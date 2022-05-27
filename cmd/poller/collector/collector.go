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
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/netapp/harvest/v2/pkg/errors"
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
	LoadPlugins(*node.Node, string) error
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
// The important thing done here is too look what tasks are defined
// in the "schedule" parameter of the collector and create a pointer
// to the corresponding method of the collector. Example, parameter is:
//
// schedule:
//   data: 10s
//   instance: 20s
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
		return errors.New(errors.MissingParam, "schedule")
	}

	s := schedule.New()

	// Each task will be mapped to a collector method
	// Example: "data" will be aligned to method PollData()
	for _, task := range tasks.GetChildren() {

		methodName := "Poll" + strings.Title(task.GetNameS())

		if m := reflect.ValueOf(c).MethodByName(methodName); m.IsValid() {
			if foo, ok := m.Interface().(func() (map[string]*matrix.Matrix, error)); ok {
				if err := s.NewTaskString(task.GetNameS(), task.GetContentS(), foo, true, "Collector_"+c.GetName()+"_"+c.GetObject()); err != nil {
					return errors.New(errors.InvalidParam, "schedule ("+task.GetNameS()+"): "+err.Error())
				}
			} else {
				return errors.New(errors.ErrImplement, methodName+" has not signature 'func() (*matrix.Matrix, error)'")
			}
		} else {
			return errors.New(errors.ErrImplement, methodName)
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
		if err := c.LoadPlugins(plugins, c.GetObject()); err != nil {
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

	md.NewMetricInt64("poll_time")
	md.NewMetricInt64("task_time")
	md.NewMetricInt64("api_time")
	md.NewMetricInt64("parse_time")
	md.NewMetricInt64("calc_time")
	md.NewMetricInt64("plugin_time")
	md.NewMetricUint64("count")
	//md.AddLabel("task", "")
	//md.AddLabel("interval", "")

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
func (me *AbstractCollector) GetMetadata() *matrix.Matrix {
	return me.Metadata
}

func (me *AbstractCollector) GetHostModel() string {
	return me.HostModel
}

func (me *AbstractCollector) GetHostVersion() string {
	return me.HostVersion
}

func (me *AbstractCollector) GetHostUUID() string {
	return me.HostUUID
}

// Start will run the collector in an infinity loop
func (me *AbstractCollector) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			me.Logger.Error().Stack().Err(errors.New(errors.GoRoutinePanic, "")).
				Msgf("Collector panicked %s", r)
		}
	}()

	// keep track of connection errors
	// to increment time before retry
	// @TODO add to metadata
	retryDelay := 1
	me.SetStatus(0, "running")

	for {

		// We can't reset metadata here because autosupport metadata is reset
		// https://github.com/NetApp/harvest-private/issues/114 for details
		//me.Metadata.Reset()

		results := make([]*matrix.Matrix, 0)

		// run all scheduled tasks
		for _, task := range me.Schedule.GetTasks() {
			if !task.IsDue() {
				continue
			}

			if me.Schedule.IsStandBy() && !me.Schedule.IsTaskStandBy(task) {
				me.Logger.Info().
					Str("task", task.Name).
					Msg("schedule is in standby mode skipping")
				continue
			}

			var (
				start, pluginStart   time.Time
				taskTime, pluginTime time.Duration
			)

			// reset task metadata
			me.Metadata.ResetInstance(task.Name)

			start = time.Now()
			data, err := task.Run()
			taskTime = time.Since(start)

			// poll returned error, try to understand what to do
			if err != nil {

				if !me.Schedule.IsStandBy() {
					me.Logger.Debug().Msgf("handling error during [%s] poll...", task.Name)
				}
				switch {
				// target system is unreachable
				// enter standby mode and retry with some delay that will be increased if we fail again
				case errors.IsErr(err, errors.ErrConnection):
					if retryDelay < 1024 {
						retryDelay *= 4
					}
					if !me.Schedule.IsStandBy() {
						me.Logger.Warn().
							Str("task", task.Name).
							Int("retryDelay", retryDelay).
							Msgf("target unreachable, entering standby mode (retry in retryDelay s)")
					}
					me.Logger.Debug().
						Err(err).
						Int("retryDelay", retryDelay).
						Str("task", task.Name).
						Msg("Target unreachable, entering standby mode (retry in retryDelay s)")
					me.Schedule.SetStandByMode(task, time.Duration(retryDelay)*time.Second)
					me.SetStatus(1, errors.ErrConnection)
				// there are no instances to collect
				case errors.IsErr(err, errors.ErrNoInstance):
					me.Schedule.SetStandByMode(task, 5*time.Minute)
					me.SetStatus(1, errors.ErrNoInstance)
					me.Logger.Info().
						Str("task", task.Name).
						Str("object", me.Object).
						Msg("no instances of object on system, entering standby mode")
				// no metrics available
				case errors.IsErr(err, errors.ErrNoMetric):
					me.SetStatus(1, errors.ErrNoMetric)
					me.Schedule.SetStandByMode(task, 1*time.Hour)
					me.Logger.Info().
						Str("task", task.Name).
						Str("object", me.Object).
						Msg("no metrics of object on system, entering standby mode")
				// not an error we are expecting, so enter failed state and terminate
				default:
					me.Logger.Error().Stack().Err(err).Str("task", task.Name).Msg("")
					if errmsg := errors.GetClass(err); errmsg != "" {
						me.SetStatus(2, errmsg)
					} else {
						me.SetStatus(2, err.Error())
					}
					break
				}
				// stop here if we had errors
				continue
			} else if me.Schedule.IsStandBy() {
				// recover from standby mode
				me.Schedule.Recover()
				retryDelay = 1
				me.SetStatus(0, "running")
				me.Logger.Info().Str("task", task.Name).Msg("recovered from standby mode, back to normal schedule")
			} else {
				me.SetStatus(0, "running")
			}

			if data != nil {

				for _, value := range data {
					results = append(results, value)
				}

				// run plugins after data poll
				if task.Name == "data" {

					pluginStart = time.Now()

					for k, v := range me.Plugins {
						for _, plg := range v {
							if pluginData, err := plg.Run(data[k]); err != nil {
								me.Logger.Error().Stack().Err(err).Msgf("plugin [%s]: ", plg.GetName())
							} else if pluginData != nil {
								results = append(results, pluginData...)
								me.Logger.Debug().Msgf("plugin [%s] added (%d) data", plg.GetName(), len(pluginData))
							} else {
								me.Logger.Debug().Msgf("plugin [%s]: completed", plg.GetName())
							}
						}
					}

					pluginTime = time.Since(pluginStart)
					me.Metadata.LazySetValueInt64("plugin_time", task.Name, pluginTime.Microseconds())
				}
			}

			// update task metadata
			me.Metadata.LazySetValueInt64("poll_time", task.Name, task.GetDuration().Microseconds())
			me.Metadata.LazySetValueInt64("task_time", task.Name, taskTime.Microseconds())
		}

		// pass results to exporters

		me.Logger.Debug().Msgf("exporting collected (%d) data", len(results))

		// @TODO better handling when exporter is standby/failed state
		for _, e := range me.Exporters {
			if code, status, reason := e.GetStatus(); code != 0 {
				me.Logger.Warn().Msgf("exporter [%s] down (%d - %s) (%s), skip export", e.GetName(), code, status, reason)
				continue
			}

			if err := e.Export(me.Metadata); err != nil {
				me.Logger.Warn().Msgf("export metadata to [%s]: %s", e.GetName(), err.Error())
			}

			// continue if metadata failed, since it might be specific to metadata
			for _, data := range results {
				if data.IsExportable() {
					if err := e.Export(data); err != nil {
						me.Logger.Error().Stack().Err(err).Msgf("export data to [%s]:", e.GetName())
						break
					}
				} else {
					me.Logger.Debug().Msgf("skipped data (%s) (%s) - set non-exportable", data.UUID, data.Object)
				}
			}
		}

		if nd := me.Schedule.NextDue(); nd > 0 {
			me.Logger.Debug().Msgf("sleeping %s until next poll", nd.String()) //DEBUG
			me.Schedule.Sleep()
			// log if lagging by more than 50 ms
			// < is used since larger durations are more negative
		} else if nd.Milliseconds() <= -50 && !me.Schedule.IsStandBy() {
			me.Logger.Warn().Msgf("lagging behind schedule %s", (-nd).String())
		}
	}
}

// GetName returns name of the collector
func (me *AbstractCollector) GetName() string {
	return me.Name
}

// GetObject returns object of the collector
func (me *AbstractCollector) GetObject() string {
	return me.Object
}

// GetCollectCount retrieves and resets count of collected data
// this and next method are only to report the poller
// how much data we have collected (independent of poll interval)
func (me *AbstractCollector) GetCollectCount() uint64 {
	me.countMux.Lock()
	count := me.collectCount
	me.collectCount = 0
	me.countMux.Unlock()
	return count
}

// AddCollectCount adds n to collectCount atomically
func (me *AbstractCollector) AddCollectCount(n uint64) {
	me.countMux.Lock()
	me.collectCount += n
	me.countMux.Unlock()
}

// GetStatus returns current state of the collector
func (me *AbstractCollector) GetStatus() (uint8, string, string) {
	return me.Status, Status[me.Status], me.Message
}

// SetStatus sets the current state of the collector to one
// of the values defined by CollectorStatus
func (me *AbstractCollector) SetStatus(status uint8, msg string) {
	if status < 0 || status >= uint8(len(Status)) {
		panic("invalid status code " + strconv.Itoa(int(status)))
	}
	me.Status = status
	me.Message = msg
}

// GetParams returns the parameters of the collector
func (me *AbstractCollector) GetParams() *node.Node {
	return me.Params
}

// GetOptions returns the poller options passed to the collector
func (me *AbstractCollector) GetOptions() *options.Options {
	return me.Options
}

// SetSchedule set Schedule s as a field of the collector
func (me *AbstractCollector) SetSchedule(s *schedule.Schedule) {
	me.Schedule = s
}

// SetMatrix set Matrix m as a field of the collector
func (me *AbstractCollector) SetMatrix(m map[string]*matrix.Matrix) {
	me.Matrix = m
}

// SetMetadata set the metadata Matrix m as a field of the collector
func (me *AbstractCollector) SetMetadata(m *matrix.Matrix) {
	me.Metadata = m
}

// WantedExporters returns the list of exporters the receiver will export data to
func (me *AbstractCollector) WantedExporters(exporters []string) []string {
	return conf.GetUniqueExporters(exporters)
}

// LinkExporter appends exporter e to the list of exporters of the collector
func (me *AbstractCollector) LinkExporter(e exporter.Exporter) {
	// @TODO: add lock if we want to add exporters while collector is running
	me.Exporters = append(me.Exporters, e)
}

func (me *AbstractCollector) LoadPlugin(s string, abc *plugin.AbstractPlugin) plugin.Plugin {
	return nil
}

//LoadPlugins loads built-in plugins or dynamically loads custom plugins
//and adds them to the collector
func (me *AbstractCollector) LoadPlugins(params *node.Node, key string) error {

	var p plugin.Plugin
	var abc *plugin.AbstractPlugin
	var plugins []plugin.Plugin
	me.Plugins = make(map[string][]plugin.Plugin)

	for _, x := range params.GetChildren() {

		name := x.GetNameS()
		if name == "" {
			name = x.GetContentS() // some plugins are defined as list elements others as dicts
			x.SetNameS(name)
		}

		abc = plugin.New(me.Name, me.Options, x, me.Params, me.Object)

		// case 1: available as built-in plugin
		if p = GetBuiltinPlugin(name, abc); p != nil {
			me.Logger.Debug().Msgf("loaded built-in plugin [%s]", name)
			// case 2: available as dynamic plugin
		} else {
			p = me.LoadPlugin(name, abc)
			me.Logger.Debug().Msgf("loaded plugin [%s]", name)
		}
		if p == nil {
			continue
		}

		if err := p.Init(); err != nil {
			me.Logger.Error().Stack().Err(err).Msgf("init plugin [%s]:", name)
			return err
		}
		plugins = append(plugins, p)
	}
	me.Logger.Debug().Msgf("initialized %d plugins", len(me.Plugins))
	me.Plugins[key] = plugins
	return nil
}

// CollectAutoSupport allows a Collector to add autosupport information
func (me *AbstractCollector) CollectAutoSupport(_ *Payload) {
}
