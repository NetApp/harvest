package grafana

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"regexp"
	"strings"
	"testing"
)

func TestCheckVersion(t *testing.T) {

	inputVersion := []string{"8.4.0-beta1", "7.2.3.4", "abc.1.3", "4.5.4", "7.1.0", "7.5.5"}
	expectedOutPut := []bool{true, true, false, false, true, true}

	for i, s := range inputVersion {
		c := checkVersion(s)
		if c != expectedOutPut[i] {
			t.Errorf("Expected %t but got %t for input %s", expectedOutPut[i], c, inputVersion[i])
		}
	}
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
		dashboard                      map[string]interface{}
		oldExpressions, newExpressions []string
		updatedData                    []byte
		err                            error
	)

	prefix := "xx_"
	dir := "../../../grafana/dashboards/cmode"
	visitDashboards(dir, func(path string, data []byte) {
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

		for i := 0; i < len(newExpressions); i++ {
			if newExpressions[i] != prefix+oldExpressions[i] {
				t.Errorf("\nExpected: [%s]\n     Got: [%s]", prefix+oldExpressions[i], newExpressions[i])
			}
		}
	})
}

func getExp(expr string, expressions *[]string) {
	regex := regexp.MustCompile(`([a-zA-Z_+]+)\s?{.+?}`)
	match := regex.FindAllStringSubmatch(expr, -1)
	for _, m := range match {
		// multiple metrics used to summarize
		if strings.Contains(m[1], "+") {
			submatch := strings.Split(m[1], "+")
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
		doExpr("", key, value, func(path string, expr string) {
			getExp(expr, &expressions)
		})
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			pathPrefix := fmt.Sprintf("panels[%d].", key.Int())
			doExpr(pathPrefix, key2, value2, func(path string, expr string) {
				getExp(expr, &expressions)
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
