/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package tree

import (
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/tree/xml"
	y3 "gopkg.in/yaml.v3"
	"os"
)

func ImportYaml(filepath string) (*node.Node, error) {
	data, err := os.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	return LoadYaml(data)
}

func LoadYaml(data []byte) (*node.Node, error) {
	root := y3.Node{}
	err := y3.Unmarshal(data, &root)
	if err != nil {
		return nil, err
	}
	// treat an empty file as an error
	if len(root.Content) == 0 {
		return nil, errs.New(errs.ErrConfig, "template file is empty or does not exist")
	}

	r := node.New([]byte("Root"))
	consume(r, "", root.Content[0], false)
	return r, nil
}

func consume(r *node.Node, key string, y *y3.Node, makeNewChild bool) {
	switch y.Kind {
	case y3.ScalarNode:
		r.NewChildS(key, y.Value)
	case y3.MappingNode:
		var s = r
		if key != "" || makeNewChild {
			s = r.NewChildS(key, "")
		}
		for i := 0; i < len(y.Content); i += 2 {
			k := y.Content[i].Value
			// special case to handle incorrectly indented LabelAgent
			if k == "LabelAgent" && y.Content[i+1].Kind == y3.ScalarNode {
				s = r.NewChildS(k, "")
				continue
			}
			consume(s, k, y.Content[i+1], false)
		}
	case y3.DocumentNode, y3.SequenceNode, y3.AliasNode:
		s := r.NewChildS(key, "")
		for _, child := range y.Content {
			makeNewChild := false
			if child.Tag == "!!map" {
				makeNewChild = key == "endpoints" || key == "events" || key == "matches"
			}
			consume(s, "", child, makeNewChild)
		}
	}
}

func LoadXML(data []byte) (*node.Node, error) {
	return xml.Load(data)
}

func DumpXML(n *node.Node) ([]byte, error) {
	return xml.Dump(n)
}

func ImportXML(filepath string) (*node.Node, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return LoadXML(data)
}
