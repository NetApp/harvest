/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package max

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"regexp"
	"strconv"
	"strings"
)

type Max struct {
	*plugin.AbstractPlugin
	rules []*rule
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Max{AbstractPlugin: p}
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

func (a *Max) Init() error {

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

func (a *Max) parseRules() error {

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

func (a *Max) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	matrices := make(map[string]*matrix.Matrix)

	// initialize cache
	for i, rule := range a.rules {

		for k := range data.GetMetrics() {

			key := strconv.Itoa(i) + k

			//Create matrix for each metric as each metric may have an instance with different label
			matrices[key] = data.Clone(false, true, false)
			matrices[key].RemoveExceptMetric(k)
			if rule.object != "" {
				matrices[key].Object = rule.object
			} else {
				matrices[key].Object = strings.ToLower(rule.label) + "_" + data.Object
			}
			//UUID needs to be unique
			matrices[key].UUID += key
			matrices[key].SetExportOptions(matrix.DefaultExportOptions())
			matrices[key].SetExportable(true)
			rule.counts = make(map[string]map[string]int)
		}
	}

	// create instances and summarize metric values

	var (
		objName, objKey string
		objInstance     *matrix.Instance
		objMetric       matrix.Metric
		value           float64
		ok              bool
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

			objKey = objName
			a.Logger.Trace().Msgf("instance (%s= %s): formatted key [%s]", rule.label, objName, objKey)

			for key, metric := range data.GetMetrics() {

				matrixKey := strconv.Itoa(i) + key

				if objInstance = matrices[matrixKey].GetInstance(objKey); objInstance == nil {
					rule.counts[objKey] = make(map[string]int)
					if objInstance, err = matrices[matrixKey].NewInstance(objKey); err != nil {
						return nil, err
					}
				}

				if value, ok = metric.GetValueFloat64(instance); !ok {
					continue
				}

				if objMetric = matrices[matrixKey].GetMetric(key); objMetric == nil {
					a.Logger.Warn().Msgf("metric [%s] not found in [%s] cache", key, rule.label)
					continue
				}

				m, _ := objMetric.GetValueFloat64(objInstance)

				if value > m {
					if err = objMetric.SetValueFloat64(objInstance, value); err != nil {
						a.Logger.Error().Stack().Err(err).Msgf("add value [%s] [%s]:", key, objName)
					} else {
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
				}

				rule.counts[objKey][key]++
			}
		}
	}

	var matricesArray []*matrix.Matrix

	for _, v := range matrices {
		matricesArray = append(matricesArray, v)
	}

	return matricesArray, nil
}
