package dashboard

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/data"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"net/url"
	"testing"
	"time"
)

type GrafanaDb struct {
	ID        int64      `yaml:"apiVersion"`
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
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", data.PrometheusURL,
		url.QueryEscape(query))
	resp, err := utils.GetResponse(queryURL)
	if err == nil && gjson.Get(resp, "status").String() == "success" {
		value := gjson.Get(resp, "data.result")
		if value.Exists() && value.IsArray() && (len(value.Array()) > limit) {
			return true
		}
	}
	log.Info().Str("Query", query).Str("Query Url", queryURL).Str("Response", resp).Msg("failed query info")
	return false
}

func TestIfCounterExists(t *testing.T, restCollector string, query string) {
	checkCounter(t, fmt.Sprintf(`count(%s{datacenter="%s"})`, query, restCollector))
	checkCounter(t, fmt.Sprintf(`count(%s{datacenter!="%s"})`, query, restCollector))
}

func checkCounter(t *testing.T, query string) {
	maxCount := 10
	startCount := 1
	now := time.Now()
	for startCount < maxCount {
		queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", data.PrometheusURL,
			url.QueryEscape(query))
		resp, err := utils.GetResponse(queryURL)
		if err == nil && gjson.Get(resp, "status").String() == "success" {
			value := gjson.Get(resp, "data.result")
			if value.Exists() && value.IsArray() && (len(value.Array()) > 0) {
				metricArray := gjson.Get(value.Array()[0].String(), "value").Array()
				if len(metricArray) > 1 {
					totalRecord := metricArray[1].Int()
					if totalRecord >= 5 {
						log.Info().
							Int64("numRecs", totalRecord).
							Str("query", query).
							Str("dur", time.Since(now).Round(time.Millisecond).String()).
							Msg("Data is present")
						return
					}
				}
			}
		}
		startCount++
		time.Sleep(30 * time.Second)
	}
	log.Info().
		Str("query", query).
		Str("took", time.Since(now).String()).
		Msg("Data is NOT present")
	t.Errorf("Data for counter %s not found. Check Workload counters are uncommented", query)
}
