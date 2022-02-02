package shelf

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"strings"
	"time"
)

type Shelf struct {
	*plugin.AbstractPlugin
	data           map[string]*matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
	client         *rest.Client
	query          string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Shelf{AbstractPlugin: p}
}

func (my *Shelf) Init() error {

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

	if err = my.InitAbc(); err != nil {
		return err
	}

	timeout := rest.DefaultTimeout * time.Second
	if my.client, err = rest.New(conf.ZapiPoller(my.ParentParams), timeout); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.query = "api/storage/shelves"

	my.data = make(map[string]*matrix.Matrix)
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]*dict.Dict)

	for attribute, childObj := range shelfMetric {

		objectName := strings.ReplaceAll(attribute, "-", "_")

		if x := strings.Split(attribute, "=>"); len(x) == 2 {
			attribute = strings.TrimSpace(x[0])
			objectName = strings.TrimSpace(x[1])
		}

		my.instanceLabels[attribute] = dict.New()

		my.data[attribute] = matrix.New(my.Parent+".Shelf", "shelf_"+objectName, "shelf_"+objectName)
		my.data[attribute].SetGlobalLabel("datacenter", my.ParentParams.GetChildContentS("datacenter"))

		exportOptions := node.NewS("export_options")
		instanceLabels := exportOptions.NewChildS("instance_labels", "")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")
		instanceKeys.NewChildS("", "shelf")
		instanceKeys.NewChildS("", "channel")

		// artificial metric for status of child object of shelf
		my.data[attribute].NewMetricUint8("status")

		for _, c := range childObj {

			metricName, display, kind, _ := util.ParseMetric(c)

			switch kind {
			case "key":
				my.instanceKeys[attribute] = metricName
				my.instanceLabels[attribute].Set(metricName, display)
				instanceKeys.NewChildS("", display)
				my.Logger.Debug().Msgf("added instance key: (%s) [%s]", attribute, display)
			case "label":
				my.instanceLabels[attribute].Set(metricName, display)
				instanceLabels.NewChildS("", display)
				my.Logger.Debug().Msgf("added instance label: (%s) [%s]", attribute, display)
			case "float":
				metric, err := my.data[attribute].NewMetricFloat64(metricName)
				if err != nil {
					my.Logger.Error().Stack().Err(err).Msg("add metric")
					return err
				}
				metric.SetName(display)
				my.Logger.Debug().Msgf("added metric: (%s) [%s]", attribute, display)
			}
		}

		my.Logger.Debug().Msgf("added data for [%s] with %d metrics", attribute, len(my.data[attribute].GetMetrics()))

		my.data[attribute].SetExportOptions(exportOptions)
	}

	my.Logger.Debug().Msgf("initialized with data [%d] objects", len(my.data))
	return nil
}

func (my *Shelf) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		records []interface{}
		content []byte
		err     error
	)

	// Set all global labels from rest.go if already not exist
	for a := range my.instanceLabels {
		my.data[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	href := rest.BuildHref("", "*", nil, "", "", "", "", my.query)

	err = rest.FetchData(my.client, href, &records)
	if err != nil {
		my.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
		return nil, err
	}

	all := rest.Pagination{
		Records:    records,
		NumRecords: len(records),
	}

	content, err = json.Marshal(all)
	if err != nil {
		my.Logger.Error().Err(err).Str("ApiPath", my.query).Msg("Unable to marshal rest pagination")
	}

	if !gjson.ValidBytes(content) {
		return nil, fmt.Errorf("json is not valid for: %s", my.query)
	}

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no "+my.query+" instances on cluster")
	}

	my.Logger.Debug().Msgf("fetching %d shelf counters", numRecords.Int())

	var output []*matrix.Matrix

	// Purge and reset data
	for _, data1 := range my.data {
		data1.PurgeInstances()
		data1.Reset()
	}

	results[1].ForEach(func(shelfKey, shelf gjson.Result) bool {

		if !shelf.IsObject() {
			my.Logger.Warn().Str("type", shelf.Type.String()).Msg("Shelf is not object, skipping")
			return true
		}

		shelfName := shelf.Get("name").String()
		shelfId := shelf.Get("uid").String()

		for attribute, data1 := range my.data {
			if statusMetric := data1.GetMetric("status"); statusMetric != nil {

				if my.instanceKeys[attribute] == "" {
					my.Logger.Warn().Str("attribute", attribute).Msg("no instance keys defined for object, skipping")
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

							if key := obj.Get(my.instanceKeys[attribute]); key.Exists() {
								instanceKey := shelfId + "." + attribute + "." + key.String()
								shelfChildInstance, errs := data1.NewInstance(instanceKey)

								if errs != nil {
									my.Logger.Error().Err(err).Str("attribute", attribute).Str("instanceKey", instanceKey).Msg("Failed to add instance")
									break
								}
								my.Logger.Debug().Msgf("add (%s) instance: %s.%s.%s", attribute, shelfId, attribute, key)

								for label, labelDisplay := range my.instanceLabels[attribute].Map() {
									if value := obj.Get(label); value.Exists() {
										if value.IsArray() {
											var labelArray []string
											for _, r := range value.Array() {
												labelString := r.String()
												labelArray = append(labelArray, labelString)
											}
											shelfChildInstance.SetLabel(labelDisplay, strings.Join(labelArray, ","))
										} else {
											shelfChildInstance.SetLabel(labelDisplay, value.String())
										}
									} else {
										// spams a lot currently due to missing label mappings. Moved to debug for now till rest gaps are filled
										my.Logger.Debug().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
									}
								}

								shelfChildInstance.SetLabel("shelf", shelfName)
								shelfChildInstance.SetLabel("shelf_id", shelfId)

								// Each child would have different possible values which is ugly way to write all of them,
								// so normal value would be mapped to 1 and rest all are mapped to 0.
								if shelfChildInstance.GetLabel("status") == "ok" {
									statusMetric.SetValueInt(shelfChildInstance, 1)
								} else {
									statusMetric.SetValueInt(shelfChildInstance, 0)
								}

								for metricKey, m := range data1.GetMetrics() {

									if value := obj.Get(metricKey); value.Exists() {
										if err = m.SetValueString(shelfChildInstance, value.String()); err != nil { // float
											my.Logger.Error().Err(err).Str("key", metricKey).Str("metric", m.GetName()).Str("value", value.String()).
												Msg("Unable to set float key on metric")
										} else {
											my.Logger.Debug().Str("metricKey", metricKey).Str("value", value.String()).Msg("added")
										}
									}
								}

							} else {
								my.Logger.Debug().Msgf("instance without [%s], skipping", my.instanceKeys[attribute])
							}
						}
					}
				} else {
					my.Logger.Warn().Msgf("no [%s] instances on this system", attribute)
					continue
				}
				output = append(output, data1)
			}
		}
		return true
	})

	return output, nil
}
