package matrix

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
)

func setUpMatrix() *Matrix {
	m := New("TestRemoveInstance", "test", "test")
	speed, _ := m.NewMetricUint64("max_speed")
	instanceNames := []string{"A", "B", "C", "D"}
	for _, instanceName := range instanceNames {
		instance, _ := m.NewInstance(instanceName)
		speed.SetValueInt64(instance, 10)
	}
	return m
}

func TestMatrix_RemoveInstance(t *testing.T) {

	type args struct {
		key string
	}

	type test struct {
		name             string
		args             args
		maxInstanceIndex int
	}

	tests := []test{
		{"removeExistingKey", args{key: "A"}, 2},
		{"removeAbsentKey", args{key: "E"}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpMatrix()
			m.RemoveInstance(tt.args.key)
			maxIndex := 0
			for _, i := range m.GetInstances() {
				if i.index > maxIndex {
					maxIndex = i.index
				}
			}
			if maxIndex != tt.maxInstanceIndex {
				assert.Equal(t, maxIndex, tt.maxInstanceIndex)
			}
		})
	}
}

func TestMatrix_RemoveMetric_ClearsDisplayMetrics(t *testing.T) {
	m := New("TestRemoveMetric", "test", "test")
	_, err := m.NewMetricFloat64("speed", "Speed")
	assert.Nil(t, err)
	assert.NotNil(t, m.DisplayMetric("Speed"))

	m.RemoveMetric("speed")
	assert.Nil(t, m.DisplayMetric("Speed"))
	assert.Equal(t, m.DisplayMetricKey("Speed"), "")

	// Reusing the same display name for a new metric must not resolve to the old key.
	_, err = m.NewMetricFloat64("speed2", "Speed")
	assert.Nil(t, err)
	got := m.DisplayMetric("Speed")
	assert.NotNil(t, got)
	assert.Equal(t, m.DisplayMetricKey("Speed"), "speed2")
}

func TestMatrix_PurgeMetrics_ClearsDisplayMetrics(t *testing.T) {
	m := New("TestPurgeMetrics", "test", "test")
	_, err := m.NewMetricFloat64("speed", "Speed")
	assert.Nil(t, err)

	m.PurgeMetrics()

	assert.Equal(t, len(m.GetMetrics()), 0)
	assert.Nil(t, m.DisplayMetric("Speed"))
	assert.Equal(t, m.DisplayMetricKey("Speed"), "")
}

func TestMatrix_CloneProfiles(t *testing.T) {
	m := New("TestCloneProfiles", "test", "identifier")
	m.SetGlobalLabel("cluster", "source")
	metric, err := m.NewMetricFloat64("speed", "Speed")
	assert.Nil(t, err)
	instance, err := m.NewInstance("instance")
	assert.Nil(t, err)
	instance.SetLabel("node", "node1")
	instance.SetLabel("uuid", "uuid1")
	instance.SetExportable(false)
	instance.SetPartial(true)
	metric.SetValueFloat64(instance, 42)

	template := m.CloneMetricTemplate()
	assert.Equal(t, len(template.GetMetrics()), 1)
	assert.Equal(t, len(template.GetInstances()), 0)
	template.SetGlobalLabel("cluster", "template")
	assert.Equal(t, m.GetGlobalLabels()["cluster"], "source")

	collection := m.CloneForCollection()
	collectionInstance := collection.GetInstance("instance")
	assert.NotNil(t, collectionInstance)
	assert.False(t, collectionInstance.IsPartial())
	assert.Equal(t, len(collection.GetMetric("speed").record), 0)

	snapshot := m.Clone()
	snapshotInstance := snapshot.GetInstance("instance")
	assert.NotNil(t, snapshotInstance)
	assert.False(t, snapshotInstance.IsExportable())
	assert.True(t, snapshotInstance.IsPartial())
	value, ok := snapshot.GetMetric("speed").GetValueFloat64(snapshotInstance)
	assert.True(t, ok)
	assert.Equal(t, value, 42.0)

	selected := m.CloneSelected([]string{"speed"}, []string{"uuid"})
	selectedInstance := selected.GetInstance("instance")
	assert.NotNil(t, selected.GetMetric("speed"))
	assert.Equal(t, selectedInstance.GetLabel("uuid"), "uuid1")
	assert.Equal(t, selectedInstance.GetLabel("node"), "")
	assert.False(t, selectedInstance.IsExportable())

	empty := m.CloneEmpty()
	assert.Equal(t, len(empty.GetMetrics()), 0)
	assert.Equal(t, len(empty.GetInstances()), 0)
}

func TestMatrix_GetOrCreateMetric(t *testing.T) {
	m := New("TestGetOrCreateMetric", "test", "test")

	created, err := m.GetOrCreateMetric("speed")
	assert.Nil(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, created.GetType(), "float64")

	again, err := m.GetOrCreateMetric("speed")
	assert.Nil(t, err)
	assert.Equal(t, again, created)

	typed, err := m.GetOrCreateMetric("count", "int64")
	assert.Nil(t, err)
	assert.Equal(t, typed.GetType(), "int64")
}

func TestMatrix_NewMetricsFloat64(t *testing.T) {
	m := New("TestNewMetricsFloat64", "test", "test")

	err := m.NewMetricsFloat64("a", "b", "c")
	assert.Nil(t, err)
	for _, key := range []string{"a", "b", "c"} {
		assert.NotNil(t, m.GetMetric(key))
	}

	// Calling again is idempotent and must not error.
	err = m.NewMetricsFloat64("a", "b", "c")
	assert.Nil(t, err)
}

func TestMatrix_GetOrCreateInstance(t *testing.T) {
	m := New("TestGetOrCreateInstance", "test", "test")

	instance, created := m.GetOrCreateInstance("i1")
	assert.NotNil(t, instance)
	assert.True(t, created)

	again, created2 := m.GetOrCreateInstance("i1")
	assert.Equal(t, again, instance)
	assert.False(t, created2)
}

func TestMatrix_MustSetValueX_Wrappers(t *testing.T) {
	m := New("TestMustSetValueXWrappers", "test", "test")
	instance, _ := m.NewInstance("i1")

	countMetric, err := m.NewMetricInt64("count")
	assert.Nil(t, err)
	m.MustSetValueInt64("count", instance, 5)
	v, ok := countMetric.GetValueInt64(instance)
	assert.True(t, ok)
	assert.Equal(t, int64(5), v)
	assert.Panics(t, func() { m.MustSetValueInt64("missing", instance, 5) })

	u8Metric, err := m.NewMetricUint8("u8")
	assert.Nil(t, err)
	m.MustSetValueUint8("u8", instance, 5)
	u8, ok := u8Metric.GetValueUint8(instance)
	assert.True(t, ok)
	assert.Equal(t, uint8(5), u8)
	assert.Panics(t, func() { m.MustSetValueUint8("missing", instance, 5) })

	u64Metric, err := m.NewMetricUint64("u64")
	assert.Nil(t, err)
	m.MustSetValueUint64("u64", instance, 5)
	u64, ok := u64Metric.GetValueUint64(instance)
	assert.True(t, ok)
	assert.Equal(t, uint64(5), u64)
	assert.Panics(t, func() { m.MustSetValueUint64("missing", instance, 5) })

	sMetric, err := m.NewMetricFloat64("s")
	assert.Nil(t, err)
	assert.Nil(t, m.MustGetMetric("s").SetValueString(instance, "1.5"))
	assert.Equal(t, sMetric, m.MustGetMetric("s"))
	assert.Panics(t, func() { m.MustGetMetric("missing") })
}

func TestNewExportOptions(t *testing.T) {
	n := NewExportOptions("svm", "volume")
	instanceKeys := n.GetChildS("instance_keys")
	assert.NotNil(t, instanceKeys)
	children := instanceKeys.GetChildren()
	assert.Equal(t, len(children), 2)
	assert.Equal(t, children[0].GetContentS(), "svm")
	assert.Equal(t, children[1].GetContentS(), "volume")
}

func TestMatrix_DefaultExportOptionsAreOwned(t *testing.T) {
	m := New("TestExportOptions", "test", "test")
	m.GetExportOptions().PopChildS("include_all_labels")
	assert.Nil(t, m.GetExportOptions().GetChildS("include_all_labels"))

	clone := m.CloneEmpty()
	clone.GetExportOptions().NewChildS("instance_keys", "")
	assert.Nil(t, m.GetExportOptions().GetChildS("instance_keys"))
}
