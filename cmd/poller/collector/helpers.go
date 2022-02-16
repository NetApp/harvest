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
	"fmt"
	"github.com/hashicorp/go-version"
	"io/ioutil"
	"path"
	"regexp"
	"sort"
	"strings"

	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/poller/plugin/aggregator"
	"goharvest2/cmd/poller/plugin/label_agent"
	"goharvest2/pkg/tree"
	"goharvest2/pkg/tree/node"
)

// ImportTemplate retrieves the config (template) of a collector, arguments are:
// @confDir			- path of Harvest config directory (usually /etc/harvest)
// @confFn			- template filename (e.g. default.yaml or custom.yaml)
// @collectorName	- name of the collector
func ImportTemplate(confPath, confFn, collectorName string) (*node.Node, error) {
	fp := path.Join(confPath, "conf/", strings.ToLower(collectorName), confFn)
	return tree.ImportYaml(fp)
}

// ImportSubTemplate retrieves the best matching subtemplate of a collector object.
//
// This method is only applicable to the Zapi/ZapiPerf collectors which have
// multiple objects and each object is forked as a separate collector.
// The subtemplates are sorted in subdirectories that serve as "tag" for the
// matching ONTAP version. ImportSubTemplate will attempt to choose the subtemplate
// with the closest matching ONTAP version.
//
// Arguments:
// @model		- ONTAP model, either cdot or 7mode
// @filename	- name of the subtemplate
// @version		- ONTAP version triple (generation, major, minor)
//
func (c *AbstractCollector) ImportSubTemplate(model, filename string, ver [3]int) (*node.Node, error) {

	var (
		selectedVersion, pathPrefix, subTemplateFp string
		availableVersions                          []string
		err                                        error
		finalTemplate                              *node.Node
		tempTemplate                               *node.Node
	)

	//split filename by comma
	// in case of custom.yaml having same key, file names will be concatenated by comma
	filenames := strings.Split(filename, ",")

	for _, f := range filenames {
		pathPrefix = path.Join(c.Options.HomePath, "conf/", strings.ToLower(c.Name), model)
		c.Logger.Debug().Msgf("Looking for best-fitting template in [%s]", pathPrefix)
		verWithDots := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ver)), "."), "[]")

		// check for available versions, those are the subdirectories that include filename
		if files, err := ioutil.ReadDir(pathPrefix); err == nil {
			for _, file := range files {
				if match, _ := regexp.MatchString(`\d+\.\d+\.\d+`, file.Name()); match == true && file.IsDir() {
					if templates, err := ioutil.ReadDir(path.Join(pathPrefix, file.Name())); err == nil {
						for _, t := range templates {
							if t.Name() == f {
								c.Logger.Trace().Msgf("available version dir: [%s]", file.Name())
								availableVersions = append(availableVersions, file.Name())
								break
							}
						}
					}
				}
			}
		} else {
			return nil, err
		}
		c.Logger.Trace().Msgf("checking for %d available versions: %v", len(availableVersions), availableVersions)

		if len(availableVersions) > 0 {
			versions := make([]*version.Version, len(availableVersions))
			for i, raw := range availableVersions {
				v, err := version.NewVersion(raw)
				if err != nil {
					c.Logger.Trace().Msgf("error parsing version: %s err: %s", raw, err)
					continue
				}
				versions[i] = v
			}

			sort.Sort(version.Collection(versions))

			verS, err := version.NewVersion(verWithDots)
			if err != nil {
				c.Logger.Trace().Msgf("error parsing ONTAP version: %s err: %s", verWithDots, err)
				return nil, errors.New("no best-fitting subtemplate version found")
			}
			// get closest index
			idx := getClosestIndex(versions, verS)
			if idx >= 0 && idx < len(versions) {
				selectedVersion = versions[idx].String()
			}
		}

		if selectedVersion == "" {
			return nil, errors.New("no best-fit template found")
		}

		subTemplateFp = path.Join(pathPrefix, selectedVersion, f)
		c.Logger.Info().Msgf("best-fit template [%s] for [%s]", subTemplateFp, verWithDots)
		if finalTemplate == nil {
			finalTemplate, err = tree.ImportYaml(subTemplateFp)
			if err == nil {
				finalTemplate.PreprocessTemplate()
			}
		} else {
			tempTemplate, err = tree.ImportYaml(subTemplateFp)
			if tempTemplate == nil || err != nil {
				c.Logger.Warn().Err(err).Str("template", subTemplateFp).
					Msg("Unable to import template file. File is invalid or empty")
				continue
			}
			if err == nil {
				tempTemplate.PreprocessTemplate()
				// merge templates
				finalTemplate.Merge(tempTemplate, nil)
			}
		}
	}
	return finalTemplate, err
}

//getClosestIndex returns the closest left match to the sorted list of input versions
// returns -1 when the versions list is empty
// returns equal or closest match to the left
func getClosestIndex(versions []*version.Version, version *version.Version) int {
	if len(versions) == 0 {
		return -1
	}
	idx := sort.Search(len(versions), func(i int) bool {
		return versions[i].GreaterThanOrEqual(version)
	})

	// if we are at length of slice
	if idx == len(versions) {
		return len(versions) - 1
	}

	// if idx is greater than 0 but less than length of slice
	if idx > 0 && idx < len(versions) && !versions[idx].Equal(version) {
		return idx - 1
	}
	return idx
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
