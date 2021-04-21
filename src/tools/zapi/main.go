package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"strconv"

	client "goharvest2/api/ontapi/zapi"

	"goharvest2/share/tree/node"
	"goharvest2/share/config"
	"goharvest2/share/errors"
	"goharvest2/share/util"
)

var (
	MAX_SEARCH_DEPTH = 10
	CONFPATH = "/etc/harvest"
)


func main() {

	var (
		err error
		args *Args
		item, params *node.Node
		confp string
		connection *client.Client
		system *client.System
	)

	args = get_args()

	// set harvest config path
	if confp = os.Getenv("HARVEST_CONF"); confp != "" {
		CONFPATH = confp
	}

	// connect to cluster and retrieve system version
	if params, err = config.GetPoller(path.Join(CONFPATH, "harvest.yml"), args.Poller); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if connection, err = client.New(params); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if system, err = connection.GetSystem(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("connected to %s%s%s (%s)\n", util.Bold, system.Name, util.End, system.Release)

	// get requested item
	if item, err = get(connection, args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if args.Command == "show" {
		show(item, args)
	} else if err =	export(item, connection, args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}


func get(c *client.Client, args *Args) (*node.Node, error) {
	switch args.Item {
	case "system":
		return get_system(c)
	case "apis":
		return get_apis(c)
	case "objects":
		return get_objects(c)
	case "attrs":
		return get_attrs(c, args)
	case "counters":
		return get_counters(c, args)
	case "counter":
		return get_counter(c, args)
	case "instances":
		return get_instances(c, args)
	case "data":
		return get_data(c, args)
	default:
		return nil, errors.New(INVALID_ITEM, args.Item)
	}
}

func get_system(c *client.Client) (*node.Node, error) {
	return c.InvokeRequestString("system-get-version")
}

func get_apis(c *client.Client) (*node.Node, error) {
	var (
		n *node.Node
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

func get_objects(c *client.Client) (*node.Node, error) {
	var (
		n *node.Node
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

func get_counters(c *client.Client, args *Args) (*node.Node, error) {
	var (
		req, n *node.Node
		err error
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

func get_counter(c *client.Client, args *Args) (*node.Node, error) {
	var (
		counters, cnt *node.Node
		err error
	)
	if counters, err = get_counters(c, args); err != nil {
		return nil, err
	}

	for _, cnt = range counters.GetChildren() {
		if cnt.GetChildContentS("name") == args.Counter {
			return cnt, nil
		}
	}
	return nil, errors.New(ATTRIBUTE_NOT_FOUND, args.Counter)
}

func get_instances(c *client.Client, args *Args) (*node.Node, error) {
	var (
		req, n *node.Node
		err error
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

func get_data(c *client.Client, args *Args) (*node.Node, error) {
	
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
