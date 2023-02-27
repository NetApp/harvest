This document describes implementation details about ONTAP's REST performance metrics endpoints, including how we built the Harvest RESTPerf collectors. 

!!! warning

    These are implemenation details about ONTAP's REST performance metrics. You do not need to understand any of this to use Harvest. If you want to know how to **use** or **configure** Harvest's REST collectors, checkout the [Rest Collector](../configure-rest.md) documentation instead. If you're interested in the gory details. Read on. 

## Introduction

ONTAP REST metrics were introduced in [ONTAP `9.11.1`](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-strategy.md#reference) and included parity with Harvest-collected ZAPI performance metrics by ONTAP `9.12.1`.

## Performance REST queries

Mapping table

| ZAPI                                  | REST                                      | Comment                              |
|---------------------------------------|-------------------------------------------|--------------------------------------|
| `perf-object-counter-list-info`       | `/api/cluster/counter/tables`             | returns counter tables and schemas   |
| `perf-object-instance-list-info-iter` | `/api/cluster/counter/tables/{name}/rows` | returns instances and counter values |
| `perf-object-get-instances`           | `/api/cluster/counter/tables/{name}/rows` | returns instances and counter values |

Performance REST responses include `properties` and `counters`. Counters are metric-like, while properties include instance attributes.

### Examples

### Ask ONTAP for all resources that report performance metrics

```bash
curl 'https://10.193.48.154/api/cluster/counter/tables'
```

<details><summary>Response</summary>
<p>

```json
{
  "records": [
    {
      "name": "copy_manager",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/copy_manager"
        }
      }
    },
    {
      "name": "copy_manager:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/copy_manager%3Aconstituent"
        }
      }
    },
    {
      "name": "disk",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/disk"
        }
      }
    },
    {
      "name": "disk:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/disk%3Aconstituent"
        }
      }
    },
    {
      "name": "disk:raid_group",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/disk%3Araid_group"
        }
      }
    },
    {
      "name": "external_cache",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/external_cache"
        }
      }
    },
    {
      "name": "fcp",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/fcp"
        }
      }
    },
    {
      "name": "fcp:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/fcp%3Anode"
        }
      }
    },
    {
      "name": "fcp_lif",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/fcp_lif"
        }
      }
    },
    {
      "name": "fcp_lif:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/fcp_lif%3Anode"
        }
      }
    },
    {
      "name": "fcp_lif:port",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/fcp_lif%3Aport"
        }
      }
    },
    {
      "name": "fcp_lif:svm",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/fcp_lif%3Asvm"
        }
      }
    },
    {
      "name": "fcvi",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/fcvi"
        }
      }
    },
    {
      "name": "headroom_aggregate",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/headroom_aggregate"
        }
      }
    },
    {
      "name": "headroom_cpu",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/headroom_cpu"
        }
      }
    },
    {
      "name": "host_adapter",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/host_adapter"
        }
      }
    },
    {
      "name": "iscsi_lif",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/iscsi_lif"
        }
      }
    },
    {
      "name": "iscsi_lif:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/iscsi_lif%3Anode"
        }
      }
    },
    {
      "name": "iscsi_lif:svm",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/iscsi_lif%3Asvm"
        }
      }
    },
    {
      "name": "lif",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/lif"
        }
      }
    },
    {
      "name": "lif:svm",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/lif%3Asvm"
        }
      }
    },
    {
      "name": "lun",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/lun"
        }
      }
    },
    {
      "name": "lun:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/lun%3Aconstituent"
        }
      }
    },
    {
      "name": "lun:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/lun%3Anode"
        }
      }
    },
    {
      "name": "namespace",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/namespace"
        }
      }
    },
    {
      "name": "namespace:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/namespace%3Aconstituent"
        }
      }
    },
    {
      "name": "nfs_v4_diag",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/nfs_v4_diag"
        }
      }
    },
    {
      "name": "nic_common",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/nic_common"
        }
      }
    },
    {
      "name": "nvmf_lif",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/nvmf_lif"
        }
      }
    },
    {
      "name": "nvmf_lif:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/nvmf_lif%3Aconstituent"
        }
      }
    },
    {
      "name": "nvmf_lif:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/nvmf_lif%3Anode"
        }
      }
    },
    {
      "name": "nvmf_lif:port",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/nvmf_lif%3Aport"
        }
      }
    },
    {
      "name": "object_store_client_op",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/object_store_client_op"
        }
      }
    },
    {
      "name": "path",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/path"
        }
      }
    },
    {
      "name": "processor",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/processor"
        }
      }
    },
    {
      "name": "processor:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/processor%3Anode"
        }
      }
    },
    {
      "name": "qos",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/qos"
        }
      }
    },
    {
      "name": "qos:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/qos%3Aconstituent"
        }
      }
    },
    {
      "name": "qos:policy_group",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/qos%3Apolicy_group"
        }
      }
    },
    {
      "name": "qos_detail",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/qos_detail"
        }
      }
    },
    {
      "name": "qos_detail_volume",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/qos_detail_volume"
        }
      }
    },
    {
      "name": "qos_volume",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/qos_volume"
        }
      }
    },
    {
      "name": "qos_volume:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/qos_volume%3Aconstituent"
        }
      }
    },
    {
      "name": "qtree",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/qtree"
        }
      }
    },
    {
      "name": "qtree:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/qtree%3Aconstituent"
        }
      }
    },
    {
      "name": "svm_cifs",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_cifs"
        }
      }
    },
    {
      "name": "svm_cifs:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_cifs%3Aconstituent"
        }
      }
    },
    {
      "name": "svm_cifs:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_cifs%3Anode"
        }
      }
    },
    {
      "name": "svm_nfs_v3",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v3"
        }
      }
    },
    {
      "name": "svm_nfs_v3:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v3%3Aconstituent"
        }
      }
    },
    {
      "name": "svm_nfs_v3:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v3%3Anode"
        }
      }
    },
    {
      "name": "svm_nfs_v4",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v4"
        }
      }
    },
    {
      "name": "svm_nfs_v41",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v41"
        }
      }
    },
    {
      "name": "svm_nfs_v41:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v41%3Aconstituent"
        }
      }
    },
    {
      "name": "svm_nfs_v41:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v41%3Anode"
        }
      }
    },
    {
      "name": "svm_nfs_v42",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v42"
        }
      }
    },
    {
      "name": "svm_nfs_v42:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v42%3Aconstituent"
        }
      }
    },
    {
      "name": "svm_nfs_v42:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v42%3Anode"
        }
      }
    },
    {
      "name": "svm_nfs_v4:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v4%3Aconstituent"
        }
      }
    },
    {
      "name": "svm_nfs_v4:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/svm_nfs_v4%3Anode"
        }
      }
    },
    {
      "name": "system",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/system"
        }
      }
    },
    {
      "name": "system:constituent",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/system%3Aconstituent"
        }
      }
    },
    {
      "name": "system:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/system%3Anode"
        }
      }
    },
    {
      "name": "token_manager",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/token_manager"
        }
      }
    },
    {
      "name": "volume",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/volume"
        }
      }
    },
    {
      "name": "volume:node",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/volume%3Anode"
        }
      }
    },
    {
      "name": "volume:svm",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/volume%3Asvm"
        }
      }
    },
    {
      "name": "wafl",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/wafl"
        }
      }
    },
    {
      "name": "wafl_comp_aggr_vol_bin",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/wafl_comp_aggr_vol_bin"
        }
      }
    },
    {
      "name": "wafl_hya_per_aggregate",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/wafl_hya_per_aggregate"
        }
      }
    },
    {
      "name": "wafl_hya_sizer",
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/wafl_hya_sizer"
        }
      }
    }
  ],
  "num_records": 71,
  "_links": {
    "self": {
      "href": "/api/cluster/counter/tables/"
    }
  }
}
```

</p>
</details>

### Node performance metrics metadata

Ask ONTAP to return the schema for `system:node`. This will include the name, description, and metadata for all counters associated with `system:node`.

```bash
curl 'https://10.193.48.154/api/cluster/counter/tables/system:node?return_records=true '
```

<details><summary>Response</summary>
<p>

```json
{
  "name": "system:node",
  "description": "The System table reports general system activity. This includes global throughput for the main services, I/O latency, and CPU activity. The alias name for system:node is system_node.",
  "counter_schemas": [
    {
      "name": "average_processor_busy_percent",
      "description": "Average processor utilization across all processors in the system",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cpu_elapsed_time"
      }
    },
    {
      "name": "cifs_ops",
      "description": "Number of CIFS operations per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "cp",
      "description": "CP time rate",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cpu_elapsed_time"
      }
    },
    {
      "name": "cp_time",
      "description": "Processor time in CP",
      "type": "delta",
      "unit": "microsec"
    },
    {
      "name": "cpu_busy",
      "description": "System CPU resource utilization. Returns a computed percentage for the default CPU field. Basically computes a 'cpu usage summary' value which indicates how 'busy' the system is based upon the most heavily utilized domain. The idea is to determine the amount of available CPU until we're limited by either a domain maxing out OR we exhaust all available idle CPU cycles, whichever occurs first.",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cpu_elapsed_time"
      }
    },
    {
      "name": "cpu_elapsed_time",
      "description": "Elapsed time since boot",
      "type": "delta",
      "unit": "microsec"
    },
    {
      "name": "disk_data_read",
      "description": "Number of disk kilobytes (KB) read per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "disk_data_written",
      "description": "Number of disk kilobytes (KB) written per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "domain_busy",
      "description": "Array of processor time in percentage spent in various domains",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cpu_elapsed_time"
      }
    },
    {
      "name": "domain_shared",
      "description": "Array of processor time in percentage spent in various shared domains",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cpu_elapsed_time"
      }
    },
    {
      "name": "dswitchto_cnt",
      "description": "Array of processor time in percentage spent in domain switch",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cpu_elapsed_time"
      }
    },
    {
      "name": "fcp_data_received",
      "description": "Number of FCP kilobytes (KB) received per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "fcp_data_sent",
      "description": "Number of FCP kilobytes (KB) sent per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "fcp_ops",
      "description": "Number of FCP operations per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "hard_switches",
      "description": "Number of context switches per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "hdd_data_read",
      "description": "Number of HDD Disk kilobytes (KB) read per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "hdd_data_written",
      "description": "Number of HDD kilobytes (KB) written per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "idle",
      "description": "Processor idle rate percentage",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cpu_elapsed_time"
      }
    },
    {
      "name": "idle_time",
      "description": "Processor idle time",
      "type": "delta",
      "unit": "microsec"
    },
    {
      "name": "instance_name",
      "description": "Node name",
      "type": "string",
      "unit": "none"
    },
    {
      "name": "interrupt",
      "description": "Processor interrupt rate percentage",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cpu_elapsed_time"
      }
    },
    {
      "name": "interrupt_in_cp",
      "description": "Processor interrupt rate percentage",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cp_time"
      }
    },
    {
      "name": "interrupt_in_cp_time",
      "description": "Processor interrupt in CP time",
      "type": "delta",
      "unit": "microsec"
    },
    {
      "name": "interrupt_num",
      "description": "Processor interrupt number",
      "type": "delta",
      "unit": "none"
    },
    {
      "name": "interrupt_num_in_cp",
      "description": "Number of processor interrupts in CP",
      "type": "delta",
      "unit": "none"
    },
    {
      "name": "interrupt_time",
      "description": "Processor interrupt time",
      "type": "delta",
      "unit": "microsec"
    },
    {
      "name": "intr_cnt",
      "description": "Array of interrupt count per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "intr_cnt_ipi",
      "description": "IPI interrupt count per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "intr_cnt_msec",
      "description": "Millisecond interrupt count per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "intr_cnt_total",
      "description": "Total interrupt count per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "iscsi_data_received",
      "description": "iSCSI kilobytes (KB) received per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "iscsi_data_sent",
      "description": "iSCSI kilobytes (KB) sent per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "iscsi_ops",
      "description": "Number of iSCSI operations per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "memory",
      "description": "Total memory in megabytes (MB)",
      "type": "raw",
      "unit": "none"
    },
    {
      "name": "network_data_received",
      "description": "Number of network kilobytes (KB) received per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "network_data_sent",
      "description": "Number of network kilobytes (KB) sent per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "nfs_ops",
      "description": "Number of NFS operations per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "non_interrupt",
      "description": "Processor non-interrupt rate percentage",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cpu_elapsed_time"
      }
    },
    {
      "name": "non_interrupt_time",
      "description": "Processor non-interrupt time",
      "type": "delta",
      "unit": "microsec"
    },
    {
      "name": "num_processors",
      "description": "Number of active processors in the system",
      "type": "raw",
      "unit": "none"
    },
    {
      "name": "nvme_fc_data_received",
      "description": "NVMe/FC kilobytes (KB) received per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "nvme_fc_data_sent",
      "description": "NVMe/FC kilobytes (KB) sent per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "nvme_fc_ops",
      "description": "NVMe/FC operations per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "nvme_roce_data_received",
      "description": "NVMe/RoCE kilobytes (KB) received per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "nvme_roce_data_sent",
      "description": "NVMe/RoCE kilobytes (KB) sent per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "nvme_roce_ops",
      "description": "NVMe/RoCE operations per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "nvme_tcp_data_received",
      "description": "NVMe/TCP kilobytes (KB) received per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "nvme_tcp_data_sent",
      "description": "NVMe/TCP kilobytes (KB) sent per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "nvme_tcp_ops",
      "description": "NVMe/TCP operations per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "other_data",
      "description": "Other throughput",
      "type": "rate",
      "unit": "b_per_sec"
    },
    {
      "name": "other_latency",
      "description": "Average latency for all other operations in the system in microseconds",
      "type": "average",
      "unit": "microsec",
      "denominator": {
        "name": "other_ops"
      }
    },
    {
      "name": "other_ops",
      "description": "All other operations per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "partner_data_received",
      "description": "SCSI Partner kilobytes (KB) received per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "partner_data_sent",
      "description": "SCSI Partner kilobytes (KB) sent per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "processor_plevel",
      "description": "Processor plevel rate percentage",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cpu_elapsed_time"
      }
    },
    {
      "name": "processor_plevel_time",
      "description": "Processor plevel rate percentage",
      "type": "delta",
      "unit": "none"
    },
    {
      "name": "read_data",
      "description": "Read throughput",
      "type": "rate",
      "unit": "b_per_sec"
    },
    {
      "name": "read_latency",
      "description": "Average latency for all read operations in the system in microseconds",
      "type": "average",
      "unit": "microsec",
      "denominator": {
        "name": "read_ops"
      }
    },
    {
      "name": "read_ops",
      "description": "Read operations per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "sk_switches",
      "description": "Number of sk switches per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "ssd_data_read",
      "description": "Number of SSD Disk kilobytes (KB) read per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "ssd_data_written",
      "description": "Number of SSD Disk kilobytes (KB) written per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "sys_read_data",
      "description": "Network and FCP kilobytes (KB) received per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "sys_total_data",
      "description": "Network and FCP kilobytes (KB) received and sent per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "sys_write_data",
      "description": "Network and FCP kilobytes (KB) sent per second",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "tape_data_read",
      "description": "Tape bytes read per millisecond",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "tape_data_written",
      "description": "Tape bytes written per millisecond",
      "type": "rate",
      "unit": "kb_per_sec"
    },
    {
      "name": "time",
      "description": "Time in seconds since the Epoch (00:00:00 UTC January 1 1970)",
      "type": "raw",
      "unit": "sec"
    },
    {
      "name": "time_per_interrupt",
      "description": "Processor time per interrupt",
      "type": "average",
      "unit": "microsec",
      "denominator": {
        "name": "interrupt_num"
      }
    },
    {
      "name": "time_per_interrupt_in_cp",
      "description": "Processor time per interrupt in CP",
      "type": "average",
      "unit": "microsec",
      "denominator": {
        "name": "interrupt_num_in_cp"
      }
    },
    {
      "name": "total_data",
      "description": "Total throughput in bytes",
      "type": "rate",
      "unit": "b_per_sec"
    },
    {
      "name": "total_latency",
      "description": "Average latency for all operations in the system in microseconds",
      "type": "average",
      "unit": "microsec",
      "denominator": {
        "name": "total_ops"
      }
    },
    {
      "name": "total_ops",
      "description": "Total number of operations per second",
      "type": "rate",
      "unit": "per_sec"
    },
    {
      "name": "total_processor_busy",
      "description": "Total processor utilization of all processors in the system",
      "type": "percent",
      "unit": "percent",
      "denominator": {
        "name": "cpu_elapsed_time"
      }
    },
    {
      "name": "total_processor_busy_time",
      "description": "Total processor time of all processors in the system",
      "type": "delta",
      "unit": "microsec"
    },
    {
      "name": "uptime",
      "description": "Time in seconds that the system has been up",
      "type": "raw",
      "unit": "sec"
    },
    {
      "name": "wafliron",
      "description": "Wafliron counters",
      "type": "delta",
      "unit": "none"
    },
    {
      "name": "write_data",
      "description": "Write throughput",
      "type": "rate",
      "unit": "b_per_sec"
    },
    {
      "name": "write_latency",
      "description": "Average latency for all write operations in the system in microseconds",
      "type": "average",
      "unit": "microsec",
      "denominator": {
        "name": "write_ops"
      }
    },
    {
      "name": "write_ops",
      "description": "Write operations per second",
      "type": "rate",
      "unit": "per_sec"
    }
  ],
  "_links": {
    "self": {
      "href": "/api/cluster/counter/tables/system:node"
    }
  }
}
```

</p>
</details>

### Node performance metrics with all instances, properties, and counters

Ask ONTAP to return all instances of `system:node`. For each `system:node` include all of that node's properties and performance metrics.

```bash
curl 'https://10.193.48.154/api/cluster/counter/tables/system:node/rows?fields=*&return_records=true'
```

<details><summary>Response</summary>
<p>

```json
{
  "records": [
    {
      "counter_table": {
        "name": "system:node"
      },
      "id": "umeng-aff300-01:28e14eab-0580-11e8-bd9d-00a098d39e12",
      "properties": [
        {
          "name": "node.name",
          "value": "umeng-aff300-01"
        },
        {
          "name": "system_model",
          "value": "AFF-A300"
        },
        {
          "name": "ontap_version",
          "value": "NetApp Release R9.12.1xN_221108_1315: Tue Nov  8 15:32:25 EST 2022 "
        },
        {
          "name": "compile_flags",
          "value": "1"
        },
        {
          "name": "serial_no",
          "value": "721802000260"
        },
        {
          "name": "system_id",
          "value": "0537124012"
        },
        {
          "name": "hostname",
          "value": "umeng-aff300-01"
        },
        {
          "name": "name",
          "value": "umeng-aff300-01"
        },
        {
          "name": "uuid",
          "value": "28e14eab-0580-11e8-bd9d-00a098d39e12"
        }
      ],
      "counters": [
        {
          "name": "memory",
          "value": 88766
        },
        {
          "name": "nfs_ops",
          "value": 15991465
        },
        {
          "name": "cifs_ops",
          "value": 0
        },
        {
          "name": "fcp_ops",
          "value": 0
        },
        {
          "name": "iscsi_ops",
          "value": 355884195
        },
        {
          "name": "nvme_fc_ops",
          "value": 0
        },
        {
          "name": "nvme_tcp_ops",
          "value": 0
        },
        {
          "name": "nvme_roce_ops",
          "value": 0
        },
        {
          "name": "network_data_received",
          "value": 33454266379
        },
        {
          "name": "network_data_sent",
          "value": 9938586739
        },
        {
          "name": "fcp_data_received",
          "value": 0
        },
        {
          "name": "fcp_data_sent",
          "value": 0
        },
        {
          "name": "iscsi_data_received",
          "value": 4543696942
        },
        {
          "name": "iscsi_data_sent",
          "value": 3058795391
        },
        {
          "name": "nvme_fc_data_received",
          "value": 0
        },
        {
          "name": "nvme_fc_data_sent",
          "value": 0
        },
        {
          "name": "nvme_tcp_data_received",
          "value": 0
        },
        {
          "name": "nvme_tcp_data_sent",
          "value": 0
        },
        {
          "name": "nvme_roce_data_received",
          "value": 0
        },
        {
          "name": "nvme_roce_data_sent",
          "value": 0
        },
        {
          "name": "partner_data_received",
          "value": 0
        },
        {
          "name": "partner_data_sent",
          "value": 0
        },
        {
          "name": "sys_read_data",
          "value": 33454266379
        },
        {
          "name": "sys_write_data",
          "value": 9938586739
        },
        {
          "name": "sys_total_data",
          "value": 43392853118
        },
        {
          "name": "disk_data_read",
          "value": 32083838540
        },
        {
          "name": "disk_data_written",
          "value": 21102507352
        },
        {
          "name": "hdd_data_read",
          "value": 0
        },
        {
          "name": "hdd_data_written",
          "value": 0
        },
        {
          "name": "ssd_data_read",
          "value": 32083838540
        },
        {
          "name": "ssd_data_written",
          "value": 21102507352
        },
        {
          "name": "tape_data_read",
          "value": 0
        },
        {
          "name": "tape_data_written",
          "value": 0
        },
        {
          "name": "read_ops",
          "value": 33495530
        },
        {
          "name": "write_ops",
          "value": 324699398
        },
        {
          "name": "other_ops",
          "value": 13680732
        },
        {
          "name": "total_ops",
          "value": 371875660
        },
        {
          "name": "read_latency",
          "value": 14728140707
        },
        {
          "name": "write_latency",
          "value": 1568830328022
        },
        {
          "name": "other_latency",
          "value": 2132691612
        },
        {
          "name": "total_latency",
          "value": 1585691160341
        },
        {
          "name": "read_data",
          "value": 3212301497187
        },
        {
          "name": "write_data",
          "value": 4787509093524
        },
        {
          "name": "other_data",
          "value": 0
        },
        {
          "name": "total_data",
          "value": 7999810590711
        },
        {
          "name": "cpu_busy",
          "value": 790347800332
        },
        {
          "name": "cpu_elapsed_time",
          "value": 3979034040025
        },
        {
          "name": "average_processor_busy_percent",
          "value": 788429907770
        },
        {
          "name": "total_processor_busy",
          "value": 12614878524320
        },
        {
          "name": "total_processor_busy_time",
          "value": 12614878524320
        },
        {
          "name": "num_processors",
          "value": 16
        },
        {
          "name": "interrupt_time",
          "value": 118435504138
        },
        {
          "name": "interrupt",
          "value": 118435504138
        },
        {
          "name": "interrupt_num",
          "value": 1446537540
        },
        {
          "name": "time_per_interrupt",
          "value": 118435504138
        },
        {
          "name": "non_interrupt_time",
          "value": 12496443020182
        },
        {
          "name": "non_interrupt",
          "value": 12496443020182
        },
        {
          "name": "idle_time",
          "value": 51049666116080
        },
        {
          "name": "idle",
          "value": 51049666116080
        },
        {
          "name": "cp_time",
          "value": 221447740301
        },
        {
          "name": "cp",
          "value": 221447740301
        },
        {
          "name": "interrupt_in_cp_time",
          "value": 7969316828
        },
        {
          "name": "interrupt_in_cp",
          "value": 7969316828
        },
        {
          "name": "interrupt_num_in_cp",
          "value": 1639345044
        },
        {
          "name": "time_per_interrupt_in_cp",
          "value": 7969316828
        },
        {
          "name": "sk_switches",
          "value": 3830419593
        },
        {
          "name": "hard_switches",
          "value": 2786999477
        },
        {
          "name": "intr_cnt_msec",
          "value": 3978648113
        },
        {
          "name": "intr_cnt_ipi",
          "value": 1709054
        },
        {
          "name": "intr_cnt_total",
          "value": 1215253490
        },
        {
          "name": "time",
          "value": 1677516216
        },
        {
          "name": "uptime",
          "value": 3978648
        },
        {
          "name": "processor_plevel_time",
          "values": [
            3405835479577,
            2628275207938,
            1916273074545,
            1366761457118,
            964863281216,
            676002919489,
            472533086045,
            331487674159,
            234447654307,
            167247803300,
            120098535891,
            86312126550,
            61675398266,
            43549889374,
            30176461104,
            19891286233,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
          ],
          "labels": [
            "0_CPU",
            "1_CPU",
            "2_CPU",
            "3_CPU",
            "4_CPU",
            "5_CPU",
            "6_CPU",
            "7_CPU",
            "8_CPU",
            "9_CPU",
            "10_CPU",
            "11_CPU",
            "12_CPU",
            "13_CPU",
            "14_CPU",
            "15_CPU",
            "16_CPU",
            "17_CPU",
            "18_CPU",
            "19_CPU",
            "20_CPU",
            "21_CPU",
            "22_CPU",
            "23_CPU",
            "24_CPU",
            "25_CPU",
            "26_CPU",
            "27_CPU",
            "28_CPU",
            "29_CPU",
            "30_CPU",
            "31_CPU",
            "32_CPU",
            "33_CPU",
            "34_CPU",
            "35_CPU",
            "36_CPU",
            "37_CPU",
            "38_CPU",
            "39_CPU",
            "40_CPU",
            "41_CPU",
            "42_CPU",
            "43_CPU",
            "44_CPU",
            "45_CPU",
            "46_CPU",
            "47_CPU",
            "48_CPU",
            "49_CPU",
            "50_CPU",
            "51_CPU",
            "52_CPU",
            "53_CPU",
            "54_CPU",
            "55_CPU",
            "56_CPU",
            "57_CPU",
            "58_CPU",
            "59_CPU",
            "60_CPU",
            "61_CPU",
            "62_CPU",
            "63_CPU",
            "64_CPU",
            "65_CPU",
            "66_CPU",
            "67_CPU",
            "68_CPU",
            "69_CPU",
            "70_CPU",
            "71_CPU",
            "72_CPU",
            "73_CPU",
            "74_CPU",
            "75_CPU",
            "76_CPU",
            "77_CPU",
            "78_CPU",
            "79_CPU",
            "80_CPU",
            "81_CPU",
            "82_CPU",
            "83_CPU",
            "84_CPU",
            "85_CPU",
            "86_CPU",
            "87_CPU",
            "88_CPU",
            "89_CPU",
            "90_CPU",
            "91_CPU",
            "92_CPU",
            "93_CPU",
            "94_CPU",
            "95_CPU",
            "96_CPU",
            "97_CPU",
            "98_CPU",
            "99_CPU",
            "100_CPU",
            "101_CPU",
            "102_CPU",
            "103_CPU",
            "104_CPU",
            "105_CPU",
            "106_CPU",
            "107_CPU",
            "108_CPU",
            "109_CPU",
            "110_CPU",
            "111_CPU",
            "112_CPU",
            "113_CPU",
            "114_CPU",
            "115_CPU",
            "116_CPU",
            "117_CPU",
            "118_CPU",
            "119_CPU",
            "120_CPU",
            "121_CPU",
            "122_CPU",
            "123_CPU",
            "124_CPU",
            "125_CPU",
            "126_CPU",
            "127_CPU"
          ]
        },
        {
          "name": "processor_plevel",
          "values": [
            3405835479577,
            2628275207938,
            1916273074545,
            1366761457118,
            964863281216,
            676002919489,
            472533086045,
            331487674159,
            234447654307,
            167247803300,
            120098535891,
            86312126550,
            61675398266,
            43549889374,
            30176461104,
            19891286233,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
          ],
          "labels": [
            "0_CPU",
            "1_CPU",
            "2_CPU",
            "3_CPU",
            "4_CPU",
            "5_CPU",
            "6_CPU",
            "7_CPU",
            "8_CPU",
            "9_CPU",
            "10_CPU",
            "11_CPU",
            "12_CPU",
            "13_CPU",
            "14_CPU",
            "15_CPU",
            "16_CPU",
            "17_CPU",
            "18_CPU",
            "19_CPU",
            "20_CPU",
            "21_CPU",
            "22_CPU",
            "23_CPU",
            "24_CPU",
            "25_CPU",
            "26_CPU",
            "27_CPU",
            "28_CPU",
            "29_CPU",
            "30_CPU",
            "31_CPU",
            "32_CPU",
            "33_CPU",
            "34_CPU",
            "35_CPU",
            "36_CPU",
            "37_CPU",
            "38_CPU",
            "39_CPU",
            "40_CPU",
            "41_CPU",
            "42_CPU",
            "43_CPU",
            "44_CPU",
            "45_CPU",
            "46_CPU",
            "47_CPU",
            "48_CPU",
            "49_CPU",
            "50_CPU",
            "51_CPU",
            "52_CPU",
            "53_CPU",
            "54_CPU",
            "55_CPU",
            "56_CPU",
            "57_CPU",
            "58_CPU",
            "59_CPU",
            "60_CPU",
            "61_CPU",
            "62_CPU",
            "63_CPU",
            "64_CPU",
            "65_CPU",
            "66_CPU",
            "67_CPU",
            "68_CPU",
            "69_CPU",
            "70_CPU",
            "71_CPU",
            "72_CPU",
            "73_CPU",
            "74_CPU",
            "75_CPU",
            "76_CPU",
            "77_CPU",
            "78_CPU",
            "79_CPU",
            "80_CPU",
            "81_CPU",
            "82_CPU",
            "83_CPU",
            "84_CPU",
            "85_CPU",
            "86_CPU",
            "87_CPU",
            "88_CPU",
            "89_CPU",
            "90_CPU",
            "91_CPU",
            "92_CPU",
            "93_CPU",
            "94_CPU",
            "95_CPU",
            "96_CPU",
            "97_CPU",
            "98_CPU",
            "99_CPU",
            "100_CPU",
            "101_CPU",
            "102_CPU",
            "103_CPU",
            "104_CPU",
            "105_CPU",
            "106_CPU",
            "107_CPU",
            "108_CPU",
            "109_CPU",
            "110_CPU",
            "111_CPU",
            "112_CPU",
            "113_CPU",
            "114_CPU",
            "115_CPU",
            "116_CPU",
            "117_CPU",
            "118_CPU",
            "119_CPU",
            "120_CPU",
            "121_CPU",
            "122_CPU",
            "123_CPU",
            "124_CPU",
            "125_CPU",
            "126_CPU",
            "127_CPU"
          ]
        },
        {
          "name": "domain_busy",
          "values": [
            51049666116086,
            13419960088,
            13297686377,
            1735383373870,
            39183250298,
            6728050897,
            28229793795,
            17493622207,
            122290467,
            974721172619,
            47944793823,
            164946850,
            4162377932,
            407009733276,
            128199854099,
            9037374471285,
            38911301970,
            366749865,
            732045734,
            2997541695,
            14,
            18,
            40
          ],
          "labels": [
            "idle",
            "kahuna",
            "storage",
            "exempt",
            "none",
            "raid",
            "raid_exempt",
            "xor_exempt",
            "target",
            "wafl_exempt",
            "wafl_mpcleaner",
            "sm_exempt",
            "protocol",
            "nwk_exempt",
            "network",
            "hostOS",
            "ssan_exempt",
            "unclassified",
            "kahuna_legacy",
            "ha",
            "ssan_exempt2",
            "exempt_ise",
            "zombie"
          ]
        },
        {
          "name": "domain_shared",
          "values": [
            0,
            685164024474,
            0,
            0,
            0,
            24684879894,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
          ],
          "labels": [
            "idle",
            "kahuna",
            "storage",
            "exempt",
            "none",
            "raid",
            "raid_exempt",
            "xor_exempt",
            "target",
            "wafl_exempt",
            "wafl_mpcleaner",
            "sm_exempt",
            "protocol",
            "nwk_exempt",
            "network",
            "hostOS",
            "ssan_exempt",
            "unclassified",
            "kahuna_legacy",
            "ha",
            "ssan_exempt2",
            "exempt_ise",
            "zombie"
          ]
        },
        {
          "name": "dswitchto_cnt",
          "values": [
            0,
            322698663,
            172936437,
            446893016,
            96971,
            39788918,
            5,
            10,
            10670440,
            22,
            7,
            836,
            2407967,
            9798186907,
            9802868991,
            265242,
            53,
            2614118,
            4430780,
            66117706,
            1,
            1,
            1
          ],
          "labels": [
            "idle",
            "kahuna",
            "storage",
            "exempt",
            "none",
            "raid",
            "raid_exempt",
            "xor_exempt",
            "target",
            "wafl_exempt",
            "wafl_mpcleaner",
            "sm_exempt",
            "protocol",
            "nwk_exempt",
            "network",
            "hostOS",
            "ssan_exempt",
            "unclassified",
            "kahuna_legacy",
            "ha",
            "ssan_exempt2",
            "exempt_ise",
            "zombie"
          ]
        },
        {
          "name": "intr_cnt",
          "values": [
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            4191453008,
            8181232,
            1625052957,
            0,
            71854,
            0,
            71854,
            0,
            5,
            0,
            5,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
          ],
          "labels": [
            "dev_0",
            "dev_1",
            "dev_2",
            "dev_3",
            "dev_4",
            "dev_5",
            "dev_6",
            "dev_7",
            "dev_8",
            "dev_9",
            "dev_10",
            "dev_11",
            "dev_12",
            "dev_13",
            "dev_14",
            "dev_15",
            "dev_16",
            "dev_17",
            "dev_18",
            "dev_19",
            "dev_20",
            "dev_21",
            "dev_22",
            "dev_23",
            "dev_24",
            "dev_25",
            "dev_26",
            "dev_27",
            "dev_28",
            "dev_29",
            "dev_30",
            "dev_31",
            "dev_32",
            "dev_33",
            "dev_34",
            "dev_35",
            "dev_36",
            "dev_37",
            "dev_38",
            "dev_39",
            "dev_40",
            "dev_41",
            "dev_42",
            "dev_43",
            "dev_44",
            "dev_45",
            "dev_46",
            "dev_47",
            "dev_48",
            "dev_49",
            "dev_50",
            "dev_51",
            "dev_52",
            "dev_53",
            "dev_54",
            "dev_55",
            "dev_56",
            "dev_57",
            "dev_58",
            "dev_59",
            "dev_60",
            "dev_61",
            "dev_62",
            "dev_63",
            "dev_64",
            "dev_65",
            "dev_66",
            "dev_67",
            "dev_68",
            "dev_69",
            "dev_70",
            "dev_71",
            "dev_72",
            "dev_73",
            "dev_74",
            "dev_75",
            "dev_76",
            "dev_77",
            "dev_78",
            "dev_79",
            "dev_80",
            "dev_81",
            "dev_82",
            "dev_83",
            "dev_84",
            "dev_85",
            "dev_86",
            "dev_87",
            "dev_88",
            "dev_89",
            "dev_90",
            "dev_91",
            "dev_92",
            "dev_93",
            "dev_94",
            "dev_95",
            "dev_96",
            "dev_97",
            "dev_98",
            "dev_99",
            "dev_100",
            "dev_101",
            "dev_102",
            "dev_103",
            "dev_104",
            "dev_105",
            "dev_106",
            "dev_107",
            "dev_108",
            "dev_109",
            "dev_110",
            "dev_111",
            "dev_112",
            "dev_113",
            "dev_114",
            "dev_115",
            "dev_116",
            "dev_117",
            "dev_118",
            "dev_119",
            "dev_120",
            "dev_121",
            "dev_122",
            "dev_123",
            "dev_124",
            "dev_125",
            "dev_126",
            "dev_127",
            "dev_128",
            "dev_129",
            "dev_130",
            "dev_131",
            "dev_132",
            "dev_133",
            "dev_134",
            "dev_135",
            "dev_136",
            "dev_137",
            "dev_138",
            "dev_139",
            "dev_140",
            "dev_141",
            "dev_142",
            "dev_143",
            "dev_144",
            "dev_145",
            "dev_146",
            "dev_147",
            "dev_148",
            "dev_149",
            "dev_150",
            "dev_151",
            "dev_152",
            "dev_153",
            "dev_154",
            "dev_155",
            "dev_156",
            "dev_157",
            "dev_158",
            "dev_159",
            "dev_160",
            "dev_161",
            "dev_162",
            "dev_163",
            "dev_164",
            "dev_165",
            "dev_166",
            "dev_167",
            "dev_168",
            "dev_169",
            "dev_170",
            "dev_171",
            "dev_172",
            "dev_173",
            "dev_174",
            "dev_175",
            "dev_176",
            "dev_177",
            "dev_178",
            "dev_179",
            "dev_180",
            "dev_181",
            "dev_182",
            "dev_183",
            "dev_184",
            "dev_185",
            "dev_186",
            "dev_187",
            "dev_188",
            "dev_189",
            "dev_190",
            "dev_191",
            "dev_192",
            "dev_193",
            "dev_194",
            "dev_195",
            "dev_196",
            "dev_197",
            "dev_198",
            "dev_199",
            "dev_200",
            "dev_201",
            "dev_202",
            "dev_203",
            "dev_204",
            "dev_205",
            "dev_206",
            "dev_207",
            "dev_208",
            "dev_209",
            "dev_210",
            "dev_211",
            "dev_212",
            "dev_213",
            "dev_214",
            "dev_215",
            "dev_216",
            "dev_217",
            "dev_218",
            "dev_219",
            "dev_220",
            "dev_221",
            "dev_222",
            "dev_223",
            "dev_224",
            "dev_225",
            "dev_226",
            "dev_227",
            "dev_228",
            "dev_229",
            "dev_230",
            "dev_231",
            "dev_232",
            "dev_233",
            "dev_234",
            "dev_235",
            "dev_236",
            "dev_237",
            "dev_238",
            "dev_239",
            "dev_240",
            "dev_241",
            "dev_242",
            "dev_243",
            "dev_244",
            "dev_245",
            "dev_246",
            "dev_247",
            "dev_248",
            "dev_249",
            "dev_250",
            "dev_251",
            "dev_252",
            "dev_253",
            "dev_254",
            "dev_255"
          ]
        },
        {
          "name": "wafliron",
          "values": [
            0,
            0,
            0
          ],
          "labels": [
            "iron_totstarts",
            "iron_nobackup",
            "iron_usebackup"
          ]
        }
      ],
      "aggregation": {
        "count": 2,
        "complete": true
      },
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/system:node/rows/umeng-aff300-01%3A28e14eab-0580-11e8-bd9d-00a098d39e12"
        }
      }
    },
    {
      "counter_table": {
        "name": "system:node"
      },
      "id": "umeng-aff300-02:1524afca-0580-11e8-ae74-00a098d390f2",
      "properties": [
        {
          "name": "node.name",
          "value": "umeng-aff300-02"
        },
        {
          "name": "system_model",
          "value": "AFF-A300"
        },
        {
          "name": "ontap_version",
          "value": "NetApp Release R9.12.1xN_221108_1315: Tue Nov  8 15:32:25 EST 2022 "
        },
        {
          "name": "compile_flags",
          "value": "1"
        },
        {
          "name": "serial_no",
          "value": "721802000259"
        },
        {
          "name": "system_id",
          "value": "0537123843"
        },
        {
          "name": "hostname",
          "value": "umeng-aff300-02"
        },
        {
          "name": "name",
          "value": "umeng-aff300-02"
        },
        {
          "name": "uuid",
          "value": "1524afca-0580-11e8-ae74-00a098d390f2"
        }
      ],
      "counters": [
        {
          "name": "memory",
          "value": 88766
        },
        {
          "name": "nfs_ops",
          "value": 2061227971
        },
        {
          "name": "cifs_ops",
          "value": 0
        },
        {
          "name": "fcp_ops",
          "value": 0
        },
        {
          "name": "iscsi_ops",
          "value": 183570559
        },
        {
          "name": "nvme_fc_ops",
          "value": 0
        },
        {
          "name": "nvme_tcp_ops",
          "value": 0
        },
        {
          "name": "nvme_roce_ops",
          "value": 0
        },
        {
          "name": "network_data_received",
          "value": 28707362447
        },
        {
          "name": "network_data_sent",
          "value": 31199786274
        },
        {
          "name": "fcp_data_received",
          "value": 0
        },
        {
          "name": "fcp_data_sent",
          "value": 0
        },
        {
          "name": "iscsi_data_received",
          "value": 2462501728
        },
        {
          "name": "iscsi_data_sent",
          "value": 962425592
        },
        {
          "name": "nvme_fc_data_received",
          "value": 0
        },
        {
          "name": "nvme_fc_data_sent",
          "value": 0
        },
        {
          "name": "nvme_tcp_data_received",
          "value": 0
        },
        {
          "name": "nvme_tcp_data_sent",
          "value": 0
        },
        {
          "name": "nvme_roce_data_received",
          "value": 0
        },
        {
          "name": "nvme_roce_data_sent",
          "value": 0
        },
        {
          "name": "partner_data_received",
          "value": 0
        },
        {
          "name": "partner_data_sent",
          "value": 0
        },
        {
          "name": "sys_read_data",
          "value": 28707362447
        },
        {
          "name": "sys_write_data",
          "value": 31199786274
        },
        {
          "name": "sys_total_data",
          "value": 59907148721
        },
        {
          "name": "disk_data_read",
          "value": 27355740700
        },
        {
          "name": "disk_data_written",
          "value": 3426898232
        },
        {
          "name": "hdd_data_read",
          "value": 0
        },
        {
          "name": "hdd_data_written",
          "value": 0
        },
        {
          "name": "ssd_data_read",
          "value": 27355740700
        },
        {
          "name": "ssd_data_written",
          "value": 3426898232
        },
        {
          "name": "tape_data_read",
          "value": 0
        },
        {
          "name": "tape_data_written",
          "value": 0
        },
        {
          "name": "read_ops",
          "value": 29957410
        },
        {
          "name": "write_ops",
          "value": 2141657620
        },
        {
          "name": "other_ops",
          "value": 73183500
        },
        {
          "name": "total_ops",
          "value": 2244798530
        },
        {
          "name": "read_latency",
          "value": 43283636161
        },
        {
          "name": "write_latency",
          "value": 1437635703835
        },
        {
          "name": "other_latency",
          "value": 628457365
        },
        {
          "name": "total_latency",
          "value": 1481547797361
        },
        {
          "name": "read_data",
          "value": 1908711454978
        },
        {
          "name": "write_data",
          "value": 23562759645410
        },
        {
          "name": "other_data",
          "value": 0
        },
        {
          "name": "total_data",
          "value": 25471471100388
        },
        {
          "name": "cpu_busy",
          "value": 511050841704
        },
        {
          "name": "cpu_elapsed_time",
          "value": 3979039364919
        },
        {
          "name": "average_processor_busy_percent",
          "value": 509151403977
        },
        {
          "name": "total_processor_busy",
          "value": 8146422463632
        },
        {
          "name": "total_processor_busy_time",
          "value": 8146422463632
        },
        {
          "name": "num_processors",
          "value": 16
        },
        {
          "name": "interrupt_time",
          "value": 108155323601
        },
        {
          "name": "interrupt",
          "value": 108155323601
        },
        {
          "name": "interrupt_num",
          "value": 3369179127
        },
        {
          "name": "time_per_interrupt",
          "value": 108155323601
        },
        {
          "name": "non_interrupt_time",
          "value": 8038267140031
        },
        {
          "name": "non_interrupt",
          "value": 8038267140031
        },
        {
          "name": "idle_time",
          "value": 55518207375072
        },
        {
          "name": "idle",
          "value": 55518207375072
        },
        {
          "name": "cp_time",
          "value": 64306316680
        },
        {
          "name": "cp",
          "value": 64306316680
        },
        {
          "name": "interrupt_in_cp_time",
          "value": 2024956616
        },
        {
          "name": "interrupt_in_cp",
          "value": 2024956616
        },
        {
          "name": "interrupt_num_in_cp",
          "value": 2661183541
        },
        {
          "name": "time_per_interrupt_in_cp",
          "value": 2024956616
        },
        {
          "name": "sk_switches",
          "value": 2798598514
        },
        {
          "name": "hard_switches",
          "value": 1354185066
        },
        {
          "name": "intr_cnt_msec",
          "value": 3978642246
        },
        {
          "name": "intr_cnt_ipi",
          "value": 797281
        },
        {
          "name": "intr_cnt_total",
          "value": 905575861
        },
        {
          "name": "time",
          "value": 1677516216
        },
        {
          "name": "uptime",
          "value": 3978643
        },
        {
          "name": "processor_plevel_time",
          "values": [
            2878770221447,
            1882901052733,
            1209134416474,
            771086627192,
            486829133301,
            306387520688,
            193706139760,
            123419519944,
            79080346535,
            50459518003,
            31714732122,
            19476561954,
            11616026278,
            6666253598,
            3623880168,
            1790458071,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
          ],
          "labels": [
            "0_CPU",
            "1_CPU",
            "2_CPU",
            "3_CPU",
            "4_CPU",
            "5_CPU",
            "6_CPU",
            "7_CPU",
            "8_CPU",
            "9_CPU",
            "10_CPU",
            "11_CPU",
            "12_CPU",
            "13_CPU",
            "14_CPU",
            "15_CPU",
            "16_CPU",
            "17_CPU",
            "18_CPU",
            "19_CPU",
            "20_CPU",
            "21_CPU",
            "22_CPU",
            "23_CPU",
            "24_CPU",
            "25_CPU",
            "26_CPU",
            "27_CPU",
            "28_CPU",
            "29_CPU",
            "30_CPU",
            "31_CPU",
            "32_CPU",
            "33_CPU",
            "34_CPU",
            "35_CPU",
            "36_CPU",
            "37_CPU",
            "38_CPU",
            "39_CPU",
            "40_CPU",
            "41_CPU",
            "42_CPU",
            "43_CPU",
            "44_CPU",
            "45_CPU",
            "46_CPU",
            "47_CPU",
            "48_CPU",
            "49_CPU",
            "50_CPU",
            "51_CPU",
            "52_CPU",
            "53_CPU",
            "54_CPU",
            "55_CPU",
            "56_CPU",
            "57_CPU",
            "58_CPU",
            "59_CPU",
            "60_CPU",
            "61_CPU",
            "62_CPU",
            "63_CPU",
            "64_CPU",
            "65_CPU",
            "66_CPU",
            "67_CPU",
            "68_CPU",
            "69_CPU",
            "70_CPU",
            "71_CPU",
            "72_CPU",
            "73_CPU",
            "74_CPU",
            "75_CPU",
            "76_CPU",
            "77_CPU",
            "78_CPU",
            "79_CPU",
            "80_CPU",
            "81_CPU",
            "82_CPU",
            "83_CPU",
            "84_CPU",
            "85_CPU",
            "86_CPU",
            "87_CPU",
            "88_CPU",
            "89_CPU",
            "90_CPU",
            "91_CPU",
            "92_CPU",
            "93_CPU",
            "94_CPU",
            "95_CPU",
            "96_CPU",
            "97_CPU",
            "98_CPU",
            "99_CPU",
            "100_CPU",
            "101_CPU",
            "102_CPU",
            "103_CPU",
            "104_CPU",
            "105_CPU",
            "106_CPU",
            "107_CPU",
            "108_CPU",
            "109_CPU",
            "110_CPU",
            "111_CPU",
            "112_CPU",
            "113_CPU",
            "114_CPU",
            "115_CPU",
            "116_CPU",
            "117_CPU",
            "118_CPU",
            "119_CPU",
            "120_CPU",
            "121_CPU",
            "122_CPU",
            "123_CPU",
            "124_CPU",
            "125_CPU",
            "126_CPU",
            "127_CPU"
          ]
        },
        {
          "name": "processor_plevel",
          "values": [
            2878770221447,
            1882901052733,
            1209134416474,
            771086627192,
            486829133301,
            306387520688,
            193706139760,
            123419519944,
            79080346535,
            50459518003,
            31714732122,
            19476561954,
            11616026278,
            6666253598,
            3623880168,
            1790458071,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
          ],
          "labels": [
            "0_CPU",
            "1_CPU",
            "2_CPU",
            "3_CPU",
            "4_CPU",
            "5_CPU",
            "6_CPU",
            "7_CPU",
            "8_CPU",
            "9_CPU",
            "10_CPU",
            "11_CPU",
            "12_CPU",
            "13_CPU",
            "14_CPU",
            "15_CPU",
            "16_CPU",
            "17_CPU",
            "18_CPU",
            "19_CPU",
            "20_CPU",
            "21_CPU",
            "22_CPU",
            "23_CPU",
            "24_CPU",
            "25_CPU",
            "26_CPU",
            "27_CPU",
            "28_CPU",
            "29_CPU",
            "30_CPU",
            "31_CPU",
            "32_CPU",
            "33_CPU",
            "34_CPU",
            "35_CPU",
            "36_CPU",
            "37_CPU",
            "38_CPU",
            "39_CPU",
            "40_CPU",
            "41_CPU",
            "42_CPU",
            "43_CPU",
            "44_CPU",
            "45_CPU",
            "46_CPU",
            "47_CPU",
            "48_CPU",
            "49_CPU",
            "50_CPU",
            "51_CPU",
            "52_CPU",
            "53_CPU",
            "54_CPU",
            "55_CPU",
            "56_CPU",
            "57_CPU",
            "58_CPU",
            "59_CPU",
            "60_CPU",
            "61_CPU",
            "62_CPU",
            "63_CPU",
            "64_CPU",
            "65_CPU",
            "66_CPU",
            "67_CPU",
            "68_CPU",
            "69_CPU",
            "70_CPU",
            "71_CPU",
            "72_CPU",
            "73_CPU",
            "74_CPU",
            "75_CPU",
            "76_CPU",
            "77_CPU",
            "78_CPU",
            "79_CPU",
            "80_CPU",
            "81_CPU",
            "82_CPU",
            "83_CPU",
            "84_CPU",
            "85_CPU",
            "86_CPU",
            "87_CPU",
            "88_CPU",
            "89_CPU",
            "90_CPU",
            "91_CPU",
            "92_CPU",
            "93_CPU",
            "94_CPU",
            "95_CPU",
            "96_CPU",
            "97_CPU",
            "98_CPU",
            "99_CPU",
            "100_CPU",
            "101_CPU",
            "102_CPU",
            "103_CPU",
            "104_CPU",
            "105_CPU",
            "106_CPU",
            "107_CPU",
            "108_CPU",
            "109_CPU",
            "110_CPU",
            "111_CPU",
            "112_CPU",
            "113_CPU",
            "114_CPU",
            "115_CPU",
            "116_CPU",
            "117_CPU",
            "118_CPU",
            "119_CPU",
            "120_CPU",
            "121_CPU",
            "122_CPU",
            "123_CPU",
            "124_CPU",
            "125_CPU",
            "126_CPU",
            "127_CPU"
          ]
        },
        {
          "name": "domain_busy",
          "values": [
            55518207375080,
            8102895398,
            12058227646,
            991838747162,
            28174147737,
            6669066926,
            14245801778,
            9009875224,
            118982762,
            177496844302,
            5888814259,
            167280195,
            3851617905,
            484154906167,
            91240285306,
            6180138216837,
            22111798640,
            344700584,
            266304074,
            2388625825,
            16,
            21,
            19
          ],
          "labels": [
            "idle",
            "kahuna",
            "storage",
            "exempt",
            "none",
            "raid",
            "raid_exempt",
            "xor_exempt",
            "target",
            "wafl_exempt",
            "wafl_mpcleaner",
            "sm_exempt",
            "protocol",
            "nwk_exempt",
            "network",
            "hostOS",
            "ssan_exempt",
            "unclassified",
            "kahuna_legacy",
            "ha",
            "ssan_exempt2",
            "exempt_ise",
            "zombie"
          ]
        },
        {
          "name": "domain_shared",
          "values": [
            0,
            153663450171,
            0,
            0,
            0,
            11834112384,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
          ],
          "labels": [
            "idle",
            "kahuna",
            "storage",
            "exempt",
            "none",
            "raid",
            "raid_exempt",
            "xor_exempt",
            "target",
            "wafl_exempt",
            "wafl_mpcleaner",
            "sm_exempt",
            "protocol",
            "nwk_exempt",
            "network",
            "hostOS",
            "ssan_exempt",
            "unclassified",
            "kahuna_legacy",
            "ha",
            "ssan_exempt2",
            "exempt_ise",
            "zombie"
          ]
        },
        {
          "name": "dswitchto_cnt",
          "values": [
            0,
            178192633,
            143964155,
            286324250,
            2365,
            39684121,
            5,
            10,
            10715325,
            22,
            7,
            30,
            2407970,
            7865489299,
            7870331008,
            265242,
            53,
            2535145,
            3252888,
            53334340,
            1,
            1,
            1
          ],
          "labels": [
            "idle",
            "kahuna",
            "storage",
            "exempt",
            "none",
            "raid",
            "raid_exempt",
            "xor_exempt",
            "target",
            "wafl_exempt",
            "wafl_mpcleaner",
            "sm_exempt",
            "protocol",
            "nwk_exempt",
            "network",
            "hostOS",
            "ssan_exempt",
            "unclassified",
            "kahuna_legacy",
            "ha",
            "ssan_exempt2",
            "exempt_ise",
            "zombie"
          ]
        },
        {
          "name": "intr_cnt",
          "values": [
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            724698481,
            8181275,
            488080162,
            0,
            71856,
            0,
            71856,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
          ],
          "labels": [
            "dev_0",
            "dev_1",
            "dev_2",
            "dev_3",
            "dev_4",
            "dev_5",
            "dev_6",
            "dev_7",
            "dev_8",
            "dev_9",
            "dev_10",
            "dev_11",
            "dev_12",
            "dev_13",
            "dev_14",
            "dev_15",
            "dev_16",
            "dev_17",
            "dev_18",
            "dev_19",
            "dev_20",
            "dev_21",
            "dev_22",
            "dev_23",
            "dev_24",
            "dev_25",
            "dev_26",
            "dev_27",
            "dev_28",
            "dev_29",
            "dev_30",
            "dev_31",
            "dev_32",
            "dev_33",
            "dev_34",
            "dev_35",
            "dev_36",
            "dev_37",
            "dev_38",
            "dev_39",
            "dev_40",
            "dev_41",
            "dev_42",
            "dev_43",
            "dev_44",
            "dev_45",
            "dev_46",
            "dev_47",
            "dev_48",
            "dev_49",
            "dev_50",
            "dev_51",
            "dev_52",
            "dev_53",
            "dev_54",
            "dev_55",
            "dev_56",
            "dev_57",
            "dev_58",
            "dev_59",
            "dev_60",
            "dev_61",
            "dev_62",
            "dev_63",
            "dev_64",
            "dev_65",
            "dev_66",
            "dev_67",
            "dev_68",
            "dev_69",
            "dev_70",
            "dev_71",
            "dev_72",
            "dev_73",
            "dev_74",
            "dev_75",
            "dev_76",
            "dev_77",
            "dev_78",
            "dev_79",
            "dev_80",
            "dev_81",
            "dev_82",
            "dev_83",
            "dev_84",
            "dev_85",
            "dev_86",
            "dev_87",
            "dev_88",
            "dev_89",
            "dev_90",
            "dev_91",
            "dev_92",
            "dev_93",
            "dev_94",
            "dev_95",
            "dev_96",
            "dev_97",
            "dev_98",
            "dev_99",
            "dev_100",
            "dev_101",
            "dev_102",
            "dev_103",
            "dev_104",
            "dev_105",
            "dev_106",
            "dev_107",
            "dev_108",
            "dev_109",
            "dev_110",
            "dev_111",
            "dev_112",
            "dev_113",
            "dev_114",
            "dev_115",
            "dev_116",
            "dev_117",
            "dev_118",
            "dev_119",
            "dev_120",
            "dev_121",
            "dev_122",
            "dev_123",
            "dev_124",
            "dev_125",
            "dev_126",
            "dev_127",
            "dev_128",
            "dev_129",
            "dev_130",
            "dev_131",
            "dev_132",
            "dev_133",
            "dev_134",
            "dev_135",
            "dev_136",
            "dev_137",
            "dev_138",
            "dev_139",
            "dev_140",
            "dev_141",
            "dev_142",
            "dev_143",
            "dev_144",
            "dev_145",
            "dev_146",
            "dev_147",
            "dev_148",
            "dev_149",
            "dev_150",
            "dev_151",
            "dev_152",
            "dev_153",
            "dev_154",
            "dev_155",
            "dev_156",
            "dev_157",
            "dev_158",
            "dev_159",
            "dev_160",
            "dev_161",
            "dev_162",
            "dev_163",
            "dev_164",
            "dev_165",
            "dev_166",
            "dev_167",
            "dev_168",
            "dev_169",
            "dev_170",
            "dev_171",
            "dev_172",
            "dev_173",
            "dev_174",
            "dev_175",
            "dev_176",
            "dev_177",
            "dev_178",
            "dev_179",
            "dev_180",
            "dev_181",
            "dev_182",
            "dev_183",
            "dev_184",
            "dev_185",
            "dev_186",
            "dev_187",
            "dev_188",
            "dev_189",
            "dev_190",
            "dev_191",
            "dev_192",
            "dev_193",
            "dev_194",
            "dev_195",
            "dev_196",
            "dev_197",
            "dev_198",
            "dev_199",
            "dev_200",
            "dev_201",
            "dev_202",
            "dev_203",
            "dev_204",
            "dev_205",
            "dev_206",
            "dev_207",
            "dev_208",
            "dev_209",
            "dev_210",
            "dev_211",
            "dev_212",
            "dev_213",
            "dev_214",
            "dev_215",
            "dev_216",
            "dev_217",
            "dev_218",
            "dev_219",
            "dev_220",
            "dev_221",
            "dev_222",
            "dev_223",
            "dev_224",
            "dev_225",
            "dev_226",
            "dev_227",
            "dev_228",
            "dev_229",
            "dev_230",
            "dev_231",
            "dev_232",
            "dev_233",
            "dev_234",
            "dev_235",
            "dev_236",
            "dev_237",
            "dev_238",
            "dev_239",
            "dev_240",
            "dev_241",
            "dev_242",
            "dev_243",
            "dev_244",
            "dev_245",
            "dev_246",
            "dev_247",
            "dev_248",
            "dev_249",
            "dev_250",
            "dev_251",
            "dev_252",
            "dev_253",
            "dev_254",
            "dev_255"
          ]
        },
        {
          "name": "wafliron",
          "values": [
            0,
            0,
            0
          ],
          "labels": [
            "iron_totstarts",
            "iron_nobackup",
            "iron_usebackup"
          ]
        }
      ],
      "aggregation": {
        "count": 2,
        "complete": true
      },
      "_links": {
        "self": {
          "href": "/api/cluster/counter/tables/system:node/rows/umeng-aff300-02%3A1524afca-0580-11e8-ae74-00a098d390f2"
        }
      }
    }
  ],
  "num_records": 2,
  "_links": {
    "self": {
      "href": "/api/cluster/counter/tables/system:node/rows?fields=*&return_records=true"
    }
  }
}
```

</p>
</details>

# References

- [Harvest REST Strategy](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-strategy.md)
- [ONTAP 9.11.1 ONTAPI-to-REST Counter Manager Mapping](https://library.netapp.com/ecm/ecm_download_file/ECMLP2883449)
- [ONTAP REST API reference documentation](https://docs.netapp.com/us-en/ontap-automation/reference/api_reference.html#access-the-ontap-api-documentation-page)
- [ONTAP REST API ](https://devnet.netapp.com/restapi.php)
