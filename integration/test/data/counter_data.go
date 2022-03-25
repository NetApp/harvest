package data

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/tidwall/gjson"
	"time"
)

const PrometheusUrl string = "http://localhost:9090"

func GetCounterMap() map[string][]string {
	counterMap := make(map[string][]string)
	counterMap["NO_DATA_EXACT"] = []string{"fcp_util_percent", "nic_new_status", "svm_read_total",
		"svm_write_total", "node_cifs_signed_sessions", "metadata_target_ping", "snapmirror_labels",
		"qos_detail_volume_resource_latency", "qos_detail_resource_latency"}
	counterMap["NO_DATA_CONTAINS"] = []string{"nvme_lif", "flashpool", "fcp", "snapmirror_", "node_nfs",
		"fcvi", "path_", "node_cifs_", "flashcache_", "svm_cifs_"}
	//if docker.IsDockerBasedPoller() || setup.IsMac {
	counterMap["NO_DATA_CONTAINS"] = append(counterMap["NO_DATA_CONTAINS"], "poller", "metadata_exporter_count")
	//}
	return counterMap
}

func GetCounterMapByQuery(query string) map[string]string {
	var dataMap map[string]string = make(map[string]string)
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		PrometheusUrl, query, timeNow)
	data, _ := utils.GetResponse(queryUrl)
	value := gjson.Get(data, "data.result")
	if value.Exists() && value.IsArray() && (len(value.Array()) > 0) {
		metricValue := gjson.Get(value.Array()[0].String(), "metric")
		if metricValue.Exists() {
			for counterKey, counterValue := range metricValue.Map() {
				dataMap[counterKey] = counterValue.String()
			}

		}
	}
	return dataMap
}
