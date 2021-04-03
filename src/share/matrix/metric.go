package matrix

import (
	"goharvest2/share/dict"
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
	GetLabel(string) string
	GetLabels() map[string]string
	HasLabels() bool
	IsExportable() bool
	SetExportable(bool)
	// methods for resizing metric storage
	Reset(int)
	Remove(int)
	Append()
	// methods for writing to metric storage
	SetValueInt(*Instance, int) error
	SetValueInt32(*Instance, int32) error
	SetValueInt64(*Instance, int64) error
	SetValueUint32(*Instance, uint32) error
	SetValueUint64(*Instance, uint64) error
	SetValueFloat32(*Instance, float32) error
	SetValueFloat64(*Instance, float64) error
	SetValueString(*Instance, string) error
	SetValueBytes(*Instance, []byte) error
	// methods for reading from metric storage
	GetValueInt32(*Instance) (int32, bool)
	GetValueInt64(*Instance) (int64, bool)
	GetValueUint32(*Instance) (uint32, bool)
	GetValueUint64(*Instance) (uint64, bool)
	GetValueFloat32(*Instance) (float32, bool)
	GetValueFloat64(*Instance) (float64, bool)
	GetValueString(*Instance) (string, bool)
	GetValueBytes(*Instance) ([]byte, bool)
	// debugging
	Print()
}
/*
func NewMetric(name, dtype string) (Metric, error) {

	var (
		metric Metric
		err error
	)

	abm := &AbstractMetric{Name: name, Type: dtype}

	switch dtype {
	case "int32":
		metric = &MetricInt32{AbstractMetric: abm}
	case "int64":
		metric = &MetricInt64{AbstractMetric: abm}
	case "uint32":
		metric = &MetricUint32{AbstractMetric: abm}
	case "uint64":
		metric = &MetricUint64{AbstractMetric: abm}
	case "float32":
		metric = &MetricFloat32{AbstractMetric: abm}
	case "float64":
		metric = &MetricFloat64{AbstractMetric: abm}
	default:
		err = errors.New(INVALID_DTYPE, dtype)
	}
	return metric, err
}
*/

type AbstractMetric struct {
	Name string
	Type string
	Exportable bool
	Labels *dict.Dict
	record []bool
}

func (m *AbstractMetric) GetName() string {
	return m.Name
}

func (m *AbstractMetric) SetName(name string) {
	m.Name = name
}

func (m *AbstractMetric) IsExportable() bool {
	return m.Exportable
}

func (m *AbstractMetric) SetExportable(b bool) {
	m.Exportable = b
}

func (m *AbstractMetric) GetType() string {
	return m.Type
}

func (m *AbstractMetric) SetLabel(key, value string) {
	if m.Labels == nil {
		m.Labels = dict.New()
	}
	m.Labels.Set(key, value)
}

func (m *AbstractMetric) GetLabel(key string) string {
	if m.Labels != nil {
		return m.Labels.Get(key)
	}
	return ""
}

func (m *AbstractMetric) GetLabels() map[string]string {
	var labels map[string]string
	if m.HasLabels() {
		labels = m.Labels.Iter()
	}
	return labels
}

func (m *AbstractMetric) HasLabels() bool {
	return m.Labels != nil
}