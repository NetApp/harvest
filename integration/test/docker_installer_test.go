//+build install_docker

package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"testing"
)

func TestDockerInstall(t *testing.T) {
	//it will create a grafana token and configure it for dashboard export
	if !utils.IsUrlReachable(utils.GetGrafanaHttpUrl()) {
		panic(fmt.Errorf("Grafana is not reachable."))
	}
	utils.WriteToken(utils.CreateGrafanaToken())
	containerIds := docker.GetContainerID("poller")
	for _, containerId := range containerIds {
		docker.CopyFile(containerId, installer.HarvestConfigFile, installer.HarvestHome+"/"+installer.HarvestConfigFile)
	}
}
