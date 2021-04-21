package manager

import (
	"bytes"
	"fmt"
	"goharvest2/share/argparse"
	"goharvest2/share/config"
	"goharvest2/share/set"
	"goharvest2/share/tree/node"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	HARVEST_HOME string
	HARVEST_CONF string
	HARVEST_PIDS string
)

type options struct {
	Command    string
	Pollers    []string
	Verbose    bool
	Trace      bool
	Debug      bool
	Foreground bool
	Loglevel   int
	Config     string
}

func (o options) print() {
	fmt.Printf("command    = %s\n", o.Command)
	fmt.Printf("pollers    = %v\n", o.Pollers)
	fmt.Printf("verbose    = %v\n", o.Verbose)
	fmt.Printf("trace      = %v\n", o.Trace)
	fmt.Printf("debug      = %v\n", o.Debug)
	fmt.Printf("foreground = %v\n", o.Foreground)
	fmt.Printf("loglevel   = %d\n", o.Loglevel)
}

func Run() {

	if HARVEST_HOME = os.Getenv("HARVEST_HOME"); HARVEST_HOME == "" {
		HARVEST_HOME = "/opt/harvest/"
	}

	if HARVEST_CONF = os.Getenv("HARVEST_CONF"); HARVEST_CONF == "" {
		HARVEST_CONF = "/etc/harvest/"
	}

	if HARVEST_PIDS = os.Getenv("HARVEST_PIDS"); HARVEST_PIDS == "" {
		HARVEST_PIDS = "/var/run/harvest/"
	}

	// default options
	opts := &options{
		Verbose:    false,
		Debug:      false,
		Trace:      false,
		Foreground: false,
		Loglevel:   2,
		Config:     path.Join(HARVEST_CONF, "harvest.yml"),
	}

	// parse user-defined options
	parser := argparse.New("Harvest Manager", "harvest", "manage your pollers")

	parser.PosString(
		&opts.Command,
		"command",
		"action to take, one of:",
		[]string{"status", "start", "restart", "stop", "kill"},
	)

	parser.PosSlice(
		&opts.Pollers,
		"pollers",
		"names of pollers as defined in config [harvest.yml] (default: all)",
	)

	parser.Bool(
		&opts.Verbose,
		"verbose",
		"v",
		"verbose logging (equals to loglevel=1)",
	)

	parser.Bool(
		&opts.Trace,
		"trace",
		"t",
		"aggressively verbose logging (equals to loglevel=0)",
	)

	parser.Bool(
		&opts.Debug,
		"debug",
		"d",
		"debug mode (collect data, but no HTTP daemons and no writes to DBs)",
	)

	parser.Bool(
		&opts.Foreground,
		"foreground",
		"f",
		"start poller in foreground (only one poller, implies debug mode)",
	)

	parser.Int(
		&opts.Loglevel,
		"loglevel",
		"l",
		"logging level (0=trace, 1=debug, 2=info, 3=warn, 4=error, 5=fatal)",
	)

	parser.String(
		&opts.Config,
		"config",
		"c",
		"Custom config filepath (default: "+opts.Config+")",
	)

	parser.SetHelpFlag("help")

	// user asked for help or invalid options
	if !parser.Parse() {
		os.Exit(0)
	}

	//parser.PrintValues()

	if opts.Debug {
		opts.Loglevel = 1
	}

	if opts.Trace {
		opts.Loglevel = 0
	}

	pollers, err := config.GetPollers(opts.Config)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("config [%s]: not found\n", opts.Config)
		} else {
			fmt.Printf("config [%s]: %v\n", opts.Config, err)
		}
		os.Exit(1)
	}

	if len(opts.Pollers) > 0 {
		// verify poller names
		ok := true
		for _, p := range opts.Pollers {
			if pollers.GetChildS(p) == nil {
				fmt.Printf("poller [%s] not defined\n", p)
				ok = false
			}
		}
		if !ok {
			os.Exit(1)
		}
		// filter pollers
		poller_names := set.NewFrom(opts.Pollers)
		for _, p := range pollers.GetChildren() {
			if !poller_names.Has(p.GetNameS()) {
				pollers.PopChildS(p.GetNameS())
			}
		}
	}

	if opts.Foreground {
		if opts.Command != "start" {
			fmt.Printf("invalid command [%s] for foreground mode\n", opts.Command)
			os.Exit(1)
		}
		if len(pollers.GetChildren()) != 1 {

			fmt.Println("only one poller can be started in foreground mode")
			os.Exit(1)
		}
		if !opts.Debug {
			//opts.Debug = true
			fmt.Println("set debug mode ON (starting poller in foreground otherwise is unsafe)")
		}
		p := pollers.GetChildren()[0]
		startPoller(p.GetNameS(), opts)
		os.Exit(0)
	}

	c1, c2 := getMaxLengths(pollers, 20, 20)
	printHeader(c1, c2)
	printBreak(c1, c2)

	for _, p := range pollers.GetChildren() {
		name := p.GetNameS()
		datacenter := p.GetChildContentS("datacenter")
		port := p.GetChildContentS("prometheus_port")

		if opts.Command == "kill" {
			status, pid := killPoller(name)
			printStatus(c1, c2, datacenter, name, port, status, pid)
			continue
		}

		status, pid := getStatus(name)

		if opts.Command == "status" {
			printStatus(c1, c2, datacenter, name, port, status, pid)
		}

		if opts.Command == "stop" || opts.Command == "restart" {
			status, pid = stopPoller(name)
			printStatus(c1, c2, datacenter, name, port, status, pid)
		}

		if opts.Command == "start" || opts.Command == "restart" {
			// only start poller if confirmed that it's not running
			if status == "not running" || status == "stopped" {
				status, pid = startPoller(name, opts)
				printStatus(c1, c2, datacenter, name, port, status, pid)
			}
		}
	}

	printBreak(c1, c2)
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
func getStatus(poller_name string) (string, int) {

	var (
		status string
		pid    int
	)
	// running poller should have written PID to file
	pid_fp := path.Join(HARVEST_PIDS, poller_name+".pid")

	// no PID file, assume process exited or never started
	if data, err := ioutil.ReadFile(pid_fp); err != nil {
		status = "not running"
		// corrupt PID should never happen
		// might be a sign of system failure or unexpected shutdown
	} else if pid, err = strconv.Atoi(string(data)); err != nil {
		status = "invalid pid"
	}

	// docker dummy status
	if os.Getenv("HARVEST_DOCKER") == "yes" {
		return "n/a", pid
	}

	// no valid PID stops here
	if pid < 1 {
		return status, pid
	}

	// retrieve process with PID and check if running
	// (error can be safely ignored on Unix)
	proc, _ := os.FindProcess(pid)

	// send signal 0 to check if process is runnign
	// returns err if it's not running or permission is denied
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		if os.IsPermission(err) {
			fmt.Println("Insufficient priviliges to send signal to process")
		}
		return "unknown", pid
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
	if data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid)); err == nil {
		cmdline := string(bytes.ReplaceAll(data, []byte("\x00"), []byte(" ")))

		if strings.Contains(cmdline, "--daemon") {
			s := strings.SplitAfter(cmdline, "--poller ")
			if len(s) == 2 && strings.HasPrefix(s[1], poller_name) {
				status = "running"
			}
		}
	}

	// if status is not running, either process just exited and cmdline was unavailable
	// or cmdline did not confirm process is the poller we're looking for
	if status != "running" {
		status = "unknown"
	}

	return status, pid
}

func killPoller(poller_name string) (string, int) {

	var (
		status string
		pid    int
	)

	defer cleanPidFile(poller_name)

	// attempt to get pid from pid file
	status, pid = getStatus(poller_name)

	// attempt to get pid from "ps aux"
	if pid < 1 {
		data, err := exec.Command("ps", "aux").Output()
		if err != nil {
			fmt.Println("ps aux: ", err)
			return status, pid
		}

		var fields, args []string

		for _, line := range strings.Split(string(data), "\n") {
			if fields = strings.Fields(line); len(fields) != 11 {
				continue
			}
			if args = strings.Fields(fields[10]); len(args) < 3 {
				continue
			}

			if !strings.HasSuffix(args[0], "poller") && args[2] != poller_name {
				continue
			}

			if x, err := strconv.Atoi(fields[1]); err == nil {
				pid = x
			}
			break
		}
	}

	if pid < 1 {
		return status, pid
	}
	proc, _ := os.FindProcess(pid)
	if err := proc.Kill(); err != nil {
		fmt.Println("kill: ", err)
		return status, pid
	}
	return "killed", pid

}

// Stop poller if it's running or it's stoppable
//
// Returns: same as get_status()
func stopPoller(poller_name string) (string, int) {
	var (
		status string
		pid    int
	)

	// if we get no valid PID, assume process is not running
	if status, pid = getStatus(poller_name); pid < 1 {
		return status, pid
	}

	proc, _ := os.FindProcess(pid)

	// send terminate signal
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if os.IsPermission(err) {
			fmt.Println("Insufficient priviliges to terminate process")
			return "stopping failed", pid
		}
		fmt.Println(err)
		// other errors, probably mean poller already exited, so double check
		return getStatus(poller_name)
	}

	// give the poller chance to cleanup and exit
	for i := 0; i < 5; i += 1 {
		time.Sleep(200 * time.Millisecond)
		// @TODO, handle situation when PID is regained by some other process
		if proc.Signal(syscall.Signal(0)) != nil {
			return "stopped", pid
		}
	}
	return "stopping failed", pid
}

func startPoller(poller_name string, opts *options) (string, int) {

	argv := make([]string, 5)
	argv[0] = path.Join(HARVEST_HOME, "bin", "poller")
	argv[1] = "--poller"
	argv[2] = poller_name
	argv[3] = "--loglevel"
	argv[4] = strconv.Itoa(opts.Loglevel)

	if opts.Debug {
		argv = append(argv, "--debug")
	}

	if opts.Config != path.Join(HARVEST_CONF, "harvest.yml") {
		argv = append(argv, "--conf")
		argv = append(argv, opts.Config)
	}

	if opts.Foreground {
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

	cmd := exec.Command(path.Join(HARVEST_HOME, "bin", "daemonize"), argv...)
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
		if status, pid := getStatus(poller_name); pid > 0 {
			return status, pid
		}
		time.Sleep(50 * time.Millisecond)
	}

	return getStatus(poller_name)
}

// Clean PID file if it exists
// Return value indicates wether PID file existed
func cleanPidFile(poller string) bool {
	fp := path.Join(HARVEST_PIDS, poller+".pid")
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
func printStatus(c1, c2 int, dc, pn, port, status string, pid int) {
	fmt.Printf("%s%s ", dc, strings.Repeat(" ", c1-len(dc)))
	fmt.Printf("%s%s ", pn, strings.Repeat(" ", c2-len(pn)))
	if pid == 0 {
		fmt.Printf("%-10s %-10s %-20s\n", port, "", status)
	} else {
		fmt.Printf("%-10s %-10d %-20s\n", port, pid, status)
	}
}

func printHeader(c1, c2 int) {
	fmt.Printf("Datacenter%s Poller%s ", strings.Repeat(" ", c1-10), strings.Repeat(" ", c2-6))
	fmt.Printf("%-10s %-10s %-20s\n", "Port", "PID", "Status")
}

func printBreak(c1, c2 int) {
	fmt.Printf("%s %s ", strings.Repeat("+", c1), strings.Repeat("+", c2))
	fmt.Println("++++++++++ ++++++++++ ++++++++++++++++++++")
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
