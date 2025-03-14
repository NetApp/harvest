/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package labelagent

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
	// "github.com/netapp/harvest/v2/share/logger"
)

func newLabelAgent() *LabelAgent {
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
	// remove occurrences of "aaa_" from value of "A"
	params.NewChildS("replace", "").NewChildS("", "A A `aaa_` ``") //nolint:dupword
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
	// create metric "new_outage", if empty value is expected and non-empty means wrong, use default value "0"
	params.NewChildS("value_to_num", "").NewChildS("", "new_outage outage - - `0`")
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

	abc := plugin.New("Test", nil, params, nil, "", nil)
	p := &LabelAgent{AbstractPlugin: abc}

	if err := p.Init(conf.Remote{}); err != nil {
		panic(err)
	}
	return p
}

func TestSplitSimpleRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()
	instance, _ := m.NewInstance("0")
	instance.SetLabel("X", "a/b/c/d")

	_ = p.splitSimple(m)

	if instance.GetLabel("C") != "c" || instance.GetLabel("D") != "d" {
		t.Error("Labels C and D don't have expected values")
	}
}

func TestSplitRegexQtree(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	params := node.NewS("LabelAgent")
	params.NewChildS("split_regex", "").NewChildS("", "X `^/[^/]+/([^/]+)(?:/.*?|)/([^/]+)$` vol,lun")

	abc := plugin.New("Test", nil, params, nil, "", nil)
	p := &LabelAgent{AbstractPlugin: abc}
	err := p.Init(conf.Remote{})
	if err != nil {
		panic(err)
	}

	instance, _ := m.NewInstance("0")
	instance.SetLabel("X", "/vol/vol_georg_fcp401/lun401/lun-1")
	_ = p.splitRegex(m)

	if instance.GetLabel("lun") != "lun-1" {
		t.Errorf("got=%s want=lun-1", instance.GetLabel("lun"))
	}

	wantLun := "🦃\U0001FAF6🏾"
	instance.SetLabel("X", "/vol/vol_georg_fcp401/"+wantLun)
	_ = p.splitRegex(m)

	if instance.GetLabel("lun") != wantLun {
		t.Errorf("got=%s want=%s", instance.GetLabel("lun"), wantLun)
	}
}

func TestSplitRegexRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

	instance, _ := m.NewInstance("0")
	instance.SetLabel("X", "xxxA22_B333")

	_ = p.splitRegex(m)

	if instance.GetLabel("A") != "A22" || instance.GetLabel("B") != "B333" {
		t.Error("Labels A and B don't have expected values")
	}
}

func TestSplitPairsRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

	instance, _ := m.NewInstance("0")
	instance.SetLabel("X", "owner:jack contact:some@email")

	_ = p.splitPairs(m)

	if instance.GetLabel("owner") != "jack" || instance.GetLabel("contact") != "some@email" {
		t.Error("Labels owner and contact don't have expected values")
	}
}

func TestJoinSimpleRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

	instance, _ := m.NewInstance("0")
	instance.SetLabel("A", "aaa")
	instance.SetLabel("B", "bbb")

	_ = p.joinSimple(m)

	if instance.GetLabel("X") != "aaa_bbb" {
		t.Error("Label A does have expected value")
	}
}

func TestReplaceSimpleRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

	instance, _ := m.NewInstance("0")
	instance.SetLabel("A", "aaa_X")

	_ = p.replaceSimple(m)

	if instance.GetLabel("A") != "X" || instance.GetLabel("B") != "bbb_X" {
		t.Error("Labels A and B don't have expected values")
	}
}

func TestReplaceRegexRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

	instance, _ := m.NewInstance("0")
	instance.SetLabel("A", "aaa_12345_abcDEF")

	_ = p.replaceRegex(m)

	if instance.GetLabel("B") != "abcDEF-12345-bbb" {
		t.Error("Label B does not have expected value")
	}
}

func TestExcludeEqualsRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "aaa bbb ccc")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "aaa bbb")
	instanceNo.SetLabel("B", "aaa bbb ccc")

	_ = p.excludeEquals(m)

	if instanceYes.IsExportable() {
		t.Error("InstanceYes should have been excluded")
	}

	if !instanceNo.IsExportable() {
		t.Error("instanceNo should not have been excluded")
	}
}

func TestExcludeContainsRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "xxx_aaa_xxx")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "_aaa")

	_ = p.excludeContains(m)

	if instanceYes.IsExportable() {
		t.Error("InstanceYes should have been excluded")
	}

	if !instanceNo.IsExportable() {
		t.Error("instanceNo should not have been excluded")
	}
}

func TestExcludeRegexRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "aaa_123")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "aaa_123!")

	_ = p.excludeRegex(m)

	if instanceYes.IsExportable() {
		t.Error("InstanceYes should have been excluded")
	}

	if !instanceNo.IsExportable() {
		t.Error("instanceNo should not have been excluded")
	}
}

func TestIncludeEqualsRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "aaa bbb ccc")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "aaa bbb")
	instanceNo.SetLabel("B", "aaa bbb ccc")

	_ = p.includeEquals(m)

	if !instanceYes.IsExportable() {
		t.Error("InstanceYes should have been included")
	}

	if instanceNo.IsExportable() {
		t.Error("instanceNo should not have been included")
	}
}

func TestIncludeContainsRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "xxx_aaa_xxx")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "_aaa")

	_ = p.includeContains(m)

	if !instanceYes.IsExportable() {
		t.Error("InstanceYes should have been included")
	}

	if instanceNo.IsExportable() {
		t.Error("instanceNo should not have been included")
	}
}

func TestIncludeRegexRule(t *testing.T) {
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

	// should match
	instanceYes, _ := m.NewInstance("0")
	instanceYes.SetLabel("A", "aaa_123")

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "aaa_123!")

	_ = p.includeRegex(m)

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
		status, stage, outage *matrix.Metric
		v, expected           uint8
		ok                    bool
		err                   error
	)
	// should match
	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

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
		t.Fatal("metric [status] missing")
	}

	if stage = m.GetMetric("new_stage"); stage == nil {
		t.Fatal("metric [stage] missing")
	}

	if outage = m.GetMetric("new_outage"); outage == nil {
		t.Fatal("metric [outage] missing")
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

func TestValueToNumRegexRule(t *testing.T) {

	var (
		instanceA, instanceB   *matrix.Instance
		status, output, result *matrix.Metric
		v, expected            uint8
		ok                     bool
		err                    error
	)

	m := matrix.New("TestLabelAgent", "test", "test")
	p := newLabelAgent()

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
		t.Fatal("metric [status] missing")
	}

	if output = m.GetMetric("output"); output == nil {
		t.Fatal("metric [output] missing")
	}

	if result = m.GetMetric("result"); result == nil {
		t.Fatal("metric [result] missing")
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
	p := newLabelAgent()

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
	_ = p.excludeEquals(m)
	_ = p.includeEquals(m)

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
	p := newLabelAgent()

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
	_ = p.includeContains(m)
	_ = p.excludeContains(m)

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
