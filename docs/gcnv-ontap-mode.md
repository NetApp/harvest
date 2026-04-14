Harvest supports monitoring [Google Cloud NetApp Volumes (GCNV) in ONTAP mode](https://docs.cloud.google.com/netapp/volumes/docs/ontap/overview).

## Poller Configuration

Add a poller to your `harvest.yml` with the full GCNV resource path as `addr` and set `gcnv_ontap_mode: true`.

```yaml
Pollers:
  my-gcnv-poller:
    datacenter: gcp-us-central
    addr: <host>/v1beta1/projects/<project_id>/locations/<location>/storagePools/<pool_name>/ontap
    gcnv_ontap_mode: true
    credentials_script:
      path: /path/to/token-script.sh
      schedule: 5m
      timeout: 10s
    collectors:
      - Rest
      - KeyPerf
      - Ems
    exporters:
      - prometheus1
```

`addr` must be the full GCNV resource path without a scheme — do not include `http://` or `https://`.

GCNV uses Bearer token authentication. Use [`credentials_script`](configure-harvest-basic.md#credentials-script) to supply a token at runtime.

For all poller parameters, see [Poller configuration](configure-harvest-basic.md#pollers).

## Supported Harvest Metrics

Capacity and configuration metrics are available via the `Rest` collector in GCNV ONTAP mode.
However, some metrics may not be available due to permission limitations imposed by the GCNV ONTAP mode environment.
Only limited performance metrics are supported because GCNV ONTAP mode does not support the `ZapiPerf` or `RestPerf` collectors.
Instead, use the [KeyPerf](configure-keyperf.md) collector to gather latency, IOPS, and throughput performance metrics for a limited set of objects.

The following collectors are supported:

- `Rest` — capacity, configuration, and inventory metrics
- [`KeyPerf`](configure-keyperf.md) — latency, IOPS, and throughput performance metrics
- `Ems` — events and alerts

Performance metrics with the API name `KeyPerf` in the [ONTAP metrics documentation](ontap-metrics.md) are supported in GCNV ONTAP mode systems.
As a result, some panels in the dashboards may be missing information.

## Supported Harvest Dashboards

The dashboards that work with GCNV ONTAP mode are tagged with `gcnv-ontap-mode` and listed below:

* ONTAP: LUN
* ONTAP: Security
* ONTAP: SVM
* ONTAP: Volume
* ONTAP: Volume by SVM
* ONTAP: Volume Deep Dive
