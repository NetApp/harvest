package xml

import (
    "fmt"
    "strings"
    "bytes"
    "encoding/xml"
    "unicode"
    "errors"
    "local.host/share"
)


type Node struct {
    XMLName  xml.Name
    Attrs    []xml.Attr `xml:",any,attr"`
    Content  []byte     `xml:",innerxml"`
    Children []Node     `xml:",any"`
}

func New(name string) *Node {
    return &Node{ XMLName : xml.Name{"",name}}
}

func (n *Node) AddToRoot() *Node {
    var root *Node
    root = New("netapp")
    root.Attrs = append(root.Attrs, xml.Attr{Name: xml.Name{ "","xmlns"}, Value: "http://www.netapp.com/filer/admin"})
    root.Attrs = append(root.Attrs, xml.Attr{Name: xml.Name{"","version"}, Value: "1.3"})
    root.Children = append(root.Children, *n)
    return root
}

func (n *Node) CreateChild(name string, content string) {
    var child Node
    child = *New(name)
    child.Content = []byte(content)
    n.AddChild(child)
}

func (n *Node) AddChild(child Node) {
    n.Children = append(n.Children, child)
}

func (n *Node) GetChildren() []Node {
    return n.Children
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

func (n *Node) GetAttr(name string) (string, bool) {
    for _, a := range n.Attrs {
        if a.Name.Local == name {
            return a.Value, true
        }
    }
    return "", false
}

func (n *Node) GetAttrs() []string {
    var names []string
    for _, a := range n.Attrs {
        names = append(names, a.Name.Local)
    }
    return names
}

func (n *Node) Build() ([]byte, error) {
    var root *Node
    root = n.AddToRoot()
    return xml.Marshal(&root)
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

func SearchByNames(node *Node, prefix []string, paths [][]string) ([]string, bool) {
    var curr_path, matches []string
    var search func(*Node, []string)
    curr_path = make([]string, 0)

    //fmt.Printf("Prefix= %v, Paths= %v\n", prefix, paths)

    search = func(n *Node, curr []string) {
        var newcurr, path []string
        var children []Node
        var child Node
        var content []byte

        if len(curr) > 0 || n.GetName() == prefix[0] {
            newcurr = append(curr, n.GetName())
        } else {
            newcurr = make([]string, len(curr))
            copy(newcurr, curr)
        }

        //fmt.Printf("> [%s]  newcurr= %v\n", n.GetName(), newcurr)

        //fmt.Printf("[%v] with %d children - match: %v\n", mynew, len(children), EqualSlices(path, mynew))
        content, _ = n.GetContent()
        for _, path = range paths {
            if EqualSlices(newcurr, path) {
                matches = append(matches, string(content))
                //fmt.Printf("MATCH <%s>      %v ==> %s (%s)\n", n.GetName(), path, content, n.Content)
                break
            }
        }

        if len(newcurr) < share.MaxLen(paths) {
            children = n.GetChildren()
            for _, child = range children {
                search(&child, newcurr)
            }
        }
    }
    search(node, curr_path)
    //fmt.Printf("          --- Search complte: match = %d\n", len(matches))
    return matches, len(matches)==len(paths)
}

func SearchByPath(root *Node, path []string) []Node {
    var matches []Node
    var curr_path []string
    var search func(*Node, []string)
    curr_path = make([]string, 0)

    search = func(node *Node, curr []string) {
        var newcurr []string
        var children []Node
        //var child Node

        if len(curr) > 0 || node.GetName() == path[0] {
            newcurr = append(curr, node.GetName())
        } else {
            newcurr = make([]string, len(curr))
            copy(newcurr, curr)
        }
        children = node.GetChildren()

        //fmt.Printf("[%v] with %d children - match: %v\n", mynew, len(children), EqualSlices(path, mynew))
        if EqualSlices(newcurr, path) {
            matches = append(matches, *node)
            //name, found := node.GetChildContent("disk-name")
            //fmt.Printf("%s%sMATCH: <%p> <%v> => %s => %s (%v)%s\n", share.Bold, share.Red, node, &node, node.GetName(), name, found, share.End)
        } else if len(newcurr) < len(path) {
            for _, child := range children {
                search(&child, newcurr)
            }
        }
    }
    search(root, curr_path)
    return matches
}

func EqualSlices(a, b []string) bool {
    var i int
    if len(a) != len(b) {
        return false
    }
    for i=0; i<len(a); i+=1 {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

func PrintTree(n *Node, depth int) {
    var COLOR, name, attrs string
    var content []byte
    var exists bool
    var child Node

    if len(n.Children) == 0 {
        COLOR = share.Red
    } else {
        COLOR = share.Cyan
    }

    attrs_names := n.GetAttrs()
    if len(attrs_names) == 0 {
        attrs = ""
    } else {
        attrs = share.Grey + " ("
        for _, a := range attrs_names {
            value, _ := n.GetAttr(a)
            attrs += " " + a + "=\"" + value + "\""
        }
        attrs += " )"
    }

    name = share.Bold + COLOR + strings.Repeat(" ", depth) + n.GetName() + attrs + share.End
    content, exists = n.GetContent()
    if ! exists { content = []byte("-") }

    fmt.Printf("(%d) %-35s %120s (%d)\n", len(n.GetName()), name, string(content), len(content))

    for _, child = range n.Children {
        PrintTree(&child, depth+1)
    }
}
