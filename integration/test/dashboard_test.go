//+build regression dashboard

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
}

func (suite *DashboardImportTestSuite) SetupSuite() {
	log.Println("Verify Grafana and Prometheus are configured")
	if !utils.IsUrlReachable(utils.GetGrafanaHttpUrl()) {
		panic(fmt.Errorf("Grafana is not reachable."))
	}
	if !utils.IsUrlReachable(utils.GetPrometheusUrl()) {
		panic(fmt.Errorf("Prometheus is not reachable."))
	}
	status, _ := new(grafana.GrafanaMgr).Import("grafana/dashboards")
	if !status {
		assert.Fail(suite.T(), "Grafana import operation is failed")
	}
}

func (suite *DashboardImportTestSuite) TestImport() {
	log.Println("Verify harvest folder")
	data, err := utils.GetResponseBody(utils.GetGrafanaHttpUrl() + "/api/folders?limit=10")
	utils.PanicIfNotNil(err)
	var dataFolder []Folder
	err = json.Unmarshal(data, &dataFolder)
	utils.PanicIfNotNil(err)
	for _, values := range dataFolder {
		if values.Title == "Harvest 2.0" {
			return
		}
	}
	log.Println(data)
	assert.Fail(suite.T(), "Unable to find harvest folder")
}

func (suite *DashboardImportTestSuite) TestDashboardCount() {
	log.Println("Find harvest folder id")
	data, err := utils.GetResponseBody(utils.GetGrafanaHttpUrl() + "/api/folders?limit=10")
	utils.PanicIfNotNil(err)
	var dataFolder []Folder
	var folderId int64
	err = json.Unmarshal(data, &dataFolder)
	utils.PanicIfNotNil(err)
	for _, values := range dataFolder {
		if values.Title == "Harvest 2.0" {
			folderId = values.Id
			break
		}
	}
	if !(folderId > 0) {
		assert.Fail(suite.T(), "Folder id is empty or zero.")
	}
	expectedName := []string{"Harvest Metadata", "NetApp Detail: Aggregate", "NetApp Detail: Cluster",
		"NetApp Detail: Disk", "NetApp Detail: LUN", "NetApp Detail: Network", "NetApp Detail: Network  - Details",
		"NetApp Detail: Network with NVMe/FC", "NetApp Detail: Node", "NetApp Detail: Node - Details",
		"NetApp Detail: Shelf", "NetApp Detail: SnapMirror", "NetApp Detail: SVM", "NetApp Detail: SVM - Details",
		"NetApp Detail: Volume", "NetApp Detail: Volume - Details"}

	log.Println(fmt.Sprintf("Find list of dashboard for folder %d", folderId))
	url := utils.GetGrafanaHttpUrl() + fmt.Sprintf("/api/search?folderIds=%d", folderId)
	log.Println(url)
	data, err = utils.GetResponseBody(url)
	utils.PanicIfNotNil(err)
	var dataDashboard []Dashboard
	err = json.Unmarshal(data, &dataDashboard)
	utils.PanicIfNotNil(err)
	totalDashboardCount := len(expectedName)
	assert.True(suite.T(), totalDashboardCount == len(dataDashboard), fmt.Sprintf("Expected dashboard %d but found %d dashboards",
		totalDashboardCount, len(dataDashboard)))
	var actualNames []string
	var notFoundList []string
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
		assert.Fail(suite.T(), fmt.Sprintf("One or more dashboards %s were missing/ not imported", notFoundList))
	}
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
