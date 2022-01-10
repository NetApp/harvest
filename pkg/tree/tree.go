/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package tree

import (
	"fmt"
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
	consume(r, "", root.Content[0], 0)
	return r, nil
}

func consume(r *node.Node, key string, y *y3.Node, level int) {
	if y.Kind == y3.ScalarNode {
		r.NewChildS(key, y.Value)
	} else if y.Kind == y3.MappingNode {
		var s = r
		// handles below yaml structure. This induces a parent in between for grouping of child components
		/*
			endpoints:
			  - query: api/private/cli/volume
			    counters:
			      - ^^instance_uuid => instance_uuid
			      - ^node  => node
			  - query: api/private/cli/svm
			    counters:
			      - ^^instance_uuid => instance_uuid
			      - ^node  => node
		*/
		// condition s.GetNameS() != "plugins" && s.SearchAncestor("LabelAgent") == nil is needed to support 21.08 older format for plugins
		ans := s.SearchAncestor("LabelAgent")
		fmt.Println(ans)
		if key == "" && (s.GetParent() != nil && s.GetNameS() != "plugins" && s.SearchAncestor("LabelAgent") == nil) && len(y.Content) > 2 {
			key = strconv.Itoa(level)
		}
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
			consume(s, k, y.Content[i+1], level)
		}
	} else { // sequence
		s := r.NewChildS(key, "")
		for level, child := range y.Content {
			consume(s, "", child, level)
		}
	}
}

func LoadXml(data []byte) (*node.Node, error) {
	return xml.Load(data)
}

func DumpXml(n *node.Node) ([]byte, error) {
	return xml.Dump(n)
}
