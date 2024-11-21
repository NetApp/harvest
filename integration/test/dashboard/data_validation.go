package dashboard

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"net/url"
	"strings"
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
	query = strings.ReplaceAll(query, "$__range", "3h")
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", PrometheusURL,
		url.QueryEscape(query))
	resp, err := utils.GetResponse(queryURL)
	if err == nil && gjson.Get(resp, "status").ClonedString() == "success" {
		value := gjson.Get(resp, "data.result")
		if value.Exists() && value.IsArray() && (len(value.Array()) > limit) {
			return true
		}
	}
	slog.Info(
		"failed query info",
		slog.String("Query", query),
		slog.String("Query Url", queryURL),
		slog.String("Response", resp),
	)
	return false
}
