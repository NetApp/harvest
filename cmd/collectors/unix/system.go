/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package main

import (
	"goharvest2/pkg/errors"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

// System - provides memory size and boot time of the system
type System struct {
	memTotal uint64
	//cpu_total float64
	bootTime float64
}

// NewSystem - creates an initialized instance of System
func NewSystem() (*System, error) {
	s := &System{}
	return s, s.Reload()
}

// Reload - refresh metrics
func (s *System) Reload() error {

	if err := s.loadStat(); err != nil {
		return err
	}

	return s.loadMeminfo()
}

// read values from /proc/stat - system boot time
func (s *System) loadStat() error {

	var (
		//cpu float64
		num           int64
		data          []byte
		lines, fields []string
		err           error
	)

	if data, err = ioutil.ReadFile(path.Join("/proc", "stat")); err != nil {
		return err
	}

	// first line should contain summary of CPU times, example:
	// cpu  8536493 101888 2291762 23315526 39855 674853 262635 0 0 0
	lines = strings.Split(string(data), "\n")

	/* Not used currently

	fields = strings.Fields(lines[0])
	if len(fields) < 5 || strings.ToLower(fields[0]) != "cpu" {
		return errors.New("cpu sum found")
	}

	// user time
	if num, err = strconv.ParseInt(fields[1], 10, 64); err == nil {
		cpu += float64(num) / _CLK_TCK
	}

	// system time
	if num, err = strconv.ParseInt(fields[3], 10, 64); err == nil {
		cpu +=  float64(num) / _CLK_TCK
	}

	// gives total cpu times since last poll
	// if this is the first time we poll, it's total cpu times since boot time
	s.cpu_total = cpu - s.cpu_total
	*/

	// extract system boot time, since file contains lines for each cpu
	// there is no way to know the exact line we need
	for _, line := range lines[1:] {
		if strings.HasPrefix(strings.ToLower(line), "btime") {
			if fields = strings.Fields(line); len(fields) == 2 {
				if num, err = strconv.ParseInt(fields[1], 10, 64); err == nil {
					s.bootTime = float64(num)
					return nil
				}
				return errors.New(FIELD_VALUE, "/proc/stat: btime ["+fields[1]+"]")
			}
		}
	}
	return errors.New(FIELD_NOT_FOUND, "/proc/stat: btime")
}

// read values from /proc/meminfo - system memory size
func (s *System) loadMeminfo() error {

	data, err := ioutil.ReadFile(path.Join("/proc", "meminfo"))
	if err != nil {
		return err
	}

	// First line contains total memory in Kb, example
	// MemTotal:        7707284 kB
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if fields := strings.Fields(line); len(fields) > 1 {
			if strings.HasPrefix(strings.ToLower(fields[0]), "memtotal") {
				if num, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					s.memTotal = num
					return nil
				}
				break
			}
		}
	}

	return errors.New(FIELD_NOT_FOUND, "/proc/meminfo: MemTotal")
}
