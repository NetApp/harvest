

# Prometheus Exporter

## Overview

The Prometheus Exporter will format metrics into the Prometheus [line protocol](https://prometheus.io/docs/instrumenting/exposition_formats/) (a.k.a. *open metric format*) and expose it on an HTTP port (`http://<ADDR>:<PORT>/metrics`). Additionally, it serves a basic overview of available metrics and collectors on its root address (`http://<ADDR>:<PORT>`).


## Design

The Exporter has two concurrent components that makes it possible to simulatenously serve push requests by collectors and pull requests by Prometheus scrapers.

<img src="prometheus.png" width="100%" align="center">


## Parameters

All parameters of the exporter are defined in the `Exporters` section of `harvest.yml`. An exception is port, which can either be part of the exporter parameters (`port`) or part of the poller parameters (`prometheus_port`). If you go for the first option, you can define as many Prometheus exporters as you would like. We are planning to integrete the exporter with one of Prometheus' service-discovery options to simplify this in the future.


An overview of all parameters:


| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `port`                 | int, requird | port of the HTTP server                          |                        |
| `local_http_addr`      | string, optional	| address of the HTTP server Harvest starts for Prometheus to scrape:<br />use `localhost` to serve only on the local machine<br />use `0.0.0.0` (default) if Prometheus is scrapping from another machine | `0.0.0.0` |
| `global_prefx`         | string, optional | add a prefix to all metrics (e.g. `netapp_`) |                        |
| `allow_addrs`          | list of strings, optional | allow access only if host matches any of the provided addresses | |
| `allow_addrs_regex`	 | list of strings, optional | allow access only if host address matches at least one of the regular expressions | |
| `cache_max_keep`       | string (Go duration format), optional | maximum amount of time metrics are cached (in case Prometheus does not timely collect the metrics) | `180s` |
|	|	|	|	|


A few examples:

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
