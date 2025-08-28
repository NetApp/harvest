package test

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/yaml"
	"os"
	"strings"
	"testing"
)

// Merge default.yaml and custom.yaml
func TestNode_Merge(t *testing.T) {
	defaultTemplate, _ := tree.ImportYaml("testdata/default_collector.yaml")
	customTemplate, _ := tree.ImportYaml("testdata/extend_collector.yaml")
	defaultTemplate.Merge(customTemplate, []string{"objects"})

	// count number of objects post merge
	want := 10
	got := 0
	if objects := defaultTemplate.GetChildS("objects"); objects != nil {
		for range objects.GetChildren() {
			got++
		}
		assert.Equal(t, got, want)
	}

	// Compare overwritten values for object
	want1 := "node.yaml,node2.yaml"
	got1 := defaultTemplate.GetChildS("objects").GetChildS("Node").GetContentS()

	if got1 != want1 {
		t.Errorf("got %v, want %v", got1, want1)
	}

	// Check if additional object is added
	checkObject := "Sensor"
	assert.NotNil(t, defaultTemplate.GetChildS("objects").GetChildS(checkObject))

	// Compare overwritten values for schedule
	want2 := "650s"
	got2 := defaultTemplate.GetChildS("schedule").GetChildS("instance").GetContentS()

	assert.Equal(t, got2, want2)
}

// merge collector templates for 21.08.6+ versions
// change is LabelAgent child will have list of rules instead of key-value pair
func TestNode_MergeCollector(t *testing.T) {
	defaultTemplate, _ := tree.ImportYaml("testdata/lun.yaml")
	customTemplate, _ := tree.ImportYaml("testdata/extend_lun.yaml")
	defaultTemplate.PreprocessTemplate()
	customTemplate.PreprocessTemplate()
	defaultTemplate.Merge(customTemplate, nil)

	gotString1, _ := yaml.Dump(defaultTemplate)
	gotString := strings.TrimSpace(string(gotString1))
	expected, _ := os.ReadFile("mergeTemplates/lun_merge.yaml")
	expectedString := strings.TrimSpace(string(expected))

	assert.Equal(t, gotString, expectedString)

	// object name overwrite
	want := "customLun"
	var got string
	if name := defaultTemplate.GetChildS("name"); name != nil {
		got = name.GetContentS()
		assert.Equal(t, got, want)
	}

	// Add new counter
	want1 := 9
	got1 := 0
	counters := defaultTemplate.GetChildS("counters").GetChildS("lun-info")

	if counters != nil {
		for range counters.GetChildren() {
			got1++
		}
	}

	assert.Equal(t, got1, want1)

	// plugins labelagent add same child to existing plugin
	want3 := 1
	got3 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("value_to_num")
	if counters != nil {
		for range counters.GetChildren() {
			got3++
		}
	}

	assert.Equal(t, got3, want3)

	// plugins labelagent add same child to existing plugin
	want4 := 1
	got4 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("new_mapping")
	if counters != nil {
		for range counters.GetChildren() {
			got4++
		}
	}

	assert.Equal(t, got4, want4)

	// plugins labelagent existing child no change
	want5 := 1
	got5 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("split")
	if counters != nil {
		for range counters.GetChildren() {
			got5++
		}
	}

	assert.Equal(t, got5, want5)

	// plugins aggregator add new child
	want8 := 2
	got8 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("Aggregator")
	if counters != nil {
		for range counters.GetChildren() {
			got8++
		}
	}

	assert.Equal(t, got8, want8)

	// export_options add new instance_key
	want6 := 6
	got6 := 0
	counters = defaultTemplate.GetChildS("export_options").GetChildS("instance_keys")
	if counters != nil {
		for range counters.GetChildren() {
			got6++
		}
	}

	assert.Equal(t, got6, want6)

	// export_options add same instance_labels
	want7 := 1
	got7 := 0
	counters = defaultTemplate.GetChildS("export_options").GetChildS("instance_labels")
	if counters != nil {
		for range counters.GetChildren() {
			got7++
		}
	}

	assert.Equal(t, got7, want7)

	// override block
	want9 := 2
	got9 := 0
	counters = defaultTemplate.GetChildS("override")
	if counters != nil {
		for range counters.GetChildren() {
			got9++
		}
	}

	assert.Equal(t, got9, want9)

	//export block

	export := defaultTemplate.GetChildS("export")
	assert.Nil(t, export)
	assert.Equal(t, got9, want9)
}

// Merge collector templates where custom templates are from 21.08.6 and before
// LabelAgent child did have key-value pair of rules instead of a list
func TestNode_MergeCollectorOld(t *testing.T) {
	defaultTemplate, _ := tree.ImportYaml("testdata/lun.yaml")
	customTemplate, _ := tree.ImportYaml("testdata/21.08.0_extend_lun.yaml")
	defaultTemplate.PreprocessTemplate()
	customTemplate.PreprocessTemplate()
	defaultTemplate.Merge(customTemplate, nil)

	// plugins labelagent add same child to existing plugin
	want3 := 1
	got3 := 0
	counters := defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("value_to_num")
	if counters != nil {
		for range counters.GetChildren() {
			got3++
		}
	}

	assert.Equal(t, got3, want3)

	// plugins labelagent add same child to existing plugin
	want4 := 1
	got4 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("LabelAgent").GetChildS("new_mapping")
	if counters != nil {
		for range counters.GetChildren() {
			got4++
		}
	}

	assert.Equal(t, got4, want4)

	// plugins aggregator add new child
	want5 := 3
	got5 := 0
	counters = defaultTemplate.GetChildS("plugins").GetChildS("Aggregator")
	if counters != nil {
		for range counters.GetChildren() {
			got5++
		}
	}

	assert.Equal(t, got5, want5)
}

func TestNode_PreProcessCollector(t *testing.T) {

	tests := []struct {
		name        string
		sourceFile  string
		compareFile string
	}{
		{
			name:        "preprocess template from 21.08.0",
			sourceFile:  "testdata/21.08.0_extend_lun.yaml",
			compareFile: "preProcessResultData/p_21.08.0_extend_lun.yaml",
		},
		{
			name:        "preprocess template after 21.08.0",
			sourceFile:  "testdata/21.08.0_lun.yaml",
			compareFile: "preProcessResultData/p_21.08.0_lun.yaml",
		},
		{
			name:        "process collector template",
			sourceFile:  "testdata/default_collector.yaml",
			compareFile: "preProcessResultData/p_default_collector.yaml",
		},
		{
			name:        "process extended collector template",
			sourceFile:  "testdata/extend_collector.yaml",
			compareFile: "preProcessResultData/p_extend_collector.yaml",
		},
		{
			name:        "process extended object template",
			sourceFile:  "testdata/extend_lun.yaml",
			compareFile: "preProcessResultData/p_extend_lun.yaml",
		},
		{
			name:        "process object template",
			sourceFile:  "testdata/lun.yaml",
			compareFile: "preProcessResultData/p_lun.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, _ := tree.ImportYaml(tt.sourceFile)
			template.PreprocessTemplate()
			got, _ := yaml.Dump(template)
			expected, _ := os.ReadFile(tt.compareFile)
			gotString := strings.TrimSpace(string(got))
			expectedString := strings.TrimSpace(string(expected))

			assert.Equal(t, gotString, expectedString)
		})
	}
}

func TestNode_PreProcessMergeCollector(t *testing.T) {

	tests := []struct {
		name           string
		baseTemplate   string
		extendTemplate string
		mergeTemplate  string
	}{
		{
			name:           "Case1: Both base and extended template follow new convention for labelagent which is list",
			baseTemplate:   "testdata/lun.yaml",
			extendTemplate: "testdata/extend_lun.yaml",
			mergeTemplate:  "mergeTemplates/lun_merge.yaml",
		},
		{
			name:           "Case2: base template follow new convention for labelagent and extended template follow 21.08.0",
			baseTemplate:   "testdata/lun.yaml",
			extendTemplate: "testdata/21.08.0_extend_lun.yaml",
			mergeTemplate:  "mergeTemplates/lun_merge_21.08.0_extended.yaml",
		},
		{
			name:           "Case3: base template follow old convention for labelagent and extended template follow 21.08.0",
			baseTemplate:   "testdata/21.08.0_lun.yaml",
			extendTemplate: "testdata/21.08.0_extend_lun.yaml",
			mergeTemplate:  "mergeTemplates/21.08.0_lun_merge_21.08.0_extended.yaml",
		},
		{
			name:           "Case4: base template follow old convention for labelagent and extended template follow new",
			baseTemplate:   "testdata/21.08.0_lun.yaml",
			extendTemplate: "testdata/extend_lun.yaml",
			mergeTemplate:  "mergeTemplates/21.08.0_lun_merge_extended.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseTemplate, _ := tree.ImportYaml(tt.baseTemplate)
			extendTemplate, _ := tree.ImportYaml(tt.extendTemplate)
			baseTemplate.PreprocessTemplate()
			extendTemplate.PreprocessTemplate()
			baseTemplate.Merge(extendTemplate, nil)
			gotString1, _ := yaml.Dump(baseTemplate)
			gotString := strings.TrimSpace(string(gotString1))
			expected, _ := os.ReadFile(tt.mergeTemplate)
			expectedString := strings.TrimSpace(string(expected))

			assert.Equal(t, gotString, expectedString)
		})
	}
}
