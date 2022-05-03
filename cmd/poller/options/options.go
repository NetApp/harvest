/*
   Copyright NetApp Inc, 2021 All rights reserved

   Package options provides Poller start-up options. These are fetched from CLI arguments,
   default values and/or environment variables. Some of the options are left blank and will
   be set by the Poller.

   Options is declared in a separate file to make it possible for collector/exporters
   to access it.

*/

package options

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/conf"
	"os"
	"strings"
)

type Options struct {
	Poller string // name of the Poller
	Daemon bool   // if true, Poller is started as daemon
	Debug  bool   // if true, Poller is started in debug mode
	// this mostly means that no data will be exported
	PromPort   int      // HTTP port that is assigned to Poller and can be used by the Prometheus exporter
	Config     string   // filepath of Harvest config (defaults to "harvest.yml") can be relative or absolute path
	HomePath   string   // path to harvest home (usually "/opt/harvest")
	LogPath    string   // log files location (usually "/var/log/harvest")
	LogLevel   int      // logging level, 0 for trace, 5 for fatal
	LogToFile  bool     // when running in the foreground, log to file instead of stdout
	Version    string   // harvest version
	Hostname   string   // hostname of the machine harvest is running
	Collectors []string // name of collectors to load (override poller config)
	Objects    []string // objects to load (overrides collector config)
	Profiling  int      // in case of profiling, the HTTP port used to display results
	Asup       bool     // if true, invoke autosupport at start up
}

// String provides a string representation of Options
func (o *Options) String() string {
	x := []string{
		fmt.Sprintf("%s= %s", "Poller", o.Poller),
		fmt.Sprintf("%s = %v", "Daemon", o.Daemon),
		fmt.Sprintf("%s = %v", "Debug", o.Debug),
		fmt.Sprintf("%s = %d", "Profiling", o.Profiling),
		fmt.Sprintf("%s = %d", "PromPort", o.PromPort),
		fmt.Sprintf("%s = %d", "LogLevel", o.LogLevel),
		fmt.Sprintf("%s = %s", "HomePath", o.HomePath),
		fmt.Sprintf("%s = %s", "LogPath", o.LogPath),
		fmt.Sprintf("%s = %s", "Config", o.Config),
		fmt.Sprintf("%s = %s", "Hostname", o.Hostname),
		fmt.Sprintf("%s = %s", "Version", o.Version),
		fmt.Sprintf("%s = %v", "Asup", o.Asup),
	}
	return strings.Join(x, ", ")
}

// Print writes Options to STDOUT
func (o *Options) Print() {
	fmt.Println(o.String())
}

func SetPathsAndHostname(args *Options) {
	if hostname, err := os.Hostname(); err == nil {
		args.Hostname = hostname
	}

	args.HomePath = conf.GetHarvestHomePath()

	args.LogPath = conf.GetHarvestLogPath()
}
