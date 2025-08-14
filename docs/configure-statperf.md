# StatPerf Collector

StatPerf collects performance metrics from ONTAP by invoking the ONTAP CLI statistics command via the private Rest CLI. The full ONTAP CLI command used is:

```bash
statistics show -raw -object $object
```

This collector is designed for performance metrics collection in environments where the ZapiPerf/RestPerf/KeyPerf collectors can not be used.

**Note:** The StatPerf collector requires additional ONTAP permissions. Please refer to the [StatPerf Collector Permissions](prepare-cdot-clusters.md#statperf-least-privilege-role) section for details.
If you are using multi-admin verification in your cluster, you need to allow diagnostic mode queries in the MAV rules for the StatPerf collector to run. For more details, refer to the [Multi-Admin Verification documentation](https://docs.netapp.com/us-en/ontap/multi-admin-verify/).

## Metrics

StatPerf metrics are calculated similar to those in ZapiPerf. For details on how performance metrics are processed and calculated,
please refer to the [ZapiPerf Metrics Documentation](configure-zapi.md#metrics_1).

## Parameters

The parameters for the StatPerf collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- StatPerf configuration file (default: `conf/statperf/default.yaml`)
- Object configuration file (located in `conf/statperf/$version/`)

Except for `addr`, `datacenter` and `auth_style`, all other parameters of the StatPerf collector can be defined in either the Harvest configuration, the main StatPerf configuration file, or the object configuration files. Lower-level file definitions override higher-level ones. This enables configuring objects on an individual basis or applying common settings across all objects.

---

## StatPerf Configuration File

The StatPerf configuration file (also known as the "template") includes a list of objects to collect and their corresponding configuration file names. Additionally, it defines parameters applied as defaults to all objects. As with RestPerf, parameters defined in lower-level files (object or Harvest configuration files) will override the ones provided here.

| Parameter          | Type                           | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |   Default |
|--------------------|--------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------:|
| `use_insecure_tls` | bool, optional                 | Skip verifying the TLS certificate of the target system.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |     false |
| `client_timeout`   | duration (Go-syntax)           | Maximum time to wait for server responses.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |       30s |
| `latency_io_reqd`  | int, optional                  | Threshold of IOPs for calculating latency metrics; latencies based on very few IOPs are unreliable.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |        10 |
| `jitter`           | duration (Go-syntax), optional | Randomly delay collector startup by up to the specified duration to prevent simultaneous REST queries during startup. For example, a jitter value of `1m` will delay startup by a random duration between 0 seconds and 60 seconds. For more details, see [this discussion](https://github.com/NetApp/harvest/discussions/2856).                                                                                                                                                                                                                                                                                    |           |
| `schedule`         | list, required                 | Specifies the polling frequencies, which must include exactly these three elements in the exact specified order:                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |           |
| - `counter`        | duration (Go-syntax)           | Poll frequency for updating the counter metadata cache.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |  24 hours |
| - `instance`       | duration (Go-syntax)           | Poll frequency for updating the instance cache.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | 10 minutes|
| - `data`           | duration (Go-syntax)           | Poll frequency for updating the data cache. Note that while Harvest allows sub-second poll intervals (e.g. `1ms`), factors such as API response times and system load should be considered. In short intervals, performance counters may not be aggregated accurately, potentially leading to a failed state in the collector if the poll interval is less than `client_timeout`. Additionally, very short intervals may cause heavier loads on the ONTAP system and lead to less meaningful metric values (e.g. for latencies).                                                                                    |  1 minute |

The template should list objects in the `objects` section. For example:

```yaml
objects:
   Flexcache:   flexcache.yaml
```

Each object is defined by the filename of its configuration file. The file is located in the subdirectory corresponding to the ONTAP version (e.g., `conf/statperf/$version/`). At runtime, StatPerf selects the object configuration file that best matches the ONTAP system version. In case of mismatches, StatPerf will still fetch and validate counter metadata from the system.

---

## Object Configuration File

The object configuration file allows users to specify object-level parameters for StatPerf. It follows the same concept as other collectors and includes details such as instances, counters, and export options.

For further details, please refer to the guiding documentation below:

- [Object Configuration File](configure-rest.md#restperf-configuration-file)
- [Counters](configure-rest.md#counters_1)
- [Export Options](configure-rest.md#export_options)

#### Template Example

```yaml
name:                     FlexCache
query:                    flexcache_per_volume
object:                   flexcache

counters:
  - ^^instance_name                                     => volume
  - ^^instance_uuid                                     => svm
  - blocks_requested_from_client
  - blocks_retrieved_from_origin
  - evict_rw_cache_skipped_reason_disconnected
  - evict_skipped_reason_config_noent
  - evict_skipped_reason_disconnected
  - evict_skipped_reason_offline
  - invalidate_skipped_reason_config_noent
  - invalidate_skipped_reason_disconnected
  - invalidate_skipped_reason_offline
  - nix_retry_skipped_reason_initiator_retrieve
  - nix_skipped_reason_config_noent
  - nix_skipped_reason_disconnected
  - nix_skipped_reason_in_progress
  - nix_skipped_reason_offline
  - reconciled_data_entries
  - reconciled_lock_entries

plugins:
  - FlexCache
  - MetricAgent:
      compute_metric:
        - miss_percent PERCENT blocks_retrieved_from_origin blocks_requested_from_client

export_options:
  instance_keys:
    - svm
    - volume
```

### Filter

This guide provides instructions on how to use the `filter` feature in StatPerf. Filtering is useful when you need to query a subset of instances. For example, suppose you have a small number of high-value volumes from which you want Harvest to collect performance metrics every five seconds. Collecting data from all volumes at this frequency would be too resource-intensive. Therefore, filtering allows you to create/modify a template that includes only the high-value volumes.

In StatPerf templates, you can set up filters under the `counters` section using the `filter` key as shown below.

To filter `volume` performance instances by instance name, where the name is either `NS_svm_nvme` or contains `Test`, and the `vserver_name` is `osc`, use the following configuration in the `volume.yaml` file under the `counters` section of StatPerf:

```yaml
counters:
  ...
  - filter:
     - instance_name=NS_svm_nvme|instance_name=*Test*
     - vserver_name=osc
```

## Partial Aggregation

For more details about partial aggregation behavior and configuration, see [Partial Aggregation](partial-aggregation.md).