package psutil

import (
	"path"
	"sync"
	"strings"
	"strconv"
	"errors"
	"io/ioutil"
	"math"
	"fmt"
	"reflect"
	"encoding/json"
	"github.com/shirou/gopsutil/v3/process"
	"poller/yaml"
	"poller/exporter"
	"poller/schedule"
	"poller/structs/matrix"
	"poller/structs/opts"
	"poller/share/logger"
)

var Log *logger.Logger = logger.New(1, "")


type Psutil struct {
	Class string
	Name string
	Params *yaml.Node
	Options *opts.Opts
	Exporters []exporter.Exporter
	Schedule *schedule.Schedule
	Data *matrix.Matrix
}

func New(class string, params *yaml.Node, options *opts.Opts) []*Psutil {
	var subcollectors []*Psutil
	Log = logger.New(options.LogLevel, class)
	
	c := Psutil{
		Class : class,
		Name : "local_proc",
		Params : params,
		Options : options,
	}
	c.Exporters = make([]exporter.Exporter, 0)
	subcollectors = append(subcollectors, &c)
	return subcollectors
}

func (c *Psutil) GetName() string {
    return c.Name
}

func (c *Psutil) GetClass() string {
    return c.Class
}

func (c *Psutil) GetExporterNames() []string {
    var names []string
    e := c.Params.GetChild("exporters")
    if e != nil {
        names = e.Values
	}
	Log.Info("OK, my exporters are: %v", names)
    return names
}

func (c *Psutil) AddExporter(e exporter.Exporter) {
	Log.Info("Adding exporter [%s]", e.GetName())
    c.Exporters = append(c.Exporters, e)
}

func (c *Psutil) Init() error {

	var err error

	template, err := yaml.Import(path.Join(c.Options.Path, "var", strings.ToLower(c.Class), "default.yaml"))
	//template, err := ImportDefaultTemplate(class, options.Path)
	if err != nil {
		return err
	}

	fmt.Printf("imported Template:")
	template.PrintTree(0)

	c.Params.Union(template, false)

	fmt.Printf("merged Template:")
	c.Params.PrintTree(0)

    object := c.Params.GetChildValue("object")
    if object == "" {
        Log.Warn("Missing object in template")
    }
    counters := c.Params.GetChild("counters")
    if counters == nil {
        Log.Warn("Missing counters in template")
    }
    if object == "" || counters == nil {
        return errors.New("missing parameters")
    }

	c.Data = matrix.New(object, c.Class, "", c.Params.GetChild("export_options"))
	
	c.load_metrics(counters)

	interval := c.Params.GetChild("schedule").GetChild("data").Value
	i, _ := strconv.Atoi(interval)
	c.Schedule = schedule.New(i)

	Log.Info("Collector started, poll interval: %s s", interval)

	return nil

}

func (c *Psutil) Start(wg *sync.WaitGroup) {

	defer wg.Done()

	for {
		c.Schedule.Start()
		Log.Debug("Starting poll session")

		err := c.poll_instance()
		if err != nil {
			Log.Error("instance poll: %v", err)
			continue
		}
		Log.Info("Completed instance poll")

		data, err := c.poll_data()
		if err != nil {
			Log.Error("data poll: %v", err)
			continue
		}
		Log.Info("Completed data poll")

		for _, exp := range c.Exporters {
			err := exp.Export(data)
			if err != nil {
				Log.Error("export data to [%s]: %v", exp.GetName(), err)
			} else {
				Log.Debug("exported data to [%s]", exp.GetName())
			}
		}

		c.Schedule.Pause()
	}
}

func (c *Psutil) poll_data() (*matrix.Matrix, error) {

	m := c.Data
	m.InitData()

	for key, instance := range m.Instances {
		pid, _ := m.GetInstanceLabel(instance, "pid")
		if pid == "" {
			Log.Debug("Skipping instance [%s]: not running", key)
			continue
		}

		pid_i,_ := strconv.Atoi(pid)
		proc, err := process.NewProcess(int32(pid_i))

		if err != nil {
			Log.Debug("Skipping instance [%s]: proc not found", key)
			continue
		}

		for key, metric := range m.Metrics {

			values := make(map[string]string)
			var value float64
			
			switch key {
			case "CPUPercent":
				value, _ = proc.CPUPercent()
				break
			case "CreateTime":
				v, _ := proc.CreateTime()
				value = float64(v)
				break
			case "MemoryPercent":
				v, _ := proc.MemoryPercent()
				value = float64(v)
				break
			case "NumFds":
				v, _ := proc.NumFDs()
				value = float64(v)
				break
			case "NumThreads":
				v, _ := proc.NumThreads()
				value = float64(v)
				break
			case "IOCounters":
				data, _ := proc.IOCounters()
				str, _ := json.Marshal(data)
				json.Unmarshal(str, &values)
				break
			case "MemoryInfo":
				data, _ := proc.MemoryInfo()
				str, _ := json.Marshal(data)
				json.Unmarshal(str, &values)
				break
			case "NumCtxSwitches":
				data, _ := proc.NumCtxSwitches()
				str, _ := json.Marshal(data)
				json.Unmarshal(str, &values)
				break
			case "Times":
				data, _ := proc.Times()
				str, _ := json.Marshal(data)
				json.Unmarshal(str, &values)
				break
			}

			if metric.Scalar {
				//Log.Debug("Handling metric [%s] with value [%v]", )
				m.SetValue(metric, instance, float64(value))
				Log.Debug("+ [%s] [%s] => [%f]", key, metric.Display, value)
			} else {
				float_values := make([]float64, 0)
				for _, labelUpper := range metric.Labels {
					label := strings.ToLower(labelUpper)
					v, ok := values[label]
					if !ok {
						Log.Warn("For metric [%s] label [%s] not found", metric.Display, label)
					}
					f, err := to_float64(v)
					if err != nil {
						Log.Warn("For metric [%s] label [%s]: value [%s] failed to convert", metric.Display, label, v)
						float_values = append(float_values, math.NaN())
					} else {
						float_values = append(float_values, f)
					}
					Log.Debug("+ [%s] [%s:%s] => [%f]", key, metric.Display, label, f)
					
				}
				m.SetArrayValues(metric, instance, float_values)
			}
		}
	}

	Log.Info("Data poll completed!")
	return m, nil
}

func (c *Psutil) load_metrics(counters *yaml.Node) {

	m := c.Data

	for _, child := range counters.Children {
		name := child.Name
		labels := child.Values
		
		m.AddMetricArray(name, name, labels, true)

		Log.Debug("Added array metric [%s] with %d submetrics", name, len(labels))
	}

	for _, value := range counters.Values {
		m.AddMetric(value, value, true)
		Log.Debug("Added scalar metric [%s]", value)
	}

	Log.Info("Loaded %d metrics", m.MetricsIndex)
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
			c.Data.SetInstanceLabel(instance, "state", "1")
			c.Data.SetInstanceLabel(instance, "pid", "")
		} else {
			Log.Debug("Adding instance [%s] - up and running", name)

			instance, _ := c.Data.AddInstance(name+"."+pid_s)

			c.Data.SetInstanceLabel(instance, "poller", name)
			c.Data.SetInstanceLabel(instance, "state", "1")
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

func to_float64(x interface{}) (float64, error) {
	var floatType = reflect.TypeOf(float64(0))
	var stringType = reflect.TypeOf("")
	switch i := x.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		return strconv.ParseFloat(i, 64)
	default:
		v := reflect.ValueOf(x)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Float(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			return strconv.ParseFloat(s, 64)
		} else {
			return math.NaN(), fmt.Errorf("Can't convert %v to float64", v.Type())
		}
	}
}