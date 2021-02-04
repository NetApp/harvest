package main

import (
	"os"
	"path"
	"sync"
	"strings"
	"strconv"
	"errors"
	"io/ioutil"
	"github.com/shirou/gopsutil/process"
	"goharvest2/poller/yaml"
	"goharvest2/poller/structs/matrix"
	"goharvest2/poller/structs/options"
	"goharvest2/poller/share/logger"
    "goharvest2/poller/collector"
)

var Log *logger.Logger = logger.New(1, "")

var extractors = map[string]interface{}{
	"Times" 	     : cpu_times,
	"MemoryInfo"     : memory_info,
	"IOCounters" 	 : io_counters,
	"NetIOCounters"  : net_io_counters,
	"NumCtxSwitches" : ctx_switches,
}


type Psutil struct {
	*collector.AbstractCollector
	/*
	Class string
	Name string
	Params *yaml.Node
	Options *opts.Opts
	Exporters []exporter.Exporter
	Schedule *schedule.Schedule
	Data *matrix.Matrix*/
}

func New(name, obj string, options *options.Options, params *yaml.Node) collector.Collector {
	a := collector.New(name, obj, options, params)
	return &Psutil{AbstractCollector: a}
}

func (c *Psutil) Init() error {

    if err := c.InitAbc(); err != nil {
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

	Log.Info("Collector initialized")

	return nil

}

func (c *Psutil) Start(wg *sync.WaitGroup) {

	defer wg.Done()

	for {
		c.Metadata.InitData()

		for _, task := range c.Schedule.GetTasks() {

			if c.Schedule.IsDue(task) {
				c.Schedule.Start(task)

				data, err := c.poll(task)

				if err != nil {
					Log.Warn("%s poll failed: %v", task, err)
					return
				}

				Log.Debug("%s poll completed", task)

				duration := c.Schedule.Stop(task)
				c.Metadata.SetValueForMetricAndInstance("poll_time", task, duration.Seconds())

				if data != nil {

					Log.Debug("exporting to %d exporters", len(c.Exporters))

					for _, e := range c.Exporters {
						if err := e.Export(data); err != nil {
							Log.Warn("export to [%s] failed: %v", e.GetName(), err)
						}
					}
				}
			}

			Log.Debug("exporting metadata")

			for _, e := range c.Exporters {
				if err := e.Export(c.Metadata); err != nil {
					Log.Warn("metadata export to [%s] failed: %v", e.GetName(), err)
				}
			}
		}
		c.Schedule.Sleep()
	}
}

func (c *Psutil) poll(task string) (*matrix.Matrix, error) {
    switch task {
        case "data":
            return c.poll_data()
        case "instance":
            return nil, c.poll_instance()
        default:
            return nil, errors.New("invalid task: " + task)
    }
}


func (c *Psutil) poll_data() (*matrix.Matrix, error) {

	m := c.Data
	m.InitData()

	for key, instance := range m.Instances {
		pid, _ := m.GetInstanceLabel(instance, "pid")
		poller, _ := m.GetInstanceLabel(instance, "poller")

		// assume not running
		c.Data.SetValueForMetric("status", instance, float64(1))

		if pid == "" {
			Log.Debug("Skip instance [%s]: not running", key)
			continue
		}

		pid_i, err := strconv.Atoi(pid)
		if err != nil {
			Log.Warn("Skip instance [%s], failed convert PID: %v", key, err)
			continue
		}

		proc, err := process.NewProcess(int32(pid_i))
		if err != nil {
			Log.Debug("Skip instance [%s], proc not found: %v", key, err)
			continue
		}

		name, _ := proc.Name()
		cmdline, _ := proc.Cmdline()

		Log.Debug("Extracting instance [%s] counters (%s) [%s]\n", key, name, cmdline)

		if !strings.Contains(name, "poller") || !strings.Contains(cmdline, poller) {
			Log.Debug("Skip instance [%s]: PID might have changed")
			continue
		}

		// if we got here poller is running
		c.Data.SetValueForMetric("status", instance, float64(0))


		/*
		state, err := proc.Status()
		if err == nil {
			m.SetInstanceLabel(instance, "state", state)
		}*/

		cpu, _ := proc.CPUPercent()
		if err == nil {
			m.SetValueForMetric("CPUPercent", instance, float64(cpu))
		}

		mem, _ := proc.MemoryPercent()
		if err == nil {
			m.SetValueForMetric("MemoryPercent", instance, float64(mem))
		}
		
		create_time, _ := proc.CreateTime()
		if err == nil {
			m.SetValueForMetric("CreateTime", instance, float64(create_time))
		}
		
		num_threads, _ := proc.NumThreads()
		if err == nil {
			m.SetValueForMetric("NumThreads", instance, float64(num_threads))
		}
		
		num_fds, _ := proc.NumFDs()
		if err == nil {
			m.SetValueForMetric("NumFDs", instance, float64(num_fds))
		}
		
		children, _ := proc.Children()
		if err == nil {
			m.SetValueForMetric("NumChildren", instance, float64(len(children)))
		}
		
		socks, _ := proc.Connections()
		if err == nil {
			m.SetValueForMetric("NumSockets", instance, float64(len(socks)))
		}

		for key, metric := range m.Metrics {

			if !metric.Scalar {
				f, ok := extractors[key]

				if !ok {
					continue
				}

				values, ok := f.(func(*process.Process)([]float64, bool))(proc)

				if !ok {
					continue
				}

				if len(values) != len(metric.Labels) {
					Log.Warn("Extracted [%s] values (%d) not what expected (%d)", metric.Display, len(values), len(metric.Labels))
					continue
				}

				m.SetArrayValues(metric, instance, values)
			}
		}
	}
	Log.Info("Data poll completed!")
	return m, nil
}

func (c *Psutil) load_metrics(counters *yaml.Node) {

	m := c.Data

	for _, child := range counters.Children {
		name, display := parse_metric_name(child.Name)

		Log.Debug("Parsing [%s] => (%s => %s)", child.Name, name, display)

		labels := make([]string, len(child.Values))
		for i, label := range(child.Values) {
			_, display := parse_metric_name(label)
			labels[i] = strings.ToLower(display)
		}

		Log.Debug("Parsed (%d) labels [%v] => (%d) [%v]", len(child.Values), child.Values, len(labels), labels)
		
		m.AddArrayMetric(name, display, labels, true)
		Log.Debug("+ Array metric [%s => %s] with %d labels", name, display, len(labels))
	}

	for _, value := range counters.Values {
		name, display := parse_metric_name(value)
		m.AddMetric(name, display, true)
		Log.Debug("+ Scalar metric [%s => %s]", name, display)
	}

	//m.AddMetric("status", "status", true) // static metric

	m.AddLabelName("poller")
	m.AddLabelName("pid")
	//m.AddLabelName("state")

	Log.Info("Loaded %d metrics", m.MetricsIndex)
}

func parse_metric_name(raw_name string) (string, string) {
	if items := strings.Split(raw_name, "=>"); len(items) == 2 {
		return strings.TrimSpace(items[0]), strings.TrimSpace(items[1])
	}
	return raw_name, raw_name
}

func (c *Psutil) poll_instance() error {

	c.Data.ResetInstances()

	poller_names, err := get_poller_names(c.Options.Path, c.Options.Config)
	if err != nil {
		return err
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
			Log.Debug("Added pid [%s] from [%s]", pid_s, pidfp)
		} else {
			Log.Debug("No such pid file [%s]", pidfp)
		}

		if pid_s == "" {
			Log.Debug("Adding instance [%s] - not running", name)

			instance, _ := c.Data.AddInstance(name)

			c.Data.SetInstanceLabel(instance, "poller", name)
			c.Data.SetInstanceLabel(instance, "pid", "")
		} else {
			Log.Debug("Adding instance [%s] - up and running", name)

			instance, _ := c.Data.AddInstance(name+"."+pid_s)

			c.Data.SetInstanceLabel(instance, "poller", name)
			c.Data.SetInstanceLabel(instance, "pid", pid_s)
		}

	}
	Log.Info("InstancePoll complete: added %d instances", len(c.Data.Instances))

	return nil
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