package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"strings"
	"testing"
)

func TestDockerInstall(t *testing.T) {
	utils.SkipIfMissing(t, utils.InstallDocker)
	//it will create a grafana token and configure it for dashboard export
	isReachable := utils.WaitForGrafana()
	if !isReachable {
		t.Fatalf("Grafana is not reachable.")
	}
	utils.WriteToken(utils.CreateGrafanaToken())
	containerIds, err := docker.Containers("poller")
	if err != nil {
		panic(err)
	}
	fileZapiName := installer.GetPerfFileWithQosCounters(installer.ZapiPerfDefaultFile, "defaultZapi.yaml")
	fileRestName := installer.GetPerfFileWithQosCounters(installer.RestPerfDefaultFile, "defaultRest.yaml")
	for _, container := range containerIds {
		docker.CopyFile(container.ID, installer.HarvestConfigFile, installer.HarvestHome+"/"+installer.HarvestConfigFile)
		docker.CopyFile(container.ID, fileZapiName, installer.HarvestHome+"/"+installer.ZapiPerfDefaultFile)
		docker.CopyFile(container.ID, fileRestName, installer.HarvestHome+"/"+installer.RestPerfDefaultFile)
	}
	_ = docker.ReStartContainers("poller")
	ids, err := docker.Containers("poller")
	if err != nil {
		panic(err)
	}
	if len(ids) > 0 {
		id := ids[0].ID
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
