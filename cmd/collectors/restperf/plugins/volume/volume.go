package volume

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"regexp"
	"sort"
	"strings"
)

type Volume struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (me *Volume) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		err error
	)
	opsKeyPrefix := "temp_"

	re := regexp.MustCompile(`^(.*)__(\d{4})$`)

	flexgroupAggrsMap := make(map[string]*set.Set)
	// new metric would be volume_aggr_labels
	metricName := "labels"
	volumeAggrmetric := matrix.New(".Volume", "volume_aggr", "volume_aggr")
	volumeAggrmetric.SetGlobalLabels(data.GetGlobalLabels())

	metric, err := volumeAggrmetric.NewMetricFloat64(metricName)
	if err != nil {
		me.Logger.Error().Err(err).Msg("add metric")
		return nil, err
	}
	me.Logger.Trace().Msgf("added metric: (%s) %v", metricName, metric)

	cache := data.Clone(false, true, false)
	cache.UUID += ".Volume"

	// create flexgroup instance cache
	for _, i := range data.GetInstances() {
		if match := re.FindStringSubmatch(i.GetLabel("volume")); len(match) == 3 {
			// instance key is svm.flexgroup-volume
			key := i.GetLabel("svm") + "." + match[1]
			if cache.GetInstance(key) == nil {
				fg, _ := cache.NewInstance(key)
				fg.SetLabels(i.GetLabels().Copy())
				fg.SetLabel("volume", match[1])
				// Flexgroup don't show any aggregate, node
				fg.SetLabel("aggr", "")
				fg.SetLabel("node", "")
				fg.SetLabel("style", "flexgroup")
			}

			if volumeAggrmetric.GetInstance(key) == nil {
				flexgroupInstance, _ := volumeAggrmetric.NewInstance(key)
				flexgroupInstance.SetLabels(i.GetLabels().Copy())
				flexgroupInstance.SetLabel("volume", match[1])
				// Flexgroup don't show any node
				flexgroupInstance.SetLabel("node", "")
				flexgroupInstance.SetLabel("style", "flexgroup")
				flexgroupAggrsMap[key] = set.New()
				if err := metric.SetValueFloat64(flexgroupInstance, 1); err != nil {
					me.Logger.Error().Err(err).Str("metric", metricName).Msg("Unable to set value on metric")
				}
			}
			flexgroupAggrsMap[key].Add(i.GetLabel("aggr"))
			i.SetLabel("style", "flexgroup_constituent")
			i.SetExportable(false)
		} else {
			i.SetLabel("style", "flexvol")
			key := i.GetLabel("svm") + "." + i.GetLabel("volume")
			flexvolInstance, err := volumeAggrmetric.NewInstance(key)
			if err != nil {
				me.Logger.Error().Err(err).Str("key", key).Msg("Failed to create new instance")
				continue
			}
			flexvolInstance.SetLabels(i.GetLabels().Copy())
			flexvolInstance.SetLabel("style", "flexvol")
			if err := metric.SetValueFloat64(flexvolInstance, 1); err != nil {
				me.Logger.Error().Err(err).Str("metric", metricName).Msg("Unable to set value on metric")
			}
		}
	}

	me.Logger.Debug().Int("flexgroup volume count", len(cache.GetInstances())).Msg("")

	//cache.Reset()

	// create summary
	for _, i := range data.GetInstances() {
		if match := re.FindStringSubmatch(i.GetLabel("volume")); len(match) == 3 {
			// instance key is svm.flexgroup-volume
			key := i.GetLabel("svm") + "." + match[1]

			// set aggrs label for flexgroup in new metrics
			flexgroupInstance := volumeAggrmetric.GetInstance(key)
			if flexgroupInstance != nil {
				// make sure the order of aggregate is same for each poll
				aggrs := flexgroupAggrsMap[key].Values()
				sort.Strings(aggrs)
				flexgroupInstance.SetLabel("aggr", strings.Join(aggrs, ","))
			}

			fg := cache.GetInstance(key)
			if fg == nil {
				me.Logger.Error().Msgf("instance [%s] not in local cache", key)
				continue
			}

			for mkey, m := range data.GetMetrics() {

				if !m.IsExportable() && m.GetType() != "float64" {
					continue
				}

				fgm := cache.GetMetric(mkey)
				if fgm == nil {
					me.Logger.Error().Msgf("metric [%s] not in local cache", mkey)
					continue
				}

				me.Logger.Trace().Msgf("(%s) handling metric (%s)", fg.GetLabel("volume"), mkey)

				if value, ok := m.GetValueFloat64(i); ok {

					fgv, _ := fgm.GetValueFloat64(fg)

					// non-latency metrics: simple sum
					if !strings.HasSuffix(m.GetName(), "_latency") {

						err := fgm.SetValueFloat64(fg, fgv+value)
						if err != nil {
							me.Logger.Error().Err(err).Msg("error")
						}
						// just for debugging
						fgv2, _ := fgm.GetValueFloat64(fg)

						me.Logger.Trace().Msgf("   > simple increment %f + %f = %f", fgv, value, fgv2)
						continue
					}

					// latency metric: weighted sum
					opsKey := ""
					if strings.Contains(mkey, "_latency") {
						opsKey = m.GetComment()
					}
					me.Logger.Trace().Msgf("    > weighted increment <%s * %s>", mkey, opsKey)

					if ops := data.GetMetric(opsKey); ops != nil {
						if opsValue, ok := ops.GetValueFloat64(i); ok {
							var tempOpsV float64

							prod := value * opsValue
							tempOpsKey := opsKeyPrefix + opsKey
							// Create temp ops metrics. These should not be exported.
							// A base counter can be part base for multiple metrics hence we should not be changing base counter for weighted average calculation
							tempOps := cache.GetMetric(tempOpsKey)

							if tempOps == nil {
								if tempOps, err = cache.NewMetricFloat64(tempOpsKey); err != nil {
									return nil, err
								}
								tempOps.SetExportable(false)
							} else {
								tempOpsV, _ = tempOps.GetValueFloat64(fg)
							}
							// If latency value is 0 then it's ops value is not used in weighted average calculation
							if value != 0 {
								err = tempOps.SetValueFloat64(fg, tempOpsV+opsValue)
								if err != nil {
									me.Logger.Error().Err(err).Msg("error")
								}
							}
							err = fgm.SetValueFloat64(fg, fgv+prod)
							if err != nil {
								me.Logger.Error().Err(err).Msg("error")
							}

							// debugging
							fgv2, _ := fgm.GetValueFloat64(fg)

							me.Logger.Trace().Msgf("       %f + (%f * %f) (=%f) = %f", fgv, value, opsValue, prod, fgv2)
						} else {
							me.Logger.Trace().Msg("       no ops value SKIP")
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

					opsKey := ""
					if strings.Contains(mkey, "_latency") {
						opsKey = m.GetComment()
					}

					if ops := cache.GetMetric(opsKeyPrefix + opsKey); ops != nil {

						if opsValue, ok := ops.GetValueFloat64(i); ok && opsValue != 0 {
							err := m.SetValueFloat64(i, value/opsValue)
							if err != nil {
								me.Logger.Error().Err(err).Msgf("error")
							}
						} else {
							m.SetValueNAN(i)
						}
					}
				}

			}
		}
	}

	return []*matrix.Matrix{cache, volumeAggrmetric}, nil
}
