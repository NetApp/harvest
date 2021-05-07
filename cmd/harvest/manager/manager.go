/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package manager

import (
	"bytes"
	"fmt"
	"goharvest2/pkg/argparse"
	"goharvest2/pkg/config"
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

var (
	HarvestHomePath string
	HarvestConfPath string
	HarvestPidPath  string
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
}

type pollerStatus struct {
	status        string
	pid           int
	profilingPort string
}

func Run() {
	HarvestHomePath = config.GetHarvestHome()
	HarvestConfPath = config.GetHarvestConf()

	if HarvestPidPath = os.Getenv("HARVEST_PIDS"); HarvestPidPath == "" {
		HarvestPidPath = "/var/run/harvest/"
	}

	// default options
	opts := &options{
		loglevel: 2,
		config:   path.Join(HarvestConfPath, "harvest.yml"),
	}

	// parse user-defined options
	parser := argparse.New("Harvest Manager", "harvest", "manage your pollers")
	parser.SetOffset(2)
	parser.SetHelpFlag("help")
	parser.SetHelpFlag("manager")

	parser.PosString(
		&opts.command,
		"command",
		"action to take, one of:",
		[]string{"status", "start", "restart", "stop", "kill"},
	)

	parser.PosSlice(
		&opts.pollers,
		"pollers",
		"names of pollers as defined in config [harvest.yml] (default: all)",
	)

	parser.Bool(
		&opts.verbose,
		"verbose",
		"v",
		"verbose logging (equals to loglevel=1)",
	)

	parser.Bool(
		&opts.trace,
		"trace",
		"t",
		"aggressively verbose logging (equals to loglevel=0)",
	)

	parser.Bool(
		&opts.debug,
		"debug",
		"d",
		"debug mode (collect data, but no HTTP daemons and no writes to DBs)",
	)

	parser.Bool(
		&opts.foreground,
		"foreground",
		"f",
		"start poller in foreground (only one poller, implies debug mode)",
	)

	parser.Int(
		&opts.loglevel,
		"loglevel",
		"",
		"logging level (0=trace, 1=debug, 2=info, 3=warn, 4=error, 5=fatal)",
	)

	parser.Slice(
		&opts.collectors,
		"collectors",
		"",
		"start poller with only these collectors (overrides harvest.yml)",
	)

	parser.Slice(
		&opts.objects,
		"objects",
		"",
		"start collectors with only these objects (overrides collector template)",
	)

	parser.String(
		&opts.config,
		"config",
		"c",
		"Custom config filepath (default: "+opts.config+")",
	)

	parser.Bool(
		&opts.longStatus,
		"long",
		"l",
		"show advanced options in poller status",
	)

	parser.Bool(
		&opts.profiling,
		"profiling",
		"p",
		"If true enables profiling via locahost:PORT/debug/pprof/",
	)

	// exit if user asked for help or invalid options
	parser.ParseOrExit()

	if opts.verbose {
		opts.loglevel = 1
	}

	if opts.trace {
		opts.loglevel = 0
	}

	pollers, err := config.GetPollers(opts.config)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("config [%s]: not found\n", opts.config)
		} else {
			fmt.Printf("config [%s]: %v\n", opts.config, err)
		}
		os.Exit(1)
	}

	if len(opts.pollers) > 0 {
		// verify poller names
		ok := true
		for _, p := range opts.pollers {
			if pollers.GetChildS(p) == nil {
				fmt.Printf("poller [%s] not defined\n", p)
				ok = false
			}
		}
		if !ok {
			os.Exit(1)
		}
		// filter pollers
		pollerNames := set.NewFrom(opts.pollers)
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
		startPoller(p.GetNameS(), opts)
		os.Exit(0)
	}

	// get max lengths of datacanter and poller names
	// so that output does not get distorted if they are too long
	c1, c2 := getMaxLengths(pollers, 20, 20)
	printHeader(opts.longStatus, c1, c2)
	printBreak(opts.longStatus, c1, c2)

	for _, p := range pollers.GetChildren() {

		var s *pollerStatus

		name := p.GetNameS()
		datacenter := p.GetChildContentS("datacenter")
		port := p.GetChildContentS("prometheus_port")

		if opts.command == "kill" {
			s = killPoller(name)
			printStatus(opts.longStatus, c1, c2, datacenter, name, port, s)
			continue
		}

		s = getStatus(name)

		if opts.command == "status" {
			printStatus(opts.longStatus, c1, c2, datacenter, name, port, s)
		}

		if opts.command == "stop" || opts.command == "restart" {
			s = stopPoller(name)
			printStatus(opts.longStatus, c1, c2, datacenter, name, port, s)
		}

		if opts.command == "start" || opts.command == "restart" {
			// only start poller if confirmed that it's not running
			if s.status == "not running" || s.status == "stopped" {
				s = startPoller(name, opts)
				printStatus(opts.longStatus, c1, c2, datacenter, name, port, s)
			} else {
				fmt.Println("can't verify status of [%s]: kill poller and try again", name)
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

	// send signal 0 to check if process is runnign
	// returns err if it's not running or permission is denied
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		if os.IsPermission(err) {
			fmt.Println("Insufficient priviliges to send signal to process")
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
	// since PID might have changed (although very inlikely)
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
		time.Sleep(200 * time.Millisecond)
		// @TODO, handle situation when PID is regained by some other process
		if proc.Signal(syscall.Signal(0)) != nil {
			s.status = "stopped"
		}
	}
	s.status = "stopping failed"
	return s
}

func startPoller(pollerName string, opts *options) *pollerStatus {

	argv := make([]string, 5)
	argv[0] = path.Join(HarvestHomePath, "bin", "poller")
	argv[1] = "--poller"
	argv[2] = pollerName
	argv[3] = "--loglevel"
	argv[4] = strconv.Itoa(opts.loglevel)

	if opts.debug {
		argv = append(argv, "--debug")
	}

	if opts.config != path.Join(HarvestConfPath, "harvest.yml") {
		argv = append(argv, "--conf")
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
		os.Stdout.Sync()
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
func printStatus(long bool, c1, c2 int, dc, pn, prometheusPort string, s *pollerStatus) {
	fmt.Printf("%s%s ", dc, strings.Repeat(" ", c1-len(dc)))
	fmt.Printf("%s%s ", pn, strings.Repeat(" ", c2-len(pn)))
	if long {
		if s.pid == 0 {
			fmt.Printf("%-10s %-10s %-10s %-20s\n", "", prometheusPort, s.profilingPort, s.status)
		} else {
			fmt.Printf("%-10d %-10s %-10s %-20s\n", s.pid, prometheusPort, s.profilingPort, s.status)
		}
	} else if s.pid == 0 {
		fmt.Printf("%-10s %-10s %-20s\n", "", prometheusPort, s.status)
	} else {
		fmt.Printf("%-10d %-10s %-20s\n", s.pid, prometheusPort, s.status)
	}
}

func printHeader(long bool, c1, c2 int) {
	fmt.Printf("Datacenter%s Poller%s ", strings.Repeat(" ", c1-10), strings.Repeat(" ", c2-6))
	if long {
		fmt.Printf("%-10s %-10s %-10s %-20s\n", "PID", "Port", "Profiling", "Status")
	} else {
		fmt.Printf("%-10s %-10s %-20s\n", "PID", "Port", "Status")
	}
}

func printBreak(long bool, c1, c2 int) {
	fmt.Printf("%s %s ", strings.Repeat("+", c1), strings.Repeat("+", c2))
	if long {
		fmt.Println("++++++++++ ++++++++++ ++++++++++ ++++++++++++++++++++")
	} else {
		fmt.Println("++++++++++ ++++++++++ ++++++++++++++++++++")
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
	defer dial.Close()
	return dial.Addr().(*net.TCPAddr).Port, nil
}
