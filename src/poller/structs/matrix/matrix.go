package matrix

import (
	"fmt"
	"errors"
    "math"
    "strconv"
    "strings"
    "goharvest2/poller/share"
    "goharvest2/poller/yaml"
)


type Matrix struct {
    Collector string
    Object string
    Plugin string
	GlobalLabels map[string]string
	LabelNames map[string]string
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
    m.GlobalLabels = map[string]string{}
    m.LabelNames = map[string]string{}
    m.InstanceKeys = make([][]string, 0)
    m.Instances = map[string]*Instance{}
    m.Metrics = map[string]*Metric{}
    return &m
}

func (m *Matrix) InitData() error {
	var x, y, i, j int
	x = m.MetricsIndex
	y = len(m.Instances)
	if x == 0 || y == 0 {
		return errors.New("Counter or Instance cache empty")
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




func (m *Matrix) SetValueString(metric *Metric, instance *Instance, value []byte) error {
	var numeric float64
	var err error

    numeric, err = strconv.ParseFloat(string(value), 64)

    if err == nil {
		m.SetValue(metric, instance, numeric)
    }
	return err
}

func (m *Matrix) SetValue(metric *Metric, instance *Instance, value float64) {
	m.Data[metric.Index][instance.Index] = value
}

func (m *Matrix) SetValueForMetric(key string, instance *Instance, value float64) {
    metric, found := m.GetMetric(key)
    if found {
        m.SetValue(metric, instance, value)
    }
}

func (m *Matrix) SetValueForMetricAndInstance(metric_key, instance_key string, value float64) {
    if instance, found := m.GetInstance(instance_key); found {
        m.SetValueForMetric(metric_key, instance, value)
    }
}

func (m *Matrix) SetArrayValues(metric *Metric, instance *Instance, values []float64) {
    for i:=0; i<len(metric.Labels); i+=1 {
        m.Data[metric.Index+i][instance.Index] = values[i]        
    }
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
    fmt.Printf("%s+ InstancLabel [%s] => [%s] %s\n", share.Bold, key, name, share.End)
	m.LabelNames[key] = name
    fmt.Printf("%s= %v %s\n", share.Red, m.InstanceKeys, share.End)
}

func (m *Matrix) GetLabel(key string) (string, bool) {
    display, found := m.LabelNames[key]
    return display, found
}

func (m *Matrix) AddInstanceKey(key []string) {
    copied := make([]string, len(key))
    copy(copied, key)

    m.InstanceKeys = append(m.InstanceKeys, copied)
    fmt.Printf("%s+ InstancKey %v%s\n", share.Bold, copied, share.End)
    fmt.Printf("%s= %v %s\n", share.Red, m.InstanceKeys, share.End)
}

func (m *Matrix) GetInstanceKeys() [][]string {
    return m.InstanceKeys
}

func (m *Matrix) SetInstanceLabel(instance *Instance, display, value string) {
    instance.Labels[display] = value
}

func (m *Matrix) GetInstanceLabel(instance *Instance, display string) (string, bool) {
    var label string
    var found bool
    label, found = instance.Labels[display]
    return label, found
}

func (m *Matrix) GetInstanceLabels(instance *Instance) map[string]string {
	return instance.Labels
}

func (m *Matrix) SetGlobalLabel(label, value string) {
	m.GlobalLabels[label] = value
}

func (m *Matrix) GetGlobalLabels() map[string]string {
	return m.GlobalLabels
}

func (m *Matrix) SetExportOptions(options *yaml.Node) {
    m.ExportOptions = options
}

func DefaultExportOptions() *yaml.Node {
    n := yaml.New("export_options", "")
    n.AddNewChild("include_all_labels", "True")
    return n
}
