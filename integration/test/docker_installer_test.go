//go:build install_docker

package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/setup"
	"github.com/Netapp/harvest-automation/test/utils"
	"strings"
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
	ids := docker.GetContainerID("poller")
	if len(ids) > 0 {
		id := ids[0]
		if !isValidAsup(id) {
			panic("Asup validation failed")
		}
	} else {
		panic("No pollers running")
	}

}

func isValidAsup(containerName string) bool {
	out, err := utils.Exec("", "docker", nil, "container", "exec", containerName, "autosupport/asup", "--version")
	if err != nil {
		fmt.Printf("error %s\n", err)
		return false
	}
	if !strings.Contains(out, "endpoint:stable") {
		fmt.Printf("asup endpoint is not stable %s\n", out)
		return false
	}
	fmt.Printf("asup validation successful %s\n", out)
	return true
}
