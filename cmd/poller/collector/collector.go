/*
	Copyright NetApp Inc, 2021 All rights reserved

	Package collector provides the Collector interface and
	the AbstractCollector type which implements most basic
	attributes.

	A Harvest collector should normally "inherit" all these
	attributes and implement only the PollData function.
	The AbstractCollector will make sure that the collector
	is properly initializied, metadata are updated and
	data poll(s) and plugins run as scheduled. The collector
	can also choose to override any of the attributes
	implemented by AbstractCollector.
*/
package collector

import (
	"goharvest2/pkg/config"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"goharvest2/pkg/dload"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"

	"goharvest2/cmd/poller/exporter"
	"goharvest2/cmd/poller/options"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/poller/schedule"
)

// Collector defines the attributes of a collector
// The poll functions (PollData, PollInstance, etc)
// are not part of the interface and are linked dynamically
// All required functions are implemented by AbstractCollector
//
// Note that many of the functions required by the interface
// are only there to facilitate "inheritance" through AbstractCollector.
type Collector interface {
	Init() error
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
	SetMatrix(*matrix.Matrix)
	SetMetadata(*matrix.Matrix)
	WantedExporters(configFp string) []string
	LinkExporter(exporter.Exporter)
	LoadPlugins(*node.Node) error
}

// CollectorStatus defines the possible states of a collector
var CollectorStatus = [3]string{
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
	Object  string           // object of the collector, describes what thet collector is collecting
	Prefix  string           // prefix used for logging
	Status  uint8            // current state of th
	Message string           // reason if collector is in failed state
	Options *options.Options // poller options
	Params  *node.Node       // collector parameters
	// note that this is a merge of poller parameters, collector conf and object conf ("subtemplate")
	Schedule     *schedule.Schedule  // schedule of the collector
	Matrix       *matrix.Matrix      // the data storage of the collector
	Metadata     *matrix.Matrix      // metadata of the collector, such as poll duration, collected data points etc.
	Exporters    []exporter.Exporter // the exporters that the collector will emit data to
	Plugins      []plugin.Plugin     // built-in or custom plugins
	collectCount uint64              // count of collected data points
	// this is different from what the collector will have in its metadata, since this variable
	// holds count independent of the poll interval of the collector, used to give stats to Poller
	countMux *sync.Mutex // used for atomic access to collectCount
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
		Params:   params,
		countMux: &sync.Mutex{},
	}
	c.Prefix = "(collector) (" + name + ":" + object + ")"

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
// then we expect that the collector has methods PollDdta and PollInstance
// that need to be invoked every 10 and 20 seconds respectively.
// Names of the polls are arbitrary, only "ddta" is a special case, since
// plugins are executed after the data poll (this might change).
func Init(c Collector) error {

	params := c.GetParams()
	options := c.GetOptions()
	name := c.GetName()
	object := c.GetObject()

	// Initialize schedule and tasks (polls)
	tasks := params.GetChildS("schedule")
	if tasks == nil || len(tasks.GetChildren()) == 0 {
		return errors.New(errors.MISSING_PARAM, "schedule")
	}

	s := schedule.New()

	// Each task will be mapped to a collector method
	// Example: "data" will be alligned to method PollData()
	for _, task := range tasks.GetChildren() {

		methodName := "Poll" + strings.Title(task.GetNameS())

		if m := reflect.ValueOf(c).MethodByName(methodName); m.IsValid() {
			if foo, ok := m.Interface().(func() (*matrix.Matrix, error)); ok {
				if err := s.NewTaskString(task.GetNameS(), task.GetContentS(), foo); err != nil {
					return errors.New(errors.INVALID_PARAM, "schedule ("+task.GetNameS()+"): "+err.Error())
				}
			} else {
				return errors.New(errors.ERR_IMPLEMENT, methodName+" has not signature 'func() (*matrix.Matrix, error)'")
			}
		} else {
			return errors.New(errors.ERR_IMPLEMENT, methodName)
		}
	}
	c.SetSchedule(s)

	// Initialize Matrix, the container of collected data
	mx := matrix.New(name, object)
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

	c.SetMatrix(mx)

	// Initialize Plugins
	if plugins := params.GetChildS("plugins"); plugins != nil {
		if err := c.LoadPlugins(plugins); err != nil {
			return err
		}
	}

	// Initialize metadata
	md := matrix.New(name, "metadata_collector")

	md.SetGlobalLabel("hostname", options.Hostname)
	md.SetGlobalLabel("version", options.Version)
	md.SetGlobalLabel("poller", options.Poller)
	md.SetGlobalLabel("collector", name)
	md.SetGlobalLabel("object", object)

	md.NewMetricInt64("poll_time")
	md.NewMetricInt64("task_time")
	md.NewMetricInt64("api_time")
	md.NewMetricInt64("parse_time")
	md.NewMetricInt64("calc_time")
	md.NewMetricInt64("plugin_time")
	md.NewMetricInt64("content_length")
	md.NewMetricFloat64("api_time_percent")
	md.NewMetricUint64("count")
	//md.AddLabel("task", "")
	//md.AddLabel("interval", "")

	// add tasks of the collecor as metadata instances
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

// Start will run the collector in an infinity loop
func (me *AbstractCollector) Start(wg *sync.WaitGroup) {

	defer wg.Done()

	// keep track of connection errors
	// to increment time before retry
	// @TODO add to metadata
	retryDelay := 1
	me.SetStatus(0, "running")

	for {

		me.Metadata.Reset()

		results := make([]*matrix.Matrix, 0)

		// run all scheduled tasks
		for _, task := range me.Schedule.GetTasks() {
			if !task.IsDue() {
				continue
			}

			var (
				start, pluginStart   time.Time
				taskTime, pluginTime time.Duration
			)

			start = time.Now()
			data, err := task.Run()
			taskTime = time.Since(start)

			// poll returned error, try to understand what to do
			if err != nil {

				if !me.Schedule.IsStandBy() {
					logger.Debug(me.Prefix, "handling error during [%s] poll...", task.Name)
				}
				switch {
				// target system is unreachable
				// enter standby mode and retry with some delay that will be increased if we fail again
				case errors.IsErr(err, errors.ERR_CONNECTION):
					if retryDelay < 1024 {
						retryDelay *= 4
					}
					if !me.Schedule.IsStandBy() {
						//logger.Error(me.Prefix, err.Error())
						logger.Warn(me.Prefix, "target unreachable, entering standby mode (retry to connect in %d s)", retryDelay)
					}
					me.Schedule.SetStandByMode(task, time.Duration(retryDelay)*time.Second)
					me.SetStatus(1, errors.ERR_CONNECTION)
				// there are no instances to collect
				case errors.IsErr(err, errors.ERR_NO_INSTANCE):
					me.Schedule.SetStandByMode(task, 5*time.Minute)
					me.SetStatus(1, errors.ERR_NO_INSTANCE)
					logger.Info(me.Prefix, "no [%s] instances on system, entering standby mode", me.Object)
				// no metrics available
				case errors.IsErr(err, errors.ERR_NO_METRIC):
					me.SetStatus(1, errors.ERR_NO_METRIC)
					me.Schedule.SetStandByMode(task, 1*time.Hour)
					logger.Warn(me.Prefix, "no [%s] metrics on system, entering standby mode", me.Object)
				// not an error we are expecting, so enter failed state and terminate
				default:
					logger.Error(me.Prefix, err.Error())
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
				me.SetStatus(0, "running")
				logger.Info(me.Prefix, "recovered from standby mode, back to normal schedule")
			}

			if data != nil {
				results = append(results, data)

				// run plugins after data poll
				if task.Name == "data" {

					pluginStart = time.Now()

					for _, plg := range me.Plugins {
						if pluginData, err := plg.Run(data); err != nil {
							logger.Error(me.Prefix, "plugin [%s]: %s", plg.GetName(), err.Error())
						} else if pluginData != nil {
							results = append(results, pluginData...)
							logger.Debug(me.Prefix, "plugin [%s] added (%d) data", plg.GetName(), len(pluginData))
						} else {
							logger.Debug(me.Prefix, "plugin [%s]: completed", plg.GetName())
						}
					}

					pluginTime = time.Since(pluginStart)
					me.Metadata.LazySetValueInt64("plugin_time", task.Name, pluginTime.Microseconds())
				}

				// update some metadata
				me.Metadata.LazySetValueInt64("poll_time", task.Name, task.GetDuration().Microseconds())
				me.Metadata.LazySetValueInt64("task_time", task.Name, taskTime.Microseconds())

				if apiTime, ok := me.Metadata.LazyGetValueInt64("api_time", task.Name); ok && apiTime != 0 {
					me.Metadata.LazySetValueFloat64("api_time_percent", task.Name, float64(apiTime)/float64(taskTime.Microseconds())*100)
				}

			}
		}

		// pass results to exporters

		logger.Debug(me.Prefix, "exporting collected (%d) data", len(results))

		// @TODO better handling when exporter is standby/failed state
		for _, e := range me.Exporters {
			if code, status, reason := e.GetStatus(); code != 0 {
				logger.Warn(me.Prefix, "exporter [%s] down (%d - %s) (%s), skip export", e.GetName(), code, status, reason)
				continue
			}

			if err := e.Export(me.Metadata); err != nil {
				logger.Warn(me.Prefix, "export metadata to [%s]: %s", e.GetName(), err.Error())
			}

			// continue if metadata failed, since it might be specific to metadata
			for _, data := range results {
				if data.IsExportable() {
					if err := e.Export(data); err != nil {
						logger.Error(me.Prefix, "export data to [%s]: %s", e.GetName(), err.Error())
						break
					}
				} else {
					logger.Debug(me.Prefix, "skipped data (%s) (%s) - set non-exportable", data.UUID, data.Object)
				}
			}
		}

		if nd := me.Schedule.NextDue(); nd > 0 {
			logger.Debug(me.Prefix, "sleeping %s until next poll", nd.String()) //DEBUG
			me.Schedule.Sleep()
		} else if nd.Milliseconds() > -50 { // avoid warning for small delays
			logger.Warn(me.Prefix, "lagging behind schedule %s", (-nd).String())
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
	return me.Status, CollectorStatus[me.Status], me.Message
}

// SetStatus sets the current state of the collector to one
// of the values defined by CollectorStatus
func (me *AbstractCollector) SetStatus(status uint8, msg string) {
	if status < 0 || status >= uint8(len(CollectorStatus)) {
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
func (me *AbstractCollector) SetMatrix(m *matrix.Matrix) {
	me.Matrix = m
}

// SetMetadata set the metadata Matrix m as a field of the collector
func (me *AbstractCollector) SetMetadata(m *matrix.Matrix) {
	me.Metadata = m
}

// WantedExporters retrievs the names of the exporters to which the collector
// needs to export data
func (me *AbstractCollector) WantedExporters(configFp string) []string {
	names, err := config.GetUniqueExporters(me.Params, configFp)
	if err != nil {
		logger.Error(me.Prefix, "Error while fetching exporters %v", err)
	}
	return names
}

// LinkExporter appends exporter e to the list of exporters of the collector
func (me *AbstractCollector) LinkExporter(e exporter.Exporter) {
	// @TODO: add lock if we want to add exporters while collector is running
	me.Exporters = append(me.Exporters, e)
}

// LoadPlugins loads built-in plugins or dynamically loads custom plugins
// and adds them to the collector
func (me *AbstractCollector) LoadPlugins(params *node.Node) error {

	var p plugin.Plugin
	var abc *plugin.AbstractPlugin

	for _, x := range params.GetChildren() {

		name := x.GetNameS()
		if name == "" {
			name = x.GetContentS() // some plugins are defined as list elements others as dicts
			x.SetNameS(name)
		}

		abc = plugin.New(me.Name, me.Options, x, me.Params)

		// case 1: available as built-in plugin
		if p = getBuiltinPlugin(name, abc); p != nil {
			logger.Debug(me.Prefix, "loaded built-in plugin [%s]", name)
			// case 2: available as dynamic plugin
		} else {
			binpath := path.Join(me.Options.HomePath, "bin", "plugins", strings.ToLower(me.Name))
			module, err := dload.LoadFuncFromModule(binpath, strings.ToLower(name), "New")
			if err != nil {
				//logger.Error(c.LongName, "load plugin [%s]: %v", name, err)
				return errors.New(errors.ERR_DLOAD, "plugin "+name+": "+err.Error())
			}

			NewFunc, ok := module.(func(*plugin.AbstractPlugin) plugin.Plugin)
			if !ok {
				//logger.Error(c.LongName, "load plugin [%s]: New() has not expected signature", name)
				return errors.New(errors.ERR_DLOAD, name+": New()")
			}
			p = NewFunc(abc)
			logger.Debug(me.Prefix, "loaded dynamic plugin [%s]", name)
		}

		if err := p.Init(); err != nil {
			logger.Error(me.Prefix, "init plugin [%s]: %v", name, err)
			return err
		}
		me.Plugins = append(me.Plugins, p)
	}
	logger.Debug(me.Prefix, "initialized %d plugins", len(me.Plugins))
	return nil
}
