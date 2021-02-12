package exporter

import (
	"sync"
	"strconv"
	"goharvest2/poller/struct/yaml"
	"goharvest2/poller/struct/matrix"
	"goharvest2/poller/struct/options"
)

type Exporter interface {
	//New(string, *yaml.Node, *structs.Options) Collector
	Init() error
	GetClass() string
	GetName() string
	GetStatus() (int, string, string)
	Export(*matrix.Matrix) error
}

var ExporterStatus = [4]string{
	"undefined",
	"up",
	"standby",
	"failed",
}

type AbstractExporter struct {
	Name string
	Class string
	Prefix string
	Status int
	Message string
	Options *options.Options
	Params *yaml.Node
	Metadata *matrix.Matrix
	mu *sync.Mutex
}

func New(c, n string, o *options.Options, p *yaml.Node) *AbstractExporter {
	abc := AbstractExporter{
		Name: n,
		Class: c,
		Options: o,
		Params: p,
		Prefix: "(exporter) (" + n + ")",
	}
	return &abc
}

func (e *AbstractExporter) InitAbc() error {
    e.Metadata = matrix.New(e.Class, e.Name, "")
	e.Metadata.IsMetadata = true
	e.Metadata.MetadataType = "exporter"
	e.Metadata.MetadataObject = "export"
	e.Metadata.SetGlobalLabel("hostname", e.Options.Hostname)
	e.Metadata.SetGlobalLabel("version", e.Options.Version)
	e.Metadata.SetGlobalLabel("poller", e.Options.Poller)
	e.Metadata.SetGlobalLabel("exporter", e.Class)
	e.Metadata.SetGlobalLabel("target", e.Name)

	if _, err := e.Metadata.AddMetric("time", "time", true); err != nil {
        return err
    }
	if _, err := e.Metadata.AddMetric("count", "count", true); err != nil {
        return err
    }

    e.Metadata.AddLabelName("task")
    if instance, err := e.Metadata.AddInstance("render"); err == nil {
		e.Metadata.SetInstanceLabel(instance, "task", "render")
		e.Metadata.SetExportOptions(matrix.DefaultExportOptions())
	} else {
		return err
	}
	
	if err := e.Metadata.InitData(); err != nil {
		return err
	}

	e.SetStatus(1, "")
	return nil
}


func (e *AbstractExporter) GetClass() string {
	return e.Class
}

func (e *AbstractExporter) GetName() string {
	return e.Name
}

func (e *AbstractExporter) GetStatus() (int, string, string) {
	return e.Status, ExporterStatus[e.Status], e.Message
}

func (e *AbstractExporter) SetStatus(code int, msg string) {
	if code < 0 || code >= len(ExporterStatus) {
		panic("invalid status code " + strconv.Itoa(code))
	}
	e.Status = code
	e.Message = msg
}

func (e *AbstractExporter) Export(data *matrix.Matrix) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.ExportData(data)
}

func (e *AbstractExporter) ExportData(data *matrix.Matrix) error {
	panic(e.Class + " did not implement ExportData()")
}

/*
func (e *AbstractExporter) Render(data *matrix.Matrix) ([][]byte, error) {
	panic(e.Class + " did not implement Render()")
}*/