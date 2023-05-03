package dashboard

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"net/url"
)

const PrometheusURL string = "http://localhost:9090"

type GrafanaDB struct {
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
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", PrometheusURL,
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
