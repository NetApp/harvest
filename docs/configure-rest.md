## Rest Collector

The Rest collectors uses the REST protocol to collect data from ONTAP systems.

The [RestPerf collector](#restperf-collector) is an extension of this collector, therefore they share many parameters
and configuration settings.

### Target System

Target system can be cDot ONTAP system. 9.12.1 and after are supported, however the default configuration files
may not completely match with all versions.
See [REST Strategy](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-strategy.md) for more details.

### Requirements

No SDK or other requirements. It is recommended to create a read-only user for Harvest on the ONTAP system (see
[prepare monitored clusters](prepare-cdot-clusters.md#all-apis-read-only-approach) for details)

### Metrics

The collector collects a dynamic set of metrics. ONTAP returns JSON documents and
Harvest allows you to define templates to extract values from the JSON document via a dot notation path. You can view
ONTAP's full set of REST APIs by
visiting `https://docs.netapp.com/us-en/ontap-automation/reference/api_reference.html#access-a-copy-of-the-ontap-rest-api-reference-documentation`

As an example, the `/api/storage/aggregates` endpoint, lists all data aggregates in the cluster. Below is an example
response from this endpoint:

```json
{
  "records": [
    {
      "uuid": "3e59547d-298a-4967-bd0f-8ae96cead08c",
      "name": "umeng_aff300_aggr2",
      "space": {
        "block_storage": {
          "size": 8117898706944,
          "available": 4889853616128
        }
      },
      "state": "online",
      "volume_count": 36
    }
  ]
}
```

The Rest collector will take this document, extract the `records` section and convert the metrics above
into: `name`, `space.block_storage.size`, `space.block_storage.available`, `state` and `volume_count`. Metric names will
be taken, as is, unless you specify a short display name. See [counters](configure-templates.md#counters) for more
details.

### Parameters

The parameters of the collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- Rest configuration file (default: `conf/rest/default.yaml`)
- Each object has its own configuration file (located in `conf/rest/$version/`)

Except for `addr` and `datacenter`, all other parameters of the Rest collector can be
defined in either of these three files. Parameters defined in the lower-level file, override parameters in the
higher-level ones. This allows you to configure each object individually, or use the same parameters for all
objects.

The full set of parameters are described [below](#collector-configuration-file).

#### Collector configuration file

This configuration file contains a list of objects that should be collected and the filenames of their templates (
explained in the next section).

Additionally, this file contains the parameters that are applied as defaults to all objects. As mentioned before, any
of these parameters can be defined in the Harvest or object configuration files as well.

| parameter        | type                 | description                                                             | default   |
|------------------|----------------------|-------------------------------------------------------------------------|-----------|
| `client_timeout` | duration (Go-syntax) | how long to wait for server responses                                   | 30s       |
| `schedule`       | list, **required**   | how frequently to retrieve metrics from ONTAP                           |           |
| - `data`         | duration (Go-syntax) | how frequently this collector/object should retrieve metrics from ONTAP | 3 minutes |

The template should define objects in the `objects` section. Example:

```yaml
objects:
  Aggregate: aggr.yaml
```

For each object, we define the filename of the object configuration file. The object configuration files
are located in subdirectories matching the ONTAP version that was used to create these files. It is possible to
have multiple version-subdirectories for multiple ONTAP versions. At runtime, the collector will select the object
configuration file that closest matches the version of the target ONTAP system.

### Object configuration file

The Object configuration file ("subtemplate") should contain the following parameters:

| parameter        | type                 | description                                                 | default |
|------------------|----------------------|-------------------------------------------------------------|---------|
| `name`           | string, **required** | display name of the collector that will collect this object |         |
| `query`          | string, **required** | REST endpoint used to issue a REST request                  |         |
| `object`         | string, **required** | short name of the object                                    |         |
| `counters`       | string               | list of counters to collect (see notes below)               |         |
| `plugins`        | list                 | plugins and their parameters to run on the collected data   |         |
| `export_options` | list                 | parameters to pass to exporters (see notes below)           |         |

#### `counters`

This section defines the list of counters that will be collected. These counters can be labels, numeric metrics or
histograms. The exact property of each counter is fetched from ONTAP and updated periodically.

The display name of a counter can be changed with `=>` (e.g., `space.block_storage.size => space_total`).

Counters that are stored as labels will only be exported if they are included in the `export_options` section.

#### `export_options`

Parameters in this section tell the exporters how to handle the collected data. The set of parameters varies by
exporter. For [Prometheus](prometheus-exporter.md) and [InfluxDB](influxdb-exporter.md)
exporters, the following parameters can be defined:

* `instances_keys` (list): display names of labels to export with each data-point
* `instance_labels` (list): display names of labels to export as a separate data-point
* `include_all_labels` (bool): export all labels with each data-point (overrides previous two parameters)

## RestPerf Collector

RestPerf collects performance metrics from ONTAP systems using the REST protocol. The collector is designed to be easily
extendable to collect new objects or to collect additional counters from already configured objects.

This collector is an extension of the [Rest collector](#rest-collector). The major difference between them is that
RestPerf collects only the performance (`perf`) APIs. Additionally, RestPerf always calculates final values from the
deltas of two subsequent polls.

## Metrics

RestPerf metrics are calculated the same as ZapiPerf metrics. More details about how
performance metrics are calculated can be found [here](configure-zapi.md#metrics_1).

## Parameters

The parameters of the collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- RestPerf configuration file (default: `conf/restperf/default.yaml`)
- Each object has its own configuration file (located in `conf/restperf/$version/`)

Except for `addr`, `datacenter` and `auth_style`, all other parameters of the RestPerf collector can be
defined in either of these three files. Parameters defined in the lower-level file, override parameters in the
higher-level file. This allows the user to configure each objects individually, or use the same parameters for all
objects.

The full set of parameters are described [below](#restperf-configuration-file).

### RestPerf configuration file

This configuration file (the "template") contains a list of objects that should be collected and the filenames of their
configuration (explained in the next section).

Additionally, this file contains the parameters that are applied as defaults to all objects. (As mentioned before, any
of these parameters can be defined in the Harvest or object configuration files as well).

| parameter          | type                 | description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |    default |
|--------------------|----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------:|
| `use_insecure_tls` | bool, optional       | skip verifying TLS certificate of the target system                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |      false |
| `client_timeout`   | duration (Go-syntax) | how long to wait for server responses                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |        30s |
| `latency_io_reqd`  | int, optional        | threshold of IOPs for calculating latency metrics (latencies based on very few IOPs are unreliable)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |        100 |
| `schedule`         | list, required       | the poll frequencies of the collector/object, should include exactly these three elements in the exact same other:                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |            |
| - `counter`        | duration (Go-syntax) | poll frequency of updating the counter metadata cache                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | 20 minutes |
| - `instance`       | duration (Go-syntax) | poll frequency of updating the instance cache                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | 10 minutes |
| - `data`           | duration (Go-syntax) | poll frequency of updating the data cache <br /><br />**Note** Harvest allows defining poll intervals on sub-second level (e.g. `1ms`), however keep in mind the following:<br /><ul><li>API response of an ONTAP system can take several seconds, so the collector is likely to enter failed state if the poll interval is less than `client_timeout`.</li><li>Small poll intervals will create significant workload on the ONTAP system, as many counters are aggregated on-demand.</li><li>Some metric values become less significant if they are calculated for very short intervals (e.g. latencies)</li></ul> |  1  minute |

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
RestPerf will fetch and validate counter metadata from the system.)

### Object configuration file

Refer [Object configuration file](configure-rest.md#object-configuration-file)

#### `counters`

Refer [Counters](configure-rest.md#counters)

Some counters require a "base-counter" for post-processing. If the base-counter is missing, RestPerf will still run, but
the missing data won't be exported.

#### `export_options`

Refer [Export Options](configure-rest.md#export_options)
