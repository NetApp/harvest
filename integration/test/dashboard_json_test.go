package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/dashboard"
	"github.com/Netapp/harvest-automation/test/errs"
	"github.com/Netapp/harvest-automation/test/request"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	rest     = "REST"
	zapi     = "ZAPI"
	statperf = "StatPerf"
)

var isStringAlphabetic = regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString
var fileSet []string

// zapiCounterMap are additional counters, above and beyond the ones from counterMap, which should be excluded from Zapi
var zapiCounterMap = map[string]struct{}{
	"aggr_object_store_logical_used":         {},
	"aggr_object_store_physical_used":        {},
	"cluster_schedule_labels":                {},
	"cluster_tags":                           {},
	"fpolicy_svm_failedop_notifications":     {},
	"fru_status":                             {},
	"igroup_labels":                          {},
	"net_route_labels":                       {},
	"ontaps3_labels":                         {},
	"ontaps3_logical_used_size":              {},
	"ontaps3_object_count":                   {},
	"ontaps3_policy_labels":                  {},
	"ontaps3_services_labels":                {},
	"ontaps3_size":                           {},
	"ontaps3_used_percent":                   {},
	"snapshot_policy_labels":                 {},
	"volume_arw_status":                      {},
	"volume_capacity_tier_footprint":         {},
	"volume_capacity_tier_footprint_percent": {},
	"volume_num_compress_attempts":           {},
	"volume_num_compress_fail":               {},
	"volume_snaplock_labels":                 {},
	// sar is experiencing high api time for ZapiPerf. The u2 cluster does not have fabricpool added for the collection of these counters. Remove the following once sar is capable of running ZapiPerf.
	"volume_performance_tier_footprint":         {},
	"volume_performance_tier_footprint_percent": {},
	"volume_tags": {},
	// Skip this counter in CI environments because it was introduced in version 9.15.
	// The CI currently operates with clusters running versions earlier than 9.15 for the ZAPI collector.
	"volume_total_metadata_footprint": {},
	"volume_hot_data":                 {},
}

// restCounterMap are additional counters, above and beyond the ones from counterMap, which should be excluded from Rest
var restCounterMap = map[string]struct{}{
	"aggr_snapshot_inode_used_percent": {},
	"flexcache_":                       {},
	"rw_ctx_":                          {},
	"snapshot_policy_labels":           {},
	"support_labels":                   {},
}

// excludeCounters consists of counters which should be excluded from both Zapi/Rest in CI test
var excludeCounters = []string{
	"aggr_physical_",
	"audit_log",
	"change_log",
	"cifs_session",
	"cluster_peer",
	"cluster_software_",
	"efficiency_savings",
	"ems_destination_labels",
	"ems_events",
	"ethernet_switch_port",
	"external_service_op_num_",
	"external_service_op_request_",
	"fabricpool_cloud_bin_op_latency_average",
	"fabricpool_cloud_bin_operation",
	"fcp",
	"fcvi",
	"flashcache_",
	"flashpool",
	"health_",
	"iw",
	"logical_used",
	"mav_request_",
	"metadata_exporter_count",
	"metadata_target_ping",
	"metrocluster_check_",
	"ndmp_session_",
	"nfs_clients_",
	"nfs_clients_idle_duration",
	"nic_",
	"node_cifs_",
	"node_nfs",
	"node_nvmf_ops",
	"nvme_lif",
	"nvm_mirror_",
	"ontaps3_svm_",
	"path_",
	"poller",
	"qos_workload_min_throughput_iops",
	"qtree_cifs_",
	"qtree_internal_",
	"qtree_nfs_",
	"qtree_total_",
	"quota_disk_used_pct_disk_limit",
	"quota_files_used_pct_file_limit",
	"security_login",
	"smb2_",
	"snapmirror_",
	"snapshot_labels",
	"snapshot_restore_size",
	"snapshot_create_time",
	"snapshot_volume_violation_count",
	"snapshot_volume_violation_total_size",
	"support_auto_update_labels",
	"svm_cifs_",
	"svm_ldap",
	"svm_nfs_latency_hist_bucket",
	"svm_nfs_read_latency_hist_bucket",
	"svm_nfs_write_latency_hist_bucket",
	"svm_read_total",
	"svm_vscan",
	"svm_write_total",
	"volume_top_clients_",
	"volume_top_files_",
	"volume_top_users_",
}

var flakyCounters = []string{
	"namespace",
	"flexcache",
}

var validateQueries = []string{
	`volume_read_ops{style="flexgroup"}`,
}

// Validate few counters from flexcache and system_node templates
var statPerfCounters = []string{
	"flexcache_blocks_requested_from_client",
	"flexcache_blocks_retrieved_from_origin",
	"node_cpu_busy",
	"node_total_data",
}

func TestMain(m *testing.M) {
	cmds.SetupLogging()
	os.Exit(m.Run())
}

func TestDashboardsLoad(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.Regression)
	jsonDir := cmds.GetHarvestRootDir() + "/grafana/dashboards"
	slog.Info("Dashboard directory path", slog.String("jsonDir", jsonDir))
	fileSet = GetAllJsons(jsonDir)
	if len(fileSet) == 0 {
		t.Fatalf("No json file found @ %s", jsonDir)
	}
	slog.Info("Json files", slog.Int("fileSet", len(fileSet)))
}

func validateStatPerfCounters(t *testing.T) {
	_, exists := os.LookupEnv(cmds.TestStatPerf)
	if exists {
		for _, counter := range statPerfCounters {
			query := counter + `{datacenter=~"` + statperf + `"}`
			if !hasDataInDB(query, 0) {
				t.Errorf("No records for Prometheus query: %s", query)
			}
		}
	}
}

func TestJsonExpression(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.Regression)
	cmds.SkipIfFipsSet(t)

	if len(fileSet) == 0 {
		TestDashboardsLoad(t)
	}
	var (
		zapiFails   int
		restFails   int
		exprIgnored int
		numCounters int
		zapiFlaky   int
		restFlaky   int
	)

	now := time.Now()
	// QoS counters have the longest schedule so check for them before checking for any of the other counters
	precheckCounters := []string{"qos_read_data"}
	for _, counter := range precheckCounters {
		if counterIsMissing(rest, counter, 7*time.Minute) {
			t.Fatalf("rest qos counters not found dur=%s", time.Since(now).Round(time.Millisecond).String())
		}
		if counterIsMissing(zapi, counter, 2*time.Minute) {
			t.Fatalf("zapi qos counters not found dur=%s", time.Since(now).Round(time.Millisecond).String())
		}
	}

	slog.Info(
		"Pre-check counters done",
		slog.String("dur", time.Since(now).Round(time.Millisecond).String()),
	)

	validateStatPerfCounters(t)

	for _, filePath := range fileSet {
		dashPath := shortPath(filePath)
		if shouldSkipDashboard(filePath) {
			slog.Info("Skip", slog.String("path", dashPath))
			continue
		}
		sub := time.Now()
		subCounters := 0
		subFlaky := 0
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
		allExpr = cmds.RemoveDuplicateStr(allExpr)

		for _, expression := range allExpr {
			counters := getAllCounters(expression)

			exprFlaky := false

			if len(counters) == 0 {
				exprIgnored++
				continue
			}
			for _, counter := range counters {
				numCounters++
				subCounters++
				if counterIsMissing(rest, counter, 1*time.Second) {
					if counterIsFlaky(counter) {
						subFlaky++
						restFlaky++
						exprFlaky = true
						continue
					}
					t.Errorf("%s counter=%s path=%s not in DB expr=%s", rest, counter, dashPath, expression)
					restFails++
					sumMissing++
				}
				if counterIsMissing(zapi, counter, 1*time.Second) {
					if counterIsFlaky(counter) {
						subFlaky++
						zapiFlaky++
						exprFlaky = true
						continue
					}
					t.Errorf("%s counter=%s path=%s not in DB expr=%s", zapi, counter, dashPath, expression)
					zapiFails++
					sumMissing++
				}
			}

			// if expression contains flaky counters then ignore validation
			if exprFlaky {
				continue
			}
			queryStatus, actualExpression := validateExpr(expression)
			if !queryStatus {
				t.Errorf("query validation failed counters=%s expr=%s ",
					strings.Join(counters, " "), actualExpression)
			}
		}
		slog.Info("Dashboard validation completed",
			slog.String("path", dashPath),
			slog.Int("numCounters", subCounters),
			slog.Int("missing", sumMissing),
			slog.Int("flaky", subFlaky),
			slog.String("dur", time.Since(sub).Round(time.Millisecond).String()),
		)
	}

	if restFails > 0 {
		t.Errorf("Rest validation failures=%d", restFails)
	} else {
		slog.Info("Rest Validation looks good!!")
	}

	if zapiFails > 0 {
		t.Errorf("Zapi validation failures=%d", zapiFails)
	} else {
		slog.Info("Zapi Validation looks good!!")
	}
	slog.Info("Dashboard Json validated",
		slog.String("durMs", time.Since(now).Round(time.Millisecond).String()),
		slog.Int("exprIgnored", exprIgnored),
		slog.Int("numCounters", numCounters),
		slog.Int("restMiss", restFails),
		slog.Int("restFlaky", restFlaky),
		slog.Int("zapiMiss", zapiFails),
		slog.Int("zapiFlaky", zapiFlaky),
	)

	// Add checks for queries in Prometheus
	for _, query := range validateQueries {
		if !hasDataInDB(query, 0) {
			t.Errorf("No records for Prometheus query: %s", query)
		}
	}
}

func counterIsFlaky(counter string) bool {
	for _, flakyCounter := range flakyCounters {
		if strings.Contains(counter, flakyCounter) {
			return true
		}
	}
	return false
}

func counterIsMissing(flavor string, counter string, waitFor time.Duration) bool {
	if shouldIgnoreCounter(counter, flavor) {
		return false
	}
	query := counter + `{datacenter=~"` + rest + `"}`
	if flavor == zapi {
		query = counter + `{datacenter!~"` + rest + `"}`
	}
	return !hasDataInDB(query, waitFor)
}

func shouldIgnoreCounter(counter, flavor string) bool {
	if counter == "" {
		return true
	}

	switch flavor {
	case zapi:
		if _, ok := zapiCounterMap[counter]; ok {
			return true
		}
	case rest:
		for k := range restCounterMap {
			if strings.Contains(counter, k) {
				return true
			}
		}
	default:
		for _, pattern := range excludeCounters {
			if strings.Contains(counter, pattern) {
				return true
			}
		}
	}

	return false
}

func shouldSkipDashboard(path string) bool {
	// Skip dashboards that are not cmode or use dynamic expr variables
	if !strings.Contains(path, "cmode") {
		return true
	}
	skip := []string{"nfs4storePool", "headroom", "fsa", "health", "nfsTroubleshooting", "vscan"}
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
		if counter == "" {
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
	if expression != "" {
		counters := FindStringBetweenTwoChar(expression, "{", "(")
		newExpression := expression
		if len(counters) > 0 {
			for _, counter := range counters {
				// if expression contains an excluded counter then return true
				if shouldIgnoreCounter(counter, "") {
					return true, newExpression
				}
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
		func(path string, _ os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fileInfo, err := os.Stat(path)
			errs.PanicIfNotNil(err)
			if !fileInfo.IsDir() && strings.Contains(path, ".json") {
				fileSet = append(fileSet, path)
			}
			return nil
		})
	errs.PanicIfNotNil(err)
	return fileSet
}

func FindStringBetweenTwoChar(stringValue string, startChar string, endChar string) []string {
	var counters = make([]string, 0)
	firstSet := strings.SplitSeq(stringValue, startChar)
	for actualString := range firstSet {
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

func hasDataInDB(query string, waitFor time.Duration) bool {
	now := time.Now()
	for {
		q := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d", dashboard.PrometheusURL, query, time.Now().Unix())
		response, _ := request.GetResponse(q)
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
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d", dashboard.PrometheusURL, query, timeNow)
	response, _ := request.GetResponse(queryURL)
	newExpression := expression
	/**
	We are not following standard naming convention for variables in the json
	*/
	newExpression = strings.ReplaceAll(newExpression, "$TopResources", "1")
	newExpression = strings.ReplaceAll(newExpression, "$Topresources", "1")
	newExpression = strings.ReplaceAll(newExpression, "$Aggregate", "$aggr") // dashboard has $Aggregate
	newExpression = strings.ReplaceAll(newExpression, "$Eth", "$Nic")
	newExpression = strings.ReplaceAll(newExpression, "$NFSv", "$Nfsv")
	newExpression = strings.ReplaceAll(newExpression, "$DestinationNode", "$Destination_node")
	newExpression = strings.ReplaceAll(newExpression, "$SourceSVM", "$Source_vserver")
	newExpression = strings.ReplaceAll(newExpression, "$DestinationSVM", "$Destination_vserver")
	newExpression = strings.ReplaceAll(newExpression, "$System", "$Cluster")
	value := gjson.Get(response, "data.result")
	caser := cases.Title(language.Und)
	if value.Exists() && value.IsArray() && (len(value.Array()) > 0) {
		metricMap := gjson.Get(value.Array()[0].ClonedString(), "metric").Map()
		for k, v := range metricMap {
			newExpression = strings.ReplaceAll(newExpression, "$"+caser.String(k), v.ClonedString())
			newExpression = strings.ReplaceAll(newExpression, "$"+k, v.ClonedString())
			newExpression = strings.ReplaceAll(newExpression, "$"+strings.ToLower(k), v.ClonedString())
			newExpression = strings.ReplaceAll(newExpression, "$"+strings.ToUpper(k), v.ClonedString())
		}
		return newExpression
	}
	return ""
}

func shortPath(dashPath string) string {
	splits := strings.Split(dashPath, string(filepath.Separator))
	return strings.Join(splits[len(splits)-2:], string(filepath.Separator))
}
