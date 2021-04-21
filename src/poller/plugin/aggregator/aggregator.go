package aggregator

import (
	"goharvest2/poller/plugin"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"goharvest2/share/errors"
	"strings"
	"regexp"
)

type Aggregator struct {
	*plugin.AbstractPlugin
	rules []*rule
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Aggregator{AbstractPlugin: p}
}

type rule struct {
	label string
	object string
	checkLabel string
	checkValue string
	checkRegex *regexp.Regexp
	includeLabels []string
	allLabels bool
	counts map[string]map[string]int
}

func (me *Aggregator) Init() error {

	if err := me.AbstractPlugin.Init(); err != nil {
		return err
	}

	me.rules = make([]*rule, 0)
	if err := me.parseRules(); err != nil {
		return err
	}

	if len(me.rules) == 1 {
		logger.Debug(me.Prefix, "parsed 1 aggregation rule")
	} else {
		logger.Debug(me.Prefix, "parsed %d aggregation rules", len(me.rules))
	}
	return nil
}

func (me *Aggregator) parseRules() error {

	var err error

	for _, line := range me.Params.GetAllChildContentS() {

		logger.Trace(me.Prefix, "parsing raw rule: [%s]", line)

		r := rule{}

		fields := strings.Fields(line)
		if len(fields) == 2 || len(fields) == 1 {
			// parse label, possibly followed by value and object
			logger.Trace(me.Prefix, "handling first field: [%s]", fields[0])
			prefix := strings.SplitN(fields[0], "<", 2)
			r.label = strings.TrimSpace(prefix[0])
			if len(prefix) == 2 {
				// rule part in <>
				suffix := strings.SplitN(prefix[1], ">", 2)
				value := ""
				if s := strings.SplitN(suffix[0], "=", 2); len(s) == 2 {
					r.checkLabel = s[0]
					value = s[1]
				} else {
					r.checkLabel = r.label
					value = s[0]
				}

				if strings.HasPrefix(value, "`") {
					value = strings.TrimPrefix(strings.TrimSuffix(value, "`"), "`")
					if r.checkRegex, err = regexp.Compile(value); err != nil {
						logger.Error(me.Prefix, "rule [%s]: compile regex: %v", line, err)
						return err
					} else {
						logger.Trace(me.Prefix, "parsed regex: [%s]", r.checkRegex.String())
					}
				} else if value != "" {
					r.checkValue = value
				}

				if len(suffix) == 2 && suffix[1] != "" {
					r.object = strings.ToLower(suffix[1])
				}
			}
			if len(fields) == 2 {
				logger.Trace(me.Prefix, "handling second field: [%s]", fields[1])
				if strings.TrimSpace(fields[1]) == "..." {
					r.allLabels = true
				} else {
					r.includeLabels = strings.Split(fields[1], ",")
				}
			}
			me.rules = append(me.rules, &r)
			logger.Debug(me.Prefix, "parsed rule [%v]", r)
		} else {
			logger.Warn(me.Prefix, "invalid rule syntax [%s]", line)
			return errors.New(errors.INVALID_PARAM, "invalid rule")
		}
	}
	return nil
}

func (me *Aggregator) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	matrices := make([]*matrix.Matrix, len(me.rules))

	// initialize cache
	for i, rule := range me.rules {

		matrices[i] = data.Clone(false, true, false)
		if rule.object != "" {
			matrices[i].Object = rule.object
		} else {
			matrices[i].Object = strings.ToLower(rule.label) + "_" + data.Object
		}
		matrices[i].UUID += ".Aggregator"
		matrices[i].SetExportOptions(matrix.DefaultExportOptions())
		rule.counts = make(map[string]map[string]int)
	}

	// create instances and summarize metric values

	var (
		obj_name, obj_key string
		obj_instance *matrix.Instance
		obj_metric   matrix.Metric
		value        float64
		ok           bool
		err          error
	)

	for _, instance := range data.GetInstances() {

		logger.Trace(me.Prefix, "handling instance with labels [%s]", instance.GetLabels().String())

		for i, rule := range me.rules {
			if obj_name = instance.GetLabel(rule.label); obj_name == "" {
				logger.Warn(me.Prefix, "label name for [%s] missing, skipped", rule.label)
				continue
			}

			if rule.checkLabel != "" {
				if rule.checkRegex != nil {
					if ! rule.checkRegex.MatchString(instance.GetLabel(rule.checkLabel)) {
						continue
					}
				} else if instance.GetLabel(rule.checkLabel) != rule.checkValue {
					continue
				}
			}
			
			if rule.allLabels {
				obj_key = strings.Join(instance.GetLabels().Values(), ".")
			} else if len(rule.includeLabels) != 0 {
				obj_key = obj_name
				for _, k := range rule.includeLabels {
					obj_key += "."+instance.GetLabel(k)
				}
			} else {
				obj_key = obj_name
			}
			logger.Trace(me.Prefix, "instance (%s= %s): formatted key [%s]", rule.label, obj_name, obj_key)

			if obj_instance = matrices[i].GetInstance(obj_key); obj_instance == nil {
				rule.counts[obj_key] = make(map[string]int)
				if obj_instance, err = matrices[i].NewInstance(obj_key); err != nil {
					return nil, err
				}
				if rule.allLabels {
					obj_instance.SetLabels(instance.GetLabels())
				} else if len(rule.includeLabels) != 0 {
					for _, k := range rule.includeLabels {
						obj_instance.SetLabel(k, instance.GetLabel(k))
					}
					obj_instance.SetLabel(rule.label, obj_name)
				} else {
					obj_instance.SetLabel(rule.label, obj_name)
				}
			}

			for key, metric := range data.GetMetrics() {

				if value, ok = metric.GetValueFloat64(instance); ! ok {
					continue
				}

				if obj_metric = matrices[i].GetMetric(key); obj_metric == nil {
					logger.Warn(me.Prefix, "metric [%s] not found in [%s] cache", key, rule.label)
					continue
				}

				//logger.Debug(me.Prefix, "(%s) (%s) handling metric [%s] (%s)", obj, obj_name, key, obj_metric.GetName())
				//obj_metric.Print()

				if err = obj_metric.AddValueFloat64(obj_instance, value); err != nil {
					logger.Error(me.Prefix, "add value [%s] [%s]: %v", key, obj_name, err)
				}

				rule.counts[obj_key][key]++
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
			} else if strings.Contains(mn, "_latency") {
				avg = true
			}

			if !avg {
				continue
			}

			logger.Debug(me.Prefix, "[%s] (%s) normalizing values as average", mk, mn)

			for key, instance := range m.GetInstances() {

				if value, ok = metric.GetValueFloat64(instance); !ok {
					continue
				}

				if count, ok = me.rules[i].counts[key][mk]; !ok {
					continue
				}

				if err = metric.SetValueFloat64(instance, value/float64(count)); err != nil {
					logger.Error(me.Prefix, "set value [%s] [%s]: %v", mn, key, err)
				}
			}
		}
	}

	return matrices, nil
}
