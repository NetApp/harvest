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
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strings"
)

type Matrix struct {
	UUID           string
	Object         string
	Identifier     string
	globalLabels   *dict.Dict
	instances      map[string]*Instance
	metrics        map[string]*Metric // ONTAP metric name => metric (in templates, this is left side)
	displayMetrics map[string]string  // display name of metric to => metric name (in templates, this is right side)
	exportOptions  *node.Node
	exportable     bool
}

func New(uuid, object string, identifier string) *Matrix {
	me := Matrix{UUID: uuid, Object: object, Identifier: identifier}
	me.globalLabels = dict.New()
	me.instances = make(map[string]*Instance, 0)
	me.metrics = make(map[string]*Metric, 0)
	me.displayMetrics = make(map[string]string, 0)
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
		fmt.Printf("(%s) (type=%s) (exportable=%v) values= ", key, metric.GetType(), metric.IsExportable())
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
			clone.displayMetrics[c.GetName()] = key
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

func (m *Matrix) DisplayMetric(name string) *Metric {
	if metricKey, has := m.displayMetrics[name]; has {
		return m.GetMetric(metricKey)
	}
	return nil
}

func (m *Matrix) GetMetric(key string) *Metric {
	if metric, has := m.metrics[key]; has {
		return metric
	}
	return nil
}

func (m *Matrix) GetMetrics() map[string]*Metric {
	return m.metrics
}

func (m *Matrix) NewMetricInt64(key string, display ...string) (*Metric, error) {
	metric := newAbstract(key, "int64", display...)
	return metric, m.addMetric(key, metric)
}

func (m *Matrix) NewMetricUint8(key string, display ...string) (*Metric, error) {
	metric := newAbstract(key, "uint8", display...)
	return metric, m.addMetric(key, metric)
}

func (m *Matrix) NewMetricUint64(key string, display ...string) (*Metric, error) {
	metric := newAbstract(key, "uint64", display...)
	return metric, m.addMetric(key, metric)
}

func (m *Matrix) NewMetricFloat64(key string, display ...string) (*Metric, error) {
	metric := newAbstract(key, "float64", display...)
	return metric, m.addMetric(key, metric)
}

func (m *Matrix) NewMetricType(key string, dataType string, display ...string) (*Metric, error) {

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

func newAbstract(key string, dataType string, display ...string) *Metric {
	name := key
	if len(display) > 0 && display[0] != "" {
		name = display[0]
	}
	return &Metric{name: name, dataType: dataType, exportable: true}
}

func (m *Matrix) addMetric(key string, metric *Metric) error {
	if _, has := m.metrics[key]; has { // Fail if a metric with the same key already exists
		return errs.New(ErrDuplicateMetricKey, key)
	}
	// Histograms and arrays don't support display metrics yet, last write wins
	metric.Reset(len(m.instances))
	m.metrics[key] = metric
	m.displayMetrics[metric.GetName()] = key
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
	m.metrics = make(map[string]*Metric)
	m.displayMetrics = make(map[string]string)
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

// Delta vector arithmetics
func (m *Matrix) Delta(metricKey string, prevMat *Matrix, logger *logging.Logger) (int, error) {
	var skips int
	prevMetric := prevMat.GetMetric(metricKey)
	curMetric := m.GetMetric(metricKey)
	prevRaw := prevMetric.values
	prevRecord := prevMetric.GetRecords()
	for key, currInstance := range m.GetInstances() {
		// check if this instance key exists in previous matrix
		prevInstance := prevMat.GetInstance(key)
		currIndex := currInstance.index
		curRaw := curMetric.values[currIndex]
		if prevInstance != nil {
			prevIndex := prevInstance.index
			if curMetric.record[currIndex] && prevRecord[prevIndex] {
				curMetric.values[currIndex] -= prevRaw[prevIndex]
				// Sometimes ONTAP sends spurious zeroes or values less than the previous poll.
				// Detect and don't publish negative deltas or the subsequent poll will show a large spike.
				isInvalidZero := (curRaw == 0 || prevRaw[prevIndex] == 0) && curMetric.values[prevIndex] != 0
				isNegative := curMetric.values[currIndex] < 0
				if isInvalidZero || isNegative {
					curMetric.record[currIndex] = false
					skips++
					logger.Trace().
						Str("metric", curMetric.GetName()).
						Str("key", key).
						Float64("currentRaw", curRaw).
						Float64("previousRaw", prevRaw[prevIndex]).
						Str("instKey", key).
						Msg("Negative cooked value")
				}
			} else {
				curMetric.record[currIndex] = false
				skips++
				logger.Trace().
					Str("metric", curMetric.GetName()).
					Str("key", key).
					Float64("currentRaw", curRaw).
					Float64("previousRaw", prevRaw[prevIndex]).
					Str("instKey", key).
					Msg("Delta calculation skipped")
			}
		} else {
			curMetric.record[currIndex] = false
			skips++
			logger.Trace().
				Str("metric", curMetric.GetName()).
				Str("key", key).
				Float64("currentRaw", curRaw).
				Str("instKey", key).
				Msg("New instance added")
		}
	}
	return skips, nil
}

func (m *Matrix) Divide(metricKey string, baseKey string, logger *logging.Logger) (int, error) {
	var skips int
	metric := m.GetMetric(metricKey)
	base := m.GetMetric(baseKey)
	sValues := base.values
	sRecord := base.GetRecords()
	if len(metric.values) != len(sValues) {
		return 0, errs.New(ErrUnequalVectors, fmt.Sprintf("numerator=%d, denominator=%d", len(metric.values), len(sValues)))
	}
	for i := 0; i < len(metric.values); i++ {
		if metric.record[i] && sRecord[i] {
			if sValues[i] != 0 {
				// Don't pass along the value if the numerator or denominator is < 0
				// A denominator of zero is fine
				if metric.values[i] < 0 || sValues[i] < 0 {
					metric.record[i] = false
					skips++
					logger.Trace().
						Str("metric", metric.GetName()).
						Str("key", metricKey).
						Float64("numerator", metric.values[i]).
						Float64("denominator", sValues[i]).
						Msg("Divide calculation skipped")
				}
				metric.values[i] /= sValues[i]
			} else {
				metric.values[i] = 0
			}
		} else {
			metric.record[i] = false
			skips++
			logger.Trace().
				Str("metric", metric.GetName()).
				Str("key", metricKey).
				Float64("numerator", metric.values[i]).
				Float64("denominator", sValues[i]).
				Msg("Divide calculation skipped")
		}
	}
	return skips, nil
}

func (m *Matrix) DivideWithThreshold(metricKey string, baseKey string, threshold int, logger *logging.Logger) (int, error) {
	var skips int
	x := float64(threshold)
	metric := m.GetMetric(metricKey)
	base := m.GetMetric(baseKey)
	sValues := base.values
	sRecord := base.GetRecords()
	if len(metric.values) != len(sValues) {
		return 0, errs.New(ErrUnequalVectors, fmt.Sprintf("numerator=%d, denominator=%d", len(metric.values), len(sValues)))
	}
	for i := 0; i < len(metric.values); i++ {
		v := metric.values[i]
		// Don't pass along the value if the numerator or denominator is < 0
		// It is important to check sValues[i] < 0 and allow a zero so pass=true and m.values[i] remains unchanged
		if metric.values[i] < 0 || sValues[i] < 0 {
			metric.record[i] = false
			skips++
			logger.Trace().
				Str("metric", metric.GetName()).
				Str("key", metricKey).
				Float64("numerator", v).
				Float64("denominator", sValues[i]).
				Msg("Negative values")
			return skips, nil
		}
		if metric.record[i] && sRecord[i] {
			if sValues[i] >= x {
				metric.values[i] /= sValues[i]
			} else {
				metric.values[i] = 0
			}
		} else {
			metric.record[i] = false
			skips++
			logger.Trace().
				Str("metric", metric.GetName()).
				Str("key", metricKey).
				Float64("numerator", metric.values[i]).
				Float64("denominator", sValues[i]).
				Msg("Divide threshold calculation skipped")
		}
	}
	return skips, nil
}

func (m *Matrix) MultiplyByScalar(metricKey string, s uint, logger *logging.Logger) (int, error) {
	var skips int
	x := float64(s)
	metric := m.GetMetric(metricKey)
	for i := 0; i < len(metric.values); i++ {
		if metric.record[i] {
			// if current is <= 0
			if metric.values[i] < 0 {
				metric.record[i] = false
				skips++
				logger.Trace().
					Str("metric", metric.GetName()).
					Str("key", metricKey).
					Float64("currentRaw", metric.values[i]).
					Uint("scalar", s).
					Msg("Negative value")
			}
			metric.values[i] *= x
		} else {
			metric.record[i] = false
			skips++
			logger.Trace().
				Str("metric", metric.GetName()).
				Str("key", metricKey).
				Float64("currentRaw", metric.values[i]).
				Uint("scalar", s).
				Msg("Scalar multiplication skipped")
		}
	}
	return skips, nil
}
