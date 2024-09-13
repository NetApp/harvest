package collectors

import (
	"fmt"
	template2 "github.com/netapp/harvest/v2/cmd/tools/template"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

type AlertRule struct {
	name   string
	exprs  []string
	labels []string
}

var labelRegex = regexp.MustCompile(`\{{.*?}\}`)
var isStringAlphabetic = regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString
var pluginGeneratedMetric = map[string][]string{
	"change_log": []string{"svm", "state", "type", "anti_ransomware_state", "object", "node", "location", "healthy", "volume", "style", "aggr", "status"},
}
var exceptionMetrics = []string{
	"up",
	"node_nfs_latency", // need to check why it was skipped in restperf
}

func TestParseAlertRules(t *testing.T) {
	metrics := getRestMetrics("../../")
	updateMetrics(metrics)

	alertRules := GetAllAlertRules("../../container/prometheus/", "alert_rules.yml")
	for _, alertRule := range alertRules {
		if strings.Contains(strings.Join(alertRule.exprs, ","), "_labels") {
			continue
		}
		for _, expr := range alertRule.exprs {
			if !slices.Contains(exceptionMetrics, expr) {
				metricLabels := metrics[expr]
				for _, label := range alertRule.labels {
					if !slices.Contains(metricLabels, label) {
						t.Errorf("%s is not available in %s metric", label, expr)
					}
				}
			}
		}
	}
}

func GetAllAlertRules(dir string, fileName string) []AlertRule {
	alertRules := make([]AlertRule, 0)
	alertNames := make([]string, 0)
	exprList := make([]string, 0)
	summaryList := make([]string, 0)
	alertRulesFilePath := dir + "/" + fileName
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
		alertRules = append(alertRules, AlertRule{name: alertNames[i], exprs: getAllExpressions(exprList[i]), labels: getAllLabels(summaryList[i])})
	}
	return alertRules
}

func getAllExpressions(expression string) []string {
	all := FindStringBetweenTwoChar(expression, "{", "(")
	filtered := make([]string, 0)

	for _, counter := range all {
		if counter == "" {
			continue
		}
		filtered = append(filtered, counter)
	}
	return filtered
}

func FindStringBetweenTwoChar(stringValue string, startChar string, endChar string) []string {
	var counters = make([]string, 0)
	firstSet := strings.Split(stringValue, startChar)
	for _, actualString := range firstSet {
		counterArray := strings.Split(actualString, endChar)
		switch {
		case strings.Contains(actualString, ")"): // check for inner expression such as top:
			counterArray = strings.Split(actualString, ")")
		case strings.Contains(actualString, "+"): // check for inner expression such as top:
			counterArray = strings.Split(actualString, "+")
		case strings.Contains(actualString, "/"): // check for inner expression such as top:
			counterArray = strings.Split(actualString, "/")
		case strings.Contains(actualString, ","): // check for inner expression such as top:
			counterArray = strings.Split(actualString, ",")
		}
		counter := strings.TrimSpace(counterArray[len(counterArray)-1])
		counterArray = strings.Split(counter, endChar)
		counter = strings.TrimSpace(counterArray[len(counterArray)-1])
		if _, err := strconv.Atoi(counter); err == nil {
			continue
		}
		if isStringAlphabetic(counter) && counter != "" {
			counters = append(counters, counter)
		}
	}
	return counters
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

func updateMetrics(metrics map[string][]string) {
	for k, v := range pluginGeneratedMetric {
		metrics[k] = v
	}
}

func getRestMetrics(path string) map[string][]string {
	var (
		err error
	)

	_, err = conf.LoadHarvestConfig(path + "harvest.yml")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	restCounters := processRestCounters(path)
	return restCounters
}

func processRestCounters(path string) map[string][]string {
	restPerfCounters := visitRestTemplates(path+"conf/restperf", processRestPerfCounters)
	restCounters := visitRestTemplates(path+"conf/rest", processRestConfigCounters)

	for k, v := range restPerfCounters {
		restCounters[k] = v
	}
	return restCounters
}

func visitRestTemplates(dir string, eachTemp func(path string) map[string][]string) map[string][]string {
	result := make(map[string][]string)
	err := filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			log.Fatal("failed to read directory:", err)
		}
		ext := filepath.Ext(path)
		if ext != ".yaml" {
			return nil
		}
		if strings.HasSuffix(path, "default.yaml") {
			return nil
		}
		r := eachTemp(path)
		for k, v := range r {
			result[k] = v
		}
		return nil
	})

	if err != nil {
		log.Fatal("failed to read template:", err)
		return nil
	}
	return result
}

func processRestPerfCounters(path string) map[string][]string {
	var (
		counters = make(map[string][]string)
	)
	t, err := tree.ImportYaml(path)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty\n", path)
		return nil
	}
	model, err := template2.ReadTemplate(path)
	if err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", path, err)
		return nil
	}
	noExtraMetrics := len(model.MultiplierMetrics) == 0 && len(model.PluginMetrics) == 0
	templateCounters := t.GetChildS("counters")
	if model.ExportData == "false" && noExtraMetrics {
		return nil
	}
	if templateCounters == nil {
		return nil
	}
	counterMap := make(map[string]string)
	counterMapNoPrefix := make(map[string]string)
	labels := getAllExportedLabels(t, templateCounters.GetAllChildContentS())
	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := util.ParseMetric(c)
			if m == "float" {
				counterMap[name] = model.Object + "_" + display
				counterMapNoPrefix[name] = display
				counters[name] = labels
			}
		}
	}

	return counters
}

func processRestConfigCounters(path string) map[string][]string {
	var (
		counters = make(map[string][]string)
	)
	t, err := tree.ImportYaml(path)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", path, err)
		return nil
	}

	model, err := template2.ReadTemplate(path)
	if err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", path, err)
		return nil
	}
	noExtraMetrics := len(model.MultiplierMetrics) == 0 && len(model.PluginMetrics) == 0
	templateCounters := t.GetChildS("counters")
	if model.ExportData == "false" && noExtraMetrics {
		return nil
	}

	if templateCounters != nil {
		processCounters(t, templateCounters.GetAllChildContentS(), &model, counters)
	}

	endpoints := t.GetChildS("endpoints")
	if endpoints != nil {
		for _, endpoint := range endpoints.GetChildren() {
			for _, line := range endpoint.GetChildren() {
				if line.GetNameS() == "counters" {
					processCounters(line, line.GetAllChildContentS(), &model, counters)
				}
			}
		}
	}
	return counters
}

func processCounters(t *node.Node, counterContents []string, model *template2.Model, counters map[string][]string) {
	labels := getAllExportedLabels(t, counterContents)
	for _, c := range counterContents {
		if c == "" {
			continue
		}
		_, display, m, _ := util.ParseMetric(c)
		harvestName := model.Object + "_" + display
		if m == "float" {
			counters[harvestName] = labels
		}
	}
	harvestName := model.Object + "_" + "labels"
	counters[harvestName] = labels
}

func getAllExportedLabels(t *node.Node, counterContents []string) []string {
	labels := make([]string, 0)
	if exportOptions := t.GetChildS("export_options"); exportOptions != nil {
		if iAllLabels := exportOptions.GetChildS("include_all_labels"); iAllLabels != nil {
			if iAllLabels.GetContentS() == "true" {
				for _, c := range counterContents {
					if c == "" {
						continue
					}
					if _, display, m, _ := util.ParseMetric(c); m == "key" || m == "label" {
						labels = append(labels, display)
					}
				}
				return labels
			}
		}

		if iKeys := exportOptions.GetChildS("instance_keys"); iKeys != nil {
			labels = append(labels, iKeys.GetAllChildContentS()...)
		}
	}
	return labels
}
