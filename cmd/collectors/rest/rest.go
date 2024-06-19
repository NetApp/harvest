package rest

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/certificate"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/disk"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/health"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/metroclustercheck"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/netroute"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/ontaps3service"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/qospolicyadaptive"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/qospolicyfixed"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/qtree"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/securityaccount"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/securitycertificate"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/shelf"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/snapmirror"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/svm"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/systemnode"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/volumeanalytics"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/workload"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Regular expression to match dot notation or single word without dot
// It allows one or more alphanumeric characters or underscores, optionally followed by a dot and more characters
// This pattern can repeat any number of times
// This version does not allow numeric-only segments
var validPropRegex = regexp.MustCompile(`^([a-zA-Z_]\w*\.)*[a-zA-Z_]\w*$`)

type Rest struct {
	*collector.AbstractCollector
	Client                       *rest.Client
	Prop                         *prop
	endpoints                    []*endPoint
	isIgnoreUnknownFieldsEnabled bool
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
	ReturnTimeOut  *int
	Fields         []string
	HiddenFields   []string
	IsPublic       bool
	Filter         []string
	Href           string
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

func (r *Rest) isValidFormat(prop *prop) bool {
	for _, str := range prop.Fields {
		if !validPropRegex.MatchString(str) {
			return false
		}
	}
	return true
}

func (r *Rest) Fields(prop *prop) []string {
	fields := prop.Fields
	if prop.IsPublic {
		// applicable for public API only
		if !r.isIgnoreUnknownFieldsEnabled || !r.isValidFormat(prop) {
			fields = []string{"*"}
		}
	}
	if len(prop.HiddenFields) > 0 {
		fieldsMap := make(map[string]bool)
		for _, field := range fields {
			fieldsMap[field] = true
		}

		// append hidden fields
		for _, hiddenField := range prop.HiddenFields {
			if _, exists := fieldsMap[hiddenField]; !exists {
				fields = append(fields, hiddenField)
				fieldsMap[hiddenField] = true
			}
		}
	}
	return fields
}

func (r *Rest) filter(p *endPoint) []string {
	return p.prop.Filter
}

func (r *Rest) Init(a *collector.AbstractCollector) error {

	var err error

	r.AbstractCollector = a

	r.InitProp()

	if err := r.InitClient(); err != nil {
		return err
	}

	if r.Prop.TemplatePath, err = r.LoadTemplate(); err != nil {
		return err
	}

	r.InitVars(a.Params)

	if err := r.initEndPoints(); err != nil {
		return err
	}

	if err := collector.Init(r); err != nil {
		return err
	}

	if err := r.InitCache(); err != nil {
		return err
	}

	if err := r.InitMatrix(); err != nil {
		return err
	}

	r.Logger.Debug().
		Int("numMetrics", len(r.Prop.Metrics)).
		Str("timeout", r.Client.Timeout.String()).
		Msg("initialized cache")

	return nil
}

func (r *Rest) InitVars(config *node.Node) {

	var err error

	clientTimeout := config.GetChildContentS("client_timeout")
	if clientTimeout == "" {
		clientTimeout = rest.DefaultTimeout
	}

	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		r.Client.Timeout = duration
	} else {
		r.Logger.Info().Str("timeout", rest.DefaultTimeout).Msg("Using default timeout")
	}
}

func (r *Rest) InitClient() error {

	var err error
	a := r.AbstractCollector
	if r.Client, err = r.getClient(a, r.Auth); err != nil {
		return err
	}

	if r.Options.IsTest {
		return nil
	}
	if err := r.Client.Init(5); err != nil {
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

func (r *Rest) getClient(a *collector.AbstractCollector, c *auth.Credentials) (*rest.Client, error) {
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
	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if a.Options.IsTest {
		return &rest.Client{Metadata: &util.Metadata{}}, nil
	}
	if client, err = rest.New(poller, timeout, c); err != nil {
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

			p := prop{}

			p.InstanceKeys = make([]string, 0)
			p.InstanceLabels = make(map[string]string)
			p.Counters = make(map[string]string)
			p.Metrics = make(map[string]*Metric)
			p.IsPublic = true
			p.ReturnTimeOut = r.Prop.ReturnTimeOut
			p.TemplatePath = r.Prop.TemplatePath

			for _, line1 := range line.GetChildren() {
				if line1.GetNameS() == "query" {
					p.Query = line1.GetContentS()
					p.IsPublic = util.IsPublicAPI(p.Query)
				}
				if line1.GetNameS() == "counters" {
					r.ParseRestCounters(line1, &p)
				}
			}
			e.prop = &p
			r.endpoints = append(r.endpoints, &e)
		}
	}
	return nil
}

func TemplateFn(n *node.Node, obj string) string {
	var fn string
	objects := n.GetChildS("objects")
	if objects != nil {
		fn = objects.GetChildContentS(obj)
	}
	return fn
}

// Returns a slice of keys in dot notation from json
func getFieldName(source string, parent string) []string {
	res := make([]string, 0)
	var arr map[string]gjson.Result
	r := gjson.Parse(source)
	switch {
	case r.IsArray():
		newR := r.Get("0")
		arr = newR.Map()
	case r.IsObject():
		arr = r.Map()
	default:
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

// PollCounter performs daily tasks such as updating the cluster info and caching href.
func (r *Rest) PollCounter() (map[string]*matrix.Matrix, error) {

	startTime := time.Now()
	// Update the cluster info to track if customer version is updated
	err := r.Client.UpdateClusterInfo(5)
	if err != nil {
		return nil, err
	}
	apiD := time.Since(startTime)

	startTime = time.Now()
	v, err := util.VersionAtLeast(r.Client.Cluster().GetVersion(), "9.11.1")
	if err != nil {
		return nil, err
	}
	// Check the version if it is 9.11.1 then pass relevant fields and not *
	if v {
		r.isIgnoreUnknownFieldsEnabled = true
	} else {
		r.isIgnoreUnknownFieldsEnabled = false
	}
	r.updateHref()
	parseD := time.Since(startTime)

	// update metadata for collector logs
	_ = r.Metadata.LazySetValueInt64("api_time", "counter", apiD.Microseconds())
	_ = r.Metadata.LazySetValueInt64("parse_time", "counter", parseD.Microseconds())
	return nil, nil
}

func (r *Rest) updateHref() {
	r.Prop.Href = rest.NewHrefBuilder().
		APIPath(r.Prop.Query).
		Fields(r.Fields(r.Prop)).
		Filter(r.Prop.Filter).
		ReturnTimeout(r.Prop.ReturnTimeOut).
		IsIgnoreUnknownFieldsEnabled(r.isIgnoreUnknownFieldsEnabled).
		Build()

	for _, e := range r.endpoints {
		e.prop.Href = rest.NewHrefBuilder().
			APIPath(r.query(e)).
			Fields(r.Fields(e.prop)).
			Filter(r.filter(e)).
			ReturnTimeout(r.Prop.ReturnTimeOut).
			IsIgnoreUnknownFieldsEnabled(r.isIgnoreUnknownFieldsEnabled).
			Build()
	}
}

func (r *Rest) PollData() (map[string]*matrix.Matrix, error) {

	var (
		startTime time.Time
		err       error
		records   []gjson.Result
	)

	r.Matrix[r.Object].Reset()
	r.Client.Metadata.Reset()

	startTime = time.Now()

	if records, err = r.GetRestData(r.Prop.Href); err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no "+r.Object+" instances on cluster")
	}

	return r.pollData(startTime, records, func(e *endPoint) ([]gjson.Result, time.Duration, error) {
		return r.processEndPoint(e)
	})
}

func (r *Rest) pollData(
	startTime time.Time,
	records []gjson.Result,
	endpointFunc func(e *endPoint) ([]gjson.Result, time.Duration, error),
) (map[string]*matrix.Matrix, error) {

	var (
		count        uint64
		apiD, parseD time.Duration
	)

	apiD = time.Since(startTime)
	startTime = time.Now()

	count = r.HandleResults(records, r.Prop, false)

	// process endpoints
	eCount, endpointAPID := r.processEndPoints(endpointFunc)
	count += eCount
	parseD = time.Since(startTime)

	numRecords := len(r.Matrix[r.Object].GetInstances())

	_ = r.Metadata.LazySetValueInt64("api_time", "data", (apiD + endpointAPID).Microseconds())
	_ = r.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = r.Metadata.LazySetValueUint64("metrics", "data", count)
	_ = r.Metadata.LazySetValueUint64("instances", "data", uint64(numRecords))
	_ = r.Metadata.LazySetValueUint64("bytesRx", "data", r.Client.Metadata.BytesRx)
	_ = r.Metadata.LazySetValueUint64("numCalls", "data", r.Client.Metadata.NumCalls)

	r.AddCollectCount(count)

	return r.Matrix, nil
}

func (r *Rest) processEndPoint(e *endPoint) ([]gjson.Result, time.Duration, error) {
	now := time.Now()
	data, err := r.GetRestData(e.prop.Href)
	if err != nil {
		return nil, 0, err
	}
	return data, time.Since(now), nil
}

func (r *Rest) processEndPoints(endpointFunc func(e *endPoint) ([]gjson.Result, time.Duration, error)) (uint64, time.Duration) {
	var (
		err       error
		count     uint64
		totalAPID time.Duration
	)

	for _, endpoint := range r.endpoints {
		var (
			records []gjson.Result
			apiD    time.Duration
		)

		records, apiD, err = endpointFunc(endpoint)
		totalAPID += apiD

		if err != nil {
			r.Logger.Error().Err(err).Str("api", endpoint.prop.Query).Send()
			continue
		}

		if len(records) == 0 {
			r.Logger.Debug().Str("APIPath", endpoint.prop.Query).Msg("no instances on cluster")
			continue
		}
		count = r.HandleResults(records, endpoint.prop, true)
	}

	return count, totalAPID
}

func (r *Rest) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Disk":
		return disk.New(abc)
	case "Health":
		return health.New(abc)
	case "NetRoute":
		return netroute.New(abc)
	case "Qtree":
		return qtree.New(abc)
	case "Snapmirror":
		return snapmirror.New(abc)
	case "Volume":
		return volume.New(abc)
	case "VolumeAnalytics":
		return volumeanalytics.New(abc)
	case "SecurityCertificate":
		return securitycertificate.New(abc)
	case "SVM":
		return svm.New(abc)
	case "Sensor":
		return collectors.NewSensor(abc)
	case "Shelf":
		return shelf.New(abc)
	case "SecurityAccount":
		return securityaccount.New(abc)
	case "QosPolicyFixed":
		return qospolicyfixed.New(abc)
	case "QosPolicyAdaptive":
		return qospolicyadaptive.New(abc)
	case "OntapS3Service":
		return ontaps3service.New(abc)
	case "MetroclusterCheck":
		return metroclustercheck.New(abc)
	case "SystemNode":
		return systemnode.New(abc)
	case "Workload":
		return workload.New(abc)
	case "Certificate":
		return certificate.New(abc)
	default:
		r.Logger.Warn().Str("kind", kind).Msg("no rest plugin found ")
	}
	return nil
}

// HandleResults function is used for handling the rest response for parent as well as endpoints calls,
// isEndPoint would be true only for the endpoint call, and it can't create/delete instance.
func (r *Rest) HandleResults(result []gjson.Result, prop *prop, isEndPoint bool) uint64 {
	var (
		err   error
		count uint64
	)

	oldInstances := set.New()
	currentInstances := set.New()
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

		if len(prop.InstanceKeys) != 0 {
			// extract instance key(s)
			for _, k := range prop.InstanceKeys {
				value := instanceData.Get(k)
				if value.Exists() {
					instanceKey += value.String()
				}
			}

			if instanceKey == "" {
				continue
			}
		}

		instance = mat.GetInstance(instanceKey)

		// Used for endpoints as we don't want to create additional instances
		if isEndPoint && instance == nil {
			// Moved to trace as with filter, this log may spam
			continue
		}

		if instance == nil {
			if instance, err = mat.NewInstance(instanceKey); err != nil {
				r.Logger.Error().Err(err).Str("instKey", instanceKey).Msg("Failed to create new missing instance")
				continue
			}
		}

		if currentInstances.Has(instanceKey) {
			r.Logger.Warn().Str("instKey", instanceKey).Msg("This instance is already processed. instKey is not unique")
		} else {
			currentInstances.Add(instanceKey)
		}
		oldInstances.Remove(instanceKey)

		// clear all instance labels as there are some fields which may be missing between polls
		// Don't remove instance labels when endpoints are being processed because endpoints uses parent instance only.
		if !isEndPoint {
			instance.ClearLabels()
		}
		for label, display := range prop.InstanceLabels {
			value := instanceData.Get(label)
			if value.Exists() {
				if value.IsArray() {
					var labelArray []string
					for _, r := range value.Array() {
						labelString := r.String()
						labelArray = append(labelArray, labelString)
					}
					sort.Strings(labelArray)
					instance.SetLabel(display, strings.Join(labelArray, ","))
				} else {
					instance.SetLabel(display, value.String())
				}
				count++
			}
		}

		for _, metric := range prop.Metrics {
			metr, ok := mat.GetMetrics()[metric.Name]
			if !ok {
				if metr, err = mat.NewMetricFloat64(metric.Name, metric.Label); err != nil {
					r.Logger.Error().Err(err).
						Str("name", metric.Name).
						Msg("NewMetricFloat64")
				}
			}
			f := instanceData.Get(metric.Name)
			if f.Exists() {
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

		// for endpoints, we want to remove common keys from metric count
		if isEndPoint {
			count -= uint64(len(prop.InstanceKeys))
		}
	}

	// Used for parent as we don't want to remove instances for endpoints
	if !isEndPoint {
		// remove deleted instances
		for key := range oldInstances.Iter() {
			mat.RemoveInstance(key)
		}
	}

	return count
}

func (r *Rest) GetRestData(href string) ([]gjson.Result, error) {
	r.Logger.Debug().Str("href", href).Send()
	if href == "" {
		return nil, errs.New(errs.ErrConfig, "empty url")
	}

	result, err := rest.Fetch(r.Client, href)
	if err != nil {
		return r.handleError(err)
	}

	return result, nil
}

func (r *Rest) handleError(err error) ([]gjson.Result, error) {
	if errs.IsRestErr(err, errs.MetroClusterNotConfigured) {
		// MetroCluster is not configured, return ErrMetroClusterNotConfigured
		return nil, errors.Join(errs.ErrAPIRequestRejected, errs.New(errs.ErrMetroClusterNotConfigured, err.Error()))
	}
	return nil, fmt.Errorf("failed to fetch data: %w", err)
}

func (r *Rest) CollectAutoSupport(p *collector.Payload) {
	exporterTypes := make([]string, 0, len(r.Exporters))
	for _, exporter := range r.Exporters {
		exporterTypes = append(exporterTypes, exporter.GetClass())
	}

	var counters = make([]string, 0, len(r.Prop.Counters))
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
	md := r.GetMetadata()
	info := collector.InstanceInfo{
		Count:      md.LazyValueInt64("instances", "data"),
		DataPoints: md.LazyValueInt64("metrics", "data"),
		PollTime:   md.LazyValueInt64("poll_time", "data"),
		APITime:    md.LazyValueInt64("api_time", "data"),
		ParseTime:  md.LazyValueInt64("parse_time", "data"),
		PluginTime: md.LazyValueInt64("plugin_time", "data"),
	}

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
		InstanceInfo:  &info,
	})

	if (r.Name == "Rest" && (r.Object == "Volume" || r.Object == "Node")) || r.Name == "Ems" {
		version := r.Client.Cluster().Version
		p.Target.Version = strconv.Itoa(version[0]) + "." + strconv.Itoa(version[1]) + "." + strconv.Itoa(version[2])
		p.Target.Model = "cdot"
		if p.Target.Serial == "" {
			p.Target.Serial = r.Client.Cluster().UUID
		}
		p.Target.ClusterUUID = r.Client.Cluster().UUID

		if r.Object == "Node" || r.Name == "ems" {
			var (
				nodeIDs []collector.ID
				err     error
			)
			nodeIDs, err = r.getNodeUuids()
			if err != nil {
				// log but don't return so the other info below is collected
				r.Logger.Error().
					Err(err).
					Msg("Unable to get nodes.")
			}
			info.Ids = nodeIDs
			p.Nodes = &info

			// Since the serial number is bogus in c-mode
			// use the first node's serial number instead (the nodes were ordered in getNodeUuids())
			if len(nodeIDs) > 0 {
				p.Target.Serial = nodeIDs[0].SerialNumber
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
	)
	query := "api/cluster/nodes"

	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields([]string{"serial_number", "system_id"}).
		ReturnTimeout(r.Prop.ReturnTimeOut).
		Build()

	if records, err = r.GetRestData(href); err != nil {
		return nil, err
	}

	infos := make([]collector.ID, 0, len(records))
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
