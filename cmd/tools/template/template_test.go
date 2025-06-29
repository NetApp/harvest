package template

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	template2 "github.com/netapp/harvest/v2/pkg/template"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

const toConf = "../../../conf"

var allTemplatesButEms = []string{"rest", "restperf", "storagegrid", "zapi", "zapiperf", "keyperf", "statperf"}

// validates each template file name:
//   - ends with yaml
//   - is all lowercase
//   - uses `_` and not `-` as word separators
func TestTemplateFileNames(t *testing.T) {
	for _, dir := range []string{toConf} {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Fatal("failed to read directory:", err)
			}
			if info.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if ext != ".yaml" {
				t.Errorf("got extension=[%s] want=[.yaml] path=%s", ext, shortPath(path))
			}
			base := filepath.Base(path)
			if base != strings.ToLower(base) {
				t.Errorf("got base=[%s] want=[%s] path=%s", base, strings.ToLower(base), shortPath(path))
			}
			if strings.Contains(base, "-") {
				t.Errorf("got base=[%s] with - used _ instead path=%s", base, shortPath(path))
			}
			return nil
		})
		if err != nil {
			log.Fatal("failed to read template:", err)
		}
	}
}

type objectMap map[string]string

// TestTemplateNamesMatchDefault checks that:
//   - the name used in default.yaml matches the `name:` field in the template
//   - all files listed in the default.yaml exist and are parseable
func TestTemplateNamesMatchDefault(t *testing.T) {
	modelsByTemplate := make(map[string]objectMap)
	defaults, err := readDefaults(allTemplatesButEms)

	if err != nil {
		log.Fatal("failed to read defaults:", err)

	}

	visitTemplates(t, func(path string, model Model) {
		sp := collectorPath(path)
		o := modelsByTemplate[sp]
		if o == nil {
			o = make(objectMap)
		}
		o[model.Name] = path
		modelsByTemplate[sp] = o
	}, allTemplatesButEms...)

	for kind, om := range defaults {
		templates := modelsByTemplate[kind]
		for name, templatePath := range om {
			_, ok := templates[name]
			if !ok {
				t.Errorf("template %s/default.yaml defines object=[%s] but %s does not include that name",
					kind, name, templatePath)
			}
		}
	}

	// Ensure files contained in default.yaml exist and are parseable
	for kind, om := range defaults {
		for _, fileRef := range om {
			kindDir := filepath.Join(toConf, kind)
			// find all templates named fileRef
			var matchingTemplates []string
			err := filepath.WalkDir(kindDir, func(path string, _ fs.DirEntry, _ error) error {
				if strings.Contains(path, fileRef) {
					matchingTemplates = append(matchingTemplates, path)
				}
				return nil
			})
			if err != nil {
				t.Errorf("failed to walk dir=%s err=%v", shortPath(kindDir), err)
				continue
			}
			if len(matchingTemplates) == 0 {
				t.Errorf("no templates matching file ref=%s from %s/default.yaml", fileRef, shortPath(kindDir))
				continue
			}
			for _, template := range matchingTemplates {
				_, err = parser.ParseFile(template, 0)
				if err != nil {
					t.Errorf("failed to parse template file=%s from %s/default.yaml", template, shortPath(kindDir))
				}
			}
		}
	}
}

// TestTotals prints the number of unique objects and counters that Harvest collects, excluding EMS.
func TestTotals(t *testing.T) {
	totalObject := make(objectMap)
	var totalCounters int

	visitTemplates(t, func(path string, model Model) {
		totalObject[model.Name] = path
		totalCounters += len(model.metrics)

		for _, ep := range model.Endpoints {
			totalCounters += len(ep.Metrics)
		}
	}, allTemplatesButEms...)

	fmt.Printf("%d objects, %d counters\n", len(totalObject), totalCounters)
}

func readDefaults(dirs []string) (map[string]objectMap, error) {
	defaults := make(map[string]objectMap)

	for _, dir := range dirs {
		dirPath := toConf + "/" + dir
		err := filepath.WalkDir(dirPath, func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !strings.HasSuffix(path, "default.yaml") {
				return nil
			}

			astFile, err := parser.ParseFile(path, 0)
			if err != nil {
				return fmt.Errorf("failed to parse template err=%w", err)
			}
			defaults[collectorPath(path)] = newObjectMap(astFile.Docs[0].Body)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return defaults, nil
}

func newObjectMap(n ast.Node) objectMap {
	om := objectMap{}
	objects := searchNode(n, "objects")

	mn, ok := objects.(*ast.MappingNode)
	if ok {
		for _, child := range mn.Values {
			k := node.ToString(child.Key)
			v := node.ToString(child.Value)
			om[k] = v
		}
	}

	return om
}

func TestMetricColumnAlignmentAndCase(t *testing.T) {
	visitTemplates(t, func(path string, model Model) {
		columnSet := make(map[int]int)
		for _, m := range model.metrics {
			if m.renameColumn > 0 {
				columnSet[m.renameColumn]++
			}
			// left side can use - since Harvest wil convert to underscore automatically,
			// right side can't
			if m.right != "" {
				if strings.Contains(m.right, "-") {
					t.Errorf("metric=%s should use _ not - on right side path=%s", m.right, shortPath(path))
				}
			}
			first, _ := utf8.DecodeRuneInString(m.right)
			if m.right != "" && unicode.IsUpper(first) {
				t.Errorf("metric=%s should start with a lowercase, path=%s", m.line, shortPath(path))
			}
		}

		for _, ep := range model.Endpoints {
			for _, m := range ep.Metrics {
				if m.renameColumn > 0 {
					columnSet[m.renameColumn]++
				}
				if strings.Contains(m.left, "-") {
					t.Errorf("metric=%s should use _ not - path=%s", m.left, shortPath(path))
				}
			}
		}

		if len(columnSet) > 1 {
			t.Errorf("=> should be column aligned but isn't, got columnSet=%+v path=%s", columnSet, shortPath(path))
		}
	}, allTemplatesButEms...)
}

func TestNoTabs(t *testing.T) {
	visitTemplates(t, func(path string, _ Model) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read path=%s err=%v", shortPath(path), err)
			return
		}
		tabs := bytes.Count(data, []byte("\t"))
		if tabs != 0 {
			t.Errorf("want zero tabs got=%d path=%s", tabs, shortPath(path))
		}
	}, allTemplatesButEms...)
}

// TestExportLabelsExist ensures that the values in the `export_options` section of a template are listed in the
// counters section of the template or created from plugins.
// Few notes on special cases:
//   - ZAPI collector computes a display name for counters that do not include =>.
//     This unit test uses the same logic so for example, in template conf/zapi/cdot/9.8.0/shelf.yaml
//     the counter named `^shelf-model` will be renamed to `model`
//   - ZAPI Perf collector adds the template's object field as a label
func TestExportLabelsExist(t *testing.T) {
	// these templates have a unique shape or programmatically set labels
	ignoreTemplates := []string{
		"perf/.*/workload.*.yaml",
		"zapi/cdot/9.8.0/qos_policy_fixed.yaml",
	}

	visitTemplates(t, func(path string, model Model) {
		shortenedPath := shortPath(path)
		isZapi := strings.Contains(path, "zapi")
		isZapiPerf := strings.Contains(path, "zapiperf")

		for _, template := range ignoreTemplates {
			re := regexp.MustCompile(template)
			if re.MatchString(shortenedPath) {
				return
			}
		}
		allLabelNames := make(map[string]bool)
		if isZapiPerf {
			// ZapiPerf templates implicitly include the template's object field as a label
			allLabelNames[model.Object] = true
		}

		for _, m := range model.metrics {
			if m.right != "" {
				allLabelNames[m.right] = true
			} else {
				if isZapi {
					zapiPaths := m.parents
					zapiPaths = append(zapiPaths, m.left)
					display := template2.ParseZAPIDisplay(model.Object, zapiPaths)
					allLabelNames[display] = true
				} else {
					allLabelNames[m.left] = true
				}
			}
		}
		for _, ep := range model.Endpoints {
			for _, m := range ep.Metrics {
				if m.right != "" {
					allLabelNames[m.right] = true
				} else {
					allLabelNames[m.left] = true
				}
			}
		}
		for _, label := range model.pluginLabels {
			allLabelNames[label] = true
		}
		for _, ik := range model.ExportOptions.InstanceKeys {
			if !allLabelNames[ik] {
				t.Errorf("export_options instance_key=%s does not exist path=%s", ik, shortenedPath)
			}
		}
		for _, il := range model.ExportOptions.InstanceLabels {
			if !allLabelNames[il] {
				t.Errorf("export_options instance_label=%s does not exist path=%s", il, shortenedPath)
			}
		}
	}, allTemplatesButEms...)
}

func TestOverrideMetricsExist(t *testing.T) {
	// ignore base counters which are not collected in templates but used in collector to cook values
	ignore := map[string]bool{
		"compound.total":    true,
		"compound_total":    true,
		"read_io_type_base": true,
	}

	visitTemplates(t, func(path string, model Model) {
		// Check if the path contains "zapiperf" or "restperf"
		isPerf := strings.Contains(path, "zapiperf") || strings.Contains(path, "restperf")

		// If the path is not a performance path, skip the rest of the loop
		if !isPerf {
			return
		}

		override := make(map[string]bool)
		for key := range model.Override {
			override[key] = false
		}

		for _, m := range model.metrics {
			if _, exists := model.Override[m.left]; exists {
				override[m.left] = true
			}
		}

		// Check if each key in the override map exists and is not ignored
		for k, v := range override {
			if !v && !ignore[k] {
				t.Errorf("override option=%s does not exist in counters path=%s", k, shortPath(path))
			}
		}
	}, allTemplatesButEms...)
}

func TestNoHyphenInMetrics(t *testing.T) {
	visitTemplates(t, func(path string, model Model) {
		isRest := strings.Contains(path, "rest")

		if !isRest {
			return
		}

		for _, m := range model.metrics {
			if strings.Contains(m.left, "-") || strings.Contains(m.right, "-") {
				t.Errorf("Metric fields 'left' and 'right' should not contain hyphens. Found in path=%s, left=%s, right=%s", shortPath(path), m.left, m.right)
			}
		}
	}, allTemplatesButEms...)
}

type sorted struct {
	got  string
	want string
}

func (m Metric) pathString() string {
	return strings.Join(m.parents, "/") + "/" + m.left
}

func TestQueryPrefix(t *testing.T) {
	visitTemplates(t, func(path string, model Model) {
		if model.Ignore == "true" {
			return
		}
		if !strings.HasPrefix(model.Query, "api/") {
			t.Errorf("query should be prefixed with api/, got=%s path=[%s]", model.Query, shortPath(path))
		}
	}, []string{"rest", "restperf"}...)
}

// Tests that keys and metrics are sorted in the following order:
// - double hats (alphabetically)
// - single hats (alphabetically)
// - metrics (alphabetically)
// ZAPI parent attributes are sorted alphabetically
// Tests that exported keys and labels are in sorted order
func TestMetricsAreSortedAndNoDuplicates(t *testing.T) {
	visitTemplates(t, func(path string, model Model) {
		sortedCounters := checkSortedCounters(model.metrics)
		if sortedCounters.got != sortedCounters.want {
			t.Errorf("counters should be sorted path=[%s]", shortPath(path))
			t.Errorf("use this instead\n")
			t.Errorf("\n%s", sortedCounters.want)
		}

		for _, endpoint := range model.Endpoints {
			sortedCounters := checkSortedCounters(endpoint.Metrics)
			if sortedCounters.got != sortedCounters.want {
				t.Errorf("endpoint=%s counters should be sorted path=[%s]", endpoint.Query, shortPath(path))
				t.Errorf("use this instead\n")
				t.Errorf("\n%s", sortedCounters.want)
			}
		}

		checkForDuplicateMetrics(t, model, path)

		// check sorted exported instance keys
		sortedKeys := checkSortedKeyLabels(model.ExportOptions.InstanceKeys)
		if sortedKeys.got != sortedKeys.want {
			t.Errorf("instance_keys should be sorted path=[%s]", shortPath(path))
			t.Errorf("use this instead\n")
			t.Errorf("\n%s", sortedKeys.want)
		}

		// check sorted exported instance labels
		sortedLabels := checkSortedKeyLabels(model.ExportOptions.InstanceLabels)
		if sortedLabels.got != sortedLabels.want {
			t.Errorf("instance_labels should be sorted path=[%s]", shortPath(path))
			t.Errorf("use this instead\n")
			t.Errorf("\n%s", sortedLabels.want)
		}

	}, allTemplatesButEms...)
}

func checkForDuplicateMetrics(t *testing.T, model Model, path string) {
	dupSet := make(map[string]bool)
	for _, m := range model.metrics {
		p := m.pathString()
		_, ok := dupSet[p]
		if ok {
			t.Errorf("duplicate metric=%s in %s", p, shortPath(path))
		}
		dupSet[p] = true
	}

	for _, endpoint := range model.Endpoints {
		// endpoints are independent metrics
		dupSet = make(map[string]bool)
		for _, m := range endpoint.Metrics {
			p := m.pathString()
			_, ok := dupSet[p]
			if ok {
				t.Errorf("duplicate endpoint metric=%s in %s", p, shortPath(path))
			}
			dupSet[p] = true
		}
	}
}

func checkSortedCounters(counters []Metric) sorted {
	got := countersStr(counters)
	sortZapiCounters(counters)
	want := countersStr(counters)
	return sorted{got: got, want: want}
}

func countersStr(counters []Metric) string {
	builder := strings.Builder{}
	parentSeen := make(map[string]bool)
	for _, counter := range counters {
		for i, p := range counter.parents {
			if parentSeen[p] {
				continue
			}
			prefix := strings.Repeat(" ", i*2)
			builder.WriteString(prefix)
			if i > 0 {
				builder.WriteString("- ")
			}
			builder.WriteString(p)
			builder.WriteString(":")
			builder.WriteString("\n")
			parentSeen[p] = true
		}
		prefix := strings.Repeat(" ", counter.column-2)
		builder.WriteString(prefix)
		builder.WriteString("- ")
		builder.WriteString(counter.line)
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
	return builder.String()
}

func sortZapiCounters(counters []Metric) {
	sort.SliceStable(counters, func(i, j int) bool {
		a := counters[i]
		b := counters[j]
		pa := strings.Join(a.parents, " ")
		pb := strings.Join(b.parents, " ")
		if pa != pb {
			return pa < pb
		}
		a2Hat := strings.Contains(a.line, "^^")
		b2Hat := strings.Contains(b.line, "^^")
		if a2Hat && b2Hat {
			return a.left < b.left
		}
		if a2Hat {
			return true
		}
		if b2Hat {
			return false
		}
		if a.hasSigil && b.hasSigil {
			return a.left < b.left
		}
		if a.hasSigil {
			return true
		}
		if b.hasSigil {
			return false
		}
		return a.left < b.left
	})
}

func checkSortedKeyLabels(keyLabels []string) sorted {
	got := labelsToString(keyLabels)
	sort.Strings(keyLabels)
	want := labelsToString(keyLabels)
	return sorted{got: got, want: want}
}

func labelsToString(labels []string) string {
	b := strings.Builder{}
	for _, label := range labels {
		b.WriteString("  - ")
		b.WriteString(label)
		b.WriteString("\n")
	}
	return b.String()
}

func visitTemplates(t *testing.T, eachTemplate func(path string, model Model), dirs ...string) {
	if len(dirs) == 0 {
		t.Fatalf("must pass list of directories")
	}
	for _, dir := range dirs {
		dirPath := toConf + "/" + dir
		err := filepath.WalkDir(dirPath, func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			ext := filepath.Ext(path)
			if ext != ".yaml" {
				return nil
			}
			if strings.HasSuffix(path, "custom.yaml") || strings.HasSuffix(path, "default.yaml") || strings.HasSuffix(path, "static_counter_definitions.yaml") {
				return nil
			}
			model, err := ReadTemplate(path)
			if err != nil {
				return fmt.Errorf("failed to read template path=%s err=%w", shortPath(path), err)
			}
			eachTemplate(path, model)
			return nil
		})
		if err != nil {
			t.Errorf("failed to visitTemplate dir=%s err=%v", dir, err)
		}
	}
}

func collectorPath(path string) string {
	const conf string = "conf/"
	index := strings.Index(path, conf)
	if index > 0 {
		splits := strings.Split(path[index+len(conf):], "/")
		return splits[0]
	}
	return path
}

func shortPath(path string) string {
	const conf string = "conf/"
	index := strings.Index(path, conf)
	if index > 0 {
		return path[index+len(conf):]
	}
	return path
}
