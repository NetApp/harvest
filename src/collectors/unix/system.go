package main

import (
	"strings"
	"strconv"
	"path"
	"io/ioutil"
	"errors"
)

type System struct {
	mem_total float64
	//cpu_total float64
    boot_time float64
}

func NewSystem() (*System, error) {
	s := &System{}
	return s, s.Reload()
}

func (s *System) Reload() error {

	if err := s.load_stat(); err != nil {
		return err
	}

	if err := s.load_meminfo(); err != nil {
		return err
	}

	return nil
}

func (s *System) load_stat() error {

	var (
		//cpu float64
		num int64
		data []byte
		lines, fields []string
		err error
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
		cpu += float64(num) / CLK_TCK
	}

	// system time
	if num, err = strconv.ParseInt(fields[3], 10, 64); err == nil {
		cpu +=  float64(num) / CLK_TCK
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
					s.boot_time = float64(num)
					return nil
				} else {
					return err
				}
			}
		}
	}
	return errors.New("btime not found")
}

// Read total Memory size
func (s *System) load_meminfo() error {

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
				if num, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					s.mem_total = float64(num)
					return nil
				}
				break
			}
		}
	}
	
	return errors.New("field MemTotal not found")
}