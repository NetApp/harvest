/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/logging"
	"strconv"
)

type MetricFloat64 struct {
	*AbstractMetric
	values map[string]float64
}

func (m *MetricFloat64) Clone(deep bool) Metric {
	clone := MetricFloat64{AbstractMetric: m.AbstractMetric.Clone(deep)}
	if deep && len(m.values) != 0 {
		clone.values = make(map[string]float64, len(m.values))
		for key, element := range m.values {
			clone.values[key] = element
		}
	} else {
		clone.values = make(map[string]float64)
	}
	return &clone
}

// Storage resizing methods

func (m *MetricFloat64) Reset(size int) {
	m.record = make(map[string]bool, size)
	m.pass = make(map[string]bool, size)
	m.values = make(map[string]float64, size)
}

// Remove element at index, shift everything to the left
func (m *MetricFloat64) Remove(key string) {
	delete(m.record, key)
	delete(m.pass, key)
	delete(m.values, key)
}

// Write methods

func (m *MetricFloat64) SetValueInt64(i *Instance, v int64) error {
	m.record[i.key] = true
	m.pass[i.key] = true
	m.values[i.key] = float64(v)
	return nil
}

func (m *MetricFloat64) SetValueUint8(i *Instance, v uint8) error {
	m.record[i.key] = true
	m.pass[i.key] = true
	m.values[i.key] = float64(v)
	return nil
}

func (m *MetricFloat64) SetValueUint64(i *Instance, v uint64) error {
	m.record[i.key] = true
	m.pass[i.key] = true
	m.values[i.key] = float64(v)
	return nil
}

func (m *MetricFloat64) SetValueFloat64(i *Instance, v float64) error {
	m.record[i.key] = true
	m.pass[i.key] = true
	m.values[i.key] = v
	return nil
}

func (m *MetricFloat64) SetValueString(i *Instance, v string) error {
	var x float64
	var err error
	if x, err = strconv.ParseFloat(v, 64); err == nil {
		m.record[i.key] = true
		m.pass[i.key] = true
		m.values[i.key] = x
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
	return int(m.values[i.key]), m.record[i.key], m.pass[i.key]
}

func (m *MetricFloat64) GetValueInt64(i *Instance) (int64, bool, bool) {
	return int64(m.values[i.key]), m.record[i.key], m.pass[i.key]
}

func (m *MetricFloat64) GetValueUint8(i *Instance) (uint8, bool, bool) {
	return uint8(m.values[i.key]), m.record[i.key], m.pass[i.key]
}

func (m *MetricFloat64) GetValueUint64(i *Instance) (uint64, bool, bool) {
	return uint64(m.values[i.key]), m.record[i.key], m.pass[i.key]
}

func (m *MetricFloat64) GetValueFloat64(i *Instance) (float64, bool, bool) {
	return m.values[i.key], m.record[i.key], m.pass[i.key]
}

func (m *MetricFloat64) GetValueString(i *Instance) (string, bool, bool) {
	return strconv.FormatFloat(m.values[i.key], 'f', -1, 64), m.record[i.key], m.pass[i.key]
}

func (m *MetricFloat64) GetValueBytes(i *Instance) ([]byte, bool, bool) {
	s, ok, pass := m.GetValueString(i)
	return []byte(s), ok, pass
}

// vector arithmetics

func (m *MetricFloat64) GetValuesFloat64() map[string]float64 {
	return m.values
}

func (m *MetricFloat64) Delta(s Metric, logger *logging.Logger) (int, error) {
	var skips int
	prevRaw := s.GetValuesFloat64()
	prevRecord := s.GetRecords()
	pass := m.GetPass()

	for k := range m.values {
		// reset pass
		pass[k] = true
		if m.record[k] && prevRecord[k] {
			curRaw := m.values[k]
			m.values[k] -= prevRaw[k]
			// Sometimes ONTAP sends spurious zeroes or values less than the previous poll.
			// Detect and don't publish negative deltas or the subsequent poll will show a large spike.
			isInvalidZero := (curRaw == 0 || prevRaw[k] == 0) && m.values[k] != 0
			isNegative := m.values[k] < 0
			if isInvalidZero || isNegative {
				pass[k] = false
				skips++
				logger.Trace().
					Str("metric", m.GetName()).
					Float64("currentRaw", curRaw).
					Float64("previousRaw", prevRaw[k]).
					Str("instIndex", k).
					Msg("Negative cooked value")
			}
		} else {
			// It could be a new or deleted instance
			pass[k] = false
			skips++
			logger.Trace().
				Str("metric", m.GetName()).
				Str("instIndex", k).
				Msg("New or deleted instance")
		}
	}
	return skips, nil
}

func (m *MetricFloat64) Divide(s Metric, logger *logging.Logger) (int, error) {
	var skips int
	sValues := s.GetValuesFloat64()
	sRecord := s.GetRecords()
	pass := m.GetPass()

	for k := range m.values {
		// reset pass
		pass[k] = true
		if m.record[k] && sRecord[k] && sValues[k] != 0 {
			// Don't pass along the value if the numerator or denominator is < 0
			// A denominator of zero is fine
			if m.values[k] < 0 || sValues[k] < 0 {
				pass[k] = false
				skips++
				logger.Trace().
					Str("metric", m.GetName()).
					Float64("numerator", m.values[k]).
					Float64("denominator", sValues[k]).
					Msg("No pass values")
			}
			m.values[k] /= sValues[k]
		} else {
			// It could be a new or deleted instance or a 0 denominator
			pass[k] = false
			skips++
			logger.Trace().
				Str("metric", m.GetName()).
				Str("instIndex", k).
				Msg("New or deleted instance or zero denominator")
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

	for k := range m.values {
		// reset pass
		pass[k] = true
		if m.record[k] && sRecord[k] {
			v := m.values[k]
			// Don't pass along the value if the numerator or denominator is < 0
			// It's important to check sValues[i] < 0 and allow a zero so pass=true and m.values[i] remains unchanged
			if m.values[k] < 0 || sValues[k] < 0 {
				pass[k] = false
				skips++
				logger.Trace().
					Str("metric", m.GetName()).
					Float64("numerator", v).
					Float64("denominator", sValues[k]).
					Msg("Negative values")
			}
			if sValues[k] >= x {
				m.values[k] /= sValues[k]
			}
		} else {
			// It could be a new or deleted instance
			pass[k] = false
			skips++
			logger.Trace().
				Str("metric", m.GetName()).
				Str("instIndex", k).
				Msg("New or deleted instance")
		}
	}
	return skips, nil
}

func (m *MetricFloat64) MultiplyByScalar(s uint, logger *logging.Logger) (int, error) {
	var skips int
	x := float64(s)
	pass := m.GetPass()
	for k := range m.values {
		// reset pass
		pass[k] = true
		if m.record[k] {
			// if current is <= 0
			if m.values[k] < 0 {
				pass[k] = false
				skips++
				logger.Trace().
					Str("metric", m.GetName()).
					Float64("currentRaw", m.values[k]).
					Uint("scalar", s).
					Msg("Negative value")
			}
			m.values[k] *= x
		}
	}
	return skips, nil
}

func (m *MetricFloat64) Print() {
	for k := range m.values {
		if m.record[k] && m.pass[k] {
			fmt.Printf("%s%v%s ", color.Green, m.values[k], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, m.values[k], color.End)
		}
	}
}
