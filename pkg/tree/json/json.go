/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package json

import (
	"bytes"
	"errors"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strconv"
	"strings"
)

var (
	SquareOpen  = []byte(`[`)
	SquareClose = []byte(`]`)
	CurlyOpen   = []byte(`{`)
	CurlyClose  = []byte(`}`)
	COMMA       = []byte(`,`)
	COLON       = []byte(`:`)
	QUOTE       = []byte(`"`)
	EMPTY       = []byte(``)
	SPACE       = []byte(` `)
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

	if bytes.HasPrefix(x, SquareOpen) && bytes.HasSuffix(x, SquareClose) {
		// parse as children
		x = bytes.TrimPrefix(x, SquareOpen)
		x = bytes.TrimSuffix(x, SquareClose)
		if elements := bytes.Split(x, []byte(`],[`)); len(elements) > 1 {
			for _, e := range elements {
				child := n.NewChildS("", "")
				parse(child, e)
			}
		} else {
			return parse(n, elements[0])
		}
	} else if bytes.HasPrefix(x, CurlyOpen) && bytes.HasSuffix(x, CurlyClose) {
		x = bytes.TrimPrefix(x, CurlyOpen)
		x = bytes.TrimSuffix(x, CurlyClose)
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
	var childValues [][]byte
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
		childValues = append(childValues, dump(ch))
	}
	return bytes.Join([][]byte{key, CurlyOpen, bytes.Join(childValues, COMMA), CurlyClose}, EMPTY)

}
