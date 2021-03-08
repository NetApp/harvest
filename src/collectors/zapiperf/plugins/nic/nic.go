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
    "math"
    "strconv"
	"strings"   
	"goharvest2/poller/collector/plugin"
    "goharvest2/share/matrix"
    "goharvest2/share/logger"
)

type Nic struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Nic{AbstractPlugin: p}
}

func (p *Nic) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

    var rx, tx, util, nic_state *matrix.Metric
    var err error

    if rx = data.GetMetric("rx_percent"); rx == nil {
        if rx, err = data.AddMetric("rx_percent", "rx_percent", true); err == nil {
            rx.Properties = "raw"
        } else {
            return err
        }
        
    }
    if tx = data.GetMetric("tx_percent"); tx == nil {
        if tx, err = data.AddMetric("tx_percent", "tx_percent", true); err == nil {
            tx.Properties = "raw"
        } else {
            return err
        }
    }

    if util = data.GetMetric("util_percent"); util == nil {
        if util, err = data.AddMetric("util_percent", "util_percent", true); err == nil {
            util.Properties = "raw"
        } else {
            return err
        }
    }

    if nic_state = data.GetMetric("nice_state"); nic_state == nil {
        if nic_state, err = data.AddMetric("nice_state", "nice_state", true); err == nil {
            nic_state.Properties = "raw"
        } else {
            return err
        }   
    }

	for _, instance := range data.GetInstances() {

        if x := instance.Labels.Get("link_speed"); strings.HasSuffix(x, "M") {
            base, err := strconv.Atoi(strings.TrimSuffix(x, "M"))
            if err != nil {
                logger.Debug(p.Prefix, "skip, can't convert speed (%s) to numeric", x)
            } else {
                speed := base * 125000
                instance.Labels.Set("link_speed", strconv.Itoa(speed))
                logger.Trace(p.Prefix, "converted speed (%s) to numeric (%d)", x, speed)

                if speed != 0 {

                    var rx_bytes, tx_bytes, rx_percent, tx_percent float64
                    var ok bool

                    if rx_bytes, ok = data.GetValueS("rx_bytes", instance); ok {
                        rx_percent = rx_bytes / float64(speed)
                        data.SetValue(rx, instance, rx_percent)
                    }

                    if tx_bytes, ok = data.GetValueS("tx_bytes", instance); ok {
                        tx_percent = tx_bytes / float64(speed)
                        data.SetValue(tx, instance, tx_percent)
                    }

                    if ok {
                        data.SetValue(util, instance, math.Max(rx_percent, tx_percent))
                    }
                }
            }
        }

        if state := instance.Labels.Get("link_current_state"); state == "up" {
            data.SetValue(nic_state, instance, float64(0))
        } else {
            data.SetValue(nic_state, instance, float64(1))
        }
	}

	return nil, nil
}
