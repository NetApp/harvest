//go:build copy_docker_logs

package main

import (
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"testing"
)

func TestCopyLogs(t *testing.T) {
	utils.SetupLogging()
	installer.CleanLogDir()
	installer.CreateLogDir()
	pollerProcessName := "bin/poller"
	harvestLogDir := installer.LOG_DIR
	containerIds := docker.GetContainerID(pollerProcessName)
	for _, containerId := range containerIds {
		containerShortId := containerId[:10]
		docker.StoreContainerLog(containerShortId, harvestLogDir+"/"+containerShortId+".log")
	}
}
