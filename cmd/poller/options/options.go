/*
 * Copyright NetApp Inc, 2021 All rights reserved

   Package options provides Poller start-up options. These are fetched from CLI arguments,
   default values and/or environment variables. Some of the options are left blank and will
   be set by the Poller.

   Options is declared in a seperate file to make it possible for collector/exporters
   to access it.

*/
package options

import (
	"fmt"
	"goharvest2/cmd/harvest/version"
	"goharvest2/pkg/argparse"
	"os"
	"path"
	"strings"
)

type Options struct {
	Poller string // name of the Poller
	Daemon bool   // if true, Poller is started as daemon
	Debug  bool   // if true, Poller is started in debug mode
	// this mostly means that no data will be exported
	PrometheusPort string   // HTTP port that is assigned to Poller and can be used by the Prometheus exporter
	Config         string   // filename of Harvest config (e.g. "harvest.yml")
	ConfPath       string   // path to config directory (usually "/etc/harvest")
	HomePath       string   // path to harvest home (usually "/opt/harvest")
	LogPath        string   // log files location (usually "/var/log/harvest")
	PidPath        string   // pid files location (usually "/var/run/harvest")
	LogLevel       int      // logging level, 0 for trace, 5 for fatal
	Version        string   // harvest version
	Hostname       string   // hostname of the machine harvest is running
	Collectors     []string // name of collectors to load (override poller config)
	Objects        []string // objects to load (overrides collector config)
	Profiling      int      // in case of profiling, the HTTP port used to display results
}

// String provides a string representation of Options
func (o *Options) String() string {
	x := []string{
		fmt.Sprintf("%s= %s", "Poller", o.Poller),
		fmt.Sprintf("%s = %v", "Daemon", o.Daemon),
		fmt.Sprintf("%s = %v", "Debug", o.Debug),
		fmt.Sprintf("%s = %d", "Profiling", o.Profiling),
		fmt.Sprintf("%s = %s", "PrometheusPort", o.PrometheusPort),
		fmt.Sprintf("%s = %d", "LogLevel", o.LogLevel),
		fmt.Sprintf("%s = %s", "HomePath", o.HomePath),
		fmt.Sprintf("%s = %s", "ConfPath", o.ConfPath),
		fmt.Sprintf("%s = %s", "LogPath", o.LogPath),
		fmt.Sprintf("%s = %s", "PidPath", o.PidPath),
		fmt.Sprintf("%s = %s", "Config", o.Config),
		fmt.Sprintf("%s = %s", "Hostname", o.Hostname),
		fmt.Sprintf("%s = %s", "Version", o.Version),
	}
	return strings.Join(x, ", ")
}

// Print writes Options to STDOUT
func (o *Options) Print() {
	fmt.Println(o.String())
}

// Get retrieves options from CLI flags, env variables and defaults
func Get() (*Options, string) {
	var args Options
	args = Options{}

	// set defaults
	args.Daemon = false
	args.Debug = false
	args.LogLevel = 2
	args.Version = version.VERSION
	if hostname, err := os.Hostname(); err == nil {
		args.Hostname = hostname
	}

	if args.HomePath = os.Getenv("HARVEST_HOME"); args.HomePath == "" {
		args.HomePath = "/opt/harvest/"
	}
	if args.ConfPath = os.Getenv("HARVEST_CONF"); args.ConfPath == "" {
		args.ConfPath = "/etc/harvest/"
	}
	if args.LogPath = os.Getenv("HARVEST_LOGS"); args.LogPath == "" {
		args.LogPath = "/var/log/harvest/"
	}
	if args.PidPath = os.Getenv("HARVEST_PIDS"); args.PidPath == "" {
		args.PidPath = "/var/run/harvest/"
	}

	args.Config = path.Join(args.ConfPath, "harvest.yml")

	// parse from command line
	parser := argparse.New("Harvest Poller", "poller", "Runs collectors and exporters for a target system")
	parser.String(&args.Poller, "poller", "p", "Poller name as defined in config")
	parser.Bool(&args.Debug, "debug", "d", "Debug mode, no data will be exported")
	parser.Bool(&args.Daemon, "daemon", "", "Start as daemon")
	parser.Int(&args.LogLevel, "loglevel", "l", "Logging level (0=trace, 1=debug, 2=info, 3=warning, 4=error, 5=critical)")
	parser.Int(&args.Profiling, "profiling", "", "If profiling port > 0, enables profiling via locahost:PORT/debug/pprof/")
	parser.String(&args.Config, "conf", "", "Custom config filepath (default: "+args.Config+")")
	parser.Slice(&args.Collectors, "collectors", "c", "Only start these collectors (overrides harvest.yml)")
	parser.Slice(&args.Objects, "objects", "o", "Only start these objects (overrides collector config)")

	parser.SetHelpFlag("help")
	parser.ParseOrExit()

	if args.Poller == "" {
		fmt.Println("Missing required argument: poller")
		os.Exit(1)
	}

	return &args, args.Poller
}
