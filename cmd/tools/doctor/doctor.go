package doctor

import (
	"errors"
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	harvestyaml "github.com/netapp/harvest/v2/pkg/tree/yaml"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type options struct {
	ShouldPrintConfig  bool
	Color              string
	BaseTemplate       string
	MergeTemplate      string
	zapiDataCenterName string
	restDataCenterName string
	prometheusURL      string
	expandVar          bool
}

var opts = &options{
	ShouldPrintConfig: false,
	Color:             "auto",
}

type validation struct {
	isValid bool
	invalid []string // collect invalid results
}

var Cmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check for potential problems",
	Long:  "Check for potential problems",
	Run:   doDoctorCmd,
}

var mergeCmd = &cobra.Command{
	Use:    "merge",
	Hidden: true,
	Short:  "merge templates",
	Run:    doMergeCmd,
}

var compareZapiRestMetricsCmd = &cobra.Command{
	Use:    "compareZRMetrics",
	Hidden: true,
	Short:  "Compare metrics between Zapi and Rest",
	Run:    doCompareZapiRestCmd,
}

func doCompareZapiRestCmd(_ *cobra.Command, _ []string) {
	missingMetrics, err := DoCompareZapiRest(opts.zapiDataCenterName, opts.restDataCenterName, opts.prometheusURL)
	if err != nil {
		fmt.Println("Metrics missing (excluding those in the ignore list):", strings.Join(missingMetrics, ", "))
		os.Exit(1)
	}
}

func doMergeCmd(_ *cobra.Command, _ []string) {
	doMerge(opts.BaseTemplate, opts.MergeTemplate)
}

func doMerge(path1 string, path2 string) {
	template, err := tree.ImportYaml(path1)
	if err != nil || template == nil {
		fmt.Printf("error reading template file [%s]. err=%+v\n", path1, err)
		return
	}
	subTemplate, err := tree.ImportYaml(path2)
	if err != nil || subTemplate == nil {
		fmt.Printf("error reading template file [%s] err=%+v\n", path2, err)
		return
	}
	template.PreprocessTemplate()
	subTemplate.PreprocessTemplate()
	template.Merge(subTemplate, nil)
	data, err := harvestyaml.Dump(template)
	if err != nil {
		fmt.Printf("error reading parsing template file [%s]  err=%+v\n", data, err)
		return
	}
	fmt.Println(string(data))
}

func doDoctorCmd(cmd *cobra.Command, _ []string) {
	var config = cmd.Root().PersistentFlags().Lookup("config")
	var confPaths = cmd.Root().PersistentFlags().Lookup("confpath")

	pathI := conf.ConfigPath(config.Value.String())
	confPath := confPaths.Value.String()
	out := doDoctor(pathI)
	if opts.ShouldPrintConfig {
		fmt.Println(out)
	}
	checkAll(pathI, confPath)
}

func doDoctor(aPath string) string {
	contents, err := os.ReadFile(aPath)
	if err != nil {
		fmt.Printf("error reading config file. err=%+v\n", err)
		return ""
	}

	if opts.expandVar {
		contents, err = conf.ExpandVars(contents)
		if err != nil {
			fmt.Printf("error reading config file. err=%+v\n", err)
			return ""
		}
	}

	parentRoot, err := printRedactedConfig(aPath, contents)
	if err != nil {
		fmt.Printf("error processing parent config file=[%s] %+v\n", aPath, err)
		return ""
	}

	// Extract Poller_files field from parentRoot
	var pollerFiles []string
	for _, kv := range parentRoot.(*ast.MappingNode).Values {
		if node.ToString(kv.Key) == "Poller_files" {
			seq := kv.Value.(*ast.SequenceNode)
			for _, value := range seq.Values {
				pollerFiles = append(pollerFiles, node.ToString(value))
			}
			break
		}
	}

	for _, childPathPattern := range pollerFiles {
		matches, err := filepath.Glob(childPathPattern)
		if err != nil {
			fmt.Printf("error matching glob pattern: %v\n", err)
			continue
		}
		for _, childPath := range matches {
			childContents, err := os.ReadFile(childPath)
			if err != nil {
				fmt.Printf("error reading child file. err=%+v\n", err)
				continue
			}
			childRoot, err := printRedactedConfig(childPath, childContents)
			if err != nil {
				fmt.Printf("error processing child config file=[%s] %+v\n", childPath, err)
				continue
			}

			// Merge childRoot into parentRoot
			mergeYamlNodes(parentRoot, childRoot)
		}
	}

	marshaled, err := yaml.Marshal(parentRoot)
	if err != nil {
		fmt.Printf("error marshalling yaml sanitized from config file=[%s] %+v\n", aPath, err)
		return ""
	}
	out := string(marshaled)
	return out
}

func mergeYamlNodes(parent ast.Node, child ast.Node) {
	// Find the Pollers section in the parent node
	var parentPollers ast.Node

	pmn := parent.(*ast.MappingNode)
	cmn := child.(*ast.MappingNode)

	for _, kv := range pmn.Values {
		if node.ToString(kv.Key) == "Pollers" {
			parentPollers = kv.Value
			break
		}
	}

	// If the parent node doesn't have a Pollers section, create one in the parent node
	if parentPollers == nil {
		pos := &token.Position{Column: 1, IndentLevel: 0, IndentNum: 0}
		newPollersNode := ast.Mapping(token.New("", "", pos), false)
		pmn.Values = append(pmn.Values, ast.MappingValue(
			nil,
			ast.String(token.New("Pollers", "", pos)),
			newPollersNode,
		))
		parentPollers = newPollersNode
	}

	// Create a map of the parent node's pollers
	parentPollerNames := make(map[string]bool)
	for _, kv := range parentPollers.(*ast.MappingNode).Values {
		parentPollerNames[node.ToString(kv.Key)] = true
	}

	// Find the Pollers section in the child node and append any child pollers that aren't already in the parent node
	for _, kv := range cmn.Values {
		if node.ToString(kv.Key) == "Pollers" {
			childPollers := kv.Value
			if childPollers == nil {
				continue
			}
			for _, kv := range childPollers.(*ast.MappingNode).Values {
				if _, exists := parentPollerNames[node.ToString(kv.Key)]; !exists {
					parentPollers.(*ast.MappingNode).Values = append(parentPollers.(*ast.MappingNode).Values, kv)
				}
			}
			break
		}
	}
}

// checkAll runs all doctor checks
// If all checks succeed, print nothing and exit with a return code of 0
// Otherwise, print what failed and exit with a return code of 1
func checkAll(aPath string, confPath string) {
	// See https://github.com/NetApp/harvest/issues/16 for more checks to add
	color.DetectConsole(opts.Color)

	_, err := conf.LoadHarvestConfig(aPath)
	if err != nil {
		fmt.Printf("error reading config file=[%s] %+v\n", aPath, err)
		os.Exit(1)
		return
	}

	cfg := conf.Config
	confPaths := filepath.SplitList(confPath)
	var anyFailed bool
	anyFailed = !checkExportersExist(cfg).isValid
	anyFailed = !checkUniquePromPorts(cfg).isValid || anyFailed
	anyFailed = !checkPollersExportToUniquePromPorts(cfg).isValid || anyFailed
	anyFailed = !checkExporterTypes(cfg).isValid || anyFailed
	anyFailed = !checkConfTemplates(confPaths).isValid || anyFailed
	anyFailed = !checkCollectorName(cfg).isValid || anyFailed
	anyFailed = !checkPollerPromPorts(cfg).isValid || anyFailed

	if anyFailed {
		os.Exit(1)
	}
	os.Exit(0)
}

// checkCollectorName checks if the collector names in the config struct are valid
func checkCollectorName(config conf.HarvestConfig) validation {
	valid := validation{isValid: true}

	var isDefaultCollectorExist bool
	// Verify default collectors
	if config.Defaults != nil {
		defaultCollectors := config.Defaults.Collectors
		for _, c := range defaultCollectors {
			isDefaultCollectorExist = true
			// Check if the collector name is valid
			_, ok := conf.IsCollector[c.Name]
			if !ok {
				fmt.Printf("Default Section uses an invalid collector %s \n", color.Colorize(c.Name, color.Red))
				valid.isValid = false
			}
		}
	}

	var isPollerCollectorExist bool
	// Verify poller collectors
	for k, v := range config.Pollers {
		for _, c := range v.Collectors {
			isPollerCollectorExist = true
			// Check if the collector name is valid
			_, ok := conf.IsCollector[c.Name]
			if !ok {
				fmt.Printf("Poller [%s] uses an invalid collector [%s] \n", color.Colorize(k, color.Yellow), color.Colorize(c.Name, color.Red))
				valid.isValid = false
			}
		}
	}

	// if no collector is configured in default and poller
	if !isDefaultCollectorExist && !isPollerCollectorExist {
		fmt.Printf("%s: No collectors are defined. Nothing will be collected.\n", color.Colorize("Error", color.Red))
		valid.isValid = false
	}

	// Print the valid collector names if there are invalid collector names
	if !valid.isValid {
		fmt.Printf("Valid collector names %v \n", color.Colorize(conf.GetCollectorSlice(), color.Green))
	}

	return valid
}

func checkConfTemplates(confPaths []string) validation {
	valid := validation{isValid: true}

	for _, confDir := range confPaths {
		dir, err := os.ReadDir(confDir)
		if err != nil {
			fmt.Printf("unable to read directory=%s err=%s\n", confDir, err)
			continue
		}
		for _, f := range dir {
			if !f.IsDir() {
				continue
			}
			flavor := f.Name()
			custom := path.Join(confDir, flavor, "custom.yaml")
			if _, err := os.Stat(custom); errors.Is(err, os.ErrNotExist) {
				continue
			}
			template, err := collector.ImportTemplate(confPaths, "custom.yaml", flavor)
			if err != nil {
				valid.isValid = false
				valid.invalid = append(valid.invalid, fmt.Sprintf(`%s is empty or invalid err=%+v`, custom, err))
				continue
			}
			s := template.GetChildS("objects")
			if s == nil {
				valid.isValid = false
				msg := custom + ` should have a top-level "objects" key`
				valid.invalid = append(valid.invalid, msg)
				continue
			}
			if s.Children == nil {
				valid.isValid = false
				msg := custom + " objects section should be a map of object: path"
				valid.invalid = append(valid.invalid, msg)
			} else {
				for _, t := range s.Children {
					if len(t.Content) == 0 {
						valid.isValid = false
						msg := custom + " objects section should be a map of object: path"
						valid.invalid = append(valid.invalid, msg)
						continue
					}
					searchDir := path.Join(confDir, flavor)
					fileNames := strings.Split(t.GetContentS(), ",")
					for _, fileName := range fileNames {
						fileName = strings.TrimSpace(fileName) // Remove any leading or trailing spaces
						if !templateExists(searchDir, fileName) {
							valid.isValid = false
							msg := fmt.Sprintf(`%s references template file %q which does not exist in %s`,
								custom, fileName, path.Join(searchDir, "**"))
							valid.invalid = append(valid.invalid, msg)
							continue
						}
					}
				}
			}
		}
	}

	if len(valid.invalid) > 0 {
		fmt.Printf("%s: Problems found in custom.yaml files\n", color.Colorize("Error", color.Red))
		for _, s := range valid.invalid {
			fmt.Printf("  %s\n", s)
		}
	}
	return valid
}

func templateExists(searchDir string, templateName string) bool {
	// recursively search searchDir for a file named templateName
	found := false
	err := filepath.WalkDir(searchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, templateName) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		fmt.Printf("failed to walk dir=%s err=%v", searchDir, err)
	}
	return found
}

// checkExporterTypes validates that all exporters are of valid types
func checkExporterTypes(config conf.HarvestConfig) validation {
	if config.Exporters == nil {
		return validation{}
	}
	invalidTypes := make(map[string]string)
	for name, exporter := range config.Exporters {
		if exporter.Type == "" {
			continue
		}
		if exporter.Type == "Prometheus" || exporter.Type == "InfluxDB" {
			continue
		}
		invalidTypes[name] = exporter.Type
	}

	valid := validation{isValid: true}

	if len(invalidTypes) > 0 {
		valid.isValid = false
		fmt.Printf("%s Unknown Exporter types found\n", color.Colorize("Error:", color.Red))
		fmt.Println("These are probably misspellings or the wrong case.")
		fmt.Println("Exporter types must start with a capital letter.")
		fmt.Println("The following exporters are unknown:")
		for name, eType := range invalidTypes {
			valid.invalid = append(valid.invalid, eType)
			fmt.Printf("  exporter named: [%s] has unknown type: [%s]\n", color.Colorize(name, color.Red), color.Colorize(eType, color.Yellow))
		}
		fmt.Println()
	}
	return valid
}

func checkExportersExist(config conf.HarvestConfig) validation {
	if config.Exporters == nil {
		fmt.Printf("%s: No Exporters section defined. No metrics will be exported.\n", color.Colorize("Error", color.Red))
		return validation{}
	}
	valid := validation{isValid: true}

	return valid
}

// checkUniquePromPorts checks that all Prometheus exporters, which specify a port, do so uniquely.
// Embedded exporters can conflict with non-embedded exporters, but not with each other.
func checkUniquePromPorts(config conf.HarvestConfig) validation {
	if config.Exporters == nil {
		return validation{}
	}
	// Add all exporters that have a port to a
	// map of portNum -> list of names
	seen := make(map[int][]string)
	var embeddedExporters []conf.ExporterDef

	for name, exporter := range config.Exporters {
		// ignore configuration with both port and portrange defined. PortRange takes precedence
		if exporter.Port == nil || exporter.Type != "Prometheus" || exporter.PortRange != nil {
			continue
		}
		// Ignore embedded exporters
		if exporter.IsEmbedded {
			embeddedExporters = append(embeddedExporters, conf.ExporterDef{Exporter: exporter, Name: name})
			continue
		}
		previous := seen[*exporter.Port]
		previous = append(previous, name)
		seen[*exporter.Port] = previous
	}

	// Update PortRanges
	for name, exporter := range config.Exporters {
		if exporter.PortRange == nil || exporter.Type != "Prometheus" {
			continue
		}
		portRange := exporter.PortRange
		start := portRange.Min
		end := portRange.Max
		for i := start; i <= end; i++ {
			previous := seen[i]
			previous = append(previous, name)
			seen[i] = previous
		}
	}

	valid := validation{isValid: true}
	for _, exporterNames := range seen {
		if len(exporterNames) == 1 {
			continue
		}
		valid.isValid = false
		break
	}

	// Check that embedded exports do not conflict with each other
	embeddedPorts := make(map[int][]string)
	for _, embeddedExporter := range embeddedExporters {
		// Check if the embedded exporter has a port
		if embeddedExporter.Port == nil {
			continue
		}
		// Check if the port is unique
		previous := embeddedPorts[*embeddedExporter.Port]
		previous = append(previous, embeddedExporter.Name)
		embeddedPorts[*embeddedExporter.Port] = previous
		if len(previous) > 1 {
			valid.isValid = false
			break
		}
	}

	if !valid.isValid {
		fmt.Printf("%s: Exporter PromPort conflict\n", color.Colorize("Error", color.Red))
		fmt.Println("  Prometheus exporters must specify unique ports. Change the following exporters to use unique ports:")
		for port, exporterNames := range seen {
			if len(exporterNames) == 1 {
				continue
			}
			names := strings.Join(exporterNames, ", ")
			valid.invalid = append(valid.invalid, exporterNames...)
			fmt.Printf("  port: [%s] duplicate exporters: [%s]\n", color.Colorize(port, color.Red), color.Colorize(names, color.Yellow))
		}
		for port, exporterNames := range embeddedPorts {
			if len(exporterNames) == 1 {
				continue
			}
			pollerNames := make([]string, 0, len(exporterNames))
			for _, exporterName := range exporterNames {
				index := strings.LastIndex(exporterName, "-")
				if index == -1 {
					pollerNames = append(pollerNames, exporterName)
					continue
				}
				pollerNames = append(pollerNames, exporterName[:index])
			}
			names := strings.Join(pollerNames, ", ")
			valid.invalid = append(valid.invalid, exporterNames...)
			fmt.Printf("  port: [%s] duplicate embedded exporters for pollers: [%s]\n", color.Colorize(port, color.Red), color.Colorize(names, color.Yellow))
		}
		fmt.Println()
	}
	return valid
}

// checkPollersExportToUniquePromPorts checks that all pollers which export
// to a Prometheus exporter, do so to a unique promPort
func checkPollersExportToUniquePromPorts(config conf.HarvestConfig) validation {
	if config.Exporters == nil {
		return validation{}
	}

	// Add all exporters that have a port to a
	// map of portNum -> list of names
	seen := make(map[int][]string)
	for name, exporter := range config.Exporters {
		// ignore configuration with both port and portrange defined. PortRange takes precedence
		if exporter.Port == nil || exporter.Type != "Prometheus" || exporter.PortRange != nil {
			continue
		}
		previous := seen[*exporter.Port]
		previous = append(previous, name)
		seen[*exporter.Port] = previous
	}

	// Look for pollers that export to the same Prometheus exporter that is not a port range exporter
	pollerExportsTo := make(map[string][]string)

	for name, poller := range config.Pollers {
		if poller.Exporters == nil {
			continue
		}
		for _, exporterName := range poller.Exporters {
			exporter, ok := config.Exporters[exporterName]
			if !ok {
				continue
			}
			if exporter.Type != "Prometheus" || exporter.Port == nil || exporter.PortRange != nil {
				continue
			}
			pollerExportsTo[exporterName] = append(pollerExportsTo[exporterName], name)
		}
	}

	valid := validation{isValid: true}
	for _, pollerNames := range pollerExportsTo {
		if len(pollerNames) == 1 {
			continue
		}
		valid.isValid = false
		valid.invalid = append(valid.invalid, pollerNames...)
		break
	}

	if !valid.isValid {
		fmt.Printf("%s: Multiple pollers export to the same PromPort\n", color.Colorize("Error", color.Red))
		fmt.Println("  Each poller should export to a unique Prometheus exporter or use PortRange. Change the following pollers to use unique exporters:")
		for port, pollerNames := range pollerExportsTo {
			if len(pollerNames) == 1 {
				continue
			}
			names := strings.Join(pollerNames, ", ")
			fmt.Printf("  pollers [%s] export to the same static PrometheusExporter: [%s]\n", color.Colorize(names, color.Yellow), color.Colorize(port, color.Red))
		}
		fmt.Println()
	}
	return valid
}

func printRedactedConfig(aPath string, contents []byte) (ast.Node, error) {
	// do not read comments since they may contain sensitive information
	astFile, err := parser.ParseBytes(contents, 0)
	if err != nil {
		fmt.Printf("error reading config file=[%s] %+v\n", aPath, err)
		return nil, err
	}

	root := astFile.Docs[0].Body

	sanitize(root)
	return root, nil
}

// checkPollerPromPorts checks that
//   - pollers that define a prom_port do so uniquely.
//   - when a prom_port is defined, but there are no Prometheus exporters
func checkPollerPromPorts(config conf.HarvestConfig) validation {
	seen := make(map[int][]string)

	for _, pName := range config.PollersOrdered {
		poller := config.Pollers[pName]
		if poller.PromPort == 0 {
			continue
		}
		previous := seen[poller.PromPort]
		previous = append(previous, pName)
		seen[poller.PromPort] = previous
	}

	valid := validation{isValid: true}
	for _, pollerNames := range seen {
		if len(pollerNames) == 1 {
			continue
		}
		valid.isValid = false
		break
	}

	if !valid.isValid {
		fmt.Printf("%s: Multiple pollers use the same prom_port.\n", color.Colorize("Error", color.Red))
		fmt.Println("  Each poller's prom_port should be unique. Change the following pollers to use unique prom_ports:")

		for port, pollerNames := range seen {
			if len(pollerNames) == 1 {
				continue
			}
			names := strings.Join(pollerNames, ", ")
			fmt.Printf("  pollers [%s] specify the same prom_port: [%s]\n", color.Colorize(names, color.Yellow), color.Colorize(port, color.Red))
			valid.invalid = append(valid.invalid, names)
		}
		fmt.Println()
	}

	// Check if there are any pollers that define a prom_port but there are no Prometheus exporters
	if config.Exporters == nil {
		fmt.Printf("%s: No Exporters section defined. At least one Prometheus exporter is needed for prom_port to export.\n", color.Colorize("Error", color.Red))
		valid.invalid = append(valid.invalid, "No Prometheus exporters defined")
		valid.isValid = false
	} else {
		hasPromExporter := false
		for _, exporter := range config.Exporters {
			if exporter.Type == "Prometheus" {
				hasPromExporter = true
				break
			}
		}
		if !hasPromExporter {
			fmt.Printf("%s: No Prometheus exporters defined. At least one Prometheus exporter is needed for prom_port to export.\n", color.Colorize("Error", color.Red))
			valid.invalid = append(valid.invalid, "No Prometheus exporters defined")
			valid.isValid = false
		}
	}

	return valid
}

// Update this map when there are additional tokens to sanitize
var sanitizeWords = map[string]bool{
	"username":          true,
	"password":          true,
	"grafana_api_token": true,
	"token":             true,
	"host":              true,
	"addr":              true,
}

func collectMapNodes(n ast.Node, nodes *[]*ast.MappingValueNode) {
	if n == nil {
		return
	}

	if n.Type() == ast.MappingType {
		mappingNode := n.(*ast.MappingNode)
		for _, kv := range mappingNode.Values {
			if kv.Key.Type() == ast.StringType {
				*nodes = append(*nodes, kv)
			}
			collectMapNodes(kv.Value, nodes)
		}
	} else if n.Type() == ast.SequenceType {
		sequenceNode := n.(*ast.SequenceNode)
		for _, value := range sequenceNode.Values {
			collectMapNodes(value, nodes)
		}
	}
}

func sanitize(root ast.Node) {

	var nodes []*ast.MappingValueNode
	collectMapNodes(root, &nodes)

	for _, n := range nodes {
		aNode := *n
		if aNode.Key.Type() == ast.StringType {
			keyName := node.ToString(aNode.Key)
			_, ok := sanitizeWords[keyName]
			if ok {
				if aNode.Value.Type() == ast.StringType {
					// sanitize the value
					aNode.Value.(*ast.StringNode).Value = "-REDACTED-"
				}
			}
		}
	}
}

func init() {
	Cmd.AddCommand(mergeCmd)
	Cmd.AddCommand(compareZapiRestMetricsCmd)
	dFlags := compareZapiRestMetricsCmd.PersistentFlags()
	mFlags := mergeCmd.PersistentFlags()

	dFlags.StringVarP(&opts.zapiDataCenterName, "zapiDc", "", "Zapi", "Zapi Datacenter Name ")
	dFlags.StringVarP(&opts.restDataCenterName, "restDc", "", "Rest", "Rest Datacenter Name. ")
	dFlags.StringVarP(&opts.prometheusURL, "promUrl", "", "", "Prometheus URL ")

	_ = compareZapiRestMetricsCmd.MarkPersistentFlagRequired("promUrl")

	mFlags.StringVarP(&opts.BaseTemplate, "template", "", "", "Base template path ")
	mFlags.StringVarP(&opts.MergeTemplate, "with", "", "", "Extended file path. ")

	_ = mergeCmd.MarkPersistentFlagRequired("template")
	_ = mergeCmd.MarkPersistentFlagRequired("with")
	Cmd.Flags().BoolVarP(
		&opts.ShouldPrintConfig,
		"print",
		"p",
		false,
		"Print config to console with sensitive info redacted",
	)

	Cmd.Flags().StringVar(&opts.Color, "color", "auto", "When to use colors. One of: auto | always | never. Auto will guess based on tty.")
	Cmd.Flags().BoolVar(&opts.expandVar, "expand-var", false, "Expand environment variables in config (default: false)")
}
