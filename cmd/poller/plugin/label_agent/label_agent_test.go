/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package label_agent

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
	//"github.com/netapp/harvest/v2/share/logger"
)

var p *LabelAgent

func TestInitPlugin(t *testing.T) {

	// uncomment for debugging
	//logger.SetLevel(0)

	// define plugin rules
	params := node.NewS("LabelAgent")
	// split value of "X" into 4 and take last 2 as new labels
	params.NewChildS("split", "").NewChildS("", "X `/` ,,C,D")
	// split value of "X" and take first 2 as new labels
	params.NewChildS("split_regex", "").NewChildS("", "X `.*(A\\d+)_(B\\d+)` A,B")
	// split value of "X" into key-value pairs
	params.NewChildS("split_pairs", "").NewChildS("", "X ` ` `:`")
	// join values of "A" and "B" and set as label "X"
	params.NewChildS("join", "").NewChildS("", "X `_` A,B")
	// replace "aaa_" with "bbb_" and set as label "B"
	params.NewChildS("replace", "").NewChildS("", "A B `aaa_` `bbb_`")
	// remove occurences of "aaa_" from value of "A"
	params.NewChildS("replace", "").NewChildS("", "A A `aaa_` ``")
	// reverse the order of matching elements, replace underscore with dash and "aaa" with "bbb"
	params.NewChildS("replace_regex", "").NewChildS("", "A B `^(aaa)_(\\d+)_(\\w+)$` `$3-$2-bbb`")
	// exclude instance if label "A" has value "aaa bbb ccc"
	params.NewChildS("exclude_equals", "").NewChildS("", "A `aaa bbb ccc`")
	// exclude instance if label "A" contains "_aaa_"
	params.NewChildS("exclude_contains", "").NewChildS("", "A `_aaa_`")
	// exclude instance of label value has prefix "aaa_" followed by at least one digit
	params.NewChildS("exclude_regex", "").NewChildS("", "A `^aaa_\\d+$`")
	// include instance if label "A" has value "aaa bbb ccc"
	params.NewChildS("include_equals", "").NewChildS("", "A `aaa bbb ccc`")
	// include instance if label "A" contains "_aaa_"
	params.NewChildS("include_contains", "").NewChildS("", "A `_aaa_`")
	// include instance of label value has prefix "aaa_" followed by at least one digit
	params.NewChildS("include_regex", "").NewChildS("", "A `^aaa_\\d+$`")
	// create metric "new_status", if label "state" is one of the up/ok[zapi/rest], map metric value to respective index
	params.NewChildS("value_to_num", "").NewChildS("", "new_status state up ok")
	// create metric "new_stage", but if none of the values is matching, use default value "4"
	params.NewChildS("value_to_num", "").NewChildS("", "new_stage stage init start `4`")
	// create metric "new_outage", if empty value is expected and non empty means wrong, use default value "0"
	params.NewChildS("value_to_num", "").NewChildS("", "new_outage outage - - `0`")
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
	// create metric "status", if label "state" contains one of the up/ok[zapi/rest], map metric value to respective index
	params.NewChildS("value_to_num_regex", "").NewChildS("", "status state ^up$ ^ok$ `0`")
	// create metric "output", if label "stage" matches regex then map to 1 else use default value "4"
	params.NewChildS("value_to_num_regex", "").NewChildS("", "output stage ^transfer.*$ ^run.*$ `4`")
	// create metric "result", if label "state" matches regex then map to 1 else use default value "4"
	params.NewChildS("value_to_num_regex", "").NewChildS("", "result value ^test\\d+ ^error `4`")

	// These both are mutually exclusive, and should honour the above one's filtered result.
	// exclude instance if label "volstate" has value "offline"
	params.NewChildS("exclude_equals", "").NewChildS("", "volstate `offline`")
	// include instance if label "voltype" has value "rw"
	params.NewChildS("include_equals", "").NewChildS("", "voltype `rw`")
	// include instance if label "volstyle" has value which starts with "flexvol_"
	params.NewChildS("include_contains", "").NewChildS("", "volstyle `flexvol`")
	// exclude instance if label "volstatus" has value which starts with "stopped_"
	params.NewChildS("exclude_contains", "").NewChildS("", "volstatus `stop`")

	abc := plugin.New("Test", nil, params, nil, "")
	p = &LabelAgent{AbstractPlugin: abc}

	if err := p.Init(); err != nil {
		t.Fatal(err)
	}
}

func TestSplitSimpleRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	instance, _ := m.NewInstance("0")
	instance.SetLabel("X", "a/b/c/d")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.splitSimple(m)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("C") == "c" && instance.GetLabel("D") == "d" {
		// OK
	} else {
		t.Error("Labels C and D don't have expected values")
	}
}

func TestSplitRegexRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	instance, _ := m.NewInstance("0")
	instance.SetLabel("X", "xxxA22_B333")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.splitRegex(m)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("A") == "A22" && instance.GetLabel("B") == "B333" {
		// OK
	} else {
		t.Error("Labels A and B don't have expected values")
	}
}

func TestSplitPairsRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	instance, _ := m.NewInstance("0")
	instance.SetLabel("X", "owner:jack contact:some@email")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.splitPairs(m)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("owner") == "jack" && instance.GetLabel("contact") == "some@email" {
		// OK
	} else {
		t.Error("Labels owner and contact don't have expected values")
	}
}

func TestJoinSimpleRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	instance, _ := m.NewInstance("0")
	instance.SetLabel("A", "aaa")
	instance.SetLabel("B", "bbb")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.joinSimple(m)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("X") == "aaa_bbb" {
		// OK
	} else {
		t.Error("Label A does have expected value")
	}
}

func TestReplaceSimpleRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	instance, _ := m.NewInstance("0")
	instance.SetLabel("A", "aaa_X")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.replaceSimple(m)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("A") == "X" && instance.GetLabel("B") == "bbb_X" {
		// OK
	} else {
		t.Error("Labels A and B don't have expected values")
	}
}

func TestReplaceRegexRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	instance, _ := m.NewInstance("0")
	instance.SetLabel("A", "aaa_12345_abcDEF")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.replaceRegex(m)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("B") == "abcDEF-12345-bbb" {
		// OK
	} else {
		t.Error("Label B does not have expected value")
	}
}

func TestExcludeEqualsRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "aaa bbb ccc")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "aaa bbb")
	instanceNo.SetLabel("B", "aaa bbb ccc")

	p.excludeEquals(m)

	if instanceYes.IsExportable() {
		t.Error("InstanceYes should have been excluded")
	}

	if !instanceNo.IsExportable() {
		t.Error("instanceNo should not have been excluded")
	}
}

func TestExcludeContainsRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "xxx_aaa_xxx")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "_aaa")

	p.excludeContains(m)

	if instanceYes.IsExportable() {
		t.Error("InstanceYes should have been excluded")
	}

	if !instanceNo.IsExportable() {
		t.Error("instanceNo should not have been excluded")
	}
}

func TestExcludeRegexRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "aaa_123")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "aaa_123!")

	p.excludeRegex(m)

	if instanceYes.IsExportable() {
		t.Error("InstanceYes should have been excluded")
	}

	if !instanceNo.IsExportable() {
		t.Error("instanceNo should not have been excluded")
	}
}

func TestIncludeEqualsRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "aaa bbb ccc")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "aaa bbb")
	instanceNo.SetLabel("B", "aaa bbb ccc")

	p.includeEquals(m)

	if !instanceYes.IsExportable() {
		t.Error("InstanceYes should have been included")
	}

	if instanceNo.IsExportable() {
		t.Error("instanceNo should not have been included")
	}
}

func TestIncludeContainsRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "xxx_aaa_xxx")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "_aaa")

	p.includeContains(m)

	if !instanceYes.IsExportable() {
		t.Error("InstanceYes should have been included")
	}

	if instanceNo.IsExportable() {
		t.Error("instanceNo should not have been included")
	}
}

func TestIncludeRegexRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "aaa_123")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "aaa_123!")

	p.includeRegex(m)

	if !instanceYes.IsExportable() {
		t.Error("InstanceYes should have been included")
	}

	if instanceNo.IsExportable() {
		t.Error("instanceNo should not have been included")
	}
}

func TestValueToNumRule(t *testing.T) {

	var (
		instanceA, instanceB  *matrix.Instance
		status, stage, outage matrix.Metric
		v, expected           uint8
		ok                    bool
		err                   error
	)
	// should match
	m := matrix.New("TestLabelAgent", "test", "test")

	if instanceA, err = m.NewInstance("A"); err != nil {
		t.Fatal(err)
	}
	instanceA.SetLabel("state", "up")   // "status" should be 1
	instanceA.SetLabel("stage", "init") // "stage" should be 1
	instanceA.SetLabel("outage", "")    // "outageStatus" should be 1

	if instanceB, err = m.NewInstance("B"); err != nil {
		t.Fatal(err)
	}
	instanceB.SetLabel("state", "unknown") // "status" should not be set
	instanceB.SetLabel("stage", "unknown") // "stage" should be 4 (default)
	instanceB.SetLabel("outage", "failed") // "outage" should be 0 (default)

	if err = p.mapValueToNum(m); err != nil {
		t.Fatal(err)
	}

	if status = m.GetMetric("new_status"); status == nil {
		t.Error("metric [status] missing")
	}

	if stage = m.GetMetric("new_stage"); stage == nil {
		t.Error("metric [stage] missing")
	}

	if outage = m.GetMetric("new_outage"); outage == nil {
		t.Error("metric [outage] missing")
	}

	// check "status" for instanceA
	expected = 1
	if v, ok = status.GetValueUint8(instanceA); !ok {
		t.Error("metric [status]: value for InstanceA not set")
	} else if v != expected {
		t.Errorf("metric [status]: value for InstanceA is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [status]: value for instanceA set to %d", v)
	}

	// check "status" for instanceB
	if v, ok = status.GetValueUint8(instanceB); !ok {
		t.Log("OK - metric [status]: value for InstanceB not set")
	} else {
		t.Errorf("metric [status]: value for InstanceB is %d, should not be set", v)
	}

	// check "stage" for instanceA
	expected = 1
	if v, ok = stage.GetValueUint8(instanceA); !ok {
		t.Error("metric [stage]: value for InstanceA not set")
	} else if v != expected {
		t.Errorf("metric [stage]: value for InstanceA is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [stage]: value for instanceA set to %d", v)
	}

	// check "stage" for instanceB
	expected = 4
	if v, ok = stage.GetValueUint8(instanceB); !ok {
		t.Error("metric [stage]: value for InstanceB not set")
	} else if v != expected {
		t.Errorf("metric [stage]: value for InstanceB is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [stage]: value for instanceB set to %d", v)
	}

	// check "outage" for instanceA
	expected = 1
	if v, ok = outage.GetValueUint8(instanceA); !ok {
		t.Error("metric [outage]: value for InstanceA not set")
	} else if v != expected {
		t.Errorf("metric [outage]: value for InstanceA is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [outage]: value for instanceA set to %d", v)
	}

	// check "outage" for instanceB
	expected = 0
	if v, ok = outage.GetValueUint8(instanceB); !ok {
		t.Error("metric [outage]: value for InstanceB not set")
	} else if v != expected {
		t.Errorf("metric [outage]: value for InstanceB is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [outage]: value for instanceB set to %d", v)
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
		metricAvail.SetValueFloat64(instanceA, 1010101010)
	} else {
		t.Error("metric [space_available]  not created for InstanceA")
	}
	if metricUsed, err = m.NewMetricFloat64("space_used"); err == nil {
		metricUsed.SetValueFloat64(instanceA, 5050505050)
	} else {
		t.Error("metric [space_used]  not created for InstanceA")
	}
	// files metric for subtraction
	if metricFiles, err = m.NewMetricFloat64("files"); err == nil {
		metricFiles.SetValueFloat64(instanceA, 1024)
	} else {
		t.Error("metric [files]  not created for InstanceA")
	}
	if metricFilesUsed, err = m.NewMetricFloat64("files_used"); err == nil {
		metricFilesUsed.SetValueFloat64(instanceA, 216)
	} else {
		t.Error("metric [files_used]  not created for InstanceA")
	}

	if instanceB, err = m.NewInstance("B"); err != nil {
		t.Fatal(err)
	}

	// disk metric for addition
	if metricDiskP, err = m.NewMetricFloat64("primary.disk_count"); err == nil {
		metricDiskP.SetValueFloat64(instanceB, 8)
	} else {
		t.Error("metric [primary.disk_count]  not created for InstanceB")
	}
	if metricDiskS, err = m.NewMetricFloat64("secondary.disk_count"); err == nil {
		metricDiskS.SetValueFloat64(instanceB, 10)
	} else {
		t.Error("metric [secondary.disk_count]  not created for InstanceB")
	}
	if metricDiskH, err = m.NewMetricFloat64("hybrid.disk_count"); err == nil {
		metricDiskH.SetValueFloat64(instanceB, 4)
	} else {
		t.Error("metric [hybrid.disk_count]  not created for InstanceB")
	}
	// bytes metric for multiplication
	if metricBytesPSector, err = m.NewMetricFloat64("bytes_per_sector"); err == nil {
		metricBytesPSector.SetValueFloat64(instanceB, 10000)
	} else {
		t.Error("metric [bytes_per_sector]  not created for InstanceB")
	}
	if metricSectorCount, err = m.NewMetricFloat64("sector_count"); err == nil {
		metricSectorCount.SetValueFloat64(instanceB, 12)
	} else {
		t.Error("metric [sector_count]  not created for InstanceB")
	}
	// transmission metric for division
	if metricBytesTransferred, err = m.NewMetricFloat64("transfer.bytes_transferred"); err == nil {
		metricBytesTransferred.SetValueFloat64(instanceB, 9000000)
	} else {
		t.Error("metric [transfer.bytes_transferred]  not created for InstanceB")
	}
	if metricTotalDuration, err = m.NewMetricFloat64("transfer.total_duration"); err == nil {
		metricTotalDuration.SetValueFloat64(instanceB, 3600)
	} else {
		t.Error("metric [transfer.total_duration]  not created for InstanceB")
	}

	if err = p.computeMetrics(m); err != nil {
		t.Fatal(err)
	}

	// check "space_total" for instanceA
	expected = 6060606060
	if metricTotal = m.GetMetric("space_total"); metricTotal != nil {
		if metricTotalVal, ok := metricTotal.GetValueFloat64(instanceA); !ok {
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
		if metricDiskTotalVal, ok := metricDiskTotal.GetValueFloat64(instanceB); !ok {
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
		if metricFilesAvailableVal, ok := metricFilesAvailable.GetValueFloat64(instanceA); !ok {
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
		if metricTotalBytesVal, ok := metricTotalBytes.GetValueFloat64(instanceB); !ok {
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
		if metricTransmissionRateVal, ok := metricTransmissionRate.GetValueFloat64(instanceB); !ok {
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

func TestValueToNumRegexRule(t *testing.T) {

	var (
		instanceA, instanceB   *matrix.Instance
		status, output, result matrix.Metric
		v, expected            uint8
		ok                     bool
		err                    error
	)

	m := matrix.New("TestLabelAgent", "test", "test")

	if instanceA, err = m.NewInstance("A"); err != nil {
		t.Fatal(err)
	}
	instanceA.SetLabel("state", "up")      // "status" should be 1
	instanceA.SetLabel("stage", "stopped") // "output" should be 4 (default)
	instanceA.SetLabel("value", "test11")  // "result" should be 1

	if instanceB, err = m.NewInstance("B"); err != nil {
		t.Fatal(err)
	}
	instanceB.SetLabel("state", "error")   // "status" should be 0 (default)
	instanceB.SetLabel("stage", "running") // "output" should be 1
	instanceB.SetLabel("value", "done")    // "result" should be 4 (default)

	if err = p.mapValueToNumRegex(m); err != nil {
		t.Fatal(err)
	}

	if status = m.GetMetric("status"); status == nil {
		t.Error("metric [status] missing")
	}

	if output = m.GetMetric("output"); output == nil {
		t.Error("metric [output] missing")
	}

	if result = m.GetMetric("result"); result == nil {
		t.Error("metric [result] missing")
	}

	// check "status" for instanceA
	expected = 1
	if v, ok = status.GetValueUint8(instanceA); !ok {
		t.Error("metric [status]: value for InstanceA not set")
	} else if v != expected {
		t.Errorf("metric [status]: value for InstanceA is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [status]: value for instanceA set to %d", v)
	}

	// check "status" for instanceB
	expected = 0
	if v, ok = status.GetValueUint8(instanceB); !ok {
		t.Log("metric [status]: value for InstanceB not set")
	} else if v != expected {
		t.Errorf("metric [status]: value for InstanceB is %d, expected %d", v, expected)
	} else {
		t.Logf("Ok - metric [status]: value for InstanceB set to %d", v)
	}

	// check "output" for instanceA
	expected = 4
	if v, ok = output.GetValueUint8(instanceA); !ok {
		t.Error("metric [stage]: value for InstanceA not set")
	} else if v != expected {
		t.Errorf("metric [stage]: value for InstanceA is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [stage]: value for instanceA set to %d", v)
	}

	// check "output" for instanceB
	expected = 1
	if v, ok = output.GetValueUint8(instanceB); !ok {
		t.Error("metric [stage]: value for InstanceB not set")
	} else if v != expected {
		t.Errorf("metric [stage]: value for InstanceB is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [stage]: value for instanceB set to %d", v)
	}

	// check "result" for instanceA
	expected = 1
	if v, ok = result.GetValueUint8(instanceA); !ok {
		t.Error("metric [outage]: value for InstanceA not set")
	} else if v != expected {
		t.Errorf("metric [outage]: value for InstanceA is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [outage]: value for instanceA set to %d", v)
	}

	// check "result" for instanceB
	expected = 4
	if v, ok = result.GetValueUint8(instanceB); !ok {
		t.Error("metric [outage]: value for InstanceB not set")
	} else if v != expected {
		t.Errorf("metric [outage]: value for InstanceB is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [outage]: value for instanceB set to %d", v)
	}
}

func TestExcludeEqualIncludeEqualRuleOrder(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")

	instanceOne, _ := m.NewInstance("1")
	instanceOne.SetLabel("volstate", "offline")
	instanceOne.SetLabel("voltype", "rw")

	instanceTwo, _ := m.NewInstance("2")
	instanceTwo.SetLabel("volstate", "online")
	instanceTwo.SetLabel("voltype", "rw")

	instanceThree, _ := m.NewInstance("3")
	instanceThree.SetLabel("volstate", "online")
	instanceThree.SetLabel("voltype", "dp")

	// After the rules, only instanceTwo would be exportable
	p.excludeEquals(m)
	p.includeEquals(m)

	if instanceOne.IsExportable() {
		t.Error("InstanceOne should have been excluded")
	}

	if !instanceTwo.IsExportable() {
		t.Error("InstanceTwo should not have been excluded")
	}

	if instanceThree.IsExportable() {
		t.Error("instanceThree should have been excluded")
	}
}

func TestIncludeContainExcludeContainRuleOrder(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")

	instanceFour, _ := m.NewInstance("4")
	instanceFour.SetLabel("volstyle", "flexvol")
	instanceFour.SetLabel("volstatus", "running")

	instanceFive, _ := m.NewInstance("5")
	instanceFive.SetLabel("volstyle", "flexvol")
	instanceFive.SetLabel("volstatus", "stopped")

	instanceSix, _ := m.NewInstance("6")
	instanceSix.SetLabel("volstyle", "flexgroup")
	instanceSix.SetLabel("volstatus", "stopped")

	// After the rules, only instanceFour would be exportable
	p.includeContains(m)
	p.excludeContains(m)

	if !instanceFour.IsExportable() {
		t.Error("InstanceFour should not have been excluded")
	}

	if instanceFive.IsExportable() {
		t.Error("InstanceFive should have been excluded")
	}

	if instanceSix.IsExportable() {
		t.Error("instanceSix should have been excluded")
	}
}
