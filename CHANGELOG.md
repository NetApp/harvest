# Change Log
## [Releases](https://github.com/NetApp/harvest/releases)

## 24.02.0 / 2024-02-16 Release
:pushpin: Highlights of this major release include:
- New Datacenter dashboard which contains node health, capacity, performance, storage efficiency, issues, snapshot, power, and temperature details.
- Harvest includes SnapMirror active sync EMS events with alert rules. Thanks @Nikhita-13 for reporting.
- Harvest monitors FlexCache performance metrics and includes a new FlexCache dashboard to visualize them. Thanks to @ewilts for raising.
- Harvest detects HA pair down and sensor failures. These are shown in the Health dashboard. Thanks to @johnwarlick for raising.
- Harvest monitors MetroCluster diagnostics and shows them in the MetroCluster dashboard. Thanks to @wagneradrian92 for reporting.
- We improved the performance of all dashboards that include topk queries. Thanks to @mamoep for reporting!
- We added filter support for the ZapiPerf collector. See [filter](https://netapp.github.io/harvest/nightly/configure-zapi/#filter) for more detail. Thanks to @debbrata-netapp for reporting.
- A `bin/harvest grafana customize` command that writes the dashboards to the filesystem so other programs can manage them. Thanks to @nicolai-hornung-bl for reporting!

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
  - Updated how to [collect Harvest logs](https://netapp.github.io/harvest/latest/help/log-collection/)  
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
- Handled Missing Uuid In Volume For Change_log (#2478) ([#2479](https://github.com/NetApp/harvest/pull/2479))

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
