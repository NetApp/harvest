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
	// @TODO, add methods for (convenience of collectors)
	// Property
	// BaseCounter

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
	SetValueInt(*Instance, int) error
	SetValueInt32(*Instance, int32) error
	SetValueInt64(*Instance, int64) error
	SetValueUint8(*Instance, uint8) error
	SetValueUint32(*Instance, uint32) error
	SetValueUint64(*Instance, uint64) error
	SetValueFloat32(*Instance, float32) error
	SetValueFloat64(*Instance, float64) error
	SetValueString(*Instance, string) error
	SetValueBytes(*Instance, []byte) error
	SetValueBool(*Instance, bool) error

	AddValueInt(*Instance, int) error
	AddValueInt32(*Instance, int32) error
	AddValueInt64(*Instance, int64) error
	AddValueUint8(*Instance, uint8) error
	AddValueUint32(*Instance, uint32) error
	AddValueUint64(*Instance, uint64) error
	AddValueFloat32(*Instance, float32) error
	AddValueFloat64(*Instance, float64) error
	AddValueString(*Instance, string) error
	//SetValueBytes(*Instance, []byte) error

	SetValueNAN(*Instance)
	// methods for reading from metric storage
	GetValueInt(*Instance) (int, bool, bool)
	GetValueInt32(*Instance) (int32, bool, bool)
	GetValueInt64(*Instance) (int64, bool, bool)
	GetValueUint8(*Instance) (uint8, bool, bool)
	GetValueUint32(*Instance) (uint32, bool, bool)
	GetValueUint64(*Instance) (uint64, bool, bool)
	GetValueFloat32(*Instance) (float32, bool, bool)
	GetValueFloat64(*Instance) (float64, bool, bool)
	GetValueString(*Instance) (string, bool, bool)
	GetValueBytes(*Instance) ([]byte, bool, bool)
	// methods for doing vector arithmetics
	// currently only supported for float64!
	GetRecords() []bool
	GetSkips() []bool
	GetValuesFloat64() []float64
	Delta(Metric, *logging.Logger) (VectorSummary, error)
	Divide(Metric, *logging.Logger) (VectorSummary, error)
	DivideWithThreshold(Metric, int, *logging.Logger) (VectorSummary, error)
	MultiplyByScalar(int, *logging.Logger) (VectorSummary, error)
	// debugging
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
	skip       []bool
}

func (me *AbstractMetric) Clone(deep bool) *AbstractMetric {
	clone := AbstractMetric{
		name:       me.name,
		dtype:      me.dtype,
		property:   me.property,
		comment:    me.comment,
		exportable: me.exportable,
		array:      me.array,
	}
	if me.labels != nil {
		clone.labels = me.labels.Copy()
	}
	if deep {
		if len(me.record) != 0 {
			clone.record = make([]bool, len(me.record))
			copy(clone.record, me.record)
		}
		if len(me.skip) != 0 {
			clone.skip = make([]bool, len(me.skip))
			copy(clone.skip, me.skip)
		}
	}
	return &clone
}

func (me *AbstractMetric) GetName() string {
	return me.name
}

func (me *AbstractMetric) SetName(name string) {
	me.name = name
}

func (me *AbstractMetric) IsExportable() bool {
	return me.exportable
}

func (me *AbstractMetric) SetExportable(b bool) {
	me.exportable = b
}

func (me *AbstractMetric) GetType() string {
	return me.dtype
}

func (me *AbstractMetric) GetProperty() string {
	return me.property
}

func (me *AbstractMetric) SetProperty(p string) {
	me.property = p
}

func (me *AbstractMetric) GetComment() string {
	return me.comment
}

func (me *AbstractMetric) SetComment(c string) {
	me.comment = c
}

func (me *AbstractMetric) IsArray() bool {
	return me.array
}

func (me *AbstractMetric) SetArray(c bool) {
	me.array = c
}

func (me *AbstractMetric) SetLabel(key, value string) {
	if me.labels == nil {
		me.labels = dict.New()
	}
	me.labels.Set(key, value)
}

func (me *AbstractMetric) SetLabels(labels *dict.Dict) {
	me.labels = labels
}

func (me *AbstractMetric) GetLabel(key string) string {
	if me.labels != nil {
		return me.labels.Get(key)
	}
	return ""
}

func (me *AbstractMetric) GetLabels() *dict.Dict {
	return me.labels

}
func (me *AbstractMetric) HasLabels() bool {
	return me.labels != nil && me.labels.Size() != 0
}

func (me *AbstractMetric) GetRecords() []bool {
	return me.record
}

func (me *AbstractMetric) GetSkips() []bool {
	return me.skip
}

func (me *AbstractMetric) SetValueNAN(i *Instance) {
	me.record[i.index] = false
}

func (me *AbstractMetric) Delta(Metric, *logging.Logger) (VectorSummary, error) {
	return VectorSummary{}, errs.New(errs.ErrImplement, me.dtype)
}

func (me *AbstractMetric) Divide(Metric, *logging.Logger) (VectorSummary, error) {
	return VectorSummary{}, errs.New(errs.ErrImplement, me.dtype)
}

func (me *AbstractMetric) DivideWithThreshold(Metric, int, *logging.Logger) (VectorSummary, error) {
	return VectorSummary{}, errs.New(errs.ErrImplement, me.dtype)
}

func (me *AbstractMetric) MultiplyByScalar(int, *logging.Logger) (VectorSummary, error) {
	return VectorSummary{}, errs.New(errs.ErrImplement, me.dtype)
}

func (me *AbstractMetric) AddValueString(*Instance, string) error {
	return errs.New(errs.ErrImplement, me.dtype)
}

func (me *AbstractMetric) SetValueBool(*Instance, bool) error {
	return errs.New(errs.ErrImplement, me.dtype)
}
