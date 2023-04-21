package main

import (
	"encoding/json"
	"fmt"
	"github.com/Netapp/harvest-automation/test/grafana"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/cmd/harvest/version"
	"github.com/rs/zerolog/log"
	"testing"
	"time"
)

type Folder struct {
	Id    int64  `json:"id"`
	Uid   string `json:"uid"`
	Title string `json:"title"`
}
type Dashboard struct {
	Id          int64  `json:"id"`
	Uid         string `json:"uid"`
	Title       string `json:"title"`
	FolderTitle string `json:"folderTitle"`
	FolderUrl   string `json:"folderUrl"`
	FolderId    int64  `json:"folderId"`
}

var cDotFolder, sevenModeFolder string

func TestGrafanaAndPrometheusAreConfigured(t *testing.T) {
	utils.SkipIfMissing(t, utils.Regression)
	log.Info().Msg("Verify Grafana and Prometheus are configured")
	if !utils.IsURLReachable(utils.GetGrafanaHTTPURL()) {
		panic(fmt.Errorf("grafana is not reachable"))
	}
	if !utils.IsURLReachable(utils.GetPrometheusURL()) {
		panic(fmt.Errorf("prometheus is not reachable"))
	}
	cDotFolder = "Harvest-" + version.VERSION + "-cDOT"
	sevenModeFolder = "Harvest-" + version.VERSION + "-7mode"
	log.Info().Str("cMode", cDotFolder).Str("7mode", sevenModeFolder).Msg("Folder name details")
	status, out := new(grafana.Mgr).Import("") //send empty so that it will import all dashboards
	if !status {
		t.Errorf("Grafana import operation failed out=%s", out)
	}
	time.Sleep(30 * time.Second)
}

func TestImport(t *testing.T) {
	utils.SkipIfMissing(t, utils.Regression)
	log.Info().Msg("Verify harvest folder")
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
	log.Info().Bytes("Data", data).Msg("Folder data")
	t.Error("Unable to find harvest folder")
}

func TestCModeDashboardCount(t *testing.T) {
	utils.SkipIfMissing(t, utils.Regression)
	folderId := getFolderId(cDotFolder, t)
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
		"ONTAP: SnapMirror",
		"ONTAP: SVM",
		"ONTAP: Volume",
		"ONTAP: MetroCluster",
		"ONTAP: Data Protection Snapshots",
		"ONTAP: Qtree",
		"ONTAP: Security",
		"ONTAP: Compliance",
		"ONTAP: Power",
		"ONTAP: cDOT",
	}

	verifyDashboards(folderId, expectedName, t)
}

func TestSevenModeDashboardCount(t *testing.T) {
	utils.SkipIfMissing(t, utils.Regression)
	folderId := getFolderId(sevenModeFolder, t)
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
	verifyDashboards(folderId, expectedName, t)
}

func getFolderId(folderName string, t *testing.T) int64 {
	log.Info().Msg("Find " + folderName + " folder id")
	data, err := utils.GetResponseBody(utils.GetGrafanaHTTPURL() + "/api/folders?limit=100")
	utils.PanicIfNotNil(err)
	var dataFolder []Folder
	var folderId int64
	err = json.Unmarshal(data, &dataFolder)
	utils.PanicIfNotNil(err)
	for _, values := range dataFolder {
		if values.Title == folderName {
			folderId = values.Id
			break
		}
	}
	if !(folderId > 0) {
		t.Errorf("Folder id is empty or zero for folder=[%s]", folderName)
	}
	return folderId
}

func verifyDashboards(folderId int64, expectedName []string, t *testing.T) {
	log.Info().Msg(fmt.Sprintf("Find list of dashboard for folder %d", folderId))
	url := utils.GetGrafanaHTTPURL() + "/api/search?type=dash-db"
	log.Info().Msg(url)
	data, err := utils.GetResponseBody(url)
	utils.PanicIfNotNil(err)
	var dataDashboard []Dashboard
	err = json.Unmarshal(data, &dataDashboard)
	utils.PanicIfNotNil(err)
	var actualNames []string
	var notFoundList []string
	log.Info().Int64("folderId", folderId).Msg("Folder details")
	for _, values := range dataDashboard {
		actualNames = append(actualNames, values.Title)
	}
	for _, title := range expectedName {
		if !(utils.Contains(actualNames, title)) {
			notFoundList = append(notFoundList, title)
		}
	}
	if len(notFoundList) > 0 {
		log.Info().Msg("The following dashboards were not imported successfully.")
		t.Errorf("One or more dashboards %s were missing/ not imported", notFoundList)
	}
}
