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
	GetStatus() (uint8, string, string)
	IsMaster() bool
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
	Status   uint8
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

func (me *AbstractExporter) InitAbc() error {
	me.Metadata = matrix.New(me.Class, me.Name, "")
	me.Metadata.IsMetadata = true
	me.Metadata.MetadataType = "exporter"
	me.Metadata.MetadataObject = "export"
	me.Metadata.SetGlobalLabel("hostname", me.Options.Hostname)
	me.Metadata.SetGlobalLabel("version", me.Options.Version)
	me.Metadata.SetGlobalLabel("poller", me.Options.Poller)
	me.Metadata.SetGlobalLabel("exporter", me.Class)
	me.Metadata.SetGlobalLabel("target", me.Name)

	if _, err := me.Metadata.AddMetricInt64("time"); err != nil {
		return err
	}
	if _, err := me.Metadata.AddMetricUint64("count"); err != nil {
		return err
	}

	//e.Metadata.AddLabel("task", "")
	if instance, err := me.Metadata.AddInstance("render"); err == nil {
		instance.SetLabel("task", "render")
	} else {
		return err
	}

	if err := me.Metadata.Reset(); err != nil {
		return err
	}

	me.SetStatus(0, "initialized")
	return nil
}

func (me *AbstractExporter) GetClass() string {
	return me.Class
}

func (me *AbstractExporter) GetName() string {
	return me.Name
}

func (me *AbstractExporter) GetCount() uint64 {
	count := me.Count
	atomic.StoreUint64(&me.Count, 0)
	return count
}

func (me *AbstractExporter) AddCount(n int) {
	atomic.AddUint64(&me.Count, uint64(n))
}

func (me *AbstractExporter) GetStatus() (uint8, string, string) {
	return me.Status, ExporterStatus[me.Status], me.Message
}

func (me *AbstractExporter) SetStatus(code uint8, msg string) {
	if code < 0 || code >= uint8(len(ExporterStatus)) {
		panic("invalid status code " + strconv.Itoa(int(code)))
	}
	me.Status = code
	me.Message = msg
}

// @TODO: implement!
func (me *AbstractExporter) IsMaster() bool {
	return true
}

/*
func (me *AbstractExporter) Export(data *matrix.Matrix) error {
	me.mu.Lock()
	defer me.mu.Unlock()
	return me.ExportData(data)
}

func (me *AbstractExporter) ExportData(data *matrix.Matrix) error {
	panic(me.Class + " did not implement ExportData()")
}

func (me *AbstractExporter) Render(data *matrix.Matrix) ([][]byte, error) {
	panic(me.Class + " did not implement Render()")
}*/
