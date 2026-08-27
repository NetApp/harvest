/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package node

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/netapp/harvest/v2/pkg/slogx"
)

type Node struct {
	parent   *Node
	name     []byte
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  []byte     `xml:",innerxml"`
	Children []*Node    `xml:",any"`
}

func New(name []byte) *Node {
	return &Node{name: name}
}

func NewS(name string) *Node {
	return New([]byte(name))
}

func NewXML(name []byte) *Node {
	return NewXMLS(string(name))
}

func NewXMLS(name string) *Node {
	// ugly solution to support xml
	return &Node{XMLName: xml.Name{Local: name}}
}

func (n *Node) GetXMLNameS() string {
	return n.XMLName.Local
}

func (n *Node) SetXMLNameS(name string) {
	n.XMLName = xml.Name{Local: name}
}

func (n *Node) GetName() []byte {
	if name := n.GetXMLNameS(); name != "" {
		return []byte(name)
	}
	return n.name
}

func (n *Node) GetNameS() string {
	if name := n.GetXMLNameS(); name != "" {
		return name
	}
	return string(n.name)
}

func (n *Node) SetNameS(name string) {
	n.name = []byte(name)
}

func (n *Node) GetParent() *Node {
	return n.parent
}

func (n *Node) GetAttr(name string) (xml.Attr, bool) {
	var attr xml.Attr
	for _, attr = range n.Attrs {
		if attr.Name.Local == name {
			return attr, true
		}
	}
	return attr, false
}

func (n *Node) GetAttrValueS(name string) (string, bool) {
	if attr, ok := n.GetAttr(name); ok {
		return attr.Value, true
	}
	return "", false
}

func (n *Node) AddAttr(attr xml.Attr) {
	n.Attrs = append(n.Attrs, attr)
}

func (n *Node) NewAttrS(name, value string) {
	n.AddAttr(xml.Attr{Name: xml.Name{Local: name}, Value: value})
}

func (n *Node) GetChildren() []*Node {
	if n == nil {
		return nil
	}
	return n.Children
}

func (n *Node) GetChildS(name string) *Node {
	for _, child := range n.Children {
		if child.GetNameS() == name {
			return child
		}
	}
	return nil
}

func (n *Node) HasChildS(name string) bool {
	return n.GetChildS(name) != nil
}

func (n *Node) PopChildS(name string) *Node {
	for i, child := range n.Children {
		if child.GetNameS() == name {
			n.Children[i] = n.Children[len(n.Children)-1]
			n.Children = n.Children[:len(n.Children)-1]
			return child
		}
	}
	return nil
}

func (n *Node) NewChildS(name, content string) *Node {
	var child *Node
	if n.GetXMLNameS() != "" {
		child = NewXMLS(name)
	} else {
		child = NewS(name)
	}
	child.parent = n
	child.Content = []byte(content)
	n.AddChild(child)
	return child
}

func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
}

func (n *Node) GetContent() []byte {
	if content := bytes.TrimSpace(n.Content); len(content) != 0 {
		if content[0] != '<' {
			return content
		}
	}
	return []byte("")
}

func (n *Node) GetContentS() string {
	return string(n.Content)
}

/*
func (n *Node) GetContentIfHas() []byte {
    content, _ := n.GetContent()
    return content
}

func (n *Node) GetContentIfHasS() string {
    return string(GetContentIfHas())
}*/

func (n *Node) GetChildContentS(name string) string {
	if child := n.GetChildS(name); child != nil {
		return child.GetContentS()
	}
	return ""
}

// GetChildByContent Compare child content
func (n *Node) GetChildByContent(content string) *Node {
	for _, child := range n.Children {
		if child.GetContentS() == content {
			return child
		}
	}
	return nil
}

func (n *Node) SetChildContentS(name, content string) {
	if child := n.GetChildS(name); child != nil {
		child.SetContentS(content)
	} else {
		n.NewChildS(name, content)
	}
}

func (n *Node) GetAllChildContentS() []string {
	content := make([]string, 0, len(n.Children))
	for _, ch := range n.Children {
		content = append(content, ch.GetContentS())
	}
	return content
}

func (n *Node) GetAllChildNamesS() []string {
	names := make([]string, 0, len(n.Children))
	for _, ch := range n.Children {
		names = append(names, ch.GetNameS())
	}
	return names
}

func (n *Node) SetContent(content []byte) {
	n.Content = content
}

func (n *Node) SetContentS(content string) {
	n.SetContent([]byte(content))
}

func (n *Node) Copy() *Node {
	var clone *Node
	if n.GetXMLNameS() != "" {
		clone = NewXML(n.GetName())
	} else {
		clone = New(n.GetName())
	}
	clone.SetContent(n.GetContent())
	for _, child := range n.Children {
		clone.Children = append(clone.Children, child.Copy())
	}
	return clone
}

func (n *Node) Union(source *Node) {
	if len(n.GetContent()) == 0 {
		n.SetContent(source.GetContent())
	}
	for _, child := range source.Children {
		switch {
		case !n.HasChildS(child.GetNameS()):
			n.AddChild(child)
		case child.GetChildren() != nil:
			n.GetChildS(child.GetNameS()).Union(child)
		default:
			n.SetChildContentS(child.GetNameS(), child.GetContentS())
		}
	}
}

// fetchRoot return if a parent name ancestor exists
func (n *Node) searchAncestor(ancestor string) *Node {
	if n == nil {
		return nil
	}
	p := n.GetParent()
	if p == nil {
		return nil
	}
	if p.GetNameS() == ancestor {
		return n
	}
	return p.searchAncestor(ancestor)
}

func (n *Node) PreprocessTemplate() {
	for _, child := range n.Children {
		mine := n.GetChildS(child.GetNameS())
		if mine != nil && len(child.GetName()) > 0 {
			if mine.searchAncestor("LabelAgent") != nil {
				if mine.GetContentS() != "" {
					mine.NewChildS("", child.GetContentS())
					mine.SetContentS("")
				}
			}
			mine.PreprocessTemplate()
		}
	}
}

// Merge method will merge the subtemplate into the receiver, modifying the receiver in-place.
// skipOverwrite is a readonly list of keys that will not be overwritten in the receiver.
func (n *Node) Merge(subtemplate *Node, skipOverwrite []string) {
	if subtemplate == nil {
		return
	}
	if len(n.Content) == 0 {
		n.Content = subtemplate.Content
	}
	for _, child := range subtemplate.Children {
		mine := n.GetChildS(child.GetNameS())
		switch {
		case len(child.GetName()) == 0:
			if mine != nil && mine.GetParent() != nil && mine.GetParent().GetChildByContent(child.GetContentS()) == nil {
				mine.GetParent().AddChild(child)
			} else if n.GetChildByContent(child.GetContentS()) == nil {
				n.AddChild(child)
			}
		case mine == nil:
			n.AddChild(child)
		default:
			if mine.GetParent() != nil && slices.Contains(skipOverwrite, mine.GetParent().GetNameS()) {
				mine.SetContentS(mine.GetContentS() + "," + child.GetContentS())
			} else {
				mine.SetContentS(child.GetContentS())
			}
			mine.Merge(child, skipOverwrite)
		}
	}
}

func (n *Node) UnmarshalXML(dec *xml.Decoder, root xml.StartElement) error {
	n.Attrs = root.Attr
	type node Node
	return dec.DecodeElement((*node)(n), &root)
}

func (n *Node) FlatList(list *[]string, prefix string) {
	if n == nil {
		return
	}
	if len(n.Children) == 0 {
		var sub string
		if prefix != "" {
			sub = prefix + " " + simpleName(n.GetContentS())
		} else {
			sub = simpleName(n.GetContentS())
		}
		*list = append(*list, sub)
	} else {
		nameS := n.GetNameS()
		if nameS != "" && nameS != "counters" {
			if prefix == "" {
				prefix = nameS
			} else {
				prefix += " " + nameS
			}
		}
		for _, child := range n.Children {
			child.FlatList(list, prefix)
		}
	}
}

var wordRegex = regexp.MustCompile(`(\w|-)+`)

// simpleName returns the first word in the string s
// ignoring non-word characters. see node_test for examples
func simpleName(s string) string {
	return wordRegex.FindString(s)
}

func (n *Node) DebugString() string {
	return n.Print(0)
}

func (n *Node) Print(depth int) string {
	builder := strings.Builder{}
	n.printN(depth, &builder)
	return builder.String()
}

func (n *Node) printN(depth int, b *strings.Builder) {
	name := "* "
	content := " *"
	if n.GetNameS() != "" {
		name = n.GetNameS()
	}

	if n.GetContentS() != "" && n.GetContentS()[0] != '<' {
		content = n.GetContentS()
	}
	fname := fmt.Sprintf("%s[%s]", strings.Repeat("  ", depth), name)
	_, _ = fmt.Fprintf(b, "%-50s - %35s\n", fname, content)
	for _, child := range n.Children {
		child.printN(depth+1, b)
	}
}

func (n *Node) SearchContent(prefix []string, paths [][]string) ([]string, bool) {
	var (
		search  func(*Node, []string)
		matches []string
	)

	search = func(node *Node, currentPath []string) {
		var newPath []string
		if len(currentPath) > 0 || prefix[0] == node.GetNameS() {
			newPath = currentPath
			newPath = append(newPath, node.GetNameS())
		} else {
			newPath = slices.Clone(currentPath)
		}
		for _, path := range paths {
			if slices.Equal(newPath, path) {
				matches = append(matches, node.GetContentS())
				break
			}
		}
		if len(newPath) < MaxLen(paths) {
			for _, child := range node.GetChildren() {
				search(child, newPath)
			}
		}
	}

	search(n, []string{})

	return matches, len(matches) > 0
}

func (n *Node) SearchChildren(path []string) []*Node {
	if len(path) == 0 {
		return nil
	}

	var (
		search  func(*Node, []string)
		matches []*Node
	)

	search = func(node *Node, currentPath []string) {
		var newPath []string
		if len(currentPath) > 0 || path[0] == node.GetNameS() {
			newPath = currentPath
			newPath = append(newPath, node.GetNameS())
		} else {
			newPath = slices.Clone(currentPath)
		}
		if slices.Equal(newPath, path) {
			matches = append(matches, node)
		} else if len(newPath) < len(path) {
			for _, child := range node.GetChildren() {
				search(child, newPath)
			}
		}
	}
	search(n, []string{})
	return matches
}

func DecodeHTML(x string) string {
	x = strings.ReplaceAll(x, "&amp;", "&")
	x = strings.ReplaceAll(x, "&lt;", "<")
	x = strings.ReplaceAll(x, "&gt;", ">")
	x = strings.ReplaceAll(x, "&apos;", "'")
	x = strings.ReplaceAll(x, "&quot;", "\"")
	return Normalize(x)
}

func Normalize(x string) string {
	x = strings.ReplaceAll(x, " ", "_")
	x = strings.ReplaceAll(x, "-", "_")
	return x
}

func ToString(n ast.Node) string {
	switch v := n.(type) {
	case *ast.StringNode:
		return v.Value
	case *ast.NullNode:
		return ""
	default:
		return n.String()
	}
}

func MinLen(elements [][]string) int {
	var smallest, i int
	smallest = len(elements[0])
	for i = 1; i < len(elements); i++ {
		if len(elements[i]) < smallest {
			smallest = len(elements[i])
		}
	}
	return smallest
}

func MaxLen(elements [][]string) int {
	var largest, i int
	largest = len(elements[0])
	for i = 1; i < len(elements); i++ {
		if len(elements[i]) > largest {
			largest = len(elements[i])
		}
	}
	return largest
}

func AllSame(elements [][]string, k int) bool {
	var i int
	for i = 1; i < len(elements); i++ {
		if elements[i][k] != elements[0][k] {
			return false
		}
	}
	return true
}

// ParamType is the set of types GetParam can parse a child's content into.
type ParamType interface {
	bool | int | int64 | float64 | string
}

// GetParam returns the content of the child named key, parsed as T. It is the
// preferred way to read a typed template parameter. Collectors and plugins should
// generally reach for their own LoadParam wrapper instead, which adds logging on top.
//
// When the child is missing or its content is empty, def is returned with a nil
// error -- an absent parameter is not an error, it means "use the default".
// When the content is present but cannot be parsed as T, def is returned along
// with the parse error.
//
// GetParam deliberately cannot distinguish "absent" from "present and equal to
// def". Call sites that must tell those apart (for example ones that set a
// pointer field only when the key is present) should keep using GetChildContentS.
func (n *Node) GetParam[T ParamType](key string, def T) (T, error) {
	x := n.GetChildContentS(key)
	if x == "" {
		return def, nil
	}

	var (
		parsed any
		err    error
	)

	switch any(def).(type) {
	case bool:
		parsed, err = strconv.ParseBool(x)
	case int:
		parsed, err = strconv.Atoi(x)
	case int64:
		parsed, err = strconv.ParseInt(x, 10, 64)
	case float64:
		parsed, err = strconv.ParseFloat(x, 64)
	case string:
		parsed = x
	}

	if err != nil {
		return def, err
	}

	// parsed always holds a T: the switch above is exhaustive over ParamType
	// and every branch produces T's underlying type. The guard is belt-and-braces
	// so a future addition to ParamType without a matching case cannot panic.
	v, ok := parsed.(T)
	if !ok {
		return def, fmt.Errorf("cannot parse %q as %T", x, def)
	}

	return v, nil
}

// LoadParam reads the parameter name from n and parses it as T, falling back to
// def when the parameter is absent or malformed. A malformed value is reported at
// Warn on logger and does not fail the caller -- the default is used instead.
func LoadParam[T ParamType](n *Node, logger *slog.Logger, name string, def T) T {
	v, err := n.GetParam(name, def)
	if err != nil {
		logger.Warn(
			"invalid parameter, using default",
			slog.String("name", name),
			slog.Any("default", def),
			slogx.Err(err),
		)

		return def
	}

	logger.Debug(
		"using",
		slog.String("name", name),
		slog.Any("value", v),
		// so operators can tell a template override from a built-in default
		slog.Bool("isDefault", n.GetChildContentS(name) == ""),
	)

	return v
}
