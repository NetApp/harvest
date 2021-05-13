
# NetApp Harvest 2.0

The *swiss-army knife* for monitoring datacenters. The default package collects performance, capacity and hardware metrics from *ONTAP* clusters. New metrics can be collected by editing the config files. Metrics can be delivered to Prometheus and InfluxDB databases - and displayed in Grafana dashboards.

Harvest's architecture is flexible in how it collects, augments, and exports data. Think of it as a framework for running collectors and exporters concurrently. You are more than welcome to contribute your own collector, plugin or exporter (start with our [ARCHITECTURE.md](ARCHITECTURE.md)).

<img src="docs/examples/dashboard_cluster.png" width="40%" align="center"><img src="docs/examples/dashboard_shelf.png" width="40%" align="center">
## Requirements

Harvest is written in Go, which means it runs on recent Linux systems. It also runs on Macs, but the process isn't as smooth yet.

Optional prerequisites:
- `dialog` or `whiptail` (used by the `config` utility)
- `openssl` (used by `config`)
  
Hardware requirements depend on how many clusters you monitor and the number of metrics you chose to collect. With the default configuration, when  monitoring 10 clusters, we recommend:

- CPU: 2 cores
- Memory: 1 GB
- Disk: 500 MB (mostly used by log files)

Harvest is compatible with:
- Prometheus: `2.24` or higher
- InfluxDB: `v2`
- Grafana: `7.4.2` or higher

# Installation / Upgrade

We provide pre-compiled binaries for Linux, RPMs and Debs.

## Pre-compiled Binaries
Download the latest version of [Harvest](https://github.com/NetApp/harvest/releases/latest) from the releases tab and extract it.

```
wget https://github.com/NetApp/harvest/releases/download/v21.05.0/harvest-21.05.0.tar.gz
tar -xf harvest-21.05.0.tar.gz
cd harvest-21.05.0

# Run Harvest with the default unix localhost collector
bin/harvest start
```

## Redhat
Download the latest rpm of [Harvest](https://github.com/NetApp/harvest/releases/latest) from the releases tab and install with yum.

```
sudo yum install harvest.rpm
```

## Debian
Download the latest deb of [Harvest](https://github.com/NetApp/harvest/releases/latest) from the releases tab and install with apt.

```
sudo apt install harvest.deb
```

## Docker

WIP. Coming soon

## Building from source

To build Harvest from source code, first make sure you have a working Go environment with [version 1.15 or greater installed](https://golang.org/doc/install). You'll also need an Internet connection to install go dependencies. If you need to build from an air-gapped machine, use `go mod vendor` from an Internet connected machine first and then copy the `vendor` directory to the air-gapped machine.

Clone the repo and build everything.

```
git clone https://github.com/NetApp/harvest.git
cd harvest
make
bin/harvest version
```

Checkout the `Makefile` for other targets of interest.
# Quick start

## 1. Configuration file

Harvest's configuration information is defined in `harvest.yml`. There are a few ways to tell Harvest how to load this file:

* If you don't use the `--config` flag, the `harvest.yml` file located in the current working directory will be used

* If you specify the `--config` flag like so `harvest status --config /opt/harvest/harvest.yml`, Harvest will use that file

To start collecting metrics, you need to define at least one `poller` and one `exporter` in your  configuration file. The default configuration comes with a pre-configured poller named `unix` which collects metrics from the local system. This is useful if you want to monitor resource usage by Harvest and serves as a good example. Feel free to delete it if you want.

The next step is to add pollers for your ONTAP clusters in the [Pollers](#pollers) section of the configuration file. Refer to the [Harvest Configuration] Section(#harvest-configuration) for more details.

## 2. Start Harvest

Start *all* Harvest pollers as daemons:

```bash
$ bin/harvest start
```

Or start a specific poller(s):

```bash
$ bin/harvest start jamaica grenada
```

(replace `jamaica` and `grenada` with the poller names that you defined in `harvest.yml`). The logs of each poller can be found in `/var/log/harvest/`.

## 3. Import Grafana dashboards

The Grafana dashboards are located in the `grafana` directory. You can manually import the dashboards or use the `harvest grafana` utility. The utility requires the address (hostname or IP), port of the Grafana server, and a Grafana API token. The port can be omitted if Grafana is configured to redirect the URL. Use the `-d` flag to point to the directory that contains the dashboards.

### Grafana API token

The utility tool asks for an API token which can be generated from the Grafana web-gui. Click on `Configuration` in the left menu bar (1), click on `API Keys` (2) and click on the `New API Key` button. Choose a Key name (3), choose `Editor` for role (4) and click on add (5). Copy the generated key and paste it in your terminal or addd the token to the `Tools` section of your configuration file. (see below)

For example, let's say your Grafana server is on `http://my.grafana.server:3000` and you want to import the Prometheus-based dashboards from the `grafana` directory. You would run this:

```
$ bin/grafana import --addr my.grafana.server:3000 --directory grafana/prometheus
```

By default the dashboards are connected to the `Prometheus` datasource defined in Grafana. If your datasource has a different name, use the `--datasource` flag during import.

## 4. Verify the metrics

If you use a Prometheus Exporter, open a browser and navigate to [http://0.0.0.0:12990/](http://0.0.0.0:12990/) (replace `12990` with the port number of your poller). This is the Harvest created HTTP end-point for your Prometheus exporter. This page provides a real-time generated list of running collectors and names of exported metrics. 

The metric data that's exposed for Prometheus to scrap is available at [http://0.0.0.0:12990/metrics/](http://0.0.0.0:12990/metrics/). For more help on how to configure Prometheus DB, see the [Prometheus](#prometheus) section.

If you can't access the URL, check the logs of your pollers. These are located in `/var/log/harvest/`.

# Harvest Configuration 

The main configuration file, `harvest.yml`, consists of the following sections, described below:

## Pollers
All pollers are defined in `harvest.yml`, the main configuration file of Harvest, under the section `Pollers`. 

| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| Poller name (header)   | **required** |  poller name, user-defined value                 |                        |
| `datacenter`           | **required** |  datacenter name, user-defined value                               |                        |
| `addr`                 | required by some collectors |  IPv4 or FQDN of the target system                     |                        |
| `collectors`           | **required** | list of collectors to run for this poller |   |
| `exporters`            | **required** | list of exporter names from the `Exporters` section. Note: this should be the name of the exporter (e.g. `prometheus1`), not the value of the `exporter` key (e.g. `Prometheus`)   |                   |
| `auth_style`           | required by Zapi* collectors |  either `basic_auth` or `certificate_auth`  | `basic_auth` |
| `username`, `password` | required if `auth_style` is `basic_auth` |  |              |
| `cert`, `key`          | required if `auth_style` is `certificate_auth` | certificate and key files which should be in the directory `/etc/harvest/cert/`. If these two parameters are not provided files matching the poller name will be used (for example if poller name is `jamaica` than the files should be `jamaica.key` and `jamaica.cert`).                        |              |
| `use_insecure_tls`     | optional, bool |  If true, disable TLS verification when connecting to ONTAP cluster  | false         |
| `log_max_bytes`        |  | Maximum size of the log file before it will be rotated | `10000000` (10 mb) |
| `log_max_files`        |  | Number of rotated log files to keep | `10` |
| |  | | |

## Defaults
This section is optional. If there are parameters identical for all your pollers (e.g. datacenter, authentication method, login preferences), they can be grouped under this section. The poller section will be checked first and then the defaults consulted.

## Exporters

All exporters need two types of parameters:

- `exporter parameters` - defined in `harvest.yml` under `Exporters` section
- `export_options` - these options are defined in the `Matrix` datastructure that is emitted from collectors and plugins

The following two parameters are required for all exporters:

| parameter     | type         | description                                                                             | default      |
|---------------|--------------|-----------------------------------------------------------------------------------------|--------------|
| Exporter name (header) | **required** | Name of the exporter instance, this is a user-defined value |              |
| `exporter`    | **required** | Name of the exporter class (e.g. Prometheus, InfluxDB, Http) - these can be found under the `cmd/exporters/` directory           |              |

Note: when we talk about *Prometheus Exporter*, *InfluxDB Exporter*, etc., we mean the Harvest modules that send the data to a database, NOT the names used to refer to the actual databases.

### Prometheus Exporter
***parameters:***
| parameter     | type         | description                                                                             | default      |
|---------------|--------------|-----------------------------------------------------------------------------------------|--------------|
| `local_http_addr`    | optional  | Local address of the HTTP service (`localhost` or `127.0.0.1` makes the metrics accessible only on local machine, `0.0.0.0` makes it public).| `0.0.0.0` |
| `port`               | required  | Local HTTP service port Prometheus will scrape.  |
| `allow_addrs`        | optional, list | List of clients that can access the HTTP service, each "URL" should be a hostname or IP address. If the client is not in this list, the HTTP request will be rejected. | allow all URLs |
| `allow_addrs_regex`  | optional, list | Same as `allow_addrs`, but client will be only allowed if it matches any of the regular expressions | allow all URLs |
| `global_prefix`      | optional, string | globally add a prefix to all metrics, e.g settings this parameter to `netapp_`, would change the metric `cluster_status` (and all other metrics) into `netapp_cluster_status` | |
| `cache_max_keep`     | optional, duration | How long the HTTP end-point will cache the metrics. This can be useful if Prometheus is unavailable or if it scraps Harvest less frequently than the polling interval of collectors. Examples: `10s`, `1m`, `1h`, etc| `180s` |
| |  | |

Don't forget to update your Prometheus configuration and add a new target for each of the ports defined in Harvest configuration. As an example, if we defined four prometheus exporters at ports: 12990, 12991, 14567, and 14568 you need to add four sections to your `prometheus.yml`.

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
## Tools

This section is optional. You can uncomment the `grafana_api_token` key and add your Grafana API token so `harvest` does not prompt you for the key when importing dashboards.

```
Tools:
  #grafana_api_token: 'aaa-bbb-ccc-ddd'
```