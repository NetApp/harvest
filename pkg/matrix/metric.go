//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package matrix

import (
    "fmt"
    //"goharvest2/pkg/errors"
    "goharvest2/pkg/dict"
)

// Metric struct and related methods

type Metric struct {
    Index   int
    Name    string
    Enabled bool
    Size    int // 1 for scalar metrics
    /* extended fields for ZapiPerf counters */
    Properties  string
    BaseCounter string
    /* fields for Array counters */
    Dimensions int
    Labels     *dict.Dict
}

func (m *Metric) IsScalar() bool {
    if !(m.Size >= 0) {
        panic(fmt.Sprintf("metric [%s] has size %d", m.Name, m.Size))
    }
    return m.Size == 1
}

func (m *Metric) HasLabels() bool {
    return m.Labels != nil && m.Labels.Size() != 0
}

func (m *Matrix) GetMetric(key string) *Metric {

    if metric, found := m.Metrics[key]; found {
        return metric
    }
    return nil
}

func (m *Matrix) add_metric(key string, metric *Metric) error {

    if _, exists := m.Metrics[key]; exists {
        //return errors.New(errors.MATRIX_HASH, "metric [" + key + "] already in cache")
        panic("metric [" + key + "] already in cache")
    }
    metric.Index = m.SizeMetrics()
    m.Metrics[key] = metric

    if !m.IsEmpty() {
        m.Data = append(m.Data, make([]float64, m.SizeInstances()))
        for j := 0; j < m.SizeInstances(); j += 1 {
            m.Data[metric.Index][j] = NAN
        }
    }

    return nil
}

// Create new metric and add to cache
func (m *Matrix) AddMetric(key, name string, enabled bool) (*Metric, error) {
    metric := &Metric{Name: name, Enabled: enabled, Size: 1}
    return metric, m.add_metric(key, metric)
}

func (m *Matrix) AddMetricExtended(key, name, base, properties string, enabled bool) (*Metric, error) {
    metric := &Metric{Name: name, Enabled: enabled, Size: 1, BaseCounter: base, Properties: properties}
    metric.Labels = dict.New()
    return metric, m.add_metric(key, metric)
}

/*
// Create 1D Array Matric
func (m *Matrix) AddArrayMetric(key, name string, labels []string, enabled bool) (*Metric, error) {
    metric := &Metric{Name: name, Labels: labels, Enabled: enabled, Dimensions: 1, Size: len(labels)}
    return metric, m.add_metric(key, metric)
}

// Similar to AddMetric, but metric is initialized. This allows collectors
// to add extended fields to metric or create multidimensional Array metric.
//
// Method should be used with caution: incorrect "size" will corrupt data
// or make Harvest panic
func (m *Matrix) AddCustomMetric(key string, metric *Metric) error {

    // sanity check: metric should come with size
    if metric.Size == 0 {
        return errors.New(errors.MATRIX_INV_PARAM, "array metric has 0 size")
    }
    return m.add_metric(key, metric)
}
*/
