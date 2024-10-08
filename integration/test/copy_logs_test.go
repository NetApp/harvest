package main

import (
	"errors"
	"fmt"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"os/exec"
	"testing"
)

func TestCopyLogs(t *testing.T) {
	utils.SkipIfMissing(t, utils.CopyDockerLogs)
	installer.CleanLogDir()
	installer.CreateLogDir()
	pollerProcessName := "bin/poller"
	harvestLogDir := installer.LogDir
	containerIDs, err := docker.Containers(pollerProcessName)
	if err != nil {
		panic(err)
	}
	for _, container := range containerIDs {
		containerShortID := container.ID[:10]
		dest := harvestLogDir + "/" + containerShortID + ".log"
		err = docker.StoreContainerLog(containerShortID, dest)
		if err != nil {
			slog.Error(
				"Unable to copy logs",
				slogx.Err(err),
				slog.String("id", containerShortID),
				slog.String("dest", dest),
			)
		}
	}
}

func TestNoErrors(t *testing.T) {
	utils.SkipIfMissing(t, utils.AnalyzeDockerLogs)
	containerIDs, err := docker.Containers("bin/poller")
	if err != nil {
		panic(err)
	}
	for _, container := range containerIDs {
		checkLogs(t, container)
	}
}

func checkLogs(t *testing.T, container docker.Container) {
	cli := fmt.Sprintf(`docker logs %s 2>&1 | grep -Ev '%s' | grep -E "ERR"`, container.ID, ignoreList())
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
		t.Errorf("ERR checking logs name=%s container=%s cli=%s err=%v output=%s",
			container.Name(), container.ID, cli, err, string(output))
		return
	}
	if len(output) > 0 {
		t.Errorf("ERRs found in poller logs name=%s id=%s size=%d. Dump of errors follows:\n%s",
			container.Name(), container.ID[:6], len(output), string(output))
	}
}

// ignoreList returns a list of regex patterns that will be ignored
func ignoreList() any {
	return `RPC: Remote system error|connection error|Code: 2426405`
}
