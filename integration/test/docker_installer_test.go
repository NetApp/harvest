package main

import (
	"crypto/fips140"
	"fmt"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"log/slog"
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
	containers, err := docker.Containers("poller")
	if err != nil {
		panic(err)
	}

	if len(containers) == 0 {
		panic("No pollers running")
	}

	for _, container := range containers {
		if fips140.Enabled() && container.Name() == "/poller-umeng-aff300-05-06" {
			// FIPS 140-3 is only supported on ONTAP 9.11.1+
			// u2 is running version 9.9.1 so ignore FIPs failures on it
			slog.Warn("Skipping FIPS validation for container", slog.String("containerName", container.Name()))
			continue
		}
		if !isValidAsup(container.ID) {
			panic("Asup validation failed")
		}
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
