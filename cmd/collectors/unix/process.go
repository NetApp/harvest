/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package main

import (
	"bytes"
	"fmt"
	"goharvest2/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// Process - identity and stats about resource usage of a process
// most values are simple counters; elapsedTime & cpuTotal are the exception: they are deltas
type Process struct {
	pid         int
	dirpath     string
	cmdline     string
	name        string
	state       string
	startTime   float64
	elapsedTime float64
	timestamp   float64
	numThreads  uint64
	numFds      uint64
	cpuTotal    float64
	cpu         map[string]float64
	mem         map[string]uint64
	io          map[string]uint64
	net         map[string]uint64
	ctx         map[string]uint64
}

// NewProcess - returns an initialzed instance of Process
// if no process with *pid* exists, returns PROCESS_NOT_FOUND
func NewProcess(pid int) (*Process, error) {
	me := &Process{pid: pid}
	me.cpu = make(map[string]float64)
	me.mem = make(map[string]uint64)
	me.io = make(map[string]uint64)
	me.net = make(map[string]uint64)
	me.ctx = make(map[string]uint64)
	return me, me.Reload()
}

// Name - common name of the process
func (me *Process) Name() string {
	return me.name
}

// Cmdline - command-line vector of the process
func (me *Process) Cmdline() string {
	return me.cmdline
}

// State - state of the process
func (me *Process) State() string {
	return me.state
}

// Reload - load or refresh stats
func (me *Process) Reload() error {

	var err error

	me.dirpath = path.Join("/proc", strconv.Itoa(me.pid)+"/")

	if s, err := os.Stat(me.dirpath); err != nil || !s.IsDir() {
		if err == nil {
			return errors.New(PROCESS_NOT_FOUND, fmt.Sprintf("%s is not dir", me.dirpath))
		} else {
			return errors.New(PROCESS_NOT_FOUND, err.Error())
		}
	}

	if err = me.loadCmdline(); err != nil {
		return err
	}

	if err = me.loadStatus(); err != nil {
		return err
	}

	if err = me.loadStat(); err != nil {
		return err
	}

	if err = me.loadSmaps(); err != nil {
		return err
	}

	if err = me.loadIo(); err != nil {
		return err
	}

	if err = me.loadNetDev(); err != nil {
		return err
	}

	if err = me.loadFdinfo(); err != nil {
		return err
	}

	ts := float64(time.Now().Unix())
	if me.timestamp != 0 {
		me.elapsedTime = ts - me.timestamp
	}
	me.timestamp = ts

	return nil
}

func (me *Process) loadCmdline() error {

	var (
		data []byte
		err  error
	)
	if data, err = ioutil.ReadFile(path.Join(me.dirpath, "cmdline")); err != nil {
		return errors.New(FILE_READ, err.Error())
	}
	me.cmdline = string(bytes.ReplaceAll(data, []byte("\x00"), []byte(" ")))
	return nil
}

func (me *Process) loadStatus() error {
	var (
		data             []byte
		err              error
		line, key, value string
		fields           []string
		num              uint64
	)

	if data, err = ioutil.ReadFile(path.Join(me.dirpath, "status")); err != nil {
		return errors.New(FILE_READ, "status: "+err.Error())
	}

	for _, line = range strings.Split(string(data), "\n") {
		if fields = strings.Split(line, ":"); len(fields) == 2 {

			key = strings.ToLower(fields[0])
			value = strings.TrimSpace(fields[1])

			switch key {
			case "name":
				me.name = value
			case "state":
				me.state = value
			case "threads":
				if num, err = strconv.ParseUint(value, 10, 32); err == nil {
					me.numThreads = num
				}
			case "voluntary_ctxt_switches":
				if num, err = strconv.ParseUint(value, 10, 64); err == nil {
					me.ctx["voluntary"] = num
				}
			case "nonvoluntary_ctxt_switches":
				if num, err = strconv.ParseUint(value, 10, 64); err == nil {
					me.ctx["involuntary"] = num
				}
			}
		}
	}
	return err
}

func (me *Process) loadStat() error {

	var (
		data          []byte
		err           error
		prevTotal     float64
		num           int64
		after, fields []string
	)

	if data, err = ioutil.ReadFile(path.Join(me.dirpath, "stat")); err != nil {
		return errors.New(FILE_READ, "stat: "+err.Error())
	}

	// store previous values to calculate deltas
	prevTotal = me.cpu["user"] + me.cpu["system"]

	after = strings.Split(string(data), ")") // anythingn after status field

	if fields = strings.Fields(after[len(after)-1]); len(fields) >= 40 {

		// utime
		if num, err = strconv.ParseInt(fields[11], 10, 64); err == nil {
			me.cpu["user"] = float64(num) / _CLK_TCK
		}

		// stime
		if num, err = strconv.ParseInt(fields[12], 10, 64); err == nil {
			me.cpu["system"] = float64(num) / _CLK_TCK
		}

		// delayacct_blkio_ticks
		if num, err = strconv.ParseInt(fields[39], 10, 64); err == nil {
			me.cpu["iowait"] = float64(num) / _CLK_TCK
		}

		// process start time (since system boot time)
		if num, err = strconv.ParseInt(fields[19], 10, 64); err == nil {
			me.startTime = float64(num) / _CLK_TCK
		}
	}

	me.cpuTotal = (me.cpu["user"] + me.cpu["system"]) - prevTotal

	return err
}

func (me *Process) loadSmaps() error {

	var (
		data      []byte
		err       error
		num       uint64
		line, key string
		fields    []string
	)

	if data, err = ioutil.ReadFile(path.Join(me.dirpath, "smaps")); err != nil {
		return errors.New(FILE_READ, "smaps: "+err.Error())
	}

	me.mem = make(map[string]uint64)

	for _, line = range strings.Split(string(data), "\n") {

		if fields = strings.Fields(line); len(fields) == 3 {
			if num, err = strconv.ParseUint(fields[1], 10, 64); err == nil {

				key = strings.ToLower(strings.TrimSuffix(strings.Split(fields[0], "_")[0], ":"))

				if key == "rss" || key == "swap" || key == "anonymous" || key == "shared" || key == "private" {
					me.mem[key] += num
				} else if key == "size" {
					me.mem["vms"] += num
				}

			}
		}
	}
	return nil
}

func (me *Process) loadIo() error {

	var (
		data   []byte
		err    error
		line   string
		values []string
		num    uint64
	)

	if data, err = ioutil.ReadFile(path.Join(me.dirpath, "io")); err != nil {
		return errors.New(FILE_READ, "io: "+err.Error())
	}

	for _, line = range strings.Split(string(data), "\n") {
		if values = strings.Split(line, ":"); len(values) == 2 {
			if num, err = strconv.ParseUint(strings.TrimSpace(values[1]), 10, 64); err == nil {
				me.io[values[0]] = num
			}
		}
	}
	return err
}

func (me *Process) loadNetDev() error {

	var (
		data                               []byte
		label, line, value                 string
		lines, labels, raw, legend, fields []string
		i                                  int
		num                                uint64
		err                                error
	)

	if data, err = ioutil.ReadFile(path.Join(me.dirpath, "net", "dev")); err != nil {
		return errors.New(FILE_READ, "net/dev: "+err.Error())
	}

	me.net = make(map[string]uint64)

	labels = make([]string, 0)

	// extract counter labels
	if lines = strings.Split(string(data), "\n"); len(lines) > 2 {
		if legend = strings.Split(lines[1], "|"); len(legend) > 2 {
			raw = strings.Fields(legend[len(legend)-1])

			for _, label = range raw {
				labels = append(labels, "rcv_"+strings.TrimSpace(label))
			}

			for _, label = range raw {
				labels = append(labels, "sent_"+strings.TrimSpace(label))
			}
		}

		// extract counter values
		for _, line = range lines[2:] {
			if fields = strings.Fields(line); len(fields) == len(labels)+1 {
				for i, value = range fields[1:] {
					if num, err = strconv.ParseUint(value, 10, 64); err == nil {
						me.net[labels[i]] += num
					}
				}
			}
		}
	}

	return nil

}

func (me *Process) loadFdinfo() error {
	files, err := ioutil.ReadDir(path.Join(me.dirpath, "fdinfo"))
	if err != nil {
		return errors.New(DIR_READ, "fdinfo: "+err.Error())
	}
	me.numFds = uint64(len(files))
	return nil
}
