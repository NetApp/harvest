In order to understand how Harvest works, it's important to understand the following concepts:

<div class="grid cards" markdown>

- :material-reload: [Poller](#poller)
- :fontawesome-solid-bucket: [Collectors](#collectors)
- :simple-yaml: [Templates](#templates)
- :material-export: [Exporters](#exporters)
- :simple-grafana: [Dashboards](#dashboards)
- :material-filter-plus: [Port Map](#port-map)
</div>

In addition to the above concepts, Harvest uses the following software that you will want to be familiar with:

<div class="grid cards" markdown>

- :simple-prometheus: [Prometheus](#prometheus)
- :simple-influxdb: [InfluxDB](#influxdb)
- :simple-grafana: [Dashboards](#dashboards)
- :material-arrow-decision: [Prometheus Auto-discover](#prometheus-auto-discover)
- :simple-docker: [Docker](#docker)
- :material-cube-outline: [NABox](#nabox)

</div>

## Poller

The poller is the resident daemon process that coordinates the collectors and exporters. There will be one poller per monitored cluster.

## Collectors

Collectors implement the necessary protocol required to speak to the cluster. Harvest ships with ZAPI, REST, EMS, and StorageGRID collectors. Collectors use a set of per-object template files to determine which metrics to collect.

**More information:**

- [Configuring Collectors](configure-harvest-basic.md/#configuring-collectors)

## Templates

Templates define which metrics should be collected for an object (e.g. volume, lun, SVM, etc.). Harvest ships with a set of templates for each collector. The templates are written in YAML and are straightforward to read and modify. The templates are located in the `conf` directory.

There are two kinds of templates:

### Collector Templates
Collector templates (e.g. `conf/rest/default.yaml`) define which set of objects Harvest should collect from the system being monitored when that collector runs. For example, the `conf/rest/default.yaml` collector template defines which objects should be collected by the REST collector, while `conf/storagegrid/default.yaml` lists which objects should be collected by the StorageGRID collector.

### Object Templates
Object templates (e.g. `conf/rest/9.12.0/disk.yaml`) define which metrics should be collected and exported for an object. For example, the `disk.yaml` object template defines which disk metrics should be collected (e.g. `disk_bytes_per_sector`, `disk_stats_average_latency`, `disk_uptime`, etc.) 

**More information:**

- [Templates](configure-templates.md)
- [Templates and Metrics](resources/templates-and-metrics.md)
 
## Exporters

Exporters are responsible for encoding the collected metrics and making them available to time-series databases. Harvest ships with [Prometheus](#prometheus) and [InfluxDB](#influxdb) exporters. Harvest does not include Prometheus and InfluxDB, only the exporters for them. Prometheus and InfluxDB must be installed separately via Docker, NAbox, or other means.

## Prometheus

[Prometheus](https://prometheus.io/) is an open-source time-series database. It is a popular choice for storing and querying metrics. 

> Don't call us, we'll call you

None of the [pollers](#poller) know anything about Prometheus. That's because Prometheus pulls metrics from the poller's Prometheus exporter. The exporter creates an HTTP(s) endpoint that Prometheus scrapes on its own schedule. 

**More information:**

- [Prometheus Exporter](prometheus-exporter.md)

## InfluxDB

[InfluxDB](https://www.influxdata.com/) is an open-source time-series database. Harvest ships with some sample Grafana dashboards that are designed to work with InfluxDB. Unlike the Prometheus exporter, Harvest's InfluxDB exporter pushes metrics from the poller to InfluxDB via InfluxDB's line protocol. The exporter is compatible with InfluxDB v2.0. 

!!! note

    Harvest includes a subset of dashboards for InfluxDB. There is a richer set of dashboards available for Prometheus.

**More information:**

- [InfluxDB Exporter](influxdb-exporter.md)

## Dashboards

Harvest ships with a set of [Grafana](https://grafana.com/) dashboards that are primarily designed to work with Prometheus. The dashboards are located in the `grafana/dashboards` directory. Harvest does not include Grafana, only the dashboards for it. Grafana must be installed separately via Docker, NAbox, or other means.

Harvest includes CLI tools to import and export dashboards to Grafana. The CLI tools are available by running `bin/harvest grafana --help`

**More information:**

- [Import or Export Dashboards](dashboards.md)
- [How to Create A New Dashboard](dashboards.md#creating-a-custom-grafana-dashboard-with-harvest-metrics-stored-in-prometheus)

## Prometheus Auto-Discovery

Because of Prometheus's pull model, you need to configure Prometheus to tell it where to pull metrics from. There are two ways to tell Prometheus how to scrape Harvest: 1) listing each poller's address and port individually in Prometheus's config file or 2) using HTTP service discovery. 

Harvest's admin node implements Prometheus's HTTP service discovery API. Each poller registers its address and port with the admin node and Prometheus consults with the admin node for the list of targets it should scrape.

**More information:**

- [Configure Prometheus to scrape Harvest pollers](prometheus-exporter.md#configure-prometheus-to-scrape-harvest-pollers)
- [Prometheus Admin Node](prometheus-exporter.md#prometheus-http-service-discovery)
- [Prometheus HTTP Service Discovery](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config)

## Docker

Harvest runs natively in containers. The [Harvest container](https://github.com/NetApp/harvest/pkgs/container/harvest) includes the `harvest` and `poller` binaries as well as all templates and dashboards. If you want to standup Harvest, Prometheus, and Grafana all together, you can use the Docker Compose workflow. The Docker Compose workflow is a good way to quickly get started with Harvest.

**More information:**

- [Running Harvest in Docker](install/containers.md)
- [Running Harvest, Prometheus, and Grafana in Docker](install/containers.md#docker-compose)

## NABox

NABox is a separate virtual appliance (.ova) that acts as a front-end to Harvest and includes Promethus and Grafana setup to use with Harvest. NABox is a great option for customers that prefer a virtual appliance over containers.

**More information:**

- [NABox](https://nabox.org/documentation/installation/)

## Port Map

The default ports for ONTAP, Grafana, and Prometheus are shown below, along with three pollers. Poller1 is using the [PrometheusExporter](prometheus-exporter.md#static-scrape-targets) with a statically defined port in `harvest.yml`. Poller2 and Poller3 are using Harvest's admin node, [port range](prometheus-exporter.md#prometheus-http-service-discovery-and-port-range), and Prometheus HTTP service discovery. 

``` mermaid
graph LR
  Poller1 -->|:443|ONTAP1;
  Prometheus -->|:promPort1|Poller1;
  Prometheus -->|:promPort2|Poller2;
  Prometheus -->|:promPort3|Poller3;
  Prometheus -->|:8887|AdminNode;
  
  Poller2 -->|:443|ONTAP2;
  AdminNode <-->|:8887|Poller3;
  Poller3 -->|:443|ONTAP3;
  AdminNode <-->|:8887|Poller2;
  
  Grafana -->|:9090|Prometheus;
  Browser -->|:3000|Grafana;
```

- Grafana's default port is `3000` and is used to access the Grafana user-interface via a web browser
- Prometheus's default port is `9090` and Grafana talks to the Prometheus datasource on that port
- Prometheus scrapes each poller-exposed Prometheus port (`promPort1`, `promPort2`, `promPort3`)
- Poller2 and Poller3 are configured to use a PrometheusExporter with [port range](prometheus-exporter.md#prometheus-http-service-discovery-and-port-range). Each pollers picks a free port within the port_range and sends that port to the AdminNode.
- The Prometheus config file, `prometheus.yml` is updated with two scrape targets:
    1. the static `address:port` for Poller1
    2. the `address:port` for the AdminNode

- Poller1 creates an HTTP endpoint on the static port defined in the `harvest.yml` file
- All pollers use ZAPI or REST to communicate with ONTAP on port `443`

## Reference
- [Architecture.md](https://github.com/NetApp/harvest/blob/main/ARCHITECTURE.md)
