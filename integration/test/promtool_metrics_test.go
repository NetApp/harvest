package main

import (
	"errors"
	"fmt"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/netapp/harvest/v2/cmd/tools/grafana"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"github.com/netapp/harvest/v2/third_party/tidwall/sjson"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const (
	TopresourceConstant      = "999999"
	RangeConstant            = "888888"
	RangeReverseConstant     = "10d6h54m48s"
	IntervalConstant         = "777777"
	IntervalDurationConstant = "666666"
)

var allowedList = []string{
	"aggr_power",
	"cluster_software_status",
	"health_lif_alerts",
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

func TestFormatQueries(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.CheckFormat)
	grafana.VisitDashboards(
		[]string{
			"../../grafana/dashboards/cisco",
			"../../grafana/dashboards/cmode",
			"../../grafana/dashboards/cmode-details",
			"../../grafana/dashboards/storagegrid",
		},
		func(path string, data []byte) {
			changeExpr(t, path, data, "promtool")
		},
	)
}

func changeExpr(t *testing.T, path string, data []byte, promtoolPath string) {
	var (
		updatedData  []byte
		notFormatted bool
		errorStr     []string
		err          error
	)

	updatedData = slices.Clone(data)
	dashPath := grafana.ShortPath(path)

	// Change all panel expressions
	grafana.VisitAllPanels(updatedData, func(path string, _, value gjson.Result) {
		title := value.Get("title").ClonedString()
		// Rewrite expressions
		value.Get("targets").ForEach(func(targetKey, target gjson.Result) bool {
			expr := target.Get("expr")
			if expr.Exists() && expr.ClonedString() != "" {
				updatedExpr := format(expr.ClonedString(), promtoolPath)
				if updatedExpr != expr.ClonedString() {
					notFormatted = true
					updatedData, err = sjson.SetBytes(updatedData, path+".targets."+targetKey.ClonedString()+".expr", []byte(updatedExpr))
					if err != nil {
						fmt.Printf("Error while updating the panel query format: %v\n", err)
					}
					errorStr = append(errorStr, fmt.Sprintf("query not formatted in dashboard %s panel `%s`, it should be \n %s\n", dashPath, title, updatedExpr))
				}
			}
			return true
		})
	})
	if notFormatted {
		sortedPath := writeFormatted(t, dashPath, updatedData)
		path = "grafana/dashboards/" + dashPath
		t.Errorf("%v \nFormatted version created at path=%s.\ncp %s %s",
			errorStr, sortedPath, sortedPath, path)
	}
}

func format(query string, path string) string {
	replacedQuery := strings.ReplaceAll(query, "$TopResources", TopresourceConstant)
	replacedQuery = strings.ReplaceAll(replacedQuery, "$__range", RangeConstant)
	replacedQuery = strings.ReplaceAll(replacedQuery, "$__interval", IntervalConstant)
	replacedQuery = strings.ReplaceAll(replacedQuery, "${Interval}", IntervalDurationConstant)

	command := exec.Command(path, "--experimental", "promql", "format", replacedQuery)
	output, err := command.CombinedOutput()
	updatedQuery := strings.TrimSuffix(string(output), "\n")
	if strings.HasPrefix(updatedQuery, "  ") {
		updatedQuery = strings.TrimLeft(updatedQuery, " ")
	}
	if err != nil {
		// An exit code can't be used since we need to ignore metrics that are not formatted but can't change
		fmt.Printf("ERR formating metrics query=%s err=%v output=%s", query, err, string(output))
		return query
	}

	if len(output) == 0 {
		return query
	}

	updatedQuery = strings.ReplaceAll(updatedQuery, TopresourceConstant, "$TopResources")
	updatedQuery = strings.ReplaceAll(updatedQuery, RangeReverseConstant, "$__range")
	updatedQuery = strings.ReplaceAll(updatedQuery, IntervalConstant, "$__interval")
	updatedQuery = strings.ReplaceAll(updatedQuery, IntervalDurationConstant, "${Interval}")
	return updatedQuery
}

func writeFormatted(t *testing.T, path string, updatedData []byte) string {
	dir, file := filepath.Split(path)
	dir = filepath.Dir(dir)
	tempDir := "/tmp"
	dest := filepath.Join(tempDir, dir, file)
	destDir := filepath.Dir(dest)
	err := os.MkdirAll(destDir, 0750)
	if err != nil {
		t.Errorf("failed to create dir=%s err=%v", destDir, err)
		return ""
	}
	create, err := os.Create(dest)

	if err != nil {
		t.Errorf("failed to create file=%s err=%v", dest, err)
		return ""
	}
	_, err = create.Write(updatedData)
	if err != nil {
		t.Errorf("failed to write formatted json to file=%s err=%v", dest, err)
		return ""
	}
	err = create.Close()
	if err != nil {
		t.Errorf("failed to close file=%s err=%v", dest, err)
		return ""
	}
	return dest
}
