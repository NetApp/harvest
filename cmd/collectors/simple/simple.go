// Copyright NetApp Inc, 2021 All rights reserved

package simple

import (
	"github.com/netapp/harvest/v2/cmd/harvest/version"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
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

	if err := collector.Init(n); err != nil {
		return err
	}

	// load list of counters from template
	if counters := n.Params.GetChildS("counters"); counters != nil {
		if err = n.loadMetrics(counters); err != nil {
			n.Logger.Error("load metrics", slogx.Err(err))
			return err
		}
	} else {
		return errs.New(errs.ErrMissingParam, "counters")
	}
	return nil
}

func (n *NodeMon) loadMetrics(counters *node.Node) error {
	var (
		err error
	)

	n.Logger.Debug("initializing metric cache")
	mat := n.Matrix[n.Object]
	// fetch list of counters from template
	for _, cnt := range counters.GetChildren() {

		name, display := parseMetricName(cnt.GetNameS())
		if cnt.GetNameS() == "" {
			name, display = parseMetricName(cnt.GetContentS())
		}
		dtype := "int64"

		if _, err = mat.NewMetricType(name, dtype, display); err != nil {
			return err
		}
		n.Logger.Debug("added metric", slog.String("name", name), slog.String("display", display))
	}

	if _, err = mat.NewMetricUint8("status"); err != nil {
		return err
	}

	n.Logger.Debug("initialized metric cache", slog.Int("numMetrics", len(mat.GetMetrics())))
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

	var err error
	name := "simple"
	if instance := mat.GetInstance(name); instance == nil {
		if instance, err = mat.NewInstance(name); err != nil {
			return nil, err
		}
		instance.SetLabel("poller", name)
		instance.SetLabel("version", version.VERSION)
		instance.SetLabel("pid", strconv.Itoa(os.Getpid()))
	}

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
			n.Logger.Error(
				"initializing metric cache",
				slogx.Err(err),
				slog.String("key", key),
			)
		}
		for _, key2 := range toQuery {
			if metric := mat.GetMetric(key2); metric != nil {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				switch key2 {
				case "alloc":
					metric.SetValueUint64(instance, m.Alloc)
				case "num_gc":
					metric.SetValueUint64(instance, uint64(m.NumGC))
				case "num_cpu":
					metric.SetValueInt64(instance, int64(runtime.NumCPU()))
				}
			}
		}
	}

	var count uint64
	n.AddCollectCount(count)
	n.Logger.Debug("poll complete", slog.Uint64("count", count))
	return n.Matrix, nil
}

// Interface guards
var (
	_ collector.Collector = (*NodeMon)(nil)
)
