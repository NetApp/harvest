package generate

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/tools"
	"github.com/netapp/harvest/v2/cmd/tools/grafana"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"github.com/netapp/harvest/v2/third_party/tidwall/sjson"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
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

var metricRe = regexp.MustCompile(`(\w+)\{`)

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
	// reset metricsPanelMap map
	metricsPanelMap := make(map[string]tools.PanelData)
	panelKeyMap = make(map[string]bool)

	visitDashboard(
		[]string{
			"grafana/dashboards/asar2",
			"grafana/dashboards/cmode",
			"grafana/dashboards/cmode-details",
			"grafana/dashboards/cisco",
			"grafana/dashboards/storagegrid",
			"grafana/dashboards/eseries",
		}, metricsPanelMap,
		func(data []byte, metricsPanelMap map[string]tools.PanelData) {
			visitExpressions(data, metricsPanelMap)
		})
	counters, cluster := tools.BuildMetrics("", "", opts.Poller, opts, metricsPanelMap)
	tools.GenerateOntapCounterTemplate(counters, cluster.Version)
	sgCounters, ciscoCounters, eseriesCounters := generateCounterTemplate(metricsPanelMap)
	generateMetadataFiles(counters, sgCounters, ciscoCounters, eseriesCounters)
}

func doDescription(cmd *cobra.Command, _ []string) {
	addRootOptions(cmd)
	counters, _ := tools.BuildMetrics("", "", opts.Poller, opts, make(map[string]tools.PanelData))
	grafana.VisitDashboards(
		[]string{"grafana/dashboards/cmode"},
		func(path string, data []byte) {
			generateDescription(path, data, counters)
		})
}

func addRootOptions(cmd *cobra.Command) {
	opts.ConfigPath = conf.ConfigPath(cmd.Root().PersistentFlags().Lookup("config").Value.String())
	opts.ConfPath = cmd.Root().PersistentFlags().Lookup("confpath").Value.String()
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
		out            *os.File
	)

	pollerTemplate = PollerTemplate{}
	promTemplate := PromTemplate{
		opts.GrafanaPort,
		opts.PromPort,
	}
	_, err := conf.LoadHarvestConfig(opts.ConfigPath)
	if err != nil {
		tools.LogErrAndExit(err)
	}
	configFilePath = asComposePath(opts.ConfigPath)
	certDirPath = asComposePath(opts.CertDir)
	filesd := make([]string, 0, len(conf.Config.PollersOrdered))

	for _, pollerName := range conf.Config.PollersOrdered {
		poller, ok := conf.Config.Pollers[pollerName]
		if !ok || poller == nil || poller.IsDisabled {
			continue
		}
		port, _ := conf.GetLastPromPort(pollerName, true)
		pollerInfo := PollerInfo{
			ServiceName:   normalizeContainerNames(pollerName),
			PollerName:    pollerName,
			ConfigFile:    configFilePath,
			Port:          port,
			LogLevel:      opts.Loglevel,
			Image:         opts.Image,
			ContainerName: normalizeContainerNames("poller_" + pollerName),
			ShowPorts:     opts.ShowPorts,
			IsFull:        kind == full,
			CertDir:       certDirPath,
			Mounts:        makeMounts(pollerName),
		}
		pollerTemplate.Pollers = append(pollerTemplate.Pollers, pollerInfo)
		filesd = append(filesd, fmt.Sprintf("- targets: ['%s:%d']", pollerInfo.ServiceName, pollerInfo.Port))
	}

	t, err := template.New("docker-compose.tmpl").ParseFiles("container/onePollerPerContainer/docker-compose.tmpl")
	if err != nil {
		tools.LogErrAndExit(err)
	}

	color.DetectConsole("")

	out, err = os.Create(opts.OutputPath)
	if err != nil {
		tools.LogErrAndExit(err)
	}
	defer out.Close()

	if kind == harvest {
		// generate admin service if configuration is present in harvest config
		if conf.Config.Admin.Httpsd.Listen != "" {
			httpsd := conf.Config.Admin.Httpsd.Listen

			adminPort := 8887
			if s := strings.Split(httpsd, ":"); len(s) == 2 {
				adminPort, err = strconv.Atoi(s[1])
				if err != nil {
					tools.LogErrAndExit(errors.New("invalid httpsd listen configuration. Valid configuration are <<addr>>:PORT or :PORT"))
				}
			} else {
				tools.LogErrAndExit(errors.New("invalid httpsd listen configuration. Valid configuration are <<addr>>:PORT or :PORT"))
			}

			pollerTemplate.Admin = AdminInfo{
				ServiceName:   "admin",
				ConfigFile:    configFilePath,
				Port:          adminPort,
				Image:         opts.Image,
				ContainerName: "admin",
				Enabled:       true,
				CertDir:       certDirPath,
			}
		}
	} else {
		pt, err := template.New("prom-stack.tmpl").ParseFiles("prom-stack.tmpl")
		if err != nil {
			tools.LogErrAndExit(err)
		}

		promStackOut, err := os.Create("prom-stack.yml")
		if err != nil {
			tools.LogErrAndExit(err)
		}
		defer promStackOut.Close()
		err = pt.Execute(promStackOut, promTemplate)
		if err != nil {
			tools.LogErrAndExit(err)
		}
	}

	err = t.Execute(out, pollerTemplate)
	if err != nil {
		tools.LogErrAndExit(err)
	}

	f, err := os.Create(opts.FilesdPath)
	if err != nil {
		tools.LogErrAndExit(err)
	}
	defer silentClose(f)
	for _, line := range filesd {
		_, _ = fmt.Fprintln(f, line)
	}
	_, _ = fmt.Fprintf(os.Stderr, "Wrote file_sd targets to %s\n", opts.FilesdPath)

	if os.Getenv("HARVEST_DOCKER") != "" {
		srcFolder := "/opt/harvest"
		destFolder := "/opt/temp"

		err = copyFiles(srcFolder, destFolder)
		if err != nil {
			tools.LogErrAndExit(err)
		}
	}

	if kind == harvest {
		_, _ = fmt.Fprint(os.Stderr,
			"Start containers with:\n"+
				color.Colorize("docker compose -f "+opts.OutputPath+" up -d --remove-orphans\n", color.Green))
	}
	if kind == full {
		_, _ = fmt.Fprint(os.Stderr,
			"Start containers with:\n"+
				color.Colorize("docker compose -f prom-stack.yml -f "+opts.OutputPath+" up -d --remove-orphans\n", color.Green))
	}
}

// setup mount(s) for the confpath and any CLI-passed mounts
func makeMounts(pollerName string) []string {
	var mounts = opts.Mounts

	p, err := conf.PollerNamed(pollerName)
	if err != nil {
		tools.LogErrAndExit(err)
	}

	confPath := opts.ConfPath
	if confPath == "conf" {
		confPath = p.ConfPath
	}

	if confPath == "" {
		mounts = append(mounts, toMount("./conf"))
	} else {
		paths := strings.SplitSeq(confPath, ":")
		for path := range paths {
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
		"container": 0755,
		"grafana":   0755,
	}
	// requires specific permissions
	filePermissionsInDir := map[string]os.FileMode{
		"container":  0644,
		"prometheus": 0644,
		"grafana":    0644,
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
			// Check if the current directory or any of its parent directories are in dirsPermissions
			dirCreated := false
			for dir, perm := range dirsPermissions {
				if strings.HasPrefix(relPath, dir) {
					err = os.MkdirAll(dest, perm)
					if err != nil {
						return err
					}
					dirCreated = true
					break
				}
			}
			if !dirCreated {
				err = os.MkdirAll(dest, 0750)
				if err != nil {
					return err
				}
			}
			err = changeOwner(dest)
			if err != nil {
				return err
			}

			return nil
		}

		// Skip excluded files
		if filesToExclude[info.Name()] {
			return nil
		}

		// Check if the file is under a directory in the filePermissions map
		for dir, perm := range filePermissionsInDir {
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
	if err != nil {
		return err
	}

	err = changeOwner(destPath)
	if err != nil {
		return err
	}

	return nil
}

func changeOwner(path string) error {
	// Get the UID and GID from the environment variables
	uidStr := os.Getenv("UID")
	gidStr := os.Getenv("GID")

	// If the UID and GID are set, change the owner and group of the file
	if uidStr != "" && gidStr != "" {
		uid, err := strconv.Atoi(uidStr)
		if err != nil {
			return err
		}
		gid, err := strconv.Atoi(gidStr)
		if err != nil {
			return err
		}
		err = os.Chown(path, uid, gid)
		if err != nil {
			return err
		}
	}

	return nil
}

func asComposePath(path string) string {
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "./") {
		return path
	}
	return "./" + path
}

func silentClose(body io.ReadCloser) {
	_ = body.Close()
}

func generateSystemd() {
	var adminService string
	_, err := conf.LoadHarvestConfig(opts.ConfigPath)
	if err != nil {
		tools.LogErrAndExit(err)
	}
	if conf.Config.Pollers == nil {
		return
	}
	t, err := template.New("target.tmpl").ParseFiles("service/contrib/target.tmpl")
	if err != nil {
		tools.LogErrAndExit(err)
	}
	color.DetectConsole("")
	println("Save the following to " + color.Colorize("/etc/systemd/system/harvest.target", color.Green) +
		" or " + color.Colorize("| sudo tee /etc/systemd/system/harvest.target", color.Green))
	if conf.Config.Admin.Httpsd.Listen != "" {
		adminService = harvestAdminService + " "
		println("and " + color.Colorize("cp "+harvestAdminService+" /etc/systemd/system/", color.Green))
	}
	println("and then run " + color.Colorize("systemctl daemon-reload", color.Green))
	writeAdminSystemd(opts.ConfigPath)
	pollers := make([]string, 0)
	unixPollers := make([]string, 0)

	for _, pollerName := range conf.Config.PollersOrdered {
		poller, ok := conf.Config.Pollers[pollerName]
		if !ok || poller == nil || poller.IsDisabled {
			continue
		}
		// reorder list of pollers so that unix collectors are last, see https://github.com/NetApp/harvest/issues/643
		// if unix is in the poller's list of collectors, remove it from the list of pollers
		skipPoller := false
		for _, c := range poller.Collectors {
			if c.Name == "Unix" {
				unixPollers = append(unixPollers, pollerName)
				skipPoller = true
			}
		}
		if !skipPoller {
			pollers = append(pollers, pollerName)
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
		tools.LogErrAndExit(err)
	}
}

func writeAdminSystemd(configFp string) {
	if conf.Config.Admin.Httpsd.Listen == "" {
		return
	}
	t, err := template.New("httpsd.tmpl").ParseFiles("service/contrib/httpsd.tmpl")
	if err != nil {
		tools.LogErrAndExit(err)
	}
	f, err := os.Create(harvestAdminService)
	if err != nil {
		tools.LogErrAndExit(err)
	}
	defer silentClose(f)
	configAbsPath, err := filepath.Abs(configFp)
	if err != nil {
		configAbsPath = "/opt/harvest/harvest.yml"
	}
	_ = t.Execute(f, configAbsPath)
	println(color.Colorize("✓", color.Green) + " HTTP SD file: " + harvestAdminService + " created")
}

func generateDescription(dPath string, data []byte, counters map[string]tools.Counter) {
	var err error
	dashPath := grafana.ShortPath(dPath)
	panelDescriptionMap := make(map[string]string)
	ignoreDashboards := []string{
		"cmode/health.json", "cmode/headroom.json",
	}
	if slices.Contains(ignoreDashboards, dashPath) {
		return
	}

	grafana.VisitAllPanels(data, func(path string, _, value gjson.Result) {
		kind := value.Get("type").ClonedString()
		if kind == "row" || kind == "text" {
			return
		}
		description := value.Get("description").ClonedString()
		targetsSlice := value.Get("targets").Array()

		if description == "" {
			if len(targetsSlice) == 1 {
				expr := targetsSlice[0].Get("expr").ClonedString()
				if !strings.Contains(expr, "/") && !strings.Contains(expr, "+") && !strings.Contains(expr, "-") && !strings.Contains(expr, "on") {
					allMatches := metricRe.FindAllStringSubmatch(expr, -1)
					for _, match := range allMatches {
						m := match[1]
						if m == "" {
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
	result := gjson.GetBytes(data, `@pretty:{"sortKeys":true, "indent":"  ", "width":0}`)

	if err = os.WriteFile(dPath, []byte(result.Raw), grafana.GPerm); err != nil {
		log.Fatalf("failed to write dashboard=%s err=%v\n", dPath, err)
	}
}

func generatePanelPathWithDescription(path string, desc string) (string, string) {
	if desc != "" && !strings.HasSuffix(desc, ".") {
		desc += "."
	}
	return strings.ReplaceAll(strings.ReplaceAll(path, "[", "."), "]", ".") + "description", desc
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
	flags.StringVarP(&opts.Poller, "poller", "p", "dc1", "name of poller, e.g. 10.193.48.154")
	_ = metricCmd.MarkPersistentFlagRequired("poller")

	flag := descCmd.PersistentFlags()
	flag.StringVarP(&opts.Poller, "poller", "p", "dc1", "name of poller, e.g. 10.193.48.154")
	_ = descCmd.MarkPersistentFlagRequired("poller")

	dFlags.IntVarP(&opts.Loglevel, "loglevel", "l", 2,
		"logging level (0=trace, 1=debug, 2=info, 3=warning, 4=error, 5=critical)",
	)
	dFlags.StringVar(&opts.Image, "image", "ghcr.io/netapp/harvest:latest", "Harvest image. Use rahulguptajss/harvest:latest to pull from Docker Hub")
	dFlags.StringVar(&opts.CertDir, "certdir", "./cert", "Harvest certificate dir path")
	dFlags.StringVarP(&opts.OutputPath, "output", "o", "", "Output file path. ")
	dFlags.BoolVarP(&opts.ShowPorts, "port", "p", true, "Expose poller ports to host machine")
	_ = dockerCmd.MarkPersistentFlagRequired("output")
	dFlags.StringSliceVar(&opts.Mounts, "volume", []string{}, "Additional volume mounts to include in compose file")

	fFlags.StringVar(&opts.FilesdPath, "filesdpath", "container/prometheus/harvest_targets.yml",
		"Prometheus file_sd target path. Written when the --output is set")
	fFlags.IntVar(&opts.PromPort, "promPort", 9090, "Prometheus Port")
	fFlags.IntVar(&opts.GrafanaPort, "grafanaPort", 3000, "Grafana Port")

	metricCmd.PersistentFlags().StringVar(&opts.PromURL, "prom-url", "", "Prometheus URL for CI validation")
}
