package main

import (
	"strings"
	"strconv"
	"bytes"
	"os"
	"path"
	"errors"
	"time"
	"io/ioutil"
)

type Process struct {
	pid int
	cmdline string
	name string
	state string
	start_time float64
	elapsed_time float64
	timestamp float64
	num_threads float64
	num_fds float64
	cpu_total float64
	cpu map[string]float64
	mem map[string]float64
	io map[string]float64
	net map[string]float64
	ctx map[string]float64
}

func NewProcess(pid int) (*Process, error) {
	p := &Process{pid: pid}
	p.cpu = make(map[string]float64)
	p.mem = make(map[string]float64)
	p.io = make(map[string]float64)
	p.net = make(map[string]float64)
	p.ctx = make(map[string]float64)
	return p, p.Reload()
}

func (p *Process) Name() string {
	return p.name
}

func (p *Process) Cmdline() string {
	return p.cmdline
}

func (p *Process) State() string {
	return p.state
}

func (p *Process) Reload() error {
	return p.load_all(path.Join("/proc", strconv.Itoa(p.pid)+"/"))
}

func (p *Process) load_all(dp string) error {

	var err error

	if _, err = os.Stat(dp); err != nil {
		return errors.New("process path: " + err.Error())
	}

	if err = p.load_cmdline(dp); err != nil {
		return errors.New("cmdline: " + err.Error())
	}

	if err = p.load_status(dp); err != nil {
		return errors.New("status: " + err.Error())
	}

	if err = p.load_stat(dp); err != nil {
		return errors.New("stat: " + err.Error())
	}

	if err = p.load_smaps(dp); err != nil {
		return errors.New("smaps: " + err.Error())
	}

	if err = p.load_io(dp); err != nil {
		return errors.New("io: " + err.Error())
	}

	if err = p.load_net_dev(dp); err != nil {
		return errors.New("net/dev: " + err.Error())
	}

	if err = p.load_fdinfo(dp); err != nil {
		return errors.New("fdinfo: " + err.Error())
	}

	ts := float64(time.Now().Unix())
	if p.timestamp != 0 {
		p.elapsed_time = ts - p.timestamp
	}
	p.timestamp = ts

	return nil
}

func (p *Process) load_cmdline(dp string) error {
	data, err := ioutil.ReadFile(path.Join(dp, "cmdline"))
	if err == nil {
		p.cmdline = string(bytes.ReplaceAll(data, []byte("\x00"), []byte(" ")))
	}
	return err
}

func (p *Process) load_status(dp string) error {
	data, err := ioutil.ReadFile(path.Join(dp, "status"))
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if fields := strings.Split(line, ":"); len(fields) == 2 {

				key := strings.ToLower(fields[0])
				value := strings.TrimSpace(fields[1])

				switch key {
				case "name":
					p.name = value
				case "state":
					p.state = value
				case "threads":
					if num, err := strconv.ParseInt(value, 10, 64); err == nil {
						p.num_threads = float64(num)
					}					
				case "voluntary_ctxt_switches":
					if num, err := strconv.ParseInt(value, 10, 64); err == nil {
						p.ctx["voluntary"] = float64(num)
					}
				case "nonvoluntary_ctxt_switches":
					if num, err := strconv.ParseInt(value, 10, 64); err == nil {
						p.ctx["involuntary"] = float64(num)
					}
				}
			}
		}
	}
	return err
}

func (p *Process) load_stat(dp string) error {

	prev_total := p.cpu["user"] + p.cpu["system"]

	data, err := ioutil.ReadFile(path.Join(dp, "stat"))
	
	if err == nil {

		after := strings.Split(string(data), ")") // anythingn after status field
		
		if fields := strings.Fields(after[len(after)-1]); len(fields) >= 40 {

			// utime
			if num, err := strconv.ParseInt(fields[11], 10, 64); err == nil {
				p.cpu["user"] = float64(num) / CLK_TCK
			}

			// stime
			if num, err := strconv.ParseInt(fields[12], 10, 64); err == nil {
				p.cpu["system"] = float64(num) / CLK_TCK
			}

			// delayacct_blkio_ticks
			if num, err := strconv.ParseInt(fields[39], 10, 64); err == nil {
				p.cpu["iowait"] = float64(num) / CLK_TCK
			}

			// process start time (since system boot time)
			if num, err := strconv.ParseInt(fields[19], 10, 64); err == nil {
				p.start_time = float64(num) / CLK_TCK
			}
		}

		p.cpu_total = (p.cpu["user"] + p.cpu["system"]) - prev_total
	}
	
	return err
}

func (p *Process) load_smaps(dp string) error {

	p.mem = make(map[string]float64)

	data, err := ioutil.ReadFile(path.Join(dp, "smaps"))

	if err == nil {

		for _, line := range strings.Split(string(data), "\n") {

			if fields := strings.Fields(line); len(fields) == 3 {
				if num, err := strconv.ParseInt(fields[1], 10, 64); err == nil {

					key := strings.ToLower(strings.TrimSuffix(strings.Split(fields[0], "_")[0], ":"))

					if key == "rss" || key == "swap" || key == "anonymous" || key == "shared" || key == "private" {
						p.mem[key] += float64(num)
					} else if key == "size" {
						p.mem["vms"] += float64(num)
					}

				}
			}
		}
	}
	return err
}

func (p *Process) load_io(dp string) error {

	//p.io = make(map[string]float64)

	data, err := ioutil.ReadFile(path.Join(dp, "io"))
	
	if err == nil {

		for _, line := range strings.Split(string(data), "\n") {

			if values := strings.Split(line, ":"); len(values) == 2 {
				if num, err := strconv.ParseInt(strings.TrimSpace(values[1]), 10, 64); err == nil {
					p.io[values[0]] = float64(num)
				}
			}
		}

	}
	return err
}

func (p *Process) load_net_dev(dp string) error {

	var (
		data []byte
		labels, lines []string
		err error
	)


	if data, err = ioutil.ReadFile(path.Join(dp, "net", "dev")); err != nil {
		return err
	}

	lines = strings.Split(string(data), "\n")
	p.net = make(map[string]float64)
	labels = make([]string, 0)
	
	// extract counter labels
	if len(lines) > 2 {
		if legend := strings.Split(lines[1], "|"); len(legend) > 2 {
			raw_labels := strings.Fields(legend[len(legend)-1])

			for _, label := range raw_labels {
				labels = append(labels, "rcv_" + strings.TrimSpace(label))
			}

			for _, label := range raw_labels {
				labels = append(labels, "sent_" + strings.TrimSpace(label))
			}
		}

		// extract counter values
		for _, line := range lines[2:] {

			if fields := strings.Fields(line); len(fields) == len(labels) + 1 {
				for i, v := range fields[1:] {
					if num, err := strconv.ParseInt(v, 10, 64); err == nil {
						p.net[labels[i]] += float64(num)
					}
				}
			}			
		}
	}

	return nil

}

func (p *Process) load_fdinfo(dp string) error {
	files, err := ioutil.ReadDir(path.Join(dp, "fdinfo"))
	if err == nil {
		p.num_fds = float64(len(files))
	}
	return err
}