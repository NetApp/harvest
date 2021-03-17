package xml

import (
	"bytes"
	"encoding/xml"
	"goharvest2/share/tree/node"
)

func Load(data []byte) (*node.Node, error) {
	var root *node.Node
	buf := bytes.NewBuffer(data)
	dec := xml.NewDecoder(buf)
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}
	return root, nil
}

func Dump(n *node.Node) ([]byte, error) {
	return xml.Marshal(&n)
}
