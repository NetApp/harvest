package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/netapp/harvest/v2/cmd/tools"
	"github.com/netapp/harvest/v2/pkg/template"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/go-version"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

type AlertRule struct {
	name   string
	exprs  []string
	labels []string
}

var labelRegex = regexp.MustCompile(`\{{.*?}}`)
var pluginGeneratedMetric = map[string][]string{
	"change_log": {"svm", "state", "type", "anti_ransomware_state", "object", "node", "location", "healthy", "volume", "style", "aggr", "status"},
}
var exceptionMetrics = []string{
	"up",
	"node_nfs_latency", // need to check why it was skipped in restperf
}

func TestAlertRules(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.Regression)
	metrics, _ := tools.BuildMetrics("../..", "integration/test/harvest.yml", "dc1", nil, make(map[string]tools.PanelData))
	for pluginMetric, pluginLabels := range pluginGeneratedMetric {
		metrics[pluginMetric] = tools.Counter{Name: pluginMetric, Labels: pluginLabels}
	}

	alertRules := GetAllAlertRules("../../container/prometheus/", "alert_rules.yml", false)
	for _, alertRule := range alertRules {
		for _, label := range alertRule.labels {
			found := false
			for _, expr := range alertRule.exprs {
				if !slices.Contains(exceptionMetrics, expr) {
					metricCounters := metrics[expr]
					if slices.Contains(metricCounters.Labels, label) {
						found = true
						break
					}
				} else {
					found = true
				}
			}
			if !found {
				t.Errorf("%s is not available in %s metric", label, alertRule.exprs)
			}
		}
	}
}

func TestEmsAlertRules(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.Regression)
	templateEmsLabels := getEmsLabels("../../conf/ems/9.6.0/ems.yaml")
	emsAlertRules := GetAllAlertRules("../../container/prometheus/", "ems_alert_rules.yml", true)
	for _, alertRule := range emsAlertRules {
		for _, ems := range alertRule.exprs {
			emsLabels := templateEmsLabels[ems]
			for _, label := range alertRule.labels {
				if !slices.Contains(emsLabels, label) {
					t.Errorf("%s is not available in %s ems", label, ems)
				}
			}
		}
	}
}

// eseriesMelGlobalLabels: set by the collector via mat.SetGlobalLabel, not present in events.yaml.
var eseriesMelGlobalLabels = []string{"array", "chassis_serial"}

func TestEseriesAlertRules(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.Regression)

	versionDir, err := latestVersionDir("../../conf/eseriesmel")
	if err != nil {
		t.Fatalf("unable to find eseriesmel template version: %s", err)
	}
	eventsTemplatePath := filepath.Join(versionDir, "events.yaml")

	data, err := tree.ImportYaml(eventsTemplatePath)
	if data == nil || err != nil {
		t.Fatalf("unable to import template file %s: %s", eventsTemplatePath, err)
	}

	eventTypes := getEseriesMelEventTypes(data)
	exportedLabels := getEseriesMelExportedLabels(data)

	// isEms=true reuses FindEms, which assumes a single-label selector like {event_type="N"}.
	eseriesAlertRules := GetAllAlertRules("../../container/prometheus/", "eseries_alert_rules.yml", true)
	for _, alertRule := range eseriesAlertRules {
		for _, eventType := range alertRule.exprs {
			if !slices.Contains(eventTypes, eventType) {
				t.Errorf("event_type %s used by alert %q is not defined in %s", eventType, alertRule.name, eventsTemplatePath)
			}
		}
		for _, label := range alertRule.labels {
			if !slices.Contains(exportedLabels, label) {
				t.Errorf("%s is not an exported label for alert %q", label, alertRule.name)
			}
		}
	}
}

func getEseriesMelEventTypes(data *node.Node) []string {
	children := data.GetChildS("events").GetChildren()
	eventTypes := make([]string, 0, len(children))
	for _, e := range children {
		eventTypes = append(eventTypes, e.GetChildContentS("event_type"))
	}
	return eventTypes
}

func getEseriesMelExportedLabels(data *node.Node) []string {
	labels := slices.Clone(eseriesMelGlobalLabels)

	if counters := data.GetChildS("counters"); counters != nil {
		for _, counter := range counters.GetAllChildContentS() {
			_, display, _, _ := template.ParseMetric(counter)
			if display != "" {
				labels = append(labels, display)
			}
		}
	}

	if exportOptions := data.GetChildS("export_options"); exportOptions != nil {
		if instanceKeys := exportOptions.GetChildS("instance_keys"); instanceKeys != nil {
			labels = append(labels, instanceKeys.GetAllChildContentS()...)
		}
	}

	return labels
}

// latestVersionDir returns the subdirectory of baseDir with the highest version-numbered name,
// so the test keeps covering the newest eseriesmel template instead of a hardcoded version.
func latestVersionDir(baseDir string) (string, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return "", err
	}

	var (
		latestName string
		latestVer  *version.Version
	)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		v, err := version.NewVersion(e.Name())
		if err != nil {
			continue
		}
		if latestVer == nil || v.GreaterThan(latestVer) {
			latestVer = v
			latestName = e.Name()
		}
	}
	if latestVer == nil {
		return "", fmt.Errorf("no version directories found in %s", baseDir)
	}
	return filepath.Join(baseDir, latestName), nil
}

func GetAllAlertRules(dir string, fileName string, isEms bool) []AlertRule {
	alertRules := make([]AlertRule, 0) //nolint:prealloc
	alertNames := make([]string, 0)
	exprList := make([]string, 0)
	summaryList := make([]string, 0)
	alertRulesFilePath := filepath.Join(dir, fileName)
	data, err := tree.ImportYaml(alertRulesFilePath)
	if err != nil {
		panic(err)
	}

	for _, v := range data.GetChildS("groups").GetChildren() {
		if v.GetNameS() == "rules" {
			for _, a := range v.GetChildren() {
				switch a.GetNameS() {
				case "alert":
					alertname := a.GetContentS()
					alertNames = append(alertNames, alertname)
				case "expr":
					alertexp := a.GetContentS()
					exprList = append(exprList, alertexp)
				case "annotations":
					alertSummary := a.GetChildS("summary")
					summaryList = append(summaryList, alertSummary.GetContentS())
				}
			}
		}
	}

	for i := range alertNames {
		alertRules = append(alertRules, AlertRule{name: alertNames[i], exprs: getAllExpressions(exprList[i], isEms), labels: getAllLabels(summaryList[i])})
	}
	return alertRules
}

func getAllExpressions(expression string, isEms bool) []string {
	filtered := make([]string, 0)
	var all []string
	if isEms {
		all = FindEms(expression, "{", "}")
	} else {
		all = FindStringBetweenTwoChar(expression, "{", "(")
	}
	for _, counter := range all {
		if counter == "" {
			continue
		}
		filtered = append(filtered, counter)
	}
	return filtered
}

func FindEms(stringValue string, startChar string, endChar string) []string {
	var emsSlice = make([]string, 0)
	var firstSet, counterArray []string
	if firstSet = strings.Split(stringValue, startChar); len(firstSet) < 2 {
		return emsSlice
	}
	actualString := strings.TrimSpace(firstSet[1])
	if counterArray = strings.Split(actualString, endChar); len(counterArray) < 2 {
		return emsSlice
	}
	ems := strings.TrimSpace(counterArray[0])
	if ems != "" {
		counterArray = strings.Split(ems, "=")
		emsSlice = append(emsSlice, strings.ReplaceAll(counterArray[1], "\"", ""))
	}
	return emsSlice
}

func getAllLabels(summary string) []string {
	var labels []string
	labelSlice := labelRegex.FindAllString(summary, -1)
	for _, label := range labelSlice {
		label = strings.Trim(label, "{")
		label = strings.Trim(label, "}")
		label = strings.TrimSpace(label)
		if strings.HasPrefix(label, "$labels") {
			labels = append(labels, strings.Split(label, ".")[1])
		}
	}
	return labels
}

func getEmsLabels(path string) map[string][]string {
	var (
		emsLabels = make(map[string][]string)
	)
	emsNames := make([]string, 0)
	labels := make([]string, 0)
	data, err := tree.ImportYaml(path)
	if data == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", path, err)
		return nil
	}

	for _, e := range data.GetChildS("events").GetChildren() {
		emsNames = append(emsNames, e.GetChildContentS("name"))
		labels = append(labels, parseEmsLabels(e.GetChildS("exports")))
	}

	for i := range emsNames {
		emsLabels[emsNames[i]] = strings.Split(labels[i], ",")
	}
	return emsLabels
}

func parseEmsLabels(exports *node.Node) string {
	var labels []string
	if exports != nil {
		for _, export := range exports.GetAllChildContentS() {
			_, display, _, _ := template.ParseMetric(export)
			if display != "" {
				labels = append(labels, display)
			}
		}
	}
	return strings.Join(labels, ",")
}
