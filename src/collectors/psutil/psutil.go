package main

import (
	"os"
	"path"
	"strings"
	"strconv"
	"errors"
	"io/ioutil"
	"goharvest2/share/logger"
	"github.com/shirou/gopsutil/process"
	"goharvest2/poller/struct/yaml"
	"goharvest2/poller/struct/matrix"
    "goharvest2/poller/collector"
)


var extractors = map[string]interface{}{
	"Times" 	     : cpu_times,
	"MemoryInfo"     : memory_info,
	"IOCounters" 	 : io_counters,
	"NetIOCounters"  : net_io_counters,
	"NumCtxSwitches" : ctx_switches,
}


type Psutil struct {
	*collector.AbstractCollector
}


func New(a *collector.AbstractCollector) collector.Collector {
    return &Psutil{AbstractCollector: a}
}

func (c *Psutil) Init() error {

    if err := collector.Init(c); err != nil {
        return err
	}

    if counters := c.Params.GetChild("counters"); counters == nil {
		return errors.New("Missing counters in template")
	} else {
		c.load_metrics(counters)
	}

	//c.Data = matrix.New(object, c.Class, "", c.Params.GetChild("export_options"))
	hostname, _ := os.Hostname()
	c.Data.SetGlobalLabel("hostname", hostname)
	c.Data.SetGlobalLabel("datacenter", c.Params.GetChildValue("datacenter"))

	logger.Info(c.Prefix, "Collector initialized")

	return nil

}

func (c *Psutil) PollData() (*matrix.Matrix, error) {

	m := c.Data
	m.InitData()

	for key, instance := range m.Instances {
		pid, _ := m.GetInstanceLabel(instance, "pid")
		poller, _ := m.GetInstanceLabel(instance, "poller")

		// assume not running
		c.Data.SetValueS("status", instance, float32(1))

		if pid == "" {
			logger.Debug(c.Prefix, "Skip instance [%s]: not running", key)
			continue
		}

		pid_i, err := strconv.Atoi(pid)
		if err != nil {
			logger.Warn(c.Prefix, "Skip instance [%s], failed convert PID: %v", key, err)
			continue
		}

		proc, err := process.NewProcess(int32(pid_i))
		if err != nil {
			logger.Debug(c.Prefix, "Skip instance [%s], proc not found: %v", key, err)
			continue
		}

		name, _ := proc.Name()
		cmdline, _ := proc.Cmdline()

		logger.Debug(c.Prefix, "Extracting instance [%s] counters (%s) [%s]\n", key, name, cmdline)

		if !strings.Contains(name, "poller") || !strings.Contains(cmdline, poller) {
			logger.Debug(c.Prefix, "Skip instance [%s]: PID might have changed")
			continue
		}

		// if we got here poller is running
		c.Data.SetValueS("status", instance, float32(0))


		/*
		state, err := proc.Status()
		if err == nil {
			m.SetInstanceLabel(instance, "state", state)
		}*/

		cpu, _ := proc.CPUPercent()
		if err == nil {
			m.SetValueS("CPUPercent", instance, float32(cpu))
		}

		mem, _ := proc.MemoryPercent()
		if err == nil {
			m.SetValueS("MemoryPercent", instance, float32(mem))
		}

		create_time, _ := proc.CreateTime()
		if err == nil {
			m.SetValueS("CreateTime", instance, float32(create_time))
		}

		num_threads, _ := proc.NumThreads()
		if err == nil {
			m.SetValueS("NumThreads", instance, float32(num_threads))
		}

		num_fds, _ := proc.NumFDs()
		if err == nil {
			m.SetValueS("NumFDs", instance, float32(num_fds))
		}
		
		children, _ := proc.Children()
		if err == nil {
			m.SetValueS("NumChildren", instance, float32(len(children)))
		}
		
		socks, _ := proc.Connections()
		if err == nil {
			m.SetValueS("NumSockets", instance, float32(len(socks)))
		}

		for key, metric := range m.Metrics {

			if !metric.Scalar {
				f, ok := extractors[key]

				if !ok {
					continue
				}

				values, ok := f.(func(*process.Process)([]float32, bool))(proc)

				if !ok {
					continue
				}

				if len(values) != len(metric.Labels) {
					logger.Warn(c.Prefix, "Extracted [%s] values (%d) not what expected (%d)", metric.Display, len(values), len(metric.Labels))
					continue
				}

				m.SetArrayValues(metric, instance, values)
			}
		}
	}
	logger.Info(c.Prefix, "Data poll completed!")
	return m, nil
}

func (c *Psutil) load_metrics(counters *yaml.Node) {

	m := c.Data

	for _, child := range counters.Children {
		name, display := parse_metric_name(child.Name)

		logger.Debug(c.Prefix, "Parsing [%s] => (%s => %s)", child.Name, name, display)

		labels := make([]string, len(child.Values))
		for i, label := range(child.Values) {
			_, display := parse_metric_name(label)
			labels[i] = strings.ToLower(display)
		}

		logger.Debug(c.Prefix, "Parsed (%d) labels [%v] => (%d) [%v]", len(child.Values), child.Values, len(labels), labels)
		
		m.AddArrayMetric(name, display, labels, true)
		logger.Debug(c.Prefix, "+ Array metric [%s => %s] with %d labels", name, display, len(labels))
	}

	for _, value := range counters.Values {
		name, display := parse_metric_name(value)
		m.AddMetric(name, display, true)
		logger.Debug(c.Prefix, "+ Scalar metric [%s => %s]", name, display)
	}

	//m.AddMetric("status", "status", true) // static metric

	m.AddLabelName("poller")
	m.AddLabelName("pid")
	//m.AddLabelName("state")

	logger.Info(c.Prefix, "Loaded %d metrics", m.MetricsIndex)
}

func parse_metric_name(raw_name string) (string, string) {
	if items := strings.Split(raw_name, "=>"); len(items) == 2 {
		return strings.TrimSpace(items[0]), strings.TrimSpace(items[1])
	}
	return raw_name, raw_name
}

func (c *Psutil) PollInstance() (*matrix.Matrix, error) {

	c.Data.ResetInstances()

	poller_names, err := get_poller_names(c.Options.Path, c.Options.Config)
	if err != nil {
		return nil, err
	}

	for _, name := range poller_names {

		pid_s := ""

		pidfp := path.Join(c.Options.Path, "var", "." + name + ".pid")
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

			instance, _ := c.Data.AddInstance(name+"."+pid_s)

			c.Data.SetInstanceLabel(instance, "poller", name)
			c.Data.SetInstanceLabel(instance, "pid", pid_s)
		}

	}
	logger.Info(c.Prefix, "InstancePoll complete: added %d instances", len(c.Data.Instances))

	return nil, nil
}

func get_poller_names(harvest_path, config_fn string) ([]string, error){

	var poller_names []string

	config, err := yaml.Import(path.Join(harvest_path, config_fn))
	if err != nil {
		return poller_names, err
	} else if config == nil {
		return poller_names, errors.New("no content")
	}

	pollers := config.GetChild("Pollers")
	if pollers == nil {
		return poller_names, errors.New("no pollers")
	}

	for _, poller := range pollers.GetChildren() {
		poller_names = append(poller_names, poller.Name)
	}

	return poller_names, nil
}

func memory_info(proc *process.Process) ([]float32, bool) {

	values := make([]float32, 7)

	mem, err := proc.MemoryInfo()
	if err != nil {
		return values, false
	}

	values[0] = float32(mem.RSS)
	values[1] = float32(mem.VMS)
	values[2] = float32(mem.HWM)
	values[3] = float32(mem.Data)
	values[4] = float32(mem.Stack)
	values[5] = float32(mem.Locked)
	values[6] = float32(mem.Swap)

	return values, true
}

func cpu_times(proc *process.Process) ([]float32, bool) {

	values := make([]float32, 3)

	cpu, err := proc.Times()
	if err != nil {
		return values, false
	}

	values[0] = float32(cpu.User)
	values[1] = float32(cpu.System)
	values[2] = float32(cpu.Iowait)

	return values, true
}

func ctx_switches(proc *process.Process) ([]float32, bool) {

	values := make([]float32, 2)

	ctx, err := proc.NumCtxSwitches()
	if err != nil {
		return values, false
	}

	values[0] = float32(ctx.Voluntary)
	values[1] = float32(ctx.Involuntary)

	return values, true
}

func io_counters(proc *process.Process) ([]float32, bool) {

	values := make([]float32, 4)

	iocounter, err := proc.IOCounters()
	if err != nil {
		return values, false
	}

	values[0] = float32(iocounter.ReadCount)
	values[1] = float32(iocounter.WriteCount)
	values[2] = float32(iocounter.ReadBytes)
	values[3] = float32(iocounter.WriteBytes)

	return values, true
}

func net_io_counters(proc *process.Process) ([]float32, bool) {

	values := make([]float32, 8)

	netio, err := proc.NetIOCounters(false)
	if err != nil {
		return values, false
	}

	if len(netio) != 1 || netio[0].Name != "all" {
		return values, false
	}

	values[0] = float32(netio[0].BytesSent)
	values[1] = float32(netio[0].BytesRecv)
	values[2] = float32(netio[0].PacketsSent)
	values[3] = float32(netio[0].PacketsRecv)
	values[4] = float32(netio[0].Errin)
	values[5] = float32(netio[0].Errout)
	values[6] = float32(netio[0].Dropin)
	values[7] = float32(netio[0].Dropout)

	return values, true
}
