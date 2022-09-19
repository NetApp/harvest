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
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
)

type Metric interface {
	// methods related to metric attributes

	GetName() string
	SetName(string)
	GetType() string
	SetLabel(string, string)
	SetLabels(*dict.Dict)
	GetLabel(string) string
	GetLabels() *dict.Dict
	HasLabels() bool
	IsExportable() bool
	SetExportable(bool)
	GetProperty() string
	SetProperty(string)
	GetComment() string
	SetComment(string)
	IsArray() bool
	SetArray(bool)
	Clone(bool) Metric

	// methods for resizing metric storage

	Reset(int)
	Remove(int)
	Append()

	// methods for writing to metric storage

	SetValueInt64(*Instance, int64) error
	SetValueUint8(*Instance, uint8) error
	SetValueUint64(*Instance, uint64) error
	SetValueFloat64(*Instance, float64) error
	SetValueString(*Instance, string) error
	SetValueBytes(*Instance, []byte) error
	SetValueBool(*Instance, bool) error

	AddValueInt64(*Instance, int64) error
	AddValueUint8(*Instance, uint8) error
	AddValueUint64(*Instance, uint64) error
	AddValueFloat64(*Instance, float64) error
	AddValueString(*Instance, string) error

	SetValueNAN(*Instance)
	// methods for reading from metric storage

	GetValueInt(*Instance) (int, bool, bool)
	GetValueInt64(*Instance) (int64, bool, bool)
	GetValueUint8(*Instance) (uint8, bool, bool)
	GetValueUint64(*Instance) (uint64, bool, bool)
	GetValueFloat64(*Instance) (float64, bool, bool)
	GetValueString(*Instance) (string, bool, bool)
	GetValueBytes(*Instance) ([]byte, bool, bool)
	GetRecords() []bool
	GetPass() []bool
	GetValuesFloat64() []float64

	// methods for doing vector arithmetics
	// currently only supported for float64!

	Delta(Metric, *logging.Logger) (VectorSummary, error)
	Divide(Metric, *logging.Logger) (VectorSummary, error)
	DivideWithThreshold(Metric, int, *logging.Logger) (VectorSummary, error)
	MultiplyByScalar(uint, *logging.Logger) (VectorSummary, error)
	// Print is used for debugging
	Print()
}

type AbstractMetric struct {
	name       string
	dtype      string
	property   string
	comment    string
	array      bool
	exportable bool
	labels     *dict.Dict
	record     []bool
	pass       []bool
}

func (m *AbstractMetric) Clone(deep bool) *AbstractMetric {
	clone := AbstractMetric{
		name:       m.name,
		dtype:      m.dtype,
		property:   m.property,
		comment:    m.comment,
		exportable: m.exportable,
		array:      m.array,
	}
	if m.labels != nil {
		clone.labels = m.labels.Copy()
	}
	if deep {
		if len(m.record) != 0 {
			clone.record = make([]bool, len(m.record))
			copy(clone.record, m.record)
		}
		if len(m.pass) != 0 {
			clone.pass = make([]bool, len(m.pass))
			copy(clone.pass, m.pass)
		}
	}
	return &clone
}

func (m *AbstractMetric) GetName() string {
	return m.name
}

func (m *AbstractMetric) SetName(name string) {
	m.name = name
}

func (m *AbstractMetric) IsExportable() bool {
	return m.exportable
}

func (m *AbstractMetric) SetExportable(b bool) {
	m.exportable = b
}

func (m *AbstractMetric) GetType() string {
	return m.dtype
}

func (m *AbstractMetric) GetProperty() string {
	return m.property
}

func (m *AbstractMetric) SetProperty(p string) {
	m.property = p
}

func (m *AbstractMetric) GetComment() string {
	return m.comment
}

func (m *AbstractMetric) SetComment(c string) {
	m.comment = c
}

func (m *AbstractMetric) IsArray() bool {
	return m.array
}

func (m *AbstractMetric) SetArray(c bool) {
	m.array = c
}

func (m *AbstractMetric) SetLabel(key, value string) {
	if m.labels == nil {
		m.labels = dict.New()
	}
	m.labels.Set(key, value)
}

func (m *AbstractMetric) SetLabels(labels *dict.Dict) {
	m.labels = labels
}

func (m *AbstractMetric) GetLabel(key string) string {
	if m.labels != nil {
		return m.labels.Get(key)
	}
	return ""
}

func (m *AbstractMetric) GetLabels() *dict.Dict {
	return m.labels

}
func (m *AbstractMetric) HasLabels() bool {
	return m.labels != nil && m.labels.Size() != 0
}

func (m *AbstractMetric) GetRecords() []bool {
	return m.record
}

func (m *AbstractMetric) GetPass() []bool {
	return m.pass
}

func (m *AbstractMetric) SetValueNAN(i *Instance) {
	m.record[i.index] = false
}

func (m *AbstractMetric) Delta(Metric, *logging.Logger) (VectorSummary, error) {
	return VectorSummary{}, errs.New(errs.ErrImplement, m.dtype)
}

func (m *AbstractMetric) Divide(Metric, *logging.Logger) (VectorSummary, error) {
	return VectorSummary{}, errs.New(errs.ErrImplement, m.dtype)
}

func (m *AbstractMetric) DivideWithThreshold(Metric, int, *logging.Logger) (VectorSummary, error) {
	return VectorSummary{}, errs.New(errs.ErrImplement, m.dtype)
}

func (m *AbstractMetric) MultiplyByScalar(uint, *logging.Logger) (VectorSummary, error) {
	return VectorSummary{}, errs.New(errs.ErrImplement, m.dtype)
}

func (m *AbstractMetric) AddValueString(*Instance, string) error {
	return errs.New(errs.ErrImplement, m.dtype)
}

func (m *AbstractMetric) SetValueBool(*Instance, bool) error {
	return errs.New(errs.ErrImplement, m.dtype)
}
