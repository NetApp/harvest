package main

import (

	"strings"
	"strconv"
	"time"
	"fmt"
    "goharvest2/share/logger"
	"goharvest2/share/errors"
	"goharvest2/poller/collector"
    "goharvest2/poller/struct/matrix"
	"goharvest2/poller/struct/set"
	"goharvest2/poller/struct/dict"
    "goharvest2/poller/struct/xml"
    "goharvest2/poller/util"
	
	client "goharvest2/poller/api/zapi"
)

// default parameter values
const (
	INSTANCE_KEY = "uuid"
	BATCH_SIZE = 500
	LATENCY_IO_REQD = 10
)


type ZapiPerf struct {
    *collector.AbstractCollector
    connection *client.Client
    system *client.System
	object string
	query string
	template_fn string
	template_type string
	instance_key string
	batch_size int
	latency_io_reqd int
}

func New(a *collector.AbstractCollector) collector.Collector {
    return &ZapiPerf{AbstractCollector: a}
}

func (c *ZapiPerf) Init() error {

	var err error
	
	// create client to talk to ONTAP system
    if c.connection, err = client.New(c.Params); err != nil {
        return err
    }

	// fetch system info
    if c.system, err = c.connection.GetSystem(); err != nil {
        return err
    }
	logger.Debug(c.Prefix, "Connected to: %s", c.system.String())
	
    template_fn := c.Params.GetChild("objects").GetChildValue(c.Object) // @TODO err handling

    template, err := collector.ImportSubTemplate(c.Options.Path, "default", template_fn, c.Name, c.system.Version)
    if err != nil {
        logger.Error(c.Prefix, "Error importing subtemplate: %s", err)
        return err
    }
	c.Params.Union(template, false)
	
    // object name from subtemplate
    if c.object = c.Params.GetChildValue("object"); c.object == "" {
        return errors.New(errors.MISSING_PARAM, "object")
    }
 
    // Invoke generic initializer
    // this will load Schedule, initialize Data and Metadata
    if err := collector.Init(c); err != nil {
        return err
    }

    // Add system (cluster) name 
    c.Data.SetGlobalLabel("system", c.system.Name)

	// @TODO cleanup
    c.Data.Object = c.object
    c.Metadata.Object = c.object
	
    if c.Params.GetChild("counters") == nil {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	if c.query = c.Params.GetChildValue("query"); c.query == "" {
		return errors.New(errors.MISSING_PARAM, "query")
	}

	c.instance_key = c.loadParamStr("instance_key", INSTANCE_KEY)
	c.batch_size = c.loadParamInt("batch_size", BATCH_SIZE)
	c.latency_io_reqd = c.loadParamInt("latency_io_reqd", LATENCY_IO_REQD)
	
	logger.Debug(c.Prefix, "Successfully initialized")
	return nil

}

func (c *ZapiPerf) loadParamStr(name, default_value string) string {

	var x string

	if x = c.Params.GetChildValue(name); x != "" {
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

	if x = c.Params.GetChildValue(name); x != "" {
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

	// list of instance keys (instance names or uuids) for which
	// we will request counter data
	instance_keys := make([]string, 0, len(NewData.Instances))
	for key, _ := range NewData.Instances {
		instance_keys = append(instance_keys, key)
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

		// build API request
		request := xml.New("perf-object-get-instances")
		request.CreateChild("objectname", c.query)
		//request.CreateChild("max-records", strconv.Itoa(c.batch_size))

		// load batch instances
		key_name := "instance-uuid"
		if c.instance_key == "name" {
			key_name = "instance"
		}

		request_instances := xml.New(key_name+"s")
		for _, key := range instance_keys[start_index:end_index] {
			request_instances.CreateChild(key_name, key)
		}
		start_index = end_index
		request.AddChild(request_instances)

		request_counters := xml.New("counters")
		for key, _ := range NewData.Metrics {
			request_counters.CreateChild("counter", key)
		}
		for key,_ := range NewData.LabelNames.Iter() {
			request_counters.CreateChild("counter", key)
		}
		request.AddChild(request_counters)

		if err = c.connection.BuildRequest(request); err != nil {
			logger.Error(c.Prefix, "build request: %v", err)
			//break
			return nil, err
		}

		response, rd, pd, err := c.connection.InvokeWithTimers()
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
		instances, found := response.GetChild("instances")
		if !found || instances == nil || len(instances.GetChildren()) == 0 {
			logger.Warn(c.Prefix, "no instances")
			//@TODO ErrNoInstances
			break
			return nil, errors.New(errors.ERR_NO_INSTANCE, "")
		}

		logger.Debug(c.Prefix, "fetched batch with %d instances", len(instances.GetChildren()))

		// timestamp for batch instances
		//ts, e := strconv.ParseFloat(response.GetChildContentS("timestamp"), 32)
		//if e != nil {
		//	logger.Warn(c.Prefix, "invalid timestamp value [%s]", response.GetChildContentS("timestamp"))
		//	//@TODO ...
		//}
		ts := float32(time.Now().UnixNano()) / 1000000000

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

			counters, _ := i.GetChild("counters")
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

				if NewData.LabelNames.Has(name) { // @TODO implement
					NewData.SetInstanceLabel(instance, name, value)
					logger.Debug(c.Prefix, "+ label [%s= %s%s%s]", name, util.Yellow, value, util.End)
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

				if counter.Scalar {
					if e := NewData.SetValueString(counter, instance, string(value)); e != nil {
						logger.Error(c.Prefix, "set metric [%s] with value [%s]: %v", name, value, e)
						data_count += 1
					} else {
						logger.Debug(c.Prefix, "+ scalar metric [%s= %s%s%s]", name, util.Cyan, value, util.End)
						v, ok := NewData.GetValue(counter, instance)
						logger.Debug(c.Prefix, "%s(%f) (%v)%s", util.Grey, v, ok, util.End)
					}
				} else {
					if e := NewData.SetArrayValuesString(counter, instance, strings.Split(string(value), ",")); e != nil {
						logger.Error(c.Prefix, "set array metric [%s] with values [%s]: %v", name, value, e)
						data_count += 1
					} else {
						logger.Debug(c.Prefix, "+ array metric [%s]", util.Green, value, util.End)
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
	c.Metadata.SetValueSS("response_time", "data", float32(response_d.Seconds()) / float32(batch_count))
	c.Metadata.SetValueSS("parse_time", "data", float32(parse_d.Seconds()) / float32(batch_count))
	c.Metadata.SetValueSS("count", "data", float32(data_count))

	logger.Debug(c.Prefix, "collected data: %d batch polls, %d data points", batch_count, data_count)

	// fmt.Println()
	// fmt.Println()
	for _, m := range NewData.GetMetrics() {
		print_vector(fmt.Sprintf("%s(%d) %s%s%s", util.Grey, m.Index, util.Cyan, m.Display, util.End), NewData.Data[m.Index])
	}
	// fmt.Println()
	// fmt.Println()

	// skip calculating from delta if no data from previous poll
	if c.Data.IsEmpty()  {
		logger.Debug(c.Prefix, "no cache from previous poll, so postprocessing until next poll (new data empty: %v)", NewData.IsEmpty())
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
	NewData.Delta(c.Data, timestamp.Index)
	// fmt.Printf("\npostprocessing %s%s%s - %s%v%s\n", util.Red, timestamp.Display, util.End, util.Bold, timestamp.Properties, util.End)
	print_vector("current", NewData.Data[timestamp.Index])
	print_vector("previous", c.Data.Data[timestamp.Index])


	for _, m := range ordered_metrics {

		// raw counters don't require postprocessing
		if strings.Contains(m.Properties, "raw") {
			continue
		}

		//logger.Debug(c.Prefix, "Postprocessing %s%s%s (%s%v%s)", util.Red, m.Display, util.End, util.Bold, m.Properties, util.End)
		// fmt.Printf("\npostprocessing %s%s%s - %s%v%s\n", util.Red, m.Display, util.End, util.Bold, m.Properties, util.End)
		// scalar not depending on base counter
		if m.Scalar {
			if m.BaseCounter == "" {
				// fmt.Printf("scalar - no basecounter\n")
				c.calculate_from_delta(NewData, m.Display, m.Index, -1, m.Properties) // -1 indicates no base counter
			} else if b := NewData.GetMetric(m.BaseCounter); b != nil {
				// fmt.Printf("scalar - with basecounter %s%s%s (%s)\n", util.Red, m.BaseCounter, util.End, b.Properties)
				c.calculate_from_delta(NewData, m.Display, m.Index, b.Index, m.Properties)
			} else {
				logger.Error(c.Prefix, "required base [%s] for scalar [%s] missing", m.BaseCounter, m.Display)
			}
			continue
		}
		// array metric, it becomes a bit complicated here
		// since base counter can be array as well
		if m.BaseCounter == "" {
			// fmt.Printf("array - no basecounter\n")
			for i:=0; i<m.Size; i+=1 {
				c.calculate_from_delta(NewData, m.Display, m.Index+i, -1, m.Properties)
			}
		} else if b := NewData.GetMetric(m.BaseCounter); b != nil {
			if b.Scalar {
				// fmt.Printf("array - scalar basecounter %s%s%s (%s)\n", util.Red, m.BaseCounter, util.End, b.Properties)
				for i:=m.Index; i<m.Size; i+=1 {
					c.calculate_from_delta(NewData, m.Display, m.Index+i, b.Index, m.Properties)
				}
			} else if m.Size == b.Size {
				// fmt.Printf("array - array basecounter %s%s%s (%s)\n", util.Red, m.BaseCounter, util.End, b.Properties)
				for i:=0; i<m.Size; i+= 1 {
					c.calculate_from_delta(NewData, m.Display, m.Index+i, b.Index+i, m.Properties)
				}
			} else {
				logger.Error(c.Prefix, "size of [%s] (%d) does not match with base [%s] (%d)", m.Display, m.Size, b.Display, b.Size)
			}
		} else {
			logger.Error(c.Prefix, "required base [%s] for array [%s] missing", m.BaseCounter, m.Display)
		}
	}

	// store cache for next poll
	c.Data = CachedData
	//c.Data.IsEmpty = false // @redundant

	return NewData, nil	
}


func print_vector(tag string, x []float32) {
	// fmt.Printf("%-35s", tag)
	for i:=0; i<len(x); i+=1 {
		// fmt.Printf("%25f", x[i])
	}
	// fmt.Println()
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
			NewData.Divide(metricIndex, ts.Index, float32(0))
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
			NewData.Divide(metricIndex, baseIndex, float32(c.latency_io_reqd))
		} else {
			NewData.Divide(metricIndex, baseIndex, float32(0))
		}
		print_vector(fmt.Sprintf("%s average%s", util.Green, util.End), NewData.Data[metricIndex])
		return
	}

	if strings.Contains(properties, "percent") {
		NewData.Divide(metricIndex, baseIndex, float32(0))
		NewData.MultByScalar(metricIndex, float32(100))
		print_vector(fmt.Sprintf("%s percent%s", util.Green, util.End), NewData.Data[metricIndex])
		return
	}

	logger.Error(c.Prefix, "unexpected properties: %s", properties)
	return
}

func (c *ZapiPerf) PollCounter() (*matrix.Matrix, error) {

	replaced := set.New() // deprecated and replaced counters
	missing := set.New() // missing base counters
	wanted := dict.New() // counters listed in template

	if counters := c.Params.GetChildValues("counters"); len(counters) != 0 {
		for _, c := range counters {
			if x := strings.Split(c, "=>"); len(x) == 2 {
				wanted.Set(strings.TrimSpace(x[0]), strings.TrimSpace(x[1]))
			} else {
				wanted.Set(c, "")
			}
		}
	} else {
		return nil, errors.New(errors.MISSING_PARAM, "no counters defined in template")
	}

	logger.Debug(c.Prefix, "Updating metric cache (old cache: %d metrics and %d labels", 
		c.Data.LabelNames.Size(),
		len(c.Data.Metrics))

	c.Data.ResetMetrics()
	c.Data.ResetLabelNames()

	// build request
	request := xml.New("perf-object-counter-list-info")
	request.CreateChild("objectname", c.Params.GetChildValue("query"))

	if err := c.connection.BuildRequest(request); err != nil {
		return nil, err
	}
			
	response, err := c.connection.Invoke()
	if err != nil {
		return nil, err
	}

	// fetch counter elements
	counters, _ := response.GetChild("counters")
	if counters == nil {
		return nil, errors.New(errors.ERR_NO_METRIC, "no counters in response")
	}
	
	for _, counter := range counters.GetChildren() {
		name := counter.GetChildContentS("name")

		display, ok := wanted.GetHas(name)
		// counter not requested
		if !ok {
			logger.Debug(c.Prefix, "Skipping [%s]", name)
			continue
		}

		// deprecated and possibly replaced counter
		if counter.GetChildContentS("is-deprecated") == "true" {

			if r := counter.GetChildContentS("replaced-by"); r != "" {
				logger.Info(c.Prefix, "Counter [%s] deprecated, replacing with [%s]", name, r)
				if !wanted.Has(r) {
					replaced.Add(r)
				}
			} else {
				logger.Info(c.Prefix, "Counter [%s] deprecated, skipping", name)
			}
			continue
		}

		// add counter to our cache
		if r := c.add_counter(counter, name, display, true); r != "" && !wanted.Has(r) {
			missing.Add(r) // required base counter, missing in template
		}
	}

	// second loop for replaced counters
	if !replaced.IsEmpty() {
		for _, counter := range counters.GetChildren() {
			name := counter.GetName()
			if replaced.Has(name) {
				if r:= c.add_counter(counter, name, "", true); r != "" && !wanted.Has(r) {
					missing.Add(r)  // required base counter, missing in template
				}
			}
		}
	}
	
	// third loop for required base counters, not in template
	if !missing.IsEmpty() {
		for _, counter := range counters.GetChildren() {
			name := counter.GetName()
			if missing.Has(name) {
				logger.Debug(c.Prefix, "Adding missing base counter [%s]", name)
				c.add_counter(counter, name, "", false)
			}
		}
	}

	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	ts := matrix.Metric{Display: "timestamp", Properties: "raw", Scalar: true, Enabled: false}
	if err := c.Data.AddCustomMetric("timestamp", &ts); err != nil {
		logger.Error(c.Prefix, "add timestamp metric: %v", err)
	}

	logger.Debug(c.Prefix, "Added %d label and %d numeric metrics to cache", c.Data.LabelNames.Size(), len(c.Data.Metrics))

	// @TODO - return ErrNoMetrics if not counters were loaded
	//         and enter standby mode
	return nil, nil
}

func (c *ZapiPerf) add_counter(counter *xml.Node, name, display string, enabled bool) string {

	properties := counter.GetChildContentS("properties")
	base_counter := counter.GetChildContentS("base-counter")
	unit := counter.GetChildContentS("unit")
	if display == "" {
		display = strings.ReplaceAll(name, "-", "_") // redundant
	}
	
	logger.Debug(c.Prefix, "Handling counter [%s] with properties [%s] and unit [%s]", name, properties, unit)

	// string counters, add as instance label name
	if strings.Contains(properties, "string") {
		c.Data.AddLabelKeyName(name, display)
		logger.Debug(c.Prefix, "%s+[%s] added as label name%s", util.Yellow, name, util.End)
		return ""
	}

	// numerical counter

	// make sure counter is not already in cache
	// this might happen with base counters that were marked missing

	if m := c.Data.GetMetric(name); m != nil {
		logger.Debug(c.Prefix, "Skipping counter [%s], already in cache", name)
		return ""
	}

	m := matrix.Metric{Display: display, Properties: properties, BaseCounter: base_counter, Enabled: enabled}

	// counter type is array

	if counter.GetChildContentS("type") == "array" {

		m.Scalar = false

		labels_element, _ := counter.GetChild("labels")
		if labels_element == nil {
			logger.Warn(c.Prefix, "Counter [%s] type is array, but subcounters missing", name)
			return ""
		}

		labels := labels_element.GetChildren()

		if len(labels) == 0 || len(labels) > 2 {

			logger.Warn(c.Prefix, "Skipping [%s] type array, unexpected (%d) dimensions", name, len(labels))
			return ""
		}

		labelsA := xml.DecodeHtml(labels[0].GetContentS())
		
		if len(labels) == 1 {

			m.Labels = strings.Split(labelsA, ",")
			m.Dimensions = 1
			m.Size = len(m.Labels)

		} else if len(labels) == 2 {

			labelsB := xml.DecodeHtml(labels[1].GetContentS())

			m.SubLabels = strings.Split(labelsB, ",")
			m.Dimensions = 2
			m.Size = len(m.Labels) * len(m.SubLabels)
		}

	} else {
	// coutner type is scalar
		m.Scalar = true
	}

	if err := c.Data.AddCustomMetric(name, &m); err != nil {
		logger.Error(c.Prefix, "add metric [%s]: %v", name, err)
	}

	return base_counter
}

func (c *ZapiPerf) PollInstance() (*matrix.Matrix, error) {

	var err error

	logger.Debug(c.Prefix, "Updating instance cache (old cache has: %d)", len(c.Data.Instances))
	c.Data.ResetInstances()

	batch_tag := "initial"

	for batch_tag != "" {

		// build request
		request := xml.New("perf-object-instance-list-info-iter")
		request.CreateChild("objectname", c.query)
		request.CreateChild("max-records", strconv.Itoa(c.batch_size))
		if batch_tag != "initial" {
			request.CreateChild("tag", batch_tag)
		}

		if err = c.connection.BuildRequest(request); err != nil {
			logger.Error(c.Prefix, "build request: %v", err)
			break
		}

		response, err := c.connection.Invoke()
		if err != nil {
			logger.Error(c.Prefix, "instance request: %v", err)
			break
		}

		// @TODO next-tag bug
		batch_tag = response.GetChildContentS("next-tag")

		// fetch instances
		instances, _ := response.GetChild("attributes-list")
		if instances == nil || len(instances.GetChildren()) == 0 {
			break
		}

		for _, i := range instances.GetChildren() {

			key := i.GetChildContentS(c.instance_key)

			if key == "" {
				logger.Debug(c.Prefix, "skip instance, no key [%s] (name=%s, uuid=%s)", 
					c.instance_key, 
					i.GetChildContentS("name"),
					i.GetChildContentS("uuid"),
				)
			} else if _, e := c.Data.AddInstance(key); e != nil {
				logger.Warn(c.Prefix, "add instance: %v", e)
			}
			logger.Debug(c.Prefix, "added instance [%s]", key)
		}
	}

	logger.Debug(c.Prefix, "Added %d instances", len(c.Data.Instances))

	// @TODO ErrNoInstances
	
	return nil, err
	
}
