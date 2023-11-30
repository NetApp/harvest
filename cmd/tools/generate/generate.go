package generate

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/tools/grafana"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/cmd/tools/utils"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/tidwall/sjson"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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
	CertDir       string
	Mounts        []string
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
	certDir     string
	promPort    int
	grafanaPort int
	mounts      []string
	configPath  string
	confPath    string
}

var metricRe = regexp.MustCompile(`(\w+)\{`)

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

var descCmd = &cobra.Command{
	Use:    "desc",
	Short:  "generate Description of panels",
	Hidden: true,
	Run:    doDescription,
}

func doDockerFull(cmd *cobra.Command, _ []string) {
	addRootOptions(cmd)
	generateDocker(full)
}

func doSystemd(cmd *cobra.Command, _ []string) {
	addRootOptions(cmd)
	generateSystemd()
}

func doDockerCompose(cmd *cobra.Command, _ []string) {
	addRootOptions(cmd)
	generateDocker(harvest)
}

func doGenerateMetrics(cmd *cobra.Command, _ []string) {
	addRootOptions(cmd)
	counters, cluster := generateMetrics()
	generateCounterTemplate(counters, cluster.Version)
}

func doDescription(cmd *cobra.Command, _ []string) {
	addRootOptions(cmd)
	counters, _ := generateMetrics()
	grafana.VisitDashboards(
		[]string{"grafana/dashboards/cmode"},
		func(path string, data []byte) {
			generateDescription(path, data, counters)
		})
}

func addRootOptions(cmd *cobra.Command) {
	opts.configPath = conf.ConfigPath(cmd.Root().PersistentFlags().Lookup("config").Value.String())
	opts.confPath = cmd.Root().PersistentFlags().Lookup("confpath").Value.String()
}

const (
	full                = 0
	harvest             = 1
	harvestAdminService = "harvest-admin.service"
)

func normalizeContainerNames(name string) string {
	re := regexp.MustCompile("[._]")
	return strings.ToLower(re.ReplaceAllString(name, "-"))
}

func generateDocker(kind int) {
	var (
		pollerTemplate PollerTemplate
		configFilePath string
		certDirPath    string
		filesd         []string
		out            *os.File
	)

	pollerTemplate = PollerTemplate{}
	promTemplate := PromTemplate{
		opts.grafanaPort,
		opts.promPort,
	}
	_, err := conf.LoadHarvestConfig(opts.configPath)
	if err != nil {
		logErrAndExit(err)
	}
	configFilePath = asComposePath(opts.configPath)
	certDirPath = asComposePath(opts.certDir)

	for _, v := range conf.Config.PollersOrdered {
		port, _ := conf.GetPrometheusExporterPorts(v, true)
		pollerInfo := PollerInfo{
			ServiceName:   normalizeContainerNames(v),
			PollerName:    v,
			ConfigFile:    configFilePath,
			Port:          port,
			LogLevel:      opts.loglevel,
			Image:         opts.image,
			ContainerName: normalizeContainerNames("poller_" + v),
			ShowPorts:     opts.showPorts,
			IsFull:        kind == full,
			CertDir:       certDirPath,
			Mounts:        makeMounts(v),
		}
		pollerTemplate.Pollers = append(pollerTemplate.Pollers, pollerInfo)
		filesd = append(filesd, fmt.Sprintf("- targets: ['%s:%d']", pollerInfo.ServiceName, pollerInfo.Port))
	}

	t, err := template.New("docker-compose.tmpl").ParseFiles("container/onePollerPerContainer/docker-compose.tmpl")
	if err != nil {
		logErrAndExit(err)
	}

	color.DetectConsole("")
	out, err = os.Create(opts.outputPath)
	if err != nil {
		logErrAndExit(err)
	}

	if kind == harvest {
		// generate admin service if configuration is present in harvest config
		if conf.Config.Admin.Httpsd.Listen != "" {
			httpsd := conf.Config.Admin.Httpsd.Listen

			adminPort := 8887
			if s := strings.Split(httpsd, ":"); len(s) == 2 {
				adminPort, err = strconv.Atoi(s[1])
				if err != nil {
					logErrAndExit(fmt.Errorf("invalid httpsd listen configuration. Valid configuration are <<addr>>:PORT or :PORT"))
				}
			} else {
				logErrAndExit(fmt.Errorf("invalid httpsd listen configuration. Valid configuration are <<addr>>:PORT or :PORT"))
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
			logErrAndExit(err)
		}

		promStackOut, err := os.Create("prom-stack.yml")
		if err != nil {
			logErrAndExit(err)
		}
		err = pt.Execute(promStackOut, promTemplate)
		if err != nil {
			logErrAndExit(err)
		}
	}

	err = t.Execute(out, pollerTemplate)
	if err != nil {
		logErrAndExit(err)
	}

	f, err := os.Create(opts.filesdPath)
	if err != nil {
		logErrAndExit(err)
	}
	defer silentClose(f)
	for _, line := range filesd {
		_, _ = fmt.Fprintln(f, line)
	}
	_, _ = fmt.Fprintf(os.Stderr, "Wrote file_sd targets to %s\n", opts.filesdPath)

	if os.Getenv("HARVEST_DOCKER") != "" {
		srcFolder := "/opt/harvest"
		destFolder := "/opt/temp"

		err = copyFiles(srcFolder, destFolder)
		if err != nil {
			logErrAndExit(err)
		}
	}

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

// setup mount(s) for the confpath and any CLI-passed mounts
func makeMounts(pollerName string) []string {
	var mounts = opts.mounts

	p, err := conf.PollerNamed(pollerName)
	if err != nil {
		logErrAndExit(err)
	}

	confPath := opts.confPath
	if confPath == "conf" {
		confPath = p.ConfPath
	}

	if confPath == "" {
		mounts = append(mounts, toMount("./conf"))
	} else {
		paths := strings.Split(confPath, ":")
		for _, path := range paths {
			mounts = append(mounts, toMount(path))
		}
	}

	return mounts
}

func toMount(hostPath string) string {
	hostPath = asComposePath(hostPath)
	if strings.HasPrefix(hostPath, "./") {
		return hostPath + ":" + "/opt/harvest/" + hostPath[2:]
	}
	return hostPath + ":" + hostPath
}

func copyFiles(srcPath, destPath string) error {
	filesToExclude := map[string]bool{
		"harvest.yml":         true,
		"harvest.yml.example": true,
		"prom-stack.tmpl":     true,
	}
	dirsToExclude := map[string]bool{
		"bin":                   true,
		"autosupport":           true,
		"onePollerPerContainer": true,
	}
	// requires specific permissions
	dirsPermissions := map[string]os.FileMode{
		"container":  0755,
		"prometheus": 0755,
	}
	// requires specific permissions
	filePermissions := map[string]os.FileMode{
		"container":  0644,
		"prometheus": 0644,
		"grafana":    0640,
	}

	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Generate the destination path
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(destPath, relPath)

		if info.IsDir() {
			// Skip excluded directories
			if dirsToExclude[info.Name()] {
				return filepath.SkipDir
			}
			// Check if the directory is in the dirsPermissions map
			if perm, ok := dirsPermissions[info.Name()]; ok {
				return os.MkdirAll(dest, perm)
			}
			return os.MkdirAll(dest, 0750)
		}

		// Skip excluded files
		if filesToExclude[info.Name()] {
			return nil
		}

		// Check if the file is under a directory in the filePermissions map
		for dir, perm := range filePermissions {
			if strings.HasPrefix(relPath, dir) {
				return copyFile(path, dest, perm)
			}
		}
		return copyFile(path, dest, 0600)
	})
}

func copyFile(srcPath, destPath string, perm os.FileMode) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer silentClose(srcFile)

	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer silentClose(destFile)

	_, err = io.Copy(destFile, srcFile)
	return err
}

func asComposePath(path string) string {
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "./") {
		return path
	}
	return "./" + path
}

func logErrAndExit(err error) {
	fmt.Printf("%v\n", err)
	os.Exit(1)
}

func silentClose(body io.ReadCloser) {
	_ = body.Close()
}

func generateSystemd() {
	var adminService string
	_, err := conf.LoadHarvestConfig(opts.configPath)
	if err != nil {
		logErrAndExit(err)
	}
	if conf.Config.Pollers == nil {
		return
	}
	t, err := template.New("target.tmpl").ParseFiles("service/contrib/target.tmpl")
	if err != nil {
		logErrAndExit(err)
	}
	color.DetectConsole("")
	println("Save the following to " + color.Colorize("/etc/systemd/system/harvest.target", color.Green) +
		" or " + color.Colorize("| sudo tee /etc/systemd/system/harvest.target", color.Green))
	if conf.Config.Admin.Httpsd.Listen != "" {
		adminService = harvestAdminService + " "
		println("and " + color.Colorize("cp "+harvestAdminService+" /etc/systemd/system/", color.Green))
	}
	println("and then run " + color.Colorize("systemctl daemon-reload", color.Green))
	writeAdminSystemd(opts.configPath)
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
		logErrAndExit(err)
	}
}

func writeAdminSystemd(configFp string) {
	if conf.Config.Admin.Httpsd.Listen == "" {
		return
	}
	t, err := template.New("httpsd.tmpl").ParseFiles("service/contrib/httpsd.tmpl")
	if err != nil {
		logErrAndExit(err)
	}
	f, err := os.Create(harvestAdminService)
	if err != nil {
		logErrAndExit(err)
	}
	defer silentClose(f)
	configAbsPath, err := filepath.Abs(configFp)
	if err != nil {
		configAbsPath = "/opt/harvest/harvest.yml"
	}
	_ = t.Execute(f, configAbsPath)
	println(color.Colorize("✓", color.Green) + " HTTP SD file: " + harvestAdminService + " created")
}

func generateMetrics() (map[string]utils.Counter, rest.Cluster) {
	var (
		poller     *conf.Poller
		err        error
		restClient *rest.Client
		zapiClient *zapi.Client
	)

	_, err = conf.LoadHarvestConfig(opts.configPath)
	if err != nil {
		logErrAndExit(err)
	}

	if poller, _, err = rest.GetPollerAndAddr(opts.Poller); err != nil {
		logErrAndExit(err)
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	credentials := auth.NewCredentials(poller, logging.Get())
	if restClient, err = rest.New(poller, timeout, credentials); err != nil {
		fmt.Printf("error creating new client %+v\n", err)
		os.Exit(1)
	}
	if err = restClient.Init(2); err != nil {
		fmt.Printf("error init rest client %+v\n", err)
		os.Exit(1)
	}

	if zapiClient, err = zapi.New(poller, credentials); err != nil {
		fmt.Printf("error creating new client %+v\n", err)
		os.Exit(1)
	}

	utils.SwaggerBytes = readSwaggerJSON()
	restCounters := utils.ProcessRestCounters(restClient, nil, "conf/rest", "conf/restperf")
	zapiCounters := utils.ProcessZapiCounters(zapiClient, nil, "conf/zapi/cdot", "conf/zapiperf/cdot")
	counters := mergeCounters(restCounters, zapiCounters)
	counters = processExternalCounters(counters)
	return counters, restClient.Cluster()
}

func generateDescription(dPath string, data []byte, counters map[string]utils.Counter) {
	var err error
	dashPath := grafana.ShortPath(dPath)
	panelDescriptionMap := make(map[string]string)
	ignoreDashboards := []string{
		"cmode/health.json", "cmode/headroom.json",
	}
	if slices.Contains(ignoreDashboards, dashPath) {
		return
	}

	grafana.VisitAllPanels(data, func(path string, key, value gjson.Result) {
		kind := value.Get("type").String()
		if kind == "row" || kind == "text" {
			return
		}
		description := value.Get("description").String()
		targetsSlice := value.Get("targets").Array()

		if description == "" {
			if len(targetsSlice) == 1 {
				expr := targetsSlice[0].Get("expr").String()
				if !(strings.Contains(expr, "/") || strings.Contains(expr, "+") || strings.Contains(expr, "-") || strings.Contains(expr, "on")) {
					allMatches := metricRe.FindAllStringSubmatch(expr, -1)
					for _, match := range allMatches {
						m := match[1]
						if len(m) == 0 {
							continue
						}
						expr = m
					}
					panelPath, updatedDescription := generatePanelPathWithDescription(path, counters[expr].Description)
					panelDescriptionMap[panelPath] = updatedDescription
				}
			}
		} else if !strings.HasPrefix(description, "$") && !strings.HasSuffix(description, ".") {
			// Few panels have description text from variable, which would be ignored.
			panelPath, updatedDescription := generatePanelPathWithDescription(path, description)
			panelDescriptionMap[panelPath] = updatedDescription
		}
	})

	// Update the dashboard with description
	for path, value := range panelDescriptionMap {
		data, err = sjson.SetBytes(data, path, value)
		if err != nil {
			log.Fatalf("error while updating the panel in dashboard %s err: %+v", dPath, err)
		}
	}

	// Sorted json
	sorted := pretty.PrettyOptions(data, &pretty.Options{
		SortKeys: true,
		Indent:   "  ",
	})

	if err = os.WriteFile(dPath, sorted, grafana.GPerm); err != nil {
		log.Fatalf("failed to write dashboard=%s err=%v\n", dPath, err)
	}
}

func generatePanelPathWithDescription(path string, desc string) (string, string) {
	if desc != "" && !strings.HasSuffix(desc, ".") {
		desc = desc + "."
	}
	return strings.Replace(strings.Replace(path, "[", ".", -1), "]", ".", -1) + "description", desc
}

func init() {
	Cmd.AddCommand(systemdCmd)
	Cmd.AddCommand(metricCmd)
	Cmd.AddCommand(descCmd)
	Cmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(fullCmd)

	dFlags := dockerCmd.PersistentFlags()
	fFlags := fullCmd.PersistentFlags()

	flags := metricCmd.PersistentFlags()
	flags.StringVarP(&opts.Poller, "poller", "p", "sar", "name of poller, e.g. 10.193.48.154")
	_ = metricCmd.MarkPersistentFlagRequired("poller")

	flag := descCmd.PersistentFlags()
	flag.StringVarP(&opts.Poller, "poller", "p", "sar", "name of poller, e.g. 10.193.48.154")
	_ = descCmd.MarkPersistentFlagRequired("poller")

	dFlags.IntVarP(&opts.loglevel, "loglevel", "l", 2,
		"logging level (0=trace, 1=debug, 2=info, 3=warning, 4=error, 5=critical)",
	)
	dFlags.StringVar(&opts.image, "image", "ghcr.io/netapp/harvest:latest", "Harvest image. Use rahulguptajss/harvest:latest to pull from Docker Hub")
	dFlags.StringVar(&opts.certDir, "certdir", "./cert", "Harvest certificate dir path")
	dFlags.StringVarP(&opts.outputPath, "output", "o", "", "Output file path. ")
	dFlags.BoolVarP(&opts.showPorts, "port", "p", true, "Expose poller ports to host machine")
	_ = dockerCmd.MarkPersistentFlagRequired("output")
	dFlags.StringSliceVar(&opts.mounts, "volume", []string{}, "Additional volume mounts to include in compose file")

	fFlags.StringVar(&opts.filesdPath, "filesdpath", "container/prometheus/harvest_targets.yml",
		"Prometheus file_sd target path. Written when the --output is set")
	fFlags.IntVar(&opts.promPort, "promPort", 9090, "Prometheus Port")
	fFlags.IntVar(&opts.grafanaPort, "grafanaPort", 3000, "Grafana Port")
}
