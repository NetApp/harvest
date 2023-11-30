package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/installer"
	utils2 "github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/cmd/tools/grafana"
	rest2 "github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/cmd/tools/utils"
	zapi2 "github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestCounterUsage(t *testing.T) {
	utils2.SkipIfMissing(t, utils2.Regression)
	log.Info().Msg("Testing unused counters")
	pollerName := "dc1"
	var (
		restClient *rest2.Client
		zapiClient *zapi2.Client
		poller     *conf.Poller
		objects    = make(map[string]bool)
		err        error
	)

	_, err = conf.LoadHarvestConfig(installer.HarvestConfigFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if poller, _, err = rest2.GetPollerAndAddr(pollerName); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	timeout, _ := time.ParseDuration(rest2.DefaultTimeout)
	credentials := auth.NewCredentials(poller, logging.Get())

	if restClient, err = rest2.New(poller, timeout, credentials); err != nil {
		fmt.Printf("error creating new client %+v\n", err)
		os.Exit(1)
	}
	if err = restClient.Init(2); err != nil {
		fmt.Printf("error init rest client %+v\n", err)
		os.Exit(1)
	}
	if zapiClient, err = zapi2.New(poller, credentials); err != nil {
		fmt.Printf("error creating new client %+v\n", err)
		os.Exit(1)
	}
	if err = zapiClient.Init(2); err != nil {
		fmt.Printf("error init zapi client %+v\n", err)
		os.Exit(1)
	}
	utils.GetObjects("../../conf/restperf/default.yaml", objects)
	utils.GetObjects("../../conf/rest/default.yaml", objects)
	restCounters := utils.ProcessRestCounters(restClient, objects, "../../conf/rest", "../../conf/restperf")
	counterUsage(t, restCounters)
	utils.GetObjects("../../conf/zapiperf/default.yaml", objects)
	utils.GetObjects("../../conf/zapi/default.yaml", objects)
	zapiCounters := utils.ProcessZapiCounters(zapiClient, objects, "../../conf/zapi/cdot", "../../conf/zapiperf/cdot")
	counterUsage(t, zapiCounters)
}

func counterUsage(t *testing.T, counters map[string]utils.Counter) {
	tempMap := make(map[string]utils.Counter)
	// handled few special cases
	for counterName := range counters {
		// Case1: we use most of the histogram counter as histogram_bucket.
		if strings.Contains(counterName, "_hist") {
			updatedCounterName := strings.Join([]string{counterName, "bucket"}, "_")
			tempMap[updatedCounterName] = utils.Counter{Visited: false, Template: counters[counterName].Template}
			delete(counters, counterName)
		}

		// Case2: special cases for NFS, where # has been added, here we need to remove it.
		hashIndex := strings.Index(counterName, "#")
		if hashIndex != -1 {
			specialCounter := strings.Split(counterName, "#")
			tempMap[specialCounter[1]] = utils.Counter{Visited: false, Template: counters[counterName].Template}
			delete(counters, counterName)
		}
	}

	// update the actual map from temp map
	for k, v := range tempMap {
		counters[k] = v
	}

	// Case3: handle lun specially as here histogram is used directly without bucket.
	counters["lun_read_align_histo"] = utils.Counter{Visited: false, Template: counters["lun_read_align_histo_bucket"].Template}
	counters["lun_write_align_histo"] = utils.Counter{Visited: false, Template: counters["lun_write_align_histo_bucket"].Template}
	delete(counters, "lun_read_align_histo_bucket")
	delete(counters, "lun_write_align_histo_bucket")

	grafana.VisitDashboards(
		[]string{"../../grafana/dashboards/cmode"},
		func(path string, data []byte) {
			checkUsage(data, counters)
		})

	// These counters {"headroom_aggr_ewma", "headroom_cpu_ewma"} are used in this way: headroom_aggr_ewma_${Interval}
	specialTemplates := []string{
		"../../conf/restperf/9.12.0/resource_headroom_cpu.yaml",
		"../../conf/restperf/9.12.0/resource_headroom_aggr.yaml",
		"../../conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml",
		"../../conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml",
	}

	countMap := make(map[string][]string)
	for counterName, counter := range counters {
		if !counter.Visited {
			if slices.Contains(specialTemplates, counter.Template) {
				if strings.HasPrefix(counterName, "headroom_aggr_ewma") || strings.HasPrefix(counterName, "headroom_cpu_ewma") {
					continue
				}
				countMap[counter.Template] = append(countMap[counter.Template], counterName)
			} else {
				countMap[counter.Template] = append(countMap[counter.Template], counterName)
			}
		}
	}

	for template, names := range countMap {
		t.Errorf(`These %d counters haven't been used in dashboard from %s template %s`, len(names), template, names)
	}
}

func checkUsage(data []byte, counters map[string]utils.Counter) {
	grafana.VisitAllPanels(data, func(path string, key, value gjson.Result) {
		kind := value.Get("type").String()
		if kind == "row" || kind == "text" {
			return
		}
		targetsSlice := value.Get("targets").Array()
		for _, ts := range targetsSlice {
			updateCounterMap(ts.Get("expr").String(), counters)
		}
	})
}

func updateCounterMap(expr string, counters map[string]utils.Counter) {
	if exprs, isSplit := grafana.CheckSpecialExpr(expr); isSplit {
		for _, exp := range exprs {
			updateCounterMap(exp, counters)
		}
	} else {
		allMatches := grafana.MetricRe.FindAllStringSubmatch(expr, -1)
		for _, match := range allMatches {
			m := match[1]
			if len(m) == 0 {
				continue
			}
			expr = m
		}
		counters[expr] = utils.Counter{Visited: true, Template: counters[expr].Template}
	}
}
