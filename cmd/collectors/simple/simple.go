// Copyright NetApp Inc, 2021 All rights reserved

package simple

import (
	"github.com/netapp/harvest/v2/cmd/harvest/version"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type NodeMon struct {
	*collector.AbstractCollector
}

func init() {
	plugin.RegisterModule(&NodeMon{})
}

func (n *NodeMon) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.simple",
		New: func() plugin.Module { return new(NodeMon) },
	}
}

// Init initializes the collector
func (n *NodeMon) Init(a *collector.AbstractCollector) error {
	n.AbstractCollector = a
	var err error

	if err = collector.Init(n); err != nil {
		return err
	}

	// load list of counters from template
	if counters := n.Params.GetChildS("counters"); counters != nil {
		if err = n.loadMetrics(counters); err != nil {
			n.Logger.Error().Stack().Err(err).Msg("load metrics")
			return err
		}
	} else {
		return errs.New(errs.ErrMissingParam, "counters")
	}
	return nil
}

func (n *NodeMon) loadMetrics(counters *node.Node) error {
	var (
		metric matrix.Metric
		err    error
	)

	n.Logger.Debug().Msg("initializing metric cache")
	mat := n.Matrix[n.Object]
	// fetch list of counters from template
	for _, cnt := range counters.GetChildren() {

		name, display := parseMetricName(cnt.GetNameS())
		if cnt.GetNameS() == "" {
			name, display = parseMetricName(cnt.GetContentS())
		}
		dtype := "int"
		n.Logger.Trace().Msgf("handling (%s) (%s) dtype=%s", name, display, dtype)

		if metric, err = mat.NewMetricType(name, dtype); err != nil {
			return err
		}
		metric.SetName(display)
		n.Logger.Debug().Msgf("(%s) added metric (%s)", name, display)
	}

	if _, err = mat.NewMetricUint8("status"); err != nil {
		return err
	}

	n.Logger.Debug().Msgf("initialized cache with %d metrics", len(mat.GetMetrics()))
	return nil
}

func parseMetricName(name string) (string, string) {
	if fields := strings.Fields(name); len(fields) == 3 && fields[1] == "=>" {
		return fields[0], fields[2]
	}
	return name, name
}

// PollInstance - update instance cache with running pollers
func (n *NodeMon) PollInstance() (map[string]*matrix.Matrix, error) {
	mat := n.Matrix[n.Object]

	currInstances := set.NewFrom(mat.GetInstanceKeys())
	currSize := currInstances.Size()

	var err error
	name := "simple"
	if instance := mat.GetInstance(name); instance == nil {
		if instance, err = mat.NewInstance(name); err != nil {
			return nil, err
		}
		instance.SetLabel("poller", name)
		instance.SetLabel("version", version.VERSION)
		instance.SetLabel("pid", strconv.Itoa(os.Getpid()))
		n.Logger.Debug().Msgf("add instance (%s)", name)
	}
	t := len(mat.GetInstances())
	r := currInstances.Size()
	a := t - (currSize - r)
	n.Logger.Debug().Msgf("added %d, removed %d, total instances %d", a, r, t)

	return nil, nil
}

// PollData - update data cache
func (n *NodeMon) PollData() (map[string]*matrix.Matrix, error) {
	mat := n.Matrix[n.Object]
	mat.Reset()

	toQuery := []string{"alloc", "num_gc", "num_cpu"}

	for key, instance := range mat.GetInstances() {
		err := mat.LazySetValueUint64("status", key, 0)
		if err != nil {
			n.Logger.Error().Stack().Err(err).Msgf("error while parsing metric key [%s]", key)
		}
		for _, key2 := range toQuery {
			if metric := mat.GetMetric(key2); metric != nil {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				switch key2 {
				case "alloc":
					_ = metric.SetValueUint64(instance, m.Alloc)
				case "num_gc":
					_ = metric.SetValueUint64(instance, uint64(m.NumGC))
				case "num_cpu":
					_ = metric.SetValueInt64(instance, int64(runtime.NumCPU()))
				}
			}
		}
	}

	var count uint64
	n.AddCollectCount(count)
	n.Logger.Debug().Msgf("poll complete, added %d data points", count)
	return n.Matrix, nil
}

// Interface guards
var (
	_ collector.Collector = (*NodeMon)(nil)
)
