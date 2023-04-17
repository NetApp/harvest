package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/dashboard"
	"github.com/Netapp/harvest-automation/test/data"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	cross = "✖"
)

var restDataCollectors = []string{"Rest"}

var fileSet []string

// counterMap consists of counters which need to be excluded from both Zapi/Rest in CI test
var counterMap = data.GetCounterMap()

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
	"aggr_snapshot_inode_used_percent":                {},
	"external_service_op_request_latency":             {},
	"external_service_op_request_latency_hist_bucket": {},
}

type ResultInfo struct {
	expression  string
	counterName string
	result      bool
	reason      string
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
	waitForCollectors(t)
	var isZapiFailed = false
	var isRestFailed = false
	var restErrorInfoList []ResultInfo
	var zapiErrorInfoList []ResultInfo
	for _, filePath := range fileSet {
		if isValidFile(filePath) {
			continue
		}
		if shouldSkipDashboard(filePath) {
			log.Info().Str("path", filePath).Msg("Skip")
			continue
		}
		now := time.Now()
		jsonFile, err := os.Open(filePath)
		if err != nil {
			log.Panic().Err(err).Str("dashboard", filePath).Msg("Failed to open dashboard")
		}
		byteValue, _ := io.ReadAll(jsonFile)
		_ = jsonFile.Close()
		var allExpr []string
		value := gjson.Get(string(byteValue), "panels")
		if value.IsArray() {
			for _, record := range value.Array() {
				allExpr = append(allExpr, getAllExpr(record)...)
				for _, targets := range record.Map()["targets"].Array() {
					allExpr = append(allExpr, targets.Map()["expr"].Str)
				}
			}
		}
		allExpr = utils.RemoveDuplicateStr(allExpr)
	ExpressionFor:
		for _, expression := range allExpr {
			counters := getAllCounters(expression)
			for _, counter := range counters {
				if len(counter) == 0 {
					continue
				}
				// find exact counter which has no data
				for _, noDataCounter := range counterMap[data.NoDataExact] {
					if noDataCounter == counter {
						continue ExpressionFor
					}
				}
				// find exact counter which has no data
				for _, noDataCounter := range counterMap[data.NoDataContains] {
					if strings.Contains(counter, noDataCounter) {
						continue ExpressionFor
					}
				}

				//Test for Rest
				query := counter + "{datacenter=~\"" + strings.Join(restDataCollectors, "|") + "\"}"

				if !hasDataInDB(query) {
					if _, ok := restCounterMap[counter]; ok {
						continue ExpressionFor
					}
					errorInfo := ResultInfo{
						expression,
						counter,
						false,
						"No data found in the database",
					}
					restErrorInfoList = append(restErrorInfoList, errorInfo)
				}

				//Test for Zapi
				query = counter + "{datacenter!~\"" + strings.Join(restDataCollectors, "|") + "\"}"
				if !hasDataInDB(query) {
					if _, ok := zapiCounterMap[counter]; ok {
						continue ExpressionFor
					}
					errorInfo := ResultInfo{
						expression,
						counter,
						false,
						"No data found in the database",
					}
					zapiErrorInfoList = append(zapiErrorInfoList, errorInfo)
					continue ExpressionFor
				}
			}
			if len(counters) == 0 {
				expression = strings.ToLower(expression)
				for k, v := range data.GetCounterMapByQuery("svm_nfs_access_avg_latency") {
					key := fmt.Sprintf("$%s", k)
					expression = strings.ReplaceAll(expression, key, v)
				}
			}
			if strings.Contains(expression, `resource=\"cloud\"`) ||
				strings.Contains(expression, `group_type=\"vol\"`) {
				continue
			}
			queryStatus, actualExpression := validateExpr(expression)
			if queryStatus {
				errorInfo := ResultInfo{
					actualExpression,
					strings.Join(counters, " "),
					true,
					"",
				}
				zapiErrorInfoList = append(zapiErrorInfoList, errorInfo)
			} else {
				errorInfo := ResultInfo{
					actualExpression,
					strings.Join(counters, " "),
					false,
					"Query execution has failed",
				}
				zapiErrorInfoList = append(zapiErrorInfoList, errorInfo)
			}
		}
		log.Info().
			Str("path", filePath).
			Str("dur", time.Since(now).Round(time.Millisecond).String()).
			Msg("Dashboard validation completed")
	}

	for _, resultInfo := range zapiErrorInfoList {
		if !resultInfo.result {
			isZapiFailed = true
			fmt.Println(cross, fmt.Sprintf(" Zapi Collector ERROR: %s for expr [%s]", resultInfo.reason,
				resultInfo.expression))
		}
	}

	for _, resultInfo := range restErrorInfoList {
		if !resultInfo.result {
			isRestFailed = true
			fmt.Println(cross, fmt.Sprintf(" Rest Collector ERROR: %s for expr [%s]", resultInfo.reason,
				resultInfo.expression))
		}
	}

	if isRestFailed {
		t.Errorf("Rest Test validation is failed. Pls check logs above. Count of Missing Rest counters %d", len(restErrorInfoList))
	} else {
		log.Info().Msg("Rest Validation looks good!!")
	}

	// Fail if either Rest or Zapi collectors have failures
	if isZapiFailed {
		t.Errorf("Zapi Test validation is failed. Pls check logs above. Count of Missing Zapi counters %d", len(zapiErrorInfoList))
	} else {
		log.Info().Msg("Zapi Validation looks good!!")
	}
}

func waitForCollectors(t *testing.T) {
	log.Info().Msg("Exclude map info")
	log.Info().Str("Exclude Mapping", fmt.Sprint(counterMap)).Msg("List of counter")
	log.Info().Msg("Wait for data to be collected")
	countersToCheck := []string{
		"copy_manager_kb_copied",
		"qos_read_latency",
		"qos_volume_read_data",
		"qos_volume_read_latency",
		"qos_volume_read_ops",
		"qos_volume_sequential_reads",
		"qos_volume_sequential_writes",
		"qos_volume_write_data",
		"qos_volume_write_latency",
		"qos_volume_write_ops",
		"svm_nfs_throughput",
	}
	for _, counterData := range countersToCheck {
		dashboard.TestIfCounterExists(t, restDataCollectors[0], counterData)
	}
}

func shouldSkipDashboard(path string) bool {
	// Ignore dashboards that use dynamic expr variables
	skip := []string{"nfs4storePool", "headroom", "tenant", "fsa", "overview"}
	for _, s := range skip {
		if strings.Contains(path, s) {
			return true
		}
	}
	return false
}

func isValidFile(filePath string) bool {
	ignoreList := []string{"influxdb", "7mode"}
	for _, ignoreFilePath := range ignoreList {
		if strings.Contains(filePath, ignoreFilePath) {
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
	return FindStringBetweenTwoChar(expression, "{", "(")
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

func hasDataInDB(query string) bool {
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		data.PrometheusURL, query, timeNow)
	response, _ := utils.GetResponse(queryUrl)
	value := gjson.Get(response, "data.result")
	return value.Exists() && value.IsArray() && (len(value.Array()) > 0)
}

func generateQueryWithValue(query string, expression string) string {
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		data.PrometheusURL, query, timeNow)
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
