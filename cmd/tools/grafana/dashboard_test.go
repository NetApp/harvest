package grafana

import (
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestDatasource(t *testing.T) {
	visitDashboards("../../../grafana/dashboards", func(path string) {
		checkDashboardForDatasource(t, path)
	})
}

func checkDashboardForDatasource(t *testing.T, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read dashboards path=%s err=%v", path, err)
	}
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
}

func visitDashboards(dir string, eachDash func(path string)) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal("failed to read directory:", err)
		}
		ext := filepath.Ext(path)
		if ext != ".json" {
			return nil
		}
		eachDash(path)
		return nil
	})
	if err != nil {
		log.Fatal("failed to read dashboards:", err)
	}
}

func TestUnitsAndExprMatch(t *testing.T) {
	mt := newMetricsTable()
	dir := "../../../grafana/dashboards/cmode"
	visitDashboards(dir, func(path string) {
		checkUnits(path, mt)
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

func checkUnits(dashboardPath string, mt *metricsTable) {
	data, err := os.ReadFile(dashboardPath)
	if err != nil {
		log.Fatalf("failed to read dashboards dashboardPath=%s err=%v", dashboardPath, err)
	}

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
	splits := strings.Split(dashboardPath, string(filepath.Separator))
	shortPath := strings.Join(splits[len(splits)-2:], string(filepath.Separator))

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
		mt.addMetric(e.metric, unit, path, shortPath, title)
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
