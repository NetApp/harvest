package main

import (
	"os"
	"fmt"
	"strings"
	"goharvest2/share/config"
	"goharvest2/share/dialog"
	"goharvest2/share/tree/node"
)

var USAGE = `
Harvest 2.0 - Config utility

Configure a new poller or exporter

Usage: harvest2 config add ["poller" | "exporter"]

Poller:
  A poller is an Harvest instance for monitoring one single
  storage system. This utility helps you to create a poller 
  for a NetApp System (Cdot or 7Mode). For a custom poller,
  just edit your config.yaml manually.

Exporter:
  An exporter is an interface that forwards data to a database.
  The same exporter can be used by more than one pollers, i.e.
  you need to define only one exporter for each of your DBs.
  This utility helps you to create exporters for three DBs:
  Prometheus, InfluxDB and Graphite
`

var PATH = "/opt/harvest201"

var DIALOG dialog.Dialog

func print_usage() {
	fmt.Println(USAGE)
}


func main() {

	var item string
	var conf *node.Node
	var err error

	fmt.Println("hello")

	if len(os.Args) == 1 {
		print_usage()
		os.Exit(0)
	}

	if command := strings.ReplaceAll(os.Args[1], "-", ""); command == "help" {
		print_usage()
		os.Exit(0)
	} else if command != "add" {
		fmt.Printf("Unknown command: %s\nRun \"harvest2 config help\" for usage\n", command)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Printf("Missing option: \"poller\" or \"exporter\"\n")
		os.Exit(1)
	}

	if item = strings.ReplaceAll(os.Args[2], "-", ""); item != "poller" && item != "exporter" {

		DIALOG.Menu("add", "what do you want to add?", "poller", "exporter")

		item = "poller"
	}

	DIALOG = dialog.New()

	// load config
	if conf, err = config.LoadConfig(PATH, "config.yaml"); err != nil {
		//if DIALOG.YesNo("harvest config", "you don't have existing config file, create new?") {
		if DIALOG.YesNo("config", "_create_new") {
			conf = node.NewS("config")
		} else {
			os.Exit(0)
		}
	}

	//conf.Print(0)

	if item == "poller" {
		add_poller(conf)
	} else {
		add_exporter(conf)
	}
	
}

func add_poller(config *node.Node) bool {

	// load config, if does not exist notify user (create new)

	// ask for address

	// ask for authentication method

	// a. generate ssl certificates

	// b. add existing key/cert

	// ask for confirmation

	// safe / merge
	return true
}

func add_exporter(config *node.Node) bool {
	return true

}