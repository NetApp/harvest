/*
 * Copyright NetApp Inc, 2021 All rights reserved

NetApp Harvest : the swiss-army-knife for datacenter monitoring

Authors:
   Georg Mey & Vachagan Gratian
Contact:
   ng-harvest-maintainers@netapp.com

This project is based on NetApp Harvest, authored by
Chris Madden in 2015.

*/
package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"goharvest2/cmd/harvest/config"
	"goharvest2/cmd/harvest/stub"
	"goharvest2/cmd/harvest/version"
	"goharvest2/cmd/tools/doctor"
	"goharvest2/cmd/tools/generate"
	"goharvest2/cmd/tools/grafana"
	"goharvest2/cmd/tools/zapi"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/set"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"net"
	_ "net/http/pprof" // #nosec since pprof is off by default
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type options struct {
	command    string
	pollers    []string
	collectors []string
	objects    []string
	verbose    bool
	trace      bool
	debug      bool
	foreground bool
	loglevel   int
	config     string
	profiling  bool
	longStatus bool
	daemon     bool
	promPort   int
}

type pollerStatus struct {
	status        string
	pid           int
	profilingPort string
	promPort      string
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
	opts.command = cmd.Name()
	var err error
	HarvestHomePath = conf.GetHarvestHomePath()
	HarvestConfigPath, err = conf.GetDefaultHarvestConfigPath()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if opts.verbose {
		_ = cmd.Flags().Set("loglevel", "1")
	}
	if opts.trace {
		_ = cmd.Flags().Set("loglevel", "0")
	}

	//cmd.DebugFlags()  // uncomment to print flags

	pollers, err := conf.GetPollers(opts.config)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("config [%s]: not found\n", opts.config)
		} else {
			fmt.Printf("config [%s]: %v\n", opts.config, err)
		}
		os.Exit(1)
	}

	var pollerNames []string
	for _, p := range pollers.GetChildren() {
		pollerNames = append(pollerNames, p.GetNameS())
	}
	// do this before filtering of pollers
	// stop pollers which may have been renamed or no longer exists in harvest.yml
	if opts.command == "start" || opts.command == "restart" {
		stopGhostPollers("poller", pollerNames)
	}

	pollersFromCmdLine := args
	if len(pollersFromCmdLine) > 0 {
		// verify poller names
		ok := true
		for _, p := range pollersFromCmdLine {
			if pollers.GetChildS(p) == nil {
				fmt.Printf("poller [%s] not defined\n", p)
				ok = false
			}
		}
		if !ok {
			os.Exit(1)
		}
		// filter pollers
		pollerNames := set.NewFrom(pollersFromCmdLine)
		for _, p := range pollers.GetChildren() {
			if !pollerNames.Has(p.GetNameS()) {
				pollers.PopChildS(p.GetNameS())
			}
		}
	}

	if opts.foreground {
		if opts.command != "start" {
			fmt.Printf("invalid command [%s] for foreground mode\n", opts.command)
			os.Exit(1)
		}
		if len(pollers.GetChildren()) != 1 {
			fmt.Println("only one poller can be started in foreground mode")
			os.Exit(1)
		}
		p := pollers.GetChildren()[0]
		startPoller(p.GetNameS(), getPollerPrometheusPort(p, opts), opts)
		os.Exit(0)
	}

	// get max lengths of datacenter and poller names
	// so that output does not get distorted if they are too long
	c1, c2 := getMaxLengths(pollers, 20, 20)
	printHeader(opts.longStatus, c1, c2)
	printBreak(opts.longStatus, c1, c2)

	for _, p := range pollers.GetChildren() {

		var s *pollerStatus

		name := p.GetNameS()
		datacenter := p.GetChildContentS("datacenter")

		s = getStatus(name)
		if opts.command == "kill" {
			s = killPoller(name)
			printStatus(opts.longStatus, c1, c2, datacenter, name, s.promPort, s)
			continue
		}

		if opts.command == "status" {
			printStatus(opts.longStatus, c1, c2, datacenter, name, s.promPort, s)
		}

		if opts.command == "stop" || opts.command == "restart" {
			s = stopPoller(name)
			printStatus(opts.longStatus, c1, c2, datacenter, name, s.promPort, s)
		}

		if opts.command == "start" || opts.command == "restart" {
			// only start poller if confirmed that it's not running
			// if it's running do nothing
			switch s.status {
			case "running":
				// do nothing but print current status, idempotent
				printStatus(opts.longStatus, c1, c2, datacenter, name, s.promPort, s)
				break
			case "not running", "stopped", "killed":
				promPort := getPollerPrometheusPort(p, opts)
				s = startPoller(name, promPort, opts)
				printStatus(opts.longStatus, c1, c2, datacenter, name, s.promPort, s)
			default:
				fmt.Printf("can't verify status of [%s]: kill poller and try again\n", name)
			}
		}
	}

	printBreak(opts.longStatus, c1, c2)
}

// Get status of a poller. This is partially guesswork and
// there is no guarantee that this will always be correct.
// The general logic is:
// - use prep to find a matching poller
// - the check that the found pid has a harvest tag in its environ
//
// Returns:
//	@status - status of poller
//  @pid  - PID of poller (0 means no PID)
func getStatus(pollerName string) *pollerStatus {

	s := &pollerStatus{status: "not running"}

	// docker dummy status
	if os.Getenv("HARVEST_DOCKER") == "yes" {
		s.status = "n/a (docker)"
		return s
	}

	pids, err := util.GetPid(pollerName)
	if err != nil {
		return s
	}
	if len(pids) != 1 {
		if len(pids) > 1 {
			fmt.Printf("exepcted one pid for %s, instead pids=%+v\n", pollerName, pids)
		}
		return s
	}
	s.pid = pids[0]

	// no valid PID stops here
	if s.pid < 1 {
		return s
	}

	// retrieve process with PID and check if running
	// (error can be safely ignored on Unix)
	proc, _ := os.FindProcess(s.pid)

	// send signal 0 to check if process is running
	// returns err if it's not running or permission is denied
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		if os.IsPermission(err) {
			fmt.Println("Insufficient privileges to send signal to process")
		}
		return s
	}

	// process is running. GetPid ensures this is the correct poller
	// Extract cmdline args for status struct
	if cmdline, err := util.GetCmdLine(s.pid); err == nil {
		s.status = "running"

		if strings.Contains(cmdline, "--profiling") {
			r := regexp.MustCompile(`--profiling (\d+)`)
			matches := r.FindStringSubmatch(cmdline)
			if len(matches) > 0 {
				s.profilingPort = matches[1]
			}
		}

		if strings.Contains(cmdline, "--promPort") {
			r := regexp.MustCompile(`--promPort (\d+)`)
			matches := r.FindStringSubmatch(cmdline)
			if len(matches) > 0 {
				s.promPort = matches[1]
			}
		}
	}
	return s
}

func stopGhostPollers(search string, skipPoller []string) {
	pids, err := util.GetPids(search)
	if err != nil {
		fmt.Printf("Error while executing pgrep %v \n", err)
		return
	}
	for _, p := range pids {
		c, err := util.GetCmdLine(p)
		if err != nil {
			fmt.Printf("Missing pid %d %v \n", p, err)
			continue
		}

		// skip if this poller is defined in harvest config
		var skip bool
		for _, s := range skipPoller {
			if util.ContainsWholeWord(c, s) {
				skip = true
				break
			}
		}
		// if poller doesn't exists in harvest config
		if !skip {
			proc, err := os.FindProcess(p)
			if err != nil {
				fmt.Printf("process not found for pid %d %v \n", p, err)
				continue
			}
			// send terminate signal
			if err := proc.Signal(syscall.SIGTERM); err != nil {
				if os.IsPermission(err) {
					fmt.Printf("Insufficient priviliges to terminate process %v \n", err)
				}
			}
		}
	}
}

func killPoller(pollerName string) *pollerStatus {

	s := getStatus(pollerName)
	// exit if pid was not found
	if s.pid < 1 {
		return s
	}

	// send kill signal
	// TODO handle already exited
	proc, _ := os.FindProcess(s.pid)
	if err := proc.Kill(); err != nil {
		if strings.HasSuffix(err.Error(), "process already finished") {
			s.status = "already exited"
		} else {
			fmt.Println("kill:", err)
			os.Exit(1)
		}
	} else {
		s.status = "killed"
	}

	return s
}

// Stop poller if it's running or it's stoppable
//
// Returns: same as get_status()
func stopPoller(pollerName string) *pollerStatus {
	var s *pollerStatus
	// if we get no valid PID, assume process is not running
	if s = getStatus(pollerName); s.pid < 1 {
		return s
	}

	proc, _ := os.FindProcess(s.pid)

	// send terminate signal
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if os.IsPermission(err) {
			fmt.Println("Insufficient privileges to terminate process")
			s.status = "stopping failed"
			return s
		}
		fmt.Println(err)
		// other errors, probably mean poller already exited, so double check
		return getStatus(pollerName)
	}

	// give the poller chance to cleanup and exit
	for i := 0; i < 5; i += 1 {
		time.Sleep(50 * time.Millisecond)
		// @TODO, handle situation when PID is regained by some other process
		if proc.Signal(syscall.Signal(0)) != nil {
			s.status = "stopped"
			return s
		}
	}

	// couldn't verify poller exited
	// just try to kill it and cleanup
	return killPoller(pollerName)
}

func startPoller(pollerName string, promPort int, opts *options) *pollerStatus {

	argv := make([]string, 5)
	argv[0] = path.Join(HarvestHomePath, "bin", "poller")
	argv[1] = "--poller"
	argv[2] = pollerName
	argv[3] = "--loglevel"
	argv[4] = strconv.Itoa(opts.loglevel)

	if promPort != 0 {
		argv = append(argv, "--promPort")
		argv = append(argv, strconv.Itoa(promPort))
	}
	if opts.debug {
		argv = append(argv, "--debug")
	}

	if opts.config != HarvestConfigPath {
		argv = append(argv, "--config")
		argv = append(argv, opts.config)
	}

	if opts.profiling {
		if opts.foreground {
			// Always pick the same port when profiling in foreground
			argv = append(argv, "--profiling")
			argv = append(argv, "6060")
		} else {
			if port, err := freePort(); err != nil {
				// No free port, log it and move on
				fmt.Println("profiling disabled due to no free ports")
			} else {
				argv = append(argv, "--profiling")
				argv = append(argv, strconv.Itoa(port))
			}
		}
	}

	if len(opts.collectors) > 0 {
		argv = append(argv, "--collectors")
		argv = append(argv, opts.collectors...)
	}

	if len(opts.objects) > 0 {
		argv = append(argv, "--objects")
		argv = append(argv, opts.objects...)
	}

	if opts.foreground {
		cmd := exec.Command(argv[0], argv[1:]...)
		//fmt.Println(cmd.String())
		fmt.Println("starting in foreground, enter CTRL+C or close terminal to stop poller")
		_ = os.Stdout.Sync()
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		//fmt.Println("stopped")
		os.Exit(0)
	}

	argv = append(argv, "--daemon")

	// special case if we are in container, don't actually daemonize
	if os.Getenv("HARVEST_DOCKER") == "yes" {
		cmd := exec.Command(argv[0], argv[1:]...)
		if err := cmd.Start(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := cmd.Wait(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cmd := exec.Command(path.Join(HarvestHomePath, "bin", "daemonize"), argv...)
	//fmt.Println(cmd.String())
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Allow for some delay and retry checking status a few times
	for i := 0; i < 2; i += 1 {
		if s := getStatus(pollerName); s.pid > 0 {
			return s
		}
		time.Sleep(10 * time.Millisecond)
	}

	return getStatus(pollerName)
}

// print status of poller, first two arguments are column lengths
func printStatus(long bool, c1, c2 int, dc, pn, promPort string, s *pollerStatus) {
	fmt.Printf("%s%s ", dc, strings.Repeat(" ", c1-len(dc)))
	fmt.Printf("%s%s ", pn, strings.Repeat(" ", c2-len(pn)))
	if long {
		if s.pid == 0 {
			fmt.Printf("%-10s %-15s %-10s %-20s\n", "", promPort, s.profilingPort, s.status)
		} else {
			fmt.Printf("%-10d %-15s %-10s %-20s\n", s.pid, promPort, s.profilingPort, s.status)
		}
	} else if s.pid == 0 {
		fmt.Printf("%-10s %-15s %-20s\n", "", promPort, s.status)
	} else {
		fmt.Printf("%-10d %-15s %-20s\n", s.pid, promPort, s.status)
	}
}

func printHeader(long bool, c1, c2 int) {
	fmt.Printf("Datacenter%s Poller%s ", strings.Repeat(" ", c1-10), strings.Repeat(" ", c2-6))
	if long {
		fmt.Printf("%-10s %-15s %-10s %-20s\n", "PID", "PromPort", "Profiling", "Status")
	} else {
		fmt.Printf("%-10s %-15s %-20s\n", "PID", "PromPort", "Status")
	}
}

func printBreak(long bool, c1, c2 int) {
	fmt.Printf("%s %s ", strings.Repeat("+", c1), strings.Repeat("+", c2))
	if long {
		fmt.Println("++++++++++ +++++++++++++++ ++++++++++ ++++++++++++++++++++")
	} else {
		fmt.Println("++++++++++ +++++++++++++++ ++++++++++++++++++++")
	}

}

// maximum size of datacenter and poller names, if exceed defaults
func getMaxLengths(pollers *node.Node, pn, dc int) (int, int) {
	for _, p := range pollers.GetChildren() {
		if len(p.GetNameS()) > pn {
			pn = len(p.GetNameS())
		}
		if len(p.GetChildContentS("datacenter")) > dc {
			dc = len(p.GetChildContentS("datacenter"))
		}
	}
	return dc + 1, pn + 1
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

func getPollerPrometheusPort(p *node.Node, opts *options) int {
	var promPort int
	var err error

	// check first if poller argument has promPort defined
	// else in exporter config of poller
	if opts.promPort != 0 {
		promPort = opts.promPort
	} else {
		promPort, err = conf.GetPrometheusExporterPorts(p.GetNameS())
		if err != nil {
			fmt.Println(err)
			return 0
		}
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
	rootCmd.AddCommand(config.ConfigCmd, zapi.Cmd, grafana.GrafanaCmd, stub.NewCmd)
	rootCmd.AddCommand(generate.Cmd)
	rootCmd.AddCommand(doctor.Cmd)

	rootCmd.PersistentFlags().StringVar(&opts.config, "config", "./harvest.yml", "harvest config file path")
	rootCmd.Version = version.String()
	rootCmd.SetVersionTemplate(version.String())
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + `
Feedback
  Open issues at https://github.com/NetApp/harvest
`)

	startCmd.PersistentFlags().BoolVarP(
		&opts.debug,
		"debug",
		"d",
		false,
		"debug mode collects data, but no HTTP daemons and no writes to DBs",
	)
	startCmd.PersistentFlags().BoolVarP(
		&opts.verbose,
		"verbose",
		"v",
		false,
		"verbose logging (loglevel=1)",
	)
	startCmd.PersistentFlags().BoolVarP(
		&opts.trace,
		"trace",
		"t",
		false,
		"trace logging (loglevel=0)",
	)
	startCmd.PersistentFlags().BoolVarP(
		&opts.foreground,
		"foreground",
		"f",
		false,
		"start poller in foreground (only one poller, implies debug mode)",
	)
	startCmd.PersistentFlags().BoolVar(
		&opts.daemon,
		"daemon",
		true,
		"start poller in background",
	)
	startCmd.PersistentFlags().IntVarP(
		&opts.loglevel,
		"loglevel",
		"l",
		2,
		"logging level (0=trace, 1=debug, 2=info, 3=warning, 4=error, 5=critical)",
	)
	startCmd.PersistentFlags().BoolVar(
		&opts.profiling,
		"profiling",
		false,
		"if profiling port > 0, enables profiling via localhost:PORT/debug/pprof/",
	)
	startCmd.PersistentFlags().IntVar(
		&opts.promPort,
		"promPort",
		0,
		"prometheus port to use for HTTP endpoint",
	)
	startCmd.PersistentFlags().StringSliceVarP(
		&opts.collectors,
		"collectors",
		"c",
		[]string{},
		"only start these collectors (overrides harvest.yml)",
	)
	startCmd.PersistentFlags().StringSliceVarP(
		&opts.objects,
		"objects",
		"o",
		[]string{},
		"only start these objects (overrides collector config)",
	)
}

// The management commands: start|status|stop|restart|kill
// are created with this function - all but start are hidden
// to save space
func manageCmd(use string, shouldHide bool) *cobra.Command {
	return &cobra.Command{
		Use:    fmt.Sprintf("%s [POLLER...]", use),
		Short:  "Stop/restart/status/kill - all or individual pollers",
		Long:   "Harvest Manager - manage your pollers",
		Args:   cobra.ArbitraryArgs,
		Hidden: shouldHide,
		Run:    doManageCmd,
	}
}

func main() {
	// Prefer our order to alphabetical
	cobra.EnableCommandSorting = false
	cobra.CheckErr(rootCmd.Execute())
}
