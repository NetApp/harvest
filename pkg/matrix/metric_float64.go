/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"strconv"
)

type MetricFloat64 struct {
	*AbstractMetric
	values []float64
}

func (m *MetricFloat64) Clone(deep bool) Metric {
	clone := MetricFloat64{AbstractMetric: m.AbstractMetric.Clone(deep)}
	if deep && len(m.values) != 0 {
		clone.values = make([]float64, len(m.values))
		copy(clone.values, m.values)
	}
	return &clone
}

// Storage resizing methods

func (m *MetricFloat64) Reset(size int) {
	m.record = make([]bool, size)
	m.pass = make([]bool, size)
	m.values = make([]float64, size)
}

func (m *MetricFloat64) Append() {
	m.record = append(m.record, false)
	m.pass = append(m.pass, false)
	m.values = append(m.values, 0)
}

// Remove element at index, shift everything to the left
func (m *MetricFloat64) Remove(index int) {
	for i := index; i < len(m.values)-1; i++ {
		m.record[i] = m.record[i+1]
		m.pass[i] = m.pass[i+1]
		m.values[i] = m.values[i+1]
	}
	m.record = m.record[:len(m.record)-1]
	m.pass = m.pass[:len(m.pass)-1]
	m.values = m.values[:len(m.values)-1]
}

// Write methods

func (m *MetricFloat64) SetValueInt64(i *Instance, v int64) error {
	m.record[i.index] = true
	m.pass[i.index] = true
	m.values[i.index] = float64(v)
	return nil
}

func (m *MetricFloat64) SetValueUint8(i *Instance, v uint8) error {
	m.record[i.index] = true
	m.pass[i.index] = true
	m.values[i.index] = float64(v)
	return nil
}

func (m *MetricFloat64) SetValueUint64(i *Instance, v uint64) error {
	m.record[i.index] = true
	m.pass[i.index] = true
	m.values[i.index] = float64(v)
	return nil
}

func (m *MetricFloat64) SetValueFloat64(i *Instance, v float64) error {
	m.record[i.index] = true
	m.pass[i.index] = true
	m.values[i.index] = v
	return nil
}

func (m *MetricFloat64) SetValueString(i *Instance, v string) error {
	var x float64
	var err error
	if x, err = strconv.ParseFloat(v, 64); err == nil {
		m.record[i.index] = true
		m.pass[i.index] = true
		m.values[i.index] = x
		return nil
	}
	return err
}

func (m *MetricFloat64) SetValueBytes(i *Instance, v []byte) error {
	return m.SetValueString(i, string(v))
}

func (m *MetricFloat64) AddValueInt64(i *Instance, n int64) error {
	v, _, _ := m.GetValueInt64(i)
	return m.SetValueInt64(i, v+n)
}

func (m *MetricFloat64) AddValueUint8(i *Instance, n uint8) error {
	v, _, _ := m.GetValueUint8(i)
	return m.SetValueUint8(i, v+n)
}

func (m *MetricFloat64) AddValueUint64(i *Instance, n uint64) error {
	v, _, _ := m.GetValueUint64(i)
	return m.SetValueUint64(i, v+n)
}

func (m *MetricFloat64) AddValueFloat64(i *Instance, n float64) error {
	v, _, _ := m.GetValueFloat64(i)
	return m.SetValueFloat64(i, v+n)
}

func (m *MetricFloat64) AddValueString(i *Instance, v string) error {
	var (
		x, n float64
		err  error
		has  bool
	)
	if x, err = strconv.ParseFloat(v, 64); err != nil {
		return err
	}
	if n, has, _ = m.GetValueFloat64(i); has {
		return m.SetValueFloat64(i, x+n)
	}
	return m.SetValueFloat64(i, x)
}

// Read methods

func (m *MetricFloat64) GetValueInt(i *Instance) (int, bool, bool) {
	return int(m.values[i.index]), m.record[i.index], m.pass[i.index]
}

func (m *MetricFloat64) GetValueInt64(i *Instance) (int64, bool, bool) {
	return int64(m.values[i.index]), m.record[i.index], m.pass[i.index]
}

func (m *MetricFloat64) GetValueUint8(i *Instance) (uint8, bool, bool) {
	return uint8(m.values[i.index]), m.record[i.index], m.pass[i.index]
}

func (m *MetricFloat64) GetValueUint64(i *Instance) (uint64, bool, bool) {
	return uint64(m.values[i.index]), m.record[i.index], m.pass[i.index]
}

func (m *MetricFloat64) GetValueFloat64(i *Instance) (float64, bool, bool) {
	return m.values[i.index], m.record[i.index], m.pass[i.index]
}

func (m *MetricFloat64) GetValueString(i *Instance) (string, bool, bool) {
	return strconv.FormatFloat(m.values[i.index], 'f', -1, 64), m.record[i.index], m.pass[i.index]
}

func (m *MetricFloat64) GetValueBytes(i *Instance) ([]byte, bool, bool) {
	s, ok, pass := m.GetValueString(i)
	return []byte(s), ok, pass
}

// vector arithmetics

func (m *MetricFloat64) GetValuesFloat64() []float64 {
	return m.values
}

func (m *MetricFloat64) Delta(s Metric, logger *logging.Logger) (int, error) {
	var skips int
	prevRaw := s.GetValuesFloat64()
	sRecord := s.GetRecords()
	pass := m.GetPass()
	if len(m.values) != len(prevRaw) || len(pass) != len(prevRaw) {
		return 0, errs.New(ErrUnequalVectors, fmt.Sprintf("minuend=%d, subtrahend=%d", len(m.values), len(prevRaw)))
	}
	for i := range m.values {
		if m.record[i] && sRecord[i] {
			curRaw := m.values[i]
			// reset pass
			pass[i] = true
			m.values[i] -= prevRaw[i]
			// Sometimes ONTAP sends spurious zeroes. Detect and don't publish the negative delta
			// or the next poll that will show a large spike.
			// Distinguish invalid zeros from valid ones. Invalid ones happen when the delta != 0
			if (curRaw == 0 || prevRaw[i] == 0) && m.values[i] != 0 {
				pass[i] = false
				skips++
				logger.Trace().
					Str("metric", m.GetName()).
					Float64("currentRaw", curRaw).
					Float64("previousRaw", prevRaw[i]).
					Msg("Negative cooked value")
			}
		}
	}
	return skips, nil
}

func (m *MetricFloat64) Divide(s Metric, logger *logging.Logger) (int, error) {
	var skips int
	sValues := s.GetValuesFloat64()
	sRecord := s.GetRecords()
	pass := m.GetPass()
	if len(m.values) != len(sValues) || len(pass) != len(sValues) {
		return 0, errs.New(ErrUnequalVectors, fmt.Sprintf("numerator=%d, denominator=%d", len(m.values), len(sValues)))
	}
	for i := 0; i < len(m.values); i++ {
		if m.record[i] && sRecord[i] && sValues[i] != 0 {
			// reset pass
			pass[i] = true
			// Don't pass along the value if the numerator or denominator is < 0
			// A denominator of zero is fine
			if m.values[i] < 0 || sValues[i] < 0 {
				pass[i] = false
				skips++
				logger.Trace().
					Str("metric", m.GetName()).
					Float64("numerator", m.values[i]).
					Float64("denominator", sValues[i]).
					Msg("No pass values")
			}
			m.values[i] /= sValues[i]
		}
	}
	return skips, nil
}

func (m *MetricFloat64) DivideWithThreshold(s Metric, t int, logger *logging.Logger) (int, error) {
	var skips int
	x := float64(t)
	sValues := s.GetValuesFloat64()
	sRecord := s.GetRecords()
	pass := m.GetPass()
	if len(m.values) != len(sValues) || len(pass) != len(sValues) {
		return 0, errs.New(ErrUnequalVectors, fmt.Sprintf("numerator=%d, denominator=%d", len(m.values), len(sValues)))
	}
	for i := 0; i < len(m.values); i++ {
		v := m.values[i]
		// reset pass
		pass[i] = true
		// Don't pass along the value if the numerator or denominator is < 0
		// It's important to check sValues[i] < 0 and allow a zero so pass=true and m.values[i] remains unchanged
		if m.values[i] < 0 || sValues[i] < 0 {
			pass[i] = false
			skips++
			logger.Trace().
				Str("metric", m.GetName()).
				Float64("numerator", v).
				Float64("denominator", sValues[i]).
				Msg("Negative values")
		}
		if m.record[i] && sRecord[i] && sValues[i] >= x {
			m.values[i] /= sValues[i]
		}
	}
	return skips, nil
}

func (m *MetricFloat64) MultiplyByScalar(s uint, logger *logging.Logger) (int, error) {
	var skips int
	x := float64(s)
	pass := m.GetPass()
	for i := 0; i < len(m.values); i++ {
		if m.record[i] {
			// reset pass
			pass[i] = true
			skips++
			// if current is <= 0
			if m.values[i] < 0 {
				pass[i] = false
				logger.Trace().
					Str("metric", m.GetName()).
					Float64("currentRaw", m.values[i]).
					Uint("scalar", s).
					Msg("Negative value")
			}
			m.values[i] *= x
		}
	}
	return skips, nil
}

func (m *MetricFloat64) Print() {
	for i := range m.values {
		if m.record[i] && m.pass[i] {
			fmt.Printf("%s%v%s ", color.Green, m.values[i], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, m.values[i], color.End)
		}
	}
}
