/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-version"
	"goharvest2/pkg/argparse"
	"goharvest2/pkg/config"
	"goharvest2/pkg/tree/node"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	clientTimeout      = 5
	grafanaFolderTitle = "Harvest 2.0"
	grafanaFolderUid   = "harvest2.0folder"
	grafanaDataSource  = "Prometheus"
)

var (
	grafanaMinVers = "7.1.0" // lowest grafana version we require
	confPath       string
)

type options struct {
	command    string // one of: import, export, clean
	addr       string // URL of Grafana server (e.g. "http://localhost:3000")
	token      string // API token issued by Grafana server
	dir        string // Directory from which to import dashboards (e.g. "opt/harvest/grafana/prometheus")
	folder     string // Grafana folder where to upload from where to download dashboards
	folderId   int64
	folderUid  string
	datasource string
	variable   bool
	client     *http.Client
	headers    http.Header
}

func main() {

	var (
		opts   *options
		err    error
		exists bool
	)

	// set harvest config path
	confPath = config.GetHarvestConf()

	// parse CLI args
	opts = getOptions()

	if opts.command == "" {
		fmt.Println("missing positional argument: command")
		os.Exit(1)
	}

	// assume command is "import"
	// other commands not implemented yet

	// ask for API token if not provided as arg and validate
	if err = checkToken(opts, false); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// check if Grafana folder exists
	if exists, err = checkFolder(opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if opts.command == "import" {
		if exists {
			fmt.Printf("folder [%s] exists in Grafana - OK\n", opts.folder)
		} else if err = createFolder(opts); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Printf("created Grafana folder [%s] - OK\n", opts.folder)
		}
		if err = importDashboards(opts); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if opts.command == "export" {
		if !exists {
			fmt.Printf("folder [%s] not found in Grafana\n", opts.folder)
			os.Exit(1)
		} else if err = exportDashboards(opts); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func exportDashboards(opts *options) error {
	var (
		//request *node.Node
		err   error
		uids  map[string]string
		dir   string
		count int
	)

	fmt.Printf("querying for content of folder id [%d]\n", opts.folderId)
	/*
	   request = node.NewS("")
	   request.NewChildS("type", "dash-db")
	   fd := request.NewChildS("folderIds", "")
	   fd.NewChildS("", opts.folderId)
	*/

	//result, status, code, err := sendRequest(opts, "POST", "/api/search?folderIds=", json.Dump(request))
	result, status, code, err := sendRequestArray(opts, "GET", "/api/search?folderIds="+strconv.FormatInt(opts.folderId, 10), nil)
	if err != nil && code != 200 {
		fmt.Printf("server response [%d: %s]: %v\n", code, status, err)
		return err
	}
	//result.Print(0)

	uids = make(map[string]string)
	for _, elem := range result {
		uid := elem["uid"].(string)
		uri := elem["uri"].(string)
		if uid != "" && uri != "" {
			uids[uid] = strings.ReplaceAll(strings.ReplaceAll(uri, "/", "_"), "-", "_")
		}
	}

	if opts.dir == "" {
		dir = path.Join("./", strings.ReplaceAll(opts.folder, " ", "_"))
	} else {
		dir = path.Join(opts.dir, strings.ReplaceAll(opts.folder, " ", "_"))
	}
	if err = os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("error makedir [%s]: %v\n", dir, err)
		return err
	}
	fmt.Printf("exporting dashboards to directory [%s]\n", dir)
	//fmt.Printf("fetching %d dashboards from folder [%s]...\n", len(uids), opts.folder)

	for uid, uri := range uids {
		//fmt.Printf("(debug) [%s] => [%s]\n", uid, uri)
		if result, status, code, err := sendRequest(opts, "GET", "/api/dashboards/uid/"+uid, nil); err != nil {
			fmt.Printf("error requesting [%s]: [%d: %s] %v\n", uid, code, status, err)
			return err
		} else if dashboard, ok := result["dashboard"]; ok {
			fp := path.Join(dir, uri+".json")
			if data, err := json.Marshal(dashboard); err != nil {
				fmt.Printf("error marshall dashboard [%s]: %v\n\n", uid, err)
				return err
			} else if err = ioutil.WriteFile(fp, data, 0644); err != nil {
				fmt.Printf("error write to [%s]: %v\n", fp, err)
				return err
			} else {
				fmt.Printf("OK - exported [%s]\n", fp)
				count++
			}
		}
	}
	fmt.Printf("exported %d dashboards to [%s]\n", count, dir)
	return nil
}

func importDashboards(opts *options) error {

	var (
		files              []os.FileInfo
		request, dashboard map[string]interface{}
		data               []byte
		err                error
	)

	if files, err = ioutil.ReadDir(opts.dir); err != nil {
		// TODO check for not exist
		return err
	}

	fmt.Printf("preparing to import %d dashboards...\n", len(files))

	for _, f := range files {

		if !strings.HasSuffix(f.Name(), ".json") {
			//fmt.Printf("Skipping [%s]...\n", f.Name())
			continue
		}

		//fmt.Printf("Importing [%s] ", f.Name())

		if data, err = ioutil.ReadFile(path.Join(opts.dir, f.Name())); err != nil {
			fmt.Printf("error reading file [%s]\n", f.Name())
			return err
		}

		data = bytes.ReplaceAll(data, []byte("${DS_PROMETHEUS}"), []byte(opts.datasource))

		if err = json.Unmarshal(data, &dashboard); err != nil {
			fmt.Printf("error parsing file [%s]\n", f.Name())
			fmt.Println("-------------------------------")
			fmt.Println(string(data))
			fmt.Println("-------------------------------")
			return err
		}

		request = make(map[string]interface{})
		request["overwrite"] = true
		request["folderId"] = opts.folderId
		request["dashboard"] = dashboard

		result, status, code, err := sendRequest(opts, "POST", "/api/dashboards/db", request)

		if err != nil {
			fmt.Printf("error importing [%s]\n", f.Name())
			return err
		}

		if code != 200 {
			fmt.Printf("error - server response (%d - %s) %v\n", code, status, result)
			return errors.New(status)
		}
		fmt.Printf("OK - imported [%s]\n", f.Name())
	}
	return nil
}

func getOptions() *options {

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
		[]string{"import", "export"},
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

	opts.dir = "prometheus"
	parser.String(
		&opts.dir,
		"directory",
		"d",
		"Directory from which to import or where to export dashboards (default: prometheus)",
	)

	opts.folder = grafanaFolderTitle
	parser.String(
		&opts.folder,
		"folder",
		"f",
		"Grafana folder name for the dashboards (default: \""+grafanaFolderTitle+"\")",
	)

	opts.datasource = grafanaDataSource
	parser.String(
		&opts.datasource,
		"datasource",
		"s",
		"Grafana datasource for the dashboards (default: \""+grafanaDataSource+"\")",
	)

	parser.Bool(
		&opts.variable,
		"variable",
		"v",
		"Use datasource as variable (overrides: --datasource, default: false)",
	)

	parser.Bool(
		&use_https,
		"https",
		"S",
		"Force to use HTTPS (default: false)",
	)

	parser.SetHelpFlag("help")

	parser.ParseOrExit()

	// full path
	if opts.command == "import" {
		opts.dir = path.Join(confPath, "grafana", opts.dir)
	}

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

func checkToken(opts *options, ignoreConfig bool) error {

	// @TODO check and handle expired API token

	var (
		params, tools              *node.Node
		token, config_path, answer string
		err                        error
	)

	config_path = path.Join(confPath, "harvest.yml")

	if params, err = config.LoadConfig(config_path); err != nil {
		return err
	} else if params == nil {
		return errors.New(fmt.Sprintf("config [%s] not found", config_path))
	}

	if tools = params.GetChildS("Tools"); tools != nil {
		if !ignoreConfig {
			token = tools.GetChildContentS("grafana_api_token")
			fmt.Println("using API token from config")
		}
	}

	if opts.token == "" && token == "" {
		fmt.Printf("enter API token: ")
		fmt.Scanf("%s\n", &opts.token)
	} else if opts.token == "" {
		opts.token = token
	}

	// build headers for HTTP request
	opts.headers = http.Header{}
	opts.headers.Add("Accept", "application/json")
	opts.headers.Add("Content-Type", "application/json")
	opts.headers.Add("Authorization", "Bearer "+opts.token)

	opts.client = &http.Client{Timeout: time.Duration(clientTimeout) * time.Second}
	if strings.HasPrefix(opts.addr, "https://") {
		opts.client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	// send random request to validate token
	result, status, code, err := sendRequest(opts, "GET", "/api/folders/aaaaaaa", nil)
	if err != nil {
		return err
	} else if code != 200 && code != 404 {
		msg := result["message"].(string)
		fmt.Printf("error connect: (%d - %s) %s\n", code, status, msg)
		opts.token = ""
		return checkToken(opts, true)
	}

	// ask user to safe API key
	if opts.token != tools.GetChildContentS("grafana_api_token") {

		fmt.Printf("safe API key for later use? [Y/n]: ")
		fmt.Scanf("%s\n", &answer)

		if answer == "y" || answer == "yes" || answer == "" {
			if tools == nil {
				tools = params.NewChildS("Tools", "")
			}
			tools.SetChildContentS("grafana_api_token", opts.token)
			fmt.Printf("saving config file [%s]\n", config_path)
			if err = config.SafeConfig(params, config_path); err != nil {
				return err
			}
		}
	}

	// get grafana version, we are more or less guaranteed this succeeds
	if result, status, code, err = sendRequest(opts, "GET", "/api/health", nil); err != nil {
		return err
	}

	version := result["version"].(string)
	fmt.Printf("connected to Grafana server (version: %s)\n", version)
	// if we are going to import check grafana version
	if opts.command == "import" && !checkVersion(version) {
		fmt.Printf("warning: current set of dashboards require Grafana version (%s) or higher\n", grafanaMinVers)
		fmt.Printf("continue anyway? [y/N]: ")
		fmt.Scanf("%s\n", &answer)
		if answer != "y" && answer != "yes" {
			os.Exit(0)
		}
	}

	return nil
}

func checkVersion(inputVersion string) bool {
	v1, err := version.NewVersion(inputVersion)
	if err != nil {
		fmt.Println(err)
		return false
	}
	constraints, err := version.NewConstraint(">= " + grafanaMinVers)

	if err != nil {
		fmt.Println(err)
		return false
	}

	// Check if input version is greater than or equal to min version required
	if constraints.Check(v1) {
		return true
	} else {
		fmt.Printf("%s does not satisfies constraints %s", v1, constraints)
		return false
	}
}

func createFolder(opts *options) error {

	var request map[string]interface{}

	request = make(map[string]interface{})

	request["title"] = opts.folder
	//fmt.Println("REQUEST:") // DEBUG
	//request.Print(0)

	result, status, code, err := sendRequest(opts, "POST", "/api/folders", request)

	if err != nil {
		return err
	}

	if code != 200 {
		return errors.New("server response: " + status)
	}

	opts.folderId = int64(result["id"].(float64))
	//opts.folderId = strconv.FormatFloat(result["id"].(float64), 'f', 0, 32)
	//opts.folderId = result["id"].(string)
	opts.folderUid = result["uid"].(string)

	return nil
}

func checkFolder(opts *options) (bool, error) {

	result, status, code, err := sendRequestArray(opts, "GET", "/api/folders?limit=1000", nil)

	if err != nil {
		return false, err
	}

	if code != 200 {
		return false, errors.New("server response: " + status)
	}

	if result == nil || len(result) == 0 {
		return false, nil
	}

	for _, x := range result {
		//elem := x.(map[string]interface{})
		if x["title"].(string) == opts.folder {
			//opts.folderId = strconv.FormatFloat(x["id"].(float64), 'f', 0, 32)
			opts.folderId = int64(x["id"].(float64))
			opts.folderUid = x["uid"].(string)

			// DEBUG
			//fmt.Println("FOUND FOLDER!")
			//x.Print(0)
			return true, nil
		}
	}

	return false, nil
}

func sendRequest(opts *options, method, url string, query map[string]interface{}) (map[string]interface{}, string, int, error) {

	var result map[string]interface{}

	data, status, code, err := doRequest(opts, method, url, query)
	if err != nil {
		return result, status, code, err
	}

	if err = json.Unmarshal(data, &result); err != nil {
		fmt.Printf("raw response (%d - %s):\n", code, status)
		fmt.Println(string(data))
	}
	return result, status, code, err
}

func sendRequestArray(opts *options, method, url string, query map[string]interface{}) ([]map[string]interface{}, string, int, error) {

	var result []map[string]interface{}

	data, status, code, err := doRequest(opts, method, url, query)
	if err != nil {
		return result, status, code, err
	}

	if err = json.Unmarshal(data, &result); err != nil {
		fmt.Printf("raw response (%d - %s):\n", code, status)
		fmt.Println(string(data))
	}
	return result, status, code, err
}

func doRequest(opts *options, method, url string, query map[string]interface{}) ([]byte, string, int, error) {

	var (
		request  *http.Request
		response *http.Response
		status   string
		code     int
		err      error
		buf      *bytes.Buffer
		data     []byte
	)

	if query != nil {
		if data, err = json.Marshal(query); err != nil {
			return nil, status, code, err
		}
		buf = bytes.NewBuffer(data)
		request, err = http.NewRequest(method, opts.addr+url, buf)
	} else {
		request, err = http.NewRequest(method, opts.addr+url, nil)
	}

	if err != nil {
		return nil, status, code, err
	}

	//fmt.Printf("(debug) send request [%s]\n", request.URL.String())

	request.Header = opts.headers

	if response, err = opts.client.Do(request); err != nil {
		return nil, status, code, err
	}

	status = response.Status
	code = response.StatusCode

	defer response.Body.Close()
	data, err = ioutil.ReadAll(response.Body)
	return data, status, code, err
}
