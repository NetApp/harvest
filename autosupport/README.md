# Harvest Autosupport

By default, Harvest sends basic poller information to NetApp on a daily
cadence. This behavior can be disabled by adding `autosupport_disabled: true` to
the `Tools` section of your `harvest.yml` file. e.g.

```
Tools:
  autosupport_disabled: true
```

This information is collected and sent via the `autosupport/asup` executable. Harvest
autosupport does not gather or transmit Personally Identifiable Information
(PII) or Personal Information. `autosupport/asup` comes with a [end user license
agreement](NetApp-EULA.md) that is not applicable to the other Harvest binaries
in `./bin`. This EULA only applies to `autosupport/asup`. 

You can learn more about NetApp's commitment to data security and trust
[here](https://www.netapp.com/us/company/trust-center/index.aspx).

## Example Autosupport Information

An example payload sent by Harvest looks like this. You can see exactly what Harvest sends by checking the `./asup/payload/` directory.

```
{
  "Target": {
    "Version": "9.10.1",
    "Model": "cdot",
    "Serial": "1-80-000011",
    "Ping": 0,
    "ClusterUuid": "37387241-8b57-11e9-8974-00a098e0219a"
  },
  "Harvest": {
    "HostHash": "df62c133cbd0fef8ccda100e3c04c33eb3b2d911",
    "UUID": "fe19da98aea5c555d95292170e7251e037819ee6",
    "Version": "2.0.2",
    "Release": "rc2",
    "Commit": "HEAD",
    "BuildDate": "undefined",
    "NumClusters": 1
  },
  "Platform": {
    "OS": "darwin",
    "Arch": "darwin",
    "Memory": {
      "TotalKb": 33554432,
      "AvailableKb": 13753660,
      "UsedKb": 19800772
    },
    "CPUs": 1
  },
  "Nodes": {
    "Count": 2,
    "DataPoints": 33,
    "PollTime": 225683,
    "ApiTime": 225158,
    "ParseTime": 232,
    "PluginTime": 5,
    "Uuids": [
      "95f94e8d-8b4e-11e9-8974-00a098e0219a",
      "e0362156-8b4d-11e9-b263-00a098e0060e"
    ]
  },
  "Volumes": {
    "Count": 462,
    "DataPoints": 16618,
    "PollTime": 2151647,
    "ApiTime": 632082,
    "ParseTime": 66963,
    "PluginTime": 2309
  },
  "Collectors": [
    {
      "Name": "Zapi",
      "Query": "system-node-get-iter",
      "BatchSize": "500",
      "ClientTimeout": "180",
      "Schedules": [
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "180s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 17,
        "List": [
          "node-details-info cpu-busytime",
          "node-details-info env-failed-fan-count",
          "node-details-info env-failed-fan-message",
          "node-details-info env-failed-power-supply-count",
          "node-details-info env-failed-power-supply-message",
          "node-details-info env-over-temperature",
          "node-details-info is-node-healthy",
          "node-details-info maximum-aggregate-size",
          "node-details-info maximum-number-of-volumes",
          "node-details-info maximum-volume-size",
          "node-details-info node",
          "node-details-info node-location",
          "node-details-info node-model",
          "node-details-info node-serial-number",
          "node-details-info node-uptime",
          "node-details-info node-vendor",
          "node-details-info product-version"
        ]
      }
    },
    {
      "Name": "Zapi",
      "Query": "aggr-get-iter",
      "BatchSize": "500",
      "ClientTimeout": "180",
      "Schedules": [
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "180s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 46,
        "List": [
          "aggr-attributes aggregate-name",
          "aggr-attributes aggregate-uuid",
          "aggr-attributes aggr-ownership-attributes home-name",
          "aggr-attributes aggr-raid-attributes aggregate-type",
          "aggr-attributes aggr-raid-attributes disk-count",
          "aggr-attributes aggr-raid-attributes plex-count",
          "aggr-attributes aggr-raid-attributes raid-size",
          "aggr-attributes aggr-raid-attributes state",
          "aggr-attributes aggr-inode-attributes files-private-used",
          "aggr-attributes aggr-inode-attributes files-total",
          "aggr-attributes aggr-inode-attributes files-used",
          "aggr-attributes aggr-inode-attributes inodefile-private-capacity",
          "aggr-attributes aggr-inode-attributes inodefile-public-capacity",
          "aggr-attributes aggr-inode-attributes maxfiles-available",
          "aggr-attributes aggr-inode-attributes maxfiles-possible",
          "aggr-attributes aggr-inode-attributes maxfiles-used",
          "aggr-attributes aggr-inode-attributes percent-inode-used-capacity",
          "aggr-attributes aggr-space-attributes capacity-tier-used",
          "aggr-attributes aggr-space-attributes data-compacted-count",
          "aggr-attributes aggr-space-attributes data-compaction-space-saved",
          "aggr-attributes aggr-space-attributes data-compaction-space-saved-percent",
          "aggr-attributes aggr-space-attributes hybrid-cache-size-total",
          "aggr-attributes aggr-space-attributes percent-used-capacity",
          "aggr-attributes aggr-space-attributes performance-tier-inactive-user-data",
          "aggr-attributes aggr-space-attributes performance-tier-inactive-user-data-percent",
          "aggr-attributes aggr-space-attributes physical-used",
          "aggr-attributes aggr-space-attributes physical-used-percent",
          "aggr-attributes aggr-space-attributes sis-shared-count",
          "aggr-attributes aggr-space-attributes sis-space-saved",
          "aggr-attributes aggr-space-attributes sis-space-saved-percent",
          "aggr-attributes aggr-space-attributes size-available",
          "aggr-attributes aggr-space-attributes size-total",
          "aggr-attributes aggr-space-attributes size-used",
          "aggr-attributes aggr-space-attributes total-reserved-space",
          "aggr-attributes aggr-volume-count-attributes flexvol-count",
          "aggr-attributes aggr-snapshot-attributes files-total",
          "aggr-attributes aggr-snapshot-attributes files-used",
          "aggr-attributes aggr-snapshot-attributes maxfiles-available",
          "aggr-attributes aggr-snapshot-attributes maxfiles-possible",
          "aggr-attributes aggr-snapshot-attributes maxfiles-used",
          "aggr-attributes aggr-snapshot-attributes percent-inode-used-capacity",
          "aggr-attributes aggr-snapshot-attributes percent-used-capacity",
          "aggr-attributes aggr-snapshot-attributes size-available",
          "aggr-attributes aggr-snapshot-attributes size-total",
          "aggr-attributes aggr-snapshot-attributes size-used",
          "aggr-attributes aggr-snapshot-attributes snapshot-reserve-percent"
        ]
      }
    },
    {
      "Name": "Zapi",
      "Query": "volume-get-iter",
      "BatchSize": "500",
      "ClientTimeout": "180",
      "Schedules": [
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "180s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 38,
        "List": [
          "volume-attributes volume-autosize-attributes maximum-size",
          "volume-attributes volume-autosize-attributes grow-threshold-percent",
          "volume-attributes volume-id-attributes instance-uuid",
          "volume-attributes volume-id-attributes name",
          "volume-attributes volume-id-attributes node",
          "volume-attributes volume-id-attributes owning-vserver-name",
          "volume-attributes volume-id-attributes containing-aggregate-name",
          "volume-attributes volume-id-attributes style-extended",
          "volume-attributes volume-inode-attributes files-used",
          "volume-attributes volume-inode-attributes files-total",
          "volume-attributes volume-sis-attributes compression-space-saved",
          "volume-attributes volume-sis-attributes deduplication-space-saved",
          "volume-attributes volume-sis-attributes total-space-saved",
          "volume-attributes volume-sis-attributes percentage-compression-space-saved",
          "volume-attributes volume-sis-attributes percentage-deduplication-space-saved",
          "volume-attributes volume-sis-attributes percentage-total-space-saved",
          "volume-attributes volume-space-attributes expected-available",
          "volume-attributes volume-space-attributes filesystem-size",
          "volume-attributes volume-space-attributes logical-available",
          "volume-attributes volume-space-attributes logical-used",
          "volume-attributes volume-space-attributes logical-used-by-afs",
          "volume-attributes volume-space-attributes logical-used-by-snapshots",
          "volume-attributes volume-space-attributes logical-used-percent",
          "volume-attributes volume-space-attributes physical-used",
          "volume-attributes volume-space-attributes physical-used-percent",
          "volume-attributes volume-space-attributes size",
          "volume-attributes volume-space-attributes size-available",
          "volume-attributes volume-space-attributes size-total",
          "volume-attributes volume-space-attributes size-used",
          "volume-attributes volume-space-attributes percentage-size-used",
          "volume-attributes volume-space-attributes size-used-by-snapshots",
          "volume-attributes volume-space-attributes size-available-for-snapshots",
          "volume-attributes volume-space-attributes snapshot-reserve-available",
          "volume-attributes volume-space-attributes snapshot-reserve-size",
          "volume-attributes volume-space-attributes percentage-snapshot-reserve",
          "volume-attributes volume-space-attributes percentage-snapshot-reserve-used",
          "volume-attributes volume-state-attributes state",
          "volume-attributes volume-state-attributes status"
        ]
      }
    },
    {
      "Name": "Zapi",
      "Query": "snapmirror-get-iter",
      "BatchSize": "500",
      "ClientTimeout": "180",
      "Schedules": [
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "180s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 27,
        "List": [
          "snapmirror-info break-failed-count",
          "snapmirror-info break-successful-count",
          "snapmirror-info destination-volume",
          "snapmirror-info destination-volume-node",
          "snapmirror-info destination-vserver",
          "snapmirror-info is-healthy",
          "snapmirror-info lag-time",
          "snapmirror-info last-transfer-duration",
          "snapmirror-info last-transfer-end-timestamp",
          "snapmirror-info last-transfer-size",
          "snapmirror-info last-transfer-type",
          "snapmirror-info newest-snapshot-timestamp",
          "snapmirror-info relationship-id",
          "snapmirror-info relationship-status",
          "snapmirror-info relationship-type",
          "snapmirror-info relationship-group-type",
          "snapmirror-info resync-failed-count",
          "snapmirror-info resync-successful-count",
          "snapmirror-info schedule",
          "snapmirror-info source-volume",
          "snapmirror-info source-vserver",
          "snapmirror-info source-node",
          "snapmirror-info total-transfer-bytes",
          "snapmirror-info total-transfer-time-secs",
          "snapmirror-info unhealthy-reason",
          "snapmirror-info update-failed-count",
          "snapmirror-info update-successful-count"
        ]
      }
    },
    {
      "Name": "Zapi",
      "Query": "storage-disk-get-iter",
      "BatchSize": "500",
      "ClientTimeout": "180",
      "Schedules": [
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "180s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 20,
        "List": [
          "storage-disk-info disk-uid",
          "storage-disk-info disk-name",
          "storage-disk-info disk-inventory-info bytes-per-sector",
          "storage-disk-info disk-inventory-info capacity-sectors",
          "storage-disk-info disk-inventory-info disk-type",
          "storage-disk-info disk-inventory-info is-shared",
          "storage-disk-info disk-inventory-info model",
          "storage-disk-info disk-inventory-info serial-number",
          "storage-disk-info disk-inventory-info shelf",
          "storage-disk-info disk-inventory-info shelf-bay",
          "storage-disk-info disk-ownership-info home-node-name",
          "storage-disk-info disk-ownership-info owner-node-name",
          "storage-disk-info disk-ownership-info is-failed",
          "storage-disk-info disk-stats-info average-latency",
          "storage-disk-info disk-stats-info disk-io-kbps",
          "storage-disk-info disk-stats-info power-on-time-interval",
          "storage-disk-info disk-stats-info sectors-read",
          "storage-disk-info disk-stats-info sectors-written",
          "storage-disk-info disk-raid-info disk-outage-info reason",
          "storage-disk-info disk-raid-info disk-shared-info aggregate-list shared-aggregate-info aggregate-name"
        ]
      }
    },
    {
      "Name": "Zapi",
      "Query": "storage-shelf-info-get-iter",
      "BatchSize": "500",
      "ClientTimeout": "180",
      "Schedules": [
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "180s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 9,
        "List": [
          "storage-shelf-info disk-count",
          "storage-shelf-info module-type",
          "storage-shelf-info serial-number",
          "storage-shelf-info shelf",
          "storage-shelf-info shelf-model",
          "storage-shelf-info shelf-uid",
          "storage-shelf-info state",
          "storage-shelf-info vendor-name",
          "storage-shelf-info op-status"
        ]
      }
    },
    {
      "Name": "Zapi",
      "Query": "diagnosis-status-get",
      "ClientTimeout": "180",
      "Schedules": [
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "180s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 1,
        "List": [
          "status"
        ]
      }
    },
    {
      "Name": "Zapi",
      "Query": "diagnosis-subsystem-config-get-iter",
      "BatchSize": "500",
      "ClientTimeout": "180",
      "Schedules": [
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "180s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 4,
        "List": [
          "diagnosis-subsystem-config-info health",
          "diagnosis-subsystem-config-info subsystem",
          "diagnosis-subsystem-config-info outstanding-alert-count",
          "diagnosis-subsystem-config-info suppressed-alert-count"
        ]
      }
    },
    {
      "Name": "Zapi",
      "Query": "lun-get-iter",
      "BatchSize": "500",
      "ClientTimeout": "180",
      "Schedules": [
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "180s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 9,
        "List": [
          "lun-info node",
          "lun-info path",
          "lun-info qtree",
          "lun-info size",
          "lun-info size-used",
          "lun-info state",
          "lun-info uuid",
          "lun-info volume",
          "lun-info vserver"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "system:node",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 25,
        "List": [
          "instance_name",
          "avg_processor_busy",
          "cpu_elapsed_time",
          "uptime",
          "memory",
          "total_data",
          "total_latency",
          "total_ops",
          "cifs_ops",
          "nfs_ops",
          "iscsi_ops",
          "fcp_ops",
          "nvmf_ops",
          "disk_data_read",
          "disk_data_written",
          "hdd_data_read",
          "hdd_data_written",
          "ssd_data_read",
          "ssd_data_written",
          "net_data_recv",
          "net_data_sent",
          "fcp_data_recv",
          "fcp_data_sent",
          "nvmf_data_recv",
          "mvmf_data_sent"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "hostadapter",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 4,
        "List": [
          "instance_name",
          "node_name",
          "bytes_read",
          "bytes_written"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "path",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 11,
        "List": [
          "instance_name",
          "instance_uuid",
          "node_name",
          "read_iops",
          "read_data",
          "write_data",
          "write_iops",
          "total_data",
          "total_iops",
          "read_latency",
          "write_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "disk:constituent",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 23,
        "List": [
          "instance_uuid",
          "instance_name",
          "cp_read_chain",
          "cp_read_latency",
          "cp_reads",
          "disk_busy",
          "disk_capacity",
          "disk_speed",
          "io_pending",
          "io_queued",
          "node_name",
          "physical_disk_name",
          "raid_group",
          "raid_type",
          "total_transfers",
          "user_read_chain",
          "user_read_blocks",
          "user_read_latency",
          "user_reads",
          "user_write_chain",
          "user_write_blocks",
          "user_write_latency",
          "user_writes"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "fcvi",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 6,
        "List": [
          "instance_name",
          "instance_uuid",
          "node_name",
          "rdma_write_throughput",
          "rdma_write_ops",
          "rdma_write_avg_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "ext_cache_obj",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 20,
        "List": [
          "instance_name",
          "instance_uuid",
          "node_name",
          "disk_reads_replaced",
          "accesses",
          "hit_percent",
          "inserts",
          "evicts",
          "invalidates",
          "usage",
          "hit",
          "hit_normal_lev0",
          "hit_metadata_file",
          "hit_directory",
          "hit_indirect",
          "miss",
          "miss_normal_lev0",
          "miss_metadata_file",
          "miss_directory",
          "miss_indirect"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "wafl",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 23,
        "List": [
          "instance_uuid",
          "node_name",
          "avg_non_wafl_msg_latency",
          "avg_wafl_msg_latency",
          "avg_wafl_repl_msg_latency",
          "wafl_msg_total",
          "wafl_repl_msg_total",
          "non_wafl_msg_total",
          "cp_count",
          "cp_phase_times",
          "read_io_type",
          "total_cp_msecs",
          "total_cp_util",
          "wafl_memory_used",
          "wafl_memory_free",
          "wafl_reads_from_cache",
          "wafl_reads_from_cloud",
          "wafl_reads_from_cloud_s2c_bin",
          "wafl_reads_from_disk",
          "wafl_reads_from_ext_cache",
          "wafl_reads_from_fc_miss",
          "wafl_reads_from_pmem",
          "wafl_reads_from_ssd"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "wafl_hya_per_aggr",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 19,
        "List": [
          "instance_name",
          "node_name",
          "ssd_total",
          "ssd_total_used",
          "ssd_available",
          "ssd_read_cached",
          "ssd_write_cached",
          "read_ops_replaced",
          "read_ops_replaced_percent",
          "write_blks_replaced",
          "write_blks_replaced_percent",
          "hya_read_hit_latency_average",
          "hya_read_miss_latency_average",
          "hya_write_ssd_latency_average",
          "hya_write_hdd_latency_average",
          "read_cache_ins_rate",
          "wc_write_blks_total",
          "evict_remove_rate",
          "evict_destage_rate"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "wafl_hya_sizer",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 3,
        "List": [
          "instance_name",
          "node_name",
          "cache_stats"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "nic_common",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 17,
        "List": [
          "instance_name",
          "instance_uuid",
          "node_name",
          "rx_bytes",
          "tx_bytes",
          "link_speed",
          "link_current_state",
          "link_up_to_downs",
          "nic_type",
          "rx_alignment_errors",
          "rx_crc_errors",
          "rx_length_errors",
          "rx_total_errors",
          "rx_errors",
          "tx_errors",
          "tx_hw_errors",
          "tx_total_errors"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "namespace",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 13,
        "List": [
          "instance_name",
          "vserver_name",
          "read_data",
          "write_data",
          "read_ops",
          "write_ops",
          "other_ops",
          "avg_read_latency",
          "avg_write_latency",
          "avg_other_latency",
          "queue_full",
          "remote_bytes",
          "remote_ops"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "nvmf_fc_lif",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 15,
        "List": [
          "instance_name",
          "instance_uuid",
          "vserver_name",
          "node_name",
          "port_id",
          "read_data",
          "read_ops",
          "avg_read_latency",
          "write_data",
          "write_ops",
          "avg_write_latency",
          "other_ops",
          "avg_other_latency",
          "total_ops",
          "avg_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "fcp_port",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 54,
        "List": [
          "instance_name",
          "instance_uuid",
          "node_name",
          "avg_other_latency",
          "avg_read_latency",
          "avg_write_latency",
          "read_data",
          "write_data",
          "total_data",
          "read_ops",
          "write_ops",
          "other_ops",
          "total_ops",
          "link_speed",
          "link_down",
          "link_failure",
          "loss_of_signal",
          "loss_of_sync",
          "prim_seq_err",
          "queue_full",
          "reset_count",
          "shared_int_count",
          "spurious_int_count",
          "threshold_full",
          "discarded_frames_count",
          "int_count",
          "invalid_transmission_word",
          "isr_count",
          "invalid_crc",
          "nvmf_avg_other_latency",
          "nvmf_avg_read_latency",
          "nvmf_avg_remote_other_latency",
          "nvmf_avg_remote_read_latency",
          "nvmf_avg_remote_write_latenc",
          "nvmf_avg_write_latency",
          "nvmf_caw_data",
          "nvmf_caw_ops",
          "nvmf_command_slots",
          "nvmf_other_ops",
          "nvmf_read_data",
          "nvmf_read_ops",
          "nvmf_remote_caw_data",
          "nvmf_remote_caw_ops",
          "nvmf_remote_other_ops",
          "nvmf_remote_read_data",
          "nvmf_remote_read_ops",
          "nvmf_remote_total_data",
          "nvmf_remote_total_ops",
          "nvmf_remote_write_data",
          "nvmf_remote_write_ops",
          "nvmf_total_data",
          "nvmf_total_ops",
          "nvmf_write_data",
          "nvmf_write_ops"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "object_store_client_op",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 8,
        "List": [
          "instance_name",
          "instance_uuid",
          "node_name",
          "average_latency",
          "get_throughput_bytes",
          "put_throughput_bytes",
          "stats",
          "throughput_ops"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "volume:node",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 33,
        "List": [
          "instance_name",
          "cifs_read_data",
          "cifs_write_data",
          "cifs_read_ops",
          "cifs_write_ops",
          "cifs_other_ops",
          "cifs_read_latency",
          "cifs_write_latency",
          "cifs_other_latency",
          "nfs_read_data",
          "nfs_write_data",
          "nfs_read_ops",
          "nfs_write_ops",
          "nfs_other_ops",
          "nfs_read_latency",
          "nfs_write_latency",
          "nfs_other_latency",
          "iscsi_read_data",
          "iscsi_write_data",
          "iscsi_read_ops",
          "iscsi_write_ops",
          "iscsi_other_ops",
          "iscsi_read_latency",
          "iscsi_write_latency",
          "iscsi_other_latency",
          "fcp_read_data",
          "fcp_write_data",
          "fcp_read_ops",
          "fcp_write_ops",
          "fcp_other_ops",
          "fcp_read_latency",
          "fcp_write_latency",
          "fcp_other_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "token_manager",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 12,
        "List": [
          "node_name",
          "instance_name",
          "instance_uuid",
          "token_copy_success",
          "token_create_success",
          "token_zero_success",
          "token_copy_failure",
          "token_create_failure",
          "token_zero_failure",
          "token_copy_bytes",
          "token_create_bytes",
          "token_zero_bytes"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "resource_headroom_cpu",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 13,
        "List": [
          "instance_name",
          "node_name",
          "current_latency",
          "current_ops",
          "current_utilization",
          "optimal_point_latency",
          "optimal_point_ops",
          "optimal_point_utilization",
          "optimal_point_confidence_factor",
          "ewma_daily",
          "ewma_hourly",
          "ewma_monthly",
          "ewma_weekly"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "resource_headroom_aggr",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 13,
        "List": [
          "instance_name",
          "node_name",
          "current_latency",
          "current_ops",
          "current_utilization",
          "optimal_point_latency",
          "optimal_point_ops",
          "optimal_point_utilization",
          "optimal_point_confidence_factor",
          "ewma_daily",
          "ewma_hourly",
          "ewma_monthly",
          "ewma_weekly"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "processor",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 5,
        "List": [
          "node_name",
          "instance_name",
          "instance_uuid",
          "domain_busy",
          "processor_busy"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "cifs:node",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 11,
        "List": [
          "instance_name",
          "cifs_op_count",
          "cifs_ops",
          "cifs_read_ops",
          "cifs_write_ops",
          "cifs_latency",
          "cifs_read_latency",
          "cifs_write_latency",
          "connections",
          "established_sessions",
          "open_files"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "nfsv3:node",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 52,
        "List": [
          "instance_name",
          "nfsv3_ops",
          "nfsv3_read_ops",
          "nfsv3_write_ops",
          "nfsv3_throughput",
          "nfsv3_read_throughput",
          "nfsv3_write_throughput",
          "read_avg_latency",
          "write_avg_latency",
          "latency",
          "access_total",
          "commit_total",
          "create_total",
          "fsinfo_total",
          "fsstat_total",
          "getattr_total",
          "link_total",
          "lookup_total",
          "mkdir_total",
          "mknod_total",
          "null_total",
          "pathconf_total",
          "read_symlink_total",
          "read_total",
          "readdir_total",
          "readdirplus_total",
          "remove_total",
          "rename_total",
          "rmdir_total",
          "setattr_total",
          "symlink_total",
          "write_total",
          "access_avg_latency",
          "commit_avg_latency",
          "create_avg_latency",
          "fsinfo_avg_latency",
          "fsstat_avg_latency",
          "getattr_avg_latency",
          "link_avg_latency",
          "lookup_avg_latency",
          "mkdir_avg_latency",
          "mknod_avg_latency",
          "null_avg_latency",
          "pathconf_avg_latency",
          "read_symlink_avg_latency",
          "readdir_avg_latency",
          "readdirplus_avg_latency",
          "remove_avg_latency",
          "rename_avg_latency",
          "rmdir_avg_latency",
          "setattr_avg_latency",
          "symlink_avg_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "nfsv4:node",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 82,
        "List": [
          "instance_name",
          "latency",
          "total_ops",
          "nfs4_read_throughput",
          "nfs4_throughput",
          "nfs4_write_throughput",
          "access_total",
          "close_total",
          "commit_total",
          "create_total",
          "delegpurge_total",
          "delegreturn_total",
          "getattr_total",
          "getfh_total",
          "link_total",
          "lock_total",
          "lockt_total",
          "locku_total",
          "lookup_total",
          "lookupp_total",
          "null_total",
          "nverify_total",
          "open_confirm_total",
          "open_downgrade_total",
          "open_total",
          "openattr_total",
          "putfh_total",
          "putpubfh_total",
          "putrootfh_total",
          "read_total",
          "readdir_total",
          "readlink_total",
          "release_lock_owner_total",
          "remove_total",
          "rename_total",
          "renew_total",
          "restorefh_total",
          "savefh_total",
          "secinfo_total",
          "setattr_total",
          "setclientid_confirm_total",
          "setclientid_total",
          "verify_total",
          "write_total",
          "access_avg_latency",
          "close_avg_latency",
          "commit_avg_latency",
          "create_avg_latency",
          "delegpurge_avg_latency",
          "delegreturn_avg_latency",
          "getattr_avg_latency",
          "getfh_avg_latency",
          "link_avg_latency",
          "lock_avg_latency",
          "lockt_avg_latency",
          "locku_avg_latency",
          "lookup_avg_latency",
          "lookupp_avg_latency",
          "null_avg_latency",
          "nverify_avg_latency",
          "open_avg_latency",
          "open_confirm_avg_latency",
          "open_downgrade_avg_latency",
          "openattr_avg_latency",
          "putfh_avg_latency",
          "putpubfh_avg_latency",
          "putrootfh_avg_latency",
          "read_avg_latency",
          "readdir_avg_latency",
          "readlink_avg_latency",
          "release_lock_owner_avg_latency",
          "remove_avg_latency",
          "rename_avg_latency",
          "renew_avg_latency",
          "restorefh_avg_latency",
          "savefh_avg_latency",
          "secinfo_avg_latency",
          "setattr_avg_latency",
          "setclientid_avg_latency",
          "setclientid_confirm_avg_latency",
          "verify_avg_latency",
          "write_avg_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "nfsv4_1:node",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 110,
        "List": [
          "instance_name",
          "latency",
          "total_ops",
          "nfs41_read_throughput",
          "nfs41_throughput",
          "nfs41_write_throughput",
          "access_total",
          "backchannel_ctl_total",
          "bind_conn_to_session_total",
          "close_total",
          "commit_total",
          "create_session_total",
          "create_total",
          "delegpurge_total",
          "delegreturn_total",
          "destroy_clientid_total",
          "destroy_session_total",
          "exchange_id_total",
          "free_stateid_total",
          "get_dir_delegation_total",
          "getattr_total",
          "getdeviceinfo_total",
          "getdevicelist_total",
          "getfh_total",
          "layoutcommit_total",
          "layoutget_total",
          "layoutreturn_total",
          "link_total",
          "lock_total",
          "lockt_total",
          "locku_total",
          "lookup_total",
          "lookupp_total",
          "null_total",
          "nverify_total",
          "open_downgrade_total",
          "open_total",
          "openattr_total",
          "putfh_total",
          "putpubfh_total",
          "putrootfh_total",
          "read_total",
          "readdir_total",
          "readlink_total",
          "reclaim_complete_total",
          "remove_total",
          "rename_total",
          "restorefh_total",
          "savefh_total",
          "secinfo_no_name_total",
          "secinfo_total",
          "sequence_total",
          "set_ssv_total",
          "setattr_total",
          "test_stateid_total",
          "verify_total",
          "want_delegation_total",
          "write_total",
          "access_avg_latency",
          "backchannel_ctl_avg_latency",
          "bind_conn_to_session_avg_latency",
          "close_avg_latency",
          "commit_avg_latency",
          "create_avg_latency",
          "create_session_avg_latency",
          "delegpurge_avg_latency",
          "delegreturn_avg_latency",
          "destroy_clientid_avg_latency",
          "destroy_session_avg_latency",
          "exchange_id_avg_latency",
          "free_stateid_avg_latency",
          "get_dir_delegation_avg_latency",
          "getattr_avg_latency",
          "getdeviceinfo_avg_latency",
          "getdevicelist_avg_latency",
          "getfh_avg_latency",
          "layoutcommit_avg_latency",
          "layoutget_avg_latency",
          "layoutreturn_avg_latency",
          "link_avg_latency",
          "lock_avg_latency",
          "lockt_avg_latency",
          "locku_avg_latency",
          "lookup_avg_latency",
          "lookupp_avg_latency",
          "null_avg_latency",
          "nverify_avg_latency",
          "open_avg_latency",
          "open_downgrade_avg_latency",
          "openattr_avg_latency",
          "putfh_avg_latency",
          "putpubfh_avg_latency",
          "putrootfh_avg_latency",
          "read_avg_latency",
          "readdir_avg_latency",
          "readlink_avg_latency",
          "reclaim_complete_avg_latency",
          "remove_avg_latency",
          "rename_avg_latency",
          "restorefh_avg_latency",
          "savefh_avg_latency",
          "secinfo_avg_latency",
          "secinfo_no_name_avg_latency",
          "sequence_avg_latency",
          "set_ssv_avg_latency",
          "setattr_avg_latency",
          "test_stateid_avg_latency",
          "verify_avg_latency",
          "want_delegation_avg_latency",
          "write_avg_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "qtree",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 9,
        "List": [
          "instance_uuid",
          "instance_name",
          "vserver_name",
          "node_name",
          "parent_vol",
          "cifs_ops",
          "nfs_ops",
          "internal_ops",
          "total_ops"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "volume",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 15,
        "List": [
          "instance_uuid",
          "instance_name",
          "vserver_name",
          "node_name",
          "parent_aggr",
          "read_data",
          "write_data",
          "read_ops",
          "write_ops",
          "other_ops",
          "total_ops",
          "read_latency",
          "write_latency",
          "other_latency",
          "avg_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "lun",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 23,
        "List": [
          "instance_name",
          "vserver_name",
          "read_data",
          "write_data",
          "read_ops",
          "write_ops",
          "xcopy_reqs",
          "writesame_reqs",
          "writesame_unmap_reqs",
          "caw_reqs",
          "avg_read_latency",
          "avg_write_latency",
          "avg_xcopy_latency",
          "unmap_reqs",
          "writesame_unmap_reqs",
          "enospc",
          "queue_full",
          "remote_bytes",
          "remote_ops",
          "read_partial_blocks",
          "write_partial_blocks",
          "write_align_histo",
          "read_align_histo"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "nfsv3",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 52,
        "List": [
          "instance_name",
          "nfsv3_ops",
          "nfsv3_read_ops",
          "nfsv3_write_ops",
          "nfsv3_throughput",
          "nfsv3_read_throughput",
          "nfsv3_write_throughput",
          "read_avg_latency",
          "write_avg_latency",
          "latency",
          "access_total",
          "commit_total",
          "create_total",
          "fsinfo_total",
          "fsstat_total",
          "getattr_total",
          "link_total",
          "lookup_total",
          "mkdir_total",
          "mknod_total",
          "null_total",
          "pathconf_total",
          "read_symlink_total",
          "read_total",
          "readdir_total",
          "readdirplus_total",
          "remove_total",
          "rename_total",
          "rmdir_total",
          "setattr_total",
          "symlink_total",
          "write_total",
          "access_avg_latency",
          "commit_avg_latency",
          "create_avg_latency",
          "fsinfo_avg_latency",
          "fsstat_avg_latency",
          "getattr_avg_latency",
          "link_avg_latency",
          "lookup_avg_latency",
          "mkdir_avg_latency",
          "mknod_avg_latency",
          "null_avg_latency",
          "pathconf_avg_latency",
          "read_symlink_avg_latency",
          "readdir_avg_latency",
          "readdirplus_avg_latency",
          "remove_avg_latency",
          "rename_avg_latency",
          "rmdir_avg_latency",
          "setattr_avg_latency",
          "symlink_avg_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "nfsv4",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 82,
        "List": [
          "instance_name",
          "latency",
          "total_ops",
          "nfs4_read_throughput",
          "nfs4_throughput",
          "nfs4_write_throughput",
          "access_total",
          "close_total",
          "commit_total",
          "create_total",
          "delegpurge_total",
          "delegreturn_total",
          "getattr_total",
          "getfh_total",
          "link_total",
          "lock_total",
          "lockt_total",
          "locku_total",
          "lookup_total",
          "lookupp_total",
          "null_total",
          "nverify_total",
          "open_confirm_total",
          "open_downgrade_total",
          "open_total",
          "openattr_total",
          "putfh_total",
          "putpubfh_total",
          "putrootfh_total",
          "read_total",
          "readdir_total",
          "readlink_total",
          "release_lock_owner_total",
          "remove_total",
          "rename_total",
          "renew_total",
          "restorefh_total",
          "savefh_total",
          "secinfo_total",
          "setattr_total",
          "setclientid_confirm_total",
          "setclientid_total",
          "verify_total",
          "write_total",
          "access_avg_latency",
          "close_avg_latency",
          "commit_avg_latency",
          "create_avg_latency",
          "delegpurge_avg_latency",
          "delegreturn_avg_latency",
          "getattr_avg_latency",
          "getfh_avg_latency",
          "link_avg_latency",
          "lock_avg_latency",
          "lockt_avg_latency",
          "locku_avg_latency",
          "lookup_avg_latency",
          "lookupp_avg_latency",
          "null_avg_latency",
          "nverify_avg_latency",
          "open_avg_latency",
          "open_confirm_avg_latency",
          "open_downgrade_avg_latency",
          "openattr_avg_latency",
          "putfh_avg_latency",
          "putpubfh_avg_latency",
          "putrootfh_avg_latency",
          "read_avg_latency",
          "readdir_avg_latency",
          "readlink_avg_latency",
          "release_lock_owner_avg_latency",
          "remove_avg_latency",
          "rename_avg_latency",
          "renew_avg_latency",
          "restorefh_avg_latency",
          "savefh_avg_latency",
          "secinfo_avg_latency",
          "setattr_avg_latency",
          "setclientid_avg_latency",
          "setclientid_confirm_avg_latency",
          "verify_avg_latency",
          "write_avg_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "nfsv4_1",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 110,
        "List": [
          "instance_name",
          "latency",
          "total_ops",
          "nfs41_read_throughput",
          "nfs41_throughput",
          "nfs41_write_throughput",
          "access_total",
          "backchannel_ctl_total",
          "bind_conn_to_session_total",
          "close_total",
          "commit_total",
          "create_session_total",
          "create_total",
          "delegpurge_total",
          "delegreturn_total",
          "destroy_clientid_total",
          "destroy_session_total",
          "exchange_id_total",
          "free_stateid_total",
          "get_dir_delegation_total",
          "getattr_total",
          "getdeviceinfo_total",
          "getdevicelist_total",
          "getfh_total",
          "layoutcommit_total",
          "layoutget_total",
          "layoutreturn_total",
          "link_total",
          "lock_total",
          "lockt_total",
          "locku_total",
          "lookup_total",
          "lookupp_total",
          "null_total",
          "nverify_total",
          "open_downgrade_total",
          "open_total",
          "openattr_total",
          "putfh_total",
          "putpubfh_total",
          "putrootfh_total",
          "read_total",
          "readdir_total",
          "readlink_total",
          "reclaim_complete_total",
          "remove_total",
          "rename_total",
          "restorefh_total",
          "savefh_total",
          "secinfo_no_name_total",
          "secinfo_total",
          "sequence_total",
          "set_ssv_total",
          "setattr_total",
          "test_stateid_total",
          "verify_total",
          "want_delegation_total",
          "write_total",
          "access_avg_latency",
          "backchannel_ctl_avg_latency",
          "bind_conn_to_session_avg_latency",
          "close_avg_latency",
          "commit_avg_latency",
          "create_avg_latency",
          "create_session_avg_latency",
          "delegpurge_avg_latency",
          "delegreturn_avg_latency",
          "destroy_clientid_avg_latency",
          "destroy_session_avg_latency",
          "exchange_id_avg_latency",
          "free_stateid_avg_latency",
          "get_dir_delegation_avg_latency",
          "getattr_avg_latency",
          "getdeviceinfo_avg_latency",
          "getdevicelist_avg_latency",
          "getfh_avg_latency",
          "layoutcommit_avg_latency",
          "layoutget_avg_latency",
          "layoutreturn_avg_latency",
          "link_avg_latency",
          "lock_avg_latency",
          "lockt_avg_latency",
          "locku_avg_latency",
          "lookup_avg_latency",
          "lookupp_avg_latency",
          "null_avg_latency",
          "nverify_avg_latency",
          "open_avg_latency",
          "open_downgrade_avg_latency",
          "openattr_avg_latency",
          "putfh_avg_latency",
          "putpubfh_avg_latency",
          "putrootfh_avg_latency",
          "read_avg_latency",
          "readdir_avg_latency",
          "readlink_avg_latency",
          "reclaim_complete_avg_latency",
          "remove_avg_latency",
          "rename_avg_latency",
          "restorefh_avg_latency",
          "savefh_avg_latency",
          "secinfo_avg_latency",
          "secinfo_no_name_avg_latency",
          "sequence_avg_latency",
          "set_ssv_avg_latency",
          "setattr_avg_latency",
          "test_stateid_avg_latency",
          "verify_avg_latency",
          "want_delegation_avg_latency",
          "write_avg_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "cifs:vserver",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 13,
        "List": [
          "instance_uuid",
          "instance_name",
          "cifs_op_count",
          "cifs_ops",
          "cifs_read_ops",
          "cifs_write_ops",
          "cifs_latency",
          "cifs_read_latency",
          "cifs_write_latency",
          "connections",
          "established_sessions",
          "open_files",
          "signed_sessions"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "lif",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 11,
        "List": [
          "instance_uuid",
          "instance_name",
          "node_name",
          "vserver_name",
          "current_port",
          "recv_data",
          "recv_packet",
          "recv_errors",
          "sent_data",
          "sent_packet",
          "sent_errors"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "iscsi_lif",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 15,
        "List": [
          "instance_name",
          "instance_uuid",
          "vserver_name",
          "node_name",
          "protocol_errors",
          "read_data",
          "write_data",
          "iscsi_read_ops",
          "avg_read_latency",
          "iscsi_write_ops",
          "avg_write_latency",
          "iscsi_other_ops",
          "avg_other_latency",
          "cmd_transfered",
          "avg_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "fcp_lif",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 15,
        "List": [
          "instance_name",
          "instance_uuid",
          "vserver_name",
          "node_name",
          "port_id",
          "read_data",
          "read_ops",
          "avg_read_latency",
          "write_data",
          "write_ops",
          "avg_write_latency",
          "other_ops",
          "avg_other_latency",
          "total_ops",
          "avg_latency"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "copy_manager",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 7,
        "List": [
          "instance_name",
          "instance_uuid",
          "bce_copy_count_curr",
          "ocs_copy_count_curr",
          "sce_copy_count_curr",
          "spince_copy_count_curr",
          "KB_copied"
        ]
      }
    },
    {
      "Name": "ZapiPerf",
      "Query": "wafl_comp_aggr_vol_bin",
      "ClientTimeout": "60s",
      "Schedules": [
        {
          "Name": "counter",
          "Schedule": "1200s"
        },
        {
          "Name": "instance",
          "Schedule": "600s"
        },
        {
          "Name": "data",
          "Schedule": "60s"
        }
      ],
      "Exporters": [
        "Prometheus"
      ],
      "Counters": {
        "Count": 5,
        "List": [
          "instance_name",
          "vserver_name",
          "vol_name",
          "cloud_bin_operation",
          "cloud_bin_op_latency_average"
        ]
      }
    }
  ]
}

```