package doctor

import (
	"fmt"
	"github.com/spf13/cobra"
	"goharvest2/pkg/color"
	"goharvest2/pkg/conf"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strings"
)

type options struct {
	ShouldPrintConfig bool
	Color             string
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

func doDoctorCmd(cmd *cobra.Command, _ []string) {
	var config = cmd.Root().PersistentFlags().Lookup("config")
	doDoctor(config.Value.String())
}

func doDoctor(path string) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("error reading config file. err=%+v\n", err)
		return
	}
	if opts.ShouldPrintConfig {
		printRedactedConfig(path, contents)
	}
	checkAll(path, contents)
}

// checkAll runs all doctor checks
// If all checks succeed, print nothing and exit with a return code of 0
// Otherwise, print what failed and exit with a return code of 1
func checkAll(path string, contents []byte) {
	// See https://github.com/NetApp/harvest/issues/16 for more checks to add
	color.DetectConsole(opts.Color)
	// Validate that the config file can be parsed
	harvestConfig := &conf.HarvestConfig{}
	err := yaml.Unmarshal(contents, harvestConfig)
	if err != nil {
		fmt.Printf("error reading config file=[%s] %+v\n", path, err)
		os.Exit(1)
		return
	}

	anyFailed := false
	anyFailed = !checkUniquePromPorts(*harvestConfig).isValid || anyFailed
	anyFailed = !checkExporterTypes(*harvestConfig).isValid || anyFailed

	if anyFailed {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

// checkExporterTypes validates that all exporters are of valid types
func checkExporterTypes(config conf.HarvestConfig) validation {
	if config.Exporters == nil {
		return validation{}
	}
	invalidTypes := make(map[string]string)
	for name, exporter := range *config.Exporters {
		if exporter.Type == nil {
			continue
		}
		switch *exporter.Type {
		case "Prometheus", "InfluxDB":
			break
		default:
			invalidTypes[name] = *exporter.Type
		}
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

// checkUniquePromPorts checks that all Prometheus exporters
// that specify a port, do so uniquely
func checkUniquePromPorts(config conf.HarvestConfig) validation {
	if config.Exporters == nil {
		return validation{}
	}
	// Add all exporters that have a port to a
	// map of portNum -> list of names
	seen := make(map[int][]string)
	for name, exporter := range *config.Exporters {
		// ignore configuration with both port and portrange defined. PortRange takes precedence
		if exporter.Port == nil || exporter.Type == nil || *exporter.Type != "Prometheus" || exporter.PortRange != nil {
			continue
		}
		previous := seen[*exporter.Port]
		previous = append(previous, name)
		seen[*exporter.Port] = previous
	}

	// Update PortRanges
	for name, exporter := range *config.Exporters {
		if exporter.PortRange == nil || exporter.Type == nil || *exporter.Type != "Prometheus" {
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
		for _, name := range exporterNames {
			valid.invalid = append(valid.invalid, name)
		}
		break
	}

	if !valid.isValid {
		fmt.Printf("%s: Exporter PromPort conflict\n", color.Colorize("Error", color.Red))
		fmt.Println("  Prometheus exporters must specify unique ports. Change the following exporters to use unique ports:")
		for port, exporterNames := range seen {
			if len(exporterNames) == 1 {
				continue
			}
			names := strings.Join(exporterNames, ", ")
			fmt.Printf("  port: [%s] duplicateExporters: [%s]\n", color.Colorize(port, color.Red), color.Colorize(names, color.Yellow))
		}
		fmt.Println()
	}
	return valid
}

func printRedactedConfig(path string, contents []byte) string {
	root := &yaml.Node{}
	err := yaml.Unmarshal(contents, root)
	if err != nil {
		fmt.Printf("error reading config file=[%s] %+v\n", path, err)
		return ""
	}
	var nodes []*yaml.Node
	collectNodes(root, &nodes)
	sanitize(nodes)
	removeComments(root)

	marshaled, err := yaml.Marshal(root)
	if err != nil {
		fmt.Printf("error marshalling yaml sanitized from config file=[%s] %+v\n", path, err)
		return ""
	}
	result := string(marshaled)
	fmt.Println(result)
	return result
}

func sanitize(nodes []*yaml.Node) {
	// Update this list when there are additional tokens to sanitize
	sanitizeWords := []string{"username", "password", "grafana_api_token", "token",
		"host", "addr"}
	for i, node := range nodes {
		if node == nil {
			continue
		}
		if node.Kind == yaml.ScalarNode && node.ShortTag() == "!!str" {
			value := node.Value
			for _, word := range sanitizeWords {
				if value == word {
					if nodes[i-1].Value == "auth_style" {
						continue
					}
					nodes[i+1].SetString("-REDACTED-")
				}
			}
		}
		removeComments(node)
	}
}

func removeComments(node *yaml.Node) {
	// Strip all comments since they may contain sensitive information
	node.HeadComment = ""
	node.LineComment = ""
	node.FootComment = ""
}

func collectNodes(root *yaml.Node, nodes *[]*yaml.Node) {
	for _, node := range root.Content {
		*nodes = append(*nodes, node)
		collectNodes(node, nodes)
	}
}

func init() {
	Cmd.Flags().BoolVarP(
		&opts.ShouldPrintConfig,
		"print",
		"p",
		false,
		"print config to console with sensitive info redacted",
	)

	Cmd.Flags().StringVar(&opts.Color, "color", "auto", "When to use colors. One of: auto | always | never. Auto will guess based on tty.")
}
