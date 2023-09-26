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

func Load(data io.Reader, log bool) (*node.Node, []byte, error) {
	var buf bytes.Buffer
	reader := data

	if log {
		reader = io.TeeReader(data, &buf)
	}

	root := new(node.Node)
	dec := xml.NewDecoder(reader)
	if err := dec.Decode(&root); err != nil {
		return nil, nil, err
	}

	return root, buf.Bytes(), nil
}

func Dump(n *node.Node) ([]byte, error) {
	return xml.Marshal(&n)
}
