package tree

import (
	"testing"
)

func TestImportYaml(t *testing.T) {
	template, _ := ImportYaml("testdata/testTemplate.yaml")

	// check key value pairs like
	// name: Volume
	want := 0
	got := 0
	if name := template.GetChildS("name"); name != nil {
		for range name.GetChildren() {
			got++
		}
		if name.GetContentS() == "" {
			t.Errorf("empty content")
		}
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	// check counters
	want = 6
	got = 0
	if name := template.GetChildS("counters"); name != nil {
		for range name.GetChildren() {
			got++
		}
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		hiddenFields := name.GetChildS("hidden_fields")
		hiddenFieldsWant := 2
		hiddenFieldsGot := 0
		for range hiddenFields.GetChildren() {
			hiddenFieldsGot++
		}
		if hiddenFieldsGot != hiddenFieldsWant {
			t.Errorf("got %v, want %v", hiddenFieldsGot, hiddenFieldsWant)
		}
	}

	// check endpoints
	want = 2
	got = 0
	if endpoints := template.GetChildS("endpoints"); endpoints != nil {
		for range endpoints.GetChildren() {
			got++
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	// check plugins
	want = 2
	got = 0
	if plugins := template.GetChildS("plugins"); plugins != nil {
		for range plugins.GetChildren() {
			got++
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		aggregator := plugins.GetChildS("Aggregator")
		aggregatorWant := 2
		aggregatorGot := 0
		for range aggregator.GetChildren() {
			aggregatorGot++
		}
		if aggregatorGot != aggregatorWant {
			t.Errorf("got %v, want %v", aggregatorWant, aggregatorGot)
		}
	}

	// export_options
	want = 2
	got = 0
	if exportOptions := template.GetChildS("export_options"); exportOptions != nil {
		for range exportOptions.GetChildren() {
			got++
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		instanceKeys := exportOptions.GetChildS("instance_keys")
		instanceKeysWant := 2
		instanceKeysGot := 0
		for range instanceKeys.GetChildren() {
			instanceKeysGot++
		}
		if instanceKeysGot != instanceKeysWant {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestHarvestConfigImportYaml(t *testing.T) {
	template, _ := ImportYaml("../../cmd/tools/doctor/testdata/testConfig.yml")

	want := 2
	got := 0
	if name := template.GetChildS("Tools"); name != nil {
		for range name.GetChildren() {
			got++
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	want = 12
	got = 0
	if exporters := template.GetChildS("Exporters"); exporters != nil {
		for range exporters.GetChildren() {
			got++
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		prometheus := exporters.GetChildS("prometheusrange")
		prometheusWant := 2
		prometheusGot := 0
		for range prometheus.GetChildren() {
			prometheusGot++
		}
		if prometheusGot != prometheusWant {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	want = 5
	got = 0
	if defaults := template.GetChildS("Defaults"); defaults != nil {
		for range defaults.GetChildren() {
			got++
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		collectors := defaults.GetChildS("collectors")
		collectorsWant := 2
		collectorsGot := 0
		for range collectors.GetChildren() {
			collectorsGot++
		}
		if collectorsGot != collectorsWant {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	want = 12
	got = 0
	if pollers := template.GetChildS("Pollers"); pollers != nil {
		for range pollers.GetChildren() {
			got++
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		unix := pollers.GetChildS("unix")
		unixWant := 5
		unixGot := 0
		for range unix.GetChildren() {
			unixGot++
		}
		if unixGot != unixWant {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestImport2108Yaml(t *testing.T) {
	template, _ := ImportYaml("testdata/testTemplate21.08.yaml")
	// check plugins 21.08 (old backward compatibility)
	want := 1
	got := 0
	if plugins := template.GetChildS("plugins"); plugins != nil {
		for range plugins.GetChildren() {
			got++
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		labelAgent := plugins.GetChildS("LabelAgent")
		labelAgentWant := 2
		labelAgentGot := 0
		for range labelAgent.GetChildren() {
			labelAgentGot++
		}
		if labelAgentGot != labelAgentWant {
			t.Errorf("got %v, want %v", labelAgentWant, labelAgentGot)
		}
	}

	// counters
	want = 1
	got = 0
	if counters := template.GetChildS("counters"); counters != nil {
		for range counters.GetChildren() {
			got++
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		volumeAttributes := counters.GetChildS("volume-attributes")
		volumeAttributesWant := 1
		volumeAttributesGot := 0
		for range volumeAttributes.GetChildren() {
			volumeAttributesGot++
		}
		if volumeAttributesGot != volumeAttributesWant {
			t.Errorf("got %v, want %v", volumeAttributesWant, volumeAttributesGot)
		}

		volumeAutoSizeAttributes := volumeAttributes.GetChildS("volume-autosize-attributes")
		volumeAutoSizeAttributesWant := 2
		volumeAutoSizeAttributesGot := 0
		for range volumeAutoSizeAttributes.GetChildren() {
			volumeAutoSizeAttributesGot++
		}
		if volumeAutoSizeAttributesGot != volumeAutoSizeAttributesWant {
			t.Errorf("got %v, want %v", volumeAutoSizeAttributesGot, volumeAttributesGot)
		}
	}
}
