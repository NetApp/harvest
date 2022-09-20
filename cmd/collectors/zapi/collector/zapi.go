/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package zapi

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/certificate"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/qtree"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/security"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/sensor"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/shelf"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/snapmirror"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/svm"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"

	client "github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
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

	z.Logger.Debug().Msg("initialized")
	return nil
}

func (z *Zapi) InitVars() error {

	var err error

	if z.Client, err = client.New(conf.ZapiPoller(z.Params)); err != nil { // convert to connection error, so poller aborts
		return errs.New(errs.ErrConnection, err.Error())
	}
	z.Client.TraceLogSet(z.Name, z.Params)
	if err = z.Client.Init(5); err != nil { // 5 retries before giving up to connect
		return errs.New(errs.ErrConnection, err.Error())
	}
	z.Logger.Debug().Msgf("connected to: %s", z.Client.Info())

	model := "cdot"
	if !z.Client.IsClustered() {
		model = "7mode"
	}

	// save for ASUP messaging
	z.HostUUID = z.Client.Serial()
	version := z.Client.Version()
	z.HostVersion = strconv.Itoa(version[0]) + "." + strconv.Itoa(version[1]) + "." + strconv.Itoa(version[2])
	z.HostModel = model
	templateName := z.Params.GetChildS("objects").GetChildContentS(z.Object)

	template, templatePath, err := z.ImportSubTemplate(model, templateName, z.Client.Version())
	if err != nil {
		return fmt.Errorf("unable to import template=[%s] %w", templatePath, err)
	}

	z.TemplatePath = templatePath

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
	case "Shelf":
		return shelf.New(abc)
	case "Qtree":
		return qtree.New(abc)
	case "Volume":
		return volume.New(abc)
	case "Sensor":
		return sensor.New(abc)
	case "Certificate":
		return certificate.New(abc)
	case "SVM":
		return svm.New(abc)
	case "Security":
		return security.New(abc)
	default:
		z.Logger.Info().Msgf("no zapi plugin found for %s", kind)
	}
	return nil
}

func (z *Zapi) InitCache() error {

	if z.Client.IsClustered() {
		if b := z.Params.GetChildContentS("batch_size"); b != "" {
			if _, err := strconv.Atoi(b); err == nil {
				z.Logger.Trace().Msgf("using batch-size [%s]", z.batchSize)
				z.batchSize = b
			}
		}
		if z.batchSize == "" && z.Params.GetChildContentS("no_max_records") != "true" {
			z.Logger.Trace().Msgf("using default batch-size [%s]", BatchSize)
			z.batchSize = BatchSize
		} else {
			z.Logger.Trace().Msg("using default no batch-size")
		}
	}

	z.instanceLabelPaths = make(map[string]string)

	counters := z.Params.GetChildS("counters")
	if counters == nil {
		return errs.New(errs.ErrMissingParam, "counters")
	}

	var ok bool

	z.Logger.Debug().Msgf("Parsing counters: %d values", len(counters.GetChildren()))

	if ok, z.desiredAttributes = z.LoadCounters(counters); !ok {
		if z.Params.GetChildContentS("collect_only_labels") != "true" {
			return errs.New(errs.ErrNoMetric, "failed to parse any")
		}
	}

	z.Logger.Debug().Msgf("initialized cache with %d metrics and %d labels", len(z.Matrix[z.GetObject()].GetInstances()), len(z.instanceLabelPaths))

	// unless cluster is the only instance, require instance keys
	if len(z.instanceKeyPaths) == 0 && z.Params.GetChildContentS("only_cluster_instance") != "true" {
		return errs.New(errs.ErrMissingParam, "no instance keys indicated")
	}

	// @TODO validate
	z.shortestPathPrefix = ParseShortestPath(z.Matrix[z.GetObject()], z.instanceLabelPaths)
	z.Logger.Debug().Msgf("Parsed Instance Keys: %v", z.instanceKeyPaths)
	z.Logger.Debug().Msgf("Parsed Instance Key Prefix: %v", z.shortestPathPrefix)
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
	var (
		request, response *node.Node
		instances         []*node.Node
		oldCount, count   uint64
		keys              []string
		tag               string
		found             bool
		err               error
	)

	z.Logger.Debug().Msg("starting instance poll")
	mat := z.Matrix[z.Object]

	if len(z.shortestPathPrefix) == 0 {
		msg := fmt.Sprintf("There is an issue with the template [%s]. It could be due to wrong counter structure.", z.TemplatePath)
		return nil, errs.New(errs.ErrTemplate, msg)
	}

	oldCount = uint64(len(mat.GetInstances()))
	mat.PurgeInstances()

	count = 0

	// special case when only "instance" is the cluster
	if z.Params.GetChildContentS("only_cluster_instance") == "true" {
		if mat.GetInstance("cluster") == nil {
			if _, err := mat.NewInstance("cluster"); err != nil {
				return nil, err
			}
		}
		count = 1
	} else {
		request = node.NewXMLS(z.Query)
		if z.Client.IsClustered() && z.batchSize != "" {
			request.NewChildS("max-records", z.batchSize)
		}

		tag = "initial"

		for {

			response, tag, err = z.Client.InvokeBatchRequest(request, tag)

			if err != nil {
				return nil, err
			}

			if response == nil {
				break
			}

			instances = response.SearchChildren(z.shortestPathPrefix)
			if len(instances) == 0 {
				break
			}

			z.Logger.Debug().Msgf("fetching %d instances", len(instances))

			for _, instance := range instances {
				//c.logger.Printf(c.Prefix, "Handling instance element <%v> [%s]", &instance, instance.GetName())
				keys, found = instance.SearchContent(z.shortestPathPrefix, z.instanceKeyPaths)

				z.Logger.Trace().
					Strs("keys", keys).
					Bool("found", found).
					Msgf("keypaths=%v", z.instanceKeyPaths)

				if !found {
					z.Logger.Debug().Msg("skipping element, no instance keys found")
				} else {
					if _, err = mat.NewInstance(strings.Join(keys, ".")); err != nil {
						z.Logger.Error().Stack().Err(err).Msg("")
					} else {
						z.Logger.Trace().Msgf("added instance [%s]", strings.Join(keys, "."))
						count++
					}
				}
			}
		}
	}

	err = z.Metadata.LazySetValueUint64("count", "instance", count)
	if err != nil {
		z.Logger.Error().Stack().Err(err).Msg("error")
	}
	z.Logger.Debug().Msgf("added %d instances to cache (old cache had %d)", count, oldCount)

	if len(mat.GetInstances()) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no instances fetched")
	}

	return nil, nil
}

func (z *Zapi) PollData() (map[string]*matrix.Matrix, error) {
	var (
		request, response *node.Node
		count, skipped    uint64
		tag               string
		err               error
		ad, pd            time.Duration // Request/API time, Parse time, Fetch time
		fetch             func(*matrix.Instance, *node.Node, []string)
		instances         []*node.Node
	)

	count = 0
	skipped = 0

	apiT := 0 * time.Second
	parseT := 0 * time.Second

	mat := z.Matrix[z.Object]

	fetch = func(instance *matrix.Instance, node *node.Node, path []string) {

		newpath := append(path, node.GetNameS())
		key := strings.Join(newpath, ".")
		z.Logger.Trace().Msgf(" > %s(%s)%s <%s%d%s> name=[%s%s%s%s] value=[%s%s%s]", color.Grey, newpath, color.End, color.Red, len(node.GetChildren()), color.End, color.Bold, color.Cyan, node.GetNameS(), color.End, color.Yellow, node.GetContentS(), color.End)
		if value := node.GetContentS(); value != "" {
			if label, has := z.instanceLabelPaths[key]; has {
				instance.SetLabel(label, value)
				z.Logger.Trace().Msgf(" > %slabel (%s) [%s] set value (%s)%s", color.Yellow, key, label, value, color.End)
				count++
			} else if metric := mat.GetMetric(key); metric != nil {
				if err := metric.SetValueString(instance, value); err != nil {
					z.Logger.Error().Msgf("%smetric (%s) set value (%s): %v%s", color.Red, key, value, err, color.End)
					skipped++
				} else {
					z.Logger.Trace().Msgf(" > %smetric (%s) set value (%s)%s", color.Green, key, value, color.End)
					count++
				}
			} else {
				z.Logger.Trace().Msgf(" > %sskipped (%s) with value (%s): not in metric or label cache%s", color.Blue, key, value, color.End)
				skipped++
			}
		} else {
			z.Logger.Trace().Msgf(" > %sskippped (%s) with no value%s", color.Cyan, key, color.End)
			skipped++
		}

		for _, child := range node.GetChildren() {
			fetch(instance, child, newpath)
		}
	}

	z.Logger.Trace().Msg("starting data poll")

	mat.Reset()

	request = node.NewXMLS(z.Query)
	if z.Client.IsClustered() {

		if z.Params.GetChildContentS("no_desired_attributes") != "true" {
			request.AddChild(z.desiredAttributes)
		}
		if z.batchSize != "" {
			request.NewChildS("max-records", z.batchSize)
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

		apiT += ad
		parseT += pd

		instances = response.SearchChildren(z.shortestPathPrefix)

		if len(instances) == 0 {
			break
		}

		z.Logger.Debug().Msgf("fetched %d instance elements", len(instances))

		if z.Params.GetChildContentS("only_cluster_instance") == "true" {
			if instance := mat.GetInstance("cluster"); instance != nil {
				fetch(instance, instances[0], make([]string, 0))
			} else {
				z.Logger.Error().Stack().Err(nil).Msg("cluster instance not found in cache")
			}
			break
		}

		for _, instanceElem := range instances {
			//c.logger.Printf(c.Prefix, "Handling instance element <%v> [%s]", &instance, instance.GetName())
			keys, found := instanceElem.SearchContent(z.shortestPathPrefix, z.instanceKeyPaths)
			//logger.Debug(z.Prefix, "Fetched instance keys: %s", strings.Join(keys, "."))

			if !found {
				//logger.Debug(z.Prefix, "Skipping instance: no keys fetched")
				continue
			}

			instance := mat.GetInstance(strings.Join(keys, "."))

			if instance == nil {
				z.Logger.Error().Stack().Err(nil).Msgf("skipped instance [%s]: not found in cache", strings.Join(keys, "."))
				continue
			}
			fetch(instance, instanceElem, make([]string, 0))
		}
	}

	z.Logger.Info().
		Int("instances", len(instances)).
		Uint64("metrics", count).
		Str("apiD", apiT.Round(time.Millisecond).String()).
		Str("parseD", parseT.Round(time.Millisecond).String()).
		Msg("Collected")

	// update metadata
	_ = z.Metadata.LazySetValueInt64("api_time", "data", apiT.Microseconds())
	_ = z.Metadata.LazySetValueInt64("parse_time", "data", parseT.Microseconds())
	_ = z.Metadata.LazySetValueUint64("count", "data", count)
	z.AddCollectCount(count)

	if len(mat.GetInstances()) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "")
	}

	return z.Matrix, nil
}

func (z *Zapi) CollectAutoSupport(p *collector.Payload) {
	var exporterTypes []string
	for _, exporter := range z.Exporters {
		exporterTypes = append(exporterTypes, exporter.GetClass())
	}

	var counters = make([]string, 0)
	if z.Params != nil {
		c := z.Params.GetChildS("counters")
		c.FlatList(&counters, "")
	}

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

	clientTimeout := strconv.Itoa(client.DefaultTimeout)
	newTimeout := z.Params.GetChildContentS("client_timeout")
	if newTimeout != "" {
		clientTimeout = newTimeout
	}

	// Add collector information
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
	})

	if z.Name == "Zapi" && (z.Object == "Volume" || z.Object == "Node") {
		p.Target.Version = z.GetHostVersion()
		p.Target.Model = z.GetHostModel()
		if p.Target.Serial == "" {
			p.Target.Serial = z.GetHostUUID()
		}
		p.Target.ClusterUUID = z.Client.ClusterUUID()

		md := z.GetMetadata()
		info := collector.InstanceInfo{
			Count:      md.LazyValueInt64("count", "instance"),
			DataPoints: md.LazyValueInt64("count", "data"),
			PollTime:   md.LazyValueInt64("poll_time", "data"),
			APITime:    md.LazyValueInt64("api_time", "data"),
			ParseTime:  md.LazyValueInt64("parse_time", "data"),
			PluginTime: md.LazyValueInt64("plugin_time", "data"),
		}

		if z.Object == "Node" {
			nodeIds, err := z.getNodeUuids()
			if err != nil {
				// log but don't return so the other info below is collected
				z.Logger.Error().
					Err(err).
					Msg("Unable to get nodes.")
				nodeIds = make([]collector.ID, 0)
			}
			info.Ids = nodeIds
			p.Nodes = &info
			if z.Client.IsClustered() {
				// Since the serial number is bogus in c-mode
				// use the first node's serial number instead (the nodes were ordered in getNodeUuids())
				if len(nodeIds) > 0 {
					p.Target.Serial = nodeIds[0].SerialNumber
				}
			}
		} else if z.Object == "Volume" {
			p.Volumes = &info
		}
	}
}

func (z *Zapi) getNodeUuids() ([]collector.ID, error) {
	var (
		response *node.Node
		nodes    []*node.Node
		err      error
		infos    []collector.ID
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
