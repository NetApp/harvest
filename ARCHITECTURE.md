# Architecture

This document describes the high-level architecture of Harvest. If you want to familiarize yourself with the code base, you're in the right place!

Harvest has a strong emphasis on modular design, the core code-base is isolated from the code-base of secondary components - the philosophy is: adding new collectors, exporters or plugins should be simple.

## Bird's Eye View

Harvest consists of several processes/packages. All, except Poller, are short-lived processes. Here is an overview of the important ones:

<center><img src="docs/harvest.svg" width="75%"></center>

* **harvest** the main executable and entry-point. `bin/harvest` manages pollers, imports dashboards, generates required files, checks for potential configuration problem, starts the HTTP service discovery node, and more
* **poller** daemon process that polls a target system
* **config** helps to validate and configure harvest
* **tools** auxiliary utilities, normally not part of the workflow of harvest itself

### Poller

The Poller is a daemon process that polls metrics from a target system and emits to one or more databases. It is agnostic about the target system, the API used to retrieve data, and how to export to databases. The actual work is delegated to a set of collectors, plugins, and exporters.

<center><img src="docs/poller.png" width="70%" align="center"></center>

The package poller provides three interfaces:
* [Collector](cmd/poller/collector/collector.go)
* [Plugin](cmd/poller/plugin/plugin.go)
* [Exporter](cmd/poller/exporter/exporter.go)

It also provides three "abstract" types:
* [AbstractCollector](cmd/poller/collector/collector.go)
* [AbstractPlugin](cmd/poller/plugin/plugin.go)
* [AbstractExporter](cmd/poller/exporter/exporter.go)

These types implement most of the methods of their respective interfaces. A collector, plugin or exporter will usually "inherit" (or override) these methods, meaning that adding a new component requires implementing only a few methods.

### Configuration

One of the tasks of the Poller is to build a model of the configuration from CLI flags and your `harvest.yml` file. This model is passed to a poller's collectors and exporters:
* Poller Options (type *[poller.Options](cmd/poller/options/options.go))
* Params (type *[node.Node](pkg/tree/node/node.go))

<center><img src="docs/config.png" width="60%"></center><br />

For exporters, *Params*, is the exact parameters of the exporter as defined in `harvest.yml`. For collectors, *Params*, is a top-down merge of:
* poller parameters from `harvest.yml` (can include `addr`, `auth_style`, etc.)
* collector default template (can include poll frequency, list of counters, etc.)
* collector custom template (same)
See [conf/README.md](https://github.com/NetApp/harvest/blob/main/conf/README.md) for details.

Since the Poller know nothing about the system being monitored, it is the developer's and customer's responsibility to make sure parameters are in the right place.

### Collectors

Collectors are responsible for collecting metrics from a data source and writing them into a [matrix](https://github.com/NetApp/harvest/blob/main/pkg/matrix/README.md).

A collector's metrics are grouped together by the logical unit they describe (such as volume, node, process, file). If a collector contains multiple object, a new collector instance will be created for each object (example of such "multi-object" collectors are [Zapi](cmd/collectors/zapi/) and [ZapiPerf](cmd/collectors/zapiperf/). This means that customers only need to add a new template file if they want to collect a new object.

Most of the auxiliary jobs that a collector needs to do (such as initializing, running on scheduled time, reporting status to a poller, updating metadata and handling errors) are implemented by the `AbstractCollector`. Writing a new collector, typically only requires implementing the `PollData()` method.

### Plug-Ins

Plug-ins are optional and run as part of the collector. They allow you to post-process metrics from collectors, aggregate new metrics or trigger user-defined actions. They sometimes can be use to extend Harvest without writing a new collector. Some plugins will collect additional metrics on their own (example: [Zapi/SnapMirror](cmd/collectors/zapi/plugins/snapmirror/snapmirror.go)).

Harvest includes a set of built-in, collector-independent plugins for generic customization (see [documentation](cmd/poller/plugin/README.md)).

### Exporters 

Exporters write metrics to an external data source. A collector passes a matrix to an exporter, which encode the data to the required format and sends it to a database.

The way "sending" is done, can be quite different. For example, the [InfluxDB exporter](cmd/exporters/influxdb/README.md) pushes, via an HTTP PUT, to the InfluxDB, while the [Prometheus exporter](cmd/exporters/prometheus/README.md) will create a web endpoint that exposes the metrics and waits for Prometheus scrapers to collect them.

## Data Structures

### Matrix

A matrix provides storage for numerical and label data. It is the backbone of Harvest's architecture. Components of Harvest can work independently, yet interact with each other, only because they all know how to read/write a matrix.

The matrix has some constraints. For example, metrics should be typed, histograms should be converted into "flat" metrics, instance and metric names should be unique. For details see [documentation](pkg/matrix/README.md).

### Tree

The Tree data structure ([*node.Node](pkg/tree/node/node.go)) is used for unstructured and untyped data. It provides read/write methods that are independent of the underlying data format (`xml`, `yaml`, `json`). It is mainly used for API calls and template configuration.

Often collectors will receive XML from their target system, parse it into a tree, extract meaningful information, and write it into a matrix.

## Code Map

This section describes the directories of the project and how source files are organized:

### `/` 
The root directories contain scripts for building Harvest:
* `MakeFile` - script for building and installing Harvest
* `package` - script for building distribution packages (uses the subdirectories `deb/` and `rpm/`)
* `harvest.yml` - main configuration file

### `conf/`
Contains configuration files of collectors organized in subdirectories. The exact contents of each subdirectory depends on the design of the collector. Each collector should have a `default.yaml` and may optionally contain have a `custom.yaml` file.

### `cmd/`
This directory contains all packages that are compiled as executables
* `collectors/` - each subdirectory contains a different kind of collector, if you create a new collector, you should add a new subdirectory here
* `exporters/` - same as above, each subdirectory contains an exporter
* `harvest/` - contains the package and subpackages that are compiled as the main executable `harvest`
  * `version/` - harvest version and build info
* `poller/` - contains the poller program and subpackages:
  * `collector/` - provides interface Collector and type AbstractCollector
  * `exporter/` - provides interface Exporter and type AbstractExporter
  * `plugin/` - provides interface Plugin, type AbstractPlugin and built-in plugins in subdirectories
  * `options/` - provides type Options (poller start-up options)
  * `schedule/` - provides type Schedule
* `tools/` - each subdirectory is a tool, compiled as an executable
  
### `pkg/`
This directory contains all packages that are imported and used as shared libraries. List below contains only important ones:
* `api/` - each subdirectory is a package for communication with a (remote) system using some protocol
* `conf/` - package for importing harvest configuration file
* `errors/` - harvest errors
* `logging/` - wrapper methods around the [Zerolog](https://github.com/rs/zerolog)
* `matrix/` - the Matrix data structure
* `tree/` - the Tree data structure
* `util/` - helper functions

