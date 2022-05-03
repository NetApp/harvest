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

func dumpRecursive(node *node.Node, data *[][]byte, depth int) {
	indentation := bytes.Repeat([]byte("  "), depth)
	parentName := "Root"
	if node.GetParent() != nil {
		parentName = node.GetParent().GetNameS()
	}
	// workaround to handle list of maps
	if parentName == "labels" {
		*data = append(*data, joinAll(indentation, []byte("- "), node.GetName(), []byte(": "), node.GetContent()))
	} else if len(node.GetName()) != 0 && len(node.GetContent()) != 0 && parentName != "labels" {
		*data = append(*data, joinAll(indentation, node.GetName(), []byte(": "), node.GetContent()))
	} else if len(node.GetName()) != 0 {
		*data = append(*data, joinAll(indentation, node.GetName(), []byte(":")))
	} else {
		*data = append(*data, joinAll(indentation, []byte("- "), node.GetContent()))
	}
	for _, child := range node.GetChildren() {
		dumpRecursive(child, data, depth+1)
	}
}

func joinAll(slices ...[]byte) []byte {
	return bytes.Join(slices, []byte(""))
}
