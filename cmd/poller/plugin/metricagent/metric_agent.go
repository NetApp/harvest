/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package metricagent

import (
	"errors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"maps"
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

	ee := make([]error, 0, len(a.actions))

	for _, foo := range a.actions {
		err := foo(dataMap)
		ee = append(ee, err)
	}

	return nil, nil, errors.Join(ee...)
}

func (a *MetricAgent) computeMetrics(dataMap map[string]*matrix.Matrix) error {

	var (
		metric                    *matrix.Metric
		metricVal, firstMetricVal *matrix.Metric
		m                         []*matrix.Matrix
		sgMatrix                  bool
		skipMetric                bool
		sgIndex                   string
		sgLabelMap                map[string]string
		err                       error
		metricNotFound            []error
	)

	data := dataMap[a.Object]

	// map values for compute_metric mapping rules
	for _, r := range a.computeMetricRules {
		m = make([]*matrix.Matrix, len(r.metricNames))
		sgLabelMap = make(map[string]string)
		for i := range r.metricNames {
			if data == nil {
				m[i] = dataMap[r.metricNames[i]]
				sgMatrix = true
			} else {
				m[i] = data
			}
		}
		if m[0] == nil {
			return errs.New(errs.ErrMissingMetric, "matrix not found for metric "+r.metricNames[0])
		}
		if metric = a.getMetric(m[0], r.metric); metric == nil {
			if metric, err = m[0].NewMetricFloat64(r.metric); err != nil {
				a.SLogger.Error("Failed to create metric", slogx.Err(err), slog.String("metric", r.metric))
				return err
			}
			metric.SetProperty("compute_metric mapping")
		}

		for iKey, instance := range m[0].GetInstances() {
			skipMetric = false
			if sgMatrix {
				idx := strings.LastIndex(iKey, "-")
				if idx == -1 || idx == len(iKey)-1 {
					a.SLogger.Warn("computeMetrics: unexpected instance key format", slog.String("instanceKey", iKey))
					continue
				}
				sgIndex = iKey[idx+1:]
				sgLabelMap = instance.GetLabels()
			}
			var result float64
			var otherInstance *matrix.Instance

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
					otherInstance = instance
					if sgMatrix {
						iKey = r.metricNames[i] + "-" + sgIndex
						if m[i] == nil || m[i].GetInstance(iKey) == nil {
							a.SLogger.Warn("computeMetrics: matrix or instance not found for metric", slog.String("metric", r.metricNames[i]))
							skipMetric = true
							break
						}
						otherInstance = m[i].GetInstance(iKey)
						if len(sgLabelMap) != len(otherInstance.GetLabels()) || !maps.Equal(sgLabelMap, otherInstance.GetLabels()) {
							a.SLogger.Warn("computeMetrics: skipping compute metric calculation as instance labels are not matching for given metrics", slog.String("metric", r.metric), slog.Any("given metrics", r.metricNames))
							skipMetric = true
							break
						}
					}

					metricVal = a.getMetric(m[i], r.metricNames[i])
					if metricVal != nil {
						v, _ = metricVal.GetValueFloat64(otherInstance)
					} else {
						metricNotFound = append(metricNotFound, err)
						break
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

			if !skipMetric {
				metric.SetValueFloat64(instance, result)
			}

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
