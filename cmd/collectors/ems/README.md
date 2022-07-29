

# EMS collector

EMS collector collects ems events from ONTAP systems using the REST protocol. The collector is designed to be easily extendible to collect new ems events or to collect additional labels from already configured ems events. (The [default template](../../../conf/ems/9.6.0/ems.yaml) file contains 60 ems events).

## Target System
Target system can be any 9.6+ cDot ONTAP system. Any higher version than 9.6 would be supported.

## Requirements
No SDK or any other requirement. It is recommended to create a read-only user for Harvest on the ONTAP system (see the [Authentication document](../../../docs/AuthAndPermissions.md))

## Metrics

The EMS collector collects ems events from ONTAP and create new metric `ems_events` for each of the received ems events.

Harvest supports handling of 2 types of ems events:
1. **Non-bookend ems events**  :
Whenever something goes wrong in ONTAP, ems event would be raised. If the problem would be fixed, no changes reflected in this ems event.  

2. **Bookend ems events:**. 
They are pair of 2 ems events [`Issuing ems event`, `Resolving ems event`].
Whenever something goes wrong in ONTAP, Issuing ems event would be raised. If the problem would be fixed, Resolving ems event would be raised.

## Parameters

The parameters of the collector are distributed across three files:
- Harvest configuration file (default: `harvest.yml`)
- Ems configuration file (default: `conf/ems/default.yaml`)
- Ems template file (located in `conf/ems/9.6.0/ems.yaml`)

With the exception of `addr`, `datacenter` and `auth_style`, all other parameters of the Ems collector can be defined in either of these three files. Parameters defined in the lower-level file, override parameters in the higher-level file. This allows the user to configure each ems event individually, or use the same parameters for all ems.

For the sake of brevity, these parameters are described only in the section [Collector configuration file](#collector-configuration-file).


### Harvest configuration file

Parameters in poller section should define (at least) the address and authentication method of the target system:

| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `addr`             | string, required | address (IP or FQDN) of the ONTAP system         |                        |
| `datacenter`       | string, required | name of the datacenter where the target system is located |               |
| `auth_style` | string, optional | authentication method: either `basic_auth` or `certificate_auth` | `basic_auth` |
| `ssl_cert`, `ssl_key` | string, optional | full path of the SSL certificate and key pairs (when using `certificate_auth`) | |
| `username`, `password` | string, optional | full path of the SSL certificate and key pairs (when using `basic_auth`) | |

It is recommended creating a read-only user on the ONTAP system dedicated to Harvest. See section [Authentication](#authentication) for guidance.

### Collector configuration file

This configuration file contains the parameters that are applied as defaults to all ems. (As mentioned before, any of these parameters can be defined in the Harvest or object configuration files as well).

| parameter              | type        | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `client_timeout`   | duration (Go-syntax)  | how long to wait for server responses             | 1m                  |
| `schedule`         | list, required | the poll frequencies of the collector/object, should include exactly these two elements in the exact same order: | |
|    - `instance`         | duration (Go-syntax) | poll frequency of updating the instance cache (example value: `24h` = `1440m`) | |
|    - `data`         | duration (Go-syntax) | poll frequency of updating the data cache (example value: `3m`)<br /><br />**Note** Harvest allows defining poll intervals on sub-second level (e.g. `1ms`), however keep in mind the following:<br /><ul><li>API response of an ONTAP system can take several seconds, so the collector is likely to enter failed state if the poll interval is less than `client_timeout`.</li><li>Small poll intervals will create significant workload on the ONTAP system.</li></ul> | |

The configuration file should define Ems as object in the `objects` section. Example:

```yaml
objects:
  Ems:             ems.yaml
```

Note that for each object we only define the filename of the object configuration file. The object configuration files are located in subdirectories matching to the ONTAP version that was used to create these files. It is possible to have multiple version-subdirectories for multiple ONTAP versions. At runtime, the collector will select the object configuration file that closest matches to the version of the target ONTAP system.

### Object configuration file

The Object configuration file ("template") should contain the following parameters:

| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `name`                 | string       | display name of the collector that will collect this object |             |
| `object`               | string       | short name of the object                         |                        |
| `query`                | string       | api endpoint used for a REST request     |                        |
| `exports`              | list         | list of default labels which gets appended to individual event   |                        |
| `events`               | list         | list of ems events to collect (see notes below) |                        |

#### `events`

This section defines the list of ems events which will be collected. The exact name of ems with requested labels is fetched from ONTAP and later requested match filter would be applied.

The display name of a label can be changed with `=>` (e.g., `filePath => file_path`). 

Labels will only be exported if they are included in the `exports` section.

As an example, the `LUN.offline` ems event would provide the following attribute tree:

```yaml
  - name: LUN.offline
    matches:
      - name: volume_name
        value: volume_name_1
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
**Note**: values mentioned in the below description are for reference purpose only.

* `name`: name of ems event (Issuing ems in case of Bookend ems) to collect (`LUN.offline`)
* `matches` (list): name-value pair to filter out this ems event's data-point. Only collect ems event where `volume_name` is `volume_name_1`.
    * `name`: name of label (`volume_name`)
    * `value`: value of the above label (`volume_name_1`)
* `exports` (list): export all mentioned labels with each data-point 
  * label starts with `^^` (`object_uuid`) would be used as bookend key in case of Bookend ems events
* `resolve_when_ems*`: resolving ems detail which would resolve the above issuing ems
  * `name`: name of the resolving ems event to collect (`LUN.online`). As soon as receiving the resolving ems, issuing ems would be resolved.
  * `resolve_after`: issuing ems would be expected to resolve by resolving ems within this duration (`672h` = `28d`), If not then Harvest removes issuing ems from cache. [`optional`] (default: 672h)
  * `resolve_key` (list): would be used as bookend key in Bookend ems events [`optional`] (default: taken from `exports` section where label starts with `^^`)


**Note**:

`*` indicates it's only applicable for Bookend ems events
