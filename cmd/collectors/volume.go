package collectors

import (
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/util"
	"maps"
	"regexp"
	"sort"
	"strings"
)

var flexgroupRegex = regexp.MustCompile(`^(.*)__(\d{4})$`)

func ProcessFlexGroupData(logger *logging.Logger, data *matrix.Matrix, style string, includeConstituents bool, opsKeyPrefix string) ([]*matrix.Matrix, *util.Metadata, error) {
	var err error

	fgAggrMap := make(map[string]*set.Set)
	flexgroupAggrsMap := make(map[string]*set.Set)

	metricName := "labels"
	volumeAggrmetric := matrix.New(".Volume", "volume_aggr", "volume_aggr")
	volumeAggrmetric.SetGlobalLabels(data.GetGlobalLabels())

	metric, err := volumeAggrmetric.NewMetricFloat64(metricName)
	if err != nil {
		logger.Error().Err(err).Msg("add metric")
		return nil, nil, err
	}

	cache := data.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: true})
	cache.UUID += ".Volume"

	for _, i := range data.GetInstances() {
		if match := flexgroupRegex.FindStringSubmatch(i.GetLabel("volume")); len(match) == 3 {
			key := i.GetLabel("svm") + "." + match[1]
			if cache.GetInstance(key) == nil {
				fg, _ := cache.NewInstance(key)
				fg.SetLabels(maps.Clone(i.GetLabels()))
				fg.SetLabel("volume", match[1])
				fg.SetLabel("node", "")
				fg.SetLabel(style, "flexgroup")
				fgAggrMap[key] = set.New()
			}

			if volumeAggrmetric.GetInstance(key) == nil {
				flexgroupInstance, _ := volumeAggrmetric.NewInstance(key)
				flexgroupInstance.SetLabels(maps.Clone(i.GetLabels()))
				flexgroupInstance.SetLabel("volume", match[1])
				flexgroupInstance.SetLabel("node", "")
				flexgroupInstance.SetLabel(style, "flexgroup")
				flexgroupAggrsMap[key] = set.New()
				if err := metric.SetValueFloat64(flexgroupInstance, 1); err != nil {
					logger.Error().Err(err).Str("metric", metricName).Msg("Unable to set value on metric")
				}
			}
			fgAggrMap[key].Add(i.GetLabel("aggr"))
			flexgroupAggrsMap[key].Add(i.GetLabel("aggr"))
			i.SetLabel(style, "flexgroup_constituent")
			i.SetExportable(includeConstituents)
		} else {
			i.SetLabel(style, "flexvol")
			key := i.GetLabel("svm") + "." + i.GetLabel("volume")
			flexvolInstance, err := volumeAggrmetric.NewInstance(key)
			if err != nil {
				logger.Error().Err(err).Str("key", key).Msg("Failed to create new instance")
				continue
			}
			flexvolInstance.SetLabels(maps.Clone(i.GetLabels()))
			flexvolInstance.SetLabel(style, "flexvol")
			if err := metric.SetValueFloat64(flexvolInstance, 1); err != nil {
				logger.Error().Err(err).Str("metric", metricName).Msg("Unable to set value on metric")
			}
		}
	}

	logger.Debug().Int("flexgroup volume count", len(cache.GetInstances())).Send()

	recordFGFalse := make(map[string]*set.Set)
	for _, i := range data.GetInstances() {
		match := flexgroupRegex.FindStringSubmatch(i.GetLabel("volume"))
		if len(match) != 3 {
			continue
		}
		key := i.GetLabel("svm") + "." + match[1]

		flexgroupInstance := volumeAggrmetric.GetInstance(key)
		if flexgroupInstance != nil {
			aggrs := flexgroupAggrsMap[key].Values()
			sort.Strings(aggrs)
			flexgroupInstance.SetLabel("aggr", strings.Join(aggrs, ","))
		}

		fg := cache.GetInstance(key)
		if fg == nil {
			logger.Error().Msgf("instance [%s] not in local cache", key)
			continue
		}

		aggrs := fgAggrMap[key].Values()
		sort.Strings(aggrs)
		fg.SetLabel("aggr", strings.Join(aggrs, ","))

		for mkey, m := range data.GetMetrics() {
			if !m.IsExportable() && m.GetType() != "float64" {
				continue
			}

			fgm := cache.GetMetric(mkey)
			if fgm == nil {
				logger.Error().Msgf("metric [%s] not in local cache", mkey)
				continue
			}

			if value, ok := m.GetValueFloat64(i); ok {
				fgv, _ := fgm.GetValueFloat64(fg)

				if !strings.HasSuffix(m.GetName(), "_latency") {
					err := fgm.SetValueFloat64(fg, fgv+value)
					if err != nil {
						logger.Error().Err(err).Msg("error")
					}
					continue
				}

				opsKey := ""
				if strings.Contains(mkey, "_latency") {
					opsKey = m.GetComment()
				}

				if ops := data.GetMetric(opsKey); ops != nil {
					if opsValue, ok := ops.GetValueFloat64(i); ok {
						var tempOpsV float64

						prod := value * opsValue
						tempOpsKey := opsKeyPrefix + opsKey
						tempOps := cache.GetMetric(tempOpsKey)

						if tempOps == nil {
							if tempOps, err = cache.NewMetricFloat64(tempOpsKey); err != nil {
								return nil, nil, err
							}
							tempOps.SetExportable(false)
						} else {
							tempOpsV, _ = tempOps.GetValueFloat64(fg)
						}
						if value != 0 {
							err = tempOps.SetValueFloat64(fg, tempOpsV+opsValue)
							if err != nil {
								logger.Error().Err(err).Msg("error")
							}
						}
						err = fgm.SetValueFloat64(fg, fgv+prod)
						if err != nil {
							logger.Error().Err(err).Msg("error")
						}
					} else {
						s, ok := recordFGFalse[key]
						if !ok {
							s = set.New()
							recordFGFalse[key] = s
						}
						s.Add(fgm.GetName())
					}
				}
			} else {
				s, ok := recordFGFalse[key]
				if !ok {
					s = set.New()
					recordFGFalse[key] = s
				}
				s.Add(fgm.GetName())
			}
		}
	}

	for k, i := range cache.GetInstances() {
		if !i.IsExportable() {
			continue
		}
		for mkey, m := range cache.GetMetrics() {
			if mNames, ok := recordFGFalse[k]; ok {
				if mNames.Has(m.GetName()) {
					m.SetValueNAN(i)
					continue
				}
			}
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
								logger.Error().Err(err).Msgf("error")
							}
						} else {
							m.SetValueNAN(i)
						}
					}
				}
			}
		}
	}

	return []*matrix.Matrix{cache, volumeAggrmetric}, nil, nil
}
