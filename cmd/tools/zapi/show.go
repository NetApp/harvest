/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package zapi

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func show(n *node.Node, args *Args) {
	switch args.Item {
	case "system":
		showSystem(n)
	case "apis":
		showApis(n)
	case "objects":
		showObjects(n)
	case "attrs":
		showAttrs(n)
	case "counters":
		showCounters(n)
	case "counter":
		showCounter(n)
	case "instances":
		showInstances(n)
	case "data":
		showData(n, args)
	default:
		fmt.Printf("Sorry, I don't know how to show [%s]\n", args.Item)
	}
}

func showSystem(n *node.Node) {
	fmt.Println(n.Print(0))
}

func showApis(n *node.Node) {
	fmt.Println(n.Print(0))
}

func showObjects(item *node.Node) {

	for _, o := range item.GetChildren() {

		if o.GetChildContentS("is-deprecated") != "true" {
			fmt.Printf("%s%s%45s%s: %s\n", color.Bold, color.Cyan, o.GetChildContentS("name"), color.End, o.GetChildContentS("description"))
		} else {
			fmt.Printf("%s%s%45s%s: %s%s%s\n", color.Bold, color.Red, o.GetChildContentS("name"), color.End, color.Grey, o.GetChildContentS("description"), color.End)
		}
	}
}

func showAttrs(n *node.Node) {
	fmt.Println(n.Print(0))
}

func showCounters(n *node.Node) {
	fmt.Println(n.Print(0))
}

func showCounter(n *node.Node) {
	fmt.Println(n.Print(0))
}

func showInstances(n *node.Node) {
	fmt.Println(n.Print(0))
}

func showData(n *node.Node, a *Args) {
	if a.OutputFormat == "xml" {
		// the root node was stripped earlier, add back here
		fmt.Println("<root>")
		fmt.Println(string(n.Content))
		fmt.Println("</root>")
	} else {
		fmt.Println(n.Print(0))
	}
}
