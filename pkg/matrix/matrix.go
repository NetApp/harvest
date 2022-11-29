/*
* Copyright NetApp Inc, 2021 All rights reserved

	Package matrix provides the Matrix data-structure and auxiliary structures
	for high performance storage, manipulation and transmission of numeric
	metrics and string labels.

	See attached README for examples
*/

package matrix

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strings"
)

type Matrix struct {
	UUID           string
	Object         string
	Identifier     string
	globalLabels   *dict.Dict
	instances      map[string]*Instance
	metrics        map[string]Metric // ONTAP metric name => metric (in templates, this is left side)
	displayMetrics map[string]Metric // display name of metric to => metric (in templates, this is right side)
	exportOptions  *node.Node
	exportable     bool
}

func New(uuid, object string, identifier string) *Matrix {
	me := Matrix{UUID: uuid, Object: object, Identifier: identifier}
	me.globalLabels = dict.New()
	me.instances = make(map[string]*Instance, 0)
	me.metrics = make(map[string]Metric, 0)
	me.displayMetrics = make(map[string]Metric, 0)
	me.exportable = true
	return &me
}

// Print is only for debugging
func (m *Matrix) Print() {
	fmt.Println()
	fmt.Printf(">>> Metrics = %d\n", len(m.metrics))
	fmt.Printf(">>> Instances = %d\n", len(m.instances))
	fmt.Println()

	for key, metric := range m.GetMetrics() {
		fmt.Printf("(%s%s%s%s) (type=%s) (exportable=%v) values= ", color.Bold, color.Cyan, key, color.End, metric.GetType(), metric.IsExportable())
		metric.Print()
		fmt.Println()
	}
	fmt.Println()
}

// IsExportable indicates whether this matrix is meant to be exported or not
// (some data is only collected to be aggregated by plugins)
func (m *Matrix) IsExportable() bool {
	return m.exportable
}

func (m *Matrix) SetExportable(b bool) {
	m.exportable = b
}

func (m *Matrix) Clone(withData, withMetrics, withInstances bool) *Matrix {
	clone := New(m.UUID, m.Object, m.Identifier)
	clone.globalLabels = m.globalLabels
	clone.exportOptions = m.exportOptions
	clone.exportable = m.exportable

	if withInstances {
		for key, instance := range m.GetInstances() {
			clone.instances[key] = instance.Clone()
		}
	}

	if withMetrics {
		for key, metric := range m.GetMetrics() {
			c := metric.Clone(withData)
			clone.metrics[key] = c
			clone.displayMetrics[c.GetName()] = c
		}
	}

	return clone
}

// Reset all data
func (m *Matrix) Reset() {
	size := len(m.instances)
	for _, metric := range m.GetMetrics() {
		metric.Reset(size)
	}
}

func (m *Matrix) DisplayMetric(name string) Metric {
	if metric, has := m.displayMetrics[name]; has {
		return metric
	}
	return nil
}

func (m *Matrix) GetMetric(key string) Metric {
	if metric, has := m.metrics[key]; has {
		return metric
	}
	return nil
}

func (m *Matrix) GetMetrics() map[string]Metric {
	return m.metrics
}

func (m *Matrix) NewMetricInt64(key string, display ...string) (Metric, error) {
	metric := &MetricInt64{AbstractMetric: newAbstract(key, "int64", display...)}
	return metric, m.addMetric(key, metric)
}

func (m *Matrix) NewMetricUint8(key string, display ...string) (Metric, error) {
	metric := &MetricUint8{AbstractMetric: newAbstract(key, "uint8", display...)}
	return metric, m.addMetric(key, metric)
}

func (m *Matrix) NewMetricUint64(key string, display ...string) (Metric, error) {
	metric := &MetricUint64{AbstractMetric: newAbstract(key, "uint64", display...)}
	return metric, m.addMetric(key, metric)
}

func (m *Matrix) NewMetricFloat64(key string, display ...string) (Metric, error) {
	metric := &MetricFloat64{AbstractMetric: newAbstract(key, "float64", display...)}
	return metric, m.addMetric(key, metric)
}

func (m *Matrix) NewMetricType(key string, dataType string, display ...string) (Metric, error) {

	switch dataType {
	case "int64":
		return m.NewMetricInt64(key, display...)
	case "uint8":
		return m.NewMetricUint8(key, display...)
	case "uint64":
		return m.NewMetricUint64(key, display...)
	case "float64":
		return m.NewMetricFloat64(key, display...)
	default:
		return nil, errs.New(ErrInvalidDtype, dataType)
	}
}

func newAbstract(key string, dataType string, display ...string) *AbstractMetric {
	name := key
	if len(display) > 0 {
		name = display[0]
	}
	return &AbstractMetric{name: name, dtype: dataType, exportable: true}
}

func (m *Matrix) addMetric(key string, metric Metric) error {
	if _, has := m.metrics[key]; has { // Fail if a metric with the same key already exists
		return errs.New(ErrDuplicateMetricKey, key)
	}
	metric.Reset(len(m.instances))
	m.metrics[key] = metric
	m.displayMetrics[metric.GetName()] = metric
	return nil
}

func (m *Matrix) RemoveMetric(key string) {
	delete(m.metrics, key)
}

func (m *Matrix) RemoveExceptMetric(key string) {
	prev, ok := m.metrics[key]
	if !ok {
		return
	}
	m.metrics = make(map[string]Metric)
	m.displayMetrics = make(map[string]Metric)
	_ = m.addMetric(key, prev)
}

func (m *Matrix) GetInstance(key string) *Instance {
	if instance, has := m.instances[key]; has {
		return instance
	}
	return nil
}

func (m *Matrix) GetInstancesBySuffix(subKey string) []*Instance {
	var instances []*Instance
	if subKey != "" {
		for key, instance := range m.instances {
			if strings.HasSuffix(key, subKey) {
				instances = append(instances, instance)
			}
		}
	}
	return instances
}

func (m *Matrix) GetInstances() map[string]*Instance {
	return m.instances
}

func (m *Matrix) PurgeInstances() {
	m.instances = make(map[string]*Instance)
}

func (m *Matrix) GetInstanceKeys() []string {
	keys := make([]string, 0, len(m.instances))
	for k := range m.instances {
		keys = append(keys, k)
	}
	return keys
}

func (m *Matrix) NewInstance(key string) (*Instance, error) {

	var instance *Instance

	if _, has := m.instances[key]; has {
		return nil, errs.New(ErrDuplicateInstanceKey, key)
	}

	instance = NewInstance(len(m.instances)) // index is current count of instances

	for _, metric := range m.GetMetrics() {
		metric.Append()
	}

	m.instances[key] = instance
	return instance, nil
}

func (m *Matrix) ResetInstance(key string) {
	if instance, has := m.instances[key]; has {
		for _, metric := range m.GetMetrics() {
			metric.SetValueNAN(instance)
		}
	}
}

func (m *Matrix) RemoveInstance(key string) {
	if instance, has := m.instances[key]; has {
		// re-arrange columns in metrics
		for _, metric := range m.GetMetrics() {
			metric.Remove(instance.index)
		}
		deletedIndex := instance.index
		delete(m.instances, key)
		// If there were removals, the indexes need to be rewritten since gaps were created
		// Map is not ordered hence recreating map will cause mapping issue with metrics
		for _, i := range m.instances {
			if i.index > deletedIndex {
				// reduce index by 1
				i.index = i.index - 1
			}
		}
	}
}

func (m *Matrix) SetGlobalLabel(label, value string) {
	m.globalLabels.Set(label, value)
}

// SetGlobalLabels sets all global labels that do not already exist
func (m *Matrix) SetGlobalLabels(allLabels *dict.Dict) {
	m.globalLabels.SetAll(allLabels)
}

func (m *Matrix) GetGlobalLabels() *dict.Dict {
	return m.globalLabels
}

func (m *Matrix) GetExportOptions() *node.Node {
	if m.exportOptions != nil {
		return m.exportOptions
	}
	return DefaultExportOptions()
}

func (m *Matrix) SetExportOptions(e *node.Node) {
	m.exportOptions = e
}

func DefaultExportOptions() *node.Node {
	n := node.NewS("export_options")
	n.NewChildS("include_all_labels", "true")
	return n
}

func CreateMetric(key string, data *Matrix) error {
	var err error
	at := data.GetMetric(key)
	if at == nil {
		if _, err = data.NewMetricFloat64(key); err != nil {
			return err
		}
	}
	return nil
}
