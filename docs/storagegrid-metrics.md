This document describes which StorageGRID metrics are collected and what those metrics are named in Harvest, including:

- Details about which Harvest metrics each dashboard uses.
These can be generated on demand by running `bin/harvest grafana metrics`. See
[#1577](https://github.com/NetApp/harvest/issues/1577#issue-1471478260) for details.

```
Creation Date : 2026-Feb-20
StorageGrid Version: 11.6.0
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

storagegrid_tenant_usage_data_bytes <span class="key">Name of the metric exported by Harvest</span>

The logical size of all objects for the tenant. <span class="key">Description of the StorageGrid metric</span>

* <span class="key">API</span> will be REST depending on which protocol is used to collect the metric
* <span class="key">Endpoint</span> name of the REST api used to collect this metric
* <span class="key">Metric</span> name of the StorageGrid metric
* <span class="key">Template</span> path of the template that collects the metric

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
|REST | `api/grid/accounts-cache` | dataBytes | conf/storagegrid/11.6.0/tenant.yaml|


## Metrics


### storagegrid_content_buckets_and_containers

Total number of S3 buckets and Swift containers


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_content_buckets_and_containers` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_content_buckets_and_containers` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: S3 | Content | timeseries | [Top $TopResources Nodes by S3 Buckets and Swift Containers](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=10) |
///



### storagegrid_content_objects

Total number of S3 and Swift objects (excluding empty objects)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_content_objects` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_content_objects` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: S3 | Content | timeseries | [Top $TopResources Nodes by S3 and Swift Objects (excluding empty objects)](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=9) |
///



### storagegrid_ilm_awaiting_client_objects

Total number of objects on this node awaiting ILM evaluation because of client operation (for example, ingest)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_ilm_awaiting_client_objects` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_ilm_awaiting_client_objects` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Information Lifecycle Management (ILM) | timeseries | [ILM queue (Objects)](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=16) |
| StorageGrid: S3 | ILM | timeseries | [Top $TopResources Nodes by ILM Awaiting Object Evaluation (incoming from clients)](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=12) |
///



### storagegrid_ilm_awaiting_total_objects

Total number of objects on this node awaiting ILM evaluation


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_ilm_awaiting_total_objects` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_ilm_awaiting_total_objects` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: S3 | ILM | timeseries | [Top $TopResources Nodes by ILM Awaiting Object Evaluation](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=13) |
///



### storagegrid_ilm_objects_processed

Objects processed by ILM that actually had work done


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_ilm_objects_processed` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_ilm_objects_processed` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Information Lifecycle Management (ILM) | timeseries | [ILM evaluation rate (objects/second)](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=18) |
///



### storagegrid_ilm_scan_objects_per_second

ILM scan rate (objects per second)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_ilm_scan_objects_per_second` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_ilm_scan_objects_per_second` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: S3 | ILM | timeseries | [Top $TopResources Nodes by ILM Scan Rate](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=11) |
///



### storagegrid_metadata_queries_average_latency_milliseconds

Average metadata query latency in milliseconds


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_metadata_queries_average_latency_milliseconds` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_metadata_queries_average_latency_milliseconds` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: S3 | Content | timeseries | [Top $TopResources Nodes by Metadata Query Latency](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=6) |
///



### storagegrid_network_received_bytes

Total amount of data received


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_network_received_bytes` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_network_received_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Performance | timeseries | [Network traffic](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=26) |
///



### storagegrid_network_transmitted_bytes

Total amount of data sent


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_network_transmitted_bytes` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_network_transmitted_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Performance | timeseries | [Network traffic](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=26) |
///



### storagegrid_node_cpu_utilization_percentage




| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_node_cpu_utilization_percentage` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_node_cpu_utilization_percentage` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Nodes | timeseries | [Top $TopResources nodes by CPU usage](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=22) |
///



### storagegrid_private_ilm_awaiting_delete_objects

The total number of objects on this node awaiting deletion


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_private_ilm_awaiting_delete_objects` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_private_ilm_awaiting_delete_objects` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Information Lifecycle Management (ILM) | timeseries | [ILM queue (Objects)](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=16) |
///



### storagegrid_private_load_balancer_storage_request_body_bytes_bucket




| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_private_load_balancer_storage_request_body_bytes_bucket` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_private_load_balancer_storage_request_body_bytes_bucket` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Traffic Classification | heatmap | [Write Request Rate by Object Size and Policy](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=50) |
| StorageGrid: Overview | Traffic Classification | heatmap | [Read Request Rate by Object Size and Policy](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=52) |
///



### storagegrid_private_load_balancer_storage_request_count




| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_private_load_balancer_storage_request_count` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_private_load_balancer_storage_request_count` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| ONTAP: StorageGrid FabricPool | Highlights | timeseries | [Top $TopResources Load Balancer Request Completion Rates by Policy](/d/cdot-storagegrid-fabricpool/ontap3a-storagegrid fabricpool?orgId=1&viewPanel=41) |
| StorageGrid: Overview | Performance | timeseries | [Error rate (requests/second)](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=30) |
| StorageGrid: Overview | Performance | timeseries | [Average request duration (non-error)](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=32) |
| StorageGrid: Overview | Traffic Classification | timeseries | [Top $TopResources Load Balancer Request Completion Rates by Policy](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=47) |
| StorageGrid: Overview | Traffic Classification | timeseries | [Top $TopResources Error Response Rates by Policy](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=48) |
| StorageGrid: Overview | Traffic Classification | timeseries | [Top $TopResources Average Request Durations (Non-Error) By Policy](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=49) |
///



### storagegrid_private_load_balancer_storage_request_time




| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_private_load_balancer_storage_request_time` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_private_load_balancer_storage_request_time` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Performance | timeseries | [Average request duration (non-error)](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=32) |
| StorageGrid: Overview | Traffic Classification | timeseries | [Top $TopResources Average Request Durations (Non-Error) By Policy](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=49) |
///



### storagegrid_private_load_balancer_storage_rx_bytes




| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_private_load_balancer_storage_rx_bytes` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_private_load_balancer_storage_rx_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| ONTAP: StorageGrid FabricPool | Highlights | timeseries | [Top $TopResources Load Balancer Request Traffic by Policy](/d/cdot-storagegrid-fabricpool/ontap3a-storagegrid fabricpool?orgId=1&viewPanel=39) |
| StorageGrid: Overview | Traffic Classification | timeseries | [Top $TopResources Load Balancer Request Traffic by Policy](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=46) |
///



### storagegrid_private_load_balancer_storage_tx_bytes




| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_private_load_balancer_storage_tx_bytes` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_private_load_balancer_storage_tx_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| ONTAP: StorageGrid FabricPool | Highlights | timeseries | [Top $TopResources Load Balancer Request Traffic by Policy](/d/cdot-storagegrid-fabricpool/ontap3a-storagegrid fabricpool?orgId=1&viewPanel=39) |
| StorageGrid: Overview | Traffic Classification | timeseries | [Top $TopResources Load Balancer Request Traffic by Policy](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=46) |
///



### storagegrid_private_s3_total_requests

Number of S3 requests.


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_private_s3_total_requests` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_private_s3_total_requests` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Performance | timeseries | [S3 operations](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=24) |
| StorageGrid: S3 | Highlights | timeseries | [Top $TopResources Nodes by GET Operations](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=201) |
| StorageGrid: S3 | Highlights | timeseries | [Top $TopResources Nodes by DELETE Operations](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=203) |
| StorageGrid: S3 | Highlights | timeseries | [Top $TopResources Nodes by PUT Operations](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=202) |
| StorageGrid: S3 | Highlights | timeseries | [Top $TopResources Nodes by HEAD Operations](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=204) |
///



### storagegrid_s3_data_transfers_bytes_ingested

S3 data upload rate (ingestion) in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_s3_data_transfers_bytes_ingested` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_s3_data_transfers_bytes_ingested` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: S3 | Highlights | timeseries | [Top $TopResources Nodes by S3 Upload Rate](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=4) |
///



### storagegrid_s3_data_transfers_bytes_retrieved

S3 data download rate (retrieval) in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_s3_data_transfers_bytes_retrieved` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_s3_data_transfers_bytes_retrieved` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: S3 | Highlights | timeseries | [Top $TopResources Nodes by S3 Download Rate](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=5) |
///



### storagegrid_s3_operations_failed

The total number of failed S3 operations (HTTP status codes 4xx and 5xx), excluding those caused by S3 authorization failure


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_s3_operations_failed` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_s3_operations_failed` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Performance | timeseries | [S3 API requests](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=28) |
| StorageGrid: S3 | Highlights | timeseries | [Top $TopResources Nodes by S3 Operations per Second (failed)](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=2) |
///



### storagegrid_s3_operations_successful

The total number of successful S3 operations (HTTP status code 2xx)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_s3_operations_successful` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_s3_operations_successful` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Performance | timeseries | [S3 API requests](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=28) |
| StorageGrid: S3 | Highlights | timeseries | [Top $TopResources Nodes by S3 Successful Operations](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=1) |
///



### storagegrid_s3_operations_unauthorized

The total number of failed S3 operations that are the result of an authorization failure (HTTP status codes 4xx)


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_s3_operations_unauthorized` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_s3_operations_unauthorized` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Performance | timeseries | [S3 API requests](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=28) |
| StorageGrid: S3 | Highlights | timeseries | [Top $TopResources Nodes by S3 Operations per Second (unauthorized)](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=3) |
///



### storagegrid_storage_utilization_data_bytes

An estimate of the total size of replicated and erasure coded object data on the Storage Node


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_storage_utilization_data_bytes` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_storage_utilization_data_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Highlights | table | [Data space usage breakdown](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=37) |
| StorageGrid: Overview | Highlights | timeseries | [Data storage over time](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=40) |
| StorageGrid: Overview | Nodes | timeseries | [Top $TopResources nodes by data usage](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=20) |
| StorageGrid: S3 | Capacity | timeseries | [Top $TopResources Nodes by Used Storage for Data](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=8) |
///



### storagegrid_storage_utilization_metadata_allowed_bytes

The total space available on storage volume 0 for object metadata


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_storage_utilization_metadata_allowed_bytes` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_storage_utilization_metadata_allowed_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Highlights | table | [Metadata allowed space usage breakdown](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=38) |
///



### storagegrid_storage_utilization_metadata_bytes

The amount of object metadata on storage volume 0, in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_storage_utilization_metadata_bytes` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_storage_utilization_metadata_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Highlights | table | [Metadata allowed space usage breakdown](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=38) |
///



### storagegrid_storage_utilization_total_space_bytes

Total storage space available in bytes


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_storage_utilization_total_space_bytes` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_storage_utilization_total_space_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: S3 | Capacity | timeseries | [Top $TopResources Nodes by Percent Usable Space](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=7) |
///



### storagegrid_storage_utilization_usable_space_bytes

The total amount of object storage space remaining


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `prometheus` | `storagegrid_storage_utilization_usable_space_bytes` | conf/storagegrid/11.6.0/storagegrid_metrics.yaml |

The `storagegrid_storage_utilization_usable_space_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Highlights | table | [Data space usage breakdown](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=37) |
| StorageGrid: Overview | Nodes | timeseries | [Top $TopResources nodes by data usage](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=20) |
| StorageGrid: S3 | Capacity | timeseries | [Top $TopResources Nodes by Percent Usable Space](/d/storagegrid-s3/storagegrid3a-s3?orgId=1&viewPanel=7) |
///



### storagegrid_tenant_usage_data_bytes

The logical size of all objects for the tenant


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `grid/accounts-cache` | `dataBytes` | conf/storagegrid/11.6.0/tenant.yaml |

The `storagegrid_tenant_usage_data_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Storage | timeseries | [Top $TopResources tenants by logical space used](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=12) |
| StorageGrid: Overview | Storage | timeseries | [Top $TopResources tenants by quota usage](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=14) |
///



### storagegrid_tenant_usage_quota_bytes

The maximum amount of logical space available for the tenant's object. If a quota metric is not provided, an unlimited amount of space is available


| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `grid/accounts-cache` | `policy.quotaObjectBytes` | conf/storagegrid/11.6.0/tenant.yaml |

The `storagegrid_tenant_usage_quota_bytes` metric is visualized in the following Grafana dashboards:

/// html | div.grafana-table
| Dashboard | Row | Type | Panel |
|--------|----------|--------|--------|
| StorageGrid: Overview | Storage | timeseries | [Top $TopResources tenants by quota usage](/d/storagegrid-overview/storagegrid3a-overview?orgId=1&viewPanel=14) |
///



