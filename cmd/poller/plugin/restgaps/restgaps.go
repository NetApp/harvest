/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package restgaps

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"strings"
	"time"
)

const (
	// DefaultTimeout should be > than ONTAP's default REST timeout, which is 15 seconds for GET requests
	DefaultTimeout = 30
)

type RestGaps struct {
	*plugin.AbstractPlugin
	rules []*rule
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &RestGaps{AbstractPlugin: p}
}

type rule struct {
	props *prop
	name  string
}

type prop struct {
	query          string
	instanceKeys   []string
	instanceLabels map[string]string
	metrics        []string
	counters       map[string]string
}

func (me *RestGaps) Init() error {

	if err := me.AbstractPlugin.Init(); err != nil {
		return err
	}

	me.rules = make([]*rule, 0)
	if err := me.parseRules(); err != nil {
		return err
	}

	return nil
}

func (me *RestGaps) parseRules() error {

	for _, line := range me.Params.GetChildren() {
		var (
			counters            *node.Node
			display, name, kind string
		)

		n := line.GetNameS()
		r := rule{name: n}

		prop := prop{}

		prop.instanceKeys = make([]string, 0)
		prop.instanceLabels = make(map[string]string)
		prop.counters = make(map[string]string)
		prop.metrics = make([]string, 0)

		for _, line1 := range line.GetChildren() {
			if line1.GetNameS() == "query" {
				prop.query = line1.GetContentS()
			}
			if line1.GetNameS() == "counters" {
				counters = line.GetChildS("counters")

				for _, c := range counters.GetAllChildContentS() {
					name, display, kind = util.ParseMetric(c)

					prop.counters[name] = display
					switch kind {
					case "key":
						prop.instanceLabels[name] = display
						prop.instanceKeys = append(prop.instanceKeys, name)
					case "label":
						prop.instanceLabels[name] = display
					default:
						prop.instanceLabels[name] = display
						prop.metrics = append(prop.metrics, name)
					}
				}
			}
		}
		r.props = &prop
		me.rules = append(me.rules, &r)
	}
	return nil
}

func (me *RestGaps) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		err error
	)
	p := conf.ZapiPoller(me.ParentParams)
	// using template client_timeout. do we need client_timeout at plugin level?
	clientTimeout := p.ClientTimeout
	timeout := DefaultTimeout * time.Second
	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		timeout = duration
		me.Logger.Debug().Str("timeout", timeout.String()).Msg("Using timeout")
	} else {
		me.Logger.Debug().Str("timeout", timeout.String()).Msg("Using default timeout")
	}
	client, _ := rest.New(&p, timeout)

	for _, rule := range me.rules {
		var (
			records []interface{}
			content []byte
		)
		counterKey := make([]string, len(rule.props.counters))
		i := 0
		for k := range rule.props.counters {
			counterKey[i] = k
			i++
		}
		href := rest.BuildHref(rule.props.query, strings.Join(counterKey, ","), nil, "", "", "", "", rule.props.query)
		me.Logger.Debug().Str("href", href).Msg("")
		err = rest.FetchData(client, href, &records)
		if err != nil {
			me.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
			return nil, err
		}

		all := rest.Pagination{
			Records:    records,
			NumRecords: len(records),
		}

		content, err = json.Marshal(all)
		if err != nil {
			me.Logger.Error().Err(err).Str("ApiPath", rule.props.query).Msg("Unable to marshal rest pagination")
		}

		if !gjson.ValidBytes(content) {
			return nil, fmt.Errorf("json is not valid for: %s", rule.props.query)
		}

		results := gjson.GetManyBytes(content, "num_records", "records")
		numRecords := results[0]
		if numRecords.Int() == 0 {
			return nil, errors.New(errors.ERR_NO_INSTANCE, "no "+rule.props.query+" instances on cluster")
		}

		results[1].ForEach(func(key, instanceData gjson.Result) bool {
			var (
				instanceKey string
				instance    *matrix.Instance
			)

			if !instanceData.IsObject() {
				me.Logger.Warn().Str("type", instanceData.Type.String()).Msg("Instance data is not object, skipping")
				return true
			}

			// extract instance key(s)
			for _, k := range rule.props.instanceKeys {
				value := instanceData.Get(k)
				if value.Exists() {
					instanceKey += value.String()
				} else {
					me.Logger.Warn().Str("key", k).Msg("skip instance, missing key")
					break
				}
			}

			if instance = data.GetInstance(instanceKey); instance != nil {

				for label, display := range rule.props.instanceLabels {
					value := instanceData.Get(label)
					if value.Exists() {
						if value.IsArray() {
							var labelArray []string
							for _, r := range value.Array() {
								labelString := r.String()
								labelArray = append(labelArray, labelString)
							}
							instance.SetLabel(display, strings.Join(labelArray, ","))
						} else {
							instance.SetLabel(display, value.String())
						}
					} else {
						// spams a lot currently due to missing label mappings. Moved to debug for now till rest gaps are filled
						me.Logger.Debug().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
					}
				}

				for _, metric := range rule.props.metrics {
					f := instanceData.Get(metric)
					if f.Exists() {
						if metr, ok := data.GetMetrics()[metric]; !ok {
							if metr, _ = data.NewMetricFloat64(metric); err != nil {
								me.Logger.Error().Err(err).
									Str("name", metric).
									Msg("NewMetricFloat64")
							}
							metr.SetName(rule.props.instanceLabels[metric])
							if err = metr.SetValueFloat64(instance, f.Float()); err != nil {
								me.Logger.Error().Err(err).Str("key", metric).Str("metric", rule.props.instanceLabels[metric]).
									Msg("Unable to set float key on metric")
							}
						} else {
							metr.SetName(rule.props.instanceLabels[metric])
							if err = metr.SetValueFloat64(instance, f.Float()); err != nil {
								me.Logger.Error().Err(err).Str("key", metric).Str("metric", rule.props.instanceLabels[metric]).
									Msg("Unable to set float key on metric")
							}
						}
					}
				}
			}
			return true
		})

	}

	return nil, nil
}
