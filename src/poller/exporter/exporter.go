package exporter

import (
	"goharvest2/poller/options"
	"goharvest2/share/matrix"
	"goharvest2/share/tree/node"
	"strconv"
	"sync"
	"sync/atomic"
)

type Exporter interface {
	//New(string, *yaml.Node, *structs.Options) Collector
	Init() error
	GetClass() string
	GetName() string
	GetCount() uint64
	AddCount(int)
	GetStatus() (int, string, string)
	Export(*matrix.Matrix) error
}

var ExporterStatus = [3]string{
	"up",
	"standby",
	"failed",
}

type AbstractExporter struct {
	Name     string
	Class    string
	Prefix   string
	Status   int
	Message  string
	Count    uint64
	Options  *options.Options
	Params   *node.Node
	Metadata *matrix.Matrix
	*sync.Mutex
}

func New(c, n string, o *options.Options, p *node.Node) *AbstractExporter {
	abc := AbstractExporter{
		Name:    n,
		Class:   c,
		Options: o,
		Params:  p,
		Prefix:  "(exporter) (" + n + ")",
		Mutex:   &sync.Mutex{},
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

	if _, err := e.Metadata.AddMetricInt64("time"); err != nil {
		return err
	}
	if _, err := e.Metadata.AddMetricUint64("count"); err != nil {
		return err
	}

	//e.Metadata.AddLabel("task", "")
	if instance, err := e.Metadata.AddInstance("render"); err == nil {
		instance.SetLabel("task", "render")
	} else {
		return err
	}

	if err := e.Metadata.Reset(); err != nil {
		return err
	}

	e.SetStatus(0, "initialized")
	return nil
}

func (e *AbstractExporter) GetClass() string {
	return e.Class
}

func (e *AbstractExporter) GetName() string {
	return e.Name
}

func (e *AbstractExporter) GetCount() uint64 {
	count := e.Count
	atomic.StoreUint64(&e.Count, 0)
	return count
}

func (e *AbstractExporter) AddCount(n int) {
	atomic.AddUint64(&e.Count, uint64(n))
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

/*
func (e *AbstractExporter) Export(data *matrix.Matrix) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.ExportData(data)
}

func (e *AbstractExporter) ExportData(data *matrix.Matrix) error {
	panic(e.Class + " did not implement ExportData()")
}

func (e *AbstractExporter) Render(data *matrix.Matrix) ([][]byte, error) {
	panic(e.Class + " did not implement Render()")
}*/
