package dashboard

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/data"
	"github.com/Netapp/harvest-automation/test/utils"
	log "github.com/cihub/seelog"
	"github.com/tidwall/gjson"
	"net/url"
	"time"
)

func HasValidData(query string) bool {
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d", data.PrometheusUrl, url.QueryEscape(query), timeNow)
	data, err := utils.GetResponse(queryUrl)
	if err == nil && gjson.Get(data, "status").String() == "success" {
		value := gjson.Get(data, "data.result")
		if value.Exists() && value.IsArray() && (len(value.Array()) > -1) {
			return true
		}
	}
	log.Info("Failed Query --> " + queryUrl)
	log.Info("Failed Query Response --> " + data)
	return false
}
