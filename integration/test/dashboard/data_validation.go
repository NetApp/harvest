package dashboard

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/data"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"net/url"
	"time"
)

func HasValidData(query string) bool {
	return HasMinRecord(query, 0) // to make sure that no syntax error
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
	log.Info().Str("Query", query).Msg("Failed Query --> " + queryUrl)
	log.Info().Msg("Failed Query Response --> " + data)
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
					if totalRecord > 5 {
						return
					}
				}
			}
		}
		startCount++
		time.Sleep(30 * time.Second)
	}
	panic("Data for counter " + query + " not found after 5 min")
}
