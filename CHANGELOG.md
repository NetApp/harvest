# Change Log

[Releases](https://github.com/NetApp/harvest/releases)

## 22.08.0 / 2022-08-19
<!-- git log --no-decorate --no-merges --cherry-pick --right-only --oneline origin/release/22.05.0...origin/release/22.08.0 -->

:rocket: Highlights of this major release include:

- :sparkler: an ONTAP event management system (EMS) events collector

- Two new dashboards added in this release
  - Headroom dashboard
  - Quota dashboard

- We've made lots of improvements to the REST Perf collector. The REST Perf collector should be considered early-access as we continue to improve it. This feature requires ONTAP versions 9.11.1 and higher.

- New [`max` plugin](https://github.com/NetApp/harvest/blob/main/cmd/poller/plugin/README.md#max) that creates new metrics from the maximum of existing metrics by label.

- New [`compute_metric` plugin](https://github.com/NetApp/harvest/blob/main/cmd/poller/plugin/README.md#compute_metric) that creates new metrics by combining existing metrics with mathematical operations.

- 48 feature, 45 bug fixes, and 11 documentation commits this release

**IMPORTANT** :bangbang: NetApp is moving their communities from Slack to [NetApp's Discord](https://discord.gg/ZmmWPHTBHw) with a plan to lock the Slack channel at the end of August. Please join us on [Discord](https://discordapp.com/channels/855068651522490400/1001963189124206732)!

**IMPORTANT** :bangbang: Prometheus version `2.26` or higher is required for the EMS Collector. 

**IMPORTANT** :bangbang: After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the `bin/harvest/grafana import` CLI or from the Grafana UI.

**Known Issues**

Podman is unable to pull from NetApp's container registry `cr.netapp.io`. Until [issue](https://github.com/containers/podman/issues/15187) is resolved, Podman users can pull from a separate proxy like this `podman pull netappdownloads.jfrog.io/oss-docker-harvest-production/harvest:latest`.

**IMPORTANT** 7-mode filers that are not on the latest release of ONTAP may experience TLS connection issues with errors like `tls: server selected unsupported protocol version 301` This is caused by a change in Go 1.18. The [default for TLS client connections was changed to TLS 1.2](https://tip.golang.org/doc/go1.18#tls10) in Go 1.18. Please upgrade your 7-mode filers (recommended) or set `tls_min_version: tls10` in your `harvest.yml` [poller section](https://github.com/NetApp/harvest/tree/release/22.05.0#pollers). See #1007 for more details.

The Unix collector is unable to monitor pollers running in containers. See [#249](https://github.com/NetApp/harvest/issues/249) for details.

### Enhancements

- :sparkler: Harvest adds an [ONTAP event management system (EMS) events](https://github.com/NetApp/harvest/blob/main/cmd/collectors/ems/README.md) collector in this release. It collects ONTAP events, exports them to Prometheus, and provides integration with Prometheus AlertManager.

- New Harvest Headroom dashboard. [#1039](https://github.com/NetApp/harvest/issues/1039) Thanks to @faguayot for reporting.
 
- New Quota dashboard. [#1111](https://github.com/NetApp/harvest/issues/1111) Thanks to @ev1963 for raising this feature request.

- We've made lots of improvements to the REST Perf collector and filled several gaps in this release. [#881](https://github.com/NetApp/harvest/issues/881)

- Harvest Power dashboard should include `Min Ambient Temp` and `Min Temp`. Thanks to Papadopoulos Anastasios for reporting. 

- Harvest Disk dashboard should include the `Back-to-back CP Count` and `Write Latency` metrics. [#1040](https://github.com/NetApp/harvest/issues/1040) Thanks to @faguayot for reporting.

- Rest templates should be disabled [by default](https://github.com/NetApp/harvest/blob/main/conf/rest/default.yaml) until ONTAP removes ZAPI support. That way, Harvest does not double collect and store metrics.

- Harvest dashboards name prefix should be `ONTAP:` instead of `NetApp Detail:`. [#1080](https://github.com/NetApp/harvest/pull/1080). Thanks to `Martin Möbius` for reporting.

- Harvest Qtree dashboard should show `Total Qtree IOPs` and `Internal IOPs` panels and `Qtree` filter. [#1079](https://github.com/NetApp/harvest/issues/1079) Thanks to @mamoep for reporting.

- Harvest Cluster dashboard should show `SVM Performance` panel. [#1117](https://github.com/NetApp/harvest/issues/1117) Thanks to @Falcon667 for reporting.

- Combine `SnapMirror` and `Data Protection` dashboards. [#1082](https://github.com/NetApp/harvest/issues/1082). Thanks to `Martin Möbius` for reporting.

- `vscan` performance object should be enabled by default. [#1182](https://github.com/NetApp/harvest/pull/1182) Thanks to `Gabriel Conne` for reporting on Slack.

- Lun and Volume dashboard should use `topk` range. [#1184](https://github.com/NetApp/harvest/pull/1184) Thanks to `Papadopoulos Anastasios` for reporting on Slack. These changes make these dashboards more consistent with Harvest 1.6.

- New [MetricAgent](https://github.com/NetApp/harvest/blob/main/cmd/poller/plugin/README.md#metricagent) plugin. It is used to manipulate `metrics` based on a set of rules.

- New [Max](https://github.com/NetApp/harvest/blob/main/cmd/poller/plugin/README.md#max) plugin. It creates a new collection of metrics by calculating max of metric values from an existing matrix for a given label.

- `bin/zapi` should support querying multiple performance counters. [#1167](https://github.com/NetApp/harvest/pull/1167)

- Harvest REST private CLI should include filter support

- Harvest should support request/response logging in Rest/RestPerf Collector.

- Harvest maximum log file size is reduced from 10mb to 5mb. The maximum number of log files are reduced from 10 to 5.

- Harvest should consolidate log messages and reduce noise.

### Fixes

- Missing Ambient Temperature for AFF900 in Power Dashboard. [#1173](https://github.com/NetApp/harvest/issues/1173) Thanks to @iStep2Step for reporting.

- Flexgroup latency should match the values reported by ONTAP CLI. [#1060](https://github.com/NetApp/harvest/pull/1060) Thanks to @josepaulog for reporting.

- Perf Zapi Volume label should match Zapi Volume label. The label `type` was changed to `style` for Perf ZAPI Volume. [#1055](https://github.com/NetApp/harvest/pull/1055) Thanks to Papadopoulos Anastasios for reporting.

- Zapi:SecurityCert should handle certificates per SVM instead of reporting `duplicate instance key` errors. [#1075](https://github.com/NetApp/harvest/issues/1075) Thanks to @mamoep for reporting.

- Zapi:SecurityAccount should handle per switch SNMP users instead of reporting `duplicate instance key` errors. [#1088](https://github.com/NetApp/harvest/issues/1088) Thanks to @mamoep for reporting.

- Wrong throughput units in Disk dashboard. [#1091](https://github.com/NetApp/harvest/issues/1091) Thanks to @Falcon667 for reporting.

- Qtree Dashboard shows no data when SVM/Volume are selected from dropdown. [#1099](https://github.com/NetApp/harvest/issues/1099) Thanks to `Papadopoulos Anastasios` for reporting.

- `Virus Scan connections Active panel` in SVM dashboard shows decimal places in Y axis. [#1101](https://github.com/NetApp/harvest/issues/1101) Thanks to `Rene Meier` for reporting.

- Add `Disk Utilization per Aggregate` description in Disk Dashboard. [#1193](https://github.com/NetApp/harvest/issues/1193) Thanks to @faguayot for reporting.

- Prometheus exporter should escape label_value. [#1128](https://github.com/NetApp/harvest/issues/1128) Thanks to @vavdoshka for reporting.

- Grafana import dashboard fails if anonymous access is enabled. [@1132](https://github.com/NetApp/harvest/issues/1132) Thanks @iStep2Step for reporting.

- Improve color consistency and hover information on Compliance/Data Protection dashboards. [#1083](https://github.com/NetApp/harvest/issues/1083)  Thanks to `Rene Meier` for reporting.

- Compliance & Security Dashboards the text is unreadable with Grafana light theme. [#1078](https://github.com/NetApp/harvest/issues/1078) Thanks to @mamoep for reporting.

- InfluxDB exporter should not require bucket, org, port, or precision fields when using url. [#1155](https://github.com/NetApp/harvest/issues/1155) Thanks to `li fi` for reporting.

- `Node CPU Busy` and `Disk Utilization` should match the same metrics reported by ONTAP `sysstat -m` CLI. [#1152](https://github.com/NetApp/harvest/issues/1152) Thanks to `Papadopoulos Anastasios` for reporting.

- Harvest should detect counter overflow and report it as `0`. [#762] Thanks to @rodenj1 for reporting.

- Zerolog console logger fails to log stack traces. [#1044](https://github.com/NetApp/harvest/pull/1044)

---

## 22.05.0 / 2022-05-11
<!-- git log --no-decorate --no-merges --cherry-pick --right-only --oneline origin/release/22.02.0...origin/release/22.05.0 -->

:rocket: Highlights of this major release include:

- Early access to ONTAP REST perf collector from ONTAP 9.11.1GA+

- :hourglass: New Container Registry - Several of you have mentioned that you are being rate-limited when pulling Harvest Docker images from DockerHub. To alleviate this problem, we're publishing Harvest images to NetApp's container registry (cr.netapp.io). Going forward, we'll publish images to both DockerHub and cr.netapp.io. More information in the [FAQ](https://github.com/NetApp/harvest/wiki/FAQ#where-are-harvest-container-images-published). No action is required unless you want to switch from DockerHub to cr.netapp.io. If so, the FAQ has you covered.

- Five new dashboards added in this release
  - Power dashboard
  - Compliance dashboard
  - Security dashboard
  - Qtree dashboard
  - NFSv4 Store Pool dashboard (disabled by default)

- New `value_to_num_regex` plugin allows you to map all matching expressions to 1 and non-matching ones to 0.

- Harvest pollers can optionally [read credentials](https://github.com/NetApp/harvest/discussions/884) from a mounted volume or file. This enables [Hashicorp Vault](https://www.vaultproject.io/) support and works especially well with [Vault agent](https://www.vaultproject.io/docs/agent) 

- `bin/grafana import` provides a `--multi` flag that rewrites dashboards to include multi-select dropdowns for each variable at the top of the dashboard

- The `conf/rest` collector templates received a lot of attentions this release. All known gaps between the ZAPI and REST collector have been filled and there is full parity between the two from ONTAP 9.11+. :metal:

- 24 bug fixes, 47 feature, and 5 documentation commits this release

**IMPORTANT** :bangbang: After upgrade, don't forget to re-import your dashboards so you get all the new enhancements and fixes. You can import via `bin/harvest/grafana import` cli or from the Grafana UI.

**IMPORTANT** The `conf/zapiperf/cdot/9.8.0/object_store_client_op.yaml` ZapiPerf template is being deprecated in this release and will be removed in the next release of Harvest. No dashboards use the counters defined in this template and all counters are being deprecated by ONTAP. If you are using these counters, please create your own copy of the template.

**Known Issues**

**IMPORTANT** 7-mode filers that are not on the latest release of ONTAP may experience TLS connection issues with errors like `tls: server selected unsupported protocol version 301` This is caused by a change in Go 1.18. The [default for TLS client connections was changed to TLS 1.2](https://tip.golang.org/doc/go1.18#tls10) in Go 1.18. Please upgrade your 7-mode filers (recommended) or set `tls_min_version: tls10` in your `harvest.yml` [poller section](https://github.com/NetApp/harvest/tree/release/22.05.0#pollers). See #1007 for more details.

The Unix collector is unable to monitor pollers running in containers. See [#249](https://github.com/NetApp/harvest/issues/249) for details.

### Enhancements

- Harvest should include a [Power dashboard](https://github.com/NetApp/harvest/discussions/951) that shows power consumed, temperatures and fan speeds at a node and shelf level [#932](https://github.com/NetApp/harvest/pull/932) and [#903](https://github.com/NetApp/harvest/issues/903)

- Harvest should include a Security dashboard that shows authentication methods and certificate expiration details for clusters, volume encryption and status of anti-ransomware for volumes and SVMs [#935](https://github.com/NetApp/harvest/pull/935)

- Harvest should include a Compliance dashboard that shows compliance status of clusters and SVMs along with individual compliance attributes [#935](https://github.com/NetApp/harvest/pull/935)

- SVM dashboard should show antivirus counters in the CIFS drill-down section [#913](https://github.com/NetApp/harvest/issues/913) Thanks to @burkl for reporting

- Cluster and Aggregate dashboards should show Storage Efficiency Ratio metrics [#888](https://github.com/NetApp/harvest/issues/888) Thanks to @Falcon667 for reporting

- :construction: This is another step in the ZAPI to REST road map. In earlier releases, we focused on config ZAPIs and in this release we've added early access to an ONTAP REST perf collector. :confetti_ball: The REST perf collector and thirty-nine templates included in this release, require ONTAP 9.11.1GA+ :astonished: These should be considered early access as we continue to improve them. If you try them out or have any feedback, let us know on Slack or [GitHub](https://github.com/NetApp/harvest/discussions) [#881](https://github.com/NetApp/harvest/issues/881)

- Harvest should collect NFS v4.2 counters which are new in ONTAP 9.11+ releases [#572](https://github.com/NetApp/harvest/issues/572)

- Plugin logging should include object detail [#986](https://github.com/NetApp/harvest/issues/986)

- Harvest dashboards should use Time series panels instead of Graph (old) panels [#972](https://github.com/NetApp/harvest/issues/972). Thanks to @ybizeul for raising

- New regex based plugin [value_to_num_regex](https://github.com/NetApp/harvest/blob/main/cmd/poller/plugin/README.md#value_to_num_regex) helps map labels to numeric values for Grafana dashboards.

- Harvest status should run on systems without pgrep [#937](https://github.com/NetApp/harvest/pull/937) Thanks to Dan Butler for reporting this on Slack

- When using a credentials file and the poller is not found, also consult the defaults section of the `harvest.yml` file [#936](https://github.com/NetApp/harvest/pull/936)

- Harvest should include an NFSv4 StorePool dashboard that shows NFSv4 store pool locks and allocation detail [#921](https://github.com/NetApp/harvest/pull/921) Thanks to Rusty Brown for contributing this dashboard.

- REST collector should report cpu-busytime for node [#918](https://github.com/NetApp/harvest/issues/918) Thanks to @pilot7777 for reporting this on Slack

- Harvest should include a Qtree dashboard that shows Qtree NFS/CIFS metrics [#812](https://github.com/NetApp/harvest/issues/812) Thanks to @ev1963 for reporting

- Harvest should support reading credentials from an external file or mounted volume [#905](https://github.com/NetApp/harvest/pull/905)

- Grafana dashboards should have checkbox to show multiple objects in variable drop-down. See [comment](https://github.com/NetApp/harvest/issues/815#issuecomment-1050833081) for details. [#815](https://github.com/NetApp/harvest/issues/815) [#939](https://github.com/NetApp/harvest/issues/939) Thanks to @manuelbock, @bcase303 for reporting

- Harvest should include Prometheus port (promport) to metadata metric [#878](https://github.com/NetApp/harvest/pull/878)

- Harvest should use NetApp's container registry for Docker images [#874](https://github.com/NetApp/harvest/pull/874)

- Increase ZAPI client timeout for default and volume object [#1005](https://github.com/NetApp/harvest/pull/1005)

- REST collector should support retrieving a subset of objects via template filtering support [#950](https://github.com/NetApp/harvest/pull/950)

- Harvest should support minimum TLS version config [#1007](https://github.com/NetApp/harvest/issues/1007) Thanks to @jmg011 for reporting and verifying this

### Fixes

- SVM Latency numbers differ significantly on Harvest 1.6 vs Harvest 2.0 [#1003](https://github.com/NetApp/harvest/issues/1003) See [discussion](https://github.com/NetApp/harvest/discussions/940) as well. Thanks to @jmg011 for reporting

- Harvest should include regex patterns to ignore transient volumes related to backup [#929](https://github.com/NetApp/harvest/issues/929). Not enabled by default, see `conf/zapi/cdot/9.8.0/volume.yaml` for details. Thanks to @ybizeul for reporting

- Exclude OS aggregates from capacity used graph [#327](https://github.com/NetApp/harvest/issues/327) Thanks to @matejzero for raising

- Few panels need to have instant property in Data protection dashboard [#945](https://github.com/NetApp/harvest/pull/945)

- CPU overload when there are several thousands of quotas [#733](https://github.com/NetApp/harvest/issues/733) Thanks to @Flo-Fly for reporting

- Include 7-mode CLI role commands for Harvest user [#891](https://github.com/NetApp/harvest/issues/891) Thanks to @ybizeul for reporting and providing the changes!

- Zapi Collector fails to collect data if number of records on a poller is equal to batch size [#870](https://github.com/NetApp/harvest/issues/870) Thanks to @unbreakabl3 on Slack for reporting

- Wrong object name used in `conf/zapi/cdot/9.8.0/snapshot.yaml` [#862](https://github.com/NetApp/harvest/issues/862) Thanks to @pilot7777 for reporting

- Field access-time returned by `snapshot-get-iter` should be creation-time [#861](https://github.com/NetApp/harvest/issues/861) Thanks to @pilot7777 for reporting

- Harvest panics when trying to merge empty template [#859](https://github.com/NetApp/harvest/issues/859) Thanks to @pilot7777 for raising

---

## 22.02.0 / 2022-02-15
<!-- git log --no-decorate --no-merges --cherry-pick --right-only --oneline origin/release/21.11.0...origin/release/22.02.0 -->

:boom: Highlights of this major release include:

- Continued progress on the ONTAP REST config collector. Most of the template changes are in place and we're working on closing the gaps between ZAPI and REST. We've made lots of improvements to the REST collector and included 13 REST templates in this release. The REST collector  should be considered early-access as we continue to improve it. If you try it out or have any feedback, let us know on Slack or [GitHub](https://github.com/NetApp/harvest/discussions). :book: You can find more information about when you should switch from ZAPI to REST, what versions of ONTAP are supported by Harvest's REST collector, and how to fill ONTAP gaps between REST and ZAPI documented [here](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-collector.md)
  
- Many of you asked for nightly builds. [We have them](https://github.com/NetApp/harvest/releases/tag/nightly). :confetti_ball: We're also working on publishing to multiple Docker registries since you've told us you're running into rate-limiting problems with DockerHub. We'll announce here and Slack when we have a solution in place.

- Two new Data Protection dashboards

- `bin/grafana` cli should not overwrite dashboard changes, making it simpler to import/export dashboards, and enabling round-tripping dashboards (import, export, re-import)

- New `include_contains` plugin allows you to select a subset of objects. e.g. selecting only volumes with custom-defined ONTAP metadata

- We've included more out-of-the-box [Prometheus alerts](https://github.com/NetApp/harvest/blob/main/docker/prometheus/alert_rules.yml). Keep sharing your most useful alerts!

- 7mode workflows continue to be improved :heart: Harvest now collects Qtree and Quotas counters from 7mode filers (these are already collected in cDOT)
    
- 28 bug fixes, 52 feature, and 11 documentation commits this release

**IMPORTANT** Admin node certificate file location changed. Certificate files have been consolidated into the `cert` directory. If you created self-signed admin certs, you need to move the `admin-cert.pem` and `admin-key.pem` files into the `cert` directory.

**IMPORTANT** In earlier versions of Harvest, the Qtree template exported the `vserver` metric. This counter was changed to `svm` to be consistent with other templates. If you are using the qtree `vserver` metric, you will need to update your queries to use `svm` instead.   

**IMPORTANT** :bangbang: After upgrade, don't forget to re-import your dashboards so you get all the new enhancements and fixes. 
You can import via `bin/harvest/grafana import` cli or from the Grafana UI.

**IMPORTANT** The LabelAgent `value_mapping` plugin was deprecated in the `21.11` release and removed in `22.02`.
Use LabelAgent `value_to_num` instead. See [docs](https://github.com/NetApp/harvest/blob/main/cmd/poller/plugin/README.md#value_to_num)
for details. 

**Known Issues**

The Unix collector is unable to monitor pollers running in containers. See [#249](https://github.com/NetApp/harvest/issues/249) for details.

### Enhancements

- Harvest should include a Data Protection dashboard that shows volumes protected by snapshots, which ones have exceeded their reserve copy, and which are unprotected #664

- Harvest should include a Data Protection SnapMirror dashboard that shows which volumes are protected, how they're protected, their protection relationship, along with their health and lag durations.
  
- Harvest should provide nightly builds to GitHub and DockerHub #713

- Harvest `bin/grafana` cli should not overwrite dashboard changes, making it simpler to import/export dashboards, and enabling round-tripping dashboards (import, export, re-import) #831 Thanks to @luddite516 for reporting and @florianmulatz for iterating with us on a solution

- Harvest should provide a `include_contains` label agent plugin for filtering #735 Thanks to @chadpruden for reporting

- Improve Harvest's container compatibility with K8s via kompose. [#655](https://github.com/NetApp/harvest/pull/655) See [also](https://github.com/NetApp/harvest/tree/main/docker/k8) and [discussion](https://github.com/NetApp/harvest/discussions/827)

- The ZAPI cli tool should include counter types when querying ZAPIs #663

- Harvest should include a richer set of Prometheus alerts #254 Thanks @demalik for raising

- Template plugins should run in the order they are defined and compose better. 
The output of one plugin can be fed into the input of the next one. #736 Thanks to @chadpruden for raising

- Harvest should collect Antivirus counters when ONTAP offbox vscan is configured [#346](https://github.com/NetApp/harvest/issues/346) Thanks to @burkl and @Falcon667 for reporting

- [Document](https://github.com/NetApp/harvest/tree/main/docker/containerd) how to run Harvest with `containerd` and `Rancher`
     
- Qtree counters should be collected for 7-mode filers #766 Thanks to @jmg011 for raising this issue and iterating with us on a solution

- Harvest admin node should work with pollers running in Docker compose [#678](https://github.com/NetApp/harvest/pull/678)

- [Document](https://github.com/NetApp/harvest/tree/main/docker/podman) how to run Harvest with Podman. Several RHEL customers asked about this since Podman ships as the default container runtime on RHEL8+.

- Harvest should include a Systemd service file for the HTTP service discovery admin node [#656](https://github.com/NetApp/harvest/pull/656)

- [Document](https://github.com/NetApp/harvest/blob/main/docs/TemplatesAndMetrics.md) how ZAPI collectors, templates, and exporting work together. Thanks @jmg011 and others for asking for this 

- Remove redundant dashboards (Network, Node, SVM, Volume) [#703](https://github.com/NetApp/harvest/issues/703) Thanks to @mamoep for reporting this

- Harvest `generate docker` command should support customer-supplied Prometheus and Grafana ports. [#584](https://github.com/NetApp/harvest/issues/584)

- Harvest certificate authentication should work with self-signed subject alternative name (SAN) certificates. Improve documentation on how to use [certificate authentication](https://github.com/NetApp/harvest/blob/main/docs/AuthAndPermissions.md#using-certificate-authentication). Thanks to @edd1619 for raising this issue

- Harvest's Prometheus exporter should optionally sort labels. Without sorting, [VictoriaMetrics](https://github.com/VictoriaMetrics/VictoriaMetrics/issues/1857) marks metrics stale. #756 Thanks to @mamoep for reporting and verifying

- Harvest should optionally log to a file when running in the foreground. Handy for instantiated instances running on OSes that have poor support for `jounalctl` #813 and #810 Thanks to @mamoep for reporting and verifying this works in a nightly build

- Harvest should collect workload concurrency [#714](https://github.com/NetApp/harvest/pull/714)

- Harvest certificate directory should be included in a container's volume mounts #725 

- MetroCluster dashboard should show path object metrics #746

- Harvest should collect namespace resources from ONTAP #749

- Harvest should be more resilient to cluster connectivity issues #480

- Harvest Grafana dashboard version string should match the Harvest release #631

- REST collector improvements
  - Harvest REST collector should support ONTAP private cli endpoints #766 
  
  - REST collector should support ZAPI-like object prefixing #786

  - REST collector should support computing new customer-defined metrics #780

  - REST collector should collect aggregate, qtree and quota counters #780

  - REST collector metrics should be reported in autosupport #841
    
  - REST collector should collect sensor counters #789

  - Collect network port interface information not available via ZAPI #691 Thanks to @pilot7777, @mamoep amd @wagneradrian92 for working on this with us

  - Publish REST collector [document](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-collector.md) that highlights when you should switch from ZAPI to REST, what versions of ONTAP are supported by Harvest's REST collectors and how to fill ONTAP gaps between REST and ZAPI

  - REST collector should support Qutoa, Shelf, Snapmirror, and Volume plugins #799 and #811

- Improve troubleshooting and documentation on validating certificates from macOS [#723](https://github.com/NetApp/harvest/pull/723)

- Harvest should read its config information from the environment variable `HARVEST_CONFIG` when supplied. This env var has higher precedence than the `--config` command-line argument. #645

### Fixes

- FlexGroup statistics should be aggregated across node and aggregates [#706](https://github.com/NetApp/harvest/issues/706) Thanks to @wally007 for reporting
  
- Network Details dashboard should use correct units and support variable sorting [#673](https://github.com/NetApp/harvest/issues/673) Thanks to @mamoep for reporting and reviewing the fix
 
- Harvest Systemd service should wait for network to start [#707](https://github.com/NetApp/harvest/pull/707) Thanks to @mamoep for reporting and fixing

- MetroCluster dashboard should use correct units and support variable sorting [#685](https://github.com/NetApp/harvest/issues/685) Thanks to @mamoep and @chris4789 for reporting this
    
- 7mode shelf plugin should handle cases where multiple channels have the same shelf id [#692](https://github.com/NetApp/harvest/issues/692) Thanks to @pilot7777 for reporting this on Slack

- Improve template YAML parsing when indentation varies [#704](https://github.com/NetApp/harvest/issues/704) Thanks to @mamoep for reporting this.
  
- Harvest should not include version information in its container name. [#660](https://github.com/NetApp/harvest/issues/660). Thanks to @wally007 for raising this. 

- Ignore missing Qtrees and improve uniqueness check on 7mode filers #782 and #797. Thanks to @jmg011 for reporting

- Qtree instance key order should be unified between 7mode and cDOT #807 Thanks to @jmg011 for reporting

- Workload detail volume collection should not try to create duplicate counters #803 Thanks to @luddite516 for reporting

- Harvest HTTP service discovery node should not attempt to publish Prometheus metrics to InfluxDB [#684](https://github.com/NetApp/harvest/pull/684)

- Grafana import should save auth token to the config file referenced by `HARVEST_CONFIG` when that environnement variable exists [#681](https://github.com/NetApp/harvest/pull/681)

- `bin/zapi` should print output [#715](https://github.com/NetApp/harvest/pull/715)

- Snapmirror dashboard should show correct number of SVM-DR relationships, last transfer, and health status #728 Thanks to Gaël Cantarero on Slack for reporting
  
- Ensure that properties defined in object templates override their parent properties #765

- Increase time that metrics are retained in Prometheus exporter from 3 minutes to 5 minutes #778 

- Remove the misplaced `SVM FCP Throughput` panel from the iSCSI drilldown section of the SVM details dashboard #821 Thanks to @florianmulatz for reporting and fixing

- When importing Grafana dashboards, remove the existing `id` and `uid` so Grafana treats the import as a create instead of an overwrite #825 Thanks to @luddite516 for reporting
  
- Relax the Grafana version check constraint so version `8.4.0-beta1` is considered `>=7.1` #828 Thanks to @ybizeul for reporting

- `bin/harvest status` should report `running` for pollers exporting to InfluxDB, instead of reporting that they are not running #835
  
- Pin the Grafana and Prometheus versions in the Docker compose workflow instead of pulling latest #822

---
## 21.11.1 / 2021-12-10
<!-- git log --no-decorate --no-merges --cherry-pick --right-only --oneline origin/release/21.08.0...origin/release/21.11.0 -->

This release is the same as [21.11.0](https://github.com/NetApp/harvest/releases/tag/v21.11.0) with an FSx dashboard fix for #737. If you are not monitoring an FSx system the 21.11.0 release is the same, no need to upgrade. We reverted a node-labels check in those dashboards because Harvest does not collect node data from FSx systems.

## 21.11.0 / 2021-11-08
<!-- git log --no-decorate --no-merges --cherry-pick --right-only --oneline origin/release/21.08.0...origin/release/21.11.0 -->

Highlights of this major release include:

- Early access to ONTAP REST collector
- Support for [Prometheus HTTP service discovery](https://github.com/NetApp/harvest/blob/main/cmd/exporters/prometheus/README.md#prometheus-http-service-discovery)
- New MetroCluster dashboard
- Qtree and Quotas collection
- We heard your ask, and we made it happen. We've separated cDOT and 7mode dashboards so each can evolve independently
- [Label sets](https://github.com/NetApp/harvest#labels) allow you to add additional key-value pairs to a poller's metrics[#538](https://github.com/NetApp/harvest/pull/538) and expose those labels in your dashboards
- Template merging was improved to keep your template changes separate from Harvest's
- Harvest poller's are more [deterministic](https://github.com/NetApp/harvest/pull/595) about picking [free ports](https://github.com/NetApp/harvest/pull/596)
- 37 bug fixes

**IMPORTANT** The LabelAgent `value_mapping` plugin is being deprecated in this
release and will be removed in the next release of Harvest. Use LabelAgent
`value_to_num` instead. See
[docs](https://github.com/NetApp/harvest/blob/main/cmd/poller/plugin/README.md#value_to_num)
for details. 

**IMPORTANT** After upgrade, don't forget to re-import all dashboards so you get new dashboard enhancements and fixes. 
You can re-import via `bin/harvest/grafana` cli or from the Grafana UI.

**IMPORTANT** RPM and Debian packages will be deprecated in the future, replaced
with Docker and native binaries. See
[#330](https://github.com/NetApp/harvest/issues/330) for details and tell us
what you think. Several of you have already weighed-in. Thanks! If you haven't,
please do.

**Known Issues**

The Unix collector is unable to monitor pollers running in containers. See [#249](https://github.com/NetApp/harvest/issues/249) for details.

### Enhancements

- :construction: [ONTAP started moving their APIs from ZAPI to REST](https://devnet.netapp.com/restapi.php) in ONTAP 9.6. Harvest adds an early access ONTAP REST collector in this release (config only). :confetti_ball: This is our first step among several as we prepare for the day that ZAPIs are turned off. The REST collector and seven templates are included in 21.11. These should be considered early access as we continue to improve them. If you try them out or have any feedback, let us know on Slack or [GitHub](https://github.com/NetApp/harvest/discussions). [#402](https://github.com/NetApp/harvest/issues/402)
  
- Harvest should have a [Prometheus HTTP service discovery](https://github.com/NetApp/harvest/blob/main/cmd/exporters/prometheus/README.md#prometheus-http-service-discovery) end-point to make it easier to add/remove pollers [#575](https://github.com/NetApp/harvest/pull/575)
  
- Harvest should include a MetroCluster dashboard [#539](https://github.com/NetApp/harvest/issues/539) Thanks @darthVikes for reporting
   
- Harvest should collect Qtree and Quota metrics [#522](https://github.com/NetApp/harvest/issues/522) Thanks @jmg011 for reporting and validating this works in your environment 
  
- SVM dashboard: Make NFS version a variable. SVM variable should allow selecting all SVMs for a cluster wide view [#454](https://github.com/NetApp/harvest/pull/454)

- Harvest should monitor ONTAP chassis sensors [#384](https://github.com/NetApp/harvest/issues/384) Thanks to @hashi825 for raising this issue and reviewing the pull request

- Harvest cluster dashboard should include `All` option in dropdown for clusters [#630](https://github.com/NetApp/harvest/pull/630) thanks @TopSpeed for raising this on Slack
 
- Harvest should collect volume `sis status` [#519](https://github.com/NetApp/harvest/issues/519) Thanks to @jmg011 for raising

- Separate cDOT and 7-mode dashboards allowing each to change independently [#489](https://github.com/NetApp/harvest/pull/489) [#501](https://github.com/NetApp/harvest/pull/501) [#547](https://github.com/NetApp/harvest/pull/547)

- Improve collector and object template merging and documentation [#493](https://github.com/NetApp/harvest/issues/493) [#555](https://github.com/NetApp/harvest/pull/555) Thanks @hashi825 for reviewing and suggesting improvements
 
- Harvest should support [label sets](https://github.com/NetApp/harvest#labels), allowing you to add additional key-value pairs to a poller's metrics[#538](https://github.com/NetApp/harvest/pull/538)

- `bin/grafana import` should create a matching label and rewrite queries to use chained variable when using label sets [#550](https://github.com/NetApp/harvest/pull/550)

- Harvest poller's should reuse their previous Prometheus port when restarted and be more deterministic about picking free ports [#596](https://github.com/NetApp/harvest/pull/596) [#595](https://github.com/NetApp/harvest/pull/595) Thanks to @cordelster for reporting

- Improve instantiated systemd template by specifying user/group, requires, and moving Unix pollers to the end of the list. [#643](https://github.com/NetApp/harvest/issues/643) Thanks to @mamoep for reporting and providing the changes! :sparkles:

- Harvest's Docker container should use local `conf` directory instead of copying into image. Makes upgrade and changing template files easier. [#511](https://github.com/NetApp/harvest/issues/511)
  
- Improve Disk dashboard by showing total number of disks by node and aggregate [#583](https://github.com/NetApp/harvest/pull/583)

- Harvest 7-mode dashboards should be provisioned when using Docker Compose workflow [#544](https://github.com/NetApp/harvest/pull/554)
  
- When upgrading, `bin/harvest grafana import` should add dashboards to a release-named folder so earlier dashboards are not overwritten [#616](https://github.com/NetApp/harvest/pull/616)

- `client_timeout` should be overridable in object template files [#563](https://github.com/NetApp/harvest/pull/563)

- Increase ZAPI client timeouts for volume and workloads objects [#617](https://github.com/NetApp/harvest/pull/617)

- Doctor: Ensure that all pollers export to unique Prometheus ports [#597](https://github.com/NetApp/harvest/pull/597)

- Improve execution performance of Harvest management commands :rocket: `bin/harvest start|stop|restart` [#600](https://github.com/NetApp/harvest/pull/600)
  
- Include eight cDOT dashboards that use InfluxDB datasource [#466](https://github.com/NetApp/harvest/pull/466). Harvest does not support InfluxDB dashboards for 7-mode. Thanks to @SamyukthaM for working on these

- Docs: Describe how Harvest converts template labels into Prometheus labels [#585](https://github.com/NetApp/harvest/issues/585)

- Docs: Improve Matrix documentation to better align with code [#485](https://github.com/NetApp/harvest/pull/485)

- Docs: Improve [ARCHITECTURE.md](https://github.com/NetApp/harvest/blob/main/ARCHITECTURE.md) [#603](https://github.com/NetApp/harvest/pull/603)

### Fixes

- Poller should report metadata when running on BusyBox [#529](https://github.com/NetApp/harvest/issues/529) Thanks to @charz for reporting issue and providing details
  
- Space used % calculation was incorrect for Cluster and Aggregate dashboards [#624](https://github.com/NetApp/harvest/issues/624) Thanks to @faguayot and @jorbour for reporting.  

- When ONTAP indicates a counter is deprecated, but a replacement is not provided, continue using the deprecated counter [#498](https://github.com/NetApp/harvest/pull/498)

- Harvest dashboard panels must specify a Prometheus datasource to correctly handles cases were a non-Prometheus default datasource is defined in Grafana. [#639](https://github.com/NetApp/harvest/issues/639) Thanks for reporting @MrObvious
 
- Prometheus datasource was missing on five dashboards (Network and Disk) [#566](https://github.com/NetApp/harvest/pull/566) Thanks to @survive-wustl for reporting 
  
- Document permissions that Harvest requires to monitor ONTAP with a read-only user [#559](https://github.com/NetApp/harvest/pull/559) Thanks to @survive-wustl for reporting and working with us to chase this down. :thumbsup: 

- Metadata dashboard should show correct status for running/stopped pollers [#567](https://github.com/NetApp/harvest/issues/567) Thanks to @cordelster for reporting

- Harvest should serve a human-friendly :corn: overview page of metric types when hitting the Prometheus end-point [#613](https://github.com/NetApp/harvest/issues/613) Thanks @cordelster for reporting
  
- SnapMirror plugin should include source_node [#608](https://github.com/NetApp/harvest/issues/608)
  
- Disk dashboard should use better labels in table details [#578](https://github.com/NetApp/harvest/pull/578) 
  
- SVM dashboard should show correct units and remove duplicate graph [#454](https://github.com/NetApp/harvest/pull/454)

- FCP plugin should work with 7-mode clusters [#464](https://github.com/NetApp/harvest/pull/464)

- Node values are missing from some 7-mode perf counters [#467](https://github.com/NetApp/harvest/pull/467)

- Nic state is missing from several network related dashboards [486](https://github.com/NetApp/harvest/issues/486)

- Reduce log noise when templates are not found since this is often expected [#606](https://github.com/NetApp/harvest/pull/606)
  
- Use `diagnosis-config-get-iter` to collect node status from 7-mode systems [#499](https://github.com/NetApp/harvest/pull/499)

- Node status is missing from 7-mode [#527](https://github.com/NetApp/harvest/pull/527)

- Improve 7-mode templates. Remove cluster from 7-mode. `yamllint` all templates  [#531](https://github.com/NetApp/harvest/pull/531)

- When saving Grafana auth token, make sure `bin/grafana` writes valid Yaml [#544](https://github.com/NetApp/harvest/pull/544)

- Improve Yaml parsing when different levels of indention are used in `harvest.yml`. You should see fewer `invalid indentation` messages. :clap: [#626](https://github.com/NetApp/harvest/pull/626)

- Unix poller should ignore `/proc` files that aren't readable [#249](https://github.com/NetApp/harvest/issues/249)

---

## 21.08.0 / 2021-08-31
<!-- git log --no-merges --cherry-pick --right-only --oneline release/21.05.4...release/21.08.0 -->

This major release introduces a Docker workflow that makes it a breeze to
standup Grafana, Prometheus, and Harvest with auto-provisioned dashboards. There
are seven new dashboards, example Prometheus alerts, and a bunch of fixes
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

**IMPORTANT** After upgrade, don't forget to re-import all dashboards so you get new dashboard enhancements and fixes. You can re-import via `bin/harvest/grafana` cli or from the Grafana UI.

**Known Issues**

We've improved several of the 7-mode dashboards this release, but there are still a number of gaps with 7-mode dashboards when compared to c-mode. We will address these in a point release by splitting the c-mode and 7-mode dashboards. See [#423](https://github.com/NetApp/harvest/issues/423) for details.

On RHEL and Debian, the example Unix collector does not work at the moment due to the `harvest` user lacking permissions to read the `/proc` filesystem. See [#249](https://github.com/NetApp/harvest/issues/249) for details.

### Enhancements

- Make it easy to install Grafana, Prometheus, and Harvest with Docker Compose and auto-provisioned dashboards. [#349](https://github.com/NetApp/harvest/pull/349)

- Lun, Volume Details, Node Details, Network Details, and SVM dashboards added to Harvest. Thanks to @jgasher for contributing five solid dashboards. [#458](https://github.com/NetApp/harvest/pull/458) [#482](https://github.com/NetApp/harvest/pull/482)
      
- Disk dashboard added to Harvest with disk type, status, uptime, and aggregate information. Thanks to @faguayot, @bengoldenberg, and @talshechanovitz for helping with this feature [#348](https://github.com/NetApp/harvest/issues/348) [#375](https://github.com/NetApp/harvest/pull/375) [#367](https://github.com/NetApp/harvest/pull/367) [#361](https://github.com/NetApp/harvest/pull/361)
  
- New SVM dashboard with NFS v3, v4, and v4.1 frontend drill-downs. Thanks to @burkl for contributing these. :tada: [#344](https://github.com/NetApp/harvest/issues/344)

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

- ZAPI templates with the same object name overwrite matrix data (impacted nfs and object_store_client_op templates). Thanks to @hashi825 for reporting this [#462](https://github.com/NetApp/harvest/issues/462)

- Lots of fixes for 7-mode dashboards and data collection. Thanks to @madhusudhanarya and @ybizeul for reporting. There's still more work to do for 7-mode, but we understand some of our customers rely on Harvest to help them monitor these legacy systems. [#383](https://github.com/NetApp/harvest/issues/383) [#441](https://github.com/NetApp/harvest/issues/441) [#423](https://github.com/NetApp/harvest/issues/423) [#415](https://github.com/NetApp/harvest/issues/415) [#376](https://github.com/NetApp/harvest/issues/376)

- Aggregate dashboard "space used column" should use current fill grade. Thanks to @florianmulatz for reporting. [#351](https://github.com/NetApp/harvest/issues/351)

- When building RPMs don't compile Harvest Python test code. Thanks to @madhusudhanarya for reporting. [#385](https://github.com/NetApp/harvest/issues/385)

- Harvest should collect include NVMe and fiber channel port counters. Thanks to @jgasher for submitting these. [#363](https://github.com/NetApp/harvest/issues/363)

- Harvest should export NFS v4 metrics. It does for v3 and v4.1, but did not for v4 due to a typo in the v4 ZAPI template. Thanks to @jgasher for reporting. [#481](https://github.com/NetApp/harvest/pull/481)
  
- Harvest panics when port_range is used in the Prometheus exporter and address is missing. Thanks to @ybizeul for reporting. [#357](https://github.com/NetApp/harvest/issues/357)

- Network dashboard fiber channel ports (FCP) should report read and write throughput [#445](https://github.com/NetApp/harvest/pull/445)

- Aggregate dashboard panel titles should match the information being displayed [#133](https://github.com/NetApp/harvest/issues/133)

- Harvest should handle ZAPIs that include signed integers. Most ZAPIs use unsigned integers, but a few return signed ones. Thanks for reporting @hashi825 [#384](https://github.com/NetApp/harvest/issues/384)

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
