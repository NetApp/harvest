/*
Copyright NetApp Inc, 2021 All rights reserved

NetApp Harvest : the swiss-army-knife for datacenter monitoring

Authors:

	Chris Grindstaff
	Georg Mey
	Hardik Leuva
	Rahul Gupta
	Vachagan Gratian

Contact:

	ng-harvest-maintainers@netapp.com

This project is based on NetApp Harvest, authored by
Chris Madden in 2015.
*/
package main

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/admin"
	"github.com/netapp/harvest/v2/cmd/harvest/version"
	"github.com/netapp/harvest/v2/cmd/tools/doctor"
	"github.com/netapp/harvest/v2/cmd/tools/generate"
	"github.com/netapp/harvest/v2/cmd/tools/grafana"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/cmd/tools/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/util"
	tw "github.com/netapp/harvest/v2/third_party/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"log"
	"net"
	_ "net/http/pprof" // #nosec since pprof is off by default
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	maxCol           = 40
	defaultLogFormat = "plain"
)

type options struct {
	command    string
	collectors []string
	objects    []string
	verbose    bool
	trace      bool
	debug      bool
	foreground bool
	loglevel   int
	logToFile  bool   // only used when running in the foreground
	logFormat  string // one of plain or json
	config     string
	confPath   string
	profiling  bool
	longStatus bool
	daemon     bool
	promPort   int
}

var (
	HarvestHomePath   string
	HarvestConfigPath string
)

var rootCmd = &cobra.Command{
	Use:   "harvest <command> <subcommand> [flags]",
	Short: "NetApp Harvest - application for monitoring storage systems",
	Long: `NetApp Harvest - application for monitoring storage systems
`,
	Args: cobra.ArbitraryArgs,
}

var opts = &options{
	loglevel: 2,
	config:   HarvestConfigPath,
}

func doManageCmd(cmd *cobra.Command, args []string) {
	var (
		name                                             string
		pollerNames, pollersFromCmdLine, pollersFiltered []string
		pollerNamesSet                                   *set.Set
		ok, has                                          bool
		err                                              error
	)
	opts.command = cmd.Name()
	HarvestHomePath = conf.Path("")
	HarvestConfigPath = conf.Path(conf.HarvestYML)

	if opts.verbose {
		_ = cmd.Flags().Set("loglevel", "1")
	}
	if opts.trace {
		_ = cmd.Flags().Set("loglevel", "0")
	}

	_, err = conf.LoadHarvestConfig(opts.config)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("config [%s]: not found\n", opts.config)
		}
		log.Fatalf("config [%s]: %v\n", opts.config, err)
	}

	pollerNames = conf.Config.PollersOrdered
	pollers := conf.Config.Pollers

	// do this before filtering of pollers
	// stop pollers which may have been renamed or no longer exists in harvest.yml
	if opts.command == "start" || opts.command == "restart" {
		stopGhostPollers(pollerNames)
	}

	pollersFromCmdLine = args
	if len(pollersFromCmdLine) > 0 {
		// verify poller names
		ok = true
		for _, name = range pollersFromCmdLine {
			if _, has = pollers[name]; !has {
				fmt.Printf("poller [%s] not defined\n", name)
				ok = false
			}
		}
		if !ok {
			os.Exit(1)
		}
		// leave only requested pollers
		pollerNamesSet = set.NewFrom(pollersFromCmdLine)
		for name = range pollers {
			if pollerNamesSet.Has(name) {
				pollersFiltered = append(pollersFiltered, name)
			}
		}
	} else {
		// if no pollers in cmdline, use all pollers
		pollersFiltered = pollerNames
	}

	if opts.foreground {
		if opts.command != "start" {
			fmt.Printf("invalid command [%s] for foreground mode\n", opts.command)
			os.Exit(1)
		}
		if len(pollersFiltered) != 1 {
			fmt.Println("only one poller can be started in foreground mode")
			os.Exit(1)
		}
		name = pollersFiltered[0]
		startPoller(name, getPollerPrometheusPort(name, opts), opts)
		os.Exit(0)
	}

	statusesByName := getPollersStatus()
	switch opts.command {
	case "restart":
		restartPollers(pollersFiltered, statusesByName)
	case "stop", "kill":
		stopAllPollers(pollersFiltered, statusesByName)
	case "start":
		startAllPollers(pollersFiltered, statusesByName)
	}
	printTable(pollersFiltered, statusesByName)
}

func printTable(filteredPollers []string, statusesByName map[string][]*util.PollerStatus) {
	table := tw.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(false)
	if opts.longStatus {
		table.SetHeader([]string{"Datacenter", "Poller", "PID", "PromPort", "Profiling", "Status"})
	} else {
		table.SetHeader([]string{"Datacenter", "Poller", "PID", "PromPort", "Status"})
	}
	table.SetColumnAlignment([]int{tw.ALIGN_LEFT, tw.ALIGN_LEFT, tw.ALIGN_RIGHT, tw.ALIGN_RIGHT, tw.ALIGN_RIGHT})
	notRunning := &util.PollerStatus{Status: util.StatusNotRunning}
	disabled := &util.PollerStatus{Status: util.StatusDisabled}

	for _, name := range filteredPollers {
		var (
			poller       *conf.Poller
			pollerExists bool
		)
		if poller, pollerExists = conf.Config.Pollers[name]; !pollerExists {
			// should never happen, ignore since this was handled earlier
			continue
		}

		if statuses, ok := statusesByName[name]; ok {
			// print each status, annotate extra rows with a +
			for i, status := range statuses {
				if i > 0 {
					printStatus(table, opts.longStatus, poller.Datacenter, "+"+name, status)
				} else {
					printStatus(table, opts.longStatus, poller.Datacenter, name, status)
				}
			}
		} else {
			// poller not running
			if poller.IsDisabled {
				printStatus(table, opts.longStatus, poller.Datacenter, name, disabled)
			} else {
				printStatus(table, opts.longStatus, poller.Datacenter, name, notRunning)
			}
		}
	}
	table.Render()
}

// stop all pollers then start them instead of stop/starting each individually
func restartPollers(pollersFiltered []string, statusesByName map[string][]*util.PollerStatus) {
	stopAllPollers(pollersFiltered, statusesByName)
	startAllPollers(pollersFiltered, statusesByName)
}

func startAllPollers(pollersFiltered []string, statusesByName map[string][]*util.PollerStatus) {
	for _, name := range pollersFiltered {
		if statuses, wasRunning := statusesByName[name]; wasRunning {
			for _, ss := range statuses {
				if ss.Status == util.StatusRunning || ss.Status == util.StatusStoppingFailed {
					continue
				}
				promPort := getPollerPrometheusPort(name, opts)
				startPoller(name, promPort, opts)
			}
		} else {
			// poller not already running, just stopped, or disabled
			poller, _ := conf.PollerNamed(name)
			if poller == nil || poller.IsDisabled {
				continue
			}
			promPort := getPollerPrometheusPort(name, opts)
			startPoller(name, promPort, opts)
		}
	}
}

func stopAllPollers(pollersFiltered []string, statusesByName map[string][]*util.PollerStatus) {
	for _, name := range pollersFiltered {
		if statuses, isRunning := statusesByName[name]; isRunning {
			for _, s := range statuses {
				stopPoller(s)
			}
		}
	}
}

func getPollersStatus() map[string][]*util.PollerStatus {
	var statuses []util.PollerStatus
	statusesByName := map[string][]*util.PollerStatus{}

	statuses, err := util.GetPollerStatuses()
	if err != nil {
		fmt.Printf("Unable to GetPollerStatuses err: %+v\n", err)
		return nil
	}
	// create a map of status names
	for _, status := range statuses {
		statusesByName[status.Name] = append(statusesByName[status.Name], &status) // #nosec G601
	}
	return statusesByName
}

func stopGhostPollers(skipPoller []string) {
	statuses, err := util.GetPollerStatuses()
	if err != nil {
		fmt.Printf("Unable to get poller statatuses err=%v\n", err)
		return
	}
	for _, p := range statuses {
		// skip if this poller is defined in harvest config
		var skip bool
		for _, s := range skipPoller {
			if p.Name == s {
				skip = true
				break
			}
		}
		// if poller doesn't exist in harvest config
		if !skip {
			proc, err := os.FindProcess(p.Pid)
			if err != nil {
				fmt.Printf("process not found for pid %d %v \n", p.Pid, err)
				continue
			}
			// send terminate signal
			if err := proc.Signal(syscall.SIGTERM); err != nil {
				if os.IsPermission(err) {
					fmt.Printf("Insufficient privileges to terminate process %v \n", err)
				}
			}
		}
	}
}

func killPoller(ps *util.PollerStatus) {
	// exit if pid was not found
	if ps.Pid < 1 {
		return
	}

	// send kill signal
	proc, _ := os.FindProcess(ps.Pid)
	if err := proc.Kill(); err != nil {
		if strings.HasSuffix(err.Error(), "process already finished") {
			ps.Status = util.StatusAlreadyExited
		} else {
			fmt.Println("kill:", err)
			os.Exit(1)
		}
	} else {
		ps.Status = util.StatusKilled
	}
}

// Stop the poller if it's stoppable
func stopPoller(ps *util.PollerStatus) {
	// if we get no valid PID, assume process is not running
	if ps.Pid < 1 {
		return
	}

	proc, _ := os.FindProcess(ps.Pid)

	// send terminate signal
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if os.IsPermission(err) {
			fmt.Printf("Insufficient privileges to terminate process for pid=%d poller=%s\n", ps.Pid, ps.Name)
			ps.Status = util.StatusStoppingFailed
			return
		}
		fmt.Println(err)
		return
	}

	// give the poller a chance to clean up and exit
	for range 5 {
		if proc.Signal(syscall.Signal(0)) != nil {
			ps.Status = util.StatusStopped
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	// couldn't verify poller exited
	// just try to kill it and cleanup
	killPoller(ps)
}

func startPoller(pollerName string, promPort int, opts *options) {
	isDocker := os.Getenv("HARVEST_DOCKER") == "yes"
	argv := []string{
		filepath.Join(HarvestHomePath, "bin", "poller"),
		"--poller",
		pollerName,
		"--loglevel",
		strconv.Itoa(opts.loglevel),
	}

	if promPort != 0 {
		argv = append(argv, "--promPort", strconv.Itoa(promPort))
	}

	if opts.debug {
		argv = append(argv, "--debug")
	}

	if opts.config != HarvestConfigPath {
		argv = append(argv, "--config", opts.config)
	}

	if opts.confPath != conf.DefaultConfPath {
		argv = append(argv, "--confpath", opts.confPath)
	}

	if opts.logFormat != defaultLogFormat {
		argv = append(argv, "--logformat", opts.logFormat)
	}

	if opts.profiling {
		if opts.foreground {
			// Always pick the same port when profiling in foreground
			argv = append(argv, "--profiling", "6060")
		} else {
			if port, err := freePort(); err != nil {
				// No free port, log it and move on
				fmt.Println("profiling disabled due to no free ports")
			} else {
				argv = append(argv, "--profiling", strconv.Itoa(port))
			}
		}
	}

	if len(opts.collectors) > 0 {
		argv = append(argv, "--collectors", strings.Join(opts.collectors, ","))
	}

	if len(opts.objects) > 0 {
		argv = append(argv, "--objects", strings.Join(opts.objects, ","))
	}

	if opts.foreground {
		if opts.logToFile {
			argv = append(argv, "--logtofile")
		}
		cmd := exec.Command(argv[0], argv[1:]...) //nolint:gosec
		fmt.Println("starting in foreground, enter CTRL+C or close terminal to stop poller")
		_ = os.Stdout.Sync()
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	// special case if we are in container, don't daemonize
	if !isDocker {
		argv = append(argv, "--daemon")
	}

	// Set the Setsid attribute to true, which creates a new session for the child process
	// This effectively detaches the child process from the parent process
	cmd := exec.Command(argv[0], argv[1:]...) //nolint:gosec
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	// Redirect standard file descriptors to /dev/null
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		fmt.Println("Error opening /dev/null: ", err)
		os.Exit(1)
	}

	defer closeDevNull(devNull)

	if isDocker {
		cmd.Stdin = nil
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdin = devNull
		cmd.Stdout = devNull
		cmd.Stderr = devNull
	}

	// Start the poller process in the background
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		defer os.Exit(1)
	}
}

func closeDevNull(devNull *os.File) {
	if err := devNull.Close(); err != nil {
		fmt.Println("Error closing /dev/null: ", err)
	}
}

func printStatus(table *tw.Table, long bool, dc, pn string, ps *util.PollerStatus) {
	dct := truncate(dc)
	pnt := truncate(pn)
	var row []string
	if long {
		row = []string{dct, pnt, "", ps.PromPort, ps.ProfilingPort, string(ps.Status)}
	} else {
		row = []string{dct, pnt, "", ps.PromPort, string(ps.Status)}
	}
	if ps.Pid != 0 {
		row[2] = strconv.Itoa(ps.Pid)
	}
	table.Append(row)
}

func truncate(name string) string {
	if len(name) < maxCol {
		return name
	}
	return name[0:maxCol-3] + "→"
}

func freePort() (int, error) {
	// TODO add range support [min, max] and read the range from harvest.yml
	// Ask the kernel for a free open port
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	dial, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer closeDial(dial)
	return dial.Addr().(*net.TCPAddr).Port, nil
}

func closeDial(dial *net.TCPListener) {
	_ = dial.Close()
}

func getPollerPrometheusPort(pollerName string, opts *options) int {
	var promPort int
	var err error

	// check first if poller argument has promPort defined
	// else in exporter config of poller
	if opts.promPort != 0 {
		return opts.promPort
	}

	if promPort, err = conf.GetLastPromPort(pollerName, false); err != nil {
		fmt.Println(err)
		return 0
	}
	return promPort
}

func init() {
	startCmd := manageCmd("start", false)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(manageCmd("status", true))
	rootCmd.AddCommand(manageCmd("stop", true))
	rootCmd.AddCommand(manageCmd("restart", true))
	rootCmd.AddCommand(manageCmd("kill", true))
	rootCmd.AddCommand(zapi.Cmd, rest.Cmd, grafana.Cmd)
	rootCmd.AddCommand(generate.Cmd)
	rootCmd.AddCommand(doctor.Cmd)
	rootCmd.AddCommand(version.Cmd())
	rootCmd.AddCommand(admin.Cmd())

	rootCmd.PersistentFlags().StringVar(&opts.config, "config", "./harvest.yml", "Harvest config file path")
	rootCmd.PersistentFlags().StringVar(&opts.confPath, "confpath", "conf", "colon-separated paths to search for Harvest templates")

	rootCmd.Version = version.String()
	rootCmd.SetVersionTemplate(version.String())
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + `
Feedback
  Open issues at https://github.com/NetApp/harvest
`)

	start := startCmd.PersistentFlags()
	start.BoolVarP(&opts.debug, "debug", "d", false, "enable debug logging (same as -loglevel 1). If both debug and loglevel are specified, loglevel wins")
	start.BoolVarP(&opts.verbose, "verbose", "v", false, "verbose logging (loglevel=1)")
	start.BoolVarP(&opts.trace, "trace", "t", false, "trace logging (loglevel=0)")
	start.BoolVarP(&opts.foreground, "foreground", "f", false, "start single poller in foreground")
	start.BoolVar(&opts.daemon, "daemon", true, "start poller in background")
	start.IntVarP(&opts.loglevel, "loglevel", "l", 2, "logging level (0=trace, 1=debug, 2=info, 3=warning, 4=error, 5=critical)")
	start.BoolVar(&opts.logToFile, "logtofile", false, "when running in the foreground, log to file instead of stdout")
	start.StringVar(&opts.logFormat, "logformat", defaultLogFormat, "log format (plain or json)")
	start.BoolVar(&opts.profiling, "profiling", false, "if profiling port > 0, enables profiling via localhost:PORT/debug/pprof/")
	start.IntVar(&opts.promPort, "promPort", 0, "prometheus port to use for HTTP endpoint")
	start.StringSliceVarP(&opts.collectors, "collectors", "c", []string{}, "only start these collectors (overrides harvest.yml)")
	start.StringSliceVarP(&opts.objects, "objects", "o", []string{}, "only start these objects (overrides collector config)")

	_ = start.MarkHidden("logtofile")
	_ = start.MarkHidden("verbose")
	_ = start.MarkHidden("trace")
}

// The management commands: start|status|stop|restart|kill
// are created with this function - all but start are hidden
// to save space
func manageCmd(use string, shouldHide bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:    use + " [POLLER...]",
		Short:  "Stop/restart/status/kill - all or individual pollers",
		Long:   "Harvest Manager - manage your pollers",
		Args:   cobra.ArbitraryArgs,
		Hidden: shouldHide,
		Run:    doManageCmd,
	}
	cmd.PersistentFlags().BoolVar(&opts.longStatus, "long", false, "show advanced status options")
	return cmd
}

func main() {
	// Prefer our order to alphabetical
	cobra.EnableCommandSorting = false
	cobra.CheckErr(rootCmd.Execute())
}
