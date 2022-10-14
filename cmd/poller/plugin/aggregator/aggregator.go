/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package aggregator

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"regexp"
	"strings"
)

type Aggregator struct {
	*plugin.AbstractPlugin
	rules []*rule
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Aggregator{AbstractPlugin: p}
}

type rule struct {
	label         string
	object        string
	checkLabel    string
	checkValue    string
	checkRegex    *regexp.Regexp
	includeLabels []string
	allLabels     bool
	counts        map[string]map[string]int
}

func (a *Aggregator) Init() error {

	if err := a.AbstractPlugin.Init(); err != nil {
		return err
	}

	a.rules = make([]*rule, 0)
	if err := a.parseRules(); err != nil {
		return err
	}

	if len(a.rules) == 1 {
		a.Logger.Debug().Msg("parsed 1 aggregation rule")
	} else {
		a.Logger.Debug().Msgf("parsed %d aggregation rules", len(a.rules))
	}
	return nil
}

func (a *Aggregator) parseRules() error {

	var err error

	for _, line := range a.Params.GetAllChildContentS() {

		a.Logger.Trace().Msgf("parsing raw rule: [%s]", line)

		r := rule{}

		fields := strings.Fields(line)
		if len(fields) == 2 || len(fields) == 1 {
			// parse label, possibly followed by value and object
			a.Logger.Trace().Msgf("handling first field: [%s]", fields[0])
			prefix := strings.SplitN(fields[0], "<", 2)
			r.label = strings.TrimSpace(prefix[0])
			if len(prefix) == 2 {
				// rule part in <>
				suffix := strings.SplitN(prefix[1], ">", 2)
				value := ""
				if s := strings.SplitN(suffix[0], "=", 2); len(s) == 2 {
					r.checkLabel = s[0]
					value = s[1]
				} else if s[0] != "" {
					r.checkLabel = r.label
					value = s[0]
				}

				if strings.HasPrefix(value, "`") {
					value = strings.TrimPrefix(strings.TrimSuffix(value, "`"), "`")
					if r.checkRegex, err = regexp.Compile(value); err != nil {
						a.Logger.Error().Stack().Err(err).Msgf("rule [%s]: compile regex:", line)
						return err
					}
					a.Logger.Trace().Msgf("parsed regex: [%s]", r.checkRegex.String())
				} else if value != "" {
					r.checkValue = value
				}

				if len(suffix) == 2 && suffix[1] != "" {
					r.object = strings.ToLower(suffix[1])
				}
			}
			if len(fields) == 2 {
				a.Logger.Trace().Msgf("handling second field: [%s]", fields[1])
				if strings.TrimSpace(fields[1]) == "..." {
					r.allLabels = true
				} else {
					r.includeLabels = strings.Split(fields[1], ",")
				}
			}
			a.rules = append(a.rules, &r)
			a.Logger.Debug().Msgf("parsed rule [%v]", r)
		} else {
			a.Logger.Warn().Msgf("invalid rule syntax [%s]", line)
			return errs.New(errs.ErrInvalidParam, "invalid rule")
		}
	}
	return nil
}

func (a *Aggregator) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	matrices := make([]*matrix.Matrix, len(a.rules))

	// initialize cache
	for i, rule := range a.rules {

		matrices[i] = data.Clone(false, true, false)
		if rule.object != "" {
			matrices[i].Object = rule.object
		} else {
			matrices[i].Object = strings.ToLower(rule.label) + "_" + data.Object
		}
		matrices[i].UUID += ".Aggregator"
		matrices[i].SetExportOptions(matrix.DefaultExportOptions())
		matrices[i].SetExportable(true)
		rule.counts = make(map[string]map[string]int)
	}

	// create instances and summarize metric values

	var (
		objName, objKey string
		objInstance     *matrix.Instance
		objMetric       matrix.Metric
		value           float64
		ok              bool
		pass            bool
		err             error
	)

	for _, instance := range data.GetInstances() {

		if !instance.IsExportable() {
			continue
		}

		a.Logger.Trace().Msgf("handling instance with labels [%s]", instance.GetLabels().String())

		for i, rule := range a.rules {

			a.Logger.Trace().Msgf("handling rule [%v]", rule)
			if objName = instance.GetLabel(rule.label); objName == "" {
				a.Logger.Warn().Msgf("label name for [%s] missing, skipped", rule.label)
				continue
			}

			if rule.checkLabel != "" {
				a.Logger.Trace().Msgf("checking label (%s => %s)....", rule.checkLabel, rule.checkValue)
				if rule.checkRegex != nil {
					if !rule.checkRegex.MatchString(instance.GetLabel(rule.checkLabel)) {
						continue
					}
				} else if instance.GetLabel(rule.checkLabel) != rule.checkValue {
					continue
				}
			}

			if rule.allLabels {
				objKey = strings.Join(instance.GetLabels().Values(), ".")
			} else if len(rule.includeLabels) != 0 {
				objKey = objName
				for _, k := range rule.includeLabels {
					objKey += "." + instance.GetLabel(k)
				}
			} else {
				objKey = objName
			}
			a.Logger.Trace().Msgf("instance (%s= %s): formatted key [%s]", rule.label, objName, objKey)

			if objInstance = matrices[i].GetInstance(objKey); objInstance == nil {
				rule.counts[objKey] = make(map[string]int)
				if objInstance, err = matrices[i].NewInstance(objKey); err != nil {
					return nil, err
				}
				if rule.allLabels {
					objInstance.SetLabels(instance.GetLabels())
				} else if len(rule.includeLabels) != 0 {
					for _, k := range rule.includeLabels {
						objInstance.SetLabel(k, instance.GetLabel(k))
					}
					objInstance.SetLabel(rule.label, objName)
				} else {
					objInstance.SetLabel(rule.label, objName)
				}
			}

			for key, metric := range data.GetMetrics() {

				if value, ok, pass = metric.GetValueFloat64(instance); !ok || !pass {
					continue
				}

				if objMetric = matrices[i].GetMetric(key); objMetric == nil {
					a.Logger.Warn().Msgf("metric [%s] not found in [%s] cache", key, rule.label)
					continue
				}

				if err = objMetric.AddValueFloat64(objInstance, value); err != nil {
					a.Logger.Error().Stack().Err(err).Msgf("add value [%s] [%s]:", key, objName)
				}

				rule.counts[objKey][key]++
			}
		}
	}

	// normalize values into averages if we are able to identify it as an percentage or average metric

	for i, m := range matrices {
		for mk, metric := range m.GetMetrics() {

			var (
				value   float64
				count   int
				ok, avg bool
				err     error
			)

			mn := metric.GetName()
			if metric.GetProperty() == "average" || metric.GetProperty() == "percent" {
				avg = true
			} else if strings.Contains(mn, "average_") || strings.Contains(mn, "avg_") {
				avg = true
			} else if !metric.IsHistogram() && strings.Contains(mn, "_latency") {
				avg = true
			}

			if !avg {
				continue
			}

			a.Logger.Trace().Msgf("[%s] (%s) normalizing values as average", mk, mn)

			for key, instance := range m.GetInstances() {

				if value, ok, pass = metric.GetValueFloat64(instance); !ok || !pass {
					continue
				}

				if count, ok = a.rules[i].counts[key][mk]; !ok {
					continue
				}

				if err = metric.SetValueFloat64(instance, value/float64(count)); err != nil {
					a.Logger.Error().Stack().Err(err).Msgf("set value [%s] [%s]:", mn, key)
				}
			}
		}
	}

	return matrices, nil
}
