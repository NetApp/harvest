package statperf

import (
	"encoding/json"
	"errors"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var unitRegex = regexp.MustCompile(`^([\d.]+)(us|ms|%)$`)
var aggRegex = regexp.MustCompile(`Number of Constituents:\s*\d+\s+\(([^)]+)\)`)
var dividedRegex = regexp.MustCompile(`^[-\s]+$`)

type CounterProperty struct {
	Counter     string
	Name        string
	BaseCounter string
	Properties  string
	Type        string
	Deprecated  string
	ReplacedBy  string
	Unit        string
	Description string
}

func (s *StatPerf) parseCounters(input string) (map[string]CounterProperty, error) {
	linesFiltered := FilterNonEmpty(input)

	// Search for the header row, which is expected to have at least 9 columns when split.
	var headerIndex = -1
	for i, line := range linesFiltered {
		fields := strings.Split(line, collector.StatPerfSeparator)
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
		fields := strings.Split(row, collector.StatPerfSeparator)
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

func (s *StatPerf) parseInstances(input string) ([]InstanceInfo, error) {
	linesFiltered := FilterNonEmpty(input)

	// Locate the header row: look for a row that, when split, returns at least 6 fields and contains "instance"
	var headerIndex = -1
	for i, line := range linesFiltered {
		fields := strings.Split(line, collector.StatPerfSeparator)
		if len(fields) >= 6 {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "instance") {
				headerIndex = i
				break
			}
		}
	}

	if headerIndex < 0 {
		return nil, errors.New("no valid header row found")
	}

	// Expect the table to have at least two header rows.
	if len(linesFiltered) < headerIndex+2 {
		return nil, errors.New("not enough header rows in data")
	}

	dataStart := headerIndex + 2
	if dataStart >= len(linesFiltered) {
		return nil, errors.New("no data rows found")
	}

	estimatedRows := len(linesFiltered) - dataStart
	results := make([]InstanceInfo, 0, estimatedRows)
	// Process data rows.
	for _, row := range linesFiltered[dataStart:] {
		fields := strings.Split(row, collector.StatPerfSeparator)
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

	return results, nil
}

func removeUnitRegex(s string, counter string) string {
	if counter == "instance_name" || counter == "instance_uuid" {
		return s
	}
	matches := unitRegex.FindStringSubmatch(s)
	if len(matches) == 3 {
		return matches[1]
	}
	return s
}

type Row struct {
	Instance    string `json:"instance"`
	Counter     string `json:"counter"`
	Value       string `json:"value"`
	Aggregation string `json:"aggregation,omitempty"`
}

// getIndent returns the count of leading space characters (and tabs treated as spaces) in a string.
func getIndent(s string) int {
	return len(s) - len(strings.TrimLeft(s, " \t"))
}

// FilterNonEmpty splits input into lines and returns only nonblank ones.
func FilterNonEmpty(input string) []string {
	var lines []string
	for line := range strings.Lines(input) {
		line = strings.TrimSuffix(line, "\n")
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// getDividerWidth returns the number of '-' characters before the first space in the divider line.
func getDividerWidth(line string) int {
	trim := strings.TrimLeft(line, " ")
	index := strings.Index(trim, " ")
	if index == -1 {
		return len(trim)
	}
	return index
}

// parseData processes an input string, extracts and groups rows (by instance),
// and returns a gjson.Result.
func (s *StatPerf) parseData(input string) (gjson.Result, error) {
	groups, err := parseRows(input, s.Logger)
	if err != nil {
		return gjson.Result{}, err
	}

	// Marshal groups to indented JSON.
	groupedJSON, err := json.Marshal(groups)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.ParseBytes(groupedJSON), nil
}

func parseRows(input string, logger *slog.Logger) ([]map[string]string, error) {
	defaultTimestamp := float64(time.Now().UnixNano() / collector.BILLION)
	var timestamp float64
	lines := FilterNonEmpty(input)
	var groups []map[string]string

	var aggregation string
	currentGroup := make(map[string]string)

	inTable := false
	var tableLines []string
	dividerWidth := 0

	flushTable := func() {
		if len(tableLines) == 0 {
			return
		}

		seenCounters := set.New()

		for i := 0; i < len(tableLines); i++ {
			line := tableLines[i]
			curRowLine := strings.TrimSpace(line)
			if curRowLine == "" || strings.Contains(curRowLine, "entries were displayed") {
				continue
			}
			tokens := strings.Fields(curRowLine)
			if len(tokens) < 2 {
				logger.Warn("skipping unexpected line", slog.String("row", curRowLine))
				continue
			}
			counter := strings.Join(tokens[:len(tokens)-1], " ")
			value := tokens[len(tokens)-1]

			// Check for continuation lines.
			for i+1 < len(tableLines) {
				nextLineRaw := tableLines[i+1]
				nextTokens := strings.Fields(nextLineRaw)
				if len(nextTokens) != 1 {
					break
				}
				indent := getIndent(nextLineRaw)
				if dividerWidth > 0 && indent < dividerWidth {
					counter += nextTokens[0]
				} else {
					value += nextTokens[0]
				}
				i++
			}

			// Check for duplicate counters
			// The object `object_store_server` has a different table format where multiple node
			// counters are present in the same table, leading to duplicate counters.
			// StatPerf does not handle this situation and prints warning messages.
			// So far, we have observed this issue only with this object.
			trimmedCounter := strings.TrimSpace(counter)
			if seenCounters.Has(trimmedCounter) {
				logger.Warn("duplicate counter detected", slog.String("counter", trimmedCounter), slog.String("value", value))
			} else {
				seenCounters.Add(trimmedCounter)
			}

			currentGroup[trimmedCounter] = removeUnitRegex(strings.TrimSpace(value), counter)
		}

		// Save the aggregation and instance info for this group.
		if aggregation != "" {
			currentGroup["_aggregation"] = aggregation
		}
		if timestamp != 0 {
			currentGroup["timestamp"] = strconv.FormatFloat(timestamp, 'f', -1, 64)
		} else {
			currentGroup["timestamp"] = strconv.FormatFloat(defaultTimestamp, 'f', -1, 64)
		}
		groups = append(groups, currentGroup)
		tableLines = []string{}
		currentGroup = make(map[string]string)
		timestamp = 0
	}

	// Process each line of the input.
	for i := 0; i < len(lines); i++ {
		trimLine := strings.TrimSpace(lines[i])
		if trimLine == "" {
			continue
		}

		// New Object: flush current tableLines.
		if strings.HasPrefix(trimLine, "Object:") {
			flushTable()
			inTable = false
			aggregation = ""
			continue
		}

		if strings.HasPrefix(trimLine, "End-time:") {
			endTimeStr := strings.TrimPrefix(trimLine, "End-time:")
			endTimeStr = strings.TrimSpace(endTimeStr)
			endTime, err := time.Parse("1/2/2006 15:04:05", endTimeStr)
			if err != nil {
				logger.Warn("unable to parse end-time", slog.String("end-time", endTimeStr))
				continue
			}
			timestamp = float64(endTime.UnixNano()) / collector.BILLION
			continue
		}

		// Skip lines with timing or scope.
		if strings.HasPrefix(trimLine, "Start-time:") ||
			strings.HasPrefix(trimLine, "Scope:") ||
			strings.HasPrefix(trimLine, "Instance:") {
			continue
		}

		// Capture aggregation information.
		if strings.HasPrefix(trimLine, "Number of Constituents:") {
			matches := aggRegex.FindStringSubmatch(trimLine)
			if len(matches) >= 2 {
				aggregation = matches[1]
			} else {
				aggregation = ""
			}
			continue
		}

		// Identify the beginning of a table block.
		if strings.HasPrefix(trimLine, "Counter") && strings.Contains(trimLine, "Value") {
			inTable = true
			// Next nonempty line is assumed to be the divider.
			if i+1 < len(lines) && dividedRegex.MatchString(strings.TrimSpace(lines[i+1])) {
				dividerWidth = getDividerWidth(lines[i+1])
				i++ // Skip divider line.
			}
			continue
		}

		// If inside a table block, accumulate the raw line.
		if inTable {
			// If new metadata (for example, "Object:") begins, flush the current table.
			if strings.HasPrefix(trimLine, "Object:") {
				flushTable()
				inTable = false
				i-- // reprocess current line as metadata.
				continue
			}
			tableLines = append(tableLines, lines[i])
		}
	}
	flushTable()

	if len(groups) == 0 {
		return nil, errors.New("no data found")
	}
	return groups, nil
}
