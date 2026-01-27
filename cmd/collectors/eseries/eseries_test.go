package eseries

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

const (
	pollerName = "test"
)

// JSONToGson reads JSON test data and converts to gjson.Result array
func JSONToGson(path string, flatten bool) []gjson.Result {
	var (
		result []gjson.Result
		err    error
	)

	var reader io.Reader
	if filepath.Ext(path) == ".gz" {
		open, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer open.Close()
		reader, err = gzip.NewReader(open)
		if err != nil {
			panic(err)
		}
		defer reader.(*gzip.Reader).Close()
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		reader = bytes.NewReader(data)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	if flatten {
		result = gjson.GetBytes(data, "..").Array()
	} else {
		result = gjson.ParseBytes(data).Array()
	}

	return result
}

// Params creates a minimal params tree for testing
func Params(object string, path string) *node.Node {
	yml := `
schedule:
  - data: 9999h
type: %s
objects:
  %s: %s
`
	yml = fmt.Sprintf(yml, object, object, path)
	root, err := tree.LoadYaml([]byte(yml))
	if err != nil {
		panic(err)
	}
	return root
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestMain(m *testing.M) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	os.Exit(m.Run())
}

func TestESeries_PollData(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	tests := []struct {
		name           string
		object         string
		template       string
		pollDataPath   string
		numInstances   int
		numMetrics     uint64
		instanceKey    string
		expectedLabels []string
	}{
		{
			name:         "volume",
			object:       "Volume",
			template:     "volume.yaml",
			pollDataPath: "testdata/volume.json",
			numInstances: 22,
			numMetrics:   44,
			instanceKey:  "vmware-hdd1",
			expectedLabels: []string{
				"block_size",
				"offline",
				"raid_level",
				"status",
				"volume",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newESeries(tt.object, tt.template)

			e.arrayID = "600a098000f63714000000005e5cf5d2"
			e.arrayName = "eseries-test-system"
			mat := e.Matrix[e.Object]
			mat.SetGlobalLabel("array", e.arrayName)

			pollData := JSONToGson(tt.pollDataPath, false)

			count := e.pollData(mat, pollData)

			instances := mat.GetInstances()
			assert.Equal(t, len(instances), tt.numInstances)

			assert.Equal(t, count, tt.numMetrics)

			instanceKeys := set.New()
			for key := range instances {
				if instanceKeys.Has(key) {
					t.Errorf("duplicate instance key found: %s", key)
				}
				instanceKeys.Add(key)
			}

			instance := mat.GetInstance(tt.instanceKey)
			assert.NotNil(t, instance)

			if instance != nil {
				labels := instance.GetLabels()
				for _, expectedLabel := range tt.expectedLabels {
					value, ok := labels[expectedLabel]
					if !ok || value == "" {
						t.Errorf("label %s should not be empty", expectedLabel)
					}
				}
			}

			globalLabels := mat.GetGlobalLabels()
			assert.Equal(t, globalLabels["array"], e.arrayName)

			metrics := mat.GetMetrics()

			expectedMetricNames := []string{"capacity", "totalSizeInBytes"}
			for _, metricName := range expectedMetricNames {
				metric := metrics[metricName]
				assert.NotNil(t, metric)
				if metric != nil {
					assert.True(t, metric.IsExportable())
				}
			}
		})
	}
}

func TestESeries_Init(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	tests := []struct {
		name         string
		object       string
		template     string
		wantErr      bool
		numMetrics   int
		numLabels    int
		expectedKeys []string
	}{
		{
			name:         "volume_init",
			object:       "Volume",
			template:     "volume.yaml",
			wantErr:      false,
			numMetrics:   2,
			numLabels:    9,
			expectedKeys: []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newESeries(tt.object, tt.template)

			assert.NotNil(t, e.Prop)
			assert.NotNil(t, e.Client)
			assert.NotNil(t, e.Matrix)

			assert.Equal(t, len(e.Prop.Metrics), tt.numMetrics)

			assert.Equal(t, len(e.Prop.InstanceLabels), tt.numLabels)

			assert.Equal(t, len(e.Prop.InstanceKeys), len(tt.expectedKeys))
			for i, expectedKey := range tt.expectedKeys {
				assert.Equal(t, e.Prop.InstanceKeys[i], expectedKey)
			}

			if e.Prop.Query == "" {
				t.Error("query should not be empty")
			}
			if !contains(e.Prop.Query, "{array_id}") {
				t.Error("query should contain array_id placeholder")
			}
		})
	}
}

func TestESeries_PollCounter(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	e := newESeries("Volume", "volume.yaml")

	arrayJSON := JSONToGson("testdata/storage-systems.json", false)

	if len(arrayJSON) > 0 {
		array := arrayJSON[0]
		e.arrayID = array.Get("id").ClonedString()
		e.arrayName = array.Get("name").ClonedString()

		mat := e.Matrix[e.Object]
		mat.SetGlobalLabel("array", e.arrayName)
	}

	if e.arrayName == "" {
		t.Error("array should be set")
	}
	assert.Equal(t, e.arrayID, "600a098000f63714000000005e5cf5d2")
	assert.Equal(t, e.arrayName, "eseries-test-system")

	mat := e.Matrix[e.Object]
	globalLabels := mat.GetGlobalLabels()
	assert.Equal(t, globalLabels["array"], e.arrayName)
}

func TestESeries_URLBuilder(t *testing.T) {
	tests := []struct {
		name     string
		apiPath  string
		systemID string
		filters  []string
		expected string
	}{
		{
			name:     "volume_query",
			apiPath:  "storage-systems/{array_id}/volumes",
			systemID: "600a098000f63714000000005e5cf5d2",
			expected: "storage-systems/600a098000f63714000000005e5cf5d2/volumes",
		},
		{
			name:     "controller_query",
			apiPath:  "storage-systems/{array_id}/controllers",
			systemID: "test-system-123",
			expected: "storage-systems/test-system-123/controllers",
		},
		{
			name:     "with_filters",
			apiPath:  "storage-systems/{array_id}/volumes",
			systemID: "test-sys",
			filters:  []string{"type=volume", "status=optimal"},
			expected: "storage-systems/test-sys/volumes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := rest.NewURLBuilder().
				APIPath(tt.apiPath).
				ArrayID(tt.systemID)

			// Note: Filters method is not exported, skipping for now

			result := builder.Build()
			assert.Equal(t, result, tt.expected)
		})
	}
}

func TestESeries_MetricExtraction(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	e := newESeries("Volume", "volume.yaml")

	mat := e.Matrix[e.Object]
	mat.SetGlobalLabel("array", e.arrayName)

	// Load test data
	pollData := JSONToGson("testdata/volume.json", false)

	// Poll data
	_ = e.pollData(mat, pollData)

	// Get first instance
	instances := mat.GetInstances()
	if len(instances) == 0 {
		t.Error("should have at least one instance")
		return
	}

	// Get first instance key
	var firstKey string
	for key := range instances {
		firstKey = key
		break
	}

	instance := mat.GetInstance(firstKey)
	assert.NotNil(t, instance)

	// Verify metric values are numeric and valid
	metrics := mat.GetMetrics()
	for metricName, metric := range metrics {
		value, ok := metric.GetValueFloat64(instance)
		if ok && value < 0 {
			t.Errorf("metric %s should be >= 0", metricName)
		}
	}
}

func TestESeries_LabelMapping(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	e := newESeries("Volume", "volume.yaml")

	// Verify counter mappings from template
	expectedMappings := map[string]string{
		"worldWideName":    "wwn",
		"sourceController": "source_controller",
		"action":           "action",
		"label":            "label",
		"name":             "volume",
		"offline":          "offline",
		"protectionType":   "protection_type",
		"raidLevel":        "raid_level",
		"status":           "status",
		"thinProvisioned":  "thin_provisioned",
		"volumeUse":        "volume_use",
	}

	for apiField, expectedDisplay := range expectedMappings {
		displayName, exists := e.Prop.InstanceLabels[apiField]
		if exists {
			assert.Equal(t, displayName, expectedDisplay)
		}
	}

	// Verify metric mappings
	expectedMetrics := map[string]string{
		"capacity":         "reported_capacity",
		"totalSizeInBytes": "allocated_capacity",
	}

	for apiField, expectedLabel := range expectedMetrics {
		metric, exists := e.Prop.Metrics[apiField]
		if !exists {
			t.Errorf("metric %s should exist", apiField)
		}
		if exists {
			assert.Equal(t, metric.Label, expectedLabel)
		}
	}
}

// Helper function to create and initialize a new ESeries collector for testing
func newESeries(object string, path string) *ESeries {
	var err error
	opts := options.New(options.WithConfPath("../../../conf"))
	opts.Poller = pollerName
	opts.HomePath = "testdata"
	opts.IsTest = true

	ac := collector.New("ESeries", object, opts, Params(object, path), nil, conf.Remote{})
	e := ESeries{}
	err = e.Init(ac)
	if err != nil {
		panic(err)
	}
	return &e
}
