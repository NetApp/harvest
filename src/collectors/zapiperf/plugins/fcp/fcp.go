package main

/*  Some postprocessing on counter data "nic_common"
    Converts link_speed to numeric MBs
    Adds custom metrics:
        - "rc_percent":    receive data utilization percent
        - "tx_percent":    sent data utilization percent
        - "util_percent":  max utilization percent
        - "nic_state":     0 if port is up, 1 otherwise
*/

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
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

func (p *Fcp) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var rx, tx, util *matrix.Metric
	var err error

	if rx = data.GetMetric("read_percent"); rx == nil {
		if rx, err = data.AddMetric("read_percent", "read_percent", true); err == nil {
			rx.Properties = "raw"
		} else {
			return nil, err
		}

	}
	if tx = data.GetMetric("write_percent"); tx == nil {
		if tx, err = data.AddMetric("write_percent", "write_percent", true); err == nil {
			tx.Properties = "raw"
		} else {
			return nil, err
		}
	}

	if util = data.GetMetric("util_percent"); util == nil {
		if util, err = data.AddMetric("util_percent", "util_percent", true); err == nil {
			util.Properties = "raw"
		} else {
			return nil, err
		}
	}

	/*
		if nic_state = data.GetMetric("status"); nic_state == nil {
			if nic_state, err = data.AddMetric("status", "status", true); err == nil {
				nic_state.Properties = "raw"
			} else {
				return nil, err
			}
		}
	*/

	for _, instance := range data.GetInstances() {

		instance.Labels.Set("port", strings.TrimPrefix(instance.Labels.Get("port"), "port."))

		var speed int
		var s string
		var err error

		if speed, err = strconv.Atoi(instance.Labels.Get("speed")); err != nil {
			logger.Debug(p.Prefix, "skip, can't convert speed (%s) to numeric", s)
		}

		if speed != 0 {

			var rx_bytes, tx_bytes, rx_percent, tx_percent float64
			var ok bool

			if rx_bytes, ok = data.GetValueS("write_data", instance); ok {
				rx_percent = rx_bytes / float64(speed)
				data.SetValue(rx, instance, rx_percent)
			}

			if tx_bytes, ok = data.GetValueS("read_data", instance); ok {
				tx_percent = tx_bytes / float64(speed)
				data.SetValue(tx, instance, tx_percent)
			}

			if ok {
				data.SetValue(util, instance, math.Max(rx_percent, tx_percent))
			}
		}

		/*
			if state := instance.Labels.Get("state"); state == "up" {
				data.SetValue(nic_state, instance, float64(0))
			} else {
				data.SetValue(nic_state, instance, float64(1))
			}*/

	}

	return nil, nil
}
