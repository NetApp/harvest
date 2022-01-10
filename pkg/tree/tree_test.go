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
			got += 1
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
			got += 1
		}
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		hiddenFields := name.GetChildS("hidden_fields")
		hiddenFieldsWant := 2
		hiddenFieldsGot := 0
		for range hiddenFields.GetChildren() {
			hiddenFieldsGot += 1
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
			got += 1
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
			got += 1
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		aggregator := plugins.GetChildS("Aggregator")
		aggregatorWant := 2
		aggregatorGot := 0
		for range aggregator.GetChildren() {
			aggregatorGot += 1
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
			got += 1
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		instanceKeys := exportOptions.GetChildS("instance_keys")
		instanceKeysWant := 2
		instanceKeysGot := 0
		for range instanceKeys.GetChildren() {
			instanceKeysGot += 1
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
			got += 1
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	want = 12
	got = 0
	if exporters := template.GetChildS("Exporters"); exporters != nil {
		for range exporters.GetChildren() {
			got += 1
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		prometheus := exporters.GetChildS("prometheusrange")
		prometheusWant := 2
		prometheusGot := 0
		for range prometheus.GetChildren() {
			prometheusGot += 1
		}
		if prometheusGot != prometheusWant {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	want = 5
	got = 0
	if defaults := template.GetChildS("Defaults"); defaults != nil {
		for range defaults.GetChildren() {
			got += 1
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		collectors := defaults.GetChildS("collectors")
		collectorsWant := 2
		collectorsGot := 0
		for range collectors.GetChildren() {
			collectorsGot += 1
		}
		if collectorsGot != collectorsWant {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	want = 12
	got = 0
	if pollers := template.GetChildS("Pollers"); pollers != nil {
		for range pollers.GetChildren() {
			got += 1
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		unix := pollers.GetChildS("unix")
		unixWant := 5
		unixGot := 0
		for range unix.GetChildren() {
			unixGot += 1
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
			got += 1
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		labelAgent := plugins.GetChildS("LabelAgent")
		labelAgentWant := 2
		labelAgentGot := 0
		for range labelAgent.GetChildren() {
			labelAgentGot += 1
		}
		if labelAgentGot != labelAgentWant {
			t.Errorf("got %v, want %v", labelAgentWant, labelAgentGot)
		}
	}
}
