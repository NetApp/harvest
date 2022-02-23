/*
	Copyright NetApp Inc, 2022 All rights reserved

	Security collects and processes metrics from the "security" APIs of the
	REST protocol. This collector inherits some methods and fields of
	the Rest collector (as they use the same protocol).
*/

package security

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/collectors/security/plugins/certificate"
	"goharvest2/cmd/collectors/security/plugins/securityaccount"
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//type Security struct {
//	*rest.Rest // provides: AbstractCollector, Client, Prop, Endpoints
//}
//
//func init() {
//	plugin.RegisterModule(Security{})
//}
//
//func (Security) HarvestModule() plugin.ModuleInfo {
//	return plugin.ModuleInfo{
//		ID:  "harvest.collector.security",
//		New: func() plugin.Module { return new(Security) },
//	}
//}

type Security struct {
	*collector.AbstractCollector
	Client    *rest.Client
	prop      *Prop
	endpoints []*EndPoint
}

type EndPoint struct {
	prop *Prop
	name string
}

type Prop struct {
	object         string
	query          string
	templatePath   string
	instanceKeys   []string
	instanceLabels map[string]string
	metrics        []metric
	counters       map[string]string
	returnTimeOut  string
	fields         []string
	apiType        string // public, private
}

type metric struct {
	label      string
	name       string
	metricType string
}

func init() {
	plugin.RegisterModule(Security{})
}

func (Security) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.security",
		New: func() plugin.Module { return new(Security) },
	}
}

func (r *Security) Init(a *collector.AbstractCollector) error {

	var err error

	r.AbstractCollector = a

	r.prop = &Prop{}

	if r.Client, err = r.GetClient(a, a.Params); err != nil {
		return err
	}

	if err = r.Client.Init(5); err != nil {
		return err
	}

	if err = r.LoadTemplate(); err != nil {
		return err
	}

	if err = r.InitEndPoints(); err != nil {
		return err
	}

	if err = collector.Init(r); err != nil {
		return err
	}

	if err = r.InitCache(); err != nil {
		return err
	}

	if err = r.InitMatrix(); err != nil {
		return err
	}
	r.Logger.Info().Msgf("initialized cache with %d metrics", len(r.Matrix.GetMetrics()))

	return nil
}

func (r *Security) InitMatrix() error {
	// overwrite from abstract collector
	r.Matrix.Object = r.prop.object
	// Add system (cluster) name
	r.Matrix.SetGlobalLabel("cluster", r.Client.Cluster().Name)
	if r.Params.HasChildS("labels") {
		for _, l := range r.Params.GetChildS("labels").GetChildren() {
			r.Matrix.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}

	return nil
}

func (r *Security) GetClient(a *collector.AbstractCollector, config *node.Node) (*rest.Client, error) {
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
	if client, err = rest.New(*poller, timeout); err != nil {
		r.Logger.Error().Stack().Err(err).Str("poller", opt.Poller).Msg("error creating new client")
		os.Exit(1)
	}

	return client, err
}

func (r *Security) InitEndPoints() error {

	endpoints := r.Params.GetChildS("endpoints")
	if endpoints != nil {
		for _, line := range endpoints.GetChildren() {

			n := line.GetNameS()
			e := EndPoint{name: n}

			prop := Prop{}

			prop.instanceKeys = make([]string, 0)
			prop.instanceLabels = make(map[string]string)
			prop.counters = make(map[string]string)
			prop.metrics = make([]metric, 0)
			prop.apiType = "public"
			prop.returnTimeOut = r.prop.returnTimeOut

			for _, line1 := range line.GetChildren() {
				if line1.GetNameS() == "query" {
					prop.query = line1.GetContentS()
					prop.apiType = checkQueryType(prop.query)
				}
				if line1.GetNameS() == "counters" {
					r.ParseRestCounters(line1, &prop)
				}
			}
			e.prop = &prop
			r.endpoints = append(r.endpoints, &e)
		}
	}
	return nil
}

func (r *Security) getTemplateFn() string {
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

func (r *Security) PollData() (*matrix.Matrix, error) {

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

	if records, err = r.GetRestData(r.prop, r.Client); err != nil {
		r.Logger.Error().Stack().Err(err).Msg("Failed to fetch data")
		return nil, err
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

	if !gjson.ValidBytes(content) {
		return nil, fmt.Errorf("json is not valid for: %s", r.prop.query)
	}

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no "+r.Object+" instances on cluster")
	}

	r.Logger.Debug().Str("object", r.Object).Str("number of records extracted", numRecords.String()).Msg("")

	count = r.HandleResults(results[1], r.prop, true)

	// process endpoints
	startTime = time.Now()
	err = r.ProcessEndPoints()
	if err != nil {
		r.Logger.Error().Err(err).Msg("Error while processing end points")
	}
	parseD = time.Since(startTime)

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

func (r *Security) ProcessEndPoints() error {
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

		if records, err = r.GetRestData(endpoint.prop, r.Client); err != nil {
			r.Logger.Error().Stack().Err(err).Msg("Failed to fetch data")
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

		r.HandleResults(results[1], endpoint.prop, false)
	}

	return nil
}

// returns private if api endpoint has private keyword in it else public
func checkQueryType(query string) string {
	if strings.Contains(query, "private") {
		return "private"
	} else {
		return "public"
	}
}

// HandleResults function is used for handling the rest response for parent as well as endpoints calls,
// allowInstanceCreation would be true only for the parent rest, as only parent rest can create instance.
func (r *Security) HandleResults(result gjson.Result, prop *Prop, allowInstanceCreation bool) uint64 {
	var (
		err   error
		count uint64
	)

	result.ForEach(func(key, instanceData gjson.Result) bool {
		var (
			instanceKey string
			instance    *matrix.Instance
		)

		if !instanceData.IsObject() {
			r.Logger.Warn().Str("type", instanceData.Type.String()).Msg("Instance data is not object, skipping")
			return true
		}

		// extract instance key(s)
		for _, k := range prop.instanceKeys {
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

		instance = r.Matrix.GetInstance(instanceKey)

		if !allowInstanceCreation && instance == nil {
			r.Logger.Warn().Str("Instance key", instanceKey).Msg("Instance not found")
			return true
		}

		if instance == nil {
			if instance, err = r.Matrix.NewInstance(instanceKey); err != nil {
				r.Logger.Error().Err(err).Str("Instance key", instanceKey).Msg("")
				return true
			}
		}

		for label, display := range prop.instanceLabels {
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

		for _, metric := range prop.metrics {
			metr, ok := r.Matrix.GetMetrics()[metric.name]
			if !ok {
				if metr, err = r.Matrix.NewMetricFloat64(metric.name); err != nil {
					r.Logger.Error().Err(err).
						Str("name", metric.name).
						Msg("NewMetricFloat64")
				}
			}
			f := instanceData.Get(metric.name)
			if f.Exists() {
				metr.SetName(metric.label)

				var floatValue float64
				switch metric.metricType {
				case "duration":
					floatValue = HandleDuration(f.String())
				case "timestamp":
					floatValue = HandleTimestamp(f.String())
				case "":
					floatValue = f.Float()
				default:
					r.Logger.Warn().Str("type", metric.metricType).Str("metric", metric.name).Msg("unknown metric type")
				}

				if err = metr.SetValueFloat64(instance, floatValue); err != nil {
					r.Logger.Error().Err(err).Str("key", metric.name).Str("metric", metric.label).
						Msg("Unable to set float key on metric")
				}
				count++
			}
		}

		return true
	})
	return count
}

func (r *Security) GetRestData(prop *Prop, client *rest.Client) ([]interface{}, error) {
	var (
		err     error
		records []interface{}
	)

	href := rest.BuildHref(prop.query, strings.Join(prop.fields, ","), nil, "", "", "", prop.returnTimeOut, prop.query)

	r.Logger.Debug().Str("href", href).Msg("")
	if href == "" {
		return nil, errors.New(errors.ERR_CONFIG, "empty url")
	}

	err = rest.FetchData(client, href, &records)
	if err != nil {
		r.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
		return nil, err
	}

	return records, nil
}

func (r *Security) LoadTemplate() error {

	var (
		template     *node.Node
		templatePath string
		err          error
	)

	// import template
	if template, templatePath, err = r.ImportSubTemplate("", r.getTemplateFn(), r.Client.Cluster().Version); err != nil {
		return err
	}

	r.prop.templatePath = templatePath

	r.Params.Union(template)
	return nil
}

func (r *Security) InitCache() error {

	var (
		counters *node.Node
	)

	if x := r.Params.GetChildContentS("object"); x != "" {
		r.prop.object = x
	} else {
		r.prop.object = strings.ToLower(r.Object)
	}

	if e := r.Params.GetChildS("export_options"); e != nil {
		r.Matrix.SetExportOptions(e)
	}

	if r.prop.query = r.Params.GetChildContentS("query"); r.prop.query == "" {
		return errors.New(errors.MISSING_PARAM, "query")
	}

	// create metric cache
	if counters = r.Params.GetChildS("counters"); counters == nil {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	// default value for ONTAP is 15 sec
	if returnTimeout := r.Params.GetChildContentS("return_timeout"); returnTimeout != "" {
		r.prop.returnTimeOut = returnTimeout
	}

	r.prop.instanceKeys = make([]string, 0)
	r.prop.instanceLabels = make(map[string]string)
	r.prop.counters = make(map[string]string)

	// private end point do not support * as fields. We need to pass fields in endpoint
	query := r.Params.GetChildS("query")
	r.prop.apiType = "public"
	if query != nil {
		r.prop.apiType = checkQueryType(query.GetContentS())
	}

	r.ParseRestCounters(counters, r.prop)

	r.Logger.Info().Strs("extracted Instance Keys", r.prop.instanceKeys).Msg("")
	r.Logger.Info().Int("count metrics", len(r.prop.metrics)).Int("count labels", len(r.prop.instanceLabels)).Msg("initialized metric cache")

	return nil
}

func (r *Security) ParseRestCounters(counter *node.Node, prop *Prop) {
	var (
		display, name, kind, metricType string
	)

	for _, c := range counter.GetAllChildContentS() {
		if c != "" {
			name, display, kind, metricType = util.ParseMetric(c)
			r.Logger.Debug().
				Str("kind", kind).
				Str("name", name).
				Str("display", display).
				Msg("Collected")

			prop.counters[name] = display
			switch kind {
			case "key":
				prop.instanceLabels[name] = display
				prop.instanceKeys = append(prop.instanceKeys, name)
			case "label":
				prop.instanceLabels[name] = display
			case "float":
				m := metric{label: display, name: name, metricType: metricType}
				prop.metrics = append(prop.metrics, m)
			}
		}
	}

	if prop.apiType == "private" {
		counterKey := make([]string, len(prop.counters))
		i := 0
		for k := range prop.counters {
			counterKey[i] = k
			i++
		}
		prop.fields = counterKey
	}

	if prop.apiType == "public" {
		prop.fields = []string{"*"}
		if counter != nil {
			if x := counter.GetChildS("hidden_fields"); x != nil {
				prop.fields = append(prop.fields, x.GetAllChildContentS()...)
			}
		}
	}

}

func HandleDuration(value string) float64 {
	// Example: duration: PT8H35M42S
	timeDurationRegex := `^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:.\d+)?)S)?$`

	regexTimeDuration := regexp.MustCompile(timeDurationRegex)
	if match := regexTimeDuration.MatchString(value); match {
		// example: PT8H35M42S   ==>  30942
		matches := regexTimeDuration.FindStringSubmatch(value)
		if matches == nil {
			return 0
		}

		seconds := 0.0

		//years
		//months

		//days
		if matches[3] != "" {
			f, err := strconv.ParseFloat(matches[3], 64)
			if err != nil {
				fmt.Printf("%v", err)
				return 0
			}
			seconds += f * 24 * 60 * 60
		}

		//hours
		if matches[4] != "" {
			f, err := strconv.ParseFloat(matches[4], 64)
			if err != nil {
				fmt.Printf("%v", err)
				return 0
			}
			seconds += f * 60 * 60
		}

		//minutes
		if matches[5] != "" {
			f, err := strconv.ParseFloat(matches[5], 64)
			if err != nil {
				fmt.Printf("%v", err)
				return 0
			}
			seconds += f * 60
		}

		//seconds & milliseconds
		if matches[6] != "" {
			f, err := strconv.ParseFloat(matches[6], 64)
			if err != nil {
				fmt.Printf("%v", err)
				return 0
			}
			seconds += f
		}
		return seconds
	}

	return 0
}

func HandleTimestamp(value string) float64 {
	var timestamp time.Time
	var err error

	// Example: timestamp: 2020-12-02T18:36:19-08:00
	timestampRegex := `[+-]?\d{4}(-[01]\d(-[0-3]\d(T[0-2]\d:[0-5]\d:?([0-5]\d(\.\d+)?)?[+-][0-2]\d:[0-5]\d?)?)?)?`

	regexTimeStamp := regexp.MustCompile(timestampRegex)
	if match := regexTimeStamp.MatchString(value); match {
		// example: 2020-12-02T18:36:19-08:00   ==>  1606962979
		if timestamp, err = time.Parse(time.RFC3339, value); err != nil {
			fmt.Printf("%v", err)
			return 0
		}
		return float64(timestamp.Unix())
	}
	return 0
}

func (s *Security) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Certificate":
		return certificate.New(abc)
	case "SecurityAccount":
		return securityaccount.New(abc)
	default:
		s.Logger.Warn().Str("kind", kind).Msg("no security plugin found ")
	}
	return nil
}

// Interface guards
var (
	_ collector.Collector = (*Security)(nil)
)
