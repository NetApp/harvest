package yaml

import (
    "fmt"
    "bytes"
    "strings"
    "io/ioutil"
    "poller/share"
)


func Import(filepath string) (*Node, error) {
/* Imports and parses a Yaml file. Result is a tree of *Nodes.
   Returns a pointer to the root of this tree.
*/
    var err error
    var root *Node
    var filecontent []byte

    //fmt.Printf("importing from [%s]\n", filepath)
    filecontent, err = ioutil.ReadFile(filepath) /* TODO: issue when file doesn't exist! */
    if err == nil {
        root = &Node{ Name : "Root" }
        err = root.parse(bytes.Split(filecontent, []byte("\n")), 0, 0)
    }
    return root, err
}

type Node struct {
    parent *Node
    Name string
    Value string
    Values []string
    Children []*Node
}

func New(name, value string) *Node {
    return &Node{Name: name, Value: value}
}

func (node *Node) Copy() *Node {
    n := Node{ Name : node.Name }
    n.Value = node.Value
    n.Values = make([]string, len(node.Values))
    copy(n.Values, node.Values)
    
    for _, child := range node.Children {
        n.Children = append(n.Children, child.Copy())
    }
    return &n
}

func (node *Node) AddChild(child *Node) {
    node.Children = append(node.Children, child)
}

func (node *Node) AddNewChild(name, value string) {
    node.Children = append(node.Children, New(name, value))
}

func (node *Node) GetChild(name string) *Node {
    var child *Node
    for _, child = range node.Children {
        if child.Name == name {
            return child
        }
    }
    return nil
}

func (node *Node) PopChild(name string) *Node {
    var child *Node
    var i, size int
    size = len(node.Children)
    for i, child = range node.Children {
        if child.Name == name {
            node.Children[i] = node.Children[size-1]
            node.Children = node.Children[:size-1]
            return child
        }
    }
    return nil
}

func (node *Node) HasChild(name string) bool {
    return node.GetChild(name) != nil
}


func (node *Node) GetChildren() []*Node {
    return node.Children
}

func (node *Node) SetValue(value string) {
    node.Value = value
}

func (node *Node) AddValue(value string) {
    node.Values = append(node.Values, value)
}

func (node *Node) HasInValues(value string) bool {
    for _, v := range node.Values {
        if v == value {
            return true
        }
    }
    return false
}

func (node *Node) GetChildValue(name string) string {
    var child *Node

    if child = node.GetChild(name); child != nil {
        return child.Value
    }
    return ""
}

func (node *Node) GetChildValues(name string) []string {
    var child *Node
    var values []string

    if child = node.GetChild(name); child != nil {
        return child.Values
    }
    return values
}

func (node *Node) Union(source *Node, deep bool) {
    var child *Node
    var value string

    /* merge children */
    for _, child = range source.Children {  
        if ! node.HasChild(child.Name) {
            node.AddChild(child)
        } else if deep {  /* optionally do a deep merge */
            node.GetChild(child.Name).Union(child, true)
        }
    }
    /* merge value */
    if node.Value == "" && source.Value != "" {
        node.SetValue(source.Value)
    }
    /* merge values */
    for _, value = range source.Values {
        if ! node.HasInValues(value) {
            node.AddValue(value)
        }
    }
}

/* Function for debugging / testing
   Should never be called by a daemon */
func (node *Node) PrintTree(depth int) {
    var child *Node

    fmt.Println(node.ToString(depth))

    for _, child = range node.Children {
        child.PrintTree(depth+1)
    }
}

func (node *Node) ToString(depth int) string {
    var name string
    name = fmt.Sprintf("%s%s%s%s%s (%d)", 
        strings.Repeat("  ", depth), 
        share.Bold, 
        share.Cyan, 
        node.Name, 
        share.End, 
        len(node.Children),
    )
    return fmt.Sprintf("%-50s - %s%-35s%s - %s%s%s", 
        name, 
        share.Green, 
        node.Value, 
        share.End, 
        share.Pink, 
        strings.Join(node.Values, ", "),
        share.End,
    )
}
