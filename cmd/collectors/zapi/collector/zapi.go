/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package zapi

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/aggregate"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/certificate"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/qospolicyadaptive"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/qospolicyfixed"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/qtree"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/security"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/shelf"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/snapmirror"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/snapshotpolicy"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/svm"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/systemnode"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/workload"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	client "github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slice"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

const BatchSize = "500"

type Zapi struct {
	*collector.AbstractCollector
	Client             *client.Client
	object             string
	Query              string
	TemplatePath       string
	batchSize          string
	desiredAttributes  *node.Node
	instanceKeyPaths   [][]string
	instanceLabelPaths map[string]string
	shortestPathPrefix []string
}

func init() {
	plugin.RegisterModule(&Zapi{})
}

func (z *Zapi) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.zapi",
		New: func() plugin.Module { return new(Zapi) },
	}
}

func (z *Zapi) Init(a *collector.AbstractCollector) error {
	z.AbstractCollector = a
	if err := z.InitVars(); err != nil {
		return err
	}
	// Invoke generic initializer
	// this will load Schedule, initialize Data and Metadata
	if err := collector.Init(z); err != nil {
		return err
	}

	if err := z.InitMatrix(); err != nil {
		return err
	}

	if err := z.InitCache(); err != nil {
		return err
	}

	z.Logger.Debug("initialized")
	return nil
}

func (z *Zapi) InitVars() error {
	jitter := z.Params.GetChildContentS("jitter")
	// It's used for unit tests only
	if z.Options.IsTest {
		z.Client = client.NewTestClient()
		templateName := z.Params.GetChildS("objects").GetChildContentS(z.Object)
		template, path, err := z.ImportSubTemplate("cdot", templateName, jitter, "9.8.0")
		if err != nil {
			return err
		}
		z.TemplatePath = path
		z.Params.Union(template)
		return nil
	}

	var err error
	if z.Client, err = client.New(conf.ZapiPoller(z.Params), z.Auth); err != nil { // convert to connection error, so poller aborts
		return errs.New(errs.ErrConnection, err.Error())
	}
	z.Client.TraceLogSet(z.Name, z.Params)

	if err = z.Client.Init(5, z.Remote); err != nil { // 5 retries before giving up to connect
		return errs.New(errs.ErrConnection, err.Error())
	}
	z.Logger.Debug("connected", slog.String("client", z.Client.Info()))

	model := "cdot"
	if !z.Client.IsClustered() {
		model = "7mode"
	}

	templateName := z.Params.GetChildS("objects").GetChildContentS(z.Object)
	template, path, err := z.ImportSubTemplate(model, templateName, jitter, z.Remote.Version)
	if err != nil {
		return err
	}

	z.TemplatePath = path

	z.Params.Union(template)

	// object name from subtemplate
	if z.object = z.Params.GetChildContentS("object"); z.object == "" {
		return errs.New(errs.ErrMissingParam, "object")
	}

	// api query literal
	if z.Query = z.Params.GetChildContentS("query"); z.Query == "" {
		return errs.New(errs.ErrMissingParam, "query")
	}

	// if the object template includes a client_timeout, use it
	if timeout := z.Params.GetChildContentS("client_timeout"); timeout != "" {
		z.Client.SetTimeout(timeout)
	}
	return nil
}

func (z *Zapi) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Snapmirror":
		return snapmirror.New(abc)
	case "SnapshotPolicy":
		return snapshotpolicy.New(abc)
	case "Shelf":
		return shelf.New(abc)
	case "Qtree":
		return qtree.New(abc)
	case "Volume":
		return volume.New(abc)
	case "LIF":
		return collectors.NewLif(abc)
	case "Sensor":
		return collectors.NewSensor(abc)
	case "Certificate":
		return certificate.New(abc)
	case "SVM":
		return svm.New(abc)
	case "Security":
		return security.New(abc)
	case "QosPolicyFixed":
		return qospolicyfixed.New(abc)
	case "QosPolicyAdaptive":
		return qospolicyadaptive.New(abc)
	case "Aggregate":
		return aggregate.New(abc)
	case "SystemNode":
		return systemnode.New(abc)
	case "Workload":
		return workload.New(abc)
	default:
		z.Logger.Info("no zapi plugin found", slog.String("kind", kind))
	}
	return nil
}

func (z *Zapi) InitCache() error {

	if z.Client.IsClustered() {
		if b := z.Params.GetChildContentS("batch_size"); b != "" {
			if _, err := strconv.Atoi(b); err == nil {
				z.batchSize = b
			}
		}
		if z.batchSize == "" && z.Params.GetChildContentS("no_max_records") != "true" {
			z.batchSize = BatchSize
		}
	}

	z.instanceLabelPaths = make(map[string]string)

	counters := z.Params.GetChildS("counters")
	if counters == nil {
		return errs.New(errs.ErrMissingParam, "counters")
	}

	var ok bool

	z.Logger.Debug("Parsing counters", slog.Int("counters", len(counters.GetChildren())))

	if ok, z.desiredAttributes = z.LoadCounters(counters); !ok {
		if z.Params.GetChildContentS("collect_only_labels") != "true" {
			return errs.New(errs.ErrNoMetric, "failed to parse any")
		}
	}

	z.Logger.Debug(
		"initialized cache",
		slog.Int("metrics", len(z.Matrix[z.GetObject()].GetInstances())),
		slog.Int("labels", len(z.instanceLabelPaths)),
	)

	// unless cluster is the only instance, require instance keys
	if len(z.instanceKeyPaths) == 0 && z.Params.GetChildContentS("only_cluster_instance") != "true" {
		return errs.New(errs.ErrMissingParam, "no instance keys indicated")
	}

	// @TODO validate
	z.shortestPathPrefix = ParseShortestPath(z.Matrix[z.GetObject()], z.instanceLabelPaths)
	z.Logger.Debug("Parsed Instance Keys", slog.Any("keys", z.instanceKeyPaths))
	z.Logger.Debug("Parsed Instance Key Prefix", slog.Any("keys", z.shortestPathPrefix))
	return nil

}

func (z *Zapi) InitMatrix() error {
	mat := z.Matrix[z.GetObject()]
	mat.Object = z.object

	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", z.Client.Name())
	if z.Params.HasChildS("labels") {
		for _, l := range z.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}
	// For 7mode cluster is same as node
	if !z.Client.IsClustered() {
		mat.SetGlobalLabel("node", z.Client.Name())
	}
	return nil
}

func (z *Zapi) PollInstance() (map[string]*matrix.Matrix, error) {
	// for backward compatibility
	return nil, nil
}

func (z *Zapi) PollData() (map[string]*matrix.Matrix, error) {
	var (
		request, response *node.Node
		count, skipped    uint64
		tag               string
		err               error
		apiD, parseD      time.Duration // Request/API time, Parse time, Fetch time
		ad, pd            time.Duration // Request/API time, Parse time, Fetch time
		fetch             func(*matrix.Instance, *node.Node, []string, bool)
		instances         []*node.Node
	)

	oldInstances := set.New()
	mat := z.Matrix[z.Object]
	z.Client.Metadata.Reset()

	// copy keys of current instances. This is used to remove deleted instances from matrix later
	for key := range mat.GetInstances() {
		oldInstances.Add(key)
	}

	fetch = func(instance *matrix.Instance, node *node.Node, path []string, isAppend bool) {
		newPath := path
		newPath = append(newPath, node.GetNameS())
		key := strings.Join(newPath, ".")

		if value := node.GetContentS(); value != "" {
			if label, has := z.instanceLabelPaths[key]; has {
				// Handling array with comma separated values
				previousValue := instance.GetLabel(label)
				if isAppend && previousValue != "" {
					currentVal := strings.Split(previousValue+","+value, ",")
					sort.Strings(currentVal)
					instance.SetLabel(label, strings.Join(currentVal, ","))
				} else {
					instance.SetLabel(label, value)
				}
				count++
			} else if metric := mat.GetMetric(key); metric != nil {
				if err := metric.SetValueString(instance, value); err != nil {
					z.Logger.Error(
						"failed to set value",
						slogx.Err(err),
						slog.String("key", key),
						slog.String("value", value),
					)
					skipped++
				} else {
					count++
				}
			} else {
				skipped++
			}
		} else {
			skipped++
		}

		for _, child := range node.GetChildren() {
			if slice.HasDuplicates(child.GetAllChildNamesS()) {
				fetch(instance, child, newPath, true)
			} else {
				fetch(instance, child, newPath, isAppend)
			}
		}
	}

	mat.Reset()

	request = node.NewXMLS(z.Query)
	if z.Client.IsClustered() {

		if z.Params.GetChildContentS("no_desired_attributes") != "true" {
			request.AddChild(z.desiredAttributes)
		}
		if z.batchSize != "" {
			request.NewChildS("max-records", z.batchSize)
		}
		// special check for snapmirror as we would pass extra input "expand=true"
		if z.Query == "snapmirror-get-iter" {
			request.NewChildS("expand", "true")
		}
	}

	tag = "initial"

	for {
		response, tag, ad, pd, err = z.Client.InvokeBatchWithTimers(request, tag)

		if err != nil {
			return nil, err
		}

		if response == nil {
			break
		}

		apiD += ad
		parseD += pd

		instances = response.SearchChildren(z.shortestPathPrefix)

		if len(instances) == 0 {
			break
		}

		if z.Params.GetChildContentS("only_cluster_instance") == "true" {
			instance := mat.GetInstance("cluster")
			if instance == nil {
				if instance, err = mat.NewInstance("cluster"); err != nil {
					return nil, err
				}
			}
			fetch(instance, instances[0], make([]string, 0), false)
			oldInstances.Remove("cluster")
			break
		}

		for _, instanceElem := range instances {
			keys, found := instanceElem.SearchContent(z.shortestPathPrefix, z.instanceKeyPaths)

			if !found {
				continue
			}

			key := strings.Join(keys, ".")
			instance := mat.GetInstance(key)

			if instance == nil {
				if instance, err = mat.NewInstance(key); err != nil {
					z.Logger.Error(
						"Failed to create new missing instance",
						slogx.Err(err),
						slog.String("instKey", key),
					)
					continue
				}
			}
			instance.SetExportable(true)
			oldInstances.Remove(key)
			// clear all instance labels as there are some fields which may be missing between polls
			instance.ClearLabels()
			fetch(instance, instanceElem, make([]string, 0), false)
		}
	}

	// remove deleted instances
	for key := range oldInstances.Iter() {
		mat.RemoveInstance(key)
		z.Logger.Debug("removed instance", slog.String("key", key))
	}

	numInstances := len(mat.GetInstances())
	// update metadata
	_ = z.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = z.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = z.Metadata.LazySetValueUint64("metrics", "data", count)
	_ = z.Metadata.LazySetValueUint64("instances", "data", uint64(numInstances))
	_ = z.Metadata.LazySetValueUint64("bytesRx", "data", z.Client.Metadata.BytesRx)
	_ = z.Metadata.LazySetValueUint64("numCalls", "data", z.Client.Metadata.NumCalls)

	z.AddCollectCount(count)

	if numInstances == 0 {
		return nil, errs.New(errs.ErrNoInstance, "")
	}

	return z.Matrix, nil
}

func (z *Zapi) CollectAutoSupport(p *collector.Payload) {
	exporterTypes := make([]string, 0, len(z.Exporters))
	for _, exporter := range z.Exporters {
		exporterTypes = append(exporterTypes, exporter.GetClass())
	}

	var counters = make([]string, 0)
	if z.Params != nil {
		c := z.Params.GetChildS("counters")
		c.FlatList(&counters, "")
	}
	slices.Sort(counters)

	var schedules = make([]collector.Schedule, 0)
	tasks := z.Params.GetChildS("schedule")
	if tasks != nil && len(tasks.GetChildren()) > 0 {
		for _, task := range tasks.GetChildren() {
			schedules = append(schedules, collector.Schedule{
				Name:     task.GetNameS(),
				Schedule: task.GetContentS(),
			})
		}
	}

	clientTimeout := client.DefaultTimeout
	newTimeout := z.Params.GetChildContentS("client_timeout")
	if newTimeout != "" {
		clientTimeout = newTimeout
	}

	// Add collector information
	md := z.GetMetadata()
	info := collector.InstanceInfo{
		Count:      md.LazyValueInt64("instances", "data"),
		DataPoints: md.LazyValueInt64("metrics", "data"),
		PollTime:   md.LazyValueInt64("poll_time", "data"),
		APITime:    md.LazyValueInt64("api_time", "data"),
		ParseTime:  md.LazyValueInt64("parse_time", "data"),
		PluginTime: md.LazyValueInt64("plugin_time", "data"),
	}

	p.AddCollectorAsup(collector.AsupCollector{
		Name:      z.Name,
		Query:     z.Query,
		BatchSize: z.batchSize,
		Exporters: exporterTypes,
		Counters: collector.Counters{
			Count: len(counters),
			List:  counters,
		},
		Schedules:     schedules,
		ClientTimeout: clientTimeout,
		InstanceInfo:  &info,
	})

	if z.Name == "Zapi" && (z.Object == "Volume" || z.Object == "Node" || z.Object == "Qtree") {
		p.Target.Version = z.Remote.Version
		p.Target.Model = z.Remote.Model
		if p.Target.Serial == "" {
			p.Target.Serial = z.Remote.UUID
		}
		p.Target.ClusterUUID = z.Client.ClusterUUID()

		switch z.Object {
		case "Node":
			var (
				nodeIDs []collector.ID
				err     error
			)
			nodeIDs, err = z.getNodeUuids()
			if err != nil {
				// log but don't return so the other info below is collected
				z.Logger.Error("Unable to get nodes", slogx.Err(err))
			}
			info.Ids = nodeIDs
			p.Nodes = &info
			if z.Client.IsClustered() {
				// Since the serial number is bogus in c-mode
				// use the first node's serial number instead (the nodes were ordered in getNodeUuids())
				if len(nodeIDs) > 0 {
					p.Target.Serial = nodeIDs[0].SerialNumber
				}
			}
		case "Volume":
			p.Volumes = &info
		case "Qtree":
			info = collector.InstanceInfo{
				PluginInstances: md.LazyValueInt64("pluginInstances", "data"),
			}
			p.Quotas = &info
		}
	}
}

func (z *Zapi) getNodeUuids() ([]collector.ID, error) {
	var (
		response *node.Node
		nodes    []*node.Node
		err      error
	)

	// Since 7-mode is like single node, return the ids for it
	if !z.Client.IsClustered() {
		return []collector.ID{{
			SerialNumber: z.Client.Serial(),
			SystemID:     z.Client.ClusterUUID(),
		}}, nil
	}
	request := "system-node-get-iter"

	if response, err = z.Client.InvokeRequestString(request); err != nil {
		return nil, fmt.Errorf("failure invoking zapi: %s %w", request, err)
	}

	if attrs := response.GetChildS("attributes-list"); attrs != nil {
		nodes = attrs.GetChildren()
	}

	infos := make([]collector.ID, 0, len(nodes))
	for _, n := range nodes {
		sn := n.GetChildContentS("node-serial-number")
		systemID := n.GetChildContentS("node-system-id")
		infos = append(infos, collector.ID{SerialNumber: sn, SystemID: systemID})
	}
	// When Harvest monitors a c-mode system, the first node is picked.
	// Sort so there's a higher chance the same node is picked each time this method is called
	sort.SliceStable(infos, func(i, j int) bool {
		return infos[i].SerialNumber < infos[j].SerialNumber
	})
	return infos, nil
}

// Interface guards
var (
	_ collector.Collector = (*Zapi)(nil)
)
