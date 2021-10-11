## Creating/editing templates

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

This document scope is on two kind of template yamls as below. [Collector templates](#Collector-templates)  and [Object templates](#Object-templates)

## Collector templates

Example of collector template `conf/zapi/default.yaml`
By default Harvest reads from `conf/zapi/default.yaml` (shipped with harvest) and `conf/zapi/custom.yaml` (used to extend `conf/zapi/default.yaml`). 
We can define the configuration file of the collector. If no configuration file is specified, the default configuration file (`conf/zapiperf/default.yaml`) will be used and if the file `conf/zapiperf/custom.yaml` is present, it will be merged to the default one. If we specify our own configuration file for the collector, it can have any name, and it will not be merged.

Examples:
1. Define a poller that will run the ZapiPerf collector using its default configuration file:

```yaml
Pollers:
  jamaica:  # name of the poller
    datacenter: munich
    addr: 10.10.10.10
    auth_style: basic_auth
    username: harvest
    password: pass
    collectors:
      - Zapi # will use conf/zapi/default.yaml and optionally merge with conf/zapi/custom.yaml
```

2. Define a poller that will run the Zapi collector using a custom configuration file:

```yaml
Pollers:
  jamaica:  # name of the poller
    addr: 10.10.10.10
    auth_style: basic_auth
    username: harvest
    password: pass
    collectors:
      - ZapiPerf:
        - limited.yaml # will use conf/zapi/limited.yaml
        # if more templates are added, they will be merged
```

### Object Templates
Object templates contains counter details of an object. Example object template `conf/zapi/cdot/9.8.0/lun.yaml` 

Instead of editing one of the existing templates, it's better to extend existing templates. That way, your custom template will not be overwritten when upgrading Harvest. For example, if you want to extend `conf/zapi/cdot/9.8.0/aggr.yaml`, first create a copy (e.g., `conf/zapi/cdot/9.8.0/custom_aggr.yaml`), then add these lines to `conf/zapi/custom.yaml`:

```yaml
objects:
  Aggregate: custom_aggr.yaml
```

After restarting your pollers, `aggr.yaml` and `custom_aggr.yaml` will be merged.

#### 1. Create a new objectTemplate

In this example, we want to collect sensor metrics from the `environment-sensors-get-iter` API. These are the steps that we need to follow:

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

#### 2. Enable the new objectTemplate

To enable the new objectTemplate, create `conf/zapi/custom.yaml` with the lines shown below.

```yaml
objects:
  Sensor: sensor.yaml
```
In the future, if you add more objectTemplates, you can add those in this same file.

#### 3. Extend an existing objectTemplate

In this example, we want to extend existing default objectTemplate `conf/zapi/cdot/9.8.0/lun.yaml` and collect additional metrics from the `lun-get-iter` zapi. These are the steps that one need to follow:

Existing default `conf/zapi/cdot/9.8.0/lun.yaml` as below

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
    value_mapping:
      - status state online `1`
    # metric label zapi_value rest_value `default_value`
    value_to_num:
      - new_status state online online `0`
    # path is something like "/vol/vol_georg_fcp401/lun401"
    # we only want lun name, which is 4th element
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

To extend this objectTemplate, create `conf/zapi/custom.yaml` with the lines shown below.

```yaml
objects:
  Lun: custom_lun.yaml
```

Create a new objectTemplate `conf/zapi/cdot/9.8.0/custom_lun.yaml` with the lines shown below. 

Use Case:
1. Add additional counters `multiprotocol-type`, `application`
2. Configure a new `value_to_num` plugin
3. Add `application` to instance_keys
4. Overwrite object name

```yaml
name:  CustomLun
counters:
  lun-info:
    - ^multiprotocol-type
    - ^application

plugins:
  - LabelAgent:
    value_to_num:
      - new_status state online online `0`

export_options:
  instance_keys:
    - application
 ```

Harvest will merge these templates during poller start and merged template will be as below

```yaml
name: CustomLun
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
    value_mapping:
      - status state online `1`
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
    - application
```

To view the merged template, you can also run below command to verify the template.
```sh
bin/harvest doctor merge --template lun.yaml --with custom_lun.yaml
```

#### 4. Test your changes and restart pollers

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