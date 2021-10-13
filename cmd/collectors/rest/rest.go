package rest

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"os"
	"strconv"
	"strings"
	"time"
)

type Rest struct {
	*collector.AbstractCollector
	client         *rest.Client
	apiPath        string
	instanceKeys   []string
	instanceLabels map[string]string
	counters       map[string]string
	fields         []string
	misses         []string
	returnTimeOut  string
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
	if err = collector.Init(r); err != nil {
		return err
	}

	if r.client, err = r.getClient(a, a.Params); err != nil {
		return err
	}

	if err = r.client.Init(5); err != nil {
		return err
	}

	r.Logger.Info().Msgf("connected to %s: %s", r.client.ClusterName(), r.client.Info())

	r.Matrix.SetGlobalLabel("cluster", r.client.ClusterName())

	if err = r.initCache(r.getTemplateFn(), r.client.Version()); err != nil {
		return err
	}
	r.Logger.Info().Msgf("initialized cache with %d metrics", len(r.Matrix.GetMetrics()))
	return nil
}

func (r *Rest) getClient(a *collector.AbstractCollector, config *node.Node) (*rest.Client, error) {
	var (
		poller *conf.Poller
		addr   string
		err    error
		client *rest.Client
	)

	opt := a.GetOptions()
	if poller, err = conf.GetPoller2(opt.Config, opt.Poller); err != nil {
		r.Logger.Error().Stack().Err(err).Str("poller", opt.Poller).Msgf("")
		return nil, err
	}
	if addr = util.Value(poller.Addr, ""); addr == "" {
		r.Logger.Error().Stack().Str("poller", opt.Poller).Str("addr", addr).Msgf("Invalid address")
		return nil, errors.New(errors.MISSING_PARAM, "addr")
	}

	timeout := rest.DefaultTimeout

	if t, err := strconv.Atoi(config.GetChildContentS("client_timeout")); err == nil {
		timeout = time.Duration(t) * time.Second
	} else {
		// default timeout
		timeout = rest.DefaultTimeout
	}

	if client, err = rest.New(poller, timeout); err != nil {
		fmt.Printf("error creating new client %+v\n", err)
		os.Exit(1)
	}

	return client, err
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
	if gjson.Parse(source).IsObject() {
		arr = gjson.Parse(source).Map()
	} else {
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

	r.Logger.Info().Msgf("starting data poll")
	r.Matrix.Reset()

	startTime = time.Now()
	// Check if fields are set for rest call
	if len(r.fields) > 0 {
		href := rest.BuildHref(r.apiPath, strings.Join(r.fields[:], ","), nil, "", "", "", r.returnTimeOut)
		r.Logger.Info().Msgf("rest end point [%s]", href)
		err = rest.FetchData(r.client, href, &records)
		if err != nil {
			r.Logger.Error().Stack().Err(err).Msgf("")
			return nil, err
		}
	} else {
		href := rest.BuildHref(r.apiPath, "*", nil, "", "", "", r.returnTimeOut)
		r.Logger.Info().Msgf("rest end point [%s]", href)
		err = rest.FetchData(r.client, href, &records)
		if err != nil {
			r.Logger.Error().Stack().Err(err).Msgf("")
			return nil, err
		}
		all := rest.Pagination{
			Records:    records,
			NumRecords: len(records),
		}
		c, err := json.Marshal(all)
		if err != nil {
			r.Logger.Error().Stack().Err(err).Msgf("")
			return nil, err
		} else {
			results := gjson.GetBytes(c, "records")
			if len(results.String()) > 0 {
				// fetch first record from json
				firstValue := results.Get("1")
				res := getFieldName(firstValue.String(), "")
				var searchKeys []string
				for k := range r.counters {
					searchKeys = append(searchKeys, k)
				}
				// find keys from rest json response which matches counters defined in templates
				matches, misses := util.Intersection(res, searchKeys)
				for _, param := range matches {
					r.fields = append(r.fields, param.(string))
				}
				for _, param := range misses {
					r.misses = append(r.misses, param.(string))
				}
			}
		}
	}

	if len(r.misses) > 0 {
		r.Logger.Warn().
			Str("Missing Counters", strings.Join(r.misses[:], ",")).
			Str("ApiPath", r.apiPath).
			Msg("Mis configured counters")
	}

	all := rest.Pagination{
		Records:    records,
		NumRecords: len(records),
	}
	apiD = time.Since(startTime)

	content, err = json.Marshal(all)
	if err != nil {
		r.Logger.Error().Stack().Err(err).Msgf("")
	}

	startTime = time.Now()
	if !gjson.ValidBytes(content) {
		return nil, fmt.Errorf("json is not valid for: %s", r.apiPath)
	}
	parseD = time.Since(startTime)

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no "+r.Object+" instances on cluster")
	}

	r.Logger.Debug().Msgf("extracted %d [%s] instances", numRecords, r.Object)

	results[1].ForEach(func(key, instanceData gjson.Result) bool {
		var (
			instanceKey string
			instance    *matrix.Instance
		)

		if !instanceData.IsObject() {
			r.Logger.Warn().Str("type", instanceData.Type.String()).Msg("skip instance")
			return true
		}

		// extract instance key(s)
		for _, k := range r.instanceKeys {
			value := instanceData.Get(k)
			if value.Exists() {
				instanceKey += value.String()
			} else {
				r.Logger.Warn().Str("key", k).Msg("skip instance, missing key")
				break
			}
		}

		if instanceKey == "" {
			return true
		}

		if instance = r.Matrix.GetInstance(instanceKey); instance == nil {
			if instance, err = r.Matrix.NewInstance(instanceKey); err != nil {
				r.Logger.Error().Msgf("NewInstance [key=%s]: %v", instanceKey, err)
				return true
			}
		}

		for label, display := range r.instanceLabels {
			value := instanceData.Get(label)
			if value.Exists() {
				instance.SetLabel(display, value.String())
				count++
			}
		}

		for key, metric := range r.Matrix.GetMetrics() {

			if metric.GetProperty() == "etl.bool" {
				b := instanceData.Get(key)
				if b.Exists() {
					if err = metric.SetValueBool(instance, b.Bool()); err != nil {
						r.Logger.Error().Err(err).Str("key", key).Msg("SetValueBool metric")
					}
					count++
				}
			} else if metric.GetProperty() == "etl.float" {
				f := instanceData.Get(key)
				if f.Exists() {
					if err = metric.SetValueFloat64(instance, f.Float()); err != nil {
						r.Logger.Error().Err(err).Str("key", key).Msg("SetValueFloat64 metric")
					}
					count++
				}
			}
		}
		return true
	})

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

// Interface guards
var (
	_ collector.Collector = (*Rest)(nil)
)
