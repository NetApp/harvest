/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package metricagent

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"log/slog"
	"strings"
)

// parse rules from plugin parameters and return number of rules parsed
func (a *MetricAgent) parseRules() int {

	a.computeMetricRules = make([]computeMetricRule, 0)

	for _, c := range a.Params.GetChildren() {
		name := c.GetNameS()

		rules := c.GetChildren()
		// loop over all rules
		for _, rc := range rules {
			rule := strings.TrimSpace(rc.GetContentS())

			switch name {
			case "compute_metric":
				a.parseComputeMetricRule(rule)
			default:
				a.SLogger.Warn(
					"Unknown rule name",
					slog.String("object", a.ParentParams.GetChildContentS("object")),
					slog.String("name", name),
				)
			}
		}
	}

	a.actions = make([]func(matrix *matrix.Matrix) error, 0)
	count := 0

	for _, c := range a.Params.GetChildren() {
		name := c.GetNameS()
		switch name {
		case "compute_metric":
			if len(a.computeMetricRules) != 0 {
				a.actions = append(a.actions, a.computeMetrics)
				count += len(a.computeMetricRules)
			}
		default:
			a.SLogger.Warn(
				"Unknown rule name",
				slog.String("object", a.ParentParams.GetChildContentS("object")),
				slog.String("name", name),
			)
		}
	}

	return count
}

type computeMetricRule struct {
	metric      string
	operation   string
	metricNames []string
}

func (a *MetricAgent) parseComputeMetricRule(rule string) {
	if fields := strings.Fields(rule); len(fields) >= 4 {
		r := computeMetricRule{metric: fields[0], operation: fields[1], metricNames: make([]string, 0)}

		for i := 2; i < len(fields); i++ {
			r.metricNames = append(r.metricNames, fields[i])
		}

		a.computeMetricRules = append(a.computeMetricRules, r)
		a.SLogger.Debug(
			"(compute_metric) parsed rule",
			slog.String("metric", r.metric),
			slog.String("operation", r.operation),
		)
		return
	}
	a.SLogger.Warn("(compute_metric) rule has invalid format", slog.String("rule", rule))
}
