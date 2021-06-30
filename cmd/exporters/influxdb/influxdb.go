/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package influxdb

import (
	"bytes"
	"fmt"
	"goharvest2/cmd/poller/exporter"
	"goharvest2/pkg/color"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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

	var (
		url, addr, bucket, org, token, version, precision *string
		port                                              *int
	)

	// check required / optional parameters

	if bucket = e.Params.Bucket; bucket == nil {
		return errors.New(errors.MISSING_PARAM, "bucket")
	}
	e.Logger.Debug().Msgf("using bucket [%s]", *bucket)

	if org = e.Params.Org; org == nil {
		return errors.New(errors.MISSING_PARAM, "org")
	}
	e.Logger.Debug().Msgf("using organization [%s]", *org)

	if token = e.Params.Token; token == nil {
		return errors.New(errors.MISSING_PARAM, "token")
	} else {
		e.token = *token
		e.Logger.Debug().Msg("will use authorization with api token")
	}

	if version = e.Params.Version; version == nil {
		v := defaultApiVersion
		version = &v
	}
	e.Logger.Debug().Msgf("using api version [%s]", *version)

	if precision = e.Params.Precision; precision == nil {
		p := defaultApiPrecision
		precision = &p
	}
	e.Logger.Debug().Msgf("using api precision [%s]", *precision)

	// user should provide either url or addr
	// url is expected to be the full write URL with all query params specified (optionally with scheme)
	// addr is expected to include host only (no scheme, no port)
	if url = e.Params.Url; url == nil {
		if addr = e.Params.Addr; addr == nil {
			return errors.New(errors.MISSING_PARAM, "url or addr")
		}

		if port = e.Params.Port; port == nil {
			e.Logger.Debug().Msgf("using default port [%d]", defaultPort)
			defPort := defaultPort
			port = &defPort
		}

		urlToUSe := "http://" + *addr + ":" + strconv.Itoa(*port)
		url = &urlToUSe
		e.url = fmt.Sprintf("%s/api/v%s/write?org=%s&bucket=%s&precision=%s", *url, *version, *org, *bucket, *precision)
	} else {
		e.url = *url
		/* Example url: http://localhost:8088/api/v4/write?org=harvest&bucket=harvest&precision=s
		   step 1: localhost:8088/api/v4/write?org=harvest&bucket=harvest&precision=s
		   step 2: localhost:8088 value in addrAndPort var
		   step 3: addr has localhost and port has 8088
		*/
		addrAndPort := strings.Split(strings.Split(e.url, "//")[1], "/")[0]
		res := strings.Split(addrAndPort, ":")
		addr = &res[0]
		if p, err := strconv.Atoi(res[1]); err != nil {
			e.Logger.Warn().Msgf("invalid port [%s]", p)
		} else {
			port = &p
		}
	}

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

	e.Logger.Debug().Msgf("url= [%s]", e.url)

	// construct HTTP client
	e.client = &http.Client{Timeout: timeout}

	e.Logger.Debug().Msgf("initialized exporter, ready to emit to [%s:%s]", *addr, *port)
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
			// otherwise to the actual export: send to the DB
		} else if err = e.Emit(metrics); err != nil {
			e.Logger.Error().Stack().Err(err).Msgf("(%s.%s) --> %s", data.Object, data.UUID)
			return err
		}
	}

	e.Logger.Debug().Msgf("(%s.%s) --> exported %d data points", data.Object, data.UUID, len(metrics))

	// update metadata
	if err = e.Metadata.LazyAddValueInt64("time", "export", time.Since(s).Microseconds()); err != nil {
		e.Logger.Error().Stack().Err(err).Msg("metadata export time")
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
			e.Logger.Debug().Msgf("skip instance (%s), no tag set parsed from labels (%v)", key, instance.GetLabels().Map())
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

		e.Logger.Trace().Msgf("rendering from: %s", m.String())

		// skip instance with no tag set (no metrics)
		if len(m.field_set) == 0 {
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
