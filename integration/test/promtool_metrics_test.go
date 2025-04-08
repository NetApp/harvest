package main

import (
	"errors"
	"fmt"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var PromToolLocation = utils.GetHarvestRootDir() + "/integration/test/" + "promtool"

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

func TestFormatQueries(t *testing.T) {
	utils.SkipIfMissing(t, utils.CheckFormat)

	jsonDir := utils.GetHarvestRootDir() + "/grafana/dashboards"
	slog.Info("Dashboard directory path", slog.String("jsonDir", jsonDir))
	fileSet = GetAllJsons(jsonDir)
	if len(fileSet) == 0 {
		t.Fatalf("No json file found @ %s", jsonDir)
	}
	slog.Info("Json files", slog.Int("fileSet", len(fileSet)))

	if len(fileSet) == 0 {
		TestDashboardsLoad(t)
	}

	for _, filePath := range fileSet {
		dashPath := shortPath(filePath)
		if shouldSkipDashboard(filePath) {
			slog.Info("Skip", slog.String("path", dashPath))
			continue
		}
		byteValue, _ := os.ReadFile(filePath)
		var allExpr []string
		value := gjson.Get(string(byteValue), "panels")
		for _, record := range value.Array() {
			allExpr = append(allExpr, getAllExpr(record)...)
			for _, targets := range record.Map()["targets"].Array() {
				allExpr = append(allExpr, targets.Map()["expr"].Str)
			}
		}
		allExpr = utils.RemoveDuplicateStr(allExpr)

		for _, expression := range allExpr {
			updatedExpr := util.Format(expression, PromToolLocation)
			if updatedExpr != expression {
				t.Errorf("query %s not formatted in dashboard %s, it should be %s", expression, dashPath, updatedExpr)
			}
		}
	}
}
