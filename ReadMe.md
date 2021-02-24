

# NetApp Harvest 2.0

The *swiss-army knife* for monitoring datacenters. The default package collects performance, capacity and hardware metrics from *ONTAP* clusters. New metrics can be collected by editing the config files. Metrics can be delivered to multiple databases - Prometheus, InfluxDB and Graphite - and displayed in Grafana dashboards.

Harvest 2.0 is not a collector, but a framework for running collectors and exporters concurrently. You are more than welcome to contribute your own collector, plugin or exporter (see our Developers documentation).

<img src="docs/examples/dashboard_cluster.png" width="50%" align="center"><img src="docs/examples/dashboard_shelf.png" width="50%" align="center">
<br />

## Requirements

- linux system
- glibc
- tar
- python3 and pyyaml
- dialog or whiptail (optional)
- openssl (optional)
  
Hardware requirements depend configured metrics and number of pollers (i.e. number of clustered monitored). With default configuration, monitoring 10 clusteres, optimal resources are:

- CPU: 2 cores
- Memory: 1 GB
- Disk: 500 MB (mostly used by log files)
<br /><br />

## Deployment options
- [RPM-based installation](#rpm-based_installation)
- [Build from source](#build_from_source)
- [Docker containers](#docker_container)

<br />

# Installation

Download the latest package to your machine. For CentOS, RHEL, etc this is the latest `harvest-2***.x86_64.rpm` package. To build from source or run in a Docker container, download the latest tarball (`harvest-2***.tgz`).


## RPM-based installation

Install with `yum`:
```sh
$ yum install harvest-2***.x86_64.rpm
```
or with `rpm`:

```sh
$ rpm -ivh harvest-2***.x86_64.rpm
```

## Build from source
Requires Go 1.15 or higher, as well as internet connection to install go-dependencies.

```sh
$ tar -xzvf harvest-2***.tgz -C /opt/ && cd /opt/harvest2/
$ ./cmd/build.sh all
$ ./cmd/install.sh
```

## Docker container
*description coming soon*

<br />

# Quick start

## 1. Basic configuration

To start collecting metrics, you need to define at least one poller (talking to a storage system) and one exporter (talking to a DB). You can run the config tool that will run you through the steps of configuring both:

```sh
$ harvest config
```

This tool creates a main configuration file, `config.yaml` which you can also create or edit manually. The tool can also create client certificates and install them on your Ontap systems (we highly recommend using certificate-authentication rather than username/password).

## 2. Start Harvest

Start your *all* Harvest pollers as daemons:
```bash
$ harvest start
```
Or start specific poller(s):
```bash
$ harvest start jamaica grenada
```

(replace `jamaica` and `grenada` with the poller names that you defined in `config.yaml`). The logs of each poller can be found in the subdirectory `log`.