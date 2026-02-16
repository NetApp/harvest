package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/netapp/harvest/v2/cmd/collectors/keyperf"
	"github.com/netapp/harvest/v2/cmd/collectors/statperf"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/cmd/tools/rest/clirequestbuilder"
	template2 "github.com/netapp/harvest/v2/cmd/tools/template"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/set"
	template3 "github.com/netapp/harvest/v2/pkg/template"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	tw "github.com/netapp/harvest/v2/third_party/olekukonko/tablewriter"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"io"
	"log"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"text/template"
	"time"
)

const (
	keyPerfAPI = "KeyPerf"
)

type Options struct {
	Poller      string
	Loglevel    int
	Image       string
	FilesdPath  string
	ShowPorts   bool
	OutputPath  string
	CertDir     string
	PromPort    int
	GrafanaPort int
	Mounts      []string
	ConfigPath  string
	ConfPath    string
	PromURL     string
}

type CounterMetaData struct {
	Date           string
	OntapVersion   string
	SGVersion      string
	CiscoVersion   string
	ESeriesVersion string
}

type CounterTemplate struct {
	Counters        []Counter
	CounterMetaData CounterMetaData
}

type Counters struct {
	C []Counter `yaml:"counters"`
}

type Counter struct {
	Object      string      `yaml:"-"`
	Name        string      `yaml:"Name"`
	Description string      `yaml:"Description"`
	APIs        []MetricDef `yaml:"APIs"`
	Panels      []PanelDef  `yaml:"Panels"`
	Labels      []string    `yaml:"Labels"`
}

type MetricDef struct {
	API            string `yaml:"API"`
	Endpoint       string `yaml:"Endpoint"`
	ONTAPCounter   string `yaml:"ONTAPCounter"`
	CiscoCounter   string `yaml:"CiscoCounter"`
	SGCounter      string `yaml:"SGCounter"`
	ESeriesCounter string `yaml:"ESeriesCounter"`
	Template       string `yaml:"Template"`
	Unit           string `yaml:"Unit"`
	Type           string `yaml:"Type"`
	BaseCounter    string `yaml:"BaseCounter"`
}

type PanelDef struct {
	Dashboard string `yaml:"Dashboard"`
	Row       string `yaml:"Row"`
	Type      string `yaml:"Type"`
	Panel     string `yaml:"Panel"`
	PanelLink string `yaml:"PanelLink"`
}

type PanelData struct {
	Panels []PanelDef
}

// Regex to match NFS version and operation
var reRemove = regexp.MustCompile(`NFSv\d+\.\d+`)

var (
	replacer         = strings.NewReplacer("\n", "", ":", "")
	objectSwaggerMap = map[string]string{
		"aggr":               "xc_aggregate",
		"environment_sensor": "sensors",
		"fcp":                "fc_port",
		"flexcache":          "volume",
		"lif":                "ip_interface",
		"namespace":          "nvme_namespace",
		"net_port":           "xc_broadcast_domain",
		"ontaps3":            "xc_s3_bucket",
		"security_ssh":       "cluster_ssh_server",
		"svm_cifs":           "cifs_service",
		"svm_nfs":            "nfs_service",
		"volume":             "xc_volume",
	}
	swaggerBytes         []byte
	excludePerfTemplates = map[string]struct{}{
		"volume_node.yaml":            {}, // Similar metrics node_volume_* are generated via KeyPerf volume.yaml
		"workload_detail.yaml":        {},
		"workload_detail_volume.yaml": {},
	}
	excludeRestPerfTemplates = map[string]struct{}{
		"volume.yaml": {}, // Volume performance metrics now collected via KeyPerf
	}
	excludeCounters = map[string]struct{}{
		"latency_histogram":       {},
		"nfs4_latency_hist":       {},
		"nfs41_latency_hist":      {},
		"nfsv3_latency_hist":      {},
		"read_latency_hist":       {},
		"read_latency_histogram":  {},
		"total.latency_histogram": {},
		"write_latency_hist":      {},
		"write_latency_histogram": {},
	}

	// Special handling perf objects
	specialPerfObjects = map[string]bool{
		"node_nfs": true,
		"svm_nfs":  true,
	}

	excludeDocumentedRestMetrics = []string{
		"audit_log",
		"aggr_hybrid_disk_count",
		"availability_zone_",
		"change_log",
		"cifs_session_idle_duration",
		"cluster_space_available",
		"cluster_software",
		"ems_events",
		"ethernet_switch_port_",
		"export_rule_labels",
		"fcp_util_percent",
		"fcvi_",
		"flashpool_",
		"health_",
		"igroup_labels",
		"iw_",
		"mav_request_",
		"mediator_labels",
		"metrocluster_",
		"ndmp_session",
		"net_connection_labels",
		"nfs_clients_idle_duration",
		"nfs_diag_",
		"node_cifs_",
		"nvme_lif_",
		"nvmf_",
		"ontaps3_svm_",
		"path_",
		"poller_cpu_percent",
		"qtree_",
		"smb2_",
		"snapshot_labels",
		"snapshot_restore_size",
		"snapshot_create_time",
		"snapshot_volume_violation_count",
		"snapshot_volume_violation_total_size",
		"storage_unit_",
		"svm_cifs_",
		"svm_ontaps3_svm_",
		"svm_vscan_",
		"token_",
		"volume_top_clients",
		"volume_top_files",
		"vscan_",
	}

	excludeDocumentedZapiMetrics = []string{
		"ems_events",
		"external_service_",
		"fabricpool_",
		"flexcache_",
		"fpolicy_svm_failedop_notifications",
		"netstat_",
		"nvm_mirror_",
		"quota_disk_used_pct_threshold",
		"snapshot_volume_violation_count",
		"snapshot_volume_violation_total_size",
	}

	// Exclude extra metrics for REST
	excludeNotDocumentedRestMetrics = []string{
		"ALERTS",
		"cluster_space_",
		"flexcache_",
		"hist_",
		"igroup_",
		"storage_unit_",
		"volume_aggr_labels",
		"volume_arw_status",
	}

	// Exclude extra metrics for ZAPI
	excludeNotDocumentedZapiMetrics = []string{
		"ALERTS",
		"hist_",
		"security_",
		"svm_ldap",
		"volume_aggr_labels",
	}

	// include StatPerf Templates
	includeStatPerfTemplates = map[string]struct{}{
		"flexcache.yaml":   {},
		"system_node.yaml": {},
	}

	// Excludes these Rest gaps from logs
	excludeLogRestCounters = []string{
		"external_service_op_",
		"fabricpool_average_latency",
		"fabricpool_get_throughput_bytes",
		"fabricpool_put_throughput_bytes",
		"fabricpool_stats",
		"fabricpool_throughput_ops",
		"iw_",
		"netstat_",
		"nvmf_rdma_port_",
		"nvmf_tcp_port_",
		"ontaps3_svm_",
		"smb2_",
	}

	knownDescriptionGaps = map[string]struct{}{
		"availability_zone_space_available":             {},
		"availability_zone_space_physical_used":         {},
		"availability_zone_space_physical_used_percent": {},
		"availability_zone_space_size":                  {},
		"ontaps3_object_count":                          {},
		"security_certificate_expiry_time":              {},
		"storage_unit_space_efficiency_ratio":           {},
		"storage_unit_space_size":                       {},
		"storage_unit_space_used":                       {},
		"volume_capacity_tier_footprint":                {},
		"volume_capacity_tier_footprint_percent":        {},
		"volume_num_compress_attempts":                  {},
		"volume_num_compress_fail":                      {},
		"volume_performance_tier_footprint":             {},
		"volume_performance_tier_footprint_percent":     {},
	}

	knownMappingGaps = map[string]struct{}{
		"aggr_snapshot_inode_used_percent": {},
		"aggr_space_reserved":              {},
		"flexcache_":                       {},
		"fpolicy_":                         {},
		"quota_disk_used_pct_threshold":    {},
		"rw_ctx_":                          {},
		"security_audit_destination_port":  {},
		"storage_unit_":                    {},
		"wafl_reads_from_pmem":             {},
		"node_volume_nfs_":                 {},
		"nvm_mirror_":                      {},
		"volume_nfs_":                      {},
		"svm_vol_nfs":                      {},
	}

	knownMappingGapsSG = map[string]struct{}{
		"storagegrid_node_cpu_utilization_percentage":                         {},
		"storagegrid_private_load_balancer_storage_request_body_bytes_bucket": {},
		"storagegrid_private_load_balancer_storage_request_count":             {},
		"storagegrid_private_load_balancer_storage_request_time":              {},
		"storagegrid_private_load_balancer_storage_rx_bytes":                  {},
		"storagegrid_private_load_balancer_storage_tx_bytes":                  {},
	}
)

func BuildMetrics(dir, configPath, pollerName string, opts *Options, metricsPanelMap map[string]PanelData) (map[string]Counter, conf.Remote) {
	var (
		poller         *conf.Poller
		err            error
		restClient     *rest.Client
		zapiClient     *zapi.Client
		harvestYmlPath string
	)

	if opts == nil {
		restCounters := ProcessRestCounters(dir, nil, metricsPanelMap)
		zapiCounters := ProcessZapiCounters(dir, nil, metricsPanelMap)
		counters := MergeCounters(restCounters, zapiCounters)
		counters = ProcessExternalCounters(dir, counters, metricsPanelMap)
		return counters, conf.Remote{}
	}

	if opts.ConfigPath != "" {
		harvestYmlPath = filepath.Join(dir, opts.ConfigPath)
	} else {
		harvestYmlPath = filepath.Join(dir, configPath)
	}
	_, err = conf.LoadHarvestConfig(harvestYmlPath)
	if err != nil {
		LogErrAndExit(err)
	}

	if poller, _, err = rest.GetPollerAndAddr(pollerName); err != nil {
		LogErrAndExit(err)
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	credentials := auth.NewCredentials(poller, slog.Default())
	if restClient, err = rest.New(poller, timeout, credentials); err != nil {
		fmt.Printf("error creating new client %+v\n", err)
		os.Exit(1)
	}
	if err = restClient.Init(2, conf.Remote{}); err != nil {
		fmt.Printf("error init rest client %+v\n", err)
		os.Exit(1)
	}

	if zapiClient, err = zapi.New(poller, credentials); err != nil {
		fmt.Printf("error creating new client %+v\n", err)
		os.Exit(1)
	}

	swaggerBytes = readSwaggerJSON(opts)
	restCounters := ProcessRestCounters(dir, restClient, metricsPanelMap)
	zapiCounters := ProcessZapiCounters(dir, zapiClient, metricsPanelMap)
	counters := MergeCounters(restCounters, zapiCounters)
	counters = ProcessExternalCounters(dir, counters, metricsPanelMap)

	if opts.PromURL != "" {
		prometheusRest, prometheusZapi, err := FetchAndCategorizePrometheusMetrics(opts.PromURL)
		if err != nil {
			LogErrAndExit(err)
		}

		documentedRest, documentedZapi := CategorizeCounters(counters)

		if err := ValidateMetrics(documentedRest, documentedZapi, prometheusRest, prometheusZapi); err != nil {
			LogErrAndExit(err)
		}
	}

	for k, counter := range counters {
		// Generically handle latency metrics to specify microseconds in the description if the unit is microseconds
		// and the description does not already mention microseconds
		if strings.Contains(counter.Name, "latency") {
			for _, metricDef := range counter.APIs {
				if metricDef.Unit == "microsec" && !strings.Contains(counter.Description, "microsec") {
					counter.Description = strings.Replace(counter.Description, "latency", "latency in microseconds", 1)
					counters[k] = counter
					break
				}
			}
		}

		// Generically handle throughput metrics to specify bytes per second in the description
		// does not already contain it
		if strings.HasSuffix(counter.Name, "_data") {
			for _, metricDef := range counter.APIs {
				if metricDef.Unit == "b_per_sec" && !strings.Contains(counter.Description, "per second") {
					counter.Description = strings.Replace(counter.Description, "operations", "operations in bytes per seconds", 1)
					counters[k] = counter
					break
				}
			}
		}
	}

	return counters, restClient.Remote()
}

func GenerateCounters(dir string, counters map[string]Counter, collectorName string, metricsPanelMap map[string]PanelData) map[string]Counter {
	dat, err := os.ReadFile(filepath.Join(dir, "cmd", "tools", "generate", collectorName+"_counter.yaml"))
	if err != nil {
		fmt.Printf("error while reading file %v", err)
		return nil
	}
	var c Counters

	err = yaml.Unmarshal(dat, &c)
	if err != nil {
		fmt.Printf("error while parsing file %v", err)
		return nil
	}

	for k, m := range metricsPanelMap {
		if !strings.HasPrefix(k, collectorName) {
			continue
		}
		if _, ok := counters[k]; !ok {
			counters[k] = Counter{Name: k, Panels: m.Panels}
		}
	}

	for _, v := range c.C {
		if v1, ok := counters[v.Name]; !ok {
			v.Panels = metricsPanelMap[v.Name].Panels
			counters[v.Name] = v
		} else {
			if v.Description != "" {
				v1.Description = v.Description
			}
			if len(v.APIs) > 0 {
				v1.APIs = v.APIs
			}
			counters[v.Name] = v1
		}
	}
	return counters
}

// ProcessRestCounters parse rest and restperf templates
func ProcessRestCounters(dir string, client *rest.Client, metricsPanelMap map[string]PanelData) map[string]Counter {

	restPerfCounters := visitRestTemplates(filepath.Join(dir, "conf", "restperf"), client, func(path string, client *rest.Client) map[string]Counter {
		if _, ok := excludePerfTemplates[filepath.Base(path)]; ok {
			return nil
		}
		if _, ok := excludeRestPerfTemplates[filepath.Base(path)]; ok {
			return nil
		}
		return processRestPerfCounters(path, client, metricsPanelMap)
	})

	restCounters := visitRestTemplates(filepath.Join(dir, "conf", "rest"), client, func(path string, client *rest.Client) map[string]Counter { // revive:disable-line:unused-parameter
		return processRestConfigCounters(path, "REST", metricsPanelMap)
	})

	keyPerfCounters := visitRestTemplates(filepath.Join(dir, "conf", "keyperf"), client, func(path string, client *rest.Client) map[string]Counter { // revive:disable-line:unused-parameter
		return processRestConfigCounters(path, keyPerfAPI, metricsPanelMap)
	})

	statPerfCounters := visitRestTemplates(filepath.Join(dir, "conf", "statperf"), client, func(path string, client *rest.Client) map[string]Counter {
		if _, ok := includeStatPerfTemplates[filepath.Base(path)]; ok {
			return processStatPerfCounters(path, client, metricsPanelMap)
		}
		return nil
	})
	maps.Copy(restCounters, restPerfCounters)

	keyPerfKeys := slices.Sorted(maps.Keys(keyPerfCounters))
	for _, k := range keyPerfKeys {
		if strings.Contains(k, "timestamp") || strings.Contains(k, "labels") {
			continue
		}
		v := keyPerfCounters[k]
		if v1, ok := restCounters[k]; !ok {
			restCounters[k] = v
		} else {
			v1.APIs = append(v1.APIs, v.APIs...)
			restCounters[k] = v1
		}
	}

	statPerfKeys := slices.Sorted(maps.Keys(statPerfCounters))
	for _, k := range statPerfKeys {
		if strings.Contains(k, "labels") {
			continue
		}
		v := statPerfCounters[k]
		if v1, ok := restCounters[k]; !ok {
			restCounters[k] = v
		} else {
			v1.APIs = append(v1.APIs, v.APIs...)
			restCounters[k] = v1
		}
	}
	return restCounters
}

// ProcessZapiCounters parse zapi and zapiperf templates
func ProcessZapiCounters(dir string, client *zapi.Client, metricsPanelMap map[string]PanelData) map[string]Counter {
	zapiCounters := visitZapiTemplates(filepath.Join(dir, "conf", "zapi", "cdot"), client, func(path string, client *zapi.Client) map[string]Counter { // revive:disable-line:unused-parameter
		return processZapiConfigCounters(path, metricsPanelMap)
	})
	zapiPerfCounters := visitZapiTemplates(filepath.Join(dir, "conf", "zapiperf", "cdot"), client, func(path string, client *zapi.Client) map[string]Counter {
		if _, ok := excludePerfTemplates[filepath.Base(path)]; ok {
			return nil
		}
		return processZAPIPerfCounters(path, client, metricsPanelMap)
	})

	maps.Copy(zapiCounters, zapiPerfCounters)
	return zapiCounters
}

func MergeCounters(restCounters map[string]Counter, zapiCounters map[string]Counter) map[string]Counter {
	// handle special counters
	restKeys := slices.Sorted(maps.Keys(restCounters))
	for _, k := range restKeys {
		v := restCounters[k]
		found := strings.Contains(k, "#")
		if found {
			if v1, ok := restCounters[v.Name]; !ok {
				v.Description = reRemove.ReplaceAllString(v.Description, "")
				// Remove extra spaces from the description
				v.Description = strings.Join(strings.Fields(v.Description), " ")
				restCounters[v.Name] = v
			} else {
				v1.APIs = append(v1.APIs, v.APIs...)
				restCounters[v.Name] = v1
			}
			delete(restCounters, k)
		}
	}

	zapiKeys := slices.Sorted(maps.Keys(zapiCounters))
	for _, k := range zapiKeys {
		v := zapiCounters[k]
		found := strings.Contains(k, "#")
		if found {
			if v1, ok := zapiCounters[v.Name]; !ok {
				v.Description = reRemove.ReplaceAllString(v.Description, "")
				// Remove extra spaces from the description
				v.Description = strings.Join(strings.Fields(v.Description), " ")
				zapiCounters[v.Name] = v
			} else {
				v1.APIs = append(v1.APIs, v.APIs...)
				zapiCounters[v.Name] = v1
			}
			delete(zapiCounters, k)
		}
	}

	// special keys are deleted hence sort again
	zapiKeys = slices.Sorted(maps.Keys(zapiCounters))
	for _, k := range zapiKeys {
		v := zapiCounters[k]
		if v1, ok := restCounters[k]; ok {
			v1.APIs = append(v1.APIs, v.APIs...)
			restCounters[k] = v1
		} else {
			zapiDef := v.APIs[0]
			if zapiDef.ONTAPCounter == "instance_name" || zapiDef.ONTAPCounter == "instance_uuid" {
				continue
			}
			co := Counter{
				Object:      v.Object,
				Name:        v.Name,
				Description: v.Description,
				APIs:        []MetricDef{zapiDef},
				Panels:      v.Panels,
			}
			restCounters[v.Name] = co
		}
	}
	return restCounters
}

func ProcessExternalCounters(dir string, counters map[string]Counter, metricsPanelMap map[string]PanelData) map[string]Counter {
	dat, err := os.ReadFile(filepath.Join(dir, "cmd", "tools", "generate", "counter.yaml"))
	if err != nil {
		fmt.Printf("error while reading file %v", err)
		return nil
	}
	var c Counters

	err = yaml.Unmarshal(dat, &c)
	if err != nil {
		fmt.Printf("error while parsing file %v", err)
		return nil
	}

	// Check that there are not duplicates in counter.yaml
	duplicates := make(map[string]int)
	for _, v := range c.C {
		duplicates[v.Name]++
	}

	dupsFound := false
	for k, count := range duplicates {
		if count > 1 {
			fmt.Printf("error: duplicate counter definition found for counter '%s' in counter.yaml file\n", k)
			dupsFound = true
		}
	}

	if dupsFound {
		os.Exit(1)
	}

	for _, v := range c.C {
		if v1, ok := counters[v.Name]; !ok {
			v.Panels = metricsPanelMap[v.Name].Panels
			counters[v.Name] = v
		} else {
			if v.Description != "" {
				v1.Description = v.Description
			}
			for _, m := range v.APIs {
				indices := findAPI(v1.APIs, m)
				if len(indices) == 0 {
					v1.APIs = append(v1.APIs, m)
				} else {
					for _, index := range indices {
						r := &v1.APIs[index]
						if m.ONTAPCounter != "" {
							r.ONTAPCounter = m.ONTAPCounter
						}
						if m.Template != "" {
							r.Template = m.Template
						}
						if m.Endpoint != "" {
							r.Endpoint = m.Endpoint
						}
						if m.Type != "" {
							r.Type = m.Type
						}
						if m.Unit != "" {
							r.Unit = m.Unit
						}
						if m.BaseCounter != "" {
							r.BaseCounter = m.BaseCounter
						}
					}
				}
			}
			counters[v.Name] = v1
		}
	}
	return counters
}

func visitRestTemplates(dir string, client *rest.Client, eachTemp func(path string, client *rest.Client) map[string]Counter) map[string]Counter {
	result := make(map[string]Counter)
	err := filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			log.Fatal("failed to read directory:", err)
		}
		ext := filepath.Ext(path)
		if ext != ".yaml" {
			return nil
		}
		if strings.HasSuffix(path, "default.yaml") || strings.HasSuffix(path, "static_counter_definitions.yaml") {
			return nil
		}
		r := eachTemp(path, client)
		maps.Copy(result, r)
		return nil
	})

	if err != nil {
		log.Fatal("failed to read template:", err)
		return nil
	}
	return result
}

func visitZapiTemplates(dir string, client *zapi.Client, eachTemp func(path string, client *zapi.Client) map[string]Counter) map[string]Counter {
	result := make(map[string]Counter)
	err := filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			log.Fatal("failed to read directory:", err)
		}
		ext := filepath.Ext(path)
		if ext != ".yaml" {
			return nil
		}

		r := eachTemp(path, client)
		maps.Copy(result, r)
		return nil
	})

	if err != nil {
		log.Fatal("failed to read template:", err)
		return nil
	}
	return result
}

// processRestConfigCounters process Rest config templates
func processRestConfigCounters(path string, api string, metricsPanelMap map[string]PanelData) map[string]Counter {
	var (
		counters         = make(map[string]Counter)
		isInstanceLabels bool
	)
	var metricLabels []string
	var labels []string
	t, err := tree.ImportYaml(path)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", path, err)
		return nil
	}

	model, err := template2.ReadTemplate(path)
	if err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", path, err)
		return nil
	}
	noExtraMetrics := len(model.MultiplierMetrics) == 0 && len(model.PluginMetrics) == 0
	templateCounters := t.GetChildS("counters")
	if model.ExportData == "false" && noExtraMetrics {
		return nil
	}

	if templateCounters != nil {
		metricLabels, labels, isInstanceLabels = getAllExportedLabels(t, templateCounters.GetAllChildContentS())
		processCounters(templateCounters.GetAllChildContentS(), &model, path, model.Query, counters, metricLabels, api, metricsPanelMap)
		if isInstanceLabels {
			// This is for object_labels metrics
			harvestName := model.Object + "_" + "labels"
			description := "This metric provides information about " + model.Name
			counters[harvestName] = Counter{
				Name:        harvestName,
				Description: description,
				APIs: []MetricDef{
					{
						API:          "REST",
						Endpoint:     model.Query,
						Template:     path,
						ONTAPCounter: "Harvest generated",
					},
				},
				Panels: metricsPanelMap[harvestName].Panels,
				Labels: labels,
			}
		}
	}

	endpoints := t.GetChildS("endpoints")
	if endpoints != nil {
		for _, endpoint := range endpoints.GetChildren() {
			var query string
			for _, line := range endpoint.GetChildren() {
				if line.GetNameS() == "query" {
					query = line.GetContentS()
				}
				if line.GetNameS() == "counters" {
					processCounters(line.GetAllChildContentS(), &model, path, query, counters, metricLabels, api, metricsPanelMap)
				}
			}
		}
	}

	// If the template has any PluginMetrics, add them
	for _, metric := range model.PluginMetrics {
		co := Counter{
			Object: model.Object,
			Name:   model.Object + "_" + metric.Name,
			APIs: []MetricDef{
				{
					API:          api,
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: metric.Source,
				},
			},
			Panels: metricsPanelMap[model.Object+"_"+metric.Name].Panels,
		}
		if model.ExportData != "false" {
			counters[co.Name] = co
		}
	}

	if api == keyPerfAPI {
		// handling for templates with common object names
		if specialPerfObjects[model.Object] {
			return specialHandlingPerfCounters(counters, model)
		}
	}

	return counters
}

func processCounters(counterContents []string, model *template2.Model, path, query string, counters map[string]Counter, metricLabels []string, api string, metricsPanelMap map[string]PanelData) {
	var (
		staticCounterDef keyperf.ObjectCounters
		err              error
		defLocation      string
	)
	if api == keyPerfAPI {
		logger := logging.Get()
		// CLI conf/keyperf/9.15.0/aggr.yaml
		// CI  ../../conf/keyperf/9.15.0/volume.yaml
		defLocation = filepath.Join(filepath.Dir(filepath.Dir(path)), "static_counter_definitions.yaml")

		staticCounterDef, err = keyperf.LoadStaticCounterDefinitions(model.Object, defLocation, logger)
		if err != nil {
			fmt.Printf("Failed to load static counter definitions=%s\n", err)
		}
	}

	for _, c := range counterContents {
		if c == "" {
			continue
		}
		var co Counter
		name, display, m, _ := template3.ParseMetric(c)
		if _, ok := excludeCounters[name]; ok {
			continue
		}
		description := searchDescriptionSwagger(model.Object, name)
		harvestName := model.Object + "_" + display
		if m == "float" {
			if api == keyPerfAPI {
				var (
					unit        string
					counterType string
					denominator string
				)
				switch {
				case strings.Contains(name, "latency"):
					counterType = "average"
					unit = "microsec"
					denominator = model.Object + "_" + strings.Replace(name, "latency", "iops", 1)
				case strings.Contains(name, "iops"):
					counterType = "rate"
					unit = "per_sec"
				case strings.Contains(name, "throughput"):
					counterType = "rate"
					unit = "b_per_sec"
				case strings.Contains(name, "timestamp"):
					counterType = "delta"
					unit = "sec"
				default:
					// look up metric in staticCounterDef
					if counterDef, exists := staticCounterDef.CounterDefinitions[name]; exists {
						counterType = counterDef.Type
						unit = counterDef.BaseCounter
					}
				}

				co = Counter{
					Object:      model.Object,
					Name:        harvestName,
					Description: description,
					APIs: []MetricDef{
						{
							API:          api,
							Endpoint:     model.Query,
							Template:     path,
							ONTAPCounter: name,
							Unit:         unit,
							Type:         counterType,
							BaseCounter:  denominator,
						},
					},
					Panels: metricsPanelMap[harvestName].Panels,
					Labels: metricLabels,
				}
			} else {
				co = Counter{
					Object:      model.Object,
					Name:        harvestName,
					Description: description,
					APIs: []MetricDef{
						{
							API:          api,
							Endpoint:     query,
							Template:     path,
							ONTAPCounter: name,
						},
					},
					Panels: metricsPanelMap[harvestName].Panels,
					Labels: metricLabels,
				}
			}
			counters[harvestName] = co

			// If the template has any MultiplierMetrics, add them
			for _, metric := range model.MultiplierMetrics {
				mc := co
				addAggregatedCounter(&mc, metric, harvestName, display, metricsPanelMap)
				counters[mc.Name] = mc
			}
		}
	}
}

// processZAPIPerfCounters process ZapiPerf counters
func processZAPIPerfCounters(path string, client *zapi.Client, metricsPanelMap map[string]PanelData) map[string]Counter {
	var (
		counters                = make(map[string]Counter)
		request, response       *node.Node
		zapiUnitMap             = make(map[string]string)
		zapiTypeMap             = make(map[string]string)
		zapiDescMap             = make(map[string]string)
		zapiBaseCounterMap      = make(map[string]string)
		zapiDeprecateCounterMap = make(map[string]string)
	)
	t, err := tree.ImportYaml(path)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty\n", path)
		return nil
	}
	model, err := template2.ReadTemplate(path)
	if err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", path, err)
		return nil
	}

	noExtraMetrics := len(model.MultiplierMetrics) == 0 && len(model.PluginMetrics) == 0
	templateCounters := t.GetChildS("counters")
	override := t.GetChildS("override")

	if model.ExportData == "false" && noExtraMetrics {
		return nil
	}

	if templateCounters == nil {
		return nil
	}

	if client != nil {
		// build request
		request = node.NewXMLS("perf-object-counter-list-info")
		request.NewChildS("objectname", model.Query)

		if err = client.BuildRequest(request); err != nil {
			fmt.Printf("error while building request %+v\n", err)
			return nil
		}

		if response, err = client.Invoke(""); err != nil {
			fmt.Printf("error while invoking api %+v\n", err)
			return nil
		}
		// fetch counter elements
		if elems := response.GetChildS("counters"); elems != nil && len(elems.GetChildren()) != 0 {
			for _, counter := range elems.GetChildren() {
				name := counter.GetChildContentS("name")
				if name == "" {
					continue
				}
				ty := counter.GetChildContentS("properties")
				if override != nil {
					oty := override.GetChildContentS(name)
					if oty != "" {
						ty = oty
					}
				}

				zapiUnitMap[name] = counter.GetChildContentS("unit")
				zapiDescMap[name] = updateDescription(counter.GetChildContentS("desc"))
				zapiTypeMap[name] = ty
				zapiBaseCounterMap[name] = counter.GetChildContentS("base-counter")

				if counter.GetChildContentS("is-deprecated") == "true" {
					if r := counter.GetChildContentS("replaced-by"); r != "" {
						zapiDeprecateCounterMap[name] = r
					}
				}
			}
		}
	}

	metricLabels, labels, isInstanceLabels := getAllExportedLabels(t, templateCounters.GetAllChildContentS())
	if isInstanceLabels {
		// This is for object_labels metrics
		harvestName := model.Object + "_" + "labels"
		description := "This metric provides information about " + model.Name
		counters[harvestName] = Counter{
			Name:        harvestName,
			Description: description,
			APIs: []MetricDef{
				{
					API:          "ZAPI",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: "Harvest generated",
				},
			},
			Panels: metricsPanelMap[harvestName].Panels,
			Labels: labels,
		}
	}
	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := template3.ParseMetric(c)
			if after, ok := strings.CutPrefix(display, model.Object); ok {
				display = after
				display = strings.TrimPrefix(display, "_")
			}
			harvestName := model.Object + "_" + display
			if m == "float" {
				if _, ok := excludeCounters[name]; ok {
					continue
				}
				if zapiTypeMap[name] != "string" {
					description := zapiDescMap[name]
					if strings.Contains(path, "volume.yaml") && model.Object == "volume" {
						if description != "" {
							description += " "
						}
						description += "(Note: This is applicable only for ONTAP 9.9 and below. Harvest uses KeyPerf collector for ONTAP 9.10 onwards.)"
					}
					co := Counter{
						Object:      model.Object,
						Name:        harvestName,
						Description: description,
						APIs: []MetricDef{
							{
								API:          "ZapiPerf",
								Endpoint:     "perf-object-get-instances" + " " + model.Query,
								Template:     path,
								ONTAPCounter: name,
								Unit:         zapiUnitMap[name],
								Type:         zapiTypeMap[name],
								BaseCounter:  zapiBaseCounterMap[name],
							},
						},
						Panels: metricsPanelMap[harvestName].Panels,
						Labels: metricLabels,
					}
					if model.ExportData != "false" {
						counters[harvestName] = co
					}

					// handle deprecate counters
					if rName, ok := zapiDeprecateCounterMap[name]; ok {
						hName := model.Object + "_" + rName
						ro := Counter{
							Object:      model.Object,
							Name:        hName,
							Description: zapiDescMap[rName],
							APIs: []MetricDef{
								{
									API:          "ZapiPerf",
									Endpoint:     "perf-object-get-instances" + " " + model.Query,
									Template:     path,
									ONTAPCounter: rName,
									Unit:         zapiUnitMap[rName],
									Type:         zapiTypeMap[rName],
									BaseCounter:  zapiBaseCounterMap[rName],
								},
							},
							Panels: metricsPanelMap[hName].Panels,
						}
						if model.ExportData != "false" {
							counters[hName] = ro
						}
					}

					// If the template has any MultiplierMetrics, add them
					for _, metric := range model.MultiplierMetrics {
						mc := co
						addAggregatedCounter(&mc, metric, harvestName, display, metricsPanelMap)
						counters[mc.Name] = mc
					}
				}
			}
		}
	}

	// If the template has any PluginMetrics, add them
	for _, metric := range model.PluginMetrics {
		co := Counter{
			Object: model.Object,
			Name:   model.Object + "_" + metric.Name,
			APIs: []MetricDef{
				{
					API:          "ZapiPerf",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: metric.Source,
				},
			},
			Panels: metricsPanelMap[model.Object+"_"+metric.Name].Panels,
		}
		counters[co.Name] = co
	}
	// handling for templates with common object names
	if specialPerfObjects[model.Object] {
		return specialHandlingPerfCounters(counters, model)
	}
	return counters
}

func processZapiConfigCounters(path string, metricsPanelMap map[string]PanelData) map[string]Counter {
	var (
		counters         = make(map[string]Counter)
		isInstanceLabels bool
	)
	var metricLabels []string
	var labels []string
	t, err := tree.ImportYaml(path)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty\n", path)
		return nil
	}
	model, err := template2.ReadTemplate(path)
	if err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", path, err)
		return nil
	}
	noExtraMetrics := len(model.MultiplierMetrics) == 0 && len(model.PluginMetrics) == 0
	templateCounters := t.GetChildS("counters")
	if model.ExportData == "false" && noExtraMetrics {
		return nil
	}
	if templateCounters == nil {
		return nil
	}

	zc := make(map[string]string)
	metricLabels, labels, isInstanceLabels = getAllExportedLabels(t, templateCounters.GetAllChildContentS())
	if isInstanceLabels {
		// This is for object_labels metrics
		harvestName := model.Object + "_" + "labels"
		description := "This metric provides information about " + model.Name
		counters[harvestName] = Counter{
			Name:        harvestName,
			Description: description,
			APIs: []MetricDef{
				{
					API:          "ZAPI",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: "Harvest generated",
				},
			},
			Panels: metricsPanelMap[harvestName].Panels,
			Labels: labels,
		}
	}
	for _, c := range templateCounters.GetChildren() {
		parseZapiCounters(c, []string{}, model.Object, zc)
	}

	for k, v := range zc {
		if _, ok := excludeCounters[k]; ok {
			continue
		}
		co := Counter{
			Object: model.Object,
			Name:   k,
			APIs: []MetricDef{
				{
					API:          "ZAPI",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: v,
				},
			},
			Panels: metricsPanelMap[k].Panels,
			Labels: metricLabels,
		}
		if model.ExportData != "false" {
			counters[k] = co
		}

		// If the template has any MultiplierMetrics, add them
		for _, metric := range model.MultiplierMetrics {
			mc := co
			addAggregatedCounter(&mc, metric, co.Name, model.Object, metricsPanelMap)
			counters[mc.Name] = mc
		}
	}

	// If the template has any PluginMetrics, add them
	for _, metric := range model.PluginMetrics {
		co := Counter{
			Object: model.Object,
			Name:   model.Object + "_" + metric.Name,
			APIs: []MetricDef{
				{
					API:          "ZAPI",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: metric.Source,
				},
			},
			Panels: metricsPanelMap[model.Object+"_"+metric.Name].Panels,
		}
		counters[co.Name] = co
	}
	return counters
}

func processRestPerfCounters(path string, client *rest.Client, metricsPanelMap map[string]PanelData) map[string]Counter {
	var (
		records       []gjson.Result
		counterSchema gjson.Result
		counters      = make(map[string]Counter)
	)
	t, err := tree.ImportYaml(path)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty\n", path)
		return nil
	}
	model, err := template2.ReadTemplate(path)
	if err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", path, err)
		return nil
	}
	noExtraMetrics := len(model.MultiplierMetrics) == 0 && len(model.PluginMetrics) == 0
	templateCounters := t.GetChildS("counters")
	override := t.GetChildS("override")
	if model.ExportData == "false" && noExtraMetrics {
		return nil
	}
	if templateCounters == nil {
		return nil
	}
	counterMap := make(map[string]string)
	counterMapNoPrefix := make(map[string]string)
	metricLabels, labels, isInstanceLabels := getAllExportedLabels(t, templateCounters.GetAllChildContentS())
	if isInstanceLabels {
		description := "This metric provides information about " + model.Name
		// This is for object_labels metrics
		harvestName := model.Object + "_" + "labels"
		counters[harvestName] = Counter{
			Name:        harvestName,
			Description: description,
			APIs: []MetricDef{
				{
					API:          "RestPerf",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: "Harvest generated",
				},
			},
			Panels: metricsPanelMap[harvestName].Panels,
			Labels: labels,
		}
	}
	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := template3.ParseMetric(c)
			if m == "float" {
				counterMap[name] = model.Object + "_" + display
				counterMapNoPrefix[name] = display
			}
		}
	}

	if client != nil {
		href := rest.NewHrefBuilder().
			APIPath(model.Query).
			Build()
		records, err = rest.FetchAll(client, href)
		if err != nil {
			fmt.Printf("error while invoking api %+v\n", err)
			return nil
		}
		firstRecord := records[0]
		if firstRecord.Exists() {
			counterSchema = firstRecord.Get("counter_schemas")
		} else {
			return nil
		}
		counterSchema.ForEach(func(_, r gjson.Result) bool {
			if !r.IsObject() {
				return true
			}
			ontapCounterName := r.Get("name").ClonedString()
			if _, ok := excludeCounters[ontapCounterName]; ok {
				return true
			}

			description := r.Get("description").ClonedString()
			ty := r.Get("type").ClonedString()
			if override != nil {
				oty := override.GetChildContentS(ontapCounterName)
				if oty != "" {
					ty = oty
				}
			}
			if v, ok := counterMap[ontapCounterName]; ok {
				if ty == "string" {
					return true
				}
				c := Counter{
					Object:      model.Object,
					Name:        v,
					Description: description,
					APIs: []MetricDef{
						{
							API:          "RestPerf",
							Endpoint:     model.Query,
							Template:     path,
							ONTAPCounter: ontapCounterName,
							Unit:         r.Get("unit").ClonedString(),
							Type:         ty,
							BaseCounter:  r.Get("denominator.name").ClonedString(),
						},
					},
					Panels: metricsPanelMap[v].Panels,
					Labels: metricLabels,
				}
				if model.ExportData != "false" {
					counters[c.Name] = c
				}

				// If the template has any MultiplierMetrics, add them
				for _, metric := range model.MultiplierMetrics {
					mc := c
					addAggregatedCounter(&mc, metric, v, counterMapNoPrefix[ontapCounterName], metricsPanelMap)
					counters[mc.Name] = mc
				}
			}
			return true
		})
	} else {
		//
		for ontapName, counterName := range counterMap {
			c := Counter{
				Object: model.Object,
				Name:   counterName,
				APIs: []MetricDef{
					{
						API:          "RestPerf",
						Endpoint:     model.Query,
						Template:     path,
						ONTAPCounter: ontapName,
					},
				},
				Panels: metricsPanelMap[counterName].Panels,
				Labels: metricLabels,
			}
			if model.ExportData != "false" {
				counters[c.Name] = c
			}

			// If the template has any MultiplierMetrics, add them
			for _, metric := range model.MultiplierMetrics {
				mc := c
				addAggregatedCounter(&mc, metric, counterName, counterMapNoPrefix[ontapName], metricsPanelMap)
				counters[mc.Name] = mc
			}
		}
	}

	// If the template has any PluginMetrics, add them
	for _, metric := range model.PluginMetrics {
		co := Counter{
			Object: model.Object,
			Name:   model.Object + "_" + metric.Name,
			APIs: []MetricDef{
				{
					API:          "RestPerf",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: metric.Source,
				},
			},
			Panels: metricsPanelMap[model.Object+"_"+metric.Name].Panels,
		}
		counters[co.Name] = co
	}
	// handling for templates with common object names/metric name
	if specialPerfObjects[model.Object] {
		return specialHandlingPerfCounters(counters, model)
	}
	return counters
}

func processStatPerfCounters(path string, client *rest.Client, metricsPanelMap map[string]PanelData) map[string]Counter {
	var (
		records    []gjson.Result
		counters   = make(map[string]Counter)
		cliCommand []byte
	)
	t, err := tree.ImportYaml(path)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty\n", path)
		return nil
	}
	model, err := template2.ReadTemplate(path)
	if err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", path, err)
		return nil
	}
	noExtraMetrics := len(model.MultiplierMetrics) == 0 && len(model.PluginMetrics) == 0
	templateCounters := t.GetChildS("counters")
	override := t.GetChildS("override")
	if model.ExportData == "false" && noExtraMetrics {
		return nil
	}
	if templateCounters == nil {
		return nil
	}
	counterMap := make(map[string]string)
	counterMapNoPrefix := make(map[string]string)
	metricLabels, labels, isInstanceLabels := getAllExportedLabels(t, templateCounters.GetAllChildContentS())
	if isInstanceLabels {
		description := "This metric provides information about " + model.Name
		// This is for object_labels metrics
		harvestName := model.Object + "_" + "labels"
		counters[harvestName] = Counter{
			Name:        harvestName,
			Description: description,
			APIs: []MetricDef{
				{
					API:          "StatPerf",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: "Harvest generated",
				},
			},
			Panels: metricsPanelMap[harvestName].Panels,
			Labels: labels,
		}
	}
	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := template3.ParseMetric(c)
			if m == "float" {
				counterMap[name] = model.Object + "_" + display
				counterMapNoPrefix[name] = display
			}
		}
	}

	if client != nil {
		cliCommand, err = clirequestbuilder.New().
			BaseSet(statperf.GetCounterInstanceBaseSet()).
			Query("statistics catalog counter show").
			Object(model.Query).
			Fields([]string{"counter", "base-counter", "properties", "type", "is-deprecated", "replaced-by", "unit", "description"}).
			Build()
		if err != nil {
			fmt.Printf("error while build clicommand %+v\n", err)
			return nil
		}
		records, err = rest.FetchPost(client, "api/private/cli", cliCommand)
		if err != nil {
			fmt.Printf("error while invoking api %+v\n", err)
			return nil
		}
		firstRecord := records[0]
		fr := firstRecord.ClonedString()
		if fr == "" {
			fmt.Printf("no data found for query %s template %s", model.Query, path)
			return nil
		}
		s := &statperf.StatPerf{}
		pCounters, err := s.ParseCounters(fr)
		if err != nil {
			fmt.Printf("error while parsing records for query %s template %s: %+v\n", model.Query, path, err)
			return nil
		}
		for _, p := range pCounters {
			ontapCounterName := statperf.NormalizeCounterValue(p.Name)
			description := statperf.NormalizeCounterValue(p.Description)
			ty := p.Type
			if override != nil {
				oty := override.GetChildContentS(ontapCounterName)
				if oty != "" {
					ty = oty
				}
			}
			if v, ok := counterMap[ontapCounterName]; ok {
				if ty == "string" {
					continue
				}
				c := Counter{
					Object:      model.Object,
					Name:        v,
					Description: description,
					APIs: []MetricDef{
						{
							API:          "StatPerf",
							Endpoint:     model.Query,
							Template:     path,
							ONTAPCounter: ontapCounterName,
							Unit:         statperf.NormalizeCounterValue(p.Unit),
							Type:         statperf.NormalizeCounterValue(ty),
							BaseCounter:  statperf.NormalizeCounterValue(p.BaseCounter),
						},
					},
					Panels: metricsPanelMap[v].Panels,
					Labels: metricLabels,
				}
				if model.ExportData != "false" {
					counters[c.Name] = c
				}

				// If the template has any MultiplierMetrics, add them
				for _, metric := range model.MultiplierMetrics {
					mc := c
					addAggregatedCounter(&mc, metric, v, counterMapNoPrefix[ontapCounterName], metricsPanelMap)
					counters[mc.Name] = mc
				}
			}
		}
	} else {
		for ontapName, counterName := range counterMap {
			c := Counter{
				Object: model.Object,
				Name:   counterName,
				APIs: []MetricDef{
					{
						API:          "StatPerf",
						Endpoint:     model.Query,
						Template:     path,
						ONTAPCounter: ontapName,
					},
				},
				Panels: metricsPanelMap[counterName].Panels,
				Labels: metricLabels,
			}
			if model.ExportData != "false" {
				counters[c.Name] = c
			}

			// If the template has any MultiplierMetrics, add them
			for _, metric := range model.MultiplierMetrics {
				mc := c
				addAggregatedCounter(&mc, metric, counterName, counterMapNoPrefix[ontapName], metricsPanelMap)
				counters[mc.Name] = mc
			}
		}
	}

	// If the template has any PluginMetrics, add them
	for _, metric := range model.PluginMetrics {
		co := Counter{
			Object: model.Object,
			Name:   model.Object + "_" + metric.Name,
			APIs: []MetricDef{
				{
					API:          "StatPerf",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: metric.Source,
				},
			},
			Panels: metricsPanelMap[model.Object+"_"+metric.Name].Panels,
		}
		counters[co.Name] = co
	}
	return counters
}

func specialHandlingPerfCounters(counters map[string]Counter, model template2.Model) map[string]Counter {
	// handling for templates with common object names
	modifiedCounters := make(map[string]Counter)
	for originalKey, value := range counters {
		modifiedKey := model.Name + "#" + originalKey
		modifiedCounters[modifiedKey] = value
	}
	return modifiedCounters
}

func addAggregatedCounter(c *Counter, metric plugin.DerivedMetric, withPrefix string, noPrefix string, metricsPanelMap map[string]PanelData) {
	if !strings.HasSuffix(c.Description, ".") {
		c.Description += "."
	}

	if metric.IsMax {
		c.Name = metric.Name + "_" + noPrefix
		c.Description = fmt.Sprintf("%s %s is the maximum of [%s](#%s) for label `%s`.",
			c.Description, c.Name, withPrefix, withPrefix, metric.Source)
	} else {
		c.Name = metric.Name + "_" + c.Name
		if metric.HasCustomName {
			c.Name = metric.Source + "_" + noPrefix
		}
		c.Description = fmt.Sprintf("%s %s is [%s](#%s) aggregated by `%s`.",
			c.Description, c.Name, withPrefix, withPrefix, metric.Name)
	}
	c.Panels = metricsPanelMap[c.Name].Panels
}

func findAPI(apis []MetricDef, other MetricDef) []int {
	var indices []int
	for i, a := range apis {
		if a.API == other.API {
			indices = append(indices, i)
		}
	}
	return indices
}

func FetchAndCategorizePrometheusMetrics(promURL string) (map[string]bool, map[string]bool, error) {
	urlStr := promURL + "/api/v1/series?match[]={datacenter!=\"\"}"

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch metrics from Prometheus: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected status code from Prometheus: %d", resp.StatusCode)
	}

	var result struct {
		Status string              `json:"status"`
		Data   []map[string]string `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, fmt.Errorf("failed to decode Prometheus response: %w", err)
	}
	if result.Status != "success" {
		return nil, nil, fmt.Errorf("unexpected status from Prometheus: %s", result.Status)
	}

	// Categorize metrics
	restMetrics := make(map[string]bool)
	zapiMetrics := make(map[string]bool)
	for _, series := range result.Data {
		metricName := series["__name__"]
		switch series["datacenter"] {
		case "REST":
			restMetrics[metricName] = true
		case "ZAPI":
			zapiMetrics[metricName] = true
		case "keyperf":
			restMetrics[metricName] = true
		}
	}

	return restMetrics, zapiMetrics, nil
}
func ValidateMetrics(documentedRest, documentedZapi map[string]Counter, prometheusRest, prometheusZapi map[string]bool) error {
	var documentedButMissingRestMetrics []string
	var notDocumentedRestMetrics []string
	var documentedButMissingZapiMetrics []string
	var notDocumentedZapiMetrics []string

	// Helper function to check if a REST metric should be excluded
	shouldExcludeRest := func(metric string, apis []MetricDef) bool {
		for _, c := range excludeDocumentedRestMetrics {
			if strings.Contains(metric, c) {
				return true
			}
		}

		for _, api := range apis {
			if api.API == "ZAPI" || api.API == "ZapiPerf" {
				return true
			}
		}

		return false
	}

	// Helper function to check if a ZAPI metric should be excluded
	shouldExcludeZapi := func(metric string, apis []MetricDef) bool {
		for _, prefix := range excludeDocumentedZapiMetrics {
			if strings.Contains(metric, prefix) {
				return true
			}
		}

		for _, api := range apis {
			if api.API == "REST" || api.API == "RestPerf" {
				return true
			}
		}
		return false
	}

	// Helper function to check if an extra REST metric should be excluded
	shouldExcludeExtraRest := func(metric string, set *set.Set) bool {

		for _, c := range excludeNotDocumentedRestMetrics {
			if strings.Contains(metric, c) {
				return true
			}
		}

		var isRestObject bool
		for o := range set.Iter() {
			if strings.HasPrefix(metric, o) {
				isRestObject = true
			}
		}
		return !isRestObject
	}

	// Helper function to check if an extra ZAPI metric should be excluded
	shouldExcludeExtraZapi := func(metric string) bool {

		for _, c := range excludeNotDocumentedZapiMetrics {
			if strings.Contains(metric, c) {
				return true
			}
		}

		return false
	}

	restObjects := set.New()

	for metric, counter := range documentedRest {
		if counter.Object != "" {
			restObjects.Add(counter.Object)
		}
		if !prometheusRest[metric] && !shouldExcludeRest(metric, counter.APIs) {
			documentedButMissingRestMetrics = append(documentedButMissingRestMetrics, metric)
		}
	}

	for metric := range prometheusRest {
		if _, ok := documentedRest[metric]; !ok && !shouldExcludeExtraRest(metric, restObjects) {
			notDocumentedRestMetrics = append(notDocumentedRestMetrics, metric)
		}
	}

	for metric, counter := range documentedZapi {
		if !prometheusZapi[metric] && !shouldExcludeZapi(metric, counter.APIs) {
			documentedButMissingZapiMetrics = append(documentedButMissingZapiMetrics, metric)
		}
	}

	for metric := range prometheusZapi {
		if _, ok := documentedZapi[metric]; !ok && !shouldExcludeExtraZapi(metric) {
			notDocumentedZapiMetrics = append(notDocumentedZapiMetrics, metric)
		}
	}

	// Sort the slices
	sort.Strings(documentedButMissingRestMetrics)
	sort.Strings(notDocumentedRestMetrics)
	sort.Strings(documentedButMissingZapiMetrics)
	sort.Strings(notDocumentedZapiMetrics)

	if len(documentedButMissingRestMetrics) > 0 || len(notDocumentedRestMetrics) > 0 || len(documentedButMissingZapiMetrics) > 0 || len(notDocumentedZapiMetrics) > 0 {
		errorMessage := "Validation failed:\n"
		if len(documentedButMissingRestMetrics) > 0 {
			errorMessage += fmt.Sprintf("Missing Rest metrics in Prometheus but documented: %v\n", documentedButMissingRestMetrics)
		}
		if len(notDocumentedRestMetrics) > 0 {
			errorMessage += fmt.Sprintf("Extra Rest metrics in Prometheus but not documented: %v\n", notDocumentedRestMetrics)
		}
		if len(documentedButMissingZapiMetrics) > 0 {
			errorMessage += fmt.Sprintf("Missing Zapi metrics in Prometheus but documented: %v\n", documentedButMissingZapiMetrics)
		}
		if len(notDocumentedZapiMetrics) > 0 {
			errorMessage += fmt.Sprintf("Extra Zapi metrics in Prometheus but not documented: %v\n", notDocumentedZapiMetrics)
		}
		return errors.New(errorMessage)
	}

	return nil
}

func CategorizeCounters(counters map[string]Counter) (map[string]Counter, map[string]Counter) {
	restCounters := make(map[string]Counter)
	zapiCounters := make(map[string]Counter)

	for _, counter := range counters {
		for _, api := range counter.APIs {
			switch api.API {
			case "REST":
				restCounters[counter.Name] = counter
			case "RestPerf":
				restCounters[counter.Name] = counter
			case "ZAPI":
				zapiCounters[counter.Name] = counter
			case "ZapiPerf":
				zapiCounters[counter.Name] = counter
			case "KeyPerf":
				restCounters[counter.Name] = counter
			}
		}
	}

	return restCounters, zapiCounters
}

func getAllExportedLabels(t *node.Node, counterContents []string) ([]string, []string, bool) {
	metricLabels := make([]string, 0)
	labels := make([]string, 0)
	isInstanceLabels := false
	eData := true
	if exportData := t.GetChildS("export_data"); exportData != nil {
		if exportData.GetContentS() == "false" {
			eData = false
		}
	}
	if exportOptions := t.GetChildS("export_options"); exportOptions != nil {
		if iAllLabels := exportOptions.GetChildS("include_all_labels"); iAllLabels != nil {
			if iAllLabels.GetContentS() == "true" {
				for _, c := range counterContents {
					if c == "" {
						continue
					}
					if _, display, m, _ := template3.ParseMetric(c); m == "key" || m == "label" {
						metricLabels = append(metricLabels, display)
					}
				}
				return metricLabels, metricLabels, false
			}
		}

		if iKeys := exportOptions.GetChildS("instance_keys"); iKeys != nil {
			metricLabels = append(metricLabels, iKeys.GetAllChildContentS()...)
		}
		if iLabels := exportOptions.GetChildS("instance_labels"); iLabels != nil {
			labels = append(labels, iLabels.GetAllChildContentS()...)
			isInstanceLabels = eData
		}
	}
	return metricLabels, append(labels, metricLabels...), isInstanceLabels
}

// searchDescriptionSwagger returns ontap counter description from swagger
func searchDescriptionSwagger(objName string, ontapCounterName string) string {
	val, ok := objectSwaggerMap[objName]
	if ok {
		objName = val
	}
	searchQuery := strings.Join([]string{"definitions", objName, "properties"}, ".")
	cs := strings.Split(ontapCounterName, ".")
	for i, c := range cs {
		if i < len(cs)-1 {
			searchQuery = strings.Join([]string{searchQuery, c, "properties"}, ".")
		} else {
			searchQuery = strings.Join([]string{searchQuery, c, "description"}, ".")
		}
	}
	t := gjson.GetBytes(swaggerBytes, searchQuery)
	return updateDescription(t.ClonedString())
}

// parseZapiCounters parse zapi template counters
func parseZapiCounters(elem *node.Node, path []string, object string, zc map[string]string) {

	name := elem.GetNameS()
	newPath := path

	if elem.GetNameS() != "" {
		newPath = append(newPath, name)
	}

	if elem.GetContentS() != "" {
		v, k := handleZapiCounter(newPath, elem.GetContentS(), object)
		if k != "" {
			zc[k] = v
		}
	}

	for _, child := range elem.GetChildren() {
		parseZapiCounters(child, newPath, object, zc)
	}
}

// handleZapiCounter returns zapi ontap and display counter name
func handleZapiCounter(path []string, content string, object string) (string, string) {
	var (
		name, display, key    string
		splitValues, fullPath []string
	)

	splitValues = strings.Split(content, "=>")
	if len(splitValues) == 1 {
		name = content
	} else {
		name = splitValues[0]
		display = strings.TrimSpace(splitValues[1])
	}

	name = strings.TrimSpace(strings.TrimLeft(name, "^"))
	fullPath = path
	fullPath = append(fullPath, name)
	key = strings.Join(fullPath, ".")
	if display == "" {
		display = template3.ParseZAPIDisplay(object, fullPath)
	}

	if content[0] != '^' {
		return key, object + "_" + display
	}

	return "", ""
}

func updateDescription(description string) string {
	s := replacer.Replace(description)
	return s
}

// [Top $TopResources Average Disk Utilization Per Aggregate](GRAFANA_HOST/d/cdot-aggregate/ontap3a-aggregate?orgId=1&viewPanel=63)
// [p.Panel](GRAFANA_HOST/p.PanelLink)

func (c Counter) Header() string {
	return `
| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|`
}

func (c Counter) PanelHeader() string {
	return `
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|`
}

func (c Counter) HasAPIs() bool {
	return len(c.APIs) > 0
}

func (c Counter) HasPanels() bool {
	return len(c.Panels) > 0
}

func (m MetricDef) TableRow() string {
	switch {
	case strings.Contains(m.Template, "eseries"):
		return fmt.Sprintf("| %s | `%s` | `%s` | %s |", m.API, m.Endpoint, m.ESeriesCounter, m.Template)
	case strings.Contains(m.Template, "perf"):
		unitTypeBase := `<br><span class="key">Unit:</span> ` + m.Unit +
			`<br><span class="key">Type:</span> ` + m.Type +
			`<br><span class="key">Base:</span> ` + m.BaseCounter
		return fmt.Sprintf("| %s | `%s` | `%s`%s | %s |",
			m.API, m.Endpoint, m.ONTAPCounter, unitTypeBase, m.Template)
	case m.Unit != "":
		unit := `<br><span class="key">Unit:</span> ` + m.Unit
		return fmt.Sprintf("| %s | `%s` | `%s`%s | %s | ",
			m.API, m.Endpoint, m.ONTAPCounter, unit, m.Template)
	case strings.Contains(m.Template, "ciscorest"):
		return fmt.Sprintf("| %s | `%s` | `%s` | %s |", m.API, m.Endpoint, m.CiscoCounter, m.Template)
	case strings.Contains(m.Template, "storagegrid"):
		return fmt.Sprintf("| %s | `%s` | `%s` | %s |", m.API, m.Endpoint, m.SGCounter, m.Template)
	default:
		return fmt.Sprintf("| %s | `%s` | `%s` | %s |", m.API, m.Endpoint, m.ONTAPCounter, m.Template)
	}
}

func (p PanelDef) DashboardTableRow() string {
	return fmt.Sprintf("| %s | %s | %s | [%s](/%s) |", p.Dashboard, p.Row, p.Type, p.Panel, p.PanelLink)
}

// readSwaggerJSON downloads poller swagger and convert to json format
func readSwaggerJSON(opts *Options) []byte {
	var f []byte
	path, err := downloadSwaggerForPoller(opts.Poller)
	if err != nil {
		log.Fatal("failed to download swagger:", err)
		return nil
	}
	cmd := fmt.Sprintf("cat %s | dasel -i yaml -o json", path)
	f, err = exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatal("Failed to execute command:", cmd, err)
		return nil
	}
	return f
}

func downloadSwaggerForPoller(pName string) (string, error) {
	var (
		poller         *conf.Poller
		err            error
		addr           string
		shouldDownload = true
		swagTime       time.Time
	)

	if poller, addr, err = rest.GetPollerAndAddr(pName); err != nil {
		return "", err
	}

	tmp := os.TempDir()
	swaggerPath := filepath.Join(tmp, addr+"-swagger.yaml")
	fileInfo, err := os.Stat(swaggerPath)

	if os.IsNotExist(err) {
		fmt.Printf("%s does not exist downloading\n", swaggerPath)
	} else {
		swagTime = fileInfo.ModTime()
		twoWeeksAgo := swagTime.Local().AddDate(0, 0, -14)
		if swagTime.Before(twoWeeksAgo) {
			fmt.Printf("%s is more than two weeks old, re-download", swaggerPath)
		} else {
			shouldDownload = false
		}
	}

	if shouldDownload {
		swaggerURL := "https://" + addr + "/docs/api/swagger.yaml"
		bytesDownloaded, err := downloadSwagger(poller, swaggerPath, swaggerURL, false)
		if err != nil {
			fmt.Printf("error downloading swagger %s\n", err)
			if bytesDownloaded == 0 {
				// if the tmp file exists, remove it since it is empty
				_ = os.Remove(swaggerPath)
			}
			return "", err
		}
		fmt.Printf("downloaded %d bytes from %s\n", bytesDownloaded, swaggerURL)
	}

	fmt.Printf("Using downloaded file %s with timestamp %s\n", swaggerPath, swagTime)
	return swaggerPath, nil
}

func downloadSwagger(poller *conf.Poller, path string, urlStr string, verbose bool) (int64, error) {
	out, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("unable to create %s to save swagger.yaml", path)
	}
	defer func(out *os.File) { _ = out.Close() }(out)
	request, err := requests.New("GET", urlStr, nil)
	if err != nil {
		return 0, err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	credentials := auth.NewCredentials(poller, slog.Default())
	transport, err := credentials.Transport(request, poller)
	if err != nil {
		return 0, err
	}
	httpclient := &http.Client{Transport: transport, Timeout: timeout}

	if verbose {
		requestOut, _ := httputil.DumpRequestOut(request, false)
		fmt.Printf("REQUEST: %s\n%s\n", urlStr, requestOut)
	}
	response, err := httpclient.Do(request)
	if err != nil {
		return 0, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer response.Body.Close()

	if verbose {
		debugResp, _ := httputil.DumpResponse(response, false)
		fmt.Printf("RESPONSE: \n%s", debugResp)
	}
	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error making request. server response statusCode=[%d]", response.StatusCode)
	}
	n, err := io.Copy(out, response.Body)
	if err != nil {
		return 0, fmt.Errorf("error while downloading %s err=%w", urlStr, err)
	}
	return n, nil
}

func LogErrAndExit(err error) {
	fmt.Printf("%v\n", err)
	os.Exit(1)
}

func GenerateOntapCounterTemplate(counters map[string]Counter, version string) {
	targetPath := "docs/ontap-metrics.md"
	t, err := template.New("counter.tmpl").ParseFiles("cmd/tools/generate/counter.tmpl")
	if err != nil {
		panic(err)
	}
	out, err := os.Create(targetPath)
	if err != nil {
		panic(err)
	}

	keys := make([]string, 0, len(counters))
	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	values := make([]Counter, 0, len(keys))

	table := tw.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Missing", "Counter", "APIs", "Endpoint", "ONTAPCounter", "Template"})

	for _, k := range keys {
		if k == "" {
			continue
		}
		counter := counters[k]

		if counter.Description == "" {
			for _, def := range counter.APIs {
				if _, ok := knownDescriptionGaps[counter.Name]; !ok {
					appendRow(table, "Description", counter, def)
				}
			}
		}
		values = append(values, counter)
	}

	for _, k := range keys {
		if k == "" {
			continue
		}
		counter := counters[k]

		// Print such counters which are missing Rest mapping
		if len(counter.APIs) == 1 {
			if counter.APIs[0].API == "ZAPI" {
				isPrint := true
				for _, substring := range excludeLogRestCounters {
					if strings.HasPrefix(counter.Name, substring) {
						isPrint = false
						break
					}
				}
				// missing Rest Mapping
				if isPrint {
					for _, def := range counter.APIs {
						hasPrefix := false
						for prefix := range knownMappingGaps {
							if strings.HasPrefix(counter.Name, prefix) {
								hasPrefix = true
								break
							}
						}
						if !hasPrefix {
							appendRow(table, "REST", counter, def)
						}
					}
				}
			}
		}

		for _, def := range counter.APIs {
			if def.ONTAPCounter == "" {
				for _, def := range counter.APIs {
					hasPrefix := false
					for prefix := range knownMappingGaps {
						if strings.HasPrefix(counter.Name, prefix) {
							hasPrefix = true
							break
						}
					}
					if !hasPrefix {
						appendRow(table, "Mapping", counter, def)
					}
				}
			}
		}
	}
	table.Render()
	c := CounterTemplate{
		Counters: values,
		CounterMetaData: CounterMetaData{
			Date:         time.Now().Format("2006-Jan-02"),
			OntapVersion: version,
		},
	}

	err = t.Execute(out, c)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Harvest metric documentation generated at %s \n", targetPath)

	if table.NumLines() > 0 {
		slog.Error("Issues found: Please refer to the table above")
		os.Exit(1)
	}
}

func GenerateStorageGridCounterTemplate(counters map[string]Counter, version string) {
	targetPath := "docs/storagegrid-metrics.md"
	t, err := template.New("storagegrid_counter.tmpl").ParseFiles("cmd/tools/generate/storagegrid_counter.tmpl")
	if err != nil {
		panic(err)
	}
	out, err := os.Create(targetPath)
	if err != nil {
		panic(err)
	}

	keys := make([]string, 0, len(counters))
	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	values := make([]Counter, 0, len(keys))

	table := tw.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Missing", "Counter", "APIs", "Endpoint", "SGCounter", "Template"})

	for _, k := range keys {
		if k == "" {
			continue
		}
		counter := counters[k]
		if !strings.HasPrefix(counter.Name, "storagegrid_") {
			continue
		}

		if _, ok := knownMappingGapsSG[k]; !ok {
			if counter.Description == "" {
				appendRow(table, "Description", counter, MetricDef{API: ""})
			}
		}

		values = append(values, counter)
	}

	table.Render()
	c := CounterTemplate{
		Counters: values,
		CounterMetaData: CounterMetaData{
			Date:      time.Now().Format("2006-Jan-02"),
			SGVersion: version,
		},
	}

	err = t.Execute(out, c)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Harvest metric documentation generated at %s \n", targetPath)

	if table.NumLines() > 0 {
		log.Fatalf("Issues found: refer table above")
	}
}

func GenerateCiscoSwitchCounterTemplate(counters map[string]Counter, version string) {
	targetPath := "docs/cisco-switch-metrics.md"
	t, err := template.New("cisco_counter.tmpl").ParseFiles("cmd/tools/generate/cisco_counter.tmpl")
	if err != nil {
		panic(err)
	}
	out, err := os.Create(targetPath)
	if err != nil {
		panic(err)
	}

	keys := make([]string, 0, len(counters))
	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	values := make([]Counter, 0, len(keys))

	table := tw.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Missing", "Counter", "APIs", "Endpoint", "CiscoCounter", "Template"})

	for _, k := range keys {
		if k == "" {
			continue
		}
		counter := counters[k]
		if !strings.HasPrefix(counter.Name, "cisco_") {
			continue
		}

		if counter.Description == "" {
			appendRow(table, "Description", counter, MetricDef{API: ""})
		}

		values = append(values, counter)
	}

	table.Render()
	c := CounterTemplate{
		Counters: values,
		CounterMetaData: CounterMetaData{
			Date:         time.Now().Format("2006-Jan-02"),
			CiscoVersion: version,
		},
	}

	err = t.Execute(out, c)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Harvest metric documentation generated at %s \n", targetPath)

	if table.NumLines() > 0 {
		log.Fatalf("Issues found: refer table above")
	}
}

func appendRow(table *tw.Table, missing string, counter Counter, def MetricDef) {
	if def.API != "" {
		table.Append([]string{missing, counter.Name, def.API, def.Endpoint, def.ONTAPCounter, def.Template})
	} else {
		table.Append([]string{missing, counter.Name})
	}
}

func GenerateESeriesCounterTemplate(counters map[string]Counter, version string) {
	targetPath := "docs/eseries-metrics.md"
	t, err := template.New("eseries_counter.tmpl").ParseFiles("cmd/tools/generate/eseries_counter.tmpl")
	if err != nil {
		panic(err)
	}
	out, err := os.Create(targetPath)
	if err != nil {
		panic(err)
	}

	keys := make([]string, 0, len(counters))
	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	values := make([]Counter, 0, len(keys))

	table := tw.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Missing", "Counter", "APIs", "Endpoint", "ESeriesCounter", "Template"})

	for _, k := range keys {
		if k == "" {
			continue
		}
		counter := counters[k]
		if !strings.HasPrefix(counter.Name, "eseries_") {
			continue
		}

		if counter.Description == "" {
			appendRow(table, "Description", counter, MetricDef{API: ""})
		}

		values = append(values, counter)
	}

	table.Render()
	c := CounterTemplate{
		Counters: values,
		CounterMetaData: CounterMetaData{
			Date:           time.Now().Format("2006-Jan-02"),
			ESeriesVersion: version,
		},
	}

	err = t.Execute(out, c)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Harvest metric documentation generated at %s \n", targetPath)

	if table.NumLines() > 0 {
		log.Fatalf("Issues found: refer table above")
	}
}
