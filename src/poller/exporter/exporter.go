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
    GetName() string
	Export(*matrix.Matrix) error
}

func New(class string, params *yaml.Node, options *opts.Opts) Exporter {
	return prometheus.New(class, params, options)
}
