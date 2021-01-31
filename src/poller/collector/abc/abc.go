package abc

import (
	"os"
	"strconv"
	"sync"
	"poller/share/logger"
	"poller/yaml"
	"poller/exporter"
	"poller/errors"
	"poller/schedule"
	"poller/structs/matrix"
	"poller/structs/opts"
)

var Log *logger.Logger = logger.New(1, "")

type AbstractCollector struct {
	Name string
	Object string
	Options *opts.Opts
	Params *yaml.Node
	Data *matrix.Matrix
	Metadata *matrix.Matrix
	Exporters []exporter.Exporter
	Schedule *schedule.Schedule
	//plugins []plugin.Plugin
}

func New(name, object string, options *opts.Opts, params *yaml.Node) *AbstractCollector {
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
		return errors.MissingParam("schedule")
	}

	c.Schedule = schedule.New()
	for _, task := range items.GetChildren() {
		err := c.Schedule.AddTaskString(task.Name, task.Value, nil)
		if err != nil {
			return errors.InvalidParam("schedule (" + task.Name + "): " + err.Error())
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
