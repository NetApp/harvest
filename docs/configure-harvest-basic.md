The main configuration file, `harvest.yml`, consists of the following sections, described below:

## Pollers

All pollers are defined in `harvest.yml`, the main configuration file of Harvest, under the section `Pollers`.

| parameter              | type                                           | description                                                                                                                                                                                                                                                                                                                                                               | default          |
|------------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------|
| Poller name (header)   | **required**                                   | Poller name, user-defined value                                                                                                                                                                                                                                                                                                                                           |                  |
| `datacenter`           | **required**                                   | Datacenter name, user-defined value                                                                                                                                                                                                                                                                                                                                       |                  |
| `addr`                 | required by some collectors                    | IPv4 or FQDN of the target system                                                                                                                                                                                                                                                                                                                                         |                  |
| `collectors`           | **required**                                   | List of collectors to run for this poller                                                                                                                                                                                                                                                                                                                                 |                  |
| `exporters`            | **required**                                   | List of exporter names from the `Exporters` section. Note: this should be the name of the exporter (e.g. `prometheus1`), not the value of the `exporter` key (e.g. `Prometheus`)                                                                                                                                                                                          |                  |
| `auth_style`           | required by Zapi* collectors                   | Either `basic_auth` or `certificate_auth` See [authentication](#authentication) for details                                                                                                                                                                                                                                                                               | `basic_auth`     |
| `username`, `password` | required if `auth_style` is `basic_auth`       |                                                                                                                                                                                                                                                                                                                                                                           |                  |
| `ssl_cert`, `ssl_key`  | optional if `auth_style` is `certificate_auth` | Paths to SSL (client) certificate and key used to authenticate with the target system.<br /><br />If not provided, the poller will look for `<hostname>.key` and `<hostname>.pem` in `$HARVEST_HOME/cert/`.<br/><br/>To create certificates for ONTAP systems, see [using certificate authentication](prepare-cdot-clusters.md#using-certificate-authentication)          |                  |
| `ca_cert`              | optional if `auth_style` is `certificate_auth` | Path to file that contains PEM encoded certificates. Harvest will append these certificates to the system-wide set of root certificate authorities (CA).<br /><br />If not provided, the OS's root CAs will be used.<br/><br/>To create certificates for ONTAP systems, see [using certificate authentication](prepare-cdot-clusters.md#using-certificate-authentication) |                  |
| `use_insecure_tls`     | optional, bool                                 | If true, disable TLS verification when connecting to ONTAP cluster                                                                                                                                                                                                                                                                                                        | false            |
| `credentials_file`     | optional, string                               | Path to a yaml file that contains cluster credentials. The file should have the same shape as `harvest.yml`. See [here](configure-harvest-basic.md#credentials-file) for examples. Path can be relative to `harvest.yml` or absolute.                                                                                                                                     |                  |          
| `credentials_script`   | optional, section                              | Section that defines how Harvest should fetch credentials via external script. See [here](configure-harvest-basic.md#credentials-script) for details.                                                                                                                                                                                                                     |                  |          
| `tls_min_version`      | optional, string                               | Minimum TLS version to use when connecting to ONTAP cluster: One of tls10, tls11, tls12 or tls13                                                                                                                                                                                                                                                                          | Platform decides | 
| `labels`               | optional, list of key-value pairs              | Each of the key-value pairs will be added to a poller's metrics. Details [below](configure-harvest-basic.md#labels)                                                                                                                                                                                                                                                       |                  |
| `log_max_bytes`        |                                                | Maximum size of the log file before it will be rotated                                                                                                                                                                                                                                                                                                                    | `10 MB`          |
| `log_max_files`        |                                                | Number of rotated log files to keep                                                                                                                                                                                                                                                                                                                                       | `5`              |
| `log`                  | optional, list of collector names              | Matching collectors log their ZAPI request/response                                                                                                                                                                                                                                                                                                                       |                  |
| `prefer_zapi`          | optional, bool                                 | Use the ZAPI API if the cluster supports it, otherwise allow Harvest to choose REST or ZAPI, whichever is appropriate to the ONTAP version. See [rest-strategy](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-strategy.md) for details.                                                                                                              |                  |
| `conf_path`            | optional, `:` seperated list of directories    | The search path Harvest uses to load its [templates](configure-templates.md). Harvest walks each directory in order, stopping at the first one that contains the desired template.                                                                                                                                                                                        | conf             |

## Defaults

This section is optional.
If there are parameters identical for all your pollers (e.g., datacenter, authentication
method, login preferences), they can be grouped under this section.
The poller section will be checked first, and if the
values aren't found there, the defaults will be consulted.

## Exporters

All exporters need two types of parameters:

- `exporter parameters` - defined in `harvest.yml` under `Exporters` section
- `export_options` - these options are defined in the `Matrix` data structure emitted from collectors and
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

## Poller_files

Harvest supports loading pollers from multiple files specified in the `Poller_files` section of your `harvest.yml` file.
For example, the following snippet tells harvest to load pollers from all the `*.yml` files under the `configs` directory, 
and from the `path/to/single.yml` file.

Paths may be relative or absolute.

```yaml
Poller_files:
    - configs/*.yml
    - path/to/single.yml

Pollers:
    u2:
        datacenter: dc-1
```

Each referenced file can contain one or more unique pollers.
Ensure that you include the top-level `Pollers` section in these files.
All other top-level sections will be ignored.
For example:

```yaml
# contents of configs/00-rtp.yml
Pollers:
  ntap3:
    datacenter: rtp

  ntap4:
    datacenter: rtp
---
# contents of configs/01-rtp.yml
Pollers:
  ntap5:
    datacenter: blr
---
# contents of path/to/single.yml
Pollers:
  ntap1:
    datacenter: dc-1

  ntap2:
    datacenter: dc-1
```

At runtime, all files will be read and combined into a single configuration.
The example above would result in the following set of pollers in this order.
```yaml
- u2
- ntap3
- ntap4
- ntap5
- ntap1
- ntap2
```

When using glob patterns, the list of matching paths will be sorted before they are read.
Errors will be logged for all duplicate pollers and Harvest will refuse to start.

## Configuring collectors

Collectors are configured by their own configuration files ([templates](configure-templates.md)), which are stored in subdirectories
in [conf/](https://github.com/NetApp/harvest/tree/main/conf).
Most collectors run concurrently and collect a subset of related metrics.
For example, node related metrics are grouped together and run independently of the disk-related metrics.
Below is a snippet from `conf/zapi/default.yaml`

In this example, the `default.yaml` template contains a list of objects (e.g., Node) that reference sub-templates (e.g.,
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

# Authentication

When authenticating with ONTAP and StorageGRID clusters,
Harvest supports both client certificates and basic authentication.

These methods of authentication are defined in the `Pollers` or `Defaults` section of your `harvest.yml` using one or more
of the following parameters.

| parameter            | description                                                                                              | default      | Link                        |
|----------------------|----------------------------------------------------------------------------------------------------------|--------------|-----------------------------|
| `auth_sytle`         | One of `basic_auth` or `certificate_auth` Optional when using `credentials_file` or `credentials_script` | `basic_auth` | [link](#Pollers)            |
| `username`           | Username used for authenticating to the remote system                                                    |              | [link](#Pollers)            |
| `password`           | Password used for authenticating to the remote system                                                    |              | [link](#Pollers)            |
| `credentials_file`   | Relative or absolute path to a yaml file that contains cluster credentials                               |              | [link](#credentials-file)   |
| `credentials_script` | External script Harvest executes to retrieve credentials                                                 |              | [link](#credentials-script) |

## Precedence

When multiple authentication parameters are defined at the same time,
Harvest tries each method listed below, in the following order, to resolve authentication requests. 
The first method that returns a non-empty password stops the search. 

When these parameters exist in both the `Pollers` and `Defaults` section,
the `Pollers` section will be consulted before the `Defaults`.

| section    | parameter                                           |
|------------|-----------------------------------------------------|
| `Pollers`  | auth_style: `certificate_auth`                      |
| `Pollers`  | auth_style: `basic_auth` with username and password |
| `Pollers`  | `credentials_script`                                |
| `Pollers`  | `credentials_file`                                  |
| `Defaults` | auth_style: `certificate_auth`                      |
| `Defaults` | auth_style: `basic_auth` with username and password |
| `Defaults` | `credentials_script`                                |
| `Defaults` | `credentials_file`                                  |

## Credentials File

If you would rather not list cluster credentials in your `harvest.yml`, you can use the `credentials_file` section
in your `harvest.yml` to point to a file that contains the credentials.
At runtime, the `credentials_file` will be read and the included credentials will be used to authenticate with the
matching cluster(s).

This is handy when integrating with 3rd party credential stores.
See [#884](https://github.com/NetApp/harvest/discussions/884) for examples.

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

The `credentials_script` feature allows you to fetch authentication information via an external script. This can be configured in the `Pollers` section of your `harvest.yml` file, as shown in the example below.

At runtime, Harvest will invoke the script specified in the `credentials_script` `path` section. Harvest will call the script with one or two arguments depending on how your poller is configured in the `harvest.yml` file. The script will be called like this: `./script $addr` or `./script $addr $username`.

- The first argument `$addr` is the address of the cluster taken from the `addr` field under the `Pollers` section of your `harvest.yml` file.
- The second argument `$username` is the username for the cluster taken from the `username` field under the `Pollers` section of your `harvest.yml` file. If your `harvest.yml` does not include a username, nothing will be passed.

The script should  communicate the credentials to Harvest by writing the response to its standard output (stdout). Harvest supports two output formats from the script:

1. **YAML format:** If the script outputs a YAML object with `username` and `password` keys, Harvest will use both the `username` and `password` from the output. For example, if the script writes the following, Harvest will use `myuser` and `mypassword` for the poller's credentials.
   ```yaml
   username: myuser
   password: mypassword
   ```
   If only the `password` is provided, Harvest will use the `username` from the `harvest.yml` file, if available. If your username or password contains spaces, `#`, or other characters with special meaning in YAML, make sure you quote the value like so:
   `password: "my password with spaces"`

2. **Plain text format:** If the script outputs plain text, Harvest will use the output as the password. The `username` will be taken from the `harvest.yml` file, if available.  For example, if the script writes the following to its stdout, Harvest will use the username defined in that poller's section of the `harvest.yml` and `mypassword` for the poller's credentials.
   ```
   mypassword
   ```

If the script doesn't finish within the specified `timeout`, Harvest will terminate the script and any spawned processes.

Credential scripts are defined under the `credentials_script` section within `Pollers` in your `harvest.yml`. Below are the options for the `credentials_script` section:

| parameter | type                    | description                                                                                                                                                                  | default |
|-----------|-------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| path      | string                  | Absolute path to the script that takes two arguments: `addr` and `username`, in that order.                                                                                  |         |
| schedule  | go duration or `always` | Schedule for calling the authentication script. If set to `always`, the script is called every time a password is requested; otherwise, the previously cached value is used. | 24h     |
| timeout   | go duration             | Maximum time Harvest will wait for the script to finish before terminating it and its descendants.                                                                           | 10s     |

### Example

Here is an example of how to configure the `credentials_script` in the `harvest.yml` file:

```yaml
Pollers:
  ontap1:
    datacenter: rtp
    addr: 10.1.1.1
    username: admin # Optional: if not provided, the script must return the username
    collectors:
      - Rest
      - RestPerf
    credentials_script:
      path: ./get_credentials
      schedule: 3h
      timeout: 10s
```

In this example, the `get_credentials` script should be located in the same directory as the `harvest.yml` file and should be executable. It should output the credentials in either YAML or plain text format. Here are two example scripts:

`get_credentials` that outputs YAML:
```bash
#!/bin/bash
cat << EOF
username: myuser
password: mypassword
EOF
```

`get_credentials` that outputs only the password in plain text:
```bash
#!/bin/bash
echo "mypassword"
```

### Troubleshooting

* Make sure your script is executable 
* Ensure the user/group that executes your poller also has read and execute permissions on the script. 
  `su` as the user/group that runs Harvest and make sure you can execute the script too. 
