# ZapiPerf

ZapiPerf collects performance metrics from ONTAP systems using the ZAPI protocol. The collector is designed to be easily extendible to collect new objects or to collect additional counters from already configured objects. (The [default configuration](../../../conf/zapiperf/default.yaml) file contains 25 objects)

This collector is an extension of the [Zapi collector](../zapi/README.md) with the major difference between that ZapiPerf collects only the `perf` subfamily of the ZAPIs. Additionally, ZapiPerf always calculates final values from deltas of two subsequent polls.

## Target System
Target system can be any cDot or 7Mode ONTAP system. Any version is supported, however the default configuration files may not completely match with an older system.

## Requirements
No SDK or any other requirement. It is recommended to create a read-only user for Harvest on the ONTAP system (see the [Authentication document](../../../docs/AuthAndPermissions.md))

## Parameters


The parameters of the collector are distributed across three files:
- Harvest configuration file (default: `harvest.yml`)
- ZapiPerf configuration file (default: `conf/zapiperf/default.yaml`)
- Each object has its own configuration file (located in `conf/zapiperf/cdot/` and `conf/zapiperf/7mode/` for cDot and 7Mode systems respectively)

With the exception of `addr`, `datacenter` and `auth_style`, all other parameters of the ZapiPerf collector can be defined in either of these three files. Parameters defined in the lower-level file, override parameters in the higher-level file. This allows the user to configure each objects individually, or use same parameters for all objects.

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

We can define the configuration file of the collector. If no configuration file is specified, the default configuration file (`conf/zapiperf/default.yaml`) will be used and if the file `conf/zapiperf/default.yaml` is present, it will be merged to the default one. If we specify our own configuration file for the collector, it can have any name, and it will not be merged.

Examples:

Define a poller that will run the ZapiPerf collector using its default configuration file:

```yaml
Pollers:
  jamaica:  # name of the poller
    datacenter: munich
    addr: 10.65.55.2
    auth_style: basic_auth
    username: harvest
    password: 3t4ERTW%$W%c
    collectors:
      - ZapiPerf # will use conf/zapiperf/default.yaml and optionally merge with conf/zapiperf/custom.yaml
```

Define a poller that will run the ZapiPerf collector using a custom configuration file:

```yaml
Pollers:
  jamaica:  # name of the poller
    addr: 10.65.55.2
    auth_style: basic_auth
    username: harvest
    password: 3t4ERTW%$W%c
    collectors:
      - ZapiPerf:
        - limited.yaml # will use conf/zapiperf/limited.yaml
        # if more templates are added, they will be merged
```

### Collector configuration file

This configuration file (the "template") contains a list of objects that should be collected and the filenames of their configuration (explained in the next section).

Additionally, this file contains the parameters that are applied as defaults to all objects. (As mentioned before, any of these parameters can be defined in the Harvest or object configuration files as well).

| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `use_insecure_tls` | bool, optional | skip verifying TLS certificate of the target system | `false`               |
| `client_timeout`   | int, optional  | max seconds to wait for server response             | `10`                  |
| `batch_size`       | int, optional  | max instances per API request                       | `500`                 |
| `latency_io_reqd`  | int, optional  | threshold of IOPs for calculating latency metrics (latencies based on very few IOPs are unreliable) | `100`  |
| `schedule`         | list, required | the poll frequencies of the collector/object, should include exactly these three elements in the exact same other: | |
|    - `counter`         | duration (Go-syntax) | poll frequency of updating the counter metadata cache (example value: `1200s` = `20m`) | |
|    - `instance`         | duration (Go-syntax) | poll frequency of updating the instance cache (example value: `600s` = `10m`) | |
|    - `data`         | duration (Go-syntax) | poll frequency of updating the data cache (example value: `60s` = `1m`)<br /><br />**Note** Harvest allows defining poll intervals on sub-second level (e.g. `1ms`), however keep in mind the following:<br /><ul><li>API response of an ONTAP system can take several seconds, so the collector is likely to enter failed state if the poll interval is less than `client_timeout`.</li><li>Small poll intervals will create significant workload on the ONTAP system, as many counters are aggregated on-demand.</li><li>Some metric values become less significant if they are calculated for very short intervals (e.g. latencies)</li></ul> | |

The template should define objects in the `objects` section. Example:

```yaml

objects:
  SystemNode:             system_node.yaml
  HostAdapter:            hostadapter.yaml
```

Note that for each object we only define the filename of the object configuration file. The object configuration files are located in subdirectories matching to the ONTAP version that was used to create these files. It is possible to have multiple version-subdirectories for multiple ONTAP versions. At runtime, the collector will select the object configuration file that closest matches to the version of the target ONTAP system. (A mismatch is tolerated since ZapiPerf will fetch and validate counter metadata from the system.)

### Object configuration file

The Object configuration file ("subtemplate") should contain the following parameters:

| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `name`                 | string       | display name of the collector that will collect this object |             |
| `object`               | string       | short name of the object                         |                        |
| `query`                | string       | raw object name used to issue a ZAPI request     |                        |
| `counters`             | list         | list of counters to collect (see notes below) |                        |
| `instance_key`         | string | label to use as instance key (either `name` or `uuid`) |                        |
| `override` | list of key-value pairs | override counter properties that we get from ONTAP (allows to circumvent ZAPI bugs) | |
| `plugins`  | list | plugins and their parameters to run on the collected data | |
| `export_options` | list | parameters to pass to exporters (see notes below) | |

#### `counters`

This section defines the list of counters that will be collected. These counters can be labels, numeric metrics or histograms. The exact property of each counter is fetched from ONTAP and updated periodically.

Some counters require a "base-counter" for post-processing. If the base-counter is missing, ZapiPerf will still run, but the missing data won't be exported.

The display name of a counter can be changed with `=>` (e.g. `nfsv3_ops => ops`). The special counter `instance_name` will be renamed to the value of `object` by default.

Counters that are stored as labels will only be exported if they are included in the `export_options` section.

#### `export_options`

Parameters in this section tell the exporters how to handle the collected data. The set of parameters varies by exporter. For [Prometheus](../../exporters/prometheus/README.md) and [InfluxDB]((../../exporters/influxdb/README.md)) exporters, the following parameters can be defined:

* `instances_keys` (list): display names of labels to export with each data-point
* `instance_labels` (list): display names of labels to export as a separate data-point
* `include_all_labels` (bool): export all labels with each data-point (overrides previous two parameters)

## Metrics

The collector collects a dynamic set of metrics. The metric values are calculated from two consecutive polls (therefore no metrics are emitted after the first poll). The calculation algorithm depends on the `property` and `base-counter` attributes of each metric, the following properties are supported:

| property  | formula                                    |  description                                              |
|-----------|--------------------------------------------|-----------------------------------------------------------|
| raw       | x = x<sub>i</sub>                          | no post-processing, value **x** is submitted as it is       |
| delta    | x = x<sub>i</sub> - x<sub>i-1</sub> | delta of two poll values, **x<sub>i<sub>** and **x<sub>i-1<sub>** |
| rate | x = (x<sub>i</sub> - x<sub>i-1</sub>) / (t<sub>i</sub> - t<sub>i-1</sub>) | delta divided by the interval of the two polls in seconds |
| average | x = (x<sub>i</sub> - x<sub>i-1</sub>) / (y<sub>i</sub> - y<sub>i-1</sub>) | delta divided by the delta of the base counter **y** |
| percent | x = 100 * (x<sub>i</sub> - x<sub>i-1</sub>) / (y<sub>i</sub> - y<sub>i-1</sub>) | average multiplied by 100 |

## Creating/editing subtemplates

Instead of editing one of the existing templates, it's better to copy one and edit the copy. That way, your custom template will not be overwritten when upgrading Harvest. Below is an example of collecting sensor metrics.

### Copy an existing template or create a new one

Create `conf/zapi/cdot/9.8.0/sensor.yaml` with the following content:

```
name:                      Sensor
query:                     environment-sensors-get-iter
object:                    sensor

counters:
  environment-sensors-info:
    - critical-high-threshold    => critical_high
    - critical-low-threshold     => critical_low
    - ^discrete-sensor-state     => discrete_state
    - ^discrete-sensor-value     => discrete_value
    - ^^node-name                => node
    - ^^sensor-name              => sensor
    - ^sensor-type               => type
    - ^threshold-sensor-state    => threshold_state
    - threshold-sensor-value     => threshold_value
    - ^value-units               => unit
    - ^warning-high-threshold    => warning_high
    - ^warning-low-threshold     => warning_low

export_options:
  include_all_labels: true
```

### Enable the new template

To enable the new template, create `conf/zapi/custom.yaml` with the lines shown below.

In the future, if you add more templates, you can add those in this same file.

```
objects:
  Sensor: sensor.yaml
```

### Test your changes and restart pollers

Test your new `Sensor` template with a single poller like this:
```
./bin/harvest start --foreground --verbose --collectors Zapi --objects Sensor <poller>
```
Replace `<poller>` with the name of one of your ONTAP pollers.

Once you have confirmed that the new template works, restart any already running pollers that you want to pick up the new template(s).

### Check the metrics

If you are using the Prometheus exporter, check the metrics on the HTTP endpoint with `curl` or a web browser. e.g. my poller is exporting its data on port `15001`. Adjust as needed for your exporter.

```
curl -s 'http://localhost:15001/metrics' | grep sensor_

sensor_value{datacenter="WDRF",cluster="shopfloor",critical_high="3664",node="shopfloor-02",sensor="P3.3V STBY",type="voltage",warning_low="3040",critical_low="2960",threshold_state="normal",unit="mV",warning_high="3568"} 3280
sensor_value{datacenter="WDRF",cluster="shopfloor",sensor="P1.2V STBY",type="voltage",threshold_state="normal",warning_high="1299",warning_low="1105",critical_low="1086",node="shopfloor-02",critical_high="1319",unit="mV"} 1193
sensor_value{datacenter="WDRF",cluster="shopfloor",unit="mV",critical_high="15810",critical_low="0",node="shopfloor-02",sensor="P12V STBY",type="voltage",threshold_state="normal"} 11842
sensor_value{datacenter="WDRF",cluster="shopfloor",sensor="P12V STBY Curr",type="current",threshold_state="normal",unit="mA",critical_high="3182",critical_low="0",node="shopfloor-02"} 748
sensor_value{datacenter="WDRF",cluster="shopfloor",critical_low="1470",node="shopfloor-02",sensor="Sysfan2 F2 Speed",type="fan",threshold_state="normal",unit="RPM",warning_low="1560"} 2820
sensor_value{datacenter="WDRF",cluster="shopfloor",sensor="PSU2 Fan1 Speed",type="fan",threshold_state="normal",unit="RPM",warning_low="4600",critical_low="4500",node="shopfloor-01"} 6900
sensor_value{datacenter="WDRF",cluster="shopfloor",sensor="PSU1 InPwr Monitor",type="unknown",threshold_state="normal",unit="mW",node="shopfloor-01"} 132000
sensor_value{datacenter="WDRF",cluster="shopfloor",critical_high="58",type="thermal",unit="C",warning_high="53",critical_low="0",node="shopfloor-01",sensor="Bat Temp",threshold_state="normal",warning_low="5"} 24
sensor_value{datacenter="WDRF",cluster="shopfloor",critical_high="9000",node="shopfloor-01",sensor="Bat Charge Volt",type="voltage",threshold_state="normal",unit="mV",warning_high="8900"} 8200
sensor_value{datacenter="WDRF",cluster="shopfloor",node="shopfloor-02",sensor="PSU1 InPwr Monitor",type="unknown",threshold_state="normal",unit="mW"} 132000
```
