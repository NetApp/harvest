package config

import (
	"path"
	"goharvest2/share/tree"
	"goharvest2/share/tree/node"
	"goharvest2/share/errors"
)

func LoadConfig(harvest_path, config_fn string) (*node.Node, error) {
	return tree.ImportYaml(path.Join(harvest_path, config_fn))
}

func GetExporters(harvest_path, config_fn string) (*node.Node, error) {
	var err error
	var config, exporters *node.Node

	if config, err = LoadConfig(harvest_path, config_fn); err != nil {
		return nil, err
	}

	if exporters = config.GetChildS("Exporters"); exporters == nil {
		err = errors.New(errors.ERR_CONFIG, "[Exporters] section not found")	
		return nil, err
	}

	return exporters, nil
}


func GetPoller(harvest_path, config_fn, poller_name string) (*node.Node, error) {
	var err error
	var config, pollers, poller, defaults *node.Node

	if config, err = LoadConfig(harvest_path, config_fn); err != nil {
		return nil, err
	}

	pollers = config.GetChildS("Pollers")
	defaults = config.GetChildS("Defaults")

	if pollers != nil {
		if poller = pollers.GetChildS(poller_name); poller != nil {
			if defaults != nil { // optional
				poller.Union(defaults)
			}
		} else {
			err = errors.New(errors.ERR_CONFIG, "poller [" + poller_name + "] not found")
		}
	} else {
		err = errors.New(errors.ERR_CONFIG, "[Pollers] section not found")	
	}

	return poller, err
}