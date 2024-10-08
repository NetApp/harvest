/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package max

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

type Max struct {
	*plugin.AbstractPlugin
	rules []*rule
}

func New(p *plugin.AbstractPlugin) *Max {
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

func (m *Max) Init() error {

	if err := m.AbstractPlugin.Init(); err != nil {
		return err
	}

	m.rules = make([]*rule, 0)
	if err := m.parseRules(); err != nil {
		return err
	}

	if len(m.rules) == 1 {
		m.SLogger.Debug("parsed 1 max rule")
	} else {
		m.SLogger.Debug("parsed max rules", slog.Int("count", len(m.rules)))
	}
	return nil
}

func (m *Max) parseRules() error {

	var err error

	for _, line := range m.Params.GetAllChildContentS() {

		r := rule{}

		fields := strings.Fields(line)
		if len(fields) == 2 || len(fields) == 1 {
			// parse label, possibly followed by value and object
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
						m.SLogger.Error(
							"rule compile regex",
							slogx.Err(err),
							slog.String("line", line),
						)
						return err
					}
				} else if value != "" {
					r.checkValue = value
				}

				if len(suffix) == 2 && suffix[1] != "" {
					r.object = strings.ToLower(suffix[1])
				}
			}
			if len(fields) == 2 {
				if strings.TrimSpace(fields[1]) == "..." {
					r.allLabels = true
				} else {
					r.includeLabels = strings.Split(fields[1], ",")
				}
			}
			m.rules = append(m.rules, &r)
			m.SLogger.Debug("parsed rule", slog.Any("rule", r))
		} else {
			m.SLogger.Warn("invalid rule syntax", slog.String("line", line))
			return errs.New(errs.ErrInvalidParam, "invalid rule")
		}
	}
	return nil
}

func (m *Max) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	data := dataMap[m.Object]
	matrices := make(map[string]*matrix.Matrix)

	// initialize cache
	for i, rule := range m.rules {

		for k := range data.GetMetrics() {

			key := strconv.Itoa(i) + k

			// Create matrix for each metric as each metric may have an instance with different label
			matrices[key] = data.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: true})

			matrices[key].RemoveExceptMetric(k)
			if rule.object != "" {
				matrices[key].Object = rule.object
			} else {
				matrices[key].Object = strings.ToLower(rule.label) + "_" + data.Object
			}
			// UUID needs to be unique
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
		objMetric       *matrix.Metric
		value           float64
		ok              bool
		err             error
	)

	for _, instance := range data.GetInstances() {

		if !instance.IsExportable() {
			continue
		}

		for i, rule := range m.rules {

			if objName = instance.GetLabel(rule.label); objName == "" {
				m.SLogger.Warn("label name missing, skipped", slog.String("label", rule.label))
				continue
			}

			if rule.checkLabel != "" {
				if rule.checkRegex != nil {
					if !rule.checkRegex.MatchString(instance.GetLabel(rule.checkLabel)) {
						continue
					}
				} else if instance.GetLabel(rule.checkLabel) != rule.checkValue {
					continue
				}
			}

			objKey = objName
			for key, metric := range data.GetMetrics() {

				matrixKey := strconv.Itoa(i) + key

				if objInstance = matrices[matrixKey].GetInstance(objKey); objInstance == nil {
					rule.counts[objKey] = make(map[string]int)
					if objInstance, err = matrices[matrixKey].NewInstance(objKey); err != nil {
						return nil, nil, err
					}
				}

				if value, ok = metric.GetValueFloat64(instance); !ok {
					continue
				}

				if objMetric = matrices[matrixKey].GetMetric(key); objMetric == nil {
					m.SLogger.Warn("metric not found in cache", slog.String("key", key), slog.String("label", rule.label))
					continue
				}

				v, _ := objMetric.GetValueFloat64(objInstance)

				if value > v {
					if err = objMetric.SetValueFloat64(objInstance, value); err != nil {
						m.SLogger.Error(
							"add value",
							slogx.Err(err),
							slog.String("key", key),
							slog.String("objName", objName),
						)
					} else {
						switch {
						case rule.allLabels:
							objInstance.SetLabels(instance.GetLabels())
						case len(rule.includeLabels) != 0:
							for _, k := range rule.includeLabels {
								objInstance.SetLabel(k, instance.GetLabel(k))
							}
							objInstance.SetLabel(rule.label, objName)
						default:
							objInstance.SetLabel(rule.label, objName)
						}
					}
				}

				rule.counts[objKey][key]++
			}
		}
	}

	matricesArray := make([]*matrix.Matrix, 0, len(matrices))

	for _, v := range matrices {
		matricesArray = append(matricesArray, v)
	}

	return matricesArray, nil, nil
}

// NewMetrics returns the new metrics the receiver creates
func (m *Max) NewMetrics() []plugin.DerivedMetric {
	derivedMetrics := make([]plugin.DerivedMetric, 0, len(m.rules))
	for _, r := range m.rules {
		derivedMetrics = append(derivedMetrics, plugin.DerivedMetric{Name: r.object, Source: r.label, IsMax: true})
	}

	return derivedMetrics
}
