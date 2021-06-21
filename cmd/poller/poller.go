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

	Usually the poller will run as a daemon. In this case it will
	write logs to a file. For debugging and testing
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
	"gopkg.in/yaml.v3"
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
	schedule       *schedule.Schedule
	collectors     []collector.Collector
	exporters      []exporter.Exporter
	exporterParams *node.Node
	params         *conf.Poller
	metadata       *matrix.Matrix
	status         *matrix.Matrix
}

// Init starts Poller, reads parameters, opens zeroLog handler, initializes metadata,
// starts collectors and exporters
func (p *Poller) Init() error {

	var err error

	// read options
	options.SetPathsAndHostname(&args)
	p.options = &args
	p.name = args.Poller

	fileLoggingEnabled := false
	consoleLoggingEnabled := false
	zeroLogLevel := logging.GetZerologLevel(p.options.LogLevel)
	// if we are daemon, use file logging
	if p.options.Daemon {
		logFileName = "poller_" + p.name + ".log"
		fileLoggingEnabled = true
	} else {
		consoleLoggingEnabled = true
	}

	if p.params, err = conf.GetPoller2(p.options.Config, p.name); err != nil {
		// separate logger is not yet configured as it depends on setting logMaxMegaBytes, logMaxFiles later
		// Using default instance of logger which logs below error to harvest.log
		logging.SubLogger("Poller", p.name).Error().Stack().Err(err).Msg("read config")
		return err
	}

	// log handling parameters
	// size of file before rotating
	if s := p.params.LogMaxBytes; s != nil {
		logMaxMegaBytes = int(*s / (1024 * 1024))
	}

	// maximum number of rotated files to keep
	if s := p.params.LogMaxFiles; s != nil {
		logMaxBackups = *p.params.LogMaxFiles
	}

	logConfig := logging.LogConfig{ConsoleLoggingEnabled: consoleLoggingEnabled,
		PrefixKey:          "Poller",
		PrefixValue:        p.name,
		LogLevel:           zeroLogLevel,
		FileLoggingEnabled: fileLoggingEnabled,
		Directory:          p.options.LogPath,
		Filename:           logFileName,
		MaxSize:            logMaxMegaBytes,
		MaxBackups:         logMaxBackups,
		MaxAge:             logMaxAge}

	logger = logging.Configure(logConfig)
	logger.Info().Msgf("log level used: %s", zeroLogLevel.String())
	logger.Info().Msgf("options config: %s", p.options.Config)

	// if profiling port > 0 start profiling service
	if p.options.Profiling > 0 {
		addr := fmt.Sprintf("localhost:%d", p.options.Profiling)
		logger.Info().Msgf("profiling enabled on [%s]", addr)
		go func() {
			fmt.Println(http.ListenAndServe(addr, nil))
		}()
	}

	// useful info for debugging
	logger.Debug().Msgf("* %s *s", version.String())
	logger.Debug().Msgf("options= %s", p.options.String())

	// set signal handler for graceful termination
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, SIGNALS...)
	go p.handleSignals(signalChannel)
	logger.Debug().Msgf("set signal handler for %v", SIGNALS)

	// announce startup
	if p.options.Daemon {
		logger.Info().Msgf("started as daemon [pid=%d]", os.Getpid())
	} else {
		logger.Info().Msgf("started in foreground [pid=%d]", os.Getpid())
	}

	// load parameters from config (harvest.yml)
	logger.Debug().Msgf("importing config [%s]", p.options.Config)

	// each poller is associated with a remote host
	// if no address is specified, assume that is local host
	// @TODO: remove, redundant and error-prone
	if p.params.Addr == nil {
		p.target = "localhost"
	} else {
		p.target = *p.params.Addr
	}
	// check optional parameter auth_style
	// if certificates are missing use default paths
	if p.params.AuthStyle != nil && *p.params.AuthStyle == "certificate_auth" {
		if p.params.SslCert != nil {
			fp := path.Join(p.options.HomePath, "cert/", p.options.Hostname+".pem")
			p.params.SslCert = &fp
			logger.Debug().Msgf("using default [ssl_cert] path: [%s]", fp)
			if _, err = os.Stat(fp); err != nil {
				logger.Error().Stack().Err(err).Msgf("ssl_cert")
				return errors.New(errors.MISSING_PARAM, "ssl_cert: "+err.Error())
			}
		}
		if p.params.SslKey != nil {
			fp := path.Join(p.options.HomePath, "cert/", p.options.Hostname+".key")
			p.params.SslKey = &fp
			logger.Debug().Msgf("using default [ssl_key] path: [%s]", fp)
			if _, err = os.Stat(fp); err != nil {
				logger.Error().Stack().Err(err).Msgf("ssl_key")
				return errors.New(errors.MISSING_PARAM, "ssl_key: "+err.Error())
			}
		}
	}

	// initialize our metadata, the metadata will host status of our
	// collectors and exporters, as well as ping stats to target host
	p.loadMetadata()

	if p.exporterParams, err = conf.GetExporters(p.options.Config); err != nil {
		logger.Warn().Msgf("read exporter params: %v", err)
		// @TODO just warn or abort?
	}

	// iterate over list of collectors and initialize them
	// exporters are initialized on the fly, if at least one collector
	// has requested them
	if p.params.Collectors == nil {
		logger.Warn().Msg("no collectors defined for this poller in config")
		return errors.New(errors.ERR_NO_COLLECTOR, "no collectors")
	} else {
		for _, c := range *p.params.Collectors {
			ok := true
			// if requested, filter collectors
			if len(p.options.Collectors) != 0 {
				ok = false
				for _, x := range p.options.Collectors {
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

			if err = p.loadCollector(c, ""); err != nil {
				logger.Error().Stack().Err(err).Msgf("load collector (%s):", c)
			}
		}
	}

	// at least one collector should successfully initialize
	if len(p.collectors) == 0 {
		logger.Warn().Msg("no collectors initialized, stopping")
		return errors.New(errors.ERR_NO_COLLECTOR, "no collectors")
	}

	logger.Debug().Msgf("initialized %d collectors", len(p.collectors))

	// we are more tolerable against exporters, since we might only
	// want to debug collectors without actually exporting
	if len(p.exporters) == 0 {
		logger.Warn().Msg("no exporters initialized, continuing without exporters")
	} else {
		logger.Debug().Msgf("initialized %d exporters", len(p.exporters))
	}

	// initialize a schedule for the poller, this is the interval at which
	// we will check the status of collectors, exporters and target system,
	// and send metadata to exporters
	if p.params.PollerSchedule != nil {
		pollerSchedule = *p.params.PollerSchedule
	}
	p.schedule = schedule.New()
	if err = p.schedule.NewTaskString("poller", pollerSchedule, nil); err != nil {
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
func (p *Poller) Start() {

	var (
		wg  sync.WaitGroup
		col collector.Collector
	)

	// start collectors
	for _, col = range p.collectors {
		logger.Debug().Msgf("launching collector (%s:%s)", col.GetName(), col.GetObject())
		wg.Add(1)
		go col.Start(&wg)
	}

	// run concurrently and update metadata
	go p.Run()

	wg.Wait()

	// ...until there are no collectors running anymore
	logger.Info().Msg("no active collectors -- terminating")

	p.Stop()
}

// Run will periodically check the status of collectors/exporters,
// report metadata and do some housekeeping
func (p *Poller) Run() {

	// poller schedule has just one task
	task := p.schedule.GetTask("poller")

	// number of collectors/exporters that are still up
	upCollectors := 0
	upExporters := 0

	for {

		if task.IsDue() {

			task.Start()

			// flush metadata
			p.status.Reset()
			p.metadata.Reset()

			// ping target system
			if ping, ok := p.ping(); ok {
				p.status.LazySetValueUint8("status", "host", 0)
				p.status.LazySetValueFloat32("ping", "host", ping)
			} else {
				p.status.LazySetValueUint8("status", "host", 1)
			}

			// add number of goroutines to metadata
			// @TODO: cleanup, does not belong to "status"
			p.status.LazySetValueInt("goroutines", "host", runtime.NumGoroutine())

			upc := 0 // up collectors
			upe := 0 // up exporters

			// update status of collectors
			for _, c := range p.collectors {
				code, status, msg := c.GetStatus()
				logger.Debug().Msgf("collector (%s:%s) status: (%d - %s) %s", c.GetName(), c.GetObject(), code, status, msg)

				if code == 0 {
					upc++
				}

				key := c.GetName() + "." + c.GetObject()

				p.metadata.LazySetValueUint64("count", key, c.GetCollectCount())
				p.metadata.LazySetValueUint8("status", key, code)

				if msg != "" {
					if instance := p.metadata.GetInstance(key); instance != nil {
						instance.SetLabel("reason", msg)
					}
				}
			}

			// update status of exporters
			for _, e := range p.exporters {
				code, status, msg := e.GetStatus()
				logger.Debug().Msgf("exporter (%s) status: (%d - %s) %s", e.GetName(), code, status, msg)

				if code == 0 {
					upe++
				}

				key := e.GetClass() + "." + e.GetName()

				p.metadata.LazySetValueUint64("count", key, e.GetExportCount())
				p.metadata.LazySetValueUint8("status", key, code)

				if msg != "" {
					if instance := p.metadata.GetInstance(key); instance != nil {
						instance.SetLabel("reason", msg)
					}
				}
			}

			// @TODO if there are no "master" exporters, don't collect metadata
			for _, e := range p.exporters {
				if err := e.Export(p.metadata); err != nil {
					logger.Error().Stack().Err(err).Msg("export component metadata:")
				}
				if err := e.Export(p.status); err != nil {
					logger.Error().Stack().Err(err).Msg("export target metadata:")
				}
			}

			// only zeroLog when numbers have changes, since hopefully that happens rarely
			if upc != upCollectors || upe != upExporters {
				logger.Info().Msgf("updated status, up collectors: %d (of %d), up exporters: %d (of %d)", upc, len(p.collectors), upe, len(p.exporters))
			}
			upCollectors = upc
			upExporters = upe
		}

		p.schedule.Sleep()
	}
}

// Stop gracefully exits the program by closing zeroLog
func (p *Poller) Stop() {
	logger.Info().Msgf("cleaning up and stopping [pid=%d]", os.Getpid())
}

// set up signal disposition
func (p *Poller) handleSignals(signalChannel chan os.Signal) {
	for {
		sig := <-signalChannel
		logger.Info().Msgf("caught signal [%s]", sig)
		p.Stop()
		os.Exit(0)
	}
}

// ping target system, report if it's available or not
// and if available, response time
func (p *Poller) ping() (float32, bool) {

	cmd := exec.Command("ping", p.target, "-w", "5", "-c", "1", "-q")

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
func (p *Poller) loadCollector(class, object string) error {

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
	if template, err = collector.ImportTemplate(p.options.HomePath, "default.yaml", class); err != nil {
		return err
	} else if template == nil { // probably redundant
		return errors.New(errors.MISSING_PARAM, "collector template")
	}

	if custom, err = collector.ImportTemplate(p.options.HomePath, "custom.yaml", class); err == nil && custom != nil {
		template.Merge(custom)
		logger.Debug().Msg("merged custom and default templates")
	}
	// add the poller's parameters to the collector's parameters
	Union2(template, p.params)

	// if we don't know object, try load from template
	if object == "" {
		object = template.GetChildContentS("object")
	}

	// if object is defined, we only initialize 1 sub-collector / object
	if object != "" {
		col, err = p.newCollector(class, object, template)
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
			if len(p.options.Objects) != 0 {
				ok = false
				for _, o := range p.options.Objects {
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

			col, err = p.newCollector(class, object.GetNameS(), template)
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

	p.collectors = append(p.collectors, collectors...)
	logger.Debug().Msgf("initialized (%s) with %d objects", class, len(collectors))
	// link each collector with requested exporter & update metadata
	for _, col = range collectors {
		if col == nil {
			logger.Warn().Msg("ignoring nil collector")
			continue
		}
		name := col.GetName()
		obj := col.GetObject()

		for _, expName := range col.WantedExporters(p.options.Config) {
			logger.Trace().Msgf("expName %s", expName)
			if exp := p.loadExporter(expName); exp != nil {
				col.LinkExporter(exp)
				logger.Debug().Msgf("linked (%s:%s) to exporter (%s)", name, obj, expName)
			} else {
				logger.Warn().Msgf("exporter (%s) requested by (%s:%s) not available", expName, name, obj)
			}
		}

		// update metadata

		if instance, err := p.metadata.NewInstance(name + "." + obj); err != nil {
			return err
		} else {
			instance.SetLabel("type", "collector")
			instance.SetLabel("name", name)
			instance.SetLabel("target", obj)
		}
	}

	return nil
}

// Union2 merges the fields of a Poller with the fields of a node.
// This is a way to bridge the struct world with the string typed world.
// If one of the poller field's does not exist in hNode, it will be copied
// from poller to hNode.
// If the field already exists in hNode, nothing is copied.
// Instead of comparing each field of the poller individually and being forced
// to keep this method in sync with the Poller struct, reflection via yaml marshaling
// is used to do the comparison. First the poller is marshaled to yaml and then
// unmarshalled into a list of generic yaml node. Each generic yaml node is walked, checking
// if there is a corresponding node in hNode, when there isn't one, a new hNode is created
// and populated with the yaml node's content. Finally, the new hNode is added to its parent

func Union2(hNode *node.Node, poller *conf.Poller) {
	marshal, err := yaml.Marshal(poller)
	if err != nil {
		return
	}
	root := yaml.Node{}
	err = yaml.Unmarshal(marshal, &root)
	if err != nil {
		return
	}
	rootContent := root.Content[0]
	if rootContent.Kind == yaml.MappingNode {
		for index, yNode := range rootContent.Content {
			// since rootContent is a mapping node every other yNode is a key
			if index%2 == 0 && yNode.Tag == "!!str" {
				// If the harvest node is missing this key, add it the harvest node
				if !hNode.HasChildS(yNode.Value) {
					// create a new harvest node to contain the missing content
					newNode := node.NewS(yNode.Value)
					// this is the value that goes along with the key from yNode
					valNode := rootContent.Content[index+1]
					//fmt.Printf("node type=%s val=%s %s\n", yNode.Value, valNode.Tag, valNode.Value)
					switch valNode.Tag {
					case "!!str", "!!bool":
						newNode.Content = []byte(valNode.Value)
					case "!!seq":
						// the poller node that's missing is a sequence so add all the children of the sequence
						for _, seqNode := range valNode.Content {
							newNode.NewChildS(seqNode.Value, seqNode.Value)
						}
					}
					hNode.AddChild(newNode)
				}
			}
		}
	}
}

func (p *Poller) newCollector(class string, object string, template *node.Node) (collector.Collector, error) {
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
	delegate := collector.New(class, object, p.options, template.Copy())
	err = col.Init(delegate)
	return col, err
}

// returns exporter that matches to name, if exporter is not loaded
// tries to load and return
func (p *Poller) loadExporter(name string) exporter.Exporter {

	var (
		err    error
		class  string
		params *node.Node
		exp    exporter.Exporter
	)

	// stop here if exporter is already loaded
	if exp = p.getExporter(name); exp != nil {
		return exp
	}

	if params = p.exporterParams.GetChildS(name); params == nil {
		logger.Warn().Msgf("exporter (%s) not defined in config", name)
		return nil
	}

	if class = params.GetChildContentS("exporter"); class == "" {
		logger.Warn().Msgf("exporter (%s) has no exporter class defined", name)
		return nil
	}

	absExp := exporter.New(class, name, p.options, params)
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

	p.exporters = append(p.exporters, exp)
	logger.Debug().Msgf("initialized exporter (%s)", name)

	// update metadata
	if instance, err := p.metadata.NewInstance(exp.GetClass() + "." + exp.GetName()); err != nil {
		logger.Error().Msgf("add metadata instance: %v", err)
	} else {
		instance.SetLabel("type", "exporter")
		instance.SetLabel("name", exp.GetClass())
		instance.SetLabel("target", exp.GetName())
	}
	return exp

}

func (p *Poller) getExporter(name string) exporter.Exporter {
	for _, exp := range p.exporters {
		if exp.GetName() == name {
			return exp
		}
	}
	return nil
}

// initialize matrices to be used as metadata
func (p *Poller) loadMetadata() {

	p.metadata = matrix.New("poller", "metadata_component")
	p.metadata.NewMetricUint8("status")
	p.metadata.NewMetricUint64("count")
	p.metadata.SetGlobalLabel("poller", p.name)
	p.metadata.SetGlobalLabel("version", p.options.Version)
	p.metadata.SetGlobalLabel("hostname", p.options.Hostname)
	p.metadata.SetExportOptions(matrix.DefaultExportOptions())

	// metadata for target system
	p.status = matrix.New("poller", "metadata_target")
	p.status.NewMetricUint8("status")
	p.status.NewMetricFloat32("ping")
	p.status.NewMetricUint32("goroutines")

	instance, _ := p.status.NewInstance("host")
	instance.SetLabel("addr", p.target)
	p.status.SetGlobalLabel("poller", p.name)
	p.status.SetGlobalLabel("version", p.options.Version)
	p.status.SetGlobalLabel("hostname", p.options.Hostname)
	p.status.SetExportOptions(matrix.DefaultExportOptions())
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
	flags.IntVar(&args.PromPort, "promPort", 0, "Prometheus Port")
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
