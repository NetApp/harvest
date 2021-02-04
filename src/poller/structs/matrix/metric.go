package matrix

// Metric struct and related methods

type Metric struct {
	Index int
    Display string
	Enabled bool
	Scalar bool
	/* extended fields for ZapiPerf counters */
	Properties string
	BaseCounter string
	/* fields for Array counters */
	Size int
	Dimensions int
	Labels []string
	SubLabels []string
}

func (m *Matrix) GetMetric(key string) *Metric {

    if metric, found := m.Metrics[key]; found {
		return metric
	}
	return nil
}

// Create new metric and add to cache
func (m *Matrix) AddMetric(key, display string, enabled bool) (*Metric, error) {

	if _, exists := m.Metrics[key]; exists {
		err = errors.New(fmt.Sprintf("Metric [%s] already in cache", key))
		return nil, err
	}

	metric = Metric{Index: m.MetricsIndex, Display: display, Scalar: true, Enabled: enabled}
	m.Metrics[key] = &metric
	m.MetricsIndex += 1

	return &metric, nil
}

// Create 1D Array Matric
func (m *Matrix) AddArrayMetric(key, display string, labels []string, enabled bool) (*Metric, error) {
    if metric, err := m.AddMetric(key, display, enabled); err == nil {
		metric.Scalar = false
		metric.Dimensions = 1
		metric.Size = len(labels)
		metric.Labels = labels
		m.MetricsIndex += metric.Size - 1 // already incremented by 1
		return metric, nil
	} else {
		return nil, err
	}
}

// Similar to AddMetric, but metric is initialized. This allows collectors
// to add extended fields to metric or create multidimensional Array metric.
//
// Method should be used with caution: incorrect "size" will corrupt data
// or make Harvest panic
func (m *Matrix) AddCustomMetric(key string, metric *Metric) error {
	if _, exists := m.Metrics[key]; exists {
		return errors.New(fmt.Sprintf("Metric [%s] already in cache", key))
	}
	// sanity check: array should come with size
	if !m.Scalar && m.Size == 0 {
		return errors.New("Array metric with no valid size")
	}
	m.Metrics[key] = metric
	m.MetricsIndex += metric.Size
}

