package grafana

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
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
		[]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
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

			for i := range len(newExpressions) {
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
			file := filepath.Base(path)
			out := addSvmRegex(data, file, regex)
			if file == "svm.json" {
				r := gjson.GetBytes(out, "templating.list.#(name=\"SVM\").regex")
				if r.String() != regex {
					t.Errorf("path: %s \nExpected: [%s]\n     Got: [%s]", path, regex, r.String())
				}
			} else if file == "snapmirror.json" {
				r := gjson.GetBytes(out, "templating.list.#(name=\"DestinationSVM\").regex")
				if r.String() != regex {
					t.Errorf("path: %s \nExpected: [%s]\n     Got: [%s]", path, regex, r.String())
				}
				r = gjson.GetBytes(out, "templating.list.#(name=\"SourceSVM\").regex")
				if r.String() != regex {
					t.Errorf("path: %s \nExpected: [%s]\n     Got: [%s]", path, regex, r.String())
				}
			}
		})
}

func getExp(expr string, expressions *[]string) {
	regex := regexp.MustCompile(`([a-zA-Z0-9_+-]+)\s?{.+?}`)
	match := regex.FindAllStringSubmatch(expr, -1)
	for _, m := range match {
		// multiple metrics used with `+`
		if strings.Contains(m[1], "+") {
			submatch := strings.Split(m[1], "+")
			*expressions = append(*expressions, submatch...)
			// multiple metrics used with `-`
		} else if strings.Contains(m[1], "-") {
			submatch := strings.Split(m[1], "-")
			*expressions = append(*expressions, submatch...)
			// single metric
		} else {
			*expressions = append(*expressions, m[1])
		}
	}
}

func readExprs(data []byte) []string {
	expressions := make([]string, 0)
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		doExpr("", key, value, func(anExpr expression) {
			getExp(anExpr.metric, &expressions)
		})
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			pathPrefix := fmt.Sprintf("panels[%d].", key.Int())
			doExpr(pathPrefix, key2, value2, func(anExpr expression) {
				getExp(anExpr.metric, &expressions)
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
			wrappedInDef := fmt.Sprintf(`"definition": "%s"`, tt.json)
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
			"prometheus": map[string]any{
				"type": DefaultDataSource,
			},
		},
	}
	multiPrometheusDSWithSameDS := map[string]any{
		"datasources": map[string]any{
			"Influx": map[string]any{
				"type": "influxdb",
			},
			"prometheus": map[string]any{
				"type": DefaultDataSource,
			},
			"NetProm": map[string]any{
				"type": DefaultDataSource,
			},
		},
	}
	multiPrometheusDSWithOtherDS := map[string]any{
		"datasources": map[string]any{
			"Influx": map[string]any{
				"type": "influxdb",
			},
			"prometheus": map[string]any{
				"type": DefaultDataSource,
			},
			"NetProm": map[string]any{
				"type": DefaultDataSource,
			},
		},
	}

	tests := []test{
		{name: "empty", result: nil, dsArg: DefaultDataSource, want: false},
		{name: "nil datasource", result: noDS, dsArg: DefaultDataSource, want: false},
		{name: "non prometheus datasource", result: nonPrometheusDS, dsArg: DefaultDataSource, want: false},
		{name: "valid prometheus datasource", result: defaultPrometheusDS, dsArg: DefaultDataSource, want: true},
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
