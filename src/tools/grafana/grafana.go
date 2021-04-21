package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"goharvest2/share/argparse"
	"goharvest2/share/config"
	"goharvest2/share/tree/json"
	"goharvest2/share/tree/node"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	CLIENT_TIMEOUT       = 5
	GRAFANA_FOLDER_TITLE = "Harvest 2.0"
	GRAFANA_FOLDER_UID   = "harvest2.0folder"
	GRAFANA_DATASOURCE   = "Prometheus" // default datasource to use, @TODO: support others
)

var (
	CONF_PATH string
)

type options struct {
	command         string // one of: import, export, clean
	addr            string // URL of Grafana server (e.g. "http://localhost:3000")
	token           string // API token issued by Grafana server
	import_dir      string // Directory from which to import dashboards (e.g. "/etc/harvest/grafana/prometheus")
	grafana_dir     string // Grafana folder where to upload from where to download dashboards
	grafana_dir_id  string
	grafana_dir_uid string
	datasource      string
	client          *http.Client
	headers         http.Header
}

func main() {

	var (
		opts   *options
		err    error
		exists bool
	)

	// set harvest config path
	if CONF_PATH = os.Getenv("HARVEST_CONF"); CONF_PATH == "" {
		CONF_PATH = "/etc/harvest"
	}

	// parse CLI args
	opts = get_opts()

	if opts.command == "" {
		fmt.Println("missing positional argument: command")
		os.Exit(1)
	}

	// assume command is "import"
	// other commands not implemented yet

	// ask for API token if not provided as arg
	if opts.token == "" {
		if opts.token, err = get_token(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// build headers for HTTP request
	opts.headers = http.Header{}
	opts.headers.Add("Accept", "application/json")
	opts.headers.Add("Content-Type", "application/json")
	opts.headers.Add("Authorization", "Bearer "+opts.token)

	opts.client = &http.Client{Timeout: time.Duration(CLIENT_TIMEOUT) * time.Second}
	if strings.HasPrefix(opts.addr, "https://") {
		opts.client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}

	// check if Grafana folder exists
	if exists, err = check_folder(opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else if exists {
		fmt.Printf("Grafana folder [%s] already exists - OK\n", opts.grafana_dir)
	} else if err = create_folder(opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Printf("Created Grafana folder [%s] - OK\n", opts.grafana_dir)
	}

	if err = import_dashboards(opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func import_dashboards(opts *options) error {

	var (
		files   []os.FileInfo
		request *node.Node
		data    []byte
		err     error
	)

	if files, err = ioutil.ReadDir(opts.import_dir); err != nil {
		// TODO check for not exist
		return err
	}

	for _, f := range files {

		if !strings.HasSuffix(f.Name(), ".json") {
			//fmt.Printf("Skipping [%s]...\n", f.Name())
			continue
		}

		fmt.Printf("Importing [%s] ", f.Name())

		if data, err = ioutil.ReadFile(path.Join(opts.import_dir, f.Name())); err != nil {
			fmt.Println("ERR")
			return err
		}

		request = node.NewS("")
		request.NewChildS("overwrite", "true")
		request.NewChildS("folderId", opts.grafana_dir_id)
		request.NewChild([]byte("dashboard"), bytes.ReplaceAll(data, []byte("${DS_PROMETHEUS}"), []byte(opts.datasource)))

		result, status, code, err := send_request(opts, "POST", "/api/dashboards/db", json.Dump(request))

		if err != nil {
			fmt.Println("ERR")
			return err
		}

		if code != 200 {
			fmt.Println("ERR")
			if result != nil {
				result.Print(0)
			}
			return errors.New("server response: " + status)
		}
		fmt.Println("OK")
	}
	return nil
}

func get_opts() *options {

	var (
		opts      *options
		use_https bool
	)

	opts = &options{}

	parser := argparse.New("Grafana tool", "harvest grafana", "Import/Export Grafana dashboards")

	parser.PosString(
		&opts.command,
		"command",
		"command to execute",
		[]string{"import"},
	)

	opts.addr = "http://127.0.0.1:3000"
	parser.String(
		&opts.addr,
		"addr",
		"a",
		"Address of Grafana server (IP, FQDN or hostname) (default: "+opts.addr+")",
	)

	parser.String(
		&opts.token,
		"token",
		"t",
		"API token issued by Grafana server for authentication",
	)

	opts.import_dir = "prometheus"
	parser.String(
		&opts.import_dir,
		"directory",
		"d",
		"Directory from which to import or where to export dashboards (default: prometheus)",
	)

	opts.grafana_dir = GRAFANA_FOLDER_TITLE
	parser.String(
		&opts.grafana_dir,
		"folder",
		"f",
		"Grafana folder name for the dashboards (default: \""+GRAFANA_FOLDER_TITLE+"\")",
	)

	opts.datasource = GRAFANA_DATASOURCE
	parser.String(
		&opts.datasource,
		"datasource",
		"s",
		"Grafana datasource for the dashboards (default: \""+GRAFANA_DATASOURCE+"\")",
	)

	parser.Bool(
		&use_https,
		"https",
		"s",
		"Force to use HTTPS (default: false)",
	)

	parser.SetHelpFlag("help")

	if !parser.Parse() {
		os.Exit(0)
	}

	// full path
	opts.import_dir = path.Join(CONF_PATH, "grafana", opts.import_dir)

	// full URL
	opts.addr = strings.TrimPrefix(opts.addr, "http://")
	opts.addr = strings.TrimPrefix(opts.addr, "https://")
	opts.addr = strings.TrimSuffix(opts.addr, "/")

	if use_https {
		opts.addr = "https://" + opts.addr
	} else {
		opts.addr = "http://" + opts.addr
	}

	return opts
}

func get_token() (string, error) {

	// @TODO check and handle expired API token

	var (
		params, tools              *node.Node
		token, config_path, answer string
		err                        error
	)

	config_path = path.Join(CONF_PATH, "harvest.yml")

	if params, err = config.LoadConfig(config_path); err != nil {
		return token, err
	} else if params == nil {
		return token, errors.New(fmt.Sprintf("config [%s] not found", config_path))
	}

	if tools = params.GetChildS("Tools"); tools != nil {
		token = tools.GetChildContentS("grafana_api_token")
		fmt.Println("Using API token from config")
	}

	if token == "" {
		fmt.Printf("enter API token: ")
		fmt.Scanf("%s\n", &token)

		fmt.Printf("safe for later use? [y/n]: ")
		fmt.Scanf("%s\n", &answer)

		if answer == "y" || answer == "yes" {
			if tools == nil {
				tools = params.NewChildS("Tools", "")
			}
			tools.SetChildContentS("grafana_api_token", token)
			fmt.Printf("saving config file [%s]\n", config_path)
			if err = config.SafeConfig(params, config_path); err != nil {
				return token, err
			}
		}
	}
	return token, nil
}

func create_folder(opts *options) error {

	var request *node.Node

	request = node.NewS("")
	request.NewChildS("title", opts.grafana_dir)
	//fmt.Println("REQUEST:") // DEBUG
	//request.Print(0)

	result, status, code, err := send_request(opts, "POST", "/api/folders", json.Dump(request))

	if err != nil {
		return err
	}

	if code != 200 {
		return errors.New("server response: " + status)
	}

	opts.grafana_dir_id = result.GetChildContentS("id")
	opts.grafana_dir_uid = result.GetChildContentS("uid")

	// DEBUG
	//fmt.Println("FOLDER CREATED!")
	//result.Print(0)

	return nil
}

func check_folder(opts *options) (bool, error) {

	result, status, code, err := send_request(opts, "GET", "/api/folders?limit=1000", nil)

	if err != nil {
		return false, err
	}

	if code != 200 {
		return false, errors.New("server response: " + status)
	}

	for _, x := range result.GetChildren() {
		if x.GetChildContentS("title") == opts.grafana_dir {
			opts.grafana_dir_id = x.GetChildContentS("id")
			opts.grafana_dir_uid = x.GetChildContentS("uid")

			// DEBUG
			//fmt.Println("FOUND FOLDER!")
			//x.Print(0)
			return true, nil
		}
	}

	return false, nil
}

func send_request(opts *options, method, url string, data []byte) (*node.Node, string, int, error) {

	var (
		request  *http.Request
		response *http.Response
		result   *node.Node
		status   string
		code     int
		err      error
	)

	if method == "GET" {
		request, err = http.NewRequest("GET", opts.addr+url, nil)
	} else {
		request, err = http.NewRequest("POST", opts.addr+url, bytes.NewBuffer(data))
	}

	if err != nil {
		return result, status, code, err
	}

	request.Header = opts.headers

	if response, err = opts.client.Do(request); err != nil {
		return result, status, code, err
	}

	status = response.Status
	code = response.StatusCode

	defer response.Body.Close()
	if data, err = ioutil.ReadAll(response.Body); err == nil {
		result, err = json.Load(data)
	}

	// DEBUG
	if err != nil {
		fmt.Println("raw response body:")
		fmt.Println(string(data))
	}
	return result, status, code, err
}
