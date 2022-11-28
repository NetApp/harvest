/*
Copyright NetApp Inc, 2021 All rights reserved

Package collector provides the Collector interface and
the AbstractCollector type which implements most basic
attributes.

A Harvest collector should normally "inherit" all these
attributes and implement only the PollData function.
The AbstractCollector will make sure that the collector
is properly initialized, metadata are updated and
data poll(s) and plugins run as scheduled. The collector
can also choose to override any of the attributes
implemented by AbstractCollector.
*/

package collector

import (
	"errors"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"

	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/poller/schedule"
)

// Collector defines the attributes of a collector
// The poll functions (PollData, PollInstance, etc)
// are not part of the interface and are linked dynamically
// All required functions are implemented by AbstractCollector
//
// Note that many of the functions required by the interface
// are only there to facilitate "inheritance" through AbstractCollector.
type Collector interface {
	Init(*AbstractCollector) error
	Start(*sync.WaitGroup)
	GetName() string
	GetObject() string
	GetParams() *node.Node
	GetOptions() *options.Options
	GetCollectCount() uint64
	AddCollectCount(uint64)
	GetStatus() (uint8, string, string)
	SetStatus(uint8, string)
	SetSchedule(*schedule.Schedule)
	SetMatrix(map[string]*matrix.Matrix)
	SetMetadata(*matrix.Matrix)
	WantedExporters([]string) []string
	LinkExporter(exporter.Exporter)
	LoadPlugins(*node.Node, Collector, string) error
	LoadPlugin(string, *plugin.AbstractPlugin) plugin.Plugin
	CollectAutoSupport(p *Payload)
}

// Status defines the possible states of a collector
var Status = [3]string{
	"up",
	"standby",
	"failed",
}

// AbstractCollector implements all required attributes of Collector.
// A "real" collector will "inherit" all these attributes and has
// the option to override them. The real collector should implement
// at least one poll function (usually PollData). AbstractCollector
// will link these functions to its Schedule and make sure that they
// are properly and timely executed.
type AbstractCollector struct {
	Name    string           // name of the collector, CamelCased
	Object  string           // object of the collector, describes what that collector is collecting
	Logger  *logging.Logger  // logger used for logging
	Status  uint8            // current state of th
	Message string           // reason if collector is in failed state
	Options *options.Options // poller options
	Params  *node.Node       // collector parameters
	// note that this is a merge of poller parameters, collector conf and object conf ("subtemplate")
	Schedule     *schedule.Schedule         // schedule of the collector
	Matrix       map[string]*matrix.Matrix  // the data storage of the collector
	Metadata     *matrix.Matrix             // metadata of the collector, such as poll duration, collected data points etc.
	Exporters    []exporter.Exporter        // the exporters that the collector will emit data to
	Plugins      map[string][]plugin.Plugin // built-in or custom plugins
	collectCount uint64                     // count of collected data points
	// this is different from what the collector will have in its metadata, since this variable
	// holds count independent of the poll interval of the collector, used to give stats to Poller
	countMux    *sync.Mutex // used for atomic access to collectCount
	HostVersion string
	HostModel   string
	HostUUID    string
}

// New creates an AbstractCollector with the given arguments:
// @name	- name of the collector
// @object	- object of the collector (something that best describes the data)
// @options	- poller options
// @params	- collector parameters
func New(name, object string, options *options.Options, params *node.Node) *AbstractCollector {
	c := AbstractCollector{
		Name:     name,
		Object:   object,
		Options:  options,
		Logger:   logging.Get().SubLogger("collector", name+":"+object),
		Params:   params,
		countMux: &sync.Mutex{},
	}
	return &c
}

// Init initializes a collector and does the trick of "inheritance",
// hence a function and not a method.
// A collector can to choose to call this function
// inside its Init method, or leave it to be called
// by the poller during dynamic load.
//
// The important thing done here is to look what tasks are defined
// in the "schedule" parameter of the collector and create a pointer
// to the corresponding method of the collector. Example, parameter is:
//
// schedule:
//
//	data: 10s
//	instance: 20s
//
// then we expect that the collector has methods PollData and PollInstance
// that need to be invoked every 10 and 20 seconds respectively.
// Names of the polls are arbitrary, only "data" is a special case, since
// plugins are executed after the data poll (this might change).
func Init(c Collector) error {

	params := c.GetParams()
	opts := c.GetOptions()
	name := c.GetName()
	object := c.GetObject()

	// Initialize schedule and tasks (polls)
	tasks := params.GetChildS("schedule")
	if tasks == nil || len(tasks.GetChildren()) == 0 {
		return errs.New(errs.ErrMissingParam, "schedule")
	}

	s := schedule.New()

	// Each task will be mapped to a collector method
	// Example: "data" will be aligned to method PollData()
	caser := cases.Title(language.Und)
	for _, task := range tasks.GetChildren() {

		methodName := "Poll" + caser.String(task.GetNameS())

		if m := reflect.ValueOf(c).MethodByName(methodName); m.IsValid() {
			if foo, ok := m.Interface().(func() (map[string]*matrix.Matrix, error)); ok {
				if err := s.NewTaskString(task.GetNameS(), task.GetContentS(), foo, true, "Collector_"+c.GetName()+"_"+c.GetObject()); err != nil {
					return errs.New(errs.ErrInvalidParam, "schedule ("+task.GetNameS()+"): "+err.Error())
				}
			} else {
				return errs.New(errs.ErrImplement, methodName+" has not signature 'func() (*matrix.Matrix, error)'")
			}
		} else {
			return errs.New(errs.ErrImplement, methodName)
		}
	}
	c.SetSchedule(s)

	// Initialize Matrix, the container of collected data
	mx := matrix.New(name, object, object)
	if exportOptions := params.GetChildS("export_options"); exportOptions != nil {
		mx.SetExportOptions(exportOptions)
	} else {
		mx.SetExportOptions(matrix.DefaultExportOptions())
		// @TODO log warning for user
	}
	mx.SetGlobalLabel("datacenter", params.GetChildContentS("datacenter"))

	// Add user-defined global labels
	if gl := params.GetChildS("global_labels"); gl != nil {
		for _, c := range gl.GetChildren() {
			mx.SetGlobalLabel(c.GetNameS(), c.GetContentS())
		}
	}

	// Some data should not be exported and is only used for plugins
	if params.GetChildContentS("export_data") == "false" {
		mx.SetExportable(false)
	}

	var m = make(map[string]*matrix.Matrix)

	m[mx.Object] = mx

	c.SetMatrix(m)

	// Initialize Plugins
	if plugins := params.GetChildS("plugins"); plugins != nil {
		if err := c.LoadPlugins(plugins, c, c.GetObject()); err != nil {
			return err
		}
	}

	// Initialize metadata
	md := matrix.New(name, "metadata_collector", "metadata_collector")

	md.SetGlobalLabel("hostname", opts.Hostname)
	md.SetGlobalLabel("version", opts.Version)
	md.SetGlobalLabel("poller", opts.Poller)
	md.SetGlobalLabel("collector", name)
	md.SetGlobalLabel("object", object)
	md.SetGlobalLabel("datacenter", params.GetChildContentS("datacenter"))

	_, _ = md.NewMetricInt64("poll_time")
	_, _ = md.NewMetricInt64("task_time")
	_, _ = md.NewMetricInt64("api_time")
	_, _ = md.NewMetricInt64("parse_time")
	_, _ = md.NewMetricInt64("calc_time")
	_, _ = md.NewMetricInt64("plugin_time")
	_, _ = md.NewMetricUint64("metrics")
	_, _ = md.NewMetricUint64("instances")

	// add tasks of the collector as metadata instances
	for _, task := range s.GetTasks() {
		instance, _ := md.NewInstance(task.Name)
		instance.SetLabel("task", task.Name)
		t := task.GetInterval().Seconds()
		instance.SetLabel("interval", strconv.FormatFloat(t, 'f', 4, 32))
	}

	md.SetExportOptions(matrix.DefaultExportOptions())

	c.SetMetadata(md)
	c.SetStatus(0, "initialized")

	return nil
}

// @TODO unsafe to read concurrently

func (c *AbstractCollector) GetMetadata() *matrix.Matrix {
	return c.Metadata
}

func (c *AbstractCollector) GetHostModel() string {
	return c.HostModel
}

func (c *AbstractCollector) GetHostVersion() string {
	return c.HostVersion
}

func (c *AbstractCollector) GetHostUUID() string {
	return c.HostUUID
}

// Start will run the collector in an infinite loop
func (c *AbstractCollector) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			c.Logger.Error().Stack().Err(errs.New(errs.ErrPanic, "")).
				Msgf("Collector panicked %s", r)
		}
	}()

	// keep track of connection errors
	// to increment time before retry
	// @TODO add to metadata
	retryDelay := 1
	c.SetStatus(0, "running")

	for {

		// We can't reset metadata here because autosupport metadata is reset
		// https://github.com/NetApp/harvest-private/issues/114 for details
		//c.Metadata.Reset()

		results := make([]*matrix.Matrix, 0)

		// run all scheduled tasks
		for _, task := range c.Schedule.GetTasks() {
			if !task.IsDue() {
				continue
			}

			if c.Schedule.IsStandBy() && !c.Schedule.IsTaskStandBy(task) {
				c.Logger.Info().
					Str("task", task.Name).
					Msg("skip, schedule is in standby")
				continue
			}

			var (
				start, pluginStart   time.Time
				taskTime, pluginTime time.Duration
			)

			// reset task metadata
			c.Metadata.ResetInstance(task.Name)

			start = time.Now()
			data, err := task.Run()
			taskTime = time.Since(start)

			// poll returned error, try to understand what to do
			if err != nil {
				if !c.Schedule.IsStandBy() {
					c.Logger.Debug().Msgf("handling error during [%s] poll...", task.Name)
				}
				switch {
				// target system is unreachable
				// enter standby mode and retry with some delay that will be increased if we fail again
				case errors.Is(err, errs.ErrConnection):
					if retryDelay < 1024 {
						retryDelay *= 4
					}
					if !c.Schedule.IsStandBy() {
						c.Logger.Warn().
							Str("task", task.Name).
							Int("retryDelaySecs", retryDelay).
							Msg("target unreachable, entering standby mode and retry")
					}
					c.Logger.Debug().
						Err(err).
						Str("task", task.Name).
						Int("retryDelaySecs", retryDelay).
						Msg("target unreachable, entering standby mode and retry")
					c.Schedule.SetStandByMode(task, time.Duration(retryDelay)*time.Second)
					c.SetStatus(1, errs.ErrConnection.Error())
				// there are no instances to collect
				case errors.Is(err, errs.ErrNoInstance):
					c.Schedule.SetStandByMode(task, 5*time.Minute)
					c.SetStatus(1, errs.ErrNoInstance.Error())
					c.Logger.Info().
						Str("task", task.Name).
						Msg("no instances, entering standby")
				// no metrics available
				case errors.Is(err, errs.ErrNoMetric):
					c.SetStatus(1, errs.ErrNoMetric.Error())
					c.Schedule.SetStandByMode(task, 1*time.Hour)
					c.Logger.Info().
						Str("task", task.Name).
						Str("object", c.Object).
						Msg("no metrics of object on system, entering standby mode")
				// not an error we are expecting, so enter failed state and terminate
				default:
					var herr errs.HarvestError
					errMsg := err.Error()

					if ok := errors.As(err, &herr); ok {
						errMsg = herr.Inner.Error()
					}
					// API was rejected, this happens when a resource is not available or does not exist
					if errors.Is(err, errs.ErrAPIRequestRejected) {
						c.Logger.Info().Str("task", task.Name).Msg(err.Error())
					} else {
						c.Logger.Error().Err(err).Str("task", task.Name).Msg("")
					}
					c.SetStatus(2, errMsg)
				}
				// stop here if we had errors
				continue
			} else if c.Schedule.IsStandBy() {
				// recover from standby mode
				c.Schedule.Recover()
				retryDelay = 1
				c.SetStatus(0, "running")
				c.Logger.Info().Str("task", task.Name).Msg("recovered from standby mode, back to normal schedule")
			} else {
				c.SetStatus(0, "running")
			}

			if data != nil {

				for _, value := range data {
					results = append(results, value)
				}

				// run plugins after data poll
				if task.Name == "data" {

					pluginStart = time.Now()

					for k, v := range c.Plugins {
						for _, plg := range v {
							if pluginData, err := plg.Run(data[k]); err != nil {
								c.Logger.Error().Err(err).Str("plugin", plg.GetName()).Send()
							} else if pluginData != nil {
								results = append(results, pluginData...)
								c.Logger.Debug().Msgf("plugin [%s] added (%d) data", plg.GetName(), len(pluginData))
							} else {
								c.Logger.Debug().Msgf("plugin [%s]: completed", plg.GetName())
							}
						}
					}

					pluginTime = time.Since(pluginStart)
					_ = c.Metadata.LazySetValueInt64("plugin_time", task.Name, pluginTime.Microseconds())
				}
			}
			if task.Name == "data" {
				c.logMetadata()
			}

			// update task metadata
			_ = c.Metadata.LazySetValueInt64("poll_time", task.Name, task.GetDuration().Microseconds())
			_ = c.Metadata.LazySetValueInt64("task_time", task.Name, taskTime.Microseconds())
		}

		// pass results to exporters

		c.Logger.Debug().Msgf("exporting collected (%d) data", len(results))

		// @TODO better handling when exporter is standby/failed state
		for _, e := range c.Exporters {
			if code, status, reason := e.GetStatus(); code != 0 {
				c.Logger.Warn().Msgf("exporter [%s] down (%d - %s) (%s), skip export", e.GetName(), code, status, reason)
				continue
			}

			if err := e.Export(c.Metadata); err != nil {
				c.Logger.Warn().Err(err).Str("exporter", e.GetName()).Msg("Unable to export metadata")
			}

			// continue if metadata failed, since it might be specific to metadata
			for _, data := range results {
				if data.IsExportable() {
					if err := e.Export(data); err != nil {
						c.Logger.Error().Stack().Err(err).Msgf("export data to [%s]:", e.GetName())
						break
					}
				} else {
					c.Logger.Debug().Msgf("skipped data (%s) (%s) - set non-exportable", data.UUID, data.Object)
				}
			}
		}

		if nd := c.Schedule.NextDue(); nd > 0 {
			c.Logger.Debug().Msgf("sleeping %s until next poll", nd.String()) //DEBUG
			c.Schedule.Sleep()
			// log if lagging by more than 50 ms
			// < is used since larger durations are more negative
		} else if nd.Milliseconds() <= -50 && !c.Schedule.IsStandBy() {
			c.Logger.Warn().
				Str("lag", (-nd).String()).
				Msg("lagging behind schedule")
		}
	}
}

func (c *AbstractCollector) logMetadata() {
	metrics := c.Metadata.GetMetrics()
	info := c.Logger.Info()
	dataInstance := c.Metadata.GetInstance("data")
	if dataInstance == nil {
		return
	}
	for _, metric := range metrics {
		mName := metric.GetName()
		if mName == "poll_time" || mName == "task_time" {
			// don't log these since they're covered by other durations
			continue
		}
		value, _ := metric.GetValueFloat64(dataInstance)
		if strings.HasSuffix(mName, "_time") {
			// convert microseconds to milliseconds and names ending with _time into -> *Ms
			v := int64(math.Round(value / 1000))
			info.Int64(mName[0:len(mName)-5]+"Ms", v)
		} else {
			info.Int64(mName, int64(value))
		}
	}
	info.Msg("Collected")
}

// GetName returns name of the collector
func (c *AbstractCollector) GetName() string {
	return c.Name
}

// GetObject returns object of the collector
func (c *AbstractCollector) GetObject() string {
	return c.Object
}

// GetCollectCount retrieves and resets count of collected data
// this and next method are only to report the poller
// how much data we have collected (independent of poll interval)
func (c *AbstractCollector) GetCollectCount() uint64 {
	c.countMux.Lock()
	count := c.collectCount
	c.collectCount = 0
	c.countMux.Unlock()
	return count
}

// AddCollectCount adds n to collectCount atomically
func (c *AbstractCollector) AddCollectCount(n uint64) {
	c.countMux.Lock()
	c.collectCount += n
	c.countMux.Unlock()
}

// GetStatus returns current state of the collector
func (c *AbstractCollector) GetStatus() (uint8, string, string) {
	return c.Status, Status[c.Status], c.Message
}

// SetStatus sets the current state of the collector to one
// of the values defined by CollectorStatus
func (c *AbstractCollector) SetStatus(status uint8, msg string) {
	if status >= uint8(len(Status)) {
		panic("invalid status code " + strconv.Itoa(int(status)))
	}
	c.Status = status
	c.Message = msg
}

// GetParams returns the parameters of the collector
func (c *AbstractCollector) GetParams() *node.Node {
	return c.Params
}

// GetOptions returns the poller options passed to the collector
func (c *AbstractCollector) GetOptions() *options.Options {
	return c.Options
}

// SetSchedule set Schedule s as a field of the collector
func (c *AbstractCollector) SetSchedule(s *schedule.Schedule) {
	c.Schedule = s
}

// SetMatrix set Matrix m as a field of the collector
func (c *AbstractCollector) SetMatrix(m map[string]*matrix.Matrix) {
	c.Matrix = m
}

// SetMetadata set the metadata Matrix m as a field of the collector
func (c *AbstractCollector) SetMetadata(m *matrix.Matrix) {
	c.Metadata = m
}

// WantedExporters returns the list of exporters the receiver will export data to
func (c *AbstractCollector) WantedExporters(exporters []string) []string {
	return conf.GetUniqueExporters(exporters)
}

// LinkExporter appends exporter e to the list of exporters of the collector
func (c *AbstractCollector) LinkExporter(e exporter.Exporter) {
	// @TODO: add lock if we want to add exporters while collector is running
	c.Exporters = append(c.Exporters, e)
}

func (c *AbstractCollector) LoadPlugin(_ string, _ *plugin.AbstractPlugin) plugin.Plugin {
	return nil
}

// LoadPlugins loads built-in plugins or dynamically loads custom plugins
// and adds them to the collector
func (c *AbstractCollector) LoadPlugins(params *node.Node, collector Collector, key string) error {

	var p plugin.Plugin
	var abc *plugin.AbstractPlugin
	var plugins []plugin.Plugin
	c.Plugins = make(map[string][]plugin.Plugin)

	for _, x := range params.GetChildren() {

		name := x.GetNameS()
		if name == "" {
			name = x.GetContentS() // some plugins are defined as list elements others as dicts
			x.SetNameS(name)
		}

		abc = plugin.New(c.Name, c.Options, x, c.Params, c.Object)

		// case 1: available as built-in plugin
		if p = GetBuiltinPlugin(name, abc); p != nil {
			c.Logger.Debug().Msgf("loaded built-in plugin [%s]", name)
			// case 2: available as dynamic plugin
		} else {
			p = collector.LoadPlugin(name, abc)
			c.Logger.Debug().Msgf("loaded plugin [%s]", name)
		}
		if p == nil {
			continue
		}

		if err := p.Init(); err != nil {
			c.Logger.Error().Stack().Err(err).Msgf("init plugin [%s]:", name)
			return err
		}
		plugins = append(plugins, p)
	}
	c.Plugins[key] = plugins
	c.Logger.Debug().Msgf("initialized %d plugins", len(c.Plugins))
	return nil
}

// CollectAutoSupport allows a Collector to add autosupport information
func (c *AbstractCollector) CollectAutoSupport(_ *Payload) {
}
