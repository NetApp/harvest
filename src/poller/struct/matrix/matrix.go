package matrix

import (
	//"fmt"
    "math"
    "strings"
    "strconv"
	"goharvest2/poller/errors"
    //"goharvest2/poller/util"
    "goharvest2/poller/struct/yaml"
    "goharvest2/poller/struct/dict"
)


type Matrix struct {
    Collector string
    Object string
    Plugin string
	GlobalLabels *dict.Dict
	LabelNames *dict.Dict
    InstanceKeys [][]string
	ExportOptions *yaml.Node
	Instances map[string]*Instance
	Metrics map[string]*Metric
	MetricsIndex int /* since some metrics are arrays and we can't relay on len(Metrics) */
    Data [][]float64
    IsMetadata bool
    MetadataType string
    MetadataObject string
}

func New(collector, object, plugin string) *Matrix {
    m := Matrix{Collector: collector, Object: object, Plugin: plugin, MetricsIndex: 0 }
    m.GlobalLabels = dict.New()
    m.LabelNames = dict.New()
    m.InstanceKeys = make([][]string, 0)
    m.Instances = map[string]*Instance{}
    m.Metrics = map[string]*Metric{}
    return &m
}

func (m *Matrix) IsEmpty() bool {
    return len(m.Data) == 0
}

func (m *Matrix) Clone() *Matrix {
    n := &Matrix{ 
        Collector      : m.Collector, 
        Object         : m.Object, 
        Plugin         : m.Plugin,
        Instances      : m.Instances,
        Metrics        : m.Metrics,
        MetricsIndex   : m.MetricsIndex,
        GlobalLabels   : m.GlobalLabels,
        LabelNames     : m.LabelNames,
        InstanceKeys   : m.InstanceKeys,
        ExportOptions  : m.ExportOptions,
        IsMetadata     : m.IsMetadata,
        MetadataType   : m.MetadataType,
        MetadataObject : m.MetadataObject,
    }
    n.Data = make([][]float64, n.MetricsIndex)
    if !m.IsEmpty() {
        for i:=0; i<n.MetricsIndex; i+=1 {
            n.Data[i] = make([]float64, len(n.Instances))
            copy(n.Data[i], m.Data[i])
        }
    }
    return n
}

func (m *Matrix) InitData() error {
	var x, y, i, j int
	x = m.MetricsIndex
	y = len(m.Instances)
	if x == 0 || y == 0 {
		return errors.New(errors.MATRIX_EMPTY, "counter or instance cache empty")
	}
	m.Data = make([][]float64, x)
	for i=0; i<x; i+=1 {
		m.Data[i] = make([]float64, y)
		for j=0; j<y; j+=1 {
			m.Data[i][j] = math.NaN()
		}
	}
	return nil
}

func (m *Matrix) ResetData() {
    m.Data = make([][]float64, 0)
}

func (m *Matrix) ResetMetrics() {
    if len(m.Metrics) != 0 {
        m.Metrics = make(map[string]*Metric)
    }
    m.ResetData()
}


func (m *Matrix) ResetInstances() {
    if len(m.Instances) != 0 {
        m.Instances = make(map[string]*Instance)
    }
    m.ResetData()
}


func (m *Matrix) ResetLabelNames() {
    m.LabelNames = dict.New()
}


func (m *Matrix) GetInstances() map[string]*Instance {
    return m.Instances
}

func (m *Matrix) GetMetrics() map[string]*Metric {
    return m.Metrics
}

func (m *Matrix) SetValueString(metric *Metric, instance *Instance, value string) error {
	var numeric float64
	var err error

    numeric, err = strconv.ParseFloat(value, 32)

    if err == nil {
		m.SetValue(metric, instance, numeric)
    }
	return err
}

func (m *Matrix) SetValue(metric *Metric, instance *Instance, value float64) {
	m.Data[metric.Index][instance.Index] = value
}

func (m *Matrix) SetValueS(key string, instance *Instance, value float64) {
    if metric := m.GetMetric(key); metric != nil {
        m.SetValue(metric, instance, value)
    }
}

func (m *Matrix) SetValueSS(metric_key, instance_key string, value float64) {
    if instance := m.GetInstance(instance_key); instance != nil {
        m.SetValueS(metric_key, instance, value)
    }
}

func (m *Matrix) SetArrayValues(metric *Metric, instance *Instance, values []float64) {
    for i:=0; i<len(metric.Labels); i+=1 {
        m.Data[metric.Index+i][instance.Index] = values[i]        
    }
}

func (m *Matrix) SetArrayValuesString(metric *Metric, instance *Instance, values []string) error {
    var ok bool

    numeric := make([]float64, len(values))
    for _, v := range values {
        if n, err := strconv.ParseFloat(v, 32); err == nil {
            numeric = append(numeric, n)
            ok = true // at least one parsed
        } else {
            numeric = append(numeric, math.NaN())
        }
    }

    m.SetArrayValues(metric, instance, numeric)

    if ok {
        return nil
    }
    return errors.New(errors.MATRIX_PARSE_STR, "no number parsed from: [" + strings.Join(values, ", ") + "]")
}

func (m *Matrix) GetValue(metric *Metric, instance *Instance) (float64, bool) {
	var value float64
	value = m.Data[metric.Index][instance.Index]
	return value, value==value
}

func (m *Matrix) GetArrayValues(metric *Metric, instance *Instance) []float64 {
    values := make([]float64, len(metric.Labels))
    for i:=0; i<len(metric.Labels); i+=1 {
        values[i] = m.Data[metric.Index+i][instance.Index]        
    }
    return values
}

func (m *Matrix) AddLabelName(name string) {
    m.AddLabelKeyName(name, name)
}

func (m *Matrix) AddLabelKeyName(key, name string) {
    //fmt.Printf("%s+ InstancLabel [%s] => [%s] %s\n", util.Bold, key, name, util.End)
	m.LabelNames.Set(key, name)
    //fmt.Printf("%s= %v %s\n", util.Red, m.InstanceKeys, util.End)
}

func (m *Matrix) GetLabel(key string) (string, bool) {
    return m.LabelNames.GetHas(key)
}

func (m *Matrix) AddInstanceKey(key []string) {
    copied := make([]string, len(key))
    copy(copied, key)

    m.InstanceKeys = append(m.InstanceKeys, copied)
    //fmt.Printf("%s+ InstancKey %v%s\n", util.Bold, copied, util.End)
    //fmt.Printf("%s= %v %s\n", util.Red, m.InstanceKeys, util.End)
}

func (m *Matrix) GetInstanceKeys() [][]string {
    return m.InstanceKeys
}

func (m *Matrix) SetInstanceLabel(instance *Instance, key, value string) {
    display := m.LabelNames.Get(key)
    instance.Labels.Set(display, value)
}

func (m *Matrix) GetInstanceLabel(instance *Instance, display string) (string, bool) {
    return instance.Labels.GetHas(display)
}

func (m *Matrix) GetInstanceLabels(instance *Instance) *dict.Dict {
	return instance.Labels
}

func (m *Matrix) SetGlobalLabel(label, value string) {
	m.GlobalLabels.Set(label, value)
}

func (m *Matrix) GetGlobalLabels() *dict.Dict {
	return m.GlobalLabels
}

func (m *Matrix) SetExportOptions(options *yaml.Node) {
    m.ExportOptions = options
}

func DefaultExportOptions() *yaml.Node {
    n := yaml.New("export_options", "")
    n.CreateChild("include_all_labels", "True")
    return n
}
