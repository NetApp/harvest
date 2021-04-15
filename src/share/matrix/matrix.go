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
	me := Matrix{Collector: collector, Object: object, Plugin: plugin}
	me.GlobalLabels = dict.New()
	me.Instances = make(map[string]*Instance, 0)
	me.Metrics = make(map[string]Metric, 0)
	me.Exportable = true
	me.Empty = true
	return &me
}

func (me *Matrix) Print() {
	fmt.Println()
	fmt.Printf(">>> Metrics = %d\n", me.SizeMetrics())
	fmt.Printf(">>> Instances = %d\n", me.SizeInstances())
	fmt.Println()

	for key, metric := range me.GetMetrics() {
		fmt.Printf("(%s%s%s%s) (type=%s) (exportable=%v) values= ", util.Bold, util.Cyan, key, util.End, metric.GetType(), metric.IsExportable())
		metric.Print()
		fmt.Println()
	}
	fmt.Println()
}

func (me *Matrix) IsEmpty() bool {
	return me.Empty
}

func (me *Matrix) IsExportable() bool {
	return me.Exportable
}

func (me *Matrix) SetExportable(b bool) {
	me.Exportable = b
}

func (me *Matrix) Clone(with_data, with_metrics, with_instances bool) *Matrix {
	clone := New(me.Collector, me.Object, me.Plugin)
	clone.GlobalLabels = me.GlobalLabels
	clone.ExportOptions = me.ExportOptions
	clone.IsMetadata = me.IsMetadata
	clone.MetadataType = me.MetadataType
	clone.MetadataObject = me.MetadataObject
	clone.Exportable = me.Exportable

	if with_instances {
		for key, instance := range me.GetInstances() {
			clone.Instances[key] = instance.Clone()
		}
	}

	if with_metrics {
		for key, metric := range me.GetMetrics() {
			clone.Metrics[key] = metric.Clone(with_data)
		}
	}

	if with_data && len(clone.Metrics) != 0 && len(clone.Instances) != 0 {
		clone.Empty = false
	} 
	return clone
}

func (me *Matrix) Reset() error {
	if me.SizeMetrics() == 0 && me.SizeInstances() == 0 {
		return errors.New(errors.MATRIX_EMPTY, "counter and instance cache empty")
	}
	size := me.SizeInstances()
	for _, metric := range me.GetMetrics() {
		metric.Reset(size)
	}
	me.Empty = false
	return nil
}

func (me *Matrix) SizeMetrics() int {
	return len(me.Metrics)
}

func (me *Matrix) GetMetric(key string) Metric {
	if metric, has := me.Metrics[key]; has {
		return metric
	}
	return nil
}

func (me *Matrix) GetMetrics() map[string]Metric {
	return me.Metrics
}

func (me *Matrix) AddMetricInt(key string) (Metric, error) {
	metric := &MetricInt{AbstractMetric: &AbstractMetric{name: key, dtype: "int", exportable: true}}
	return metric, me.AddMetric(key, metric)
}

func (me *Matrix) AddMetricInt32(key string) (Metric, error) {
	metric := &MetricInt32{AbstractMetric: &AbstractMetric{name: key, dtype: "int32", exportable: true}}
	return metric, me.AddMetric(key, metric)
}

func (me *Matrix) AddMetricInt64(key string) (Metric, error) {
	metric := &MetricInt64{AbstractMetric: &AbstractMetric{name: key, dtype: "int64", exportable: true}}
	return metric, me.AddMetric(key, metric)
}

func (me *Matrix) AddMetricUint8(key string) (Metric, error) {
	metric := &MetricUint8{AbstractMetric: &AbstractMetric{name: key, dtype: "uint8", exportable: true}}
	return metric, me.AddMetric(key, metric)
}

func (me *Matrix) AddMetricUint32(key string) (Metric, error) {
	metric := &MetricUint32{AbstractMetric: &AbstractMetric{name: key, dtype: "uint32", exportable: true}}
	return metric, me.AddMetric(key, metric)
}

func (me *Matrix) AddMetricUint64(key string) (Metric, error) {
	metric := &MetricUint64{AbstractMetric: &AbstractMetric{name: key, dtype: "uint64", exportable: true}}
	return metric, me.AddMetric(key, metric)
}

func (me *Matrix) AddMetricFloat32(key string) (Metric, error) {
	metric := &MetricFloat32{AbstractMetric: &AbstractMetric{name: key, dtype: "float32", exportable: true}}
	return metric, me.AddMetric(key, metric)
}

func (me *Matrix) AddMetricFloat64(key string) (Metric, error) {
	metric := &MetricFloat64{AbstractMetric: &AbstractMetric{name: key, dtype: "float64", exportable: true}}
	return metric, me.AddMetric(key, metric)
}

func (me *Matrix) AddMetricType(key, dtype string) (Metric, error) {

	switch dtype {
	case "int":
		return me.AddMetricInt(key)
	case "int32":
		return me.AddMetricInt32(key)
	case "int64":
		return me.AddMetricInt64(key)
	case "uint8":
		return me.AddMetricUint8(key)
	case "uint32":
		return me.AddMetricUint32(key)
	case "uint64":
		return me.AddMetricUint64(key)
	case "float32":
		return me.AddMetricFloat32(key)
	case "float64":
		return me.AddMetricFloat64(key)
	default:
		return nil, errors.New(INVALID_DTYPE, dtype)
	}
}

func (me *Matrix) ChangeMetricType(key, dtype string) (Metric, error) {
	me.RemoveMetric(key)
	return me.AddMetricType(key, dtype)
}

func (me *Matrix) AddMetric(key string, metric Metric) (error) {

	if _, has := me.Metrics[key]; has {
		return errors.New(DUPLICATE_METRIC_KEY, key)
	}

	if ! me.IsEmpty() {
		metric.Reset(me.SizeInstances())
	}

	me.Metrics[key] = metric
	
	return nil
}

func (me *Matrix) RemoveMetric(key string) {
	delete(me.Metrics, key)
}

func (me *Matrix) SizeInstances() int {
	return len(me.Instances)
}

func (me *Matrix) GetInstance(key string) *Instance {
	if instance, has := me.Instances[key]; has {
		return instance
	}
	return nil
}

func (me *Matrix) GetInstances() map[string]*Instance {
	return me.Instances
}

func (me *Matrix) PurgeInstances() {
	me.Instances = make(map[string]*Instance)
}

func (me *Matrix) GetInstanceKeys() []string {
	keys := make([]string, 0, len(me.Instances))
	for k := range me.Instances {
		keys = append(keys, k)
	}
	return keys
}

func (me *Matrix) AddInstance(key string) (*Instance, error) {

	var instance *Instance

	if _, has := me.Instances[key]; has {
		return nil, errors.New(DUPLICATE_INSTANCE_KEY, key)
	}

	instance = NewInstance(len(me.Instances)) // index is current count of instances

	if ! me.IsEmpty() {
		for _, metric := range me.GetMetrics() {
			metric.Append()
		}
	}

	me.Instances[key] = instance
	return instance, nil
}

func (me *Matrix) RemoveInstance(key string) {
	if instance, has := me.Instances[key]; has {

		if ! me.IsEmpty() {
			// re-arrange columns in metrics
			for _, metric := range me.GetMetrics() {

				metric.Remove(instance.index)
			}
		}
		delete(me.Instances, key)
	}
}

func (me *Matrix) SetGlobalLabel(label, value string) {
	me.GlobalLabels.Set(label, value)
}

func (me *Matrix) GetGlobalLabels() map[string]string {
	return me.GlobalLabels.Iter()
}

func (me *Matrix) SetExportOptions(options *node.Node) {
	me.ExportOptions = options
}

func DefaultExportOptions() *node.Node {
	n := node.NewS("export_options")
	n.NewChildS("include_all_labels", "True")
	return n
}
