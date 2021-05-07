/*
 * Copyright NetApp Inc, 2021 All rights reserved

NetApp Harvest 2.0: the swiss-army-knife for datacenter monitoring

Authors:
   Georg Mey & Vachagan Gratian
Contact:
   ng-harvest-maintainers@netapp.com

This project is based on NetApp Harvest, authored by
Chris Madden in 2015.

*/
package main

import (
	"fmt"
	"goharvest2/cmd/harvest/config"
	"goharvest2/cmd/harvest/manager"
	pkgConfig "goharvest2/pkg/config"
	"goharvest2/cmd/harvest/stub"
	"goharvest2/cmd/harvest/version"
	"os"
	"os/exec"
	"path"
)

var usage = `
NetApp Harvest 2.0 - application for monitoring storage systems

Usage:
    harvest <command> [arguments]

The commands are:

	status                     show status of pollers
	start, stop...             manage pollers
	config                     run the config utility
	zapi                       explore ZAPI objects and counters
	grafana                    import/export Grafana dashboards
	new                        create new collector, plugin or exporter (for developers)
	version                    show version and exit

Use "harvest <command> help" for more information about a command
Use "harvest manager help" for more options on managing pollers
`

func main() {

	var (
		command, bin, harvestPath string
		cmd                       *exec.Cmd
	)

	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	if command == "" || command == "help" || command == "-h" || command == "--help" {
		fmt.Println(usage)
		os.Exit(0)
	}

	harvestPath = pkgConfig.GetHarvestHome()

	switch command {
	case "version":
		fmt.Println(version.String())
	case "manager", "status", "start", "restart", "stop", "kill":
		manager.Run()
	case "config":
		config.Run()
	case "new":
		stub.Run()
	case "zapi":
		bin = "bin/zapi"
	case "grafana":
		bin = "bin/grafana"
	/*
		case "build":
			bin = "cmd/build.sh"
	*/
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	if bin != "" {
		cmd = exec.Command(path.Join(harvestPath, bin), os.Args[2:]...)
		os.Stdout.Sync()
		os.Stdin.Sync()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
