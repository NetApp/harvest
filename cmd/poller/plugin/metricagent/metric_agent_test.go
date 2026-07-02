/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package metricagent

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
	// "github.com/netapp/harvest/v2/share/logger"
)

func newAgent() *MetricAgent {

	// uncomment for debugging
	// logger.SetLevel(0)

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

	abc := plugin.New("Test", nil, params, nil, "", nil)
	p := &MetricAgent{AbstractPlugin: abc}

	if err := p.Init(conf.Remote{}); err != nil {
		panic(err)
	}
	return p
}

func TestComputeMetricsRuleWithSingleMatrix(t *testing.T) {

	var (
		instanceA, instanceB                                                *matrix.Instance
		metricAvail, metricUsed, metricTotal                                *matrix.Metric
		metricDiskP, metricDiskS, metricDiskH, metricDiskTotal              *matrix.Metric
		metricFiles, metricFilesUsed, metricFilesAvailable                  *matrix.Metric
		metricBytesPSector, metricSectorCount, metricTotalBytes             *matrix.Metric
		metricBytesTransferred, metricTotalDuration, metricTransmissionRate *matrix.Metric
		expected                                                            float64
		err                                                                 error
	)

	p := newAgent()
	m := matrix.New("TestLabelAgent", "test", "test")

	instanceA, err = m.NewInstance("A")
	assert.Nil(t, err)

	// space metric for addition
	metricAvail, err = m.NewMetricFloat64("space_available")
	assert.Nil(t, err)
	metricAvail.SetValueFloat64(instanceA, 1010101010)

	metricUsed, err = m.NewMetricFloat64("space_used")
	assert.Nil(t, err)
	metricUsed.SetValueFloat64(instanceA, 5050505050)

	// files metric for subtraction
	metricFiles, err = m.NewMetricFloat64("files")
	assert.Nil(t, err)
	metricFiles.SetValueFloat64(instanceA, 1024)

	metricFilesUsed, err = m.NewMetricFloat64("files_used")
	assert.Nil(t, err)
	metricFilesUsed.SetValueFloat64(instanceA, 216)

	instanceB, err = m.NewInstance("B")
	assert.Nil(t, err)

	// disk metric for addition
	metricDiskP, err = m.NewMetricFloat64("primary.disk_count")
	assert.Nil(t, err)
	metricDiskP.SetValueFloat64(instanceB, 8)

	metricDiskS, err = m.NewMetricFloat64("secondary.disk_count")
	assert.Nil(t, err)
	metricDiskS.SetValueFloat64(instanceB, 10)

	metricDiskH, err = m.NewMetricFloat64("hybrid.disk_count")
	assert.Nil(t, err)
	metricDiskH.SetValueFloat64(instanceB, 4)

	// bytes metric for multiplication
	metricBytesPSector, err = m.NewMetricFloat64("bytes_per_sector")
	assert.Nil(t, err)
	metricBytesPSector.SetValueFloat64(instanceB, 10000)

	metricSectorCount, err = m.NewMetricFloat64("sector_count")
	assert.Nil(t, err)
	metricSectorCount.SetValueFloat64(instanceB, 12)

	// transmission metric for division
	metricBytesTransferred, err = m.NewMetricFloat64("transfer.bytes_transferred")
	assert.Nil(t, err)
	metricBytesTransferred.SetValueFloat64(instanceB, 9000000)

	metricTotalDuration, err = m.NewMetricFloat64("transfer.total_duration")
	assert.Nil(t, err)
	metricTotalDuration.SetValueFloat64(instanceB, 3600)

	dataMap := make(map[string]*matrix.Matrix)
	dataMap[""] = m
	_, _, err = p.Run(dataMap)
	assert.Nil(t, err)

	// check "space_total" for instanceA
	expected = 6060606060
	metricTotal = m.GetMetric("space_total")
	assert.NotNil(t, metricTotal)
	metricTotalVal, ok := metricTotal.GetValueFloat64(instanceA)
	assert.True(t, ok)
	assert.Equal(t, metricTotalVal, expected)

	// check "disk_count" for instanceB
	expected = 22
	metricDiskTotal = m.GetMetric("disk_count")
	assert.NotNil(t, metricDiskTotal)
	metricDiskTotalVal, ok := metricDiskTotal.GetValueFloat64(instanceB)
	assert.True(t, ok)
	assert.Equal(t, metricDiskTotalVal, expected)

	// check "files_available" for instanceA
	expected = 808
	metricFilesAvailable = m.GetMetric("files_available")
	assert.NotNil(t, metricFilesAvailable)
	metricFilesAvailableVal, ok := metricFilesAvailable.GetValueFloat64(instanceA)
	assert.True(t, ok)
	assert.Equal(t, metricFilesAvailableVal, expected)

	// check "total_bytes" for instanceB
	expected = 120000
	metricTotalBytes = m.GetMetric("total_bytes")
	assert.NotNil(t, metricTotalBytes)
	metricTotalBytesVal, ok := metricTotalBytes.GetValueFloat64(instanceB)
	assert.True(t, ok)
	assert.Equal(t, metricTotalBytesVal, expected)

	// check "transmission_rate" for instanceB
	expected = 2500
	metricTransmissionRate = m.GetMetric("transmission_rate")
	metricTransmissionRateVal, ok := metricTransmissionRate.GetValueFloat64(instanceB)
	assert.True(t, ok)
	assert.Equal(t, metricTransmissionRateVal, expected)
}

func TestComputeMetricsRuleWithMultiMatrix(t *testing.T) {

	var (
		instanceA, instanceB                                   *matrix.Instance
		metricDataBytes, metricTotalSpaceBytes, metricUsedPerc *matrix.Metric
		expected                                               float64
		err                                                    error
	)

	params := node.NewS("MetricAgent")
	// create metric "storagegrid_storage_utilization_used_percent", which is percent of the metric value of storagegrid_storage_utilization_data_bytes by storagegrid_storage_utilization_total_space_bytes
	params.NewChildS("compute_metric", "").NewChildS("", "storagegrid_storage_utilization_used_percent PERCENT storagegrid_storage_utilization_data_bytes storagegrid_storage_utilization_total_space_bytes")
	abc := plugin.New("Test", nil, params, nil, "", nil)
	p := &MetricAgent{AbstractPlugin: abc}
	if err := p.Init(conf.Remote{}); err != nil {
		panic(err)
	}

	m1 := matrix.New("Matrix1", "", "Prometheus")
	m2 := matrix.New("Matrix2", "", "Prometheus")

	instanceA, err = m1.NewInstance("A")
	assert.Nil(t, err)

	instanceB, err = m2.NewInstance("B")
	assert.Nil(t, err)

	metricDataBytes, err = m1.NewMetricFloat64("storagegrid_storage_utilization_data_bytes")
	assert.Nil(t, err)
	metricDataBytes.SetValueFloat64(instanceA, 9000000)

	metricTotalSpaceBytes, err = m2.NewMetricFloat64("storagegrid_storage_utilization_total_space_bytes")
	assert.Nil(t, err)
	metricTotalSpaceBytes.SetValueFloat64(instanceB, 36000000)

	dataMap := make(map[string]*matrix.Matrix)
	dataMap["storagegrid_storage_utilization_data_bytes"] = m1
	dataMap["storagegrid_storage_utilization_total_space_bytes"] = m2
	_, _, err = p.Run(dataMap)
	assert.Nil(t, err)

	expected = 25
	metricUsedPerc = m1.GetMetric("storagegrid_storage_utilization_used_percent")
	metricUsedPercVal, ok := metricUsedPerc.GetValueFloat64(instanceA)
	assert.True(t, ok)
	assert.Equal(t, metricUsedPercVal, expected)
}
