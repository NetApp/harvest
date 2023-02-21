package grafana

import (
	"fmt"
	"github.com/tidwall/gjson"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestDatasource(t *testing.T) {
	visitDashboards([]string{"../../../grafana/dashboards"}, func(path string, data []byte) {
		checkDashboardForDatasource(t, path, data)
	})
}

func checkDashboardForDatasource(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		dsResult := value.Get("datasource")
		if dsResult.Type == gjson.Null {
			// if the panel is a row, it is OK if there is no datasource
			if value.Get("type").String() == "row" {
				return true
			}
			t.Errorf("dashboard=%s panel=%s has a null datasource", path, key.String())
		}
		return true
	})

	// Check that the variable DS_PROMETHEUS exist
	doesDsPromExist := false
	gjson.GetBytes(data, "templating.list").ForEach(func(key, value gjson.Result) bool {
		if value.Get("name").String() == "DS_PROMETHEUS" {
			doesDsPromExist = true
			query := value.Get("query").String()
			if query != "prometheus" {
				t.Errorf("dashboard=%s var=DS_PROMETHEUS query want=prometheus got=%s", path, query)
			}
			theType := value.Get("type").String()
			if theType != "datasource" {
				t.Errorf("dashboard=%s var=DS_PROMETHEUS type want=datasource got=%s", path, theType)
			}
		}
		return true
	})
	if !doesDsPromExist {
		t.Errorf("dashboard=%s should define variable has DS_PROMETHEUS", path)
	}
}

func TestUnitsAndExprMatch(t *testing.T) {
	mt := newMetricsTable()
	visitDashboards([]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkUnits(path, mt, data)
		})

	// Exceptions are meant to reduce false negatives
	allowedSuffix := map[string][]string{
		"_count":    {"none", "short", "locale"},
		"_lag_time": {"", "s", "short"},
	}

	metricNames := make([]string, 0, len(mt.metricsByUnit))
	for m := range mt.metricsByUnit {
		metricNames = append(metricNames, m)
	}
	sort.Strings(metricNames)

	for _, metric := range metricNames {
		u := mt.metricsByUnit[metric]

		failText := strings.Builder{}
		numUnits := len(u.units)
		for unit, location := range u.units {
			if unit == "" || unit == "none" {
				// Fail this metric if it contains an empty or none unit
				t.Errorf(`%s should not have unit=none %s path=%s title="%s"`,
					metric, location[0].dashboard, location[0].path, location[0].title)
			}
			if numUnits == 1 {
				continue
			}
		locCheck:
			for _, row := range location {
				for suffix, allowedUnits := range allowedSuffix {
					if strings.HasSuffix(metric, suffix) {
						for _, allowedUnit := range allowedUnits {
							if unit == allowedUnit {
								continue locCheck
							}
						}
					}
				}
				failText.WriteString(fmt.Sprintf("unit=%s %s path=%s title=\"%s\"\n",
					unit, row.dashboard, row.path, row.title))
			}
		}
		if failText.Len() > 0 {
			t.Errorf("%s has conflicting units\n%s", metric, failText.String())
		}
	}
}

var metricName = regexp.MustCompile(`(\w+){`)

type metricsTable struct {
	metricsByUnit map[string]*units
}

type override struct {
	id      string
	options string
	unit    string
	path    string
}
type expression struct {
	metric string
	refID  string
	kind   string
}
type units struct {
	units map[string][]*metricLoc
}

func (u *units) addUnit(unit string, path string, dashboard string, title string) {
	locs, ok := u.units[unit]
	if !ok {
		locs = make([]*metricLoc, 0)
	}
	locs = append(locs, &metricLoc{
		path:      path,
		dashboard: dashboard,
		title:     title,
	})
	u.units[unit] = locs
}

func (t *metricsTable) addMetric(metric string, unit string, path string, dashboard string, title string) {
	u, ok := t.metricsByUnit[metric]
	if !ok {
		u = &units{
			units: make(map[string][]*metricLoc),
		}
		t.metricsByUnit[metric] = u
	}
	u.addUnit(unit, path, dashboard, title)
}

type metricLoc struct {
	path      string
	dashboard string
	title     string
}

func newMetricsTable() *metricsTable {
	return &metricsTable{
		metricsByUnit: make(map[string]*units),
	}
}

func checkUnits(dashboardPath string, mt *metricsTable, data []byte) {
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		doPanel("", key, value, mt, dashboardPath)
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			pathPrefix := fmt.Sprintf("panels[%d].", key.Int())
			doPanel(pathPrefix, key2, value2, mt, dashboardPath)
			return true
		})
		return true
	})
}

// detects two metrics divided by each other with labels
// e.g. 100*sum(aggr_space_used{datacenter=~"$Datacenter",cluster=~"$Cluster",node=~"$Node",aggr=~"$Aggregate"})/sum(aggr_space_total{datacenter=~"$Datacenter",cluster=~"$Cluster",node=~"$Node",aggr=~"$Aggregate"})
var metricDivideMetric1 = regexp.MustCompile(`(\w+){.*?/.*?(\w+){`)

// detects metric without labels divided by metric with labels
// e.g. lun_write_data/lun_write_ops{datacenter=~"$Datacenter",cluster=~"$Cluster",svm=~"$SVM",volume=~"$Volume",lun=~"$LUN"}/1024
var metricDivideMetric2 = regexp.MustCompile(`(\w+)/.*?(\w+){`)

// detects arrays
var metricWithArray = regexp.MustCompile(`metric=~*"(.*?)"`)

func doPanel(pathPrefix string, key gjson.Result, value gjson.Result, mt *metricsTable, dashboardPath string) bool {
	kind := value.Get("type").String()
	if kind == "row" {
		return true
	}
	path := fmt.Sprintf("%spanels[%d]", pathPrefix, key.Int())
	defaultUnit := value.Get("fieldConfig.defaults.unit").String()
	overridesSlice := value.Get("fieldConfig.overrides").Array()
	targetsSlice := value.Get("targets").Array()
	transformationsSlice := value.Get("transformations").Array()
	title := value.Get("title").String()
	sPath := shortPath(dashboardPath)

	overrides := make([]override, 0, len(overridesSlice))
	expressions := make([]expression, 0)
	valueToName := make(map[string]string) // only used with panels[*].transformations[*].options.renameByName

	for oi, overrideN := range overridesSlice {
		matcherID := overrideN.Get("matcher.id")
		matcherOptions := overrideN.Get("matcher.options")
		propertiesN := overrideN.Get("properties").Array()
		for pi, propN := range propertiesN {
			propID := propN.Get("id").String()
			propVal := propN.Get("value").String()
			if propID == "unit" {
				o := override{
					id:      matcherID.String(),
					options: matcherOptions.String(),
					unit:    propVal,
					path: fmt.Sprintf("%s.panels[%d].fieldConfig.overrides.%d.properties.%d",
						path, key.Int(), oi, pi),
				}
				overrides = append(overrides, o)
			}
		}
	}

	// Heatmap units are saved in a different place
	if kind == "heatmap" && defaultUnit == "" {
		defaultUnit = value.Get("yAxis.format").String()
	}

	for _, targetN := range targetsSlice {
		expr := targetN.Get("expr").String()
		matches := metricName.FindStringSubmatch(expr)
		if len(matches) != 2 {
			continue
		}
		// If the expression includes count( ignore since it is unit-less
		if strings.Contains(expr, "count(") {
			continue
		}
		// Filter percentages since they are unit-less
		if strings.Contains(expr, "/") {
			match := metricDivideMetric1.FindStringSubmatch(expr)
			if len(match) > 0 {
				continue
			}
			match = metricDivideMetric2.FindStringSubmatch(expr)
			if len(match) > 0 {
				continue
			}
		}

		metric := matches[1]
		// If the expression contain metric="\w+" that means this is an array
		// combine the array metric, with the metric, and include
		match := metricWithArray.FindStringSubmatch(expr)
		if len(match) > 0 {
			metric = metric + "-" + match[1]
		}

		exprRefID := targetN.Get("refId").String()
		expressions = append(expressions, expression{
			metric: metric,
			refID:  exprRefID,
		})
	}
	for _, transformN := range transformationsSlice {
		transformID := transformN.Get("id").String()
		if transformID == "organize" {
			rbn := transformN.Get("options.renameByName").Map()
			for k, v := range rbn {
				valueToName[k] = v.String()
			}
		}
	}
	numExpressions := len(expressions)
	for _, e := range expressions {
		// Ignore labels and _status
		if strings.HasSuffix(e.metric, "_labels") || strings.HasSuffix(e.metric, "_status") {
			continue
		}
		unit := unitForExpr(e, overrides, defaultUnit, valueToName, numExpressions)
		mt.addMetric(e.metric, unit, path, sPath, title)
	}
	return true
}

func unitForExpr(e expression, overrides []override, defaultUnit string,
	valueToName map[string]string, numExpressions int) string {

	if len(overrides) == 0 {
		return defaultUnit
	}
	// search the slice of overrides for a match, and use the overridden unit if found.
	// When the valueToName map is not empty, that means some objects were renamed in Grafana.
	// Consult the map to track the renames.
	// Special case: If there is a single expression, Grafana will name it "Value" instead of "Value #refId"
	byNameQuery := "Value"
	if numExpressions > 1 {
		byNameQuery = fmt.Sprintf("Value #%s", e.refID)
	}
	byNameQuery2 := ""
	for _, o := range overrides {
		if len(valueToName) > 0 {
			n, ok := valueToName[byNameQuery]
			if ok {
				byNameQuery2 = n
			}
		}
		if o.id == "byName" {
			if o.options == byNameQuery || o.options == byNameQuery2 {
				return o.unit
			}
		} else if o.id == "byFrameRefID" {
			if o.options == e.refID || o.options == byNameQuery2 {
				return o.unit
			}
		}
	}
	return defaultUnit
}

func TestVariablesAreSorted(t *testing.T) {
	visitDashboards([]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkVariablesAreSorted(t, path, data)
		})
}

func checkVariablesAreSorted(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "templating.list").ForEach(func(key, value gjson.Result) bool {
		// The datasource variable does not need to be sorted, ignore
		if value.Get("type").String() == "datasource" {
			return true
		}
		// If the variable is not visible, ignore
		if value.Get("hide").Int() != 0 {
			return true
		}
		// If the variable is custom, ignore
		if value.Get("type").String() == "custom" {
			return true
		}

		sortVal := value.Get("sort").Int()
		if sortVal != 1 {
			varName := value.Get("name").String()
			t.Errorf("dashboard=%s path=templating.list[%s].sort variable=%s is not 1. Should be \"sort\": 1,",
				shortPath(path), key.String(), varName)
		}
		return true
	})
}

func TestNoUnusedVariables(t *testing.T) {
	visitDashboards(
		[]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkUnusedVariables(t, path, data)
		})
}

func checkUnusedVariables(t *testing.T, path string, data []byte) {
	// collect are variable names, expressions except data source
	vars := make([]string, 0)
	varExpression := make([]string, 0)
	gjson.GetBytes(data, "templating.list").ForEach(func(key, value gjson.Result) bool {
		if value.Get("type").String() == "datasource" {
			return true
		}
		// name of variable
		vars = append(vars, value.Get("name").String())
		// query expression of variable
		varExpression = append(varExpression, value.Get("definition").String())
		return true
	})

	expressions := allExpressions(data)

	// check that each variable is used in at least one expression
varLoop:
	for _, variable := range vars {
		for _, expr := range expressions {
			if strings.Contains(expr.metric, variable) {
				continue varLoop
			}
		}
		for _, varExpr := range varExpression {
			if strings.Contains(varExpr, variable) {
				continue varLoop
			}
		}

		t.Errorf("dashboard=%s has unused variable [%s]", shortPath(path), variable)
	}
}

func allExpressions(data []byte) []expression {
	exprs := make([]expression, 0)
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		doExpr("", key, value, func(expr expression) {
			exprs = append(exprs, expr)
		})
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			pathPrefix := fmt.Sprintf("panels[%d].", key.Int())
			doExpr(pathPrefix, key2, value2, func(expr expression) {
				exprs = append(exprs, expr)
			})
			return true
		})
		return true
	})
	return exprs
}

func doExpr(pathPrefix string, key gjson.Result, value gjson.Result, exprFunc func(exp expression)) {
	kind := value.Get("type").String()
	if kind == "row" {
		return
	}
	path := fmt.Sprintf("%spanels[%d]", pathPrefix, key.Int())
	targetsSlice := value.Get("targets").Array()
	for i, targetN := range targetsSlice {
		expr := targetN.Get("expr").String()
		pathWithTarget := path + ".targets[" + strconv.Itoa(i) + "]"
		exprFunc(expression{
			refID:  pathWithTarget,
			metric: expr,
			kind:   kind,
		})
	}
}

func TestIDIsBlank(t *testing.T) {
	visitDashboards(
		[]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkUIDIsBlank(t, path, data)
		})
}

func checkUIDIsBlank(t *testing.T, path string, data []byte) {
	uid := gjson.GetBytes(data, "uid").String()
	if uid != "" {
		t.Errorf(`dashboard=%s uid should be "" but is %s`, shortPath(path), uid)
	}
}

func TestUniquePanelIDs(t *testing.T) {
	visitDashboards(
		[]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkUniquePanelIDs(t, path, data)
		})
}

func checkUniquePanelIDs(t *testing.T, path string, data []byte) {
	ids := make(map[int64]struct{})

	// visit all panel ids
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		id := value.Get("id").Int()
		_, ok := ids[id]
		if ok {
			t.Errorf(`dashboard=%s path=panels[%d] has multiple panels with id=%d`,
				shortPath(path), key.Int(), id)
		}
		ids[id] = struct{}{}
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			where := fmt.Sprintf("panels[%d].", key.Int())
			id := value2.Get("id").Int()
			_, ok := ids[id]
			if ok {
				t.Errorf(`dashboard=%s path=%spanels[%d] has multiple panels with id=%d`,
					shortPath(path), where, key2.Int(), id)
			}
			ids[id] = struct{}{}
			return true
		})
		return true
	})
}

// - collect all expressions that include "topk". Ignore expressions that are:
// 		- part of a table or stat or
//      - calculate a percentage
// - for each expression - check if any variable used in the expression has a topk range
//   a) if it does, pass
//   b) otherwise fail, printing the expression, path, dashboard

func TestTopKRange(t *testing.T) {
	visitDashboards(
		[]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkTopKRange(t, path, data)
		})
}

func checkTopKRange(t *testing.T, path string, data []byte) {
	// collect all expressions
	expressions := make([]exprP, 0)
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		doTarget("", key, value, func(path string, expr string, format string) {
			if format == "table" || format == "stat" {
				return
			}
			expressions = append(expressions, newExpr(path, expr))
		})
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			pathPrefix := fmt.Sprintf("panels[%d].", key.Int())
			doTarget(pathPrefix, key2, value2, func(path string, expr string, format string) {
				if format == "table" || format == "stat" {
					return
				}
				expressions = append(expressions, newExpr(path, expr))
			})
			return true
		})
		return true
	})

	// collect all variables
	variables := allVariables(data)

	for _, expr := range expressions {
		if !strings.Contains(expr.expr, "topk") {
			continue
		}
		if strings.Contains(expr.expr, "/") {
			continue
		}
		hasRange := false
	vars:
		for _, name := range expr.vars {
			for _, v := range variables {
				if v.name == name && strings.Contains(v.query, "__range") {
					hasRange = true
					break vars
				}
			}
		}
		if !hasRange {
			t.Errorf(`dashboard=%s path=%s use topk but no variable has range. expr=%s`,
				shortPath(path), expr.path, expr.expr)
		}
	}
}

func TestOnlyHighlightsExpanded(t *testing.T) {
	exceptions := map[string]int{
		"cmode/shelf.json":    2,
		"cmode/security.json": 3,
		"cmode/fsa.json":      2,
	}
	// count number of expanded sections in dashboard and ensure num expanded = 1
	visitDashboards(
		[]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkExpansion(t, exceptions, path, data)
		})
}

func checkExpansion(t *testing.T, exceptions map[string]int, path string, data []byte) {
	pathCollapsed := make(map[string]bool)
	titles := make([]string, 0)
	// visit all panel
	visitAllPanels(data, func(path string, key, value gjson.Result) {
		collapsed := value.Get("collapsed")
		if !collapsed.Exists() {
			return
		}
		title := value.Get("title")
		pathCollapsed[path] = collapsed.Bool()
		if !collapsed.Bool() {
			titles = append(titles, title.String())
		}
	})
	if len(titles) == 0 {
		return
	}
	dashPath := shortPath(path)
	// By default, a single expanded row is allowed.
	allowed := 1
	v, ok := exceptions[dashPath]
	if ok {
		allowed = v
	}
	if len(titles) != allowed {
		t.Errorf("%s expanded section(s) want=%d got=%d", dashPath, allowed, len(titles))
	}
}

func visitAllPanels(data []byte, handle func(path string, key gjson.Result, value gjson.Result)) {
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		path := fmt.Sprintf("panels[%d]", key.Int())
		handle(path, key, value)
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			path2 := fmt.Sprintf("%spanels[%d]", path, key2.Int())
			handle(path2, key2, value2)
			return true
		})
		return true
	})
}

func TestLegends(t *testing.T) {
	visitDashboards(
		[]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkLegends(t, path, data)
		})
}

func checkLegends(t *testing.T, path string, data []byte) {
	// collect all legends
	dashPath := shortPath(path)

	visitAllPanels(data, func(path string, key, value gjson.Result) {
		doLegends(t, value, dashPath)
	})
}

func doLegends(t *testing.T, value gjson.Result, dashPath string) {
	wantDisplayMode := "table"
	wantPlacement := "bottom"

	kind := value.Get("type").String()
	if kind == "row" || kind == "piechart" {
		return
	}
	optionsData := value.Get("options")
	if legendData := optionsData.Get("legend"); legendData.Exists() {
		legendDisplayMode := legendData.Get("displayMode").String()
		legendPlacementData := legendData.Get("placement").String()
		title := value.Get("title").String()
		calcsData := legendData.Get("calcs").Array()
		var calcsSlice []string
		for _, result := range calcsData {
			calcsSlice = append(calcsSlice, result.String())
		}
		checkLegendCalculations(t, calcsSlice, dashPath, title)

		// Skip hidden legends
		if legendDisplayMode == "hidden" {
			return
		}
		if legendDisplayMode != wantDisplayMode {
			t.Errorf(`dashboard=%s, panel="%s", display mode want=%s got=%s val %v`, dashPath, title, wantDisplayMode, legendDisplayMode, legendData)
		}

		if legendPlacementData != wantPlacement {
			t.Errorf(`dashboard=%s, panel="%s", legend placement want=%s got=%s val %v`, dashPath, title, wantPlacement, legendPlacementData, legendData)
		}
	}
}

func checkLegendCalculations(t *testing.T, gotLegendCalculations []string, dashPath, title string) {
	wantLegendNoMin := strings.Join([]string{"mean", "lastNotNull", "max"}, ",")
	wantLegendWithMin := "min," + wantLegendNoMin
	got := strings.Join(gotLegendCalculations, ",")
	if strings.Contains(got, "sum") {
		return
	}
	if strings.Contains(got, "min") {
		if got != wantLegendWithMin {
			t.Errorf(`dashboard=%s, panel="%s", got=[%s] want=[%s]`, dashPath, title, got, wantLegendWithMin)
		}
	} else {
		if got != wantLegendNoMin {
			t.Errorf(`dashboard=%s, panel="%s", got=[%s] want=[%s]`, dashPath, title, got, wantLegendNoMin)
		}
	}
}

func TestConnectNullValues(t *testing.T) {
	visitDashboards(
		[]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkConnectNullValues(t, path, data)
		})
}

func checkConnectNullValues(t *testing.T, path string, data []byte) {
	dashPath := shortPath(path)

	visitAllPanels(data, func(path string, key, value gjson.Result) {
		spanNulls := value.Get("fieldConfig.defaults.custom.spanNulls")
		if !spanNulls.Exists() {
			return
		}
		if !spanNulls.Bool() {
			t.Errorf(`dashboard=%s panel="%s got=[%s] want=true`, dashPath, value.Get("title").String(), spanNulls.String())
		}
	})
}

func TestPanelChildPanels(t *testing.T) {
	visitDashboards(
		[]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkPanelChildPanels(t, shortPath(path), data)
		})
}

func checkPanelChildPanels(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		// Check all collapsed panels should have child panels
		if value.Get("collapsed").Bool() && len(value.Get("panels").Array()) == 0 {
			t.Errorf("dashboard=%s, panel=%s, has child panels outside of row", path, value.Get("title").String())
		}
		return true
	})
}

func TestRatesAreNot1m(t *testing.T) {
	visitDashboards(
		[]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkRate1m(t, shortPath(path), data)
		},
	)
}

func checkRate1m(t *testing.T, path string, data []byte) {
	expressions := allExpressions(data)
	for _, expr := range expressions {
		if strings.Contains(expr.metric, "[1m]") {
			t.Errorf("dashboard=%s, expr should not use rate of [1m] expr=%s", path, expr)
		}
	}
}

func TestTitlesOfTopN(t *testing.T) {
	visitDashboards(
		[]string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"},
		func(path string, data []byte) {
			checkTitlesOfTopN(t, shortPath(path), data)
		},
	)
}

func checkTitlesOfTopN(t *testing.T, path string, data []byte) {
	expressions := allExpressions(data)
	for _, expr := range expressions {
		if !strings.Contains(expr.metric, "topk") || expr.kind == "stat" {
			continue
		}
		titleRef := asTitle(expr.refID)
		title := gjson.GetBytes(data, titleRef)

		// Check that the title contains are variable
		if !strings.Contains(title.String(), "$") {
			t.Errorf("dashboard=%s, title=%s at=%s does not include TopResource var", path, title, titleRef)
		}
	}
}

func asTitle(id string) string {
	// Replace the last segment with title and gjson-ify the path
	// This `panels[26].panels[0].targets[0]` becomes `panels.26.panels.0.title`
	splits := strings.Split(id, ".")
	if len(splits) < 2 {
		return id
	}
	splits[len(splits)-1] = "title"
	path := strings.Join(splits, ".")
	replacer := strings.NewReplacer("[", ".", "]", "")
	return replacer.Replace(path)
}
