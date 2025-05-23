package volumetopmetrics

import (
	"testing"

	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

var globalDataMatrix *matrix.Matrix

type MockVolume struct {
	*TopMetrics
	testFilePath string
}

func (mv *MockVolume) fetchTopClients(_ *set.Set, _ *set.Set, _ string) ([]gjson.Result, error) {
	return collectors.InvokeRestCallWithTestFile(nil, "", mv.testFilePath)
}

func (mv *MockVolume) fetchTopFiles(_ *set.Set, _ *set.Set, _ string) ([]gjson.Result, error) {
	return collectors.InvokeRestCallWithTestFile(nil, "", mv.testFilePath)
}

func (mv *MockVolume) fetchVolumesWithActivityTrackingEnabled() (*set.Set, error) {
	va := set.New()
	va.Add("osc" + keyToken + "osc_vol01")
	va.Add("osc" + keyToken + "volharvest")
	return va, nil
}

func NewMockVolume(p *plugin.AbstractPlugin, testFilePath string) *MockVolume {
	v := &TopMetrics{AbstractPlugin: p}
	mockVolume := &MockVolume{
		TopMetrics:   v,
		testFilePath: testFilePath,
	}
	mockVolume.tracker = mockVolume
	return mockVolume
}

func setupMockDataMatrix() *matrix.Matrix {
	data := matrix.New("volume", "volume", "volume")
	instance1, _ := data.NewInstance("1")
	instance1.SetLabel("volume", "osc_vol01")
	instance1.SetLabel("svm", "osc")

	instance2, _ := data.NewInstance("2")
	instance2.SetLabel("volume", "volharvest")
	instance2.SetLabel("svm", "osc")

	readOpsMetric, _ := data.NewMetricFloat64("total_read_ops", "read_ops")
	readOpsMetric.SetValueFloat64(instance1, 1)
	readOpsMetric.SetValueFloat64(instance2, 241)

	writeOpsMetric, _ := data.NewMetricFloat64("total_write_ops", "write_ops")
	writeOpsMetric.SetValueFloat64(instance1, 100)
	writeOpsMetric.SetValueFloat64(instance2, 341)

	readDataMetric, _ := data.NewMetricFloat64("bytes_read", "read_data")
	readDataMetric.SetValueFloat64(instance1, 100000)
	readDataMetric.SetValueFloat64(instance2, 341000)

	writeDataMetric, _ := data.NewMetricFloat64("bytes_written", "write_data")
	writeDataMetric.SetValueFloat64(instance1, 100000)
	writeDataMetric.SetValueFloat64(instance2, 341000)
	return data
}

func init() {
	globalDataMatrix = setupMockDataMatrix()
}

func TestProcessTopClients(t *testing.T) {
	testCases := []struct {
		name          string
		matrixName    string
		testFilePath  string
		expectedCount int
	}{
		{"Client Read Ops", topClientReadOPSMatrix, "testdata/client_readops.json", 1},
		{"Client Write Ops", topClientWriteOPSMatrix, "testdata/client_writeops.json", 4},
		{"Client Read Data", topClientReadDataMatrix, "testdata/client_readdata.json", 1},
		{"Client Write Data", topClientWriteDataMatrix, "testdata/client_writedata.json", 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockVolume := NewMockVolume(&plugin.AbstractPlugin{}, tc.testFilePath)
			mockVolume.maxVolumeCount = 5

			err := mockVolume.InitAllMatrix()
			if err != nil {
				t.Errorf("InitAllMatrix should not return an error: %v", err)
			}

			data := globalDataMatrix

			metrics, err := mockVolume.processTopMetrics(data)
			if err != nil {
				return
			}

			err = mockVolume.processTopClients(metrics)
			if err != nil {
				t.Errorf("processTopClients should not return an error: %v", err)
			}

			resultMatrix := mockVolume.data[tc.matrixName]

			if resultMatrix == nil {
				t.Errorf("%s Matrix should be initialized", tc.matrixName)
			}
			if len(resultMatrix.GetInstances()) != tc.expectedCount {
				t.Errorf("%s Matrix should have %d instance(s), got %d", tc.matrixName, tc.expectedCount, len(resultMatrix.GetInstances()))
			}
		})
	}
}

func TestProcessTopFiles(t *testing.T) {
	testCases := []struct {
		name          string
		matrixName    string
		testFilePath  string
		expectedCount int
	}{
		{"File Read Ops", topFileReadOPSMatrix, "testdata/file_readops.json", 1},
		{"File Write Ops", topFileWriteOPSMatrix, "testdata/file_writeops.json", 6},
		{"File Read Data", topFileReadDataMatrix, "testdata/file_readdata.json", 1},
		{"File Write Data", topFileWriteDataMatrix, "testdata/file_writedata.json", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockVolume := NewMockVolume(&plugin.AbstractPlugin{}, tc.testFilePath)
			mockVolume.maxVolumeCount = 5

			err := mockVolume.InitAllMatrix()
			if err != nil {
				t.Errorf("InitAllMatrix should not return an error: %v", err)
			}

			data := globalDataMatrix

			metrics, err := mockVolume.processTopMetrics(data)
			if err != nil {
				return
			}

			err = mockVolume.processTopFiles(metrics)
			if err != nil {
				t.Errorf("processTopClients should not return an error: %v", err)
			}

			resultMatrix := mockVolume.data[tc.matrixName]

			if resultMatrix == nil {
				t.Errorf("%s Matrix should be initialized", tc.matrixName)
			}
			if len(resultMatrix.GetInstances()) != tc.expectedCount {
				t.Errorf("%s Matrix should have %d instance(s), got %d", tc.matrixName, tc.expectedCount, len(resultMatrix.GetInstances()))
			}
		})
	}
}
