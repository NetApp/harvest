package data

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/tidwall/gjson"
	"time"
)

const PrometheusURL string = "http://localhost:9090"

func GetCounterMap() map[string][]string {
	counterMap := make(map[string][]string)
	counterMap["NO_DATA_EXACT"] = []string{
		"fcp_util_percent",
		"metadata_target_ping",
		"nic_new_status",
		"node_cifs_signed_sessions",
		"qos_detail_resource_latency",
		"qos_detail_volume_resource_latency",
		"snapmirror_labels",
		"svm_read_total",
		"svm_write_total",
	}
	counterMap["NO_DATA_CONTAINS"] = []string{
		"logical_used",
		"efficiency_savings",
		"aggr_physical_",
		"fcp",
		"fcvi",
		"flashcache_",
		"flashpool",
		"nic_",
		"node_cifs_",
		"node_nfs",
		"node_nvmf_ops",
		"nvme_lif",
		"path_",
		"snapmirror_",
		"svm_cifs_",
		"svm_nfs_latency_hist_bucket",
		"svm_nfs_read_latency_hist_bucket",
		"svm_nfs_write_latency_hist_bucket",
		"smb2_",
		"health_",
	}
	//if docker.IsDockerBasedPoller() || setup.IsMac {
	counterMap["NO_DATA_CONTAINS"] = append(counterMap["NO_DATA_CONTAINS"], "poller", "metadata_exporter_count")
	//}

	// CI clusters don't have cluster peer and svm ldap/vscan metrics, security_login metrics, fabricpool metrics
	counterMap["NO_DATA_CONTAINS"] = append(
		counterMap["NO_DATA_CONTAINS"],
		"cluster_peer",
		"svm_ldap",
		"svm_vscan",
		"security_login",
		"quota_files_used_pct_file_limit",
		"quota_disk_used_pct_disk_limit",
		"nfs_clients_idle_duration",
		"fabricpool_cloud_bin_operation",
		"fabricpool_cloud_bin_op_latency_average",
	)
	return counterMap
}

func GetCounterMapByQuery(query string) map[string]string {
	var dataMap = make(map[string]string)
	timeNow := time.Now().Unix()
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		PrometheusURL, query, timeNow)
	data, _ := utils.GetResponse(queryURL)
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
