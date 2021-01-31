package collector

import (
	"sync"
	"strings"
	"errors"
	"poller/yaml"
	"poller/collector/abc"
	"poller/collector/zapi"
	"poller/collector/psutil"
	"poller/structs/opts"
	"poller/exporter"
)

type Collector interface {
	Init() error
	Start(*sync.WaitGroup)
	GetName() string
	GetObject() string
	IsUp() bool
	WantedExporters() []string
	LinkExporter(exporter.Exporter)
}

func Load(class, object string, options *opts.Opts, params *yaml.Node) ([]Collector, error) {
	var collectors []Collector
	var err error

	template, err := abc.ImportTemplate(options.Path, class)
	if err != nil {
		return nil, err
	} else if template != nil {
		template.Union(params, false)
		// log: imported and merged template...
	}

	if object == "" {
		object = template.GetChildValue("object")
	}

	if object != "" {
		if c, err := New(class, object, options, template.Copy()); err == nil {
			if err = c.Init(); err == nil {
				collectors = append(collectors, c)
			} else {
				return collectors, err
			}
		} else {
			return collectors, err
		}
	} else if objects := template.GetChild("objects"); objects != nil {
		for _, object := range objects.GetChildren() {
			if c, err := New(class, object.Name, options, params.Copy()); err == nil {
				if err = c.Init(); err == nil {
					collectors = append(collectors, c)
				} else {
					return collectors, err
				}	
			} else {
				return collectors, err
			}
		}
	} else {
		return collectors, errors.New("no object defined in template")
	}

	return collectors, err
}

func New(class, object string, options *opts.Opts, params *yaml.Node) (Collector, error) {

	switch strings.ToLower(class) {
	case "zapi":
		return zapi.New(class, object, options, params), nil
	case "psutil":
		return psutil.New(class, object, options, params), nil
	default:
		return nil, errors.New("unknown collector class: " + class)
	}
}
