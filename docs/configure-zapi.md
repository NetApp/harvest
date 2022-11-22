!!! tip "What about REST?"

    ZAPI will reach end of availablity in ONTAP  9.13.1 released Q2 2023.
    Don't worry, Harvest has you covered. Switch to Harvest's REST collectors
    and collect idential metrics. See [REST Strategy](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-strategy.md) for more details. 

## Zapi Collector

The Zapi collectors uses the ZAPI protocol to collect data from ONTAP systems. The collector submits data as received
from the target system, and does not perform any calculations or post-processing. Since the attributes of most APIs have
an irregular tree structure, sometimes a plugin will be required to collect all metrics from an API.

The [ZapiPerf collector](#zapiperf-collector) is an extension of this collector, therefore they share many parameters
and configuration settings.

### Target System

Target system can be any cDot or 7Mode ONTAP system. Any version is supported, however the default configuration files
may not completely match with older systems.

### Requirements

No SDK or other requirements. It is recommended to create a read-only user for Harvest on the ONTAP system (see 
[prepare monitored clusters](prepare-cdot-clusters.md) for details)

### Metrics

The collector collects a dynamic set of metrics. Since most ZAPIs have a tree structure, the collector converts that
structure into a flat metric representation. No post-processing or calculation is performed on the collected data
itself.

As an example, the `aggr-get-iter` ZAPI provides the following partial attribute tree:

```yaml
aggr-attributes:
  - aggr-raid-attributes:
      - disk-count
  - aggr-snapshot-attributes:
      - files-total
```

The Zapi collector will convert this tree into two "flat" metrics: `aggr_raid_disk_count`
and `aggr_snapshot_files_total`. (The algorithm to generate a name for the metrics will attempt to keep it as simple as
possible, but sometimes it's useful to manually set a short display name. See [counters](configure-templates.md#counters)
for more details.

### Parameters

The parameters and configuration are similar to those of the [ZapiPerf collector](#zapiperf-collector). Only the
differences will be discussed below.

#### Collector configuration file

Parameters different from ZapiPerf:

| parameter               | type           | description                                                                                                  | default |
|-------------------------|----------------|--------------------------------------------------------------------------------------------------------------|---------|
| `schedule`              | required       | same as for ZapiPerf, but only two elements: `instance` and `data` (collector does not run a `counter` poll) ||
| `no_max_records`        | bool, optional | don't add `max-records` to the ZAPI request                                                                  |         |
| `collect_only_labels`   | bool, optional | don't look for numeric metrics, only submit labels  (suppresses the `ErrNoMetrics` error)                    |         |
| `only_cluster_instance` | bool, optional | don't look for instance keys and assume only instance is the cluster itself                                  ||

#### Object configuration file

The Zapi collector does not have the parameters `instance_key` and `override` parameters. The optional
parameter `metric_type` allows you to override the default metric type (`uint64`). The value of this parameter should be
one of the metric types supported by [the matrix data-structure](resources/matrix.md).

## ZapiPerf Collector

# ZapiPerf

ZapiPerf collects performance metrics from ONTAP systems using the ZAPI protocol. The collector is designed to be easily
extendable to collect new objects or to collect additional counters from already configured objects.

This collector is an extension of the [Zapi collector](#zapi-collector). The major difference between them is that
ZapiPerf collects only the performance (`perf`) APIs. Additionally, ZapiPerf always calculates final values from the
deltas of two subsequent polls.

## Metrics

The collector collects a dynamic set of metrics. The metric values are calculated from two consecutive polls (therefore,
no metrics are emitted after the first poll). The calculation algorithm depends on the `property` and `base-counter`
attributes of each metric, the following properties are supported:

| property | formula                                                                         | description                                                       |
|----------|---------------------------------------------------------------------------------|-------------------------------------------------------------------|
| raw      | x = x<sub>i</sub>                                                               | no post-processing, value **x** is submitted as it is             |
| delta    | x = x<sub>i</sub> - x<sub>i-1</sub>                                             | delta of two poll values, **x<sub>i<sub>** and **x<sub>i-1<sub>** |
| rate     | x = (x<sub>i</sub> - x<sub>i-1</sub>) / (t<sub>i</sub> - t<sub>i-1</sub>)       | delta divided by the interval of the two polls in seconds         |
| average  | x = (x<sub>i</sub> - x<sub>i-1</sub>) / (y<sub>i</sub> - y<sub>i-1</sub>)       | delta divided by the delta of the base counter **y**              |
| percent  | x = 100 * (x<sub>i</sub> - x<sub>i-1</sub>) / (y<sub>i</sub> - y<sub>i-1</sub>) | average multiplied by 100                                         |

## Parameters

The parameters of the collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- ZapiPerf configuration file (default: `conf/zapiperf/default.yaml`)
- Each object has its own configuration file (located in `conf/zapiperf/cdot/` and `conf/zapiperf/7mode/` for cDot and
  7Mode systems respectively)

Except for `addr`, `datacenter` and `auth_style`, all other parameters of the ZapiPerf collector can be
defined in either of these three files. Parameters defined in the lower-level file, override parameters in the
higher-level file. This allows the user to configure each objects individually, or use the same parameters for all
objects.

The full set of parameters are described [below](#zapiperf-configuration-file).

### Harvest configuration file

Parameters in poller section should define (at least) the address and authentication method of the target system:

| parameter              | type             | description                                                                    | default      |
|------------------------|------------------|--------------------------------------------------------------------------------|--------------|
| `addr`                 | string, required | address (IP or FQDN) of the ONTAP system                                       |              |
| `datacenter`           | string, required | name of the datacenter where the target system is located                      |              |
| `auth_style`           | string, optional | authentication method: either `basic_auth` or `certificate_auth`               | `basic_auth` |
| `ssl_cert`, `ssl_key`  | string, optional | full path of the SSL certificate and key pairs (when using `certificate_auth`) |              |
| `username`, `password` | string, optional | full path of the SSL certificate and key pairs (when using `basic_auth`)       |              |

### ZapiPerf configuration file

This configuration file (the "template") contains a list of objects that should be collected and the filenames of their
configuration (explained in the next section).

Additionally, this file contains the parameters that are applied as defaults to all objects. (As mentioned before, any
of these parameters can be defined in the Harvest or object configuration files as well).

| parameter          | type                 | description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | default |
|--------------------|----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| `use_insecure_tls` | bool, optional       | skip verifying TLS certificate of the target system                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | `false` |
| `client_timeout`   | duration (Go-syntax) | how long to wait for server responses                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | 30s     |
| `batch_size`       | int, optional        | max instances per API request                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | `500`   |
| `latency_io_reqd`  | int, optional        | threshold of IOPs for calculating latency metrics (latencies based on very few IOPs are unreliable)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | `100`   |
| `schedule`         | list, required       | the poll frequencies of the collector/object, should include exactly these three elements in the exact same other:                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |         |
| - `counter`        | duration (Go-syntax) | poll frequency of updating the counter metadata cache (example value: `1200s` = `20m`)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |         |
| - `instance`       | duration (Go-syntax) | poll frequency of updating the instance cache (example value: `600s` = `10m`)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |         |
| - `data`           | duration (Go-syntax) | poll frequency of updating the data cache (example value: `60s` = `1m`)<br /><br />**Note** Harvest allows defining poll intervals on sub-second level (e.g. `1ms`), however keep in mind the following:<br /><ul><li>API response of an ONTAP system can take several seconds, so the collector is likely to enter failed state if the poll interval is less than `client_timeout`.</li><li>Small poll intervals will create significant workload on the ONTAP system, as many counters are aggregated on-demand.</li><li>Some metric values become less significant if they are calculated for very short intervals (e.g. latencies)</li></ul> |         |

The template should define objects in the `objects` section. Example:

```yaml
objects:
  SystemNode: system_node.yaml
  HostAdapter: hostadapter.yaml
```

Note that for each object we only define the filename of the object configuration file. The object configuration files
are located in subdirectories matching to the ONTAP version that was used to create these files. It is possible to have
multiple version-subdirectories for multiple ONTAP versions. At runtime, the collector will select the object
configuration file that closest matches to the version of the target ONTAP system. (A mismatch is tolerated since
ZapiPerf will fetch and validate counter metadata from the system.)

### Object configuration file

The Object configuration file ("subtemplate") should contain the following parameters:

| parameter        | type                    | description                                                                         | default |
|------------------|-------------------------|-------------------------------------------------------------------------------------|---------|
| `name`           | string                  | display name of the collector that will collect this object                         |         |
| `object`         | string                  | short name of the object                                                            |         |
| `query`          | string                  | raw object name used to issue a ZAPI request                                        |         |
| `counters`       | list                    | list of counters to collect (see notes below)                                       |         |
| `instance_key`   | string                  | label to use as instance key (either `name` or `uuid`)                              |         |
| `override`       | list of key-value pairs | override counter properties that we get from ONTAP (allows circumventing ZAPI bugs) |         |
| `plugins`        | list                    | plugins and their parameters to run on the collected data                           |         |
| `export_options` | list                    | parameters to pass to exporters (see notes below)                                   |         |

#### `counters`

This section defines the list of counters that will be collected. These counters can be labels, numeric metrics or
histograms. The exact property of each counter is fetched from ONTAP and updated periodically.

Some counters require a "base-counter" for post-processing. If the base-counter is missing, ZapiPerf will still run, but
the missing data won't be exported.

The display name of a counter can be changed with `=>` (e.g., `nfsv3_ops => ops`). There's one conversion Harvest does
for you by default, the `instance_name` counter will be renamed to the value of `object`.

Counters that are stored as labels will only be exported if they are included in the `export_options` section.

#### `export_options`

Parameters in this section tell the exporters how to handle the collected data. The set of parameters varies by
exporter. For [Prometheus](prometheus-exporter.md) and [InfluxDB](influxdb-exporter.md)
exporters, the following parameters can be defined:

* `instances_keys` (list): display names of labels to export with each data-point
* `instance_labels` (list): display names of labels to export as a separate data-point
* `include_all_labels` (bool): export all labels with each data-point (overrides previous two parameters)
