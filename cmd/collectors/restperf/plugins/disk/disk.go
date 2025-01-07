package disk

import (
	"context"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"time"
)

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
	client         *rest.Client
	query          string
	aggrMap        map[string]*aggregate
	diskMap        map[string]*disk  // disk UID to disk info containing shelf name
	diskNameMap    map[string]*disk  // disk Name to disk info containing shelf name. Used for 9.12 versions where disk uuid is absent rest perf response
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

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if d.client, err = rest.New(conf.ZapiPoller(d.ParentParams), timeout, d.Auth); err != nil {
		d.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := d.client.Init(5, remote); err != nil {
		return err
	}

	d.query = "api/storage/shelves"

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

		for _, c := range obj.GetAllChildContentS() {
			metricName, display, kind, _ := util.ParseMetric(c)

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

		d.shelfData[attribute].SetExportOptions(exportOptions)
	}

	d.initShelfPowerMatrix()
	d.initAggrPowerMatrix()
	d.initMaps()

	d.SLogger.Debug("initialized", slog.Any("shelfData", len(d.shelfData)))
	return nil
}

func (d *Disk) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	data := dataMap[d.Object]
	d.client.Metadata.Reset()

	// Set all global labels from rest.go if already not exist
	for a := range d.instanceLabels {
		d.shelfData[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	for a := range d.powerData {
		d.powerData[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	href := rest.NewHrefBuilder().
		APIPath(d.query).
		Fields([]string{"*"}).
		MaxRecords(collectors.DefaultBatchSize).
		Build()

	records, err := rest.FetchAll(d.client, href)
	if err != nil {
		d.SLogger.Error("Failed to fetch shelfData", slogx.Err(err), slog.String("href", href))
		return nil, nil, err
	}

	if len(records) == 0 {
		return nil, nil, nil
	}

	d.initMaps()

	var output []*matrix.Matrix
	noSet := make(map[string]struct{})

	// Purge and reset data
	for _, data1 := range d.shelfData {
		data1.PurgeInstances()
		data1.Reset()
	}
	for _, shelf := range records {

		if !shelf.IsObject() {
			d.SLogger.Warn("Shelf is not object, skipping", slog.String("type", shelf.Type.String()))
			continue
		}

		shelfName := shelf.Get("name").ClonedString()
		shelfSerialNumber := shelf.Get("serial_number").ClonedString()

		for attribute, data1 := range d.shelfData {
			if statusMetric := data1.GetMetric("status"); statusMetric != nil {

				if len(d.instanceKeys[attribute]) == 0 {
					d.SLogger.Warn("no instance keys defined for object, skipping", slog.String("attribute", attribute))
					continue
				}

				if childObj := shelf.Get(attribute); childObj.Exists() {
					if childObj.IsArray() {
						for _, obj := range childObj.Array() {

							if keys := d.instanceKeys[attribute]; len(keys) != 0 {

								var skey []string
								for _, k := range keys {
									v := obj.Get(k)
									skey = append(skey, v.ClonedString())
								}

								combinedKey := strings.Join(skey, "")
								instanceKey := shelfSerialNumber + "#" + attribute + "#" + combinedKey
								shelfChildInstance, err2 := data1.NewInstance(instanceKey)

								if err2 != nil {
									d.SLogger.Error(
										"Failed to add instance",
										slog.Any("err", err2),
										slog.String("attribute", attribute),
										slog.String("instanceKey", instanceKey),
									)
									break
								}

								for label, labelDisplay := range d.instanceLabels[attribute] {
									if value := obj.Get(label); value.Exists() {
										if value.IsArray() {
											var labelArray []string
											for _, r := range value.Array() {
												labelString := r.ClonedString()
												labelArray = append(labelArray, labelString)
											}
											shelfChildInstance.SetLabel(labelDisplay, strings.Join(labelArray, ","))
										} else {
											valueString := value.ClonedString()
											// For shelf child level objects' status field, Rest `ok` value maps Zapi `normal` value.
											if labelDisplay == "status" {
												valueString = strings.ReplaceAll(valueString, "ok", "normal")
											}
											shelfChildInstance.SetLabel(labelDisplay, valueString)
										}
									} else {
										// spams a lot currently due to missing label mappings. Moved to debug for now till rest gaps are filled
										d.SLogger.Debug("Missing label value", slog.String("Instance key", instanceKey), slog.String("label", label))
									}
								}

								shelfChildInstance.SetLabel("shelf", shelfName)

								// Each child would have different possible values which is an ugly way to write all of them,
								// so normal value would be mapped to 1, and the rest all are mapped to 0.
								if shelfChildInstance.GetLabel("status") == "normal" {
									_ = statusMetric.SetValueInt64(shelfChildInstance, 1)
								} else {
									_ = statusMetric.SetValueInt64(shelfChildInstance, 0)
								}

								for metricKey, m := range data1.GetMetrics() {

									if value := obj.Get(metricKey); value.Exists() {
										if err = m.SetValueString(shelfChildInstance, value.ClonedString()); err != nil { // float
											d.SLogger.Error(
												"Unable to set float key on metric",
												slogx.Err(err),
												slog.String("key", metricKey),
												slog.String("metric", m.GetName()),
												slog.String("value", value.ClonedString()),
											)
										}
									}
								}

							} else if d.SLogger.Enabled(context.Background(), slog.LevelDebug) {
								d.SLogger.Debug(
									"instance without keys, skipping",
									slog.Any("attribute", d.instanceKeys[attribute]),
								)
							}
						}
					}
				} else {
					noSet[attribute] = struct{}{}
					continue
				}
				output = append(output, data1)
			}
		}
	}

	if len(noSet) > 0 {
		attributes := slices.Sorted(maps.Keys(noSet))
		if d.SLogger.Enabled(context.Background(), slog.LevelDebug) {
			d.SLogger.Warn("No instances", slog.Any("attributes", attributes))
		}
	}

	output, err = d.handleShelfPower(records, output)
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

	totalTransfers := data.GetMetric("total_transfer_count")
	if totalTransfers == nil {
		return output, errs.New(errs.ErrNoMetric, "total_transfer_count")
	}

	// calculate power for returned disks in perf response
	for key, instance := range data.GetInstances() {
		if v, ok := totalTransfers.GetValueFloat64(instance); ok {
			diskUUID := instance.GetLabel("disk_uuid")
			diskName := instance.GetLabel("disk")
			aggrName := instance.GetLabel("aggr")
			var a *aggregate
			a, ok = d.aggrMap[aggrName]
			if !ok {
				d.SLogger.Warn("Missing Aggregate info", slog.String("aggrName", aggrName))
				continue
			}

			di, has := d.diskMap[diskUUID]
			// search via diskName
			if !has {
				di, has = d.diskNameMap[diskName]
			}
			if has {
				shelfID := di.shelfID
				sh, ok := d.ShelfMap[shelfID]
				if ok {
					diskPower := v * sh.power / sh.iops
					a.power += diskPower
				}
			} else {
				d.SLogger.Warn(
					"Missing disk info",
					slog.String("diskUUID", diskUUID),
					slog.String("diskName", diskName),
				)
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
		err = m.SetValueFloat64(instance, v.power)
		if err != nil {
			d.SLogger.Error("Failed to set value", slogx.Err(err), slog.String("key", instanceKey))
			continue
		}
	}

	output = append(output, aggrData)
	return output, nil

}

func (d *Disk) populateShelfIOPS(data *matrix.Matrix) error {

	totalTransfers := data.GetMetric("total_transfer_count")
	if totalTransfers == nil {
		return errs.New(errs.ErrNoMetric, "total_transfer_count")
	}

	for _, instance := range data.GetInstances() {
		v, ok := totalTransfers.GetValueFloat64(instance)
		if !ok {
			continue
		}
		diskUUID := instance.GetLabel("disk_uuid")
		diskName := instance.GetLabel("disk")

		di, has := d.diskMap[diskUUID]
		// search via diskName
		if !has {
			di, has = d.diskNameMap[diskName]
		}
		if has {
			shelfID := di.shelfID
			sh, ok := d.ShelfMap[shelfID]
			if ok {
				sh.iops += v
			}
		} else {
			d.SLogger.Warn(
				"Missing disk info",
				slog.String("diskUUID", diskUUID),
				slog.String("diskName", diskName),
			)
		}
	}
	return nil
}

func (d *Disk) getDisks() error {

	var (
		err error
	)

	query := "api/storage/disks"

	href := rest.NewHrefBuilder().
		APIPath(query).
		MaxRecords(collectors.DefaultBatchSize).
		Fields([]string{"name", "uid", "shelf.uid", "type", "aggregates"}).
		Build()

	records, err := rest.FetchAll(d.client, href)
	if err != nil {
		d.SLogger.Error("Failed to fetch data", slogx.Err(err), slog.String("href", href))
		return err
	}

	if len(records) == 0 {
		return nil
	}

	for _, v := range records {
		if !v.IsObject() {
			d.SLogger.Warn("Shelf is not object, skipping", slog.String("type", v.Type.String()))
			continue
		}

		diskName := v.Get("name").ClonedString()
		diskUID := v.Get("uid").ClonedString()
		shelfID := v.Get("shelf.uid").ClonedString()

		diskType := v.Get("type").ClonedString()
		aN := v.Get("aggregates.#.name")
		var aggrNames []string
		for _, name := range aN.Array() {
			aggrNames = append(aggrNames, name.ClonedString())
		}

		dis := &disk{
			name:       diskName,
			shelfID:    shelfID,
			id:         diskUID,
			aggregates: aggrNames,
			diskType:   diskType,
		}
		d.diskMap[diskUID] = dis
		d.diskNameMap[diskName] = dis

		sh, ok := d.ShelfMap[shelfID]
		if ok {
			sh.disks = append(sh.disks, dis)
		}
	}
	return nil
}

func (d *Disk) getAggregates() error {

	var (
		err error
	)

	query := "api/private/cli/aggr"

	href := rest.NewHrefBuilder().
		APIPath(query).
		MaxRecords(collectors.DefaultBatchSize).
		Fields([]string{"aggregate", "composite", "node", "uses_shared_disks", "storage_type"}).
		Build()

	records, err := rest.FetchAll(d.client, href)
	if err != nil {
		d.SLogger.Error("Failed to fetch data", slogx.Err(err), slog.String("href", href))
		return err
	}

	if len(records) == 0 {
		return nil
	}

	for _, aggr := range records {
		if !aggr.IsObject() {
			d.SLogger.Warn("Aggregate is not object, skipping", slog.String("type", aggr.Type.String()))
			continue
		}
		aggrName := aggr.Get("aggregate").ClonedString()
		usesSharedDisks := aggr.Get("uses_shared_disks").ClonedString()
		isC := aggr.Get("composite").ClonedString()
		aggregateType := aggr.Get("storage_type").ClonedString()
		nodeName := aggr.Get("node").ClonedString()
		isShared := usesSharedDisks == "true"
		isComposite := isC == "true"
		derivedType := getAggregateDerivedType(aggregateType, isComposite, isShared)
		d.aggrMap[aggrName] = &aggregate{
			name:        aggrName,
			isShared:    isShared,
			derivedType: derivedType,
			node:        nodeName,
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

func (d *Disk) handleShelfPower(shelves []gjson.Result, output []*matrix.Matrix) ([]*matrix.Matrix, error) {
	// Purge and reset data
	data := d.powerData["shelf"]
	data.PurgeInstances()
	data.Reset()

	for _, s := range shelves {

		if !s.IsObject() {
			d.SLogger.Warn("Shelf is not object, skipping", slog.String("type", s.Type.String()))
			continue
		}
		shelfName := s.Get("name").ClonedString()
		shelfSerialNumber := s.Get("serial_number").ClonedString()
		shelfID := s.Get("id").ClonedString()
		shelfUID := s.Get("uid").ClonedString()
		instanceKey := shelfSerialNumber
		instance, err := data.NewInstance(instanceKey)
		if err != nil {
			d.SLogger.Error("Failed to add instance", slogx.Err(err), slog.String("key", instanceKey))
			return output, err
		}
		instance.SetLabel("shelf", shelfName)
		instance.SetLabel("shelfID", shelfID)
		instance.SetLabel("shelfUID", shelfUID)
	}
	d.calculateEnvironmentMetrics(data)

	output = append(output, data)
	return output, nil
}

func (d *Disk) initShelfPowerMatrix() {
	d.powerData = make(map[string]*matrix.Matrix)
	d.powerData["shelf"] = matrix.New(d.Parent+".Shelf", "shelf", "shelf")

	for _, k := range shelfMetrics {
		err := matrix.CreateMetric(k, d.powerData["shelf"])
		if err != nil {
			d.SLogger.Warn("error while creating metric", slogx.Err(err), slog.String("key", k))
		}
	}
}

func (d *Disk) initAggrPowerMatrix() {
	d.powerData["aggr"] = matrix.New(d.Parent+".Aggr", "aggr", "aggr")

	for _, k := range aggrMetrics {
		err := matrix.CreateMetric(k, d.powerData["aggr"])
		if err != nil {
			d.SLogger.Warn("error while creating metric", slogx.Err(err), slog.String("key", k))
		}
	}
}

func (d *Disk) initMaps() {
	// reset shelf Power
	d.ShelfMap = make(map[string]*shelf)

	// reset diskmap
	d.diskMap = make(map[string]*disk)
	d.diskNameMap = make(map[string]*disk)

	// reset aggrmap
	d.aggrMap = make(map[string]*aggregate)
}

func (d *Disk) calculateEnvironmentMetrics(data *matrix.Matrix) {
	var err error
	shelfEnvironmentMetricMap := make(map[string]*shelfEnvironmentMetric)
	for _, o := range d.shelfData {
		for k, instance := range o.GetInstances() {
			firstInd := strings.Index(k, "#")
			lastInd := strings.LastIndex(k, "#")
			iKey := k[:firstInd]
			iKey2 := k[lastInd+1:]
			if _, ok := shelfEnvironmentMetricMap[iKey]; !ok {
				shelfEnvironmentMetricMap[iKey] = &shelfEnvironmentMetric{key: iKey, ambientTemperature: []float64{}, nonAmbientTemperature: []float64{}, fanSpeed: []float64{}}
			}
			for mkey, metric := range o.GetMetrics() {
				switch {
				case o.Object == "shelf_temperature":
					if mkey == "temperature" {
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
				case o.Object == "shelf_fan":
					if mkey == "rpm" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							shelfEnvironmentMetricMap[iKey].fanSpeed = append(shelfEnvironmentMetricMap[iKey].fanSpeed, value)
						}
					}
				case o.Object == "shelf_voltage":
					if mkey == "voltage" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							if shelfEnvironmentMetricMap[iKey].voltageSensor == nil {
								shelfEnvironmentMetricMap[iKey].voltageSensor = make(map[string]float64)
							}
							shelfEnvironmentMetricMap[iKey].voltageSensor[iKey2] = value
						}
					}
				case o.Object == "shelf_sensor":
					if mkey == "current" {
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

	for _, k := range shelfMetrics {
		err := matrix.CreateMetric(k, data)
		if err != nil {
			d.SLogger.Warn("error while creating metric", slogx.Err(err), slog.String("key", k))
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

				err = m.SetValueFloat64(instance, sumPower)
				if err != nil {
					d.SLogger.Error("Unable to set power", slogx.Err(err), slog.Float64("power", sumPower))
				} else {
					d.ShelfMap[instance.GetLabel("shelfUID")] = &shelf{power: sumPower}
				}

			case "average_ambient_temperature":
				if len(v.ambientTemperature) > 0 {
					aaT := util.Avg(v.ambientTemperature)
					err = m.SetValueFloat64(instance, aaT)
					if err != nil {
						d.SLogger.Error(
							"Unable to set average_ambient_temperature",
							slogx.Err(err),
							slog.Float64("average_ambient_temperature", aaT),
						)
					}
				}
			case "min_ambient_temperature":
				maT := util.Min(v.ambientTemperature)
				err = m.SetValueFloat64(instance, maT)
				if err != nil {
					d.SLogger.Error(
						"Unable to set min_ambient_temperature",
						slogx.Err(err),
						slog.Float64("min_ambient_temperature", maT),
					)
				}
			case "max_temperature":
				mT := util.Max(v.nonAmbientTemperature)
				err = m.SetValueFloat64(instance, mT)
				if err != nil {
					d.SLogger.Error(
						"Unable to set max_temperature",
						slogx.Err(err),
						slog.Float64("max_temperature", mT),
					)
				}
			case "average_temperature":
				if len(v.nonAmbientTemperature) > 0 {
					nat := util.Avg(v.nonAmbientTemperature)
					err = m.SetValueFloat64(instance, nat)
					if err != nil {
						d.SLogger.Error(
							"Unable to set average_temperature",
							slogx.Err(err),
							slog.Float64("average_temperature", nat),
						)
					}
				}
			case "min_temperature":
				mT := util.Min(v.nonAmbientTemperature)
				err = m.SetValueFloat64(instance, mT)
				if err != nil {
					d.SLogger.Error(
						"Unable to set min_temperature",
						slogx.Err(err),
						slog.Float64("min_temperature", mT),
					)
				}
			case "average_fan_speed":
				if len(v.fanSpeed) > 0 {
					afs := util.Avg(v.fanSpeed)
					err = m.SetValueFloat64(instance, afs)
					if err != nil {
						d.SLogger.Error(
							"Unable to set average_fan_speed",
							slogx.Err(err),
							slog.Float64("average_fan_speed", afs),
						)
					}
				}
			case "max_fan_speed":
				mfs := util.Max(v.fanSpeed)
				err = m.SetValueFloat64(instance, mfs)
				if err != nil {
					d.SLogger.Error(
						"Unable to set max_fan_speed",
						slogx.Err(err),
						slog.Float64("max_fan_speed", mfs),
					)
				}
			case "min_fan_speed":
				mfs := util.Min(v.fanSpeed)
				err = m.SetValueFloat64(instance, mfs)
				if err != nil {
					d.SLogger.Error(
						"Unable to set min_fan_speed",
						slogx.Err(err),
						slog.Float64("min_fan_speed", mfs),
					)
				}
			}
		}
	}
}
