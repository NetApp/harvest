ONTAP may report partial aggregate results for certain objects during events such as node outages or cluster disruptions.
When this occurs, Harvest's performance collectors (ZapiPerf, RestPerf, KeyPerf, and StatPerf) will skip reporting performance counters for the affected objects to ensure data accuracy.

#### Identifying Partial Aggregation

To determine whether partial aggregation affects an object, check the `numPartials` entry in the Harvest logs.
If `numPartials` is greater than zero, it indicates that partial aggregations have occurred for that object.

For example, consider the log entry:
`Collected Poller=aff-251 collector=ZapiPerf:NFSv4 apiMs=870 bytesRx=3640 calcMs=0 exportMs=0 instances=56 instancesExported=41 metrics=400 metricsExported=340 numCalls=1 numPartials=15`

In this example, 15 out of 56 NFSv4 instances experienced partial aggregation.

#### Node-Scoped Objects Exception

Node-scoped objects are not affected by issues related to partial aggregation. For these objects, Harvest emits metrics even when ONTAP reports partial aggregation.
This behavior is controlled by the `allow_partial_aggregation` flag in the object's template configuration.

When `allow_partial_aggregation: true` is set in a template, Harvest will continue to collect and emit metrics for that object regardless of the partial aggregation status.