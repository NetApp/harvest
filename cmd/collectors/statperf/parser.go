package statperf

import (
	"encoding/json"
	"errors"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/node"
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
var arrayOneToManyRegex = regexp.MustCompile(`^([^,]+),"([^"]+)"$`)
var arrayManyToManyRegex = regexp.MustCompile(`^"([^"]+)","([^"]+)"$`)
var arrayRegex = regexp.MustCompile(`^"([^"]+)"$`)

type CounterProperty struct {
	Counter     string
	Name        string
	BaseCounter string
	Properties  string
	Type        string
	Deprecated  string
	ReplacedBy  string
	Label       string
	LabelCount  int
	Description string
	Unit        string
}

func (s *StatPerf) ParseCounters(input string) (map[string]CounterProperty, error) {
	linesFiltered := FilterNonEmpty(input)

	// Search for the header row, which is expected to have at least 11 columns when split.
	var expectedFieldCount = 11
	var headerIndex = -1
	var headers []string
	for i, line := range linesFiltered {
		fields := strings.Split(line, collector.StatPerfSeparator)
		if len(fields) >= expectedFieldCount {
			// Check if this header row contains a known header word like "counter"
			lower := strings.ToLower(line)
			if strings.Contains(lower, "counter") {
				headerIndex = i
				headers = fields
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

	// Create a map of header names to their indices
	headerMap := make(map[string]int, len(headers))
	for index, header := range headers {
		headerMap[strings.TrimSpace(header)] = index
	}

	for _, row := range linesFiltered[dataStart:] {
		fields := strings.Split(row, collector.StatPerfSeparator)
		if len(fields) <= expectedFieldCount {
			s.Logger.Warn("skipping incomplete row", slog.String("row", row))
			continue
		}

		counterType := strings.TrimSpace(fields[headerMap["type"]])
		var combinedLabels []string

		labelField := strings.ToLower(node.Normalize(fields[headerMap["label"]]))
		if counterType == "array" {
			if match := arrayRegex.FindStringSubmatch(labelField); match != nil {
				labelParts := strings.Split(match[1], ",")
				combinedLabels = labelParts
			} else if match = arrayOneToManyRegex.FindStringSubmatch(labelField); match != nil {
				identifier := strings.TrimSpace(match[1])
				quotedLabels := strings.SplitSeq(match[2], ",")
				for label := range quotedLabels {
					combinedLabels = append(combinedLabels, label+"."+identifier)
				}
			} else if match = arrayManyToManyRegex.FindStringSubmatch(labelField); match != nil {
				firstSet := strings.Split(match[1], ",")
				secondSet := strings.Split(match[2], ",")
				for _, firstLabel := range firstSet {
					for _, secondLabel := range secondSet {
						combinedLabels = append(combinedLabels, secondLabel+"."+firstLabel)
					}
				}
			} else {
				// Fallback for simple comma-separated lists
				labelParts := strings.SplitSeq(labelField, ",")
				for label := range labelParts {
					label = strings.TrimSpace(label)
					if label != "" {
						combinedLabels = append(combinedLabels, label)
					}
				}
			}
		}

		cp := CounterProperty{
			Counter:     strings.TrimSpace(fields[headerMap["counter"]]),
			Name:        strings.TrimSpace(fields[headerMap["name"]]),
			BaseCounter: strings.TrimSpace(fields[headerMap["base-counter"]]),
			Label:       strings.Join(combinedLabels, ","),
			LabelCount:  len(combinedLabels),
			Description: strings.Trim(fields[headerMap["description"]], ` "'`),
			Properties:  strings.TrimSpace(fields[headerMap["properties"]]),
			Type:        counterType,
			Deprecated:  strings.TrimSpace(fields[headerMap["is-deprecated"]]),
			ReplacedBy:  strings.TrimSpace(fields[headerMap["replaced-by"]]),
			Unit:        strings.TrimSpace(fields[headerMap["unit"]]),
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
			Instance: strings.TrimSpace(fields[2]),
			// Remove quotes from InstanceUUID when present. UUIDs may be quoted in input data,
			// but during poll-data calls we receive InstanceUUID without quotes due to the
			// tabular format, so we remove surrounding quotes here.
			// Do not modify instance names: they are used as filter values in data-fetch calls
			// and must retain their quotes.
			InstanceUUID: strings.Trim(strings.TrimSpace(fields[4]), "\""),
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
	groups, err := s.parseRows(input)
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

func (s *StatPerf) parseRows(input string) ([]map[string]any, error) {
	defaultTimestamp := float64(time.Now().UnixNano() / collector.BILLION)
	var timestamp float64
	lines := FilterNonEmpty(input)
	var data []map[string]any

	var aggregation string
	currentGroup := make(map[string]any)
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
			// Split on first space - counter name has no spaces, value can have spaces
			spaceIndex := strings.Index(curRowLine, " ")
			if spaceIndex == -1 {
				s.Logger.Warn("skipping unexpected line - no space found", slog.String("row", curRowLine))
				continue
			}

			if spaceIndex >= len(curRowLine)-1 {
				s.Logger.Warn("skipping line - space at end", slog.String("row", curRowLine))
				continue
			}

			counter := strings.TrimSpace(curRowLine[:spaceIndex])
			var valueBuilder strings.Builder
			valueBuilder.WriteString(strings.TrimSpace(curRowLine[spaceIndex+1:]))

			if counter == "" {
				s.Logger.Warn("skipping line - empty counter", slog.String("row", curRowLine))
				continue
			}

			// Check if the counter is an array-like element
			if s.perfProp != nil && s.perfProp.counterInfo != nil {
				if c, exists := s.perfProp.counterInfo[counter]; exists {
					if c.counterType == "array" {
						arrayData := make(map[string]string)
						for range c.labelCount {
							nextLineRaw := preprocessArrayLine(tableLines[i+1])
							nextTokens := strings.Fields(nextLineRaw)
							if len(nextTokens) != 2 {
								break
							}
							arrayData[nextTokens[0]] = nextTokens[1]
							i++
						}
						currentGroup[counter] = arrayData
						continue
					}
				}
			}

			// Check for continuation lines.
			var cb strings.Builder
			cb.WriteString(counter)
			for i+1 < len(tableLines) {
				nextLineRaw := tableLines[i+1]
				nextTokens := strings.Fields(nextLineRaw)
				if len(nextTokens) != 1 {
					break
				}
				indent := getIndent(nextLineRaw)
				if dividerWidth > 0 && indent < dividerWidth {
					cb.WriteString(nextTokens[0])
				} else {
					valueBuilder.WriteString(nextTokens[0])
				}
				i++
			}
			counter = cb.String()
			value := valueBuilder.String()

			// Check for duplicate counters
			// The objects `object_store_server` and `smb2` have a different table format where multiple node
			// counters are present in the same table, leading to duplicate counters.
			// StatPerf does not handle this situation and prints warning messages.
			// So far, we have observed this issue only with this object.
			trimmedCounter := strings.TrimSpace(counter)
			if seenCounters.Has(trimmedCounter) {
				s.Logger.Warn("duplicate counter detected", slog.String("counter", trimmedCounter), slog.String("value", value))
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

		data = append(data, currentGroup)
		tableLines = []string{}
		currentGroup = make(map[string]any)
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

		if after, ok := strings.CutPrefix(trimLine, "End-time:"); ok {
			endTimeStr := after
			endTimeStr = strings.TrimSpace(endTimeStr)
			endTime, err := time.Parse("1/2/2006 15:04:05", endTimeStr)
			if err != nil {
				s.Logger.Warn("unable to parse end-time", slog.String("end-time", endTimeStr))
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

	if len(data) == 0 {
		return nil, errors.New("no data found")
	}
	return data, nil
}

func preprocessArrayLine(line string) string {
	if line == "" {
		return line
	}
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return line
	}
	modifiedText := strings.ToLower(node.Normalize(strings.Join(parts[:len(parts)-1], " ")))
	number := parts[len(parts)-1]
	result := modifiedText + " " + number
	return result
}
