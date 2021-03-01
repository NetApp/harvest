package main

import (
	"os"
	"os/exec"
	"path"
    "path/filepath"
	"fmt"
    "strings"
    "io/ioutil"
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
	alerts                  manage alerts
	zapitool                explore ZAPI objects and counters
	grafanatool             import dashboards to Grafana
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

    if cwd, err := os.Getwd(); err == nil {
        fmt.Println("CWD: ", cwd)
    } else {
        fmt.Println("Getwed() ", err)
    }

    if abs, err := filepath.Abs("."); err == nil {
        fmt.Println("ABS: ", abs)
    } else {
        fmt.Println("Abs() ", err)
    }

    var harvest_path string
	if harvest_path = os.Getenv("HARVEST_HOME"); harvest_path == "" {
        harvest_path = "/opt/harvest/"
    }
	// very dirty way of "sourcing" path variables
	// temporary 
    if data, err := ioutil.ReadFile(path.Join(harvest_path, "sources.sh")); err == nil {
        for _, line := range strings.Split(string(data), "\n") {
            //line = strings.TrimSpace(line)
            if ! strings.HasPrefix(line, "#") {
                s := strings.Split(line, " ")
                if len(s) == 2 && s[0] == "export" {
                    v := strings.Split(s[1], "=")
                    if len(v) == 2 {
                        os.Setenv(v[0], strings.ReplaceAll(v[1], "\"", ""))
                        //fmt.Printf("%s ==> %s\n", v[0], v[1])
                    }
                }
            }
        }
    } else {
        fmt.Println(err)
    }

    //fmt.Printf("HARVEST_HOME = %s\n", harvest_path)
    //fmt.Printf("HARVEST_CONF = %s\n", os.Getenv("HARVEST_CONF"))
    //fmt.Printf("HARVEST_LOGS = %s\n", os.Getenv("HARVEST_LOGS"))
    //fmt.Printf("HARVEST_PIDS = %s\n", os.Getenv("HARVEST_PIDS"))

	var bin string

	switch command {
	case "status", "start", "restart", "stop":
		bin = "manager"
	case "alerts":
		fmt.Println("alert manager not available.")
	case "config":
		bin = "config"
	case "zapitool":
		bin = "zapitool"
	case "grafanatool":
		bin = "grafanatool"
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

        //fmt.Println(cmd.String())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

	}

}
