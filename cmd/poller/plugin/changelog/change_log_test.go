package changelog

import (
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
)

func newChangeLog() *ChangeLog {

	params := node.NewS("ChangeLog")
	parentParams := node.NewS("parent")
	parentParams.NewChildS("object", "svm")

	abc := plugin.New("Test", nil, params, parentParams, "", nil)
	p := &ChangeLog{AbstractPlugin: abc}
	p.Options = &options.Options{
		Poller: "Test",
	}
	p.Object = "svm"

	if err := p.Init(); err != nil {
		panic(err)
	}
	return p
}

func TestChangeLogModified(t *testing.T) {
	p := newChangeLog()
	m := matrix.New("TestChangeLog", "svm", "svm")
	data := map[string]*matrix.Matrix{
		"svm": m,
	}
	instance, _ := m.NewInstance("0")
	instance.SetLabel("uuid", "u1")
	instance.SetLabel("svm", "s1")
	instance.SetLabel("type", "t1")

	_, _ = p.Run(data)

	m1 := matrix.New("TestChangeLog", "svm", "svm")
	data1 := map[string]*matrix.Matrix{
		"svm": m1,
	}
	instance1, _ := m1.NewInstance("0")
	instance1.SetLabel("uuid", "u1")
	instance1.SetLabel("svm", "s2")
	instance1.SetLabel("type", "t2")

	o, _ := p.Run(data1)

	if len(o) == 1 {
		cl := o[0]
		if len(cl.GetInstances()) != 2 {
			t.Errorf("ChangeLog instances size expected %d, actual %d", 2, len(cl.GetInstances()))
		} else {
			for _, i := range cl.GetInstances() {
				if i.GetLabel(changeTypeLabel) != modify {
					t.Errorf("ChangeLog %s label expected %s, actual %s", changeTypeLabel, modify, i.GetLabel(changeTypeLabel))
				}
			}
		}
	} else {
		t.Error("ChangeLog slice size is wrong")
	}
}

func TestChangeLogCreated(t *testing.T) {
	p := newChangeLog()
	m := matrix.New("TestChangeLog", "svm", "svm")
	data := map[string]*matrix.Matrix{
		"svm": m,
	}
	instance, _ := m.NewInstance("0")
	instance.SetLabel("uuid", "u1")
	instance.SetLabel("svm", "s1")
	instance.SetLabel("type", "t1")

	_, _ = p.Run(data)

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

	o, _ := p.Run(data1)

	if len(o) == 1 {
		cl := o[0]
		if len(cl.GetInstances()) != 1 {
			t.Errorf("ChangeLog instances size expected %d, actual %d", 1, len(cl.GetInstances()))
		} else {
			for _, i := range cl.GetInstances() {
				if i.GetLabel(changeTypeLabel) != create {
					t.Errorf("ChangeLog %s label expected %s, actual %s", changeTypeLabel, create, i.GetLabel(changeTypeLabel))
				}
			}
		}
	} else {
		t.Error("ChangeLog slice size is wrong")
	}
}

func TestChangeLogDeleted(t *testing.T) {
	p := newChangeLog()
	m := matrix.New("TestChangeLog", "svm", "svm")
	data := map[string]*matrix.Matrix{
		"svm": m,
	}
	instance, _ := m.NewInstance("0")
	instance.SetLabel("uuid", "u1")
	instance.SetLabel("svm", "s1")
	instance.SetLabel("type", "t1")

	_, _ = p.Run(data)

	m1 := matrix.New("TestChangeLog", "svm", "svm")
	data1 := map[string]*matrix.Matrix{
		"svm": m1,
	}

	o, _ := p.Run(data1)

	if len(o) == 1 {
		cl := o[0]
		if len(cl.GetInstances()) != 1 {
			t.Errorf("ChangeLog instances size expected %d, actual %d", 1, len(cl.GetInstances()))
		} else {
			for _, i := range cl.GetInstances() {
				if i.GetLabel(changeTypeLabel) != del {
					t.Errorf("ChangeLog %s label expected %s, actual %s", changeTypeLabel, del, i.GetLabel(changeTypeLabel))
				}
			}
		}
	} else {
		t.Error("ChangeLog slice size is wrong")
	}
}

func TestChangeLogUnsupported(t *testing.T) {
	p := newChangeLog()
	m := matrix.New("TestChangeLog", "lun", "lun")
	data := map[string]*matrix.Matrix{
		"svm": m,
	}
	instance, _ := m.NewInstance("0")
	instance.SetLabel("uuid", "u1")
	instance.SetLabel("lun", "l1")

	_, _ = p.Run(data)

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

	o, _ := p.Run(data1)

	if len(o) != 0 {
		t.Errorf("ChangeLog matric size expected %d, actual %d", 0, len(o))
	}
}
