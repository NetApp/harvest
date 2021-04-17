package aggregator

import (
	"goharvest2/poller/plugin"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"strings"
)

type Aggregator struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Aggregator{AbstractPlugin: p}
}

type cache struct {
    matrix *matrix.Matrix
    counts map[string]map[string]int
    labels []string
}

func (me *Aggregator) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

    all := make(map[string]*cache, 0)

	// initialize cache
	for _, obj := range me.Params.GetAllChildContentS() {

        c := &cache{}

		if fields := strings.Fields(obj); len(fields) > 1 {
			obj = fields[0]
			c.labels = fields[1:]
		}
        all[obj] = c

		c.matrix = data.Clone(false, true, false)
		c.matrix.Object = obj + "_" + data.Object
		c.matrix.Plugin = "Aggregator"
		c.matrix.SetExportOptions(matrix.DefaultExportOptions())
		c.counts = make(map[string]map[string]int)

		if err := c.matrix.Reset(); err != nil {
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

        logger.Trace(me.Prefix, "handling instance with labels [%s]", instance.GetLabels().String())

		for obj, c := range all {
			if obj_name = instance.GetLabel(obj); obj_name == "" {
				logger.Warn(me.Prefix, "label name for [%s] missing, skipped", obj)
				continue
			}

			if obj_instance = c.matrix.GetInstance(obj_name); obj_instance == nil {
                c.counts[obj_name] = make(map[string]int)
				if obj_instance, err = c.matrix.AddInstance(obj_name); err != nil {
					return nil, err
				} else {
					obj_instance.SetLabel(obj, obj_name)
					for _, k := range c.labels {
						obj_instance.SetLabel(k, instance.GetLabel(k))
					}
				}
			}

			for key, metric := range data.GetMetrics() {

				if value, ok = metric.GetValueFloat64(instance); ! ok {
					continue
				}

				if obj_metric = c.matrix.GetMetric(key); obj_metric == nil {
					logger.Warn(me.Prefix, "metric [%s] not found in [%s] cache", key, obj)
					continue
				}

				//logger.Debug(me.Prefix, "(%s) (%s) handling metric [%s] (%s)", obj, obj_name, key, obj_metric.GetName())
				//obj_metric.Print()

				if err = obj_metric.AddValueFloat64(obj_instance, value); err != nil {
					logger.Error(me.Prefix, "add value [%s] [%s]: %v", key, obj_name, err)
				}

                c.counts[obj_name][key]++
			}
		}
	}

	// normalize values into averages if we are able to identify it as an percentage or average metric

	for _, c := range all {
		for mk, metric := range c.matrix.GetMetrics() {

			var (
				value float64
                count int
				ok, avg bool
				err   error
			)

            mn := metric.GetName()
			if metric.GetProperty() == "average" || metric.GetProperty() == "percent" {
				avg = true
			} else if strings.Contains(mn, "average_") || strings.Contains(mn, "avg_") {
                avg = true
            } else if strings.Contains(mn, "_latency") {
                avg = true
            }

            if ! avg {
                continue
            }

            logger.Debug(me.Prefix, "[%s] (%s) normalizing values as average", mk, mn)

			for key, instance := range c.matrix.GetInstances() {

				if value, ok = metric.GetValueFloat64(instance); ! ok {
					continue
				}

                if count, ok = c.counts[key][mk]; ! ok {
                    continue
                }

				if err = metric.SetValueFloat64(instance, value/float64(count)); err != nil {
					logger.Error(me.Prefix, "set value [%s] [%s]: %v", mn, key, err)
				}
			}
		}
	}

	results := make([]*matrix.Matrix, len(all))
    i := 0
	for _, c := range all {
		results[i] = c.matrix
        i++
	}

	return results, nil
}
