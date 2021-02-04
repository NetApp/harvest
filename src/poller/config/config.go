package config

import (
	"errors"
	"path"
	"goharvest2/poller/yaml"

)

func GetConfig(harvest_path, config_fn string) (*yaml.Node, error) {
	return yaml.Import(path.Join(harvest_path, config_fn))
}

func GetPoller(harvest_path, config_fn, poller_name string) (*yaml.Node, error) {
	var err error
	var config, pollers, poller, defaults *yaml.Node

	if config, err = GetConfig(harvest_path, config_fn); err != nil {
		return nil, err
	}

	pollers = config.GetChild("Pollers")
	defaults = config.GetChild("Defaults")

	if pollers != nil {
		if poller = pollers.GetChild(poller_name); poller != nil {
			if defaults != nil { // optional
				poller.Union(defaults, false)
			}
		} else {
			err = errors.New("Poller [" + poller_name + "] not found")
		}
	} else {
		err = errors.New("No [Pollers] section found")	
	}

	return poller, err
}