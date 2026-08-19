This document describes which E-Series metrics are collected and what those metrics are named in Harvest, including:

- Details about which Harvest metrics each dashboard uses.
These can be generated on demand by running `bin/harvest grafana metrics`. See
[#1577](https://github.com/NetApp/harvest/issues/1577#issue-1471478260) for details.

```
Creation Date : 2026-Aug-17
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


### eseries_application_other_ops

Application other I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `otherOps` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_other_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Highlights | timeseries | [Top $TopResources Applications by Other IOPs](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=8) |
///



### eseries_application_queue_depth_average

Average queue depth per I/O operation


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_queue_depth_average` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Queue Depth | timeseries | [Top $TopResources Applications by Queue Depth Average](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=14) |
///



### eseries_application_queue_depth_max

Maximum queue depth seen over the observation window


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `queueDepthMax` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_queue_depth_max` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Queue Depth | timeseries | [Top $TopResources Applications by Queue Depth Max](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=15) |
///



### eseries_application_read_cache_hit_ratio

Application read cache hit ratio calculated from read hit operations and total read operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/application.yaml (CacheHitRatio plugin) |

The `eseries_application_read_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Cache | timeseries | [Top $TopResources Applications by Read Cache Hit Ratio](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=10) |
///



### eseries_application_read_data

Application read data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readBytes` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_read_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Highlights | timeseries | [Top $TopResources Applications by Read Throughput](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=6) |
///



### eseries_application_read_hit_ops

Number of read operations that hit cache


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readHitOps` | conf/eseriesperf/11.80.0/application.yaml |


### eseries_application_read_latency

Read response time average in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readTimeTotal` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_read_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Highlights | timeseries | [Top $TopResources Applications by Read Latency](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=2) |
///



### eseries_application_read_ops

Application read I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readOps` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_read_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Highlights | timeseries | [Top $TopResources Applications by Read IOPs](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=4) |
| E-Series: Application | Application Table | table | [Applications](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=21) |
///



### eseries_application_read_utilization

Percentage of the observation window the application spent servicing read I/O


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_read_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Utilization | timeseries | [Top $TopResources Applications by Read Utilization](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=17) |
///



### eseries_application_total_cache_hit_ratio

Application total cache hit ratio combining read and write cache hit operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/application.yaml (CacheHitRatio plugin) |

The `eseries_application_total_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Cache | timeseries | [Top $TopResources Applications by Total Cache Hit Ratio](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=12) |
///



### eseries_application_total_utilization

Percentage of the observation window the application spent servicing I/O (read + write)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_total_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Utilization | timeseries | [Top $TopResources Applications by Total Utilization](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=19) |
///



### eseries_application_write_cache_hit_ratio

Application write cache hit ratio calculated from write hit operations and total write operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/application.yaml (CacheHitRatio plugin) |

The `eseries_application_write_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Cache | timeseries | [Top $TopResources Applications by Write Cache Hit Ratio](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=11) |
///



### eseries_application_write_data

Application write data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeBytes` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_write_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Highlights | timeseries | [Top $TopResources Applications by Write Throughput](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=7) |
///



### eseries_application_write_hit_ops

Application write cache hit operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeHitOps` | conf/eseriesperf/11.80.0/application.yaml |


### eseries_application_write_latency

Write response time average in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeTimeTotal` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_write_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Highlights | timeseries | [Top $TopResources Applications by Write Latency](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=3) |
///



### eseries_application_write_ops

Application write I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeOps` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_write_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Highlights | timeseries | [Top $TopResources Applications by Write IOPs](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=5) |
///



### eseries_application_write_utilization

Percentage of the observation window the application spent servicing write I/O


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/application.yaml |

The `eseries_application_write_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Application | Utilization | timeseries | [Top $TopResources Applications by Write Utilization](/d/eseries-application/e-series3a-application?orgId=1&viewPanel=18) |
///



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
| REST | `storage-systems/{array_id}/hardware-inventory` | `cacheBackupDevices.backupDeviceCapacity` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

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
| REST | `storage-systems/{array_id}/hardware-inventory` | `controllers.processorMemorySize` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

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
| REST | `storage-systems/{array_id}/hardware-inventory` | `controllers.physicalCacheMemorySize` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

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
| REST | `storage-systems/{array_id}/hardware-inventory` | `controllers.cacheMemorySize` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

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
| E-Series: Drive | Drive Details | table | [Drives](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=29) |
| E-Series: Hardware | Drives | table | [Drives](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=112) |
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
| E-Series: Drive | Drive Details | table | [Drives](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=29) |
| E-Series: Hardware | Drives | table | [Drives](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=112) |
///



### eseries_drive_capacity

Usable capacity of the drive in bytes, after accounting for space reserved by the array controller for overhead


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `drives.usableCapacity` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_drive_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Drive | Drive Details | table | [Drives](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=29) |
| E-Series: Hardware | Drives | table | [Drives](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=112) |
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
| E-Series: Drive | Drive Details | table | [Drives](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=29) |
| E-Series: Hardware | Drives | table | [Drives](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=112) |
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
| E-Series: Drive | Drive Details | table | [Drives](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=29) |
| E-Series: Hardware | Drives | table | [Drives](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=112) |
///



### eseries_drive_raw_capacity

Raw physical capacity of the drive in bytes, before array controller overhead is reserved


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `drives.rawCapacity` | conf/eseries/11.80.0/hardware.yaml (Hardware plugin) |

The `eseries_drive_raw_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Drive | Drive Details | table | [Drives](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=29) |
| E-Series: Hardware | Drives | table | [Drives](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=112) |
///



### eseries_drive_read_data

Drive read throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readBytes` | conf/eseriesperf/11.80.0/drive.yaml |

The `eseries_drive_read_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Drive | Highlights | timeseries | [Top $TopResources Drives by Read Throughput](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=5) |
///



### eseries_drive_read_latency

Average drive read latency in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readTimeTotal` | conf/eseriesperf/11.80.0/drive.yaml |

The `eseries_drive_read_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Drive | Highlights | timeseries | [Top $TopResources Drives by Read Latency](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=1) |
///



### eseries_drive_read_ops

Drive read I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readOps` | conf/eseriesperf/11.80.0/drive.yaml |

The `eseries_drive_read_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Drive | Highlights | timeseries | [Top $TopResources Drives by Read IOPs](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=3) |
///



### eseries_drive_read_utilization

Percentage of time the drive was busy servicing read I/O requests


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest generated` | conf/eseriesperf/11.80.0/drive.yaml |

The `eseries_drive_read_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Drive | Utilization | timeseries | [Top $TopResources Drives by Read Utilization](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=8) |
///



### eseries_drive_total_utilization

Percentage of time the drive was busy servicing all I/O requests (read and write combined)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest generated` | conf/eseriesperf/11.80.0/drive.yaml |

The `eseries_drive_total_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Drive | Utilization | timeseries | [Top $TopResources Drives by Total Utilization](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=10) |
///



### eseries_drive_write_data

Drive write throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeBytes` | conf/eseriesperf/11.80.0/drive.yaml |

The `eseries_drive_write_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Drive | Highlights | timeseries | [Top $TopResources Drives by Write Throughput](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=6) |
///



### eseries_drive_write_latency

Average drive write latency in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeTimeTotal` | conf/eseriesperf/11.80.0/drive.yaml |

The `eseries_drive_write_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Drive | Highlights | timeseries | [Top $TopResources Drives by Write Latency](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=2) |
///



### eseries_drive_write_ops

Drive write I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeOps` | conf/eseriesperf/11.80.0/drive.yaml |

The `eseries_drive_write_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Drive | Highlights | timeseries | [Top $TopResources Drives by Write IOPs](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=4) |
///



### eseries_drive_write_utilization

Percentage of time the drive was busy servicing write I/O requests


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest generated` | conf/eseriesperf/11.80.0/drive.yaml |

The `eseries_drive_write_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Drive | Utilization | timeseries | [Top $TopResources Drives by Write Utilization](/d/eseries-drive/e-series3a-drive?orgId=1&viewPanel=9) |
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



### eseries_firmware_version_labels

This metric provides information about the array-wide firmware and software code versions (raid, management, iom, bundle, etc.).


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `firmware/embedded-firmware/{array_id}/versions` | `Harvest generated` | conf/eseries/11.80.0/firmware.yaml (Firmware plugin) |

The `eseries_firmware_version_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Array | Firmware Versions | table | [Firmware Versions](/d/eseries-array/e-series3a-array?orgId=1&viewPanel=18) |
///



### eseries_host_board_labels

This metric provides information about IO modules (host interface cards) installed in the controllers.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `Harvest generated` | conf/eseries/11.80.0/hardware.yaml |

The `eseries_host_board_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Hardware | IO Module / Host Board | table | [IO Module / Host Board](/d/eseries-hardware/e-series3a-hardware?orgId=1&viewPanel=111) |
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



### eseries_interface_channel_error_count

Number of errors detected on the channel


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `channelErrorCount` | conf/eseriesperf/11.80.0/interface.yaml |

The `eseries_interface_channel_error_count` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Interface | Channel Errors | timeseries | [Top $TopResources Interfaces by Channel Error Count](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=14) |
///



### eseries_interface_interface

Friendly interface/port name (physicalLocation.label) resolved from hardware-inventory


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/hardware-inventory` | `physicalLocation.label` | conf/eseriesperf/11.80.0/interface.yaml (Interface plugin) |


### eseries_interface_other_latency

Other command response time average in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `otherTimeTotal` | conf/eseriesperf/11.80.0/interface.yaml |

The `eseries_interface_other_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Interface | Highlights | timeseries | [Top $TopResources Interfaces by Other Latency](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=9) |
///



### eseries_interface_other_ops

Interface other I/O operations per second (e.g. control/management commands)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `otherOps` | conf/eseriesperf/11.80.0/interface.yaml |

The `eseries_interface_other_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Interface | Highlights | timeseries | [Top $TopResources Interfaces by Other IOPs](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=8) |
///



### eseries_interface_queue_depth_average

Average queue depth per I/O operation


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/interface.yaml |

The `eseries_interface_queue_depth_average` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Interface | Queue Depth | timeseries | [Top $TopResources Interfaces by Queue Depth Average](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=11) |
///



### eseries_interface_queue_depth_max

Maximum queue depth seen over the observation window


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `queueDepthMax` | conf/eseriesperf/11.80.0/interface.yaml |

The `eseries_interface_queue_depth_max` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Interface | Queue Depth | timeseries | [Top $TopResources Interfaces by Queue Depth Max](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=12) |
///



### eseries_interface_read_data

Interface read data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readBytes` | conf/eseriesperf/11.80.0/interface.yaml |

The `eseries_interface_read_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Interface | Highlights | timeseries | [Top $TopResources Interfaces by Read Throughput](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=6) |
///



### eseries_interface_read_latency

Read response time average in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readTimeTotal` | conf/eseriesperf/11.80.0/interface.yaml |

The `eseries_interface_read_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Interface | Highlights | timeseries | [Top $TopResources Interfaces by Read Latency](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=2) |
///



### eseries_interface_read_ops

Interface read I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readOps` | conf/eseriesperf/11.80.0/interface.yaml |

The `eseries_interface_read_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Interface | Highlights | timeseries | [Top $TopResources Interfaces by Read IOPs](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=4) |
| E-Series: Interface | Interface Table | table | [Interfaces](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=16) |
///



### eseries_interface_write_data

Interface write data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeBytes` | conf/eseriesperf/11.80.0/interface.yaml |

The `eseries_interface_write_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Interface | Highlights | timeseries | [Top $TopResources Interfaces by Write Throughput](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=7) |
///



### eseries_interface_write_latency

Write response time average in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeTimeTotal` | conf/eseriesperf/11.80.0/interface.yaml |

The `eseries_interface_write_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Interface | Highlights | timeseries | [Top $TopResources Interfaces by Write Latency](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=3) |
///



### eseries_interface_write_ops

Interface write I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeOps` | conf/eseriesperf/11.80.0/interface.yaml |

The `eseries_interface_write_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Interface | Highlights | timeseries | [Top $TopResources Interfaces by Write IOPs](/d/eseries-interface/e-series3a-interface?orgId=1&viewPanel=5) |
///



### eseries_pool_block_size

Recommended block size of the storage pool in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `blkSizeRecommended` | conf/eseries/11.80.0/pool.yaml |

The `eseries_pool_block_size` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Pool Table | table | [Pools](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=21) |
///



### eseries_pool_block_sizes_supported

Volume block sizes supported by the storage pool (e.g. 512,4096)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `blkSizeSupported` | conf/eseries/11.80.0/pool.yaml (Pool plugin) |


### eseries_pool_da_capable

Whether all drives in the storage pool are Data Assurance (T10 PI) capable


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `protectionInformationCapabilities.protectionInformationCapable` | conf/eseries/11.80.0/pool.yaml |


### eseries_pool_drive_media_type

Media type of the drives that make up the storage pool (e.g. ssd, hdd)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `driveMediaType` | conf/eseries/11.80.0/pool.yaml |


### eseries_pool_drive_physical_type

Physical type of the drives that make up the storage pool (e.g. nvme4k)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `drivePhysicalType` | conf/eseries/11.80.0/pool.yaml |


### eseries_pool_free_capacity

Free capacity of the storage pool in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `freeSpace` | conf/eseries/11.80.0/pool.yaml |

The `eseries_pool_free_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Pool Table | table | [Pools](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=21) |
///



### eseries_pool_labels

This metric provides information about storage pools (volume groups and disk pools).


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `Harvest generated` | conf/eseries/11.80.0/pool.yaml |

The `eseries_pool_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Pool Table | table | [Pools](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=21) |
///



### eseries_pool_other_ops

Pool other I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `otherOps` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_other_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Highlights | timeseries | [Top $TopResources Pools by Other IOPs](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=8) |
///



### eseries_pool_pool

Friendly storage pool name resolved from storage-pools, attached directly to every pool performance metric


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `name` | conf/eseriesperf/11.80.0/pool.yaml (Pool plugin) |


### eseries_pool_queue_depth_average

Average queue depth per I/O operation


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_queue_depth_average` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Queue Depth | timeseries | [Top $TopResources Pools by Queue Depth Average](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=14) |
///



### eseries_pool_queue_depth_max

Maximum queue depth seen over the observation window


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `queueDepthMax` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_queue_depth_max` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Queue Depth | timeseries | [Top $TopResources Pools by Queue Depth Max](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=15) |
///



### eseries_pool_raid_level

RAID level of the storage pool


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `raidLevel` | conf/eseries/11.80.0/pool.yaml |


### eseries_pool_read_cache_hit_ratio

Pool read cache hit ratio calculated from read hit operations and total read operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/pool.yaml (CacheHitRatio plugin) |

The `eseries_pool_read_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Cache | timeseries | [Top $TopResources Pools by Read Cache Hit Ratio](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=10) |
///



### eseries_pool_read_data

Pool read data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readBytes` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_read_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Highlights | timeseries | [Top $TopResources Pools by Read Throughput](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=6) |
///



### eseries_pool_read_hit_ops

Number of read operations that hit cache


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readHitOps` | conf/eseriesperf/11.80.0/pool.yaml |


### eseries_pool_read_latency

Read response time average in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readTimeTotal` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_read_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Highlights | timeseries | [Top $TopResources Pools by Read Latency](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=2) |
///



### eseries_pool_read_ops

Pool read I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readOps` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_read_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Highlights | timeseries | [Top $TopResources Pools by Read IOPs](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=4) |
///



### eseries_pool_read_utilization

Percentage of the observation window the pool spent servicing read I/O


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_read_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Utilization | timeseries | [Top $TopResources Pools by Read Utilization](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=17) |
///



### eseries_pool_security_level

Security level of the storage pool (e.g. fde)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `securityLevel` | conf/eseries/11.80.0/pool.yaml |


### eseries_pool_security_type

Security capability type of the storage pool (e.g. capable)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `securityType` | conf/eseries/11.80.0/pool.yaml |


### eseries_pool_shelf_loss_protection

Whether the storage pool is protected against the loss of an entire drive shelf/tray


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `trayLossProtection` | conf/eseries/11.80.0/pool.yaml |


### eseries_pool_status

RAID status of the storage pool


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/storage-pools` | `raidStatus` | conf/eseries/11.80.0/pool.yaml |


### eseries_pool_total_cache_hit_ratio

Pool total cache hit ratio combining read and write cache hit operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/pool.yaml (CacheHitRatio plugin) |

The `eseries_pool_total_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Cache | timeseries | [Top $TopResources Pools by Total Cache Hit Ratio](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=12) |
///



### eseries_pool_total_utilization

Percentage of the observation window the pool spent servicing I/O (read + write)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_total_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Utilization | timeseries | [Top $TopResources Pools by Total Utilization](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=19) |
///



### eseries_pool_write_cache_hit_ratio

Pool write cache hit ratio calculated from write hit operations and total write operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/pool.yaml (CacheHitRatio plugin) |

The `eseries_pool_write_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Cache | timeseries | [Top $TopResources Pools by Write Cache Hit Ratio](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=11) |
///



### eseries_pool_write_data

Pool write data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeBytes` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_write_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Highlights | timeseries | [Top $TopResources Pools by Write Throughput](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=7) |
///



### eseries_pool_write_hit_ops

Pool write cache hit operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeHitOps` | conf/eseriesperf/11.80.0/pool.yaml |


### eseries_pool_write_latency

Write response time average in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeTimeTotal` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_write_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Highlights | timeseries | [Top $TopResources Pools by Write Latency](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=3) |
///



### eseries_pool_write_ops

Pool write I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeOps` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_write_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Highlights | timeseries | [Top $TopResources Pools by Write IOPs](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=5) |
///



### eseries_pool_write_utilization

Percentage of the observation window the pool spent servicing write I/O


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/pool.yaml |

The `eseries_pool_write_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Pool | Utilization | timeseries | [Top $TopResources Pools by Write Utilization](/d/eseries-pool/e-series3a-pool?orgId=1&viewPanel=18) |
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



### eseries_ssd_cache_additional_capacity

Additional SSD cache capacity that can still be added to the array in bytes (maximum capacity minus current capacity)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/capabilities` | `Harvest generated` | conf/eseries/11.80.0/ssd_cache.yaml (SsdCacheCapacity plugin) |

The `eseries_ssd_cache_additional_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Capacity | table | [SSD Cache Overview](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=11) |
| E-Series: SSD Cache | Capacity | timeseries | [Top $TopResources Additional Capacity Allowed](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=18) |
///



### eseries_ssd_cache_allocated_size

Allocated size of the SSD cache per controller in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.allocatedBytes` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_allocated_size` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Capacity | timeseries | [Top $TopResources Allocated Size](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=13) |
///



### eseries_ssd_cache_allocation_percent

SSD cache allocation percentage per controller, calculated as allocated bytes divided by total cache size (allocated + available bytes)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `Harvest generated` | conf/eseriesperf/11.80.0/ssd_cache.yaml (SsdCacheStats plugin) |

The `eseries_ssd_cache_allocation_percent` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Highlights | timeseries | [Top $TopResources Cache Allocation %](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=83) |
| E-Series: SSD Cache | Capacity | timeseries | [Top $TopResources Allocation %](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=15) |
///



### eseries_ssd_cache_available_size

Available size of the SSD cache per controller in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.availableBytes` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_available_size` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Capacity | timeseries | [Top $TopResources Available Size](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=14) |
///



### eseries_ssd_cache_complete_cache_miss_block_ops

Number of complete cache miss block operations per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.completeCacheMissBlocks` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_complete_cache_miss_block_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Block Operations | timeseries | [Top $TopResources Cache Miss Block Ops/s](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=65) |
///



### eseries_ssd_cache_complete_cache_miss_ops

Number of complete cache miss operations per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.completeCacheMiss` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_complete_cache_miss_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | IOPS | timeseries | [Top $TopResources Cache Miss IOPS](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=55) |
///



### eseries_ssd_cache_complete_cache_miss_percent

Percentage of read operations that resulted in a complete cache miss per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `Harvest generated` | conf/eseriesperf/11.80.0/ssd_cache.yaml (SsdCacheStats plugin) |

The `eseries_ssd_cache_complete_cache_miss_percent` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Cache Hit Performance | timeseries | [Top $TopResources Complete Cache Miss %](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=44) |
///



### eseries_ssd_cache_current_capacity

Current used capacity of the SSD cache in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches` | `usedCapacity` | conf/eseries/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_current_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Capacity | table | [SSD Cache Overview](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=11) |
| E-Series: SSD Cache | Capacity | timeseries | [Top $TopResources Current Capacity](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=12) |
///



### eseries_ssd_cache_drive_labels

This metric provides information about drives assigned to an SSD cache.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/drives` | `Harvest generated` | conf/eseries/11.80.0/ssd_cache.yaml (SsdCacheCapacity plugin) |

The `eseries_ssd_cache_drive_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Drives | table | [SSD Cache Drives](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=21) |
///



### eseries_ssd_cache_drive_raw_capacity

Raw capacity of each drive contributing to the SSD cache in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/drives` | `rawCapacity` | conf/eseries/11.80.0/ssd_cache.yaml (SsdCacheCapacity plugin) |

The `eseries_ssd_cache_drive_raw_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Drives | table | [SSD Cache Drives](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=21) |
///



### eseries_ssd_cache_full_cache_hit_block_ops

Number of full cache hit block operations per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.fullCacheHitBlocks` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_full_cache_hit_block_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Block Operations | timeseries | [Top $TopResources Full Cache Hit Block Ops/s](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=63) |
///



### eseries_ssd_cache_full_cache_hit_ops

Number of full cache hit operations per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.fullCacheHits` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_full_cache_hit_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | IOPS | timeseries | [Top $TopResources Full Cache Hit IOPS](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=53) |
///



### eseries_ssd_cache_full_cache_hit_percent

Percentage of read operations that resulted in a full cache hit per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `Harvest generated` | conf/eseriesperf/11.80.0/ssd_cache.yaml (SsdCacheStats plugin) |

The `eseries_ssd_cache_full_cache_hit_percent` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Cache Hit Performance | timeseries | [Top $TopResources Full Cache Hit %](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=42) |
///



### eseries_ssd_cache_hit_percent

SSD cache hit percentage per controller, calculated as full cache hits divided by total I/O operations (reads + writes)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `Harvest generated` | conf/eseriesperf/11.80.0/ssd_cache.yaml (SsdCacheStats plugin) |

The `eseries_ssd_cache_hit_percent` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Highlights | timeseries | [Top $TopResources Cache Hit %](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=2) |
| E-Series: SSD Cache | Cache Hit Performance | timeseries | [Top $TopResources Cache Hit %](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=41) |
///



### eseries_ssd_cache_invalidate_ops

Number of cache invalidate operations per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.invalidates` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_invalidate_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Cache Sizing | timeseries | [Top $TopResources Invalidate Ops/s](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=73) |
///



### eseries_ssd_cache_labels

This metric provides information about SSD caches.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches` | `Harvest generated` | conf/eseries/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Capacity | table | [SSD Cache Overview](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=11) |
///



### eseries_ssd_cache_max_capacity

Maximum SSD cache capacity allowed for the array in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/capabilities` | `featureParameters.maxFlashCacheSize` | conf/eseries/11.80.0/ssd_cache.yaml (SsdCacheCapacity plugin) |

The `eseries_ssd_cache_max_capacity` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Capacity | table | [SSD Cache Overview](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=11) |
| E-Series: SSD Cache | Capacity | timeseries | [Top $TopResources Maximum Capacity Allowed](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=17) |
///



### eseries_ssd_cache_partial_cache_hit_block_ops

Number of partial cache hit block operations per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.partialCacheHitBlocks` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_partial_cache_hit_block_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Block Operations | timeseries | [Top $TopResources Partial Cache Hit Block Ops/s](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=64) |
///



### eseries_ssd_cache_partial_cache_hit_ops

Number of partial cache hit operations per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.partialCacheHits` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_partial_cache_hit_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | IOPS | timeseries | [Top $TopResources Partial Cache Hit IOPS](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=54) |
///



### eseries_ssd_cache_partial_cache_hit_percent

Percentage of read operations that resulted in a partial cache hit per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `Harvest generated` | conf/eseriesperf/11.80.0/ssd_cache.yaml (SsdCacheStats plugin) |

The `eseries_ssd_cache_partial_cache_hit_percent` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Cache Hit Performance | timeseries | [Top $TopResources Partial Cache Hit %](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=43) |
///



### eseries_ssd_cache_populate_on_read_ops

Number of populate-on-read operations per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.populateOnReads` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_populate_on_read_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | IOPS | timeseries | [Top $TopResources Populate-on-Read IOPS](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=56) |
///



### eseries_ssd_cache_populate_on_write_ops

Number of populate-on-write operations per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.populateOnWrites` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_populate_on_write_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | IOPS | timeseries | [Top $TopResources Populate-on-Write IOPS](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=57) |
///



### eseries_ssd_cache_populated_clean_size

Amount of clean (unmodified) data populated in the SSD cache per controller in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.populatedCleanBytes` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_populated_clean_size` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Cache Sizing | timeseries | [Top $TopResources Populated Clean Size](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=71) |
///



### eseries_ssd_cache_populated_dirty_size

Amount of dirty (modified) data populated in the SSD cache per controller in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.populatedDirtyBytes` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_populated_dirty_size` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Cache Sizing | timeseries | [Top $TopResources Populated Dirty Size](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=72) |
///



### eseries_ssd_cache_read_block_ops

SSD cache read block operations per second per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.readBlocks` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_read_block_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Block Operations | timeseries | [Top $TopResources Read Block Ops/s](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=61) |
///



### eseries_ssd_cache_read_ops

SSD cache read operations per second per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.reads` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_read_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | IOPS | timeseries | [Top $TopResources Read IOPS](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=51) |
///



### eseries_ssd_cache_recycle_ops

Number of cache recycle operations per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.recycles` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_recycle_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Cache Sizing | timeseries | [Top $TopResources Recycle Ops/s](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=74) |
///



### eseries_ssd_cache_utilization_percent

SSD cache utilization percentage per controller, calculated as populated data (clean + dirty bytes) divided by allocated bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `Harvest generated` | conf/eseriesperf/11.80.0/ssd_cache.yaml (SsdCacheStats plugin) |

The `eseries_ssd_cache_utilization_percent` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Highlights | timeseries | [Top $TopResources Cache Utilization %](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=84) |
| E-Series: SSD Cache | Capacity | timeseries | [Top $TopResources Utilization %](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=16) |
///



### eseries_ssd_cache_volume_labels

This metric provides information about volumes mapped to an SSD cache.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/volumes` | `Harvest generated` | conf/eseries/11.80.0/ssd_cache.yaml (SsdCacheCapacity plugin) |

The `eseries_ssd_cache_volume_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Cached Volumes | table | [Cached Volumes](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=31) |
///



### eseries_ssd_cache_write_block_ops

SSD cache write block operations per second per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.writeBlocks` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_write_block_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | Block Operations | timeseries | [Top $TopResources Write Block Ops/s](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=62) |
///



### eseries_ssd_cache_write_ops

SSD cache write operations per second per controller


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/ssd-caches/{ssd_cache_id}/statistics` | `statistics.writes` | conf/eseriesperf/11.80.0/ssd_cache.yaml |

The `eseries_ssd_cache_write_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: SSD Cache | IOPS | timeseries | [Top $TopResources Write IOPS](/d/eseries-ssd-cache/e-series3a-ssd cache?orgId=1&viewPanel=52) |
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

Allocated capacity of the volume in bytes. Reports the same value as eseries_volume_reported_capacity


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/volumes` | `totalSizeInBytes` | conf/eseries/11.80.0/volume.yaml |


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



### eseries_volume_other_ops

Volume other I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `otherOps` | conf/eseriesperf/11.80.0/volume.yaml |

The `eseries_volume_other_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Highlights | timeseries | [Top $TopResources Volumes by Other IOPs](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=46) |
///



### eseries_volume_queue_depth_average

Average queue depth per I/O operation


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/volume.yaml |

The `eseries_volume_queue_depth_average` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Queue Depth | timeseries | [Top $TopResources Volumes by Queue Depth Average](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=44) |
///



### eseries_volume_queue_depth_max

Maximum queue depth seen over the observation window


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `queueDepthMax` | conf/eseriesperf/11.80.0/volume.yaml |

The `eseries_volume_queue_depth_max` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Volume | Queue Depth | timeseries | [Top $TopResources Volumes by Queue Depth Max](/d/eseries-volume/e-series3a-volume?orgId=1&viewPanel=45) |
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

The capacity in bytes of the volume. Reports the same value as eseries_volume_allocated_capacity


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



### eseries_workload_labels

This metric provides information about workloads.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/workloads` | `Harvest generated` | conf/eseries/11.80.0/workload.yaml |

The `eseries_workload_labels` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Workload Table | table | [Workloads](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=21) |
///



### eseries_workload_other_ops

Workload other I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `otherOps` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_other_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Highlights | timeseries | [Top $TopResources Workloads by Other IOPs](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=8) |
///



### eseries_workload_profile_id

Workload profile identifier assigned to the workload (links to the corresponding Application object)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/workloads` | `workloadAttributes.0.value` | conf/eseries/11.80.0/workload.yaml |


### eseries_workload_queue_depth_average

Average queue depth per I/O operation


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_queue_depth_average` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Queue Depth | timeseries | [Top $TopResources Workloads by Queue Depth Average](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=14) |
///



### eseries_workload_queue_depth_max

Maximum queue depth seen over the observation window


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `queueDepthMax` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_queue_depth_max` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Queue Depth | timeseries | [Top $TopResources Workloads by Queue Depth Max](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=15) |
///



### eseries_workload_read_cache_hit_ratio

Workload read cache hit ratio calculated from read hit operations and total read operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/workload.yaml (CacheHitRatio plugin) |

The `eseries_workload_read_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Cache | timeseries | [Top $TopResources Workloads by Read Cache Hit Ratio](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=10) |
///



### eseries_workload_read_data

Workload read data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readBytes` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_read_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Highlights | timeseries | [Top $TopResources Workloads by Read Throughput](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=6) |
///



### eseries_workload_read_hit_ops

Number of read operations that hit cache


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readHitOps` | conf/eseriesperf/11.80.0/workload.yaml |


### eseries_workload_read_latency

Read response time average in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readTimeTotal` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_read_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Highlights | timeseries | [Top $TopResources Workloads by Read Latency](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=2) |
///



### eseries_workload_read_ops

Workload read I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `readOps` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_read_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Highlights | timeseries | [Top $TopResources Workloads by Read IOPs](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=4) |
///



### eseries_workload_read_utilization

Percentage of the observation window the workload spent servicing read I/O


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_read_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Utilization | timeseries | [Top $TopResources Workloads by Read Utilization](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=17) |
///



### eseries_workload_total_cache_hit_ratio

Workload total cache hit ratio combining read and write cache hit operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/workload.yaml (CacheHitRatio plugin) |

The `eseries_workload_total_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Cache | timeseries | [Top $TopResources Workloads by Total Cache Hit Ratio](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=12) |
///



### eseries_workload_total_utilization

Percentage of the observation window the workload spent servicing I/O (read + write)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_total_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Utilization | timeseries | [Top $TopResources Workloads by Total Utilization](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=19) |
///



### eseries_workload_workload

Friendly workload name resolved from workloads, attached directly to every workload performance metric


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/workloads` | `name` | conf/eseriesperf/11.80.0/workload.yaml (Workload plugin) |


### eseries_workload_write_cache_hit_ratio

Workload write cache hit ratio calculated from write hit operations and total write operations


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/workload.yaml (CacheHitRatio plugin) |

The `eseries_workload_write_cache_hit_ratio` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Cache | timeseries | [Top $TopResources Workloads by Write Cache Hit Ratio](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=11) |
///



### eseries_workload_write_data

Workload write data throughput in bytes per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeBytes` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_write_data` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Highlights | timeseries | [Top $TopResources Workloads by Write Throughput](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=7) |
///



### eseries_workload_write_hit_ops

Workload write cache hit operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeHitOps` | conf/eseriesperf/11.80.0/workload.yaml |


### eseries_workload_write_latency

Write response time average in microseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeTimeTotal` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_write_latency` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Highlights | timeseries | [Top $TopResources Workloads by Write Latency](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=3) |
///



### eseries_workload_write_ops

Workload write I/O operations per second


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `writeOps` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_write_ops` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Highlights | timeseries | [Top $TopResources Workloads by Write IOPs](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=5) |
///



### eseries_workload_write_utilization

Percentage of the observation window the workload spent servicing write I/O


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `storage-systems/{array_id}/live-statistics` | `Harvest Generated` | conf/eseriesperf/11.80.0/workload.yaml |

The `eseries_workload_write_utilization` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| E-Series: Workload | Utilization | timeseries | [Top $TopResources Workloads by Write Utilization](/d/eseries-workload/e-series3a-workload?orgId=1&viewPanel=18) |
///



