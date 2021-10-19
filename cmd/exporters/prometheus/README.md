

# Prometheus Exporter

## Overview

The Prometheus exporter is responsible for:
- formatting metrics into the Prometheus [line protocol](https://prometheus.io/docs/instrumenting/exposition_formats/) 
- creating a web-endpoint on `http://<ADDR>:<PORT>/metrics` for Prometheus to scrape 
  
A web end-point is required because Prometheus scrapes Harvest by polling that end-point.

In addition to the `/metrics` end-point, the Prometheus exporter also serves an overview of all metrics and collectors available on its root address `http://<ADDR>:<PORT>/`.

Because Prometheus polls Harvest, don't forget to [update your Prometheus configuration](#configure-prometheus-to-scrape-harvest-pollers) and tell Prometheus how to scrape each poller.

There are two ways to configure the Prometheus exporter: using a `port range` or individual `port`s. 

The `port range` is more flexible and should be used when you want multiple pollers all exporting to the same instance of Prometheus. Both options are explained below.

## Parameters

All parameters of the exporter are defined in the `Exporters` section of `harvest.yml`. 

An overview of all parameters:

| parameter           | type                                           | description                                                                                                                                                                                                     | default   |
| ------------------- | ---------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| `port_range`        | int-int (range), overrides `port` if specified | lower port to upper port (inclusive) of the HTTP end-point to create when a poller specifies this exporter. Starting at lower port, each free port will be tried sequentially up to the upper port.             |           |
| `port`              | int, required if port_range is not specified   | port of the HTTP end-point                                                                                                                                                                                      |           |
| `local_http_addr`   | string, optional                               | address of the HTTP server Harvest starts for Prometheus to scrape:<br />use `localhost` to serve only on the local machine<br />use `0.0.0.0` (default) if Prometheus is scrapping from another machine        | `0.0.0.0` |
| `global_prefix`     | string, optional                               | add a prefix to all metrics (e.g. `netapp_`)                                                                                                                                                                    |           |
| `allow_addrs`       | list of strings, optional                      | allow access only if host matches any of the provided addresses                                                                                                                                                 |           |
| `allow_addrs_regex` | list of strings, optional                      | allow access only if host address matches at least one of the regular expressions                                                                                                                               |           |
| `cache_max_keep`    | string (Go duration format), optional          | maximum amount of time metrics are cached (in case Prometheus does not timely collect the metrics)                                                                                                              | `180s`    |
| `add_meta_tags`     | bool, optional                                 | add `HELP` and `TYPE` [metatags](https://prometheus.io/docs/instrumenting/exposition_formats/#comments-help-text-and-type-information) to metrics (currently no useful information, but required by some tools) | `false`   |

A few examples:

#### port_range

```yaml
Exporters:
  prom-prod:
    exporter: Prometheus
    port_range: 2000-2030
Pollers:
  cluster-01:
    exporters:
    - prom-prod
  cluster-02:
    exporters:
    - prom-prod
  cluster-03:
    exporters:
    - prom-prod
  ...
  cluster-16:
    exporters:
    - prom-prod
```

Sixteen pollers will collect metrics from 16 clusters and make those metrics available to a single instance of Prometheus named `prom-prod`. Sixteen web end-points will be created on the first 16 available free ports between 2000 and 2030 (inclusive). 

After staring the pollers in the example above, running `bin/harvest status` shows the following. Note that ports 2000 and 2003 were not available so the next free port in the range was selected. If no free port can be found an error will be logged.

```
Datacenter   Poller       PID     PromPort  Status              
++++++++++++ ++++++++++++ +++++++ +++++++++ ++++++++++++++++++++
DC-01        cluster-01   2339    2001      running         
DC-01        cluster-02   2343    2002      running         
DC-01        cluster-03   2351    2004      running         
...
DC-01        cluster-14   2405    2015      running         
DC-01        cluster-15   2502    2016      running         
DC-01        cluster-16   2514    2017      running         
```

#### allow_addrs

```yaml
Exporters:
  my_prom:
    allow_addrs:
      - 192.168.0.102
      - 192.168.0.103
```
will only allow access from exactly these two addresses.

#### allow_addrs_regex

```yaml
Exporters:
  my_prom:
    allow_addrs_regex:
      - `^192.168.0.\d+$`
```
will only allow access from the IP4 range `192.168.0.0`-`192.168.0.255`.

## Configure Prometheus to scrape Harvest pollers

There are two ways to tell Prometheus how to scrape Harvest: using HTTP service discovery (SD) or listing each poller individually.

HTTP service discovery is the more flexible of the two. It is also less error-prone, and easier to manage. Combined with the port_range configuration described above, SD is the least effort to configure Prometheus and the easiest way to keep both Harvest and Prometheus in sync. 

**NOTE** HTTP service discovery does not work with Docker yet. With Docker, you will need to list each poller individually or if possible, use the [Docker Compose](https://github.com/NetApp/harvest/tree/main/docker) workflow that uses file service discovery to achieve a similar ease-of-use as HTTP service discovery.

See the [example](#prometheus-http-service-discovery-and-port-range) below for how to use HTTP SD and port_range together. 

### Prometheus HTTP Service Discovery

[HTTP service discovery](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config) was introduced in Prometheus version `2.28.0`. Make sure you're using that version or later.

To use HTTP service discovery you need to:
1. tell [Harvest to start the HTTP service discovery process](#enable-http-service-discovery-in-harvest)
2. tell [Prometheus to use the HTTP service discovery endpoint](#enable-http-service-discovery-in-prometheus)

#### Enable HTTP service discovery in Harvest

Add the following to your `harvest.yml`

```
Admin:
  httpsd:
    listen: :8887
```

This tells Harvest to create an HTTP service discovery end-point on interface `0.0.0.0:8887`. If you want to only listen on localhost, use `127.0.0.1:<port>` instead. See [net.Dial](https://pkg.go.dev/net#Dial) for details on the supported listen formats.

Start the SD process by running `bin/harvest admin start`. Once it is started, you can curl the end-point for the list of running Harvest pollers.

```
curl -s 'http://localhost:8887/api/v1/sd' | jq .
[
  {
    "targets": [
      "10.0.1.55:12990",
      "10.0.1.55:15037",
      "127.0.0.1:15511",
      "127.0.0.1:15008",
      "127.0.0.1:15191",
      "10.0.1.55:15343"
    ]
  }
]
```

#### Harvest HTTP Service Discovery options

HTTP service discovery (SD) is configured in the `Admin > httpsd` section of your `harvest.yml`.

| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `listen`   | **required** |  Interface and port to listen on, use localhost:PORT or :PORT for all interfaces                 |                        |
| `auth_basic`           | optional |  If present, enables basic authentication on `/api/v1/sd` end-point                               |                        |
| auth_basic `username`, `password`     | **required** child of `auth_basic` |                             |                        |
| `tls`                 | optional |  If present, enables TLS transport                     |                        |
| tls `cert_file`, `key_file`  | **required** child of `tls` | Relative or absolute path to TLS certificate and key file. TLS 1.3 certificates required.<br />FIPS complaint P-256 TLS 1.3 certificates can be created with `bin/harvest admin tls create server` |   |
| `ssl_cert`, `ssl_key`  | optional if `auth_style` is `certificate_auth` | Absolute paths to SSL (client) certificate and key used to authenticate with the target system.<br /><br />If not provided, the poller will look for `<hostname>.key` and `<hostname>.pem` in `$HARVEST_HOME/cert/`.<br/><br/>To create certificates for ONTAP systems, see [using certificate authentication](docs/AuthAndPermissions.md#using-certificate-authentication)                        |              |
| `heart_beat` | optional, [Go Duration format](https://pkg.go.dev/time#ParseDuration) | How frequently each poller sends a heartbeat message to the SD node | 45s |
| `expire_after` | optional, [Go Duration format](https://pkg.go.dev/time#ParseDuration) | If a poller fails to send a heartbeat, the SD node removes the poller after this duration  | 1m |

#### Enable HTTP service discovery in Prometheus

Edit your `prometheus.yml` and add the following section

`$ vim /etc/prometheus/prometheus.yml`

```
scrape_configs:
  - job_name: harvest
    http_sd_configs:
    - url: http://localhost:8887/api/v1/sd
```
 
Harvest and Prometheus both support basic authentication for HTTP SD end-points. To enable basic auth, add the following to your Harvest config.

```
Admin:
  httpsd:
    listen: :8887
    # Basic auth protects GETs and publishes
    auth_basic: 
      username: admin
      password: admin
```

Don't forget to also update your Prometheus config with the matching [basic_auth](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config) credentials.

### Prometheus HTTP Service Discovery and Port Range

HTTP SD combined with Harvest's `port_range` feature leads to significantly less configuration in your `harvest.yml`. For example, if your clusters all export to the same Prometheus instance, you can refactor the per-poller exporter into a single exporter shared by all clusters in `Defaults` as shown below:

Notice that none of the pollers specify an exporter. Instead, all the pollers share the single exporter named `prometheus-r` listed in `Defaults`. `prometheus-r` is the only exporter defined and as specified will manage up to 1,000 Harvest Prometheus exporters.

If you add or remove more clusters in the `Pollers` section, you do not have to change Prometheus since it dynamically pulls the targets from the Harvest admin node.

```
Admin:
  httpsd:
    listen: :8887

Exporters:
  prometheus-r:
    exporter: Prometheus
    port_range: 13000-13999

Defaults:
  collectors:
    - Zapi
    - ZapiPerf
  use_insecure_tls: false
  auth_style: password
  username: admin
  password: pass
  exporters:
    - prometheus-r

Pollers:
  umeng_aff300:
    datacenter: meg
    addr: 10.193.48.11
    
  F2240-127-26:
    datacenter: meg
    addr: 10.193.6.61

  # ... add more clusters
```
### Static Scrape Targets

If we define four prometheus exporters at ports: 12990, 12991, 14567, and 14568 you need to add four sections to your `prometheus.yml`.

```bash
$ vim /etc/prometheus/prometheus.yml
```

Scroll down to near the end of file and add the following lines:

```yaml
  - job_name: 'harvest'
    scrape_interval:     60s 
    static_configs:
      - targets: 
        - 'localhost:12990'
        - 'localhost:12991'
        - 'localhost:14567'
        - 'localhost:14568'
```
**NOTE** If Prometheus is not on the same machine as Harvest, then replace `localhost` with the IP address of your Harvest machine. Also note the scrape interval above is set to 60s. That matches the polling frequency of the default Harvest collectors. If you change the polling frequency of a Harvest collector to a lower value, you should also change the scrape interval.
