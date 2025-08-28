package changelog

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
)

func newChangeLog(object string, includeAll bool) *ChangeLog {
	params := node.NewS("ChangeLog")
	parentParams := node.NewS("parent")
	parentParams.NewChildS("object", object)
	exportOptions := parentParams.NewChildS("export_options", "")
	if includeAll {
		exportOptions.NewChildS("include_all_labels", "true")
	} else {
		instanceKeys := exportOptions.NewChildS("instance_keys", "")
		instanceKeys.NewChildS("", "svm")
	}

	return createChangeLog(params, parentParams)
}

func createChangeLog(params, parentParams *node.Node) *ChangeLog {
	abc := plugin.New("Test", nil, params, parentParams, "", nil)
	p := &ChangeLog{AbstractPlugin: abc}
	p.Options = &options.Options{
		Poller: "Test",
	}
	p.Object = "svm"

	if err := p.Init(conf.Remote{}); err != nil {
		panic(err)
	}
	return p
}

func newChangeLogUnsupportedTrack(object string) *ChangeLog {
	params := node.NewS("ChangeLog")
	t := params.NewChildS("track", "")
	t.NewChildS("", "abcd")
	parentParams := node.NewS("parent")
	parentParams.NewChildS("object", object)

	return createChangeLog(params, parentParams)
}

func newChangeLogMetricTrack(object string) *ChangeLog {
	params := node.NewS("ChangeLog")
	t := params.NewChildS("track", "")
	t.NewChildS("", "svm")
	t.NewChildS("", "type")
	t.NewChildS("", "cpu_usage")
	t1 := params.NewChildS("publish_labels", "")
	t1.NewChildS("", "svm")
	parentParams := node.NewS("parent")
	parentParams.NewChildS("object", object)

	return createChangeLog(params, parentParams)
}

func checkChangeLogInstances(t *testing.T, o []*matrix.Matrix, expectedInstances, expectedLabels int, expectedOpLabel, opLabel string) {
	assert.Equal(t, len(o), 1)
	cl := o[0]
	assert.Equal(t, len(cl.GetInstances()), expectedInstances)
	for _, i := range cl.GetInstances() {
		assert.Equal(t, i.GetLabel(opLabel), expectedOpLabel)
		assert.Equal(t, len(i.GetLabels()), expectedLabels)
	}
}

func TestChangeLogModified(t *testing.T) {
	p := newChangeLog("svm", true)
	m := matrix.New("TestChangeLog", "svm", "svm")
	data := map[string]*matrix.Matrix{
		"svm": m,
	}
	instance, _ := m.NewInstance("0")
	instance.SetLabel("uuid", "u1")
	instance.SetLabel("svm", "s1")
	instance.SetLabel("type", "t1")

	_, _, _ = p.Run(data)

	m1 := matrix.New("TestChangeLog", "svm", "svm")
	data1 := map[string]*matrix.Matrix{
		"svm": m1,
	}
	instance1, _ := m1.NewInstance("0")
	instance1.SetLabel("uuid", "u1")
	instance1.SetLabel("svm", "s2")
	instance1.SetLabel("type", "t2")

	o, _, _ := p.Run(data1)

	checkChangeLogInstances(t, o, 2, 10, update, opLabel)
}

func TestChangeLogModifiedWithMetrics(t *testing.T) {
	p := newChangeLogMetricTrack("svm")
	m := matrix.New("TestChangeLog", "svm", "svm")
	data := map[string]*matrix.Matrix{
		"svm": m,
	}
	instance, _ := m.NewInstance("0")
	instance.SetLabel("uuid", "u1")
	instance.SetLabel("svm", "s1")
	instance.SetLabel("type", "t1")

	// Add a metric to the instance
	metric, _ := m.NewMetricFloat64("cpu_usage")
	metric.SetValueFloat64(instance, 10.0)

	_, _, _ = p.Run(data)

	m1 := matrix.New("TestChangeLog", "svm", "svm")
	data1 := map[string]*matrix.Matrix{
		"svm": m1,
	}
	instance1, _ := m1.NewInstance("0")
	instance1.SetLabel("uuid", "u1")
	instance1.SetLabel("svm", "s2")
	instance1.SetLabel("type", "t2")

	// Modify the metric value
	metric1, _ := m1.NewMetricFloat64("cpu_usage")
	metric1.SetValueFloat64(instance1, 20.0)

	o, _, _ := p.Run(data1)

	checkChangeLogInstances(t, o, 3, 8, update, opLabel)
}

func TestChangeLogCreated(t *testing.T) {
	p := newChangeLog("svm", false)
	m := matrix.New("TestChangeLog", "svm", "svm")
	data := map[string]*matrix.Matrix{
		"svm": m,
	}
	instance, _ := m.NewInstance("0")
	instance.SetLabel("uuid", "u1")
	instance.SetLabel("svm", "s1")
	instance.SetLabel("type", "t1")

	_, _, _ = p.Run(data)

	m1 := matrix.New("TestChangeLog", "svm", "svm")
	data1 := map[string]*matrix.Matrix{
		"svm": m1,
	}
	instance1, _ := m1.NewInstance("1")
	instance1.SetLabel("uuid", "u2")
	instance1.SetLabel("svm", "s2")
	instance1.SetLabel("type", "t2")

	instance2, _ := m1.NewInstance("0")
	instance2.SetLabel("uuid", "u1")
	instance2.SetLabel("svm", "s1")
	instance2.SetLabel("type", "t1")

	o, _, _ := p.Run(data1)

	checkChangeLogInstances(t, o, 1, 4, create, opLabel)
}

func TestChangeLogDeleted(t *testing.T) {
	p := newChangeLog("svm", false)
	m := matrix.New("TestChangeLog", "svm", "svm")
	data := map[string]*matrix.Matrix{
		"svm": m,
	}
	instance, _ := m.NewInstance("0")
	instance.SetLabel("uuid", "u1")
	instance.SetLabel("svm", "s1")
	instance.SetLabel("type", "t1")

	_, _, _ = p.Run(data)

	m1 := matrix.New("TestChangeLog", "svm", "svm")
	data1 := map[string]*matrix.Matrix{
		"svm": m1,
	}

	o, _, _ := p.Run(data1)

	checkChangeLogInstances(t, o, 1, 4, del, opLabel)
}

func TestChangeLogUnsupported(t *testing.T) {
	p := newChangeLog("lun", false)
	m := matrix.New("TestChangeLog", "lun", "lun")
	data := map[string]*matrix.Matrix{
		"svm": m,
	}
	instance, _ := m.NewInstance("0")
	instance.SetLabel("uuid", "u1")
	instance.SetLabel("lun", "l1")

	_, _, _ = p.Run(data)

	m1 := matrix.New("TestChangeLog", "lun", "lun")
	data1 := map[string]*matrix.Matrix{
		"svm": m1,
	}
	instance1, _ := m1.NewInstance("1")
	instance1.SetLabel("uuid", "u2")
	instance1.SetLabel("lun", "l2")

	instance2, _ := m1.NewInstance("0")
	instance2.SetLabel("uuid", "u1")
	instance2.SetLabel("lun", "l3")

	o, _, _ := p.Run(data1)

	assert.Equal(t, len(o), 0)
}

func TestChangeLogModifiedUnsupportedTrack(t *testing.T) {
	p := newChangeLogUnsupportedTrack("svm")

	m := matrix.New("TestChangeLog", "svm", "svm")
	data := map[string]*matrix.Matrix{
		"svm": m,
	}
	instance, _ := m.NewInstance("0")
	instance.SetLabel("uuid", "u1")
	instance.SetLabel("svm", "s1")

	_, _, _ = p.Run(data)

	m1 := matrix.New("TestChangeLog", "svm", "svm")
	data1 := map[string]*matrix.Matrix{
		"svm": m1,
	}

	instance1, _ := m1.NewInstance("0")
	instance1.SetLabel("uuid", "u1")
	instance1.SetLabel("svm", "s2")

	o, _, _ := p.Run(data1)

	checkChangeLogInstances(t, o, 0, 0, "", "")
}
