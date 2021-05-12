/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package main

import (
	"goharvest2/cmd/poller/collector"
	"goharvest2/pkg/config"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/set"
	"goharvest2/pkg/tree/node"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Relying on Wikipedia for the list of supporting platforms
// https://en.wikipedia.org/wiki/Procfs
var _SUPPORTED_PLATFORMS = []string{
	"aix",
	"andriod", // available in termux
	"dragonfly",
	"freebsd", // available, but not mounted by default
	"linux",
	"netbsd", // same as freebsd
	"plan9",
	"solaris",
}

var _MOUNT_POINT = "/proc"

var _CLK_TCK float64

// list of histograms provided by the collector, mapped
// to functions extracting them from a Process instance
var _HISTOGRAMS = map[string]func(matrix.Metric, string, *matrix.Instance, *Process){
	"cpu":    setCpu,
	"memory": setMemory,
	"io":     setIo,
	"net":    setNet,
	"ctx":    setCtx,
}

// list of (scalar) metrics
var _METRICS = map[string]func(matrix.Metric, *matrix.Instance, *Process, *System){
	"start_time":     setStartTime,
	"cpu_percent":    setCpuPercent,
	"memory_percent": setMemoryPercent,
	"threads":        setNumThreads,
	"fds":            setNumFds,
}

var _DTYPES = map[string]string{
	"cpu":            "float64",
	"memory":         "uint64",
	"io":             "uint64",
	"net":            "uint64",
	"ctx":            "uint64",
	"start_time":     "float64",
	"cpu_percent":    "float64",
	"memory_percent": "float64",
	"threads":        "uint64",
	"fds":            "uint64",
}

func getClockTicks() {
	_CLK_TCK = 100.0
	if data, err := exec.Command("getconf", "CLK_TCK").Output(); err == nil {
		if num, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err != nil {
			_CLK_TCK = float64(num)
		}
	}
}

func parseMetricName(name string) (string, string) {
	if fields := strings.Fields(name); len(fields) == 3 && fields[1] == "=>" {
		return fields[0], fields[2]
	}
	return name, name
}

func getHistogramLabels(p *Process, name string) []string {

	var labels []string
	var m map[string]uint64

	// dirty fast solution

	if name == "cpu" {
		for key := range p.cpu {
			labels = append(labels, key)
		}
	} else {

		switch name {
		case "memory":
			m = p.mem
		case "io":
			m = p.io
		case "net":
			m = p.net
		case "ctx":
			m = p.ctx
		}
		if m != nil {
			for key := range m {
				labels = append(labels, key)
			}
		}
	}
	return labels
}

// Unix - collector providing basic stats about harvest pollers
// @TODO - extend to monitor any user-defined process
type Unix struct {
	*collector.AbstractCollector
	system          *System
	histogramLabels map[string][]string
	processes       map[string]*Process
}

// New - create new, uninitialized collector
func New(a *collector.AbstractCollector) collector.Collector {
	return &Unix{AbstractCollector: a}
}

// Init - initialize the collector
func (me *Unix) Init() error {

	var err error

	if !set.NewFrom(_SUPPORTED_PLATFORMS).Has(runtime.GOOS) {
		return errors.New(errors.ERR_IMPLEMENT, "platform not supported")
	}

	if err = collector.Init(me); err != nil {
		return err
	}

	// optionally let user define mount point of the fs
	if mp := me.Params.GetChildContentS("mount_point"); mp != "" {
		_MOUNT_POINT = mp
	}

	// assert fs is avilable
	if fi, err := os.Stat(_MOUNT_POINT); err != nil || !fi.IsDir() {
		return errors.New(errors.ERR_IMPLEMENT, "filesystem ["+_MOUNT_POINT+"] not available")
	}

	// load list of counters from template
	if counters := me.Params.GetChildS("counters"); counters != nil {
		if err = me.loadMetrics(counters); err != nil {
			logger.Error(me.Prefix, "load metrics: %v", err)
			return err
		}
	} else {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	getClockTicks()
	if me.system, err = NewSystem(); err != nil {
		logger.Error(me.Prefix, "load system: %v", err)
		return err
	}

	me.Matrix.SetGlobalLabel("hostname", me.Options.Hostname)
	me.Matrix.SetGlobalLabel("datacenter", me.Params.GetChildContentS("datacenter"))

	logger.Debug(me.Prefix, "initialized")
	return nil
}

func (me *Unix) loadMetrics(counters *node.Node) error {
	var (
		proc           *Process
		metric         matrix.Metric
		labels, wanted *set.Set
		err            error
	)

	logger.Debug(me.Prefix, "initializing metric cache")

	me.processes = make(map[string]*Process)
	me.histogramLabels = make(map[string][]string)

	// process instance for self, we will use this
	// to get size/labels of histograms at runtime
	if proc, err = NewProcess(os.Getpid()); err != nil {
		return err
	}

	// fetch list of counters from template
	for _, cnt := range counters.GetChildren() {

		name, display := parseMetricName(cnt.GetNameS())
		if cnt.GetNameS() == "" {
			name, display = parseMetricName(cnt.GetContentS())
		}

		dtype := _DTYPES[name]

		logger.Trace(me.Prefix, "handling (%s) (%s) dtype=%s", name, display, dtype)

		// counter is scalar metric
		if _, has := _METRICS[name]; has {

			if metric, err = me.Matrix.NewMetricType(name, dtype); err != nil {
				return err
			}
			metric.SetName(display)
			logger.Debug(me.Prefix, "(%s) added metric (%s)", name, display)

			// counter is histogram
		} else if _, has := _HISTOGRAMS[name]; has {

			labels = set.NewFrom(getHistogramLabels(proc, name))

			me.histogramLabels[name] = make([]string, 0)

			// if template defines labels, only collect those
			// otherwise get everything available

			if len(cnt.GetChildren()) != 0 {
				wanted = set.NewFrom(cnt.GetAllChildContentS())
			} else {
				wanted = labels
			}

			// validate
			for w := range wanted.Iter() {
				// parse label name and display name
				label, ldisplay := parseMetricName(w)

				if !labels.Has(label) {
					logger.Warn(me.Prefix, "invalid histogram metric [%s]", label)
					wanted.Delete(w)
					continue
				}

				if metric, err = me.Matrix.NewMetricType(name+"."+label, dtype); err != nil {
					return err
				}
				metric.SetName(name)
				metric.SetLabel("metric", ldisplay)
				me.histogramLabels[name] = append(me.histogramLabels[name], label)
			}

			logger.Debug(me.Prefix, "(%s) added histogram (%s) with %d submetrics", name, display, len(me.histogramLabels[name]))

			// invalid counter
		} else {
			logger.Warn(me.Prefix, "(%s) skipped unknown metric", name)
		}
	}

	if _, err = me.Matrix.NewMetricUint8("status"); err != nil {
		return err
	}

	logger.Debug(me.Prefix, "initialized cache with %d metrics", len(me.Matrix.GetMetrics()))
	return nil
}

// PollInstance - update instance cache with running pollers
func (me *Unix) PollInstance() (*matrix.Matrix, error) {

	currInstances := set.NewFrom(me.Matrix.GetInstanceKeys())
	currSize := currInstances.Size()

	pollerNames, err := config.GetPollerNames(me.Options.Config)
	if err != nil {
		return nil, err
	}

	for _, name := range pollerNames {
		pidf := path.Join(me.Options.PidPath, name+".pid")

		pid := ""

		if x, err := ioutil.ReadFile(pidf); err == nil {
			//logger.Debug(me.Prefix, "skip instance (%s), err pidf: %v", name, err)
			pid = string(x)
		}

		if instance := me.Matrix.GetInstance(name); instance == nil {
			if instance, err = me.Matrix.NewInstance(name); err != nil {
				return nil, err
			}
			instance.SetLabel("poller", name)
			instance.SetLabel("pid", pid)
			logger.Debug(me.Prefix, "add instance (%s) with PID (%s)", name, pid)
		} else {
			currInstances.Delete(name)
			instance.SetLabel("pid", pid)
			logger.Debug(me.Prefix, "update instance (%s) with PID (%s)", name, pid)
		}
	}

	for name := range currInstances.Iter() {
		me.Matrix.RemoveInstance(name)
		logger.Debug(me.Prefix, "remove instance (%s)", name)
	}

	t := len(me.Matrix.GetInstances())
	r := currInstances.Size()
	a := t - (currSize - r)
	logger.Debug(me.Prefix, "added %d, removed %d, total instances %d", a, r, t)

	return nil, nil
}

// PollData - update data cache
func (me *Unix) PollData() (*matrix.Matrix, error) {

	var (
		pid   int
		count uint64
		err   error
		ok    bool
		proc  *Process
	)

	me.Matrix.Reset()

	if err = me.system.Reload(); err != nil {
		return nil, err
	}

	for key, instance := range me.Matrix.GetInstances() {

		// assume not running
		err = me.Matrix.LazySetValueUint8("status", key, 1)
		if err != nil {
			logger.Error(me.Prefix, "error while parsing metric key [%s], %v", key, err)
		}

		if proc, ok = me.processes[key]; ok {
			if err = proc.Reload(); err != nil {
				delete(me.processes, key)
				proc = nil
			}
		}

		if proc == nil {
			if instance.GetLabel("pid") == "" {
				logger.Debug(me.Prefix, "skip instance [%s]: not running", key)
				continue
			}
			if pid, err = strconv.Atoi(instance.GetLabel("pid")); err != nil {
				logger.Warn(me.Prefix, "skip instance [%s], invalid PID: %v", key, err)
				continue
			}
			if proc, err = NewProcess(pid); err != nil {
				logger.Warn(me.Prefix, "skip instance [%s], process: %v", key, err)
				continue
			}
			me.processes[key] = proc
		}

		poller := instance.GetLabel("poller")
		cmd := proc.Cmdline()

		if !set.NewFrom(strings.Fields(cmd)).Has(poller) {
			logger.Debug(me.Prefix, "skip instance [%s]: PID (%d) not matched with [%s]", key, pid, cmd)
			continue
		}

		// if we got here poller is running
		err = me.Matrix.LazySetValueUint32("status", key, 0)
		if err != nil {
			logger.Error(me.Prefix, "error while parsing metric key [%s], %v", key, err)
		}

		logger.Debug(me.Prefix, "populating instance [%s]: PID (%d) with [%s]\n", key, pid, cmd)

		// process scalar metrics
		for key, foo := range _METRICS {
			if metric := me.Matrix.GetMetric(key); metric != nil {
				foo(metric, instance, proc, me.system)
				//logger.Trace(me.Prefix, "+ (%s) [%f]", key, value)
				count++
			}
		}

		// process histograms
		for key, foo := range _HISTOGRAMS {
			if labels, ok := me.histogramLabels[key]; ok {
				//logger.Trace(me.Prefix, "+++ (%s) [%v]", key, values)
				for _, label := range labels {
					if metric := me.Matrix.GetMetric(key + "." + label); metric != nil {
						foo(metric, label, instance, proc)
						count++
					}
				}
			}
		}
	}

	me.AddCollectCount(count)
	logger.Debug(me.Prefix, "poll complete, added %d data points", count)
	return me.Matrix, nil
}

func setStartTime(m matrix.Metric, i *matrix.Instance, p *Process, s *System) {
	err := m.SetValueFloat64(i, p.startTime+s.bootTime)
	if err != nil {
		logger.Error("", "error %v", err)
	}
}

func setNumThreads(m matrix.Metric, i *matrix.Instance, p *Process, s *System) {
	err := m.SetValueUint64(i, p.numThreads)
	if err != nil {
		logger.Error("", "error %v", err)
	}
}

func setNumFds(m matrix.Metric, i *matrix.Instance, p *Process, s *System) {
	err := m.SetValueUint64(i, p.numFds)
	if err != nil {
		logger.Error("", "error %v", err)
	}
}

func setMemoryPercent(m matrix.Metric, i *matrix.Instance, p *Process, s *System) {
	err := m.SetValueFloat64(i, float64(p.mem["rss"])/float64(s.memTotal)*100)
	if err != nil {
		logger.Error("", "error %v", err)
	}
}

func setCpuPercent(m matrix.Metric, i *matrix.Instance, p *Process, s *System) {
	if p.elapsedTime != 0 {
		err := m.SetValueFloat64(i, p.cpuTotal/p.elapsedTime*100)
		if err != nil {
			logger.Error("", "error %v", err)
		}
	} else {
		err := m.SetValueFloat64(i, p.cpuTotal/(float64(time.Now().Unix())-p.startTime)*100)
		if err != nil {
			logger.Error("", "error %v", err)
		}
	}
}

func setCpu(m matrix.Metric, l string, i *matrix.Instance, p *Process) {
	if value, ok := p.cpu[l]; ok {
		err := m.SetValueFloat64(i, value)
		if err != nil {
			logger.Error("", "error %v", err)
		}
	}
}

func setMemory(m matrix.Metric, l string, i *matrix.Instance, p *Process) {
	if value, ok := p.mem[l]; ok {
		err := m.SetValueUint64(i, value)
		if err != nil {
			logger.Error("", "error %v", err)
		}
	}
}

func setIo(m matrix.Metric, l string, i *matrix.Instance, p *Process) {
	if value, ok := p.io[l]; ok {
		err := m.SetValueUint64(i, value)
		if err != nil {
			logger.Error("", "error %v", err)
		}
	}
}

func setNet(m matrix.Metric, l string, i *matrix.Instance, p *Process) {
	if value, ok := p.net[l]; ok {
		err := m.SetValueUint64(i, value)
		if err != nil {
			logger.Error("", "error %v", err)
		}
	}
}

func setCtx(m matrix.Metric, l string, i *matrix.Instance, p *Process) {
	if value, ok := p.ctx[l]; ok {
		err := m.SetValueUint64(i, value)
		if err != nil {
			logger.Error("", "error %v", err)
		}
	}
}

// Need to appease go build - see https://github.com/golang/go/issues/20312
func main() {}
