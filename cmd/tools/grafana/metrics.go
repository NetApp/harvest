package grafana

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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
	VisitDashboards([]string{opts.dir}, func(path string, data []byte) {
		visitExpressionsAndQueries(path, data)
	})
}

type exprP struct {
	path       string
	expr       string
	vars       []string
	panelTitle string
}

func visitExpressionsAndQueries(path string, data []byte) {
	// collect all expressions
	expressions := make([]exprP, 0)
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		doTarget("", key, value, func(path string, expr string, _ string, title string) {
			expressions = append(expressions, newExpr(path, expr, title))
		})
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			pathPrefix := fmt.Sprintf("panels[%d].", key.Int())
			doTarget(pathPrefix, key2, value2, func(path string, expr string, _ string, title string) {
				expressions = append(expressions, newExpr(path, expr, title))
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

	fmt.Printf("%s\n", ShortPath(path))
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
	options []gjson.Result
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
			options: value.Get("options").Array(),
			path:    key.String(),
		})
		return true
	})
	return variables
}

func newExpr(path string, expr string, title string) exprP {
	allMatches := varRe.FindAllStringSubmatch(expr, -1)
	vars := make([]string, 0, len(allMatches))
	for _, match := range allMatches {
		vars = append(vars, match[1])
	}
	return exprP{
		path:       path,
		expr:       expr,
		vars:       vars,
		panelTitle: title,
	}
}

func doTarget(pathPrefix string, key gjson.Result, value gjson.Result,
	exprFunc func(path string, expr string, format string, title string)) {
	kind := value.Get("type").String()
	title := value.Get("title").String()
	if kind == "row" {
		return
	}
	path := fmt.Sprintf("%spanels[%d]", pathPrefix, key.Int())
	targetsSlice := value.Get("targets").Array()
	for i, targetN := range targetsSlice {
		expr := targetN.Get("expr").String()
		pathWithTarget := path + ".targets[" + strconv.Itoa(i) + "]"
		exprFunc(pathWithTarget, expr, kind, title)
	}
}

func VisitDashboards(dirs []string, eachDash func(path string, data []byte)) {
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
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

func VisitAllPanels(data []byte, handle func(path string, key gjson.Result, value gjson.Result)) {
	visitPanels(data, "panels", "", handle)
}

func visitPanels(data []byte, panelPath string, pathPrefix string, handle func(path string, key gjson.Result, value gjson.Result)) {
	gjson.GetBytes(data, panelPath).ForEach(func(key, value gjson.Result) bool {
		path := fmt.Sprintf("%s[%d]", panelPath, key.Int())
		fullPath := fmt.Sprintf("%s%s", pathPrefix, path)
		handle(fullPath, key, value)
		visitPanels([]byte(value.Raw), "panels", fullPath, handle)
		return true
	})
}

// ShortPath turns ../../../grafana/dashboards/cmode/aggregate.json into cmode/aggregate.json
func ShortPath(dashPath string) string {
	splits := strings.Split(dashPath, string(filepath.Separator))

	// Join the elements after "dashboards"
	index := slices.Index(splits, "dashboards")
	if index == -1 || index+1 >= len(splits) {
		return dashPath
	}
	return strings.Join(splits[index+1:], string(filepath.Separator))
}
