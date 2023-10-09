/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package volume

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"strconv"
	"time"
)

const HoursInMonth = 24 * 30

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
		arwInstance       *matrix.Instance
		arwStartTimeValue time.Time
		err               error
	)

	// Purge and reset data
	my.arw.PurgeInstances()
	my.arw.Reset()

	// Set all global labels
	my.arw.SetGlobalLabels(data.GetGlobalLabels())
	arwStatusValue := "Active Mode"
	// Case where cluster don't have any volumes, arwStatus show as 'Not Monitoring'
	if len(data.GetInstances()) == 0 {
		arwStatusValue = "Not Monitoring"
	}

	// This is how cluster level arwStatusValue has been calculated based on each volume
	// If any one volume arwStatus is disabled --> "Not Monitoring"
	// If any one volume has been completed learning mode --> "Switch to Active Mode"
	// If all volumes are in learning mode --> "Learning Mode"
	// Else indicates arwStatus for all volumes are enabled --> "Active Mode"
	for _, volume := range data.GetInstances() {
		arwState := volume.GetLabel("antiRansomwareState")
		if arwState == "" {
			// Case where REST calls don't return `antiRansomwareState` field, arwStatus show as 'Not Monitoring'
			arwStatusValue = "Not Monitoring"
			break
		}
		if arwState == "disabled" {
			arwStatusValue = "Not Monitoring"
			break
		} else if arwState == "dry_run" || arwState == "enable_paused" {
			arwStartTime := volume.GetLabel("anti_ransomware_start_time")
			if arwStartTime == "" || arwStatusValue == "Switch to Active Mode" {
				continue
			}
			// If ARW startTime is more than 30 days old, which indicates that learning mode has been finished.
			if arwStartTimeValue, err = time.Parse(time.RFC3339, arwStartTime); err != nil {
				my.Logger.Error().Err(err).Msg("Failed to parse arw start time")
				arwStartTimeValue = time.Now()
			}
			if time.Since(arwStartTimeValue).Hours() > HoursInMonth {
				arwStatusValue = "Switch to Active Mode"
			} else {
				arwStatusValue = "Learning Mode"
			}
		}
	}

	arwInstanceKey := data.GetGlobalLabels()["cluster"] + data.GetGlobalLabels()["datacenter"]
	if arwInstance, err = my.arw.NewInstance(arwInstanceKey); err != nil {
		my.Logger.Error().Err(err).Str("arwInstanceKey", arwInstanceKey).Msg("Failed to create arw instance")
		return
	}

	arwInstance.SetLabel("ArwStatus", arwStatusValue)
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

	fields := []string{"aggregates.name", "aggregates.uuid"}
	query := "api/storage/disks"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{"protection_mode=!data|full"}).
		Build()

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
