//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package config

import (
	"fmt"
	"goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/argparse"
	"goharvest2/pkg/config"
	"goharvest2/pkg/dialog"
	"goharvest2/pkg/tree"
	"goharvest2/pkg/tree/node"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
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

const (
	HARVEST_USER          = "harvest2-user"
	HARVEST_ROLE          = "harvest2-role"
	PROMETHEUS_PORT_START = 7202
)

var CONF_PATH string
var CONF_FILE string

var DIALOG *dialog.Dialog

func exitError(msg string, err error) {
	DIALOG.Close()
	fmt.Printf("Error (%s): %v\n", msg, err)
	os.Exit(1)
}

func Run() {

	if CONF_PATH = os.Getenv("HARVEST_CONF"); CONF_PATH == "" {
		CONF_PATH = "/etc/harvest/"
	}

	CONF_FILE = path.Join(CONF_PATH, "harvest.yml")

	var item string
	var err error
	var conf, pollers, exporters *node.Node

	parser := argparse.New("Config utility", "harvest config", "configure pollers")
	parser.PosString(&item, "item", "item to configure", []string{"poller", "exporter", "welcome", "help"})
	parser.String(&CONF_FILE, "config", "c", "custom config filepath (default: "+CONF_FILE+")")
	parser.SetHelp(USAGE)

	if !parser.Parse() {
		os.Exit(0)
	}

	if item == "help" {
		fmt.Println(USAGE)
		os.Exit(0)
	}

	if DIALOG = dialog.New(); !DIALOG.Enabled() {
		fmt.Println("This program requires [dialog] or [whiptail].")
		os.Exit(1)
	}

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
		DIALOG.Close()
		os.Exit(0)
	}

	if conf, err = config.LoadConfig(CONF_FILE); err != nil {
		conf = node.NewS("")
	}

	if pollers = conf.GetChildS("Pollers"); pollers == nil {
		pollers = conf.NewChildS("Pollers", "")
	}

	if exporters = conf.GetChildS("Exporters"); exporters == nil {
		exporters = conf.NewChildS("Exporters", "")
	}

	for {

		if item == "" {
			item, err = DIALOG.Menu("Add new:", "poller", "exporter", "safe and exit")
			if err != nil {
				// error means user clicked on Cancel
				item = "exit"
				break
			}
		}

		if item == "poller" {
			if new_poller := add_poller(); new_poller != nil {

				if len(exporters.GetChildren()) == 0 {
					question := "You don't have any exporters defined.\n" +
						"Create Prometheus exporter with default " +
						"parameters and add to this poller?"
					if DIALOG.YesNo(question) {
						prometheus := exporters.NewChildS("prometheus", "")
						prometheus.NewChildS("exporter", "Prometheus")
						prometheus.NewChildS("addr", "0.0.0.0")
						prometheus.NewChildS("master", "True")

						poller_exporters := new_poller.NewChildS("exporters", "")
						poller_exporters.NewChildS("", "prometheus")
						new_poller.NewChildS("prometheus_port", strconv.Itoa(PROMETHEUS_PORT_START))
					}

				} else if len(exporters.GetChildren()) == 1 {
					exporter := exporters.GetChildren()[0]

					question := "Add exporter [" + exporter.GetNameS() + "] to poller?"
					if DIALOG.YesNo(question) {

						poller_exporters := new_poller.NewChildS("exporters", "")
						poller_exporters.NewChildS("", exporter.GetNameS())

						if exporter.GetChildContentS("exporter") == "Prometheus" {
							new_poller.NewChildS("prometheus_port", strconv.Itoa(PROMETHEUS_PORT_START+len(pollers.GetChildren())+1))
						}
					}
				} else {
					choices := make([]string, 0, len(exporters.GetChildren()))

					for _, exp := range exporters.GetChildren() {
						choices = append(choices, exp.GetNameS())
					}
					choices = append(choices, "skip")

					// @TODO allow multiple choices
					item, err = DIALOG.Menu("Choose exporter for this poller:", choices...)

					if item != "skip" {
						if exp := exporters.GetChildS(item); exp != nil {

							poller_exporters := new_poller.NewChildS("exporters", "")
							poller_exporters.NewChildS("", item)

							if exp.GetChildContentS("exporter") == "Prometheus" {
								new_poller.NewChildS("prometheus_port", strconv.Itoa(PROMETHEUS_PORT_START+len(pollers.GetChildren())+1))
							}
						} else {
							DIALOG.Message("You don't have any exporter named [" + item + "].")
						}
					}
				}

				if pollers.GetChildS(new_poller.GetNameS()) == nil {
					pollers.AddChild(new_poller)
				} else if DIALOG.YesNo("poller [" + new_poller.GetNameS() + "] already exists, overwrite?") {
					pollers.AddChild(new_poller)
				}
			} else {
				item = "exit"
			}
		}

		if item == "exporter" {
			if new_exporter := add_exporter(); new_exporter != nil {
				if exporters.GetChildS(new_exporter.GetNameS()) == nil {
					exporters.AddChild(new_exporter)
				} else if DIALOG.YesNo("exporter [" + new_exporter.GetNameS() + "] already exists, overwrite?") {
					exporters.AddChild(new_exporter)
				}
			} else {
				item = "exit"
			}
		}

		if item == "exit" || item == "safe and exit" {
			break
		}

		item = ""
	}

	if item == "safe and exit" {

		use_tmp := false
		fp := CONF_FILE

		dir, fn := path.Split(CONF_FILE)

		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			if os.Mkdir(CONF_PATH, 0644) != nil {
				fp = path.Join("/tmp", fn)
				use_tmp = true
			}
		}

		if err = tree.Export(conf, "yaml", fp); err != nil {
			exitError("export yaml", err)
		}

		msg := "Saved results as [" + fp + "]"
		if use_tmp {
			msg = "You don't have write permissions in [" + CONF_PATH + "]!!\n" +
				"Config file saved as [" + fp + "]. Please move it\n" +
				"to [" + CONF_PATH + "] with a privileged user."
		}
		DIALOG.Message(msg)
	}

	DIALOG.Close()
}

func add_poller() *node.Node {

	poller := node.NewS("")

	// ask for datacenter & address

	datacenter, err := DIALOG.Input("Datacenter name:")
	if err != nil {
		return nil
	}
	poller.NewChildS("datacenter", datacenter)

	addr, err := DIALOG.Input("Enter address (IPv4, IPv6, hostname or URL)")
	if err != nil {
		return nil
	}
	poller.NewChildS("addr", addr)

	// ask for authentication method
	auth, err := DIALOG.Menu("Choose authentication method", "client certificate", "password")
	if err != nil {
		return nil
	}

	create_cert := false

	if auth == "client certificate" {
		if DIALOG.YesNo("Create client certificate and key pair?") {
			if exec.Command("which", "openssl").Run() != nil {
				DIALOG.Message("You don't have openssl installed, please install and try again")
				return nil
			}
			create_cert = true
			DIALOG.Message("This requires one-time admin password to create \na read-only user and install certificate on your system")
		} else {
			msg := fmt.Sprintf("Copy your cert/key pair to [%s/cert/] as [<SYSTEM_NAME>.key] and [<SYSTEM_NAME>.pem] to continue", CONF_PATH)
			DIALOG.Message(msg)
			poller.NewChildS("auth_style", "certificate_auth")
		}
	}

	if auth == "password" || create_cert {
		poller.NewChildS("auth_style", "password")
		username, err := DIALOG.Input("username: ")
		if err != nil {
			return nil
		}
		password, err := DIALOG.Password("password: ")
		if err != nil {
			return nil
		}
		poller.NewChildS("username", username)
		poller.NewChildS("password", password)
	}

	// connect and get system info
	DIALOG.Message("Connecting to system...")

	var client *zapi.Client
	var system *zapi.System

	if client, err = zapi.New(poller); err == nil {
		system, err = client.GetSystem()
	}

	if err != nil {
		if DIALOG.YesNo("Unable to connect to system. Add poller anyway?") {
			name, err := DIALOG.Input("Name of poller / cluster:")
			if err != nil {
				return nil
			}
			poller.SetNameS(name)
		} else {
			return nil
		}
	} else {
		DIALOG.Message("Connected to:\n" + system.String())
		poller.SetNameS(system.Name)
	}

	if err == nil && create_cert {

		cert_path := path.Join(CONF_PATH, "cert", system.Name+".pem")
		key_path := path.Join(CONF_PATH, "cert", system.Name+".key")

		cmd := exec.Command(
			"openssl",
			"req",
			"-x509",
			"-nodes",
			"-days",
			"1095",
			"-newkey",
			"rsa:2048",
			"-keyout",
			key_path,
			"-out",
			cert_path,
			"-subj",
			"/CN="+HARVEST_USER,
		)

		if err := cmd.Run(); err != nil {
			exitError("openssl", err)
		}

		DIALOG.Message(fmt.Sprintf("Generated certificate/key pair:\n  - %s\n  - %s\n", cert_path, key_path))

		req := node.NewXmlS("security-login-role-create")
		req.NewChildS("access-level", "readonly")
		req.NewChildS("command-directory-name", "DEFAULT")
		req.NewChildS("role-name", HARVEST_ROLE)
		req.NewChildS("vserver", system.Name)

		if _, err := client.InvokeRequest(req); err != nil {
			exitError("create role", err)
		}

		req = node.NewXmlS("security-login-create")
		req.NewChildS("application", "ontapi")
		req.NewChildS("authentication-method", "cert")
		req.NewChildS("comment", "readonly user for harvest2")
		req.NewChildS("role-name", HARVEST_ROLE)
		req.NewChildS("user-name", HARVEST_USER)
		req.NewChildS("vserver", system.Name)

		if _, err := client.InvokeRequest(req); err != nil {
			exitError("create user", err)
		}

		DIALOG.Message(fmt.Sprintf("Created read-only user [%s] and role [%s]", HARVEST_USER, HARVEST_ROLE))

		cert_content, err := ioutil.ReadFile(cert_path)
		if err != nil {
			exitError("cert content", err)
		}

		req = node.NewXmlS("security-certificate-install")
		req.NewChildS("cert-name", HARVEST_USER)
		req.NewChildS("certificate", string(cert_content))
		req.NewChildS("type", "client_ca")
		req.NewChildS("vserver", system.Name)

		if _, err := client.InvokeRequest(req); err != nil {
			exitError("install cert", err)
		}

		DIALOG.Message("Certificate installed on system.")

		// forget password immediately
		poller.PopChildS("auth_style")
		poller.PopChildS("username")
		poller.PopChildS("password")

		// new auth parameters
		poller.NewChildS("auth_style", "certificate_auth")
		poller.NewChildS("ssl_cert", cert_path)
		poller.NewChildS("ssl_key", key_path)
	}

	collectors := poller.NewChildS("collectors", "")
	collectors.NewChildS("", "Zapi")
	collectors.NewChildS("", "ZapiPerf")

	return poller
}

func add_exporter() *node.Node {

	exporter := node.NewS("")

	item, err := DIALOG.Menu("Choose exporter type:", "prometheus", "influxdb", "graphite")
	if err != nil {
		return nil
	}
	exporter.NewChildS("exporter", item)

	name, err := DIALOG.Input("Choose name for exporter instance:")
	if err != nil {
		return nil
	}
	exporter.SetNameS(name)

	port, err := DIALOG.Input("Port of the HTTP service:")
	if err != nil {
		exitError("input exporter port", err)
	}
	exporter.NewChildS("port", port)

	if DIALOG.YesNo("Make HTTP serve publicly on your network?\n(Choose no to serve it only on localhst)") {
		exporter.NewChildS("addr", "0.0.0.0")
	} else {
		exporter.NewChildS("addr", "localhost")
	}

	exporter.NewChildS("master", "True")

	return exporter
}
