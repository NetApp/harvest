/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package zapi

import (
	"fmt"
	client "github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

var (
	defaultTimeout  = "1m"
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
	API string
	// which attr to show (when Item is "attrs")
	Attr string
	// which object to show (when Item is "object")
	Object string
	// which counter(s) to show when "show data" is called
	Counters   []string
	Instance   string
	Duration   int
	MaxRecords int
	// additional parameters to add to the ZAPI request, in "key:value" format
	Parameters   []string
	Config       string // filepath of Harvest config (defaults to "harvest.yml") can be relative or absolute path
	OutputFormat string
	Timeout      string
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

func validateArgs(slice []string) {
	ok := true
	if len(slice) == 0 {
		args.Item = ""
	} else {
		args.Item = slice[0]
	}

	if args.Item == "data" && args.API == "" && args.Object == "" {
		fmt.Println("show data: requires --api or --object")
		ok = false
	}

	if args.Item == "attrs" && args.API == "" {
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

	if args.Item == "counter" && (args.Object == "" || len(args.Counters) == 0) {
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

	_, err = conf.LoadHarvestConfig(args.Config)
	if err != nil {
		log.Fatal(err)
	}
	// connect to a cluster and retrieve the system version
	if poller, err = conf.PollerNamed(args.Poller); err != nil {
		log.Fatal(err)
	}
	if connection, err = client.New(poller, auth.NewCredentials(poller, slog.Default())); err != nil {
		log.Fatal(err)
	}

	if err = connection.Init(2, conf.Remote{}); err != nil {
		log.Fatal(err)
	}

	connection.SetTimeout(args.Timeout)

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
		return nil, errs.New(errs.ErrInvalidItem, args.Item)
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
		return nil, errs.New(errs.ErrAttributeNotFound, "apis")
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
		return nil, errs.New(errs.ErrAttributeNotFound, "objects")
	}
	return n, nil
}

func getCounters(c *client.Client, args *Args) (*node.Node, error) {
	var (
		req, n *node.Node
		err    error
	)

	req = node.NewXMLS("perf-object-counter-list-info")
	req.NewChildS("objectname", args.Object)

	if n, err = c.InvokeRequest(req); err != nil {
		return nil, err
	}

	if n = n.GetChildS("counters"); n == nil {
		return nil, errs.New(errs.ErrAttributeNotFound, "counters")
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

	result := node.NewS("counters")
	for _, cnt = range counters.GetChildren() {
		for _, counter := range args.Counters {
			if cnt.GetChildContentS("name") == counter {
				result.AddChild(cnt)
			}
		}
	}
	if len(result.GetChildren()) == 0 {
		return nil, errs.New(errs.ErrAttributeNotFound, strings.Join(args.Counters, ","))
	}
	return result, nil
}

func getInstances(c *client.Client, args *Args) (*node.Node, error) {
	var (
		req, n *node.Node
		err    error
	)

	req = node.NewXMLS("perf-object-instance-list-info-iter")
	req.NewChildS("objectname", args.Object)

	if c.IsClustered() && args.MaxRecords != 0 {
		req.NewChildS("max-records", strconv.Itoa(args.MaxRecords))
	}

	if n, err = c.InvokeRequest(req); err != nil {
		return nil, err
	}

	if n = n.GetChildS("attributes-list"); n == nil {
		return nil, errs.New(errs.ErrAttributeNotFound, "attributes-list")
	}
	return n, nil

}

func getData(c *client.Client, args *Args) (*node.Node, error) {

	var req *node.Node

	// requested data is for an Zapi query
	if args.API != "" {
		req = node.NewXMLS(args.API)
		if c.IsClustered() && args.MaxRecords != 0 {
			req.NewChildS("max-records", strconv.Itoa(args.MaxRecords))
		}
		// requested data is for a ZapiPerf object
	} else {
		if c.IsClustered() {
			req = node.NewXMLS("perf-object-get-instances")
			instances := req.NewChildS("instances", "")
			instances.NewChildS("instance", "*")
			if args.MaxRecords != 0 {
				req.NewChildS("max", strconv.Itoa(args.MaxRecords))
			}
		} else {
			req = node.NewXMLS("perf-object-get-instances")
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

	if len(args.Counters) > 0 {
		counters := req.NewChildS("counters", "")
		for _, counter := range args.Counters {
			counters.NewChildS("counter", counter)
		}
	}
	return c.InvokeRequest(req)
}

var args = &Args{}

func init() {
	configPath := conf.Path(conf.HarvestYML)
	Cmd.AddCommand(showCmd, exportCmd)
	flags := Cmd.PersistentFlags()
	flags.StringVarP(&args.Poller, "poller", "p", "", "name of poller (cluster), as defined in your harvest config")
	_ = Cmd.MarkPersistentFlagRequired("poller")

	flags.StringVarP(&args.OutputFormat, "write", "w", "xml",
		fmt.Sprintf("Output format to use: one of %s", outputFormats))
	flags.StringVarP(&args.API, "api", "a", "", "ZAPI query to show")
	flags.StringVarP(&args.Attr, "attr", "t", "", "ZAPI attribute to show")
	flags.StringVarP(&args.Object, "object", "o", "", "ZapiPerf object to show")
	flags.StringSliceVarP(&args.Counters, "counter", "c", []string{}, "ZapiPerf counter(s) to show. Can be specified multiple times")
	flags.IntVarP(&args.MaxRecords, "max-records", "m", 100, "max instances per API request")
	flags.IntVar(&args.MaxRecords, "max", 100, "max instances per API request (Deprecated: Use --max-records instead)")
	_ = flags.MarkDeprecated("max", "Please use --max-records instead")
	flags.StringSliceVarP(&args.Parameters, "parameters", "r", []string{}, "parameter to add to the ZAPI query")
	flags.StringVar(&args.Config, "config", configPath, "harvest config file path")
	flags.StringVar(&args.Timeout, "timeout", defaultTimeout, "Go duration how long to wait for server responses")

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
  harvest zapi --poller aff-900 show data --object lun \
         --counter avg_read_latency --counter read_ops                    Query cluster aff-900 and print performance counters for average
                                                                          read latency and number of read operations per second on each LUN
`)
}
