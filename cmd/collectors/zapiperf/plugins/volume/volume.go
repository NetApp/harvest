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
	styleType           string
	includeConstituents bool
	client              *zapi.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (v *Volume) Init() error {
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

	// Read template to decide inclusion of flexgroup constituents
	v.includeConstituents = collectors.ReadPluginKey(v.Params, "include_constituents")

	if v.client, err = zapi.New(conf.ZapiPoller(v.ParentParams), v.Auth); err != nil {
		v.SLogger.Error("connecting", slog.Any("err", err))
		return err
	}
	return v.client.Init(5)
}

func (v *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[v.Object]
	style := v.styleType
	opsKeyPrefix := "temp_"
	volumesMap := v.fetchVolumes()

	return collectors.ProcessFlexGroupData(v.SLogger, data, style, v.includeConstituents, opsKeyPrefix, volumesMap)
}

func (v *Volume) fetchVolumes() map[string]string {
	var (
		result     *node.Node
		volumes    []*node.Node
		volumesMap map[string]string
	)

	volumesMap = make(map[string]string)
	query := "volume-get-iter"
	tag := "initial"
	request := node.NewXMLS(query)
	request.NewChildS("max-records", batchSize)
	desired := node.NewXMLS("desired-attributes")
	volumeAttributes := node.NewXMLS("desired-attributes")
	volumeIDAttributes := node.NewXMLS("volume-id-attributes")
	volumeIDAttributes.NewChildS("name", "")
	volumeIDAttributes.NewChildS("style-extended", "")
	volumeAttributes.AddChild(volumeIDAttributes)
	desired.AddChild(volumeAttributes)
	request.AddChild(desired)

	for {
		responseData, err := v.client.InvokeBatchRequest(request, tag, "")
		if err != nil {
			return nil
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
			return nil
		}

		for _, volume := range volumes {
			styleExtended := volume.GetChildS("volume-id-attributes").GetChildContentS("style-extended")
			name := volume.GetChildS("volume-id-attributes").GetChildContentS("name")
			volumesMap[name] = styleExtended
		}
	}

	return volumesMap
}
