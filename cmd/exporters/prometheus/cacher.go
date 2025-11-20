package prometheus

import (
	"github.com/netapp/harvest/v2/pkg/set"
	"io"
)

type CacheStats struct {
	NumCollectors int
	NumObjects    int
	NumMetrics    int
	UniqueData    map[string]map[string][]string
}

type cacher interface {
	Put(key string, data [][]byte, metricNames *set.Set)
	Clean()
}

type memoryCacher interface {
	cacher
	Get() map[string][][]byte
}

type diskCacher interface {
	cacher
	StreamToWriter(w io.Writer) error
	GetMetricCount() int
	GetStats() (*CacheStats, error)
}
