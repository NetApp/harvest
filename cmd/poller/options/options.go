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
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
)

type Options struct {
	Poller     string   // name of the Poller
	Daemon     bool     // if true, Poller is started as daemon
	Debug      bool     // if true, Poller is started in debug mode, which means data will not be exported
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
	Asup       bool     // if true, invoke autosupport at start-up
	IsTest     bool     // true when run from unit test
	ConfPath   string   // colon-separated paths to search for templates
	ConfPaths  []string // sliced version of `ConfPath`, list of paths to search for templates
}

func New(opts ...Option) *Options {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}
	o.SetDefaults()
	return o
}

type Option func(*Options)

func WithConfPath(path string) Option {
	return func(o *Options) {
		o.ConfPath = path
	}
}

func WithConfigPath(path string) Option {
	return func(o *Options) {
		o.Config = path
	}
}

func (o *Options) MarshalZerologObject(e *zerolog.Event) {
	e.Str("config", o.Config)
	e.Str("confPath", o.ConfPath)
	e.Bool("daemon", o.Daemon)
	e.Bool("debug", o.Debug)
	e.Int("profiling", o.Profiling)
	e.Int("promPort", o.PromPort)
	e.Str("homePath", o.HomePath)
	e.Str("logPath", o.LogPath)
	e.Str("logPath", o.LogPath)
	e.Str("hostname", o.Hostname)
	e.Bool("asup", o.Asup)
}

func (o *Options) SetDefaults() *Options {
	if hostname, err := os.Hostname(); err == nil {
		o.Hostname = hostname
	}

	o.HomePath = conf.Path()
	o.LogPath = conf.GetHarvestLogPath()
	o.SetConfPath(o.ConfPath)

	return o
}

func (o *Options) SetConfPath(colonSeperatedPath string) {
	o.ConfPath = colonSeperatedPath
	o.ConfPaths = filepath.SplitList(colonSeperatedPath)
}
