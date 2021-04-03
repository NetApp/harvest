package main

import (
	"os"
	"os/exec"
	"time"
	"runtime"
	"strings"
	"strconv"
	"path"
	"io/ioutil"
	"goharvest2/poller/collector"
	"goharvest2/share/matrix"
	"goharvest2/share/logger"
	"goharvest2/share/errors"
	"goharvest2/share/config"
	"goharvest2/share/set"
	"goharvest2/share/tree/node"
)

// Relying on Wikipedia for the list of supporting platforms
// https://en.wikipedia.org/wiki/Procfs
var SUPPORTED_PLATFORMS = []string{
	"aix",
	"andriod", // available in termux
	"dragonfly",
	"freebsd",  // available, but not mounted by default
	"linux",
	"netbsd",  // same as freebsd
	"plan9",
	"solaris",
}

var MOUNT_POINT = "/proc"

var CLK_TCK float64

// list of histograms provided by the collector, mapped
// to functions extracting them from a Process instance
var HISTOGRAMS = map[string]func(*Process) (map[string]float64){
	"cpu":			cpu,
	"memory":     	memory,
	"io":     		io,
	"net":  		net,
	"ctx": 			ctx,
}

// list of (scalar) metrics
var METRICS = map[string]func(*Process, *System) (float64){
	"start_time":		start_time,
	"cpu_percent":		cpu_prc,
	"memory_percent":	memory_prc,
	"threads":			num_threads,
	"fds":				num_fds,
}

func get_histogram_labels(p *Process, name string) []string {
	var labels []string
	if foo, ok := HISTOGRAMS[name]; ok {
		values := foo(p)
		for key := range values {
			labels = append(labels, key)
		}
	}
	return labels
}

type Unix struct {
	*collector.AbstractCollector
	system *System
	histogram_labels map[string][]string
	processes map[string]*Process
}

func New(a *collector.AbstractCollector) collector.Collector {
	return &Unix{AbstractCollector: a}
}

func (c *Unix) Init() error {

	var err error

	if ! set.NewFrom(SUPPORTED_PLATFORMS).Has(runtime.GOOS) {
		return errors.New(errors.ERR_IMPLEMENT, "platform not supported")
	}

	if err = collector.Init(c); err != nil {
		return err
	}

	// optionally let user define mount point of the fs
	if mp := c.Params.GetChildContentS("mount_point"); mp != "" {
		MOUNT_POINT = mp
	}

	// assert fs is avilable
	if fi, err := os.Stat(MOUNT_POINT); err != nil || ! fi.IsDir() {
		return errors.New(errors.ERR_IMPLEMENT, "filesystem [" + MOUNT_POINT + "] not available")
	}

	// load list of counters from template
	if counters := c.Params.GetChildS("counters"); counters != nil {
		if err = c.load_metrics(counters); err != nil {
			return err
		}
	} else {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	GetClockTicks()
	if c.system, err = NewSystem(); err != nil {
		return err
	}

	c.processes = make(map[string]*Process)

	c.Data.SetGlobalLabel("hostname", c.Options.Hostname)
	c.Data.SetGlobalLabel("datacenter", c.Params.GetChildContentS("datacenter"))

	logger.Debug(c.Prefix, "initialized")
	return nil
}

func (c *Unix) load_metrics(counters *node.Node) error {

	var (
		p *Process
		m matrix.Metric
		labels, wanted *set.Set
		err error
	)
	// process instance for self, we will use this
	// to get size/labels of histograms at runtime
	if p, err = NewProcess(os.Getpid()); err != nil {
		return err
	}

	c.histogram_labels = make(map[string][]string)

	// fetch list of counters from template
	for _, cnt := range counters.GetChildren() {

		name, display := parse_metric_name(cnt.GetNameS())
		if cnt.GetNameS() == "" {
			name, display = parse_metric_name(cnt.GetContentS())
		}

		logger.Trace(c.Prefix, "handling (%s) (%s)", name, display)

		// counter is scalar metric
		if _, has := METRICS[name]; has {

			if m, err = c.Data.AddMetricFloat64(name); err != nil {
				return err
			}
			m.SetName(display)
			logger.Debug(c.Prefix, "(%s) added metric (%s)", name, display)

		// counter is histogram
		} else if _, has := HISTOGRAMS[name]; has {

			labels = set.NewFrom(get_histogram_labels(p, name))

			c.histogram_labels[name] = make([]string, 0)

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
				label, label_display := parse_metric_name(w)

				if ! labels.Has(label) {
					logger.Warn(c.Prefix, "invalid histogram metric [%s]", label)
					wanted.Delete(w)
					continue
				}

				if m, err = c.Data.AddMetricFloat64(name+"."+label); err != nil {
					return err
				}
				m.SetName(name)
				m.SetLabel("metric", label_display)
				c.histogram_labels[name] = append(c.histogram_labels[name], label)				
			}

			logger.Debug(c.Prefix, "(%s) added histogram (%s) with %d submetrics", name, display, len(c.histogram_labels[name]))

		// invalid counter
		} else {
			logger.Warn(c.Prefix, "(%s) skipped unknown metric", name)
		}
	}

	//c.Data.AddLabel("poller", "")
	//c.Data.AddLabel("pid", "")

	if _, err = c.Data.AddMetricUint32("status"); err != nil {
		return err
	}

	logger.Debug(c.Prefix, "initialized cache with %d metrics", c.Data.SizeMetrics())
	return nil
}

func parse_metric_name(raw_name string) (string, string) {
	if fields := strings.Fields(raw_name); len(fields) == 3 && fields[1] == "=>" {
		return fields[0], fields[2]
	}
	return raw_name, raw_name
}

func (c *Unix) PollInstance() (*matrix.Matrix, error) {

	curr_instances := set.NewFrom(c.Data.GetInstanceKeys())
	curr_size := curr_instances.Size()

	poller_names, err := config.GetPollerNames(path.Join(c.Options.ConfPath, "harvest.yml"))
	if err != nil {
		return nil, err
	}

	for _, name := range poller_names {
		pidf := path.Join(c.Options.PidPath, name+".pid")

		pid := ""

		if x, err := ioutil.ReadFile(pidf); err == nil {
			//logger.Debug(c.Prefix, "skip instance (%s), err pidf: %v", name, err)
			pid = string(x)
		}

		if instance := c.Data.GetInstance(name); instance == nil {
			if instance, err = c.Data.AddInstance(name); err != nil {
				return nil, err
			}
			instance.SetLabel("poller", name)
			instance.SetLabel("pid", pid)
			logger.Debug(c.Prefix, "add instance (%s) with PID (%s)", name, pid)
		} else {
			curr_instances.Delete(name)
			instance.SetLabel("pid", pid)
			logger.Debug(c.Prefix, "update instance (%s) with PID (%s)", name, pid)
		}
	}

	for name := range curr_instances.Iter() {
		c.Data.RemoveInstance(name)
		logger.Debug(c.Prefix, "remove instance (%s)")
	}

	t := c.Data.SizeInstances()
	r := curr_instances.Size()
	a := t - (curr_size - r)
	logger.Debug(c.Prefix, "added %d, removed %d, total instances %d", a, r, t)

	return nil, nil
}

func (c *Unix) PollData() (*matrix.Matrix, error) {

	var (
		count, pid int
		err error
		ok bool
		proc *Process
	)

	if err = c.Data.Reset(); err != nil {
		return nil, err
	}

	if err = c.system.Reload(); err != nil {
		return nil, err
	}

	for key, instance := range c.Data.GetInstances() {

		// assume not running
		c.Data.LazySetValueUint32("status", key, 1)

		if proc, ok = c.processes[key]; ok {
			if err = proc.Reload(); err != nil {
				delete(c.processes, key)
				proc = nil
			}
		}

		if proc == nil {
			if instance.GetLabel("pid") == "" {
				logger.Debug(c.Prefix, "skip instance [%s]: not running", key)
				continue
			}
			if pid, err = strconv.Atoi(instance.GetLabel("pid")); err != nil {
				logger.Warn(c.Prefix, "skip instance [%s], invalid PID: %v", key, err)
				continue
			}
			if proc, err = NewProcess(pid); err != nil {
				logger.Warn(c.Prefix, "skip instance [%s], process: %v", key, err)
				continue
			}
			c.processes[key] = proc
		}

		poller := instance.GetLabel("poller")
		cmd := proc.Cmdline()
		
		if ! set.NewFrom(strings.Fields(cmd)).Has(poller) {
			logger.Debug(c.Prefix, "skip instance [%s]: PID (%d) not matched with [%s]", key, pid, cmd)
			continue
		}

		// if we got here poller is running
		c.Data.LazySetValueUint32("status", key, 0)

		logger.Debug(c.Prefix, "populating instance [%s]: PID (%d) with [%s]\n", key, pid, cmd)

		// process scalar metrics
		for key, foo := range METRICS {
			if metric := c.Data.GetMetric(key); metric != nil {
				value := foo(proc, c.system)
				logger.Trace(c.Prefix, "+ (%s) [%f]", key, value)
				metric.SetValueFloat64(instance, value)
				count++
			}
		}

		// process histograms
		for key, foo := range HISTOGRAMS {
			if labels, ok := c.histogram_labels[key]; ok {
				values := foo(proc)
				logger.Trace(c.Prefix, "+++ (%s) [%v]", key, values)
				for _, label := range labels {
					if metric := c.Data.GetMetric(key+"."+label); metric != nil {
						if value, ok := values[label]; ok {
							metric.SetValueFloat64(instance, value)
							count++
						}
					}
				}
			}
		}

	}

	c.AddCount(count)
	logger.Debug(c.Prefix, "poll complete, added %d data points", count)
	return c.Data, nil
}

func GetClockTicks() {
	CLK_TCK = 100.0
	if data, err := exec.Command("getconf", "CLK_TCK").Output(); err == nil {
		if num, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err != nil {
			CLK_TCK = float64(num)
		}
	}
}

func start_time(p *Process, s *System) float64 {
	return p.start_time + s.boot_time
}

func num_threads(p *Process, s *System) float64 {
	return p.num_threads
}

func num_fds(p *Process, s *System) float64 {
	return p.num_fds
}

func memory_prc(p *Process, s *System) float64 {
	return p.mem["rss"] / s.mem_total * 100
}

func cpu_prc(p *Process, s *System) float64 {
	if p.elapsed_time != 0 {
		return p.cpu_total / p.elapsed_time * 100
	}
	return p.cpu_total / (float64(time.Now().Unix()) - p.start_time) * 100
}

func cpu(p *Process) map[string]float64 {
	return p.cpu	
}

func memory(p *Process) map[string]float64 {
	return p.mem	
}

func io(p *Process) map[string]float64 {
	return p.io	
}

func net(p *Process) map[string]float64 {
	return p.net	
}

func ctx(p *Process) map[string]float64 {
	return p.ctx	
}