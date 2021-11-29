/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package tree

import (
	"errors"
	"goharvest2/pkg/tree/json"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/tree/xml"
	"goharvest2/pkg/tree/yaml"
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
		for _, child := range y.Content {
			consume(s, "", child)
		}
	}
}

func Import(format, filepath string) (*node.Node, error) {

	data, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	switch format {
	case "yaml":
		return yaml.Load(data)
	case "xml":
		return xml.Load(data)
	case "json":
		return json.Load(data)
	}

	return nil, errors.New("unknown format: " + format)
}

func LoadXml(data []byte) (*node.Node, error) {
	return xml.Load(data)
}

func DumpXml(n *node.Node) ([]byte, error) {
	return xml.Dump(n)
}
