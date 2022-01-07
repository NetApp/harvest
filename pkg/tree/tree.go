/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package tree

import (
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/tree/xml"
	y3 "gopkg.in/yaml.v3"
	"io/ioutil"
	"strconv"
)

func ImportYaml(filepath string) (*node.Node, error) {
	data, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	root := y3.Node{}
	err = y3.Unmarshal(data, &root)
	if err != nil || len(root.Content) == 0 {
		return nil, err
	}
	r := node.New([]byte("Root"))
	consume(r, "", root.Content[0])
	return r, nil
}

func consume(r *node.Node, key string, y *y3.Node) {
	if y.Kind == y3.ScalarNode {
		r.NewChildS(key, y.Value)
	} else if y.Kind == y3.MappingNode {
		var s = r
		if key != "" {
			s = r.NewChildS(key, "")
		}
		for i := 0; i < len(y.Content); i += 2 {
			k := y.Content[i].Value
			// special case to handle incorrectly indented LabelAgent
			if k == "LabelAgent" && y.Content[i+1].Kind == y3.ScalarNode {
				s = r.NewChildS(k, "")
				continue
			}
			consume(s, k, y.Content[i+1])
		}
	} else { // sequence
		s := r.NewChildS(key, "")
		for i, child := range y.Content {
			if key == "endpoints" {
				consume(s, strconv.Itoa(i), child)
			} else {
				consume(s, "", child)
			}
		}
	}
}

func LoadXml(data []byte) (*node.Node, error) {
	return xml.Load(data)
}

func DumpXml(n *node.Node) ([]byte, error) {
	return xml.Dump(n)
}
