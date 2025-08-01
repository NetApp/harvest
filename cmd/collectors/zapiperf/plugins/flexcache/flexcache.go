package flexcache

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

const (
	token = "#"
)

type FlexCache struct {
	*plugin.AbstractPlugin
	client *zapi.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &FlexCache{AbstractPlugin: p}
}

func (f *FlexCache) Init(remote conf.Remote) error {

	var err error

	if err := f.InitAbc(); err != nil {
		return err
	}

	if f.client, err = zapi.New(conf.ZapiPoller(f.ParentParams), f.Auth); err != nil {
		f.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := f.client.Init(5, remote); err != nil {
		return err
	}
	return nil
}

func (f *FlexCache) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[f.Object]
	f.client.Metadata.Reset()

	flexCache, err := f.getFlexCaches()
	if err != nil {
		return nil, nil, err
	}
	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		svm := instance.GetLabel("svm")
		volume := instance.GetLabel("volume")
		s := svm + token + volume
		// Handles the scenario where the performance call returns records even though the config call does not.
		if !flexCache.Has(s) {
			instance.SetExportable(false)
		}
	}
	return nil, f.client.Metadata, nil
}

func (f *FlexCache) getFlexCaches() (*set.Set, error) {

	var (
		result       *node.Node
		flexCaches   []*node.Node
		flexCacheSet *set.Set
	)

	flexCacheSet = set.New()
	query := "flexcache-get-iter"
	tag := "initial"
	request := node.NewXMLS(query)
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	desired := node.NewXMLS("desired-attributes")
	flexCacheInfo := node.NewXMLS("flexcache-info")
	flexCacheInfo.NewChildS("volume", "")
	flexCacheInfo.NewChildS("vserver", "")
	desired.AddChild(flexCacheInfo)
	request.AddChild(desired)

	for {
		responseData, err := f.client.InvokeBatchRequest(request, tag, "")
		if err != nil {
			return nil, err
		}
		result = responseData.Result
		tag = responseData.Tag

		if result == nil {
			break
		}

		if x := result.GetChildS("attributes-list"); x != nil {
			flexCaches = x.GetChildren()
		}
		if len(flexCaches) == 0 {
			// Handles the case where Perf call has records but Config call doesn't.
			break
		}

		for _, flexCache := range flexCaches {
			volume := flexCache.GetChildS("volume")
			svm := flexCache.GetChildS("vserver")
			flexCacheSet.Add(svm.GetContentS() + token + volume.GetContentS())
		}
	}
	return flexCacheSet, nil
}
