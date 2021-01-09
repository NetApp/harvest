package lib

import (
    "fmt"
    "strings"
    "bytes"
    "encoding/xml"
    "unicode"
    "errors"
)

var RED string = "\033[31m"
var PURPLE string = "\033[35m"
var CYAN string = "\033[36m"
var BOLD string = "\033[1m"
var END string = "\033[0m"
var GREY string = "\033[90m"


type Node struct {
    XMLName  xml.Name
    Attrs    []xml.Attr `xml:",any,attr"`
    Content  []byte     `xml:",innerxml"`
    Children []Node     `xml:",any"`
}

func NewNode(name string) *Node {
    var node Node
    var xmlname xml.Name
    xmlname = xml.Name{ "", name }
    node = Node{ XMLName : xmlname }
    return &node
}

func (n *Node) CreateChild(name string, content string) {
    var child Node
    child = *NewNode(name)
    child.Content = []byte(content)
    n.AddChild(child)
}

func (n *Node) AddChild(child Node) {
    n.Children = append(n.Children, child)
}

func (n *Node) GetChild(name string) (*Node, bool) {
    var child Node
    for _, child = range n.Children {
        if child.GetName() == name {
            return &child, true
        }
    }
    return nil, false
}

func (n *Node) GetChildContent(name string) ([]byte, bool) {
    var child *Node
    var found bool
    child, found = n.GetChild(name)
    if found == true {
        return child.GetContent()
    }
    return nil, false
}

func (n *Node) GetName() string {
    return strings.TrimFunc(n.XMLName.Local, unicode.IsSpace)
}

func (n *Node) GetContent() ([]byte, bool) {
    var content []byte
    if len(n.Content) != 0 {
        content = bytes.TrimFunc(n.Content, unicode.IsSpace)
        if content[0] != '<' {
            return content, true
        }
    }
    return nil, false
}

func (n *Node) GetAttribute(name string) (string, bool) {
    for _, a := range n.Attrs {
        if a.Name.Local == name {
            return a.Value, true
        }
    }
    return "", false
}

func (n *Node) GetAttributeNames() []string {
    var names []string
    for _, a := range n.Attrs {
        names = append(names, a.Name.Local)
    }
    return names
}

func (n *Node) Build() ([]byte, error) {
    return xml.Marshal(&n)
}

func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
    n.Attrs = start.Attr
    type node Node
    return d.DecodeElement((*node)(n), &start)
}

func Parse(data []byte) (*Node, error) {
    var buffer *bytes.Buffer
    var decoder *xml.Decoder
    var node, root *Node
    var found bool
    var err error

    buffer = bytes.NewBuffer(data)
    decoder = xml.NewDecoder(buffer)
    err = decoder.Decode(&node)

    if err == nil {
        root, found = node.GetChild("results")
        if found != true {
            err = errors.New("Root element \"results\" missing")
        }
    }
    return root, err
}

func PrintTree(n *Node, depth int) {
    var COLOR, name, attrs string
    var content []byte
    var exists bool
    var child Node

    if len(n.Children) == 0 {
        COLOR = RED
    } else {
        COLOR = CYAN
    }

    attrs_names := n.GetAttributeNames()
    if len(attrs_names) == 0 {
        attrs = ""
    } else {
        attrs = GREY + " ("
        for _, a := range attrs_names {
            value, _ := n.GetAttribute(a)
            attrs += " " + a + "=\"" + value + "\""
        }
        attrs += " )"
    }

    name = BOLD + COLOR + strings.Repeat(" ", depth) + n.GetName() + attrs + END
    content, exists = n.GetContent()
    if ! exists { content = []byte("-") }

    fmt.Printf("(%d) %-35s %120s (%d)\n", len(n.GetName()), name, string(content), len(content))

    for _, child = range n.Children {
        PrintTree(&child, depth+1)
    }
}
