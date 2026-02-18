This document describes which E-Series metrics are collected and what those metrics are named in Harvest, including:

- Details about which Harvest metrics each dashboard uses.
These can be generated on demand by running `bin/harvest grafana metrics`. See
[#1577](https://github.com/NetApp/harvest/issues/1577#issue-1471478260) for details.

```
Creation Date : 2026-Feb-18
E-Series Version: 11.80.0
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

eseries_volume_read_ops <span class="key">Name of the metric exported by Harvest</span>

Volume read I/O operations per second. <span class="key">Description of the E-Series metric</span>

* <span class="key">API</span> will be REST since E-Series uses the REST API
* <span class="key">Endpoint</span> name of the REST API endpoint used to collect this metric
* <span class="key">Metric</span> name of the E-Series counter
* <span class="key">Template</span> path of the template that collects the metric

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
|REST | `storage-systems/{array_id}/live-statistics` | readOps | conf/eseriesperf/11.80.0/volume.yaml|


## Metrics


### eseries_array_cache_hit_ops

Total number of IO operations that hit cache on the array


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `cacheHitsIopsTotal` | conf/eseriesperf/11.80.0/array.yaml |

The `eseries_array_cache_hit_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Array | Highlights | timeseries | [Top $TopResources Arrays by Cache Hit IOPS](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=201) |
///



### eseries_array_drive_count

Total number of drives in the storage array


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems` | `driveCount` | conf/eseries/11.80.0/array.yaml |


### eseries_array_free_pool_space

Free space available in storage pools in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems` | `freePoolSpace` | conf/eseries/11.80.0/array.yaml |

The `eseries_array_free_pool_space` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Array | Capacity | timeseries | [Top $TopResources Systems by Storage Capacity Used %](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=4) |
| E-Series: Array | Capacity | timeseries | [Top $TopResources Systems by Free Space](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=6) |
///



### eseries_array_host_spares_used

Number of hot spare drives currently in use


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems` | `hostSparesUsed` | conf/eseries/11.80.0/array.yaml |


### eseries_array_labels

This metric provides information about E-Series storage arrays.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems` | `Harvest generated` | conf/eseries/11.80.0/array.yaml |

The `eseries_array_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Array | Capacity | table | [Array Configuration](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=2) |
///



### eseries_array_read_data

Array-wide read data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readBytesTotal` | conf/eseriesperf/11.80.0/array.yaml |

The `eseries_array_read_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Array | Highlights | timeseries | [Top $TopResources Arrays by Read Throughput](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=204) |
///



### eseries_array_read_ops

Array-wide read I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readIopsTotal` | conf/eseriesperf/11.80.0/array.yaml |

The `eseries_array_read_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Array | Highlights | timeseries | [Top $TopResources Arrays by Read IOPS](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=202) |
///



### eseries_array_tray_count

Number of drive trays in the storage array


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems` | `trayCount` | conf/eseries/11.80.0/array.yaml |


### eseries_array_unconfigured_space

Unconfigured space available in the storage array in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems` | `unconfiguredSpace` | conf/eseries/11.80.0/array.yaml |

The `eseries_array_unconfigured_space` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Array | Capacity | timeseries | [Top $TopResources Systems by Unconfigured Space](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=7) |
///



### eseries_array_used_pool_space

Used space in storage pools in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems` | `usedPoolSpace` | conf/eseries/11.80.0/array.yaml |

The `eseries_array_used_pool_space` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Array | Capacity | timeseries | [Top $TopResources Systems by Storage Capacity Used %](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=4) |
| E-Series: Array | Capacity | timeseries | [Top $TopResources Systems by Used Space](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=5) |
///



### eseries_array_write_data

Array-wide write data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeBytesTotal` | conf/eseriesperf/11.80.0/array.yaml |

The `eseries_array_write_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Array | Highlights | timeseries | [Top $TopResources Arrays by Write Throughput](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=205) |
///



### eseries_array_write_ops

Array-wide write I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeIopsTotal` | conf/eseriesperf/11.80.0/array.yaml |

The `eseries_array_write_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Array | Highlights | timeseries | [Top $TopResources Arrays by Write IOPS](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=203) |
///



### eseries_battery_labels

This metric provides information about batteries.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml |

The `eseries_battery_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Battery | table | [Battery](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=101) |
///



### eseries_cache_backup_device_capacity

Capacity of the cache backup device in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `cacheBackupDevices.capacityInMegabytes` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_cache_backup_device_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Cache | table | [Cache Backup Devices](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=13) |
///



### eseries_cache_backup_device_labels

This metric provides information about cache backup devices.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_cache_backup_device_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Cache | table | [Cache Backup Devices](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=13) |
///



### eseries_cache_memory_dimm_capacity

Capacity of the cache memory DIMM in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `cacheMemoryDimms.capacityInMegabytes` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_cache_memory_dimm_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Cache | table | [Cache Memory DIMMs](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=12) |
///



### eseries_cache_memory_dimm_labels

This metric provides information about cache memory DIMMs.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_cache_memory_dimm_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Cache | table | [Cache Memory DIMMs](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=12) |
///



### eseries_controller_cache_hit_ops

Total number of IO operations that hit cache on the controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `cacheHitsIopsTotal` | conf/eseriesperf/11.80.0/controller.yaml |

The `eseries_controller_cache_hit_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Controller | Highlights | timeseries | [Top $TopResources Controllers by Cache Hit Ops](/d/eseries-controller/e-series3a-controller?orgId=1&viewPanel=2) |
///



### eseries_controller_code_version_labels

This metric provides information about controller code versions.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_controller_code_version_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Highlights | table | [Code Versions](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=14) |
///



### eseries_controller_cpu_utilization

Controller CPU utilization percentage


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `cpuUtilizationStats.0.sumCpuUtilization` | conf/eseriesperf/11.80.0/controller.yaml |

The `eseries_controller_cpu_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Controller | Highlights | timeseries | [Top $TopResources Controllers by CPU Utilization](/d/eseries-controller/e-series3a-controller?orgId=1&viewPanel=1) |
///



### eseries_controller_drive_interface_labels

This metric provides information about controller drive-side interfaces.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/interfaces?channelType=driveside` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_controller_drive_interface_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Drive Interfaces | table | [Drive Interfaces](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=6) |
///



### eseries_controller_host_interface_labels

This metric provides information about controller host-side interfaces.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/interfaces?channelType=hostside` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_controller_host_interface_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Host Interfaces | table | [Host Interfaces](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=5) |
///



### eseries_controller_labels

This metric provides information about controllers.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml |

The `eseries_controller_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Controller | Controller Details | table | [Controller Configuration](/d/eseries-controller/e-series3a-controller?orgId=1&viewPanel=5) |
| E-Series: Hardware | Highlights | table | [Controller Configuration](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=2) |
///



### eseries_controller_net_interface_labels

This metric provides information about controller network interfaces.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_controller_net_interface_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Management | table | [Management Ports](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=7) |
| E-Series: Hardware | Management | table | [DNS & NTP Configuration](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=8) |
///



### eseries_controller_processor_memory

Controller processor memory size in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `controllers.processorMemorySizeMiB` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_controller_processor_memory` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Controller | Controller Details | table | [Controller Cache & Memory](/d/eseries-controller/e-series3a-controller?orgId=1&viewPanel=6) |
| E-Series: Hardware | Cache | table | [Controller Cache & Memory](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=3) |
| E-Series: Hardware | Cache | timeseries | [Top $TopResources Controllers by Processor Cache](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=11) |
///



### eseries_controller_read_data

Total number of bytes read by the controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readBytesTotal` | conf/eseriesperf/11.80.0/controller.yaml |

The `eseries_controller_read_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Controller | Highlights | timeseries | [Top $TopResources Controllers by Read Throughput](/d/eseries-controller/e-series3a-controller?orgId=1&viewPanel=4) |
///



### eseries_controller_read_ops

Total number of read IO operations serviced by the controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readIopsTotal` | conf/eseriesperf/11.80.0/controller.yaml |

The `eseries_controller_read_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Controller | Highlights | timeseries | [Top $TopResources Controllers by Read IOPS](/d/eseries-controller/e-series3a-controller?orgId=1&viewPanel=3) |
///



### eseries_controller_total_cache_memory

Total cache memory on the controller in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `controllers.cacheMemorySize` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_controller_total_cache_memory` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Controller | Controller Details | table | [Controller Cache & Memory](/d/eseries-controller/e-series3a-controller?orgId=1&viewPanel=6) |
| E-Series: Hardware | Cache | table | [Controller Cache & Memory](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=3) |
| E-Series: Hardware | Cache | timeseries | [Top $TopResources Controllers by Data Cache Total](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=10) |
///



### eseries_controller_used_cache_memory

Used cache memory on the controller in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `controllers` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_controller_used_cache_memory` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Controller | Controller Details | table | [Controller Cache & Memory](/d/eseries-controller/e-series3a-controller?orgId=1&viewPanel=6) |
| E-Series: Hardware | Cache | table | [Controller Cache & Memory](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=3) |
| E-Series: Hardware | Cache | timeseries | [Top $TopResources Controllers by Data Cache Used](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=4) |
///



### eseries_controller_write_data

Total number of bytes written by the controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeBytesTotal` | conf/eseriesperf/11.80.0/controller.yaml |

The `eseries_controller_write_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Controller | Highlights | timeseries | [Top $TopResources Controllers by Write Throughput](/d/eseries-controller/e-series3a-controller?orgId=1&viewPanel=31) |
///



### eseries_controller_write_ops

Total number of write IO operations serviced by the controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeIopsTotal` | conf/eseriesperf/11.80.0/controller.yaml |

The `eseries_controller_write_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Controller | Highlights | timeseries | [Top $TopResources Controllers by Write IOPS](/d/eseries-controller/e-series3a-controller?orgId=1&viewPanel=30) |
///



### eseries_drive_block_size

Logical block size of the drive in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `drives.blkSize` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_drive_block_size` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Drives | table | [Drives](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=31) |
///



### eseries_drive_block_size_physical

Physical block size of the drive in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `drives.blkSizePhysical` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_drive_block_size_physical` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Drives | table | [Drives](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=31) |
///



### eseries_drive_capacity

Raw capacity of the drive in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `drives.rawCapacity` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_drive_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Drives | table | [Drives](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=31) |
///



### eseries_drive_labels

This metric provides information about drives.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_drive_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Drives | table | [Drives](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=31) |
///



### eseries_drive_percent_endurance_used

Percentage of SSD endurance used for solid state drives


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `drives.ssdWearLife.percentEnduranceUsed` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_drive_percent_endurance_used` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Drives | table | [Drives](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=31) |
///



### eseries_fan_labels

This metric provides information about fans.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml |

The `eseries_fan_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Fan | table | [Fan](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=103) |
///



### eseries_host_labels

This metric provides information about hosts connected to the storage array.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hosts` | `Harvest generated` | conf/eseries/11.80.0/host.yaml |

The `eseries_host_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Array | Host | table | [Host Configuration](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=15) |
///



### eseries_power_supply_labels

This metric provides information about power supplies.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml |

The `eseries_power_supply_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Power Supply | table | [Power Supply](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=105) |
///



### eseries_sfp_labels

This metric provides information about SFP transceivers.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_sfp_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | SFP | table | [SFP](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=109) |
///



### eseries_thermal_sensor_labels

This metric provides information about thermal sensors.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_thermal_sensor_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | Thermal Sensor | table | [Thermal Sensor](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=107) |
///



### eseries_volume_allocated_capacity

Allocated capacity of the volume in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/volumes` | `totalSizeInBytes` | conf/eseries/11.80.0/volume.yaml |

The `eseries_volume_allocated_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Volume Table | table | [Volumes](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=18) |
///



### eseries_volume_block_size

Block size of the volume in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/volumes` | `blkSize` | conf/eseries/11.80.0/volume.yaml |

The `eseries_volume_block_size` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Volume Table | table | [Volumes](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=18) |
///



### eseries_volume_labels

This metric provides information about volumes.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/volumes` | `Harvest generated` | conf/eseries/11.80.0/volume.yaml |

The `eseries_volume_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Volume Table | table | [Volumes](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=18) |
///



### eseries_volume_read_cache_hit_ratio

Volume read cache hit ratio calculated from read hit operations and total read operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/volume.yaml (CacheHitRatio plugin) |

The `eseries_volume_read_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Cache | timeseries | [Top $TopResources Volumes by Read Cache Hit Ratio](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=40) |
///



### eseries_volume_read_data

Volume read data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readBytes` | conf/eseriesperf/11.80.0/volume.yaml |

The `eseries_volume_read_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Highlights | timeseries | [Top $TopResources Volumes by Read Throughput](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=31) |
///



### eseries_volume_read_hit_ops

Number of read operations that hit cache


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readHitOps` | conf/eseriesperf/11.80.0/volume.yaml |


### eseries_volume_read_latency

Read response time average in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readTimeTotal` | conf/eseriesperf/11.80.0/volume.yaml |

The `eseries_volume_read_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Highlights | timeseries | [Top $TopResources Volumes by Read Latency](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=30) |
///



### eseries_volume_read_ops

Volume read I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readOps` | conf/eseriesperf/11.80.0/volume.yaml |

The `eseries_volume_read_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Highlights | timeseries | [Top $TopResources Volumes by Read IOPs](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=32) |
///



### eseries_volume_reported_capacity

The capacity in bytes of the volume


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/volumes` | `capacity` | conf/eseries/11.80.0/volume.yaml |

The `eseries_volume_reported_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Volume Table | table | [Volumes](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=18) |
///



### eseries_volume_total_cache_hit_ratio

Volume total cache hit ratio combining read and write cache hit operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/volume.yaml (CacheHitRatio plugin) |

The `eseries_volume_total_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Cache | timeseries | [Top $TopResources Volumes by Total Cache Hit Ratio](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=42) |
///



### eseries_volume_write_cache_hit_ratio

Volume write cache hit ratio calculated from write hit operations and total write operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/volume.yaml (CacheHitRatio plugin) |

The `eseries_volume_write_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Cache | timeseries | [Top $TopResources Volumes by Write Cache Hit Ratio](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=41) |
///



### eseries_volume_write_data

Volume write data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeBytes` | conf/eseriesperf/11.80.0/volume.yaml |

The `eseries_volume_write_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Highlights | timeseries | [Top $TopResources Volumes by Write Throughput](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=34) |
///



### eseries_volume_write_hit_ops

Volume write cache hit operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeHitOps` | conf/eseriesperf/11.80.0/volume.yaml |


### eseries_volume_write_latency

Write response time average in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeTimeTotal` | conf/eseriesperf/11.80.0/volume.yaml |

The `eseries_volume_write_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Highlights | timeseries | [Top $TopResources Volumes by Write Latency](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=33) |
///



### eseries_volume_write_ops

Volume write I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeOps` | conf/eseriesperf/11.80.0/volume.yaml |

The `eseries_volume_write_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Highlights | timeseries | [Top $TopResources Volumes by Write IOPs](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=35) |
///



