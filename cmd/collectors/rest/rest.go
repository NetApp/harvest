package rest

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/collectors/rest/plugins/disk"
	"goharvest2/cmd/collectors/rest/plugins/qtree"
	"goharvest2/cmd/collectors/rest/plugins/shelf"
	"goharvest2/cmd/collectors/rest/plugins/snapmirror"
	"goharvest2/cmd/collectors/rest/plugins/volume"
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"os"
	"sort"
	"strconv"
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
	object         string
	query          string
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
	plugin.RegisterModule(Rest{})
}

func (Rest) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.rest",
		New: func() plugin.Module { return new(Rest) },
	}
}

func (r *Rest) query(p *endPoint) string {
	return p.prop.query
}

func (r *Rest) fields(p *endPoint) []string {
	return p.prop.fields
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

	if err = r.LoadTemplate(); err != nil {
		return err
	}

	if err = r.initEndPoints(); err != nil {
		return err
	}

	if err = collector.Init(r); err != nil {
		return err
	}

	if err = r.initCache(); err != nil {
		return err
	}

	if err = r.InitMatrix(); err != nil {
		return err
	}
	r.Logger.Info().Msgf("initialized cache with %d metrics", len(r.Matrix.GetMetrics()))

	return nil
}

func (r *Rest) InitMatrix() error {
	// overwrite from abstract collector
	r.Matrix.Object = r.prop.object
	// Add system (cluster) name
	r.Matrix.SetGlobalLabel("cluster", r.client.Cluster().Name)
	if r.Params.HasChildS("labels") {
		for _, l := range r.Params.GetChildS("labels").GetChildren() {
			r.Matrix.SetGlobalLabel(l.GetNameS(), l.GetContentS())
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
	if client, err = rest.New(*poller, timeout); err != nil {
		r.Logger.Error().Stack().Err(err).Str("poller", opt.Poller).Msg("error creating new client")
		os.Exit(1)
	}

	return client, err
}

func (r *Rest) initEndPoints() error {

	endpoints := r.Params.GetChildS("endpoints")
	if endpoints != nil {
		for _, line := range endpoints.GetChildren() {

			n := line.GetNameS()
			e := endPoint{name: n}

			prop := prop{}

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

	if records, err = r.GetRestData(r.prop.query, strings.Join(r.prop.fields, ","), r.prop.returnTimeOut, r.client); err != nil {
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
	err = r.processEndPoints()
	if err != nil {
		r.Logger.Error().Err(err).Msg("Error while processing end points")
	}
	parseD = time.Since(startTime)

	r.Logger.Info().
		Uint64("instances", numRecords.Uint()).
		Uint64("dataPoints", count).
		Str("apiTime", apiD.String()).
		Str("parseTime", parseD.String()).
		Msg("Collected")

	_ = r.Metadata.LazySetValueInt64("count", "data", numRecords.Int())
	_ = r.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = r.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = r.Metadata.LazySetValueUint64("datapoint_count", "data", count)
	r.AddCollectCount(count)

	return r.Matrix, nil
}

func (r *Rest) processEndPoints() error {
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

		if records, err = r.GetRestData(r.query(endpoint), strings.Join(r.fields(endpoint), ","), r.prop.returnTimeOut, r.client); err != nil {
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

func (r *Rest) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Disk":
		return disk.New(abc)
	case "Qtree":
		return qtree.New(abc)
	case "Shelf":
		return shelf.New(abc)
	case "Snapmirror":
		return snapmirror.New(abc)
	case "Volume":
		return volume.New(abc)
	default:
		r.Logger.Warn().Str("kind", kind).Msg("no rest plugin found ")
	}
	return nil
}

// HandleResults function is used for handling the rest response for parent as well as endpoints calls,
// allowInstanceCreation would be true only for the parent rest, as only parent rest can create instance.
func (r *Rest) HandleResults(result gjson.Result, prop *prop, allowInstanceCreation bool) uint64 {
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

func (r *Rest) GetRestData(query string, fields string, returnTimeOut string, client *rest.Client) ([]interface{}, error) {
	var (
		err     error
		records []interface{}
	)

	href := rest.BuildHref(query, fields, nil, "", "", "", returnTimeOut, query)

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

func (r *Rest) CollectAutoSupport(p *collector.Payload) {
	var exporterTypes []string
	for _, exporter := range r.Exporters {
		exporterTypes = append(exporterTypes, exporter.GetClass())
	}

	var counters = make([]string, 0)
	for k := range r.prop.counters {
		counters = append(counters, k)
	}

	var schedules = make([]collector.Schedule, 0)
	tasks := r.Params.GetChildS("schedule")
	if tasks != nil && len(tasks.GetChildren()) > 0 {
		for _, task := range tasks.GetChildren() {
			schedules = append(schedules, collector.Schedule{
				Name:     task.GetNameS(),
				Schedule: task.GetContentS(),
			})
		}
	}

	// Add collector information
	p.AddCollectorAsup(collector.AsupCollector{
		Name:      r.Name,
		Query:     r.prop.query,
		Exporters: exporterTypes,
		Counters: collector.Counters{
			Count: len(counters),
			List:  counters,
		},
		Schedules:     schedules,
		ClientTimeout: r.client.Timeout.String(),
	})

	if r.Name == "Rest" && (r.Object == "Volume" || r.Object == "Node") {
		version := r.client.Cluster().Version
		p.Target.Version = strconv.Itoa(version[0]) + "." + strconv.Itoa(version[1]) + "." + strconv.Itoa(version[2])
		p.Target.Model = "cdot"
		if p.Target.Serial == "" {
			p.Target.Serial = r.client.Cluster().Uuid
		}
		p.Target.ClusterUuid = r.client.Cluster().Uuid

		md := r.GetMetadata()
		info := collector.InstanceInfo{
			Count:      md.LazyValueInt64("count", "data"),
			DataPoints: md.LazyValueInt64("datapoint_count", "data"),
			PollTime:   md.LazyValueInt64("poll_time", "data"),
			ApiTime:    md.LazyValueInt64("api_time", "data"),
			ParseTime:  md.LazyValueInt64("parse_time", "data"),
			PluginTime: md.LazyValueInt64("plugin_time", "data"),
		}

		if r.Object == "Node" {
			nodeIds, err := r.getNodeUuids()
			if err != nil {
				// log but don't return so the other info below is collected
				r.Logger.Error().
					Err(err).
					Msg("Unable to get nodes.")
				nodeIds = make([]collector.Id, 0)
			}
			info.Ids = nodeIds
			p.Nodes = &info

			// Since the serial number is bogus in c-mode
			// use the first node's serial number instead (the nodes were ordered in getNodeUuids())
			if len(nodeIds) > 0 {
				p.Target.Serial = nodeIds[0].SerialNumber
			}
		} else if r.Object == "Volume" {
			p.Volumes = &info
		}
	}
}

func (r *Rest) getNodeUuids() ([]collector.Id, error) {
	var (
		records []interface{}
		err     error
		infos   []collector.Id
	)
	query := "api/cluster/nodes"

	if records, err = r.GetRestData(query, "serial_number,system_id", r.prop.returnTimeOut, r.client); err != nil {
		r.Logger.Error().Stack().Err(err).Msg("Failed to fetch data")
		return nil, err
	}

	all := rest.Pagination{
		Records:    records,
		NumRecords: len(records),
	}

	content, err := json.Marshal(all)
	if err != nil {
		r.Logger.Error().Err(err).Str("ApiPath", query).Msg("Unable to marshal rest pagination")
	}

	if !gjson.ValidBytes(content) {
		return nil, fmt.Errorf("json is not valid for: %s", r.prop.query)
	}

	results := gjson.GetManyBytes(content, "num_records", "records")

	results[1].ForEach(func(key, instanceData gjson.Result) bool {
		infos = append(infos, collector.Id{SerialNumber: instanceData.Get("serial_number").String(), SystemId: instanceData.Get("system_id").String()})
		return true
	})

	// When Harvest monitors a c-mode system, the first node is picked.
	// Sort so there's a higher chance the same node is picked each time this method is called
	sort.SliceStable(infos, func(i, j int) bool {
		return infos[i].SerialNumber < infos[j].SerialNumber
	})
	return infos, nil
}

// Interface guards
var (
	_ collector.Collector = (*Rest)(nil)
)
