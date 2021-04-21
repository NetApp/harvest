//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package main

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/matrix"
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

func (me *Fcp) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var rx, tx, util, read, write matrix.Metric
	var err error

	if read = data.GetMetric("read_data"); read == nil {
		return nil, errors.New(errors.ERR_NO_METRIC, "read_data")
	}

	if write = data.GetMetric("write_data"); write == nil {
		return nil, errors.New(errors.ERR_NO_METRIC, "write_data")
	}

	if rx = data.GetMetric("read_percent"); rx == nil {
		if rx, err = data.NewMetricFloat64("read_percent"); err == nil {
			rx.SetProperty("raw")
		} else {
			return nil, err
		}

	}
	if tx = data.GetMetric("write_percent"); tx == nil {
		if tx, err = data.NewMetricFloat64("write_percent"); err == nil {
			tx.SetProperty("raw")
		} else {
			return nil, err
		}
	}

	if util = data.GetMetric("util_percent"); util == nil {
		if util, err = data.NewMetricFloat64("util_percent"); err == nil {
			util.SetProperty("raw")
		} else {
			return nil, err
		}
	}

	for _, instance := range data.GetInstances() {

		instance.SetLabel("port", strings.TrimPrefix(instance.GetLabel("port"), "port."))

		var speed int
		var s string
		var err error

		if speed, err = strconv.Atoi(instance.GetLabel("speed")); err != nil {
			logger.Debug(me.Prefix, "skip, can't convert speed (%s) to numeric", s)
		}

		if speed != 0 {

			var rx_bytes, tx_bytes, rx_percent, tx_percent float64
			var rx_ok, tx_ok bool

			if rx_bytes, rx_ok = write.GetValueFloat64(instance); rx_ok {
				rx_percent = rx_bytes / float64(speed)
				rx.SetValueFloat64(instance, rx_percent)
			}

			if tx_bytes, tx_ok = read.GetValueFloat64(instance); tx_ok {
				tx_percent = tx_bytes / float64(speed)
				tx.SetValueFloat64(instance, tx_percent)
			}

			if rx_ok || tx_ok {
				util.SetValueFloat64(instance, math.Max(rx_percent, tx_percent))
			}
		}
	}
	return nil, nil
}
