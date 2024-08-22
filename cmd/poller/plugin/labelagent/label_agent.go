/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package labelagent

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"strings"
)

type LabelAgent struct {
	*plugin.AbstractPlugin
	actions              []func(*matrix.Matrix) error
	newLabelNames        []string
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

func New(p *plugin.AbstractPlugin) *LabelAgent {
	return &LabelAgent{AbstractPlugin: p}
}

func (a *LabelAgent) Init() error {

	var (
		err   error
		count int
	)

	if err := a.AbstractPlugin.Init(); err != nil {
		return err
	}

	if count = a.parseRules(); count == 0 {
		err = errs.New(errs.ErrMissingParam, "valid rules")
	} else {
		a.Logger.Debug().Msgf("parsed %d rules for %d actions", count, len(a.actions))
	}

	return err
}

func (a *LabelAgent) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	var err error

	data := dataMap[a.Object]
	for _, foo := range a.actions {
		_ = foo(data)
	}

	return nil, nil, err
}

// splits one label value into multiple labels using separator symbol
func (a *LabelAgent) splitSimple(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
		for _, r := range a.splitSimpleRules {
			if values := strings.Split(instance.GetLabel(r.source), r.sep); len(values) >= len(r.targets) {
				for i := range r.targets {
					if r.targets[i] != "" && values[i] != "" {
						instance.SetLabel(r.targets[i], values[i])
					}
				}
			}
		}
	}
	return nil
}

// splits one label value into multiple labels based on regex match
func (a *LabelAgent) splitRegex(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
		for _, r := range a.splitRegexRules {
			if m := r.reg.FindStringSubmatch(instance.GetLabel(r.source)); m != nil && len(m) == len(r.targets)+1 {
				for i := range r.targets {
					if r.targets[i] != "" && m[i+1] != "" {
						instance.SetLabel(r.targets[i], m[i+1])
					}
				}
			}
		}
	}
	return nil
}

// splits one label value into multiple key-value pairs
func (a *LabelAgent) splitPairs(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
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
func (a *LabelAgent) joinSimple(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
		for _, r := range a.joinSimpleRules {
			values := make([]string, 0)
			for _, label := range r.sources {
				if v := instance.GetLabel(label); v != "" {
					values = append(values, v)
				}
			}
			if len(values) != 0 {
				instance.SetLabel(r.target, strings.Join(values, r.sep))
			}
		}
	}
	return nil
}

// replace in source label, if present, and add as new label
func (a *LabelAgent) replaceSimple(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
		for _, r := range a.replaceSimpleRules {
			if old := instance.GetLabel(r.source); old != "" {
				if value := strings.ReplaceAll(old, r.old, r.new); value != old {
					instance.SetLabel(r.target, value)
				}
			}
		}
	}
	return nil
}

// same as replaceSimple, but use regex
func (a *LabelAgent) replaceRegex(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
		for _, r := range a.replaceRegexRules {
			old := instance.GetLabel(r.source)
			if m := r.reg.FindStringSubmatch(old); m != nil {
				s := make([]interface{}, 0)
				for _, i := range r.indices {
					if i < len(m)-1 {
						s = append(s, m[i+1])
					} else {
						// probably we need to throw warning
						s = append(s, "")
					}
				}
				if value := fmt.Sprintf(r.format, s...); value != "" && value != old {
					instance.SetLabel(r.target, value)
				}
			}
		}
	}
	return nil
}

// if label equals to value, set instance as non-exportable
func (a *LabelAgent) excludeEquals(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
		for _, r := range a.excludeEqualsRules {
			if instance.GetLabel(r.label) == r.value {
				instance.SetExportable(false)
				break
			}
		}
	}
	return nil
}

// if label contains value, set instance as non-exportable
func (a *LabelAgent) excludeContains(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
		for _, r := range a.excludeContainsRules {
			if strings.Contains(instance.GetLabel(r.label), r.value) {
				instance.SetExportable(false)
				break
			}
		}
	}
	return nil
}

// if label equals to value, set instance as non-exportable
func (a *LabelAgent) excludeRegex(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
		for _, r := range a.excludeRegexRules {
			if r.reg.MatchString(instance.GetLabel(r.label)) {
				instance.SetExportable(false)
				break
			}
		}
	}
	return nil
}

// if label is not equal to value, set instance as non-exportable
func (a *LabelAgent) includeEquals(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
		if instance.IsExportable() {
			isExport := false
			for _, r := range a.includeEqualsRules {
				if instance.GetLabel(r.label) == r.value {
					isExport = true
					break
				}
			}
			instance.SetExportable(isExport)
		}
	}
	return nil
}

// if label does not contains value, set instance as non-exportable
func (a *LabelAgent) includeContains(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
		if instance.IsExportable() {
			isExport := false
			for _, r := range a.includeContainsRules {
				if strings.Contains(instance.GetLabel(r.label), r.value) {
					isExport = true
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
func (a *LabelAgent) includeRegex(aMatrix *matrix.Matrix) error {
	for _, instance := range aMatrix.GetInstances() {
		if instance.IsExportable() {
			isExport := false
			for _, r := range a.includeRegexRules {
				if r.reg.MatchString(instance.GetLabel(r.label)) {
					isExport = true
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
		metric *matrix.Metric
		err    error
	)

	// map values for value_to_num mapping rules
	for _, r := range a.valueToNumRules {

		if metric = m.GetMetric(r.metric); metric == nil {
			if metric, err = m.NewMetricUint8(r.metric); err != nil {
				a.Logger.Error().Err(err).Str("metric", r.metric).Msg("valueToNumMapping")
				return err
			}
			metric.SetProperty("value_to_num mapping")
		}

		for _, instance := range m.GetInstances() {
			if v, ok := r.mapping[instance.GetLabel(r.label)]; ok {
				_ = metric.SetValueUint8(instance, v)
			} else if r.hasDefault {
				_ = metric.SetValueUint8(instance, r.defaultValue)
			}
		}
	}

	return nil
}

func (a *LabelAgent) mapValueToNumRegex(m *matrix.Matrix) error {
	var (
		metric *matrix.Metric
		err    error
	)

	// map values for value_to_num mapping rules
	for _, r := range a.valueToNumRegexRules {
		if metric = m.GetMetric(r.metric); metric == nil {
			if metric, err = m.NewMetricUint8(r.metric); err != nil {
				a.Logger.Error().Err(err).Str("metric", r.metric).Msg("valueToNumRegexMapping")
				return err
			}
			metric.SetProperty("value_to_num_regex mapping")
		}

		for _, instance := range m.GetInstances() {
			value := instance.GetLabel(r.label)
			if r.reg[0].MatchString(value) || r.reg[1].MatchString(value) {
				_ = metric.SetValueUint8(instance, uint8(1))
			} else if r.hasDefault {
				_ = metric.SetValueUint8(instance, r.defaultValue)
			}
		}
	}
	return nil
}

// NewLabels returns the new labels the receiver creates
func (a *LabelAgent) NewLabels() []string {
	return a.newLabelNames
}
