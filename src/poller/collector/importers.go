package collector

import (
	"errors"
	"goharvest2/share/logger"
	"goharvest2/share/tree"
	"goharvest2/share/tree/node"
	"io/ioutil"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func ImportTemplate(conf_path, collector_name string) (*node.Node, error) {
	fp := path.Join(conf_path, "conf/", strings.ToLower(collector_name), "default.yaml")
	return tree.Import("yaml", fp)
}

func (c *AbstractCollector) ImportSubTemplate(model, dirname, filename string, version [3]int) (*node.Node, error) {

	var err error
	var selected_version string
	var template *node.Node

	path_prefix := path.Join(c.Options.ConfPath, "conf/", strings.ToLower(c.Name), dirname, model)
	logger.Debug(c.Prefix, "Looking for best-fitting template in [%s]", path_prefix)

	available := make(map[string]bool)
	files, _ := ioutil.ReadDir(path_prefix)
	for _, file := range files {
		logger.Trace(c.Prefix, "Found version dir: [%s]", file.Name())
		if match, _ := regexp.MatchString(`\d+\.\d+\.\d+`, file.Name()); match == true && file.IsDir() {
			available[file.Name()] = true
		}
	}

	vers := version[0]*100 + version[1]*10 + version[2]

	for max := 300; max > 0 && vers > 0; max -= 1 {
		str := strings.Join(strings.Split(strconv.Itoa(vers), ""), ".")
		if _, exists := available[str]; exists == true {
			selected_version = str
			break
		}
		vers -= 1
	}

	if selected_version == "" {
		logger.Debug(c.Prefix, "looking for newer version")

		vers = version[0]*100 + version[1]*10 + version[2]

		for max := 300; max > 0 && vers > 0; max -= 1 {
			str := strings.Join(strings.Split(strconv.Itoa(vers), ""), ".")
			if _, exists := available[str]; exists == true {
				selected_version = str
				break
			}
			vers += 1
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
