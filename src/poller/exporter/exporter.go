package exporter

import (
	"poller/exporter/prometheus"
	"poller/yaml"
	"poller/structs/matrix"
	"poller/structs/opts"
)

type Exporter interface {
	//New(string, *yaml.Node, *structs.Options) Collector
	Init() error
	GetClass() string
	GetName() string
	IsUp() bool
	Export(*matrix.Matrix) error
}

func New(class, name string, options *opts.Opts, params *yaml.Node) Exporter {
	return prometheus.New(class, name, options, params)
}
