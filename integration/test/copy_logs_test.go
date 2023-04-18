package main

import (
	"errors"
	"fmt"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"os/exec"
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

func TestNoErrors(t *testing.T) {
	utils.SkipIfMissing(t, utils.AnalyzeDockerLogs)
	containerIds := docker.GetContainerID("bin/poller")
	for _, containerId := range containerIds {
		checkLogs(t, containerId)
	}
}

func checkLogs(t *testing.T, container string) {
	cli := fmt.Sprintf(`docker logs %s 2>&1 | grep -E "ERR"`, container)
	command := exec.Command("bash", "-c", cli)
	output, err := command.CombinedOutput()
	// The grep checks for matching lines.
	// An exit status of:
	//   0 means one or more lines matched
	//   1 means no lines matched
	//  >1 means an error occurred
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			// an exit code of 1 means there were no matches, that's OK
			if ee.ExitCode() == 1 {
				return

			}
		}
		t.Errorf("ERR checking logs container=%s cli=%s err=%v output=%s", container, cli, err, string(output))
		return
	}
	if len(output) > 0 {
		t.Errorf("ERRs found in poller logs container=%s size=%d. Dump of errors follows:\n%s",
			container, len(output), string(output))
	}
}
