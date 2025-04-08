Since Harvest 24.11.1, NetApp ASA r2 systems monitoring is supported.

## Supported Harvest Metrics

Most of the capacity metrics collected via the REST collector are available in ASA r2 monitoring. However, there are limited performance metrics supported in ASA r2 systems.
In general, ASA r2 clusters only provide latency, IOPS, and throughput metrics for a handful of objects.
Harvest collects these performance metrics using the [KeyPerf](configure-keyperf.md) collector.

Performance metrics with the API name `KeyPerf` in the [ONTAP metrics documentation](ontap-metrics.md) are supported in ASA r2 systems.
This means that some panels in the dashboards may be missing information.