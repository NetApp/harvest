package main

import (
	"os"
	"fmt"
	"strings"
    "path"
	"goharvest2/share/config"
	"goharvest2/share/dialog"
	"goharvest2/share/tree"
	"goharvest2/share/tree/node"
    "goharvest2/poller/api/zapi"
)

var USAGE = `
Harvest 2.0 - Config utility

Configure a new poller or exporter

Usage: harvest2 config ["poller" | "exporter"]

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

//var PATH = "/opt/harvest201"
var PATH = "/home/imandes0/GoCode/goharvest2"

var DIALOG *dialog.Dialog

func print_usage() {
	fmt.Println(USAGE)
}

func main() {

	var item string
    var err error
	var conf, params, pollers, exporters *node.Node

	if len(os.Args) > 1 {
		item = strings.ReplaceAll(os.Args[1], "-", "")
	}

	if item == "help" {
		print_usage()
		os.Exit(0)
	}

	DIALOG = dialog.New()

	if item == "welcome" {

        DIALOG.SetTitle("harvest 2.0 - welcome")
		DIALOG.Message("Your installation is complete. Welcome to Harvest 2.0!")

		if DIALOG.YesNo("Do you want to quickly configure Harvest?") {
			item = ""
		} else {
		    item = "exit"
        }
	}

	DIALOG.SetTitle("harvest 2.0 - config")

    if item == "exit" {
        DIALOG.Message("Bye! If you want my help next time, run: \"harvest config\"")
    }

	if item == "" {
		item, err = DIALOG.Menu("Add new:", "poller", "exporter")
		if err != nil {
            exitError("menu add new", err)
        }
	}

    if item == "poller" {
        params = add_poller()
    } else if item == "exporter" {
        params = add_exporter()
    }

    if conf, err = config.LoadConfig(PATH, "config.yaml"); err != nil {
        conf = node.NewS("")
    }

    if item == "poller" {
        if pollers = conf.GetChildS("Pollers"); pollers == nil {
            pollers = conf.NewChildS("Pollers", "")
        }
        pollers.AddChild(params)
    } else if item == "exporters" {
        if exporters = conf.GetChildS("Exporters"); exporters == nil {
            exporters = conf.NewChildS("Exporters", "")
        }
        exporters.AddChild(params)
    }

    fp := path.Join(PATH, "config.yaml")
    if err = tree.ExportYaml(conf, fp); err != nil {
        exitError("export yaml", err)
    }
    DIALOG.Message(fmt.Sprintf("Saved results to:\n[%s]", fp))
    DIALOG.Close()
}

func exitError(msg string, err error) {
    DIALOG.Close()
    fmt.Println(msg)
    fmt.Println(err)
    os.Exit(1)
}

func add_poller() *node.Node {

    poller := node.NewS("")

	// ask for address
    addr, err := DIALOG.Input("Enter address (IPv4, IPv6, hostname or URL)")
    if err != nil {
        exitError("input addr", err)
    }
    poller.NewChildS("url", addr)

    // ask for authentication method
    auth, err := DIALOG.Menu("Choose authentication method", "password", "certificate_auth")
    if err != nil {
        exitError("menu auth", err)
    }
    poller.NewChildS("auth_style", auth)

    if auth == "password" {
        username, _ := DIALOG.Input("username: ")
        password, _ := DIALOG.Password("password: ")
        poller.NewChildS("username", username)
        poller.NewChildS("password", password)

    }

    // connect and get system info
    client, err := zapi.New(poller)
    if err != nil {
        exitError("client", err)
    }

    system, err := client.GetSystem()
    if err != nil {
        DIALOG.Message("Failed to connect to system. Are you sure your credentials are correct?")
        //exitError("system", err)
    } else {
        poller.SetNameS(system.Name)
        DIALOG.Message("Connected to:\n" + system.String())
    }

	// a. generate ssl certificates

	// b. add existing key/cert

	// ask for confirmation

	// safe / merge
	return poller
}

func add_exporter() *node.Node {
	return nil

}
