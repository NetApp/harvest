package collectors

import (
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"golang.org/x/exp/maps"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`^(.*)__(\d{4})$`)

func GetFlexGroupFabricPoolMetrics(dataMap map[string]*matrix.Matrix, object string, opName string, includeConstituents bool, l *logging.Logger) (*matrix.Matrix, error) {
	var (
		err                 error
		latencyCacheMetrics []string
	)

	data := dataMap[object]
	opsKeyPrefix := "temp_"

	cache := data.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: true})
	cache.UUID += ".FabricPool"

	// collect latency_average metrics names
	for mKey := range cache.GetMetrics() {
		if strings.HasPrefix(mKey, "cloud_bin_op_latency_average") {
			latencyCacheMetrics = append(latencyCacheMetrics, mKey)
		}
	}

	// create flexgroup instance cache
	for _, i := range data.GetInstances() {
		if !i.IsExportable() {
			continue
		}
		if match := re.FindStringSubmatch(i.GetLabel("volume")); len(match) == 3 {
			key := i.GetLabel("svm") + "." + match[1] + i.GetLabel("cloud_target")
			if cache.GetInstance(key) == nil {
				fg, _ := cache.NewInstance(key)
				fg.SetLabels(maps.Clone(i.GetLabels()))
				fg.SetLabel("volume", match[1])
			}
			i.SetExportable(includeConstituents)
		}
	}

	l.Logger.Debug().Int("size", len(cache.GetInstances())).Msg("extracted  flexgroup volumes")

	// create summary
	for _, i := range data.GetInstances() {
		if match := re.FindStringSubmatch(i.GetLabel("volume")); len(match) == 3 {
			// instance key is svm.flexgroup-volume.cloud-target-name
			key := i.GetLabel("svm") + "." + match[1] + i.GetLabel("cloud_target")

			fg := cache.GetInstance(key)
			if fg == nil {
				l.Logger.Error().Str("key", key).Msg("instance not in local cache")
				continue
			}

			for mkey, m := range data.GetMetrics() {
				if !m.IsExportable() && m.GetType() != "float64" {
					continue
				}

				fgm := cache.GetMetric(mkey)
				if fgm == nil {
					l.Logger.Error().Str("key", mkey).Msg("metric not in local cache")
					continue
				}

				if value, ok := m.GetValueFloat64(i); ok {
					fgv, _ := fgm.GetValueFloat64(fg)

					// non-latency metrics: simple sum
					if !strings.HasPrefix(mkey, "cloud_bin_op_latency_average") {
						err := fgm.SetValueFloat64(fg, fgv+value)
						if err != nil {
							l.Logger.Error().Err(err).Msg("error")
						}
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
								err = tempOps.SetValueFloat64(fg, tempOpsV+opsValue)
								if err != nil {
									l.Logger.Error().Err(err).Msg("error")
								}
							}
							err = fgm.SetValueFloat64(fg, fgv+prod)
							if err != nil {
								l.Logger.Error().Err(err).Msg("error")
							}
						}
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
							err := m.SetValueFloat64(i, value/opsValue)
							if err != nil {
								l.Logger.Error().Err(err).Msg("error")
							}
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
