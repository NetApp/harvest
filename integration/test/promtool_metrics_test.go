package main

import (
	"errors"
	"fmt"
	"github.com/Netapp/harvest-automation/test/cmds"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

var allowedList = []string{
	"aggr_power",
	"cluster_software_status",
	"health_lif_alerts",
	"health_support_alerts",
	"security_certificate_labels",
	"shelf_average_ambient_temperature",
	"shelf_average_fan_speed",
	"shelf_average_temperature",
	"shelf_labels",
	"shelf_max_fan_speed",
	"shelf_max_temperature",
	"shelf_min_ambient_temperature",
	"shelf_min_fan_speed",
	"shelf_min_temperature",
	"shelf_power",
	"snapmirror_labels",
	"volume_arw_status",
	"volume_labels",
}

func TestPrometheusMetrics(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.CheckMetrics)
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
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			// An exit code can't be used since we need to ignore metrics that are not valid but can't change
			t.Errorf("ERR checking metrics cli=%s err=%v output=%s", cli, err, string(output))
			return
		}
	}

	if len(output) == 0 {
		return
	}

	// Read the output, line by line, and check for errors, non-errors are ignored
	lines := strings.SplitSeq(string(output), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "label names should be written in 'snake_case' not 'camelCase'") {
			metricName := strings.Split(line, " ")[0]
			if !slices.Contains(allowedList, metricName) {
				t.Errorf("ERR %s", line)
			}
		}
		if strings.Contains(line, "error while linting: ") {
			t.Errorf("promtool: %s", line)
		}
	}
}
