## KeyPerf Collector

The KeyPerf collector is designed to gather performance counters from ONTAP objects that include a statistics field in their REST responses. This collector is an alternative to ZapiPerf and RestPerf collectors when these collectors cannot be used due to unavailable relevant APIs.

### Target System

The KeyPerf collector targets ONTAP systems that support the statistics field in their REST responses.

### Requirements

No additional SDK or software requirements are needed. It is recommended to create a read-only user for Harvest on the ONTAP system. For more details, refer to the [prepare monitored clusters](prepare-cdot-clusters.md#all-apis-read-only-approach) documentation.

### Metrics

The KeyPerf collector gathers a dynamic set of performance metrics, including IOPS, latency, and throughput. These metrics are extracted from the statistics fields in the ONTAP REST responses.

### Parameters

The parameters for the KeyPerf collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- KeyPerf configuration file (default: `conf/keyperf/default.yaml`)
- Each object has its own configuration file (located in `conf/keyperf/$version/`)

Except for `addr` and `datacenter`, all other parameters of the KeyPerf collector can be defined in any of these three files. Parameters defined in the lower-level file override parameters in the higher-level ones, allowing for individual object configuration or shared parameters across all objects.

### Collector Configuration File

This configuration file contains a list of objects that should be collected and the filenames of their templates (explained in the next section).

Additionally, this file contains the parameters that are applied as defaults to all objects. As mentioned before, any of these parameters can be defined in the Harvest or object configuration files as well.

| parameter          | type                           | description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |    default |
|--------------------|--------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------:|
| `use_insecure_tls` | bool, optional                 | skip verifying TLS certificate of the target system                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |      false |
| `client_timeout`   | duration (Go-syntax)           | how long to wait for server responses                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |        30s |
| `latency_io_reqd`  | int, optional                  | threshold of IOPs for calculating latency metrics (latencies based on very few IOPs are unreliable)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |         10 |
| `jitter`           | duration (Go-syntax), optional | Each Harvest collector runs independently, which means that at startup, each collector may send its REST queries at nearly the same time. To spread out the collector startup times over a broader period, you can use `jitter` to randomly distribute collector startup across a specified duration. For example, a `jitter` of `1m` starts each collector after a random delay between 0 and 60 seconds. For more details, refer to [this discussion](https://github.com/NetApp/harvest/discussions/2856).                                                                                                        |            |
| `schedule`         | list, required                 | the poll frequencies of the collector/object, should include exactly these three elements in the exact same other:                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |            |
| - `counter`        | duration (Go-syntax)           | poll frequency of updating the counter metadata cache                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |   24 hours |
| - `data`           | duration (Go-syntax)           | poll frequency of updating the data cache <br /><br />**Note** Harvest allows defining poll intervals on sub-second level (e.g. `1ms`), however keep in mind the following:<br /><ul><li>API response of an ONTAP system can take several seconds, so the collector is likely to enter failed state if the poll interval is less than `client_timeout`.</li><li>Small poll intervals will create significant workload on the ONTAP system, as many counters are aggregated on-demand.</li><li>Some metric values become less significant if they are calculated for very short intervals (e.g. latencies)</li></ul> |  1  minute |

The template should define objects in the `objects` section. Example:

```yaml
objects:
  Aggregate: aggr.yaml
```

For each object, we define the filename of the object configuration file. The object configuration files are located in subdirectories matching the ONTAP version that was used to create these files. It is possible to have multiple version-subdirectories for multiple ONTAP versions. At runtime, the collector will select the object configuration file that closest matches the version of the target ONTAP system.

### Object Configuration File

The object configuration file should contain the following parameters:

| Parameter        | Type                 | Description                                                 | Default |
|------------------|----------------------|-------------------------------------------------------------|---------|
| `name`           | string, **required** | Display name of the collector that will collect this object |         |
| `query`          | string, **required** | REST endpoint used to issue a REST request                  |         |
| `object`         | string, **required** | Short name of the object                                    |         |
| `counters`       | string               | List of counters to collect (see notes below)               |         |
| `plugins`        | list                 | Plugins and their parameters to run on the collected data   |         |
| `export_options` | list                 | Parameters to pass to exporters (see notes below)           |         |

#### Template Example

```yaml
name:                     Volume
query:                    api/storage/volumes
object:                   volume

counters:
  - ^^name                                => volume
  - ^^svm.name                            => svm
  - ^statistics.status                    => status
  - ^style                                => style
  - statistics.iops_raw.other             => other_ops
  - statistics.iops_raw.read              => read_ops
  - statistics.iops_raw.total             => total_ops
  - statistics.iops_raw.write             => write_ops
  - statistics.latency_raw.other          => other_latency
  - statistics.latency_raw.read           => read_latency
  - statistics.latency_raw.total          => avg_latency
  - statistics.latency_raw.write          => write_latency
  - statistics.throughput_raw.other       => other_data
  - statistics.throughput_raw.read        => read_data
  - statistics.throughput_raw.total       => total_data
  - statistics.throughput_raw.write       => write_data
  - statistics.timestamp(timestamp)       => timestamp
  - hidden_fields:
    - statistics
  - filter:
    - statistics.timestamp=!"-"

endpoints:
  - query: api/private/cli/volume
    counters:
      - ^^volume                          => volume
      - ^^vserver                         => svm
      - ^aggr_list                        => aggr
      - ^nodes                            => node

plugins:
  - Aggregator:
      # Plugin will create summary/average for each object
      # Any names after the object names will be treated as label names that will be added to instances
      - node
      - svm<>svm_vol

export_options:
  instance_keys:
    - aggr
    - node
    - style
    - svm
    - volume
```

### Counters

The `counters` section defines the list of counters to be collected. These counters can be labels or numeric metrics. The display name of a counter can be changed using `=>`.

#### Hidden Fields

Some fields are not returned by ONTAP unless explicitly requested. The `hidden_fields` parameter specifies additional fields to include in the REST response.

#### Filter

The `filter` parameter constrains the data returned by the endpoint, allowing for more targeted data retrieval. The filtering uses ONTAP's REST record filtering.

### Export Options

The `export_options` section defines how to handle the collected data. It includes parameters such as `instance_keys` and `instance_labels` to specify which labels to export with the metrics and instance labels.

### Endpoints

Refer to the [endpoints](configure-rest.md#endpoints) section for more details.

## Partial Aggregation

For more details about partial aggregation behavior and configuration, see [Partial Aggregation](partial-aggregation.md).