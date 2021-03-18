
# NetApp Harvest 2.0

The *swiss-army knife* for monitoring datacenters. The default package collects performance, capacity and hardware metrics from *ONTAP* clusters. New metrics can be collected by editing the config files. Metrics can be delivered to multiple databases - Prometheus, InfluxDB and Graphite - and displayed in Grafana dashboards.

Harvest 2.0 is not a collector, but a framework for running collectors and exporters concurrently. You are more than welcome to contribute your own collector, plugin or exporter (see our Developers documentation).

<img src="docs/examples/dashboard_cluster.png" width="40%" align="center"><img src="docs/examples/dashboard_shelf.png" width="40%" align="center">


## Requirements

A Linux system with libraries:
- `glibc`
- `libc6-dev`
- `tar`
  
Optional prerequisites:
- `dialog` or `whiptail` (used by the `config` utility)
- `openssl` (used by `config`)
- `python3` (used by the `grafana` utility)
  
Hardware requirements depend on number of pollers (i.e. number of clustered monitored) and collected metrics. With default configuration, for monitoring 10 clusteres, optimal resources are:

- CPU: 2 cores
- Memory: 1 GB
- Disk: 500 MB (mostly used by log files)

Harvest is compatible with:
- Prometheus: `2.24` or higher
- InfluxDB: `v2`
- Grafana: `7.4.2` or higher


# Installation / Upgrade

Download the latest package to your machine. For CentOS, RHEL, etc this is the latest `harvest-2*.x86_64.rpm` package. To build from source or run in a Docker container, download the latest tarball (`harvest-2*.tgz`).

## RPM-based installation

Install with `rpm`:

```sh
$ rpm -ivh harvest-2*.x86_64.rpm
```

Upgrade with `rpm`:

```sh
$ rpm -Uvh harvest-2*.x86_64.rpm
```

## Build from source
Requires Go 1.15 or higher, as well as internet connection to install go-dependencies.

```sh
$ tar -xzvf harvest-2*.tgz -C /opt/ && cd /opt/harvest2/
$ ./cmd/build.sh all
$ ./cmd/install.sh
```

# Quick start

## 1. Basic configuration

To start collecting metrics, you need to define at least one poller (collecting from a storage system) and one exporter (exporting to a DB) in your main configuration file: `harvest.yml` (located in `/etc/harvest/`). The default configuration comes with a pre-configured poller (`localsys`) which collects metrics local system. This is useful if you want to monitor resource usage by Harvest, otherwise it is safe to delete this poller.

The next step is to configure your Ontap clusters as pollers. There are two ways of doning this:

(1) You can run the config tool that will walk you through the steps of adding a cluster:

```sh
$ harvest config
```

(2) You can edit `harvest.yml` manually. See section Advanced configuration for more help.

## 2. Start Harvest

Start your *all* Harvest pollers as daemons:
```bash
$ harvest start
```
Or start specific poller(s):
```bash
$ harvest start jamaica grenada
```

(replace `jamaica` and `grenada` with the poller names that you defined in `harvest.yaml`). The logs of each poller can be found in `/var/log/harvest/`.

## 3. Import Grafana dashboards

The Grafana dashboards are located in `etc/harvest/grafana`. You can manually import the Grafana dashboards or use the `grafana` utility. It requires the address (hostname or IP) and port of the Grafana server. Port should emitted if the HTTP server is configured to redirect the URL. Use the `-d` flag for pointing to the directory from which the dashboards should be loaded. 

For example, to import the Prometheus-based dashboards from the directory `/opt/netapp-harvest/grafana/prometheus/` we will run (assuming `http://10.12.12.12:3000` points to our Grafana server):

```
$ harvest grafana -u 10.12.12.12 -p 3000 -d prometheus
```

The utility tool will ask for an API token which can be generated from the Grafana web-gui. Click on `Configuration` in the left menu bar (1), click on `API Keys` (2) and click on the button `New API Key`. Choose a Key name (3), choose `Editor` for role (4) and click on add (5). Copy the generated key and paste it in your terminal.

## 4. Verify the metrics

If you use a Prometheus Exporter, open a browser and navigate to [http://0.0.0.0:12990/](http://0.0.0.0:12990/) (replace `12990` with the port number of your poller). This is the info portal of the exporter which provides a real-time generated list of running collectors and names of exported metrics. 

The actual metric data that should be scraped by the Prometheus DB is available at [http://0.0.0.0:12990/metrics/](http://0.0.0.0:12990/metrics/). For more help on how to configure Prometheus DB, see section Prometheus.

If you can't access the URL, check the logs of your pollers. These are located in `/var/log/harvest/`.


# Advanced Configuration

If you need to edit the configuration of Harvest manually, you will find all configuration files in `/etc/harvest/`. The main configuration file, `harvest.yml`, consists of three sections, that are described below:


## Pollers
All pollers are defined in `harvest.yaml`, the main configuration file of Harvest, under the section `Pollers`. 

| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| Poller name (header)   | **required** |  poller name, user-defined value                 |                        |
| `datacenter`           | **required** |  datacenter name, user-defined value                               |                        |
| `addr`                  | required by some collectors |  IPv4 or FQDN of the target system                     |                        |
| `collectors`            | **required**              | list of collectors to run in this poller |   |
| `exporters` | **required**  | list of exporter names from the `Exporters` section. Note: this should be the exporter instance (e.g. `prometheus01`), not the exporter class (e.g. `Prometheus`)   |                   |
| `prometheus_port` | **optional** | local HTTP that the Prmometheus exporter will use. | |
| `auth_style`           | required by Zapi* collectors |  either `basic_auth` or `certificate_auth`  | `basic_auth` |
| `username`, `password` | required if `auth_style` is `basic_auth` |  |              |
| `cert`, `key`          | required if `auth_style` is `certificate_auth` | certificate and key files which should be in the directory `/etc/harvest/cert/`. If these two parameters are not provided files matching the poller name will be used (for example if poller name is `jamaica` than the files should be `jamaica.key` and `jamaica.cert`).                        |              |
| |  | |

## Defaults
This section is optional. If there are parameters identical for all your pollers (e.g. datacenter, authentication method), they can be grouped under section this section to safe space.

## Exporters

All exporters need two types of parameters:
- `exporter parameters` - defined in `harvest.yaml` under section `Exporters`
- `export_options` - which they should get from the datastructure (`Matrix`) that is emitted from collectors and plugins

The following two parameters are required for all exporters:

| parameter     | type         | description                                                                             | default      |
|---------------|--------------|-----------------------------------------------------------------------------------------|--------------|
| Exporter name (header) | **required** | Name of the exporter instance, this is a user-defined value |              |
| `exporter`    | **required** | Name of the exporter class (e.g. Prometheus, Graphite, InfluxDB, Http) which will be imported from the directory `harvest/exporters/`           |              |

Note: when we talk about *Prometheus Exporter*, *Graphite Exporter*, etc., we mean the Harvest modules that send the data to a database, do not confuse those names with the actual databases.

### Prometheus Exporter
***parameters:***
| parameter     | type         | description                                                                             | default      |
|---------------|--------------|-----------------------------------------------------------------------------------------|--------------|
| `url`    | optional  | Local address of the HTTP service (`localhost` or `127.0.0.1` makes the metrics accessible only on local machine, `0.0.0.0` makes it public).| `0.0.0.0` |
| `port`        | required  | Local port of the HTTP service. This value can be also defined under the poller section as `prometheus_port`.  |
| `allowed_urls`        | optional, list | List of clients that can access the HTTP service, each "URL" should be a hostname or IP address (regexes will be supported in the future). If the client is not in thist list, the HTTP request will be rejected. | allow all URLs |
| |  | |

Notice that you should define a new job in the configuration of your Prometheus database and add a target for each of the ports defined in Harvest configuration. As an example, let's assume we defined the port range `12990-12992` for the Prometheus Exporter. Open the configuration of the Prometheus database:

```bash
$ vim /etc/prometheus/prometheus.yml
```

Scroll down until the end of file and add the following lines:

```yaml
  - job_name: 'harvest'
    
    scrape_interval:     60s 
    static_configs:
    
      - targets: 
        - 'localhost:12990'
        - 'localhost:12991'
        - 'localhost:12992'
```
Notice that if Prometheus is not on the same machine as Harvest, then replace `localhost` with the IP address of your Harvest machine. Notice also that we set the scrape interval to 60s, since that matches with the lowest poll frequency of Harvest collectors (with default configuration). If you change the poll frequencies of Harvest collectors to a lower value, you should also change the scrape interval.
