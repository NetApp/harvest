package grafana

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Metric struct {
	Metric         string `yaml:"metric"`
	OntapUnit      string `yaml:"ontapUnit"`
	GrafanaJSON    string `yaml:"grafanaJson"`
	GrafanaDisplay string `yaml:"grafanaDisplay"`
	Comment        string `yaml:"comment"`
	skipValidate   bool
}

func parseUnits() map[string]Metric {

	excludeValidationMap := map[string]struct{}{
		"svm_nfs_read_throughput":  {},
		"svm_nfs_throughput":       {},
		"svm_nfs_write_throughput": {},
	}
	filePath := "units.yaml"

	// Read the YAML file
	yamlData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	var metrics []Metric

	// Unmarshal the YAML data into the metrics slice
	err = yaml.Unmarshal(yamlData, &metrics)
	if err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Create a map to store the metrics
	metricsMap := make(map[string]Metric)

	// Populate the map using the metric name as the key
	for _, metric := range metrics {
		_, metric.skipValidate = excludeValidationMap[metric.Metric]
		metricsMap[metric.Metric] = metric
	}
	return metricsMap
}
