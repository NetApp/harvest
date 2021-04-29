/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package exporter

import (
	"goharvest2/cmd/poller/options"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strconv"
	"sync"
)

type Exporter interface {
	Init() error
	GetClass() string
	GetName() string
	GetExportCount() uint64
	AddExportCount(uint64)
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
	Options  *options.Options
	Params   *node.Node
	Metadata *matrix.Matrix
	*sync.Mutex
	exportCount uint64
	countMux    *sync.Mutex
}

func New(c, n string, o *options.Options, p *node.Node) *AbstractExporter {
	abc := AbstractExporter{
		Name:     n,
		Class:    c,
		Options:  o,
		Params:   p,
		Prefix:   "(exporter) (" + n + ")",
		Mutex:    &sync.Mutex{},
		countMux: &sync.Mutex{},
	}
	return &abc
}

func (me *AbstractExporter) InitAbc() error {
	me.Metadata = matrix.New(me.Name, "metadata_exporter")
	me.Metadata.SetGlobalLabel("hostname", me.Options.Hostname)
	me.Metadata.SetGlobalLabel("version", me.Options.Version)
	me.Metadata.SetGlobalLabel("poller", me.Options.Poller)
	me.Metadata.SetGlobalLabel("exporter", me.Class)
	me.Metadata.SetGlobalLabel("target", me.Name)

	if _, err := me.Metadata.NewMetricInt64("time"); err != nil {
		return err
	}
	if _, err := me.Metadata.NewMetricUint64("count"); err != nil {
		return err
	}

	//e.Metadata.AddLabel("task", "")
	if instance, err := me.Metadata.NewInstance("export"); err == nil {
		instance.SetLabel("task", "export")
	} else {
		return err
	}

	if instance, err := me.Metadata.NewInstance("render"); err == nil {
		instance.SetLabel("task", "render")
	} else {
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

// get count of exported data points and reset counter
// this and next methods are only to report the poller
// how much data we have exported (independent of poll/export interval)
func (me *AbstractExporter) GetExportCount() uint64 {
	me.countMux.Lock()
	count := me.exportCount
	me.exportCount = 0
	me.countMux.Unlock()
	return count
}

// add count to the export counter
func (me *AbstractExporter) AddExportCount(n uint64) {
	me.countMux.Lock()
	me.exportCount += n
	me.countMux.Unlock()
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
