package main

import (
	"os"
	"path"
	"errors"
	"runtime"
	"math/rand"
	"strconv"
	"runtime/pprof"
	"poller/share/logger"
	"poller/collector"
	"poller/exporter"
	"poller/structs/opts"
	"poller/yaml"
)

var Log *logger.Logger = logger.New(1, "")

func init_collector(args *opts.Opts, params *yaml.Node) (collector.Collector, error) {

	var col collector.Collector

	instances := collector.New(args.Collector, params, args)
	
	if len(instances) == 0 {
		return col, errors.New("no collectors")
	}

	for _, i := range instances {

		if i.GetClass() == args.Collector {

			if args.Object == "" {
				col = i
				break
			} else if i.GetName() == args.Object {
				col = i
				break
			}
		}
	}

	if col == nil {
		return col, errors.New("collector/object not found")
	}
	
	return col, col.Init()
}

func init_exporter(args *opts.Opts, params *yaml.Node) (exporter.Exporter, error) {
	var exp exporter.Exporter

	if args.Exporter == "" {
		return exp, errors.New("no exporter requested")
	}

	exp = exporter.New("Prometheus", params, args)
	if exp == nil {
		panic("nil exporter")
	}
	return exp, exp.Init()
}


func run_sessions(col collector.Collector, exp exporter.Exporter, iter int) error {
	
	var err error

	Log.Info("Running instance poll")
	if err = col.PollInstance(); err != nil {
		return err
	}

	for i:=0; i<iter; i+=1 {
		col.ScheduleStart()
		Log.Info("Running poll cycle %s", i+1)
		data, err := col.PollData()
		if err != nil {
			return err
		}
		//data.Print()
		if exp != nil {
			exp.Export(data)
		}
		d := col.SchedulePause()
		Log.Info("Slept %s seconds", d.String())
	}
	return nil
}

func main() {

	args, name, err := opts.GetOpts()
	if err != nil {
		Log.Error("args: %v", err)
		return
	}

	Log = logger.New(args.LogLevel, name)

	params, all_exporter_params, err := ReadConfig(args.Path, args.Config, name)
	if err != nil {
		Log.Error("read config: %v", err)
		return
	}

	c, err := init_collector(args, params)
	if err !=nil {
		Log.Error("initialize collector: %v", err)
		return
	} else {
		Log.Info("initialized collector [%s:%s]", c.GetClass(), c.GetName())
	}

	exporter_params := all_exporter_params.PopChild(args.Exporter)

	if exporter_params == nil {
		if args.Exporter != "" {
			Log.Error("exporter [%s] not found in config")
			return
		}
	}

	e, err := init_exporter(args, exporter_params)
	if err != nil {
		Log.Warn("no exporter initialized")
	} else {
		c.AddExporter(e)
		Log.Info("added exporter [%s]", e.GetName())
	}

	filename_base := c.GetClass() + "_" + c.GetName() + "_" + strconv.Itoa(rand.Intn(1000000))
	cpuFile, err := os.Create("tests/" + filename_base + ".cpu")
	if err != nil {
		Log.Error("create cpu file: %v", err)
		return
	}

	pprof.StartCPUProfile(cpuFile)
	
	err = run_sessions(c, e, 1)
	if err != nil {
		Log.Error("run collector: %v", err)
	} else {
		Log.Info("running collector complete")
	}

	pprof.StopCPUProfile()
	cpuFile.Close()
	runtime.GC()

	memFile, err := os.Create("tests/" + filename_base + ".mem")
	if err != nil {
		Log.Error("create mem file: %v", err)
	}

	//pprof.StartCPUProfile(memFile)

	err = run_sessions(c, e, 1)
	if err != nil {
		Log.Error("run collector: %v", err)
	} else {
		Log.Info("running collector complete")
	}

	err = pprof.WriteHeapProfile(memFile)
	if err != nil {
		Log.Error("mem profile: %v", err)
	}
	memFile.Close()

	Log.Info("Generated output: [%s.cpu]", filename_base)
	Log.Info("Generated output: [%s.mem]", filename_base)

}


func ReadConfig(harvest_path, config_fn, name string) (*yaml.Node, *yaml.Node, error) {
	var err error
	var config, pollers, p, exporters, defaults *yaml.Node

	config, err = yaml.Import(path.Join(harvest_path, config_fn))

	if err == nil {

		pollers = config.GetChild("Pollers")
		defaults = config.GetChild("Defaults")

		if pollers == nil {
			err = errors.New("No pollers defined")
		} else {
			p = pollers.GetChild(name)
			if p == nil {
				err = errors.New("Poller [" + name + "] not defined")
			} else if defaults != nil {
				p.Union(defaults, false)
			}
		}
	}

	if err == nil && p != nil {

		exporters = config.GetChild("Exporters")
		if exporters == nil {
			Log.Warn("No exporters defined in config [%s]", config)
		} else {
			requested := p.GetChild("exporters")
			redundant := make([]*yaml.Node, 0)
			if requested != nil {
				for _, e := range exporters.Children {
					if !requested.HasInValues(e.Name) {
						redundant = append(redundant, e)
					}
				}
				for _, e := range redundant {
					exporters.PopChild(e.Name)
				}
			}
		}
	}

	return p, exporters, err
}