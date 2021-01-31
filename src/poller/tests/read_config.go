package main

import (
	"path"
	"poller/yaml"
)

func ReadConfig(harvest_path, config_fn, name string) (*yaml.Node, *yaml.Node, error) {
	var err error
	var config, pollers, p, exporters, defaults *yaml.Node

	config, err = yaml.Import(path.Join(harvest_path, config_fn))

	if err == nil {

		pollers = config.GetChild("Pollers")
		defaults = config.GetChild("Defaults")

		if pollers == nil {
			err = errors.New("No pollers defined")
		} else {
			p = pollers.GetChild(name)
			if p == nil {
				err = errors.New("Poller [" + name + "] not defined")
			} else if defaults != nil {
				p.Union(defaults, false)
			}
		}
	}

	if err == nil && p != nil {

		exporters = config.GetChild("Exporters")
		if exporters == nil {
			Log.Warn("No exporters defined in config [%s]", config)
		} else {
			requested := p.GetChild("exporters")
			redundant := make([]*yaml.Node, 0)
			if requested != nil {
				for _, e := range exporters.Children {
					if !requested.HasInValues(e.Name) {
						redundant = append(redundant, e)
					}
				}
				for _, e := range redundant {
					exporters.PopChild(e.Name)
				}
			}
		}
	}

	return p, exporters, err
}