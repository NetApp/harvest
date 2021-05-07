/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package config

import (
	"fmt"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/tree"
	"goharvest2/pkg/tree/node"
	"os"
	"path/filepath"
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

/*
if HARVEST_CONF variable exists then return HARVEST_CONF value
else return parent path of executable. For example : harvest binary is in /opt/harvest/bin. This methid will return /opt/harvest
*/
func GetHarvestConf() string {
	var confPath string
	configFileName := "harvest.yml"
	if confPath = os.Getenv("HARVEST_CONF"); confPath == "" {
		path, _ := os.Executable()
		exPath := filepath.Dir(filepath.Dir(path))
		if _, err := os.Stat(exPath + string(os.PathSeparator) + configFileName); os.IsNotExist(err) {
			fmt.Printf("Config file %s does not exist at %s\n", configFileName, exPath)
			os.Exit(1)
		} else {
			confPath = exPath
		}
	}
	//fmt.Printf("Config file %s read from %s\n", configFileName, confPath)
	return confPath
}

/*
This method returns the parent folder path of executable binary
For example : harvest binary is in /opt/harvest/bin. This methid will return /opt/harvest
*/
func GetHarvestHome() string {
	var homePath string
	if homePath = os.Getenv("HARVEST_HOME"); homePath == "" {
		path, _ := os.Executable()
		exPath := filepath.Dir(filepath.Dir(path))
		homePath = exPath
	}
	//fmt.Printf("Harvest path %s\n", homePath)
	return homePath
}
