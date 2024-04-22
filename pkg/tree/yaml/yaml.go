/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package yaml

import (
	"bytes"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func Dump(root *node.Node) ([]byte, error) {
	data := make([][]byte, 0)
	for _, n := range root.GetChildren() {
		dumpRecursive(n, &data, 0)
	}
	return bytes.Join(data, []byte("\n")), nil
}

func dumpRecursive(aNode *node.Node, data *[][]byte, depth int) {
	indentation := bytes.Repeat([]byte("  "), depth)
	parentName := "Root"
	if aNode.GetParent() != nil {
		parentName = aNode.GetParent().GetNameS()
	}
	// workaround to handle a list of maps
	switch {
	case parentName == "labels":
		*data = append(*data, joinAll(indentation, []byte("- "), aNode.GetName(), []byte(": "), aNode.GetContent()))
	case len(aNode.GetName()) != 0 && len(aNode.GetContent()) != 0:
		*data = append(*data, joinAll(indentation, aNode.GetName(), []byte(": "), aNode.GetContent()))
	case len(aNode.GetName()) != 0:
		*data = append(*data, joinAll(indentation, aNode.GetName(), []byte(":")))
	default:
		*data = append(*data, joinAll(indentation, []byte("- "), aNode.GetContent()))
	}
	for _, child := range aNode.GetChildren() {
		dumpRecursive(child, data, depth+1)
	}
}

func joinAll(slices ...[]byte) []byte {
	return bytes.Join(slices, []byte(""))
}
