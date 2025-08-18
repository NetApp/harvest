package matrix

import (
	"log/slog"
	"testing"
)

type matrixOp string

const (
	oneInstance       = matrixOp("oneInstance")
	twoInstance       = matrixOp("twoInstance")
	deleteInstance    = matrixOp("deleteInstance")
	addInstance       = matrixOp("addInstance")
	addDeleteInstance = matrixOp("addDeleteInstance")
)

type instNames struct {
	prev []string
	cur  []string
}

var instanceNames = map[matrixOp]instNames{
	oneInstance: {
		prev: []string{"A"},
		cur:  []string{"A"},
	},
	twoInstance: {
		prev: []string{"A", "B"},
		cur:  []string{"A", "B"},
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
		speed.SetValueFloat64(instance, previousRaw)
	}

	m1 := New("Test", "test", "test")
	speed1, _ := m1.NewMetricFloat64("speed")

	for _, instanceName := range names.cur {
		instance, _ := m1.NewInstance(instanceName)
		speed1.SetValueFloat64(instance, currentRaw)
	}
	return m, m1
}

func setupMatrixAdv(latency string, previousRaw []rawData, currentRaw []rawData, mop matrixOp) (*Matrix, *Matrix) {
	prevMat := New("Test", "test", "test")
	averageLatency, _ := prevMat.NewMetricFloat64(latency)
	totalOps, _ := prevMat.NewMetricFloat64("total_ops")
	timestamp, _ := prevMat.NewMetricFloat64("timestamp")
	names := instanceNames[mop]
	for i, instanceName := range names.prev {
		instance, _ := prevMat.NewInstance(instanceName)
		averageLatency.SetValueFloat64(instance, previousRaw[i].latency)
		totalOps.SetValueFloat64(instance, previousRaw[i].ops)
		timestamp.SetValueFloat64(instance, previousRaw[i].timestamp)
	}

	currentMat := New("Test", "test", "test")
	averageLatency1, _ := currentMat.NewMetricFloat64(latency)
	totalOps1, _ := currentMat.NewMetricFloat64("total_ops")
	timestamp1, _ := currentMat.NewMetricFloat64("timestamp")
	for i, instanceName := range names.cur {
		instance, _ := currentMat.NewInstance(instanceName)
		averageLatency1.SetValueFloat64(instance, currentRaw[i].latency)
		totalOps1.SetValueFloat64(instance, currentRaw[i].ops)
		timestamp1.SetValueFloat64(instance, currentRaw[i].timestamp)
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

type rawData struct {
	latency   float64
	ops       float64
	timestamp float64
}

type testAdv struct {
	name      string
	curRaw    []rawData
	prevRaw   []rawData
	cooked    []float64
	skips     int
	threshold int
	record    []bool
	matrixOp  matrixOp
	latency   string
}

func TestMetricFloat64_Delta(t *testing.T) {
	testDelta(t, oneInstance)
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
			skips, err := current.Delta("speed", previous, current, false, slog.Default())
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
			skips, err := current.Delta("speed", previous, current, false, slog.Default())
			matrixTest(t, tt, current, skips, err)
		})
	}
}

func TestMetricFloat64_Delta_PartialAggregation(t *testing.T) {
	tests := []struct {
		name                   string
		curRaw                 float64
		prevRaw                float64
		expectedSkips          int
		prevPartialAggregation bool
		currPartialAggregation bool
	}{
		{"No Partial Aggregation", 20, 10, 0, false, false},
		{"Previous Partial Aggregation", 20, 10, 1, true, false},
		{"Current Partial Aggregation", 20, 10, 1, false, true},
		{"Both Partial Aggregation", 20, 10, 1, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previous, current := setupMatrixForPartialAggregation(tt.prevRaw, tt.curRaw, tt.prevPartialAggregation, tt.currPartialAggregation)
			skips, err := current.Delta("speed", previous, current, false, slog.Default())
			if err != nil {
				t.Errorf("Delta method returned an error: %v", err)
			}
			if skips != tt.expectedSkips {
				t.Errorf("Expected %d skips, got %d", tt.expectedSkips, skips)
			}
		})
	}
}

// setupMatrixForPartialAggregation sets up two Matrix objects with one instance each, and marks them as partial aggregation instances based on the input flags.
func setupMatrixForPartialAggregation(prevRaw, curRaw float64, prevPartial, currPartial bool) (*Matrix, *Matrix) {
	// Create the previous Matrix with one instance
	prevMatrix := New("Test", "test", "test")
	prevSpeed, _ := prevMatrix.NewMetricFloat64("speed")
	prevInstance, _ := prevMatrix.NewInstance("A")
	prevSpeed.SetValueFloat64(prevInstance, prevRaw)
	if prevPartial {
		prevInstance.SetPartial(true)
	}

	// Create the current Matrix with one instance
	currMatrix := New("Test", "test", "test")
	currSpeed, _ := currMatrix.NewMetricFloat64("speed")
	currInstance, _ := currMatrix.NewInstance("A")
	currSpeed.SetValueFloat64(currInstance, curRaw)
	if currPartial {
		currInstance.SetPartial(true)
	}

	return prevMatrix, currMatrix
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
			skips, err := current.Delta("speed", previous, current, false, slog.Default())
			matrixTest(t, tt, current, skips, err)
		})
	}
}

func TestMetricFloat64_Divide(t *testing.T) {
	testsAdv := []testAdv{
		{prevRaw: []rawData{{10, 5, 60}}, curRaw: []rawData{{20, 10, 120}}, cooked: []float64{2}, skips: 0, matrixOp: oneInstance, record: []bool{true}, name: "normal"},
		{prevRaw: []rawData{{10, 5, 60}}, curRaw: []rawData{{20, 5, 120}}, cooked: []float64{0}, skips: 0, matrixOp: oneInstance, record: []bool{true}, name: "allow zero den"},
		{prevRaw: []rawData{{10, 5, 60}}, curRaw: []rawData{{10, 5, 120}}, cooked: []float64{0}, skips: 0, matrixOp: oneInstance, record: []bool{true}, name: "allow zero both"},
		{prevRaw: []rawData{{10, 5, 60}}, curRaw: []rawData{{20, 0, 120}}, cooked: []float64{10}, skips: 1, matrixOp: oneInstance, record: []bool{false}, name: "bug negative den"},
		{prevRaw: []rawData{{10, 5, 60}}, curRaw: []rawData{{5, 10, 120}}, cooked: []float64{-5}, skips: 1, matrixOp: oneInstance, record: []bool{false}, name: "bug negative num"},
		{prevRaw: []rawData{{20, 5, 60}}, curRaw: []rawData{{10, 0, 120}}, cooked: []float64{-10}, skips: 1, matrixOp: oneInstance, record: []bool{false}, name: "bug negative both"},
		{prevRaw: []rawData{{20, 5, 60}, {10, 5, 60}}, curRaw: []rawData{{10, 10, 120}, {20, 10, 120}}, cooked: []float64{-10, 2}, skips: 1, record: []bool{false, true}, matrixOp: twoInstance, name: "verify multiple instance"},
	}

	for _, tt := range testsAdv {
		t.Run(tt.name, func(t *testing.T) {
			latency := tt.latency
			if latency == "" {
				latency = "average_latency"
			}
			prevMat, curMat := setupMatrixAdv(latency, tt.prevRaw, tt.curRaw, tt.matrixOp)
			for k := range curMat.GetMetrics() {
				_, err := curMat.Delta(k, prevMat, curMat, false, slog.Default())
				if err != nil {
					t.Error("unexpected error", err)
					return
				}
			}
			skips, err := curMat.Divide(latency, "total_ops")
			matrixTestAdv(t, tt, curMat, skips, err, latency)
		})
	}
}

func TestMetricFloat64_DivideWithThreshold(t *testing.T) {
	testsAdv := []testAdv{
		{prevRaw: []rawData{{1000, 500, 60}}, curRaw: []rawData{{9000, 2500, 120}}, threshold: 10, cooked: []float64{4}, skips: 0, record: []bool{true}, matrixOp: oneInstance, name: "normal"},
		{prevRaw: []rawData{{1000, 500, 60}}, curRaw: []rawData{{9000, 1000, 120}}, threshold: 10, cooked: []float64{0}, skips: 0, record: []bool{true}, matrixOp: oneInstance, name: "normal < threshold"},
		{prevRaw: []rawData{{1000, 500, 60}}, curRaw: []rawData{{2000, 300, 120}}, threshold: 10, cooked: []float64{1000}, skips: 1, record: []bool{false}, matrixOp: oneInstance, name: "bug negative den"},
		{prevRaw: []rawData{{1000, 500, 60}}, curRaw: []rawData{{500, 1000, 120}}, threshold: 10, cooked: []float64{-500}, skips: 1, record: []bool{false}, matrixOp: oneInstance, name: "bug negative num"},
		{prevRaw: []rawData{{1000, 500, 60}}, curRaw: []rawData{{500, 300, 120}}, threshold: 10, cooked: []float64{-500}, skips: 1, record: []bool{false}, matrixOp: oneInstance, name: "bug negative both"},
		{prevRaw: []rawData{{1000, 500, 60}}, curRaw: []rawData{{2000, 500, 120}}, threshold: 10, cooked: []float64{0}, skips: 0, record: []bool{true}, matrixOp: oneInstance, name: "zero ops delta"},
		{prevRaw: []rawData{{1000, 500, 60}}, curRaw: []rawData{{1000, 500, 120}}, threshold: 10, cooked: []float64{0}, skips: 0, record: []bool{true}, matrixOp: oneInstance, name: "zero latency/ops delta"},
		{prevRaw: []rawData{{2000, 500, 60}, {1000, 5000, 60}}, curRaw: []rawData{{1000, 1000, 120}, {10000, 8000, 120}}, threshold: 10, cooked: []float64{-1000, 3}, skips: 1, record: []bool{false, true}, matrixOp: twoInstance, name: "verify multiple instance"},
		{prevRaw: []rawData{{10, 5, 60}}, curRaw: []rawData{{20, 10, 120}}, threshold: 10, cooked: []float64{2}, skips: 0, record: []bool{true}, matrixOp: oneInstance, name: "no threshold check for optimal_point_latency", latency: "optimal_point_latency"},
	}

	for _, tt := range testsAdv {
		t.Run(tt.name, func(t *testing.T) {
			latency := tt.latency
			if latency == "" {
				latency = "average_latency"
			}
			prevMat, curMat := setupMatrixAdv(latency, tt.prevRaw, tt.curRaw, tt.matrixOp)
			cachedData := curMat.Clone(With{Data: true, Metrics: true, Instances: true, ExportInstances: true})

			for k := range curMat.GetMetrics() {
				_, err := curMat.Delta(k, prevMat, curMat, false, slog.Default())
				if err != nil {
					t.Error("unexpected error", err)
					return
				}
			}

			skips, err := curMat.DivideWithThreshold(latency, "total_ops", tt.threshold, cachedData, prevMat, "timestamp", slog.Default())
			matrixTestAdv(t, tt, curMat, skips, err, latency)
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
			_, current := setupMatrix(tt.prevRaw, tt.curRaw, oneInstance)
			skips, err := current.MultiplyByScalar("speed", tt.scalar)
			matrixTest(t, tt, current, skips, err)
		})
	}
}

func matrixTestAdv(t *testing.T, tt testAdv, cur *Matrix, skips int, err error, latency string) {
	if err != nil {
		t.Error("unexpected error", err)
		return
	}
	if skips != tt.skips {
		t.Errorf("skips expected = %d, got %d", tt.skips, skips)
	}

	cooked := cur.GetMetric(latency).values
	for i := range cooked {
		if cooked[i] != tt.cooked[i] {
			t.Errorf("cooked expected = %v, got %v", tt.cooked, cooked)
		}
	}

	record := cur.GetMetric(latency).GetRecords()
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
