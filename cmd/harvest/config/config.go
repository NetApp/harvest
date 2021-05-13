/*
Copyright NetApp Inc, 2021 All rights reserved
*/
package config

import (
	"fmt"
	"github.com/spf13/cobra"
	"goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/dialog"
	"goharvest2/pkg/tree"
	"goharvest2/pkg/tree/node"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
)

const (
	// if we create user/role on ONTAP use these names
	harvestUserName = "harvest2-user"
	harvestRoleName = "harvest2-role"
	// PrometheusPortStart use this port if included
	PrometheusPortStart = 7202
)

var (
	harvestConfigPath string
	harvestHomePath   string
	_dialog           *dialog.Dialog
)

func exitError(msg string, err error) {
	_dialog.Close()
	fmt.Printf("Error (%s): %v\n", msg, err)
	os.Exit(1)
}

const (
	pollerUsage = `
    A pollers monitors a single storage system. This utility creates a 
    poller for ONTAP clusters (CDOT or 7Mode). For a custom poller, edit your 
    config.yaml manually.`

	exporterUsage = `
    An exporter forwards data to a database. The same exporter 
    can be used by more than one pollers, i.e. you need to define
    only one exporter for each of your DBs. This utility creates
    exporters for: Prometheus and InfluxDB.`

	welcome  = "welcome"
	exporter = "exporter"
	poller   = "poller"
)

var ConfigCmd = &cobra.Command{
	Use:    "config",
	Short:  "run the config utility",
	Long:   "Harvest 2.0 - Config utility",
	Hidden: true,
}

var exportCmd = &cobra.Command{
	Use:   exporter,
	Short: "create a new exporter " + exporterUsage,
	Long:  exporterUsage,
	Args:  cobra.OnlyValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		Run(exporter)
	},
}

var pollerCmd = &cobra.Command{
	Use:   poller,
	Short: "create a new poller " + pollerUsage,
	Long:  pollerUsage,
	Args:  cobra.OnlyValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		Run(poller)
	},
}

var welcomeCmd = &cobra.Command{
	Use:    welcome,
	Short:  "run welcome helper",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		Run(welcome)
	},
}

func init() {
	ConfigCmd.AddCommand(pollerCmd, exportCmd, welcomeCmd)
}

func Run(item string) {
	var err error
	harvestConfigPath, err = conf.GetDefaultHarvestConfigPath()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	harvestHomePath = conf.GetHarvestHomePath()

	var confNode, pollers, exporters *node.Node

	if _dialog = dialog.New(); !_dialog.Enabled() {
		fmt.Println("This program requires [dialog] or [whiptail].")
		os.Exit(1)
	}

	if item == welcome {

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

	if confNode, err = conf.LoadConfig(harvestConfigPath); err != nil {
		confNode = node.NewS("")
	}

	if pollers = confNode.GetChildS("Pollers"); pollers == nil {
		pollers = confNode.NewChildS("Pollers", "")
	}

	if exporters = confNode.GetChildS("Exporters"); exporters == nil {
		exporters = confNode.NewChildS("Exporters", "")
	}

	for {

		if item == "" {
			item, err = _dialog.Menu("Add new:", poller, exporter, "save and exit")
			if err != nil {
				// error means user clicked on Cancel
				item = "exit"
				break
			}
		}

		if item == poller {
			if newPoller := addPoller(); newPoller != nil {

				if len(exporters.GetChildren()) == 0 {
					question := "You don't have any exporters defined.\n" +
						"Create Prometheus exporter with default " +
						"parameters and add to this poller?"
					if _dialog.YesNo(question) {
						prometheus := exporters.NewChildS("prometheus", "")
						prometheus.NewChildS("exporter", "Prometheus")
						prometheus.NewChildS("addr", "0.0.0.0")
						prometheus.NewChildS("port", strconv.Itoa(PrometheusPortStart))

						pollerExporters := newPoller.NewChildS("exporters", "")
						pollerExporters.NewChildS("", "prometheus")
					}

				} else if len(exporters.GetChildren()) == 1 {
					exporter := exporters.GetChildren()[0]

					question := "Add exporter [" + exporter.GetNameS() + "] to poller?"
					if _dialog.YesNo(question) {

						pollerExporters := newPoller.NewChildS("exporters", "")
						pollerExporters.NewChildS("", exporter.GetNameS())
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

		if item == exporter {
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

		if item == "exit" || item == "save and exit" {
			break
		}

		item = ""
	}

	if item == "save and exit" {

		useTmp := false
		fp := harvestConfigPath

		dir, fn := path.Split(harvestConfigPath)

		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			if os.Mkdir(harvestHomePath, 0644) != nil {
				fp = path.Join("/tmp", fn)
				useTmp = true
			}
		}

		if err = tree.Export(confNode, "yaml", fp); err != nil {
			exitError("export yaml", err)
		}

		msg := "Saved results as [" + fp + "]"
		if useTmp {
			msg = "You don't have write permissions in [" + harvestHomePath + "]!!\n" +
				"Config file saved as [" + fp + "]. Please move it\n" +
				"to [" + harvestHomePath + "] with a privileged user."
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
			msg := fmt.Sprintf("Copy your cert/key pair to [%s/cert/] as [<SYSTEM_NAME>.key] and [<SYSTEM_NAME>.pem] to continue", harvestHomePath)
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

		certPath := path.Join(harvestHomePath, "cert", client.Name()+".pem")
		keyPath := path.Join(harvestHomePath, "cert", client.Name()+".key")

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

	if _dialog.YesNo("Make HTTP serve publicly on your network?\n(Choose no to serve it only on localhost)") {
		exporter.NewChildS("addr", "0.0.0.0")
	} else {
		exporter.NewChildS("addr", "localhost")
	}

	port, err := _dialog.Input("Port of the HTTP service:")
	if err != nil {
		exitError("input exporter port", err)
	}
	exporter.NewChildS("port", port)

	return exporter
}
