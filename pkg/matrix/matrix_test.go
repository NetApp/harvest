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
		name             string
		args             args
		maxInstanceIndex int
	}

	tests := []test{
		{"removeExistingKey", args{key: "A"}, 2},
		{"removeAbsentKey", args{key: "E"}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpMatrix()
			m.RemoveInstance(tt.args.key)
			maxIndex := 0
			for _, i := range m.GetInstances() {
				if i.index > maxIndex {
					maxIndex = i.index
				}
			}
			if maxIndex != tt.maxInstanceIndex {
				t.Errorf("expected = %d, got %d", tt.maxInstanceIndex, maxIndex)
			}
		})
	}
}
