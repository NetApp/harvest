package dashboard

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/data"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/url"
	"time"
)

type GrafanaDb struct {
	Id        int64      `yaml:"apiVersion"`
	Providers []Provider `yaml:"providers"`
}

type Provider struct {
	Name       string `yaml:"name"`
	FolderName string `yaml:"folder"`
}

func HasValidData(query string) bool {
	return HasMinRecord(query, -1) // to make sure that there are no syntax error
}

func HasMinRecord(query string, limit int) bool {
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s", data.PrometheusUrl,
		url.QueryEscape(query))
	data, err := utils.GetResponse(queryUrl)
	if err == nil && gjson.Get(data, "status").String() == "success" {
		value := gjson.Get(data, "data.result")
		if value.Exists() && value.IsArray() && (len(value.Array()) > limit) {
			return true
		}
	}
	log.Info().Str("Query", query).Str("Query Url", queryUrl).Str("Response", data).Msg("failed query info")
	return false
}

func AssertIfNotPresent(query string) {
	maxCount := 10
	startCount := 1
	query = fmt.Sprintf("count(%s)", query)
	log.Info().Msg("Checking whether data is present or not for counter " + query)
	for startCount < maxCount {
		queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s", data.PrometheusUrl,
			url.QueryEscape(query))
		data, err := utils.GetResponse(queryUrl)
		if err == nil && gjson.Get(data, "status").String() == "success" {
			value := gjson.Get(data, "data.result")
			if value.Exists() && value.IsArray() && (len(value.Array()) > 0) {
				metricArray := gjson.Get(value.Array()[0].String(), "value").Array()
				fmt.Println(metricArray)
				if len(metricArray) > 1 {
					totalRecord := metricArray[1].Int()
					log.Info().Int64("Total Record", totalRecord).Msg("")
					if totalRecord >= 5 {
						time.Sleep(2 * time.Minute)
						return
					}
				}
			}
		}
		startCount++
		time.Sleep(30 * time.Second)
	}
	panic("Data for counter " + query + " not found after 8 min. Check Workload counters are uncommented from conf/zapiperf/default.yml")
}

func GetFolderNameFromYml(filePath string) string {
	var yamlData GrafanaDb
	yamlFile, err := ioutil.ReadFile(filePath)
	utils.PanicIfNotNil(err)
	err = yaml.Unmarshal(yamlFile, &yamlData)
	utils.PanicIfNotNil(err)
	if len(yamlData.Providers) == 0 {
		log.Error().Msg("No providers found at " + filePath)
		panic("No providers found at " + filePath)
	}
	return yamlData.Providers[0].FolderName
}
