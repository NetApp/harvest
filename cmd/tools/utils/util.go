package utils

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	template2 "github.com/netapp/harvest/v2/cmd/tools/template"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	SwaggerBytes         []byte
	ExcludePerfTemplates = map[string]struct{}{
		"workload_detail.yaml":        {},
		"workload_detail_volume.yaml": {},
	}
	ExcludeCounters = map[string]struct{}{
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
	ExcludeLogRestCounters = []string{
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
	SpecialPerfObjects = map[string]bool{
		"svm_nfs":  true,
		"node_nfs": true,
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

type Counter struct {
	Name        string      `yaml:"Name"`
	Description string      `yaml:"Description"`
	APIs        []MetricDef `yaml:"APIs"`
	Visited     bool
	Template    string
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

func (c Counter) Header() string {
	return `
| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|`
}

func (c Counter) HasAPIs() bool {
	return len(c.APIs) > 0
}

// ParseZapiCounters parse zapi template counters
func ParseZapiCounters(elem *node.Node, path []string, object string, zc map[string]string) {

	name := elem.GetNameS()
	newPath := path

	if len(elem.GetNameS()) != 0 {
		newPath = append(newPath, name)
	}

	if len(elem.GetContentS()) != 0 {
		v, k := handleZapiCounter(newPath, elem.GetContentS(), object)
		if k != "" {
			zc[k] = v
		}
	}

	for _, child := range elem.GetChildren() {
		ParseZapiCounters(child, newPath, object, zc)
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
		return key, strings.Join([]string{object, display}, "_")
	}

	return "", ""
}

func VisitRestTemplates(dir string, objects map[string]bool, client *rest.Client, eachTemp func(path string, client *rest.Client) map[string]Counter) map[string]Counter {
	result := make(map[string]Counter)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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

		if objects != nil && !objects[filepath.Base(path)] {
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

func GetObjects(defaultPath string, objects map[string]bool) {
	t, err := tree.ImportYaml(defaultPath)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", defaultPath, err)
		return
	}

	if err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", defaultPath, err)
		return
	}
	templateObjects := t.GetChildS("objects")
	if templateObjects == nil {
		return
	}
	for _, c := range templateObjects.GetAllChildContentS() {
		objects[c] = true
	}
}

func AddAggregatedCounter(c *Counter, metric plugin.DerivedMetric, withPrefix string, noPrefix string) {
	if !strings.HasSuffix(c.Description, ".") {
		c.Description = c.Description + "."
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

// ProcessRestCounters parse rest and restperf templates
func ProcessRestCounters(client *rest.Client, objects map[string]bool, restPath, restPerfPath string) map[string]Counter {
	restPerfCounters := VisitRestTemplates(restPerfPath, objects, client, func(path string, client *rest.Client) map[string]Counter {
		if _, ok := ExcludePerfTemplates[filepath.Base(path)]; ok {
			return nil
		}
		return processRestPerfCounters(path, client)
	})

	restCounters := VisitRestTemplates(restPath, objects, client, func(path string, client *rest.Client) map[string]Counter {
		return processRestConfigCounters(path)
	})

	for k, v := range restPerfCounters {
		restCounters[k] = v
	}
	return restCounters
}

// processRestConfigCounters process Rest config templates
func processRestConfigCounters(path string) map[string]Counter {
	var (
		counters = make(map[string]Counter)
	)
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

	if templateCounters == nil {
		return nil
	}

	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := util.ParseMetric(c)
			if _, ok := ExcludeCounters[name]; ok {
				continue
			}
			description := searchDescriptionSwagger(model.Object, name)
			harvestName := strings.Join([]string{model.Object, display}, "_")
			if m == "float" {
				co := Counter{
					Name:        harvestName,
					Description: description,
					APIs: []MetricDef{
						{
							API:          "REST",
							Endpoint:     model.Query,
							Template:     path,
							ONTAPCounter: name,
						},
					},
					Visited:  false,
					Template: path,
				}
				counters[harvestName] = co

				// If the template has any MultiplierMetrics, add them
				for _, metric := range model.MultiplierMetrics {
					mc := co
					AddAggregatedCounter(&mc, metric, harvestName, display)
					mc.Visited = false
					mc.Template = path
					counters[mc.Name] = mc
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
			Visited:  false,
			Template: path,
		}
		counters[co.Name] = co
	}
	return counters
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
	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := util.ParseMetric(c)
			if m == "float" {
				counterMap[name] = strings.Join([]string{model.Object, display}, "_")
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
	counterSchema.ForEach(func(key, r gjson.Result) bool {
		if !r.IsObject() {
			return true
		}
		ontapCounterName := r.Get("name").String()
		if _, ok := ExcludeCounters[ontapCounterName]; ok {
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
				Visited:  false,
				Template: path,
			}
			counters[c.Name] = c

			// If the template has any MultiplierMetrics, add them
			for _, metric := range model.MultiplierMetrics {
				mc := c
				AddAggregatedCounter(&mc, metric, v, counterMapNoPrefix[ontapCounterName])
				mc.Visited = false
				mc.Template = path
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
			Visited:  false,
			Template: path,
		}
		counters[co.Name] = co
	}
	// handling for templates with common object names/metric name
	if SpecialPerfObjects[model.Object] {
		return SpecialHandlingPerfCounters(counters, model)
	}
	return counters
}

func SpecialHandlingPerfCounters(counters map[string]Counter, model template2.Model) map[string]Counter {
	// handling for templates with common object names
	modifiedCounters := make(map[string]Counter)
	for originalKey, value := range counters {
		modifiedKey := model.Name + "#" + originalKey
		modifiedCounters[modifiedKey] = value
	}
	return modifiedCounters
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
	t := gjson.GetBytes(SwaggerBytes, searchQuery)
	return UpdateDescription(t.String())
}

func UpdateDescription(description string) string {
	s := replacer.Replace(description)
	return s
}

// ProcessZapiCounters parse zapi and zapiperf templates
func ProcessZapiCounters(client *zapi.Client, objects map[string]bool, zapiPath, zapiPerfPath string) map[string]Counter {
	zapiCounters := VisitZapiTemplates(zapiPath, objects, client, func(path string, client *zapi.Client) map[string]Counter {
		return processZapiConfigCounters(path)
	})
	zapiPerfCounters := VisitZapiTemplates(zapiPerfPath, objects, client, func(path string, client *zapi.Client) map[string]Counter {
		if _, ok := ExcludePerfTemplates[filepath.Base(path)]; ok {
			return nil
		}
		return processZAPIPerfCounters(path, client)
	})

	for k, v := range zapiPerfCounters {
		zapiCounters[k] = v
	}
	return zapiCounters
}

// processZAPIPerfCounters process ZapiPerf counters
func processZAPIPerfCounters(path string, client *zapi.Client) map[string]Counter {
	var (
		counters           = make(map[string]Counter)
		request, response  *node.Node
		zapiUnitMap        = make(map[string]string)
		zapiTypeMap        = make(map[string]string)
		zapiDescMap        = make(map[string]string)
		zapiBaseCounterMap = make(map[string]string)
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
			if name := counter.GetChildContentS("name"); name != "" {
				ty := counter.GetChildContentS("properties")
				if override != nil {
					oty := override.GetChildContentS(name)
					if oty != "" {
						ty = oty
					}
				}
				zapiUnitMap[name] = counter.GetChildContentS("unit")
				zapiDescMap[name] = UpdateDescription(counter.GetChildContentS("desc"))
				zapiTypeMap[name] = ty
				zapiBaseCounterMap[name] = counter.GetChildContentS("base-counter")
			}
		}
	}

	if templateCounters == nil {
		return nil
	}

	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := util.ParseMetric(c)
			if strings.HasPrefix(display, model.Object) {
				display = strings.TrimPrefix(display, model.Object)
				display = strings.TrimPrefix(display, "_")
			}
			harvestName := strings.Join([]string{model.Object, display}, "_")
			if m == "float" {
				if _, ok := ExcludeCounters[name]; ok {
					continue
				}
				// uuid field type is "string,no-display"
				if !(zapiTypeMap[name] == "string" || zapiTypeMap[name] == "string,no-display") {
					co := Counter{
						Name:        harvestName,
						Description: zapiDescMap[name],
						APIs: []MetricDef{
							{
								API:          "ZAPI",
								Endpoint:     strings.Join([]string{"perf-object-get-instances", model.Query}, " "),
								Template:     path,
								ONTAPCounter: name,
								Unit:         zapiUnitMap[name],
								Type:         zapiTypeMap[name],
								BaseCounter:  zapiBaseCounterMap[name],
							},
						},
						Visited:  false,
						Template: path,
					}
					counters[harvestName] = co

					// If the template has any MultiplierMetrics, add them
					for _, metric := range model.MultiplierMetrics {
						mc := co
						AddAggregatedCounter(&mc, metric, harvestName, name)
						mc.Visited = false
						mc.Template = path
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
			Visited:  false,
			Template: path,
		}
		counters[co.Name] = co
	}
	// handling for templates with common object names
	if SpecialPerfObjects[model.Object] {
		return SpecialHandlingPerfCounters(counters, model)
	}
	return counters
}

func processZapiConfigCounters(path string) map[string]Counter {
	var (
		counters = make(map[string]Counter)
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
	if model.ExportData == "false" && noExtraMetrics {
		return nil
	}
	if templateCounters == nil {
		return nil
	}

	zc := make(map[string]string)

	for _, c := range templateCounters.GetChildren() {
		ParseZapiCounters(c, []string{}, model.Object, zc)
	}

	for k, v := range zc {
		if _, ok := ExcludeCounters[k]; ok {
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
			Visited:  false,
			Template: path,
		}
		counters[k] = co

		// If the template has any MultiplierMetrics, add them
		for _, metric := range model.MultiplierMetrics {
			mc := co
			AddAggregatedCounter(&mc, metric, co.Name, model.Object)
			mc.Visited = false
			mc.Template = path
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
			Visited:  false,
			Template: path,
		}
		counters[co.Name] = co
	}
	return counters
}

func VisitZapiTemplates(dir string, objects map[string]bool, client *zapi.Client, eachTemp func(path string, client *zapi.Client) map[string]Counter) map[string]Counter {
	result := make(map[string]Counter)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal("failed to read directory:", err)
		}
		ext := filepath.Ext(path)
		if ext != ".yaml" {
			return nil
		}

		if objects != nil && !objects[filepath.Base(path)] {
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
