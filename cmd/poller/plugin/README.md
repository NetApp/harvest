# Built-in Plugins


The `plugin` feature allows users to manipulate and customize data collected by collectors without changing the collectors. Plugins have the same capabilities as collectors and therefore can collect data on their own as well. Furthermore, multiple plugins can be put in a pipeline to perform more complex operations.

Harvest architecture defines three types of plugins:

**built-in generic** - Statically compiled, generic plugins. "Generic" means the plugin is collector-agnostic. These plugins are provided in this package.

**dynamic-generic** - These are generic plugins as well, but they are compiled as shared objects and dynamically loaded. These plugins are living in the directory src/plugins.

**dynamic-custom** - These plugins are collector-specific. Their source code should reside inside the plugins/ subdirectory of the collector package. Custom plugins have access to all the parameters of their parent collector and should be therefore treated with great care.

This documentation gives an overview of builtin plugins. For other plugins, see their respective documentation. For writing your own plugin, see Developer's documentation.

The built-in plugins are:

- [Aggregator](#aggregator)
- [LabelAgent](#labelagent)

# Aggregator

Aggregator creates a new collection of metrics (Matrix) by summarizing and/or averaging metric values from an existing Matrix for a given label. For example, if the collected metrics are for volumes, you can create an aggregation for nodes or svms.

### Rule syntax

simplest case:

```yaml
plugins:
  Aggregator:
    - LABEL
# will aggregate a new Matrix based on target label LABEL
```

If you want to specify which labels should be included in the new instances, you can add those space-seperated after `LABEL`:

```yaml
    - LABEL LABEL1,LABEL2
# same, but LABEL1 and LABEL2 will be copied into the new instances
# (default is to only copy LABEL and any global labels (such as cluster and datacenter)
```

Or include all labels:

```yaml
    - LABEL ...
# copy all labels of the original instance
```
By default, aggregated metrics will be prefixed with `LABEL`. For example if the object of the original Matrix is `volume` (meaning metrics are prefixed with `volume_`) and `LABEL` is `aggr`, then the metric `volume_read_ops` will become `aggr_volume_read_ops`, etc. You can override this by providing the `<>OBJ` using the following syntax:

```yaml
    - LABEL<>OBJ
# use OBJ as the object of the new matrix, e.g. if the original object is "volume" and you
# want to leave metric names unchanged, use "volume"
```

Finally, sometimes you only want to aggregate instances with a specific label value. You can use `<VALUE>` for that (optionally follow by `OBJ`):

```yaml
    - LABEL<VALUE>
# aggregate all instances if LABEL has value VALUE
    - LABEL<`VALUE`>
# same, but VALUE is regular expression
    - LABEL<LABELX=`VALUE`>
# same, but check against "LABELX" (instead of "LABEL")
```

Examples:

```yaml
plugins:
  Aggregator:
    # will aggregate metrics of the aggregate. The labels "node" and "type" are included in the new instances
    - aggr node type
    # aggregate instances if label "type" has value "flexgroup"
    # include all original labels
    - type<flexgroup> ...
    # aggregate all instances if value of "volume" ends with underscore and 4 digits
    - volume<`_\d{4}$`>
```

### Aggregation rules

The plugin tries to intelligently aggregate metrics based on a few rules:

- **Sum** - the default rule, if no other rules apply
- **Average** - if any of the following is true:
  - metric name has suffix `_percent` or `_percentage`
  - metric name has prefix `average_` or `avg_`
  - metric has property (`metric.GetProperty()`) `percent` or `average`
- **Weighted Average** - applied if metric has property `average` and suffix `_latency` and if there is a matching `_ops` metric. (This is currently only matching to ZapiPerf metrics, which use the Property field of metrics.)
- **Ignore** - metrics created by some plugins, such as value_to_num by LabelAgent

# LabelAgent
LabelAgent are used to manipulate instance labels based on rules. You can define multiple rules, here is an example of what you could add to the yaml file of a collector:


```yaml
plugins:
  LabelAgent:
  # our rules:
    split_simple: node `/` ,aggr,plex,disk
    replace_regex: node node `^(node)_(\d+)_.*$` `Node-$2`
```

Notice that the rules are executed in the same order as you've added them. List of currently available rules:

- [Built-in Plugins](#built-in-plugins)
- [Aggregator](#aggregator)
    - [Rule syntax](#rule-syntax)
    - [Aggregation rules](#aggregation-rules)
- [LabelAgent](#labelagent)
  - [split](#split)
  - [split_regex](#split_regex)
  - [split_pairs](#split_pairs)
  - [join](#join)
  - [replace](#replace)
  - [replace_regex](#replace_regex)
  - [exclude_equals](#exclude_equals)
  - [exclude_contains](#exclude_contains)
  - [exclude_regex](#exclude_regex)
  - [include_equals](#include_equals)
  - [include_contains](#include_contains)
  - [include_regex](#include_regex)
  - [value_mapping](#value_mapping)
  - [value_to_num](#value_to_num)
  - [compute_metric](#compute_metric)
  - [value_to_num_regex](#value_to_num_regex)

## split

Rule syntax:

```yaml
split: 
  - LABEL `SEP` LABEL1,LABEL2,LABEL3
# source label - separator - comma-seperated target labels
```

Splits the value of a given label by separator `SEP` and creates new labels if their number matches to the number of target labels defined in rule. To discard a subvalue, just add a redundant `,` in the names of the target labels. 

Example:

```yaml
split: 
  - node `/` ,aggr,plex,disk
# will split the value of "node" using separator "/"
# will expect 4 values: first will be discarded, remaining
# three will be stored as labels "aggr", "plex" and "disk"
```

## split_regex

Does the same as `split_regex` but uses a regular expression instead of a separator.

Rule syntax:

```yaml
split_simple: 
  - LABEL `REGEX` LABEL1,LABEL2,LABEL3
```

Example:

```yaml
split_simple: 
  - node `.*_(ag\d+)_(p\d+)_(d\d+)` aggr,plex,disk
# will look for "_ag", "_p", "_d", each followed by one
# or more numbers, if there is a match, the submatches
# will be stored as "aggr", "plex" and "disk"
```


## split_pairs

Rule syntax:

```yaml
split_pairs: 
  - LABEL `SEP1` `SEP2`
# source label - pair separator - key-value separator
```

Extracts key-value pairs from the value of source label `LABEL`. Note that you need to add these keys in the export options, otherwise they will not be exported.

Example:

```yaml
split_pairs: 
  - comment ` ` `:`
# will split pairs using a single space and split key-values using colon
# e.g. if comment="owner:jack contact:some@email", the result wll be
# two new labels: owner="jack" and contact="some@email"
```

## join

Join multiple label values using separator `SEP` and create a new label. 

Rule syntax:

```yaml
join: 
  - LABEL `SEP` LABEL1,LABEL2,LABEL3
# target label - separator - comma-seperated source labels
```

Example:

```yaml
join: 
  - plex_long `_` aggr,plex
# will look for the values of labels "aggr" and "plex",
# if they are set, a new "plex_long" label will be added
# by joining their values with "_"
```

## replace

Substitute substring `OLD` with `NEW` in label `SOURCE` and store in `TARGET`. Note that target and source labels can be the same.

Rule syntax:

```yaml
replace: 
  - SOURCE TARGET `OLD` `NEW`
# source label - target label - substring to replace - replace with
```

Example:

```yaml
replace: 
  - node node_short `node_` ``
# this rule will just remove "node_" from all values of label
# "node". E.g. if label is "node_jamaica1", it will rewrite it 
# as "jamaica1"
```

## replace_regex
Same as `replace_simple`, but will use a regular expression instead of `OLD`. Note you can use `$n` to specify `n`th submatch in `NEW`.

Rule syntax:

```yaml
replace_regex: 
  - SOURCE TARGET `REGEX` `NEW`
# source label - target label - substring to replace - replace with
```

Example:

```yaml
replace_regex: 
  - node node `^(node)_(\d+)_.*$` `Node-$2`
# if there is a match, will capitalize "Node" and remove suffixes.
# E.g. if label is "node_10_dc2", it will rewrite it as
# will rewrite it as "Node-10"
```

## exclude_equals

Exclude each instance, if the value of `LABEL` is exactly `VALUE`. Exclude means that metrics for this instance will not be exported.

Rule syntax:

```yaml
exclude_equals: 
  - LABEL `VALUE`
# label name - label value
```

Example:

```yaml
exclude_equals: 
  - vol_type `flexgroup_constituent`
# all instances, which have label "vol_type" with value
# "flexgroup_constituent" will not be exported
```

## exclude_contains

Same as `exclude_equals`, but all labels that *contain* `VALUE` will be excluded

Rule syntax:

```yaml
exclude_contains: 
  - LABEL `VALUE`
# label name - label value
```

Example:

```yaml
exclude_contains: 
  - vol_type `flexgroup_`
# all instances, which have label "vol_type" which contain
# "flexgroup_" will not be exported
```

## exclude_regex

Same as `exclude_equals`, but will use a regular expression and all matching instances will be excluded.

Rule syntax:

```yaml
exclude_regex: 
  - LABEL `REGEX`
# label name - regular expression
```

Example:

```yaml
exclude_regex: 
  - vol_type `^flex`
# all instances, which have label "vol_type" which starts with
# "flex" will not be exported
```

## include_equals

Include each instance, if the value of `LABEL` is exactly `VALUE`. Include means that metrics for this instance will be exported and instances that do not match will not be exported.

Rule syntax:

```yaml
include_equals: 
  - LABEL `VALUE`
# label name - label value
```

Example:

```yaml
include_equals: 
  - vol_type `flexgroup_constituent`
# all instances, which have label "vol_type" with value
# "flexgroup_constituent" will be exported
```

## include_contains

Same as `include_equals`, but all labels that *contain* `VALUE` will be included

Rule syntax:

```yaml
include_contains: 
  - LABEL `VALUE`
# label name - label value
```

Example:

```yaml
include_contains: 
  - vol_type `flexgroup_`
# all instances, which have label "vol_type" which contain
# "flexgroup_" will be exported
```

## include_regex

Same as `include_equals`, but a regular expression will be used for inclusion. Similar to the other includes, all matching instances will be included and all non-matching will not be exported.

Rule syntax:

```yaml
include_regex: 
  - LABEL `REGEX`
# label name - regular expression
```

Example:

```yaml
include_regex: 
  - vol_type `^flex`
# all instances, which have label "vol_type" which starts with
# "flex" will be exported
```

## value_mapping

value_mapping was deprecated in 21.11 and removed in 22.02. Use [value_to_num mapping](#value_to_num) instead.

## value_to_num

Map values of a given label to a numeric metric (of type `uint8`).
This rule maps values of a given label to a numeric metric (of type `unit8`). Healthy is mapped to 1 and all non-healthy values are mapped to 0.

This is handy to manipulate the data in the DB or Grafana (e.g. change color based on status or create alert).

Note that you don't define the numeric values yourself, instead, you only provide the possible (expected) values, the plugin will map each value to its index in the rule.

Rule syntax:

```yaml
value_to_num: 
  - METRIC LABEL ZAPI_VALUE REST_VALUE `N`
# map values of LABEL to 1 if it is ZAPI_VALUE or REST_VALUE
# otherwise, value of METRIC is set to N
```
The default value `N` is optional, if no default value is given and the label value does not match any of the given values, the metric value will not be set.

Examples:

```yaml
value_to_num: 
  - status state up online `0`
# a new metric will be created with the name "status"
# if an instance has label "state" with value "up", the metric value will be 1,
# if it's "online", the value will be set to 1,
# if it's any other value, it will be set to the specified default, 0
```

```yaml
value_to_num: 
  - status state up online `4`
# metric value will be set to 1 if "state" is "up", otherwise to **4**
```

```yaml
value_to_num: 
  - status outage - - `0` #ok_value is empty value. 
# metric value will be set to 1 if "outage" is empty, if it's any other value, it will be set to the default, 0
# '-' is a special symbol in this mapping, and it will be converted to blank while processing.
```
## compute_metric

This rule creates the new metric (of type `float64`) using the existing metric values.

Note that you don't define the numeric values yourself, instead, you only provide the metric names with operation, 
the plugin will fetch the value of each given metric and store the result of mathematical operation in new given custom metric.

Currently, we support these operations: ADD SUBTRACT MULTIPLY DIVIDE

Rule syntax:

```yaml
compute_metric: 
  - METRIC OPERATION METRIC1 METRIC2 METRIC3 
# target new metric - mathematical operation - input metric names 
# apply OPERATION on metric values of METRIC1, METRIC2 and METRIC3 and set result in METRIC
```

Examples:

```yaml
compute_metric:
  - space_total ADD space_available space_used
# a new metric will be created with the name "space_total"
# if an instance has metric "space_available" with value "1000", and "space_used" with value "400",
# the result value will be "1400" and set to metric "space_total".
```

```yaml
compute_metric:
  - disk_count ADD primary.disk_count secondary.disk_count hybrid.disk_count
# value of metric "disk_count" would be addition of all the given disk_counts metric values.
# disk_count = primary.disk_count + secondary.disk_count + hybrid.disk_count
```

```yaml
compute_metric:
  - files_available SUBTRACT files files_used
# value of metric "files_available" would be subtraction of the metric value of files_used from metric value of files.
# files_available = files - files_used
```

```yaml
compute_metric:
  - total_bytes MULTIPLY bytes_per_sector sector_count
# value of metric "total_bytes" would be multiplication of metric value of bytes_per_sector and metric value of sector_count.
# total_bytes = bytes_per_sector * sector_count
```

```yaml
compute_metric:
  - transmission_rate DIVIDE transfer.bytes_transferred transfer.total_duration
# value of metric "transmission_rate" would be division of metric value of transfer.bytes_transferred by metric value of transfer.total_duration.
# transmission_rate = transfer.bytes_transferred / transfer.total_duration
```

## value_to_num_regex

Same as value_to_num, but will use a regular expression and all matching are mapped to 1 and other mapped to 0.

This is handy to manipulate the data in the DB or Grafana (e.g. change color based on status or create alert).

Note that you don't define the numeric values yourself, instead, you only provide the possible (expected) values, the plugin will map each value to its index in the rule.

Rule syntax:

```yaml
value_to_num_regex:
  - METRIC LABEL ZAPI_REGEX REST_REGEX `N`
# map values of LABEL to 1 if it matches ZAPI_REGEX or REST_REGEX
# otherwise, value of METRIC is set to N
```
The default value `N` is optional, if no default value is given and the label value does not match any of the given values, the metric value will not be set.

Examples:

```yaml
value_to_num_regex:
  - certificateuser methods .*cert.*$ .*certificate.*$ `0`
# a new metric will be created with the name "certificateuser"
# if an instance has label "methods" with value contains "cert", the metric value will be 1,
# if value contains "certificate", the value will be set to 1,
# if value doesn't contain "cert" and "certificate", it will be set to the specified default, 0
```

```yaml
value_to_num_regex:
  - status state ^up$ ^ok$ `4`
# metric value will be set to 1 if label "state" matches regex, otherwise set to **4**
```
