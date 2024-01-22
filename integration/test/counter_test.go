package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/hashicorp/go-version"
	"github.com/netapp/harvest/v2/cmd/collectors"
	rest2 "github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type counterData struct {
	api          string
	restCounters []string
	perfCounters []string
}

var replacer = strings.NewReplacer("{", "", "}", "", "^^", "", "^", "")

// TestCounters extracts non-hidden counters from all of the rest and restperf templates and then invokes an HTTP GET for each api path + counters.
// Valid responses are status code = 200. Objects do not need to exist on the cluster, only the api path and counter names are checked.
func TestCounters(t *testing.T) {
	var (
		poller *conf.Poller
		client *rest2.Client
	)

	utils.SkipIfMissing(t, utils.Regression)
	_, err := conf.LoadHarvestConfig(installer.HarvestConfigFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to load harvest config")
	}

	pollerName := "dc1"
	if poller, err = conf.PollerNamed(pollerName); err != nil {
		log.Fatal().Err(err).Str("poller", pollerName).Msgf("")
	}
	if poller.Addr == "" {
		log.Fatal().Str("poller", pollerName).Msg("Address is empty")
	}
	timeout, _ := time.ParseDuration(rest2.DefaultTimeout)

	if client, err = rest2.New(poller, timeout, auth.NewCredentials(poller, logging.Get())); err != nil {
		log.Fatal().Err(err).Str("poller", pollerName).Msg("error creating new client")
	}

	if err = client.Init(5); err != nil {
		log.Fatal().Err(err).Msg("client init failed")
	}

	restCounters := processRestCounters(client)
	if err = invokeRestCall(client, restCounters); err != nil {
		log.Error().Err(err).Msg("rest call failed")
	}

}

func invokeRestCall(client *rest2.Client, counters map[string][]counterData) error {
	for _, countersDetail := range counters {
		for _, counterDetail := range countersDetail {
			href := rest2.NewHrefBuilder().
				APIPath(counterDetail.api).
				Fields(counterDetail.restCounters).
				CounterSchema(counterDetail.perfCounters).
				Build()

			if _, err := collectors.InvokeRestCall(client, href, logging.Get()); err != nil {
				return fmt.Errorf("failed to invoke rest href=%s call: %w", href, err)
			}
		}
	}
	return nil
}

func processRestCounters(client *rest2.Client) map[string][]counterData {
	restPerfCounters := visitRestTemplates("../../conf/restperf", client, func(path string, currentVersion string, client *rest2.Client) map[string][]counterData {
		return processRestConfigCounters(path, currentVersion, "perf")
	})

	restCounters := visitRestTemplates("../../conf/rest", client, func(path string, currentVersion string, client *rest2.Client) map[string][]counterData {
		return processRestConfigCounters(path, currentVersion, "rest")
	})

	for k, v := range restPerfCounters {
		restCounters[k] = v
	}
	return restCounters
}

func visitRestTemplates(dir string, client *rest2.Client, eachTemp func(path string, currentVersion string, client *rest2.Client) map[string][]counterData) map[string][]counterData {
	result := make(map[string][]counterData)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read directory:")
		}
		ext := filepath.Ext(path)
		if ext != ".yaml" {
			return nil
		}
		if strings.HasSuffix(path, "default.yaml") {
			return nil
		}
		r := eachTemp(path, client.Cluster().GetVersion(), client)
		for k, v := range r {
			result[k] = v
		}
		return nil
	})

	if err != nil {
		log.Fatal().Err(err).Msgf("failed to walk directory: %s", dir)
	}

	return result
}

func processRestConfigCounters(path string, currentVersion string, kind string) map[string][]counterData {
	countersData := make(map[string][]counterData)
	templateVersion := filepath.Base(filepath.Dir(path))
	templateV, err := version.NewVersion(templateVersion)
	if err != nil {
		return nil
	}
	currentV, err := version.NewVersion(currentVersion)
	if err != nil {
		return nil
	}
	if templateV.GreaterThan(currentV) {
		return nil
	}

	t, err := tree.ImportYaml(path)
	if t == nil || err != nil {
		fmt.Printf("Unable to import template file %s. File is invalid or empty err=%s\n", path, err)
		return nil
	}

	readCounters(t, path, kind, countersData)
	if templateEndpoints := t.GetChildS("endpoints"); templateEndpoints != nil {
		for _, endpoint := range templateEndpoints.GetChildren() {
			readCounters(endpoint, path, kind, countersData)
		}
	}

	return countersData
}

func readCounters(t *node.Node, path, kind string, countersData map[string][]counterData) {
	var counters []string
	if templateCounters := t.GetChildS("counters"); templateCounters != nil {
		templateQuery := t.GetChildS("query")
		counters = make([]string, 0)
		for _, c := range templateCounters.GetAllChildContentS() {
			if c != "" {
				if strings.Contains(c, "filter") || strings.Contains(c, "hidden_fields") {
					continue
				}
				if strings.HasPrefix(c, "^") && kind == "perf" {
					continue
				}
				name, _, _, _ := util.ParseMetric(c)
				nameArray := strings.Split(name, ".#")
				counters = append(counters, replacer.Replace(nameArray[0]))
			}
		}
		if kind == "rest" {
			countersData[path] = append(countersData[path], counterData{api: templateQuery.GetContentS(), restCounters: counters})
		} else {
			countersData[path] = append(countersData[path], counterData{api: templateQuery.GetContentS(), perfCounters: counters})
		}
	}
}
