package health

import (
	"log/slog"
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

func TestEndPoll(t *testing.T) {
	// Create a new Health struct
	h := &Health{AbstractPlugin: plugin.New("health", nil, nil, nil, "health", nil)}
	h.SLogger = slog.Default()
	h.data = make(map[string]*matrix.Matrix)
	h.previousData = make(map[string]*matrix.Matrix)
	_ = h.InitAllMatrix()

	matName := "health_lif"

	// Initialize some test data
	prevMat := matrix.New("UUID", "object", "identifier")
	prevInstance1, _ := prevMat.NewInstance("0")
	prevInstance1.SetLabel("label0", "value0")
	prevInstance2, _ := prevMat.NewInstance("1")
	prevInstance2.SetLabel("label1", "value1")
	h.previousData[matName] = prevMat

	curMat := matrix.New("UUID", "object", "identifier")
	curInstance, _ := curMat.NewInstance("2")
	curInstance.SetLabel("label2", "value2")
	h.data[matName] = curMat

	h.generateResolutionMetrics()

	// Check that resolutionData has the expected values
	resMat, ok := h.resolutionData[matName]
	assert.True(t, ok)

	// Check the count of instances in the resolution matrix
	assert.Equal(t, len(resMat.GetInstances()), 2)

	// Check that previousData is correctly updated
	assert.Equal(t, len(h.previousData[matName].GetInstances()), 1)

	// Check the instances in the resolution matrix
	for _, instanceKey := range []string{"0", "1"} {
		resInstance := resMat.GetInstance(instanceKey)
		assert.NotNil(t, resInstance)

		expectedLabel := "value" + instanceKey
		actualLabel := resInstance.GetLabel("label" + instanceKey)
		assert.Equal(t, actualLabel, expectedLabel)
	}
}
