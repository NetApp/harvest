package volumetopclients

import (
	"testing"

	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/tidwall/gjson"
)

var globalDataMatrix *matrix.Matrix

type MockVolume struct {
	*Volume
	testFilePath string
}

func (mv *MockVolume) fetchTopClients(_ []string, _ []string, _ string) ([]gjson.Result, error) {
	return collectors.InvokeRestCallWithTestFile(nil, "", nil, mv.testFilePath)
}

func (mv *MockVolume) fetchVolumesWithActivityTrackingEnabled() (*set.Set, error) {
	va := set.New()
	va.Add("osc" + keyToken + "osc_vol01")
	va.Add("osc" + keyToken + "volharvest")
	return va, nil
}

func NewMockVolume(p *plugin.AbstractPlugin, testFilePath string) *MockVolume {
	v := &Volume{AbstractPlugin: p}
	mockVolume := &MockVolume{
		Volume:       v,
		testFilePath: testFilePath,
	}
	mockVolume.volumeInterface = mockVolume
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

	readOpsMetric, _ := data.NewMetricFloat64("total_read_ops")
	_ = readOpsMetric.SetValueFloat64(instance1, 1)
	_ = readOpsMetric.SetValueFloat64(instance2, 241)

	writeOpsMetric, _ := data.NewMetricFloat64("total_write_ops")
	_ = writeOpsMetric.SetValueFloat64(instance1, 100)
	_ = writeOpsMetric.SetValueFloat64(instance2, 341)

	readDataMetric, _ := data.NewMetricFloat64("bytes_read")
	_ = readDataMetric.SetValueFloat64(instance1, 100000)
	_ = readDataMetric.SetValueFloat64(instance2, 341000)

	writeDataMetric, _ := data.NewMetricFloat64("bytes_written")
	_ = writeDataMetric.SetValueFloat64(instance1, 100000)
	_ = writeDataMetric.SetValueFloat64(instance2, 341000)
	return data
}

func init() {
	globalDataMatrix = setupMockDataMatrix()
}

func TestProcessTopClients(t *testing.T) {
	testCases := []struct {
		name          string
		metric        string
		matrixName    string
		testFilePath  string
		expectedCount int
	}{
		{"Read Ops", "iops.read", topClientReadOPSMatrix, "testdata/readops.json", 1},
		{"Write Ops", "iops.write", topClientWriteOPSMatrix, "testdata/writeops.json", 4},
		{"Read Data", "throughput.read", topClientReadDataMatrix, "testdata/readdata.json", 1},
		{"Write Data", "throughput.write", topClientWriteDataMatrix, "testdata/writedata.json", 3},
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

			err = mockVolume.processTopClients(data)
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
