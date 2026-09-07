This document describes which Arista switch metrics are collected and what those metrics are named in Harvest, including:

- Details about which Harvest metrics each dashboard uses.
These can be generated on demand by running `bin/harvest grafana metrics`. See
[#1577](https://github.com/NetApp/harvest/issues/1577#issue-1471478260) for details.

```
Creation Date : 2026-Sep-03
EOS Version: 4.18.5
```

??? "Navigate to Grafana dashboards"

    Add your Grafana instance to the following form and save it. When you click on dashboard links on this page, a link to your dashboard will be opened. NAbox hosts Grafana on a subdomain like so: https://localhost/grafana/

    <div>
        <label for="grafanaHost">Grafana Host</label>
        <input type="text" id="grafanaHost" name="grafanaHost" placeholder="e.g. http://localhost:3000" style="width: 80%;margin-left:1em">
        <button type="button" onclick="saveGrafanaHost()">Save</button>
    </div>

## Understanding the structure

Below is an <span class="key">annotated</span> example of how to interpret the structure of each of the [metrics](#metrics).

arista_switch_uptime <span class="key">Name of the metric exported by Harvest</span>

Displays uptime duration of the Arista switch. <span class="key">Description of the Arista switch metric</span>

* <span class="key">API</span> Harvest uses the Arista eAPI (JSON-RPC runCmds) to collect metrics
* <span class="key">Endpoint</span> name of the CLI command used to collect this metric
* <span class="key">Metric</span> name of the field in the eAPI response
* <span class="key">Template</span> path of the template that collects the metric

| API   | Endpoint | Metric | Template |
|-------|----------|--------|----------|
| eAPI | `show version` | bootupTimestamp | conf/aristarest/eos/4.18.5/version.yaml |


??? "Example to invoke eAPI show commands via curl"

    Arista EOS exposes a JSON-RPC endpoint at `/command-api`. The `runCmds` method accepts an
    array of CLI commands and returns their structured output.
    Replace USER, PASSWORD, and ARISTA_SWITCH_IP with your actual username, password, and the switch's IP address.

    ```
    curl -sk -u USER:PASSWORD 'https://ARISTA_SWITCH_IP/command-api' \
      -H 'Content-Type: application/json' -d
    '{
      "jsonrpc": "2.0",
      "method": "runCmds",
      "params": {
        "version": 1,
        "format": "json",
        "cmds": ["show version"]
      },
      "id": "1"
    }'
    ```

    After invoking the above curl command, you would get this response:
    ```
    {
        "jsonrpc": "2.0",
        "id": "1",
        "result": [
            {
                "modelName": "DCS-7010T-48-R",
                "internalVersion": "4.18.5M",
                "systemMacAddress": "00:1c:73:00:00:00",
                "serialNumber": "JPE00000000",
                "hardwareRevision": "00.00",
                "architecture": "i386",
                "version": "4.18.5M",
                "bootupTimestamp": 1625520901.0,
                "uptime": 123456.78,
                "memTotal": 4017088,
                "memFree": 1872308
            }
        ]
    }
    ```


## Metrics


### arista_environment_ambient_temp

Displays the switch ambient (inlet) temperature.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show environment cooling` | `ambientTemperature` | conf/aristarest/eos/4.18.5/environment.yaml |


### arista_environment_fan_speed

Displays the current fan speed.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show environment cooling` | `actualSpeed` | conf/aristarest/eos/4.18.5/environment.yaml |

The `arista_environment_fan_speed` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Temperature and Fan | table | [Fan Details](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=18) |
///



### arista_environment_fan_up

Indicates whether the fan is operating normally (1) or not (0).


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show environment cooling` | `status` | conf/aristarest/eos/4.18.5/environment.yaml |

The `arista_environment_fan_up` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Temperature and Fan | table | [Fan Details](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=18) |
///



### arista_environment_power_capacity

Displays the rated capacity of a power supply.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show environment power` | `capacity` | conf/aristarest/eos/4.18.5/environment.yaml |

The `arista_environment_power_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Power | table | [PSU Details](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=35) |
///



### arista_environment_power_in

Displays the input power of a power supply.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show environment power` | `inputPower` | conf/aristarest/eos/4.18.5/environment.yaml |

The `arista_environment_power_in` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Power | stat | [Total Power](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=33) |
| Arista: Switch | Power | stat | [PSU Efficiency](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=9) |
| Arista: Switch | Power | timeseries | [Top $TopResources Power Consumption](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=8) |
///



### arista_environment_power_out

Displays the output power of a power supply.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show environment power` | `outputPower` | conf/aristarest/eos/4.18.5/environment.yaml |

The `arista_environment_power_out` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Power | stat | [PSU Efficiency](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=9) |
| Arista: Switch | Power | timeseries | [Top $TopResources Power Consumption](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=8) |
| Arista: Switch | Power | table | [PSU Details](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=35) |
///



### arista_environment_power_up

Indicates whether the power supply is operating normally (1) or not (0).


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show environment power` | `state` | conf/aristarest/eos/4.18.5/environment.yaml |

The `arista_environment_power_up` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Power | table | [PSU Details](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=35) |
///



### arista_environment_sensor_temp

Displays the current temperature reported by a switch temperature sensor.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show environment temperature` | `currentTemperature` | conf/aristarest/eos/4.18.5/environment.yaml |

The `arista_environment_sensor_temp` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Temperature and Fan | timeseries | [Top $TopResources Sensor Temperatures](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=10) |
///



### arista_interface_admin_up

Indicates whether the interface is administratively enabled (1) or disabled (0).


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces` | `interfaceStatus` | conf/aristarest/eos/4.18.5/interface.yaml |

The `arista_interface_admin_up` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Interfaces | stat | [Down (Last 24h)](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=37) |
| Arista: Switch | Interfaces | table | [Down (Last 24h)](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=39) |
| Arista: Switch | Interfaces | timeseries | [Down (Last 24h)](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=40) |
///



### arista_interface_crc_errors

Displays the number of CRC errors received on the interface.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces` | `interfaceCounters.inputErrorsDetail.fcsErrors` | conf/aristarest/eos/4.18.5/interface.yaml |

The `arista_interface_crc_errors` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Traffic | timeseries | [Top $TopResources Interface CRC error](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=34) |
///



### arista_interface_receive_broadcast

Displays the number of inbound broadcast packets received on the interface.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces` | `interfaceCounters.inBroadcastPkts` | conf/aristarest/eos/4.18.5/interface.yaml |


### arista_interface_receive_bytes

Displays the number of bytes received on the interface.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces` | `interfaceCounters.inOctets` | conf/aristarest/eos/4.18.5/interface.yaml |

The `arista_interface_receive_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Traffic | table | [Traffic on Switch](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=48) |
| Arista: Switch | Traffic | timeseries | [Top $TopResources Interface Receive Throughput](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=29) |
///



### arista_interface_receive_drops

Displays the number of inbound packets dropped on the interface.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces` | `interfaceCounters.inDiscards` | conf/aristarest/eos/4.18.5/interface.yaml |

The `arista_interface_receive_drops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Traffic | timeseries | [Top $TopResources Interface Drops](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=31) |
///



### arista_interface_receive_errors

Displays the number of inbound errors on the interface.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces` | `interfaceCounters.totalInErrors` | conf/aristarest/eos/4.18.5/interface.yaml |

The `arista_interface_receive_errors` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Traffic | timeseries | [Top $TopResources Interface Errors](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=32) |
///



### arista_interface_receive_multicast

Displays the number of inbound multicast packets received on the interface.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces` | `interfaceCounters.inMulticastPkts` | conf/aristarest/eos/4.18.5/interface.yaml |


### arista_interface_transmit_bytes

Displays the number of bytes transmitted on the interface.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces` | `interfaceCounters.outOctets` | conf/aristarest/eos/4.18.5/interface.yaml |

The `arista_interface_transmit_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Traffic | table | [Traffic on Switch](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=48) |
| Arista: Switch | Traffic | timeseries | [Top $TopResources Interface Send Throughput](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=28) |
///



### arista_interface_transmit_drops

Displays the number of outbound packets dropped on the interface.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces` | `interfaceCounters.outDiscards` | conf/aristarest/eos/4.18.5/interface.yaml |

The `arista_interface_transmit_drops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Traffic | timeseries | [Top $TopResources Interface Drops](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=31) |
///



### arista_interface_transmit_errors

Displays the number of outbound errors on the interface.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces` | `interfaceCounters.totalOutErrors` | conf/aristarest/eos/4.18.5/interface.yaml |

The `arista_interface_transmit_errors` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Traffic | timeseries | [Top $TopResources Interface Errors](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=32) |
///



### arista_interface_up

Indicates whether the interface line protocol is up (1) or down (0).


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces` | `lineProtocolStatus` | conf/aristarest/eos/4.18.5/interface.yaml |

The `arista_interface_up` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Interfaces | table | [Down (Last 24h)](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=39) |
///



### arista_lldp_neighbor_labels

Displays link layer discovery protocol information about neighbors of the Arista switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show lldp neighbors detail` | `neighborDevice, neighborPort, systemName, systemDescription, chassisId, capabilities` | conf/aristarest/eos/4.18.5/lldp.yaml |

The `arista_lldp_neighbor_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Neighbors | table | [Link Layer Discovery Protocol](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=44) |
///



### arista_optic_rx

Displays the receive optical power (dBm) of the transceiver.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces transceiver` | `rxPower` | conf/aristarest/eos/4.18.5/optic.yaml |

The `arista_optic_rx` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Transceiver | timeseries | [Top $TopResources Transceiver RX Power](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=21) |
///



### arista_optic_temperature

Displays the transceiver temperature in degrees Celsius.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces transceiver` | `temperature` | conf/aristarest/eos/4.18.5/optic.yaml |


### arista_optic_tx

Displays the transmit optical power (dBm) of the transceiver.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces transceiver` | `txPower` | conf/aristarest/eos/4.18.5/optic.yaml |

The `arista_optic_tx` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Transceiver | timeseries | [Top $TopResources Transceiver TX Power](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=23) |
///



### arista_optic_voltage

Displays the transceiver supply voltage.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show interfaces transceiver` | `voltage` | conf/aristarest/eos/4.18.5/optic.yaml |


### arista_switch_labels

Displays identifying configuration detail of the Arista switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show version ; show hostname` | `modelName, serialNumber, hostname, version, internalVersion, systemMacAddress, hardwareRevision, architecture` | conf/aristarest/eos/4.18.5/version.yaml |

The `arista_switch_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Highlights | table | [Switch Details](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=5) |
///



### arista_switch_uptime

Displays uptime duration of the Arista switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| eAPI | `show version` | `bootupTimestamp` | conf/aristarest/eos/4.18.5/version.yaml |

The `arista_switch_uptime` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Arista: Switch | Highlights | table | [Switch Details](/d/arista-switch/arista3a-switch?orgId=1&viewPanel=5) |
///



