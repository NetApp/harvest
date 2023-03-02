/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package influxdb

import (
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"testing"
)

func setupInfluxDB(t *testing.T, exporterName string) *InfluxDB {
	opts := &options.Options{}
	opts.Debug = true

	err := conf.LoadHarvestConfig("../../tools/doctor/testdata/testConfig.yml")
	if err != nil {
		panic(err)
	}
	e, ok := conf.Config.Exporters[exporterName]
	if !ok {
		t.Fatalf(`exporter (%v) not defined in config`, exporterName)
	}

	influx := &InfluxDB{AbstractExporter: exporter.New("InfluxDB", exporterName, opts, e, nil)}
	if err := influx.Init(); err != nil {
		t.Fatal(err)
	}

	return influx
}

// test that the addr (and port) parameters
// are handled properly to construct server URL
func TestAddrParameter(t *testing.T) {
	expectedURL := "http://localhost:8086/api/v2/write?org=netapp&bucket=harvest&precision=s"
	exporterName := "influx-test-addr"
	influx := setupInfluxDB(t, exporterName)

	if influx.url == expectedURL {
		t.Logf("OK - url: [%s]", expectedURL)
	} else {
		t.Fatalf("FAIL - expected [%s]\n                             got [%s]", expectedURL, influx.url)
	}
}

// test that the addr (and port) parameters
// are handled properly to construct server URL
func TestUrlParameter(t *testing.T) {
	expectedURL := "https://some-valid-domain-name.net:8888/api/v2/write?org=netapp&bucket=harvest&precision=s"
	exporterName := "influx-test-url"
	influx := setupInfluxDB(t, exporterName)

	if influx.url == expectedURL {
		t.Logf("OK - url: [%s]", expectedURL)
	} else {
		t.Fatalf("FAIL - expected [%s]\n       got [%s]", expectedURL, influx.url)
	}
}

// test that the addr, port and version parameters are handled properly to construct server URL
func TestVersionParameter(t *testing.T) {
	expectedURL := "http://localhost:8088/api/v4/write?org=harvest&bucket=harvest&precision=s"
	exporterName := "influx-test-version"
	influx := setupInfluxDB(t, exporterName)

	if influx.url == expectedURL {
		t.Logf("OK - url: [%s]", expectedURL)
	} else {
		t.Fatalf("FAIL - expected [%s]\n       got [%s]", expectedURL, influx.url)
	}
}

// test that `bucket`, `org`, `port`, and `precision` fields are ignored when using the `url` field
func TestUrlIgnores(t *testing.T) {
	expectedURL := "https://example.com:8086/api/v2/write?org=harvest&bucket=harvest&precision=s"
	exporterName := "influx-with-url"
	influx := setupInfluxDB(t, exporterName)

	if influx.url == expectedURL {
		t.Logf("OK - url: [%s]", expectedURL)
	} else {
		t.Fatalf("FAIL - expected [%s]\n       got [%s]", expectedURL, influx.url)
	}
}

// test rendering in debug mode
// this does not send to influxdb, but simply prints
// rendered data
func TestExportDebug(t *testing.T) {
	exporterName := "influx-test-url"
	influx := setupInfluxDB(t, exporterName)

	// matrix with fake data
	data := matrix.New("test_exporter", "influxd_test_data", "influxd_test_data")
	data.SetExportOptions(matrix.DefaultExportOptions())

	// add metric
	m, err := data.NewMetricInt64("test_metric")
	if err != nil {
		t.Fatal(err)
	}

	// add instance
	i, err := data.NewInstance("test_instance")
	if err != nil {
		t.Fatal(err)
	}
	i.SetLabel("test_label", "test_label_value")

	// add numeric data

	if err := m.SetValueInt64(i, 42); err != nil {
		t.Fatal(err)
	}

	// render data
	if err := influx.Export(data); err != nil {
		t.Fatal(err)
	}
}

// test that whitespace is escaped in the  parameters
// are handled properly to construct server URL
func TestWhiteSpaceInParameter(t *testing.T) {
	expectedURL := "http://localhost:8086/api/v2/write?org=harvest%202&bucket=harvest%20%2009&precision=s"
	exporterName := "influx-test-space"
	influx := setupInfluxDB(t, exporterName)

	if influx.url != expectedURL {
		t.Fatalf("FAIL - expected [%s]\n                             got [%s]", expectedURL, influx.url)
	}
}
