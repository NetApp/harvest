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
	params.NewChildS("split", "X `/` ,,C,D")
	// split value of "X" and take first 2 as new labels
	params.NewChildS("split_regex", "X `.*(A\\d+)_(B\\d+)` A,B")
	// split value of "X" into key-value pairs
	params.NewChildS("split_pairs", "X ` ` `:`")
	// join values of "A" and "B" and set as label "X"
	params.NewChildS("join", "X `_` A,B")
	// replace "aaa_" with "bbb_" and set as label "B"
	params.NewChildS("replace", "A B `aaa_` `bbb_`")
	// remove occurences of "aaa_" from value of "A"
	params.NewChildS("replace", "A A `aaa_` ``")
	// reverse the order of matching elements, replace underscore with dash and "aaa" with "bbb"
	params.NewChildS("replace_regex", "A B `^(aaa)_(\\d+)_(\\w+)$` `$3-$2-bbb`")
	// exclude instance if label "A" has value "aaa bbb ccc"
	params.NewChildS("exclude_equals", "A `aaa bbb ccc`")
	// exclude instance if label "A" contains "_aaa_"
	params.NewChildS("exclude_contains", "A `_aaa_`")
	// exclude instance of label value has prefix "aaa_" followed by at least one digit
	params.NewChildS("exclude_regex", "A `^aaa_\\d+$`")
	// create metric "status", if label "state" is one of the 3, map metric value to respective index
	params.NewChildS("value_mapping", "status state up,sleeping,down")
	// similar to above, but if none of the values is matching, use default value "4"
	params.NewChildS("value_mapping", "stage stage init `1`")

	abc := plugin.New("Test", nil, params, nil)
	p = &LabelAgent{}

	if err := p.Init(abc); err != nil {
		t.Fatal(err)
	}
}

func TestSplitSimpleRule(t *testing.T) {
	instance := matrix.NewInstance(0)
	instance.SetLabel("X", "a/b/c/d")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.splitSimple(instance)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("C") == "c" && instance.GetLabel("D") == "d" {
		// OK
	} else {
		t.Error("Labels C and D don't have expected values")
	}
}

func TestSplitRegexRule(t *testing.T) {
	instance := matrix.NewInstance(0)
	instance.SetLabel("X", "xxxA22_B333")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.splitRegex(instance)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("A") == "A22" && instance.GetLabel("B") == "B333" {
		// OK
	} else {
		t.Error("Labels A and B don't have expected values")
	}
}

func TestSplitPairsRule(t *testing.T) {
	instance := matrix.NewInstance(0)
	instance.SetLabel("X", "owner:jack contact:some@email")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.splitPairs(instance)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("owner") == "jack" && instance.GetLabel("contact") == "some@email" {
		// OK
	} else {
		t.Error("Labels owner and contact don't have expected values")
	}
}

func TestJoinSimpleRule(t *testing.T) {
	instance := matrix.NewInstance(0)
	instance.SetLabel("A", "aaa")
	instance.SetLabel("B", "bbb")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.joinSimple(instance)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("X") == "aaa_bbb" {
		// OK
	} else {
		t.Error("Label A does have expected value")
	}
}

func TestReplaceSimpleRule(t *testing.T) {
	instance := matrix.NewInstance(0)
	instance.SetLabel("A", "aaa_X")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.replaceSimple(instance)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("A") == "X" && instance.GetLabel("B") == "bbb_X" {
		// OK
	} else {
		t.Error("Labels A and B don't have expected values")
	}
}

func TestReplaceRegexRule(t *testing.T) {
	instance := matrix.NewInstance(0)
	instance.SetLabel("A", "aaa_12345_abcDEF")

	t.Logf("before = [%s]\n", instance.GetLabels().String())
	p.replaceRegex(instance)
	t.Logf("after  = [%s]\n", instance.GetLabels().String())

	if instance.GetLabel("B") == "abcDEF-12345-bbb" {
		// OK
	} else {
		t.Error("Label B does not have expected value")
	}
}

func TestExcludeEqualsRule(t *testing.T) {
	// should match
	instanceYes := matrix.NewInstance(0)
	instanceYes.SetLabel("A", "aaa bbb ccc")

	// should not match
	instanceNo := matrix.NewInstance(1)
	instanceNo.SetLabel("A", "aaa bbb")
	instanceNo.SetLabel("B", "aaa bbb ccc")

	p.excludeEquals(instanceYes)
	p.excludeEquals(instanceNo)

	if instanceYes.IsExportable() {
		t.Error("InstanceYes should have been excluded")
	}

	if !instanceNo.IsExportable() {
		t.Error("instanceNo should not have been excluded")
	}
}

func TestExcludeContainsRule(t *testing.T) {
	// should match
	instanceYes := matrix.NewInstance(0)
	instanceYes.SetLabel("A", "xxx_aaa_xxx")

	// should not match
	instanceNo := matrix.NewInstance(1)
	instanceNo.SetLabel("A", "_aaa")

	p.excludeContains(instanceYes)
	p.excludeContains(instanceNo)

	if instanceYes.IsExportable() {
		t.Error("InstanceYes should have been excluded")
	}

	if !instanceNo.IsExportable() {
		t.Error("instanceNo should not have been excluded")
	}
}

func TestExcludeRegexRule(t *testing.T) {
	// should match
	instanceYes := matrix.NewInstance(0)
	instanceYes.SetLabel("A", "aaa_123")

	// should not match
	instanceNo := matrix.NewInstance(1)
	instanceNo.SetLabel("A", "aaa_123!")

	p.excludeRegex(instanceYes)
	p.excludeRegex(instanceNo)

	if instanceYes.IsExportable() {
		t.Error("InstanceYes should have been excluded")
	}

	if !instanceNo.IsExportable() {
		t.Error("instanceNo should not have been excluded")
	}
}

func TestValueMappingRule(t *testing.T) {

	var (
		instanceA, instanceB *matrix.Instance
		status, stage        matrix.Metric
		v, expected          uint8
		ok                   bool
		err                  error
	)
	// should match
	m := matrix.New("TestLabelAgent", "test")

	if instanceA, err = m.NewInstance("A"); err != nil {
		t.Fatal(err)
	}
	instanceA.SetLabel("state", "down") // "status" should be 1
	instanceA.SetLabel("stage", "init") // "stage" should be 0

	if instanceB, err = m.NewInstance("B"); err != nil {
		t.Fatal(err)
	}
	instanceB.SetLabel("state", "unknown") // "status" should not be set
	instanceB.SetLabel("stage", "unknown") // "stage" should be 1 (default)

	if err = p.mapValues(m); err != nil {
		t.Fatal(err)
	}

	if status = m.GetMetric("status"); status == nil {
		t.Error("metric [status] missing")
	}

	if stage = m.GetMetric("stage"); stage == nil {
		t.Error("metric [stage] missing")
	}

	// check "status" for instanceA
	expected = 2
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
	expected = 0
	if v, ok = stage.GetValueUint8(instanceA); !ok {
		t.Error("metric [stage]: value for InstanceA not set")
	} else if v != expected {
		t.Errorf("metric [stage]: value for InstanceA is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [stage]: value for instanceA set to %d", v)
	}

	// check "stage" for instanceB
	expected = 1
	if v, ok = stage.GetValueUint8(instanceB); !ok {
		t.Error("metric [stage]: value for InstanceB not set")
	} else if v != expected {
		t.Errorf("metric [stage]: value for InstanceB is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [stage]: value for instanceB set to %d", v)
	}
}
