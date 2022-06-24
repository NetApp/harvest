package matrix

import (
	"testing"
)

func setUpMatrix() *Matrix {
	m := New("TestRemoveInstance", "test", "test")
	speed, _ := m.NewMetricUint32("max_speed")
	length, _ := m.NewMetricFloat32("length_in_mm")
	instanceNames := []string{"A", "B", "C", "D"}
	for _, instanceName := range instanceNames {
		instance, _ := m.NewInstance(instanceName)
		_ = speed.SetValueInt32(instance, 10)
		_ = length.SetValueFloat32(instance, 1.0)
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
