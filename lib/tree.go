package lib

import (
    "fmt"
    "strings"
    "bytes"
    "encoding/xml"
    "unicode"
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

func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
    n.Attrs = start.Attr
    type node Node
    return d.DecodeElement((*node)(n), &start)
}

func (n *Node) GetName() string {
    return strings.TrimFunc(n.XMLName.Local, unicode.IsSpace)
}

func (n *Node) GetContent() ([]byte, bool) {
    //var i int
    var content []byte
    if len(n.Content) != 0 {
        /*
        i = 0
        for {
            if !unicode.IsSpace(rune(n.Content[i])) || i==len(n.Content)-1 {
                break
            }
            i += 1
        }
        */
        //fmt.Printf("n.Content[%d] = %s\n", i, string(n.Content[i]))
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

func NewTree(data []byte) (*Node, error) {
    var buffer *bytes.Buffer
    var decoder *xml.Decoder
    var node Node
    var err error
    buffer = bytes.NewBuffer(data)
    decoder = xml.NewDecoder(buffer)
    err = decoder.Decode(&node)
    return &node, err
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
