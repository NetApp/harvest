package disk

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"sort"
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
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
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
	export      bool
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

func (d *Disk) Init() error {

	var err error
	shelfMetric := make(map[string][]string)

	shelfMetric["fans => fan"] = []string{
		"^^id => fan_id",
		"^location",
		"^state => status",
		"rpm",
	}
	shelfMetric["current_sensors => sensor"] = []string{
		"^^id => sensor_id",
		"^location",
		"^state => status",
		"current => reading",
	}
	shelfMetric["frus => psu"] = []string{
		"^^id => psu_id",
		"^installed => enabled",
		//"^location",
		"^part_number",
		"^serial_number => serial",
		"^psu.model => type",
		"^state => status",
		"psu.power_drawn => power_drawn",
		"psu.power_rating => power_rating",
	}
	shelfMetric["temperature_sensors => temperature"] = []string{
		"^^id => sensor_id",
		"^threshold.high.critical => high_critical",
		"^threshold.high.warning => high_warning",
		"^ambient => temp_is_ambient",
		"^threshold.low.critical => low_critical",
		"^threshold.low.warning => low_warning",
		"^state => status",
		"temperature => reading",
	}
	shelfMetric["voltage_sensors => voltage"] = []string{
		"^^id => sensor_id",
		"^location",
		"^state => status",
		"voltage => reading",
	}

	if err = d.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if d.client, err = rest.New(conf.ZapiPoller(d.ParentParams), timeout, d.Auth); err != nil {
		d.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = d.client.Init(5); err != nil {
		return err
	}

	d.query = "api/storage/shelves"

	d.shelfData = make(map[string]*matrix.Matrix)
	d.powerData = make(map[string]*matrix.Matrix)

	d.instanceKeys = make(map[string]string)
	d.instanceLabels = make(map[string]*dict.Dict)

	for attribute, childObj := range shelfMetric {

		objectName := strings.ReplaceAll(attribute, "-", "_")

		if x := strings.Split(attribute, "=>"); len(x) == 2 {
			attribute = strings.TrimSpace(x[0])
			objectName = strings.TrimSpace(x[1])
		}

		d.instanceLabels[attribute] = dict.New()

		d.shelfData[attribute] = matrix.New(d.Parent+".Shelf", "shelf_"+objectName, "shelf_"+objectName)
		d.shelfData[attribute].SetGlobalLabel("datacenter", d.ParentParams.GetChildContentS("datacenter"))

		exportOptions := node.NewS("export_options")
		instanceLabels := exportOptions.NewChildS("instance_labels", "")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")
		instanceKeys.NewChildS("", "shelf")
		instanceKeys.NewChildS("", "channel")

		// artificial metric for status of child object of shelf
		_, _ = d.shelfData[attribute].NewMetricUint8("status")

		for _, c := range childObj {

			metricName, display, kind, _ := util.ParseMetric(c)

			switch kind {
			case "key":
				d.instanceKeys[attribute] = metricName
				d.instanceLabels[attribute].Set(metricName, display)
				instanceKeys.NewChildS("", display)
				d.Logger.Debug().Msgf("added instance key: (%s) [%s]", attribute, display)
			case "label":
				d.instanceLabels[attribute].Set(metricName, display)
				instanceLabels.NewChildS("", display)
				d.Logger.Debug().Msgf("added instance label: (%s) [%s]", attribute, display)
			case "float":
				_, err := d.shelfData[attribute].NewMetricFloat64(metricName, display)
				if err != nil {
					d.Logger.Error().Stack().Err(err).Msg("add metric")
					return err
				}
				d.Logger.Debug().Msgf("added metric: (%s) [%s]", attribute, display)
			}
		}

		d.Logger.Debug().Msgf("added shelfData for [%s] with %d metrics", attribute, len(d.shelfData[attribute].GetMetrics()))

		d.shelfData[attribute].SetExportOptions(exportOptions)

		d.initShelfPowerMatrix()

		d.initAggrPowerMatrix()

		d.initMaps()
	}

	d.Logger.Debug().Msgf("initialized with shelfData [%d] objects", len(d.shelfData))
	return nil
}

func (d *Disk) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {

	data := dataMap[d.Object]
	// Set all global labels from rest.go if already not exist
	for a := range d.instanceLabels {
		d.shelfData[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	for a := range d.powerData {
		d.powerData[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	href := rest.BuildHref("", "*", nil, "", "", "", "", d.query)

	records, err := rest.Fetch(d.client, href)
	if err != nil {
		d.Logger.Error().Err(err).Str("href", href).Msg("Failed to fetch shelfData")
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	d.initMaps()

	var output []*matrix.Matrix
	noSet := make(map[string]any)

	// Purge and reset data
	for _, data1 := range d.shelfData {
		data1.PurgeInstances()
		data1.Reset()
	}
	for _, shelf := range records {

		if !shelf.IsObject() {
			d.Logger.Warn().Str("type", shelf.Type.String()).Msg("Shelf is not object, skipping")
			continue
		}

		shelfName := shelf.Get("name").String()
		shelfSerialNumber := shelf.Get("serial_number").String()

		for attribute, data1 := range d.shelfData {
			if statusMetric := data1.GetMetric("status"); statusMetric != nil {

				if d.instanceKeys[attribute] == "" {
					d.Logger.Warn().Str("attribute", attribute).Msg("no instance keys defined for object, skipping")
					continue
				}

				if childObj := shelf.Get(attribute); childObj.Exists() {
					if childObj.IsArray() {
						for _, obj := range childObj.Array() {

							// This is special condition, because child records can't be filterable in parent REST call
							// frus type can be [module, psu] and we would only need psu for our use-case.
							if attribute == "frus" && obj.Get("type").Exists() && obj.Get("type").String() != "psu" {
								continue
							}

							if key := obj.Get(d.instanceKeys[attribute]); key.Exists() {
								instanceKey := shelfSerialNumber + "#" + attribute + "#" + key.String()
								shelfChildInstance, err2 := data1.NewInstance(instanceKey)

								if err2 != nil {
									d.Logger.Error().Err(err).Str("attribute", attribute).Str("instanceKey", instanceKey).Msg("Failed to add instance")
									break
								}
								d.Logger.Debug().Msgf("add (%s) instance: %s.%s.%s", attribute, shelfSerialNumber, attribute, key)

								for label, labelDisplay := range d.instanceLabels[attribute].Map() {
									if value := obj.Get(label); value.Exists() {
										if value.IsArray() {
											var labelArray []string
											for _, r := range value.Array() {
												labelString := r.String()
												labelArray = append(labelArray, labelString)
											}
											shelfChildInstance.SetLabel(labelDisplay, strings.Join(labelArray, ","))
										} else {
											valueString := value.String()
											// For shelf child level objects' status field, Rest `ok` value maps Zapi `normal` value.
											if labelDisplay == "status" {
												valueString = strings.ReplaceAll(valueString, "ok", "normal")
											}
											shelfChildInstance.SetLabel(labelDisplay, valueString)
										}
									} else {
										// spams a lot currently due to missing label mappings. Moved to debug for now till rest gaps are filled
										d.Logger.Debug().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
									}
								}

								shelfChildInstance.SetLabel("shelf", shelfName)

								// Each child would have different possible values which is ugly way to write all of them,
								// so normal value would be mapped to 1 and rest all are mapped to 0.
								if shelfChildInstance.GetLabel("status") == "normal" {
									_ = statusMetric.SetValueInt64(shelfChildInstance, 1)
								} else {
									_ = statusMetric.SetValueInt64(shelfChildInstance, 0)
								}

								for metricKey, m := range data1.GetMetrics() {

									if value := obj.Get(metricKey); value.Exists() {
										if err = m.SetValueString(shelfChildInstance, value.String()); err != nil { // float
											d.Logger.Error().Err(err).Str("key", metricKey).Str("metric", m.GetName()).Str("value", value.String()).
												Msg("Unable to set float key on metric")
										} else {
											d.Logger.Debug().Str("metricKey", metricKey).Str("value", value.String()).Msg("added")
										}
									}
								}

							} else {
								d.Logger.Debug().Msgf("instance without [%s], skipping", d.instanceKeys[attribute])
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
		attributes := make([]string, 0)
		for k := range noSet {
			attributes = append(attributes, k)
		}
		sort.Strings(attributes)
		d.Logger.Warn().Strs("attributes", attributes).Msg("No instances")
	}

	output, err = d.handleShelfPower(records, output)
	if err != nil {
		return output, err
	}

	err = d.getAggregates()
	if err != nil {
		return output, err
	}

	err = d.getDisks()
	if err != nil {
		return output, err
	}

	err = d.populateShelfIOPS(data)
	if err != nil {
		return output, err
	}

	output, err = d.calculateAggrPower(data, output)
	if err != nil {
		return output, err
	}

	return output, nil
}

func (d *Disk) calculateAggrPower(data *matrix.Matrix, output []*matrix.Matrix) ([]*matrix.Matrix, error) {

	totalTransfers := data.GetMetric("total_transfer_count")
	if totalTransfers == nil {
		return output, errs.New(errs.ErrNoMetric, "total_transfer_count")
	}
	totaliops := make(map[string]float64)

	// calculate power for returned disks in perf response
	for _, instance := range data.GetInstances() {
		if v, ok := totalTransfers.GetValueFloat64(instance); ok {
			diskUUID := instance.GetLabel("disk_uuid")
			diskName := instance.GetLabel("disk")
			aggrName := instance.GetLabel("aggr")
			var a *aggregate
			a, ok = d.aggrMap[aggrName]
			if !ok {
				d.Logger.Warn().Str("aggrName", aggrName).Msg("Missing Aggregate info")
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
					totaliops[shelfID] = totaliops[shelfID] + v
					aggrPower := a.power + diskPower
					a.power = aggrPower
				} else {
					d.Logger.Warn().Str("shelfID", shelfID).Msg("Missing shelf info")
				}
			} else {
				d.Logger.Warn().Str("diskUUID", diskUUID).
					Str("diskName", diskName).
					Msg("Missing disk info")
			}
		} else {
			d.Logger.Warn().Msg("Instance not exported")
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
					diskWithAggregateCount += 1
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
								d.Logger.Warn().Str("aggrName", a1).Msg("Missing Aggregate info")
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
	for k, v := range d.aggrMap {
		if v.export {
			instanceKey := k
			instance, err := aggrData.NewInstance(instanceKey)
			if err != nil {
				d.Logger.Error().Err(err).Str("key", instanceKey).Msg("Failed to add instance")
				continue
			}
			instance.SetLabel("aggr", k)
			instance.SetLabel("derivedType", string(v.derivedType))
			instance.SetLabel("node", v.node)

			m := aggrData.GetMetric("power")
			err = m.SetValueFloat64(instance, v.power)
			if err != nil {
				d.Logger.Error().Err(err).Str("key", instanceKey).Msg("Failed to set value")
				continue
			}
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
		if v, ok := totalTransfers.GetValueFloat64(instance); ok {
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
				} else {
					d.Logger.Warn().Str("shelfID", shelfID).Msg("Missing shelf info")
				}
			} else {
				d.Logger.Warn().Str("diskUUID", diskUUID).
					Str("diskName", diskName).
					Msg("Missing disk info")
			}
		}
	}
	return nil
}

func (d *Disk) getDisks() error {

	var (
		err error
	)

	query := "api/storage/disks"

	href := rest.BuildHref("", "name,uid,shelf.uid,type,aggregates", nil, "", "", "", "", query)

	records, err := rest.Fetch(d.client, href)
	if err != nil {
		d.Logger.Error().Err(err).Str("href", href).Msg("Failed to fetch data")
		return err
	}

	if len(records) == 0 {
		return nil
	}

	for _, v := range records {
		if !v.IsObject() {
			d.Logger.Warn().Str("type", v.Type.String()).Msg("Shelf is not object, skipping")
			continue
		}

		diskName := v.Get("name").String()
		diskUID := v.Get("uid").String()
		shelfID := v.Get("shelf.uid").String()

		diskType := v.Get("type").String()
		aN := v.Get("aggregates.#.name")
		var aggrNames []string
		for _, name := range aN.Array() {
			aggrNames = append(aggrNames, name.String())
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

	href := rest.BuildHref("", "aggregate,composite,node,uses_shared_disks,root,storage_type", nil, "", "", "", "", query)

	records, err := rest.Fetch(d.client, href)
	if err != nil {
		d.Logger.Error().Err(err).Str("href", href).Msg("Failed to fetch data")
		return err
	}

	if len(records) == 0 {
		return nil
	}

	for _, aggr := range records {
		if !aggr.IsObject() {
			d.Logger.Warn().Str("type", aggr.Type.String()).Msg("aggregate is not object, skipping")
			continue
		}
		aggrName := aggr.Get("aggregate").String()
		usesSharedDisks := aggr.Get("uses_shared_disks").String()
		isC := aggr.Get("composite").String()
		isR := aggr.Get("root").String()
		aggregateType := aggr.Get("storage_type").String()
		nodeName := aggr.Get("node").String()
		isShared := usesSharedDisks == "true"
		isRootAggregate := isR == "true"
		isComposite := isC == "true"
		derivedType := getAggregateDerivedType(aggregateType, isComposite, isShared)
		d.aggrMap[aggrName] = &aggregate{
			name:        aggrName,
			isShared:    isShared,
			derivedType: derivedType,
			node:        nodeName,
			export:      !isRootAggregate,
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
			d.Logger.Warn().Str("type", s.Type.String()).Msg("Shelf is not object, skipping")
			continue
		}
		shelfName := s.Get("name").String()
		shelfSerialNumber := s.Get("serial_number").String()
		shelfID := s.Get("id").String()
		shelfUID := s.Get("uid").String()
		instanceKey := shelfSerialNumber
		instance, err := data.NewInstance(instanceKey)
		if err != nil {
			d.Logger.Error().Err(err).Str("key", instanceKey).Msg("Failed to add instance")
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
			d.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
		}
	}
}

func (d *Disk) initAggrPowerMatrix() {
	d.powerData["aggr"] = matrix.New(d.Parent+".Aggr", "aggr", "aggr")

	for _, k := range aggrMetrics {
		err := matrix.CreateMetric(k, d.powerData["aggr"])
		if err != nil {
			d.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
		}
	}
}

func (d *Disk) initMaps() {
	//reset shelf Power
	d.ShelfMap = make(map[string]*shelf)

	//reset diskmap
	d.diskMap = make(map[string]*disk)
	d.diskNameMap = make(map[string]*disk)

	//reset aggrmap
	d.aggrMap = make(map[string]*aggregate)
}

func (d *Disk) calculateEnvironmentMetrics(data *matrix.Matrix) {
	var err error
	shelfEnvironmentMetricMap := make(map[string]*shelfEnvironmentMetric, 0)
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
				if o.Object == "shelf_temperature" {
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
				} else if o.Object == "shelf_fan" {
					if mkey == "rpm" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							shelfEnvironmentMetricMap[iKey].fanSpeed = append(shelfEnvironmentMetricMap[iKey].fanSpeed, value)
						}
					}
				} else if o.Object == "shelf_voltage" {
					if mkey == "voltage" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							if shelfEnvironmentMetricMap[iKey].voltageSensor == nil {
								shelfEnvironmentMetricMap[iKey].voltageSensor = make(map[string]float64, 0)
							}
							shelfEnvironmentMetricMap[iKey].voltageSensor[iKey2] = value
						}
					}
				} else if o.Object == "shelf_sensor" {
					if mkey == "current" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							if shelfEnvironmentMetricMap[iKey].currentSensor == nil {
								shelfEnvironmentMetricMap[iKey].currentSensor = make(map[string]float64, 0)
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
			d.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
		}
	}
	for key, v := range shelfEnvironmentMetricMap {
		for _, k := range shelfMetrics {
			m := data.GetMetric(k)
			instance := data.GetInstance(key)
			if instance == nil {
				d.Logger.Warn().Str("key", key).Msg("Instance not found")
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
						d.Logger.Warn().Str("voltage sensor id", k1).Msg("missing current sensor")
					}
				}

				err = m.SetValueFloat64(instance, sumPower)
				if err != nil {
					d.Logger.Error().Float64("power", sumPower).Err(err).Msg("Unable to set power")
				} else {
					d.ShelfMap[instance.GetLabel("shelfUID")] = &shelf{power: sumPower}
				}

			case "average_ambient_temperature":
				if len(v.ambientTemperature) > 0 {
					aaT := util.Avg(v.ambientTemperature)
					err = m.SetValueFloat64(instance, aaT)
					if err != nil {
						d.Logger.Error().Float64("average_ambient_temperature", aaT).Err(err).Msg("Unable to set average_ambient_temperature")
					}
				}
			case "min_ambient_temperature":
				maT := util.Min(v.ambientTemperature)
				err = m.SetValueFloat64(instance, maT)
				if err != nil {
					d.Logger.Error().Float64("min_ambient_temperature", maT).Err(err).Msg("Unable to set min_ambient_temperature")
				}
			case "max_temperature":
				mT := util.Max(v.nonAmbientTemperature)
				err = m.SetValueFloat64(instance, mT)
				if err != nil {
					d.Logger.Error().Float64("max_temperature", mT).Err(err).Msg("Unable to set max_temperature")
				}
			case "average_temperature":
				if len(v.nonAmbientTemperature) > 0 {
					nat := util.Avg(v.nonAmbientTemperature)
					err = m.SetValueFloat64(instance, nat)
					if err != nil {
						d.Logger.Error().Float64("average_temperature", nat).Err(err).Msg("Unable to set average_temperature")
					}
				}
			case "min_temperature":
				mT := util.Min(v.nonAmbientTemperature)
				err = m.SetValueFloat64(instance, mT)
				if err != nil {
					d.Logger.Error().Float64("min_temperature", mT).Err(err).Msg("Unable to set min_temperature")
				}
			case "average_fan_speed":
				if len(v.fanSpeed) > 0 {
					afs := util.Avg(v.fanSpeed)
					err = m.SetValueFloat64(instance, afs)
					if err != nil {
						d.Logger.Error().Float64("average_fan_speed", afs).Err(err).Msg("Unable to set average_fan_speed")
					}
				}
			case "max_fan_speed":
				mfs := util.Max(v.fanSpeed)
				err = m.SetValueFloat64(instance, mfs)
				if err != nil {
					d.Logger.Error().Float64("max_fan_speed", mfs).Err(err).Msg("Unable to set max_fan_speed")
				}
			case "min_fan_speed":
				mfs := util.Min(v.fanSpeed)
				err = m.SetValueFloat64(instance, mfs)
				if err != nil {
					d.Logger.Error().Float64("min_fan_speed", mfs).Err(err).Msg("Unable to set min_fan_speed")
				}
			}
		}
	}
}
