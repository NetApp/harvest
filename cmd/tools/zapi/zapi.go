/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package zapi

import (
	"fmt"
	"github.com/spf13/cobra"
	client "goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/color"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/tree/node"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	AttributeNotFound = "attribute not found"
	InvalidItem       = "invalid item"
)

var (
	maxSearchDepth  = 5
	validShowArgs   = []string{"data", "apis", "attrs", "objects", "instances", "counters", "counter", "system"}
	validExportArgs = []string{"attrs", "counters"}
	outputFormats   = []string{"xml", "color"}
)

type Args struct {
	// main command: show, export, etc
	Command string
	// second command: what to show, export,
	Item string
	// poller that contains parameters of an Ontap cluster
	Poller string
	// which API to show (when Item is "api")
	Api string
	// which attr to show (when Item is "attrs")
	Attr string
	// which object to show (when Item is "object")
	Object string
	// which counter to show (when Item is "counter")
	Counter string
	// ???
	Counters []string
	// ??
	Instance string
	// ??
	Duration int
	// ??
	MaxRecords int
	// additional parameters to add to the ZAPI request, in "key:value" format
	Parameters   []string
	Config       string // filepath of Harvest config (defaults to "harvest.yml") can be relative or absolute path
	OutputFormat string
}

var Cmd = &cobra.Command{
	Use:   "zapi",
	Short: "Zapi Utility",
	Long:  "Zapi Utility - Explore available ZAPI counters of an ONTAP system",
}

var exportCmd = &cobra.Command{
	Use:       "export",
	Short:     "item to export, one of: " + strings.Join(validExportArgs, ", "),
	Args:      cobra.OnlyValidArgs,
	ValidArgs: validExportArgs,
	Run:       doExport,
}

var showCmd = &cobra.Command{
	Use:       "show",
	Short:     "item to show, one of: " + strings.Join(validShowArgs, ", "),
	Args:      cobra.OnlyValidArgs,
	ValidArgs: validShowArgs,
	Run:       doShow,
}

func doExport(_ *cobra.Command, a []string) {
	validateArgs(a)
	doCmd("export")
}

func doShow(_ *cobra.Command, a []string) {
	validateArgs(a)
	doCmd("show")
}

func validateArgs(strings []string) {
	ok := true
	if len(strings) == 0 {
		args.Item = ""
	} else {
		args.Item = strings[0]
	}

	if args.Item == "data" && args.Api == "" && args.Object == "" {
		fmt.Println("show data: requires --api or --object")
		ok = false
	}

	if args.Item == "attrs" && args.Api == "" {
		fmt.Println("show attrs: requires --api")
		ok = false
	}

	if args.Item == "counters" && args.Object == "" {
		fmt.Println("show counters: requires --object")
		ok = false
	}

	if args.Item == "instances" && args.Object == "" {
		fmt.Println("show instances: requires --object")
		ok = false
	}

	if args.Item == "counter" && (args.Object == "" || args.Counter == "") {
		fmt.Println("show counter: requires --object and --counter")
		ok = false
	}

	if !ok {
		os.Exit(1)
	}
}

func doCmd(cmd string) {
	var (
		err        error
		item       *node.Node
		poller     *conf.Poller
		connection *client.Client
	)

	err = conf.LoadHarvestConfig(args.Config)
	if err != nil {
		log.Fatal(err)
	}
	// connect to cluster and retrieve system version
	if poller, err = conf.PollerNamed(args.Poller); err != nil {
		log.Fatal(err)
	}
	if connection, err = client.New(*poller); err != nil {
		log.Fatal(err)
	}

	if err = connection.Init(2); err != nil {
		log.Fatal(err)
	}

	color.DetectConsole("")
	_, _ = fmt.Fprintf(os.Stderr, "connected to %s%s%s (%s)\n", color.Bold, connection.Name(), color.End, connection.Release())

	// get requested item
	if item, err = get(connection, args); err != nil {
		log.Fatal(err)
	}

	if cmd == "show" {
		show(item, args)
	} else if err = export(item, connection, args); err != nil {
		log.Fatal(err)
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
		return nil, errors.New(InvalidItem, args.Item)
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
		return nil, errors.New(AttributeNotFound, "apis")
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
		return nil, errors.New(AttributeNotFound, "objects")
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
		return nil, errors.New(AttributeNotFound, "counters")
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
	return nil, errors.New(AttributeNotFound, args.Counter)
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
		return nil, errors.New(AttributeNotFound, "attributes-list")
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

	if args.Counter != "" {
		counters := req.NewChildS("counters", "")
		counters.NewChildS("counter", args.Counter)
	}
	return c.InvokeRequest(req)
}

var args = &Args{}

func init() {
	configPath := conf.GetDefaultHarvestConfigPath()

	Cmd.AddCommand(showCmd, exportCmd)
	flags := Cmd.PersistentFlags()
	flags.StringVarP(&args.Poller, "poller", "p", "", "name of poller (cluster), as defined in your harvest config")
	_ = Cmd.MarkPersistentFlagRequired("poller")

	flags.StringVarP(&args.OutputFormat, "write", "w", "xml",
		fmt.Sprintf("Output format to use: one of %s", outputFormats))
	flags.StringVarP(&args.Api, "api", "a", "", "ZAPI query to show")
	flags.StringVarP(&args.Attr, "attr", "t", "", "ZAPI attribute to show")
	flags.StringVarP(&args.Object, "object", "o", "", "ZapiPerf object to show")
	flags.StringVarP(&args.Counter, "counter", "c", "", "ZapiPerf counter to show")
	flags.IntVarP(&args.MaxRecords, "max", "m", 100, "max-records: max instances per API request")
	flags.StringSliceVarP(&args.Parameters, "parameters", "r", []string{}, "parameter to add to the ZAPI query")
	flags.StringVar(&args.Config, "config", configPath, "harvest config file path")

	showCmd.SetUsageTemplate("item to show should be one of: " + strings.Join(validShowArgs, ", "))

	// Append usage examples
	Cmd.SetUsageTemplate(Cmd.UsageTemplate() + `
Examples:
  harvest zapi -p infinity show apis                                      Query cluster infinity for available APIs
  harvest zapi -p infinity show attrs --api volume-get-iter               Query cluster infinity for volume-get-iter metrics
                                                                          Typically APIs suffixed with 'get-iter' have interesting metrics 
  harvest zapi -p infinity show data --api volume-get-iter                Query cluster infinity and print attribute tree of volume-get-iter
  harvest zapi -p infinity show counters --object workload_detail_volume  Query cluster infinity and print performance counter metadata 
  harvest zapi -p infinity show data --object qtree --counter nfs_ops     Query cluster infinity and print performance counters on the 
                                                                          number of NFS operations per second on each qtree
`)
}
