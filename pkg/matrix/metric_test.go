package matrix

import (
	"github.com/netapp/harvest/v2/pkg/logging"
	"math"
	"testing"
)

type matrixOp string

const (
	sameInstance      = matrixOp("sameInstance")
	deleteInstance    = matrixOp("deleteInstance")
	addInstance       = matrixOp("addInstance")
	addDeleteInstance = matrixOp("addDeleteInstance")
)

type instNames struct {
	prev []string
	cur  []string
}

var instanceNames = map[matrixOp]instNames{
	sameInstance: {
		prev: []string{"A"},
		cur:  []string{"A"},
	},
	deleteInstance: {
		prev: []string{"A", "B"},
		cur:  []string{"A"},
	},
	addInstance: {
		prev: []string{"A"},
		cur:  []string{"A", "B"},
	},
	addDeleteInstance: {
		prev: []string{"A"},
		cur:  []string{"C"},
	},
}

func setupMatrix(previousRaw float64, currentRaw float64, mop matrixOp) (*Matrix, *Matrix) {
	m := New("Test", "test", "test")
	speed, _ := m.NewMetricFloat64("speed")
	names := instanceNames[mop]
	for _, instanceName := range names.prev {
		instance, _ := m.NewInstance(instanceName)
		_ = speed.SetValueFloat64(instance, previousRaw)
	}

	m1 := New("Test", "test", "test")
	speed1, _ := m1.NewMetricFloat64("speed")

	for _, instanceName := range names.cur {
		instance, _ := m1.NewInstance(instanceName)
		_ = speed1.SetValueFloat64(instance, currentRaw)
	}
	return m, m1
}

func setupMatrixAdv(previousRaw []float64, currentRaw []float64) (*Matrix, *Matrix) {
	prevMat := New("Test", "test", "test")
	averageLatency, _ := prevMat.NewMetricFloat64("average_latency")
	totalOps, _ := prevMat.NewMetricFloat64("total_ops")
	names := instanceNames[sameInstance]
	for _, instanceName := range names.prev {
		instance, _ := prevMat.NewInstance(instanceName)
		_ = averageLatency.SetValueFloat64(instance, previousRaw[0])
		_ = totalOps.SetValueFloat64(instance, previousRaw[1])
	}

	currentMat := New("Test", "test", "test")
	averageLatency1, _ := currentMat.NewMetricFloat64("average_latency")
	totalOps1, _ := currentMat.NewMetricFloat64("total_ops")
	for _, instanceName := range names.cur {
		instance, _ := currentMat.NewInstance(instanceName)
		_ = averageLatency1.SetValueFloat64(instance, currentRaw[0])
		_ = totalOps1.SetValueFloat64(instance, currentRaw[1])
	}
	return prevMat, currentMat
}

type test struct {
	name    string
	curRaw  float64
	prevRaw float64
	cooked  []float64
	scalar  uint
	skips   int
	record  []bool
}

type testAdv struct {
	name      string
	curRaw    []float64
	prevRaw   []float64
	cooked    []float64
	skips     int
	threshold int
	record    []bool
}

func TestMetricFloat64_Delta(t *testing.T) {
	testDelta(t, sameInstance)
	testDelta(t, deleteInstance)

	tests2 := []test{
		{curRaw: 10, prevRaw: 10, cooked: []float64{0, 10}, skips: 1, record: []bool{true, false}, name: "no increase AddInstance"},
		{curRaw: 20, prevRaw: 10, cooked: []float64{10, 20}, skips: 1, record: []bool{true, false}, name: "normal increase AddInstance"},
		{curRaw: 10, prevRaw: 20, cooked: []float64{-10, 10}, skips: 2, record: []bool{false, false}, name: "bug negative AddInstance"},
		{curRaw: 10, prevRaw: 0, cooked: []float64{10, 10}, skips: 2, record: []bool{false, false}, name: "bug zeroPrev AddInstance"},
		{curRaw: 0, prevRaw: 10, cooked: []float64{-10, 0}, skips: 2, record: []bool{false, false}, name: "bug zeroCur AddInstance"},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			previous, current := setupMatrix(tt.prevRaw, tt.curRaw, addInstance)
			skips, err := current.Delta("speed", previous, logging.Get())
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
			previous, current := setupMatrix(tt.prevRaw, tt.curRaw, addDeleteInstance)
			skips, err := current.Delta("speed", previous, logging.Get())
			matrixTest(t, tt, current, skips, err)
		})
	}
}

func testDelta(t *testing.T, op matrixOp) {
	tests := []test{
		{curRaw: 10, prevRaw: 10, cooked: []float64{0}, skips: 0, record: []bool{true}, name: "no increase"},
		{curRaw: 20, prevRaw: 10, cooked: []float64{10}, skips: 0, record: []bool{true}, name: "normal increase"},
		{curRaw: 10, prevRaw: 20, cooked: []float64{-10}, skips: 1, record: []bool{false}, name: "bug negative"},
		{curRaw: 10, prevRaw: 0, cooked: []float64{10}, skips: 1, record: []bool{false}, name: "bug zeroPrev"},
		{curRaw: 0, prevRaw: 10, cooked: []float64{-10}, skips: 1, record: []bool{false}, name: "bug zeroCur"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+string(op), func(t *testing.T) {
			previous, current := setupMatrix(tt.prevRaw, tt.curRaw, op)
			skips, err := current.Delta("speed", previous, logging.Get())
			matrixTest(t, tt, current, skips, err)
		})
	}
}

func TestMetricFloat64_Divide(t *testing.T) {
	testsAdv := []testAdv{
		{prevRaw: []float64{10, 5}, curRaw: []float64{20, 10}, cooked: []float64{2}, skips: 0, record: []bool{true}, name: "normal"},
		{prevRaw: []float64{10, 5}, curRaw: []float64{20, 5}, cooked: []float64{0}, skips: 0, record: []bool{true}, name: "allow zero den"},
		{prevRaw: []float64{10, 5}, curRaw: []float64{10, 5}, cooked: []float64{0}, skips: 0, record: []bool{true}, name: "allow zero both"},
		{prevRaw: []float64{10, 5}, curRaw: []float64{20, 0}, cooked: []float64{10}, skips: 1, record: []bool{false}, name: "bug negative den"},
		{prevRaw: []float64{10, 5}, curRaw: []float64{5, 10}, cooked: []float64{-5}, skips: 1, record: []bool{false}, name: "bug negative num"},
		{prevRaw: []float64{20, 5}, curRaw: []float64{10, 0}, cooked: []float64{-10}, skips: 1, record: []bool{false}, name: "bug negative both"},
	}

	for _, tt := range testsAdv {
		t.Run(tt.name, func(t *testing.T) {
			prevMat, curMat := setupMatrixAdv(tt.prevRaw, tt.curRaw)
			for k := range curMat.GetMetrics() {
				_, err := curMat.Delta(k, prevMat, logging.Get())
				if err != nil {
					t.Error("unexpected error", err)
					return
				}
			}
			skips, err := curMat.Divide("average_latency", "total_ops", logging.Get())
			matrixTestAdv(t, tt, curMat, skips, err)
		})
	}
}

func TestMetricFloat64_DivideWithThreshold(t *testing.T) {
	testsAdv := []testAdv{
		{prevRaw: []float64{10, 5}, curRaw: []float64{20, 10}, threshold: 1, cooked: []float64{2}, skips: 0, record: []bool{true}, name: "normal"},
		{prevRaw: []float64{10, 5}, curRaw: []float64{20, 10}, threshold: 15, cooked: []float64{0}, skips: 0, record: []bool{true}, name: "normal < threshold"},
		{prevRaw: []float64{10, 5}, curRaw: []float64{20, 0}, threshold: 1, cooked: []float64{10}, skips: 1, record: []bool{false}, name: "bug negative den"},
		{prevRaw: []float64{10, 5}, curRaw: []float64{5, 10}, threshold: 1, cooked: []float64{-5}, skips: 1, record: []bool{false}, name: "bug negative num"},
		{prevRaw: []float64{20, 5}, curRaw: []float64{10, 0}, threshold: 1, cooked: []float64{-10}, skips: 1, record: []bool{false}, name: "bug negative both"},
		{prevRaw: []float64{10, 10}, curRaw: []float64{20, 10}, threshold: 0, cooked: []float64{math.Inf(1)}, skips: 0, record: []bool{true}, name: "allow no threshold"},
		{prevRaw: []float64{10, 5}, curRaw: []float64{10, 5}, threshold: 5, cooked: []float64{0}, skips: 0, record: []bool{true}, name: "allow zero both"},
	}

	for _, tt := range testsAdv {
		t.Run(tt.name, func(t *testing.T) {
			prevMat, curMat := setupMatrixAdv(tt.prevRaw, tt.curRaw)
			for k := range curMat.GetMetrics() {
				_, err := curMat.Delta(k, prevMat, logging.Get())
				if err != nil {
					t.Error("unexpected error", err)
					return
				}
			}
			skips, err := curMat.DivideWithThreshold("average_latency", "total_ops", tt.threshold, logging.Get())
			matrixTestAdv(t, tt, curMat, skips, err)
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
			_, current := setupMatrix(tt.prevRaw, tt.curRaw, sameInstance)
			skips, err := current.MultiplyByScalar("speed", tt.scalar, logging.Get())
			matrixTest(t, tt, current, skips, err)
		})
	}
}

func matrixTestAdv(t *testing.T, tt testAdv, cur *Matrix, skips int, err error) {
	if err != nil {
		t.Error("unexpected error", err)
		return
	}
	if skips != tt.skips {
		t.Errorf("skips expected = %d, got %d", tt.skips, skips)
	}

	cooked := cur.GetMetric("average_latency").values
	for i := range cooked {
		if cooked[i] != tt.cooked[i] {
			t.Errorf("cooked expected = %v, got %v", tt.cooked, cooked)
		}
	}

	record := cur.GetMetric("average_latency").GetRecords()
	for i := range record {
		if record[i] != tt.record[i] {
			t.Errorf("record expected = %t, got %t", tt.record, record)
		}
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
	cooked := cur.GetMetric("speed").values
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
