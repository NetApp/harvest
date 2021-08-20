/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package influxdb

import (
	"goharvest2/cmd/poller/exporter"
	"goharvest2/cmd/poller/options"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
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

// exporter should deal with situation when metric key
// conflicts with instance labels
func TestFieldKeyConflict(t *testing.T) {

	influx := setupInfluxDB("influx-test-url", t)

	data := matrix.New("collector", "object")
	exp := node.NewS("")
	exp.NewChildS("instance_keys", "").NewChildS("", "name")
	exp.NewChildS("instance_labels", "").NewChildS("", "status")
	data.SetExportOptions(exp)

	i, err := data.NewInstance("instance")
	if err != nil {
		t.Fatal(err)
	}
	i.SetLabel("name", "instance")
	i.SetLabel("status", "ok")

	m, err := data.NewMetricInt("status")
	if err != nil {
		t.Fatal(err)
	}

	if err := m.SetValueInt(i, 0); err != nil {
		t.Fatal(err)
	}

	// order can differ, since we use hash
	expected1 := `object,name=instance status="ok",status_num=0`
	expected2 := `object,name=instance status_num=0,status="ok"`

	if r, err := influx.Render(data); err != nil {
		t.Fatal(err)
	} else if len(r) != 1 {
		t.Errorf("expected 1 result, but got %d", len(r))
	} else if string(r[0]) != expected1 && string(r[0]) != expected2 {
		t.Errorf("metric [status] not renamed: %s", string(r[0]))
	} else {
		t.Logf("rendered correctly:\n%s", string(r[0]))
	}
}

// test rendering in debug mode
// this does not send to influxdb, but simply prints
// rendered data
func TestExportDebug(t *testing.T) {
	exporterName := "influx-test-url"
	influx := setupInfluxDB(exporterName, t)

	// matrix with fake data
	data := matrix.New("test_exporter", "influxd_test_data")
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
