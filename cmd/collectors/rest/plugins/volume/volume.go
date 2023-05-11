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

type Volume struct {
	*plugin.AbstractPlugin
	currentVal int
	client     *rest.Client
	aggrsMap   map[string]string // aggregate-uuid -> aggregate-name map
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
