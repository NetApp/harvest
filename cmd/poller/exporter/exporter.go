/*
	Copyright NetApp Inc, 2021 All rights reserved

	Package exporter provides the Exporter interface and
	type AbstractExporter that implements most basic
	functions.
*/
package exporter

import (
	"goharvest2/cmd/poller/options"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/logging"
	"goharvest2/pkg/matrix"
	"strconv"
	"sync"
)

// Exporter defines the required attributes of an exporter
// All except, Export() are implemented by AbstractExporter
type Exporter interface {
	Init() error      // initialize exporter
	GetClass() string // the class of the exporter, e.g. Prometheus, InfluxDB
	GetName() string  // the name of the exporter instance
	// Name is different from Class, since we can have multiple instances of the same Class
	GetExportCount() uint64             // return and reset number of exported data points, used by Poller to keep stats
	AddExportCount(uint64)              // add count to the export count, called by the exporter itself
	GetStatus() (uint8, string, string) // return current state of the exporter
	Export(*matrix.Matrix) error        // render data in matrix to the desired format and emit
	// this is the only function that should be implemented by "real" exporters
}

// ExporterStatus defines the possible states of an exporter
var ExporterStatus = [3]string{
	"up",
	"standby",
	"failed",
}

// AbstractExporter implements all methods of the Exporter interface, except Export()
// It defines attributes that will be "inherited" by child exporters
type AbstractExporter struct {
	Class       string
	Name        string
	Logger      *logging.Logger // logger used for logging
	Status      uint8
	Message     string
	Options     *options.Options
	Params      conf.Exporter
	Metadata    *matrix.Matrix // metadata about the export
	*sync.Mutex                // mutex to block exporter during export
	exportCount uint64         // atomic
	countMux    *sync.Mutex
}

// New creates an AbstractExporter instance with the given arguments:
// @c - exporter class
// @n - exporter name
// @o - poller options
// @p - exporter parameters
func New(c, n string, o *options.Options, p conf.Exporter) *AbstractExporter {
	abc := AbstractExporter{
		Class:    c,
		Name:     n,
		Options:  o,
		Params:   p,
		Logger:   logging.SubLogger("exporter", n),
		Mutex:    &sync.Mutex{},
		countMux: &sync.Mutex{},
	}
	return &abc
}

// InitAbc() initializes AbstractExporter
func (me *AbstractExporter) InitAbc() error {
	me.Metadata = matrix.New(me.Name, "metadata_exporter", "metadata_exporter")
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

// GetClass returns the class of the AbstractExporter
func (me *AbstractExporter) GetClass() string {
	return me.Class
}

// GetName returns the name of the AbstractExporter
func (me *AbstractExporter) GetName() string {
	return me.Name
}

// GetExportCount reports and resets count of exported data points "atomically"
// this and next methods are only to report the poller
// how much data we have exported (independent of poll/export interval)
func (me *AbstractExporter) GetExportCount() uint64 {
	me.countMux.Lock()
	count := me.exportCount
	me.exportCount = 0
	me.countMux.Unlock()
	return count
}

// AddExportCount adds count n to the export counter
func (me *AbstractExporter) AddExportCount(n uint64) {
	me.countMux.Lock()
	me.exportCount += n
	me.countMux.Unlock()
}

// GetStatus returns current state of exporter
func (me *AbstractExporter) GetStatus() (uint8, string, string) {
	return me.Status, ExporterStatus[me.Status], me.Message
}

// SetStatus sets the current state of exporter
func (me *AbstractExporter) SetStatus(code uint8, msg string) {
	if code < 0 || code >= uint8(len(ExporterStatus)) {
		panic("invalid status code " + strconv.Itoa(int(code)))
	}
	me.Status = code
	me.Message = msg
}

// @TODO: implement!
/*	This method/attribute is intended to tell Poller/collectors
	wheither or not they should metadata to this exporter.
	Currently not implemented
func (me *AbstractExporter) IsMaster() bool {
	return true
}
*/
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
