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
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (p *Prometheus) startHTTPD(addr string, port int) {

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.ServeInfo)
	mux.HandleFunc("/metrics", p.ServeMetrics)

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

	p.Logger.Info().Str("url", url).Msg("server listen")

	if p.Params.TLS.KeyFile != "" {
		if err := server.ListenAndServeTLS(p.Params.TLS.CertFile, p.Params.TLS.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			p.Logger.Fatal().Err(err).
				Str("url", url).
				Str("cert_file", p.Params.TLS.CertFile).
				Str("key_file", p.Params.TLS.KeyFile).
				Msg("Failed to start server")
		}
	} else {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			p.Logger.Fatal().Err(err).
				Str("url", url).
				Msg("Failed to start server")
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

	p.Logger.Debug().Msgf("(httpd) denied request [%s] (%s)", r.RequestURI, r.RemoteAddr)
	w.WriteHeader(http.StatusForbidden)
	w.Header().Set("content-type", "text/plain")
	_, err := w.Write([]byte("403 Forbidden"))
	if err != nil {
		p.Logger.Error().Stack().Err(err).Msg("error")
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
	for _, metrics := range p.cache.Get() {
		data = append(data, metrics...)
		count += len(metrics)
	}
	p.cache.Unlock()

	// serve our own metadata
	// notice that some values are always taken from previous session
	md, _ := p.render(p.Metadata)
	data = append(data, md...)
	count += len(md)
	/*

		e.Metadata.SetValueSS("count", "render", float64(count))

		if md, err := e.Render(e.Metadata); err == nil {
			data = append(data, md...)
		}
	*/

	if p.addMetaTags {
		data = filterMetaTags(data)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "text/plain")
	_, err := w.Write(bytes.Join(data, []byte("\n")))
	if err != nil {
		p.Logger.Error().Err(err).Msg("write metrics")
	} else {
		// make sure stream ends with newline
		if _, err2 := w.Write([]byte("\n")); err2 != nil {
			p.Logger.Error().Err(err2).Msg("write ending newline")
		}
	}

	// update metadata
	p.Metadata.Reset()
	err = p.Metadata.LazySetValueInt64("time", "http", time.Since(start).Microseconds())
	if err != nil {
		p.Logger.Error().Stack().Err(err).Msg("error")
	}
	err = p.Metadata.LazySetValueInt64("count", "http", int64(count))
	if err != nil {
		p.Logger.Error().Stack().Err(err).Msg("error")
	}
}

// filterMetaTags removes duplicate TYPE/HELP tags in the metrics
// Note: this is a workaround, normally Render() will only add
// one TYPE/HELP for each metric type, however since some metric
// types (e.g. metadata_collector_metrics) are submitted from multiple
// collectors, we end up with duplicates in the final batch delivered
// over HTTP.
func filterMetaTags(metrics [][]byte) [][]byte {

	filtered := make([][]byte, 0)

	metricsWithTags := make(map[string]bool)

	for i, m := range metrics {
		if bytes.HasPrefix(m, []byte("# ")) {
			if fields := strings.Fields(string(m)); len(fields) > 3 {
				name := fields[2]
				if !metricsWithTags[name] {
					metricsWithTags[name] = true
					filtered = append(filtered, m)
					if i+1 < len(metrics) {
						filtered = append(filtered, metrics[i+1])
					}
				}
			}
		} else {
			filtered = append(filtered, m)
		}
	}
	return filtered
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

	p.Logger.Debug().Msgf("(httpd) serving info request [%s] (%s)", r.RequestURI, r.RemoteAddr)

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

	p.Logger.Debug().Msgf("(httpd) fetching %d cached elements", len(cache))

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
	w.Header().Set("content-type", "text/html")
	_, err := w.Write([]byte(bodyFlat))
	if err != nil {
		p.Logger.Error().Stack().Err(err).Msg("error")
	}

	err = p.Metadata.LazyAddValueInt64("time", "info", time.Since(start).Microseconds())
	if err != nil {
		p.Logger.Error().Stack().Err(err).Msg("error")
	}
}
