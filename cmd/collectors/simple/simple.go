// Copyright NetApp Inc, 2021 All rights reserved

package simple

import (
	"goharvest2/cmd/harvest/version"
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/set"
	"goharvest2/pkg/tree/node"
	"runtime"
	"strings"
)

type NodeMon struct {
	*collector.AbstractCollector
}

func init() {
	plugin.RegisterModule(NodeMon{})
}

func (NodeMon) HarvestModule() plugin.ModuleInfo {
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
		return errors.New(errors.MISSING_PARAM, "counters")
	}
	return nil
}

func (n *NodeMon) loadMetrics(counters *node.Node) error {
	var (
		metric matrix.Metric
		err    error
	)

	n.Logger.Debug().Msg("initializing metric cache")

	// fetch list of counters from template
	for _, cnt := range counters.GetChildren() {

		name, display := parseMetricName(cnt.GetNameS())
		if cnt.GetNameS() == "" {
			name, display = parseMetricName(cnt.GetContentS())
		}
		dtype := "int"
		n.Logger.Trace().Msgf("handling (%s) (%s) dtype=%s", name, display, dtype)

		if metric, err = n.Matrix.NewMetricType(name, dtype); err != nil {
			return err
		}
		metric.SetName(display)
		n.Logger.Debug().Msgf("(%s) added metric (%s)", name, display)
	}

	if _, err = n.Matrix.NewMetricUint8("status"); err != nil {
		return err
	}

	n.Logger.Debug().Msgf("initialized cache with %d metrics", len(n.Matrix.GetMetrics()))
	return nil
}

func parseMetricName(name string) (string, string) {
	if fields := strings.Fields(name); len(fields) == 3 && fields[1] == "=>" {
		return fields[0], fields[2]
	}
	return name, name
}

// PollInstance - update instance cache with running pollers
func (n *NodeMon) PollInstance() (*matrix.Matrix, error) {

	currInstances := set.NewFrom(n.Matrix.GetInstanceKeys())
	currSize := currInstances.Size()

	var err error
	name := "simple"
	if instance := n.Matrix.GetInstance(name); instance == nil {
		if instance, err = n.Matrix.NewInstance(name); err != nil {
			return nil, err
		}
		instance.SetLabel("poller", name)
		instance.SetLabel("version", version.VERSION)
		n.Logger.Debug().Msgf("add instance (%s)", name)
	}
	t := len(n.Matrix.GetInstances())
	r := currInstances.Size()
	a := t - (currSize - r)
	n.Logger.Debug().Msgf("added %d, removed %d, total instances %d", a, r, t)

	return nil, nil
}

// PollData - update data cache
func (n *NodeMon) PollData() (*matrix.Matrix, error) {
	n.Matrix.Reset()

	toQuery := []string{"alloc", "num_gc", "num_cpu"}

	for key, instance := range n.Matrix.GetInstances() {
		err := n.Matrix.LazySetValueUint32("status", key, 0)
		if err != nil {
			n.Logger.Error().Stack().Err(err).Msgf("error while parsing metric key [%s]", key)
		}
		for _, key2 := range toQuery {
			if metric := n.Matrix.GetMetric(key2); metric != nil {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				switch key2 {
				case "alloc":
					_ = metric.SetValueUint64(instance, m.Alloc)
				case "num_gc":
					_ = metric.SetValueUint32(instance, m.NumGC)
				case "num_cpu":
					_ = metric.SetValueInt(instance, runtime.NumCPU())
				}
				//logger.Trace(me.Prefix, "+ (%s) [%f]", key, value)
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
