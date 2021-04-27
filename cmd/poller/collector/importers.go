//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package collector

import (
	"errors"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/tree"
	"goharvest2/pkg/tree/node"
	"io/ioutil"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func ImportTemplate(conf_path, collector_name, template string) (*node.Node, error) {
	fp := path.Join(conf_path, "conf/", strings.ToLower(collector_name), template)
	return tree.Import("yaml", fp)
}

func (c *AbstractCollector) ImportSubTemplate(model, dirname, filename string, version [3]int) (*node.Node, error) {

	var err error
	var selected_version string
	var template *node.Node

	path_prefix := path.Join(c.Options.ConfPath, "conf/", strings.ToLower(c.Name), model)
	logger.Debug(c.Prefix, "Looking for best-fitting template in [%s]", path_prefix)

	available := make(map[string]bool)
	files, _ := ioutil.ReadDir(path_prefix)
	for _, file := range files {
		if match, _ := regexp.MatchString(`\d+\.\d+\.\d+`, file.Name()); match == true && file.IsDir() {
			if templates, err := ioutil.ReadDir(path.Join(path_prefix, file.Name())); err == nil {
				for _, t := range templates {
					if t.Name() == filename {
						logger.Trace(c.Prefix, "available version dir: [%s]", file.Name())
						available[file.Name()] = true
						break
					}
				}
			}
		}
	}
	logger.Trace(c.Prefix, "checking for %d available versions: %v", len(available), available)

	vers := version[0]*100 + version[1]*10 + version[2]

	//@TODO: this will become a bug once ontap version is higher than 9.9.0!
	for max := 0; max <= 100; max++ {
		// check older version
		str := strings.Join(strings.Split(strconv.Itoa(vers-max), ""), ".")
		if _, exists := available[str]; exists == true {
			selected_version = str
			break
		}
		// check newer version
		str = strings.Join(strings.Split(strconv.Itoa(vers+max), ""), ".")
		if _, exists := available[str]; exists == true {
			selected_version = str
			break
		}
	}

	if selected_version == "" {
		err = errors.New("No best-fitting subtemplate version found")
	} else {
		template_path := path.Join(path_prefix, selected_version, filename)
		logger.Debug(c.Prefix, "selected best-fitting subtemplate [%s]", template_path)
		template, err = tree.Import("yaml", template_path)
	}
	return template, err
}
