# Change Log

[Releases](https://github.com/NetApp/harvest/releases)

## 21.05.0 / 2021-05-20
**IMPORTANT** v21.05 increased Harvest's out-of-the-box security posture - self-signed certificates are rejected by default. You have two options:
 
 1. [Setup client certificates for each cluster](https://github.com/NetApp/harvest-private/blob/main/cmd/collectors/zapi/README.md)
 2. Disable the TLS check in Harvest. To disable, you need to edit `harvest.yml` and add `use_insecure_tls=true` to each poller or add it to the `Defaults` section. Doing so tells Harvest to ignore invalid TLS certificates.

**IMPORTANT** RPM and Debian packages will be deprecated in the future, replaced with Docker and native packages.

 **IMPORTANT** Harvest 1.6 is end of support. We recommend you upgrade to Harvest 21.05 to take advantage of the improvements.

Changes since rc2
### Fixes
- Log mistyped exporter names and continue, instead of stopping
- `harvest grafana` should work with custom `harvest.yml` files passed via `--config`
- Harvest will try harder to stop pollers when they're stuck
- Add Grafana version check to ensure Harvest can talk to a supported version of Grafana
- Normalize rate counter calculations - improves latency values
- Workload latency calculations improved by using related objects operations
- Make cli flags consistent across programs and subcommands
- Reduce aggressive logging; if first object has fatal errors, abort to avoid repetitive errors
- Throw error when use_insecure_tls is false and there are no certificates setup for the cluster
- Harvest status fails to print port number after restart
- RPM install should create required directories
- Collector now warns if it falls behind schedule
- package.sh fails without internet connection

### Enhancements
- Add new exporter for InfluxDB
- Add native install package
- Add ARCHITECTURE.md and improve overall documentation
- Use systemd harvest.service on RPM and Debian installs to manage Harvest
- Add runtime profiling support - off by default, enabled with `--profiling` flag. See `harvest start --help` for details
- Document how to use ONTAP client certificates for password-less polling
- Add per-poller Prometheus end-point support with `promPort`
- The release, commit and build date information are baked into the release executables
- You can pick a subset of pollers to manage by passing the name of the poller to harvest. e.g. `harvest start|stop|restart POLLERS`

## rc2

### Fixes
- RPM package should create Harvest user and group
- Fixed many bugs (and possibly created new ones)
- Don't restart pollers without stopping them first
- Improve XML parse time by changing ZAPI collectors to request less data from ONTAP
- Fixed race condition in the Prometheus exporter (thanks to Yann Bizeul)
- Fixed non-portable function calls that would cause Harvest to crash on ARM architectures

### Enhancements
- Add Debian package
- Improved metric architecture, eliminated race conditions in matrix data structure. This paves the way for other developers to create custom collectors
    - Matrix can be manipulated by collectors and plugins safely
    - Size of the matrix can be changed dynamically
    - Label data is collected (in early versions, at least one numeric metric was required)
- [New plugin architecture](cmd/poller/plugin/README.md) - creating new plugins is easier and existing plugins made more generic
    - You can use built-in plugins by adding rules to a collector's template. RC2 includes two built-in plugins:
      - **Aggregator**: Aggregates metrics for a given label, e.g. volume data can be used to create an aggregation at the node or SVM-level
       - **LabelAgent**: Defines rules for rewriting instance labels, creating new labels or create ignore-lists based on regular expressions  
## rc1

 **IMPORTANT** Harvest has been rewritten in Go

 **IMPORTANT** Harvest no longer gathers data from AIQ Unified Manager. An install of AIQ.UM is not required.

### Fixes

### Enhancements
- RPM installation now conforms to a more standard Linux filesystem layout - needed to support container deployments
- Unified Grafana dashboards for cDOT and 7-mode systems
- Early release of `harvest config` tool to help you edit your `harvest.yml` file
- Add ZAPI collectors for performance, capacity and hardware metrics - gather directly from ONTAP
- Add new exporter for Prometheus
- Add plugin for Prometheus alert manager integration