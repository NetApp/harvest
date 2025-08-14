package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"strings"
	"testing"
)

func TestDockerInstall(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.InstallDocker)
	// it will create a grafana token and configure it for dashboard export
	isReachable := cmds.WaitForGrafana()
	if !isReachable {
		t.Fatalf("Grafana is not reachable.")
	}
	cmds.WriteToken(cmds.CreateGrafanaToken())
	containerIDs, err := docker.Containers("poller")
	if err != nil {
		panic(err)
	}
	fileZapiName := installer.GetPerfFileWithQosCounters(installer.ZapiPerfDefaultFile, "defaultZapi.yaml")
	fileRestName := installer.GetPerfFileWithQosCounters(installer.RestPerfDefaultFile, "defaultRest.yaml")
	for _, container := range containerIDs {
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
		for _, id := range ids {
			if !isValidAsup(id.ID) {
				panic("Asup validation failed")
			}
		}
	} else {
		panic("No pollers running")
	}

}

func isValidAsup(containerName string) bool {
	out, err := cmds.Exec("", "docker", nil, "container", "exec", containerName, "autosupport/asup", "--version")
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
