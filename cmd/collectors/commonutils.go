package collectors

import (
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBatchSize    = "500"
	MaxAllowedTimeDrift = 10 * time.Second
)

var validUnits = map[string]bool{
	"mW":    true,
	"W":     true,
	"mW*hr": true,
	"W*hr":  true,
}

type embedShelf struct {
	model, moduleType string
}

type PortData struct {
	Node  string
	Port  string
	Read  float64
	Write float64
}

// Reference https://kb.netapp.com/onprem/ontap/hardware/FAQ%3A_How_do_shelf_product_IDs_and_modules_in_ONTAP_map_to_a_model_of_a_shelf_or_storage_system_with_embedded_storage
// There are two ways to identify embedded disk shelves:
// 1. The shelf's module type ends with E
// 2. The shelf is listed in the link above
var combinations = map[embedShelf]bool{
	{"FS424-12", "IOM12F"}: true,
	{"DS212-12", "IOM12G"}: true,
}

func IsEmbedShelf(model string, moduleType string) bool {
	model = strings.ToUpper(model)
	moduleType = strings.ToUpper(moduleType)

	// if module type ends with E
	if strings.HasSuffix(moduleType, "E") {
		return true
	}

	return combinations[embedShelf{model, moduleType}]
}

func InvokeRestCallWithTestFile(client *rest.Client, href string, logger *slog.Logger, testFilePath string) ([]gjson.Result, error) {
	if testFilePath != "" {
		b, err := os.ReadFile(testFilePath)
		if err != nil {
			return []gjson.Result{}, err
		}
		testData := gjson.Result{Type: gjson.JSON, Raw: string(b)}
		return testData.Get("records").Array(), nil
	}
	return InvokeRestCall(client, href, logger)
}

func InvokeRestCall(client *rest.Client, href string, logger *slog.Logger) ([]gjson.Result, error) {
	result, err := rest.Fetch(client, href)
	if err != nil {
		logger.Error(
			"Failed to fetch data",
			slog.Any("err", err),
			slog.String("href", href),
			slog.Int("hrefLength", len(href)),
		)
		return []gjson.Result{}, err
	}

	if len(result) == 0 {
		return []gjson.Result{}, nil
	}

	return result, nil
}

func GetClusterTime(client *rest.Client, returnTimeOut *int, logger *slog.Logger) (time.Time, error) {
	var (
		err         error
		records     []gjson.Result
		clusterTime time.Time
		timeOfNodes []int64
	)

	query := "private/cli/cluster/date"
	fields := []string{"date"}

	maxRecords := 1

	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(&maxRecords).
		ReturnTimeout(returnTimeOut).
		Build()

	if records, err = rest.Fetch(client, href); err != nil {
		return clusterTime, err
	}
	if len(records) == 0 {
		return clusterTime, errs.New(errs.ErrConfig, " date not found on cluster")
	}

	for _, instanceData := range records {
		currentClusterDate := instanceData.Get("date")
		if currentClusterDate.Exists() {
			t, err := time.Parse(time.RFC3339, currentClusterDate.String())
			if err != nil {
				logger.Error(
					"Failed to load cluster date",
					slog.Any("err", err),
					slog.String("date", currentClusterDate.String()),
				)
				continue
			}
			clusterTime = t
			timeOfNodes = append(timeOfNodes, t.UnixNano())
		}
	}

	for _, timeOfEachNode := range timeOfNodes {
		timeDrift := time.Duration(timeOfEachNode - timeOfNodes[0]).Abs()
		if timeDrift >= MaxAllowedTimeDrift {
			logger.Warn(
				"Time drift exists between nodes",
				slog.Float64("timedrift(in sec)", timeDrift.Seconds()),
			)
			break
		}
	}

	return clusterTime, nil
}

// GetDataInterval fetch pollData interval
func GetDataInterval(param *node.Node, defaultInterval time.Duration) (time.Duration, error) {
	var dataIntervalStr string
	var durationVal time.Duration
	var err error
	schedule := param.GetChildS("schedule")
	if schedule != nil {
		dataInterval := schedule.GetChildS("data")
		if dataInterval != nil {
			dataIntervalStr = dataInterval.GetContentS()
			if durationVal, err = time.ParseDuration(dataIntervalStr); err == nil {
				return durationVal, nil
			}
			return defaultInterval, err
		}
	}
	return defaultInterval, nil
}

func UpdateProtectedFields(instance *matrix.Instance) {

	// check for group_type
	// Supported group_type are: "none", "vserver", "infinitevol", "consistencygroup", "flexgroup"
	if instance.GetLabel("group_type") != "" {

		groupType := instance.GetLabel("group_type")
		destinationVolume := instance.GetLabel("destination_volume")
		sourceVolume := instance.GetLabel("source_volume")
		destinationLocation := instance.GetLabel("destination_location")

		isSvmDr := groupType == "vserver" && destinationVolume == "" && sourceVolume == ""
		isCg := groupType == "consistencygroup" && strings.Contains(destinationLocation, ":/cg/")
		isConstituentVolumeRelationshipWithinSvmDr := groupType == "vserver" && !strings.HasSuffix(destinationLocation, ":")
		isConstituentVolumeRelationshipWithinCG := groupType == "consistencygroup" && !strings.Contains(destinationLocation, ":/cg/")

		// Update protectedBy label
		switch {
		case isSvmDr || isConstituentVolumeRelationshipWithinSvmDr:
			instance.SetLabel("protectedBy", "storage_vm")
		case isCg || isConstituentVolumeRelationshipWithinCG:
			instance.SetLabel("protectedBy", "cg")
		default:
			instance.SetLabel("protectedBy", "volume")
		}

		// SVM-DR related information is populated, Update protectionSourceType label
		switch {
		case isSvmDr:
			instance.SetLabel("protectionSourceType", "storage_vm")
		case isCg:
			instance.SetLabel("protectionSourceType", "cg")
		case isConstituentVolumeRelationshipWithinSvmDr || isConstituentVolumeRelationshipWithinCG || groupType == "none" || groupType == "flexgroup":
			instance.SetLabel("protectionSourceType", "volume")
		default:
			instance.SetLabel("protectionSourceType", "not_mapped")
		}
	}

	// Update derived_relationship_type field based on the policyType
	relationshipType := instance.GetLabel("relationship_type")
	if policyType := instance.GetLabel("policy_type"); policyType != "" {
		switch {
		case policyType == "strict_sync_mirror":
			instance.SetLabel("derived_relationship_type", "sync_mirror_strict")
		case policyType == "sync_mirror":
			instance.SetLabel("derived_relationship_type", "sync_mirror")
		case policyType == "mirror_vault":
			instance.SetLabel("derived_relationship_type", "mirror_vault")
		case policyType == "automated_failover":
			instance.SetLabel("derived_relationship_type", "automated_failover")
		case policyType == "automated_failover_duplex":
			instance.SetLabel("derived_relationship_type", "automated_failover_duplex")
		default:
			instance.SetLabel("derived_relationship_type", relationshipType)
		}
	} else {
		instance.SetLabel("derived_relationship_type", relationshipType)
	}
}

func SetNameservice(nsDB, nsSource, nisDomain string, instance *matrix.Instance) {
	requiredNSDb := false
	requiredNSSource := false

	if strings.Contains(nsDB, "passwd") || strings.Contains(nsDB, "group") || strings.Contains(nsDB, "netgroup") {
		requiredNSDb = true
		if strings.Contains(nsSource, "nis") {
			requiredNSSource = true
		}
	}

	if nisDomain != "" && requiredNSDb && requiredNSSource {
		instance.SetLabel("nis_authentication_enabled", "true")
	} else {
		instance.SetLabel("nis_authentication_enabled", "false")
	}
}

// IsTimestampOlderThanDuration - timestamp units are micro seconds
// The `begin` argument lets us virtualize time without requiring sleeps in test code
func IsTimestampOlderThanDuration(nowish time.Time, timestamp float64, duration time.Duration) bool {
	return nowish.Sub(time.UnixMicro(int64(timestamp))) > duration
}

func UpdateLagTime(instance *matrix.Instance, lastTransferSize *matrix.Metric, lagTime *matrix.Metric, logger *slog.Logger) {
	healthy := instance.GetLabel("healthy")
	schedule := instance.GetLabel("schedule")
	lastError := instance.GetLabel("last_transfer_error")

	// If SM relationship is healthy, has a schedule, last_transfer_error is empty, and last_transfer_bytes is 0, Then we are setting lag_time to 0
	// Otherwise, report the lag_time which ONTAP has originally reported.
	if lastBytes, ok := lastTransferSize.GetValueFloat64(instance); ok {
		if healthy == "true" && schedule != "" && lastError == "" && lastBytes == 0 {
			if err := lagTime.SetValueFloat64(instance, 0); err != nil {
				logger.Error("Unable to set value on metric", slog.Any("err", err), slog.String("metric", lagTime.GetName()))
			}
		}
	}
}

func IsValidUnit(unit string) bool {
	return validUnits[unit]
}

func ReadPluginKey(param *node.Node, key string) bool {
	if val := param.GetChildContentS(key); val != "" {
		if boolValue, err := strconv.ParseBool(val); err == nil {
			return boolValue
		}
	}
	return false
}

type VscanNames struct {
	Svm     string
	Scanner string
	Node    string
}

// SplitVscanName splits the vscan name into three parts and returns them as a VscanNames
func SplitVscanName(ontapName string, isZapi bool) (VscanNames, bool) {
	// colon separated list of fields
	// ZapiPerf uses this format: svm:scanner:node
	// RestPerf uses this format: node:svm:scanner

	// ZapiPerf examples:
	// svm      : scanner                  : node
	// vs_test4 : 2.2.2.2                  : umeng-aff300-05
	// moon-ad  : 2a03:1e80:a15:60c::1:2a5 : moon-02

	// RestPerf examples:
	// node                 : svm      : scanner
	// sti46-vsim-ucs519d   : vs0      : 172.29.120.57

	firstColon := strings.Index(ontapName, ":")
	if firstColon == -1 {
		return VscanNames{}, false
	}
	lastColon := strings.LastIndex(ontapName, ":")
	if lastColon == -1 {
		return VscanNames{}, false
	}
	if firstColon == lastColon {
		return VscanNames{}, false
	}

	if isZapi {
		return VscanNames{Svm: ontapName[:firstColon], Scanner: ontapName[firstColon+1 : lastColon], Node: ontapName[lastColon+1:]}, true
	}
	return VscanNames{Node: ontapName[:firstColon], Svm: ontapName[firstColon+1 : lastColon], Scanner: ontapName[lastColon+1:]}, true
}

func AggregatePerScanner(logger *slog.Logger, data *matrix.Matrix, latencyKey string, rateKey string) ([]*matrix.Matrix, *util.Metadata, error) {
	// When isPerScanner=true, Harvest 1.6 uses this form:
	// netapp.perf.dev.nltl-fas2520.vscan.scanner.10_64_30_62.scanner_stats_pct_mem_used 18 1501765640

	// These counters are per scanner and need averaging:
	// 		scanner_stats_pct_cpu_used
	// 		scanner_stats_pct_mem_used
	// 		scanner_stats_pct_network_used
	// These counters need to be summed:
	// 		scan_request_dispatched_rate
	// These counters need weighted averages:
	// 		scan_latency

	// create per scanner instance cache
	cache := data.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: true})
	var err error
	cache.UUID += ".Vscan"
	opsKeyPrefix := "temp_"

	for _, i := range data.GetInstances() {
		if !i.IsExportable() {
			continue
		}
		scanner := i.GetLabel("scanner")
		if cache.GetInstance(scanner) == nil {
			s, _ := cache.NewInstance(scanner)
			s.SetLabel("scanner", scanner)
		}
		i.SetExportable(false)
	}

	// aggregate per scanner
	counts := make(map[string]map[string]int) // map[scanner][counter] => value

	for _, i := range data.GetInstances() {
		scanner := i.GetLabel("scanner")
		ps := cache.GetInstance(scanner)
		if ps == nil {
			logger.Error("Failed to find scanner instance in cache", slog.String("scanner", scanner))
			continue
		}
		_, ok := counts[scanner]
		if !ok {
			counts[scanner] = make(map[string]int)
		}
		for mKey, m := range data.GetMetrics() {
			if !m.IsExportable() && m.GetType() != "float64" {
				continue
			}
			psm := cache.GetMetric(mKey)
			if psm == nil {
				logger.Error(
					"Failed to find metric in scanner cache",
					slog.String("scanner", scanner),
					slog.String("metric", mKey),
				)
				continue
			}
			if value, ok := m.GetValueFloat64(i); ok {
				fv, _ := psm.GetValueFloat64(ps)

				if mKey == latencyKey {
					// weighted average for scan.latency
					opsKey := m.GetComment()

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
								tempOpsV, _ = tempOps.GetValueFloat64(ps)
							}

							if value != 0 {
								err = tempOps.SetValueFloat64(ps, tempOpsV+opsValue)
								if err != nil {
									logger.Error("error", slog.Any("err", err))
								}
							}
							err = psm.SetValueFloat64(ps, fv+prod)
							if err != nil {
								logger.Error("error", slog.Any("err", err))
							}
						}
					}

					continue
				}

				// sum for scan_request_dispatched_rate
				if mKey == rateKey {
					err := psm.SetValueFloat64(ps, fv+value)
					if err != nil {
						logger.Error(
							"Error setting metric value",
							slog.Any("err", err),
							slog.String("metric", "scan_request_dispatched_rate"),
						)
					}

					continue
				} else if strings.HasSuffix(mKey, "_used") {
					// these need averaging
					counts[scanner][mKey]++
					runningTotal, _ := psm.GetValueFloat64(ps)
					value, _ := m.GetValueFloat64(ps)
					err := psm.SetValueFloat64(ps, runningTotal+value)
					if err != nil {
						logger.Error("Failed to set value", slog.Any("err", err), slog.String("mKey", mKey))
					}
				}
			}
		}
	}

	// cook averaged values and latencies
	for scanner, i := range cache.GetInstances() {
		for mKey, m := range cache.GetMetrics() {
			if !m.IsExportable() {
				continue
			}
			if strings.HasSuffix(m.GetName(), "_used") {
				count := counts[scanner][mKey]
				value, ok := m.GetValueFloat64(i)
				if !ok {
					continue
				}
				if err := m.SetValueFloat64(i, value/float64(count)); err != nil {
					logger.Error(
						"Unable to set average",
						slog.Any("err", err),
						slog.String("mKey", mKey),
						slog.String("name", m.GetName()),
					)
				}
			} else if strings.HasSuffix(m.GetName(), "_latency") {
				if value, ok := m.GetValueFloat64(i); ok {
					opsKey := m.GetComment()

					if ops := cache.GetMetric(opsKeyPrefix + opsKey); ops != nil {
						if opsValue, ok := ops.GetValueFloat64(i); ok && opsValue != 0 {
							err := m.SetValueFloat64(i, value/opsValue)
							if err != nil {
								logger.Error("error", slog.Any("err", err))
							}
						} else {
							m.SetValueNAN(i)
						}
					}
				}

			}
		}
	}

	return []*matrix.Matrix{cache}, nil, nil
}

func PopulateIfgroupMetrics(portIfgroupMap map[string]string, portDataMap map[string]PortData, nData *matrix.Matrix, logger *slog.Logger) error {
	var err error
	for portKey, ifgroupName := range portIfgroupMap {
		portInfo, ok := portDataMap[portKey]
		if !ok {
			continue
		}
		nodeName := portInfo.Node
		port := portInfo.Port
		readBytes := portInfo.Read
		writeBytes := portInfo.Write

		ifgrpupInstanceKey := nodeName + ifgroupName
		ifgroupInstance := nData.GetInstance(ifgrpupInstanceKey)
		if ifgroupInstance == nil {
			ifgroupInstance, err = nData.NewInstance(ifgrpupInstanceKey)
			if err != nil {
				logger.Debug(
					"Failed to add instance",
					slog.Any("err", err),
					slog.String("ifgrpupInstanceKey", ifgrpupInstanceKey),
				)
				return err
			}
		}

		// set labels
		ifgroupInstance.SetLabel("node", nodeName)
		ifgroupInstance.SetLabel("ifgroup", ifgroupName)
		if ifgroupInstance.GetLabel("ports") != "" {
			portSlice := []string{ifgroupInstance.GetLabel("ports"), port}
			// make sure ports are always in sorted order
			sort.Strings(portSlice)
			ifgroupInstance.SetLabel("ports", strings.Join(portSlice, ","))
		} else {
			ifgroupInstance.SetLabel("ports", port)
		}

		rx := nData.GetMetric("rx_bytes")
		rxv, _ := rx.GetValueFloat64(ifgroupInstance)
		if err = rx.SetValueFloat64(ifgroupInstance, readBytes+rxv); err != nil {
			logger.Debug("Failed to parse value", slog.Any("value", readBytes), slog.Any("err", err))
		}

		tx := nData.GetMetric("tx_bytes")
		txv, _ := tx.GetValueFloat64(ifgroupInstance)
		if err = tx.SetValueFloat64(ifgroupInstance, writeBytes+txv); err != nil {
			logger.Debug("Failed to parse value", slog.Any("value", writeBytes), slog.Any("err", err))
		}
	}
	return nil
}
