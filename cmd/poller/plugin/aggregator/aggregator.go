/*
/*
 * Copyright NetApp Inc, 2021 All rights reserved
*/

package aggregator

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strings"
)

type Aggregator struct {
	*plugin.AbstractPlugin
	rules []*rule
}

func New(p *plugin.AbstractPlugin) *Aggregator {
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
	counts        map[string]map[string]float64
}

func (a *Aggregator) Init(remote conf.Remote) error {

	if err := a.AbstractPlugin.Init(remote); err != nil {
		return err
	}

	a.rules = make([]*rule, 0)
	if err := a.parseRules(); err != nil {
		return err
	}

	a.SLogger.Debug("parsed aggregation rules", slog.Int("numRules", len(a.rules)))
	return nil
}

func (a *Aggregator) parseRules() error {

	var err error

	for _, line := range a.Params.GetAllChildContentS() {

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
						a.SLogger.Error("ignore rule", slogx.Err(err))
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
			a.rules = append(a.rules, &r)
			a.SLogger.Debug("parsed rule", slog.String("label", r.label), slog.String("object", r.object))
		} else {
			return errs.New(errs.ErrInvalidParam, "invalid rule syntax "+line)
		}
	}
	return nil
}

func (a *Aggregator) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[a.Object]
	matrices := make([]*matrix.Matrix, len(a.rules))

	// initialize cache
	for i, rule := range a.rules {
		matrices[i] = data.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: true})
		if rule.object != "" {
			matrices[i].Object = rule.object
		} else {
			matrices[i].Object = strings.ToLower(rule.label) + "_" + data.Object
		}
		matrices[i].UUID += ".Aggregator"
		matrices[i].SetExportOptions(matrix.DefaultExportOptions())
		matrices[i].SetExportable(true)
		rule.counts = make(map[string]map[string]float64)
	}

	// create instances and summarize metric values
	var (
		objName, objKey string
		objInstance     *matrix.Instance
		objMetric       *matrix.Metric
		opsMetric       *matrix.Metric
		opsValue        float64
		value           float64
		ok              bool
		err             error
	)

	metadata := &util.Metadata{}

	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}

		for i, rule := range a.rules {
			if objName = instance.GetLabel(rule.label); objName == "" {
				a.SLogger.Warn("label missing, skipped", slog.String("label", rule.label))
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

			switch {
			case rule.allLabels:
				values := slices.Collect(maps.Keys(instance.GetLabels()))
				objKey = strings.Join(values, ".")
			case len(rule.includeLabels) != 0:
				objKey = objName
				for _, k := range rule.includeLabels {
					objKey += "." + instance.GetLabel(k)
				}
			default:
				objKey = objName
			}

			if objInstance = matrices[i].GetInstance(objKey); objInstance == nil {
				rule.counts[objKey] = make(map[string]float64)
				if objInstance, err = matrices[i].NewInstance(objKey); err != nil {
					return nil, nil, err
				}
				metadata.PluginInstances++
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

			for key, metric := range data.GetMetrics() {
				if value, ok = metric.GetValueFloat64(instance); !ok {
					continue
				}

				if objMetric = matrices[i].GetMetric(key); objMetric == nil {
					a.SLogger.Warn("metric not found in cache", slog.String("metric", key), slog.String("label", rule.label))
					continue
				}

				// latency metric: weighted sum
				if strings.Contains(key, "_latency") {
					opsKey := objMetric.GetComment()
					if opsKey != "" {
						if opsMetric = data.GetMetric(opsKey); opsMetric == nil {
							a.SLogger.Warn(
								"metric not found in response",
								slog.String("metric", opsKey),
								slog.String("label", rule.label),
							)
							continue
						}
						if opsValue, ok = opsMetric.GetValueFloat64(instance); !ok {
							continue
						}
						objMetric.AddValueFloat64(objInstance, opsValue*value)
						rule.counts[objKey][key] += opsValue
					}
				} else {
					objMetric.AddValueFloat64(objInstance, value)
					rule.counts[objKey][key]++
				}
			}
		}
	}

	// normalize values into averages if we are able to identify it as a percentage or average metric
	for i, m := range matrices {
		for mk, metric := range m.GetMetrics() {
			var (
				v       float64
				count   float64
				ok, avg bool
			)

			mn := metric.GetName()
			switch {
			case metric.GetProperty() == "average" || metric.GetProperty() == "percent":
				avg = true
			case strings.Contains(mn, "average_") || strings.Contains(mn, "avg_"):
				avg = true
			case !metric.IsHistogram() && strings.Contains(mn, "_latency"):
				avg = true
			}

			if !avg {
				continue
			}

			for key, instance := range m.GetInstances() {
				if v, ok = metric.GetValueFloat64(instance); !ok {
					continue
				}

				if count, ok = a.rules[i].counts[key][mk]; !ok {
					continue
				}

				// if no ops happened
				if count == 0 {
					metric.SetValueFloat64(instance, 0)
				} else {
					metric.SetValueFloat64(instance, v/count)
				}
			}
		}
	}

	return matrices, metadata, nil
}

// NewLabels returns the new labels the receiver creates
func (a *Aggregator) NewLabels() []string {
	var newLabelNames []string
	for _, r := range a.rules {
		newLabelNames = append(newLabelNames, r.includeLabels...)
	}

	return newLabelNames
}

// NewMetrics returns the new metrics the receiver creates
func (a *Aggregator) NewMetrics() []plugin.DerivedMetric {
	derivedMetrics := make([]plugin.DerivedMetric, 0, len(a.rules))
	for _, r := range a.rules {
		customName := r.object != ""
		derivedMetrics = append(derivedMetrics, plugin.DerivedMetric{Name: r.label, Source: r.object, HasCustomName: customName})
	}

	return derivedMetrics
}
