//go:build regression || dashboard
// +build regression dashboard

package main

import (
	"encoding/json"
	"fmt"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/grafana"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
	"time"
)

type DashboardImportTestSuite struct {
	suite.Suite
}

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

func (suite *DashboardImportTestSuite) SetupSuite() {
	log.Println("Verify Grafana and Prometheus are configured")
	if !utils.IsUrlReachable(utils.GetGrafanaHttpUrl()) {
		panic(fmt.Errorf("Grafana is not reachable."))
	}
	if !utils.IsUrlReachable(utils.GetPrometheusUrl()) {
		panic(fmt.Errorf("Prometheus is not reachable."))
	}
	status, _ := new(grafana.GrafanaMgr).Import("") //send empty so that it will import all dashboards
	if !status {
		assert.Fail(suite.T(), "Grafana import operation is failed")
	}
	time.Sleep(30 * time.Second)
}

func (suite *DashboardImportTestSuite) TestImport() {
	log.Println("Verify harvest folder")
	data, err := utils.GetResponseBody(utils.GetGrafanaHttpUrl() + "/api/folders?limit=10")
	utils.PanicIfNotNil(err)
	var dataFolder []Folder
	err = json.Unmarshal(data, &dataFolder)
	utils.PanicIfNotNil(err)
	for _, values := range dataFolder {
		if values.Title == "Harvest 2.0 - cDOT" {
			return
		}
	}
	log.Println(data)
	assert.Fail(suite.T(), "Unable to find harvest folder")
}

func (suite *DashboardImportTestSuite) TestCModeDashboardCount() {
	folderId := GetFolderId("Harvest 2.0 - cDOT", suite.T())
	expectedName := []string{"Harvest Metadata", "NetApp Detail: Aggregate", "NetApp Detail: Cluster",
		"NetApp Detail: Disk", "NetApp Detail: LUN", "NetApp Detail: Network", "NetApp Detail: Network  - Details",
		"NetApp Detail: Network with NVMe/FC", "NetApp Detail: Node", "NetApp Detail: Node - Details",
		"NetApp Detail: Shelf", "NetApp Detail: SnapMirror", "NetApp Detail: SVM", "NetApp Detail: SVM - Details",
		"NetApp Detail: Volume", "NetApp Detail: Volume - Details", "NetApp Detail: MetroCluster"}

	VerifyDashboards(folderId, expectedName, suite.T())
}

func (suite *DashboardImportTestSuite) TestSevenModeDashboardCount() {

	folderId := GetFolderId("Harvest 2.0 - 7-mode", suite.T())
	expectedName := []string{"NetApp Detail: Aggregate 7 mode", "NetApp Detail: Cluster 7 mode",
		"NetApp Detail: Disk 7 mode", "NetApp Detail: LUN 7 mode", "NetApp Detail: Network 7 mode", "NetApp Detail: Network with NVMe/FC 7 mode",
		"NetApp Detail: Node 7 mode", "NetApp Detail: Shelf 7 mode", "NetApp Detail: Volume 7 mode"}
	VerifyDashboards(folderId, expectedName, suite.T())
}

func (suite *DashboardImportTestSuite) TestImportForInvalidJson() {
	jsonDir := "grafana/dashboard_jsons"
	if docker.IsDockerBasedPoller() {
		containerId := docker.GetOnePollerContainers()
		docker.CopyFile(containerId, "dashboard_jsons", "/opt/harvest/"+jsonDir)
	} else {
		utils.Run("cp", "-R", "dashboard_jsons", installer.HarvestHome+"/grafana")
	}
	status, _ := new(grafana.GrafanaMgr).Import(jsonDir)
	if status {
		assert.Fail(suite.T(), "Grafana import should fail but it is succeeded")
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDashboardImportSuite(t *testing.T) {
	suite.Run(t, new(DashboardImportTestSuite))
}

func GetFolderId(folderName string, t *testing.T) int64 {
	log.Println("Find " + folderName + " folder id")
	data, err := utils.GetResponseBody(utils.GetGrafanaHttpUrl() + "/api/folders?limit=100")
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
		assert.Fail(t, "Folder id is empty or zero.")
	}
	return folderId
}

func VerifyDashboards(folderId int64, expectedName []string, t *testing.T) {
	log.Println(fmt.Sprintf("Find list of dashboard for folder %d", folderId))
	url := utils.GetGrafanaHttpUrl() + "/api/search?type=dash-db"
	log.Println(url)
	data, err := utils.GetResponseBody(url)
	utils.PanicIfNotNil(err)
	var dataDashboard []Dashboard
	err = json.Unmarshal(data, &dataDashboard)
	utils.PanicIfNotNil(err)
	var actualNames []string
	var notFoundList []string
	log.Println(folderId)
	for _, values := range dataDashboard {
		actualNames = append(actualNames, values.Title)
	}
	for _, title := range expectedName {
		if !(utils.Contains(actualNames, title)) {
			notFoundList = append(notFoundList, title)
		}
	}
	if len(notFoundList) > 0 {
		log.Println("The following dashboards were not imported successfully.")
		assert.Fail(t, fmt.Sprintf("One or more dashboards %s were missing/ not imported", notFoundList))
	}
}
