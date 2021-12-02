/*
 * Copyright NetApp Inc, 2021 All rights reserved

Package Description:
    Some postprocessing on counter data "nic_common"
      Converts link_speed to numeric MBs
      Adds custom metrics:
          - "rc_percent":    receive data utilization percent
          - "tx_percent":    sent data utilization percent
          - "util_percent":  max utilization percent
		  - "nic_state":     0 if port is up, 1 otherwise

*/
package nic

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/errors"
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

func (me *Nic) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var read, write, rx, tx, util matrix.Metric
	var err error

	if read = data.GetMetric("rx_bytes"); read == nil {
		return nil, errors.New(errors.ERR_NO_METRIC, "rx_bytes")
	}

	if write = data.GetMetric("tx_bytes"); write == nil {
		return nil, errors.New(errors.ERR_NO_METRIC, "tx_bytes")
	}

	if rx = data.GetMetric("rx_percent"); rx == nil {
		if rx, err = data.NewMetricFloat64("rx_percent"); err == nil {
			rx.SetProperty("raw")
		} else {
			return nil, err
		}

	}
	if tx = data.GetMetric("tx_percent"); tx == nil {
		if tx, err = data.NewMetricFloat64("tx_percent"); err == nil {
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

		var speed, base int
		var s string
		var err error

		if s = instance.GetLabel("speed"); strings.HasSuffix(s, "M") {
			base, err = strconv.Atoi(strings.TrimSuffix(s, "M"))
			if err != nil {
				me.Logger.Warn().Msgf("convert speed [%s]", s)
			} else {
				// NIC speed value converted from Mbps to Bps(bytes per second)
				speed = base * 125000
				instance.SetLabel("speed", strconv.Itoa(speed))
				me.Logger.Debug().Msgf("converted speed (%s) to numeric (%d)", s, speed)
			}
		} else if speed, err = strconv.Atoi(s); err != nil {
			me.Logger.Warn().Msgf("convert speed [%s]", s)
		}

		if speed != 0 {

			var rx_bytes, tx_bytes, rx_percent, tx_percent float64
			var rx_ok, tx_ok bool

			if rx_bytes, rx_ok = read.GetValueFloat64(instance); rx_ok {
				rx_percent = rx_bytes / float64(speed)
				err := rx.SetValueFloat64(instance, rx_percent)
				if err != nil {
					me.Logger.Error().Stack().Err(err).Msg("error")
				}
			}

			if tx_bytes, tx_ok = write.GetValueFloat64(instance); tx_ok {
				tx_percent = tx_bytes / float64(speed)
				err := tx.SetValueFloat64(instance, tx_percent)
				if err != nil {
					me.Logger.Error().Stack().Err(err).Msg("error")
				}
			}

			if rx_ok || tx_ok {
				err := util.SetValueFloat64(instance, math.Max(rx_percent, tx_percent))
				if err != nil {
					me.Logger.Error().Stack().Err(err).Msg("error")
				}
			}
		}

		// truncate redundant prefix in nic type
		if t := instance.GetLabel("type"); strings.HasPrefix(t, "nic_") {
			instance.SetLabel("type", strings.TrimPrefix(t, "nic_"))
		}

	}

	return nil, nil
}
