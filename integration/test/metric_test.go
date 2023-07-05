package main

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/rs/zerolog/log"
	"sort"
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
	var duplicateMetrics []string
	for _, pollerName := range conf.Config.PollersOrdered {
		port, _ := conf.GetPrometheusExporterPorts(pollerName, true)
		portString := strconv.Itoa(port)
		var validCounters = 0
		uniqueSetOfMetricLabels := make(map[string]bool)
		sb, err2 := utils.GetResponse("http://localhost:" + strings.TrimSpace(portString) + "/metrics")
		if err2 != nil {
			t.Fatalf("Unable to get metric data for %s", pollerName)
		}
		rows := strings.Split(sb, "\n")
		for i := range rows {
			row := rows[i]
			if len(row) == 0 {
				continue
			}
			// Ignore comments
			if strings.HasPrefix(row, "#") {
				continue
			}
			openBracket := strings.Index(row, "{")
			firstSpace := strings.Index(row, " ")
			if openBracket == -1 {
				// this means the metric has this form
				// metric_without_labels 12.47
				if firstSpace == -1 {
					continue
				}
				key := metricAndLabelKey(row[:firstSpace], "")
				if uniqueSetOfMetricLabels[key] {
					duplicateMetrics = append(duplicateMetrics,
						fmt.Sprintf("Duplicate metric poller=%s, got >1 want 1 of %s", pollerName, key))
				}
				uniqueSetOfMetricLabels[key] = true
				continue
			}
			closeBracket := strings.Index(row, "}")

			if openBracket > 0 && closeBracket > 0 && firstSpace > 0 {
				// Turn metric and labels into a unique key
				key := metricAndLabelKey(row[:openBracket], row[openBracket+1:])
				if uniqueSetOfMetricLabels[key] {
					duplicateMetrics = append(duplicateMetrics,
						fmt.Sprintf("Duplicate metric poller=%s, got >1 want 1 of %s", pollerName, key))
				}
				uniqueSetOfMetricLabels[key] = true
				metricValue, _ := strconv.Atoi(strings.TrimSpace(row[(closeBracket + 1):]))
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
	sort.Strings(duplicateMetrics)
	for _, dupMetric := range duplicateMetrics {
		if shouldSkipMetric(dupMetric) {
			log.Info().Str("metric", dupMetric).Msg("Skip")
			continue
		}
		t.Errorf(dupMetric)
	}
}

func metricAndLabelKey(metric string, rest string) string {
	var (
		scanner int
		labels  []string
	)

	for {
		label, equalIndex := readLabel(rest, scanner)
		if equalIndex == 0 {
			break
		}
		if equalIndex != 0 {
			equalIndex++
			if string(rest[equalIndex]) == `"` {
				// Scan until you hit another unescaped quote.
				// Can be any sequence of UTF-8 characters, but the backslash (\),
				// and double-quote (") characters have to be
				// escaped as \, and \"
				labelEnd := 0
				for i := equalIndex + 1; i < len(rest); i++ {
					s := string(rest[i])
					if s == `\` {
						i++
						continue
					}
					if s == `"` {
						// done reading quoted
						labelEnd = i
						break
					}
				}
				labelValue := rest[equalIndex+1 : labelEnd]
				labels = append(labels, fmt.Sprintf(`%s="%s"`, label, labelValue))
				scanner = labelEnd + 1
				if string(rest[scanner]) == "," {
					scanner++
				}
			}
		}
	}

	sort.Strings(labels)

	return metric + "{" + strings.Join(labels, ",") + "}"
}

func readLabel(sample string, i int) (string, int) {
	sub := sample[i:]
	equalsIndex := strings.Index(sub, "=")
	if equalsIndex == -1 {
		return sample, 0
	}
	end := i + equalsIndex
	return sample[i:end], end
}

func shouldSkipMetric(dupMetric string) bool {
	// Skip metrics that are belongs to aggr_efficiency template as Rest collector don't have separate template
	skip := []string{"logical_used_wo_snapshots", "logical_used_wo_snapshots_flexclones", "physical_used_wo_snapshots", "physical_used_wo_snapshots_flexclones", "total_logical_used", "total_physical_used"}
	for _, s := range skip {
		if strings.Contains(dupMetric, s) {
			return true
		}
	}
	return false
}
