//go:build regression

package main

import (
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"goharvest2/pkg/conf"
	"log"
	"strconv"
	"strings"
	"testing"
)

func TestPollerMetrics(t *testing.T) {
	conf.LoadHarvestConfig(installer.HarvestConfigFile)
	for _, pollerName := range conf.Config.PollersOrdered {
		port, _ := conf.GetPrometheusExporterPorts(pollerName)
		portString := strconv.Itoa(port)
		var validCounters = 0
		sb, error := utils.GetResponse("http://localhost:" + strings.TrimSpace(portString) + "/metrics")
		if error != nil {
			panic("Unable to get metric data")
		}
		rows := strings.Split(sb, "\n")
		for i := range rows {
			row := rows[i]
			length := len(row)
			if length == 0 {
				continue
			}
			open := strings.Index(row, "{")
			close := strings.Index(row, "}")
			space := strings.Index(row, " ")
			if open > 0 && close > 0 && space > 0 {
				//objectName := row[0:open]
				//metricContent := row[open:(close+1)]
				metricValue, _ := strconv.Atoi(strings.TrimSpace(row[(close + 1):length]))
				if metricValue > 0 {
					validCounters++
				}
			} else {
				log.Println("invalid string data found in the metric output " + row)
			}
		}
		if validCounters == 0 {
			panic("Empty values found for all counters for poller " + pollerName)
		}
		log.Printf("Total number of counters verified %d for poller '%s' \n", validCounters, pollerName)
	}

}
