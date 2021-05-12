/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package config

import (
	"goharvest2/pkg/constant"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/tree"
	"goharvest2/pkg/tree/node"
	"os"
	"path"
)

func LoadConfig(config_fp string) (*node.Node, error) {
	return tree.Import("yaml", config_fp)
}

func SafeConfig(n *node.Node, fp string) error {
	return tree.Export(n, "yaml", fp)
}

func GetExporters(config_fp string) (*node.Node, error) {
	var err error
	var config, exporters *node.Node

	if config, err = LoadConfig(config_fp); err != nil {
		return nil, err
	}

	if exporters = config.GetChildS("Exporters"); exporters == nil {
		err = errors.New(errors.ERR_CONFIG, "[Exporters] section not found")
		return nil, err
	}

	return exporters, nil
}

func GetPollerNames(config_fp string) ([]string, error) {

	var poller_names []string
	var config, pollers *node.Node
	var err error

	if config, err = LoadConfig(config_fp); err != nil {
		return poller_names, err
	}

	if pollers = config.GetChildS("Pollers"); pollers == nil {
		return poller_names, errors.New(errors.ERR_CONFIG, "[Pollers] section not found")
	}

	poller_names = make([]string, 0)

	for _, p := range pollers.GetChildren() {
		poller_names = append(poller_names, p.GetNameS())
	}

	return poller_names, nil
}

func GetPollers(config_fp string) (*node.Node, error) {
	var config, pollers, defaults *node.Node
	var err error

	if config, err = LoadConfig(config_fp); err != nil {
		return nil, err
	}

	pollers = config.GetChildS("Pollers")
	defaults = config.GetChildS("Defaults")

	if pollers == nil {
		err = errors.New(errors.ERR_CONFIG, "[Pollers] section not found")
	} else if defaults != nil { // optional
		for _, p := range pollers.GetChildren() {
			p.Union(defaults)
		}
	}
	return pollers, err
}

func GetPoller(config_fp, poller_name string) (*node.Node, error) {
	var err error
	var pollers, poller *node.Node

	if pollers, err = GetPollers(config_fp); err == nil {
		if poller = pollers.GetChildS(poller_name); poller == nil {
			err = errors.New(errors.ERR_CONFIG, "poller ["+poller_name+"] not found")
		}
	}

	return poller, err
}

/*GetDefaultHarvestConfigPath*/
//This method is used to return the default absolute path of harvest config file.
func GetDefaultHarvestConfigPath() (string, error) {
	var configPath string
	var err error
	configFileName := constant.ConfigFileName
	if configPath = os.Getenv("HARVEST_CONF"); configPath == "" {
		var homePath string
		homePath = GetHarvestHomePath()
		configPath = path.Join(homePath, configFileName)
	} else {
		configPath = path.Join(configPath, configFileName)
	}
	return configPath, err
}

/*GetHarvestHomePath*/
//This method is used to return current working directory
func GetHarvestHomePath() string {
	return "./"
}

/*
This method returns port configured in prometheus exporter for given poller
If there are more than 1 exporter configured for a poller then return string will have ports as comma seperated
*/
func GetPrometheusExporterPorts(p *node.Node, configFp string) (string, error) {
	var port string
	exporters := p.GetChildS("exporters")
	if exporters != nil {
		exportChildren := exporters.GetAllChildContentS()
		definedExporters, err := GetExporters(configFp)
		if err != nil {
			return "", err
		}
		for _, ec := range exportChildren {
			exporterType := definedExporters.GetChildS(ec).GetChildContentS("exporter")
			if exporterType == "Prometheus" {
				currentPort := definedExporters.GetChildS(ec).GetChildContentS("port")
				port = currentPort
			}
		}
	}
	return port, nil
}

// Returns unique type of exporters for the poller
// For example: If 2 prometheus exporters are configured for a poller then last one defined is returned
func GetUniqueExporters(p *node.Node, configFp string) ([]string, error) {
	var resultExporters []string
	exporters := p.GetChildS("exporters")
	if exporters != nil {
		exportChildren := exporters.GetAllChildContentS()
		definedExporters, err := GetExporters(configFp)
		if err != nil {
			return nil, err
		}
		exporterMap := make(map[string]string)
		for _, ec := range exportChildren {
			exporterType := definedExporters.GetChildS(ec).GetChildContentS("exporter")
			exporterMap[exporterType] = ec
		}

		for _, value := range exporterMap {
			resultExporters = append(resultExporters, value)
		}
	}
	return resultExporters, nil
}
