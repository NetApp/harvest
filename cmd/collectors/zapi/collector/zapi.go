/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package zapi

import (
	"strconv"
	"strings"
	"time"

	"goharvest2/cmd/poller/collector"
	"goharvest2/pkg/color"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logger"
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

func New(a *collector.AbstractCollector) collector.Collector {
	return &Zapi{AbstractCollector: a}
}

func NewZapi(a *collector.AbstractCollector) *Zapi {
	return &Zapi{AbstractCollector: a}
}

func (me *Zapi) Init() error {

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

	logger.Debug(me.Prefix, "initialized")
	return nil
}

func (me *Zapi) InitVars() error {

	var err error

	if me.Client, err = client.New(me.Params); err != nil { // convert to connection error, so poller aborts
		return errors.New(errors.ERR_CONNECTION, err.Error())
	}

	if me.Client.Init(5); err != nil { // 5 retries before giving up to connect
		return errors.New(errors.ERR_CONNECTION, err.Error())
	}
	logger.Debug(me.Prefix, "connected to: %s", me.Client.Info())

	me.TemplateFn = me.Params.GetChildS("objects").GetChildContentS(me.Object) // @TODO err handling

	model := "cdot"
	if !me.Client.IsClustered() {
		model = "7mode"
	}

	template, err := me.ImportSubTemplate(model, me.TemplateFn, me.Client.Version())
	if err != nil {
		logger.Error(me.Prefix, "Error importing subtemplate: %s", err)
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
	return nil
}

func (me *Zapi) InitCache() error {

	if me.Client.IsClustered() {
		if b := me.Params.GetChildContentS("batch_size"); b != "" {
			if _, err := strconv.Atoi(b); err == nil {
				logger.Trace(me.Prefix, "using batch-size [%s]", me.batchSize)
				me.batchSize = b
			}
		}
		if me.batchSize == "" && me.Params.GetChildContentS("no_max_records") != "true" {
			logger.Trace(me.Prefix, "using default batch-size [%s]", BatchSize)
			me.batchSize = BatchSize
		} else {
			logger.Trace(me.Prefix, "using default no batch-size")
		}
	}

	me.instanceLabelPaths = make(map[string]string)

	counters := me.Params.GetChildS("counters")
	if counters == nil {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	var ok bool

	logger.Debug(me.Prefix, "Parsing counters: %d values", len(counters.GetChildren()))

	if ok, me.desiredAttributes = me.LoadCounters(counters); !ok {
		if me.Params.GetChildContentS("collect_only_labels") != "true" {
			return errors.New(errors.ERR_NO_METRIC, "failed to parse any")
		}
	}

	logger.Debug(me.Prefix, "initialized cache with %d metrics and %d labels", len(me.Matrix.GetInstances()), len(me.instanceLabelPaths))

	if len(me.instanceKeyPaths) == 0 && me.Params.GetChildContentS("only_cluster_instance") != "true" {
		return errors.New(errors.INVALID_PARAM, "no instance keys indicated")
	}

	// @TODO validate
	me.shortestPathPrefix = ParseShortestPath(me.Matrix, me.instanceLabelPaths)
	logger.Debug(me.Prefix, "Parsed Instance Keys: %v", me.instanceKeyPaths)
	logger.Debug(me.Prefix, "Parsed Instance Key Prefix: %v", me.shortestPathPrefix)
	return nil

}

func (me *Zapi) InitMatrix() error {
	// overwrite from abstract collector
	//me.Matrix.Collector = me.Matrix.Collector + ":" + me.Matrix.Object
	me.Matrix.Object = me.object

	// Add system (cluster) name
	me.Matrix.SetGlobalLabel("cluster", me.Client.Name())
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

	logger.Debug(me.Prefix, "starting instance poll")

	oldCount = uint64(len(me.Matrix.GetInstances()))
	me.Matrix.PurgeInstances()

	count = 0

	// special case when there is only once instance
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

			logger.Debug(me.Prefix, "fetching %d instances", len(instances))

			for _, instance := range instances {
				//c.logger.Printf(c.Prefix, "Handling instance element <%v> [%s]", &instance, instance.GetName())
				keys, found = instance.SearchContent(me.shortestPathPrefix, me.instanceKeyPaths)

				logger.Debug(me.Prefix, "keys=%v keypaths=%v found=%v", keys, me.instanceKeyPaths, found)
				logger.Debug(me.Prefix, "fetched instance keys (%v): %v", me.instanceKeyPaths, keys)

				if !found {
					logger.Debug(me.Prefix, "skipping element, no instance keys found")
				} else {
					if _, err = me.Matrix.NewInstance(strings.Join(keys, ".")); err != nil {
						logger.Error(me.Prefix, err.Error())
					} else {
						logger.Debug(me.Prefix, "added instance [%s]", strings.Join(keys, "."))
						count += 1
					}
				}
			}
		}
	}

	err = me.Metadata.LazySetValueUint64("count", "instance", count)
	if err != nil {
		logger.Error(me.Prefix, "error: %v", err)
	}
	logger.Debug(me.Prefix, "added %d instances to cache (old cache had %d)", count, oldCount)

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
		logger.Debug(me.Prefix, " > %s(%s)%s <%s%d%s> name=[%s%s%s%s] value=[%s%s%s]", color.Grey, newpath, color.End, color.Red, len(node.GetChildren()), color.End, color.Bold, color.Cyan, node.GetNameS(), color.End, color.Yellow, node.GetContentS(), color.End)

		if value := node.GetContentS(); value != "" {
			key := strings.Join(newpath, ".")
			if metric := me.Matrix.GetMetric(key); metric != nil {
				if err := metric.SetValueString(instance, value); err != nil {
					//logger.Warn(me.Prefix, "%sskipped metric (%s) set value (%s): %v%s", color.Red, key, value, err, color.End)
					skipped += 1
				} else {
					//logger.Trace(me.Prefix, "%smetric (%s) set value (%s)%s", color.Green, key, value, color.End)
					count += 1
				}
			} else if label, has := me.instanceLabelPaths[key]; has {
				instance.SetLabel(label, value)
				//logger.Trace(me.Prefix, "%slabel (%s) [%s] set value (%s)%s", color.Yellow, key, label, value, color.End)
				count += 1
			} else {
				logger.Debug(me.Prefix, "%sskipped (%s) with value (%s): not in metric or label cache%s", color.Blue, key, value, color.End)
				skipped += 1
			}
		} else {
			//logger.Trace(me.Prefix, "%sskippped (%s) with no value%s", color.Cyan, key, color.End)
			skipped += 1
		}

		for _, child := range node.GetChildren() {
			fetch(instance, child, newpath)
		}
	}

	logger.Debug(me.Prefix, "starting data poll")

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

		logger.Debug(me.Prefix, "fetched %d instance elements", len(instances))

		if me.Params.GetChildContentS("only_cluster_instance") == "true" {
			if instance := me.Matrix.GetInstance("cluster"); instance != nil {
				fetch(instance, instances[0], make([]string, 0))
			} else {
				logger.Error(me.Prefix, "cluster instance not found in cache")
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
				logger.Error(me.Prefix, "skipped instance [%s]: not found in cache", strings.Join(keys, "."))
				continue
			}
			fetch(instance, instanceElem, make([]string, 0))
		}
	}

	// update metadata
	me.Metadata.LazySetValueInt64("api_time", "data", apiT.Microseconds())
	me.Metadata.LazySetValueInt64("parse_time", "data", parseT.Microseconds())
	me.Metadata.LazySetValueUint64("count", "data", count)
	me.AddCollectCount(count)

	return me.Matrix, nil
}
