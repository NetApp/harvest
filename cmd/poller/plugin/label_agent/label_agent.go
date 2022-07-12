/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package label_agent

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
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
	valueToNumRegexRules []valueToNumRegexRule
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &LabelAgent{AbstractPlugin: p}
}

func (a *LabelAgent) Init() error {

	var (
		err   error
		count int
	)

	if err = a.AbstractPlugin.Init(); err != nil {
		return err
	}

	if count = a.parseRules(); count == 0 {
		err = errs.New(errs.ErrMissingParam, "valid rules")
	} else {
		a.Logger.Debug().Msgf("parsed %d rules for %d actions", count, len(a.actions))
	}

	return err
}

func (a *LabelAgent) Run(m *matrix.Matrix) ([]*matrix.Matrix, error) {

	var err error

	for _, foo := range a.actions {
		_ = foo(m)
	}

	return nil, err
}

// splits one label value into multiple labels using separator symbol
func (a *LabelAgent) splitSimple(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range a.splitSimpleRules {
			if values := strings.Split(instance.GetLabel(r.source), r.sep); len(values) >= len(r.targets) {
				for i := range r.targets {
					if r.targets[i] != "" && values[i] != "" {
						instance.SetLabel(r.targets[i], values[i])
						a.Logger.Trace().Msgf("splitSimple: (%s) [%s] => (%s) [%s]", r.source, instance.GetLabel(r.source), r.targets[i], values[i])
					}
				}
			}
		}
	}
	return nil
}

// splits one label value into multiple labels based on regex match
func (a *LabelAgent) splitRegex(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range a.splitRegexRules {
			if m := r.reg.FindStringSubmatch(instance.GetLabel(r.source)); m != nil && len(m) == len(r.targets)+1 {
				for i := range r.targets {
					if r.targets[i] != "" && m[i+1] != "" {
						instance.SetLabel(r.targets[i], m[i+1])
						a.Logger.Trace().Msgf("splitRegex: (%s) [%s] => (%s) [%s]", r.source, instance.GetLabel(r.source), r.targets[i], m[i+1])
					}
				}
			}
		}
	}
	return nil
}

// splits one label value into multiple key-value pairs
func (a *LabelAgent) splitPairs(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range a.splitPairsRules {
			if value := instance.GetLabel(r.source); value != "" {
				for _, pair := range strings.Split(value, r.sep1) {
					if kv := strings.Split(pair, r.sep2); len(kv) == 2 {
						instance.SetLabel(kv[0], kv[1])
					}
				}
			}
		}
	}
	return nil
}

// joins multiple labels into one label
func (a *LabelAgent) joinSimple(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range a.joinSimpleRules {
			values := make([]string, 0)
			for _, label := range r.sources {
				if v := instance.GetLabel(label); v != "" {
					values = append(values, v)
				}
			}
			if len(values) != 0 {
				instance.SetLabel(r.target, strings.Join(values, r.sep))
				a.Logger.Trace().Msgf("joinSimple: (%v) => (%s) [%s]", r.sources, r.target, instance.GetLabel(r.target))
			}
		}
	}
	return nil
}

// replace in source label, if present, and add as new label
func (a *LabelAgent) replaceSimple(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range a.replaceSimpleRules {
			if old := instance.GetLabel(r.source); old != "" {
				if value := strings.ReplaceAll(old, r.old, r.new); value != old {
					instance.SetLabel(r.target, value)
					a.Logger.Trace().Msgf("replaceSimple: (%s) [%s] => (%s) [%s]", r.source, old, r.target, value)
				}
			}
		}
	}
	return nil
}

// same as replaceSimple, but use regex
func (a *LabelAgent) replaceRegex(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range a.replaceRegexRules {
			old := instance.GetLabel(r.source)
			if m := r.reg.FindStringSubmatch(old); m != nil {
				a.Logger.Trace().Msgf("replaceRegex: (%d) matches= %v", len(m)-1, m[1:])
				s := make([]interface{}, 0)
				for _, i := range r.indices {
					if i < len(m)-1 {
						s = append(s, m[i+1])
						a.Logger.Trace().Msgf("substring [%d] = (%s)", i, m[i+1])
					} else {
						// probably we need to throw warning
						s = append(s, "")
						a.Logger.Trace().Msgf("substring [%d] = no match!", i)
					}
				}
				a.Logger.Trace().Msgf("replaceRegex: (%d) substitution strings= %v", len(s), s)
				if value := fmt.Sprintf(r.format, s...); value != "" && value != old {
					instance.SetLabel(r.target, value)
					a.Logger.Trace().Msgf("replaceRegex: (%s) [%s] => (%s) [%s]", r.source, old, r.target, value)
				}
			}
		}
	}
	return nil
}

// if label equals to value, set instance as non-exportable
func (a *LabelAgent) excludeEquals(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range a.excludeEqualsRules {
			if instance.GetLabel(r.label) == r.value {
				instance.SetExportable(false)
				a.Logger.Trace().Str("label", r.label).
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
func (a *LabelAgent) excludeContains(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range a.excludeContainsRules {
			if strings.Contains(instance.GetLabel(r.label), r.value) {
				instance.SetExportable(false)
				a.Logger.Trace().Str("label", r.label).
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
func (a *LabelAgent) excludeRegex(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		for _, r := range a.excludeRegexRules {
			if r.reg.MatchString(instance.GetLabel(r.label)) {
				instance.SetExportable(false)
				a.Logger.Trace().Str("label", r.label).
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
func (a *LabelAgent) includeEquals(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		if instance.IsExportable() {
			isExport := false
			for _, r := range a.includeEqualsRules {
				if instance.GetLabel(r.label) == r.value {
					isExport = true
					a.Logger.Trace().Str("label", r.label).
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
func (a *LabelAgent) includeContains(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		if instance.IsExportable() {
			isExport := false
			for _, r := range a.includeContainsRules {
				if strings.Contains(instance.GetLabel(r.label), r.value) {
					isExport = true
					a.Logger.Trace().Str("label", r.label).
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
func (a *LabelAgent) includeRegex(matrix *matrix.Matrix) error {
	for _, instance := range matrix.GetInstances() {
		if instance.IsExportable() {
			isExport := false
			for _, r := range a.includeRegexRules {
				if r.reg.MatchString(instance.GetLabel(r.label)) {
					isExport = true
					a.Logger.Trace().Str("label", r.label).
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

func (a *LabelAgent) mapValueToNum(m *matrix.Matrix) error {

	var (
		metric matrix.Metric
		err    error
	)

	// map values for value_to_num mapping rules
	for _, r := range a.valueToNumRules {

		if metric = m.GetMetric(r.metric); metric == nil {
			if metric, err = m.NewMetricUint8(r.metric); err != nil {
				a.Logger.Error().Stack().Err(err).Msgf("valueToNumMapping: new metric [%s]:", r.metric)
				return err
			} else {
				metric.SetProperty("value_to_num mapping")
			}
		}

		for key, instance := range m.GetInstances() {
			if v, ok := r.mapping[instance.GetLabel(r.label)]; ok {
				_ = metric.SetValueUint8(instance, v)
				a.Logger.Trace().Msgf("valueToNumMapping: [%s] [%s] mapped (%s) value to %d", r.metric, key, instance.GetLabel(r.label), v)
			} else if r.hasDefault {
				_ = metric.SetValueUint8(instance, r.defaultValue)
				a.Logger.Trace().Msgf("valueToNumMapping: [%s] [%s] mapped (%s) value to default %d", r.metric, key, instance.GetLabel(r.label), r.defaultValue)
			}
		}
	}

	return nil
}

func (a *LabelAgent) mapValueToNumRegex(m *matrix.Matrix) error {
	var (
		metric matrix.Metric
		err    error
	)

	// map values for value_to_num mapping rules
	for _, r := range a.valueToNumRegexRules {
		if metric = m.GetMetric(r.metric); metric == nil {
			if metric, err = m.NewMetricUint8(r.metric); err != nil {
				a.Logger.Error().Stack().Err(err).Msgf("valueToNumRegexMapping: new metric [%s]:", r.metric)
				return err
			} else {
				metric.SetProperty("value_to_num_regex mapping")
			}
		}

		for key, instance := range m.GetInstances() {
			value := instance.GetLabel(r.label)
			if r.reg[0].MatchString(value) || r.reg[1].MatchString(value) {
				_ = metric.SetValueUint8(instance, uint8(1))
				a.Logger.Trace().Msgf("valueToNumRegexMapping: [%s] [%s] mapped (%s) value to %d, regex1: %s, regex2: %s", r.metric, key, value, 1, r.reg[0], r.reg[1])
			} else if r.hasDefault {
				_ = metric.SetValueUint8(instance, r.defaultValue)
				a.Logger.Trace().Msgf("valueToNumRegexMapping: [%s] [%s] mapped (%s) value to default %d, regex1: %s, regex2: %s", r.metric, key, value, r.defaultValue, r.reg[0], r.reg[1])
			}
		}
	}
	return nil
}
