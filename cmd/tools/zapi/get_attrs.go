/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package main

import (
	"fmt"
	client "goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/tree/node"
	"strings"
)

func getAttrs(c *client.Client, a *Args) (*node.Node, error) {

	var (
		req, apis, results, attr *node.Node
		err                      error
	)

	req = node.NewXmlS("system-api-get-elements")
	apis = req.NewChildS("api-list", "")
	apis.NewChildS("api-list-info", a.Api)

	if results, err = c.InvokeRequest(req); err != nil {
		return nil, err
	}

	output := node.NewS("output")
	input := node.NewS("input")

	if entries := results.GetChildS("api-entries"); entries != nil && len(entries.GetChildren()) > 0 {
		if elements := entries.GetChildren()[0].GetChildS("api-elements"); elements != nil {
			for _, x := range elements.GetChildren() {
				if x.GetChildContentS("is-output") == "true" {
					x.PopChildS("is-output")
					output.AddChild(x)
				} else {
					input.AddChild(x)
				}
			}
		}
	}

	fmt.Println("############################        INPUT        ##########################")
	input.Print(0)
	fmt.Println()
	fmt.Println()

	fmt.Println("############################        OUPUT        ##########################")
	output.Print(0)
	fmt.Println()
	fmt.Println()

	// fetch root attribute
	attr_key := ""
	attr_name := ""

	for _, x := range output.GetChildren() {
		if t := x.GetChildContentS("type"); t == "string" || t == "integer" {
			continue
		}
		if name := x.GetChildContentS("name"); true {
			attr_key = name
			attr_name = x.GetChildContentS("type")
			break
		}
	}

	if attr_name == "" {
		fmt.Println("no root attribute, stopping here.")
		return nil, errors.New(ATTRIBUTE_NOT_FOUND, "root attribute")
	}

	if strings.HasSuffix(attr_name, "[]") {
		attr_name = strings.TrimSuffix(attr_name, "[]")
	}

	fmt.Printf("building tree for attribute [%s] => [%s]\n", attr_key, attr_name)

	if results, err = c.InvokeRequestString("system-api-list-types"); err != nil {
		return nil, err
	}

	entries := results.GetChildS("type-entries")
	if entries == nil {
		fmt.Println("Error: missing [type-entries]")
		return nil, errors.New(ATTRIBUTE_NOT_FOUND, "type-entries")
	}

	attr = node.NewS(attr_name)
	search_entries(attr, entries)

	//fmt.Println("############################        ATTR         ##########################")
	//attr.Print(0)
	//fmt.Println()
	/*
		if args.Export {
			fn := path.Join("/tmp", args.Api+".yml")
			if err = tree.Export(attr, "yaml", fn); err != nil {
				fmt.Printf("failed to export to [%s]:\n", fn)
				fmt.Println(err)
			} else {
				fmt.Printf("exported to [%s]\n", fn)
			}
		}
	*/
	return attr, nil
}

func search_entries(root, entries *node.Node) {

	cache := make(map[string]*node.Node)
	cache[root.GetNameS()] = root

	for i := 0; i < maxSearchDepth; i += 1 {
		for _, entry := range entries.GetChildren() {
			name := entry.GetChildContentS("name")
			if parent, ok := cache[name]; ok {
				delete(cache, name)
				if elems := entry.GetChildS("type-elements"); elems != nil {
					for _, elem := range elems.GetChildren() {
						child := parent.NewChildS(elem.GetChildContentS("name"), "")
						attr_type := strings.TrimSuffix(elem.GetChildContentS("type"), "[]")
						if !knownTypes.Has(attr_type) {
							cache[attr_type] = child
						}
					}
				}
			}
		}
	}
}
