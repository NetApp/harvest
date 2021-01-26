package prometheus

import (
	"fmt"
	"net/http"
	"bytes"
)


func (p *Prometheus) StartHttpd(url, port string) {

	fmt.Printf("I am here!!!!\n")
	
	mux := http.NewServeMux()
	//mux.HandleFunc("/", p.ServeInfo)
	mux.HandleFunc("/metrics", p.ServeMetrics)

	PORT := ":"+port
	Log.Info("Starting server at [%s]", PORT)
	server := &http.Server{ Addr: PORT, Handler: mux}
	go server.ListenAndServe()

}

func (p *Prometheus) ServeInfo(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    w.Write([]byte("success!\n"))
}

func (p *Prometheus) ServeMetrics(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Wow!!\n")

	Log.Info("Serving metrics!!")
	sep := []byte("\n")
	var data [][]byte
	for _, m := range p.Cache {
		Log.Info("Rendering metrics [%s:%s]", m.Collector, m.Object)
		rendered := p.Render(m)
		for _, metric := range rendered {
			data = append(data, []byte(metric))
		}
	}

	data = append(data, sep)
	w.WriteHeader(200)
	w.Write(bytes.Join(data, sep))
}