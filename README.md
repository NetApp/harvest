
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

We provide pre-compiled binaries for Linux, RPMs, and Debs.

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

Work in progress. Coming soon

## Building from source

To build Harvest from source code, first make sure you have a working Go environment with [version 1.15 or greater installed](https://golang.org/doc/install). You'll also need an Internet connection to install go dependencies. If you need to build from an air-gapped machine, use `go mod vendor` from an Internet connected machine first, and then copy the `vendor` directory to the air-gapped machine.

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

The next step is to add pollers for your ONTAP clusters in the [Pollers](#pollers) section of the configuration file. Refer to the [Harvest Configuration](#harvest-configuration) Section for more details.

## 2. Start Harvest

Start all Harvest pollers as daemons:

```bash
$ bin/harvest start
```

Or start a specific poller(s):

```bash
$ bin/harvest start jamaica grenada
```

Replace `jamaica` and `grenada` with the poller names you defined in `harvest.yml`. The logs of each poller can be found in `/var/log/harvest/`.

## 3. Import Grafana dashboards

The Grafana dashboards are located in the `$HARVEST_HOME/grafana` directory. You can manually import the dashboards or use the `harvest grafana` command ([more documentation](cmd/tools/grafana/README.md)).

Note: the current dashboards specify Prometheus as the datasource. If you use the InfluxDB exporter, you will need to create your own dashboards.

## 4. Verify the metrics

If you use a Prometheus Exporter, open a browser and navigate to [http://0.0.0.0:12990/](http://0.0.0.0:12990/) (replace `12990` with the port number of your poller). This is the Harvest created HTTP end-point for your Prometheus exporter. This page provides a real-time generated list of running collectors and names of exported metrics. 

The metric data that's exposed for Prometheus to scrap is available at [http://0.0.0.0:12990/metrics/](http://0.0.0.0:12990/metrics/). For more help on how to configure Prometheus DB, see the [Prometheus exporter](cmd/exporters/prometheus/README.md) documentation.

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
| `ssl_cert`, `ssl_key`  | optional if `auth_style` is `certificate_auth` | Absolute paths to SSL (client) certificate and key used to authenticate with the target system.<br /><br />If not provided, the poller will look for `<hostname>.key` and `<hostname>.pem` in `$HARVEST_HOME/cert/`.<br/><br/>To create certificates for ONTAP systems, see the [Zapi documentation](cmd/collectors/zapi/README.md#authentication)                        |              |
| `use_insecure_tls`     | optional, bool |  If true, disable TLS verification when connecting to ONTAP cluster  | false         |
| `log_max_bytes`        |  | Maximum size of the log file before it will be rotated | `10000000` (10 mb) |
| `log_max_files`        |  | Number of rotated log files to keep | `10` |
| |  | | |

## Defaults
This section is optional. If there are parameters identical for all your pollers (e.g. datacenter, authentication method, login preferences), they can be grouped under this section. The poller section will be checked first and if the values aren't found there, the defaults will be consulted.

## Exporters

All exporters need two types of parameters:

- `exporter parameters` - defined in `harvest.yml` under `Exporters` section
- `export_options` - these options are defined in the `Matrix` datastructure that is emitted from collectors and plugins

The following two parameters are required for all exporters:

| parameter     | type         | description                                                                             | default      |
|---------------|--------------|-----------------------------------------------------------------------------------------|--------------|
| Exporter name (header) | **required** | Name of the exporter instance, this is a user-defined value |              |
| `exporter`    | **required** | Name of the exporter class (e.g. Prometheus, InfluxDB, Http) - these can be found under the `cmd/exporters/` directory           |              |

Note: when we talk about the *Prometheus Exporter* or *InfluxDB Exporter*, we mean the Harvest modules that send the data to a database, NOT the names used to refer to the actual databases.

### [Prometheus Exporter](cmd/exporters/prometheus/README.md)

### [InfluxDB Exporter](cmd/exporters/influxdb/README.md)

## Tools

This section is optional. You can uncomment the `grafana_api_token` key and add your Grafana API token so `harvest` does not prompt you for the key when importing dashboards.

```
Tools:
  #grafana_api_token: 'aaa-bbb-ccc-ddd'
```


## Configuring collectors

Collectors are configured by their own configuration files, which are subdirectories in [conf/](conf/). Each collector can define its own set of parameters.

### [Zapi](cmd/collectors/zapi/README.md)

### [ZapiPerf](cmd/collectors/zapiperf/README.md)

### [Unix](cmd/collectors/unix/README.md)

