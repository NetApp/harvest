# Change Log
## [Releases](https://github.com/NetApp/harvest/releases)

## 26.02.0 / 2026-02-11 Release
:pushpin: Highlights of this major release include:
## :star: New Features

- :medal_sports: Harvest includes a new BETA E-Series collector for inventory and performance metrics along with four E-Series dashboards and 47 panels.
  E-Series newly added dashboards:
  - E-Series: Array
  - E-Series: Controller
  - E-Series: Hardware
  - E-Series: Volume
    Thanks to @mamoep, @ReBaunana, @erikgruetter, @mark.pendrick, @darthVikes, @crollorc, @heinowalther, @ngocchiongnoi, @summertony15, @Venumadhu for raising.

- :medal_sports: Harvest already supports VictoriaMetrics exporter in pull mode, and with this release Harvest also supports VictoriaMetrics in push mode. More details are available for VictoriaMetrics push mode https://docs.victoriametrics.com/victoriametrics/single-server-victoriametrics/#how-to-import-time-series-data

- :medal_sports: Harvest includes an opt-in disk-based cache for improved memory efficiency. More details https://netapp.github.io/harvest/latest/prometheus-exporter/#disk_cache

- Harvest MCP now supports overriding the default HARVEST_TSDB_URL on a per-request basis. Thanks @gautcher for raising.

- Harvest includes fix to avoid double-counting of the shelf power. Thanks to @rmilkowski for raising and contributing a fix 🤘

- **IMPORTANT** KeyPerf is the default collector for volume performance metrics starting in `25.11` and this release `26.02` fixes an issue where KeyPerf Collector did not calculate FlexGroup latency.

- :gem: New dashboards and additional panels:
  - Harvest includes LUN Serial hex numbers in Lun table. Thanks @Venumadhu for raising.

- :closed_book: Documentation additions:
  - Added documentation recommending binding to all interfaces in HTTP mode. Thanks to Chris Gautcher for reporting!
  - Added new E-Series documentation
  - Added mcp installation.md documentation

- `harvest grafana metrics` prints the template path for each metric consumed in all Grafana dashboards. Thanks @songlin-rgb for raising.
- `harvest grafana import` adds additional options to customize the orgId, title, uid, and tags when importing the Grafana dashboards. Thanks @spapadop for reporting.

- Harvest provides an option to limit concurrent ONTAP HTTP connections. Thanks @songlin-rgb for raising.

- Harvest MCP enhancement:
  - Harvest MCP includes metric unit and type information for performance metrics. Thanks @gautcher for reporting.
  - Harvest adds health endpoint in harvest mcp. Thanks @Yann for reporting.
  - Harvest supports prometheus/victoriametrics endpoint retry in mcp-server. Thanks @Yann for reporting.

## Announcements

:bulb: **IMPORTANT** After upgrading, don't forget to re-import your dashboards to get all the new enhancements and fixes. You can import them via the `bin/harvest grafana import` CLI, from the Grafana UI, or from the 'Maintenance > Reset Harvest Dashboards' button in NAbox3. For NAbox4, this step is not needed.

:bulb: E-Series collector and dashboards are beta as we collect more feedback.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards for this release:

@mamoep, @gautcher, @Venumadhu, Chris Gautcher, @songlin-rgb, @spapadop

:seedling: This release includes 15 features, 21 bug fixes, 6 documentation, 1 testing, 4 refactoring, 22 miscellaneous, and 7 ci pull requests.

<details>

<summary>Expand for full list of pull requests</summary>

### :rocket: Features
- Limit concurrent collectors ([#4024](https://github.com/NetApp/harvest/pull/4024))
- Adding serial in hex format in LUN table ([#4034](https://github.com/NetApp/harvest/pull/4034))
- Override default HARVEST_TSDB_URL on a per-request basis for Harvest MCP ([#4042](https://github.com/NetApp/harvest/pull/4042))
- Disk cache ([#4033](https://github.com/NetApp/harvest/pull/4033))
- Add templatePath in grafana metric harvest cli ([#4056](https://github.com/NetApp/harvest/pull/4056))
- Support victoria metrics push exporter ([#4031](https://github.com/NetApp/harvest/pull/4031))
- Add units for ONTAP metrics missing them ([#4082](https://github.com/NetApp/harvest/pull/4082))
- Add support for specifying organization ID when importing dashb… ([#4087](https://github.com/NetApp/harvest/pull/4087))
- Add units for ONTAP metrics missing them ([#4091](https://github.com/NetApp/harvest/pull/4091))
- Add health endpoint in harvest mcp ([#4097](https://github.com/NetApp/harvest/pull/4097))
- E-Series collector infrastructure ([#4088](https://github.com/NetApp/harvest/pull/4088))
- Include options for modifying title, uid, tags when importing d… ([#4101](https://github.com/NetApp/harvest/pull/4101))
- Time series retry in harvest MCP ([#4105](https://github.com/NetApp/harvest/pull/4105))
- E-Series hardware metrics ([#4112](https://github.com/NetApp/harvest/pull/4112))
- Enable Qtree latency collection via KeyPerf collector ([#4139](https://github.com/NetApp/harvest/pull/4139))


### :bug: Bug Fixes
- StorageGRID Cached credential script tokens not expired on 401 ([#4011](https://github.com/NetApp/harvest/pull/4011))
- Fix activity label for volume analytics ([#4026](https://github.com/NetApp/harvest/pull/4026))
- Handle local/remote records in cluster schedule ([#4027](https://github.com/NetApp/harvest/pull/4027))
- Better error handling when trying to monitor an ONTAP cluster wi… ([#4041](https://github.com/NetApp/harvest/pull/4041))
- Update logic of hot/cold changes if total_footprint is missing in volume template ([#4046](https://github.com/NetApp/harvest/pull/4046))
- Support different ontap port Zapi/ZapiPerf collector ([#4061](https://github.com/NetApp/harvest/pull/4061))
- Volume top metrics are not available for Flexgroup volumes ([#4072](https://github.com/NetApp/harvest/pull/4072))
- Adding dynamic threshold of link speed in 2 network tables ([#4062](https://github.com/NetApp/harvest/pull/4062))
- RestPerf:Volume volume aggregate metrics error handling ([#4098](https://github.com/NetApp/harvest/pull/4098))
- Remove unsafe GetChildS GetChildren chaining ([#4102](https://github.com/NetApp/harvest/pull/4102))
- E-Series dashboard names ([#4114](https://github.com/NetApp/harvest/pull/4114))
- Avoid double-counting shelf power ([#4116](https://github.com/NetApp/harvest/pull/4116))
- E-Series dashboard names for container ([#4118](https://github.com/NetApp/harvest/pull/4118))
- Volume uuid should be instance_uuid ([#4121](https://github.com/NetApp/harvest/pull/4121))
- Top file note should point to correct discussion ([#4123](https://github.com/NetApp/harvest/pull/4123))
- Enable multiselect for array dashboard ([#4125](https://github.com/NetApp/harvest/pull/4125))
- Move cache log to debug ([#4127](https://github.com/NetApp/harvest/pull/4127))
- Fix units for capacity ([#4130](https://github.com/NetApp/harvest/pull/4130))
- Use clonedstring for gjson ([#4132](https://github.com/NetApp/harvest/pull/4132))
- Duplicate metrics for flashpool ([#4134](https://github.com/NetApp/harvest/pull/4134))
- Rest client should use its own error struct ([#4135](https://github.com/NetApp/harvest/pull/4135))
- KeyPerf collector doesn't calculate flexgroup latency ([#4137](https://github.com/NetApp/harvest/pull/4137))

### :closed_book: Documentation
- Update rest strategy guide ([#4039](https://github.com/NetApp/harvest/pull/4039))
- For HTTP mode recommend binding to all interfaces ([#4065](https://github.com/NetApp/harvest/pull/4065))
- Mention flagship models and which model was used for examples ([#4084](https://github.com/NetApp/harvest/pull/4084))
- E-Series documentation ([#4115](https://github.com/NetApp/harvest/pull/4115))
- Fix mcp installation.md ([#4124](https://github.com/NetApp/harvest/pull/4124))
- E-Series metric documentation ([#4133](https://github.com/NetApp/harvest/pull/4133))

### :wrench: Testing
- Ensure node GetChildren does not panic ([#4110](https://github.com/NetApp/harvest/pull/4110))

### Refactoring
- Make mcp go runnable ([#4018](https://github.com/NetApp/harvest/pull/4018))
- Address lint warnings ([#4032](https://github.com/NetApp/harvest/pull/4032))
- Lint issues ([#4038](https://github.com/NetApp/harvest/pull/4038))
- Fix potential resource leak ([#4064](https://github.com/NetApp/harvest/pull/4064))
- Address lint warnings ([#4081](https://github.com/NetApp/harvest/pull/4081))

### Miscellaneous
- Improve changelog twistie formatting ([#4025](https://github.com/NetApp/harvest/pull/4025))
- Update all dependencies ([#4019](https://github.com/NetApp/harvest/pull/4019))
- Merge release/25.11.0 to main ([#4023](https://github.com/NetApp/harvest/pull/4023))
- Update all dependencies ([#4029](https://github.com/NetApp/harvest/pull/4029))
- Update all dependencies ([#4036](https://github.com/NetApp/harvest/pull/4036))
- Update all dependencies ([#4047](https://github.com/NetApp/harvest/pull/4047))
- Track upstream go-version changes ([#4048](https://github.com/NetApp/harvest/pull/4048))
- Bump go ([#4054](https://github.com/NetApp/harvest/pull/4054))
- Lint issues ([#4055](https://github.com/NetApp/harvest/pull/4055))
- Update all dependencies ([#4058](https://github.com/NetApp/harvest/pull/4058))
- Update all dependencies ([#4066](https://github.com/NetApp/harvest/pull/4066))
- Update all dependencies ([#4068](https://github.com/NetApp/harvest/pull/4068))
- Track upstream gopsutil changes ([#4074](https://github.com/NetApp/harvest/pull/4074))
- Update all dependencies ([#4083](https://github.com/NetApp/harvest/pull/4083))
- Bump go ([#4090](https://github.com/NetApp/harvest/pull/4090))
- Update all dependencies ([#4092](https://github.com/NetApp/harvest/pull/4092))
- Update all dependencies ([#4104](https://github.com/NetApp/harvest/pull/4104))
- Bring harvest.cue up to date ([#4107](https://github.com/NetApp/harvest/pull/4107))
- Update all dependencies ([#4111](https://github.com/NetApp/harvest/pull/4111))
- Bump go ([#4113](https://github.com/NetApp/harvest/pull/4113))
- Update all dependencies ([#4119](https://github.com/NetApp/harvest/pull/4119))
- Bump dependencies ([#4122](https://github.com/NetApp/harvest/pull/4122))
- Bump go ([#4128](https://github.com/NetApp/harvest/pull/4128))

### :hammer: CI
- Unit test should fail ZapiPerf templates when detect caret ([#4043](https://github.com/NetApp/harvest/pull/4043))
- V-zhuravlev has signed the CCLA ([#4070](https://github.com/NetApp/harvest/pull/4070))
- Allow first party GitHub actions to use unpinned references ([#4079](https://github.com/NetApp/harvest/pull/4079))
- Fix ci issue ([#4094](https://github.com/NetApp/harvest/pull/4094))
- KuaJnio has signed the CCLA ([#4109](https://github.com/NetApp/harvest/pull/4109))
- Rmilkowski has signed the CCLA ([#4117](https://github.com/NetApp/harvest/pull/4117))
- Fix lint ([#4126](https://github.com/NetApp/harvest/pull/4126))
</details>


---
## 25.11.0 / 2025-11-10 Release
:pushpin: Highlights of this major release include:
## :star: New Features

- :medal_sports: We've created a [Harvest Model Context Protocol](https://netapp.github.io/harvest/latest/mcp/overview/) (MCP) server. The Harvest MCP server provides MCP clients like GitHub Copilot, Claude Desktop, and other large language models (LLMs) access to your infrastructure monitoring data collected by Harvest from ONTAP, StorageGRID, and Cisco systems.

- :fire: Harvest supports monitoring NetApp AFX clusters with this release. Performance metrics with the API name KeyPerf or StatPerf in the [ONTAP metrics documentation](https://netapp.github.io/harvest/latest/ontap-metrics/) are supported in AFX systems. As a result, some panels in the dashboards may be missing information.

- :gem: New dashboards and additional panels:
  - Harvest includes an ASAr2 dashboard with storage units and SAN initiator group panels.
  - Harvest includes a StorageGRID S3 dashboard. Thanks to @ofu48167 for raising!
  - Harvest includes a Hosts dashboard with SAN initiator groups. Thanks to @CJLvU for raising!
  - Harvest collects FlexCache metrics from FSx.
  - The StorageGRID Tenants dashboard includes tenant descriptions and bucket versioning. Thanks to @jowanw for raising!
  - The Volume dashboard includes an autosize table panel. Thanks to @roybatty2019 for raising!
  - The Network dashboard shows all ethernet port errors. Thanks to RobertWatson for raising!
  - The Datacenter dashboard includes a System Manager panel with links to ONTAP System Manager. Thanks to Ed Barron for raising!
  - The Data Protection dashboard includes a Snapshot Policy Violations panel that shows the number of snapshots outside the defined policy scope. Thanks to Lora NeyMan for raising!
  - The Volume dashboard includes panels on hot and cold data. Thanks to prime_kiwi_05259 for raising!
  - The Snapmirror Destination dashboard includes a "TopN Destination Volumes by Average Throughput" panel. Thanks to @roybatty2019 for raising!
  - The Volume dashboard includes a Snaplock panel. Thanks to @BrendonA667 for raising!
  - The MetroCluster dashboard includes IWarp and NVM mirror metrics. Thanks to @mamoep for raising!
  - The Security dashboard includes an anti-ransomware snapshots table. Thanks to @ybizeul for raising!
  - The Workload dashboard includes min IOPs and workload size in the adaptive QoS workload table. Thanks to Paqui for raising!
  - The LUN dashboard includes a LUN's block size in the LUN table. Thanks to Venumadhu for raising!

- :ear_of_rice: `harvest grafana import` includes a new command-line interface option (`show-datasource`) to show the datasource variable dropdown in dashboards, useful for multi-datasource setups. Thanks to @RockSolidScripts for raising!

- `harvest grafana import` includes a new command-line interface option (`add-cluster-label`) to rewrite all panel expressions to add the specified cluster label and variable. Thanks to @RockSolidScripts for raising!

- :closed_book: Documentation additions:
  - Added a tutorial for how to include StorageGRID-supplied dashboards into Harvest. Thanks to @ofu48167 for raising!
  - Included [ONTAP permissions](https://netapp.github.io/harvest/latest/prepare-cdot-clusters/#statperf-least-privilege-role) required for the [StatPerf collector](https://netapp.github.io/harvest/latest/configure-statperf/).
  - Clarified which APIs are used to collect each metric.
  - Clarified that the StatPerf collector does not work for FSx clusters due to ONTAP limitations.

- Harvest reports node-scoped metrics even when some nodes are down.

- Harvest's poller includes a `/health` endpoint for liveness checks. Thanks to @RockSolidScripts for raising!

- The FcpPort and NicCommon templates work with the StatPerf collector. This means the Network dashboard works with AFX and ASAr2 clusters.

## Announcements

:bangbang: **IMPORTANT** We've made changes to how volume performance metrics are collected. These changes are automatic and require no action from you unless you've customized Harvest's default `volume.yaml` templates. Continue reading for more details on the reasons behind this change and how to accommodate it.

By default, Harvest will now use the `KeyPerf` collector for volume performance metrics. This better aligns with ONTAP's recommendations and what System Manager shows.

The `default.yaml` files for `ZapiPerf` and `RestPerf` now include a `KeyPerf:` prefix for the volume template (e.g., `KeyPerf:volume.yaml`). This instructs Harvest to use the `KeyPerf` collector for volumes. More details are available at: #3900

:bangbang: **IMPORTANT** If you are using Docker Compose and want to keep your historical Prometheus data, please
read [how to migrate your Prometheus volume](https://github.com/NetApp/harvest/blob/main/docs/MigratePrometheusDocker.md).

:bulb: **IMPORTANT** After upgrading, don't forget to re-import your dashboards to get all the new enhancements and fixes. You can import them via the `bin/harvest grafana import` CLI, from the Grafana UI, or from the 'Maintenance > Reset Harvest Dashboards' button in NAbox3. For NAbox4, this step is not needed.

## Known Issues

- #3941 disabled the `restperf/volume_node.yaml` and `zapiperf/volume_node.yaml` templates because ONTAP provided incomplete metrics for them. The `node_vol` prefixed metrics are not used in any Harvest dashboard. If you still need these metrics, you can re-enable the templates in their corresponding `default.yaml`. See #3900 for details.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards for this release:

@BrendonA667, @CJLvU, @Falcon667, @RockSolidScripts, @jowanw, @mamoep, @ofu48167, @roybatty2019, @ybizeul

:seedling: This release includes 34 features, 20 bug fixes, 17 documentation, 1 testing, 8 refactoring, 21 miscellaneous, and 14 ci pull requests.

<details>

<summary>Expand for full list of pull requests</summary>

### :rocket: Features
- Allow Partial Aggregation For Node Scoped Objects ([#3811](https://github.com/NetApp/harvest/pull/3811))
- Grafana Import Should Include Option To Show Datasource Var ([#3830](https://github.com/NetApp/harvest/pull/3830))
- Adding Description And Versioning Enable In Tenant Dashboard ([#3833](https://github.com/NetApp/harvest/pull/3833))
- Adding Volume Autosize Details In Volume Dashboard ([#3851](https://github.com/NetApp/harvest/pull/3851))
- Network Dashboard Ethernet Port Errors Should Show All Errors ([#3852](https://github.com/NetApp/harvest/pull/3852))
- Datacenter Dashboard Should Include Links To System Manager ([#3853](https://github.com/NetApp/harvest/pull/3853))
- Monitor Snapshot Policy Compliance ([#3857](https://github.com/NetApp/harvest/pull/3857))
- Display Hot/Cold Data Of Volumes ([#3858](https://github.com/NetApp/harvest/pull/3858))
- Adding Storage-Units Rest Call In Asar2 ([#3867](https://github.com/NetApp/harvest/pull/3867))
- Adding Availability-Zones Rest Call In Asar2 ([#3870](https://github.com/NetApp/harvest/pull/3870))
- Harvest Should Load Asar2 Templates When Monitoring Asar2 Clusters ([#3871](https://github.com/NetApp/harvest/pull/3871))
- Add Health Endpoint To Harvest Poller ([#3879](https://github.com/NetApp/harvest/pull/3879))
- Adding Total Throughput Panel Of Destination Volume Of Sm-Sv ([#3880](https://github.com/NetApp/harvest/pull/3880))
- Collect Volume Snaplock Information ([#3883](https://github.com/NetApp/harvest/pull/3883))
- Adding Nvm_mirror Zapiperf Object And Removed Read_ops And Ops From Iwarp ([#3884](https://github.com/NetApp/harvest/pull/3884))
- Honour Volume Filter In Top Client/File In Volume ([#3888](https://github.com/NetApp/harvest/pull/3888))
- Harvest Mcp ([#3895](https://github.com/NetApp/harvest/pull/3895))
- Asar2 Storage Unit Dashboard ([#3898](https://github.com/NetApp/harvest/pull/3898))
- Use Keyperf Collector For Volume Performance Metrics ([#3909](https://github.com/NetApp/harvest/pull/3909))
- Include Node Model In Aggregate Dashboard ([#3929](https://github.com/NetApp/harvest/pull/3929))
- Arw Snapshot Template With Private Cli ([#3933](https://github.com/NetApp/harvest/pull/3933))
- Adding Static Counter File For Keyperf Asar2 Folder ([#3934](https://github.com/NetApp/harvest/pull/3934))
- Adding Min Iops And Workload Size In Adaptive Qos ([#3937](https://github.com/NetApp/harvest/pull/3937))
- Storagegrid S3 Dashboard ([#3940](https://github.com/NetApp/harvest/pull/3940))
- Disable Volumenode Metrics ([#3941](https://github.com/NetApp/harvest/pull/3941))
- Adding Unique Type Field In Metroclustercheck ([#3948](https://github.com/NetApp/harvest/pull/3948))
- Add Mcp Tool Details ([#3950](https://github.com/NetApp/harvest/pull/3950))
- Cluster-Label Flag Adds New Cluster Var/Label And Update All Panels ([#3955](https://github.com/NetApp/harvest/pull/3955))
- Add Plugins For Statperf Collector ([#3969](https://github.com/NetApp/harvest/pull/3969))
- Root Volume Enable/Disable Handled In Template ([#3975](https://github.com/NetApp/harvest/pull/3975))
- Include Tiering_minimum_cooling_days In Volume Template ([#3977](https://github.com/NetApp/harvest/pull/3977))
- Adding Block_size In Lun Perf ([#3982](https://github.com/NetApp/harvest/pull/3982))
- Add Nic And Fcp Port Support In Statperf ([#3989](https://github.com/NetApp/harvest/pull/3989))
- Hosts Dashboard ([#3994](https://github.com/NetApp/harvest/pull/3994))

### :bug: Bug Fixes
- Check Asup For All Pollers In Docker-Ci ([#3836](https://github.com/NetApp/harvest/pull/3836))
- Don't Fail Poller Startup When Zapi Is Disabled ([#3839](https://github.com/NetApp/harvest/pull/3839))
- Handle Ha Alerts For Non Ha Nodes ([#3881](https://github.com/NetApp/harvest/pull/3881))
- Cluster Label Rewriting Should Not Modify Metrics That Contain T… ([#3896](https://github.com/NetApp/harvest/pull/3896))
- Ignore Empty Templates ([#3911](https://github.com/NetApp/harvest/pull/3911))
- Indent Aggregate Json ([#3932](https://github.com/NetApp/harvest/pull/3932))
- Remove Cdot Tags From Storagegrid Fabricpool Dashboard ([#3938](https://github.com/NetApp/harvest/pull/3938))
- Include User And Group Columns In Quota Dashboard ([#3944](https://github.com/NetApp/harvest/pull/3944))
- Update Flexgroup Latency To 0 In Case Of No Ops ([#3952](https://github.com/NetApp/harvest/pull/3952))
- Handle Workload Filter Batch And Instance Removal ([#3958](https://github.com/NetApp/harvest/pull/3958))
- Handle Afx Ha Error ([#3960](https://github.com/NetApp/harvest/pull/3960))
- Disable `Latency_io_reqd` ([#3963](https://github.com/NetApp/harvest/pull/3963))
- Join With Group_left And Adding Workload Field ([#3968](https://github.com/NetApp/harvest/pull/3968))
- Auditlog Dashboard Duplicate Count In Victoriametrics ([#3973](https://github.com/NetApp/harvest/pull/3973))
- Fix Versioning Mapping In Dashboard ([#4000](https://github.com/NetApp/harvest/pull/4000))
- Correct Query With Group_left ([#4001](https://github.com/NetApp/harvest/pull/4001))
- Harvest Target File Should Use Soft Dependency ([#4002](https://github.com/NetApp/harvest/pull/4002))
- Cisco Lldp Should Handle Instances With The Same Chassisid ([#4004](https://github.com/NetApp/harvest/pull/4004))
- Storagegrid Cached Credential Script Tokens Not Expired On 401 ([#4010](https://github.com/NetApp/harvest/pull/4010))
- Statperf Multi Line Handling ([#4017](https://github.com/NetApp/harvest/pull/4017))

### :closed_book: Documentation
- Remove Invalid Api Url From Permissions ([#3835](https://github.com/NetApp/harvest/pull/3835))
- Fix Asar2 Spelling ([#3906](https://github.com/NetApp/harvest/pull/3906))
- Mcp Documentation ([#3916](https://github.com/NetApp/harvest/pull/3916))
- Update Readme To Mention Mcp ([#3923](https://github.com/NetApp/harvest/pull/3923))
- Use Consistent Role Name In Documentation ([#3928](https://github.com/NetApp/harvest/pull/3928))
- Add Configuration Link For Mcp ([#3931](https://github.com/NetApp/harvest/pull/3931))
- Update Statperf Permissions ([#3949](https://github.com/NetApp/harvest/pull/3949))
- Update Volume Metric Doc ([#3954](https://github.com/NetApp/harvest/pull/3954))
- Update Ontap Metrics With Actual Api ([#3956](https://github.com/NetApp/harvest/pull/3956))
- Tutorial To Add Storagegrid Supplied Dashboards Into Harvest ([#3957](https://github.com/NetApp/harvest/pull/3957))
- Add Afx Testing In Release Checklist ([#3961](https://github.com/NetApp/harvest/pull/3961))
- Build From Source Instructions For Mcp ([#3964](https://github.com/NetApp/harvest/pull/3964))
- Add Nabox4 Config Collection ([#3978](https://github.com/NetApp/harvest/pull/3978))
- Remove Statperf For Fsx ([#3981](https://github.com/NetApp/harvest/pull/3981))
- Add Log File Location Change Steps To Quickstart ([#3993](https://github.com/NetApp/harvest/pull/3993))
- Add Flexcache As Supported With Fsx ([#4005](https://github.com/NetApp/harvest/pull/4005))
- Update Metric Docs ([#4014](https://github.com/NetApp/harvest/pull/4014))

### :wrench: Testing
- Use Asserts In Tests ([#3863](https://github.com/NetApp/harvest/pull/3863))

### Refactoring
- Rename Tag Mapper Part 1 ([#3864](https://github.com/NetApp/harvest/pull/3864))
- Rename Tag Mapper Part 2 ([#3865](https://github.com/NetApp/harvest/pull/3865))
- Rename Asar2 Dashboard To Overview ([#3920](https://github.com/NetApp/harvest/pull/3920))
- Only Check If Zapis Exist When A Zapi Collector Is Desired ([#3921](https://github.com/NetApp/harvest/pull/3921))
- Only Check If Zapis Exist When A Zapi Collector Is Desired ([#3924](https://github.com/NetApp/harvest/pull/3924))
- Add Debug Logging For Rest Href ([#3943](https://github.com/NetApp/harvest/pull/3943))
- Remove Checksum Generation Since Github Provides Digests Now ([#3983](https://github.com/NetApp/harvest/pull/3983))
- Use Strings.builder To Improve Performance ([#3984](https://github.com/NetApp/harvest/pull/3984))

### Miscellaneous
- Merge Release/25.08.0 To Main ([#3831](https://github.com/NetApp/harvest/pull/3831))
- Update All Dependencies ([#3841](https://github.com/NetApp/harvest/pull/3841))
- Merge Release/25.08.1 To Main ([#3848](https://github.com/NetApp/harvest/pull/3848))
- Update All Dependencies ([#3856](https://github.com/NetApp/harvest/pull/3856))
- Track Upstream Gopsutil Changes ([#3872](https://github.com/NetApp/harvest/pull/3872))
- Go Bump ([#3873](https://github.com/NetApp/harvest/pull/3873))
- Update All Dependencies ([#3876](https://github.com/NetApp/harvest/pull/3876))
- Update Astral-Sh/Setup-Uv Digest To B75a909 ([#3887](https://github.com/NetApp/harvest/pull/3887))
- Update Astral-Sh/Setup-Uv Digest To 208B0c0 ([#3903](https://github.com/NetApp/harvest/pull/3903))
- Update All Dependencies ([#3918](https://github.com/NetApp/harvest/pull/3918))
- Track Upstream Gopsutil Changes ([#3922](https://github.com/NetApp/harvest/pull/3922))
- Update All Dependencies ([#3930](https://github.com/NetApp/harvest/pull/3930))
- Bump Go ([#3936](https://github.com/NetApp/harvest/pull/3936))
- Update All Dependencies ([#3945](https://github.com/NetApp/harvest/pull/3945))
- Update All Dependencies ([#3965](https://github.com/NetApp/harvest/pull/3965))
- Update All Dependencies ([#3971](https://github.com/NetApp/harvest/pull/3971))
- Bump Modelcontextprotocol/Go-Sdk ([#3987](https://github.com/NetApp/harvest/pull/3987))
- Update All Dependencies ([#3991](https://github.com/NetApp/harvest/pull/3991))
- Bump Modelcontextprotocol/Go-Sdk ([#3996](https://github.com/NetApp/harvest/pull/3996))
- Track Upstream Gopsutil Changes ([#3997](https://github.com/NetApp/harvest/pull/3997))
- Update Jenkins Href To New Server ([#4015](https://github.com/NetApp/harvest/pull/4015))

### :hammer: CI
- Update Integration Dependency ([#3818](https://github.com/NetApp/harvest/pull/3818))
- Bump Go ([#3834](https://github.com/NetApp/harvest/pull/3834))
- Disable Fips Check For Clusters That Don't Support Fips ([#3837](https://github.com/NetApp/harvest/pull/3837))
- Update Makefile Go Version ([#3850](https://github.com/NetApp/harvest/pull/3850))
- Fix Ci Issues ([#3877](https://github.com/NetApp/harvest/pull/3877))
- Enable The Gopls Modernize Analyzer ([#3890](https://github.com/NetApp/harvest/pull/3890))
- Update Mcp Version ([#3913](https://github.com/NetApp/harvest/pull/3913))
- Mcp Container Publish ([#3914](https://github.com/NetApp/harvest/pull/3914))
- Bump Go ([#3951](https://github.com/NetApp/harvest/pull/3951))
- Reduce Nightly Build Time ([#3953](https://github.com/NetApp/harvest/pull/3953))
- Add Sha File To Nightly Builds ([#3980](https://github.com/NetApp/harvest/pull/3980))
- Update Ci Machine ([#3985](https://github.com/NetApp/harvest/pull/3985))
- Bump Go ([#4008](https://github.com/NetApp/harvest/pull/4008))
- Handle Renovate[Bot] Pull Requests In Changelog ([#4012](https://github.com/NetApp/harvest/pull/4012))
</details>
---

## 25.08.1 / 2025-08-18 Release
:pushpin: This release is the same as version 25.08.0, with a fix for an issue where the ONTAP REST collector fails to start if ZAPIs are disabled on the cluster.

**Upgrade Recommendation:** Only upgrade if you are monitoring clusters with ZAPIs disabled. If ZAPIs are enabled, you can continue using the 25.08.0.

**Full Changelog**: https://github.com/NetApp/harvest/compare/v25.08.0...v25.08.1

---

## 25.08.0 / 2025-08-13 Release
:pushpin: Highlights of this major release include:
## :star: New Features

- [StatPerf](https://netapp.github.io/harvest/latest/configure-statperf/) Collector
  - This collector is designed for environments where ZapiPerf, RestPerf, or KeyPerf collectors can not be used and uses the well known ONTAP statistics CLI command to gather performance statistics.

- :gem: Three new dashboards:
  - Multi-admin verification (MAV) Dashboard provides a real-time overview of Multi-Admin Verification requests, tracking their status, approvals, and pending actions for enhanced security and operational visibility.
  - FPolicy dashboard for monitoring FPolicy performance metrics at the policy, SVM, and server levels.
  - ONTAP:Switch dashboard that provides details about switches connected to ONTAP.

- Cisco switch dashboard updates: :100: Thanks to @roybatty2019 for raising this issue and providing valuable guidance and examples.
  - Individual fan speeds are now displayed separately from zone speeds.
  - LLDP and CDP parsing have been refined with consistent field naming and improved data handling
  - New traffic monitoring metrics

- :star:
  - Quota and FSA dashboards now support filtering by volume tags.
  - Added a Junction Path variable in the Volume dashboard.
  - Added bucket quotas in StorageGrid Tenant dashboard.
  - Added "Volume" and "Idle Timeout" columns to the CIFS sessions table in the SMB Dashboard.
  - Added Used% in the bucket table within Tenant dashboard.

- :closed_book: Documentation additions
  - Navigate to your local Grafana dashboards from the metrics documentation by linking to your Grafana instance.
  - Added documentation for Cisco Switch and StorageGrid metrics.

## Announcements

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the 'bin/harvest grafana import' CLI, from the Grafana UI, or from the 'Maintenance > Reset Harvest Dashboards' button in NAbox3. For NAbox4, this step is not needed.

## Known Issues

:bulb: **IMPORTANT** FSx ZapiPerf workload collector fails to collect metrics, please use RestPerf instead.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards
this release:

@BrendonA667, @Falcon667, @T1r0l, @anguswilliams, @datamuc, @jowanw, @mamoep, @mhbeh, @mishraavinash88, @roybatty2019

:seedling: This release includes 18 features, 19 bug fixes, 10 documentation, 4 refactoring, 1 miscellaneous, and 9 ci pull requests.

### :rocket: Features
- Include Shelf Power Usage In "Average Power/Used Tb" And "Avera… ([#3705](https://github.com/NetApp/harvest/pull/3705))
- Add Array Support To Statperf Collector ([#3706](https://github.com/NetApp/harvest/pull/3706))
- Doc Includes Usage Detail Of Given Metric In Dashboards ([#3710](https://github.com/NetApp/harvest/pull/3710))
- Add Filtering Support For Statperf ([#3713](https://github.com/NetApp/harvest/pull/3713))
- Adding Junction Path Var In Volume Dashboard ([#3716](https://github.com/NetApp/harvest/pull/3716))
- Add Storagegrid Bucket Quota ([#3720](https://github.com/NetApp/harvest/pull/3720))
- Adding Volume Column In Smb Dashboard In Cifs Sessions Table ([#3721](https://github.com/NetApp/harvest/pull/3721))
- Added Used% In Bucket Table In Tenant Dashboard ([#3729](https://github.com/NetApp/harvest/pull/3729))
- Adding Legend In Dashboard Panels ([#3734](https://github.com/NetApp/harvest/pull/3734))
- Adding Idle Duration In Cifs Session Table ([#3744](https://github.com/NetApp/harvest/pull/3744))
- Mav Request Dashboard ([#3746](https://github.com/NetApp/harvest/pull/3746))
- Disable Statperf Templates ([#3753](https://github.com/NetApp/harvest/pull/3753))
- Add Volume Tags Support For Quota And Fsa Dashboard ([#3769](https://github.com/NetApp/harvest/pull/3769))
- Fpolicy Dashboard ([#3777](https://github.com/NetApp/harvest/pull/3777))
- Tags As Labels For Volume ([#3786](https://github.com/NetApp/harvest/pull/3786))
- Enhance Fan Metrics And Parsing For Cisco Switches ([#3790](https://github.com/NetApp/harvest/pull/3790))
- Add Tagmapper Plugin For Volume Labels ([#3796](https://github.com/NetApp/harvest/pull/3796))
- Replace Ping With Go Code To Reduce External Dependencies ([#3801](https://github.com/NetApp/harvest/pull/3801))

### :bug: Bug Fixes
- Update Instance Key In Snapshot Policy ([#3701](https://github.com/NetApp/harvest/pull/3701))
- Storagegrid Overview Dashboard - "S3 Api Requests" Panel Should … ([#3726](https://github.com/NetApp/harvest/pull/3726))
- "Svm Cifs Connections And Open Files" Panel Should Include Svm I… ([#3728](https://github.com/NetApp/harvest/pull/3728))
- Fsa Time Formatting ([#3733](https://github.com/NetApp/harvest/pull/3733))
- Storagegrid Panel Should Include Units ([#3737](https://github.com/NetApp/harvest/pull/3737))
- Statperf Collector Should Retry With Smaller Batch Size When Ont… ([#3748](https://github.com/NetApp/harvest/pull/3748))
- Handle Sorting Of The Port Labels In Ifgroup ([#3750](https://github.com/NetApp/harvest/pull/3750))
- After Failover/Giveback Volume Dashboard Data Query Failed ([#3758](https://github.com/NetApp/harvest/pull/3758))
- Only Show Shelves Which Are Local In Mcc Cluster Case ([#3759](https://github.com/NetApp/harvest/pull/3759))
- Publish Ontap Array As A Gauge Instead Of Histogram ([#3760](https://github.com/NetApp/harvest/pull/3760))
- Cisco Dashboard Switch Details Should Only Show Columns Once ([#3761](https://github.com/NetApp/harvest/pull/3761))
- Support Alerts Should Publish Recent Data ([#3766](https://github.com/NetApp/harvest/pull/3766))
- Handle Outage Parsing In Disk Rest Template ([#3768](https://github.com/NetApp/harvest/pull/3768))
- Reduce Reason Label In Metric Metadata_component_status ([#3776](https://github.com/NetApp/harvest/pull/3776))
- Duplicate Rows In The Target Systems Panel Of The Metadata Dashboard ([#3789](https://github.com/NetApp/harvest/pull/3789))
- Remove Index From Fsa ([#3792](https://github.com/NetApp/harvest/pull/3792))
- Add Sort For Deterministic Order In Test ([#3794](https://github.com/NetApp/harvest/pull/3794))
- Aggregate Dashboard Should Show % Inactive ([#3803](https://github.com/NetApp/harvest/pull/3803))
- Ciscorest Collector In Default List Causes Ontap Collectors To Fail ([#3805](https://github.com/NetApp/harvest/pull/3805))

### :closed_book: Documentation
- Add Storagegrid Port Information ([#3699](https://github.com/NetApp/harvest/pull/3699))
- Add Instance_add Documentation For Endpoints ([#3715](https://github.com/NetApp/harvest/pull/3715))
- Add Cisco And Sg Metric Documentation ([#3764](https://github.com/NetApp/harvest/pull/3764))
- Add Details For Node Total Data ([#3767](https://github.com/NetApp/harvest/pull/3767))
- Add Mav Note For Statperf Collector ([#3782](https://github.com/NetApp/harvest/pull/3782))
- Add Flexcache Support For Rest/Statperf Collector ([#3793](https://github.com/NetApp/harvest/pull/3793))
- Add Quota Dashboard Information ([#3809](https://github.com/NetApp/harvest/pull/3809))
- Clarify That Restperf Is Upgraded To Keyperf For Asa R2 Clusters ([#3810](https://github.com/NetApp/harvest/pull/3810))
- Clarify That Statperf Is Needed For Asa R2 Clusters ([#3815](https://github.com/NetApp/harvest/pull/3815))
- Update Metric Docs ([#3820](https://github.com/NetApp/harvest/pull/3820))

### Refactoring
- Reuse Parsecounter Method For Generate Metric Doc ([#3709](https://github.com/NetApp/harvest/pull/3709))
- Remove Unused Templates For Statperf Collector ([#3784](https://github.com/NetApp/harvest/pull/3784))
- Simplify Parsing Cisco Fan Metrics ([#3797](https://github.com/NetApp/harvest/pull/3797))
- Catch Nil Dereferences ([#3807](https://github.com/NetApp/harvest/pull/3807))

### Miscellaneous
- Merge Release/25.05.1 ([#3698](https://github.com/NetApp/harvest/pull/3698))

### :hammer: CI
- Remove Unused Flags ([#3712](https://github.com/NetApp/harvest/pull/3712))
- Add Statperf Integration Tests ([#3714](https://github.com/NetApp/harvest/pull/3714))
- Remove Cr.netapp.io And Jfrog ([#3749](https://github.com/NetApp/harvest/pull/3749))
- Bump Go ([#3754](https://github.com/NetApp/harvest/pull/3754))
- Remove Smoke Test From Fips And Rpm ([#3756](https://github.com/NetApp/harvest/pull/3756))
- Disable Nolintlint Due To False Positives ([#3775](https://github.com/NetApp/harvest/pull/3775))
- Stop Pollers For Rpm And Containers After Tests ([#3787](https://github.com/NetApp/harvest/pull/3787))
- Bump Go ([#3808](https://github.com/NetApp/harvest/pull/3808))
- Update Integration Dependency (#3818) ([#3819](https://github.com/NetApp/harvest/pull/3819))

---


## 25.05.1 / 2025-06-10 Release
:pushpin: This release is identical to 25.05.0, if you are using the Cisco collector, we recommend upgrading to version 25.05.1 to reduce cardinality issues caused by storing a switch's uptime as a label instead of a metric value.

This release also includes:

1. Introduced a new `ONTAP: Switch` dashboard that provides detailed information about switches connected to ONTAP.
2. Enhanced functionality to parse the Cisco version when the RCF is missing.
3. Updated to Golang 1.23.4, which includes several security vulnerability fixes (CVEs).
4. MetroCluster internal SVMs and volumes are no longer exported when they are offline.

---

## 25.05.0 / 2025-05-19 Release
pushpin: Highlights of this major release include:
## :star: New Features

- Cisco Switch collector:
  - Harvest collects metrics from all supported MetroCluster Cisco switches. More details [here](https://netapp.github.io/harvest/latest/configure-cisco-rest).
  - Harvest collects environmental, ethernet, optics, interface, link layer discovery protocol (LLDP), Cisco discovery protocol (CDP), and version related details.
  - Harvest includes a new Cisco switch dashboard. Thanks to @BrendonA667, Mamoep, and Eric Brüning for reporting and providing valuable feedback on this feature.

- Harvest includes a new performance collector named KeyPerf, designed to gather performance counters from ONTAP objects that include a `statistics` field in their REST responses. More details [here](https://netapp.github.io/harvest/latest/configure-keyperf).

- Harvest supports auditing volume operations such as create,delete and modify via ONTAP CLI or REST commands, tracked through the `ONTAP: AuditLog` dashboard. Thanks @mvilam79 for reporting. More details [here](https://github.com/NetApp/harvest/discussions/3478).

- Harvest supports filtering for the RestPerf collector. See [Filter](https://netapp.github.io/harvest/latest/configure-rest/#filter) for more detail.

- Harvest collects vscan server pool active connection. Thanks @BrendonA667 for reporting.

- Harvest collects uptime in lif perf templates and shows them in the SVM dashboard. Thanks to @Pengng88 for reporting.

- Harvest collects volume footprint metrics and displays them through the Volume dashboard. Thanks to @Robert Brown for reporting.

- Harvest includes a beta template to collect ethernet switch ports. Thanks to @Robert Watson for reporting!

- :star: Several of the existing dashboards include new panels in this release:
  - The `Disk` dashboard updates CP panels `Disk Utilization` panel.
  - The `Node` dashboard include the Node column in the `Node Detail` panel.
  - The `Quota` dashboard includes `Space Used` panel. Thanks @razaahmed for reporting.
  - The `Aggregate` dashboard includes `Growth Rate` panel. Thanks @Preston Nguyen for reporting.
  - The `Volume` dashboard includes `Growth Rate` panel. Thanks @Preston Nguyen for reporting.
  - The `Volume` dashboard includes volume footprint metrics in `FabricPool` panel. Thanks @RBrown for reporting.

## Announcements

:bangbang: **IMPORTANT** If using Docker Compose and you want to keep your historical Prometheus data, please read [how to migrate your Prometheus volume](https://github.com/NetApp/harvest/blob/main/docs/MigratePrometheusDocker.md)

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the 'bin/harvest grafana import' CLI, from the Grafana UI, or from the 'Maintenance > Reset Harvest Dashboards' button in NAbox3. For NAbox4, this step is not needed.

## Known Issues

:bulb: **IMPORTANT** FSx ZapiPerf workload collector fails to collect metrics, please use RestPerf instead.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards this release:

@WayneShen2, @mvilam79, @RobbW, @Robert Watson, @roller, @Pengng88, @gaur-piyush, @Chris Gautcher, @BrendonA667, @razaahmed, @nicolai-hornung-bl, @Preston Nguyen, @Robert Brown, @jay-law

:seedling: This release includes 28 features, 28 bug fixes, 13 documentation, 17 refactoring, 16 miscellaneous, and 11 ci pull requests.

### :rocket: Features
- Disable qtree perf metrics for KeyPerf collector ([#3488](https://github.com/NetApp/harvest/pull/3488))
- Volume Audit log ([#3479](https://github.com/NetApp/harvest/pull/3479))
- Handled duplicate instance issue in clustersoftware plugin ([#3486](https://github.com/NetApp/harvest/pull/3486))
- Split cp panels in disk dashboard ([#3496](https://github.com/NetApp/harvest/pull/3496))
- Adding uptime in lif perf templates ([#3507](https://github.com/NetApp/harvest/pull/3507))
- Harvest EMS Events label plugin ([#3511](https://github.com/NetApp/harvest/pull/3511))
- Filter support for RestPerf Collector ([#3514](https://github.com/NetApp/harvest/pull/3514))
- Adding vscan server pool rest template and plugin changes ([#3519](https://github.com/NetApp/harvest/pull/3519))
- Synthesize a timestamp when it is missing from KeyPerf responses ([#3544](https://github.com/NetApp/harvest/pull/3544))
- Node dashboard should include the Node column in the Node detai… ([#3553](https://github.com/NetApp/harvest/pull/3553))
- Adding format for promql in cluster dashboard ([#3538](https://github.com/NetApp/harvest/pull/3538))
- Harvest should monitor Cisco 3K and 9K switches ([#3559](https://github.com/NetApp/harvest/pull/3559))
- Adding space used time series panel in quota dashboard ([#3561](https://github.com/NetApp/harvest/pull/3561))
- Cisco collector should collect optics metrics ([#3575](https://github.com/NetApp/harvest/pull/3575))
- Private CLI perf collector StatPerf ([#3566](https://github.com/NetApp/harvest/pull/3566))
- Cisco collector should collect optics metrics for transceivers … ([#3580](https://github.com/NetApp/harvest/pull/3580))
- Add growth rate panel for Aggregate ([#3582](https://github.com/NetApp/harvest/pull/3582))
- Use timestamp provided by CLI in statperf ([#3585](https://github.com/NetApp/harvest/pull/3585))
- Add crc error for switch interface ([#3590](https://github.com/NetApp/harvest/pull/3590))
- Dedup statperf against other perf collectors ([#3592](https://github.com/NetApp/harvest/pull/3592))
- Harvest should collect volume footprint metrics ([#3598](https://github.com/NetApp/harvest/pull/3598))
- Harvest should collect ethernet switch ports ([#3601](https://github.com/NetApp/harvest/pull/3601))
- Adding cisco switch dashboard ([#3574](https://github.com/NetApp/harvest/pull/3574))
- Add growth rate for volume and aggregate ([#3610](https://github.com/NetApp/harvest/pull/3610))
- Update Cisco dashboard units and comment ([#3613](https://github.com/NetApp/harvest/pull/3613))
- Add Volume footprint metrics to Volume Dashboard ([#3624](https://github.com/NetApp/harvest/pull/3624))
- Include checksums with release artifacts ([#3628](https://github.com/NetApp/harvest/pull/3628))
- Cisco collector should collect CDP and LLDP metrics ([#3638](https://github.com/NetApp/harvest/pull/3638))

### :bug: Bug Fixes
- Handled empty node name in clustersoftware plugin ([#3460](https://github.com/NetApp/harvest/pull/3460))
- Duplicate timeseries in volume dashboard ([#3483](https://github.com/NetApp/harvest/pull/3483))
- Update title of number of snapmirror transfers ([#3485](https://github.com/NetApp/harvest/pull/3485))
- Network dashboard link speed units should be Megabits per second ([#3491](https://github.com/NetApp/harvest/pull/3491))
- Workload and workload_volume templates should invoke the instance task before the data task ([#3498](https://github.com/NetApp/harvest/pull/3498))
- Handled empty scanner and export false case for vscan ([#3502](https://github.com/NetApp/harvest/pull/3502))
- KeyPerf Collector Volume stats are incorrect for flexgroup ([#3520](https://github.com/NetApp/harvest/pull/3520))
- EMS cache handling ([#3524](https://github.com/NetApp/harvest/pull/3524))
- IWARP read and write IOPS for ZAPI should be expressed as rate ([#3550](https://github.com/NetApp/harvest/pull/3550))
- Aligning Harvest Dashboard node metrics with ONTAP CLI Data ([#3549](https://github.com/NetApp/harvest/pull/3549))
- Handle system:node deprecate metrics in ZapiPerf ([#3554](https://github.com/NetApp/harvest/pull/3554))
- Update namespace counters ([#3558](https://github.com/NetApp/harvest/pull/3558))
- StorageGrid Collector handles global_prefix inconsistently ([#3565](https://github.com/NetApp/harvest/pull/3565))
- `grafana import` should add labels to all panel expressions when… ([#3567](https://github.com/NetApp/harvest/pull/3567))
- Cisco environment plugin should trim watts ([#3572](https://github.com/NetApp/harvest/pull/3572))
- Handle string parsing for switch templates ([#3578](https://github.com/NetApp/harvest/pull/3578))
- yaml parsing should handle key/values with spaces, colons, quotes ([#3581](https://github.com/NetApp/harvest/pull/3581))
- Handle array element for optic metrics ([#3589](https://github.com/NetApp/harvest/pull/3589))
- Filter label for ems destination is missing ([#3596](https://github.com/NetApp/harvest/pull/3596))
- Harvest should collect ethernet switch ports when timestamp is m… ([#3603](https://github.com/NetApp/harvest/pull/3603))
- Handle histogram skips in exporter ([#3606](https://github.com/NetApp/harvest/pull/3606))
- Handled nil aggr instance in aggr plugin ([#3607](https://github.com/NetApp/harvest/pull/3607))
- Handle HA and volume move alerts ([#3611](https://github.com/NetApp/harvest/pull/3611))
- Poller Union2 should handle prom_port ([#3614](https://github.com/NetApp/harvest/pull/3614))
- Handle empty values in template ([#3626](https://github.com/NetApp/harvest/pull/3626))
- Improve Cisco RCF parsing ([#3629](https://github.com/NetApp/harvest/pull/3629))
- Grafana import should refuse to redirect ([#3632](https://github.com/NetApp/harvest/pull/3632))
- Handle empty values in template ([#3627](https://github.com/NetApp/harvest/pull/3627))
- Vscanpool plugin should only ask for fields it uses ([#3639](https://github.com/NetApp/harvest/pull/3639))
- Handle uname in qtree zapi plugin ([#3641](https://github.com/NetApp/harvest/pull/3641))

### :closed_book: Documentation
- Add changelog discussion link ([#3495](https://github.com/NetApp/harvest/pull/3495))
- Handled plugin custom prefix name for metrics ([#3493](https://github.com/NetApp/harvest/pull/3493))
- Asar2 support ([#3535](https://github.com/NetApp/harvest/pull/3535))
- Add labels metric doc ([#3532](https://github.com/NetApp/harvest/pull/3532))
- Update private cli ONTAP link ([#3591](https://github.com/NetApp/harvest/pull/3591))
- Harvest should document volume footprint metrics ([#3599](https://github.com/NetApp/harvest/pull/3599))
- StatPerf collector documentation ([#3600](https://github.com/NetApp/harvest/pull/3600))
- Document ethernet switch port counters ([#3604](https://github.com/NetApp/harvest/pull/3604))
- Document CiscoRest collector ([#3619](https://github.com/NetApp/harvest/pull/3619))
- Fix restperf filter doc ([#3622](https://github.com/NetApp/harvest/pull/3622))
- Update metric doc ([#3634](https://github.com/NetApp/harvest/pull/3634))
- Add beta to StatPerf docs ([#3635](https://github.com/NetApp/harvest/pull/3635))
- Fix default schedule values for collector ([#3642](https://github.com/NetApp/harvest/pull/3642))

### Refactoring
- Remove tidwall match and pretty dependencies ([#3503](https://github.com/NetApp/harvest/pull/3503))
- Update log message ([#3526](https://github.com/NetApp/harvest/pull/3526))
- Debug build logs ([#3536](https://github.com/NetApp/harvest/pull/3536))
- Revert debug build logs ([#3537](https://github.com/NetApp/harvest/pull/3537))
- Replace benchmark.N with benchmark.Loop() ([#3547](https://github.com/NetApp/harvest/pull/3547))
- Remove zapiperf debug log for qos ([#3560](https://github.com/NetApp/harvest/pull/3560))
- Support root aggrs in rest template ([#3569](https://github.com/NetApp/harvest/pull/3569))
- Replace `gopkg.in/yaml` with `github.com/goccy/go-yaml` ([#3573](https://github.com/NetApp/harvest/pull/3573))
- Remove unnecessary debug logs ([#3579](https://github.com/NetApp/harvest/pull/3579))
- Correct error messages for health ([#3583](https://github.com/NetApp/harvest/pull/3583))
- Workaround Cisco truncation issue by using cli_show_array ([#3586](https://github.com/NetApp/harvest/pull/3586))
- Eliminate superfluous error ([#3588](https://github.com/NetApp/harvest/pull/3588))
- Handle histogram skips in exporter ([#3608](https://github.com/NetApp/harvest/pull/3608))
- Capitalize the Grafana Cisco folder ([#3612](https://github.com/NetApp/harvest/pull/3612))
- Improve Grafana import logging (#3620) ([#3630](https://github.com/NetApp/harvest/pull/3630))
- Update instance generation in quota plugin ([#3637](https://github.com/NetApp/harvest/pull/3637))
- Remove unused errors ([#3640](https://github.com/NetApp/harvest/pull/3640))

### Miscellaneous
- Merge release/25.02.0 into main ([#3474](https://github.com/NetApp/harvest/pull/3474))
- Bump go.mod ([#3476](https://github.com/NetApp/harvest/pull/3476))
- Update all dependencies ([#3477](https://github.com/NetApp/harvest/pull/3477))
- Update all dependencies ([#3487](https://github.com/NetApp/harvest/pull/3487))
- Update all dependencies ([#3499](https://github.com/NetApp/harvest/pull/3499))
- Update all dependencies ([#3508](https://github.com/NetApp/harvest/pull/3508))
- Update astral-sh/setup-uv digest to a4fd982 ([#3521](https://github.com/NetApp/harvest/pull/3521))
- Update astral-sh/setup-uv digest to 2269511 ([#3525](https://github.com/NetApp/harvest/pull/3525))
- Update all dependencies ([#3539](https://github.com/NetApp/harvest/pull/3539))
- Update all dependencies ([#3548](https://github.com/NetApp/harvest/pull/3548))
- Fix formatting ([#3552](https://github.com/NetApp/harvest/pull/3552))
- Update astral-sh/setup-uv digest to 594f292 ([#3556](https://github.com/NetApp/harvest/pull/3556))
- Update astral-sh/setup-uv digest to fb3a0a9 ([#3568](https://github.com/NetApp/harvest/pull/3568))
- Update all dependencies ([#3576](https://github.com/NetApp/harvest/pull/3576))
- Update all dependencies ([#3595](https://github.com/NetApp/harvest/pull/3595))
- Update all dependencies ([#3615](https://github.com/NetApp/harvest/pull/3615))


### :hammer: CI
- The issue burn-down list should ignore status/done issues ([#3459](https://github.com/NetApp/harvest/pull/3459))
- Bump go ([#3504](https://github.com/NetApp/harvest/pull/3504))
- style: format match gjson file ([#3506](https://github.com/NetApp/harvest/pull/3506))
- Bump dependencies ([#3517](https://github.com/NetApp/harvest/pull/3517))
- Update config path ([#3523](https://github.com/NetApp/harvest/pull/3523))
- Update rest role in cert ([#3527](https://github.com/NetApp/harvest/pull/3527))
- Upgrade golangci-lint to v2.0.1 ([#3529](https://github.com/NetApp/harvest/pull/3529))
- Bump go ([#3543](https://github.com/NetApp/harvest/pull/3543))
- Fix lint warnings ([#3557](https://github.com/NetApp/harvest/pull/3557))
- Update promtool path ([#3571](https://github.com/NetApp/harvest/pull/3571))
- Handle ems_events error for ZAPI datacenter ([#3597](https://github.com/NetApp/harvest/pull/3597))
- Bump go ([#3602](https://github.com/NetApp/harvest/pull/3602))
- Handle duplicated definition of symbol dlopen error ([#3605](https://github.com/NetApp/harvest/pull/3605))

---

## 25.02.0 / 2025-02-11 Release
:pushpin: Highlights of this major release include:
## :star: New Features

- :star: The Volume dashboard was updated to clarify that volume latencies are missing some latencies from NAS protocols. Use the workload volume metrics in the QoS row for a more detailed breakdown. Thanks to MatthiasS for reporting.

- All Harvest dashboards default to Datacenter=All instead of the first datacenter in the list. Thanks to @roybatty2019 for reporting.

- Harvest provides a [FIPS 140-3 compliant](https://go.dev/doc/security/fips140) container image, available as a separate image at `ghcr.io/netapp/harvest:25.02.0-1-fips`.

- :ear_of_rice: Harvest `bin/grafana import`
  - Supports nested Grafana folders. Thanks to @IvanZenger for reporting.
  - Supports setting variables' default values during import. See [#3384](https://github.com/NetApp/harvest/issues/3384) for details. Thanks to @mamoep for reporting.

- Harvest collects shelf firmware versions and shows them in the Shelf dashboard, Module row. Thanks to @summertony15 for reporting.

- :star: Several of the existing dashboards include new panels in this release:
  - The `Disk` dashboard includes a `Top Disk and Tape Drives Throughput by Host Adapter` panel. Thanks to Amir for reporting.
  - The `Datacenter` and `Data Protection` dashboards were updated with data protection buckets and policy rows.

- The volumes templates exclude transient volumes by default. Thanks to Yann for reporting.

- Harvest collects rewind context (rwctx) metrics for ONTAP 9.16.0 and later. Thanks to @shawnahall71 for reporting.

- :closed_book: Documentation additions
  - Document [Podman Quadlet](https://netapp.github.io/harvest/nightly/install/quadlet/) as a deployment option. Thanks to ttlexceeded for reporting.
  - Describe how to use a [Go binary as a credential script](https://github.com/NetApp/harvest/discussions/3380) for Harvest. Thanks to AdiZ for reporting.

## :rocket: Performance Improvements

- RestPerf collector uses less memory by streaming results.

In case you missed the previous `24.11.1` dot release, here are the features included in it:

## :rocket: Performance Improvements in 24.11.1

- Significant memory footprint improvements for the REST collector. More details [here](https://github.com/NetApp/harvest/pull/3310#issue-2676698124). Thanks to @Ryan for reporting it.
- Reduced memory footprint by using streaming in the REST collector.

## :star: New Features in 24.11.1

- Harvest supports Top files metrics collection. More details [here](https://github.com/NetApp/harvest/discussions/3130).
- Volume and Cluster tags are supported via Volume and Cluster dashboards.
- Field Replaceable Unit (FRU) details have been added to the power dashboard.
- Track ONTAP image update progress for a cluster via the Cluster dashboard. Thanks to @knappmi for reporting it.
- `prom_port` is now supported within the poller. More details [here](https://netapp.github.io/harvest/nightly/prometheus-exporter/#per-poller-prom_port).
- We've fixed an intermittent latency/operations spike issue in the plugin-generated Harvest performance metrics. Thanks to @wooyoungAhn for reporting it.

## Announcements

:bangbang: **IMPORTANT** Harvest version 25.02.0 disables the out-of-the-box `Qtree` templates because of reported ONTAP slowdowns when collecting a large number of qtree objects. If you want to enable the Qtree templates, please see these [instructions](https://github.com/NetApp/harvest/discussions/3446).

:bangbang: **IMPORTANT** Harvest version 25.02.0 removes the `WorkloadDetail` and `WorkloadDetailVolume` templates and all dashboard panels that use them. These templates are removed because they are expensive to collect and currently there is no way to collect them from ONTAP without introducing an unacceptable amount of skew in the results. See [#3423](https://github.com/NetApp/harvest/issues/3423) for details.

:bangbang: **IMPORTANT** If using Docker Compose and you want to keep your historical Prometheus data, please read [how to migrate your Prometheus volume](https://github.com/NetApp/harvest/blob/main/docs/MigratePrometheusDocker.md)

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the 'bin/harvest grafana import' CLI, from the Grafana UI, or from the 'Maintenance > Reset Harvest Dashboards' button in NAbox3. For NAbox4, this step is not needed.

## Known Issues

:bulb: **IMPORTANT** FSx ZapiPerf workload collector fails to collect metrics, please use RestPerf instead.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards this release:

@Falcon667, @IvanZenger, @cheese1, @embusalacchi, @mamoep, @roybatty2019, @summertony15, AdiZ, Amir, MatthiasS, Yann, ttlexceeded

:seedling: This release includes 18 features, 23 bug fixes, 8 documentation, 1 performance, 1 refactoring, 6 miscellaneous, and 12 ci pull requests.

### :rocket: Features
- Hide Transient Volumes ([#3337](https://github.com/NetApp/harvest/pull/3337))
- Adding Ifgrp Api To Fetch Ifgrp Labels In Net_port ([#3342](https://github.com/NetApp/harvest/pull/3342))
- Adding Rwctx Template For Restperf In 9.16.0 ([#3349](https://github.com/NetApp/harvest/pull/3349))
- Add Disk And Tape Drives Throughput By Host Adapter ([#3372](https://github.com/NetApp/harvest/pull/3372))
- Adding The Bucket And Policy Rest Template ([#3374](https://github.com/NetApp/harvest/pull/3374))
- Include Lun And Namespace Templates In Keyperf ([#3379](https://github.com/NetApp/harvest/pull/3379))
- Dashboard Variables Set Default Value On Import ([#3399](https://github.com/NetApp/harvest/pull/3399))
- Optimize Workload_detail And Workload_detail_volume Through Delay Center Filter ([#3406](https://github.com/NetApp/harvest/pull/3406))
- Handled Empty Admin Svm In Plugin Call ([#3410](https://github.com/NetApp/harvest/pull/3410))
- Support Metric For Module Type In Frus ([#3411](https://github.com/NetApp/harvest/pull/3411))
- Harvest Grafana Import Should Support Nested Grafana Folders ([#3417](https://github.com/NetApp/harvest/pull/3417))
- Remove Workload Detail Templates ([#3433](https://github.com/NetApp/harvest/pull/3433))
- Add Streaming To Keyperf Collector ([#3435](https://github.com/NetApp/harvest/pull/3435))
- Negative Metrics Spike Handling ([#3439](https://github.com/NetApp/harvest/pull/3439))
- Disabled Qtree Perf Template And Update Docs ([#3445](https://github.com/NetApp/harvest/pull/3445))
- Harvest Dashboards Should Default To Datacenter=All ([#3448](https://github.com/NetApp/harvest/pull/3448))
- Update Qos Row In Volume Dashboard ([#3453](https://github.com/NetApp/harvest/pull/3453))
- Improve Cp Summary In Disk Dashboard ([#3456](https://github.com/NetApp/harvest/pull/3456))

### :bug: Bug Fixes
- Include Instances Generated By Inbuilt Plugins In Plugininstances Log ([#3343](https://github.com/NetApp/harvest/pull/3343))
- Handled Duplicate Key In Securityauditdestination ([#3348](https://github.com/NetApp/harvest/pull/3348))
- Harvest Permissions Should Include Fru ([#3354](https://github.com/NetApp/harvest/pull/3354))
- No Instances Handling In Rest Collector ([#3358](https://github.com/NetApp/harvest/pull/3358))
- Rest No Instance Handling ([#3360](https://github.com/NetApp/harvest/pull/3360))
- Don't Clear Performance Volume Cache When There Is An Error ([#3361](https://github.com/NetApp/harvest/pull/3361))
- Failed To Find Scanner Instance In Cache Zapiperf ([#3366](https://github.com/NetApp/harvest/pull/3366))
- Installation Broken On Debian 11 Bullseye ([#3368](https://github.com/NetApp/harvest/pull/3368))
- Changed Var Label To Ne Null From Empty ([#3385](https://github.com/NetApp/harvest/pull/3385))
- Update Snapshot Policy Endpoint ([#3391](https://github.com/NetApp/harvest/pull/3391))
- Update Export Rule Endpoint ([#3392](https://github.com/NetApp/harvest/pull/3392))
- Upgrade Golang.org/X/Net Due To Dependabot Alert ([#3395](https://github.com/NetApp/harvest/pull/3395))
- Update Export Rule Endpoint ([#3396](https://github.com/NetApp/harvest/pull/3396))
- Remove Redundant Label From Node Template ([#3404](https://github.com/NetApp/harvest/pull/3404))
- Enable Request/Response Logging For Restperf ([#3408](https://github.com/NetApp/harvest/pull/3408))
- Disable Cache To Avoid Cache Poisoning Attack ([#3409](https://github.com/NetApp/harvest/pull/3409))
- Duplicate Time Series In Volume Dashboard ([#3418](https://github.com/NetApp/harvest/pull/3418))
- Update Tr Link In Security Dashboard ([#3419](https://github.com/NetApp/harvest/pull/3419))
- Typo ([#3425](https://github.com/NetApp/harvest/pull/3425))
- Handle Only Labels In Zapi Snapshotpolicy ([#3444](https://github.com/NetApp/harvest/pull/3444))
- "Top Ethernet Ports By Utilization %" Panel Legend Should Not In… ([#3451](https://github.com/NetApp/harvest/pull/3451))
- Handle Cp Labels In Dashboard ([#3455](https://github.com/NetApp/harvest/pull/3455))
- Cloud Target Template Should Not Export Access_Key ([#3470](https://github.com/NetApp/harvest/pull/3470))

### :closed_book: Documentation
- Fix Release Announcements ([#3330](https://github.com/NetApp/harvest/pull/3330))
- Keyperf Documentation ([#3345](https://github.com/NetApp/harvest/pull/3345))
- Updating Doc For Custom.yaml ([#3352](https://github.com/NetApp/harvest/pull/3352))
- Rest Endpoint Permissions ([#3359](https://github.com/NetApp/harvest/pull/3359))
- Add Go Binary Steps For Credential Script ([#3381](https://github.com/NetApp/harvest/pull/3381))
- Fix Alignment Of Template ([#3421](https://github.com/NetApp/harvest/pull/3421))
- Document Podman Quadlet As A Deployment Option ([#3442](https://github.com/NetApp/harvest/pull/3442))
- Add Description About Cp In Disk Dashboard ([#3454](https://github.com/NetApp/harvest/pull/3454))

### :zap: Performance
- Restperf Should Stream Results While Parsing ([#3356](https://github.com/NetApp/harvest/pull/3356))

### Refactoring
- Remove Openapi Dependency ([#3436](https://github.com/NetApp/harvest/pull/3436))

### Miscellaneous
- Merge 24.11.1 To Main ([#3328](https://github.com/NetApp/harvest/pull/3328))
- Update Module Github.com/Shirou/Gopsutil/V4 To V4.24.11 ([#3347](https://github.com/NetApp/harvest/pull/3347))
- Update All Dependencies ([#3365](https://github.com/NetApp/harvest/pull/3365))
- Update All Dependencies ([#3407](https://github.com/NetApp/harvest/pull/3407))
- Update All Dependencies ([#3450](https://github.com/NetApp/harvest/pull/3450))
- Update All Dependencies ([#3457](https://github.com/NetApp/harvest/pull/3457))

### :hammer: CI
- Fix Go Vet Errors ([#3331](https://github.com/NetApp/harvest/pull/3331))
- Add Misspell, Nakedret, And Predeclared Linter ([#3350](https://github.com/NetApp/harvest/pull/3350))
- Bump Go ([#3351](https://github.com/NetApp/harvest/pull/3351))
- Add Role Permissions Validation ([#3375](https://github.com/NetApp/harvest/pull/3375))
- Enable More Linters ([#3413](https://github.com/NetApp/harvest/pull/3413))
- Add Workflow Permissions At Codeql Recommendation ([#3426](https://github.com/NetApp/harvest/pull/3426))
- Add Workflow Permissions At Codeql Recommendation ([#3427](https://github.com/NetApp/harvest/pull/3427))
- Pin Actions To Sha ([#3428](https://github.com/NetApp/harvest/pull/3428))
- Cheese1 Has Signed The Ccla ([#3429](https://github.com/NetApp/harvest/pull/3429))
- Bump Go ([#3430](https://github.com/NetApp/harvest/pull/3430))
- Ci-Local Requires Passing Admin Argument ([#3431](https://github.com/NetApp/harvest/pull/3431))
- Bump Go ([#3452](https://github.com/NetApp/harvest/pull/3452))

---

## 24.11.1 / 2024-11-25 Release
:pushpin: Highlights of this major release include:
## :rocket: Performance Improvements

- Significant memory footprint improvements for the REST collector. More details [here](https://github.com/NetApp/harvest/pull/3310#issue-2676698124). Thanks to @Ryan for reporting it.
- Reduced memory footprint by using streaming in the REST collector.

## :star: New Features

- Harvest supports Top files metrics collection. More details [here](https://github.com/NetApp/harvest/discussions/3130).
- Volume and Cluster tags are supported via Volume and Cluster dashboards.
- Field Replaceable Unit (FRU) details have been added to the power dashboard.
- Track ONTAP image update progress for a cluster via the Cluster dashboard. Thanks to @knappmi for reporting it.
- `prom_port` is now supported within the poller. More details [here](https://netapp.github.io/harvest/nightly/prometheus-exporter/#per-poller-prom_port).
- We've fixed an intermittent latency/operations spike issue in the plugin-generated Harvest performance metrics. Thanks to @wooyoungAhn for reporting it.

## Announcements

:bangbang: **IMPORTANT** NetApp moved their communities from Slack to [Discord](https://discord.gg/ZmmWPHTBHw), please join us [there](https://discordapp.com/channels/855068651522490400/1001963189124206732)!

:bangbang: **IMPORTANT** If using Docker Compose and you want to keep your historical Prometheus data, please
read [how to migrate your Prometheus volume](https://github.com/NetApp/harvest/blob/main/docs/MigratePrometheusDocker.md)

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the 'bin/harvest grafana import' CLI, from the Grafana UI, or from the 'Maintenance > Reset Harvest Dashboards' button in NAbox3. For NAbox4, this step is not needed.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards
this release:

@70tas, @BrendonA667, @Falcon667, @Mark Jordan, @Paqui, @Ryan, @cashnmoney, @ceojinhak, @ekolove, @knappmi, @wooyoungAhn

:seedling: This release includes 14 features, 8 bug fixes, 2 documentation, 3 performance, 1 testing, 1 styling, 7 refactoring, 2 miscellaneous, and 3 ci pull requests.

### :rocket: Features
- Add Tags To The Volume And Cluster Dashboards ([#3273](https://github.com/NetApp/harvest/pull/3273))
- Harvest Should Request Cluster Version Once ([#3274](https://github.com/NetApp/harvest/pull/3274))
- Top Files Collection ([#3279](https://github.com/NetApp/harvest/pull/3279))
- Enable Iface And Recvcheck Linters ([#3280](https://github.com/NetApp/harvest/pull/3280))
- Harvest Should Support Per-Poller Prom_ports ([#3281](https://github.com/NetApp/harvest/pull/3281))
- Harvest Should Log Number Of Renderedbytes For Each Collector ([#3282](https://github.com/NetApp/harvest/pull/3282))
- Asa R2 Should Use Keyperf Instead Of Restperf ([#3289](https://github.com/NetApp/harvest/pull/3289))
- Add Top Files Panels In Volume Dashboard ([#3292](https://github.com/NetApp/harvest/pull/3292))
- Adding The Ems Doc Link In The Health Dashboard Table ([#3295](https://github.com/NetApp/harvest/pull/3295))
- Add Dimm Panels In Power Dashboard ([#3296](https://github.com/NetApp/harvest/pull/3296))
- Adding Is_space_enforcement_logical, Is_space_reporting_logical… ([#3301](https://github.com/NetApp/harvest/pull/3301))
- Harvest Should Monitor `Wafl.dir.size.warning` ([#3304](https://github.com/NetApp/harvest/pull/3304))
- Add Flexcache Keyperf Template ([#3309](https://github.com/NetApp/harvest/pull/3309))
- Add Top Metrics Plugin To Keyperf ([#3315](https://github.com/NetApp/harvest/pull/3315))

### :bug: Bug Fixes
- Set Dashboard Variable To Refresh To Time Range Change. ([#3269](https://github.com/NetApp/harvest/pull/3269))
- Correct The Mtu Unit In Network Dashboard ([#3278](https://github.com/NetApp/harvest/pull/3278))
- Zapi Collection ([#3285](https://github.com/NetApp/harvest/pull/3285))
- Metroclustercheck Collector Should Report Standby When Metroclus… ([#3287](https://github.com/NetApp/harvest/pull/3287))
- Missing Volumes After Vol Move ([#3312](https://github.com/NetApp/harvest/pull/3312))
- Metroclustercheck Collector Should Report "No Instances" ([#3314](https://github.com/NetApp/harvest/pull/3314))
- Panic If No Volumes Have Analytics Enabled ([#3323](https://github.com/NetApp/harvest/pull/3323))
- Partial Aggregation Handling In Plugins ([#3324](https://github.com/NetApp/harvest/pull/3324))

### :closed_book: Documentation
- Update Top Clients Doc ([#3311](https://github.com/NetApp/harvest/pull/3311))
- Harvest Should Include Network Port Ifgrp Permissions ([#3318](https://github.com/NetApp/harvest/pull/3318))

### :zap: Performance
- Reduce The Memory Footprint Of Rest Collector ([#3303](https://github.com/NetApp/harvest/pull/3303))
- Add Streaming To Rest Collector ([#3305](https://github.com/NetApp/harvest/pull/3305))
- Improve Memory And Cpu Performance Of Rest Collector ([#3310](https://github.com/NetApp/harvest/pull/3310))

### :wrench: Testing
- Sort Exporters For Deterministic Tests ([#3290](https://github.com/NetApp/harvest/pull/3290))

### Styling
- Fix Logs ([#3307](https://github.com/NetApp/harvest/pull/3307))

### Refactoring
- Remove Extra Log ([#3257](https://github.com/NetApp/harvest/pull/3257))
- Remove Env Logging ([#3277](https://github.com/NetApp/harvest/pull/3277))
- Simplify Negotiateontapapi ([#3288](https://github.com/NetApp/harvest/pull/3288))
- Keyperf Node Template Should Match Restperf Object Name ([#3298](https://github.com/NetApp/harvest/pull/3298))
- Remove Uses Of `Nolint:gocritic` ([#3299](https://github.com/NetApp/harvest/pull/3299))
- Remove Unused Method In Rest Collector ([#3308](https://github.com/NetApp/harvest/pull/3308))
- Sync Template Names For Keyperf ([#3316](https://github.com/NetApp/harvest/pull/3316))

### Miscellaneous
- Update All Dependencies ([#3275](https://github.com/NetApp/harvest/pull/3275))
- Update Chizkiyahu/Delete-Untagged-Ghcr-Action Action To V5 ([#3300](https://github.com/NetApp/harvest/pull/3300))

### :hammer: CI
- Bump Go ([#3270](https://github.com/NetApp/harvest/pull/3270))
- Lint Errors ([#3276](https://github.com/NetApp/harvest/pull/3276))
- Ignore Volume_top_files_ Counters ([#3293](https://github.com/NetApp/harvest/pull/3293))

---


## 24.11.0 / 2024-11-06 Release
:pushpin: Highlights of this major release include:

- :gem: New dashboards:
- SnapMirror Destinations Dashboard which displays relationship details from the destination perspective.
- Vscan Dashboard which shows SVM-level and connection scanner details.


- :star: Several of the existing dashboards include new panels in this release:
- SnapMirror dashboard now includes relationship details from the source perspective and has been renamed to "ONTAP: SnapMirror Sources".
- Health Dashboard's emergency events panel now includes all emergency EMS events from the last 24 hours.
- Network Dashboard
  - Includes Link Aggregation Group (LAG) metrics
  - Adds Ethernet port details
- s3 Object Storage dashboard includes panels for s3 metrics for SVM.
- Tenant Dashboard
  - Adds Tenant/Bucket Capacity Growth Chart
  - Includes average size per object details for each bucket
- Metadata Dashboard includes a panel displaying the number of instances collected.
- Power Dashboard includes a new "Average Power Consumption (kWh) Over Last Hour" panel.
- SVM Dashboard now features panels for logical space and physical space at the SVM level.
- Volume Deep Dive dashboard includes "Other IOPs" panel.

- :rocket: Performance Improvements:
- Reduced memory footprint by optimizing memory allocations when serving metrics.
- Reduced API calls when using the RestPerf collector.

- Harvest supports Top clients metrics collection. [More details](https://netapp.github.io/harvest/latest/plugins/#volumetopclients).
- Harvest supports recording and replaying HTTP requests.
- Harvest now provides a FIPS-compliant container image, available as a separate image (ghcr.io/netapp/harvest:24.08.0-1-fips).
- Grafana import allows rewriting the cluster label during import.

## Announcements

:bangbang: **IMPORTANT** NetApp moved their communities from Slack to [Discord](https://discord.gg/ZmmWPHTBHw), please join us [there](https://discordapp.com/channels/855068651522490400/1001963189124206732)!

:bangbang: **IMPORTANT** If using Docker Compose and you want to keep your historical Prometheus data, please
read [how to migrate your Prometheus volume](https://github.com/NetApp/harvest/blob/main/docs/MigratePrometheusDocker.md)

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the 'bin/harvest grafana import' CLI, from the Grafana UI, or from the 'Maintenance > Reset Harvest Dashboards' button in NAbox.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards
this release:

- @ofu48167
- @WayneShen2
- @T1r0l
- @Daniel-Vaz
- @razaahmed
- @gaow1423
- @BrendonA667
- @70tas
- @annapook-netapp
- @buller7929
- @florent4155
- @heinowalther
- @db-wally007
- @embusalacchi

:seedling: This release includes 36 features, 24 bug fixes, 7 documentation, 7 performance, 1 testing, 3 styling, 5 refactoring, 9 miscellaneous, and 15 ci pull requests.

### :rocket: Features
- Tenant Dashboard Buckets Panel Should Include ([#3101](https://github.com/NetApp/harvest/pull/3101))
- Use Docker Buildx Secret For Token ([#3108](https://github.com/NetApp/harvest/pull/3108))
- Enable Pprof Endpoints On Localhost ([#3110](https://github.com/NetApp/harvest/pull/3110))
- Generate Fips Compliant Container Image For Harvest ([#3113](https://github.com/NetApp/harvest/pull/3113))
- Support Ifgroup Level Throughput Metrics ([#3117](https://github.com/NetApp/harvest/pull/3117))
- Harvest Should Include A Vscan Dashboard ([#3121](https://github.com/NetApp/harvest/pull/3121))
- Vscan Dashboard Should Include Topk ([#3127](https://github.com/NetApp/harvest/pull/3127))
- Top Clients Metrics Collection ([#3132](https://github.com/NetApp/harvest/pull/3132))
- Adding Panels For Ontaps3svm Object ([#3134](https://github.com/NetApp/harvest/pull/3134))
- Grafana Import Should Allow Rewriting Cluster Label ([#3135](https://github.com/NetApp/harvest/pull/3135))
- Replace Zerolog With Slog ([#3146](https://github.com/NetApp/harvest/pull/3146))
- Harvest Should Include Time-Series Panels For Tenants And Buckets ([#3147](https://github.com/NetApp/harvest/pull/3147))
- Send The Harvest Version To Ontap ([#3152](https://github.com/NetApp/harvest/pull/3152))
- Replace Zerolog With Slog ([#3164](https://github.com/NetApp/harvest/pull/3164))
- Add Documentation For Plugin-Generated Metrics And Enable Ci ([#3169](https://github.com/NetApp/harvest/pull/3169))
- Add Instances Collected Panel To Metadata Dashboard ([#3178](https://github.com/NetApp/harvest/pull/3178))
- Harvest Should Use Slogs Text Format By Default ([#3179](https://github.com/NetApp/harvest/pull/3179))
- Add "Average Power Consumption (Kwh) Over Last Hour" Panel To Power Dashboard ([#3180](https://github.com/NetApp/harvest/pull/3180))
- Replacing connector webhook with MS workflow ([#3183](https://github.com/NetApp/harvest/pull/3183))
- Handle Url Limit In Rest ([#3186](https://github.com/NetApp/harvest/pull/3186))
- Keyperf Collector Templates ([#3194](https://github.com/NetApp/harvest/pull/3194))
- Harvest Rest And Restperf Collectors Should Support Batching ([#3195](https://github.com/NetApp/harvest/pull/3195))
- Add Top Svm By Space In Svm Dashboard ([#3200](https://github.com/NetApp/harvest/pull/3200))
- All Harvest Dashboards Should Include Tags ([#3202](https://github.com/NetApp/harvest/pull/3202))
- Support Destination/Source Level View - Parity With Sm ([#3204](https://github.com/NetApp/harvest/pull/3204))
- Add Other Ops Panel In Volume Deep Dive Dashboard ([#3209](https://github.com/NetApp/harvest/pull/3209))
- Add Nfs Templates For Keyperf Collector ([#3215](https://github.com/NetApp/harvest/pull/3215))
- Adding Snapmirror Sources dashboard - 1 ([#3216](https://github.com/NetApp/harvest/pull/3216))
- Keyperf Collector Templates ([#3219](https://github.com/NetApp/harvest/pull/3219))
- Adding Ethernet Port Table From Netport Template ([#3221](https://github.com/NetApp/harvest/pull/3221))
- Fail Ci When There Are Errors In Prometheus Or Grafana ([#3232](https://github.com/NetApp/harvest/pull/3232))
- Log Cluster Name And Version With Poller Metadata ([#3234](https://github.com/NetApp/harvest/pull/3234))
- Harvest Should Support Recording And Replaying Http Requests ([#3235](https://github.com/NetApp/harvest/pull/3235))
- Add Emergency Events To Health Dashboard ([#3238](https://github.com/NetApp/harvest/pull/3238))
- Add Keyperf Metric Docs ([#3240](https://github.com/NetApp/harvest/pull/3240))
- Improve Harvest Memory Logging ([#3244](https://github.com/NetApp/harvest/pull/3244))
- Doctor should handle embedded exporters ([#3258](https://github.com/NetApp/harvest/pull/3258))

### :bug: Bug Fixes
- Handled Non Exported Qtrees In Template ([#3105](https://github.com/NetApp/harvest/pull/3105))
- Handled Nameservices In Svm Zapi Plugin ([#3124](https://github.com/NetApp/harvest/pull/3124))
- Fix Disk Count In Disk Dashboard ([#3126](https://github.com/NetApp/harvest/pull/3126))
- Handled Quota Index Key In Rest Template With Tests ([#3128](https://github.com/NetApp/harvest/pull/3128))
- Vscan Panels Throws 422 Error ([#3133](https://github.com/NetApp/harvest/pull/3133))
- Correcting The Alert Rule Expression For Required Labels ([#3143](https://github.com/NetApp/harvest/pull/3143))
- Svm Dashboard - Volume Capacity Row Ordering ([#3158](https://github.com/NetApp/harvest/pull/3158))
- Fsa History Data Should Work When Multi Select ([#3159](https://github.com/NetApp/harvest/pull/3159))
- Do Not Log Stdout When A Credential Script Fails ([#3163](https://github.com/NetApp/harvest/pull/3163))
- Remove '*' As 'All' Option In Workload Dropdown On Workload Dashboard ([#3165](https://github.com/NetApp/harvest/pull/3165))
- `Bin/Harvest Rest` Should Read Credentials Before Fetching Data ([#3166](https://github.com/NetApp/harvest/pull/3166))
- Remove Embedded Shelf Power From Total Power In Series Panel To Match Stats Panel ([#3167](https://github.com/NetApp/harvest/pull/3167))
- Volume_aggr_labels Should Not Include Uuid Label ([#3171](https://github.com/NetApp/harvest/pull/3171))
- Add Embedded Shelf Type For Power Calculation ([#3174](https://github.com/NetApp/harvest/pull/3174))
- Using Instancename Instead Of Volname In Fabricpool Perf ([#3175](https://github.com/NetApp/harvest/pull/3175))
- Correct Failed State In Workflow ([#3190](https://github.com/NetApp/harvest/pull/3190))
- Handled Flexgroup Based On Volume Config Call ([#3199](https://github.com/NetApp/harvest/pull/3199))
- Filter By Svm, Volume In Sm Destination Dashboard ([#3220](https://github.com/NetApp/harvest/pull/3220))
- Remove _Labels From Metric Docs ([#3222](https://github.com/NetApp/harvest/pull/3222))
- Update Datacenter And Cluster Variables In Dashboards ([#3227](https://github.com/NetApp/harvest/pull/3227))
- Don't Double Export Aggregate Efficiency Metrics ([#3230](https://github.com/NetApp/harvest/pull/3230))
- Update Keyperf Collector Static Counter File Path ([#3241](https://github.com/NetApp/harvest/pull/3241))
- Fix Numbering In Quickstart ([#3249](https://github.com/NetApp/harvest/pull/3249))
- Fix Value Mapping In Tenant Dashboard ([#3253](https://github.com/NetApp/harvest/pull/3253))
- Rename volume latency in keyperf ([#3261](https://github.com/NetApp/harvest/pull/3261))

### :closed_book: Documentation
- Fix Typo In Docs ([#3112](https://github.com/NetApp/harvest/pull/3112))
- Clarify Ipv6 Support ([#3119](https://github.com/NetApp/harvest/pull/3119))
- Topclients Plugin Document ([#3151](https://github.com/NetApp/harvest/pull/3151))
- Add More Credential Script Troubleshooting Steps ([#3154](https://github.com/NetApp/harvest/pull/3154))
- Remove Qos Service Latency Counter From Metric Docs ([#3188](https://github.com/NetApp/harvest/pull/3188))
- Add Space To Datacenter Dashboard Title ([#3225](https://github.com/NetApp/harvest/pull/3225))
- Update Release Months To Match Harvest Release Cadence ([#3236](https://github.com/NetApp/harvest/pull/3236))
- Update KeyPerf metric docs ([#3260](https://github.com/NetApp/harvest/pull/3260))

### :zap: Performance
- Reduce Allocs When Reading Credential Files ([#3111](https://github.com/NetApp/harvest/pull/3111))
- Reduce Allocs In Prometheus Render ([#3168](https://github.com/NetApp/harvest/pull/3168))
- Reduce Allocations When Serving Prometheus Metrics ([#3172](https://github.com/NetApp/harvest/pull/3172))
- Reduce Poller Footprint By Not Collecting Smb Histogram Metrics … ([#3177](https://github.com/NetApp/harvest/pull/3177))
- Restperf Collectors Should Only Run Pollinstance For Workloads ([#3207](https://github.com/NetApp/harvest/pull/3207))
- Reduce Allocs When Serving Metrics ([#3208](https://github.com/NetApp/harvest/pull/3208))
- Reduce Allocs When Rendering Metrics ([#3214](https://github.com/NetApp/harvest/pull/3214))

### :wrench: Testing
- Add Authtoken With Password Testcase ([#3176](https://github.com/NetApp/harvest/pull/3176))

### Styling
- Ensure Slogging Uses Attributes Only ([#3197](https://github.com/NetApp/harvest/pull/3197))
- Add Debug Logs For Volume Plugin ([#3233](https://github.com/NetApp/harvest/pull/3233))
- Bring Harvest.cue Up To Date ([#3256](https://github.com/NetApp/harvest/pull/3256))

### Refactoring
- Rename Volumetopclients Maxvolumecount To Max_volumes ([#3141](https://github.com/NetApp/harvest/pull/3141))
- Improve Logging ([#3182](https://github.com/NetApp/harvest/pull/3182))
- Change Auth Credential Script Logging To Debug ([#3191](https://github.com/NetApp/harvest/pull/3191))
- Improve Slog Error Logging ([#3198](https://github.com/NetApp/harvest/pull/3198))
- Truncate Href When Logging ([#3245](https://github.com/NetApp/harvest/pull/3245))

### Miscellaneous
- Update All Dependencies ([#3096](https://github.com/NetApp/harvest/pull/3096))
- Merge Release/24.08.0 Into Main ([#3099](https://github.com/NetApp/harvest/pull/3099))
- Remove Zerolog Stack Calls ([#3118](https://github.com/NetApp/harvest/pull/3118))
- Update Module Github.com/Shirou/Gopsutil/V4 To V4.24.8 ([#3129](https://github.com/NetApp/harvest/pull/3129))
- Update All Dependencies ([#3139](https://github.com/NetApp/harvest/pull/3139))
- Remove Calls To Msgf("") ([#3144](https://github.com/NetApp/harvest/pull/3144))
- Update All Dependencies ([#3196](https://github.com/NetApp/harvest/pull/3196))
- Update Prometheus Version V2.55.0 (#3223) ([#3226](https://github.com/NetApp/harvest/pull/3226))
- Update Module Github.com/Shirou/Gopsutil/V4 To V4.24.10 ([#3250](https://github.com/NetApp/harvest/pull/3250))

### :hammer: CI
- Bump Go ([#3106](https://github.com/NetApp/harvest/pull/3106))
- Fix Container Image Push Order ([#3116](https://github.com/NetApp/harvest/pull/3116))
- Bump Go ([#3138](https://github.com/NetApp/harvest/pull/3138))
- Add Metric Generate Docs To Ci ([#3150](https://github.com/NetApp/harvest/pull/3150))
- Fix Ci Datacenter Name ([#3170](https://github.com/NetApp/harvest/pull/3170))
- Replacing Connector Webhook With Ms Workflow ([#3183](https://github.com/NetApp/harvest/pull/3183))
- Bump Gopsutil ([#3185](https://github.com/NetApp/harvest/pull/3185))
- Bump Go ([#3187](https://github.com/NetApp/harvest/pull/3187))
- Add Missing Dependency Of Purego ([#3189](https://github.com/NetApp/harvest/pull/3189))
- Add Trivy To Ci ([#3217](https://github.com/NetApp/harvest/pull/3217))
- Embusalacchi Has Signed The Ccla ([#3224](https://github.com/NetApp/harvest/pull/3224))
- Add Docker Login To Ci ([#3237](https://github.com/NetApp/harvest/pull/3237))
- Harvest Should Lint Metrics With Promtool ([#3246](https://github.com/NetApp/harvest/pull/3246))
- Use Zizmor To Find Security Issues In Github Actions Setups ([#3247](https://github.com/NetApp/harvest/pull/3247))
- Keyperf Collector Does Not Exist In Harvest Version 22.11 ([#3248](https://github.com/NetApp/harvest/pull/3248))

---

## 24.08.0 / 2024-08-12 Release

- :gem: Harvest dashboards now include links to other relevant dashboards. This makes it easier to navigate relationships between cluster objects.

- :star: Several of the existing dashboards include new panels in this release:
  - The Security dashboard shows SSL certificate expiration dates and warns if certificates are expiring soon. Prometheus alerts are created for expired certificates and certificates that will expire within the next month. Thanks to @timstiller for the suggestion.
  - The Volume and Aggregate dashboards include new panels showing inactive data trends. Thanks to @razaahmed for the suggestion.
  - The Workload dashboard includes panels showing the QoS percentage utilization at the policy level for shared QoS policies. Thanks to Rusty Brown for the suggestion.
  - The Datacenter dashboard includes the number of Qtrees, Quotas, and Workloads in the Object Count panel.
  - The Aggregate dashboard now includes topk timeseries.
  - The Metadata dashboard now includes a stats panel showing the number of failed collectors. Thanks to @mamoep for the suggestion.
  - The Metadata dashboard Pollers table includes the resident set size of each poller process.
  - The StorageGRID Tenant dashboard now includes an "average size per object" column in the Tenant Quota panel. Thanks to @ofu48167 for the contribution.

- :ear_of_rice: Quotas and Qtrees templates are separated into individual templates instead of being combined as in earlier versions of Harvest.

- The ChangeLog plugin monitors metric value changes in addition to label changes. Thanks to @pilot7777 for the suggestion.

- Harvest collects quotas even when there are no qtrees. Thanks to @qrm1982 for reporting.

- The StorageGRID collector supports single sign-on via a credential script auth token. Thanks to @santosh725 for suggesting.

- Harvest supports OAuth 2.0 ONTAP collectors via a credential script auth token.

- Harvest handles lun and namespace metrics with simple names.

- Harvest collects `virtual_used` and `virtual_used_percent` metrics from volumes via REST on ONTAP versions 9.14.1+

- Prometheus metrics retention has been increased to one year in the Docker compose workflow.

- Harvest creates resolution metrics for health alerts. Thanks to @faguayot for suggesting.

- Pollers report their status as the `poller_status` in native and container environments.

- Grafana import allows you to specify a custom all value when importing. Thanks to ChrisGautcher for the suggestion.

- Harvest includes remediation steps for EMS active sync events in the [EMS alert runbook](https://netapp.github.io/harvest/latest/resources/ems-alert-runbook/). Thanks to @Nikhita-13 for the contribution.

- `bin/harvest doctor` reports when exporters are missing

- Harvest allows exporting metrics without a prefix. This can be handy when collecting from a StorageGRID Prometheus instance. See the [storagegrid_metrics.yaml](https://github.com/NetApp/harvest/blob/4e3945c1f299f0f5e9cd0fff899c25121fd3599d/conf/storagegrid/11.6.0/storagegrid_metrics.yaml#L3) template for an example. Thanks to @Bhagyasri-Dolly for suggesting.

- :closed_book: Documentation Additions:
  - Harvest includes a new "Getting Started" tutorial. Thanks to MichelePardini for the suggestion.

## Announcements

:bangbang: **IMPORTANT** Harvest removed the Service Center row from the Workload dashboard and disabled collection of `qos_detail_service_time_latency` metrics. The metrics can be reenabled by setting `with_service_latency: true` in the WorkloadDetailVolume template file. See [#3015](https://github.com/NetApp/harvest/issues/3015) for details.

:bangbang: **IMPORTANT** If using Docker Compose and you want to keep your historical Prometheus data, please
read [how to migrate your Prometheus volume](https://github.com/NetApp/harvest/blob/main/docs/MigratePrometheusDocker.md)

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the 'bin/harvest grafana import' CLI, from the Grafana UI, or from the 'Maintenance > Reset Harvest Dashboards' button in NAbox3.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards
this release:

- @timstiller
- @razaahmed
- @mamoep
- @ofu48167
- @pilot7777
- @qrm1982
- @santosh725
- @faguayot
- @Nikhita
- @Bhagyasri
- @Falcon667
- RustyBrown
- ChrisGautcher
- MichelePardini

:seedling:
This release includes 40 features, 28 bug fixes, 13 documentation, 1 performance,
2 testing, 5 refactoring, 12 miscellaneous, and 11 ci pull requests.

### :rocket: Features
- Prometheus Should Retain Data For Up To One Year ([#2919](https://github.com/NetApp/harvest/pull/2919))
- Log Jitter During Best-Fit Template Loading ([#2920](https://github.com/NetApp/harvest/pull/2920))
- Add Failed Collectors Stats In Metadata Dashboard ([#2929](https://github.com/NetApp/harvest/pull/2929))
- Linking Dashboard Part-1 ([#2931](https://github.com/NetApp/harvest/pull/2931))
- Poller's Should Collect And Export Their Status And Memory ([#2944](https://github.com/NetApp/harvest/pull/2944))
- Include Rss In Poller Table Of Metadata Dashboard ([#2948](https://github.com/NetApp/harvest/pull/2948))
- Grafana Import Should Allow You To Specify A Custom All Value ([#2953](https://github.com/NetApp/harvest/pull/2953))
- Harvest Should Include Remediation Steps For Ems Active Sync Ev… ([#2963](https://github.com/NetApp/harvest/pull/2963))
- Linking Dashboards Part-2 ([#2968](https://github.com/NetApp/harvest/pull/2968))
- Support For Qos Percentage Utilization At Policy Level For Shared Qos Policies ([#2972](https://github.com/NetApp/harvest/pull/2972))
- Linking Dashboards Part-3 ([#2976](https://github.com/NetApp/harvest/pull/2976))
- Create Resolution Metrics For Health Alerts ([#2977](https://github.com/NetApp/harvest/pull/2977))
- Add Qtree,Quota,Workload Counts To Datacenter Dashboard ([#2978](https://github.com/NetApp/harvest/pull/2978))
- Harvest Should Track Poller Maxrss In Auto-Support ([#2982](https://github.com/NetApp/harvest/pull/2982))
- Add Topk To Aggregate Dashboard Timeseries Panels ([#2987](https://github.com/NetApp/harvest/pull/2987))
- Harvest Should Handle Lun And Namespace Metrics With Simple Names ([#2998](https://github.com/NetApp/harvest/pull/2998))
- Harvest Should Log Rss And Maxrss Every Hour ([#2999](https://github.com/NetApp/harvest/pull/2999))
- Implementing Certificate Expiry Detail In Security Dashboard ([#3000](https://github.com/NetApp/harvest/pull/3000))
- Remove Topk Vars From Storagegrid Dashboards ([#3002](https://github.com/NetApp/harvest/pull/3002))
- Add Inactive Data Metrics For Aggregate And Volume ([#3003](https://github.com/NetApp/harvest/pull/3003))
- Harvest Should Remove Service Center Metrics ([#3019](https://github.com/NetApp/harvest/pull/3019))
- Adding Quotas Detail In Asup ([#3020](https://github.com/NetApp/harvest/pull/3020))
- Harvest Should Allow Exporting Metrics Without A Prefix ([#3022](https://github.com/NetApp/harvest/pull/3022))
- Remove Service_time_latency Counter From Tests ([#3027](https://github.com/NetApp/harvest/pull/3027))
- Harvest Should Collect Virtual_used And Virtual_used_percent ([#3031](https://github.com/NetApp/harvest/pull/3031))
- Harvest Should Log Template Loading Errors ([#3036](https://github.com/NetApp/harvest/pull/3036))
- Enable Changelog Plugin To Monitor Metric Value Change ([#3041](https://github.com/NetApp/harvest/pull/3041))
- `--Debug` Cli Argument Should Enable Debug Logging ([#3043](https://github.com/NetApp/harvest/pull/3043))
- Harvest Should Support Storagegrid Credentials Script With Auth… ([#3048](https://github.com/NetApp/harvest/pull/3048))
- Harvest Doctor Should Report When Exporters Are Missing ([#3049](https://github.com/NetApp/harvest/pull/3049))
- Update Qtree Template Doc -  Collect Quotas When No Qtrees ([#3056](https://github.com/NetApp/harvest/pull/3056))
- Handled User/Group Quota In Historicallabels ([#3060](https://github.com/NetApp/harvest/pull/3060))
- Support Oauth2.0 Via Credential Script - Phase1 ([#3066](https://github.com/NetApp/harvest/pull/3066))
- Harvest Should Not Simultaneously Publish Quota Metrics From Qt… ([#3067](https://github.com/NetApp/harvest/pull/3067))
- Split Qtree/Quota Rest Templates ([#3068](https://github.com/NetApp/harvest/pull/3068))
- Adding Generated Instances/Metrics Count In Health Plugin Log ([#3074](https://github.com/NetApp/harvest/pull/3074))
- Health Dashboard Should Indicate When There Are No Events ([#3077](https://github.com/NetApp/harvest/pull/3077))
- Keyperfmetrics Collector Infrastructure ([#3078](https://github.com/NetApp/harvest/pull/3078))
- Adding Ut For Qtree Non Exported Case ([#3085](https://github.com/NetApp/harvest/pull/3085))
- Tenant Dashboard Should Include An `Average Size Per Object` Co… ([#3091](https://github.com/NetApp/harvest/pull/3091))

### :bug: Bug Fixes
- Zapi Rest Parity ([#2934](https://github.com/NetApp/harvest/pull/2934))
- Rest Templates Should Not Have Hyphon ([#2943](https://github.com/NetApp/harvest/pull/2943))
- Restore The Svm, Qtree, User, And Group Columns To The Quota Das… ([#2950](https://github.com/NetApp/harvest/pull/2950))
- Harvest Should Log Errors When Grafana Import Fails ([#2962](https://github.com/NetApp/harvest/pull/2962))
- Correct Details Folder Name While Import ([#2966](https://github.com/NetApp/harvest/pull/2966))
- Handling Min-Max In Gradient ([#2969](https://github.com/NetApp/harvest/pull/2969))
- Use Read/Write Data Due To Missing Historical Data In Dashboards ([#2979](https://github.com/NetApp/harvest/pull/2979))
- Fixing Non-Exported Flexgroup Instances Error ([#2980](https://github.com/NetApp/harvest/pull/2980))
- Add Shared Column For Workload Used % Tables ([#2986](https://github.com/NetApp/harvest/pull/2986))
- Qos Sequential Reads And Writes % Panels ([#2992](https://github.com/NetApp/harvest/pull/2992))
- Power Plugin Should Not Fail ([#2993](https://github.com/NetApp/harvest/pull/2993))
- Use Avg_over_time For Qos Used % ([#3004](https://github.com/NetApp/harvest/pull/3004))
- Add Missing Filtering For Metadata Dashboard ([#3005](https://github.com/NetApp/harvest/pull/3005))
- Handle Endpoints In Metric Doc ([#3011](https://github.com/NetApp/harvest/pull/3011))
- Handle Partial Aggregation For Flexgroup Perf Metrics ([#3018](https://github.com/NetApp/harvest/pull/3018))
- Handle Volume Analytics Error Logging ([#3026](https://github.com/NetApp/harvest/pull/3026))
- Vscan Plugin Should Handle Ipv6 Scanners ([#3028](https://github.com/NetApp/harvest/pull/3028))
- Vscan Plugin Should Handle Ipv6 Scanners ([#3034](https://github.com/NetApp/harvest/pull/3034))
- Object Store Metrics Collection For Aggregate ([#3045](https://github.com/NetApp/harvest/pull/3045))
- Throughput Should Use Sum Aggregation ([#3052](https://github.com/NetApp/harvest/pull/3052))
- Harvest Should Collect Power Metrics From A1000 And A900 Clusters ([#3063](https://github.com/NetApp/harvest/pull/3063))
- Quota Dashboard Should Use Kibibytes Instead Of Kilobytes ([#3072](https://github.com/NetApp/harvest/pull/3072))
- Namespace Dashboard Legends Have A Dangling } ([#3075](https://github.com/NetApp/harvest/pull/3075))
- Add Color To Relevant Value Mapping Columns In Dashboards ([#3080](https://github.com/NetApp/harvest/pull/3080))
- Poller Rss Panel Should Ignore Pid ([#3083](https://github.com/NetApp/harvest/pull/3083))
- Remove Quota Asup From Rest ([#3087](https://github.com/NetApp/harvest/pull/3087))
- Remove Threshold From Quota Rest Template ([#3093](https://github.com/NetApp/harvest/pull/3093))
- Add Datacenter, Cluster Columns In Tables With Links ([#3094](https://github.com/NetApp/harvest/pull/3094))

### :closed_book: Documentation
- Update Docker Instructions ([#2940](https://github.com/NetApp/harvest/pull/2940))
- Update Metric Docs For 9.15 ([#2957](https://github.com/NetApp/harvest/pull/2957))
- Add Note For Hardware Requirement For Harvest ([#2964](https://github.com/NetApp/harvest/pull/2964))
- Fix Standalone Harvest Container Deployment Steps ([#2981](https://github.com/NetApp/harvest/pull/2981))
- Release Notes For 24.05.2 ([#2985](https://github.com/NetApp/harvest/pull/2985))
- Add Description In Subsystem Latency Panels ([#3017](https://github.com/NetApp/harvest/pull/3017))
- Update List Of Supported Fsx Dashboards ([#3037](https://github.com/NetApp/harvest/pull/3037))
- Harvest Should Document The Least-Privilege Approach For Rest ([#3047](https://github.com/NetApp/harvest/pull/3047))
- Harvest Getting Started Tutorial ([#3054](https://github.com/NetApp/harvest/pull/3054))
- Describe How To Collect Support Bundle From Nabox4 ([#3071](https://github.com/NetApp/harvest/pull/3071))
- Doc Update For Oauth 2.0 Support In Harvest ([#3073](https://github.com/NetApp/harvest/pull/3073))
- Add Ems Permissions For Rest Least Privilege Approach ([#3088](https://github.com/NetApp/harvest/pull/3088))
- Add container troubleshooting steps ([#3097](https://github.com/NetApp/harvest/pull/3097))

### :zap: Performance
- Improve Prometheus Render Escaping By 23% ([#2922](https://github.com/NetApp/harvest/pull/2922))

### :wrench: Testing
- Quota Tests ([#2924](https://github.com/NetApp/harvest/pull/2924))
- Harvest Should Use Go-Cmp Instead Of Reflect.deepequal ([#3025](https://github.com/NetApp/harvest/pull/3025))

### Refactoring
- Use Builtin Maps Instead Of 3Rd Party ([#3009](https://github.com/NetApp/harvest/pull/3009))
- Remove Dead Code And Reduce 3Rd Party Dependencies ([#3039](https://github.com/NetApp/harvest/pull/3039))
- Remove Obsolete `Version` From Compose Files ([#3042](https://github.com/NetApp/harvest/pull/3042))
- Update Description For Volume Arw Panel ([#3076](https://github.com/NetApp/harvest/pull/3076))
- Remove Deprecated Compliance Dashboard ([#3081](https://github.com/NetApp/harvest/pull/3081))

### Miscellaneous
- Update Module Github.com/Zekrotja/Timedmap To V2 ([#2910](https://github.com/NetApp/harvest/pull/2910))
- Update All Dependencies ([#2926](https://github.com/NetApp/harvest/pull/2926))
- Bump Hashicorp Go-Version ([#2933](https://github.com/NetApp/harvest/pull/2933))
- Update All Dependencies ([#2955](https://github.com/NetApp/harvest/pull/2955))
- Move Gopsutil To V4 ([#2961](https://github.com/NetApp/harvest/pull/2961))
- Update All Dependencies ([#2975](https://github.com/NetApp/harvest/pull/2975))
- Update All Dependencies ([#2995](https://github.com/NetApp/harvest/pull/2995))
- Update Module Github.com/Zekrotja/Timedmap To V2 ([#3010](https://github.com/NetApp/harvest/pull/3010))
- Remove Unused Code ([#3016](https://github.com/NetApp/harvest/pull/3016))
- Update All Dependencies ([#3040](https://github.com/NetApp/harvest/pull/3040))
- Update Golang.org/X/Exp Digest To 8A7402a ([#3058](https://github.com/NetApp/harvest/pull/3058))
- Update All Dependencies ([#3084](https://github.com/NetApp/harvest/pull/3084))

### :hammer: CI
- Fix Flaky Test For Expression ([#2927](https://github.com/NetApp/harvest/pull/2927))
- Update To Use Colored-Line-Number For Linter ([#2930](https://github.com/NetApp/harvest/pull/2930))
- Stop Pollers After Tests In Ci ([#2939](https://github.com/NetApp/harvest/pull/2939))
- Add Zapi Rest Comparison To Ci ([#2945](https://github.com/NetApp/harvest/pull/2945))
- Fix Container Stop In Ci ([#2946](https://github.com/NetApp/harvest/pull/2946))
- Stop Containers After Tests ([#2958](https://github.com/NetApp/harvest/pull/2958))
- Bump Go ([#2965](https://github.com/NetApp/harvest/pull/2965))
- Run Tests Before Docker Publish ([#2990](https://github.com/NetApp/harvest/pull/2990))
- Add Flexgroup Tests ([#3001](https://github.com/NetApp/harvest/pull/3001))
- Bump Go ([#3032](https://github.com/NetApp/harvest/pull/3032))
- Bump Go ([#3089](https://github.com/NetApp/harvest/pull/3089))

---

## 24.05.2 / 2024-06-13 Release
:pushpin: This release is identical to 24.05.0, with the addition of two fixes:

1. A fix that makes the NFS Troubleshooting dashboards load in NAbox and via `bin/harvest grafana import`.
2. A fix for a regression introduced in 24.05.1, that causes FlexGroup volume performance metrics to be skipped.

**Upgrade Recommendation**:
You should upgrade to `24.05.2` if any of the following apply to you:
- You want to use the NFS troubleshooting dashboards.
- You are on version 24.05.1 and your cluster includes FlexGroup volumes.

---

## 24.05.1 / 2024-05-29 Release
:pushpin: This release is the same as 24.05.0 with a fix that makes the NFS Troubleshooting dashboards load in NAbox. If you are not using NAbox or you do not use the NFS trouble shooting dashboards, you can ignore this release.

---

## 24.05.0 / 2024-05-20 Release
:pushpin: Highlights of this major release include:
- Harvest supports consistency groups (CG) in the SnapMirror dashboard. Thanks to @Nikhita-13 for reporting this.
- We've fixed an intermittent latency/ops spike problem caused by Harvest incorrectly handling ONTAP partial aggregation. This impacted all perf objects. A big thank you to @summertony15 for reporting this critical issue.
- Harvest dashboards are compatible with Grafana 10.x.x versions.
- :gem: LUN, Flexgroup and cDot dashboard updated to work with FSx. Some panels are blank because FSx does not have that data.
- The credentials script supports providing both username and password. Thanks to @kbhalaki for reporting.
- Harvest configuration file supports reading parameters from environment variables. Kudos to @wally007 for the suggestion.
- Harvest includes [remediation steps](https://netapp.github.io/harvest/nightly/resources/ems-alert-runbook/) for EMS alerts.

- :gem: New Dashboards:
  - `NFS Troubleshooting` which provides links to detailed dashboards. Thanks to RustyBrown for contributing these.
  - Detailed Dashboards: `Volume by SVM` and `Volume Deep Dive`.

- :rocket: Performance Improvements:
  - Rest/RestPerf Collector only requests metrics defined in templates, reducing API time, payload size, and collection load.
  - TopK queries in dashboards are now faster. Thanks to AlessandroN for reporting.

- :star: Several of the existing dashboards include new panels in this release:
  - Workload dashboard includes adaptive QoS used percentage tracking. Thanks to @faguayot for reporting.
  - Network dashboard includes ethernet errors. Thanks to Rusty Brown for contributing.
  - Node dashboard includes the BMC firmware version. Thanks to @summertony15 for reporting.
  - SVM dashboard now includes NFS4.2 panels. Thanks to Didlier for reporting.
  - The Volume dashboard includes several new panels:
    - Volume growth rate panels. Thanks to AlessandroN for reporting.
    - I/O density panels. Thanks to @jgasher for reporting.
    - Volume capacity forecasting panels, predicting a volume's used sized over the next 15 days. Thanks to @s-kuchi for reporting.


- :ear_of_rice: Harvest includes a new template to collect lock counts at the node, SVM, LIF, and volume levels.. Thanks to @troysmullerna for reporting.

- :closed_book: Documentation Additions:
  - How to customize Prometheus's retention period in a Docker deployment. Thanks to @WayneShen2 for the suggestion.
  - How to use endpoints in a REST collector template. Thanks to Hubert for reporting.
  - Harvest includes [remediation steps](https://netapp.github.io/harvest/nightly/resources/ems-alert-runbook/) for EMS alerts.
  - How to use `confpath` to extend templates.

- Harvest supports embedded exporters in Harvest configuration. This means you can define your exporters in one place instead of multiple. Thanks to @wagneradrian92 for reporting.
- Harvest supports exporting to multiple InfluxDB instances. Thanks to @figeac888 for reporting.
- Node label metrics include HA partner details. Thanks to @johnwarlick for reporting.

## Announcements

:bangbang: **IMPORTANT** Release `24.05` removes duplicate quota metrics. If you wish to enable them, refer [here](https://github.com/NetApp/harvest/discussions/2895).

:bulb: **IMPORTANT** After upgrading, don't forget to re-import your dashboards to get all the new enhancements and fixes. You can import them via the 'bin/harvest grafana import' CLI, from the Grafana UI, or from the 'Maintenance > Reset Harvest Dashboards' button in NAbox.

## Known Issues

- :warning: Harvest does not calculate power metrics for AFF A250 systems. This data is not available from ONTAP via ZAPI or REST. See ONTAP bug [1511476](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1511476) for more details.
- :warning: ONTAP does not include REST metrics for `offbox_vscan_server` and `offbox_vscan` until ONTAP 9.13.1. See ONTAP bug [1473892](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1473892) for more details.

**IMPORTANT** 7-mode filers that are not on the latest release of ONTAP may experience TLS connection issues with errors like `tls: server selected unsupported protocol version 301`. This is caused by a change in Go 1.18. The [default for TLS client connections was changed to TLS 1.2](https://tip.golang.org/doc/go1.18#tls10) in Go 1.18. Please upgrade your 7-mode filers (recommended) or set `tls_min_version: tls10` in your `harvest.yml` [poller section](https://github.com/NetApp/harvest/tree/release/22.05.0#pollers). See [#1007](https://github.com/NetApp/harvest/issues/1007) for more details.

## Thanks to all the awesome contributors

:metal: A big thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards this release:

@BrendonA667, @Nikhita-13, @WayneShen2, @derDaywalker, @faguayot, @figeac888, @jgasher, @johnwarlick, @kbhalaki, @rdecaneva, @s-kuchi, @summertony15, @troysmullerna, @wagneradrian92, @wally007, @ybizeul, AlessandroN, Didlier, Hubert, Rusty Brow, Tamas Zsolt

:seedling: This release includes 42 features, 38 bug fixes, 10 documentation, 1 performance, 6 styling, 9 refactoring, 16 miscellaneous, and 17 ci pull requests.

### :rocket: Features
- Adding Zapi/Rest Templates For Lock-Get-Iter & Protocols/Locks ([#2706](https://github.com/NetApp/harvest/pull/2706))
- Dashboards Would Work With Grafana 10.X.x ([#2713](https://github.com/NetApp/harvest/pull/2713))
- Add Harvest.yml Environment Variable Expansion ([#2714](https://github.com/NetApp/harvest/pull/2714))
- Metadata Dashboard Should Include Poller Rss Panels And Time Se… ([#2716](https://github.com/NetApp/harvest/pull/2716))
- Harvest Should Export To Multiple Influxdb Exporters ([#2722](https://github.com/NetApp/harvest/pull/2722))
- Adding Ha_partner Info In Node ([#2723](https://github.com/NetApp/harvest/pull/2723))
- Improve Rest Collector ([#2740](https://github.com/NetApp/harvest/pull/2740))
- Harvest Should Track Network Bytes Received And Number Of Ontap… ([#2745](https://github.com/NetApp/harvest/pull/2745))
- Harvest Should Handle Ontap Counter Manager Rejection Errors ([#2747](https://github.com/NetApp/harvest/pull/2747))
- Harvest Network Dashboard Should Show Ethernet Errors ([#2748](https://github.com/NetApp/harvest/pull/2748))
- Changed Plugin Generated Metric Naming For Lock Object ([#2750](https://github.com/NetApp/harvest/pull/2750))
- Usage Of Predict_linear Function In Volume Dashboard ([#2763](https://github.com/NetApp/harvest/pull/2763))
- Improve Restperf Collector ([#2765](https://github.com/NetApp/harvest/pull/2765))
- Harvest Should Include Nfs Troubleshooting Dashboards ([#2766](https://github.com/NetApp/harvest/pull/2766))
- Adding Volume Growth Rate Panels In Volume Dashboard ([#2768](https://github.com/NetApp/harvest/pull/2768))
- Harvest Should Reduce Batch Size And Retry When Ontap Times Out ([#2770](https://github.com/NetApp/harvest/pull/2770))
- Ignore Performance Counters With Partial Aggregation ([#2775](https://github.com/NetApp/harvest/pull/2775))
- Harvest Should Reduce Batch Size And Retry When Ontap Times Out ([#2776](https://github.com/NetApp/harvest/pull/2776))
- Harvest Should Log When The Template Is Missing ([#2779](https://github.com/NetApp/harvest/pull/2779))
- Add Instance Log For Latency Calculation ([#2794](https://github.com/NetApp/harvest/pull/2794))
- Harvest Should Collect The Bmc Firmware Version ([#2800](https://github.com/NetApp/harvest/pull/2800))
- Add I/O Density Panels To Volume Dashboard ([#2805](https://github.com/NetApp/harvest/pull/2805))
- Reduce Dependencies ([#2812](https://github.com/NetApp/harvest/pull/2812))
- Use Constrained Topk To Improve Dashboard Performance ([#2825](https://github.com/NetApp/harvest/pull/2825))
- Supporting Consistency Group Drilldowns In Snapmirror Dashboard ([#2830](https://github.com/NetApp/harvest/pull/2830))
- Harvest Should Include Remediation Steps For Ems Alerts ([#2836](https://github.com/NetApp/harvest/pull/2836))
- Harvest Svm Dashboard Should Include Nfsv4.2 Panels ([#2846](https://github.com/NetApp/harvest/pull/2846))
- Adding Description To Svm Panels ([#2861](https://github.com/NetApp/harvest/pull/2861))
- Harvest Should Support Embedded Exporters ([#2864](https://github.com/NetApp/harvest/pull/2864))
- Adaptive Qos Used% Tracking ([#2865](https://github.com/NetApp/harvest/pull/2865))
- Credentials Script Should Support Both Username And Password ([#2870](https://github.com/NetApp/harvest/pull/2870))
- Adding Panel Descriptions In All Dashboards ([#2878](https://github.com/NetApp/harvest/pull/2878))
- Remove Hidden Topk Variables From Dashboards ([#2881](https://github.com/NetApp/harvest/pull/2881))
- Remove Duplicate Quota Metrics ([#2886](https://github.com/NetApp/harvest/pull/2886))
- Remove Hidden Topk Variables From Dashboards ([#2889](https://github.com/NetApp/harvest/pull/2889))
- Adding Description To Panels ([#2891](https://github.com/NetApp/harvest/pull/2891))
- Add Test Case For Join Queries In A Table ([#2892](https://github.com/NetApp/harvest/pull/2892))
- Adding Details Folder In Docker ([#2896](https://github.com/NetApp/harvest/pull/2896))
- Enable Request/Response Logging For Rest And Restperf Plugins ([#2898](https://github.com/NetApp/harvest/pull/2898))
- Flexgroup And Lun Dashboards Work With Fsx ([#2899](https://github.com/NetApp/harvest/pull/2899))
- Remove Hidden Topk From Aggregation Dashboard ([#2900](https://github.com/NetApp/harvest/pull/2900))
- Cdot Dashboards Work With Fsx ([#2903](https://github.com/NetApp/harvest/pull/2903))

### :bug: Bug Fixes
- Handle Inter-Cluster Snapmirrors When Different Datacenter ([#2688](https://github.com/NetApp/harvest/pull/2688))
- Display Poller Status With Harvest_docker Env ([#2705](https://github.com/NetApp/harvest/pull/2705))
- Sync Svm_labels With Ontap Cli For Zapi Collector ([#2711](https://github.com/NetApp/harvest/pull/2711))
- Harvest Should Not Panic When A Poller Has No Config ([#2718](https://github.com/NetApp/harvest/pull/2718))
- Convert Qos Adaptive Policy Configuration Ops To Tb ([#2720](https://github.com/NetApp/harvest/pull/2720))
- Align Make Build Version With Prod Version ([#2732](https://github.com/NetApp/harvest/pull/2732))
- Add Volume Filter For Per Volume Statistics ([#2742](https://github.com/NetApp/harvest/pull/2742))
- Restperf Panics When Pollinstance Fails ([#2743](https://github.com/NetApp/harvest/pull/2743))
- Storagegrid Should Honor Template Api Version ([#2744](https://github.com/NetApp/harvest/pull/2744))
- Qospolicyfixed Should Ignore Missing Min-Throughput ([#2754](https://github.com/NetApp/harvest/pull/2754))
- Remove Unused Error From Rest ([#2758](https://github.com/NetApp/harvest/pull/2758))
- Harvest Dashboard Variables Should Use Fsx Friendly Queries Wher… ([#2778](https://github.com/NetApp/harvest/pull/2778))
- Harvest Should Support Poller Names With Spaces In Their Names ([#2780](https://github.com/NetApp/harvest/pull/2780))
- Restperf Ignore Performance Counters With Partial Aggregation ([#2783](https://github.com/NetApp/harvest/pull/2783))
- Using Volume_total_data Instead Of Read_data And Write_data ([#2786](https://github.com/NetApp/harvest/pull/2786))
- Change Diskperf Warn To Debug When Metrics Have Record False ([#2796](https://github.com/NetApp/harvest/pull/2796))
- Parity Of Id Value In Disk Restperf And Zapiperf ([#2807](https://github.com/NetApp/harvest/pull/2807))
- Iops Should Not Have Decimals In Dashboards ([#2810](https://github.com/NetApp/harvest/pull/2810))
- Node Dashboard, Bmc Column Should Have String Unit ([#2815](https://github.com/NetApp/harvest/pull/2815))
- Add Root_volume Label For Vol0 Volumes ([#2816](https://github.com/NetApp/harvest/pull/2816))
- The Unix Poller Should Detect Poller Names With Spaces ([#2818](https://github.com/NetApp/harvest/pull/2818))
- Add Missing Label Snapshot_autodelete To Rest Volume Template ([#2822](https://github.com/NetApp/harvest/pull/2822))
- Resolve Duplicate Skip Increment ([#2826](https://github.com/NetApp/harvest/pull/2826))
- Zapiperf Pollcounter Error ([#2831](https://github.com/NetApp/harvest/pull/2831))
- Update Workload Templates To Use Default Schedule For Counter And Instance ([#2835](https://github.com/NetApp/harvest/pull/2835))
- Fix Dashboard Sort Test ([#2844](https://github.com/NetApp/harvest/pull/2844))
- Svm Cifs Total Ops Should Sum All Types Of Ops ([#2853](https://github.com/NetApp/harvest/pull/2853))
- Volume Count In Datacenter Dashboard ([#2854](https://github.com/NetApp/harvest/pull/2854))
- Load Cert Pool When Ca_cert Is Defined In Harvest.yml ([#2855](https://github.com/NetApp/harvest/pull/2855))
- Qos Mbps Should Report With Precision ([#2871](https://github.com/NetApp/harvest/pull/2871))
- Remove Duplicate Columns From Qos Adaptive Dashboard Tables ([#2872](https://github.com/NetApp/harvest/pull/2872))
- Adaptive Qos Table Grafana 9 Workaround ([#2873](https://github.com/NetApp/harvest/pull/2873))
- Handling Index For Quota ([#2874](https://github.com/NetApp/harvest/pull/2874))
- Add Regex For Node Table ([#2884](https://github.com/NetApp/harvest/pull/2884))
- Add All To Svm Dropdown In Volume Deep Dive Dashboard ([#2901](https://github.com/NetApp/harvest/pull/2901))
- Add Restgap For Volume_space_logical_available ([#2904](https://github.com/NetApp/harvest/pull/2904))
- Handling Missing Protection_mode In Disk Rest Call ([#2905](https://github.com/NetApp/harvest/pull/2905))
- Duplicate Instance Key Issue Quota Metrics ([#2913](https://github.com/NetApp/harvest/pull/2913))

### :closed_book: Documentation
- Describe How To Use Confpath To Extend Templates ([#2725](https://github.com/NetApp/harvest/pull/2725))
- Prometheus Retention Period Customization In Docker Instructions ([#2760](https://github.com/NetApp/harvest/pull/2760))
- Update Container Faq ([#2809](https://github.com/NetApp/harvest/pull/2809))
- Harvest Should Document How To Use Endpoints In Rest Collector Template ([#2811](https://github.com/NetApp/harvest/pull/2811))
- Add Workload-Class As Supported Field For Workload Perf Filter ([#2828](https://github.com/NetApp/harvest/pull/2828))
- Clarify Include_all_labels And Export Options ([#2839](https://github.com/NetApp/harvest/pull/2839))
- Fix Latency Average Units ([#2851](https://github.com/NetApp/harvest/pull/2851))
- Add Jitter Documentation ([#2860](https://github.com/NetApp/harvest/pull/2860))
- Update List Of Supported Fsx Dashboards ([#2906](https://github.com/NetApp/harvest/pull/2906))
- 24.05 Ontap Metric Docs ([#2907](https://github.com/NetApp/harvest/pull/2907))

### :zap: Performance
- Remove Visits Counter From Workload Detail Templates ([#2824](https://github.com/NetApp/harvest/pull/2824))

### Styling
- Remove Potential Nil Dereference ([#2675](https://github.com/NetApp/harvest/pull/2675))
- Fix Spelling Error In Description ([#2712](https://github.com/NetApp/harvest/pull/2712))
- Log Via Send Instead Of Msg("") ([#2737](https://github.com/NetApp/harvest/pull/2737))
- Jitter Logging ([#2781](https://github.com/NetApp/harvest/pull/2781))
- Improving Debug Log Clarity And Reducing Noise ([#2795](https://github.com/NetApp/harvest/pull/2795))
- Reduce Log Noise When Disk Attributes Are Missing ([#2798](https://github.com/NetApp/harvest/pull/2798))

### Refactoring
- Use Range Over Int Go 1.22 Feature ([#2684](https://github.com/NetApp/harvest/pull/2684))
- Don't Double Log Error ([#2739](https://github.com/NetApp/harvest/pull/2739))
- Remove Unused Schedule In Security Account Rest Template ([#2751](https://github.com/NetApp/harvest/pull/2751))
- Log With Msg Not Msgf ([#2753](https://github.com/NetApp/harvest/pull/2753))
- Changed Field To Use # To Handle Counter Testcase ([#2759](https://github.com/NetApp/harvest/pull/2759))
- Remove Trace Logging ([#2813](https://github.com/NetApp/harvest/pull/2813))
- Fetch Constituents When Asked From Template ([#2838](https://github.com/NetApp/harvest/pull/2838))
- Use Cmp.or For Envvar-Default Pattern ([#2845](https://github.com/NetApp/harvest/pull/2845))
- Changed Allvalue To .* ([#2887](https://github.com/NetApp/harvest/pull/2887))

### Miscellaneous
- Merge Release/24.02.0 To Main ([#2701](https://github.com/NetApp/harvest/pull/2701))
- Update Golang.org/X/Exp Digest To 814Bf88 ([#2709](https://github.com/NetApp/harvest/pull/2709))
- Bump ([#2721](https://github.com/NetApp/harvest/pull/2721))
- Update Module Github.com/Shirou/Gopsutil/V3 To V3.24.2 ([#2724](https://github.com/NetApp/harvest/pull/2724))
- Update Module Github.com/Go-Openapi/Spec To V0.21.0 ([#2733](https://github.com/NetApp/harvest/pull/2733))
- Update Golang.org/X/Exp Digest To C7f7c64 ([#2755](https://github.com/NetApp/harvest/pull/2755))
- Update All Dependencies ([#2772](https://github.com/NetApp/harvest/pull/2772))
- Use The Correct Format For .Golangci.yml ([#2791](https://github.com/NetApp/harvest/pull/2791))
- Update All Dependencies ([#2797](https://github.com/NetApp/harvest/pull/2797))
- Commitlint Changed File Extension ([#2801](https://github.com/NetApp/harvest/pull/2801))
- Update All Dependencies ([#2814](https://github.com/NetApp/harvest/pull/2814))
- Update Golang.org/X/Exp Digest To 93D18d7 ([#2833](https://github.com/NetApp/harvest/pull/2833))
- Address Code Scanning Issues In Thirdparty ([#2842](https://github.com/NetApp/harvest/pull/2842))
- Update Golang.org/X/Exp Digest To Fe59bbe ([#2843](https://github.com/NetApp/harvest/pull/2843))
- Update All Dependencies ([#2877](https://github.com/NetApp/harvest/pull/2877))
- Update Module Github.com/Zekrotja/Timedmap To V2 ([#2888](https://github.com/NetApp/harvest/pull/2888))

### :hammer: CI
- Merge 24.02 Into Main ([#2683](https://github.com/NetApp/harvest/pull/2683))
- Use Go Run To Track Tool Dependencies ([#2685](https://github.com/NetApp/harvest/pull/2685))
- Run Go Mod Tidy Before Linting And Govulncheck ([#2708](https://github.com/NetApp/harvest/pull/2708))
- Preallocate Slices That Can Be ([#2727](https://github.com/NetApp/harvest/pull/2727))
- Bump Go ([#2728](https://github.com/NetApp/harvest/pull/2728))
- Bump Dependencies ([#2730](https://github.com/NetApp/harvest/pull/2730))
- Improve Lint ([#2764](https://github.com/NetApp/harvest/pull/2764))
- Enable Integration Dep Updates With Renovate ([#2789](https://github.com/NetApp/harvest/pull/2789))
- Move More Ci From Github Actions To Makefile ([#2803](https://github.com/NetApp/harvest/pull/2803))
- Bump Go ([#2808](https://github.com/NetApp/harvest/pull/2808))
- Add Flexcache To Flaky Counter List ([#2817](https://github.com/NetApp/harvest/pull/2817))
- Enable More Linters ([#2841](https://github.com/NetApp/harvest/pull/2841))
- Enable More Linters ([#2847](https://github.com/NetApp/harvest/pull/2847))
- Enable Golanglint "Canonicalheader" Linter ([#2876](https://github.com/NetApp/harvest/pull/2876))
- Bump Go ([#2880](https://github.com/NetApp/harvest/pull/2880))
- Bump Dependencies ([#2882](https://github.com/NetApp/harvest/pull/2882))
- Increase Golangci-Lint Timeout ([#2912](https://github.com/NetApp/harvest/pull/2912))

---

## 24.02.0 / 2024-02-21 Release
:pushpin: Highlights of this major release include:
- New Datacenter dashboard which contains node health, capacity, performance, storage efficiency, issues, snapshot, power, and temperature details.
- Harvest includes SnapMirror active sync EMS events with alert rules. Thanks @Nikhita-13 for reporting.
- Harvest monitors FlexCache performance metrics and includes a new FlexCache dashboard to visualize them. Thanks to @ewilts for raising.
- Harvest detects HA pair down and sensor failures. These are shown in the Health dashboard. Thanks to @johnwarlick for raising.
- Harvest monitors MetroCluster diagnostics and shows them in the MetroCluster dashboard. Thanks to @wagneradrian92 for reporting.
- We improved the performance of all dashboards that include topk queries. Thanks to @mamoep for reporting!
- We added filter support for the ZapiPerf collector. See [filter](https://netapp.github.io/harvest/nightly/configure-zapi/#filter) for more detail. Thanks to @debbrata-netapp for reporting.
- A `bin/harvest grafana customize` command that writes the dashboards to the filesystem so other programs can manage them. Thanks to @nicolai-hornung-bl for reporting!
- We fixed an intermittent latency spike problem that impacted all perf objects. Thanks to @summertony15 and @rodenj1 for reporting this critical issue.

- :star: Several of the existing dashboards include new panels in this release:
- Node and Aggregate dashboard include volume stats panels. Thanks to @BrendonA667 for raising.
- SVM dashboard includes volume capacity panels. Thanks to @BrendonA667 for raising.
- SnapMirror dashboard includes automated_failover and automated_failover_duplex policies.

- More Harvest dashboard dropdown variables include the `All` option. Making it easier to get an overview of your environment.
- All EMS alerts include an impact annotation. Thanks to @Divya for raising.

- :ear_of_rice: Harvest includes new templates to collect:
- Network filesystem (NFS) rewinds performance metrics (rw_ctx). Thanks to @shawnahall71 for raising
- Network data management protocol (NDMP) session metrics. Thanks to @schumijo for raising.

- :closed_book: Documentation additions
- Harvest describe why and how to configure Docker's logging drivers [Docker logging configuration](https://netapp.github.io/harvest/nightly/install/containers/#note-on-docker-logging-configuration) Thanks to @Madaan for raising.
- How to create templates that use ONTAP's private CLI [details](https://netapp.github.io/harvest/nightly/configure-rest/#ontap-private-cli)
- How to create custom Grafana dashboards [Steps](https://netapp.github.io/harvest/nightly/dashboards/#creating-a-custom-grafana-dashboard-with-harvest-metrics-stored-in-prometheus)
- How to validate your `harvest.yml` file and share a redacted copy with the Harvest team. [Details](https://netapp.github.io/harvest/nightly/help/config-collection/)
- Harvest describes high-level concepts [here](https://netapp.github.io/harvest/nightly/concepts/) Thanks to @norespers for raising.

- All constituents are disabled by default for workload detail performance templates.
- The `bin/harvest zapi` CLI now supports a `timeout` argument.
- Harvest performance collectors (ZapiPerf and RestPerf) ask ONTAP for performance counter metadata every 24 hours instead of every 20 minutes. Thanks to BrianMa for raising.
- The Harvest REST collector's `api_time` metric now includes the API time for all template endpoints. Thanks to ChristopherWilcox for raising.

## Announcements

:bangbang: **IMPORTANT** Release `24.02` disables four templates that collected metrics not used in dashboards.
These four templates are disabled by default: `ObjectStoreClient`, `TokenManager`, `OntapS3SVM`, and `Vscan`.
This change was made to reduce the number of collected metrics.
If you require these templates, you can enable them by uncommenting them in their corresponding `default.yaml` or by extending the [existing object template](https://netapp.github.io/harvest/latest/configure-templates/#extend-an-existing-object-template).

:small_red_triangle: **IMPORTANT** The minimum version of Prometheus required to run Harvest is now 2.33.
Version 2.33 is required to take advantage of [Prometheus's `@` modifier](https://prometheus.io/docs/prometheus/latest/querying/basics/#modifier).
Please upgrade your Prometheus server to at least 2.33 before upgrading Harvest.

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the 'bin/harvest grafana import' CLI, from the Grafana UI, or from the 'Maintenance > Reset Harvest Dashboards' button in NAbox.

## Known Issues

- Harvest does not calculate power metrics for AFF A250 systems. This data is not available from ONTAP via ZAPI or REST.
  See ONTAP bug [1511476](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1511476) for more details.

- ONTAP does not include REST metrics for `offbox_vscan_server` and `offbox_vscan` until ONTAP 9.13.1. See ONTAP bug
  [1473892](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1473892) for more details.

**IMPORTANT** 7-mode filers that are not on the latest release of ONTAP may experience TLS connection issues with errors
like `tls: server selected unsupported protocol version 301` This is caused by a change in Go 1.18.
The [default for TLS client connections was changed to TLS 1.2](https://tip.golang.org/doc/go1.18#tls10) in Go 1.18.
Please upgrade your 7-mode filers (recommended) or set `tls_min_version: tls10` in
your `harvest.yml` [poller section](https://github.com/NetApp/harvest/tree/release/22.05.0#pollers).
See [#1007](https://github.com/NetApp/harvest/issues/1007) for more details.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards
this release:

@shawnahall71, @pilot7777, @ben, @Madaan, @johnwarlick, @jfong5040, @santosh725, @summertony15, @jmg011, @cheese1, @mamoep, @Falcon667, @Dess, @debbrata-netapp, @ewilts,
@Nikhita-13, @norespers, @nicolai-hornung-bl, @BrendonA667, @schumijo, @Divya, @joshuacook-tamu, @wagneradrian92, @george-strother

:seedling: This release includes 26 features, 24 bug fixes, 20 documentation, 3 styling, 5 refactoring, 11 miscellaneous, and 12 ci pull requests.

### :rocket: Features
- Include Start Time, Exported Metrics, And Poll Duration In Collector logs ([#2493](https://github.com/NetApp/harvest/pull/2493))
- Adding Rw_ctx Zapiperf Object Template ([#2494](https://github.com/NetApp/harvest/pull/2494))
- Change Pollcounter Schedule To 24H ([#2499](https://github.com/NetApp/harvest/pull/2499))
- Add Ha Down And Sensor Issues In Health Dashboard ([#2519](https://github.com/NetApp/harvest/pull/2519))
- Adding Ndmp Session Rest Template ([#2531](https://github.com/NetApp/harvest/pull/2531))
- Use Modifier For Topk To Improve Svm Dashboard Performance ([#2553](https://github.com/NetApp/harvest/pull/2553))
- Add Timeout For Zapi Cli ([#2566](https://github.com/NetApp/harvest/pull/2566))
- Restperf Disk Plugin Should Support Metric Customization ([#2573](https://github.com/NetApp/harvest/pull/2573))
- Add Filter Support For Zapiperf Collector ([#2575](https://github.com/NetApp/harvest/pull/2575))
- FlexCache Monitoring ([#2583](https://github.com/NetApp/harvest/pull/2583))
- Supporting Automated_failover, Automated_failover_duplex Policy In Sm ([#2584](https://github.com/NetApp/harvest/pull/2584))
- Disabled The Templates Whose All Metrics Are Not Consumed In Dashboards ([#2587](https://github.com/NetApp/harvest/pull/2587))
- Harvest Should Include Snapmirror Active Sync Ems Events ([#2588](https://github.com/NetApp/harvest/pull/2588))
- Use Modifier For Topk To Improve Dashboard Performance ([#2590](https://github.com/NetApp/harvest/pull/2590))
- Harvest Should Include A Snapmirror Active Sync Template ([#2596](https://github.com/NetApp/harvest/pull/2596))
- Disable Constituents By Default For Workload Detail Performance Templates ([#2598](https://github.com/NetApp/harvest/pull/2598))
- Adding Template For Metrocluster Diagnostics Check ([#2601](https://github.com/NetApp/harvest/pull/2601))
- Adding Per Volume Panels In Svm Dashboard ([#2602](https://github.com/NetApp/harvest/pull/2602))
- Add Grafana Customize Command ([#2619](https://github.com/NetApp/harvest/pull/2619))
- Add Volume Stats To Node And Aggregate Dashboard ([#2627](https://github.com/NetApp/harvest/pull/2627))
- Ems Alerts Should Include An Impact Annotation ([#2631](https://github.com/NetApp/harvest/pull/2631))
- Improving Debug Log Clarity And Reducing Noise ([#2637](https://github.com/NetApp/harvest/pull/2637))
- Datacenter Dashboard ([#2650](https://github.com/NetApp/harvest/pull/2650))
- Harvest Dashboards Should Include An All Option ([#2661](https://github.com/NetApp/harvest/pull/2661))
- Percent Unit Panels Should Use Decimal Points ([#2663](https://github.com/NetApp/harvest/pull/2663))
- Change Stat Panel For Uptime,Power Status,Fan Status To Table In Node Dashboard ([#2668](https://github.com/NetApp/harvest/pull/2668))

### :bug: Bug Fixes
- Handled Missing Uuid In Volume For Change_log ([#2478](https://github.com/NetApp/harvest/pull/2478))
- Remove Docs From Deb Binary ([#2489](https://github.com/NetApp/harvest/pull/2489))
- Parsed Logger Changes ([#2490](https://github.com/NetApp/harvest/pull/2490))
- Array Metrics Should Have Correct Base Label In Zapiperf ([#2496](https://github.com/NetApp/harvest/pull/2496))
- Harvest Should Collect Luns In Qtress ([#2502](https://github.com/NetApp/harvest/pull/2502))
- Grafana Export Should Set Correct Permissions ([#2505](https://github.com/NetApp/harvest/pull/2505))
- `Begin` Log For Pollcounter And Pollinstance Should Be In Ms ([#2509](https://github.com/NetApp/harvest/pull/2509))
- Quickstart.md Docs Should Not Duplicate Pollers ([#2521](https://github.com/NetApp/harvest/pull/2521))
- Print Results If Not Nil For Rest Cli ([#2525](https://github.com/NetApp/harvest/pull/2525))
- `Storage Efficiency Ratios` Panels Should Show Cluster Capacity ([#2529](https://github.com/NetApp/harvest/pull/2529))
- Qos Fixed% Should Include Admin Svm Qos Policy ([#2532](https://github.com/NetApp/harvest/pull/2532))
- Handling Shelf_new_status For 7Mode ([#2535](https://github.com/NetApp/harvest/pull/2535))
- Rest Aggr.yaml Template Should Be In The 9.11.0 Folder ([#2538](https://github.com/NetApp/harvest/pull/2538))
- Storagegrid Collectors Should Support Only_cluster_instance ([#2542](https://github.com/NetApp/harvest/pull/2542))
- Intermittent Latency Spike ([#2548](https://github.com/NetApp/harvest/pull/2548))
- Hide Idle Metric And Max To Auto For Cpu_domain_busy ([#2555](https://github.com/NetApp/harvest/pull/2555))
- Storagegrid Error When Password Has ([#2576](https://github.com/NetApp/harvest/pull/2576))
- Container Workflow Creates Files As Root Even When The Commands Are Executed By A Non-Root User ([#2581](https://github.com/NetApp/harvest/pull/2581))
- Clone_split_estimate Parse Error ([#2613](https://github.com/NetApp/harvest/pull/2613))
- Qos Latency Spikes Due To Low Iops ([#2615](https://github.com/NetApp/harvest/pull/2615))
- Fix Datacenter Count In Metadata Dashboard ([#2622](https://github.com/NetApp/harvest/pull/2622))
- Doctor Print Should Include Child Pollers Into Optional Parent Pollers ([#2641](https://github.com/NetApp/harvest/pull/2641))
- Remove Max Percent Limit From 'Volumes Per Snapshot Reserve Used' Panel ([#2662](https://github.com/NetApp/harvest/pull/2662))
- Align Template Name With Object Name For Ndmp ([#2667](https://github.com/NetApp/harvest/pull/2667))
- Honor absolute paths from the HARVEST_CONF environment variable ([#2674](https://github.com/NetApp/harvest/pull/2674))
- Rest collector should include endpoint `api_time`s ([#2679](https://github.com/NetApp/harvest/pull/2679))
- StorageGrid Rest collector doesn't remove deleted Objects ([#2677](https://github.com/NetApp/harvest/pull/2677))
- NABox doctor command errors for custom.yaml ([#2691](https://github.com/NetApp/harvest/pull/2691))
- WaflSizer RestPerf template panics ([#2695](https://github.com/NetApp/harvest/pull/2695))
- Purging unused metrics from shelf template for 7mode ([#2696](https://github.com/NetApp/harvest/pull/2696))
- Handle inter-cluster snapmirrors when different datacenter ([#2697](https://github.com/NetApp/harvest/pull/2697))
- Multi poller in a container should route logs to console ([#2698](https://github.com/NetApp/harvest/pull/2698))

### :closed_book: Documentation
- Fix Service Latency ([#2492](https://github.com/NetApp/harvest/pull/2492))
- Fix Doc Link From Changelog Dashboard ([#2510](https://github.com/NetApp/harvest/pull/2510))
- How To Use Harvest With Rest Private Cli ([#2523](https://github.com/NetApp/harvest/pull/2523))
- Add Docker Logging Configuration Guide ([#2524](https://github.com/NetApp/harvest/pull/2524))
- Mention Iec Is Base2 And Source ([#2527](https://github.com/NetApp/harvest/pull/2527))
- Change Link From Netapp.io To Github ([#2533](https://github.com/NetApp/harvest/pull/2533))
- Steps To Create Custom Grafana Dashboard ([#2550](https://github.com/NetApp/harvest/pull/2550))
- Add Type And Base For Qos Detail Metrics ([#2557](https://github.com/NetApp/harvest/pull/2557))
- Ems Doc Update ([#2561](https://github.com/NetApp/harvest/pull/2561))
- Unit For Nfs Throughput Should Be B_per_sec ([#2562](https://github.com/NetApp/harvest/pull/2562))
- Consolidate Upgrade Steps With Install ([#2567](https://github.com/NetApp/harvest/pull/2567))
- Bump The Minimum Prometheus Version To 2.33 ([#2569](https://github.com/NetApp/harvest/pull/2569))
- Add Restart Information In Power Document ([#2572](https://github.com/NetApp/harvest/pull/2572))
- Add Rest Strategy Under Left Nav ([#2578](https://github.com/NetApp/harvest/pull/2578))
- Add Rest Permissions ([#2604](https://github.com/NetApp/harvest/pull/2604))
- Add Fsa Template Description ([#2606](https://github.com/NetApp/harvest/pull/2606))
- Update Grafana Datasource Docs ([#2614](https://github.com/NetApp/harvest/pull/2614))
- Add Vserver For Rest Role Creation ([#2620](https://github.com/NetApp/harvest/pull/2620))
- Fix Broken Link And Remove Todo ([#2624](https://github.com/NetApp/harvest/pull/2624))
- Harvest Should Describe High-Level Concepts ([#2625](https://github.com/NetApp/harvest/pull/2625))
- Add doctor print commands for each platform ([2670](https://github.com/NetApp/harvest/pull/2670))
- Release 24.02 metric docs ([#2694](https://github.com/NetApp/harvest/pull/2694))
- Debian upgrade documentation ([#2699](https://github.com/NetApp/harvest/pull/2699))

### Styling
- Resolve Spell Check Warnings ([#2461](https://github.com/NetApp/harvest/pull/2461))
- Address All Lint Errors ([#2643](https://github.com/NetApp/harvest/pull/2643))
- Address Lint Warnings In Integration ([#2659](https://github.com/NetApp/harvest/pull/2659))

### Refactoring
- Move `Begin` Logging To The End Of The Line ([#2513](https://github.com/NetApp/harvest/pull/2513))
- Update Aggr Dashboard To Sync With Sm ([#2568](https://github.com/NetApp/harvest/pull/2568))
- Remove Dead Code ([#2570](https://github.com/NetApp/harvest/pull/2570))
- Address Data Flow Analysis Warnings ([#2589](https://github.com/NetApp/harvest/pull/2589))
- Revert Ontap Mediator Alert Names ([#2618](https://github.com/NetApp/harvest/pull/2618))

### Miscellaneous
- Update All Dependencies ([#2481](https://github.com/NetApp/harvest/pull/2481))
- Merge 23.11.0 To Main ([#2488](https://github.com/NetApp/harvest/pull/2488))
- Update All Dependencies ([#2522](https://github.com/NetApp/harvest/pull/2522))
- Update All Dependencies ([#2543](https://github.com/NetApp/harvest/pull/2543))
- Update All Dependencies ([#2558](https://github.com/NetApp/harvest/pull/2558))
- Update All Dependencies ([#2564](https://github.com/NetApp/harvest/pull/2564))
- Update All Dependencies ([#2579](https://github.com/NetApp/harvest/pull/2579))
- Update All Dependencies ([#2585](https://github.com/NetApp/harvest/pull/2585))
- Update Golang.org/X/Exp Digest To 1B97071 ([#2592](https://github.com/NetApp/harvest/pull/2592))
- Update All Dependencies ([#2629](https://github.com/NetApp/harvest/pull/2629))
- Update Golangci/Golangci-Lint-Action Action To V4 ([#2653](https://github.com/NetApp/harvest/pull/2653))
- Update all dependencies ([#2687](https://github.com/NetApp/harvest/pull/2687))

### :hammer: CI
- Keep Mkdocs Version Fixed For Build Servers ([#2511](https://github.com/NetApp/harvest/pull/2511))
- Bump Go ([#2536](https://github.com/NetApp/harvest/pull/2536))
- Update Range In Query To 3H Before Validation ([#2571](https://github.com/NetApp/harvest/pull/2571))
- Bump Go ([#2582](https://github.com/NetApp/harvest/pull/2582))
- Template Validation For Rest, Restperf ([#2586](https://github.com/NetApp/harvest/pull/2586))
- Bump Dependencies ([#2600](https://github.com/NetApp/harvest/pull/2600))
- Detect Poller Logs Errors ([#2603](https://github.com/NetApp/harvest/pull/2603))
- Detect Poller Logs Errors ([#2609](https://github.com/NetApp/harvest/pull/2609))
- Fix Nightly Build ([#2630](https://github.com/NetApp/harvest/pull/2630))
- Bump Go And Dependencies ([#2649](https://github.com/NetApp/harvest/pull/2649))
- Disable Dockerfile Updtes By Renovate ([#2655](https://github.com/NetApp/harvest/pull/2655))
- Ignore Metrocluster Error In Counter Test ([#2664](https://github.com/NetApp/harvest/pull/2664))
- Bump go ([#2671](https://github.com/NetApp/harvest/pull/2671))
- Update makefile go version ([#2678](https://github.com/NetApp/harvest/pull/2678))

---

## 23.11.0 / 2023-11-13 Release
:pushpin: Highlights of this major release include:

- New FlexGroup dashboard that includes FlexGroup constituents. Thanks to @sandromuc and @ewilts for raising.

- Harvest [ChangeLog](https://netapp.github.io/harvest/latest/plugins/#changelog-plugin) plugin to detect and monitor changes related to object creation, modification, and deletion.

- We improved how Harvest calculates power. As a result, you may notice a decrease in the reported power metrics compared to previous versions. Details [here](https://netapp.github.io/harvest/latest/resources/power-algorithm/). Thanks to Evan Lee for reporting!

- Added `conf_path` variable for specifying the search path of Harvest templates.

- :package: Streamlined the Harvest container installation process by eliminating the need to download a tar file. Running Harvest in a container is now simpler and more convenient.

- :star: Several of the existing dashboards include new panels in this release:
  - Aggregate and Volume dashboard includes performance and capacity tier data. Thanks to @ewilts for raising.
  - Workload dashboard includes QoS fixed Utilization % panels. Thanks to @faguayot for raising.
  - Disk Dashboard features performance panels at the disk raid-group level. Thanks to @kinderr95 for raising.

- :ear_of_rice: Harvest includes new templates to collect:
  - Cloud target metrics. Thanks to @mamoep for raising
  - CIFS Share metrics. Thanks to @s-kuchi for raising
  - IWarp metrics are included in RestPerf
  - object_store_server metrics are included in RestPerf
  - SMB2 metrics are included in RestPerf

- :closed_book: Documentation additions
  - Enhanced Quickstart guide for Harvest
  - NABox logs collection guide
  - Document poller `ca_cert` property. Thanks to Marvin Montanus for reporting!
  - Describe how Harvest calculates power. Thanks to Evan Lee for reporting!
  - Details about hidden_fields and filter for the Rest Collector. Thanks to Johnathan Warlick for raising!

- Enhanced the Volume dashboard to include clone information.

- :zap: Optimized the Harvest binaries, significantly reducing their size.

- The Metadata dashboard works inside container deployments.

- The FabricPool panels in the Volume dashboard now support FlexGroup volumes. Thanks to @sriniji for reporting.

- Large `harvest.yml` files can be refactoring into smaller ones. Thanks to @llelik and @Pengng88 for raising.

- :bulb: Added help text about metrics to more Harvest dashboard panels.

## Announcements

:bangbang: **IMPORTANT** Due to ONTAP bug [1585893](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1585893) the Harvest team recommends using ZapiPerf instead of RestPerf when collecting performance metrics. The RestPerf collector can be used once you upgrade your cluster to a version of ONTAP with the fix. Details in 1585893.

:bangbang: **IMPORTANT** Release `23.11` disables the `CIFSSession` templates by default. This change was made to prevent the generation of a large number of metrics. If you require these templates, you can enable them. Please be aware that enabling them may result in a significant increase in metric collection time, Harvest memory footprint, and Prometheus used disk space. These metrics are utilized in the SMB2 dashboard.

:bangbang: **IMPORTANT** Release `23.11` has updated its power metric calculation algorithm. As a result, you may notice a decrease in the reported power metrics compared to previous versions. To collect these metrics, Rest API permissions are required. For detailed information on the power algorithm, please refer to the power algorithm [documentation](https://netapp.github.io/harvest/latest/resources/power-algorithm/).

:bangbang: **IMPORTANT** NetApp moved their communities from Slack to [Discord](https://discord.gg/ZmmWPHTBHw), please join us [there](https://discordapp.com/channels/855068651522490400/1001963189124206732)!

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the `bin/harvest grafana import` CLI, from the Grafana UI, or from the `Maintenance > Reset Harvest Dashboards` button in NAbox.

## Known Issues

- Some AFF A250 systems do not report power metrics. See ONTAP bug [1511476](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1511476) for more details.

- ONTAP does not include REST metrics for `offbox_vscan_server` and `offbox_vscan` until ONTAP 9.13.1. See ONTAP bug
  [1473892](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1473892) for more details.

**IMPORTANT** 7-mode filers that are not on the latest release of ONTAP may experience TLS connection issues with errors like `tls: server selected unsupported protocol version 301` This is caused by a change in Go 1.18. The [default for TLS client connections was changed to TLS 1.2](https://tip.golang.org/doc/go1.18#tls10) in Go 1.18. Please upgrade your 7-mode filers (recommended) or set `tls_min_version: tls10` in your `harvest.yml` [poller section](https://github.com/NetApp/harvest/tree/release/22.05.0#pollers). See [#1007](https://github.com/NetApp/harvest/issues/1007) for more details.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards
this release:

@Garydep, @MrObvious, @Pengng88, @Sandromuc, @ewilts, @faguayot, @jmg011, @kinderr95, @llelik, @mamoep, @rodenj1, @s-kuchi, @shawnahall71, @slater0013, @sriniji, @statdigger, @wyahn1219, AlessandroN, Dave, Diane, Evan Lee, Francesco, Heaven7, Johnathan Warlick, Madaan, Martijn Moret, Marvin Montanus, NicoSeiberth, RBrown, TonyHsieh, Watson9121, dbakerletn, imthenightbird, roller, twodot0h, tymercer

:seedling: This release includes 38 features, 26 bug fixes, 24 documentation, 5 performance, 2 refactoring, 12 miscellaneous, and 7 ci pull requests.

### :rocket: Features
- Change Log Detection In Harvest ([#2178](https://github.com/NetApp/harvest/pull/2178))
- Remove Daemon Dependency ([#2195](https://github.com/NetApp/harvest/pull/2195))
- Enable More Golanglint Linters ([#2313](https://github.com/NetApp/harvest/pull/2313))
- Gcc Is Not Required To Build Harvest ([#2322](https://github.com/NetApp/harvest/pull/2322))
- Ontap Permission Errors Should Be Logged As Errors ([#2326](https://github.com/NetApp/harvest/pull/2326))
- Harvest Should Load Templates From A Set Of Conf Directories ([#2329](https://github.com/NetApp/harvest/pull/2329))
- Ontap Power Calculation For Embedded Shelf ([#2333](https://github.com/NetApp/harvest/pull/2333))
- Enable More Golanglint Linters ([#2334](https://github.com/NetApp/harvest/pull/2334))
- Harvest Auto-Support Should Include Instance Count In Collector Section ([#2337](https://github.com/NetApp/harvest/pull/2337))
- Set Allvalue To Null When Svm Regex Is Applied ([#2340](https://github.com/NetApp/harvest/pull/2340))
- Add Parity For String Types Between Restperf And Zapiperf ([#2342](https://github.com/NetApp/harvest/pull/2342))
- Tiering Data Changes For Volume - Template Change ([#2343](https://github.com/NetApp/harvest/pull/2343))
- Docker Workflow Doesn't Need Tar Download ([#2354](https://github.com/NetApp/harvest/pull/2354))
- Enable Ports By Default In Docker Generate ([#2360](https://github.com/NetApp/harvest/pull/2360))
- Support Comma Separated Aggrs In Perf Metrics ([#2376](https://github.com/NetApp/harvest/pull/2376))
- Harvest Should Support Multiple Poller Files To Allow Refactori… ([#2388](https://github.com/NetApp/harvest/pull/2388))
- Adding Iwarp Restperf Template ([#2390](https://github.com/NetApp/harvest/pull/2390))
- Adding New Panels In Disk Dashboard ([#2391](https://github.com/NetApp/harvest/pull/2391))
- Harvest Should Load Templates From A Set Of Conf Directories ([#2394](https://github.com/NetApp/harvest/pull/2394))
- Add Api To Rest Error Log ([#2401](https://github.com/NetApp/harvest/pull/2401))
- Add Clone Info To Volume Dashboard ([#2402](https://github.com/NetApp/harvest/pull/2402))
- Cifs Share Templates ([#2405](https://github.com/NetApp/harvest/pull/2405))
- Support Flexgroup Constituents In Template ([#2410](https://github.com/NetApp/harvest/pull/2410))
- Add Flexgroup To Fabricpool Panels ([#2419](https://github.com/NetApp/harvest/pull/2419))
- Smb2 Restperf Counters ([#2420](https://github.com/NetApp/harvest/pull/2420))
- Adding Fc Rest Template For Fibre Channel Switch ([#2424](https://github.com/NetApp/harvest/pull/2424))
- Metric Doc Needs To Handle Templates With Same Object Names ([#2426](https://github.com/NetApp/harvest/pull/2426))
- Antiransomwarestate Label Should Be Exported ([#2432](https://github.com/NetApp/harvest/pull/2432))
- Metadata Dashboard Should Work With Containers And Remove System Resources Panel ([#2433](https://github.com/NetApp/harvest/pull/2433))
- Adding Restperf Object_store_server Template ([#2435](https://github.com/NetApp/harvest/pull/2435))
- Update Ci To Use Docker Run And Update Permissions ([#2436](https://github.com/NetApp/harvest/pull/2436))
- Enable More Golanglint Linters ([#2439](https://github.com/NetApp/harvest/pull/2439))
- Qos Fixed Utilization % Panels ([#2445](https://github.com/NetApp/harvest/pull/2445))
- Description Fetched From Ontap Docs Via Cli ([#2454](https://github.com/NetApp/harvest/pull/2454))
- Disable Cifssession Template ([#2455](https://github.com/NetApp/harvest/pull/2455))
- Add Labels Defined In Harvest Config To Metadata Metrics ([#2456](https://github.com/NetApp/harvest/pull/2456))
- Add Link_up Counter For Fcp ([#2464](https://github.com/NetApp/harvest/pull/2464))
- Implementing Support For Randomized Start Times In Tasks ([#2465](https://github.com/NetApp/harvest/pull/2465))

### :bug: Bug Fixes
- Qos Policy Not Updated In Workload Counters ([#2318](https://github.com/NetApp/harvest/pull/2318))
- Bin Dir Prints Error When There Are No Files In It ([#2324](https://github.com/NetApp/harvest/pull/2324))
- Harvest Docker Generate Should Prefix Relative Paths With Dot ([#2346](https://github.com/NetApp/harvest/pull/2346))
- Sg Overview Dashboard Should Use Percentunit For `Top X Tenants … ([#2348](https://github.com/NetApp/harvest/pull/2348))
- Log Error When Config File Is Not Found During Generate Cli ([#2349](https://github.com/NetApp/harvest/pull/2349))
- Include Cloud_target Template For Zapi Collector ([#2350](https://github.com/NetApp/harvest/pull/2350))
- Sg Fabricpool Dashboard Should Use Bucket Name Variable ([#2353](https://github.com/NetApp/harvest/pull/2353))
- Add Quotes For Ports In Docker Compose Files ([#2355](https://github.com/NetApp/harvest/pull/2355))
- Remove Only Harvest Related Containers In Make Ci-Local ([#2356](https://github.com/NetApp/harvest/pull/2356))
- Harvest Power Algorithm Should Handle Shared Psus ([#2359](https://github.com/NetApp/harvest/pull/2359))
- Don't Expose End-Points With Port=False ([#2361](https://github.com/NetApp/harvest/pull/2361))
- Enable Root Aggregate Power ([#2363](https://github.com/NetApp/harvest/pull/2363))
- Read Templates From `Harvest_conf` When Set ([#2366](https://github.com/NetApp/harvest/pull/2366))
- Read Harvest.yml From `Harvest_conf` When Set ([#2367](https://github.com/NetApp/harvest/pull/2367))
- Sg Collector Throws Npe ([#2384](https://github.com/NetApp/harvest/pull/2384))
- Lif Template Should Use Unique Instancekeys ([#2393](https://github.com/NetApp/harvest/pull/2393))
- Don't Schedule A Task More Frequently Than Its Default Schedule ([#2408](https://github.com/NetApp/harvest/pull/2408))
- `Volume Iops Per Type` Legend Should Use Mean, Last, Max Instead… ([#2411](https://github.com/NetApp/harvest/pull/2411))
- Override In Templates Should Be From Counters ([#2417](https://github.com/NetApp/harvest/pull/2417))
- Make Zapi Rest Cli Consistent For Max-Records Arg ([#2447](https://github.com/NetApp/harvest/pull/2447))
- Exclude Dirs For Docker Generate ([#2448](https://github.com/NetApp/harvest/pull/2448))
- Support Multi Key For Disk Plugin Zapiperf ([#2449](https://github.com/NetApp/harvest/pull/2449))
- Handled Cloud_target For Fabricpool ([#2467](https://github.com/NetApp/harvest/pull/2467))
- Harvest.yml Defaults Should Be Applied To Child Harvest.yml ([#2471](https://github.com/NetApp/harvest/pull/2471))
- Update Flexgroup Text In Dashboard ([#2474](https://github.com/NetApp/harvest/pull/2474))
- Handled Missing Uuid In Volume For Change_log ([#2479](https://github.com/NetApp/harvest/pull/2479))

### :closed_book: Documentation
- Add Workload Information To Release Notes ([#2316](https://github.com/NetApp/harvest/pull/2316))
- Describe How To Collect Nabox Logs ([#2321](https://github.com/NetApp/harvest/pull/2321))
- Add Storagegrid To Extend Docs ([#2336](https://github.com/NetApp/harvest/pull/2336))
- Add Details About Hidden_fields And Filter For Rest Collector ([#2341](https://github.com/NetApp/harvest/pull/2341))
- Describe How Harvest Calculates Power ([#2362](https://github.com/NetApp/harvest/pull/2362))
- Describe How To Upgrade To Nightly Image ([#2370](https://github.com/NetApp/harvest/pull/2370))
- Fix Nightly Image Docker Generate ([#2372](https://github.com/NetApp/harvest/pull/2372))
- Document Poller `Ca_cert` Property ([#2374](https://github.com/NetApp/harvest/pull/2374))
- Harvest Configuration File From Other Location With Docker Run Generate ([#2377](https://github.com/NetApp/harvest/pull/2377))
- Fix Link To Issue ([#2382](https://github.com/NetApp/harvest/pull/2382))
- Update Container Docs ([#2397](https://github.com/NetApp/harvest/pull/2397))
- Add Download Swagger Permission ([#2398](https://github.com/NetApp/harvest/pull/2398))
- Add Curl To Twisty ([#2406](https://github.com/NetApp/harvest/pull/2406))
- `Generate Metrics` Should Include Metrics Created By Builtin Plu… ([#2413](https://github.com/NetApp/harvest/pull/2413))
- Fix Cluster_new_status Metric ([#2423](https://github.com/NetApp/harvest/pull/2423))
- Add Build Docker Image Section And Fix Links ([#2425](https://github.com/NetApp/harvest/pull/2425))
- Fix Docker Build Path ([#2429](https://github.com/NetApp/harvest/pull/2429))
- Fix Typo In Ems Docs ([#2431](https://github.com/NetApp/harvest/pull/2431))
- Example Auto-Support Payload Should Include Rest Example ([#2442](https://github.com/NetApp/harvest/pull/2442))
- Quickstart Should Mention Install Section ([#2446](https://github.com/NetApp/harvest/pull/2446))
- Changelog Doc ([#2453](https://github.com/NetApp/harvest/pull/2453))
- Generate Ordered Api List For Metrics Doc ([#2458](https://github.com/NetApp/harvest/pull/2458))
- Fix Docker Run Steps ([#2473](https://github.com/NetApp/harvest/pull/2473))
- Describe `Conf_path` ([#2480](https://github.com/NetApp/harvest/pull/2480))

### :zap: Performance
- Rest Client Allocs Improvements ([#2381](https://github.com/NetApp/harvest/pull/2381))
- Use Parse Instead Of Getmany For Gjson ([#2385](https://github.com/NetApp/harvest/pull/2385))
- Remove Extra Dict Wrapper ([#2396](https://github.com/NetApp/harvest/pull/2396))
- Avoid Byte Slice To String Conversion ([#2407](https://github.com/NetApp/harvest/pull/2407))

### Refactoring
- Simplify Conf Path Management ([#2325](https://github.com/NetApp/harvest/pull/2325))
- Update Href Creation ([#2399](https://github.com/NetApp/harvest/pull/2399))

### Miscellaneous
- Minor Release Issue Fixes ([#2293](https://github.com/NetApp/harvest/pull/2293))
- Release Doc Update ([#2294](https://github.com/NetApp/harvest/pull/2294))
- Merge 23.08.0 To Main ([#2310](https://github.com/NetApp/harvest/pull/2310))
- Init Bool Variable ([#2319](https://github.com/NetApp/harvest/pull/2319))
- Update All Dependencies ([#2339](https://github.com/NetApp/harvest/pull/2339))
- Update Actions/Checkout Action To V4 ([#2358](https://github.com/NetApp/harvest/pull/2358))
- Update Module Github.com/Tidwall/Gjson To V1.17.0 ([#2378](https://github.com/NetApp/harvest/pull/2378))
- Remove Drilldown From Panel Titles ([#2387](https://github.com/NetApp/harvest/pull/2387))
- Update All Dependencies ([#2395](https://github.com/NetApp/harvest/pull/2395))
- Update All Dependencies ([#2414](https://github.com/NetApp/harvest/pull/2414))
- Bump Golang.org/X/Net From 0.14.0 To 0.17.0 In /Integration ([#2422](https://github.com/NetApp/harvest/pull/2422))
- Update All Dependencies ([#2459](https://github.com/NetApp/harvest/pull/2459))

### :hammer: CI
- Create Draft Release Highlights ([#2314](https://github.com/NetApp/harvest/pull/2314))
- Bump Go And Dependencies ([#2351](https://github.com/NetApp/harvest/pull/2351))
- Reduce Artifacts Size ([#2403](https://github.com/NetApp/harvest/pull/2403))
- Bump Go ([#2409](https://github.com/NetApp/harvest/pull/2409))
- Bump Go ([#2418](https://github.com/NetApp/harvest/pull/2418))
- Remove Unused Volumes From Ci Machines ([#2440](https://github.com/NetApp/harvest/pull/2440))
- Bump Go ([#2469](https://github.com/NetApp/harvest/pull/2469))

---

## 23.08.0 / 2023-08-21 Release
:pushpin: Highlights of this major release include:
- Harvest Security dashboard highlights compliance using [NetApp's Security hardening guide for ONTAP](https://www.netapp.com/media/10674-tr4569.pdf)

- Harvest's credential script supports ONTAP daily credential rotation. Thanks to @mamoep for raising.

- :tophat: Harvest makes it easy to run with both the [ZAPI and REST collectors](https://netapp.github.io/harvest/latest/architecture/rest-strategy/) at the same time. Overlapping resources are deduplicated and only exported once. Harvest will automatically upgrade ZAPI conversations to REST when ZAPIs are suspended or disabled.

- :gem: Updated workload dashboard now includes Service Center, Latency Breakdown, and 50 panels

- :gem: Cluster dashboard updated to work with FSx. Some panels are blank because FSx does not have that data.

- :mega: The Harvest team published a couple of screencasts about:
  - [Why Harvest](https://youtu.be/04-66_9egJc)
  - [Harvest Quick Start: Docker Compose](https://youtu.be/4cbDKzwjGHI)

- :star: Several of the existing dashboards include new panels in this release:
  - Aggregate dashboard includes busy volume panels
  - SVM dashboard includes per NFS latency heatmaps. Thanks to @rbrownATnetapp for raising.
  - Volume dashboard includes top resources by other IOPs panel and junction paths. Thanks to @tsohst for raising.

- All Harvest dashboard tables include column filters
- Harvest dashboards use color to highlight latency and busy threshold breaches
- Harvest's Prometheus exporter supports TLS

- :ear_of_rice: Harvest includes new templates to collect:
  - Iwarp metrics
  - FCVI metrics
  - Per volume NFS metrics
  - Volume clone metrics
  - QoS workload policy metrics
  - NVME/TCP and NVME/RoCE metrics
  - Flashpool metrics are included in RestPerf. Thanks to @lobster1860 for raising

- :closed_book: Documentation additions
  - Move more documentation from GitHub to [Harvest documentation site](https://netapp.github.io/harvest/)
  - Clarify how to tell Harvest to continue using the ZAPI protocol
  - Clarify generic vs custom plugins. Thanks to GregS for raising
  - Clarify which version of Go is required to build Harvest. Thanks to MikeK for raising
  - Clarify how to prepare ONTAP cDOT clusters for Harvest data collection
  - EMS documentation should point to Harvest documentation site. Thanks to @cwaltham for raising
  - Clarify how to gather log files on all platforms
  - Explain how to use the `--labels` option of `bin/harvest grafana`. Thanks to @slater0013 for raising
  - Describe how to run docker compose generate command without required Harvest binaries

- The Harvest `doctor` command validates collector names listed in your `harvest.yml` file

- An earlier version of Harvest collected cloud store information via REST. This release adds the same for ZAPI

- When ONTAP resources are missing, Harvest tries to collect them every hour. Earlier versions of Harvest waited 24 hours before retrying, which often caused metrics to be missing after a cluster upgrade. Thanks to @Falcon667 for raising

- Earlier versions of Harvest created world writable auto-support files. These files are now only read/writeable by the current user. Thanks to Bunnygirl for raising

- `bin/harvest import` should work with Grafana 10. Thanks to @wooyoungAhn for raising

## Announcements

:bangbang: **IMPORTANT** `23.08` fixes a REST collector bug that caused partial data collection when ONTAP paginated results. See #2109 for details.

:bangbang: **IMPORTANT** Release `23.08` disables the `NetConnections` and `NFSClients` templates by default. You can enable them if needed. These templates were disabled because several customers reported that these templates created millions of metrics. None of these metrics are used in Harvest dashboards.

:bangbang: **IMPORTANT** Release `23.08` changes how Harvest monitors workloads. For detailed information, please refer to the discussion #2265.

:bulb: The Compliance dashboard was removed after its panels were moved to the Security dashboard.

:eyes: Ambient temperature metric may experience an increase due to issue #2259

:bangbang: **IMPORTANT** NetApp moved their communities from Slack to [Discord](https://discord.gg/ZmmWPHTBHw), please join us [there](https://discordapp.com/channels/855068651522490400/1001963189124206732)!

:bangbang: **IMPORTANT** If using Docker Compose and you want to keep your historical Prometheus data, please
read [how to migrate your Prometheus volume](https://github.com/NetApp/harvest/blob/main/docs/MigratePrometheusDocker.md)

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the `bin/harvest grafana import` CLI, from the Grafana UI, or from the `Maintenance > Reset Harvest Dashboards` button in NAbox.

## Known Issues

- Some AFF A250 systems do not report power metrics. See ONTAP bug [1511476](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1511476) for more details.

- ONTAP does not include REST metrics for `offbox_vscan_server` and `offbox_vscan` until ONTAP 9.13.1. See ONTAP bug
  [1473892](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1473892) for more details.

**IMPORTANT** 7-mode filers that are not on the latest release of ONTAP may experience TLS connection issues with errors like `tls: server selected unsupported protocol version 301` This is caused by a change in Go 1.18. The [default for TLS client connections was changed to TLS 1.2](https://tip.golang.org/doc/go1.18#tls10) in Go 1.18. Please upgrade your 7-mode filers (recommended) or set `tls_min_version: tls10` in your `harvest.yml` [poller section](https://github.com/NetApp/harvest/tree/release/22.05.0#pollers). See [#1007](https://github.com/NetApp/harvest/issues/1007) for more details.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards
this release:

@7840vz, @DAx-cGn, @Falcon667, @Hedius, @LukaszWasko, @MrObvious, @ReneMeier, @Sawall10, @T1r0l, @XDavidT, @amd-eulee, @aticatac, @chadpruden, @cwaltham, @cygio, @ddhti, @debert-ntap, @demalik, @electrocreative, @elsgaard, @ev1963, @faguayot, @iStep2Step, @jgasher, @jmg011, @lobster1860, @mamoep, @matejzero, @matthieu-sudo, @merdos, @pilot7777, @rbrownATnetapp, @rodenj1, @slater0013, @swordfish291, @tsohst, @wooyoungAhn, Alessandro.Nuzzo, Ed Wilts, GregS, Imthenightbird, KlausHub, MeghanaD, MikeK, Paul P2, Rusty Brown, Shubham Mer, Tudor Pascu, Watson9121, jf38800, jfong, lorenzoc, rcl23, roller, scrhobbs, troysmuller, twodot0h

:seedling: This release includes 42 features, 40 bug fixes, 20 documentation, 2 performance, 4 testing, 1 styling, 9 refactoring, 20 miscellaneous, and 12 ci pull requests.

### :rocket: Features
- Harvest Should Collect Iwarp Counters ([#2071](https://github.com/NetApp/harvest/pull/2071))
- Update Visitpanels To Be Recursive ([#2085](https://github.com/NetApp/harvest/pull/2085))
- Add Table Column Filter For Dashboards ([#2088](https://github.com/NetApp/harvest/pull/2088))
- Update Lagtime Based On Lasttransfersize ([#2091](https://github.com/NetApp/harvest/pull/2091))
- Harvest Should Add Grafana Import Rewrite Svm Filtering For Multi-Tenant Support ([#2092](https://github.com/NetApp/harvest/pull/2092))
- Fetch Cloud_store Info In Zapi Via Plugin ([#2094](https://github.com/NetApp/harvest/pull/2094))
- Collection Of Other Counters For Fcvi Perf Object ([#2096](https://github.com/NetApp/harvest/pull/2096))
- Add Nfs Io Types At The Volume Level ([#2098](https://github.com/NetApp/harvest/pull/2098))
- Add System Defined Workload Collection ([#2099](https://github.com/NetApp/harvest/pull/2099))
- Add Workload Panels In Workload Dashboard ([#2100](https://github.com/NetApp/harvest/pull/2100))
- Add Volume Clone Info In Rest ([#2102](https://github.com/NetApp/harvest/pull/2102))
- Added Volume Panels In Aggr Dashboard ([#2104](https://github.com/NetApp/harvest/pull/2104))
- Workload Policy Iops Metrics ([#2111](https://github.com/NetApp/harvest/pull/2111))
- Autoresolve Ems Would Export Metric Value As 0 And Autoresolve=True Label ([#2120](https://github.com/NetApp/harvest/pull/2120))
- Support Type Label For Volume For Backward Compatibility ([#2132](https://github.com/NetApp/harvest/pull/2132))
- Volume Clone Info For Zapi ([#2140](https://github.com/NetApp/harvest/pull/2140))
- Harvest Should Include Numpollers And Rss In Autosupport ([#2143](https://github.com/NetApp/harvest/pull/2143))
- Colors In Grafana Dashboards To Highlight Warning, Critical Severity ([#2147](https://github.com/NetApp/harvest/pull/2147))
- Security Hardening Guide ([#2150](https://github.com/NetApp/harvest/pull/2150))
- Harvest Prometheus Exporter Should Support Tls ([#2153](https://github.com/NetApp/harvest/pull/2153))
- Latency Units Should Be In Microseconds In Harvest Dashboard ([#2156](https://github.com/NetApp/harvest/pull/2156))
- Simplify Rest Auto-Upgrade ([#2167](https://github.com/NetApp/harvest/pull/2167))
- When Using A Credential Script, Re-Auth On 401S ([#2180](https://github.com/NetApp/harvest/pull/2180))
- Upgrade Zapi Conversations To Rest When Zapis Are Suspended Or … ([#2200](https://github.com/NetApp/harvest/pull/2200))
- When Using A Credential Script, Re-Auth On 401S ([#2203](https://github.com/NetApp/harvest/pull/2203))
- Merge Compliance And Security Dashboard + Added Arw Fields ([#2207](https://github.com/NetApp/harvest/pull/2207))
- Supporting Topk In S3 Dashboard ([#2208](https://github.com/NetApp/harvest/pull/2208))
- Aff250 Power Calculation ([#2211](https://github.com/NetApp/harvest/pull/2211))
- Use Single `Go Build` Command To Build Harvest And Poller Binaries ([#2221](https://github.com/NetApp/harvest/pull/2221))
- Harvest Should Include A User Agent ([#2224](https://github.com/NetApp/harvest/pull/2224))
- Add Collector Name Validation In Doctor ([#2229](https://github.com/NetApp/harvest/pull/2229))
- Harvest Should Fetch Certificates Via A Script ([#2238](https://github.com/NetApp/harvest/pull/2238))
- Include Lun Offline Ems Alert ([#2252](https://github.com/NetApp/harvest/pull/2252))
- Add Panel For Other Iops On Volume Dashboard ([#2254](https://github.com/NetApp/harvest/pull/2254))
- Update Ambient Temperature Calculation For Power Dashboard ([#2259](https://github.com/NetApp/harvest/pull/2259))
- Nvme/Tcp And Nvme/Roce Counters ([#2264](https://github.com/NetApp/harvest/pull/2264))
- Harvest Svm Dashboard Should Include Latency Heatmap Panels Nfs… ([#2268](https://github.com/NetApp/harvest/pull/2268))
- Added Table Description For Cluster Compliance ([#2269](https://github.com/NetApp/harvest/pull/2269))
- Update Ontap Metric Document ([#2270](https://github.com/NetApp/harvest/pull/2270))
- Add Cpu_firmware_release To Cluster Dashboard ([#2274](https://github.com/NetApp/harvest/pull/2274))
- Enable Cluster Dashboard For Fsx ([#2303](https://github.com/NetApp/harvest/pull/2303))
- Add Junction Paths In Volumes Dashboard ([#2309](https://github.com/NetApp/harvest/pull/2309))

### :bug: Bug Fixes
- Disk Dashboard Power On Time Should Use `Seconds` Unit ([#2039](https://github.com/NetApp/harvest/pull/2039))
- Update Metadata Cpu Times: Breakdown To Seconds ([#2055](https://github.com/NetApp/harvest/pull/2055))
- Workload Missing Label Value ([#2072](https://github.com/NetApp/harvest/pull/2072))
- Fcvi Restperf Template ([#2080](https://github.com/NetApp/harvest/pull/2080))
- Change Svm Panels Row Name ([#2097](https://github.com/NetApp/harvest/pull/2097))
- Correct Unit In Panels With Added Testcase ([#2108](https://github.com/NetApp/harvest/pull/2108))
- Rest Collector Incomplete Data If Retrieval Exceeds Return_timeout ([#2110](https://github.com/NetApp/harvest/pull/2110))
- Storagegrid Should Honor `-Logtofile` Option ([#2119](https://github.com/NetApp/harvest/pull/2119))
- Harvest Should Always Pass `Addr` Argument To Credentials_script ([#2128](https://github.com/NetApp/harvest/pull/2128))
- Handle Difference Of Pollinstance And Polldata Records Via Exportable ([#2137](https://github.com/NetApp/harvest/pull/2137))
- Cpu_busy Description In Cluster Dashboard ([#2141](https://github.com/NetApp/harvest/pull/2141))
- Reduce Auto Support Log Noise When Collecting Process Info On Mac ([#2145](https://github.com/NetApp/harvest/pull/2145))
- Correct The Flashpool Panel Units ([#2163](https://github.com/NetApp/harvest/pull/2163))
- Handling Label Count When Matches Applied In Ems ([#2165](https://github.com/NetApp/harvest/pull/2165))
- Volume Template Fix ([#2171](https://github.com/NetApp/harvest/pull/2171))
- Harvest Should Retry Every Hour When Ontap Replies With An Api-R… ([#2181](https://github.com/NetApp/harvest/pull/2181))
- Ciphers Query Was Giving Wrong Result In Promql ([#2188](https://github.com/NetApp/harvest/pull/2188))
- S3 Dashboard Fails To Import In Grafana 8.5.15 ([#2191](https://github.com/NetApp/harvest/pull/2191))
- Harvest Auto-Support Files Should Not Be World Writable ([#2193](https://github.com/NetApp/harvest/pull/2193))
- Fix Key For Qtree 7Mode ([#2196](https://github.com/NetApp/harvest/pull/2196))
- Check Existing Asup Dir Permission ([#2197](https://github.com/NetApp/harvest/pull/2197))
- Import Dashboard Failure With Editor Role In Grafana ([#2206](https://github.com/NetApp/harvest/pull/2206))
- When Using Credentials_file Make Sure Defaults Are Copied To Poller ([#2209](https://github.com/NetApp/harvest/pull/2209))
- When Using Credentials_file Make Sure Defaults Are Copied To Poller ([#2215](https://github.com/NetApp/harvest/pull/2215))
- Flashpool-Data Is Missing In Restperf ([#2217](https://github.com/NetApp/harvest/pull/2217))
- Disable Nfs_clients.yaml Template By Default In Rest Collector ([#2219](https://github.com/NetApp/harvest/pull/2219))
- Remove Duplicate Error Message ([#2222](https://github.com/NetApp/harvest/pull/2222))
- Correct Svm Rest Template Based On Version ([#2239](https://github.com/NetApp/harvest/pull/2239))
- Correct Shelf Metrics In 7Mode ([#2245](https://github.com/NetApp/harvest/pull/2245))
- Remove Source_node Label From Snapmirror Zapi ([#2255](https://github.com/NetApp/harvest/pull/2255))
- Added Version Check For Aggr-Object-Store-Get-Iter ([#2258](https://github.com/NetApp/harvest/pull/2258))
- Volume Rest Template Based On Version ([#2263](https://github.com/NetApp/harvest/pull/2263))
- Nfs Heatmap Per Cluster ([#2273](https://github.com/NetApp/harvest/pull/2273))
- Make Poller Mandatory For Metrics Generation Cmd ([#2280](https://github.com/NetApp/harvest/pull/2280))
- Handled When Metric Not Found In Plugin ([#2281](https://github.com/NetApp/harvest/pull/2281))
- Disable Netconnections In Rest By Default ([#2283](https://github.com/NetApp/harvest/pull/2283))
- Grafana Ask-For-Token Should Retry At Most 5 Times ([#2284](https://github.com/NetApp/harvest/pull/2284))
- Match Object Name With Zapiperf For Cifs_vserver.yaml ([#2288](https://github.com/NetApp/harvest/pull/2288))
- Add Bin Dir Check Before Removing Files ([#2289](https://github.com/NetApp/harvest/pull/2289))
- Adding Log Forwarding Column In Compliance Table In Security Dashboard ([#2306](https://github.com/NetApp/harvest/pull/2306))

### :closed_book: Documentation
- Explain Bin/Grafana Import --Labels ([#2032](https://github.com/NetApp/harvest/pull/2032))
- Update Release Checklist ([#2043](https://github.com/NetApp/harvest/pull/2043))
- Update Docker Compose Generation Process To Remove Binary Dependencies ([#2046](https://github.com/NetApp/harvest/pull/2046))
- Add Details About Volume Sis Stat Panel ([#2047](https://github.com/NetApp/harvest/pull/2047))
- Add Harvest-Metrics Release Branch Creation For Release Steps ([#2050](https://github.com/NetApp/harvest/pull/2050))
- Fix Rest Template Extend Instructions Path ([#2051](https://github.com/NetApp/harvest/pull/2051))
- Fsx Does Not Support Headroom Dashboard ([#2131](https://github.com/NetApp/harvest/pull/2131))
- Update Fsa Dashboard Doc ([#2159](https://github.com/NetApp/harvest/pull/2159))
- Move K8 Podman Document To Documentation Site ([#2160](https://github.com/NetApp/harvest/pull/2160))
- Clarify How To Tell Harvest To Continue Using The Zapi Protocol ([#2162](https://github.com/NetApp/harvest/pull/2162))
- Clarify Generic Vs Custom Plugins ([#2166](https://github.com/NetApp/harvest/pull/2166))
- Update Docker Docs Link To Doc Site ([#2186](https://github.com/NetApp/harvest/pull/2186))
- Clarify Which Version Of Go Is Required ([#2214](https://github.com/NetApp/harvest/pull/2214))
- Give Authentication Precedence Its Own Section ([#2226](https://github.com/NetApp/harvest/pull/2226))
- Add Note About Workload Counter In Default Templates ([#2230](https://github.com/NetApp/harvest/pull/2230))
- Simplify The Preparing Ontap Cdot Cluster Documentation ([#2231](https://github.com/NetApp/harvest/pull/2231))
- Fix Ems Link ([#2244](https://github.com/NetApp/harvest/pull/2244))
- Update Metric Generate Step Command ([#2279](https://github.com/NetApp/harvest/pull/2279))
- Move Troubleshoot Docs To Doc Site ([#2287](https://github.com/NetApp/harvest/pull/2287))
- Release 23.08 Metric Docs ([#2290](https://github.com/NetApp/harvest/pull/2290))

### :zap: Performance
- Improve Memory And Cpu Performance Of Restperf Collector ([#2053](https://github.com/NetApp/harvest/pull/2053))
- Optimize Restperf Collector Pollinstance ([#2121](https://github.com/NetApp/harvest/pull/2121))

### :wrench: Testing
- Add Unit Test For Restperf ([#2044](https://github.com/NetApp/harvest/pull/2044))
- Adding Ems Unit Tests ([#2052](https://github.com/NetApp/harvest/pull/2052))
- Add Unit Test For Rest Collector ([#2062](https://github.com/NetApp/harvest/pull/2062))
- Ensure Dashboard Time Is Now-3H ([#2275](https://github.com/NetApp/harvest/pull/2275))

### Styling
- Address All Lint Errors In Ci ([#2014](https://github.com/NetApp/harvest/pull/2014))

### Refactoring
- Move Unit Testing Json Parser To Common ([#2064](https://github.com/NetApp/harvest/pull/2064))
- Dashboard Tests ([#2090](https://github.com/NetApp/harvest/pull/2090))
- Harvest Dashboard Jsons Should Be Sorted By Key ([#2152](https://github.com/NetApp/harvest/pull/2152))
- Eliminate Usages Of Time.sleep In Test Code ([#2182](https://github.com/NetApp/harvest/pull/2182))
- Fix Inconsistent Pointer Receivers ([#2225](https://github.com/NetApp/harvest/pull/2225))
- Reduce Asup Log Noise ([#2276](https://github.com/NetApp/harvest/pull/2276))
- Increase Max Log File Size From 5Mb To 10Mb ([#2277](https://github.com/NetApp/harvest/pull/2277))
- Add Cp Command In Dashboard Sort Test ([#2278](https://github.com/NetApp/harvest/pull/2278))
- Code Cleanup ([#2282](https://github.com/NetApp/harvest/pull/2282))

### Miscellaneous
- Bump Github.com/Shirou/Gopsutil/V3 From 3.23.3 To 3.23.4 ([#2027](https://github.com/NetApp/harvest/pull/2027))
- Bump Golang.org/X/Term From 0.7.0 To 0.8.0 ([#2056](https://github.com/NetApp/harvest/pull/2056))
- Bump Golang.org/X/Sys From 0.7.0 To 0.8.0 ([#2057](https://github.com/NetApp/harvest/pull/2057))
- Add Renovate Bot ([#2075](https://github.com/NetApp/harvest/pull/2075))
- Update Module Github.com/Imdario/Mergo To V0.3.16 ([#2112](https://github.com/NetApp/harvest/pull/2112))
- Update Renovate Bot ([#2116](https://github.com/NetApp/harvest/pull/2116))
- Update Renovate Commit Prefix ([#2117](https://github.com/NetApp/harvest/pull/2117))
- Update Module Github.com/Shirou/Gopsutil/V3 To V3.23.5 ([#2122](https://github.com/NetApp/harvest/pull/2122))
- Update All Dependencies ([#2139](https://github.com/NetApp/harvest/pull/2139))
- Update Module Github.com/Imdario/Mergo To V1 ([#2144](https://github.com/NetApp/harvest/pull/2144))
- Upgrade Mergo Package ([#2157](https://github.com/NetApp/harvest/pull/2157))
- Update Module Github.com/Shirou/Gopsutil/V3 To V3.23.6 ([#2174](https://github.com/NetApp/harvest/pull/2174))
- Update All Dependencies ([#2176](https://github.com/NetApp/harvest/pull/2176))
- Update Module Golang.org/X/Term To V0.10.0 ([#2183](https://github.com/NetApp/harvest/pull/2183))
- Bump Go ([#2205](https://github.com/NetApp/harvest/pull/2205))
- Update All Dependencies ([#2243](https://github.com/NetApp/harvest/pull/2243))
- Bump Go ([#2253](https://github.com/NetApp/harvest/pull/2253))
- Update All Dependencies ([#2261](https://github.com/NetApp/harvest/pull/2261))
- Bump Go ([#2285](https://github.com/NetApp/harvest/pull/2285))
- Update Module Github.com/Tidwall/Gjson To V1.16.0 ([#2286](https://github.com/NetApp/harvest/pull/2286))

### :hammer: CI
- Wait For Qos_volume Counters ([#2045](https://github.com/NetApp/harvest/pull/2045))
- Update Docs For Nightly Builds ([#2058](https://github.com/NetApp/harvest/pull/2058))
- Add Gh-Pages Fetch Before Mkdoc Deploy ([#2067](https://github.com/NetApp/harvest/pull/2067))
- Configure Renovate ([#2074](https://github.com/NetApp/harvest/pull/2074))
- Renovate Should Ignore Integration ([#2078](https://github.com/NetApp/harvest/pull/2078))
- Renovate Should Run On A Schedule ([#2082](https://github.com/NetApp/harvest/pull/2082))
- Renovate Group All Prs ([#2136](https://github.com/NetApp/harvest/pull/2136))
- Ensure Exported Prometheus Metrics Are Unique ([#2173](https://github.com/NetApp/harvest/pull/2173))
- Run Renovate Once In A Week ([#2185](https://github.com/NetApp/harvest/pull/2185))
- Include Harvest Certification Tool ([#2241](https://github.com/NetApp/harvest/pull/2241))
- Fix Local Ci Errors ([#2266](https://github.com/NetApp/harvest/pull/2266))
- Remove Apt-Get Update ([#2271](https://github.com/NetApp/harvest/pull/2271))

---

## 23.05.0 / 2023-05-03
:pushpin: Highlights of this major release include:
- :gem: Seven new dashboards:
  - StorageGRID and ONTAP fabric pool
  - Health
  - S3 object storage
  - External service operations
  - Namespace
  - SMB
  - Workloads

- :star: Several of the existing dashboards include new panels in this release:
  - Qtree dashboard includes topK qtrees by disk-used growth
  - StorageGRID Overview dashboard includes traffic classification panels
  - Network dashboard includes net routes
  - Average CPU utilization and CPU busy are included in the cDOT, Cluster, Node, and Metrocluster dashboards
  - SVM dashboard includes LIF counters and the NFS panels filter graphs by NFS version
  - Volume dashboard includes efficiency statistics
  - Aggregate dashboard includes the amount of free space
  - Compliance dashboard only reports on data SVMs

- :closed_lock_with_key: Harvest can fetch cluster credentials via a [credential script](https://netapp.github.io/harvest/23.05/configure-harvest-basic/#credentials-script). Thanks to Ed Wilts for raising.

- :ear_of_rice: Harvest includes new templates to collect:
  - IP routes. Thanks jfong for contributing!
  - QoS fixed and adaptive policy groups. Thanks @faguayot for raising!
  - Cloud targets and storage
  - Export rules
  - Namespaces
  - CIFS clients
  - LIF counters
  - Volume efficiency stats

- Harvest containers are published to [GitHub's container registry](https://github.com/NetApp/harvest/pkgs/container/harvest) in addition to DockerHub and cr.netapp.io.
  If you're using `cr.netapp.io`, we encourage you to switch to ghcr.io or DockerHub. In 2024, we will stop publishing to `cr.netapp.io`

- Harvest uses a distroless image as its base now - reducing the size of the container and reducing the attack surface

- Harvest collects 38 additional EMS events and alert rules in this release

- Harvest EMS alert rules were updated to include better label names and align their severity with [Prometheus best practices](https://monitoring.mixins.dev/#guidelines-for-alert-names-labels-and-annotations). Thanks to @7840vz for contributing this feature!

- The `bin/harvest doctor` tool validates your `custom.yaml` template files, checking them for errors.

- :closed_book: Documentation additions
  - How to set up [Harvest with Kubernetes](https://github.com/NetApp/harvest/tree/main/container/k8)
  - Harvest [metadata metrics](https://netapp.github.io/harvest/latest/monitor-harvest/)
  - [Authenticating with ONTAP and StorageGRID](https://netapp.github.io/harvest/23.05/configure-harvest-basic/#authentication) clusters and auth precedence
  - [Pollers](https://netapp.github.io/harvest/latest/configure-harvest-basic/#pollers) include a `prefer_zapi` flag that tells Harvest to use the ZAPI API if the cluster supports it, otherwise allow Harvest to choose REST or ZAPI, whichever is appropriate to the ONTAP version. See [rest-strategy](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-strategy.md) for details.

- :tophat: Harvest makes it easy to run with both the ZAPI and REST collectors at the same time. Overlapping resources are deduplicated and only published to Prometheus once. This was the final piece in our journey to REST. See [rest-strategy.md](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-strategy.md) if you are interested in the details.

## Announcements

**IMPORTANT** The `volume_aggr_labels` metric is being deprecated in the `23.05` release and will be removed in the `23.08` release of Harvest ([#1966](https://github.com/NetApp/harvest/pull/1966)) `volume_aggr_labels` is redundant and the same labels are already available via `volume_labels`.

**IMPORTANT** To reduce image and download size, several tools were combined in `23.05`. The following binaries are no longer included: `bin/grafana`, `bin/rest`, `bin/zapi`. Use `bin/harvest grafana`, `bin/harvest rest`, and `bin/harvest zapi` instead.

:bangbang: **IMPORTANT** NetApp moved their communities from Slack to [Discord](https://discord.gg/ZmmWPHTBHw), please join us [there](https://discordapp.com/channels/855068651522490400/1001963189124206732)!

:bangbang: **IMPORTANT** If using Docker Compose and you want to keep your historical Prometheus data, please
read [how to migrate your Prometheus volume](https://github.com/NetApp/harvest/blob/main/docs/MigratePrometheusDocker.md)

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the `bin/harvest grafana import` CLI, from the Grafana UI, or from the `Maintenance > Reset Harvest Dashboards` button in NAbox.

## Known Issues

- Harvest does not calculate power metrics for AFF A250 systems. This data is not available from ONTAP via ZAPI or REST.
  See ONTAP bug [1511476](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1511476) for more details.

- ONTAP does not include REST metrics for `offbox_vscan_server` and `offbox_vscan` until ONTAP 9.13.1. See ONTAP bug
  [1473892](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1473892) for more details.

**IMPORTANT** 7-mode filers that are not on the latest release of ONTAP may experience TLS connection issues with errors like `tls: server selected unsupported protocol version 301` This is caused by a change in Go 1.18. The [default for TLS client connections was changed to TLS 1.2](https://tip.golang.org/doc/go1.18#tls10) in Go 1.18. Please upgrade your 7-mode filers (recommended) or set `tls_min_version: tls10` in your `harvest.yml` [poller section](https://github.com/NetApp/harvest/tree/release/22.05.0#pollers). See [#1007](https://github.com/NetApp/harvest/issues/1007) for more details.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards
this release:

@7840vz, @DAx-cGn, @Falcon667, @Hedius, @LukaszWasko, @MrObvious, @ReneMeier, @Sawall10, @T1r0l, @XDavidT, @aticatac, @chadpruden, @cygio, @ddhti, @debert-ntap, @demalik, @electrocreative, @elsgaard, @ev1963, @faguayot, @iStep2Step, @jgasher, @jmg011, @mamoep, @matejzero, @matthieu-sudo, @merdos, @pilot7777, @rodenj1, Alessandro.Nuzzo, Ed Wilts, Imthenightbird, KlausHub, MeghanaD, Paul P2, Rusty Brown, Shubham Mer, Tudor Pascu, Watson9121, jf38800, jfong, rcl23, troysmuller, twodot0h

:seedling: This release includes 61 features, 49 bug fixes, 22 documentation, 2 testing, 8 refactoring, 25 miscellaneous, and 32 ci pull requests.

### :rocket: Features
- Pollers Should Allow Customers To Opt Out Of Rest Upgrade ([#1744](https://github.com/NetApp/harvest/pull/1744))
- Restperf Vscan Counters ([#1751](https://github.com/NetApp/harvest/pull/1751))
- Smb2 Dashboard ([#1754](https://github.com/NetApp/harvest/pull/1754))
- Add Object Count To S3 Metrics ([#1759](https://github.com/NetApp/harvest/pull/1759))
- Enable Golanglint "Unparam" Linter ([#1769](https://github.com/NetApp/harvest/pull/1769))
- Dependabot Should Bump Dependencies ([#1777](https://github.com/NetApp/harvest/pull/1777))
- Print Missing Rest Metrics In Metric Generate Command ([#1783](https://github.com/NetApp/harvest/pull/1783))
- Add Datacenter To Metadata Exporter_time Metrics ([#1789](https://github.com/NetApp/harvest/pull/1789))
- Percentage Panels Should Clamp Min/Max To 0/100% ([#1790](https://github.com/NetApp/harvest/pull/1790))
- Qtree Dashboard Should Include Topk Qtrees By Disk Used Growth ([#1792](https://github.com/NetApp/harvest/pull/1792))
- Harvest Should Collect Ip Routes ([#1801](https://github.com/NetApp/harvest/pull/1801))
- Include Aggregate Encryption Information In Rest/Zapi Templates ([#1803](https://github.com/NetApp/harvest/pull/1803))
- Add Encrypted Field To Aggregate Dashboard ([#1804](https://github.com/NetApp/harvest/pull/1804))
- Harvest Should Include Sg Traffic Classification Panels ([#1807](https://github.com/NetApp/harvest/pull/1807))
- Harvest Should Fetch Auth Via Script ([#1819](https://github.com/NetApp/harvest/pull/1819))
- Delay Center Dashboard ([#1824](https://github.com/NetApp/harvest/pull/1824))
- Publish Harvest Images To Github Container Registry ([#1827](https://github.com/NetApp/harvest/pull/1827))
- Harvest Should Default To Pulling Images From Github Container … ([#1830](https://github.com/NetApp/harvest/pull/1830))
- Harvest Should Collect Qos Policy Groups ([#1831](https://github.com/NetApp/harvest/pull/1831))
- Ontap S3 Dashboard - Config Metrics ([#1833](https://github.com/NetApp/harvest/pull/1833))
- Harvest Should Collect Cloud Targets ([#1836](https://github.com/NetApp/harvest/pull/1836))
- Add Routes To Network Dashboard ([#1840](https://github.com/NetApp/harvest/pull/1840))
- Harvest Should Collect Export Rules ([#1843](https://github.com/NetApp/harvest/pull/1843))
- Workload Dashboard ([#1846](https://github.com/NetApp/harvest/pull/1846))
- Harvest Should Collect Adaptive Qos Policy Groups ([#1847](https://github.com/NetApp/harvest/pull/1847))
- Harvest Should Turn Dashboard Refresh Off ([#1849](https://github.com/NetApp/harvest/pull/1849))
- Namespace Dashboard ([#1850](https://github.com/NetApp/harvest/pull/1850))
- Create Release Issue Template ([#1856](https://github.com/NetApp/harvest/pull/1856))
- Enable Rest Ci Failures ([#1858](https://github.com/NetApp/harvest/pull/1858))
- Bin/Rest Should Be Able To Query All Clusters ([#1866](https://github.com/NetApp/harvest/pull/1866))
- Go Test Should Detect Races And Order Dependent Tests ([#1868](https://github.com/NetApp/harvest/pull/1868))
- Add Average Cpu Utilization And Cpu Busy In Harvest Dashboards ([#1872](https://github.com/NetApp/harvest/pull/1872))
- Harvest Should Use A Distroless Image As Its Base Image Instead… ([#1877](https://github.com/NetApp/harvest/pull/1877))
- Cluster Health Dashboard ([#1881](https://github.com/NetApp/harvest/pull/1881))
- Harvest Should Define And Document Auth Precedence ([#1882](https://github.com/NetApp/harvest/pull/1882))
- Aggregate Template Should Collect Cloud_storage ([#1883](https://github.com/NetApp/harvest/pull/1883))
- Harvest Should Include Template Unit Tests ([#1887](https://github.com/NetApp/harvest/pull/1887))
- Move Docker Folder To Container ([#1898](https://github.com/NetApp/harvest/pull/1898))
- Enable Smb2 Template ([#1923](https://github.com/NetApp/harvest/pull/1923))
- Harvest Generate Should Include A `--Volume` Option For Additio… ([#1924](https://github.com/NetApp/harvest/pull/1924))
- Harvest Should Collect Cifs Clients ([#1935](https://github.com/NetApp/harvest/pull/1935))
- Collect External_service_op Perf Object ([#1941](https://github.com/NetApp/harvest/pull/1941))
- Ci Regression Runs Locally ([#1943](https://github.com/NetApp/harvest/pull/1943))
- Harvest Should Include A Sg And Ontap Fabricpool Dashboard ([#1945](https://github.com/NetApp/harvest/pull/1945))
- Collect Lif Counters ([#1956](https://github.com/NetApp/harvest/pull/1956))
- Collect Volume Sis Stat ([#1958](https://github.com/NetApp/harvest/pull/1958))
- Update Alerts Summary ([#1967](https://github.com/NetApp/harvest/pull/1967))
- Map Ems Severity To Prom Sev ([#1973](https://github.com/NetApp/harvest/pull/1973))
- Grafana Should Retry On Err Or Status=500 ([#1974](https://github.com/NetApp/harvest/pull/1974))
- Doctor Should Validate Custom.yaml Files ([#1979](https://github.com/NetApp/harvest/pull/1979))
- Update Workload Panel Titles ([#1980](https://github.com/NetApp/harvest/pull/1980))
- Add Cifs Connection To Smb Dashboard ([#1982](https://github.com/NetApp/harvest/pull/1982))
- Add Volume Stat Panels To Volume Dashboard ([#1985](https://github.com/NetApp/harvest/pull/1985))
- Topk Variables In Dashboards Should Change With Time Range Change ([#1987](https://github.com/NetApp/harvest/pull/1987))
- Collect 38 More Ems Events ([#1988](https://github.com/NetApp/harvest/pull/1988))
- Include Ems Alerts For All Ems Events ([#1992](https://github.com/NetApp/harvest/pull/1992))
- 23.05 Metrics Docs ([#2002](https://github.com/NetApp/harvest/pull/2002))
- Update Docker Prometheus Variables ([#2003](https://github.com/NetApp/harvest/pull/2003))
- Add Missing Rest Counters For Svm_nfs V3 ([#2007](https://github.com/NetApp/harvest/pull/2007))
- Add Names To Harvest Docker Networks ([#2017](https://github.com/NetApp/harvest/pull/2017))
- Add Column Filter For Buckets In Tenant Dashboard ([#2020](https://github.com/NetApp/harvest/pull/2020))

### :bug: Bug Fixes
- Handle Min-Max For Network Dashboard ([#1763](https://github.com/NetApp/harvest/pull/1763))
- Omit Changelog Categories That Are Empty ([#1776](https://github.com/NetApp/harvest/pull/1776))
- Aggregating Latency Metrics Returns Nan When Base Counter Is 0 ([#1781](https://github.com/NetApp/harvest/pull/1781))
- Backward Compatibility For Qtree Metrics In Rest ([#1788](https://github.com/NetApp/harvest/pull/1788))
- Rename Metadata Row So Rest And Zapi Are Included ([#1791](https://github.com/NetApp/harvest/pull/1791))
- Fetch Few Counters From Ontap Instead Of Um Api ([#1793](https://github.com/NetApp/harvest/pull/1793))
- Rest Fabricpool Metric Label Should Match Zapi ([#1794](https://github.com/NetApp/harvest/pull/1794))
- Handle Array As Comma Separated Value In Zapi ([#1810](https://github.com/NetApp/harvest/pull/1810))
- Increase Lag Time Log Print ([#1832](https://github.com/NetApp/harvest/pull/1832))
- User Read|Write Panels Should Use Power Of Two Bytes ([#1834](https://github.com/NetApp/harvest/pull/1834))
- Session Setup Latency Heatmap Panel Is Duplicate On The Dashboard ([#1839](https://github.com/NetApp/harvest/pull/1839))
- Reduce Dns Storm By Disabling Netconnections ([#1845](https://github.com/NetApp/harvest/pull/1845))
- Prometheus Alert For `Node_nfs_latency` Is Microsecs ([#1853](https://github.com/NetApp/harvest/pull/1853))
- Prometheus Alert For Node_nfs_latency Is Microsecs ([#1854](https://github.com/NetApp/harvest/pull/1854))
- Fixing Security Account Plugin Generated Metrics ([#1857](https://github.com/NetApp/harvest/pull/1857))
- Explain How To Join Discord Before Harvest Channel ([#1860](https://github.com/NetApp/harvest/pull/1860))
- Latency Unit Fix In Namespace Dashboard ([#1862](https://github.com/NetApp/harvest/pull/1862))
- Bin/Rest Should Log Errors ([#1876](https://github.com/NetApp/harvest/pull/1876))
- Fix Template Object Name ([#1878](https://github.com/NetApp/harvest/pull/1878))
- Change Color Scheme For Heatmaps ([#1880](https://github.com/NetApp/harvest/pull/1880))
- Adding Available Column In Aggr Dashboard ([#1896](https://github.com/NetApp/harvest/pull/1896))
- Correct Object Name In Ontaps3_svm.yaml ([#1900](https://github.com/NetApp/harvest/pull/1900))
- Correct Ci Logs ([#1903](https://github.com/NetApp/harvest/pull/1903))
- Use Certificate Auth When Auth_style Is Certificate_auth ([#1904](https://github.com/NetApp/harvest/pull/1904))
- Log Time Drift Between Nodes In Ems Collector ([#1908](https://github.com/NetApp/harvest/pull/1908))
- Docker Fix ([#1911](https://github.com/NetApp/harvest/pull/1911))
- Handle Interface Api Call In Svm Rest/Zapi ([#1912](https://github.com/NetApp/harvest/pull/1912))
- Restrict Exemplar Flag In Dashboards ([#1925](https://github.com/NetApp/harvest/pull/1925))
- Update Harvest.cue To Match Config ([#1926](https://github.com/NetApp/harvest/pull/1926))
- Handle Custom File Of Status_7mode Object ([#1927](https://github.com/NetApp/harvest/pull/1927))
- Sorted Exported Keys And Labels Test ([#1928](https://github.com/NetApp/harvest/pull/1928))
- Storagegrid Overview Panel Is A Missing Query ([#1930](https://github.com/NetApp/harvest/pull/1930))
- Harvest Should Check For Free Promport When Restarting ([#1931](https://github.com/NetApp/harvest/pull/1931))
- Timeseries Panels With Bytes Should Set Decimals=2 ([#1934](https://github.com/NetApp/harvest/pull/1934))
- Show Only The Data Svms In Svm Compliance Dashboard ([#1939](https://github.com/NetApp/harvest/pull/1939))
- Typo In Health Dashboard ([#1948](https://github.com/NetApp/harvest/pull/1948))
- Log Noise In Rest Collector ([#1957](https://github.com/NetApp/harvest/pull/1957))
- Endpoint Key Order Fix ([#1960](https://github.com/NetApp/harvest/pull/1960))
- Deprecate Volume_aggr_labels Metric ([#1966](https://github.com/NetApp/harvest/pull/1966))
- Reduce Shelf Log Noise In Restperf Collector ([#1969](https://github.com/NetApp/harvest/pull/1969))
- Changed Warn To Error In Ems For Key/Labels ([#1981](https://github.com/NetApp/harvest/pull/1981))
- Combine Netport And Port Templates ([#1983](https://github.com/NetApp/harvest/pull/1983))
- Correct Some Mistakes In Ems.yaml ([#1984](https://github.com/NetApp/harvest/pull/1984))
- Adding Nfs Versions In Queries Where Its Missed ([#1998](https://github.com/NetApp/harvest/pull/1998))
- Add Logs For Ems Error ([#2019](https://github.com/NetApp/harvest/pull/2019))
- Restore Bin/Grafana For Nabox ([#2022](https://github.com/NetApp/harvest/pull/2022))
- Remove Cifs Clients Template ([#2024](https://github.com/NetApp/harvest/pull/2024))
- Fixing Key Order In Qtree Plugin ([#2028](https://github.com/NetApp/harvest/pull/2028))
- 7Mode Qtree Key Order In Plugin ([#2033](https://github.com/NetApp/harvest/pull/2033))

### :closed_book: Documentation
- Clarify Envvar And Overlapping Collectors ([#1709](https://github.com/NetApp/harvest/pull/1709))
- Highlight That Harvest Requires Go ([#1761](https://github.com/NetApp/harvest/pull/1761))
- Add Fsa Ontap Enable Instructions In Dashboard ([#1762](https://github.com/NetApp/harvest/pull/1762))
- Document Rest Perf Metrics Implementation Details ([#1785](https://github.com/NetApp/harvest/pull/1785))
- Fsa Dashboard Should Highlight Ontap Actions ([#1787](https://github.com/NetApp/harvest/pull/1787))
- Add Information About Enabling Template For Nfsv4 Storepool Mon… ([#1797](https://github.com/NetApp/harvest/pull/1797))
- Mention Prefer_zapi In Rest Strategy Docs ([#1799](https://github.com/NetApp/harvest/pull/1799))
- Harvest Should Fetch Auth Via Script ([#1822](https://github.com/NetApp/harvest/pull/1822))
- Update Fsa Dashboard Information ([#1851](https://github.com/NetApp/harvest/pull/1851))
- Add Permissions To Docs For Qos ([#1869](https://github.com/NetApp/harvest/pull/1869))
- Include A Link To Nabox Troubleshooting ([#1891](https://github.com/NetApp/harvest/pull/1891))
- Fixing Numbers And Use `--Port` By Default ([#1917](https://github.com/NetApp/harvest/pull/1917))
- K8 Docs ([#1932](https://github.com/NetApp/harvest/pull/1932))
- Fix Dead Link ([#1950](https://github.com/NetApp/harvest/pull/1950))
- Document Metadata Metrics Harvest Publishes ([#1951](https://github.com/NetApp/harvest/pull/1951))
- Clarify That Source And Dest Clusters Need To Export To Same Pro… ([#1995](https://github.com/NetApp/harvest/pull/1995))
- Explain How To Upgrade Docker Compose To Nightly ([#2001](https://github.com/NetApp/harvest/pull/2001))
- Highlight Ontap Rest Performance Counters ([#2009](https://github.com/NetApp/harvest/pull/2009))
- Add Workload Description ([#2011](https://github.com/NetApp/harvest/pull/2011))
- Add Unit To Workload And Metadata Counters ([#2013](https://github.com/NetApp/harvest/pull/2013))
- Exclude Bucket Histogram From Docs ([#2018](https://github.com/NetApp/harvest/pull/2018))
- Fix Grafana Spelling ([#2029](https://github.com/NetApp/harvest/pull/2029))

### :wrench: Testing
- Ensure All Dashboard Heatmaps Use The Same Colorscheme And Style ([#1884](https://github.com/NetApp/harvest/pull/1884))
- Remove Global `Validateportinuse` That Caused Test To Fail ([#1889](https://github.com/NetApp/harvest/pull/1889))

### Refactoring
- Plugins Should Accept Map Of Matrix, Like Collector ([#1798](https://github.com/NetApp/harvest/pull/1798))
- Rename Ontaps3 Perf Metrics To Ontaps3_svm ([#1899](https://github.com/NetApp/harvest/pull/1899))
- Reduce Log Noise When Ontap Apis Do Not Exist ([#1901](https://github.com/NetApp/harvest/pull/1901))
- Reduce Log Noise Disk ([#1902](https://github.com/NetApp/harvest/pull/1902))
- Remove Unnecessary Dependency ([#1936](https://github.com/NetApp/harvest/pull/1936))
- Generate Should Not Panic ([#1938](https://github.com/NetApp/harvest/pull/1938))
- Reduce Ems Logs ([#2012](https://github.com/NetApp/harvest/pull/2012))
- Move Vscan From 9.12 To 9.13 ([#2015](https://github.com/NetApp/harvest/pull/2015))

### Miscellaneous
- Bump Lumberjack ([#1713](https://github.com/NetApp/harvest/pull/1713))
- Update Integration Go Dependencies ([#1746](https://github.com/NetApp/harvest/pull/1746))
- Merge 23.02 To Main ([#1758](https://github.com/NetApp/harvest/pull/1758))
- Bump Dependencies ([#1767](https://github.com/NetApp/harvest/pull/1767))
- Bump Golang.org/X/Text From 0.7.0 To 0.8.0 In /Integration ([#1811](https://github.com/NetApp/harvest/pull/1811))
- Bump Github.com/Stretchr/Testify From 1.8.1 To 1.8.2 In /Integration ([#1812](https://github.com/NetApp/harvest/pull/1812))
- Bump Github.com/Shirou/Gopsutil/V3 From 3.23.1 To 3.23.2 ([#1813](https://github.com/NetApp/harvest/pull/1813))
- Bump Golang.org/X/Text From 0.7.0 To 0.8.0 ([#1814](https://github.com/NetApp/harvest/pull/1814))
- Bump Golang.org/X/Term From 0.5.0 To 0.6.0 ([#1815](https://github.com/NetApp/harvest/pull/1815))
- Bump Golang.org/X/Sys From 0.5.0 To 0.6.0 ([#1816](https://github.com/NetApp/harvest/pull/1816))
- Bump Github.com/Imdario/Mergo From 0.3.13 To 0.3.14 ([#1837](https://github.com/NetApp/harvest/pull/1837))
- Add Link To Release Page ([#1859](https://github.com/NetApp/harvest/pull/1859))
- Bump Github.com/Imdario/Mergo From 0.3.14 To 0.3.15 ([#1870](https://github.com/NetApp/harvest/pull/1870))
- Bump Github.com/Zekrotja/Timedmap From 1.4.0 To 1.5.1 ([#1871](https://github.com/NetApp/harvest/pull/1871))
- Bump Github.com/Docker/Docker From 23.0.1+Incompatible To 23.0.2+Incompatible In /Integration ([#1886](https://github.com/NetApp/harvest/pull/1886))
- Bump Github.com/Shirou/Gopsutil/V3 From 3.23.2 To 3.23.3 ([#1888](https://github.com/NetApp/harvest/pull/1888))
- Fix Integration Security Vulnerabilities ([#1894](https://github.com/NetApp/harvest/pull/1894))
- Update Golang To 1.20.3 ([#1905](https://github.com/NetApp/harvest/pull/1905))
- Print Number Of Object And Counters Harvest Collects ([#1916](https://github.com/NetApp/harvest/pull/1916))
- Bump Golang.org/X/Sys From 0.6.0 To 0.7.0 ([#1919](https://github.com/NetApp/harvest/pull/1919))
- Bump Github.com/Spf13/Cobra From 1.6.1 To 1.7.0 ([#1920](https://github.com/NetApp/harvest/pull/1920))
- Bump Golang.org/X/Term From 0.6.0 To 0.7.0 ([#1921](https://github.com/NetApp/harvest/pull/1921))
- Bump Golang.org/X/Text From 0.8.0 To 0.9.0 ([#1922](https://github.com/NetApp/harvest/pull/1922))
- Bump Github.com/Rs/Zerolog From 1.29.0 To 1.29.1 ([#1952](https://github.com/NetApp/harvest/pull/1952))
- Bump Github.com/Go-Openapi/Spec From 0.20.8 To 0.20.9 ([#1989](https://github.com/NetApp/harvest/pull/1989))

### :hammer: CI
- Bump Go ([#1764](https://github.com/NetApp/harvest/pull/1764))
- Update Clabot ([#1765](https://github.com/NetApp/harvest/pull/1765))
- Add Govulncheck To Workflows ([#1778](https://github.com/NetApp/harvest/pull/1778))
- Bump Go To 1.20.2 ([#1806](https://github.com/NetApp/harvest/pull/1806))
- Pull Go Version Into Ci Var ([#1808](https://github.com/NetApp/harvest/pull/1808))
- Bump Github Actions To Address Eol And Warnings ([#1809](https://github.com/NetApp/harvest/pull/1809))
- Let Dependabot[Bot] Merge Prs ([#1817](https://github.com/NetApp/harvest/pull/1817))
- Let Dependabot[Bot] Merge Prs ([#1818](https://github.com/NetApp/harvest/pull/1818))
- Prune Untagged Images ([#1885](https://github.com/NetApp/harvest/pull/1885))
- Prune Untagged Images ([#1890](https://github.com/NetApp/harvest/pull/1890))
- Make Lint Pr Match Commitlint ([#1892](https://github.com/NetApp/harvest/pull/1892))
- Prune Untagged Images ([#1893](https://github.com/NetApp/harvest/pull/1893))
- Use Go.dev/Dl To Download Artifacts ([#1906](https://github.com/NetApp/harvest/pull/1906))
- Changed Wait Time In Ems Tests From 3M To 3M15sec ([#1907](https://github.com/NetApp/harvest/pull/1907))
- Use Certificate Auth On Native ([#1909](https://github.com/NetApp/harvest/pull/1909))
- Don't Fail When Fetch-Asup Is Missing ([#1918](https://github.com/NetApp/harvest/pull/1918))
- Switch Cert To Harvest_cert.yml ([#1937](https://github.com/NetApp/harvest/pull/1937))
- Rpm Does Not Need Certificates ([#1940](https://github.com/NetApp/harvest/pull/1940))
- Add Asup Validation Check ([#1942](https://github.com/NetApp/harvest/pull/1942))
- Should Fail When Logs Contain Errors ([#1944](https://github.com/NetApp/harvest/pull/1944))
- Check Namespace Counters ([#1947](https://github.com/NetApp/harvest/pull/1947))
- Remove Docker Dependency ([#1949](https://github.com/NetApp/harvest/pull/1949))
- Simplify Counter Validation ([#1955](https://github.com/NetApp/harvest/pull/1955))
- Simplify Counter Validation ([#1962](https://github.com/NetApp/harvest/pull/1962))
- Grafana Db Locked ([#1964](https://github.com/NetApp/harvest/pull/1964))
- Test 1 Non-Bookend Ems In Ci ([#1971](https://github.com/NetApp/harvest/pull/1971))
- Improve Ems Alert Logging And Event Generation ([#1993](https://github.com/NetApp/harvest/pull/1993))
- Improve Ems Alert Logging And Event Generation ([#1999](https://github.com/NetApp/harvest/pull/1999))
- Import With Overwrite ([#2000](https://github.com/NetApp/harvest/pull/2000))
- Handle Sm.mediator.misconfigured Ems ([#2006](https://github.com/NetApp/harvest/pull/2006))
- Mediator Related Ems Changes In Ci ([#2010](https://github.com/NetApp/harvest/pull/2010))
- Fix Ci Ems Test Case ([#2016](https://github.com/NetApp/harvest/pull/2016))

---

## 23.02.0 / 2023-02-21

:pushpin: Highlights of this major release include:

- :sparkles: Harvest includes a new file system analytics (FSA) dashboard with directory growth, top directories per volume, and volume usage statistics.

- Harvest includes a new StorageGRID overview dashboard with performance, storage, information lifecycle management, and node panels. We're collecting suggestions on which StorageGRID dashboards you'd like to see next in issue [#1420](https://github.com/NetApp/harvest/issues/1420).

- :bulb: Power dashboard includes new panels for total power by aggregate disk type, average power per used TB, average IOPs/Watt, total power by aggregate, and information on sensor problems.

- :tophat: Harvest makes it easy to run with both the ZAPI and REST collectors at the same time. Overlapping resources are deduplicated and only published to Prometheus once. This was the final piece in our journey to REST. See [rest-strategy.md](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-strategy.md) if you are interested in the details.

- :closed_book: We made lots of improvements to Harvest's [new documentation site](https://netapp.github.io/harvest/) this release including one of the most requested features - a list of Harvest metrics and their corresponding ONTAP ZAPI/REST API mappings. :triangular_ruler: [Check it out](https://netapp.github.io/harvest/latest/ontap-metrics/)

- :gem: New dashboards and improvements
  - A new file system analytics (FSA) dashboard with directory growth, top directories per volume, and volume usage statistics
  - A new StorageGRID overview dashboard with performance, storage, information lifecycle management, and node panels
  - Power dashboard includes new panels for total power by aggregate disk type, average power per used TB, average IOPs/Watt, total power by aggregate, and information on sensor problems.
  - Disk dashboard shows which node/controller a disk belongs too
  - SVM dashboard shows topK resources in panel drill downs
  - SnapMirror dashboard includes transfer duration, lag time and transfer data panels in addition to new source and destination volume variables to make it easier to understand SnapMirror relationships
  - Aggregate dashboard includes a new flash pool drill down with five new panels
  - Aggregate dashboard includes four new panels showing volume statistics broken down by flexvol/flexgroup space per aggregate
  - SVM dashboard includes NFSv3 latency heatmap panels
  - Node dashboard latency panels updated to use weighted average, bringing them inline with ActiveIQ
  - Volume dashboard includes new inode usage panels

- Harvest includes a new command `bin/harvest grafana metrics` which shows which metrics each dashboard uses

## Announcements

:bangbang: **IMPORTANT** NetApp moved their communities from Slack to [Discord](https://discord.gg/ZmmWPHTBHw), please
join us [there](https://discordapp.com/channels/855068651522490400/1001963189124206732)!

:bangbang: **IMPORTANT** If using Docker Compose and you want to keep your historical Prometheus data, please
read [how to migrate your Prometheus volume](https://github.com/NetApp/harvest/blob/main/docs/MigratePrometheusDocker.md)

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and
fixes. You can import them via the `bin/harvest/grafana import` CLI, from the Grafana UI, or from
the `Maintenance > Reset Harvest Dashboards` button in NAbox.

:sunflower: In the `22.11.0` release notes, we announced that we would be removing quota metrics prefixed with qtree.
Several of you asked us to leave them. :+1: We will continue publishing them as-is.

## Known Issues

- Harvest does not calculate power metrics for AFF A250 systems. This data is not available from ONTAP via ZAPI or REST.
  See ONTAP bug [1511476](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1511476) for more details.

- ONTAP does not include REST metrics for `offbox_vscan_server` and `offbox_vscan` until ONTAP 9.13.1. See ONTAP bug
  [1473892](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1473892) for more details.

**IMPORTANT** 7-mode filers that are not on the latest release of ONTAP may experience TLS connection issues with errors
like `tls: server selected unsupported protocol version 301` This is caused by a change in Go 1.18.
The [default for TLS client connections was changed to TLS 1.2](https://tip.golang.org/doc/go1.18#tls10) in Go 1.18.
Please upgrade your 7-mode filers (recommended) or set `tls_min_version: tls10` in
your `harvest.yml` [poller section](https://github.com/NetApp/harvest/tree/release/22.05.0#pollers).
See [#1007](https://github.com/NetApp/harvest/issues/1007) for more details.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards
this release:

@Falcon667, @MrObvious, @ReneMeier, @Sawall10, @T1r0l, @aticatac, @chadpruden, @demalik, @electrocreative, @ev1963, @faguayot, @iStep2Step, @jgasher, @jmg011, @mamoep, @matejzero, @matthieu-sudo, @merdos, @rodenj1, Ed Wilts, KlausHub, MeghanaD, Paul P2, Rusty Brown, Shubham Mer, Tudor Pascu, jf38800, jfong, rcl23, troysmuller, twodot0h

:seedling: This release includes 43 features, 43 bug fixes, 19 documentation, 2 testing, 1 styling, 5 miscellaneous, and 7 ci pull requests.

### :rocket: Features
- Add Information To Which Node/Controller A Disk Belongs ([#1542](https://github.com/NetApp/harvest/pull/1542))
- Remove Pass Slice From Matrix Data Structure ([#1553](https://github.com/NetApp/harvest/pull/1553))
- Ensure Dashboards Have Only One Expanded Section ([#1554](https://github.com/NetApp/harvest/pull/1554))
- Plugins Can Use Raw Or Display Metric In Calculations ([#1567](https://github.com/NetApp/harvest/pull/1567))
- Perf Collector Delta Calculation Handling ([#1571](https://github.com/NetApp/harvest/pull/1571))
- Added Dashboard Tests For Legends Details ([#1576](https://github.com/NetApp/harvest/pull/1576))
- Add `Bin/Grafana Metrics` To Print Which Metrics Each Dashboard… ([#1578](https://github.com/NetApp/harvest/pull/1578))
- Restperf Svm_vscan Template ([#1590](https://github.com/NetApp/harvest/pull/1590))
- Simplify Metrics Storage ([#1591](https://github.com/NetApp/harvest/pull/1591))
- Include Inodes File Usage In Volume Dashboard ([#1593](https://github.com/NetApp/harvest/pull/1593))
- Handle Record Values In Metric Calculation ([#1594](https://github.com/NetApp/harvest/pull/1594))
- Refractor Matrix ([#1595](https://github.com/NetApp/harvest/pull/1595))
- Topk Support In Svm Dashboard - 1 ([#1608](https://github.com/NetApp/harvest/pull/1608))
- Topk Support In Svm Dashboard - 2 ([#1609](https://github.com/NetApp/harvest/pull/1609))
- Topk Support In Svm Dashboard - 3 ([#1611](https://github.com/NetApp/harvest/pull/1611))
- Add Failed Sensors To Power Dashboard ([#1621](https://github.com/NetApp/harvest/pull/1621))
- Include Storagegrid Dashboard In Docker Compose ([#1631](https://github.com/NetApp/harvest/pull/1631))
- Include Storagegrid Dashboard In Docker Compose ([#1632](https://github.com/NetApp/harvest/pull/1632))
- Honor Harvest_no_upgrade Envvar When Zapi Exist ([#1636](https://github.com/NetApp/harvest/pull/1636))
- Harvest Metrics Document ([#1641](https://github.com/NetApp/harvest/pull/1641))
- Added 3 Panels Supported With Relationshipid Data Link ([#1642](https://github.com/NetApp/harvest/pull/1642))
- Adding Flashpool Drilldown Panels In Aggr Dashboard ([#1649](https://github.com/NetApp/harvest/pull/1649))
- List Docker Tags On Cr.netapp.io ([#1656](https://github.com/NetApp/harvest/pull/1656))
- Fsa ([#1661](https://github.com/NetApp/harvest/pull/1661))
- Move Shelf Power To Disk Perf Template ([#1665](https://github.com/NetApp/harvest/pull/1665))
- Aggregate Power For Zapi Collector ([#1671](https://github.com/NetApp/harvest/pull/1671))
- Weighted Avg Support In Aggregator Plugin ([#1672](https://github.com/NetApp/harvest/pull/1672))
- Add Activity To Panel Names ([#1675](https://github.com/NetApp/harvest/pull/1675))
- Add Storagegrid Overview Dashboard ([#1677](https://github.com/NetApp/harvest/pull/1677))
- Calculate Power By Disk Type ([#1681](https://github.com/NetApp/harvest/pull/1681))
- Support Aggr Filter For Flexgroup Volumes ([#1691](https://github.com/NetApp/harvest/pull/1691))
- Calculate Aggr Power Rest Support ([#1692](https://github.com/NetApp/harvest/pull/1692))
- Support Aggr Filter Chnages In Zapiperf ([#1695](https://github.com/NetApp/harvest/pull/1695))
- Calculate Power Per Tb And Watt ([#1698](https://github.com/NetApp/harvest/pull/1698))
- Add Nfsv3 Latency Heatmap ([#1699](https://github.com/NetApp/harvest/pull/1699))
- Add Fsa Full Form To Dashboard Name ([#1700](https://github.com/NetApp/harvest/pull/1700))
- Add Ca Certificate Support For Rest Client ([#1705](https://github.com/NetApp/harvest/pull/1705))
- Support Node Aggregation For Flexgroup Also ([#1706](https://github.com/NetApp/harvest/pull/1706))
- Cluster Dashboard Panel Width Ux Changes ([#1723](https://github.com/NetApp/harvest/pull/1723))
- Include Sg Cluster In Panels ([#1725](https://github.com/NetApp/harvest/pull/1725))
- Shelf Power Panel Alignment Issue ([#1728](https://github.com/NetApp/harvest/pull/1728))
- Topk Panels Should Use Topresources Var In Their Titles ([#1733](https://github.com/NetApp/harvest/pull/1733))
- Pollers Should Allow Customers To Opt Out Of Rest Upgrade (#1744) ([#1747](https://github.com/NetApp/harvest/pull/1747))

### :bug: Bug Fixes
- Collapse All But Highlights In Svm Dashboard ([#1540](https://github.com/NetApp/harvest/pull/1540))
- Remove Datacenter From Snapmirror Queries ([#1546](https://github.com/NetApp/harvest/pull/1546))
- Handle Alignment In Svm Panel ([#1547](https://github.com/NetApp/harvest/pull/1547))
- Changed To Private Cli For Cifs_ntlm_enabled ([#1563](https://github.com/NetApp/harvest/pull/1563))
- Restperf Collector Causes Spikes Every Poll Counter ([#1564](https://github.com/NetApp/harvest/pull/1564))
- Import Should Not Empty Uid When --Overwrite Is Used ([#1566](https://github.com/NetApp/harvest/pull/1566))
- Remove Redundant Ems Log ([#1582](https://github.com/NetApp/harvest/pull/1582))
- Change Display Metrics Map Value To Metric Name ([#1583](https://github.com/NetApp/harvest/pull/1583))
- Node_nfs_ops Is Exposed Twice In Rest ([#1588](https://github.com/NetApp/harvest/pull/1588))
- Shelf Status Fix ([#1589](https://github.com/NetApp/harvest/pull/1589))
- Virus Scan Connections Panel Should Use Linear Scale ([#1592](https://github.com/NetApp/harvest/pull/1592))
- Histogram Issues ([#1598](https://github.com/NetApp/harvest/pull/1598))
- Suppress Missing 7Mode Template In Cdot ([#1607](https://github.com/NetApp/harvest/pull/1607))
- Don't Write Response On Error ([#1613](https://github.com/NetApp/harvest/pull/1613))
- Plugininvocationrate Fix For Plugins ([#1616](https://github.com/NetApp/harvest/pull/1616))
- Headroom Dashboard Description Fix ([#1618](https://github.com/NetApp/harvest/pull/1618))
- Cold Power Sensor Fix For Rest ([#1620](https://github.com/NetApp/harvest/pull/1620))
- Label Not Cleared When It Is Not Available In Subsequent Zapi/Rest Poll ([#1627](https://github.com/NetApp/harvest/pull/1627))
- Available Ops In Headroom Dashboard Should Be Displayed As Per Confidence Factor ([#1628](https://github.com/NetApp/harvest/pull/1628))
- Dashboard Test For Child Panels ([#1633](https://github.com/NetApp/harvest/pull/1633))
- Handling Sub-Panels While Importing With Prefix ([#1645](https://github.com/NetApp/harvest/pull/1645))
- Removing Status Metric In Security_cert As Nowhere Used ([#1651](https://github.com/NetApp/harvest/pull/1651))
- Exclude Node/Svm Vols And Include Data Svms ([#1658](https://github.com/NetApp/harvest/pull/1658))
- Unix Poller Error Unable To Read Process Cmdline ([#1674](https://github.com/NetApp/harvest/pull/1674))
- Handling Array Of Array In Rest Security Accounts ([#1676](https://github.com/NetApp/harvest/pull/1676))
- Backward Compatibility For Qtree Template ([#1679](https://github.com/NetApp/harvest/pull/1679))
- Metadata Rate Calculations Should Not Alias With Prom Scrape_int… ([#1682](https://github.com/NetApp/harvest/pull/1682))
- Variable Ds_prometheus Should Exist In Dashboard ([#1688](https://github.com/NetApp/harvest/pull/1688))
- Indent Node.yaml ([#1697](https://github.com/NetApp/harvest/pull/1697))
- Metadata_collector Metric Are Being Overwritten By Collectors ([#1708](https://github.com/NetApp/harvest/pull/1708))
- Port In Generate Docker Command Should Work For Standlone Harvest Container Deployment ([#1710](https://github.com/NetApp/harvest/pull/1710))
- Correct Name Description In Aggregate Panel ([#1715](https://github.com/NetApp/harvest/pull/1715))
- Panel Name Alias ([#1717](https://github.com/NetApp/harvest/pull/1717))
- Headroom Dashboard Available Ops Panel Is Broken ([#1720](https://github.com/NetApp/harvest/pull/1720))
- Topk In Table And Info Expanded As Default In Fsa Dashboard ([#1722](https://github.com/NetApp/harvest/pull/1722))
- Qtree Total_ops Panel Topk Mapping Is Wrong ([#1726](https://github.com/NetApp/harvest/pull/1726))
- Storagegrid Collector Should Be Included In Autosupport ([#1729](https://github.com/NetApp/harvest/pull/1729))
- Svm Dashboard Stat Panels Aggregation ([#1730](https://github.com/NetApp/harvest/pull/1730))
- Add Error Log When Collector Fails Due To Connection Issues ([#1731](https://github.com/NetApp/harvest/pull/1731))
- Topk Suppport In 3 Panels In Snapmirror Dashboard ([#1735](https://github.com/NetApp/harvest/pull/1735))
- Remove Power Panel From Cluster Dashboard ([#1736](https://github.com/NetApp/harvest/pull/1736))
- Storagegrid Cluster Name Should Be Grid Name Instead Of Admin No… ([#1737](https://github.com/NetApp/harvest/pull/1737))
- Changed Filter From Not Nil To Greater Than 0 ([#1741](https://github.com/NetApp/harvest/pull/1741))

### :closed_book: Documentation
- Add Storagegrid Collector And Prepare Docs ([#1532](https://github.com/NetApp/harvest/pull/1532))
- Add Rest/Restperf Collector Docs ([#1537](https://github.com/NetApp/harvest/pull/1537))
- Update Cdot Auth Docs ([#1548](https://github.com/NetApp/harvest/pull/1548))
- Add Cdot Auth Steps ([#1559](https://github.com/NetApp/harvest/pull/1559))
- Clarify What `--Overwrite` Does When Importing Dashboards That … ([#1574](https://github.com/NetApp/harvest/pull/1574))
- Clarify Docker Compose Upgrade ([#1604](https://github.com/NetApp/harvest/pull/1604))
- Change Env_var In Diagram To Match Code ([#1615](https://github.com/NetApp/harvest/pull/1615))
- Add Docs In Snapmirror Dashboard For Source/Destination Clusters ([#1619](https://github.com/NetApp/harvest/pull/1619))
- Update Cdot Auth Document ([#1624](https://github.com/NetApp/harvest/pull/1624))
- Connect Dashboards To Grafana Docs ([#1660](https://github.com/NetApp/harvest/pull/1660))
- Improve Harvest Metrics Display ([#1662](https://github.com/NetApp/harvest/pull/1662))
- Update Discord Links To New Forum ([#1667](https://github.com/NetApp/harvest/pull/1667))
- Link To Ontap Metrics ([#1669](https://github.com/NetApp/harvest/pull/1669))
- Update Download Instructions Link ([#1685](https://github.com/NetApp/harvest/pull/1685))
- Add `Metrics Query` Permissions For Storagegrid ([#1687](https://github.com/NetApp/harvest/pull/1687))
- Fix Yaml Formatting For Configure-Templates.md ([#1693](https://github.com/NetApp/harvest/pull/1693))
- Update Ontap Metrics And Mention How To Generate Grafana Metrics ([#1712](https://github.com/NetApp/harvest/pull/1712))
- Document Plugin Generated Metrics ([#1734](https://github.com/NetApp/harvest/pull/1734))
- Fix Links To Openssl Samples ([#1743](https://github.com/NetApp/harvest/pull/1743))

### :wrench: Testing
- Ensure Dashboards Use Spannull = True ([#1602](https://github.com/NetApp/harvest/pull/1602))
- Ensure Rate Calculations Are Not 1M ([#1683](https://github.com/NetApp/harvest/pull/1683))

### Styling
- Align Rename Markers ([#1599](https://github.com/NetApp/harvest/pull/1599))

### Refactoring

### Miscellaneous
- Merge 22.11 With Main ([#1493](https://github.com/NetApp/harvest/pull/1493))
- Merge 22.11.0 To Main ([#1528](https://github.com/NetApp/harvest/pull/1528))
- Update Third Party Dependencies ([#1543](https://github.com/NetApp/harvest/pull/1543))
- Integration Go Mod Tidy ([#1568](https://github.com/NetApp/harvest/pull/1568))
- Bump Dependencies ([#1652](https://github.com/NetApp/harvest/pull/1652))

### :hammer: CI
- Tag Closed Issues With `Status/Testme` ([#1541](https://github.com/NetApp/harvest/pull/1541))
- Enable Qos Monitoring For Rest Templates In Ci ([#1545](https://github.com/NetApp/harvest/pull/1545))
- Test Tag Closed Issues With `Status/Testme` ([#1552](https://github.com/NetApp/harvest/pull/1552))
- Unable To Make Issue Labeler Work. Removing ([#1596](https://github.com/NetApp/harvest/pull/1596))
- Ignore Node Scoped Ems In Ci ([#1601](https://github.com/NetApp/harvest/pull/1601))
- Nodescope Check Update For Bookend Ems ([#1634](https://github.com/NetApp/harvest/pull/1634))
- Node-Scoped Ems Correction ([#1638](https://github.com/NetApp/harvest/pull/1638))

---

## 22.11.0 / 2022-11-21
:pushpin: Highlights of this major release include:
- :sparkles: Harvest now includes a StorageGRID collector and a Tenant/Buckets dashboard. We're just getting started
  with StorageGRID dashboards. Please give the collector a try,
  and [let us know](https://github.com/NetApp/harvest/issues/1420) which StorageGRID dashboards you'd like to see next.

- :tophat: The REST collectors are ready! We recommend using them for ONTAP versions 9.12.1 and higher.
  Today, Harvest collects 1,546 metrics via ZAPI. Harvest includes a full set of REST templates that export identical
  metrics. All 1,546 metrics are available via Harvest's REST templates and no changes to dashboards or downstream
  metric-consumers is required. :tada:
  More details
  on [Harvest's REST strategy](https://github.com/NetApp/harvest/blob/main/docs/architecture/rest-strategy.md).

- :closed_book: Harvest has a [new documentation site](https://netapp.github.io/harvest/)! This consolidates Harvest
  documentation into one place and will make it easier to find what you need. Stay tuned for more updates here.

- :gem: New and improved dashboards
  - cDOT, high-level cluster overview dashboard
  - Headroom dashboard
  - Quota dashboard
  - Snapmirror dashboard shows source and destination side of relationship
  - NFS clients dashboard
  - Fabric Pool panels are now included in Volume dashboard
  - Tags are included for all default dashboards, making it easier to find what you need
  - Additional throughput, ops, and utilization panels were added to the Aggregate, Disk, and Clusters dashboards
  - Harvest dashboards updated to enable multi-select variables, shared crosshairs, better top n resources support,
    and all variables are sorted by default.

- :lock: Harvest code is checked for vulnerabilities on every commit
  using [Go's vulnerability management](https://go.dev/blog/vuln) scanner.

- Harvest collects additional metrics in this release
  - ONTAP S3 server config metrics
  - User defined volume workload
  - Active network connections
  - NFS connected clients
  - Network ports
  - Netstat packet loss

- Harvest now converts ONTAP histograms to Prometheus histograms, making it possible to visualize metrics as heatmaps in
  Grafana

## Announcements

:bangbang: **IMPORTANT** NetApp moved their communities from Slack to [Discord](https://discord.gg/ZmmWPHTBHw), please
join us [there](https://discordapp.com/channels/855068651522490400/1001963189124206732)!

:bomb: **Deprecation**: Earlier versions of Harvest published quota metrics prefixed with `qtree`. Harvest release 22.11
deprecates the quota metrics prefixed with `qtree` and instead publishes quota metrics prefixed with `quota`. All
dashboards have been updated. If you are consuming these metrics outside the default dashboards, please change
to `quota` prefixed metrics. Harvest release 23.02 will remove the deprecated quota metrics prefixed with `qtree`.

:bangbang: **IMPORTANT** If using Docker Compose and you want to keep your historical Prometheus data, please
read [how to migrate your Prometheus volume](https://github.com/NetApp/harvest/blob/main/docs/MigratePrometheusDocker.md)

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and
fixes. You can import them via the `bin/harvest/grafana import` CLI, from the Grafana UI, or from
the `Maintenance > Reset Harvest Dashboards` button in NAbox.

## Known Issues

- Harvest does not calculate power metrics for AFF A250 systems. This data is not available from ONTAP via ZAPI or REST.
  See ONTAP bug [1511476](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1511476) for more details.

- ONTAP does not include REST metrics for `offbox_vscan_server` and `offbox_vscan` until ONTAP 9.13.1. See ONTAP bug
  [1473892](https://burtview.netapp.com/burt/burt-bin/start?burt-id=1473892) for more details.

- Podman is unable to pull from NetApp's container registry `cr.netapp.io`.
  Until [issue](https://github.com/containers/podman/issues/15187) is resolved, Podman users can pull from a separate
  proxy like this `podman pull netappdownloads.jfrog.io/oss-docker-harvest-production/harvest:latest`.

**IMPORTANT** 7-mode filers that are not on the latest release of ONTAP may experience TLS connection issues with errors
like `tls: server selected unsupported protocol version 301` This is caused by a change in Go 1.18.
The [default for TLS client connections was changed to TLS 1.2](https://tip.golang.org/doc/go1.18#tls10) in Go 1.18.
Please upgrade your 7-mode filers (recommended) or set `tls_min_version: tls10` in
your `harvest.yml` [poller section](https://github.com/NetApp/harvest/tree/release/22.05.0#pollers). See #1007 for more
details.

The Unix collector is unable to monitor pollers running in containers.
See [#249](https://github.com/NetApp/harvest/issues/249) for details.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards
this release:

@Falcon667, @MrObvious, @ReneMeier, @Sawall10, @T1r0l, @chadpruden, @demalik, @electrocreative, @ev1963, @faguayot, @iStep2Step, @jgasher, @jmg011, @mamoep, @matthieu-sudo, @merdos, @rodenj1, Ed Wilts, KlausHub, MeghanaD, Paul P2, Rusty Brown, Shubham Mer, jf38800, rcl23, troysmuller


:seedling: This release includes 59 features, 90 bug fixes, 21 documentation, 4 testing, 2 styling, 6 refactoring, 2 miscellaneous, and 6 ci commits.

### :rocket: Features
- Enable Multi Select By Default ([#1213](https://github.com/NetApp/harvest/pull/1213))
- Merge Release 22.08 To Main ([#1218](https://github.com/NetApp/harvest/pull/1218))
- Add Avg Cifs Latency To Svm Dashboard Graph Panel ([#1221](https://github.com/NetApp/harvest/pull/1221))
- Network Port Templates ([#1231](https://github.com/NetApp/harvest/pull/1231))
- Add Node Cpu Busy To Cluster Dashboard ([#1243](https://github.com/NetApp/harvest/pull/1243))
- Improve Poller Startup Logging ([#1254](https://github.com/NetApp/harvest/pull/1254))
- Add Net Connections Template For Rest Collector ([#1257](https://github.com/NetApp/harvest/pull/1257))
- Upgrade Zapi Collector To Rest When The Ontap Version Is >= 9.12.1 ([#1261](https://github.com/NetApp/harvest/pull/1261))
- Run Govulncheck On `Make Dev` ([#1273](https://github.com/NetApp/harvest/pull/1273))
- Nfsv42 Restperf Templates ([#1275](https://github.com/NetApp/harvest/pull/1275))
- Enable User Defined Volume Workload ([#1276](https://github.com/NetApp/harvest/pull/1276))
- Prometheus Exporter Should Log Address And Port ([#1279](https://github.com/NetApp/harvest/pull/1279))
- Ensure Dashboard Units Align With Ontap's Units ([#1280](https://github.com/NetApp/harvest/pull/1280))
- Panels Should Connect Null Values ([#1281](https://github.com/NetApp/harvest/pull/1281))
- Harvest Should Collect Ontap S3 Server Metrics ([#1285](https://github.com/NetApp/harvest/pull/1285))
- `Bin/Zapi Show Counters` Should Print Xml Results To Make Parsi… ([#1286](https://github.com/NetApp/harvest/pull/1286))
- Harvest Should Collect Ontap S3 Server Config Metrics ([#1287](https://github.com/NetApp/harvest/pull/1287))
- Harvest Should Publish Cooked Zero Performance Metrics ([#1292](https://github.com/NetApp/harvest/pull/1292))
- Add Grafana Tags On Default Dashboards ([#1293](https://github.com/NetApp/harvest/pull/1293))
- Add Harvest Tags ([#1294](https://github.com/NetApp/harvest/pull/1294))
- Rest Nfs Connections Dashboard ([#1297](https://github.com/NetApp/harvest/pull/1297))
- Cmd Line `Objects` And `Collectors` Override Defaults ([#1300](https://github.com/NetApp/harvest/pull/1300))
- Harvest Should Replace Topk With Topk Range In All Dashboards Part 1 ([#1301](https://github.com/NetApp/harvest/pull/1301))
- Harvest Should Replace Topk With Topk Range In All Dashboards Part 2 ([#1302](https://github.com/NetApp/harvest/pull/1302))
- Harvest Should Replace Topk With Topk Range In All Dashboards Part 3 ([#1304](https://github.com/NetApp/harvest/pull/1304))
- Snapmirror From Source Side [Zapi Changes] ([#1307](https://github.com/NetApp/harvest/pull/1307))
- Mcc Plex Panel Fix ([#1310](https://github.com/NetApp/harvest/pull/1310))
- Add Support For Qos Min And Cp In Harvest ([#1316](https://github.com/NetApp/harvest/pull/1316))
- Add Available Ops To Headroom Dashboard ([#1317](https://github.com/NetApp/harvest/pull/1317))
- Added Panels In Cluster, Disk For 1.6 Parity ([#1320](https://github.com/NetApp/harvest/pull/1320))
- Add Storagegrid Collector And Dashboard ([#1322](https://github.com/NetApp/harvest/pull/1322))
- Export Ontap Histograms As Prometheus Histograms ([#1326](https://github.com/NetApp/harvest/pull/1326))
- Solution Based Cdot Dashboard ([#1336](https://github.com/NetApp/harvest/pull/1336))
- Cluster Var Changed To Source_cluster In Snapmirror Dashboard ([#1337](https://github.com/NetApp/harvest/pull/1337))
- Remove Pollinstance From Zapi Collector ([#1338](https://github.com/NetApp/harvest/pull/1338))
- Reduce Memory Footprint Of Set ([#1339](https://github.com/NetApp/harvest/pull/1339))
- Quota Metric Renaming ([#1345](https://github.com/NetApp/harvest/pull/1345))
- Collectors Should Log Polldata, Plugin Times, And Metadata ([#1347](https://github.com/NetApp/harvest/pull/1347))
- Export Ontap Histograms As Prometheus Histograms ([#1349](https://github.com/NetApp/harvest/pull/1349))
- Fabricpool Panels - Parity With 1.6 ([#1352](https://github.com/NetApp/harvest/pull/1352))
- All Dashboards Should Default To `Shared Crosshair` ([#1359](https://github.com/NetApp/harvest/pull/1359))
- All Dashboards Should Use Multi-Select Dropdowns For Each Variable ([#1363](https://github.com/NetApp/harvest/pull/1363))
- Perf Collector Unit Test Cases ([#1373](https://github.com/NetApp/harvest/pull/1373))
- Remove Metric Labels From Shelf Sensor Plugins ([#1378](https://github.com/NetApp/harvest/pull/1378))
- Envvar `Harvest_no_upgrade` To Skip Collector Upgrade ([#1385](https://github.com/NetApp/harvest/pull/1385))
- Rest Collector Should Not Log When Client_timeout Is Missing ([#1387](https://github.com/NetApp/harvest/pull/1387))
- Enable Rest Collector Templates ([#1391](https://github.com/NetApp/harvest/pull/1391))
- Harvest Should Use Rest Unconditionally Starting With 9.13.1 ([#1394](https://github.com/NetApp/harvest/pull/1394))
- Rest Perf Template Fixes ([#1395](https://github.com/NetApp/harvest/pull/1395))
- Only Allow One Config/Perf Collector Per Object ([#1410](https://github.com/NetApp/harvest/pull/1410))
- Histogram Support For Restperf ([#1412](https://github.com/NetApp/harvest/pull/1412))
- Volume Tag Plugin ([#1417](https://github.com/NetApp/harvest/pull/1417))
- Add Rest Validation In Ci ([#1421](https://github.com/NetApp/harvest/pull/1421))
- Add Netstat Metrics For Packet Loss ([#1423](https://github.com/NetApp/harvest/pull/1423))
- Add Datacenter To Metadata ([#1427](https://github.com/NetApp/harvest/pull/1427))
- Increase Dashboard Quality With More Tests ([#1460](https://github.com/NetApp/harvest/pull/1460))
- Add Node_disk_data_read To Units.yaml ([#1480](https://github.com/NetApp/harvest/pull/1480))
- Tag Fsx Dashboards ([#1490](https://github.com/NetApp/harvest/pull/1490))
- Restperf Submetric ([#1506](https://github.com/NetApp/harvest/pull/1506))

### :bug: Bug Fixes
- Log.fatalln Will Exit, And `Defer Resp.body.close()` Will Not Run ([#1211](https://github.com/NetApp/harvest/pull/1211))
- Remove Rewrite_as_label From Templates ([#1212](https://github.com/NetApp/harvest/pull/1212))
- Set User To Uid If Name Is Missing ([#1223](https://github.com/NetApp/harvest/pull/1223))
- Duplicate Instance Issue Quota ([#1225](https://github.com/NetApp/harvest/pull/1225))
- Create Unique Indexes For Quota Dashboard ([#1226](https://github.com/NetApp/harvest/pull/1226))
- Volume Dashboard Should Use Iec Bytes ([#1229](https://github.com/NetApp/harvest/pull/1229))
- Skipped Bookend Ems Whose Key Is Node-Name ([#1237](https://github.com/NetApp/harvest/pull/1237))
- Bin/Rest Should Support Verbose And Return Error When ([#1240](https://github.com/NetApp/harvest/pull/1240))
- Volume Plugin Should Not Fail When Snapmirror Has Empty Relationship_id ([#1241](https://github.com/NetApp/harvest/pull/1241))
- Volume.go Plugin Should Check No Instances ([#1253](https://github.com/NetApp/harvest/pull/1253))
- Remove Power 24H Panel From Shelf Dashboard ([#1256](https://github.com/NetApp/harvest/pull/1256))
- Govulncheck Scan Issue Go-2021-0113 ([#1259](https://github.com/NetApp/harvest/pull/1259))
- Negative Counter Fix And Zero Suppression ([#1260](https://github.com/NetApp/harvest/pull/1260))
- Remove User_id To Reduce Memory Load From Quota ([#1263](https://github.com/NetApp/harvest/pull/1263))
- Snapmirror Relationships From Source Side ([#1266](https://github.com/NetApp/harvest/pull/1266))
- Flashcache Dashboard Units Are Incorrect ([#1268](https://github.com/NetApp/harvest/pull/1268))
- Disable User,Group Quota By Default ([#1271](https://github.com/NetApp/harvest/pull/1271))
- Enable Dashboard Check In Ci ([#1277](https://github.com/NetApp/harvest/pull/1277))
- Http Sd Should Publish Local Ip When Exporter Is Blank ([#1278](https://github.com/NetApp/harvest/pull/1278))
- Headroom Dashboard Utilization Should Be In Percentage ([#1290](https://github.com/NetApp/harvest/pull/1290))
- Simple Poller Should Use Int64 Metric ([#1291](https://github.com/NetApp/harvest/pull/1291))
- Remove Label Warning From Rest Collector ([#1299](https://github.com/NetApp/harvest/pull/1299))
- Ignore Negative Perf Deltas ([#1303](https://github.com/NetApp/harvest/pull/1303))
- Mcc Plex Panel Fix Rest Template ([#1313](https://github.com/NetApp/harvest/pull/1313))
- Remove Duplicate Network Dashboards ([#1314](https://github.com/NetApp/harvest/pull/1314))
- 7Mode Zapi Cli Issue Due To Max ([#1321](https://github.com/NetApp/harvest/pull/1321))
- Add Scale To Headroom Dashboard ([#1323](https://github.com/NetApp/harvest/pull/1323))
- Increase Default Zapi Timeout To 30 Seconds ([#1333](https://github.com/NetApp/harvest/pull/1333))
- Zapiperf Lun Name Should Match Zapi ([#1341](https://github.com/NetApp/harvest/pull/1341))
- Record Number Of Zapi Instances In Polldata ([#1343](https://github.com/NetApp/harvest/pull/1343))
- Rest Metric Count ([#1346](https://github.com/NetApp/harvest/pull/1346))
- Aggregator.go Should Not Change Histogram Properties To Avg ([#1348](https://github.com/NetApp/harvest/pull/1348))
- Ci Ems Issue ([#1350](https://github.com/NetApp/harvest/pull/1350))
- Add Node In Warning Logs For Power Calculation ([#1351](https://github.com/NetApp/harvest/pull/1351))
- Align Aggregate Disk Utilization Panel ([#1355](https://github.com/NetApp/harvest/pull/1355))
- Correct Skip Count For Perf Percent Property ([#1358](https://github.com/NetApp/harvest/pull/1358))
- Harvest Should Keep Same Volume Name During Upgrade In Docker-Compose Workflow ([#1361](https://github.com/NetApp/harvest/pull/1361))
- Zapi Polldata Logged The Wrong Number Of Instances During Batch … ([#1366](https://github.com/NetApp/harvest/pull/1366))
- Top Latency Units Should Be Microseconds Not Milliseconds ([#1371](https://github.com/NetApp/harvest/pull/1371))
- Calculate Power From Voltage And Current Sensors When Power Units Are Not Known ([#1372](https://github.com/NetApp/harvest/pull/1372))
- Don't Add Units As Metric Labels Since It Breaks Influxdb Exporter ([#1376](https://github.com/NetApp/harvest/pull/1376))
- Handle Raidgroup/Plex Alongwith Other Changes ([#1380](https://github.com/NetApp/harvest/pull/1380))
- Disable Color Console Logging ([#1382](https://github.com/NetApp/harvest/pull/1382))
- Restperf Lun Name Should Match Zapi ([#1390](https://github.com/NetApp/harvest/pull/1390))
- Cluster Dashboard Panel Changes ([#1393](https://github.com/NetApp/harvest/pull/1393))
- Harvest Should Use Template Display Name When Exporting Histograms ([#1403](https://github.com/NetApp/harvest/pull/1403))
- Rest Collector Should Collect Cluster Level Instances ([#1404](https://github.com/NetApp/harvest/pull/1404))
- Remove Protected,Protectionrole,All_healthy Labels From Volume ([#1406](https://github.com/NetApp/harvest/pull/1406))
- Snapmirror Dashboard Changes ([#1407](https://github.com/NetApp/harvest/pull/1407))
- System Node Perf Template Fix ([#1409](https://github.com/NetApp/harvest/pull/1409))
- Svm Records Count ([#1411](https://github.com/NetApp/harvest/pull/1411))
- Dont Export Constituents Relationships In Sm ([#1414](https://github.com/NetApp/harvest/pull/1414))
- Handle Ls Relationships + Handle Dashboard ([#1415](https://github.com/NetApp/harvest/pull/1415))
- Snapmirror Dashboard Should Not Show Id Column ([#1416](https://github.com/NetApp/harvest/pull/1416))
- Handle Error When No Instances Found For Plugins In Rest ([#1428](https://github.com/NetApp/harvest/pull/1428))
- Handle Batching In Shelf Plugin ([#1429](https://github.com/NetApp/harvest/pull/1429))
- Lun Rest Perf Template Fixes ([#1430](https://github.com/NetApp/harvest/pull/1430))
- Handle Volume Panels ([#1431](https://github.com/NetApp/harvest/pull/1431))
- Align Rest Start Up Logging As Zapi ([#1435](https://github.com/NetApp/harvest/pull/1435))
- Handle Aggr_space_used_percent In Aggr ([#1439](https://github.com/NetApp/harvest/pull/1439))
- Sensor Plugin Rest Changes ([#1440](https://github.com/NetApp/harvest/pull/1440))
- Disk Dashboard - Variables Are Not Sorted ([#1443](https://github.com/NetApp/harvest/pull/1443))
- Add Missing Labels For Rest Zapi Diff ([#1445](https://github.com/NetApp/harvest/pull/1445))
- Shelf Child Obj - Status Ok To Normal ([#1451](https://github.com/NetApp/harvest/pull/1451))
- Restperf Key Handling ([#1452](https://github.com/NetApp/harvest/pull/1452))
- Rest Zapi Diff Error Handling ([#1453](https://github.com/NetApp/harvest/pull/1453))
- Restperf Fcp Template Mapping Fix ([#1455](https://github.com/NetApp/harvest/pull/1455))
- Disk Type In Lower Case In Rest ([#1456](https://github.com/NetApp/harvest/pull/1456))
- Power Fix For Cold Sensors ([#1464](https://github.com/NetApp/harvest/pull/1464))
- Svm With Private Cli ([#1465](https://github.com/NetApp/harvest/pull/1465))
- Storagegrid Collector Should Use Metadata Collection ([#1468](https://github.com/NetApp/harvest/pull/1468))
- Fix New Line Char In Headroom Dashboard ([#1473](https://github.com/NetApp/harvest/pull/1473))
- New_status Gap Issue For Cluster Scoped Zapi Call ([#1477](https://github.com/NetApp/harvest/pull/1477))
- Merge To Main From Release ([#1479](https://github.com/NetApp/harvest/pull/1479))
- Tenant Column Should Be In "Tenants And Buckets" Table Once ([#1483](https://github.com/NetApp/harvest/pull/1483))
- Fsx Headroom Dashboard Support For Rest Collector ([#1484](https://github.com/NetApp/harvest/pull/1484))
- Qos Rest Template Fix ([#1487](https://github.com/NetApp/harvest/pull/1487))
- Net Port Template Fix ([#1488](https://github.com/NetApp/harvest/pull/1488))
- Disable Netport Rest Template ([#1491](https://github.com/NetApp/harvest/pull/1491))
- Rest Sensor Template Fix ([#1492](https://github.com/NetApp/harvest/pull/1492))
- Fix Background Color In Cluster, Aggregate Panels ([#1496](https://github.com/NetApp/harvest/pull/1496))
- Smv_labels Missing In Zapi ([#1499](https://github.com/NetApp/harvest/pull/1499))
- 7Mode Shelf Plugin Fix ([#1500](https://github.com/NetApp/harvest/pull/1500))
- Remove Queue_full Counter From Namespace Template ([#1501](https://github.com/NetApp/harvest/pull/1501))
- Storagegrid Bucket Plugin Should Honor Client Timeout ([#1503](https://github.com/NetApp/harvest/pull/1503))
- Snapmirror Warn To Trace Log ([#1504](https://github.com/NetApp/harvest/pull/1504))
- Cdot Svm Panels ([#1515](https://github.com/NetApp/harvest/pull/1515))
- Svm Nfsv4 Panels Fix ([#1518](https://github.com/NetApp/harvest/pull/1518))
- Svm Copy Panel Fix ([#1520](https://github.com/NetApp/harvest/pull/1520))
- Exclude Empty Qtree In Restperf Template Through Regex ([#1522](https://github.com/NetApp/harvest/pull/1522))

### :closed_book: Documentation
- Rest Strategy Doc ([#1234](https://github.com/NetApp/harvest/pull/1234))
- Improve Security Panel Info For Ontap 9.10+ ([#1238](https://github.com/NetApp/harvest/pull/1238))
- Explain How To Enable Qos Collection When Using Least-Privilege… ([#1249](https://github.com/NetApp/harvest/pull/1249))
- Clarify When Harvest Defaults To Rest ([#1252](https://github.com/NetApp/harvest/pull/1252))
- Spelling Correction ([#1318](https://github.com/NetApp/harvest/pull/1318))
- Add Ems Alert To Ems Documentation ([#1319](https://github.com/NetApp/harvest/pull/1319))
- Explain How To Log To File With Systemd Instantiated Service ([#1325](https://github.com/NetApp/harvest/pull/1325))
- Add Help Text In Nfs Clients Dashboard About Enabling Rest Collector ([#1334](https://github.com/NetApp/harvest/pull/1334))
- Describe How To Migrate Historical Prometheus Data Generated Be… ([#1369](https://github.com/NetApp/harvest/pull/1369))
- Explain What To Do If Zapi Metrics Are Missing In Rest ([#1389](https://github.com/NetApp/harvest/pull/1389))
- Move Documentation To Separate Site ([#1433](https://github.com/NetApp/harvest/pull/1433))
- Readme Should Point To Https://Netapp.github.io/Harvest/ ([#1434](https://github.com/NetApp/harvest/pull/1434))
- Remove Unneeded Readme.md Files ([#1438](https://github.com/NetApp/harvest/pull/1438))
- Fix Image Links ([#1441](https://github.com/NetApp/harvest/pull/1441))
- Restore Rest Strategy And Migrate Docs ([#1463](https://github.com/NetApp/harvest/pull/1463))
- Explain What Negative Available Ops Means ([#1469](https://github.com/NetApp/harvest/pull/1469))
- Explain What Negative Available Ops Means (#1469) ([#1471](https://github.com/NetApp/harvest/pull/1471))
- Explain What Negative Available Ops Means ([#1474](https://github.com/NetApp/harvest/pull/1474))
- Add Amazon Fsx For Ontap Documentation ([#1485](https://github.com/NetApp/harvest/pull/1485))
- Include Docker Compose Workflow In Docs ([#1507](https://github.com/NetApp/harvest/pull/1507))
- Docker Upgrade Instructions ([#1511](https://github.com/NetApp/harvest/pull/1511))

### :wrench: Testing
- Add Unit Test That Finds Metrics Used In Dashboards With Confli… ([#1381](https://github.com/NetApp/harvest/pull/1381))
- Ensure Resetinstance Causes Metric To Be Skipped ([#1388](https://github.com/NetApp/harvest/pull/1388))
- Validate Sensor Template Fix ([#1486](https://github.com/NetApp/harvest/pull/1486))
- Add Unit Test To Detect Topk Without Range ([#1495](https://github.com/NetApp/harvest/pull/1495))

### Styling
- Address Shellcheck Strong Warnings ([#1228](https://github.com/NetApp/harvest/pull/1228))
- Correct Spelling And Lint Warning ([#1332](https://github.com/NetApp/harvest/pull/1332))

### Refactoring
- Use Map Instead Of Loop For `Targetisontap` ([#1235](https://github.com/NetApp/harvest/pull/1235))
- Remove Unused And Lightly-Used Metrics ([#1274](https://github.com/NetApp/harvest/pull/1274))
- Remove Warnings ([#1298](https://github.com/NetApp/harvest/pull/1298))
- Simplify Loadcollector ([#1329](https://github.com/NetApp/harvest/pull/1329))
- Simplify Set Api ([#1340](https://github.com/NetApp/harvest/pull/1340))
- Move Docs Out Of The Way To Make Way For New Ones ([#1422](https://github.com/NetApp/harvest/pull/1422))

### Miscellaneous
- Bump Dependencies ([#1331](https://github.com/NetApp/harvest/pull/1331))
- Increase Golangci-Lint Timeout ([#1364](https://github.com/NetApp/harvest/pull/1364))

### :hammer: CI
- Bump Go To 1.19 ([#1201](https://github.com/NetApp/harvest/pull/1201))
- Bump Go ([#1215](https://github.com/NetApp/harvest/pull/1215))
- Bump Golangci-Lint And Address Issues ([#1233](https://github.com/NetApp/harvest/pull/1233))
- Bump Go To 1.19.1 ([#1262](https://github.com/NetApp/harvest/pull/1262))
- Revert Jenkins File To 1.19 ([#1267](https://github.com/NetApp/harvest/pull/1267))
- Bump Golangci-Lint ([#1328](https://github.com/NetApp/harvest/pull/1328))

---

## 22.08.0 / 2022-08-19
<!-- git log --no-decorate --no-merges --cherry-pick --right-only --oneline origin/release/22.05.0...origin/release/22.08.0 -->

:rocket: Highlights of this major release include:

- :sparkler: an ONTAP event management system (EMS) events collector with [64 events out-of-the-box](https://github.com/NetApp/harvest/blob/main/conf/ems/9.6.0/ems.yaml)

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

- :sparkler: Harvest adds an [ONTAP event management system (EMS) events](https://github.com/NetApp/harvest/blob/main/cmd/collectors/ems/README.md) collector in this release.
  It collects ONTAP events, exports them to Prometheus, and provides integration with Prometheus AlertManager.
  [Full list of 64 events](https://github.com/NetApp/harvest/blob/main/conf/ems/9.6.0/ems.yaml)

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

- We've included more out-of-the-box [Prometheus alerts](https://github.com/NetApp/harvest/blob/main/container/prometheus/alert_rules.yml). Keep sharing your most useful alerts!

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

- Improve Harvest's container compatibility with K8s via kompose. [#655](https://github.com/NetApp/harvest/pull/655) See [also](https://github.com/NetApp/harvest/tree/main/container/k8) and [discussion](https://github.com/NetApp/harvest/discussions/827)

- The ZAPI cli tool should include counter types when querying ZAPIs #663

- Harvest should include a richer set of Prometheus alerts #254 Thanks @demalik for raising

- Template plugins should run in the order they are defined and compose better.
  The output of one plugin can be fed into the input of the next one. #736 Thanks to @chadpruden for raising

- Harvest should collect Antivirus counters when ONTAP offbox vscan is configured [#346](https://github.com/NetApp/harvest/issues/346) Thanks to @burkl and @Falcon667 for reporting

- [Document](https://github.com/NetApp/harvest/tree/main/container/containerd) how to run Harvest with `containerd` and `Rancher`

- Qtree counters should be collected for 7-mode filers #766 Thanks to @jmg011 for raising this issue and iterating with us on a solution

- Harvest admin node should work with pollers running in Docker compose [#678](https://github.com/NetApp/harvest/pull/678)

- [Document](https://github.com/NetApp/harvest/tree/main/container/podman) how to run Harvest with Podman. Several RHEL customers asked about this since Podman ships as the default container runtime on RHEL8+.

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

- Improve Docker deployment documentation [4019308](https://github.com/NetApp/harvest/tree/main/container/onePollerPerContainer)

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
