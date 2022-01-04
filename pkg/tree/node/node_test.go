package node

import (
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
			if got := simpleName(tt.s); got != tt.want {
				t.Errorf("simpleName() = %v, want %v", got, tt.want)
			}
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
			list := make([]string, 0)
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

	testNodeUnionCase1(parent, child, t)
	testNodeUnionCase2(parent, child, t)
	testNodeUnionCase3(parent, child, t)
}

// Parent don't have field and child would have, after the union, parent will be having the field
func testNodeUnionCase1(parent *Node, child *Node, t *testing.T) {
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
func testNodeUnionCase2(parent *Node, child *Node, t *testing.T) {
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
func testNodeUnionCase3(parent *Node, child *Node, t *testing.T) {
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
