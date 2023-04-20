package template

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/aggregator"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/labelagent"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	y3 "gopkg.in/yaml.v3"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode"
)

const toConf = "../../../conf"

var allTemplatesButEms = []string{"rest", "restperf", "storagegrid", "zapi", "zapiperf"}

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

	visitTemplates(t, func(path string, model TemplateModel) {
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
			err := filepath.WalkDir(kindDir, func(path string, d fs.DirEntry, err error) error {
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
				open, err := os.Open(template)
				if err != nil {
					t.Errorf("failed to read template file=%s from %s/default.yaml", template, shortPath(kindDir))
				}
				decoder := y3.NewDecoder(open)
				root := &y3.Node{}
				err = decoder.Decode(root)
				if err != nil {
					t.Errorf("failed to parse template file=%s from %s/default.yaml", template, shortPath(kindDir))
				}
				err = open.Close()
				if err != nil {
					t.Errorf("failed to close template file=%s from %s/default.yaml", template, shortPath(kindDir))
					return
				}
			}
		}
	}
}

// TestTotals prints the number of unique objects and counters that Harvest collects, excluding EMS.
func TestTotals(t *testing.T) {
	totalObject := make(objectMap)
	var totalCounters int

	visitTemplates(t, func(path string, model TemplateModel) {
		totalObject[model.Name] = path
		totalCounters += len(model.metrics)

		for _, ep := range model.Endpoints {
			totalCounters += len(ep.metrics)
		}
	}, allTemplatesButEms...)

	fmt.Printf("%d objects, %d counters\n", len(totalObject), totalCounters)
}

func readDefaults(dirs []string) (map[string]objectMap, error) {
	defaults := make(map[string]objectMap)

	for _, dir := range dirs {
		dirPath := toConf + "/" + dir
		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !strings.HasSuffix(path, "default.yaml") {
				return nil
			}
			open, err := os.Open(path)
			if err != nil {
				return err
			}
			decoder := y3.NewDecoder(open)
			root := &y3.Node{}
			err = decoder.Decode(root)
			if err != nil {
				return err
			}
			defaults[collectorPath(path)] = newObjectMap(root.Content[0])
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return defaults, nil
}

func newObjectMap(n *y3.Node) objectMap {
	om := objectMap{}
	objects := searchNode(n, "objects")
	for i := 0; i < len(objects.Content); i += 2 {
		om[objects.Content[i].Value] = objects.Content[i+1].Value
	}
	return om
}

func TestMetricColumnAlignmentAndCase(t *testing.T) {
	visitTemplates(t, func(path string, model TemplateModel) {
		columnSet := make(map[int]int)
		for _, m := range model.metrics {
			if m.renameColumn > 0 {
				columnSet[m.renameColumn]++
			}
			// left side can use - since Harvest wil convert to underscore automatically,
			// right side can't
			if len(m.right) > 0 {
				if strings.Contains(m.right, "-") {
					t.Errorf("metric=%s should use _ not - on right side path=%s", m.right, shortPath(path))
				}
			}
			if len(m.right) > 0 && unicode.IsUpper([]rune(m.right)[0]) {
				t.Errorf("metric=%s should start with a lowercase, path=%s", m.line, shortPath(path))
			}
		}

		for _, ep := range model.Endpoints {
			for _, m := range ep.metrics {
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
	visitTemplates(t, func(path string, model TemplateModel) {
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

	visitTemplates(t, func(path string, model TemplateModel) {
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
					zapiPaths := append(m.parents, m.left)
					display := util.ParseZAPIDisplay(model.Object, zapiPaths)
					allLabelNames[display] = true
				} else {
					allLabelNames[m.left] = true
				}
			}
		}
		for _, ep := range model.Endpoints {
			for _, m := range ep.metrics {
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

type sorted struct {
	got  string
	want string
}

type metric struct {
	left         string
	right        string
	line         string
	renameColumn int
	hasSigil     bool
	column       int
	parents      []string
}

func (m metric) pathString() string {
	return strings.Join(m.parents, "/") + "/" + m.left
}

// Tests that keys and metrics are sorted in the following order:
// - double hats (alphabetically)
// - single hats (alphabetically)
// - metrics (alphabetically)
// ZAPI parent attributes are sorted alphabetically
// Tests that exported keys and labels are in sorted order
func TestMetricsAreSortedAndNoDuplicates(t *testing.T) {
	visitTemplates(t, func(path string, model TemplateModel) {
		sortedCounters := checkSortedCounters(model.metrics)
		if sortedCounters.got != sortedCounters.want {
			t.Errorf("counters should be sorted path=[%s]", shortPath(path))
			t.Errorf("use this instead\n")
			t.Errorf("\n%s", sortedCounters.want)
		}

		for _, endpoint := range model.Endpoints {
			sortedCounters := checkSortedCounters(endpoint.metrics)
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

func checkForDuplicateMetrics(t *testing.T, model TemplateModel, path string) {
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
		for _, m := range endpoint.metrics {
			p := m.pathString()
			_, ok := dupSet[p]
			if ok {
				t.Errorf("duplicate endpoint metric=%s in %s", p, shortPath(path))
			}
			dupSet[p] = true
		}
	}
}

func checkSortedCounters(counters []metric) sorted {
	got := countersStr(counters)
	sortZapiCounters(counters)
	want := countersStr(counters)
	return sorted{got: got, want: want}
}

func countersStr(counters []metric) string {
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

func sortZapiCounters(counters []metric) {
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

var sigilReplacer = strings.NewReplacer("^", "", "- ", "")

func visitTemplates(t *testing.T, eachTemplate func(path string, model TemplateModel), dirs ...string) {
	if len(dirs) == 0 {
		t.Fatalf("must pass list of directories")
	}
	for _, dir := range dirs {
		dirPath := toConf + "/" + dir
		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			var model TemplateModel
			ext := filepath.Ext(path)
			if ext != ".yaml" {
				return nil
			}
			if strings.HasSuffix(path, "custom.yaml") || strings.HasSuffix(path, "default.yaml") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read template path=%s err=%w", shortPath(path), err)
			}
			model, err = unmarshalModel(data)
			if err != nil {
				return fmt.Errorf("failed to unmarshalModel template path=%s err=%w", shortPath(path), err)
			}
			err = addPluginLabels(path, &model)
			if err != nil {
				//t.Errorf("failed to addPluginLabels template path=%s err=%v", shortPath(path), err)
				return err
			}
			eachTemplate(path, model)
			return nil
		})
		if err != nil {
			t.Errorf("failed to visitTemplate dir=%s err=%v", dir, err)
		}
	}
}

func addPluginLabels(path string, model *TemplateModel) error {
	template, err := tree.ImportYaml(path)
	if err != nil {
		return fmt.Errorf("failed to ImportYaml err: %w", err)
	}
	err = findBuiltInPlugins(template, model)
	if err != nil {
		return fmt.Errorf("failed to find findBuiltInPlugins err: %w", err)
	}
	err = findCustomPlugins(path, template, model)
	if err != nil {
		return fmt.Errorf("failed to findCustomPlugins err: %w", err)
	}
	return nil
}

func flattenCounters(n *y3.Node, metrics *[]metric, parents []string) {
	switch n.Tag {
	case "!!map":
		key := n.Content[0].Value
		if key == "hidden_fields" || key == "filter" {
			return
		}
		parents = append(parents, key)
		flattenCounters(n.Content[1], metrics, parents)
	case "!!seq":
		for _, c := range n.Content {
			flattenCounters(c, metrics, parents)
		}
	case "!!str":
		*metrics = append(*metrics, newZapiMetric(n, parents))
	}
}

func newZapiMetric(n *y3.Node, parents []string) metric {
	// separate left and right and remove all sigils
	text := n.Value
	noSigils := sigilReplacer.Replace(text)
	before, after, found := strings.Cut(noSigils, "=>")
	m := metric{
		line:     text,
		left:     strings.TrimSpace(noSigils),
		hasSigil: strings.Contains(text, "^"),
		column:   n.Column,
		parents:  parents,
	}
	if found {
		m.left = strings.TrimSpace(before)
		m.right = trimComment(after)
		m.renameColumn = strings.Index(text, "=>") + n.Column
	}
	return m
}

var setRe = regexp.MustCompile(`SetLabel\("(\w+)",`)

func findCustomPlugins(path string, template *node.Node, model *TemplateModel) error {
	plug := template.SearchChildren([]string{"plugins"})
	if len(plug) == 0 {
		return nil
	}
	builtIn := map[string]bool{
		"LabelAgent":  true,
		"MetricAgent": true,
		"Aggregator":  true,
		"Max":         true,
		"Tenant":      true,
	}
	for _, child := range plug[0].Children {
		name := child.GetNameS()
		if name == "" {
			name = child.GetContentS()
		}
		if builtIn[name] {
			continue
		}

		goPluginName := strings.ToLower(name)
		splits := strings.Split(path, "/")
		pluginGo := fmt.Sprintf("../../../cmd/collectors/%s/plugins/%s/%s.go", splits[4], goPluginName, goPluginName)
		err2 := readPlugin(pluginGo, model)
		if err2 != nil {
			return err2
		}
		// special case for labels added outside normal per-object plugin
		if strings.Contains(path, "snapmirror.yaml") || strings.Contains(path, "svm.yaml") {
			err2 = readPlugin("../../../cmd/collectors/commonutils.go", model)
			if err2 != nil {
				return err2
			}
		}
	}
	return nil
}

func readPlugin(fileName string, model *TemplateModel) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := scanner.Text()
		trimmed := strings.TrimSpace(text)
		matches := setRe.FindStringSubmatch(trimmed)
		if len(matches) == 2 {
			model.pluginLabels = append(model.pluginLabels, matches[1])
		}
	}
	_ = file.Close()
	return nil
}

func findBuiltInPlugins(template *node.Node, model *TemplateModel) error {
	template.PreprocessTemplate()
	children := template.SearchChildren([]string{"plugins", "LabelAgent"})
	if len(children) == 0 {
		return nil
	}
	abc := plugin.AbstractPlugin{Params: children[0]}
	la := labelagent.New(&abc)
	err := la.Init()
	if err != nil {
		return err
	}
	model.pluginLabels = la.NewLabels()

	children = template.SearchChildren([]string{"plugins", "Aggregator"})
	if len(children) == 0 {
		return nil
	}
	abc = plugin.AbstractPlugin{Params: children[0]}
	agg := aggregator.New(&abc)
	err = agg.Init()
	if err != nil {
		return err
	}
	model.pluginLabels = append(model.pluginLabels, agg.NewLabels()...)

	return nil
}

func unmarshalModel(data []byte) (TemplateModel, error) {
	tm := TemplateModel{}
	root := &y3.Node{}
	err := y3.Unmarshal(data, root)
	if err != nil {
		return tm, fmt.Errorf("failed to unmarshal err: %w", err)
	}
	if len(root.Content) == 0 {
		return tm, errs.New(errs.ErrConfig, "template file is empty or does not exist")
	}
	contentNode := root.Content[0]
	err = readNameQueryObject(&tm, contentNode)
	if err != nil {
		return tm, err
	}
	countersNode := searchNode(contentNode, "counters")
	if countersNode == nil {
		return tm, fmt.Errorf("template has no counters")
	}
	metrics := make([]metric, 0)
	flattenCounters(countersNode, &metrics, make([]string, 0))
	addEndpoints(&tm, searchNode(contentNode, "endpoints"), make([]string, 0))
	addExportOptions(&tm, searchNode(contentNode, "export_options"))

	tm.metrics = metrics
	return tm, nil
}

func addExportOptions(tm *TemplateModel, n *y3.Node) {
	if n == nil {
		return
	}
	instanceKeys := searchNode(n, "instance_keys")
	if instanceKeys != nil {
		for _, ikn := range instanceKeys.Content {
			tm.ExportOptions.InstanceKeys = append(tm.ExportOptions.InstanceKeys, ikn.Value)
		}
	}
	instanceLabels := searchNode(n, "instance_labels")
	if instanceLabels != nil {
		for _, il := range instanceLabels.Content {
			tm.ExportOptions.InstanceLabels = append(tm.ExportOptions.InstanceLabels, il.Value)
		}
	}
}

func readNameQueryObject(tm *TemplateModel, root *y3.Node) error {
	nameNode := searchNode(root, "name")
	if nameNode != nil {
		tm.Name = nameNode.Value
	}
	queryNode := searchNode(root, "query")
	if queryNode != nil {
		tm.Query = queryNode.Value
	}
	objectNode := searchNode(root, "object")
	if objectNode != nil {
		tm.Object = objectNode.Value
	}
	if tm.Name == "" {
		return fmt.Errorf("template has no name")
	}
	if tm.Query == "" {
		return fmt.Errorf("template has no query")
	}
	if tm.Object == "" {
		return fmt.Errorf("template has no object")
	}
	return nil
}

func addEndpoints(tm *TemplateModel, n *y3.Node, parents []string) {
	if n == nil {
		return
	}
	for _, m := range n.Content {
		query := m.Content[1].Value
		metrics := make([]metric, 0)
		countersNode := m.Content[3]
		flattenCounters(countersNode, &metrics, parents)
		ep := &Endpoint{Query: query, metrics: metrics}
		tm.Endpoints = append(tm.Endpoints, ep)
	}
}

func searchNode(r *y3.Node, key string) *y3.Node {
	for i, n := range r.Content {
		if n.Tag == "!!str" && n.Value == key {
			return r.Content[i+1]
		}
	}
	return nil
}

func trimComment(text string) string {
	lastSink := strings.Index(text, "#")
	if lastSink > -1 {
		return strings.TrimSpace(text[:lastSink])
	}
	return strings.TrimSpace(text)
}

type Endpoint struct {
	Query    string   `yaml:"query"`
	Counters []string `yaml:"counters"`
	metrics  []metric
}

type TemplateModel struct {
	Name          string      `yaml:"name"`
	Query         string      `yaml:"query"`
	Object        string      `yaml:"object"`
	Endpoints     []*Endpoint `yaml:"endpoints"`
	ExportOptions struct {
		InstanceKeys     []string `yaml:"instance_keys"`
		InstanceLabels   []string `yaml:"instance_labels"`
		IncludeAllLabels bool     `yaml:"include_all_labels"`
	} `yaml:"export_options"`
	metrics      []metric
	pluginLabels []string
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
