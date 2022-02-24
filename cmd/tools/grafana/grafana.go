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
	goversion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"goharvest2/cmd/harvest/version"
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
	clientTimeout     = 5
	grafanaDataSource = "Prometheus"
)

var (
	grafanaMinVers = "7.1.0" // lowest grafana version we require
	homePath       string
	harvestRelease = version.VERSION
)

type options struct {
	command             string // one of: import, export, clean
	addr                string // URL of Grafana server (e.g. "http://localhost:3000")
	token               string // API token issued by Grafana server
	dir                 string
	serverfolder        Folder
	datasource          string
	variable            bool
	client              *http.Client
	headers             http.Header
	config              string
	prefix              string
	useHttps            bool
	useInsecureTLS      bool
	overwrite           bool
	labels              []string
	dirGrafanaFolderMap map[string]*Folder
	addMultiSelect      bool
}

type Folder struct {
	name string // Grafana folder where to upload from where to download dashboards
	id   int64
	uid  string
}

func adjustOptions() {
	opts.config = conf.ConfigPath(opts.config)
	homePath = conf.GetHarvestHomePath()
	opts.dirGrafanaFolderMap = make(map[string]*Folder)

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

func askForToken() {
	// ask for API token if not provided as arg and validate
	if err := checkToken(opts, false); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func doExport(_ *cobra.Command, _ []string) {
	adjustOptions()
	exitIfExist(opts.dir, "directory")
	askForToken()
	initExportVars()
	fmt.Printf("preparing to export dashboards from serverfolder %s directory %s\n", opts.serverfolder.name, opts.dir)
	exportDashboards(opts)
}

func initExportVars() {
	opts.dirGrafanaFolderMap[opts.dir] = &Folder{name: opts.serverfolder.name}
}

func exportDashboards(opts *options) {
	for k, v := range opts.dirGrafanaFolderMap {
		err := exportFiles(k, v)
		if err != nil {
			fmt.Printf("Error during export %v\n", err)
		}
	}
}

func exportFiles(dir string, folder *Folder) error {
	var (
		err   error
		uids  map[string]string
		count int
	)

	err = checkFolder(folder)
	if err != nil {
		fmt.Printf("folder %s error %v\n", folder.name, err)
		os.Exit(1)
	}
	if folder.id == 0 {
		fmt.Printf("error folder %s doesn't exist in grafana, unable to continue\n", folder.name)
		os.Exit(1)
	}
	fmt.Printf("querying for content of folder name [%s]\n", folder.name)

	result, status, code, err := sendRequestArray(opts, "GET", "/api/search?folderIds="+strconv.FormatInt(folder.id, 10), nil)
	if err != nil && code != 200 {
		fmt.Printf("server response [%d: %s]: %v\n", code, status, err)
		return err
	}

	uids = make(map[string]string)
	rep := strings.NewReplacer("/", "_", "-", "_")
	for _, elem := range result {
		uid := elem["uid"].(string)
		uri := elem["uri"].(string)
		if uid != "" && uri != "" {
			uids[uid] = rep.Replace(uri)
		}
	}

	if err = os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("error makedir [%s]: %v\n", dir, err)
		return err
	}
	fmt.Printf("exporting dashboards to directory [%s]\n", dir)

	for uid, uri := range uids {
		result, status, code, err := sendRequest(opts, "GET", "/api/dashboards/uid/"+uid, nil)
		if err != nil {
			fmt.Printf("error requesting [%s]: [%d: %s] %v\n", uid, code, status, err)
			return err
		}
		if dashboard, ok := result["dashboard"]; ok {
			fp := path.Join(dir, uri+".json")
			data, err := json.Marshal(dashboard)
			if err != nil {
				fmt.Printf("error marshall dashboard [%s]: %v\n\n", uid, err)
				return err
			}
			if err = ioutil.WriteFile(fp, data, 0644); err != nil {
				fmt.Printf("error write to [%s]: %v\n", fp, err)
				return err
			}
			fmt.Printf("OK - exported [%s]\n", fp)
			count++
		}
	}
	fmt.Printf("exported %d dashboards to [%s]\n", count, dir)
	return nil
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
	validateImport()
	askForToken()
	initImportVars()

	fmt.Printf("preparing to import dashboards...\n")
	if err := importDashboards(opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func validateImport() {
	// default case
	if opts.dir == "" && opts.serverfolder.name == "" {
		opts.dir = "grafana/dashboards"
		opts.dir = path.Join(homePath, opts.dir)
	}

	exitIfMissing(opts.dir, "directory")
	files, err := ioutil.ReadDir(opts.dir)
	if err != nil {
		fmt.Printf("Error %v while reading dir [%s] is the directory correct?\n", err, opts.dir)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Printf("No dashboards found in [%s] is the directory correct?\n", opts.dir)
		os.Exit(1)
	}
}

func initImportVars() {
	m := make(map[string]*Folder)

	// default behaviour
	if opts.dir == "grafana/dashboards" && opts.serverfolder.name == "" {
		m[path.Join(opts.dir, "/cmode")] = &Folder{name: "Harvest-" + harvestRelease + "-cDOT"}
		m[path.Join(opts.dir, "/7mode")] = &Folder{name: "Harvest-" + harvestRelease + "-7mode"}
	} else if opts.dir != "" && opts.serverfolder.name != "" {
		m[opts.dir] = &Folder{name: opts.serverfolder.name}
	}

	for k, v := range m {
		err := checkAndCreateServerFolder(v)
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
		opts.dirGrafanaFolderMap[k] = v
	}
}

func checkAndCreateServerFolder(folder *Folder) error {
	err := checkFolder(folder)
	if err != nil {
		fmt.Printf("error %v for folder %s\n", err, folder.name)
		os.Exit(1)
	}

	if folder.uid != "" && folder.id != 0 {
		fmt.Printf("folder [%s] exists in Grafana - OK\n", folder.name)
	} else if err = createServerFolder(folder); err != nil {
		return err
	} else {
		fmt.Printf("created Grafana folder [%s] - OK\n", folder.name)
	}
	return nil
}

func exitIfMissing(fp string, s string) {
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		fmt.Printf("error: %s [%s] does not exist.\n", s, fp)
		os.Exit(1)
	}
}

func exitIfExist(fp string, s string) {
	if _, err := os.Stat(fp); err == nil {
		fmt.Printf("error: %s folder [%s] exists. Please specify an empty or non-existant directory\n", s, fp)
		os.Exit(1)
	}
}

func importDashboards(opts *options) error {
	for k, v := range opts.dirGrafanaFolderMap {
		importFiles(k, v)
	}
	return nil
}

func importFiles(dir string, folder *Folder) {
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

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		if data, err = ioutil.ReadFile(path.Join(dir, file.Name())); err != nil {
			fmt.Printf("error reading file [%s]\n", file.Name())
			return
		}

		data = bytes.ReplaceAll(data, []byte("${DS_PROMETHEUS}"), []byte(opts.datasource))

		// If the dashboard has an uid defined, change the uid to empty string. We do comparison for dashboard create/update based on title
		if dashboardId := gjson.GetBytes(data, "uid").String(); dashboardId != "" {
			data, err = sjson.SetBytes(data, "uid", []byte(""))
			if err != nil {
				fmt.Printf("error while updating the uid %s into dashboard %s, err: %+v", dashboardId, file.Name(), err)
				continue
			}
		}

		// If the dashboard has an id defined, change the id to empty string so Grafana treats this as a new dashboard instead of an update to an existing one
		if dashboardId := gjson.GetBytes(data, "id").String(); dashboardId != "" {
			data, err = sjson.SetBytes(data, "id", []byte(""))
			if err != nil {
				fmt.Printf("error while updating the id %s into dashboard %s, err: %+v", dashboardId, file.Name(), err)
				continue
			}
		}

		if opts.addMultiSelect {
			data = addMultiSelect(data)
			if data == nil {
				fmt.Printf("error while adding multi select into dashboard %s, err: %+v", file.Name(), err)
				continue
			}
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
		request["overwrite"] = opts.overwrite
		request["folderId"] = folder.id
		request["dashboard"] = dashboard

		result, status, code, err := sendRequest(opts, "POST", "/api/dashboards/db", request)

		if err != nil {
			fmt.Printf("error importing [%s]  to folder [%s] \n", file.Name(), folder.name)
			return
		}

		if code != 200 {
			fmt.Printf("error importing [%s] to folder [%s] - the server replied with [%s]\n", file.Name(), folder.name, status)
			if code == 412 {
				fmt.Printf("That means you are trying to overwrite an existing dashboard.\n\n")
				fmt.Printf("If you want to overwrite run with the --overwrite flag or choose a different Grafana folder with the --serverfolder flag\n\n")
			}
			fmt.Printf("Server response follows:\n")
			for k, v := range result {
				fmt.Printf("    %s => %s\n", k, v)
			}
			fmt.Println()
			return
		}
		fmt.Printf("OK - imported %s / [%s]\n", folder.name, file.Name())
		importedFiles++
	}
	if importedFiles > 0 {
		fmt.Printf("Imported %d dashboards to [%s] from [%s]\n", importedFiles, folder.name, dir)
	} else {
		fmt.Printf("No dashboards found in [%s] is the directory correct?\n", dir)
	}
}

type change struct {
	path string
	sub  interface{}
}

// addMultiSelect does the following:
// 	- sets "mutli: true" for each query variable
//  - updates the variables definition and query to use a regex equality (=~) instead of an exact match (=)
//  - updates all panels target's expressions to also use a regex equality (=~) instead of an exact match (=)
func addMultiSelect(dashboard []byte) []byte {
	var err error
	changes := make([]change, 0)
	gjson.GetBytes(dashboard, "templating.list").ForEach(func(key, value gjson.Result) bool {
		if value.Get("type").String() != "query" {
			return true
		}
		changes = append(changes, change{
			path: fmt.Sprintf("templating.list.%s.multi", key.String()),
			sub:  true,
		})
		changeEqualsToRegex(&changes, value, key.String())
		return true
	})
	gjson.GetBytes(dashboard, "panels").ForEach(func(key, value gjson.Result) bool {
		value.Get("targets").ForEach(func(key2, value2 gjson.Result) bool {
			if value2.Get("expr").String() != "" {
				changeExprEqualsToRegex(&changes, value2, fmt.Sprintf("panels.%s.targets.%s.expr",
					key.String(), key2.String()))
			}
			return true
		})
		// turtles all the way down, /panels/7/panels/0/targets/0/expr
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			value2.Get("targets").ForEach(func(key3, value3 gjson.Result) bool {
				if value3.Get("expr").String() != "" {
					changeExprEqualsToRegex(&changes, value3, fmt.Sprintf("panels.%s.panels.%s.targets.%s.expr",
						key.String(), key2.String(), key3.String()))
				}
				return true
			})
			return true
		})
		return true
	})
	modifiedDashboard := dashboard
	for _, c := range changes {
		modifiedDashboard, err = sjson.SetBytes(modifiedDashboard, c.path, c.sub)
		if err != nil {
			fmt.Printf("error while updating the query for variable %s, err: %+v", c.path, err)
			return nil
		}
	}
	return modifiedDashboard
}

var equalRegex = regexp.MustCompile(`(\w+)="\$`)

func changeExprEqualsToRegex(changes *[]change, q gjson.Result, key string) {
	query := q.Get("expr")
	if query.String() == "" {
		return
	}
	match := equalRegex.FindStringSubmatch(query.String())
	if len(match) == 0 {
		return
	}
	sub := equalRegex.ReplaceAllString(query.String(), `$1=~"$`)
	*changes = append(*changes, change{
		path: key,
		sub:  sub,
	})
}

func changeEqualsToRegex(changes *[]change, q gjson.Result, key string) {
	jsonPath := "query.query"
	query := q.Get("query.query")
	// check if the query or definition need the regex query added
	// there are two different forms of the query, check for both
	if query.String() == "" {
		query = q.Get("query")
		jsonPath = "query"
	}
	if query.String() == "" {
		return
	}
	match := equalRegex.FindStringSubmatch(query.String())
	if len(match) == 0 {
		return
	}
	sub := equalRegex.ReplaceAllString(query.String(), `$1=~"$`)
	*changes = append(*changes, change{
		path: "templating.list." + key + "." + jsonPath,
		sub:  sub,
	}, change{
		path: "templating.list." + key + ".definition",
		sub:  sub,
	})
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
	v1, err := goversion.NewVersion(inputVersion)
	if err != nil {
		fmt.Println(err)
		return false
	}

	min, _ := goversion.NewVersion(grafanaMinVers)

	// Not using a constraint check since a pre-release version (e.g. 8.4.0-beta1) never matches
	// a constraint specified without a pre-release https://github.com/hashicorp/go-version/pull/35

	satisfies := v1.GreaterThanOrEqual(min)
	if !satisfies {
		fmt.Printf("%s is not >= %s", v1, min)
	}
	return satisfies
}

func checkFolder(folder *Folder) error {

	result, status, code, err := sendRequestArray(opts, "GET", "/api/folders?limit=1000", nil)

	if err != nil {
		return err
	}

	if code != 200 {
		return errors.New("server response: " + status)
	}

	if len(result) == 0 {
		return nil
	}

	for _, x := range result {
		if name, ok := x["title"]; ok {
			if name.(string) == folder.name {
				if id, idExist := x["id"]; idExist {
					folder.id = int64(id.(float64))
					if uid, uidExist := x["uid"]; uidExist {
						folder.uid = uid.(string)
					}
				}
			}
		}
	}

	return nil
}

func createServerFolder(folder *Folder) error {

	var request map[string]interface{}
	request = make(map[string]interface{})

	request["title"] = folder.name

	result, status, code, err := sendRequest(opts, "POST", "/api/folders", request)

	if err != nil {
		return err
	}

	if code != 200 {
		return errors.New("server response: " + status)
	}

	folder.id = int64(result["id"].(float64))
	folder.uid = result["uid"].(string)

	return nil
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
	Use:   "import",
	Short: "import Grafana dashboards",
	// Added so directory and serverfolder are required arguments except when both are empty. When both are empty use long accepted Harvest defaults
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		dir := cmd.Flags().Lookup("directory")
		folder := cmd.Flags().Lookup("serverfolder")
		if dir.Value.String() == "" && folder.Value.String() == "" {
			dir.Changed = true
			folder.Changed = true
		}
	},
	Run: doImport,
	Example: `
# Add the default set of cDot and 7mode dashboards from local directory grafana/dashboards to my.grafana.server 
grafana import --addr my.grafana.server:3000

# Add the dashboards from the local directory to the server_folder on my.grafana.server
grafana import --addr my.grafana.server:3000 --directory [local] --serverfolder [server_folder]`,
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export Grafana dashboards",
	Run:   doExport,
	Example: `
# Export all of the dashboards contained in the server_folder on my.grafana.server and write them to the local directory
grafana export --addr my.grafana.server:3000 --serverfolder server_folder --directory local`,
}

func init() {
	Cmd.AddCommand(importCmd, exportCmd)

	Cmd.PersistentFlags().StringVar(&opts.config, "config", "./harvest.yml", "harvest config file path")
	Cmd.PersistentFlags().StringVarP(&opts.addr, "addr", "a", "http://127.0.0.1:3000", "Address of Grafana server (IP, FQDN or hostname)")
	Cmd.PersistentFlags().StringVarP(&opts.token, "token", "t", "", "API token issued by Grafana server for authentication")
	Cmd.PersistentFlags().StringVarP(&opts.prefix, "prefix", "p", "", "Use global metric prefix in queries")
	Cmd.PersistentFlags().StringVarP(&opts.datasource, "datasource", "s", grafanaDataSource, "Grafana datasource for the dashboards")
	Cmd.PersistentFlags().BoolVarP(&opts.variable, "variable", "v", false, "Use datasource as variable, overrides: --datasource")
	Cmd.PersistentFlags().BoolVarP(&opts.useHttps, "https", "S", false, "Use HTTPS")
	Cmd.PersistentFlags().BoolVarP(&opts.overwrite, "overwrite", "o", false, "Overwrite existing dashboard with same title")
	Cmd.PersistentFlags().BoolVarP(&opts.useInsecureTLS, "insecure", "k", false, "Allow insecure server connections when using SSL")
	Cmd.PersistentFlags().StringVarP(&opts.serverfolder.name, "serverfolder", "f", "", "Grafana folder name for dashboards")
	Cmd.PersistentFlags().StringVarP(&opts.dir, "directory", "d", "", "When importing, import dashboards from this local directory.\nWhen exporting, local directory to write dashboards to")

	importCmd.PersistentFlags().StringSliceVar(&opts.labels, "labels", nil,
		"For each label, create a variable and add as chained query to other variables")
	importCmd.PersistentFlags().BoolVar(&opts.addMultiSelect, "multi", false,
		"Modify the dashboards to add multi-select dropdowns for each variable")

	_ = Cmd.MarkPersistentFlagRequired("serverfolder")
	_ = Cmd.MarkPersistentFlagRequired("directory")
}
