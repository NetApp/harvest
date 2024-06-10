package grafana

import (
	"fmt"
	"github.com/tidwall/gjson"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var dashboards = []string{
	"../../../grafana/dashboards/cmode",
	"../../../grafana/dashboards/cmode-details",
	"../../../grafana/dashboards/storagegrid",
}

func TestThreshold(t *testing.T) {
	VisitDashboards(dashboards, func(path string, data []byte) {
		checkThreshold(t, path, data)
	})
}

var aggregationPattern = regexp.MustCompile(`\b(sum|count|min|max)\b`)

func checkThreshold(t *testing.T, path string, data []byte) {
	path = ShortPath(path)
	var thresholdMap = map[string][]string{
		// _latencies are in microseconds
		"_latency": {
			"[\"green\",\"orange\",\"red\"]",
			"[null,20000,30000]",
		},
		"_busy": {
			"[\"green\",\"orange\",\"red\"]",
			"[null,60,80]",
		},
	}
	// visit all panels for datasource test
	VisitAllPanels(data, func(_ string, _, value gjson.Result) {
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
						"stat":  {"background", "value"},
					}
					// check in default also for stat. For table, we only want the relevant column background and override settings
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
								if !slices.Contains(expectedColorBackground[kind], v.String()) {
									t.Errorf("dashboard=%s panel=%s kind=%s expr=%s don't have correct displaymode expected %s found %s", path, panelTitle, kind, expr, expectedColorBackground[kind], v.String())
								} else {
									isColorBackgroundSet = true
								}
							}
						}
					}

					if kind == "stat" {
						colorMode := value.Get("options.colorMode")
						if !slices.Contains(expectedColorBackground[kind], colorMode.String()) {
							t.Errorf("dashboard=%s panel=%s kind=%s expr=%s doesn't have correct colorMode got %s want %s", path, panelTitle, kind, expr, colorMode.String(), expectedColorBackground[kind])
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
	VisitDashboards(dashboards, func(path string, data []byte) {
		checkDashboardForDatasource(t, path, data)
	})
}

func checkDashboardForDatasource(t *testing.T, path string, data []byte) {
	path = ShortPath(path)
	// visit all panels for datasource test
	VisitAllPanels(data, func(p string, _, value gjson.Result) {
		dsResult := value.Get("datasource")
		panelTitle := value.Get("title").String()
		if !dsResult.Exists() {
			t.Errorf(`dashboard="%s" panel="%s" doesn't have a datasource`, path, panelTitle)
			return
		}

		if dsResult.Type == gjson.Null {
			// if the panel is a row, it is OK if there is no datasource
			if value.Get("type").String() == "row" {
				return
			}
			t.Errorf(`dashboard=%s panel="%s" has a null datasource, should be ${DS_PROMETHEUS}`, path, panelTitle)
		} else if dsResult.String() != "${DS_PROMETHEUS}" {
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
	// This is a list of names that are exempt from the check for a 'true' selected status.
	excludedNames := map[string]bool{
		"TopResources": true,
		"Interval":     true,
		"IncludeRoot":  true,
	}

	excludeTypes := map[string]bool{
		"textbox":    true,
		"custom":     true,
		"datasource": true,
	}

	gjson.GetBytes(data, "templating.list").ForEach(func(_, value gjson.Result) bool {
		name := value.Get("name").String()
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

		if !excludedNames[name] {
			if value.Get("current.selected").String() == "true" {
				t.Errorf(
					"dashboard=%s var=current.selected query want=false got=%s text=%s value=%s name= %s",
					path,
					"true",
					value.Get("current.text"),
					value.Get("current.value"),
					name,
				)
			}
			ttype := value.Get("type").String()
			datasource := value.Get("datasource").String()
			if !excludeTypes[ttype] && datasource != "${DS_PROMETHEUS}" {
				t.Errorf("dashboard=%s var=%s has %s datasource should be ${DS_PROMETHEUS}", path, name, datasource)
			}
		}
		return true
	})

	if !doesDsPromExist {
		t.Errorf("dashboard=%s should define the variable DS_PROMETHEUS", path)
	}
}

func TestUnitsAndExprMatch(t *testing.T) {
	defaultLatencyUnit := "µs"
	pattern := `\/\d+` // Regular expression pattern to match division by a number
	reg := regexp.MustCompile(pattern)
	mt := newMetricsTable()
	expectedMt := parseUnits()
	VisitDashboards(dashboards,
		func(path string, data []byte) {
			checkUnits(t, path, mt, data)
		})

	// Exceptions are meant to reduce false negatives
	allowedSuffix := map[string][]string{
		"_count":                          {"none", "short", "locale"},
		"_lag_time":                       {"", "s", "short"},
		"qos_detail_service_time_latency": {"µs", "percent"},
		"qos_detail_resource_latency":     {"µs", "percent"},
		"volume_space_physical_used":      {"bytes", "binBps"}, // Growth rate uses bytes/sec unit
		"volume_space_logical_used":       {"bytes", "binBps"}, // Growth rate uses bytes/sec unit
		"qos_ops":                         {"iops", "percent"},
		"qos_total_data":                  {"Bps", "percent"},
		"aggr_space_used":                 {"bytes", "percent"},
		"volume_size_used":                {"bytes", "percent"},
		"volume_num_compress_fail":        {"short", "percent"},
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
			} else if strings.HasSuffix(metric, "_latency") {
				// special case latency that dashboard uses unit microseconds µs
				expectedGrafanaUnit = defaultLatencyUnit
				if unit != expectedGrafanaUnit {
					// Check if this metric is in the allowedSuffix map and has a matching unit
					if slices.Contains(allowedSuffix[metric], unit) {
						continue
					}
					t.Errorf(`%s should not have unit=%s expected=%s %s path=%s title="%s"`,
						metric, unit, defaultLatencyUnit, location[0].dashboard, location[0].path, location[0].title)
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
				failText.WriteString(fmt.Sprintf("unit=%s %s path=%s title=%q\n",
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
	VisitAllPanels(data, func(_ string, key, value gjson.Result) {
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
	sPath := ShortPath(dashboardPath)

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
		if (properties["unit"] == "percentunit" || defaultUnit == "percentunit") && (displayMode == "gradient-gauge" || displayMode == "lcd-gauge" || displayMode == "basic") {
			if maxVal, exist := properties["max"]; !exist || maxVal != "1" {
				t.Errorf("dashboard=%s, title=%s should have max value 1", sPath, title)
			}
			if minVal, exist := properties["min"]; !exist || minVal != "0" {
				t.Errorf("dashboard=%s, title=%s should have min value 0", sPath, title)
			}
		} else if (properties["unit"] == "percent" || defaultUnit == "percent") && (displayMode == "gradient-gauge" || displayMode == "lcd-gauge" || displayMode == "basic") {
			if maxVal, exist := properties["max"]; !exist || maxVal != "100" {
				t.Errorf("dashboard=%s, title=%s should have max value 100", sPath, title)
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
		byNameQuery = "Value #" + e.refID
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
	VisitDashboards(dashboards,
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
				ShortPath(path), key.String(), varName)
		}
		return true
	})
}

func TestVariablesIncludeAllOption(t *testing.T) {
	VisitDashboards(dashboards,
		func(path string, data []byte) {
			checkVariablesHaveAll(t, path, data)
		})
}

var exceptionToAll = map[string]bool{
	"cmode-details/volumeDeepDive.json": true,
}

func checkVariablesHaveAll(t *testing.T, path string, data []byte) {
	shouldHaveAll := map[string]bool{
		"Cluster":     true,
		"Node":        true,
		"Volume":      true,
		"SVM":         true,
		"Aggregate":   true,
		"FlexGroup":   true,
		"Constituent": true,
	}
	exceptionForAllValues := map[string]bool{
		"cmode/security.json": true,
	}

	if exceptionToAll[ShortPath(path)] {
		return
	}

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

		varName := value.Get("name").String()
		if !shouldHaveAll[varName] {
			return true
		}

		includeAll := value.Get("includeAll").Bool()
		if !includeAll {
			t.Errorf("variable=%s should have includeAll=true dashboard=%s path=templating.list[%s].includeAll",
				varName, ShortPath(path), key.String())
		}

		allValues := value.Get("allValue").String()
		if allValues != ".*" {
			if exceptionForAllValues[ShortPath(path)] {
				return true
			}
			t.Errorf("dashboard=%s variable=%s is not .*. Should be \"allValues\": .*,",
				ShortPath(path), varName)
		}

		return true
	})
}

func TestNoUnusedVariables(t *testing.T) {
	VisitDashboards(
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
	gjson.GetBytes(data, "templating.list").ForEach(func(_, value gjson.Result) bool {
		if value.Get("type").String() == "datasource" {
			return true
		}
		// name of variable
		vars = append(vars, value.Get("name").String())
		// query expression of variable
		varExpression = append(varExpression, value.Get("definition").String())
		return true
	})

	VisitAllPanels(data, func(_ string, _, value gjson.Result) {
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

		t.Errorf("dashboard=%s has unused variable [%s]", ShortPath(path), variable)
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
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkUIDNotEmpty(t, path, data)
			checkIDIsNull(t, path, data)
		})
}

func TestExemplarIsFalse(t *testing.T) {
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkExemplarIsFalse(t, path, data)
		})
}

func checkExemplarIsFalse(t *testing.T, path string, data []byte) {
	if strings.Contains(string(data), "\"exemplar\": true") {
		t.Errorf(`dashboard=%s exemplar should be "false" but is "true"`, ShortPath(path))
	}
}

func checkUIDNotEmpty(t *testing.T, path string, data []byte) {
	path = ShortPath(path)
	uid := gjson.GetBytes(data, "uid").String()
	if uid == "" {
		t.Errorf(`dashboard=%s uid is "", but should not be empty`, path)
	}
}

func checkIDIsNull(t *testing.T, path string, data []byte) {
	id := gjson.GetBytes(data, "id").String()
	if id != "" {
		t.Errorf(`dashboard=%s id should be null but is %s`, ShortPath(path), id)
	}
}

func TestUniquePanelIDs(t *testing.T) {
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkUniquePanelIDs(t, path, data)
		})
}

func checkUniquePanelIDs(t *testing.T, path string, data []byte) {
	ids := make(map[int64]struct{})

	sPath := ShortPath(path)
	// visit all panel ids
	VisitAllPanels(data, func(_ string, key, value gjson.Result) {
		id := value.Get("id").Int()
		_, ok := ids[id]
		if ok {
			t.Errorf(`dashboard=%s path=panels[%d] has multiple panels with id=%d`,
				sPath, key.Int(), id)
		}
		ids[id] = struct{}{}
	})
}

// - Collect all expressions and variables that include "topk".
// Ignore expressions that are:
// 		- part of a table or stat or
//      - calculate a percentage
// - if the var|expression includes a `rate|deriv`, ensure the look-back is 4m
// - otherwise, the look-back should be 3h

func TestTopKRange(t *testing.T) {
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkTopKRange(t, path, data)
		})
}

func checkTopKRange(t *testing.T, path string, data []byte) {

	// collect all expressions
	expressions := make([]exprP, 0)

	VisitAllPanels(data, func(_ string, key, value gjson.Result) {
		doTarget("", key, value, func(path string, expr string, format string, title string) {
			if format == "table" || format == "stat" {
				return
			}
			expressions = append(expressions, newExpr(path, expr, title))
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

		for _, name := range expr.vars {
			v, ok := variables[name]
			if !ok {
				t.Errorf(`dashboard=%s path=%s is using var that does not exist. var=%s`,
					ShortPath(path), expr.path, name)
				continue
			}
			if !strings.Contains(v.query, "topk") {
				continue
			}

			problem := ensureLookBack(v.query)
			if problem != "" {
				t.Errorf(`dashboard=%s var=%s topk got=%s %s`, ShortPath(path), v.name, v.query, problem)
			}
		}

		problem := ensureLookBack(expr.expr)
		if problem != "" {
			t.Errorf(`dashboard=%s path=%s topk got=%s %s`, ShortPath(path), expr.path, expr.expr, problem)
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
						ShortPath(path), v.name, text, value)
				}

				// Test if the selected value matches the expected text "5"
				// Unless the dashboard is volumeDeepDive, in which case the text "1" should be selected
				if ShortPath(path) == "cmode/details/volumeDeepDive.json" {
					if text == "1" != selected {
						t.Errorf("In dashboard %s, variable %s uses topk, but text '%s' has incorrect selected state: %t",
							ShortPath(path), v.name, text, selected)
					}
				} else {
					if text == "5" != selected {
						t.Errorf("In dashboard %s, variable %s uses topk, but text '%s' has incorrect selected state: %t",
							ShortPath(path), v.name, text, selected)
					}
				}
			}
		}
		if !strings.Contains(v.query, "topk") || !strings.Contains(v.query, "__range") {
			continue
		}

		if v.refresh != "2" {
			t.Errorf("dashboard=%s name=%s use topk, refresh should be set to \"On time range change\". query=%s",
				ShortPath(path), v.name, v.query)
		}
	}

}

var lookBackRe = regexp.MustCompile(`\[(.*?)]`)

// ensureLookBack ensures that the look-back for a topk query is either 4m or 3h.
// If the query contains a rate or deriv function, the look-back should be 4m
// otherwise, the look-back should be 3h.
// If the look-back is incorrect, the function returns a string describing the correct look-back
func ensureLookBack(text string) string {
	if !strings.Contains(text, "[") {
		return ""
	}
	// search for the first look-back
	matches := lookBackRe.FindAllStringSubmatch(text, -1)
	indexes := lookBackRe.FindAllStringIndex(text, -1)

	for i, match := range matches {
		indexOfLookBack := indexes[i][1]

		// search backwards for the function
		openIndex := strings.LastIndex(text[:indexOfLookBack], "(")
		space := strings.LastIndex(text[:openIndex], " ")
		if space == -1 {
			space = 0
		}
		function := text[space:openIndex]

		if strings.Contains(function, "rate") || strings.Contains(function, "deriv") {
			if match[1] != "4m" {
				return "rate/deriv want=[4m]"
			}
		} else if match[1] != "3h" {
			return "range lookback want=[3h]"
		}

	}

	return ""
}

func TestOnlyHighlightsExpanded(t *testing.T) {
	exceptions := map[string]int{
		"cmode/shelf.json":              2,
		"cmode/fsa.json":                2,
		"cmode/flexcache.json":          2,
		"cmode/workload.json":           2,
		"cmode/smb.json":                2,
		"cmode/health.json":             2,
		"cmode/power.json":              2,
		"storagegrid/fabricpool.json":   2,
		"cmode/nfsTroubleshooting.json": 3,
	}
	// count the number of expanded sections in the dashboard and ensure num expanded = 1
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkExpansion(t, exceptions, path, data)
		})
}

func checkExpansion(t *testing.T, exceptions map[string]int, path string, data []byte) {
	pathCollapsed := make(map[string]bool)
	titles := make([]string, 0)
	// visit all panel
	VisitAllPanels(data, func(path string, _, value gjson.Result) {
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
	dashPath := ShortPath(path)
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

func TestLegends(t *testing.T) {
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkLegends(t, path, data)
		})
}

func checkLegends(t *testing.T, path string, data []byte) {
	// collect all legends
	dashPath := ShortPath(path)

	VisitAllPanels(data, func(_ string, _, value gjson.Result) {
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
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkConnectNullValues(t, path, data)
		})
}

func checkConnectNullValues(t *testing.T, path string, data []byte) {
	dashPath := ShortPath(path)

	VisitAllPanels(data, func(_ string, _, value gjson.Result) {
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
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkPanelChildPanels(t, ShortPath(path), data)
		})
}

func checkPanelChildPanels(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "panels").ForEach(func(_, value gjson.Result) bool {
		// Check all collapsed panels should have child panels
		if value.Get("collapsed").Bool() && len(value.Get("panels").Array()) == 0 {
			t.Errorf("dashboard=%s, panel=%s, has child panels outside of row", path, value.Get("title").String())
		}
		return true
	})
}

func TestRatesAreNot1m(t *testing.T) {
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkRate1m(t, ShortPath(path), data)
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
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkTableFilter(t, path, data)
		})
}

func checkTableFilter(t *testing.T, path string, data []byte) {
	dashPath := ShortPath(path)
	VisitAllPanels(data, func(_ string, key, value gjson.Result) {
		panelType := value.Get("type").String()
		if panelType == "table" {
			isFilterable := value.Get("fieldConfig.defaults.custom.filterable").String()
			if isFilterable != "true" {
				t.Errorf(`dashboard=%s path=panels[%d] title="%s" does not enable filtering for the table`,
					dashPath, key.Int(), value.Get("title").String())
			}
		}
	})
}

func TestJoinExpressions(t *testing.T) {
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkJoinExpressions(t, path, data)
		})
}

func checkJoinExpressions(t *testing.T, path string, data []byte) {
	dashPath := ShortPath(path)
	expectedRegex := "(.*) 1$"
	VisitAllPanels(data, func(_ string, key, value gjson.Result) {
		panelType := value.Get("type").String()
		if panelType == "table" {
			targetsSlice := value.Get("targets").Array()
			if len(targetsSlice) > 1 {
				errorFound := false
				for _, targetN := range targetsSlice {
					expr := targetN.Get("expr").String()
					if strings.Contains(expr, "label_join") {
						transformationsSlice := value.Get("transformations").Array()
						regexUsed := false
						for _, transformationN := range transformationsSlice {
							if transformationN.Get("id").String() == "renameByRegex" {
								regex := transformationN.Get("options.regex").String()
								if regex == expectedRegex {
									regexUsed = true
									break
								}
							}
						}
						if !regexUsed {
							errorFound = true
							break
						}
					}
				}
				if errorFound {
					t.Errorf(`dashboard=%s path=panels[%d] title="%s" uses label_join but does not use the expected regex`,
						dashPath, key.Int(), value.Get("title").String())
				}
			}
		}
	})
}

func TestTitlesOfTopN(t *testing.T) {
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkTitlesOfTopN(t, ShortPath(path), data)
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

func TestIOPS(t *testing.T) {
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkIOPSDecimal(t, path, data)
		})
}

func checkIOPSDecimal(t *testing.T, path string, data []byte) {
	dashPath := ShortPath(path)

	VisitAllPanels(data, func(path string, _, value gjson.Result) {
		panelType := value.Get("type").String()
		if panelType != "timeseries" {
			return
		}
		defaultUnit := value.Get("fieldConfig.defaults.unit").String()
		if defaultUnit != "iops" {
			return
		}
		decimals := value.Get("fieldConfig.defaults.decimals").String()

		if decimals != "0" {
			t.Errorf(`dashboard=%s path=%s panel="%s", decimals should be 0 got=%s`,
				dashPath, path, value.Get("title").String(), decimals)
		}
	})
}

func TestPercentHasMinMax(t *testing.T) {
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkPercentHasMinMax(t, path, data)
		})
}

func checkPercentHasMinMax(t *testing.T, path string, data []byte) {
	// These panels can show percent value more than 100.
	exceptionMap := map[string]bool{
		"CPU Busy Domains": true,
		"Top $TopResources Volumes Per Snapshot Reserve Used": true,
		"% CPU Used": true,
	}
	dashPath := ShortPath(path)

	VisitAllPanels(data, func(path string, _, value gjson.Result) {
		panelType := value.Get("type").String()
		if panelType != "timeseries" {
			return
		}
		defaultUnit := value.Get("fieldConfig.defaults.unit").String()
		if defaultUnit != "percent" && defaultUnit != "percentunit" {
			return
		}
		theMin := value.Get("fieldConfig.defaults.min").String()
		theMax := value.Get("fieldConfig.defaults.max").String()
		decimals := value.Get("fieldConfig.defaults.decimals").String()
		if theMin != "0" {
			t.Errorf(`dashboard=%s path=%s panel="%s" has unit=%s, min should be 0 got=%s`,
				dashPath, path, value.Get("title").String(), defaultUnit, theMin)
		}
		if decimals != "2" {
			t.Errorf(`dashboard=%s path=%s panel="%s", decimals should be 2 got=%s`,
				dashPath, path, value.Get("title").String(), decimals)
		}
		if defaultUnit == "percent" && !exceptionMap[value.Get("title").String()] && theMax != "100" {
			t.Errorf(`dashboard=%s path=%s panel="%s" has unit=%s, max should be 100 got=%s`,
				dashPath, path, value.Get("title").String(), defaultUnit, theMax)
		}
		if defaultUnit == "percentunit" && theMax != "1" {
			t.Errorf(`dashboard=%s path=%s panel="%s" has unit=%s, max should be 1 got=%s`,
				dashPath, path, value.Get("title").String(), defaultUnit, theMax)
		}
	})
}

func TestRefreshIsOff(t *testing.T) {
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkDashboardRefresh(t, ShortPath(path), data)
		},
	)
}

func checkDashboardRefresh(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "refresh").ForEach(func(_, value gjson.Result) bool {
		if value.String() != "" {
			t.Errorf(`dashboard=%s, got refresh=%s, want refresh="" (off)`, path, value.String())
		}
		return true
	})
}

func TestHeatmapSettings(t *testing.T) {
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkHeatmapSettings(t, ShortPath(path), data)
		},
	)
}

func checkHeatmapSettings(t *testing.T, path string, data []byte) {
	dashPath := ShortPath(path)
	const (
		wantColorScheme = "interpolateRdYlGn"
		wantColorMode   = "spectrum"
	)
	VisitAllPanels(data, func(path string, _, value gjson.Result) {
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
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			checkBytePanelsHave2Decimals(t, path, data)
		})
}

func checkBytePanelsHave2Decimals(t *testing.T, path string, data []byte) {
	dashPath := ShortPath(path)
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

	VisitAllPanels(data, func(path string, _, value gjson.Result) {
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
	VisitDashboards(
		dashboards,
		func(path string, data []byte) {
			path = ShortPath(path)
			sorted := gjson.GetBytes(data, `@pretty:{"sortKeys":true, "indent":"  ", "width":0}`).String()
			if sorted != string(data) {
				sortedPath := writeSorted(t, path, sorted)
				path = "grafana/dashboards/" + path
				t.Errorf("dashboard=%s should have sorted keys but does not. Sorted version created at path=%s.\ncp %s %s",
					path, sortedPath, sortedPath, path)
			}
		})
}

func writeSorted(t *testing.T, path string, sorted string) string {
	dir, file := filepath.Split(path)
	dir = filepath.Dir(dir)
	tempDir := "/tmp"
	dest := filepath.Join(tempDir, dir, file)
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
	_, err = create.WriteString(sorted)
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
	VisitDashboards(dashboards, func(path string, data []byte) {
		checkDashboardTime(t, path, data)
	})
}

func checkDashboardTime(t *testing.T, path string, data []byte) {
	dashPath := ShortPath(path)
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

func TestNoDrillDownRows(t *testing.T) {
	VisitDashboards(dashboards, func(path string, data []byte) {
		checkRowNames(t, path, data)
	})
}

func checkRowNames(t *testing.T, path string, data []byte) {
	path = ShortPath(path)
	VisitAllPanels(data, func(_ string, key, value gjson.Result) {
		kind := value.Get("type").String()
		if kind == "row" {
			title := value.Get("title").String()
			if strings.Contains(title, "Drilldown") {
				t.Errorf(`dashboard=%s path=panels[%d] title=[%s] got row with Drilldown in title. Remove drilldown`, path, key.Int(), title)
			}
		}
	})
}

func TestDescription(t *testing.T) {
	count := 0
	VisitDashboards(
		[]string{"../../../grafana/dashboards/cmode"},
		func(path string, data []byte) {
			checkDescription(t, path, data, &count)
		})
}

func checkDescription(t *testing.T, path string, data []byte, count *int) {
	dashPath := ShortPath(path)
	ignoreDashboards := []string{
		"cmode/health.json", "cmode/headroom.json",
	}
	if slices.Contains(ignoreDashboards, dashPath) {
		fmt.Printf(`dashboard=%s skipped\n`, dashPath)
		return
	}

	// we don't get description for below panels, we need to manually formed them as per our need.
	ignoreList := []string{
		// These are from fsa
		"Volume Access ($Activity) History", "Volume Access ($Activity) History By Percent", "Volume Modify ($Activity) History", "Volume Modify ($Activity) History By Percent",
		// This is from workload
		"Top $TopResources Workloads by Service Time from sync_repl", "Top $TopResources Workloads by Service Time from flexcache_ral", "Top $TopResources Workloads by Service Time from flexcache_spinhi",
		"Top $TopResources Workloads by Latency from sync_repl", "Top $TopResources Workloads by Latency from flexcache_ral", "Top $TopResources Workloads by Latency from flexcache_spinhi", "Service Latency by Resources",
		// These are from svm
		"NFSv3 Latency Heatmap", "NFSv3 Read Latency Heatmap", "NFSv3 Write Latency Heatmap", "NFSv3 Latency by Op Type", "NFSv3 IOPs per Type",
		"NFSv4 Latency Heatmap", "NFSv4 Read Latency Heatmap", "NFSv4 Write Latency Heatmap", "NFSv4 Latency by Op Type", "NFSv4 IOPs per Type",
		"NFSv4.1 Latency Heatmap", "NFSv4.1 Read Latency Heatmap", "NFSv4.1 Write Latency Heatmap", "NFSv4.1 Latency by Op Type", "NFSv4.1 IOPs per Type",
		"NFSv4.2 Latency by Op Type", "NFSv4.2 IOPs per Type", "SVM NVMe/FC Throughput", "Copy Manager Requests",
		// This is from volume
		"Top $TopResources Volumes by Number of Compress Attempts", "Top $TopResources Volumes by Number of Compress Fail", "Volume Latency by Op Type", "Volume IOPs per Type",
		// This is from lun
		"IO Size",
		// This is from nfs4storePool
		"Allocations over 50%", "All nodes with 1% or more allocations in $Datacenter", "SessionConnectionHolderAlloc", "ConnectionParentSessionReferenceAlloc", "SessionHolderAlloc", "SessionAlloc", "StateRefHistoryAlloc",
	}

	VisitAllPanels(data, func(_ string, _, value gjson.Result) {
		kind := value.Get("type").String()
		if kind == "row" || kind == "text" {
			return
		}
		description := value.Get("description").String()
		targetsSlice := value.Get("targets").Array()
		title := value.Get("title").String()
		panelType := value.Get("type").String()
		if slices.Contains(ignoreList, title) {
			fmt.Printf(`dashboard=%s panel="%s" has different description\n`, dashPath, title)
			return
		}

		if description == "" {
			if len(targetsSlice) == 1 {
				expr := targetsSlice[0].Get("expr").String()
				if strings.Contains(expr, "/") || strings.Contains(expr, "+") || strings.Contains(expr, "-") || strings.Contains(expr, " on ") {
					// This indicates expressions with arithmetic operations, After adding appropriate description, this will be uncommented.
					*count++
					t.Errorf(`dashboard=%s panel="%s" has arithmetic operations %d`, dashPath, value.Get("title").String(), *count)
				} else {
					*count++
					t.Errorf(`dashboard=%s panel="%s" does not have panel description %d`, dashPath, title, *count)
				}
			} else {
				// This indicates table/timeseries with more than 1 expression, After deciding next steps, this will be uncommented.
				if panelType == "table" {
					*count++
					t.Errorf(`dashboard=%s panel="%s" has table with multiple expression %d`, dashPath, title, *count)
				} else {
					*count++
					t.Errorf(`dashboard=%s panel="%s" has many expressions %d`, dashPath, title, *count)
				}
			}
		} else if !strings.HasPrefix(description, "$") && !strings.HasSuffix(description, ".") {
			// Few panels have description text from variable, which would be ignored and description must end with period(.)
			t.Errorf(`dashboard=%s panel="%s" description hasn't ended with period`, dashPath, title)
		}
	})
}

func TestFSxFriendlyVariables(t *testing.T) {
	VisitDashboards(dashboards,
		func(path string, data []byte) {
			checkVariablesAreFSxFriendly(t, path, data)
		})
}

func checkVariablesAreFSxFriendly(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "templating.list").ForEach(func(key, value gjson.Result) bool {
		// Only consider query variables
		if value.Get("type").String() != "query" {
			return true
		}

		query := value.Get("query").String()
		definition := value.Get("definition").String()
		varName := value.Get("name").String()

		if varName != "Cluster" && varName != "Datacenter" {
			return true
		}

		if strings.Contains(query, "node_labels") {
			t.Errorf(`dashboard=%s path=templating.list[%s] variable="%s" has "node_labels" in query. Use "cluster_new_status" instead.`,
				ShortPath(path), key.String(), varName)
		}

		if strings.Contains(definition, "node_labels") {
			t.Errorf(`dashboard=%s path=templating.list[%s] variable="%s" has "node_labels" in definition. Use "cluster_new_status" instead.`,
				ShortPath(path), key.String(), varName)
		}
		return true
	})
}

var linkPath = regexp.MustCompile(`/d/(.*?)/`)
var supportedLinkedObjects = []string{"cluster", "datacenter", "aggr", "svm", "volume", "node", "qtree", "home_node"}

func TestLinks(t *testing.T) {
	hasLinks := map[string][]string{}
	uids := map[string]string{}

	VisitDashboards(dashboards, func(path string, data []byte) {
		checkLinks(t, path, data, hasLinks, uids)
	})

	// Check that the links are valid URLs and that the link points to an existing dashboard
	for path, list := range hasLinks {
		hasOrgID := false
		for _, link := range list {
			parse, err := url.Parse(link)
			if err != nil {
				t.Errorf(`dashboard=%s link="%s" is not a valid URL`, path, link)
				continue
			}
			matches := linkPath.FindStringSubmatch(parse.Path)
			if len(matches) != 2 {
				t.Errorf(`dashboard=%s link="%s" does not have a valid path`, path, link)
				continue
			}

			// Check if the dashboard exists
			if _, ok := uids[matches[1]]; !ok {
				t.Errorf(`dashboard=%s links to not existant dashboard with link="%s"`, path, link)
			}

			query, err := url.ParseQuery(link)
			if err != nil {
				t.Errorf(`dashboard=%s link="%s" is not a valid URL`, path, link)
				continue
			}
			if len(query) < 3 {
				t.Errorf(`dashboard=%s link="%s" does not have enough query parameters`, path, link)
			}
			for k, v := range query {
				if strings.HasSuffix(k, "?orgId") {
					if v[0] != "1" {
						t.Errorf(`dashboard=%s link="%s" does not have orgId=1`, path, link)
					}
					hasOrgID = true
					continue
				}
				if strings.HasPrefix(k, "${") {
					if v[0] != "" {
						t.Errorf(`dashboard=%s link="%s" has variable in query. got key="%s" value="%s" want value=""`,
							path, link, k, v[0])
					}
				} else if strings.HasPrefix(k, "var-") {
					if v[0] == "" {
						t.Errorf(`dashboard=%s link="%s" has empty variable in query. got key="%s" value="%s" want non empty value`,
							path, link, k, v[0])
					}
				}
			}
		}

		if !hasOrgID {
			t.Errorf(`dashboard=%s does not have orgId=1`, path)
		}
	}
}

func checkLinks(t *testing.T, path string, data []byte, hasLinks map[string][]string, uids map[string]string) {
	dashPath := ShortPath(path)

	VisitAllPanels(data, func(_ string, _, value gjson.Result) {
		checkPanelLinks(t, value, dashPath, hasLinks)
	})

	uid := gjson.GetBytes(data, "uid").String()
	uids[uid] = dashPath
}

func checkPanelLinks(t *testing.T, value gjson.Result, path string, hasLinks map[string][]string) {
	linkFound := false

	if value.Get("type").String() == "table" {
		value.Get("fieldConfig.overrides").ForEach(func(_, anOverride gjson.Result) bool {
			if name := anOverride.Get("matcher.options").String(); slices.Contains(supportedLinkedObjects, name) {
				linkFound = false
				anOverride.Get("properties").ForEach(func(_, propValue gjson.Result) bool {
					propValue.Get("value").ForEach(func(_, value gjson.Result) bool {
						link := value.Get("url").String()
						if link != "" {
							linkFound = true
							hasLinks[path] = append(hasLinks[path], link)
						}
						return true
					})
					return true
				})
				if !linkFound {
					t.Errorf(`dashboard=%s panel="%s" column=%s is missing the links`, path, value.Get("title").String(), name)
				}
			}
			return true
		})
	}
}
