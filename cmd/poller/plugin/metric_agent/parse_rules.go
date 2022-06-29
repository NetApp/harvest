/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package metric_agent

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"strings"
)

// parse rules from plugin parameters and return number of rules parsed
func (me *MetricAgent) parseRules() int {

	me.computeMetricRules = make([]computeMetricRule, 0)

	for _, c := range me.Params.GetChildren() {
		name := c.GetNameS()

		rules := c.GetChildren()
		// loop over all rules
		for _, rc := range rules {
			rule := strings.TrimSpace(rc.GetContentS())

			switch name {
			case "compute_metric":
				me.parseComputeMetricRule(rule)
			default:
				me.Logger.Warn().
					Str("object", me.ParentParams.GetChildContentS("object")).
					Str("name", name).Msg("Unknown rule name")
			}
		}
	}

	me.actions = make([]func(matrix *matrix.Matrix) error, 0)
	count := 0

	for _, c := range me.Params.GetChildren() {
		name := c.GetNameS()
		switch name {
		case "compute_metric":
			if len(me.computeMetricRules) != 0 {
				me.actions = append(me.actions, me.computeMetrics)
				count += len(me.computeMetricRules)
			}
		default:
			me.Logger.Warn().
				Str("object", me.ParentParams.GetChildContentS("object")).
				Str("name", name).Msg("Unknown rule name")
		}
	}

	return count
}

type computeMetricRule struct {
	metric      string
	operation   string
	metricNames []string
}

func (me *MetricAgent) parseComputeMetricRule(rule string) {
	if fields := strings.Fields(rule); len(fields) >= 4 {
		r := computeMetricRule{metric: fields[0], operation: fields[1], metricNames: make([]string, 0)}

		for i := 2; i < len(fields); i++ {
			r.metricNames = append(r.metricNames, fields[i])
		}

		me.computeMetricRules = append(me.computeMetricRules, r)
		me.Logger.Debug().Msgf("(compute_metric) parsed rule [%v]", r)
		return
	}
	me.Logger.Warn().Msgf("(compute_metric) rule has invalid format [%s]", rule)
}
