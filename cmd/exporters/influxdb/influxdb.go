/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package influxdb

import (
	"bytes"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"io"
	"net/http"
	url2 "net/url"
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
	defaultPort          = 8086
	detaultTimeout       = 5
	defaultAPIVersion    = "2"
	defaultAPIPrecision  = "s"
	expectedResponseCode = 204
)

// some field names that we need to avoid
// first two: to avoid collision with label names
// others: protected field names by influxdb
var protectedFieldNames = map[string]string{
	"status":       "status_code",
	"new_status":   "new_status_code",
	"time":         "harvest_time",
	"_measurement": "harvest_measurement",
	"_field":       "harvest_field",
}

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

	var (
		url, addr, bucket, org, token, version, precision *string
		port                                              *int
	)

	// check the required / optional parameters
	// customer should provide either url or addr
	// url is expected to be the full write URL with all query params specified (optionally with scheme)
	// when url is defined, addr, bucket, org, port, and precision are ignored

	// addr is expected to include host only (no scheme, no port)
	// when addr is defined, bucket, org, port, and precision are required

	dbEndpoint := "addr"
	if url = e.Params.URL; url != nil {
		e.url = *url
		dbEndpoint = "url"
	} else {
		if addr = e.Params.Addr; addr == nil {
			return errs.New(errs.ErrMissingParam, "url or addr")
		}
		if port = e.Params.Port; port == nil {
			e.Logger.Debug().Msgf("using default port [%d]", defaultPort)
			defPort := defaultPort
			port = &defPort
		}
		if version = e.Params.Version; version == nil {
			v := defaultAPIVersion
			version = &v
		}
		e.Logger.Debug().Msgf("using api version [%s]", *version)

		if bucket = e.Params.Bucket; bucket == nil {
			return errs.New(errs.ErrMissingParam, "bucket")
		}
		e.Logger.Debug().Msgf("using bucket [%s]", *bucket)

		if org = e.Params.Org; org == nil {
			return errs.New(errs.ErrMissingParam, "org")
		}
		e.Logger.Debug().Msgf("using organization [%s]", *org)

		if precision = e.Params.Precision; precision == nil {
			p := defaultAPIPrecision
			precision = &p
		}
		e.Logger.Debug().Msgf("using api precision [%s]", *precision)

		urlToUSe := "http://" + *addr + ":" + strconv.Itoa(*port)
		url = &urlToUSe
		e.url = fmt.Sprintf("%s/api/v%s/write?org=%s&bucket=%s&precision=%s",
			*url, *version, url2.PathEscape(*org), url2.PathEscape(*bucket), *precision)
	}

	if token = e.Params.Token; token == nil {
		return errs.New(errs.ErrMissingParam, "token")
	}
	e.token = *token
	e.Logger.Debug().Msg("will use authorization with api token")

	// timeout parameter
	timeout := time.Duration(detaultTimeout) * time.Second
	if ct := e.Params.ClientTimeout; ct != nil {
		if t, err := strconv.Atoi(*ct); err == nil {
			timeout = time.Duration(t) * time.Second
		} else {
			e.Logger.Warn().Msgf("invalid client_timeout [%s], using default: %d s", *ct, detaultTimeout)
		}
	} else {
		e.Logger.Debug().Msgf("using default client_timeout: %d s", detaultTimeout)
	}

	e.Logger.Debug().Str("dbEndpoint", dbEndpoint).Str("url", e.url).Msg("")

	// construct HTTP client
	e.client = &http.Client{Timeout: timeout}

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
			e.Logger.Error().Stack().Err(err).Msg("metadata render time")
		}
		// in debug mode, don't actually export but write to log
		if e.Options.Debug {
			e.Logger.Debug().Msg("simulating export since in debug mode")
			for _, m := range metrics {
				e.Logger.Debug().Msgf("M= [%s%s%s]", color.Blue, m, color.End)
			}
			return nil
			// otherwise, to the actual export: send to the DB
		} else if err = e.Emit(metrics); err != nil {
			e.Logger.Error().Stack().Err(err).
				Str("object", data.Object).
				Str("uuid", data.UUID).
				Msg("Failed to emit metrics")
			return err
		}
	}

	e.Logger.Debug().Msgf("(%s.%s) --> exported %d data points", data.Object, data.UUID, len(metrics))

	// update metadata
	if err = e.Metadata.LazySetValueInt64("time", "export", time.Since(s).Microseconds()); err != nil {
		e.Logger.Error().Err(err).Msg("metadata export time")
	}

	if metrics, err = e.Render(e.Metadata); err != nil {
		e.Logger.Error().Err(err).Msg("render metadata")
	} else if err = e.Emit(metrics); err != nil {
		e.Logger.Error().Err(err).Msg("emit metadata")
	}

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
		defer func(Body io.ReadCloser) { _ = Body.Close() }(response.Body)
		if body, err := io.ReadAll(response.Body); err != nil {
			return errs.New(errs.ErrAPIResponse, err.Error())
		} else {
			return fmt.Errorf("%w: %s", errs.ErrAPIRequestRejected, string(body))
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
	var labelsToInclude, keysToInclude []string
	includeAll := false
	if data.GetExportOptions().GetChildContentS("include_all_labels") == "true" {
		includeAll = true
	}
	if x := data.GetExportOptions().GetChildS("instance_keys"); x != nil {
		keysToInclude = x.GetAllChildContentS()
	}
	if x := data.GetExportOptions().GetChildS("instance_labels"); x != nil {
		labelsToInclude = x.GetAllChildContentS()
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

		m := NewMeasurement(object, len(global.tagSet))
		copy(m.tagSet, global.tagSet)

		// tag set
		if includeAll {
			for label, value := range instance.GetLabels().Map() {
				if value != "" {
					m.AddTag(label, value)
				}
			}
		} else {
			for _, key := range keysToInclude {
				if value, has := instance.GetLabels().GetHas(key); has && value != "" {
					m.AddTag(key, value)
				}
			}
		}

		// skip instance without key tags
		if len(m.tagSet) == 0 {
			e.Logger.Debug().Msgf("skip instance (%s), no tag set parsed from labels (%v)", key, instance.GetLabels().Map())
		}

		// field set

		// strings
		for _, label := range labelsToInclude {
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

			fieldName := metric.GetName()

			if metric.HasLabels() {
				for _, label := range metric.GetLabels().Map() {
					fieldName += "_" + label
				}
			}

			if rename, has := protectedFieldNames[fieldName]; has {
				fieldName = rename
			}

			m.AddField(fieldName, value)
			countTmp++
		}

		e.Logger.Trace().Msgf("rendering from: %s", m.String())

		// skip instance with no tag set (no metrics)
		if len(m.fieldSet) == 0 {
			e.Logger.Debug().Msgf("skip instance (%s), no field set parsed", key)
		} else if r, err := m.Render(); err == nil {
			rendered = append(rendered, []byte(r))
			//logger.Debug(e.Prefix, "M= [%s%s%s]", color.Blue, r, color.End)
			count += countTmp
		} else {
			e.Logger.Debug().Msg(err.Error())
		}
	}

	e.Logger.Debug().Msgf("rendered %d measurements with %d data points for (%s)", len(rendered), count, object)

	// update metadata
	e.AddExportCount(count)
	if err := e.Metadata.LazySetValueUint64("count", "export", count); err != nil {
		e.Logger.Error().Stack().Err(err).Msg("metadata export count")
	}
	return rendered, nil
}
