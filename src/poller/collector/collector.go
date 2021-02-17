package collector

import (
	"sync"
	"strings"
	"strconv"
	"path"
	"reflect"
	"time"

	"goharvest2/share/logger"
	"goharvest2/share/util"
	"goharvest2/share/tree/node"
	"goharvest2/share/errors"

	"goharvest2/poller/struct/options"
	"goharvest2/poller/struct/matrix"
	"goharvest2/poller/schedule"
	"goharvest2/poller/exporter"
	"goharvest2/poller/collector/plugin"
)

type Collector interface {
	Init() error
	Start(*sync.WaitGroup)
	GetName() string
	GetObject() string
	GetParams() *node.Node
	GetOptions() *options.Options
	GetStatus() (int, string, string)
	SetStatus(int, string)
	SetSchedule(*schedule.Schedule)
	SetData(*matrix.Matrix)
	SetMetadata(*matrix.Matrix)
	WantedExporters() []string
	LinkExporter(exporter.Exporter)
	LoadPlugins(*node.Node) error
}

var CollectorStatus = [4]string{
	"undefined",
	"up",
	"standby",
	"failed",
}

type AbstractCollector struct {
	Name string
	Prefix string
	Object string
	Status int
	Message string
	Options *options.Options
	Params *node.Node
	Schedule *schedule.Schedule
	Data *matrix.Matrix
	Metadata *matrix.Matrix
	Exporters []exporter.Exporter
	Plugins []plugin.Plugin
}

func New(name, object string, options *options.Options, params *node.Node) *AbstractCollector {
	c := AbstractCollector{
		Name: name,
		Object: object,
		Options: options,
		Params: params,
	}
	c.Prefix = "(collector) (" + name + ":" + object + ")"

	return &c
}

// This is a func not method to enforce "inheritance"
// A collector can to choose to call this function
// inside its Init method, or leave it to be called
// by the poller during dynamic load
func Init(c Collector) error {

	params := c.GetParams()
	options := c.GetOptions()
	name := c.GetName()
	object := c.GetObject()

	/* Initialize schedule and tasks (polls) */
	tasks := params.GetChildS("schedule")
	if tasks == nil || len(tasks.GetChildren()) == 0 {
		return errors.New(errors.MISSING_PARAM, "schedule")
	}

	s := schedule.New()

	// Each task will be mapped to a collector method
	// Example: "data" will be alligned to method PollData()
	for _, task := range tasks.GetChildren() {

		method_name := "Poll"+strings.Title(task.GetNameS())

		if m := reflect.ValueOf(c).MethodByName(method_name); m.IsValid() {
			if foo, ok := m.Interface().(func() (*matrix.Matrix, error)); ok {
				if err := s.AddTaskString(task.GetNameS(), task.GetContentS(), foo); err == nil {
					//logger.Debug(c.Prefix, "scheduled task [%s] with %s interval", task.Name, task.GetInterval().String())
					;
				} else {
					return errors.New(errors.INVALID_PARAM, "schedule (" + task.GetNameS() + "): " + err.Error())
				}
			} else {
				return errors.New(errors.ERR_IMPLEMENT, method_name + " has not signature 'func() (*matrix.Matrix, error)'")
			}
		} else {
			return errors.New(errors.ERR_IMPLEMENT, method_name)
		}
	}
	c.SetSchedule(s)

	/* Initialize Matrix, the container of collected data */
	data := matrix.New(name, object, "")
	if export_options := params.GetChildS("export_options"); export_options != nil {
		data.SetExportOptions(export_options)
	} else {
		data.SetExportOptions(matrix.DefaultExportOptions())
	}
	data.SetGlobalLabel("datacenter", params.GetChildContentS("datacenter"))
	c.SetData(data)

	/* Initialize Plugins */
	if plugins := params.GetChildS("plugins"); plugins != nil {
		if err := c.LoadPlugins(plugins); err != nil {
			return err
		}
	}

	/* Initialize metadata */
	md := matrix.New(name, object, "")
	md.IsMetadata = true
	md.MetadataType = "collector"
	md.MetadataObject = "task"

	md.SetGlobalLabel("hostname", options.Hostname)
	md.SetGlobalLabel("version", options.Version)
	md.SetGlobalLabel("poller", options.Poller)
	md.SetGlobalLabel("collector", name)
	md.SetGlobalLabel("object", object)

	md.AddMetric("poll_time", "poll_time", true)
	md.AddMetric("api_time", "api_time", true)
	md.AddMetric("parse_time", "parse_time", true)
	md.AddMetric("calc_time", "calc_time", true)
	md.AddMetric("count", "count", true)
	md.AddLabelName("task")
	md.AddLabelName("interval")

	/* each task we run is an "instance" */
	for _, task := range s.GetTasks() {
		instance, _ := md.AddInstance(task.Name)
		md.SetInstanceLabel(instance, "task", task.Name)
		t := task.GetInterval().Seconds()
		md.SetInstanceLabel(instance, "interval", strconv.FormatFloat(t, 'f', 4, 32))
	}

	md.SetExportOptions(matrix.DefaultExportOptions())

	/* initialize underlaying arrays */
	if err := md.InitData(); err != nil {
		return err
	}

	c.SetMetadata(md)
	c.SetStatus(1, "")

	return nil
}

func (c *AbstractCollector) Start(wg *sync.WaitGroup) {

	defer wg.Done()

	// keep track of connection errors
	// to increment time before retry
	retry_delay := 1

	for {

		c.Metadata.InitData() // @TODO handle err (can occur if collector messed up)

		results := make([]*matrix.Matrix, 0)

		for _, task := range c.Schedule.GetTasks() {
			if !task.IsDue() {
				continue
			}
			data, err := task.Run()

			if err != nil {

				if !c.Schedule.IsStandBy() {
					logger.Debug(c.Prefix, "handling error during [%s] poll...", task.Name)
				}
				switch {
				case errors.IsErr(err, errors.ERR_CONNECTION):
					if retry_delay < 1024 {
						retry_delay *= 4
					}
					if !c.Schedule.IsStandBy() {
						logger.Error(c.Prefix, err.Error())
						logger.Error(c.Prefix, "target system unreachable, entering standby mode (retry to connect in %d s)", retry_delay)
					}
					c.Schedule.SetStandByMode(task.Name, time.Duration(retry_delay) * time.Second)
					c.SetStatus(2, err.Error())
				case errors.IsErr(err, errors.ERR_NO_INSTANCE):
					c.Schedule.SetStandByMode(task.Name, 5 * time.Minute)
					c.SetStatus(2, err.Error())
					logger.Error(c.Prefix, "no [%s] instances on system, entering standby mode", c.Object)
				case errors.IsErr(err, errors.ERR_NO_METRIC):
					c.SetStatus(2, err.Error())
					c.Schedule.SetStandByMode(task.Name, 1 * time.Hour)
					logger.Error(c.Prefix, "no [%s] metrics on system, entering standby mode", c.Object)
				default:
					// enter failed state
					c.SetStatus(3, err.Error())
					logger.Error(c.Prefix, err.Error())
					return
				}
				// don't continue on errors
				break
			} else if c.Schedule.IsStandBy() {
				// recover from standby mode
				c.SetStatus(1, "")
				c.Schedule.Recover()
				logger.Info(c.Prefix, "recovered from standby mode, back to normal schedule")
			}

			c.Metadata.SetValueSS("poll_time", task.Name, float32(task.Runtime().Seconds()))

			if data != nil {
				results = append(results, data)

				if task.Name == "data" {
					for _, plg := range c.Plugins {
						if plg_data_slice, err := plg.Run(data); err != nil {
							logger.Error(c.Prefix, "plugin [%s]: %s", plg.GetName(), err.Error())
						} else if plg_data_slice != nil {
							results = append(results, plg_data_slice...)
							logger.Debug(c.Prefix, "plugin [%s] added (%d) data", plg.GetName(), len(plg_data_slice))
						} else {
							logger.Debug(c.Prefix, "plugin [%s]: completed", plg.GetName())
						}
					}
				}
			}
		}

		logger.Debug(c.Prefix, "exporting collected (%d) data", len(results))

		// @TODO better handling when exporter is standby/failed state
		for _, e := range c.Exporters {
			if status, _, _ := e.GetStatus(); status != 1 {
				logger.Warn(c.Prefix, "exporter [%s] down, skipping export", e.GetName())
			} else if err := e.Export(c.Metadata); err != nil {
				logger.Warn(c.Prefix, "export metadata to [%s]: %s", e.GetName(), err.Error())
			}
			// continue if metadata failed, since it might be specific to metadata
			for _, data := range results {
				if err := e.Export(data); err != nil {
					logger.Error(c.Prefix, "export data to [%s]: %s", e.GetName(), err.Error())
					break
				}
			}
		}

		logger.Debug(c.Prefix, "sleeping %s until next poll", c.Schedule.NextDue().String())
		c.Schedule.Sleep()
	}
}

func (c *AbstractCollector) GetName() string {
	return c.Name
}

func (c *AbstractCollector) GetObject() string {
	return c.Object
}

func (c *AbstractCollector) GetStatus() (int, string, string) {
	return c.Status, CollectorStatus[c.Status], c.Message
}

func (c *AbstractCollector) SetStatus(status int, msg string) {
	if status < 0 || status >= len(CollectorStatus) {
		panic("invalid status code " + strconv.Itoa(status))
	}
	c.Status = status
	c.Message = msg
}

func (c *AbstractCollector) GetParams() *node.Node {
	return c.Params
}

func (c *AbstractCollector) GetOptions() *options.Options {
	return c.Options
}

func (c *AbstractCollector) SetSchedule(s *schedule.Schedule) {
	c.Schedule = s
}

func (c *AbstractCollector) SetData(m *matrix.Matrix) {
	c.Data = m
}

func (c *AbstractCollector) SetMetadata(m *matrix.Matrix) {
	c.Metadata = m
}

func (c *AbstractCollector) WantedExporters() []string {
	var names []string
	if e := c.Params.GetChildS("exporters"); e != nil {
		names = e.GetAllChildContentS()
	}
	return names
}

func (c *AbstractCollector) LinkExporter(e exporter.Exporter) {
	// @TODO: add lock if we want to add exporters while collector is running
	//logger.Info(c.LongName, "Adding exporter [%s:%s]", e.GetClass(), e.GetName())
	c.Exporters = append(c.Exporters, e)
}

func (c *AbstractCollector) LoadPlugins(params *node.Node) error {

	for _, x := range params.GetChildren() {
		name := x.GetNameS()

		binpath := path.Join(c.Options.Path, "bin", "plugins", strings.ToLower(c.Name))

		module, err := util.LoadFuncFromModule(binpath, strings.ToLower(name), "New")
		if err != nil {
			//logger.Error(c.LongName, "load plugin [%s]: %v", name, err)
			return errors.New(errors.ERR_DLOAD, name + ": " + err.Error())
		}

		NewFunc, ok := module.(func(*plugin.AbstractPlugin) plugin.Plugin)
		if !ok {
			//logger.Error(c.LongName, "load plugin [%s]: New() has not expected signature", name)
			return errors.New(errors.ERR_DLOAD, name + ": New()")
		}

		p := NewFunc(plugin.New(c.Name, c.Options, x, c.Params))
		if err := p.Init(); err != nil {
			//logger.Error(c.LongName, "init plugin [%s]: %v", name, err)
			return errors.New(errors.ERR_DLOAD, name + ": Init(): " + err.Error())
		}

		c.Plugins = append(c.Plugins, p)
	}
	//logger.Debug(c.LongName, "initialized %d plugins", len(c.Plugins))
	return nil
}
