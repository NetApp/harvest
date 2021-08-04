package harvest

Exporters: [Name=_]: #Prom | #Influx

#Prom: {
	local_http_addr: "0.0.0.0" | "localhost" | "127.0.0.1"
	addr: string // deprecated
	exporter:    "Prometheus"
	port?:       int
	port_range?: string
	allow_addrs_regex?: [...string]
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
	log: [...string]
}