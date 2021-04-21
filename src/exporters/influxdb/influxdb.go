package main

import (
	"bytes"
	"fmt"
	"goharvest2/poller/exporter"
	"goharvest2/share/errors"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

/* Write metrics to the InfluxDB Time-Series Database.
   The exporter follows InfluxDB's v2 documentation
   for authentication and line protocol (measurement):

   - https://docs.influxdata.com/influxdb/v2.0/write-data/developer-tools/api/
   - https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/

*/

const (
	DEFAULT_PORT          = "8086"
	DEFAULT_TIMEOUT       = 5
	DEFAULT_API_VERSION   = "2"
	DEFAULT_API_PRECISION = "s"
	EXPECTED_STATUS_CODE  = 204
)

type InfluxDB struct {
	*exporter.AbstractExporter
	client *http.Client
	url    string
	token  string
}

func New(abc *exporter.AbstractExporter) exporter.Exporter {
	return &InfluxDB{AbstractExporter: abc}
}

func (e *InfluxDB) Init() error {

	if err := e.InitAbc(); err != nil {
		return err
	}

	var addr, port, bucket, org, v, p string
	var err error

	// check required / optional parameters
	if addr = e.Params.GetChildContentS("addr"); addr == "" {
		return errors.New(errors.MISSING_PARAM, "addr")
	}

	if port = e.Params.GetChildContentS("port"); port == "" {
		logger.Debug(e.Prefix, "using default port [%s]", DEFAULT_PORT)
		port = DEFAULT_PORT
	} else if _, err = strconv.Atoi(port); err != nil {
		return errors.New(errors.INVALID_PARAM, "port")
	}

	if bucket = e.Params.GetChildContentS("bucket"); bucket == "" {
		return errors.New(errors.MISSING_PARAM, "bucket")
	}
	logger.Debug(e.Prefix, "using bucket [%s]", bucket)

	if org = e.Params.GetChildContentS("org"); org == "" {
		return errors.New(errors.MISSING_PARAM, "org")
	}
	logger.Debug(e.Prefix, "using organization [%s]", org)

	if e.token = e.Params.GetChildContentS("token"); e.token == "" {
		return errors.New(errors.MISSING_PARAM, "token")
	} else {
		logger.Debug(e.Prefix, "will use authorization with api token")
	}

	if v = e.Params.GetChildContentS("version"); v == "" {
		v = DEFAULT_API_VERSION
	}
	logger.Debug(e.Prefix, "using api version [%s]", v)

	if p = e.Params.GetChildContentS("precision"); p == "" {
		p = DEFAULT_API_PRECISION
	}
	logger.Debug(e.Prefix, "using api precision [%s]", p)

	// timeout parameter
	timeout := time.Duration(DEFAULT_TIMEOUT) * time.Second
	if ct := e.Params.GetChildContentS("client_timeout"); ct != "" {
		if t, err := strconv.Atoi(ct); err == nil {
			timeout = time.Duration(t) * time.Second
		} else {
			logger.Warn(e.Prefix, "invalid client_timeout [%s], using default: %d s", ct, DEFAULT_TIMEOUT)
		}
	} else {
		logger.Debug(e.Prefix, "using default client_timeout: %d s", DEFAULT_TIMEOUT)
	}

	// construct client URL
	e.url = fmt.Sprintf("http://%s:%s/api/v%s/write?org=%s&bucket=%s&precision=%s", addr, port, v, org, bucket, p)
	logger.Debug(e.Prefix, "url= [%s]", e.url)

	// construct HTTP client
	e.client = &http.Client{Timeout: timeout}

	logger.Debug(e.Prefix, "initialized exporter, ready to emit to [%s:%s]", addr, port)
	return nil
}

func (e *InfluxDB) Export(data *matrix.Matrix) error {

	var metrics [][]byte
	var err error

	e.Lock()
	defer e.Unlock()

	if metrics, err = e.Render(data); err == nil {
		if e.Options.Debug {
			logger.Debug(e.Prefix, "simulating export since in debug mode")
			for _, m := range metrics {
				logger.Debug(e.Prefix, "M= [%s]", m)
			}
		} else {
			err = e.Emit(metrics)
		}
	}
	return err
}

func (e *InfluxDB) Emit(data [][]byte) error {
	var buffer *bytes.Buffer
	var request *http.Request
	var response *http.Response
	var err error

	buffer = bytes.NewBuffer(bytes.Join(data, []byte("\n")))

	if request, err = http.NewRequest("POST", e.url, buffer); err != nil {
		return err
	}

	request.Header.Set("Authorization", "Token "+e.token)

	if response, err = e.client.Do(request); err != nil {
		return err
	}

	if response.StatusCode != EXPECTED_STATUS_CODE {
		defer response.Body.Close()
		if body, err := ioutil.ReadAll(response.Body); err != nil {
			return errors.New(errors.API_RESPONSE, err.Error())
		} else {
			return errors.New(errors.API_REQ_REJECTED, string(body))
		}
	}
	return nil
}

func (e *InfluxDB) Render(data *matrix.Matrix) ([][]byte, error) {

	rendered := make([][]byte, 0)

	object := data.Object

	// user-defined preferences for export
	var labels_to_include, keys_to_include []string
	include_all := false
	if data.GetExportOptions().GetChildContentS("include_all_labels") == "true" {
		include_all = true
	}
	if x := data.GetExportOptions().GetChildS("instance_keys"); x != nil {
		keys_to_include = x.GetAllChildContentS()
	}
	if x := data.GetExportOptions().GetChildS("instance_labels"); x != nil {
		labels_to_include = x.GetAllChildContentS()
	}

	// count number of data points rendered
	count := uint64(0)

	// not all collectors provide timestamp, so this might be nil
	//timestamp := data.GetMetric("timestamp")
	//var timestamp *matrix.Metric
	// temporarily disabled, influx expects nanosecs, we get something else from zapis

	// measurement that we will not emit
	// only to store global labels that we'll
	// add to all instances
	global := NewMeasurement("", 0)
	for key, value := range data.GetGlobalLabels().Map() {
		global.AddTag(key, value)
	}

	// render one measurement for each instance
	for key, instance := range data.GetInstances() {

		if !instance.IsExportable() {
			continue
		}

		m := NewMeasurement(object, len(global.tag_set))
		copy(m.tag_set, global.tag_set)

		// tag set
		if include_all {
			for label, value := range instance.GetLabels().Map() {
                if value != "" {
				    m.AddTag(label, value)
                }
			}
		} else {
			for _, key := range keys_to_include {
				if value, has := instance.GetLabels().GetHas(key); has && value != "" {
					m.AddTag(key, value)
				}
			}
		}

		// skip instance without key tags
		if len(m.tag_set) == 0 {
			logger.Debug(e.Prefix, "skip instance (%s), no tag set parsed from labels (%v)", key, instance.GetLabels().Map())
		}

		// field set

		// strings
		for _, label := range labels_to_include {
			if value, has := instance.GetLabels().GetHas(label); has && value != "" {
				m.AddField(label, value)
				count += 1
			}
		}

		// numeric
		for _, metric := range data.GetMetrics() {

			if !metric.IsExportable() {
				continue
			}

			value, ok := metric.GetValueString(instance)

			if !ok {
				continue
			}

			field_name := metric.GetName()

			if metric.HasLabels() {
				for _, label := range metric.GetLabels().Map() {
					field_name += "_" + label
				}
			}

			m.AddField(field_name, value)

			count += 1
		}

		/*
			// optionially add timestamp
			if timestamp != nil {
				if value, ok := data.GetValue(timestamp, instance); ok {
					m.SetTimestamp(strconv.FormatFloat(value, 'f', 0, 64))
				}
			}*/

		if r, err := m.Render(); err == nil {
			rendered = append(rendered, []byte(r))
		} else {
			logger.Debug(e.Prefix, err.Error())
		}
	}
	e.AddExportCount(count)
	logger.Debug(e.Prefix, "rendered %d measurements with %d data points for (%s)", len(rendered), count, object)
	return rendered, nil
}
