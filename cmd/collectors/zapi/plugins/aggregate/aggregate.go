package aggregate

import (
	"errors"
	goversion "github.com/hashicorp/go-version"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
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

	aggrFootprintMap, err := a.getAggrFootprint()
	if err != nil {
		a.Logger.Error().Err(err).Msg("Failed to update footprint data")
	}

	// update aggregate instance label with cloud stores info
	for aggrUUID, aggr := range data.GetInstances() {
		if !aggr.IsExportable() {
			continue
		}
		aggr.SetLabel("cloud_stores", strings.Join(a.aggrCloudStoresMap[aggrUUID], ","))

		// Handling aggr footprint metrics
		aggrName := aggr.GetLabel("aggr")
		if af, ok := aggrFootprintMap[aggrName]; ok {
			for afKey, afVal := range af {
				vfMetric := data.GetMetric(afKey)
				if vfMetric == nil {
					if vfMetric, err = data.NewMetricFloat64(afKey); err != nil {
						a.Logger.Error().Err(err).Str("metric", afKey).Msg("add metric")
						continue
					}
				}

				if afVal != "" {
					vfMetricVal, err := strconv.ParseFloat(afVal, 64)
					if err != nil {
						a.Logger.Error().Err(err).Str(afKey, afVal).Msg("parse")
						continue
					}
					if err = vfMetric.SetValueFloat64(aggr, vfMetricVal); err != nil {
						a.Logger.Error().Err(err).Str(afKey, afVal).Msg("set")
						continue
					}
				}
			}
		}
	}
	return nil, nil
}

func (a *Aggregate) getCloudStores() error {
	var (
		result []*node.Node
		err    error
	)

	// aggr-object-store-get-iter Zapi was introduced in 9.2.
	version := a.client.Version()
	clusterVersion := strconv.Itoa(version[0]) + "." + strconv.Itoa(version[1]) + "." + strconv.Itoa(version[2])
	ontapVersion, err := goversion.NewVersion(clusterVersion)
	if err != nil {
		a.Logger.Error().Err(err).
			Str("version", clusterVersion).
			Msg("Failed to parse version")
		return err
	}
	version92 := "9.2"
	version92After, err := goversion.NewVersion(version92)
	if err != nil {
		a.Logger.Error().Err(err).
			Str("version", version92).
			Msg("Failed to parse version")
		return err
	}

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
		footprintMatrics map[string]string
		err              error
	)

	aggrFootprintMap = make(map[string]map[string]string)
	request := node.NewXMLS("aggr-space-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	if result, err = a.client.InvokeZapiCall(request); err != nil {
		return aggrFootprintMap, err
	}

	if len(result) == 0 || result == nil {
		return aggrFootprintMap, nil
	}

	for _, footprint := range result {
		footprintMatrics = make(map[string]string)
		aggr := footprint.GetChildContentS("aggregate")
		performanceTierUsed := footprint.GetChildContentS("volume-footprints")
		performanceTierUsedPerc := footprint.GetChildContentS("volume-footprints-percent")
		if performanceTierUsed != "" || performanceTierUsedPerc != "" {
			footprintMatrics["space_performance_tier_used"] = performanceTierUsed
			footprintMatrics["space_performance_tier_used_percent"] = performanceTierUsedPerc
			aggrFootprintMap[aggr] = footprintMatrics
		}
	}

	return aggrFootprintMap, nil
}
