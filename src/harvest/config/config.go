package main

import (
	"os"
	"fmt"
	"strings"
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


func print_usage() {
	fmt.Println(USAGE)
}


func main() {
	if len(os.Args) == 1 {
		print_usage()
		os.Exit(0)
	}

	if command := strings.ReplaceAll(os.Args[1], "-", ""); command != "add" {
		fmt.Printf("Unknown command: %s\nRun \"harvest2 config help\" for usage\n", command)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Printf("Missing option: \"poller\" or \"exporter\"\n")
		os.Exit(1)
	}

	if option := strings.ReplaceAll(os.Args[2], "-", ""); option != "poller" && option != "exporter" {
		fmt.Printf("Unkown option: %s\nExpected \"poller\" or \"exporter\"\n", option)
		os.Exit(1)
	} else {
		fmt.Printf("adding new %s ......\n", option)
	}
}