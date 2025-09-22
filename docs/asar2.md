Since Harvest 24.11.1, NetApp ASA r2 systems monitoring is supported. We recommend using the latest Harvest version to ensure compatibility with ASA r2 systems.


## Prepare ASA r2 cluster

You need to prepare your ASA r2 cluster for monitoring by following the steps outlined in [monitoring ONTAP systems](prepare-cdot-clusters.md#rest-least-privilege-role), and then perform the following steps.

#### REST least-privilege role

Verify there are no errors when you copy/paste these. Warnings are fine.

```shell
security login rest-role create -role harvest-rest-role -access readonly -api /api/storage/storage-units
security login rest-role create -role harvest-rest-role -access readonly -api /api/storage/availability-zones
```

## Supported Harvest Metrics

Most capacity metrics collected via the REST collector are available in ASA r2 monitoring. However, only limited performance metrics are supported in ASA r2 systems. That is because ASA r2 clusters do not support the `RestPerf` collector. Instead, Harvest uses the [KeyPerf](configure-keyperf.md) collector to gather latency, IOPS, and throughput performance metrics for a limited set of objects.

Harvest automatically detects ASA r2 systems and replaces any `ZapiPerf` or `RestPerf` collectors with the `KeyPerf` collector. 

We also recommend enabling the [StatPerf](configure-statperf.md) collector for ASA r2 systems to collect performance metrics that are not available via the `KeyPerf` collector. You need to make sure the `StatPerf` collector is listed first in your list of collectors. e.g. 

```yaml
collectors:
    - StatPerf
    - Rest
    - RestPerf
```

Performance metrics with the API name `KeyPerf` or `StatPerf` in the [ONTAP metrics documentation](ontap-metrics.md) are supported in ASA r2 systems.
As a result, some panels in the dashboards may be missing information. 