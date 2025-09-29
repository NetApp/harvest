package template

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/aggregator"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/labelagent"
	max2 "github.com/netapp/harvest/v2/cmd/poller/plugin/maxplugin"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/metricagent"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"os"
	"regexp"
	"strings"
)

type Model struct {
	Name          string      `yaml:"name"`
	Query         string      `yaml:"query"`
	Object        string      `yaml:"object"`
	Ignore        string      `yaml:"ignore"`
	ExportData    string      `yaml:"export_data"`
	Endpoints     []*Endpoint `yaml:"endpoints"`
	ExportOptions struct {
		InstanceKeys     []string `yaml:"instance_keys"`
		InstanceLabels   []string `yaml:"instance_labels"`
		IncludeAllLabels bool     `yaml:"include_all_labels"`
	} `yaml:"export_options"`
	Override          map[string]string `yaml:"override"`
	metrics           []Metric
	pluginLabels      []string
	PluginMetrics     []plugin.DerivedMetric
	MultiplierMetrics []plugin.DerivedMetric
}

type Metric struct {
	left         string
	right        string
	line         string
	renameColumn int
	hasSigil     bool
	column       int
	parents      []string
}

type Endpoint struct {
	Query    string   `yaml:"query"`
	Counters []string `yaml:"counters"`
	Metrics  []Metric
}

func ReadTemplate(path string) (Model, error) {
	var model Model
	data, err := os.ReadFile(path)
	if err != nil {
		return Model{}, err
	}
	model, err = unmarshalModel(data)
	if err != nil {
		return Model{}, err
	}
	err = readPlugins(path, &model)
	if err != nil {
		return Model{}, err
	}
	return model, nil
}

func unmarshalModel(data []byte) (Model, error) {
	tm := Model{}

	astFile, err := parser.ParseBytes(data, 0)
	if err != nil {
		return tm, fmt.Errorf("failed to parse template err=%w", err)
	}
	// treat an empty file as an error
	if len(astFile.Docs) == 0 {
		return tm, errs.New(errs.ErrConfig, "template file is empty or does not exist")
	}

	contentNode := astFile.Docs[0].Body

	ignoreNode := searchNode(contentNode, "ignore")
	if ignoreNode != nil && node.ToString(ignoreNode) == "true" {
		tm.Ignore = node.ToString(ignoreNode)
		nameNode := searchNode(contentNode, "name")
		if nameNode != nil {
			tm.Name = node.ToString(nameNode)
		}
		return tm, nil
	}
	err = readNameQueryObject(&tm, contentNode)
	if err != nil {
		return tm, err
	}
	countersNode := searchNode(contentNode, "counters")
	if countersNode == nil {
		return tm, errors.New("template has no counters")
	}
	metrics := make([]Metric, 0)
	flattenCounters(countersNode, &metrics, make([]string, 0))
	addEndpoints(&tm, searchNode(contentNode, "endpoints"), make([]string, 0))
	addExportOptions(&tm, searchNode(contentNode, "export_options"))
	addOverride(&tm, searchNode(contentNode, "override"))

	tm.metrics = metrics
	return tm, nil
}

func addOverride(tm *Model, n ast.Node) {
	if n == nil {
		return
	}

	if tm.Override == nil {
		tm.Override = make(map[string]string)
	}
	sn, ok := n.(*ast.SequenceNode)
	if ok {
		for _, child := range sn.Values {
			mn, ok := child.(*ast.MappingNode)
			if ok {
				if len(mn.Values) == 0 {
					continue
				}
				key := node.ToString(mn.Values[0].Key)
				val := node.ToString(mn.Values[0].Value)
				tm.Override[key] = val
			}
		}
	}
}

func readPlugins(path string, model *Model) error {
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

func readNameQueryObject(tm *Model, root ast.Node) error {
	nameNode := searchNode(root, "name")
	if nameNode != nil {
		tm.Name = node.ToString(nameNode)
	}
	queryNode := searchNode(root, "query")
	if queryNode != nil {
		tm.Query = node.ToString(queryNode)
	}
	objectNode := searchNode(root, "object")
	if objectNode != nil {
		tm.Object = node.ToString(objectNode)
	}
	if tm.Name == "" {
		return errors.New("template has no name")
	}
	if tm.Query == "" {
		return errors.New("template has no query")
	}
	// A template with query=prometheus is allowed to have no object
	if tm.Object == "" && tm.Query != "prometheus" {
		return errors.New("template has no object")
	}
	return nil
}

func addEndpoints(tm *Model, n ast.Node, parents []string) {
	if n == nil {
		return
	}
	sn, ok := n.(*ast.SequenceNode)
	if ok {
		for _, ikn := range sn.Values {
			mn, ok := ikn.(*ast.MappingNode)
			if ok {
				query := node.ToString(mn.Values[0].Key)
				metrics := make([]Metric, 0)
				countersNode := searchNode(mn, "counters")
				flattenCounters(countersNode, &metrics, parents)
				ep := &Endpoint{Query: query, Metrics: metrics}
				tm.Endpoints = append(tm.Endpoints, ep)
			}
		}
	}
}

func searchNode(r ast.Node, key string) ast.Node {
	if mn, ok := r.(*ast.MappingNode); ok {
		for _, child := range mn.Values {
			if node.ToString(child.Key) == key {
				return child.Value
			}
		}
	}

	return nil
}

func addExportOptions(tm *Model, n ast.Node) {
	if n == nil {
		return
	}
	instanceKeys := searchNode(n, "instance_keys")
	if instanceKeys != nil {
		sn, ok := instanceKeys.(*ast.SequenceNode)
		if ok {
			for _, ikn := range sn.Values {
				tm.ExportOptions.InstanceKeys = append(tm.ExportOptions.InstanceKeys, node.ToString(ikn))
			}
		}
	}
	instanceLabels := searchNode(n, "instance_labels")
	if instanceLabels != nil {
		sn, ok := instanceKeys.(*ast.SequenceNode)
		if ok {
			for _, ikn := range sn.Values {
				tm.ExportOptions.InstanceLabels = append(tm.ExportOptions.InstanceLabels, node.ToString(ikn))
			}
		}
	}
}

func flattenCounters(n ast.Node, metrics *[]Metric, parents []string) {
	switch n.Type() { //nolint:exhaustive
	case ast.MappingType:
		mn := n.(*ast.MappingNode)
		key := node.ToString(mn.Values[0].Key)
		if key == "hidden_fields" || key == "filter" {
			return
		}
		parents = append(parents, key)
		flattenCounters(mn.Values[0].Value, metrics, parents)
	case ast.SequenceType:
		sn := n.(*ast.SequenceNode)
		for _, c := range sn.Values {
			flattenCounters(c, metrics, parents)
		}
	case ast.StringType:
		*metrics = append(*metrics, newMetric(n, parents))
	default:
		// ignore
	}
}

var sigilReplacer = strings.NewReplacer("^", "", "- ", "")

func newMetric(n ast.Node, parents []string) Metric {
	// separate left and right and remove all sigils
	text := node.ToString(n)
	noSigils := sigilReplacer.Replace(text)
	before, after, found := strings.Cut(noSigils, "=>")
	column := n.GetToken().Position.Column
	m := Metric{
		line:     text,
		left:     strings.TrimSpace(noSigils),
		hasSigil: strings.Contains(text, "^"),
		column:   column,
		parents:  parents,
	}
	if found {
		m.left = strings.TrimSpace(before)
		m.right = trimComment(after)
		m.renameColumn = strings.Index(text, "=>") + column
	}
	return m
}

func trimComment(text string) string {
	lastSink := strings.Index(text, "#")
	if lastSink > -1 {
		return strings.TrimSpace(text[:lastSink])
	}
	return strings.TrimSpace(text)
}

func findBuiltInPlugins(template *node.Node, model *Model) error {
	var ee []error
	template.PreprocessTemplate()

	err := readLabelAgent(template, model)
	if err != nil {
		ee = append(ee, err)
	}

	err = readAggregator(template, model)
	if err != nil {
		ee = append(ee, err)
	}

	err = readMetricAgent(template, model)
	if err != nil {
		ee = append(ee, err)
	}

	err = readMax(template, model)
	if err != nil {
		ee = append(ee, err)
	}

	return errors.Join(ee...)
}

func readMax(template *node.Node, model *Model) error {
	children := template.SearchChildren([]string{"plugins", "Max"})
	if len(children) != 0 {
		abc := plugin.AbstractPlugin{Params: children[0]}
		mm := max2.New(&abc)
		err := mm.Init(conf.Remote{})
		if err != nil {
			return err
		}
		model.MultiplierMetrics = append(model.MultiplierMetrics, mm.NewMetrics()...)
	}

	return nil
}

func readMetricAgent(template *node.Node, model *Model) error {
	children := template.SearchChildren([]string{"plugins", "MetricAgent"})
	if len(children) == 0 {
		return nil
	}
	abc := plugin.AbstractPlugin{Params: children[0]}
	ma := metricagent.New(&abc)
	err := ma.Init(conf.Remote{})
	if err != nil {
		return err
	}
	model.PluginMetrics = append(model.PluginMetrics, ma.NewMetrics()...)
	return nil
}

func readAggregator(template *node.Node, model *Model) error {
	children := template.SearchChildren([]string{"plugins", "Aggregator"})
	if len(children) != 0 {
		abc := plugin.AbstractPlugin{Params: children[0]}
		agg := aggregator.New(&abc)
		err := agg.Init(conf.Remote{})
		if err != nil {
			return err
		}
		model.pluginLabels = append(model.pluginLabels, agg.NewLabels()...)
		model.MultiplierMetrics = append(model.MultiplierMetrics, agg.NewMetrics()...)
	}

	return nil
}

func readLabelAgent(template *node.Node, model *Model) error {
	children := template.SearchChildren([]string{"plugins", "LabelAgent"})
	if len(children) == 0 {
		return nil
	}
	abc := plugin.AbstractPlugin{Params: children[0]}
	la := labelagent.New(&abc)
	err := la.Init(conf.Remote{})
	if err != nil {
		return err
	}
	model.pluginLabels = la.NewLabels()
	return nil
}

var setRe = regexp.MustCompile(`[sS]etLabel\("?(\w+)"?,`)

func findCustomPlugins(path string, template *node.Node, model *Model) error {
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
		pluginGo := toPluginPath(path, goPluginName)

		if err := readPlugin(pluginGo, model); err != nil {
			return err
		}

		// special case for labels added outside normal per-object plugin
		if strings.Contains(path, "snapmirror.yaml") || strings.Contains(path, "svm.yaml") {
			pluginGo2 := toPluginPath(path, "commonutils")
			if err := readPlugin(pluginGo2, model); err != nil {
				return err
			}
		}
		if strings.Contains(path, "volume.yaml") && strings.Contains(path, "perf") {
			pluginGo2 := toPluginPath(path, "volume")
			if err := readPlugin(pluginGo2, model); err != nil {
				return err
			}
		}
	}
	return nil
}

func toPluginPath(path string, pluginName string) string {
	// ../../../conf/rest/9.10.0/sensor.yaml -> ../../../cmd/collectors/rest/plugins/sensor/sensor.go
	// conf/rest/9.10.0/sensor.yaml          -> cmd/collectors/rest/plugins/sensor/sensor.go
	// ../../../conf/rest/9.12.0/lif.yaml -> ../../../cmd/collectors/rest/plugins/lif/lif.go
	// conf/rest/9.12.0/lif.yaml          -> cmd/collectors/rest/plugins/lif/lif.go

	before, after, _ := strings.Cut(path, "conf/")

	// Both Zapi and REST sensor.yaml templates uses a single plugin defined in power.go
	if strings.Contains(path, "sensor.yaml") {
		return before + "cmd/collectors/power.go"
	}

	// Both Zapi and REST lif.yaml templates uses a single plugin defined in lif.go
	if strings.Contains(path, "lif.yaml") {
		return before + "cmd/collectors/lif.go"
	}

	// Both Zapi and REST volume.yaml templates uses a single plugin defined in volume.go
	if strings.Contains(path, "volume.yaml") && strings.Contains(path, "perf") {
		return before + "cmd/collectors/volume.go"
	}

	// Both REST and KeyPerf storage_unit.yaml templates uses a single plugin defined in volume.go
	if strings.Contains(path, "storage_unit.yaml") {
		return before + "cmd/collectors/storageunit.go"
	}

	base := strings.Split(after, "/")
	p := fmt.Sprintf("%scmd/collectors/%s/plugins/%s/%s.go", before, base[0], pluginName, pluginName)

	// special case for labels added outside normal per-object plugin
	if pluginName == "commonutils" {
		return before + "cmd/collectors/commonutils.go"
	}

	return p
}

func readPlugin(fileName string, model *Model) error {
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
