# VictoriaMetrics Exporter

???+ note "VictoriaMetrics Install"

    The information below describes how to setup Harvest's VictoriaMetrics exporter.
    If you need help installing or setting up VictoriaMetrics, check
    out [their documentation](https://docs.victoriametrics.com/victoriametrics/).

## Overview

The VictoriaMetrics Exporter will format metrics into [Prometheus exposition format](https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/#naming-restrictions).
The Exporter is compatible with VictoriaMetrics v1.129.1.

## Parameters

Overview of all parameters is provided below. Only one of `url` or `addr` should be provided and at least one of them is
required.
If `addr` is specified, it should be a valid TCP address or hostname of the VictoriaMetrics server and should not include the
scheme and port.

> `addr` only works with HTTP. If you need to use HTTPS, you should use `url` instead.

If `url` is specified, you must add all arguments to the url.
Harvest will do no additional processing and use exactly what you specify. (
e.g. `url: http://localhost:8428/api/v1/import/prometheus`.
When using `url`, the `addr` and `port` field will be ignored.

| parameter        | type                         | description                                                                                        | default |
|------------------|------------------------------|----------------------------------------------------------------------------------------------------|---------|
| `url`            | string                       | URL of the database, format: `SCHEME://HOST[:PORT]`                                                |         |
| `addr`           | string                       | address of the database, format: `HOST` (HTTP only)                                                |         |
| `port`           | int, optional                | port of the database                                                                               | `8086`  |
| `client_timeout` | int, optional                | client timeout in seconds                                                                          | `5`     |

### Example

snippet from `harvest.yml` using `addr`: (supports HTTP only))

```yaml
Exporters:
  my_victoriametrics:
    exporter: VictoriaMetrics
    addr: localhost
```

snippet from `harvest.yml` using `url`: (supports both HTTP/HTTPS))

```yaml
Exporters:
  victoriametrics2:
    exporter: VictoriaMetrics
    url: http://localhost:8428/api/v1/import/prometheus
```

