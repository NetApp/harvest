package main

import (
	"os"
	"os/exec"
	"path"
	"fmt"
	"goharvest2/share/version"
)

var USAGE = `
NetApp Harvest 2.0 - application for monitoring storage systems

Usage:
    harvest <command> [arguments]

The commands are:

	status                  show status of pollers
	start/restart/stop      manage pollers
	config                  run the config utility
	zapi                    explore ZAPI objects and counters
	grafan                  import dashboards to Grafana
	version                 show Harvest2 version

Use "harvest <command> help" for more information about a command
`

func main() {

	command := ""
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	if command == "" || command == "help" || command == "-h" || command == "--help" {
		fmt.Println(USAGE)
		os.Exit(0)
	}

	if command == "version" {
		fmt.Println(version.VERSION)
		os.Exit(0)
	}

    var harvest_path string
	if harvest_path = os.Getenv("HARVEST_HOME"); harvest_path == "" {
        harvest_path = "/opt/harvest/"
    }

	var bin string

	switch command {
	case "status", "start", "restart", "stop":
		bin = "manager"
	case "alerts":
		fmt.Println("alert manager not available.")
	case "config":
		bin = "config"
	case "zapi":
		bin = "zapi"
	case "grafana":
		bin = "grafana"
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}

	if bin != "" {

		var cmd *exec.Cmd

		if bin == "manager" {
			cmd = exec.Command(path.Join(harvest_path, "bin/", bin), os.Args[1:]...)
		} else {
			cmd = exec.Command(path.Join(harvest_path, "bin/", bin), os.Args[2:]...)
		}

        os.Stdout.Sync()
        os.Stdin.Sync()
        //fmt.Println(cmd.String())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

	}

}
