/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package zapi

import (
	"fmt"
	"goharvest2/cmd/collectors/zapi/plugins/quota"
	"goharvest2/cmd/collectors/zapi/plugins/shelf"
	"goharvest2/cmd/collectors/zapi/plugins/snapmirror"
	"goharvest2/cmd/collectors/zapi/plugins/volume"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/conf"
	"sort"
	"strconv"
	"strings"
	"time"

	"goharvest2/cmd/poller/collector"
	"goharvest2/pkg/color"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"

	client "goharvest2/pkg/api/ontapi/zapi"
)

const BatchSize = "500"

type Zapi struct {
	*collector.AbstractCollector
	Client             *client.Client
	object             string
	Query              string
	TemplateFn         string
	TemplateType       string
	batchSize          string
	desiredAttributes  *node.Node
	instanceKeyPaths   [][]string
	instanceLabelPaths map[string]string
	shortestPathPrefix []string
}

func init() {
	plugin.RegisterModule(Zapi{})
}

func (Zapi) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.zapi",
		New: func() plugin.Module { return new(Zapi) },
	}
}

func (me *Zapi) Init(a *collector.AbstractCollector) error {
	me.AbstractCollector = a
	if err := me.InitVars(); err != nil {
		return err
	}
	// Invoke generic initializer
	// this will load Schedule, initialize Data and Metadata
	if err := collector.Init(me); err != nil {
		return err
	}

	if err := me.InitMatrix(); err != nil {
		return err
	}

	if err := me.InitCache(); err != nil {
		return err
	}

	me.Logger.Debug().Msg("initialized")
	return nil
}

func (me *Zapi) InitVars() error {

	var err error

	if me.Client, err = client.New(conf.ZapiPoller(me.Params)); err != nil { // convert to connection error, so poller aborts
		return errors.New(errors.ERR_CONNECTION, err.Error())
	}
	me.Client.TraceLogSet(me.Name, me.Params)
	if err = me.Client.Init(5); err != nil { // 5 retries before giving up to connect
		return errors.New(errors.ERR_CONNECTION, err.Error())
	}
	me.Logger.Debug().Msgf("connected to: %s", me.Client.Info())

	me.TemplateFn = me.Params.GetChildS("objects").GetChildContentS(me.Object) // @TODO err handling

	model := "cdot"
	if !me.Client.IsClustered() {
		model = "7mode"
	}

	// save for ASUP messaging
	me.HostUUID = me.Client.Serial()
	version := me.Client.Version()
	me.HostVersion = strconv.Itoa(version[0]) + "." + strconv.Itoa(version[1]) + "." + strconv.Itoa(version[2])
	me.HostModel = model

	template, err := me.ImportSubTemplate(model, me.TemplateFn, me.Client.Version())
	if err != nil {
		me.Logger.Warn().Stack().Err(err).Str("template", me.TemplateFn).Msg("Unable to import template")
		return err
	}

	me.Params.Union(template)

	// object name from subtemplate
	if me.object = me.Params.GetChildContentS("object"); me.object == "" {
		return errors.New(errors.MISSING_PARAM, "object")
	}

	// api query literal
	if me.Query = me.Params.GetChildContentS("query"); me.Query == "" {
		return errors.New(errors.MISSING_PARAM, "query")
	}

	// if the object template includes a client_timeout, use it
	if timeout := me.Params.GetChildContentS("client_timeout"); timeout != "" {
		me.Client.SetTimeout(timeout)
	}
	return nil
}

func (me *Zapi) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Snapmirror":
		return snapmirror.New(abc)
	case "Shelf":
		return shelf.New(abc)
	case "Qtree":
		return quota.New(abc)
	case "Volume":
		return volume.New(abc)
	default:
		me.Logger.Info().Msgf("no zapi plugin found for %s", kind)
	}
	return nil
}

func (me *Zapi) InitCache() error {

	if me.Client.IsClustered() {
		if b := me.Params.GetChildContentS("batch_size"); b != "" {
			if _, err := strconv.Atoi(b); err == nil {
				me.Logger.Trace().Msgf("using batch-size [%s]", me.batchSize)
				me.batchSize = b
			}
		}
		if me.batchSize == "" && me.Params.GetChildContentS("no_max_records") != "true" {
			me.Logger.Trace().Msgf("using default batch-size [%s]", BatchSize)
			me.batchSize = BatchSize
		} else {
			me.Logger.Trace().Msg("using default no batch-size")
		}
	}

	me.instanceLabelPaths = make(map[string]string)

	counters := me.Params.GetChildS("counters")
	if counters == nil {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	var ok bool

	me.Logger.Debug().Msgf("Parsing counters: %d values", len(counters.GetChildren()))

	if ok, me.desiredAttributes = me.LoadCounters(counters); !ok {
		if me.Params.GetChildContentS("collect_only_labels") != "true" {
			return errors.New(errors.ERR_NO_METRIC, "failed to parse any")
		}
	}

	me.Logger.Debug().Msgf("initialized cache with %d metrics and %d labels", len(me.Matrix.GetInstances()), len(me.instanceLabelPaths))

	// unless cluster is the only instance, require instance keys
	if len(me.instanceKeyPaths) == 0 && me.Params.GetChildContentS("only_cluster_instance") != "true" {
		return errors.New(errors.MISSING_PARAM, "no instance keys indicated")
	}

	// @TODO validate
	me.shortestPathPrefix = ParseShortestPath(me.Matrix, me.instanceLabelPaths)
	me.Logger.Debug().Msgf("Parsed Instance Keys: %v", me.instanceKeyPaths)
	me.Logger.Debug().Msgf("Parsed Instance Key Prefix: %v", me.shortestPathPrefix)
	return nil

}

func (me *Zapi) InitMatrix() error {
	// overwrite from abstract collector
	//me.Matrix.Collector = me.Matrix.Collector + ":" + me.Matrix.Object
	me.Matrix.Object = me.object

	// Add system (cluster) name
	me.Matrix.SetGlobalLabel("cluster", me.Client.Name())
	if me.Params.HasChildS("labels") {
		for _, l := range me.Params.GetChildS("labels").GetChildren() {
			me.Matrix.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}
	// For 7mode cluster is same as node
	if !me.Client.IsClustered() {
		me.Matrix.SetGlobalLabel("node", me.Client.Name())
	}
	return nil
}

func (me *Zapi) PollInstance() (*matrix.Matrix, error) {
	var (
		request, response *node.Node
		instances         []*node.Node
		oldCount, count   uint64
		keys              []string
		tag               string
		found             bool
		err               error
	)

	me.Logger.Debug().Msg("starting instance poll")

	oldCount = uint64(len(me.Matrix.GetInstances()))
	me.Matrix.PurgeInstances()

	count = 0

	// special case when only "instance" is the cluster
	if me.Params.GetChildContentS("only_cluster_instance") == "true" {
		if me.Matrix.GetInstance("cluster") == nil {
			if _, err := me.Matrix.NewInstance("cluster"); err != nil {
				return nil, err
			}
		}
		count = 1
	} else {
		request = node.NewXmlS(me.Query)
		if me.Client.IsClustered() && me.batchSize != "" {
			request.NewChildS("max-records", me.batchSize)
		}

		tag = "initial"

		for {

			response, tag, err = me.Client.InvokeBatchRequest(request, tag)

			if err != nil {
				return nil, err
			}

			if response == nil {
				break
			}

			instances = response.SearchChildren(me.shortestPathPrefix)
			if len(instances) == 0 {
				return nil, errors.New(errors.ERR_NO_INSTANCE, "no instances in server response")
			}

			me.Logger.Debug().Msgf("fetching %d instances", len(instances))

			for _, instance := range instances {
				//c.logger.Printf(c.Prefix, "Handling instance element <%v> [%s]", &instance, instance.GetName())
				keys, found = instance.SearchContent(me.shortestPathPrefix, me.instanceKeyPaths)

				me.Logger.Debug().Msgf("keys=%v keypaths=%v found=%v", keys, me.instanceKeyPaths, found)
				me.Logger.Debug().Msgf("fetched instance keys (%v): %v", me.instanceKeyPaths, keys)

				if !found {
					me.Logger.Debug().Msg("skipping element, no instance keys found")
				} else {
					if _, err = me.Matrix.NewInstance(strings.Join(keys, ".")); err != nil {
						me.Logger.Error().Stack().Err(err).Msg("")
					} else {
						me.Logger.Debug().Msgf("added instance [%s]", strings.Join(keys, "."))
						count += 1
					}
				}
			}
		}
	}

	err = me.Metadata.LazySetValueUint64("count", "instance", count)
	if err != nil {
		me.Logger.Error().Stack().Err(err).Msg("error")
	}
	me.Logger.Debug().Msgf("added %d instances to cache (old cache had %d)", count, oldCount)

	if len(me.Matrix.GetInstances()) == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no instances fetched")
	}

	return nil, nil
}

func (me *Zapi) PollData() (*matrix.Matrix, error) {
	var err error
	var request, response *node.Node
	var fetch func(*matrix.Instance, *node.Node, []string)
	var count, skipped uint64
	var ad, pd time.Duration // Request/API time, Parse time, Fetch time
	var tag string

	count = 0
	skipped = 0

	apiT := 0 * time.Second
	parseT := 0 * time.Second

	fetch = func(instance *matrix.Instance, node *node.Node, path []string) {

		newpath := append(path, node.GetNameS())
		key := strings.Join(newpath, ".")
		me.Logger.Debug().Msgf(" > %s(%s)%s <%s%d%s> name=[%s%s%s%s] value=[%s%s%s]", color.Grey, newpath, color.End, color.Red, len(node.GetChildren()), color.End, color.Bold, color.Cyan, node.GetNameS(), color.End, color.Yellow, node.GetContentS(), color.End)
		if value := node.GetContentS(); value != "" {
			if label, has := me.instanceLabelPaths[key]; has {
				instance.SetLabel(label, value)
				me.Logger.Debug().Msgf(" > %slabel (%s) [%s] set value (%s)%s", color.Yellow, key, label, value, color.End)
				count += 1
			} else if metric := me.Matrix.GetMetric(key); metric != nil {
				if err := metric.SetValueString(instance, value); err != nil {
					me.Logger.Error().Msgf("%smetric (%s) set value (%s): %v%s", color.Red, key, value, err, color.End)
					skipped += 1
				} else {
					me.Logger.Debug().Msgf(" > %smetric (%s) set value (%s)%s", color.Green, key, value, color.End)
					count += 1
				}
			} else {
				me.Logger.Debug().Msgf(" > %sskipped (%s) with value (%s): not in metric or label cache%s", color.Blue, key, value, color.End)
				skipped += 1
			}
		} else {
			me.Logger.Debug().Msgf(" > %sskippped (%s) with no value%s", color.Cyan, key, color.End)
			skipped += 1
		}

		for _, child := range node.GetChildren() {
			fetch(instance, child, newpath)
		}
	}

	me.Logger.Debug().Msg("starting data poll")

	me.Matrix.Reset()

	request = node.NewXmlS(me.Query)
	if me.Client.IsClustered() {

		if me.Params.GetChildContentS("no_desired_attributes") != "true" {
			request.AddChild(me.desiredAttributes)
		}
		if me.batchSize != "" {
			request.NewChildS("max-records", me.batchSize)
		}
	}

	tag = "initial"

	for {
		response, tag, ad, pd, err = me.Client.InvokeBatchWithTimers(request, tag)

		if err != nil {
			return nil, err
		}

		if response == nil {
			break
		}

		apiT += ad
		parseT += pd

		instances := response.SearchChildren(me.shortestPathPrefix)

		if len(instances) == 0 {
			return nil, errors.New(errors.ERR_NO_INSTANCE, "")
		}

		me.Logger.Debug().Msgf("fetched %d instance elements", len(instances))

		if me.Params.GetChildContentS("only_cluster_instance") == "true" {
			if instance := me.Matrix.GetInstance("cluster"); instance != nil {
				fetch(instance, instances[0], make([]string, 0))
			} else {
				me.Logger.Error().Stack().Err(nil).Msg("cluster instance not found in cache")
			}
			break
		}

		for _, instanceElem := range instances {
			//c.logger.Printf(c.Prefix, "Handling instance element <%v> [%s]", &instance, instance.GetName())
			keys, found := instanceElem.SearchContent(me.shortestPathPrefix, me.instanceKeyPaths)
			//logger.Debug(me.Prefix, "Fetched instance keys: %s", strings.Join(keys, "."))

			if !found {
				//logger.Debug(me.Prefix, "Skipping instance: no keys fetched")
				continue
			}

			instance := me.Matrix.GetInstance(strings.Join(keys, "."))

			if instance == nil {
				me.Logger.Error().Stack().Err(nil).Msgf("skipped instance [%s]: not found in cache", strings.Join(keys, "."))
				continue
			}
			fetch(instance, instanceElem, make([]string, 0))
		}
	}

	// update metadata
	_ = me.Metadata.LazySetValueInt64("api_time", "data", apiT.Microseconds())
	_ = me.Metadata.LazySetValueInt64("parse_time", "data", parseT.Microseconds())
	_ = me.Metadata.LazySetValueUint64("count", "data", count)
	me.AddCollectCount(count)

	return me.Matrix, nil
}

func (me *Zapi) CollectAutoSupport(p *collector.Payload) {
	var exporterTypes []string
	for _, exporter := range me.Exporters {
		exporterTypes = append(exporterTypes, exporter.GetClass())
	}

	var counters = make([]string, 0)
	if me.Params != nil {
		c := me.Params.GetChildS("counters")
		c.FlatList(&counters, "")
	}

	var schedules = make([]collector.Schedule, 0)
	tasks := me.Params.GetChildS("schedule")
	if tasks != nil && len(tasks.GetChildren()) > 0 {
		for _, task := range tasks.GetChildren() {
			schedules = append(schedules, collector.Schedule{
				Name:     task.GetNameS(),
				Schedule: task.GetContentS(),
			})
		}
	}

	clientTimeout := strconv.Itoa(client.DefaultTimeout)
	newTimeout := me.Params.GetChildContentS("client_timeout")
	if newTimeout != "" {
		clientTimeout = newTimeout
	}

	// Add collector information
	p.AddCollectorAsup(collector.AsupCollector{
		Name:      me.Name,
		Query:     me.Query,
		BatchSize: me.batchSize,
		Exporters: exporterTypes,
		Counters: collector.Counters{
			Count: len(counters),
			List:  counters,
		},
		Schedules:     schedules,
		ClientTimeout: clientTimeout,
	})

	if me.Name == "Zapi" && (me.Object == "Volume" || me.Object == "Node") {
		p.Target.Version = me.GetHostVersion()
		p.Target.Model = me.GetHostModel()
		if p.Target.Serial == "" {
			p.Target.Serial = me.GetHostUUID()
		}
		p.Target.ClusterUuid = me.Client.ClusterUuid()

		md := me.GetMetadata()
		info := collector.InstanceInfo{
			Count:      md.LazyValueInt64("count", "instance"),
			DataPoints: md.LazyValueInt64("count", "data"),
			PollTime:   md.LazyValueInt64("poll_time", "data"),
			ApiTime:    md.LazyValueInt64("api_time", "data"),
			ParseTime:  md.LazyValueInt64("parse_time", "data"),
			PluginTime: md.LazyValueInt64("plugin_time", "data"),
		}

		if me.Object == "Node" {
			nodeIds, err := me.getNodeUuids()
			if err != nil {
				// log but don't return so the other info below is collected
				me.Logger.Error().
					Err(err).
					Msg("Unable to get nodes.")
				nodeIds = make([]collector.Id, 0)
			}
			info.Ids = nodeIds
			p.Nodes = &info
			if me.Client.IsClustered() {
				// Since the serial number is bogus in c-mode
				// use the first node's serial number instead (the nodes were ordered in getNodeUuids())
				if len(nodeIds) > 0 {
					p.Target.Serial = nodeIds[0].SerialNumber
				}
			}
		} else if me.Object == "Volume" {
			p.Volumes = &info
		}
	}
}

func (me *Zapi) getNodeUuids() ([]collector.Id, error) {
	var (
		response *node.Node
		nodes    []*node.Node
		err      error
		infos    []collector.Id
	)

	// Since 7-mode is like single node, return the ids for it
	if !me.Client.IsClustered() {
		return []collector.Id{{
			SerialNumber: me.Client.Serial(),
			SystemId:     me.Client.ClusterUuid(),
		}}, nil
	}
	request := "system-node-get-iter"

	if response, err = me.Client.InvokeRequestString(request); err != nil {
		return nil, fmt.Errorf("failure invoking zapi: %s %w", request, err)
	}

	if attrs := response.GetChildS("attributes-list"); attrs != nil {
		nodes = attrs.GetChildren()
	}

	for _, n := range nodes {
		sn := n.GetChildContentS("node-serial-number")
		systemId := n.GetChildContentS("node-system-id")
		infos = append(infos, collector.Id{SerialNumber: sn, SystemId: systemId})
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
