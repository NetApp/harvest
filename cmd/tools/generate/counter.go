package generate

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	template2 "github.com/netapp/harvest/v2/cmd/tools/template"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	tw "github.com/netapp/harvest/v2/third_party/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
	"log"
	"maps"
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
		return fmt.Sprintf("| %s | `%s` | `%s`%s | %s | ",
			m.API, m.Endpoint, m.ONTAPCounter, unitTypeBase, m.Template)
	} else if m.Unit != "" {
		unit := `<br><span class="key">Unit:</span> ` + m.Unit
		return fmt.Sprintf("| %s | `%s` | `%s`%s | %s | ",
			m.API, m.Endpoint, m.ONTAPCounter, unit, m.Template)
	}
	return fmt.Sprintf("| %s | `%s` | `%s` | %s |", m.API, m.Endpoint, m.ONTAPCounter, m.Template)
}

type Counter struct {
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
	path, err := rest.ReadOrDownloadSwagger(opts.Poller)
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
	return updateDescription(t.String())
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
		return processRestConfigCounters(path)
	})

	for k, v := range restPerfCounters {
		restCounters[k] = v
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
func processRestConfigCounters(path string) map[string]Counter {
	var (
		counters = make(map[string]Counter)
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
		metricLabels, labels = getAllExportedLabels(t, templateCounters.GetAllChildContentS())
		processCounters(templateCounters.GetAllChildContentS(), &model, path, model.Query, counters, metricLabels)
		// This is for object_labels metrics
		harvestName := model.Object + "_" + "labels"
		counters[harvestName] = Counter{Name: harvestName, Labels: labels}
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
					processCounters(line.GetAllChildContentS(), &model, path, query, counters, metricLabels)
				}
			}
		}
	}

	// If the template has any PluginMetrics, add them
	for _, metric := range model.PluginMetrics {
		co := Counter{
			Name: model.Object + "_" + metric.Name,
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

	return counters
}

func processCounters(counterContents []string, model *template2.Model, path, query string, counters map[string]Counter, metricLabels []string) {
	for _, c := range counterContents {
		if c == "" {
			continue
		}
		name, display, m, _ := util.ParseMetric(c)
		if _, ok := excludeCounters[name]; ok {
			continue
		}
		description := searchDescriptionSwagger(model.Object, name)
		harvestName := model.Object + "_" + display
		if m == "float" {
			co := Counter{
				Name:        harvestName,
				Description: description,
				APIs: []MetricDef{
					{
						API:          "REST",
						Endpoint:     query,
						Template:     path,
						ONTAPCounter: name,
					},
				},
				Labels: metricLabels,
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

	metricLabels, _ := getAllExportedLabels(t, templateCounters.GetAllChildContentS())
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
					counters[harvestName] = co

					// handle deprecate counters
					if rName, ok := zapiDeprecateCounterMap[name]; ok {
						hName := model.Object + "_" + rName
						ro := Counter{
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
						counters[hName] = ro
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
			Name: model.Object + "_" + metric.Name,
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
		counters = make(map[string]Counter)
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
	metricLabels, labels = getAllExportedLabels(t, templateCounters.GetAllChildContentS())
	// This is for object_labels metrics
	harvestName := model.Object + "_" + "labels"
	counters[harvestName] = Counter{Name: harvestName, Labels: labels}
	for _, c := range templateCounters.GetChildren() {
		parseZapiCounters(c, []string{}, model.Object, zc)
	}

	for k, v := range zc {
		if _, ok := excludeCounters[k]; ok {
			continue
		}
		co := Counter{
			Name: k,
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
		counters[k] = co

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
			Name: model.Object + "_" + metric.Name,
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
		if strings.HasSuffix(path, "default.yaml") {
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

func generateCounterTemplate(counters map[string]Counter, version [3]int) {
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
	verWithDots := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(version)), "."), "[]")
	c := CounterTemplate{
		Counters: values,
		CounterMetaData: CounterMetaData{
			Date:         time.Now().Format("2006-Jan-02"),
			OntapVersion: verWithDots,
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
	metricLabels, _ := getAllExportedLabels(t, templateCounters.GetAllChildContentS())
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
	records, err = rest.Fetch(client, href)
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
		ontapCounterName := r.Get("name").String()
		if _, ok := excludeCounters[ontapCounterName]; ok {
			return true
		}

		description := r.Get("description").String()
		ty := r.Get("type").String()
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
				Name:        v,
				Description: description,
				APIs: []MetricDef{
					{
						API:          "REST",
						Endpoint:     model.Query,
						Template:     path,
						ONTAPCounter: ontapCounterName,
						Unit:         r.Get("unit").String(),
						Type:         ty,
						BaseCounter:  r.Get("denominator.name").String(),
					},
				},
				Labels: metricLabels,
			}
			counters[c.Name] = c

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
			Name: model.Object + "_" + metric.Name,
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

func getAllExportedLabels(t *node.Node, counterContents []string) ([]string, []string) {
	metricLabels := make([]string, 0)
	labels := make([]string, 0)
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
				return metricLabels, metricLabels
			}
		}

		if iKeys := exportOptions.GetChildS("instance_keys"); iKeys != nil {
			metricLabels = append(metricLabels, iKeys.GetAllChildContentS()...)
		}
		if iLabels := exportOptions.GetChildS("instance_labels"); iLabels != nil {
			labels = append(labels, iLabels.GetAllChildContentS()...)
		}
	}
	return metricLabels, append(labels, metricLabels...)
}
