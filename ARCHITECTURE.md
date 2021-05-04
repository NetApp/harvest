# Architecture

This document describes the high-level architecture of Harvest.
If you want to familiarize yourself with the code base, you're in the right place!

## Bird's Eye View

Harvest consists of several programs. All, except Poller, are short-lived processes. Here is an overview of the important ones, followed by more detailed description in the following sections.

<center><img src="docs/harvest.png" width="60%"></center>

* **harvest**: the main executable and entry-point for the user, it's task is mainly to trigger the other components
* **manager**: starts, stops and shows the status of pollers
* **poller**: daemon process that polls a target system
* **config**: helps to validate and configure harvest
* **tools**: auxilary utilities, normally not related to the workflow of harvest itself

## Poller


<center><img src="docs/poller.png" width="80%" align="center"></center>

### Matrix


### Collectors 

<img src="docs/config.png" width="80%" align="center"><br />
Collectors are responsible for collecting metrics from data sources
* ZAPI collectors for: 
  * Performance
  * Shelf
  * Capacity
  * Process and Hardware monitors

### Plug-Ins

Plug-Ins post-process metrics from collectors, aggregate new counters, or trigger actions. 

They allow you to extend Harvest without writing a new collector, instead you can augment an existing one.

### Exporters 

Exporters encode metrics and send then to a time-series database. Out of the box
Harvest includes Prometheus and InfluxDB exporters.

### Tools
currently these include:
  * grafana - tool to export and import dashboards from a Grafana server
  * zapi -   

## Code Map

This section describes the important directories and data structures.



## Request Flows

## Plugins

## Life-cycle
