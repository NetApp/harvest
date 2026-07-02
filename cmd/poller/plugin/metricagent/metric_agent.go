/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package metricagent

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"strconv"
	"strings"
)

type MetricAgent struct {
	*plugin.AbstractPlugin
	actions            []func(map[string]*matrix.Matrix) error
	computeMetricRules []computeMetricRule
}

func New(p *plugin.AbstractPlugin) *MetricAgent {
	return &MetricAgent{AbstractPlugin: p}
}

func (a *MetricAgent) Init(remote conf.Remote) error {

	var (
		err   error
		count int
	)

	if err := a.AbstractPlugin.Init(remote); err != nil {
		return err
	}

	if count = a.parseRules(); count == 0 {
		err = errs.New(errs.ErrMissingParam, "valid rules")
	} else {
		a.SLogger.Debug("parsed rules", slog.Int("count", count), slog.Int("actions", len(a.actions)))
	}

	return err
}

func (a *MetricAgent) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {

	var err error

	for _, foo := range a.actions {
		_ = foo(dataMap)
	}

	return nil, nil, err
}

func (a *MetricAgent) computeMetrics(dataMap map[string]*matrix.Matrix) error {

	var (
		metric                    *matrix.Metric
		metricVal, firstMetricVal *matrix.Metric
		m                         []*matrix.Matrix
		err                       error
		metricNotFound            []error
	)

	data := dataMap[a.Object]

	// map values for compute_metric mapping rules
	for _, r := range a.computeMetricRules {
		m = make([]*matrix.Matrix, len(r.metricNames))
		for i := range len(r.metricNames) {
			if data == nil {
				m[i] = dataMap[r.metricNames[i]]
			} else {
				m[i] = data
			}
		}
		if metric = a.getMetric(m[0], r.metric); metric == nil {
			if metric, err = m[0].NewMetricFloat64(r.metric); err != nil {
				a.SLogger.Error("Failed to create metric", slogx.Err(err), slog.String("metric", r.metric))
				return err
			}
			metric.SetProperty("compute_metric mapping")
		}

		for iKey, instance := range m[0].GetInstances() {
			var result float64

			// Parse first operand and store in result for further processing
			if firstMetricVal = a.getMetric(m[0], r.metricNames[0]); firstMetricVal != nil {
				if val, ok := firstMetricVal.GetValueFloat64(instance); ok {
					result = val
				} else {
					continue
				}
			} else {
				a.SLogger.Warn("computeMetrics: metric not found", slogx.Err(err), slog.String("metricName", r.metricNames[0]))
			}

			// Parse other operands and process them
			for i := 1; i < len(r.metricNames); i++ {
				var v float64
				if value, err := strconv.Atoi(r.metricNames[i]); err == nil {
					v = float64(value)
				} else {
					if m[i] != nil && m[i].GetInstance(iKey) != nil {
						metricVal = a.getMetric(m[i], r.metricNames[i])
						if metricVal != nil {
							v, _ = metricVal.GetValueFloat64(m[i].GetInstance(iKey))
						} else {
							metricNotFound = append(metricNotFound, err)
							break
						}
					} else {
						a.SLogger.Warn("computeMetrics: instance not found", slog.String("iKey", iKey))
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
						// don't divide by zero
						result = 0
					}
				case "PERCENT":
					if v != 0 {
						result = (result / v) * 100
					} else {
						// don't divide by zero
						result = 0
					}
				default:
					a.SLogger.Warn("Unknown operation", slog.String("operation", r.operation))
				}
			}

			metric.SetValueFloat64(instance, result)
		}
	}
	if len(metricNotFound) > 0 {
		a.SLogger.Warn("", slog.Any("computeMetrics: errors for metric not found", metricNotFound))
	}
	return nil
}

func (a *MetricAgent) getMetric(m *matrix.Matrix, name string) *matrix.Metric {
	metric := m.DisplayMetric(name)
	if metric != nil {
		return metric
	}
	return m.GetMetric(name)
}

// NewMetrics returns the new metrics the receiver creates
func (a *MetricAgent) NewMetrics() []plugin.DerivedMetric {
	derivedMetrics := make([]plugin.DerivedMetric, 0, len(a.computeMetricRules))
	for _, rule := range a.computeMetricRules {
		derivedMetrics = append(derivedMetrics, plugin.DerivedMetric{
			Name:   rule.metric,
			Source: strings.Join(rule.metricNames, ", "),
		})
	}
	return derivedMetrics
}
