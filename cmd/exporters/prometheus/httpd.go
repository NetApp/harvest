/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

// Package prometheus creates an HTTP end-point for Prometheus to scrape on `/metrics`
// It also publishes a list of available metrics for human consumption on `/`
package prometheus

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/exporters"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

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

	_, err := p.aCache.streamMetrics(w, tagsSeen, nil)
	if err != nil {
		p.Logger.Error("failed to stream metrics", slogx.Err(err))
	}

	// serve our own metadata
	// notice that some values are always taken from previous session
	md, _, _ := exporters.Render(p.Metadata, p.addMetaTags, p.Params.SortLabels, p.globalPrefix, p.Logger, "")
	_, err = p.aCache.streamMetrics(w, tagsSeen, md)
	if err != nil {
		p.Logger.Error("failed to stream metadata metrics", slogx.Err(err))
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

// ServeInfo provides a human-friendly overview of metric types and source collectors
// this is done in a very inefficient way, by "reverse engineering" the metrics.
// That's probably ok, since we don't expect this to be called often.
func (p *Prometheus) ServeInfo(w http.ResponseWriter, r *http.Request) {

	var (
		numCollectors int
		numObjects    int
		numMetrics    int
		uniqueData    map[string]map[string][]string
	)

	start := time.Now()

	if !p.checkAddr(r.RemoteAddr) {
		p.denyAccess(w, r)
		return
	}

	p.Logger.Debug("serving info request", slog.String("url", r.RequestURI), slog.String("remote_addr", r.RemoteAddr))

	body := make([]string, 0)

	overview, err := p.aCache.getOverview()
	if err != nil {
		p.Logger.Error("failed to get cache statistics", slogx.Err(err))
		http.Error(w, "Failed to collect cache statistics", http.StatusInternalServerError)
		return
	}
	numCollectors = overview.NumCollectors
	numObjects = overview.NumObjects
	numMetrics = overview.NumMetrics
	uniqueData = overview.UniqueData

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
	_, err = w.Write([]byte(bodyFlat))
	if err != nil {
		p.Logger.Error("write info", slogx.Err(err))
	}

	err = p.Metadata.LazyAddValueInt64("time", "info", time.Since(start).Microseconds())
	if err != nil {
		p.Logger.Error("metadata time", slogx.Err(err))
	}
}
