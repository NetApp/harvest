//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
// The HTTP daemon exposes metrics for the Prometheus database
// as well as a list of the names of available metrics for humans
//
// Examples:
//
package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/set"
)

func (me *Prometheus) startHttpD(addr, port string) {

	mux := http.NewServeMux()
	mux.HandleFunc("/", me.ServeInfo)
	mux.HandleFunc("/metrics", me.ServeMetrics)

	logger.Debug(me.Prefix+" (httpd)", "starting server at [%s:%s]", addr, port)
	server := &http.Server{Addr: addr + ":" + port, Handler: mux}

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal(me.Prefix+" (httpd)", err.Error())
	} else {
		logger.Info(me.Prefix+" (httpd)", "listening at [http://%s:%s]", addr, port)
	}
}

// checks if address is allowed access
// current implementation only checks for addresses, discarding ports
func (me *Prometheus) checkAddr(addr string) bool {
	if !me.checkAddrs {
		return true
	}

	if value, ok := me.cacheAddrs[addr]; ok {
		return value
	}

	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.Split(addr, ":")[0]

	if me.allowAddrs != nil {
		for _, a := range me.allowAddrs {
			if a == addr {
				me.cacheAddrs[addr] = true
				return true
			}
		}
	}

	if me.allowAddrsRegex != nil {
		for _, r := range me.allowAddrsRegex {
			if r.MatchString(addr) {
				me.cacheAddrs[addr] = true
				return true
			}
		}
	}

	me.cacheAddrs[addr] = false
	return false
}

// send a deny request response
func (me *Prometheus) denyAccess(w http.ResponseWriter, r *http.Request) {

	logger.Debug(me.Prefix+" (httpd) ", "denied request [%s] (%s)", r.RequestURI, r.RemoteAddr)
	w.WriteHeader(403)
	w.Header().Set("content-type", "text/plain")
	w.Write([]byte("403 Forbidden"))
}

func (me *Prometheus) ServeMetrics(w http.ResponseWriter, r *http.Request) {

	var (
		data  [][]byte
		count int
	)

	start := time.Now()

	if !me.checkAddr(r.RemoteAddr) {
		me.denyAccess(w, r)
		return
	}

	logger.Debug(me.Prefix+" (httpd) ", "serving request [%s] (%s)", r.RequestURI, r.RemoteAddr)

	me.cache.Lock()
	for _, metrics := range me.cache.Get() {
		data = append(data, metrics...)
		count += len(metrics)
	}
	me.cache.Unlock()

	// serve our own metadata
	// notice that some values are always taken from previous session
	if md, err := me.render(me.Metadata); err == nil {
		data = append(data, md...)
		count += len(md)
	} else {
		logger.Error(me.Prefix+" (httpd) ", "render metadata: ", err)
	}
	/*

		e.Metadata.SetValueSS("count", "render", float64(count))

		if md, err := e.Render(e.Metadata); err == nil {
			data = append(data, md...)
		}*/

	w.WriteHeader(200)
	w.Header().Set("content-type", "text/plain")
	w.Write(bytes.Join(data, []byte("\n")))

	me.Metadata.Reset()
	me.Metadata.LazySetValueInt64("time", "http", time.Since(start).Microseconds())
	me.Metadata.LazySetValueInt("count", "http", count)
}

// provide a human-friendly overview of metric types and source collectors
// this is than in a very inefficient way, by "reverse engineering" the metrics,
// but probably that's ok, since we don't expected this to be requested
// very often.
// @TODO: also add plugins and plugin metrics
func (me *Prometheus) ServeInfo(w http.ResponseWriter, r *http.Request) {

	start := time.Now()

	if !me.checkAddr(r.RemoteAddr) {
		me.denyAccess(w, r)
		return
	}

	logger.Debug(me.Prefix+" (httpd)", "serving info request [%s] (%s)", r.RequestURI, r.RemoteAddr)

	body := make([]string, 0)

	num_collectors := 0
	num_objects := 0
	num_metrics := 0

	unique_data := map[string]map[string][]string{}

	// copy cache so we don't lock it
	me.cache.Lock()
	cache := make(map[string][][]byte)
	for key, data := range me.cache.Get() {
		cache[key] = make([][]byte, len(data))
		copy(cache[key], data)
	}
	me.cache.Unlock()

	logger.Debug(me.Prefix+" (httpd)", "fetching %d cached elements", len(cache))

	for key, data := range cache {
		logger.Debug(me.Prefix+" (httpd)", "key => [%s] (%d)", key, len(data))
		var collector, object string

		if keys := strings.Split(key, "."); len(keys) == 2 {
			collector = keys[0]
			object = keys[1]
			logger.Debug(me.Prefix+" (httpd)", "collector [%s] - object [%s]", collector, object)
		} else {
			continue
		}

		// skip metadata
		if strings.HasPrefix(object, "metadata_") {
			continue
		}

		metric_names := set.New()
		for _, m := range data {
			if x := strings.Split(string(m), "{"); len(x) >= 2 && x[0] != "" {
				metric_names.Add(x[0])
			}
		}
		num_metrics += metric_names.Size()

		if _, exists := unique_data[collector]; !exists {
			unique_data[collector] = make(map[string][]string)
		}
		unique_data[collector][object] = metric_names.Values()

	}

	for col, per_object := range unique_data {
		objects := make([]string, 0)
		for obj, metric_names := range per_object {
			metrics := make([]string, 0)
			for _, m := range metric_names {
				if m != "" {
					metrics = append(metrics, fmt.Sprintf(metric_template, m))
				}
			}
			objects = append(objects, fmt.Sprintf(object_template, obj, strings.Join(metrics, "\n")))
			num_objects += 1
		}

		body = append(body, fmt.Sprintf(collector_template, col, strings.Join(objects, "\n")))
		num_collectors += 1
	}

	poller := me.Options.Poller
	body_flat := fmt.Sprintf(html_template, poller, poller, poller, num_collectors, num_objects, num_metrics, strings.Join(body, "\n\n"))

	w.WriteHeader(200)
	w.Header().Set("content-type", "text/html")
	w.Write([]byte(body_flat))

	me.Metadata.LazyAddValueInt64("time", "info", time.Since(start).Microseconds())
}

