/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package servicecontrol

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/servicecontrol/v1"
	"google.golang.org/api/transport"
)

/* Write metrics to the ServiceControl Time-Series Database.
   The exporter follows ServiceControl's v2 documentation
   for authentication and line protocol (measurement):

   - https://docs.influxdata.com/influxdb/v2.0/write-data/developer-tools/api/
   - https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/

*/

const (
	defaultPort          = 8086
	defaultTimeout       = 5
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

type ServiceControl struct {
	*exporter.AbstractExporter
	url            string
	serviceName    string
	serviceControl *servicecontrol.Service
}

func New(abc *exporter.AbstractExporter) exporter.Exporter {
	return &ServiceControl{AbstractExporter: abc}
}

func (e *ServiceControl) Init() error {

	if err := e.InitAbc(); err != nil {
		return err
	}

	e.url = *e.Params.URL
	e.serviceName = *e.Params.ServiceName

	ctx := context.Background()

	httpClient, _, err := transport.NewHTTPClient(ctx, option.WithTokenSource(google.ComputeTokenSource("", servicecontrol.ServicecontrolScope)))
	if err != nil {
		e.Logger.Error("Failed to create http client for creating service control: %v", err)
		return err
	}

	e.serviceControl, err = servicecontrol.NewService(ctx,
		option.WithHTTPClient(httpClient),
		option.WithEndpoint(e.url),
		option.WithScopes(servicecontrol.ServicecontrolScope),
	)
	if err != nil {
		e.Logger.Error("Failed to create service control client: %v", err)
		return err
	}

	e.Logger.Info("Created client with workload identity: rootUrl: %s", e.url)

	return nil

}

func (e *ServiceControl) Export(data *matrix.Matrix) (exporter.Stats, error) {

	if data.Identifier != "Volume" {
		return exporter.Stats{}, nil
	}

	var (
		metrics []*servicecontrol.Operation
		err     error
		s       time.Time
		stats   exporter.Stats
	)

	e.Lock()
	defer e.Unlock()

	s = time.Now()

	// render the metrics, i.e. convert to InfluxDb line protocol
	if metrics, stats, err = e.Render(data); err == nil && len(metrics) != 0 {
		// fix render time
		if err = e.Metadata.LazyAddValueInt64("time", "render", time.Since(s).Microseconds()); err != nil {
			e.Logger.Error("metadata render time", slogx.Err(err))
		}
		// in test mode, don't emit metrics
		if e.Options.IsTest {
			return stats, nil
			// otherwise, to the actual export: send to the DB
		} else if err = e.Emit(metrics); err != nil {
			return stats, fmt.Errorf("unable to emit object: %s, uuid: %s, err=%w", data.Object, data.UUID, err)
		}
	}

	e.Logger.Debug(
		"exported",
		slog.String("object", data.Object),
		slog.String("uuid", data.UUID),
		slog.Int("numMetric", len(metrics)),
	)

	return stats, nil
}

func (e *ServiceControl) Emit(data []*servicecontrol.Operation) error {

	req := &servicecontrol.ReportRequest{
		Operations: data,
	}

	resp, err := e.serviceControl.Services.Report(e.serviceName, req).Do()
	if err != nil {
		e.Logger.Error("Failed to report operations to service control", slogx.Err(err))
		return err
	}
	e.Logger.Info("%v", resp)

	return nil
}

func (e *ServiceControl) Render(data *matrix.Matrix) ([]*servicecontrol.Operation, exporter.Stats, error) {

	var (
		count, countTmp, instancesExported uint64
	)

	rendered := make([]*servicecontrol.Operation, 0)

	object := data.Object
	if data.Identifier == "Volume" {
		fmt.Printf("Volume\n")
	} else {
		return nil, exporter.Stats{}, nil
	}

	// user-defined preferences for export
	var labelsToInclude, keysToInclude []string
	includeAll := data.GetExportOptions().GetChildContentS("include_all_labels") == "true"
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
	for key, value := range data.GetGlobalLabels() {
		global.AddTag(key, value)
	}

	// render one measurement for each instance
	for key, instance := range data.GetInstances() {

		countTmp = 0

		if !instance.IsExportable() {
			continue
		}

		instancesExported++

		m := NewMeasurement(object, len(global.tagSet))
		copy(m.tagSet, global.tagSet)

		// tag set
		if includeAll {
			for label, value := range instance.GetLabels() {
				if value != "" {
					m.AddTag(label, value)
				}
			}
		} else {
			for _, key := range keysToInclude {
				if value, has := instance.GetLabels()[key]; has && value != "" {
					m.AddTag(key, value)
				}
			}
		}

		// skip instance without key tags
		if len(m.tagSet) == 0 {
			e.Logger.Debug(
				"skip instance, no tag set parsed from labels",
				slog.String("key", key),
				slog.Any("labels", instance.GetLabels()),
			)
		}

		// field set

		// strings
		for _, label := range labelsToInclude {
			if value, has := instance.GetLabels()[label]; has && value != "" {
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
				for _, label := range metric.GetLabels() {
					fieldName += "_" + label
				}
			}

			if rename, has := protectedFieldNames[fieldName]; has {
				fieldName = rename
			}

			m.AddField(fieldName, value)
			countTmp++
		}

		// skip instance with no tag set (no metrics)
		if len(m.fieldSet) == 0 {
			e.Logger.Debug("skip instance, no field set parsed", slog.Any("instance", instance))
		} else if r, err := m.Render(); err == nil {
			rendered = append(rendered, r)
			count += countTmp
		} else {
			e.Logger.Debug(err.Error())
		}
	}

	e.Logger.Debug("rendered", slog.Int("instances", len(rendered)), slog.Uint64("metrics", count))

	// update metadata
	e.AddExportCount(count)
	if err := e.Metadata.LazySetValueUint64("count", "export", count); err != nil {
		e.Logger.Error("metadata export count", slogx.Err(err))
	}
	return rendered, exporter.Stats{InstancesExported: instancesExported, MetricsExported: count}, nil
}
