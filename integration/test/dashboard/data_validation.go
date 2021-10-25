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
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d", data.PrometheusUrl,
		url.QueryEscape(query), timeNow)
	data, err := utils.GetResponse(queryUrl)
	if err == nil && gjson.Get(data, "status").String() == "success" {
		value := gjson.Get(data, "data.result")
		if value.Exists() && value.IsArray() && (len(value.Array()) > -1) {
			return true
		}
	}
	log.Info().Msg("Failed Query --> " + queryUrl)
	log.Info().Msg("Failed Query Response --> " + data)
	return false
}

func AssertIfNotPresent(query string) {
	maxCount := 10
	startCount := 1
	for startCount < maxCount {
		if HasValidData(query) {
			log.Info().Int("Iteration", startCount).Msg("Qos counters are present")
			return
		}
		startCount++
		time.Sleep(30 * time.Second)
	}
	panic("No Qos data found after 5 min")
}
