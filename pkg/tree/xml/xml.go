/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package xml

import (
	"encoding/xml"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"io"
)

func Load(data io.Reader) (*node.Node, error) {
	root := new(node.Node)
	dec := xml.NewDecoder(data)
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}
	return root, nil
}

func Dump(n *node.Node) ([]byte, error) {
	return xml.Marshal(&n)
}
