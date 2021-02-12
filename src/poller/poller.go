package main

import (
	"runtime"
	"errors"
	"sync"
	"os"
	"os/signal"
	"syscall"
	"strconv"
	"path"
	"plugin"
	"io/ioutil"
	"strings"
    "goharvest2/share/logger"
	"goharvest2/poller/schedule"
	"goharvest2/poller/collector"
	"goharvest2/poller/exporter"
	"goharvest2/poller/struct/yaml"
	"goharvest2/poller/struct/options"
)

var SIGNALS = []os.Signal{
			syscall.SIGHUP, 
			syscall.SIGINT, 
			syscall.SIGTERM,
			syscall.SIGQUIT,
}

type Poller struct {
	Name string
	prefix string
	options *options.Options
	pid int
	pidf string
	schedule *schedule.Schedule
	collectors []collector.Collector
	exporters []exporter.Exporter
	exporter_params *yaml.Node
	params *yaml.Node
	//metadata *metadata.Metadata
}

func New() *Poller {
	return &Poller{}
}

func (p *Poller) Init() error {

	var err error
	/* Set poller main attributes */
	p.options, p.Name, err = options.GetOpts()

	p.prefix = "(poller) (" + p.Name + ")"

	//p.options.Print()

	/* If daemon, make sure handler outputs to file */
	if p.options.Daemon {
		err := logger.OpenFileOutput(p.options.Path, "harvest_poller_" + p.Name + ".log")
		if err != nil {
			return err
		}
	}
	logger.SetLevel(p.options.LogLevel)

	/* Useful info for debugging */
	if p.options.Debug {
		p.LogDebugInfo()
	}

	signal_channel := make(chan os.Signal, 1)
	signal.Notify(signal_channel, SIGNALS...)
	go p.handleSignals(signal_channel)
	logger.Debug(p.prefix, "Set signal handler for %v", SIGNALS)

	/* Write PID to file */ 
	err = p.registerPid()
	if err != nil {
		logger.Warn(p.prefix, "Failed to write PID file: %v", err)
	}

	/* Announce startup */
	if p.options.Daemon {
		logger.Info(p.prefix, "Starting as daemon [pid=%d] [pid file=%s]", p.pid, p.pidf)
	} else {
		logger.Info(p.prefix, "Starting in foreground [pid=%d] [pid file=%s]", p.pid, p.pidf)
	}

	/* Set Harvest API handler */
	//go p.handleFifo()

	/* Initialize exporters and collectors */
	if p.params, p.exporter_params, err = ReadConfig(p.options.Path, p.options.Config, p.Name); err != nil {
		logger.Error(p.prefix, "Failed to read config: %v", err)
		return err
	} else if p.exporter_params == nil {
		logger.Warn(p.prefix, "No exporters defined in config")
	}

	if collectors := p.params.GetChild("collectors"); collectors != nil {
		if len(p.options.Collectors) > 0 {
			collectors.FilterValues(p.options.Collectors)
			logger.Debug(p.prefix, "Filtered collectors: %v (=%d)", p.options.Collectors, len(collectors.Children))
		}
		for _, c := range collectors.Values {
			p.load_collector(c, "")
		}
	} else {
		logger.Warn(p.prefix, "No collectors defined for poller")
		return errors.New("No collectors")
	}

	if len(p.collectors) == 0 {
		logger.Warn(p.prefix, "No collectors initialized, stopping")
		return errors.New("No collectors")
	}
	logger.Debug(p.prefix, "Initialized %d collectors", len(p.collectors))
	
	if len(p.exporters) == 0 {
		logger.Warn(p.prefix, "No exporters initialized, continuing without exporters")
	} else {
		logger.Debug(p.prefix, "Initialized %d exporters", len(p.exporters))
	}

	//@todo interval default = 60s
	p.schedule = schedule.New()
	if err := p.schedule.AddTaskString("poller", "20s", nil); err != nil {
		logger.Error(p.prefix, "Setting schedule: %v", err)
		return err
	}

	/* Famous last words */
	logger.Info(p.prefix, "Poller start-up complete.")

	return nil

}

func (p *Poller) load_module(binpath, name string) (*plugin.Plugin, error) {

	var err error
	var files []os.FileInfo
	var fn string

	if files, err = ioutil.ReadDir(binpath); err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.Name() == name + ".so" {
			fn = f.Name()
			break
		}
	}

	if fn == "" {
		logger.Warn(p.prefix, "Failed to find %s.so file in [%s]", name, binpath)
		return nil, errors.New(".so file not found")
	}

	return plugin.Open(path.Join(binpath, fn))
}

func (p *Poller) load_collector(class, object string) error {

	var err error
	var module *plugin.Plugin
	var sym plugin.Symbol
	var binpath string
	var template *yaml.Node
	var subcollectors []collector.Collector

	binpath = path.Join(p.options.Path, "bin", "collectors")

	if module, err = p.load_module(binpath, strings.ToLower(class)); err != nil {
		logger.Error(p.prefix, "load .so: %v", err)
		return err
	}

	if sym, err = module.Lookup("New"); err != nil {
		logger.Error(p.prefix, "load New(): %v", err)
		return err
	}

	NewFunc, ok := sym.(func(*collector.AbstractCollector) collector.Collector)
	if !ok {
		logger.Error(p.prefix, "New() has not expected signature")
		return errors.New("incompatible New()")
	}

	if template, err = collector.ImportTemplate(p.options.Path, class); err != nil {
		logger.Error(p.prefix, "load template: %v", err)
		return err
	} else if template == nil {  // probably redundant
		logger.Error(p.prefix, "empty template")
		return errors.New("empty template")
	}
	// log: imported and merged template...
	template.Union(p.params, false)

	// if we don't know object, try load from template
	if object == "" {
		object = template.GetChildValue("object")
	}

	// if object is defined, we only initialize 1 subcollector / object
	if object != "" {
		c := NewFunc(collector.New(class, object, p.options, template.Copy()))
		if err = c.Init(); err != nil {
			logger.Error(p.prefix, "init [%s:%s]: %v", class, object, err)
			return err
		} else {
			subcollectors = append(subcollectors, c)
			logger.Debug(p.prefix, "intialized collector [%s:%s]", class, object)
		}
	// if template has list of objects, initialiez 1 subcollector for each
	} else if objects := template.GetChild("objects"); objects != nil {
		
		if len(p.options.Objects) > 0 {
			objects.FilterChildren(p.options.Objects)
			logger.Debug(p.prefix, "Filtered Objects: %v (=%d)", p.options.Objects, len(objects.Children))
		}
		for _, object := range objects.GetChildren() {
			c := NewFunc(collector.New(class, object.Name, p.options, template.Copy()))
			if err = c.Init(); err != nil {
				logger.Error(p.prefix, "init [%s:%s]: %v", class, object.Name, err)
				return err
			} else {
				subcollectors = append(subcollectors, c)
				logger.Debug(p.prefix, "intialized subcollector [%s:%s]", class, object.Name)
			}
		}
	} else {
		return errors.New("no object defined in template")
	}

	p.collectors = append(p.collectors, subcollectors...)
	logger.Debug(p.prefix, "initialized [%s] with %d subcollectors", class, len(subcollectors))

	// link each collector with requested exporter
	for _, c := range subcollectors {
		for _, e := range c.WantedExporters() {
			if exp := p.load_exporter(e); exp != nil {
				c.LinkExporter(exp)
				logger.Debug(p.prefix, "Linked [%s:%s] to exporter [%s]", c.GetName(), c.GetObject(), e)
			} else {
				logger.Warn(p.prefix, "Exporter [%s] requested by [%s:%s] not available", e, c.GetName(), c.GetObject())
			}
		}
	}
	return nil
}


func (p *Poller) get_exporter(name string) exporter.Exporter {
	for _, exp := range p.exporters {
		if exp.GetName() == name {
			return exp
		}
	}
	return nil
}


func (p *Poller) load_exporter(name string) exporter.Exporter {

	var err error
	var module *plugin.Plugin
	var sym plugin.Symbol
	var binpath string
	var params, class *yaml.Node

	if e := p.get_exporter(name); e != nil {
		return e
	}

	if params = p.exporter_params.GetChild(name); params == nil {
		logger.Warn(p.prefix, "Exporter [%s] not defined in config", name)
		return nil
	}

	if class = params.GetChild("exporter"); class == nil {
		logger.Warn(p.prefix, "Exporter [%s] missing field \"exporter\"", name)
		return nil
	}
	binpath = path.Join(p.options.Path, "bin", "exporters")

	if module, err = p.load_module(binpath, strings.ToLower(class.Value)); err != nil {
		logger.Error(p.prefix, "load .so: %v", err)
		return nil
	}

	if sym, err = module.Lookup("New"); err != nil {
		logger.Error(p.prefix, "load New(): %v", err)
		return nil
	}

	NewFunc, ok := sym.(func(*exporter.AbstractExporter) exporter.Exporter)
	if !ok {
		logger.Error(p.prefix, "New() has not expected signature")
		return nil
	}

	e := NewFunc(exporter.New(class.Value, name, p.options, params))
	if err = e.Init(); err != nil {
		logger.Error(p.prefix, "Failed initializing exporter [%s]: %v", name, err)
		return nil
	}

	p.exporters = append(p.exporters, e)
	logger.Info(p.prefix, "Initialized exporter [%s]", name)
	return e
	
}

func (p *Poller) Start() {

	var wg sync.WaitGroup

	/* Start collectors */
	for _, col := range p.collectors {
		logger.Debug(p.prefix, "Starting collector [%s]", col.GetName())
		wg.Add(1)
		go col.Start(&wg)
	}

	go p.selfMonitor()

	wg.Wait()
	//time.Sleep(30 * time.Second)

	logger.Info(p.prefix, "No active collectors. Poller terminating.")
	p.Stop()

	//os.Exit(0)
	return
}

func (p *Poller) Stop() {
	logger.Info(p.prefix, "Cleaning up and stopping Poller [pid=%d]", p.pid)

	if p.options.Daemon {

		var err error

		err = os.Remove(p.pidf)
		if err != nil {
			logger.Warn(p.prefix, "Failed to clean pid file: %v", err)
		} else {
			logger.Debug(p.prefix, "Clean pid file [%s]", p.pidf)
		}

		err = logger.CloseFileOutput()
		if err != nil {
			logger.Error(p.prefix, "Failed to close log file: %v", err)
		}
	}
}

func (p *Poller) selfMonitor() {

	task, _ := p.schedule.GetTask("poller")

	for {

		if task.IsDue() {

			task.Start()

			up_collectors := 0
			up_exporters := 0

			for _, c := range p.collectors {
				if code, status, msg := c.GetStatus(); code == 1 {
					logger.Debug(p.prefix, "collector status: %d (%s) - %s", code, status, msg)
					up_collectors += 1
				}
			}

			for _, e := range p.exporters {
				if code, status, msg := e.GetStatus(); code == 1 {
					logger.Debug(p.prefix, "exporter status: %d (%s) - %s", code, status, msg)
					up_exporters += 1
				}
			}

			logger.Info(p.prefix, "Updated status: %d up collectors (of %d) and %d up exporters (of %d)", up_collectors, len(p.collectors), up_exporters, len(p.exporters))

		}
		
		p.schedule.Sleep()

	}
}

func (p *Poller) handleSignals(signal_channel chan os.Signal) {
	for {
		sig := <-signal_channel
		logger.Info(p.prefix, "Caught signal [%s]", sig)
		p.Stop()
		os.Exit(0)
	}
}

func (p *Poller) handleFifo() {
	logger.Info(p.prefix, "Serving APIs for Harvest2 daemon")
	for {
		;
	}
}

func (p *Poller) registerPid() error {
	var err error
	p.pid = os.Getpid()
	if p.options.Daemon {
		var file *os.File
		p.pidf = path.Join(p.options.Path, "var", "." + p.Name + ".pid")
		file, err = os.Create(p.pidf)
		if err == nil {
			_, err = file.WriteString(strconv.Itoa(p.pid))
			if err == nil {
				file.Sync()
			}
			file.Close()
		}
	}
	return err
}

func (p *Poller) LogDebugInfo() {

	var err error
	var hostname string
	var st syscall.Sysinfo_t

	logger.Debug(p.prefix, "Options: path=[%s], config=[%s], daemon=%v, debug=%v, loglevel=%d", 
		p.options.Path, p.options.Config, p.options.Daemon, p.options.Debug, p.options.LogLevel)
	hostname, err  = os.Hostname()
	logger.Debug(p.prefix, "Running on [%s]: system [%s], arch [%s], CPUs=%d", 
		hostname, runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
	logger.Debug(p.prefix, "Poller Go build version [%s]", runtime.Version())
	
	st = syscall.Sysinfo_t{}
	err = syscall.Sysinfo(&st)
	if err == nil {
		logger.Debug(p.prefix, "System uptime [%d], Memory [%d] / Free [%d]. Running processes [%d]", 
			st.Uptime, st.Totalram, st.Freeram, st.Procs)
	}
}


func ReadConfig(harvest_path, config_fn, name string) (*yaml.Node, *yaml.Node, error) {
	var err error
	var config, pollers, p, exporters, defaults *yaml.Node

	config, err = yaml.Import(path.Join(harvest_path, config_fn))

	if err == nil {

		pollers = config.GetChild("Pollers")
		defaults = config.GetChild("Defaults")

		if pollers == nil {
			err = errors.New("No pollers defined")
		} else {
			p = pollers.GetChild(name)
			if p == nil {
				err = errors.New("Poller [" + name + "] not defined")
			} else if defaults != nil {
				p.Union(defaults, false)
			}
		}
	}

	if err == nil && p != nil {

		exporters = config.GetChild("Exporters")
		if exporters == nil {
			//logger.Warn(p.prefix, "No exporters defined in config [%s]", config)
			;
		} else {
			requested := p.GetChild("exporters")
			redundant := make([]*yaml.Node, 0)
			if requested != nil {
				for _, e := range exporters.Children {
					if !requested.HasInValues(e.Name) {
						redundant = append(redundant, e)
					}
				}
				for _, e := range redundant {
					exporters.PopChild(e.Name)
				}
			}
		}
	}

	return p, exporters, err
}

func main() {

	/*
	filepath := path.Join("tests", "Poller_shopfloor_003.cpu")
	cpuFile, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}

	pprof.StartCPUProfile(cpuFile)
	*/

    p := New()

    if err := p.Init(); err == nil {

		p.Start()

	} else {
		p.Stop()
	}

	/*
	pprof.StopCPUProfile()
	cpuFile.Close()

	os.Exit(0)
	*/
}
