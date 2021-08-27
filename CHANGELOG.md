# Change Log

[Releases](https://github.com/NetApp/harvest/releases)

## 21.08.0 / 2021-08-31
<!-- git log --no-merges --cherry-pick --right-only --oneline release/21.05.4...release/21.08.0 -->

This major release introduces a Docker workflow that makes it a breeze to
standup Grafana, Prometheus, and Harvest with auto-provisioned dashboards. There
are six new dashboards, example Prometheus alerts, and a bunch of fixes
detailed below. We haven't forgotten about our 7-mode customers either and have
a number of improvements in 7-mode dashboards with more to come.

This release Harvest also sports the most external contributions to date.
:metal: Thanks!

With 284 commits since 21.05, there is a lot to summarize! Make sure you check out the full list of enhancements and improvements in the [CHANGELOG.md](CHANGELOG.md) since 21.05.

**IMPORTANT** Harvest relies on the autosupport sidecar binary to periodically
send usage and support telemetry data to NetApp by default. Usage of the
harvest-autosupport binary falls under the NetApp [EULA](autosupport/NetApp-EULA.md).
Automatic sending of this data can be disabled with the `autosupport_disabled`
option. See [Autosupport](autosupport) for details.

**IMPORTANT** RPM and Debian packages will be deprecated in the future, replaced
with Docker and native binaries. See
[#330](https://github.com/NetApp/harvest/issues/330) for details and tell us
what you think. Several of you have already weighed-in. Thanks! If you haven't,
please do.

**Known Issues**

We've improved several of the 7-mode dashboards this release, but there are still a number of gaps with 7-mode dashboards when compared to c-mode. We will address these in a point release by splitting the c-mode and 7-mode dashboards. See [#423](https://github.com/NetApp/harvest/issues/423) for details.

On RHEL and Debian, the example Unix collector does not work at the moment due to the `harvest` user lacking permissions to read the `/proc` filesystem. See [#249](https://github.com/NetApp/harvest/issues/249) for details.

### Enhancements

- Make it easy to install Grafana, Prometheus, and Harvest with Docker Compose and auto-provisioned dashboards. [#349](https://github.com/NetApp/harvest/pull/349)

- Lun, Volume Details, Node Details, and Network Details dashboards added to Harvest. Thanks to Jeff Asher for contributing four solid dashboards. [#458](https://github.com/NetApp/harvest/pull/458)
      
- Disk dashboard added to Harvest with disk type, status, uptime, and aggregate information. Thanks to @faguayot, @bengoldenberg, and @talshechanovitz for helping with this feature [#348](https://github.com/NetApp/harvest/issues/348) [#375](https://github.com/NetApp/harvest/pull/375) [#367](https://github.com/NetApp/harvest/pull/367) [#361](https://github.com/NetApp/harvest/pull/361)
  
- New SVM dashboard with NFS v3 and v4 frontend drill-downs. Thanks to @burkl for contributing these. :tada: [#344](https://github.com/NetApp/harvest/issues/344)

- Harvest templates should be extendible without modifying the originals. Thanks to @madhusudhanarya and @albinpopote for reporting. [#394](https://github.com/NetApp/harvest/issues/394) [#396](https://github.com/NetApp/harvest/issues/396) [#391](https://github.com/NetApp/harvest/pull/391)

- Sort all variables in Harvest dashboards in ascending order. Thanks to @florianmulatz for raising [#350](https://github.com/NetApp/harvest/issues/350)

- Harvest should include example Prometheus alerting rules [#414](https://github.com/NetApp/harvest/pull/414)

- Improved documentation on how to send new ZAPIs and modify existing ZAPI templates. Thanks to @albinpopote for reporting. [#397](https://github.com/NetApp/harvest/issues/397) 
 
- Improve Harvest ZAPI template selection when monitoring a broader set of ONTAP clusters including 7-mode and 9.10.X [#407](https://github.com/NetApp/harvest/pull/407)

- Collectors should log their full ZAPI request/response(s) when their poller includes a `log` section [#382](https://github.com/NetApp/harvest/pull/382)

- Harvest should load config information from the `HARVEST_CONF` environment variable when set. Thanks to @ybizeul for reporting. [#368](https://github.com/NetApp/harvest/issues/368)

- Document how to delete time series data from Prometheus [#393](https://github.com/NetApp/harvest/pull/393)
  
- Harvest ZAPI tool supports printing results in XML and colors. This makes it easier to post-process responses in downstream pipelines [#353](https://github.com/NetApp/harvest/pull/353)
  
- Harvest `version` should check for a new release and display it when available [#323](https://github.com/NetApp/harvest/issues/323)

- Document how client authentication works and how to troubleshoot [#325](https://github.com/NetApp/harvest/pull/325)  

### Fixes

- ZAPI collector should recover after losing connection with ONTAP cluster for several hours. Thanks to @hashi825 for reporting this and helping us track it down [#356](https://github.com/NetApp/harvest/issues/356)

- Lots of fixes for 7-mode dashboards and data collection. Thanks to @madhusudhanarya and @ybizeul for reporting. There's still more work to do for 7-mode, but we understand some of our customers rely on Harvest to help them monitor these legacy systems. [#383](https://github.com/NetApp/harvest/issues/383) [#441](https://github.com/NetApp/harvest/issues/441) [#423](https://github.com/NetApp/harvest/issues/423) [#415](https://github.com/NetApp/harvest/issues/415) [#376](https://github.com/NetApp/harvest/issues/376)

- Aggregate dashboard "space used column" should use current fill grade. Thanks to @florianmulatz for reporting. [#351](https://github.com/NetApp/harvest/issues/351)

- When building RPMs don't compile Harvest Python test code. Thanks to @madhusudhanarya for reporting. [#385](https://github.com/NetApp/harvest/issues/385)

- Harvest should collect include NVMe and fiber chanel port counters. Thanks to @jgasher for submitting these. [#363](https://github.com/NetApp/harvest/issues/363)
- 
- Harvest panics when port_range is used in the Prometheus exporter and address is missing. Thanks to @ybizeul for reporting. [#357](https://github.com/NetApp/harvest/issues/357)

- Network dashboard fiber channel ports (FCP) should report read and write throughput [#445](https://github.com/NetApp/harvest/pull/445)

- Aggregate dashboard panel titles should match the information being displayed [#133](https://github.com/NetApp/harvest/issues/133)

- Harvest should handle ZAPIs that include signed integers. Most ZAPIs use unsigned integers, but a few return signed ones [#387](https://github.com/NetApp/harvest/issues/387)

---

## 21.05.4 / 2021-07-22
<!-- git log --no-merges --cherry-pick --right-only --oneline release/21.05.3...release/21.05.4 -->

This release introduces Qtree protocol collection, improved Docker and client authentication documentation, publishing to Docker Hub, and a new plugin that helps build richer dashboards, as well as a couple of important fixes for collector panics.

**IMPORTANT** RPM and Debian packages will be deprecated in the future, replaced with Docker and native binaries. See [#330](https://github.com/NetApp/harvest/issues/330) for details and tell us what you think.

**Known Issues**

On RHEL and Debian, the example Unix collector does not work at the moment due to the `harvest` user lacking permissions to read the `/proc` filesystem. See [#249](https://github.com/NetApp/harvest/issues/249) for details.

### Enhancements

- Harvest collects Qtree protocol ops [#298](https://github.com/NetApp/harvest/pull/298). Thanks to Martin Möbius for contributing

- Harvest Grafana tool (optionally) adds a user-specified prefix to all Dashboard metrics during import. See  `harvest grafana --help` [#87](https://github.com/NetApp/harvest/issues/87)
  
- Harvest is taking its first steps to talk REST: query ONTAP, show Swagger API, model, and definitions [#292](https://github.com/NetApp/harvest/pull/292)
 
- Tagged releases of Harvest are published to [Docker Hub](https://hub.docker.com/r/rahulguptajss/harvest) 
  
- Harvest honors Go's http(s) environment variable proxy information. See https://pkg.go.dev/net/http#ProxyFromEnvironment for details [#252](https://github.com/NetApp/harvest/pull/252)

- New plugin [value_to_num](https://github.com/NetApp/harvest/blob/main/cmd/poller/plugin/README.md#value_to_num) helps map labels to numeric values for Grafana dashboards. Current dashboards updated to use this plugin [#319](https://github.com/NetApp/harvest/pull/319)

- `harvest.yml` supports YAML flow style. E.g. `collectors: [Zapi]` [#260](https://github.com/NetApp/harvest/pull/260)

- New Simple collector that runs on Macos and Unix [#270](https://github.com/NetApp/harvest/pull/270)  

- Improve client certificate authentication [documentation](https://github.com/NetApp/harvest/issues/314#issuecomment-882120238)

- Improve Docker deployment documentation [4019308](https://github.com/NetApp/harvest/tree/main/docker/onePollerPerContainer)

### Fixes

- Harvest collector should not panic when resources are deleted from ONTAP [#174](https://github.com/NetApp/harvest/issues/174) and [#302](https://github.com/NetApp/harvest/issues/302). Thanks to @hashi825 and @mamoep for providing steps to reproduce

- Shelf metrics should report on op-status for components. Thanks to @hashi825 for working with us on this fix and dashboard improvements [#262](https://github.com/NetApp/harvest/issues/262)
  
- Harvest should not panic when InfluxDB is the only exporter [#286](https://github.com/NetApp/harvest/issues/284)

- Volume dashboard space-used column should display with percentage filled. Thanks to @florianmulatz for reporting and suggesting a fix [#303](https://github.com/NetApp/harvest/issues/303)

- Certificate authentication should honor path in `harvest.yml` [#318](https://github.com/NetApp/harvest/pull/318)
  
- Harvest should not kill processes with `poller` in their arguments [#328](https://github.com/NetApp/harvest/issues/328)
  
- Harvest ZAPI command line tool should limit perf-object-get-iter to subset of counters when using `--counter` [#299](https://github.com/NetApp/harvest/pull/299) 

---

## 21.05.3 / 2021-06-28

This release introduces a significantly simplified way to connect Harvest and Prometheus, improves Harvest builds times by 7x, reduces executable sizes by 3x, enables cross compiling support, and includes several Grafana dashboard and collector fixes.

**Known Issues**

On RHEL and Debian, the example Unix collector does not work at the moment due to the `harvest` user lacking permissions to read the `/proc` filesystem. See [#249](https://github.com/NetApp/harvest/issues/249) for details.

### Enhancements
- Create Prometheus port range exporter that allows you to connect multiple pollers to Prometheus without needing to specify a port-per-poller. This makes it much easier to connect Prometheus and Harvest; especially helpful when you're monitoring many clusters [#172](https://github.com/NetApp/harvest/issues/172)

- Improve Harvest build times by 7x and reduce executable sizes by 3x [#100](https://github.com/NetApp/harvest/issues/100)
 
- Improve containerization with the addition of a poller-per-container Dockerfile. Create a new subcommand `harvest generate docker` which generates a `docker-compose.yml` file for all pollers defined in your config

- Improve systemd integration by using instantiated units for each poller and a harvest target to tie them together. Create a new subcommand `harvest generate systemd` which generates a Harvest systemd target for all pollers defined in your config [#systemd](https://github.com/NetApp/harvest/tree/main/service/contrib)
  
- Harvest doctor checks that all Prometheus exporters specify a unique port [#118](https://github.com/NetApp/harvest/issues/118)
  
- Harvest doctor warns when an unknown exporter type is specified (likely a spelling error) [#118](https://github.com/NetApp/harvest/issues/118)
  
- Add Harvest [CUE](https://cuelang.org/) validation and type-checking [#208](https://github.com/NetApp/harvest/pull/208)
  
- `bin/zapi` uses the `--config` command line option to read the harvest config file. This brings this tool inline with other Harvest tools. This makes it easier to switch between multiple sets of harvest.yml files.
  
- Harvest no longer writes pidfiles; simplifying management code and install [#159](https://github.com/NetApp/harvest/pull/159)

### Fixes
- Ensure that the Prometheus exporter does not create duplicate labels [#132](https://github.com/NetApp/harvest/issues/132)

- Ensure that the Prometheus exporter includes `HELP` and `TYPE` metatags when requested. Some tools require these [#104](https://github.com/NetApp/harvest/issues/104)
      
- Disk status should return zero for a failed disk and one for a healthy disk. Thanks to @hashi825 for reporting and fixing [#182](https://github.com/NetApp/harvest/issues/182)

- Lun info should be collected by Harvest. Thanks to @hashi825 for reporting and fixing [#230](https://github.com/NetApp/harvest/issues/230)

- Grafana dashboard units, typo, and filtering fixes. Thanks to @mamoep, @matejzero, and @florianmulatz for reporting these :tada: [#184](https://github.com/NetApp/harvest/issues/184) [#186](https://github.com/NetApp/harvest/issues/186) [#190](https://github.com/NetApp/harvest/issues/190) [#192](https://github.com/NetApp/harvest/issues/192) [#195](https://github.com/NetApp/harvest/issues/195) [#202](https://github.com/NetApp/harvest/issues/202)

- Unix collector should not panic when harvest.yml is changed [#160](https://github.com/NetApp/harvest/issues/160)

- Reduce log noise about poller lagging behind by few milliseconds. Thanks @hashi825 [#214](https://github.com/NetApp/harvest/issues/214)

- Don't assume debug when foregrounding the poller process. Thanks to @florianmulatz for reporting. [#246](https://github.com/NetApp/harvest/issues/246)

- Improve Docker all-in-one-container argument handling and simplify building in air gapped environments. Thanks to @optiz0r for reporting these issues and creating fixes. [#166](https://github.com/NetApp/harvest/pull/166) [#167](https://github.com/NetApp/harvest/pull/167) [#168](https://github.com/NetApp/harvest/pull/168)

---
## 21.05.2 / 2021-06-14

This release adds support for user-defined URLs for InfluxDB exporter, a new command to validate your `harvest.yml` file, improved logging, panic handling, and collector documentation. We also enabled GitHub security code scanning for the Harvest repo to catch issues sooner. These scans happen on every push.

There are also several quality-of-life bug fixes listed below.

### Fixes
- Handle special characters in cluster credentials [#79](https://github.com/NetApp/harvest/pull/79)
- TLS server verification works with basic auth [#51](https://github.com/NetApp/harvest/issues/51)
- Collect metrics from all disk shelves instead of one [#75](https://github.com/NetApp/harvest/issues/75)
- Disk serial number and is-failed are missing from cdot query [#60](https://github.com/NetApp/harvest/issues/60)
- Ensure collectors and pollers recover from panics [#105](https://github.com/NetApp/harvest/issues/105)
- Cluster status is initially reported, but then stops being reported [#66](https://github.com/NetApp/harvest/issues/66)
- Performance metrics don't display volume names [#40](https://github.com/NetApp/harvest/issues/40)
- Allow insecure Grafana TLS connections `--insecure` and honor requested transport. See `harvest grafana --help` for details [#111](https://github.com/NetApp/harvest/issues/111)
- Prometheus dashboards don't load when `exemplar` is true. Thanks to @sevenval-admins, @florianmulatz, and @unbreakabl3 for their help tracking this down and suggesting a fix. [#96](https://github.com/NetApp/harvest/issues/96)
- `harvest stop` does not stop pollers that have been renamed [#20](https://github.com/NetApp/harvest/issues/20)
- Harvest stops working after reboot on rpm/deb [#50](https://github.com/NetApp/harvest/issues/50)
- `harvest start` shall start as harvest user in rpm/deb [#129](https://github.com/NetApp/harvest/issues/129)
- `harvest start` detects stale pidfiles and makes start idempotent [#123](https://github.com/NetApp/harvest/issues/123)
- Don't include unknown metrics when talking with older versions of ONTAP [#116](https://github.com/NetApp/harvest/issues/116)
### Enhancements
- InfluxDB exporter supports [user-defined URLs](https://github.com/NetApp/harvest/blob/main/cmd/exporters/influxdb/README.md#parameters)
- Add workload counters to ZapiPerf [#9](https://github.com/NetApp/harvest/issues/9)
- Add new command to validate `harvest.yml` file and optionally redact sensitive information [#16](https://github.com/NetApp/harvest/issues/16) e.g. `harvest doctor --config ./harvest.yml`
- Improve documentation for [Unix](https://github.com/NetApp/harvest/tree/main/cmd/collectors/unix), [Zapi](https://github.com/NetApp/harvest/tree/main/cmd/collectors/zapi), and [ZapiPerf](https://github.com/NetApp/harvest/tree/main/cmd/collectors/zapiperf) collectors
- Add Zerolog framework for structured logging [#61](https://github.com/NetApp/harvest/issues/61)
- Vendor 3rd party code to increase reliability and make it easier to build in air-gapped environments [#26](https://github.com/NetApp/harvest/pull/26)
- Make contributing easier with a digital CCLA instead of 1970's era PDF :)
- Enable GitHub security code scanning
- InfluxDB exporter provides the option to pass the URL end-point unchanged. Thanks to @steverweber for their suggestion and validation. [#63](https://github.com/NetApp/harvest/issues/63)

---
## 21.05.1 / 2021-05-20

Announcing the release of Harvest2. With this release the core of Harvest has been completely rewritten in Go. Harvest2 is a replacement for the older versions of Harvest 1.6 and below.

If you're using one of the Harvest 2.x release candidates, you can do a direct upgrade.

Going forward Harvest2 will follow a `year.month.fix` release naming convention with the first release being 21.05.0. See [SUPPORT.md](SUPPORT.md) for details.

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
- Version flag is missing new line on some shells [#4](https://github.com/NetApp/harvest/issues/4)
- Poller should not ignore --config [#28](https://github.com/NetApp/harvest/issues/28)
- Handle special characters in cluster credentials [#79](https://github.com/NetApp/harvest/pull/79)
- TLS server verification works with basic auth [#51](https://github.com/NetApp/harvest/issues/51)
- Collect metrics from all disk shelves instead of one [#75](https://github.com/NetApp/harvest/issues/75)
- Disk serial number and is-failed are missing from cdot query [#60](https://github.com/NetApp/harvest/issues/60)
- Ensure collectors and pollers recover from panics [#105](https://github.com/NetApp/harvest/issues/105)
- Cluster status is initially reported, but then stops being reported [#66](https://github.com/NetApp/harvest/issues/66)
- Performance metrics don't display volume names [#40](https://github.com/NetApp/harvest/issues/40)
- Allow insecure Grafana TLS connections `--insecure` and honor requested transport. See `harvest grafana --help` for details [#111](https://github.com/NetApp/harvest/issues/111)
- Prometheus dashboards don't load when `exemplar` is true. Thanks to @sevenval-admins, @florianmulatz, and @unbreakabl3 for their help tracking this down and suggesting a fix. [#96](https://github.com/NetApp/harvest/issues/96)

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
- InfluxDB exporter supports [user-defined URLs](https://github.com/NetApp/harvest/blob/main/cmd/exporters/influxdb/README.md#parameters)
- Add workload counters to ZapiPerf [#9](https://github.com/NetApp/harvest/issues/9)
- Add new command to validate `harvest.yml` file and optionally redact sensitive information [#16](https://github.com/NetApp/harvest/issues/16) e.g. `harvest doctor --config ./harvest.yml`
- Improve documentation for [Unix](https://github.com/NetApp/harvest/tree/main/cmd/collectors/unix), [Zapi](https://github.com/NetApp/harvest/tree/main/cmd/collectors/zapi), and [ZapiPerf](https://github.com/NetApp/harvest/tree/main/cmd/collectors/zapiperf) collectors
- Add Zerolog framework for structured logging [#61](https://github.com/NetApp/harvest/issues/61)
- Vendor 3rd party code to increase reliability and make it easier to build in air-gapped environments [#26](https://github.com/NetApp/harvest/pull/26)
- Make contributing easier with a digital CCLA instead of 1970's era PDF :)
- Enable GitHub security code scanning

---
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

---
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