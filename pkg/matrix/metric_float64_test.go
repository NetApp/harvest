package matrix

import (
	"github.com/netapp/harvest/v2/pkg/logging"
	"math"
	"testing"
)

func setupMatrix(previousRaw float64, currentRaw float64) (*Matrix, *Matrix) {
	m := New("Test", "test", "test")
	speed, _ := m.NewMetricFloat64("speed")
	instanceNames := []string{"A"}
	for _, instanceName := range instanceNames {
		instance, _ := m.NewInstance(instanceName)
		_ = speed.SetValueFloat64(instance, previousRaw)
	}

	m1 := New("Test", "test", "test")
	speed1, _ := m1.NewMetricFloat64("speed")
	instanceNames = []string{"A"}
	for _, instanceName := range instanceNames {
		instance, _ := m1.NewInstance(instanceName)
		_ = speed1.SetValueFloat64(instance, currentRaw)
	}
	return m, m1
}

type test struct {
	name      string
	curRaw    float64
	prevRaw   float64
	cooked    float64
	threshold int
	scalar    uint
	skips     int
	export    bool
}

func TestMetricFloat64_Delta(t *testing.T) {
	tests := []test{
		{curRaw: 10, prevRaw: 10, cooked: 0, skips: 0, export: true, name: "no increase"},
		{curRaw: 20, prevRaw: 10, cooked: 10, skips: 0, export: true, name: "normal increase"},
		{curRaw: 10, prevRaw: 20, cooked: -10, skips: 1, export: false, name: "bug negative"},
		{curRaw: 10, prevRaw: 0, cooked: 10, skips: 1, export: false, name: "bug zeroPrev"},
		{curRaw: 0, prevRaw: 10, cooked: -10, skips: 1, export: false, name: "bug zeroCur"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previous, current := setupMatrix(tt.prevRaw, tt.curRaw)
			skips, err := current.GetMetric("speed").Delta(previous.GetMetric("speed"), logging.Get())
			matrixTest(t, tt, current, skips, err)
		})
	}
}

func TestMetricFloat64_Divide(t *testing.T) {
	tests := []test{
		{prevRaw: 20, curRaw: 10, cooked: 2, skips: 0, export: true, name: "normal"},
		{prevRaw: -20, curRaw: 10, cooked: -2, skips: 1, export: false, name: "bug negative num"},
		{prevRaw: 20, curRaw: -10, cooked: -2, skips: 1, export: false, name: "bug negative den"},
		{prevRaw: -20, curRaw: -10, cooked: 2, skips: 1, export: false, name: "bug negative both"},
		{prevRaw: 20, curRaw: 0, cooked: 0, skips: 0, export: true, name: "allow zero den"},
		{prevRaw: 0, curRaw: 0, cooked: 0, skips: 0, export: true, name: "allow zero both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numerator, denominator := setupMatrix(tt.prevRaw, tt.curRaw)
			skips, err := numerator.GetMetric("speed").Divide(denominator.GetMetric("speed"), logging.Get())
			matrixTest(t, tt, numerator, skips, err)
		})
	}
}

func TestMetricFloat64_DivideWithThreshold(t *testing.T) {
	tests := []test{
		{prevRaw: 20, curRaw: 10, threshold: 5, cooked: 2, skips: 0, export: true, name: "normal"},
		{prevRaw: 20, curRaw: 10, threshold: 15, cooked: 0, skips: 0, export: true, name: "normal < threshold"},
		{prevRaw: 20, curRaw: -10, threshold: 5, cooked: 0, skips: 1, export: false, name: "bug negative den"},
		{prevRaw: -20, curRaw: -10, threshold: 5, cooked: 0, skips: 1, export: false, name: "bug negative both"},
		{prevRaw: 20, curRaw: 0, threshold: 0, cooked: math.Inf(1), skips: 0, export: true, name: "allow no threshold"},
		{prevRaw: 0, curRaw: 0, threshold: 5, cooked: 0, skips: 0, export: true, name: "allow zero both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numerator, denominator := setupMatrix(tt.prevRaw, tt.curRaw)
			skips, err := numerator.GetMetric("speed").
				DivideWithThreshold(denominator.GetMetric("speed"), tt.threshold, logging.Get())
			matrixTest(t, tt, numerator, skips, err)
		})
	}
}

func TestMetricFloat64_MultiplyByScalar(t *testing.T) {
	tests := []test{
		{prevRaw: 10, curRaw: 10, scalar: 5, cooked: 50, skips: 0, export: true, name: "normal"},
		{prevRaw: 10, curRaw: -10, scalar: 5, cooked: -50, skips: 1, export: false, name: "bug negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, current := setupMatrix(tt.prevRaw, tt.curRaw)
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
	cooked := cur.GetMetric("speed").GetValuesFloat64()["A"]
	if cooked != tt.cooked {
		t.Errorf("cooked expected = %v, got %v", tt.cooked, cooked)
	}
	shallSkip := cur.GetMetric("speed").GetSkips()["A"]
	if !shallSkip != tt.export {
		t.Errorf("export expected = %t, got %t", tt.export, !shallSkip)
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
