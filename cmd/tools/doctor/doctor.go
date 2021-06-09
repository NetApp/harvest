package doctor

import (
	"fmt"
	"github.com/spf13/cobra"
	"goharvest2/pkg/conf"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type options struct {
	ShouldPrintConfig bool
}

var opts = &options{
	ShouldPrintConfig: false,
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

func checkAll(path string, contents []byte) {
	// TODO add checks here, see https://github.com/NetApp/harvest/issues/16
	// print nothing and exit with 0 when all checks pass

	// Validate that the config file can be parsed
	harvestConfig := &conf.HarvestConfig{}
	err := yaml.Unmarshal(contents, harvestConfig)
	if err != nil {
		fmt.Printf("error reading config file=[%s] %+v\n", path, err)
		os.Exit(1)
		return
	}
	os.Exit(0)
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
}
