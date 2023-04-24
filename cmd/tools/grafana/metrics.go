package grafana

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var varRe = regexp.MustCompile(`\$(\w+)`)
var metricRe = regexp.MustCompile(`(\w+)\{`)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "print dashboard metrics",
	Run:   doMetrics,
}

func doMetrics(_ *cobra.Command, _ []string) {
	adjustOptions()
	validateImport()
	visitDashboards([]string{opts.dir}, func(path string, data []byte) {
		visitExpressionsAndQueries(path, data)
	})
}

type exprP struct {
	path string
	expr string
	vars []string
}

func visitExpressionsAndQueries(path string, data []byte) {
	// collect all expressions
	expressions := make([]exprP, 0)
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		doTarget("", key, value, func(path string, expr string, format string) {
			expressions = append(expressions, newExpr(path, expr))
		})
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			pathPrefix := fmt.Sprintf("panels[%d].", key.Int())
			doTarget(pathPrefix, key2, value2, func(path string, expr string, format string) {
				expressions = append(expressions, newExpr(path, expr))
			})
			return true
		})
		return true
	})

	metricsSeen := make(map[string]struct{})
	for _, expr := range expressions {
		allMatches := metricRe.FindAllStringSubmatch(expr.expr, -1)
		for _, match := range allMatches {
			m := match[1]
			if len(m) == 0 {
				continue
			}
			metricsSeen[m] = struct{}{}
		}
	}

	// collect all variables
	variables := allVariables(data)
	for _, v := range variables {
		allMatches := metricRe.FindAllStringSubmatch(v.query, -1)
		for _, match := range allMatches {
			m := match[1]
			if len(m) == 0 {
				continue
			}
			metricsSeen[m] = struct{}{}
		}
	}

	fmt.Printf("%s\n", shortPath(path))
	metrics := setToList(metricsSeen)
	for _, metric := range metrics {
		fmt.Printf("- %s\n", metric)
	}
	fmt.Println()
}

func setToList(seen map[string]struct{}) []string {
	list := make([]string, 0)
	for k := range seen {
		list = append(list, k)
	}
	sort.Strings(list)
	return list
}

type variable struct {
	name    string
	kind    string
	query   string
	refresh string
	path    string
}

func allVariables(data []byte) []variable {
	variables := make([]variable, 0)
	gjson.GetBytes(data, "templating.list").ForEach(func(key, value gjson.Result) bool {
		// The datasource variable can be ignored
		if value.Get("type").String() == "datasource" {
			return true
		}

		variables = append(variables, variable{
			name:    value.Get("name").String(),
			kind:    value.Get("type").String(),
			query:   value.Get("query.query").String(),
			refresh: value.Get("refresh").String(),
			path:    key.String(),
		})
		return true
	})
	return variables
}

func newExpr(path string, expr string) exprP {
	allMatches := varRe.FindAllStringSubmatch(expr, -1)
	vars := make([]string, 0, len(allMatches))
	for _, match := range allMatches {
		vars = append(vars, match[1])
	}
	return exprP{
		path: path,
		expr: expr,
		vars: vars,
	}
}

func doTarget(pathPrefix string, key gjson.Result, value gjson.Result,
	exprFunc func(path string, expr string, format string)) {
	kind := value.Get("type").String()
	if kind == "row" {
		return
	}
	path := fmt.Sprintf("%spanels[%d]", pathPrefix, key.Int())
	targetsSlice := value.Get("targets").Array()
	for i, targetN := range targetsSlice {
		expr := targetN.Get("expr").String()
		pathWithTarget := path + ".targets[" + strconv.Itoa(i) + "]"
		exprFunc(pathWithTarget, expr, kind)
	}
}

func visitDashboards(dirs []string, eachDash func(path string, data []byte)) {
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if strings.Contains(path, "influxdb") {
				return nil
			}
			if err != nil {
				log.Fatal("failed to read directory:", err)
			}
			ext := filepath.Ext(path)
			if ext != ".json" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				log.Fatalf("failed to read dashboards path=%s err=%v", path, err)
			}
			eachDash(path, data)
			return nil
		})
		if err != nil {
			log.Fatal("failed to read dashboards:", err)
		}
	}
}

func shortPath(dashPath string) string {
	splits := strings.Split(dashPath, string(filepath.Separator))
	return strings.Join(splits[len(splits)-2:], string(filepath.Separator))
}
