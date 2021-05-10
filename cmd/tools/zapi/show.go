/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package main

import (
	"fmt"
	"goharvest2/pkg/color"
	"goharvest2/pkg/tree/node"
)

func show(n *node.Node, args *Args) {
	switch args.Item {
	case "system":
		showSystem(n, args)
	case "apis":
		showApis(n, args)
	case "objects":
		showObjects(n, args)
	case "attrs":
		showAttrs(n, args)
	case "counters":
		showCounters(n, args)
	case "counter":
		showCounter(n, args)
	case "instances":
		showInstances(n, args)
	case "data":
		showData(n, args)
	default:
		fmt.Printf("Sorry, I don't know how to show [%s]\n", args.Item)
	}
}

func showSystem(n *node.Node, args *Args) {
	n.Print(0)
}

func showApis(n *node.Node, args *Args) {
	n.Print(0)
}

func showObjects(item *node.Node, args *Args) {

	for _, o := range item.GetChildren() {

		if o.GetChildContentS("is-deprecated") != "true" {
			fmt.Printf("%s%s%45s%s: %s\n", color.Bold, color.Cyan, o.GetChildContentS("name"), color.End, o.GetChildContentS("description"))
		} else {
			fmt.Printf("%s%s%45s%s: %s%s%s\n", color.Bold, color.Red, o.GetChildContentS("name"), color.End, color.Grey, o.GetChildContentS("description"), color.End)
		}
	}
}

func showAttrs(n *node.Node, args *Args) {
	n.Print(0)
}

func showCounters(n *node.Node, args *Args) {
	n.Print(0)
}

func showCounter(n *node.Node, args *Args) {
	n.Print(0)
}

func showInstances(n *node.Node, args *Args) {
	n.Print(0)
}

func showData(n *node.Node, args *Args) {
	n.Print(0)
}
