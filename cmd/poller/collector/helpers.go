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
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/aggregator"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/labelagent"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/max"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/metricagent"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

// ImportTemplate retrieves the config (template) of a collector, arguments are:
// @confDir			- path of Harvest config directory (usually /etc/harvest)
// @confFn			- template filename (e.g. default.yaml or custom.yaml)
// @collectorName	- name of the collector
func ImportTemplate(confPath, confFn, collectorName string) (*node.Node, error) {
	fp := path.Join(confPath, "conf/", strings.ToLower(collectorName), confFn)
	return tree.ImportYaml(fp)
}

var versionRegex = regexp.MustCompile(`\d+\.\d+\.\d+`)

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
func (c *AbstractCollector) ImportSubTemplate(model, filename string, ver [3]int) (*node.Node, string, error) {

	var (
		selectedVersion, pathPrefix, templatePath string
		availableVersions                         []string
		err                                       error
		customTemplateErr                         error
		finalTemplate                             *node.Node
		customTemplate                            *node.Node
	)

	//split filename by comma
	// in case of custom.yaml having same key, file names will be concatenated by comma
	filenames := strings.Split(filename, ",")

	for _, f := range filenames {
		pathPrefix = c.GetTemplatePathPrefix(model)
		c.Logger.Debug().Msgf("Looking for best-fitting template in [%s]", pathPrefix)
		verWithDots := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ver)), "."), "[]")

		// check for available versions, those are the subdirectories that include filename
		if files, err := os.ReadDir(pathPrefix); err == nil {
			for _, file := range files {
				if !file.IsDir() {
					continue
				}
				submatch := versionRegex.FindStringSubmatch(file.Name())
				if len(submatch) > 0 {
					if templates, err := os.ReadDir(path.Join(pathPrefix, file.Name())); err == nil {
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
			return nil, "", err
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
				return nil, "", errors.New("no best-fitting subtemplate version found")
			}
			// get closest index
			idx := getClosestIndex(versions, verS)
			if idx >= 0 && idx < len(versions) {
				selectedVersion = versions[idx].String()
			}
		}

		if selectedVersion == "" {
			// workaround for 7mode template that will always be missing in cdot
			if filename == "status_7.yaml" && model == "cdot" {
				return nil, "", errs.New(errs.ErrWrongTemplate, "unable to load status_7.yaml on cdot")
			}
			return nil, "", errors.New("no best-fit template found")
		}

		templatePath = path.Join(pathPrefix, selectedVersion, f)
		c.Logger.Info().
			Str("path", templatePath).
			Str("v", verWithDots).
			Msg("best-fit template")
		if finalTemplate == nil {
			finalTemplate, err = tree.ImportYaml(templatePath)
			if err == nil {
				finalTemplate.PreprocessTemplate()
			}
		} else {
			// any errors w.r.t customTemplate are warnings and should not be returned to caller
			customTemplate, customTemplateErr = tree.ImportYaml(templatePath)
			if customTemplate == nil || customTemplateErr != nil {
				c.Logger.Warn().Err(err).Str("template", templatePath).
					Msg("Unable to import template file. File is invalid or empty")
				continue
			}
			if customTemplateErr == nil {
				customTemplate.PreprocessTemplate()
				// merge templates
				finalTemplate.Merge(customTemplate, nil)
			}
		}
	}
	return finalTemplate, templatePath, err
}

func (c *AbstractCollector) GetTemplatePathPrefix(model string) string {
	return path.Join(c.Options.HomePath, "conf/", strings.ToLower(c.Name), model)
}

// getClosestIndex returns the closest left match to the sorted list of input versions
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

// GetBuiltinPlugin returns built-in plugin with name if it exists, otherwise nil
func GetBuiltinPlugin(name string, abc *plugin.AbstractPlugin) plugin.Plugin {

	if name == "Aggregator" {
		return aggregator.New(abc)
	}

	if name == "Max" {
		return max.New(abc)
	}
	/* this will be added in soon
	if name == "Calculator" {--4
		return calculator.New(abc)
	}
	*/

	if name == "LabelAgent" {
		return labelagent.New(abc)
	}

	if name == "MetricAgent" {
		return metricagent.New(abc)
	}

	return nil
}
