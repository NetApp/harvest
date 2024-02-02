> Harvest is the open-metrics endpoint for ONTAP and StorageGRID

NetApp Harvest brings observability to ONTAP and StorageGRID clusters.
Harvest collects performance, capacity and hardware metrics from ONTAP and StorageGRID,
transforms them, and routes them to your choice of a time-series database.

The included Grafana dashboards deliver the datacenter insights you need, while
new metrics can be collected with a few edits of the included template files.

Harvest is open-source, released under an [Apache2 license](https://github.com/NetApp/harvest/blob/main/LICENSE),
and offers great flexibility in how you collect, augment, and export your
datacenter [metrics](https://netapp.github.io/harvest/latest/ontap-metrics/). 

Out-of-the-box Harvest provides a set of pollers, collectors, templates, exporters, an optional auto-discover daemon, and a set of StorageGRID and ONTAP dashboards for Prometheus and Grafana.
Harvest collects the metrics and makes them available to a separately installed instance of Prometheus/InfluxDB and Grafana.

<div class="grid cards" markdown>

- :material-toolbox: [Concepts](concepts.md)
- :material-arrow-right: [Quickstart Guide](quickstart.md)

</div>

If you'd like to familiarize yourself with Harvest's core concepts, we recommend reading [concepts](concepts.md).

If you feel comfortable with the concepts, we recommend our [quickstart guide](quickstart.md),
which takes you through a practical example.

!!! note

    Hop onto our [Discord](https://github.com/NetApp/harvest/blob/main/SUPPORT.md#getting-help) 
    or GitHub [discussions](https://github.com/NetApp/harvest/discussions) and say hi. 👋🏽
