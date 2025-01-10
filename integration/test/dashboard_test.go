package main

import (
	"encoding/json"
	"errors"
	"github.com/Netapp/harvest-automation/test/grafana"
	"github.com/Netapp/harvest-automation/test/utils"
	"log/slog"
	"testing"
	"time"
)

type Folder struct {
	ID    int64  `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
}
type Dashboard struct {
	ID          int64  `json:"id"`
	UID         string `json:"uid"`
	Title       string `json:"title"`
	FolderTitle string `json:"folderTitle"`
	FolderURL   string `json:"folderUrl"`
	FolderID    int64  `json:"folderId"`
}

var cDotFolder, sevenModeFolder string

func TestGrafanaAndPrometheusAreConfigured(t *testing.T) {
	utils.SkipIfMissing(t, utils.Regression)
	slog.Info("Verify Grafana and Prometheus are configured")
	if !utils.IsURLReachable(utils.GetGrafanaHTTPURL()) {
		panic(errors.New("grafana is not reachable"))
	}
	if !utils.IsURLReachable(utils.GetPrometheusURL()) {
		panic(errors.New("prometheus is not reachable"))
	}
	cDotFolder = "Harvest-main-cDOT"
	sevenModeFolder = "Harvest-main-7mode"
	slog.Info("Folder name details", slog.String("cMode", cDotFolder), slog.String("7mode", sevenModeFolder))
	status, out := new(grafana.Mgr).Import()
	if !status {
		t.Errorf("Grafana import operation failed out=%s", out)
	}
	time.Sleep(30 * time.Second)
}

func TestImport(t *testing.T) {
	utils.SkipIfMissing(t, utils.Regression)
	slog.Info("Verify harvest folder")
	data, err := utils.GetResponseBody(utils.GetGrafanaHTTPURL() + "/api/folders?limit=10")
	utils.PanicIfNotNil(err)
	var dataFolder []Folder
	err = json.Unmarshal(data, &dataFolder)
	utils.PanicIfNotNil(err)
	for _, values := range dataFolder {
		if values.Title == cDotFolder {
			return
		}
	}
	slog.Info("Folder data", slog.String("Data", string(data)))
	t.Error("Unable to find harvest folder")
}

func TestCModeDashboardCount(t *testing.T) {
	utils.SkipIfMissing(t, utils.Regression)
	folderID := getFolderID(t, cDotFolder)
	expectedName := []string{
		"Harvest Metadata",
		"ONTAP: Aggregate",
		"ONTAP: Cluster",
		"ONTAP: Disk",
		"ONTAP: LUN",
		"ONTAP: Network",
		"ONTAP: NFS Clients",
		"ONTAP: Node",
		"ONTAP: Shelf",
		"ONTAP: SnapMirror Sources",
		"ONTAP: SVM",
		"ONTAP: Volume",
		"ONTAP: MetroCluster",
		"ONTAP: Data Protection",
		"ONTAP: Qtree",
		"ONTAP: Security",
		"ONTAP: Power",
		"ONTAP: cDOT",
	}

	verifyDashboards(t, folderID, expectedName)
}

func TestSevenModeDashboardCount(t *testing.T) {
	utils.SkipIfMissing(t, utils.Regression)
	folderID := getFolderID(t, sevenModeFolder)
	expectedName := []string{
		"ONTAP: Aggregate 7 mode",
		"ONTAP: Cluster 7 mode",
		"ONTAP: Disk 7 mode",
		"ONTAP: LUN 7 mode",
		"ONTAP: Network 7 mode",
		"ONTAP: Node 7 mode",
		"ONTAP: Shelf 7 mode",
		"ONTAP: Volume 7 mode",
	}
	verifyDashboards(t, folderID, expectedName)
}

func getFolderID(t *testing.T, folderName string) int64 {
	slog.Info("Find " + folderName + " folder id")
	data, err := utils.GetResponseBody(utils.GetGrafanaHTTPURL() + "/api/folders?limit=100")
	utils.PanicIfNotNil(err)
	var dataFolder []Folder
	var folderID int64
	err = json.Unmarshal(data, &dataFolder)
	utils.PanicIfNotNil(err)
	for _, values := range dataFolder {
		if values.Title == folderName {
			folderID = values.ID
			break
		}
	}
	if !(folderID > 0) {
		t.Errorf("Folder id is empty or zero for folder=[%s]", folderName)
	}
	return folderID
}

func verifyDashboards(t *testing.T, folderID int64, expectedName []string) {
	slog.Info("Find list of dashboard for folder", slog.Int64("folderID", folderID))
	url := utils.GetGrafanaHTTPURL() + "/api/search?type=dash-db"
	slog.Info(url)
	data, err := utils.GetResponseBody(url)
	utils.PanicIfNotNil(err)
	var dataDashboard []Dashboard
	err = json.Unmarshal(data, &dataDashboard)
	utils.PanicIfNotNil(err)
	actualNames := make([]string, 0, len(dataDashboard))
	var notFoundList []string
	slog.Info("Folder details", slog.Int64("folderID", folderID))
	for _, values := range dataDashboard {
		actualNames = append(actualNames, values.Title)
	}
	for _, title := range expectedName {
		if !(utils.Contains(actualNames, title)) {
			notFoundList = append(notFoundList, title)
		}
	}
	if len(notFoundList) > 0 {
		slog.Info("The following dashboards were not imported successfully.")
		t.Errorf("One or more dashboards %s were missing/ not imported", notFoundList)
	}
}
