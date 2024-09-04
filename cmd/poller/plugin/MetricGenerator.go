package plugin

type CustomMetric struct {
	Name         string
	Endpoint     string
	ONTAPCounter string
	Description  string
	Prefix       string
}

type MetricGenerator interface {
	GetGeneratedMetrics() []CustomMetric
}
