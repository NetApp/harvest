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

| parameter        | type                           | description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | default   |
|------------------|--------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------|
| `client_timeout` | duration (Go-syntax)           | how long to wait for server responses                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |       30s |
| `jitter`         | duration (Go-syntax), optional | Each Harvest collector runs independently, which means that at startup, each collector may send its REST queries at nearly the same time. To spread out the collector startup times over a broader period, you can use `jitter` to randomly distribute collector startup across a specified duration. For example, a `jitter` of `1m` starts each collector after a random delay between 0 and 60 seconds. For more details, refer to [this discussion](https://github.com/NetApp/harvest/discussions/2856). |           |
| `schedule`       | list, **required**             | how frequently to retrieve metrics from ONTAP                                                                                                                                                                                                                                                                                                                                                                                                                                                                |           |
| - `counter`      | duration (Go-syntax)           | poll frequency of updating the counter metadata cache                                                                                                                                                                                                                                                                                                                                                                                                                                                        |  24 hours |
| - `data`         | duration (Go-syntax)           | how frequently this collector/object should retrieve metrics from ONTAP                                                                                                                                                                                                                                                                                                                                                                                                                                      | 3 minutes |

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

#### Template Example:

```yaml
name:                     Volume
query:                    api/storage/volumes
object:                   volume

counters:
  - ^^name                                        => volume
  - ^^svm.name                                    => svm
  - ^aggregates.#.name                            => aggr
  - ^anti_ransomware.state                        => antiRansomwareState
  - ^state                                        => state
  - ^style                                        => style
  - space.available                               => size_available
  - space.overwrite_reserve                       => overwrite_reserve_total
  - space.overwrite_reserve_used                  => overwrite_reserve_used
  - space.percent_used                            => size_used_percent
  - space.physical_used                           => space_physical_used
  - space.physical_used_percent                   => space_physical_used_percent
  - space.size                                    => size
  - space.used                                    => size_used
  - hidden_fields:
      - anti_ransomware.state
      - space
  - filter:
      - name=*harvest*

plugins:
  - LabelAgent:
      exclude_equals:
        - style `flexgroup_constituent`

export_options:
  instance_keys:
    - aggr
    - style
    - svm
    - volume
  instance_labels:
    - antiRansomwareState
    - state
```

#### Counters

This section defines the list of counters that will be collected. These counters can be labels, numeric metrics or
histograms. The exact property of each counter is fetched from ONTAP and updated periodically.

The display name of a counter can be changed with `=>` (e.g., `space.block_storage.size => space_total`).

Counters that are stored as labels will only be exported if they are included in the `export_options` section.

The `counters` section allows you to specify `hidden_fields` and `filter` parameters. Please find the detailed explanation below.

##### Hidden_fields

There are some fields that ONTAP will not return unless you explicitly ask for them, even when using the URL parameter `fields=**`. `hidden_fields` is how you tell ONTAP which additional fields it should include in the REST response.

##### Filter

The `filter` is used to constrain the data returned by the endpoint, allowing for more targeted data retrieval. The filtering uses ONTAP's REST record filtering. The example above asks ONTAP to only return records where a volume's name matches `*harvest*`.

If you're familiar with ONTAP's REST record filtering, the [example](#template-example) above would become `name=*harvest*` and appended to the final URL like so:

```
https://CLUSTER_IP/api/storage/volumes?fields=*,anti_ransomware.state,space&name=*harvest*
```

Refer to the ONTAP API specification, sections: `query parameters` and `record filtering`, for more details.

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

#### Endpoints

In Harvest REST templates, `endpoints` are additional queries that enhance the data collected from the main query. The main query, identified by the `query` parameter, is the primary REST API for data collection. For example, the main query for a `disk` object is `api/storage/disks`.
This main query collects disk objects from the ONTAP API and converts them into a [matrix](resources/matrix.md).

Typically `endpoints` are used to query the private CLI to add metrics that are not available via ONTAP's public REST API.
Within the `endpoints` section of a Harvest REST template, you can define multiple endpoint entries. Each entry supports its own `query` and associated `counters`, allowing you to collect additional metrics or labels from various API.
These additional metrics or labels are combined with the main matrix via a key. The key is denoted by the `^^` notation in the counters of both the main query and the `endpoints`.

If the `instance_add` flag is set to `true` within an endpoint, new records will be created rather than modifying existing ones.
This allows for the collection of additional instances without altering the existing matrix.

In the example below, the `endpoints` section makes an additional query to `api/private/cli/disk`, which collects metrics such as `stats_io_kbps`, `stats_sectors_read`, and `stats_sectors_written`. The `uuid` is the key that links the data from the `api/storage/disks` and `api/private/cli/disk` API.
The `type` label from the `api/private/cli/disk` endpoint is included as outlined in the `export_options`.

```yaml
name:             Disk
query:            api/storage/disks
object:           disk

counters:
  - ^^uid                       => uuid
  - ^bay                        => shelf_bay
  - ^container_type
  - ^home_node.name             => owner_node
  - ^model
  - ^name                       => disk
  - ^node.name                  => node
  - ^node.uuid
  - ^outage.reason              => outage
  - ^serial_number
  - ^shelf.uid                  => shelf
  - ^state
  - bytes_per_sector            => bytes_per_sector
  - sector_count                => sectors
  - stats.average_latency       => stats_average_latency
  - stats.power_on_hours        => power_on_hours
  - usable_size

endpoints:
  - query: api/private/cli/disk
    counters:
      - ^^uid                   => uuid
      - ^type
      - disk_io_kbps_total      => stats_io_kbps
      - sectors_read            => stats_sectors_read
      - sectors_written         => stats_sectors_written

plugins:
  - Disk
  - LabelAgent:
      value_to_num:
        - new_status outage - - `0` #ok_value is empty value, '-' would be converted to blank while processing.
      join:
        - index `_` node,disk
  - MetricAgent:
      compute_metric:
        - uptime MULTIPLY stats.power_on_hours 60 60 #convert to second for zapi parity

export_options:
  instance_keys:
    - disk
    - index
    - node
  instance_labels:
    - container_type
    - failed
    - model
    - outage
    - owner_node
    - serial_number
    - shared
    - shelf
    - shelf_bay
    - type
```

Example with `instance_add`


In the example below, the use of `instance_add` is necessary to collect both flexvols and their flexgroup constituents, which cannot be retrieved in a single ONTAP API call.
Therefore, two separate API calls are required. Initially, volume data excluding flexgroups is gathered from `api/storage/volumes` and added to the `matrix`.
The endpoint with `instance_add: true` enables the collection and addition of flexgroup constituent volumes to the matrix.
Subsequently, the endpoint query `api/private/cli/volume` is used to add `aggr` and `node` labels to the data collected from both the main query and the first endpoint query,
modifying the matrix with additional details.

This example is from `conf/keyperf/9.15.0/volume.yaml`.

```yaml
name:                     Volume
query:                    api/storage/volumes
object:                   volume

counters:
  - ^^name                                    => volume
  - ^^svm.name                                => svm
  - ^statistics.status                        => status
  - ^style                                    => style
  - statistics.iops_raw.other                 => other_ops
  - statistics.iops_raw.read                  => read_ops
  - statistics.iops_raw.total                 => total_ops
  - statistics.iops_raw.write                 => write_ops
  - statistics.latency_raw.other              => other_latency
  - statistics.latency_raw.read               => read_latency
  - statistics.latency_raw.total              => avg_latency
  - statistics.latency_raw.write              => write_latency
  - statistics.throughput_raw.other           => other_data
  - statistics.throughput_raw.read            => read_data
  - statistics.throughput_raw.total           => total_data
  - statistics.throughput_raw.write           => write_data
  - statistics.timestamp(timestamp)           => timestamp
  - hidden_fields:
      - statistics
  - filter:
      - statistics.timestamp=!"-"
      - style=!flexgroup     # collected via endpoints

endpoints:
  - query: api/storage/volumes
    instance_add: true
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
          - is_constituent=true

  - query: api/private/cli/volume
    counters:
      - ^^volume                              => volume
      - ^^vserver                             => svm
      - ^aggr_list                            => aggr
      - ^nodes                                => node
      - filter:
          - is_constituent=*
```


## RestPerf Collector

RestPerf collects performance metrics from ONTAP systems using the REST protocol. The collector is designed to be easily
extendable to collect new objects or to collect additional counters from already configured objects.

This collector is an extension of the [Rest collector](#rest-collector). The major difference between them is that
RestPerf collects only the performance (`perf`) APIs. Additionally, RestPerf always calculates final values from the
deltas of two subsequent polls.

### Metrics

RestPerf metrics are calculated the same as ZapiPerf metrics. More details about how
performance metrics are calculated can be found [here](configure-zapi.md#metrics_1).

### Parameters

The parameters of the collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- RestPerf configuration file (default: `conf/restperf/default.yaml`)
- Each object has its own configuration file (located in `conf/restperf/$version/`)

Except for `addr`, `datacenter` and `auth_style`, all other parameters of the RestPerf collector can be
defined in either of these three files. Parameters defined in the lower-level file, override parameters in the
higher-level file. This allows the user to configure each objects individually, or use the same parameters for all
objects.

The full set of parameters are described [below](#restperf-configuration-file).

#### RestPerf configuration file

This configuration file (the "template") contains a list of objects that should be collected and the filenames of their
configuration (explained in the next section).

Additionally, this file contains the parameters that are applied as defaults to all objects. (As mentioned before, any
of these parameters can be defined in the Harvest or object configuration files as well).

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

#### Counters

See [Counters](configure-rest.md#counters)

Some counters require a "base-counter" for post-processing. If the base-counter is missing, RestPerf will still run, but
the missing data won't be exported.

#### Export_options

See [Export Options](configure-rest.md#export_options)

### Filter

This guide provides instructions on how to use the `filter` feature in RestPerf. Filtering is useful when you need to query a subset of instances. For example, suppose you have a small number of high-value volumes from which you want Harvest to collect performance metrics every five seconds. Collecting data from all volumes at this frequency would be too resource-intensive. Therefore, filtering allows you to create/modify a template that includes only the high-value volumes.

#### Objects (Excluding Workload)

In RestPerf templates, you can set up filters under `counters`. Wildcards like `*` are useful if you don't want to specify all instances.

For instance, to filter `volume` performance instances by volume name, use the following configuration in RestPerf `volume.yaml` under `counters`. This example will return the volumes named `NS_svm_nvme` or the volumes with `Test` in their name.

```yaml
counters:
...
  - filter:
    - query=NS_svm_nvme|*Test*
    - query_fields=properties.value
```

### Workload Templates

Performance workload templates require a different syntax because instances are retrieved from the `api/storage/qos/workloads` Rest call.

The `api/storage/qos/workloads` Rest supports filtering on the following fields:

- name
- workload_class
- wid
- policy.name
- svm.name
- volume
- lun
- file
- qtree

You can include these fields under the `filter` parameter. For example, to filter Workload performance instances by `name` where the name contains `NS` or `Test` and `svm` is `vs1`, use the following configuration in RestPerf `workload.yaml` under `counters`:

```yaml
counters:
...
  - filter:
    - name=*NS*|*Test*
    - svm.name=vs1
```

## ONTAP Private CLI

The ONTAP private CLI allows for more granular control and access to non-public counters. It can be used to fill gaps in the REST API, especially in cases where certain data is not yet available through the REST API. Harvest's REST collector can make full use of ONTAP's private CLI. This means when ONTAP's public REST API is missing counters, Harvest can still collect them as long as those counters are available via ONTAP's CLI.

For more information on using the ONTAP private CLI with the REST API, you can refer to the following resources:

- [NetApp Documentation: Accessing ONTAP CLI through REST APIs](https://library.netapp.com/ecmdocs/ECMLP2885799/html/#/Using_the_private_CLI_passthrough_with_the_ONTAP_REST_API)
- [NetApp Blog: Private CLI Passthrough with ONTAP REST API](https://netapp.io/2020/11/09/private-cli-passthrough-ontap-rest-api/)


### Creating Templates That Use ONTAP's Private CLI

Let's take an example of how we can make Harvest use the `system fru-check show` CLI command.

```bash
system fru-check show
```

REST APIs endpoint:

```http
/api/private/cli/system/fru-check?fields=node,fru_name,fru_status
```

Converting the CLI command `system fru-check show` for use with a private CLI REST API can be achieved by adhering to the path rules outlined in the ONTAP [documentation](https://docs.netapp.com/us-en/ontap-restapi/ontap/getting_started_with_the_ontap_rest_api.html#Using_the_private_CLI_passthrough_with_the_ONTAP_REST_API). Generally, this involves substituting all spaces within the CLI command with a forward slash (/), and converting the ONTAP CLI verb into the corresponding REST verb.

The `show` command gets converted to the HTTP method GET call. From the CLI, look at the required field names and pass them as a comma-separated value in `fields=` in the API endpoint.

Note: If the field name contains a hyphen (`-`), it should be converted to an underscore (`_`) in the REST API field. For example, `fru-name` becomes `fru_name`. ONTAP is flexible with the input format and can freely convert between hyphen (`-`) and underscore (`_`) forms. However, when it comes to output, ONTAP returns field names with underscores. For compatibility and consistency, it is mandatory to use underscores in field names when working with Harvest REST templates for ONTAP private CLI.

### Advanced and Diagnostic Mode Commands

The CLI pass through allows you to execute advanced and diagnostic mode CLI commands by including the `privilege_level` field in your request under the `filter` setting like so:
```
counters:
  - filter:
      - privilege_level=diagnostic
```          


### Creating a Harvest Template for Private CLI

Here's a Harvest template that uses ONTAP's private CLI to collect field-replaceable units (FRU) counters by using ONTAP's CLI command `system fru-check show`

```yaml
name:                         FruCheck
query:                        api/private/cli/system/fru-check
object:                       fru_check

counters:
  - ^^node
  - ^^serial_number              => serial_number
  - ^fru_name                    => name
  - ^fru_status                  => status

export_options:
  instance_keys:
    - node
    - serial_number
  instance_labels:
    - name
    - status
```

In this template, the `query` field specifies the private CLI command to be used (`system fru-check show`). The `counters` field maps the output of the private CLI command to the fields of the `fru_check` object.
To identify the ONTAP counter names (the left side of the '=>' symbol in the template, such as `fru_name`), you can establish an SSH connection to your ONTAP cluster. Once connected, leverage ONTAP's command completion functionality to reveal the counter names. For instance, you can type `system fru-check show -fields`, then press the '?' key. This will display a list of ONTAP field names, as demonstrated below.

```
cluster-01::> system fru-check show -fields ?
  node                        Node
  serial-number               FRU Serial Number
  fru-name                    FRU Name
  fru-type                    FRU Type
  fru-status                  Status
  display-name                Display Name
  location                    Location
  additional-info             Additional Info
  reason                      Details
```

The `export_options` field specifies how the data should be exported. The `instance_keys` field lists the fields that will be added as labels to all exported instances of the `fru_check` object. The `instance_labels` field lists the fields that should be included as labels in the exported data.

The output of this template would look like:

```
fru_check_labels{cluster="umeng-aff300-01-02",datacenter="u2",name="DIMM-1",node="umeng-aff300-02",serial_number="s2",status="pass"} 1.0
fru_check_labels{cluster="umeng-aff300-01-02",datacenter="u2",name="PCIe Devices",node="umeng-aff300-02",serial_number="s1",status="pass"} 1.0
```

## Partial Aggregation

For more details about partial aggregation behavior and configuration, see [Partial Aggregation](partial-aggregation.md).