package generate

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
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
	CertDir       string
}

type AdminInfo struct {
	ServiceName   string
	Port          int
	ConfigFile    string
	Image         string
	ContainerName string
	Enabled       bool
	CertDir       string
}

type PollerTemplate struct {
	Pollers []PollerInfo
	Admin   AdminInfo
}

type PromTemplate struct {
	GrafanaPort int
	PromPort    int
}

type options struct {
	Poller      string
	loglevel    int
	image       string
	filesdPath  string
	showPorts   bool
	outputPath  string
	templateDir string
	certDir     string
	promPort    int
	grafanaPort int
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

var metricCmd = &cobra.Command{
	Use:    "metrics",
	Short:  "generate Harvest metrics documentation",
	Hidden: true,
	Run:    doGenerateMetrics,
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

func doGenerateMetrics(cmd *cobra.Command, _ []string) {
	var config = cmd.Root().PersistentFlags().Lookup("config")
	generateMetrics(conf.ConfigPath(config.Value.String()))
}

const (
	full                = 0
	harvest             = 1
	harvestAdminService = "harvest-admin.service"
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
	promTemplate := PromTemplate{
		opts.grafanaPort,
		opts.promPort,
	}
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
	certDirPath, err := filepath.Abs(opts.certDir)
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
			ContainerName: normalizeContainerNames("poller_" + v),
			ShowPorts:     kind == harvest || opts.showPorts,
			IsFull:        kind == full,
			TemplateDir:   templateDirPath,
			CertDir:       certDirPath,
		}
		pollerTemplate.Pollers = append(pollerTemplate.Pollers, pollerInfo)
		filesd = append(filesd, fmt.Sprintf("- targets: ['%s:%d']", pollerInfo.ServiceName, pollerInfo.Port))
	}

	t, err := template.New("docker-compose.tmpl").ParseFiles("docker/onePollerPerContainer/docker-compose.tmpl")
	if err != nil {
		panic(err)
	}

	var out *os.File
	color.DetectConsole("")
	out, err = os.Create(opts.outputPath)
	if err != nil {
		panic(err)
	}

	if kind == harvest {
		// generate admin service if configuration is present in harvest config
		if conf.Config.Admin.Httpsd.Listen != "" {
			httpsd := conf.Config.Admin.Httpsd.Listen

			adminPort := 8887
			if s := strings.Split(httpsd, ":"); len(s) == 2 {
				adminPort, err = strconv.Atoi(s[1])
				if err != nil {
					panic("Invalid httpsd listen configuration. Valid configuration are <<addr>>:PORT or :PORT")
				}
			} else {
				panic("Invalid httpsd listen configuration. Valid configuration are <<addr>>:PORT or :PORT")
			}

			pollerTemplate.Admin = AdminInfo{
				ServiceName:   "admin",
				ConfigFile:    configFilePath,
				Port:          adminPort,
				Image:         opts.image,
				ContainerName: "admin",
				Enabled:       true,
				CertDir:       certDirPath,
			}
		}
	} else {
		pt, err := template.New("prom-stack.tmpl").ParseFiles("prom-stack.tmpl")
		if err != nil {
			panic(err)
		}

		promStackOut, err := os.Create("prom-stack.yml")
		if err != nil {
			panic(err)
		}
		err = pt.Execute(promStackOut, promTemplate)
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

	if kind == harvest {
		_, _ = fmt.Fprintf(os.Stderr,
			"Start containers with:\n"+
				color.Colorize("docker-compose -f "+opts.outputPath+" up -d --remove-orphans\n", color.Green))
	}
	if kind == full {
		_, _ = fmt.Fprintf(os.Stderr,
			"Start containers with:\n"+
				color.Colorize("docker-compose -f prom-stack.yml -f "+opts.outputPath+" up -d --remove-orphans\n", color.Green))
	}
}

func silentClose(body io.ReadCloser) {
	_ = body.Close()
}

func generateSystemd(path string) {
	var adminService string
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
	if conf.Config.Admin.Httpsd.Listen != "" {
		adminService = harvestAdminService + " "
		println("and " + color.Colorize("cp "+harvestAdminService+" /etc/systemd/system/", color.Green))
	}
	println("and then run " + color.Colorize("systemctl daemon-reload", color.Green))
	writeAdminSystemd(path)
	// reorder list of pollers so that unix collectors are last, see https://github.com/NetApp/harvest/issues/643
	pollers := make([]string, 0)
	unixPollers := make([]string, 0)
	pollers = append(pollers, conf.Config.PollersOrdered...)
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
	pollers = append(pollers, unixPollers...)
	err = t.Execute(os.Stdout, struct {
		Admin          string
		PollersOrdered []string
	}{
		Admin:          adminService,
		PollersOrdered: pollers,
	})
	if err != nil {
		panic(err)
	}
}

func writeAdminSystemd(configFp string) {
	if conf.Config.Admin.Httpsd.Listen == "" {
		return
	}
	t, err := template.New("httpsd.tmpl").ParseFiles("service/contrib/httpsd.tmpl")
	if err != nil {
		panic(err)
	}
	f, err := os.Create(harvestAdminService)
	if err != nil {
		panic(err)
	}
	defer silentClose(f)
	configAbsPath, err := filepath.Abs(configFp)
	if err != nil {
		configAbsPath = "/opt/harvest/harvest.yml"
	}
	_ = t.Execute(f, configAbsPath)
	println(color.Colorize("✓", color.Green) + " HTTP SD file: " + harvestAdminService + " created")
}

func generateMetrics(path string) {
	var (
		poller     *conf.Poller
		err        error
		restClient *rest.Client
		zapiClient *zapi.Client
	)

	err = conf.LoadHarvestConfig(path)
	if err != nil {
		panic(err)
	}

	if poller, _, err = rest.GetPollerAndAddr(opts.Poller); err != nil {
		return
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if restClient, err = rest.New(poller, timeout); err != nil {
		fmt.Printf("error creating new client %+v\n", err)
		os.Exit(1)
	}
	if err = restClient.Init(2); err != nil {
		fmt.Printf("error init rest client %+v\n", err)
		os.Exit(1)
	}

	if zapiClient, err = zapi.New(poller); err != nil {
		fmt.Printf("error creating new client %+v\n", err)
		os.Exit(1)
	}

	swaggerBytes = readSwaggerJSON()
	restCounters := processRestCounters(restClient)
	zapiCounters := processZapiCounters(zapiClient)
	counters := mergeCounters(restCounters, zapiCounters)
	counters = processExternalCounters(counters)
	generateCounterTemplate(counters, restClient)
}

func init() {
	Cmd.AddCommand(systemdCmd)
	Cmd.AddCommand(metricCmd)
	Cmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(fullCmd)

	dFlags := dockerCmd.PersistentFlags()
	fFlags := fullCmd.PersistentFlags()

	flags := metricCmd.PersistentFlags()
	flags.StringVarP(&opts.Poller, "poller", "p", "sar", "name of poller, e.g. 10.193.48.154")
	dFlags.IntVarP(&opts.loglevel, "loglevel", "l", 2,
		"logging level (0=trace, 1=debug, 2=info, 3=warning, 4=error, 5=critical)",
	)
	dFlags.StringVar(&opts.image, "image", "cr.netapp.io/harvest:latest", "Harvest image. Use rahulguptajss/harvest:latest to pull from Docker Hub")
	dFlags.StringVar(&opts.templateDir, "templatedir", "./conf", "Harvest template dir path")
	dFlags.StringVar(&opts.certDir, "certdir", "./cert", "Harvest certificate dir path")
	dFlags.StringVarP(&opts.outputPath, "output", "o", "", "Output file path. ")

	dFlags.BoolVarP(&opts.showPorts, "port", "p", false, "Expose poller ports to host machine")
	_ = dockerCmd.MarkPersistentFlagRequired("output")
	fFlags.StringVar(&opts.filesdPath, "filesdpath", "docker/prometheus/harvest_targets.yml",
		"Prometheus file_sd target path. Written when the --output is set")
	fFlags.IntVar(&opts.promPort, "promPort", 9090, "Prometheus Port")
	fFlags.IntVar(&opts.grafanaPort, "grafanaPort", 3000, "Grafana Port")
}
