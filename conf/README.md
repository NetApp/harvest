## Creating/editing templates

This document covers how to use [Collector](#Collector-templates) and [Object](#Object-templates) templates to extend Harvest.

1. [How to add a new object template](#create-a-new-object-template)
2. [How to extend an existing object template](#extend-an-existing-object-template)
3. [How to replace an existing object template](#replace-an-existing-object-template)

There are a couple of ways to learn about ZAPIs and their attributes:
- [ONTAP's documentation](https://mysupport.netapp.com/documentation/productlibrary/index.html?productID=60427) 
- Using Harvest's `zapi` tool to explore available APIs and metrics on your cluster. Examples:

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

## Collector templates

Collector templates define which set of objects Harvest should collect from the system being monitored. In your `harvest.yml` configuration file, when you say that you want to use a `Zapi` collector, that collector will read the matching `conf/zapi/default.yaml` - same with `ZapiPerf`, it will read the `conf/zapiperf/default.yaml` file. Belows's a snippet from `conf/zapi/default.yaml`. Each object is mapped to a corresponding [object template](#object-templates) file. For example, the `Node` object searches for the [most appropriate version](#harvest-versioned-templates) of the `node.yaml` file in the `conf/zapi/cdot/**` directory.

```
collector:          Zapi
objects:
  Node:             node.yaml
  Aggregate:        aggr.yaml
  Volume:           volume.yaml
  Disk:             disk.yaml
```

Each collector will also check if a matching file named, `custom.yaml` exists, and if it does, it will read that file and merge it with `default.yaml`. The `custom.yaml` file should be located beside the matching `default.yaml` file. (eg. `conf/zapi/custom.yaml` is beside `conf/zapi/default.yaml`). 

Let's take a look at some examples.

1. Define a poller that uses the default Zapi collector. Using the default template is the easiest and most used option.

```yaml
Pollers:
  jamaica:
    datacenter: munich
    addr: 10.10.10.10
    collectors:
      - Zapi # will use conf/zapi/default.yaml and optionally merge with conf/zapi/custom.yaml
```

2. Define a poller that uses the Zapi collector, but with a custom template file:

```yaml
Pollers:
  jamaica:
    datacenter: munich
    addr: 10.10.10.10
    collectors:
      - ZapiPerf:
        - limited.yaml # will use conf/zapiperf/limited.yaml
        # more templates can be added, they will be merged
```

## Object Templates

Object templates (example: `conf/zapi/cdot/9.8.0/lun.yaml`) describe what to collect and export. These templates are used by collectors to gather metrics and send them to your time-series db.

Object templates are made up of the following parts:
1. the name of the object (or resource) to collect
2. the ZAPI or REST query used to collect the object
3. a list of object counters to collect and how to export them

Instead of editing one of the existing templates, it's better to extend one of them. That way, your custom template will not be overwritten when upgrading Harvest. For example, if you want to extend `conf/zapi/cdot/9.8.0/aggr.yaml`, first create a copy (e.g., `conf/zapi/cdot/9.8.0/custom_aggr.yaml`), and then tell Harvest to use your custom template by adding these lines to `conf/zapi/custom.yaml`:

```yaml
objects:
  Aggregate: custom_aggr.yaml
```

After restarting your pollers, `aggr.yaml` and `custom_aggr.yaml` will be merged.

### Create a new object template

In this example, imagine that Harvest doesn't already collect environment sensor data and you wanted to collect it. Sensor does comes from the `environment-sensors-get-iter` ZAPI. Here are the steps to add a new object template.

Create the file `conf/zapi/cdot/9.8.0/sensor.yaml` (optionally replace `9.8.0` with the earliest version of ONTAP that supports sensor data. Refer to [Harvest Versioned Templates](#harvest-versioned-templates) for more information. Add the following content to your new `sensor.yaml` file.

```yaml
name:     Sensor                      # this name must match the key in your custom.yaml file
query:    environment-sensors-get-iter
object:   sensor

metric_type:      int64

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
### Enable the new object template

To enable the new sensor object template, create the `conf/zapi/custom.yaml` file with the lines shown below.

```yaml
objects:
  Sensor: sensor.yaml                 # this key must match the name in your sensor.yaml file
```
The `Sensor` key used in the `custom.yaml` must match the name defined in the `sensor.yaml` file. That mapping is what connects this object with its template. In the future, if you add more object templates, you can add those in your existing `custom.yaml` file.

### Test your object template changes

Test your new `Sensor` template with a single poller like this:
```
./bin/harvest start <poller> --foreground --verbose --collectors Zapi --objects Sensor
```
Replace `<poller>` with the name of one of your ONTAP pollers.

Once you have confirmed that the new template works, restart any already running pollers that you want to use the new template(s).

### Check the metrics

If you are using the Prometheus exporter, you can scrape the poller's HTTP endpoint with curl or a web browser. E.g., my poller exports its data on port 15001. Adjust as needed for your exporter.
```
curl -s 'http://localhost:15001/metrics' | grep ^sensor_  # sensor_ name matches the object: value in your sensor.yaml file.

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
### Extend an existing object template

In this example, we want to extend one of the existing object templates that Harvest ships with, e.g. `conf/zapi/cdot/9.8.0/lun.yaml` and collect additional information as outlined below.

Lets's say you want to extend `lun.yaml` to:

1. Increase `client_timeout` (You want to increase the default timeout of the lun ZAPI because it keeps [timing out](https://github.com/NetApp/harvest/wiki/Troubleshooting-Harvest#client_timeout))
2. Add additional counters, e.g. `multiprotocol-type`, `application`
3. Add a new counter to the already collected lun metrics using the `value_to_num` plugin
4. Add a new `application` instance_keys and labels to the collected metrics

Let's assume the existing template is located at conf/zapi/cdot/9.8.0/lun.yaml and contains the following. 

```yaml
name:                       Lun
query:                      lun-get-iter
object:                     lun

counters:
  lun-info:
    - ^node
    - ^path
    - ^qtree
    - size
    - size-used
    - ^state
    - ^^uuid
    - ^volume
    - ^vserver => svm

plugins:
  - LabelAgent:
    # metric label zapi_value rest_value `default_value`
    value_to_num:
      - new_status state online online `0`
    split:
      - path `/` ,,,lun

export_options:
  instance_keys:
    - node
    - qtree
    - lun
    - volume
    - svm
  instance_labels:
    - state
 ```

To extend the out-of-the-box `lun.yaml` template, create a `conf/zapi/custom.yaml` file if it doesn't already exist and add the lines shown below:

```yaml
objects:
  Lun: custom_lun.yaml
```

Create a new object template `conf/zapi/cdot/9.8.0/custom_lun.yaml` with the lines shown below.

```yaml
client_timeout: 5m
counters:
  lun-info:
    - ^multiprotocol-type
    - ^application

plugins:
  - LabelAgent:
    value_to_num:
      - custom_status state online online `0`

export_options:
  instance_keys:
    - application
 ```

When you restart your pollers, Harvest will take the out-of-the-box template (`lun.yaml`) and your new one (`custom_lun.yaml`) and merge them into the following:

```yaml
name: Lun
query: lun-get-iter
object: lun
counters:
  lun-info:
    - ^node
    - ^path
    - ^qtree
    - size
    - size-used
    - ^state
    - ^^uuid
    - ^volume
    - ^vserver => svm
    - ^multiprotocol-type
    - ^application
plugins:
  LabelAgent:
    value_to_num:
      - new_status state online online `0`
      - custom_status state online online `0`
    split:
      - path `/` ,,,lun
export_options:
  instance_keys:
    - node
    - qtree
    - lun
    - volume
    - svm
    - application
 client_timeout: 5m
```

To help understand the merging process and the resulting combined template, you can view the result with:
```sh
bin/harvest doctor merge --template conf/zapi/cdot/9.8.0/lun.yaml --with conf/zapi/cdot/9.8.0/custom_lun.yaml
```

### Replace an existing object template

You can only extend existing templates as explained [above](#extend-an-existing-object-template). If you need to replace one of the existing object templates, let us know more on Slack or GitHub.

## Harvest Versioned Templates

Harvest ships with a set of versioned templates tailored for specific versions of ONTAP. At runtime, Harvest uses a BestFit heuristic to pick the most appropriate template. The BestFit heuristic compares the list of Harvest templates with the ONTAP version and selects the best match. There are versioned templates for both the ZAPI and REST collectors. Below is an example of how the BestFit algorithm works - assume Harvest has these templated versions:

- 9.6.0
- 9.6.1
- 9.8.0
- 9.9.0
- 9.10.1

if you are monitoring a cluster at these versions, Harvest will select the indicated template:

- ONTAP version 9.4.1, Harvest will select the templates for 9.6.0
- ONTAP version 9.6.0, Harvest will select the templates for 9.6.0
- ONTAP version 9.7.X, Harvest will select the templates for 9.6.1
- ONTAP version  9.12, Harvest will select the templates for 9.10.1

### counters

This section contains the complete or partial attribute tree of the queried API. Since the collector does not get counter metadata from the ONTAP system, two additional symbols are used for non-numeric attributes:

- `^` used as a prefix indicates that the attribute should be stored as a label
- `^^` indicates that the attribute is a label and an instance key (i.e., a label that uniquely identifies an instance, such as `name`, `uuid`). If a single label does not uniquely identify an instance, then multiple instance keys should be indicated.

Additionally, the symbol `=>` can be used to set a custom display name for both instance labels and numeric counters. Example:

```yaml
name:                     Spare
query:                    aggr-spare-get-iter
object:                   spare
collect_only_labels:      true
counters:
  aggr-spare-disk-info:
    - ^^disk                                # creates label aggr-disk
    - ^disk-type                            # creates label aggr-disk-type
    - ^is-disk-zeroed   => is_disk_zeroed   # creates label is_disk_zeroed
    - ^^original-owner  => original_owner   # creates label original_owner
export_options:
  instance_keys:
    - disk
    - original_owner
  instance_labels:
    - disk_type
    - is_disk_zeroed
```

Harvest does its best to determine a unique display name for each template's label and metric. Instead of relying on this heuristic, it is better to be explicit in your templates and define a display name using the caret (`^`) mapping. For example, instead of this:
```
aggr-spare-disk-info:
    - ^^disk
    - ^disk-type
```
do this:
```
aggr-spare-disk-info:
    - ^^disk      => disk
    - ^disk-type  => disk_type
```
See also [#585](https://github.com/NetApp/harvest/issues/585)
