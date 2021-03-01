package main

import (
	"os"
	"os/exec"
	"fmt"
	"path"
	"io/ioutil"
	"strings"
	"strconv"
	"syscall"
	"bytes"
	"time"
	"goharvest2/share/argparse"
	"goharvest2/share/config"
    "goharvest2/share/set"
)

var (
	HARVEST_HOME string
	HARVEST_CONF string
	HARVEST_PIDS string
)

type options struct {
	Command string
	Pollers []string
	Verbose bool
	Trace bool
	Debug bool
	Foreground bool
	Loglevel int
    Config string
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

func main() {

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
        Verbose: false,
        Debug: false,
        Trace: false,
        Foreground: false,
        Loglevel: 2,
        Config: path.Join(HARVEST_CONF, "harvest.yml"),
    }

	// parse user-defined options
	parser := argparse.New("Harvest Manager", "harvest", "manage your pollers")

	parser.PosString(
		&opts.Command,
		"command",
		"action to take, one of:",
		[]string{"status", "start", "restart", "stop"},
	)

	parser.PosSlice(
		&opts.Pollers,
		"pollers",
		"names of pollers as defined in config [harvest.yml], yield all pollers if left empty",
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
        "Custom config filepath (default: " + opts.Config + ")",
    )

	// user asked for help or invalid options
	if ! parser.Parse() {
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
			fmt.Printf("config [%s] not found\n", opts.Config)
		} else {
			fmt.Println(err)
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
            if ! poller_names.Has(p.GetNameS()) {
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
			opts.Debug = true
			fmt.Println("set debug mode ON (starting poller in foreground otherwise is unsafe)")
		}
		p := pollers.GetChildren()[0]
		fmt.Printf("starting [%s] in foreground mode...\n", p.GetNameS())
		start_poller(p.GetNameS(), opts)
		os.Exit(0)
	}

	// normal start/stop/status operation
	fmt.Printf("%-20s%-20s%-20s%-20s%-10s\n", "Datacenter", "Poller", "Prometheus Port", "Status", "PID")
	fmt.Println("+++++++++++++++++++ +++++++++++++++++++ +++++++++++++++++++ +++++++++++++++++++ +++++++++")

	for _, p := range pollers.GetChildren() {
		name := p.GetNameS()
		datacenter := p.GetChildContentS("datacenter")
		port := p.GetChildContentS("prometheus_port")

		if opts.Command == "status" {
			status, pid := get_status(name)
			fmt.Printf("%-20s%-20s%-20s%-20s%-10d\n", datacenter, name, port, status, pid)
		}

		if opts.Command == "stop" || opts.Command == "restart" {
			status, pid := stop_poller(name)
			fmt.Printf("%-20s%-20s%-20s%-20s%-10d\n", datacenter, name, port, status, pid)
		}

		if opts.Command == "start" || opts.Command == "restart" {
			status, pid := start_poller(name, opts)
			fmt.Printf("%-20s%-20s%-20s%-20s%-10d\n", datacenter, name, port, status, pid)
		}
	}
	fmt.Println("+++++++++++++++++++ +++++++++++++++++++ +++++++++++++++++++ +++++++++++++++++++ +++++++++")
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
func get_status(poller_name string) (string, int) {
	var status string
	var pid int

	// running poller should have written PID to file
	pid_fp := path.Join(HARVEST_PIDS, poller_name + ".pid")

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
        return "na", pid
    }

	// no valid PID stops here
	if pid < 1 {
		return status, pid
	}

	// retrieve process with PID and check if running
	// (error can be safely ignored on Unix)
	proc, _ := os.FindProcess(pid)

	// sendingg signal 0, return error if
	//  process is not running or permission is denied
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		if os.IsPermission(err) {
			fmt.Println("Error: you have no permission to send signal to process")
			return "unknown", pid
		}
		// process not running, but did not clean PID file
		// maybe it just exited, so give it a chance to clean
		time.Sleep(500 * time.Millisecond)
		if clean_pidf(pid_fp) {
			return "interrupted", pid
		}
		return "exited", pid
	}

	// process is running, validate that it's the poller we're looking fore
	// since PID might have changed (although very inlikely)
	if data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid)); err == nil {
		cmdline := string(bytes.ReplaceAll(data, []byte("\x00"), []byte(" ")))

		if strings.Contains(cmdline, "--daemon") {
			s := strings.SplitAfter(cmdline, "--poller ")
			if len(s)==2 && strings.HasPrefix(s[1], poller_name) {
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


// Stop poller if it's running or it's stoppable
//
// Returns: same as get_status()
func stop_poller(poller_name string) (string, int) {
	var status string
	var pid int

	// if we get a valid PID, assume process is "stoppable"
	if status, pid = get_status(poller_name); pid < 1 {
		return status, pid
	}

	proc, _ := os.FindProcess(pid)

	// send terminate signal
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if os.IsPermission(err) {
			fmt.Println("Error: you have no permission to send signal to process")
			return "stopping failed", pid
		}
		// other errors, probably mean poller already exited, so double check
		return get_status(poller_name)
	}

	// give the poller chance to cleanup and exit
	for i:=0; i<5; i+=1 {
		time.Sleep(200 * time.Millisecond)
		// @TODO, handle situation when PID is regained by some other process
		if proc.Signal(syscall.Signal(0)) != nil {
			return "stopped", pid
		}
	}
	return "stopping failed", pid
}

func start_poller(poller_name string, opts *options) (string, int) {

	argv := make([]string, 5)
	argv[0] = path.Join(HARVEST_HOME, "bin", "poller")
	argv[1] = "--poller"
	argv[2] = poller_name
	argv[3] = "--loglevel"
	argv[4] = strconv.Itoa(opts.Loglevel)

	if opts.Debug {
		argv = append(argv, "--debug")
	}

	if opts.Foreground {
        cmd := exec.Command(argv[0], argv[1:]...)
        fmt.Println(cmd.String())
		fmt.Println("starting in foreground, enter CTRL+C or close terminal to stop poller")
		os.Stdout.Sync()
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
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
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

    // Poller should immediately write its PID to file at startup
	// Allow for some delay and retry checking status a few times
	time.Sleep(50 * time.Millisecond)
	for i:=0; i<10; i+=1 {
		// @TODO, handle situation when PID is regained by some other process
		if status, pid := get_status(poller_name); pid > 0 {
			return status, pid
		}
		time.Sleep(100 * time.Millisecond)
	}

	return get_status(poller_name)
}

// Clean PID file if it exists
// Return value indicates wether PID file existed
func clean_pidf(fp string) bool {
	if err := os.Remove(fp); err != nil {
		if os.IsPermission(err) {
			fmt.Printf("Error: you have no permission to remove [%s]\n", fp)
		} else if ! os.IsNotExist(err) {
			fmt.Printf("Error: %v\n", err)
		}
		return false
	}
	return true
}
