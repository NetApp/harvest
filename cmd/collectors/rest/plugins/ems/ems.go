/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package ems

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strings"
	"time"
)

const DefaultTimeInterval = 24 * time.Hour

type Ems struct {
	*plugin.AbstractPlugin
	timeInterval float64
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Ems{AbstractPlugin: p}
}

func (my *Ems) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	my.timeInterval = my.GetInterval(my.Params, DefaultTimeInterval)

	return nil
}

func (my *Ems) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		updateTimeMetric matrix.Metric
	)

	replaceStr := strings.NewReplacer("\"", "")
	for _, instance := range data.GetInstances() {
		instance.SetExportable(false)
		if updateTimeMetric = data.GetMetric("time"); updateTimeMetric == nil {
			my.Logger.Error().Stack().Msg("missing update time metric")
			return nil, errors.New(errors.MISSING_PARAM, "update_time")
		}

		if updateTime, ok := updateTimeMetric.GetValueFloat64(instance); ok {
			// convert expiryTime from float64 to int64 and find difference
			timestampDiff := time.Unix(int64(updateTime), 0).Sub(time.Now()).Hours()

			// timestampDiff will be more than 0 if it has reached this point, convert to days
			if timestampDiff <= my.timeInterval {
				instance.SetExportable(true)
				instance.SetLabel("log_message", replaceStr.Replace(instance.GetLabel("log_message")))
			}
		}

	}
	return nil, nil
}

func (my *Ems) GetInterval(param *node.Node, defaultInterval time.Duration) float64 {
	interval := param.GetChildS("interval")
	if interval != nil {
		timeInterval := interval.GetChildContentS("time")
		if timeInterval != "" {
			if durationVal, err := time.ParseDuration(timeInterval); err == nil {
				return durationVal.Hours()
			}
		}
	}
	return defaultInterval.Hours()
}
