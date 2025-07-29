package collectors

import (
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"maps"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

var flexgroupRegex = regexp.MustCompile(`^(.*)__(\d{4})$`)

var footprintMetrics = map[string]struct{}{
	"delayed_free_footprint":       {}, // Rest, Zapi
	"flexvol_metadata_footprint":   {}, // Rest
	"total_footprint":              {}, // Rest, Zapi
	"total_metadata_footprint":     {}, // Rest, Zapi
	"volume_blocks_footprint_bin0": {}, // Rest
	"volume_blocks_footprint_bin1": {}, // Rest
	"volume_guarantee_footprint":   {}, // Rest
	"metadata_footprint":           {}, // Zapi
	"guarantee_footprint":          {}, // Zapi
	"capacity_tier_footprint":      {}, // Zapi
	"performance_tier_footprint":   {}, // Zapi

}

type OpsData struct {
	TotalOps       float64
	Timestamp      float64
	priorTimestamp float64
}

func ProcessFlexGroupData(logger *slog.Logger, data *matrix.Matrix, style string, includeConstituents bool, opsKeyPrefix string, volumesMap map[string]string, enableVolumeAggrMatrix bool, zombieVolumeMatrix *matrix.Matrix, volumePastOpsMap map[string]OpsData, zombieMetricName string) ([]*matrix.Matrix, *collector.Metadata, error) {
	var err error

	if volumesMap == nil {
		logger.Debug("volumes config data is empty")
		return nil, nil, nil
	}

	fgAggrMap := make(map[string]*set.Set)
	flexgroupAggrsMap := make(map[string]*set.Set)

	metricName := "labels"
	volumeAggrMatrix := matrix.New(".Volume", "volume_aggr", "volume_aggr")
	volumeAggrMatrix.SetGlobalLabels(data.GetGlobalLabels())
	zombieVolumeMatrix.SetGlobalLabels(data.GetGlobalLabels())

	metric, err := volumeAggrMatrix.NewMetricFloat64(metricName)
	if err != nil {
		logger.Error("add metric", slogx.Err(err), slog.String("key", metricName))
		return nil, nil, err
	}

	cache := data.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: true})
	cache.UUID += ".Volume"

	for _, i := range data.GetInstances() {
		volName := i.GetLabel("volume")
		svmName := i.GetLabel("svm")
		switch volumesMap[svmName+volName] {
		case "flexgroup_constituent":
			match := flexgroupRegex.FindStringSubmatch(volName)
			key := svmName + "." + match[1]
			if cache.GetInstance(key) == nil {
				fg, _ := cache.NewInstance(key)
				fg.SetLabels(maps.Clone(i.GetLabels()))
				fg.SetLabel("volume", match[1])
				fg.SetLabel("node", "")
				fg.SetLabel("uuid", "")
				fg.SetLabel(style, "flexgroup")
				fgAggrMap[key] = set.New()
			}

			if volumeAggrMatrix.GetInstance(key) == nil {
				flexgroupInstance, _ := volumeAggrMatrix.NewInstance(key)
				flexgroupInstance.SetLabels(maps.Clone(i.GetLabels()))
				flexgroupInstance.SetLabel("volume", match[1])
				flexgroupInstance.SetLabel("node", "")
				flexgroupInstance.SetLabel("uuid", "")
				flexgroupInstance.SetLabel(style, "flexgroup")
				flexgroupAggrsMap[key] = set.New()
				metric.SetValueFloat64(flexgroupInstance, 1)
			}
			//nolint:gocritic
			//if zombieVolumeMatrix.GetInstance(key) == nil {
			//	zombieFlexgroupInstance, _ := zombieVolumeMatrix.NewInstance(key)
			//	zombieFlexgroupInstance.SetLabels(maps.Clone(i.GetLabels()))
			//	zombieFlexgroupInstance.SetLabel("volume", match[1])
			//	zombieFlexgroupInstance.SetLabel("node", "")
			//	zombieFlexgroupInstance.SetLabel("uuid", "")
			//	zombieFlexgroupInstance.SetLabel(style, "flexgroup")
			//	//metric.SetValueFloat64(zombieFlexgroupInstance, 1)
			//}
			fgAggrMap[key].Add(i.GetLabel("aggr"))
			flexgroupAggrsMap[key].Add(i.GetLabel("aggr"))
			i.SetLabel(style, "flexgroup_constituent")
			i.SetExportable(includeConstituents)
		case "flexvol":
			i.SetLabel(style, "flexvol")
			key := svmName + "." + volName
			flexvolInstance, err := volumeAggrMatrix.NewInstance(key)
			if err != nil {
				logger.Error("Failed to create new instance", slogx.Err(err), slog.String("key", key))
				continue
			}
			flexvolInstance.SetLabels(maps.Clone(i.GetLabels()))
			flexvolInstance.SetLabel(style, "flexvol")
			metric.SetValueFloat64(flexvolInstance, 1)
			zombieFlexvolInstance, err := zombieVolumeMatrix.NewInstance(key)
			if err != nil {
				logger.Error("Failed to create new instance", slogx.Err(err), slog.String("key", key))
				continue
			}
			zombieFlexvolInstance.SetLabels(maps.Clone(i.GetLabels()))
			zombieFlexvolInstance.SetLabel(style, "flexvol")
			totalOps := data.GetMetric("total_ops")
			totalOpsVal, _ := totalOps.GetValueFloat64(i)
			zombieMetric := zombieVolumeMatrix.GetMetric(zombieMetricName)
			if volumePastOpsMap[key].Timestamp > 0.0 && (math.Abs(volumePastOpsMap[key].priorTimestamp-float64(time.Now().UnixMicro())) > float64(time.Minute.Microseconds()) && math.Abs(volumePastOpsMap[key].TotalOps-totalOpsVal) < 10.0) {
				zombieMetric.SetValueFloat64(zombieFlexvolInstance, 1)
				volumePastOpsMap[key] = OpsData{TotalOps: totalOpsVal, Timestamp: float64(time.Now().UnixMicro())}
			} else {
				zombieMetric.SetValueFloat64(zombieFlexvolInstance, 0)
				volumePastOpsMap[key] = OpsData{TotalOps: totalOpsVal, Timestamp: float64(time.Now().UnixMicro()), priorTimestamp: float64(time.Now().UnixMicro())}
			}
		}
	}

	logger.Debug("", slog.Int("flexgroup volume count", len(cache.GetInstances())))

	recordFGFalse := make(map[string]*set.Set)
	for _, i := range data.GetInstances() {
		volName := i.GetLabel("volume")
		svmName := i.GetLabel("svm")
		if volumesMap[svmName+volName] != "flexgroup_constituent" {
			continue
		}
		match := flexgroupRegex.FindStringSubmatch(volName)
		key := svmName + "." + match[1]

		flexgroupInstance := volumeAggrMatrix.GetInstance(key)
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
					fgm.SetValueFloat64(fg, fgv+value)
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
							tempOps.SetValueFloat64(fg, tempOpsV+opsValue)
						}
						fgm.SetValueFloat64(fg, fgv+prod)
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
							m.SetValueFloat64(i, value/opsValue)
						} else {
							m.SetValueNAN(i)
						}
					}
				}
			}
		}
	}

	if enableVolumeAggrMatrix {
		return []*matrix.Matrix{cache, volumeAggrMatrix, zombieVolumeMatrix}, nil, nil
	}
	return []*matrix.Matrix{cache, zombieVolumeMatrix}, nil, nil
}

func ProcessFlexGroupFootPrint(data *matrix.Matrix, logger *slog.Logger) *matrix.Matrix {
	fgAggrMap := make(map[string]*set.Set)

	cache := data.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: true})
	cache.UUID += ".VolumeFootPrint.Flexgroup"
	// remove instance_labels from this matrix otherwise it will emit volume_labels
	cache.GetExportOptions().PopChildS("instance_labels")

	for _, i := range data.GetInstances() {
		volName := i.GetLabel("volume")
		svmName := i.GetLabel("svm")
		style := i.GetLabel("style")
		if style != "flexgroup_constituent" {
			continue
		}
		match := flexgroupRegex.FindStringSubmatch(volName)
		if len(match) < 2 {
			logger.Error("regex match failed or capture group missing", slog.String("volume", volName))
			continue
		}
		key := svmName + "." + match[1]
		if cache.GetInstance(key) == nil {
			fg, _ := cache.NewInstance(key)
			fg.SetLabels(maps.Clone(i.GetLabels()))
			fg.SetLabel("volume", match[1])
			fg.SetLabel("node", "")
			fg.SetLabel("uuid", "")
			fg.SetLabel("style", "flexgroup")
			fgAggrMap[key] = set.New()
		}
		fgAggrMap[key].Add(i.GetLabel("aggr"))
	}

	for _, i := range data.GetInstances() {
		volName := i.GetLabel("volume")
		svmName := i.GetLabel("svm")
		style := i.GetLabel("style")

		if style != "flexgroup_constituent" {
			continue
		}
		match := flexgroupRegex.FindStringSubmatch(volName)
		if len(match) < 2 {
			logger.Error("regex match failed or capture group missing", slog.String("volume", volName))
			continue
		}
		key := svmName + "." + match[1]

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

			_, ok := footprintMetrics[mkey]
			if !ok {
				continue
			}

			fgm := cache.GetMetric(mkey)
			if fgm == nil {
				logger.Error("metric not in local cache", slog.String("key", mkey))
				continue
			}

			if value, ok := m.GetValueFloat64(i); ok {
				fgv, _ := fgm.GetValueFloat64(fg)
				fgm.SetValueFloat64(fg, fgv+value)
			}
		}
	}

	return cache
}
