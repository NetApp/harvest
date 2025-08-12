## 25.08.0 / 2025-08-13 Release
:pushpin: Highlights of this major release include:
## :star: New Features

- [StatPerf](https://netapp.github.io/harvest/latest/configure-statperf/) Collector
  - StatPerf collects performance metrics from ONTAP by invoking the ONTAP CLI statistics command via the private REST CLI.
  - This collector is designed for environments where ZapiPerf, RestPerf, or KeyPerf collectors cannot be used.

- :gem: Two New Dashboards:
  - **MAV (Multi Admin Verify) Request Dashboard**: Provides a real-time overview of Multi-Admin Verification requests, tracking their status, approvals, and pending actions for enhanced security and operational visibility.
  - **FPolicy Dashboard**: Facilitates monitoring of FPolicy performance metrics at the policy, SVM, and server levels.

- **Cisco Switch Dashboard Updates**: Thanks to @roybatty2019 for the input.
  - Individual fan speeds are now displayed separately from zone speeds.
  - LLDP and CDP parsing have been refined with consistent label naming in metrics and improved data handling.
  - Introduction of new traffic monitoring metrics.

- :star: Enhancements:
  - Quota and FSA dashboards now support filtering by volume tags.
  - Added a Junction Path variable in the Volume dashboard.
  - Added bucket quotas in the StorageGRID Tenant dashboard.
  - Inclusion of a Volume column in the SMB Dashboard's CIFS sessions table.
  - Added Used% in the bucket table within the Tenant dashboard.

- :closed_book: Documentation Additions:
  - Metric documentation now includes details of the dashboard panel where each metric is utilized.
  - Added documentation for Cisco Switch and StorageGRID metrics.

## Announcements

:bulb: **IMPORTANT** After upgrade, don't forget to re-import your dashboards, so you get all the new enhancements and fixes. You can import them via the 'bin/harvest grafana import' CLI, from the Grafana UI, or from the 'Maintenance > Reset Harvest Dashboards' button in NAbox3. For NAbox4, this step is not needed.

## Known Issues

:bulb: **IMPORTANT** FSx ZapiPerf workload collector fails to collect metrics, please use RestPerf instead.

## Thanks to all the awesome contributors

:metal: Thanks to all the people who've opened issues, asked questions on Discord, and contributed code or dashboards
this release:

@BrendonA667, @Falcon667, @T1r0l, @anguswilliams, @datamuc, @jowanw, @mamoep, @mhbeh, @mishraavinash88, @roybatty2019

:seedling: This release includes 18 features, 19 bug fixes, 9 documentation, 4 refactoring, 1 miscellaneous, and 8 ci pull requests.

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

---
