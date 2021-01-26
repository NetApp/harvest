package collector

import (
	"sync"
	"strings"
	"poller/yaml"
	"poller/collector/zapi"
	"poller/collector/procfs"
	"poller/structs/opts"
	"poller/exporter"
)

type Collector interface {
	//New(string, *yaml.Node, *structs.Options) Collector
	Init() error
	Start(*sync.WaitGroup)
	//Poll() error
	GetClass() string
	GetName() string
	GetExporterNames() []string
	AddExporter(exporter.Exporter)
}

func New(class string, params *yaml.Node, options *opts.Opts) ([]Collector) {
	var collectors []Collector

	switch strings.ToLower(class) {
		case "zapi":
			instances := zapi.New(class, params, options)
			for _, c := range instances {
				collectors = append(collectors, c)
			}
			break
		case "procfs":
			instances := procfs.New(class, params, options)
			for _, c := range instances {
				collectors = append(collectors, c)
			}
			break
	}

	return collectors
}
