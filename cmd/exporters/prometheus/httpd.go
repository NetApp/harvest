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
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"
)

func (p *Prometheus) startHTTPD(addr string, port int) {

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.ServeInfo)
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
		for _, a := range p.allowAddrs {
			if a == addr {
				p.cacheAddrs[addr] = true
				return true
			}
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
		data  [][]byte
		count int
	)

	start := time.Now()

	if !p.checkAddr(r.RemoteAddr) {
		p.denyAccess(w, r)
		return
	}

	p.cache.Lock()
	// Count the number of metrics so we can pre-allocate the slice to avoid reallocations
	for _, metrics := range p.cache.Get() {
		count += len(metrics)
	}

	data = make([][]byte, 0, count)
	tagsSeen := make(map[string]bool)

	for _, metrics := range p.cache.Get() {
		data = addMetricsToSlice(data, metrics, tagsSeen, p.addMetaTags)
	}
	p.cache.Unlock()

	// serve our own metadata
	// notice that some values are always taken from previous session
	md, _ := p.render(p.Metadata)
	data = addMetricsToSlice(data, md, tagsSeen, p.addMetaTags)

	count += len(md)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write(bytes.Join(data, []byte("\n")))
	if err != nil {
		p.Logger.Error("write metrics", slogx.Err(err))
	} else {
		// make sure stream ends with newline
		if _, err2 := w.Write([]byte("\n")); err2 != nil {
			p.Logger.Error("write ending newline", slog.Any("err", err2))
		}
	}

	// update metadata
	p.Metadata.Reset()
	err = p.Metadata.LazySetValueInt64("time", "http", time.Since(start).Microseconds())
	if err != nil {
		p.Logger.Error("metadata time", slogx.Err(err))
	}
	err = p.Metadata.LazySetValueInt64("count", "http", int64(count))
	if err != nil {
		p.Logger.Error("metadata count", slogx.Err(err))
	}
}

// addMetricsToSlice adds metrics to a slice, skipping duplicates. Normally
// Render() only adds one TYPE/HELP for each metric type. Some metric types
// (e.g., metadata_collector_metrics) are submitted from multiple collectors.
// That causes duplicates that are removed in this function. The seen map is
// used to keep track of which metrics have been added. The metrics slice is
// expected to be in the format: # HELP metric_name help text # TYPE metric_name
// type metric_name{tag="value"} value
func addMetricsToSlice(data [][]byte, metrics [][]byte, seen map[string]bool, addMetaTags bool) [][]byte {

	if !addMetaTags {
		return append(data, metrics...)
	}

	for i, metric := range metrics {
		if bytes.HasPrefix(metric, []byte("# ")) {
			if fields := bytes.Fields(metric); len(fields) > 3 {
				name := string(fields[2])
				if !seen[name] {
					seen[name] = true
					data = append(data, metric)
					if i+1 < len(metrics) {
						data = append(data, metrics[i+1])
					}
				}
			}
		} else {
			data = append(data, metric)
		}
	}

	return data
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

	// copy cache so we don't lock it
	p.cache.Lock()
	cache := make(map[string][][]byte)
	for key, data := range p.cache.Get() {
		cache[key] = make([][]byte, len(data))
		copy(cache[key], data)
	}
	p.cache.Unlock()

	p.Logger.Debug("fetching cached elements", slog.Int("count", len(cache)))

	for key, data := range cache {
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
		}
		uniqueData[collector][object] = metricNames.Values()

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
			numObjects++
		}

		body = append(body, fmt.Sprintf(collectorTemplate, col, strings.Join(objects, "\n")))
		numCollectors++
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
