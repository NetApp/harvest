package collectors

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"maps"
	"regexp"
	"sort"
	"strings"
)

var flexgroupRegex = regexp.MustCompile(`^(.*)__(\d{4})$`)

func ProcessFlexGroupData(logger *slog.Logger, data *matrix.Matrix, style string, includeConstituents bool, opsKeyPrefix string) ([]*matrix.Matrix, *util.Metadata, error) {
	var err error

	fgAggrMap := make(map[string]*set.Set)
	flexgroupAggrsMap := make(map[string]*set.Set)

	metricName := "labels"
	volumeAggrmetric := matrix.New(".Volume", "volume_aggr", "volume_aggr")
	volumeAggrmetric.SetGlobalLabels(data.GetGlobalLabels())

	metric, err := volumeAggrmetric.NewMetricFloat64(metricName)
	if err != nil {
		logger.Error("add metric", slogx.Err(err), slog.String("key", metricName))
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
				fg.SetLabel("uuid", "")
				fg.SetLabel(style, "flexgroup")
				fgAggrMap[key] = set.New()
			}

			if volumeAggrmetric.GetInstance(key) == nil {
				flexgroupInstance, _ := volumeAggrmetric.NewInstance(key)
				flexgroupInstance.SetLabels(maps.Clone(i.GetLabels()))
				flexgroupInstance.SetLabel("volume", match[1])
				flexgroupInstance.SetLabel("node", "")
				flexgroupInstance.SetLabel("uuid", "")
				flexgroupInstance.SetLabel(style, "flexgroup")
				flexgroupAggrsMap[key] = set.New()
				if err := metric.SetValueFloat64(flexgroupInstance, 1); err != nil {
					logger.Error("set value", slogx.Err(err), slog.String("metric", metricName))
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
				logger.Error("Failed to create new instance", slogx.Err(err), slog.String("key", key))
				continue
			}
			flexvolInstance.SetLabels(maps.Clone(i.GetLabels()))
			flexvolInstance.SetLabel(style, "flexvol")
			if err := metric.SetValueFloat64(flexvolInstance, 1); err != nil {
				logger.Error("Unable to set value on metric", slogx.Err(err), slog.String("metric", metricName))
			}
		}
	}

	logger.Debug("", slog.Int("flexgroup volume count", len(cache.GetInstances())))

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
			logger.Error("instance not in local cache", slog.String("key", key))
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
				logger.Error("metric not in local cache", slog.String("key", mkey))
				continue
			}

			if value, ok := m.GetValueFloat64(i); ok {
				fgv, _ := fgm.GetValueFloat64(fg)

				if !strings.HasSuffix(m.GetName(), "_latency") {
					err := fgm.SetValueFloat64(fg, fgv+value)
					if err != nil {
						logger.Error("error setting value", slogx.Err(err))
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
								logger.Error("error setting value", slogx.Err(err))
							}
						}
						err = fgm.SetValueFloat64(fg, fgv+prod)
						if err != nil {
							logger.Error("error setting value", slogx.Err(err))
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
								logger.Error("error setting value", slogx.Err(err))
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
