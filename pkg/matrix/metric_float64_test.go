package matrix

import (
	"github.com/netapp/harvest/v2/pkg/logging"
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

func TestMetricFloat64_Delta(t *testing.T) {
	// When previous == current and previous > 0 and current > 0
	//      then don't skip and pass is true
	previous, current := setupMatrix(10, 10)
	skip, _ := current.GetMetric("speed").Delta(previous.GetMetric("speed"), logging.Get())
	if skip != 0 {
		t.Errorf("expected = %d, got %d", 0, skip)
	}
	v := current.GetMetric("speed").GetValuesFloat64()[0]
	if v != 0 {
		t.Errorf("expected = %v, got %v", 0, v)
	}
	pass := current.GetMetric("speed").GetPass()[0]
	if !pass {
		t.Errorf("expected = %t, got %t", pass, !pass)
	}

	// When previous < current and previous > 0 and current > 0
	//      then don't skip and pass is true
	previous, current = setupMatrix(10, 20)
	skip, _ = current.GetMetric("speed").Delta(previous.GetMetric("speed"), logging.Get())
	if skip != 0 {
		t.Errorf("expected = %d, got %d", 0, skip)
	}
	v = current.GetMetric("speed").GetValuesFloat64()[0]
	if v != 10 {
		t.Errorf("expected = %v, got %v", 10, v)
	}
	pass = current.GetMetric("speed").GetPass()[0]
	if !pass {
		t.Errorf("expected = %t, got %t", pass, !pass)
	}

	// When previous > current and previous > 0 and current > 0
	//      then skip and pass is false
	previous, current = setupMatrix(20, 10)
	skip, _ = current.GetMetric("speed").Delta(previous.GetMetric("speed"), logging.Get())
	if skip != 1 {
		t.Errorf("expected = %d, got %d", 1, skip)
	}
	v = current.GetMetric("speed").GetValuesFloat64()[0]
	if v != -10 {
		t.Errorf("expected = %v, got %v", -10, v)
	}
	pass = current.GetMetric("speed").GetPass()[0]
	if pass {
		t.Errorf("expected = %t, got %t", !pass, pass)
	}

	// When previous == 0 current > 0
	//      then skip and pass is false
	previous, current = setupMatrix(0, 10)
	skip, _ = current.GetMetric("speed").Delta(previous.GetMetric("speed"), logging.Get())
	if skip != 1 {
		t.Errorf("expected = %d, got %d", 1, skip)
	}
	v = current.GetMetric("speed").GetValuesFloat64()[0]
	if v != 10 {
		t.Errorf("expected = %v, got %v", 10, v)
	}
	pass = current.GetMetric("speed").GetPass()[0]
	if pass {
		t.Errorf("expected = %t, got %t", !pass, pass)
	}
}

func TestMetricFloat64_Divide(t *testing.T) {
	// When numerator == denominator and numerator > 0 and denominator > 0
	//      then don't skip and pass is true
	numerator, denominator := setupMatrix(20, 10)
	skip, _ := numerator.GetMetric("speed").Divide(denominator.GetMetric("speed"), logging.Get())
	if skip != 0 {
		t.Errorf("expected = %d, got %d", 0, skip)
	}
	v := numerator.GetMetric("speed").GetValuesFloat64()[0]
	if v != 2 {
		t.Errorf("expected = %v, got %v", 2, v)
	}
	pass := numerator.GetMetric("speed").GetPass()[0]
	if !pass {
		t.Errorf("expected = %t, got %t", pass, !pass)
	}

	// When numerator < 0 and denominator > 0
	//      then skip and pass is false
	numerator, denominator = setupMatrix(-20, 10)
	skip, _ = numerator.GetMetric("speed").Divide(denominator.GetMetric("speed"), logging.Get())
	if skip != 1 {
		t.Errorf("expected = %d, got %d", 1, skip)
	}
	v = numerator.GetMetric("speed").GetValuesFloat64()[0]
	if v != -2 {
		t.Errorf("expected = %v, got %v", -2, v)
	}
	pass = numerator.GetMetric("speed").GetPass()[0]
	if pass {
		t.Errorf("expected = %t, got %t", !pass, pass)
	}

	// When numerator > 0 denominator < 0
	//      then skip and pass is false
	numerator, denominator = setupMatrix(20, -10)
	skip, _ = numerator.GetMetric("speed").Divide(denominator.GetMetric("speed"), logging.Get())
	if skip != 1 {
		t.Errorf("expected = %d, got %d", 1, skip)
	}
	v = numerator.GetMetric("speed").GetValuesFloat64()[0]
	if v != -2 {
		t.Errorf("expected = %v, got %v", -2, v)
	}
	pass = numerator.GetMetric("speed").GetPass()[0]
	if pass {
		t.Errorf("expected = %t, got %t", !pass, pass)
	}

	// When numerator < 0 denominator < 0
	//      then skip and pass is false
	numerator, denominator = setupMatrix(-20, -10)
	skip, _ = numerator.GetMetric("speed").Divide(denominator.GetMetric("speed"), logging.Get())
	if skip != 1 {
		t.Errorf("expected = %d, got %d", 1, skip)
	}
	v = numerator.GetMetric("speed").GetValuesFloat64()[0]
	if v != 2 {
		t.Errorf("expected = %v, got %v", 2, v)
	}
	pass = numerator.GetMetric("speed").GetPass()[0]
	if pass {
		t.Errorf("expected = %t, got %t", !pass, pass)
	}
}

func TestMetricFloat64_DivideWithThreshold(t *testing.T) {
	// When numerator == denominator and numerator > 0 and denominator > 0 and denominator > threshold
	//      then don't skip and pass is true
	numerator, denominator := setupMatrix(20, 10)
	skip, _ := numerator.GetMetric("speed").DivideWithThreshold(denominator.GetMetric("speed"), 5, logging.Get())
	if skip != 0 {
		t.Errorf("expected = %d, got %d", 0, skip)
	}
	v := numerator.GetMetric("speed").GetValuesFloat64()[0]
	if v != 2 {
		t.Errorf("expected = %v, got %v", 2, v)
	}
	pass := numerator.GetMetric("speed").GetPass()[0]
	if !pass {
		t.Errorf("expected = %t, got %t", pass, !pass)
	}

	// When numerator == denominator and numerator > 0 and denominator > 0 and denominator < threshold
	//      then don't skip and pass is true
	numerator, denominator = setupMatrix(20, 10)
	skip, _ = numerator.GetMetric("speed").DivideWithThreshold(denominator.GetMetric("speed"), 15, logging.Get())
	if skip != 0 {
		t.Errorf("expected = %d, got %d", 0, skip)
	}
	v = numerator.GetMetric("speed").GetValuesFloat64()[0]
	if v != 0 {
		t.Errorf("expected = %v, got %v", 0, v)
	}
	pass = numerator.GetMetric("speed").GetPass()[0]
	if !pass {
		t.Errorf("expected = %t, got %t", pass, !pass)
	}

	// When numerator < 0 and denominator > 0 and denominator > threshold
	//      then skip and pass is false
	numerator, denominator = setupMatrix(-20, 10)
	skip, _ = numerator.GetMetric("speed").DivideWithThreshold(denominator.GetMetric("speed"), 5, logging.Get())
	if skip != 1 {
		t.Errorf("expected = %d, got %d", 1, skip)
	}
	v = numerator.GetMetric("speed").GetValuesFloat64()[0]
	if v != -2 {
		t.Errorf("expected = %v, got %v", -2, v)
	}
	pass = numerator.GetMetric("speed").GetPass()[0]
	if pass {
		t.Errorf("expected = %t, got %t", !pass, pass)
	}

	// When numerator > 0 denominator < 0 and denominator < threshold
	//      then skip and pass is false
	numerator, denominator = setupMatrix(20, -10)
	skip, _ = numerator.GetMetric("speed").DivideWithThreshold(denominator.GetMetric("speed"), 5, logging.Get())
	if skip != 1 {
		t.Errorf("expected = %d, got %d", 1, skip)
	}
	v = numerator.GetMetric("speed").GetValuesFloat64()[0]
	if v != 0 {
		t.Errorf("expected = %v, got %v", 0, v)
	}
	pass = numerator.GetMetric("speed").GetPass()[0]
	if pass {
		t.Errorf("expected = %t, got %t", !pass, pass)
	}

	// When numerator < 0 denominator < 0 and denominator < threshold
	//      then skip and pass is false
	numerator, denominator = setupMatrix(-20, -10)
	skip, _ = numerator.GetMetric("speed").DivideWithThreshold(denominator.GetMetric("speed"), 5, logging.Get())
	if skip != 1 {
		t.Errorf("expected = %d, got %d", 1, skip)
	}
	v = numerator.GetMetric("speed").GetValuesFloat64()[0]
	if v != 0 {
		t.Errorf("expected = %v, got %v", 0, v)
	}
	pass = numerator.GetMetric("speed").GetPass()[0]
	if pass {
		t.Errorf("expected = %t, got %t", !pass, pass)
	}
}

func TestMetricFloat64_MultiplyByScalar(t *testing.T) {
	// current > 0
	//      then don't skip and pass is true
	_, current := setupMatrix(10, 10)
	skip, _ := current.GetMetric("speed").MultiplyByScalar(5, logging.Get())
	if skip != 0 {
		t.Errorf("expected = %d, got %d", 0, skip)
	}
	v := current.GetMetric("speed").GetValuesFloat64()[0]
	if v != 50 {
		t.Errorf("expected = %v, got %v", 50, v)
	}
	pass := current.GetMetric("speed").GetPass()[0]
	if !pass {
		t.Errorf("expected = %t, got %t", pass, !pass)
	}

	// When current < 0
	//      then skip and pass is false
	_, current = setupMatrix(10, -10)
	skip, _ = current.GetMetric("speed").MultiplyByScalar(5, logging.Get())
	if skip != 1 {
		t.Errorf("expected = %d, got %d", 1, skip)
	}
	v = current.GetMetric("speed").GetValuesFloat64()[0]
	if v != -50 {
		t.Errorf("expected = %v, got %v", -50, v)
	}
	pass = current.GetMetric("speed").GetPass()[0]
	if pass {
		t.Errorf("expected = %t, got %t", !pass, pass)
	}
}
