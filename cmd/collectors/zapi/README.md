

# Zapi Collector

Zapi collects data from ONTAP systems using the ZAPI protocol. The collector submits data as it is and does not perform any calculations (therefore it is not able to collect `perf` objects). Since the attributes of most APIs have an irregular tree structure, sometimes a plugin will be required to collect metrics from an API.


## Configuration

The parameters and configuration are similar to those of the [ZapiPerf collector](../zapiperf/README.md). Only the differences will be discussed below.

### Collector configuration file

Parameters different from ZapiPerf:


| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `schedule`             | required     | same as for ZapiPerf, but only two elements: `instance` and `data` (collector does not run a `counter` poll) ||
| `no_max_records`       | bool, optional | don't add `max-records` to the ZAPI request    |                        |
| `collect_only_labels`  | bool, optional | don't look for numeric metrics, only submit labels  (suppresses the `ErrNoMetrics` error)| |
| `only_cluster_instance` | bool, optional | don't look for instance keys and assume only instance is the cluster itself ||


### Object configuration file

The Zapi collector does not have the parameters `instance_key` and `override` parameters.

#### `counters`

This section contains the complete or partial attribute tree of the queried API. Since the collector does get counter metadata from the ONTAP system, two additional symbols are used for non-numeric attributes:

- `^` used as a prefix indicates that the attribute should be stored as a label
- `^^` indicates that the attribute is a label and an instance key (i.e. a label that uniquely identifies an instance, such as `name`, `uuid`). If a single label does not uniquely identify an instance, then multiple instance keys should be indicated.

#### Creating/editing object configurations

The Zapi tool can help to create or edit subtemplates. Examples:

```sh
$ harvest zapi --poller <poller> show apis
  # will print list of apis that are available
  # usually apis with the "get-iter" suffix can provide useful metrics
$ harvest zapi --poller <poller> show attrs --api volume-get-iter
  # will print the attribute tree of the API
$ harvest zapi --poller <poller> show data --api volume-get-iter
  # will print raw data of the API attribute tree
```

Replace `<poller>` with the name of a poller that can connect to an ONTAP system.
