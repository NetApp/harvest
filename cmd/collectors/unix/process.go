/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package unix

import (
	"bytes"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/errs"
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

// NewProcess - returns an initialized instance of Process
// if no process with *pid* exists, returns ErrProcessNotFound
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
func (p *Process) Name() string {
	return p.name
}

// Cmdline - command-line vector of the process
func (p *Process) Cmdline() string {
	return p.cmdline
}

// State - state of the process
func (p *Process) State() string {
	return p.state
}

// Reload - load or refresh stats
func (p *Process) Reload() error {

	var err error

	p.dirpath = path.Join("/proc", strconv.Itoa(p.pid)+"/")

	if s, err := os.Stat(p.dirpath); err != nil || !s.IsDir() {
		if err == nil {
			return errs.New(ErrProcessNotFound, fmt.Sprintf("%s is not dir", p.dirpath))
		}
		return errs.New(ErrProcessNotFound, err.Error())
	}

	if err = p.loadCmdline(); err != nil {
		return err
	}

	if err = p.loadStatus(); err != nil {
		return err
	}

	if err = p.loadStat(); err != nil {
		return err
	}

	if err = p.loadSmaps(); err != nil {
		return err
	}

	if err = p.loadIo(); err != nil {
		return err
	}

	if err = p.loadNetDev(); err != nil {
		return err
	}

	if err = p.loadFdinfo(); err != nil {
		return err
	}

	ts := float64(time.Now().Unix())
	if p.timestamp != 0 {
		p.elapsedTime = ts - p.timestamp
	}
	p.timestamp = ts

	return nil
}

func (p *Process) loadCmdline() error {

	var (
		data []byte
		err  error
	)
	if data, err = os.ReadFile(path.Join(p.dirpath, "cmdline")); err != nil {
		return errs.New(ErrFileRead, err.Error())
	}
	p.cmdline = string(bytes.ReplaceAll(data, []byte("\x00"), []byte(" ")))
	return nil
}

func (p *Process) loadStatus() error {
	var (
		data             []byte
		err              error
		line, key, value string
		fields           []string
		num              uint64
	)

	if data, err = os.ReadFile(path.Join(p.dirpath, "status")); err != nil {
		return errs.New(ErrFileRead, "status: "+err.Error())
	}

	for _, line = range strings.Split(string(data), "\n") {
		if fields = strings.Split(line, ":"); len(fields) == 2 {

			key = strings.ToLower(fields[0])
			value = strings.TrimSpace(fields[1])

			switch key {
			case "name":
				p.name = value
			case "state":
				p.state = value
			case "threads":
				if num, err = strconv.ParseUint(value, 10, 32); err == nil {
					p.numThreads = num
				}
			case "voluntary_ctxt_switches":
				if num, err = strconv.ParseUint(value, 10, 64); err == nil {
					p.ctx["voluntary"] = num
				}
			case "nonvoluntary_ctxt_switches":
				if num, err = strconv.ParseUint(value, 10, 64); err == nil {
					p.ctx["involuntary"] = num
				}
			}
		}
	}
	return err
}

func (p *Process) loadStat() error {

	var (
		data          []byte
		err           error
		prevTotal     float64
		num           int64
		after, fields []string
	)

	if data, err = os.ReadFile(path.Join(p.dirpath, "stat")); err != nil {
		return errs.New(ErrFileRead, "stat: "+err.Error())
	}

	// store previous values to calculate deltas
	prevTotal = p.cpu["user"] + p.cpu["system"]

	after = strings.Split(string(data), ")") // anythingn after status field

	if fields = strings.Fields(after[len(after)-1]); len(fields) >= 40 {

		// utime
		if num, err = strconv.ParseInt(fields[11], 10, 64); err == nil {
			p.cpu["user"] = float64(num) / clkTck
		}

		// stime
		if num, err = strconv.ParseInt(fields[12], 10, 64); err == nil {
			p.cpu["system"] = float64(num) / clkTck
		}

		// delayacct_blkio_ticks
		if num, err = strconv.ParseInt(fields[39], 10, 64); err == nil {
			p.cpu["iowait"] = float64(num) / clkTck
		}

		// process start time (since system boot time)
		if num, err = strconv.ParseInt(fields[19], 10, 64); err == nil {
			p.startTime = float64(num) / clkTck
		}
	}

	p.cpuTotal = (p.cpu["user"] + p.cpu["system"]) - prevTotal

	return err
}

func (p *Process) loadSmaps() error {

	var (
		data      []byte
		err       error
		num       uint64
		line, key string
		fields    []string
	)

	// this may fail see https://github.com/NetApp/harvest/issues/249
	// when it does, ignore so the other /proc checks are given a chance to run
	if data, err = os.ReadFile(path.Join(p.dirpath, "smaps")); err != nil {
		return nil //nolint:nilerr
	}

	p.mem = make(map[string]uint64)

	for _, line = range strings.Split(string(data), "\n") {

		if fields = strings.Fields(line); len(fields) == 3 {
			if num, err = strconv.ParseUint(fields[1], 10, 64); err == nil {

				key = strings.ToLower(strings.TrimSuffix(strings.Split(fields[0], "_")[0], ":"))

				if key == "rss" || key == "swap" || key == "anonymous" || key == "shared" || key == "private" {
					p.mem[key] += num
				} else if key == "size" {
					p.mem["vms"] += num
				}

			}
		}
	}
	return nil
}

func (p *Process) loadIo() error {

	var (
		data   []byte
		err    error
		line   string
		values []string
		num    uint64
	)

	// this may fail see https://github.com/NetApp/harvest/issues/249
	// when it does, ignore so the other /proc checks are given a chance to run
	if data, err = os.ReadFile(path.Join(p.dirpath, "io")); err != nil {
		return nil //nolint:nilerr
	}

	for _, line = range strings.Split(string(data), "\n") {
		if values = strings.Split(line, ":"); len(values) == 2 {
			if num, err = strconv.ParseUint(strings.TrimSpace(values[1]), 10, 64); err == nil {
				p.io[values[0]] = num
			}
		}
	}
	return nil
}

func (p *Process) loadNetDev() error {

	var (
		data                               []byte
		label, line, value                 string
		lines, labels, raw, legend, fields []string
		i                                  int
		num                                uint64
		err                                error
	)

	if data, err = os.ReadFile(path.Join(p.dirpath, "net", "dev")); err != nil {
		return errs.New(ErrFileRead, "net/dev: "+err.Error())
	}

	p.net = make(map[string]uint64)

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
						p.net[labels[i]] += num
					}
				}
			}
		}
	}

	return nil

}

func (p *Process) loadFdinfo() error {
	// this may fail see https://github.com/NetApp/harvest/issues/249
	// when it does, ignore so the other /proc checks are given a chance to run
	files, err := os.ReadDir(path.Join(p.dirpath, "fdinfo"))
	if err != nil {
		return nil //nolint:nilerr
	}
	p.numFds = uint64(len(files))
	return nil
}
