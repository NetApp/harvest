package main

import (
	"crypto/fips140"
	"fmt"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/request"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Skip aggr_efficiency template metrics since the Rest collector does not have a separate template

var skipDuplicates = map[string]bool{
	"aggr_logical_used_wo_snapshots":             true,
	"aggr_logical_used_wo_snapshots_flexclones":  true,
	"aggr_physical_used_wo_snapshots":            true,
	"aggr_physical_used_wo_snapshots_flexclones": true,
	"aggr_total_logical_used":                    true,
	"aggr_total_physical_used":                   true,
}

func TestPollerMetrics(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.Regression)
	_, err := conf.LoadHarvestConfig(installer.HarvestConfigFile)
	if err != nil {
		slog.Error("Unable to load harvest config", slogx.Err(err))
		os.Exit(1)
	}
	var duplicateMetrics []string
	for _, pollerName := range conf.Config.PollersOrdered {
		if fips140.Enabled() && pollerName == "umeng-aff300-05-06" {
			// FIPS 140-3 is only supported on ONTAP 9.11.1+
			// umeng-aff300-05-06 is running version 9.9.1 so ignore FIPs failures on it
			slog.Warn("Skipping poller", slog.String("pollerName", pollerName))
			continue
		}
		port, _ := conf.GetLastPromPort(pollerName, true)
		portString := strconv.Itoa(port)
		var validCounters = 0
		uniqueSetOfMetricLabels := make(map[string]bool)
		sb, err2 := request.GetResponse("http://localhost:" + strings.TrimSpace(portString) + "/metrics")
		if err2 != nil {
			t.Fatalf("Unable to get metric data for %s", pollerName)
		}
		rows := strings.Split(sb, "\n")
		for i := range rows {
			row := rows[i]
			if row == "" {
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
				metricName := row[:openBracket]
				key := metricAndLabelKey(metricName, row[openBracket+1:])
				if uniqueSetOfMetricLabels[key] {
					_, ok := skipDuplicates[metricName]
					if ok {
						continue
					}
					duplicateMetrics = append(duplicateMetrics,
						fmt.Sprintf("Duplicate metric poller=%s, got >1 want 1 of %s", pollerName, key))
				}
				uniqueSetOfMetricLabels[key] = true
				metricValue, _ := strconv.Atoi(strings.TrimSpace(row[(closeBracket + 1):]))
				if metricValue > 0 {
					validCounters++
				}
			} else {
				slog.Error("Invalid string data found in the metric output", slog.String("row", row))
			}
		}
		if validCounters == 0 {
			panic("Empty values found for all counters for poller " + pollerName)
		}
		slog.Info(
			"Valid Counters for poller",
			slog.Int("numCounters", validCounters),
			slog.String("poller", pollerName),
		)
	}
	sort.Strings(duplicateMetrics)
	for _, dupMetric := range duplicateMetrics {
		t.Error(dupMetric)
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
			labels = append(labels, fmt.Sprintf(`%s=%q`, label, labelValue))
			scanner = labelEnd + 1
			if string(rest[scanner]) == "," {
				scanner++
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
