package zapi

import (
	//"fmt"

	"path"
	"strconv"
	"strings"
	"time"

	"goharvest2/poller/collector"
	"goharvest2/share/errors"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	//"goharvest2/share/tree"
	"goharvest2/share/tree/node"
	"goharvest2/share/util"

	client "goharvest2/apis/zapi"
)

const (
	BATCH_SIZE = "500"
)

type Zapi struct {
	*collector.AbstractCollector
	Connection   *client.Client
	System       *client.System
	object       string
	Query        string
	TemplateFn   string
	TemplateType string
	batch_size   string
	//data_attrs	      *node.Node
	// @TODO: lowercase, since we don't want to export
	desired_attributes   *node.Node
	INSTANCE_KEY_PATHS   [][]string
	INSTANCE_LABEL_PATHS map[string]string
	SHORTEST_PATH_PREFIX []string
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

	// @TODO check if cert/key files exist
	if me.Params.GetChildContentS("auth_style") == "certificate_auth" {
		if me.Params.GetChildS("ssl_cert") == nil {
			cert_path := path.Join(me.Options.ConfPath, "cert", me.Options.Poller+".pem")
			me.Params.NewChildS("ssl_cert", cert_path)
			logger.Debug(me.Prefix, "added ssl_cert path [%s]", cert_path)
		}

		if me.Params.GetChildS("ssl_key") == nil {
			key_path := path.Join(me.Options.ConfPath, "cert", me.Options.Poller+".key")
			me.Params.NewChildS("ssl_key", key_path)
			logger.Debug(me.Prefix, "added ssl_key path [%s]", key_path)
		}
	}

	if me.Connection, err = client.New(me.Params); err != nil {
		return err
	}

	// @TODO handle connectivity-related errors (retry a few times)
	if me.System, err = me.Connection.GetSystem(); err != nil {
		//logger.Error(c.Prefix, "system info: %v", err)
		return err
	}
	logger.Debug(me.Prefix, "connected to: %s", me.System.String())

	me.TemplateFn = me.Params.GetChildS("objects").GetChildContentS(me.Object) // @TODO err handling

	model := "cdot"
	if !me.System.Clustered {
		model = "7mode"
	}

	template, err := me.ImportSubTemplate(model, "default", me.TemplateFn, me.System.Version)
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

	if me.System.Clustered {
		if b := me.Params.GetChildContentS("batch_size"); b != "" {
			if _, err := strconv.Atoi(b); err == nil {
				logger.Trace(me.Prefix, "using batch-size [%s]", me.batch_size)
				me.batch_size = b
			}
		}
		if me.batch_size == "" {
			logger.Trace(me.Prefix, "using default batch-size [%s]", BATCH_SIZE)
			me.batch_size = BATCH_SIZE
		}
	}

	me.INSTANCE_LABEL_PATHS = make(map[string]string)

	counters := me.Params.GetChildS("counters")
	if counters == nil {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	var ok bool

	logger.Debug(me.Prefix, "Parsing counters: %d values", len(counters.GetChildren()))

	if ok, me.desired_attributes = me.LoadCounters(counters); !ok {
		if me.Params.GetChildContentS("collect_only_labels") != "true" {
			return errors.New(errors.ERR_NO_METRIC, "failed to parse any")
		}
	}

	logger.Debug(me.Prefix, "initialized cache with %d metrics and %d labels", me.Matrix.SizeMetrics(), len(me.INSTANCE_LABEL_PATHS))

	if len(me.INSTANCE_KEY_PATHS) == 0 && me.Params.GetChildContentS("only_cluster_instance") != "true" {
		return errors.New(errors.INVALID_PARAM, "no instance keys indicated")
	}

	// @TODO validate
	me.SHORTEST_PATH_PREFIX = ParseShortestPath(me.Matrix, me.INSTANCE_LABEL_PATHS)
	logger.Debug(me.Prefix, "Parsed Instance Keys: %v", me.INSTANCE_KEY_PATHS)
	logger.Debug(me.Prefix, "Parsed Instance Key Prefix: %v", me.SHORTEST_PATH_PREFIX)
	return nil

}

func (me *Zapi) InitMatrix() error {
	// overwrite from abstract collector
	me.Matrix.Collector = me.Matrix.Collector + ":" + me.Matrix.Object
	me.Matrix.Object = me.object

	// Add system (cluster) name
	me.Matrix.SetGlobalLabel("cluster", me.System.Name)
	if !me.System.Clustered {
		me.Matrix.SetGlobalLabel("node", me.System.Name)
	}
	return nil
}

func (me *Zapi) PollInstance() (*matrix.Matrix, error) {
	var err error
	var request, response *node.Node
	var instances []*node.Node
	var old_count, count uint64
	var keys []string
	var tag string
	var found bool

	logger.Debug(me.Prefix, "starting instance poll")

	old_count = uint64(me.Matrix.SizeInstances())
	me.Matrix.PurgeInstances()

	count = 0

	// special case when there is only once instance
	if me.Params.GetChildContentS("only_cluster_instance") == "true" {
		if _, err := me.Matrix.AddInstance("cluster"); err != nil {
			return nil, err
		}
		count = 1
	} else {
		request = node.NewXmlS(me.Query)
		if me.System.Clustered && me.batch_size != "" {
			request.NewChildS("max-records", me.batch_size)
		}

		tag = "initial"

		for {

			response, tag, err = me.Connection.InvokeBatchRequest(request, tag)

			if err != nil {
				return nil, err
			}

			if response == nil {
				break
			}

			instances = response.SearchChildren(me.SHORTEST_PATH_PREFIX)
			if len(instances) == 0 {
				return nil, errors.New(errors.ERR_NO_INSTANCE, "no instances in server response")
			}

			logger.Debug(me.Prefix, "fetching %d instances", len(instances))

			for _, instance := range instances {
				//c.logger.Printf(c.Prefix, "Handling instance element <%v> [%s]", &instance, instance.GetName())
				keys, found = instance.SearchContent(me.SHORTEST_PATH_PREFIX, me.INSTANCE_KEY_PATHS)

				logger.Debug(me.Prefix, "keys=%v keypaths=%v found=%v", keys, me.INSTANCE_KEY_PATHS, found)
				logger.Debug(me.Prefix, "fetched instance keys (%v): %v", me.INSTANCE_KEY_PATHS, keys)

				if !found {
					logger.Debug(me.Prefix, "skipping element, no instance keys found")
				} else {
					if _, err = me.Matrix.AddInstance(strings.Join(keys, ".")); err != nil {
						logger.Error(me.Prefix, err.Error())
					} else {
						logger.Debug(me.Prefix, "added instance [%s]", strings.Join(keys, "."))
						count += 1
					}
				}
			}
		}
	}

	me.Metadata.LazySetValueUint64("count", "instance", count)
	logger.Debug(me.Prefix, "added %d instances to cache (old cache had %d)", count, old_count)

	if me.Matrix.SizeInstances() == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no instances fetched")
	}

	return nil, nil
}

func (me *Zapi) PollData() (*matrix.Matrix, error) {
	var err error
	var request, response *node.Node
	var fetch func(*matrix.Instance, *node.Node, []string)
	var count, skipped uint64
	var ad, rd, pd, bd2, fd, sd, sd2, bd, id time.Duration // Request/API time, Parse time, Fetch time
	var tag string
	var cl, content_length int64

	task_start := time.Now()

	count = 0
	skipped = 0

	api_d := time.Duration(0 * time.Second)
	read_d := time.Duration(0 * time.Second)
	parse_d := time.Duration(0 * time.Second)
	build_d := time.Duration(0 * time.Second)

	fd = time.Duration(0 * time.Second)
	sd = time.Duration(0 * time.Second)
	sd2 = time.Duration(0 * time.Second)
	bd = time.Duration(0 * time.Second)
	id = time.Duration(0 * time.Second)

	fetch = func(instance *matrix.Instance, node *node.Node, path []string) {

		newpath := append(path, node.GetNameS())
		logger.Debug(me.Prefix, " > %s(%s)%s <%s%d%s> name=[%s%s%s%s] value=[%s%s%s]", util.Grey, newpath, util.End, util.Red, len(node.GetChildren()), util.End, util.Bold, util.Cyan, node.GetNameS(), util.End, util.Yellow, node.GetContentS(), util.End)

		if value := node.GetContentS(); value != "" {
			key := strings.Join(newpath, ".")
			if metric := me.Matrix.GetMetric(key); metric != nil {
				if err := metric.SetValueString(instance, value); err != nil {
					//logger.Warn(me.Prefix, "%sskipped metric (%s) set value (%s): %v%s", util.Red, key, value, err, util.End)
					skipped += 1
				} else {
					//logger.Trace(me.Prefix, "%smetric (%s) set value (%s)%s", util.Green, key, value, util.End)
					count += 1
				}
			} else if label, has := me.INSTANCE_LABEL_PATHS[key]; has {
				instance.SetLabel(label, value)
				//logger.Trace(me.Prefix, "%slabel (%s) [%s] set value (%s)%s", util.Yellow, key, label, value, util.End)
				count += 1
			} else {
				logger.Debug(me.Prefix, "%sskipped (%s) with value (%s): not in metric or label cache%s", util.Blue, key, value, util.End)
				skipped += 1
			}
		} else {
			//logger.Trace(me.Prefix, "%sskippped (%s) with no value%s", util.Cyan, key, util.End)
			skipped += 1
		}

		for _, child := range node.GetChildren() {
			fetch(instance, child, newpath)
		}
	}

	logger.Debug(me.Prefix, "starting data poll")

	if err = me.Matrix.Reset(); err != nil {
		return nil, err
	}

	request = node.NewXmlS(me.Query)
	if me.System.Clustered {

		if me.Params.GetChildContentS("no_desired_attributes") != "true" {
			request.AddChild(me.desired_attributes)
		}
		if me.batch_size != "" && me.Params.GetChildContentS("no-max-records") != "true" {
			request.NewChildS("max-records", me.batch_size)
		}
	}

	/*
		fmt.Println("\n>>> my request tree <<<")
		request.Print(0)

		fmt.Println("\n>>> my xml dump <<<")
		if dump, err := tree.DumpXml(request); err == nil {
			fmt.Println(string(dump))
		} else {
			fmt.Println(err)
		}
	*/

	tag = "initial"

	batch_start := time.Now()

	for {

		invoke_start := time.Now()
		response, tag, cl, bd2, ad, rd, pd, err = me.Connection.InvokeBatchWithMoreTimers(request, tag)
		id += time.Since(invoke_start)

		content_length += cl

		if err != nil {
			return nil, err
		}

		if response == nil {
			break
		}

		/*
			fmt.Println(">>> got response tree <<<")
			response.Print(0)
		*/

		build_d += bd2
		api_d += ad
		read_d += rd
		parse_d += pd

		search_start := time.Now()
		instances := response.SearchChildren(me.SHORTEST_PATH_PREFIX)
		sd += time.Since(search_start)

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
			search2 := time.Now()
			keys, found := instanceElem.SearchContent(me.SHORTEST_PATH_PREFIX, me.INSTANCE_KEY_PATHS)
			sd2 += time.Since(search2)
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
			fetch_start := time.Now()
			fetch(instance, instanceElem, make([]string, 0))
			fd += time.Since(fetch_start)
		}
	}

	bd = time.Since(batch_start)

	// update metadata
	me.Metadata.LazySetValueInt64("api_time", "data", api_d.Microseconds())
	me.Metadata.LazySetValueInt64("read_time", "data", read_d.Microseconds())
	me.Metadata.LazySetValueInt64("build_time", "data", build_d.Microseconds())
	me.Metadata.LazySetValueInt64("parse_time", "data", parse_d.Microseconds())
	me.Metadata.LazySetValueInt64("fetch_time", "data", fd.Microseconds())
	me.Metadata.LazySetValueInt64("search_time", "data", sd.Microseconds())
	me.Metadata.LazySetValueInt64("search2_time", "data", sd2.Microseconds())
	me.Metadata.LazySetValueInt64("batch_time", "data", bd.Microseconds())
	me.Metadata.LazySetValueInt64("invoke_time", "data", id.Microseconds())
	me.Metadata.LazySetValueInt64("content_length", "data", content_length)
	me.Metadata.LazySetValueUint64("count", "data", count)
	me.AddCount(int(count))

	me.Metadata.LazySetValueInt64("task2_time", "data", time.Since(task_start).Microseconds())
	return me.Matrix, nil
}
