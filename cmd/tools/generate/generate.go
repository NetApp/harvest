package generate

import (
	"fmt"
	"github.com/spf13/cobra"
	"goharvest2/cmd/harvest/version"
	"goharvest2/pkg/color"
	"goharvest2/pkg/conf"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type PollerInfo struct {
	ServiceName   string
	PollerName    string
	Port          int
	ConfigFile    string
	LogLevel      int
	Image         string
	ContainerName string
	ShowPorts     bool
	IsFull        bool
	TemplateDir   string
}

type PollerTemplate struct {
	Pollers []PollerInfo
}

type options struct {
	loglevel    int
	image       string
	filesdPath  string
	showPorts   bool
	outputPath  string
	templateDir string
}

var opts = &options{
	loglevel: 2,
	image:    "harvest:latest",
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

var fullCmd = &cobra.Command{
	Use:   "full",
	Short: "generate Harvest, Grafana, Prometheus docker-compose.yml target for all pollers defined in config",
	Run:   doDockerFull,
}

func doDockerFull(cmd *cobra.Command, _ []string) {
	var config = cmd.Root().PersistentFlags().Lookup("config")
	generateFullCompose(conf.ConfigPath(config.Value.String()))
}
func doSystemd(cmd *cobra.Command, _ []string) {
	var config = cmd.Root().PersistentFlags().Lookup("config")
	generateSystemd(conf.ConfigPath(config.Value.String()))
}

func doDockerCompose(cmd *cobra.Command, _ []string) {
	var config = cmd.Root().PersistentFlags().Lookup("config")
	generateDockerCompose(conf.ConfigPath(config.Value.String()))
}

const (
	full    = 0
	harvest = 1
)

func generateFullCompose(path string) {
	generateDocker(path, full)
}

func generateDockerCompose(path string) {
	generateDocker(path, harvest)
}

func normalizeContainerNames(name string) string {
	re := regexp.MustCompile("[._]")
	return strings.ToLower(re.ReplaceAllString(name, "-"))
}

func generateDocker(path string, kind int) {
	pollerTemplate := PollerTemplate{}
	err := conf.LoadHarvestConfig(path)
	if err != nil {
		panic(err)
	}
	configFilePath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	templateDirPath, err := filepath.Abs(opts.templateDir)
	if err != nil {
		panic(err)
	}
	conf.ValidatePortInUse = true
	var filesd []string
	for _, v := range conf.Config.PollersOrdered {
		port, _ := conf.GetPrometheusExporterPorts(v)
		pollerInfo := PollerInfo{
			ServiceName:   normalizeContainerNames(v),
			PollerName:    v,
			ConfigFile:    configFilePath,
			Port:          port,
			LogLevel:      opts.loglevel,
			Image:         opts.image,
			ContainerName: normalizeContainerNames("poller_" + v + "_v" + version.VERSION),
			ShowPorts:     kind == harvest || opts.showPorts,
			IsFull:        kind == full,
			TemplateDir:   templateDirPath,
		}
		pollerTemplate.Pollers = append(pollerTemplate.Pollers, pollerInfo)
		filesd = append(filesd, fmt.Sprintf("- targets: ['%s:%d']", pollerInfo.ServiceName, pollerInfo.Port))
	}

	t, err := template.New("docker-compose.tmpl").ParseFiles("docker/onePollerPerContainer/docker-compose.tmpl")
	if err != nil {
		panic(err)
	}

	out := os.Stdout
	color.DetectConsole("")
	if kind == harvest {
		println("Save the following to " + color.Colorize("docker-compose.yml", color.Green) +
			" or " + color.Colorize("> docker-compose.yml", color.Green))
		println("and then run " + color.Colorize("docker-compose -f docker-compose.yml up -d --remove-orphans", color.Green))
	} else {
		out, err = os.Create(opts.outputPath)
		if err != nil {
			panic(err)
		}
	}
	err = t.Execute(out, pollerTemplate)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(opts.filesdPath)
	if err != nil {
		panic(err)
	}
	defer silentClose(f)
	for _, line := range filesd {
		_, _ = fmt.Fprintln(f, line)
	}
	_, _ = fmt.Fprintf(os.Stderr, "Wrote file_sd targets to %s\n", opts.filesdPath)
	if kind == full {
		_, _ = fmt.Fprintf(os.Stderr,
			"Start containers with:\n"+
				color.Colorize("docker-compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans\n", color.Green))
	}
}

func silentClose(body io.ReadCloser) {
	_ = body.Close()
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
	// reorder list of pollers so that unix collectors are last, see https://github.com/NetApp/harvest/issues/643
	pollers := make([]string, 0)
	unixPollers := make([]string, 0)
	for _, k := range conf.Config.PollersOrdered {
		pollers = append(pollers, k)
	}
	// iterate over the pollers backwards, so we don't skip any when removing
	for i := len(pollers) - 1; i >= 0; i-- {
		pollerName := pollers[i]
		poller, ok := conf.Config.Pollers[pollerName]
		if !ok || poller == nil {
			continue
		}
		// if unix is in the poller's list of collectors, remove it from the list of pollers
		for _, c := range poller.Collectors {
			if c.Name == "Unix" {
				pollers = append(pollers[:i], pollers[i+1:]...)
				unixPollers = append(unixPollers, pollerName)
				break
			}
		}
	}
	for _, poller := range unixPollers {
		pollers = append(pollers, poller)
	}
	err = t.Execute(os.Stdout, struct {
		PollersOrdered []string
	}{
		PollersOrdered: pollers,
	})
	if err != nil {
		panic(err)
	}
}

func init() {
	Cmd.AddCommand(systemdCmd)
	Cmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(fullCmd)

	dFlags := dockerCmd.PersistentFlags()
	fFlags := fullCmd.PersistentFlags()

	dFlags.IntVarP(&opts.loglevel, "loglevel", "l", 2,
		"logging level (0=trace, 1=debug, 2=info, 3=warning, 4=error, 5=critical)",
	)
	dFlags.StringVar(&opts.image, "image", "rahulguptajss/harvest:latest", "Harvest image")
	dFlags.StringVar(&opts.templateDir, "templatedir", "./conf", "Harvest template dir path")

	fFlags.BoolVarP(&opts.showPorts, "port", "p", false, "Expose poller ports to host machine")
	fFlags.StringVarP(&opts.outputPath, "output", "o", "", "Output file path. ")
	_ = fullCmd.MarkPersistentFlagRequired("output")
	fFlags.StringVar(&opts.filesdPath, "filesdpath", "docker/prometheus/harvest_targets.yml",
		"Prometheus file_sd target path. Written when the --output is set")
}
