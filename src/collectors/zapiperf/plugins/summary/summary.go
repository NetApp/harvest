package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"strings"
)

type Summary struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Summary{AbstractPlugin: p}
}

func (me *Summary) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	count := make(map[string]map[string]int)
	cache := make(map[string]*matrix.Matrix)
	labels := make(map[string][]string)

	// initialize cache
	for _, obj := range me.Params.GetAllChildContentS() {

		if fields := strings.Fields(obj); len(fields) > 1 {
			obj = fields[0]
			labels[obj] = fields[1:]
		}
		cache[obj] = data.Clone(false, true, false)
		cache[obj].Object = obj + "_" + data.Object
		cache[obj].Plugin = "Summary"
		cache[obj].SetExportOptions(matrix.DefaultExportOptions())
		count[obj] = make(map[string]int)

		if err := cache[obj].Reset(); err != nil {
			return nil, err
		}
	}

	// create instances and summarize metric values

	var (
		obj_name     string
		obj_instance *matrix.Instance
		obj_metric   matrix.Metric
		value        float64
		ok           bool
		err          error
	)

	for _, instance := range data.GetInstances() {

		for obj := range cache {

			if obj_name = instance.GetLabel(obj); obj_name == "" {
				logger.Warn(me.Prefix, "label name for [%s] missing, skipped", obj)
				continue
			}

			if obj_instance = cache[obj].GetInstance(obj_name); obj_instance == nil {
				if obj_instance, err = cache[obj].AddInstance(obj_name); err != nil {
					return nil, err
				} else {
					obj_instance.SetLabel(obj, obj_name)
					for _, k := range labels[obj] {
						obj_instance.SetLabel(k, instance.GetLabel(k))
					}
				}
			}

			count[obj][obj_name]++

			for key, metric := range data.GetMetrics() {

				if value, ok = metric.GetValueFloat64(instance); !ok {
					continue
				}

				if obj_metric = cache[obj].GetMetric(key); obj_metric == nil {
					logger.Warn(me.Prefix, "metric [%s] not found in [%s] cache", key, obj)
					continue
				}

				//logger.Debug(me.Prefix, "(%s) (%s) handling metric [%s] (%s)", obj, obj_name, key, obj_metric.GetName())
				//obj_metric.Print()

				if err = obj_metric.AddValueFloat64(obj_instance, value); err != nil {
					logger.Error(me.Prefix, "add value [%s] [%s]: %v", key, obj_name, err)
				}
			}
		}
	}

	// normalize values

	for obj := range cache {
		for _, metric := range cache[obj].GetMetrics() {

			var (
				value float64
				ok    bool
				err   error
			)

			if metric.GetProperty() != "average" && metric.GetProperty() != "percent" {
				continue
			}

			for key, instance := range cache[obj].GetInstances() {

				if value, ok = metric.GetValueFloat64(instance); !ok {
					continue
				}

				if err = metric.SetValueFloat64(instance, value/float64(count[obj][key])); err != nil {

					logger.Error(me.Prefix, "set value [%s] [%s]: %v", metric.GetName(), key, err)
				}
			}
		}
	}

	results := make([]*matrix.Matrix, len(cache))
	i := 0
	for _, c := range cache {
		results[i] = c
		i++
	}

	return results, nil

}
