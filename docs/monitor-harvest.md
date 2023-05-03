## Harvest Metadata

Harvest publishes metadata metrics about the key components of Harvest.
Many of these metrics are used in the `Harvest Metadata` dashboard.

If you want to understand more about these metrics, read on!

Metrics are published for:

- collectors
- pollers
- clusters being monitored
- exporters

Here's a high-level summary of the metadata metrics Harvest publishes with details below.

| Metric                         | Description                                                                                                                                                                                                   | Units        |
|:-------------------------------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-------------|
| metadata_collector_api_time    | amount of time to collect data from monitored cluster object                                                                                                                                                  | microseconds |
| metadata_collector_instances   | number of objects collected from monitored cluster                                                                                                                                                            | scalar       |
| metadata_collector_metrics     | number of counters collected from monitored cluster                                                                                                                                                           | scalar       |
| metadata_collector_parse_time  | amount of time to parse XML, JSON, etc. for cluster object                                                                                                                                                    | microseconds |
| metadata_collector_plugin_time | amount of time for all plugins to post-process metrics                                                                                                                                                        | microseconds |
| metadata_collector_poll_time   | amount of time it took for the poll to finish                                                                                                                                                                 | microseconds |
| metadata_collector_task_time   | amount of time it took for each collector's subtasks to complete                                                                                                                                              | microseconds |
| metadata_component_count       | number of metrics collected for each object                                                                                                                                                                   | scalar       |
| metadata_component_status      | status of the collector - 0 means running, 1 means standby, 2 means failed                                                                                                                                    | enum         |
| metadata_exporter_count        | number of metrics and labels exported                                                                                                                                                                         | scalar       |
| metadata_exporter_time         | amount of time it took to render, export, and serve exported data                                                                                                                                             | microseconds |
| metadata_target_goroutines     | number of goroutines that exist within the poller                                                                                                                                                             | scalar       |
| metadata_target_status         | status of the system being monitored. 0 means reachable, 1 means unreachable                                                                                                                                  | enum         |
| metadata_collector_calc_time   | amount of time it took to compute metrics between two successive polls, specifically using properties like raw, delta, rate, average, and percent. This metric is available for ZapiPerf/RestPerf collectors. | microseconds |
| metadata_collector_skips       | number of metrics that were not calculated between two successive polls. This metric is available for ZapiPerf/RestPerf collectors.                                                                           | scalar       |

## Collector Metadata

A poller publishes the metadata metrics for each collector and exporter associated with it.

Let's say we start a poller with the `Zapi` collector and the out-of-the-box `default.yaml` exporting metrics to
Prometheus. That means you will be monitoring 22 different objects (uncommented lines in `default.yaml` as of 23.02).

When we start this poller, we expect it to export 23 `metadata_component_status` metrics. 
One for each of the 22 objects, plus one for the Prometheus exporter.

The following `curl` confirms there are 23 `metadata_component_status` metrics reported.

```bash
curl -s http://localhost:12990/metrics | grep -v "#" | grep metadata_component_status | wc -l
      23
```

These metrics also tell us which collectors are in an standby or failed state. 
For example, filtering on components not in the `running` state shows the following since this cluster doesn't have any `ClusterPeers`, `SecurityAuditDestinations`, or `SnapMirrors`. The `reason` is listed as `no instances` and the metric value is 1 which means standby. 

```bash
curl -s http://localhost:12990/metrics | grep -v "#" | grep metadata_component_status | grep -Evo "running"
metadata_component_status{name="Zapi", reason="no instances",target="ClusterPeer",type="collector",version="23.04.1417"} 1
metadata_component_status{name="Zapi", reason="no instances",target="SecurityAuditDestination",type="collector",version="23.04.1417"} 1
metadata_component_status{name="Zapi", reason="no instances",target="SnapMirror",type="collector",version="23.04.1417"} 1
```

The log files for the poller show a similar story. The poller starts with 22 collectors, but drops to 19 after three of the collectors go to standby because there are no instances to collect. 

```bash
2023-04-17T13:14:18-04:00 INF ./poller.go:539 > updated status, up collectors: 22 (of 22), up exporters: 1 (of 1) Poller=u2
2023-04-17T13:14:18-04:00 INF collector/collector.go:342 > no instances, entering standby Poller=u2 collector=Zapi:SecurityAuditDestination task=data
2023-04-17T13:14:18-04:00 INF collector/collector.go:342 > no instances, entering standby Poller=u2 collector=Zapi:ClusterPeer task=data
2023-04-17T13:14:18-04:00 INF collector/collector.go:342 > no instances, entering standby Poller=u2 collector=Zapi:SnapMirror task=data
2023-04-17T13:15:18-04:00 INF ./poller.go:539 > updated status, up collectors: 19 (of 22), up exporters: 1 (of 1) Poller=u2
```
