This document describes how Harvest metrics relate to their relevant Cisco Switch, including:

- Details about which Harvest metrics each dashboard uses.
These can be generated on demand by running `bin/harvest grafana metrics`. See
[#1577](https://github.com/NetApp/harvest/issues/1577#issue-1471478260) for details.

```
Creation Date : 2025-Jul-15
ONTAP Version: 9.16.1
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

cisco_switch_labels <span class="key">Name of the metric exported by Harvest</span>

Displays detail of the Cisco switch. <span class="key">Description of the Cisco Switch metric</span>

* <span class="key">Metric</span> name of the Cisco Switch metric
* <span class="key">Template</span> path of the template that collects the metric

| Metric | Template |
|--------|---------|
|`cisco_switch_labels` | conf/ciscorest/nxos/9.3.12/version.yaml |

## Metrics


### cisco_cdp_neighbor_labels

Displays cisco discovery protocol information about neighbors in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/cdp.yaml |

The `cisco_cdp_neighbor_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Neighbors | table | [Cisco Discovery Protocol](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=46) |
///



### cisco_environment_fan_speed

Displays zone speed of the fan in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_fan_speed` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Highlights | table | [Switch Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=5) |
///



### cisco_environment_fan_up

Displays Present/Absent Status of the fan in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_fan_up` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Temperature and Fan | table | [Fan Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=18) |
///



### cisco_environment_power_capacity

Displays total capacity of the power supply in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_power_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Power | table | [PSU Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=35) |
///



### cisco_environment_power_in

Displays actual input power in watts of power supply in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_power_in` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Power | stat | [Total Power](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=33) |
| Cisco: Switch | Power | stat | [PSU Efficiency](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=9) |
| Cisco: Switch | Power | timeseries | [Top $TopResources Power Consumption](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=8) |
///



### cisco_environment_power_mode

Displays redundant or operational Mode of power supply in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_power_mode` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Highlights | table | [Switch Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=5) |
///



### cisco_environment_power_out

Displays actual output power in watts of power supply in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_power_out` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Power | stat | [PSU Efficiency](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=9) |
| Cisco: Switch | Power | timeseries | [Top $TopResources Power Consumption](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=8) |
| Cisco: Switch | Power | table | [PSU Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=35) |
///



### cisco_environment_power_up

Displays power supply status in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_power_up` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Power | table | [PSU Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=35) |
///



### cisco_environment_sensor_temp

Displays current temperature of sensor in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_sensor_temp` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Temperature and Fan | timeseries | [Top $TopResources Sensor Temperatures](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=10) |
///



### cisco_interface_admin_up

Displays admin state of the interface in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_admin_up` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Interfaces | stat | [Down (Last 24h)](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=37) |
| Cisco: Switch | Interfaces | table | [Down (Last 24h)](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=39) |
| Cisco: Switch | Interfaces | timeseries | [Down (Last 24h)](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=40) |
///



### cisco_interface_crc_errors

Displays CRC of interface in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_crc_errors` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface CRC error](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=34) |
///



### cisco_interface_receive_bytes

Displays bytes input of the interface in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_receive_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Receive Traffic](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=29) |
///



### cisco_interface_receive_drops

Displays input if-down drops of interface in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_receive_drops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Receive Drops](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=31) |
///



### cisco_interface_receive_errors

Displays input errors of interface in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_receive_errors` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Errors](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=32) |
///



### cisco_interface_transmit_bytes

Displays bytes output of interface in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_transmit_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Transmit Traffic](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=28) |
///



### cisco_interface_transmit_drops

Displays output drops of interface in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_transmit_drops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Receive Drops](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=31) |
///



### cisco_interface_transmit_errors

Displays output errors of interface in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_transmit_errors` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Errors](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=32) |
///



### cisco_lldp_neighbor_labels

Displays link layer discovery protocol information about neighbours in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/lldp.yaml |

The `cisco_lldp_neighbor_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Neighbors | table | [Link Layer Discovery Protocol](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=44) |
///



### cisco_optic_rx

Displays rx power of the interface in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/optic.yaml |

The `cisco_optic_rx` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Transceiver | timeseries | [Top $TopResources Transceiver RX Power](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=21) |
///



### cisco_optic_tx

Displays tx power of the interface in the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/optic.yaml |

The `cisco_optic_tx` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Transceiver | timeseries | [Top $TopResources Transceiver TX Power](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=23) |
///



### cisco_switch_labels

Displays configuration detail of the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/version.yaml |

The `cisco_switch_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Highlights | table | [Switch Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=5) |
///



### cisco_switch_uptime

Displays uptime duration of the Cisco switch.


| Template |
|---------|
| conf/ciscorest/nxos/9.3.12/version.yaml |

The `cisco_switch_uptime` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Highlights | table | [Switch Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=5) |
///



