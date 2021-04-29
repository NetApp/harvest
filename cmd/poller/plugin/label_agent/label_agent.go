/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package label_agent

import (
	"fmt"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/matrix"
	"strings"
)

type LabelAgent struct {
	*plugin.AbstractPlugin
	actions              []func(*matrix.Instance)
	splitSimpleRules     []splitSimpleRule
	splitRegexRules      []splitRegexRule
	splitPairsRules      []splitPairsRule
	joinSimpleRules      []joinSimpleRule
	replaceSimpleRules   []replaceSimpleRule
	replaceRegexRules    []replaceRegexRule
	excludeEqualsRules   []excludeEqualsRule
	excludeContainsRules []excludeContainsRule
	excludeRegexRules    []excludeRegexRule
	valueMappingRules    []valueMappingRule
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
		err = errors.New(errors.MISSING_PARAM, "valid rules")
	} else {
		logger.Debug(me.Prefix, "parsed %d rules for %d actions", count, len(me.actions))
	}

	return err
}

func (me *LabelAgent) Run(m *matrix.Matrix) ([]*matrix.Matrix, error) {

	var err error

	for _, instance := range m.GetInstances() {
		for _, foo := range me.actions {
			foo(instance)
		}
	}

	if len(me.valueMappingRules) != 0 {
		err = me.mapValues(m)
	}

	return nil, err
}

// splits one label value into multiple labels using seperator symbol
func (me *LabelAgent) splitSimple(instance *matrix.Instance) {
	for _, r := range me.splitSimpleRules {
		if values := strings.Split(instance.GetLabel(r.source), r.sep); len(values) >= len(r.targets) {
			for i := range r.targets {
				if r.targets[i] != "" && values[i] != "" {
					instance.SetLabel(r.targets[i], values[i])
					logger.Trace(me.Prefix, "splitSimple: (%s) [%s] => (%s) [%s]", r.source, instance.GetLabel(r.source), r.targets[i], values[i])
				}
			}
		}
	}
}

// splits one label value into multiple labels based on regex match
func (me *LabelAgent) splitRegex(instance *matrix.Instance) {
	for _, r := range me.splitRegexRules {
		if m := r.reg.FindStringSubmatch(instance.GetLabel(r.source)); m != nil && len(m) == len(r.targets)+1 {
			for i := range r.targets {
				if r.targets[i] != "" && m[i+1] != "" {
					instance.SetLabel(r.targets[i], m[i+1])
					logger.Trace(me.Prefix, "splitRegex: (%s) [%s] => (%s) [%s]", r.source, instance.GetLabel(r.source), r.targets[i], m[i+1])
				}
			}
		}
	}
}

// splits one label value into multiple key-value pairs
func (me *LabelAgent) splitPairs(instance *matrix.Instance) {
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

// joins multiple labels into one label
func (me *LabelAgent) joinSimple(instance *matrix.Instance) {
	for _, r := range me.joinSimpleRules {
		values := make([]string, 0)
		for _, label := range r.sources {
			if v := instance.GetLabel(label); v != "" {
				values = append(values, v)
			}
		}
		if len(values) != 0 {
			instance.SetLabel(r.target, strings.Join(values, r.sep))
			logger.Trace(me.Prefix, "joinSimple: (%v) => (%s) [%s]", r.sources, r.target, instance.GetLabel(r.target))
		}
	}
}

// replace in source label, if present, and add as new label
func (me *LabelAgent) replaceSimple(instance *matrix.Instance) {
	for _, r := range me.replaceSimpleRules {
		if old := instance.GetLabel(r.source); old != "" {
			if value := strings.ReplaceAll(old, r.old, r.new); value != old {
				instance.SetLabel(r.target, value)
				logger.Trace(me.Prefix, "replaceSimple: (%s) [%s] => (%s) [%s]", r.source, old, r.target, value)
			}
		}
	}
}

// same as replaceSimple, but use regex
func (me *LabelAgent) replaceRegex(instance *matrix.Instance) {
	for _, r := range me.replaceRegexRules {
		old := instance.GetLabel(r.source)
		if m := r.reg.FindStringSubmatch(old); m != nil {
			logger.Trace(me.Prefix, "replaceRegex: (%d) matches= %v", len(m)-1, m[1:])
			s := make([]interface{}, 0)
			for _, i := range r.indices {
				if i < len(m)-1 {
					s = append(s, m[i+1])
					logger.Trace(me.Prefix, "substring [%d] = (%s)", i, m[i+1])
				} else {
					// probably we need to throw warning
					s = append(s, "")
					logger.Trace(me.Prefix, "substring [%d] = no match!", i)
				}
			}
			logger.Trace(me.Prefix, "replaceRegex: (%d) substitution strings= %v", len(s), s)
			if value := fmt.Sprintf(r.format, s...); value != "" && value != old {
				instance.SetLabel(r.target, value)
				logger.Trace(me.Prefix, "replaceRegex: (%s) [%s] => (%s) [%s]", r.source, old, r.target, value)
			}
		}
	}
}

// if label equals to value, set instance as non-exportable
func (me *LabelAgent) excludeEquals(instance *matrix.Instance) {
	for _, r := range me.excludeEqualsRules {
		if instance.GetLabel(r.label) == r.value {
			instance.SetExportable(false)
			logger.Trace(me.Prefix, "excludeEquals: (%s) [%s] instance with labels [%s] => excluded", r.label, r.value, instance.GetLabels().String())
			break
		}
	}
}

// if label contains value, set instance as non-exportable
func (me *LabelAgent) excludeContains(instance *matrix.Instance) {
	for _, r := range me.excludeContainsRules {
		if strings.Contains(instance.GetLabel(r.label), r.value) {
			instance.SetExportable(false)
			logger.Trace(me.Prefix, "excludeContains: (%s) [%s] instance with labels [%s] => excluded", r.label, r.value, instance.GetLabels().String())
			break
		}
	}
}

// if label equals to value, set instance as non-exportable
func (me *LabelAgent) excludeRegex(instance *matrix.Instance) {
	for _, r := range me.excludeRegexRules {
		if r.reg.MatchString(instance.GetLabel(r.label)) {
			instance.SetExportable(false)
			logger.Trace(me.Prefix, "excludeEquals: (%s) [%s] instance with labels [%s] => excluded", r.label, r.reg.String(), instance.GetLabels().String())
			break
		}
	}
}

func (me *LabelAgent) mapValues(m *matrix.Matrix) error {

	var (
		metric matrix.Metric
		err    error
	)

	for _, r := range me.valueMappingRules {

		if metric = m.GetMetric(r.metric); metric == nil {
			if metric, err = m.NewMetricUint8(r.metric); err != nil {
				logger.Error(me.Prefix, "valueMapping: new metric [%s]: %v", r.metric, err)
				return err
			} else {
				metric.SetProperty("mapping")
			}
		}

		for key, instance := range m.GetInstances() {
			if v, ok := r.mapping[instance.GetLabel(r.label)]; ok {
				metric.SetValueUint8(instance, v)
				logger.Trace(me.Prefix, "valueMapping: [%s] [%s] mapped (%s) value to %d", r.metric, key, instance.GetLabel(r.label), v)
			} else if r.hasDefault {
				metric.SetValueUint8(instance, r.defaultValue)
				logger.Trace(me.Prefix, "valueMapping: [%s] [%s] mapped (%s) value to default %d", r.metric, key, instance.GetLabel(r.label), r.defaultValue)
			}
		}
	}

	return nil
}
