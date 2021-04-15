package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"regexp"
	"strings"
)

type Volume struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

//@TODO cleanup logging
//@TODO rewrite using vector arithmetic
// will simplify the code a whole!!!

func (me *Volume) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	re := regexp.MustCompile(`^(.*)__(\d{4})$`)

	cache := data.Clone(false, true, false)
	cache.Plugin = "volume.flexgroup"

	// create flexgroup instance cache
	for _, i := range data.GetInstances() {
		if match := re.FindStringSubmatch(i.GetLabel("volume")); match != nil && len(match) == 3 {
			key := i.GetLabel("node") + "." + i.GetLabel("svm") + "." + match[1]
			if cache.GetInstance(key) == nil {
				fg, _ := cache.AddInstance(key)
				fg.SetLabels(i.GetLabels().Copy())
				fg.SetLabel("volume", match[1])
				fg.SetLabel("type", "flexgroup")
			}
			i.SetLabel("type", "flexgroup_constituent")
			i.SetExportable(false)
		} else {
			i.SetLabel("type", "flexvol")
		}
	}

	logger.Debug(me.Prefix, "extracted %d flexgroup volumes", cache.SizeInstances())

	if err := cache.Reset(); err != nil {
		return nil, err
	}

	// create summary
	for _, i := range data.GetInstances() {
		if match := re.FindStringSubmatch(i.GetLabel("volume")); match != nil && len(match) == 3 {
			key := i.GetLabel("node") + "." + i.GetLabel("svm") + "." + match[1]
			fg := cache.GetInstance(key)
			if fg == nil {
				logger.Error(me.Prefix, "instance [%s] not in local cache", key)
				continue
			}

			for mkey, m := range data.GetMetrics() {

				if !m.IsExportable() && m.GetType() != "float64" {
					continue
				}

				fgm := cache.GetMetric(mkey)
				if fgm == nil {
					logger.Error(me.Prefix, "metric [%s] not in local cache", mkey)
					continue
				}

				logger.Trace(me.Prefix, "(%s) handling metric (%s)", fg.GetLabel("volume"), mkey)

				if value, ok := m.GetValueFloat64(i); ok {

					fgv, _ := fgm.GetValueFloat64(fg)

					// non-latency metrics: simple sum
					if !strings.HasSuffix(m.GetName(), "_latency") {

						fgm.SetValueFloat64(fg, fgv+value)
						// just for debugging
						fgv2, _ := fgm.GetValueFloat64(fg)

						logger.Trace(me.Prefix, "   > simple increment %f + %f = %f", fgv, value, fgv2)
						continue
					}

					// latency metric: weighted sum
					ops_key := strings.Replace(mkey, "avg_latency", "total_ops", 1)
					ops_key = strings.Replace(mkey, "_latency", "_ops", 1)
					logger.Trace(me.Prefix, "    > weighted increment <%s * %s>", mkey, ops_key)

					if ops := data.GetMetric(ops_key); ops != nil {
						if ops_value, ok := ops.GetValueFloat64(i); ok {

							prod := value * ops_value
							fgm.SetValueFloat64(fg, fgv+prod)

							// debugging
							fgv2, _ := fgm.GetValueFloat64(fg)

							logger.Trace(me.Prefix, "       %f + (%f * %f) (=%f) = %f", fgv, value, ops_value, prod, fgv2)
						} else {
							logger.Trace(me.Prefix, "       no ops value SKIP")
						}
					}

				}

			}
		}
	}

	// normalize latency values
	for _, i := range cache.GetInstances() {
		for mkey, m := range cache.GetMetrics() {
			if m.IsExportable() && strings.HasSuffix(m.GetName(), "_latency") {

				if value, ok := m.GetValueFloat64(i); ok {

					ops_key := strings.Replace(mkey, "avg_latency", "total_ops", 1)
					ops_key = strings.Replace(mkey, "_latency", "_ops", 1)

					if ops := cache.GetMetric(ops_key); ops != nil {

						if ops_value, ok := ops.GetValueFloat64(i); ok && ops_value != 0 {
							m.SetValueFloat64(i, value/ops_value)
						} else {
							m.SetValueNAN(i)
						}
					}
				}

			}
		}
	}

	return []*matrix.Matrix{cache}, nil
}
