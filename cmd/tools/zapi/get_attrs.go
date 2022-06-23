/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package zapi

import (
	"fmt"
	client "github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strings"
)

func getAttrs(c *client.Client, a *Args) (*node.Node, error) {

	var (
		req, apis, results, attr *node.Node
		err                      error
	)

	req = node.NewXMLS("system-api-get-elements")
	apis = req.NewChildS("api-list", "")
	apis.NewChildS("api-list-info", a.API)

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
	fmt.Println(input.Print(0))
	fmt.Println()
	fmt.Println()

	fmt.Println("############################        OUPUT        ##########################")
	fmt.Println(output.Print(0))
	fmt.Println()
	fmt.Println()

	// fetch root attribute
	attrKey := ""
	attrName := ""

	for _, x := range output.GetChildren() {
		if t := x.GetChildContentS("type"); t == "string" || t == "integer" {
			continue
		}
		if name := x.GetChildContentS("name"); true {
			attrKey = name
			attrName = x.GetChildContentS("type")
			break
		}
	}

	if attrName == "" {
		fmt.Println("no root attribute, stopping here.")
		return nil, errs.New(errs.ErrAttributeNotFound, "root attribute")
	}

	attrName = strings.TrimSuffix(attrName, "[]")

	fmt.Printf("building tree for attribute [%s] => [%s]\n", attrKey, attrName)

	if results, err = c.InvokeRequestString("system-api-list-types"); err != nil {
		return nil, err
	}

	entries := results.GetChildS("type-entries")
	if entries == nil {
		fmt.Println("Error: missing [type-entries]")
		return nil, errs.New(errs.ErrAttributeNotFound, "type-entries")
	}

	attr = node.NewS(attrName)
	searchEntries(attr, entries)

	return attr, nil
}

func searchEntries(root, entries *node.Node) {

	cache := make(map[string]*node.Node)
	cache[root.GetNameS()] = root

	for i := 0; i < maxSearchDepth; i++ {
		for _, entry := range entries.GetChildren() {
			name := entry.GetChildContentS("name")
			if parent, ok := cache[name]; ok {
				delete(cache, name)
				if elems := entry.GetChildS("type-elements"); elems != nil {
					for _, elem := range elems.GetChildren() {
						child := parent.NewChildS(elem.GetChildContentS("name"), "")
						attrType := strings.TrimSuffix(elem.GetChildContentS("type"), "[]")
						cache[attrType] = child
						if strings.Contains(attrType, "-info") {
							child.SetContentS(" ")
						} else {
							child.SetContentS(attrType)
						}
					}
				}
			}
		}
	}
}
