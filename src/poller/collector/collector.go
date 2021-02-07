package collector

import (
	"sync"
	"os"
	"strings"
	"strconv"
	"path"
	
	"goharvest2/poller/struct/yaml"
	"goharvest2/poller/struct/options"
	"goharvest2/poller/struct/matrix"
	"goharvest2/poller/schedule"
	"goharvest2/poller/exporter"
	"goharvest2/poller/util"
	"goharvest2/poller/util/logger"
	"goharvest2/poller/errors"
	"goharvest2/poller/collector/plugin"
)

var Log *logger.Logger = logger.New(1, "")

type Collector interface {
	Init() error
	Start(*sync.WaitGroup)
	GetName() string
	GetObject() string
	IsUp() bool
	WantedExporters() []string
	LinkExporter(exporter.Exporter)
}


type AbstractCollector struct {
	Name string
	Object string
	Status string
	Message string
	Options *options.Options
	Params *yaml.Node
	Data *matrix.Matrix
	Metadata *matrix.Matrix
	Exporters []exporter.Exporter
	Plugins []plugin.Plugin
	Schedule *schedule.Schedule
	//plugins []plugin.Plugin
}

func New(name, object string, options *options.Options, params *yaml.Node) *AbstractCollector {
	c := AbstractCollector{
		Name: name,
		Object: object,
		Options: options,
		Params: params,
	}
	return &c
}


func (c *AbstractCollector) InitAbc() error {

	/* Initialize schedule and tasks (polls) */
	items := c.Params.GetChild("schedule")
	if items == nil || len(items.GetChildren()) == 0 {
		return errors.New(errors.INVALID_PARAM, "schedule")
	}

	c.Schedule = schedule.New()
	for _, task := range items.GetChildren() {
		err := c.Schedule.AddTaskString(task.Name, task.Value, nil)
		if err != nil {
			return errors.New(errors.INVALID_PARAM, "schedule (" + task.Name + "): " + err.Error())
		}
	}

	/* Initialize Matrix, the container of collected data */
	c.Data = matrix.New(c.Name, c.Object, "")
	if expo := c.Params.GetChild("export_options"); expo != nil {
		c.Data.SetExportOptions(expo)
	} else {
		c.Data.SetExportOptions(matrix.DefaultExportOptions())
	}
	c.Data.SetGlobalLabel("datacenter", c.Params.GetChildValue("datacenter"))


	/* Initialize Plugins */
	if plugins := c.Params.GetChild("plugins"); plugins != nil {
		c.LoadPlugins(plugins)
	}

	/* Initialize metadata */
	c.Metadata = matrix.New(c.Name, c.Object, "")
	c.Metadata.IsMetadata = true
	c.Metadata.MetadataType = "collector"
	c.Metadata.MetadataObject = "task"

	hostname, _ := os.Hostname()
	c.Metadata.SetGlobalLabel("hostname", hostname)
	c.Metadata.SetGlobalLabel("version", c.Options.Version)
	c.Metadata.SetGlobalLabel("poller", c.Options.Poller)
	c.Metadata.SetGlobalLabel("collector", c.Name)
	c.Metadata.SetGlobalLabel("object", c.Object)

	c.Metadata.AddMetric("poll_time", "poll_time", true)
	c.Metadata.AddMetric("count", "count", true)
	c.Metadata.AddLabelName("task")
	c.Metadata.AddLabelName("interval")

	c.Metadata.SetExportOptions(matrix.DefaultExportOptions())

	/* each task we run is an "instance" */
	for _, task := range c.Schedule.GetTasks() {
		instance, _ := c.Metadata.AddInstance(task)
		c.Metadata.SetInstanceLabel(instance, "task", task)
		s := c.Schedule.GetInterval(task).Seconds()
		c.Metadata.SetInstanceLabel(instance, "interval", strconv.FormatFloat(s, 'f', 4, 32))
	}

	/* initialize underlaying arrays */
	err := c.Metadata.InitData()

	return err
}

func (c *AbstractCollector) Start(wg *sync.WaitGroup) {
	panic(c.Name + " has not implemented Start()")
}

func (c *AbstractCollector) GetName() string {
	return c.Name
}

func (c *AbstractCollector) GetObject() string {
	return c.Object
}

func (c *AbstractCollector) IsUp() bool {
	return true
}

func (c *AbstractCollector) WantedExporters() []string {
	var names []string
	if e := c.Params.GetChild("exporters"); e != nil {
		names = e.Values
	}
	return names
}

func (c *AbstractCollector) LinkExporter(e exporter.Exporter) {
	// @TODO: add lock if we want to add exporters while collector is running
	Log.Info("Adding exporter [%s:%s]", e.GetClass(), e.GetName())
	c.Exporters = append(c.Exporters, e)
}

func (c *AbstractCollector) LoadPlugins(params *yaml.Node) error {

	for _, x := range params.GetChildren() {
		name := x.Name

		binpath := path.Join(c.Options.Path, "bin", "plugins", strings.ToLower(c.Name))

		module, err := util.LoadFromModule(binpath, strings.ToLower(name), "New")
		if err != nil {
			Log.Error("load plugin [%s]: %v", name, err)
			return err
		}

		NewFunc, ok := module.(func(string, *options.Options, *yaml.Node, *yaml.Node) plugin.Plugin)
		if !ok {
			Log.Error("load plugin [%s]: New() has not expected signature", name)
			return errors.New(errors.ERR_DLOAD, "New()")
		}

		p := NewFunc(c.Name, c.Options, x, c.Params)
		if err := p.Init(); err != nil {
			Log.Error("init plugin [%s]: %v", name, err)
			return errors.New(errors.ERR_DLOAD, "Init()")
		}

		c.Plugins = append(c.Plugins, p)
	}
	Log.Debug("initialized %d plugins", len(c.Plugins))
	return nil
}