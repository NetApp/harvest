package prometheus

import (
	"fmt"
	"net/http"
	"bytes"
	"sort"
	"strings"
	"time"
	"poller/structs/matrix"
)

var html_template = `
        <!DOCTYPE html>
        <html>
            <head>
            <meta charset="utf-8">
            <title>NetApp Harvest 2.0 -%s- Prometheus Exporter</title>
            </head>
            <body>
            <br/>
            <h2 style="color:#404040">NetApp Harvest 2.0 - %s</h2>
            <p style="color:#303030">
            Welcome to the Prometheus Exporter of poller <em>%s</em>!<br/>
            If you are Prometheus scraper, get the metric data <a href="/metrics">here</a>.<br/><br/>

            Below is the list of metrics provided by my collectors and plugins.<br/>
            Exposing data from %d collectors and %d objects, %d metrics in total.<br/><br/>
            Note: this is a real-time generated list and might change over time.<br/>
            If you just started Harvest, you might need to wait a few minutes<br/>
            before the full list of counters is available here.<br/><br/>
            </p>
            %s
            </body>
        </html>`

var body_template = `
				<div style="margin-left:40px; color:#303030">
					%s
				</div>`

var collector_template = `
					<h3 style="color:#404040">%s</h3>
					<small><em>collector</em></small>
					<ul>
						%s
					</ul>
					`

var object_template = `
						<h4 style="color:#404040">%s</h4>
						<small><em>object</em></small>
						%s`

var metric_template = `<li>%s</li>`


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