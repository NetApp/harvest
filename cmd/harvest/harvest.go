/*
 * Copyright NetApp Inc, 2021 All rights reserved

NetApp Harvest 2.0: the swiss-army-knife for datacenter monitoring

Authors:
   Georg Mey & Vachagan Gratian
Contact:
   ng-harvest-maintainers@netapp.com

This project is based on NetApp Harvest, authored by
Chris Madden in 2015.

*/
package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"goharvest2/cmd/harvest/config"
	"goharvest2/cmd/harvest/stub"
	"goharvest2/cmd/harvest/version"
	"goharvest2/cmd/tools/grafana"
	"goharvest2/cmd/tools/zapi"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/set"
	"goharvest2/pkg/tree/node"
	"io/ioutil"
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
	HarvestPidPath    string
)

var rootCmd = &cobra.Command{
	Use:   "harvest <command> <subcommand> [flags]",
	Short: "NetApp Harvest 2.0 - application for monitoring storage systems",
	Long: `NetApp Harvest 2.0 - application for monitoring storage systems
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

	if HarvestPidPath = os.Getenv("HARVEST_PIDS"); HarvestPidPath == "" {
		HarvestPidPath = "/var/run/harvest/"
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
		if !opts.debug {
			opts.debug = true
			fmt.Println("set debug mode ON (starting poller in foreground otherwise is unsafe)")
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
		promPort := getPollerPrometheusPort(p, opts)

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
			if s.status == "not running" || s.status == "stopped" || s.status == "killed" {
				s = startPoller(name, promPort, opts)
				printStatus(opts.longStatus, c1, c2, datacenter, name, s.promPort, s)
			} else {
				fmt.Printf("can't verify status of [%s]: kill poller and try again\n", name)
			}
		}
	}

	printBreak(opts.longStatus, c1, c2)
}

// Trace status of a poller. This is partially guesswork and
// there is no guarantee that we will be always correct
// The general logic is:
// - if no PID file, assume poller exited or never started
// - if PID file exists, but no process with PID exists, assume poller interrupted
// - if a process with PID is running, assume poller is running if it has expected cmdline
//
// Returns:
//	@status - status of poller
//  @pid  - PID of poller (0 means no PID)
func getStatus(pollerName string) *pollerStatus {

	s := &pollerStatus{}
	// running poller should have written PID to file
	pidFp := path.Join(HarvestPidPath, pollerName+".pid")

	// no PID file, assume process exited or never started
	if data, err := ioutil.ReadFile(pidFp); err != nil {
		s.status = "not running"
		// corrupt PID should never happen
		// might be a sign of system failure or unexpected shutdown
	} else if s.pid, err = strconv.Atoi(string(data)); err != nil {
		s.status = "invalid pid"
	}

	// docker dummy status
	if os.Getenv("HARVEST_DOCKER") == "yes" {
		s.status = "n/a (docker)"
		return s
	}

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
		s.status = "unknown: " + err.Error()
		return s
		// process not running, but did not clean PID file
		// maybe it just exited, so give it a chance to clean
		/*
			time.Sleep(500 * time.Millisecond)
			if clean_pidf(pid_fp) {
				return "interrupted", pid
			}
			return "exited", pid
		*/
	}

	// process is running, validate that it's the poller we're looking fore
	// since PID might have changed (although very unlikely)
	if data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", s.pid)); err == nil {
		cmdline := string(bytes.ReplaceAll(data, []byte("\x00"), []byte(" ")))

		if checkPollerIdentity(cmdline, pollerName) {
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
	}

	// if status is not running, either process just exited and cmdline was unavailable
	// or cmdline did not confirm process is the poller we're looking for
	if s.status == "" {
		s.status = "unknown/unmatched"
	}

	return s
}

func checkPollerIdentity(cmdline, pollerName string) bool {
	if x := strings.Fields(cmdline); len(x) == 0 || !strings.HasSuffix(x[0], "poller") {
		return false
	}

	if !strings.Contains(cmdline, "--daemon") {
		return false
	}

	if !strings.Contains(cmdline, "--poller") {
		return false
	}

	x := strings.SplitAfter(cmdline, "--poller ")
	if len(x) != 2 {
		return false
	}

	if y := strings.Fields(x[1]); len(y) == 0 || y[0] != pollerName {
		return false
	}
	return true
}

func killPoller(pollerName string) *pollerStatus {

	defer cleanPidFile(pollerName)

	// attempt to get pid from pid file
	s := getStatus(pollerName)

	// attempt to get pid from "ps aux"
	if s.pid < 1 {
		data, err := exec.Command("ps", "aux").Output()
		if err != nil {
			fmt.Println("ps aux: ", err)
			return s
		}

		for _, line := range strings.Split(string(data), "\n") {
			// BSD format should have 11 columns
			// last column can contain whitespace, so we should get at least 11
			if fields := strings.Fields(line); len(fields) > 10 {
				// CLI args are everything after 10th column
				if checkPollerIdentity(strings.Join(fields[10:], " "), pollerName) {
					if x, err := strconv.Atoi(fields[1]); err == nil {
						s.pid = x
					}
					break
				}
			}
		}
	}

	// stop if couldn't find pid
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
			fmt.Println("Insufficient priviliges to terminate process")
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

func startPoller(pollerName string, promPort string, opts *options) *pollerStatus {

	argv := make([]string, 5)
	argv[0] = path.Join(HarvestHomePath, "bin", "poller")
	argv[1] = "--poller"
	argv[2] = pollerName
	argv[3] = "--loglevel"
	argv[4] = strconv.Itoa(opts.loglevel)

	if len(promPort) != 0 {
		argv = append(argv, "--promPort")
		argv = append(argv, promPort)
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

	// if pid directory doesn't exist, create full path, otherwise poller will complain
	if info, err := os.Stat(HarvestPidPath); err != nil || !info.IsDir() {
		// don't abort on error, since another poller might have done the job
		if err = os.MkdirAll(HarvestPidPath, 0755); err != nil && !os.IsExist(err) {
			fmt.Printf("error mkdir [%s]: %v\n", HarvestPidPath, err)
		}
	}

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

	// Poller should immediately write its PID to file at startup
	// Allow for some delay and retry checking status a few times
	time.Sleep(50 * time.Millisecond)
	for i := 0; i < 10; i += 1 {
		// @TODO, handle situation when PID is regained by some other process
		if s := getStatus(pollerName); s.pid > 0 {
			return s
		}
		time.Sleep(50 * time.Millisecond)
	}

	return getStatus(pollerName)
}

// Clean PID file if it exists
// Return value indicates wether PID file existed
func cleanPidFile(pollerName string) bool {
	fp := path.Join(HarvestPidPath, pollerName+".pid")
	if err := os.Remove(fp); err != nil {
		if os.IsPermission(err) {
			fmt.Printf("Error: you have no permission to remove [%s]\n", fp)
		} else if !os.IsNotExist(err) {
			fmt.Printf("Error: %v\n", err)
		}
		return false
	}
	return true
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

func getPollerPrometheusPort(p *node.Node, opts *options) string {
	var promPort string
	var err error
	// check first if poller argument has promPort defined
	// else in exporter config of poller
	if opts.promPort != 0 {
		promPort = strconv.Itoa(opts.promPort)
	} else {
		promPort, err = conf.GetPrometheusExporterPorts(p, opts.config)
		if err != nil {
			fmt.Println(err)
			promPort = "error"
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
	rootCmd.AddCommand(config.ConfigCmd, zapi.ZapiCmd, grafana.GrafanaCmd, stub.NewCmd)

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
		Short:  "stop/restart/status/kill - all or individual pollers",
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
