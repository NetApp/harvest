package main

import (
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"testing"
)

func TestCopyLogs(t *testing.T) {
	utils.SkipIfMissing(t, utils.CopyDockerLogs)
	installer.CleanLogDir()
	installer.CreateLogDir()
	pollerProcessName := "bin/poller"
	harvestLogDir := installer.LogDir
	containerIds := docker.GetContainerID(pollerProcessName)
	for _, containerId := range containerIds {
		containerShortId := containerId[:10]
		docker.StoreContainerLog(containerShortId, harvestLogDir+"/"+containerShortId+".log")
	}
}
