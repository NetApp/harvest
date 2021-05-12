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
	"goharvest2/cmd/harvest/version"
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/exporter"
	"goharvest2/cmd/poller/options"
	"goharvest2/cmd/poller/schedule"
	"goharvest2/pkg/config"
	"goharvest2/pkg/dload"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"log/syslog"
	"net/http"
	_ "net/http/pprof" // #nosec since pprof is off by default
	"os"
	"os/exec"
	"os/signal"
	"path"
	"plugin"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// default params
var (
	_POLLER_SCHEDULE string = "60s"
	_LOG_FILE_NAME   string = ""
	_LOG_MAX_BYTES   int64  = 10000000 // 10MB
	_LOG_MAX_FILES   int    = 10
)

// signals to catch
var SIGNALS = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT,
}

// deprecated collectors to throw warning
var _DEPRECATED_COLLECTORS = map[string]string{
	"psutil": "Unix",
}

// Poller is the instance that starts and monitors a
// group of collectors and exporters as a single UNIX process
type Poller struct {
	name            string
	prefix          string
	target          string
	options         *options.Options
	pid             int
	pidf            string
	schedule        *schedule.Schedule
	collectors      []collector.Collector
	exporters       []exporter.Exporter
	exporter_params *node.Node
	params          *node.Node
	metadata        *matrix.Matrix
	status          *matrix.Matrix
}

// New returns a new instance of Poller
func New() *Poller {
	return &Poller{}
}

// Init starts Poller, reads parameters, opens log handler, initializes metadata,
// starts collectors and exporters
func (me *Poller) Init() error {

	var err error

	// read options
	me.options, me.name, err = options.Get()
	if err != nil {
		logger.Error(me.prefix, "error: %s", err.Error())
		return err
	}

	// use prefix for logging
	me.prefix = "(poller) (" + me.name + ")"

	// if we are daemon, use file logging
	if me.options.Daemon {
		_LOG_FILE_NAME = "poller_" + me.name + ".log"
		if err = logger.OpenFileOutput(me.options.LogPath, _LOG_FILE_NAME); err != nil {
			return err
		}
	}

	// set logging level
	if err = logger.SetLevel(me.options.LogLevel); err != nil {
		logger.Warn(me.prefix, "using default loglevel=2 (info): %s", err.Error())
	}

	// if profiling port > 0 start profiling service
	if me.options.Profiling > 0 {
		addr := fmt.Sprintf("localhost:%d", me.options.Profiling)
		logger.Info(me.prefix, "profiling enabled on [%s]", addr)
		go func() {
			fmt.Println(http.ListenAndServe(addr, nil))
		}()
	}

	// useful info for debugging
	logger.Debug(me.prefix, "* %s *s", version.String())
	logger.Debug(me.prefix, "options= %s", me.options.String())

	// set signal handler for graceful termination
	signal_channel := make(chan os.Signal, 1)
	signal.Notify(signal_channel, SIGNALS...)
	go me.handleSignals(signal_channel)
	logger.Debug(me.prefix, "set signal handler for %v", SIGNALS)

	// write PID to file
	if err = me.registerPid(); err != nil {
		logger.Warn(me.prefix, "failed to write PID file: %v", err)
		return err
	}

	// announce startup
	if me.options.Daemon {
		logger.Info(me.prefix, "started as daemon [pid=%d] [pid file=%s]", me.pid, me.pidf)
	} else {
		logger.Info(me.prefix, "started in foreground [pid=%d]", me.pid)
	}

	// load parameters from config (harvest.yml)
	logger.Debug(me.prefix, "importing config [%s]", me.options.Config)
	if me.params, err = config.GetPoller(me.options.Config, me.name); err != nil {
		logger.Error(me.prefix, "read config: %v", err)
		return err
	}

	// log handling parameters
	// size of file before rotating
	if s := me.params.GetChildContentS("log_max_bytes"); s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			_LOG_MAX_BYTES = i
		}
	}

	// maximum number of rotated files to keep
	if s := me.params.GetChildContentS("log_max_files"); s != "" {
		if i, err := strconv.Atoi(s); err == nil {
			_LOG_MAX_FILES = i
		}
	}

	// each poller is associated with a remote host
	// if no address is specified, assume that is local host
	// @TODO: remove, redundany and error-prone
	if me.target = me.params.GetChildContentS("addr"); me.target == "" {
		me.target = "localhost"
	}

	// initialize our metadata, the metadata will host status of our
	// collectors and exporters, as well as ping stats to target host
	me.load_metadata()

	if me.exporter_params, err = config.GetExporters(me.options.Config); err != nil {
		logger.Warn(me.prefix, "read exporter params: %v", err)
		// @TODO just warn or abort?
	}

	// iterate over list of collectors and initialize them
	// exporters are initialized on the fly, if at least one collector
	// has requested them
	if collectors := me.params.GetChildS("collectors"); collectors == nil {
		logger.Warn(me.prefix, "no collectors defined for this poller in config")
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
				logger.Debug(me.prefix, "skipping collector [%s]", c)
				continue
			}

			if err = me.load_collector(c, ""); err != nil {
				logger.Error(me.prefix, "load collector (%s): %v", c, err)
			}
		}
	}

	// at least one collector should successfully initialize
	if len(me.collectors) == 0 {
		logger.Warn(me.prefix, "no collectors initialized, stopping")
		return errors.New(errors.ERR_NO_COLLECTOR, "no collectors")
	}

	logger.Debug(me.prefix, "initialized %d collectors", len(me.collectors))

	// we are more tolerable against exporters, since we might only
	// want to debug collectors without actually exporting
	if len(me.exporters) == 0 {
		logger.Warn(me.prefix, "no exporters initialized, continuing without exporters")
	} else {
		logger.Debug(me.prefix, "initialized %d exporters", len(me.exporters))
	}

	// initialze a schedule for the poller, this is the interval at which
	// we will check the status of collectors, exporters and target system,
	// and send metadata to exporters
	if s := me.params.GetChildContentS("poller_schedule"); s != "" {
		_POLLER_SCHEDULE = s
	}
	me.schedule = schedule.New()
	if err = me.schedule.NewTaskString("poller", _POLLER_SCHEDULE, nil); err != nil {
		logger.Error(me.prefix, "set schedule: %v", err)
		return err
	}
	logger.Debug(me.prefix, "set poller schedule with %s frequency", _POLLER_SCHEDULE)

	// famous last words
	logger.Info(me.prefix, "poller start-up complete")

	return nil

}

// Start will run the collectors and the poller itself
// in seperate goroutines, leaving the main goroutine
// to the exporters
func (me *Poller) Start() {

	var (
		wg  sync.WaitGroup
		col collector.Collector
	)

	// start collectors
	for _, col = range me.collectors {
		logger.Debug(me.prefix, "launching collector (%s:%s)", col.GetName(), col.GetObject())
		wg.Add(1)
		go col.Start(&wg)
	}

	// run concurrently and update metadata
	go me.Run()

	wg.Wait()

	// ...until there are no collectors running anymore
	logger.Info(me.prefix, "no active collectors -- terminating")

	me.Stop()
}

// Run will periodicaly check the status of collectors/exporters,
// report metadata and do some housekeeping
func (me *Poller) Run() {

	// poller schedule has just one task
	task := me.schedule.GetTask("poller")

	// number of collectors/exporters that are still up
	up_collectors := 0
	up_exporters := 0

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
				logger.Debug(me.prefix, "collector (%s:%s) status: (%d - %s) %s", c.GetName(), c.GetObject(), code, status, msg)

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
				logger.Debug(me.prefix, "exporter (%s) status: (%d - %s) %s", e.GetName(), code, status, msg)

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
					logger.Error(me.prefix, "export component metadata: %v", err)
				}
				if err := e.Export(me.status); err != nil {
					logger.Error(me.prefix, "export target metadata: %v", err)
				}
			}

			// only log when numbers have changes, since hopefully that happens rarely
			if upc != up_collectors || upe != up_exporters {
				logger.Info(me.prefix, "updated status, up collectors: %d (of %d), up exporters: %d (of %d)", upc, len(me.collectors), upe, len(me.exporters))
			}
			up_collectors = upc
			up_exporters = upe

			// some housekeeping jobs if we are daemon
			// @TODO: syslog or panic on log file related errors (might mean fs is corrupt or unavailable)
			// @TODO: probably delegate to log handler (both rotating and panicing)
			if me.options.Daemon {
				// check size of log file
				if stat, err := os.Stat(path.Join(me.options.LogPath, _LOG_FILE_NAME)); err != nil {
					logger.Error(me.prefix, "stat: %v", err)
					// rotate if exceeds threshold
				} else if stat.Size() >= _LOG_MAX_BYTES {
					logger.Debug(me.prefix, "rotating log (size= %d bytes)", stat.Size())
					if err = logger.Rotate(me.options.LogPath, _LOG_FILE_NAME, _LOG_MAX_FILES); err != nil {
						logger.Error(me.prefix, "rotating log: %v", err)
					}
				}
			}
		}

		me.schedule.Sleep()
	}
}

// Stop gracefully exits the program by closing log file and removing pid file
func (me *Poller) Stop() {
	logger.Info(me.prefix, "cleaning up and stopping [pid=%d]", me.pid)

	if me.options.Daemon {
		if err := os.Remove(me.pidf); err != nil {
			logger.Error(me.prefix, "clean pid file: %v", err)
		} else {
			logger.Debug(me.prefix, "cleaned pid file [%s]", me.pidf)
		}

		if err := logger.CloseFileOutput(); err != nil {
			logger.Error(me.prefix, "close log file: %v", err)
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
func (me *Poller) handleSignals(signal_channel chan os.Signal) {
	for {
		sig := <-signal_channel
		logger.Info(me.prefix, "caught signal [%s]", sig)
		me.Stop()
		os.Exit(0)
	}
}

// ping target system, report weither it's available or not
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
func (me *Poller) load_collector(class, object string) error {

	var (
		err              error
		sym              plugin.Symbol
		binpath          string
		template, custom *node.Node
		collectors       []collector.Collector
		col              collector.Collector
	)

	// path to the shared object (.so file)
	binpath = path.Join(me.options.HomePath, "bin", "collectors")

	// throw warning for deprecated collectors
	if r, d := _DEPRECATED_COLLECTORS[strings.ToLower(class)]; d {
		if r != "" {
			logger.Warn(me.prefix, "collector (%s) is deprecated, please use (%s) instead", class, r)
		} else {
			logger.Warn(me.prefix, "collector (%s) is deprecated, see documentation for help", class)
		}
	}

	if sym, err = dload.LoadFuncFromModule(binpath, strings.ToLower(class), "New"); err != nil {
		return err
	}

	NewFunc, ok := sym.(func(*collector.AbstractCollector) collector.Collector)
	if !ok {
		return errors.New(errors.ERR_DLOAD, "New() has not expected signature")
	}

	// load the template file(s) of the collector where we expect to find
	// object name or list of objects
	if template, err = collector.ImportTemplate(me.options.ConfPath, "default.yaml", class); err != nil {
		return err
	} else if template == nil { // probably redundant
		return errors.New(errors.MISSING_PARAM, "collector template")
	}

	if custom, err = collector.ImportTemplate(me.options.ConfPath, "custom.yaml", class); err == nil && custom != nil {
		template.Merge(custom)
		logger.Debug(me.prefix, "merged custom and default templates")
	}

	// add Poller's parametres to the collector parameters
	template.Union(me.params)

	// if we don't know object, try load from template
	if object == "" {
		object = template.GetChildContentS("object")
	}

	// if object is defined, we only initialize 1 subcollector / object
	if object != "" {
		col = NewFunc(collector.New(class, object, me.options, template.Copy()))
		if err = col.Init(); err != nil {
			logger.Error(me.prefix, "init collector (%s:%s): %v", class, object, err)
		} else {
			collectors = append(collectors, col)
			logger.Debug(me.prefix, "initialized collector (%s:%s)", class, object)
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
				logger.Debug(me.prefix, "skipping object [%s]", object.GetNameS())
				continue
			}

			col = NewFunc(collector.New(class, object.GetNameS(), me.options, template.Copy()))
			if err = col.Init(); err != nil {
				logger.Warn(me.prefix, "init collector-object (%s:%s): %v", class, object.GetNameS(), err)
				if errors.IsErr(err, errors.ERR_CONNECTION) {
					logger.Warn(me.prefix, "aborting collector (%s)", class)
					break
				}
			} else {
				collectors = append(collectors, col)
				logger.Debug(me.prefix, "initialized collector-object (%s:%s)", class, object.GetNameS())
			}
		}
	} else {
		return errors.New(errors.MISSING_PARAM, "collector object")
	}

	me.collectors = append(me.collectors, collectors...)
	logger.Debug(me.prefix, "initialized (%s) with %d objects", class, len(collectors))

	// link each collector with requested exporter & update metadata
	for _, col = range collectors {

		name := col.GetName()
		obj := col.GetObject()

		for _, exp_name := range col.WantedExporters(me.options.Config) {
			logger.Trace(me.prefix, "exp_name %s", exp_name)
			if exp := me.load_exporter(exp_name); exp != nil {
				col.LinkExporter(exp)
				logger.Debug(me.prefix, "linked (%s:%s) to exporter (%s)", name, obj, exp_name)
			} else {
				logger.Warn(me.prefix, "exporter (%s) requested by (%s:%s) not available", exp_name, name, obj)
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

// returns exporter that matches to name, if exporter is not loaded
// tries to load and return
func (me *Poller) load_exporter(name string) exporter.Exporter {

	var (
		err            error
		sym            plugin.Symbol
		binpath, class string
		params         *node.Node
		exp            exporter.Exporter
	)

	// stop here if exporter is already loaded
	if exp = me.get_exporter(name); exp != nil {
		return exp
	}

	if params = me.exporter_params.GetChildS(name); params == nil {
		logger.Warn(me.prefix, "exporter (%s) not defined in config", name)
		return nil
	}

	if class = params.GetChildContentS("exporter"); class == "" {
		logger.Warn(me.prefix, "exporter (%s) has no exporter class defined", name)
		return nil
	}

	binpath = path.Join(me.options.HomePath, "bin", "exporters")

	if sym, err = dload.LoadFuncFromModule(binpath, strings.ToLower(class), "New"); err != nil {
		logger.Error(me.prefix, "dload: %v", err.Error())
		return nil
	}

	NewFunc, ok := sym.(func(*exporter.AbstractExporter) exporter.Exporter)
	if !ok {
		logger.Error(me.prefix, "New() has not expected signature")
		return nil
	}

	exp = NewFunc(exporter.New(class, name, me.options, params))
	if err = exp.Init(); err != nil {
		logger.Error(me.prefix, "init exporter (%s): %v", name, err)
		return nil
	}

	me.exporters = append(me.exporters, exp)
	logger.Debug(me.prefix, "initialized exporter (%s)", name)

	// update metadata
	if instance, err := me.metadata.NewInstance(exp.GetClass() + "." + exp.GetName()); err != nil {
		logger.Error(me.prefix, "add metadata instance: %v", err)
	} else {
		instance.SetLabel("type", "exporter")
		instance.SetLabel("name", exp.GetClass())
		instance.SetLabel("target", exp.GetName())
	}
	return exp

}

func (me *Poller) get_exporter(name string) exporter.Exporter {
	for _, exp := range me.exporters {
		if exp.GetName() == name {
			return exp
		}
	}
	return nil
}

// initialize matrices to be used as metadata
func (me *Poller) load_metadata() {

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

// start poller, if fails try to write to syslog
func main() {

	// don't recover if a goroutine has paniced, instead
	// try to log as much as possible, since normally it's
	// not properly logged
	defer func() {
		//logger.Warn("(main) ", "defer func here")
		if r := recover(); r != nil {
			syslogger, err := syslog.NewLogger(syslog.LOG_ERR|syslog.LOG_DAEMON, logger.LOG_FLAGS)
			if err == nil {
				syslogger.Printf("harvest poller paniced: %v", r)
			}
			// if logger still abailable try to write there as well
			// do this last, since might make us panic as again
			logger.Fatal("(main) ", "%v", r)
			logger.Fatal("(main) ", "terminating abnormally, tip: run in foreground with \"-l 0\" mode to debug")

			os.Exit(1)
		}
	}()

	poller := New()

	if poller.Init() != nil {
		// error already logger by poller
		poller.Stop()
		os.Exit(1)
	}

	poller.Start()
	os.Exit(0)
}
