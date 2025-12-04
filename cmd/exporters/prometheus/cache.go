/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package prometheus

import (
	"bytes"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

type cacher interface {
	getOverview() (*CacheStats, error)
	exportMetrics(key string, data [][]byte, names *set.Set)
	streamMetrics(w http.ResponseWriter, seen map[string]struct{}, metrics [][]byte) (int, error)
	isValid() bool
}

type memCache struct {
	mu     *sync.Mutex
	logger *slog.Logger
	data   map[string][][]byte
	timers map[string]time.Time
	expire time.Duration
}

func (c *memCache) isValid() bool {
	return true
}

func (c *memCache) getOverview() (*CacheStats, error) {
	c.mu.Lock()
	cacheData := make(map[string][][]byte)
	for key, data := range c.Get() {
		cacheData[key] = make([][]byte, len(data))
		copy(cacheData[key], data)
	}
	c.mu.Unlock()

	stats := &CacheStats{
		UniqueData: make(map[string]map[string][]string),
	}

	for key, data := range cacheData {
		var collector, object string

		if keys := strings.Split(key, "."); len(keys) == 3 {
			collector = keys[0]
			object = keys[1]
		} else {
			continue
		}

		// skip metadata
		if strings.HasPrefix(object, "metadata_") {
			continue
		}

		metricNames := set.New()
		for _, m := range data {
			if x := strings.Split(string(m), "{"); len(x) >= 2 && x[0] != "" {
				metricNames.Add(x[0])
			}
		}
		stats.NumMetrics += metricNames.Size()

		if _, exists := stats.UniqueData[collector]; !exists {
			stats.UniqueData[collector] = make(map[string][]string)
			stats.NumCollectors++
		}
		if _, exists := stats.UniqueData[collector][object]; !exists {
			stats.NumObjects++
		}
		stats.UniqueData[collector][object] = metricNames.Values()
	}

	return stats, nil
}

func (c *memCache) exportMetrics(key string, data [][]byte, metricNames *set.Set) {
	c.Put(key, data, metricNames)
}

func (c *memCache) streamMetrics(w http.ResponseWriter, tagsSeen map[string]struct{}, metrics [][]byte) (int, error) {
	c.mu.Lock()
	var count int
	if metrics == nil {
		// stream all cached metrics
		for _, metrics := range c.Get() {
			count += c.writeMetrics(w, metrics, tagsSeen)
		}
	} else {
		// stream only provided metrics
		count += c.writeMetrics(w, metrics, tagsSeen)
	}

	c.mu.Unlock()

	return count, nil
}

// writeMetrics writes metrics to the writer, skipping duplicates.
// Normally Render() only adds one TYPE/HELP for each metric type.
// Some metric types (e.g., metadata_collector_metrics) are submitted from multiple collectors.
// That causes duplicates that are suppressed in this function.
// The seen map is used to keep track of which metrics have been added.
func (c *memCache) writeMetrics(w io.Writer, metrics [][]byte, tagsSeen map[string]struct{}) int {

	var count int

	for i := 0; i < len(metrics); i++ {
		metric := metrics[i]
		if bytes.HasPrefix(metric, []byte("# ")) {

			// Find the metric name and check if it has been seen before
			var (
				spacesSeen  int
				space2Index int
			)

			for j := range metric {
				if metric[j] == ' ' {
					spacesSeen++
					if spacesSeen == 2 {
						space2Index = j
					} else if spacesSeen == 3 {
						name := string(metric[space2Index+1 : j])
						if _, ok := tagsSeen[name]; !ok {
							tagsSeen[name] = struct{}{}
							c.writeMetric(w, metric)
							count++
							if i+1 < len(metrics) {
								c.writeMetric(w, metrics[i+1])
								count++
								i++
							}
						}
						break
					}
				}
			}
		} else {
			c.writeMetric(w, metric)
			count++
		}
	}

	return count
}

func (c *memCache) writeMetric(w io.Writer, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		c.logger.Error("write metrics", slogx.Err(err))
	}
	_, err = w.Write([]byte("\n"))
	if err != nil {
		c.logger.Error("write newline", slogx.Err(err))
	}
}

func newMemCache(l *slog.Logger, d time.Duration) *memCache {
	c := memCache{mu: &sync.Mutex{}, expire: d, logger: l}
	c.data = make(map[string][][]byte)
	c.timers = make(map[string]time.Time)
	return &c
}

func (c *memCache) Get() map[string][][]byte {
	c.Clean()
	return c.data
}

func (c *memCache) Put(key string, data [][]byte, _ *set.Set) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = data
	c.timers[key] = time.Now()
}

func (c *memCache) Clean() {
	for k, t := range c.timers {
		if time.Since(t) > c.expire {
			delete(c.timers, k)
			delete(c.data, k)
		}
	}
}
