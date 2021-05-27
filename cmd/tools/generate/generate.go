package generate

import (
	"github.com/spf13/cobra"
	"goharvest2/pkg/color"
	"goharvest2/pkg/conf"
	"os"
	"text/template"
)

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

func doSystemd(cmd *cobra.Command, _ []string) {
	var config = cmd.Root().PersistentFlags().Lookup("config")
	generateSystemd(config.Value.String())
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
}
