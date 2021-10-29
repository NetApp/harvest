/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package util

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/process"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"reflect"
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

var pidAndPollerRegex = regexp.MustCompile(`(\d+).*?poller.*?--poller (.*?) .*?--promPort (\d+)`)
var profRegex = regexp.MustCompile(`--profiling (\d+)`)
var promRegex = regexp.MustCompile(`--promPort (\d+)`)

func GetPollerStatuses() ([]PollerStatus, error) {
	var (
		result []PollerStatus
		ee     *exec.ExitError
		pe     *os.PathError
		cmd    *exec.Cmd
	)
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("pgrep", "-fl", "poller ")
	} else {
		cmd = exec.Command("pgrep", "-fa", "poller ")
	}
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
	// read each line of out and extract the pid, poller name, and port
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		matches := pidAndPollerRegex.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}
		pid, err := strconv.ParseInt(matches[1], 10, 32)
		if err != nil {
			continue
		}
		s := PollerStatus{
			Name:   matches[2],
			Pid:    int32(pid),
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

func Intersection(a interface{}, b interface{}) ([]interface{}, []interface{}) {
	matches := make([]interface{}, 0)
	misses := make([]interface{}, 0)
	hash := make(map[interface{}]bool)
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	for i := 0; i < av.Len(); i++ {
		el := av.Index(i).Interface()
		hash[el] = true
	}

	for i := 0; i < bv.Len(); i++ {
		el := bv.Index(i).Interface()
		if _, found := hash[el]; found {
			matches = append(matches, el)
		} else {
			misses = append(misses, el)
		}
	}

	return matches, misses
}
