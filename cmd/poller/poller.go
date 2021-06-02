/*
	Copyright NetApp Inc, 2021 All rights reserved

	Package poller implements the program that monitors a target system.
	This can be either a remote host or the local system. Poller is however
	agnostic about the target system and the APIs used to communicate with it.

	Polling the target is done by collectors (sometimes plugins). Conversely,
	storing collected data is done by exporters. All the poller does is
	initialize the collectors and exporters defined in its configuration
	and start them up. All poller parameters are passed on to the collectors.
	Conversely, exporters get only what is explicitly defined as their
	parameters.

	After start-up, poller will periodically check the status of collectors
	and exporters, ping the target system, generate metadata and do some
	housekeeping.

	Usually the poller will run as a daemon. In this case it will create
	a PID file and write logs to a file. For debugging and testing
	it can also be started as a foreground process, in this case
	logs are sent to STDOUT.
*/
package main

import (
	"fmt"
	"github.com/spf13/cobra"
	_ "goharvest2/cmd/collectors/unix"
	_ "goharvest2/cmd/collectors/zapi/collector"
	_ "goharvest2/cmd/collectors/zapiperf"
	"goharvest2/cmd/exporters/influxdb"
	"goharvest2/cmd/exporters/prometheus"
	"goharvest2/cmd/harvest/version"
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/exporter"
	"goharvest2/cmd/poller/options"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/poller/schedule"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logging"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"net/http"
	_ "net/http/pprof" // #nosec since pprof is off by default
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// default params
var (
	pollerSchedule  = "60s"
	logFileName     = ""
	logMaxMegaBytes = logging.DefaultLogMaxMegaBytes // 10MB
	logMaxBackups   = logging.DefaultLogMaxBackups
	logMaxAge       = logging.DefaultLogMaxAge
)

// init with default configuration by default it gets logged both to console and  harvest.log
var logger = logging.Get()

// SIGNALS to catch
var SIGNALS = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT,
}

// deprecated collectors to throw warning
var deprecatedCollectors = map[string]string{
	"psutil": "Unix",
}

// Poller is the instance that starts and monitors a
// group of collectors and exporters as a single UNIX process
type Poller struct {
	name           string
	target         string
	options        *options.Options
	pid            int
	pidf           string
	schedule       *schedule.Schedule
	collectors     []collector.Collector
	exporters      []exporter.Exporter
	exporterParams *node.Node
	params         *node.Node
	metadata       *matrix.Matrix
	status         *matrix.Matrix
}

// Init starts Poller, reads parameters, opens zeroLog handler, initializes metadata,
// starts collectors and exporters
func (me *Poller) Init() error {

	var err error

	// read options
	options.SetPathsAndHostname(&args)
	me.options = &args
	me.name = args.Poller

	fileLoggingEnabled := false
	consoleLoggingEnabled := false
	zeroLogLevel := logging.GetZerologLevel(me.options.LogLevel)
	// if we are daemon, use file logging
	if me.options.Daemon {
		logFileName = "poller_" + me.name + ".log"
		fileLoggingEnabled = true
	} else {
		consoleLoggingEnabled = true
	}

	if me.params, err = conf.GetPoller(me.options.Config, me.name); err != nil {
		// separate logger is not yet configured as it depends on setting logMaxMegaBytes, logMaxFiles later
		// Using default instance of logger which logs below error to harvest.log
		logging.SubLogger("Poller", me.name).Error().Stack().Err(err).Msg("read config")
		return err
	}

	// log handling parameters
	// size of file before rotating
	if s := me.params.GetChildContentS("log_max_bytes"); s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			logMaxMegaBytes = int(i / (1024 * 1024))
		}
	}

	// maximum number of rotated files to keep
	if s := me.params.GetChildContentS("log_max_files"); s != "" {
		if i, err := strconv.Atoi(s); err == nil {
			logMaxBackups = i
		}
	}

	logConfig := logging.LogConfig{ConsoleLoggingEnabled: consoleLoggingEnabled,
		PrefixKey:          "Poller",
		PrefixValue:        me.name,
		LogLevel:           zeroLogLevel,
		FileLoggingEnabled: fileLoggingEnabled,
		Directory:          me.options.LogPath,
		Filename:           logFileName,
		MaxSize:            logMaxMegaBytes,
		MaxBackups:         logMaxBackups,
		MaxAge:             logMaxAge}

	logger = logging.Configure(logConfig)
	logger.Info().Msgf("log level used: %s", zeroLogLevel.String())
	logger.Info().Msgf("options config: %s", me.options.Config)

	// if profiling port > 0 start profiling service
	if me.options.Profiling > 0 {
		addr := fmt.Sprintf("localhost:%d", me.options.Profiling)
		logger.Info().Msgf("profiling enabled on [%s]", addr)
		go func() {
			fmt.Println(http.ListenAndServe(addr, nil))
		}()
	}

	// useful info for debugging
	logger.Debug().Msgf("* %s *s", version.String())
	logger.Debug().Msgf("options= %s", me.options.String())

	// set signal handler for graceful termination
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, SIGNALS...)
	go me.handleSignals(signalChannel)
	logger.Debug().Msgf("set signal handler for %v", SIGNALS)

	// write PID to file
	if err = me.registerPid(); err != nil {
		logger.Warn().Msgf("failed to write PID file: %v", err)
		return err
	}

	// announce startup
	if me.options.Daemon {
		logger.Info().Msgf("started as daemon [pid=%d] [pid file=%s]", me.pid, me.pidf)
	} else {
		logger.Info().Msgf("started in foreground [pid=%d]", me.pid)
	}

	// load parameters from config (harvest.yml)
	logger.Debug().Msgf("importing config [%s]", me.options.Config)

	// each poller is associated with a remote host
	// if no address is specified, assume that is local host
	// @TODO: remove, redundant and error-prone
	if me.target = me.params.GetChildContentS("addr"); me.target == "" {
		me.target = "localhost"
	}
	// check optional parameter auth_style
	// if certificates are missing use default paths
	if me.params.GetChildContentS("auth_style") == "certificate_auth" {
		filenames := [2]string{"ssl_cert", "ssl_key"}
		extensions := [2]string{".pem", ".key"}
		fp := ""
		for i := range filenames {
			if fp = me.params.GetChildContentS(filenames[i]); fp == "" {
				// use default paths
				// example: /opt/harvest/cert/hostname.key, /opt/harvest/cert/hostname.pem
				fp = path.Join(me.options.HomePath, "cert/", me.options.Hostname+extensions[i])
				me.params.SetChildContentS(filenames[i], fp)
				logger.Debug().Msgf("using default [%s] path: [%s]", filenames[i], fp)
			}
			if _, err = os.Stat(fp); err != nil {
				logger.Error().Stack().Err(err).Msgf("%s", filenames[i])
				return errors.New(errors.MISSING_PARAM, filenames[i]+": "+err.Error())
			}
		}
	}

	// initialize our metadata, the metadata will host status of our
	// collectors and exporters, as well as ping stats to target host
	me.loadMetadata()

	if me.exporterParams, err = conf.GetExporters(me.options.Config); err != nil {
		logger.Warn().Msgf("read exporter params: %v", err)
		// @TODO just warn or abort?
	}

	// iterate over list of collectors and initialize them
	// exporters are initialized on the fly, if at least one collector
	// has requested them
	if collectors := me.params.GetChildS("collectors"); collectors == nil {
		logger.Warn().Msg("no collectors defined for this poller in config")
		return errors.New(errors.ERR_NO_COLLECTOR, "no collectors")
	} else {
		for _, c := range collectors.GetAllChildContentS() {
			ok := true
			// if requested, filter collectors
			if len(me.options.Collectors) != 0 {
				ok = false
				for _, x := range me.options.Collectors {
					if x == c {
						ok = true
						break
					}
				}
			}
			if !ok {
				logger.Debug().Msgf("skipping collector [%s]", c)
				continue
			}

			if err = me.loadCollector(c, ""); err != nil {
				logger.Error().Stack().Err(err).Msgf("load collector (%s):", c)
			}
		}
	}

	// at least one collector should successfully initialize
	if len(me.collectors) == 0 {
		logger.Warn().Msg("no collectors initialized, stopping")
		return errors.New(errors.ERR_NO_COLLECTOR, "no collectors")
	}

	logger.Debug().Msgf("initialized %d collectors", len(me.collectors))

	// we are more tolerable against exporters, since we might only
	// want to debug collectors without actually exporting
	if len(me.exporters) == 0 {
		logger.Warn().Msg("no exporters initialized, continuing without exporters")
	} else {
		logger.Debug().Msgf("initialized %d exporters", len(me.exporters))
	}

	// initialize a schedule for the poller, this is the interval at which
	// we will check the status of collectors, exporters and target system,
	// and send metadata to exporters
	if s := me.params.GetChildContentS("poller_schedule"); s != "" {
		pollerSchedule = s
	}
	me.schedule = schedule.New()
	if err = me.schedule.NewTaskString("poller", pollerSchedule, nil); err != nil {
		logger.Error().Stack().Err(err).Msg("set schedule:")
		return err
	}
	logger.Debug().Msgf("set poller schedule with %s frequency", pollerSchedule)

	// famous last words
	logger.Info().Msg("poller start-up complete")

	return nil

}

// Start will run the collectors and the poller itself
// in separate goroutines, leaving the main goroutine
// to the exporters
func (me *Poller) Start() {

	var (
		wg  sync.WaitGroup
		col collector.Collector
	)

	// start collectors
	for _, col = range me.collectors {
		logger.Debug().Msgf("launching collector (%s:%s)", col.GetName(), col.GetObject())
		wg.Add(1)
		go col.Start(&wg)
	}

	// run concurrently and update metadata
	go me.Run()

	wg.Wait()

	// ...until there are no collectors running anymore
	logger.Info().Msg("no active collectors -- terminating")

	me.Stop()
}

// Run will periodically check the status of collectors/exporters,
// report metadata and do some housekeeping
func (me *Poller) Run() {

	// poller schedule has just one task
	task := me.schedule.GetTask("poller")

	// number of collectors/exporters that are still up
	upCollectors := 0
	upExporters := 0

	for {

		if task.IsDue() {

			task.Start()

			// flush metadata
			me.status.Reset()
			me.metadata.Reset()

			// ping target system
			if ping, ok := me.ping(); ok {
				me.status.LazySetValueUint8("status", "host", 0)
				me.status.LazySetValueFloat32("ping", "host", ping)
			} else {
				me.status.LazySetValueUint8("status", "host", 1)
			}

			// add number of goroutines to metadata
			// @TODO: cleanup, does not belong to "status"
			me.status.LazySetValueInt("goroutines", "host", runtime.NumGoroutine())

			upc := 0 // up collectors
			upe := 0 // up exporters

			// update status of collectors
			for _, c := range me.collectors {
				code, status, msg := c.GetStatus()
				logger.Debug().Msgf("collector (%s:%s) status: (%d - %s) %s", c.GetName(), c.GetObject(), code, status, msg)

				if code == 0 {
					upc++
				}

				key := c.GetName() + "." + c.GetObject()

				me.metadata.LazySetValueUint64("count", key, c.GetCollectCount())
				me.metadata.LazySetValueUint8("status", key, code)

				if msg != "" {
					if instance := me.metadata.GetInstance(key); instance != nil {
						instance.SetLabel("reason", msg)
					}
				}
			}

			// update status of exporters
			for _, e := range me.exporters {
				code, status, msg := e.GetStatus()
				logger.Debug().Msgf("exporter (%s) status: (%d - %s) %s", e.GetName(), code, status, msg)

				if code == 0 {
					upe++
				}

				key := e.GetClass() + "." + e.GetName()

				me.metadata.LazySetValueUint64("count", key, e.GetExportCount())
				me.metadata.LazySetValueUint8("status", key, code)

				if msg != "" {
					if instance := me.metadata.GetInstance(key); instance != nil {
						instance.SetLabel("reason", msg)
					}
				}
			}

			// @TODO if there are no "master" exporters, don't collect metadata
			for _, e := range me.exporters {
				if err := e.Export(me.metadata); err != nil {
					logger.Error().Stack().Err(err).Msg("export component metadata:")
				}
				if err := e.Export(me.status); err != nil {
					logger.Error().Stack().Err(err).Msg("export target metadata:")
				}
			}

			// only zeroLog when numbers have changes, since hopefully that happens rarely
			if upc != upCollectors || upe != upExporters {
				logger.Info().Msgf("updated status, up collectors: %d (of %d), up exporters: %d (of %d)", upc, len(me.collectors), upe, len(me.exporters))
			}
			upCollectors = upc
			upExporters = upe
		}

		me.schedule.Sleep()
	}
}

// Stop gracefully exits the program by closing zeroLog file and removing pid file
func (me *Poller) Stop() {
	logger.Info().Msgf("cleaning up and stopping [pid=%d]", me.pid)

	if me.options.Daemon {
		if err := os.Remove(me.pidf); err != nil {
			logger.Error().Stack().Err(err).Msg("clean pid file")
		} else {
			logger.Debug().Msgf("cleaned pid file [%s]", me.pidf)
		}
	}
}

// get PID and write to file if we are daemon
func (me *Poller) registerPid() error {
	me.pid = os.Getpid()

	if me.options.Daemon {

		me.pidf = path.Join(me.options.PidPath, me.name+".pid")

		file, err := os.Create(me.pidf)

		if err == nil {
			if _, err = file.WriteString(strconv.Itoa(me.pid)); err == nil {
				file.Sync()
			}
			file.Close()
		}
		return err
	}
	return nil
}

// set up signal disposition
func (me *Poller) handleSignals(signalChannel chan os.Signal) {
	for {
		sig := <-signalChannel
		logger.Info().Msgf("caught signal [%s]", sig)
		me.Stop()
		os.Exit(0)
	}
}

// ping target system, report if it's available or not
// and if available, response time
func (me *Poller) ping() (float32, bool) {

	cmd := exec.Command("ping", me.target, "-w", "5", "-c", "1", "-q")

	if out, err := cmd.Output(); err == nil {
		if x := strings.Split(string(out), "mdev = "); len(x) > 1 {
			if y := strings.Split(x[len(x)-1], "/"); len(y) > 1 {
				if p, err := strconv.ParseFloat(y[0], 32); err == nil {
					return float32(p), true
				}
			}
		}
	}
	return 0, false
}

// dynamically load and initialize a collector
// if there are more than one objects defined for a collector,
// then multiple collectors will be initialized
func (me *Poller) loadCollector(class, object string) error {

	var (
		err              error
		template, custom *node.Node
		collectors       []collector.Collector
		col              collector.Collector
	)

	// throw warning for deprecated collectors
	if r, d := deprecatedCollectors[strings.ToLower(class)]; d {
		if r != "" {
			logger.Warn().Msgf("collector (%s) is deprecated, please use (%s) instead", class, r)
		} else {
			logger.Warn().Msgf("collector (%s) is deprecated, see documentation for help", class)
		}
	}

	// load the template file(s) of the collector where we expect to find
	// object name or list of objects
	if template, err = collector.ImportTemplate(me.options.HomePath, "default.yaml", class); err != nil {
		return err
	} else if template == nil { // probably redundant
		return errors.New(errors.MISSING_PARAM, "collector template")
	}

	if custom, err = collector.ImportTemplate(me.options.HomePath, "custom.yaml", class); err == nil && custom != nil {
		template.Merge(custom)
		logger.Debug().Msg("merged custom and default templates")
	}
	// add Poller's parameters to the collector parameters
	template.Union(me.params)

	// if we don't know object, try load from template
	if object == "" {
		object = template.GetChildContentS("object")
	}

	// if object is defined, we only initialize 1 sub-collector / object
	if object != "" {
		col, err = me.newCollector(class, object, template)
		if col != nil {
			if err != nil {
				logger.Error().Msgf("init collector (%s:%s): %v", class, object, err)
			} else {
				collectors = append(collectors, col)
				logger.Debug().Msgf("initialized collector (%s:%s)", class, object)
			}
		}
		// if template has list of objects, initialize 1 subcollector for each
	} else if objects := template.GetChildS("objects"); objects != nil {
		for _, object := range objects.GetChildren() {

			ok := true

			// if requested filter objects
			if len(me.options.Objects) != 0 {
				ok = false
				for _, o := range me.options.Objects {
					if o == object.GetNameS() {
						ok = true
						break
					}
				}
			}

			if !ok {
				logger.Debug().Msgf("skipping object [%s]", object.GetNameS())
				continue
			}

			col, err = me.newCollector(class, object.GetNameS(), template)
			if col == nil {
				logger.Warn().Msgf("collector is nil for collector-object (%s:%s)", class, object.GetNameS())
				continue
			}
			if err != nil {
				logger.Warn().Msgf("init collector-object (%s:%s): %v", class, object.GetNameS(), err)
				if errors.IsErr(err, errors.ERR_CONNECTION) {
					logger.Warn().Msgf("aborting collector (%s)", class)
					break
				}
			} else {
				collectors = append(collectors, col)
				logger.Debug().Msgf("initialized collector-object (%s:%s)", class, object.GetNameS())
			}
		}
	} else {
		return errors.New(errors.MISSING_PARAM, "collector object")
	}

	me.collectors = append(me.collectors, collectors...)
	logger.Debug().Msgf("initialized (%s) with %d objects", class, len(collectors))
	// link each collector with requested exporter & update metadata
	for _, col = range collectors {
		if col == nil {
			logger.Warn().Msg("ignoring nil collector")
			continue
		}
		name := col.GetName()
		obj := col.GetObject()

		for _, expName := range col.WantedExporters(me.options.Config) {
			logger.Trace().Msgf("expName %s", expName)
			if exp := me.loadExporter(expName); exp != nil {
				col.LinkExporter(exp)
				logger.Debug().Msgf("linked (%s:%s) to exporter (%s)", name, obj, expName)
			} else {
				logger.Warn().Msgf("exporter (%s) requested by (%s:%s) not available", expName, name, obj)
			}
		}

		// update metadata

		if instance, err := me.metadata.NewInstance(name + "." + obj); err != nil {
			return err
		} else {
			instance.SetLabel("type", "collector")
			instance.SetLabel("name", name)
			instance.SetLabel("target", obj)
		}
	}

	return nil
}

func (me *Poller) newCollector(class string, object string, template *node.Node) (collector.Collector, error) {
	name := "harvest.collector." + strings.ToLower(class)
	mod, err := plugin.GetModule(name)
	if err != nil {
		logger.Error().Msgf("error getting module %s", name)
		return nil, err
	}
	inst := mod.New()
	col, ok := inst.(collector.Collector)
	if !ok {
		logger.Error().Msgf("collector '%s' is not a Collector", name)
		return nil, errors.New(errors.ERR_NO_COLLECTOR, "no collectors")
	}
	delegate := collector.New(class, object, me.options, template.Copy())
	err = col.Init(delegate)
	return col, err
}

// returns exporter that matches to name, if exporter is not loaded
// tries to load and return
func (me *Poller) loadExporter(name string) exporter.Exporter {

	var (
		err    error
		class  string
		params *node.Node
		exp    exporter.Exporter
	)

	// stop here if exporter is already loaded
	if exp = me.getExporter(name); exp != nil {
		return exp
	}

	if params = me.exporterParams.GetChildS(name); params == nil {
		logger.Warn().Msgf("exporter (%s) not defined in config", name)
		return nil
	}

	if class = params.GetChildContentS("exporter"); class == "" {
		logger.Warn().Msgf("exporter (%s) has no exporter class defined", name)
		return nil
	}

	absExp := exporter.New(class, name, me.options, params)
	switch class {
	case "Prometheus":
		exp = prometheus.New(absExp)
	case "InfluxDB":
		exp = influxdb.New(absExp)
	default:
		logger.Error().Msgf("no exporter of name:type %s:%s", name, class)
		return nil
	}
	if err = exp.Init(); err != nil {
		logger.Error().Msgf("init exporter (%s): %v", name, err)
		return nil
	}

	me.exporters = append(me.exporters, exp)
	logger.Debug().Msgf("initialized exporter (%s)", name)

	// update metadata
	if instance, err := me.metadata.NewInstance(exp.GetClass() + "." + exp.GetName()); err != nil {
		logger.Error().Msgf("add metadata instance: %v", err)
	} else {
		instance.SetLabel("type", "exporter")
		instance.SetLabel("name", exp.GetClass())
		instance.SetLabel("target", exp.GetName())
	}
	return exp

}

func (me *Poller) getExporter(name string) exporter.Exporter {
	for _, exp := range me.exporters {
		if exp.GetName() == name {
			return exp
		}
	}
	return nil
}

// initialize matrices to be used as metadata
func (me *Poller) loadMetadata() {

	me.metadata = matrix.New("poller", "metadata_component")
	me.metadata.NewMetricUint8("status")
	me.metadata.NewMetricUint64("count")
	me.metadata.SetGlobalLabel("poller", me.name)
	me.metadata.SetGlobalLabel("version", me.options.Version)
	me.metadata.SetGlobalLabel("hostname", me.options.Hostname)
	me.metadata.SetExportOptions(matrix.DefaultExportOptions())

	// metadata for target system
	me.status = matrix.New("poller", "metadata_target")
	me.status.NewMetricUint8("status")
	me.status.NewMetricFloat32("ping")
	me.status.NewMetricUint32("goroutines")

	instance, _ := me.status.NewInstance("host")
	instance.SetLabel("addr", me.target)
	me.status.SetGlobalLabel("poller", me.name)
	me.status.SetGlobalLabel("version", me.options.Version)
	me.status.SetGlobalLabel("hostname", me.options.Hostname)
	me.status.SetExportOptions(matrix.DefaultExportOptions())
}

var pollerCmd = &cobra.Command{
	Use:   "poller -p name [flags]",
	Short: "Harvest Poller - Runs collectors and exporters for a target system",
	Args:  cobra.NoArgs,
	Run:   startPoller,
}

func startPoller(_ *cobra.Command, _ []string) {
	//cmd.DebugFlags()  // uncomment to print flags
	poller := &Poller{}
	poller.options = &args
	if poller.Init() != nil {
		// error already logger by poller
		poller.Stop()
		os.Exit(1)
	}
	poller.Start()
	os.Exit(0)
}

var args = options.Options{
	Version: version.VERSION,
}

func init() {
	configPath, _ := conf.GetDefaultHarvestConfigPath()

	var flags = pollerCmd.Flags()
	flags.StringVarP(&args.Poller, "poller", "p", "", "Poller name as defined in config")
	flags.BoolVarP(&args.Debug, "debug", "d", false, "Debug mode, no data will be exported")
	flags.BoolVar(&args.Daemon, "daemon", false, "Start as daemon")
	flags.IntVarP(&args.LogLevel, "loglevel", "l", 2, "Logging level (0=trace, 1=debug, 2=info, 3=warning, 4=error, 5=critical)")
	flags.IntVar(&args.Profiling, "profiling", 0, "If profiling port > 0, enables profiling via localhost:PORT/debug/pprof/")
	flags.StringVar(&args.PromPort, "promPort", "", "Prometheus Port")
	flags.StringVar(&args.Config, "config", configPath, "harvest config file path")
	flags.StringSliceVarP(&args.Collectors, "collectors", "c", []string{}, "only start these collectors (overrides harvest.yml)")
	flags.StringSliceVarP(&args.Objects, "objects", "o", []string{}, "only start these objects (overrides collector config)")

	_ = pollerCmd.MarkFlagRequired("poller")
}

// start poller, if fails try to write to syslog
func main() {

	// don't recover if a goroutine has panicked, instead
	// try to zeroLog as much as possible, since normally it's
	// not properly logged
	defer func() {
		//logger.Warn("(main) ", "defer func here")
		if r := recover(); r != nil {
			logger.Info().Msgf("harvest poller panicked: %v", r)
			// if logger still available try to write there as well
			// do this last, since might make us panic as again
			logger.Fatal().Msgf("(main) %v", r)
			logger.Fatal().Msg(`(main) terminating abnormally, tip: run in foreground mode (with "--loglevel 0") to debug`)

			os.Exit(1)
		}
	}()

	cobra.CheckErr(pollerCmd.Execute())
}
