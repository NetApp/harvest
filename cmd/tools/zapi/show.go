package main

import (
	"fmt"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
)

func show(n *node.Node, args *Args) {
	switch args.Item {
	case "system":
		show_system(n, args)
	case "apis":
		show_apis(n, args)
	case "objects":
		show_objects(n, args)
	case "attrs":
		show_attrs(n, args)
	case "counters":
		show_counters(n, args)
	case "counter":
		show_counter(n, args)
	case "instances":
		show_instances(n, args)
	case "data":
		show_data(n, args)
	default:
		fmt.Printf("Sorry, I don't know how to show [%s]\n", args.Item)
	}
}

func show_system(n *node.Node, args *Args) {
	n.Print(0)
}

func show_apis(n *node.Node, args *Args) {
	n.Print(0)
}

func show_objects(item *node.Node, args *Args) {

	for _, o := range item.GetChildren() {

		if o.GetChildContentS("is-deprecated") != "true" {
			fmt.Printf("%s%s%45s%s: %s\n", util.Bold, util.Cyan, o.GetChildContentS("name"), util.End, o.GetChildContentS("description"))
		} else {
			fmt.Printf("%s%s%45s%s: %s%s%s\n", util.Bold, util.Red, o.GetChildContentS("name"), util.End, util.Grey, o.GetChildContentS("description"), util.End)
		}
	}
}

func show_attrs(n *node.Node, args *Args) {
	n.Print(0)
}

func show_counters(n *node.Node, args *Args) {
	n.Print(0)
}

func show_counter(n *node.Node, args *Args) {
	n.Print(0)
}

func show_instances(n *node.Node, args *Args) {
	n.Print(0)
}

func show_data(n *node.Node, args *Args) {
	n.Print(0)
}
