package harvest

Exporters: [Name=_]: #Prom | #Influx | #PromConsul

#Prom: {
	addr: string
	exporter:    "Prometheus"
	port?:       int
	port_range?: string
	allow_addrs_regex?: [...string]
}

#PromConsul: {
	addr: string
	exporter:    "PrometheusConsul"
    service_name: string
    tags: [...string]
}

#Influx: {
	addr?: string  // one of addr|url
	url?: string
	exporter: "InfluxDB"
	bucket:   string
	org:      string
	token?:   string
	allow_addrs_regex: [...string]
}

Pollers: [Name=_]: #Poller

#Poller: {
	datacenter?:       string
	auth_style?:       "basic_auth" | "certificate_auth"
	ssl_cert?:         string
	ssl_key?:          string
	username?:         string
	password?:         string
	use_insecure_tls?: bool
	is_kfs?:           bool

	addr?:          string
	log_max_bytes?: int
	log_max_files?: int
	collectors: [...string]
	exporters: [...string]
}