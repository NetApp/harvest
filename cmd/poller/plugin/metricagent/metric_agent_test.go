/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package metricagent

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
	//"github.com/netapp/harvest/v2/share/logger"
)

var p *MetricAgent

func TestInitPlugin(t *testing.T) {

	// uncomment for debugging
	//logger.SetLevel(0)

	// define plugin rules
	params := node.NewS("MetricAgent")
	// create metric "space_total", which is addition of the metric value of space_available and space_used
	params.NewChildS("compute_metric", "").NewChildS("", "space_total ADD space_available space_used")
	// create metric "disk_count", which is addition of the metric value of primary.disk_count, secondary.disk_count and hybrid.disk_count
	params.NewChildS("compute_metric", "").NewChildS("", "disk_count ADD primary.disk_count secondary.disk_count hybrid.disk_count")
	// create metric "files_available", which is subtraction of the metric value of files_used from metric value of files
	params.NewChildS("compute_metric", "").NewChildS("", "files_available SUBTRACT files files_used")
	// create metric "total_bytes", which is multiplication of the metric value of bytes_per_sector and sector_count
	params.NewChildS("compute_metric", "").NewChildS("", "total_bytes MULTIPLY bytes_per_sector sector_count")
	// create metric "transmission_rate", which is division of the metric value of transfer.bytes_transferred by transfer.total_duration
	params.NewChildS("compute_metric", "").NewChildS("", "transmission_rate DIVIDE transfer.bytes_transferred transfer.total_duration")

	abc := plugin.New("Test", nil, params, nil, "")
	p = &MetricAgent{AbstractPlugin: abc}

	if err := p.Init(); err != nil {
		t.Fatal(err)
	}
}

func TestComputeMetricsRule(t *testing.T) {

	var (
		instanceA, instanceB                                                *matrix.Instance
		metricAvail, metricUsed, metricTotal                                matrix.Metric
		metricDiskP, metricDiskS, metricDiskH, metricDiskTotal              matrix.Metric
		metricFiles, metricFilesUsed, metricFilesAvailable                  matrix.Metric
		metricBytesPSector, metricSectorCount, metricTotalBytes             matrix.Metric
		metricBytesTransferred, metricTotalDuration, metricTransmissionRate matrix.Metric
		expected                                                            float64
		err                                                                 error
	)

	m := matrix.New("TestLabelAgent", "test", "test")

	if instanceA, err = m.NewInstance("A"); err != nil {
		t.Fatal(err)
	}

	// space metric for addition
	if metricAvail, err = m.NewMetricFloat64("space_available"); err == nil {
		_ = metricAvail.SetValueFloat64(instanceA, 1010101010)
	} else {
		t.Error("metric [space_available]  not created for InstanceA")
	}
	if metricUsed, err = m.NewMetricFloat64("space_used"); err == nil {
		_ = metricUsed.SetValueFloat64(instanceA, 5050505050)
	} else {
		t.Error("metric [space_used]  not created for InstanceA")
	}
	// files metric for subtraction
	if metricFiles, err = m.NewMetricFloat64("files"); err == nil {
		_ = metricFiles.SetValueFloat64(instanceA, 1024)
	} else {
		t.Error("metric [files]  not created for InstanceA")
	}
	if metricFilesUsed, err = m.NewMetricFloat64("files_used"); err == nil {
		_ = metricFilesUsed.SetValueFloat64(instanceA, 216)
	} else {
		t.Error("metric [files_used]  not created for InstanceA")
	}

	if instanceB, err = m.NewInstance("B"); err != nil {
		t.Fatal(err)
	}

	// disk metric for addition
	if metricDiskP, err = m.NewMetricFloat64("primary.disk_count"); err == nil {
		_ = metricDiskP.SetValueFloat64(instanceB, 8)
	} else {
		t.Error("metric [primary.disk_count]  not created for InstanceB")
	}
	if metricDiskS, err = m.NewMetricFloat64("secondary.disk_count"); err == nil {
		_ = metricDiskS.SetValueFloat64(instanceB, 10)
	} else {
		t.Error("metric [secondary.disk_count]  not created for InstanceB")
	}
	if metricDiskH, err = m.NewMetricFloat64("hybrid.disk_count"); err == nil {
		_ = metricDiskH.SetValueFloat64(instanceB, 4)
	} else {
		t.Error("metric [hybrid.disk_count]  not created for InstanceB")
	}
	// bytes metric for multiplication
	if metricBytesPSector, err = m.NewMetricFloat64("bytes_per_sector"); err == nil {
		_ = metricBytesPSector.SetValueFloat64(instanceB, 10000)
	} else {
		t.Error("metric [bytes_per_sector]  not created for InstanceB")
	}
	if metricSectorCount, err = m.NewMetricFloat64("sector_count"); err == nil {
		_ = metricSectorCount.SetValueFloat64(instanceB, 12)
	} else {
		t.Error("metric [sector_count]  not created for InstanceB")
	}
	// transmission metric for division
	if metricBytesTransferred, err = m.NewMetricFloat64("transfer.bytes_transferred"); err == nil {
		_ = metricBytesTransferred.SetValueFloat64(instanceB, 9000000)
	} else {
		t.Error("metric [transfer.bytes_transferred]  not created for InstanceB")
	}
	if metricTotalDuration, err = m.NewMetricFloat64("transfer.total_duration"); err == nil {
		_ = metricTotalDuration.SetValueFloat64(instanceB, 3600)
	} else {
		t.Error("metric [transfer.total_duration]  not created for InstanceB")
	}

	if err = p.computeMetrics(m); err != nil {
		t.Fatal(err)
	}

	// check "space_total" for instanceA
	expected = 6060606060
	if metricTotal = m.GetMetric("space_total"); metricTotal != nil {
		if metricTotalVal, ok, pass := metricTotal.GetValueFloat64(instanceA); !ok || !pass {
			t.Error("metric [space_total]: value for InstanceA not set")
		} else if metricTotalVal != expected {
			t.Errorf("metric [space_total]: value for InstanceA is %f, expected %f", metricTotalVal, expected)
		} else {
			t.Logf("OK - metric [space_total]: value for instanceA set to %f", metricTotalVal)
		}
	} else {
		t.Error("metric [space_total] missing")
	}

	// check "disk_count" for instanceB
	expected = 22
	if metricDiskTotal = m.GetMetric("disk_count"); metricDiskTotal != nil {
		if metricDiskTotalVal, ok, pass := metricDiskTotal.GetValueFloat64(instanceB); !ok || !pass {
			t.Error("metric [disk_count]: value for InstanceB not set")
		} else if metricDiskTotalVal != expected {
			t.Errorf("metric [disk_count]: value for InstanceB is %f, expected %f", metricDiskTotalVal, expected)
		} else {
			t.Logf("OK - metric [disk_count]: value for instanceB set to %f", metricDiskTotalVal)
		}
	} else {
		t.Error("metric [disk_count] missing")
	}

	// check "files_available" for instanceA
	expected = 808
	if metricFilesAvailable = m.GetMetric("files_available"); metricFilesAvailable != nil {
		if metricFilesAvailableVal, ok, pass := metricFilesAvailable.GetValueFloat64(instanceA); !ok || !pass {
			t.Error("metric [files_available]: value for InstanceA not set")
		} else if metricFilesAvailableVal != expected {
			t.Errorf("metric [files_available]: value for InstanceA is %f, expected %f", metricFilesAvailableVal, expected)
		} else {
			t.Logf("OK - metric [files_available]: value for instanceA set to %f", metricFilesAvailableVal)
		}
	} else {
		t.Error("metric [files_available] missing")
	}

	// check "total_bytes" for instanceB
	expected = 120000
	if metricTotalBytes = m.GetMetric("total_bytes"); metricTotalBytes != nil {
		if metricTotalBytesVal, ok, pass := metricTotalBytes.GetValueFloat64(instanceB); !ok || !pass {
			t.Error("metric [total_bytes]: value for InstanceB not set")
		} else if metricTotalBytesVal != expected {
			t.Errorf("metric [total_bytes]: value for InstanceB is %f, expected %f", metricTotalBytesVal, expected)
		} else {
			t.Logf("OK - metric [total_bytes]: value for instanceB set to %f", metricTotalBytesVal)
		}
	} else {
		t.Error("metric [total_bytes] missing")
	}

	// check "transmission_rate" for instanceB
	expected = 2500
	if metricTransmissionRate = m.GetMetric("transmission_rate"); metricTransmissionRate != nil {
		if metricTransmissionRateVal, ok, pass := metricTransmissionRate.GetValueFloat64(instanceB); !ok || !pass {
			t.Error("metric [transmission_rate]: value for InstanceB not set")
		} else if metricTransmissionRateVal != expected {
			t.Errorf("metric [transmission_rate]: value for InstanceB is %f, expected %f", metricTransmissionRateVal, expected)
		} else {
			t.Logf("OK - metric [transmission_rate]: value for instanceB set to %f", metricTransmissionRateVal)
		}
	} else {
		t.Error("metric [transmission_rate] missing")
	}

}
