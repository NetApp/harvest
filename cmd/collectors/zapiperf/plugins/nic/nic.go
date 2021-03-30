//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
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
    "goharvest2/cmd/poller/collector/plugin"
    "goharvest2/pkg/logger"
    "goharvest2/pkg/matrix"
    "math"
    "strconv"
    "strings"
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
            return nil, err
        }

    }
    if tx = data.GetMetric("tx_percent"); tx == nil {
        if tx, err = data.AddMetric("tx_percent", "tx_percent", true); err == nil {
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

    if nic_state = data.GetMetric("state"); nic_state == nil {
        if nic_state, err = data.AddMetric("state", "state", true); err == nil {
            nic_state.Properties = "raw"
        } else {
            return nil, err
        }
    }

    for _, instance := range data.GetInstances() {

        var speed, base int
        var s string
        var err error

        if s = instance.Labels.Get("speed"); strings.HasSuffix(s, "M") {
            base, err = strconv.Atoi(strings.TrimSuffix(s, "M"))
            if err != nil {
                logger.Debug(p.Prefix, "skip, can't convert speed (%s) to numeric", s)
            } else {
                speed = base * 125000
                instance.Labels.Set("speed", strconv.Itoa(speed))
                logger.Trace(p.Prefix, "converted speed (%s) to numeric (%d)", s, speed)
            }
        } else if speed, err = strconv.Atoi(s); err != nil {
            logger.Debug(p.Prefix, "skip, can't convert speed (%s) to numeric", s)
        }

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

        if state := instance.Labels.Get("state"); state == "up" {
            data.SetValue(nic_state, instance, float64(0))
        } else {
            data.SetValue(nic_state, instance, float64(1))
        }

        // truncate redundant prefix in nic type
        if t := instance.Labels.Get("type"); strings.HasPrefix(t, "nic_") {
            instance.Labels.Set("type", strings.TrimPrefix(t, "nic_"))
        }

    }

    return nil, nil
}
