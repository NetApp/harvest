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
