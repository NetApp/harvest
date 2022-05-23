/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package tree

import (
	"github.com/netapp/harvest/v2/pkg/errors"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/tree/xml"
	y3 "gopkg.in/yaml.v3"
	"io/ioutil"
)

func ImportYaml(filepath string) (*node.Node, error) {
	data, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	root := y3.Node{}
	err = y3.Unmarshal(data, &root)
	if err != nil {
		return nil, err
	}
	// treat an empty file as an error
	if len(root.Content) == 0 {
		return nil, errors.New(errors.ErrConfig, "template file is empty or does not exist")
	}

	r := node.New([]byte("Root"))
	consume(r, "", root.Content[0], false)
	return r, nil
}

func consume(r *node.Node, key string, y *y3.Node, makeNewChild bool) {
	if y.Kind == y3.ScalarNode {
		r.NewChildS(key, y.Value)
	} else if y.Kind == y3.MappingNode {
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
	} else { // sequence
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

func LoadXml(data []byte) (*node.Node, error) {
	return xml.Load(data)
}

func DumpXml(n *node.Node) ([]byte, error) {
	return xml.Dump(n)
}
