package matrix

import (
	"goharvest2/share/dict"
	"goharvest2/share/errors"
)

type Metric interface {
	// methods related to metric attributes
	// @TODO, add methods for (conveniency for some collectors)
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
	// methods for reading from metric storage
	GetValueInt32(*Instance) (int32, bool)
	GetValueInt64(*Instance) (int64, bool)
	GetValueUint8(*Instance) (uint8, bool)
	GetValueUint32(*Instance) (uint32, bool)
	GetValueUint64(*Instance) (uint64, bool)
	GetValueFloat32(*Instance) (float32, bool)
	GetValueFloat64(*Instance) (float64, bool)
	GetValueString(*Instance) (string, bool)
	GetValueBytes(*Instance) ([]byte, bool)
	// methods for doing vector arithmetics
	// currently only supported for float64!
	GetRecords() []bool
	GetValuesFloat64() []float64
	Delta(Metric) error
	Divide(Metric) error
	DivideWithThreshold(Metric, int) error
	MultiplyByScalar(int) error
	// debugging
	Print()
}

type AbstractMetric struct {
	name string
	dtype string
	property string
	comment string
	exportable bool
	labels *dict.Dict
	record []bool
}

func (my *AbstractMetric) Clone(deep bool) *AbstractMetric {
	clone := AbstractMetric{
		name: my.name,
		dtype: my.dtype,
		property: my.property,
		comment: my.comment,
		exportable: my.exportable,
	}
	if deep {
		if my.labels != nil {
			clone.labels = my.labels.Copy()
		}
		if len(my.record) != 0 {
			clone.record = make([]bool, len(my.record))
			for i,v := range my.record {
				clone.record[i] = v
			}
		}
	}
	return &clone
}

func (my *AbstractMetric) GetName() string {
	return my.name
}

func (my *AbstractMetric) SetName(name string) {
	my.name = name
}

func (my *AbstractMetric) IsExportable() bool {
	return my.exportable
}

func (my *AbstractMetric) SetExportable(b bool) {
	my.exportable = b
}

func (my *AbstractMetric) GetType() string {
	return my.dtype
}

func (my *AbstractMetric) GetProperty() string {
	return my.property
}

func (my *AbstractMetric) SetProperty(p string) {
	my.property = p
}

func (my *AbstractMetric) GetComment() string {
	return my.comment
}

func (my *AbstractMetric) SetComment(c string) {
	my.comment = c
}

func (my *AbstractMetric) SetLabel(key, value string) {
	if my.labels == nil {
		my.labels = dict.New()
	}
	my.labels.Set(key, value)
}

func (my *AbstractMetric) SetLabels(labels *dict.Dict) {
	my.labels = labels
}

func (my *AbstractMetric) GetLabel(key string) string {
	if my.labels != nil {
		return my.labels.Get(key)
	}
	return ""
}

func (my *AbstractMetric) GetLabels() *dict.Dict {
	return my.labels

}
func (my *AbstractMetric) HasLabels() bool {
	return my.labels != nil && my.labels.Size() != 0
}

func (my *AbstractMetric) GetRecords() []bool {
	return my.record
}

func (my *AbstractMetric) Delta(s Metric) error {
	return errors.New(errors.ERR_IMPLEMENT, my.dtype)
}

func (my *AbstractMetric) Divide(s Metric) error {
	return errors.New(errors.ERR_IMPLEMENT, my.dtype)
}

func (my *AbstractMetric) DivideWithThreshold(s Metric, t int) error {
	return errors.New(errors.ERR_IMPLEMENT, my.dtype)
}

func (my *AbstractMetric) MultiplyByScalar(s int) error {
	return errors.New(errors.ERR_IMPLEMENT, my.dtype)
}

