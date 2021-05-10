/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	client "goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/color"
	"goharvest2/pkg/config"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/tree/node"
)

var (
	maxSearchDepth  = 10
	harvestConfPath = config.GetHarvestConf()
)

func main() {

	var (
		err          error
		args         *Args
		item, params *node.Node
		confp        string
		connection   *client.Client
	)

	args = getArgs()

	// set harvest config path
	if confp = os.Getenv("HARVEST_CONF"); confp != "" {
		harvestConfPath = confp
	}

	// connect to cluster and retrieve system version
	if params, err = config.GetPoller(path.Join(harvestConfPath, "harvest.yml"), args.Poller); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if connection, err = client.New(params); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err = connection.Init(2); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("connected to %s%s%s (%s)\n", color.Bold, connection.Name(), color.End, connection.Release())

	// get requested item
	if item, err = get(connection, args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if args.Command == "show" {
		show(item, args)
	} else if err = export(item, connection, args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func get(c *client.Client, args *Args) (*node.Node, error) {
	switch args.Item {
	case "system":
		return getSystem(c)
	case "apis":
		return getApis(c)
	case "objects":
		return getObjects(c)
	case "attrs":
		return getAttrs(c, args)
	case "counters":
		return getCounters(c, args)
	case "counter":
		return getCounter(c, args)
	case "instances":
		return getInstances(c, args)
	case "data":
		return getData(c, args)
	default:
		return nil, errors.New(INVALID_ITEM, args.Item)
	}
}

func getSystem(c *client.Client) (*node.Node, error) {
	return c.InvokeRequestString("system-get-version")
}

func getApis(c *client.Client) (*node.Node, error) {
	var (
		n   *node.Node
		err error
	)

	if n, err = c.InvokeRequestString("system-api-list"); err != nil {
		return nil, err
	}

	if n = n.GetChildS("apis"); n == nil {
		return nil, errors.New(ATTRIBUTE_NOT_FOUND, "apis")
	}
	return n, nil
}

func getObjects(c *client.Client) (*node.Node, error) {
	var (
		n   *node.Node
		err error
	)

	if n, err = c.InvokeRequestString("perf-object-list-info"); err != nil {
		return nil, err
	}

	if n = n.GetChildS("objects"); n == nil {
		return nil, errors.New(ATTRIBUTE_NOT_FOUND, "objects")
	}
	return n, nil
}

func getCounters(c *client.Client, args *Args) (*node.Node, error) {
	var (
		req, n *node.Node
		err    error
	)

	req = node.NewXmlS("perf-object-counter-list-info")
	req.NewChildS("objectname", args.Object)

	if n, err = c.InvokeRequest(req); err != nil {
		return nil, err
	}

	if n = n.GetChildS("counters"); n == nil {
		return nil, errors.New(ATTRIBUTE_NOT_FOUND, "counters")
	}
	return n, nil
}

func getCounter(c *client.Client, args *Args) (*node.Node, error) {
	var (
		counters, cnt *node.Node
		err           error
	)
	if counters, err = getCounters(c, args); err != nil {
		return nil, err
	}

	for _, cnt = range counters.GetChildren() {
		if cnt.GetChildContentS("name") == args.Counter {
			return cnt, nil
		}
	}
	return nil, errors.New(ATTRIBUTE_NOT_FOUND, args.Counter)
}

func getInstances(c *client.Client, args *Args) (*node.Node, error) {
	var (
		req, n *node.Node
		err    error
	)

	req = node.NewXmlS("perf-object-instance-list-info-iter")
	req.NewChildS("objectname", args.Object)

	if args.MaxRecords != 0 {
		req.NewChildS("max-records", strconv.Itoa(args.MaxRecords))
	}

	if n, err = c.InvokeRequest(req); err != nil {
		return nil, err
	}

	if n = n.GetChildS("attributes-list"); n == nil {
		return nil, errors.New(ATTRIBUTE_NOT_FOUND, "attributes-list")
	}
	return n, nil

}

func getData(c *client.Client, args *Args) (*node.Node, error) {

	var req *node.Node

	// requested data is for an Zapi query
	if args.Api != "" {
		req = node.NewXmlS(args.Api)
		// requested data is for a ZapiPerf object
	} else {
		if c.IsClustered() {
			req = node.NewXmlS("perf-object-get-instances")
			instances := req.NewChildS("instances", "")
			instances.NewChildS("instance", "*")
		} else {
			req = node.NewXmlS("perf-object-get-instances")
		}
		req.NewChildS("objectname", args.Object)
	}

	// add custom parameters to request
	for _, p := range args.Parameters {
		if x := strings.Split(p, ":"); len(x) == 2 {
			req.NewChildS(x[0], x[1])
		} else {
			fmt.Printf("Invalid parameter [%s]\n", p)
		}
	}

	return c.InvokeRequest(req)
}
