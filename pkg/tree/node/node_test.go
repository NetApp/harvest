package node

import (
	"github.com/netapp/harvest/v2/assert"
	"testing"
)

func Test_simpleName(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{name: "empty", s: "", want: ""},
		{name: "simple", s: "abc", want: "abc"},
		{name: "simple 1 space", s: "abc ", want: "abc"},
		{name: "simple more space", s: "abc      ", want: "abc"},
		{name: "simple prefix", s: "^^abc  => asdf ", want: "abc"},
		{name: "space prefix", s: "   abc  => asdf ", want: "abc"},
		{name: "dashed", s: "cpu-busy => zig", want: "cpu-busy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, simpleName(tt.s), tt.want)
		})
	}
}

func TestNode_FlatList(t *testing.T) {
	tests := []struct {
		name  string
		tree  *Node
		want  string
		count int
	}{
		{name: "perf", tree: makeTree("counters", "instance_name"), want: "instance_name", count: 1},
		{name: "conf", tree: makeTree("counters", "node-details-info", "cpu-busytime"),
			want: "node-details-info cpu-busytime", count: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var list []string
			tt.tree.FlatList(&list, "")
			if len(list) != tt.count {
				t.Errorf("flat list has size= %v, want %v", len(list), tt.count)
			}
			if list[0] != tt.want {
				t.Errorf("flat list[0] got=[%v], want=[%v]", list[0], tt.want)
			}
		})
	}
}

func makeTree(names ...string) *Node {
	tree := Node{
		name:     []byte("root"),
		Children: make([]*Node, 0),
	}
	cursor := &tree
	for i, n := range names {
		if i == len(names)-1 {
			cursor.AddChild(&Node{Content: []byte(n)})
		} else {
			child := &Node{name: []byte(n)}
			cursor.AddChild(child)
			cursor = child
		}
	}
	return tree.GetChildS(names[0])
}

func TestNode_Union(t *testing.T) {
	parent := &Node{
		name:     []byte("default"),
		Children: make([]*Node, 0),
	}
	child := &Node{
		name:     []byte("volume"),
		Children: make([]*Node, 0),
	}

	testNodeUnionCase1(t, parent, child)
	testNodeUnionCase2(t, parent, child)
	testNodeUnionCase3(t, parent, child)
}

// Parent don't have field and child would have, after the union, parent will be having the field
func testNodeUnionCase1(t *testing.T, parent *Node, child *Node) {
	childClientTimeout := &Node{name: []byte("client_timeout"), Content: []byte("2m")}
	child.AddChild(childClientTimeout)

	parent.Union(child)

	if timeout := parent.GetChildS("client_timeout"); timeout != nil {
		if timeoutVal := timeout.GetContentS(); timeoutVal != "2m" {
			t.Errorf("client timeout after union got=[%v], want=[%v]", timeoutVal, "2m")
		}
	} else {
		t.Errorf("client timeout after union got=[%v], want=[%v]", nil, "2m")
	}
}

// Parent and child both have field but different in sub-child level, after the union, parent will be having union of both
func testNodeUnionCase2(t *testing.T, parent *Node, child *Node) {
	parentScheduleInstance := &Node{name: []byte("instance"), Content: []byte("600s")}
	parentScheduleData := &Node{name: []byte("data"), Content: []byte("180s")}
	parentScheduleCounter := &Node{name: []byte("counter"), Content: []byte("1200s")}
	parentSchedule := &Node{name: []byte("schedule"), Children: []*Node{parentScheduleInstance, parentScheduleData, parentScheduleCounter}}
	parent.AddChild(parentSchedule)

	childScheduleData := &Node{name: []byte("data"), Content: []byte("360s")}
	childSchedule := &Node{name: []byte("schedule"), Children: []*Node{childScheduleData}}
	child.AddChild(childSchedule)

	parent.Union(child)

	if schedule := parent.GetChildS("schedule"); schedule != nil {
		if instanceVal := schedule.GetChildContentS("instance"); instanceVal != "600s" {
			t.Errorf("schedule instance value after union got=[%v], want=[%v]", instanceVal, "600s")
		}
		if dataVal := schedule.GetChildContentS("data"); dataVal != "360s" {
			t.Errorf("schedule data value after union got=[%v], want=[%v]", dataVal, "360s")
		}
		if counterVal := schedule.GetChildContentS("counter"); counterVal != "1200s" {
			t.Errorf("schedule counter value after union got=[%v], want=[%v]", counterVal, "1200s")
		}
	} else {
		t.Errorf("schedule after union got=[%v], want=[%v]", nil, "instance: 600s, data: 360s, counter: 1200s")
	}
}

// Parent and child both have field but different value, after the union, parent will be having child's value
func testNodeUnionCase3(t *testing.T, parent *Node, child *Node) {
	parentClientTimeout := &Node{name: []byte("client_timeout"), Content: []byte("1m")}
	parent.AddChild(parentClientTimeout)

	childClientTimeout := &Node{name: []byte("client_timeout"), Content: []byte("3m")}
	child.AddChild(childClientTimeout)

	parent.Union(child)

	if timeout := parent.GetChildS("client_timeout"); timeout != nil {
		if timeoutVal := timeout.GetContentS(); timeoutVal != "3m" {
			t.Errorf("client timeout after union got=[%v], want=[%v]", timeoutVal, "3m")
		}
	} else {
		t.Errorf("client timeout after union got=[%v], want=[%v]", nil, "3m")
	}
}

func TestGetChildrenNoPanic(t *testing.T) {
	s := NewS("parent")
	child := s.NewChildS("child1", "value1")
	child.AddChild(NewS("sub"))

	assert.Equal(t, 1, len(s.GetChildS("child1").GetChildren()))
	assert.Equal(t, 0, len(s.GetChildS("child2").GetChildren()))
}

func TestNodeGetParam(t *testing.T) {
	n := New([]byte("root"))
	n.NewChildS("aBool", "true")
	n.NewChildS("anInt", "42")
	n.NewChildS("anInt64", "9000000000")
	n.NewChildS("aFloat", "1.5")
	n.NewChildS("aString", "hello")
	n.NewChildS("empty", "")
	n.NewChildS("bad", "notanumber")

	// present and parsable
	assertGetParam(t, n, "aBool", false, true)
	assertGetParam(t, n, "anInt", 7, 42)
	assertGetParam(t, n, "anInt64", int64(7), int64(9000000000))
	assertGetParam(t, n, "aFloat", 0.0, 1.5)
	assertGetParam(t, n, "aString", "def", "hello")

	// absent child and empty content both yield the default, with no error
	assertGetParam(t, n, "missing", 500, 500)
	assertGetParam(t, n, "empty", 500, 500)
	assertGetParam(t, n, "missing", "def", "def")

	// malformed content yields the default *and* an error
	for _, tc := range []struct{ name string }{{"bad"}} {
		if v, err := n.GetParam(tc.name, 500); err == nil || v != 500 {
			t.Errorf("GetParam(%q, 500) = (%d, %v), want (500, non-nil error)", tc.name, v, err)
		}
		if v, err := n.GetParam(tc.name, false); err == nil || v {
			t.Errorf("GetParam(%q, false) = (%v, %v), want (false, non-nil error)", tc.name, v, err)
		}
		if v, err := n.GetParam(tc.name, 0.0); err == nil || v != 0.0 {
			t.Errorf("GetParam(%q, 0.0) = (%v, %v), want (0, non-nil error)", tc.name, v, err)
		}
	}

	// a string default never fails to parse, even for content that isn't a number
	assertGetParam(t, n, "bad", "def", "notanumber")
}

func assertGetParam[T ParamType](t *testing.T, n *Node, key string, def, want T) {
	t.Helper()
	got, err := n.GetParam(key, def)
	assert.Nil(t, err)
	assert.Equal(t, got, want)
}
