package volumetag

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

const batchSize = "500"

type VolumeTag struct {
	*plugin.AbstractPlugin
	client *zapi.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &VolumeTag{AbstractPlugin: p}
}

func (v *VolumeTag) Init(remote conf.Remote) error {
	var err error
	if err := v.InitAbc(); err != nil {
		return err
	}

	if v.client, err = zapi.New(conf.ZapiPoller(v.ParentParams), v.Auth); err != nil {
		v.SLogger.Error("connecting", slogx.Err(err))
		return err
	}
	return v.client.Init(5, remote)
}

func (v *VolumeTag) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {

	var (
		result  *node.Node
		volumes []*node.Node
	)

	data := dataMap[v.Object]
	v.client.Metadata.Reset()

	query := "volume-get-iter"
	tag := "initial"
	request := node.NewXMLS(query)
	request.NewChildS("max-records", batchSize)
	desired := node.NewXMLS("desired-attributes")
	volumeAttributes := node.NewXMLS("desired-attributes")
	volumeIDAttributes := node.NewXMLS("volume-id-attributes")
	volumeIDAttributes.NewChildS("comment", "")
	volumeIDAttributes.NewChildS("instance-uuid", "")
	volumeAttributes.AddChild(volumeIDAttributes)
	desired.AddChild(volumeAttributes)
	request.AddChild(desired)

	for {
		responseData, err := v.client.InvokeBatchRequest(request, tag, "")
		if err != nil {
			return nil, nil, err
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
			return nil, nil, nil
		}

		for _, volume := range volumes {
			key := volume.GetChildS("volume-id-attributes").GetChildContentS("instance-uuid")
			comment := volume.GetChildS("volume-id-attributes").GetChildContentS("comment")
			instance := data.GetInstance(key)
			if instance != nil && comment != "" {
				instance.SetLabel("comment", comment)
			}
		}
	}

	if exportOption := v.ParentParams.GetChildS("export_options"); exportOption != nil {
		if exportedKeys := exportOption.GetChildS("instance_keys"); exportedKeys != nil {
			if exportedKeys.GetChildByContent("comment") == nil {
				exportedKeys.NewChildS("", "comment")
			}
		}
	}

	return nil, v.client.Metadata, nil
}
