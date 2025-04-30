package aggregate

import (
	"errors"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	goversion "github.com/netapp/harvest/v2/third_party/go-version"
	"log/slog"
	"strconv"
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

func (a *Aggregate) Init(remote conf.Remote) error {

	var err error

	if err := a.InitAbc(); err != nil {
		return err
	}

	if a.client, err = zapi.New(conf.ZapiPoller(a.ParentParams), a.Auth); err != nil {
		a.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := a.client.Init(5, remote); err != nil {
		return err
	}

	a.aggrCloudStoresMap = make(map[string][]string)
	return nil
}

func (a *Aggregate) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[a.Object]
	a.client.Metadata.Reset()

	// invoke aggr-object-store-get-iter zapi and populate cloud stores info
	if err := a.getCloudStores(); err != nil {
		if !errors.Is(err, errs.ErrNoInstance) {
			a.SLogger.Error("Failed to update get cloud stores", slogx.Err(err))
		}
	}

	aggrFootprintMap, err := a.getAggrFootprint()
	if err != nil {
		a.SLogger.Error("Failed to update footprint data", slogx.Err(err))
		// clean the map in case of the error
		clear(aggrFootprintMap)
	}

	// update aggregate instance label with cloud stores info
	for uuid, aggr := range data.GetInstances() {
		if !aggr.IsExportable() {
			continue
		}
		aggr.SetLabel("cloud_stores", strings.Join(a.aggrCloudStoresMap[uuid], ","))

		// Handling aggr footprint metrics
		aggrName := aggr.GetLabel("aggr")
		if af, ok := aggrFootprintMap[aggrName]; ok {
			for afKey, afVal := range af {
				vfMetric := data.GetMetric(afKey)
				if vfMetric == nil {
					if vfMetric, err = data.NewMetricFloat64(afKey); err != nil {
						a.SLogger.Error("add metric", slogx.Err(err), slog.String("metric", afKey))
						continue
					}
				}

				if afVal != "" {
					vfMetricVal, err := strconv.ParseFloat(afVal, 64)
					if err != nil {
						a.SLogger.Error("parse", slogx.Err(err), slog.String(afKey, afVal))
						continue
					}
					vfMetric.SetValueFloat64(aggr, vfMetricVal)
				}
			}
		}
	}
	return nil, a.client.Metadata, nil
}

func (a *Aggregate) getCloudStores() error {
	var (
		result []*node.Node
		err    error
	)

	// aggr-object-store-get-iter Zapi was introduced in 9.2.
	clusterVersion := a.client.Version()
	ontapVersion, err := goversion.NewVersion(clusterVersion)
	if err != nil {
		a.SLogger.Error(
			"Failed to parse version",
			slogx.Err(err),
			slog.String("version", clusterVersion),
		)
		return err
	}
	version92 := "9.2"
	version92After, _ := goversion.NewVersion(version92)

	if ontapVersion.LessThan(version92After) {
		return nil
	}

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

func (a *Aggregate) getAggrFootprint() (map[string]map[string]string, error) {
	var (
		result           []*node.Node
		aggrFootprintMap map[string]map[string]string
		err              error
	)

	aggrFootprintMap = make(map[string]map[string]string)
	request := node.NewXMLS("aggr-space-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	desired := node.NewXMLS("desired-attributes")
	spaceInfo := node.NewXMLS("space-information")
	spaceInfo.NewChildS("aggregate", "")
	spaceInfo.NewChildS("volume-footprints", "")
	spaceInfo.NewChildS("volume-footprints-percent", "")
	desired.AddChild(spaceInfo)
	request.AddChild(desired)

	if result, err = a.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return aggrFootprintMap, nil
	}

	for _, footprint := range result {
		footprintMetrics := make(map[string]string)
		aggr := footprint.GetChildContentS("aggregate")
		performanceTierUsed := footprint.GetChildContentS("volume-footprints")
		performanceTierUsedPerc := footprint.GetChildContentS("volume-footprints-percent")
		if performanceTierUsed != "" || performanceTierUsedPerc != "" {
			footprintMetrics["space_performance_tier_used"] = performanceTierUsed
			footprintMetrics["space_performance_tier_used_percent"] = performanceTierUsedPerc
			aggrFootprintMap[aggr] = footprintMetrics
		}
	}

	return aggrFootprintMap, nil
}
