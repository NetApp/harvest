package aggregate

import (
	"errors"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strings"
)

type Aggregate struct {
	*plugin.AbstractPlugin
	client             *zapi.Client
	aggrCloudStoresMap map[string][]string // aggregate-uuid -> slice of cloud stores map
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Aggregate{AbstractPlugin: p}
}

func (a *Aggregate) Init() error {

	var err error

	if err = a.InitAbc(); err != nil {
		return err
	}

	if a.client, err = zapi.New(conf.ZapiPoller(a.ParentParams), a.Auth); err != nil {
		a.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = a.client.Init(5); err != nil {
		return err
	}

	a.aggrCloudStoresMap = make(map[string][]string)
	return nil
}

func (a *Aggregate) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[a.Object]

	// invoke aggr-object-store-get-iter zapi and populate cloud stores info
	if err := a.getCloudStores(); err != nil {
		if errors.Is(err, errs.ErrNoInstance) {
			a.Logger.Debug().Err(err).Msg("Failed to collect cloud store data")
			return nil, nil
		}
		return nil, err
	}

	// update aggregate instance label with cloud stores info
	if len(a.aggrCloudStoresMap) > 0 {
		for uuid, aggr := range data.GetInstances() {
			if !aggr.IsExportable() {
				continue
			}
			aggr.SetLabel("cloud_stores", strings.Join(a.aggrCloudStoresMap[uuid], ","))
		}
	}
	return nil, nil
}

func (a *Aggregate) getCloudStores() error {
	var (
		result []*node.Node
		err    error
	)

	a.aggrCloudStoresMap = make(map[string][]string)
	request := node.NewXMLS("aggr-object-store-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)

	desired := node.NewXMLS("desired-attributes")
	objectStoreInfo := node.NewXMLS("object-store-information")
	objectStoreInfo.NewChildS("aggregate-uuid", "")
	objectStoreInfo.NewChildS("object-store-name", "")
	desired.AddChild(objectStoreInfo)
	request.AddChild(desired)

	if result, err = a.client.InvokeZapiCall(request); err != nil {
		return err
	}

	if len(result) == 0 || result == nil {
		return errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, objectStore := range result {
		aggregateUUID := objectStore.GetChildContentS("aggregate-uuid")
		objectStoreName := objectStore.GetChildContentS("object-store-name")
		a.aggrCloudStoresMap[aggregateUUID] = append(a.aggrCloudStoresMap[aggregateUUID], objectStoreName)
	}
	return nil
}
