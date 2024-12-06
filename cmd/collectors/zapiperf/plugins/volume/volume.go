/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package volume

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
)

const batchSize = "500"

type Volume struct {
	*plugin.AbstractPlugin
	currentVal          int
	styleType           string
	includeConstituents bool
	client              *zapi.Client
	volumesMap          map[string]string // volume-name -> volume-extended-style map
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (v *Volume) Init(remote conf.Remote) error {
	var err error
	if err := v.InitAbc(); err != nil {
		return err
	}

	v.styleType = "style"

	if v.Params.HasChildS("historicalLabels") {
		v.styleType = "type"
	}

	if v.Options.IsTest {
		return nil
	}

	v.volumesMap = make(map[string]string)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	v.currentVal = v.SetPluginInterval()

	// Read template to decide inclusion of flexgroup constituents
	v.includeConstituents = collectors.ReadPluginKey(v.Params, "include_constituents")

	if v.client, err = zapi.New(conf.ZapiPoller(v.ParentParams), v.Auth); err != nil {
		v.SLogger.Error("connecting", slog.Any("err", err))
		return err
	}
	return v.client.Init(5, remote)
}

func (v *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[v.Object]
	style := v.styleType
	opsKeyPrefix := "temp_"
	if v.currentVal >= v.PluginInvocationRate {
		v.currentVal = 0
		// Attempt to fetch new volumes
		newVolumesMap, err := v.fetchVolumes()
		if err != nil {
			v.SLogger.Error("Failed to fetch volumes, retaining cached volumesMap", slog.Any("err", err))
		} else {
			// Only update volumesMap if fetchVolumes was successful
			v.volumesMap = newVolumesMap
		}
	}

	v.currentVal++
	return collectors.ProcessFlexGroupData(v.SLogger, data, style, v.includeConstituents, opsKeyPrefix, v.volumesMap)
}

func (v *Volume) fetchVolumes() (map[string]string, error) {
	var (
		result     *node.Node
		volumes    []*node.Node
		volumesMap = make(map[string]string)
	)

	query := "volume-get-iter"
	tag := "initial"
	request := node.NewXMLS(query)
	request.NewChildS("max-records", batchSize)
	desired := node.NewXMLS("desired-attributes")
	volumeAttributes := node.NewXMLS("desired-attributes")
	volumeIDAttributes := node.NewXMLS("volume-id-attributes")
	volumeIDAttributes.NewChildS("name", "")
	volumeIDAttributes.NewChildS("owning-vserver-name", "")
	volumeIDAttributes.NewChildS("style-extended", "")
	volumeAttributes.AddChild(volumeIDAttributes)
	desired.AddChild(volumeAttributes)
	request.AddChild(desired)

	for {
		responseData, err := v.client.InvokeBatchRequest(request, tag, "")
		if err != nil {
			v.SLogger.Error("Failed to fetch data", slog.Any("err", err))
			return nil, err
		}
		result = responseData.Result
		tag = responseData.Tag

		if result == nil {
			break
		}

		if x := result.GetChildS("attributes-list"); x != nil {
			volumes = x.GetChildren()
		}
		if len(volumes) == 0 {
			return volumesMap, nil
		}

		for _, volume := range volumes {
			styleExtended := volume.GetChildS("volume-id-attributes").GetChildContentS("style-extended")
			name := volume.GetChildS("volume-id-attributes").GetChildContentS("name")
			svm := volume.GetChildS("volume-id-attributes").GetChildContentS("owning-vserver-name")
			volumesMap[svm+name] = styleExtended
		}
	}

	return volumesMap, nil
}
