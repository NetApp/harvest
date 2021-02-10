package main

import (

	"strings"
	"strconv"
	"sync"
	"time"
    "goharvest2/share/logger"
	"goharvest2/poller/errors"
    "goharvest2/poller/struct/matrix"
	"goharvest2/poller/struct/options"
	"goharvest2/poller/struct/set"
    "goharvest2/poller/struct/yaml"
    "goharvest2/poller/struct/xml"
    "goharvest2/poller/share"
	"goharvest2/poller/collector"
	
	client "goharvest2/poller/api/zapi"
)

const (
	DEFAULT_BATCH_SIZE = 500
	DEFAULT_INSTANCE_KEY = "uuid"
)

var Log *logger.Logger = logger.New(1, "")

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
}

func New(name, obj string, options *options.Options, params *yaml.Node) collector.Collector {
    a := collector.New(name, obj, options, params)
    return &ZapiPerf{AbstractCollector: a}
}

func (c *ZapiPerf) Init() error {

	var err error

    Log = logger.New(c.Options.LogLevel, c.Name+":"+c.Object)
	
	// create client to talk to ONTAP system
    if c.connection, err = client.New(c.Params); err != nil {
        return err
    }

	// fetch system info
    if c.system, err = c.connection.GetSystem(); err != nil {
        return err
    }

	Log.Debug("Connected to: %s", c.system.String())
	
	c.template_type = "default" // @temp

	// fetch filename of sub-template from ZapiPerf template (default.yaml/custom.yaml)
    if o := c.Params.GetChild("objects"); o != nil {
		c.template_fn = o.GetChildValue(c.Object)
		// import subtemplate of current object
		if t, err := collector.ImportSubTemplate(c.Options.Path, c.template_type, c.template_fn, c.Name, c.system.Version); err != nil {
			return err
		} else { // merge subtemplate to inherited main template
			c.Params.Union(t, false)
		}
	} else {
		return errors.New(errors.ERR_CONFIG, "subtemplate filename in main template")
	}
 
	// @TODO verify order fields are initialized
    if err := c.InitAbc(); err != nil {
        return err
    }

	// Set export options for data
    if export_opts := c.Params.GetChild("export_options"); export_opts != nil {
        c.Data.SetExportOptions(export_opts)
    } else {
        return errors.New(errors.ERR_CONFIG, "no export options in subtemplate")
    }

	// @TODO move to ABC
    c.Metadata.AddMetric("api_time", "api_time", true) // extra metric for measuring api time

    if c.object = c.Params.GetChildValue("object"); c.object == "" {
		return errors.New(errors.ERR_CONFIG, "object name")
    }

    c.Data.Object = c.object
    c.Metadata.Object = c.object
	
    if c.Params.GetChild("counters") == nil {
		return errors.New(errors.MISSING_PARAM, "counters list")
	}

	if c.query = c.Params.GetChildValue("query"); c.query == "" {
		return errors.New(errors.MISSING_PARAM, "query")
	}

	if c.instance_key = c.Params.GetChildValue("instance_key"); c.instance_key == "" {
		c.instance_key = DEFAULT_INSTANCE_KEY
		Log.Debug("Using [%s] as instance key (default)", DEFAULT_INSTANCE_KEY)
	} else if c.instance_key != "name" && c.instance_key != "uuid" {
		Log.Warn("Invalid instance key [%s], using default [%s] instead", c.instance_key, DEFAULT_INSTANCE_KEY)
		c.instance_key = DEFAULT_INSTANCE_KEY
	} else {
		Log.Debug("Using [%s] as instance key", c.instance_key)
	}

	if c.batch_size, err = strconv.Atoi(c.Params.GetChildValue("batch_size")); err != nil {
		c.batch_size = DEFAULT_BATCH_SIZE
		Log.Debug("Using batch-size [%d] (default)", c.batch_size)
	} else if c.batch_size > 500 || c.batch_size < 1 {
		Log.Warn("Invalid batch-size [%d], using [%d] instead (default)", c.batch_size, DEFAULT_BATCH_SIZE)
		c.batch_size = DEFAULT_BATCH_SIZE
	} else {
		Log.Debug("Using batch-size [%d]", c.batch_size)
	}
	
	Log.Debug("Successfully initialized")
	return nil

}


func (c *ZapiPerf) Start(wg *sync.WaitGroup) {

    defer wg.Done()

    for {

        c.Metadata.InitData()

        for _, task := range c.Schedule.GetTasks() {

            if c.Schedule.IsDue(task) {

                c.Schedule.Start(task)

                data, err := c.poll(task)

                if err != nil {
                    Log.Warn("%s poll failed: %v", task, err)
                    return
                }
                
                Log.Debug("%s poll completed", task)

                duration := c.Schedule.Stop(task)
                c.Metadata.SetValueSS("poll_time", task, duration.Seconds())
                
                if data != nil {

					//data.Print()
                    
                    Log.Debug("exporting to %d exporters", len(c.Exporters))

                    for _, e := range c.Exporters {
                        if err := e.Export(data); err != nil {
                            Log.Warn("export to [%s] failed: %v", e.GetName(), err)
                        }
                    }
                }
            }

            Log.Debug("exporting metadata")

            for _, e := range c.Exporters {
                if err := e.Export(c.Metadata); err != nil {
                    Log.Warn("Metadata export to [%s] failed: %v", e.GetName(), err)
                }
            }
        }

        d := c.Schedule.SleepDuration()
        Log.Debug("Sleeping %s until next poll session", d.String())
        c.Schedule.Sleep()
    }
}

func (c *ZapiPerf) poll(task string) (*matrix.Matrix, error) {
    switch task {
        case "data":
            return c.poll_data()
        case "instance":
			return nil, c.poll_instance()
		case "counter":
			return nil, c.poll_counter()
        default:
            return nil, errors.New(errors.ERR_CONFIG, "invalid task: " + task)
    }
}


func (c *ZapiPerf) poll_data() (*matrix.Matrix, error) {

	var err error

	Log.Debug("Updating data cache")

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

		Log.Debug("starting batch poll for instances [%d:%d]", start_index, end_index)

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
			Log.Error("build request: %v", err)
			break
		}

		response, rd, pd, err := c.connection.InvokeWithTimers()
		if err != nil {
			Log.Error("data request: %v", err)
			//@TODO handle "resource limit exceeded"
			break
		}

		response_d += rd
		parse_d += pd
		batch_count += 1

		// fetch instances
		instances, found := response.GetChild("instances")
		if !found || instances == nil || len(instances.GetChildren()) == 0 {
			Log.Warn("no instances")
			//@TODO ErrNoInstances
			break
		}

		Log.Debug("fetched batch with %d instances", len(instances.GetChildren()))

		// timestamp for batch instances
		ts, e := strconv.ParseFloat(response.GetChildContentS("timestamp"), 32)
		if e != nil {
			Log.Warn("invalid timestamp value [%s]", response.GetChildContentS("timestamp"))
			//@TODO ...
		}

		for _, i := range instances.GetChildren() {

			key := i.GetChildContentS(c.instance_key)
			if key == "" {
				Log.Debug("skip instance, no key [%s] (name=%s, uuid=%s)", c.instance_key, i.GetChildContentS("name"), i.GetChildContentS("uuid"))
				continue
			}
			
			instance := NewData.GetInstance(key)
			if instance == nil {
				Log.Warn("skip instance [%s], not found in cache", key)
				continue
			}

			counters, _ := i.GetChild("counters")
			if counters == nil {
				Log.Warn("skip instance [%s], no data counters", key)
				continue
			}
	
			Log.Debug("fetching data of instance [%s]", key)

			// add batch timestamp as custom counter
			NewData.SetValue(timestamp, instance, ts)

			for _, c := range counters.GetChildren() {

				name := c.GetChildContentS("name")
				value := c.GetChildContentS("value")

				// ZAPI counter for us is either instance label (string)
				// or numeric metric (scalar or string)

				if NewData.LabelNames.Has(name) { // @TODO implement
					NewData.SetInstanceLabel(instance, name, value)
					Log.Debug("+ label [%s= %s]", name, value)
					data_count += 1
					continue
				}

				// numeric 
				counter := NewData.GetMetric(name)
				if counter == nil {
					Log.Debug("metric [%s] [=%s] not in cache, skip", name, value)
					continue
				}
				Log.Debug("+ metric [%s] [=%s]", name, value)

				if counter.Scalar {
					if e := NewData.SetValueString(counter, instance, string(value)); e != nil {
						Log.Error("set metric [%s] with value [%s]: %v", name, value, e)
						data_count += 1
					} else {
						Log.Debug("+ scalar metric [%s= %s]", name, value)
					}
				} else {
					if e := NewData.SetArrayValuesString(counter, instance, strings.Split(string(value), ",")); e != nil {
						Log.Error("set array metric [%s] with values [%s]: %v", name, value, e)
						data_count += 1
					} else {
						Log.Debug("+ array metric [%s]", value)
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
	c.Metadata.SetValueSS("response_time", "data", response_d.Seconds() / float64(batch_count))
	c.Metadata.SetValueSS("parse_time", "data", parse_d.Seconds() / float64(batch_count))
	c.Metadata.SetValueSS("count", "data", float64(data_count))

	Log.Debug("collected data: %d batch polls, %d data points", batch_count, data_count)

	// skip calculating from delta if no data from previous poll
	if c.Data.IsEmpty()  {
		Log.Debug("no cache from previous poll, so postprocessing until next poll (new data empty: %v)", NewData.IsEmpty())
		c.Data = NewData
		return nil, nil
	}

	Log.Debug("starting delta calculations from previous poll")

	// cache data, to store after calculations
	CacheData := NewData.Clone()

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

	for _, m := range ordered_metrics {

		// raw counters don't require postprocessing
		if strings.Contains(m.Properties, "raw") {
			continue
		}

		Log.Debug("Postprocessing %s (%v)", m.Display, m.Properties)
		// scalar not depending on base counter
		if m.Scalar {
			if m.BaseCounter == "" {
				c.calculate_from_delta(NewData, m.Index, -1, m.Properties) // -1 indicates no base counter
			} else if b := NewData.GetMetric(m.BaseCounter); b != nil {
				c.calculate_from_delta(NewData, m.Index, b.Index, m.Properties)
			} else {
				Log.Error("required base [%s] for scalar [%s] missing", m.BaseCounter, m.Display)
			}
			continue
		}
		// array metric, it becomes a bit complicated here
		// since base counter can be array as well
		if m.BaseCounter == "" {
			for i:=0; i<m.Size; i+=1 {
				c.calculate_from_delta(NewData, m.Index+i, -1, m.Properties)
			}
		} else if b := NewData.GetMetric(m.BaseCounter); b != nil {
			if b.Scalar {
				for i:=m.Index; i<m.Size; i+=1 {
					c.calculate_from_delta(NewData, m.Index+i, b.Index, m.Properties)
				}
			} else if m.Size == b.Size {
				for i:=0; i<m.Size; i+= 1 {
					c.calculate_from_delta(NewData, m.Index+i, b.Index+i, m.Properties)
				}
			} else {
				Log.Error("size of [%s] (%d) does not match with base [%s] (%d)", m.Display, m.Size, b.Display, b.Size)
			}
		} else {
			Log.Error("required base [%s] for array [%s] missing", m.BaseCounter, m.Display)
		}
	}

	// store cache for next poll
	c.Data = CacheData
	//c.Data.IsEmpty = false // @redundant

	return NewData, nil	
}

func (c *ZapiPerf) calculate_from_delta(NewData *matrix.Matrix, metricIndex, baseIndex int, properties string) {
	
	PrevData := c.Data // for convenience

	Log.Debug("\nNewValues: %v", NewData.Data[metricIndex])
	Log.Debug("OldValues: %v", PrevData.Data[metricIndex])

	// calculate metric delta for all instances from previous cache
	NewData.Delta(PrevData, metricIndex)

	Log.Debug("Delta:    %v\n\n", NewData.Data[metricIndex])

	if strings.Contains(properties, "delta") {
		//B.Data[metric.Index] = delta
		return
	}

	if strings.Contains(properties, "rate") {
		if ts := NewData.GetMetric("timestamp"); ts != nil {
			NewData.Divide(metricIndex, ts.Index)
		} else {
			Log.Error("timestamp counter not found")
		}
		return
	}

	// For the next two properties we need base counters
	// We assume that delta of base counters is already calculated

	if baseIndex < 0 {
		Log.Error("no base counter index for") // should never happen
		return
	}

	NewData.Divide(metricIndex, baseIndex)

	// @TODO minimum ops threshold
	if strings.Contains(properties, "average") {
		return
	}

	if strings.Contains(properties, "percent") {
		NewData.MultByScalar(metricIndex, float64(100))
		return
	}

	Log.Error("unexpected properties: %s", properties)
	return
}

func (c *ZapiPerf) poll_counter() error {

	replaced := set.New() // deprecated and replaced counters
	missing := set.New() // missing base counters
	wanted := set.New() // counters listed in template

	if counters := c.Params.GetChildValues("counters"); len(counters) != 0 {
		for _, c := range counters {
			wanted.Add(c)
		}
	} else {
		return errors.New(errors.MISSING_PARAM, "no counters defined in template")
	}

	Log.Debug("Updating metric cache (old cache: %d metrics and %d labels", 
		c.Data.LabelNames.Size(),
		len(c.Data.Metrics))

	c.Data.ResetMetrics()
	c.Data.ResetLabelNames()

	// build request
	request := xml.New("perf-object-counter-list-info")
	request.CreateChild("objectname", c.Params.GetChildValue("query"))

	if err := c.connection.BuildRequest(request); err != nil {
		return err
	}
			
	response, err := c.connection.Invoke()
	if err != nil {
		return err
	}

	// fetch counter elements
	counters, _ := response.GetChild("counters")
	if counters == nil {
		return errors.New(errors.NO_METRICS, "no counters in response")
	}
	
	for _, counter := range counters.GetChildren() {
		name := counter.GetChildContentS("name")

		// counter not requested
		if !wanted.Has(name) {
			Log.Debug("Skipping [%s]", name)
			continue
		}

		// deprecated and possibly replaced counter
		if counter.GetChildContentS("is-deprecated") == "true" {

			if r := counter.GetChildContentS("replaced-by"); r != "" {
				Log.Info("Counter [%s] deprecated, replacing with [%s]", name, r)
				if !wanted.Has(r) {
					replaced.Add(r)
				}
			} else {
				Log.Info("Counter [%s] deprecated, skipping", name)
			}
			continue
		}

		// add counter to our cache
		if r := c.add_counter(counter, name, true); r != "" && !wanted.Has(r) {
			missing.Add(r) // required base counter, missing in template
		}
	}

	// second loop for replaced counters
	if !replaced.IsEmpty() {
		for _, counter := range counters.GetChildren() {
			name := counter.GetName()
			if replaced.Has(name) {
				if r:= c.add_counter(counter, name, true); r != "" && !wanted.Has(r) {
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
				Log.Debug("Adding missing base counter [%s]", name)
				c.add_counter(counter, name, false)
			}
		}
	}

	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	ts := matrix.Metric{Display: "timestamp", Properties: "raw", Scalar: true, Enabled: false}
	if err := c.Data.AddCustomMetric("timestamp", &ts); err != nil {
		Log.Error("add timestamp metric: %v", err)
	}

	Log.Debug("Added %d label and %d numeric metrics to cache", c.Data.LabelNames.Size(), len(c.Data.Metrics))

	// @TODO - return ErrNoMetrics if not counters were loaded
	//         and enter standby mode
	return nil
}

func (c *ZapiPerf) add_counter(counter *xml.Node, name string, enabled bool) string {

	properties := counter.GetChildContentS("properties")
	base_counter := counter.GetChildContentS("base-counter")
	unit := counter.GetChildContentS("unit")
	display := strings.ReplaceAll(name, "-", "_")

	Log.Debug("Handling counter [%s] with properties [%s] and unit [%s]", name, properties, unit)

	// string counters, add as instance label name
	if strings.Contains(properties, "string") {
		c.Data.AddLabelKeyName(name, display)
		Log.Debug("%s+[%s] added as label name%s", share.Yellow, name, share.End)
		return ""
	}

	// numerical counter

	// make sure counter is not already in cache
	// this might happen with base counters that were marked missing

	if m := c.Data.GetMetric(name); m != nil {
		Log.Debug("Skipping counter [%s], already in cache", name)
		return ""
	}

	m := matrix.Metric{Display: display, Properties: properties, BaseCounter: base_counter, Enabled: enabled}

	// counter type is array

	if counter.GetChildContentS("type") == "array" {

		m.Scalar = false

		labels_element, _ := counter.GetChild("labels")
		if labels_element == nil {
			Log.Warn("Counter [%s] type is array, but subcounters missing", name)
			return ""
		}

		labels := labels_element.GetChildren()

		if len(labels) == 0 || len(labels) > 2 {

			Log.Warn("Skipping [%s] type array, unexpected (%d) dimensions", name, len(labels))
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
		Log.Error("add metric [%s]: %v", name, err)
	}

	return base_counter
}

func (c *ZapiPerf) poll_instance() error {

	var err error

	Log.Debug("Updating instance cache (old cache has: %d)", len(c.Data.Instances))
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
			Log.Error("build request: %v", err)
			break
		}

		response, err := c.connection.Invoke()
		if err != nil {
			Log.Error("instance request: %v", err)
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
				Log.Debug("skip instance, no key [%s] (name=%s, uuid=%s)", 
					c.instance_key, 
					i.GetChildContentS("name"),
					i.GetChildContentS("uuid"),
				)
			} else if _, e := c.Data.AddInstance(key); e != nil {
				Log.Warn("add instance: %v", e)
			}
			Log.Debug("added instance [%s]", key)
		}
	}

	Log.Debug("Added %d instances", len(c.Data.Instances))

	// @TODO ErrNoInstances
	
	return err
	
}
