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

// computeMetrics applies each compute_metric rule.
//
// Most collectors return a single matrix for their object, so all operands of a rule share
// an instance. The StorageGrid Prometheus collector returns one matrix per metric instead,
// so a rule's operands live in sibling matrices and are matched by instance labels.
func (a *MetricAgent) computeMetrics(dataMap map[string]*matrix.Matrix) error {

	data := dataMap[a.Object]

	for _, r := range a.computeMetricRules {
		if data != nil {
			a.computeRule(r, data)
		} else {
			a.computeRuleAcrossMatrices(r, dataMap)
		}
	}

	return nil
}

// computeRule computes a rule whose operands all live in matrix m.
func (a *MetricAgent) computeRule(r computeMetricRule, m *matrix.Matrix) {

	first := a.getMetric(m, r.metricNames[0])
	if first == nil {
		a.SLogger.Warn("computeMetrics: metric not found", slog.String("metric", r.metricNames[0]))
		return
	}

	target := a.targetMetric(m, r.metric)
	if target == nil {
		return
	}

	operands := make([]operand, 0, len(r.metricNames)-1)
	for _, name := range r.metricNames[1:] {
		o := a.newOperand(name, m, nil)
		// A missing metric is a config error (typo/misconfig). Do not treat it like an
		// optional field that is merely unset on some instances.
		if !o.isLiteral && o.metric == nil {
			return
		}
		operands = append(operands, o)
	}

	for _, instance := range m.GetInstances() {
		result, ok := first.GetValueFloat64(instance)
		if !ok {
			continue
		}
		for _, o := range operands {
			// An operand without a value is an optional field that is absent for this
			// instance, e.g. hybrid_cache.disk_count on an aggregate without a cache tier.
			v, _ := o.value(instance)
			result = a.apply(r.operation, result, v)
		}
		target.SetValueFloat64(instance, result)
	}
}

// computeRuleAcrossMatrices computes a rule whose operands may live in different matrices.
// The new metric is added to the matrix of the rule's first operand, which is why later
// rules can use it: operands are searched for by metric name across all matrices.
func (a *MetricAgent) computeRuleAcrossMatrices(r computeMetricRule, dataMap map[string]*matrix.Matrix) {

	base := a.findMatrix(dataMap, r.metricNames[0])
	if base == nil {
		a.SLogger.Warn("computeMetrics: metric not found", slog.String("metric", r.metricNames[0]))
		return
	}

	first := a.getMetric(base, r.metricNames[0])
	target := a.targetMetric(base, r.metric)
	if target == nil {
		return
	}

	operands := make([]operand, 0, len(r.metricNames)-1)
	for _, name := range r.metricNames[1:] {
		o := a.newOperand(name, base, dataMap)
		if !o.isLiteral && o.metric == nil {
			return
		}
		operands = append(operands, o)
	}

	for _, instance := range base.GetInstances() {
		result, ok := first.GetValueFloat64(instance)
		if !ok {
			continue
		}
		skip := false
		for _, o := range operands {
			// Unlike the single matrix case, a missing value means the operand describes a
			// different set of instances, so computing with zero would be misleading.
			v, ok := o.value(instance)
			if !ok {
				if o.sibling != nil && instanceWithLabels(o.sibling, instance.GetLabels()) == nil {
					a.SLogger.Error("computeMetrics: skip compute metric since instance labels do not match",
						slog.String("metric", r.metric),
						slog.String("operand", o.name),
						slog.Any("labels", instance.GetLabels()))
				}
				skip = true
				break
			}
			result = a.apply(r.operation, result, v)
		}
		if !skip {
			target.SetValueFloat64(instance, result)
		}
	}
}

// operand is one input of a compute_metric rule: either a literal number or a metric.
// sibling is set when the metric lives in a different matrix than the instances the rule
// iterates over.
type operand struct {
	name      string
	literal   float64
	isLiteral bool
	metric    *matrix.Metric
	sibling   *matrix.Matrix
}

// value returns the operand's value for instance. When the operand comes from a sibling
// matrix, the instance with the same labels is used, since instance keys differ per matrix.
func (o operand) value(instance *matrix.Instance) (float64, bool) {
	switch {
	case o.isLiteral:
		return o.literal, true
	case o.metric == nil:
		return 0, false
	case o.sibling == nil:
		return o.metric.GetValueFloat64(instance)
	default:
		other := instanceWithLabels(o.sibling, instance.GetLabels())
		if other == nil {
			return 0, false
		}
		return o.metric.GetValueFloat64(other)
	}
}

// newOperand resolves an operand name to a literal, a metric of base, or, when dataMap is
// not nil, a metric of one of the sibling matrices.
func (a *MetricAgent) newOperand(name string, base *matrix.Matrix, dataMap map[string]*matrix.Matrix) operand {

	if literal, err := strconv.Atoi(name); err == nil {
		return operand{name: name, literal: float64(literal), isLiteral: true}
	}

	if metric := a.getMetric(base, name); metric != nil {
		return operand{name: name, metric: metric}
	}

	if dataMap != nil {
		if m := a.findMatrix(dataMap, name); m != nil {
			return operand{name: name, metric: a.getMetric(m, name), sibling: m}
		}
	}

	a.SLogger.Warn("computeMetrics: metric not found", slog.String("metric", name))
	return operand{name: name}
}

func (a *MetricAgent) apply(operation string, result float64, v float64) float64 {
	switch operation {
	case "ADD":
		return result + v
	case "SUBTRACT":
		return result - v
	case "MULTIPLY":
		return result * v
	case "DIVIDE":
		if v == 0 {
			return 0
		}
		return result / v
	case "PERCENT":
		if v == 0 {
			return 0
		}
		return (result / v) * 100
	default:
		a.SLogger.Warn("Unknown operation", slog.String("operation", operation))
		return result
	}
}

// targetMetric returns the metric the rule writes to, creating it when needed.
func (a *MetricAgent) targetMetric(m *matrix.Matrix, name string) *matrix.Metric {
	if metric := a.getMetric(m, name); metric != nil {
		return metric
	}
	metric, err := m.NewMetricFloat64(name)
	if err != nil {
		a.SLogger.Error("Failed to create metric", slogx.Err(err), slog.String("metric", name))
		return nil
	}
	metric.SetProperty("compute_metric mapping")
	return metric
}

// findMatrix returns the matrix containing the named metric. Collectors that return one
// matrix per metric key dataMap by metric name, but metrics created by earlier rules are
// added to an existing matrix, so fall back to searching every matrix.
func (a *MetricAgent) findMatrix(dataMap map[string]*matrix.Matrix, name string) *matrix.Matrix {
	if m := dataMap[name]; m != nil && a.getMetric(m, name) != nil {
		return m
	}
	for _, m := range dataMap {
		if a.getMetric(m, name) != nil {
			return m
		}
	}
	return nil
}

// instanceWithLabels returns the first instance whose labels exactly match. Label sets are
// expected to uniquely identify a series within a matrix.
func instanceWithLabels(m *matrix.Matrix, labels map[string]string) *matrix.Instance {
	for _, instance := range m.GetInstances() {
		if maps.Equal(instance.GetLabels(), labels) {
			return instance
		}
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
