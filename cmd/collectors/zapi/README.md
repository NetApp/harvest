

# Zapi

Zapi collects data from ONTAP systems using the ZAPI protocol. The collector submits data as received from the target system, and does not perform any calculations or postprocessng. Since the attributes of most APIs have an irregular tree structure, sometimes a plugin will be required to collect all metrics from an API.

Note that the [ZapiPerf collector](../zapiperf/README.md) is an extension of this collector, therefore many parameters and configuration settings will coincide.

### Table of Contents
- [Target System](#target-system)
- [Requirements](#requirements)
- [Parameters](#parameters)
- [Metrics](#metrics)

## Target System
Target system can be any cDot or 7Mode ONTAP system. Any version is supported, however the default configuration files may not completely match with an older system.

## Requirements
No SDK or any other requirement. It is recommended to create a read-only user for Harvest on the ONTAP system (see the [Authentication document](../../../docs/AuthAndPermissions.md))

## Parameters

The parameters and configuration are similar to those of the [ZapiPerf collector](../zapiperf/README.md). Only the differences will be discussed below.

### Collector configuration file

Parameters different from ZapiPerf:


| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `schedule`             | required     | same as for ZapiPerf, but only two elements: `instance` and `data` (collector does not run a `counter` poll) ||
| `no_max_records`       | bool, optional | don't add `max-records` to the ZAPI request    |                        |
| `collect_only_labels`  | bool, optional | don't look for numeric metrics, only submit labels  (suppresses the `ErrNoMetrics` error)| |
| `only_cluster_instance` | bool, optional | don't look for instance keys and assume only instance is the cluster itself ||


### Object configuration file

The Zapi collector does not have the parameters `instance_key` and `override` parameters.

#### `counters`

This section contains the complete or partial attribute tree of the queried API. Since the collector does not get counter metadata from the ONTAP system, two additional symbols are used for non-numeric attributes:

- `^` used as a prefix indicates that the attribute should be stored as a label
- `^^` indicates that the attribute is a label and an instance key (i.e. a label that uniquely identifies an instance, such as `name`, `uuid`). If a single label does not uniquely identify an instance, then multiple instance keys should be indicated.

Additionally, the symbol `=>` can be used to set a custom display name for for both instance labels and numeric counters. Example:

```yaml
aggr-attributes:
  - aggr-raid-attributes:
    - ^aggregate-type    => type
	- disk-count     => disks
```

will force to use `aggr_type` and `aggr_disks` for the label and the metric respectively.

#### Creating/editing object configurations

The Zapi tool can help to create or edit subtemplates. Examples:

```sh
$ harvest zapi --poller <poller> show apis
  # will print list of apis that are available
  # usually apis with the "get-iter" suffix can provide useful metrics
$ harvest zapi --poller <poller> show attrs --api volume-get-iter
  # will print the attribute tree of the API
$ harvest zapi --poller <poller> show data --api volume-get-iter
  # will print raw data of the API attribute tree
```

Replace `<poller>` with the name of a poller that can connect to an ONTAP system.

## Metrics

The collector collects a dynamic set of metrics. Since most ZAPIs have a tree structure, the collector converts that structure into a flat metric represtantation. No postprocessing or calculation is performed on the collected data itself. 

As an example, the `aggr-get-iter` ZAPI provides the following partial attribute tree:

```yaml
aggr-attributes:
  - aggr-raid-attributes:
    - disk-count
  - aggr-snapshot-attributes:
    - files-total
```

The Zapi collector will convert this tree into two "flat" metrics: `aggr_raid_disk_count` and `aggr_snapshot_files_total`. (The algorithm to generate a name for the metrics will attempt to keep it as simple as possible, but sometimes it's useful to manually set a short display name (see [#counters](#counters)))