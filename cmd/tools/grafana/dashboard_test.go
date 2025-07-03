package grafana

import (
	"fmt"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"github.com/netapp/harvest/v2/third_party/tidwall/sjson"
	"maps"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	TopResourceConstant      = "999999"
	RangeConstant            = "888888"
	RangeReverseConstant     = "10d6h54m48s"
	IntervalConstant         = "777777"
	IntervalDurationConstant = "666666"
	FormatPromQL             = "FORMAT_PROMQL"
)

var cDotDashboards = []string{
	"../../../grafana/dashboards/cmode",
	"../../../grafana/dashboards/cmode-details",
}

var exceptionLegendMap = map[string][]string{
	"cmode/metadata.json":                {"exporter", "target"},
	"cmode/cluster.json":                 {"node"},
	"cmode/headroom.json":                {"node"},
	"cmode/node.json":                    {"metric"},
	"cmode/power.json":                   {"node"},
	"cmode/smb.json":                     {"protocol"},
	"cmode/snapmirror_destinations.json": {"destination_location", "source_location"},
	"cmode/snapmirror.json":              {"destination_location", "source_location"},
	"cmode/svm.json":                     {"node", "port", "lif", "metric", "__name__"},
	"cmode/volume.json":                  {"__name__"},
}

var exceptionList = []string{
	"Total Power Consumed", "Average Power Consumption (kWh) Over Last Hour",
	"NICs Send Errors by Cluster", "NICs Receive Errors by Cluster", "FCPs Transmission interrupts", "FCPs Transmission errors",
	"Average Latency", "Throughput", "IOPs", "System Utilization", "NFSv3 Read and Write Latency", "NFSv3 Read and Write Throughput", "NFSv3 Read and Write IOPs", "CIFS Connections", "Protocol Backend IOPs",
	"SVM Average Latency", "SVM Throughput", "SVM IOPs", "SVM CIFS Latency", "SVM CIFS IOPs", "SVM FCP Average Latency", "SVM FCP IOPs", "SVM FCP Throughput", "SVM iSCSI Average Latency", "SVM iSCSI Throughput", "SVM NVMe/FC Average Latency", "SVM NVMe/FC Throughput", "SVM NVMe/FC IOPs", "Copy Offload Data Copied",
}

var legendName = regexp.MustCompile(`{{.*?}}`)

var throughputPattern = regexp.MustCompile(`(throughput|read_data|write_data|total_data)`)
var aggregationThroughputPattern = regexp.MustCompile(`(?i)(\w+)\(`)

func TestThroughput(t *testing.T) {
	VisitDashboards(Dashboards, func(path string, data []byte) {
		checkThroughput(t, path, data)
	})
}

func checkThroughput(t *testing.T, path string, data []byte) {
	path = ShortPath(path)
	// visit all panels for datasource test
	VisitAllPanels(data, func(_ string, _, value gjson.Result) {
		panelTitle := value.Get("title").ClonedString()
		kind := value.Get("type").ClonedString()
		targetsSlice := value.Get("targets").Array()
		for _, targetN := range targetsSlice {
			expr := targetN.Get("expr").ClonedString()
			if !throughputPattern.MatchString(expr) {
				continue
			}
			matches := aggregationThroughputPattern.FindStringSubmatch(expr)
			if len(matches) > 1 {
				aggregation := matches[1]
				if strings.EqualFold(aggregation, "avg") {
					t.Errorf("dashboard=%s panel=%s kind=%s expr=%s should use sum for throughput", path, panelTitle, kind, expr)
				}
			}
		}
	})
}

func TestThreshold(t *testing.T) {
	VisitDashboards(Dashboards, func(path string, data []byte) {
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
		panelTitle := value.Get("title").ClonedString()
		kind := value.Get("type").ClonedString()
		if kind == "table" || kind == "stat" {
			targetsSlice := value.Get("targets").Array()
			for _, targetN := range targetsSlice {
				expr := targetN.Get("expr").ClonedString()
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
						isThresholdSet = color.ClonedString() == th[0] && v.ClonedString() == th[1]
					}

					// check if any override has threshold set
					overridesSlice := value.Get("fieldConfig.overrides").Array()
					for _, overrideN := range overridesSlice {
						propertiesSlice := overrideN.Get("properties").Array()
						for _, propertiesN := range propertiesSlice {
							id := propertiesN.Get("id").ClonedString()
							if id == "thresholds" {
								color := propertiesN.Get("value.steps.#.color")
								v := propertiesN.Get("value.steps.#.value")
								isThresholdSet = color.ClonedString() == th[0] && v.ClonedString() == th[1]
							} else if id == "custom.displayMode" && kind == "table" {
								v := propertiesN.Get("value")
								if !slices.Contains(expectedColorBackground[kind], v.ClonedString()) {
									t.Errorf("dashboard=%s panel=%s kind=%s expr=%s don't have correct displaymode expected %s found %s", path, panelTitle, kind, expr, expectedColorBackground[kind], v.ClonedString())
								} else {
									isColorBackgroundSet = true
								}
							}
						}
					}

					if kind == "stat" {
						colorMode := value.Get("options.colorMode")
						if !slices.Contains(expectedColorBackground[kind], colorMode.ClonedString()) {
							t.Errorf("dashboard=%s panel=%s kind=%s expr=%s doesn't have correct colorMode got %s want %s", path, panelTitle, kind, expr, colorMode.ClonedString(), expectedColorBackground[kind])
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
	VisitDashboards(Dashboards, func(path string, data []byte) {
		checkDashboardForDatasource(t, path, data)
	})
}

func checkDashboardForDatasource(t *testing.T, path string, data []byte) {
	path = ShortPath(path)
	// visit all panels for datasource test
	VisitAllPanels(data, func(p string, _, value gjson.Result) {
		dsResult := value.Get("datasource")
		panelTitle := value.Get("title").ClonedString()
		if !dsResult.Exists() {
			t.Errorf(`dashboard="%s" panel="%s" doesn't have a datasource`, path, panelTitle)
			return
		}

		if dsResult.Type == gjson.Null {
			// if the panel is a row, it is OK if there is no datasource
			if value.Get("type").ClonedString() == "row" {
				return
			}
			t.Errorf(`dashboard=%s panel="%s" has a null datasource, should be ${DS_PROMETHEUS}`, path, panelTitle)
		} else if dsResult.ClonedString() != "${DS_PROMETHEUS}" {
			t.Errorf("dashboard=%s panel=%s has %s datasource should be ${DS_PROMETHEUS}", path, panelTitle, dsResult.ClonedString())
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
			if ds.ClonedString() != "${DS_PROMETHEUS}" {
				targetPath := fmt.Sprintf("%s.target[%d].datasource", p, i)
				t.Errorf(
					"dashboard=%s path=%s panel=%s has %s datasource shape that breaks older versions of Grafana",
					path,
					targetPath,
					panelTitle,
					dsResult.ClonedString(),
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
		name := value.Get("name").ClonedString()
		if value.Get("name").ClonedString() == "DS_PROMETHEUS" {
			doesDsPromExist = true
			query := value.Get("query").ClonedString()
			if query != "prometheus" {
				t.Errorf("dashboard=%s var=DS_PROMETHEUS query want=prometheus got=%s", path, query)
			}
			theType := value.Get("type").ClonedString()
			if theType != "datasource" {
				t.Errorf("dashboard=%s var=DS_PROMETHEUS type want=datasource got=%s", path, theType)
			}
		}

		if !excludedNames[name] {
			if value.Get("current.selected").ClonedString() == "true" {
				t.Errorf(
					"dashboard=%s var=current.selected query want=false got=%s text=%s value=%s name= %s",
					path,
					"true",
					value.Get("current.text"),
					value.Get("current.value"),
					name,
				)
			}
			ttype := value.Get("type").ClonedString()
			datasource := value.Get("datasource").ClonedString()
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
	VisitDashboards(Dashboards,
		func(path string, data []byte) {
			checkUnits(t, path, mt, data)
		})

	// Exceptions are meant to reduce false negatives
	allowedSuffix := map[string][]string{
		"_count":                     {"none", "short", "locale"},
		"_lag_time":                  {"", "s", "short"},
		"aggr_total_physical_used":   {"bytes", "binBps"}, // Growth rate uses bytes/sec unit
		"aggr_total_logical_used":    {"bytes", "binBps"}, // Growth rate uses bytes/sec unit
		"volume_space_physical_used": {"bytes", "binBps"}, // Growth rate uses bytes/sec unit
		"volume_space_logical_used":  {"bytes", "binBps"}, // Growth rate uses bytes/sec unit
		"qos_ops":                    {"iops", "percent"},
		"qos_total_data":             {"Bps", "percent"},
		"aggr_space_used":            {"bytes", "percent"},
		"volume_size_used":           {"bytes", "percent"},
		"shelf_power":                {"watt", "watth"},
		"environment_sensor_power":   {"watt", "watth"},
		"volume_num_compress_fail":   {"percent", "short"},
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
				match := reg.MatchString(strings.ReplaceAll(strings.ReplaceAll(l.expr, "\n", ""), " ", ""))
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
	kind := value.Get("type").ClonedString()
	if kind == "row" {
		return
	}
	path := fmt.Sprintf("%spanels[%d]", pathPrefix, key.Int())
	defaultUnit := value.Get("fieldConfig.defaults.unit").ClonedString()
	overridesSlice := value.Get("fieldConfig.overrides").Array()
	targetsSlice := value.Get("targets").Array()
	transformationsSlice := value.Get("transformations").Array()
	title := value.Get("title").ClonedString()
	sPath := ShortPath(dashboardPath)

	propertiesMap := make(map[string]map[string]string)
	overrides := make([]override, 0, len(overridesSlice))
	expressions := make([]Expression, 0)
	valueToName := make(map[string]string) // only used with panels[*].transformations[*].options.renameByName

	for oi, overrideN := range overridesSlice {
		matcherID := overrideN.Get("matcher.id")
		// make sure that mapKey is unique for each override element
		propertiesMapKey := matcherID.ClonedString() + strconv.Itoa(oi)
		propertiesMap[propertiesMapKey] = make(map[string]string)
		matcherOptions := overrideN.Get("matcher.options")
		propertiesN := overrideN.Get("properties").Array()
		for pi, propN := range propertiesN {
			propID := propN.Get("id").ClonedString()
			propVal := propN.Get("value").ClonedString()
			propertiesMap[propertiesMapKey][propID] = propVal
			if propID == "unit" {
				o := override{
					id:      matcherID.ClonedString(),
					options: matcherOptions.ClonedString(),
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
		defaultUnit = value.Get("yAxis.format").ClonedString()
	}

	for _, targetN := range targetsSlice {
		expr := targetN.Get("expr").ClonedString()
		expr = strings.ReplaceAll(strings.ReplaceAll(expr, "\n", ""), " ", "")
		matches := metricName.FindStringSubmatch(expr)
		if len(matches) != 2 {
			continue
		}
		// If the expression includes count( ignore since it is unit-less
		if strings.Contains(expr, "count(") {
			continue
		}

		// If the expression includes count by ignore since it is unit-less
		if strings.Contains(expr, "countby") {
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

		exprRefID := targetN.Get("refId").ClonedString()
		expressions = append(expressions, Expression{
			Metric: metric,
			refID:  exprRefID,
			expr:   expr,
		})
	}
	for _, transformN := range transformationsSlice {
		transformID := transformN.Get("id").ClonedString()
		if transformID == "organize" {
			rbn := transformN.Get("options.renameByName").Map()
			for k, v := range rbn {
				valueToName[k] = v.ClonedString()
			}
		}
	}
	numExpressions := len(expressions)
	for _, e := range expressions {
		// Ignore labels and _status
		if strings.HasSuffix(e.Metric, "_labels") || strings.HasSuffix(e.Metric, "_status") || strings.HasSuffix(e.Metric, "_events") || strings.HasSuffix(e.Metric, "_alerts") {
			continue
		}
		unit := unitForExpr(e, overrides, defaultUnit, valueToName, numExpressions)
		mt.addMetric(e.Metric, unit, path, sPath, title, e.expr)
	}
}

func unitForExpr(e Expression, overrides []override, defaultUnit string,
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
		switch o.id {
		case "byName":
			if o.options == byNameQuery || o.options == byNameQuery2 {
				return o.unit
			}
		case "byFrameRefID":
			if o.options == e.refID || o.options == byNameQuery2 {
				return o.unit
			}
		}
	}
	return defaultUnit
}

func TestVariablesRefresh(t *testing.T) {
	VisitDashboards(Dashboards,
		func(path string, data []byte) {
			checkVariablesRefresh(t, path, data)
		})
}

func checkVariablesRefresh(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "templating.list").ForEach(func(key, value gjson.Result) bool {
		if value.Get("type").ClonedString() == "datasource" {
			return true
		}
		// If the variable is not visible, ignore
		if value.Get("hide").Int() != 0 {
			return true
		}
		// If the variable is custom, ignore
		if value.Get("type").ClonedString() == "custom" {
			return true
		}

		refreshVal := value.Get("refresh").Int()
		if refreshVal != 2 {
			varName := value.Get("name").ClonedString()
			t.Errorf("dashboard=%s path=templating.list[%s].refresh variable=%s is not 2. Should be \"refresh\": 2,",
				ShortPath(path), key.ClonedString(), varName)
		}

		return true
	})
}

func TestVariablesAreSorted(t *testing.T) {
	VisitDashboards(Dashboards,
		func(path string, data []byte) {
			checkVariablesAreSorted(t, path, data)
		})
}

func checkVariablesAreSorted(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "templating.list").ForEach(func(key, value gjson.Result) bool {
		// The datasource variable does not need to be sorted, ignore
		if value.Get("type").ClonedString() == "datasource" {
			return true
		}
		// If the variable is not visible, ignore
		if value.Get("hide").Int() != 0 {
			return true
		}
		// If the variable is custom, ignore
		if value.Get("type").ClonedString() == "custom" {
			return true
		}

		sortVal := value.Get("sort").Int()
		if sortVal != 1 {
			varName := value.Get("name").ClonedString()
			t.Errorf("dashboard=%s path=templating.list[%s].sort variable=%s is not 1. Should be \"sort\": 1,",
				ShortPath(path), key.ClonedString(), varName)
		}
		return true
	})
}

func TestVariablesIncludeAllOption(t *testing.T) {
	VisitDashboards(Dashboards,
		func(path string, data []byte) {
			checkVariablesHaveAll(t, path, data)
		})
}

var exceptionToAll = map[string]bool{
	"cmode-details/volumeDeepDive.json": true,
}

func checkVariablesHaveAll(t *testing.T, path string, data []byte) {
	shouldHaveAll := map[string]bool{
		"Datacenter":  true,
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
		"cmode/cluster.json":  true,
	}

	if exceptionToAll[ShortPath(path)] {
		return
	}

	gjson.GetBytes(data, "templating.list").ForEach(func(key, value gjson.Result) bool {
		// The datasource variable does not need to be sorted, ignore
		if value.Get("type").ClonedString() == "datasource" {
			return true
		}
		// If the variable is not visible, ignore
		if value.Get("hide").Int() != 0 {
			return true
		}
		// If the variable is custom, ignore
		if value.Get("type").ClonedString() == "custom" {
			return true
		}

		varName := value.Get("name").ClonedString()
		if !shouldHaveAll[varName] {
			return true
		}

		includeAll := value.Get("includeAll").Bool()
		if !includeAll {
			t.Errorf("variable=%s should have includeAll=true dashboard=%s path=templating.list[%s].includeAll",
				varName, ShortPath(path), key.ClonedString())
		}

		allValues := value.Get("allValue").ClonedString()
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
		Dashboards,
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
		if value.Get("type").ClonedString() == "datasource" {
			return true
		}
		// name of variable
		vars = append(vars, value.Get("name").ClonedString())
		// query expression of variable
		varExpression = append(varExpression, value.Get("definition").ClonedString())
		return true
	})

	VisitAllPanels(data, func(_ string, _, value gjson.Result) {
		d := value.Get("description").ClonedString()
		if d != "" {
			description = append(description, d)
		}
	})

	expressions := AllExpressions(data)

	// check that each variable is used in at least one expression
varLoop:
	for _, variable := range vars {
		for _, expr := range expressions {
			if strings.Contains(expr.Metric, variable) {
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

func TestIDIsBlank(t *testing.T) {
	VisitDashboards(
		Dashboards,
		func(path string, data []byte) {
			checkUIDNotEmpty(t, path, data)
			checkIDIsNull(t, path, data)
		})
}

func TestExemplarIsFalse(t *testing.T) {
	VisitDashboards(
		Dashboards,
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
	uid := gjson.GetBytes(data, "uid").ClonedString()
	if uid == "" {
		t.Errorf(`dashboard=%s uid is "", but should not be empty`, path)
	}
}

func checkIDIsNull(t *testing.T, path string, data []byte) {
	id := gjson.GetBytes(data, "id").ClonedString()
	if id != "" {
		t.Errorf(`dashboard=%s id should be null but is %s`, ShortPath(path), id)
	}
}

func TestUniquePanelIDs(t *testing.T) {
	VisitDashboards(
		Dashboards,
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
		Dashboards,
		func(path string, data []byte) {
			checkTopKRange(t, path, data)
		})
}

func checkTopKRange(t *testing.T, path string, data []byte) {

	// collect all expressions
	expressions := make([]ExprP, 0)

	VisitAllPanels(data, func(_ string, key, value gjson.Result) {
		DoTarget("", "", key, value, func(path string, expr string, format string, id string, title string, rowTitle string) {
			if format == "table" || format == "stat" {
				return
			}
			expressions = append(expressions, NewExpr(path, expr, format, id, title, rowTitle))
		})
	})

	// collect all variables
	variables := allVariables(data)

	for _, expr := range expressions {
		if !strings.Contains(expr.Expr, "topk") {
			continue
		}
		if strings.Contains(expr.Expr, "/") {
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

		problem := ensureLookBack(expr.Expr)
		if problem != "" {
			t.Errorf(`dashboard=%s path=%s topk got=%s %s`, ShortPath(path), expr.path, expr.Expr, problem)
		}
	}

	for _, v := range variables {
		if v.name == "TopResources" {
			for _, optionVal := range v.options {
				selected := optionVal.Get("selected").Bool()
				text := optionVal.Get("text").ClonedString()
				value := optionVal.Get("value").ClonedString()

				// Test if text and value match, except for the special case with "All" and "$__all"
				if text != value && (text != "All" || value != "$__all") {
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
		// Ignore special case where code filter has been applied as `code=~"[45].*"`, which cause the match[1] to be 45.
		// This pattern is used in the StorageGRID Overview dashboard to check HTTP StatusCodes.
		if strings.Contains(text, "code=~\"[45].*\"") {
			continue
		}
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
		"cmode/auditlog.json":           2,
		"cmode/flexcache.json":          2,
		"cmode/fsa.json":                2,
		"cmode/health.json":             2,
		"cmode/nfsTroubleshooting.json": 3,
		"cmode/power.json":              2,
		"cmode/shelf.json":              2,
		"cmode/smb.json":                2,
		"cmode/switch.json":             2,
		"cmode/workload.json":           2,
		"storagegrid/fabricpool.json":   2,
	}
	// count the number of expanded sections in the dashboard and ensure num expanded = 1
	VisitDashboards(
		Dashboards,
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
			titles = append(titles, title.ClonedString())
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
		Dashboards,
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

	kind := value.Get("type").ClonedString()
	if kind == "row" || kind == "piechart" {
		return
	}
	optionsData := value.Get("options")
	if legendData := optionsData.Get("legend"); legendData.Exists() {
		legendDisplayMode := legendData.Get("displayMode").ClonedString()
		legendPlacementData := legendData.Get("placement").ClonedString()
		title := value.Get("title").ClonedString()
		calcsData := legendData.Get("calcs").Array()
		var calcsSlice []string
		for _, result := range calcsData {
			calcsSlice = append(calcsSlice, result.ClonedString())
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
	if strings.Contains(got, "diff") {
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
		Dashboards,
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
			t.Errorf(`dashboard=%s panel="%s fieldConfig.defaults.custom.spanNulls got=[%s] want=true`,
				dashPath, value.Get("title").ClonedString(), spanNulls.ClonedString())
		}
	})
}

func TestPanelChildPanels(t *testing.T) {
	VisitDashboards(
		Dashboards,
		func(path string, data []byte) {
			checkPanelChildPanels(t, ShortPath(path), data)
		})
}

func checkPanelChildPanels(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "panels").ForEach(func(_, value gjson.Result) bool {
		// Check all collapsed panels should have child panels
		if value.Get("collapsed").Bool() && len(value.Get("panels").Array()) == 0 {
			t.Errorf("dashboard=%s, panel=%s, has child panels outside of row", path, value.Get("title").ClonedString())
		}
		return true
	})
}

func TestRatesAreNot1m(t *testing.T) {
	VisitDashboards(
		Dashboards,
		func(path string, data []byte) {
			checkRate1m(t, ShortPath(path), data)
		},
	)
}

func checkRate1m(t *testing.T, path string, data []byte) {
	expressions := AllExpressions(data)
	for _, expr := range expressions {
		if strings.Contains(expr.Metric, "[1m]") {
			t.Errorf("dashboard=%s, expr should not use rate of [1m] expr=%s", path, expr.Metric)
		}
	}
}

func TestTableFilter(t *testing.T) {
	VisitDashboards(
		Dashboards,
		func(path string, data []byte) {
			checkTableFilter(t, path, data)
		})
}

func checkTableFilter(t *testing.T, path string, data []byte) {
	dashPath := ShortPath(path)
	VisitAllPanels(data, func(_ string, key, value gjson.Result) {
		panelType := value.Get("type").ClonedString()
		if panelType == "table" {
			isFilterable := value.Get("fieldConfig.defaults.custom.filterable").ClonedString()
			if isFilterable != "true" {
				t.Errorf(`dashboard=%s path=panels[%d] title="%s" does not enable filtering for the table`,
					dashPath, key.Int(), value.Get("title").ClonedString())
			}
		}
	})
}

func TestJoinExpressions(t *testing.T) {
	VisitDashboards(
		Dashboards,
		func(path string, data []byte) {
			checkJoinExpressions(t, path, data)
		})
}

func checkJoinExpressions(t *testing.T, path string, data []byte) {
	dashPath := ShortPath(path)
	expectedRegex := "(.*) 1$"
	VisitAllPanels(data, func(_ string, key, value gjson.Result) {
		panelType := value.Get("type").ClonedString()
		if panelType == "table" {
			targetsSlice := value.Get("targets").Array()
			if len(targetsSlice) > 1 {
				errorFound := false
				for _, targetN := range targetsSlice {
					expr := targetN.Get("expr").ClonedString()
					if strings.Contains(expr, "label_join") {
						transformationsSlice := value.Get("transformations").Array()
						regexUsed := false
						for _, transformationN := range transformationsSlice {
							if transformationN.Get("id").ClonedString() == "renameByRegex" {
								regex := transformationN.Get("options.regex").ClonedString()
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
						dashPath, key.Int(), value.Get("title").ClonedString())
				}
			}
		}
	})
}

func TestTitlesOfTopN(t *testing.T) {
	VisitDashboards(
		Dashboards,
		func(path string, data []byte) {
			checkTitlesOfTopN(t, ShortPath(path), data)
		},
	)
}

func checkTitlesOfTopN(t *testing.T, path string, data []byte) {
	expressions := AllExpressions(data)
	for _, expr := range expressions {
		if !strings.Contains(expr.Metric, "topk") || expr.Kind == "stat" {
			continue
		}
		titleRef := asTitle(expr.refID)
		title := gjson.GetBytes(data, titleRef)

		// Check that the title contains are variable
		if !strings.Contains(title.ClonedString(), "$") {
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
		Dashboards,
		func(path string, data []byte) {
			checkIOPSDecimal(t, path, data)
		})
}

func checkIOPSDecimal(t *testing.T, path string, data []byte) {
	dashPath := ShortPath(path)

	VisitAllPanels(data, func(path string, _, value gjson.Result) {
		panelType := value.Get("type").ClonedString()
		if panelType != "timeseries" {
			return
		}
		defaultUnit := value.Get("fieldConfig.defaults.unit").ClonedString()
		if defaultUnit != "iops" {
			return
		}
		decimals := value.Get("fieldConfig.defaults.decimals").ClonedString()

		if decimals != "0" {
			t.Errorf(`dashboard=%s path=%s panel="%s", decimals should be 0 got=%s`,
				dashPath, path, value.Get("title").ClonedString(), decimals)
		}
	})
}

func TestPercentHasMinMax(t *testing.T) {
	VisitDashboards(
		Dashboards,
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
		panelType := value.Get("type").ClonedString()
		if panelType != "timeseries" {
			return
		}
		defaultUnit := value.Get("fieldConfig.defaults.unit").ClonedString()
		if defaultUnit != "percent" && defaultUnit != "percentunit" {
			return
		}
		theMin := value.Get("fieldConfig.defaults.min").ClonedString()
		theMax := value.Get("fieldConfig.defaults.max").ClonedString()
		decimals := value.Get("fieldConfig.defaults.decimals").ClonedString()
		if theMin != "0" {
			t.Errorf(`dashboard=%s path=%s panel="%s" has unit=%s, min should be 0 got=%s`,
				dashPath, path, value.Get("title").ClonedString(), defaultUnit, theMin)
		}
		if decimals != "2" {
			t.Errorf(`dashboard=%s path=%s panel="%s", decimals should be 2 got=%s`,
				dashPath, path, value.Get("title").ClonedString(), decimals)
		}
		if defaultUnit == "percent" && !exceptionMap[value.Get("title").ClonedString()] && theMax != "100" {
			t.Errorf(`dashboard=%s path=%s panel="%s" has unit=%s, max should be 100 got=%s`,
				dashPath, path, value.Get("title").ClonedString(), defaultUnit, theMax)
		}
		if defaultUnit == "percentunit" && theMax != "1" {
			t.Errorf(`dashboard=%s path=%s panel="%s" has unit=%s, max should be 1 got=%s`,
				dashPath, path, value.Get("title").ClonedString(), defaultUnit, theMax)
		}
	})
}

func TestRefreshIsOff(t *testing.T) {
	VisitDashboards(
		Dashboards,
		func(path string, data []byte) {
			checkDashboardRefresh(t, ShortPath(path), data)
		},
	)
}

func checkDashboardRefresh(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "refresh").ForEach(func(_, value gjson.Result) bool {
		if value.ClonedString() != "" {
			t.Errorf(`dashboard=%s, got refresh=%s, want refresh="" (off)`, path, value.ClonedString())
		}
		return true
	})
}

func TestHeatmapSettings(t *testing.T) {
	VisitDashboards(
		Dashboards,
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
		panelType := value.Get("type").ClonedString()
		if panelType != "heatmap" {
			return
		}
		colorScheme := value.Get("color.colorScheme").ClonedString()
		colorMode := value.Get("color.mode").ClonedString()
		if colorScheme != wantColorScheme {
			t.Errorf(`dashboard=%s path=%s panel="%s" got color.scheme=%s, want=%s`,
				dashPath, path, value.Get("title").ClonedString(), colorScheme, wantColorScheme)
		}
		if colorMode != wantColorMode {
			t.Errorf(`dashboard=%s path=%s panel="%s" got color.mode=%s, want=%s`,
				dashPath, path, value.Get("title").ClonedString(), colorMode, wantColorMode)
		}
	})
}

func TestBytePanelsHave2Decimals(t *testing.T) {
	VisitDashboards(
		Dashboards,
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
		panelType := value.Get("type").ClonedString()
		if panelType != "timeseries" {
			return
		}
		defaultUnit := value.Get("fieldConfig.defaults.unit").ClonedString()
		if !byteTypes[defaultUnit] {
			return
		}
		decimals := value.Get("fieldConfig.defaults.decimals").ClonedString()
		if decimals != "2" {
			t.Errorf(`dashboard=%s path=%s panel="%s" got decimals=%s, want decimals=2`,
				dashPath, path, value.Get("title").ClonedString(), decimals)
		}
	})
}

func TestDashboardKeysAreSorted(t *testing.T) {
	VisitDashboards(
		Dashboards,
		func(path string, data []byte) {
			path = ShortPath(path)
			sorted := gjson.GetBytes(data, `@pretty:{"sortKeys":true, "indent":"  ", "width":0}`).ClonedString()
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
	VisitDashboards(Dashboards, func(path string, data []byte) {
		checkDashboardTime(t, path, data)
	})
}

func checkDashboardTime(t *testing.T, path string, data []byte) {
	dashPath := ShortPath(path)
	from := gjson.GetBytes(data, "time.from")
	to := gjson.GetBytes(data, "time.to")

	expectedTimeRanges := map[string]struct {
		from string
		to   string
	}{
		"cmode/auditlog.json": {"now-24h", "now"},
		"default":             {"now-3h", "now"},
	}

	expected, exists := expectedTimeRanges[dashPath]
	if !exists {
		expected = expectedTimeRanges["default"]
	}

	if from.ClonedString() != expected.from {
		t.Errorf("dashboard=%s time.from got=%s want=%s", dashPath, from.ClonedString(), expected.from)
	}
	if to.ClonedString() != expected.to {
		t.Errorf("dashboard=%s time.to got=%s want=%s", dashPath, to.ClonedString(), expected.to)
	}
}

func TestNoDrillDownRows(t *testing.T) {
	VisitDashboards(Dashboards, func(path string, data []byte) {
		checkRowNames(t, path, data)
	})
}

func checkRowNames(t *testing.T, path string, data []byte) {
	path = ShortPath(path)
	VisitAllPanels(data, func(_ string, key, value gjson.Result) {
		kind := value.Get("type").ClonedString()
		if kind == "row" {
			title := value.Get("title").ClonedString()
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
		"NFSv3 Latency Heatmap", "NFSv3 Read Latency Heatmap", "NFSv3 Write Latency Heatmap", "NFSv3 Latency by Op Type", "NFSv3 IOPs Per Type",
		"NFSv4 Latency Heatmap", "NFSv4 Read Latency Heatmap", "NFSv4 Write Latency Heatmap", "NFSv4 Latency by Op Type", "NFSv4 IOPs Per Type",
		"NFSv4.1 Latency Heatmap", "NFSv4.1 Read Latency Heatmap", "NFSv4.1 Write Latency Heatmap", "NFSv4.1 Latency by Op Type", "NFSv4.1 IOPs Per Type",
		"NFSv4.2 Latency by Op Type", "NFSv4.2 IOPs Per Type", "SVM NVMe/FC Throughput", "Copy Manager Requests",
		// This is from volume
		"Top $TopResources Volumes by Number of Compress Attempts", "Top $TopResources Volumes by Number of Compress Fail", "Volume Latency by Op Type", "Volume IOPs Per Type",
		// This is from lun
		"IO Size",
		// This is from nfs4storePool
		"Allocations over 50%", "All nodes with 1% or more allocations in $Datacenter", "SessionConnectionHolderAlloc", "ConnectionParentSessionReferenceAlloc", "SessionHolderAlloc", "SessionAlloc", "StateRefHistoryAlloc",
	}

	VisitAllPanels(data, func(_ string, _, value gjson.Result) {
		kind := value.Get("type").ClonedString()
		if kind == "row" || kind == "text" {
			return
		}
		description := value.Get("description").ClonedString()
		targetsSlice := value.Get("targets").Array()
		title := value.Get("title").ClonedString()
		panelType := value.Get("type").ClonedString()
		if slices.Contains(ignoreList, title) {
			return
		}

		if description == "" {
			if len(targetsSlice) == 1 {
				expr := targetsSlice[0].Get("expr").ClonedString()
				if strings.Contains(expr, "/") || strings.Contains(expr, "+") || strings.Contains(expr, "-") || strings.Contains(expr, " on ") {
					// This indicates expressions with arithmetic operations, After adding appropriate description, this will be uncommented.
					*count++
					t.Errorf(`dashboard=%s panel="%s" has arithmetic operations %d`, dashPath, value.Get("title").ClonedString(), *count)
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
		} else if !strings.HasPrefix(description, "$") && !strings.HasSuffix(strings.TrimSpace(description), ".") {
			// A few panels take their description text from a variable.
			// Those can be ignored.
			// Descriptions must end with a period (.)
			t.Errorf(`dashboard=%s panel="%s" description "%s" should end with a period`, dashPath, title, description)
		}
	})
}

func TestFSxFriendlyVariables(t *testing.T) {
	VisitDashboards(cDotDashboards,
		func(path string, data []byte) {
			checkVariablesAreFSxFriendly(t, path, data)
		})
}

func checkVariablesAreFSxFriendly(t *testing.T, path string, data []byte) {

	exceptionValues := map[string]bool{
		"cmode/metadata.json":                true,
		"cmode/snapmirror_destinations.json": true,
		"cmode/snapmirror.json":              true,
	}

	gjson.GetBytes(data, "templating.list").ForEach(func(key, value gjson.Result) bool {
		// Only consider query variables
		if value.Get("type").ClonedString() != "query" {
			return true
		}

		query := value.Get("query").ClonedString()
		definition := value.Get("definition").ClonedString()
		varName := value.Get("name").ClonedString()

		sPath := ShortPath(path)
		isExceptionPath := exceptionValues[sPath]

		if isExceptionPath || (varName != "Cluster" && varName != "Datacenter") {
			return true
		}

		if !strings.Contains(query, "cluster_new_status") {
			t.Errorf(`dashboard=%s path=templating.list[%s] variable="%s" does not have "cluster_new_status" in query. Found "%s" instead.`,
				sPath, key.ClonedString(), varName, definition)
		}

		if !strings.Contains(definition, "cluster_new_status") {
			t.Errorf(`dashboard=%s path=templating.list[%s] variable="%s" does not have "cluster_new_status" in definition. Found "%s" instead.`,
				sPath, key.ClonedString(), varName, definition)
		}
		return true
	})
}

func TestLabelsNullVariables(t *testing.T) {
	VisitDashboards(cDotDashboards,
		func(path string, data []byte) {
			checkVariablesLabelNull(t, path, data)
		})
}

func checkVariablesLabelNull(t *testing.T, path string, data []byte) {
	gjson.GetBytes(data, "templating.list").ForEach(func(key, value gjson.Result) bool {
		if value.Get("label").Type != gjson.Null && value.Get("label").ClonedString() == "" {
			varName := value.Get("name").ClonedString()
			t.Errorf("dashboard=%s path=templating.list[%s]. variable=%s label should not be empty",
				ShortPath(path), key.ClonedString(), varName)
		}
		return true
	})
}

var linkPath = regexp.MustCompile(`/d/(.*?)/`)
var supportedLinkedObjects = []string{"cluster", "datacenter", "aggr", "svm", "volume", "node", "qtree", "home_node", "tenant"}
var exceptionPathPanelObject = []string{
	"cmode/s3ObjectStorage.json-Bucket Overview-volume",                        // bucket volumes starts with fg_oss_xx and volume dashboard don't support them
	"storagegrid/fabricpool.json-Aggregates-cluster",                           // There is no datacenter var to be passed for linking to cluster dashboard
	"storagegrid/fabricpool.json-Aggregates-aggr",                              // There is no datacenter var to be passed for linking to aggregate dashboard
	"storagegrid/fabricpool.json-Buckets-cluster",                              // There is no storagegrid cluster dashboard available
	"storagegrid/overview.json-Data space usage breakdown-cluster",             // There is no storagegrid cluster dashboard available
	"storagegrid/overview.json-Metadata allowed space usage breakdown-cluster", // There is no storagegrid cluster dashboard available
	"storagegrid/tenant.json-Tenant Quota-cluster",                             // There is no storagegrid cluster dashboard available
	"storagegrid/tenant.json-Buckets-cluster",                                  // There is no storagegrid cluster dashboard available
}

func TestLinks(t *testing.T) {
	hasLinks := map[string][]string{}
	uids := map[string]string{}

	VisitDashboards(Dashboards, func(path string, data []byte) {
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
				t.Errorf(`dashboard=%s links to not existent dashboard with link="%s"`, path, link)
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

	uid := gjson.GetBytes(data, "uid").ClonedString()
	uids[uid] = dashPath
}

func checkPanelLinks(t *testing.T, value gjson.Result, path string, hasLinks map[string][]string) {
	linkFound := false

	if value.Get("type").ClonedString() == "table" {
		value.Get("fieldConfig.overrides").ForEach(func(_, anOverride gjson.Result) bool {
			if name := anOverride.Get("matcher.options").ClonedString(); slices.Contains(supportedLinkedObjects, name) {
				title := value.Get("title").ClonedString()
				linkFound = false
				anOverride.Get("properties").ForEach(func(_, propValue gjson.Result) bool {
					propValue.Get("value").ForEach(func(_, value gjson.Result) bool {
						link := value.Get("url").ClonedString()
						if link != "" {
							linkFound = true
							hasLinks[path] = append(hasLinks[path], link)
						}
						return true
					})
					return true
				})
				if !linkFound && !slices.Contains(exceptionPathPanelObject, path+"-"+title+"-"+name) {
					t.Errorf(`dashboard=%s panel="%s" column=%s is missing the links`, path, title, name)
				}
			}
			return true
		})
	}
}

func TestTags(t *testing.T) {
	VisitDashboards(Dashboards,
		func(path string, data []byte) {
			checkTags(t, path, data)
		})
}

func checkTags(t *testing.T, path string, data []byte) {
	allowedTagsMap := map[string]bool{
		"cdot":        true,
		"cisco":       true,
		"fsx":         true,
		"harvest":     true,
		"ontap":       true,
		"storagegrid": true,
	}

	path = ShortPath(path)
	tags := gjson.GetBytes(data, "tags").Array()
	if len(tags) == 0 {
		t.Errorf(`dashboard=%s got tags are empty, but should have tags`, path)
		return
	}

	for _, tag := range tags {
		if !allowedTagsMap[tag.ClonedString()] {
			allowedTags := slices.Sorted(maps.Keys(allowedTagsMap))
			t.Errorf(`dashboard=%s got tag=%s, which is not in the allowed set=%v`, path, tag.ClonedString(), allowedTags)
		}
	}
}

func TestFormatedPromQL(t *testing.T) {
	SkipIfMissing(t, FormatPromQL)

	VisitDashboards(
		Dashboards,
		func(path string, data []byte) {
			checkIfPromQLIsFormatted(t, path, data)
		},
	)
}

func checkIfPromQLIsFormatted(t *testing.T, path string, data []byte) {
	var (
		updatedData  []byte
		notFormatted bool
		errorStr     []string
		err          error
	)

	updatedData = slices.Clone(data)
	dashPath := ShortPath(path)

	// Change all panel expressions
	VisitAllPanels(updatedData, func(path string, _, value gjson.Result) {
		title := value.Get("title").ClonedString()
		// Rewrite expressions
		value.Get("targets").ForEach(func(targetKey, target gjson.Result) bool {
			expr := target.Get("expr")
			if expr.Exists() && expr.ClonedString() != "" {
				updatedExpr := formatPromQL(expr.ClonedString())
				if updatedExpr != expr.ClonedString() {
					notFormatted = true
					updatedData, err = sjson.SetBytes(updatedData, path+".targets."+targetKey.ClonedString()+".expr", []byte(updatedExpr))
					if err != nil {
						fmt.Printf("Error while updating the panel query format: %v\n", err)
					}
					errorStr = append(errorStr, fmt.Sprintf("query not formatted in dashboard %s panel `%s`, it should be \n %s\n", dashPath, title, updatedExpr))
				}
			}
			return true
		})
	})
	if notFormatted {
		sortedPath := writeFormatted(t, dashPath, updatedData)
		path = "grafana/dashboards/" + dashPath
		t.Errorf("%v \nFormatted version created at path=%s.\ncp %s %s",
			errorStr, sortedPath, sortedPath, path)
	}
}

func formatPromQL(query string) string {
	replacedQuery := strings.ReplaceAll(query, "$TopResources", TopResourceConstant)
	replacedQuery = strings.ReplaceAll(replacedQuery, "$__range", RangeConstant)
	replacedQuery = strings.ReplaceAll(replacedQuery, "$__interval", IntervalConstant)
	replacedQuery = strings.ReplaceAll(replacedQuery, "${Interval}", IntervalDurationConstant)

	command := exec.Command("promtool", "--experimental", "promql", "format", replacedQuery)
	output, err := command.CombinedOutput()
	updatedQuery := strings.TrimSuffix(string(output), "\n")
	if strings.HasPrefix(updatedQuery, "  ") {
		updatedQuery = strings.TrimLeft(updatedQuery, " ")
	}
	if err != nil {
		// An exit code can't be used since we need to ignore metrics that are not formatted but can't change
		fmt.Printf("ERR formating metrics query=%s err=%v output=%s", query, err, string(output))
		return query
	}

	if len(output) == 0 {
		return query
	}

	updatedQuery = strings.ReplaceAll(updatedQuery, TopResourceConstant, "$TopResources")
	updatedQuery = strings.ReplaceAll(updatedQuery, RangeReverseConstant, "$__range")
	updatedQuery = strings.ReplaceAll(updatedQuery, IntervalConstant, "$__interval")
	updatedQuery = strings.ReplaceAll(updatedQuery, IntervalDurationConstant, "${Interval}")
	return updatedQuery
}

func writeFormatted(t *testing.T, path string, updatedData []byte) string {
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
	_, err = create.Write(updatedData)
	if err != nil {
		t.Errorf("failed to write formatted json to file=%s err=%v", dest, err)
		return ""
	}
	err = create.Close()
	if err != nil {
		t.Errorf("failed to close file=%s err=%v", dest, err)
		return ""
	}
	return dest
}

func SkipIfMissing(t *testing.T, vars ...string) {
	t.Helper()
	anyMatches := false
	for _, v := range vars {
		e := os.Getenv(v)
		if e != "" {
			anyMatches = true
			break
		}
	}
	if !anyMatches {
		t.Skipf("Set one of %s envvars to run this test", strings.Join(vars, ", "))
	}
}

func TestLegendFormat(t *testing.T) {
	VisitDashboards(Dashboards, func(path string, data []byte) {
		checkLegendFormat(t, path, data)
	})
}

func checkLegendFormat(t *testing.T, path string, data []byte) {
	path = ShortPath(path)
	possibleLegends := exceptionLegendMap[path]

	gjson.GetBytes(data, "templating.list").ForEach(func(_, value gjson.Result) bool {
		varName := strings.ToLower(value.Get("name").ClonedString())
		if varName == "aggregate" {
			varName = "aggr"
		}
		possibleLegends = append(possibleLegends, varName)
		return true
	})

	VisitAllPanels(data, func(_ string, _, value gjson.Result) {
		panelTitle := value.Get("title").ClonedString()
		if slices.Contains(exceptionList, panelTitle) {
			return
		}
		kind := value.Get("type").ClonedString()
		if kind == "row" || kind == "table" || kind == "stat" || kind == "bargauge" || kind == "piechart" || kind == "gauge" || kind == "heatmap" {
			return
		}
		targetsSlice := value.Get("targets").Array()
		for _, targetN := range targetsSlice {
			legendFormat := targetN.Get("legendFormat").ClonedString()
			legendExist := false
			if !strings.Contains(legendFormat, "{{") {
				t.Errorf("dashboard=%s panel=%s kind=%s legendFormat=%s should have {{object}} in legendFormat", path, panelTitle, kind, legendFormat)
			} else {
				matches := legendName.FindAllString(legendFormat, -1)
				for _, match := range matches {
					if slices.Contains(possibleLegends, match[2:len(match)-2]) {
						legendExist = true
						break
					}
				}
				if !legendExist {
					t.Errorf("dashboard=%s panel=%s kind=%s legendFormat=%s should have legends from %s in legendFormat", path, panelTitle, kind, legendFormat, possibleLegends)
				}
			}
		}
	})
}
