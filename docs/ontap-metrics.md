This document contains details about Harvest metrics and their relevant ONTAP ZAPI and REST API mappings.

Details about which Harvest metrics each dashboard uses can be generated on demand by running `bin/harvest grafana metrics`. See
[#1577](https://github.com/NetApp/harvest/issues/1577#issue-1471478260) for details.

```
Creation Date : 2023-Apr-27
ONTAP Version: 9.12.1
```
## Understanding the structure

Below is an <span class="key">annotated</span> example of how to interpret the structure of each of the [metrics](#metrics).

disk_io_queued <span class="key">Name of the metric exported by Harvest</span>

Number of I/Os queued to the disk but not yet issued <span class="key">Description of the ONTAP metric</span>

* <span class="key">API</span> will be one of REST or ZAPI depending on which collector is used to collect the metric
* <span class="key">Endpoint</span> name of the REST or ZAPI API used to collect this metric
* <span class="key">Metric</span> name of the ONTAP metric
 <span class="key">Template</span> path of the template that collects the metric

Performance related metrics also include:

- <span class="key">Unit</span> the unit of the metric
- <span class="key">Type</span> describes how to calculate a cooked metric from two consecutive ONTAP raw metrics
- <span class="key">Base</span> some counters require a `base counter` for post-processing. When required, this property lists the `base counter`

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
|REST | `api/cluster/counter/tables/disk:constituent` | `io_queued`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> base_for_disk_busy | conf/restperf/9.12.0/disk.yaml|
|ZAPI | `perf-object-get-instances disk:constituent` | `io_queued`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> base_for_disk_busy | conf/zapiperf/cdot/9.8.0/disk.yaml|

## Metrics


### aggr_efficiency_savings

Space saved by storage efficiencies (logical_used - used)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.efficiency.savings` | conf/rest/9.12.0/aggr.yaml |


### aggr_efficiency_savings_wo_snapshots

Space saved by storage efficiencies (logical_used - used)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.efficiency_without_snapshots.savings` | conf/rest/9.12.0/aggr.yaml |


### aggr_efficiency_savings_wo_snapshots_flexclones

Space saved by storage efficiencies (logical_used - used)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.efficiency_without_snapshots_flexclones.savings` | conf/rest/9.12.0/aggr.yaml |


### aggr_hybrid_cache_size_total

Total usable space in bytes of SSD cache. Only provided when hybrid_cache.enabled is 'true'.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `block_storage.hybrid_cache.size` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.hybrid-cache-size-total` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_hybrid_disk_count

Number of disks used in the cache tier of the aggregate. Only provided when hybrid_cache.enabled is 'true'.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `block_storage.hybrid_cache.disk_count` | conf/rest/9.12.0/aggr.yaml |


### aggr_inode_files_private_used

Number of system metadata files used. If the referenced file system is restricted or offline, a value of 0 is returned.This is an advanced property; there is an added computational cost to retrieving its value. The field is not populated for either a collection GET or an instance GET unless it is explicitly requested using the <i>fields</i> query parameter containing either footprint or **.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `inode_attributes.files_private_used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-inode-attributes.files-private-used` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_inode_files_total

Maximum number of user-visible files that this referenced file system can currently hold. If the referenced file system is restricted or offline, a value of 0 is returned.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `inode_attributes.files_total` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-inode-attributes.files-total` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_inode_files_used

Number of user-visible files used in the referenced file system. If the referenced file system is restricted or offline, a value of 0 is returned.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `inode_attributes.files_used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-inode-attributes.files-used` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_inode_inodefile_private_capacity

Number of files that can currently be stored on disk for system metadata files. This number will dynamically increase as more system files are created.This is an advanced property; there is an added computationl cost to retrieving its value. The field is not populated for either a collection GET or an instance GET unless it is explicitly requested using the <i>fields</i> query parameter containing either footprint or **.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `inode_attributes.file_private_capacity` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-inode-attributes.inodefile-private-capacity` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_inode_inodefile_public_capacity

Number of files that can currently be stored on disk for user-visible files.  This number will dynamically increase as more user-visible files are created.This is an advanced property; there is an added computational cost to retrieving its value. The field is not populated for either a collection GET or an instance GET unless it is explicitly requested using the <i>fields</i> query parameter containing either footprint or **.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `inode_attributes.file_public_capacity` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-inode-attributes.inodefile-public-capacity` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_inode_maxfiles_available

The count of the maximum number of user-visible files currently allowable on the referenced file system.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `inode_attributes.max_files_available` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-inode-attributes.maxfiles-available` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_inode_maxfiles_possible

The largest value to which the maxfiles-available parameter can be increased by reconfiguration, on the referenced file system.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `inode_attributes.max_files_possible` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-inode-attributes.maxfiles-possible` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_inode_maxfiles_used

The number of user-visible files currently in use on the referenced file system.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `inode_attributes.max_files_used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-inode-attributes.maxfiles-used` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_inode_used_percent

The percentage of disk space currently in use based on user-visible file count on the referenced file system.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `inode_attributes.used_percent` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-inode-attributes.percent-inode-used-capacity` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_logical_used_wo_snapshots

Logical used

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.efficiency_without_snapshots.logical_used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-efficiency-get-iter` | `aggr-efficiency-info.aggr-efficiency-cumulative-info.total-data-reduction-logical-used-wo-snapshots` | conf/zapi/cdot/9.9.0/aggr_efficiency.yaml |


### aggr_logical_used_wo_snapshots_flexclones

Logical used

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.efficiency_without_snapshots_flexclones.logical_used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-efficiency-get-iter` | `aggr-efficiency-info.aggr-efficiency-cumulative-info.total-data-reduction-logical-used-wo-snapshots-flexclones` | conf/zapi/cdot/9.9.0/aggr_efficiency.yaml |


### aggr_physical_used_wo_snapshots

Total Data Reduction Physical Used Without Snapshots

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `aggr-efficiency-get-iter` | `aggr-efficiency-info.aggr-efficiency-cumulative-info.total-data-reduction-physical-used-wo-snapshots` | conf/zapi/cdot/9.9.0/aggr_efficiency.yaml |
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/aggr.yaml |


### aggr_physical_used_wo_snapshots_flexclones

Total Data Reduction Physical Used without snapshots and flexclones

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `aggr-efficiency-get-iter` | `aggr-efficiency-info.aggr-efficiency-cumulative-info.total-data-reduction-physical-used-wo-snapshots-flexclones` | conf/zapi/cdot/9.9.0/aggr_efficiency.yaml |
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/aggr.yaml |


### aggr_power

Power consumed by aggregate in Watts.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/disk.yaml | 
| ZAPI | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/disk.yaml | 


### aggr_primary_disk_count

Number of disks used in the aggregate. This includes parity disks, but excludes disks in the hybrid cache.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `block_storage.primary.disk_count` | conf/rest/9.12.0/aggr.yaml |


### aggr_raid_disk_count

Number of disks in the aggregate.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-raid-attributes.disk-count` | conf/zapi/cdot/9.8.0/aggr.yaml |
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/aggr.yaml |


### aggr_raid_plex_count

Number of plexes in the aggregate

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `block_storage.plexes.#` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-raid-attributes.plex-count` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_raid_size

Option to specify the maximum number of disks that can be included in a RAID group.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `block_storage.primary.raid_size` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-raid-attributes.raid-size` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_snapshot_files_total

Total files allowed in Snapshot copies

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `snapshot.files_total` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-snapshot-attributes.files-total` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_snapshot_files_used

Total files created in Snapshot copies

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `snapshot.files_used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-snapshot-attributes.files-used` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_snapshot_inode_used_percent

The percentage of disk space currently in use based on user-visible file (inode) count on the referenced file system.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-snapshot-attributes.percent-inode-used-capacity` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_snapshot_maxfiles_available

Maximum files available for Snapshot copies

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `snapshot.max_files_available` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-snapshot-attributes.maxfiles-available` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_snapshot_maxfiles_possible

The largest value to which the maxfiles-available parameter can be increased by reconfiguration, on the referenced file system.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-snapshot-attributes.maxfiles-possible` | conf/zapi/cdot/9.8.0/aggr.yaml |
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/aggr.yaml |


### aggr_snapshot_maxfiles_used

Files in use by Snapshot copies

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `snapshot.max_files_used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-snapshot-attributes.maxfiles-used` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_snapshot_reserve_percent

Percentage of space reserved for Snapshot copies

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.snapshot.reserve_percent` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-snapshot-attributes.snapshot-reserve-percent` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_snapshot_size_available

Available space for Snapshot copies in bytes

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.snapshot.available` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-snapshot-attributes.size-available` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_snapshot_size_total

Total space for Snapshot copies in bytes

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.snapshot.total` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-snapshot-attributes.size-total` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_snapshot_size_used

Space used by Snapshot copies in bytes

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.snapshot.used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-snapshot-attributes.size-used` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_snapshot_used_percent

Percentage of disk space used by Snapshot copies

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.snapshot.used_percent` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-snapshot-attributes.percent-used-capacity` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_available

Space available in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.available` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.size-available` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_capacity_tier_used

Used space in bytes in the cloud store. Only applicable for aggregates with a cloud store tier.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.cloud_storage.used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.capacity-tier-used` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_data_compacted_count

Amount of compacted data in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.data_compacted_count` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.data-compacted-count` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_data_compaction_saved

Space saved in bytes by compacting the data.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.data_compaction_space_saved` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.data-compaction-space-saved` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_data_compaction_saved_percent

Percentage saved by compacting the data.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.data_compaction_space_saved_percent` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.data-compaction-space-saved-percent` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_performance_tier_inactive_user_data

The size that is physically used in the block storage and has a cold temperature, in bytes. This property is only supported if the aggregate is either attached to a cloud store or can be attached to a cloud store.This is an advanced property; there is an added computational cost to retrieving its value. The field is not populated for either a collection GET or an instance GET unless it is explicitly requested using the <i>fields</i> query parameter containing either block_storage.inactive_user_data or **.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.inactive_user_data` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.performance-tier-inactive-user-data` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_performance_tier_inactive_user_data_percent

The percentage of inactive user data in the block storage. This property is only supported if the aggregate is either attached to a cloud store or can be attached to a cloud store.This is an advanced property; there is an added computational cost to retrieving its value. The field is not populated for either a collection GET or an instance GET unless it is explicitly requested using the <i>fields</i> query parameter containing either block_storage.inactive_user_data_percent or **.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.inactive_user_data_percent` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.performance-tier-inactive-user-data-percent` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_physical_used

Total physical used size of an aggregate in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.physical_used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.physical-used` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_physical_used_percent

Physical used percentage.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.physical_used_percent` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.physical-used-percent` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_reserved

The total disk space in bytes that is reserved on the referenced file system. The reserved space is already counted in the used space, so this element can be used to see what portion of the used space represents space reserved for future use.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.total-reserved-space` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_sis_saved

Amount of space saved in bytes by storage efficiency.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.volume_deduplication_space_saved` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.sis-space-saved` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_sis_saved_percent

Percentage of space saved by storage efficiency.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.volume_deduplication_space_saved_percent` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.sis-space-saved-percent` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_sis_shared_count

Amount of shared bytes counted by storage efficiency.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.volume_deduplication_shared_count` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.sis-shared-count` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_total

Total usable space in bytes, not including WAFL reserve and aggregate Snapshot copy reserve.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.size` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.size-total` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_used

Space used or reserved in bytes. Includes volume guarantees and aggregate metadata.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.block_storage.used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.size-used` | conf/zapi/cdot/9.8.0/aggr.yaml |


### aggr_space_used_percent

The percentage of disk space currently in use on the referenced file system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-space-attributes.percent-used-capacity` | conf/zapi/cdot/9.8.0/aggr.yaml |
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/aggr.yaml |


### aggr_total_logical_used

Logical used

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `space.efficiency.logical_used` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-efficiency-get-iter` | `aggr-efficiency-info.aggr-efficiency-cumulative-info.total-logical-used` | conf/zapi/cdot/9.9.0/aggr_efficiency.yaml |


### aggr_total_physical_used

Total Physical Used

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `aggr-efficiency-get-iter` | `aggr-efficiency-info.aggr-efficiency-cumulative-info.total-physical-used` | conf/zapi/cdot/9.9.0/aggr_efficiency.yaml |
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/aggr.yaml |


### aggr_volume_count_flexvol

Number of flexvol volumes in the aggregate.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/aggregates` | `volume_count` | conf/rest/9.12.0/aggr.yaml |
| ZAPI | `aggr-get-iter` | `aggr-attributes.aggr-volume-count-attributes.flexvol-count` | conf/zapi/cdot/9.8.0/aggr.yaml |


### cifs_session_connection_count

A counter used to track requests that are sent to the volumes to the node.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/protocols/cifs/sessions` | `connection_count` | conf/rest/9.8.0/cifs_session.yaml |
| ZAPI | `cifs-session-get-iter` | `cifs-session.connection-count` | conf/zapi/cdot/9.8.0/cifs_session.yaml |


### cloud_target_used

The amount of cloud space used by all the aggregates attached to the target, in bytes. This field is only populated for FabricPool targets. The value is recalculated once every 5 minutes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cloud/targets` | `used` | conf/rest/9.12.0/cloud_target.yaml |


### cluster_new_status

It is an indicator of the overall health status of the cluster, with a value of 1 indicating a healthy status and a value of 0 indicating an unhealthy status.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/subsystem.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/subsystem.yaml |


### cluster_subsystem_outstanding_alerts

Number of outstanding alerts

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/system/health/subsystem` | `outstanding_alert_count` | conf/rest/9.12.0/subsystem.yaml |
| ZAPI | `diagnosis-subsystem-config-get-iter` | `diagnosis-subsystem-config-info.outstanding-alert-count` | conf/zapi/cdot/9.8.0/subsystem.yaml |


### cluster_subsystem_suppressed_alerts

Number of suppressed alerts

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/system/health/subsystem` | `suppressed_alert_count` | conf/rest/9.12.0/subsystem.yaml |
| ZAPI | `diagnosis-subsystem-config-get-iter` | `diagnosis-subsystem-config-info.suppressed-alert-count` | conf/zapi/cdot/9.8.0/subsystem.yaml |


### copy_manager_bce_copy_count_curr

Current number of copy requests being processed by the Block Copy Engine.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/copy_manager` | `block_copy_engine_current_copy_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/copy_manager.yaml | 
| ZAPI | `perf-object-get-instances copy_manager` | `bce_copy_count_curr`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/copy_manager.yaml | 


### copy_manager_kb_copied

Sum of kilo-bytes copied.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/copy_manager` | `KB_copied`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/copy_manager.yaml | 
| ZAPI | `perf-object-get-instances copy_manager` | `KB_copied`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/copy_manager.yaml | 


### copy_manager_ocs_copy_count_curr

Current number of copy requests being processed by the ONTAP copy subsystem.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/copy_manager` | `ontap_copy_subsystem_current_copy_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/copy_manager.yaml | 
| ZAPI | `perf-object-get-instances copy_manager` | `ocs_copy_count_curr`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/copy_manager.yaml | 


### copy_manager_sce_copy_count_curr

Current number of copy requests being processed by the System Continuous Engineering.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/copy_manager` | `system_continuous_engineering_current_copy_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/copy_manager.yaml | 
| ZAPI | `perf-object-get-instances copy_manager` | `sce_copy_count_curr`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/copy_manager.yaml | 


### copy_manager_spince_copy_count_curr

Current number of copy requests being processed by the SpinCE.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/copy_manager` | `spince_current_copy_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/copy_manager.yaml | 
| ZAPI | `perf-object-get-instances copy_manager` | `spince_copy_count_curr`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/copy_manager.yaml | 


### disk_bytes_per_sector

Bytes per sector.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/disks` | `bytes_per_sector` | conf/rest/9.12.0/disk.yaml |
| ZAPI | `storage-disk-get-iter` | `storage-disk-info.disk-inventory-info.bytes-per-sector` | conf/zapi/cdot/9.8.0/disk.yaml |


### disk_power_on_hours

Hours powered on.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/disks` | `stats.power_on_hours` | conf/rest/9.12.0/disk.yaml |


### disk_sectors

Number of sectors on the disk.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/disks` | `sector_count` | conf/rest/9.12.0/disk.yaml |
| ZAPI | `storage-disk-get-iter` | `storage-disk-info.disk-inventory-info.capacity-sectors` | conf/zapi/cdot/9.8.0/disk.yaml |


### disk_stats_average_latency

Average I/O latency across all active paths, in milliseconds.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/disks` | `stats.average_latency` | conf/rest/9.12.0/disk.yaml |
| ZAPI | `storage-disk-get-iter` | `storage-disk-info.disk-stats-info.average-latency` | conf/zapi/cdot/9.8.0/disk.yaml |


### disk_stats_io_kbps

Total Disk Throughput in KBPS Across All Active Paths

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `storage-disk-get-iter` | `storage-disk-info.disk-stats-info.disk-io-kbps` | conf/zapi/cdot/9.8.0/disk.yaml |
| REST | `api/private/cli/disk` | `disk_io_kbps_total` | conf/rest/9.12.0/disk.yaml |


### disk_stats_sectors_read

Number of Sectors Read

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `storage-disk-get-iter` | `storage-disk-info.disk-stats-info.sectors-read` | conf/zapi/cdot/9.8.0/disk.yaml |
| REST | `api/private/cli/disk` | `sectors_read` | conf/rest/9.12.0/disk.yaml |


### disk_stats_sectors_written

Number of Sectors Written

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `storage-disk-get-iter` | `storage-disk-info.disk-stats-info.sectors-written` | conf/zapi/cdot/9.8.0/disk.yaml |
| REST | `api/private/cli/disk` | `sectors_written` | conf/rest/9.12.0/disk.yaml |


### disk_uptime

Number of seconds the drive has been powered on

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `storage-disk-get-iter` | `storage-disk-info.disk-stats-info.power-on-time-interval` | conf/zapi/cdot/9.8.0/disk.yaml |
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/disk.yaml |


### disk_usable_size

Usable size of each disk, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/disks` | `usable_size` | conf/rest/9.12.0/disk.yaml |


### environment_sensor_average_ambient_temperature

Average temperature of all ambient sensors for node in Celsius.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/sensor.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/sensor.yaml |


### environment_sensor_average_fan_speed

Average fan speed for node in rpm.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/sensor.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/sensor.yaml |


### environment_sensor_average_temperature

Average temperature of all non-ambient sensors for node in Celsius.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/sensor.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/sensor.yaml |


### environment_sensor_max_fan_speed

Maximum fan speed for node in rpm.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/sensor.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/sensor.yaml |


### environment_sensor_max_temperature

Maximum temperature of all non-ambient sensors for node in Celsius.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/sensor.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/sensor.yaml |


### environment_sensor_min_ambient_temperature

Minimum temperature of all ambient sensors for node in Celsius.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/sensor.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/sensor.yaml |


### environment_sensor_min_fan_speed

Minimum fan speed for node in rpm.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/sensor.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/sensor.yaml |


### environment_sensor_min_temperature

Minimum temperature of all non-ambient sensors for node in Celsius.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/sensor.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/sensor.yaml |


### environment_sensor_power

Power consumed by a node in Watts.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/sensor.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/sensor.yaml |


### environment_sensor_threshold_value

Provides the sensor reading.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/sensors` | `value` | conf/rest/9.12.0/sensor.yaml |
| ZAPI | `environment-sensors-get-iter` | `environment-sensors-info.threshold-sensor-value` | conf/zapi/cdot/9.8.0/sensor.yaml |


### fabricpool_average_latency

Note This counter is deprecated and will be removed in a future release.  Average latencies executed during various phases of command execution. The execution-start latency represents the average time taken to start executing a operation. The request-prepare latency represent the average time taken to prepare the commplete request that needs to be sent to the server. The send latency represents the average time taken to send requests to the server. The execution-start-to-send-complete represents the average time taken to send a operation out since its execution started. The execution-start-to-first-byte-received represent the average time taken to to receive the first byte of a response since the command&apos;s request execution started. These counters can be used to identify performance bottlenecks within the object store client module.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_client_op` | `average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> ops | conf/zapiperf/cdot/9.8.0/object_store_client_op.yaml | 


### fabricpool_cloud_bin_op_latency_average

Cloud bin operation latency average in milliseconds.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_comp_aggr_vol_bin` | `cloud_bin_op_latency_average`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_comp_aggr_vol_bin.yaml | 
| ZAPI | `perf-object-get-instances wafl_comp_aggr_vol_bin` | `cloud_bin_op_latency_average`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_comp_aggr_vol_bin.yaml | 


### fabricpool_cloud_bin_operation

Cloud bin operation counters.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_comp_aggr_vol_bin` | `cloud_bin_op`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_comp_aggr_vol_bin.yaml | 
| ZAPI | `perf-object-get-instances wafl_comp_aggr_vol_bin` | `cloud_bin_operation`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_comp_aggr_vol_bin.yaml | 


### fabricpool_get_throughput_bytes

Note This counter is deprecated and will be removed in a future release.  Counter that indicates the throughput for GET command in bytes per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_client_op` | `get_throughput_bytes`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/object_store_client_op.yaml | 


### fabricpool_put_throughput_bytes

Note This counter is deprecated and will be removed in a future release.  Counter that indicates the throughput for PUT command in bytes per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_client_op` | `put_throughput_bytes`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/object_store_client_op.yaml | 


### fabricpool_stats

Note This counter is deprecated and will be removed in a future release.  Counter that indicates the number of object store operations sent, and their success and failure counts. The objstore_client_op_name array indicate the operation name such as PUT, GET, etc. The objstore_client_op_stats_name array contain the total number of operations, their success and failure counter for each operation.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_client_op` | `stats`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/object_store_client_op.yaml | 


### fabricpool_throughput_ops

Counter that indicates the throughput for commands in ops per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_client_op` | `throughput_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/object_store_client_op.yaml | 


### fcp_avg_other_latency

Average latency for operations other than read and write

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `average_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> other_ops | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `avg_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> other_ops | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_avg_read_latency

Average latency for read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `average_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `avg_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_avg_write_latency

Average latency for write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `average_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `avg_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_discarded_frames_count

Number of discarded frames.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `discarded_frames_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `discarded_frames_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_int_count

Number of interrupts

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `interrupt_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `int_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_invalid_crc

Number of invalid cyclic redundancy checks (CRC count)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `invalid.crc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `invalid_crc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_invalid_transmission_word

Number of invalid transmission words

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `invalid.transmission_word`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `invalid_transmission_word`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_isr_count

Number of interrupt responses

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `isr.count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `isr_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_lif_avg_latency

Average latency for FCP operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp_lif` | `average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/restperf/9.12.0/fcp_lif.yaml | 
| ZAPI | `perf-object-get-instances fcp_lif` | `avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/zapiperf/cdot/9.8.0/fcp_lif.yaml | 


### fcp_lif_avg_other_latency

Average latency for operations other than read and write

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp_lif` | `average_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> other_ops | conf/restperf/9.12.0/fcp_lif.yaml | 
| ZAPI | `perf-object-get-instances fcp_lif` | `avg_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> other_ops | conf/zapiperf/cdot/9.8.0/fcp_lif.yaml | 


### fcp_lif_avg_read_latency

Average latency for read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp_lif` | `average_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/restperf/9.12.0/fcp_lif.yaml | 
| ZAPI | `perf-object-get-instances fcp_lif` | `avg_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.8.0/fcp_lif.yaml | 


### fcp_lif_avg_write_latency

Average latency for write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp_lif` | `average_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/restperf/9.12.0/fcp_lif.yaml | 
| ZAPI | `perf-object-get-instances fcp_lif` | `avg_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/zapiperf/cdot/9.8.0/fcp_lif.yaml | 


### fcp_lif_other_ops

Number of operations that are not read or write.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp_lif` | `other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp_lif.yaml | 
| ZAPI | `perf-object-get-instances fcp_lif` | `other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp_lif.yaml | 


### fcp_lif_read_data

Amount of data read from the storage system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp_lif` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp_lif.yaml | 
| ZAPI | `perf-object-get-instances fcp_lif` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp_lif.yaml | 


### fcp_lif_read_ops

Number of read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp_lif` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp_lif.yaml | 
| ZAPI | `perf-object-get-instances fcp_lif` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp_lif.yaml | 


### fcp_lif_total_ops

Total number of operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp_lif` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp_lif.yaml | 
| ZAPI | `perf-object-get-instances fcp_lif` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp_lif.yaml | 


### fcp_lif_write_data

Amount of data written to the storage system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp_lif` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp_lif.yaml | 
| ZAPI | `perf-object-get-instances fcp_lif` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp_lif.yaml | 


### fcp_lif_write_ops

Number of write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp_lif` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp_lif.yaml | 
| ZAPI | `perf-object-get-instances fcp_lif` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp_lif.yaml | 


### fcp_link_down

Number of times the Fibre Channel link was lost

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `link.down`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `link_down`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_link_failure

Number of link failures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `link_failure`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `link_failure`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_loss_of_signal

Number of times this port lost signal

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `loss_of_signal`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `loss_of_signal`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_loss_of_sync

Number of times this port lost sync

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `loss_of_sync`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `loss_of_sync`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_nvmf_avg_other_latency

Average latency for operations other than read and write (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.average_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf.other_ops | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_avg_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf_other_ops | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_avg_read_latency

Average latency for read operations (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.average_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf.read_ops | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_avg_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf_read_ops | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_avg_remote_other_latency

Average latency for remote operations other than read and write (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.average_remote_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf_remote.other_ops | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_avg_remote_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf_remote_other_ops | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_avg_remote_read_latency

Average latency for remote read operations (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.average_remote_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf_remote.read_ops | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_avg_remote_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf_remote_read_ops | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_avg_remote_write_latency

Average latency for remote write operations (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.average_remote_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf_remote.write_ops | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_avg_remote_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf_remote_write_ops | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_avg_write_latency

Average latency for write operations (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.average_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf.write_ops | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_avg_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nvmf_write_ops | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_caw_data

Amount of CAW data sent to the storage system (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.caw_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_caw_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_caw_ops

Number of FC-NVMe CAW operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.caw_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_caw_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_command_slots

Number of command slots that have been used by initiators logging into this port. This shows the command fan-in on the port.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.command_slots`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_command_slots`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_other_ops

Number of NVMF operations that are not read or write.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_read_data

Amount of data read from the storage system (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_read_ops

Number of FC-NVMe read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_remote_caw_data

Amount of remote CAW data sent to the storage system (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf_remote.caw_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_remote_caw_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_remote_caw_ops

Number of FC-NVMe remote CAW operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf_remote.caw_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_remote_caw_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_remote_other_ops

Number of NVMF remote operations that are not read or write.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf_remote.other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_remote_other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_remote_read_data

Amount of remote data read from the storage system (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf_remote.read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_remote_read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_remote_read_ops

Number of FC-NVMe remote read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf_remote.read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_remote_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_remote_total_data

Amount of remote FC-NVMe traffic to and from the storage system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf_remote.total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_remote_total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_remote_total_ops

Total number of remote FC-NVMe operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf_remote.total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_remote_total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_remote_write_data

Amount of remote data written to the storage system (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf_remote.write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_remote_write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_remote_write_ops

Number of FC-NVMe remote write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf_remote.write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_remote_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_total_data

Amount of FC-NVMe traffic to and from the storage system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_total_ops

Total number of FC-NVMe operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_write_data

Amount of data written to the storage system (FC-NVMe)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_nvmf_write_ops

Number of FC-NVMe write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `nvmf.write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `nvmf_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/fcp.yaml | 


### fcp_other_ops

Number of operations that are not read or write.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_prim_seq_err

Number of primitive sequence errors

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `primitive_seq_err`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `prim_seq_err`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_queue_full

Number of times a queue full condition occurred.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `queue_full`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `queue_full`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_read_data

Amount of data read from the storage system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_read_ops

Number of read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_reset_count

Number of physical port resets

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `reset_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `reset_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_shared_int_count

Number of shared interrupts

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `shared_interrupt_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `shared_int_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_spurious_int_count

Number of spurious interrupts

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `spurious_interrupt_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `spurious_int_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_threshold_full

Number of times the total number of outstanding commands on the port exceeds the threshold supported by this port.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `threshold_full`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `threshold_full`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_total_data

Amount of FCP traffic to and from the storage system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_total_ops

Total number of FCP operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_write_data

Amount of data written to the storage system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcp_write_ops

Number of write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcp` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcp.yaml | 
| ZAPI | `perf-object-get-instances fcp_port` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcp.yaml | 


### fcvi_rdma_write_avg_latency

Average RDMA write I/O latency.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcvi` | `rdma.write_average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> rdma.write_ops | conf/restperf/9.12.0/fcvi.yaml | 
| ZAPI | `perf-object-get-instances fcvi` | `rdma_write_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> rdma_write_ops | conf/zapiperf/cdot/9.8.0/fcvi.yaml | 


### fcvi_rdma_write_ops

Number of RDMA write I/Os issued per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcvi` | `rdma.write_ops`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcvi.yaml | 
| ZAPI | `perf-object-get-instances fcvi` | `rdma_write_ops`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcvi.yaml | 


### fcvi_rdma_write_throughput

RDMA write throughput in bytes per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/fcvi` | `rdma.write_throughput`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/fcvi.yaml | 
| ZAPI | `perf-object-get-instances fcvi` | `rdma_write_throughput`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/fcvi.yaml | 


### flashcache_accesses

External cache accesses per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `accesses`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `accesses`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_disk_reads_replaced

Estimated number of disk reads per second replaced by cache

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `disk_reads_replaced`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `disk_reads_replaced`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_evicts

Number of blocks evicted from the external cache to make room for new blocks

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `evicts`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `evicts`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_hit

Number of WAFL buffers served off the external cache

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `hit.total`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `hit`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_hit_directory

Number of directory buffers served off the external cache

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `hit.directory`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `hit_directory`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_hit_indirect

Number of indirect file buffers served off the external cache

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `hit.indirect`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `hit_indirect`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_hit_metadata_file

Number of metadata file buffers served off the external cache

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `hit.metadata_file`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `hit_metadata_file`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_hit_normal_lev0

Number of normal level 0 WAFL buffers served off the external cache

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `hit.normal_level_zero`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `hit_normal_lev0`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_hit_percent

External cache hit rate

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `hit.percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> average<br><span class="key">Base:</span> accesses | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `hit_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> accesses | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_inserts

Number of WAFL buffers inserted into the external cache

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `inserts`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `inserts`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_invalidates

Number of blocks invalidated in the external cache

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `invalidates`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `invalidates`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_miss

External cache misses

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `miss.total`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `miss`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_miss_directory

External cache misses accessing directory buffers

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `miss.directory`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `miss_directory`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_miss_indirect

External cache misses accessing indirect file buffers

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `miss.indirect`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `miss_indirect`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_miss_metadata_file

External cache misses accessing metadata file buffers

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `miss.metadata_file`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `miss_metadata_file`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_miss_normal_lev0

External cache misses accessing normal level 0 buffers

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `miss.normal_level_zero`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `miss_normal_lev0`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashcache_usage

Percentage of blocks in external cache currently containing valid data

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/external_cache` | `usage`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/ext_cache_obj.yaml | 
| ZAPI | `perf-object-get-instances ext_cache_obj` | `usage`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml | 


### flashpool_cache_stats

Automated Working-set Analyzer (AWA) per-interval pseudo cache statistics for the most recent intervals. The number of intervals defined as recent is CM_WAFL_HYAS_INT_DIS_CNT. This array is a table with fields corresponding to the enum type of hyas_cache_stat_type_t.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_sizer` | `cache_stats`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_sizer.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_sizer` | `cache_stats`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_sizer.yaml | 


### flashpool_evict_destage_rate

Number of block destage per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `evict_destage_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `evict_destage_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_evict_remove_rate

Number of block free per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `evict_remove_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `evict_remove_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_hya_read_hit_latency_average

Average of RAID I/O latency on read hit.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `hya_read_hit_latency_average`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> hya_read_hit_latency_count | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `hya_read_hit_latency_average`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> hya_read_hit_latency_count | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_hya_read_miss_latency_average

Average read miss latency.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `hya_read_miss_latency_average`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> hya_read_miss_latency_count | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `hya_read_miss_latency_average`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> hya_read_miss_latency_count | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_hya_write_hdd_latency_average

Average write latency to HDD.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `hya_write_hdd_latency_average`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> hya_write_hdd_latency_count | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `hya_write_hdd_latency_average`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> hya_write_hdd_latency_count | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_hya_write_ssd_latency_average

Average of RAID I/O latency on write to SSD.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `hya_write_ssd_latency_average`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> hya_write_ssd_latency_count | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `hya_write_ssd_latency_average`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> hya_write_ssd_latency_count | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_read_cache_ins_rate

Cache insert rate blocks/sec.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `read_cache_insert_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `read_cache_ins_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_read_ops_replaced

Number of HDD read operations replaced by SSD reads per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `read_ops_replaced`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `read_ops_replaced`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_read_ops_replaced_percent

Percentage of HDD read operations replace by SSD.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `read_ops_replaced_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_ops_total | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `read_ops_replaced_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_ops_total | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_ssd_available

Total SSD blocks available.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `ssd_available`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `ssd_available`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_ssd_read_cached

Total read cached SSD blocks.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `ssd_read_cached`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `ssd_read_cached`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_ssd_total

Total SSD blocks.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `ssd_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `ssd_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_ssd_total_used

Total SSD blocks used.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `ssd_total_used`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `ssd_total_used`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_ssd_write_cached

Total write cached SSD blocks.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `ssd_write_cached`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `ssd_write_cached`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_wc_write_blks_total

Number of write-cache blocks written per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `wc_write_blocks_total`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `wc_write_blks_total`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_write_blks_replaced

Number of HDD write blocks replaced by SSD writes per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `write_blocks_replaced`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `write_blks_replaced`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### flashpool_write_blks_replaced_percent

Percentage of blocks overwritten to write-cache among all disk writes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl_hya_per_aggregate` | `write_blocks_replaced_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> average<br><span class="key">Base:</span> estimated_write_blocks_total | conf/restperf/9.12.0/wafl_hya_per_aggr.yaml | 
| ZAPI | `perf-object-get-instances wafl_hya_per_aggr` | `write_blks_replaced_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> average<br><span class="key">Base:</span> est_write_blks_total | conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml | 


### headroom_aggr_current_latency

This is the storage aggregate average latency per message at the disk level.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_aggregate` | `current_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> current_ops | conf/restperf/9.12.0/resource_headroom_aggr.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_aggr` | `current_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> current_ops | conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml | 


### headroom_aggr_current_ops

Total number of I/Os processed by the aggregate per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_aggregate` | `current_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/resource_headroom_aggr.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_aggr` | `current_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml | 


### headroom_aggr_current_utilization

This is the storage aggregate average utilization of all the data disks in the aggregate.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_aggregate` | `current_utilization`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> current_utilization_denominator | conf/restperf/9.12.0/resource_headroom_aggr.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_aggr` | `current_utilization`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> current_utilization_total | conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml | 


### headroom_aggr_ewma_daily

Daily exponential weighted moving average.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_aggregate` | `ewma.daily`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/resource_headroom_aggr.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_aggr` | `ewma_daily`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml | 


### headroom_aggr_ewma_hourly

Hourly exponential weighted moving average.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_aggregate` | `ewma.hourly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/resource_headroom_aggr.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_aggr` | `ewma_hourly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml | 


### headroom_aggr_ewma_monthly

Monthly exponential weighted moving average.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_aggregate` | `ewma.monthly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/resource_headroom_aggr.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_aggr` | `ewma_monthly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml | 


### headroom_aggr_ewma_weekly

Weekly exponential weighted moving average.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_aggregate` | `ewma.weekly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/resource_headroom_aggr.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_aggr` | `ewma_weekly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml | 


### headroom_aggr_optimal_point_confidence_factor

The confidence factor for the optimal point value based on the observed resource latency and utilization.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_aggregate` | `optimal_point.confidence_factor`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point.samples | conf/restperf/9.12.0/resource_headroom_aggr.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_aggr` | `optimal_point_confidence_factor`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point_samples | conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml | 


### headroom_aggr_optimal_point_latency

The latency component of the optimal point of the latency/utilization curve.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_aggregate` | `optimal_point.latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point.samples | conf/restperf/9.12.0/resource_headroom_aggr.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_aggr` | `optimal_point_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point_samples | conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml | 


### headroom_aggr_optimal_point_ops

The ops component of the optimal point derived from the latency/utilzation curve.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_aggregate` | `optimal_point.ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point.samples | conf/restperf/9.12.0/resource_headroom_aggr.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_aggr` | `optimal_point_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point_samples | conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml | 


### headroom_aggr_optimal_point_utilization

The utilization component of the optimal point of the latency/utilization curve.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_aggregate` | `optimal_point.utilization`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point.samples | conf/restperf/9.12.0/resource_headroom_aggr.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_aggr` | `optimal_point_utilization`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point_samples | conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml | 


### headroom_cpu_current_latency

Current operation latency of the resource.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_cpu` | `current_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> current_ops | conf/restperf/9.12.0/resource_headroom_cpu.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_cpu` | `current_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> current_ops | conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml | 


### headroom_cpu_current_ops

Total number of operations per second (also referred to as dblade ops).

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_cpu` | `current_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/resource_headroom_cpu.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_cpu` | `current_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml | 


### headroom_cpu_current_utilization

Average processor utilization across all processors in the system.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_cpu` | `current_utilization`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> elapsed_time | conf/restperf/9.12.0/resource_headroom_cpu.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_cpu` | `current_utilization`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> current_utilization_total | conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml | 


### headroom_cpu_ewma_daily

Daily exponential weighted moving average for current_ops, optimal_point_ops, current_latency, optimal_point_latency, current_utilization, optimal_point_utilization and optimal_point_confidence_factor.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_cpu` | `ewma.daily`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/resource_headroom_cpu.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_cpu` | `ewma_daily`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml | 


### headroom_cpu_ewma_hourly

Hourly exponential weighted moving average for current_ops, optimal_point_ops, current_latency, optimal_point_latency, current_utilization, optimal_point_utilization and optimal_point_confidence_factor.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_cpu` | `ewma.hourly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/resource_headroom_cpu.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_cpu` | `ewma_hourly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml | 


### headroom_cpu_ewma_monthly

Monthly exponential weighted moving average for current_ops, optimal_point_ops, current_latency, optimal_point_latency, current_utilization, optimal_point_utilization and optimal_point_confidence_factor.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_cpu` | `ewma.monthly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/resource_headroom_cpu.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_cpu` | `ewma_monthly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml | 


### headroom_cpu_ewma_weekly

Weekly exponential weighted moving average for current_ops, optimal_point_ops, current_latency, optimal_point_latency, current_utilization, optimal_point_utilization and optimal_point_confidence_factor.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_cpu` | `ewma.weekly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/resource_headroom_cpu.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_cpu` | `ewma_weekly`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml | 


### headroom_cpu_optimal_point_confidence_factor

Confidence factor for the optimal point value based on the observed resource latency and utilization. The possible values are: 0 - unknown, 1 - low, 2 - medium, 3 - high. This counter can provide an average confidence factor over a range of time.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_cpu` | `optimal_point.confidence_factor`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point.samples | conf/restperf/9.12.0/resource_headroom_cpu.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_cpu` | `optimal_point_confidence_factor`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point_samples | conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml | 


### headroom_cpu_optimal_point_latency

Latency component of the optimal point of the latency/utilization curve. This counter can provide an average latency over a range of time.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_cpu` | `optimal_point.latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point.samples | conf/restperf/9.12.0/resource_headroom_cpu.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_cpu` | `optimal_point_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point_samples | conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml | 


### headroom_cpu_optimal_point_ops

Ops component of the optimal point derived from the latency/utilization curve. This counter can provide an average ops over a range of time.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_cpu` | `optimal_point.ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point.samples | conf/restperf/9.12.0/resource_headroom_cpu.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_cpu` | `optimal_point_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point_samples | conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml | 


### headroom_cpu_optimal_point_utilization

Utilization component of the optimal point of the latency/utilization curve. This counter can provide an average utilization over a range of time.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `/api/cluster/counter/tables/headroom_cpu` | `optimal_point.utilization`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point.samples | conf/restperf/9.12.0/resource_headroom_cpu.yaml | 
| ZAPI | `perf-object-get-instances resource_headroom_cpu` | `optimal_point_utilization`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> average<br><span class="key">Base:</span> optimal_point_samples | conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml | 


### hostadapter_bytes_read

Bytes read through a host adapter

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/host_adapter` | `bytes_read`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/hostadapter.yaml | 
| ZAPI | `perf-object-get-instances hostadapter` | `bytes_read`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/hostadapter.yaml | 


### hostadapter_bytes_written

Bytes written through a host adapter

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/host_adapter` | `bytes_written`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/hostadapter.yaml | 
| ZAPI | `perf-object-get-instances hostadapter` | `bytes_written`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/hostadapter.yaml | 


### iscsi_lif_avg_latency

Average latency for iSCSI operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/iscsi_lif` | `average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cmd_transferred | conf/restperf/9.12.0/iscsi_lif.yaml | 
| ZAPI | `perf-object-get-instances iscsi_lif` | `avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cmd_transfered | conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml | 


### iscsi_lif_avg_other_latency

Average latency for operations other than read and write (for example, Inquiry, Report LUNs, SCSI Task Management Functions)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/iscsi_lif` | `average_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi_other_ops | conf/restperf/9.12.0/iscsi_lif.yaml | 
| ZAPI | `perf-object-get-instances iscsi_lif` | `avg_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi_other_ops | conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml | 


### iscsi_lif_avg_read_latency

Average latency for read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/iscsi_lif` | `average_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi_read_ops | conf/restperf/9.12.0/iscsi_lif.yaml | 
| ZAPI | `perf-object-get-instances iscsi_lif` | `avg_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi_read_ops | conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml | 


### iscsi_lif_avg_write_latency

Average latency for write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/iscsi_lif` | `average_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi_write_ops | conf/restperf/9.12.0/iscsi_lif.yaml | 
| ZAPI | `perf-object-get-instances iscsi_lif` | `avg_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi_write_ops | conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml | 


### iscsi_lif_cmd_transfered

Command transfered by this iSCSI conn

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances iscsi_lif` | `cmd_transfered`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml | 


### iscsi_lif_cmd_transferred

Command transferred by this iSCSI connection

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/iscsi_lif` | `cmd_transferred`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/iscsi_lif.yaml | 


### iscsi_lif_iscsi_other_ops

iSCSI other operations per second on this logical interface (LIF)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/iscsi_lif` | `iscsi_other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/iscsi_lif.yaml | 
| ZAPI | `perf-object-get-instances iscsi_lif` | `iscsi_other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml | 


### iscsi_lif_iscsi_read_ops

iSCSI read operations per second on this logical interface (LIF)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/iscsi_lif` | `iscsi_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/iscsi_lif.yaml | 
| ZAPI | `perf-object-get-instances iscsi_lif` | `iscsi_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml | 


### iscsi_lif_iscsi_write_ops

iSCSI write operations per second on this logical interface (LIF)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/iscsi_lif` | `iscsi_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/iscsi_lif.yaml | 
| ZAPI | `perf-object-get-instances iscsi_lif` | `iscsi_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml | 


### iscsi_lif_protocol_errors

Number of protocol errors from iSCSI sessions on this logical interface (LIF)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/iscsi_lif` | `protocol_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/iscsi_lif.yaml | 
| ZAPI | `perf-object-get-instances iscsi_lif` | `protocol_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml | 


### iscsi_lif_read_data

Amount of data read from the storage system in bytes

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/iscsi_lif` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/iscsi_lif.yaml | 
| ZAPI | `perf-object-get-instances iscsi_lif` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml | 


### iscsi_lif_write_data

Amount of data written to the storage system in bytes

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/iscsi_lif` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/iscsi_lif.yaml | 
| ZAPI | `perf-object-get-instances iscsi_lif` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml | 


### lif_recv_data

Number of bytes received per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lif` | `received_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lif.yaml | 
| ZAPI | `perf-object-get-instances lif` | `recv_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lif.yaml | 


### lif_recv_errors

Number of received Errors per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lif` | `received_errors`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lif.yaml | 
| ZAPI | `perf-object-get-instances lif` | `recv_errors`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lif.yaml | 


### lif_recv_packet

Number of packets received per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lif` | `received_packets`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lif.yaml | 
| ZAPI | `perf-object-get-instances lif` | `recv_packet`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lif.yaml | 


### lif_sent_data

Number of bytes sent per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lif` | `sent_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lif.yaml | 
| ZAPI | `perf-object-get-instances lif` | `sent_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lif.yaml | 


### lif_sent_errors

Number of sent errors per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lif` | `sent_errors`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lif.yaml | 
| ZAPI | `perf-object-get-instances lif` | `sent_errors`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lif.yaml | 


### lif_sent_packet

Number of packets sent per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lif` | `sent_packets`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lif.yaml | 
| ZAPI | `perf-object-get-instances lif` | `sent_packet`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lif.yaml | 


### lun_avg_read_latency

Average read latency in microseconds for all operations on the LUN

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `average_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `avg_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_avg_write_latency

Average write latency in microseconds for all operations on the LUN

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `average_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `avg_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_avg_xcopy_latency

Average latency in microseconds for xcopy requests

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `average_xcopy_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> xcopy_requests | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `avg_xcopy_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> xcopy_reqs | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_caw_reqs

Number of compare and write requests

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `caw_requests`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `caw_reqs`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_enospc

Number of operations receiving ENOSPC errors

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `enospc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `enospc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_queue_full

Queue full responses

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `queue_full`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `queue_full`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_read_align_histo

Histogram of WAFL read alignment (number sectors off WAFL block start)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `read_align_histogram`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_ops_sent | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `read_align_histo`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_ops_sent | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_read_data

Read bytes

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_read_ops

Number of read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_read_partial_blocks

Percentage of reads whose size is not a multiple of WAFL block size

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `read_partial_blocks`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_ops | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `read_partial_blocks`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_remote_bytes

I/O to or from a LUN which is not owned by the storage system handling the I/O.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `remote_bytes`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `remote_bytes`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_remote_ops

Number of operations received by a storage system that does not own the LUN targeted by the operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `remote_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `remote_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_size

The total provisioned size of the LUN. The LUN size can be increased but not be made smaller using the REST interface.<br/>The maximum and minimum sizes listed here are the absolute maximum and absolute minimum sizes in bytes. The actual minimum and maxiumum sizes vary depending on the ONTAP version, ONTAP platform and the available space in the containing volume and aggregate.<br/>For more information, see _Size properties_ in the _docs_ section of the ONTAP REST API documentation.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/luns` | `space.size` | conf/rest/9.12.0/lun.yaml |
| ZAPI | `lun-get-iter` | `lun-info.size` | conf/zapi/cdot/9.8.0/lun.yaml |


### lun_size_used

The amount of space consumed by the main data stream of the LUN.<br/>This value is the total space consumed in the volume by the LUN, including filesystem overhead, but excluding prefix and suffix streams. Due to internal filesystem overhead and the many ways SAN filesystems and applications utilize blocks within a LUN, this value does not necessarily reflect actual consumption/availability from the perspective of the filesystem or application. Without specific knowledge of how the LUN blocks are utilized outside of ONTAP, this property should not be used as an indicator for an out-of-space condition.<br/>For more information, see _Size properties_ in the _docs_ section of the ONTAP REST API documentation.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/luns` | `space.used` | conf/rest/9.12.0/lun.yaml |
| ZAPI | `lun-get-iter` | `lun-info.size-used` | conf/zapi/cdot/9.8.0/lun.yaml |


### lun_unmap_reqs

Number of unmap command requests

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `unmap_requests`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `unmap_reqs`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_write_align_histo

Histogram of WAFL write alignment (number of sectors off WAFL block start)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `write_align_histogram`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> write_ops_sent | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `write_align_histo`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> write_ops_sent | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_write_data

Write bytes

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_write_ops

Number of write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_write_partial_blocks

Percentage of writes whose size is not a multiple of WAFL block size

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `write_partial_blocks`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> write_ops | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `write_partial_blocks`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> write_ops | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_writesame_reqs

Number of write same command requests

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `writesame_requests`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `writesame_reqs`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_writesame_unmap_reqs

Number of write same commands requests with unmap bit set

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `writesame_unmap_requests`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `writesame_unmap_reqs`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### lun_xcopy_reqs

Total number of xcopy operations on the LUN

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/lun` | `xcopy_requests`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/lun.yaml | 
| ZAPI | `perf-object-get-instances lun` | `xcopy_reqs`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/lun.yaml | 


### metadata_collector_api_time

amount of time to collect data from monitored cluster object

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_collector_instances

number of objects collected from monitored cluster

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_collector_metrics

number of counters collected from monitored cluster

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_collector_parse_time

amount of time to parse XML, JSON, etc. for cluster object

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_collector_plugin_time

amount of time for all plugins to post-process metrics

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_collector_poll_time

amount of time it took for the poll to finish

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_collector_task_time

amount of time it took for each collector's subtasks to complete

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_component_count

number of metrics collected for each object

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_component_status

status of the collector - 0 means running, 1 means standby, 2 means failed

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_exporter_count

number of metrics and labels exported

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_exporter_time

amount of time it took to render, export, and serve exported data

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_target_goroutines

number of goroutines that exist within the poller

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### metadata_target_status

status of the system being monitored. 0 means reachable, 1 means unreachable

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | NA |
| ZAPI | `NA` | `Harvest generated` | NA |


### namespace_avg_other_latency

Average other ops latency in microseconds for all operations on the Namespace

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/namespace` | `average_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> other_ops | conf/restperf/9.12.0/namespace.yaml | 
| ZAPI | `perf-object-get-instances namespace` | `avg_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> other_ops | conf/zapiperf/cdot/9.10.1/namespace.yaml | 


### namespace_avg_read_latency

Average read latency in microseconds for all operations on the Namespace

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/namespace` | `average_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/restperf/9.12.0/namespace.yaml | 
| ZAPI | `perf-object-get-instances namespace` | `avg_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.10.1/namespace.yaml | 


### namespace_avg_write_latency

Average write latency in microseconds for all operations on the Namespace

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/namespace` | `average_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/restperf/9.12.0/namespace.yaml | 
| ZAPI | `perf-object-get-instances namespace` | `avg_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/zapiperf/cdot/9.10.1/namespace.yaml | 


### namespace_block_size

The size of blocks in the namespace in bytes.<br/>Valid in POST when creating an NVMe namespace that is not a clone of another. Disallowed in POST when creating a namespace clone. Valid in POST.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/namespaces` | `space.block_size` | conf/rest/9.12.0/namespace.yaml |
| ZAPI | `nvme-namespace-get-iter` | `nvme-namespace-info.block-size` | conf/zapi/cdot/9.8.0/namespace.yaml |


### namespace_other_ops

Number of other operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/namespace` | `other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/namespace.yaml | 
| ZAPI | `perf-object-get-instances namespace` | `other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/namespace.yaml | 


### namespace_read_data

Read bytes

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/namespace` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/namespace.yaml | 
| ZAPI | `perf-object-get-instances namespace` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/namespace.yaml | 


### namespace_read_ops

Number of read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/namespace` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/namespace.yaml | 
| ZAPI | `perf-object-get-instances namespace` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/namespace.yaml | 


### namespace_remote_bytes

Remote read bytes

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/namespace` | `remote.read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/namespace.yaml | 
| ZAPI | `perf-object-get-instances namespace` | `remote_bytes`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/namespace.yaml | 


### namespace_remote_ops

Number of remote read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/namespace` | `remote.read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/namespace.yaml | 
| ZAPI | `perf-object-get-instances namespace` | `remote_ops`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/namespace.yaml | 


### namespace_size

The total provisioned size of the NVMe namespace. Valid in POST and PATCH. The NVMe namespace size can be increased but not be made smaller using the REST interface.<br/>The maximum and minimum sizes listed here are the absolute maximum and absolute minimum sizes in bytes. The maximum size is variable with respect to large NVMe namespace support in ONTAP. If large namespaces are supported, the maximum size is 128 TB (140737488355328 bytes) and if not supported, the maximum size is just under 16 TB (17557557870592 bytes). The minimum size supported is always 4096 bytes.<br/>For more information, see _Size properties_ in the _docs_ section of the ONTAP REST API documentation.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/namespaces` | `space.size` | conf/rest/9.12.0/namespace.yaml |
| ZAPI | `nvme-namespace-get-iter` | `nvme-namespace-info.size` | conf/zapi/cdot/9.8.0/namespace.yaml |


### namespace_size_used

The amount of space consumed by the main data stream of the NVMe namespace.<br/>This value is the total space consumed in the volume by the NVMe namespace, including filesystem overhead, but excluding prefix and suffix streams. Due to internal filesystem overhead and the many ways NVMe filesystems and applications utilize blocks within a namespace, this value does not necessarily reflect actual consumption/availability from the perspective of the filesystem or application. Without specific knowledge of how the namespace blocks are utilized outside of ONTAP, this property should not be used and an indicator for an out-of-space condition.<br/>For more information, see _Size properties_ in the _docs_ section of the ONTAP REST API documentation.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/namespaces` | `space.used` | conf/rest/9.12.0/namespace.yaml |
| ZAPI | `nvme-namespace-get-iter` | `nvme-namespace-info.size-used` | conf/zapi/cdot/9.8.0/namespace.yaml |


### namespace_write_data

Write bytes

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/namespace` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/namespace.yaml | 
| ZAPI | `perf-object-get-instances namespace` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/namespace.yaml | 


### namespace_write_ops

Number of write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/namespace` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/namespace.yaml | 
| ZAPI | `perf-object-get-instances namespace` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/namespace.yaml | 


### net_port_mtu

Maximum transmission unit, largest packet size on this network

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/network/ethernet/ports` | `mtu` | conf/rest/9.12.0/netport.yaml |
| ZAPI | `net-port-get-iter` | `net-port-info.mtu` | conf/zapi/cdot/9.8.0/netport.yaml |


### netstat_bytes_recvd

Number of bytes received by a TCP connection

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances netstat` | `bytes_recvd`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/netstat.yaml | 


### netstat_bytes_sent

Number of bytes sent by a TCP connection

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances netstat` | `bytes_sent`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/netstat.yaml | 


### netstat_cong_win

Congestion window of a TCP connection

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances netstat` | `cong_win`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/netstat.yaml | 


### netstat_cong_win_th

Congestion window threshold of a TCP connection

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances netstat` | `cong_win_th`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/netstat.yaml | 


### netstat_ooorcv_pkts

Number of out-of-order packets received by this TCP connection

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances netstat` | `ooorcv_pkts`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/netstat.yaml | 


### netstat_recv_window

Receive window size of a TCP connection

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances netstat` | `recv_window`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/netstat.yaml | 


### netstat_rexmit_pkts

Number of packets retransmitted by this TCP connection

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances netstat` | `rexmit_pkts`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/netstat.yaml | 


### netstat_send_window

Send window size of a TCP connection

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances netstat` | `send_window`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/netstat.yaml | 


### nfs_clients_idle_duration

Specifies an ISO-8601 format of date and time to retrieve the idle time duration in hours, minutes, and seconds format.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/protocols/nfs/connected-clients` | `idle_duration` | conf/rest/9.7.0/nfs_clients.yaml |


### nfs_diag_storePool_ByteLockAlloc

Current number of byte range lock objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.byte_lock_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_ByteLockAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_ByteLockMax

Maximum number of byte range lock objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.byte_lock_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_ByteLockMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_ClientAlloc

Current number of client objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.client_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_ClientAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_ClientMax

Maximum number of client objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.client_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_ClientMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_ConnectionParentSessionReferenceAlloc

Current number of connection parent session reference objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.connection_parent_session_reference_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_ConnectionParentSessionReferenceAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_ConnectionParentSessionReferenceMax

Maximum number of connection parent session reference objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.connection_parent_session_reference_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_ConnectionParentSessionReferenceMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_CopyStateAlloc

Current number of copy state objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.copy_state_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_CopyStateAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_CopyStateMax

Maximum number of copy state objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.copy_state_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_CopyStateMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_DelegAlloc

Current number of delegation lock objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.delegation_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_DelegAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_DelegMax

Maximum number delegation lock objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.delegation_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_DelegMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_DelegStateAlloc

Current number of delegation state objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.delegation_state_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_DelegStateAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_DelegStateMax

Maximum number of delegation state objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.delegation_state_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_DelegStateMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_LayoutAlloc

Current number of layout objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.layout_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_LayoutAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_LayoutMax

Maximum number of layout objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.layout_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_LayoutMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_LayoutStateAlloc

Current number of layout state objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.layout_state_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_LayoutStateAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_LayoutStateMax

Maximum number of layout state objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.layout_state_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_LayoutStateMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_LockStateAlloc

Current number of lock state objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.lock_state_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_LockStateAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_LockStateMax

Maximum number of lock state objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.lock_state_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_LockStateMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_OpenAlloc

Current number of share objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.open_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_OpenAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_OpenMax

Maximum number of share lock objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.open_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_OpenMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_OpenStateAlloc

Current number of open state objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.openstate_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_OpenStateAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_OpenStateMax

Maximum number of open state objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.openstate_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_OpenStateMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_OwnerAlloc

Current number of owner objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.owner_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_OwnerAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_OwnerMax

Maximum number of owner objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.owner_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_OwnerMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_SessionAlloc

Current number of session objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.session_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_SessionAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_SessionConnectionHolderAlloc

Current number of session connection holder objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.session_connection_holder_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_SessionConnectionHolderAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_SessionConnectionHolderMax

Maximum number of session connection holder objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.session_connection_holder_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_SessionConnectionHolderMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_SessionHolderAlloc

Current number of session holder objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.session_holder_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_SessionHolderAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_SessionHolderMax

Maximum number of session holder objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.session_holder_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_SessionHolderMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_SessionMax

Maximum number of session objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.session_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_SessionMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_StateRefHistoryAlloc

Current number of state reference callstack history objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.state_reference_history_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_StateRefHistoryAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_StateRefHistoryMax

Maximum number of state reference callstack history objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.state_reference_history_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_StateRefHistoryMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_StringAlloc

Current number of string objects allocated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.string_allocated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_StringAlloc`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nfs_diag_storePool_StringMax

Maximum number of string objects.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nfs_v4_diag` | `storepool.string_maximum`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_pool.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_diag` | `storePool_StringMax`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml | 


### nic_link_up_to_downs

Number of link state change from UP to DOWN.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nic_common` | `link_up_to_down`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nic_common.yaml | 
| ZAPI | `perf-object-get-instances nic_common` | `link_up_to_downs`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nic_common.yaml | 


### nic_rx_alignment_errors

Alignment errors detected on received packets

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nic_common` | `receive_alignment_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nic_common.yaml | 
| ZAPI | `perf-object-get-instances nic_common` | `rx_alignment_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nic_common.yaml | 


### nic_rx_bytes

Bytes received

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nic_common` | `receive_bytes`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nic_common.yaml | 
| ZAPI | `perf-object-get-instances nic_common` | `rx_bytes`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nic_common.yaml | 


### nic_rx_crc_errors

CRC errors detected on received packets

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nic_common` | `receive_crc_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nic_common.yaml | 
| ZAPI | `perf-object-get-instances nic_common` | `rx_crc_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nic_common.yaml | 


### nic_rx_errors

Error received

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nic_common` | `receive_errors`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nic_common.yaml | 
| ZAPI | `perf-object-get-instances nic_common` | `rx_errors`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nic_common.yaml | 


### nic_rx_length_errors

Length errors detected on received packets

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nic_common` | `receive_length_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nic_common.yaml | 
| ZAPI | `perf-object-get-instances nic_common` | `rx_length_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nic_common.yaml | 


### nic_rx_total_errors

Total errors received

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nic_common` | `receive_total_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nic_common.yaml | 
| ZAPI | `perf-object-get-instances nic_common` | `rx_total_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nic_common.yaml | 


### nic_tx_bytes

Bytes sent

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nic_common` | `transmit_bytes`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nic_common.yaml | 
| ZAPI | `perf-object-get-instances nic_common` | `tx_bytes`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nic_common.yaml | 


### nic_tx_errors

Error sent

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nic_common` | `transmit_errors`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nic_common.yaml | 
| ZAPI | `perf-object-get-instances nic_common` | `tx_errors`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nic_common.yaml | 


### nic_tx_hw_errors

Transmit errors reported by hardware

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nic_common` | `transmit_hw_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nic_common.yaml | 
| ZAPI | `perf-object-get-instances nic_common` | `tx_hw_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nic_common.yaml | 


### nic_tx_total_errors

Total errors sent

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nic_common` | `transmit_total_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nic_common.yaml | 
| ZAPI | `perf-object-get-instances nic_common` | `tx_total_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nic_common.yaml | 


### node_avg_processor_busy

Average processor utilization across all processors in the system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `average_processor_busy_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> cpu_elapsed_time | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `avg_processor_busy`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> cpu_elapsed_time | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_cifs_connections

Number of connections

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs:node` | `connections`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_node.yaml | 
| ZAPI | `perf-object-get-instances cifs:node` | `connections`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_node.yaml | 


### node_cifs_established_sessions

Number of established SMB and SMB2 sessions

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs:node` | `established_sessions`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_node.yaml | 
| ZAPI | `perf-object-get-instances cifs:node` | `established_sessions`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_node.yaml | 


### node_cifs_latency

Average latency for CIFS operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs:node` | `latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> latency_base | conf/restperf/9.12.0/cifs_node.yaml | 
| ZAPI | `perf-object-get-instances cifs:node` | `cifs_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs_latency_base | conf/zapiperf/cdot/9.8.0/cifs_node.yaml | 


### node_cifs_op_count

Array of select CIFS operation counts

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs:node` | `op_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_node.yaml | 
| ZAPI | `perf-object-get-instances cifs:node` | `cifs_op_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_node.yaml | 


### node_cifs_open_files

Number of open files over SMB and SMB2

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs:node` | `open_files`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_node.yaml | 
| ZAPI | `perf-object-get-instances cifs:node` | `open_files`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_node.yaml | 


### node_cifs_ops

Number of CIFS operations per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `cifs_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `cifs_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_cifs_read_latency

Average latency for CIFS read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs:node` | `average_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_read_ops | conf/restperf/9.12.0/cifs_node.yaml | 
| ZAPI | `perf-object-get-instances cifs:node` | `cifs_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs_read_ops | conf/zapiperf/cdot/9.8.0/cifs_node.yaml | 


### node_cifs_read_ops

Total number of CIFS read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs:node` | `total_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_node.yaml | 
| ZAPI | `perf-object-get-instances cifs:node` | `cifs_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_node.yaml | 


### node_cifs_total_ops

Total number of CIFS operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs:node` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_node.yaml | 
| ZAPI | `perf-object-get-instances cifs:node` | `cifs_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_node.yaml | 


### node_cifs_write_latency

Average latency for CIFS write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs:node` | `average_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_write_ops | conf/restperf/9.12.0/cifs_node.yaml | 
| ZAPI | `perf-object-get-instances cifs:node` | `cifs_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs_write_ops | conf/zapiperf/cdot/9.8.0/cifs_node.yaml | 


### node_cifs_write_ops

Total number of CIFS write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs:node` | `total_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_node.yaml | 
| ZAPI | `perf-object-get-instances cifs:node` | `cifs_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_node.yaml | 


### node_cpu_busy

System CPU resource utilization. Returns a computed percentage for the default CPU field. Basically computes a 'cpu usage summary' value which indicates how 'busy' the system is based upon the most heavily utilized domain. The idea is to determine the amount of available CPU until we're limited by either a domain maxing out OR we exhaust all available idle CPU cycles, whichever occurs first.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `cpu_busy`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> cpu_elapsed_time | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `cpu_busy`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> cpu_elapsed_time | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_cpu_busytime

The time (in hundredths of a second) that the CPU has been doing useful work since the last boot

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `system-node-get-iter` | `node-details-info.cpu-busytime` | conf/zapi/cdot/9.8.0/node.yaml |
| REST | `api/private/cli/node` | `cpu_busy_time` | conf/rest/9.12.0/node.yaml |


### node_cpu_domain_busy

Array of processor time in percentage spent in various domains

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `domain_busy`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> cpu_elapsed_time | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `domain_busy`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> cpu_elapsed_time | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_cpu_elapsed_time

Elapsed time since boot

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `cpu_elapsed_time`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `cpu_elapsed_time`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-display<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_disk_data_read

Number of disk kilobytes (KB) read per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `disk_data_read`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `disk_data_read`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_disk_data_written

Number of disk kilobytes (KB) written per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `disk_data_written`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `disk_data_written`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_failed_fan

Specifies a count of the number of chassis fans that are not operating within the recommended RPM range.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/nodes` | `controller.failed_fan.count` | conf/rest/9.12.0/node.yaml |
| ZAPI | `system-node-get-iter` | `node-details-info.env-failed-fan-count` | conf/zapi/cdot/9.8.0/node.yaml |


### node_failed_power

Number of failed power supply units.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/nodes` | `controller.failed_power_supply.count` | conf/rest/9.12.0/node.yaml |
| ZAPI | `system-node-get-iter` | `node-details-info.env-failed-power-supply-count` | conf/zapi/cdot/9.8.0/node.yaml |


### node_fcp_data_recv

Number of FCP kilobytes (KB) received per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `fcp_data_received`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `fcp_data_recv`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_fcp_data_sent

Number of FCP kilobytes (KB) sent per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `fcp_data_sent`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `fcp_data_sent`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_fcp_ops

Number of FCP operations per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `fcp_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `fcp_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_hdd_data_read

Number of HDD Disk kilobytes (KB) read per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `hdd_data_read`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `hdd_data_read`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_hdd_data_written

Number of HDD kilobytes (KB) written per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `hdd_data_written`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `hdd_data_written`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_iscsi_ops

Number of iSCSI operations per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `iscsi_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `iscsi_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_memory

Total memory in megabytes (MB)

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `memory`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `memory`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_net_data_recv

Number of network kilobytes (KB) received per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `network_data_received`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `net_data_recv`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_net_data_sent

Number of network kilobytes (KB) sent per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `network_data_sent`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `net_data_sent`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_nfs_access_avg_latency

Average latency of ACCESS procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `access.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> access.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `access_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> access_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_access_total

Total number of ACCESS procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `access.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `access_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_backchannel_ctl_avg_latency

Average latency of NFSv4.2 BACKCHANNEL_CTL operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `backchannel_ctl.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> backchannel_ctl.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `backchannel_ctl_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> backchannel_ctl_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_backchannel_ctl_total

Total number of NFSv4.2 BACKCHANNEL_CTL operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `backchannel_ctl.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `backchannel_ctl_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_bind_conn_to_session_avg_latency

Average latency of NFSv4.2 BIND_CONN_TO_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `bind_conn_to_session.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> bind_conn_to_session.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `bind_conn_to_session_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> bind_conn_to_session_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_bind_conn_to_session_total

Total number of NFSv4.2 BIND_CONN_TO_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `bind_conn_to_session.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `bind_conn_to_session_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_close_avg_latency

Average latency of CLOSE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `close.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> close.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `close_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> close_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_close_total

Total number of CLOSE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `close.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `close_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_commit_avg_latency

Average latency of COMMIT procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `commit.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> commit.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `commit_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> commit_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_commit_total

Total number of COMMIT procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `commit.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `commit_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_create_avg_latency

Average latency of CREATE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `create.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> create.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `create_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> create_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_create_session_avg_latency

Average latency of NFSv4.2 CREATE_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `create_session.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> create_session.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `create_session_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> create_session_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_create_session_total

Total number of NFSv4.2 CREATE_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `create_session.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `create_session_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_create_total

Total number of CREATE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `create.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `create_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_delegpurge_avg_latency

Average latency of DELEGPURGE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `delegpurge.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> delegpurge.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `delegpurge_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> delegpurge_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_delegpurge_total

Total number of DELEGPURGE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `delegpurge.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `delegpurge_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_delegreturn_avg_latency

Average latency of DELEGRETURN procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `delegreturn.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> delegreturn.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `delegreturn_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> delegreturn_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_delegreturn_total

Total number of DELEGRETURN procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `delegreturn.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `delegreturn_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_destroy_clientid_avg_latency

Average latency of NFSv4.2 DESTROY_CLIENTID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `destroy_clientid.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> destroy_clientid.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `destroy_clientid_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> destroy_clientid_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_destroy_clientid_total

Total number of NFSv4.2 DESTROY_CLIENTID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `destroy_clientid.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `destroy_clientid_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_destroy_session_avg_latency

Average latency of NFSv4.2 DESTROY_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `destroy_session.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> destroy_session.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `destroy_session_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> destroy_session_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_destroy_session_total

Total number of NFSv4.2 DESTROY_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `destroy_session.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `destroy_session_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_exchange_id_avg_latency

Average latency of NFSv4.2 EXCHANGE_ID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `exchange_id.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> exchange_id.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `exchange_id_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> exchange_id_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_exchange_id_total

Total number of NFSv4.2 EXCHANGE_ID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `exchange_id.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `exchange_id_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_free_stateid_avg_latency

Average latency of NFSv4.2 FREE_STATEID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `free_stateid.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> free_stateid.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `free_stateid_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> free_stateid_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_free_stateid_total

Total number of NFSv4.2 FREE_STATEID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `free_stateid.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `free_stateid_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_fsinfo_avg_latency

Average latency of FSInfo procedure requests. The counter keeps track of the average response time of FSInfo requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `fsinfo.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> fsinfo.total | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `fsinfo_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> fsinfo_total | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_fsinfo_total

Total number FSInfo of procedure requests. It is the total number of FSInfo success and FSInfo error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `fsinfo.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `fsinfo_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_fsstat_avg_latency

Average latency of FSStat procedure requests. The counter keeps track of the average response time of FSStat requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `fsstat.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> fsstat.total | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `fsstat_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> fsstat_total | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_fsstat_total

Total number FSStat of procedure requests. It is the total number of FSStat success and FSStat error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `fsstat.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `fsstat_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_get_dir_delegation_avg_latency

Average latency of NFSv4.2 GET_DIR_DELEGATION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `get_dir_delegation.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> get_dir_delegation.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `get_dir_delegation_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> get_dir_delegation_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_get_dir_delegation_total

Total number of NFSv4.2 GET_DIR_DELEGATION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `get_dir_delegation.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `get_dir_delegation_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_getattr_avg_latency

Average latency of GETATTR procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `getattr.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> getattr.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `getattr_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> getattr_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_getattr_total

Total number of GETATTR procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `getattr.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `getattr_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_getdeviceinfo_avg_latency

Average latency of NFSv4.2 GETDEVICEINFO operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `getdeviceinfo.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> getdeviceinfo.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `getdeviceinfo_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> getdeviceinfo_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_getdeviceinfo_total

Total number of NFSv4.2 GETDEVICEINFO operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `getdeviceinfo.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `getdeviceinfo_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_getdevicelist_avg_latency

Average latency of NFSv4.2 GETDEVICELIST operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `getdevicelist.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> getdevicelist.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `getdevicelist_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> getdevicelist_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_getdevicelist_total

Total number of NFSv4.2 GETDEVICELIST operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `getdevicelist.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `getdevicelist_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_getfh_avg_latency

Average latency of GETFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `getfh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> getfh.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `getfh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> getfh_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_getfh_total

Total number of GETFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `getfh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `getfh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_latency

Average latency of NFSv4 requests. This counter keeps track of the average response time of NFSv4 requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> total_ops | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_layoutcommit_avg_latency

Average latency of NFSv4.2 LAYOUTCOMMIT operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `layoutcommit.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> layoutcommit.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `layoutcommit_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> layoutcommit_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_layoutcommit_total

Total number of NFSv4.2 LAYOUTCOMMIT operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `layoutcommit.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `layoutcommit_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_layoutget_avg_latency

Average latency of NFSv4.2 LAYOUTGET operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `layoutget.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> layoutget.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `layoutget_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> layoutget_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_layoutget_total

Total number of NFSv4.2 LAYOUTGET operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `layoutget.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `layoutget_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_layoutreturn_avg_latency

Average latency of NFSv4.2 LAYOUTRETURN operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `layoutreturn.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> layoutreturn.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `layoutreturn_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> layoutreturn_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_layoutreturn_total

Total number of NFSv4.2 LAYOUTRETURN operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `layoutreturn.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `layoutreturn_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_link_avg_latency

Average latency of LINK procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `link.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> link.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `link_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> link_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_link_total

Total number of LINK procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `link.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `link_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_lock_avg_latency

Average latency of LOCK procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `lock.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> lock.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `lock_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> lock_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_lock_total

Total number of LOCK procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `lock.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `lock_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_lockt_avg_latency

Average latency of LOCKT procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `lockt.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> lockt.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `lockt_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> lockt_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_lockt_total

Total number of LOCKT procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `lockt.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `lockt_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_locku_avg_latency

Average latency of LOCKU procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `locku.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> locku.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `locku_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> locku_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_locku_total

Total number of LOCKU procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `locku.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `locku_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_lookup_avg_latency

Average latency of LOOKUP procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `lookup.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> lookup.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `lookup_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> lookup_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_lookup_total

Total number of LOOKUP procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `lookup.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `lookup_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_lookupp_avg_latency

Average latency of LOOKUPP procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `lookupp.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> lookupp.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `lookupp_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> lookupp_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_lookupp_total

Total number of LOOKUPP procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `lookupp.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `lookupp_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_mkdir_avg_latency

Average latency of MkDir procedure requests. The counter keeps track of the average response time of MkDir requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `mkdir.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> mkdir.total | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `mkdir_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> mkdir_total | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_mkdir_total

Total number MkDir of procedure requests. It is the total number of MkDir success and MkDir error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `mkdir.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `mkdir_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_mknod_avg_latency

Average latency of MkNod procedure requests. The counter keeps track of the average response time of MkNod requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `mknod.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> mknod.total | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `mknod_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> mknod_total | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_mknod_total

Total number MkNod of procedure requests. It is the total number of MkNod success and MkNod error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `mknod.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `mknod_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_null_avg_latency

Average Latency of NULL procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `null.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> null.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `null_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> null_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_null_total

Total number of NULL procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `null.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `null_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_nverify_avg_latency

Average latency of NVERIFY procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `nverify.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nverify.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `nverify_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> nverify_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_nverify_total

Total number of NVERIFY procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `nverify.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `nverify_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_open_avg_latency

Average latency of OPEN procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `open.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> open.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `open_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> open_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_open_confirm_avg_latency

Average latency of OPEN_CONFIRM procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `open_confirm.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> open_confirm.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `open_confirm_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> open_confirm_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_open_confirm_total

Total number of OPEN_CONFIRM procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `open_confirm.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `open_confirm_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_open_downgrade_avg_latency

Average latency of OPEN_DOWNGRADE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `open_downgrade.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> open_downgrade.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `open_downgrade_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> open_downgrade_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_open_downgrade_total

Total number of OPEN_DOWNGRADE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `open_downgrade.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `open_downgrade_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_open_total

Total number of OPEN procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `open.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `open_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_openattr_avg_latency

Average latency of OPENATTR procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `openattr.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> openattr.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `openattr_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> openattr_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_openattr_total

Total number of OPENATTR procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `openattr.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `openattr_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_ops

Number of NFS operations per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `nfs_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `nfs_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_nfs_pathconf_avg_latency

Average latency of PathConf procedure requests. The counter keeps track of the average response time of PathConf requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `pathconf.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> pathconf.total | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `pathconf_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> pathconf_total | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_pathconf_total

Total number PathConf of procedure requests. It is the total number of PathConf success and PathConf error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `pathconf.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `pathconf_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_putfh_avg_latency

Average latency of PUTFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `putfh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> putfh.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `putfh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> putfh_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_putfh_total

Total number of PUTFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `putfh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `putfh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_putpubfh_avg_latency

Average latency of PUTPUBFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `putpubfh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> putpubfh.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `putpubfh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> putpubfh_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_putpubfh_total

Total number of PUTPUBFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `putpubfh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `putpubfh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_putrootfh_avg_latency

Average latency of PUTROOTFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `putrootfh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> putrootfh.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `putrootfh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> putrootfh_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_putrootfh_total

Total number of PUTROOTFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `putrootfh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `putrootfh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_read_avg_latency

Average latency of READ procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `read.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `read_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> read_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_read_ops

Total observed NFSv3 read operations per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `nfsv3_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_read_symlink_avg_latency

Average latency of ReadSymLink procedure requests. The counter keeps track of the average response time of ReadSymLink requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `read_symlink.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_symlink.total | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `read_symlink_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> read_symlink_total | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_read_symlink_total

Total number of ReadSymLink procedure requests. It is the total number of read symlink success and read symlink error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `read_symlink.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `read_symlink_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_read_throughput

NFSv4 read data transfers

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `total.read_throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `nfs4_read_throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_read_total

Total number of READ procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `read.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `read_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_readdir_avg_latency

Average latency of READDIR procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `readdir.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> readdir.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `readdir_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> readdir_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_readdir_total

Total number of READDIR procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `readdir.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `readdir_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_readdirplus_avg_latency

Average latency of ReadDirPlus procedure requests. The counter keeps track of the average response time of ReadDirPlus requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `readdirplus.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> readdirplus.total | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `readdirplus_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> readdirplus_total | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_readdirplus_total

Total number ReadDirPlus of procedure requests. It is the total number of ReadDirPlus success and ReadDirPlus error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `readdirplus.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `readdirplus_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_readlink_avg_latency

Average latency of READLINK procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `readlink.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> readlink.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `readlink_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> readlink_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_readlink_total

Total number of READLINK procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `readlink.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `readlink_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_reclaim_complete_avg_latency

Average latency of NFSv4.2 RECLAIM_complete operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `reclaim_complete.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> reclaim_complete.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `reclaim_complete_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> reclaim_complete_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_reclaim_complete_total

Total number of NFSv4.2 RECLAIM_complete operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `reclaim_complete.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `reclaim_complete_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_release_lock_owner_avg_latency

Average Latency of RELEASE_LOCKOWNER procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `release_lock_owner.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> release_lock_owner.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `release_lock_owner_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> release_lock_owner_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_release_lock_owner_total

Total number of RELEASE_LOCKOWNER procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `release_lock_owner.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `release_lock_owner_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_remove_avg_latency

Average latency of REMOVE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `remove.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> remove.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `remove_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> remove_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_remove_total

Total number of REMOVE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `remove.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `remove_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_rename_avg_latency

Average latency of RENAME procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `rename.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> rename.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `rename_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> rename_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_rename_total

Total number of RENAME procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `rename.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `rename_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_renew_avg_latency

Average latency of RENEW procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `renew.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> renew.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `renew_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> renew_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_renew_total

Total number of RENEW procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `renew.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `renew_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_restorefh_avg_latency

Average latency of RESTOREFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `restorefh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> restorefh.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `restorefh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> restorefh_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_restorefh_total

Total number of RESTOREFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `restorefh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `restorefh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_rmdir_avg_latency

Average latency of RmDir procedure requests. The counter keeps track of the average response time of RmDir requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `rmdir.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> rmdir.total | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `rmdir_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> rmdir_total | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_rmdir_total

Total number RmDir of procedure requests. It is the total number of RmDir success and RmDir error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `rmdir.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `rmdir_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_savefh_avg_latency

Average latency of SAVEFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `savefh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> savefh.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `savefh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> savefh_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_savefh_total

Total number of SAVEFH procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `savefh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `savefh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_secinfo_avg_latency

Average latency of SECINFO procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `secinfo.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> secinfo.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `secinfo_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> secinfo_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_secinfo_no_name_avg_latency

Average latency of NFSv4.2 SECINFO_NO_NAME operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `secinfo_no_name.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> secinfo_no_name.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `secinfo_no_name_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> secinfo_no_name_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_secinfo_no_name_total

Total number of NFSv4.2 SECINFO_NO_NAME operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `secinfo_no_name.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `secinfo_no_name_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_secinfo_total

Total number of SECINFO procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `secinfo.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `secinfo_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_sequence_avg_latency

Average latency of NFSv4.2 SEQUENCE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `sequence.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> sequence.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `sequence_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> sequence_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_sequence_total

Total number of NFSv4.2 SEQUENCE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `sequence.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `sequence_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_set_ssv_avg_latency

Average latency of NFSv4.2 SET_SSV operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `set_ssv.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> set_ssv.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `set_ssv_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> set_ssv_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_set_ssv_total

Total number of NFSv4.2 SET_SSV operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `set_ssv.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `set_ssv_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_setattr_avg_latency

Average latency of SETATTR procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `setattr.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> setattr.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `setattr_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> setattr_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_setattr_total

Total number of SETATTR procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `setattr.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `setattr_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_setclientid_avg_latency

Average latency of SETCLIENTID procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `setclientid.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> setclientid.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `setclientid_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> setclientid_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_setclientid_confirm_avg_latency

Average latency of SETCLIENTID_CONFIRM procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `setclientid_confirm.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> setclientid_confirm.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `setclientid_confirm_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> setclientid_confirm_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_setclientid_confirm_total

Total number of SETCLIENTID_CONFIRM procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `setclientid_confirm.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `setclientid_confirm_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_setclientid_total

Total number of SETCLIENTID procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `setclientid.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `setclientid_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_symlink_avg_latency

Average latency of SymLink procedure requests. The counter keeps track of the average response time of SymLink requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `symlink.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> symlink.total | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `symlink_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> symlink_total | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_symlink_total

Total number SymLink of procedure requests. It is the total number of SymLink success and create SymLink requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `symlink.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `symlink_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_test_stateid_avg_latency

Average latency of NFSv4.2 TEST_STATEID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `test_stateid.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> test_stateid.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `test_stateid_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> test_stateid_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_test_stateid_total

Total number of NFSv4.2 TEST_STATEID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `test_stateid.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `test_stateid_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_throughput

NFSv4 data transfers

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `total.throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `nfs4_throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_total_ops

Total number of NFSv4 requests per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_verify_avg_latency

Average latency of VERIFY procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `verify.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> verify.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `verify_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> verify_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_verify_total

Total number of VERIFY procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `verify.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `verify_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_want_delegation_avg_latency

Average latency of NFSv4.2 WANT_DELEGATION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `want_delegation.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> want_delegation.total | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `want_delegation_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> want_delegation_total | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_want_delegation_total

Total number of NFSv4.2 WANT_DELEGATION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42:node` | `want_delegation.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1:node` | `want_delegation_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml | 


### node_nfs_write_avg_latency

Average Latency of WRITE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `write.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write.total | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `write_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> write_total | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_write_ops

Total observed NFSv3 write operations per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3:node` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv3:node` | `nfsv3_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml | 


### node_nfs_write_throughput

NFSv4 write data transfers

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `total.write_throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `nfs4_write_throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nfs_write_total

Total number of WRITE procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4:node` | `write.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_node.yaml | 
| ZAPI | `perf-object-get-instances nfsv4:node` | `write_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml | 


### node_nvmf_data_recv

NVMe/FC kilobytes (KB) received per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `nvme_fc_data_received`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `nvmf_data_recv`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_nvmf_data_sent

NVMe/FC kilobytes (KB) sent per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `nvme_fc_data_sent`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `nvmf_data_sent`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_nvmf_ops

NVMe/FC operations per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `nvme_fc_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `nvmf_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_ssd_data_read

Number of SSD Disk kilobytes (KB) read per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `ssd_data_read`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `ssd_data_read`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_ssd_data_written

Number of SSD Disk kilobytes (KB) written per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `ssd_data_written`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `ssd_data_written`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_total_data

Total throughput in bytes

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_total_latency

Average latency for all operations in the system in microseconds

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `total_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `total_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_total_ops

Total number of operations per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/system:node` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/system_node.yaml | 
| ZAPI | `perf-object-get-instances system:node` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/system_node.yaml | 


### node_uptime

The total time, in seconds, that the node has been up.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/nodes` | `uptime` | conf/rest/9.12.0/node.yaml |
| ZAPI | `system-node-get-iter` | `node-details-info.node-uptime` | conf/zapi/cdot/9.8.0/node.yaml |


### node_vol_cifs_other_latency

Average time for the WAFL filesystem to process other CIFS operations to the volume; not including CIFS protocol request processing or network communication time which will also be included in client observed CIFS request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `cifs.other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs.other_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `cifs_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs_other_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_cifs_other_ops

Number of other CIFS operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `cifs.other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `cifs_other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_cifs_read_data

Bytes read per second via CIFS

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `cifs.read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `cifs_read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_cifs_read_latency

Average time for the WAFL filesystem to process CIFS read requests to the volume; not including CIFS protocol request processing or network communication time which will also be included in client observed CIFS request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `cifs.read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs.read_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `cifs_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs_read_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_cifs_read_ops

Number of CIFS read operations per second from the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `cifs.read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `cifs_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_cifs_write_data

Bytes written per second via CIFS

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `cifs.write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `cifs_write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_cifs_write_latency

Average time for the WAFL filesystem to process CIFS write requests to the volume; not including CIFS protocol request processing or network communication time which will also be included in client observed CIFS request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `cifs.write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs.write_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `cifs_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs_write_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_cifs_write_ops

Number of CIFS write operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `cifs.write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `cifs_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_fcp_other_latency

Average time for the WAFL filesystem to process other FCP protocol operations to the volume; not including FCP protocol request processing or network communication time which will also be included in client observed FCP protocol request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `fcp.other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> fcp.other_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `fcp_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> fcp_other_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_fcp_other_ops

Number of other block protocol operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `fcp.other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `fcp_other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_fcp_read_data

Bytes read per second via block protocol

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `fcp.read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `fcp_read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_fcp_read_latency

Average time for the WAFL filesystem to process FCP protocol read operations to the volume; not including FCP protocol request processing or network communication time which will also be included in client observed FCP protocol request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `fcp.read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> fcp.read_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `fcp_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> fcp_read_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_fcp_read_ops

Number of block protocol read operations per second from the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `fcp.read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `fcp_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_fcp_write_data

Bytes written per second via block protocol

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `fcp.write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `fcp_write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_fcp_write_latency

Average time for the WAFL filesystem to process FCP protocol write operations to the volume; not including FCP protocol request processing or network communication time which will also be included in client observed FCP protocol request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `fcp.write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> fcp.write_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `fcp_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> fcp_write_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_fcp_write_ops

Number of block protocol write operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `fcp.write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `fcp_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_iscsi_other_latency

Average time for the WAFL filesystem to process other iSCSI protocol operations to the volume; not including iSCSI protocol request processing or network communication time which will also be included in client observed iSCSI protocol request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `iscsi.other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi.other_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `iscsi_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi_other_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_iscsi_other_ops

Number of other block protocol operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `iscsi.other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `iscsi_other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_iscsi_read_data

Bytes read per second via block protocol

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `iscsi.read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `iscsi_read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_iscsi_read_latency

Average time for the WAFL filesystem to process iSCSI protocol read operations to the volume; not including iSCSI protocol request processing or network communication time which will also be included in client observed iSCSI protocol request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `iscsi.read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi.read_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `iscsi_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi_read_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_iscsi_read_ops

Number of block protocol read operations per second from the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `iscsi.read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `iscsi_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_iscsi_write_data

Bytes written per second via block protocol

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `iscsi.write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `iscsi_write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_iscsi_write_latency

Average time for the WAFL filesystem to process iSCSI protocol write operations to the volume; not including iSCSI protocol request processing or network communication time which will also be included in client observed iSCSI request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `iscsi.write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi.write_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `iscsi_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> iscsi_write_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_iscsi_write_ops

Number of block protocol write operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `iscsi.write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `iscsi_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_nfs_other_latency

Average time for the WAFL filesystem to process other NFS operations to the volume; not including NFS protocol request processing or network communication time which will also be included in client observed NFS request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `nfs.other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nfs.other_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `nfs_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nfs_other_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_nfs_other_ops

Number of other NFS operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `nfs.other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `nfs_other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_nfs_read_data

Bytes read per second via NFS

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `nfs.read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `nfs_read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_nfs_read_latency

Average time for the WAFL filesystem to process NFS protocol read requests to the volume; not including NFS protocol request processing or network communication time which will also be included in client observed NFS request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `nfs.read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nfs.read_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `nfs_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nfs_read_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_nfs_read_ops

Number of NFS read operations per second from the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `nfs.read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `nfs_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_nfs_write_data

Bytes written per second via NFS

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `nfs.write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `nfs_write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_nfs_write_latency

Average time for the WAFL filesystem to process NFS protocol write requests to the volume; not including NFS protocol request processing or network communication time, which will also be included in client observed NFS request latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `nfs.write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nfs.write_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `nfs_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nfs_write_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_nfs_write_ops

Number of NFS write operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `nfs.write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `nfs_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_read_latency

Average latency in microseconds for the WAFL filesystem to process read request to the volume; not including request processing or network communication time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_read_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### node_vol_write_latency

Average latency in microseconds for the WAFL filesystem to process write request to the volume; not including request processing or network communication time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:node` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_write_ops | conf/restperf/9.12.0/volume_node.yaml | 
| ZAPI | `perf-object-get-instances volume:node` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/zapiperf/cdot/9.8.0/volume_node.yaml | 


### nvme_lif_avg_latency

Average latency for NVMF operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nvmf_lif` | `average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/restperf/9.12.0/nvmf_lif.yaml | 
| ZAPI | `perf-object-get-instances nvmf_fc_lif` | `avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml | 


### nvme_lif_avg_other_latency

Average latency for operations other than read, write, compare or compare-and-write.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nvmf_lif` | `average_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> other_ops | conf/restperf/9.12.0/nvmf_lif.yaml | 
| ZAPI | `perf-object-get-instances nvmf_fc_lif` | `avg_other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> other_ops | conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml | 


### nvme_lif_avg_read_latency

Average latency for read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nvmf_lif` | `average_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/restperf/9.12.0/nvmf_lif.yaml | 
| ZAPI | `perf-object-get-instances nvmf_fc_lif` | `avg_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml | 


### nvme_lif_avg_write_latency

Average latency for write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nvmf_lif` | `average_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/restperf/9.12.0/nvmf_lif.yaml | 
| ZAPI | `perf-object-get-instances nvmf_fc_lif` | `avg_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml | 


### nvme_lif_other_ops

Number of operations that are not read, write, compare or compare-and-write.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nvmf_lif` | `other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nvmf_lif.yaml | 
| ZAPI | `perf-object-get-instances nvmf_fc_lif` | `other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml | 


### nvme_lif_read_data

Amount of data read from the storage system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nvmf_lif` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nvmf_lif.yaml | 
| ZAPI | `perf-object-get-instances nvmf_fc_lif` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml | 


### nvme_lif_read_ops

Number of read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nvmf_lif` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nvmf_lif.yaml | 
| ZAPI | `perf-object-get-instances nvmf_fc_lif` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml | 


### nvme_lif_total_ops

Total number of operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nvmf_lif` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nvmf_lif.yaml | 
| ZAPI | `perf-object-get-instances nvmf_fc_lif` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml | 


### nvme_lif_write_data

Amount of data written to the storage system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nvmf_lif` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nvmf_lif.yaml | 
| ZAPI | `perf-object-get-instances nvmf_fc_lif` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml | 


### nvme_lif_write_ops

Number of write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/nvmf_lif` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nvmf_lif.yaml | 
| ZAPI | `perf-object-get-instances nvmf_fc_lif` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml | 


### ontaps3_logical_used_size

Specifies the bucket logical used size up to this point.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/protocols/s3/buckets` | `logical_used_size` | conf/rest/9.7.0/ontap_s3.yaml |


### ontaps3_size

Specifies the bucket size in bytes; ranges from 80MB to 64TB.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/protocols/s3/buckets` | `size` | conf/rest/9.7.0/ontap_s3.yaml |


### ontaps3_svm_abort_multipart_upload_failed

Number of failed Abort Multipart Upload operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `abort_multipart_upload_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_abort_multipart_upload_failed_client_close

Number of times Abort Multipart Upload operation failed because client terminated connection for operation pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `abort_multipart_upload_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_abort_multipart_upload_latency

Average latency for Abort Multipart Upload operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `abort_multipart_upload_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> abort_multipart_upload_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_abort_multipart_upload_rate

Number of Abort Multipart Upload operations per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `abort_multipart_upload_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_abort_multipart_upload_total

Number of Abort Multipart Upload operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `abort_multipart_upload_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_allow_access

Number of times access was allowed.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `allow_access`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_anonymous_access

Number of times anonymous access was allowed.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `anonymous_access`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_anonymous_deny_access

Number of times anonymous access was denied.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `anonymous_deny_access`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_authentication_failures

Number of authentication failures.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `authentication_failures`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_chunked_upload_reqs

Total number of object store server chunked object upload requests

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `chunked_upload_reqs`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_complete_multipart_upload_failed

Number of failed Complete Multipart Upload operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `complete_multipart_upload_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_complete_multipart_upload_failed_client_close

Number of times Complete Multipart Upload operation failed because client terminated connection for operation pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `complete_multipart_upload_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_complete_multipart_upload_latency

Average latency for Complete Multipart Upload operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `complete_multipart_upload_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> complete_multipart_upload_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_complete_multipart_upload_rate

Number of Complete Multipart Upload operations per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `complete_multipart_upload_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_complete_multipart_upload_total

Number of Complete Multipart Upload operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `complete_multipart_upload_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_connected_connections

Number of object store server connections currently established

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `connected_connections`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_connections

Total number of object store server connections.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `connections`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_create_bucket_failed

Number of failed Create Bucket operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `create_bucket_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_create_bucket_failed_client_close

Number of times Create Bucket operation failed because client terminated connection for operation pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `create_bucket_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_create_bucket_latency

Average latency for Create Bucket operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `create_bucket_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> create_bucket_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_create_bucket_rate

Number of Create Bucket operations per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `create_bucket_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_create_bucket_total

Number of Create Bucket operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `create_bucket_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_default_deny_access

Number of times access was denied by default and not through any policy statement.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `default_deny_access`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_bucket_failed

Number of failed Delete Bucket operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_bucket_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_bucket_failed_client_close

Number of times Delete Bucket operation failed because client terminated connection for operation pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_bucket_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_bucket_latency

Average latency for Delete Bucket operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_bucket_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> delete_bucket_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_bucket_rate

Number of Delete Bucket operations per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_bucket_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_bucket_total

Number of Delete Bucket operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_bucket_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_object_failed

Number of failed DELETE object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_object_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_object_failed_client_close

Number of times DELETE object operation failed due to the case where client closed the connection while the operation was still pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_object_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_object_latency

Average latency for DELETE object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_object_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> delete_object_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_object_rate

Number of DELETE object operations per sec

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_object_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_object_tagging_failed

Number of failed DELETE object tagging operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_object_tagging_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_object_tagging_failed_client_close

Number of times DELETE object tagging operation failed because client terminated connection for operation pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_object_tagging_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_object_tagging_latency

Average latency for DELETE object tagging operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_object_tagging_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> delete_object_tagging_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_object_tagging_rate

Number of DELETE object tagging operations per sec.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_object_tagging_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_object_tagging_total

Number of DELETE object tagging operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_object_tagging_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_delete_object_total

Number of DELETE object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `delete_object_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_explicit_deny_access

Number of times access was denied explicitly by a policy statement.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `explicit_deny_access`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_bucket_acl_failed

Number of failed GET Bucket ACL operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_bucket_acl_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_bucket_acl_total

Number of GET Bucket ACL operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_bucket_acl_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_bucket_versioning_failed

Number of failed Get Bucket Versioning operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_bucket_versioning_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_bucket_versioning_total

Number of Get Bucket Versioning operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_bucket_versioning_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_data

Rate of GET object data transfers per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_acl_failed

Number of failed GET Object ACL operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_acl_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_acl_total

Number of GET Object ACL operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_acl_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_failed

Number of failed GET object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_failed_client_close

Number of times GET object operation failed due to the case where client closed the connection while the operation was still pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_lastbyte_latency

Average last-byte latency for GET object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_lastbyte_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> get_object_lastbyte_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_latency

Average first-byte latency for GET object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> get_object_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_rate

Number of GET object operations per sec

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_tagging_failed

Number of failed GET object tagging operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_tagging_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_tagging_failed_client_close

Number of times GET object tagging operation failed due to the case where client closed the connection while the operation was still pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_tagging_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_tagging_latency

Average latency for GET object tagging operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_tagging_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> get_object_tagging_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_tagging_rate

Number of GET object tagging operations per sec

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_tagging_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_tagging_total

Number of GET object tagging operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_tagging_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_get_object_total

Number of GET object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `get_object_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_group_policy_evaluated

Number of times group policies were evaluated.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `group_policy_evaluated`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_head_bucket_failed

Number of failed HEAD bucket operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `head_bucket_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_head_bucket_failed_client_close

Number of times HEAD bucket operation failed due to the case where client closed the connection while the operation was still pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `head_bucket_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_head_bucket_latency

Average latency for HEAD bucket operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `head_bucket_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> head_bucket_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_head_bucket_rate

Number of HEAD bucket operations per sec

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `head_bucket_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_head_bucket_total

Number of HEAD bucket operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `head_bucket_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_head_object_failed

Number of failed HEAD Object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `head_object_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_head_object_failed_client_close

Number of times HEAD object operation failed due to the case where client closed the connection while the operation was still pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `head_object_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_head_object_latency

Average latency for HEAD object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `head_object_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> head_object_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_head_object_rate

Number of HEAD Object operations per sec

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `head_object_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_head_object_total

Number of HEAD Object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `head_object_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_initiate_multipart_upload_failed

Number of failed Initiate Multipart Upload operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `initiate_multipart_upload_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_initiate_multipart_upload_failed_client_close

Number of times Initiate Multipart Upload operation failed because client terminated connection for operation pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `initiate_multipart_upload_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_initiate_multipart_upload_latency

Average latency for Initiate Multipart Upload operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `initiate_multipart_upload_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> initiate_multipart_upload_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_initiate_multipart_upload_rate

Number of Initiate Multipart Upload operations per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `initiate_multipart_upload_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_initiate_multipart_upload_total

Number of Initiate Multipart Upload operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `initiate_multipart_upload_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_input_flow_control_entry

Number of times input flow control was entered.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `input_flow_control_entry`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_input_flow_control_exit

Number of times input flow control was exited.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `input_flow_control_exit`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_buckets_failed

Number of failed LIST Buckets operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_buckets_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_buckets_failed_client_close

Number of times LIST Bucket operation failed due to the case where client closed the connection while the operation was still pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_buckets_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_buckets_latency

Average latency for LIST Buckets operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_buckets_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> head_object_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_buckets_rate

Number of LIST Buckets operations per sec

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_buckets_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_buckets_total

Number of LIST Buckets operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_buckets_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_object_versions_failed

Number of failed LIST object versions operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_object_versions_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_object_versions_failed_client_close

Number of times LIST object versions operation failed due to the case where client closed the connection while the operation was still pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_object_versions_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_object_versions_latency

Average latency for LIST Object versions operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_object_versions_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> list_object_versions_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_object_versions_rate

Number of LIST Object Versions operations per sec

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_object_versions_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_object_versions_total

Number of LIST Object Versions operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_object_versions_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_objects_failed

Number of failed LIST objects operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_objects_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_objects_failed_client_close

Number of times LIST objects operation failed due to the case where client closed the connection while the operation was still pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_objects_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_objects_latency

Average latency for LIST Objects operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_objects_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> list_objects_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_objects_rate

Number of LIST Objects operations per sec

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_objects_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_objects_total

Number of LIST Objects operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_objects_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_uploads_failed

Number of failed LIST Uploads operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_uploads_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_uploads_failed_client_close

Number of times LIST Uploads operation failed due to the case where client closed the connection while the operation was still pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_uploads_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_uploads_latency

Average latency for LIST Uploads operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_uploads_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> list_uploads_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_uploads_rate

Number of LIST Uploads operations per sec

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_uploads_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_list_uploads_total

Number of LIST Uploads operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `list_uploads_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_max_cmds_per_connection

Maximum commands pipelined at any instance on a connection.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `max_cmds_per_connection`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_max_connected_connections

Maximum number of object store server connections established at one time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `max_connected_connections`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_max_requests_outstanding

Maximum number of object store server requests in process at one time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `max_requests_outstanding`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_multi_delete_reqs

Total number of object store server multiple object delete requests

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `multi_delete_reqs`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_output_flow_control_entry

Number of output flow control was entered.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `output_flow_control_entry`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_output_flow_control_exit

Number of times output flow control was exited.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `output_flow_control_exit`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_presigned_url_reqs

Total number of presigned object store server URL requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `presigned_url_reqs`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_bucket_versioning_failed

Number of failed Put Bucket Versioning operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_bucket_versioning_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_bucket_versioning_total

Number of Put Bucket Versioning operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_bucket_versioning_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_data

Rate of PUT object data transfers per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_object_failed

Number of failed PUT object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_object_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_object_failed_client_close

Number of times PUT object operation failed due to the case where client closed the connection while the operation was still pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_object_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_object_latency

Average latency for PUT object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_object_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> put_object_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_object_rate

Number of PUT object operations per sec

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_object_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_object_tagging_failed

Number of failed PUT object tagging operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_object_tagging_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_object_tagging_failed_client_close

Number of times PUT object tagging operation failed because client terminated connection for operation pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_object_tagging_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_object_tagging_latency

Average latency for PUT object tagging operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_object_tagging_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> put_object_tagging_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_object_tagging_rate

Number of PUT object tagging operations per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_object_tagging_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_object_tagging_total

Number of PUT object tagging operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_object_tagging_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_put_object_total

Number of PUT object operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `put_object_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_request_parse_errors

Number of request parser errors due to malformed requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `request_parse_errors`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_requests

Total number of object store server requests

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `requests`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_requests_outstanding

Number of object store server requests in process

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `requests_outstanding`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_root_user_access

Number of times access was done by root user.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `root_user_access`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_server_connection_close

Number of connection closes triggered by server due to fatal errors.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `server_connection_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_signature_v2_reqs

Total number of object store server signature V2 requests

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `signature_v2_reqs`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_signature_v4_reqs

Total number of object store server signature V4 requests

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `signature_v4_reqs`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_tagging

Number of requests with tagging specified.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `tagging`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_upload_part_failed

Number of failed Upload Part operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `upload_part_failed`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_upload_part_failed_client_close

Number of times Upload Part operation failed because client terminated connection for operation pending on server.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `upload_part_failed_client_close`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_upload_part_latency

Average latency for Upload Part operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `upload_part_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> upload_part_latency_base | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_upload_part_rate

Number of Upload Part operations per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `upload_part_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### ontaps3_svm_upload_part_total

Number of Upload Part operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances object_store_server` | `upload_part_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/ontap_s3_svm.yaml | 


### path_read_data

The average read throughput in kilobytes per second read from the indicated target port by the controller.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/path` | `read_data`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/path.yaml | 
| ZAPI | `perf-object-get-instances path` | `read_data`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/path.yaml | 


### path_read_iops

The number of I/O read operations sent from the initiator port to the indicated target port.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/path` | `read_iops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/path.yaml | 
| ZAPI | `perf-object-get-instances path` | `read_iops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/path.yaml | 


### path_read_latency

The average latency of I/O read operations sent from this controller to the indicated target port.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/path` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_iops | conf/restperf/9.12.0/path.yaml | 
| ZAPI | `perf-object-get-instances path` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_iops | conf/zapiperf/cdot/9.8.0/path.yaml | 


### path_total_data

The average throughput in kilobytes per second read and written from/to the indicated target port by the controller.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/path` | `total_data`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/path.yaml | 
| ZAPI | `perf-object-get-instances path` | `total_data`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/path.yaml | 


### path_total_iops

The number of total read/write I/O operations sent from the initiator port to the indicated target port.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/path` | `total_iops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/path.yaml | 
| ZAPI | `perf-object-get-instances path` | `total_iops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/path.yaml | 


### path_write_data

The average write throughput in kilobytes per second written to the indicated target port by the controller.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/path` | `write_data`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/path.yaml | 
| ZAPI | `perf-object-get-instances path` | `write_data`<br><span class="key">Unit:</span> kb_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/path.yaml | 


### path_write_iops

The number of I/O write operations sent from the initiator port to the indicated target port.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/path` | `write_iops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/path.yaml | 
| ZAPI | `perf-object-get-instances path` | `write_iops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/path.yaml | 


### path_write_latency

The average latency of I/O write operations sent from this controller to the indicated target port.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/path` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_iops | conf/restperf/9.12.0/path.yaml | 
| ZAPI | `perf-object-get-instances path` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_iops | conf/zapiperf/cdot/9.8.0/path.yaml | 


### qos_concurrency

This is the average number of concurrent requests for the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `concurrency`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `concurrency`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_detail_resource_latency

average latency for workload on Data ONTAP subsystems

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_detail` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload_detail.yaml | 
| ZAPI | `perf-object-get-instances workload_detail` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/9.12.0/workload_detail.yaml | 


### qos_detail_volume_resource_latency

average latency for volume on Data ONTAP subsystems

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_detail_volume` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload_detail_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_detail_volume` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/9.12.0/workload_detail_volume.yaml | 


### qos_latency

This is the average response time for requests that were initiated by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> ops | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> ops | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_ops

Workload operations executed per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_read_data

This is the amount of data read per second from the filer by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_read_io_type

This is the percentage of read requests served from various components (such as buffer cache, ext_cache, disk, etc.).

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `read_io_type_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_io_type_base | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `read_io_type`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_io_type_base | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_read_latency

This is the average response time for read requests that were initiated by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_read_ops

This is the rate of this workload's read operations that completed during the measurement interval.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_sequential_reads

This is the percentage of reads, performed on behalf of the workload, that were sequential.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `sequential_reads_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> sequential_reads_base | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `sequential_reads`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent,no-zero-values<br><span class="key">Base:</span> sequential_reads_base | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_sequential_writes

This is the percentage of writes, performed on behalf of the workload, that were sequential. This counter is only available on platforms with more than 4GB of NVRAM.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `sequential_writes_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> sequential_writes_base | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `sequential_writes`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent,no-zero-values<br><span class="key">Base:</span> sequential_writes_base | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_total_data

This is the total amount of data read/written per second from/to the filer by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_volume_latency

This is the average response time for requests that were initiated by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> ops | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> ops | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_volume_ops

This field is the workload's rate of operations that completed during the measurement interval; measured per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_volume_read_data

This is the amount of data read per second from the filer by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_volume_read_io_type

This is the percentage of read requests served from various components (such as buffer cache, ext_cache, disk, etc.).

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `read_io_type_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_io_type_base | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `read_io_type`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_io_type_base | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_volume_read_latency

This is the average response time for read requests that were initiated by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_volume_read_ops

This is the rate of this workload's read operations that completed during the measurement interval.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_volume_sequential_reads

This is the percentage of reads, performed on behalf of the workload, that were sequential.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `sequential_reads_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> sequential_reads_base | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `sequential_reads`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent,no-zero-values<br><span class="key">Base:</span> sequential_reads_base | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_volume_sequential_writes

This is the percentage of writes, performed on behalf of the workload, that were sequential. This counter is only available on platforms with more than 4GB of NVRAM.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `sequential_writes_percent`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> sequential_writes_base | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `sequential_writes`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent,no-zero-values<br><span class="key">Base:</span> sequential_writes_base | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_volume_total_data

This is the total amount of data read/written per second from/to the filer by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `total_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_volume_write_data

This is the amount of data written per second to the filer by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_volume_write_latency

This is the average response time for write requests that were initiated by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> write_ops | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_volume_write_ops

This is the workload's write operations that completed during the measurement interval; measured per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos_volume` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload_volume.yaml | 
| ZAPI | `perf-object-get-instances workload_volume` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload_volume.yaml | 


### qos_write_data

This is the amount of data written per second to the filer by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_write_latency

This is the average response time for write requests that were initiated by the workload.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> write_ops | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qos_write_ops

This is the workload's write operations that completed during the measurement interval; measured per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qos` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/workload.yaml | 
| ZAPI | `perf-object-get-instances workload` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/workload.yaml | 


### qtree_cifs_ops

Number of CIFS operations per second to the qtree

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qtree` | `cifs_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/qtree.yaml | 
| ZAPI | `perf-object-get-instances qtree` | `cifs_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/qtree.yaml | 


### qtree_id

The identifier for the qtree, unique within the qtree's volume.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/qtrees` | `id` | conf/rest/9.12.0/qtree.yaml |


### qtree_internal_ops

Number of internal operations generated by activites such as snapmirror and backup per second to the qtree

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qtree` | `internal_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/qtree.yaml | 
| ZAPI | `perf-object-get-instances qtree` | `internal_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/qtree.yaml | 


### qtree_nfs_ops

Number of NFS operations per second to the qtree

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qtree` | `nfs_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/qtree.yaml | 
| ZAPI | `perf-object-get-instances qtree` | `nfs_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/qtree.yaml | 


### qtree_total_ops

Summation of NFS ops, CIFS ops, CSS ops and internal ops

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/qtree` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/qtree.yaml | 
| ZAPI | `perf-object-get-instances qtree` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/qtree.yaml | 


### quota_disk_limit

Maximum amount of disk space, in kilobytes, allowed for the quota target (hard disk space limit). The value is -1 if the limit is unlimited.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/quota/reports` | `space.hard_limit` | conf/rest/9.12.0/qtree.yaml |
| ZAPI | `quota-report-iter` | `disk-limit` | conf/zapi/cdot/9.8.0/qtree.yaml |


### quota_disk_used

Current amount of disk space, in kilobytes, used by the quota target.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/quota/reports` | `space.used.total` | conf/rest/9.12.0/qtree.yaml |
| ZAPI | `quota-report-iter` | `disk-used` | conf/zapi/cdot/9.8.0/qtree.yaml |


### quota_disk_used_pct_disk_limit

Current disk space used expressed as a percentage of hard disk limit.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/quota/reports` | `space.used.hard_limit_percent` | conf/rest/9.12.0/qtree.yaml |
| ZAPI | `quota-report-iter` | `disk-used-pct-disk-limit` | conf/zapi/cdot/9.8.0/qtree.yaml |


### quota_disk_used_pct_soft_disk_limit

Current disk space used expressed as a percentage of soft disk limit.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/quota/reports` | `space.used.soft_limit_percent` | conf/rest/9.12.0/qtree.yaml |
| ZAPI | `quota-report-iter` | `disk-used-pct-soft-disk-limit` | conf/zapi/cdot/9.8.0/qtree.yaml |


### quota_disk_used_pct_threshold

Current disk space used expressed as a percentage of threshold.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `quota-report-iter` | `disk-used-pct-threshold` | conf/zapi/cdot/9.8.0/qtree.yaml |


### quota_file_limit

Maximum number of files allowed for the quota target (hard files limit). The value is -1 if the limit is unlimited.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/quota/reports` | `files.hard_limit` | conf/rest/9.12.0/qtree.yaml |
| ZAPI | `quota-report-iter` | `file-limit` | conf/zapi/cdot/9.8.0/qtree.yaml |


### quota_files_used

Current number of files used by the quota target.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/quota/reports` | `files.used.total` | conf/rest/9.12.0/qtree.yaml |
| ZAPI | `quota-report-iter` | `files-used` | conf/zapi/cdot/9.8.0/qtree.yaml |


### quota_files_used_pct_file_limit

Current number of files used expressed as a percentage of hard file limit.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/quota/reports` | `files.used.hard_limit_percent` | conf/rest/9.12.0/qtree.yaml |
| ZAPI | `quota-report-iter` | `files-used-pct-file-limit` | conf/zapi/cdot/9.8.0/qtree.yaml |


### quota_files_used_pct_soft_file_limit

Current number of files used expressed as a percentage of soft file limit.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/quota/reports` | `files.used.soft_limit_percent` | conf/rest/9.12.0/qtree.yaml |
| ZAPI | `quota-report-iter` | `files-used-pct-soft-file-limit` | conf/zapi/cdot/9.8.0/qtree.yaml |


### quota_soft_disk_limit

soft disk space limit, in kilobytes, for the quota target. The value is -1 if the limit is unlimited.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/quota/reports` | `space.soft_limit` | conf/rest/9.12.0/qtree.yaml |
| ZAPI | `quota-report-iter` | `soft-disk-limit` | conf/zapi/cdot/9.8.0/qtree.yaml |


### quota_soft_file_limit

Soft file limit, in number of files, for the quota target. The value is -1 if the limit is unlimited.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/quota/reports` | `files.soft_limit` | conf/rest/9.12.0/qtree.yaml |
| ZAPI | `quota-report-iter` | `soft-file-limit` | conf/zapi/cdot/9.8.0/qtree.yaml |


### quota_threshold

Disk space threshold, in kilobytes, for the quota target. The value is -1 if the limit is unlimited.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `quota-report-iter` | `threshold` | conf/zapi/cdot/9.8.0/qtree.yaml |


### security_audit_destination_port

The destination port used to forward the message.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `cluster-log-forward-get-iter` | `cluster-log-forward-info.port` | conf/zapi/cdot/9.8.0/security_audit_dest.yaml |


### security_certificate_expiry_time

Certificate expiration time. Can be provided on POST if creating self-signed certificate. The expiration time range is between 1 day to 10 years.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/security/certificates` | `expiry_time` | conf/rest/9.12.0/security_certificate.yaml |
| ZAPI | `security-certificate-get-iter` | `certificate-info.expiration-date` | conf/zapi/cdot/9.8.0/security_certificate.yaml |


### security_ssh_max_instances

Maximum possible simultaneous connections.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/security/ssh` | `max_instances` | conf/rest/9.12.0/security_ssh.yaml |


### shelf_average_ambient_temperature

Average temperature of all ambient sensors for shelf in Celsius.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/disk.yaml | 
| ZAPI | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/disk.yaml | 


### shelf_average_fan_speed

Average fan speed for shelf in rpm.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/disk.yaml | 
| ZAPI | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/disk.yaml | 


### shelf_average_temperature

Average temperature of all non-ambient sensors for shelf in Celsius.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/disk.yaml | 
| ZAPI | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/disk.yaml | 


### shelf_disk_count

Disk count in a shelf.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/shelves` | `disk_count` | conf/rest/9.12.0/shelf.yaml |
| ZAPI | `storage-shelf-info-get-iter` | `storage-shelf-info.disk-count` | conf/zapi/cdot/9.8.0/shelf.yaml |


### shelf_max_fan_speed

Maximum fan speed for shelf in rpm.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/disk.yaml | 
| ZAPI | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/disk.yaml | 


### shelf_max_temperature

Maximum temperature of all non-ambient sensors for shelf in Celsius.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/disk.yaml | 
| ZAPI | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/disk.yaml | 


### shelf_min_ambient_temperature

Minimum temperature of all ambient sensors for shelf in Celsius.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/disk.yaml | 
| ZAPI | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/disk.yaml | 


### shelf_min_fan_speed

Minimum fan speed for shelf in rpm.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/disk.yaml | 
| ZAPI | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/disk.yaml | 


### shelf_min_temperature

Minimum temperature of all non-ambient sensors for shelf in Celsius.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/disk.yaml | 
| ZAPI | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/disk.yaml | 


### shelf_power

Power consumed by shelf in Watts.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/restperf/9.12.0/disk.yaml | 
| ZAPI | `NA` | `Harvest generated`<br><span class="key">Unit:</span> <br><span class="key">Type:</span> <br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/disk.yaml | 


### smb2_close_latency

Average latency for SMB2_COM_CLOSE operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `close_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> close_latency_base | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_close_latency_histogram

Latency histogram for SMB2_COM_CLOSE operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `close_latency_histogram`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_close_ops

Number of SMB2_COM_CLOSE operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `close_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_create_latency

Average latency for SMB2_COM_CREATE operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `create_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> create_latency_base | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_create_latency_histogram

Latency histogram for SMB2_COM_CREATE operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `create_latency_histogram`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_create_ops

Number of SMB2_COM_CREATE operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `create_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_lock_latency

Average latency for SMB2_COM_LOCK operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `lock_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> lock_latency_base | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_lock_latency_histogram

Latency histogram for SMB2_COM_LOCK operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `lock_latency_histogram`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_lock_ops

Number of SMB2_COM_LOCK operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `lock_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_negotiate_latency

Average latency for SMB2_COM_NEGOTIATE operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `negotiate_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> negotiate_latency_base | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_negotiate_ops

Number of SMB2_COM_NEGOTIATE operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `negotiate_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_oplock_break_latency

Average latency for SMB2_COM_OPLOCK_BREAK operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `oplock_break_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> oplock_break_latency_base | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_oplock_break_latency_histogram

Latency histogram for SMB2_COM_OPLOCK_BREAK operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `oplock_break_latency_histogram`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_oplock_break_ops

Number of SMB2_COM_OPLOCK_BREAK operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `oplock_break_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_query_directory_latency

Average latency for SMB2_COM_QUERY_DIRECTORY operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `query_directory_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> query_directory_latency_base | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_query_directory_latency_histogram

Latency histogram for SMB2_COM_QUERY_DIRECTORY operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `query_directory_latency_histogram`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_query_directory_ops

Number of SMB2_COM_QUERY_DIRECTORY operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `query_directory_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_query_info_latency

Average latency for SMB2_COM_QUERY_INFO operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `query_info_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> query_info_latency_base | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_query_info_latency_histogram

Latency histogram for SMB2_COM_QUERY_INFO operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `query_info_latency_histogram`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_query_info_ops

Number of SMB2_COM_QUERY_INFO operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `query_info_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_read_latency

Average latency for SMB2_COM_READ operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_read_ops

Number of SMB2_COM_READ operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_session_setup_latency

Average latency for SMB2_COM_SESSION_SETUP operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `session_setup_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> session_setup_latency_base | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_session_setup_latency_histogram

Latency histogram for SMB2_COM_SESSION_SETUP operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `session_setup_latency_histogram`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_session_setup_ops

Number of SMB2_COM_SESSION_SETUP operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `session_setup_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_set_info_latency

Average latency for SMB2_COM_SET_INFO operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `set_info_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> set_info_latency_base | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_set_info_latency_histogram

Latency histogram for SMB2_COM_SET_INFO operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `set_info_latency_histogram`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_set_info_ops

Number of SMB2_COM_SET_INFO operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `set_info_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_tree_connect_latency

Average latency for SMB2_COM_TREE_CONNECT operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `tree_connect_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> tree_connect_latency_base | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_tree_connect_ops

Number of SMB2_COM_TREE_CONNECT operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `tree_connect_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_write_latency

Average latency for SMB2_COM_WRITE operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_latency_base | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### smb2_write_ops

Number of SMB2_COM_WRITE operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances smb2` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/smb2.yaml | 


### snapmirror_break_failed_count

The number of failed SnapMirror break operations for the relationship

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `break_failed_count` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.break-failed-count` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_break_successful_count

The number of successful SnapMirror break operations for the relationship

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `break_successful_count` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.break-successful-count` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_lag_time

Amount of time since the last snapmirror transfer in seconds

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `lag_time` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.lag-time` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_last_transfer_duration

Duration of the last SnapMirror transfer in seconds

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `last_transfer_duration` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.last-transfer-duration` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_last_transfer_end_timestamp

The Timestamp of the end of the last transfer

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `last_transfer_end_timestamp` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.last-transfer-end-timestamp` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_last_transfer_size

Size in kilobytes (1024 bytes) of the last transfer

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `last_transfer_size` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.last-transfer-size` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_newest_snapshot_timestamp

The timestamp of the newest Snapshot copy on the destination volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `newest_snapshot_timestamp` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.newest-snapshot-timestamp` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_resync_failed_count

The number of failed SnapMirror resync operations for the relationship

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `resync_failed_count` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.resync-failed-count` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_resync_successful_count

The number of successful SnapMirror resync operations for the relationship

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `resync_successful_count` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.resync-successful-count` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_total_transfer_bytes

Cumulative bytes transferred for the relationship

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `total_transfer_bytes` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.total-transfer-bytes` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_total_transfer_time_secs

Cumulative total transfer time in seconds for the relationship

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `total_transfer_time_secs` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.total-transfer-time-secs` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_update_failed_count

The number of successful SnapMirror update operations for the relationship

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `update_failed_count` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.update-failed-count` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapmirror_update_successful_count

Number of Successful Updates

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapmirror` | `update_successful_count` | conf/rest/9.12.0/snapmirror.yaml |
| ZAPI | `snapmirror-get-iter` | `snapmirror-info.update-successful-count` | conf/zapi/cdot/9.8.0/snapmirror.yaml |


### snapshot_policy_total_schedules

Total Number of Schedules in this Policy

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/private/cli/snapshot/policy` | `total_schedules` | conf/rest/9.12.0/snapshotpolicy.yaml |
| ZAPI | `snapshot-policy-get-iter` | `snapshot-policy-info.total-schedules` | conf/zapi/cdot/9.8.0/snapshotpolicy.yaml |


### svm_cifs_connections

Number of connections

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs` | `connections`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_svm.yaml | 
| ZAPI | `perf-object-get-instances cifs:vserver` | `connections`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml | 


### svm_cifs_established_sessions

Number of established SMB and SMB2 sessions

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs` | `established_sessions`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_svm.yaml | 
| ZAPI | `perf-object-get-instances cifs:vserver` | `established_sessions`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml | 


### svm_cifs_latency

Average latency for CIFS operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs` | `latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> latency_base | conf/restperf/9.12.0/cifs_svm.yaml | 
| ZAPI | `perf-object-get-instances cifs:vserver` | `cifs_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs_latency_base | conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml | 


### svm_cifs_op_count

Array of select CIFS operation counts

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs` | `op_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_svm.yaml | 
| ZAPI | `perf-object-get-instances cifs:vserver` | `cifs_op_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml | 


### svm_cifs_open_files

Number of open files over SMB and SMB2

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs` | `open_files`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_svm.yaml | 
| ZAPI | `perf-object-get-instances cifs:vserver` | `open_files`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml | 


### svm_cifs_ops

Total number of CIFS operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_svm.yaml | 
| ZAPI | `perf-object-get-instances cifs:vserver` | `cifs_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml | 


### svm_cifs_read_latency

Average latency for CIFS read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs` | `average_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_read_ops | conf/restperf/9.12.0/cifs_svm.yaml | 
| ZAPI | `perf-object-get-instances cifs:vserver` | `cifs_read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs_read_ops | conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml | 


### svm_cifs_read_ops

Total number of CIFS read operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs` | `total_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_svm.yaml | 
| ZAPI | `perf-object-get-instances cifs:vserver` | `cifs_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml | 


### svm_cifs_signed_sessions

Number of signed SMB and SMB2 sessions.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs` | `signed_sessions`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_svm.yaml | 
| ZAPI | `perf-object-get-instances cifs:vserver` | `signed_sessions`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml | 


### svm_cifs_write_latency

Average latency for CIFS write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs` | `average_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_write_ops | conf/restperf/9.12.0/cifs_svm.yaml | 
| ZAPI | `perf-object-get-instances cifs:vserver` | `cifs_write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> cifs_write_ops | conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml | 


### svm_cifs_write_ops

Total number of CIFS write operations

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_cifs` | `total_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/cifs_svm.yaml | 
| ZAPI | `perf-object-get-instances cifs:vserver` | `cifs_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml | 


### svm_nfs_access_avg_latency

Average latency of NFSv4.2 ACCESS operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `access.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> access.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `access_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> access_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_access_total

Total number of NFSv4.2 ACCESS operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `access.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `access_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_backchannel_ctl_avg_latency

Average latency of NFSv4.2 BACKCHANNEL_CTL operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `backchannel_ctl.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> backchannel_ctl.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `backchannel_ctl_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> backchannel_ctl_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_backchannel_ctl_total

Total number of NFSv4.2 BACKCHANNEL_CTL operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `backchannel_ctl.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `backchannel_ctl_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_bind_conn_to_session_avg_latency

Average latency of NFSv4.2 BIND_CONN_TO_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `bind_conn_to_session.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> bind_conn_to_session.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `bind_conn_to_session_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> bind_conn_to_session_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_bind_conn_to_session_total

Total number of NFSv4.2 BIND_CONN_TO_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `bind_conn_to_session.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `bind_conn_to_session_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_close_avg_latency

Average latency of NFSv4.2 CLOSE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `close.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> close.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `close_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> close_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_close_total

Total number of NFSv4.2 CLOSE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `close.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `close_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_commit_avg_latency

Average latency of NFSv4.2 COMMIT operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `commit.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> commit.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `commit_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> commit_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_commit_total

Total number of NFSv4.2 COMMIT operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `commit.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `commit_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_create_avg_latency

Average latency of NFSv4.2 CREATE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `create.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> create.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `create_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> create_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_create_session_avg_latency

Average latency of NFSv4.2 CREATE_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `create_session.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> create_session.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `create_session_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> create_session_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_create_session_total

Total number of NFSv4.2 CREATE_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `create_session.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `create_session_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_create_total

Total number of NFSv4.2 CREATE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `create.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `create_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_delegpurge_avg_latency

Average latency of NFSv4.2 DELEGPURGE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `delegpurge.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> delegpurge.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `delegpurge_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> delegpurge_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_delegpurge_total

Total number of NFSv4.2 DELEGPURGE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `delegpurge.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `delegpurge_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_delegreturn_avg_latency

Average latency of NFSv4.2 DELEGRETURN operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `delegreturn.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> delegreturn.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `delegreturn_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> delegreturn_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_delegreturn_total

Total number of NFSv4.2 DELEGRETURN operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `delegreturn.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `delegreturn_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_destroy_clientid_avg_latency

Average latency of NFSv4.2 DESTROY_CLIENTID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `destroy_clientid.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> destroy_clientid.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `destroy_clientid_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> destroy_clientid_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_destroy_clientid_total

Total number of NFSv4.2 DESTROY_CLIENTID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `destroy_clientid.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `destroy_clientid_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_destroy_session_avg_latency

Average latency of NFSv4.2 DESTROY_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `destroy_session.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> destroy_session.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `destroy_session_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> destroy_session_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_destroy_session_total

Total number of NFSv4.2 DESTROY_SESSION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `destroy_session.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `destroy_session_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_exchange_id_avg_latency

Average latency of NFSv4.2 EXCHANGE_ID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `exchange_id.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> exchange_id.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `exchange_id_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> exchange_id_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_exchange_id_total

Total number of NFSv4.2 EXCHANGE_ID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `exchange_id.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `exchange_id_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_free_stateid_avg_latency

Average latency of NFSv4.2 FREE_STATEID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `free_stateid.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> free_stateid.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `free_stateid_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> free_stateid_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_free_stateid_total

Total number of NFSv4.2 FREE_STATEID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `free_stateid.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `free_stateid_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_fsinfo_avg_latency

Average latency of FSInfo procedure requests. The counter keeps track of the average response time of FSInfo requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `fsinfo.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> fsinfo.total | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `fsinfo_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> fsinfo_total | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_fsinfo_total

Total number FSInfo of procedure requests. It is the total number of FSInfo success and FSInfo error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `fsinfo.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `fsinfo_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_fsstat_avg_latency

Average latency of FSStat procedure requests. The counter keeps track of the average response time of FSStat requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `fsstat.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> fsstat.total | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `fsstat_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> fsstat_total | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_fsstat_total

Total number FSStat of procedure requests. It is the total number of FSStat success and FSStat error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `fsstat.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `fsstat_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_get_dir_delegation_avg_latency

Average latency of NFSv4.2 GET_DIR_DELEGATION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `get_dir_delegation.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> get_dir_delegation.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `get_dir_delegation_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> get_dir_delegation_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_get_dir_delegation_total

Total number of NFSv4.2 GET_DIR_DELEGATION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `get_dir_delegation.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `get_dir_delegation_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_getattr_avg_latency

Average latency of NFSv4.2 GETATTR operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `getattr.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> getattr.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `getattr_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> getattr_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_getattr_total

Total number of NFSv4.2 GETATTR operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `getattr.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `getattr_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_getdeviceinfo_avg_latency

Average latency of NFSv4.2 GETDEVICEINFO operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `getdeviceinfo.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> getdeviceinfo.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `getdeviceinfo_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> getdeviceinfo_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_getdeviceinfo_total

Total number of NFSv4.2 GETDEVICEINFO operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `getdeviceinfo.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `getdeviceinfo_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_getdevicelist_avg_latency

Average latency of NFSv4.2 GETDEVICELIST operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `getdevicelist.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> getdevicelist.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `getdevicelist_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> getdevicelist_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_getdevicelist_total

Total number of NFSv4.2 GETDEVICELIST operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `getdevicelist.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `getdevicelist_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_getfh_avg_latency

Average latency of NFSv4.2 GETFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `getfh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> getfh.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `getfh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> getfh_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_getfh_total

Total number of NFSv4.2 GETFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `getfh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `getfh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_latency

Average latency of nfsv42 requests. This counter keeps track of the average response time of nfsv42 requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> total_ops | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_latency_hist

Histogram of latency for NFSv3 operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances nfsv3` | `nfsv3_latency_hist`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_layoutcommit_avg_latency

Average latency of NFSv4.2 LAYOUTCOMMIT operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `layoutcommit.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> layoutcommit.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `layoutcommit_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> layoutcommit_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_layoutcommit_total

Total number of NFSv4.2 LAYOUTCOMMIT operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `layoutcommit.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `layoutcommit_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_layoutget_avg_latency

Average latency of NFSv4.2 LAYOUTGET operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `layoutget.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> layoutget.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `layoutget_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> layoutget_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_layoutget_total

Total number of NFSv4.2 LAYOUTGET operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `layoutget.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `layoutget_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_layoutreturn_avg_latency

Average latency of NFSv4.2 LAYOUTRETURN operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `layoutreturn.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> layoutreturn.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `layoutreturn_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> layoutreturn_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_layoutreturn_total

Total number of NFSv4.2 LAYOUTRETURN operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `layoutreturn.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `layoutreturn_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_link_avg_latency

Average latency of NFSv4.2 LINK operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `link.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> link.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `link_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> link_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_link_total

Total number of NFSv4.2 LINK operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `link.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `link_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_lock_avg_latency

Average latency of NFSv4.2 LOCK operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `lock.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> lock.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `lock_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> lock_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_lock_total

Total number of NFSv4.2 LOCK operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `lock.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `lock_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_lockt_avg_latency

Average latency of NFSv4.2 LOCKT operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `lockt.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> lockt.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `lockt_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> lockt_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_lockt_total

Total number of NFSv4.2 LOCKT operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `lockt.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `lockt_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_locku_avg_latency

Average latency of NFSv4.2 LOCKU operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `locku.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> locku.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `locku_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> locku_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_locku_total

Total number of NFSv4.2 LOCKU operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `locku.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `locku_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_lookup_avg_latency

Average latency of NFSv4.2 LOOKUP operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `lookup.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> lookup.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `lookup_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> lookup_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_lookup_total

Total number of NFSv4.2 LOOKUP operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `lookup.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `lookup_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_lookupp_avg_latency

Average latency of NFSv4.2 LOOKUPP operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `lookupp.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> lookupp.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `lookupp_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> lookupp_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_lookupp_total

Total number of NFSv4.2 LOOKUPP operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `lookupp.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `lookupp_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_mkdir_avg_latency

Average latency of MkDir procedure requests. The counter keeps track of the average response time of MkDir requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `mkdir.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> mkdir.total | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `mkdir_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> mkdir_total | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_mkdir_total

Total number MkDir of procedure requests. It is the total number of MkDir success and MkDir error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `mkdir.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `mkdir_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_mknod_avg_latency

Average latency of MkNod procedure requests. The counter keeps track of the average response time of MkNod requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `mknod.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> mknod.total | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `mknod_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> mknod_total | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_mknod_total

Total number MkNod of procedure requests. It is the total number of MkNod success and MkNod error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `mknod.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `mknod_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_null_avg_latency

Average latency of NFSv4.2 NULL procedures.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `null.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> null.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `null_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> null_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_null_total

Total number of NFSv4.2 NULL procedures.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `null.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `null_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_nverify_avg_latency

Average latency of NFSv4.2 NVERIFY operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `nverify.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> nverify.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `nverify_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> nverify_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_nverify_total

Total number of NFSv4.2 NVERIFY operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `nverify.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `nverify_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_open_avg_latency

Average latency of NFSv4.2 OPEN operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `open.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> open.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `open_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> open_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_open_confirm_avg_latency

Average latency of OPEN_CONFIRM procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4` | `open_confirm.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> open_confirm.total | conf/restperf/9.12.0/nfsv4.yaml | 
| ZAPI | `perf-object-get-instances nfsv4` | `open_confirm_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> open_confirm_total | conf/zapiperf/cdot/9.8.0/nfsv4.yaml | 


### svm_nfs_open_confirm_total

Total number of OPEN_CONFIRM procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4` | `open_confirm.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4.yaml | 
| ZAPI | `perf-object-get-instances nfsv4` | `open_confirm_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4.yaml | 


### svm_nfs_open_downgrade_avg_latency

Average latency of NFSv4.2 OPEN_DOWNGRADE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `open_downgrade.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> open_downgrade.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `open_downgrade_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> open_downgrade_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_open_downgrade_total

Total number of NFSv4.2 OPEN_DOWNGRADE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `open_downgrade.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `open_downgrade_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_open_total

Total number of NFSv4.2 OPEN operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `open.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `open_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_openattr_avg_latency

Average latency of NFSv4.2 OPENATTR operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `openattr.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> openattr.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `openattr_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> openattr_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_openattr_total

Total number of NFSv4.2 OPENATTR operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `openattr.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `openattr_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_ops

Total number of nfsv42 requests per sec.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_pathconf_avg_latency

Average latency of PathConf procedure requests. The counter keeps track of the average response time of PathConf requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `pathconf.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> pathconf.total | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `pathconf_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> pathconf_total | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_pathconf_total

Total number PathConf of procedure requests. It is the total number of PathConf success and PathConf error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `pathconf.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `pathconf_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_putfh_avg_latency

Average latency of NFSv4.2 PUTFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `putfh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> putfh.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `putfh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> putfh_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_putfh_total

Total number of NFSv4.2 PUTFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `putfh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `putfh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_putpubfh_avg_latency

Average latency of NFSv4.2 PUTPUBFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `putpubfh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> putpubfh.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `putpubfh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> putpubfh_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_putpubfh_total

Total number of NFSv4.2 PUTPUBFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `putpubfh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `putpubfh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_putrootfh_avg_latency

Average latency of NFSv4.2 PUTROOTFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `putrootfh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> putrootfh.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `putrootfh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> putrootfh_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_putrootfh_total

Total number of NFSv4.2 PUTROOTFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `putrootfh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `putrootfh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_read_avg_latency

Average latency of NFSv4.2 READ operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `read.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `read_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> read_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_read_latency_hist

Histogram of latency for Read operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances nfsv3` | `read_latency_hist`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_read_ops

Total observed NFSv3 read operations per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `nfsv3_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_read_symlink_avg_latency

Average latency of ReadSymLink procedure requests. The counter keeps track of the average response time of ReadSymLink requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `read_symlink.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_symlink.total | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `read_symlink_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> read_symlink_total | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_read_symlink_total

Total number of ReadSymLink procedure requests. It is the total number of read symlink success and read symlink error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `read_symlink.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `read_symlink_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_read_throughput

NFSv4.2 read data transfers.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `total.read_throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `nfs41_read_throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_read_total

Total number of NFSv4.2 READ operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `read.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `read_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_readdir_avg_latency

Average latency of NFSv4.2 READDIR operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `readdir.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> readdir.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `readdir_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> readdir_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_readdir_total

Total number of NFSv4.2 READDIR operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `readdir.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `readdir_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_readdirplus_avg_latency

Average latency of ReadDirPlus procedure requests. The counter keeps track of the average response time of ReadDirPlus requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `readdirplus.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> readdirplus.total | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `readdirplus_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> readdirplus_total | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_readdirplus_total

Total number ReadDirPlus of procedure requests. It is the total number of ReadDirPlus success and ReadDirPlus error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `readdirplus.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `readdirplus_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_readlink_avg_latency

Average latency of NFSv4.2 READLINK operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `readlink.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> readlink.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `readlink_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> readlink_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_readlink_total

Total number of NFSv4.2 READLINK operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `readlink.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `readlink_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_reclaim_complete_avg_latency

Average latency of NFSv4.2 RECLAIM_complete operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `reclaim_complete.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> reclaim_complete.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `reclaim_complete_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> reclaim_complete_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_reclaim_complete_total

Total number of NFSv4.2 RECLAIM_complete operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `reclaim_complete.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `reclaim_complete_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_release_lock_owner_avg_latency

Average Latency of RELEASE_LOCKOWNER procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4` | `release_lock_owner.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> release_lock_owner.total | conf/restperf/9.12.0/nfsv4.yaml | 
| ZAPI | `perf-object-get-instances nfsv4` | `release_lock_owner_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> release_lock_owner_total | conf/zapiperf/cdot/9.8.0/nfsv4.yaml | 


### svm_nfs_release_lock_owner_total

Total number of RELEASE_LOCKOWNER procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4` | `release_lock_owner.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4.yaml | 
| ZAPI | `perf-object-get-instances nfsv4` | `release_lock_owner_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4.yaml | 


### svm_nfs_remove_avg_latency

Average latency of NFSv4.2 REMOVE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `remove.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> remove.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `remove_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> remove_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_remove_total

Total number of NFSv4.2 REMOVE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `remove.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `remove_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_rename_avg_latency

Average latency of NFSv4.2 RENAME operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `rename.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> rename.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `rename_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> rename_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_rename_total

Total number of NFSv4.2 RENAME operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `rename.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `rename_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_renew_avg_latency

Average latency of RENEW procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4` | `renew.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> renew.total | conf/restperf/9.12.0/nfsv4.yaml | 
| ZAPI | `perf-object-get-instances nfsv4` | `renew_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> renew_total | conf/zapiperf/cdot/9.8.0/nfsv4.yaml | 


### svm_nfs_renew_total

Total number of RENEW procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4` | `renew.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4.yaml | 
| ZAPI | `perf-object-get-instances nfsv4` | `renew_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4.yaml | 


### svm_nfs_restorefh_avg_latency

Average latency of NFSv4.2 RESTOREFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `restorefh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> restorefh.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `restorefh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> restorefh_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_restorefh_total

Total number of NFSv4.2 RESTOREFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `restorefh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `restorefh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_rmdir_avg_latency

Average latency of RmDir procedure requests. The counter keeps track of the average response time of RmDir requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `rmdir.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> rmdir.total | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `rmdir_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> rmdir_total | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_rmdir_total

Total number RmDir of procedure requests. It is the total number of RmDir success and RmDir error requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `rmdir.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `rmdir_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_savefh_avg_latency

Average latency of NFSv4.2 SAVEFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `savefh.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> savefh.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `savefh_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> savefh_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_savefh_total

Total number of NFSv4.2 SAVEFH operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `savefh.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `savefh_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_secinfo_avg_latency

Average latency of NFSv4.2 SECINFO operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `secinfo.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> secinfo.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `secinfo_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> secinfo_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_secinfo_no_name_avg_latency

Average latency of NFSv4.2 SECINFO_NO_NAME operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `secinfo_no_name.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> secinfo_no_name.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `secinfo_no_name_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> secinfo_no_name_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_secinfo_no_name_total

Total number of NFSv4.2 SECINFO_NO_NAME operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `secinfo_no_name.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `secinfo_no_name_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_secinfo_total

Total number of NFSv4.2 SECINFO operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `secinfo.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `secinfo_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_sequence_avg_latency

Average latency of NFSv4.2 SEQUENCE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `sequence.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> sequence.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `sequence_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> sequence_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_sequence_total

Total number of NFSv4.2 SEQUENCE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `sequence.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `sequence_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_set_ssv_avg_latency

Average latency of NFSv4.2 SET_SSV operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `set_ssv.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> set_ssv.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `set_ssv_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> set_ssv_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_set_ssv_total

Total number of NFSv4.2 SET_SSV operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `set_ssv.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `set_ssv_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_setattr_avg_latency

Average latency of NFSv4.2 SETATTR operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `setattr.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> setattr.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `setattr_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> setattr_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_setattr_total

Total number of NFSv4.2 SETATTR operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `setattr.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `setattr_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_setclientid_avg_latency

Average latency of SETCLIENTID procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4` | `setclientid.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> setclientid.total | conf/restperf/9.12.0/nfsv4.yaml | 
| ZAPI | `perf-object-get-instances nfsv4` | `setclientid_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> setclientid_total | conf/zapiperf/cdot/9.8.0/nfsv4.yaml | 


### svm_nfs_setclientid_confirm_avg_latency

Average latency of SETCLIENTID_CONFIRM procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4` | `setclientid_confirm.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> setclientid_confirm.total | conf/restperf/9.12.0/nfsv4.yaml | 
| ZAPI | `perf-object-get-instances nfsv4` | `setclientid_confirm_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> setclientid_confirm_total | conf/zapiperf/cdot/9.8.0/nfsv4.yaml | 


### svm_nfs_setclientid_confirm_total

Total number of SETCLIENTID_CONFIRM procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4` | `setclientid_confirm.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4.yaml | 
| ZAPI | `perf-object-get-instances nfsv4` | `setclientid_confirm_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4.yaml | 


### svm_nfs_setclientid_total

Total number of SETCLIENTID procedures

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v4` | `setclientid.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4.yaml | 
| ZAPI | `perf-object-get-instances nfsv4` | `setclientid_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4.yaml | 


### svm_nfs_symlink_avg_latency

Average latency of SymLink procedure requests. The counter keeps track of the average response time of SymLink requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `symlink.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> symlink.total | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `symlink_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> symlink_total | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_symlink_total

Total number SymLink of procedure requests. It is the total number of SymLink success and create SymLink requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `symlink.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `symlink_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_test_stateid_avg_latency

Average latency of NFSv4.2 TEST_STATEID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `test_stateid.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> test_stateid.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `test_stateid_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> test_stateid_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_test_stateid_total

Total number of NFSv4.2 TEST_STATEID operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `test_stateid.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `test_stateid_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_throughput

NFSv4.2 write data transfers.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `total.write_throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `nfs41_throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_verify_avg_latency

Average latency of NFSv4.2 VERIFY operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `verify.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> verify.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `verify_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> verify_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_verify_total

Total number of NFSv4.2 VERIFY operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `verify.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `verify_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_want_delegation_avg_latency

Average latency of NFSv4.2 WANT_DELEGATION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `want_delegation.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> want_delegation.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `want_delegation_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> want_delegation_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_want_delegation_total

Total number of NFSv4.2 WANT_DELEGATION operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `want_delegation.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `want_delegation_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_write_avg_latency

Average latency of NFSv4.2 WRITE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `write.average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write.total | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `write_avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average,no-zero-values<br><span class="key">Base:</span> write_total | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_write_latency_hist

Histogram of latency for Write operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances nfsv3` | `write_latency_hist`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_write_ops

Total observed NFSv3 write operations per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v3` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv3.yaml | 
| ZAPI | `perf-object-get-instances nfsv3` | `nfsv3_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv3.yaml | 


### svm_nfs_write_throughput

NFSv4.2 data transfers.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `total.throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `nfs41_write_throughput`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate,no-zero-values<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_nfs_write_total

Total number of NFSv4.2 WRITE operations.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/svm_nfs_v42` | `write.total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/nfsv4_2.yaml | 
| ZAPI | `perf-object-get-instances nfsv4_1` | `write_total`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml | 


### svm_vol_avg_latency

Average latency in microseconds for the WAFL filesystem to process all the operations on the volume; not including request processing or network communication time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:svm` | `average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/restperf/9.12.0/volume_svm.yaml | 
| ZAPI | `perf-object-get-instances volume:vserver` | `avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/zapiperf/cdot/9.8.0/volume_svm.yaml | 


### svm_vol_other_latency

Average latency in microseconds for the WAFL filesystem to process other operations to the volume; not including request processing or network communication time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:svm` | `other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_other_ops | conf/restperf/9.12.0/volume_svm.yaml | 
| ZAPI | `perf-object-get-instances volume:vserver` | `other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> other_ops | conf/zapiperf/cdot/9.8.0/volume_svm.yaml | 


### svm_vol_other_ops

Number of other operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:svm` | `total_other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_svm.yaml | 
| ZAPI | `perf-object-get-instances volume:vserver` | `other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_svm.yaml | 


### svm_vol_read_data

Bytes read per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:svm` | `bytes_read`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_svm.yaml | 
| ZAPI | `perf-object-get-instances volume:vserver` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_svm.yaml | 


### svm_vol_read_latency

Average latency in microseconds for the WAFL filesystem to process read request to the volume; not including request processing or network communication time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:svm` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_read_ops | conf/restperf/9.12.0/volume_svm.yaml | 
| ZAPI | `perf-object-get-instances volume:vserver` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.8.0/volume_svm.yaml | 


### svm_vol_read_ops

Number of read operations per second from the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:svm` | `total_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_svm.yaml | 
| ZAPI | `perf-object-get-instances volume:vserver` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_svm.yaml | 


### svm_vol_total_ops

Number of operations per second serviced by the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:svm` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_svm.yaml | 
| ZAPI | `perf-object-get-instances volume:vserver` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_svm.yaml | 


### svm_vol_write_data

Bytes written per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:svm` | `bytes_written`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_svm.yaml | 
| ZAPI | `perf-object-get-instances volume:vserver` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_svm.yaml | 


### svm_vol_write_latency

Average latency in microseconds for the WAFL filesystem to process write request to the volume; not including request processing or network communication time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:svm` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_write_ops | conf/restperf/9.12.0/volume_svm.yaml | 
| ZAPI | `perf-object-get-instances volume:vserver` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/zapiperf/cdot/9.8.0/volume_svm.yaml | 


### svm_vol_write_ops

Number of write operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume:svm` | `total_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume_svm.yaml | 
| ZAPI | `perf-object-get-instances volume:vserver` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume_svm.yaml | 


### svm_vscan_connections_active

Total number of current active connections

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances offbox_vscan` | `connections_active`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/vscan_svm.yaml | 


### svm_vscan_dispatch_latency

Average dispatch latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances offbox_vscan` | `dispatch_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> dispatch_latency_base | conf/zapiperf/cdot/9.8.0/vscan_svm.yaml | 


### svm_vscan_scan_latency

Average scan latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances offbox_vscan` | `scan_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> scan_latency_base | conf/zapiperf/cdot/9.8.0/vscan_svm.yaml | 


### svm_vscan_scan_noti_received_rate

Total number of scan notifications received by the dispatcher per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances offbox_vscan` | `scan_noti_received_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/vscan_svm.yaml | 


### svm_vscan_scan_request_dispatched_rate

Total number of scan requests sent to the Vscanner per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances offbox_vscan` | `scan_request_dispatched_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/vscan_svm.yaml | 


### token_copy_bytes

Total number of bytes copied.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/token_manager` | `token_copy.bytes`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/token_manager.yaml | 
| ZAPI | `perf-object-get-instances token_manager` | `token_copy_bytes`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/token_manager.yaml | 


### token_copy_failure

Number of failed token copy requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/token_manager` | `token_copy.failures`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/token_manager.yaml | 
| ZAPI | `perf-object-get-instances token_manager` | `token_copy_failure`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/token_manager.yaml | 


### token_copy_success

Number of successful token copy requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/token_manager` | `token_copy.successes`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/token_manager.yaml | 
| ZAPI | `perf-object-get-instances token_manager` | `token_copy_success`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/token_manager.yaml | 


### token_create_bytes

Total number of bytes for which tokens are created.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/token_manager` | `token_create.bytes`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/token_manager.yaml | 
| ZAPI | `perf-object-get-instances token_manager` | `token_create_bytes`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/token_manager.yaml | 


### token_create_failure

Number of failed token create requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/token_manager` | `token_create.failures`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/token_manager.yaml | 
| ZAPI | `perf-object-get-instances token_manager` | `token_create_failure`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/token_manager.yaml | 


### token_create_success

Number of successful token create requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/token_manager` | `token_create.successes`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/token_manager.yaml | 
| ZAPI | `perf-object-get-instances token_manager` | `token_create_success`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/token_manager.yaml | 


### token_zero_bytes

Total number of bytes zeroed.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/token_manager` | `token_zero.bytes`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/token_manager.yaml | 
| ZAPI | `perf-object-get-instances token_manager` | `token_zero_bytes`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/token_manager.yaml | 


### token_zero_failure

Number of failed token zero requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/token_manager` | `token_zero.failures`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/token_manager.yaml | 
| ZAPI | `perf-object-get-instances token_manager` | `token_zero_failure`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/token_manager.yaml | 


### token_zero_success

Number of successful token zero requests.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/token_manager` | `token_zero.successes`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/token_manager.yaml | 
| ZAPI | `perf-object-get-instances token_manager` | `token_zero_success`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/token_manager.yaml | 


### volume_autosize_grow_threshold_percent

Used space threshold size, in percentage, for the automatic growth of the volume. When the amount of used space in the volume becomes greater than this threhold, the volume automatically grows unless it has reached the maximum size. The volume grows when 'space.used' is greater than this percent of 'space.size'. The 'grow_threshold' size cannot be less than or equal to the 'shrink_threshold' size..

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `autosize.grow_threshold` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-autosize-attributes.grow-threshold-percent` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_autosize_maximum_size

Maximum size in bytes up to which a volume grows automatically. This size cannot be less than the current volume size, or less than or equal to the minimum size of volume.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `autosize.maximum` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-autosize-attributes.maximum-size` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_avg_latency

Average latency in microseconds for the WAFL filesystem to process all the operations on the volume; not including request processing or network communication time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume` | `average_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/restperf/9.12.0/volume.yaml | 
| ZAPI | `perf-object-get-instances volume` | `avg_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_ops | conf/zapiperf/cdot/9.8.0/volume.yaml | 


### volume_filesystem_size

Total usable size of the volume, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.filesystem_size` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.filesystem-size` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_inode_files_total

Total user-visible file (inode) count, i.e., current maximum number of user-visible files (inodes) that this volume can currently hold.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `volume-get-iter` | `volume-attributes.volume-inode-attributes.files-total` | conf/zapi/cdot/9.8.0/volume.yaml |
| REST | `api/private/cli/volume` | `files` | conf/rest/9.12.0/volume.yaml |


### volume_inode_files_used

Number of user-visible files (inodes) used. This field is valid only when the volume is online.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `volume-get-iter` | `volume-attributes.volume-inode-attributes.files-used` | conf/zapi/cdot/9.8.0/volume.yaml |
| REST | `api/private/cli/volume` | `files_used` | conf/rest/9.12.0/volume.yaml |


### volume_inode_used_percent

volume_inode_files_used / volume_inode_total

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_other_latency

Average latency in microseconds for the WAFL filesystem to process other operations to the volume; not including request processing or network communication time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume` | `other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_other_ops | conf/restperf/9.12.0/volume.yaml | 
| ZAPI | `perf-object-get-instances volume` | `other_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> other_ops | conf/zapiperf/cdot/9.8.0/volume.yaml | 


### volume_other_ops

Number of other operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume` | `total_other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume.yaml | 
| ZAPI | `perf-object-get-instances volume` | `other_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume.yaml | 


### volume_overwrite_reserve_available

amount of storage space that is currently available for overwrites, calculated by subtracting the total amount of overwrite reserve space from the amount that has already been used.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_overwrite_reserve_total

Reserved space for overwrites, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.overwrite_reserve` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.overwrite-reserve` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_overwrite_reserve_used

Overwrite logical reserve space used, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.overwrite_reserve_used` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.overwrite-reserve-used` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_read_data

Bytes read per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume` | `bytes_read`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume.yaml | 
| ZAPI | `perf-object-get-instances volume` | `read_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume.yaml | 


### volume_read_latency

Average latency in microseconds for the WAFL filesystem to process read request to the volume; not including request processing or network communication time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_read_ops | conf/restperf/9.12.0/volume.yaml | 
| ZAPI | `perf-object-get-instances volume` | `read_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> read_ops | conf/zapiperf/cdot/9.8.0/volume.yaml | 


### volume_read_ops

Number of read operations per second from the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume` | `total_read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume.yaml | 
| ZAPI | `perf-object-get-instances volume` | `read_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume.yaml | 


### volume_sis_compress_saved

The total disk space (in bytes) that is saved by compressing blocks on the referenced file system.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `volume-get-iter` | `volume-attributes.volume-sis-attributes.compression-space-saved` | conf/zapi/cdot/9.8.0/volume.yaml |
| REST | `api/private/cli/volume` | `compression_space_saved` | conf/rest/9.12.0/volume.yaml |


### volume_sis_compress_saved_percent

Percentage of the total disk space that is saved by compressing blocks on the referenced file system

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `volume-get-iter` | `volume-attributes.volume-sis-attributes.percentage-compression-space-saved` | conf/zapi/cdot/9.8.0/volume.yaml |
| REST | `api/private/cli/volume` | `compression_space_saved_percent` | conf/rest/9.12.0/volume.yaml |


### volume_sis_dedup_saved

The total disk space (in bytes) that is saved by deduplication and file cloning.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `volume-get-iter` | `volume-attributes.volume-sis-attributes.deduplication-space-saved` | conf/zapi/cdot/9.8.0/volume.yaml |
| REST | `api/private/cli/volume` | `dedupe_space_saved` | conf/rest/9.12.0/volume.yaml |


### volume_sis_dedup_saved_percent

Percentage of the total disk space that is saved by deduplication and file cloning.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `volume-get-iter` | `volume-attributes.volume-sis-attributes.percentage-deduplication-space-saved` | conf/zapi/cdot/9.8.0/volume.yaml |
| REST | `api/private/cli/volume` | `dedupe_space_saved_percent` | conf/rest/9.12.0/volume.yaml |


### volume_sis_total_saved

Total space saved (in bytes) in the volume due to deduplication, compression, and file cloning.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `volume-get-iter` | `volume-attributes.volume-sis-attributes.total-space-saved` | conf/zapi/cdot/9.8.0/volume.yaml |
| REST | `api/private/cli/volume` | `sis_space_saved` | conf/rest/9.12.0/volume.yaml |


### volume_sis_total_saved_percent

Percentage of total disk space that is saved by compressing blocks, deduplication and file cloning.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `volume-get-iter` | `volume-attributes.volume-sis-attributes.percentage-total-space-saved` | conf/zapi/cdot/9.8.0/volume.yaml |
| REST | `api/private/cli/volume` | `sis_space_saved_percent` | conf/rest/9.12.0/volume.yaml |


### volume_size

Total provisioned size. The default size is equal to the minimum size of 20MB, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.size` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.size` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_size_available

The available space, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.available` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.size-available` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_size_total

Total size of AFS, excluding snap-reserve, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.afs_total` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.size-total` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_size_used

The virtual space used (includes volume reserves) before storage efficiency, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.used` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.size-used` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_size_used_percent

percentage of utilized storage space in a volume relative to its total capacity

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.percent_used` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.percentage-size-used` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_snapshot_count

Number of Snapshot copies in the volume.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `snapshot_count` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-snapshot-attributes.snapshot-count` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_snapshot_reserve_available

Size available for Snapshot copies within the Snapshot copy reserve, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.snapshot.reserve_available` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.snapshot-reserve-available` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_snapshot_reserve_percent

The space that has been set aside as a reserve for Snapshot copy usage, in percent.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.snapshot.reserve_percent` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.percentage-snapshot-reserve` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_snapshot_reserve_size

Size in the volume that has been set aside as a reserve for Snapshot copy usage, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.snapshot.reserve_size` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.snapshot-reserve-size` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_snapshot_reserve_used

amount of storage space currently used by a volume's snapshot reserve, which is calculated by subtracting the snapshot reserve available space from the snapshot reserve size.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `NA` | `Harvest generated` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `NA` | `Harvest generated` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_snapshot_reserve_used_percent

Percentage of snapshot reserve size that has been used.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.snapshot.space_used_percent` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.percentage-snapshot-reserve-used` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_snapshots_size_available

Available space for Snapshot copies from snap-reserve, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.size_available_for_snapshots` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.size-available-for-snapshots` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_snapshots_size_used

The total space used by Snapshot copies in the volume, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.snapshot.used` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.size-used-by-snapshots` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_space_expected_available

Size that should be available for the volume, irrespective of available size in the aggregate, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.expected_available` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.expected-available` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_space_logical_available

The amount of space available in this volume with storage efficiency space considered used, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.logical_space.available` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.logical-available` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_space_logical_used

SUM of (physical-used, shared_refs, compression_saved_in_plane0, vbn_zero, future_blk_cnt), in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.logical_space.used` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.logical-used` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_space_logical_used_by_afs

The virtual space used by AFS alone (includes volume reserves) and along with storage efficiency, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.logical_space.used_by_afs` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.logical-used-by-afs` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_space_logical_used_by_snapshots

Size that is logically used across all Snapshot copies in the volume, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.logical_space.used_by_snapshots` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.logical-used-by-snapshots` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_space_logical_used_percent

SUM of (physical-used, shared_refs, compression_saved_in_plane0, vbn_zero, future_blk_cnt), as a percentage.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.logical_space.used_percent` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.logical-used-percent` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_space_physical_used

Size that is physically used in the volume, in bytes.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.physical_used` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.physical-used` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_space_physical_used_percent

Size that is physically used in the volume, as a percentage.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/storage/volumes` | `space.physical_used_percent` | conf/rest/9.12.0/volume.yaml |
| ZAPI | `volume-get-iter` | `volume-attributes.volume-space-attributes.physical-used-percent` | conf/zapi/cdot/9.8.0/volume.yaml |


### volume_total_ops

Number of operations per second serviced by the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume.yaml | 
| ZAPI | `perf-object-get-instances volume` | `total_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume.yaml | 


### volume_write_data

Bytes written per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume` | `bytes_written`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume.yaml | 
| ZAPI | `perf-object-get-instances volume` | `write_data`<br><span class="key">Unit:</span> b_per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume.yaml | 


### volume_write_latency

Average latency in microseconds for the WAFL filesystem to process write request to the volume; not including request processing or network communication time

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> total_write_ops | conf/restperf/9.12.0/volume.yaml | 
| ZAPI | `perf-object-get-instances volume` | `write_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> write_ops | conf/zapiperf/cdot/9.8.0/volume.yaml | 


### volume_write_ops

Number of write operations per second to the volume

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/volume` | `total_write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/volume.yaml | 
| ZAPI | `perf-object-get-instances volume` | `write_ops`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/volume.yaml | 


### vscan_scan_latency

Average scan latency

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances offbox_vscan_server` | `scan_latency`<br><span class="key">Unit:</span> microsec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> scan_latency_base | conf/zapiperf/cdot/9.8.0/vscan.yaml | 


### vscan_scan_request_dispatched_rate

Total number of scan requests sent to the Vscanner per second

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances offbox_vscan_server` | `scan_request_dispatched_rate`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/vscan.yaml | 


### vscan_scanner_stats_pct_cpu_used

Percentage CPU utilization on scanner

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances offbox_vscan_server` | `scanner_stats_pct_cpu_used`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/vscan.yaml | 


### vscan_scanner_stats_pct_mem_used

Percentage RAM utilization on scanner

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances offbox_vscan_server` | `scanner_stats_pct_mem_used`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/vscan.yaml | 


### vscan_scanner_stats_pct_network_used

Percentage network utilization on scanner

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances offbox_vscan_server` | `scanner_stats_pct_network_used`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/vscan.yaml | 


### wafl_avg_msg_latency

Average turnaround time for WAFL messages in milliseconds.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `average_msg_latency`<br><span class="key">Unit:</span> millisec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> msg_total | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `avg_wafl_msg_latency`<br><span class="key">Unit:</span> millisec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> wafl_msg_total | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_avg_non_wafl_msg_latency

Average turnaround time for non-WAFL messages in milliseconds.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `average_non_wafl_msg_latency`<br><span class="key">Unit:</span> millisec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> non_wafl_msg_total | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `avg_non_wafl_msg_latency`<br><span class="key">Unit:</span> millisec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> non_wafl_msg_total | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_avg_repl_msg_latency

Average turnaround time for replication WAFL messages in milliseconds.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `average_replication_msg_latency`<br><span class="key">Unit:</span> millisec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> replication_msg_total | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `avg_wafl_repl_msg_latency`<br><span class="key">Unit:</span> millisec<br><span class="key">Type:</span> average<br><span class="key">Base:</span> wafl_repl_msg_total | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_cp_count

Array of counts of different types of Consistency Points (CP).

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `cp_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `cp_count`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_cp_phase_times

Array of percentage time spent in different phases of Consistency Point (CP).

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `cp_phase_times`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> total_cp_msecs | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `cp_phase_times`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> total_cp_msecs | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_memory_free

The current WAFL memory available in the system.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `memory_free`<br><span class="key">Unit:</span> mb<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `wafl_memory_free`<br><span class="key">Unit:</span> mb<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_memory_used

The current WAFL memory used in the system.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `memory_used`<br><span class="key">Unit:</span> mb<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `wafl_memory_used`<br><span class="key">Unit:</span> mb<br><span class="key">Type:</span> raw<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_msg_total

Total number of WAFL messages per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `msg_total`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `wafl_msg_total`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_non_wafl_msg_total

Total number of non-WAFL messages per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `non_wafl_msg_total`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `non_wafl_msg_total`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_read_io_type

Percentage of reads served from buffer cache, external cache, or disk.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `read_io_type`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_io_type_base | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `read_io_type`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> read_io_type_base | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_reads_from_cache

WAFL reads from cache.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `reads_from_cache`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `wafl_reads_from_cache`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_reads_from_cloud

WAFL reads from cloud storage.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `reads_from_cloud`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `wafl_reads_from_cloud`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_reads_from_cloud_s2c_bin

WAFL reads from cloud storage via s2c bin.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `reads_from_cloud_s2c_bin`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `wafl_reads_from_cloud_s2c_bin`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_reads_from_disk

WAFL reads from disk.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `reads_from_disk`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `wafl_reads_from_disk`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_reads_from_ext_cache

WAFL reads from external cache.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `reads_from_external_cache`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `wafl_reads_from_ext_cache`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_reads_from_fc_miss

WAFL reads from remote volume for fc_miss.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `reads_from_fc_miss`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `wafl_reads_from_fc_miss`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_reads_from_pmem

Wafl reads from persistent mmeory.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| ZAPI | `perf-object-get-instances wafl` | `wafl_reads_from_pmem`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_reads_from_ssd

WAFL reads from SSD.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `reads_from_ssd`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `wafl_reads_from_ssd`<br><span class="key">Unit:</span> none<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_repl_msg_total

Total number of replication WAFL messages per second.

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `replication_msg_total`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `wafl_repl_msg_total`<br><span class="key">Unit:</span> per_sec<br><span class="key">Type:</span> rate<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_total_cp_msecs

Milliseconds spent in Consistency Point (CP).

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `total_cp_msecs`<br><span class="key">Unit:</span> millisec<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `total_cp_msecs`<br><span class="key">Unit:</span> millisec<br><span class="key">Type:</span> delta<br><span class="key">Base:</span>  | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


### wafl_total_cp_util

Percentage of time spent in a Consistency Point (CP).

| API    | Endpoint | Metric | Template |
|--------|----------|--------|---------|
| REST | `api/cluster/counter/tables/wafl` | `total_cp_util`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> cpu_elapsed_time | conf/restperf/9.12.0/wafl.yaml | 
| ZAPI | `perf-object-get-instances wafl` | `total_cp_util`<br><span class="key">Unit:</span> percent<br><span class="key">Type:</span> percent<br><span class="key">Base:</span> cpu_elapsed_time | conf/zapiperf/cdot/9.8.0/wafl.yaml | 


