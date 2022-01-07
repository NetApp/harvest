package rest

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/collectors/rest/plugins/disk"
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"os"
	"strings"
	"time"
)

type Rest struct {
	*collector.AbstractCollector
	client    *rest.Client
	prop      *prop
	endpoints []*endPoint
}

type endPoint struct {
	prop *prop
	name string
}

type prop struct {
	query          string
	instanceKeys   []string
	instanceLabels map[string]string
	metrics        []string
	counters       map[string]string
	returnTimeOut  string
	fields         []string
	apiType        string // public, private
}

func init() {
	plugin.RegisterModule(Rest{})
}

func (Rest) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.rest",
		New: func() plugin.Module { return new(Rest) },
	}
}

func (r *Rest) Init(a *collector.AbstractCollector) error {

	var err error

	r.AbstractCollector = a

	r.prop = &prop{}

	if r.client, err = r.getClient(a, a.Params); err != nil {
		return err
	}

	if err = r.client.Init(5); err != nil {
		return err
	}

	if err = r.InitVars(); err != nil {
		return err
	}

	if err = r.initEndPoints(); err != nil {
		return err
	}

	if err = collector.Init(r); err != nil {
		return err
	}

	r.Logger.Info().Str("cluster", r.client.Cluster().Name).Msgf("connected to %s", r.client.Cluster().Info)

	r.Matrix.SetGlobalLabel("cluster", r.client.Cluster().Name)

	if err = r.initCache(); err != nil {
		return err
	}
	r.Logger.Info().Msgf("initialized cache with %d metrics", len(r.Matrix.GetMetrics()))

	return nil
}

func (r *Rest) InitVars() error {

	var (
		template *node.Node
		err      error
	)

	// import template
	if template, err = r.ImportSubTemplate("", r.getTemplateFn(), r.client.Cluster().Version); err != nil {
		return err
	}

	r.Params.Union(template)
	// private end point do not support * as fields. We need to pass fields in endpoint
	query := r.Params.GetChildS("query")
	if query != nil && strings.Contains(query.GetContentS(), "private") {
		r.prop.apiType = "private"
	} else {
		r.prop.apiType = "public"
	}

	r.prop.fields = []string{"*"}
	if c := r.Params.GetChildS("counters"); c != nil {
		if x := c.GetChildS("hidden_fields"); x != nil {
			r.prop.fields = append(r.prop.fields, x.GetAllChildContentS()...)
		}
	}
	return nil
}

func (r *Rest) getClient(a *collector.AbstractCollector, config *node.Node) (*rest.Client, error) {
	var (
		poller *conf.Poller
		err    error
		client *rest.Client
	)

	opt := a.GetOptions()
	if poller, err = conf.PollerNamed(opt.Poller); err != nil {
		r.Logger.Error().Stack().Err(err).Str("poller", opt.Poller).Msgf("")
		return nil, err
	}
	if poller.Addr == "" {
		r.Logger.Error().Stack().Str("poller", opt.Poller).Msg("Address is empty")
		return nil, errors.New(errors.MISSING_PARAM, "addr")
	}

	clientTimeout := config.GetChildContentS("client_timeout")
	timeout := rest.DefaultTimeout * time.Second
	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		timeout = duration
		r.Logger.Info().Str("timeout", timeout.String()).Msg("Using timeout")
	} else {
		r.Logger.Info().Str("timeout", timeout.String()).Msg("Using default timeout")
	}
	if client, err = rest.New(poller, timeout); err != nil {
		r.Logger.Error().Stack().Err(err).Str("poller", opt.Poller).Msg("error creating new client")
		os.Exit(1)
	}

	return client, err
}

func (r *Rest) initEndPoints() error {

	endpoints := r.Params.GetChildS("endpoints")
	if endpoints != nil {
		for _, line := range endpoints.GetChildren() {
			var (
				display, name, kind string
			)

			n := line.GetNameS()
			e := endPoint{name: n}

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

					for _, c := range line1.GetAllChildContentS() {
						name, display, kind = ParseMetric(c)

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
			e.prop = &prop
			r.endpoints = append(r.endpoints, &e)
		}
	}
	return nil
}

func (r *Rest) getTemplateFn() string {
	var fn string
	objects := r.Params.GetChildS("objects")
	if objects != nil {
		fn = objects.GetChildContentS(r.Object)
	}
	return fn
}

// Returns a slice of keys in dot notation from json
func getFieldName(source string, parent string) []string {
	res := make([]string, 0)
	var arr map[string]gjson.Result
	r := gjson.Parse(source)
	if r.IsArray() {
		newR := r.Get("0")
		arr = newR.Map()
	} else if r.IsObject() {
		arr = r.Map()
	} else {
		return []string{parent}
	}
	if len(arr) == 0 {
		return []string{parent}
	}
	for key, val := range arr {
		var temp []string
		if parent == "" {
			temp = getFieldName(val.Raw, key)
		} else {
			temp = getFieldName(val.Raw, parent+"."+key)
		}
		res = append(res, temp...)
	}
	return res
}

func (r *Rest) PollData() (*matrix.Matrix, error) {

	var (
		content      []byte
		count        uint64
		apiD, parseD time.Duration
		startTime    time.Time
		err          error
		records      []interface{}
	)

	r.Logger.Debug().Msg("starting data poll")
	r.Matrix.Reset()

	startTime = time.Now()

	if r.prop.apiType == "public" {
		href := rest.BuildHref(r.prop.query, strings.Join(r.prop.fields, ","), nil, "", "", "", r.prop.returnTimeOut, r.prop.query)
		r.Logger.Debug().Str("href", href).Msg("")
		err = rest.FetchData(r.client, href, &records)
		if err != nil {
			r.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
			return nil, err
		}
	}

	if r.prop.apiType == "private" {
		counterKey := make([]string, len(r.prop.counters))
		i := 0
		for k := range r.prop.counters {
			counterKey[i] = k
			i++
		}
		href := rest.BuildHref(r.prop.query, strings.Join(counterKey, ","), nil, "", "", "", r.prop.returnTimeOut, r.prop.query)
		r.Logger.Debug().Str("href", href).Msg("")
		err = rest.FetchData(r.client, href, &records)
		if err != nil {
			r.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
			return nil, err
		}
	}

	all := rest.Pagination{
		Records:    records,
		NumRecords: len(records),
	}
	apiD = time.Since(startTime)

	content, err = json.Marshal(all)
	if err != nil {
		r.Logger.Error().Err(err).Str("ApiPath", r.prop.query).Msg("Unable to marshal rest pagination")
	}

	startTime = time.Now()
	if !gjson.ValidBytes(content) {
		return nil, fmt.Errorf("json is not valid for: %s", r.prop.query)
	}
	parseD = time.Since(startTime)

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no "+r.Object+" instances on cluster")
	}

	r.Logger.Debug().Str("object", r.Object).Str("number of records extracted", numRecords.String()).Msg("")

	results[1].ForEach(func(key, instanceData gjson.Result) bool {
		var (
			instanceKey string
			instance    *matrix.Instance
		)

		if !instanceData.IsObject() {
			r.Logger.Warn().Str("type", instanceData.Type.String()).Msg("Instance data is not object, skipping")
			return true
		}

		// extract instance key(s)
		for _, k := range r.prop.instanceKeys {
			value := instanceData.Get(k)
			if value.Exists() {
				instanceKey += value.String()
			} else {
				r.Logger.Warn().Str("key", k).Msg("skip instance, missing key")
				break
			}
		}

		if r.Params.GetChildContentS("only_cluster_instance") != "true" {
			if instanceKey == "" {
				return true
			}
		}

		if instance = r.Matrix.GetInstance(instanceKey); instance == nil {
			if instance, err = r.Matrix.NewInstance(instanceKey); err != nil {
				r.Logger.Error().Err(err).Str("Instance key", instanceKey).Msg("")
				return true
			}
		}

		for label, display := range r.prop.instanceLabels {
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
				count++
			} else {
				// spams a lot currently due to missing label mappings. Moved to debug for now till rest gaps are filled
				r.Logger.Debug().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
			}
		}

		for key, metric := range r.Matrix.GetMetrics() {
			f := instanceData.Get(key)
			if f.Exists() {
				if err = metric.SetValueFloat64(instance, f.Float()); err != nil {
					r.Logger.Error().Err(err).Str("key", key).Str("metric", metric.GetName()).
						Msg("Unable to set float key on metric")
				}
				count++
			}
		}
		return true
	})

	// process endpoints
	err = r.processEndPoints(r.Matrix)
	if err != nil {
		r.Logger.Error().Err(err).Msg("Error while processing end points")
	}

	r.Logger.Info().
		Uint64("dataPoints", count).
		Str("apiTime", apiD.String()).
		Str("parseTime", parseD.String()).
		Msg("Collected")

	_ = r.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = r.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = r.Metadata.LazySetValueUint64("count", "data", count)
	r.AddCollectCount(count)

	return r.Matrix, nil
}

func (r *Rest) processEndPoints(data *matrix.Matrix) error {
	var (
		err error
	)

	for _, endpoint := range r.endpoints {
		var (
			records []interface{}
			content []byte
		)
		counterKey := make([]string, len(endpoint.prop.counters))
		i := 0
		for k := range endpoint.prop.counters {
			counterKey[i] = k
			i++
		}
		href := rest.BuildHref(endpoint.prop.query, strings.Join(counterKey, ","), nil, "", "", "", "", endpoint.prop.query)
		r.Logger.Debug().Str("href", href).Msg("")
		err = rest.FetchData(r.client, href, &records)
		if err != nil {
			r.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
			return err
		}

		all := rest.Pagination{
			Records:    records,
			NumRecords: len(records),
		}

		content, err = json.Marshal(all)
		if err != nil {
			r.Logger.Error().Err(err).Str("ApiPath", endpoint.prop.query).Msg("Unable to marshal rest pagination")
		}

		if !gjson.ValidBytes(content) {
			return fmt.Errorf("json is not valid for: %s", endpoint.prop.query)
		}

		results := gjson.GetManyBytes(content, "num_records", "records")
		numRecords := results[0]
		if numRecords.Int() == 0 {
			return errors.New(errors.ERR_NO_INSTANCE, "no "+endpoint.prop.query+" instances on cluster")
		}

		results[1].ForEach(func(key, instanceData gjson.Result) bool {
			var (
				instanceKey string
				instance    *matrix.Instance
			)

			if !instanceData.IsObject() {
				r.Logger.Warn().Str("type", instanceData.Type.String()).Msg("Instance data is not object, skipping")
				return true
			}

			// extract instance key(s)
			for _, k := range endpoint.prop.instanceKeys {
				value := instanceData.Get(k)
				if value.Exists() {
					instanceKey += value.String()
				} else {
					r.Logger.Warn().Str("key", k).Msg("skip instance, missing key")
					break
				}
			}

			if instance = data.GetInstance(instanceKey); instance != nil {

				for label, display := range endpoint.prop.instanceLabels {
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
						r.Logger.Debug().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
					}
				}

				for _, metric := range endpoint.prop.metrics {
					f := instanceData.Get(metric)
					if f.Exists() {
						if metr, ok := data.GetMetrics()[metric]; !ok {
							if metr, _ = data.NewMetricFloat64(metric); err != nil {
								r.Logger.Error().Err(err).
									Str("name", metric).
									Msg("NewMetricFloat64")
							}
							metr.SetName(endpoint.prop.instanceLabels[metric])
							if err = metr.SetValueFloat64(instance, f.Float()); err != nil {
								r.Logger.Error().Err(err).Str("key", metric).Str("metric", endpoint.prop.instanceLabels[metric]).
									Msg("Unable to set float key on metric")
							}
						} else {
							metr.SetName(endpoint.prop.instanceLabels[metric])
							if err = metr.SetValueFloat64(instance, f.Float()); err != nil {
								r.Logger.Error().Err(err).Str("key", metric).Str("metric", endpoint.prop.instanceLabels[metric]).
									Msg("Unable to set float key on metric")
							}
						}
					}
				}
			}
			return true
		})

	}

	return nil
}

func (r *Rest) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Disk":
		return disk.New(abc)
	default:
		r.Logger.Warn().Str("kind", kind).Msg("no rest plugin found ")
	}
	return nil
}

// Interface guards
var (
	_ collector.Collector = (*Rest)(nil)
)
