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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	goversion "github.com/hashicorp/go-version"
	_ "github.com/netapp/harvest/v2/cmd/collectors/ems"
	"github.com/netapp/harvest/v2/cmd/collectors/rest"
	_ "github.com/netapp/harvest/v2/cmd/collectors/restperf"
	_ "github.com/netapp/harvest/v2/cmd/collectors/simple"
	_ "github.com/netapp/harvest/v2/cmd/collectors/storagegrid"
	_ "github.com/netapp/harvest/v2/cmd/collectors/unix"
	_ "github.com/netapp/harvest/v2/cmd/collectors/zapi/collector"
	_ "github.com/netapp/harvest/v2/cmd/collectors/zapiperf"
	"github.com/netapp/harvest/v2/cmd/exporters/influxdb"
	"github.com/netapp/harvest/v2/cmd/exporters/prometheus"
	"github.com/netapp/harvest/v2/cmd/harvest/version"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/poller/schedule"
	rest2 "github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"math"
	"net/http"
	_ "net/http/pprof" // #nosec since pprof is off by default
	"os"
	"os/exec"
	"os/signal"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// default params
var (
	pollerSchedule   = "60s"
	logFileName      = ""
	logMaxMegaBytes  = logging.DefaultLogMaxMegaBytes
	logMaxBackups    = logging.DefaultLogMaxBackups
	logMaxAge        = logging.DefaultLogMaxAge
	asupSchedule     = "24h" // send every 24 hours
	asupFirstWrite   = "4m"  // after this time, write 1st autosupport payload (for testing)
	isOntapCollector = map[string]struct{}{
		"ZapiPerf": {},
		"Zapi":     {},
		"Rest":     {},
		"RestPerf": {},
		"Ems":      {},
	}
)

// init with default configuration that logs to both console and harvest.log
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

var pingRegex = regexp.MustCompile(` = (.*?)/`)

// Poller is the instance that starts and monitors a
// group of collectors and exporters as a single UNIX process
type Poller struct {
	name            string
	target          string
	options         *options.Options
	schedule        *schedule.Schedule
	collectors      []collector.Collector
	exporters       []exporter.Exporter
	exporterParams  map[string]conf.Exporter
	params          *conf.Poller
	metadata        *matrix.Matrix
	status          *matrix.Matrix
	certPool        *x509.CertPool
	client          *http.Client
	hasPromExporter bool
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
		fileLoggingEnabled = true
	} else {
		consoleLoggingEnabled = !p.options.LogToFile
		fileLoggingEnabled = p.options.LogToFile
	}
	if fileLoggingEnabled {
		logFileName = "poller_" + p.name + ".log"
	}

	err = conf.LoadHarvestConfig(p.options.Config)
	if err != nil {
		// separate logger is not yet configured as it depends on setting logMaxMegaBytes, logMaxFiles later
		// Using default instance of logger which logs below error to harvest.log
		logging.Get().SubLogger("Poller", p.name).Error().
			Str("config", p.options.Config).Err(err).Msg("Unable to read config")
		return err
	}
	p.params, err = conf.PollerNamed(p.name)
	if err != nil {
		logging.Get().SubLogger("Poller", p.name).Error().
			Str("config", p.options.Config).Err(err).Msg("Failed to find poller")
		return err
	}

	// log handling parameters
	// size of file before rotating
	if p.params.LogMaxBytes != 0 {
		logMaxMegaBytes = int(p.params.LogMaxBytes / (1024 * 1024))
	}

	// maximum number of rotated files to keep
	if p.params.LogMaxFiles != 0 {
		logMaxBackups = p.params.LogMaxFiles
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
	logger.Info().
		Str("logLevel", zeroLogLevel.String()).
		Str("configPath", p.options.Config).
		Str("version", version.String()).
		Msg("Init")

	// if profiling port > 0 start profiling service
	if p.options.Profiling > 0 {
		addr := fmt.Sprintf("localhost:%d", p.options.Profiling)
		logger.Info().Msgf("profiling enabled on [%s]", addr)
		go func() {
			fmt.Println(http.ListenAndServe(addr, nil)) //nolint:gosec
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

	if conf.Config.Admin.Httpsd.TLS.CertFile != "" {
		util.CheckCert(conf.Config.Admin.Httpsd.TLS.CertFile, "ssl_cert", p.options.Config, *logger.Logger)
		cert, err := os.ReadFile(conf.Config.Admin.Httpsd.TLS.CertFile)
		if err != nil {
			logger.Fatal().Str("certFile", conf.Config.Admin.Httpsd.TLS.CertFile).Msg("Unable to read cert file")
		}
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(cert); !ok {
			logger.Fatal().Str("certFile", conf.Config.Admin.Httpsd.TLS.CertFile).Msg("Unable to parse cert")
		}
		p.certPool = certPool
	}
	// announce startup
	if p.options.Daemon {
		logger.Info().Int("pid", os.Getpid()).Msg("started as daemon")
	} else {
		logger.Info().Int("pid", os.Getpid()).Msg("started in foreground")
	}

	// each poller is associated with a remote host
	// if no address is specified, assume that is local host
	if p.params.Addr == "" {
		p.target = "localhost"
	} else {
		p.target = p.params.Addr
	}
	// check optional parameter auth_style
	// if certificates are missing use default paths
	if p.params.AuthStyle == "certificate_auth" {
		if p.params.SslCert == "" {
			fp := path.Join(p.options.HomePath, "cert/", p.options.Hostname+".pem")
			p.params.SslCert = fp
			logger.Debug().Msgf("using default [ssl_cert] path: [%s]", fp)
			if _, err = os.Stat(fp); err != nil {
				logger.Error().Stack().Err(err).Msgf("ssl_cert")
				return errs.New(errs.ErrMissingParam, "ssl_cert: "+err.Error())
			}
		}
		if p.params.SslKey == "" {
			fp := path.Join(p.options.HomePath, "cert/", p.options.Hostname+".key")
			p.params.SslKey = fp
			logger.Debug().Msgf("using default [ssl_key] path: [%s]", fp)
			if _, err = os.Stat(fp); err != nil {
				logger.Error().Stack().Err(err).Msgf("ssl_key")
				return errs.New(errs.ErrMissingParam, "ssl_key: "+err.Error())
			}
		}
	}

	// initialize our metadata, the metadata will host status of our
	// collectors and exporters, as well as ping stats to target host
	p.loadMetadata()
	p.exporterParams = conf.Config.Exporters

	// iterate over the list of collectors and initialize them
	// exporters are initialized on the fly when at least one collector references them

	filteredCollectors := p.params.Collectors
	// If the customer requested a specific collector, use it
	if len(p.options.Collectors) > 0 {
		filteredCollectors = make([]conf.Collector, 0, len(p.options.Collectors))
		for _, collectorName := range p.options.Collectors {
			filteredCollectors = append(filteredCollectors, conf.NewCollector(collectorName))
		}
	}
	if len(filteredCollectors) == 0 {
		logger.Warn().Msg("no collectors defined for this poller in config or CLI")
		return errs.New(errs.ErrNoCollector, "no collectors")
	}

	for _, c := range filteredCollectors {
		if err = p.loadCollector(c); err != nil {
			logger.Error().Stack().Err(err).Msgf("load collector (%s) templates=%s:", c.Name, *c.Templates)
		}
	}

	// at least one collector should successfully initialize
	if len(p.collectors) == 0 {
		logger.Warn().Msg("no collectors initialized, stopping")
		return errs.New(errs.ErrNoCollector, "no collectors")
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
	if p.params.PollerSchedule != "" {
		pollerSchedule = p.params.PollerSchedule
	}
	p.schedule = schedule.New()
	if err = p.schedule.NewTaskString("poller", pollerSchedule, nil, true, "poller_"+p.name); err != nil {
		logger.Error().Stack().Err(err).Msg("set schedule:")
		return err
	}
	logger.Debug().Msgf("set poller schedule with %s frequency", pollerSchedule)

	// Check if autosupport is enabled
	tools := conf.Config.Tools
	if tools != nil && tools.AsupDisabled {
		logger.Info().Msgf("Autosupport is disabled")
	} else {
		if p.targetIsOntap() {
			// Write the payload after asupFirstWrite.
			// This is to examine the autosupport contents
			// Nothing is sent, sending happens based on the asupSchedule
			duration, err := time.ParseDuration(asupFirstWrite)
			if err != nil {
				logger.Error().Err(err).
					Str("asupFirstWrite", asupFirstWrite).
					Msg("Failed to write 1st autosupport payload.")
			} else {
				time.AfterFunc(duration, func() {
					p.firstAutoSupport()
				})
			}
			if err = p.schedule.NewTaskString("asup", asupSchedule, p.startAsup, p.options.Asup, "asup_"+p.name); err != nil {
				return err
			}
			logger.Info().
				Str("asupSchedule", asupSchedule).
				Msg("Autosupport scheduled.")
		} else {
			logger.Info().
				Str("poller", p.name).
				Msg("Autosupport disabled since poller not connected to ONTAP.")
		}
	}

	// famous last words
	logger.Info().Msg("poller start-up complete")

	return nil

}

func (p *Poller) firstAutoSupport() {
	if p.collectors == nil {
		return
	}
	if _, err := collector.BuildAndWriteAutoSupport(p.collectors, p.status, p.name); err != nil {
		logger.Error().Err(err).
			Str("poller", p.name).
			Msg("First autosupport failed.")
	}
}

func (p *Poller) startAsup() (map[string]*matrix.Matrix, error) {
	if p.collectors != nil {
		if err := collector.SendAutosupport(p.collectors, p.status, p.name); err != nil {
			logger.Error().Err(err).
				Str("poller", p.name).
				Msg("Start autosupport failed.")
			return nil, err
		}
	}
	return nil, nil
}

// Start will run the collectors and the poller itself
// in separate goroutines, leaving the main goroutine
// to the exporters
func (p *Poller) Start() {

	var (
		wg  sync.WaitGroup
		col collector.Collector
	)

	go p.startHeartBeat()

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

	// poller schedule has the poller and asup task (when enabled)
	task := p.schedule.GetTask("poller")
	asuptask := p.schedule.GetTask("asup")

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
				_ = p.status.LazySetValueUint8("status", "host", 0)
				_ = p.status.LazySetValueFloat64("ping", "host", float64(ping))
			} else {
				_ = p.status.LazySetValueUint8("status", "host", 1)
			}

			// add number of goroutines to metadata
			// @TODO: cleanup, does not belong to "status"
			_ = p.status.LazySetValueInt64("goroutines", "host", int64(runtime.NumGoroutine()))

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

				_ = p.metadata.LazySetValueUint64("count", key, c.GetCollectCount())
				_ = p.metadata.LazySetValueUint8("status", key, code)

				if msg != "" {
					if instance := p.metadata.GetInstance(key); instance != nil {
						// replace quotes with empty, in case of rest error may have quotes around endpoint which fails prometheus discovery
						instance.SetLabel("reason", strings.ReplaceAll(msg, "\"", ""))
					}
				}
			}

			// update status of exporters
			for _, ee := range p.exporters {
				code, status, msg := ee.GetStatus()
				logger.Debug().Msgf("exporter (%s) status: (%d - %s) %s", ee.GetName(), code, status, msg)

				if code == 0 {
					upe++
				}

				key := ee.GetClass() + "." + ee.GetName()

				_ = p.metadata.LazySetValueUint64("count", key, ee.GetExportCount())
				_ = p.metadata.LazySetValueUint8("status", key, code)

				if msg != "" {
					if instance := p.metadata.GetInstance(key); instance != nil {
						instance.SetLabel("reason", msg)
					}
				}
			}

			// @TODO if there are no "master" exporters, don't collect metadata
			for _, ee := range p.exporters {
				if err := ee.Export(p.metadata); err != nil {
					logger.Error().Stack().Err(err).Msg("export component metadata:")
				}
				if err := ee.Export(p.status); err != nil {
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

		// asup task will be nil when autosupport is disabled
		if asuptask != nil && asuptask.IsDue() {
			_, _ = asuptask.Run()
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

	cmd := exec.Command("ping", p.target, "-w", "5", "-c", "1", "-q") //nolint:gosec
	output, err := cmd.Output()
	if err != nil {
		return 0, false
	}
	return p.parsePing(string(output))
}

func (p *Poller) parsePing(out string) (float32, bool) {
	if strings.Contains(out, "min/avg/max") {
		match := pingRegex.FindStringSubmatch(out)
		if len(match) > 0 {
			if p, err := strconv.ParseFloat(match[1], 32); err == nil {
				return float32(p), true
			}
		}
	}
	return 0, false
}

// dynamically load and initialize a collector
// if there are more than one objects defined for a collector,
// then multiple collectors will be initialized
func (p *Poller) loadCollector(c conf.Collector) error {

	var (
		class                 string
		err                   error
		template, subTemplate *node.Node
		collectors            []collector.Collector
		col                   collector.Collector
	)

	if c, err = p.upgradeCollector(c); err != nil {
		return err
	}
	class = c.Name
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
	if c.Templates != nil {
		for _, t := range *c.Templates {
			if subTemplate, err = collector.ImportTemplate(p.options.HomePath, t, class); err != nil {
				logEvent := logger.Warn()
				if t == "custom.yaml" {
					// make this less noisy since it won't exist for most people
					logEvent = logger.Debug()
				}
				logEvent.
					Str("err", err.Error()).
					Msg("Unable to load template.")
				continue
			}
			if template == nil {
				template = subTemplate
			} else {
				logger.Debug().
					Str("template", t).
					Msg("Merged template.")
				if c.Name == "Zapi" || c.Name == "ZapiPerf" {
					// do not overwrite child of objects. They will be concatenated
					template.Merge(subTemplate, []string{"objects"})
				} else {
					template.Merge(subTemplate, []string{""})
				}
			}
		}
	}
	if template == nil {
		return fmt.Errorf("no templates loaded for %s", c.Name)
	}
	// add the poller's parameters to the collector's parameters
	Union2(template, p.params)
	template.NewChildS("poller_name", p.params.Name)

	objects := make([]string, 0)
	templateObject := template.GetChildContentS("object")

	// if `objects` was passed at the cmdline, use them instead of the defaults
	if len(p.options.Objects) != 0 {
		objects = append(objects, p.options.Objects...)
	} else if templateObject != "" {
		// if object is defined, we only initialize 1 sub-collector / object
		objects = append(objects, templateObject)
		// if template has list of objects, initialize 1 sub-collector for each
	} else if templateObjects := template.GetChildS("objects"); templateObjects != nil {
		for _, object := range templateObjects.GetChildren() {
			objects = append(objects, object.GetNameS())
		}
	} else {
		return errs.New(errs.ErrMissingParam, "collector object")
	}

	for _, object := range objects {
		col, err = p.newCollector(class, object, template)
		if err != nil {
			if errors.Is(err, errs.ErrConnection) {
				logger.Warn().
					Str("collector", class).
					Str("object", object).
					Msg("abort collector")
				break
			}
			logger.Warn().Err(err).
				Str("collector", class).
				Str("object", object).
				Msg("init collector-object")
		} else {
			collectors = append(collectors, col)
			logger.Debug().
				Str("collector", class).
				Str("object", object).
				Msg("initialized collector-object")
		}
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

		for _, expName := range col.WantedExporters(p.params.Exporters) {
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
							if seqNode.Tag == "!!str" {
								newNode.NewChildS(seqNode.Value, seqNode.Value)
							} else if seqNode.Tag == "!!map" {
								for ci := 0; ci < len(seqNode.Content); ci += 2 {
									newNode.NewChildS(seqNode.Content[ci].Value, seqNode.Content[ci+1].Value)
								}
							}
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
		return nil, fmt.Errorf("error getting module %s err: %w", name, err)
	}
	inst := mod.New()
	col, ok := inst.(collector.Collector)
	if !ok {
		return nil, errs.New(errs.ErrNoCollector, "no collectors")
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
		params conf.Exporter
		exp    exporter.Exporter
	)

	// stop here if exporter is already loaded
	if exp = p.getExporter(name); exp != nil {
		return exp
	}

	params, ok := p.exporterParams[name]
	if !ok {
		logger.Warn().Msgf("exporter (%s) not defined in config", name)
		return nil
	}

	if class = params.Type; class == "" {
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
		logger.Error().Err(err).Str("name", name).Msg("Unable to init exporter")
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

	p.metadata = matrix.New("poller", "metadata_component", "metadata_component")
	_, _ = p.metadata.NewMetricUint8("status")
	_, _ = p.metadata.NewMetricUint64("count")
	p.metadata.SetGlobalLabel("poller", p.name)
	p.metadata.SetGlobalLabel("version", p.options.Version)
	p.metadata.SetGlobalLabel("hostname", p.options.Hostname)
	if p.options.PromPort != 0 {
		p.metadata.SetGlobalLabel("promport", strconv.Itoa(p.options.PromPort))
	}
	p.metadata.SetExportOptions(matrix.DefaultExportOptions())

	// metadata for target system
	p.status = matrix.New("poller", "metadata_target", "metadata_component")
	_, _ = p.status.NewMetricUint8("status")
	_, _ = p.status.NewMetricFloat64("ping")
	_, _ = p.status.NewMetricUint64("goroutines")

	instance, _ := p.status.NewInstance("host")
	instance.SetLabel("addr", p.target)
	p.status.SetGlobalLabel("poller", p.name)
	p.status.SetGlobalLabel("version", p.options.Version)
	p.status.SetGlobalLabel("hostname", p.options.Hostname)
	if p.options.PromPort != 0 {
		p.status.SetGlobalLabel("promport", strconv.Itoa(p.options.PromPort))
	}
	p.status.SetExportOptions(matrix.DefaultExportOptions())
}

var pollerCmd = &cobra.Command{
	Use:   "poller -p name [flags]",
	Short: "Harvest Poller - Runs collectors and exporters for a target system",
	Args:  cobra.NoArgs,
	Run:   startPoller,
}

// Returns true if at least one collector is known
// to collect from an Ontap system (needs to be updated
// when we add other Ontap collectors, e.g. REST)

func (p *Poller) targetIsOntap() bool {
	for _, c := range p.collectors {
		_, ok := isOntapCollector[c.GetName()]
		if ok {
			return true
		}
	}
	return false
}

type pollerDetails struct {
	Name string `json:"Name,omitempty"`
	IP   string `json:"IP,omitempty"`
	Port int    `json:"Port,omitempty"`
}

func (p *Poller) publishDetails() {
	localIP, err := util.FindLocalIP()
	if err != nil {
		logger.Err(err).Msg("Unable to find local IP")
		return
	}
	if p.client == nil {
		return
	}
	exporterIP := "127.0.0.1"
	heartBeatURL := ""
	for _, exporterName := range p.params.Exporters {
		exp, ok := p.exporterParams[exporterName]
		if !ok {
			continue
		}
		if exp.Type != "Prometheus" {
			continue
		}
		p.hasPromExporter = true
		if exp.LocalHTTPAddr == "" || exp.LocalHTTPAddr == "0.0.0.0" {
			exporterIP = localIP
		} else {
			exporterIP = exp.LocalHTTPAddr
		}
		if exp.HeartBeatURL != "" {
			heartBeatURL = exp.HeartBeatURL
		}
	}

	if !p.hasPromExporter {
		// no prometheus exporter, don't publish details
		return
	}

	details := pollerDetails{
		Name: p.name,
		IP:   exporterIP,
		Port: p.options.PromPort,
	}
	payload, err := json.Marshal(details)
	if err != nil {
		logger.Error().Err(err).Str("poller", p.name).Msg("Unable to marshal poller details")
		return
	}
	defaultURL := p.makePublishURL()

	if heartBeatURL == "" {
		heartBeatURL = defaultURL
	}
	req, err := http.NewRequest("PUT", heartBeatURL, bytes.NewBuffer(payload))
	if err != nil {
		logger.Err(err).Msg("failed to connect to admin")
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	user := conf.Config.Admin.Httpsd.AuthBasic.Username
	if user != "" {
		req.SetBasicAuth(user, conf.Config.Admin.Httpsd.AuthBasic.Password)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		rErr := errors.Unwrap(err)
		if rErr == nil {
			rErr = err
		}
		// check if this is a connection error, if so, the admin node is down
		// log as warning instead of error
		event := logger.Error()
		if strings.Contains(rErr.Error(), "connection refused") {
			event = logger.Warn()
		}
		event.Err(rErr).Str("admin", conf.Config.Admin.Httpsd.Listen).Msg("Failed connecting to admin node")
		return
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Err(err).Msg("failed to read publishDetails response to admin")
		return
	}
	p.client.CloseIdleConnections()
	if resp.StatusCode != 200 {
		txt := string(body)
		txt = txt[0:int(math.Min(float64(len(txt)), 48))]
		logger.Error().
			Str("admin", conf.Config.Admin.Httpsd.Listen).
			Str("body", txt).
			Int("httpStatusCode", resp.StatusCode).
			Msg("Admin node problem")
	}
}

// startHeartBeat never returns unless the receiver does not have a Prometheus exporter
// Publish the receiver's discovery details to the admin node
func (p *Poller) startHeartBeat() {
	if conf.Config.Admin.Httpsd.Listen == "" {
		return
	}
	p.createClient()
	p.publishDetails()
	if !p.hasPromExporter {
		return
	}
	if conf.Config.Admin.Httpsd.HeartBeat == "" {
		conf.Config.Admin.Httpsd.HeartBeat = "45s"
	}
	duration, err := time.ParseDuration(conf.Config.Admin.Httpsd.HeartBeat)
	if err != nil {
		logger.Warn().Str("heart_beat", conf.Config.Admin.Httpsd.HeartBeat).
			Err(err).Msg("Invalid heart_beat using 1m")
		duration = 1 * time.Minute
	}
	tick := time.Tick(duration)
	for range tick {
		p.publishDetails()
	}
}

func (p *Poller) makePublishURL() string {
	// Listen will be one of: localhost:port, :port, ip:port
	schema := "http"
	if conf.Config.Admin.Httpsd.TLS.CertFile != "" {
		schema = "https"
	}
	if strings.HasPrefix(conf.Config.Admin.Httpsd.Listen, ":") {
		return fmt.Sprintf("%s://127.0.0.1:%s/api/v1/sd", schema, conf.Config.Admin.Httpsd.Listen[1:])
	}
	return fmt.Sprintf("%s://%s/api/v1/sd", schema, conf.Config.Admin.Httpsd.Listen)
}

func (p *Poller) createClient() {
	if conf.Config.Admin.Httpsd.TLS.CertFile != "" {
		p.client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:    p.certPool,
					MinVersion: tls.VersionTLS13,
				},
			},
		}
	} else {
		p.client = &http.Client{
			Transport: &http.Transport{},
		}
	}
}

// upgradeCollector checks if the collector c should be upgraded to a REST collector.
// If an upgrade is possible, a new collector will be returned, otherwise the original will be.
//
// ZAPI collectors should be upgraded to REST collectors when the ONTAP version is >= 9.12.1
// The check is performed by:
//   - use the REST API to query the cluster version
//   - compare the cluster version with the upgradeAfter version
//
// If any error happens during the REST query, the upgrade is aborted and the original collector c returned.
func (p *Poller) upgradeCollector(c conf.Collector) (conf.Collector, error) {
	// Only attempt to upgrade Zapi* collectors
	if !strings.HasPrefix(c.Name, "Zapi") {
		return c, nil
	}
	noUpgrade := os.Getenv("HARVEST_NO_COLLECTOR_UPGRADE")
	if noUpgrade != "" {
		logger.Debug().Str("collector", c.Name).Msg("No upgrade due to env var. Use collector")
		return c, nil
	}
	r, err := p.newRestClient()
	if err != nil {
		logger.Debug().Err(err).Str("collector", c.Name).Msg("Failed to upgrade to Rest. Use collector")
		return c, nil
	}
	ver := r.Client.Cluster().Version
	verWithDots := fmt.Sprintf("%d.%d.%d", ver[0], ver[1], ver[2])
	ontapVersion, err2 := goversion.NewVersion(verWithDots)
	if err2 != nil {
		logger.Error().Err(err2).
			Str("version", verWithDots).
			Str("collector", c.Name).
			Msg("Failed to parse version")
		return c, nil
	}
	upgradeVersion := "9.12.1"
	upgradeAfter, err3 := goversion.NewVersion(upgradeVersion)
	if err3 != nil {
		logger.Error().Err(err3).
			Str("upgradeVersion", upgradeVersion).
			Str("collector", c.Name).
			Msg("Failed to parse upgradeVersion")
		return c, nil
	}

	if ontapVersion.GreaterThanOrEqual(upgradeAfter) {
		upgradeCollector := strings.ReplaceAll(c.Name, "Zapi", "Rest")
		logger.Info().
			Str("from", c.Name).
			Str("to", upgradeCollector).
			Str("v", verWithDots).
			Str("upgradeVersion", upgradeVersion).
			Msg("Upgrade collector")
		return conf.Collector{
			Name:      upgradeCollector,
			Templates: c.Templates,
		}, nil
	}
	logger.Debug().
		Str("collector", c.Name).
		Str("v", verWithDots).
		Str("upgradeVersion", upgradeVersion).
		Msg("Do not upgrade collector")
	return c, nil
}

func (p *Poller) newRestClient() (*rest.Rest, error) {
	params := node.NewS("")
	// Set client_timeout to suppress logging a msg about the default client_timeout during Rest client creation
	params.NewChildS("client_timeout", rest2.DefaultTimeout)
	Union2(params, p.params)
	delegate := collector.New("Rest", "", p.options, params)
	r := &rest.Rest{
		AbstractCollector: delegate,
	}
	err := r.InitClient()
	return r, err
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
	configPath := conf.GetDefaultHarvestConfigPath()

	var flags = pollerCmd.Flags()
	flags.StringVarP(&args.Poller, "poller", "p", "", "Poller name as defined in config")
	flags.BoolVarP(&args.Debug, "debug", "d", false, "Debug mode, no data will be exported")
	flags.BoolVar(&args.Daemon, "daemon", false, "Start as daemon")
	flags.IntVarP(&args.LogLevel, "loglevel", "l", 2, "Logging level (0=trace, 1=debug, 2=info, 3=warning, 4=error, 5=critical)")
	flags.BoolVar(&args.LogToFile, "logtofile", false, "When running in the foreground, log to file instead of stdout")
	flags.IntVar(&args.Profiling, "profiling", 0, "If profiling port > 0, enables profiling via localhost:PORT/debug/pprof/")
	flags.IntVar(&args.PromPort, "promPort", 0, "Prometheus Port")
	flags.StringVar(&args.Config, "config", configPath, "Harvest config file path")
	flags.StringSliceVarP(&args.Collectors, "collectors", "c", []string{}, "Only start these collectors (overrides harvest.yml)")
	flags.StringSliceVarP(&args.Objects, "objects", "o", []string{}, "Only start these objects (overrides collector config)")

	// Used to test autosupport at startup. An environment variable is used instead of a cmdline
	// arg, so we don't have to also add this testing arg to harvest cli
	if isAsup := os.Getenv("ASUP"); isAsup != "" {
		args.Asup = true
	}

	_ = pollerCmd.MarkFlagRequired("poller")
	_ = pollerCmd.Flags().MarkHidden("logtofile")
}

// start poller, if fails try to write to syslog
func main() {
	// don't recover if a goroutine has panicked, instead
	// log as much as possible
	defer func() {
		if r := recover(); r != nil {
			e := r.(error)
			logger.Error().Stack().Err(e).Msg("Poller panicked")
			logger.Fatal().Msg(`(main) terminating abnormally, tip: run in foreground mode (with "--loglevel 0") to debug`)
		}
	}()

	cobra.CheckErr(pollerCmd.Execute())
}
