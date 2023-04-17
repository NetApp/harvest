package main

import (
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"testing"
)

func TestCopyLogs(t *testing.T) {
	utils.SkipIfMissing(t, utils.CopyDockerLogs)
	installer.CleanLogDir()
	installer.CreateLogDir()
	pollerProcessName := "bin/poller"
	harvestLogDir := installer.LogDir
	containerIds, err := docker.Containers(pollerProcessName)
	if err != nil {
		panic(err)
	}
	for _, container := range containerIds {
		containerShortId := container.Id[:10]
		dest := harvestLogDir + "/" + containerShortId + ".log"
		err = docker.StoreContainerLog(containerShortId, dest)
		if err != nil {
			log.Error().Err(err).Str("id", containerShortId).Str("dest", dest).Msg("Unable to copy logs")
		}
	}
}
