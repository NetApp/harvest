package main

import (
	"errors"
	"fmt"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"os/exec"
	"testing"
)

func TestCopyLogs(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.CopyDockerLogs)
	installer.CleanLogDir()
	installer.CreateLogDir()
	pollerProcessName := "bin/poller"
	harvestLogDir := installer.LogDir
	containerIDs, err := docker.Containers(pollerProcessName)
	assert.Nil(t, err)
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

type containerInfo struct {
	name          string
	ignorePattern string
	errorPattern  string
}

func TestNoErrors(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.AnalyzeDockerLogs)

	containerPatterns := []containerInfo{
		{name: "bin/poller", ignorePattern: pollerIgnore(), errorPattern: "ERR"},
		{name: "prometheus", errorPattern: "level=error|level=warn"},
		{name: "grafana", errorPattern: "level=error"},
	}

	for _, containerPattern := range containerPatterns {
		containers, err := docker.Containers(containerPattern.name)
		assert.Nil(t, err)
		for _, container := range containers {
			checkLogs(t, container, containerPattern)
		}
	}
}

func checkLogs(t *testing.T, container docker.Container, info containerInfo) {
	cli := fmt.Sprintf(`docker logs %s 2>&1 | grep -Ev '%s' | grep -E '%s'`, container.ID, info.ignorePattern, info.errorPattern)
	command := exec.Command("bash", "-c", cli)
	output, err := command.CombinedOutput()
	// The grep checks for matching lines.
	// An exit status of:
	//   0 means one or more lines matched
	//   1 means no lines matched
	//  >1 means an error occurred
	if err != nil {
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
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

// pollerIgnore returns a list of regex patterns that will be ignored
func pollerIgnore() string {
	return `RPC: Remote system error|connection error|Code: 2426405|failed to fetch data: error making request StatusCode: 403, Error: Permission denied, Message: not authorized for that command, API: (/api/support/autosupport)`
}
