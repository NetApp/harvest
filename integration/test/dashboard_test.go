//+build regression

package main

import (
	"encoding/json"
	"fmt"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
)

const TOTAL_DASHBOARD = 11

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
		panic(fmt.Errorf("Grafana is not reachable."))
	}
	log.Println("Import dashboard from grafana/dashboards")
	params := []string{"import", "--addr", utils.GetGrafanaUrl(), "--directory", "grafana/dashboards"}
	containerID := docker.GetContainerID("poller")
	if len(containerID) == 0 {
		//assuming non docker based harvest
		log.Println("It is non docker based harvest")
		importOutput := utils.Exec("/opt/harvest", "bin/grafana", params...)
		log.Println(importOutput)
	} else {
		execParam := []string{"exec", containerID, "bin/grafana"}
		for index, _ := range params {
			execParam = append(execParam, params[index])
		}
		importOutput := utils.Run("docker", execParam...)
		log.Println(importOutput)
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
	log.Println(fmt.Sprintf("Find list of dashboard for folder %d", folderId))
	url := utils.GetGrafanaHttpUrl() + fmt.Sprintf("/api/search?folderIds=%d", folderId)
	log.Println(url)
	data, err = utils.GetResponseBody(url)
	utils.PanicIfNotNil(err)
	var dataDashboard []Dashboard
	err = json.Unmarshal(data, &dataDashboard)
	utils.PanicIfNotNil(err)
	assert.True(suite.T(), TOTAL_DASHBOARD == len(dataDashboard), fmt.Sprintf("Expected dashboard %d but found %d dashboards",
		TOTAL_DASHBOARD, len(dataDashboard)))
	var actualNames []string
	var notFoundList []string
	for _, values := range dataDashboard {
		actualNames = append(actualNames, values.Title)
	}
	expectedName := []string{"Harvest Metadata", "NetApp Detail: Aggregate", "NetApp Detail: Cluster",
		"NetApp Detail: Disk", "NetApp Detail: Network",
		"NetApp Detail: Network with NVMe/FC", "NetApp Detail: Node",
		"NetApp Detail: Shelf", "NetApp Detail: SnapMirror", "NetApp Detail: SVM", "NetApp Detail: Volume"}
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDashboardImportSuite(t *testing.T) {
	suite.Run(t, new(DashboardImportTestSuite))
}
