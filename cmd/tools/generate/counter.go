package generate

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/netapp/harvest/v2/cmd/collectors/keyperf"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	template2 "github.com/netapp/harvest/v2/cmd/tools/template"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
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

var (
	replacer         = strings.NewReplacer("\n", "", ":", "")
	objectSwaggerMap = map[string]string{
		"volume":             "xc_volume",
		"aggr":               "xc_aggregate",
		"net_port":           "xc_broadcast_domain",
		"environment_sensor": "sensors",
		"ontaps3":            "xc_s3_bucket",
		"security_ssh":       "cluster_ssh_server",
		"namespace":          "nvme_namespace",
		"fcp":                "fc_port",
		"svm_cifs":           "cifs_service",
		"svm_nfs":            "nfs_service",
		"lif":                "ip_interface",
		"flexcache":          "volume",
	}
	swaggerBytes         []byte
	excludePerfTemplates = map[string]struct{}{
		"workload_detail.yaml":        {},
		"workload_detail_volume.yaml": {},
	}
	excludeCounters = map[string]struct{}{
		"write_latency_histogram": {},
		"read_latency_histogram":  {},
		"latency_histogram":       {},
		"nfsv3_latency_hist":      {},
		"nfs4_latency_hist":       {},
		"read_latency_hist":       {},
		"write_latency_hist":      {},
		"total.latency_histogram": {},
		"nfs41_latency_hist":      {},
	}
	// Excludes these Rest gaps from logs
	excludeLogRestCounters = []string{
		"smb2_",
		"ontaps3_svm_",
		"nvmf_rdma_port_",
		"nvmf_tcp_port_",
		"netstat_",
		"external_service_op_",
		"fabricpool_average_latency",
		"fabricpool_get_throughput_bytes",
		"fabricpool_put_throughput_bytes",
		"fabricpool_stats",
		"fabricpool_throughput_ops",
		"iw_",
	}
	// Special handling perf objects
	specialPerfObjects = map[string]bool{
		"svm_nfs":  true,
		"node_nfs": true,
	}

	knownDescriptionGaps = map[string]struct{}{
		"ontaps3_object_count":                      {},
		"security_certificate_expiry_time":          {},
		"volume_capacity_tier_footprint":            {},
		"volume_capacity_tier_footprint_percent":    {},
		"volume_num_compress_attempts":              {},
		"volume_num_compress_fail":                  {},
		"volume_performance_tier_footprint":         {},
		"volume_performance_tier_footprint_percent": {},
	}

	knownMappingGaps = map[string]struct{}{
		"aggr_snapshot_inode_used_percent":                      {},
		"aggr_space_reserved":                                   {},
		"flexcache_blocks_requested_from_client":                {},
		"flexcache_blocks_retrieved_from_origin":                {},
		"flexcache_evict_rw_cache_skipped_reason_disconnected":  {},
		"flexcache_evict_skipped_reason_config_noent":           {},
		"flexcache_evict_skipped_reason_disconnected":           {},
		"flexcache_evict_skipped_reason_offline":                {},
		"flexcache_invalidate_skipped_reason_config_noent":      {},
		"flexcache_invalidate_skipped_reason_disconnected":      {},
		"flexcache_invalidate_skipped_reason_offline":           {},
		"flexcache_miss_percent":                                {},
		"flexcache_nix_retry_skipped_reason_initiator_retrieve": {},
		"flexcache_nix_skipped_reason_config_noent":             {},
		"flexcache_nix_skipped_reason_disconnected":             {},
		"flexcache_nix_skipped_reason_in_progress":              {},
		"flexcache_nix_skipped_reason_offline":                  {},
		"flexcache_reconciled_data_entries":                     {},
		"flexcache_reconciled_lock_entries":                     {},
		"quota_disk_used_pct_threshold":                         {},
		"rw_ctx_cifs_giveups":                                   {},
		"rw_ctx_cifs_rewinds":                                   {},
		"rw_ctx_nfs_giveups":                                    {},
		"rw_ctx_nfs_rewinds":                                    {},
		"rw_ctx_qos_flowcontrol":                                {},
		"rw_ctx_qos_rewinds":                                    {},
		"security_audit_destination_port":                       {},
		"wafl_reads_from_pmem":                                  {},
	}

	excludeDocumentedRestMetrics = []string{
		"ontaps3_svm_",
		"svm_ontaps3_svm_",
		"nvme_lif_",
		"fcvi_",
		"smb2_",
		"nvmf_",
		"flashpool_",
		"fcvi_",
		"nfs_diag_",
		"iw_",
		"node_cifs_",
		"svm_cifs_",
		"vscan_",
		"svm_vscan_",
		"token_",
		"metrocluster_",
		"path_",
		"ndmp_session",
		"export_rule_labels",
		"mediator_labels",
		"net_connection_labels",
		"health_",
		"aggr_hybrid_disk_count",
		"nfs_clients_idle_duration",
		"ems_events",
		"volume_top_clients",
		"volume_top_files",
		"cluster_software",
	}

	excludeDocumentedZapiMetrics = []string{
		"fabricpool_",
		"external_service_",
		"netstat_",
		"flexcache_",
		"quota_disk_used_pct_threshold",
		"ems_events",
	}

	// Exclude extra metrics for REST
	excludeNotDocumentedRestMetrics = []string{
		"volume_aggr_labels",
		"flexcache_",
		"hist_",
		"volume_arw_status",
		"ALERTS",
	}

	// Exclude extra metrics for ZAPI
	excludeNotDocumentedZapiMetrics = []string{
		"volume_aggr_labels",
		"hist_",
		"security_",
		"svm_ldap",
		"ALERTS",
	}
)

type Counters struct {
	C []Counter `yaml:"counters"`
}

type CounterMetaData struct {
	Date         string
	OntapVersion string
}

type CounterTemplate struct {
	Counters        []Counter
	CounterMetaData CounterMetaData
}

type MetricDef struct {
	API          string `yaml:"API"`
	Endpoint     string `yaml:"Endpoint"`
	ONTAPCounter string `yaml:"ONTAPCounter"`
	Template     string `yaml:"Template"`
	Unit         string `yaml:"Unit"`
	Type         string `yaml:"Type"`
	BaseCounter  string `yaml:"BaseCounter"`
}

func (m MetricDef) TableRow() string {
	if strings.Contains(m.Template, "perf") {
		unitTypeBase := `<br><span class="key">Unit:</span> ` + m.Unit +
			`<br><span class="key">Type:</span> ` + m.Type +
			`<br><span class="key">Base:</span> ` + m.BaseCounter
		return fmt.Sprintf("| %s | `%s` | `%s`%s | %s |",
			m.API, m.Endpoint, m.ONTAPCounter, unitTypeBase, m.Template)
	} else if m.Unit != "" {
		unit := `<br><span class="key">Unit:</span> ` + m.Unit
		return fmt.Sprintf("| %s | `%s` | `%s`%s | %s | ",
			m.API, m.Endpoint, m.ONTAPCounter, unit, m.Template)
	}
	return fmt.Sprintf("| %s | `%s` | `%s` | %s |", m.API, m.Endpoint, m.ONTAPCounter, m.Template)
}

type Counter struct {
	Object      string      `yaml:"-"`
	Name        string      `yaml:"Name"`
	Description string      `yaml:"Description"`
	APIs        []MetricDef `yaml:"APIs"`
	Labels      []string    `yaml:"Labels"`
}

func (c Counter) Header() string {
	return `
| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|`
}

func (c Counter) HasAPIs() bool {
	return len(c.APIs) > 0
}

// readSwaggerJSON downloads poller swagger and convert to json format
func readSwaggerJSON() []byte {
	var f []byte
	path, err := downloadSwaggerForPoller(opts.Poller)
	if err != nil {
		log.Fatal("failed to download swagger:", err)
		return nil
	}
	cmd := fmt.Sprintf("dasel -f %s -r yaml -w json", path)
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

// processRestCounters parse rest and restperf templates
func processRestCounters(dir string, client *rest.Client) map[string]Counter {
	restPerfCounters := visitRestTemplates(filepath.Join(dir, "conf", "restperf"), client, func(path string, client *rest.Client) map[string]Counter {
		if _, ok := excludePerfTemplates[filepath.Base(path)]; ok {
			return nil
		}
		return processRestPerfCounters(path, client)
	})

	restCounters := visitRestTemplates(filepath.Join(dir, "conf", "rest"), client, func(path string, client *rest.Client) map[string]Counter { // revive:disable-line:unused-parameter
		return processRestConfigCounters(path, "REST")
	})

	keyPerfCounters := visitRestTemplates(filepath.Join(dir, "conf", "keyperf"), client, func(path string, client *rest.Client) map[string]Counter { // revive:disable-line:unused-parameter
		return processRestConfigCounters(path, keyPerfAPI)
	})

	for k, v := range restPerfCounters {
		restCounters[k] = v
	}

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
	return restCounters
}

// processZapiCounters parse zapi and zapiperf templates
func processZapiCounters(dir string, client *zapi.Client) map[string]Counter {
	zapiCounters := visitZapiTemplates(filepath.Join(dir, "conf", "zapi", "cdot"), client, func(path string, client *zapi.Client) map[string]Counter { // revive:disable-line:unused-parameter
		return processZapiConfigCounters(path)
	})
	zapiPerfCounters := visitZapiTemplates(filepath.Join(dir, "conf", "zapiperf", "cdot"), client, func(path string, client *zapi.Client) map[string]Counter {
		if _, ok := excludePerfTemplates[filepath.Base(path)]; ok {
			return nil
		}
		return processZAPIPerfCounters(path, client)
	})

	for k, v := range zapiPerfCounters {
		zapiCounters[k] = v
	}
	return zapiCounters
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
		display = util.ParseZAPIDisplay(object, fullPath)
	}

	if content[0] != '^' {
		return key, object + "_" + display
	}

	return "", ""
}

// processRestConfigCounters process Rest config templates
func processRestConfigCounters(path string, api string) map[string]Counter {
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
		processCounters(templateCounters.GetAllChildContentS(), &model, path, model.Query, counters, metricLabels, api)
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
					processCounters(line.GetAllChildContentS(), &model, path, query, counters, metricLabels, api)
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

func processCounters(counterContents []string, model *template2.Model, path, query string, counters map[string]Counter, metricLabels []string, api string) {
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
		name, display, m, _ := util.ParseMetric(c)
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
					Labels: metricLabels,
				}
			}
			counters[harvestName] = co

			// If the template has any MultiplierMetrics, add them
			for _, metric := range model.MultiplierMetrics {
				mc := co
				addAggregatedCounter(&mc, metric, harvestName, display)
				counters[mc.Name] = mc
			}
		}
	}
}

// processZAPIPerfCounters process ZapiPerf counters
func processZAPIPerfCounters(path string, client *zapi.Client) map[string]Counter {
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
			Labels: labels,
		}
	}
	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := util.ParseMetric(c)
			if strings.HasPrefix(display, model.Object) {
				display = strings.TrimPrefix(display, model.Object)
				display = strings.TrimPrefix(display, "_")
			}
			harvestName := model.Object + "_" + display
			if m == "float" {
				if _, ok := excludeCounters[name]; ok {
					continue
				}
				if zapiTypeMap[name] != "string" {
					co := Counter{
						Object:      model.Object,
						Name:        harvestName,
						Description: zapiDescMap[name],
						APIs: []MetricDef{
							{
								API:          "ZAPI",
								Endpoint:     "perf-object-get-instances" + " " + model.Query,
								Template:     path,
								ONTAPCounter: name,
								Unit:         zapiUnitMap[name],
								Type:         zapiTypeMap[name],
								BaseCounter:  zapiBaseCounterMap[name],
							},
						},
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
									API:          "ZAPI",
									Endpoint:     "perf-object-get-instances" + " " + model.Query,
									Template:     path,
									ONTAPCounter: rName,
									Unit:         zapiUnitMap[rName],
									Type:         zapiTypeMap[rName],
									BaseCounter:  zapiBaseCounterMap[rName],
								},
							},
						}
						if model.ExportData != "false" {
							counters[hName] = ro
						}
					}

					// If the template has any MultiplierMetrics, add them
					for _, metric := range model.MultiplierMetrics {
						mc := co
						addAggregatedCounter(&mc, metric, harvestName, display)
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
					API:          "ZAPI",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: metric.Source,
				},
			},
		}
		counters[co.Name] = co
	}
	// handling for templates with common object names
	if specialPerfObjects[model.Object] {
		return specialHandlingPerfCounters(counters, model)
	}
	return counters
}

func processZapiConfigCounters(path string) map[string]Counter {
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
			Labels: metricLabels,
		}
		if model.ExportData != "false" {
			counters[k] = co
		}

		// If the template has any MultiplierMetrics, add them
		for _, metric := range model.MultiplierMetrics {
			mc := co
			addAggregatedCounter(&mc, metric, co.Name, model.Object)
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
		}
		counters[co.Name] = co
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
		for k, v := range r {
			result[k] = v
		}
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
		for k, v := range r {
			result[k] = v
		}
		return nil
	})

	if err != nil {
		log.Fatal("failed to read template:", err)
		return nil
	}
	return result
}

func updateDescription(description string) string {
	s := replacer.Replace(description)
	return s
}

func generateCounterTemplate(counters map[string]Counter, version string) {
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
	table.SetHeader([]string{"Missing", "Counter", "APIs", "Endpoint", "Counter", "Template"})

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
						if _, ok := knownMappingGaps[counter.Name]; !ok {
							appendRow(table, "REST", counter, def)
						}
					}
				}
			}
		}

		for _, def := range counter.APIs {
			if def.ONTAPCounter == "" {
				for _, def := range counter.APIs {
					if _, ok := knownMappingGaps[counter.Name]; !ok {
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
		log.Fatalf("Issues found: refer table above")
	}
}

func appendRow(table *tw.Table, missing string, counter Counter, def MetricDef) {
	table.Append([]string{missing, counter.Name, def.API, def.Endpoint, def.ONTAPCounter, def.Template})
}

// Regex to match NFS version and operation
var reRemove = regexp.MustCompile(`NFSv\d+\.\d+`)

func mergeCounters(restCounters map[string]Counter, zapiCounters map[string]Counter) map[string]Counter {
	// handle special counters
	restKeys := slices.Sorted(maps.Keys(restCounters))
	for _, k := range restKeys {
		v := restCounters[k]
		hashIndex := strings.Index(k, "#")
		if hashIndex != -1 {
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
		hashIndex := strings.Index(k, "#")
		if hashIndex != -1 {
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
			}
			restCounters[v.Name] = co
		}
	}
	return restCounters
}

func processRestPerfCounters(path string, client *rest.Client) map[string]Counter {
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
					API:          "REST",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: "Harvest generated",
				},
			},
			Labels: labels,
		}
	}
	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := util.ParseMetric(c)
			if m == "float" {
				counterMap[name] = model.Object + "_" + display
				counterMapNoPrefix[name] = display
			}
		}
	}
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
						API:          "REST",
						Endpoint:     model.Query,
						Template:     path,
						ONTAPCounter: ontapCounterName,
						Unit:         r.Get("unit").ClonedString(),
						Type:         ty,
						BaseCounter:  r.Get("denominator.name").ClonedString(),
					},
				},
				Labels: metricLabels,
			}
			if model.ExportData != "false" {
				counters[c.Name] = c
			}

			// If the template has any MultiplierMetrics, add them
			for _, metric := range model.MultiplierMetrics {
				mc := c
				addAggregatedCounter(&mc, metric, v, counterMapNoPrefix[ontapCounterName])
				counters[mc.Name] = mc
			}
		}
		return true
	})

	// If the template has any PluginMetrics, add them
	for _, metric := range model.PluginMetrics {
		co := Counter{
			Object: model.Object,
			Name:   model.Object + "_" + metric.Name,
			APIs: []MetricDef{
				{
					API:          "REST",
					Endpoint:     model.Query,
					Template:     path,
					ONTAPCounter: metric.Source,
				},
			},
		}
		counters[co.Name] = co
	}
	// handling for templates with common object names/metric name
	if specialPerfObjects[model.Object] {
		return specialHandlingPerfCounters(counters, model)
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

func addAggregatedCounter(c *Counter, metric plugin.DerivedMetric, withPrefix string, noPrefix string) {
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
}

func processExternalCounters(dir string, counters map[string]Counter) map[string]Counter {
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
	for _, v := range c.C {
		if v1, ok := counters[v.Name]; !ok {
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

func findAPI(apis []MetricDef, other MetricDef) []int {
	var indices []int
	for i, a := range apis {
		if a.API == other.API {
			indices = append(indices, i)
		}
	}
	return indices
}

func fetchAndCategorizePrometheusMetrics(promURL string) (map[string]bool, map[string]bool, error) {
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
		}
	}

	return restMetrics, zapiMetrics, nil
}
func validateMetrics(documentedRest, documentedZapi map[string]Counter, prometheusRest, prometheusZapi map[string]bool) error {
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
			if api.API == "ZAPI" {
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
			if api.API == "REST" {
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

func categorizeCounters(counters map[string]Counter) (map[string]Counter, map[string]Counter) {
	restCounters := make(map[string]Counter)
	zapiCounters := make(map[string]Counter)

	for _, counter := range counters {
		for _, api := range counter.APIs {
			switch api.API {
			case "REST":
				restCounters[counter.Name] = counter
			case "ZAPI":
				zapiCounters[counter.Name] = counter
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
					if _, display, m, _ := util.ParseMetric(c); m == "key" || m == "label" {
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
