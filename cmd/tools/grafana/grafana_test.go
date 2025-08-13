package grafana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TestCheckVersion(t *testing.T) {
	type vCheck struct {
		version string
		want    bool
	}
	checks := []vCheck{
		{version: "8.4.0-beta1", want: true},
		{version: "7.2.3.4", want: true},
		{version: "abc.1.3", want: false},
		{version: "4.5.4", want: false},
		{version: "7.1.0", want: true},
		{version: "7.5.5", want: true},
	}
	for _, check := range checks {
		got := checkVersion(check.version)
		if got != check.want {
			t.Errorf("Expected %t but got %t for input %s", check.want, got, check.version)
		}
	}
	t.Log("") // required so the test is not marked as terminated
}

func TestHttpsAddr(t *testing.T) {
	opts.addr = "https://1.1.1.1:3000"
	opts.dir = "../../../grafana/dashboards"
	opts.config = "../doctor/testdata/testConfig.yml"
	adjustOptions()
	if opts.addr != "https://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "https://1.1.1.1:3000", opts.addr)
	}

	opts.addr = "https://1.1.1.1:3000"
	opts.useHTTPS = false // addr takes precedence over useHTTPS
	adjustOptions()
	if opts.addr != "https://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "https://1.1.1.1:3000", opts.addr)
	}

	opts.addr = "http://1.1.1.1:3000"
	adjustOptions()
	if opts.addr != "http://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "http://1.1.1.1:3000", opts.addr)
	}

	// Old way of specifying https
	opts.addr = "http://1.1.1.1:3000"
	opts.useHTTPS = true
	adjustOptions()
	if opts.addr != "https://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "https://1.1.1.1:3000", opts.addr)
	}
}

func TestAddPrefixToMetricNames(t *testing.T) {

	var (
		dashboard                      map[string]any
		oldExpressions, newExpressions []string
		updatedData                    []byte
		err                            error
	)

	prefix := "xx_"
	VisitDashboards(
		[]string{"../../../grafana/dashboards/cisco", "../../../grafana/dashboards/cmode", "../../../grafana/dashboards/cmode-details", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			oldExpressions = readExprs(data)
			if err = json.Unmarshal(data, &dashboard); err != nil {
				fmt.Printf("error parsing file [%s] %+v\n", path, err)
				fmt.Println("-------------------------------")
				fmt.Println(string(data))
				fmt.Println("-------------------------------")
				return
			}
			addGlobalPrefix(dashboard, prefix)

			if updatedData, err = json.Marshal(dashboard); err != nil {
				fmt.Printf("error parsing file [%s] %+v\n", path, err)
				fmt.Println("-------------------------------")
				fmt.Println(string(updatedData))
				fmt.Println("-------------------------------")
				return
			}
			newExpressions = readExprs(updatedData)

			for i := range newExpressions {
				if newExpressions[i] != prefix+oldExpressions[i] {
					t.Errorf("path: %s \nExpected: [%s]\n     Got: [%s]", path, prefix+oldExpressions[i], newExpressions[i])
				}
			}
		})
}

func TestAddSvmRegex(t *testing.T) {

	regex := ".*ABC.*"
	VisitDashboards(
		[]string{"../../../grafana/dashboards/cmode/svm.json", "../../../grafana/dashboards/cmode/snapmirror.json"},
		func(path string, data []byte) {
			out := addSvmRegex(data, regex)
			r := gjson.GetBytes(out, "templating.list.#(name=\"SVM\").regex")
			if r.ClonedString() != regex {
				t.Errorf("path: %s \nExpected: [%s]\n     Got: [%s]", path, regex, r.ClonedString())
			}
		})
}

func getExp(expr string, expressions *[]string) {
	regex := regexp.MustCompile(`([a-zA-Z0-9_+-]+)\s?{.+?}`)
	match := regex.FindAllStringSubmatch(expr, -1)
	for _, m := range match {
		// multiple metrics used with `+`
		switch {
		case strings.Contains(m[1], "+"):
			submatch := strings.Split(m[1], "+")
			*expressions = append(*expressions, submatch...)
		case strings.Contains(m[1], "-"):
			submatch := strings.Split(m[1], "-")
			*expressions = append(*expressions, submatch...)
		default:
			*expressions = append(*expressions, m[1])
		}
	}
}

func readExprs(data []byte) []string {
	expressions := make([]string, 0)
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		DoExpr("", key, value, func(anExpr Expression) {
			getExp(anExpr.Metric, &expressions)
		})
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			pathPrefix := fmt.Sprintf("panels[%d].", key.Int())
			DoExpr(pathPrefix, key2, value2, func(anExpr Expression) {
				getExp(anExpr.Metric, &expressions)
			})
			return true
		})
		return true
	})
	return expressions
}

func TestChainedParsing(t *testing.T) {
	type test struct {
		name string
		json string
		want string
	}
	tests := []test{
		{name: "empty", json: "", want: ``},
		{name: "no label", json: "label_values(datacenter)", want: `"definition": "label_values({foo=~"$Foo"}, datacenter)"`},
		{name: "one label", json: "label_values(poller_status, datacenter)",
			want: `"definition": "label_values(poller_status{foo=~"$Foo"}, datacenter)"`},
		{name: "typical", json: `label_values(aggr_new_status{datacenter="$Datacenter",cluster="$Cluster"}, node)`,
			want: `"definition": "label_values(aggr_new_status{datacenter="$Datacenter",cluster="$Cluster",foo=~"$Foo"}, node)"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrappedInDef := `"definition": "` + tt.json + `"`
			got := toChainedVar(wrappedInDef, "foo")
			if got != tt.want {
				t.Errorf("TestChainedParsing\n got=[%v]\nwant=[%v]", got, tt.want)
			}
		})
	}
}

func TestIsValidDatasource(t *testing.T) {
	type test struct {
		name   string
		result map[string]any
		dsArg  string
		want   bool
	}

	const prometheusDS = "prometheus"

	noDS := map[string]any{
		"datasources": nil,
	}
	nonPrometheusDS := map[string]any{
		"datasources": map[string]any{
			"Grafana": map[string]any{
				"type": "dashboard",
			},
			"Influx": map[string]any{
				"type": "influxdb",
			},
		},
	}
	defaultPrometheusDS := map[string]any{
		"datasources": map[string]any{
			"Influx": map[string]any{
				"type": "influxdb",
			},
			prometheusDS: map[string]any{
				"type": prometheusDS,
			},
		},
	}
	legacyPrometheusDS := map[string]any{
		"datasources": map[string]any{
			"Influx": map[string]any{
				"type": "influxdb",
			},
			"Prometheus": map[string]any{
				"type": prometheusDS,
			},
		},
	}
	multiPrometheusDSWithSameDS := map[string]any{
		"datasources": map[string]any{
			"Influx": map[string]any{
				"type": "influxdb",
			},
			prometheusDS: map[string]any{
				"type": prometheusDS,
			},
			"NetProm": map[string]any{
				"type": prometheusDS,
			},
		},
	}
	multiPrometheusDSWithOtherDS := map[string]any{
		"datasources": map[string]any{
			"Influx": map[string]any{
				"type": "influxdb",
			},
			prometheusDS: map[string]any{
				"type": prometheusDS,
			},
			"NetProm": map[string]any{
				"type": prometheusDS,
			},
		},
	}

	tests := []test{
		{name: "empty", result: nil, dsArg: "", want: false},
		{name: "nil datasource", result: noDS, dsArg: "", want: false},
		{name: "non prometheus datasource", result: nonPrometheusDS, dsArg: "influxdb", want: false},
		{name: "valid prometheus datasource", result: defaultPrometheusDS, dsArg: prometheusDS, want: true},
		{name: "legacy valid prometheus datasource", result: legacyPrometheusDS, dsArg: "Prometheus", want: true},
		{name: "multiple prometheus datasource with same datasource given", result: multiPrometheusDSWithSameDS, dsArg: "NetProm", want: true},
		{name: "multiple prometheus datasource with different datasource given", result: multiPrometheusDSWithOtherDS, dsArg: "UpdateProm", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts.datasource = tt.dsArg
			got := isValidDatasource(tt.result)
			if got != tt.want {
				t.Errorf("TestIsValidDatasource\n got=[%v]\nwant=[%v]", got, tt.want)
			}
		})
	}
}

func TestAddLabel(t *testing.T) {
	type test struct {
		name           string
		json           string
		labels         []string
		want           string
		customAllValue string
	}
	tests := []test{
		{
			name:           "includeAll is false",
			labels:         []string{"nate"},
			customAllValue: "boo",
			json: `{
        "allValue": "foo",
        "definition": "label_values(cluster_new_status{system_type!=\"7mode\"},datacenter)",
        "includeAll": false,
        "name": "Datacenter",
        "query": {
          "query": "label_values(cluster_new_status{system_type!=\"7mode\"},datacenter)",
          "refId": "Prometheus-Datacenter-Variable-Query"
        }
      }`,
			want: `{
          "templating": {
            "list": [
              {
                "allValue": "foo",
                "definition": "label_values(cluster_new_status{system_type!=\"7mode\",nate=~\"$Nate\"},datacenter)",
                "includeAll": false,
                "name": "Datacenter",
                "query": {
                  "query": "label_values(cluster_new_status{system_type!=\"7mode\",nate=~\"$Nate\"},datacenter)",
                  "refId": "Prometheus-Datacenter-Variable-Query"
                }
              },
              {
                "allValue": ".*",
                "current": {
                  "selected": false
                },
                "definition": "label_values(nate)",
                "hide": 0,
                "includeAll": true,
                "multi": true,
                "name": "Nate",
                "options": [],
                "query": {
                  "query": "label_values(nate)",
                  "refId": "StandardVariableQuery"
                },
                "refresh": 2,
                "regex": "",
                "skipUrlSync": false,
                "sort": 0,
                "type": "query"
              }
            ]
          }
        }`,
		},
		{
			name: "include all is true no custom all value",
			json: `{
        "allValue": ".*",
        "current": {},
        "datasource": "${DS_PROMETHEUS}",
        "definition": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\"},cluster)",
        "description": null,
        "error": null,
        "hide": 0,
        "includeAll": true,
        "label": "",
        "multi": true,
        "name": "Cluster",
        "options": [],
        "query": {
          "query": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\"},cluster)",
          "refId": "StandardVariableQuery"
        },
        "refresh": 2,
        "regex": "",
        "skipUrlSync": false,
        "sort": 1,
        "tagValuesQuery": "",
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      }`,
			labels: []string{"nate"},
			want: `{
          "templating": {
            "list": [
              {
                "allValue": ".*",
                "current": {},
                "datasource": "${DS_PROMETHEUS}",
                "definition": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\",nate=~\"$Nate\"},cluster)",
                "description": null,
                "error": null,
                "hide": 0,
                "includeAll": true,
                "label": "",
                "multi": true,
                "name": "Cluster",
                "options": [],
                "query": {
                  "query": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\",nate=~\"$Nate\"},cluster)",
                  "refId": "StandardVariableQuery"
                },
                "refresh": 2,
                "regex": "",
                "skipUrlSync": false,
                "sort": 1,
                "tagValuesQuery": "",
                "tagsQuery": "",
                "type": "query",
                "useTags": false
              },
              {
                "allValue": ".*",
                "current": {
                  "selected": false
                },
                "definition": "label_values(nate)",
                "hide": 0,
                "includeAll": true,
                "multi": true,
                "name": "Nate",
                "options": [],
                "query": {
                  "query": "label_values(nate)",
                  "refId": "StandardVariableQuery"
                },
                "refresh": 2,
                "regex": "",
                "skipUrlSync": false,
                "sort": 0,
                "type": "query"
              }
            ]
          }
        }`,
		},
		{
			name:           "include all with null custom all value",
			labels:         []string{"nate"},
			customAllValue: "null",
			json: `{
        "allValue": ".*",
        "current": {},
        "datasource": "${DS_PROMETHEUS}",
        "definition": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\"},cluster)",
        "description": null,
        "error": null,
        "hide": 0,
        "includeAll": true,
        "label": "",
        "multi": true,
        "name": "Cluster",
        "options": [],
        "query": {
          "query": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\"},cluster)",
          "refId": "StandardVariableQuery"
        },
        "refresh": 2,
        "regex": "",
        "skipUrlSync": false,
        "sort": 1,
        "tagValuesQuery": "",
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      }`,
			want: `{
          "templating": {
            "list": [
              {
                "allValue": null,
                "current": {},
                "datasource": "${DS_PROMETHEUS}",
                "definition": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\",nate=~\"$Nate\"},cluster)",
                "description": null,
                "error": null,
                "hide": 0,
                "includeAll": true,
                "label": "",
                "multi": true,
                "name": "Cluster",
                "options": [],
                "query": {
                  "query": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\",nate=~\"$Nate\"},cluster)",
                  "refId": "StandardVariableQuery"
                },
                "refresh": 2,
                "regex": "",
                "skipUrlSync": false,
                "sort": 1,
                "tagValuesQuery": "",
                "tagsQuery": "",
                "type": "query",
                "useTags": false
              },
              {
                "allValue": ".*",
                "current": {
                  "selected": false
                },
                "definition": "label_values(nate)",
                "hide": 0,
                "includeAll": true,
                "multi": true,
                "name": "Nate",
                "options": [],
                "query": {
                  "query": "label_values(nate)",
                  "refId": "StandardVariableQuery"
                },
                "refresh": 2,
                "regex": "",
                "skipUrlSync": false,
                "sort": 0,
                "type": "query"
              }
            ]
          }
        }`,
		},
		{
			name:           "include all with custom all value",
			labels:         []string{"nate"},
			customAllValue: ".*",
			json: `{
        "allValue": null,
        "current": {},
        "datasource": "${DS_PROMETHEUS}",
        "definition": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\"},cluster)",
        "description": null,
        "error": null,
        "hide": 0,
        "includeAll": true,
        "label": "",
        "multi": true,
        "name": "Cluster",
        "options": [],
        "query": {
          "query": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\"},cluster)",
          "refId": "StandardVariableQuery"
        },
        "refresh": 2,
        "regex": "",
        "skipUrlSync": false,
        "sort": 1,
        "tagValuesQuery": "",
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      }`,
			want: `{
          "templating": {
            "list": [
              {
                "allValue": ".*",
                "current": {},
                "datasource": "${DS_PROMETHEUS}",
                "definition": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\",nate=~\"$Nate\"},cluster)",
                "description": null,
                "error": null,
                "hide": 0,
                "includeAll": true,
                "label": "",
                "multi": true,
                "name": "Cluster",
                "options": [],
                "query": {
                  "query": "label_values(cluster_new_status{system_type!=\"7mode\",datacenter=~\"$Datacenter\",nate=~\"$Nate\"},cluster)",
                  "refId": "StandardVariableQuery"
                },
                "refresh": 2,
                "regex": "",
                "skipUrlSync": false,
                "sort": 1,
                "tagValuesQuery": "",
                "tagsQuery": "",
                "type": "query",
                "useTags": false
              },
              {
                "allValue": ".*",
                "current": {
                  "selected": false
                },
                "definition": "label_values(nate)",
                "hide": 0,
                "includeAll": true,
                "multi": true,
                "name": "Nate",
                "options": [],
                "query": {
                  "query": "label_values(nate)",
                  "refId": "StandardVariableQuery"
                },
                "refresh": 2,
                "regex": "",
                "skipUrlSync": false,
                "sort": 0,
                "type": "query"
              }
            ]
          }
        }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrappedInDef := `{"templating": {"list": [` + tt.json + `]}}`

			labelMap := make(map[string]string)
			caser := cases.Title(language.Und)
			for _, label := range tt.labels {
				labelMap[caser.String(label)] = label
			}

			opts.customAllValue = tt.customAllValue

			data := []byte(wrappedInDef)
			for i := len(tt.labels) - 1; i >= 0; i-- {
				data = addLabel(data, tt.labels[i], labelMap)
			}

			formattedGot, err := formatJSON(data)
			if err != nil {
				t.Errorf("TestAddLabel\n failed to format json %v", err)
			}

			formattedWant, err := formatJSON([]byte(tt.want))
			if err != nil {
				t.Errorf("TestAddLabel\n failed to format wanted json %v", err)
			}

			if !bytes.Equal(formattedGot, formattedWant) {
				t.Errorf("TestAddLabel\n got=%v\nwant=%v", string(formattedGot), string(formattedWant))
			}
		})
	}
}

func TestClusterRewrite(t *testing.T) {
	type test struct {
		name    string
		input   string
		want    string
		cluster string
	}

	tests := []test{
		{
			name:    "no cluster label",
			input:   `abc{datacenter=~\"$Datacenter\"})`,
			want:    `abc{datacenter=~\"$Datacenter\"})`,
			cluster: "netapp_cluster",
		},
		{
			name:    "multiple cluster labels",
			input:   `sum by(site_name,cluster,datacenter)(storagegrid_storage_utilization_data_bytes{datacenter=~\"$Datacenter\",cluster=~\"$Cluster\"})`,
			want:    `sum by(site_name,netapp_cluster,datacenter)(storagegrid_storage_utilization_data_bytes{datacenter=~\"$Datacenter\",netapp_cluster=~\"$Cluster\"})`,
			cluster: "netapp_cluster",
		},
		{
			name:    "by cluster label",
			input:   `sum by (cluster) (abc{cluster="$Cluster"}[2m)`,
			want:    `sum by (netapp_cluster) (abc{netapp_cluster="$Cluster"}[2m)`,
			cluster: "netapp_cluster",
		},
		{
			name:    "cluster_new_status should not change",
			input:   `"label_values(cluster_new_status{}, cluster)`,
			want:    `"label_values(cluster_new_status{}, netapp_cluster)`,
			cluster: "netapp_cluster",
		},
		{
			name:    "cluster_new_status should not change 2",
			input:   `cluster_new_status{}, cluster`,
			want:    `cluster_new_status{}, netapp_cluster`,
			cluster: "netapp_cluster",
		},
		{
			name:    "snapmirror var",
			input:   `snapmirror_labels{source_cluster) "source_cluster", "$1", "cluster", "(.*)")`,
			want:    `snapmirror_labels{source_cluster) "source_cluster", "$1", "netapp_cluster", "(.*)")`,
			cluster: "netapp_cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rewriteCluster(tt.input, tt.cluster)
			if got != tt.want {
				t.Errorf("TestClusterLabel\n got=%v\nwant=%v", got, tt.want)
			}
		})
	}
}

func TestValidateVarDefaults(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "Datacenter=DC1,DC2;Cluster=Cluster1", want: true},
		{input: "Datacenter=DC1;Cluster=Cluster1,Cluster2", want: true},
		{input: "Datacenter=DC1,DC2;Cluster=Cluster1;Region=US,EU", want: true},
		{input: "Datacenter=DC1,DC2;Cluster=Cluster1;Region=US,EU;SVM=SAN,NAS", want: true},
		{input: "Datacenter=DC1,DC2", want: true},
		{input: "Datacenter=nane,rtp;Cluster=Cluster2,A250-15-28-29", want: true},
		{input: "Datacenter=nane,rtp;Cluster=Cluster2,A250-15-28#29", want: true},
		{input: "Datacenter=DC1,DC2;", want: false},                  // trailing semicolon
		{input: "=DC1,DC2;Cluster=Cluster1", want: false},            // missing variable name
		{input: "Datacenter=;Cluster=Cluster1", want: false},         // missing value
		{input: "Datacenter=DC1,DC2;Cluster=", want: false},          // missing value
		{input: "Datacenter=DC1,DC2;Cluster", want: false},           // missing equals sign
		{input: "Datacenter=DC1,DC2;Cluster=Cluster1;", want: false}, // trailing semicolon
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := validateVarDefaults(tt.input)
			if got != tt.want {
				t.Errorf("validateVarDefaults(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
