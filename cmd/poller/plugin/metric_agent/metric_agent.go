/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package metric_agent

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"strconv"
)

type MetricAgent struct {
	*plugin.AbstractPlugin
	actions            []func(*matrix.Matrix) error
	computeMetricRules []computeMetricRule
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &MetricAgent{AbstractPlugin: p}
}

func (me *MetricAgent) Init() error {

	var (
		err   error
		count int
	)

	if err = me.AbstractPlugin.Init(); err != nil {
		return err
	}

	if count = me.parseRules(); count == 0 {
		err = errs.New(errs.ErrMissingParam, "valid rules")
	} else {
		me.Logger.Debug().Msgf("parsed %d rules for %d actions", count, len(me.actions))
	}

	return err
}

func (me *MetricAgent) Run(m *matrix.Matrix) ([]*matrix.Matrix, error) {

	var err error

	for _, foo := range me.actions {
		_ = foo(m)
	}

	return nil, err
}

func (me *MetricAgent) computeMetrics(m *matrix.Matrix) error {

	var (
		metric                    matrix.Metric
		metricVal, firstMetricVal matrix.Metric
		err                       error
	)

	// map values for compute_metric mapping rules
	for _, r := range me.computeMetricRules {

		if metric = m.GetMetric(r.metric); metric == nil {
			if metric, err = m.NewMetricFloat64(r.metric); err != nil {
				me.Logger.Error().Stack().Err(err).Str("new metric", r.metric).Msg("computeMetrics: failed to create metric")
				return err
			} else {
				metric.SetProperty("compute_metric mapping")
			}
		}

		for _, instance := range m.GetInstances() {
			var result float64

			// Parse first operand and store in result for further processing
			if firstMetricVal = m.GetMetric(r.metricNames[0]); firstMetricVal != nil {
				if val, ok := firstMetricVal.GetValueFloat64(instance); ok {
					result = val
				} else {
					continue
				}
			} else {
				me.Logger.Warn().Err(err).Str("metricName", r.metricNames[0]).Msg("computeMetrics: metric not found")
			}

			// Parse other operands and process them
			for i := 1; i < len(r.metricNames); i++ {
				var v float64
				if value, err := strconv.Atoi(r.metricNames[i]); err == nil {
					v = float64(value)
				} else {
					metricVal = m.GetMetric(r.metricNames[i])
					if metricVal != nil {
						v, _ = metricVal.GetValueFloat64(instance)
					} else {
						me.Logger.Warn().Err(err).Str("metricName", r.metricNames[i]).Msg("computeMetrics: metric not found")
						return nil
					}
				}

				switch r.operation {
				case "ADD":
					result += v
				case "SUBTRACT":
					result -= v
				case "MULTIPLY":
					result *= v
				case "DIVIDE":
					if v != 0 {
						result /= v
					} else {
						me.Logger.Error().
							Str("operation", r.operation).
							Msg("Division by zero operation")
					}
				default:
					me.Logger.Warn().
						Str("operation", r.operation).
						Msg("Unknown operation")
				}

			}

			_ = metric.SetValueFloat64(instance, result)
			me.Logger.Trace().Str("metricName", r.metric).Float64("metricValue", result).Msg("computeMetrics: new metric created")
		}
	}
	return nil
}
