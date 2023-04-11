package harvest

Exporters: [Name=_]: #Prom | #Influx

label: [string]: string

#Admin: {
	addr?: string
}

#Prom: {
	local_http_addr?: "0.0.0.0" | "localhost" | "127.0.0.1"
	addr?:            string // deprecated
	exporter:         "Prometheus"
	port?:            int
	port_range?:      string
	allow_addrs_regex?: [...string]
	add_meta_tags?: bool
	sort_labels?: bool
}

#Influx: {
	addr?:    string // one of addr|url
	url?:     string
	exporter: "InfluxDB"
	bucket?:   string
	org?:      string
	token?:   string
	allow_addrs_regex: [...string]
}

#CredentialsScript: {
    path: string
    schedule?: string
    timeout?: string
}

#CollectorDef: {
	[Name=_]: [...string]
}

Pollers: [Name=_]: #Poller

#Poller: {
	datacenter?:         string
	auth_style?:         "basic_auth" | "certificate_auth"
	ssl_cert?:           string
	ssl_key?:            string
	username?:           string
	password?:           string
	use_insecure_tls?:   bool
	is_kfs?:             bool
	addr?:               string
	log_max_bytes?:      int
	log_max_files?:      int
	client_timeout?:     string
	collectors?:         [...#CollectorDef] | [...string]
	exporters: [...string]
	log: [...string]
	labels?: [...label]
	credentials_script?: #CredentialsScript
	prefer_zapi?:        bool
}
