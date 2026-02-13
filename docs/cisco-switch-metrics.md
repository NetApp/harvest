This document describes which Cisco switch metrics are collected and what those metrics are named in Harvest, including:

- Details about which Harvest metrics each dashboard uses.
These can be generated on demand by running `bin/harvest grafana metrics`. See
[#1577](https://github.com/NetApp/harvest/issues/1577#issue-1471478260) for details.

```
Creation Date : 2026-Feb-13
NX-OS Version: 9.3.12
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

cisco_switch_uptime <span class="key">Name of the metric exported by Harvest</span>

Displays uptime duration of the Cisco switch. <span class="key">Description of the Cisco switch metric</span>

* <span class="key">API</span> Harvest uses the NXAPI protocol to collect metrics
* <span class="key">Endpoint</span> name of the CLI used to collect this metric
* <span class="key">Metric</span> name of the Cisco switch metric
* <span class="key">Template</span> path of the template that collects the metric

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
|NXAPI | `show version` | kern_uptm_days, kern_uptm_hrs, kern_uptm_mins, kern_uptm_secs | conf/ciscorest/nxos/9.3.12/version.yaml|


??? "Example to invoke CLI show commands via curl"

    In this example, we would demonstrate invoking the `show version` CLI command via curl.

    To do this, send a POST request to your switch’s IP address with the desired command as input.
    Replace RO_USER, PASSWORD, and CISCO_SWITCH_IP with your actual read-only username, password, and the switch’s IP address.

    ```
    curl -sk -u RO_USER:PASSWORD POST 'https://CISCO_SWITCH_IP/ins_api' -d
    '{
    "ins_api": {
    "version": "1.0",
    "type": "cli_show",
    "chunk": "0",
    "sid": "1",
    "input": "show version",
    "output_format": "json"
    }
    }'
    ```

    After invoking the above Curl command, You would get this response
    ```
    {
            "ins_api":      {
                    "type": "cli_show",
                    "version":      "1.0",
                    "sid":  "eoc",
                    "outputs":      {
                            "output":       {
                                    "input": "show version",
                                    "msg":  "Success",
                                    "code": "200",
                                    "body": {
                                            "header_str":   "Cisco Nexus Operating System (NX-OS) Software\nTAC support: http://www.cisco.com/tac\nCopyright (C) 2002-2023, Cisco and/or its affiliates.\nAll rights reserved.\nThe copyrights to certain works contained in this software are\nowned by other third parties and used and distributed under their own\nlicenses, such as open source.  This software is provided \"as is,\" and unless\notherwise stated, there is no warranty, express or implied, including but not\nlimited to warranties of merchantability and fitness for a particular purpose.\nCertain components of this software are licensed under\nthe GNU General Public License (GPL) version 2.0 or \nGNU General Public License (GPL) version 3.0  or the GNU\nLesser General Public License (LGPL) Version 2.1 or \nLesser General Public License (LGPL) Version 2.0. \nA copy of each such license is available at\nhttp://www.opensource.org/licenses/gpl-2.0.php and\nhttp://opensource.org/licenses/gpl-3.0.html and\nhttp://www.opensource.org/licenses/lgpl-2.1.php and\nhttp://www.gnu.org/licenses/old-licenses/library.txt.\n",
                                            "bios_ver_str": "04.25",
                                            "kickstart_ver_str":    "9.3(12)",
                                            "nxos_ver_str": "9.3(12)",
                                            "bios_cmpl_time":       "05/22/2019",
                                            "kick_file_name":       "bootflash:///nxos.9.3.12.bin",
                                            "nxos_file_name":       "bootflash:///nxos.9.3.12.bin",
                                            "kick_cmpl_time":       "6/20/2023 12:00:00",
                                            "nxos_cmpl_time":       "6/20/2023 12:00:00",
                                            "kick_tmstmp":  "06/23/2023 17:33:36",
                                            "nxos_tmstmp":  "06/23/2023 17:33:36",
                                            "chassis_id":   "Nexus 3132QV Chassis",
                                            "cpu_name":     "Intel(R) Core(TM) i3- CPU @ 2.50GHz",
                                            "memory":       16399572,
                                            "mem_type":     "kB",
                                            "proc_board_id":        "FOC24213H5C",
                                            "host_name":    "Switch-A1",
                                            "bootflash_size":       15137792,
                                            "slot0_size":   0,
                                            "kern_uptm_days":       256,
                                            "kern_uptm_hrs":        19,
                                            "kern_uptm_mins":       3,
                                            "kern_uptm_secs":       50,
                                            "rr_usecs":     24056,
                                            "rr_ctime":     "Wed Nov  6 14:02:05 2024",
                                            "rr_reason":    "Reset Requested by CLI command reload",
                                            "rr_sys_ver":   "9.3(12)",
                                            "rr_service":   "",
                                            "plugins":      "Core Plugin, Ethernet Plugin",
                                            "manufacturer": "Cisco Systems, Inc.",
                                            "TABLE_package_list":   {
                                                    "ROW_package_list":     {
                                                            "package_id":   ""
                                                    }
                                            }
                                    }
                            }
                    }
            }
    }
    ```


## Metrics


### cisco_cdp_neighbor_labels

Displays cisco discovery protocol information about neighbors in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show cdp neighbors detail` | `device_id, platform_id, port_id, ttl, version, local_intf_mac, remote_intf_mac, capability` | conf/ciscorest/nxos/9.3.12/cdp.yaml |

The `cisco_cdp_neighbor_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Neighbors | table | [Cisco Discovery Protocol](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=46) |
///



### cisco_environment_fan_speed

Displays fan speed.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show environment fan detail` | `speed` | conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_fan_speed` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Temperature and Fan | table | [Fan Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=18) |
///



### cisco_environment_fan_up

Displays Present/Absent Status of the fan in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show environment` | `fanstatus` | conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_fan_up` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Temperature and Fan | table | [Fan Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=18) |
///



### cisco_environment_fan_zone_speed

Displays the zone fan speed of the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show environment fan detail` | `zonespeed` | conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_fan_zone_speed` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Highlights | table | [Switch Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=5) |
///



### cisco_environment_power_capacity

Displays total capacity of the power supply in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show environment` | `tot_capa` | conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_power_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Power | table | [PSU Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=35) |
///



### cisco_environment_power_in

Displays actual input power in watts of power supply in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show environment` | `actual_input OR watts` | conf/ciscorest/nxos/9.3.12/environment.yaml |

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


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show environment` | `ps_redun_mode, ps_oper_mode` | conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_power_mode` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Highlights | table | [Switch Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=5) |
///



### cisco_environment_power_out

Displays actual output power in watts of power supply in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show environment` | `actual_out` | conf/ciscorest/nxos/9.3.12/environment.yaml |

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


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show environment` | `ps_status` | conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_power_up` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Power | table | [PSU Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=35) |
///



### cisco_environment_sensor_temp

Displays current temperature of sensor in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show environment` | `curtemp` | conf/ciscorest/nxos/9.3.12/environment.yaml |

The `cisco_environment_sensor_temp` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Temperature and Fan | timeseries | [Top $TopResources Sensor Temperatures](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=10) |
///



### cisco_interface_admin_up

Displays admin state of the interface in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show interface` | `admin_state` | conf/ciscorest/nxos/9.3.12/interface.yaml |

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


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show interface` | `eth_crc` | conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_crc_errors` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface CRC error](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=34) |
///



### cisco_interface_receive_bytes

Displays bytes input of the interface in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show interface` | `eth_inbytes` | conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_receive_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | table | [Traffic on Switch](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=48) |
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Receive Throughput](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=29) |
///



### cisco_interface_receive_drops

Displays input if-down drops of interface in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show interface` | `eth_in_ifdown_drops` | conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_receive_drops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Receive Drops](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=31) |
///



### cisco_interface_receive_errors

Displays input errors of interface in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show interface` | `eth_inerr` | conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_receive_errors` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Errors](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=32) |
///



### cisco_interface_transmit_bytes

Displays bytes output of interface in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show interface` | `eth_outbytes` | conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_transmit_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | table | [Traffic on Switch](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=48) |
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Send Throughput](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=28) |
///



### cisco_interface_transmit_drops

Displays output drops of interface in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show interface` | `eth_out_drops` | conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_transmit_drops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Receive Drops](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=31) |
///



### cisco_interface_transmit_errors

Displays output errors of interface in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show interface` | `eth_outerr` | conf/ciscorest/nxos/9.3.12/interface.yaml |

The `cisco_interface_transmit_errors` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Traffic | timeseries | [Top $TopResources Interface Errors](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=32) |
///



### cisco_lldp_neighbor_labels

Displays link layer discovery protocol information about neighbours in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show lldp neighbors detail` | `sys_name, sys_desc, chassis_id, l_port_id, ttl, port_id, enabled_capability` | conf/ciscorest/nxos/9.3.12/lldp.yaml |

The `cisco_lldp_neighbor_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Neighbors | table | [Link Layer Discovery Protocol](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=44) |
///



### cisco_optic_rx

Displays rx power of the interface in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show interface transceiver details` | `rx_pwr` | conf/ciscorest/nxos/9.3.12/optic.yaml |

The `cisco_optic_rx` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Transceiver | timeseries | [Top $TopResources Transceiver RX Power](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=21) |
///



### cisco_optic_tx

Displays tx power of the interface in the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show interface transceiver details` | `tx_pwr` | conf/ciscorest/nxos/9.3.12/optic.yaml |

The `cisco_optic_tx` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Transceiver | timeseries | [Top $TopResources Transceiver TX Power](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=23) |
///



### cisco_switch_labels

Displays configuration detail of the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show version` | `bios_ver_str, chassis_id, host_name, nxos_ver_str,` | conf/ciscorest/nxos/9.3.12/version.yaml |
| NXAPI | `show banner motd` | `banner_msg.b_msg` | conf/ciscorest/nxos/9.3.12/version.yaml |


### cisco_switch_uptime

Displays uptime duration of the Cisco switch.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| NXAPI | `show version` | `kern_uptm_days, kern_uptm_hrs, kern_uptm_mins, kern_uptm_secs` | conf/ciscorest/nxos/9.3.12/version.yaml |

The `cisco_switch_uptime` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| Cisco: Switch | Highlights | table | [Switch Details](/d/cisco-switch/cisco3a-switch?orgId=1&viewPanel=5) |
///



