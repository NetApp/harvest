/*
	Copyright NetApp Inc, 2021 All rights reserved

	This file contains helper functions and methods for Poller,
	AbstractCollector and collectors

	@TODO: review which functions really belong here
	@TODO: review which methods should actually be functions
		(e.g. ImportSubTemplate is not abstract enough to be a method
		of AbstractCollector)
*/
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

// ImportTemplate retrieves the config (template) of a collector, arguments are:
// @confDir			- path of Harvest config durectory (usually /etc/harvest)
// @confFn			- template filename (e.g. default.yaml or custom.yaml)
// @collectorName	- name of the collector
func ImportTemplate(confPath, confFn, collectorName string) (*node.Node, error) {
	fp := path.Join(confPath, "conf/", strings.ToLower(collectorName), confFn)
	return tree.Import("yaml", fp)
}

// ImportSubTemplate retrieves the best matching subtemplate of a collector object.
//
// This method is only applicable to the Zapi/ZapiPerf collectors which have
// multiple objects and each object is forked as a seperate collector.
// The subtemplates are sorted in subdirectories that serve as "tag" for the
// matching ONTAP version. ImportSubTemplate will attempt to choose the subtemplate
// with closest matching ONTAP version.
//
// Arguments:
// @model		- ONTAP model, either cdot or 7mode
// @filename	- name of the subtemplate
// @version		- ONTAP version triple (generation, major, minor)
//
// BUG: ImportSubTemplate will break if ONTAP version is higher than 9.9.0.
func (c *AbstractCollector) ImportSubTemplate(model, filename string, version [3]int) (*node.Node, error) {

	var (
		selectedVersion, pathPrefix, subTemplateFp string
		availableVersions                          map[string]bool
		template                                   *node.Node
		versionDecimal                             int
		err                                        error
	)

	pathPrefix = path.Join(c.Options.ConfPath, "conf/", strings.ToLower(c.Name), model)
	logger.Debug(c.Prefix, "Looking for best-fitting template in [%s]", pathPrefix)

	// check for available versons, those are the subdirectories that include filename
	availableVersions = make(map[string]bool)
	if files, err := ioutil.ReadDir(pathPrefix); err == nil {
		for _, file := range files {
			if match, _ := regexp.MatchString(`\d+\.\d+\.\d+`, file.Name()); match == true && file.IsDir() {
				if templates, err := ioutil.ReadDir(path.Join(pathPrefix, file.Name())); err == nil {
					for _, t := range templates {
						if t.Name() == filename {
							logger.Trace(c.Prefix, "available version dir: [%s]", file.Name())
							availableVersions[file.Name()] = true
							break
						}
					}
				}
			}
		}
	} else {
		return nil, err
	}
	logger.Trace(c.Prefix, "checking for %d available versions: %v", len(availableVersions), availableVersions)

	versionDecimal = version[0]*100 + version[1]*10 + version[2]

	for max := 0; max <= 100; max++ {
		// check older version
		str := strings.Join(strings.Split(strconv.Itoa(versionDecimal-max), ""), ".")
		if _, exists := availableVersions[str]; exists == true {
			selectedVersion = str
			break
		}
		// check newer version
		str = strings.Join(strings.Split(strconv.Itoa(versionDecimal+max), ""), ".")
		if _, exists := availableVersions[str]; exists == true {
			selectedVersion = str
			break
		}
	}

	if selectedVersion == "" {
		return nil, errors.New("No best-fitting subtemplate version found")
	}

	subTemplateFp = path.Join(pathPrefix, selectedVersion, filename)
	logger.Debug(c.Prefix, "selected best-fitting subtemplate [%s]", subTemplateFp)
	return tree.Import("yaml", subTemplateFp)
}

// ParseMetricName parses display name from the raw name of the metric as defined in (sub)template.
// Users can rename a metric with "=>" (e.g. some_long_metric_name => short).
// Trailing "^" characters are ignored/cleaned as they have special meaning in some collectors.
func ParseMetricName(raw string) (string, string) {

	var name, display string

	name = strings.ReplaceAll(raw, "^", "")

	if x := strings.Split(name, "=>"); len(x) == 2 {
		name = strings.TrimSpace(x[0])
		display = strings.TrimSpace(x[1])
	} else {
		display = strings.ReplaceAll(name, "-", "_")
	}

	return name, display
}

// getBuiltinPlugin returns built-in plugin with name if it exists, otherwise nil
func getBuiltinPlugin(name string, abc *plugin.AbstractPlugin) plugin.Plugin {

	if name == "Aggregator" {
		return aggregator.New(abc)
	}

	/* this will be added in soon
	if name == "Calculator" {
		return calculator.New(abc)
	}
	*/

	if name == "LabelAgent" {
		return label_agent.New(abc)
	}

	return nil
}
