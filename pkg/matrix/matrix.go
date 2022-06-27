/*
 * Copyright NetApp Inc, 2021 All rights reserved

   Package matrix provides the Matrix data-structure and auxiliary structures
   for high performance storage, manipulation and transmission of numeric
   metrics and string labels. See detailed documentation in README.md

   See attached README for examples
*/
package matrix

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

type Matrix struct {
	UUID          string
	Object        string
	Identifier    string
	globalLabels  *dict.Dict
	instances     map[string]*Instance
	metrics       map[string]Metric
	exportOptions *node.Node
	exportable    bool
}

func New(uuid, object string, identifier string) *Matrix {
	me := Matrix{UUID: uuid, Object: object, Identifier: identifier}
	me.globalLabels = dict.New()
	me.instances = make(map[string]*Instance, 0)
	me.metrics = make(map[string]Metric, 0)
	me.exportable = true
	return &me
}

// only for debugging
func (me *Matrix) Print() {
	fmt.Println()
	fmt.Printf(">>> Metrics = %d\n", len(me.metrics))
	fmt.Printf(">>> Instances = %d\n", len(me.instances))
	fmt.Println()

	for key, metric := range me.GetMetrics() {
		fmt.Printf("(%s%s%s%s) (type=%s) (exportable=%v) values= ", color.Bold, color.Cyan, key, color.End, metric.GetType(), metric.IsExportable())
		metric.Print()
		fmt.Println()
	}
	fmt.Println()
}

// indicates wether this matrix is meant to be exported or not
// (some data is only collected to be aggregated by plugins)
func (me *Matrix) IsExportable() bool {
	return me.exportable
}

func (me *Matrix) SetExportable(b bool) {
	me.exportable = b
}

func (me *Matrix) Clone(withData, withMetrics, withInstances bool) *Matrix {
	clone := New(me.UUID, me.Object, me.Identifier)
	clone.globalLabels = me.globalLabels
	clone.exportOptions = me.exportOptions
	clone.exportable = me.exportable

	if withInstances {
		for key, instance := range me.GetInstances() {
			clone.instances[key] = instance.Clone()
		}
	}

	if withMetrics {
		for key, metric := range me.GetMetrics() {
			clone.metrics[key] = metric.Clone(withData)
		}
	}

	return clone
}

// flush all existing data
func (me *Matrix) Reset() {
	size := len(me.instances)
	for _, metric := range me.GetMetrics() {
		metric.Reset(size)
	}
}

func (me *Matrix) GetMetric(key string) Metric {
	if metric, has := me.metrics[key]; has {
		return metric
	}
	return nil
}

func (me *Matrix) GetMetrics() map[string]Metric {
	return me.metrics
}

func (me *Matrix) NewMetricInt(key string) (Metric, error) {
	metric := &MetricInt{AbstractMetric: &AbstractMetric{name: key, dtype: "int", exportable: true}}
	return metric, me.addMetric(key, metric)
}

func (me *Matrix) NewMetricInt32(key string) (Metric, error) {
	metric := &MetricInt32{AbstractMetric: &AbstractMetric{name: key, dtype: "int32", exportable: true}}
	return metric, me.addMetric(key, metric)
}

func (me *Matrix) NewMetricInt64(key string) (Metric, error) {
	metric := &MetricInt64{AbstractMetric: &AbstractMetric{name: key, dtype: "int64", exportable: true}}
	return metric, me.addMetric(key, metric)
}

func (me *Matrix) NewMetricUint8(key string) (Metric, error) {
	metric := &MetricUint8{AbstractMetric: &AbstractMetric{name: key, dtype: "uint8", exportable: true}}
	return metric, me.addMetric(key, metric)
}

func (me *Matrix) NewMetricUint32(key string) (Metric, error) {
	metric := &MetricUint32{AbstractMetric: &AbstractMetric{name: key, dtype: "uint32", exportable: true}}
	return metric, me.addMetric(key, metric)
}

func (me *Matrix) NewMetricUint64(key string) (Metric, error) {
	metric := &MetricUint64{AbstractMetric: &AbstractMetric{name: key, dtype: "uint64", exportable: true}}
	return metric, me.addMetric(key, metric)
}

func (me *Matrix) NewMetricFloat32(key string) (Metric, error) {
	metric := &MetricFloat32{AbstractMetric: &AbstractMetric{name: key, dtype: "float32", exportable: true}}
	return metric, me.addMetric(key, metric)
}

func (me *Matrix) NewMetricFloat64(key string) (Metric, error) {
	metric := &MetricFloat64{AbstractMetric: &AbstractMetric{name: key, dtype: "float64", exportable: true}}
	return metric, me.addMetric(key, metric)
}

func (me *Matrix) NewMetricType(key, dtype string) (Metric, error) {

	switch dtype {
	case "int":
		return me.NewMetricInt(key)
	case "int32":
		return me.NewMetricInt32(key)
	case "int64":
		return me.NewMetricInt64(key)
	case "uint8":
		return me.NewMetricUint8(key)
	case "uint32":
		return me.NewMetricUint32(key)
	case "uint64":
		return me.NewMetricUint64(key)
	case "float32":
		return me.NewMetricFloat32(key)
	case "float64":
		return me.NewMetricFloat64(key)
	default:
		return nil, errs.New(ErrInvalidDtype, dtype)
	}
}

func (me *Matrix) ChangeMetricType(key, dtype string) (Metric, error) {
	me.RemoveMetric(key)
	return me.NewMetricType(key, dtype)
}

func (me *Matrix) addMetric(key string, metric Metric) error {
	if _, has := me.metrics[key]; has {
		return errs.New(ErrDuplicateMetricKey, key)
	}
	metric.Reset(len(me.instances))
	me.metrics[key] = metric
	return nil
}

func (me *Matrix) RemoveMetric(key string) {
	delete(me.metrics, key)
}

func (me *Matrix) GetInstance(key string) *Instance {
	if instance, has := me.instances[key]; has {
		return instance
	}
	return nil
}

func (me *Matrix) GetInstances() map[string]*Instance {
	return me.instances
}

func (me *Matrix) PurgeInstances() {
	me.instances = make(map[string]*Instance)
}

func (me *Matrix) GetInstanceKeys() []string {
	keys := make([]string, 0, len(me.instances))
	for k := range me.instances {
		keys = append(keys, k)
	}
	return keys
}

func (me *Matrix) NewInstance(key string) (*Instance, error) {

	var instance *Instance

	if _, has := me.instances[key]; has {
		return nil, errs.New(ErrDuplicateInstanceKey, key)
	}

	instance = NewInstance(len(me.instances)) // index is current count of instances

	for _, metric := range me.GetMetrics() {
		metric.Append()
	}

	me.instances[key] = instance
	return instance, nil
}

func (me *Matrix) ResetInstance(key string) {
	if instance, has := me.instances[key]; has {
		for _, metric := range me.GetMetrics() {
			metric.SetValueNAN(instance)
		}
	}
}

func (me *Matrix) RemoveInstance(key string) {
	if instance, has := me.instances[key]; has {
		// re-arrange columns in metrics
		for _, metric := range me.GetMetrics() {
			metric.Remove(instance.index)
		}
		deletedIndex := instance.index
		delete(me.instances, key)
		// If there were removals, the indexes need to be rewritten since gaps were created
		// Map is not ordered hence recreating map will cause mapping issue with metrics
		for _, i := range me.instances {
			if i.index > deletedIndex {
				// reduce index by 1
				i.index = i.index - 1
			}
		}
	}
}

func (me *Matrix) SetGlobalLabel(label, value string) {
	me.globalLabels.Set(label, value)
}

// Set all global labels if already not exist
func (me *Matrix) SetGlobalLabels(allLabels *dict.Dict) {
	me.globalLabels.SetAll(allLabels)
}

func (me *Matrix) GetGlobalLabels() *dict.Dict {
	return me.globalLabels
}

func (me *Matrix) GetExportOptions() *node.Node {
	if me.exportOptions != nil {
		return me.exportOptions
	}
	return DefaultExportOptions()
}

func (me *Matrix) SetExportOptions(e *node.Node) {
	me.exportOptions = e
}

func DefaultExportOptions() *node.Node {
	n := node.NewS("export_options")
	n.NewChildS("include_all_labels", "true")
	return n
}
