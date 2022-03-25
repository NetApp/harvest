//go:build install_docker
// +build install_docker

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
	if !utils.IsUrlReachable(utils.GetGrafanaHttpUrl()) {
		panic(fmt.Errorf("Grafana is not reachable."))
	}
	utils.WriteToken(utils.CreateGrafanaToken())
	containerIds := docker.GetContainerID("poller")
	fileName := setup.GetZapiPerfFileWithQosCounters()
	for _, containerId := range containerIds {
		docker.CopyFile(containerId, installer.HarvestConfigFile, installer.HarvestHome+"/"+installer.HarvestConfigFile)
		docker.CopyFile(containerId, fileName, installer.HarvestHome+"/"+setup.ZapiPerfDefaultFile)
	}
	docker.ReStartContainers("poller")
}
