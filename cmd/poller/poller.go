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
	"context"
	"crypto/fips140"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/exporters/victoriametrics"
	"io"
	"log/slog"
	"math"
	"net"
	"net/http"
	_ "net/http/pprof" // #nosec since pprof is off by default
	"os"
	"os/signal"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/netapp/harvest/v2/cmd/collectors"
	_ "github.com/netapp/harvest/v2/cmd/collectors/cisco"
	_ "github.com/netapp/harvest/v2/cmd/collectors/ems"
	_ "github.com/netapp/harvest/v2/cmd/collectors/eseries"
	_ "github.com/netapp/harvest/v2/cmd/collectors/eseriesperf"
	_ "github.com/netapp/harvest/v2/cmd/collectors/keyperf"
	_ "github.com/netapp/harvest/v2/cmd/collectors/restperf"
	_ "github.com/netapp/harvest/v2/cmd/collectors/simple"
	_ "github.com/netapp/harvest/v2/cmd/collectors/statperf"
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
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	version2 "github.com/netapp/harvest/v2/pkg/version"
	goversion "github.com/netapp/harvest/v2/third_party/go-version"
	"github.com/spf13/cobra"
)

// default params
var (
	pollerSchedule    = "1m"
	pollerLogSchedule = "1h"
	logFileName       = ""
	logMaxMegaBytes   = logging.DefaultLogMaxMegaBytes
	logMaxBackups     = logging.DefaultLogMaxBackups
	logMaxAge         = logging.DefaultLogMaxAge
	asupSchedule      = "24h" // send every 24 hours
	asupFirstWrite    = "4m"  // after this time, write 1st autosupport payload (for testing)
	opts              *options.Options
)

// init with default configuration that logs to the console and harvest.log
var logger = slog.Default()

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
	name                 string
	target               string
	options              *options.Options
	schedule             *schedule.Schedule
	collectors           []collector.Collector
	exporters            []exporter.Exporter
	exporterParams       map[string]conf.Exporter
	params               *conf.Poller
	metadata             *matrix.Matrix
	metadataTarget       *matrix.Matrix // exported as metadata_target_
	status               *matrix.Matrix // exported as poller_status
	certPool             *x509.CertPool
	client               *http.Client
	auth                 *auth.Credentials
	hasPromExporter      bool
	maxRssBytes          uint64
	startTime            time.Time
	remote               conf.Remote
	concurrentCollectors *atomic.Int32 // tracks the number of currently active collector tasks
}

// Init starts Poller, reads parameters, opens zeroLog handler, initializes metadata,
// starts collectors and exporters
func (p *Poller) Init() error {

	var (
		err                   error
		fileLoggingEnabled    bool
		consoleLoggingEnabled bool
		configPath            string
	)

	p.startTime = time.Now()
	p.options = opts.SetDefaults()
	p.name = opts.Poller
	p.concurrentCollectors = &atomic.Int32{}

	logLevel := logging.GetLogLevel(p.options.LogLevel)
	// if we are a daemon, use file logging
	if p.options.Daemon {
		fileLoggingEnabled = true
	} else {
		consoleLoggingEnabled = !p.options.LogToFile
		fileLoggingEnabled = p.options.LogToFile
	}
	if fileLoggingEnabled {
		logFileName = "poller_" + p.name + ".log"
	}

	configPath, err = conf.LoadHarvestConfig(p.options.Config)
	if err != nil {
		// separate logger is not yet configured as it depends on setting logMaxMegaBytes, logMaxFiles later
		// Using default instance of logger which logs below error to harvest.log
		slog.Default().With(slog.String("Poller", p.name)).Error(
			"Unable to read config",
			slogx.Err(err),
			slog.String("config", p.options.Config),
			slog.String("configPath", configPath),
		)
		return err
	}
	p.params, err = conf.PollerNamed(p.name)
	if err != nil {
		slog.Default().With(slog.String("Poller", p.name)).Error(
			"Failed to find poller",
			slogx.Err(err),
			slog.String("config", p.options.Config),
			slog.String("configPath", configPath),
		)
		return err
	}

	p.mergeConfPath()

	// log handling parameters
	// size of file before rotating
	if p.params.LogMaxBytes != 0 {
		logMaxMegaBytes = int(p.params.LogMaxBytes / (1024 * 1024))
	}

	// maximum number of rotated files to keep
	if p.params.LogMaxFiles != 0 {
		logMaxBackups = p.params.LogMaxFiles
	}

	logConfig := logging.LogConfig{
		ConsoleLoggingEnabled: consoleLoggingEnabled,
		PrefixKey:             "Poller",
		PrefixValue:           p.name,
		LogLevel:              logLevel,
		LogFormat:             p.options.LogFormat,
		FileLoggingEnabled:    fileLoggingEnabled,
		Directory:             p.options.LogPath,
		Filename:              logFileName,
		MaxSize:               logMaxMegaBytes,
		MaxBackups:            logMaxBackups,
		MaxAge:                logMaxAge,
	}

	logger = logging.Configure(logConfig)

	// If profiling port > 0, start an HTTP server on that port with the profiling endpoints setup.
	// When using the Prometheus exporter, the profiling endpoints will be setup automatically in
	// cmd/exporters/prometheus/httpd.go
	if p.options.Profiling > 0 {
		addr := fmt.Sprintf("localhost:%d", p.options.Profiling)
		slog.Info("profiling enabled", slog.String("addr", addr))
		go func() {
			fmt.Println(http.ListenAndServe(addr, nil)) //nolint:gosec
		}()
	}

	getwd, err := os.Getwd()
	if err != nil {
		slog.Error("Unable to get current working directory", slogx.Err(err))
		getwd = ""
	}
	slog.Info("Init",
		slog.String("logLevel", logLevel.String()),
		slog.String("configPath", configPath),
		slog.String("cwd", getwd),
		slog.String("version", strings.TrimSpace(version.String())),
		slog.Bool("fips", fips140.Enabled()),
		slog.Any("options", p.options),
	)

	// set signal handler for graceful termination
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, SIGNALS...)
	go p.handleSignals(signalChannel)

	if conf.Config.Admin.Httpsd.TLS.CertFile != "" {
		requests.CheckCert(conf.Config.Admin.Httpsd.TLS.CertFile, "ssl_cert", p.options.Config, slog.Default())
		cert, err := os.ReadFile(conf.Config.Admin.Httpsd.TLS.CertFile)
		if err != nil {
			slog.Error(
				"Unable to read cert file",
				slogx.Err(err),
				slog.String("certFile", conf.Config.Admin.Httpsd.TLS.CertFile),
			)
			os.Exit(1)
		}
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(cert); !ok {
			slog.Error(
				"Unable to parse cert file",
				slog.String("certFile", conf.Config.Admin.Httpsd.TLS.CertFile),
			)
			os.Exit(1)
		}
		p.certPool = certPool
	}
	// announce startup
	if p.options.Daemon {
		slog.Info("started as daemon", slog.Int("pid", os.Getpid()))
	} else {
		slog.Info("started in foreground", slog.Int("pid", os.Getpid()))
	}

	// each poller is associated with a remote host
	// if no address is specified, assume localhost
	if p.params.Addr == "" {
		p.target = "localhost"
	} else {
		p.target = p.params.Addr
	}

	// create a shared auth service that all collectors will use
	p.auth = auth.NewCredentials(p.params, logger)

	// initialize our metadata, the metadata will host the status of our
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
		slog.Warn("no collectors defined for this poller in config or CLI")
		return errs.New(errs.ErrNoCollector, "no collectors")
	}

	filteredCollectors = p.negotiateAPI(filteredCollectors)

	objectsToCollectors := make(map[string][]objectCollector)
	for _, c := range filteredCollectors {
		_, ok := conf.IsCollector[c.Name]
		if !ok {
			valid := strings.Join(conf.GetCollectorSlice(), ", ")
			slog.Error("Valid collectors are: "+valid, slog.String("Detected invalid collector", c.Name))
			continue
		}
		objects, err := p.readObjects(c)
		if err != nil {
			slog.Error(
				"Failed to read objects",
				slogx.Err(err),
				slog.String("collector", c.Name),
				slog.String("templates", strings.Join(*c.Templates, ",")),
				slog.String("error", err.Error()),
			)
			continue
		}
		for _, oc := range objects {
			upgradedOC := p.upgradeObjectCollector(oc)
			objectsToCollectors[oc.object] = append(objectsToCollectors[oc.object], upgradedOC)
		}
	}

	// for each object, only allow one of config & perf collectors to start
	uniqueOCs := uniquifyObjectCollectors(objectsToCollectors)

	// start the uniqueified collectors
	err = p.loadCollectorObject(uniqueOCs)
	if err != nil {
		logger.Error("Failed to load collector", slogx.Err(err))
	}

	// at least one collector should successfully initialize
	if len(p.collectors) == 0 {
		logger.Warn("no collectors initialized, stopping")
		return errs.New(errs.ErrNoCollector, "no collectors")
	}

	logger.Debug("collectors initialized", slog.Int("count", len(p.collectors)))

	// we are more tolerable against exporters, since we might only
	// want to debug collectors without actually exporting
	if len(p.exporters) == 0 {
		logger.Warn("no exporters initialized, continuing without exporters")
	} else {
		logger.Debug("exporters initialized", slog.Int("count", len(p.exporters)))
	}

	// initialize a schedule for the poller, this is the interval at which
	// we will check the status of collectors, exporters and target system,
	// and send metadata to exporters
	if p.params.PollerSchedule != "" {
		pollerSchedule = p.params.PollerSchedule
	}
	p.schedule = schedule.New()
	if err = p.schedule.NewTaskString("poller", pollerSchedule, 0, nil, true, "poller_"+p.name); err != nil {
		logger.Error("set schedule:", slogx.Err(err))
		return err
	}

	if p.params.PollerLogSchedule != "" {
		pollerLogSchedule = p.params.PollerLogSchedule
	}
	if err = p.schedule.NewTaskString("log", pollerLogSchedule, 0, p.logPollerMetadata, true, "poller_log_"+p.name); err != nil {
		logger.Error("set log schedule:", slogx.Err(err))
		return err
	}

	logger.Debug(
		"set poller schedule",
		slog.String("pollerSchedule", pollerSchedule),
		slog.String("pollerLogSchedule", pollerLogSchedule),
	)

	// Check if autosupport is enabled
	tools := conf.Config.Tools
	if tools != nil && tools.AsupDisabled {
		logger.Info("Autosupport is disabled")
	} else {
		if p.collectorIsBuiltin() {
			// Write the payload after asupFirstWrite.
			// This is to examine the autosupport contents
			// Nothing is sent, sending happens based on the asupSchedule
			duration, err := time.ParseDuration(asupFirstWrite)
			if err != nil {
				logger.Error(
					"Failed to write first autosupport payload.",
					slogx.Err(err),
					slog.String("asupFirstWrite", asupFirstWrite),
				)
			} else {
				time.AfterFunc(duration, func() {
					p.firstAutoSupport()
				})
			}
			if err := p.schedule.NewTaskString("asup", asupSchedule, 0, p.startAsup, p.options.Asup, "asup_"+p.name); err != nil {
				return err
			}
			logger.Info("Autosupport scheduled", slog.String("asupSchedule", asupSchedule))
		} else {
			logger.Info(
				`Autosupport disabled since poller not connected to ONTAP.`,
				slog.String("poller", p.name),
			)
		}
	}

	// famous last words
	logger.Info("poller start-up complete")

	return nil

}

func uniquifyObjectCollectors(objectsToCollectors map[string][]objectCollector) []objectCollector {
	uniqueOCs := make([]objectCollector, 0, len(objectsToCollectors))

	specialCaseQtree(objectsToCollectors)

	for _, col := range objectsToCollectors {
		uniqueOCs = append(uniqueOCs, nonOverlappingCollectors(col)...)
	}

	return uniqueOCs
}

func specialCaseQtree(objectsToCollectors map[string][]objectCollector) {
	// ZAPI Qtree also publishes quota metrics. This means that if ZAPI Qtree
	// appears earlier in the slice than REST Qtree, REST Quota should be
	// disabled to prevent both ZAPI Qtree and REST Quota from publishing
	// quota metrics simultaneously.

	qtreeCollectors := objectsToCollectors["Qtree"]
	quotaCollectors := objectsToCollectors["Quota"]
	zapiQtreeWillRun := false

	if len(quotaCollectors) == 0 {
		return
	}

	qtreeNoOverlaps := nonOverlappingCollectors(qtreeCollectors)
	for _, oc := range qtreeNoOverlaps {
		if oc.class == "Zapi" {
			zapiQtreeWillRun = true
			break
		}
	}

	if !zapiQtreeWillRun {
		return
	}

	// Disable REST Quota, if it is enabled
	quotaNoOverlaps := nonOverlappingCollectors(quotaCollectors)
	deleteIndex := -1
	for i, oc := range quotaNoOverlaps {
		if oc.class == "Rest" {
			deleteIndex = i
			break
		}
	}
	if deleteIndex != -1 {
		quotaNoOverlaps = slices.Delete(quotaNoOverlaps, deleteIndex, deleteIndex+1)
		objectsToCollectors["Quota"] = quotaNoOverlaps
	}
}

func (p *Poller) firstAutoSupport() {
	if p.collectors == nil {
		return
	}
	if _, err := collector.BuildAndWriteAutoSupport(p.collectors, p.metadataTarget, p.name, p.maxRssBytes); err != nil {
		slog.Error(
			"First autosupport failed",
			slogx.Err(err),
			slog.String("poller", p.name),
		)
	}
}

func (p *Poller) startAsup() (map[string]*matrix.Matrix, error) {
	if p.collectors != nil {
		if err := collector.SendAutosupport(p.collectors, p.metadataTarget, p.name, p.maxRssBytes); err != nil {
			slog.Error(
				"Start autosupport failed.",
				slogx.Err(err),
				slog.String("poller", p.name),
			)
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
		wg        sync.WaitGroup
		semaphore chan struct{}
		col       collector.Collector
	)

	go p.startHeartBeat()

	if p.params.Pool.IsEnabled() {
		// Create a semaphore channel to limit concurrent collector execution.
		// Buffered channel acts as a counting semaphore
		semaphore = make(chan struct{}, p.params.Pool.Limit)
		slog.Info("pool enabled", slog.Int("limit", p.params.Pool.Limit))
	}

	// start collectors
	for _, col = range p.collectors {
		wg.Add(1)
		go col.Start(&wg, semaphore, p.concurrentCollectors)
	}

	// run concurrently and update metadata
	go p.Run()

	wg.Wait()

	// ...until there are no collectors running anymore
	logger.Info("no active collectors -- terminating")

	p.Stop()
}

// Run will periodically check the status of collectors/exporters,
// report metadata and do some housekeeping
func (p *Poller) Run() {

	// poller schedule has the poller and asup task (when enabled)
	task := p.schedule.GetTask("poller")
	asupTask := p.schedule.GetTask("asup")
	logTask := p.schedule.GetTask("log")

	// number of collectors/exporters that are still up
	upCollectors := 0
	upExporters := 0

	// warm-up ping to avoid cold start effects of first ping
	_, _ = p.ping()

	for {
		if task.IsDue() {
			task.Start()
			// flush metadata
			p.metadataTarget.Reset()
			p.status.Reset()
			p.metadata.Reset()

			// ping target system
			if ping, ok := p.ping(); ok {
				_ = p.metadataTarget.LazySetValueUint8("status", "host", 0)
				_ = p.metadataTarget.LazySetValueFloat64("ping", "host", float64(ping))
				_ = p.status.LazySetValueUint8("status", "host", 1)
				_ = p.status.LazySetValueFloat64("ping", "host", float64(ping))
			} else {
				_ = p.metadataTarget.LazySetValueUint8("status", "host", 1)
				_ = p.status.LazySetValueUint8("status", "host", 0)
			}

			p.addMemoryMetadata()

			// add number of goroutines to metadata
			_ = p.metadataTarget.LazySetValueInt64("goroutines", "host", int64(runtime.NumGoroutine()))

			upc := 0 // up collectors
			upe := 0 // up exporters

			// update status of collectors
			for _, c := range p.collectors {
				code, _, msg := c.GetStatus()

				if code == 0 {
					upc++
				}

				key := c.GetName() + "." + c.GetObject()

				_ = p.metadata.LazySetValueUint64("count", key, c.GetCollectCount())
				_ = p.metadata.LazySetValueUint8("status", key, code)

				if msg != "" {
					if instance := p.metadata.GetInstance(key); instance != nil {
						instance.SetLabel("reason", p.truncateReason(msg))
					}
				}
			}

			// add remote version and name to metadata
			p.status.GetInstance("remote").SetLabel("version", p.remote.Version)
			p.status.GetInstance("remote").SetLabel("name", p.remote.Name)

			// update status of exporters
			for _, ee := range p.exporters {
				code, status, msg := ee.GetStatus()
				logger.Debug(
					"exporter status",
					slog.String("name", ee.GetName()),
					slog.Int("code", int(code)),
					slog.String("status", status),
					slog.String("msg", msg),
				)

				if code == 0 {
					upe++
				}

				key := ee.GetClass() + "." + ee.GetName()

				_ = p.metadata.LazySetValueUint64("count", key, ee.GetExportCount())
				_ = p.metadata.LazySetValueUint8("status", key, code)

				if msg != "" {
					if instance := p.metadata.GetInstance(key); instance != nil {
						instance.SetLabel("reason", p.truncateReason(msg))
					}
				}
			}

			for _, ee := range p.exporters {
				if _, err := ee.Export(p.metadata); err != nil {
					logger.Error("export component metadata", slogx.Err(err))
				}
				if _, err := ee.Export(p.metadataTarget); err != nil {
					logger.Error("export target metadata", slogx.Err(err))
				}
				if _, err := ee.Export(p.status); err != nil {
					logger.Error("export poller status", slogx.Err(err))
				}
			}

			// only log when there are changes, which we expect to be infrequent
			if upc != upCollectors || upe != upExporters {
				logger.Info(
					"updated status",
					slog.Group("collectors",
						slog.Int("up", upc),
						slog.Int("total", len(p.collectors)),
					),
					slog.Group("exporters",
						slog.Int("up", upe),
						slog.Int("total", len(p.exporters)),
					),
				)
			}
			upCollectors = upc
			upExporters = upe
		}

		// asup task will be nil when autosupport is disabled
		if asupTask != nil && asupTask.IsDue() {
			_, _ = asupTask.Run()
		}

		if logTask.IsDue() {
			_, _ = logTask.Run()
		}

		p.schedule.Sleep()
	}
}

// Stop gracefully exits the program by closing zeroLog
func (p *Poller) Stop() {
	logger.Info("stopping poller", slog.Int("pid", os.Getpid()))
}

// set up signal disposition
func (p *Poller) handleSignals(signalChannel chan os.Signal) {
	for {
		sig := <-signalChannel
		slog.Info("caught signal", slog.String("signal", sig.String()))
		p.Stop()
		os.Exit(0)
	}
}

// If the target is available, return the TCP connection time in milliseconds and true.
// If the target is not available, return 0 and false.
// If the target is not eligible for a TCP connection check, return 0 and true.
func (p *Poller) ping() (float32, bool) {

	isPingable := false
	for _, col := range p.collectors {
		if conf.IsPingableCollector(col.GetName()) {
			isPingable = true
			break
		}
	}

	if !isPingable {
		return 0, true
	}

	// If the host includes a port, use that port, otherwise use portsToTry
	target := p.target
	portsToTry := []int{443}

	// Extract host and port. This also handles IPv6
	if host, port, err := net.SplitHostPort(target); err == nil {
		if parsedPort, err := strconv.Atoi(port); err == nil && parsedPort > 0 && parsedPort <= 65535 {
			portsToTry = []int{parsedPort}
			target = host
		} else {
			logger.Error("invalid port in target", slog.String("target", p.target), slog.String("port", port))
			return 0, false
		}
	}

	// Attempt to connect to each port, returning after the first successful connection
	for _, port := range portsToTry {
		address := net.JoinHostPort(target, strconv.Itoa(port))
		start := time.Now()
		conn, err := net.DialTimeout("tcp", address, 5*time.Second)
		if err != nil {
			logger.Error("ping failed", slog.String("address", address), slogx.Err(err))
			continue
		}
		elapsed := time.Since(start)
		_ = conn.Close()
		return float32(elapsed) / float32(time.Millisecond), true
	}

	return 0, false
}

// read templates for this collector and return a list of object collectors. If there are
// multiple objects defined for a collector, multiple object collectors will be returned.
func (p *Poller) readObjects(c conf.Collector) ([]objectCollector, error) {
	var (
		class                 string
		err                   error
		template, subTemplate *node.Node
	)

	newC := p.upgradeCollector(c, p.remote)
	if newC.Name != c.Name {
		logger.Info("upgraded collector", slog.String("old", c.Name), slog.String("new", newC.Name))
	}
	c = newC
	class = c.Name
	// throw warning for deprecated collectors
	if r, d := deprecatedCollectors[strings.ToLower(class)]; d {
		if r != "" {
			logger.Warn(
				"collector is deprecated, please use replacement",
				slog.String("collector", class),
				slog.String("replacement", r),
			)
		} else {
			logger.Warn(
				"collector is deprecated, see documentation for help",
				slog.String("collector", class),
			)
		}
	}

	// load the template file(s) of the collector where we expect to find
	// object name or list of objects
	if c.Templates != nil {
		for _, t := range *c.Templates {
			if subTemplate, err = collector.ImportTemplate(p.options.ConfPaths, t, class); err != nil {
				level := slog.LevelWarn
				// When the template is custom.yaml, log at debug level to reduce noise, since that template
				// won't exist for most people
				if strings.Contains(t, "custom.yaml") {
					level = slog.LevelDebug
				}
				if p.remote.IsASAr2() && strings.Contains(t, "asar2/default.yaml") {
					level = slog.LevelDebug
				}
				logger.LogAttrs(
					context.Background(),
					level,
					"Unable to load template",
					slogx.Err(err),
					slog.String("template", t),
					slog.String("collector", class),
					slog.Any("confPaths", p.options.ConfPaths),
				)
				continue
			}
			if template == nil {
				template = subTemplate
			} else {
				logger.Debug("Merging template", slog.String("template", t))
				if c.Name == "Zapi" || c.Name == "ZapiPerf" {
					// Do not overwrite child of objects. They will be concatenated
					template.Merge(subTemplate, []string{"objects"})
				} else {
					template.Merge(subTemplate, []string{""})
				}
			}
		}
	}
	if template == nil {
		return nil, fmt.Errorf("no templates loaded for %s", c.Name)
	}

	// add the poller's parameters to the collector's parameters
	err = p.mergePollerParametersIntoTemplate(template)
	if err != nil {
		return nil, err
	}

	objects := make([]objectCollector, 0)
	templateObject := template.GetChildContentS("object")

	// if `objects` was passed at the cmdline, use them instead of the defaults
	if len(p.options.Objects) != 0 {
		for _, object := range p.options.Objects {
			objects = append(objects, objectCollector{class: class, object: object, template: template})
		}
	} else if templateObject != "" {
		// if object is defined, we only initialize 1 sub-collector / object
		objects = append(objects, objectCollector{class: class, object: templateObject, template: template})
		// if template has list of objects, initialize 1 sub-collector for each
	} else if templateObjects := template.GetChildS("objects"); templateObjects != nil {
		for _, object := range templateObjects.GetChildren() {
			objects = append(objects, objectCollector{class: class, object: object.GetNameS(), template: template})
		}
	} else {
		return nil, errs.New(errs.ErrMissingParam, "collector object")
	}

	return objects, nil
}

type objectCollector struct {
	class          string
	object         string
	template       *node.Node
	viaRedirection bool // true if this collector was created by redirecting from another collector type
}

// dynamically load and initialize a collector
func (p *Poller) loadCollectorObject(ocs []objectCollector) error {

	var cols []collector.Collector

	logger.Debug("Starting collectors", slog.Int("collectors", len(ocs)))

	for _, oc := range ocs {
		col, err := p.newCollector(oc.class, oc.object, oc.template)
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrConnection):
				logger.Warn(
					"abort collector",
					slogx.Err(err),
					slog.String("collector", oc.class),
					slog.String("object", oc.object),
				)
			case errors.Is(err, errs.ErrWrongTemplate):
				logger.Debug("Zapi Status_7mode failed to load", slogx.Err(err))
			default:
				logger.Warn(
					"init collector-object",
					slogx.Err(err),
					slog.String("collector", oc.class),
					slog.String("object", oc.object),
				)
			}
		} else {
			if shouldIgnore := col.GetParams().GetChildContentS("ignore"); shouldIgnore == "true" {
				logger.Debug("ignoring collector", slog.String("collector", oc.class), slog.String("object", oc.object))
				continue
			}
			cols = append(cols, col)
			logger.Debug(
				"initialized collector-object",
				slog.String("collector", oc.class),
				slog.String("object", oc.object),
			)
		}
	}

	p.collectors = append(p.collectors, cols...)
	// link each collector with requested exporter & update metadata
	for _, col := range cols {
		if col == nil {
			logger.Warn("ignoring nil collector")
			continue
		}
		name := col.GetName()
		obj := col.GetObject()

		for _, expName := range col.WantedExporters(p.params.Exporters) {
			if exp := p.loadExporter(expName); exp != nil {
				col.LinkExporter(exp)
			} else {
				logger.Warn(
					"exporter requested not available",
					slog.String("exporterName", expName),
					slog.String("name", name),
					slog.String("object", obj),
				)
			}
		}

		// update metadata

		instance, err := p.metadata.NewInstance(name + "." + obj)
		if err != nil {
			return err
		}
		instance.SetLabel("type", "collector")
		instance.SetLabel("name", name)
		instance.SetLabel("target", obj)
	}

	return nil
}

func nonOverlappingCollectors(objectCollectors []objectCollector) []objectCollector {
	if len(objectCollectors) == 0 {
		return []objectCollector{}
	}
	if len(objectCollectors) == 1 {
		return objectCollectors
	}

	unique := make([]objectCollector, 0)
	conflicts := map[string][]string{
		"Zapi":     {"Rest"},
		"ZapiPerf": {"RestPerf", "KeyPerf", "StatPerf"},
		"Rest":     {"Zapi"},
		"RestPerf": {"ZapiPerf", "KeyPerf", "StatPerf"},
		"KeyPerf":  {"ZapiPerf", "RestPerf", "StatPerf"},
		"StatPerf": {"ZapiPerf", "RestPerf", "KeyPerf"},
	}

	// Sort collectors so native ones (viaRedirection=false) come before redirected ones
	// This ensures native collectors take precedence in conflict resolution
	slices.SortFunc(objectCollectors, func(a, b objectCollector) int {
		if a.viaRedirection == b.viaRedirection {
			return 0
		}
		if a.viaRedirection {
			return 1 // a comes after b (native comes first)
		}
		return -1 // a comes before b
	})

	for _, c := range objectCollectors {
		conflict, ok := conflicts[c.class]
		if ok {
			if collectorContains(unique, conflict, c.class) {
				continue
			}
		}
		unique = append(unique, c)
	}
	return unique
}

func collectorContains(unique []objectCollector, conflicts []string, search string) bool {
	for _, o := range unique {
		if o.class == search {
			return true
		}
		has := slices.Contains(conflicts, o.class)
		if has {
			return true
		}
	}
	return false
}

func (p *Poller) mergePollerParametersIntoTemplate(template *node.Node) error {
	err := Union2(template, p.params)
	if err != nil {
		return fmt.Errorf("failed to merge poller parameters: %w", err)
	}
	template.NewChildS("poller_name", p.params.Name)
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
func Union2(hNode *node.Node, poller *conf.Poller) error {
	marshal2, err := yaml.Marshal(poller)
	if err != nil {
		return fmt.Errorf("failed to marshal poller: %w", err)
	}
	file, err := parser.ParseBytes(marshal2, 0)
	if err != nil {
		return fmt.Errorf("failed to parse poller: %w", err)
	}

	body := file.Docs[0].Body

	if body.Type() == ast.MappingType {
		mn := body.(*ast.MappingNode)
		for _, mvn := range mn.Values {
			if mvn.Key.Type() != ast.StringType {
				continue
			}

			// check if the key exists in the hNode
			key := node.ToString(mvn.Key)
			if hNode.HasChildS(key) {
				// if it does, skip it
				continue
			}
			// if it doesn't, create a new node with the key and value
			newNode := node.NewS(key)

			switch mvn.Value.Type() { //nolint:exhaustive
			case ast.StringType, ast.BoolType, ast.IntegerType:
				newNode.Content = []byte(node.ToString(mvn.Value))
			case ast.SequenceType:
				// the poller node that is missing is a sequence so add all the children of the sequence
				for _, seqNode := range mvn.Value.(*ast.SequenceNode).Values {
					switch seqNode.Type() { //nolint:exhaustive
					case ast.StringType:
						seqStr := node.ToString(seqNode)
						newNode.NewChildS(seqStr, seqStr)
					case ast.MappingType:
						for _, v := range seqNode.(*ast.MappingNode).Values {
							newNode.NewChildS(node.ToString(v.Key), node.ToString(v.Value))
						}
					default:
						return fmt.Errorf("unknown sequence type: %s", seqNode.Type().String())
					}
				}
			case ast.MappingType:
				// the poller node that is missing is a map, add all the children of the map
				for _, v := range mvn.Value.(*ast.MappingNode).Values {
					newNode.NewChildS(node.ToString(v.Key), node.ToString(v.Value))
				}
			default:
				return fmt.Errorf("unknown mapping type: %s", mvn.Value.Type().String())
			}
			hNode.AddChild(newNode)
		}
	}

	return nil
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
	delegate := collector.New(class, object, p.options, template.Copy(), p.auth, p.remote)
	err = col.Init(delegate)
	return col, err
}

// Returns the exporter with the matching name.
// If the exporter is not loaded, load and return it.
func (p *Poller) loadExporter(name string) exporter.Exporter {

	var (
		err    error
		class  string
		params conf.Exporter
		exp    exporter.Exporter
	)

	// stop here if exporter is already loaded
	if exp := p.getExporter(name); exp != nil {
		return exp
	}

	params, ok := p.exporterParams[name]
	if !ok {
		logger.Warn("exporter not defined in config", slog.String("name", name))
		return nil
	}

	if class = params.Type; class == "" {
		logger.Warn("exporter has no exporter class defined", slog.String("name", name))
		return nil
	}

	absExp := exporter.New(class, name, p.options, params, p.params)
	switch class {
	case "Prometheus":
		exp = prometheus.New(absExp)
	case "InfluxDB":
		exp = influxdb.New(absExp)
	case "VictoriaMetrics":
		exp = victoriametrics.New(absExp)
	default:
		logger.Error("no exporter of name:type", slog.String("name", name), slog.String("type", class))
		return nil
	}
	if err = exp.Init(); err != nil {
		logger.Error("Unable to init exporter", slogx.Err(err), slog.String("name", name))
		return nil
	}

	p.exporters = append(p.exporters, exp)
	logger.Debug("initialized exporter", slog.String("name", name), slog.String("type", class))

	// update metadata
	if instance, err := p.metadata.NewInstance(exp.GetClass() + "." + exp.GetName()); err != nil {
		logger.Error("add metadata instance", slogx.Err(err))
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
	p.metadata.SetGlobalLabel("datacenter", p.params.Datacenter)
	p.metadata.SetGlobalLabel("hostname", p.options.Hostname)
	if p.options.PromPort != 0 {
		p.metadata.SetGlobalLabel("promport", strconv.Itoa(p.options.PromPort))
	}
	p.metadata.SetExportOptions(matrix.DefaultExportOptions())

	// metadata for the target system
	p.metadataTarget = matrix.New("poller", "metadata_target", "metadata_component")
	_, _ = p.metadataTarget.NewMetricUint8("status")
	_, _ = p.metadataTarget.NewMetricFloat64("ping")
	_, _ = p.metadataTarget.NewMetricUint64("goroutines")

	// metadata for the poller itself
	p.status = matrix.New("poller", "poller", "poller_target")
	_, _ = p.status.NewMetricUint8("status")
	_, _ = p.status.NewMetricFloat64("memory_percent")
	newMemoryMetric(p.status, "memory", "rss")
	newMemoryMetric(p.status, "memory", "vms")
	newMemoryMetric(p.status, "memory", "swap")
	_, _ = p.status.NewMetricFloat64("concurrent_collectors")

	instance, _ := p.metadataTarget.NewInstance("host")
	pInstance, _ := p.status.NewInstance("host")
	pRemote, _ := p.status.NewInstance("remote")
	pRemote.SetExportable(false)

	instance.SetLabel("addr", p.target)
	pInstance.SetLabel("addr", p.target)

	globalKVs := []string{
		"pid", strconv.Itoa(os.Getpid()),
		"poller", p.name,
		"version", p.options.Version,
		"datacenter", p.params.Datacenter,
		"hostname", p.options.Hostname,
	}

	for i := 0; i < len(globalKVs); i += 2 {
		p.metadataTarget.SetGlobalLabel(globalKVs[i], globalKVs[i+1])
		p.status.SetGlobalLabel(globalKVs[i], globalKVs[i+1])
	}

	if p.options.PromPort != 0 {
		p.metadataTarget.SetGlobalLabel("promport", strconv.Itoa(p.options.PromPort))
		p.status.SetGlobalLabel("promport", strconv.Itoa(p.options.PromPort))
	}

	labels := p.params.Labels
	if labels != nil {
		for _, labelPtr := range *labels {
			p.metadata.SetGlobalLabels(labelPtr)
			p.metadataTarget.SetGlobalLabels(labelPtr)
			p.status.SetGlobalLabels(labelPtr)
		}
	}
	p.metadataTarget.SetExportOptions(matrix.DefaultExportOptions())
	p.status.SetExportOptions(matrix.DefaultExportOptions())
}

func newMemoryMetric(status *matrix.Matrix, label string, sub string) {
	fullLabel := label + "." + sub
	mm, _ := status.NewMetricType(fullLabel, "uint64", label)
	mm.SetLabel("metric", sub)
}

var pollerCmd = &cobra.Command{
	Use:   "poller -p name [flags]",
	Short: "Harvest Poller - Runs collectors and exporters for a target system",
	Args:  cobra.NoArgs,
	Run:   startPoller,
}

// Returns true if at least one collector is one of the builtin collectors.
func (p *Poller) collectorIsBuiltin() bool {
	for _, c := range p.collectors {
		_, ok := conf.IsCollector[c.GetName()]
		if ok {
			return true
		}
	}
	return false
}

// Returns true if at least one collector is known to collect from an Ontap system.
func (p *Poller) targetIsOntap() bool {
	for _, c := range p.collectors {
		_, ok := conf.IsONTAPCollector[c.GetName()]
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
	localIP, err := requests.FindLocalIP()
	if err != nil {
		logger.Error("Unable to find local IP", slogx.Err(err))
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
		logger.Error("Unable to marshal poller details", slogx.Err(err), slog.String("poller", p.name))
		return
	}
	defaultURL := p.makePublishURL()

	if heartBeatURL == "" {
		heartBeatURL = defaultURL
	}
	req, err := requests.New("PUT", heartBeatURL, bytes.NewBuffer(payload))
	if err != nil {
		logger.Error("failed to connect to admin", slogx.Err(err))
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
		level := slog.LevelError
		if strings.Contains(rErr.Error(), "connection refused") {
			level = slog.LevelWarn
		}
		logger.LogAttrs(
			context.Background(),
			level,
			"Failed connecting to admin node",
			slog.Any("err", rErr),
			slog.String("admin", conf.Config.Admin.Httpsd.Listen),
		)
		return
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read publishDetails response to admin", slogx.Err(err))
		return
	}
	p.client.CloseIdleConnections()
	if resp.StatusCode != http.StatusOK {
		txt := string(body)
		txt = txt[0:int(math.Min(float64(len(txt)), 48))]
		logger.Error(
			"Admin node problem",
			slog.String("admin", conf.Config.Admin.Httpsd.Listen),
			slog.String("body", txt),
			slog.Int("httpStatusCode", resp.StatusCode),
		)
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
		logger.Warn(
			"Invalid heart_beat using 1m",
			slogx.Err(err),
			slog.String("heart_beat", conf.Config.Admin.Httpsd.HeartBeat),
		)
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
// ZAPI collectors should be upgraded to REST collectors when the cluster no longer speaks Zapi
func (p *Poller) upgradeCollector(c conf.Collector, remote conf.Remote) conf.Collector {
	// If REST is desired, use REST
	// If ZAPI is desired, check that the cluster speaks ZAPI and if so, use ZAPI, otherwise use REST
	// If KeyPerf is desired, negotiate the API
	// EMS and StorageGRID are ignored

	if _, ok := conf.IsONTAPCollector[c.Name]; !ok {
		return c
	}

	isKeyPerf := remote.IsKeyPerf()
	replaced := c.Name

	if strings.HasPrefix(replaced, "Zapi") {

		if remote.ZAPIsExist {
			switch replaced {
			case "Zapi":
				return c
			case "ZapiPerf":
				if isKeyPerf {
					replaced = "RestPerf"
				} else {
					return c
				}
			}
		}

		replaced = strings.ReplaceAll(replaced, "Zapi", "Rest")
		if isKeyPerf {
			replaced = strings.ReplaceAll(replaced, "RestPerf", "KeyPerf")
		}
		return conf.Collector{
			Name:      replaced,
			Templates: c.Templates,
		}
	}

	if isKeyPerf {
		replaced := strings.ReplaceAll(c.Name, "RestPerf", "KeyPerf")
		return conf.Collector{
			Name:      replaced,
			Templates: c.Templates,
		}
	}

	return c
}

func (p *Poller) upgradeObjectCollector(oc objectCollector) objectCollector {
	templateObjects := oc.template.GetChildS("objects")
	if templateObjects == nil {
		return oc
	}
	object := templateObjects.GetChildS(oc.object)
	if object == nil {
		return oc
	}
	objectValue := object.GetContentS()
	collectorName, templateName, isUpgraded := collector.ParseTemplateRef(objectValue)
	if !isUpgraded {
		return oc
	}

	// Handle KeyPerf upgrades from ZapiPerf
	if oc.class == "ZapiPerf" && collectorName == "KeyPerf" {
		// Check version compatibility for KeyPerf upgrades on older ONTAP versions
		// Don't upgrade to KeyPerf for versions below 9.10
		if supported, err := version2.AtLeast(p.remote.Version, "9.10.0"); err != nil || !supported {
			// This check is needed because KeyPerf uses an endpoint /api/storage/volumes
			// which requires is_constituent parameter which is not available prior to 9.10.
			// For versions below 9.10, we skip the KeyPerf upgrade and fall back to the original collector.
			// Check if object name contains "volume"
			if strings.Contains(strings.ToLower(oc.object), "volume") {
				object.SetContentS(templateName)
				logger.Warn(
					"volume KeyPerf upgrade skipped due to ONTAP version",
					slog.String("object", oc.object),
					slog.String("template", templateName),
					slog.String("ontapVersion", p.remote.Version),
					slog.String("requiredVersion", "9.10.0+"),
				)
				return oc
			}
		}

		// When upgrading from ZapiPerf to KeyPerf, strip any extended templates
		// as it's not needed/supported in KeyPerf.
		parts := strings.SplitN(templateName, ",", 2)
		templateName = strings.TrimSpace(parts[0])
	}

	// Find the appropriate default templates for the target collector class
	targetTemplates := p.fetchCollectorTemplates(collectorName)

	var upgradedTemplate *node.Node
	var err error
	for _, t := range targetTemplates {
		upgradedTemplate, err = collector.ImportTemplate(p.options.ConfPaths, t, collectorName)
		if err == nil {
			break
		}
	}

	if err != nil || upgradedTemplate == nil {
		logger.Warn(
			"failed to load upgraded template",
			slogx.Err(err),
			slog.String("class", collectorName),
			slog.String("object", oc.object),
			slog.Any("templates", targetTemplates),
		)
		return oc
	}

	// Validate that the upgraded template contains the target object
	if templatesObjects := upgradedTemplate.GetChildS("objects"); templatesObjects != nil {
		if templatesObjects.GetChildS(oc.object) == nil {
			logger.Warn(
				"target template does not contain required object",
				slog.String("object", oc.object),
				slog.String("class", collectorName),
				slog.String("template", templateName),
			)
			return oc
		}
	} else {
		logger.Warn(
			"target template has no objects section",
			slog.String("class", collectorName),
			slog.String("template", templateName),
		)
		return oc
	}

	// Create a new template with only the specific object we want
	newTemplate := upgradedTemplate.Copy()

	// Merge poller parameters into upgraded template
	err = p.mergePollerParametersIntoTemplate(newTemplate)
	if err != nil {
		logger.Warn(
			"failed to merge poller parameters into upgraded template",
			slogx.Err(err),
			slog.String("object", oc.object),
			slog.String("class", collectorName),
		)
		return oc
	}

	if templatesObjects := newTemplate.GetChildS("objects"); templatesObjects != nil {
		// Filter children to keep only the target object
		var filteredChildren []*node.Node
		for _, child := range templatesObjects.GetChildren() {
			if child.GetNameS() == oc.object {
				filteredChildren = append(filteredChildren, child)
			}
		}
		templatesObjects.Children = filteredChildren

		// Update the object definition to use the specified template name
		if objDef := templatesObjects.GetChildS(oc.object); objDef != nil {
			objDef.SetContentS(templateName)
		}
	}

	logger.Info(
		"object upgraded",
		slog.String("object", oc.object),
		slog.String("from", oc.class),
		slog.String("to", collectorName),
		slog.String("template", templateName),
	)

	return objectCollector{
		class:          collectorName,
		object:         oc.object,
		template:       newTemplate,
		viaRedirection: true, // this collector was created via redirection from another type
	}
}

func (p *Poller) fetchCollectorTemplates(collectorClass string) []string {
	for _, col := range p.params.Collectors {
		if col.Name == collectorClass && col.Templates != nil {
			return *col.Templates
		}
	}

	return *conf.DefaultTemplates
}

// set the poller's confPath using the following precedence:
// CLI, harvest.yml, default (conf)
func (p *Poller) mergeConfPath() {
	path := conf.DefaultConfPath
	if p.params.ConfPath != "" {
		path = p.params.ConfPath
	}
	if p.options.ConfPath != conf.DefaultConfPath {
		path = p.options.ConfPath
	}
	p.options.SetConfPath(path)
}

func (p *Poller) addMemoryMetadata() {

	memMetrics := collector.MemoryMetrics()

	// The unix poller used KB for memory so use the same here
	_ = p.status.LazySetValueUint64("memory.rss", "host", memMetrics.RSSBytes/1024)
	_ = p.status.LazySetValueUint64("memory.vms", "host", memMetrics.VMSBytes/1024)
	_ = p.status.LazySetValueUint64("memory.swap", "host", memMetrics.SwapBytes/1024)
	_ = p.status.LazySetValueFloat64("memory_percent", "host", memMetrics.PercentageRssUsed)
	_ = p.status.LazyAddValueInt64("concurrent_collectors", "host", int64(p.concurrentCollectors.Load()))

	// Update maxRssBytes
	p.maxRssBytes = max(p.maxRssBytes, memMetrics.RSSBytes)
}

func (p *Poller) logPollerMetadata() (map[string]*matrix.Matrix, error) {
	err := p.sendHarvestVersion()
	if err != nil {
		slog.Error("Failed to send Harvest version", slogx.Err(err))
	}

	remoteName := p.status.GetInstance("remote").GetLabel("name")
	remoteVersion := p.status.GetInstance("remote").GetLabel("version")

	memMetrics := collector.MemoryMetrics()
	p.maxRssBytes = max(p.maxRssBytes, memMetrics.RSSBytes)

	slog.Info(
		"Metadata",
		slog.Group("remote",
			slog.String("name", remoteName),
			slog.String("version", remoteVersion),
		),
		slog.String("version", strings.TrimSpace(version.String())),
		slog.Group("mem",
			slog.Uint64("liveHeapMB", memMetrics.LiveHeapBytes/1024/1024),
			slog.Uint64("heapMB", memMetrics.HeapSizeBytes/1024/1024),
			slog.Uint64("heapGoalMB", memMetrics.HeapGoalBytes/1024/1024),
			slog.Uint64("rssMB", memMetrics.RSSBytes/1024/1024),
			slog.Uint64("maxRssMB", p.maxRssBytes/1024/1024),
		),
		slog.Uint64("uptimeSeconds", uint64(time.Since(p.startTime).Seconds())),
	)

	return nil, nil
}

func (p *Poller) sendHarvestVersion() error {
	var (
		poller     *conf.Poller
		connection *rest.Client
		err        error
	)

	if !p.targetIsOntap() {
		return nil
	}
	if !p.remote.HasREST {
		return nil
	}

	// If the cluster is running ONTAP 9.11.1 or later,
	// send a harvestTag to the cluster to indicate that Harvest is running.
	// Otherwise, do nothing

	ontapVersion, err := goversion.NewVersion(p.remote.Version)
	if err != nil {
		return err
	}

	if ontapVersion.LessThan(goversion.Must(goversion.NewVersion("9.11.1"))) {
		return nil
	}

	// connect to the cluster
	if poller, err = conf.PollerNamed(opts.Poller); err != nil {
		return err
	}
	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if connection, err = rest.New(poller, timeout, p.auth); err != nil {
		return err
	}

	// Send the harvestTag to the ONTAP cluster including the OS name, sha1(hostname), Harvest version, and max RSS in MB
	osName := collector.GetOSName()
	hostname, _ := os.Hostname()
	sha1Hostname := collector.Sha1Sum(hostname)
	rssMB := p.maxRssBytes / 1024 / 1024
	fields := []string{osName, sha1Hostname, version.VERSION, strconv.FormatUint(rssMB, 10)}

	href := `api/cluster?ignore_unknown_fields=true&fields=harvestTag,` + strings.Join(fields, ",")
	_, err = connection.GetPlainRest(href, false)
	if err != nil {
		return err
	}

	return nil
}

func (p *Poller) negotiateAPI(cols []conf.Collector) []conf.Collector {
	ontapCols := p.filterONTAPCollectors(cols)
	ciscoCols := p.filterCiscoCollectors(cols)
	sgCols := p.filterStorageGridCollectors(cols)
	eseriesCols := p.filterEseriesCollectors(cols)
	otherCols := p.filterOtherCollectors(cols)

	var validCollectors []conf.Collector

	if len(ontapCols) > 0 {
		if p.negotiateConnection("ONTAP", ontapCols) {
			if p.remote.IsASAr2() {
				for _, col := range ontapCols {
					if slices.Equal(*col.Templates, *conf.DefaultTemplates) {
						col.Templates = new(append(*col.Templates, "asar2/default.yaml", "asar2/custom.yaml"))
						validCollectors = append(validCollectors, col)
					} else {
						validCollectors = append(validCollectors, col)
					}
				}
			} else {
				validCollectors = append(validCollectors, ontapCols...)
			}
		} else {
			logger.Warn("ONTAP connection failed, skipping ONTAP collectors")
		}
	}

	if len(ciscoCols) > 0 {
		if p.negotiateConnection("Cisco", ciscoCols) {
			validCollectors = append(validCollectors, ciscoCols...)
		} else {
			logger.Warn("Cisco connection failed, skipping Cisco collectors")
		}
	}

	if len(sgCols) > 0 {
		if p.negotiateConnection("StorageGrid", sgCols) {
			validCollectors = append(validCollectors, sgCols...)
		} else {
			logger.Warn("Storage Grid connection failed, skipping StorageGrid collectors")
		}
	}

	if len(eseriesCols) > 0 {
		if p.negotiateConnection("Eseries", eseriesCols) {
			validCollectors = append(validCollectors, eseriesCols...)
		} else {
			logger.Warn("ESeries connection failed, skipping Eseries collectors")
		}
	}

	// Include other collectors without connection validation
	if len(otherCols) > 0 {
		validCollectors = append(validCollectors, otherCols...)
		logger.Debug("including other collectors", slog.Int("count", len(otherCols)))
	}

	return validCollectors
}

func (p *Poller) filterOtherCollectors(cols []conf.Collector) []conf.Collector {
	otherCollectors := make([]conf.Collector, 0, len(cols))
	for _, c := range cols {
		if _, ok := conf.IsONTAPCollector[c.Name]; ok {
			continue
		}
		if _, ok := conf.IsNonONTAPCollector[c.Name]; ok {
			continue
		}
		// Include everything else
		otherCollectors = append(otherCollectors, c)
	}
	return otherCollectors
}

func (p *Poller) negotiateConnection(connectionType string, cols []conf.Collector) bool {
	var remote conf.Remote
	var err error

	switch connectionType {
	case "ONTAP":
		remote, err = collectors.GatherClusterInfo(opts.Poller, p.auth, cols)
	case "Cisco":
		remote, err = collectors.GatherCiscoSwitchInfo(opts.Poller, p.auth)
	case "StorageGrid":
		remote, err = collectors.GatherStorageGridInfo(opts.Poller, p.auth)
	case "Eseries":
		remote, err = collectors.GatherEseriesInfo(opts.Poller, p.auth)
	default:
		logger.Warn("unknown connection type", slog.String("type", connectionType))
		return false
	}

	if err != nil {
		logger.Warn("gather remote info failed",
			slog.String("connectionType", connectionType),
			slog.Any("remote", remote),
			slogx.Err(err))
		return false
	}

	p.remote = remote
	if remote.Version != "" {
		slog.Info("Cluster info", slog.Any("remote", remote))
	}
	return true
}

func (p *Poller) filterONTAPCollectors(cols []conf.Collector) []conf.Collector {
	var ontapCollectors []conf.Collector
	for _, c := range cols {
		if _, ok := conf.IsONTAPCollector[c.Name]; ok {
			ontapCollectors = append(ontapCollectors, c)
		}
	}
	return ontapCollectors
}

func (p *Poller) filterCiscoCollectors(cols []conf.Collector) []conf.Collector {
	var ciscoCollectors []conf.Collector
	for _, c := range cols {
		if c.Name == "CiscoRest" {
			ciscoCollectors = append(ciscoCollectors, c)
		}
	}
	return ciscoCollectors
}

func (p *Poller) filterStorageGridCollectors(cols []conf.Collector) []conf.Collector {
	var sgCollectors []conf.Collector
	for _, c := range cols {
		if c.Name == "StorageGrid" {
			sgCollectors = append(sgCollectors, c)
		}
	}
	return sgCollectors
}

func (p *Poller) filterEseriesCollectors(cols []conf.Collector) []conf.Collector {
	var eseriesCollectors []conf.Collector
	for _, c := range cols {
		if _, ok := conf.IsESeriesCollector[c.Name]; ok {
			eseriesCollectors = append(eseriesCollectors, c)
		}
	}
	return eseriesCollectors
}

func (p *Poller) truncateReason(msg string) string {
	// truncate the reason so it is not too long. This will turn
	// "failed to fetch data: error making request connection error Get https://xxx/api/private/cl"
	// into "failed to fetch data"
	msg, _, _ = strings.Cut(msg, ":")
	// replace quotes with empty, in case of rest error may have quotes around endpoint which fails prometheus discovery
	return strings.ReplaceAll(msg, "\"", "")
}

func startPoller(_ *cobra.Command, _ []string) {
	poller := &Poller{}
	poller.options = opts
	if poller.Init() != nil {
		// error already logger by poller
		poller.Stop()
		os.Exit(1)
	}
	poller.Start()
	os.Exit(0)
}

func init() {
	opts = options.New()
	opts.Version = version.VERSION

	var flags = pollerCmd.Flags()
	flags.StringVarP(&opts.Poller, "poller", "p", "", "Poller name as defined in config")
	flags.BoolVarP(&opts.Debug, "debug", "d", false, "Enable debug logging (same as --loglevel 1). If both debug and loglevel are specified, loglevel wins")
	flags.BoolVar(&opts.Daemon, "daemon", false, "Start as daemon")
	flags.IntVarP(&opts.LogLevel, "loglevel", "l", 2, "Logging level (0=trace, 1=debug, 2=info, 3=warning, 4=error, 5=critical)")
	flags.StringVar(&opts.LogFormat, "logformat", "plain", "Log format (plain or json)")
	flags.BoolVar(&opts.LogToFile, "logtofile", false, "Log to the poller_ name prefixed file in the "+logging.GetLogPath()+" directory, instead of stdout")
	flags.IntVar(&opts.Profiling, "profiling", 0, "If profiling port > 0, enables profiling via localhost:PORT/debug/pprof/")
	flags.IntVar(&opts.PromPort, "promPort", 0, "Prometheus Port")
	flags.StringVar(&opts.Config, "config", conf.HarvestYML, "Harvest config file path")
	flags.StringSliceVarP(&opts.Collectors, "collectors", "c", []string{}, "Only start these collectors (overrides harvest.yml)")
	flags.StringSliceVarP(&opts.Objects, "objects", "o", []string{}, "Only start these objects (overrides collector config)")
	flags.StringVar(&opts.ConfPath, "confpath", conf.DefaultConfPath, "colon-separated paths to search for Harvest templates")

	// Used to test autosupport at startup. An environment variable is used instead of a cmdline
	// arg, so we don't have to also add this testing arg to harvest cli
	if isAsup := os.Getenv("ASUP"); isAsup != "" {
		opts.Asup = true
	}

	_ = pollerCmd.MarkFlagRequired("poller")
	_ = pollerCmd.Flags().MarkHidden("logtofile")
}

// start poller, if fails try to write to syslog
func main() {
	cobra.CheckErr(pollerCmd.Execute())
}
