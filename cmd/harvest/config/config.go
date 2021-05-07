/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
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

var usage = `
Harvest 2.0 - Config utility

Configure a new poller or exporter

Usage: harvest config ["poller" | "exporter"]

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
	// if we create user/role on ONTAP use these names
	harvestUserName = "harvest2-user"
	harvestRoleName = "harvest2-role"
	// if we add Promtheus port to config start from this
	PrometheusPortStart = 7202
)

var (
	harvestConfPath string
	harvestConfFile string
	_dialog         *dialog.Dialog
)

func exitError(msg string, err error) {
	_dialog.Close()
	fmt.Printf("Error (%s): %v\n", msg, err)
	os.Exit(1)
}

func Run() {

	harvestConfPath = config.GetHarvestConf()
	harvestConfFile = path.Join(harvestConfPath, "harvest.yml")

	var item string
	var err error
	var conf, pollers, exporters *node.Node

	parser := argparse.New("Config utility", "harvest config", "configure pollers")
	parser.PosString(&item, "item", "item to configure", []string{"poller", "exporter", "welcome", "help"})
	parser.String(&harvestConfFile, "config", "c", "custom config filepath (default: "+harvestConfFile+")")
	parser.SetHelp(usage)
	parser.SetHelpFlag("help")
	parser.SetOffset(2)

	parser.ParseOrExit()

	if _dialog = dialog.New(); !_dialog.Enabled() {
		fmt.Println("This program requires [dialog] or [whiptail].")
		os.Exit(1)
	}

	if item == "welcome" {

		_dialog.SetTitle("harvest 2.0 - welcome")
		_dialog.Message("Your installation is complete. Welcome to Harvest 2.0!")

		if _dialog.YesNo("Do you want to quickly configure Harvest?") {
			item = ""
		} else {
			item = "exit"
		}
	}

	_dialog.SetTitle("harvest 2.0 - config")

	if item == "exit" {
		_dialog.Message("Bye! If you want my help next time, run: \"harvest config\"")
		_dialog.Close()
		os.Exit(0)
	}

	if conf, err = config.LoadConfig(harvestConfFile); err != nil {
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
			item, err = _dialog.Menu("Add new:", "poller", "exporter", "safe and exit")
			if err != nil {
				// error means user clicked on Cancel
				item = "exit"
				break
			}
		}

		if item == "poller" {
			if newPoller := addPoller(); newPoller != nil {

				if len(exporters.GetChildren()) == 0 {
					question := "You don't have any exporters defined.\n" +
						"Create Prometheus exporter with default " +
						"parameters and add to this poller?"
					if _dialog.YesNo(question) {
						prometheus := exporters.NewChildS("prometheus", "")
						prometheus.NewChildS("exporter", "Prometheus")
						prometheus.NewChildS("addr", "0.0.0.0")
						prometheus.NewChildS("master", "True")

						pollerExporters := newPoller.NewChildS("exporters", "")
						pollerExporters.NewChildS("", "prometheus")
						newPoller.NewChildS("prometheus_port", strconv.Itoa(PrometheusPortStart))
					}

				} else if len(exporters.GetChildren()) == 1 {
					exporter := exporters.GetChildren()[0]

					question := "Add exporter [" + exporter.GetNameS() + "] to poller?"
					if _dialog.YesNo(question) {

						pollerExporters := newPoller.NewChildS("exporters", "")
						pollerExporters.NewChildS("", exporter.GetNameS())

						if exporter.GetChildContentS("exporter") == "Prometheus" {
							newPoller.NewChildS("prometheus_port", strconv.Itoa(PrometheusPortStart+len(pollers.GetChildren())+1))
						}
					}
				} else {
					choices := make([]string, 0, len(exporters.GetChildren()))

					for _, exp := range exporters.GetChildren() {
						choices = append(choices, exp.GetNameS())
					}
					choices = append(choices, "skip")

					// @TODO allow multiple choices
					item, err = _dialog.Menu("Choose exporter for this poller:", choices...)

					if item != "skip" {
						if exp := exporters.GetChildS(item); exp != nil {

							pollerExporters := newPoller.NewChildS("exporters", "")
							pollerExporters.NewChildS("", item)

							if exp.GetChildContentS("exporter") == "Prometheus" {
								newPoller.NewChildS("prometheus_port", strconv.Itoa(PrometheusPortStart+len(pollers.GetChildren())+1))
							}
						} else {
							_dialog.Message("You don't have any exporter named [" + item + "].")
						}
					}
				}

				if pollers.GetChildS(newPoller.GetNameS()) == nil {
					pollers.AddChild(newPoller)
				} else if _dialog.YesNo("poller [" + newPoller.GetNameS() + "] already exists, overwrite?") {
					pollers.AddChild(newPoller)
				}
			} else {
				item = "exit"
			}
		}

		if item == "exporter" {
			if newExporter := addExporter(); newExporter != nil {
				if exporters.GetChildS(newExporter.GetNameS()) == nil {
					exporters.AddChild(newExporter)
				} else if _dialog.YesNo("exporter [" + newExporter.GetNameS() + "] already exists, overwrite?") {
					exporters.AddChild(newExporter)
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

		useTmp := false
		fp := harvestConfFile

		dir, fn := path.Split(harvestConfFile)

		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			if os.Mkdir(harvestConfPath, 0644) != nil {
				fp = path.Join("/tmp", fn)
				useTmp = true
			}
		}

		if err = tree.Export(conf, "yaml", fp); err != nil {
			exitError("export yaml", err)
		}

		msg := "Saved results as [" + fp + "]"
		if useTmp {
			msg = "You don't have write permissions in [" + harvestConfPath + "]!!\n" +
				"Config file saved as [" + fp + "]. Please move it\n" +
				"to [" + harvestConfPath + "] with a privileged user."
		}
		_dialog.Message(msg)
	}

	_dialog.Close()
}

func addPoller() *node.Node {

	var (
		client *zapi.Client
		err    error
	)

	poller := node.NewS("")

	// ask for datacenter & address

	datacenter, err := _dialog.Input("Datacenter name:")
	if err != nil {
		return nil
	}
	poller.NewChildS("datacenter", datacenter)

	addr, err := _dialog.Input("Enter address (IPv4, IPv6, hostname or URL)")
	if err != nil {
		return nil
	}
	poller.NewChildS("addr", addr)

	// ask for authentication method
	auth, err := _dialog.Menu("Choose authentication method", "client certificate", "password")
	if err != nil {
		return nil
	}

	createCert := false

	if auth == "client certificate" {
		if _dialog.YesNo("Create client certificate and key pair?") {
			if exec.Command("which", "openssl").Run() != nil {
				_dialog.Message("You don't have openssl installed, please install and try again")
				return nil
			}
			createCert = true
			_dialog.Message("This requires one-time admin password to create \na read-only user and install certificate on your system")
		} else {
			msg := fmt.Sprintf("Copy your cert/key pair to [%s/cert/] as [<SYSTEM_NAME>.key] and [<SYSTEM_NAME>.pem] to continue", harvestConfPath)
			_dialog.Message(msg)
			poller.NewChildS("auth_style", "certificate_auth")
		}
	}

	if auth == "password" || createCert {
		poller.NewChildS("auth_style", "password")
		username, err := _dialog.Input("username: ")
		if err != nil {
			return nil
		}
		password, err := _dialog.Password("password: ")
		if err != nil {
			return nil
		}
		poller.NewChildS("username", username)
		poller.NewChildS("password", password)
	}

	// connect and get system info
	_dialog.Message("Connecting to system...")

	if client, err = zapi.New(poller); err != nil {
		exitError("client", err)
	}

	if err = client.Init(5); err != nil {
		if _dialog.YesNo("Unable to connect to system. Add poller anyway?") {
			if name, err := _dialog.Input("Name of poller / cluster:"); err != nil {
				return nil
			} else {
				poller.SetNameS(name)
			}
		} else {
			return nil
		}
	} else {
		_dialog.Message("Connected to:\n" + client.Info())
		poller.SetNameS(client.Name())
	}

	if err == nil && createCert {

		certPath := path.Join(harvestConfPath, "cert", client.Name()+".pem")
		keyPath := path.Join(harvestConfPath, "cert", client.Name()+".key")

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
			keyPath,
			"-out",
			certPath,
			"-subj",
			"/CN="+harvestUserName,
		)

		if err := cmd.Run(); err != nil {
			exitError("openssl", err)
		}

		_dialog.Message(fmt.Sprintf("Generated certificate/key pair:\n  - %s\n  - %s\n", certPath, keyPath))

		req := node.NewXmlS("security-login-role-create")
		req.NewChildS("access-level", "readonly")
		req.NewChildS("command-directory-name", "DEFAULT")
		req.NewChildS("role-name", harvestRoleName)
		req.NewChildS("vserver", client.Name())

		if _, err := client.InvokeRequest(req); err != nil {
			exitError("create role", err)
		}

		req = node.NewXmlS("security-login-create")
		req.NewChildS("application", "ontapi")
		req.NewChildS("authentication-method", "cert")
		req.NewChildS("comment", "readonly user for harvest2")
		req.NewChildS("role-name", harvestRoleName)
		req.NewChildS("user-name", harvestUserName)
		req.NewChildS("vserver", client.Name())

		if _, err := client.InvokeRequest(req); err != nil {
			exitError("create user", err)
		}

		_dialog.Message(fmt.Sprintf("Created read-only user [%s] and role [%s]", harvestUserName, harvestRoleName))

		certContent, err := ioutil.ReadFile(certPath)
		if err != nil {
			exitError("cert content", err)
		}

		req = node.NewXmlS("security-certificate-install")
		req.NewChildS("cert-name", harvestUserName)
		req.NewChildS("certificate", string(certContent))
		req.NewChildS("type", "client_ca")
		req.NewChildS("vserver", client.Name())

		if _, err := client.InvokeRequest(req); err != nil {
			exitError("install cert", err)
		}

		_dialog.Message("Certificate installed on system.")

		// forget password immediately
		poller.PopChildS("auth_style")
		poller.PopChildS("username")
		poller.PopChildS("password")

		// new auth parameters
		poller.NewChildS("auth_style", "certificate_auth")
		poller.NewChildS("ssl_cert", certPath)
		poller.NewChildS("ssl_key", keyPath)
	}

	collectors := poller.NewChildS("collectors", "")
	collectors.NewChildS("", "Zapi")
	collectors.NewChildS("", "ZapiPerf")

	return poller
}

func addExporter() *node.Node {

	exporter := node.NewS("")

	item, err := _dialog.Menu("Choose exporter type:", "prometheus", "influxdb")
	if err != nil {
		return nil
	}
	exporter.NewChildS("exporter", item)

	name, err := _dialog.Input("Choose name for exporter instance:")
	if err != nil {
		return nil
	}
	exporter.SetNameS(name)

	port, err := _dialog.Input("Port of the HTTP service:")
	if err != nil {
		exitError("input exporter port", err)
	}
	exporter.NewChildS("port", port)

	if _dialog.YesNo("Make HTTP serve publicly on your network?\n(Choose no to serve it only on localhst)") {
		exporter.NewChildS("addr", "0.0.0.0")
	} else {
		exporter.NewChildS("addr", "localhost")
	}

	exporter.NewChildS("master", "True")

	return exporter
}
