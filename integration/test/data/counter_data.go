package data

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/tidwall/gjson"
	"time"
)

const PrometheusURL string = "http://localhost:9090"

const (
	NoDataExact    = "NO_DATA_EXACT"
	NoDataContains = "NO_DATA_CONTAINS"
)

func GetCounterMap() map[string][]string {
	counterMap := make(map[string][]string)
	counterMap[NoDataExact] = []string{
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
	counterMap[NoDataContains] = []string{
		"aggr_physical_",
		"efficiency_savings",
		"ems_events",
		"fcp",
		"fcvi",
		"flashcache_",
		"flashpool",
		"health_",
		"logical_used",
		"nic_",
		"node_cifs_",
		"node_nfs",
		"node_nvmf_ops",
		"nvme_lif",
		"path_",
		"smb2_",
		"snapmirror_",
		"svm_cifs_",
		"svm_nfs_latency_hist_bucket",
		"svm_nfs_read_latency_hist_bucket",
		"svm_nfs_write_latency_hist_bucket",
	}
	//if docker.IsDockerBasedPoller() || setup.IsMac {
	counterMap[NoDataContains] = append(counterMap[NoDataContains], "poller", "metadata_exporter_count")
	//}

	// CI clusters don't have cluster peer and svm ldap/vscan metrics, security_login metrics, fabricpool metrics
	counterMap[NoDataContains] = append(
		counterMap[NoDataContains],
		"cluster_peer",
		"fabricpool_cloud_bin_op_latency_average",
		"fabricpool_cloud_bin_operation",
		"nfs_clients_idle_duration",
		"quota_disk_used_pct_disk_limit",
		"quota_files_used_pct_file_limit",
		"security_login",
		"svm_ldap",
		"svm_vscan",
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
