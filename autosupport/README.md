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
  "Version": "9.13.1",
  "Model": "cdot",
  "Serial": "721802000259",
  "Ping": 0.41200000047683716,
  "ClusterUUID": "cbd1757b-0580-11e8-bd9d-00a098d39e12"
 },
 "Harvest": {
  "HostHash": "8de858c125e674ff0d3c8ea1882e84ef31b7afd2",
  "UUID": "915c13f51cf9265b6ead640cd0a0ac4b35fe30a4",
  "Version": "dev",
  "Release": "v23.08.0",
  "Commit": "bf06676d",
  "BuildDate": "2023-10-12T14:07:44-0400",
  "NumClusters": 1,
  "NumPollers": 14,
  "NumExporters": 1,
  "NumPortRange": 1
 },
 "Platform": {
  "OS": "linux",
  "Arch": "ubuntu",
  "Memory": {
   "TotalKb": 10231524,
   "AvailableKb": 5948860,
   "UsedKb": 3969240
  },
  "CPUs": 8,
  "NumProcesses": 215,
  "Processes": [
   {
    "Pid": 1620,
    "User": "prometheus",
    "Ppid": 1,
    "Ctime": 1687348024000,
    "RssBytes": 2732367872,
    "Threads": 15,
    "Cmdline": "/usr/bin/prometheus --config.file /etc/prometheus/prometheus.yml --storage.tsdb.path=/var/lib/prometheus/metrics2/ --web.enable-admin-api"
   },
   {
    "Pid": 2492,
    "User": "grafana",
    "Ppid": 1,
    "Ctime": 1687348029000,
    "RssBytes": 111095808,
    "Threads": 81,
    "Cmdline": "/usr/sbin/grafana-server --config=/etc/grafana/grafana.ini --pidfile=/var/run/grafana/grafana-server.pid --packaging=deb cfg:default.paths.logs=/var/log/grafana cfg:default.paths.data=/var/lib/grafana cfg:default.paths.plugins=/var/lib/grafana/plugins cfg:default.paths.provisioning=/etc/grafana/provisioning"
   },
   {
    "Pid": 3534,
    "User": "harvest",
    "Ppid": 1,
    "Ctime": 1697134192000,
    "RssBytes": 75702272,
    "Threads": 16,
    "Cmdline": "bin/poller --poller sar --loglevel 2 --promPort 13002 --config ./harvest.yml --daemon"
   },
   {
    "Pid": 15155,
    "User": "harvest",
    "Ppid": 15098,
    "Ctime": 1695400595000,
    "RssBytes": 6549504,
    "Threads": 12,
    "Cmdline": "bin/harvest admin start"
   },
   {
    "Pid": 16376,
    "User": "prometheus",
    "Ppid": 1,
    "Ctime": 1695401228000,
    "RssBytes": 16408576,
    "Threads": 20,
    "Cmdline": "/usr/bin/prometheus-node-exporter --collector.diskstats.ignored-devices=^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvmed+nd+p)d+$ --collector.filesystem.ignored-mount-points=^/(sys|proc|dev|run)($|/) --collector.netdev.ignored-devices=^lo$ --collector.textfile.directory=/var/lib/prometheus/node-exporter"
   }
  ]
 },
 "Nodes": {
  "Count": 2,
  "DataPoints": 40,
  "PollTime": 588253,
  "APITime": 310107,
  "ParseTime": 278009,
  "PluginTime": 19,
  "Ids": [
   {
    "serial-number": "721802000259",
    "system-id": "0537123843"
   },
   {
    "serial-number": "721802000260",
    "system-id": "0537124012"
   }
  ]
 },
 "Volumes": {
  "Count": 200,
  "DataPoints": 7982,
  "PollTime": 6253384,
  "APITime": 4120720,
  "ParseTime": 2131821,
  "PluginTime": 720
 },
 "Collectors": [
  {
   "Name": "Rest",
   "Query": "api/security/ssh",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 3,
    "List": [
     "ciphers",
     "mac_algorithms",
     "max_instances"
    ]
   },
   "InstanceInfo": {
    "Count": 1,
    "DataPoints": 3,
    "PollTime": 271719,
    "APITime": 271650,
    "ParseTime": 20,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/svm_cifs:node",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 12,
    "List": [
     "latency",
     "total_ops",
     "average_write_latency",
     "connections",
     "average_read_latency",
     "established_sessions",
     "op_count",
     "open_files",
     "total_read_ops",
     "total_write_ops",
     "id",
     "node.name"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/svm_vscan",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 7,
    "List": [
     "scan.notification_received_rate",
     "scan.request_dispatched_rate",
     "id",
     "svm.name",
     "connections_active",
     "dispatch.latency",
     "scan.latency"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/support/ems/destinations",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 7,
    "List": [
     "system_defined",
     "destination",
     "name",
     "type",
     "certificate.ca",
     "filter",
     "syslog"
    ]
   },
   "InstanceInfo": {
    "Count": 87,
    "DataPoints": 435,
    "PollTime": 355181,
    "APITime": 354391,
    "ParseTime": 709,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/storage/namespaces",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 13,
    "List": [
     "status.read_only",
     "status.state",
     "subsystem_map.nsid",
     "space.size",
     "space.used",
     "location.volume.name",
     "name",
     "os_type",
     "location.node.name",
     "svm.name",
     "space.block_size",
     "uuid",
     "subsystem"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 22,
    "PollTime": 189798,
    "APITime": 189695,
    "ParseTime": 52,
    "PluginTime": 4
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/namespace",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 13,
    "List": [
     "write_ops",
     "average_other_latency",
     "average_read_latency",
     "write_data",
     "svm.name",
     "remote.read_ops",
     "read_data",
     "remote.read_data",
     "id",
     "average_write_latency",
     "read_ops",
     "name",
     "other_ops"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 26,
    "PollTime": 215755,
    "APITime": 215463,
    "ParseTime": 224,
    "PluginTime": 4
   }
  },
  {
   "Name": "Rest",
   "Query": "api/security/login/messages",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 4,
    "List": [
     "uuid",
     "banner",
     "scope",
     "svm.name"
    ]
   },
   "InstanceInfo": {
    "Count": 26,
    "DataPoints": 81,
    "PollTime": 279191,
    "APITime": 279031,
    "ParseTime": 107,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/wafl_hya_per_aggregate",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 20,
    "List": [
     "id",
     "hya_aggregate_name",
     "hya_write_ssd_latency_average",
     "read_ops_replaced_percent",
     "ssd_total",
     "write_blocks_replaced_percent",
     "node.name",
     "evict_destage_rate",
     "hya_read_miss_latency_average",
     "read_cache_insert_rate",
     "read_ops_replaced",
     "ssd_read_cached",
     "ssd_total_used",
     "ssd_write_cached",
     "write_blocks_replaced",
     "evict_remove_rate",
     "hya_read_hit_latency_average",
     "hya_write_hdd_latency_average",
     "ssd_available",
     "wc_write_blocks_total"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/volume:svm",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 12,
    "List": [
     "id",
     "name",
     "average_latency",
     "bytes_written",
     "other_latency",
     "total_other_ops",
     "write_latency",
     "bytes_read",
     "read_latency",
     "total_ops",
     "total_read_ops",
     "total_write_ops"
    ]
   },
   "InstanceInfo": {
    "Count": 27,
    "DataPoints": 324,
    "PollTime": 343186,
    "APITime": 339786,
    "ParseTime": 3204,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/private/cli/vserver/object-store-server/bucket/policy",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 7,
    "List": [
     "resource",
     "bucket",
     "effect",
     "index",
     "vserver",
     "action",
     "principal"
    ]
   },
   "InstanceInfo": {
    "Count": 8,
    "DataPoints": 56,
    "PollTime": 282069,
    "APITime": 281849,
    "ParseTime": 131,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/storage/volumes",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 41,
    "List": [
     "svm.name",
     "encryption.enabled",
     "uuid",
     "space.afs_total",
     "space.physical_used_percent",
     "space.size_available_for_snapshots",
     "snapshot_count",
     "space.snapshot.reserve_size",
     "anti_ransomware.dry_run_start_time",
     "autosize.grow_threshold",
     "space.overwrite_reserve_used",
     "space.percent_used",
     "space.filesystem_size",
     "space.logical_space.available",
     "name",
     "aggregates.#.uuid",
     "anti_ransomware.state",
     "is_svm_root",
     "nas.path",
     "snaplock.type",
     "space.logical_space.used",
     "space.physical_used",
     "aggregates.#.name",
     "snapshot_policy.name",
     "state",
     "type",
     "space.snapshot.reserve_percent",
     "space.snapshot.autodelete_enabled",
     "space.logical_space.used_by_afs",
     "space.overwrite_reserve",
     "space.used",
     "style",
     "space.available",
     "space.expected_available",
     "space.logical_space.used_percent",
     "space.snapshot.reserve_available",
     "space.snapshot.used",
     "autosize.maximum",
     "space.logical_space.used_by_snapshots",
     "space.size",
     "space.snapshot.space_used_percent"
    ]
   },
   "InstanceInfo": {
    "Count": 200,
    "DataPoints": 7982,
    "PollTime": 6253384,
    "APITime": 4120720,
    "ParseTime": 2131821,
    "PluginTime": 720
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/volume",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 32,
    "List": [
     "parent_aggregate",
     "nfs.lookup_latency",
     "nfs.punch_hole_latency",
     "nfs.write_latency",
     "nfs.other_latency",
     "nfs.setattr_ops",
     "write_latency",
     "nfs.lookup_ops",
     "nfs.punch_hole_ops",
     "read_latency",
     "total_other_ops",
     "other_latency",
     "total_read_ops",
     "name",
     "bytes_read",
     "nfs.getattr_latency",
     "nfs.read_latency",
     "svm.name",
     "nfs.other_ops",
     "node.name",
     "nfs.write_ops",
     "total_write_ops",
     "average_latency",
     "nfs.access_latency",
     "nfs.access_ops",
     "nfs.setattr_latency",
     "nfs.total_ops",
     "total_ops",
     "uuid",
     "bytes_written",
     "nfs.getattr_ops",
     "nfs.read_ops"
    ]
   },
   "InstanceInfo": {
    "Count": 229,
    "DataPoints": 7328,
    "PollTime": 621127,
    "APITime": 585392,
    "ParseTime": 31094,
    "PluginTime": 3386
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/svm_cifs",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 14,
    "List": [
     "id",
     "svm.name",
     "average_write_latency",
     "op_count",
     "connections",
     "total_ops",
     "total_read_ops",
     "node.name",
     "average_read_latency",
     "established_sessions",
     "total_write_ops",
     "latency",
     "open_files",
     "signed_sessions"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/iscsi_lif",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 15,
    "List": [
     "average_latency",
     "write_data",
     "id",
     "name",
     "iscsi_write_ops",
     "read_data",
     "node.name",
     "average_read_latency",
     "svm.name",
     "iscsi_read_ops",
     "cmd_transferred",
     "iscsi_other_ops",
     "protocol_errors",
     "average_other_latency",
     "average_write_latency"
    ]
   },
   "InstanceInfo": {
    "Count": 4,
    "DataPoints": 60,
    "PollTime": 163015,
    "APITime": 162314,
    "ParseTime": 616,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/qos_detail_volume",
   "ClientTimeout": "1m30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 4,
    "List": [
     "id",
     "service_time",
     "visits",
     "wait_time"
    ]
   },
   "InstanceInfo": {
    "Count": 211,
    "DataPoints": 748,
    "PollTime": 875693,
    "APITime": 364550,
    "ParseTime": 509125,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/fcp",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 54,
    "List": [
     "name",
     "shared_interrupt_count",
     "total_ops",
     "isr.count",
     "nvmf_remote.other_ops",
     "nvmf.write_data",
     "nvmf_remote.total_data",
     "write_data",
     "nvmf.average_read_latency",
     "nvmf.average_remote_other_latency",
     "nvmf_remote.read_data",
     "loss_of_sync",
     "nvmf.average_other_latency",
     "nvmf.caw_data",
     "average_other_latency",
     "average_read_latency",
     "invalid.transmission_word",
     "nvmf.average_remote_write_latency",
     "nvmf_remote.write_ops",
     "id",
     "primitive_seq_err",
     "nvmf.total_data",
     "nvmf.write_ops",
     "spurious_interrupt_count",
     "total_data",
     "link.speed",
     "average_write_latency",
     "link.down",
     "nvmf.total_ops",
     "nvmf_remote.caw_ops",
     "read_data",
     "threshold_full",
     "nvmf.other_ops",
     "nvmf_remote.read_ops",
     "other_ops",
     "reset_count",
     "link_failure",
     "loss_of_signal",
     "nvmf.average_write_latency",
     "nvmf.command_slots",
     "nvmf.read_data",
     "queue_full",
     "discarded_frames_count",
     "interrupt_count",
     "read_ops",
     "invalid.crc",
     "nvmf.average_remote_read_latency",
     "nvmf.caw_ops",
     "nvmf_remote.total_ops",
     "node.name",
     "nvmf.read_ops",
     "nvmf_remote.write_data",
     "write_ops",
     "nvmf_remote.caw_data"
    ]
   },
   "InstanceInfo": {
    "Count": 8,
    "DataPoints": 432,
    "PollTime": 140399,
    "APITime": 137220,
    "ParseTime": 2922,
    "PluginTime": 10
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/qos_volume",
   "ClientTimeout": "1m30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 14,
    "List": [
     "uuid",
     "sequential_reads_percent",
     "read_ops",
     "total_data",
     "write_data",
     "latency",
     "ops",
     "write_ops",
     "read_io_type_percent",
     "read_latency",
     "sequential_writes_percent",
     "write_latency",
     "concurrency",
     "read_data"
    ]
   },
   "InstanceInfo": {
    "Count": 211,
    "DataPoints": 783,
    "PollTime": 184891,
    "APITime": 181557,
    "ParseTime": 2092,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/storage/aggregates",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 54,
    "List": [
     "inode_attributes.used_percent",
     "snapshot.files_total",
     "space.block_storage.volume_deduplication_space_saved",
     "space.snapshot.total",
     "block_storage.primary.raid_type",
     "space.block_storage.available",
     "inode_attributes.files_used",
     "block_storage.primary.disk_type",
     "block_storage.plexes.#",
     "snapshot.files_used",
     "space.efficiency_without_snapshots.logical_used",
     "space.snapshot.reserve_percent",
     "volume_count",
     "name",
     "space.block_storage.data_compacted_count",
     "space.block_storage.data_compaction_space_saved",
     "space.block_storage.inactive_user_data_percent",
     "space.footprint_percent",
     "home_node.name",
     "block_storage.primary.raid_size",
     "space.snapshot.used_percent",
     "state",
     "space.block_storage.size",
     "space.block_storage.volume_deduplication_shared_count",
     "space.efficiency.logical_used",
     "space.efficiency_without_snapshots_flexclones.logical_used",
     "inode_attributes.file_public_capacity",
     "space.snapshot.available",
     "space.snapshot.used",
     "space.block_storage.volume_deduplication_space_saved_percent",
     "space.efficiency.savings",
     "block_storage.primary.disk_count",
     "inode_attributes.files_private_used",
     "inode_attributes.max_files_possible",
     "inode_attributes.max_files_used",
     "snapshot.max_files_used",
     "space.block_storage.data_compaction_space_saved_percent",
     "space.block_storage.physical_used",
     "cloud_storage.stores.#.cloud_store.name",
     "inode_attributes.file_private_capacity",
     "snapshot.max_files_available",
     "space.cloud_storage.used",
     "space.efficiency_without_snapshots_flexclones.savings",
     "data_encryption.software_encryption_enabled",
     "space.block_storage.physical_used_percent",
     "space.block_storage.inactive_user_data",
     "space.efficiency_without_snapshots.savings",
     "space.footprint",
     "inode_attributes.max_files_available",
     "space.block_storage.used",
     "block_storage.hybrid_cache.disk_count",
     "block_storage.hybrid_cache.size",
     "inode_attributes.files_total",
     "uuid"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 103,
    "PollTime": 397450,
    "APITime": 396846,
    "ParseTime": 515,
    "PluginTime": 22
   }
  },
  {
   "Name": "Rest",
   "Query": "api/network/ip/interfaces",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 13,
    "List": [
     "location.home_port.name",
     "name",
     "ipspace.name",
     "location.home_node.name",
     "location.node.name",
     "svm.name",
     "location.is_home",
     "ip.address",
     "location.port.name",
     "services",
     "state",
     "subnet.name",
     "uuid"
    ]
   },
   "InstanceInfo": {
    "Count": 39,
    "DataPoints": 489,
    "PollTime": 962240,
    "APITime": 366472,
    "ParseTime": 595704,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/lif",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 11,
    "List": [
     "id",
     "current_port",
     "node.name",
     "svm.name",
     "received_data",
     "received_packets",
     "sent_data",
     "sent_errors",
     "name",
     "received_errors",
     "sent_packets"
    ]
   },
   "InstanceInfo": {
    "Count": 38,
    "DataPoints": 418,
    "PollTime": 160110,
    "APITime": 159253,
    "ParseTime": 750,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/svm_nfs_v42:node",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 111,
    "List": [
     "node.name",
     "backchannel_ctl.average_latency",
     "putpubfh.average_latency",
     "secinfo.total",
     "getdevicelist.total",
     "lock.total",
     "lookupp.total",
     "total_ops",
     "sequence.average_latency",
     "delegpurge.average_latency",
     "layoutcommit.average_latency",
     "lookupp.average_latency",
     "null.total",
     "nverify.average_latency",
     "id",
     "destroy_clientid.average_latency",
     "get_dir_delegation.average_latency",
     "bind_conn_to_session.average_latency",
     "lookup.total",
     "want_delegation.average_latency",
     "create_session.average_latency",
     "total.write_throughput",
     "layoutcommit.total",
     "open_downgrade.total",
     "secinfo_no_name.average_latency",
     "open_downgrade.average_latency",
     "secinfo.average_latency",
     "savefh.average_latency",
     "exchange_id.total",
     "layoutget.average_latency",
     "link.average_latency",
     "openattr.average_latency",
     "readdir.total",
     "delegpurge.total",
     "layoutreturn.total",
     "set_ssv.total",
     "setattr.average_latency",
     "verify.average_latency",
     "destroy_session.average_latency",
     "getdeviceinfo.total",
     "openattr.total",
     "remove.total",
     "set_ssv.average_latency",
     "create.average_latency",
     "layoutreturn.average_latency",
     "locku.average_latency",
     "open.average_latency",
     "read.average_latency",
     "putpubfh.total",
     "write.average_latency",
     "access.average_latency",
     "getfh.average_latency",
     "putfh.total",
     "reclaim_complete.average_latency",
     "restorefh.total",
     "bind_conn_to_session.total",
     "lockt.total",
     "sequence.total",
     "destroy_clientid.total",
     "getattr.average_latency",
     "open.total",
     "test_stateid.total",
     "commit.total",
     "create.total",
     "delegreturn.average_latency",
     "getdevicelist.average_latency",
     "readdir.average_latency",
     "free_stateid.total",
     "getdeviceinfo.average_latency",
     "getfh.total",
     "get_dir_delegation.total",
     "latency",
     "total.throughput",
     "verify.total",
     "want_delegation.total",
     "write.total",
     "access.total",
     "close.total",
     "free_stateid.average_latency",
     "destroy_session.total",
     "null.average_latency",
     "read.total",
     "savefh.total",
     "secinfo_no_name.total",
     "getattr.total",
     "setattr.total",
     "reclaim_complete.total",
     "rename.average_latency",
     "lockt.average_latency",
     "readlink.average_latency",
     "readlink.total",
     "remove.average_latency",
     "lookup.average_latency",
     "putrootfh.total",
     "backchannel_ctl.total",
     "close.average_latency",
     "commit.average_latency",
     "create_session.total",
     "exchange_id.average_latency",
     "layoutget.total",
     "lock.average_latency",
     "locku.total",
     "nverify.total",
     "delegreturn.total",
     "putfh.average_latency",
     "restorefh.average_latency",
     "link.total",
     "putrootfh.average_latency",
     "rename.total",
     "test_stateid.average_latency",
     "total.read_throughput"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 222,
    "PollTime": 159031,
    "APITime": 158373,
    "ParseTime": 414,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/fcp_lif",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 15,
    "List": [
     "port.id",
     "average_read_latency",
     "write_data",
     "id",
     "node.name",
     "average_latency",
     "read_data",
     "total_ops",
     "average_other_latency",
     "average_write_latency",
     "read_ops",
     "write_ops",
     "name",
     "svm.name",
     "other_ops"
    ]
   },
   "InstanceInfo": {
    "Count": 8,
    "DataPoints": 120,
    "PollTime": 288753,
    "APITime": 287366,
    "ParseTime": 1309,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/wafl_comp_aggr_vol_bin",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 6,
    "List": [
     "cloud_bin_op",
     "cloud_bin_op_latency_average",
     "id",
     "cloud_target.name",
     "svm.name",
     "volume.name"
    ]
   },
   "InstanceInfo": {
    "Count": 178,
    "DataPoints": 2848,
    "PollTime": 377845,
    "APITime": 367586,
    "ParseTime": 9186,
    "PluginTime": 725
   }
  },
  {
   "Name": "Rest",
   "Query": "api/protocols/s3/buckets",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 11,
    "List": [
     "qos_policy.name",
     "svm.name",
     "logical_used_size",
     "uuid",
     "name",
     "protection_status.destination.is_ontap",
     "protection_status.is_protected",
     "encryption.enabled",
     "protection_status.destination.is_cloud",
     "volume.name",
     "size"
    ]
   },
   "InstanceInfo": {
    "Count": 5,
    "DataPoints": 58,
    "PollTime": 1553529,
    "APITime": 1088950,
    "ParseTime": 233106,
    "PluginTime": 231401
   }
  },
  {
   "Name": "Rest",
   "Query": "api/private/cli/qos/adaptive-policy-group",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 9,
    "List": [
     "peak_iops",
     "peak_iops_allocation",
     "absolute_min_iops",
     "expected_iops",
     "num_workloads",
     "policy_group",
     "vserver",
     "uuid",
     "expected_iops_allocation"
    ]
   },
   "InstanceInfo": {
    "Count": 4,
    "DataPoints": 36,
    "PollTime": 121145,
    "APITime": 121038,
    "ParseTime": 42,
    "PluginTime": 19
   }
  },
  {
   "Name": "Rest",
   "Query": "api/cluster",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 1,
    "List": [
     "health"
    ]
   },
   "InstanceInfo": {
    "Count": 1,
    "DataPoints": 1,
    "PollTime": 436572,
    "APITime": 436509,
    "ParseTime": 8,
    "PluginTime": 2
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/external_cache",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 19,
    "List": [
     "accesses",
     "hit.total",
     "miss.directory",
     "hit.metadata_file",
     "inserts",
     "invalidates",
     "miss.normal_level_zero",
     "node.name",
     "disk_reads_replaced",
     "hit.directory",
     "hit.indirect",
     "usage",
     "evicts",
     "hit.percent",
     "miss.metadata_file",
     "miss.total",
     "id",
     "hit.normal_level_zero",
     "miss.indirect"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/fcvi",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 13,
    "List": [
     "firmware.loss_of_sync_count",
     "soft_reset_count",
     "firmware.loss_of_signal_count",
     "firmware.systat.discard_frames",
     "hard_reset_count",
     "rdma.write_ops",
     "node.name",
     "firmware.invalid_crc_count",
     "rdma.write_throughput",
     "id",
     "firmware.invalid_transmit_word_count",
     "firmware.link_failure_count",
     "rdma.write_average_latency"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/svm_nfs_v3",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 56,
    "List": [
     "commit.average_latency",
     "link.total",
     "readdir.total",
     "readdirplus.total",
     "remove.total",
     "setattr.total",
     "write.average_latency",
     "access.total",
     "read.average_latency",
     "read_ops",
     "mknod.average_latency",
     "ops",
     "readdirplus.average_latency",
     "write_latency_histogram",
     "id",
     "getattr.total",
     "read_symlink.total",
     "rename.total",
     "access.average_latency",
     "create.average_latency",
     "fsstat.average_latency",
     "throughput",
     "write.total",
     "fsstat.total",
     "create.total",
     "read_latency_histogram",
     "read_throughput",
     "remove.average_latency",
     "mkdir.total",
     "null.total",
     "symlink.average_latency",
     "commit.total",
     "mknod.total",
     "readdir.average_latency",
     "lookup.total",
     "null.average_latency",
     "pathconf.total",
     "write_throughput",
     "getattr.average_latency",
     "latency",
     "lookup.average_latency",
     "read.total",
     "name",
     "fsinfo.total",
     "rename.average_latency",
     "latency_histogram",
     "mkdir.average_latency",
     "write_ops",
     "fsinfo.average_latency",
     "link.average_latency",
     "rmdir.average_latency",
     "setattr.average_latency",
     "pathconf.average_latency",
     "read_symlink.average_latency",
     "rmdir.total",
     "symlink.total"
    ]
   },
   "InstanceInfo": {
    "Count": 13,
    "DataPoints": 2262,
    "PollTime": 358349,
    "APITime": 351695,
    "ParseTime": 5936,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/cluster/ntp/servers",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 2,
    "List": [
     "server",
     "authentication_enabled"
    ]
   },
   "InstanceInfo": {
    "Count": 4,
    "DataPoints": 8,
    "PollTime": 125569,
    "APITime": 125513,
    "ParseTime": 12,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/storage/qtrees",
   "ClientTimeout": "2m0s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 6,
    "List": [
     "id",
     "name",
     "svm.name",
     "volume.name",
     "export_policy.name",
     "security_style"
    ]
   },
   "InstanceInfo": {
    "Count": 44,
    "DataPoints": 352,
    "PollTime": 2661093,
    "APITime": 288268,
    "ParseTime": 1281861,
    "PluginTime": 1090913
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/qtree",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 8,
    "List": [
     "nfs_ops",
     "total_ops",
     "id",
     "node.name",
     "parent_volume.name",
     "svm.name",
     "cifs_ops",
     "internal_ops"
    ]
   },
   "InstanceInfo": {
    "Count": 229,
    "DataPoints": 1832,
    "PollTime": 449266,
    "APITime": 443026,
    "ParseTime": 4351,
    "PluginTime": 1499
   }
  },
  {
   "Name": "Rest",
   "Query": "api/security/accounts",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 5,
    "List": [
     "name",
     "owner.name",
     "locked",
     "password_hash_algorithm",
     "role.name"
    ]
   },
   "InstanceInfo": {
    "Count": 37,
    "DataPoints": 183,
    "PollTime": 4415684,
    "APITime": 2300962,
    "ParseTime": 939,
    "PluginTime": 2113647
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/volume:node",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 36,
    "List": [
     "cifs.read_latency",
     "nfs.other_ops",
     "fcp.read_data",
     "iscsi.other_latency",
     "cifs.read_ops",
     "fcp.write_data",
     "fcp.write_latency",
     "iscsi.write_data",
     "cifs.other_latency",
     "cifs.other_ops",
     "fcp.read_ops",
     "iscsi.read_ops",
     "nfs.read_latency",
     "nfs.write_latency",
     "cifs.write_data",
     "fcp.other_ops",
     "iscsi.other_ops",
     "iscsi.write_ops",
     "nfs.other_latency",
     "nfs.read_data",
     "nfs.write_ops",
     "cifs.write_ops",
     "fcp.write_ops",
     "iscsi.read_data",
     "id",
     "cifs.read_data",
     "read_latency",
     "write_latency",
     "fcp.read_latency",
     "iscsi.read_latency",
     "fcp.other_latency",
     "iscsi.write_latency",
     "nfs.read_ops",
     "nfs.write_data",
     "node.name",
     "cifs.write_latency"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 76,
    "PollTime": 388590,
    "APITime": 388184,
    "ParseTime": 294,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/svm_nfs_v42",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 111,
    "List": [
     "nverify.total",
     "setattr.average_latency",
     "destroy_clientid.total",
     "get_dir_delegation.average_latency",
     "layoutreturn.average_latency",
     "lookupp.total",
     "create_session.total",
     "readdir.total",
     "rename.total",
     "total.throughput",
     "verify.total",
     "create.average_latency",
     "getdeviceinfo.average_latency",
     "putrootfh.total",
     "remove.total",
     "read.average_latency",
     "readlink.average_latency",
     "savefh.total",
     "access.average_latency",
     "lookup.average_latency",
     "total.read_throughput",
     "access.total",
     "getdevicelist.total",
     "putrootfh.average_latency",
     "reclaim_complete.average_latency",
     "layoutreturn.total",
     "want_delegation.average_latency",
     "layoutcommit.average_latency",
     "layoutget.average_latency",
     "secinfo_no_name.average_latency",
     "setattr.total",
     "close.total",
     "commit.average_latency",
     "create.total",
     "delegpurge.total",
     "delegpurge.average_latency",
     "create_session.average_latency",
     "readlink.total",
     "set_ssv.total",
     "readdir.average_latency",
     "set_ssv.average_latency",
     "get_dir_delegation.total",
     "latency",
     "lock.average_latency",
     "openattr.total",
     "putfh.total",
     "rename.average_latency",
     "secinfo_no_name.total",
     "destroy_session.total",
     "nverify.average_latency",
     "open.average_latency",
     "open_downgrade.total",
     "verify.average_latency",
     "svm.name",
     "getfh.total",
     "open_downgrade.average_latency",
     "putpubfh.total",
     "secinfo.average_latency",
     "test_stateid.average_latency",
     "test_stateid.total",
     "want_delegation.total",
     "destroy_clientid.average_latency",
     "getfh.average_latency",
     "lock.total",
     "lockt.total",
     "sequence.total",
     "free_stateid.total",
     "openattr.average_latency",
     "putpubfh.average_latency",
     "sequence.average_latency",
     "read.total",
     "write.total",
     "exchange_id.total",
     "lookup.total",
     "savefh.average_latency",
     "backchannel_ctl.average_latency",
     "layoutcommit.total",
     "restorefh.average_latency",
     "restorefh.total",
     "open.total",
     "remove.average_latency",
     "total_ops",
     "commit.total",
     "delegreturn.average_latency",
     "delegreturn.total",
     "link.average_latency",
     "getdeviceinfo.total",
     "backchannel_ctl.total",
     "bind_conn_to_session.total",
     "reclaim_complete.total",
     "lookupp.average_latency",
     "null.average_latency",
     "close.average_latency",
     "exchange_id.average_latency",
     "getattr.average_latency",
     "locku.total",
     "link.total",
     "lockt.average_latency",
     "total.write_throughput",
     "id",
     "getdevicelist.average_latency",
     "putfh.average_latency",
     "free_stateid.average_latency",
     "layoutget.total",
     "write.average_latency",
     "bind_conn_to_session.average_latency",
     "getattr.total",
     "secinfo.total",
     "destroy_session.average_latency",
     "locku.average_latency",
     "null.total"
    ]
   },
   "InstanceInfo": {
    "Count": 9,
    "DataPoints": 999,
    "PollTime": 157694,
    "APITime": 154093,
    "ParseTime": 2907,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/cluster/peers",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 3,
    "List": [
     "uuid",
     "encryption.state",
     "name"
    ]
   },
   "InstanceInfo": {
    "Count": 1,
    "DataPoints": 3,
    "PollTime": 83061,
    "APITime": 82933,
    "ParseTime": 20,
    "PluginTime": 4
   }
  },
  {
   "Name": "Rest",
   "Query": "api/private/cli/system/health/subsystem",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 4,
    "List": [
     "outstanding_alert_count",
     "suppressed_alert_count",
     "subsystem",
     "health"
    ]
   },
   "InstanceInfo": {
    "Count": 13,
    "DataPoints": 52,
    "PollTime": 273690,
    "APITime": 273566,
    "ParseTime": 59,
    "PluginTime": 8
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/headroom_cpu",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 14,
    "List": [
     "current_latency",
     "ewma.hourly",
     "optimal_point.utilization",
     "name",
     "ewma.daily",
     "ewma.monthly",
     "ewma.weekly",
     "optimal_point.confidence_factor",
     "optimal_point.latency",
     "id",
     "current_utilization",
     "optimal_point.ops",
     "node.name",
     "current_ops"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 96,
    "PollTime": 302465,
    "APITime": 302091,
    "ParseTime": 259,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/svm_nfs_v3:node",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 53,
    "List": [
     "name",
     "mkdir.total",
     "remove.average_latency",
     "null.total",
     "read_throughput",
     "remove.total",
     "rmdir.average_latency",
     "access.average_latency",
     "latency",
     "link.total",
     "mkdir.average_latency",
     "mknod.total",
     "read.total",
     "read_ops",
     "read.average_latency",
     "access.total",
     "commit.total",
     "getattr.average_latency",
     "setattr.total",
     "fsinfo.average_latency",
     "fsstat.average_latency",
     "pathconf.total",
     "symlink.average_latency",
     "link.average_latency",
     "read_symlink.total",
     "readdirplus.total",
     "rename.total",
     "commit.average_latency",
     "rmdir.total",
     "fsinfo.total",
     "pathconf.average_latency",
     "read_symlink.average_latency",
     "readdirplus.average_latency",
     "create.average_latency",
     "fsstat.total",
     "lookup.total",
     "write_ops",
     "mknod.average_latency",
     "getattr.total",
     "lookup.average_latency",
     "null.average_latency",
     "rename.average_latency",
     "write.average_latency",
     "create.total",
     "ops",
     "readdir.average_latency",
     "setattr.average_latency",
     "id",
     "readdir.total",
     "symlink.total",
     "throughput",
     "write.total",
     "write_throughput"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 108,
    "PollTime": 345851,
    "APITime": 344855,
    "ParseTime": 763,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/cloud/targets",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 14,
    "List": [
     "uuid",
     "certificate_validation_enabled",
     "container",
     "name",
     "ipspace.name",
     "scope",
     "ssl_enabled",
     "access_key",
     "port",
     "provider_type",
     "server",
     "authentication_type",
     "owner",
     "used"
    ]
   },
   "InstanceInfo": {
    "Count": 1,
    "DataPoints": 14,
    "PollTime": 165197,
    "APITime": 165120,
    "ParseTime": 27,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/private/cli/snapshot/policy",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 4,
    "List": [
     "policy",
     "vserver",
     "comment",
     "total_schedules"
    ]
   },
   "InstanceInfo": {
    "Count": 3,
    "DataPoints": 12,
    "PollTime": 358864,
    "APITime": 358760,
    "ParseTime": 20,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/support/autosupport",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 3,
    "List": [
     "is_minimal",
     "transport",
     "enabled"
    ]
   },
   "InstanceInfo": {
    "Count": 1,
    "DataPoints": 3,
    "PollTime": 50220,
    "APITime": 50088,
    "ParseTime": 13,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/system:node",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 27,
    "List": [
     "average_processor_busy_percent",
     "cpu_elapsed_time",
     "fcp_data_sent",
     "nvme_fc_data_received",
     "ssd_data_read",
     "nfs_ops",
     "nvme_fc_data_sent",
     "total_data",
     "disk_data_written",
     "fcp_data_received",
     "iscsi_ops",
     "id",
     "fcp_ops",
     "hdd_data_written",
     "nvme_fc_ops",
     "node.name",
     "network_data_sent",
     "total_latency",
     "domain_busy",
     "cifs_ops",
     "cpu_busy",
     "hdd_data_read",
     "network_data_received",
     "total_ops",
     "disk_data_read",
     "memory",
     "ssd_data_written"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 98,
    "PollTime": 155481,
    "APITime": 154845,
    "ParseTime": 489,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/svm_nfs_v41",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 114,
    "List": [
     "readdir.average_latency",
     "total.throughput",
     "null.average_latency",
     "delegpurge.average_latency",
     "layoutget.total",
     "bind_connections_to_session.average_latency",
     "nverify.total",
     "putpubfh.average_latency",
     "lookup.average_latency",
     "getdevicelist.average_latency",
     "secinfo_no_name.average_latency",
     "secinfo_no_name.total",
     "test_stateid.total",
     "getdeviceinfo.average_latency",
     "getfh.total",
     "destroy_session.total",
     "layoutcommit.average_latency",
     "setattr.average_latency",
     "destroy_clientid.total",
     "openattr.total",
     "putpubfh.total",
     "destroy_clientid.average_latency",
     "link.average_latency",
     "putrootfh.total",
     "read.average_latency",
     "close.total",
     "layoutreturn.average_latency",
     "link.total",
     "set_ssv.total",
     "create_session.average_latency",
     "lock.total",
     "putfh.average_latency",
     "want_delegation.average_latency",
     "write.total",
     "delegreturn.total",
     "backchannel_ctl.total",
     "create_session.total",
     "locku.total",
     "read_latency_histogram",
     "total.read_throughput",
     "verify.average_latency",
     "commit.average_latency",
     "layoutget.average_latency",
     "open_downgrade.average_latency",
     "readlink.average_latency",
     "readlink.total",
     "reclaim_complete.total",
     "restorefh.average_latency",
     "savefh.total",
     "getdevicelist.total",
     "total.write_throughput",
     "total_ops",
     "sequence.average_latency",
     "close.average_latency",
     "lockt.average_latency",
     "open.average_latency",
     "putrootfh.average_latency",
     "rename.average_latency",
     "id",
     "get_dir_delegation.average_latency",
     "getattr.total",
     "null.total",
     "write.average_latency",
     "name",
     "secinfo.total",
     "create.average_latency",
     "savefh.average_latency",
     "open_downgrade.total",
     "remove.total",
     "restorefh.total",
     "locku.average_latency",
     "lookupp.total",
     "secinfo.average_latency",
     "get_dir_delegation.total",
     "lookupp.average_latency",
     "read.total",
     "test_stateid.average_latency",
     "exchange_id.total",
     "readdir.total",
     "total.latency_histogram",
     "exchange_id.average_latency",
     "bind_connections_to_session.total",
     "getattr.average_latency",
     "backchannel_ctl.average_latency",
     "lockt.total",
     "latency",
     "open.total",
     "destroy_session.average_latency",
     "putfh.total",
     "reclaim_complete.average_latency",
     "layoutreturn.total",
     "lookup.total",
     "sequence.total",
     "layoutcommit.total",
     "free_stateid.total",
     "nverify.average_latency",
     "verify.total",
     "delegreturn.average_latency",
     "remove.average_latency",
     "set_ssv.average_latency",
     "access.average_latency",
     "lock.average_latency",
     "setattr.total",
     "write_latency_histogram",
     "access.total",
     "create.total",
     "delegpurge.total",
     "getdeviceinfo.total",
     "getfh.average_latency",
     "commit.total",
     "openattr.average_latency",
     "rename.total",
     "want_delegation.total",
     "free_stateid.average_latency"
    ]
   },
   "InstanceInfo": {
    "Count": 9,
    "DataPoints": 2079,
    "PollTime": 210542,
    "APITime": 205589,
    "ParseTime": 4409,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/storage/disks",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 17,
    "List": [
     "stats.power_on_hours",
     "uid",
     "outage.reason",
     "serial_number",
     "shelf.uid",
     "node.uuid",
     "state",
     "stats.average_latency",
     "bay",
     "container_type",
     "home_node.name",
     "name",
     "model",
     "bytes_per_sector",
     "sector_count",
     "usable_size",
     "node.name"
    ]
   },
   "InstanceInfo": {
    "Count": 23,
    "DataPoints": 460,
    "PollTime": 531096,
    "APITime": 153322,
    "ParseTime": 377662,
    "PluginTime": 54
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/disk:constituent",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 25,
    "List": [
     "physical_disk_name",
     "raid.type",
     "user_read_latency",
     "cp_read_chain",
     "cp_read_count",
     "disk_busy_percent",
     "io_queued",
     "user_read_block_count",
     "id",
     "raid_group",
     "capacity",
     "cp_read_latency",
     "user_read_count",
     "name",
     "speed",
     "io_pending",
     "user_write_block_count",
     "user_write_latency",
     "total_data",
     "user_write_count",
     "node.name",
     "physical_disk_id",
     "total_transfer_count",
     "user_read_chain",
     "user_write_chain"
    ]
   },
   "InstanceInfo": {
    "Count": 64,
    "DataPoints": 1728,
    "PollTime": 650551,
    "APITime": 437801,
    "ParseTime": 15459,
    "PluginTime": 196871
   }
  },
  {
   "Name": "Rest",
   "Query": "api/cluster",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 2,
    "List": [
     "uuid",
     "name"
    ]
   },
   "InstanceInfo": {
    "Count": 1,
    "DataPoints": 2,
    "PollTime": 2538656,
    "APITime": 451859,
    "ParseTime": 12,
    "PluginTime": 2086722
   }
  },
  {
   "Name": "Rest",
   "Query": "api/storage/shelves",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 8,
    "List": [
     "disk_count",
     "serial_number",
     "manufacturer.name",
     "model",
     "module_type",
     "name",
     "state",
     "uid"
    ]
   },
   "InstanceInfo": {
    "Count": 1,
    "DataPoints": 9,
    "PollTime": 324451,
    "APITime": 119224,
    "ParseTime": 205161,
    "PluginTime": 10
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/qos",
   "ClientTimeout": "1m30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 16,
    "List": [
     "sequential_writes_percent",
     "total_data",
     "write_ops",
     "other_ops",
     "read_latency",
     "sequential_reads_percent",
     "latency",
     "read_data",
     "write_data",
     "write_latency",
     "uuid",
     "ops",
     "read_io_type_percent",
     "name",
     "concurrency",
     "read_ops"
    ]
   },
   "InstanceInfo": {
    "Count": 70,
    "DataPoints": 174,
    "PollTime": 213437,
    "APITime": 212668,
    "ParseTime": 370,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/storage/luns",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 9,
    "List": [
     "uuid",
     "name",
     "space.used",
     "location.node.name",
     "location.qtree.name",
     "location.volume.name",
     "status.state",
     "svm.name",
     "space.size"
    ]
   },
   "InstanceInfo": {
    "Count": 13,
    "DataPoints": 104,
    "PollTime": 159138,
    "APITime": 158781,
    "ParseTime": 254,
    "PluginTime": 28
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/lun",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 22,
    "List": [
     "write_data",
     "id",
     "enospc",
     "read_align_histogram",
     "remote_bytes",
     "writesame_unmap_requests",
     "average_read_latency",
     "average_xcopy_latency",
     "read_ops",
     "write_partial_blocks",
     "svm.name",
     "queue_full",
     "read_partial_blocks",
     "xcopy_requests",
     "unmap_requests",
     "write_align_histogram",
     "write_ops",
     "writesame_requests",
     "average_write_latency",
     "caw_requests",
     "read_data",
     "remote_ops"
    ]
   },
   "InstanceInfo": {
    "Count": 4,
    "DataPoints": 152,
    "PollTime": 141289,
    "APITime": 140499,
    "ParseTime": 512,
    "PluginTime": 61
   }
  },
  {
   "Name": "Rest",
   "Query": "api/network/ip/routes",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 9,
    "List": [
     "destination.family",
     "svm.name",
     "{interfaces.#.name,interfaces.#.ip.address}",
     "uuid",
     "destination.address",
     "destination.netmask",
     "gateway",
     "ipspace.name",
     "scope"
    ]
   },
   "InstanceInfo": {
    "Count": 13,
    "DataPoints": 114,
    "PollTime": 88311,
    "APITime": 87437,
    "ParseTime": 556,
    "PluginTime": 193
   }
  },
  {
   "Name": "Rest",
   "Query": "api/private/cli/vserver",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 5,
    "List": [
     "vserver",
     "anti_ransomware_default_volume_state",
     "operational_state",
     "type",
     "uuid"
    ]
   },
   "InstanceInfo": {
    "Count": 28,
    "DataPoints": 162,
    "PollTime": 2422001,
    "APITime": 267220,
    "ParseTime": 1856872,
    "PluginTime": 297830
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/svm_nfs_v4:node",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 83,
    "List": [
     "open_confirm.total",
     "rename.total",
     "setclientid_confirm.average_latency",
     "delegpurge.average_latency",
     "lookupp.average_latency",
     "remove.total",
     "setclientid.average_latency",
     "lookupp.total",
     "readdir.average_latency",
     "getfh.average_latency",
     "null.average_latency",
     "putrootfh.total",
     "release_lock_owner.average_latency",
     "renew.total",
     "setclientid_confirm.total",
     "access.average_latency",
     "close.average_latency",
     "delegreturn.average_latency",
     "rename.average_latency",
     "id",
     "create.total",
     "lookup.total",
     "openattr.average_latency",
     "read.average_latency",
     "setclientid.total",
     "total.read_throughput",
     "total_ops",
     "getattr.total",
     "lockt.average_latency",
     "verify.average_latency",
     "link.average_latency",
     "locku.total",
     "create.average_latency",
     "link.total",
     "open_downgrade.total",
     "readlink.average_latency",
     "readlink.total",
     "restorefh.total",
     "close.total",
     "commit.average_latency",
     "savefh.total",
     "total.throughput",
     "null.total",
     "nverify.total",
     "open_confirm.average_latency",
     "write.average_latency",
     "getfh.total",
     "lookup.average_latency",
     "nverify.average_latency",
     "putpubfh.average_latency",
     "read.total",
     "release_lock_owner.total",
     "secinfo.average_latency",
     "secinfo.total",
     "node.name",
     "getattr.average_latency",
     "total.write_throughput",
     "putfh.average_latency",
     "commit.total",
     "open.average_latency",
     "open.total",
     "putfh.total",
     "write.total",
     "delegreturn.total",
     "latency",
     "remove.average_latency",
     "savefh.average_latency",
     "delegpurge.total",
     "lock.average_latency",
     "readdir.total",
     "renew.average_latency",
     "setattr.total",
     "locku.average_latency",
     "putpubfh.total",
     "open_downgrade.average_latency",
     "verify.total",
     "access.total",
     "lockt.total",
     "putrootfh.average_latency",
     "restorefh.average_latency",
     "setattr.average_latency",
     "lock.total",
     "openattr.total"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 166,
    "PollTime": 139134,
    "APITime": 138307,
    "ParseTime": 612,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/token_manager",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 11,
    "List": [
     "token_copy.bytes",
     "token_copy.failures",
     "token_zero.bytes",
     "token_zero.successes",
     "id",
     "node.name",
     "token_copy.successes",
     "token_create.bytes",
     "token_create.failures",
     "token_create.successes",
     "token_zero.failures"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 22,
    "PollTime": 157804,
    "APITime": 157680,
    "ParseTime": 54,
    "PluginTime": 4
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/copy_manager",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 7,
    "List": [
     "ontap_copy_subsystem_current_copy_count",
     "spince_current_copy_count",
     "system_continuous_engineering_current_copy_count",
     "id",
     "name",
     "KB_copied",
     "block_copy_engine_current_copy_count"
    ]
   },
   "InstanceInfo": {
    "Count": 25,
    "DataPoints": 175,
    "PollTime": 372155,
    "APITime": 368898,
    "ParseTime": 3135,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/security",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 3,
    "List": [
     "fips.enabled",
     "management_protocols.rsh_enabled",
     "management_protocols.telnet_enabled"
    ]
   },
   "InstanceInfo": {
    "Count": 1,
    "DataPoints": 3,
    "PollTime": 187081,
    "APITime": 187005,
    "ParseTime": 28,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/private/cli/snapmirror",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 29,
    "List": [
     "total_transfer_time_secs",
     "update_successful_count",
     "break_failed_count",
     "destination_volume",
     "policy_type",
     "relationship_type",
     "newest_snapshot_timestamp",
     "resync_failed_count",
     "destination_volume_node",
     "healthy",
     "last_transfer_type",
     "source_vserver",
     "last_transfer_size",
     "relationship_id",
     "total_transfer_bytes",
     "break_successful_count",
     "schedule",
     "lag_time",
     "last_transfer_duration",
     "update_failed_count",
     "destination_vserver",
     "relationship_group_type",
     "source_volume",
     "status",
     "destination_path",
     "unhealthy_reason",
     "last_transfer_end_timestamp",
     "resync_successful_count",
     "cg_item_mappings"
    ]
   },
   "InstanceInfo": {
    "Count": 6,
    "DataPoints": 116,
    "PollTime": 761309,
    "APITime": 600524,
    "ParseTime": 160711,
    "PluginTime": 16
   }
  },
  {
   "Name": "Rest",
   "Query": "api/support/auto-update",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 1,
    "List": [
     "enabled"
    ]
   },
   "InstanceInfo": {
    "Count": 1,
    "DataPoints": 1,
    "PollTime": 130755,
    "APITime": 130646,
    "ParseTime": 4,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/headroom_aggregate",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 14,
    "List": [
     "name",
     "ewma.daily",
     "ewma.monthly",
     "current_latency",
     "ewma.weekly",
     "optimal_point.ops",
     "ewma.hourly",
     "optimal_point.confidence_factor",
     "optimal_point.latency",
     "current_utilization",
     "optimal_point.utilization",
     "id",
     "node.name",
     "current_ops"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 96,
    "PollTime": 245928,
    "APITime": 244677,
    "ParseTime": 436,
    "PluginTime": 7
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/path",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 10,
    "List": [
     "read_data",
     "read_iops",
     "total_data",
     "total_iops",
     "id",
     "node.name",
     "read_latency",
     "write_data",
     "write_iops",
     "write_latency"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/wafl_hya_sizer",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 3,
    "List": [
     "id",
     "node.name",
     "cache_stats"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/protocols/cifs/sessions",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 14,
    "List": [
     "identifier",
     "large_mtu",
     "mapped_unix_user",
     "node.name",
     "server_ip",
     "smb_encryption",
     "smb_signing",
     "client_ip",
     "connection_id",
     "connection_count",
     "authentication",
     "user",
     "protocol",
     "svm.name"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/cluster/nodes",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 14,
    "List": [
     "model",
     "state",
     "name",
     "controller.failed_power_supply.message.message",
     "uuid",
     "version.full",
     "controller.failed_fan.count",
     "controller.failed_power_supply.count",
     "location",
     "serial_number",
     "controller.cpu.firmware_release",
     "controller.failed_fan.message.message",
     "controller.over_temperature",
     "uptime"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 40,
    "PollTime": 588253,
    "APITime": 310107,
    "ParseTime": 278009,
    "PluginTime": 19,
    "Ids": [
     {
      "serial-number": "721802000259",
      "system-id": "0537123843"
     },
     {
      "serial-number": "721802000260",
      "system-id": "0537124012"
     }
    ]
   }
  },
  {
   "Name": "Rest",
   "Query": "api/private/cli/qos/policy-group",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 7,
    "List": [
     "throughput_policy",
     "vserver",
     "uuid",
     "class",
     "is_shared",
     "num_workloads",
     "policy_group"
    ]
   },
   "InstanceInfo": {
    "Count": 11,
    "DataPoints": 77,
    "PollTime": 85223,
    "APITime": 85035,
    "ParseTime": 74,
    "PluginTime": 58
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/nic_common",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 16,
    "List": [
     "link_current_state",
     "receive_crc_errors",
     "transmit_hw_errors",
     "node.name",
     "link_up_to_down",
     "receive_length_errors",
     "transmit_bytes",
     "transmit_errors",
     "id",
     "receive_bytes",
     "receive_errors",
     "receive_total_errors",
     "transmit_total_errors",
     "link_speed",
     "type",
     "receive_alignment_errors"
    ]
   },
   "InstanceInfo": {
    "Count": 14,
    "DataPoints": 224,
    "PollTime": 193284,
    "APITime": 190079,
    "ParseTime": 2969,
    "PluginTime": 47
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/smb2",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 35,
    "List": [
     "tree_connect_latency",
     "query_directory_latency",
     "create_ops",
     "session_setup_latency",
     "set_info_latency_histogram",
     "close_latency_histogram",
     "oplock_break_latency",
     "session_setup_latency_histogram",
     "id",
     "lock_ops",
     "query_directory_ops",
     "query_info_ops",
     "set_info_ops",
     "svm.name",
     "close_latency",
     "lock_latency",
     "negotiate_latency",
     "read_latency",
     "node.name",
     "create_latency_histogram",
     "negotiate_ops",
     "query_directory_latency_histogram",
     "query_info_latency_histogram",
     "create_latency",
     "query_info_latency",
     "session_setup_ops",
     "oplock_break_latency_histogram",
     "lock_latency_histogram",
     "oplock_break_ops",
     "read_ops",
     "set_info_latency",
     "tree_connect_ops",
     "write_latency",
     "write_ops",
     "close_ops"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/security/audit/destinations",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 1,
    "List": [
     "protocol"
    ]
   },
   "InstanceInfo": {
    "Count": 1,
    "DataPoints": 1,
    "PollTime": 169996,
    "APITime": 169898,
    "ParseTime": 12,
    "PluginTime": 6
   }
  },
  {
   "Name": "Rest",
   "Query": "api/security/certificates",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 8,
    "List": [
     "type",
     "expiry_time",
     "uuid",
     "name",
     "public_certificate",
     "scope",
     "serial_number",
     "svm.name"
    ]
   },
   "InstanceInfo": {
    "Count": 6,
    "DataPoints": 42,
    "PollTime": 3807190,
    "APITime": 3724532,
    "ParseTime": 276,
    "PluginTime": 82288
   }
  },
  {
   "Name": "Rest",
   "Query": "api/cluster/sensors",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 13,
    "List": [
     "index",
     "critical_high_threshold",
     "threshold_state",
     "type",
     "critical_low_threshold",
     "discrete_state",
     "node.name",
     "warning_high_threshold",
     "discrete_value",
     "name",
     "value_units",
     "warning_low_threshold",
     "value"
    ]
   },
   "InstanceInfo": {
    "Count": 186,
    "DataPoints": 1558,
    "PollTime": 1864342,
    "APITime": 1819744,
    "ParseTime": 4453,
    "PluginTime": 40093
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/host_adapter",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 4,
    "List": [
     "id",
     "node.name",
     "bytes_read",
     "bytes_written"
    ]
   },
   "InstanceInfo": {
    "Count": 8,
    "DataPoints": 32,
    "PollTime": 153351,
    "APITime": 152898,
    "ParseTime": 120,
    "PluginTime": 10
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/svm_nfs_v41:node",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 111,
    "List": [
     "id",
     "lock.total",
     "write.average_latency",
     "write.total",
     "exchange_id.total",
     "getattr.total",
     "layoutcommit.total",
     "read.average_latency",
     "reclaim_complete.total",
     "openattr.total",
     "readdir.total",
     "reclaim_complete.average_latency",
     "free_stateid.total",
     "putpubfh.average_latency",
     "putpubfh.total",
     "setattr.average_latency",
     "null.average_latency",
     "delegpurge.average_latency",
     "delegreturn.total",
     "exchange_id.average_latency",
     "free_stateid.average_latency",
     "link.total",
     "putrootfh.total",
     "rename.average_latency",
     "sequence.total",
     "commit.average_latency",
     "create_session.total",
     "lock.average_latency",
     "set_ssv.average_latency",
     "close.total",
     "nverify.total",
     "backchannel_ctl.total",
     "destroy_session.average_latency",
     "open_downgrade.total",
     "putfh.total",
     "bind_connections_to_session.average_latency",
     "putrootfh.average_latency",
     "set_ssv.total",
     "setattr.total",
     "total.read_throughput",
     "restorefh.total",
     "total_ops",
     "node.name",
     "lookup.average_latency",
     "savefh.total",
     "total.write_throughput",
     "close.average_latency",
     "delegpurge.total",
     "openattr.average_latency",
     "rename.total",
     "want_delegation.total",
     "access.total",
     "create.average_latency",
     "secinfo.total",
     "destroy_session.total",
     "getdevicelist.total",
     "locku.total",
     "open.average_latency",
     "secinfo_no_name.average_latency",
     "bind_connections_to_session.total",
     "getdevicelist.average_latency",
     "null.total",
     "open_downgrade.average_latency",
     "access.average_latency",
     "layoutreturn.average_latency",
     "link.average_latency",
     "savefh.average_latency",
     "readdir.average_latency",
     "readlink.total",
     "verify.total",
     "want_delegation.average_latency",
     "getdeviceinfo.total",
     "getfh.average_latency",
     "lockt.total",
     "sequence.average_latency",
     "destroy_clientid.average_latency",
     "getdeviceinfo.average_latency",
     "nverify.average_latency",
     "remove.total",
     "secinfo.average_latency",
     "commit.total",
     "create_session.average_latency",
     "layoutcommit.average_latency",
     "layoutreturn.total",
     "restorefh.average_latency",
     "test_stateid.average_latency",
     "backchannel_ctl.average_latency",
     "delegreturn.average_latency",
     "getfh.total",
     "latency",
     "layoutget.total",
     "lookupp.total",
     "readlink.average_latency",
     "secinfo_no_name.total",
     "lookup.total",
     "putfh.average_latency",
     "remove.average_latency",
     "test_stateid.total",
     "total.throughput",
     "verify.average_latency",
     "destroy_clientid.total",
     "get_dir_delegation.average_latency",
     "getattr.average_latency",
     "layoutget.average_latency",
     "lockt.average_latency",
     "lookupp.average_latency",
     "read.total",
     "open.total",
     "create.total",
     "get_dir_delegation.total",
     "locku.average_latency"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 222,
    "PollTime": 159998,
    "APITime": 158956,
    "ParseTime": 786,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/nvmf_lif",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 15,
    "List": [
     "node.name",
     "svm.name",
     "average_other_latency",
     "id",
     "total_ops",
     "write_ops",
     "port_id",
     "average_read_latency",
     "read_data",
     "name",
     "average_write_latency",
     "other_ops",
     "read_ops",
     "write_data",
     "average_latency"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/storage/qos/workloads",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 3,
    "List": [
     "uuid",
     "name",
     "workload_class"
    ]
   },
   "InstanceInfo": {
    "Count": 282,
    "DataPoints": 846,
    "PollTime": 1579862,
    "APITime": 1578146,
    "ParseTime": 1629,
    "PluginTime": 0
   }
  },
  {
   "Name": "Rest",
   "Query": "api/storage/volumes",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 3,
    "List": [
     "name",
     "svm.name",
     "uuid"
    ]
   },
   "InstanceInfo": {
    "Count": 4,
    "DataPoints": 12,
    "PollTime": 1893780,
    "APITime": 217014,
    "ParseTime": 26,
    "PluginTime": 1676688
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/wafl",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 22,
    "List": [
     "node.name",
     "average_non_wafl_msg_latency",
     "cp_count",
     "cp_phase_times",
     "memory_free",
     "reads_from_cloud_s2c_bin",
     "reads_from_disk",
     "reads_from_fc_miss",
     "id",
     "reads_from_cache",
     "reads_from_cloud",
     "reads_from_external_cache",
     "total_cp_msecs",
     "average_msg_latency",
     "average_replication_msg_latency",
     "memory_used",
     "non_wafl_msg_total",
     "replication_msg_total",
     "total_cp_util",
     "msg_total",
     "read_io_type",
     "reads_from_ssd"
    ]
   },
   "InstanceInfo": {
    "Count": 2,
    "DataPoints": 184,
    "PollTime": 244710,
    "APITime": 243646,
    "ParseTime": 585,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/svm_nfs_v4",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 86,
    "List": [
     "getfh.average_latency",
     "lookup.total",
     "lookupp.average_latency",
     "putpubfh.total",
     "getattr.average_latency",
     "open.total",
     "total.throughput",
     "delegreturn.total",
     "lock.total",
     "null.average_latency",
     "openattr.total",
     "rename.total",
     "savefh.total",
     "name",
     "access.average_latency",
     "close.average_latency",
     "create.total",
     "delegpurge.total",
     "lockt.total",
     "readdir.average_latency",
     "release_lock_owner.average_latency",
     "setclientid_confirm.total",
     "lock.average_latency",
     "locku.average_latency",
     "lookupp.total",
     "remove.total",
     "setclientid.average_latency",
     "latency",
     "putpubfh.average_latency",
     "renew.average_latency",
     "restorefh.average_latency",
     "savefh.average_latency",
     "setattr.total",
     "verify.average_latency",
     "commit.total",
     "nverify.average_latency",
     "open_downgrade.total",
     "renew.total",
     "write.total",
     "getattr.total",
     "getfh.total",
     "openattr.average_latency",
     "readlink.average_latency",
     "remove.average_latency",
     "setattr.average_latency",
     "total_ops",
     "read.total",
     "rename.average_latency",
     "total.write_throughput",
     "verify.total",
     "create.average_latency",
     "open_confirm.average_latency",
     "restorefh.total",
     "secinfo.average_latency",
     "setclientid_confirm.average_latency",
     "link.average_latency",
     "null.total",
     "open.average_latency",
     "write_latency_histogram",
     "access.total",
     "nverify.total",
     "putfh.total",
     "putrootfh.average_latency",
     "read_latency_histogram",
     "readlink.total",
     "locku.total",
     "lookup.average_latency",
     "open_downgrade.average_latency",
     "read.average_latency",
     "release_lock_owner.total",
     "setclientid.total",
     "total.read_throughput",
     "id",
     "commit.average_latency",
     "putfh.average_latency",
     "write.average_latency",
     "close.total",
     "delegpurge.average_latency",
     "delegreturn.average_latency",
     "link.total",
     "lockt.average_latency",
     "putrootfh.total",
     "readdir.total",
     "total.latency_histogram",
     "open_confirm.total",
     "secinfo.total"
    ]
   },
   "InstanceInfo": {
    "Count": 10,
    "DataPoints": 2030,
    "PollTime": 190175,
    "APITime": 185410,
    "ParseTime": 4250,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/vscan",
   "ClientTimeout": "30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "1m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 6,
    "List": [
     "scanner.stats_percent_cpu_used",
     "scanner.stats_percent_mem_used",
     "scanner.stats_percent_network_used",
     "id",
     "scan.latency",
     "scan.request_dispatched_rate"
    ]
   },
   "InstanceInfo": {
    "Count": 0,
    "DataPoints": 0,
    "PollTime": 0,
    "APITime": 0,
    "ParseTime": 0,
    "PluginTime": 0
   }
  },
  {
   "Name": "RestPerf",
   "Query": "api/cluster/counter/tables/qos_detail",
   "ClientTimeout": "1m30s",
   "Schedules": [
    {
     "Name": "counter",
     "Schedule": "20m"
    },
    {
     "Name": "instance",
     "Schedule": "10m"
    },
    {
     "Name": "data",
     "Schedule": "3m"
    }
   ],
   "Exporters": [
    "Prometheus"
   ],
   "Counters": {
    "Count": 5,
    "List": [
     "wait_time",
     "id",
     "node.name",
     "service_time",
     "visits"
    ]
   },
   "InstanceInfo": {
    "Count": 70,
    "DataPoints": 995,
    "PollTime": 407530,
    "APITime": 256856,
    "ParseTime": 148897,
    "PluginTime": 0
   }
  }
 ]
}
```