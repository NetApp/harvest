package collectors

import (
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

const (
	DefaultBatchSize    = "500"
	MaxAllowedTimeDrift = 10 * time.Second
)

func InvokeRestCall(client *rest.Client, href string, logger *logging.Logger) ([]gjson.Result, error) {
	result, err := rest.Fetch(client, href)
	if err != nil {
		logger.Error().Err(err).Str("href", href).Msg("Failed to fetch data")
		return []gjson.Result{}, err
	}

	if len(result) == 0 {
		return []gjson.Result{}, nil
	}

	return result, nil
}

func GetClusterTime(client *rest.Client, returnTimeOut string, logger *logging.Logger) (time.Time, error) {
	var (
		err         error
		records     []gjson.Result
		clusterTime time.Time
		timeOfNodes []int64
	)

	query := "private/cli/cluster/date"
	fields := []string{"date"}

	href := rest.BuildHref(query, strings.Join(fields, ","), nil, "", "", "1", returnTimeOut, "")

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
				logger.Error().Str("date", currentClusterDate.String()).Err(err).Msg("Failed to load cluster date")
				continue
			}
			clusterTime = t
			timeOfNodes = append(timeOfNodes, t.UnixNano())
		}
	}

	for _, timeOfEachNode := range timeOfNodes {
		timeDrift := time.Duration(timeOfEachNode - timeOfNodes[0]).Abs()
		if timeDrift >= MaxAllowedTimeDrift {
			logger.Warn().Float64("timedrift(in sec)", timeDrift.Seconds()).Msg("Time drift exist among the nodes")
			break
		}
	}

	logger.Debug().Str("cluster time", clusterTime.String()).Msg("")
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
		if isSvmDr || isConstituentVolumeRelationshipWithinSvmDr {
			instance.SetLabel("protectedBy", "storage_vm")
		} else if isCg || isConstituentVolumeRelationshipWithinCG {
			instance.SetLabel("protectedBy", "cg")
		} else {
			instance.SetLabel("protectedBy", "volume")
		}

		// SVM-DR related information is populated, Update protectionSourceType label
		if isSvmDr {
			instance.SetLabel("protectionSourceType", "storage_vm")
		} else if isCg {
			instance.SetLabel("protectionSourceType", "cg")
		} else if isConstituentVolumeRelationshipWithinSvmDr || isConstituentVolumeRelationshipWithinCG || groupType == "none" || groupType == "flexgroup" {
			instance.SetLabel("protectionSourceType", "volume")
		} else {
			instance.SetLabel("protectionSourceType", "not_mapped")
		}
	}

	// Update derived_relationship_type field based on the policyType
	relationshipType := instance.GetLabel("relationship_type")
	if policyType := instance.GetLabel("policy_type"); policyType != "" {
		if policyType == "strict_sync_mirror" {
			instance.SetLabel("derived_relationship_type", "sync_mirror_strict")
		} else if policyType == "sync_mirror" {
			instance.SetLabel("derived_relationship_type", "sync_mirror")
		} else if policyType == "mirror_vault" {
			instance.SetLabel("derived_relationship_type", "mirror_vault")
		} else if policyType == "automated_failover" {
			instance.SetLabel("derived_relationship_type", "sync_mirror")
		} else {
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

func UpdateLagTime(instance *matrix.Instance, lastTransferSize *matrix.Metric, lagTime *matrix.Metric, logger *logging.Logger) {
	healthy := instance.GetLabel("healthy")
	schedule := instance.GetLabel("schedule")
	lastError := instance.GetLabel("last_transfer_error")
	relationshipID := instance.GetLabel("relationship_id")

	// If SM relationship is healthy, has a schedule, last_transfer_error is empty, and last_transfer_bytes is 0, Then we are setting lag_time to 0
	// Otherwise, report the lag_time which ONTAP has originally reported.
	if lastBytes, ok := lastTransferSize.GetValueFloat64(instance); ok {
		if healthy == "true" && schedule != "" && lastError == "" && lastBytes == 0 {
			lag, _ := lagTime.GetValueFloat64(instance)
			if err := lagTime.SetValueFloat64(instance, 0); err != nil {
				logger.Error().Err(err).Str("metric", lagTime.GetName()).Msg("Unable to set value on metric")
			}
			logger.Debug().Msgf("lagTime value set from %f to 0 for %s. Healthy: %s, Schedule: %s, LastBytes: %f, LastError:%s", lag, relationshipID, healthy, schedule, lastBytes, lastError)
		}
	}
}
