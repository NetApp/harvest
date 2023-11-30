package generate

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/cmd/tools/utils"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"sort"
	"strings"
	"text/template"
	"time"
)

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

func generateCounterTemplate(counters map[string]utils.Counter, version [3]int) {
	targetPath := "docs/ontap-metrics.md"
	t, err := template.New("counter.tmpl").ParseFiles("cmd/tools/generate/counter.tmpl")
	if err != nil {
		panic(err)
	}
	var out *os.File
	out, err = os.Create(targetPath)
	if err != nil {
		panic(err)
	}

	keys := make([]string, 0, len(counters))

	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var values []utils.Counter
	for _, k := range keys {
		if k == "" {
			continue
		}
		counter := counters[k]

		// Print such counters which are missing Rest mapping
		if len(counter.APIs) == 1 {
			if counter.APIs[0].API == "ZAPI" {
				isPrint := true
				for _, substring := range utils.ExcludeLogRestCounters {
					if strings.HasPrefix(counter.Name, substring) {
						isPrint = false
						break
					}
				}
				// missing Rest Mapping
				if isPrint {
					fmt.Printf("Missing %s mapping for %v \n", "REST", counter)
				}
			}
		}

		values = append(values, counter)
		for _, def := range counter.APIs {
			if def.ONTAPCounter == "" {
				fmt.Printf("Missing %s mapping for %v \n", def.API, counter)
			}
		}
		if counter.Description == "" {
			fmt.Printf("Missing Description for %v \n", counter)
		}
	}

	verWithDots := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(version)), "."), "[]")
	c := utils.CounterTemplate{
		Counters: values,
		CounterMetaData: utils.CounterMetaData{
			Date:         time.Now().Format("2006-Jan-02"),
			OntapVersion: verWithDots,
		},
	}

	err = t.Execute(out, c)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Harvest metric documentation generated at %s \n", targetPath)
}

// Regex to match NFS version and operation
var reRemove = regexp.MustCompile(`NFSv\d+\.\d+`)

func mergeCounters(restCounters map[string]utils.Counter, zapiCounters map[string]utils.Counter) map[string]utils.Counter {
	// handle special counters
	restKeys := sortedKeys(restCounters)
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

	zapiKeys := sortedKeys(zapiCounters)
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
	zapiKeys = sortedKeys(zapiCounters)
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
			co := utils.Counter{
				Name:        v.Name,
				Description: v.Description,
				APIs:        []utils.MetricDef{zapiDef},
			}
			restCounters[v.Name] = co
		}
	}
	return restCounters
}

func sortedKeys(m map[string]utils.Counter) []string {
	keys := maps.Keys(m)
	slices.Sort(keys)
	return keys
}

func processExternalCounters(counters map[string]utils.Counter) map[string]utils.Counter {
	dat, err := os.ReadFile("cmd/tools/generate/counter.yaml")
	if err != nil {
		fmt.Printf("error while reading file %v", err)
		return nil
	}
	var c utils.Counters

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
				r := findAPI(v1.APIs, m)
				if r == nil {
					v1.APIs = append(v1.APIs, m)
				} else {
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
			counters[v.Name] = v1
		}
	}
	return counters
}

func findAPI(apis []utils.MetricDef, other utils.MetricDef) *utils.MetricDef {
	for _, a := range apis {
		if a.API == other.API {
			return &a
		}
	}
	return nil
}
