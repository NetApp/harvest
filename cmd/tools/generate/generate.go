package generate

import (
	"github.com/spf13/cobra"
	"goharvest2/pkg/color"
	"goharvest2/pkg/conf"
	"os"
	"path/filepath"
	"text/template"
)

type PollerPort struct {
	PollerName string
	Port       int
	ConfigFile string
}

type PollerTemplate struct {
	Pollers []PollerPort
}

var Cmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Harvest related files",
	Long:  "Generate Harvest related files",
}

var systemdCmd = &cobra.Command{
	Use:   "systemd",
	Short: "generate Harvest systemd target for all pollers defined in config",
	Run:   doSystemd,
}

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "generate Harvest docker-compose.yml target for all pollers defined in config",
	Run:   doDockerCompose,
}

func doSystemd(cmd *cobra.Command, _ []string) {
	var config = cmd.Root().PersistentFlags().Lookup("config")
	generateSystemd(config.Value.String())
}

func doDockerCompose(cmd *cobra.Command, _ []string) {
	var config = cmd.Root().PersistentFlags().Lookup("config")
	generateDockerCompose(config.Value.String())
}

func generateDockerCompose(path string) {
	pollerTemplate := PollerTemplate{}
	err := conf.LoadHarvestConfig(path)
	if err != nil {
		return
	}
	if conf.Config.Pollers == nil {
		return
	}
	// fetch absolute path of file for binding to volume
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	conf.IsDocker = true
	for _, v := range conf.Config.PollersOrdered {
		port, _ := conf.GetPrometheusExporterPorts(v)
		pollerTemplate.Pollers = append(pollerTemplate.Pollers, PollerPort{v, port, absPath})
	}

	t, err := template.New("docker-compose.tmpl").ParseFiles("docker/onePollerPerContainer/docker-compose.tmpl")
	if err != nil {
		panic(err)
	}

	color.DetectConsole("")
	println("Save the following to " + color.Colorize("docker-compose.yml", color.Green))
	println("and then run " + color.Colorize("docker-compose -f docker-compose.yml up -d --remove-orphans", color.Green))

	err = t.Execute(os.Stdout, pollerTemplate)
	if err != nil {
		panic(err)
	}
}

func generateSystemd(path string) {
	err := conf.LoadHarvestConfig(path)
	if err != nil {
		return
	}
	if conf.Config.Pollers == nil {
		return
	}
	t, err := template.New("target.tmpl").ParseFiles("service/contrib/target.tmpl")
	if err != nil {
		panic(err)
	}
	color.DetectConsole("")
	println("Save the following to " + color.Colorize("/etc/systemd/system/harvest.target", color.Green) +
		" or " + color.Colorize("| sudo tee /etc/systemd/system/harvest.target", color.Green))
	println("and then run " + color.Colorize("systemctl daemon-reload", color.Green))
	err = t.Execute(os.Stdout, conf.Config)
	if err != nil {
		panic(err)
	}
}

func init() {
	Cmd.AddCommand(systemdCmd)
	Cmd.AddCommand(dockerCmd)
}
