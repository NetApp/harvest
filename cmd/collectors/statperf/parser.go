package statperf

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"regexp"
	"strings"
)

var unitRegex = regexp.MustCompile(`^([\d.]+)(us|%)$`)

type CounterProperty struct {
	Counter     string
	Name        string
	BaseCounter string
	Properties  string
	Type        string
	Deprecated  string
	ReplacedBy  string
}

func filterNonEmpty(input string) []string {
	scanner := bufio.NewScanner(strings.NewReader(input))
	var results []string
	for scanner.Scan() {
		if trimmed := strings.TrimSpace(scanner.Text()); trimmed != "" {
			results = append(results, trimmed)
		}
	}
	return results
}

func (s *StatPerf) parseCounters(input string) (map[string]CounterProperty, error) {
	linesFiltered := filterNonEmpty(input)

	// Search for the header row, which is expected to have at least 9 columns when split.
	var headerIndex = -1
	for i, line := range linesFiltered {
		fields := strings.Split(line, util.StatPerfSeparator)
		if len(fields) >= 9 {
			// Check if this header row contains a known header word like "counter"
			lower := strings.ToLower(line)
			if strings.Contains(lower, "counter") {
				headerIndex = i
				break
			}
		}
	}

	if headerIndex < 0 {
		return nil, errors.New("no valid header row found")
	}

	// Expect at least one additional header row to follow the header row.
	if len(linesFiltered) < headerIndex+2 {
		return nil, errors.New("not enough header rows following the detected header")
	}

	// counter rows start after the second header row.
	dataStart := headerIndex + 2
	if dataStart >= len(linesFiltered) {
		return nil, errors.New("no data rows found")
	}

	counters := make(map[string]CounterProperty)

	for _, row := range linesFiltered[dataStart:] {
		fields := strings.Split(row, util.StatPerfSeparator)
		if len(fields) < 9 {
			s.Logger.Warn("skipping incomplete row", slog.String("row", row))
			continue
		}

		cp := CounterProperty{
			Counter:     strings.TrimSpace(fields[2]),
			Name:        strings.TrimSpace(fields[3]),
			BaseCounter: strings.TrimSpace(fields[4]),
			Properties:  strings.TrimSpace(fields[5]),
			Type:        strings.TrimSpace(fields[6]),
			Deprecated:  strings.TrimSpace(fields[7]),
			ReplacedBy:  strings.TrimSpace(fields[8]),
		}
		counters[cp.Counter] = cp
	}
	return counters, nil
}

type InstanceInfo struct {
	Instance     string
	InstanceUUID string
}

func (s *StatPerf) parseInstances(output string) []InstanceInfo {
	linesFiltered := filterNonEmpty(output)

	// Locate the header row: look for a row that, when split, returns at least 6 fields and contains "instance"
	var headerIndex = -1
	for i, line := range linesFiltered {
		fields := strings.Split(line, util.StatPerfSeparator)
		if len(fields) >= 6 {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "instance") {
				headerIndex = i
				break
			}
		}
	}

	if headerIndex < 0 {
		s.Logger.Warn("no valid header found in instance output")
		return nil
	}

	// Expect the table to have at least two header rows.
	if len(linesFiltered) < headerIndex+2 {
		s.Logger.Warn("not enough header rows in instance output")
		return nil
	}

	dataStart := headerIndex + 2
	if dataStart >= len(linesFiltered) {
		s.Logger.Warn("no data rows found in instance output")
		return nil
	}

	estimatedRows := len(linesFiltered) - dataStart
	results := make([]InstanceInfo, 0, estimatedRows)
	// Process data rows.
	for _, row := range linesFiltered[dataStart:] {
		fields := strings.Split(row, util.StatPerfSeparator)
		if len(fields) < 6 {
			s.Logger.Warn("skipping incomplete row", slog.String("row", row))
			continue
		}
		inst := InstanceInfo{
			Instance:     strings.TrimSpace(fields[2]),
			InstanceUUID: strings.TrimSpace(fields[4]),
		}
		results = append(results, inst)
	}

	return results
}

type Row struct {
	Instance string `json:"instance"`
	Counter  string `json:"counter"`
	Value    string `json:"value"`
}

func parseRows(output string, logger *slog.Logger) ([]Row, error) {
	linesFiltered := filterNonEmpty(output)

	// Find the header row by scanning for a line that splits into 4+ fields and contains "counter".
	headerIndex := -1
	for i, line := range linesFiltered {
		fields := strings.Split(line, util.StatPerfSeparator)
		if len(fields) >= 4 {
			if strings.Contains(strings.ToLower(line), "counter") {
				headerIndex = i
				break
			}
		}
	}

	if headerIndex < 0 {
		return nil, errors.New("no valid header row found")
	}

	// Ensure there is at least one additional header row following the header.
	if len(linesFiltered) < headerIndex+2 {
		return nil, errors.New("not enough header rows in data")
	}

	// Data rows start after headerIndex + 2.
	dataStart := headerIndex + 2
	if dataStart >= len(linesFiltered) {
		return nil, errors.New("no data rows found")
	}

	estimatedRows := len(linesFiltered) - dataStart
	rows := make([]Row, 0, estimatedRows)

	for _, line := range linesFiltered[dataStart:] {
		fields := strings.Split(line, util.StatPerfSeparator)
		if len(fields) < 4 {
			logger.Warn("skipping incomplete row", slog.String("row", line))
			continue
		}
		row := Row{
			Instance: strings.TrimSpace(fields[1]),
			Counter:  strings.TrimSpace(fields[2]),
			Value:    removeUnitRegex(strings.TrimSpace(fields[3])),
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func removeUnitRegex(s string) string {
	matches := unitRegex.FindStringSubmatch(strings.TrimSpace(s))
	if len(matches) == 3 {
		return matches[1]
	}
	return s
}

// GroupInstances The first counter encountered
// in the first row of a group is captured. When that counter appears again,
// it indicates a new instance.
func groupInstances(rows []Row) ([]map[string]string, error) {
	var results []map[string]string
	currentGroup := make(map[string]string)
	var firstCounterName string

	for _, row := range rows {
		lowerCounter := strings.ToLower(row.Counter)

		if len(currentGroup) == 0 {
			firstCounterName = lowerCounter
		}

		// If the first counter repeats, then finish the current group
		// and start a new one.
		if lowerCounter == firstCounterName && len(currentGroup) > 0 {
			if _, exists := currentGroup[firstCounterName]; exists {
				results = append(results, currentGroup)
				currentGroup = make(map[string]string)
				firstCounterName = lowerCounter
			}
		}

		currentGroup[row.Counter] = row.Value
	}
	if len(currentGroup) > 0 {
		results = append(results, currentGroup)
	}
	if len(results) == 0 {
		return nil, errors.New("no valid groups found")
	}
	return results, nil
}

func (s *StatPerf) parseData(output string, logger *slog.Logger) (gjson.Result, error) {
	rows, err := parseRows(output, logger)
	if err != nil {
		return gjson.Result{}, err
	}
	groups, err := groupInstances(rows)
	if err != nil {
		return gjson.Result{}, err
	}
	groupedJSON, err := json.MarshalIndent(groups, "", "  ")
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.ParseBytes(groupedJSON), nil
}
