

# Zapi

Zapi collects data from ONTAP systems using the ZAPI protocol. The collector submits data as received from the target system, and does not perform any calculations or post-processing. Since the attributes of most APIs have an irregular tree structure, sometimes a plugin will be required to collect all metrics from an API.

Note that the [ZapiPerf collector](../zapiperf/README.md) is an extension of this collector, therefore many parameters and configuration settings will coincide.

## Target System
Target system can be any cDot or 7Mode ONTAP system. Any version is supported, however the default configuration files may not completely match with an older system.

## Requirements
No SDK or any other requirement. It is recommended to create a read-only user for Harvest on the ONTAP system (see the [Authentication document](../../../docs/AuthAndPermissions.md))

## Metrics

The collector collects a dynamic set of metrics. Since most ZAPIs have a tree structure, the collector converts that structure into a flat metric representation. No post-processing or calculation is performed on the collected data itself. 

As an example, the `aggr-get-iter` ZAPI provides the following partial attribute tree:

```yaml
aggr-attributes:
  - aggr-raid-attributes:
    - disk-count
  - aggr-snapshot-attributes:
    - files-total
```

The Zapi collector will convert this tree into two "flat" metrics: `aggr_raid_disk_count` and `aggr_snapshot_files_total`. (The algorithm to generate a name for the metrics will attempt to keep it as simple as possible, but sometimes it's useful to manually set a short display name (see [#counters](#counters)))

## Parameters

The parameters and configuration are similar to those of the [ZapiPerf collector](../zapiperf/README.md). Only the differences will be discussed below.

### Collector configuration file

Parameters different from ZapiPerf:


| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `schedule`             | required     | same as for ZapiPerf, but only two elements: `instance` and `data` (collector does not run a `counter` poll) ||
| `no_max_records`       | bool, optional | don't add `max-records` to the ZAPI request    |                        |
| `collect_only_labels`  | bool, optional | don't look for numeric metrics, only submit labels  (suppresses the `ErrNoMetrics` error)| |
| `only_cluster_instance` | bool, optional | don't look for instance keys and assume only instance is the cluster itself ||


### Object configuration file

The Zapi collector does not have the parameters `instance_key` and `override` parameters. The optional parameter `metric_type` allows you to override the default metric type (`uint64`). The value of this parameter should be one of the metric types supported by [the Matrix data-structure](../../../pkg/matrix/README.md#add-metrics).

#### `counters`

This section contains the complete or partial attribute tree of the queried API. Since the collector does not get counter metadata from the ONTAP system, two additional symbols are used for non-numeric attributes:

- `^` used as a prefix indicates that the attribute should be stored as a label
- `^^` indicates that the attribute is a label and an instance key (i.e., a label that uniquely identifies an instance, such as `name`, `uuid`). If a single label does not uniquely identify an instance, then multiple instance keys should be indicated.

Additionally, the symbol `=>` can be used to set a custom display name for both instance labels and numeric counters. Example:

```yaml
aggr-attributes:
  - aggr-raid-attributes:
    - ^aggregate-type    => type
    - disk-count     => disks
```

will force using `aggr_type` and `aggr_disks` for the label and the metric respectively.

## Creating/editing subtemplates

You can either read [ONTAP's documentation](https://mysupport.netapp.com/documentation/productlibrary/index.html?productID=60427) or use Harvest's `zapi` tool to explore available APIs and metrics on your cluster. Examples:

```sh
$ harvest zapi --poller <poller> show apis
  # will print list of apis that are available
  # usually apis with the "get-iter" suffix can provide useful metrics
$ harvest zapi --poller <poller> show attrs --api volume-get-iter
  # will print the attribute tree of the API
$ harvest zapi --poller <poller> show data --api volume-get-iter
  # will print raw data of the API attribute tree
```

(Replace `<poller>` with the name of a poller that can connect to an ONTAP system.)

Instead of editing one of the existing templates, it's better to copy one and edit the copy. That way, your custom template will not be overwritten when upgrading Harvest. For example, if you want to change `conf/zapi/cdot/9.8.0/aggr.yaml`, first create a copy (e.g., `conf/zapi/cdot/9.8.0/custom_aggr.yaml`), then add these lines to `conf/zapi/custom.yaml`:

```yaml
objects:
  Aggregate: custom_aggr.yaml
```

After restarting your pollers, `aggr.yaml` will be ignored and the new, `custom_aggr.yaml` subtemplate will be used instead.

### Example subtemplate

In this example, we want to collect sensor metrics from the `environment-sensors-get-iter` API. These are the steps that we need to follow:

#### 1. Create a new subtemplate

Create the file `conf/zapi/cdot/9.8.0/sensor.yaml` (optionally replace `9.8.0` with the version of your ONTAP). Add following content:

```yaml
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

(See [#counters](#counters) for an explanation about the special symbols used).

#### 2. Enable the new subtemplate

To enable the new subtemplate, create `conf/zapi/custom.yaml` with the lines shown below.

```yaml
objects:
  Sensor: sensor.yaml
```
In the future, if you add more subtemplates, you can add those in this same file.

#### 3. Test your changes and restart pollers

Test your new `Sensor` template with a single poller like this:
```
./bin/harvest start <poller> --foreground --verbose --collectors Zapi --objects Sensor
```
Replace `<poller>` with the name of one of your ONTAP pollers.

Once you have confirmed that the new template works, restart any already running pollers that you want to pick up the new template(s).

### Check the metrics

If you are using the Prometheus exporter, check the metrics on the HTTP endpoint with `curl` or a web browser. E.g., my poller is exporting its data on port `15001`. Adjust as needed for your exporter.

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
