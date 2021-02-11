package exporter

import (
	//"goharvest2/poller/exporter/prometheus"
	//"goharvest2/poller/yaml"
	"goharvest2/poller/struct/matrix"
	//"goharvest2/poller/structs/options"
)

type Exporter interface {
	//New(string, *yaml.Node, *structs.Options) Collector
	Init() error
	GetClass() string
	GetName() string
	GetStatus() (int, string)
	Export(*matrix.Matrix) error
}
/*
func New(class, name string, options *options.Options, params *yaml.Node) Exporter {
	return prometheus.New(class, name, options, params)
}*/
