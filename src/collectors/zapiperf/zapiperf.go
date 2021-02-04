package main

import (

    "goharvest2/poller/structs/matrix"
	"goharvest2/poller/structs/options"
	"goharvest2/poller/structs/set"
    "goharvest2/poller/yaml"
    "goharvest2/poller/xml"
    "goharvest2/poller/share"
    "goharvest2/poller/share/logger"
    "goharvest2/poller/collector"
)

var Log *logger.Logger = logger.New(1, "")

type ZapiPerf struct {
    *collector.AbstractCollector
    connection client.Client
    system client.SystemInfo
    object_raw string
}

func New(name, obj string, options *options.Options, params *yaml.Node) collector.Collector {
    a := collector.New(name, obj, options, params)
    return &ZapiPerf{AbstractCollector: a}
}

func (z *ZapiPerf) Init() error {

	var err error

    Log = logger.New(c.Options.LogLevel, c.Name+":"+c.Object)
    
    if c.connection, err = client.New(c.Params); err != nil {
        //Log.Error("connecting: %v", err)
        return err
    }

    if c.system, err = c.connection.GetSystemInfo(); err != nil {
        //Log.Error("system info: %v", err)
        return err
    }

    Log.Debug("Connected to: %s", c.system.String())

    template_fn := c.Params.GetChild("objects").GetChildValue(c.Object) // @TODO err handling

    template, err := collector.ImportObjectTemplate(c.Options.Path, "default", template_fn, c.Name, c.system.Version)
    if err != nil {
        Log.Error("Error importing subtemplate: %s", err)
        return err
    }
    c.Params.Union(template, false)
 
    if err := c.InitAbc(); err != nil {
        return err
    }

    if expopt := c.Params.GetChild("export_options"); expopt != nil {
        c.Data.SetExportOptions(expopt)
    } else {
        return errors.New("missing export options")
    }

    c.Metadata.AddMetric("api_time", "api_time", true) // extra metric for measuring api time

    if c.object_raw = c.Params.GetChildValue("object"); c.object_raw == "" {
        Log.Warn("Missing object in template")
    }

    c.Data.Object = c.object_raw
    c.Metadata.Object = c.object_raw
	
    counters := c.Params.GetChild("counters")
    if counters == nil {
        Log.Warn("Missing counters in template")
    }

    if c.object_raw == "" || counters == nil {
        return errors.New("missing parameters")
	}
	

    
}

func (c *ZapiPerf) poll_data() error {

	var err error

	Log.Debug("Updating data cach")

	NewData := c.Data.Clone()
	if err = NewData.InitData(); err != nil {
		return err
	}

	timestamp := NewData.GetMetric("timestamp")
	if timestamp == nil {
		return errors.New("missing timestamp metric")
	}

	instance_keys := make([]string, 0, len(NewData.Instances))
	for key, _ := range NewData.Instances {
		instance_keys = append(instance_keys, key)
	}

	start_index := 0
	end_index := 0

	for end_index < len(instance_keys) {

		// no builtin min/max in Go...
		end_index += c.batch_size
		if end_index > len(instance_keys) {
			end_index = len(instance_keys)
		}

		// build API request
		request := xml.New("perf-object-get-instances")
		request.CreateChild("objectname", c.query)
		request.CreateChild("max-records", strconv.Itoa(c.batch_size))

		// load batch instances
		request_instances := xml.New(c.instance_key+"s")
		for _, key range instance_keys[start_index:end_index] {
			request_instances.CreateChild(c.instance_key, key)
		}
		request.AddChild(request_instances)
		start_index = end_index 

		request_counters := xml.New("counters")
		for key,_ range NewData.Metrics {
			request_counters.CreateChild("counter", key)
		}
		request.AddChild(request_counters)

		if err = c.connection.BuildRequest(request); err != nil {
			Log.Error("build request: %v", err)
			break
		}

		response, err := c.connection.InvokeRequest()
		if err != nil {
			Log.Error("data request: %v", err)
			break
		}

		// fetch instances
		instances := response.GetChild("attributes-list")
		if instances == nil || len(instances.GetChildren()) == 0 {
			//@TODO ErrNoInstances
			break
		}

		// timestamp for batch instances
		ts, e := strconv.ParseFloat(response.GetChildValueS("timestamp"), 32)
		if e != nil {
			Log.Warn("invalid timestamp value [%s]", response.GetChildValueS("timestamp"))
			//@TODO ...
		}


		for _, i := range instances.GetChildren() {

			key := i.GetChildContentS(c.instance_key)
			if key == "" {
				Log.Debug("skip instance, no key [%s] (name=%s, uuid=%s)", c.instance_key, i.GetChildContentS("name"), i.GetChildContent("uuid"))
				continue
			}
			
			instance := NewData.GetInstance(key)
			if instance == nil {
				Log.Warn("skip instance [%s], not found in cache", key)
				continue
			}

			counters := request.GetChild("counters")
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

				if NewData.InstanceLabels.Has(name) { // @TODO implement
					NewData.SetInstanceLabel(instance, name, value)
					Log.Debug("+ label [%s= %s"], name, value)
					continue
				}

				// numeric 
				counters := NewData.GetMetric(name)
				if counters == nil {
					Log.Debug("metric [%s] not in cache, skip", name)
					continue
				}

				if counter.Scalar {
					if e := NewData.SetValueString(counter, instance, value); e != nil {
						Log.Error("set metric [%s] with value [%s]: %v", name, value, e)
					} else {
						Log.Debug("+ scalar metric [%s= %s]", name, value)
					}
				} else {
					if e := NewData.SetArrayValuesString(counter, instance, strings.Split(value, ",")) e != nil {
						Log.Error("set array metric [%s] with values [%s]: %v", name, value, e)
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
		return err
	}

	// skip calculating from delta if no data from previous poll
	if c.Data.IsEmpty()  {
		Log.Debug("no cache from previous poll, no postprocessing until next poll")
		c.Data = NewData
		return nil
	}

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

	// calculate timestamp delta once
	NewData.Delta(c.Data, timestamp, timestamp)

	for _, m := range ordered_metrics {

		// raw counters don't require postprocessing
		if m.Property == "raw" {
			continue
		}

		// scalar not depending on base counter
		if m.Scalar {
			if m.BaseMetric == "" {
				c.calculate_from_delta(NewData, m.Index, -1, m.Properties) // -1 indicates no base counter
			} else if b := NewData.GetMetric(m.BaseMetric); b != nil {
				c.calculate_from_delta(NewData, m.Index, b.Index, m.Properties)
			} else {
				Log.Error("required base [%s] for scalar [%s] missing", m.BaseMetric, m.Display)
			}
			continue
		}
		// array metric, it becomes a bit complicated here
		// since base counter can be array as well
		if m.BaseMetric == "" {
			for i:=0; i<m.Size; i+=1 {
				c.calculate_from_delta(NewData, m.Index+i, -1, m.Properties)
			}
		} else if b := NewData.GetMetric(m.BaseMetric); b != nil {
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
			Log.Error("required base [%s] for array [%s] missing", m.BaseMetric, m.Display)
		}
	}

	// store cache for next poll
	c.Data = CacheData
	c.Data.IsEmpty = false // @redundant

	return NewData, nil	
}

func (c *ZapiPerf) calculate_from_delta(B *matrix.Matrix, metricIndex, baseIndex int, properties string) {
	
	A := c.Data // for convenience

	// calculate metric delta for all instances from previous cache
	B.Delta(A, metricIndex, metricIndex)

	if strings.Contains(properties, "delta") {
		//B.Data[metric.Index] = delta
		return
	}

	if strings.Contains(properties, "rate") {
		if ts := B.GetMetric("timestamp"); ts != nil {
			B.Divide(A, metricIndex, ts.Index)
		} else {
			Log.Error("timestamp counter not found")
		}
		return
	}

	// For the next two properties we need base counters
	// Calculate delta of base counter

	if base < 0 {
		Log.Error("no base counter index for") // should never happen
		return
	}

	B.Divide(A, metricIndex, baseIndex)

	// @TODO minimum ops threshold
	if strings.Contains(properties, "average") {
		return
	}

	if strings.Contains(properties, "percent") {
		B.MultByScalar(metricIndex, float64(100))
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
		return errors.New("no counters defined in template")
	}

	Log.Debug("Updating metric cache (old cache: %d metrics and %d labels", 
		len(c.Data.LabelNames),
		len(c.Data.Metrics))

	c.Data.ResetMetrics()
	c.Data.ResetLabelNames()

	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	ts := matrix.Metric{Display: "timestamp", Properties: "raw", Enabled: false}
	if err := c.Data.AddCustomMetric("timestamp", &ts); err != nil {
		Log.Error("add timestamp metric: %v", err)
	}

	// build request
	request := xml.New("perf-object-counter-list-info")
	request.CreateChild("objectname", c.Params.GetChildValue("query"))

	if err = c.connection.BuildRequest(request); err != nil {
		Log.Error("build request: %v", err)
		break
	}
			
	response, err := c.connection.InvokeRequest()
	if err != nil {
		return err
	}

	// fetch counter elements
	counters := response.GetChild("counters")
	if counters == nil {
		return errors.New("no counters in response")
	}
	
	var counter *xml.Node

	for _, counter := range counters.GetChildren() {
		name := counter.GetName()

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
	if !replaced.Empty() {
		for _, counter := range counters.GetChildren() {
			name := counter.GetName()
			if replaced.Has(name) {
				if r:= c.add_counter(counter, name, true); r != "" && !wanted.Has(r) {
					missing.Add(r)  // required base counter, missing in template
			}
		}
	}
	
	// third loop for required base counters, not in template
	if !missing.Empty() {
		for _, counter := range counters.GetChildren() {
			name := counter.GetName()
			if missing.Has(name) {
				Log.Debug("Adding missing base counter [%s]", name)
				c.add_counter(counter, name, false)
			}
		}
	}
	// @TODO - return ErrNoMetrics if not counters were loaded
	//         and enter standby mode
	return nil
}

func (c *ZapiPerf) add_counter(counter *xml.Node, name string, enabled bool) string {

	properties := counter.GetChildContentS("properties")
	base_counter := counter.GetChildContentS("base-counter")
	unit := counter.GetChildContentS("unit")
	display := strings.Replace(name, "-", "_")

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

	if _, exists := c.Data.GetMetric(name); exists {
		Log.Debug("Skipping counter [%s], already in cache", name)
		return ""
	}

	m := matrix.Metric{Display: display, Properties: properties, BaseCounter: base_counter, Enabled: enabled}

	// counter type is array

	if counter.GetChildContentS("type") == "array" {

		m.Scalar = false

		labels_element := counter.GetChild("labels")
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

	Log.Debug("Updating instance cache (old cache has: %d)", len(c.Data.Instances()))
	c.Data.ResetInstances()

	batch_tag := "initial"

	for batch_tag != "" {

		// build request
		request := xml.New("perf-object-counter-list-info")
		request.CreateChild("objectname", c.query)
		request.CreateChild("max-records", strconv.Itoa(c.batch_size))
		if batch_tag != "initial" {
			request.CreateChild("tag", batch_tag)
		}

		if err = c.connection.BuildRequest(request); err != nil {
			Log.Error("build request: %v", err)
			break
		}

		response, err := c.connection.InvokeRequest()
		if err != nil {
			Log.Error("instance request: %v", err)
			break
		}

		// @TODO next-tag bug
		batch_tag = response.GetChildContentS("next-tag")

		// fetch instances
		instances := response.GetChild("attributes-list")
		if instances == nil || len(instances.GetChildren()) == 0 {
			break
		}

		for _, i := range instances.GetChildren() {

			key := i.GetChildContentS(c.instance_key)

			if key == "" {
				Log.Debug("skip instance, no key [%s] (name=%s, uuid=%s)", 
					c.instance_key, 
					i.GetChildContentS("name"), 
					i.GetChildContent("uuid")
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