package main

import (
	"errors"
	"fmt"
	"github.com/Netapp/harvest-automation/test/utils"
	"os/exec"
	"strings"
	"testing"
)

func TestPrometheusMetrics(t *testing.T) {
	utils.SkipIfMissing(t, utils.CheckMetrics)

	ports := []int{12990, 12992, 12993, 12994}
	for _, port := range ports {
		checkMetrics(t, port)
	}
}

func checkMetrics(t *testing.T, port int) {
	cli := fmt.Sprintf(`curl -s http://localhost:%d/metrics | tee /tmp/metrics:%d.txt | promtool check metrics`, port, port)
	command := exec.Command("bash", "-c", cli)
	output, err := command.CombinedOutput()

	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			// An exit code can't be used since we need to ignore metrics that are not valid but can't change
			t.Errorf("ERR checking metrics cli=%s err=%v output=%s", cli, err, string(output))
			return
		}
	}

	if len(output) == 0 {
		return
	}

	// Read the output, line by line, and check for errors, non-errors are ignored

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "error while linting: ") {
			t.Errorf("promtool: %s", line)
		}
	}
}
