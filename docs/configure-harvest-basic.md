The main configuration file, `harvest.yml`, consists of the following sections, described below:

## Pollers

All pollers are defined in `harvest.yml`, the main configuration file of Harvest, under the section `Pollers`.

| parameter              | type                                           | description                                                                                                                                                                                                                                                                                                                                                               | default            |
|------------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------|
| Poller name (header)   | **required**                                   | Poller name, user-defined value                                                                                                                                                                                                                                                                                                                                           |                    |
| `datacenter`           | **required**                                   | Datacenter name, user-defined value                                                                                                                                                                                                                                                                                                                                       |                    |
| `addr`                 | required by some collectors                    | IPv4 or FQDN of the target system                                                                                                                                                                                                                                                                                                                                         |                    |
| `collectors`           | **required**                                   | List of collectors to run for this poller                                                                                                                                                                                                                                                                                                                                 |                    |
| `exporters`            | **required**                                   | List of exporter names from the `Exporters` section. Note: this should be the name of the exporter (e.g. `prometheus1`), not the value of the `exporter` key (e.g. `Prometheus`)                                                                                                                                                                                          |                    |
| `auth_style`           | required by Zapi* collectors                   | Either `basic_auth` or `certificate_auth`                                                                                                                                                                                                                                                                                                                                 | `basic_auth`       |
| `username`, `password` | required if `auth_style` is `basic_auth`       |                                                                                                                                                                                                                                                                                                                                                                           |                    |
| `ssl_cert`, `ssl_key`  | optional if `auth_style` is `certificate_auth` | Absolute paths to SSL (client) certificate and key used to authenticate with the target system.<br /><br />If not provided, the poller will look for `<hostname>.key` and `<hostname>.pem` in `$HARVEST_HOME/cert/`.<br/><br/>To create certificates for ONTAP systems, see [using certificate authentication](prepare-cdot-clusters.md#using-certificate-authentication) |                    |
| `use_insecure_tls`     | optional, bool                                 | If true, disable TLS verification when connecting to ONTAP cluster                                                                                                                                                                                                                                                                                                        | false              |
| `credentials_file`     | optional, string                               | Path to a yaml file that contains cluster credentials. The file should have the same shape as `harvest.yml`. See [here](configure-harvest-basic.md#credentials-file) for examples. Path can be relative to `harvest.yml` or absolute                                                                                                                                      |                    |          
| `credentials_script`   | optional, section                              | Section that defines how Harvest should fetch credentials via external script. See [here](configure-harvest-basic.md#credentials-script) for details.                                                                                                                                                                                                                     |                    |          
| `tls_min_version`      | optional, string                               | Minimum TLS version to use when connecting to ONTAP cluster: One of tls10, tls11, tls12 or tls13                                                                                                                                                                                                                                                                          | Platform decides   | 
| `labels`               | optional, list of key-value pairs              | Each of the key-value pairs will be added to a poller's metrics. Details [below](configure-harvest-basic.md#labels)                                                                                                                                                                                                                                                       |                    |
| `log_max_bytes`        |                                                | Maximum size of the log file before it will be rotated                                                                                                                                                                                                                                                                                                                    | `5_242_880` (5 MB) |
| `log_max_files`        |                                                | Number of rotated log files to keep                                                                                                                                                                                                                                                                                                                                       | `5`                |
| `log`                  | optional, list of collector names              | Matching collectors log their ZAPI request/response                                                                                                                                                                                                                                                                                                                       |                    |
| `prefer_zapi`          | optional, bool                                 | Use the ZAPI API if the cluster supports it, otherwise allow Harvest to choose REST or ZAPI, whichever is appropriate to the ONTAP version. See [rest-strategy](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-strategy.md) for details.                                                                                                              |                    |

## Defaults

This section is optional. If there are parameters identical for all your pollers (e.g. datacenter, authentication
method, login preferences), they can be grouped under this section. The poller section will be checked first and if the
values aren't found there, the defaults will be consulted.

## Exporters

All exporters need two types of parameters:

- `exporter parameters` - defined in `harvest.yml` under `Exporters` section
- `export_options` - these options are defined in the `Matrix` data structure that is emitted from collectors and
  plugins

The following two parameters are required for all exporters:

| parameter              | type         | description                                                                                                            | default |
|------------------------|--------------|------------------------------------------------------------------------------------------------------------------------|---------|
| Exporter name (header) | **required** | Name of the exporter instance, this is a user-defined value                                                            |         |
| `exporter`             | **required** | Name of the exporter class (e.g. Prometheus, InfluxDB, Http) - these can be found under the `cmd/exporters/` directory |         |

Note: when we talk about the *Prometheus Exporter* or *InfluxDB Exporter*, we mean the Harvest modules that send the
data to a database, NOT the names used to refer to the actual databases.

### [Prometheus Exporter](prometheus-exporter.md)

### [InfluxDB Exporter](influxdb-exporter.md)

## Tools

This section is optional. You can uncomment the `grafana_api_token` key and add your Grafana API token so `harvest` does
not prompt you for the key when importing dashboards.

```
Tools:
  #grafana_api_token: 'aaa-bbb-ccc-ddd'
```

## Configuring collectors

Collectors are configured by their own configuration files ([templates](configure-templates.md)), which are stored in subdirectories
in [conf/](https://github.com/NetApp/harvest/tree/main/conf).
Most collectors run concurrently and collect a subset of related metrics.
For example, node related metrics are grouped together and run independently of the disk related metrics.
Below is a snippet from `conf/zapi/default.yaml`

In this example, the `default.yaml` template contains a list of objects (e.g. Node) that reference sub-templates (e.g.
node.yaml). This decomposition groups related metrics together and at runtime, a `Zapi` collector per object will be
created and each of these collectors will run concurrently.

Using the snippet below, we expect there to be four `Zapi` collectors running, each with a different subtemplate and
object.

```
collector:          Zapi
objects:
  Node:             node.yaml
  Aggregate:        aggr.yaml
  Volume:           volume.yaml
  SnapMirror:       snapmirror.yaml
```

At start-up, Harvest looks for two files (`default.yaml` and `custom.yaml`) in the `conf` directory of the
collector (e.g. `conf/zapi/default.yaml`).
The `default.yaml` is installed by default, while the `custom.yaml` is an optional file
you can create
to [add new templates](configure-templates.md#creatingediting-templates).

When present, the `custom.yaml` file will be merged with the `default.yaml` file.
This behavior can be overridden in your `harvest.yml`, see
[here](https://github.com/NetApp/harvest/blob/main/pkg/conf/testdata/issue_396.yaml) for an example.

For a list of collector-specific parameters, refer to their individual documentation.

#### [Zapi and ZapiPerf](configure-zapi.md)

#### [Rest and RestPerf](configure-rest.md)

#### [EMS](configure-ems.md)

#### [StorageGRID](configure-storagegrid.md)

#### [Unix](configure-unix.md)

## Labels

Labels offer a way to add additional key-value pairs to a poller's metrics. These allow you to tag a cluster's metrics
in a cross-cutting fashion. Here's an example:

```
  cluster-03:
    datacenter: DC-01
    addr: 10.0.1.1
    labels:
      - org: meg       # add an org label with the value "meg"
      - ns:  rtp       # add a namespace label with the value "rtp"
```

These settings add two key-value pairs to each metric collected from `cluster-03` like this:

```
node_vol_cifs_write_data{org="meg",ns="rtp",datacenter="DC-01",cluster="cluster-03",node="umeng-aff300-05"} 10
```

Keep in mind that each unique combination of key-value pairs increases the amount of stored data. Use them sparingly.
See [PrometheusNaming](https://prometheus.io/docs/practices/naming/#labels) for details.

## Credentials File

If you would rather not list cluster credentials in your `harvest.yml`, you can use the `credentials_file` section
in your `harvest.yml` to point to a file that contains the credentials.
At runtime, the `credentials_file` will be read and the included credentials will be used to authenticate with the
matching cluster(s).

This is handy when integrating with 3rd party credential stores.
See #884 for examples.

The format of the `credentials_file` is similar to `harvest.yml` and can contain multiple cluster credentials.

Example:

Snippet from `harvest.yml`:

```yaml
Pollers:
  cluster1:
    addr: 10.193.48.11
    credentials_file: secrets/cluster1.yml
    exporters:
      - prom1 
```

File `secrets/cluster1.yml`:

```yaml
Pollers:
  cluster1:
    username: harvest
    password: foo
```

## Credentials Script

You can fetch authentication information via an external script by using the `credentials_script` section in
the `Pollers` section of your `harvest.yml` as shown in the [example below](#example). 

At runtime, Harvest will invoke the script referenced in the `credentials_script` `path` section. 
Harvest will call the script with two arguments via `standard in`, in this order:
1. address of the cluster taken from your `harvest.yaml` file, section `Pollers` `addr`
2. username of the cluster taken from your `harvest.yaml` file, section `Pollers` `username`

The script should use the two arguments to look up and return the password via the script's `standard out`.
If the script doesn't finish within the specified `timeout`, Harvest will kill the script and any spawned processes.

Credential scripts are defined in your `harvest.yml` under the `Pollers` `credentials_script` section. 
Below are the options for the `credentials_script` section  

| parameter | type                    | description                                                                                                                                                                    | default |
|-----------|-------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| path      | string                  | absolute path to script that takes two arguments: addr and username, in that order                                                                                             |         |
| schedule  | go duration or `always` | schedule used to call the authentication script. If the value is `always`, the script will be called everytime a password is requested, otherwise use the earlier cached value | 24h     |
| timeout   | go duration             | amount of time Harvest will wait for the script to finish before killing it and descendents                                                                                    | 10s     |

### Example

```yaml
Pollers:
    ontap1:
        datacenter: rtp
        addr: 10.1.1.1
        collectors:
            - Rest
            - RestPerf
        credentials_script:
            path: ./get_pass
            schedule: 3h
            timeout: 10s
```

### Troubleshooting

* Make sure your script is executable 
* Ensure the user/group that executes your poller also has read and execute permissions on the script. 
  `su` as the user/group that runs Harvest and make sure you can execute the script too. 
