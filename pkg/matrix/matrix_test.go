package matrix

import (
	"testing"
)

func setUpMatrix() *Matrix {
	m := New("TestRemoveInstance", "test", "test")
	speed, _ := m.NewMetricUint64("max_speed")
	instanceNames := []string{"A", "B", "C", "D"}
	for _, instanceName := range instanceNames {
		instance, _ := m.NewInstance(instanceName)
		_ = speed.SetValueInt64(instance, 10)
	}
	return m
}

func TestMatrix_RemoveInstance(t *testing.T) {

	type args struct {
		key string
	}

	type test struct {
		name          string
		args          args
		instanceCount int
	}

	tests := []test{
		{"removeExistingKey", args{key: "A"}, 3},
		{"removeAbsentKey", args{key: "E"}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpMatrix()
			m.RemoveInstance(tt.args.key)
			instanceCount := len(m.GetInstances())
			if instanceCount != tt.instanceCount {
				t.Errorf("expected = %d, got %d", tt.instanceCount, instanceCount)
			}
		})
	}
}
