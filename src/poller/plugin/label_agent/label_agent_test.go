package label_agent

import (
	"goharvest2/poller/plugin"
	"goharvest2/share/matrix"
	"goharvest2/share/tree/node"
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

	abc := plugin.New("Test", nil, params, nil)
	p = &LabelAgent{AbstractPlugin: abc}

	if err := p.Init(); err != nil {
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
