/*
Copyright NetApp Inc, 2021 All rights reserved

This file contains helper functions and methods for Poller,
AbstractCollector and collectors
*/

package collector

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/aggregator"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/changelog"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/labelagent"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/maxplugin"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/metricagent"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/go-version"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var ExcludeTemplates = map[string]map[string]struct{}{
	"ZapiPerf": {
		"workload_detail":        {},
		"workload_detail_volume": {},
	},
	"RestPerf": {
		"api/cluster/counter/tables/qos_detail":        {},
		"api/cluster/counter/tables/qos_detail_volume": {},
	},
}

// ImportTemplate looks for a collector's template by searching confPaths for the first template that exists in
// confPath/collectorName/templateName
func ImportTemplate(confPaths []string, templateName, collectorName string) (*node.Node, error) {
	homePath := conf.Path("")
	for _, confPath := range confPaths {
		fp := filepath.Join(homePath, confPath, strings.ToLower(collectorName), templateName)
		_, err := os.Stat(fp)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		return tree.ImportYaml(fp)
	}
	return nil, errors.New("template not found on confPath")
}

var versionRegex = regexp.MustCompile(`\d+\.\d+\.\d+`)

// ImportSubTemplate retrieves the best matching subtemplate of a collector object.
// This method is applicable to collectors which have multiple objects.
// Each object is forked as a separate collector.
// The sub-templates exist in subdirectories named after ONTAP versions. These directories
// are sorted, and we try to return the subtemplate that most closely matches the ONTAP version.
// Model is cdot or 7mode, filename is the name of the subtemplate, and ver is the
// ONTAP version triple (generation, major, minor)
func (c *AbstractCollector) ImportSubTemplate(model, filename, jitter string, verWithDots string) (*node.Node, string, error) {

	var (
		selectedVersion, templatePath string
		customTemplateErr             error
		finalTemplate                 *node.Node
		customTemplate                *node.Node
		importErrs                    []error
	)

	if filename == "" {
		return nil, "", fmt.Errorf("template name is empty. Make sure the object is defined in your default.yaml, confPath: [%s]", c.Options.ConfPath)
	}

	// Filename will be the name of a template (volume.yaml) or, when merging templates, a comma-separated
	// string like "volume.yaml,custom_volume.yaml"
	filenames := strings.Split(filename, ",")

	if verWithDots == "" {
		verWithDots = "0.0.0"
	}
	ontapVersion, err := version.NewVersion(verWithDots)
	if err != nil {
		return nil, "", fmt.Errorf("no best-fit template found due to err=%w", err)
	}
	homePath := conf.Path("")

nextFile:
	for _, f := range filenames {
		for _, confPath := range c.Options.ConfPaths {
			selectedVersion, err = c.findBestFit(homePath, confPath, f, model, ontapVersion)
			if err != nil || selectedVersion == "" {
				continue
			}

			templatePath = filepath.Join(selectedVersion, f)
			if jitter == "" {
				jitter = "none"
			}

			c.Logger.Info(
				"best-fit template",
				slog.String("path", templatePath),
				slog.String("v", verWithDots),
				slog.String("jitter", jitter),
			)

			if finalTemplate == nil {
				finalTemplate, err = tree.ImportYaml(templatePath)
				if err == nil {
					finalTemplate.PreprocessTemplate()
					continue nextFile
				}
				importErrs = append(importErrs, fmt.Errorf("failed to import template: %s file: %w", templatePath, err))
			} else {
				// any errors w.r.t customTemplate are warnings and should not be returned to caller
				customTemplate, customTemplateErr = tree.ImportYaml(templatePath)
				if customTemplateErr != nil {
					c.Logger.Warn("Unable to import template file. File is invalid or empty",
						slogx.Err(err),
						slog.String("path", templatePath),
					)
					continue
				}
				customTemplate.PreprocessTemplate()
				finalTemplate.Merge(customTemplate, nil)
				continue nextFile
			}
		}

		if finalTemplate == nil {
			// workaround for 7mode template that will always be missing in cdot
			if c.Object == "Status_7mode" && model == "cdot" {
				return nil, "", errs.New(errs.ErrWrongTemplate, "unable to load status_7.yaml on cdot")
			}
			return nil, "", fmt.Errorf("no best-fit template for %s on confPath %s %w",
				filename, c.Options.ConfPath, errors.Join(importErrs...))
		}
	}

	if finalTemplate != nil {
		if queries, exists := ExcludeTemplates[c.Name]; exists {
			templateQuery := finalTemplate.GetChildContentS("query")
			if _, ok := queries[templateQuery]; ok {
				return nil, "", fmt.Errorf("%w: template '%s' does not support query '%s' in template '%s'", errs.ErrTemplateNotSupported, c.Object, templateQuery, filename)
			}
		}
	}

	return finalTemplate, templatePath, err
}

func (c *AbstractCollector) findBestFit(homePath string, confPath string, name string, model string, ontapVersion *version.Version) (string, error) {
	var (
		selectedVersion   string
		availableVersions []string
	)

	pathPrefix := filepath.Join(homePath, confPath, strings.ToLower(c.Name), model)
	c.Logger.Debug("Looking for best-fitting template in pathPrefix", slog.String("pathPrefix", pathPrefix))

	// check for available versions, these are the subdirectories with matching filenames
	files, err := os.ReadDir(pathPrefix)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		if versionRegex.MatchString(file.Name()) {
			if templates, err := os.ReadDir(filepath.Join(pathPrefix, file.Name())); err == nil {
				for _, t := range templates {
					if t.Name() == name {
						availableVersions = append(availableVersions, file.Name())
						break
					}
				}
			}
		}
	}

	if len(availableVersions) == 0 {
		return "", nil
	}

	versions := make([]*version.Version, len(availableVersions))
	for i, raw := range availableVersions {
		v, err := version.NewVersion(raw)
		if err != nil {
			continue
		}
		versions[i] = v
	}
	sort.Sort(version.Collection(versions))

	// get closest index
	idx := getClosestIndex(versions, ontapVersion)
	if idx >= 0 && idx < len(versions) {
		selectedVersion = versions[idx].String()
	}

	return filepath.Join(pathPrefix, selectedVersion), nil
}

// getClosestIndex returns the closest left match to the sorted list of input versions
// returns -1 when the version's list is empty
// returns equal or closest match to the left
func getClosestIndex(versions []*version.Version, aVersion *version.Version) int {
	if len(versions) == 0 {
		return -1
	}
	idx := sort.Search(len(versions), func(i int) bool {
		return versions[i].GreaterThanOrEqual(aVersion)
	})

	// if we are at length of slice
	if idx == len(versions) {
		return len(versions) - 1
	}

	// if idx is greater than 0 but less than length of slice
	if idx > 0 && idx < len(versions) && !versions[idx].Equal(aVersion) {
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
		return maxplugin.New(abc)
	}

	if name == "LabelAgent" {
		return labelagent.New(abc)
	}

	if name == "MetricAgent" {
		return metricagent.New(abc)
	}

	if name == "ChangeLog" {
		return changelog.New(abc)
	}

	return nil
}

func RecordKeepLast(n *node.Node, logger *slog.Logger) int {
	r := n.GetChildS("recorder")
	if r == nil {
		return DefaultRecordsToSave
	}
	if !r.HasChildS("path") {
		return DefaultRecordsToSave
	}

	kl := r.GetChildContentS("keep_last")
	if kl != "" {
		keep, err := strconv.Atoi(kl)
		if err != nil {
			logger.Error(
				"invalid keep_last value. Using default.",
				slog.Int("default", DefaultRecordsToSave),
				slog.String("value", kl),
			)
			return DefaultRecordsToSave
		}
		return keep
	}

	return DefaultRecordsToSave
}
