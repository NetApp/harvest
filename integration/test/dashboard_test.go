package main

import (
	"encoding/json"
	"errors"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/errs"
	"github.com/Netapp/harvest-automation/test/grafana"
	"github.com/Netapp/harvest-automation/test/request"
	"log/slog"
	"slices"
	"testing"
	"time"
)

type Folder struct {
	UID   string `json:"uid"`
	Title string `json:"title"`
}
type Dashboard struct {
	ID          int64  `json:"id"`
	UID         string `json:"uid"`
	Title       string `json:"title"`
	FolderTitle string `json:"folderTitle"`
	FolderURL   string `json:"folderUrl"`
	FolderUID   string `json:"folderUid"`
}

var cDotFolder, sevenModeFolder string

func TestGrafanaAndPrometheusAreConfigured(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.Regression)
	slog.Info("Verify Grafana and Prometheus are configured")
	if !cmds.IsURLReachable(cmds.GetGrafanaHTTPURL()) {
		panic(errors.New("grafana is not reachable"))
	}
	if !cmds.IsURLReachable(cmds.GetPrometheusURL()) {
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
	cmds.SkipIfMissing(t, cmds.Regression)
	slog.Info("Verify harvest folder")
	data, err := request.GetResponseBody(cmds.GetGrafanaHTTPURL() + "/api/folders?limit=100")
	errs.PanicIfNotNil(err)
	var dataFolder []Folder
	err = json.Unmarshal(data, &dataFolder)
	errs.PanicIfNotNil(err)
	count := 0
	for _, values := range dataFolder {
		if values.Title == cDotFolder {
			count++
		}
	}
	// A duplicate folder means import failed to recognize an existing folder,
	// see https://github.com/NetApp/harvest/issues/4460
	if count != 1 {
		slog.Info("Folder data", slog.String("Data", string(data)))
		t.Errorf("expected exactly 1 folder titled %s, got %d", cDotFolder, count)
	}
}

func TestCModeDashboardCount(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.Regression)
	folderUID := getFolderUID(t, cDotFolder)
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

	verifyDashboards(t, folderUID, expectedName)
}

func TestSevenModeDashboardCount(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.Regression)
	folderUID := getFolderUID(t, sevenModeFolder)
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
	verifyDashboards(t, folderUID, expectedName)
}

func getFolderUID(t *testing.T, folderName string) string {
	slog.Info("Find " + folderName + " folder uid")
	data, err := request.GetResponseBody(cmds.GetGrafanaHTTPURL() + "/api/folders?limit=100")
	errs.PanicIfNotNil(err)
	var dataFolder []Folder
	var folderUID string
	err = json.Unmarshal(data, &dataFolder)
	errs.PanicIfNotNil(err)
	for _, values := range dataFolder {
		if values.Title == folderName {
			folderUID = values.UID
			break
		}
	}
	if folderUID == "" {
		t.Errorf("Folder uid is empty for folder=[%s]", folderName)
	}
	return folderUID
}

func verifyDashboards(t *testing.T, folderUID string, expectedName []string) {
	slog.Info("Find list of dashboard for folder", slog.String("folderUID", folderUID))
	url := cmds.GetGrafanaHTTPURL() + "/api/search?type=dash-db&folderUIDs=" + folderUID
	slog.Info(url)
	data, err := request.GetResponseBody(url)
	errs.PanicIfNotNil(err)
	var dataDashboard []Dashboard
	err = json.Unmarshal(data, &dataDashboard)
	errs.PanicIfNotNil(err)
	actualNames := make([]string, 0, len(dataDashboard))
	var notFoundList []string
	slog.Info("Folder details", slog.String("folderUID", folderUID))
	for _, values := range dataDashboard {
		// Assert placement rather than trusting the query filter. Without this, a dashboard
		// imported into the Dashboards root still passes,
		// see https://github.com/NetApp/harvest/issues/4460
		if values.FolderUID != folderUID {
			t.Errorf("dashboard %s is in folder uid=[%s] want uid=[%s]", values.Title, values.FolderUID, folderUID)
			continue
		}
		actualNames = append(actualNames, values.Title)
	}
	for _, title := range expectedName {
		if !(slices.Contains(actualNames, title)) {
			notFoundList = append(notFoundList, title)
		}
	}
	if len(notFoundList) > 0 {
		slog.Info("The following dashboards were not imported successfully.")
		t.Errorf("One or more dashboards %s were missing/ not imported", notFoundList)
	}
}
