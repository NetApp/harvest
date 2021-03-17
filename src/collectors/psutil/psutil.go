package main

import (
	"github.com/shirou/gopsutil/process"
	"goharvest2/poller/collector"
	"goharvest2/share/config"
	"goharvest2/share/dict"
	"goharvest2/share/errors"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"goharvest2/share/tree/node"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

var extractors = map[string]interface{}{
	"Times":          cpu_times,
	"MemoryInfo":     memory_info,
	"IOCounters":     io_counters,
	"NetIOCounters":  net_io_counters,
	"NumCtxSwitches": ctx_switches,
}

type Psutil struct {
	*collector.AbstractCollector
	array_labels map[string][]string
}

func New(a *collector.AbstractCollector) collector.Collector {
	return &Psutil{AbstractCollector: a}
}

func (c *Psutil) Init() error {

	if err := collector.Init(c); err != nil {
		return err
	}

	if counters := c.Params.GetChildS("counters"); counters != nil {
		c.array_labels = make(map[string][]string)
		c.load_metrics(counters)
	} else {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	//c.Data = matrix.New(object, c.Class, "", c.Params.GetChild("export_options"))
	c.Data.SetGlobalLabel("hostname", c.Options.Hostname)
	c.Data.SetGlobalLabel("datacenter", c.Params.GetChildContentS("datacenter"))

	logger.Debug(c.Prefix, "Collector initialized")

	return nil

}

func (c *Psutil) PollData() (*matrix.Matrix, error) {

	var count int

	count = 0

	if err := c.Data.InitData(); err != nil {
		return nil, err
	}

	for key, instance := range c.Data.GetInstances() {

		var pid int
		var err error
		var proc *process.Process

		// assume not running
		c.Data.SetValueS("status", instance, float64(1))

		if instance.Labels.Get("pid") == "" {
			logger.Debug(c.Prefix, "skip instance [%s]: not running", key)
			continue
		}

		if pid, err = strconv.Atoi(instance.Labels.Get("pid")); err != nil {
			logger.Warn(c.Prefix, "skip instance [%s], invalid PID: %v", key, err)
			continue
		}

		if proc, err = process.NewProcess(int32(pid)); err != nil {
			logger.Debug(c.Prefix, "skip instance [%s], proc not found: %v", key, err)
			continue
		}

		poller := instance.Labels.Get("poller")
		name, _ := proc.Name()
		cmdline, _ := proc.Cmdline()

		logger.Debug(c.Prefix, "parsing instance [%s] process (%s) [%s]\n", key, name, cmdline)

		if !strings.Contains(name, "poller") || !strings.Contains(cmdline, poller) {
			logger.Debug(c.Prefix, "skip instance [%s]: PID might have changed")
			continue
		}

		// if we got here poller is running
		c.Data.SetValueS("status", instance, float64(0))

		if cpu, err := proc.CPUPercent(); err == nil {
			c.Data.SetValueS("CPUPercent", instance, float64(cpu))
			count += 1
		}

		if mem, err := proc.MemoryPercent(); err == nil {
			c.Data.SetValueS("MemoryPercent", instance, float64(mem))
			count += 1
		}

		if create_time, err := proc.CreateTime(); err == nil {
			c.Data.SetValueS("CreateTime", instance, float64(create_time))
			count += 1
		}

		if num_threads, err := proc.NumThreads(); err == nil {
			c.Data.SetValueS("NumThreads", instance, float64(num_threads))
			count += 1
		}

		if num_fds, err := proc.NumFDs(); err == nil {
			c.Data.SetValueS("NumFDs", instance, float64(num_fds))
			count += 1
		}

		if children, err := proc.Children(); err == nil {
			c.Data.SetValueS("NumChildren", instance, float64(len(children)))
			count += 1
		}

		if socks, err := proc.Connections(); err == nil {
			c.Data.SetValueS("NumSockets", instance, float64(len(socks)))
			count += 1
		}

		for key, labels := range c.array_labels {

			if f, ok := extractors[key]; ok {

				if values, ok := f.(func(*process.Process) ([]float64, bool))(proc); ok {
					if len(values) != len(labels) {
						logger.Warn(c.Prefix, "metric [%s] labels don't match with values (%d, but expected %d)", key, len(values), len(labels))
						continue
					}

					for i, label := range labels {
						if m := c.Data.GetMetric(key + "." + label); m != nil {
							c.Data.SetValue(m, instance, values[i])
							count += 1
						} else {
							logger.Error(c.Prefix, "metric [%s.%s] not found in cache", key, label)
						}
					}
				}
			}
		}
	}
	c.AddCount(count)
	logger.Debug(c.Prefix, "Data poll completed. Added %d data points", count)
	return c.Data, nil
}

func (c *Psutil) load_metrics(counters *node.Node) {

	m := c.Data

	for _, child := range counters.GetChildren() {

		if name := child.GetContentS(); name != "" {
			key, display := parse_metric_name(name)
			if _, err := m.AddMetric(key, display, true); err == nil {
				logger.Debug(c.Prefix, "+ [%s] added metric (%s)", name, display)
			} else {
				panic(err)
			}
		} else if name := child.GetNameS(); name != "" {
			key, display := parse_metric_name(name)

			if labels := child.GetAllChildContentS(); len(labels) == 0 {
				logger.Warn(c.Prefix, "[%s] missing labels", key)
			} else {
				labels_clean := make([]string, 0, len(labels))
				for _, x := range labels {
					label, label_display := parse_metric_name(x)
					labels_clean = append(labels_clean, label)
					if metric, err := m.AddMetric(key+"."+label, display, true); err == nil {
						metric.Labels = dict.New()
						metric.Labels.Set("metric", label_display)
						logger.Debug(c.Prefix, "+ [%s] added metric (%s) with label (%s)", name, display, label)
					} else {
						panic(err)
					}
				}
				c.array_labels[key] = labels_clean
				logger.Debug(c.Prefix, "add key [%s] with labels [%v]", key, labels_clean)
			}
		} else {
			logger.Warn(c.Prefix, "skipping empty counter")
		}
	}

	//m.AddMetric("status", "status", true) // static metric

	m.AddLabel("poller", "")
	m.AddLabel("pid", "")
	//m.AddLabel("state")

	logger.Debug(c.Prefix, "Loaded %d metrics", m.SizeMetrics())
}

func parse_metric_name(raw_name string) (string, string) {
	if items := strings.Split(raw_name, "=>"); len(items) == 2 {
		return strings.TrimSpace(items[0]), strings.ToLower(strings.TrimSpace(items[1]))
	}
	return raw_name, raw_name
}

func (c *Psutil) PollInstance() (*matrix.Matrix, error) {

	c.Data.ResetInstances()

	poller_names, err := config.GetPollerNames(path.Join(c.Options.ConfPath, "harvest.yml"))
	if err != nil {
		return nil, err
	}

	for _, name := range poller_names {

		pid_s := ""

		pidfp := path.Join(c.Options.PidPath, name+".pid")
		pid_b, err := ioutil.ReadFile(pidfp)

		if err == nil {
			pid_s = string(pid_b)
			pid_i, err := strconv.ParseInt(pid_s, 10, 32)

			if err != nil {
				pid_s = ""
			} else if exists, _ := process.PidExists(int32(pid_i)); !exists {
				pid_s = ""
			}
			logger.Debug(c.Prefix, "Added pid [%s] from [%s]", pid_s, pidfp)
		} else {
			logger.Debug(c.Prefix, "No such pid file [%s]", pidfp)
		}

		if pid_s == "" {
			logger.Debug(c.Prefix, "Adding instance [%s] - not running", name)

			instance, _ := c.Data.AddInstance(name)

			c.Data.SetInstanceLabel(instance, "poller", name)
			c.Data.SetInstanceLabel(instance, "pid", "")
		} else {
			logger.Debug(c.Prefix, "Adding instance [%s] - up and running", name)

			instance, _ := c.Data.AddInstance(name + "." + pid_s)

			c.Data.SetInstanceLabel(instance, "poller", name)
			c.Data.SetInstanceLabel(instance, "pid", pid_s)
		}

	}
	logger.Debug(c.Prefix, "InstancePoll complete: added %d instances", len(c.Data.Instances))

	return nil, nil
}

func memory_info(proc *process.Process) ([]float64, bool) {

	values := make([]float64, 7)

	mem, err := proc.MemoryInfo()
	if err != nil {
		return values, false
	}

	values[0] = float64(mem.RSS)
	values[1] = float64(mem.VMS)
	values[2] = float64(mem.HWM)
	values[3] = float64(mem.Data)
	values[4] = float64(mem.Stack)
	values[5] = float64(mem.Locked)
	values[6] = float64(mem.Swap)

	return values, true
}

func cpu_times(proc *process.Process) ([]float64, bool) {

	values := make([]float64, 3)

	cpu, err := proc.Times()
	if err != nil {
		return values, false
	}

	values[0] = float64(cpu.User)
	values[1] = float64(cpu.System)
	values[2] = float64(cpu.Iowait)

	return values, true
}

func ctx_switches(proc *process.Process) ([]float64, bool) {

	values := make([]float64, 2)

	ctx, err := proc.NumCtxSwitches()
	if err != nil {
		return values, false
	}

	values[0] = float64(ctx.Voluntary)
	values[1] = float64(ctx.Involuntary)

	return values, true
}

func io_counters(proc *process.Process) ([]float64, bool) {

	values := make([]float64, 4)

	iocounter, err := proc.IOCounters()
	if err != nil {
		return values, false
	}

	values[0] = float64(iocounter.ReadCount)
	values[1] = float64(iocounter.WriteCount)
	values[2] = float64(iocounter.ReadBytes)
	values[3] = float64(iocounter.WriteBytes)

	return values, true
}

func net_io_counters(proc *process.Process) ([]float64, bool) {

	values := make([]float64, 8)

	netio, err := proc.NetIOCounters(false)
	if err != nil {
		return values, false
	}

	if len(netio) != 1 || netio[0].Name != "all" {
		return values, false
	}

	values[0] = float64(netio[0].BytesSent)
	values[1] = float64(netio[0].BytesRecv)
	values[2] = float64(netio[0].PacketsSent)
	values[3] = float64(netio[0].PacketsRecv)
	values[4] = float64(netio[0].Errin)
	values[5] = float64(netio[0].Errout)
	values[6] = float64(netio[0].Dropin)
	values[7] = float64(netio[0].Dropout)

	return values, true
}
