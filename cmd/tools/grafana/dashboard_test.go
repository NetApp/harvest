package grafana

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var dashboards = []string{"../../../grafana/dashboards/cmode", "../../../grafana/dashboards/storagegrid"}

func TestThreshold(t *testing.T) {
	visitDashboards(dashboards, func(path string, data []byte) {
		checkThreshold(t, path, data)
	})
}

var aggregationPattern = regexp.MustCompile(`\b(sum|count|min|max)\b`)

func checkThreshold(t *testing.T, path string, data []byte) {
	path = shortPath(path)
	var thresholdMap = map[string][]string{
		// _latency are in microseconds
		"_latency": {
			"[\"green\",\"orange\",\"red\"]",
			"[null,20000,30000]",
		},
		"_busy": {
			"[\"green\",\"orange\",\"red\"]",
			"[null,60,80]",
		},
	}
	// visit all panel for datasource test
	visitAllPanels(data, func(p string, key, value gjson.Result) {
		panelTitle := value.Get("title").String()
		kind := value.Get("type").String()
		if kind == "table" || kind == "stat" {
			targetsSlice := value.Get("targets").Array()
			for _, targetN := range targetsSlice {
				expr := targetN.Get("expr").String()
				// Check if the metric matches the aggregation pattern
				if aggregationPattern.MatchString(expr) {
					continue
				}
				if strings.Contains(expr, "_latency") || strings.Contains(expr, "_busy") {
					var th []string
					if strings.Contains(expr, "_latency") {
						th = thresholdMap["_latency"]
					} else if strings.Contains(expr, "_busy") {
						th = thresholdMap["_busy"]
					}
					isThresholdSet := false
					isColorBackgroundSet := false
					expectedColorBackground := map[string][]string{
						"table": {"color-background", "lcd-gauge"},
						"stat":  {"background"},
					}
					// check in default also for stat. For table we only want relevant column background and override settings
					if kind == "stat" {
						dS := value.Get("fieldConfig.defaults")
						tSlice := dS.Get("thresholds")
						color := tSlice.Get("steps.#.color")
						v := tSlice.Get("steps.#.value")
						isThresholdSet = color.String() == th[0] && v.String() == th[1]
					}

					// check if any override has threshold set
					overridesSlice := value.Get("fieldConfig.overrides").Array()
					for _, overrideN := range overridesSlice {
						propertiesSlice := overrideN.Get("properties").Array()
						for _, propertiesN := range propertiesSlice {
							id := propertiesN.Get("id").String()
							if id == "thresholds" {
								color := propertiesN.Get("value.steps.#.color")
								v := propertiesN.Get("value.steps.#.value")
								isThresholdSet = color.String() == th[0] && v.String() == th[1]
							} else if id == "custom.displayMode" && kind == "table" {
								v := propertiesN.Get("value")
								if !util.Contains(expectedColorBackground[kind], v.String()) {
									t.Errorf("dashboard=%s panel=%s kind=%s expr=%s don't have correct displaymode expected %s found %s", path, panelTitle, kind, expr, expectedColorBackground[kind], v.String())
								} else {
									isColorBackgroundSet = true
								}
							}
						}
					}

					if kind == "stat" {
						colorMode := value.Get("options.colorMode")
						if !util.Contains(expectedColorBackground[kind], colorMode.String()) {
							t.Errorf("dashboard=%s panel=%s kind=%s expr=%s don't have correct colorMode expected %s found %s", path, panelTitle, kind, expr, expectedColorBackground[kind], colorMode.String())
						} else {
							isColorBackgroundSet = true
						}
					}
					if !isThresholdSet {
						t.Errorf("dashboard=%s panel=%s kind=%s expr=%s don't have correct latency threshold set. expected threshold %s %s", path, panelTitle, kind, expr, th[0], th[1])
					}
					if !isColorBackgroundSet {
						t.Errorf("dashboard=%s panel=%s kind=%s expr=%s don't have displaymode expected %s", path, panelTitle, kind, expr, expectedColorBackground[kind])
					}
				}
			}
		}
	})
}

func TestDatasource(t *testing.T) {
	visitDashboards(dashboards, func(path string, data []byte) {
		checkDashboardForDatasource(t, path, data)
	})
}

func checkDashboardForDatasource(t *testing.T, path string, data []byte) {
	path = shortPath(path)
	// visit all panels for datasource test
	visitAllPanels(data, func(p string, key, value gjson.Result) {
		dsResult := value.Get("datasource")
		panelTitle := value.Get("title").String()
		if !dsResult.Exists() {
			t.Errorf("dashboard=%s panel=%s don't have datasource", path, panelTitle)
			return
		}

		if dsResult.Type == gjson.Null {
			// if the panel is a row, it is OK if there is no datasource
			if value.Get("type").String() == "row" {
				return
			}
			t.Errorf("dashboard=%s panel=%s has a null datasource", path, panelTitle)
		}
		if dsResult.String() != "${DS_PROMETHEUS}" {
			t.Errorf("dashboard=%s panel=%s has %s datasource should be ${DS_PROMETHEUS}", path, panelTitle, dsResult.String())
		}

		// Later versions of Grafana introduced a different datasource shape which causes errors
		// when used in older versions. Detect that here
		// GOOD "datasource": "${DS_PROMETHEUS}",
		// BAD  "datasource": {
		//            "type": "prometheus",
		//            "uid": "EO6UabnVz"
		//          },
		dses := value.Get("targets.#.datasource").Array()
		for i, ds := range dses {
			if ds.String() != "${DS_PROMETHEUS}" {
				targetPath := fmt.Sprintf("%s.target[%d].datasource", p, i)
				t.Errorf(
					"dashboard=%s path=%s panel=%s has %s datasource shape that breaks older versions of Grafana",
					path,
					targetPath,
					panelTitle,
					dsResult.String(),
				)
			}
		}
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
	defaultLatencyUnit := "µs"
	pattern := `\/\d+` // Regular expression pattern to match division by a number
	reg := regexp.MustCompile(pattern)
	mt := newMetricsTable()
	expectedMt := parseUnits()
	visitDashboards(dashboards,
		func(path string, data []byte) {
			checkUnits(t, path, mt, data)
		})

	// Exceptions are meant to reduce false negatives
	allowedSuffix := map[string][]string{
		"_count":    {"none", "short", "locale"},
		"_lag_time": {"", "s", "short"},
	}

	// Normalize rates to their base unit
	rates := map[string]string{
		"KiBs": "kbytes",
	}

	metricNames := make([]string, 0, len(mt.metricsByUnit))
	for m := range mt.metricsByUnit {
		metricNames = append(metricNames, m)
	}
	sort.Strings(metricNames)

	for _, metric := range metricNames {
		u := mt.metricsByUnit[metric]

		failText := strings.Builder{}
		// Normalize units if there are rates
		for unit, listMetricLoc := range u.units {
			normal, ok := rates[unit]
			if !ok {
				continue
			}
			list, ok := u.units[normal]
			if !ok {
				continue
			}
			delete(u.units, unit)
			list = append(list, listMetricLoc...)
			u.units[normal] = list
		}
		numUnits := len(u.units)
		for unit, location := range u.units {
			if unit == "" || unit == "none" {
				// Fail this metric if it contains an empty or none unit
				t.Errorf(`%s should not have unit=none %s path=%s title="%s"`,
					metric, location[0].dashboard, location[0].path, location[0].title)
			}

			var expectedGrafanaUnit string

			if v, ok := expectedMt[metric]; ok {
				expectedGrafanaUnit = v.GrafanaJSON
				if v.GrafanaJSON != unit && !v.skipValidate {
					t.Errorf(`%s should not have unit=%s expected=%s %s path=%s title="%s"`,
						metric, unit, v.GrafanaJSON, location[0].dashboard, location[0].path, location[0].title)
				}
			} else {
				// special case latency that dashboard uses unit microseconds µs
				if strings.HasSuffix(metric, "_latency") {
					expectedGrafanaUnit = defaultLatencyUnit
					if unit != expectedGrafanaUnit {
						t.Errorf(`%s should not have unit=%s expected=%s %s path=%s title="%s"`,
							metric, unit, defaultLatencyUnit, location[0].dashboard, location[0].path, location[0].title)
					}
				}
			}

			for _, l := range location {
				match := reg.MatchString(l.expr)
				if match {
					if expectedGrafanaUnit == unit {
						t.Errorf(`%s should not have unit=%s because there is a division by a number %s path=%s title="%s"`,
							metric, unit, l.dashboard, l.path, l.title)
					}
				}
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
	expr   string
}
type units struct {
	units map[string][]*metricLoc
}

func (u *units) addUnit(unit string, path string, dashboard string, title string, expr string) {
	locs, ok := u.units[unit]
	if !ok {
		locs = make([]*metricLoc, 0)
	}
	locs = append(locs, &metricLoc{
		path:      path,
		dashboard: dashboard,
		title:     title,
		expr:      expr,
	})
	u.units[unit] = locs
}

func (t *metricsTable) addMetric(metric string, unit string, path string, dashboard string, title string, expr string) {
	u, ok := t.metricsByUnit[metric]
	if !ok {
		u = &units{
			units: make(map[string][]*metricLoc),
		}
		t.metricsByUnit[metric] = u
	}
	u.addUnit(unit, path, dashboard, title, expr)
}

type metricLoc struct {
	path      string
	dashboard string
	title     string
	expr      string
}

func newMetricsTable() *metricsTable {
	return &metricsTable{
		metricsByUnit: make(map[string]*units),
	}
}

func checkUnits(t *testing.T, dashboardPath string, mt *metricsTable, data []byte) {
	visitAllPanels(data, func(path string, key, value gjson.Result) {
		doPanel(t, "", key, value, mt, dashboardPath)
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

func doPanel(t *testing.T, pathPrefix string, key gjson.Result, value gjson.Result, mt *metricsTable, dashboardPath string) {
	kind := value.Get("type").String()
	if kind == "row" {
		return
	}
	path := fmt.Sprintf("%spanels[%d]", pathPrefix, key.Int())
	defaultUnit := value.Get("fieldConfig.defaults.unit").String()
	overridesSlice := value.Get("fieldConfig.overrides").Array()
	targetsSlice := value.Get("targets").Array()
	transformationsSlice := value.Get("transformations").Array()
	title := value.Get("title").String()
	sPath := shortPath(dashboardPath)

	propertiesMap := make(map[string]map[string]string)
	overrides := make([]override, 0, len(overridesSlice))
	expressions := make([]expression, 0)
	valueToName := make(map[string]string) // only used with panels[*].transformations[*].options.renameByName

	for oi, overrideN := range overridesSlice {
		matcherID := overrideN.Get("matcher.id")
		// make sure that mapKey is unique for each override element
		propertiesMapKey := matcherID.String() + strconv.Itoa(oi)
		propertiesMap[propertiesMapKey] = make(map[string]string)
		matcherOptions := overrideN.Get("matcher.options")
		propertiesN := overrideN.Get("properties").Array()
		for pi, propN := range propertiesN {
			propID := propN.Get("id").String()
			propVal := propN.Get("value").String()
			propertiesMap[propertiesMapKey][propID] = propVal
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

	// In case of gradient-gauge and percent(0.0-1.0), we must override min and max value
	for _, properties := range propertiesMap {
		displayMode := properties["custom.displayMode"]
		if properties["unit"] == "percentunit" && (displayMode == "gradient-gauge" || displayMode == "lcd-gauge" || displayMode == "basic") {
			if maxVal, exist := properties["max"]; !exist || maxVal != "1" {
				t.Errorf("dashboard=%s, title=%s should have max value 1", sPath, title)
			}
			if minVal, exist := properties["min"]; !exist || minVal != "0" {
				t.Errorf("dashboard=%s, title=%s should have min value 0", sPath, title)
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
			expr:   expr,
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
		if strings.HasSuffix(e.metric, "_labels") || strings.HasSuffix(e.metric, "_status") || strings.HasSuffix(e.metric, "_events") || strings.HasSuffix(e.metric, "_alerts") {
			continue
		}
		unit := unitForExpr(e, overrides, defaultUnit, valueToName, numExpressions)
		mt.addMetric(e.metric, unit, path, sPath, title, e.expr)
	}
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
	visitDashboards(dashboards,
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
		dashboards,
		func(path string, data []byte) {
			checkUnusedVariables(t, path, data)
		})
}

func checkUnusedVariables(t *testing.T, path string, data []byte) {
	// collect are variable names, expressions except data source
	vars := make([]string, 0)
	description := make([]string, 0)
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

	visitAllPanels(data, func(path string, key, value gjson.Result) {
		d := value.Get("description").String()
		if d != "" {
			description = append(description, d)
		}
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

		for _, desc := range description {
			if strings.Contains(desc, "$"+variable) {
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
		dashboards,
		func(path string, data []byte) {
			checkUIDIsBlank(t, path, data)
			checkIDIsNull(t, path, data)
		})
}

func TestExemplarIsFalse(t *testing.T) {
	visitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkExemplarIsFalse(t, path, data)
		})
}

func checkExemplarIsFalse(t *testing.T, path string, data []byte) {
	if strings.Contains(string(data), "\"exemplar\": true") {
		t.Errorf(`dashboard=%s exemplar should be "false" but is "true"`, shortPath(path))
	}
}

func checkUIDIsBlank(t *testing.T, path string, data []byte) {
	uid := gjson.GetBytes(data, "uid").String()
	if uid != "" {
		t.Errorf(`dashboard=%s uid should be "" but is %s`, shortPath(path), uid)
	}
}

func checkIDIsNull(t *testing.T, path string, data []byte) {
	id := gjson.GetBytes(data, "id").String()
	if id != "" {
		t.Errorf(`dashboard=%s id should be null but is %s`, shortPath(path), id)
	}
}

func TestUniquePanelIDs(t *testing.T) {
	visitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkUniquePanelIDs(t, path, data)
		})
}

func checkUniquePanelIDs(t *testing.T, path string, data []byte) {
	ids := make(map[int64]struct{})

	sPath := shortPath(path)
	// visit all panel ids
	visitAllPanels(data, func(path string, key, value gjson.Result) {
		id := value.Get("id").Int()
		_, ok := ids[id]
		if ok {
			t.Errorf(`dashboard=%s path=panels[%d] has multiple panels with id=%d`,
				sPath, key.Int(), id)
		}
		ids[id] = struct{}{}
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
		dashboards,
		func(path string, data []byte) {
			checkTopKRange(t, path, data)
		})
}

func checkTopKRange(t *testing.T, path string, data []byte) {
	// collect all expressions
	expressions := make([]exprP, 0)

	visitAllPanels(data, func(path string, key, value gjson.Result) {
		doTarget("", key, value, func(path string, expr string, format string) {
			if format == "table" || format == "stat" {
				return
			}
			expressions = append(expressions, newExpr(path, expr))
		})
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

	for _, v := range variables {
		if v.name == "TopResources" {
			for _, optionVal := range v.options {
				selected := optionVal.Get("selected").Bool()
				text := optionVal.Get("text").String()
				value := optionVal.Get("value").String()

				// Test if text and value match, except for the special case with "All" and "$__all"
				if text != value && !(text == "All" && value == "$__all") {
					t.Errorf("In dashboard %s, variable %s uses topk, but text '%s' does not match value '%s'",
						shortPath(path), v.name, text, value)
				}

				// Test if the selected value matches the expected text "5"
				if (text == "5") != selected {
					t.Errorf("In dashboard %s, variable %s uses topk, but text '%s' has incorrect selected state: %t",
						shortPath(path), v.name, text, selected)
				}
			}
		}
		if !strings.Contains(v.query, "topk") || !strings.Contains(v.query, "__range") {
			continue
		}

		if v.refresh != "2" {
			t.Errorf("dashboard=%s name=%s use topk, refresh should be set to \"On time range change\". query=%s",
				shortPath(path), v.name, v.query)
		}
	}

}

func TestOnlyHighlightsExpanded(t *testing.T) {
	exceptions := map[string]int{
		"cmode/shelf.json":            2,
		"cmode/fsa.json":              2,
		"cmode/workload.json":         2,
		"cmode/smb.json":              2,
		"cmode/health.json":           2,
		"cmode/power.json":            2,
		"storagegrid/fabricpool.json": 2,
	}
	// count number of expanded sections in dashboard and ensure num expanded = 1
	visitDashboards(
		dashboards,
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

func TestLegends(t *testing.T) {
	visitDashboards(
		dashboards,
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
		dashboards,
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
		dashboards,
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
		dashboards,
		func(path string, data []byte) {
			checkRate1m(t, shortPath(path), data)
		},
	)
}

func checkRate1m(t *testing.T, path string, data []byte) {
	expressions := allExpressions(data)
	for _, expr := range expressions {
		if strings.Contains(expr.metric, "[1m]") {
			t.Errorf("dashboard=%s, expr should not use rate of [1m] expr=%s", path, expr.metric)
		}
	}
}

func TestTableFilter(t *testing.T) {
	visitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkTableFilter(t, path, data)
		})
}

func checkTableFilter(t *testing.T, path string, data []byte) {
	dashPath := shortPath(path)
	visitAllPanels(data, func(path string, key, value gjson.Result) {
		panelType := value.Get("type").String()
		if panelType == "table" {
			isFilterable := value.Get("fieldConfig.defaults.custom.filterable").String()
			if isFilterable != "true" {
				t.Errorf(`dashboard=%s path=panels[%d] does not enable filtering for the table`, dashPath, key.Int())
			}
		}
	})
}

func TestTitlesOfTopN(t *testing.T) {
	visitDashboards(
		dashboards,
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

func TestPercentHasMinMax(t *testing.T) {
	visitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkPercentHasMinMax(t, path, data)
		})
}

func checkPercentHasMinMax(t *testing.T, path string, data []byte) {
	dashPath := shortPath(path)

	visitAllPanels(data, func(path string, key, value gjson.Result) {
		panelType := value.Get("type").String()
		if panelType != "timeseries" {
			return
		}
		defaultUnit := value.Get("fieldConfig.defaults.unit").String()
		if defaultUnit != "percent" && defaultUnit != "percentunit" {
			return
		}
		min := value.Get("fieldConfig.defaults.min").String()
		max := value.Get("fieldConfig.defaults.max").String()
		if min != "0" {
			t.Errorf(`dashboard=%s path=%s panel="%s" has unit=%s, min should be 0 got=%s`,
				dashPath, path, value.Get("title").String(), defaultUnit, min)
		}
		if defaultUnit == "percent" && max != "100" {
			t.Errorf(`dashboard=%s path=%s panel="%s" has unit=%s, max should be 100 got=%s`,
				dashPath, path, value.Get("title").String(), defaultUnit, max)
		}
		if defaultUnit == "percentunit" && max != "1" {
			t.Errorf(`dashboard=%s path=%s panel="%s" has unit=%s, max should be 1 got=%s`,
				dashPath, path, value.Get("title").String(), defaultUnit, max)
		}
	})
}

func TestRefreshIsOff(t *testing.T) {
	visitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkDashboardRefresh(t, shortPath(path), data)
		},
	)
}

func checkDashboardRefresh(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "refresh").ForEach(func(key, value gjson.Result) bool {
		if value.String() != "" {
			t.Errorf(`dashboard=%s, got refresh=%s, want refresh="" (off)`, path, value.String())
		}
		return true
	})
}

func TestHeatmapSettings(t *testing.T) {
	visitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkHeatmapSettings(t, shortPath(path), data)
		},
	)
}

func checkHeatmapSettings(t *testing.T, path string, data []byte) {
	dashPath := shortPath(path)
	const (
		wantColorScheme = "interpolateRdYlGn"
		wantColorMode   = "spectrum"
	)
	visitAllPanels(data, func(path string, key, value gjson.Result) {
		panelType := value.Get("type").String()
		if panelType != "heatmap" {
			return
		}
		colorScheme := value.Get("color.colorScheme").String()
		colorMode := value.Get("color.mode").String()
		if colorScheme != wantColorScheme {
			t.Errorf(`dashboard=%s path=%s panel="%s" got color.scheme=%s, want=%s`,
				dashPath, path, value.Get("title").String(), colorScheme, wantColorScheme)
		}
		if colorMode != wantColorMode {
			t.Errorf(`dashboard=%s path=%s panel="%s" got color.mode=%s, want=%s`,
				dashPath, path, value.Get("title").String(), colorMode, wantColorMode)
		}
	})
}

func TestBytePanelsHave2Decimals(t *testing.T) {
	visitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkBytePanelsHave2Decimals(t, path, data)
		})
}

func checkBytePanelsHave2Decimals(t *testing.T, path string, data []byte) {
	dashPath := shortPath(path)
	byteTypes := map[string]bool{
		"bytes":     true,
		"decbytes":  true,
		"bits":      true,
		"decbits":   true,
		"kbytes":    true,
		"deckbytes": true,
		"mbytes":    true,
		"decmbytes": true,
		"gbytes":    true,
		"decgbytes": true,
		"tbytes":    true,
		"dectbytes": true,
		"pbytes":    true,
		"decpbytes": true,
	}

	visitAllPanels(data, func(path string, key, value gjson.Result) {
		panelType := value.Get("type").String()
		if panelType != "timeseries" {
			return
		}
		defaultUnit := value.Get("fieldConfig.defaults.unit").String()
		if !byteTypes[defaultUnit] {
			return
		}
		decimals := value.Get("fieldConfig.defaults.decimals").String()
		if decimals != "2" {
			t.Errorf(`dashboard=%s path=%s panel="%s" got decimals=%s, want decimals=2`,
				dashPath, path, value.Get("title").String(), decimals)
		}
	})
}

func TestDashboardKeysAreSorted(t *testing.T) {
	visitDashboards(
		dashboards,
		func(path string, data []byte) {
			path = shortPath(path)
			sorted := pretty.PrettyOptions(data, &pretty.Options{
				SortKeys: true,
				Indent:   "  ",
			})
			if string(sorted) != string(data) {
				sortedPath := writeSorted(t, path, sorted)
				path = "grafana/dashboards/" + path
				t.Errorf("dashboard=%s should have sorted keys but does not. Sorted version created at path=%s. Run cp %s %s",
					path, sortedPath, sortedPath, path)
			}
		})
}

func writeSorted(t *testing.T, path string, sorted []byte) string {
	dir, file := filepath.Split(path)
	dir = filepath.Dir(dir)
	dest := filepath.Join("/tmp", dir, file)
	destDir := filepath.Dir(dest)
	err := os.MkdirAll(destDir, 0750)
	if err != nil {
		t.Errorf("failed to create dir=%s err=%v", destDir, err)
		return ""
	}
	create, err := os.Create(dest)

	if err != nil {
		t.Errorf("failed to create file=%s err=%v", dest, err)
		return ""
	}
	_, err = create.Write(sorted)
	if err != nil {
		t.Errorf("failed to write sorted json to file=%s err=%v", dest, err)
		return ""
	}
	err = create.Close()
	if err != nil {
		t.Errorf("failed to close file=%s err=%v", dest, err)
		return ""
	}
	return dest
}

func TestDashboardTime(t *testing.T) {
	visitDashboards(dashboards, func(path string, data []byte) {
		checkDashboardTime(t, path, data)
	})
}

func checkDashboardTime(t *testing.T, path string, data []byte) {
	dashPath := shortPath(path)
	from := gjson.GetBytes(data, "time.from")
	to := gjson.GetBytes(data, "time.to")

	fromWant := "now-3h"
	toWant := "now"
	if from.String() != fromWant {
		t.Errorf("dashboard=%s time.from got=%s want=%s", dashPath, from.String(), fromWant)
	}
	if to.String() != toWant {
		t.Errorf("dashboard=%s time.to got=%s want=%s", dashPath, to.String(), toWant)
	}
}
