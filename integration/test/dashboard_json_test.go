package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/dashboard"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	rest = "REST"
	zapi = "ZAPI"
)

var restDataCollectors = []string{"Rest"}

var fileSet []string

// zapiCounterMap are additional counters, above and beyond the ones from counterMap, which should be excluded from Zapi
var zapiCounterMap = map[string]struct{}{
	"net_route_labels":          {},
	"ontaps3_labels":            {},
	"ontaps3_logical_used_size": {},
	"ontaps3_size":              {},
	"ontaps3_object_count":      {},
	"ontaps3_used_percent":      {},
	"ontaps3_services_labels":   {},
	"ontaps3_policy_labels":     {},
}

// restCounterMap are additional counters, above and beyond the ones from counterMap, which should be excluded from Rest
var restCounterMap = map[string]struct{}{
	"aggr_snapshot_inode_used_percent": {},
}

// excludeCounters consists of counters which should be excluded from both Zapi/Rest in CI test
var excludeCounters = []string{
	"aggr_physical_",
	"cluster_peer",
	"efficiency_savings",
	"ems_events",
	"fabricpool_cloud_bin_op_latency_average",
	"fabricpool_cloud_bin_operation",
	"fcp",
	"fcvi",
	"flashcache_",
	"flashpool",
	"health_",
	"logical_used",
	"metadata_exporter_count",
	"metadata_target_ping",
	"nfs_clients_idle_duration",
	"nic_",
	"node_cifs_",
	"node_nfs",
	"node_nvmf_ops",
	"nvme_lif",
	"path_",
	"poller",
	"qos_detail_resource_latency",
	"qos_detail_volume_resource_latency",
	"quota_disk_used_pct_disk_limit",
	"quota_files_used_pct_file_limit",
	"security_login",
	"smb2_",
	"snapmirror_",
	"svm_cifs_",
	"svm_ldap",
	"svm_nfs_latency_hist_bucket",
	"svm_nfs_read_latency_hist_bucket",
	"svm_nfs_write_latency_hist_bucket",
	"svm_read_total",
	"svm_vscan",
	"svm_write_total",
}

func TestMain(m *testing.M) {
	utils.SetupLogging()
	os.Exit(m.Run())
}

func TestDashboardsLoad(t *testing.T) {
	utils.SkipIfMissing(t, utils.Regression)
	jsonDir := utils.GetHarvestRootDir() + "/grafana/dashboards"
	log.Info().Str("jsonDir", jsonDir).Msg("Dashboard directory path")
	fileSet = GetAllJsons(jsonDir)
	if len(fileSet) == 0 {
		t.Fatalf("No json file found @ %s", jsonDir)
	}
	log.Info().Int("fileSet", len(fileSet)).Msg("Json files")
}

func TestJsonExpression(t *testing.T) {
	utils.SkipIfMissing(t, utils.Regression)
	if len(fileSet) == 0 {
		TestDashboardsLoad(t)
	}
	var (
		zapiFails   int
		restFails   int
		exprIgnored int
		numCounters int
	)

	now := time.Now()
	// QoS counters have the longest schedule so check for them before checking for any of the other counters
	precheckCounters := []string{"qos_read_data"}
	for _, counter := range precheckCounters {
		if counterIsMissing(rest, counter, 6*time.Minute) {
			t.Fatalf("rest qos counters not found dur=%s", time.Since(now).Round(time.Millisecond).String())
		}
		if counterIsMissing(zapi, counter, 1*time.Minute) {
			t.Fatalf("zapi qos counters not found dur=%s", time.Since(now).Round(time.Millisecond).String())
		}
	}

	log.Info().
		Str("dur", time.Since(now).Round(time.Millisecond).String()).
		Msg("Pre-check counters done")

	for _, filePath := range fileSet {
		dashPath := shortPath(filePath)
		if shouldSkipDashboard(filePath) {
			log.Info().Str("path", dashPath).Msg("Skip")
			continue
		}
		sub := time.Now()
		subCounters := 0
		sumMissing := 0
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
			counters := getAllCounters(expression)

			if len(counters) == 0 {
				exprIgnored++
				continue
			}
			for _, counter := range counters {
				numCounters++
				subCounters++
				if counterIsMissing(rest, counter, 1*time.Second) {
					t.Errorf("%s counter=%s path=%s not in DB expr=%s", rest, counter, dashPath, expression)
					restFails++
					sumMissing++
				}
				if counterIsMissing(zapi, counter, 1*time.Second) {
					t.Errorf("%s counter=%s path=%s not in DB expr=%s", zapi, counter, dashPath, expression)
					zapiFails++
					sumMissing++
				}
			}

			queryStatus, actualExpression := validateExpr(expression)
			if !queryStatus {
				t.Errorf("query validation failed counters=%s expr=%s ",
					strings.Join(counters, " "), actualExpression)
			}
		}
		log.Info().
			Str("path", dashPath).
			Int("numCounters", subCounters).
			Int("missingCounters", sumMissing).
			Str("dur", time.Since(sub).Round(time.Millisecond).String()).
			Msg("Dashboard validation completed")
	}

	if restFails > 0 {
		t.Errorf("Rest validation failures=%d", restFails)
	} else {
		log.Info().Msg("Rest Validation looks good!!")
	}

	if zapiFails > 0 {
		t.Errorf("Zapi validation failures=%d", zapiFails)
	} else {
		log.Info().Msg("Zapi Validation looks good!!")
	}
	log.Info().
		Str("durMs", time.Since(now).Round(time.Millisecond).String()).
		Int("exprIgnored", exprIgnored).
		Int("numCounters", numCounters).
		Int("restMiss", restFails).
		Int("zapiMiss", zapiFails).
		Msg("Dashboard Json validated")
}

func counterIsMissing(flavor string, counter string, waitFor time.Duration) bool {
	if shouldIgnoreCounter(counter, flavor) {
		return false
	}
	query := counter + `{datacenter=~"` + strings.Join(restDataCollectors, "|") + `"}`
	if flavor == zapi {
		query = counter + `{datacenter!~"` + strings.Join(restDataCollectors, "|") + `"}`
	}
	return !hasDataInDB(query, waitFor)
}

func shouldIgnoreCounter(counter string, flavor string) bool {
	if len(counter) == 0 {
		return true
	}
	if flavor == zapi {
		if _, ok := zapiCounterMap[counter]; ok {
			return true
		}
	} else if flavor == rest {
		if _, ok := restCounterMap[counter]; ok {
			return true
		}
	}

	return false
}

func shouldSkipDashboard(path string) bool {
	// Skip dashboards that are not cmode or use dynamic expr variables
	if !strings.Contains(path, "cmode") {
		return true
	}
	skip := []string{"nfs4storePool", "headroom", "fsa"}
	for _, s := range skip {
		if strings.Contains(path, s) {
			return true
		}
	}
	return false
}

func getAllExpr(record gjson.Result) []string {
	var expressionArray []string
	panelInfo := record.Map()["panels"]
	if panelInfo.Exists() {
		if panelInfo.IsArray() {
			for _, eachRow := range panelInfo.Array() {
				for _, targets := range eachRow.Map()["targets"].Array() {
					expressionArray = append(expressionArray, targets.Map()["expr"].Str)
				}
			}
		}
		for _, targets := range panelInfo.Map()["targets"].Array() {
			expressionArray = append(expressionArray, targets.Map()["expr"].Str)
		}
	}
	return expressionArray
}

func getAllCounters(expression string) []string {
	all := FindStringBetweenTwoChar(expression, "{", "(")
	var filtered []string

allLoop:
	for _, counter := range all {
		if len(counter) == 0 {
			continue
		}
		// check if this counter should be ignored
		for _, pattern := range excludeCounters {
			if strings.Contains(counter, pattern) {
				continue allLoop
			}
		}
		filtered = append(filtered, counter)
	}
	return filtered
}

func validateExpr(expression string) (bool, string) {
	if len(expression) > 0 {
		counters := FindStringBetweenTwoChar(expression, "{", "(")
		newExpression := expression
		if len(counters) > 0 {
			for _, counter := range counters {
				newExpression = generateQueryWithValue(counter, newExpression)
			}
		}
		if newExpression == "" {
			return false, expression
		}
		return dashboard.HasValidData(newExpression), newExpression

	}
	return false, expression
}

func GetAllJsons(dir string) []string {
	var fileSet []string
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fileInfo, err := os.Stat(path)
			utils.PanicIfNotNil(err)
			if !fileInfo.IsDir() && strings.Contains(path, ".json") {
				fileSet = append(fileSet, path)
			}
			return nil
		})
	utils.PanicIfNotNil(err)
	return fileSet
}

func FindStringBetweenTwoChar(stringValue string, startChar string, endChar string) []string {
	var counters = make([]string, 0)
	var isStringAlphabetic = regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString
	firstSet := strings.Split(stringValue, startChar)
	for _, actualString := range firstSet {
		counterArray := strings.Split(actualString, endChar)
		if strings.Contains(actualString, "+") { //check for inner expression such as top
			counterArray = strings.Split(actualString, "+")
		} else if strings.Contains(actualString, "/") { //check for inner expression such as top
			counterArray = strings.Split(actualString, "/")
		} else if strings.Contains(actualString, ",") { //check for inner expression such as top
			counterArray = strings.Split(actualString, ",")
		}
		counter := strings.TrimSpace(counterArray[len(counterArray)-1])
		counterArray = strings.Split(counter, endChar)
		counter = strings.TrimSpace(counterArray[len(counterArray)-1])
		if _, err := strconv.Atoi(counter); err == nil {
			continue
		}
		if isStringAlphabetic(counter) && len(counter) > 0 {
			counters = append(counters, counter)
		}
	}
	return counters
}

func hasDataInDB(query string, waitFor time.Duration) bool {
	now := time.Now()
	for {
		q := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d", dashboard.PrometheusURL, query, time.Now().Unix())
		response, _ := utils.GetResponse(q)
		value := gjson.Get(response, "data.result")
		if len(value.Array()) > 0 {
			return true
		}
		if time.Since(now) > waitFor {
			return false
		}
		time.Sleep(1 * time.Second)
	}
}

func generateQueryWithValue(query string, expression string) string {
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d", dashboard.PrometheusURL, query, timeNow)
	response, _ := utils.GetResponse(queryUrl)
	newExpression := expression
	/**
	We are not following standard naming convention for variables in the json
	*/
	newExpression = strings.ReplaceAll(newExpression, "$TopResources", "1")
	newExpression = strings.ReplaceAll(newExpression, "$Topresources", "1")
	newExpression = strings.ReplaceAll(newExpression, "$Aggregate", "$aggr") //dashboard has $Aggregate
	newExpression = strings.ReplaceAll(newExpression, "$Eth", "$Nic")
	newExpression = strings.ReplaceAll(newExpression, "$NFSv", "$Nfsv")
	newExpression = strings.ReplaceAll(newExpression, "$DestinationNode", "$Destination_node")
	//newExpression = strings.ReplaceAll(newExpression, "$SourceNode", "$Source_node")
	//newExpression = strings.ReplaceAll(newExpression, "$Source_node", "$Source_node")
	newExpression = strings.ReplaceAll(newExpression, "$SourceSVM", "$Source_vserver")
	newExpression = strings.ReplaceAll(newExpression, "$DestinationSVM", "$Destination_vserver")
	newExpression = strings.ReplaceAll(newExpression, "$System", "$Cluster")
	value := gjson.Get(response, "data.result")
	caser := cases.Title(language.Und)
	if value.Exists() && value.IsArray() && (len(value.Array()) > 0) {
		metricMap := gjson.Get(value.Array()[0].String(), "metric").Map()
		for k, v := range metricMap {
			newExpression = strings.ReplaceAll(newExpression, fmt.Sprintf("$%s", caser.String(k)), v.String())
			newExpression = strings.ReplaceAll(newExpression, fmt.Sprintf("$%s", k), v.String())
			newExpression = strings.ReplaceAll(newExpression, fmt.Sprintf("$%s", strings.ToLower(k)), v.String())
			newExpression = strings.ReplaceAll(newExpression, fmt.Sprintf("$%s", strings.ToUpper(k)), v.String())
		}
		return newExpression
	}
	return ""
}

func shortPath(dashPath string) string {
	splits := strings.Split(dashPath, string(filepath.Separator))
	return strings.Join(splits[len(splits)-2:], string(filepath.Separator))
}
