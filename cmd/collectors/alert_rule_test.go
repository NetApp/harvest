package collectors

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/tools/generate"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
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
	metrics, _ := generate.GeneratedMetrics("../../", "harvest.yml")
	for pluginMetric, pluginLabels := range pluginGeneratedMetric {
		metrics[pluginMetric] = generate.Counter{Name: pluginMetric, Labels: pluginLabels}
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

func GetAllAlertRules(dir string, fileName string, isEms bool) []AlertRule {
	alertRules := make([]AlertRule, 0)
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
				if a.GetNameS() == "alert" {
					alertname := a.GetContentS()
					alertNames = append(alertNames, alertname)
				}
				if a.GetNameS() == "expr" {
					alertexp := a.GetContentS()
					exprList = append(exprList, alertexp)
				}
				if a.GetNameS() == "annotations" {
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
	firstSet := strings.Split(stringValue, startChar)
	actualString := strings.TrimSpace(firstSet[1])
	counterArray := strings.Split(actualString, endChar)
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
			name, _, _, _ := util.ParseMetric(export)
			if strings.HasPrefix(name, "parameters") {
				labels = append(labels, strings.Split(name, ".")[1])
			}
		}
	}
	return strings.Join(labels, ",")
}
