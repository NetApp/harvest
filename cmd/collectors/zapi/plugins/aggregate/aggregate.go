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
	currentVal         int
	client             *zapi.Client
	aggrCloudStoresMap map[string][]string // aggregate-uuid -> slice of cloud stores map
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Aggregate{AbstractPlugin: p}
}

func (my *Aggregate) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	if my.client, err = zapi.New(conf.ZapiPoller(my.ParentParams), my.Auth); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.aggrCloudStoresMap = make(map[string][]string)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	my.currentVal = my.SetPluginInterval()
	return nil
}

func (my *Aggregate) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {

	data := dataMap[my.Object]
	if my.currentVal >= my.PluginInvocationRate {
		my.currentVal = 0

		// invoke aggr-object-store-get-iter zapi and populate cloud stores info
		if err := my.getCloudStores(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect cloud store data")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect cloud store data")
			}
		}
	}

	// update aggregate instance label with cloud stores
	for aggrUUID, aggr := range data.GetInstances() {
		aggr.SetLabel("cloud_stores", strings.Join(my.aggrCloudStoresMap[aggrUUID], ","))
	}

	my.currentVal++
	return nil, nil
}

func (my *Aggregate) getCloudStores() error {
	var (
		result []*node.Node
		err    error
	)

	my.aggrCloudStoresMap = make(map[string][]string)
	request := node.NewXMLS("aggr-object-store-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
		return err
	}

	if len(result) == 0 || result == nil {
		return errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, objectStore := range result {
		aggregateUUID := objectStore.GetChildContentS("aggregate-uuid")
		objectStoreName := objectStore.GetChildContentS("object-store-name")
		my.aggrCloudStoresMap[aggregateUUID] = append(my.aggrCloudStoresMap[aggregateUUID], objectStoreName)
	}
	return nil
}
