/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package util

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"
	"net"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func MinLen(elements [][]string) int {
	var min, i int
	min = len(elements[0])
	for i = 1; i < len(elements); i++ {
		if len(elements[i]) < min {
			min = len(elements[i])
		}
	}
	return min
}

func MaxLen(elements [][]string) int {
	var max, i int
	max = len(elements[0])
	for i = 1; i < len(elements); i++ {
		if len(elements[i]) > max {
			max = len(elements[i])
		}
	}
	return max
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

func EqualStringSlice(a, b []string) bool {
	var i int
	if len(a) != len(b) {
		return false
	}
	for i = 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

var pollerRegex = regexp.MustCompile(`poller\s+--poller\s+(.*?)\s`)
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
		matches := pollerRegex.FindStringSubmatch(line)
		if len(matches) != 2 {
			continue
		}
		s := PollerStatus{
			Name:   matches[1],
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

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
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

func CheckCert(certPath string, name string, configPath string, logger zerolog.Logger) {
	if certPath == "" {
		logger.Fatal().
			Str("config", configPath).
			Str(name, certPath).
			Msg("TLS is enabled but cert path is empty.")
	}
	absPath := certPath
	if _, err := os.Stat(absPath); err != nil {
		logger.Fatal().
			Str("config", configPath).
			Str(name, absPath).
			Msg("TLS is enabled but cert path is invalid.")
	}
}

// SaveConfig adds or updates the Grafana token in the harvest.yml config
// and saves it to fp. The Yaml marshaller is ued so comments are preserved
func SaveConfig(fp string, token string) error {
	contents, err := os.ReadFile(fp)
	if err != nil {
		return err
	}
	root := &yaml.Node{}
	err = yaml.Unmarshal(contents, root)
	if err != nil {
		return err
	}

	// Three cases to consider:
	//	1. Tools is missing
	//  2. Tools is present but empty (nil)
	//  3. Tools is present - overwrite value
	tokenSet := false
	if len(root.Content) > 0 {
		nodes := root.Content[0].Content
		for i, n := range nodes {
			if n.Tag == "!!map" && len(n.Content) > 1 && n.Content[0].Value == "grafana_api_token" {
				// Case 3
				n.Content[1].SetString(token)
				tokenSet = true
				break
			}
			if n.Value == "Tools" {
				if i+1 < len(nodes) && nodes[i+1].Tag == "!!null" {
					// Case 2
					n2 := yaml.Node{}
					_ = n2.Encode(map[string]string{"grafana_api_token": token})
					nodes[i+1] = &n2
					tokenSet = true
					break
				}
			}
		}
		if !tokenSet {
			// Case 1
			tools := yaml.Node{}
			tools.SetString("Tools")
			nodes = append(nodes, &tools)

			nToken := yaml.Node{}
			_ = nToken.Encode(map[string]string{"grafana_api_token": token})
			nodes = append(nodes, &nToken)
			root.Content[0].Content = nodes
		}
	}
	marshal, err := yaml.Marshal(root)
	if err != nil {
		return err
	}
	return os.WriteFile(fp, marshal, 0600)
}

type Status string

const (
	StatusStopped        Status = "stopped"
	StatusStoppingFailed Status = "stopping failed"
	StatusRunning        Status = "running"
	StatusNotRunning     Status = "not running"
	StatusKilled         Status = "killed"
	StatusAlreadyExited  Status = "already exited"
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
// Users can rename a metric with "=>" (e.g. some_long_metric_name => short).
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

func Max(input []float64) float64 {
	if len(input) > 0 {
		max := input[0]
		for _, v := range input {
			if v > max {
				max = v
			}
		}
		return max
	}
	return 0
}

func Min(input []float64) float64 {
	if len(input) > 0 {
		min := input[0]
		for _, v := range input {
			if v < min {
				min = v
			}
		}
		return min
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
	i = i + value
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
		} else {
			encountered[v] = true
		}
	}

	return false
}

func GetSortedKeys(m map[string]string) []string {
	var sortedKeys []string
	for k := range m {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}
