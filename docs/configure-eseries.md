!!! note "Beta Feature"
    The EseriesMel collector is new in Harvest and should be considered beta. 
    Feedback and bug reports are welcome on [GitHub Discussions](https://github.com/NetApp/harvest/discussions).

The Eseries collectors use the REST protocol to collect data from NetApp E-Series storage systems.

The [EseriesPerf collector](#eseriesperf-collector) is an extension of this collector for performance metrics, therefore they share many parameters and configuration settings.

The [EseriesMel collector](#eseriesmel-collector) collects Major Event Log (MEL) events.

### Requirements

- E-Series storage array with REST API support
- A user account with **Monitor role** permissions on the E-Series array (see [prepare-eseries.md](prepare-eseries.md#monitor-role-requirements))

No SDK or other requirements.

### Metrics

The Eseries collector collects a dynamic set of metrics from E-Series storage arrays. The E-Series REST API returns JSON documents and Harvest extracts values from the JSON using template definitions with dot notation paths.

The collector automatically discovers the storage array and extracts metrics for volumes, controllers, hardware components, and other objects.


### Parameters

The parameters of the collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- Eseries configuration file (default: `conf/eseries/default.yaml`)
- Each object has its own configuration file (located in `conf/eseries/$version/`)

Except for `addr` and `datacenter`, all other parameters of the Eseries collector can be defined in either of these three files. Parameters defined in the lower-level file override parameters in the higher-level ones. This allows you to configure each object individually, or use the same parameters for all objects.

The full set of parameters are described [below](#collector-configuration-file).

#### Harvest Configuration Example

In your `harvest.yml`, configure a poller pointing to your E-Series storage array:

```yaml
Pollers:
  eseries-array:
    datacenter: DC-01
    addr: 10.0.1.100                    # E-Series array management address
    username: monitor                   # Array user with Monitor role
    password: enterpass                 # Or use credential_script
    collectors:
      - Eseries
      - EseriesPerf
    exporters:
      - prometheus
```

#### Collector Configuration File

This configuration file contains a list of objects that should be collected and the filenames of their templates. Additionally, this file contains the parameters that are applied as defaults to all objects.

| parameter        | type                           | description                                                                                                                                                                                                                                                                                                                                                                                                                                                | default   |
|------------------|--------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------|
| `client_timeout` | duration (Go-syntax)           | how long to wait for server responses                                                                                                                                                                                                                                                                                                                                                                                                                      | 30s       |
| `jitter`         | duration (Go-syntax), optional | Each Harvest collector runs independently, which means that at startup, each collector may send its REST queries at nearly the same time. To spread out the collector startup times over a broader period, you can use `jitter` to randomly distribute collector startup across a specified duration. For example, a `jitter` of `1m` starts each collector after a random delay between 0 and 60 seconds. For more details, refer to [this discussion](https://github.com/NetApp/harvest/discussions/2856). |           |
| `schedule`       | list, **required**             | how frequently to retrieve metrics from the E-Series array                                                                                                                                                                                                                                                                                                                                                                                                 |           |
| - `counter`      | duration (Go-syntax)           | poll frequency of updating the counter metadata cache                                                                                                                                                                                                                                                                                                                                                                                                      | 24 hours  |
| - `data`         | duration (Go-syntax)           | how frequently this collector/object should retrieve metrics                                                                                                                                                                                                                                                                                                                                                                                               | 3 minutes |

The default configuration file (`conf/eseries/default.yaml`) defines the objects to collect:

```yaml
collector: Eseries

schedule:
  - counter: 24h
  - data: 3m

objects:
  Volume:            volume.yaml
  Array:             array.yaml
  Host:              host.yaml
  Controller:        controller.yaml
```

For each object, we define the filename of the object configuration file.
The object configuration files are located in subdirectories matching the SANtricity OS version (e.g., `11.80.0`).
At runtime, the collector will select the object configuration file that closest matches the version of the target E-Series system.

### Object Configuration File

The Object configuration file ("subtemplate") should contain the following parameters:

| parameter              | type                        | description                                                                                                                                                  | default |
|------------------------|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| `name`                 | string, **required**        | display name of the collector that will collect this object                                                                                                  |         |
| `object`               | string, **required**        | short name of the object, used to prefix metrics (e.g., `eseries_volume`)                                                                                   |         |
| `query`                | string, **required**        | REST API endpoint to query, relative to `/devmgr/v2/` (can include `{array_id}` placeholder)                                                                |         |
| `counters`             | list, **required**          | list of counters to collect (see [Counters](#counters) section below)                                                                                       |         |
| `plugins`              | list, optional              | list of plugins to run on the collected data (see [Plugins](#plugins) section below)                                                                        |         |
| `export_options`       | section, **required**       | defines how to export instance labels and keys                                                                                                               |         |

### Template Example

Here's an example of the Volume object template (`conf/eseries/11.80.0/volume.yaml`):

```yaml
name: Volume
query: storage-systems/{array_id}/volumes
object: eseries_volume

counters:
  - ^^name                                => volume
  - ^listOfMappings                       => list_of_mappings
  - ^metadata                             => metadata
  - ^offline                              => offline
  - ^raidLevel                            => raid_level
  - ^status                               => status
  - ^volumeGroupRef                       => volume_group_ref
  - ^wwn                                  => wwid
  - blkSize                               => block_size
  - capacity                              => reported_capacity
  - totalSizeInBytes                      => allocated_capacity

plugins:
  - Volume
  - VolumeMapping

export_options:
  instance_keys:
    - volume
  instance_labels:
    - hosts
    - luns
    - mapping_types
    - offline
    - pool
    - raid_level
    - status
    - volume
    - workload
    - wwid
```

#### Query Path and Array ID Injection

The `query` parameter can include `{array_id}` as a placeholder. The collector automatically discovers the storage array ID and injects it into the query URL at runtime.

For example:
- Template: `storage-systems/{array_id}/volumes`
- Runtime: `storage-systems/1/volumes`

### Counters

Counters define which metrics and labels to collect from the REST API response. Each counter line follows this format:

```
[prefix]json_field_name => display_name
```

**Counter Prefixes:**

- `^^` - **Instance key**: Uniquely identifies each instance (required, must have at least one)
- `^` - **Instance label**: String metadata exported to `<object>_labels` metric
- No prefix - **Numeric metric**: Exported as its own time-series metric

**Arrow Syntax (`=>`):**

The arrow renames the JSON field to a shorter display name used in metrics:

```yaml
counters:
  - ^^name                    => volume           # Instance key
  - ^raidLevel                => raid_level       # Label
  - totalSizeInBytes          => allocated_capacity  # Metric
```

This produces metrics like:
- `eseries_volume_allocated_capacity{volume="MyVolume", ...}`
- `eseries_volume_labels{volume="MyVolume", raid_level="raid6", ...}`

### Export Options

The `export_options` section controls what gets exported:

```yaml
export_options:
  instance_keys:      # Primary identifier(s)
    - volume          
  instance_labels:    # Labels to include in <object>_labels metric
    - hosts
    - luns
    - mapping_types
    - offline
    - pool
    - raid_level
    - status
    - volume
    - workload
    - wwid    
```

- **instance_keys**: Unique identifier labels (from `^^` counters)
- **instance_labels**: All labels to export in the `<object>_labels` metric

### Plugins

Eseries collectors support plugins that enrich or transform collected data:

**Plugin Configuration:**

```yaml
plugins:
  - Volume
  - VolumeMapping
```

Plugins run after data collection but before export, allowing them to add computed metrics or enrich labels.

---

## EseriesPerf Collector

The EseriesPerf collector extends Eseries to collect **performance metrics** from E-Series arrays using the `/live-statistics` endpoint. It implements delta calculations similar to RestPerf and ZapiPerf collectors.

### Static Counter Definitions

Unlike Eseries, EseriesPerf uses a **static counter definitions file** (`conf/eseriesperf/static_counter_definitions.yaml`) that defines how to process each counter:

```yaml
objects:
  eseries_volume:
    counter_definitions:
      - name: readOps
        type: rate                    # Calculate delta per second
      - name: readTimeTotal
        type: average                 # Divide by base counter
        base_counter: readOps         # readTimeTotal / readOps = avg latency
      - name: readBytes
        type: rate
      - name: writeOps
        type: rate
```

### Performance Template Example

Here's an example EseriesPerf template (`conf/eseriesperf/11.80.0/volume.yaml`):

```yaml
name: Volume
query: storage-systems/{array_id}/live-statistics
object: eseries_volume
type: volume

counters:
  - ^^volumeName                  => volume
  - lastResetTimeInMS             => last_reset_time
  - observedTimeInMS              => observed_time
  - readBytes                     => read_data
  - readHitOps                    => read_hit_ops
  - readOps                       => read_ops
  - readTimeTotal                 => read_latency
  - writeBytes                    => write_data
  - writeHitOps                   => write_hit_ops
  - writeOps                      => write_ops
  - writeTimeTotal                => write_latency

plugins:
  - CacheHitRatio

export_options:
  instance_keys:
    - volume
```

The static counter definitions file determines how each counter is processed:

```yaml
# conf/eseriesperf/static_counter_definitions.yaml
objects:
  eseries_volume:
    counter_definitions:
      - name: readOps
        type: rate                    # Becomes read_ops (ops/sec)
      - name: readTimeTotal
        type: average                 # Becomes read_latency (ms/op)
        base_counter: readOps
      - name: readBytes
        type: rate                    # Becomes read_data (bytes/sec)
```

---

## EseriesMel Collector

The EseriesMel collector uses the REST protocol to collect Major Event Log (MEL) events from NetApp E-Series storage systems, via the `storage-systems/{array_id}/mel-events` endpoint. Events that occur while the collector is not running are not captured.

### EseriesMel Requirements

- E-Series storage array with REST API support
- A user account with **Monitor role** permissions on the E-Series array (see [prepare-eseries.md](prepare-eseries.md#monitor-role-requirements))

No SDK or other requirements.

### EseriesMel Metrics

The EseriesMel collector exports a single metric, `eseries_mel_events`. Its value is the epoch timestamp of the event's most recent occurrence.

- Metric labels: `array`, `chassis_serial`, `component_type`, `datacenter`, `event_type`, `location`, `management_version`, `message`, `severity`
- Repeated occurrences of the same `(event_type, location)` pair collapse into a single series carrying the latest timestamp.

Example:

```
eseries_mel_events{datacenter="rtp",management_version="11.90",array="VolumeBlockSizes1",chassis_serial="1111111111",component_type="drive",event_type="8812",location="Tray 1, Slot 5",message="MEL_EV_CFG_DRIVE_FAILURE",severity="critical"} 1788274734
```

See the [eseries_mel_events](eseries-metrics.md#eseries_mel_events) entry in the metrics reference for details.

#### EseriesMel Event Collection

Each poll, the collector first asks `mel-events/available` for the array's current sequence-number range, then reads events with a `sequenceNumber` cursor:

- On the first poll, the cursor starts at the tip of the range, so events that occurred before the collector started are not replayed.
- Each subsequent poll reads only the events between the last cursor position and the new tip.
- A series is exported only for the poll in which its event is read. Once the cursor moves past it, the event is not retained on later polls.
- If the array purges events before the collector reads them, the collector logs a warning and realigns the cursor to the new range start; the missed events cannot be recovered.
- If the MEL sequence counter rolls back (for example after a controller replacement or log reset), the collector logs a warning and resets the cursor to the start of the new range.

### EseriesMel Parameters

The parameters of the collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- EseriesMel configuration file (default: `conf/eseriesmel/default.yaml`)
- Each object has its own configuration file (located in `conf/eseriesmel/$version/`)

Except for `addr` and `datacenter`, all other parameters of the EseriesMel collector can be defined in either of these three files. Parameters defined in the lower-level file override parameters in the higher-level ones. This allows you to configure each object individually, or use the same parameters for all objects.

The full set of parameters are described [below](#eseriesmel-collector-configuration-file).

#### EseriesMel Harvest Configuration Example

In your `harvest.yml`, configure a poller pointing to your E-Series storage array:

```yaml
Pollers:
  eseries-array:
    datacenter: DC-01
    addr: 10.0.1.100                    # E-Series array management address
    username: monitor                   # Array user with Monitor role
    password: enterpass                 # Or use credential_script
    collectors:
      - Eseries
      - EseriesPerf
      - EseriesMel
    exporters:
      - prometheus
```

#### EseriesMel Collector Configuration File

This configuration file contains a list of objects that should be collected and the filenames of their templates. Additionally, this file contains the parameters that are applied as defaults to all objects.

| parameter        | type                           | description                                                                                                                                                                                                                                                                                                                                                                                                                                                | default   |
|------------------|--------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------|
| `client_timeout` | duration (Go-syntax)           | how long to wait for server responses                                                                                                                                                                                                                                                                                                                                                                                                                      | 30s       |
| `jitter`         | duration (Go-syntax), optional | Each Harvest collector runs independently, which means that at startup, each collector may send its REST queries at nearly the same time. To spread out the collector startup times over a broader period, you can use `jitter` to randomly distribute collector startup across a specified duration. For example, a `jitter` of `1m` starts each collector after a random delay between 0 and 60 seconds. For more details, refer to [this discussion](https://github.com/NetApp/harvest/discussions/2856). |           |
| `schedule`       | list, **required**             | how frequently to retrieve events from the E-Series array                                                                                                                                                                                                                                                                                                                                                                                                  |           |
| - `counter`      | duration (Go-syntax)           | poll frequency of refreshing the array identity and `management_version` global label                                                                                                                                                                                                                                                                                                                                                                     | 24 hours  |
| - `data`         | duration (Go-syntax)           | how frequently this collector/object should retrieve events                                                                                                                                                                                                                                                                                                                                                                                                | 3 minutes |

The default configuration file (`conf/eseriesmel/default.yaml`) defines the objects to collect:

```yaml
collector: EseriesMel

schedule:
  - counter: 24h
  - data: 3m

objects:
  Events: events.yaml
```

For each object, we define the filename of the object configuration file.
The object configuration files are located in subdirectories matching the SANtricity OS version (e.g., `11.80.0`).
At runtime, the collector will select the object configuration file that closest matches the version of the target E-Series system.

### EseriesMel Object Configuration File

The Object configuration file ("subtemplate") should contain the following parameters:

| parameter        | type                   | description                                                                                                    | default |
|------------------|------------------------|------------------------------------------------------------------------------------------------------------------------------------------|---------|
| `name`           | string, **required**   | display name of the collector that will collect this object                                                   |         |
| `object`         | string, **required**   | short name of the object, used to prefix metrics (e.g., `eseries_mel`)                                         |         |
| `query`          | string, **required**   | REST API endpoint to query, relative to `/devmgr/v2/` (can include `{array_id}` placeholder)                   |         |
| `batch_size`     | int, optional          | maximum number of events requested per page; must be a positive integer                                        | 500     |
| `filter`         | list, optional         | query fragments appended verbatim to each event-request fetch                                                              |         |
| `events`         | list, **required**     | allow-list of `event_type`/`name` pairs the collector exports (see [Event Catalog](#eseriesmel-event-catalog) below) |         |
| `counters`       | list, **required**     | list of counters to collect (see [Counters](#eseriesmel-counters) section below)                               |         |
| `export_options` | section, **required**  | defines how to export instance labels and keys                                                                 |         |

### EseriesMel Template Example

Here's the shipped Events object template (`conf/eseriesmel/11.80.0/events.yaml`), with most of its 60-entry event catalog elided:

```yaml
name: Events
query: storage-systems/{array_id}/mel-events
object: eseries_mel

# includeDebug=false already matches the API's own default; kept explicit so
# behavior doesn't silently change if the array's default ever does.
filter:
  - includeDebug=false

events:
  # AVAILABILITY
  - event_type: 5123
    name: MEL_EV_EXCESSIVE_REBOOTS_DETECTED
  - event_type: 8781
    name: MEL_EV_DRV_NO_RESPONSE
  # ... 58 more, grouped into AVAILABILITY, CAPACITY, CONFIGURATION,
  # PERFORMANCE, and PROTECTION

counters:
  - ^^eventType       => event_type
  - ^^location        => location
  - ^componentType    => component_type
  - ^priority         => severity

export_options:
  instance_keys:
    - component_type
    - event_type
    - location
    - message
    - severity
```

#### EseriesMel Event Catalog

`events` is a required allow-list: only event types listed here are collected, so adding coverage for a new event means adding an entry. Each entry's `name` becomes the exported `message` label. The shipped catalog groups 60 events by comment into Availability, Capacity, Configuration, Performance, and Protection.

To see the full shipped catalog, look at [conf/eseriesmel/11.80.0/events.yaml](https://github.com/NetApp/harvest/blob/main/conf/eseriesmel/11.80.0/events.yaml). To find event types available on a live array, first read the sequence-number range, then fetch a page from the range start (Harvest itself starts at the range tip and does not replay history):

```
curl --insecure --user "user:password" \
  'https://10.0.1.100:8443/devmgr/v2/storage-systems/1/mel-events/available'

curl --insecure --user "user:password" \
  'https://10.0.1.100:8443/devmgr/v2/storage-systems/1/mel-events?startSequenceNumber=0&count=100'
```

### EseriesMel Counters

Counters define which metrics and labels to collect from the REST API response. Each counter line follows this format:

```
[prefix]json_field_name => display_name
```

**Counter Prefixes:**

- `^^` - **Instance key**: Uniquely identifies each instance (required, must have at least one)
- `^` - **Instance label**: String metadata attached to `eseries_mel_events` (there is no `eseries_mel_labels` metric)

Unlike other collectors, EseriesMel has no analogue of a bare, unprefixed numeric counter: it exports a single `eseries_mel_events` metric that the collector sets programmatically, so every counter line must use `^^` or `^`. The shipped template lists those labels under `export_options.instance_keys`, so they appear on `eseries_mel_events` itself.

**Arrow Syntax (`=>`):**

The arrow renames the JSON field to a shorter display name used in metrics:

```yaml
counters:
  - ^^eventType       => event_type       # Instance key
  - ^^location        => location         # Instance key
  - ^componentType    => component_type   # Label
  - ^priority         => severity         # Label
```

#### EseriesMel Label Normalization

Some exported label values differ from the raw API response:

- `severity` maps known `priority*` API values to a short form (e.g. `priorityCritical` becomes `critical`). Unrecognized values are exported as-is.
- `component_type` resolves the API's `relative` placeholder through `componentLocation.componentRelativeLocation.componentType`.
- The API's `__UNDEFINED` sentinel, and an unresolvable `relative`, both export as `unknown`.

### EseriesMel Export Options

The `export_options` section controls what gets exported:

```yaml
export_options:
  instance_keys:
    - component_type
    - event_type
    - location
    - message
    - severity
```

- **instance_keys**: Labels included on every exported series. `message` is set by the collector from the [event catalog](#eseriesmel-event-catalog) entry rather than coming from a counter.

See [prepare-eseries.md](prepare-eseries.md) for system setup and [troubleshooting guide](help/troubleshooting.md) for general issues.
