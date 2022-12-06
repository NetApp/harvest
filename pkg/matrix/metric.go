/*
 * Copyright NetApp Inc, 2021 All rights reserved

Package Description:
   Parse raw metric name from collector template

Examples:
   Simple name (e.g. "metric_name"), means both name and display are the same
   Custom name (e.g. "metric_name => custom_name") is parsed as display name.
*/

package matrix

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"strconv"
)

type Metric struct {
	name       string
	dataType   string
	property   string
	comment    string
	array      bool
	histogram  bool
	exportable bool
	labels     *dict.Dict
	buckets    *[]string
	record     []bool
	values     []float64
}

func (m *Metric) Clone(deep bool) *Metric {
	clone := Metric{
		name:       m.name,
		dataType:   m.dataType,
		property:   m.property,
		comment:    m.comment,
		exportable: m.exportable,
		array:      m.array,
		histogram:  m.histogram,
		buckets:    m.buckets,
	}
	if m.labels != nil {
		clone.labels = m.labels.Copy()
	}
	if deep {
		if len(m.record) != 0 {
			clone.record = make([]bool, len(m.record))
			copy(clone.record, m.record)
		}
		if len(m.values) != 0 {
			clone.values = make([]float64, len(m.values))
			copy(clone.values, m.values)
		}
	}
	return &clone
}

func (m *Metric) GetName() string {
	return m.name
}

func (m *Metric) IsExportable() bool {
	return m.exportable
}

func (m *Metric) SetExportable(b bool) {
	m.exportable = b
}

func (m *Metric) GetType() string {
	return m.dataType
}

func (m *Metric) GetProperty() string {
	return m.property
}

func (m *Metric) SetProperty(p string) {
	m.property = p
}

func (m *Metric) GetComment() string {
	return m.comment
}

func (m *Metric) SetComment(c string) {
	m.comment = c
}

func (m *Metric) IsArray() bool {
	return m.array
}

func (m *Metric) SetArray(c bool) {
	m.array = c
}

func (m *Metric) SetLabel(key, value string) {
	if m.labels == nil {
		m.labels = dict.New()
	}
	m.labels.Set(key, value)
}

func (m *Metric) SetHistogram(b bool) {
	m.histogram = b
}

func (m *Metric) IsHistogram() bool {
	return m.histogram
}

func (m *Metric) Buckets() *[]string {
	return m.buckets
}

func (m *Metric) SetBuckets(buckets *[]string) {
	m.buckets = buckets
}

func (m *Metric) SetLabels(labels *dict.Dict) {
	m.labels = labels
}

func (m *Metric) GetLabel(key string) string {
	if m.labels != nil {
		return m.labels.Get(key)
	}
	return ""
}

func (m *Metric) GetLabels() *dict.Dict {
	return m.labels

}
func (m *Metric) HasLabels() bool {
	return m.labels != nil && m.labels.Size() != 0
}

func (m *Metric) GetRecords() []bool {
	return m.record
}

func (m *Metric) SetValueNAN(i *Instance) {
	m.record[i.index] = false
}

// Storage resizing methods

func (m *Metric) Reset(size int) {
	m.record = make([]bool, size)
	m.values = make([]float64, size)
}

func (m *Metric) Append() {
	m.record = append(m.record, false)
	m.values = append(m.values, 0)
}

// Remove element at index, shift everything to the left
func (m *Metric) Remove(index int) {
	for i := index; i < len(m.values)-1; i++ {
		m.record[i] = m.record[i+1]
		m.values[i] = m.values[i+1]
	}
	m.record = m.record[:len(m.record)-1]
	m.values = m.values[:len(m.values)-1]
}

// Write methods

func (m *Metric) SetValueInt64(i *Instance, v int64) error {
	m.record[i.index] = true
	m.values[i.index] = float64(v)
	return nil
}

func (m *Metric) SetValueUint8(i *Instance, v uint8) error {
	m.record[i.index] = true
	m.values[i.index] = float64(v)
	return nil
}

func (m *Metric) SetValueUint64(i *Instance, v uint64) error {
	m.record[i.index] = true
	m.values[i.index] = float64(v)
	return nil
}

func (m *Metric) SetValueFloat64(i *Instance, v float64) error {
	m.record[i.index] = true
	m.values[i.index] = v
	return nil
}

func (m *Metric) SetValueString(i *Instance, v string) error {
	var x float64
	var err error
	if x, err = strconv.ParseFloat(v, 64); err == nil {
		m.record[i.index] = true
		m.values[i.index] = x
		return nil
	}
	return err
}

func (m *Metric) SetValueBytes(i *Instance, v []byte) error {
	return m.SetValueString(i, string(v))
}

func (m *Metric) AddValueInt64(i *Instance, n int64) error {
	v, _ := m.GetValueInt64(i)
	return m.SetValueInt64(i, v+n)
}

func (m *Metric) AddValueUint8(i *Instance, n uint8) error {
	v, _ := m.GetValueUint8(i)
	return m.SetValueUint8(i, v+n)
}

func (m *Metric) AddValueUint64(i *Instance, n uint64) error {
	v, _ := m.GetValueUint64(i)
	return m.SetValueUint64(i, v+n)
}

func (m *Metric) AddValueFloat64(i *Instance, n float64) error {
	v, _ := m.GetValueFloat64(i)
	return m.SetValueFloat64(i, v+n)
}

func (m *Metric) AddValueString(i *Instance, v string) error {
	var (
		x, n float64
		err  error
		has  bool
	)
	if x, err = strconv.ParseFloat(v, 64); err != nil {
		return err
	}
	if n, has = m.GetValueFloat64(i); has {
		return m.SetValueFloat64(i, x+n)
	}
	return m.SetValueFloat64(i, x)
}

// Read methods

func (m *Metric) GetValueInt(i *Instance) (int, bool) {
	v := m.values[i.index]
	val := int(v)
	return val, m.record[i.index]
}

func (m *Metric) GetValueInt64(i *Instance) (int64, bool) {
	v := m.values[i.index]
	val := int64(v)
	return val, m.record[i.index]
}

func (m *Metric) GetValueUint8(i *Instance) (uint8, bool) {
	v := m.values[i.index]
	return uint8(v), m.record[i.index]
}

func (m *Metric) GetValueUint64(i *Instance) (uint64, bool) {
	v := m.values[i.index]
	val := uint64(v)
	return val, m.record[i.index]
}

func (m *Metric) GetValueFloat64(i *Instance) (float64, bool) {
	v := m.values[i.index]
	return v, m.record[i.index]
}

func (m *Metric) GetValueString(i *Instance) (string, bool) {
	v := m.values[i.index]
	return strconv.FormatFloat(v, 'f', -1, 64), m.record[i.index]
}

func (m *Metric) GetValueBytes(i *Instance) ([]byte, bool) {
	s, ok := m.GetValueString(i)
	return []byte(s), ok
}

// Delta vector arithmetics
func (m *Metric) Delta(prevMetric *Metric, prevMat *Matrix, curMat *Matrix, logger *logging.Logger) (int, error) {
	var skips int
	prevRaw := prevMetric.values
	prevRecord := prevMetric.GetRecords()
	for key, currInstance := range curMat.GetInstances() {
		// check if this instance key exists in previous matrix
		prevInstance := prevMat.GetInstance(key)
		currIndex := currInstance.index
		curRaw := m.values[currIndex]
		if prevInstance != nil {
			prevIndex := prevInstance.index
			if m.record[currIndex] && prevRecord[prevIndex] {
				m.values[currIndex] -= prevRaw[prevIndex]
				// Sometimes ONTAP sends spurious zeroes or values less than the previous poll.
				// Detect and don't publish negative deltas or the subsequent poll will show a large spike.
				isInvalidZero := (curRaw == 0 || prevRaw[prevIndex] == 0) && m.values[prevIndex] != 0
				isNegative := m.values[currIndex] < 0
				if isInvalidZero || isNegative {
					m.record[currIndex] = false
					skips++
					logger.Trace().
						Str("metric", m.GetName()).
						Float64("currentRaw", curRaw).
						Float64("previousRaw", prevRaw[prevIndex]).
						Str("instKey", key).
						Msg("Negative cooked value")
				}
			} else {
				m.record[currIndex] = false
				skips++
				logger.Trace().
					Str("metric", m.GetName()).
					Float64("currentRaw", curRaw).
					Float64("previousRaw", prevRaw[prevIndex]).
					Str("instKey", key).
					Msg("Delta calculation skipped")
			}
		} else {
			m.record[currIndex] = false
			skips++
			logger.Trace().
				Str("metric", m.GetName()).
				Float64("currentRaw", curRaw).
				Str("instKey", key).
				Msg("New instance added")
		}
	}
	return skips, nil
}

func (m *Metric) Divide(s *Metric, logger *logging.Logger) (int, error) {
	var skips int
	sValues := s.values
	sRecord := s.GetRecords()
	if len(m.values) != len(sValues) {
		return 0, errs.New(ErrUnequalVectors, fmt.Sprintf("numerator=%d, denominator=%d", len(m.values), len(sValues)))
	}
	for i := 0; i < len(m.values); i++ {
		if m.record[i] && sRecord[i] {
			if sValues[i] != 0 {
				// Don't pass along the value if the numerator or denominator is < 0
				// A denominator of zero is fine
				if m.values[i] < 0 || sValues[i] < 0 {
					m.record[i] = false
					skips++
					logger.Trace().
						Str("metric", m.GetName()).
						Float64("numerator", m.values[i]).
						Float64("denominator", sValues[i]).
						Msg("Divide calculation skipped")
				}
				m.values[i] /= sValues[i]
			} else {
				m.values[i] = 0
			}
		} else {
			m.record[i] = false
			skips++
			logger.Trace().
				Str("metric", m.GetName()).
				Float64("numerator", m.values[i]).
				Float64("denominator", sValues[i]).
				Msg("Divide calculation skipped")
		}
	}
	return skips, nil
}

func (m *Metric) DivideWithThreshold(s *Metric, t int, logger *logging.Logger) (int, error) {
	var skips int
	x := float64(t)
	sValues := s.values
	sRecord := s.GetRecords()
	if len(m.values) != len(sValues) {
		return 0, errs.New(ErrUnequalVectors, fmt.Sprintf("numerator=%d, denominator=%d", len(m.values), len(sValues)))
	}
	for i := 0; i < len(m.values); i++ {
		v := m.values[i]
		// Don't pass along the value if the numerator or denominator is < 0
		// It's important to check sValues[i] < 0 and allow a zero so pass=true and m.values[i] remains unchanged
		if m.values[i] < 0 || sValues[i] < 0 {
			m.record[i] = false
			skips++
			logger.Trace().
				Str("metric", m.GetName()).
				Float64("numerator", v).
				Float64("denominator", sValues[i]).
				Msg("Negative values")
			return skips, nil
		}
		if m.record[i] && sRecord[i] {
			if sValues[i] >= x {
				m.values[i] /= sValues[i]
			} else {
				m.values[i] = 0
			}
		} else {
			m.record[i] = false
			skips++
			logger.Trace().
				Str("metric", m.GetName()).
				Float64("numerator", m.values[i]).
				Float64("denominator", sValues[i]).
				Msg("Divide threshold calculation skipped")
		}
	}
	return skips, nil
}

func (m *Metric) MultiplyByScalar(s uint, logger *logging.Logger) (int, error) {
	var skips int
	x := float64(s)
	for i := 0; i < len(m.values); i++ {
		if m.record[i] {
			// if current is <= 0
			if m.values[i] < 0 {
				m.record[i] = false
				skips++
				logger.Trace().
					Str("metric", m.GetName()).
					Float64("currentRaw", m.values[i]).
					Uint("scalar", s).
					Msg("Negative value")
			}
			m.values[i] *= x
		} else {
			m.record[i] = false
			skips++
			logger.Trace().
				Str("metric", m.GetName()).
				Float64("currentRaw", m.values[i]).
				Uint("scalar", s).
				Msg("Scalar multiplication skipped")
		}
	}
	return skips, nil
}

func (m *Metric) Print() {
	for i := range m.values {
		if m.record[i] {
			fmt.Printf("%s%v ", " ", m.values[i])
		} else {
			fmt.Printf("%s%v ", "!", m.values[i])
		}
	}
}
