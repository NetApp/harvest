


# InfluxDB Exporter

## Overview

The InfluxDB Exporter will format metrics into the InfluxDB's [line protocol](https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/#naming-restrictions) and write it into a bucket. The Exporter is compatible with InfluxDB v2.0. For explanation about `bucket`, `org` and `precision`, see [InfluxDB API documentation](https://docs.influxdata.com/influxdb/v2.0/api/#tag/Write).


## Parameters
Overview of all parameters is provided below. Notice that only one of `url` and `addr` should be provided (at least one is required).

| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `url`                  | string       | URL of the database, format: `SCHEME://HOST[:PORT]`  |		  			|
| `addr`                 | string       | address of the database, format: `[SCHEME://]HOST`   |		        	|
| `port`                 | int, optional| port of the database                             | `8086`                 |
| `bucket`               | string       | InfluxDB bucket to write                         |                        |
| `org`                  | string       | InfluxDB organization name                       |                        |
| `token`                | string       | [token for authentication](https://docs.influxdata.com/influxdb/v2.0/security/tokens/view-tokens/)                     |                        |
| `precision`            | string       | Preferred timestamp precision in seconds         | `2`                    |
| `client_timeout`       | int, optional| client timeout in seconds                        | `5`                    |
|	|	|	|	|


### Example

snippet from `harvest.yml`:
```yaml
Exporters:
  my_influx:
    exporter: InfluxDB
    addr: localhost
    bucket: harvest
    org: harvest
    token: ZTTrt%24@#WNFM2VZTTNNT25wZWUdtUmhBZEdVUmd3dl@# 
    allow_addrs_regex:
        - `^192.168.0.\d+$`
```

Notice: InfluxDB stores a token in `~/.influxdbv2/configs`, but you can also retrieve it from the UI (usually serving on `localhost:8086`): click on "Data" on the left task bar, then on "Tokens".
