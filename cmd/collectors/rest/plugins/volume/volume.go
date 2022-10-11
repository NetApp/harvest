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
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

const DefaultPluginDuration = 30 * time.Minute
const DefaultDataPollDuration = 3 * time.Minute

type Volume struct {
	*plugin.AbstractPlugin
	pluginInvocationRate int
	currentVal           int
	client               *rest.Client
	aggrsMap             map[string]string // aggregate-uuid -> aggregate-name map
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (my *Volume) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if my.client, err = rest.New(conf.ZapiPoller(my.ParentParams), timeout); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.aggrsMap = make(map[string]string)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	if my.currentVal, err = collectors.SetPluginInterval(my.ParentParams, my.Params, my.Logger, DefaultDataPollDuration, DefaultPluginDuration); err != nil {
		my.Logger.Error().Err(err).Stack().Msg("Failed while setting the plugin interval")
		return err
	}

	return nil
}

func (my *Volume) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	if my.currentVal >= my.pluginInvocationRate {
		my.currentVal = 0

		// invoke disk rest and populate info in aggrsMap
		if disks, err := my.getEncryptedDisks(); err != nil {
			if errs.IsAPINotFound(err) {
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

	my.currentVal++
	return nil, nil
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

func (my *Volume) getEncryptedDisks() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	diskFields := []string{"aggregates.name", "aggregates.uuid"}
	query := "api/storage/disks"
	href := rest.BuildHref("", strings.Join(diskFields, ","), []string{"protection_mode=!data|full"}, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(my.client, query, href, my.Logger); err != nil {
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
