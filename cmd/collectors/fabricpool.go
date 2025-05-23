package collectors

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"log/slog"
	"maps"
	"regexp"
	"strings"
)

var fabricpoolRegex = regexp.MustCompile(`^(.*)__(\d{4})_bin.*$`)

type constituentData struct {
	instance     *matrix.Instance
	flexgroupKey string
}

func GetFlexGroupFabricPoolMetrics(dataMap map[string]*matrix.Matrix, object string, opName string, includeConstituents bool, l *slog.Logger) (*matrix.Matrix, error) {
	var (
		err                      error
		latencyCacheMetrics      []string
		flexgroupConstituentsMap map[string]constituentData
	)

	data := dataMap[object]
	opsKeyPrefix := "temp_"
	flexgroupConstituentsMap = make(map[string]constituentData)

	cache := data.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: true})
	cache.UUID += ".FabricPool"

	// collect latency_average metrics names
	for mKey := range cache.GetMetrics() {
		if strings.HasPrefix(mKey, "cloud_bin_op_latency_average") {
			latencyCacheMetrics = append(latencyCacheMetrics, mKey)
		}
	}

	// create flexgroup instance cache
	for iKey, i := range data.GetInstances() {
		if !i.IsExportable() {
			continue
		}

		if match := fabricpoolRegex.FindStringSubmatch(fetchVolumeName(i)); len(match) == 3 {
			key := i.GetLabel("svm") + "." + match[1] + "." + i.GetLabel("cloud_target")
			if cache.GetInstance(key) == nil {
				fg, _ := cache.NewInstance(key)
				fg.SetLabels(maps.Clone(i.GetLabels()))
				fg.SetLabel("volume", match[1])
			}
			i.SetExportable(includeConstituents)
			flexgroupConstituentsMap[iKey] = constituentData{flexgroupKey: key, instance: i}
		}
	}

	l.Debug("extracted flexgroup volumes", slog.Int("size", len(cache.GetInstances())))

	// create summary
	for _, constituent := range flexgroupConstituentsMap {
		// instance key is svm.flexgroup-volume.cloud-target-name
		key := constituent.flexgroupKey
		i := constituent.instance

		fg := cache.GetInstance(key)
		if fg == nil {
			l.Error("instance not in local cache", slog.String("key", key))
			continue
		}

		for mkey, m := range data.GetMetrics() {
			if !m.IsExportable() && m.GetType() != "float64" {
				continue
			}

			fgm := cache.GetMetric(mkey)
			if fgm == nil {
				l.Error("metric not in local cache", slog.String("key", mkey))
				continue
			}

			if value, ok := m.GetValueFloat64(i); ok {
				fgv, _ := fgm.GetValueFloat64(fg)

				// non-latency metrics: simple sum
				if !strings.HasPrefix(mkey, "cloud_bin_op_latency_average") {
					fgm.SetValueFloat64(fg, fgv+value)
					continue
				}

				// latency metric: weighted sum
				opsKey := strings.Replace(mkey, "cloud_bin_op_latency_average", opName, 1)

				if ops := data.GetMetric(opsKey); ops != nil {
					if opsValue, ok := ops.GetValueFloat64(i); ok {
						var tempOpsV float64

						prod := value * opsValue
						tempOpsKey := opsKeyPrefix + opsKey
						tempOps := cache.GetMetric(tempOpsKey)

						if tempOps == nil {
							if tempOps, err = cache.NewMetricFloat64(tempOpsKey); err != nil {
								return nil, err
							}
							tempOps.SetExportable(false)
						} else {
							tempOpsV, _ = tempOps.GetValueFloat64(fg)
						}
						if value != 0 {
							tempOps.SetValueFloat64(fg, tempOpsV+opsValue)
						}
						fgm.SetValueFloat64(fg, fgv+prod)
					}
				}
			}
		}
	}

	// normalize latency values
	for _, i := range cache.GetInstances() {
		if !i.IsExportable() {
			continue
		}
		for _, mKey := range latencyCacheMetrics {
			m := cache.GetMetric(mKey)
			if m != nil && m.IsExportable() {
				if value, ok := m.GetValueFloat64(i); ok {
					opsKey := strings.Replace(mKey, "cloud_bin_op_latency_average", opName, 1)

					// fetch from temp metrics
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
	return cache, nil
}

func fetchVolumeName(i *matrix.Instance) string {
	// zapiperf name: test1_fg2__0001_bin_0_cfg_id_0
	// restperf name: test1_fg2__0002_bin_1_cfg_id_1
	instanceName := i.GetLabel("fabricpool")
	compAggrName := i.GetLabel("comp_aggr_name")
	if names := strings.Split(instanceName, compAggrName+"_"); len(names) == 2 {
		return names[1]
	}
	return ""
}
