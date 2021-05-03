/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package main

import (
	"bytes"
	"fmt"
	"goharvest2/cmd/poller/exporter"
	"goharvest2/pkg/color"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/matrix"
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
	defaultPort          = "8086"
	detaultTimeout       = 5
	defaultApiVersion    = "2"
	defaultApiPrecision  = "s"
	expectedResponseCode = 204
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
		logger.Debug(e.Prefix, "using default port [%s]", defaultPort)
		port = defaultPort
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
		v = defaultApiVersion
	}
	logger.Debug(e.Prefix, "using api version [%s]", v)

	if p = e.Params.GetChildContentS("precision"); p == "" {
		p = defaultApiPrecision
	}
	logger.Debug(e.Prefix, "using api precision [%s]", p)

	// timeout parameter
	timeout := time.Duration(detaultTimeout) * time.Second
	if ct := e.Params.GetChildContentS("client_timeout"); ct != "" {
		if t, err := strconv.Atoi(ct); err == nil {
			timeout = time.Duration(t) * time.Second
		} else {
			logger.Warn(e.Prefix, "invalid client_timeout [%s], using default: %d s", ct, detaultTimeout)
		}
	} else {
		logger.Debug(e.Prefix, "using default client_timeout: %d s", detaultTimeout)
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

	var (
		metrics [][]byte
		err     error
		s       time.Time
	)

	e.Lock()
	defer e.Unlock()

	s = time.Now()

	// render the metrics, i.e. convert to InfluxDb line protocol
	if metrics, err = e.Render(data); err == nil && len(metrics) != 0 {
		// fix render time
		if err = e.Metadata.LazyAddValueInt64("time", "render", time.Since(s).Microseconds()); err != nil {
			logger.Error(e.Prefix, "metadata render time: %v", err)
		}
		// in debug mode, don't actually export but write to log
		if e.Options.Debug {
			logger.Debug(e.Prefix, "simulating export since in debug mode")
			for _, m := range metrics {
				logger.Debug(e.Prefix, "M= [%s%s%s]", color.Blue, m, color.End)
			}
			return nil
			// otherwise to the actual export: send to the DB
		} else if err = e.Emit(metrics); err != nil {
			logger.Error(e.Prefix, "(%s.%s) --> %s", data.Object, data.UUID, err.Error())
			return err
		}
	}

	logger.Debug(e.Prefix, "(%s.%s) --> exported %d data points", data.Object, data.UUID, len(metrics))

	// update metadata
	if err = e.Metadata.LazyAddValueInt64("time", "export", time.Since(s).Microseconds()); err != nil {
		logger.Error(e.Prefix, "metadata export time: %v", err)
	}

	/* skipped for now, since InfluxDB complains about "time" field name

	// export metadata
	if metrics, err = e.Render(e.Metadata); err != nil {
		logger.Error(e.Prefix, "render metadata: %v", err)
	} else if err = e.Emit(metrics); err != nil {
		logger.Error(e.Prefix, "emit metadata: %v", err)
	}

	*/
	return nil
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

	if response.StatusCode != expectedResponseCode {
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

	var (
		count, countTmp uint64
	)

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

	// measurement that we will not emit
	// only to store global labels that we'll
	// add to all instances
	global := NewMeasurement("", 0)
	for key, value := range data.GetGlobalLabels().Map() {
		global.AddTag(key, value)
	}

	// render one measurement for each instance
	for key, instance := range data.GetInstances() {

		countTmp = 0

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
				if value == "true" || value == "false" {
					m.AddField(label, value)
				} else {
					m.AddFieldString(label, value)
				}

				countTmp++
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
			countTmp++
		}

		logger.Trace(e.Prefix, "rendering from: %s", m.String())

		// skip instance with no tag set (no metrics)
		if len(m.field_set) == 0 {
			logger.Debug(e.Prefix, "skip instance (%s), no field set parsed", key)
		} else if r, err := m.Render(); err == nil {
			rendered = append(rendered, []byte(r))
			//logger.Debug(e.Prefix, "M= [%s%s%s]", color.Blue, r, color.End)
			count += countTmp
		} else {
			logger.Debug(e.Prefix, err.Error())
		}
	}

	logger.Debug(e.Prefix, "rendered %d measurements with %d data points for (%s)", len(rendered), count, object)

	// update metadata
	e.AddExportCount(count)
	if err := e.Metadata.LazySetValueUint64("count", "export", count); err != nil {
		logger.Error(e.Prefix, "metadata export count: %v", err)
	}
	return rendered, nil
}

// Need to appease go build - see https://github.com/golang/go/issues/20312
func main() {}
