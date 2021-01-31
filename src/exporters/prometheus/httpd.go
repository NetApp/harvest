package main

import (
	"fmt"
	"net/http"
	"bytes"
	"sort"
	"strings"
	"time"
	"goharvest2/poller/structs/matrix"
)


func (p *Prometheus) StartHttpd(url, port string) {

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.ServeInfo)
	mux.HandleFunc("/metrics", p.ServeMetrics)

	PORT := ":"+port
	Log.Info("Starting server at [%s]", PORT)
	server := &http.Server{ Addr: PORT, Handler: mux}
	go server.ListenAndServe()

}

func (p *Prometheus) ServeInfo(w http.ResponseWriter, r *http.Request) {
	
	body := make([]string, 0)
	//matrix_by_collector := make(map[string][]*matrix.Matrix)
	num_collectors := 0
	num_objects := 0
	num_metrics := 0
	filtered_cache := map[string]map[string]*matrix.Matrix{}
	//collector_names = map[string]string
	//object_names map[string][]string

	for _, m := range p.cache {

		if _, ok := filtered_cache[m.Collector]; !ok {
			filtered_cache[m.Collector] = make(map[string]*matrix.Matrix)
		}
		filtered_cache[m.Collector][m.Object] = m
		//key := m.Collector + "." + m.Object + "." + m.Plugin
		//matrix_by_collector[m.Collector] = append(matrix_by_collector[m.Collector], m)
	}

	for c, data_per_object := range filtered_cache {

		objects := make([]string, 0)

		for _, m := range data_per_object {

			metrics := make([]string, 0)

			for _, metric := range m.Metrics {
				
				if !metric.Enabled {
					continue
				}

				num_metrics += 1

				if metric.Scalar {
					metrics = append(metrics, fmt.Sprintf(metric_template, m.Object + "_" + metric.Display))
				} else {
					array_metric := fmt.Sprintf(metric_template, m.Object + "_" + metric.Display)
					array_metric += "\n<ul>"
					for _, label := range metric.Labels {
						array_metric += "\n" + fmt.Sprintf(metric_template, label)
					}
					array_metric += "\n</ul>"
					metrics = append(metrics, array_metric)
				}
			}

			sort.Strings(metrics)

			objects = append(objects, fmt.Sprintf(object_template, m.Object, strings.Join(metrics, "\n")))

			//num_metrics += len(metrics)
			num_objects += 1
		}

		body = append(body, fmt.Sprintf(collector_template, c, strings.Join(objects, "\n")))
		num_collectors += 1
	}

	poller := p.options.Poller
	body_flat := fmt.Sprintf(html_template, poller, poller, poller, num_collectors, num_objects, num_metrics, strings.Join(body, "\n\n"))
	
	w.WriteHeader(200)
	w.Header().Set("content-type", "text/html")
	w.Write([]byte(body_flat))
}

func (p *Prometheus) ServeMetrics(w http.ResponseWriter, r *http.Request) {

	Log.Info("Serving metrics from %d cached items", len(p.cache))
	sep := []byte("\n")
	var data [][]byte

	start := time.Now()
	count := 0

	for _, m := range p.cache {
		Log.Info("Rendering metrics [%s:%s]", m.Collector, m.Object)
		rendered := p.Render(m)

		data = append(data, rendered...)
		count += len(rendered)
	}

	duration := time.Since(start)
	p.Metadata.SetValueForMetricAndInstance("time", "render", duration.Seconds())
	p.Metadata.SetValueForMetricAndInstance("count", "render", float64(count))

	md := p.Render(p.Metadata)
	data = append(data, md...)
	//data = append(data, sep)

	w.WriteHeader(200)
	w.Header().Set("content-type", "text/plain")
	w.Write(bytes.Join(data, sep))
}