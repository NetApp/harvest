/*
Copyright NetApp Inc, 2021 All rights reserved

Package exporter provides the Exporter interface and
type AbstractExporter that implements most basic
functions.
*/

package exporter

import (
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"log/slog"
	"strconv"
	"sync"
)

// Exporter defines the required attributes of an exporter
// All except, Export() are implemented by AbstractExporter
type Exporter interface {
	Init() error      // initialize exporter
	GetClass() string // the class of the exporter, e.g. Prometheus, InfluxDB
	// GetName is different from Class, since we can have multiple instances of the same Class
	GetName() string                      // the name of the exporter instance
	GetExportCount() uint64               // return and reset number of exported data points, used by Poller to keep stats
	AddExportCount(uint64)                // add count to the export count, called by the exporter itself
	GetStatus() (uint8, string, string)   // return current state of the exporter
	Export(*matrix.Matrix) (Stats, error) // render data in matrix to the desired format and emit
	// this is the only function that should be implemented by "real" exporters
}

// status defines the possible states of an exporter
var status = [3]string{
	"up",
	"standby",
	"failed",
}

// Stats capture the number of instances and metrics exported
type Stats struct {
	InstancesExported uint64
	MetricsExported   uint64
	RenderedBytes     uint64
}

// AbstractExporter implements all methods of the Exporter interface, except Export()
// It defines attributes that will be "inherited" by child exporters
type AbstractExporter struct {
	Class       string
	Name        string
	Logger      *slog.Logger
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
func New(c, n string, o *options.Options, p conf.Exporter, params *conf.Poller) *AbstractExporter {
	abc := AbstractExporter{
		Class:    c,
		Name:     n,
		Options:  o,
		Params:   p,
		Logger:   slog.Default().With(slog.String("exporter", n)),
		Mutex:    &sync.Mutex{},
		countMux: &sync.Mutex{},
		Metadata: matrix.New(n, "metadata_exporter", "metadata_exporter"),
	}
	if params != nil {
		abc.Metadata.SetGlobalLabel("datacenter", params.Datacenter)
		labels := params.Labels
		if labels != nil {
			for _, labelPtr := range *labels {
				abc.Metadata.SetGlobalLabels(labelPtr)
			}
		}
	}
	return &abc
}

// InitAbc initializes AbstractExporter
func (e *AbstractExporter) InitAbc() error {
	e.Metadata.SetGlobalLabel("hostname", e.Options.Hostname)
	e.Metadata.SetGlobalLabel("version", e.Options.Version)
	e.Metadata.SetGlobalLabel("poller", e.Options.Poller)
	e.Metadata.SetGlobalLabel("exporter", e.Class)
	e.Metadata.SetGlobalLabel("target", e.Name)

	if _, err := e.Metadata.NewMetricInt64("time"); err != nil {
		return err
	}
	if _, err := e.Metadata.NewMetricUint64("count"); err != nil {
		return err
	}

	if instance, err := e.Metadata.NewInstance("export"); err == nil {
		instance.SetLabel("task", "export")
	} else {
		return err
	}

	if instance, err := e.Metadata.NewInstance("render"); err == nil {
		instance.SetLabel("task", "render")
	} else {
		return err
	}

	e.SetStatus(0, "initialized")
	return nil
}

// GetClass returns the class of the AbstractExporter
func (e *AbstractExporter) GetClass() string {
	return e.Class
}

// GetName returns the name of the AbstractExporter
func (e *AbstractExporter) GetName() string {
	return e.Name
}

// GetExportCount reports and resets count of exported data points "atomically"
// this and next methods are only to report the poller
// how much data we have exported (independent of poll/export interval)
func (e *AbstractExporter) GetExportCount() uint64 {
	e.countMux.Lock()
	count := e.exportCount
	e.exportCount = 0
	e.countMux.Unlock()
	return count
}

// AddExportCount adds count n to the export counter
func (e *AbstractExporter) AddExportCount(n uint64) {
	e.countMux.Lock()
	e.exportCount += n
	e.countMux.Unlock()
}

// GetStatus returns current state of exporter
func (e *AbstractExporter) GetStatus() (uint8, string, string) {
	return e.Status, status[e.Status], e.Message
}

// SetStatus sets the current state of exporter
func (e *AbstractExporter) SetStatus(code uint8, msg string) {
	if code >= uint8(len(status)) {
		panic("invalid status code " + strconv.FormatUint(uint64(code), 10))
	}
	e.Status = code
	e.Message = msg
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
