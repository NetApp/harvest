/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

// Package prometheus creates an HTTP end-point for Prometheus to scrape on `/metrics`
// It also publishes a list of available metrics for human consumption on `/`
package prometheus

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
)

func (p *Prometheus) startHTTPD(addr string, port int) {

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.ServeInfo)
	mux.HandleFunc("/health", p.checkHealth)
	mux.HandleFunc("/metrics", p.ServeMetrics)
	mux.HandleFunc("localhost/debug/pprof/", pprof.Index)
	mux.HandleFunc("localhost/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("localhost/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("localhost/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("localhost/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr:              addr + ":" + strconv.Itoa(port),
		Handler:           mux,
		ReadHeaderTimeout: 60 * time.Second,
	}

	var url string
	if p.Params.TLS.KeyFile != "" {
		url = fmt.Sprintf("https://%s/metrics", net.JoinHostPort(addr, strconv.Itoa(port)))
	} else {
		url = fmt.Sprintf("%s://%s/metrics", "http", net.JoinHostPort(addr, strconv.Itoa(port)))
	}

	p.Logger.Info("server listen", slog.String("url", url))

	if p.Params.TLS.KeyFile != "" {
		if err := server.ListenAndServeTLS(p.Params.TLS.CertFile, p.Params.TLS.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			p.Logger.Error(
				"Failed to start server",
				slogx.Err(err),
				slog.String("url", url),
				slog.String("cert_file", p.Params.TLS.CertFile),
				slog.String("key_file", p.Params.TLS.KeyFile),
			)
			os.Exit(1)
		}
	} else {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			p.Logger.Error("Failed to start server", slogx.Err(err), slog.String("url", url))
			os.Exit(1)
		}
	}
}

func (p *Prometheus) checkHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		p.Logger.Error("error", slogx.Err(err))
	}
}

// checks if address is allowed access
// current implementation only checks for addresses, discarding ports
func (p *Prometheus) checkAddr(addr string) bool {
	if !p.checkAddrs {
		return true
	}

	if value, ok := p.cacheAddrs[addr]; ok {
		return value
	}

	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.Split(addr, ":")[0]

	if p.allowAddrs != nil {
		if slices.Contains(p.allowAddrs, addr) {
			p.cacheAddrs[addr] = true
			return true
		}
	}

	if p.allowAddrsRegex != nil {
		for _, r := range p.allowAddrsRegex {
			if r.MatchString(addr) {
				p.cacheAddrs[addr] = true
				return true
			}
		}
	}

	p.cacheAddrs[addr] = false
	return false
}

// send a deny request response
func (p *Prometheus) denyAccess(w http.ResponseWriter, r *http.Request) {

	p.Logger.Debug("denied request", slog.String("url", r.RequestURI), slog.String("remote_addr", r.RemoteAddr))
	w.WriteHeader(http.StatusForbidden)
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write([]byte("403 Forbidden"))
	if err != nil {
		p.Logger.Error("error", slogx.Err(err))
	}
}

func (p *Prometheus) ServeMetrics(w http.ResponseWriter, r *http.Request) {

	var (
		count int
	)

	start := time.Now()

	if !p.checkAddr(r.RemoteAddr) {
		p.denyAccess(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	tagsSeen := make(map[string]struct{})

	if p.useDiskCache {
		if diskcache, ok := p.cache.(*diskCache); ok {
			diskcache.Lock()
			count = diskcache.GetMetricCount()
			err := diskcache.StreamToWriter(w)
			diskcache.Unlock()
			if err != nil {
				p.Logger.Error("failed to stream metrics from disk cache", slogx.Err(err))
			}
		}
	} else {
		if memCache, ok := p.cache.(*cache); ok {
			memCache.Lock()
			for _, metrics := range memCache.Get() {
				count += p.writeMetrics(w, metrics, tagsSeen)
			}
			memCache.Unlock()
		}
	}

	// serve our own metadata
	// notice that some values are always taken from previous session
	md, _ := p.render(p.Metadata)
	count += p.writeMetrics(w, md, tagsSeen)

	// update metadata
	p.Metadata.Reset()
	err := p.Metadata.LazySetValueInt64("time", "http", time.Since(start).Microseconds())
	if err != nil {
		p.Logger.Error("metadata time", slogx.Err(err))
	}
	err = p.Metadata.LazySetValueInt64("count", "http", int64(count))
	if err != nil {
		p.Logger.Error("metadata count", slogx.Err(err))
	}
}

// writeMetrics writes metrics to the writer, skipping duplicates.
// Normally Render() only adds one TYPE/HELP for each metric type.
// Some metric types (e.g., metadata_collector_metrics) are submitted from multiple collectors.
// That causes duplicates that are suppressed in this function.
// The seen map is used to keep track of which metrics have been added.
func (p *Prometheus) writeMetrics(w io.Writer, metrics [][]byte, tagsSeen map[string]struct{}) int {

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
							p.writeMetric(w, metric)
							count++
							if i+1 < len(metrics) {
								p.writeMetric(w, metrics[i+1])
								count++
								i++
							}
						}
						break
					}
				}
			}
		} else {
			p.writeMetric(w, metric)
			count++
		}
	}

	return count
}

func (p *Prometheus) writeMetric(w io.Writer, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		p.Logger.Error("write metrics", slogx.Err(err))
		return
	}
	_, err = w.Write([]byte("\n"))
	if err != nil {
		p.Logger.Error("write newline", slogx.Err(err))
		return
	}
}

// ServeInfo provides a human-friendly overview of metric types and source collectors
// this is done in a very inefficient way, by "reverse engineering" the metrics.
// That's probably ok, since we don't expect this to be called often.
func (p *Prometheus) ServeInfo(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if !p.checkAddr(r.RemoteAddr) {
		p.denyAccess(w, r)
		return
	}

	p.Logger.Debug("serving info request", slog.String("url", r.RequestURI), slog.String("remote_addr", r.RemoteAddr))

	body := make([]string, 0)

	numCollectors := 0
	numObjects := 0
	numMetrics := 0

	uniqueData := map[string]map[string][]string{}

	switch c := p.cache.(type) {
	case *cache:
		c.Lock()
		cacheData := make(map[string][][]byte)
		for key, data := range c.Get() {
			cacheData[key] = make([][]byte, len(data))
			copy(cacheData[key], data)
		}
		c.Unlock()

		p.Logger.Debug("fetching cached elements", slog.Int("count", len(cacheData)))

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
			numMetrics += metricNames.Size()

			if _, exists := uniqueData[collector]; !exists {
				uniqueData[collector] = make(map[string][]string)
				numCollectors++
			}
			if _, exists := uniqueData[collector][object]; !exists {
				numObjects++
			}
			uniqueData[collector][object] = metricNames.Values()
		}
	case *diskCache:
		c.Lock()
		stats, err := c.GetStats()
		c.Unlock()
		if err != nil {
			p.Logger.Error("failed to get cache statistics", slogx.Err(err))
			http.Error(w, "Failed to collect cache statistics", http.StatusInternalServerError)
			return
		}

		numCollectors = stats.NumCollectors
		numObjects = stats.NumObjects
		numMetrics = stats.NumMetrics
		uniqueData = stats.UniqueData

	default:
		p.Logger.Error("unexpected cache type")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for col, perObject := range uniqueData {
		objects := make([]string, 0)
		for obj, metricNames := range perObject {
			metrics := make([]string, 0)
			for _, m := range metricNames {
				if m != "" {
					metrics = append(metrics, fmt.Sprintf(metricTemplate, m))
				}
			}
			objects = append(objects, fmt.Sprintf(objectTemplate, obj, strings.Join(metrics, "\n")))
		}

		body = append(body, fmt.Sprintf(collectorTemplate, col, strings.Join(objects, "\n")))
	}

	poller := p.Options.Poller
	bodyFlat := fmt.Sprintf(htmlTemplate, poller, poller, poller, numCollectors, numObjects, numMetrics, strings.Join(body, "\n\n"))

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write([]byte(bodyFlat))
	if err != nil {
		p.Logger.Error("write info", slogx.Err(err))
	}

	err = p.Metadata.LazyAddValueInt64("time", "info", time.Since(start).Microseconds())
	if err != nil {
		p.Logger.Error("metadata time", slogx.Err(err))
	}
}
