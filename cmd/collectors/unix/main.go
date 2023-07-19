/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package unix

import (
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Relying on Wikipedia for the list of supporting platforms
// https://en.wikipedia.org/wiki/Procfs
var supportedPlatforms = []string{
	"aix",
	"android", // available in termux
	"dragonfly",
	"freebsd", // available, but not mounted by default
	"linux",
	"netbsd", // same as freebsd
	"plan9",
	"solaris",
}

var mountPoint = "/proc"

var clkTck float64

// list of histograms provided by the collector, mapped
// to function extracting them from a Process instance
var _Histograms = map[string]func(*matrix.Metric, string, *matrix.Instance, *Process){
	"cpu":    setCPU,
	"memory": setMemory,
	"io":     setIo,
	"net":    setNet,
	"ctx":    setCtx,
}

// list of (scalar) metrics
var _Metrics = map[string]func(*matrix.Metric, *matrix.Instance, *Process, *System){
	"start_time":     setStartTime,
	"cpu_percent":    setCPUPercent,
	"memory_percent": setMemoryPercent,
	"threads":        setNumThreads,
	"fds":            setNumFds,
}

var _DataTypes = map[string]string{
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

func init() {
	plugin.RegisterModule(&Unix{})
}

func (u *Unix) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.unix",
		New: func() plugin.Module { return new(Unix) },
	}
}

func getClockTicks() {
	clkTck = 100.0
	if data, err := exec.Command("getconf", "CLK_TCK").Output(); err == nil {
		if num, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err != nil {
			clkTck = float64(num)
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
		for key := range m {
			labels = append(labels, key)
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

// Init - initialize the collector
func (u *Unix) Init(a *collector.AbstractCollector) error {
	u.AbstractCollector = a
	var err error

	if !set.NewFrom(supportedPlatforms).Has(runtime.GOOS) {
		return errs.New(errs.ErrImplement, "platform not supported")
	}

	if err = collector.Init(u); err != nil {
		return err
	}

	// optionally let user define mount point of the fs
	if mp := u.Params.GetChildContentS("mount_point"); mp != "" {
		mountPoint = mp
	}

	// assert fs is available
	if fi, err := os.Stat(mountPoint); err != nil || !fi.IsDir() {
		return errs.New(errs.ErrImplement, "filesystem ["+mountPoint+"] not available")
	}

	// load list of counters from template
	if counters := u.Params.GetChildS("counters"); counters != nil {
		if err = u.loadMetrics(counters); err != nil {
			u.Logger.Error().Stack().Err(err).Msg("load metrics")
			return err
		}
	} else {
		return errs.New(errs.ErrMissingParam, "counters")
	}

	getClockTicks()
	if u.system, err = NewSystem(); err != nil {
		u.Logger.Error().Stack().Err(err).Msg("load system")
		return err
	}

	u.Matrix[u.Object].SetGlobalLabel("hostname", u.Options.Hostname)
	u.Matrix[u.Object].SetGlobalLabel("datacenter", u.Params.GetChildContentS("datacenter"))

	u.Logger.Debug().Msg("initialized")
	return nil
}

func (u *Unix) loadMetrics(counters *node.Node) error {
	var (
		proc           *Process
		metric         *matrix.Metric
		labels, wanted *set.Set
		err            error
	)

	u.Logger.Debug().Msg("initializing metric cache")
	mat := u.Matrix[u.Object]

	u.processes = make(map[string]*Process)
	u.histogramLabels = make(map[string][]string)

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

		dtype := _DataTypes[name]

		u.Logger.Trace().Msgf("handling (%s) (%s) dtype=%s", name, display, dtype)

		// counter is scalar metric
		if _, has := _Metrics[name]; has {

			if _, err = mat.NewMetricType(name, dtype, display); err != nil {
				return err
			}
			u.Logger.Debug().Msgf("(%s) added metric (%s)", name, display)

			// counter is histogram
		} else if _, has := _Histograms[name]; has {

			labels = set.NewFrom(getHistogramLabels(proc, name))

			u.histogramLabels[name] = make([]string, 0)

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
					u.Logger.Warn().Msgf("invalid histogram metric [%s]", label)
					wanted.Remove(w)
					continue
				}

				if metric, err = mat.NewMetricType(name+"."+label, dtype, name); err != nil {
					return err
				}
				metric.SetLabel("metric", ldisplay)
				u.histogramLabels[name] = append(u.histogramLabels[name], label)
			}

			u.Logger.Debug().Msgf("(%s) added histogram (%s) with %d submetrics", name, display, len(u.histogramLabels[name]))

			// invalid counter
		} else {
			u.Logger.Warn().Msgf("(%s) skipped unknown metric", name)
		}
	}

	if _, err = mat.NewMetricUint8("status"); err != nil {
		return err
	}

	u.Logger.Debug().Msgf("initialized cache with %d metrics", len(mat.GetMetrics()))
	return nil
}

// PollInstance - update instance cache with running pollers
func (u *Unix) PollInstance() (map[string]*matrix.Matrix, error) {

	mat := u.Matrix[u.Object]
	currInstances := set.NewFrom(mat.GetInstanceKeys())
	currSize := currInstances.Size()

	err := conf.LoadHarvestConfig(u.Options.Config)
	if err != nil {
		return nil, err
	}

	statuses, err := util.GetPollerStatuses()
	if err != nil {
		return nil, err
	}

	for _, name := range conf.Config.PollersOrdered {
		pid := -1
		for _, pollerStatus := range statuses {
			if pollerStatus.Name == name {
				pid = int(pollerStatus.Pid)
				break
			}
		}
		if pid == -1 {
			continue
		}
		if instance := mat.GetInstance(name); instance == nil {
			if instance, err = mat.NewInstance(name); err != nil {
				return nil, err
			}
			instance.SetLabel("poller", name)
			instance.SetLabel("pid", strconv.Itoa(pid))
			u.Logger.Debug().Str("name", name).Int("pid", pid).Msg("Add instance")
		} else {
			currInstances.Remove(name)
			instance.SetLabel("pid", strconv.Itoa(pid))
			u.Logger.Debug().Str("name", name).Int("pid", pid).Msg("Update instance")
		}
	}
	rewriteIndexes := currInstances.Size() > 0
	for name := range currInstances.Iter() {
		mat.RemoveInstance(name)
		u.Logger.Debug().Msgf("remove instance (%s)", name)
	}
	// If there were removals, the indexes need to be rewritten since gaps were created
	if rewriteIndexes {
		newMatrix := mat.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: true})
		for key := range mat.GetInstances() {
			_, _ = newMatrix.NewInstance(key)
		}
		mat = newMatrix
	}
	t := len(mat.GetInstances())
	r := currInstances.Size()
	a := t - (currSize - r)
	u.Logger.Debug().Msgf("added %d, removed %d, total instances %d", a, r, t)

	return nil, nil
}

// PollData - update data cache
func (u *Unix) PollData() (map[string]*matrix.Matrix, error) {

	var (
		pid   int
		count uint64
		err   error
		ok    bool
		proc  *Process
	)
	mat := u.Matrix[u.Object]
	mat.Reset()

	if err = u.system.Reload(); err != nil {
		return nil, err
	}

	for key, instance := range mat.GetInstances() {

		// assume not running
		err = mat.LazySetValueUint8("status", key, 0)
		if err != nil {
			u.Logger.Error().Stack().Err(err).Msgf("error while parsing metric key [%s]", key)
		}

		if proc, ok = u.processes[key]; ok {
			if err = proc.Reload(); err != nil {
				delete(u.processes, key)
				proc = nil
			}
		}

		if proc == nil {
			if instance.GetLabel("pid") == "" {
				u.Logger.Debug().Msgf("skip instance [%s]: not running", key)
				continue
			}
			if pid, err = strconv.Atoi(instance.GetLabel("pid")); err != nil {
				u.Logger.Warn().Msgf("skip instance [%s], invalid PID: %v", key, err)
				continue
			}
			if proc, err = NewProcess(pid); err != nil {
				u.Logger.Warn().Msgf("skip instance [%s], process: %v", key, err)
				continue
			}
			u.processes[key] = proc
		}

		poller := instance.GetLabel("poller")
		cmd := proc.Cmdline()

		if !set.NewFrom(strings.Fields(cmd)).Has(poller) {
			u.Logger.Debug().Msgf("skip instance [%s]: PID (%d) not matched with [%s]", key, pid, cmd)
			continue
		}

		// if we got here poller is running
		err = mat.LazySetValueUint64("status", key, 1)
		if err != nil {
			u.Logger.Error().Stack().Err(err).Msgf("error while parsing metric key [%s]", key)
		}

		u.Logger.Debug().Msgf("populating instance [%s]: PID (%d) with [%s]\n", key, pid, cmd)

		// process scalar metrics
		for key, foo := range _Metrics {
			if metric := mat.GetMetric(key); metric != nil {
				foo(metric, instance, proc, u.system)
				count++
			}
		}

		// process histograms
		for key, foo := range _Histograms {
			if labels, ok := u.histogramLabels[key]; ok {
				for _, label := range labels {
					if metric := mat.GetMetric(key + "." + label); metric != nil {
						foo(metric, label, instance, proc)
						count++
					}
				}
			}
		}
	}

	u.AddCollectCount(count)
	u.Logger.Debug().Msgf("poll complete, added %d data points", count)
	return u.Matrix, nil
}

func setStartTime(m *matrix.Metric, i *matrix.Instance, p *Process, s *System) {
	err := m.SetValueFloat64(i, p.startTime+s.bootTime)
	if err != nil {
		logging.Get().Error().Stack().Err(err).Msg("error")
	}
}

func setNumThreads(m *matrix.Metric, i *matrix.Instance, p *Process, _ *System) {
	err := m.SetValueUint64(i, p.numThreads)
	if err != nil {
		logging.Get().Error().Stack().Err(err).Msg("error")
	}
}

func setNumFds(m *matrix.Metric, i *matrix.Instance, p *Process, _ *System) {
	err := m.SetValueUint64(i, p.numFds)
	if err != nil {
		logging.Get().Error().Stack().Err(err).Msg("error")
	}
}

func setMemoryPercent(m *matrix.Metric, i *matrix.Instance, p *Process, s *System) {
	err := m.SetValueFloat64(i, float64(p.mem["rss"])/float64(s.memTotal)*100)
	if err != nil {
		logging.Get().Error().Stack().Err(err).Msg("error")
	}
}

func setCPUPercent(m *matrix.Metric, i *matrix.Instance, p *Process, _ *System) {
	if p.elapsedTime != 0 {
		err := m.SetValueFloat64(i, p.cpuTotal/p.elapsedTime*100)
		if err != nil {
			logging.Get().Error().Stack().Err(err).Msg("error")
		}
	} else {
		err := m.SetValueFloat64(i, p.cpuTotal/(float64(time.Now().Unix())-p.startTime)*100)
		if err != nil {
			logging.Get().Error().Stack().Err(err).Msg("error")
		}
	}
}

func setCPU(m *matrix.Metric, l string, i *matrix.Instance, p *Process) {
	if value, ok := p.cpu[l]; ok {
		err := m.SetValueFloat64(i, value)
		if err != nil {
			logging.Get().Error().Stack().Err(err).Msg("error")
		}
	}
}

func setMemory(m *matrix.Metric, l string, i *matrix.Instance, p *Process) {
	if value, ok := p.mem[l]; ok {
		err := m.SetValueUint64(i, value)
		if err != nil {
			logging.Get().Error().Stack().Err(err).Msg("error")
		}
	}
}

func setIo(m *matrix.Metric, l string, i *matrix.Instance, p *Process) {
	if value, ok := p.io[l]; ok {
		err := m.SetValueUint64(i, value)
		if err != nil {
			logging.Get().Error().Stack().Err(err).Msg("error")
		}
	}
}

func setNet(m *matrix.Metric, l string, i *matrix.Instance, p *Process) {
	if value, ok := p.net[l]; ok {
		err := m.SetValueUint64(i, value)
		if err != nil {
			logging.Get().Error().Stack().Err(err).Msg("error")
		}
	}
}

func setCtx(m *matrix.Metric, l string, i *matrix.Instance, p *Process) {
	if value, ok := p.ctx[l]; ok {
		err := m.SetValueUint64(i, value)
		if err != nil {
			logging.Get().Error().Stack().Err(err).Msg("error")
		}
	}
}

// Interface guards
var (
	_ collector.Collector = (*Unix)(nil)
)
