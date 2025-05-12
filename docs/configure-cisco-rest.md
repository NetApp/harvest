## CiscoRest Collector

The CiscoRest collector uses NX-API REST calls to collect data from Cisco switches.

### Target System

Harvest supports all Cisco switches listed in [NetApp's Hardware Universe](https://hwu.netapp.com/).

### Requirements

The NX-API feature must be enabled on the switch. No SDK or other requirements. It is recommended to create a read-only user for Harvest on the Cisco switch (see
[prepare monitored clusters](prepare-cisco-switch.md) for details)

### Metrics

The collector collects a dynamic set of metrics via Cisco's NX-API. The switch returns JSON documents, and unlike other
Harvest collectors, the CiscoRest collector does not provide template customization.

## Parameters

The parameters of the collector are distributed across three files:

- [Harvest configuration file](configure-harvest-basic.md#pollers) (default: `harvest.yml`)
- CiscoRest configuration file (default: `conf/ciscorest/default.yaml`)
- Each object has its own configuration file (located in `conf/ciscorest/nxos/$version/`)

Except for `addr` and `datacenter`, all other parameters of the CiscoRest collector can be defined in any of these three files. Parameters defined in a lower-level file override those in higher-level files. This allows you to configure each object individually or use the same parameters for all objects.

The full set of parameters are described [below](#harvest-configuration-file).

### Harvest configuration file

Parameters in the poller section should define the following required parameters.

| parameter              | type                 | description                                                                    | default |
|------------------------|----------------------|--------------------------------------------------------------------------------|---------|
| Poller name (header)   | string, **required** | Poller name, user-defined value                                                |         |
| `addr`                 | string, **required** | IPv4, IPv6 or FQDN of the target system                                        |         |
| `datacenter`           | string, **required** | Datacenter name, user-defined value                                            |         |
| `username`, `password` | string, **required** | Cisco swicj username and password with at least `network-operator` permissions |         |
| `collectors`           | list, **required**   | Name of collector to run for this poller, use `CiscoRest` for this collector   |         |

### CiscoRest configuration file

This configuration file contains a list of objects that should be collected and the filenames of their templates (explained in the next section).

Additionally, this file contains the parameters that are applied as defaults to all objects. As mentioned before, any
of these parameters can be defined in the Harvest or object configuration files as well.

| parameter               | type                 | description                                                                   | default   |
|-------------------------|----------------------|-------------------------------------------------------------------------------|-----------|
| `client_timeout`        | duration (Go-syntax) | how long to wait for server responses                                         | 30s       |
| `schedule`              | list, **required**   | how frequently to retrieve metrics from StorageGRID                           |           |
| - `data`                | duration (Go-syntax) | how frequently this collector/object should retrieve metrics from StorageGRID | 5 minutes |
| `only_cluster_instance` | bool, optional       | don't require instance key. assume the only instance is the cluster itself    |           |

The template should define objects in the `objects` section. Example:

```yaml
objects:
  Optic: optic.yaml
```

For each object, we define the filename of the object configuration file. The object configuration files
are located in subdirectories matching the CiscoRest version that was used to create these files. It is possible to
have multiple version-subdirectories for multiple CiscoRest versions. At runtime, the collector will select the object
configuration file that closest matches the version of the target CiscoRest system.

### Object configuration file

The Object configuration file ("subtemplate") should contain the following parameters:

| parameter        | type                 | description                                                                        | default |
|------------------|----------------------|------------------------------------------------------------------------------------|---------|
| `name`           | string, **required** | display name of the collector that will collect this object                        |         |
| `query`          | string, **required** | Cisco switch CLI command used to issue a REST request                              |         |
| `object`         | string, **required** | short name of the object                                                           |         |
| `plugins`        | list                 | plugins and their parameters to run on the collected data                          |         |

