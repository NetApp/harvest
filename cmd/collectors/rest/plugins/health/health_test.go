package health

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"log/slog"
	"testing"
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
	if !ok {
		t.Fatal("expected resolutionData to have key " + matName)
	}

	// Check the count of instances in the resolution matrix
	if len(resMat.GetInstances()) != 2 {
		t.Fatalf("expected resolutionData to have 2 instances, got %d", len(resMat.GetInstances()))
	}

	// Check that previousData is correctly updated
	if len(h.previousData[matName].GetInstances()) != 1 {
		t.Fatalf("expected previousData to have 1 instance, got %d", len(h.previousData["testMatrix"].GetInstances()))
	}

	// Check the instances in the resolution matrix
	for _, instanceKey := range []string{"0", "1"} {
		resInstance := resMat.GetInstance(instanceKey)
		if resInstance == nil {
			t.Fatalf("expected resolutionData to have instance with index %s", instanceKey)
		}

		if label := resInstance.GetLabel("label" + instanceKey); label != "value"+instanceKey {
			t.Fatalf("expected instance label 'label%s' to be 'value%s', got '%s'", instanceKey, instanceKey, label)
		}
	}
}
