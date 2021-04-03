package matrix

import (
	"fmt"
	"goharvest2/share/dict"
	"goharvest2/share/util"
	"goharvest2/share/errors"
	"goharvest2/share/tree/node"
)

type Matrix struct {
	Collector      string
	Object         string
	Plugin         string
	GlobalLabels   *dict.Dict
	Instances      map[string]*Instance
	Metrics        map[string]Metric
	ExportOptions  *node.Node
	IsMetadata     bool
	MetadataType   string
	MetadataObject string
	Exportable bool
	Empty bool
}

func New(collector, object, plugin string) *Matrix {
	m := Matrix{Collector: collector, Object: object, Plugin: plugin}
	m.GlobalLabels = dict.New()
	m.Instances = make(map[string]*Instance, 0)
	m.Metrics = make(map[string]Metric, 0)
	m.Exportable = true
	m.Empty = true
	return &m
}

func (m *Matrix) Print() {
	fmt.Println()
	fmt.Printf(">>> Metrics = %d\n", m.SizeMetrics())
	fmt.Printf(">>> Instances = %d\n", m.SizeInstances())
	fmt.Println()

	for key, metric := range m.GetMetrics() {
		fmt.Printf("(%s%s%s%s) (type=%s) (exportable=%v) values= ", util.Bold, util.Cyan, key, util.End, metric.GetType(), metric.IsExportable())
		metric.Print()
		fmt.Println()
	}
	fmt.Println()
}

func (m *Matrix) IsEmpty() bool {
	return m.Empty
}

func (m *Matrix) Reset() error {
	if m.SizeMetrics() == 0 || m.SizeInstances() == 0 {
		return errors.New(errors.MATRIX_EMPTY, "counter or instance cache empty")
	}
	size := m.SizeInstances()
	for _, metric := range m.Metrics {
		metric.Reset(size)
	}
	return nil
}

func (m *Matrix) SizeMetrics() int {
	return len(m.Metrics)
}

func (m *Matrix) GetMetric(key string) Metric {
	if metric, has := m.Metrics[key]; has {
		return metric
	}
	return nil
}

func (m *Matrix) GetMetrics() map[string]Metric {
	return m.Metrics
}

func (m *Matrix) AddMetricInt(key string) (Metric, error) {
	metric := &MetricInt{AbstractMetric: &AbstractMetric{Name: key, Type: "int", Exportable: true}}
	return metric, m.AddMetric(key, metric)
}

func (m *Matrix) AddMetricInt32(key string) (Metric, error) {
	metric := &MetricInt32{AbstractMetric: &AbstractMetric{Name: key, Type: "int32", Exportable: true}}
	return metric, m.AddMetric(key, metric)
}

func (m *Matrix) AddMetricInt64(key string) (Metric, error) {
	metric := &MetricInt64{AbstractMetric: &AbstractMetric{Name: key, Type: "int64", Exportable: true}}
	return metric, m.AddMetric(key, metric)
}

func (m *Matrix) AddMetricUint32(key string) (Metric, error) {
	metric := &MetricUint32{AbstractMetric: &AbstractMetric{Name: key, Type: "uint32", Exportable: true}}
	return metric, m.AddMetric(key, metric)
}

func (m *Matrix) AddMetricUint64(key string) (Metric, error) {
	metric := &MetricUint64{AbstractMetric: &AbstractMetric{Name: key, Type: "uint64", Exportable: true}}
	return metric, m.AddMetric(key, metric)
}

func (m *Matrix) AddMetricFloat32(key string) (Metric, error) {
	metric := &MetricFloat32{AbstractMetric: &AbstractMetric{Name: key, Type: "float32", Exportable: true}}
	return metric, m.AddMetric(key, metric)
}

func (m *Matrix) AddMetricFloat64(key string) (Metric, error) {
	metric := &MetricFloat64{AbstractMetric: &AbstractMetric{Name: key, Type: "float64", Exportable: true}}
	return metric, m.AddMetric(key, metric)
}

func (m *Matrix) AddMetric(key string, metric Metric) (error) {

	if _, has := m.Metrics[key]; has {
		return errors.New(DUPLICATE_METRIC_KEY, key)
	}

	if ! m.IsEmpty() {
		metric.Reset(m.SizeInstances())
	}

	m.Metrics[key] = metric
	
	return nil
}

func (m *Matrix) RemoveMetric(key string) {
	delete(m.Metrics, key)
}

func (m *Matrix) SizeInstances() int {
	return len(m.Instances)
}

func (m *Matrix) GetInstance(key string) *Instance {
	if instance, has := m.Instances[key]; has {
		return instance
	}
	return nil
}

func (m *Matrix) GetInstances() map[string]*Instance {
	return m.Instances
}

func (m *Matrix) GetInstanceKeys() []string {
	keys := make([]string, 0, len(m.Instances))
	for k := range m.Instances {
		keys = append(keys, k)
	}
	return keys
}

func (m *Matrix) AddInstance(key string) (*Instance, error) {

	var instance *Instance

	if _, has := m.Instances[key]; has {
		return nil, errors.New(DUPLICATE_INSTANCE_KEY, key)
	}

	instance = NewInstance(len(m.Instances)) // index is current count of instances

	if ! m.IsEmpty() {
		for _, mt := range m.Metrics {
			mt.Append()
		}
	}

	m.Instances[key] = instance
	return instance, nil
}

func (m *Matrix) RemoveInstance(key string) {
	if instance, has := m.Instances[key]; has {

		if ! m.IsEmpty() {
			// re-arrange columns in metrics
			for _, metric := range m.Metrics {

				metric.Remove(instance.index)
			}
		}
		delete(m.Instances, key)
	}
}


func (m *Matrix) SetGlobalLabel(label, value string) {
	m.GlobalLabels.Set(label, value)
}

func (m *Matrix) GetGlobalLabels() map[string]string {
	return m.GlobalLabels.Iter()
}

func (m *Matrix) SetExportOptions(options *node.Node) {
	m.ExportOptions = options
}

func DefaultExportOptions() *node.Node {
	n := node.NewS("export_options")
	n.NewChildS("include_all_labels", "True")
	return n
}
