//go:build regression || dashboard_json
// +build regression dashboard_json

package main

import (
	//"encoding/json"
	"fmt"
	"github.com/Netapp/harvest-automation/test/dashboard"
	"github.com/Netapp/harvest-automation/test/data"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/julienroland/usg"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

var dataMap map[string]string

var fileSet []string

var counterMap map[string][]string = data.GetCounterMap()

type ResultInfo struct {
	expression  string
	counterName string
	result      bool
	reason      string
}

type DashboardJsonTestSuite struct {
	suite.Suite
}

func (suite *DashboardJsonTestSuite) SetupSuite() {
	jsonDir := utils.GetHarvestRootDir() + "/grafana/dashboards"
	log.Info().Str("jsonDir", jsonDir).Msg("Dashboard directory path")
	fileSet = GetAllJsons(jsonDir)
	if len(fileSet) == 0 {
		assert.Fail(suite.T(), "No json file found @ "+jsonDir)
	}
	log.Info().Int("fileSet", len(fileSet)).Msg("Json files")

	log.Info().Msg("Exclude map info")
	log.Info().Str("Exclude Mapping", fmt.Sprint(counterMap)).Msg("List of counter")
	log.Info().Msg("Wait until qos data is available")
	countersToCheck := []string{"qos_read_latency", "svm_nfs_throughput", "copy_manager_kb_copied"}
	for _, counterData := range countersToCheck {
		dashboard.AssertIfNotPresent(counterData)
	}
}

func (suite *DashboardJsonTestSuite) TestJsonExpression() {
	//fileSet = []string{"/Users/chinna/harvest/harvest/grafana/dashboards/cmode/harvest_dashboard_snapmirror.json"}
	var isFailed bool = false
	for _, filePath := range fileSet {
		var errorInfoList []ResultInfo
		if IsValidFile(filePath) {
			continue
		}
		log.Info().Str("JSON File", filePath).Msg("Started")
		jsonFile, err := os.Open(filePath)
		utils.PanicIfNotNil(err)
		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var allExpr []string
		value := gjson.Get(string(byteValue), "panels")
		if value.IsArray() {
			for _, record := range value.Array() {
				allExpr = append(allExpr, GetAllExpr(record)...)
				for _, targets := range record.Map()["targets"].Array() {
					allExpr = append(allExpr, targets.Map()["expr"].Str)
				}
			}
		}
		allExpr = utils.RemoveDuplicateStr(allExpr)
	EXPRESSION_FOR:
		for _, expression := range allExpr {
			counters := GetAllCounters(expression)
			for _, counter := range counters {
				if len(counter) == 0 {
					continue
				}
				// find exact counter which has no data
				for _, noDataCounter := range counterMap["NO_DATA_EXACT"] {
					if noDataCounter == counter {
						continue EXPRESSION_FOR
					}
				}
				// find exact counter which has no data
				for _, noDataCounter := range counterMap["NO_DATA_CONTAINS"] {
					if strings.Contains(counter, noDataCounter) {
						continue EXPRESSION_FOR
					}
				}

				if !HasDataInDB(counter) {
					errorInfo := ResultInfo{
						expression,
						counter,
						false,
						"No data found in the database",
					}
					errorInfoList = append(errorInfoList, errorInfo)
					continue EXPRESSION_FOR
				}
			}
			if len(counters) == 0 {
				expression = strings.ToLower(expression)
				for k, v := range data.GetCounterMapByQuery("svm_nfs_access_avg_latency") {
					key := fmt.Sprintf("$%s", k)
					expression = strings.ReplaceAll(expression, key, v)
				}
			}
			if strings.Contains(expression, "resource=\\\"cloud\\\"") ||
				strings.Contains(expression, "group_type=\\\"vol\\\"") {
				continue
			}
			queryStatus, actualExpression := ValidateExpr(expression)
			if queryStatus {
				errorInfo := ResultInfo{
					actualExpression,
					strings.Join(counters, " "),
					true,
					"",
				}
				errorInfoList = append(errorInfoList, errorInfo)
			} else {
				errorInfo := ResultInfo{
					actualExpression,
					strings.Join(counters, " "),
					false,
					"Query execution has failed",
				}
				errorInfoList = append(errorInfoList, errorInfo)
			}
		}
		for _, resultInfo := range errorInfoList {
			if resultInfo.result {
				fmt.Println(usg.Get.Tick, resultInfo.expression)
			} else {
				isFailed = true
				fmt.Println(usg.Get.Cross, fmt.Sprintf(" ERROR: %s for expr [%s]", resultInfo.reason,
					resultInfo.expression))
			}
		}
		log.Info().Msg("Completed.")
	}
	if isFailed {
		assert.Fail(suite.T(), "Test validation is failed. Pls check logs above")
	} else {
		log.Info().Msg("Everything looks good!!")
	}
}

func IsValidFile(filePath string) bool {
	ignoreList := []string{"influxdb", "7mode"}
	for _, ignoreFilePath := range ignoreList {
		if strings.Contains(filePath, ignoreFilePath) {
			return true
		}
	}
	return false
}

func GetAllExpr(record gjson.Result) []string {
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

func GetAllCounters(expression string) []string {
	return FindStringBetweenTwoChar(expression, "{", "(")
}

func ValidateExpr(expression string) (bool, string) {
	if len(expression) > 0 {
		counters := FindStringBetweenTwoChar(expression, "{", "(")
		newExpression := expression
		if len(counters) > 0 {
			for _, counter := range counters {
				newExpression = GenerateQueryWithValue(counter, newExpression)
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
	var counters []string = make([]string, 0)
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

func HasDataInDB(query string) bool {
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		data.PrometheusUrl, query, timeNow)
	data, _ := utils.GetResponse(queryUrl)
	value := gjson.Get(data, "data.result")
	return (value.Exists() && value.IsArray() && (len(value.Array()) > 0))
}

func GenerateQueryWithValue(query string, expression string) string {
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		data.PrometheusUrl, query, timeNow)
	data, _ := utils.GetResponse(queryUrl)
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
	value := gjson.Get(data, "data.result")
	if value.Exists() && value.IsArray() && (len(value.Array()) > 0) {
		metricMap := gjson.Get(value.Array()[0].String(), "metric").Map()
		for k, v := range metricMap {
			newExpression = strings.ReplaceAll(newExpression, fmt.Sprintf("$%s", strings.Title(k)), v.String())
			newExpression = strings.ReplaceAll(newExpression, fmt.Sprintf("$%s", k), v.String())
			newExpression = strings.ReplaceAll(newExpression, fmt.Sprintf("$%s", strings.ToLower(k)), v.String())
			newExpression = strings.ReplaceAll(newExpression, fmt.Sprintf("$%s", strings.ToUpper(k)), v.String())
		}
		return newExpression
	}
	return ""

}

func TestDashboardJsonSuite(t *testing.T) {
	utils.SetupLogging()
	suite.Run(t, new(DashboardJsonTestSuite))
}
