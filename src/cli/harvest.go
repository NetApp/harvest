package main

import (
	"os"
	"os/exec"
	"path"
	"fmt"
	"goharvest2/share/version"
	"goharvest2/share/options"
)

var PATH = "/opt/harvest2/"

var USAGE = `
NetApp Harvest 2.0 - application for monitoring storage systems

Usage:
    harvest <command> [arguments]

The commands are:

	status                  show status of pollers
	start/restart/stop      manage pollers
	config                  run the config utility
	alerts                  manage alerts
	zapitool                explore ZAPI objects and counters
	grafanatool             import dashboards to Grafana
	version                 show Harvest2 version
	doc                     serve docs over http

Use "harvest <command> help" for more information about a command
`

var COMMANDS = []string{
	"status",
	"start",
	"stop",
	"restart",
	"config",
	"alerts",
	"zapitool",
	"grafanatool",
	"version",
	"help",
	"doc",
}

func main() {

	opts := options.New("", "", "")
	opts.SetHelp(USAGE)

	harvest_path := PATH
	command := ""

	opts.String(&harvest_path, "path", "p", "Harvest installation directory")
	opts.PosString(&command, "command", "command to run", COMMANDS)

	if ! opts.Parse() {
		opts.PrintValues()
		os.Exit(0)
	}

	opts.PrintValues()

	bin := ""

	switch command {
	case "", "help":
		opts.PrintHelp()
	case "status", "start", "restart", "stop":
		bin = "manager"
	case "alerts":
		fmt.Println("alert manager not available....")
	case "config":
		bin = "config"
	case "zapitool":
		bin = "zapitool"
	case "grafanatool":
		bin = "grafanatool"
	case "version":
		fmt.Println(version.VERSION)
	default:
		fmt.Printf("Unknown command: %s\nRun \"harvest help\" for usage\n", command)
	}

	if bin != "" {
		bin_path := path.Join(harvest_path, "bin/", bin)

		var cmd *exec.Cmd
		if bin == "manager" {
			cmd = exec.Command(bin_path, os.Args[1:]...)
		} else {
			cmd = exec.Command(bin_path, os.Args[2:]...)
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Println(cmd.String())
		cmd.Run()
		/*if err := cmd.Run(); err != nil {
			fmt.Printf("error executing [%s]: %v\n", bin_path, err)
			os.Exit(1)
		}*/
	}

}
