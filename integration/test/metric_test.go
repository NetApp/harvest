package main

import (
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"testing"
)

func TestPollerMetrics(t *testing.T) {
	utils.SkipIfMissing(t, utils.Regression)
	err := conf.LoadHarvestConfig(installer.HarvestConfigFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to load harvest config")
	}
	for _, pollerName := range conf.Config.PollersOrdered {
		port, _ := conf.GetPrometheusExporterPorts(pollerName, true)
		portString := strconv.Itoa(port)
		var validCounters = 0
		sb, err2 := utils.GetResponse("http://localhost:" + strings.TrimSpace(portString) + "/metrics")
		if err2 != nil {
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
			closeBracket := strings.Index(row, "}")
			space := strings.Index(row, " ")
			if open > 0 && closeBracket > 0 && space > 0 {
				//objectName := row[0:open]
				//metricContent := row[open:(close+1)]
				metricValue, _ := strconv.Atoi(strings.TrimSpace(row[(closeBracket + 1):length]))
				if metricValue > 0 {
					validCounters++
				}
			} else {
				log.Error().Str("row", row).Msg("Invalid string data found in the metric output")
			}
		}
		if validCounters == 0 {
			panic("Empty values found for all counters for poller " + pollerName)
		}
		log.Info().Int("numCounters", validCounters).Str("poller", pollerName).Msg("Valid Counters for poller")
	}
}
