package main

import (
	"fmt"
	"goharvest2/share/argparse"
	"os"
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
	Parameters []string
}

func get_args() *Args {

	// define arguments
	args := &Args{}
	parser := argparse.New("Zapi Utility", "harvest zapi", "Explore available ZAPI counters of an ONTAP system")

	parser.PosString(
		&args.Command,
		"command",
		"command",
		[]string{"show", "export"},
	)

	parser.PosString(
		&args.Item,
		"item",
		"item to show",
		[]string{"data", "apis", "attrs", "objects", "periodic", "instances", "counters", "counter", "system"},
	)

	parser.String(
		&args.Poller,
		"poller",
		"p",
		"name of poller (cluster), as defined in your harvest config",
	)

	parser.String(
		&args.Api,
		"api",
		"a",
		"API (Zapi query) to show",
	)

	parser.String(
		&args.Attr,
		"attr",
		"t",
		"Zapi attribute to show",
	)

	parser.String(
		&args.Object,
		"object",
		"o",
		"ZapiPerf object to show",
	)

	parser.String(
		&args.Counter,
		"counter",
		"c",
		"ZapiPerf counter to show",
	)

	parser.Slice(
		&args.Counters,
		"counters",
		"l",
		"list of counters for periodic show",
	)

	parser.String(
		&args.Instance,
		"instance",
		"i",
		"instance for which to show periodic data",
	)

	parser.Int(
		&args.Duration,
		"duration",
		"d",
		"duration/interval to show periodic data",
	)

	args.MaxRecords = 100
	parser.Int(
		&args.MaxRecords,
		"max",
		"m",
		"max-records: max instances per API request (default: 100)",
	)

	parser.Slice(
		&args.Parameters,
		"parameters",
		"r",
		"parameter to add to the ZAPI query",
	)

	parser.SetHelpFlag("help")

	if !parser.Parse() {
		os.Exit(0)
	}

	// validate and warn on missing arguments
	ok := true

	if args.Poller == "" {
		fmt.Println("missing required argument: --poller")
		ok = false
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

	if args.Item == "periodic" && len(args.Counters) == 0 {
		fmt.Println("show periodic: requires --counters (list of counters)")
		ok = false
	}

	if args.Item == "periodic" && args.Instance == "" {
		fmt.Println("show periodic requires --instance (name or uuid)")
		ok = false
	}

	if args.Item == "periodic" && args.Duration == 0 {
		//fmt.Println("show periodic: using default interval [30 s]")
		args.Duration = 30
	}

	if !ok {
		os.Exit(1)
	}

	return args
}
