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
	m.skip = make(map[string]bool)
	m.values = make(map[string]float64, size)
}

// Remove element at index, shift everything to the left
func (m *MetricFloat64) Remove(key string) {
	delete(m.skip, key)
	delete(m.values, key)
}

// Write methods

func (m *MetricFloat64) SetValueInt64(i *Instance, v int64) error {
	delete(m.skip, i.key)
	m.values[i.key] = float64(v)
	return nil
}

func (m *MetricFloat64) SetValueUint8(i *Instance, v uint8) error {
	delete(m.skip, i.key)
	m.values[i.key] = float64(v)
	return nil
}

func (m *MetricFloat64) SetValueUint64(i *Instance, v uint64) error {
	delete(m.skip, i.key)
	m.values[i.key] = float64(v)
	return nil
}

func (m *MetricFloat64) SetValueFloat64(i *Instance, v float64) error {
	delete(m.skip, i.key)
	m.values[i.key] = v
	return nil
}

func (m *MetricFloat64) SetValueString(i *Instance, v string) error {
	var x float64
	var err error
	if x, err = strconv.ParseFloat(v, 64); err == nil {
		delete(m.skip, i.key)
		m.values[i.key] = x
		return nil
	}
	m.skip[i.key] = true
	return err
}

func (m *MetricFloat64) SetValueBytes(i *Instance, v []byte) error {
	return m.SetValueString(i, string(v))
}

func (m *MetricFloat64) AddValueInt64(i *Instance, n int64) error {
	v, _ := m.GetValueInt64(i)
	return m.SetValueInt64(i, v+n)
}

func (m *MetricFloat64) AddValueUint8(i *Instance, n uint8) error {
	v, _ := m.GetValueUint8(i)
	return m.SetValueUint8(i, v+n)
}

func (m *MetricFloat64) AddValueUint64(i *Instance, n uint64) error {
	v, _ := m.GetValueUint64(i)
	return m.SetValueUint64(i, v+n)
}

func (m *MetricFloat64) AddValueFloat64(i *Instance, n float64) error {
	v, _ := m.GetValueFloat64(i)
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
	if n, has = m.GetValueFloat64(i); has {
		return m.SetValueFloat64(i, x+n)
	}
	return m.SetValueFloat64(i, x)
}

// Read methods

func (m *MetricFloat64) GetValueInt(i *Instance) (int, bool) {
	// check if key exists in value map
	val, has := m.values[i.key]
	return int(val), has && !m.skip[i.key]
}

func (m *MetricFloat64) GetValueInt64(i *Instance) (int64, bool) {
	// check if key exists in value map
	val, has := m.values[i.key]
	return int64(val), has && !m.skip[i.key]
}

func (m *MetricFloat64) GetValueUint8(i *Instance) (uint8, bool) {
	// check if key exists in value map
	val, has := m.values[i.key]
	return uint8(val), has && !m.skip[i.key]
}

func (m *MetricFloat64) GetValueUint64(i *Instance) (uint64, bool) {
	// check if key exists in value map
	val, has := m.values[i.key]
	return uint64(val), has && !m.skip[i.key]
}

func (m *MetricFloat64) GetValueFloat64(i *Instance) (float64, bool) {
	// check if key exists in value map
	val, has := m.values[i.key]
	return val, has && !m.skip[i.key]
}

func (m *MetricFloat64) GetValueString(i *Instance) (string, bool) {
	// check if key exists in value map
	val, has := m.values[i.key]
	return strconv.FormatFloat(val, 'f', -1, 64), has && !m.skip[i.key]
}

func (m *MetricFloat64) GetValueBytes(i *Instance) ([]byte, bool) {
	s, ok := m.GetValueString(i)
	return []byte(s), ok
}

// vector arithmetics

func (m *MetricFloat64) GetValuesFloat64() map[string]float64 {
	return m.values
}

func (m *MetricFloat64) Delta(s Metric, logger *logging.Logger) (int, error) {
	var skips int
	prevRaw := s.GetValuesFloat64()
	prevSkips := s.GetSkips()

	for k, v := range m.values {

		if !m.skip[k] && !prevSkips[k] {
			curRaw := v
			m.values[k] -= prevRaw[k]
			// Sometimes ONTAP sends spurious zeroes or values less than the previous poll.
			// Detect and don't publish negative deltas or the subsequent poll will show a large spike.
			isInvalidZero := (curRaw == 0 || prevRaw[k] == 0) && m.values[k] != 0
			isNegative := m.values[k] < 0
			if isInvalidZero || isNegative {
				m.skip[k] = true
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
			m.skip[k] = true
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
	sSkip := s.GetSkips()

	for k, v := range m.values {
		if !m.skip[k] && !sSkip[k] {
			if sValues[k] != 0 {
				// Don't pass along the value if the numerator or denominator is < 0
				// A denominator of zero is fine
				if v < 0 || sValues[k] < 0 {
					m.skip[k] = true
					skips++
					logger.Trace().
						Str("metric", m.GetName()).
						Float64("numerator", m.values[k]).
						Float64("denominator", sValues[k]).
						Msg("No pass values")
				}
				m.values[k] /= sValues[k]
			} else {
				m.values[k] = 0
			}
		} else {
			// It could be a new or deleted instance or a 0 denominator
			m.skip[k] = true
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
	sSkip := s.GetSkips()

	for k, v := range m.values {
		if !m.skip[k] && !sSkip[k] {
			// Don't pass along the value if the numerator or denominator is < 0
			// It's important to check sValues[i] < 0 and allow a zero so pass=true and m.values[i] remains unchanged
			if v < 0 || sValues[k] < 0 {
				m.skip[k] = true
				skips++
				logger.Trace().
					Str("metric", m.GetName()).
					Float64("numerator", v).
					Float64("denominator", sValues[k]).
					Msg("Negative values")
			}
			if sValues[k] >= x {
				m.values[k] /= sValues[k]
			} else {
				m.values[k] = 0
			}
		} else {
			// It could be a new or deleted instance
			m.skip[k] = true
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
	for k, v := range m.values {
		if !m.skip[k] {
			// if current is <= 0
			if v < 0 {
				m.skip[k] = true
				skips++
				logger.Trace().
					Str("metric", m.GetName()).
					Float64("currentRaw", v).
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
		if !m.skip[k] {
			fmt.Printf("%s%v%s ", color.Green, m.values[k], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, m.values[k], color.End)
		}
	}
}
