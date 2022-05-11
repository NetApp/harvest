/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package label_agent

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errors"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"strconv"
	"strings"
)

type LabelAgent struct {
	*plugin.AbstractPlugin
	actions              []func(*matrix.Matrix) error
	splitSimpleRules     []splitSimpleRule
	splitRegexRules      []splitRegexRule
	splitPairsRules      []splitPairsRule
	joinSimpleRules      []joinSimpleRule
	replaceSimpleRules   []replaceSimpleRule
	replaceRegexRules    []replaceRegexRule
	excludeEqualsRules   []excludeEqualsRule
	excludeContainsRules []excludeContainsRule
	excludeRegexRules    []excludeRegexRule
	includeEqualsRules   []includeEqualsRule
	includeContainsRules []includeContainsRule
	includeRegexRules    []includeRegexRule
	valueToNumRules      []valueToNumRule
	computeMetricRules   []computeMetricRule
	valueToNumRegexRules []valueToNumRegexRule
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &LabelAgent{AbstractPlugin: p}
}

func (me *LabelAgent) Init() error {

	var (
		err   error
		count int
	)

	if err = me.AbstractPlugin.Init(); err != nil {
		return err
	}

	if count = me.parseRules(); count == 0 {
		err = errors.New(errors.MissingParam, "valid rules")
	} else {
		me.Logger.Debug().Msgf("parsed %d rules for %d actions", count, len(me.actions))
	}

	return err
}

func (me *LabelAgent) Run(m *matrix.Matrix) ([]*matrix.Matrix, error) {

	var err error

	for _, foo := range me.actions {
		foo(m)
	}

	return nil, err
}

// splits one label value into multiple labels using separator symbol
func (me *LabelAgent) splitSimple(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range me.splitSimpleRules {
			if values := strings.Split(instance.GetLabel(r.source), r.sep); len(values) >= len(r.targets) {
				for i := range r.targets {
					if r.targets[i] != "" && values[i] != "" {
						instance.SetLabel(r.targets[i], values[i])
						me.Logger.Trace().Msgf("splitSimple: (%s) [%s] => (%s) [%s]", r.source, instance.GetLabel(r.source), r.targets[i], values[i])
					}
				}
			}
		}
	}
	return nil
}

// splits one label value into multiple labels based on regex match
func (me *LabelAgent) splitRegex(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range me.splitRegexRules {
			if m := r.reg.FindStringSubmatch(instance.GetLabel(r.source)); m != nil && len(m) == len(r.targets)+1 {
				for i := range r.targets {
					if r.targets[i] != "" && m[i+1] != "" {
						instance.SetLabel(r.targets[i], m[i+1])
						me.Logger.Trace().Msgf("splitRegex: (%s) [%s] => (%s) [%s]", r.source, instance.GetLabel(r.source), r.targets[i], m[i+1])
					}
				}
			}
		}
	}
	return nil
}

// splits one label value into multiple key-value pairs
func (me *LabelAgent) splitPairs(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range me.splitPairsRules {
			if value := instance.GetLabel(r.source); value != "" {
				for _, pair := range strings.Split(value, r.sep1) {
					if kv := strings.Split(pair, r.sep2); len(kv) == 2 {
						instance.SetLabel(kv[0], kv[1])
						//logger.Trace(me.Prefix, "splitPair: ($s) [%s] => (%s) [%s]", r.source, value, kv[0], kv[1])
					}
				}
			}
		}
	}
	return nil
}

// joins multiple labels into one label
func (me *LabelAgent) joinSimple(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range me.joinSimpleRules {
			values := make([]string, 0)
			for _, label := range r.sources {
				if v := instance.GetLabel(label); v != "" {
					values = append(values, v)
				}
			}
			if len(values) != 0 {
				instance.SetLabel(r.target, strings.Join(values, r.sep))
				me.Logger.Trace().Msgf("joinSimple: (%v) => (%s) [%s]", r.sources, r.target, instance.GetLabel(r.target))
			}
		}
	}
	return nil
}

// replace in source label, if present, and add as new label
func (me *LabelAgent) replaceSimple(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range me.replaceSimpleRules {
			if old := instance.GetLabel(r.source); old != "" {
				if value := strings.ReplaceAll(old, r.old, r.new); value != old {
					instance.SetLabel(r.target, value)
					me.Logger.Trace().Msgf("replaceSimple: (%s) [%s] => (%s) [%s]", r.source, old, r.target, value)
				}
			}
		}
	}
	return nil
}

// same as replaceSimple, but use regex
func (me *LabelAgent) replaceRegex(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range me.replaceRegexRules {
			old := instance.GetLabel(r.source)
			if m := r.reg.FindStringSubmatch(old); m != nil {
				me.Logger.Trace().Msgf("replaceRegex: (%d) matches= %v", len(m)-1, m[1:])
				s := make([]interface{}, 0)
				for _, i := range r.indices {
					if i < len(m)-1 {
						s = append(s, m[i+1])
						me.Logger.Trace().Msgf("substring [%d] = (%s)", i, m[i+1])
					} else {
						// probably we need to throw warning
						s = append(s, "")
						me.Logger.Trace().Msgf("substring [%d] = no match!", i)
					}
				}
				me.Logger.Trace().Msgf("replaceRegex: (%d) substitution strings= %v", len(s), s)
				if value := fmt.Sprintf(r.format, s...); value != "" && value != old {
					instance.SetLabel(r.target, value)
					me.Logger.Trace().Msgf("replaceRegex: (%s) [%s] => (%s) [%s]", r.source, old, r.target, value)
				}
			}
		}
	}
	return nil
}

// if label equals to value, set instance as non-exportable
func (me *LabelAgent) excludeEquals(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range me.excludeEqualsRules {
			if instance.GetLabel(r.label) == r.value {
				instance.SetExportable(false)
				me.Logger.Trace().Str("label", r.label).
					Str("value", r.value).
					Str("instance labels", instance.GetLabels().String()).
					Msg("excludeEquals: excluded")
				break
			}
		}
	}
	return nil
}

// if label contains value, set instance as non-exportable
func (me *LabelAgent) excludeContains(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range me.excludeContainsRules {
			if strings.Contains(instance.GetLabel(r.label), r.value) {
				instance.SetExportable(false)
				me.Logger.Trace().Str("label", r.label).
					Str("value", r.value).
					Str("instance labels", instance.GetLabels().String()).
					Msg("excludeContains: excluded")
				break
			}
		}
	}
	return nil
}

// if label equals to value, set instance as non-exportable
func (me *LabelAgent) excludeRegex(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range me.excludeRegexRules {
			if r.reg.MatchString(instance.GetLabel(r.label)) {
				instance.SetExportable(false)
				me.Logger.Trace().Str("label", r.label).
					Str("regex", r.reg.String()).
					Str("instance labels", instance.GetLabels().String()).
					Msg("excludeRegex: excluded")
				break
			}
		}
	}
	return nil
}

// if label is not equal to value, set instance as non-exportable
func (me *LabelAgent) includeEquals(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		if instance.IsExportable() {
			isExport := false
			for _, r := range me.includeEqualsRules {
				if instance.GetLabel(r.label) == r.value {
					isExport = true
					me.Logger.Trace().Str("label", r.label).
						Str("value", r.value).
						Str("instance labels", instance.GetLabels().String()).
						Msg("includeEquals: included")
					break
				}
			}
			instance.SetExportable(isExport)
		}
	}
	return nil
}

// if label does not contains value, set instance as non-exportable
func (me *LabelAgent) includeContains(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		if instance.IsExportable() {
			isExport := false
			for _, r := range me.includeContainsRules {
				if strings.Contains(instance.GetLabel(r.label), r.value) {
					isExport = true
					me.Logger.Trace().Str("label", r.label).
						Str("value", r.value).
						Str("instance labels", instance.GetLabels().String()).
						Msg("includeContains: included")
					break
				}
			}
			instance.SetExportable(isExport)
		}
	}
	return nil
}

// if label does not match regex, do not export the instance or with fewer negatives
// only export instances with a matching (regex) label
// if an instance does not match the regex label it will not be exported
func (me *LabelAgent) includeRegex(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		if instance.IsExportable() {
			isExport := false
			for _, r := range me.includeRegexRules {
				if r.reg.MatchString(instance.GetLabel(r.label)) {
					isExport = true
					me.Logger.Trace().Str("label", r.label).
						Str("regex", r.reg.String()).
						Str("instance labels", instance.GetLabels().String()).
						Msg("includeRegex: included")
					break
				}
			}
			instance.SetExportable(isExport)
		}
	}
	return nil
}

func (me *LabelAgent) mapValueToNum(m *matrix.Matrix) error {

	var (
		metric matrix.Metric
		err    error
	)

	// map values for value_to_num mapping rules
	for _, r := range me.valueToNumRules {

		if metric = m.GetMetric(r.metric); metric == nil {
			if metric, err = m.NewMetricUint8(r.metric); err != nil {
				me.Logger.Error().Stack().Err(err).Msgf("valueToNumMapping: new metric [%s]:", r.metric)
				return err
			} else {
				metric.SetProperty("value_to_num mapping")
			}
		}

		for key, instance := range m.GetInstances() {
			if v, ok := r.mapping[instance.GetLabel(r.label)]; ok {
				_ = metric.SetValueUint8(instance, v)
				me.Logger.Trace().Msgf("valueToNumMapping: [%s] [%s] mapped (%s) value to %d", r.metric, key, instance.GetLabel(r.label), v)
			} else if r.hasDefault {
				_ = metric.SetValueUint8(instance, r.defaultValue)
				me.Logger.Trace().Msgf("valueToNumMapping: [%s] [%s] mapped (%s) value to default %d", r.metric, key, instance.GetLabel(r.label), r.defaultValue)
			}
		}
	}

	return nil
}

func (me *LabelAgent) computeMetrics(m *matrix.Matrix) error {

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
				me.Logger.Warn().Stack().Err(err).Str("metricName", r.metricNames[0]).Msg("computeMetrics: metric not found")
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
						me.Logger.Warn().Stack().Err(err).Str("metricName", r.metricNames[i]).Msg("computeMetrics: metric not found")
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

func (me *LabelAgent) mapValueToNumRegex(m *matrix.Matrix) error {
	var (
		metric matrix.Metric
		err    error
	)

	// map values for value_to_num mapping rules
	for _, r := range me.valueToNumRegexRules {
		if metric = m.GetMetric(r.metric); metric == nil {
			if metric, err = m.NewMetricUint8(r.metric); err != nil {
				me.Logger.Error().Stack().Err(err).Msgf("valueToNumRegexMapping: new metric [%s]:", r.metric)
				return err
			} else {
				metric.SetProperty("value_to_num_regex mapping")
			}
		}

		for key, instance := range m.GetInstances() {
			value := instance.GetLabel(r.label)
			if r.reg[0].MatchString(value) || r.reg[1].MatchString(value) {
				_ = metric.SetValueUint8(instance, uint8(1))
				me.Logger.Trace().Msgf("valueToNumRegexMapping: [%s] [%s] mapped (%s) value to %d, regex1: %s, regex2: %s", r.metric, key, value, 1, r.reg[0], r.reg[1])
			} else if r.hasDefault {
				_ = metric.SetValueUint8(instance, r.defaultValue)
				me.Logger.Trace().Msgf("valueToNumRegexMapping: [%s] [%s] mapped (%s) value to default %d, regex1: %s, regex2: %s", r.metric, key, value, r.defaultValue, r.reg[0], r.reg[1])
			}
		}
	}
	return nil
}
