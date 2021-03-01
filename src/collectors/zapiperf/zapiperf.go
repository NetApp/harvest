package main

import (
	"strings"
	"strconv"
	"time"
	"fmt"
	"path"
	//zapi_collector "goharvest2/collectors/zapi/collector"

    "goharvest2/share/logger"
    "goharvest2/share/tree/node"
	"goharvest2/share/errors"
    "goharvest2/share/util"
	"goharvest2/share/matrix"
	"goharvest2/share/set"
	"goharvest2/share/dict"
    "goharvest2/poller/collector"

	client "goharvest2/apis/zapi"
)

// default parameter values
const (
	INSTANCE_KEY = "uuid"
	BATCH_SIZE = 500
	LATENCY_IO_REQD = 0 //10
)


type ZapiPerf struct {
	*collector.AbstractCollector
	//*zapi_collector.Zapi  // provides: Connection, System, Object, Query, TemplateFn, TemplateType
    Connection *client.Client
    System *client.System
    object string
    Query string
	TemplateFn string
	TemplateType string
	batch_size int
	latency_io_reqd int
	instance_key string
}

func New(a *collector.AbstractCollector) collector.Collector {
    return &ZapiPerf{AbstractCollector: a}
}

/*
func New(a *collector.AbstractCollector) collector.Collector {
	z := zapi_collector.NewZapi(a)
    return &ZapiPerf{Zapi: z}
}*/

/* Copied from Zapi
 Don't change => merge!
*/

func (c *ZapiPerf) Init() error {

    // @TODO check if cert/key files exist
    if c.Params.GetChildContentS("auth_style") == "certificate_auth" {
        if c.Params.GetChildS("ssl_cert") == nil {
			cert_path := path.Join(c.Options.ConfPath, "cert", c.Options.Poller + ".pem")
            c.Params.NewChildS("ssl_cert", cert_path)
            logger.Debug(c.Prefix, "added ssl_cert path [%s]", cert_path)
        }

        if c.Params.GetChildS("ssl_key") == nil {
			key_path := path.Join(c.Options.ConfPath, "cert", c.Options.Poller + ".key")
            c.Params.NewChildS("ssl_key", key_path)
            logger.Debug(c.Prefix, "added ssl_key path [%s]", key_path)
        }
    }

    var err error
    if c.Connection, err = client.New(c.Params); err != nil {
        return err
    }

    // @TODO handle connectivity-related errors (retry a few times)
    if c.System, err = c.Connection.GetSystem(); err != nil {
        //logger.Error(c.Prefix, "system info: %v", err)
        return err
    }
    logger.Debug(c.Prefix, "Connected to: %s", c.System.String())

    c.TemplateFn = c.Params.GetChildS("objects").GetChildContentS(c.Object) // @TODO err handling

    model := "cdot"
    if !c.System.Clustered {
        model = "7mode"
    }
    template, err := c.ImportSubTemplate(model, "default", c.TemplateFn, c.System.Version)
    if err != nil {
        logger.Error(c.Prefix, "Error importing subtemplate: %s", err)
        return err
    }
    c.Params.Union(template)

    // object name from subtemplate
    if c.object = c.Params.GetChildContentS("object"); c.object == "" {
        return errors.New(errors.MISSING_PARAM, "object")
    }

    // api query literal
    if c.Query = c.Params.GetChildContentS("query"); c.Query == "" {
        return errors.New(errors.MISSING_PARAM, "query")
    }

    // Invoke generic initializer
    // this will load Schedule, initialize Data and Metadata
    if err := collector.Init(c); err != nil {
        return err
    }

    // overwrite from abstract collector
    c.Data.Object = c.object
    
    // Add system (cluster) name 
    c.Data.SetGlobalLabel("cluster", c.System.Name)

    // Initialize counter cache
    counters := c.Params.GetChildS("counters")
    if counters == nil {
        return errors.New(errors.MISSING_PARAM, "counters")
    }

    if err = c.InitCache(); err != nil {
        return err
    }

    logger.Debug(c.Prefix, "Successfully initialized")
    return nil
}

func (c *ZapiPerf) InitCache() error {
	c.instance_key = c.loadParamStr("instance_key", INSTANCE_KEY)
	c.batch_size = c.loadParamInt("batch_size", BATCH_SIZE)
	c.latency_io_reqd = c.loadParamInt("latency_io_reqd", LATENCY_IO_REQD)
	return nil
}

func (c *ZapiPerf) loadParamStr(name, default_value string) string {

	var x string

	if x = c.Params.GetChildContentS(name); x != "" {
		logger.Debug(c.Prefix, "using %s = [%s]", name, x)
		return x
	}
	logger.Debug(c.Prefix, "using %s = [%s] (default)", name, default_value)
	return default_value
}

func (c *ZapiPerf) loadParamInt(name string, default_value int) int {

	var x string
	var n int
	var e error

	if x = c.Params.GetChildContentS(name); x != "" {
		if n, e = strconv.Atoi(x); e == nil {
			logger.Debug(c.Prefix, "using %s = [%d]", name, n)
			return n
		}
		logger.Warn(c.Prefix, "invalid parameter %s = [%s] (expected integer)", name, x)
	}

	logger.Debug(c.Prefix, "using %s = [%d] (default)", name, default_value)
	return default_value
}

func (c *ZapiPerf) PollData() (*matrix.Matrix, error) {

	var err error

	logger.Debug(c.Prefix, "Updating data cache")

	NewData := c.Data.Clone()
	if err = NewData.InitData(); err != nil {
		return nil, err
	}

	timestamp := NewData.GetMetric("timestamp")
	if timestamp == nil {
		return nil, errors.New(errors.ERR_CONFIG, "missing timestamp metric")
	}

	// for updating metadata
	batch_count := 0
	data_count := 0
	response_d := time.Duration(0 * time.Second)
	parse_d := time.Duration(0 * time.Second)

	// determine what will serve as instance key (either "uuid" or "instance")
	key_name := "instance-uuid"
	if c.instance_key == "name" {
		key_name = "instance"
	}

	// list of instance keys (instance names or uuids) for which
	// we will request counter data
	instance_keys := make([]string, 0, len(NewData.Instances))
	for key, _ := range NewData.GetInstances() {
		instance_keys = append(instance_keys, key)
	}

	// build ZAPI request
	request := node.NewXmlS("perf-object-get-instances")
	request.NewChildS("objectname", c.Query)
	
	// load requested counters (metrics + labels)
	request_counters := request.NewChildS("counters", "")
	for key, _ := range NewData.GetMetrics() {
		request_counters.NewChildS("counter", key)
	}
	for key,_ := range NewData.GetLabels() {
		request_counters.NewChildS("counter", key)
	}
	
	// batch indices
	start_index := 0
	end_index := 0

	for end_index < len(instance_keys) {

		// update batch indices
		end_index += c.batch_size
		if end_index > len(instance_keys) {
			end_index = len(instance_keys)
		}

		logger.Debug(c.Prefix, "starting batch poll for instances [%d:%d]", start_index, end_index)

		request.PopChildS(key_name+"s")
		request_instances := request.NewChildS(key_name+"s", "")
		for _, key := range instance_keys[start_index:end_index] {
			request_instances.NewChildS(key_name, key)
		}

		start_index = end_index

		if err = c.Connection.BuildRequest(request); err != nil {
			logger.Error(c.Prefix, "build request: %v", err)
			//break
			return nil, err
		}

		response, rd, pd, err := c.Connection.InvokeWithTimers()
		if err != nil {
			//logger.Error(c.Prefix, "data request: %v", err)
			//@TODO handle "resource limit exceeded"
			//break
			return nil, err
		}

		response_d += rd
		parse_d += pd
		batch_count += 1

		// fetch instances
		instances := response.GetChildS("instances")
		if instances == nil || len(instances.GetChildren()) == 0 {
			err = errors.New(errors.ERR_NO_INSTANCE, "")
			break
		}

		logger.Debug(c.Prefix, "fetched batch with %d instances", len(instances.GetChildren()))

		// timestamp for batch instances
		// ignore timestamp from ZAPI which is always integer
		// we want float, since our poll interval can be float
		ts := float64(time.Now().UnixNano()) / 1000000000

		for _, i := range instances.GetChildren() {

			key := i.GetChildContentS(c.instance_key)
			if key == "" {
				logger.Debug(c.Prefix, "skip instance, no key [%s] (name=%s, uuid=%s)", c.instance_key, i.GetChildContentS("name"), i.GetChildContentS("uuid"))
				continue
			}
			
			instance := NewData.GetInstance(key)
			if instance == nil {
				logger.Warn(c.Prefix, "skip instance [%s], not found in cache", key)
				continue
			}

			counters := i.GetChildS("counters")
			if counters == nil {
				logger.Warn(c.Prefix, "skip instance [%s], no data counters", key)
				continue
			}
	
			logger.Debug(c.Prefix, "fetching data of instance [%s]", key)

			// add batch timestamp as custom counter
			NewData.SetValue(timestamp, instance, ts)

			for _, cnt := range counters.GetChildren() {

				name := cnt.GetChildContentS("name")
				value := cnt.GetChildContentS("value")

				// ZAPI counter for us is either instance label (string)
				// or numeric metric (scalar or string)

				if _, has := NewData.GetLabel(name); has { // @TODO implement
					NewData.SetInstanceLabel(instance, name, value)
					logger.Debug(c.Prefix, "+ label data [%s= %s%s%s]", name, util.Yellow, value, util.End)
					data_count += 1
					continue
				}

				// numeric 
				counter := NewData.GetMetric(name)
				if counter == nil {
					logger.Debug(c.Prefix, "metric [%s] [=%s] not in cache, skip", name, value)
					continue
				}
				//logger.Debug(c.Prefix, "+ metric [%s] [=%s]", name, value)

				if counter.IsScalar() {
					if e := NewData.SetValueString(counter, instance, string(value)); e != nil {
						logger.Error(c.Prefix, "set metric [%s] with value [%s]: %v", name, value, e)
					} else {
						logger.Debug(c.Prefix, "+ scalar data [%s = %s%s%s]", name, util.Cyan, value, util.End)
						data_count += 1
						v, ok := NewData.GetValue(counter, instance)
						logger.Debug(c.Prefix, "%s(%f) (%v)%s", util.Grey, v, ok, util.End)
					}
				} else {
					values := strings.Split(string(value), ",")
					if e := NewData.SetArrayValuesString(counter, instance, values); e != nil {
						logger.Error(c.Prefix, "set array metric [%s] with values [%s]: %v", name, value, e)
					} else {
						logger.Debug(c.Prefix, "+ array data [%s = %s%s%s]", name, util.Pink, value, util.End)
						data_count += len(values)
						v := NewData.GetArrayValues(counter, instance)
						logger.Debug(c.Prefix, "%s %v %s", util.Grey, v, util.End)
					}
				}
			} // end loop over counters
		} // end loop over instances
	} // end batch request

	// terminate if serious errors
	// @TODO handle...
	if err != nil {
		return nil, err
	}

	// update metadata
	c.Metadata.SetValueSS("response_time", "data", float64(response_d.Seconds()) / float64(batch_count))
	c.Metadata.SetValueSS("parse_time", "data", float64(parse_d.Seconds()) / float64(batch_count))
	c.Metadata.SetValueSS("count", "data", float64(data_count))

	logger.Debug(c.Prefix, "collected data: %d batch polls, %d data points", batch_count, data_count)

	// fmt.Println()
	// fmt.Println()
	//for _, m := range NewData.GetMetrics() {
	//	print_vector(fmt.Sprintf("%s(%d) %s%s%s", util.Grey, m.Index, util.Cyan, m.Name, util.End), NewData.Data[m.Index])
	//}
	// fmt.Println()
	// fmt.Println()

	// skip calculating from delta if no data from previous poll
	if c.Data.IsEmpty()  {
		logger.Debug(c.Prefix, "no postprocessing until next poll (new data empty: %v)", NewData.IsEmpty())
		c.Data = NewData
		return nil, nil
	}

	logger.Debug(c.Prefix, "starting delta calculations from previous poll")
	logger.Debug(c.Prefix, "data has dimensions (%d x %d)", len(NewData.Data), len(NewData.Data[0]))

	// cache data, to store after calculations
	CachedData := NewData.Clone()

	// order metrics, such that those requiring base counters are processed last
	ordered_metrics := make([]*matrix.Metric, 0, len(NewData.Metrics))
	for _, m := range NewData.Metrics {
		if m.BaseCounter == "" { // does not require base counter
			ordered_metrics = append(ordered_metrics, m)
		}
	}
	for _, m := range NewData.Metrics {
		if m.BaseCounter != "" { // requires base counter
			ordered_metrics = append(ordered_metrics, m)
		}
	}

	// calculate timestamp delta first since many counters require it for postprocessing
	// timestamp has "raw" property, so won't be postprocessed automatically
	// fmt.Printf("\npostprocessing %s%s%s - %s%v%s\n", util.Red, timestamp.Name, util.End, util.Bold, timestamp.Properties, util.End)
	logger.Debug(c.Prefix, "cooking [%s] (%s)", timestamp.Name, timestamp.Properties)
	print_vector("current", NewData.Data[timestamp.Index])
	print_vector("previous", c.Data.Data[timestamp.Index])
	NewData.Delta(c.Data, timestamp.Index)
	print_vector(util.Green+"delta"+util.End, NewData.Data[timestamp.Index])


	for _, m := range ordered_metrics {

		// raw counters don't require postprocessing
		if strings.Contains(m.Properties, "raw") {
			continue
		}

		//logger.Debug(c.Prefix, "Postprocessing %s%s%s (%s%v%s)", util.Red, m.Name, util.End, util.Bold, m.Properties, util.End)
		// fmt.Printf("\npostprocessing %s%s%s - %s%v%s\n", util.Red, m.Name, util.End, util.Bold, m.Properties, util.End)
		// scalar not depending on base counter
		if m.IsScalar() {
			if m.BaseCounter == "" {
				// fmt.Printf("scalar - no basecounter\n")
				logger.Debug(c.Prefix, "cooking [%d] [%s%s%s] (%s)", m.Index, util.Cyan, m.Name, util.End, m.Properties)
				c.calculate_from_delta(NewData, m.Name, m.Index, -1, m.Properties) // -1 indicates no base counter
			} else if b := NewData.GetMetric(m.BaseCounter); b != nil {
				// fmt.Printf("scalar - with basecounter %s%s%s (%s)\n", util.Red, m.BaseCounter, util.End, b.Properties)
				logger.Debug(c.Prefix, "cooking [%d] [%s%s%s] (%s) using base counter [%s] (%s)", m.Index, util.Cyan, m.Name, util.End, m.Properties, b.Name, b.Properties)
				c.calculate_from_delta(NewData, m.Name, m.Index, b.Index, m.Properties)
			} else {
				logger.Error(c.Prefix, "required base [%s] for scalar [%s] missing", m.BaseCounter, m.Name)
			}
			continue
		}
		// array metric, it becomes a bit complicated here
		// since base counter can be array as well
		if m.BaseCounter == "" {
			// fmt.Printf("array - no basecounter\n")
			for i:=0; i<m.Size; i+=1 {
				logger.Debug(c.Prefix, "cooking [%d] [%s%s : %s%s] (%s)", m.Index+i, util.Pink, m.Name, m.Labels[i], util.End, m.Properties)
				c.calculate_from_delta(NewData, m.Name, m.Index+i, -1, m.Properties)
			}
		} else if b := NewData.GetMetric(m.BaseCounter); b != nil {
			if b.IsScalar() {
				// fmt.Printf("array - scalar basecounter %s%s%s (%s)\n", util.Red, m.BaseCounter, util.End, b.Properties)
				for i:=0; i<m.Size; i+=1 {
					logger.Debug(c.Prefix, "cooking [%d] [%s%s : %s%s] (%s) using base counter [%s] (%s)", m.Index+i, util.Pink, m.Name, m.Labels[i], util.End, m.Properties, b.Name, b.Properties)
					c.calculate_from_delta(NewData, m.Name, m.Index+i, b.Index, m.Properties)
				}
			} else if m.Size == b.Size {
				// fmt.Printf("array - array basecounter %s%s%s (%s)\n", util.Red, m.BaseCounter, util.End, b.Properties)
				for i:=0; i<m.Size; i+= 1 {
					logger.Debug(c.Prefix, "cooking [%d] [%s%s : %s%s] (%s) using base counter [%s : %s] (%s)", m.Index+i, util.Pink, m.Name, m.Labels[i], util.End, m.Properties, b.Name, b.Labels[i], b.Properties)
					c.calculate_from_delta(NewData, m.Name, m.Index+i, b.Index+i, m.Properties)
				}
			} else {
				logger.Error(c.Prefix, "size of [%s] (%d) does not match with base [%s] (%d)", m.Name, m.Size, b.Name, b.Size)
			}
		} else {
			logger.Error(c.Prefix, "required base [%s] for array [%s] missing", m.BaseCounter, m.Name)
		}
	}

	// store cache for next poll
	c.Data = CachedData
	//c.Data.IsEmpty = false // @redundant

	return NewData, nil	
}


func print_vector(tag string, x []float64) {
	vector := []string{}
	for _, n := range x {
		vector = append(vector, strconv.FormatFloat(float64(n), 'f', 5, 64))
	}
	logger.Debug("--------", "%-35s => %v", tag, vector)
}

func (c *ZapiPerf) calculate_from_delta(NewData *matrix.Matrix, metricName string, metricIndex, baseIndex int, properties string) {
	
	PrevData := c.Data // for convenience

	print_vector("current", NewData.Data[metricIndex])
	print_vector("previous", PrevData.Data[metricIndex])

	// calculate metric delta for all instances from previous cache
	NewData.Delta(PrevData, metricIndex)

	print_vector("delta", NewData.Data[metricIndex])
	
	// fmt.Println()

	if strings.Contains(properties, "delta") {
		print_vector(fmt.Sprintf("%s delta%s", util.Green, util.End), NewData.Data[metricIndex])
		//B.Data[metric.Index] = delta
		return
	}

	if strings.Contains(properties, "rate") {
		if ts := NewData.GetMetric("timestamp"); ts != nil {
			NewData.Divide(metricIndex, ts.Index, float64(0))
			print_vector(fmt.Sprintf("%s rate%s", util.Green, util.End), NewData.Data[metricIndex])
		} else {
			logger.Error(c.Prefix, "timestamp counter not found")
		}
		return
	}

	// For the next two properties we need base counters
	// We assume that delta of base counters is already calculated

	if baseIndex < 0 {
		logger.Error(c.Prefix, "no base counter index for") // should never happen
		return
	}

	// @TODO minimum ops threshold
	if strings.Contains(properties, "average") {

		if strings.Contains(metricName, "latency") {
			NewData.Divide(metricIndex, baseIndex, float64(c.latency_io_reqd))
		} else {
			NewData.Divide(metricIndex, baseIndex, float64(0))
		}
		print_vector(fmt.Sprintf("%s average%s", util.Green, util.End), NewData.Data[metricIndex])
		return
	}

	if strings.Contains(properties, "percent") {
		NewData.Divide(metricIndex, baseIndex, float64(0))
		NewData.MultByScalar(metricIndex, float64(100))
		print_vector(fmt.Sprintf("%s percent%s", util.Green, util.End), NewData.Data[metricIndex])
		return
	}

	logger.Error(c.Prefix, "unexpected properties: %s", properties)
	return
}

func (c *ZapiPerf) PollCounter() (*matrix.Matrix, error) {

	var (
		err error
		request, response, counter_list, counter_elems *node.Node
		old_metrics, old_labels, replaced, missing *set.Set
		wanted *dict.Dict
		old_metrics_size, old_labels_size int
	)

	old_metrics = set.New() // current set of metrics, so we can remove from matrix if not updated
	old_labels = set.New() // current set of labels
	wanted = dict.New() // counters listed in template, maps raw name to display name
	missing = set.New() // required base counters, missing in template
	replaced = set.New() // deprecated and replaced counters

	for key, _ := range c.Data.GetMetrics() {
		old_metrics.Add(key)
	}
	old_metrics_size = old_metrics.Size()

	for key, _ := range c.Data.GetLabels() {
		old_labels.Add(key)
	}
	old_labels_size = old_labels.Size()

	// parse list of counters defined in template
	if counter_list = c.Params.GetChildS("counters"); counter_list != nil {
		for _, cnt := range counter_list.GetAllChildContentS() {
			if renamed := strings.Split(cnt, "=>"); len(renamed) == 2 {
				wanted.Set(strings.TrimSpace(renamed[0]), strings.TrimSpace(renamed[1]))
			} else {
				display := strings.ReplaceAll(cnt, "-", "_")
				if strings.HasPrefix(display, c.object) {
					display = strings.TrimPrefix(display, c.object)
					display = strings.TrimPrefix(display, "_")
				}
				wanted.Set(cnt, display)
			}
		}
	} else {
		return nil, errors.New(errors.MISSING_PARAM, "counters")
	}

	logger.Debug(c.Prefix, "updating metric cache (old cache has %d metrics and %d labels", old_metrics.Size(), old_labels.Size())

	// build request
	request = node.NewXmlS("perf-object-counter-list-info")
	request.NewChildS("objectname", c.Query)

	if err = c.Connection.BuildRequest(request); err != nil {
		return nil, err
	}
			
	if response, err = c.Connection.Invoke(); err != nil {
		return nil, err
	}

	// fetch counter elements
	counter_elems = response.GetChildS("counters")
	
	if counter_elems == nil || len(counter_elems.GetChildren()) == 0 {
		return nil, errors.New(errors.ERR_NO_METRIC, "no counters in response")
	}
	
	for _, counter := range counter_elems.GetChildren() {
		key := counter.GetChildContentS("name")

		display, ok := wanted.GetHas(key)
		// counter not requested
		if !ok {
			logger.Trace(c.Prefix, "%sskip [%s], not requested%s", util.Grey, key, util.End)
			continue
		}

		// deprecated and possibly replaced counter
		if counter.GetChildContentS("is-deprecated") == "true" {

			if r := counter.GetChildContentS("replaced-by"); r != "" {
				logger.Info(c.Prefix, "replaced deprecated counter [%s] with [%s]", key, r)
				if !wanted.Has(r) {
					replaced.Add(r)
				}
			} else {
				logger.Info(c.Prefix, "skip [%s], deprecated", key)
				continue
			}
		}

		// string metric, add as instance label
		if strings.Contains(counter.GetChildContentS("properties"), "string") {
			old_labels.Delete(key)
			if display == "instance_name" {
				display = c.object
			}
			c.Data.AddLabel(key, display)
			logger.Debug(c.Prefix, "%s+[%s] added as label name (%s)%s", util.Yellow, key, display, util.End)
		} else {
			// add counter as numeric metric
			old_metrics.Delete(key)
			if r := c.add_counter(counter, key, display, true); r != "" && !wanted.Has(r) {
				missing.Add(r) // required base counter, missing in template
				logger.Debug(c.Prefix, "%smarking [%s] as required base counter for [%s]%s", util.Red, r, key, util.End)
			}
		}
	}

	// second loop for replaced counters
	if replaced.Size() > 0 {
		logger.Debug(c.Prefix, "attempting to retrieve metadata of %d replaced counters", replaced.Size())
		for _, counter := range counter_elems.GetChildren() {
			name := counter.GetChildContentS("name")
			if replaced.Has(name) {
				old_metrics.Delete(name)
				logger.Debug(c.Prefix, "adding [%s] (replacment for deprecated counter)", name)
				if r:= c.add_counter(counter, name, name, true); r != "" && !wanted.Has(r) {
					missing.Add(r)  // required base counter, missing in template
					logger.Debug(c.Prefix, "%smarking [%s] as required base counter for [%s]%s", util.Red, r, name, util.End)
				}
			}
		}
	}
	
	// third loop for required base counters, not in template
	if missing.Size() > 0 {
		logger.Debug(c.Prefix, "attempting to retrieve metadata of %d missing base counters", missing.Size())
		for _, counter := range counter_elems.GetChildren() {
			name := counter.GetChildContentS("name")
			//logger.Debug(c.Prefix, "%shas??? [%s]%s", util.Grey, name, util.End)
			if missing.Has(name) {
				old_metrics.Delete(name)
				logger.Debug(c.Prefix, "adding [%s] (missing base counter)", name)
				c.add_counter(counter, name, "", false)
			}
		}
	}

	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	if !old_metrics.Has("timestamp") {
		ts := matrix.Metric{Name: "timestamp", Properties: "raw", Size: 1, Enabled: false}
		if err := c.Data.AddCustomMetric("timestamp", &ts); err != nil {
			logger.Error(c.Prefix, "add timestamp metric: %v", err)
		}
	}

	for key, _ := range old_metrics.Iter() {
		if ! (key == "timestamp") {
			c.Data.RemoveMetric(key)
			logger.Debug(c.Prefix, "removed metric [%s]", key)
		}
	}

	for key, _ := range old_labels.Iter() {
		c.Data.RemoveLabel(key)
		logger.Debug(c.Prefix, "removed label [%s]", key)
	}

	metrics_added := c.Data.SizeMetrics() - (old_metrics_size - old_metrics.Size())
	labels_added := c.Data.SizeLabels() - (old_labels_size - old_labels.Size())

	if metrics_added > 0 || old_metrics.Size() > 0 {
		logger.Info(c.Prefix, "added %d new, removed %d metrics (total: %d)", metrics_added, old_metrics.Size(), c.Data.SizeMetrics())
	} else {
		logger.Debug(c.Prefix, "added %d new, removed %d metrics (total: %d)", metrics_added, old_metrics.Size(), c.Data.SizeMetrics())
	}

	if labels_added > 0 || old_labels.Size() > 0 {
		logger.Info(c.Prefix, "added %d new, removed %d labels (total: %d)", labels_added, old_labels.Size(), c.Data.SizeLabels())
	} else {
		logger.Debug(c.Prefix, "added %d new, removed %d labels (total: %d)", labels_added, old_labels.Size(), c.Data.SizeLabels())
	}

	if c.Data.SizeMetrics() == 0 {
		return nil, errors.New(errors.ERR_NO_METRIC, "")
	}

	return nil, nil
}

func (c *ZapiPerf) add_counter(counter *node.Node, name, display string, enabled bool) string {

	properties := counter.GetChildContentS("properties")
	base_counter := counter.GetChildContentS("base-counter")
	unit := counter.GetChildContentS("unit")
	if display == "" {
		display = strings.ReplaceAll(name, "-", "_") // redundant
	}
	
	logger.Debug(c.Prefix, "handling counter [%s] with properties [%s] and unit [%s]", name, properties, unit)

	// numerical counter

	// make sure counter is not already in cache
	// this might happen with base counters that were marked missing

	if m := c.Data.GetMetric(name); m != nil {
		logger.Debug(c.Prefix, "skipping counter [%s], already in cache", name)
		return ""
	}

	m := matrix.Metric{Name: display, Properties: properties, BaseCounter: base_counter, Enabled: enabled}

	// counter type is array

	if counter.GetChildContentS("type") == "array" {

		//m.IsScalar() = false

		labels_element := counter.GetChildS("labels")
		if labels_element == nil {
			logger.Warn(c.Prefix, "counter [%s] type is array, but subcounters missing", name)
			return ""
		}

		labels := labels_element.GetChildren()

		if len(labels) == 0 || len(labels) > 2 {

			logger.Warn(c.Prefix, "skipping [%s] type array, unexpected (%d) dimensions", name, len(labels))
			return ""
		}

		labelsA := node.DecodeHtml(labels[0].GetContentS())
		
		if len(labels) == 1 {

			m.Labels = strings.Split(labelsA, ",")
			m.Dimensions = 1
			m.Size = len(m.Labels)

		} else if len(labels) == 2 {

			labelsB := node.DecodeHtml(labels[1].GetContentS())

			m.SubLabels = strings.Split(labelsB, ",")
			m.Dimensions = 2
			m.Size = len(m.Labels) * len(m.SubLabels)
		}
		logger.Debug(c.Prefix, "%s+[%s] added as array metric (%s)%s", util.Pink, name, display, util.End)

	} else {
	// coutner type is scalar
		m.Size = 1
		logger.Debug(c.Prefix, "%s+[%s] added as scalar metric (%s)%s", util.Cyan, name, display, util.End)
	}

	if err := c.Data.AddCustomMetric(name, &m); err != nil {
		logger.Error(c.Prefix, "add metric [%s]: %v", name, err)
	}

	return base_counter
}

func (c *ZapiPerf) PollInstance() (*matrix.Matrix, error) {

	var (
		err error
		request *node.Node
		old_instances *set.Set
		old_size, new_size, removed, added int
		instances_attr string
	)

	old_instances = set.New()
	for key, _ := range c.Data.GetInstances() {
		old_instances.Add(key)
	}
	old_size = old_instances.Size()

	logger.Debug(c.Prefix, "updating instance cache (old cache has: %d)", old_instances.Size())

	if c.System.Clustered {
		request = node.NewXmlS("perf-object-instance-list-info-iter")
		instances_attr = "attributes-list"
	} else {
		request = node.NewXmlS("perf-object-instance-list-info")
		instances_attr = "instances"
	}

	request.NewChildS("objectname", c.Query)
    if c.System.Clustered {
    	request.NewChildS("max-records", strconv.Itoa(c.batch_size))
    }

	batch_tag := "initial"

	for batch_tag != "" {

		// build request
		if batch_tag != "initial" {
			request.PopChildS("tag")
			request.NewChildS("tag", batch_tag)
		}

		if err = c.Connection.BuildRequest(request); err != nil {
			logger.Error(c.Prefix, "build request: %v", err)
			break
		}

		response, err := c.Connection.Invoke()
		if err != nil {
			logger.Error(c.Prefix, "instance request: %v", err)
			break
		}

		// @TODO next-tag bug
		batch_tag = response.GetChildContentS("next-tag")

		// fetch instances
		instances := response.GetChildS(instances_attr)
		if instances == nil || len(instances.GetChildren()) == 0 {
			break
		}

		for _, i := range instances.GetChildren() {

			if key := i.GetChildContentS(c.instance_key); key == "" {
				// instance key missing
				n := i.GetChildContentS("name")
				u := i.GetChildContentS("uuid")
				logger.Debug(c.Prefix, "skip instance, missing key [%s] (name=%s, uuid=%s)", c.instance_key, n, u)
			} else if old_instances.Delete(key) {
				// instance already in cache
				continue
			} else if _, e := c.Data.AddInstance(key); e != nil {
				logger.Warn(c.Prefix, "add instance: %v", e)
			} else {
				logger.Debug(c.Prefix, "added new instance [%s]", key)
			}
		}
	}

	for key, _ := range old_instances.Iter() {
		c.Data.RemoveInstance(key)
		logger.Debug(c.Prefix, "removed instance [%s]", key)
	}

	removed = old_instances.Size()
	new_size = c.Data.SizeInstances()
	added = new_size - (old_size - removed)

	if added > 0 || removed > 0 {
		logger.Info(c.Prefix, "added %d new, removed %d (total instances %d)", added, removed, new_size)
	} else {
		logger.Debug(c.Prefix, "added %d new, removed %d (total instances %d)", added, removed, new_size)
	}
	
	if new_size == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "")
	}
	
	return nil, err
	
}
