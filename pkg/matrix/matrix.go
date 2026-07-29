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
	"log/slog"
	"maps"
	"slices"

	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

// Matrix and its Metrics/Instances are NOT safe for concurrent mutation. Each poll cycle is
// expected to own and mutate exactly one Matrix from a single goroutine; share Matrices across
// goroutines only after all mutation has finished (e.g. read-only access during export).
type Matrix struct {
	UUID           string
	Object         string
	Identifier     string
	globalLabels   map[string]string
	instances      map[string]*Instance
	metrics        map[string]*Metric // ONTAP metric name => metric (in templates, this is left side)
	displayMetrics map[string]string  // display name of metric to => metric name (in templates, this is right side)
	exportOptions  *node.Node
	exportable     bool
}

func New(uuid, object string, identifier string) *Matrix {
	me := Matrix{UUID: uuid, Object: object, Identifier: identifier}
	me.globalLabels = make(map[string]string)
	me.instances = make(map[string]*Instance)
	me.metrics = make(map[string]*Metric)
	me.displayMetrics = make(map[string]string)
	me.exportOptions = DefaultExportOptions()
	me.exportable = true
	return &me
}

// IsExportable indicates whether this matrix is meant to be exported or not
// (some data is only collected to be aggregated by plugins)
func (m *Matrix) IsExportable() bool {
	return m.exportable
}

func (m *Matrix) SetExportable(b bool) {
	m.exportable = b
}

type cloneOptions struct {
	copyData                 bool
	copyMetrics              bool
	copyInstances            bool
	preserveExportability    bool
	preservePartialInstances bool
	labelNames               []string
	metricNames              []string
}

// CloneMetricTemplate copies the matrix metadata and metrics without instances or values.
func (m *Matrix) CloneMetricTemplate() *Matrix {
	return m.clone(cloneOptions{copyMetrics: true, preserveExportability: true})
}

// CloneForCollection copies the matrix metadata, metrics, and instances without metric values.
func (m *Matrix) CloneForCollection() *Matrix {
	return m.clone(cloneOptions{
		copyMetrics:           true,
		copyInstances:         true,
		preserveExportability: true,
	})
}

// Clone copies all matrix state, including metric values, exportability, and partial instances.
func (m *Matrix) Clone() *Matrix {
	return m.clone(cloneOptions{
		copyData:                 true,
		copyMetrics:              true,
		copyInstances:            true,
		preserveExportability:    true,
		preservePartialInstances: true,
	})
}

// CloneSelected copies selected metrics and labels for comparison. The clone includes metric
// values and instances, but its instances are not exportable.
func (m *Matrix) CloneSelected(metricNames, labelNames []string) *Matrix {
	return m.clone(cloneOptions{
		copyData:      true,
		copyMetrics:   len(metricNames) > 0,
		copyInstances: true,
		labelNames:    labelNames,
		metricNames:   metricNames,
	})
}

// CloneEmpty copies matrix metadata without metrics or instances.
func (m *Matrix) CloneEmpty() *Matrix {
	return m.clone(cloneOptions{preserveExportability: true})
}

func (m *Matrix) clone(options cloneOptions) *Matrix {
	clone := &Matrix{
		UUID:           m.UUID,
		Object:         m.Object,
		Identifier:     m.Identifier,
		globalLabels:   maps.Clone(m.globalLabels),
		exportOptions:  m.exportOptions.Copy(),
		exportable:     m.exportable,
		displayMetrics: make(map[string]string),
	}

	if options.copyInstances {
		clone.instances = make(map[string]*Instance, len(m.GetInstances()))
		for key, instance := range m.GetInstances() {
			if options.preserveExportability {
				clone.instances[key] = instance.clone(instance.IsExportable(), options.labelNames...)
			} else {
				clone.instances[key] = instance.clone(false, options.labelNames...)
			}
			if options.preservePartialInstances {
				clone.instances[key].SetPartial(instance.IsPartial())
			}
		}
	} else {
		clone.instances = make(map[string]*Instance)
	}

	if options.copyMetrics {
		if len(options.metricNames) > 0 {
			clone.metrics = make(map[string]*Metric, len(options.metricNames))
			for _, metricName := range options.metricNames {
				metric, ok := m.GetMetrics()[metricName]
				if ok {
					c := metric.clone(options.copyData)
					clone.metrics[metricName] = c
					clone.displayMetrics[c.GetName()] = metricName
				}
			}
		} else {
			clone.metrics = make(map[string]*Metric, len(m.GetMetrics()))
			for key, metric := range m.GetMetrics() {
				c := metric.clone(options.copyData)
				clone.metrics[key] = c
				clone.displayMetrics[c.GetName()] = key
			}
		}
	} else {
		clone.metrics = make(map[string]*Metric)
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

func (m *Matrix) DisplayMetricKey(name string) string {
	return m.displayMetrics[name]
}

func (m *Matrix) GetMetric(key string) *Metric {
	if metric, has := m.metrics[key]; has {
		return metric
	}
	return nil
}

// MustGetMetric returns the metric identified by key, or panics if no such metric exists.
// Use this only at a point where key is guaranteed to have been registered earlier (e.g. via
// NewMetricsFloat64/NewMetricFloat64 with the same key) — a panic here means a missing or
// typo'd key was introduced by a code change, not a runtime condition callers can recover
// from, so it should fail loudly and immediately rather than be silently logged and ignored.
func (m *Matrix) MustGetMetric(key string) *Metric {
	metric := m.GetMetric(key)
	if metric == nil {
		panic(fmt.Sprintf("matrix %q: metric %q is not registered", m.Object, key))
	}
	return metric
}

// MustSetValueInt64 MustSetValueUint8, MustSetValueUint64, MustSetValueFloat64, and
// MustAddValueInt64/MustAddValueUint64 are convenience wrappers around MustGetMetric for the
// common case of setting/adding a single value on a single already-resolved Instance. Prefer
// these at single-shot call sites (e.g. once-per-poll-cycle metadata updates). For per-instance
// loops, resolve the *Metric once via MustGetMetric before the loop instead, so the key lookup
// isn't repeated per instance.
func (m *Matrix) MustSetValueInt64(key string, i *Instance, v int64) {
	m.MustGetMetric(key).SetValueInt64(i, v)
}

func (m *Matrix) MustSetValueUint8(key string, i *Instance, v uint8) {
	m.MustGetMetric(key).SetValueUint8(i, v)
}

func (m *Matrix) MustSetValueUint64(key string, i *Instance, v uint64) {
	m.MustGetMetric(key).SetValueUint64(i, v)
}

func (m *Matrix) MustSetValueFloat64(key string, i *Instance, v float64) {
	m.MustGetMetric(key).SetValueFloat64(i, v)
}

func (m *Matrix) MustAddValueInt64(key string, i *Instance, v int64) {
	m.MustGetMetric(key).AddValueInt64(i, v)
}

func (m *Matrix) MustAddValueUint64(key string, i *Instance, v uint64) {
	m.MustGetMetric(key).AddValueUint64(i, v)
}

func (m *Matrix) MustAddValueFloat64(key string, i *Instance, v float64) {
	m.MustGetMetric(key).AddValueFloat64(i, v)
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

// GetOrCreateMetric returns the existing metric for key if present, otherwise it creates a new one
// with the given dataType (default "float64" when omitted) and returns it.
func (m *Matrix) GetOrCreateMetric(key string, dataType ...string) (*Metric, error) {
	if metric := m.GetMetric(key); metric != nil {
		return metric, nil
	}
	dt := "float64"
	if len(dataType) > 0 && dataType[0] != "" {
		dt = dataType[0]
	}
	return m.NewMetricType(key, dt)
}

// NewMetricsFloat64 idempotently ensures a float64 metric exists for each key, creating any that
// are missing. It returns the first error encountered, wrapped with the offending key.
func (m *Matrix) NewMetricsFloat64(keys ...string) error {
	for _, key := range keys {
		if _, err := m.GetOrCreateMetric(key); err != nil {
			return fmt.Errorf("error while creating metric %s: %w", key, err)
		}
	}
	return nil
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
	if metric, has := m.metrics[key]; has {
		delete(m.displayMetrics, metric.GetName())
	}
	delete(m.metrics, key)
}

func (m *Matrix) PurgeMetrics() {
	m.metrics = make(map[string]*Metric)
	m.displayMetrics = make(map[string]string)
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

// MustGetInstance returns the instance identified by key, or panics if no such instance exists.
// Use this only at a point where key is guaranteed to have been registered earlier (e.g. a
// collector's fixed schedule task names) — a panic here means the instance was never created,
// which is a programming error, not a runtime condition callers can recover from.
func (m *Matrix) MustGetInstance(key string) *Instance {
	instance := m.GetInstance(key)
	if instance == nil {
		panic(fmt.Sprintf("matrix %q: instance %q is not registered", m.Object, key))
	}
	return instance
}

func (m *Matrix) GetInstances() map[string]*Instance {
	return m.instances
}

func (m *Matrix) PurgeInstances() {
	m.instances = make(map[string]*Instance)
}

func (m *Matrix) GetInstanceKeys() []string {
	return slices.Collect(maps.Keys(m.instances))
}

func (m *Matrix) NewInstance(key string) (*Instance, error) {

	var instance *Instance

	if _, has := m.instances[key]; has {
		return nil, errs.New(ErrDuplicateInstanceKey, key)
	}

	instance = newInstance(len(m.instances)) // index is current count of instances

	for _, metric := range m.GetMetrics() {
		metric.Append()
	}

	m.instances[key] = instance
	return instance, nil
}

// GetOrCreateInstance returns the existing instance for key if present (created=false), otherwise
// it creates a new instance and returns it with created=true. This lets callers run one-time
// initialization only when the instance did not already exist:
//
//	instance, created := m.GetOrCreateInstance(key)
//	if created {
//	    instance.SetLabels(...)
//	}
func (m *Matrix) GetOrCreateInstance(key string) (*Instance, bool) {
	if existing := m.GetInstance(key); existing != nil {
		return existing, false
	}
	// NewInstance can only fail with ErrDuplicateInstanceKey, which GetInstance already ruled out.
	newInstance, _ := m.NewInstance(key)
	return newInstance, true
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
				i.index--
			}
		}
	}
}

func (m *Matrix) SetGlobalLabel(label, value string) {
	m.globalLabels[label] = value
}

// SetGlobalLabels copies allLabels to globalLabels when the label does not exist in globalLabels
func (m *Matrix) SetGlobalLabels(allLabels map[string]string) {
	if allLabels == nil {
		return
	}
	for key, val := range allLabels {
		if _, has := m.globalLabels[key]; !has {
			m.globalLabels[key] = val
		}
	}
}

func (m *Matrix) GetGlobalLabels() map[string]string {
	return m.globalLabels
}

func (m *Matrix) GetExportOptions() *node.Node {
	return m.exportOptions
}

func (m *Matrix) SetExportOptions(e *node.Node) {
	if e == nil {
		m.exportOptions = DefaultExportOptions()
		return
	}
	m.exportOptions = e
}

// DefaultExportOptions returns export options that mark every instance label as exportable
// ("include_all_labels"). Use this when a matrix's own export options should not be inherited
// from another matrix, e.g. after cloning a template matrix for a plugin-computed rollup.
func DefaultExportOptions() *node.Node {
	n := node.NewS("export_options")
	n.NewChildS("include_all_labels", "true")
	return n
}

// NewExportOptions builds the export_options node tree that marks the given labels as the
// instance_keys to export, equivalent to the commonly hand-built:
//
//	exportOptions := node.NewS("export_options")
//	instanceKeys := exportOptions.NewChildS("instance_keys", "")
//	for _, k := range keys {
//	    instanceKeys.NewChildS("", k)
//	}
func NewExportOptions(keys ...string) *node.Node {
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	for _, k := range keys {
		instanceKeys.NewChildS("", k)
	}
	return exportOptions
}

// NewExportOptionsWithLabels is like NewExportOptions but also adds an instance_labels child
// with one entry per label, equivalent to the commonly hand-built:
//
//	exportOptions := node.NewS("export_options")
//	instanceKeys := exportOptions.NewChildS("instance_keys", "")
//	for _, k := range keys {
//	    instanceKeys.NewChildS("", k)
//	}
//	instanceLabels := exportOptions.NewChildS("instance_labels", "")
//	for _, l := range labels {
//	    instanceLabels.NewChildS("", l)
//	}
func NewExportOptionsWithLabels(keys []string, labels []string) *node.Node {
	exportOptions := NewExportOptions(keys...)
	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	for _, l := range labels {
		instanceLabels.NewChildS("", l)
	}
	return exportOptions
}

// Delta vector arithmetics
func (m *Matrix) Delta(metricKey string, prevMat *Matrix, cachedData *Matrix, allowPartialAggregation bool, logger *slog.Logger) (int, error) {
	var skips int
	prevMetric := prevMat.GetMetric(metricKey)
	curMetric := m.GetMetric(metricKey)
	cachedMetric := cachedData.GetMetric(metricKey)
	if prevMetric == nil || curMetric == nil {
		return 0, errs.New(errs.ErrMissingMetric, metricKey)
	}
	prevRaw := prevMetric.values
	prevRecord := prevMetric.record
	for key, currInstance := range m.GetInstances() {
		// check if this instance key exists in previous matrix
		prevInstance := prevMat.GetInstance(key)
		currIndex := currInstance.index
		curRaw := curMetric.values[currIndex]
		if prevInstance != nil {
			prevIndex := prevInstance.index
			if curMetric.record[currIndex] && prevRecord[prevIndex] {
				curMetric.values[currIndex] -= prevRaw[prevIndex]
				curCooked := curMetric.values[currIndex]
				// Sometimes ONTAP sends spurious zeroes or values less than the previous poll.
				// Detect these cases and don't publish them, otherwise the subsequent poll will have large spikes.
				// Ensure that the current cooked metric (curCooked) is not zero when either the current raw metric (curRaw) or the previous raw metric (prevRaw[prevIndex]) is zero.
				// A non-zero curCooked under these conditions indicates an issue with the current or previous poll.
				isInvalidZero := (curRaw == 0 || prevRaw[prevIndex] == 0) && curCooked != 0
				isNegative := curCooked < 0

				// Check for partial Aggregation
				var ppaOk, cpaOk bool
				if !allowPartialAggregation {
					ppaOk = prevInstance.IsPartial()
					cpaOk = currInstance.IsPartial()
				}

				if isInvalidZero || isNegative || ppaOk || cpaOk {
					curMetric.record[currIndex] = false
					skips++
				}

				if isInvalidZero || isNegative {
					if cachedMetric != nil {
						cachedMetric.record[currIndex] = false
					}
				}

				if ppaOk || cpaOk {
					logger.Debug(
						"Partial Aggregation",
						slog.String("metric", curMetric.GetName()),
						slog.Float64("currentRaw", curRaw),
						slog.Float64("previousRaw", prevRaw[prevIndex]),
						slog.Bool("prevPartial", ppaOk),
						slog.Bool("curPartial", cpaOk),
						slog.Any("instanceLabels", currInstance.GetLabels()),
						slog.String("instKey", key),
					)
				}
			} else {
				curMetric.record[currIndex] = false
				skips++
			}
		} else {
			curMetric.record[currIndex] = false
			skips++
		}
	}
	return skips, nil
}

func (m *Matrix) Divide(metricKey string, baseKey string) (int, error) {
	var skips int
	metric := m.GetMetric(metricKey)
	if metric == nil {
		return 0, errs.New(errs.ErrMissingMetric, metricKey)
	}
	base := m.GetMetric(baseKey)
	if base == nil {
		return 0, errs.New(errs.ErrMissingMetric, baseKey)
	}
	sValues := base.values
	sRecord := base.record
	if len(metric.values) != len(sValues) {
		return 0, errs.New(ErrUnequalVectors, fmt.Sprintf("numerator=%d, denominator=%d", len(metric.values), len(sValues)))
	}
	for _, instance := range m.GetInstances() {
		i := instance.index
		if metric.record[i] && sRecord[i] {
			if sValues[i] != 0 {
				// Don't pass along the value if the numerator or denominator is < 0
				// A denominator of zero is fine
				if metric.values[i] < 0 || sValues[i] < 0 {
					metric.record[i] = false
					skips++
				}
				metric.values[i] /= sValues[i]
			} else {
				metric.values[i] = 0
			}
		} else {
			metric.record[i] = false
			skips++
		}
	}
	return skips, nil
}

// DivideWithThreshold applicable for latency counters
func (m *Matrix) DivideWithThreshold(metricKey string, baseKey string, threshold int, curRawMat *Matrix, prevRawMat *Matrix, timestampMetricName string, logger *slog.Logger) (int, error) {
	var skips int
	x := float64(threshold)
	curRawMetric := curRawMat.GetMetric(metricKey)
	prevRawMetric := prevRawMat.GetMetric(metricKey)
	if curRawMetric == nil || prevRawMetric == nil {
		return 0, errs.New(errs.ErrMissingMetric, metricKey)
	}
	curBaseRawMetric := curRawMat.GetMetric(baseKey)
	prevBaseRawMetric := prevRawMat.GetMetric(baseKey)
	if curBaseRawMetric == nil || prevBaseRawMetric == nil {
		return 0, errs.New(errs.ErrMissingMetric, baseKey)
	}
	metric := m.GetMetric(metricKey)
	if metric == nil {
		return 0, errs.New(errs.ErrMissingMetric, metricKey)
	}
	base := m.GetMetric(baseKey)
	if base == nil {
		return 0, errs.New(errs.ErrMissingMetric, baseKey)
	}
	time := m.GetMetric(timestampMetricName)
	var tValues []float64
	if time != nil {
		tValues = time.values
	}
	sValues := base.values
	sRecord := base.record
	if len(metric.values) != len(sValues) || len(sValues) != len(tValues) {
		return 0, errs.New(ErrUnequalVectors, fmt.Sprintf("numerator=%d, denominator=%d, time=%d", len(metric.values), len(sValues), len(tValues)))
	}
	for key, instance := range m.GetInstances() {
		i := instance.index
		v := metric.values[i]
		// Don't pass along the value if the numerator or denominator is < 0
		// It is important to check sValues[i] < 0 and allow a zero so pass=true and m.values[i] remains unchanged
		switch {
		case metric.values[i] < 0 || sValues[i] < 0:
			metric.record[i] = false
			skips++
		case metric.record[i] && sRecord[i]:
			minimumBase := tValues[i] * x
			metricName := metric.GetName()
			if metricName == "optimal_point_latency" || metricName == "scan_latency" || m.Object == "ontaps3_svm" {
				// An exception is made for these counters because the base counter always has a few IOPS
				minimumBase = 0
			}
			if sValues[i] > minimumBase {
				metric.values[i] /= sValues[i]
				// if cooked latency is greater than 5 secs log delta values
				if metric.values[i] > 5_000_000 {
					if len(metric.values) == len(curRawMetric.values) && len(curRawMetric.values) == len(prevRawMetric.values) &&
						len(prevRawMetric.values) == len(curBaseRawMetric.values) && len(curBaseRawMetric.values) == len(prevBaseRawMetric.values) {
						logger.Debug(
							"Detected high latency value in the metric",
							slog.String("metric", metric.GetName()),
							slog.String("key", metricKey),
							slog.Float64("numerator", v),
							slog.Float64("denominator", sValues[i]),
							slog.Float64("prev_raw_latency", prevRawMetric.values[i]),
							slog.Float64("current_raw_latency", curRawMetric.values[i]),
							slog.Float64("prev_raw_base", prevBaseRawMetric.values[i]),
							slog.Float64("current_raw_base", curBaseRawMetric.values[i]),
							slog.Any("instanceLabels", instance.GetLabels()),
							slog.Float64("minimumBase", minimumBase),
							slog.Int("threshold", threshold),
							slog.String("instKey", key),
						)
					}
				}
			} else {
				metric.values[i] = 0
			}
		default:
			metric.record[i] = false
			skips++
		}
	}
	return skips, nil
}

func (m *Matrix) MultiplyByScalar(metricKey string, s uint) (int, error) {
	var skips int
	x := float64(s)
	metric := m.GetMetric(metricKey)
	if metric == nil {
		return 0, errs.New(errs.ErrMissingMetric, metricKey)
	}
	for i := range len(metric.values) {
		if metric.record[i] {
			// if current is <= 0
			if metric.values[i] < 0 {
				metric.record[i] = false
				skips++
			}
			metric.values[i] *= x
		} else {
			metric.record[i] = false
			skips++
		}
	}
	return skips, nil
}

func (m *Matrix) Skip(metricKey string) int {
	var skips int
	metric := m.GetMetric(metricKey)
	if metric != nil {
		for i := range len(metric.values) {
			metric.record[i] = false
			skips++
		}
	}
	return skips
}
