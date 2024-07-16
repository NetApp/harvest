package aggregate

import (
	"fmt"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
)

const (
	aggrObjectStoreMatrix = "aggr_object_store"
	apiQuery              = "api/private/cli/aggr/show-space"
)

var metrics = []string{
	"logical_used",
	"physical_used",
}

type Aggregate struct {
	*plugin.AbstractPlugin
	client *rest.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Aggregate{AbstractPlugin: p}
}

func (a *Aggregate) Init() error {
	if err := a.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	client, err := rest.New(conf.ZapiPoller(a.ParentParams), timeout, a.Auth)
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}
	a.client = client

	if err := a.client.Init(5); err != nil {
		return fmt.Errorf("failed to initialize REST client: %w", err)
	}

	return nil
}

func (a *Aggregate) initMatrix(name string, data *matrix.Matrix) (*matrix.Matrix, error) {
	mat := matrix.New(a.Parent+name, name, name)
	dataExportOptions := data.GetExportOptions()
	iKeys := dataExportOptions.GetChildS("instance_keys")
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")

	if iKeys != nil {
		for _, child := range iKeys.GetChildren() {
			instanceKeys.NewChildS(child.GetNameS(), child.GetContentS())
		}
	}

	instanceKeys.NewChildS("", "bin_num")
	instanceKeys.NewChildS("", "tier")
	mat.SetExportOptions(exportOptions)

	for _, k := range metrics {
		if err := matrix.CreateMetric(k, mat); err != nil {
			return nil, fmt.Errorf("error while creating metric %s: %w", k, err)
		}
	}

	return mat, nil
}

func (a *Aggregate) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[a.Object]
	a.client.Metadata.Reset()

	aggrSpaceMat, err := a.initMatrix(aggrObjectStoreMatrix, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error while initializing matrix: %w", err)
	}

	// Set all global labels if they do not already exist
	aggrSpaceMat.SetGlobalLabels(data.GetGlobalLabels())

	a.collectObjectStoreData(aggrSpaceMat, data)

	return []*matrix.Matrix{aggrSpaceMat}, a.client.Metadata, nil
}

func (a *Aggregate) collectObjectStoreData(aggrSpaceMat, data *matrix.Matrix) {
	records, err := a.getObjectStoreData()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			a.Logger.Debug().Err(err).Msg("API not found")
		} else {
			a.Logger.Error().Err(err).Msg("Failed to collect object store data")
		}
		return
	}

	uuidLookup := make(map[string]string)
	// Create name-to-UUID map for instance lookup
	for _, aggrInstance := range data.GetInstances() {
		if !aggrInstance.IsExportable() {
			continue
		}
		uuid := aggrInstance.GetLabel("uuid")
		aggr := aggrInstance.GetLabel("aggr")
		uuidLookup[aggr] = uuid
	}

	for _, record := range records {
		aggrName := record.Get("aggregate_name").String()
		uuid, has := uuidLookup[aggrName]
		if !has {
			continue
		}

		binNum := record.Get("bin_num").String()
		tierName := record.Get("tier_name").String()
		logicalUsed := record.Get("object_store_logical_used").String()
		physicalUsed := record.Get("object_store_physical_used").String()
		instanceKey := aggrName + "_" + tierName + "_" + binNum

		instance, err := aggrSpaceMat.NewInstance(instanceKey)
		if err != nil {
			a.Logger.Warn().Str("key", instanceKey).Msg("error while creating instance")
			continue
		}

		instance.SetLabels(data.GetInstance(uuid).GetLabels())
		instance.SetLabel("tier", tierName)
		instance.SetLabel("bin_num", binNum)

		if logicalUsed != "" {
			if err := aggrSpaceMat.GetMetric("logical_used").SetValueString(instance, logicalUsed); err != nil {
				a.Logger.Error().Err(err).Str("metric", "logical_used").Msg("Unable to set value on metric")
			}
		}

		if physicalUsed != "" {
			if err := aggrSpaceMat.GetMetric("physical_used").SetValueString(instance, physicalUsed); err != nil {
				a.Logger.Error().Err(err).Str("metric", "physical_used").Msg("Unable to set value on metric")
			}
		}
	}
}

func (a *Aggregate) getObjectStoreData() ([]gjson.Result, error) {
	fields := []string{"aggregate_name", "bin_num", "tier_name", "object_store_logical_used", "object_store_physical_used"}
	href := rest.NewHrefBuilder().
		APIPath(apiQuery).
		Fields(fields).
		Filter([]string{`tier_name=!" "|""`}).
		Build()

	return collectors.InvokeRestCall(a.client, href, a.Logger)
}
