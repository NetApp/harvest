package rest

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/aggregate"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/auditlog"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/certificate"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/cifssession"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/cluster"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/clusterschedule"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/clustersoftware"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/disk"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/health"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/mav"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/metroclustercheck"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/netroute"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/ontaps3service"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/qospolicyadaptive"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/qospolicyfixed"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/quota"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/securityaccount"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/shelf"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/snapmirror"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/snapshotpolicy"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/svm"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/systemnode"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/tag"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/tagmapper"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/volumeanalytics"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/vscanpool"
	"github.com/netapp/harvest/v2/cmd/collectors/rest/plugins/workload"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/auth"
	collector2 "github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/version"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"sort"
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
	Endpoints                    []*EndPoint
	isIgnoreUnknownFieldsEnabled bool
	BatchSize                    string
	AllowPartialAggregation      bool
}

type EndPoint struct {
	Prop        *prop
	name        string
	instanceAdd bool
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

func (r *Rest) query(p *EndPoint) string {
	return p.Prop.Query
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
	return fields
}

func (r *Rest) filter(p *EndPoint) []string {
	return p.Prop.Filter
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

	if err := r.InitEndPoints(); err != nil {
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

	r.Logger.Debug(
		"initialized cache",
		slog.Int("numMetrics", len(r.Prop.Metrics)),
		slog.String("timeout", r.Client.Timeout.String()),
	)

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
		r.Logger.Info("Using default timeout", slog.String("timeout", rest.DefaultTimeout))
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

	if err := r.Client.Init(5, r.Remote); err != nil {
		return err
	}

	return nil
}

func (r *Rest) InitMatrix() error {
	mat := r.Matrix[r.Object]
	// overwrite from abstract collector
	mat.Object = r.Prop.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", r.Remote.Name)

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
		r.Logger.Error("", slogx.Err(err), slog.String("poller", opt.Poller))
		return nil, err
	}
	if poller.Addr == "" {
		r.Logger.Error("Address is empty", slog.String("poller", opt.Poller))
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}
	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if a.Options.IsTest {
		return &rest.Client{Metadata: &collector2.Metadata{}}, nil
	}
	if client, err = rest.New(poller, timeout, c); err != nil {
		r.Logger.Error("error creating new client", slogx.Err(err), slog.String("poller", opt.Poller))
		os.Exit(1)
	}

	return client, err
}

func (r *Rest) InitEndPoints() error {

	endpoints := r.Params.GetChildS("endpoints")
	if endpoints != nil {
		for _, line := range endpoints.GetChildren() {

			n := line.GetNameS()
			e := EndPoint{name: n}

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
					p.IsPublic = requests.IsPublicAPI(p.Query)
				}
				if line1.GetNameS() == "instance_add" {
					iAdd := line1.GetContentS()
					if iAdd == "true" {
						e.instanceAdd = true
					}
				}
				if line1.GetNameS() == "counters" {
					r.ParseRestCounters(line1, &p)
				}
			}
			e.Prop = &p
			r.Endpoints = append(r.Endpoints, &e)
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

// PollCounter performs daily tasks such as updating the cluster info and caching href.
func (r *Rest) PollCounter() (map[string]*matrix.Matrix, error) {

	startTime := time.Now()
	// Update the cluster info to track if ONTAP version is updated
	err := r.Client.UpdateClusterInfo(5)
	if err != nil {
		return nil, err
	}
	apiD := time.Since(startTime)

	startTime = time.Now()
	v, err := version.AtLeast(r.Remote.Version, "9.11.1")
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
		HiddenFields(r.Prop.HiddenFields).
		Filter(r.Prop.Filter).
		MaxRecords(r.BatchSize).
		ReturnTimeout(r.Prop.ReturnTimeOut).
		IsIgnoreUnknownFieldsEnabled(r.isIgnoreUnknownFieldsEnabled).
		Build()

	for _, e := range r.Endpoints {
		e.Prop.Href = rest.NewHrefBuilder().
			APIPath(r.query(e)).
			Fields(r.Fields(e.Prop)).
			HiddenFields(e.Prop.HiddenFields).
			Filter(r.filter(e)).
			MaxRecords(r.BatchSize).
			ReturnTimeout(r.Prop.ReturnTimeOut).
			IsIgnoreUnknownFieldsEnabled(r.isIgnoreUnknownFieldsEnabled).
			Build()
	}
}

func (r *Rest) PollData() (map[string]*matrix.Matrix, error) {
	var (
		apiD, parseD time.Duration
		metricCount  uint64
	)
	r.Matrix[r.Object].Reset()
	r.Client.Metadata.Reset()

	// Track old instances before processing batches
	oldInstances := set.New()
	for key := range r.Matrix[r.Object].GetInstances() {
		oldInstances.Add(key)
	}

	processBatch := func(records []gjson.Result, _ int64) error {
		if len(records) == 0 {
			return nil
		}

		// Process the current batch of records
		count, batchParseD := r.pollData(records, oldInstances)
		metricCount += count
		parseD += batchParseD
		apiD -= batchParseD
		return nil
	}

	startTime := time.Now()
	if err := rest.FetchAllStream(r.Client, r.Prop.Href, processBatch); err != nil {
		_, err2 := r.handleError(err)
		return nil, err2
	}
	apiD += time.Since(startTime)

	// Process endpoints after all batches have been processed
	eCount, endpointAPID := r.ProcessEndPoints(r.Matrix[r.Object], r.ProcessEndPoint, oldInstances)
	metricCount += eCount
	apiD += endpointAPID

	r.postPollData(apiD, parseD, metricCount, oldInstances)
	return r.Matrix, nil
}

func (r *Rest) postPollData(apiD time.Duration, parseD time.Duration, metricCount uint64, oldInstances *set.Set) {
	// Remove old instances that are not found in new instances
	for key := range oldInstances.Iter() {
		r.Matrix[r.Object].RemoveInstance(key)
	}

	numRecords := len(r.Matrix[r.Object].GetInstances())

	_ = r.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = r.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = r.Metadata.LazySetValueUint64("metrics", "data", metricCount)
	_ = r.Metadata.LazySetValueUint64("instances", "data", uint64(numRecords))
	_ = r.Metadata.LazySetValueUint64("bytesRx", "data", r.Client.Metadata.BytesRx)
	_ = r.Metadata.LazySetValueUint64("numCalls", "data", r.Client.Metadata.NumCalls)

	r.AddCollectCount(metricCount)
}

func (r *Rest) pollData(records []gjson.Result, oldInstances *set.Set) (uint64, time.Duration) {

	var (
		count  uint64
		parseD time.Duration
	)

	startTime := time.Now()
	mat := r.Matrix[r.Object]

	count, _ = r.HandleResults(mat, records, r.Prop, false, oldInstances, time.Now().UnixNano()/collector2.BILLION)
	parseD = time.Since(startTime)

	return count, parseD
}

func (r *Rest) ProcessEndPoint(e *EndPoint) ([]gjson.Result, time.Duration, error) {
	now := time.Now()
	data, err := r.GetRestData(e.Prop.Href)
	if err != nil {
		return nil, 0, err
	}
	return data, time.Since(now), nil
}

func (r *Rest) ProcessEndPoints(mat *matrix.Matrix, endpointFunc func(e *EndPoint) ([]gjson.Result, time.Duration, error), oldInstances *set.Set) (uint64, time.Duration) {
	var (
		err       error
		count     uint64
		totalAPID time.Duration
	)

	for _, endpoint := range r.Endpoints {
		var (
			records []gjson.Result
			apiD    time.Duration
		)

		records, apiD, err = endpointFunc(endpoint)
		totalAPID += apiD

		if err != nil {
			r.Logger.Error("", slogx.Err(err), slog.String("api", endpoint.Prop.Query))
			continue
		}

		if len(records) == 0 {
			r.Logger.Debug("no instances on cluster", slog.String("APIPath", endpoint.Prop.Query))
			continue
		}
		isImmutable := !endpoint.instanceAdd
		count, _ = r.HandleResults(mat, records, endpoint.Prop, isImmutable, oldInstances, time.Now().UnixNano()/collector2.BILLION)
	}

	return count, totalAPID
}

func (r *Rest) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Aggregate":
		return aggregate.New(abc)
	case "AuditLog":
		return auditlog.New(abc)
	case "CIFSSession":
		return cifssession.New(abc)
	case "Cluster":
		return cluster.New(abc)
	case "ClusterSchedule":
		return clusterschedule.New(abc)
	case "ClusterSoftware":
		return clustersoftware.New(abc)
	case "Disk":
		return disk.New(abc)
	case "Health":
		return health.New(abc)
	case "LIF":
		return collectors.NewLif(abc)
	case "NetRoute":
		return netroute.New(abc)
	case "MAV":
		return mav.New(abc)
	case "Quota":
		return quota.New(abc)
	case "Snapmirror":
		return snapmirror.New(abc)
	case "Volume":
		return volume.New(abc)
	case "VolumeAnalytics":
		return volumeanalytics.New(abc)
	case "Certificate":
		return certificate.New(abc)
	case "SVM":
		return svm.New(abc)
	case "Sensor":
		return collectors.NewSensor(abc)
	case "Shelf":
		return shelf.New(abc)
	case "SnapshotPolicy":
		return snapshotpolicy.New(abc)
	case "SecurityAccount":
		return securityaccount.New(abc)
	case "Tag":
		return tag.New(abc)
	case "TagMapper":
		return tagmapper.New(abc)
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
	case "VscanPool":
		return vscanpool.New(abc)
	default:
		r.Logger.Warn("no rest plugin found", slog.String("kind", kind))
	}
	return nil
}

// HandleResults function is used for handling the rest response for parent as well as endpoints calls,
// isEndPoint would be true only for the endpoint call, and it can't create/delete instance.
func (r *Rest) HandleResults(mat *matrix.Matrix, result []gjson.Result, prop *prop, isImmutable bool, oldInstances *set.Set, timestamp int64) (uint64, uint64) {
	var (
		err         error
		count       uint64
		numPartials uint64
	)

	currentInstances := set.New()

	for _, instanceData := range result {
		var (
			instanceKey string
			instance    *matrix.Instance
		)

		if !instanceData.IsObject() {
			r.Logger.Warn("Instance data is not object, skipping", slog.String("type", instanceData.Type.String()))
			continue
		}

		if len(prop.InstanceKeys) != 0 {
			// extract instance key(s)
			for _, k := range prop.InstanceKeys {
				value := instanceData.Get(k)
				if value.Exists() {
					instanceKey += value.ClonedString()
				}
			}

			if instanceKey == "" {
				continue
			}
		}

		instance = mat.GetInstance(instanceKey)

		// Used for endpoints as we don't want to create additional instances
		if isImmutable && instance == nil {
			// Moved to trace as with filter, this log may spam
			continue
		}

		if instance == nil {
			if instance, err = mat.NewInstance(instanceKey); err != nil {
				r.Logger.Error("Failed to create new instance", slogx.Err(err), slog.String("instKey", instanceKey))
				continue
			}
		}

		instance.SetExportable(true)

		if currentInstances.Has(instanceKey) {
			r.Logger.Error("This instance is already processed. instKey is not unique", slog.String("instKey", instanceKey))
		} else {
			currentInstances.Add(instanceKey)
		}
		// clear all instance labels as there are some fields which may be missing between polls
		// Don't remove instance labels when endpoints are being processed because endpoints uses parent instance only.
		if !isImmutable {
			oldInstances.Remove(instanceKey)
			instance.ClearLabels()
		}
		for label, display := range prop.InstanceLabels {
			value := instanceData.Get(label)
			if value.Exists() {
				if value.IsArray() {
					var labelArray []string
					for _, r := range value.Array() {
						labelString := r.ClonedString()
						labelArray = append(labelArray, labelString)
					}
					sort.Strings(labelArray)
					instance.SetLabel(display, strings.Join(labelArray, ","))
				} else {
					instance.SetLabel(display, value.ClonedString())
				}
				count++
			}
		}

		// This is relevant for the KeyPerf collector.
		// If the `statistics.status` is not OK, then set `partial` to true.
		if mat.UUID == "KeyPerf" {
			status := instanceData.Get("statistics.status")
			if status.Exists() {
				s := status.ClonedString()
				switch {
				case strings.HasPrefix(s, "partial"):
					// Partial aggregation detected but allowed processing - mark as complete and exportable
					if r.AllowPartialAggregation {
						instance.SetPartial(false)
						instance.SetExportable(true)
					} else {
						// Partial aggregation detected and not allowed processing - mark instance as partial and non-exportable
						instance.SetPartial(true)
						instance.SetExportable(false)
						numPartials++
					}
				case s != "ok":
					// Any non-OK status (excluding partial) - mark as partial and non-exportable
					instance.SetPartial(true)
					instance.SetExportable(false)
					numPartials++
				default:
					// Status is "ok" - mark as complete and exportable
					instance.SetPartial(false)
					instance.SetExportable(true)
				}
			}
		}

		for _, metric := range prop.Metrics {
			metr, ok := mat.GetMetrics()[metric.Name]
			if !ok {
				if metr, err = mat.NewMetricFloat64(metric.Name, metric.Label); err != nil {
					r.Logger.Error(
						"NewMetricFloat64",
						slogx.Err(err),
						slog.String("name", metric.Name),
					)
				} else {
					metr.SetExportable(metric.Exportable)
				}
			}
			f := instanceData.Get(metric.Name)

			if metric.Name == "statistics.timestamp" && mat.UUID == "KeyPerf" {
				sTimestamp := instanceData.Get(metric.Name)
				if !sTimestamp.Exists() {
					metr.SetValueInt64(instance, timestamp)
					count++
					continue
				}
			}

			if f.Exists() {
				var floatValue float64
				switch metric.MetricType {
				case "duration":
					floatValue = collectors.HandleDuration(f.ClonedString())
				case "timestamp":
					floatValue = collectors.HandleTimestamp(f.ClonedString())
				case "":
					floatValue = f.Float()
				default:
					r.Logger.Warn("unknown metric type", slog.String("type", metric.MetricType), slog.String("metric", metric.Name))
				}

				metr.SetValueFloat64(instance, floatValue)
				count++
			}
		}

		// for isImmutable, we want to remove common keys from metric count
		if isImmutable {
			count -= uint64(len(prop.InstanceKeys))
		}
	}

	return count, numPartials
}

func (r *Rest) GetRestData(href string, headers ...map[string]string) ([]gjson.Result, error) {
	r.Logger.Debug("Fetching data", slog.String("href", href))
	if href == "" {
		return nil, errs.New(errs.ErrConfig, "empty url")
	}

	result, err := rest.FetchAll(r.Client, href, headers...)
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
	slices.Sort(counters)

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

	isRest := r.Name == "Rest"
	isKeyPerf := r.Name == "KeyPerf"
	isEMS := r.Name == "Ems"
	isOneOfVolumeNode := r.Object == "Volume" || r.Object == "Node"

	if ((isRest || isKeyPerf) && isOneOfVolumeNode) || isEMS {
		p.Target.Version = r.Remote.Version
		p.Target.Model = "cdot"
		if p.Target.Serial == "" {
			p.Target.Serial = r.Remote.UUID
		}
		p.Target.ClusterUUID = r.Remote.UUID

		if r.Object == "Node" || r.Name == "ems" {
			var (
				nodeIDs []collector.ID
				err     error
			)
			nodeIDs, err = r.getNodeUuids()
			if err != nil {
				// log but don't return so the other info below is collected
				r.Logger.Error("Unable to get nodes", slogx.Err(err))
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
		MaxRecords(collectors.DefaultBatchSize).
		ReturnTimeout(r.Prop.ReturnTimeOut).
		Build()

	if records, err = r.GetRestData(href); err != nil {
		return nil, err
	}

	infos := make([]collector.ID, 0, len(records))
	for _, instanceData := range records {
		infos = append(infos, collector.ID{SerialNumber: instanceData.Get("serial_number").ClonedString(), SystemID: instanceData.Get("system_id").ClonedString()})
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
