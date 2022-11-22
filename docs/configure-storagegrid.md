## StorageGRID Collector

The StorageGRID collector uses REST calls to collect data from StorageGRID systems.

### Target System

All StorageGRID versions are supported, however the default configuration files may not completely match with older
systems.

### Requirements

No SDK or other requirements. It is recommended to create a read-only user for Harvest on the StorageGRID system (see
[prepare monitored clusters](prepare-storagegrid-clusters.md) for details)

### Metrics

The collector collects a dynamic set of metrics via StorageGRID's REST API. StorageGRID returns JSON documents and
Harvest allows you to define templates to extract values from the JSON document via a dot notation path. You can view
StorageGRID's full set of REST APIs by visiting `https://$STORAGE_GRID_HOSTNAME/grid/apidocs.html`

As an example, the `/grid/accounts-cache` endpoint, lists the tenant accounts in the cache and includes additional
information, such as objectCount and dataBytes. Below is an example response from this endpoint:

```json
{
  "data": [
    {
      "id": "95245224059574669217",
      "name": "foople",
      "policy": {
        "quotaObjectBytes": 50000000000
      },
      "objectCount": 6,
      "dataBytes": 10473454261
    }
  ]
}
```

The StorageGRID collector will take this document, extract the `data` section and convert the metrics above
into: `name`, `policy.quotaObjectBytes`, `objectCount`, and `dataBytes`. Metric names will be taken, as is, unless you
specify a short display name. See [counters](configure-templates.md#counters) for more details.

## Parameters

The parameters of the collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- StorageGRID configuration file (default: `conf/storagegrid/default.yaml`)
- Each object has its own configuration file (located in `conf/storagegrid/$version/`)

Except for `addr` and `datacenter`, all other parameters of the StorageGRID collector can be
defined in either of these three files. Parameters defined in the lower-level file, override parameters in the
higher-level ones. This allows you to configure each object individually, or use the same parameters for all
objects.

The full set of parameters are described [below](#harvest-configuration-file).

### Harvest configuration file

Parameters in the poller section should define the following required parameters.

| parameter              | type                 | description                                                                    | default |
|------------------------|----------------------|--------------------------------------------------------------------------------|---------|
| Poller name (header)   | string, **required** | Poller name, user-defined value                                                |         |
| `addr`                 | string, **required** | address (IP or FQDN) of the ONTAP system                                       |         |
| `datacenter`           | string, **required** | Datacenter name, user-defined value                                            |         |
| `username`, `password` | string, **required** | StorageGRID username and password with at least `Tenant accounts` permissions  |         |
| `collectors`           | list, **required**   | Name of collector to run for this poller, use `StorageGrid` for this collector |         |

### StorageGRID configuration file

This configuration file contains a list of objects that should be collected and the filenames of their templates (
explained in the next section).

Additionally, this file contains the parameters that are applied as defaults to all objects. As mentioned before, any
of these parameters can be defined in the Harvest or object configuration files as well.

| parameter        | type                 | description                                                                   | default   |
|------------------|----------------------|-------------------------------------------------------------------------------|-----------|
| `client_timeout` | duration (Go-syntax) | how long to wait for server responses                                         | 30s       |
| `schedule`       | list, **required**   | how frequently to retrieve metrics from StorageGRID                           |           |
| - `data`         | duration (Go-syntax) | how frequently this collector/object should retrieve metrics from StorageGRID | 5 minutes |

The template should define objects in the `objects` section. Example:

```yaml
objects:
  Tenant: tenant.yaml
```

For each object, we define the filename of the object configuration file. The object configuration files
are located in subdirectories matching the StorageGRID version that was used to create these files. It is possible to
have multiple version-subdirectories for multiple StorageGRID versions. At runtime, the collector will select the object
configuration file that closest matches the version of the target StorageGRID system.

### Object configuration file

The Object configuration file ("subtemplate") should contain the following parameters:

| parameter        | type                 | description                                                                        | default |
|------------------|----------------------|------------------------------------------------------------------------------------|---------|
| `name`           | string, **required** | display name of the collector that will collect this object                        |         |
| `query`          | string, **required** | REST endpoint used to issue a REST request                                         |         |
| `object`         | string, **required** | short name of the object                                                           |         |
| `api`            | string               | StorageGRID REST endpoint version to use, overrides default management API version | 3       |
| `counters`       | list                 | list of counters to collect (see notes below)                                      |         |
| `plugins`        | list                 | plugins and their parameters to run on the collected data                          |         |
| `export_options` | list                 | parameters to pass to exporters (see notes below)                                  |         |

#### `counters`

This section defines the list of counters that will be collected. These counters can be labels, numeric metrics or
histograms. The exact property of each counter is fetched from StorageGRID and updated periodically.

The display name of a counter can be changed with `=>` (e.g., `policy.quotaObjectBytes => logical_quota`).

Counters that are stored as labels will only be exported if they are included in the `export_options` section.

#### `export_options`

Parameters in this section tell the exporters how to handle the collected data. The set of parameters varies by
exporter. For [Prometheus](prometheus-exporter.md) and [InfluxDB](influxdb-exporter.md)
exporters, the following parameters can be defined:

* `instances_keys` (list): display names of labels to export with each data-point
* `instance_labels` (list): display names of labels to export as a separate `_label` metric
* `include_all_labels` (bool): export all labels with each data-point (overrides previous two parameters)
