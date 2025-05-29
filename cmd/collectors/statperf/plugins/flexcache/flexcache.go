package flexcache

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"time"
)

const (
	token = "#"
)

type FlexCache struct {
	*plugin.AbstractPlugin
	client *rest.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &FlexCache{AbstractPlugin: p}
}

func (f *FlexCache) Init(remote conf.Remote) error {

	var err error

	if err := f.InitAbc(); err != nil {
		return err
	}

	if f.Options.IsTest {
		return nil
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if f.client, err = rest.New(conf.ZapiPoller(f.ParentParams), timeout, f.Auth); err != nil {
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
	flexCacheSet := set.New()
	query := "api/storage/flexcache/flexcaches"
	fields := []string{"uuid,name,svm.name"}
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Build()

	records, err := collectors.InvokeRestCall(f.client, href)
	if err != nil {
		return flexCacheSet, err
	}
	for _, record := range records {
		svm := record.Get("svm.name").ClonedString()
		volume := record.Get("name").ClonedString()
		flexCacheSet.Add(svm + token + volume)
	}
	return flexCacheSet, nil
}
