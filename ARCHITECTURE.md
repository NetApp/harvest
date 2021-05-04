# Architecture

This document describes the high-level architecture of Harvest. If you want to familiarize yourself with the code base, you're in the right place!

Harvest has a relatively small, but fairly dense and complex code base. However, the architecture has a strongly modular design: if you develop a collector, you don't need to understand what the poller does. Conversely, you don't need to understand how exporters work, if you develop a plugin or collector. But most of the time, you will need to understand the data structures.

Note: in this documentation we use uppercase for a process/program (e.g. *Poller*) and lowercase for the package (*poller*).

## Bird's Eye View

Harvest consists of several processes/packages. All, except Poller, are short-lived processes. Here is an overview of the important ones, followed by more detailed description in the following sections.

<center><img src="docs/harvest.png" width="60%"></center>

* **harvest**: the main executable and entry-point for the user, it's task is mainly to trigger the other components
* **manager**: starts, stops and shows the status of pollers
* **poller**: daemon process that polls a target system
* **config**: helps to validate and configure harvest
* **tools**: auxilary utilities, normally not related to the workflow of harvest itself

### Poller

The Poller is a deamon process that will poll metrics from one target system and emit to one or many databases. Important to understand is, that the Poller is agnostic about the target system, the API used to retrieve data and how its exported to the databases. The actual work is done by collectors, plugins and exporters respectively. All of these (except built-in plugins) are not part of the main program, they are compiled seperately as `.so` binaries and loaded dynamically by the Poller at runtime.

<center><img src="docs/poller.png" width="80%" align="center"></center>

The package poller provides three interfaces:
* [Collector](cmd/poller/collector/collector.go)
* [Plugin](cmd/poller/plugin/plugin.go)
* [Exporter](cmd/poller/exporter/exporter.go)

It also provides three "abstract" types:
* AbstractCollector
* AbstractPlugin
* AbstractExporter

These types implement most of the functions of their respective interfaces, so that when we write a new component for Harvest, there are only few functions that we need to implement.

### Configuration files

One of the tasks of the Poller is to parse CLI flags and configuration files and pass them over to collectors and exporters:
* Poller Options (type *[poller.Options](cmd/poller/options/options.go))
* Params (type *[tree.Node](pkg/tree/node/node.go))

<img src="docs/config.png" width="80%" align="center"><br />

For exporters, *Params*, is the exact parameters of the exporters as defined in `harvest.yml`. For collectors, *Params*, is a buttom-up merge of:
* poller parameters from `harvest.yml` (can include `addr`, `auth_style`, etc.)
* collector default template (can include poll frequency, list of counters, etc)
* collector custom template (same)

Since the Poller is agnostic about the system collectors will poll, it is the user's (and developer's) responsibility to make sure required parmeters are available in their right place.

### Collectors 

Collectors are responsible for collecting metrics from a data source and writing them into a Matrix instance. Most of the auxilary jobs that a collector needs to do (such as initializing itself, running on scheduled time, reporting status to Poller, updating metadata and handling errors) are implemented by the AbstractCollector. Hence, writing a new collector, most of the times, only requires implementing the `PollData()` method.

Collectors are "object-oriented", which means that metrics are grouped together by the logical unit that they describe (examples: volume, node, process, file). This constraint is imposed on the collectors by the Matrix: each combination of an object and metric name should be unique in the data-flow of Harvest.

If a collector class can collect multiple-objects, then it will clone a new collector for each one of them. Example of such "mullti-object" collectors are [Zapi](cmd/collectors/zapi/) and [ZapiPerf](cmd/collectors/zapiperf/). This means that the user only needs to add a new template file if they want to collector a new object.

Collectors run concurrently in their own goroutines.

### Plug-Ins

Plug-ins are optional and run as part of the collector. They will customize, post-process metrics from collectors, aggregate new metrics or trigger actions. They allow you to extend Harvest without writing a new collector, instead you can augment an existing one. Some plugins will collect additional metrics on their own (example: [Zapi/SnapMirror](cmd/collectors/zapi/plugins/snapmirror/)).

Harvest includes a range of built-in, collector-independent plugins for generic customization (see [documentation](cmd/poller/plugin/README.md)).


### Exporters 

Exporters encode metrics and send then to a time-series database. Out of the box
Harvest includes Prometheus and InfluxDB exporters.

### Tools

## Data Structures

## Matrix
See: [pkg/matrix/README.md](pkg/matrix/README.md)
## Tree

## Code Map

This section describes the important directories and data structures.


## Request Flows

## Life-cycle
