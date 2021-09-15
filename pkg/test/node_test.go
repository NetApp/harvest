package test

import (
	"goharvest2/pkg/tree"
	"testing"
)

// Merge default.yaml and custom.yaml
func TestNode_Merge(t *testing.T) {
	defaultTemplate, _ := tree.Import("yaml", "testdata/default.yaml")
	customTemplate, _ := tree.Import("yaml", "testdata/custom.yaml")
	defaultTemplate.Merge(customTemplate, []string{"objects"})

	// count number of objects post merge
	want := 10
	got := 0
	if objects := defaultTemplate.GetChildS("objects"); objects != nil {
		for range objects.GetChildren() {
			got += 1
		}
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	// Compare overwritten values for object
	want1 := "node.yaml,node2.yaml"
	got1 := defaultTemplate.GetChildS("objects").GetChildS("Node").GetContentS()

	if got1 != want1 {
		t.Errorf("got %v, want %v", got1, want1)
	}

	// Check if additional object is added
	checkObject := "Sensor"
	if defaultTemplate.GetChildS("objects").GetChildS(checkObject) == nil {
		t.Errorf("missing object %v", checkObject)
	}

	// Compare overwritten values for schedule
	want2 := "650s"
	got2 := defaultTemplate.GetChildS("schedule").GetChildS("instance").GetContentS()

	if got2 != want2 {
		t.Errorf("got %v, want %v", got2, want2)
	}

}

// merge collector templates for 21.08.6+ versions
// change is LabelAgent child will have list of rules instead of key-value pair
func TestNode_MergeCollector(t *testing.T) {
	defaultTemplate, _ := tree.Import("yaml", "testdata/lun.yaml")
	customTemplate, _ := tree.Import("yaml", "testdata/custom_lun.yaml")
	defaultTemplate.Merge(customTemplate, nil)

	// object name overwrite
	want := "customLun"
	got := ""
	if name := defaultTemplate.GetChildS("name"); name != nil {
		got = name.GetContentS()
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	// Add new counter
	want1 := 9
	got1 := 0
	counters := defaultTemplate.GetChildS("counters").GetChildS("lun-info")

	if counters != nil {
		for range counters.GetChildren() {
			got1 += 1
		}
	}

	if got1 != want1 {
		t.Errorf("got %v, want %v", got1, want1)
	}

	// plugins labelagent add new child to existing plugin
	want2 := 2
	got2 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("value_mapping")
	if counters != nil {
		for range counters.GetChildren() {
			got2 += 1
		}
	}

	if got2 != want2 {
		t.Errorf("got %v, want %v", got2, want2)
	}

	// plugins labelagent add same child to existing plugin
	want3 := 1
	got3 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("value_to_num")
	if counters != nil {
		for range counters.GetChildren() {
			got3 += 1
		}
	}

	if got3 != want3 {
		t.Errorf("got %v, want %v", got3, want3)
	}

	// plugins labelagent add same child to existing plugin
	want4 := 1
	got4 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("new_mapping")
	if counters != nil {
		for range counters.GetChildren() {
			got4 += 1
		}
	}

	if got4 != want4 {
		t.Errorf("got %v, want %v", got4, want4)
	}

	// plugins labelagent existing child no change
	want5 := 1
	got5 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("split")
	if counters != nil {
		for range counters.GetChildren() {
			got5 += 1
		}
	}

	if got5 != want5 {
		t.Errorf("got %v, want %v", got5, want5)
	}

	// plugins aggregator add new child
	want8 := 2
	got8 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("Aggregator")
	if counters != nil {
		for range counters.GetChildren() {
			got8 += 1
		}
	}

	if got8 != want8 {
		t.Errorf("got %v, want %v", got8, want8)
	}

	//export_options add new instance_key
	want6 := 6
	got6 := 0
	counters = defaultTemplate.GetChildS("export_options").GetChildS("instance_keys")
	if counters != nil {
		for range counters.GetChildren() {
			got6 += 1
		}
	}

	if got6 != want6 {
		t.Errorf("got %v, want %v", want6, got6)
	}

	//export_options add same instance_labels
	want7 := 1
	got7 := 0
	counters = defaultTemplate.GetChildS("export_options").GetChildS("instance_labels")
	if counters != nil {
		for range counters.GetChildren() {
			got7 += 1
		}
	}

	if got7 != want7 {
		t.Errorf("got %v, want %v", want7, got7)
	}

	//override block
	want9 := 2
	got9 := 0
	counters = defaultTemplate.GetChildS("override")
	if counters != nil {
		for range counters.GetChildren() {
			got9 += 1
		}
	}

	if want9 != got9 {
		t.Errorf("got %v, want %v", want9, got9)
	}

	//export block

	export := defaultTemplate.GetChildS("export")
	if export != nil {
		t.Errorf("missing export block")
	}

	if want9 != got9 {
		t.Errorf("got %v, want %v", want9, got9)
	}
}

// Merge collector templates where custom templates are from 21.08.6 and before
// LabelAgent child did have key-value pair of rules instead of a list
func TestNode_MergeCollectorOld(t *testing.T) {
	defaultTemplate, _ := tree.Import("yaml", "testdata/lun.yaml")
	customTemplate, _ := tree.Import("yaml", "testdata/custom_lun_old.yaml")
	defaultTemplate.Merge(customTemplate, nil)

	// plugins labelagent add new child to existing plugin
	want2 := 2
	got2 := 0
	counters := defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("value_mapping")
	if counters != nil {
		for range counters.GetChildren() {
			got2 += 1
		}
	}

	if got2 != want2 {
		t.Errorf("got %v, want %v", got2, want2)
	}

	// plugins labelagent add same child to existing plugin
	want3 := 1
	got3 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("value_to_num")
	if counters != nil {
		for range counters.GetChildren() {
			got3 += 1
		}
	}

	if got3 != want3 {
		t.Errorf("got %v, want %v", got3, want3)
	}

	// plugins labelagent add same child to existing plugin
	want4 := 1
	got4 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("new_mapping")
	if counters != nil {
		for range counters.GetChildren() {
			got4 += 1
		}
	}

	if got4 != want4 {
		t.Errorf("got %v, want %v", got4, want4)
	}

	// plugins aggregator add new child
	want5 := 3
	got5 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("Aggregator")
	if counters != nil {
		for range counters.GetChildren() {
			got5 += 1
		}
	}

	if got5 != want5 {
		t.Errorf("got %v, want %v", got5, want5)
	}
}
