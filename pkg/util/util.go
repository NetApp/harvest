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
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func MinLen(elements [][]string) int {
	var min, i int
	min = len(elements[0])
	for i = 1; i < len(elements); i += 1 {
		if len(elements[i]) < min {
			min = len(elements[i])
		}
	}
	return min
}

func MaxLen(elements [][]string) int {
	var max, i int
	max = len(elements[0])
	for i = 1; i < len(elements); i += 1 {
		if len(elements[i]) > max {
			max = len(elements[i])
		}
	}
	return max
}

func AllSame(elements [][]string, k int) bool {
	var i int
	for i = 1; i < len(elements); i += 1 {
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
	for i = 0; i < len(a); i += 1 {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func GetCmdLine(pid int32) (string, error) {
	newProcess, err := process.NewProcess(pid)
	if err != nil {
		return "", err
	}
	return newProcess.Cmdline()
}

func RemoveEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func GetPid(pollerName string) ([]int32, error) {
	// ($|\s) is included to match the poller name
	// followed by a space or end of line - that way unix1 does not match unix11
	search := fmt.Sprintf(`\-\-poller %s($|\s)`, pollerName)
	if runtime.GOOS == "darwin" {
		search = fmt.Sprintf(`\-\-poller %s([[:space:]]+|$)`, pollerName)
	}
	return GetPids(search)
}

func GetPids(search string) ([]int32, error) {
	var result []int32
	var ee *exec.ExitError
	var pe *os.PathError
	cmd := exec.Command("pgrep", "-f", search)
	data, err := cmd.Output()
	if errors.As(err, &ee) {
		if ee.Stderr != nil {
			fmt.Printf("Exit error stderr=%s\n", ee.Stderr)
		}
		return result, nil // ran, but non-zero exit code
	} else if errors.As(err, &pe) {
		return result, err // "no such file ...", "permission denied" etc.
	} else if err != nil {
		return result, err // something unexpected happened!
	}
	out := string(data)
	pids := RemoveEmptyStrings(strings.Split(out, "\n"))
	for _, pid := range pids {
		p64, err := strconv.ParseInt(strings.TrimSpace(pid), 10, 32)
		if err != nil {
			return result, err
		}

		// Validate this is a Harvest process
		// env check does not work on Darwin or Unix when running as non-root
		result = append(result, int32(p64))
	}
	return result, err
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
			if !errors.Is(err, unix.EINVAL) {
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

func ContainsWholeWord(source string, search string) bool {
	if len(source) == 0 || len(search) == 0 {
		return false
	}
	fields := strings.Fields(source)
	for _, w := range fields {
		if w == search {
			return true
		}
	}
	return false
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
	contents, err := ioutil.ReadFile(fp)
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
	return ioutil.WriteFile(fp, marshal, 0644)
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
