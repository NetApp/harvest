!!! tip "What about REST?"

    ZAPI will reach end of availability in ONTAP  9.13.1 released Q2 2023.
    Don't worry, Harvest has you covered. Switch to Harvest's REST collectors
    and collect identical metrics. See [REST Strategy](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-strategy.md) for more details.

## Zapi Collector

The Zapi collectors use the ZAPI protocol to collect data from ONTAP systems. The collector submits data as received
from the target system, and does not perform any calculations or post-processing. Since the attributes of most APIs have
an irregular tree structure, sometimes a plugin will be required to collect all metrics from an API.

The [ZapiPerf collector](#zapiperf-collector) is an extension of this collector, therefore, they share many parameters
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

The parameters of the collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- ZAPI configuration file (default: `conf/zapi/default.yaml`)
- Each object has its own configuration file (located in `conf/zapi/$version/`)

Except for `addr` and `datacenter`, all other parameters of the ZAPI collector can be
defined in either of these three files.
Parameters defined in the lower-level file, override parameters in the higher-level ones.
This allows you to configure each object individually, or use the same parameters for all
objects.

The full set of parameters are described [below](#collector-configuration-file).

#### Collector configuration file

The parameters are similar to those of the [ZapiPerf collector](#zapiperf-collector).
Parameters different from ZapiPerf:

| parameter               | type                           | description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | default |
|-------------------------|--------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| `jitter`                | duration (Go-syntax), optional | Each Harvest collector runs independently, which means that at startup, each collector may send its ZAPI queries at nearly the same time. To spread out the collector startup times over a broader period, you can use `jitter` to randomly distribute collector startup across a specified duration. For example, a `jitter` of `1m` starts each collector after a random delay between 0 and 60 seconds. For more details, refer to [this discussion](https://github.com/NetApp/harvest/discussions/2856). |         |
| `schedule`              | required                       | same as for ZapiPerf, but only two elements: `instance` and `data` (collector does not run a `counter` poll)                                                                                                                                                                                                                                                                                                                                                                                                 |         |
| `no_max_records`        | bool, optional                 | don't add `max-records` to the ZAPI request                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |         |
| `collect_only_labels`   | bool, optional                 | don't look for numeric metrics, only submit labels  (suppresses the `ErrNoMetrics` error)                                                                                                                                                                                                                                                                                                                                                                                                                    |         |
| `only_cluster_instance` | bool, optional                 | don't look for instance keys and assume only instance is the cluster itself                                                                                                                                                                                                                                                                                                                                                                                                                                  |         |

#### Object configuration file

The Zapi collector does not have the parameters `instance_key` and `override` parameters. The optional
parameter `metric_type` allows you to override the default metric type (`uint64`). The value of this parameter should be
one of the metric types supported by [the matrix data-structure](resources/matrix.md).

The Object configuration file ("subtemplate") should contain the following parameters:

| parameter        | type                 | description                                                 | default |
|------------------|----------------------|-------------------------------------------------------------|---------|
| `name`           | string, **required** | display name of the collector that will collect this object |         |
| `query`          | string, **required** | REST endpoint used to issue a REST request                  |         |
| `object`         | string, **required** | short name of the object                                    |         |
| `counters`       | string               | list of counters to collect (see notes below)               |         |
| `plugins`        | list                 | plugins and their parameters to run on the collected data   |         |
| `export_options` | list                 | parameters to pass to exporters (see notes below)           |         |

#### Counters

This section defines the list of counters that will be collected. These counters can be labels, numeric metrics or
histograms. The exact property of each counter is fetched from ONTAP and updated periodically.

Some counters require a "base-counter" for post-processing. If the base-counter is missing, ZapiPerf will still run, but
the missing data won't be exported.

The display name of a counter can be changed with `=>` (e.g., `nfsv3_ops => ops`). There's one conversion Harvest does
for you by default, the `instance_name` counter will be renamed to the value of `object`.

Counters that are stored as labels will only be exported if they are included in the `export_options` section.

#### Export_options

Parameters in this section tell the exporters how to handle the collected data.

There are two different kinds of time-series that Harvest publishes: metrics and instance labels.

- Metrics are numeric data with associated labels (key-value pairs). E.g. `volume_read_ops_total{cluster="cluster1", node="node1", volume="vol1"} 123`. The `volume_read_ops_total` metric is exporting three labels: `cluster`, `node`, and `volume` and the metric value is `123`.
- Instance labels are named after their associated config object (e.g., `volume_labels`, `qtree_labels`, etc.). There will be one instance label for each object instance, and each instance label will contain a set of associated labels (key-value pairs) that are defined in the templates `instance_labels` parameter. E.g. `volume_labels{cluster="cluster1", node="node1", volume="vol1", svm="svm1"} 1`. The `volume_labels` instance label is exporting four labels: `cluster`, `node`, `volume`, and `svm`. Instance labels always export a metric value of `1`.

The `export_options` section allows you to define how to export these time-series.

* `instances_keys` (list): display names of labels to export to both metric and instance labels.
  For example, if you list the `svm` counter under `instances_keys`,
  that key-value will be included in all time-series metrics and all instance-labels.
* `instance_labels` (list): display names of labels to export with the corresponding instance label config object. For example, if you want the `volume` counter to be exported with the `volume_labels` instance label, you would list `volume` in the `instance_labels` section.
* `include_all_labels` (bool): exports all labels for all time-series metrics. If there are no metrics defined in the template, this option will do nothing. This option also overrides the previous two parameters. See also [collect_only_labels](#collector-configuration-file).

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
defined in either of these three files. 
Parameters defined in the lower-level file, override parameters in the higher-level file. 
This allows the user to configure each object individually, 
or use the same parameters for all objects.

The full set of parameters are described [below](#zapiperf-configuration-file).

### ZapiPerf configuration file

This configuration file (the "template") contains a list of objects that should be collected and the filenames of their
configuration (explained in the next section).

Additionally, this file contains the parameters that are applied as defaults to all objects. (As mentioned before, any
of these parameters can be defined in the Harvest or object configuration files as well).

| parameter          | type                           | description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | default |
|--------------------|--------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| `use_insecure_tls` | bool, optional                 | skip verifying TLS certificate of the target system                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | `false` |
| `client_timeout`   | duration (Go-syntax)           | how long to wait for server responses                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | 30s     |
| `batch_size`       | int, optional                  | max instances per API request                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | `500`   |
| `latency_io_reqd`  | int, optional                  | threshold of IOPs for calculating latency metrics (latencies based on very few IOPs are unreliable)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | `10`    |
| `jitter`           | duration (Go-syntax), optional | Each Harvest collector runs independently, which means that at startup, each collector may send its ZAPI queries at nearly the same time. To spread out the collector startup times over a broader period, you can use `jitter` to randomly distribute collector startup across a specified duration. For example, a `jitter` of `1m` starts each collector after a random delay between 0 and 60 seconds. For more details, refer to [this discussion](https://github.com/NetApp/harvest/discussions/2856).                                                                                                                             |         |
| `schedule`         | list, required                 | the poll frequencies of the collector/object, should include exactly these three elements in the exact same other:                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |         |
| - `counter`        | duration (Go-syntax)           | poll frequency of updating the counter metadata cache (example value: `20m`)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |         |
| - `instance`       | duration (Go-syntax)           | poll frequency of updating the instance cache (example value: `10m`)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |         |
| - `data`           | duration (Go-syntax)           | poll frequency of updating the data cache (example value: `1m`)<br /><br />**Note** Harvest allows defining poll intervals on sub-second level (e.g. `1ms`), however keep in mind the following:<br /><ul><li>API response of an ONTAP system can take several seconds, so the collector is likely to enter failed state if the poll interval is less than `client_timeout`.</li><li>Small poll intervals will create significant workload on the ONTAP system, as many counters are aggregated on-demand.</li><li>Some metric values become less significant if they are calculated for very short intervals (e.g. latencies)</li></ul> |         |

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

Parameters in this section tell the exporters how to handle the collected data.

There are two different kinds of time-series that Harvest publishes: metrics and instance labels.
 
- Metrics are numeric data with associated labels (key-value pairs). E.g. `volume_read_ops_total{cluster="cluster1", node="node1", volume="vol1"} 123`. The `volume_read_ops_total` metric is exporting three labels: `cluster`, `node`, and `volume` and the metric value is `123`.
- Instance labels are named after their associated config object (e.g., `volume_labels`, `nic_labels`, etc.).
  There will be one instance label for each object instance, 
  and each instance label will contain a set of associated labels
  (key-value pairs) that are defined in the templates `instance_labels` parameter.
  E.g. `volume_labels{cluster="cluster1", node="node1", volume="vol1", svm="svm1"} 1`. 
  The `volume_labels` instance label is exporting four labels:
  `cluster`, `node`, `volume`, and `svm`.
  Instance labels always export a metric value of `1`.

??? tip "Instance labels are rarely used with ZapiPerf templates"

    They can be useful for exporting labels that are not associated with a metric value.

The `export_options` section allows you to define how to export these time-series.

* `instances_keys` (list): display names of labels to export to both metric and instance labels.
  For example, if you list the `svm` counter under `instances_keys`,
  that key-value will be included in all time-series metrics and all instance-labels.
* `instance_labels` (list): display names of labels to export with the corresponding instance label config object. For example, if you want the `volume` counter to be exported with the `volume_labels` instance label, you would list `volume` in the `instance_labels` section.

### Filter

This guide provides instructions on how to use the `filter` feature in ZapiPerf. Filtering is useful when you need to query a subset of instances. For example, suppose you have a small number of high-value volumes from which you want Harvest to collect performance metrics every five seconds. Collecting data from all volumes at this frequency would be too resource-intensive. Therefore, filtering allows you to create/modify a template that includes only the high-value volumes.

#### Objects (Excluding Workload)

In ZapiPerf templates, you can set up filters under `counters`. Wildcards like * are useful if you don't want to specify all instances. Please note, ONTAP Zapi filtering does not support regular expressions, only wildcard matching with `*`.

For instance, to filter `volume` performance instances by instance name where the name is `NS_svm_nvme` or contains `Test`, use the following configuration in ZapiPerf `volume.yaml` under `counters`:

```yaml
counters:
  ...
  - filter:
     - instance_name=NS_svm_nvme|instance_name=*Test*
```

You can define multiple values within the filter array. These will be interpreted as `AND` conditions by ONTAP. Alternatively, you can specify a complete expression within a single array element, as described in the ONTAP filtering section below.

??? info "ONTAP Filtering Details"

    For a better understanding of ONTAP's filtering mechanism, it allows the use of `filter-data` for the `perf-object-instance-list-info-iter` Zapi.

    The `filter-data` is a string that signifies filter data, adhering to the format: `counter_name=counter_value`. You can define multiple pairs, separated by either a comma (",") or a pipe ("|").

    Here's the interpretation:

    - A comma (",") signifies an AND operation.
    - A pipe ("|") signifies an OR operation.
    - The precedence order is AND first, followed by OR.

    For instance, the filter string `instance_name=volA,vserver_name=vs1|vserver_name=vs2` translates to `(instance_name=volA && vserver_name=vs1) || (vserver_name=vs2)`.

    This filter will return instances on Vserver `vs1` named `volA`, and all instances on Vserver `vs2`.

#### Workload Templates

Performance workload templates require a different syntax because instances are retrieved from the `qos-workload-get-iter` ZAPI instead of `perf-object-instance-list-info-iter`.

The `qos-workload-get-iter` ZAPI supports filtering on the following fields:

- workload-uuid
- workload-name
- workload-class
- wid
- category
- policy-group
- vserver
- volume
- lun
- file
- qtree
- read-ahead
- max-throughput
- min-throughput
- is-adaptive
- is-constituent

You can include these fields under the `filter` parameter. For example, to filter Workload performance instances by `workload-name` where the name contains `NS` or `Test` and `vserver` is `vs1`, use the following configuration in ZapiPerf `workload.yaml` under `counters`:

```yaml
counters:
  ...
  - filter:
      - workload-name: "*NS*|*Test*"
      - vserver: vs1
```   

### Partial Aggregation

For more details about partial aggregation behavior and configuration, see [Partial Aggregation](partial-aggregation.md).