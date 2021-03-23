package json

import (
	"bytes"
	"errors"
	"goharvest2/share/tree/node"
	"strconv"
	"strings"
)

var (
	SQUARE_OPEN  = []byte(`[`)
	SQUARE_CLOSE = []byte(`]`)
	CURLY_OPEN   = []byte(`{`)
	CURLY_CLOSE  = []byte(`}`)
	COMMA        = []byte(`,`)
	COLON        = []byte(`:`)
	QUOTE        = []byte(`"`)
	EMPTY        = []byte(``)
	SPACE        = []byte(` `)
)

func Load(data []byte) (*node.Node, error) {
	root := node.New([]byte(""))
	//data = bytes.ReplaceAll(data, []byte("\n"), EMPTY)
	//data = bytes.ReplaceAll(data, []byte(`": `), []byte(`":`))
	return root, parse(root, data)
}

func Dump(n *node.Node) []byte {
	return dump(n)
}

func parse(n *node.Node, x []byte) error {

	if bytes.HasPrefix(x, SQUARE_OPEN) && bytes.HasSuffix(x, SQUARE_CLOSE) {
		// parse as children
		x = bytes.TrimPrefix(x, SQUARE_OPEN)
		x = bytes.TrimSuffix(x, SQUARE_CLOSE)
		if elements := bytes.Split(x, []byte(`],[`)); len(elements) > 1 {
			for _, e := range elements {
				child := n.NewChildS("", "")
				parse(child, e)
			}
		} else {
			return parse(n, elements[0])
		}
	} else if bytes.HasPrefix(x, CURLY_OPEN) && bytes.HasSuffix(x, CURLY_CLOSE) {
		x = bytes.TrimPrefix(x, CURLY_OPEN)
		x = bytes.TrimSuffix(x, CURLY_CLOSE)
		if elements := bytes.Split(x, []byte(`},{`)); len(elements) > 1 {
			for _, e := range elements {
				child := n.NewChildS("", "")
				parse(child, e)
			}
		} else {
			return parse(n, elements[0])
		}
	} else if pairs := bytes.Split(x, []byte(`,"`)); len(pairs) > 1 {
		for _, p := range pairs {

			if values := bytes.Split(p, []byte(`":`)); len(values) == 2 {
				key := bytes.TrimPrefix(values[0], QUOTE)
				value := bytes.TrimPrefix(values[1], QUOTE)
				value = bytes.TrimSuffix(value, QUOTE)
				n.NewChild(bytes.TrimSpace(key), bytes.TrimSpace(value))
			}
		}

	} else {
		return errors.New("invalid element: " + string(x))
	}
	return nil
}

func dump(n *node.Node) []byte {

	var key []byte
	var child_values [][]byte
	var value []byte
	var err error

	if len(n.GetName()) != 0 {
		key = bytes.Join([][]byte{QUOTE, n.GetName(), QUOTE, COLON}, EMPTY)
	}

	if len(n.GetContent()) != 0 {
		s := n.GetContentS()
		if s == "true" || s == "false" || s == "null" {
			value = n.GetContent()
		} else if _, err = strconv.ParseFloat(s, 64); err == nil {
			value = n.GetContent()
		} else if strings.HasPrefix(s, `{`) {
			value = n.GetContent()
		} else {
			value = bytes.Join([][]byte{QUOTE, n.GetContent(), QUOTE}, EMPTY)
		}
		return bytes.Join([][]byte{key, value}, EMPTY)
	}

	for _, ch := range n.GetChildren() {
		child_values = append(child_values, dump(ch))
	}
	return bytes.Join([][]byte{key, CURLY_OPEN, bytes.Join(child_values, COMMA), CURLY_CLOSE}, EMPTY)

}
