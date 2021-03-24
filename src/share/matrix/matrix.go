package matrix

import (
	"goharvest2/share/dict"
	"goharvest2/share/errors"
	"goharvest2/share/tree/node"
	"math"
	"strconv"
)

var NAN = float64(math.NaN())

type Matrix struct {
	Collector      string
	Object         string
	Plugin         string
	GlobalLabels   *dict.Dict
	Labels         *dict.Dict
	Instances      map[string]*Instance
	Metrics        map[string]*Metric
	InstanceKeys   [][]string // @TODO deprecate, doesn't belong here
	ExportOptions  *node.Node
	Data           [][]float64
	IsMetadata     bool
	MetadataType   string
	MetadataObject string
    Exportable bool
}

func New(collector, object, plugin string) *Matrix {
	m := Matrix{Collector: collector, Object: object, Plugin: plugin}
	m.GlobalLabels = dict.New()
	m.Labels = dict.New()
	m.Instances = map[string]*Instance{}
	m.Metrics = map[string]*Metric{}
    m.Exportable = true
	return &m
}

func (m *Matrix) IsEmpty() bool {
	return len(m.Data) == 0
}

func (m *Matrix) Clone(copy_data bool) *Matrix {
	n := &Matrix{
		Collector:      m.Collector,
		Object:         m.Object,
		Plugin:         m.Plugin,
		Instances:      m.Instances,
		InstanceKeys:   m.InstanceKeys,
		Metrics:        m.Metrics,
		GlobalLabels:   m.GlobalLabels,
		Labels:         m.Labels,
		ExportOptions:  m.ExportOptions,
		IsMetadata:     m.IsMetadata,
		MetadataType:   m.MetadataType,
		MetadataObject: m.MetadataObject,
        Exportable:     m.Exportable,
	}
	if copy_data && !m.IsEmpty() {
		n.Data = make([][]float64, n.SizeMetrics())
		for i := 0; i < n.SizeMetrics(); i += 1 {
			n.Data[i] = make([]float64, n.SizeInstances())
			copy(n.Data[i], m.Data[i])
		}
	}
	return n
}

func (m *Matrix) InitData() error {
	if m.SizeMetrics() == 0 || m.SizeInstances() == 0 {
		return errors.New(errors.MATRIX_EMPTY, "counter or instance cache empty")
	}
	m.Data = make([][]float64, m.SizeMetrics())
	for i := 0; i < m.SizeMetrics(); i += 1 {
		m.Data[i] = make([]float64, m.SizeInstances())
		for j := 0; j < m.SizeInstances(); j += 1 {
			m.Data[i][j] = NAN
		}
	}
	return nil
}

func (m *Matrix) RemoveMetric(key string) {
	if metric, ok := m.Metrics[key]; ok {
		delete(m.Metrics, key)
		if !m.IsEmpty() {
			// re-arrange rows, if metric is not last/only row
			if len(m.Data) > metric.Index+1 {
				for i := metric.Index; i < m.SizeMetrics(); i += 1 {
					m.Data[i] = m.Data[i+1]
				}
			}
			m.Data = m.Data[:len(m.Data)-1]
		}
		// re-assign indices to other metrics
		for _, other := range m.GetMetrics() {
			if other.Index > metric.Index {
				other.Index -= 1
			}
		}
	}
}

func (m *Matrix) RemoveInstance(key string) {
	if instance, ok := m.Instances[key]; ok {
		delete(m.Instances, key)
		if !m.IsEmpty() {
			// re-arrange columns
			for i := 0; i < m.SizeMetrics(); i += 1 {
				for j := instance.Index; j < m.SizeInstances(); j += 1 {
					m.Data[i][j] = m.Data[i][j+1]
				}
				m.Data[i] = m.Data[i][:m.SizeInstances()]
			}
		}
		// re-assign indices to other instances
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
	return len(m.Metrics)
}

func (m *Matrix) SizeLabels() int {
	return m.Labels.Size()
}

func (m *Matrix) SizeInstances() int {
	return len(m.Instances)
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
		m.SetValue(metric, instance, float64(numeric))
	}
	return err
}

func (m *Matrix) IncrementValue(metric *Metric, instance *Instance, value float64) {
    if _, ok := m.GetValue(metric, instance); ok {
        m.Data[metric.Index][instance.Index] += value
    } else {
        m.Data[metric.Index][instance.Index] = value
    }
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

func (m *Matrix) GetValue(metric *Metric, instance *Instance) (float64, bool) {
	var value float64
        // temporary fix plugin bug
	if m.Data == nil || len(m.Data) == 0 || metric.Index >= len(m.Data) || instance.Index >= len(m.Data[metric.Index]) {
		return value, false
	}
	value = m.Data[metric.Index][instance.Index]
	return value, value == value
}

func (m *Matrix) GetValueS(key string, instance *Instance) (float64, bool) {
	if metric := m.GetMetric(key); metric != nil {
		return m.GetValue(metric, instance)
	}
	return NAN, false
}

func (m *Matrix) GetValueSS(M, I string) (float64, bool) {
	if metric := m.GetMetric(M); metric != nil {
		if instance := m.GetInstance(I); instance != nil {
			return m.GetValue(metric, instance)
		}
	}
	return NAN, false
}

// if name is empty, key will be used as display name
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
