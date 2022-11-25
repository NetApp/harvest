//go:build install_docker

package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/setup"
	"github.com/Netapp/harvest-automation/test/utils"
	"testing"
)

func TestDockerInstall(t *testing.T) {
	utils.SetupLogging()
	//it will create a grafana token and configure it for dashboard export
	if !utils.IsURLReachable(utils.GetGrafanaHTTPURL()) {
		panic(fmt.Errorf("Grafana is not reachable."))
	}
	utils.WriteToken(utils.CreateGrafanaToken())
	containerIds := docker.GetContainerID("poller")
	fileZapiName := setup.GetPerfFileWithQosCounters(setup.ZapiPerfDefaultFile, "defaultZapi.yaml")
	fileRestName := setup.GetPerfFileWithQosCounters(setup.RestPerfDefaultFile, "defaultRest.yaml")
	for _, containerId := range containerIds {
		docker.CopyFile(containerId, installer.HarvestConfigFile, installer.HarvestHome+"/"+installer.HarvestConfigFile)
		docker.CopyFile(containerId, fileZapiName, installer.HarvestHome+"/"+setup.ZapiPerfDefaultFile)
		docker.CopyFile(containerId, fileRestName, installer.HarvestHome+"/"+setup.RestPerfDefaultFile)
	}
	docker.ReStartContainers("poller")
}
