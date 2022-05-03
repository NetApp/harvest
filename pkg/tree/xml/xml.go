/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package xml

import (
	"bytes"
	"encoding/xml"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"io"
)

func Load(data []byte) (*node.Node, error) {
	root := new(node.Node)
	buf := bytes.NewBuffer(data)
	dec := xml.NewDecoder(buf)
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}
	return root, nil
}

func LoadFromReader(r io.Reader) (*node.Node, error) {
	root := new(node.Node)
	dec := xml.NewDecoder(r)
	return root, dec.Decode(&root)
}

func Dump(n *node.Node) ([]byte, error) {
	return xml.Marshal(&n)
}
