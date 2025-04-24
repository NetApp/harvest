/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package util

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/go-version"
	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sys/unix"
	"log/slog"
	"maps"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	BILLION                  = 1_000_000_000
	TopresourceConstant      = "999999"
	RangeConstant            = "888888"
	RangeReverseConstant     = "10d6h54m48s"
	IntervalConstant         = "777777"
	IntervalDurationConstant = "666666"
)

var arrayRegex = regexp.MustCompile(`^([a-zA-Z][\w.]*)(\.[0-9#])`)

var IsONTAPCollector = map[string]struct{}{
	"ZapiPerf": {},
	"Zapi":     {},
	"Rest":     {},
	"RestPerf": {},
	"KeyPerf":  {},
	"Ems":      {},
}

var IsCollector = map[string]struct{}{
	"CiscoRest":   {},
	"Ems":         {},
	"KeyPerf":     {},
	"Rest":        {},
	"RestPerf":    {},
	"Simple":      {},
	"StorageGrid": {},
	"Unix":        {},
	"Zapi":        {},
	"ZapiPerf":    {},
}

func GetCollectorSlice() []string {
	return slices.Collect(maps.Keys(IsCollector))
}

func MinLen(elements [][]string) int {
	var smallest, i int
	smallest = len(elements[0])
	for i = 1; i < len(elements); i++ {
		if len(elements[i]) < smallest {
			smallest = len(elements[i])
		}
	}
	return smallest
}

func MaxLen(elements [][]string) int {
	var largest, i int
	largest = len(elements[0])
	for i = 1; i < len(elements); i++ {
		if len(elements[i]) > largest {
			largest = len(elements[i])
		}
	}
	return largest
}

func AllSame(elements [][]string, k int) bool {
	var i int
	for i = 1; i < len(elements); i++ {
		if elements[i][k] != elements[0][k] {
			return false
		}
	}
	return true
}

var profRegex = regexp.MustCompile(`--profiling (\d+)`)
var promRegex = regexp.MustCompile(`--promPort (\d+)`)

func GetPollerStatuses() ([]PollerStatus, error) {
	result := make([]PollerStatus, 0)
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range processes {
		line, err := p.Cmdline()
		if err != nil {
			if !errors.Is(err, unix.EINVAL) && !errors.Is(err, unix.ENOENT) {
				fmt.Printf("Unable to read process cmdline pid=%d err=%v\n", p.Pid, err)
			}
			continue
		}
		if !strings.Contains(line, "poller --poller ") {
			continue
		}

		args, err := p.CmdlineSlice()

		if err != nil {
			if !errors.Is(err, unix.EINVAL) && !errors.Is(err, unix.ENOENT) {
				fmt.Printf("Unable to read process cmdline pid=%d err=%v\n", p.Pid, err)
			}
			continue
		}

		name := ""
		for i, arg := range args {
			if arg == "--poller" && i+1 < len(args) {
				name = args[i+1]
				break
			}
		}

		if name == "" {
			continue
		}
		s := PollerStatus{
			Name:   name,
			Pid:    p.Pid,
			Status: "running",
		}
		promMatches := promRegex.FindStringSubmatch(line)
		if len(promMatches) > 0 {
			s.PromPort = promMatches[1]
		}
		profMatches := profRegex.FindStringSubmatch(line)
		if len(profMatches) > 0 {
			s.ProfilingPort = profMatches[1]
		}
		result = append(result, s)
	}
	return result, nil
}

func FindLocalIP() (string, error) {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return "", err
	}
	defer func(conn net.Conn) { _ = conn.Close() }(conn)
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func CheckCert(certPath string, name string, configPath string, logger *slog.Logger) {
	if certPath == "" {
		logger.Error("TLS is enabled but cert path is empty",
			slog.String("config", configPath),
			slog.String(name, certPath),
		)
		os.Exit(1)
	}
	absPath := certPath
	if _, err := os.Stat(absPath); err != nil {
		logger.Error("TLS is enabled but cert path is invalid",
			slogx.Err(err),
			slog.String("config", configPath),
			slog.String(name, certPath),
		)
		os.Exit(1)
	}
}

type Status string

const (
	StatusStopped        Status = "stopped"
	StatusStoppingFailed Status = "stopping failed"
	StatusRunning        Status = "running"
	StatusNotRunning     Status = "not running"
	StatusKilled         Status = "killed"
	StatusAlreadyExited  Status = "already exited"
	StatusDisabled       Status = "disabled"
)

type PollerStatus struct {
	Name          string
	Status        Status
	Pid           int32
	ProfilingPort string
	PromPort      string
}

// Intersection returns things from b that are common and missing with a
func Intersection(a []string, b []string) ([]string, []string) {
	matches := make([]string, 0)
	misses := make([]string, 0)
	hash := make(map[string]bool)

	for _, aa := range a {
		hash[aa] = true
	}

	for _, bb := range b {
		if _, found := hash[bb]; found {
			matches = append(matches, bb)
		} else {
			misses = append(misses, bb)
		}
	}

	return matches, misses
}

// ParseMetric parses display name and type of field and metric type from the raw name of the metric as defined in (sub)template.
// Users can rename a metric with "=>" (e.g., some_long_metric_name => short).
// Trailing "^" characters are ignored/cleaned as they have special meaning in some collectors.
func ParseMetric(rawName string) (string, string, string, string) {
	var (
		name, display string
		values        []string
	)
	metricType := ""
	// Ex: last_transfer_duration(duration) => last_transfer_duration
	if values = strings.SplitN(rawName, "=>", 2); len(values) == 2 {
		name = strings.TrimSpace(values[0])
		display = strings.TrimSpace(values[1])
		name, metricType = ParseMetricType(name)
	} else {
		name = rawName
		display = strings.ReplaceAll(rawName, ".", "_")
		display = strings.ReplaceAll(display, "-", "_")
	}

	if strings.HasPrefix(name, "^^") {
		return strings.TrimPrefix(name, "^^"), strings.TrimPrefix(display, "^^"), "key", ""
	}

	if strings.HasPrefix(name, "^") {
		return strings.TrimPrefix(name, "^"), strings.TrimPrefix(display, "^"), "label", ""
	}

	return name, display, "float", metricType
}

func ParseMetricType(metricName string) (string, string) {
	metricTypeRegex := regexp.MustCompile(`(.*)\((.*?)\)`)
	match := metricTypeRegex.FindAllStringSubmatch(metricName, -1)
	if match != nil {
		// For last_transfer_duration(duration), name would have 'last_transfer_duration' and metricType would have 'duration'.
		name := match[0][1]
		metricType := match[0][2]
		return name, metricType
	}
	return metricName, ""
}

func SumNumbers(s []float64) float64 {
	var total float64
	for _, num := range s {
		total += num
	}
	return total
}

// Max returns 0 when passed an empty slice, slices.Max panics if input is empty,
// This function can be removed once all callers are checked for empty slices
func Max(input []float64) float64 {
	if len(input) > 0 {
		return slices.Max(input)
	}
	return 0
}

// Min returns 0 when passed an empty slice, slices.Min panics if input is empty,
// This function can be removed once all callers are checked for empty slices
func Min(input []float64) float64 {
	if len(input) > 0 {
		return slices.Min(input)
	}
	return 0
}

func Avg(input []float64) float64 {
	if len(input) > 0 {
		return SumNumbers(input) / float64(len(input))
	}
	return 0
}

func ParseZAPIDisplay(obj string, path []string) string {
	var (
		ignore = map[string]int{"attributes": 0, "info": 0, "list": 0, "details": 0, "storage": 0}
		added  = map[string]int{}
		words  []string
	)

	for _, w := range strings.Split(obj, "_") {
		ignore[w] = 0
	}

	for _, attribute := range path {
		split := strings.Split(attribute, "-")
		for _, word := range split {
			if word == obj {
				continue
			}
			if _, exists := ignore[word]; exists {
				continue
			}
			if _, exists := added[word]; exists {
				continue
			}
			words = append(words, word)
			added[word] = 0
		}
	}
	return strings.Join(words, "_")
}

func AddIntString(input string, value int) string {
	i, _ := strconv.Atoi(input)
	i += value
	return strconv.FormatInt(int64(i), 10)
}

var metricReplacer = strings.NewReplacer("\n", "", " ", "", "\"", "")

func ArrayMetricToString(value string) string {
	s := metricReplacer.Replace(value)

	openBracket := strings.Index(s, "[")
	closeBracket := strings.Index(s, "]")
	if openBracket > -1 && closeBracket > -1 {
		return s[openBracket+1 : closeBracket]
	}
	return value
}

func GetQueryParam(href string, query string) (string, error) {
	u, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	v := u.Query()
	mr := v.Get(query)
	return mr, nil
}

func EncodeURL(href string) (string, error) {
	u, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	u.RawQuery = u.Query().Encode()
	return u.RequestURI(), nil
}

func HasDuplicates(slice []string) bool {
	encountered := map[string]bool{}

	for _, v := range slice {
		if encountered[v] {
			return true
		}
		encountered[v] = true
	}

	return false
}

func GetURLWithoutHost(r *http.Request) string {
	urlWithoutHost := r.URL.Path
	if r.URL.RawQuery != "" {
		urlWithoutHost += "?" + r.URL.RawQuery
	}
	return urlWithoutHost
}

// VersionAtLeast checks if the provided currentVersion of the cluster
// is greater than or equal to the provided minimum version (minVersion).
func VersionAtLeast(currentVersion string, minVersion string) (bool, error) {
	parsedClusterVersion, err := version.NewVersion(currentVersion)
	if err != nil {
		return false, fmt.Errorf("invalid current version: %w", err)
	}

	minSupportedVersion, err := version.NewVersion(minVersion)
	if err != nil {
		return false, fmt.Errorf("invalid minimum version: %w", err)
	}

	// Check if the current version is greater than or equal to the minimum version
	// and return the result
	return parsedClusterVersion.GreaterThanOrEqual(minSupportedVersion), nil
}

// IsPublicAPI returns false if api endpoint has private keyword in it else true
func IsPublicAPI(query string) bool {
	return !strings.Contains(query, "private")
}

func HandleArrayFormat(name string) string {
	matches := arrayRegex.FindStringSubmatch(name)
	if len(matches) > 2 {
		return matches[1]
	}
	return name
}

func SafeConvertToInt32(in int) (int32, error) {
	if in > math.MaxInt32 {
		return 0, fmt.Errorf("input %d is too large to convert to int32", in)
	}
	return int32(in), nil // #nosec G115
}

func Format(query string) string {
	replacedQuery := strings.ReplaceAll(query, "$TopResources", TopresourceConstant)
	replacedQuery = strings.ReplaceAll(replacedQuery, "$__range", RangeConstant)
	replacedQuery = strings.ReplaceAll(replacedQuery, "$__interval", IntervalConstant)
	replacedQuery = strings.ReplaceAll(replacedQuery, "${Interval}", IntervalDurationConstant)

	path, err := exec.LookPath("promtool")
	if err != nil {
		fmt.Printf("ERR failed to find promtool")
		return query
	}
	command := exec.Command(path, "--experimental", "promql", "format", replacedQuery)
	output, err := command.CombinedOutput()
	updatedQuery := strings.TrimSuffix(string(output), "\n")
	if strings.HasPrefix(updatedQuery, "  ") {
		updatedQuery = strings.TrimLeft(updatedQuery, " ")
	}
	if err != nil {
		// An exit code can't be used since we need to ignore metrics that are not formatted but can't change
		fmt.Printf("ERR formating metrics query=%s err=%v output=%s", query, err, string(output))
		return query
	}

	if len(output) == 0 {
		return query
	}

	updatedQuery = strings.ReplaceAll(updatedQuery, TopresourceConstant, "$TopResources")
	updatedQuery = strings.ReplaceAll(updatedQuery, RangeReverseConstant, "$__range")
	updatedQuery = strings.ReplaceAll(updatedQuery, IntervalConstant, "$__interval")
	updatedQuery = strings.ReplaceAll(updatedQuery, IntervalDurationConstant, "${Interval}")
	return updatedQuery
}
