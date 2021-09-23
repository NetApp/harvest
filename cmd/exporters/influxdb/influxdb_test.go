/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package influxdb

import (
	"goharvest2/cmd/poller/exporter"
	"goharvest2/cmd/poller/options"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/matrix"
	"testing"
)

func setupInfluxDB(exporterName string, t *testing.T) *InfluxDB {
	opts := &options.Options{}
	opts.Debug = true
	var exporters map[string]conf.Exporter
	var err error

	path := "../../tools/doctor/testdata/testConfig.yml"
	if exporters, err = conf.GetExporters2(path); err != nil {
		panic(err)
	}
	e, ok := exporters[exporterName]
	if !ok {
		t.Fatalf(`exporter (%v) not defined in config`, exporterName)
	}

	influx := &InfluxDB{AbstractExporter: exporter.New("InfluxDB", exporterName, opts, e)}
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
	influx := setupInfluxDB(exporterName, t)

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
	influx := setupInfluxDB(exporterName, t)

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
	influx := setupInfluxDB(exporterName, t)

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
	influx := setupInfluxDB(exporterName, t)

	// matrix with fake data
	data := matrix.New("test_exporter", "influxd_test_data", "influxd_test_data")
	data.SetExportOptions(matrix.DefaultExportOptions())

	// add metric
	m, err := data.NewMetricInt("test_metric")
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

	if err := m.SetValueInt(i, 42); err != nil {
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
	influx := setupInfluxDB(exporterName, t)

	if influx.url == expectedURL {
		t.Logf("OK - url: [%s]", expectedURL)
	} else {
		t.Fatalf("FAIL - expected [%s]\n                             got [%s]", expectedURL, influx.url)
	}
}

/* Uncomment to test against a running InfluxDB instance
   ! Edit the params values below
   ! Uncomment import "goharvest2/share/tree/node"
func TestExportProduction(t *testing.T) {

    logger.SetLevel(0)

    opts := &options.Options{}

    params := node.NewS("")
    params.NewChildS("addr", "")
    params.NewChildS("bucket", "")
    params.NewChildS("org", "")
    params.NewChildS("token", "")

    influx := New(exporter.New("InfluxDB", "influx-test", opts, params))

    if err := influx.Init(); err != nil {
        t.Fatal(err)
    }

    // matrix with fake data
    data := matrix.New("test_exporter", "influxd_test_data", "")
    data.SetExportOptions(matrix.DefaultExportOptions())

    // add metric
    m, err := data.AddMetric("test_metric", "test_metric", true)
    if err != nil {
        t.Fatal(err)
    }

    // add instance
    i, err := data.AddInstance("test_instance")
    if err != nil {
        t.Fatal(err)
    }
    i.Labels.Set("test_label", "test_label_value")

    // add numeric data
    if err := data.InitData(); err != nil {
        t.Fatal(err)
    }
    if err := data.SetValueString(m, i, "0"); err != nil {
        t.Fatal(err)
    }

    // render data
    if err := influx.Export(data); err != nil {
        t.Error(err)
    }
}
*/
