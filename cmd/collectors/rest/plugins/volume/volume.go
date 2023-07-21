/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package volume

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Volume struct {
	*plugin.AbstractPlugin
	currentVal int
	client     *rest.Client
	aggrsMap   map[string]string // aggregate-uuid -> aggregate-name map
	arw        *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (my *Volume) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	my.aggrsMap = make(map[string]string)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	my.currentVal = my.SetPluginInterval()

	if my.Options.IsTest {
		return nil
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if my.client, err = rest.New(conf.ZapiPoller(my.ParentParams), timeout, my.Auth); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.arw = matrix.New(my.Parent+".Volume", "volume_arw", "volume_arw")
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "ArwStatus")
	my.arw.SetExportOptions(exportOptions)
	_, err = my.arw.NewMetricFloat64("status", "status")
	if err != nil {
		my.Logger.Error().Stack().Err(err).Msg("add metric")
		return err
	}
	return nil
}

func (my *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[my.Object]
	if my.currentVal >= my.PluginInvocationRate {
		my.currentVal = 0

		// invoke disk rest and populate info in aggrsMap
		if disks, err := my.getEncryptedDisks(); err != nil {
			if errs.IsRestErr(err, errs.APINotFound) {
				my.Logger.Debug().Err(err).Msg("Failed to collect disk data")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect disk data")
			}
		} else {
			// update aggrsMap based on disk data
			my.updateAggrMap(disks)
		}
	}

	// update volume instance labels
	my.updateVolumeLabels(data)

	// parse anti_ransomware_start_time, antiRansomwareState for all volumes and export at cluster level
	my.handleARWProtection(data)

	my.currentVal++
	return []*matrix.Matrix{my.arw}, nil
}

func (my *Volume) updateVolumeLabels(data *matrix.Matrix) {
	for _, volume := range data.GetInstances() {
		// For flexgroup, aggrUuid in Rest should be empty for parity with Zapi response
		if volumeStyle := volume.GetLabel("style"); volumeStyle == "flexgroup" {
			volume.SetLabel("aggrUuid", "")
		}
		aggrUUID := volume.GetLabel("aggrUuid")

		_, exist := my.aggrsMap[aggrUUID]
		volume.SetLabel("isHardwareEncrypted", strconv.FormatBool(exist))
	}
}

func (my *Volume) handleARWProtection(data *matrix.Matrix) {
	var (
		arwInstance *matrix.Instance
		err         error
	)

	// Purge and reset data
	my.arw.PurgeInstances()
	my.arw.Reset()

	// Set all global labels
	my.arw.SetGlobalLabels(data.GetGlobalLabels())

	learningModeCount := 0
	learningCompleted := 0
	disabledCount := 0

	for _, volume := range data.GetInstances() {
		if arwState := volume.GetLabel("antiRansomwareState"); arwState != "" {
			if arwState == "dry_run" || arwState == "enable_paused" {
				if arwStartTime := volume.GetLabel("anti_ransomware_start_time"); arwStartTime != "" {
					// If ARW startTime is more than 30 days old, which indicates that learning mode has been finished.
					if (float64(time.Now().Unix()) - HandleTimestamp(arwStartTime)) > 2629743 {
						learningCompleted++
					}
				}
				learningModeCount++
			} else if arwState == "disabled" {
				disabledCount++
			}
		}
	}

	arwInstanceKey := data.GetGlobalLabels().Get("cluster") + data.GetGlobalLabels().Get("datacenter")
	if arwInstance, err = my.arw.NewInstance(arwInstanceKey); err != nil {
		my.Logger.Error().Stack().Err(err).Str("arwInstanceKey", arwInstanceKey).Msg("Failed to create arw instance")
		return
	}

	if disabledCount > 0 {
		arwInstance.SetLabel("ArwStatus", "Not Monitoring")
	} else if learningModeCount > 0 {
		if learningCompleted > 0 {
			arwInstance.SetLabel("ArwStatus", "Switch to Active Mode")
		} else {
			arwInstance.SetLabel("ArwStatus", "Learning Mode")
		}
	} else {
		arwInstance.SetLabel("ArwStatus", "Active Mode")
	}

	m := my.arw.GetMetric("status")
	// populate numeric data
	value := 1.0
	if err = m.SetValueFloat64(arwInstance, value); err != nil {
		my.Logger.Error().Stack().Err(err).Float64("value", value).Msg("Failed to parse value")
	} else {
		my.Logger.Debug().Float64("value", value).Msg("added value")
	}
}

func (my *Volume) getEncryptedDisks() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	diskFields := []string{"aggregates.name", "aggregates.uuid"}
	query := "api/storage/disks"
	href := rest.BuildHref("", strings.Join(diskFields, ","), []string{"protection_mode=!data|full"}, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(my.client, href, my.Logger); err != nil {
		return nil, err
	}
	return result, nil
}

func (my *Volume) updateAggrMap(disks []gjson.Result) {
	if disks != nil {
		// Clean aggrsMap map
		my.aggrsMap = make(map[string]string)

		for _, disk := range disks {
			aggrName := disk.Get("aggregates.name").String()
			aggrUUID := disk.Get("aggregates.uuid").String()
			if aggrUUID != "" {
				my.aggrsMap[aggrUUID] = aggrName
			}
		}
	}
}

// Example: timestamp: 2020-12-02T18:36:19-08:00
var regexTimeStamp = regexp.MustCompile(
	`[+-]?\d{4}(-[01]\d(-[0-3]\d(T[0-2]\d:[0-5]\d:?([0-5]\d(\.\d+)?)?[+-][0-2]\d:[0-5]\d?)?)?)?`)

func HandleTimestamp(value string) float64 {
	var timestamp time.Time
	var err error

	if match := regexTimeStamp.MatchString(value); match {
		// example: 2020-12-02T18:36:19-08:00   ==>  1606962979
		if timestamp, err = time.Parse(time.RFC3339, value); err != nil {
			fmt.Printf("%v", err)
			return 0
		}
		return float64(timestamp.Unix())
	}
	return 0
}
