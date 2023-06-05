## EMS collector

The `EMS collector`
collects [ONTAP event management system](https://mysupport.netapp.com/documentation/productlibrary/index.html?productID=62286) (
EMS) events via the ONTAP REST API.

This collector uses a YAML template file to define which events to collect, export, and what labels to attach to each
metric. This means you can collect new EMS events or attach new labels by editing
the [default template](https://github.com/NetApp/harvest/blob/main/conf/ems/default.yaml) file or by [extending existing
templates](configure-templates.md#how-to-extend-a-restrestperfems-collectors-existing-object-template).

The [default template](https://github.com/NetApp/harvest/blob/main/conf/ems/default.yaml) file contains 98 EMS events.

### Supported ONTAP Systems

Any cDOT ONTAP system using 9.6 or higher.

### Requirements

It is recommended to create a read-only user on the ONTAP system.
See [prepare an ONTAP cDOT cluster](prepare-cdot-clusters.md) for details.

### Metrics

This collector collects EMS events from ONTAP and for each received EMS event, creates new metrics prefixed
with `ems_events`.

Harvest supports two types of ONTAP EMS events:

1. **Normal EMS events**

Single shot events. When ONTAP detects a problem, an event is raised.
When the issue is addressed, ONTAP does **not** raise another event reflecting that the problem was resolved.

2. **Bookend EMS events**

ONTAP creates bookend events in matching pairs.
ONTAP creates an event when an issue is detected and another paired event when the event is resolved.
Typically, these events share a common set of properties.

### Collector Configuration

The parameters of the collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- [EMS collector configuration](#ems-collector-configuration-file) file (default: `conf/ems/default.yaml`)
- [EMS template file](#ems-template-file) (located in `conf/ems/9.6.0/ems.yaml`)

Except for `addr`, `datacenter`, and `auth_style`, all other parameters of the EMS collector can be defined
in either of these three files.
Parameters defined in the lower-level files, override parameters in the higher-level file.
This allows you to configure each EMS event individually, or use the same parameters for all events.

#### EMS Collector Configuration File

This configuration file contains the parameters that are used to configure the EMS collector.
These parameters can be defined in your `harvest.yml` or `conf/ems/default.yaml` file.

| parameter        | type           | description                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | default |
|------------------|----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| `client_timeout` | Go duration    | how long to wait for server responses                                                                                                                                                                                                                                                                                                                                                                                                                                         | 1m      |
| `schedule`       | list, required | the polling frequency of the collector/object. Should include exactly the following two elements in the order specified:                                                                                                                                                                                                                                                                                                                                                      |         |
| - `instance`     | Go duration    | polling frequency for updating the instance cache (example value: `24h` = `1440m`)                                                                                                                                                                                                                                                                                                                                                                                            |         |
| - `data`         | Go duration    | polling frequency for updating the data cache (example value: `3m`)<br /><br />**Note** Harvest allows defining poll intervals on sub-second level (e.g. `1ms`), however keep in mind the following:<br /><ul><li>API response of an ONTAP system can take several seconds, so the collector is likely to enter failed state if the poll interval is less than `client_timeout`.</li><li>Small poll intervals will create significant workload on the ONTAP system.</li></ul> |         |

The EMS configuration file should contain the following section mapping the `Ems` object to the corresponding template
file.

```yaml
objects:
  Ems: ems.yaml
```

Even though the EMS mapping shown above references a single file named `ems.yaml`,
there may be multiple versions of that file across subdirectories named after ONTAP releases.
See [cDOT](`https://github.com/NetApp/harvest/tree/main/conf/zapiperf/cdot`) for examples.

At runtime, the EMS collector will select the appropriate object configuration file that most closely matches the
targeted ONTAP system.

#### EMS Template File

The EMS template file should contain the following parameters:

| parameter | type   | description                                                                                        | default                  |
|-----------|--------|----------------------------------------------------------------------------------------------------|--------------------------|
| `name`    | string | display name of the collector. this matches the named defined in your `conf/ems/default.yaml` file | EMS                      |
| `object`  | string | short name of the object, used to prefix metrics                                                   | ems                      |
| `query`   | string | REST API endpoint used to query EMS events                                                         | `api/support/ems/events` |
| `exports` | list   | list of default labels attached to each exported metric                                            |                          |
| `events`  | list   | list of EMS events to collect. See [Event Parameters](#event-parameters)                           |                          |

##### Event Parameters

This section defines the list of EMS events you want to collect, which properties to export, what labels to attach, and
how to handle bookend pairs.
The EMS event template parameters are explained below along with an example for reference.

- `name` is the ONTAP EMS event name. (collect ONTAP EMS events with the name of `LUN.offline`)
- `matches` list of name-value pairs used to further filter ONTAP events.
  Some EMS events include arguments and these name-value pairs provide a way to filter on those arguments.
  (Only collect ONTAP EMS events where `volume_name` has the value `abc_vol`)
- `exports` list of EMS event parameters to export. These exported parameters are attached as labels to each matching
  EMS event.
    - labels that are prefixed with `^^` use that parameter to
      define [instance uniqueness](resources/templates-and-metrics.md#harvest-object-template).
- `resolve_when_ems` (applicable to bookend events only). Lists the resolving event that pairs with the issuing event
    - `name` is the ONTAP EMS event name of the resolving EMS event (`LUN.online`).
      When the resolving event is received, the issuing EMS event will be resolved. In this example, Harvest will raise
      an event when it finds the ONTAP EMS event named `LUN.offline` and that event will be resolved when the EMS event
      named `LUN.online` is received.
    - `resolve_after` (optional, Go duration, default = 28 days) resolve the issuing EMS after the specified duration
      has elapsed (`672h` = `28d`).
      If the bookend pair is not received within the `resolve_after` duration, the Issuing EMS event expires. Harvest
      would marked as auto resolved ems event and add `autoresolve` = `true` label in Issuing EMS event.
    - `resolve_key` (optional) bookend key used to match bookend EMS events. Defaults to prefixed (`^^`) labels
      in `exports` section. `resolve_key` allows you to override what is defined in the `exports` section.

Labels are only exported if they are included in the `exports` section.

Example template definition for the `LUN.offline` EMS event:

```yaml
  - name: LUN.offline
    matches:
      - name: volume_name
        value: abc_vol
    exports:
      - ^^parameters.object_uuid            => object_uuid
      - parameters.object_type              => object_type
      - parameters.lun_path                 => lun_path
      - parameters.volume_name              => volume
      - parameters.volume_dsid              => volume_ds_id
    resolve_when_ems:
      - name: LUN.online
        resolve_after: 672h
        resolve_key:
          - ^^parameters.object_uuid        => object_uuid
```

### How do I find the full list of supported EMS events?

ONTAP documents the list of EMS events created in
the [ONTAP EMS Event Catalog](https://mysupport.netapp.com/documentation/productlibrary/index.html?productID=62286).

You can also query a live system and ask the cluster for its event catalog like so:

```
curl --insecure --user "user:password" 'https://10.61.124.110/api/support/ems/messages?fields=*'
```

Example Output

```
{
  "records": [
    {
      "name": "AccessCache.NearLimits",
      "severity": "alert",
      "description": "This message occurs when the access cache module is near its limits for entries or export rules. Reaching these limits can prevent new clients from being able to mount and perform I/O on the storage system, and can also cause clients to be granted or denied access based on stale cached information.",
      "corrective_action": "Ensure that the number of clients accessing the storage system continues to be below the limits for access cache entries and export rules across those entries. If the set of clients accessing the storage system is constantly changing, consider using the \"vserver export-policy access-cache config modify\" command to reduce the harvest timeout parameter so that cache entries for clients that are no longer accessing the storage system can be evicted sooner.",
      "snmp_trap_type": "severity_based",
      "deprecated": false
    },
...
    {
      "name": "ztl.smap.online.status",
      "severity": "notice",
      "description": "This message occurs when the specified partition on a Software Defined Flash drive could not be onlined due to internal S/W or device error.",
      "corrective_action": "NONE",
      "snmp_trap_type": "severity_based",
      "deprecated": false
    }
  ],
  "num_records": 7273
}
```

## Ems Prometheus Alerts

Refer [Prometheus-Alerts](prometheus-exporter.md#prometheus-alerts)


