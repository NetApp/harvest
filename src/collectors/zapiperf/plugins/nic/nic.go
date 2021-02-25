package main

// Some postprocessing on counter data "nic_common"
// Converts link_speed to numeric MBs
// Adds label "state" with "0" indicating nic is up
// Adds custom metrics:
//  - "rc_percent":    receive data utilization percent
//  - "tx_percent":    sent data utilization percent
//  - "util_percent":  max utilization percent

import (
    "math"
    "strconv"
	"strings"
	"goharvest2/poller/collector/plugin"
    "goharvest2/share/matrix"
)

type Nic struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Nic{AbstractPlugin: p}
}

func (p *Nic) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

    var rx, tx, util *matrix.Metric

    if rx = data.GetMetric("rx_percent"); rx == nil {
        rx, _ = data.AddMetric("rx_percent", "rx_percent", true)
        rx.Properties = "raw"
    }
    if tx = data.GetMetric("tx_percent"); tx == nil {
        tx, _ = data.AddMetric("tx_percent", "tx_percent", true)
        tx.Properties = "raw"
    }

    if util = data.GetMetric("util_percent"); util == nil {
        util, _ = data.AddMetric("util_percent", "util_percent", true)
        util.Properties = "raw"
    }

	for _, instance := range data.GetInstances() {

        if x := instance.Labels.Get("link_speed"); strings.HasSuffix(x, "M") {
            if speed, err := strconv.Atoi(x); err == nil {
                instance.Labels.Set("link_speed", strconv.Itoa(speed * 125000))

                if speed != 0 {

                    var rx_bytes, tx_bytes float64
                    var ok bool

                    if rx_bytes, ok = data.GetValueS("rx_bytes", instance); ok {
                        data.SetValue(rx, instance, rx_bytes / float64(speed))
                    }

                    if tx_bytes, ok = data.GetValueS("tx_bytes", instance); ok {
                        data.SetValue(tx, instance, tx_bytes / float64(speed))
                    }

                    if ok {
                        data.SetValue(util, instance, math.Max(rx_bytes, tx_bytes))
                    }
                }
            }
        }

        if state := instance.Labels.Get("link_current_state"); state == "up" {
            instance.Labels.Set("state", "0")
        } else {
            instance.Labels.Set("state", "1")
        }
	}

	return nil, nil
}
