/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package label_agent

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"testing"
	//"goharvest2/share/logger"
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
	params.NewChildS("process_field_value", "").NewChildS("", "space_total add space_available space_used")
	// create metric "disk_count", which is addition of the metric value of primary.disk_count, secondary.disk_count and hybrid.disk_count
	params.NewChildS("process_field_value", "").NewChildS("", "disk_count add primary.disk_count secondary.disk_count hybrid.disk_count")

	abc := plugin.New("Test", nil, params, nil)
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

	if err = p.mapValues(m); err != nil {
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
		t.Errorf("metric [status]: value for InstanceA is %d, should not be set", v)
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

func TestProcessFieldValueRule(t *testing.T) {

	var (
		instanceA, instanceB                                   *matrix.Instance
		metricAvail, metricUsed, metricTotal                   matrix.Metric
		metricDiskP, metricDiskS, metricDiskH, metricDiskTotal matrix.Metric
		expected                                               float64
		err                                                    error
	)
	// should match
	m := matrix.New("TestLabelAgent", "test", "test")

	if instanceA, err = m.NewInstance("A"); err != nil {
		t.Fatal(err)
	}

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

	if instanceB, err = m.NewInstance("B"); err != nil {
		t.Fatal(err)
	}

	if metricDiskP, err = m.NewMetricFloat64("primary.disk_count"); err == nil {
		metricDiskP.SetValueFloat64(instanceB, 8)
	} else {
		t.Error("metric [primary.disk_coun]  not created for InstanceB")
	}

	if metricDiskS, err = m.NewMetricFloat64("secondary.disk_count"); err == nil {
		metricDiskS.SetValueFloat64(instanceB, 10)
	} else {
		t.Error("metric [secondary.disk_coun]  not created for InstanceB")
	}

	if metricDiskH, err = m.NewMetricFloat64("hybrid.disk_count"); err == nil {
		metricDiskH.SetValueFloat64(instanceB, 4)
	} else {
		t.Error("metric [hybrid.disk_count]  not created for InstanceB")
	}

	if err = p.processFields(m); err != nil {
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

}
