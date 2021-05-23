

# Prometheus Exporter

## Overview

The Prometheus Exporter will format metrics into the Prometheus [line protocol](https://prometheus.io/docs/instrumenting/exposition_formats/) (a.k.a. *open metric format*) and expose it on an HTTP port (`http://<ADDR>:<PORT>/metrics`). Additionally, it serves a basic overview of available metrics and collectors on its root address (`http://<ADDR>:<PORT>`).

Don't forget to [update your Prometheus configuration](#configure-prometheus-to-scrape-from-harvest) and add a new target for each of the ports defined in your Harvest configuration.

## Design

The Exporter has two concurrent components that makes it possible to simultaneously serve push requests by collectors and pull requests by Prometheus scrapers.

<img src="prometheus.png" width="100%" align="center">


## Parameters

All parameters of the exporter are defined in the `Exporters` section of `harvest.yml`. We are planning to integrate the exporter with one of Prometheus' service-discovery options to simplify this in the future.


An overview of all parameters:


| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `port`                 | int, required | port of the HTTP server                          |                        |
| `local_http_addr`      | string, optional	| address of the HTTP server Harvest starts for Prometheus to scrape:<br />use `localhost` to serve only on the local machine<br />use `0.0.0.0` (default) if Prometheus is scrapping from another machine | `0.0.0.0` |
| `global_prefx`         | string, optional | add a prefix to all metrics (e.g. `netapp_`) |                        |
| `allow_addrs`          | list of strings, optional | allow access only if host matches any of the provided addresses | |
| `allow_addrs_regex`	 | list of strings, optional | allow access only if host address matches at least one of the regular expressions | |
| `cache_max_keep`       | string (Go duration format), optional | maximum amount of time metrics are cached (in case Prometheus does not timely collect the metrics) | `180s` |
| `add_meta_tags` |	bool, optional | add `HELP` and `TYPE` [metatags](https://prometheus.io/docs/instrumenting/exposition_formats/#comments-help-text-and-type-information) to metrics (currently no useful information, but required by some tools) | `false`	|


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

## Configure Prometheus to scrape from Harvest

As an example, if we defined four prometheus exporters at ports: 12990, 12991, 14567, and 14568 you need to add four sections to your `prometheus.yml`.

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