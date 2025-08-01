package fcp

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"log/slog"
	"math"
	"strconv"
	"strings"
)

type Fcp struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Fcp{AbstractPlugin: p}
}

func (f *Fcp) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {

	var rx, tx, utilPercent, read, write *matrix.Metric
	var err error
	data := dataMap[f.Object]

	if read = data.GetMetric("read_data"); read == nil {
		return nil, nil, errs.New(errs.ErrNoMetric, "read_data")
	}

	if write = data.GetMetric("write_data"); write == nil {
		return nil, nil, errs.New(errs.ErrNoMetric, "write_data")
	}

	if rx = data.GetMetric("read_percent"); rx == nil {
		if rx, err = data.NewMetricFloat64("read_percent"); err == nil {
			rx.SetProperty("raw")
		} else {
			return nil, nil, err
		}

	}
	if tx = data.GetMetric("write_percent"); tx == nil {
		if tx, err = data.NewMetricFloat64("write_percent"); err == nil {
			tx.SetProperty("raw")
		} else {
			return nil, nil, err
		}
	}

	if utilPercent = data.GetMetric("util_percent"); utilPercent == nil {
		if utilPercent, err = data.NewMetricFloat64("util_percent"); err == nil {
			utilPercent.SetProperty("raw")
		} else {
			return nil, nil, err
		}
	}

	for _, instance := range data.GetInstances() {

		instance.SetLabel("port", strings.TrimPrefix(instance.GetLabel("port"), "port."))

		var speed int
		var s string
		var err error

		if speed, err = strconv.Atoi(instance.GetLabel("speed")); err != nil {
			f.SLogger.Debug("skip, can't convert speed to numeric", slog.String("speed", s))
		}

		if speed != 0 {

			var rxBytes, txBytes, rxPercent, txPercent float64
			var rxOk, txOk bool

			if rxBytes, rxOk = write.GetValueFloat64(instance); rxOk {
				rxPercent = rxBytes / float64(speed)
				rx.SetValueFloat64(instance, rxPercent)
			}

			if txBytes, txOk = read.GetValueFloat64(instance); txOk {
				txPercent = txBytes / float64(speed)
				tx.SetValueFloat64(instance, txPercent)
			}

			if rxOk || txOk {
				utilPercent.SetValueFloat64(instance, math.Max(rxPercent, txPercent))
			}
		}
	}
	return nil, nil, nil
}
