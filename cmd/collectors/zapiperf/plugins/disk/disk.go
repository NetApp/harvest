// Package disk Copyright NetApp Inc, 2021 All rights reserved
package disk

import (
	"context"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/num"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/template"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"strings"
)

const batchSize = "500"

type RaidAggrDerivedType string
type RaidAggrType string

const (
	radtHDD              RaidAggrDerivedType = "hdd"
	radtHDDFABRICPOOL    RaidAggrDerivedType = "hdd_fabricpool"
	radtSSD              RaidAggrDerivedType = "ssd"
	radtSSDFABRICPOOL    RaidAggrDerivedType = "ssd_fabricpool"
	radtHYBRID           RaidAggrDerivedType = "hybrid"
	radtHYBRIDFLASHPOOL  RaidAggrDerivedType = "hybrid_flash_pool"
	radtLUNFLEXARRAY     RaidAggrDerivedType = "lun_flexarray"
	radtVMDISKSDS        RaidAggrDerivedType = "vmdisk_sds"
	radtVMDISKFABRICPOOL RaidAggrDerivedType = "vmdisk_fabricpool"
	radtNotMapped        RaidAggrDerivedType = "not_mapped"
)

const (
	ratHDD    RaidAggrType = "hdd"
	ratSSD    RaidAggrType = "ssd"
	ratHYBRID RaidAggrType = "hybrid"
	ratLUN    RaidAggrType = "lun"
	ratVMDISK RaidAggrType = "vmdisk"
)

type Disk struct {
	*plugin.AbstractPlugin
	shelfData      map[string]*matrix.Matrix
	powerData      map[string]*matrix.Matrix
	instanceKeys   map[string][]string
	instanceLabels map[string]map[string]string
	batchSize      string
	client         *zapi.Client
	query          string
	aggrMap        map[string]*aggregate
	diskMap        map[string]*disk  // disk UID to disk info containing shelf name
	ShelfMap       map[string]*shelf // shelf id to power mapping
}

type shelf struct {
	iops  float64
	power float64
	disks []*disk
}

type aggregate struct {
	name        string
	node        string
	isShared    bool
	power       float64
	derivedType RaidAggrDerivedType
}

type disk struct {
	name       string
	shelfID    string
	id         string
	diskType   string
	aggregates []string
}

type shelfEnvironmentMetric struct {
	key                   string
	ambientTemperature    []float64
	nonAmbientTemperature []float64
	fanSpeed              []float64
	voltageSensor         map[string]float64
	currentSensor         map[string]float64
}

var shelfMetrics = []string{
	"average_ambient_temperature",
	"average_fan_speed",
	"average_temperature",
	"max_fan_speed",
	"max_temperature",
	"min_ambient_temperature",
	"min_fan_speed",
	"min_temperature",
	"power",
}

var aggrMetrics = []string{
	"power",
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Disk{AbstractPlugin: p}
}

func (d *Disk) Init(remote conf.Remote) error {

	var err error

	if err := d.InitAbc(); err != nil {
		return err
	}

	if d.client, err = zapi.New(conf.ZapiPoller(d.ParentParams), d.Auth); err != nil {
		d.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := d.client.Init(5, remote); err != nil {
		return err
	}

	d.query = "storage-shelf-info-get-iter"

	d.shelfData = make(map[string]*matrix.Matrix)
	d.powerData = make(map[string]*matrix.Matrix)

	d.instanceKeys = make(map[string][]string)
	d.instanceLabels = make(map[string]map[string]string)

	objects := d.Params.GetChildS("objects")
	if objects == nil {
		return errs.New(errs.ErrMissingParam, "objects")
	}

	for _, obj := range objects.GetChildren() {

		attribute := obj.GetNameS()
		objectName := strings.ReplaceAll(attribute, "-", "_")

		if x := strings.Split(attribute, "=>"); len(x) == 2 {
			attribute = strings.TrimSpace(x[0])
			objectName = strings.TrimSpace(x[1])
		}

		d.instanceLabels[attribute] = make(map[string]string)

		d.shelfData[attribute] = matrix.New(d.Parent+".Shelf", "shelf_"+objectName, "shelf_"+objectName)
		d.shelfData[attribute].SetGlobalLabel("datacenter", d.ParentParams.GetChildContentS("datacenter"))

		exportOptions := node.NewS("export_options")
		instanceLabels := exportOptions.NewChildS("instance_labels", "")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")
		instanceKeys.NewChildS("", "shelf")
		instanceKeys.NewChildS("", "channel")

		// artificial metric for status of child object of shelf
		_, _ = d.shelfData[attribute].NewMetricUint8("status")

		for _, x := range obj.GetChildren() {

			for _, c := range x.GetAllChildContentS() {

				metricName, display, kind, _ := template.ParseMetric(c)

				switch kind {
				case "key":
					d.instanceKeys[attribute] = append(d.instanceKeys[attribute], metricName)
					d.instanceLabels[attribute][metricName] = display
					instanceLabels.NewChildS("", display)
					instanceKeys.NewChildS("", display)
				case "label":
					d.instanceLabels[attribute][metricName] = display
					instanceLabels.NewChildS("", display)
				case "float":
					_, err := d.shelfData[attribute].NewMetricFloat64(metricName, display)
					if err != nil {
						d.SLogger.Error("add metric", slogx.Err(err))
						return err
					}
				}
			}
		}

		d.shelfData[attribute].SetExportOptions(exportOptions)
	}

	d.SLogger.Debug("initialized with shelfData", slog.Int("objects", len(d.shelfData)))

	// setup batchSize for request
	d.batchSize = batchSize
	if b := d.Params.GetChildContentS("batch_size"); b != "" {
		if _, err := strconv.Atoi(b); err == nil {
			d.batchSize = b
		}
	}

	d.initShelfPowerMatrix()
	d.initAggrPowerMatrix()

	d.initMaps()

	return nil
}

func (d *Disk) initShelfPowerMatrix() {
	d.powerData["shelf"] = matrix.New(d.Parent+".Shelf", "shelf", "shelf")

	for _, k := range shelfMetrics {
		err := matrix.CreateMetric(k, d.powerData["shelf"])
		if err != nil {
			d.SLogger.Warn("create metric", slogx.Err(err), slog.String("key", k))
		}
	}
}

func (d *Disk) initAggrPowerMatrix() {
	d.powerData["aggr"] = matrix.New(d.Parent+".Aggr", "aggr", "aggr")

	for _, k := range aggrMetrics {
		err := matrix.CreateMetric(k, d.powerData["aggr"])
		if err != nil {
			d.SLogger.Warn("create metric", slogx.Err(err), slog.String("key", k))
		}
	}
}

func (d *Disk) initMaps() {
	// reset shelf Power
	d.ShelfMap = make(map[string]*shelf)

	// reset diskmap
	d.diskMap = make(map[string]*disk)

	// reset aggrmap
	d.aggrMap = make(map[string]*aggregate)
}

func (d *Disk) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {

	var (
		err    error
		output []*matrix.Matrix
	)
	data := dataMap[d.Object]
	d.client.Metadata.Reset()

	// Set all global labels from zapi.go if already not exist
	for a := range d.instanceLabels {
		d.shelfData[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	for a := range d.powerData {
		d.powerData[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	request := node.NewXMLS(d.query)
	// Fetching only local shelves
	query := request.NewChildS("query", "")
	storageShelfInfo := query.NewChildS("storage-shelf-info", "")
	storageShelfInfo.NewChildS("is-local-attach", "true")
	if d.client.IsClustered() {
		request.NewChildS("max-records", d.batchSize)
	}

	result, err := d.client.InvokeZapiCall(request)
	if err != nil {
		return nil, nil, err
	}

	d.initMaps()

	output, err = d.handleCMode(result)
	if err != nil {
		return output, nil, err
	}

	output, err = d.handleShelfPower(result, output)
	if err != nil {
		return output, nil, err
	}

	err = d.getAggregates()
	if err != nil {
		return output, nil, err
	}

	err = d.getDisks()
	if err != nil {
		return output, nil, err
	}

	err = d.populateShelfIOPS(data)
	if err != nil {
		return output, nil, err
	}

	output, err = d.calculateAggrPower(data, output)
	if err != nil {
		return output, nil, err
	}

	return output, d.client.Metadata, nil
}

func (d *Disk) calculateAggrPower(data *matrix.Matrix, output []*matrix.Matrix) ([]*matrix.Matrix, error) {

	totalTransfers := data.GetMetric("total_transfers")
	if totalTransfers == nil {
		return output, errs.New(errs.ErrNoMetric, "total_transfers")
	}

	// calculate power for returned disks in zapiperf response
	for key, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		if v, ok := totalTransfers.GetValueFloat64(instance); ok {
			diskUUID := instance.GetLabel("disk_uuid")
			aggrName := instance.GetLabel("aggr")
			a, ok := d.aggrMap[aggrName]
			if !ok {
				d.SLogger.Warn("Missing Aggregate info", slog.String("aggrName", aggrName))
				continue
			}

			di, ok := d.diskMap[diskUUID]
			if ok {
				shelfID := di.shelfID
				sh, ok := d.ShelfMap[shelfID]
				if ok {
					diskPower := v * sh.power / sh.iops
					a.power += diskPower
				}
			} else {
				d.SLogger.Warn("Missing disk info", slog.String("diskUUID", diskUUID))
			}
		} else {
			d.SLogger.Debug("Instance not exported", slog.String("key", key))
		}
	}

	// If the storage shelf total IOPS is 0, we can distribute the shelf power to each disk evenly as the disks still consume power when idle.
	// If disks are spare then they will not have aggregates, and we can not calculate aggr_power in such cases. In such cases sum of aggr_power for a cluster will not match with sum shelf_power
	for _, v := range d.ShelfMap {
		if v.iops == 0 && v.power > 0 && len(v.disks) > 0 {
			// counts disks with aggregate names
			diskWithAggregateCount := 0
			for _, v1 := range v.disks {
				c := len(v1.aggregates)
				if c > 0 {
					diskWithAggregateCount++
				}
			}
			if diskWithAggregateCount != 0 {
				powerPerDisk := v.power / float64(diskWithAggregateCount)
				for _, v1 := range v.disks {
					if len(v1.aggregates) > 0 {
						powerPerAggregate := powerPerDisk / float64(len(v1.aggregates))
						for _, a1 := range v1.aggregates {
							a, ok := d.aggrMap[a1]
							if !ok {
								d.SLogger.Warn("Missing Aggregate info", slog.String("aggrName", a1))
								continue
							}
							a.power += powerPerAggregate
						}
					}
				}
			}
		}
	}

	// Purge and reset data
	aggrData := d.powerData["aggr"]
	aggrData.PurgeInstances()
	aggrData.Reset()

	// fill aggr power matrix with power calculated above
	for instanceKey, v := range d.aggrMap {
		instance, err := aggrData.NewInstance(instanceKey)
		if err != nil {
			d.SLogger.Error("Failed to add instance", slogx.Err(err), slog.String("key", instanceKey))
			continue
		}
		instance.SetLabel("aggr", instanceKey)
		instance.SetLabel("derivedType", string(v.derivedType))
		instance.SetLabel("node", v.node)

		m := aggrData.GetMetric("power")
		m.SetValueFloat64(instance, v.power)
	}
	output = append(output, aggrData)
	return output, nil

}

func (d *Disk) populateShelfIOPS(data *matrix.Matrix) error {

	totalTransfers := data.GetMetric("total_transfers")
	if totalTransfers == nil {
		return errs.New(errs.ErrNoMetric, "total_transfers")
	}

	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		if v, ok := totalTransfers.GetValueFloat64(instance); ok {
			diskUUID := instance.GetLabel("disk_uuid")
			di, ok := d.diskMap[diskUUID]
			if ok {
				shelfID := di.shelfID
				sh, ok := d.ShelfMap[shelfID]
				if ok {
					sh.iops += v
				}
			} else {
				d.SLogger.Warn("Missing disk info", slog.String("diskUUID", diskUUID))
			}
		}
	}
	return nil
}

func (d *Disk) getDisks() error {

	var (
		result *node.Node
		disks  []*node.Node
	)

	query := "storage-disk-get-iter"
	tag := "initial"
	request := node.NewXMLS(query)
	request.NewChildS("max-records", d.batchSize)
	desired := node.NewXMLS("desired-attributes")
	storageDiskInfo := node.NewXMLS("storage-disk-info")
	diskInventoryInfo := node.NewXMLS("disk-inventory-info")
	diskInventoryInfo.NewChildS("disk-uid", "")
	diskInventoryInfo.NewChildS("shelf", "")
	diskInventoryInfo.NewChildS("is-shared", "")
	diskInventoryInfo.NewChildS("disk-type", "")
	diskRaidInfo := node.NewXMLS("disk-raid-info")
	diskAggregateInfo := node.NewXMLS("disk-aggregate-info")
	diskSharedInfo := node.NewXMLS("disk-shared-info")
	diskRaidInfo.AddChild(diskAggregateInfo)
	diskRaidInfo.AddChild(diskSharedInfo)
	storageDiskInfo.AddChild(diskInventoryInfo)
	storageDiskInfo.AddChild(diskRaidInfo)
	desired.AddChild(storageDiskInfo)
	request.AddChild(desired)

	for {
		responseData, err := d.client.InvokeBatchRequest(request, tag, "")
		if err != nil {
			return err
		}
		result = responseData.Result
		tag = responseData.Tag

		if result == nil {
			break
		}

		if x := result.GetChildS("attributes-list"); x != nil {
			disks = x.GetChildren()
		}
		if len(disks) == 0 {
			return nil
		}

		for _, v := range disks {
			diskName := v.GetChildContentS("disk-name")
			dii := v.GetChildS("disk-inventory-info")
			dri := v.GetChildS("disk-raid-info")
			if dii != nil {
				diskUID := dii.GetChildContentS("disk-uid")
				shelfID := dii.GetChildContentS("shelf")
				isShared := dii.GetChildContentS("is-shared")
				diskType := dii.GetChildContentS("disk-type")
				var aggrNames []string
				if isShared == "true" {
					if dri != nil {
						dsi := dri.GetChildS("disk-shared-info")
						if dsi != nil {
							al := dsi.GetChildS("aggregate-list")
							if al != nil {
								sai := al.GetChildren()
								for _, s := range sai {
									an := s.GetAllChildContentS()
									aggrNames = append(aggrNames, an...)
								}
							}
						}
					}
				} else {
					if dri != nil {
						dai := dri.GetChildS("disk-aggregate-info")
						if dai != nil {
							aggrName := dai.GetChildContentS("aggregate-name")
							aggrNames = append(aggrNames, aggrName)
						}
					}
				}
				dis := &disk{
					name:       diskName,
					shelfID:    shelfID,
					id:         diskUID,
					aggregates: aggrNames,
					diskType:   diskType,
				}
				d.diskMap[diskUID] = dis
				sh, ok := d.ShelfMap[shelfID]
				if ok {
					sh.disks = append(sh.disks, dis)
				}
			}
		}
	}
	return nil
}

func (d *Disk) getAggregates() error {

	var (
		result *node.Node
		aggrs  []*node.Node
	)

	query := "aggr-get-iter"
	tag := "initial"
	request := node.NewXMLS(query)
	request.NewChildS("max-records", d.batchSize)
	desired := node.NewXMLS("desired-attributes")
	aggrAttributes := node.NewXMLS("aggr-attributes")
	aggrOwnerAttributes := node.NewXMLS("aggr-ownership-attributes")
	aggrOwnerAttributes.NewChildS("home-name", "")
	aggrRaidAttributes := node.NewXMLS("aggr-raid-attributes")
	aggrRaidAttributes.NewChildS("uses-shared-disks", "")
	aggrRaidAttributes.NewChildS("aggregate-type", "")
	aggrRaidAttributes.NewChildS("is-composite", "")
	aggrAttributes.AddChild(aggrRaidAttributes)
	aggrAttributes.AddChild(aggrOwnerAttributes)
	desired.AddChild(aggrAttributes)
	request.AddChild(desired)

	for {
		responseData, err := d.client.InvokeBatchRequest(request, tag, "")
		if err != nil {
			return err
		}
		result = responseData.Result
		tag = responseData.Tag

		if result == nil {
			break
		}

		if x := result.GetChildS("attributes-list"); x != nil {
			aggrs = x.GetChildren()
		}
		if len(aggrs) == 0 {
			return nil
		}

		for _, aggr := range aggrs {
			aggrName := aggr.GetChildContentS("aggregate-name")
			aggrRaidAttr := aggr.GetChildS("aggr-raid-attributes")
			aggrOwnerAttr := aggr.GetChildS("aggr-ownership-attributes")
			var nodeName string
			if aggrOwnerAttr != nil {
				nodeName = aggrOwnerAttr.GetChildContentS("home-name")
			}
			if aggrRaidAttr != nil {
				usesSharedDisks := aggrRaidAttr.GetChildContentS("uses-shared-disks")
				aggregateType := aggrRaidAttr.GetChildContentS("aggregate-type")
				isC := aggrRaidAttr.GetChildContentS("is-composite")
				isComposite := isC == "true"
				isShared := usesSharedDisks == "true"
				derivedType := getAggregateDerivedType(aggregateType, isComposite, isShared)
				d.aggrMap[aggrName] = &aggregate{
					name:        aggrName,
					isShared:    isShared,
					derivedType: derivedType,
					node:        nodeName,
				}
			}
		}
	}
	return nil
}

func getAggregateDerivedType(aggregateType string, isComposite bool, isShared bool) RaidAggrDerivedType {
	derivedType := radtNotMapped
	if aggregateType == "" {
		return derivedType
	}
	switch RaidAggrType(aggregateType) {
	case ratHDD:
		derivedType = radtHDD
		if isComposite {
			derivedType = radtHDDFABRICPOOL
		}
	case ratSSD:
		derivedType = radtSSD
		if isComposite {
			derivedType = radtSSDFABRICPOOL
		}
	case ratHYBRID:
		derivedType = radtHYBRID
		if isShared {
			derivedType = radtHYBRIDFLASHPOOL
		}
	case ratLUN:
		derivedType = radtLUNFLEXARRAY
	case ratVMDISK:
		derivedType = radtVMDISKSDS
		if isComposite {
			derivedType = radtVMDISKFABRICPOOL
		}
	}
	return derivedType
}

func (d *Disk) handleShelfPower(shelves []*node.Node, output []*matrix.Matrix) ([]*matrix.Matrix, error) {
	// Purge and reset data
	data := d.powerData["shelf"]
	data.PurgeInstances()
	data.Reset()

	for _, s := range shelves {
		shelfName := s.GetChildContentS("shelf")
		shelfUID := s.GetChildContentS("shelf-uid")
		shelfID := s.GetChildContentS("shelf-id")
		instanceKey := shelfUID
		instance, err := data.NewInstance(instanceKey)
		if err != nil {
			d.SLogger.Error("add instance", slogx.Err(err), slog.String("key", instanceKey))
			return output, err
		}
		instance.SetLabel("shelf", shelfName)
		instance.SetLabel("shelfID", shelfID)
	}
	d.calculateEnvironmentMetrics(data)

	output = append(output, data)
	return output, nil
}

func (d *Disk) calculateEnvironmentMetrics(data *matrix.Matrix) {
	shelfEnvironmentMetricMap := make(map[string]*shelfEnvironmentMetric)
	for _, o := range d.shelfData {
		for k, instance := range o.GetInstances() {
			if !instance.IsExportable() {
				continue
			}
			lastInd := strings.LastIndex(k, "#")
			iKey := k[:lastInd]
			iKey2 := k[lastInd+1:]
			if _, ok := shelfEnvironmentMetricMap[iKey]; !ok {
				shelfEnvironmentMetricMap[iKey] = &shelfEnvironmentMetric{key: iKey, ambientTemperature: []float64{}, nonAmbientTemperature: []float64{}, fanSpeed: []float64{}}
			}
			for mkey, metric := range o.GetMetrics() {
				switch o.Object {
				case "shelf_temperature":
					if mkey == "temp-sensor-reading" {
						isAmbient := instance.GetLabel("temp_is_ambient")
						if isAmbient == "true" {
							if value, ok := metric.GetValueFloat64(instance); ok {
								shelfEnvironmentMetricMap[iKey].ambientTemperature = append(shelfEnvironmentMetricMap[iKey].ambientTemperature, value)
							}
						}
						if isAmbient == "false" {
							if value, ok := metric.GetValueFloat64(instance); ok {
								shelfEnvironmentMetricMap[iKey].nonAmbientTemperature = append(shelfEnvironmentMetricMap[iKey].nonAmbientTemperature, value)
							}
						}
					}
				case "shelf_fan":
					if mkey == "fan-rpm" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							shelfEnvironmentMetricMap[iKey].fanSpeed = append(shelfEnvironmentMetricMap[iKey].fanSpeed, value)
						}
					}
				case "shelf_voltage":
					if mkey == "voltage-sensor-reading" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							if shelfEnvironmentMetricMap[iKey].voltageSensor == nil {
								shelfEnvironmentMetricMap[iKey].voltageSensor = make(map[string]float64)
							}
							shelfEnvironmentMetricMap[iKey].voltageSensor[iKey2] = value
						}
					}
				case "shelf_sensor":
					if mkey == "current-sensor-reading" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							if shelfEnvironmentMetricMap[iKey].currentSensor == nil {
								shelfEnvironmentMetricMap[iKey].currentSensor = make(map[string]float64)
							}
							shelfEnvironmentMetricMap[iKey].currentSensor[iKey2] = value
						}
					}
				}
			}
		}
	}

	for key, v := range shelfEnvironmentMetricMap {
		for _, k := range shelfMetrics {
			m := data.GetMetric(k)
			instance := data.GetInstance(key)
			if instance == nil {
				d.SLogger.Warn("Instance not found", slog.String("key", key))
				continue
			}
			switch k {
			case "power":
				var sumPower float64
				for k1, v1 := range v.voltageSensor {
					if v2, ok := v.currentSensor[k1]; ok {
						// in W
						sumPower += (v1 * v2) / 1000
					} else {
						d.SLogger.Warn("missing current sensor", slog.String("voltage sensor id", k1))
					}
				}

				m.SetValueFloat64(instance, sumPower)
				d.ShelfMap[instance.GetLabel("shelfID")] = &shelf{power: sumPower}

			case "average_ambient_temperature":
				if len(v.ambientTemperature) > 0 {
					aaT := num.Avg(v.ambientTemperature)
					m.SetValueFloat64(instance, aaT)
				}
			case "min_ambient_temperature":
				maT := num.Min(v.ambientTemperature)
				m.SetValueFloat64(instance, maT)
			case "max_temperature":
				mT := num.Max(v.nonAmbientTemperature)
				m.SetValueFloat64(instance, mT)
			case "average_temperature":
				if len(v.nonAmbientTemperature) > 0 {
					nat := num.Avg(v.nonAmbientTemperature)
					m.SetValueFloat64(instance, nat)
				}
			case "min_temperature":
				mT := num.Min(v.nonAmbientTemperature)
				m.SetValueFloat64(instance, mT)
			case "average_fan_speed":
				if len(v.fanSpeed) > 0 {
					afs := num.Avg(v.fanSpeed)
					m.SetValueFloat64(instance, afs)
				}
			case "max_fan_speed":
				mfs := num.Max(v.fanSpeed)
				m.SetValueFloat64(instance, mfs)
			case "min_fan_speed":
				mfs := num.Min(v.fanSpeed)
				m.SetValueFloat64(instance, mfs)
			}
		}
	}
}

func (d *Disk) handleCMode(shelves []*node.Node) ([]*matrix.Matrix, error) {
	var output []*matrix.Matrix
	noSet := make(map[string]struct{})

	// Purge and reset data
	for _, data1 := range d.shelfData {
		data1.PurgeInstances()
		data1.Reset()
	}

	for _, s := range shelves {

		shelfName := s.GetChildContentS("shelf")
		shelfUID := s.GetChildContentS("shelf-uid")
		shelfID := s.GetChildContentS("shelf-id")

		for attribute, data1 := range d.shelfData {
			statusMetric := data1.GetMetric("status")
			if statusMetric == nil {
				continue
			}
			if len(d.instanceKeys[attribute]) == 0 {
				d.SLogger.Warn("no instance keys defined, skipping", slog.String("attribute", attribute))
				continue
			}

			objectElem := s.GetChildS(attribute)
			if objectElem == nil {
				noSet[attribute] = struct{}{}
				continue
			}

			for _, obj := range objectElem.GetChildren() {

				if keys := d.instanceKeys[attribute]; len(keys) != 0 {
					var sKeys []string
					for _, k := range keys {
						v := obj.GetChildContentS(k)
						sKeys = append(sKeys, v)
					}
					combinedKey := strings.Join(sKeys, "")
					instanceKey := shelfUID + "#" + combinedKey
					instance, err := data1.NewInstance(instanceKey)

					if err != nil {
						d.SLogger.Error("add instance", slogx.Err(err), slog.String("attribute", attribute))
						return nil, err
					}

					for label, labelDisplay := range d.instanceLabels[attribute] {
						if value := obj.GetChildContentS(label); value != "" {
							// This is to parity with rest for modules, Convert A -> 0, B -> 1 in zapi
							if attribute == "shelf-modules" && len(value) == 1 {
								n := int(value[0] - 'A')
								value = strconv.Itoa(n)
							}
							instance.SetLabel(labelDisplay, value)
						}
					}

					instance.SetLabel("shelf", shelfName)
					instance.SetLabel("shelf_id", shelfID)

					// Each child would have different possible values which is an ugly way to write all of them,
					// so normal value would be mapped to 1 and rest all are mapped to 0.
					if instance.GetLabel("status") == "normal" {
						statusMetric.SetValueInt64(instance, 1)
					} else {
						statusMetric.SetValueInt64(instance, 0)
					}

					for metricKey, m := range data1.GetMetrics() {

						if value := strings.Split(obj.GetChildContentS(metricKey), " ")[0]; value != "" {
							if err := m.SetValueString(instance, value); err != nil {
								if value != "-" {
									d.SLogger.Debug(
										"failed to parse value",
										slog.String("metricKey", metricKey),
										slog.String("value", value),
										slogx.Err(err),
									)
								}
							}
						}
					}
				} else if d.SLogger.Enabled(context.Background(), slog.LevelDebug) {
					d.SLogger.Debug("instance without, skipping", slog.Any("skipping", d.instanceKeys[attribute]))
				}
			}

			output = append(output, data1)
		}
	}

	if len(noSet) > 0 {
		attributes := slices.Sorted(maps.Keys(noSet))
		if d.SLogger.Enabled(context.Background(), slog.LevelDebug) {
			d.SLogger.Warn("No instances", slog.Any("attributes", attributes))
		}
	}

	return output, nil
}
