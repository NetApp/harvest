package security

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
	"strconv"
	"strings"
	"time"
)

type Security struct {
	*plugin.AbstractPlugin
	data           map[string]*matrix.Matrix
	instanceKeys   map[string][]string
	instanceLabels map[string]*dict.Dict
	client         *rest.Client
	query          map[string]string
	plugins        map[string]string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Security{AbstractPlugin: p}
}

func (my *Security) Init() error {

	var err error

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

	my.data = make(map[string]*matrix.Matrix)
	my.instanceKeys = make(map[string][]string)
	my.instanceLabels = make(map[string]*dict.Dict)
	my.query = make(map[string]string)
	my.plugins = make(map[string]string)

	objects := my.Params.GetChildS("objects")
	if objects == nil {
		return errors.New(errors.MISSING_PARAM, "objects")
	}

	for _, obj := range objects.GetChildren() {
		my.Logger.Info().Msgf("obj: (%s) ", obj.GetNameS())
		var query, securityObj string

		if x := strings.Split(obj.GetNameS(), "=>"); len(x) == 2 {
			query = strings.TrimSpace(x[0])
			securityObj = strings.TrimSpace(x[1])
			my.query[securityObj] = query
			my.Logger.Info().Msgf("securityObj: (%s), query: %s ", securityObj, query)
		}

		my.instanceLabels[securityObj] = dict.New()

		my.data[securityObj] = matrix.New(my.Parent+".Security", securityObj, securityObj)
		my.data[securityObj].SetGlobalLabel("datacenter", my.ParentParams.GetChildContentS("datacenter"))

		exportOptions := node.NewS("export_options")
		instanceLabels := exportOptions.NewChildS("instance_labels", "")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")
		//instanceKeys.NewChildS("", "shelf")
		//instanceKeys.NewChildS("", "channel")

		// artificial metric for status of child object of shelf
		my.data[securityObj].NewMetricUint8("status")

		innerPlugin := obj.GetChildS("InnerPlugin")
		if innerPlugin != nil {
			my.Logger.Info().Msgf("innerPlugin [%s]", innerPlugin.GetContentS())
			obj.PopChildS(innerPlugin.GetNameS())
			my.plugins[securityObj] = innerPlugin.GetContentS()
		}

		for _, c := range obj.GetAllChildContentS() {
			my.Logger.Info().Msgf("c: (%s)", c)

			metricName, display, kind, _ := util.ParseMetric(c)
			my.Logger.Info().Msgf("metricName: (%s), display: %s , kind=%s", metricName, display, kind)

			switch kind {
			case "key":
				my.instanceKeys[securityObj] = append(my.instanceKeys[securityObj], metricName)
				my.instanceLabels[securityObj].Set(metricName, display)
				instanceKeys.NewChildS("", display)
				my.Logger.Info().Msgf("added instance key: (%s) [%s]", securityObj, display)
			case "label":
				my.instanceLabels[securityObj].Set(metricName, display)
				instanceLabels.NewChildS("", display)
				my.Logger.Info().Msgf("added instance label: (%s) [%s]", securityObj, display)
			case "float":
				metric, err := my.data[securityObj].NewMetricFloat64(metricName)
				if err != nil {
					my.Logger.Error().Stack().Err(err).Msg("add metric")
					return err
				}
				metric.SetName(display)
				my.Logger.Info().Msgf("added metric: (%s) [%s]", securityObj, display)
			}
		}

		//if my.instanceKeys[securityObj] == nil {
		//	my.data[securityObj].GetExportOptions().GetChildS("instance_keys").NewChildS("", "cluster")
		//}

		my.Logger.Info().Msgf("added data for [%s] with %d metrics", securityObj, len(my.data[securityObj].GetMetrics()))

		my.data[securityObj].SetExportOptions(exportOptions)
	}

	my.Logger.Info().Msgf("initialized with data [%d] objects", len(my.data))
	return nil
}

func (my *Security) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		resultData []gjson.Result
		output     []*matrix.Matrix
		err        error
	)

	for securityObj, data1 := range my.data {
		var securityObjInstance *matrix.Instance

		my.Logger.Info().Msgf("securityObj [%s]", securityObj)

		if resultData, err = my.getSecurityObjData(securityObj); err != nil {
			my.Logger.Err(err).Str("securityObj", securityObj).Msg("error while fetching data for object")
			continue
		}
		my.Logger.Info().Msgf("data length [%d]", len(resultData))

		// Purge and reset data
		data1.PurgeInstances()
		data1.Reset()

		// Set all global labels from rest.go if already not exist
		my.data[securityObj].SetGlobalLabels(data.GetGlobalLabels())

		for _, secObj := range resultData {
			instanceKey := ""

			if !secObj.IsObject() {
				my.Logger.Warn().Str("type", secObj.Type.String()).Msg("secObj is not object, skipping")
				return nil, nil
			}

			if my.instanceKeys[securityObj] == nil {
				my.Logger.Warn().Str("securityObj", securityObj).Msg("no instance keys defined for object")
				if data1.GetInstance("cluster") == nil {
					if securityObjInstance, err = data1.NewInstance("cluster"); err != nil {
						my.Logger.Error().Err(err).Str("securityObj", securityObj).Str("instanceKey", instanceKey).Msg("Failed to add instance")
						break
					}
					my.Logger.Info().Msgf("key [%s]", instanceKey)
				}
			} else {
				// extract instance key(s)
				for _, k := range my.instanceKeys[securityObj] {
					value := secObj.Get(k)
					if value.Exists() {
						instanceKey += value.String()
					} else {
						my.Logger.Warn().Str("key", k).Msg("skip instance, missing key")
						break
					}
				}
				if securityObjInstance, err = data1.NewInstance(instanceKey); err != nil {
					my.Logger.Error().Err(err).Str("securityObj", securityObj).Str("instanceKey", instanceKey).Msg("Failed to add instance")
					break
				}
				my.Logger.Info().Msgf("add (%s) instancekey: %s", securityObj, instanceKey)
			}

			for label, labelDisplay := range my.instanceLabels[securityObj].Map() {
				if value := secObj.Get(label); value.Exists() {
					if value.IsArray() {
						var labelArray []string
						for _, r := range value.Array() {
							labelString := r.String()
							labelArray = append(labelArray, labelString)
						}
						my.Logger.Info().Msgf("labelDisplay [%s] labelArray [%s]", labelDisplay, labelArray)
						securityObjInstance.SetLabel(labelDisplay, strings.Join(labelArray, ","))
					} else {
						my.Logger.Info().Msgf("labelDisplay [%s] [%s]", labelDisplay, value.String())
						securityObjInstance.SetLabel(labelDisplay, value.String())
					}
				} else {
					// spams a lot currently due to missing label mappings. Moved to debug for now till rest gaps are filled
					my.Logger.Info().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
				}
			}

			for metricKey, m := range data1.GetMetrics() {

				if value := secObj.Get(metricKey); value.Exists() {
					if err = m.SetValueString(securityObjInstance, value.String()); err != nil { // float
						my.Logger.Error().Err(err).Str("key", metricKey).Str("metric", m.GetName()).Str("value", value.String()).
							Msg("Unable to set float key on metric")
					} else {
						my.Logger.Info().Str("metricKey", metricKey).Str("value", value.String()).Msg("added")
					}
				}
			}
		}
		my.InnerPluginUpdate(securityObj, data1)
		output = append(output, data1)
	}

	return output, nil
}

func (my *Security) getSecurityObjData(securityObj string) ([]gjson.Result, error) {
	var (
		records []interface{}
		content []byte
		err     error
	)
	my.Logger.Info().Msgf("securityObj %s query: %s", securityObj, my.query[securityObj])

	href := rest.BuildHref("", "*", nil, "", "", "", "", my.query[securityObj])

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
		my.Logger.Error().Err(err).Str("ApiPath", my.query[securityObj]).Msg("Unable to marshal rest pagination")
	}

	if !gjson.ValidBytes(content) {
		return nil, fmt.Errorf("json is not valid for: %s", my.query[securityObj])
	}

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no "+my.query[securityObj]+" instances on cluster")
	}

	my.Logger.Info().Msgf("fetching %d shelf counters", numRecords.Int())
	return results[1].Array(), nil
}

func (my *Security) InnerPluginUpdate(securityObj string, data1 *matrix.Matrix) {
	if innerPlugin, ok := my.plugins[securityObj]; ok {
		switch innerPlugin {
		case "security_account":
			my.setSecurityAccountLoginMethod(securityObj, data1)
		default:
			my.Logger.Warn().Str("innerPlugin", innerPlugin).Msg("no rest innerPlugin plugin found ")
		}
	}
}

func (my *Security) setSecurityAccountLoginMethod(securityObj string, data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		methods := instance.GetLabel("methods")
		my.Logger.Info().Msgf("%s", methods)
		instance.SetLabel("samluser", strconv.FormatBool(strings.Contains(methods, "saml")))
		instance.SetLabel("ldapuser", strconv.FormatBool(strings.Contains(methods, "nsswitch")))
		instance.SetLabel("certificateuser", strconv.FormatBool(strings.Contains(methods, "certificate")))
		instance.SetLabel("localuser", strconv.FormatBool(strings.Contains(methods, "password")))
		instance.SetLabel("activediruser", strconv.FormatBool(strings.Contains(methods, "domain")))
	}
	my.data[securityObj].GetExportOptions().GetChildS("instance_labels").NewChildS("", "samluser")
	my.data[securityObj].GetExportOptions().GetChildS("instance_labels").NewChildS("", "ldapuser")
	my.data[securityObj].GetExportOptions().GetChildS("instance_labels").NewChildS("", "certificateuser")
	my.data[securityObj].GetExportOptions().GetChildS("instance_labels").NewChildS("", "localuser")
	my.data[securityObj].GetExportOptions().GetChildS("instance_labels").NewChildS("", "activediruser")
}
