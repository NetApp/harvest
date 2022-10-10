package rest

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/certificate"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/disk"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/qtree"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/sensor"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/shelf"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/snapmirror"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/svm"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Rest struct {
	*collector.AbstractCollector
	Client    *rest.Client
	Prop      *prop
	endpoints []*endPoint
}

type endPoint struct {
	prop *prop
	name string
}

type prop struct {
	Object         string
	Query          string
	TemplatePath   string
	InstanceKeys   []string
	InstanceLabels map[string]string
	Metrics        map[string]*Metric
	Counters       map[string]string
	ReturnTimeOut  string
	Fields         []string
	APIType        string // public, private
	Filter         []string
}

type Metric struct {
	Label      string
	Name       string
	MetricType string
	Exportable bool
}

func init() {
	plugin.RegisterModule(&Rest{})
}

func (r *Rest) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.rest",
		New: func() plugin.Module { return new(Rest) },
	}
}

func (r *Rest) query(p *endPoint) string {
	return p.prop.Query
}

func (r *Rest) fields(p *endPoint) []string {
	return p.prop.Fields
}

func (r *Rest) filter(p *endPoint) []string {
	return p.prop.Filter
}

func (r *Rest) Init(a *collector.AbstractCollector) error {

	var err error

	r.AbstractCollector = a

	r.InitProp()

	if err = r.InitClient(); err != nil {
		return err
	}

	if r.Prop.TemplatePath, err = r.LoadTemplate(); err != nil {
		return err
	}

	if err = r.initEndPoints(); err != nil {
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
	r.Logger.Info().
		Int("numMetrics", len(r.Prop.Metrics)).
		Str("timeout", r.Client.Timeout.String()).
		Msg("initialized cache")

	return nil
}

func (r *Rest) InitClient() error {

	var err error
	a := r.AbstractCollector
	if r.Client, err = r.getClient(a, a.Params); err != nil {
		return err
	}

	if err = r.Client.Init(5); err != nil {
		return err
	}
	r.Client.TraceLogSet(r.Name, r.Params)

	return nil
}

func (r *Rest) InitMatrix() error {
	mat := r.Matrix[r.Object]
	// overwrite from abstract collector
	mat.Object = r.Prop.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", r.Client.Cluster().Name)

	if r.Params.HasChildS("labels") {
		for _, l := range r.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
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
		r.Logger.Error().Err(err).Str("poller", opt.Poller).Msgf("")
		return nil, err
	}
	if poller.Addr == "" {
		r.Logger.Error().Str("poller", opt.Poller).Msg("Address is empty")
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}

	clientTimeout := config.GetChildContentS("client_timeout")
	timeout := rest.DefaultTimeout * time.Second
	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		timeout = duration
	} else {
		r.Logger.Info().Str("timeout", timeout.String()).Msg("Using default timeout")
	}
	if client, err = rest.New(*poller, timeout); err != nil {
		r.Logger.Error().Err(err).Str("poller", opt.Poller).Msg("error creating new client")
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

			prop.InstanceKeys = make([]string, 0)
			prop.InstanceLabels = make(map[string]string)
			prop.Counters = make(map[string]string)
			prop.Metrics = make(map[string]*Metric)
			prop.APIType = "public"
			prop.ReturnTimeOut = r.Prop.ReturnTimeOut
			prop.TemplatePath = r.Prop.TemplatePath

			for _, line1 := range line.GetChildren() {
				if line1.GetNameS() == "query" {
					prop.Query = line1.GetContentS()
					prop.APIType = checkQueryType(prop.Query)
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

func (r *Rest) PollData() (map[string]*matrix.Matrix, error) {

	var (
		count        uint64
		apiD, parseD time.Duration
		startTime    time.Time
		err          error
		records      []gjson.Result
	)

	r.Logger.Debug().Msg("starting data poll")

	r.Matrix[r.Object].Reset()

	startTime = time.Now()

	href := rest.BuildHref(r.Prop.Query, strings.Join(r.Prop.Fields, ","), r.Prop.Filter, "", "", "", r.Prop.ReturnTimeOut, r.Prop.Query)

	if records, err = r.GetRestData(href); err != nil {
		return nil, err
	}

	apiD = time.Since(startTime)

	if len(records) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no "+r.Object+" instances on cluster")
	}

	startTime = time.Now()
	count = r.HandleResults(records, r.Prop, true)

	// process endpoints
	err = r.processEndPoints()
	if err != nil {
		r.Logger.Error().Err(err).Msg("Error while processing end points")
	}
	parseD = time.Since(startTime)

	numRecords := len(r.Matrix[r.Object].GetInstances())

	r.Logger.Info().
		Int("instances", numRecords).
		Uint64("metrics", count).
		Str("apiD", apiD.Round(time.Millisecond).String()).
		Str("parseD", parseD.Round(time.Millisecond).String()).
		Msg("Collected")

	_ = r.Metadata.LazySetValueInt64("count", "data", int64(numRecords))
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
			records []gjson.Result
		)
		counterKey := make([]string, len(endpoint.prop.Counters))
		i := 0
		for k := range endpoint.prop.Counters {
			counterKey[i] = k
			i++
		}

		href := rest.BuildHref(r.query(endpoint), strings.Join(r.fields(endpoint), ","), r.filter(endpoint), "", "", "", r.Prop.ReturnTimeOut, r.query(endpoint))

		if records, err = r.GetRestData(href); err != nil {
			r.Logger.Error().Err(err).
				Str("api", endpoint.prop.Query).
				Msg("")
			continue
		}

		if len(records) == 0 {
			r.Logger.Warn().Str("ApiPath", endpoint.prop.Query).Msg("no " + endpoint.prop.Query + " instances on cluster")
			continue
		}
		r.HandleResults(records, endpoint.prop, false)
	}

	return nil
}

// returns private if api endpoint has private keyword in it else public
func checkQueryType(query string) string {
	if strings.Contains(query, "private") {
		return "private"
	}
	return "public"
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
	case "Certificate":
		return certificate.New(abc)
	case "SVM":
		return svm.New(abc)
	case "Sensor":
		return sensor.New(abc)
	default:
		r.Logger.Warn().Str("kind", kind).Msg("no rest plugin found ")
	}
	return nil
}

// HandleResults function is used for handling the rest response for parent as well as endpoints calls,
// allowInstanceCreation would be true only for the parent rest, as only parent rest can create instance.
func (r *Rest) HandleResults(result []gjson.Result, prop *prop, allowInstanceCreation bool) uint64 {
	var (
		err   error
		count uint64
	)

	oldInstances := set.New()
	mat := r.Matrix[r.Object]

	// copy keys of current instances. This is used to remove deleted instances from matrix later
	for key := range mat.GetInstances() {
		oldInstances.Add(key)
	}

	for _, instanceData := range result {
		var (
			instanceKey string
			instance    *matrix.Instance
		)

		if !instanceData.IsObject() {
			r.Logger.Warn().Str("type", instanceData.Type.String()).Msg("Instance data is not object, skipping")
			continue
		}

		// extract instance key(s)
		for _, k := range prop.InstanceKeys {
			value := instanceData.Get(k)
			if value.Exists() {
				instanceKey += value.String()
			} else {
				r.Logger.Warn().Str("key", k).Msg("skip instance, missing key")
				break
			}
		}

		if instanceKey == "" {
			r.Logger.Trace().Msg("Instance key is empty, skipping")
			continue
		}

		instance = mat.GetInstance(instanceKey)

		// Used for endpoints as we don't want to create additional instances
		if !allowInstanceCreation && instance == nil {
			// Moved to trace as with filter, this log may spam
			r.Logger.Trace().Str("Instance key", instanceKey).Msg("Instance not found")
			continue
		}

		if instance == nil {
			if instance, err = mat.NewInstance(instanceKey); err != nil {
				r.Logger.Error().Err(err).Str("Instance key", instanceKey).Msg("")
				continue
			}
		}

		oldInstances.Delete(instanceKey)

		for label, display := range prop.InstanceLabels {
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
				r.Logger.Trace().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
			}
		}

		for _, metric := range prop.Metrics {
			metr, ok := mat.GetMetrics()[metric.Name]
			if !ok {
				if metr, err = mat.NewMetricFloat64(metric.Name); err != nil {
					r.Logger.Error().Err(err).
						Str("name", metric.Name).
						Msg("NewMetricFloat64")
				}
			}
			f := instanceData.Get(metric.Name)
			if f.Exists() {
				metr.SetName(metric.Label)

				var floatValue float64
				switch metric.MetricType {
				case "duration":
					floatValue = HandleDuration(f.String())
				case "timestamp":
					floatValue = HandleTimestamp(f.String())
				case "":
					floatValue = f.Float()
				default:
					r.Logger.Warn().Str("type", metric.MetricType).Str("metric", metric.Name).Msg("unknown metric type")
				}

				if err = metr.SetValueFloat64(instance, floatValue); err != nil {
					r.Logger.Error().Err(err).Str("key", metric.Name).Str("metric", metric.Label).
						Msg("Unable to set float key on metric")
				}
				count++
			}
		}

	}
	// remove deleted instances
	for key := range oldInstances.Iter() {
		mat.RemoveInstance(key)
		r.Logger.Debug().Str("key", key).Msg("removed instance")
	}

	return count
}

func (r *Rest) GetRestData(href string) ([]gjson.Result, error) {
	r.Logger.Debug().Str("href", href).Msg("")
	if href == "" {
		return nil, errs.New(errs.ErrConfig, "empty url")
	}

	result, err := rest.Fetch(r.Client, href)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	return result, nil
}

func (r *Rest) CollectAutoSupport(p *collector.Payload) {
	var exporterTypes []string
	for _, exporter := range r.Exporters {
		exporterTypes = append(exporterTypes, exporter.GetClass())
	}

	var counters = make([]string, 0)
	for k := range r.Prop.Counters {
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
		Query:     r.Prop.Query,
		Exporters: exporterTypes,
		Counters: collector.Counters{
			Count: len(counters),
			List:  counters,
		},
		Schedules:     schedules,
		ClientTimeout: r.Client.Timeout.String(),
	})

	if (r.Name == "Rest" && (r.Object == "Volume" || r.Object == "Node")) || r.Name == "ems" {
		version := r.Client.Cluster().Version
		p.Target.Version = strconv.Itoa(version[0]) + "." + strconv.Itoa(version[1]) + "." + strconv.Itoa(version[2])
		p.Target.Model = "cdot"
		if p.Target.Serial == "" {
			p.Target.Serial = r.Client.Cluster().UUID
		}
		p.Target.ClusterUUID = r.Client.Cluster().UUID

		md := r.GetMetadata()
		info := collector.InstanceInfo{
			Count:      md.LazyValueInt64("count", "data"),
			DataPoints: md.LazyValueInt64("datapoint_count", "data"),
			PollTime:   md.LazyValueInt64("poll_time", "data"),
			APITime:    md.LazyValueInt64("api_time", "data"),
			ParseTime:  md.LazyValueInt64("parse_time", "data"),
			PluginTime: md.LazyValueInt64("plugin_time", "data"),
		}

		if r.Object == "Node" || r.Name == "ems" {
			nodeIds, err := r.getNodeUuids()
			if err != nil {
				// log but don't return so the other info below is collected
				r.Logger.Error().
					Err(err).
					Msg("Unable to get nodes.")
				nodeIds = make([]collector.ID, 0)
			}
			info.Ids = nodeIds
			p.Nodes = &info

			// Since the serial number is bogus in c-mode
			// use the first node's serial number instead (the nodes were ordered in getNodeUuids())
			if len(nodeIds) > 0 {
				p.Target.Serial = nodeIds[0].SerialNumber
			}
		}
		if r.Object == "Volume" {
			p.Volumes = &info
		}
	}
}

func (r *Rest) getNodeUuids() ([]collector.ID, error) {
	var (
		records []gjson.Result
		err     error
		infos   []collector.ID
	)
	query := "api/cluster/nodes"

	href := rest.BuildHref(query, "serial_number,system_id", nil, "", "", "", r.Prop.ReturnTimeOut, query)

	if records, err = r.GetRestData(href); err != nil {
		return nil, err
	}

	for _, instanceData := range records {
		infos = append(infos, collector.ID{SerialNumber: instanceData.Get("serial_number").String(), SystemID: instanceData.Get("system_id").String()})
	}

	// When Harvest monitors a c-mode system, the first node is picked.
	// Sort so there's a higher chance the same node is picked each time this method is called
	sort.SliceStable(infos, func(i, j int) bool {
		return infos[i].SerialNumber < infos[j].SerialNumber
	})
	return infos, nil
}

func (r *Rest) InitProp() {
	r.Prop = &prop{}
	r.Prop.InstanceKeys = make([]string, 0)
	r.Prop.InstanceLabels = make(map[string]string)
	r.Prop.Counters = make(map[string]string)
	r.Prop.Metrics = make(map[string]*Metric)
}

// Interface guards
var (
	_ collector.Collector = (*Rest)(nil)
)
