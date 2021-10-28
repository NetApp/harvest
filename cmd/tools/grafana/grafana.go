/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package grafana

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/util"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	clientTimeout           = 5
	harvestRelease          = "21.11"
	grafanaFolderTitle      = "Harvest " + harvestRelease + " - cDOT"
	grafana7modeFolderTitle = "Harvest " + harvestRelease + " - 7-mode"
	grafanaDataSource       = "Prometheus"
)

var (
	grafanaMinVers = "7.1.0" // lowest grafana version we require
	homePath       string
)

type options struct {
	command        string // one of: import, export, clean
	addr           string // URL of Grafana server (e.g. "http://localhost:3000")
	token          string // API token issued by Grafana server
	dir            string // Local directory for import/export-ing cDot & 7mode dashboards (e.g. "opt/harvest/grafana/dashboards")
	dircDOT        string // Local directory for import/export-ing cDOT dashboards (e.g. "opt/harvest/grafana/dashboards/cmode")
	dir7mode       string // Local directory for import/export-ing 7mode dashboards (e.g. "opt/harvest/grafana/dashboards/7mode")
	cmodeFolder    Folder // Server-side Grafana folder name for cDOT dashboards
	mode7Folder    Folder // Server-side Grafana folder name for 7-mode dashboards
	datasource     string
	variable       bool
	client         *http.Client
	headers        http.Header
	config         string
	prefix         string
	useHttps       bool
	useInsecureTLS bool
	labels         []string
}

type Folder struct {
	folderName string // Grafana folder where to upload from where to download dashboards
	folderId   int64
	folderUid  string
}

func doExport(_ *cobra.Command, _ []string) {
	adjustOptions()
	askForToken()
	var doesFolderExist = doesGrafanaFolderExist()
	var err error

	if !(doesFolderExist[&opts.cmodeFolder] && doesFolderExist[&opts.mode7Folder]) {
		if !doesFolderExist[&opts.cmodeFolder] {
			fmt.Printf("folder [%s] not found in Grafana\n", opts.cmodeFolder.folderName)
		}
		if !doesFolderExist[&opts.mode7Folder] {
			fmt.Printf("folder [%s] not found in Grafana\n", opts.mode7Folder.folderName)
		}
		os.Exit(1)
	} else if err = exportDashboards(opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addLabel(content []byte, label string, labelMap map[string]string) []byte {
	// extract the list of variables
	templateList := gjson.GetBytes(content, "templating.list")
	if !templateList.Exists() {
		fmt.Printf("No template variables found, ignoring add label")
		return content
	}
	vars := templateList.Array()

	// create a new list of vars and copy the existing ones into it, duplicate the first var since we're going to
	// overwrite it
	newVars := make([]gjson.Result, 0)
	newVars = append(newVars, vars[0])
	for _, result := range vars {
		newVars = append(newVars, result)
	}

	// Assume Datasource is first and insert the new label var as the second element.
	// If we're wrong, that's OK, no harm
	newVars[1] = gjson.ParseBytes(newLabelVar(label))

	// Write the variables into a string builder
	// Modify the existing variables by adding the new label query
	varsString := strings.Builder{}
	varsString.WriteString("[")
	for i, result := range newVars {
		aStr := addChainedVar(result, label, labelMap)
		if aStr == "" {
			varsString.WriteString(result.String())
		} else {
			varsString.WriteString(aStr)
		}
		if i < len(newVars)-1 {
			varsString.WriteString(",\n")
		}
	}
	varsString.WriteString("]")

	newContent, err := sjson.SetRawBytes(content, "templating.list", []byte(varsString.String()))
	if err != nil {
		fmt.Printf("error inserting label=[%s] into dashboard, err: %+v", label, err)
		return content
	}
	return newContent
}

func addChainedVar(result gjson.Result, label string, labelMap map[string]string) string {
	varName := result.Get("name")
	definition := result.Get("definition")
	query := result.Get("query.query")

	if !varName.Exists() || !definition.Exists() || !query.Exists() {
		return ""
	}
	// Don't modify the query if this is one of the new labels we're adding since its query is already correct
	if _, ok := labelMap[varName.String()]; ok {
		return ""
	}

	defStr := definition.String()
	if defStr != query.String() {
		return ""
	}
	chained := toChainedVar(defStr, label)
	if chained == "" {
		return ""
	}
	rString := result.String()
	var err error
	rString, err = sjson.Set(rString, "definition", chained)
	if err != nil {
		fmt.Printf("error setting definition of varName=[%s] for label=[%s], err: %+v", varName, label, err)
		return ""
	}
	rString, err = sjson.Set(rString, "query.query", chained)
	if err != nil {
		fmt.Printf("error setting query of varName=[%s] for label=[%s], err: %+v", varName, label, err)
		return ""
	}
	return rString
}

func toChainedVar(defStr string, label string) string {
	if !strings.Contains(defStr, "label_values") {
		return ""
	}

	title := strings.Title(label)
	lastBracket := strings.LastIndex(defStr, "}")
	if lastBracket == -1 {
		lastParen := strings.LastIndex(defStr, ")")
		if lastParen == -1 {
			return ""
		}

		lastComma := strings.LastIndex(defStr, ",")
		firstParen := strings.Index(defStr, "(")
		if firstParen == -1 {
			return ""
		}
		if lastComma == -1 {
			// Case 1: There are not existing labels
			// label_values(datacenter) becomes label_values({foo=~"$Foo"}, datacenter)
		} else {
			// Case 2: There is a single metric
			// label_values(poller_status, datacenter) becomes label_values(poller_status{org=~"$org"}, datacenter)
			return defStr[0:lastComma] + "{" + label + `=~"$` + title + `"},` + defStr[lastComma+1:]
		}
		if firstParen+1 > len(defStr) {
			return ""
		}
		return defStr[0:firstParen] + "({" + label + `=~"$` + title + `"}, ` + defStr[firstParen+1:]
	}
	// Case 2: There are existing metrics
	// label_values(aggr_new_status{datacenter="$Datacenter",cluster="$Cluster"}, node) becomes
	// label_values(aggr_new_status{datacenter="$Datacenter",cluster="$Cluster",foo=~"$Foo"}, node)
	return defStr[0:lastBracket] + "," + label + `=~"$` + title + `"` + defStr[lastBracket:]
}

func newLabelVar(label string) []byte {
	return []byte(fmt.Sprintf(`{
  "allValue": ".*",
  "current": {
    "selected": false
  },
  "definition": "label_values(%s)",
  "hide": 0,
  "includeAll": true,
  "multi": true,
  "name": "%s",
  "options": [],
  "query": {
    "query": "label_values(%s)",
    "refId": "StandardVariableQuery"
  },
  "refresh": 1,
  "regex": "",
  "skipUrlSync": false,
  "sort": 0,
  "type": "query"
}`, label, strings.Title(label), label))
}

func doImport(_ *cobra.Command, _ []string) {
	opts.command = "import"
	err := conf.LoadHarvestConfig(opts.config)
	if err != nil {
		return
	}
	adjustOptions()
	askForToken()
	checkAndCreateFolder()

	fmt.Printf("preparing to import dashboards...\n")
	if err := importDashboards(opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func checkAndCreateFolder() {
	var folderExists = doesGrafanaFolderExist()
	var err error

	// first folder is Harvest 2.0 and second folder is 7mode
	for folder, value := range folderExists {
		if value {
			fmt.Printf("folder [%s] exists in Grafana - OK\n", folder.folderName)
		} else if err = createFolder(opts, folder); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Printf("created Grafana folder [%s] - OK\n", folder.folderName)
		}
	}

}

func doesGrafanaFolderExist() map[*Folder]bool {
	folderExists := make(map[*Folder]bool)
	var err error
	if folderExists, err = checkFolder(opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return folderExists
}

func askForToken() {
	// ask for API token if not provided as arg and validate
	if err := checkToken(opts, false); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func adjustOptions() {
	homePath = conf.GetHarvestHomePath()
	if opts.dircDOT != "" {
		opts.dircDOT = path.Join(homePath, opts.dircDOT)
		exitIfMissing(opts.dircDOT, "directory-cdot")
	}
	if opts.dir7mode != "" {
		opts.dir7mode = path.Join(homePath, opts.dir7mode)
		exitIfMissing(opts.dir7mode, "directory-7mode")
	}
	if opts.dircDOT == "" && opts.dir7mode == "" {
		opts.dir = path.Join(homePath, opts.dir)
		exitIfMissing(opts.dir, "directory")
		opts.dircDOT = path.Join(opts.dir, "/cmode")
		opts.dir7mode = path.Join(opts.dir, "/7mode")
	}

	// When opt.addr starts with https don't change it
	if !strings.HasPrefix(opts.addr, "https://") {
		opts.addr = strings.TrimPrefix(opts.addr, "http://")
		opts.addr = strings.TrimPrefix(opts.addr, "https://")
		opts.addr = strings.TrimSuffix(opts.addr, "/")
		if opts.useHttps {
			opts.addr = "https://" + opts.addr
		} else {
			opts.addr = "http://" + opts.addr
		}
	}
}

func exitIfMissing(fp string, s string) {
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		fmt.Printf("error: %s file [%s] does not exist.\n", s, fp)
		os.Exit(1)
	}
}

func exportDashboards(opts *options) error {
	// Exporting C mode dashboards
	_ = exportFiles(opts.cmodeFolder)
	// Exporting 7mode dashboards
	_ = exportFiles(opts.mode7Folder)
	return nil
}

func exportFiles(folder Folder) error {
	var (
		err   error
		uids  map[string]string
		dir   string
		count int
	)

	fmt.Printf("querying for content of folder id [%d]\n", folder.folderId)

	result, status, code, err := sendRequestArray(opts, "GET", "/api/search?folderIds="+strconv.FormatInt(folder.folderId, 10), nil)
	if err != nil && code != 200 {
		fmt.Printf("server response [%d: %s]: %v\n", code, status, err)
		return err
	}

	uids = make(map[string]string)
	for _, elem := range result {
		uid := elem["uid"].(string)
		uri := elem["uri"].(string)
		if uid != "" && uri != "" {
			uids[uid] = strings.ReplaceAll(strings.ReplaceAll(uri, "/", "_"), "-", "_")
		}
	}

	if opts.dir == "" {
		dir = path.Join("./", strings.ReplaceAll(folder.folderName, " ", "_"))
	} else {
		dir = path.Join(opts.dir, strings.ReplaceAll(folder.folderName, " ", "_"))
	}
	if err = os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("error makedir [%s]: %v\n", dir, err)
		return err
	}
	fmt.Printf("exporting dashboards to directory [%s]\n", dir)

	for uid, uri := range uids {
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
	importFiles(opts.dircDOT, opts.cmodeFolder)
	importFiles(opts.dir7mode, opts.mode7Folder)
	return nil
}

func importFiles(dir string, folder Folder) {
	var (
		request, dashboard map[string]interface{}
		files              []os.FileInfo
		importedFiles      int
		data               []byte
		err                error
	)
	if dir == "" {
		return
	}
	if files, err = ioutil.ReadDir(dir); err != nil {
		// TODO check for not exist
		return
	}

	releaseVersion := strings.ReplaceAll(harvestRelease, ".", "-")
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		if data, err = ioutil.ReadFile(path.Join(dir, file.Name())); err != nil {
			fmt.Printf("error reading file [%s]\n", file.Name())
			return
		}

		data = bytes.ReplaceAll(data, []byte("${DS_PROMETHEUS}"), []byte(opts.datasource))

		// Updating the uid of dashboards based in the release
		uid := gjson.GetBytes(data, "uid").String()
		data, err = sjson.SetBytes(data, "uid", []byte(uid+"-"+releaseVersion))
		if err != nil {
			fmt.Printf("error while updating the uid %s into dashboard %s, err: %+v", uid, file.Name(), err)
			continue
		}

		// labelMap is used to ensure we don't modify the query of one of the new labels we're adding
		labelMap := make(map[string]string)
		for _, label := range opts.labels {
			labelMap[strings.Title(label)] = label
		}
		// The label is inserted in the list of variables first
		// Iterate backwards so the labels keep the same order as cmdline
		for i := len(opts.labels) - 1; i >= 0; i-- {
			data = addLabel(data, opts.labels[i], labelMap)
		}

		if err = json.Unmarshal(data, &dashboard); err != nil {
			fmt.Printf("error parsing file [%s] %+v\n", file.Name(), err)
			fmt.Println("-------------------------------")
			fmt.Println(string(data))
			fmt.Println("-------------------------------")
			return
		}

		// optionally add prefix to all metric names in the queries
		if opts.prefix != "" {
			addGlobalPrefix(dashboard, opts.prefix)
		}

		request = make(map[string]interface{})
		request["overwrite"] = true
		request["folderId"] = folder.folderId
		request["dashboard"] = dashboard

		result, status, code, err := sendRequest(opts, "POST", "/api/dashboards/db", request)

		if err != nil {
			fmt.Printf("error importing [%s]\n", file.Name())
			return
		}

		if code != 200 {
			fmt.Printf("error importing [%s] - server response (%d - %s) %v\n", file.Name(), code, status, result)
			return
		}
		fmt.Printf("OK - imported %s / [%s]\n", folder.folderName, file.Name())
		importedFiles++
	}
	if importedFiles > 0 {
		fmt.Printf("Imported %d dashboards to [%s] from [%s]\n", importedFiles, folder.folderName, dir)
	} else {
		fmt.Printf("No dashboards found in [%s] is the directory correct?\n", dir)
	}
}

// addGlobalPrefix adds the given prefix to all metric names in the
// dashboards. It assumes that metrics are in Prometheus-format.
//
// A more reliable implementation of this feature would be, to
// add a constant prefix to all metrics, before they are pushed
// to Github, then replace them with a user-defined prefix
// (or empty string) when the import tool is used.
func addGlobalPrefix(db map[string]interface{}, prefix string) {

	var (
		panels, targets, templates                 []interface{}
		panel, target, templating, template, query map[string]interface{}
		p, t                                       interface{}
		queryString, definition, expr              string
		ok, has                                    bool
	)

	// make sure prefix ends with _
	if !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}

	// apply to queries in panels
	if panels, ok = db["panels"].([]interface{}); !ok {
		return
	}

	for _, p = range panels {
		if panel, ok = p.(map[string]interface{}); !ok {
			continue
		}

		if _, has = panel["targets"]; !has {
			continue
		}

		if targets, ok = panel["targets"].([]interface{}); !ok {
			continue
		}

		for _, t = range targets {

			if target, ok = t.(map[string]interface{}); !ok {
				continue
			}

			if _, has = target["expr"]; has {
				if expr, ok = target["expr"].(string); ok {
					target["expr"] = addPrefixToMetricNames(expr, prefix)
				}
			}
		}
	}

	// apply to variables
	if templating, ok = db["templating"].(map[string]interface{}); !ok {
		return
	}

	if templates, ok = templating["list"].([]interface{}); !ok {
		return
	}

	for _, t = range templates {
		if template, ok = t.(map[string]interface{}); ok {
			if definition, ok = template["definition"].(string); ok {
				template["definition"] = addPrefixToMetricNames(definition, prefix)
			}
			if query, ok = template["query"].(map[string]interface{}); ok {
				if queryString, ok = query["query"].(string); ok {
					query["query"] = addPrefixToMetricNames(queryString, prefix)
				}
			}
		}
	}
}

// addPrefixToMetricNames adds prefix to metric names in expr or leaves it
// unchanged if no metric names are identified.
// Note that this function will only work with the Prometheus-dashboards of Harvest.
// It will use a number of patterns in which metrics might be used in queries.
// (E.g. a single metric, multiple metrics used in addition, etc -- for examples
// see the test). If we change queries of our dashboards, we have to review
// this function as well (or come up with a better solution).
func addPrefixToMetricNames(expr, prefix string) string {
	var (
		match    [][]string
		submatch []string
		isMatch  bool
		regex    *regexp.Regexp
		err      error
	)

	// variable queries
	if strings.HasPrefix(expr, "label_values(") {
		if isMatch, err = regexp.MatchString(`^label_values\s?\(([a-zA-Z_])+(\s?{.+?})?,\s?[a-zA-Z_]+\)$`, expr); err != nil {
			fmt.Printf("Regex error: %v\n", err)
			return expr
		} else if isMatch {
			return strings.Replace(expr, "label_values(", "label_values("+prefix, 1)
		} else {
			// no metric name
			return expr
		}
	}

	// everything else is for graph queries
	regex = regexp.MustCompile(`([a-zA-Z_+]+)\s?{.+?}`)
	match = regex.FindAllStringSubmatch(expr, -1)

	for _, m := range match {
		// multiple metrics used to summarize
		if strings.Contains(m[1], "+") {
			submatch = strings.Split(m[1], "+")
			for i := range submatch {
				submatch[i] = prefix + submatch[i]
			}
			expr = strings.Replace(expr, m[1], strings.Join(submatch, "+"), 1)
			// single metric
		} else {
			expr = strings.Replace(expr, m[1], prefix+m[1], 1)
		}
	}

	return expr
}

func checkToken(opts *options, ignoreConfig bool) error {

	// @TODO check and handle expired API token

	var (
		token, configPath, answer string
		err                       error
	)

	configPath = opts.config

	err = conf.LoadHarvestConfig(configPath)
	if err != nil {
		return err
	}

	if conf.Config.Tools != nil {
		if !ignoreConfig {
			token = conf.Config.Tools.GrafanaApiToken
			fmt.Println("using API token from config")
		}
	}

	if opts.token == "" && token == "" {
		fmt.Printf("enter API token: ")
		_, _ = fmt.Scanf("%s\n", &opts.token)
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
		tlsConfig := &tls.Config{InsecureSkipVerify: opts.useInsecureTLS}
		opts.client.Transport = &http.Transport{TLSClientConfig: tlsConfig}
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

	// ask user to save API key
	if conf.Config.Tools == nil || opts.token != conf.Config.Tools.GrafanaApiToken {

		fmt.Printf("save API key for later use? [Y/n]: ")
		_, _ = fmt.Scanf("%s\n", &answer)

		if answer == "Y" || answer == "y" || answer == "yes" || answer == "" {
			if conf.Config.Tools == nil {
				conf.Config.Tools = &conf.Tools{}
			}
			conf.Config.Tools.GrafanaApiToken = opts.token
			fmt.Printf("saving config file [%s]\n", configPath)
			if err = util.SaveConfig(configPath, opts.token); err != nil {
				return err
			}
		}
	}

	// get grafana version, we are more or less guaranteed this succeeds
	if result, status, code, err = sendRequest(opts, "GET", "/api/health", nil); err != nil {
		return err
	}

	grafanaVersion := result["version"].(string)
	fmt.Printf("connected to Grafana server (version: %s)\n", grafanaVersion)
	// if we are going to import check grafana version
	if opts.command == "import" && !checkVersion(grafanaVersion) {
		fmt.Printf("warning: current set of dashboards require Grafana version (%s) or higher\n", grafanaMinVers)
		fmt.Printf("continue anyway? [y/N]: ")
		_, _ = fmt.Scanf("%s\n", &answer)
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

func createFolder(opts *options, folder *Folder) error {

	var request map[string]interface{}
	request = make(map[string]interface{})

	request["title"] = folder.folderName

	result, status, code, err := sendRequest(opts, "POST", "/api/folders", request)

	if err != nil {
		return err
	}

	if code != 200 {
		return errors.New("server response: " + status)
	}

	folder.folderId = int64(result["id"].(float64))
	folder.folderUid = result["uid"].(string)

	return nil
}

func checkFolder(opts *options) (map[*Folder]bool, error) {

	result, status, code, err := sendRequestArray(opts, "GET", "/api/folders?limit=1000", nil)

	if err != nil {
		return map[*Folder]bool{&opts.cmodeFolder: false, &opts.mode7Folder: false}, err
	}

	if code != 200 {
		return map[*Folder]bool{&opts.cmodeFolder: false, &opts.mode7Folder: false}, errors.New("server response: " + status)
	}

	if len(result) == 0 {
		return map[*Folder]bool{&opts.cmodeFolder: false, &opts.mode7Folder: false}, nil
	}

	cmodeFolder := folderExist(result, &opts.cmodeFolder)
	nonCmodeFolder := folderExist(result, &opts.mode7Folder)

	return map[*Folder]bool{&opts.cmodeFolder: cmodeFolder, &opts.mode7Folder: nonCmodeFolder}, nil
}

func folderExist(result []map[string]interface{}, folder *Folder) bool {
	for _, x := range result {

		if name, ok := x["title"]; ok {
			if name.(string) == folder.folderName {
				if id, idExist := x["id"]; idExist {
					folder.folderId = int64(id.(float64))
					if uid, uidExist := x["uid"]; uidExist {
						folder.folderUid = uid.(string)
						return true
					}
				}
			}
		}

	}
	return false
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

	defer silentClose(response.Body)
	data, err = ioutil.ReadAll(response.Body)
	return data, status, code, err
}

func silentClose(body io.ReadCloser) {
	_ = body.Close()
}

var opts = &options{}

var Cmd = &cobra.Command{
	Use:   "grafana",
	Short: "Import/export Grafana dashboards",
	Long:  "Grafana tool - Import/Export Grafana dashboards",
}

var importCmd = &cobra.Command{
	Use:     "import",
	Short:   "import Grafana dashboards",
	Run:     doImport,
	Example: "grafana import --addr my.grafana.server:3000",
}

var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "export Grafana dashboards",
	Run:     doExport,
	Example: "grafana export --addr my.grafana.server:3000 --directory exported_dash",
}

func init() {
	Cmd.AddCommand(importCmd, exportCmd)

	Cmd.PersistentFlags().StringVar(&opts.config, "config", "./harvest.yml", "harvest config file path")
	Cmd.PersistentFlags().StringVarP(&opts.addr, "addr", "a", "http://127.0.0.1:3000", "Address of Grafana server (IP, FQDN or hostname)")
	Cmd.PersistentFlags().StringVarP(&opts.token, "token", "t", "", "API token issued by Grafana server for authentication")
	Cmd.PersistentFlags().StringVarP(&opts.dir, "directory", "d", "grafana/dashboards", "When importing, import cDOT and 7mode dashboards from this local directory.\nWhen exporting, local directory to write dashboards to")
	Cmd.PersistentFlags().StringVar(&opts.dircDOT, "directory-cdot", "", "When importing, import cDOT dashboards from this local directory.\nWhen exporting, local directory to write dashboards to")
	Cmd.PersistentFlags().StringVar(&opts.dir7mode, "directory-7mode", "", "When importing, import 7-mode dashboards from this local directory.\nWhen exporting, local directory to write dashboards to")
	Cmd.PersistentFlags().StringVarP(&opts.cmodeFolder.folderName, "folder", "f", grafanaFolderTitle, "Grafana folder name for the cDOT dashboards")
	Cmd.PersistentFlags().StringVarP(&opts.mode7Folder.folderName, "folder-7mode", "", grafana7modeFolderTitle, "Grafana folder name for the 7-mode dashboards")
	Cmd.PersistentFlags().StringVarP(&opts.prefix, "prefix", "p", "", "Use global metric prefix in queries")
	Cmd.PersistentFlags().StringVarP(&opts.datasource, "datasource", "s", grafanaDataSource, "Grafana datasource for the dashboards")
	Cmd.PersistentFlags().BoolVarP(&opts.variable, "variable", "v", false, "Use datasource as variable, overrides: --datasource")
	Cmd.PersistentFlags().BoolVarP(&opts.useHttps, "https", "S", false, "Use HTTPS")
	Cmd.PersistentFlags().BoolVarP(&opts.useInsecureTLS, "insecure", "k", false, "Allow insecure server connections when using SSL")

	importCmd.PersistentFlags().StringSliceVar(&opts.labels, "labels", nil,
		"For each label, create a variable and add as chained query to other variables")
}
