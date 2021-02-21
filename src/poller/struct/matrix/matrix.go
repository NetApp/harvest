package matrix

import (
	//"fmt"
    "math"
    "strings"
    "strconv"
    "goharvest2/share/tree/node"
	"goharvest2/share/errors"
    "goharvest2/poller/struct/dict"
)

var NAN = float32(math.NaN())

type Matrix struct {
    Collector string
    Object string
    Plugin string
	GlobalLabels *dict.Dict
	Labels *dict.Dict
    Instances map[string]*Instance
	Metrics map[string]*Metric
	MetricsIndex int  // since some metrics are arrays and we can't relay on len(Metrics)
    InstanceKeys [][]string  // @TODO deprecate, doesn't belong here
	ExportOptions *node.Node
    Data [][]float32
    IsMetadata bool
    MetadataType string
    MetadataObject string
}

func New(collector, object, plugin string) *Matrix {
    m := Matrix{Collector: collector, Object: object, Plugin: plugin, MetricsIndex: 0 }
    m.GlobalLabels = dict.New()
    m.Labels = dict.New()
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
        InstanceKeys   : m.InstanceKeys,
        Metrics        : m.Metrics,
        MetricsIndex   : m.MetricsIndex,
        GlobalLabels   : m.GlobalLabels,
        Labels         : m.Labels,
        ExportOptions  : m.ExportOptions,
        IsMetadata     : m.IsMetadata,
        MetadataType   : m.MetadataType,
        MetadataObject : m.MetadataObject,
    }
    n.Data = make([][]float32, n.MetricsIndex)
    if !m.IsEmpty() {
        for i:=0; i<n.MetricsIndex; i+=1 {
            n.Data[i] = make([]float32, len(n.Instances))
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
    m.Data = make([][]float32, x)
	for i=0; i<x; i+=1 {
		m.Data[i] = make([]float32, y)
		for j=0; j<y; j+=1 {
			m.Data[i][j] = NAN
		}
	}
	return nil
}

func (m *Matrix) RemoveMetric(key string) {
    if metric, ok := m.Metrics[key]; ok {
        delete(m.Metrics, key)
        if !m.IsEmpty() {
            if len(m.Data) > metric.Size && m.MetricsIndex > metric.Index+metric.Size {
                for i := metric.Index; i + metric.Size < m.MetricsIndex; i += 1 {
                    m.Data[i] = m.Data[i+metric.Size]
                }
            }
            m.Data = m.Data[:len(m.Data)-metric.Size]
        }
        for _, other := range m.GetMetrics() {
            if other.Index > metric.Index {
                other.Index -= metric.Size
            }
        }
        m.MetricsIndex -= metric.Size
    }
}

func (m *Matrix) RemoveInstance(key string) {
    if instance, ok := m.Instances[key]; ok {
        delete(m.Instances, key)
        if !m.IsEmpty() {
            for i := 0; i < m.MetricsIndex; i += 1 {
                for j := instance.Index; j < len(m.Instances) - 1; j += 1 {
                    m.Data[i][j] = m.Data[i][j+1]
                }
                m.Data[i] = m.Data[i][:len(m.Instances)]
            }
        }
        for _, other := range m.Instances {
            if other.Index > instance.Index {
                other.Index -= 1
            }
        }
    }
}

func (m *Matrix) RemoveLabel(key string) {
    m.Labels.Delete(key) // remove from instances as well?
}

func (m *Matrix) SizeMetrics() int {
    return len(m.Metrics) // or MetricsIndex??
}

func (m *Matrix) SizeLabels() int {
    return m.Labels.Size()
}

func (m *Matrix) SizeInstances() int {
    return len(m.Instances)
}

func (m *Matrix) ResetData() {
    m.Data = make([][]float32, 0)
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
    m.Labels = dict.New()
}


func (m *Matrix) GetInstances() map[string]*Instance {
    return m.Instances
}

func (m *Matrix) GetMetrics() map[string]*Metric {
    return m.Metrics
}

func (m *Matrix) GetLabels() map[string]string {
    return m.Labels.Iter()
}

func (m *Matrix) SetValueString(metric *Metric, instance *Instance, value string) error {
	var numeric float64
	var err error

    numeric, err = strconv.ParseFloat(value, 32)

    if err == nil {
		m.SetValue(metric, instance, float32(numeric))
    }
	return err
}

func (m *Matrix) SetValue(metric *Metric, instance *Instance, value float32) {
	m.Data[metric.Index][instance.Index] = value
}

func (m *Matrix) SetValueS(key string, instance *Instance, value float32) {
    if metric := m.GetMetric(key); metric != nil {
        m.SetValue(metric, instance, value)
    }
}

func (m *Matrix) SetValueSS(metric_key, instance_key string, value float32) {
    if instance := m.GetInstance(instance_key); instance != nil {
        m.SetValueS(metric_key, instance, value)
    }
}

func (m *Matrix) SetArrayValues(metric *Metric, instance *Instance, values []float32) {
    for i:=0; i<len(metric.Labels); i+=1 {
        m.Data[metric.Index+i][instance.Index] = values[i]        
    }
}

func (m *Matrix) SetArrayValuesString(metric *Metric, instance *Instance, values []string) error {
    var ok bool

    numeric := make([]float32, len(values))
    for _, v := range values {
        if n, err := strconv.ParseFloat(v, 32); err == nil {
            numeric = append(numeric, float32(n))
            ok = true // at least one parsed
        } else {
            numeric = append(numeric, NAN)
        }
    }

    m.SetArrayValues(metric, instance, numeric)

    if ok {
        return nil
    }
    return errors.New(errors.MATRIX_PARSE_STR, "no number parsed from: [" + strings.Join(values, ", ") + "]")
}

func (m *Matrix) GetValue(metric *Metric, instance *Instance) (float32, bool) {
	var value float32
	value = m.Data[metric.Index][instance.Index]
	return value, value==value
}

func (m *Matrix) GetArrayValues(metric *Metric, instance *Instance) []float32 {
    values := make([]float32, len(metric.Labels))
    for i:=0; i<len(metric.Labels); i+=1 {
        values[i] = m.Data[metric.Index+i][instance.Index]        
    }
    return values
}

// if name is empty, key will be used as name
func (m *Matrix) AddLabel(key, name string) {
    if name != "" {
        m.Labels.Set(key, name)
    } else {
        m.Labels.Set(key, key)
    }
}

func (m *Matrix) GetLabel(key string) (string, bool) {
    name, has := m.Labels.GetHas(key)
    if name == "" {
        return key, has
    }
    return name, has
}

func (m *Matrix) AddInstanceKey(key []string) {
    copied := make([]string, len(key))
    copy(copied, key)
    m.InstanceKeys = append(m.InstanceKeys, copied)
}

func (m *Matrix) GetInstanceKeys() [][]string {
    return m.InstanceKeys
}


func (m *Matrix) SetInstanceLabel(instance *Instance, key, value string) {
    display := m.Labels.Get(key)
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

func (m *Matrix) SetExportOptions(options *node.Node) {
    m.ExportOptions = options
}

func DefaultExportOptions() *node.Node {
    n := node.NewS("export_options")
    n.NewChildS("include_all_labels", "True")
    return n
}
