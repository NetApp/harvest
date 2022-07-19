


# InfluxDB Exporter

## Overview

The InfluxDB Exporter will format metrics into the InfluxDB's [line protocol](https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/#naming-restrictions) 
and write it into a bucket. 
The Exporter is compatible with InfluxDB v2.0. 
For explanation about `bucket`, `org` and `precision`, see [InfluxDB API documentation](https://docs.influxdata.com/influxdb/v2.0/api/#tag/Write).

If you are monitoring both CDOT and 7mode clusters, it is strongly recommended to use two different buckets.

## Parameters

Overview of all parameters is provided below. Only one of `url` or `addr` should be provided and at least one of them is required.
If `addr` is specified, it should be a valid TCP address or hostname of the InfluxDB server and should not include the scheme.  
When using `addr`, the `bucket`, `org`, and `token` key/values are required.

> `addr` only works with HTTP. If you need to use HTTPS, you should use `url` instead.

If `url` is specified, you must add all arguments to the url. 
Harvest will do no additional processing and use exactly what you specify. (e.g. `url: https://influxdb.example.com:8086/write?db=netapp&u=user&p=pass&precision=2`. 
When using `url`,  the `bucket`, `org`, `port`, and `precision` fields will be ignored.

| parameter        | type                        | description                                                                                        | default |
|------------------|-----------------------------|----------------------------------------------------------------------------------------------------|---------|
| `url`            | string                      | URL of the database, format: `SCHEME://HOST[:PORT]`                                                |         |
| `addr`           | string                      | address of the database, format: `HOST` (HTTP only)                                                |         |
| `port`           | int, optional               | port of the database                                                                               | `8086`  |
| `bucket`         | string, required with `url` | InfluxDB bucket to write                                                                           |         |
| `org`            | string, required with `url` | InfluxDB organization name                                                                         |         |
| `precision`      | string, required with `url` | Preferred timestamp precision in seconds                                                           | `2`     |
| `client_timeout` | int, optional               | client timeout in seconds                                                                          | `5`     |
| `token`          | string                      | [token for authentication](https://docs.influxdata.com/influxdb/v2.0/security/tokens/view-tokens/) |         |

### Example

snippet from `harvest.yml` using `addr`:

```yaml
Exporters:
  my_influx:
    exporter: InfluxDB
    addr: localhost
    bucket: harvest
    org: harvest
    token: ZTTrt%24@#WNFM2VZTTNNT25wZWUdtUmhBZEdVUmd3dl@# 
```

snippet from `harvest.yml` using `url`:

```yaml
Exporters:
  influx2:
    exporter: InfluxDB
    url: http://localhost:8086/api/v2/write?org=harvest&bucket=harvest&precision=s
    token: my-token== 
```

Notice: InfluxDB stores a token in `~/.influxdbv2/configs`, but you can also retrieve it from the UI (usually serving on `localhost:8086`): click on "Data" on the left task bar, then on "Tokens".
