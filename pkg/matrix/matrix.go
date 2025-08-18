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
	"strings"

	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

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

type With struct {
	Data             bool
	Metrics          bool
	Instances        bool
	ExportInstances  bool
	PartialInstances bool
	Labels           []string
	MetricsNames     []string
}

func New(uuid, object string, identifier string) *Matrix {
	me := Matrix{UUID: uuid, Object: object, Identifier: identifier}
	me.globalLabels = make(map[string]string)
	me.instances = make(map[string]*Instance)
	me.metrics = make(map[string]*Metric)
	me.displayMetrics = make(map[string]string)
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

func (m *Matrix) Clone(with With) *Matrix {
	clone := &Matrix{
		UUID:           m.UUID,
		Object:         m.Object,
		Identifier:     m.Identifier,
		globalLabels:   m.globalLabels,
		exportOptions:  nil,
		exportable:     m.exportable,
		displayMetrics: make(map[string]string),
	}
	// Deep clone exportOptions if it is not nil
	if m.exportOptions != nil {
		clone.exportOptions = m.exportOptions.Copy()
	}

	if with.Instances {
		clone.instances = make(map[string]*Instance, len(m.GetInstances()))
		for key, instance := range m.GetInstances() {
			if with.ExportInstances {
				clone.instances[key] = instance.Clone(instance.IsExportable(), with.Labels...)
			} else {
				clone.instances[key] = instance.Clone(false, with.Labels...)
			}
			if with.PartialInstances {
				clone.instances[key].SetPartial(instance.IsPartial())
			}
		}
	} else {
		clone.instances = make(map[string]*Instance)
	}

	if with.Metrics {
		if len(with.MetricsNames) > 0 {
			clone.metrics = make(map[string]*Metric, len(with.MetricsNames))
			for _, metricName := range with.MetricsNames {
				metric, ok := m.GetMetrics()[metricName]
				if ok {
					c := metric.Clone(with.Data)
					clone.metrics[metricName] = c
					clone.displayMetrics[c.GetName()] = metricName
				}
			}
		} else {
			clone.metrics = make(map[string]*Metric, len(m.GetMetrics()))
			for key, metric := range m.GetMetrics() {
				c := metric.Clone(with.Data)
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

func (m *Matrix) PurgeMetrics() {
	m.metrics = make(map[string]*Metric)
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
	return slices.Collect(maps.Keys(m.instances))
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
func (m *Matrix) Delta(metricKey string, prevMat *Matrix, cachedData *Matrix, allowPartialAggregation bool, logger *slog.Logger) (int, error) {
	var skips int
	prevMetric := prevMat.GetMetric(metricKey)
	curMetric := m.GetMetric(metricKey)
	cachedMetric := cachedData.GetMetric(metricKey)
	if prevMetric == nil || curMetric == nil {
		return 0, errs.New(errs.ErrMissingMetric, metricKey)
	}
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
	sRecord := base.GetRecords()
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
	sRecord := base.GetRecords()
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
