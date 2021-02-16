package main

import (
	"os"
	"os/exec"
	"path"
	"fmt"
	"strings"
)

var VERSION = "2.0.1"

var PATH = "/opt/harvest2/"


var USAGE = `
NetApp Harvest 2.0 - application for monitoring storage systems

Usage:
    harvest2 <command> [arguments]

The commands are:

	status                  show status of pollers
	start/restart/stop      manage pollers
	config                  run the config utility
	alerts                  manage alerts
	zapi-tool               explore ZAPI objects and counters
	grafana-tool            import dashboards to Grafana
	version                 show Harvest2 version

Use "harvest2 <command> help" for more information about a command
`
func print_usage() {
	fmt.Println(USAGE)
}

func get_opt(flag_long, flag_short, default_val string) string {
	val := default_val
	for i:=1; i<len(os.Args); i+=1 {
		if (os.Args[i] == "-" + flag_short) || (os.Args[i] == "--" + flag_long) {
			if i+1 < len(os.Args) {
				val = os.Args[i+1]
				break
			}
		}
	}
	return val
}

func main() {

	h, _ := os.Hostname()
	c, _ := os.Getwd()
	p := get_opt("path", "p", PATH)

	fmt.Printf("host=%s cwd=%s path=%s\n", h, c, p)

	if len(os.Args) == 1 {
		print_usage()
		os.Exit(0)
	}

	command := strings.ReplaceAll(os.Args[1], "-", "")
	bin := ""

	switch command {
	case "status", "start", "restart", "stop":
		bin = "manager"
	case "alerts":
		fmt.Println("starting alert manager....")
	case "config":
		fmt.Printf("ready to configure harvest?")
	case "version":
		fmt.Printf("NetApp Harvest 2.0 - Version %s\n", VERSION)
	default:
		fmt.Printf("Unknown command: %s\nRun \"harvest2 help\" for usage\n", command)
	}

	if bin != "" {
		cmd := exec.Command(path.Join(p, "bin/", bin), os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Println("OK")
		}
	}
}