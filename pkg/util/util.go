/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package util

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// HarvestTag is injected into a poller's environment to disambiguate the process
const HarvestTag = "IS_HARVEST=TRUE"

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

func readProcFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	result := string(bytes.ReplaceAll(data, []byte("\x00"), []byte(" ")))
	return result, nil
}

func GetEnviron(pid int) (string, error) {
	return readProcFile(fmt.Sprintf("/proc/%d/environ", pid))
}

func GetCmdLine(pid int) (string, error) {
	if runtime.GOOS == "darwin" {
		return darwinGetCmdLine(pid)
	}
	return readProcFile(fmt.Sprintf("/proc/%d/cmdline", pid))
}

func darwinGetCmdLine(pid int) (string, error) {
	bin, err := exec.LookPath("ps")
	if err != nil {
		return "", err
	}
	var ee *exec.ExitError
	var pe *os.PathError
	cmd := exec.Command(bin, "-x", "-o", "command", "-p", strconv.Itoa(pid))
	out, err := cmd.Output()
	if errors.As(err, &ee) {
		if ee.Stderr != nil {
			fmt.Printf("Exit error stderr=%s\n", ee.Stderr)
		}
		return "", nil // ran, but non-zero exit code
	} else if errors.As(err, &pe) {
		return "", err // "no such file ...", "permission denied" etc.
	} else if err != nil {
		return "", err // something really bad happened!
	}
	lines := strings.Split(string(out), "\n")
	var ret [][]string
	for _, line := range lines[1:] {
		var lr []string
		for _, word := range strings.Split(line, " ") {
			if word == "" {
				continue
			}
			lr = append(lr, strings.TrimSpace(word))
		}
		if len(lr) != 0 {
			ret = append(ret, lr)
		}
	}
	if ret == nil {
		return "", nil
	}
	return strings.Join(ret[0], " "), err
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

func GetPid(pollerName string) ([]int, error) {
	// ($|\s) is included to match the poller name
	// followed by a space or end of line - that way unix1 does not match unix11
	search := fmt.Sprintf(`\-\-poller %s($|\s)`, pollerName)
	if runtime.GOOS == "darwin" {
		search = fmt.Sprintf(`\-\-poller %s([[:space:]]+|$)`, pollerName)
	}
	return GetPids(search)
}

func GetPids(search string) ([]int, error) {
	var result []int
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
		return result, err // something really bad happened!
	}
	sdata := string(data)
	pids := RemoveEmptyStrings(strings.Split(sdata, "\n"))
	for _, pid := range pids {
		p, err := strconv.Atoi(strings.TrimSpace(pid))
		if err != nil {
			return result, err
		}

		// Validate this is a Harvest process
		if runtime.GOOS == "darwin" {
			// env check does not work on darwin
			result = append(result, p)
		} else {
			environ, err := GetEnviron(p)
			if err != nil {
				if errors.As(err, &pe) {
					// permission denied, no need to log
					continue
				}
				fmt.Printf("err reading environ for search=%s pid=%d err=%+v\n", search, p, err)
				continue
			}
			if strings.Contains(environ, HarvestTag) {
				result = append(result, p)
			}
		}
	}
	return result, err
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
