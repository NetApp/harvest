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
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/slogx"
	goversion "github.com/netapp/harvest/v2/third_party/go-version"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"github.com/netapp/harvest/v2/third_party/tidwall/sjson"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	clientTimeout                 = 5
	DefaultDataSource             = "prometheus"
	GPerm             os.FileMode = 0644
)

var Dashboards = []string{
	"../../../grafana/dashboards/cisco",
	"../../../grafana/dashboards/cmode",
	"../../../grafana/dashboards/cmode-details",
	"../../../grafana/dashboards/storagegrid",
}

var (
	grafanaMinVers = "7.1.0" // lowest grafana version we require
	homePath       string
	grafanaVersion *goversion.Version
)

type options struct {
	command             string // one of: import, export, clean
	addr                string // URL of Grafana server (e.g. "http://localhost:3000")
	token               string // API token issued by Grafana server
	dir                 string
	serverfolder        Folder
	datasource          string
	variable            bool
	forceImport         bool
	client              *http.Client
	headers             http.Header
	config              string
	prefix              string
	useHTTPS            bool
	useInsecureTLS      bool
	overwrite           bool
	labels              []string
	dirGrafanaFolderMap map[string]*Folder
	addMultiSelect      bool
	svmRegex            string
	customizeDir        string
	customAllValue      string
	customCluster       string
	varDefaults         string
	defaultDropdownMap  map[string][]string
	isDebug             bool
	showDatasource      bool // show datasource variable in the dashboard
}

type Folder struct {
	name      string // Grafana folder where to upload from where to download dashboards
	id        int64
	uid       string
	parentUID string // If nested folders are enabled, and the folder is nested, this is the parent folder's uid
}

func adjustOptions() {
	opts.config = conf.ConfigPath(opts.config)
	homePath = conf.Path("")
	opts.dirGrafanaFolderMap = make(map[string]*Folder)

	// When opt.addr starts with https don't change it
	if !strings.HasPrefix(opts.addr, "https://") {
		//goland:noinspection HttpUrlsUsage
		opts.addr = strings.TrimPrefix(opts.addr, "http://")
		opts.addr = strings.TrimPrefix(opts.addr, "https://")
		opts.addr = strings.TrimSuffix(opts.addr, "/")
		if opts.useHTTPS {
			opts.addr = "https://" + opts.addr
		} else {
			//goland:noinspection HttpUrlsUsage
			opts.addr = "http://" + opts.addr
		}
	}
}

func askForToken() {
	// ask for API token if not provided as arg and validate
	if err := checkToken(opts, false, 5); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func doCustomize(_ *cobra.Command, _ []string) {
	adjustOptions()
	exitIfExist(opts.customizeDir, "output-dir")

	initImportVars()
	importDashboards(opts)
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

	if err = os.MkdirAll(dir, 0750); err != nil {
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
			fp := filepath.Join(dir, uri+".json")
			data, err := json.Marshal(dashboard)
			if err != nil {
				fmt.Printf("error marshall dashboard [%s]: %v\n\n", uid, err)
				return err
			}
			// creating dashboards with group and other permissions of read are OK
			if err = os.WriteFile(fp, data, GPerm); err != nil {
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

func addSvmRegex(content []byte, val string) []byte {
	svmExpression := []string{"templating.list.#(name=\"SVM\")"}
	for _, s := range svmExpression {
		var err error
		svm := gjson.GetBytes(content, s)
		if svm.Exists() {
			content, err = sjson.SetBytes(content, s+".regex", []byte(val))
			if err != nil {
				fmt.Printf("Error while setting svm regex: %v\n", err)
				continue
			}
			content, err = sjson.SetBytes(content, s+".allValue", json.RawMessage("null"))
			if err != nil {
				fmt.Printf("Error while setting svm allValue: %v\n", err)
				continue
			}
		}
	}
	return content
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
	var newVars []gjson.Result
	newVars = append(newVars, vars[0])
	newVars = append(newVars, vars...)

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
			varsString.WriteString(result.ClonedString())
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

	// Rewrite all panel expressions to include the new label

	forLabel := varNameForLabel(label)
	VisitAllPanels(newContent, func(path string, _, value gjson.Result) {
		value.Get("targets").ForEach(func(targetKey, target gjson.Result) bool {
			expr := target.Get("expr")
			if expr.Exists() {
				newExpression := rewriteExprWith(expr.ClonedString(), label, forLabel)
				loc := path + ".targets." + targetKey.ClonedString() + ".expr"
				newContent, err = sjson.SetBytes(newContent, loc, []byte(newExpression))
				if err != nil {
					fmt.Printf("error rewriting expr at=[%s] for label=[%s], err: %+v", loc, label, err)
					return false
				}
			}
			return true
		})
	})

	return newContent
}

var labelsRegex = regexp.MustCompile(`\{([^}]+)}`)

func rewriteExprWith(input string, label string, forLabel string) string {
	result := input

	allMatches := labelsRegex.FindAllStringSubmatch(input, -1)

	for _, match := range allMatches {
		if len(match) < 2 {
			continue
		}
		// Check if the label is already present in the match
		if strings.Contains(match[1], label) {
			continue
		}
		// Add the new label to the match
		// e.g., NATE_UUID=~"$Nate_uuid"
		newLabel := label + `=~"$` + forLabel + `"`
		result = strings.ReplaceAll(result, match[0], fmt.Sprintf("{%s,%s}", match[1], newLabel))
	}

	return result
}

func addChainedVar(result gjson.Result, label string, labelMap map[string]string) string {
	varName := result.Get("name")
	definition := result.Get("definition")
	query := result.Get("query.query")

	if !varName.Exists() || !definition.Exists() || !query.Exists() {
		return ""
	}
	// Don't modify the query if this is one of the new labels we're adding since its query is already correct
	if _, ok := labelMap[varName.ClonedString()]; ok {
		return ""
	}

	defStr := definition.ClonedString()
	if defStr != query.ClonedString() {
		return ""
	}
	chained := toChainedVar(defStr, label)
	if chained == "" {
		return ""
	}
	rString := result.ClonedString()
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

	// Only rewrite allValue when a customAllValue is specified and "includeAll" is true
	if opts.customAllValue != "" && result.Get("includeAll").Bool() {
		if opts.customAllValue == "null" || opts.customAllValue == "nil" {
			rString, err = sjson.Set(rString, "allValue", nil)
		} else {
			rString, err = sjson.Set(rString, "allValue", opts.customAllValue)
		}

		if err != nil {
			fmt.Printf("error setting allValue of varName=[%s] for label=[%s], err: %+v", varName, label, err)
			return ""
		}
	}

	return rString
}

func toChainedVar(defStr string, label string) string {
	if !strings.Contains(defStr, "label_values") {
		return ""
	}

	title := cases.Title(language.Und).String(label)
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
		// There are two cases:
		// 1. There are no existing labels, e.g.,
		// label_values(datacenter) becomes label_values({foo=~"$Foo"}, datacenter)
		// 2. There is a single metric, e.g.,
		// label_values(poller_status, datacenter) becomes label_values(poller_status{org=~"$org"}, datacenter)

		if lastComma != -1 {
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
  "refresh": 2,
  "regex": "",
  "skipUrlSync": false,
  "sort": 0,
  "type": "query"
}`, label, cases.Title(language.Und).String(label), label))
}

func doImport(_ *cobra.Command, _ []string) {
	opts.command = "import"
	_, err := conf.LoadHarvestConfig(opts.config)
	if err != nil {
		printErrorAndExit(err)
	}

	setupSlog()
	adjustOptions()
	validateImport()
	askForToken()
	initImportVars()

	fmt.Printf("preparing to import dashboards...\n")
	importDashboards(opts)
}

func setupSlog() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				source.File = filepath.Base(source.File)
			}
			return a
		},
	}))
	slog.SetDefault(logger)
}

func printErrorAndExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func validateImport() {
	// default case
	if opts.dir == "" && opts.serverfolder.name == "" {
		opts.dir = "grafana/dashboards"
		opts.dir = filepath.Join(homePath, opts.dir)
	}

	exitIfMissing(opts.dir, "directory")
	files, err := os.ReadDir(opts.dir)
	if err != nil {
		fmt.Printf("Error %v while reading dir [%s] is the directory correct?\n", err, opts.dir)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Printf("No dashboards found in [%s] is the directory correct?\n", opts.dir)
		os.Exit(1)
	}

	if opts.isDebug {
		token := opts.token
		opts.token = "****"
		slog.Default().Info("validateImport", slog.Any("opts", opts))
		opts.token = token
	}
}

func initImportVars() {
	m := make(map[string]*Folder)

	// default behaviour
	switch {
	case opts.dir == "grafana/dashboards" && opts.serverfolder.name == "":
		m[filepath.Join(opts.dir, "cmode")] = &Folder{name: "Harvest-main-cDOT"}
		m[filepath.Join(opts.dir, "cmode-details")] = &Folder{name: "Harvest-main-cDOT Details"}
		m[filepath.Join(opts.dir, "cisco")] = &Folder{name: "Harvest-main-Cisco"}
		m[filepath.Join(opts.dir, "7mode")] = &Folder{name: "Harvest-main-7mode"}
		m[filepath.Join(opts.dir, "storagegrid")] = &Folder{name: "Harvest-main-StorageGrid"}
	case opts.dir != "" && opts.serverfolder.name != "":
		m[opts.dir] = &Folder{name: opts.serverfolder.name}
	case opts.dir != "" && opts.customizeDir != "":
		m[opts.dir] = &Folder{name: opts.dir}
	}

	for k, v := range m {
		if opts.customizeDir == "" {
			err := checkAndCreateServerFolder(v)
			if err != nil {
				printErrorAndExit(err)
			}
		}
		opts.dirGrafanaFolderMap[k] = v
	}

	// Parse default dropdown values
	opts.defaultDropdownMap = make(map[string][]string)
	if opts.varDefaults != "" {
		if !validateVarDefaults(opts.varDefaults) {
			fmt.Println("Error: Invalid format for --var-defaults. Expected format is 'variable1=value1,value2;variable2=value3'")
			os.Exit(1)
		}
		pairs := strings.Split(opts.varDefaults, ";")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				values := strings.Split(parts[1], ",")
				opts.defaultDropdownMap[parts[0]] = values
			}
		}
	}
}

// validateVarDefaults validates the format of the --var-defaults input string.
// The expected format is 'variable1=value1,value2;variable2=value3'.
func validateVarDefaults(input string) bool {
	re := regexp.MustCompile(`^([^=;,]+=[^=;,]+(,[^=;,]+)*)(;[^=;,]+=[^=;,]+(,[^=;,]+)*)*$`)
	return re.MatchString(input)
}

func checkAndCreateServerFolder(folder *Folder) error {
	err := checkFolder(folder)
	if err != nil {
		fmt.Printf("error %v for folder %s\n", err, folder.name)
		os.Exit(1)
	}

	folderName := folder.name
	if folder.uid != "" && folder.id != 0 {
		fmt.Printf("folder [%s] exists in Grafana - OK\n", folder.name)
	} else if err := createServerFolders(folder); err != nil {
		return err
	} else {
		fmt.Printf("created Grafana folder [%s] - OK\n", folderName)
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
		fmt.Printf("error: %s folder [%s] exists. Please specify an empty or non-existent directory.\n", s, fp)
		os.Exit(1)
	}
}

func importDashboards(opts *options) {
	if opts.overwrite {
		fmt.Printf("warning: The overwrite flag is no longer used and will be removed in a future release. Please remove this flag from your command line invocation. When importing, dashboards are always overwritten.\n")
	}
	// Set overwrite flag to true, dashboards are always overwritten.
	opts.overwrite = true

	for k, v := range opts.dirGrafanaFolderMap {
		importFiles(k, v)
	}
}

func importFiles(dir string, folder *Folder) {
	var (
		request, dashboard map[string]any
		files              []os.DirEntry
		importedFiles      int
		data               []byte
		err                error
	)
	if dir == "" {
		return
	}
	if files, err = os.ReadDir(dir); err != nil {
		printErrorAndExit(err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		if data, err = os.ReadFile(filepath.Join(dir, file.Name())); err != nil {
			fmt.Printf("error reading file [%s]\n", file.Name())
			return
		}

		data = bytes.ReplaceAll(data, []byte("${DS_PROMETHEUS}"), []byte(opts.datasource))

		// add svm regex
		if opts.svmRegex != "" {
			data = addSvmRegex(data, opts.svmRegex)
		}

		// change cluster label if needed
		if opts.customCluster != "" {
			data = changeClusterLabel(data, opts.customCluster)
		}

		// Set default dropdown values if provided
		if len(opts.defaultDropdownMap) > 0 {
			data = setDefaultDropdownValues(data, opts.defaultDropdownMap)
		}

		// labelMap is used to ensure we don't modify the query of one of the new labels we're adding
		labelMap := make(map[string]string)
		for _, label := range opts.labels {
			labelMap[varNameForLabel(label)] = label
		}
		// The label is inserted in the list of variables first
		// Iterate backwards so the labels keep the same order as cmdline
		for i := len(opts.labels) - 1; i >= 0; i-- {
			data = addLabel(data, opts.labels[i], labelMap)
		}

		if opts.showDatasource {
			data, _ = sjson.SetBytes(data, "templating.list.0.hide", 0)
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

		if opts.customizeDir != "" {
			err := writeCustomDashboard(dashboard, dir, file)
			if err != nil {
				fmt.Printf("error customizing dashboard [%s] %+v\n", file.Name(), err)
			}
			continue
		}

		request = make(map[string]any)
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
				fmt.Printf("If you want to overwrite, run with the --overwrite flag or choose a different Grafana folder with the --serverfolder flag.\n")
				fmt.Printf("NOTE: If the dashboard already exists, --overwrite will import the changes and increment the dashboard version number.\n\n")
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

	if opts.customizeDir == "" {
		if importedFiles > 0 {
			fmt.Printf("Imported %d dashboards to [%s] from [%s]\n", importedFiles, folder.name, dir)
		} else {
			fmt.Printf("No dashboards found in [%s] is the directory correct?\n", dir)
		}
	}
}

var caser = cases.Title(language.Und)

func varNameForLabel(label string) string {
	return caser.String(label)
}

// setDefaultDropdownValues sets the default values for specified dropdown variables in the dashboard JSON data.
// It takes a map of variable names to their default values and updates the JSON data accordingly.
func setDefaultDropdownValues(data []byte, defaultValues map[string][]string) []byte {
	for variable, defaultValues := range defaultValues {
		variablePath := fmt.Sprintf("templating.list.#(name=%q)", variable)
		variableData := gjson.GetBytes(data, variablePath)
		if variableData.Exists() {
			current := map[string]any{
				"selected": true,
				"text":     defaultValues,
				"value":    defaultValues,
			}
			data, _ = sjson.SetBytes(data, variablePath+".current", current)
		}
	}
	return data
}

// This function will rewrite all panel expressions in the dashboard to use the new cluster label.
// Example:
// sum(write_data{datacenter=~"$Datacenter",cluster=~"$Cluster",svm=~"$SVM"})
// with --cluster-label=na_cluster will become
// sum(write_data{datacenter=~"$Datacenter",na_cluster=~"$Cluster",svm=~"$SVM"})
// See https://github.com/NetApp/harvest/issues/3131
func changeClusterLabel(data []byte, cluster string) []byte {

	// Change all panel expressions
	VisitAllPanels(data, func(path string, _, value gjson.Result) {

		// Rewrite expressions and legends
		value.Get("targets").ForEach(func(targetKey, target gjson.Result) bool {
			expr := target.Get("expr")
			if expr.Exists() {
				newExpression := rewriteCluster(expr.ClonedString(), cluster)

				data, _ = sjson.SetBytes(data, path+".targets."+targetKey.ClonedString()+".expr", []byte(newExpression))
			}

			legendFormat := target.Get("legendFormat")
			if legendFormat.Exists() {
				newLegendFormat := rewriteCluster(legendFormat.ClonedString(), cluster)

				data, _ = sjson.SetBytes(data, path+".targets."+targetKey.ClonedString()+".legendFormat", []byte(newLegendFormat))
			}

			return true
		})

		// Rewrite tables columns
		panelType := value.Get("type")
		if panelType.ClonedString() == "table" {
			value.Get("transformations").ForEach(func(transKey, value gjson.Result) bool {
				id := value.Get("id")
				if id.ClonedString() == "organize" {

					// Check if the cluster exists in renameByName, and if so, rename it to the new cluster label
					clusterTrans := value.Get("options.renameByName.cluster")
					if clusterTrans.Exists() {
						data, _ = sjson.SetBytes(data, path+".transformations."+transKey.ClonedString()+".options.renameByName."+cluster, []byte(clusterTrans.ClonedString()))
					}

					// If the cluster column exists, remove the column, and add the new cluster label at the same index
					clusterIndex := value.Get("options.indexByName.cluster")
					if clusterIndex.Exists() {
						data, _ = sjson.SetBytes(data, path+".transformations."+transKey.ClonedString()+".options.indexByName."+cluster, clusterIndex.Int())
						data, _ = sjson.DeleteBytes(data, path+".transformations."+transKey.ClonedString()+".options.indexByName.cluster")
					}
					// Handle the case where the cluster column is named "cluster 1", "cluster 2", etc.
					for i := range 10 {
						clusterN := "cluster " + strconv.Itoa(i)
						clusterIndexI := value.Get("options.indexByName." + clusterN)
						if clusterIndexI.Exists() {
							data, _ = sjson.SetBytes(data, path+".transformations."+transKey.ClonedString()+".options.indexByName."+cluster+" "+strconv.Itoa(i), clusterIndexI.Int())
							data, _ = sjson.DeleteBytes(data, path+".transformations."+transKey.ClonedString()+".options.indexByName."+clusterN)
						}
					}

					// If cluster is excluded from the table, exclude the new cluster label too
					excludeByName := value.Get("options.excludeByName")
					if excludeByName.Exists() {
						clusterIndex := value.Get("options.excludeByName.cluster")
						if clusterIndex.Exists() {
							data, _ = sjson.SetBytes(data, path+".transformations."+transKey.ClonedString()+".options.excludeByName."+cluster, clusterIndex.Int())
							data, _ = sjson.DeleteBytes(data, path+".transformations."+transKey.ClonedString()+".options.excludeByName.cluster")
						}
						// Handle the case where the cluster column is named "cluster 1", "cluster 2", etc.
						for i := range 10 {
							clusterN := "cluster " + strconv.Itoa(i)
							clusterIndexI := value.Get("options.excludeByName." + clusterN)
							if clusterIndexI.Exists() {
								data, _ = sjson.SetBytes(data, path+".transformations."+transKey.ClonedString()+".options.excludeByName."+cluster+" "+strconv.Itoa(i), true)
								data, _ = sjson.DeleteBytes(data, path+".transformations."+transKey.ClonedString()+".options.excludeByName."+clusterN)
							}
						}
					}
				} else if id.ClonedString() == "filterFieldsByName" {
					// Check if the cluster exists in filterFieldsByName, and if so, rename it to the new cluster label
					names := value.Get("options.include.names")
					if names.Exists() {
						var nameValues []string
						hasCluster := false
						for _, name := range names.Array() {
							if name.ClonedString() == "cluster" {
								hasCluster = true
								nameValues = append(nameValues, cluster)
								continue
							}
							nameValues = append(nameValues, name.ClonedString())
						}

						if hasCluster {
							data, _ = sjson.SetBytes(data, path+".transformations."+transKey.ClonedString()+".options.include.names", nameValues)
						}
					}
				}

				return true
			})

			// Change all fieldConfig overrides that contain cluster
			value.Get("fieldConfig.overrides").ForEach(func(overrideKey, override gjson.Result) bool {
				if override.Get("matcher.id").ClonedString() == "byName" && override.Get("matcher.options").ClonedString() == "cluster" {
					data, _ = sjson.SetBytes(data, path+".fieldConfig.overrides."+overrideKey.ClonedString()+".matcher.options", cluster)
				}
				return true
			})
		}
	})

	// Change all templating variables that contain cluster
	gjson.GetBytes(data, "templating.list").ForEach(func(key, value gjson.Result) bool {

		// Change definition
		definition := value.Get("definition")
		if definition.Exists() {
			newDefinition := rewriteCluster(definition.ClonedString(), cluster)

			data, _ = sjson.SetBytes(data, "templating.list."+key.ClonedString()+".definition", []byte(newDefinition))
		}

		// Change query
		query := value.Get("query.query")
		if query.Exists() {
			newQuery := rewriteCluster(query.ClonedString(), cluster)
			data, _ = sjson.SetBytes(data, "templating.list."+key.ClonedString()+".query.query", []byte(newQuery))
		}

		return true
	})

	return data
}

func rewriteCluster(input string, cluster string) string {
	const marker = `!!^`
	hiddenNames := []string{"cluster_new_status", "source_cluster"}
	for _, name := range hiddenNames {
		if strings.Contains(input, name) {
			// hide name
			repl := strings.ReplaceAll(name, "cluster", marker)
			input = strings.ReplaceAll(input, name, repl)
		}
	}

	result := strings.ReplaceAll(input, "cluster", cluster)

	// Restore hidden names
	result = strings.ReplaceAll(result, marker, "cluster")

	return result
}

func writeCustomDashboard(dashboard map[string]any, dir string, file os.DirEntry) error {
	data, err := json.Marshal(dashboard)
	if err != nil {
		return err
	}

	data, err = formatJSON(data)
	if err != nil {
		return err
	}
	sub := filepath.Base(dir)
	fp := filepath.Join(opts.customizeDir, sub, file.Name())
	if err := os.MkdirAll(filepath.Dir(fp), 0750); err != nil {
		return fmt.Errorf("error makedir [%s]: %w", filepath.Dir(fp), err)
	}
	if err = os.WriteFile(fp, data, GPerm); err != nil {
		return fmt.Errorf("error writing customized dashboard to file %s: %w", fp, err)
	}
	fmt.Printf("OK - customized [%s]\n", fp)
	return nil
}

func formatJSON(data []byte) ([]byte, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		return nil, fmt.Errorf("failed to format json %w", err)
	}
	return prettyJSON.Bytes(), nil
}

// addGlobalPrefix adds the given prefix to all metric names in the dashboards.
// It assumes that metrics are in Prometheus format.
//
// A more reliable implementation of this feature would be, to
// add a constant prefix to all metrics, before they are pushed
// to GitHub, then replace them with a user-defined prefix
// (or empty string) when the import tool is used.
func addGlobalPrefix(db map[string]any, prefix string) {

	var (
		panels, templates, subPanels       []interface{}
		panel, templating, template, query map[string]any
		queryString, definition            string
		ok, has                            bool
	)

	// make sure prefix ends with _
	if !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}

	// apply to queries in panels
	if panels, ok = db["panels"].([]interface{}); !ok {
		return
	}

	for _, p := range panels {
		handlingPanels(p, prefix)

		// handling for sub-panels
		if panel, ok = p.(map[string]any); !ok {
			continue
		}

		if _, has = panel["panels"]; !has {
			continue
		}
		if subPanels, ok = panel["panels"].([]interface{}); ok {
			for _, subP := range subPanels {
				handlingPanels(subP, prefix)
			}
		}
	}

	// apply to variables
	if templating, ok = db["templating"].(map[string]any); !ok {
		return
	}

	if templates, ok = templating["list"].([]interface{}); !ok {
		return
	}

	for _, t := range templates {
		if template, ok = t.(map[string]any); ok {
			if definition, ok = template["definition"].(string); ok {
				template["definition"] = addPrefixToMetricNames(definition, prefix)
			}
			if query, ok = template["query"].(map[string]any); ok {
				if queryString, ok = query["query"].(string); ok {
					query["query"] = addPrefixToMetricNames(queryString, prefix)
				}
			}
		}
	}
}

func handlingPanels(p interface{}, prefix string) {
	var (
		targets       []interface{}
		panel, target map[string]any
		ok, has       bool
		expr          string
	)
	if panel, ok = p.(map[string]any); !ok {
		return
	}

	if _, has = panel["targets"]; !has {
		return
	}

	if targets, ok = panel["targets"].([]interface{}); !ok {
		return
	}

	for _, t := range targets {
		if target, ok = t.(map[string]any); !ok {
			continue
		}
		if _, has = target["expr"]; has {
			if expr, ok = target["expr"].(string); ok {
				target["expr"] = addPrefixToMetricNames(expr, prefix)
			}
		}
	}
}

// addPrefixToMetricNames adds prefix to metric names in expr or leaves it
// unchanged if no metric names are identified.
// Note that this function will only work with the Prometheus dashboards of Harvest.
// It will use a number of patterns in which metrics might be used in queries.
// (E.g., a single metric, multiple metrics used in addition, etc. -- See the tests for examples)
// If we change queries of our dashboards, we have to review
// this function as well (or come up with a better solution).
func addPrefixToMetricNames(expr, prefix string) string {
	var (
		match      [][]string
		submatch   []string
		isMatch    bool
		regex      *regexp.Regexp
		err        error
		visitedMap map[string]bool // handles if the same query exists in multiple times in one expression.
	)

	// variable queries
	if strings.HasPrefix(expr, "label_values(") {
		if isMatch, err = regexp.MatchString(`^label_values\s?\(([a-zA-Z_])+(\s?{.+?})?,\s?[a-zA-Z_]+\)$`, expr); err != nil {
			fmt.Printf("Regex error: %v\n", err)
			return expr
		} else if isMatch {
			return strings.Replace(expr, "label_values(", "label_values("+prefix, 1)
		}
		return expr
	}
	// everything else is for graph queries
	regex = regexp.MustCompile(`([a-zA-Z0-9_+-]+)\s?{.+?}`)
	match = regex.FindAllStringSubmatch(expr, -1)
	visitedMap = make(map[string]bool)
	for _, m := range match {
		if _, has := visitedMap[m[1]]; !has {
			// multiple metrics used with `+`
			switch {
			case strings.Contains(m[1], "+"):
				submatch = strings.Split(m[1], "+")
				for i := range submatch {
					submatch[i] = prefix + submatch[i]
				}
				expr = strings.ReplaceAll(expr, m[1], strings.Join(submatch, "+"))
			case strings.Contains(m[1], "-"):
				submatch = strings.Split(m[1], "-")
				for i := range submatch {
					submatch[i] = prefix + submatch[i]
				}
				expr = strings.ReplaceAll(expr, m[1], strings.Join(submatch, "-"))
			default:
				expr = strings.ReplaceAll(expr, m[1], prefix+m[1])
			}
			visitedMap[m[1]] = true
		}
	}

	return expr
}

func checkToken(opts *options, ignoreConfig bool, tries int) error {

	var (
		token, configPath, answer string
		err                       error
		isValidDS                 bool
	)

	if tries == 0 {
		return errors.New("no more attempts")
	}

	configPath = opts.config

	_, err = conf.LoadHarvestConfig(configPath)
	if err != nil {
		return err
	}

	if conf.Config.Tools != nil {
		if !ignoreConfig {
			token = conf.Config.Tools.GrafanaAPIToken
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

	opts.client = &http.Client{
		Timeout:       time.Duration(clientTimeout) * time.Second,
		CheckRedirect: refuseRedirect,
	}
	if strings.HasPrefix(opts.addr, "https://") {
		tlsConfig := &tls.Config{InsecureSkipVerify: opts.useInsecureTLS} //nolint:gosec
		opts.client.Transport = &http.Transport{TLSClientConfig: tlsConfig}
	}
	// send random request to validate token
	result, status, code, err := sendRequest(opts, "GET", "/api/org", nil)
	if err != nil {
		return err
	} else if code != 200 && code != 404 {
		msg := result["message"].(string)
		fmt.Printf("error connect: (%d - %s) %s\n", code, status, msg)
		opts.token = ""
		return checkToken(opts, true, tries-1)
	}

	// ask user to save API key
	if conf.Config.Tools == nil || opts.token != conf.Config.Tools.GrafanaAPIToken {

		fmt.Printf("save API key for later use? [Y/n]: ")
		_, _ = fmt.Scanf("%s\n", &answer)

		if answer == "Y" || answer == "y" || answer == "yes" || answer == "" {
			if conf.Config.Tools == nil {
				conf.Config.Tools = &conf.Tools{}
			}
			conf.Config.Tools.GrafanaAPIToken = opts.token
			fmt.Printf("saving config file [%s]\n", configPath)
			if err := conf.SaveConfig(configPath, opts.token); err != nil {
				return err
			}
		}
	}

	// get grafana version
	if result, _, _, err = sendRequest(opts, "GET", "/api/frontend/settings", nil); err != nil {
		return err
	}

	if opts.forceImport {
		isValidDS = true
	} else {
		isValidDS = isValidDatasource(result)
	}
	if !isValidDS {
		fmt.Printf("A Prometheus-typed datasource named \"%s\" does not exist in Grafana.\n", opts.datasource)
		os.Exit(0)
	}

	buildInfo := result["buildInfo"].(map[string]any)
	if buildInfo == nil {
		fmt.Printf("warning: unable to get grafana version. Ignoring grafana version check")
		return nil
	}
	grafanaVersion := buildInfo["version"].(string)
	if grafanaVersion == "" {
		fmt.Printf("warning: unable to get grafana version. Ignoring grafana version check")
		return nil
	}
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

func refuseRedirect(req *http.Request, _ []*http.Request) error {
	// Refuse to follow redirects, see https://github.com/NetApp/harvest/issues/3617
	if req.Response != nil {
		loc := req.Response.Header.Get("Location")
		if loc != "" {
			return fmt.Errorf("redirect not allowed. location=[%s] Check that addr is using the correct URL", loc)
		}
	}

	return errors.New("redirect not allowed. Check that addr is using the correct URL")
}

func isValidDatasource(result map[string]any) bool {
	// If opts.datasource contains a $ we assume it's a variable and skip the check
	if strings.Contains(opts.datasource, "$") {
		return true
	}
	if result == nil {
		fmt.Printf("warning: result is null.")
		return false
	}
	datasourcesInfo := result["datasources"]
	if datasourcesInfo == nil {
		fmt.Printf("warning: datasources are missing")
		return false
	}

	// iterate in deterministic order over datasources
	dsNames := slices.Sorted(maps.Keys(datasourcesInfo.(map[string]any)))

	for _, dsName := range dsNames {
		dsInfo := datasourcesInfo.(map[string]any)[dsName]
		if dsInfo == nil {
			fmt.Printf("warning: dsInfo is missing for %s\n", dsName)
			continue
		}
		dsDetail := dsInfo.(map[string]any)
		dsType := dsDetail["type"]
		if dsType == nil {
			fmt.Printf("warning: dsType is missing for %s\n", dsName)
			continue
		}
		if strings.EqualFold(dsType.(string), "prometheus") && strings.EqualFold(dsName, opts.datasource) {
			// overwrite datasource name with the one from Grafana when the names differ by case
			opts.datasource = dsName
			return true
		}
	}

	return false
}

func checkVersion(inputVersion string) bool {
	var err error

	grafanaVersion, err = goversion.NewVersion(inputVersion)
	if err != nil {
		fmt.Println(err)
		return false
	}
	minV, _ := goversion.NewVersion(grafanaMinVers)

	// Not using a constraint check since a pre-release version (e.g. 8.4.0-beta1) never matches
	// a constraint specified without a pre-release https://github.com/hashicorp/go-version/pull/35

	satisfies := grafanaVersion.GreaterThanOrEqual(minV)
	if !satisfies {
		fmt.Printf("%s is not >= %s", grafanaVersion, minV)
	}
	return satisfies
}

func checkFolder(folder *Folder) error {

	q := "/api/folders?limit=1000"

	if folder.parentUID != "" {
		q += "&parentUid=" + folder.parentUID
	}

	result, status, code, err := sendRequestArray(opts, "GET", q, nil)

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
	request := make(map[string]any)

	request["title"] = folder.name

	if folder.parentUID != "" {
		request["parentUid"] = folder.parentUID
	}

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

func createServerFolders(folder *Folder) error {

	if grafanaVersion == nil || grafanaVersion.LessThan(goversion.Must(goversion.NewVersion("11.0.0"))) {
		return createServerFolder(folder)
	}

	// handle nested folders
	folders := strings.Split(folder.name, "/")
	var parentUID string

	for _, f := range folders {
		curFolder := &Folder{name: f}
		if parentUID != "" {
			curFolder.parentUID = parentUID
		}

		if err := checkFolder(curFolder); err != nil {
			return err
		}

		if curFolder.id == 0 {
			curFolder.name = f
			if err := createServerFolder(curFolder); err != nil {
				return err
			}
		}

		parentUID = curFolder.uid
		folder.name = f
		folder.id = curFolder.id
	}

	return nil
}

func sendRequest(opts *options, method, url string, query map[string]any) (map[string]any, string, int, error) {

	var result map[string]any

	data, status, code, err := doRequestWithRetry(opts, method, url, query, 3)
	if err != nil {
		return result, status, code, err
	}

	if err = json.Unmarshal(data, &result); err != nil {
		fmt.Printf("raw response sr (%d - %s):\n", code, status)
		fmt.Println(string(data))
	}
	return result, status, code, err
}

func sendRequestArray(opts *options, method, url string, query map[string]any) ([]map[string]any, string, int, error) {

	var result []map[string]any

	data, status, code, err := doRequestWithRetry(opts, method, url, query, 3)
	if err != nil {
		return result, status, code, err
	}

	if err = json.Unmarshal(data, &result); err != nil {
		fmt.Printf("raw response sra (%d - %s):\n", code, status)
		fmt.Println(string(data))
	}
	return result, status, code, err
}

func doRequestWithRetry(opts *options, method, url string, query map[string]any, retry int) ([]byte, string, int, error) {
	var (
		status string
		code   int
		err    error
		data   []byte
	)

	for {
		retry--
		data, status, code, err = doRequest(opts, method, url, query)
		if err != nil || code >= 500 {
			fmt.Printf("  request failed, retrying. retry=%d code=%d err=%+v url=%s\n", retry, code, err, url)
			if retry > 0 {
				continue
			}
		}
		break
	}
	return data, status, code, err
}

func doRequest(opts *options, method, url string, query map[string]any) ([]byte, string, int, error) {

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
		request, err = requests.New(method, opts.addr+url, buf)
	} else {
		request, err = requests.New(method, opts.addr+url, nil)
	}

	if err != nil {
		return nil, status, code, err
	}

	request.Header = opts.headers

	if response, err = opts.client.Do(request); err != nil {
		if opts.isDebug {
			slog.Default().Info(
				"doRequest",
				slog.String("method", method),
				slog.String("url", url),
				slog.String("status", status),
				slog.Int("code", code),
				slogx.Err(err),
			)
		}
		return nil, status, code, err
	}

	status = response.Status
	code = response.StatusCode

	//goland:noinspection GoUnhandledErrorResult
	defer response.Body.Close()
	data, err = io.ReadAll(response.Body)
	if opts.isDebug {
		slog.Default().Info(
			"doRequest",
			slog.String("method", method),
			slog.String("url", url),
			slog.String("status", status),
			slog.String("data", string(data)),
		)
	}
	return data, status, code, err
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
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
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
# Export all the dashboards contained in the server_folder on my.grafana.server and write them to the local directory
grafana export --addr my.grafana.server:3000 --serverfolder server_folder --directory local`,
}

var customizeCmd = &cobra.Command{
	Use:   "customize",
	Short: "customize Grafana dashboards and write to filesystem",
	Run:   doCustomize,
	Example: `
# Customize all the dashboards recursively contained in grafana/dashboards and write them to ~/harvest-dashboards.
grafana customize --directory grafana/dashboards --output-dir ~/harvest-dashboards --prefix netapp_ --datasource my_datasource`,
}

func init() {
	Cmd.AddCommand(importCmd, exportCmd, customizeCmd, metricsCmd)
	addCommonFlags(importCmd, exportCmd, customizeCmd)
	addImportExportFlags(importCmd, exportCmd)
	addImportCustomizeFlags(importCmd, customizeCmd)
	addImportFlags(importCmd)

	customizeCmd.PersistentFlags().StringVarP(&opts.customizeDir, "output-dir", "o", "", "Write customized dashboards to the local directory. The directory must not exist")

	metricsCmd.PersistentFlags().StringVarP(&opts.dir, "directory", "d",
		"", "local directory that contains dashboards (searched recursively).")
}

func addImportFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&opts.varDefaults, "var-defaults", "", `
Default values for dropdown variables in the format 'variable1=value1,value2;variable2=value3'.

Examples:

1. Set a single variable:
   To set the default value for the 'Datacenter' variable to 'DC1':
   --var-defaults "Datacenter=DC1"

2. Set multiple values for a single variable:
   To set the default values for the 'Datacenter' variable to 'DC1' and 'DC2':
   --var-defaults "Datacenter=DC1,DC2"

3. Set multiple variables in one command:
   To set the default values for 'Datacenter' to 'DC1' and 'DC2', and 'Cluster' to 'Cluster1':
   --var-defaults "Datacenter=DC1,DC2;Cluster=Cluster1"

4. Set multiple values for multiple variables:
   To set the default values for 'Datacenter' to 'DC1' and 'DC2', and 'Cluster' to 'Cluster1' and 'Cluster2':
   --var-defaults "Datacenter=DC1,DC2;Cluster=Cluster1,Cluster2"

Note: Ensure that variable names and values do not contain the characters '=', ',', or ';' as these are used as delimiters.
`)
}

func addImportCustomizeFlags(commands ...*cobra.Command) {
	for _, cmd := range commands {
		cmd.PersistentFlags().StringSliceVar(&opts.labels, "labels", nil,
			"For each label, create a variable and add as chained query to other variables")
		cmd.PersistentFlags().StringVar(&opts.customAllValue, "customallvalue", "",
			"Modify each variable to use the specified custom all value.")
		cmd.PersistentFlags().BoolVar(&opts.addMultiSelect, "multi", true,
			"Modify the dashboards to add multi-select dropdowns for each variable")
		cmd.PersistentFlags().BoolVar(&opts.forceImport, "force", false,
			"Import even if the datasource name is not defined in Grafana")
		cmd.PersistentFlags().StringVar(&opts.customCluster, "cluster-label", "",
			"Rewrite all panel expressions to use the specified cluster label instead of the default 'cluster'")
		cmd.PersistentFlags().BoolVar(&opts.showDatasource, "show-datasource", false,
			"Show datasource variable dropdown in dashboards, useful for multi-datasource setups")

		_ = cmd.PersistentFlags().MarkHidden("multi")
		_ = cmd.PersistentFlags().MarkHidden("force")
	}
}

func addCommonFlags(commands ...*cobra.Command) {
	for _, cmd := range commands {
		cmd.PersistentFlags().StringVar(&opts.config, "config", "./harvest.yml", "harvest config file path")
		cmd.PersistentFlags().StringVar(&opts.svmRegex, "svm-variable-regex", "", "SVM variable regex to filter SVM query results")
		cmd.PersistentFlags().StringVarP(&opts.prefix, "prefix", "p", "", "Use global metric prefix in queries")
		cmd.PersistentFlags().StringVarP(&opts.datasource, "datasource", "s", DefaultDataSource, "Name of your Prometheus datasource used by the imported dashboards. Use '${DS_PROMETHEUS}' to include multiple Prometheus datasources.")
		cmd.PersistentFlags().BoolVarP(&opts.variable, "variable", "v", false, "Use datasource as variable, overrides: --datasource")
		cmd.PersistentFlags().StringVarP(&opts.dir, "directory", "d", "", "When importing, import dashboards from this local directory.\nWhen exporting, local directory to write dashboards to")
		cmd.PersistentFlags().BoolVar(&opts.isDebug, "debug", false, "Enable debug logging")

		_ = cmd.PersistentFlags().MarkHidden("svm-variable-regex")
		_ = cmd.MarkPersistentFlagRequired("directory")
	}
}

func addImportExportFlags(commands ...*cobra.Command) {
	for _, cmd := range commands {
		cmd.PersistentFlags().StringVarP(&opts.addr, "addr", "a", "http://127.0.0.1:3000", "Address of Grafana server (IP, FQDN or hostname)")
		cmd.PersistentFlags().StringVarP(&opts.token, "token", "t", "", "API token issued by Grafana server for authentication")
		cmd.PersistentFlags().BoolVarP(&opts.useHTTPS, "https", "S", false, "Use HTTPS")
		cmd.PersistentFlags().BoolVarP(&opts.overwrite, "overwrite", "o", false, "Overwrite existing dashboard with same title")
		cmd.PersistentFlags().BoolVarP(&opts.useInsecureTLS, "insecure", "k", false, "Allow insecure server connections when using SSL")
		cmd.PersistentFlags().StringVarP(&opts.serverfolder.name, "serverfolder", "f", "", "Grafana folder name for dashboards")

		_ = cmd.MarkPersistentFlagRequired("serverfolder")
	}
}
