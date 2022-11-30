package matrix

import (
	"github.com/netapp/harvest/v2/pkg/logging"
	"math"
	"testing"
)

type MatrixOp int64

const (
	SameInstance MatrixOp = iota
	DeleteInstance
	AddInstance
	AddDeleteInstance
)

func setupMatrix(previousRaw float64, currentRaw float64, mop MatrixOp) (*Matrix, *Matrix) {
	var instanceNamesPrev []string
	var instanceNamesCur []string
	m := New("Test", "test", "test")
	speed, _ := m.NewMetricFloat64("speed")
	switch mop {
	case SameInstance:
		instanceNamesPrev = []string{"A"}
		instanceNamesCur = []string{"A"}
	case DeleteInstance:
		instanceNamesPrev = []string{"A", "B"}
		instanceNamesCur = []string{"A"}
	case AddInstance:
		instanceNamesPrev = []string{"A"}
		instanceNamesCur = []string{"A", "B"}
	case AddDeleteInstance:
		instanceNamesPrev = []string{"A"}
		instanceNamesCur = []string{"C"}
	}
	for _, instanceName := range instanceNamesPrev {
		instance, _ := m.NewInstance(instanceName)
		_ = speed.SetValueFloat64(instance, previousRaw)
	}

	m1 := New("Test", "test", "test")
	speed1, _ := m1.NewMetricFloat64("speed")

	for _, instanceName := range instanceNamesCur {
		instance, _ := m1.NewInstance(instanceName)
		_ = speed1.SetValueFloat64(instance, currentRaw)
	}
	return m, m1
}

type test struct {
	name      string
	curRaw    float64
	prevRaw   float64
	cooked    []float64
	threshold int
	scalar    uint
	skips     int
	record    []bool
}

func TestMetricFloat64_Delta(t *testing.T) {
	tests := []test{
		{curRaw: 10, prevRaw: 10, cooked: []float64{0}, skips: 0, record: []bool{true}, name: "no increase SameInstance"},
		{curRaw: 20, prevRaw: 10, cooked: []float64{10}, skips: 0, record: []bool{true}, name: "normal increase SameInstance"},
		{curRaw: 10, prevRaw: 20, cooked: []float64{-10}, skips: 1, record: []bool{false}, name: "bug negative SameInstance"},
		{curRaw: 10, prevRaw: 0, cooked: []float64{10}, skips: 1, record: []bool{false}, name: "bug zeroPrev SameInstance"},
		{curRaw: 0, prevRaw: 10, cooked: []float64{-10}, skips: 1, record: []bool{false}, name: "bug zeroCur SameInstance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previous, current := setupMatrix(tt.prevRaw, tt.curRaw, SameInstance)
			skips, err := current.GetMetric("speed").Delta(previous.GetMetric("speed"), previous, current, logging.Get())
			matrixTest(t, tt, current, skips, err)
		})
	}

	tests1 := []test{
		{curRaw: 10, prevRaw: 10, cooked: []float64{0}, skips: 0, record: []bool{true}, name: "no increase DeleteInstance"},
		{curRaw: 20, prevRaw: 10, cooked: []float64{10}, skips: 0, record: []bool{true}, name: "normal increase DeleteInstance"},
		{curRaw: 10, prevRaw: 20, cooked: []float64{-10}, skips: 1, record: []bool{false}, name: "bug negative DeleteInstance"},
		{curRaw: 10, prevRaw: 0, cooked: []float64{10}, skips: 1, record: []bool{false}, name: "bug zeroPrev DeleteInstance"},
		{curRaw: 0, prevRaw: 10, cooked: []float64{-10}, skips: 1, record: []bool{false}, name: "bug zeroCur DeleteInstance"},
	}

	for _, tt := range tests1 {
		t.Run(tt.name, func(t *testing.T) {
			previous, current := setupMatrix(tt.prevRaw, tt.curRaw, DeleteInstance)
			skips, err := current.GetMetric("speed").Delta(previous.GetMetric("speed"), previous, current, logging.Get())
			matrixTest(t, tt, current, skips, err)
		})
	}

	tests2 := []test{
		{curRaw: 10, prevRaw: 10, cooked: []float64{0, 10}, skips: 1, record: []bool{true, false}, name: "no increase AddInstance"},
		{curRaw: 20, prevRaw: 10, cooked: []float64{10, 20}, skips: 1, record: []bool{true, false}, name: "normal increase AddInstance"},
		{curRaw: 10, prevRaw: 20, cooked: []float64{-10, 10}, skips: 2, record: []bool{false, false}, name: "bug negative AddInstance"},
		{curRaw: 10, prevRaw: 0, cooked: []float64{10, 10}, skips: 2, record: []bool{false, false}, name: "bug zeroPrev AddInstance"},
		{curRaw: 0, prevRaw: 10, cooked: []float64{-10, 0}, skips: 2, record: []bool{false, false}, name: "bug zeroCur AddInstance"},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			previous, current := setupMatrix(tt.prevRaw, tt.curRaw, AddInstance)
			skips, err := current.GetMetric("speed").Delta(previous.GetMetric("speed"), previous, current, logging.Get())
			matrixTest(t, tt, current, skips, err)
		})
	}

	tests3 := []test{
		{curRaw: 10, prevRaw: 10, cooked: []float64{10}, skips: 1, record: []bool{false}, name: "no increase AddDeleteInstance"},
		{curRaw: 20, prevRaw: 10, cooked: []float64{20}, skips: 1, record: []bool{false}, name: "normal increase AddDeleteInstance"},
		{curRaw: 10, prevRaw: 20, cooked: []float64{10}, skips: 1, record: []bool{false}, name: "bug negative AddDeleteInstance"},
		{curRaw: 10, prevRaw: 0, cooked: []float64{10}, skips: 1, record: []bool{false}, name: "bug zeroPrev AddDeleteInstance"},
		{curRaw: 0, prevRaw: 10, cooked: []float64{0}, skips: 1, record: []bool{false}, name: "bug zeroCur AddDeleteInstance"},
	}

	for _, tt := range tests3 {
		t.Run(tt.name, func(t *testing.T) {
			previous, current := setupMatrix(tt.prevRaw, tt.curRaw, AddDeleteInstance)
			skips, err := current.GetMetric("speed").Delta(previous.GetMetric("speed"), previous, current, logging.Get())
			matrixTest(t, tt, current, skips, err)
		})
	}
}

func TestMetricFloat64_Divide(t *testing.T) {
	tests := []test{
		{prevRaw: 20, curRaw: 10, cooked: []float64{2}, skips: 0, record: []bool{true}, name: "normal"},
		{prevRaw: -20, curRaw: 10, cooked: []float64{-2}, skips: 1, record: []bool{false}, name: "bug negative num"},
		{prevRaw: 20, curRaw: -10, cooked: []float64{-2}, skips: 1, record: []bool{false}, name: "bug negative den"},
		{prevRaw: -20, curRaw: -10, cooked: []float64{2}, skips: 1, record: []bool{false}, name: "bug negative both"},
		{prevRaw: 20, curRaw: 0, cooked: []float64{0}, skips: 0, record: []bool{true}, name: "allow zero den"},
		{prevRaw: 0, curRaw: 0, cooked: []float64{0}, skips: 0, record: []bool{true}, name: "allow zero both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numerator, denominator := setupMatrix(tt.prevRaw, tt.curRaw, SameInstance)
			skips, err := numerator.GetMetric("speed").Divide(denominator.GetMetric("speed"), logging.Get())
			matrixTest(t, tt, numerator, skips, err)
		})
	}
}

func TestMetricFloat64_DivideWithThreshold(t *testing.T) {
	tests := []test{
		{prevRaw: 20, curRaw: 10, threshold: 5, cooked: []float64{2}, skips: 0, record: []bool{true}, name: "normal"},
		{prevRaw: 20, curRaw: 10, threshold: 15, cooked: []float64{0}, skips: 0, record: []bool{true}, name: "normal < threshold"},
		{prevRaw: 20, curRaw: -10, threshold: 5, cooked: []float64{0}, skips: 1, record: []bool{false}, name: "bug negative den"},
		{prevRaw: -20, curRaw: -10, threshold: 5, cooked: []float64{0}, skips: 1, record: []bool{false}, name: "bug negative both"},
		{prevRaw: 20, curRaw: 0, threshold: 0, cooked: []float64{math.Inf(1)}, skips: 0, record: []bool{true}, name: "allow no threshold"},
		{prevRaw: 0, curRaw: 0, threshold: 5, cooked: []float64{0}, skips: 0, record: []bool{true}, name: "allow zero both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numerator, denominator := setupMatrix(tt.prevRaw, tt.curRaw, SameInstance)
			skips, err := numerator.GetMetric("speed").
				DivideWithThreshold(denominator.GetMetric("speed"), tt.threshold, logging.Get())
			matrixTest(t, tt, numerator, skips, err)
		})
	}
}

func TestMetricFloat64_MultiplyByScalar(t *testing.T) {
	tests := []test{
		{prevRaw: 10, curRaw: 10, scalar: 5, cooked: []float64{50}, skips: 0, record: []bool{true}, name: "normal"},
		{prevRaw: 10, curRaw: -10, scalar: 5, cooked: []float64{-50}, skips: 1, record: []bool{false}, name: "bug negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, current := setupMatrix(tt.prevRaw, tt.curRaw, SameInstance)
			skips, err := current.GetMetric("speed").MultiplyByScalar(tt.scalar, logging.Get())
			matrixTest(t, tt, current, skips, err)
		})
	}
}

func matrixTest(t *testing.T, tt test, cur *Matrix, skips int, err error) {
	if err != nil {
		t.Error("unexpected error", err)
		return
	}
	if skips != tt.skips {
		t.Errorf("skips expected = %d, got %d", tt.skips, skips)
	}
	cooked := cur.GetMetric("speed").GetValuesFloat64()
	for i := range cooked {
		if cooked[i] != tt.cooked[i] {
			t.Errorf("cooked expected = %v, got %v", tt.cooked, cooked)
		}
	}

	record := cur.GetMetric("speed").GetRecords()
	for i := range record {
		if record[i] != tt.record[i] {
			t.Errorf("record expected = %t, got %t", tt.record, record)
		}
	}
}

func TestMetricReset(t *testing.T) {
	m := New("Test", "test", "test")
	_, _ = m.NewInstance("task1")
	_, _ = m.NewMetricInt64("poll_time")
	_ = m.LazySetValueInt64("poll_time", "task1", 10)
	m.ResetInstance("task1")
	_, pass := m.LazyGetValueInt64("poll_time", "task1")
	if pass {
		t.Errorf("expected metric to be skipped but passed")
	}
}
