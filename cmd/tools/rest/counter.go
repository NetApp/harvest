package rest

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

var (
	replacer         = strings.NewReplacer("\n", "", ":", "")
	objectSwaggerMap = map[string]string{
		"volume":             "xc_volume",
		"aggr":               "xc_aggregate",
		"net_port":           "xc_broadcast_domain",
		"environment_sensor": "sensors",
	}
	swaggerBytes []byte
)

type Counters struct {
	C []Counter `yaml:"counters"`
}

type CounterTemplate struct {
	Counters []Counter
}

type Counter struct {
	Name             string `yaml:"Name"`
	Description      string `yaml:"Description"`
	RestAPI          string `yaml:"RestAPI"`
	RestONTAPCounter string `yaml:"RestONTAPCounter"`
	RestTemplate     string `yaml:"RestTemplate"`
	RestUnit         string `yaml:"RestUnit"`
	RestType         string `yaml:"RestType"`
	ZapiAPI          string `yaml:"ZapiAPI"`
	ZapiONTAPCounter string `yaml:"ZapiONTAPCounter"`
	ZapiTemplate     string `yaml:"ZapiTemplate"`
	ZapiUnit         string `yaml:"ZapiUnit"`
	ZapiType         string `yaml:"ZapiType"`
}

// readSwaggerJSON sownloads poller swagger and convert to json format
func readSwaggerJSON() []byte {
	targetPath := "swagger.json"
	path, err := readOrDownloadSwagger()
	if err != nil {
		log.Fatal("failed to download swagger:", err)
		return nil
	}
	cmd := "cat " + path + " | dasel -r yaml -w json > " + targetPath

	_, err = exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatal("Failed to execute command:", cmd, err)
		return nil
	}
	f, err := os.ReadFile(targetPath)
	if err != nil {
		log.Fatal("error while reading swagger", err)
		return nil
	} else {
		return f
	}
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
func processRestCounters(client *Client) map[string]Counter {
	restPerfCounters := visitRestTemplates("conf/restperf", client, func(path string, client *Client) map[string]Counter {
		return processRestPerfCounters(path, client)
	})

	restCounters := visitRestTemplates("conf/rest", client, func(path string, client *Client) map[string]Counter {
		return processRestConfigCounters(path, client)
	})

	for k, v := range restPerfCounters {
		restCounters[k] = v
	}
	return restCounters
}

// processZapiCounters parse zapi and zapiperf templates
func processZapiCounters(client *zapi.Client) map[string]Counter {
	zapiCounters := visitZapiTemplates("conf/zapi/cdot", client, func(path string, client *zapi.Client) map[string]Counter {
		return processZapiConfigCounters(path, client)
	})
	zapiPerfCounters := visitZapiTemplates("conf/zapiperf/cdot", client, func(path string, client *zapi.Client) map[string]Counter {
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

	fullPath = append(path, name)
	key = strings.Join(fullPath, ".")
	if display == "" {
		display = util.ParseZAPIDisplay(object, fullPath)
	}

	if content[0] != '^' {
		return key, strings.Join([]string{object, display}, "_")
	}

	return "", ""
}

// processRestConfigCounters process Rest config templates
func processRestConfigCounters(path string, client *Client) map[string]Counter {
	var (
		counters = make(map[string]Counter)
	)
	t, err := tree.ImportYaml(path)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty\n", path)
		return nil
	}

	query := t.GetChildContentS("query")
	object := t.GetChildContentS("object")
	templateCounters := t.GetChildS("counters")
	if templateCounters == nil {
		return nil
	}

	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := util.ParseMetric(c)
			description := searchDescriptionSwagger(object, name)
			harvestName := strings.Join([]string{object, display}, "_")
			if m == "float" {
				co := Counter{
					Name:             harvestName,
					RestONTAPCounter: name,
					RestTemplate:     path,
					RestAPI:          query,
					Description:      description,
				}
				counters[harvestName] = co
			}
		}
	}
	return counters
}

// processZAPIPerfCounters process ZapiPerf counters
func processZAPIPerfCounters(path string, client *zapi.Client) map[string]Counter {
	var (
		counters          = make(map[string]Counter)
		request, response *node.Node
		zapiUnitMap       = make(map[string]string)
		zapiTypeMap       = make(map[string]string)
		zapiDescMap       = make(map[string]string)
	)
	t, err := tree.ImportYaml(path)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty\n", path)
		return nil
	}

	query := t.GetChildContentS("query")
	object := t.GetChildContentS("object")
	templateCounters := t.GetChildS("counters")
	override := t.GetChildS("override")

	// build request
	request = node.NewXMLS("perf-object-counter-list-info")
	request.NewChildS("objectname", query)

	if err = client.BuildRequest(request); err != nil {
		fmt.Printf("error while building request %+v\n", err)
		return nil
	}

	if response, err = client.Invoke(); err != nil {
		fmt.Printf("error while invoking api %+v\n", err)
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
				zapiDescMap[name] = counter.GetChildContentS("desc")
				zapiTypeMap[name] = ty
			}
		}
	}

	if templateCounters == nil {
		return nil
	}

	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := util.ParseMetric(c)
			if strings.HasPrefix(display, object) {
				display = strings.TrimPrefix(display, object)
				display = strings.TrimPrefix(display, "_")
			}
			harvestName := strings.Join([]string{object, display}, "_")
			if m == "float" {
				if zapiTypeMap[name] != "string" {
					co := Counter{
						Name:             harvestName,
						Description:      zapiDescMap[name],
						ZapiONTAPCounter: name,
						ZapiTemplate:     path,
						ZapiAPI:          strings.Join([]string{"perf-object-get-instances", query}, " "),
						ZapiType:         zapiTypeMap[name],
						ZapiUnit:         zapiUnitMap[name],
					}
					counters[harvestName] = co
				}
			}
		}
	}
	return counters
}

func processZapiConfigCounters(path string, client *zapi.Client) map[string]Counter {
	var (
		counters = make(map[string]Counter)
	)
	t, err := tree.ImportYaml(path)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty\n", path)
		return nil
	}

	query := t.GetChildContentS("query")
	object := t.GetChildContentS("object")
	templateCounters := t.GetChildS("counters")
	if templateCounters == nil {
		return nil
	}

	zc := make(map[string]string)

	for _, c := range templateCounters.GetChildren() {
		parseZapiCounters(c, []string{}, object, zc)
	}

	for k, v := range zc {
		co := Counter{
			Name:             k,
			ZapiONTAPCounter: v,
			ZapiTemplate:     path,
			ZapiAPI:          query,
		}
		counters[k] = co
	}
	return counters
}

func visitRestTemplates(dir string, client *Client, eachTemp func(path string, client *Client) map[string]Counter) map[string]Counter {
	result := make(map[string]Counter)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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

func visitZapiTemplates(dir string, client *zapi.Client, eachTemp func(path string, client *zapi.Client) map[string]Counter) map[string]Counter {
	result := make(map[string]Counter)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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

func generateCounterTemplate(counters map[string]Counter) {
	t, err := template.New("counter.tmpl").ParseFiles("cmd/tools/rest/counter.tmpl")
	if err != nil {
		panic(err)
	}
	var out *os.File
	out, err = os.Create("docs/metrics.md")
	if err != nil {
		panic(err)
	}

	keys := make([]string, 0, len(counters))

	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var Values []Counter
	for _, k := range keys {
		if k == "" {
			continue
		}
		Values = append(Values, counters[k])
		temp := counters[k]
		if temp.ZapiONTAPCounter == "" {
			fmt.Printf("Missing zapi mapping for %v \n", temp)
		}
		if temp.RestONTAPCounter == "" {
			fmt.Printf("Missing Rest mapping for %v \n", temp)
		}

		if temp.Description == "" {
			fmt.Printf("Missing Description for %v \n", temp)
		}
	}

	c := CounterTemplate{}
	c.Counters = Values

	err = t.Execute(out, c)
	if err != nil {
		panic(err)
	}
}

func mergeRestZapiCounters(restCounters map[string]Counter, zapiCounters map[string]Counter) map[string]Counter {
	for k, v := range zapiCounters {
		if v1, ok := restCounters[k]; ok {
			v1.ZapiAPI = v.ZapiAPI
			v1.ZapiTemplate = v.ZapiTemplate
			v1.ZapiONTAPCounter = v.ZapiONTAPCounter
			v1.ZapiUnit = v.ZapiUnit
			v1.ZapiType = v.ZapiType
			restCounters[k] = v1
		} else {
			if v.ZapiONTAPCounter == "instance_name" || v.ZapiONTAPCounter == "instance_uuid" {
				continue
			}
			co := Counter{
				Name:             v.Name,
				Description:      v.Description,
				ZapiAPI:          v.ZapiAPI,
				ZapiTemplate:     v.ZapiTemplate,
				ZapiONTAPCounter: v.ZapiONTAPCounter,
				ZapiUnit:         v.ZapiUnit,
				ZapiType:         v.ZapiType,
			}
			restCounters[v.Name] = co
		}
	}
	return restCounters
}

func processRestPerfCounters(path string, client *Client) map[string]Counter {
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

	query := t.GetChildContentS("query")
	object := t.GetChildContentS("object")
	templateCounters := t.GetChildS("counters")
	override := t.GetChildS("override")

	if templateCounters == nil {
		return nil
	}
	counterMap := make(map[string]string)
	for _, c := range templateCounters.GetAllChildContentS() {
		if c != "" {
			name, display, m, _ := util.ParseMetric(c)
			if m == "float" {
				counterMap[name] = strings.Join([]string{object, display}, "_")
			}
		}
	}
	href := BuildHref(query, "", nil, "", "", "", "", query)
	records, err = Fetch(client, href)
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
				Name:             v,
				RestONTAPCounter: ontapCounterName,
				RestTemplate:     path,
				RestAPI:          query,
				Description:      description,
				RestType:         ty,
				RestUnit:         r.Get("unit").String(),
			}
			counters[c.Name] = c
		}
		return true
	})
	return counters
}

func ProcessExternalCounters(counters map[string]Counter) map[string]Counter {
	dat, err := os.ReadFile("cmd/tools/rest/Counter.yaml")
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
			co := Counter{
				Name:             v.Name,
				Description:      v.Description,
				RestONTAPCounter: v.RestONTAPCounter,
				RestTemplate:     v.RestTemplate,
				RestAPI:          v.RestAPI,
				RestType:         v.RestType,
				RestUnit:         v.RestUnit,
				ZapiONTAPCounter: v.ZapiONTAPCounter,
				ZapiTemplate:     v.ZapiTemplate,
				ZapiAPI:          v.ZapiAPI,
				ZapiType:         v.ZapiType,
				ZapiUnit:         v.ZapiUnit,
			}
			counters[v.Name] = co
		} else {
			if v.Description != "" {
				v1.Description = v.Description
			}
			if v.RestONTAPCounter != "" {
				v1.RestONTAPCounter = v.RestONTAPCounter
			}
			if v.RestTemplate != "" {
				v1.RestTemplate = v.RestTemplate
			}
			if v.RestAPI != "" {
				v1.RestAPI = v.RestAPI
			}
			if v.RestType != "" {
				v1.RestType = v.RestType
			}
			if v.RestUnit != "" {
				v1.RestUnit = v.RestUnit
			}
			if v.ZapiONTAPCounter != "" {
				v1.ZapiONTAPCounter = v.ZapiONTAPCounter
			}
			if v.ZapiTemplate != "" {
				v1.ZapiTemplate = v.ZapiTemplate
			}
			if v.ZapiAPI != "" {
				v1.ZapiAPI = v.ZapiAPI
			}
			if v.ZapiType != "" {
				v1.ZapiType = v.ZapiType
			}
			if v.ZapiUnit != "" {
				v1.ZapiUnit = v.ZapiUnit
			}
			counters[v.Name] = v1
		}
	}
	return counters
}
