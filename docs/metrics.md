```yaml

Metrics:
  - Harvest Metric: aggr_efficiency_savings
    Description: Space saved by storage efficiencies (logical_used - used)
    REST:
      endpoint: api/storage/aggregates
      metric: space.efficiency.savings
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_efficiency_savings_wo_snapshots
    Description: Space saved by storage efficiencies (logical_used - used)
    REST:
      endpoint: api/storage/aggregates
      metric: space.efficiency_without_snapshots.savings
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_efficiency_savings_wo_snapshots_flexclones
    Description: Space saved by storage efficiencies (logical_used - used)
    REST:
      endpoint: api/storage/aggregates
      metric: space.efficiency_without_snapshots_flexclones.savings
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_hybrid_cache_size_total
    Description: Total usable space in bytes of SSD cache. Only provided when hybrid_cache.enabled is 'true'.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.hybrid-cache-size-total
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: block_storage.hybrid_cache.size
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_hybrid_disk_count
    Description: Number of disks used in the cache tier of the aggregate. Only provided when hybrid_cache.enabled is 'true'.
    REST:
      endpoint: api/storage/aggregates
      metric: block_storage.hybrid_cache.disk_count
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_inode_files_private_used
    Description: Number of system metadata files used. If the referenced file system is restricted or offline, a value of 0 is returned.This is an advanced property; there is an added computational cost to retrieving its value. The field is not populated for either a collection GET or an instance GET unless it is explicitly requested using the <i>fields</i> query parameter containing either footprint or **.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-inode-attributes.files-private-used
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: inode_attributes.files_private_used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_inode_files_total
    Description: Maximum number of user-visible files that this referenced file system can currently hold. If the referenced file system is restricted or offline, a value of 0 is returned.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-inode-attributes.files-total
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: inode_attributes.files_total
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_inode_files_used
    Description: Number of user-visible files used in the referenced file system. If the referenced file system is restricted or offline, a value of 0 is returned.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-inode-attributes.files-used
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: inode_attributes.files_used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_inode_inodefile_private_capacity
    Description: Number of files that can currently be stored on disk for system metadata files. This number will dynamically increase as more system files are created.This is an advanced property; there is an added computationl cost to retrieving its value. The field is not populated for either a collection GET or an instance GET unless it is explicitly requested using the <i>fields</i> query parameter containing either footprint or **.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-inode-attributes.inodefile-private-capacity
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: inode_attributes.file_private_capacity
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_inode_inodefile_public_capacity
    Description: Number of files that can currently be stored on disk for user-visible files.  This number will dynamically increase as more user-visible files are created.This is an advanced property; there is an added computational cost to retrieving its value. The field is not populated for either a collection GET or an instance GET unless it is explicitly requested using the <i>fields</i> query parameter containing either footprint or **.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-inode-attributes.inodefile-public-capacity
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: inode_attributes.file_public_capacity
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_inode_maxfiles_available
    Description: The count of the maximum number of user-visible files currently allowable on the referenced file system.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-inode-attributes.maxfiles-available
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: inode_attributes.max_files_available
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_inode_maxfiles_possible
    Description: The largest value to which the maxfiles-available parameter can be increased by reconfiguration, on the referenced file system.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-inode-attributes.maxfiles-possible
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: inode_attributes.max_files_possible
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_inode_maxfiles_used
    Description: The number of user-visible files currently in use on the referenced file system.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-inode-attributes.maxfiles-used
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: inode_attributes.max_files_used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_inode_used_percent
    Description: The percentage of disk space currently in use based on user-visible file count on the referenced file system.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-inode-attributes.percent-inode-used-capacity
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: inode_attributes.used_percent
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_logical_used_wo_snapshots
    Description: Logical used
    ZAPI:
      endpoint: aggr-efficiency-get-iter
      metric: aggr-efficiency-info.aggr-efficiency-cumulative-info.total-data-reduction-logical-used-wo-snapshots
      template: conf/zapi/cdot/9.9.0/aggr_efficiency.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.efficiency_without_snapshots.logical_used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_logical_used_wo_snapshots_flexclones
    Description: Logical used
    ZAPI:
      endpoint: aggr-efficiency-get-iter
      metric: aggr-efficiency-info.aggr-efficiency-cumulative-info.total-data-reduction-logical-used-wo-snapshots-flexclones
      template: conf/zapi/cdot/9.9.0/aggr_efficiency.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.efficiency_without_snapshots_flexclones.logical_used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_physical_used_wo_snapshots
    Description: Total Data Reduction Physical Used Without Snapshots
    ZAPI:
      endpoint: aggr-efficiency-get-iter
      metric: aggr-efficiency-info.aggr-efficiency-cumulative-info.total-data-reduction-physical-used-wo-snapshots
      template: conf/zapi/cdot/9.9.0/aggr_efficiency.yaml
    REST:
      endpoint: NA
      metric: Harvest Plugin Generated
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_physical_used_wo_snapshots_flexclones
    Description: Total Data Reduction Physical Used without snapshots and flexclones
    ZAPI:
      endpoint: aggr-efficiency-get-iter
      metric: aggr-efficiency-info.aggr-efficiency-cumulative-info.total-data-reduction-physical-used-wo-snapshots-flexclones
      template: conf/zapi/cdot/9.9.0/aggr_efficiency.yaml
    REST:
      endpoint: NA
      metric: Harvest Plugin Generated
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_primary_disk_count
    Description: Number of disks used in the aggregate. This includes parity disks, but excludes disks in the hybrid cache.
    REST:
      endpoint: api/storage/aggregates
      metric: block_storage.primary.disk_count
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_raid_disk_count
    Description: Number of disks in the aggregate.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-raid-attributes.disk-count
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: NA
      metric: Harvest Plugin Generated
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_raid_plex_count
    Description: Number of plexes in the aggregate
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-raid-attributes.plex-count
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: block_storage.plexes.#
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_raid_size
    Description: Option to specify the maximum number of disks that can be included in a RAID group.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-raid-attributes.raid-size
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: block_storage.primary.raid_size
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_snapshot_files_total
    Description: Total files allowed in Snapshot copies
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-snapshot-attributes.files-total
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: snapshot.files_total
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_snapshot_files_used
    Description: Total files created in Snapshot copies
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-snapshot-attributes.files-used
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: snapshot.files_used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_snapshot_inode_used_percent
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-snapshot-attributes.percent-inode-used-capacity
      template: conf/zapi/cdot/9.8.0/aggr.yaml

  - Harvest Metric: aggr_snapshot_maxfiles_available
    Description: Maximum files available for Snapshot copies
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-snapshot-attributes.maxfiles-available
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: snapshot.max_files_available
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_snapshot_maxfiles_possible
    Description: The largest value to which the maxfiles-available parameter can be increased by reconfiguration, on the referenced file system.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-snapshot-attributes.maxfiles-possible
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: NA
      metric: Harvest Plugin Generated
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_snapshot_maxfiles_used
    Description: Files in use by Snapshot copies
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-snapshot-attributes.maxfiles-used
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: snapshot.max_files_used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_snapshot_reserve_percent
    Description: Percentage of space reserved for Snapshot copies
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-snapshot-attributes.snapshot-reserve-percent
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.snapshot.reserve_percent
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_snapshot_size_available
    Description: Available space for Snapshot copies in bytes
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-snapshot-attributes.size-available
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.snapshot.available
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_snapshot_size_total
    Description: Total space for Snapshot copies in bytes
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-snapshot-attributes.size-total
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.snapshot.total
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_snapshot_size_used
    Description: Space used by Snapshot copies in bytes
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-snapshot-attributes.size-used
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.snapshot.used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_snapshot_used_percent
    Description: Percentage of disk space used by Snapshot copies
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-snapshot-attributes.percent-used-capacity
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.snapshot.used_percent
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_available
    Description: Space available in bytes.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.size-available
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.available
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_capacity_tier_used
    Description: Used space in bytes in the cloud store. Only applicable for aggregates with a cloud store tier.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.capacity-tier-used
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.cloud_storage.used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_data_compacted_count
    Description: Amount of compacted data in bytes.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.data-compacted-count
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.data_compacted_count
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_data_compaction_saved
    Description: Space saved in bytes by compacting the data.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.data-compaction-space-saved
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.data_compaction_space_saved
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_data_compaction_saved_percent
    Description: Percentage saved by compacting the data.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.data-compaction-space-saved-percent
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.data_compaction_space_saved_percent
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_performance_tier_inactive_user_data
    Description: The size that is physically used in the block storage and has a cold temperature, in bytes. This property is only supported if the aggregate is either attached to a cloud store or can be attached to a cloud store.This is an advanced property; there is an added computational cost to retrieving its value. The field is not populated for either a collection GET or an instance GET unless it is explicitly requested using the <i>fields</i> query parameter containing either block_storage.inactive_user_data or **.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.performance-tier-inactive-user-data
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.inactive_user_data
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_performance_tier_inactive_user_data_percent
    Description: The percentage of inactive user data in the block storage. This property is only supported if the aggregate is either attached to a cloud store or can be attached to a cloud store.This is an advanced property; there is an added computational cost to retrieving its value. The field is not populated for either a collection GET or an instance GET unless it is explicitly requested using the <i>fields</i> query parameter containing either block_storage.inactive_user_data_percent or **.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.performance-tier-inactive-user-data-percent
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.inactive_user_data_percent
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_physical_used
    Description: Total physical used size of an aggregate in bytes.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.physical-used
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.physical_used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_physical_used_percent
    Description: Physical used percentage.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.physical-used-percent
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.physical_used_percent
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_reserved
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.total-reserved-space
      template: conf/zapi/cdot/9.8.0/aggr.yaml

  - Harvest Metric: aggr_space_sis_saved
    Description: Amount of space saved in bytes by storage efficiency.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.sis-space-saved
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.volume_deduplication_space_saved
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_sis_saved_percent
    Description: Percentage of space saved by storage efficiency.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.sis-space-saved-percent
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.volume_deduplication_space_saved_percent
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_sis_shared_count
    Description: Amount of shared bytes counted by storage efficiency.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.sis-shared-count
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.volume_deduplication_shared_count
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_total
    Description: Total usable space in bytes, not including WAFL reserve and aggregate Snapshot copy reserve.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.size-total
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.size
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_used
    Description: Space used or reserved in bytes. Includes volume guarantees and aggregate metadata.
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.size-used
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.block_storage.used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_space_used_percent
    Description: The percentage of disk space currently in use on the referenced file system
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-space-attributes.percent-used-capacity
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: NA
      metric: Harvest Plugin Generated
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_total_logical_used
    Description: Logical used
    ZAPI:
      endpoint: aggr-efficiency-get-iter
      metric: aggr-efficiency-info.aggr-efficiency-cumulative-info.total-logical-used
      template: conf/zapi/cdot/9.9.0/aggr_efficiency.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: space.efficiency.logical_used
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_total_physical_used
    Description: Total Physical Used
    ZAPI:
      endpoint: aggr-efficiency-get-iter
      metric: aggr-efficiency-info.aggr-efficiency-cumulative-info.total-physical-used
      template: conf/zapi/cdot/9.9.0/aggr_efficiency.yaml
    REST:
      endpoint: NA
      metric: Harvest Plugin Generated
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: aggr_volume_count_flexvol
    ZAPI:
      endpoint: aggr-get-iter
      metric: aggr-attributes.aggr-volume-count-attributes.flexvol-count
      template: conf/zapi/cdot/9.8.0/aggr.yaml
    REST:
      endpoint: api/storage/aggregates
      metric: volume_count
      template: conf/rest/9.12.0/aggr.yaml

  - Harvest Metric: cluster_subsystem_outstanding_alerts
    Description: Number of outstanding alerts
    ZAPI:
      endpoint: diagnosis-subsystem-config-get-iter
      metric: diagnosis-subsystem-config-info.outstanding-alert-count
      template: conf/zapi/cdot/9.8.0/subsystem.yaml
    REST:
      endpoint: api/private/cli/system/health/subsystem
      metric: outstanding_alert_count
      template: conf/rest/9.12.0/subsystem.yaml

  - Harvest Metric: cluster_subsystem_suppressed_alerts
    Description: Number of suppressed alerts
    ZAPI:
      endpoint: diagnosis-subsystem-config-get-iter
      metric: diagnosis-subsystem-config-info.suppressed-alert-count
      template: conf/zapi/cdot/9.8.0/subsystem.yaml
    REST:
      endpoint: api/private/cli/system/health/subsystem
      metric: suppressed_alert_count
      template: conf/rest/9.12.0/subsystem.yaml

  - Harvest Metric: copy_manager_bce_copy_count_curr
    Description: Current number of copy requests being processed by the Block Copy Engine.
    ZAPI:
      endpoint: perf-object-get-instances copy_manager
      metric: bce_copy_count_curr
      template: conf/zapiperf/cdot/9.8.0/copy_manager.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/copy_manager
      metric: block_copy_engine_current_copy_count
      template: conf/restperf/9.12.0/copy_manager.yaml
      Unit: none
      Type: delta

  - Harvest Metric: copy_manager_kb_copied
    Description: Sum of kilo-bytes copied.
    ZAPI:
      endpoint: perf-object-get-instances copy_manager
      metric: KB_copied
      template: conf/zapiperf/cdot/9.8.0/copy_manager.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/copy_manager
      metric: KB_copied
      template: conf/restperf/9.12.0/copy_manager.yaml
      Unit: none
      Type: delta

  - Harvest Metric: copy_manager_ocs_copy_count_curr
    Description: Current number of copy requests being processed by the ONTAP copy subsystem.
    ZAPI:
      endpoint: perf-object-get-instances copy_manager
      metric: ocs_copy_count_curr
      template: conf/zapiperf/cdot/9.8.0/copy_manager.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/copy_manager
      metric: ontap_copy_subsystem_current_copy_count
      template: conf/restperf/9.12.0/copy_manager.yaml
      Unit: none
      Type: delta

  - Harvest Metric: copy_manager_sce_copy_count_curr
    Description: Current number of copy requests being processed by the System Continuous Engineering.
    ZAPI:
      endpoint: perf-object-get-instances copy_manager
      metric: sce_copy_count_curr
      template: conf/zapiperf/cdot/9.8.0/copy_manager.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/copy_manager
      metric: system_continuous_engineering_current_copy_count
      template: conf/restperf/9.12.0/copy_manager.yaml
      Unit: none
      Type: delta

  - Harvest Metric: copy_manager_spince_copy_count_curr
    Description: Current number of copy requests being processed by the SpinCE.
    ZAPI:
      endpoint: perf-object-get-instances copy_manager
      metric: spince_copy_count_curr
      template: conf/zapiperf/cdot/9.8.0/copy_manager.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/copy_manager
      metric: spince_current_copy_count
      template: conf/restperf/9.12.0/copy_manager.yaml
      Unit: none
      Type: delta

  - Harvest Metric: disk_busy
    Description: The utilization percent of the disk
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: disk_busy
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: disk_busy_percent
      template: conf/restperf/9.12.0/disk.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: disk_bytes_per_sector
    Description: Bytes per sector.
    ZAPI:
      endpoint: storage-disk-get-iter
      metric: storage-disk-info.disk-inventory-info.bytes-per-sector
      template: conf/zapi/cdot/9.8.0/disk.yaml
    REST:
      endpoint: api/storage/disks
      metric: bytes_per_sector
      template: conf/rest/9.12.0/disk.yaml

  - Harvest Metric: disk_capacity
    Description: Disk capacity in MB
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: disk_capacity
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: mb
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: capacity
      template: conf/restperf/9.12.0/disk.yaml
      Unit: mb
      Type: raw

  - Harvest Metric: disk_cp_read_chain
    Description: Average number of blocks transferred in each consistency point read operation during a CP
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: cp_read_chain
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: none
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: cp_read_chain
      template: conf/restperf/9.12.0/disk.yaml
      Unit: none
      Type: average

  - Harvest Metric: disk_cp_read_latency
    Description: Average latency per block in microseconds for consistency point read operations
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: cp_read_latency
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: cp_read_latency
      template: conf/restperf/9.12.0/disk.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: disk_cp_reads
    Description: Number of disk read operations initiated each second for consistency point processing
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: cp_reads
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: cp_read_count
      template: conf/restperf/9.12.0/disk.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: disk_io_pending
    Description: Average number of I/Os issued to the disk for which we have not yet received the response
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: io_pending
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: none
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: io_pending
      template: conf/restperf/9.12.0/disk.yaml
      Unit: none
      Type: average

  - Harvest Metric: disk_io_queued
    Description: Number of I/Os queued to the disk but not yet issued
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: io_queued
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: none
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: io_queued
      template: conf/restperf/9.12.0/disk.yaml
      Unit: none
      Type: average

  - Harvest Metric: disk_power_on_hours
    Description: Hours powered on.
    REST:
      endpoint: api/storage/disks
      metric: stats.power_on_hours
      template: conf/rest/9.12.0/disk.yaml

  - Harvest Metric: disk_sectors
    Description: Number of sectors on the disk.
    ZAPI:
      endpoint: storage-disk-get-iter
      metric: storage-disk-info.disk-inventory-info.capacity-sectors
      template: conf/zapi/cdot/9.8.0/disk.yaml
    REST:
      endpoint: api/storage/disks
      metric: sector_count
      template: conf/rest/9.12.0/disk.yaml

  - Harvest Metric: disk_stats_average_latency
    Description: Average I/O latency across all active paths, in milliseconds.
    ZAPI:
      endpoint: storage-disk-get-iter
      metric: storage-disk-info.disk-stats-info.average-latency
      template: conf/zapi/cdot/9.8.0/disk.yaml
    REST:
      endpoint: api/storage/disks
      metric: stats.average_latency
      template: conf/rest/9.12.0/disk.yaml

  - Harvest Metric: disk_stats_io_kbps
    Description: Total Disk Throughput in KBPS Across All Active Paths
    ZAPI:
      endpoint: storage-disk-get-iter
      metric: storage-disk-info.disk-stats-info.disk-io-kbps
      template: conf/zapi/cdot/9.8.0/disk.yaml
    REST:
      endpoint: api/private/cli/disk
      metric: disk_io_kbps_total
      template: conf/rest/9.12.0/disk.yaml

  - Harvest Metric: disk_stats_sectors_read
    Description: Number of Sectors Read
    ZAPI:
      endpoint: storage-disk-get-iter
      metric: storage-disk-info.disk-stats-info.sectors-read
      template: conf/zapi/cdot/9.8.0/disk.yaml
    REST:
      endpoint: api/private/cli/disk
      metric: sectors_read
      template: conf/rest/9.12.0/disk.yaml

  - Harvest Metric: disk_stats_sectors_written
    Description: Number of Sectors Written
    ZAPI:
      endpoint: storage-disk-get-iter
      metric: storage-disk-info.disk-stats-info.sectors-written
      template: conf/zapi/cdot/9.8.0/disk.yaml
    REST:
      endpoint: api/private/cli/disk
      metric: sectors_written
      template: conf/rest/9.12.0/disk.yaml

  - Harvest Metric: disk_total_transfers
    Description: Total number of disk operations involving data transfer initiated per second
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: total_transfers
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: total_transfer_count
      template: conf/restperf/9.12.0/disk.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: disk_uptime
    Description: Number of seconds the drive has been powered on
    ZAPI:
      endpoint: storage-disk-get-iter
      metric: storage-disk-info.disk-stats-info.power-on-time-interval
      template: conf/zapi/cdot/9.8.0/disk.yaml
    REST:
      endpoint: NA
      metric: Harvest Plugin Generated
      template: conf/rest/9.12.0/disk.yaml

  - Harvest Metric: disk_usable_size
    REST:
      endpoint: api/storage/disks
      metric: usable_size
      template: conf/rest/9.12.0/disk.yaml

  - Harvest Metric: disk_user_read_blocks
    Description: Number of blocks transferred for user read operations per second
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: user_read_blocks
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: user_read_block_count
      template: conf/restperf/9.12.0/disk.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: disk_user_read_chain
    Description: Average number of blocks transferred in each user read operation
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: user_read_chain
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: none
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: user_read_chain
      template: conf/restperf/9.12.0/disk.yaml
      Unit: none
      Type: average

  - Harvest Metric: disk_user_read_latency
    Description: Average latency per block in microseconds for user read operations
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: user_read_latency
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: user_read_latency
      template: conf/restperf/9.12.0/disk.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: disk_user_reads
    Description: Number of disk read operations initiated each second for retrieving data or metadata associated with user requests
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: user_reads
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: user_read_count
      template: conf/restperf/9.12.0/disk.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: disk_user_write_blocks
    Description: Number of blocks transferred for user write operations per second
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: user_write_blocks
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: user_write_block_count
      template: conf/restperf/9.12.0/disk.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: disk_user_write_chain
    Description: Average number of blocks transferred in each user write operation
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: user_write_chain
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: none
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: user_write_chain
      template: conf/restperf/9.12.0/disk.yaml
      Unit: none
      Type: average

  - Harvest Metric: disk_user_write_latency
    Description: Average latency per block in microseconds for user write operations
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: user_write_latency
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: user_write_latency
      template: conf/restperf/9.12.0/disk.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: disk_user_writes
    Description: Number of disk write operations initiated each second for storing data or metadata associated with user requests
    ZAPI:
      endpoint: perf-object-get-instances disk:constituent
      metric: user_writes
      template: conf/zapiperf/cdot/9.8.0/disk.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/disk:constituent
      metric: user_write_count
      template: conf/restperf/9.12.0/disk.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: environment_sensor_threshold_value
    Description: Provides the sensor reading.
    ZAPI:
      endpoint: environment-sensors-get-iter
      metric: environment-sensors-info.threshold-sensor-value
      template: conf/zapi/cdot/9.8.0/sensor.yaml
    REST:
      endpoint: api/cluster/sensors
      metric: value
      template: conf/rest/9.12.0/sensor.yaml

  - Harvest Metric: fabricpool_average_latency
    Description: Note This counter is deprecated and will be removed in a future release.  Average latencies executed during various phases of command execution. The execution-start latency represents the average time taken to start executing a operation. The request-prepare latency represent the average time taken to prepare the commplete request that needs to be sent to the server. The send latency represents the average time taken to send requests to the server. The execution-start-to-send-complete represents the average time taken to send a operation out since its execution started. The execution-start-to-first-byte-received represent the average time taken to to receive the first byte of a response since the command&apos;s request execution started. These counters can be used to identify performance bottlenecks within the object store client module.
    ZAPI:
      endpoint: perf-object-get-instances object_store_client_op
      metric: average_latency
      template: conf/zapiperf/cdot/9.8.0/object_store_client_op.yaml
      Unit: microsec
      Type: average,no-zero-values

  - Harvest Metric: fabricpool_cloud_bin_op_latency_average
    Description: Cloud bin operation latency average in milliseconds.
    ZAPI:
      endpoint: perf-object-get-instances wafl_comp_aggr_vol_bin
      metric: cloud_bin_op_latency_average
      template: conf/zapiperf/cdot/9.8.0/wafl_comp_aggr_vol_bin.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/wafl_comp_aggr_vol_bin
      metric: cloud_bin_op_latency_average
      template: conf/restperf/9.12.0/wafl_comp_aggr_vol_bin.yaml
      Unit: none
      Type: raw

  - Harvest Metric: fabricpool_cloud_bin_operation
    Description: Cloud bin operation counters.
    ZAPI:
      endpoint: perf-object-get-instances wafl_comp_aggr_vol_bin
      metric: cloud_bin_operation
      template: conf/zapiperf/cdot/9.8.0/wafl_comp_aggr_vol_bin.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/wafl_comp_aggr_vol_bin
      metric: cloud_bin_op
      template: conf/restperf/9.12.0/wafl_comp_aggr_vol_bin.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fabricpool_get_throughput_bytes
    Description: Note This counter is deprecated and will be removed in a future release.  Counter that indicates the throughput for GET command in bytes per second.
    ZAPI:
      endpoint: perf-object-get-instances object_store_client_op
      metric: get_throughput_bytes
      template: conf/zapiperf/cdot/9.8.0/object_store_client_op.yaml
      Unit: b_per_sec
      Type: rate,no-zero-values

  - Harvest Metric: fabricpool_put_throughput_bytes
    Description: Note This counter is deprecated and will be removed in a future release.  Counter that indicates the throughput for PUT command in bytes per second.
    ZAPI:
      endpoint: perf-object-get-instances object_store_client_op
      metric: put_throughput_bytes
      template: conf/zapiperf/cdot/9.8.0/object_store_client_op.yaml
      Unit: b_per_sec
      Type: rate,no-zero-values

  - Harvest Metric: fabricpool_stats
    Description: Note This counter is deprecated and will be removed in a future release.  Counter that indicates the number of object store operations sent, and their success and failure counts. The objstore_client_op_name array indicate the operation name such as PUT, GET, etc. The objstore_client_op_stats_name array contain the total number of operations, their success and failure counter for each operation.
    ZAPI:
      endpoint: perf-object-get-instances object_store_client_op
      metric: stats
      template: conf/zapiperf/cdot/9.8.0/object_store_client_op.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: fabricpool_throughput_ops
    Description: Counter that indicates the throughput for commands in ops per second.
    ZAPI:
      endpoint: perf-object-get-instances object_store_client_op
      metric: throughput_ops
      template: conf/zapiperf/cdot/9.8.0/object_store_client_op.yaml
      Unit: per_sec
      Type: rate,no-zero-values

  - Harvest Metric: fcp_avg_other_latency
    Description: Average latency for operations other than read and write
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: avg_other_latency
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: average_other_latency
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_avg_read_latency
    Description: Average latency for read operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: avg_read_latency
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: average_read_latency
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_avg_write_latency
    Description: Average latency for write operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: avg_write_latency
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: average_write_latency
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_discarded_frames_count
    Description: Number of discarded frames.
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: discarded_frames_count
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: discarded_frames_count
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_int_count
    Description: Number of interrupts
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: int_count
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: interrupt_count
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_invalid_crc
    Description: Number of invalid cyclic redundancy checks (CRC count)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: invalid_crc
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: invalid.crc
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_invalid_transmission_word
    Description: Number of invalid transmission words
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: invalid_transmission_word
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: invalid.transmission_word
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_isr_count
    Description: Number of interrupt responses
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: isr_count
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: isr.count
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_lif_avg_latency
    Description: Average latency for FCP operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_lif
      metric: avg_latency
      template: conf/zapiperf/cdot/9.8.0/fcp_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp_lif
      metric: average_latency
      template: conf/restperf/9.12.0/fcp_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_lif_avg_other_latency
    Description: Average latency for operations other than read and write
    ZAPI:
      endpoint: perf-object-get-instances fcp_lif
      metric: avg_other_latency
      template: conf/zapiperf/cdot/9.8.0/fcp_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp_lif
      metric: average_other_latency
      template: conf/restperf/9.12.0/fcp_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_lif_avg_read_latency
    Description: Average latency for read operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_lif
      metric: avg_read_latency
      template: conf/zapiperf/cdot/9.8.0/fcp_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp_lif
      metric: average_read_latency
      template: conf/restperf/9.12.0/fcp_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_lif_avg_write_latency
    Description: Average latency for write operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_lif
      metric: avg_write_latency
      template: conf/zapiperf/cdot/9.8.0/fcp_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp_lif
      metric: average_write_latency
      template: conf/restperf/9.12.0/fcp_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_lif_other_ops
    Description: Number of operations that are not read or write.
    ZAPI:
      endpoint: perf-object-get-instances fcp_lif
      metric: other_ops
      template: conf/zapiperf/cdot/9.8.0/fcp_lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp_lif
      metric: other_ops
      template: conf/restperf/9.12.0/fcp_lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_lif_read_data
    Description: Amount of data read from the storage system
    ZAPI:
      endpoint: perf-object-get-instances fcp_lif
      metric: read_data
      template: conf/zapiperf/cdot/9.8.0/fcp_lif.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp_lif
      metric: read_data
      template: conf/restperf/9.12.0/fcp_lif.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_lif_read_ops
    Description: Number of read operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_lif
      metric: read_ops
      template: conf/zapiperf/cdot/9.8.0/fcp_lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp_lif
      metric: read_ops
      template: conf/restperf/9.12.0/fcp_lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_lif_total_ops
    Description: Total number of operations.
    ZAPI:
      endpoint: perf-object-get-instances fcp_lif
      metric: total_ops
      template: conf/zapiperf/cdot/9.8.0/fcp_lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp_lif
      metric: total_ops
      template: conf/restperf/9.12.0/fcp_lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_lif_write_data
    Description: Amount of data written to the storage system
    ZAPI:
      endpoint: perf-object-get-instances fcp_lif
      metric: write_data
      template: conf/zapiperf/cdot/9.8.0/fcp_lif.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp_lif
      metric: write_data
      template: conf/restperf/9.12.0/fcp_lif.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_lif_write_ops
    Description: Number of write operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_lif
      metric: write_ops
      template: conf/zapiperf/cdot/9.8.0/fcp_lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp_lif
      metric: write_ops
      template: conf/restperf/9.12.0/fcp_lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_link_down
    Description: Number of times the Fibre Channel link was lost
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: link_down
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: link.down
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_link_failure
    Description: Number of link failures
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: link_failure
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: link_failure
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_loss_of_signal
    Description: Number of times this port lost signal
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: loss_of_signal
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: loss_of_signal
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_loss_of_sync
    Description: Number of times this port lost sync
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: loss_of_sync
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: loss_of_sync
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_nvmf_avg_other_latency
    Description: Average latency for operations other than read and write (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_avg_other_latency
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.average_other_latency
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_nvmf_avg_read_latency
    Description: Average latency for read operations (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_avg_read_latency
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.average_read_latency
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_nvmf_avg_remote_other_latency
    Description: Average latency for remote operations other than read and write (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_avg_remote_other_latency
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.average_remote_other_latency
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_nvmf_avg_remote_read_latency
    Description: Average latency for remote read operations (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_avg_remote_read_latency
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.average_remote_read_latency
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_nvmf_avg_remote_write_latency
    Description: Average latency for remote write operations (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_avg_remote_write_latency
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.average_remote_write_latency
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_nvmf_avg_write_latency
    Description: Average latency for write operations (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_avg_write_latency
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.average_write_latency
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcp_nvmf_caw_data
    Description: Amount of CAW data sent to the storage system (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_caw_data
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.caw_data
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_caw_ops
    Description: Number of FC-NVMe CAW operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_caw_ops
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.caw_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_command_slots
    Description: Number of command slots that have been used by initiators logging into this port. This shows the command fan-in on the port.
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_command_slots
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.command_slots
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_other_ops
    Description: Number of NVMF operations that are not read or write.
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_other_ops
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.other_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_read_data
    Description: Amount of data read from the storage system (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_read_data
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.read_data
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_read_ops
    Description: Number of FC-NVMe read operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_read_ops
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.read_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_remote_caw_data
    Description: Amount of remote CAW data sent to the storage system (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_remote_caw_data
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf_remote.caw_data
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_remote_caw_ops
    Description: Number of FC-NVMe remote CAW operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_remote_caw_ops
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf_remote.caw_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_remote_other_ops
    Description: Number of NVMF remote operations that are not read or write.
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_remote_other_ops
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf_remote.other_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_remote_read_data
    Description: Amount of remote data read from the storage system (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_remote_read_data
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf_remote.read_data
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_remote_read_ops
    Description: Number of FC-NVMe remote read operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_remote_read_ops
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf_remote.read_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_remote_total_data
    Description: Amount of remote FC-NVMe traffic to and from the storage system
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_remote_total_data
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf_remote.total_data
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_remote_total_ops
    Description: Total number of remote FC-NVMe operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_remote_total_ops
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf_remote.total_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_remote_write_data
    Description: Amount of remote data written to the storage system (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_remote_write_data
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf_remote.write_data
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_remote_write_ops
    Description: Number of FC-NVMe remote write operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_remote_write_ops
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf_remote.write_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_total_data
    Description: Amount of FC-NVMe traffic to and from the storage system
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_total_data
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.total_data
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_total_ops
    Description: Total number of FC-NVMe operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_total_ops
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.total_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_write_data
    Description: Amount of data written to the storage system (FC-NVMe)
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_write_data
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.write_data
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_nvmf_write_ops
    Description: Number of FC-NVMe write operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: nvmf_write_ops
      template: conf/zapiperf/cdot/9.10.1/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: nvmf.write_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_other_ops
    Description: Number of operations that are not read or write.
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: other_ops
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: other_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_prim_seq_err
    Description: Number of primitive sequence errors
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: prim_seq_err
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: primitive_seq_err
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_queue_full
    Description: Number of times a queue full condition occurred.
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: queue_full
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: queue_full
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_read_data
    Description: Amount of data read from the storage system
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: read_data
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: read_data
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_read_ops
    Description: Number of read operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: read_ops
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: read_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_reset_count
    Description: Number of physical port resets
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: reset_count
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: reset_count
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_shared_int_count
    Description: Number of shared interrupts
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: shared_int_count
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: shared_interrupt_count
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_spurious_int_count
    Description: Number of spurious interrupts
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: spurious_int_count
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: spurious_interrupt_count
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_threshold_full
    Description: Number of times the total number of outstanding commands on the port exceeds the threshold supported by this port.
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: threshold_full
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: none
      Type: delta,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: threshold_full
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: none
      Type: delta

  - Harvest Metric: fcp_total_data
    Description: Amount of FCP traffic to and from the storage system
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: total_data
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: total_data
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_total_ops
    Description: Total number of FCP operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: total_ops
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: total_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcp_write_data
    Description: Amount of data written to the storage system
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: write_data
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: write_data
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: fcp_write_ops
    Description: Number of write operations
    ZAPI:
      endpoint: perf-object-get-instances fcp_port
      metric: write_ops
      template: conf/zapiperf/cdot/9.8.0/fcp.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcp
      metric: write_ops
      template: conf/restperf/9.12.0/fcp.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: fcvi_rdma_write_avg_latency
    Description: Average RDMA write I/O latency.
    ZAPI:
      endpoint: perf-object-get-instances fcvi
      metric: rdma_write_avg_latency
      template: conf/zapiperf/cdot/9.8.0/fcvi.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/fcvi
      metric: rdma.write_average_latency
      template: conf/restperf/9.12.0/fcvi.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: fcvi_rdma_write_ops
    Description: Number of RDMA write I/Os issued per second.
    ZAPI:
      endpoint: perf-object-get-instances fcvi
      metric: rdma_write_ops
      template: conf/zapiperf/cdot/9.8.0/fcvi.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcvi
      metric: rdma.write_ops
      template: conf/restperf/9.12.0/fcvi.yaml
      Unit: none
      Type: rate

  - Harvest Metric: fcvi_rdma_write_throughput
    Description: RDMA write throughput in bytes per second.
    ZAPI:
      endpoint: perf-object-get-instances fcvi
      metric: rdma_write_throughput
      template: conf/zapiperf/cdot/9.8.0/fcvi.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/fcvi
      metric: rdma.write_throughput
      template: conf/restperf/9.12.0/fcvi.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: flashcache_accesses
    Description: External cache accesses per second
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: accesses
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: accesses
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_disk_reads_replaced
    Description: Estimated number of disk reads per second replaced by cache
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: disk_reads_replaced
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: disk_reads_replaced
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_evicts
    Description: Number of blocks evicted from the external cache to make room for new blocks
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: evicts
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: evicts
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_hit
    Description: Number of WAFL buffers served off the external cache
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: hit
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: hit.total
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_hit_directory
    Description: Number of directory buffers served off the external cache
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: hit_directory
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: hit.directory
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_hit_indirect
    Description: Number of indirect file buffers served off the external cache
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: hit_indirect
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: hit.indirect
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_hit_metadata_file
    Description: Number of metadata file buffers served off the external cache
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: hit_metadata_file
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: hit.metadata_file
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_hit_normal_lev0
    Description: Number of normal level 0 WAFL buffers served off the external cache
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: hit_normal_lev0
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: hit.normal_level_zero
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_hit_percent
    Description: External cache hit rate
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: hit_percent
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: hit.percent
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: percent
      Type: average

  - Harvest Metric: flashcache_inserts
    Description: Number of WAFL buffers inserted into the external cache
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: inserts
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: inserts
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_invalidates
    Description: Number of blocks invalidated in the external cache
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: invalidates
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: invalidates
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_miss
    Description: External cache misses
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: miss
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: miss.total
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_miss_directory
    Description: External cache misses accessing directory buffers
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: miss_directory
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: miss.directory
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_miss_indirect
    Description: External cache misses accessing indirect file buffers
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: miss_indirect
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: miss.indirect
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_miss_metadata_file
    Description: External cache misses accessing metadata file buffers
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: miss_metadata_file
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: miss.metadata_file
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_miss_normal_lev0
    Description: External cache misses accessing normal level 0 buffers
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: miss_normal_lev0
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: miss.normal_level_zero
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashcache_usage
    Description: Percentage of blocks in external cache currently containing valid data
    ZAPI:
      endpoint: perf-object-get-instances ext_cache_obj
      metric: usage
      template: conf/zapiperf/cdot/9.8.0/ext_cache_obj.yaml
      Unit: percent
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/external_cache
      metric: usage
      template: conf/restperf/9.12.0/ext_cache_obj.yaml
      Unit: percent
      Type: raw

  - Harvest Metric: flashpool_cache_stats
    Description: Automated Working-set Analyzer (AWA) per-interval pseudo cache statistics for the most recent intervals. The number of intervals defined as recent is CM_WAFL_HYAS_INT_DIS_CNT. This array is a table with fields corresponding to the enum type of hyas_cache_stat_type_t.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_sizer
      metric: cache_stats
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_sizer.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_sizer
      metric: cache_stats
      template: conf/restperf/9.12.0/wafl_hya_sizer.yaml
      Unit: none
      Type: raw

  - Harvest Metric: flashpool_evict_destage_rate
    Description: Number of block destage per second.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: evict_destage_rate
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: evict_destage_rate
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashpool_evict_remove_rate
    Description: Number of block free per second.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: evict_remove_rate
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: evict_remove_rate
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashpool_hya_read_hit_latency_average
    Description: Average of RAID I/O latency on read hit.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: hya_read_hit_latency_average
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: hya_read_hit_latency_average
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: average

  - Harvest Metric: flashpool_hya_read_miss_latency_average
    Description: Average read miss latency.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: hya_read_miss_latency_average
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: hya_read_miss_latency_average
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: average

  - Harvest Metric: flashpool_hya_write_hdd_latency_average
    Description: Average write latency to HDD.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: hya_write_hdd_latency_average
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: hya_write_hdd_latency_average
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: average

  - Harvest Metric: flashpool_hya_write_ssd_latency_average
    Description: Average of RAID I/O latency on write to SSD.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: hya_write_ssd_latency_average
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: hya_write_ssd_latency_average
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: average

  - Harvest Metric: flashpool_read_cache_ins_rate
    Description: Cache insert rate blocks/sec.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: read_cache_ins_rate
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: read_cache_insert_rate
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashpool_read_ops_replaced
    Description: Number of HDD read operations replaced by SSD reads per second.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: read_ops_replaced
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: read_ops_replaced
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashpool_read_ops_replaced_percent
    Description: Percentage of HDD read operations replace by SSD.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: read_ops_replaced_percent
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: percent
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: read_ops_replaced_percent
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: percent
      Type: average

  - Harvest Metric: flashpool_ssd_available
    Description: Total SSD blocks available.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: ssd_available
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: ssd_available
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: raw

  - Harvest Metric: flashpool_ssd_read_cached
    Description: Total read cached SSD blocks.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: ssd_read_cached
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: ssd_read_cached
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: raw

  - Harvest Metric: flashpool_ssd_total
    Description: Total SSD blocks.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: ssd_total
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: ssd_total
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: raw

  - Harvest Metric: flashpool_ssd_total_used
    Description: Total SSD blocks used.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: ssd_total_used
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: ssd_total_used
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: raw

  - Harvest Metric: flashpool_ssd_write_cached
    Description: Total write cached SSD blocks.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: ssd_write_cached
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: ssd_write_cached
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: none
      Type: raw

  - Harvest Metric: flashpool_wc_write_blks_total
    Description: Number of write-cache blocks written per second.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: wc_write_blks_total
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: wc_write_blocks_total
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashpool_write_blks_replaced
    Description: Number of HDD write blocks replaced by SSD writes per second.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: write_blks_replaced
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: write_blocks_replaced
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: flashpool_write_blks_replaced_percent
    Description: Percentage of blocks overwritten to write-cache among all disk writes.
    ZAPI:
      endpoint: perf-object-get-instances wafl_hya_per_aggr
      metric: write_blks_replaced_percent
      template: conf/zapiperf/cdot/9.8.0/wafl_hya_per_aggr.yaml
      Unit: percent
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/wafl_hya_per_aggregate
      metric: write_blocks_replaced_percent
      template: conf/restperf/9.12.0/wafl_hya_per_aggr.yaml
      Unit: percent
      Type: average

  - Harvest Metric: headroom_aggr_current_latency
    Description: This is the storage aggregate average latency per message at the disk level.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_aggr
      metric: current_latency
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: /api/cluster/counter/tables/headroom_aggregate
      metric: current_latency
      template: conf/restperf/9.12.0/resource_headroom_aggr.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: headroom_aggr_current_ops
    Description: Total number of I/Os processed by the aggregate per second.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_aggr
      metric: current_ops
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: /api/cluster/counter/tables/headroom_aggregate
      metric: current_ops
      template: conf/restperf/9.12.0/resource_headroom_aggr.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: headroom_aggr_current_utilization
    Description: This is the storage aggregate average utilization of all the data disks in the aggregate.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_aggr
      metric: current_utilization
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: /api/cluster/counter/tables/headroom_aggregate
      metric: current_utilization
      template: conf/restperf/9.12.0/resource_headroom_aggr.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: headroom_aggr_ewma_daily
    Description: Daily exponential weighted moving average.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_aggr
      metric: ewma_daily
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: /api/cluster/counter/tables/headroom_aggregate
      metric: ewma.daily
      template: conf/restperf/9.12.0/resource_headroom_aggr.yaml
      Unit: none
      Type: raw

  - Harvest Metric: headroom_aggr_ewma_hourly
    Description: Hourly exponential weighted moving average.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_aggr
      metric: ewma_hourly
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: /api/cluster/counter/tables/headroom_aggregate
      metric: ewma.hourly
      template: conf/restperf/9.12.0/resource_headroom_aggr.yaml
      Unit: none
      Type: raw

  - Harvest Metric: headroom_aggr_ewma_monthly
    Description: Monthly exponential weighted moving average.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_aggr
      metric: ewma_monthly
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: /api/cluster/counter/tables/headroom_aggregate
      metric: ewma.monthly
      template: conf/restperf/9.12.0/resource_headroom_aggr.yaml
      Unit: none
      Type: raw

  - Harvest Metric: headroom_aggr_ewma_weekly
    Description: Weekly exponential weighted moving average.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_aggr
      metric: ewma_weekly
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: /api/cluster/counter/tables/headroom_aggregate
      metric: ewma.weekly
      template: conf/restperf/9.12.0/resource_headroom_aggr.yaml
      Unit: none
      Type: raw

  - Harvest Metric: headroom_aggr_optimal_point_confidence_factor
    Description: The confidence factor for the optimal point value based on the observed resource latency and utilization.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_aggr
      metric: optimal_point_confidence_factor
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml
      Unit: none
      Type: average
    REST:
      endpoint: /api/cluster/counter/tables/headroom_aggregate
      metric: optimal_point.confidence_factor
      template: conf/restperf/9.12.0/resource_headroom_aggr.yaml
      Unit: none
      Type: average

  - Harvest Metric: headroom_aggr_optimal_point_latency
    Description: The latency component of the optimal point of the latency/utilization curve.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_aggr
      metric: optimal_point_latency
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: /api/cluster/counter/tables/headroom_aggregate
      metric: optimal_point.latency
      template: conf/restperf/9.12.0/resource_headroom_aggr.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: headroom_aggr_optimal_point_ops
    Description: The ops component of the optimal point derived from the latency/utilzation curve.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_aggr
      metric: optimal_point_ops
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml
      Unit: per_sec
      Type: average
    REST:
      endpoint: /api/cluster/counter/tables/headroom_aggregate
      metric: optimal_point.ops
      template: conf/restperf/9.12.0/resource_headroom_aggr.yaml
      Unit: per_sec
      Type: average

  - Harvest Metric: headroom_aggr_optimal_point_utilization
    Description: The utilization component of the optimal point of the latency/utilization curve.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_aggr
      metric: optimal_point_utilization
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_aggr.yaml
      Unit: none
      Type: average
    REST:
      endpoint: /api/cluster/counter/tables/headroom_aggregate
      metric: optimal_point.utilization
      template: conf/restperf/9.12.0/resource_headroom_aggr.yaml
      Unit: none
      Type: average

  - Harvest Metric: headroom_cpu_current_latency
    Description: Current operation latency of the resource.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_cpu
      metric: current_latency
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: /api/cluster/counter/tables/headroom_cpu
      metric: current_latency
      template: conf/restperf/9.12.0/resource_headroom_cpu.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: headroom_cpu_current_ops
    Description: Total number of operations per second (also referred to as dblade ops).
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_cpu
      metric: current_ops
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: /api/cluster/counter/tables/headroom_cpu
      metric: current_ops
      template: conf/restperf/9.12.0/resource_headroom_cpu.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: headroom_cpu_current_utilization
    Description: Average processor utilization across all processors in the system.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_cpu
      metric: current_utilization
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: /api/cluster/counter/tables/headroom_cpu
      metric: current_utilization
      template: conf/restperf/9.12.0/resource_headroom_cpu.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: headroom_cpu_ewma_daily
    Description: Daily exponential weighted moving average for current_ops, optimal_point_ops, current_latency, optimal_point_latency, current_utilization, optimal_point_utilization and optimal_point_confidence_factor.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_cpu
      metric: ewma_daily
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: /api/cluster/counter/tables/headroom_cpu
      metric: ewma.daily
      template: conf/restperf/9.12.0/resource_headroom_cpu.yaml
      Unit: none
      Type: raw

  - Harvest Metric: headroom_cpu_ewma_hourly
    Description: Hourly exponential weighted moving average for current_ops, optimal_point_ops, current_latency, optimal_point_latency, current_utilization, optimal_point_utilization and optimal_point_confidence_factor.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_cpu
      metric: ewma_hourly
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: /api/cluster/counter/tables/headroom_cpu
      metric: ewma.hourly
      template: conf/restperf/9.12.0/resource_headroom_cpu.yaml
      Unit: none
      Type: raw

  - Harvest Metric: headroom_cpu_ewma_monthly
    Description: Monthly exponential weighted moving average for current_ops, optimal_point_ops, current_latency, optimal_point_latency, current_utilization, optimal_point_utilization and optimal_point_confidence_factor.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_cpu
      metric: ewma_monthly
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: /api/cluster/counter/tables/headroom_cpu
      metric: ewma.monthly
      template: conf/restperf/9.12.0/resource_headroom_cpu.yaml
      Unit: none
      Type: raw

  - Harvest Metric: headroom_cpu_ewma_weekly
    Description: Weekly exponential weighted moving average for current_ops, optimal_point_ops, current_latency, optimal_point_latency, current_utilization, optimal_point_utilization and optimal_point_confidence_factor.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_cpu
      metric: ewma_weekly
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: /api/cluster/counter/tables/headroom_cpu
      metric: ewma.weekly
      template: conf/restperf/9.12.0/resource_headroom_cpu.yaml
      Unit: none
      Type: raw

  - Harvest Metric: headroom_cpu_optimal_point_confidence_factor
    Description: Confidence factor for the optimal point value based on the observed resource latency and utilization. The possible values are: 0 - unknown, 1 - low, 2 - medium, 3 - high. This counter can provide an average confidence factor over a range of time.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_cpu
      metric: optimal_point_confidence_factor
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml
      Unit: none
      Type: average
    REST:
      endpoint: /api/cluster/counter/tables/headroom_cpu
      metric: optimal_point.confidence_factor
      template: conf/restperf/9.12.0/resource_headroom_cpu.yaml
      Unit: none
      Type: average

  - Harvest Metric: headroom_cpu_optimal_point_latency
    Description: Latency component of the optimal point of the latency/utilization curve. This counter can provide an average latency over a range of time.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_cpu
      metric: optimal_point_latency
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: /api/cluster/counter/tables/headroom_cpu
      metric: optimal_point.latency
      template: conf/restperf/9.12.0/resource_headroom_cpu.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: headroom_cpu_optimal_point_ops
    Description: Ops component of the optimal point derived from the latency/utilization curve. This counter can provide an average ops over a range of time.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_cpu
      metric: optimal_point_ops
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml
      Unit: per_sec
      Type: average
    REST:
      endpoint: /api/cluster/counter/tables/headroom_cpu
      metric: optimal_point.ops
      template: conf/restperf/9.12.0/resource_headroom_cpu.yaml
      Unit: per_sec
      Type: average

  - Harvest Metric: headroom_cpu_optimal_point_utilization
    Description: Utilization component of the optimal point of the latency/utilization curve. This counter can provide an average utilization over a range of time.
    ZAPI:
      endpoint: perf-object-get-instances resource_headroom_cpu
      metric: optimal_point_utilization
      template: conf/zapiperf/cdot/9.8.0/resource_headroom_cpu.yaml
      Unit: none
      Type: average
    REST:
      endpoint: /api/cluster/counter/tables/headroom_cpu
      metric: optimal_point.utilization
      template: conf/restperf/9.12.0/resource_headroom_cpu.yaml
      Unit: none
      Type: average

  - Harvest Metric: hostadapter_bytes_read
    Description: Bytes read through a host adapter
    ZAPI:
      endpoint: perf-object-get-instances hostadapter
      metric: bytes_read
      template: conf/zapiperf/cdot/9.8.0/hostadapter.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/host_adapter
      metric: bytes_read
      template: conf/restperf/9.12.0/hostadapter.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: hostadapter_bytes_written
    Description: Bytes written through a host adapter
    ZAPI:
      endpoint: perf-object-get-instances hostadapter
      metric: bytes_written
      template: conf/zapiperf/cdot/9.8.0/hostadapter.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/host_adapter
      metric: bytes_written
      template: conf/restperf/9.12.0/hostadapter.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: iscsi_lif_avg_latency
    Description: Average latency for iSCSI operations
    ZAPI:
      endpoint: perf-object-get-instances iscsi_lif
      metric: avg_latency
      template: conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/iscsi_lif
      metric: average_latency
      template: conf/restperf/9.12.0/iscsi_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: iscsi_lif_avg_other_latency
    Description: Average latency for operations other than read and write (for example, Inquiry, Report LUNs, SCSI Task Management Functions)
    ZAPI:
      endpoint: perf-object-get-instances iscsi_lif
      metric: avg_other_latency
      template: conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/iscsi_lif
      metric: average_other_latency
      template: conf/restperf/9.12.0/iscsi_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: iscsi_lif_avg_read_latency
    Description: Average latency for read operations
    ZAPI:
      endpoint: perf-object-get-instances iscsi_lif
      metric: avg_read_latency
      template: conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/iscsi_lif
      metric: average_read_latency
      template: conf/restperf/9.12.0/iscsi_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: iscsi_lif_avg_write_latency
    Description: Average latency for write operations
    ZAPI:
      endpoint: perf-object-get-instances iscsi_lif
      metric: avg_write_latency
      template: conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/iscsi_lif
      metric: average_write_latency
      template: conf/restperf/9.12.0/iscsi_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: iscsi_lif_cmd_transfered
    Description: Command transfered by this iSCSI conn
    ZAPI:
      endpoint: perf-object-get-instances iscsi_lif
      metric: cmd_transfered
      template: conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml
      Unit: none
      Type: rate

  - Harvest Metric: iscsi_lif_cmd_transferred
    Description: Command transferred by this iSCSI connection
    REST:
      endpoint: api/cluster/counter/tables/iscsi_lif
      metric: cmd_transferred
      template: conf/restperf/9.12.0/iscsi_lif.yaml
      Unit: none
      Type: rate

  - Harvest Metric: iscsi_lif_iscsi_other_ops
    Description: iSCSI other operations per second on this logical interface (LIF)
    ZAPI:
      endpoint: perf-object-get-instances iscsi_lif
      metric: iscsi_other_ops
      template: conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/iscsi_lif
      metric: iscsi_other_ops
      template: conf/restperf/9.12.0/iscsi_lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: iscsi_lif_iscsi_read_ops
    Description: iSCSI read operations per second on this logical interface (LIF)
    ZAPI:
      endpoint: perf-object-get-instances iscsi_lif
      metric: iscsi_read_ops
      template: conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/iscsi_lif
      metric: iscsi_read_ops
      template: conf/restperf/9.12.0/iscsi_lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: iscsi_lif_iscsi_write_ops
    Description: iSCSI write operations per second on this logical interface (LIF)
    ZAPI:
      endpoint: perf-object-get-instances iscsi_lif
      metric: iscsi_write_ops
      template: conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/iscsi_lif
      metric: iscsi_write_ops
      template: conf/restperf/9.12.0/iscsi_lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: iscsi_lif_protocol_errors
    Description: Number of protocol errors from iSCSI sessions on this logical interface (LIF)
    ZAPI:
      endpoint: perf-object-get-instances iscsi_lif
      metric: protocol_errors
      template: conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/iscsi_lif
      metric: protocol_errors
      template: conf/restperf/9.12.0/iscsi_lif.yaml
      Unit: none
      Type: delta

  - Harvest Metric: iscsi_lif_read_data
    Description: Amount of data read from the storage system in bytes
    ZAPI:
      endpoint: perf-object-get-instances iscsi_lif
      metric: read_data
      template: conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/iscsi_lif
      metric: read_data
      template: conf/restperf/9.12.0/iscsi_lif.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: iscsi_lif_write_data
    Description: Amount of data written to the storage system in bytes
    ZAPI:
      endpoint: perf-object-get-instances iscsi_lif
      metric: write_data
      template: conf/zapiperf/cdot/9.8.0/iscsi_lif.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/iscsi_lif
      metric: write_data
      template: conf/restperf/9.12.0/iscsi_lif.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: lif_recv_data
    Description: Number of bytes received per second
    ZAPI:
      endpoint: perf-object-get-instances lif
      metric: recv_data
      template: conf/zapiperf/cdot/9.8.0/lif.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lif
      metric: received_data
      template: conf/restperf/9.12.0/lif.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: lif_recv_errors
    Description: Number of received Errors per second
    ZAPI:
      endpoint: perf-object-get-instances lif
      metric: recv_errors
      template: conf/zapiperf/cdot/9.8.0/lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lif
      metric: received_errors
      template: conf/restperf/9.12.0/lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: lif_recv_packet
    Description: Number of packets received per second
    ZAPI:
      endpoint: perf-object-get-instances lif
      metric: recv_packet
      template: conf/zapiperf/cdot/9.8.0/lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lif
      metric: received_packets
      template: conf/restperf/9.12.0/lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: lif_sent_data
    Description: Number of bytes sent per second
    ZAPI:
      endpoint: perf-object-get-instances lif
      metric: sent_data
      template: conf/zapiperf/cdot/9.8.0/lif.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lif
      metric: sent_data
      template: conf/restperf/9.12.0/lif.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: lif_sent_errors
    Description: Number of sent errors per second
    ZAPI:
      endpoint: perf-object-get-instances lif
      metric: sent_errors
      template: conf/zapiperf/cdot/9.8.0/lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lif
      metric: sent_errors
      template: conf/restperf/9.12.0/lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: lif_sent_packet
    Description: Number of packets sent per second
    ZAPI:
      endpoint: perf-object-get-instances lif
      metric: sent_packet
      template: conf/zapiperf/cdot/9.8.0/lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lif
      metric: sent_packets
      template: conf/restperf/9.12.0/lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: lun_avg_read_latency
    Description: Average read latency in microseconds for all operations on the LUN
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: avg_read_latency
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: average_read_latency
      template: conf/restperf/9.12.0/lun.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: lun_avg_write_latency
    Description: Average write latency in microseconds for all operations on the LUN
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: avg_write_latency
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: average_write_latency
      template: conf/restperf/9.12.0/lun.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: lun_avg_xcopy_latency
    Description: Average latency in microseconds for xcopy requests
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: avg_xcopy_latency
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: average_xcopy_latency
      template: conf/restperf/9.12.0/lun.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: lun_caw_reqs
    Description: Number of compare and write requests
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: caw_reqs
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: caw_requests
      template: conf/restperf/9.12.0/lun.yaml
      Unit: none
      Type: rate

  - Harvest Metric: lun_enospc
    Description: Number of operations receiving ENOSPC errors
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: enospc
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: enospc
      template: conf/restperf/9.12.0/lun.yaml
      Unit: none
      Type: delta

  - Harvest Metric: lun_queue_full
    Description: Queue full responses
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: queue_full
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: queue_full
      template: conf/restperf/9.12.0/lun.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: lun_read_align_histo
    Description: Histogram of WAFL read alignment (number sectors off WAFL block start)
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: read_align_histo
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: read_align_histogram
      template: conf/restperf/9.12.0/lun.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: lun_read_data
    Description: Read bytes
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: read_data
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: read_data
      template: conf/restperf/9.12.0/lun.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: lun_read_ops
    Description: Number of read operations
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: read_ops
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: read_ops
      template: conf/restperf/9.12.0/lun.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: lun_read_partial_blocks
    Description: Percentage of reads whose size is not a multiple of WAFL block size
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: read_partial_blocks
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: read_partial_blocks
      template: conf/restperf/9.12.0/lun.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: lun_remote_bytes
    Description: I/O to or from a LUN which is not owned by the storage system handling the I/O.
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: remote_bytes
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: remote_bytes
      template: conf/restperf/9.12.0/lun.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: lun_remote_ops
    Description: Number of operations received by a storage system that does not own the LUN targeted by the operations.
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: remote_ops
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: remote_ops
      template: conf/restperf/9.12.0/lun.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: lun_size
    Description: The total provisioned size of the LUN. The LUN size can be increased but not be made smaller using the REST interface.<br/>The maximum and minimum sizes listed here are the absolute maximum and absolute minimum sizes in bytes. The actual minimum and maxiumum sizes vary depending on the ONTAP version, ONTAP platform and the available space in the containing volume and aggregate.<br/>For more information, see _Size properties_ in the _docs_ section of the ONTAP REST API documentation.
    ZAPI:
      endpoint: lun-get-iter
      metric: lun-info.size
      template: conf/zapi/cdot/9.8.0/lun.yaml
    REST:
      endpoint: api/storage/luns
      metric: space.size
      template: conf/rest/9.12.0/lun.yaml

  - Harvest Metric: lun_size_used
    Description: The amount of space consumed by the main data stream of the LUN.<br/>This value is the total space consumed in the volume by the LUN, including filesystem overhead, but excluding prefix and suffix streams. Due to internal filesystem overhead and the many ways SAN filesystems and applications utilize blocks within a LUN, this value does not necessarily reflect actual consumption/availability from the perspective of the filesystem or application. Without specific knowledge of how the LUN blocks are utilized outside of ONTAP, this property should not be used as an indicator for an out-of-space condition.<br/>For more information, see _Size properties_ in the _docs_ section of the ONTAP REST API documentation.
    ZAPI:
      endpoint: lun-get-iter
      metric: lun-info.size-used
      template: conf/zapi/cdot/9.8.0/lun.yaml
    REST:
      endpoint: api/storage/luns
      metric: space.used
      template: conf/rest/9.12.0/lun.yaml

  - Harvest Metric: lun_unmap_reqs
    Description: Number of unmap command requests
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: unmap_reqs
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: unmap_requests
      template: conf/restperf/9.12.0/lun.yaml
      Unit: none
      Type: rate

  - Harvest Metric: lun_write_align_histo
    Description: Histogram of WAFL write alignment (number of sectors off WAFL block start)
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: write_align_histo
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: write_align_histogram
      template: conf/restperf/9.12.0/lun.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: lun_write_data
    Description: Write bytes
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: write_data
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: write_data
      template: conf/restperf/9.12.0/lun.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: lun_write_ops
    Description: Number of write operations
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: write_ops
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: write_ops
      template: conf/restperf/9.12.0/lun.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: lun_write_partial_blocks
    Description: Percentage of writes whose size is not a multiple of WAFL block size
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: write_partial_blocks
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: write_partial_blocks
      template: conf/restperf/9.12.0/lun.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: lun_writesame_reqs
    Description: Number of write same command requests
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: writesame_reqs
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: writesame_requests
      template: conf/restperf/9.12.0/lun.yaml
      Unit: none
      Type: rate

  - Harvest Metric: lun_writesame_unmap_reqs
    Description: Number of write same commands requests with unmap bit set
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: writesame_unmap_reqs
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: writesame_unmap_requests
      template: conf/restperf/9.12.0/lun.yaml
      Unit: none
      Type: rate

  - Harvest Metric: lun_xcopy_reqs
    Description: Total number of xcopy operations on the LUN
    ZAPI:
      endpoint: perf-object-get-instances lun
      metric: xcopy_reqs
      template: conf/zapiperf/cdot/9.8.0/lun.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/lun
      metric: xcopy_requests
      template: conf/restperf/9.12.0/lun.yaml
      Unit: none
      Type: rate

  - Harvest Metric: namespace_avg_other_latency
    Description: Average other ops latency in microseconds for all operations on the Namespace
    ZAPI:
      endpoint: perf-object-get-instances namespace
      metric: avg_other_latency
      template: conf/zapiperf/cdot/9.10.1/namespace.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/namespace
      metric: average_other_latency
      template: conf/restperf/9.12.0/namespace.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: namespace_avg_read_latency
    Description: Average read latency in microseconds for all operations on the Namespace
    ZAPI:
      endpoint: perf-object-get-instances namespace
      metric: avg_read_latency
      template: conf/zapiperf/cdot/9.10.1/namespace.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/namespace
      metric: average_read_latency
      template: conf/restperf/9.12.0/namespace.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: namespace_avg_write_latency
    Description: Average write latency in microseconds for all operations on the Namespace
    ZAPI:
      endpoint: perf-object-get-instances namespace
      metric: avg_write_latency
      template: conf/zapiperf/cdot/9.10.1/namespace.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/namespace
      metric: average_write_latency
      template: conf/restperf/9.12.0/namespace.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: namespace_other_ops
    Description: Number of other operations
    ZAPI:
      endpoint: perf-object-get-instances namespace
      metric: other_ops
      template: conf/zapiperf/cdot/9.10.1/namespace.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/namespace
      metric: other_ops
      template: conf/restperf/9.12.0/namespace.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: namespace_read_data
    Description: Read bytes
    ZAPI:
      endpoint: perf-object-get-instances namespace
      metric: read_data
      template: conf/zapiperf/cdot/9.10.1/namespace.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/namespace
      metric: read_data
      template: conf/restperf/9.12.0/namespace.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: namespace_read_ops
    Description: Number of read operations
    ZAPI:
      endpoint: perf-object-get-instances namespace
      metric: read_ops
      template: conf/zapiperf/cdot/9.10.1/namespace.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/namespace
      metric: read_ops
      template: conf/restperf/9.12.0/namespace.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: namespace_remote_bytes
    Description: Remote read bytes
    ZAPI:
      endpoint: perf-object-get-instances namespace
      metric: remote_bytes
      template: conf/zapiperf/cdot/9.10.1/namespace.yaml
    REST:
      endpoint: api/cluster/counter/tables/namespace
      metric: remote.read_data
      template: conf/restperf/9.12.0/namespace.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: namespace_remote_ops
    Description: Number of remote read operations
    ZAPI:
      endpoint: perf-object-get-instances namespace
      metric: remote_ops
      template: conf/zapiperf/cdot/9.10.1/namespace.yaml
    REST:
      endpoint: api/cluster/counter/tables/namespace
      metric: remote.read_ops
      template: conf/restperf/9.12.0/namespace.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: namespace_write_data
    Description: Write bytes
    ZAPI:
      endpoint: perf-object-get-instances namespace
      metric: write_data
      template: conf/zapiperf/cdot/9.10.1/namespace.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/namespace
      metric: write_data
      template: conf/restperf/9.12.0/namespace.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: namespace_write_ops
    Description: Number of write operations
    ZAPI:
      endpoint: perf-object-get-instances namespace
      metric: write_ops
      template: conf/zapiperf/cdot/9.10.1/namespace.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/namespace
      metric: write_ops
      template: conf/restperf/9.12.0/namespace.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: net_port_mtu
    Description: Maximum transmission unit, largest packet size on this network
    ZAPI:
      endpoint: net-port-get-iter
      metric: net-port-info.mtu
      template: conf/zapi/cdot/9.8.0/netPort.yaml
    REST:
      endpoint: api/network/ethernet/ports
      metric: mtu
      template: conf/rest/9.12.0/netPort.yaml

  - Harvest Metric: netstat_bytes_recvd
    Description: Number of bytes received by a TCP connection
    ZAPI:
      endpoint: perf-object-get-instances netstat
      metric: bytes_recvd
      template: conf/zapiperf/cdot/9.8.0/netstat.yaml
      Unit: none
      Type: raw

  - Harvest Metric: netstat_bytes_sent
    Description: Number of bytes sent by a TCP connection
    ZAPI:
      endpoint: perf-object-get-instances netstat
      metric: bytes_sent
      template: conf/zapiperf/cdot/9.8.0/netstat.yaml
      Unit: none
      Type: raw

  - Harvest Metric: netstat_cong_win
    Description: Congestion window of a TCP connection
    ZAPI:
      endpoint: perf-object-get-instances netstat
      metric: cong_win
      template: conf/zapiperf/cdot/9.8.0/netstat.yaml
      Unit: none
      Type: raw

  - Harvest Metric: netstat_cong_win_th
    Description: Congestion window threshold of a TCP connection
    ZAPI:
      endpoint: perf-object-get-instances netstat
      metric: cong_win_th
      template: conf/zapiperf/cdot/9.8.0/netstat.yaml
      Unit: none
      Type: raw

  - Harvest Metric: netstat_ooorcv_pkts
    Description: Number of out-of-order packets received by this TCP connection
    ZAPI:
      endpoint: perf-object-get-instances netstat
      metric: ooorcv_pkts
      template: conf/zapiperf/cdot/9.8.0/netstat.yaml
      Unit: none
      Type: raw

  - Harvest Metric: netstat_recv_window
    Description: Receive window size of a TCP connection
    ZAPI:
      endpoint: perf-object-get-instances netstat
      metric: recv_window
      template: conf/zapiperf/cdot/9.8.0/netstat.yaml
      Unit: none
      Type: raw

  - Harvest Metric: netstat_rexmit_pkts
    Description: Number of packets retransmitted by this TCP connection
    ZAPI:
      endpoint: perf-object-get-instances netstat
      metric: rexmit_pkts
      template: conf/zapiperf/cdot/9.8.0/netstat.yaml
      Unit: none
      Type: raw

  - Harvest Metric: netstat_send_window
    Description: Send window size of a TCP connection
    ZAPI:
      endpoint: perf-object-get-instances netstat
      metric: send_window
      template: conf/zapiperf/cdot/9.8.0/netstat.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_clients_idle_duration
    Description: Specifies an ISO-8601 format of date and time to retrieve the idle time duration in hours, minutes, and seconds format.
    REST:
      endpoint: api/protocols/nfs/connected-clients
      metric: idle_duration
      template: conf/rest/9.7.0/nfs_clients.yaml

  - Harvest Metric: nfs_diag_storePool_ByteLockAlloc
    Description: Current number of byte range lock objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_ByteLockAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.byte_lock_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_ByteLockMax
    Description: Maximum number of byte range lock objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_ByteLockMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.byte_lock_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_ClientAlloc
    Description: Current number of client objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_ClientAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.client_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_ClientMax
    Description: Maximum number of client objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_ClientMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.client_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_ConnectionParentSessionReferenceAlloc
    Description: Current number of connection parent session reference objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_ConnectionParentSessionReferenceAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.connection_parent_session_reference_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_ConnectionParentSessionReferenceMax
    Description: Maximum number of connection parent session reference objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_ConnectionParentSessionReferenceMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.connection_parent_session_reference_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_CopyStateAlloc
    Description: Current number of copy state objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_CopyStateAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.copy_state_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_CopyStateMax
    Description: Maximum number of copy state objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_CopyStateMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.copy_state_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_DelegAlloc
    Description: Current number of delegation lock objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_DelegAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.delegation_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_DelegMax
    Description: Maximum number delegation lock objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_DelegMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.delegation_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_DelegStateAlloc
    Description: Current number of delegation state objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_DelegStateAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.delegation_state_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_DelegStateMax
    Description: Maximum number of delegation state objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_DelegStateMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.delegation_state_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_LayoutAlloc
    Description: Current number of layout objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_LayoutAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.layout_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_LayoutMax
    Description: Maximum number of layout objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_LayoutMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.layout_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_LayoutStateAlloc
    Description: Current number of layout state objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_LayoutStateAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.layout_state_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_LayoutStateMax
    Description: Maximum number of layout state objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_LayoutStateMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.layout_state_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_LockStateAlloc
    Description: Current number of lock state objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_LockStateAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.lock_state_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_LockStateMax
    Description: Maximum number of lock state objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_LockStateMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.lock_state_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_OpenAlloc
    Description: Current number of share objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_OpenAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.open_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_OpenMax
    Description: Maximum number of share lock objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_OpenMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.open_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_OpenStateAlloc
    Description: Current number of open state objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_OpenStateAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.openstate_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_OpenStateMax
    Description: Maximum number of open state objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_OpenStateMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.openstate_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_OwnerAlloc
    Description: Current number of owner objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_OwnerAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.owner_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_OwnerMax
    Description: Maximum number of owner objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_OwnerMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.owner_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_SessionAlloc
    Description: Current number of session objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_SessionAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.session_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_SessionConnectionHolderAlloc
    Description: Current number of session connection holder objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_SessionConnectionHolderAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.session_connection_holder_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_SessionConnectionHolderMax
    Description: Maximum number of session connection holder objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_SessionConnectionHolderMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.session_connection_holder_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_SessionHolderAlloc
    Description: Current number of session holder objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_SessionHolderAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.session_holder_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_SessionHolderMax
    Description: Maximum number of session holder objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_SessionHolderMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.session_holder_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_SessionMax
    Description: Maximum number of session objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_SessionMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.session_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_StateRefHistoryAlloc
    Description: Current number of state reference callstack history objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_StateRefHistoryAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.state_reference_history_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_StateRefHistoryMax
    Description: Maximum number of state reference callstack history objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_StateRefHistoryMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.state_reference_history_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_StringAlloc
    Description: Current number of string objects allocated.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_StringAlloc
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.string_allocated
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nfs_diag_storePool_StringMax
    Description: Maximum number of string objects.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_diag
      metric: storePool_StringMax
      template: conf/zapiperf/cdot/9.8.0/nfsv4_pool.yaml
      Unit: none
      Type: raw,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/nfs_v4_diag
      metric: storepool.string_maximum
      template: conf/restperf/9.12.0/nfsv4_pool.yaml
      Unit: none
      Type: raw

  - Harvest Metric: nic_link_up_to_downs
    Description: Number of link state change from UP to DOWN.
    ZAPI:
      endpoint: perf-object-get-instances nic_common
      metric: link_up_to_downs
      template: conf/zapiperf/cdot/9.8.0/nic_common.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/nic_common
      metric: link_up_to_down
      template: conf/restperf/9.12.0/nic_common.yaml
      Unit: none
      Type: delta

  - Harvest Metric: nic_rx_alignment_errors
    Description: Alignment errors detected on received packets
    ZAPI:
      endpoint: perf-object-get-instances nic_common
      metric: rx_alignment_errors
      template: conf/zapiperf/cdot/9.8.0/nic_common.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/nic_common
      metric: receive_alignment_errors
      template: conf/restperf/9.12.0/nic_common.yaml
      Unit: none
      Type: delta

  - Harvest Metric: nic_rx_bytes
    Description: Bytes received
    ZAPI:
      endpoint: perf-object-get-instances nic_common
      metric: rx_bytes
      template: conf/zapiperf/cdot/9.8.0/nic_common.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/nic_common
      metric: receive_bytes
      template: conf/restperf/9.12.0/nic_common.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: nic_rx_crc_errors
    Description: CRC errors detected on received packets
    ZAPI:
      endpoint: perf-object-get-instances nic_common
      metric: rx_crc_errors
      template: conf/zapiperf/cdot/9.8.0/nic_common.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/nic_common
      metric: receive_crc_errors
      template: conf/restperf/9.12.0/nic_common.yaml
      Unit: none
      Type: delta

  - Harvest Metric: nic_rx_errors
    Description: Error received
    ZAPI:
      endpoint: perf-object-get-instances nic_common
      metric: rx_errors
      template: conf/zapiperf/cdot/9.8.0/nic_common.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/nic_common
      metric: receive_errors
      template: conf/restperf/9.12.0/nic_common.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: nic_rx_length_errors
    Description: Length errors detected on received packets
    ZAPI:
      endpoint: perf-object-get-instances nic_common
      metric: rx_length_errors
      template: conf/zapiperf/cdot/9.8.0/nic_common.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/nic_common
      metric: receive_length_errors
      template: conf/restperf/9.12.0/nic_common.yaml
      Unit: none
      Type: delta

  - Harvest Metric: nic_rx_total_errors
    Description: Total errors received
    ZAPI:
      endpoint: perf-object-get-instances nic_common
      metric: rx_total_errors
      template: conf/zapiperf/cdot/9.8.0/nic_common.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/nic_common
      metric: receive_total_errors
      template: conf/restperf/9.12.0/nic_common.yaml
      Unit: none
      Type: delta

  - Harvest Metric: nic_tx_bytes
    Description: Bytes sent
    ZAPI:
      endpoint: perf-object-get-instances nic_common
      metric: tx_bytes
      template: conf/zapiperf/cdot/9.8.0/nic_common.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/nic_common
      metric: transmit_bytes
      template: conf/restperf/9.12.0/nic_common.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: nic_tx_errors
    Description: Error sent
    ZAPI:
      endpoint: perf-object-get-instances nic_common
      metric: tx_errors
      template: conf/zapiperf/cdot/9.8.0/nic_common.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/nic_common
      metric: transmit_errors
      template: conf/restperf/9.12.0/nic_common.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: nic_tx_hw_errors
    Description: Transmit errors reported by hardware
    ZAPI:
      endpoint: perf-object-get-instances nic_common
      metric: tx_hw_errors
      template: conf/zapiperf/cdot/9.8.0/nic_common.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/nic_common
      metric: transmit_hw_errors
      template: conf/restperf/9.12.0/nic_common.yaml
      Unit: none
      Type: delta

  - Harvest Metric: nic_tx_total_errors
    Description: Total errors sent
    ZAPI:
      endpoint: perf-object-get-instances nic_common
      metric: tx_total_errors
      template: conf/zapiperf/cdot/9.8.0/nic_common.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/nic_common
      metric: transmit_total_errors
      template: conf/restperf/9.12.0/nic_common.yaml
      Unit: none
      Type: delta

  - Harvest Metric: node_avg_processor_busy
    Description: Average processor utilization across all processors in the system
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: avg_processor_busy
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: average_processor_busy_percent
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: node_cifs_connections
    Description: Number of connections
    ZAPI:
      endpoint: perf-object-get-instances cifs:node
      metric: connections
      template: conf/zapiperf/cdot/9.8.0/cifs_node.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs:node
      metric: connections
      template: conf/restperf/9.12.0/cifs_node.yaml
      Unit: none
      Type: raw

  - Harvest Metric: node_cifs_established_sessions
    Description: Number of established SMB and SMB2 sessions
    ZAPI:
      endpoint: perf-object-get-instances cifs:node
      metric: established_sessions
      template: conf/zapiperf/cdot/9.8.0/cifs_node.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs:node
      metric: established_sessions
      template: conf/restperf/9.12.0/cifs_node.yaml
      Unit: none
      Type: raw

  - Harvest Metric: node_cifs_latency
    Description: Average latency for CIFS operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:node
      metric: cifs_latency
      template: conf/zapiperf/cdot/9.8.0/cifs_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs:node
      metric: latency
      template: conf/restperf/9.12.0/cifs_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_cifs_op_count
    Description: Array of select CIFS operation counts
    ZAPI:
      endpoint: perf-object-get-instances cifs:node
      metric: cifs_op_count
      template: conf/zapiperf/cdot/9.8.0/cifs_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs:node
      metric: op_count
      template: conf/restperf/9.12.0/cifs_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_cifs_open_files
    Description: Number of open files over SMB and SMB2
    ZAPI:
      endpoint: perf-object-get-instances cifs:node
      metric: open_files
      template: conf/zapiperf/cdot/9.8.0/cifs_node.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs:node
      metric: open_files
      template: conf/restperf/9.12.0/cifs_node.yaml
      Unit: none
      Type: raw

  - Harvest Metric: node_cifs_ops
    Description: Number of CIFS operations per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: cifs_ops
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: cifs_ops
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_cifs_read_latency
    Description: Average latency for CIFS read operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:node
      metric: cifs_read_latency
      template: conf/zapiperf/cdot/9.8.0/cifs_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs:node
      metric: average_read_latency
      template: conf/restperf/9.12.0/cifs_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_cifs_read_ops
    Description: Total number of CIFS read operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:node
      metric: cifs_read_ops
      template: conf/zapiperf/cdot/9.8.0/cifs_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs:node
      metric: total_read_ops
      template: conf/restperf/9.12.0/cifs_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_cifs_total_ops
    Description: Total number of CIFS operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:node
      metric: cifs_ops
      template: conf/zapiperf/cdot/9.8.0/cifs_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs:node
      metric: total_ops
      template: conf/restperf/9.12.0/cifs_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_cifs_write_latency
    Description: Average latency for CIFS write operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:node
      metric: cifs_write_latency
      template: conf/zapiperf/cdot/9.8.0/cifs_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs:node
      metric: average_write_latency
      template: conf/restperf/9.12.0/cifs_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_cifs_write_ops
    Description: Total number of CIFS write operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:node
      metric: cifs_write_ops
      template: conf/zapiperf/cdot/9.8.0/cifs_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs:node
      metric: total_write_ops
      template: conf/restperf/9.12.0/cifs_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_cpu_busy
    Description: System CPU resource utilization. Returns a computed percentage for the default CPU field. Basically computes a 'cpu usage summary' value which indicates how 'busy' the system is based upon the most heavily utilized domain. The idea is to determine the amount of available CPU until we're limited by either a domain maxing out OR we exhaust all available idle CPU cycles, whichever occurs first.
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: cpu_busy
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: cpu_busy
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: node_cpu_busytime
    Description: The time (in hundredths of a second) that the CPU has been doing useful work since the last boot
    ZAPI:
      endpoint: system-node-get-iter
      metric: node-details-info.cpu-busytime
      template: conf/zapi/cdot/9.8.0/node.yaml
    REST:
      endpoint: api/private/cli/node
      metric: cpu_busy_time
      template: conf/rest/9.12.0/node.yaml

  - Harvest Metric: node_cpu_domain_busy
    Description: Array of processor time in percentage spent in various domains
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: domain_busy
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: domain_busy
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: node_cpu_elapsed_time
    Description: Elapsed time since boot
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: cpu_elapsed_time
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: none
      Type: delta,no-display
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: cpu_elapsed_time
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: microsec
      Type: delta

  - Harvest Metric: node_disk_data_read
    Description: Number of disk kilobytes (KB) read per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: disk_data_read
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: disk_data_read
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_disk_data_written
    Description: Number of disk kilobytes (KB) written per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: disk_data_written
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: disk_data_written
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_failed_fan
    Description: Specifies a count of the number of chassis fans that are not operating within the recommended RPM range.
    ZAPI:
      endpoint: system-node-get-iter
      metric: node-details-info.env-failed-fan-count
      template: conf/zapi/cdot/9.8.0/node.yaml
    REST:
      endpoint: api/cluster/nodes
      metric: controller.failed_fan.count
      template: conf/rest/9.12.0/node.yaml

  - Harvest Metric: node_failed_power
    Description: Number of failed power supply units.
    ZAPI:
      endpoint: system-node-get-iter
      metric: node-details-info.env-failed-power-supply-count
      template: conf/zapi/cdot/9.8.0/node.yaml
    REST:
      endpoint: api/cluster/nodes
      metric: controller.failed_power_supply.count
      template: conf/rest/9.12.0/node.yaml

  - Harvest Metric: node_fcp_data_recv
    Description: Number of FCP kilobytes (KB) received per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: fcp_data_recv
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: fcp_data_received
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_fcp_data_sent
    Description: Number of FCP kilobytes (KB) sent per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: fcp_data_sent
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: fcp_data_sent
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_fcp_ops
    Description: Number of FCP operations per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: fcp_ops
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: fcp_ops
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_hdd_data_read
    Description: Number of HDD Disk kilobytes (KB) read per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: hdd_data_read
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: hdd_data_read
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_hdd_data_written
    Description: Number of HDD kilobytes (KB) written per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: hdd_data_written
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: hdd_data_written
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_iscsi_ops
    Description: Number of iSCSI operations per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: iscsi_ops
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: iscsi_ops
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_memory
    Description: Total memory in megabytes (MB)
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: memory
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: memory
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: none
      Type: raw

  - Harvest Metric: node_net_data_recv
    Description: Number of network kilobytes (KB) received per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: net_data_recv
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: network_data_received
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_net_data_sent
    Description: Number of network kilobytes (KB) sent per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: net_data_sent
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: network_data_sent
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_nfs_access_avg_latency
    Description: Average latency of ACCESS procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: access_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: access.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_access_total
    Description: Total number of ACCESS procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: access_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: access.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_backchannel_ctl_avg_latency
    Description: Average latency of NFSv4.2 BACKCHANNEL_CTL operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: backchannel_ctl_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: backchannel_ctl.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_backchannel_ctl_total
    Description: Total number of NFSv4.2 BACKCHANNEL_CTL operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: backchannel_ctl_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: backchannel_ctl.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_bind_conn_to_session_avg_latency
    Description: Average latency of NFSv4.2 BIND_CONN_TO_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: bind_conn_to_session_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: bind_conn_to_session.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_bind_conn_to_session_total
    Description: Total number of NFSv4.2 BIND_CONN_TO_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: bind_conn_to_session_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: bind_conn_to_session.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: delta

  - Harvest Metric: node_nfs_close_avg_latency
    Description: Average latency of CLOSE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: close_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: close.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_close_total
    Description: Total number of CLOSE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: close_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: close.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_commit_avg_latency
    Description: Average latency of COMMIT procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: commit_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: commit.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_commit_total
    Description: Total number of COMMIT procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: commit_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: commit.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_create_avg_latency
    Description: Average latency of CREATE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: create_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: create.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_create_session_avg_latency
    Description: Average latency of NFSv4.2 CREATE_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: create_session_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: create_session.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_create_session_total
    Description: Total number of NFSv4.2 CREATE_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: create_session_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: create_session.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_create_total
    Description: Total number of CREATE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: create_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: create.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_delegpurge_avg_latency
    Description: Average latency of DELEGPURGE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: delegpurge_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: delegpurge.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_delegpurge_total
    Description: Total number of DELEGPURGE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: delegpurge_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: delegpurge.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_delegreturn_avg_latency
    Description: Average latency of DELEGRETURN procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: delegreturn_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: delegreturn.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_delegreturn_total
    Description: Total number of DELEGRETURN procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: delegreturn_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: delegreturn.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_destroy_clientid_avg_latency
    Description: Average latency of NFSv4.2 DESTROY_CLIENTID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: destroy_clientid_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: destroy_clientid.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_destroy_clientid_total
    Description: Total number of NFSv4.2 DESTROY_CLIENTID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: destroy_clientid_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: destroy_clientid.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_destroy_session_avg_latency
    Description: Average latency of NFSv4.2 DESTROY_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: destroy_session_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: destroy_session.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_destroy_session_total
    Description: Total number of NFSv4.2 DESTROY_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: destroy_session_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: destroy_session.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_exchange_id_avg_latency
    Description: Average latency of NFSv4.2 EXCHANGE_ID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: exchange_id_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: exchange_id.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_exchange_id_total
    Description: Total number of NFSv4.2 EXCHANGE_ID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: exchange_id_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: exchange_id.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_free_stateid_avg_latency
    Description: Average latency of NFSv4.2 FREE_STATEID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: free_stateid_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: free_stateid.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_free_stateid_total
    Description: Total number of NFSv4.2 FREE_STATEID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: free_stateid_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: free_stateid.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_fsinfo_avg_latency
    Description: Average latency of FSInfo procedure requests. The counter keeps track of the average response time of FSInfo requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: fsinfo_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: fsinfo.average_latency
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_fsinfo_total
    Description: Total number FSInfo of procedure requests. It is the total number of FSInfo success and FSInfo error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: fsinfo_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: fsinfo.total
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_fsstat_avg_latency
    Description: Average latency of FSStat procedure requests. The counter keeps track of the average response time of FSStat requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: fsstat_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: fsstat.average_latency
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_fsstat_total
    Description: Total number FSStat of procedure requests. It is the total number of FSStat success and FSStat error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: fsstat_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: fsstat.total
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_get_dir_delegation_avg_latency
    Description: Average latency of NFSv4.2 GET_DIR_DELEGATION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: get_dir_delegation_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: get_dir_delegation.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_get_dir_delegation_total
    Description: Total number of NFSv4.2 GET_DIR_DELEGATION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: get_dir_delegation_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: get_dir_delegation.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_getattr_avg_latency
    Description: Average latency of GETATTR procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: getattr_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: getattr.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_getattr_total
    Description: Total number of GETATTR procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: getattr_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: getattr.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_getdeviceinfo_avg_latency
    Description: Average latency of NFSv4.2 GETDEVICEINFO operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: getdeviceinfo_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: getdeviceinfo.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_getdeviceinfo_total
    Description: Total number of NFSv4.2 GETDEVICEINFO operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: getdeviceinfo_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: getdeviceinfo.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_getdevicelist_avg_latency
    Description: Average latency of NFSv4.2 GETDEVICELIST operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: getdevicelist_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: getdevicelist.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_getdevicelist_total
    Description: Total number of NFSv4.2 GETDEVICELIST operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: getdevicelist_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: getdevicelist.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_getfh_avg_latency
    Description: Average latency of GETFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: getfh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: getfh.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_getfh_total
    Description: Total number of GETFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: getfh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: getfh.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_latency
    Description: Average latency of NFSv4 requests. This counter keeps track of the average response time of NFSv4 requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_layoutcommit_avg_latency
    Description: Average latency of NFSv4.2 LAYOUTCOMMIT operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: layoutcommit_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: layoutcommit.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_layoutcommit_total
    Description: Total number of NFSv4.2 LAYOUTCOMMIT operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: layoutcommit_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: layoutcommit.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_layoutget_avg_latency
    Description: Average latency of NFSv4.2 LAYOUTGET operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: layoutget_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: layoutget.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_layoutget_total
    Description: Total number of NFSv4.2 LAYOUTGET operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: layoutget_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: layoutget.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_layoutreturn_avg_latency
    Description: Average latency of NFSv4.2 LAYOUTRETURN operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: layoutreturn_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: layoutreturn.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_layoutreturn_total
    Description: Total number of NFSv4.2 LAYOUTRETURN operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: layoutreturn_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: layoutreturn.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_link_avg_latency
    Description: Average latency of LINK procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: link_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: link.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_link_total
    Description: Total number of LINK procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: link_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: link.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_lock_avg_latency
    Description: Average latency of LOCK procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: lock_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: lock.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_lock_total
    Description: Total number of LOCK procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: lock_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: lock.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_lockt_avg_latency
    Description: Average latency of LOCKT procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: lockt_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: lockt.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_lockt_total
    Description: Total number of LOCKT procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: lockt_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: lockt.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_locku_avg_latency
    Description: Average latency of LOCKU procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: locku_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: locku.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_locku_total
    Description: Total number of LOCKU procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: locku_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: locku.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_lookup_avg_latency
    Description: Average latency of LOOKUP procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: lookup_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: lookup.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_lookup_total
    Description: Total number of LOOKUP procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: lookup_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: lookup.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_lookupp_avg_latency
    Description: Average latency of LOOKUPP procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: lookupp_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: lookupp.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_lookupp_total
    Description: Total number of LOOKUPP procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: lookupp_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: lookupp.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_mkdir_avg_latency
    Description: Average latency of MkDir procedure requests. The counter keeps track of the average response time of MkDir requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: mkdir_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: mkdir.average_latency
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_mkdir_total
    Description: Total number MkDir of procedure requests. It is the total number of MkDir success and MkDir error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: mkdir_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: mkdir.total
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_mknod_avg_latency
    Description: Average latency of MkNod procedure requests. The counter keeps track of the average response time of MkNod requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: mknod_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: mknod.average_latency
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_mknod_total
    Description: Total number MkNod of procedure requests. It is the total number of MkNod success and MkNod error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: mknod_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: mknod.total
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_null_avg_latency
    Description: Average Latency of NULL procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: null_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: null.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_null_total
    Description: Total number of NULL procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: null_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: null.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_nverify_avg_latency
    Description: Average latency of NVERIFY procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: nverify_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: nverify.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_nverify_total
    Description: Total number of NVERIFY procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: nverify_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: nverify.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_open_avg_latency
    Description: Average latency of OPEN procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: open_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: open.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_open_confirm_avg_latency
    Description: Average latency of OPEN_CONFIRM procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: open_confirm_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: open_confirm.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_open_confirm_total
    Description: Total number of OPEN_CONFIRM procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: open_confirm_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: open_confirm.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_open_downgrade_avg_latency
    Description: Average latency of OPEN_DOWNGRADE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: open_downgrade_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: open_downgrade.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_open_downgrade_total
    Description: Total number of OPEN_DOWNGRADE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: open_downgrade_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: open_downgrade.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_open_total
    Description: Total number of OPEN procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: open_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: open.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_openattr_avg_latency
    Description: Average latency of OPENATTR procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: openattr_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: openattr.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_openattr_total
    Description: Total number of OPENATTR procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: openattr_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: openattr.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_ops
    Description: Number of NFS operations per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: nfs_ops
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: nfs_ops
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_nfs_pathconf_avg_latency
    Description: Average latency of PathConf procedure requests. The counter keeps track of the average response time of PathConf requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: pathconf_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: pathconf.average_latency
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_pathconf_total
    Description: Total number PathConf of procedure requests. It is the total number of PathConf success and PathConf error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: pathconf_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: pathconf.total
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_putfh_avg_latency
    Description: Average latency of PUTFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: putfh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: putfh.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_putfh_total
    Description: Total number of PUTFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: putfh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: putfh.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_putpubfh_avg_latency
    Description: Average latency of PUTPUBFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: putpubfh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: putpubfh.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_putpubfh_total
    Description: Total number of PUTPUBFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: putpubfh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: putpubfh.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_putrootfh_avg_latency
    Description: Average latency of PUTROOTFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: putrootfh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: putrootfh.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_putrootfh_total
    Description: Total number of PUTROOTFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: putrootfh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: putrootfh.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_read_avg_latency
    Description: Average latency of READ procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: read_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: read.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_read_ops
    Description: Total observed NFSv3 read operations per second.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: nfsv3_read_ops
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: read_ops
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_nfs_read_symlink_avg_latency
    Description: Average latency of ReadSymLink procedure requests. The counter keeps track of the average response time of ReadSymLink requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: read_symlink_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: read_symlink.average_latency
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_read_symlink_total
    Description: Total number of ReadSymLink procedure requests. It is the total number of read symlink success and read symlink error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: read_symlink_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: read_symlink.total
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: none
      Type: delta

  - Harvest Metric: node_nfs_read_throughput
    Description: NFSv4 read data transfers
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: nfs4_read_throughput
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: total.read_throughput
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_nfs_read_total
    Description: Total number of READ procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: read_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: read.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_readdir_avg_latency
    Description: Average latency of READDIR procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: readdir_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: readdir.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_readdir_total
    Description: Total number of READDIR procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: readdir_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: readdir.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_readdirplus_avg_latency
    Description: Average latency of ReadDirPlus procedure requests. The counter keeps track of the average response time of ReadDirPlus requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: readdirplus_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: readdirplus.average_latency
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_readdirplus_total
    Description: Total number ReadDirPlus of procedure requests. It is the total number of ReadDirPlus success and ReadDirPlus error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: readdirplus_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: readdirplus.total
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_readlink_avg_latency
    Description: Average latency of READLINK procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: readlink_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: readlink.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_readlink_total
    Description: Total number of READLINK procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: readlink_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: readlink.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_reclaim_complete_avg_latency
    Description: Average latency of NFSv4.2 RECLAIM_complete operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: reclaim_complete_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: reclaim_complete.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_reclaim_complete_total
    Description: Total number of NFSv4.2 RECLAIM_complete operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: reclaim_complete_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: reclaim_complete.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_release_lock_owner_avg_latency
    Description: Average Latency of RELEASE_LOCKOWNER procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: release_lock_owner_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: release_lock_owner.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_release_lock_owner_total
    Description: Total number of RELEASE_LOCKOWNER procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: release_lock_owner_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: release_lock_owner.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_remove_avg_latency
    Description: Average latency of REMOVE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: remove_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: remove.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_remove_total
    Description: Total number of REMOVE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: remove_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: remove.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_rename_avg_latency
    Description: Average latency of RENAME procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: rename_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: rename.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_rename_total
    Description: Total number of RENAME procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: rename_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: rename.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_renew_avg_latency
    Description: Average latency of RENEW procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: renew_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: renew.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_renew_total
    Description: Total number of RENEW procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: renew_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: renew.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_restorefh_avg_latency
    Description: Average latency of RESTOREFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: restorefh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: restorefh.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_restorefh_total
    Description: Total number of RESTOREFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: restorefh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: restorefh.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_rmdir_avg_latency
    Description: Average latency of RmDir procedure requests. The counter keeps track of the average response time of RmDir requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: rmdir_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: rmdir.average_latency
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_rmdir_total
    Description: Total number RmDir of procedure requests. It is the total number of RmDir success and RmDir error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: rmdir_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: rmdir.total
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_savefh_avg_latency
    Description: Average latency of SAVEFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: savefh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: savefh.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_savefh_total
    Description: Total number of SAVEFH procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: savefh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: savefh.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_secinfo_avg_latency
    Description: Average latency of SECINFO procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: secinfo_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: secinfo.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_secinfo_no_name_avg_latency
    Description: Average latency of NFSv4.2 SECINFO_NO_NAME operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: secinfo_no_name_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: secinfo_no_name.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_secinfo_no_name_total
    Description: Total number of NFSv4.2 SECINFO_NO_NAME operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: secinfo_no_name_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: secinfo_no_name.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_secinfo_total
    Description: Total number of SECINFO procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: secinfo_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: secinfo.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_sequence_avg_latency
    Description: Average latency of NFSv4.2 SEQUENCE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: sequence_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: sequence.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_sequence_total
    Description: Total number of NFSv4.2 SEQUENCE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: sequence_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: sequence.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_set_ssv_avg_latency
    Description: Average latency of NFSv4.2 SET_SSV operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: set_ssv_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: set_ssv.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_set_ssv_total
    Description: Total number of NFSv4.2 SET_SSV operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: set_ssv_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: set_ssv.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_setattr_avg_latency
    Description: Average latency of SETATTR procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: setattr_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: setattr.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_setattr_total
    Description: Total number of SETATTR procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: setattr_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: setattr.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_setclientid_avg_latency
    Description: Average latency of SETCLIENTID procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: setclientid_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: setclientid.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_setclientid_confirm_avg_latency
    Description: Average latency of SETCLIENTID_CONFIRM procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: setclientid_confirm_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: setclientid_confirm.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_setclientid_confirm_total
    Description: Total number of SETCLIENTID_CONFIRM procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: setclientid_confirm_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: setclientid_confirm.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_setclientid_total
    Description: Total number of SETCLIENTID procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: setclientid_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: setclientid.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_symlink_avg_latency
    Description: Average latency of SymLink procedure requests. The counter keeps track of the average response time of SymLink requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: symlink_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: symlink.average_latency
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_symlink_total
    Description: Total number SymLink of procedure requests. It is the total number of SymLink success and create SymLink requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: symlink_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: symlink.total
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_test_stateid_avg_latency
    Description: Average latency of NFSv4.2 TEST_STATEID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: test_stateid_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: test_stateid.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_test_stateid_total
    Description: Total number of NFSv4.2 TEST_STATEID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: test_stateid_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: test_stateid.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_throughput
    Description: NFSv4 data transfers
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: nfs4_throughput
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: total.throughput
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_nfs_total_ops
    Description: Total number of NFSv4 requests per second.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: total_ops
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: total_ops
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_nfs_verify_avg_latency
    Description: Average latency of VERIFY procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: verify_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: verify.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_verify_total
    Description: Total number of VERIFY procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: verify_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: verify.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_want_delegation_avg_latency
    Description: Average latency of NFSv4.2 WANT_DELEGATION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: want_delegation_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: want_delegation.average_latency
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_want_delegation_total
    Description: Total number of NFSv4.2 WANT_DELEGATION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1:node
      metric: want_delegation_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42:node
      metric: want_delegation.total
      template: conf/restperf/9.12.0/nfsv4_2_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nfs_write_avg_latency
    Description: Average Latency of WRITE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: write_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: write.average_latency
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_nfs_write_ops
    Description: Total observed NFSv3 write operations per second.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3:node
      metric: nfsv3_write_ops
      template: conf/zapiperf/cdot/9.8.0/nfsv3_node.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3:node
      metric: write_ops
      template: conf/restperf/9.12.0/nfsv3_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_nfs_write_throughput
    Description: NFSv4 write data transfers
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: nfs4_write_throughput
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: total.write_throughput
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_nfs_write_total
    Description: Total number of WRITE procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4:node
      metric: write_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_node.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4:node
      metric: write.total
      template: conf/restperf/9.12.0/nfsv4_node.yaml
      Unit: none
      Type: rate

  - Harvest Metric: node_nvmf_data_recv
    Description: NVMe/FC kilobytes (KB) received per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: nvmf_data_recv
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: nvme_fc_data_received
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_nvmf_data_sent
    Description: NVMe/FC kilobytes (KB) sent per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: nvmf_data_sent
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: nvme_fc_data_sent
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_nvmf_ops
    Description: NVMe/FC operations per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: nvmf_ops
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: nvme_fc_ops
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_ssd_data_read
    Description: Number of SSD Disk kilobytes (KB) read per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: ssd_data_read
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: ssd_data_read
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_ssd_data_written
    Description: Number of SSD Disk kilobytes (KB) written per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: ssd_data_written
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: ssd_data_written
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: node_total_data
    Description: Total throughput in bytes
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: total_data
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: total_data
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: node_total_latency
    Description: Average latency for all operations in the system in microseconds
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: total_latency
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: total_latency
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_total_ops
    Description: Total number of operations per second
    ZAPI:
      endpoint: perf-object-get-instances system:node
      metric: total_ops
      template: conf/zapiperf/cdot/9.8.0/system_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/system:node
      metric: total_ops
      template: conf/restperf/9.12.0/system_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_uptime
    Description: The total time, in seconds, that the node has been up.
    ZAPI:
      endpoint: system-node-get-iter
      metric: node-details-info.node-uptime
      template: conf/zapi/cdot/9.8.0/node.yaml
    REST:
      endpoint: api/cluster/nodes
      metric: uptime
      template: conf/rest/9.12.0/node.yaml

  - Harvest Metric: node_vol_cifs_other_latency
    Description: Average time for the WAFL filesystem to process other CIFS operations to the volume; not including CIFS protocol request processing or network communication time which will also be included in client observed CIFS request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: cifs_other_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: cifs.other_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_cifs_other_ops
    Description: Number of other CIFS operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: cifs_other_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: cifs.other_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_cifs_read_data
    Description: Bytes read per second via CIFS
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: cifs_read_data
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: cifs.read_data
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: node_vol_cifs_read_latency
    Description: Average time for the WAFL filesystem to process CIFS read requests to the volume; not including CIFS protocol request processing or network communication time which will also be included in client observed CIFS request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: cifs_read_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: cifs.read_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_cifs_read_ops
    Description: Number of CIFS read operations per second from the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: cifs_read_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: cifs.read_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_cifs_write_data
    Description: Bytes written per second via CIFS
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: cifs_write_data
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: cifs.write_data
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: node_vol_cifs_write_latency
    Description: Average time for the WAFL filesystem to process CIFS write requests to the volume; not including CIFS protocol request processing or network communication time which will also be included in client observed CIFS request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: cifs_write_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: cifs.write_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_cifs_write_ops
    Description: Number of CIFS write operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: cifs_write_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: cifs.write_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_fcp_other_latency
    Description: Average time for the WAFL filesystem to process other FCP protocol operations to the volume; not including FCP protocol request processing or network communication time which will also be included in client observed FCP protocol request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: fcp_other_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: fcp.other_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_fcp_other_ops
    Description: Number of other block protocol operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: fcp_other_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: fcp.other_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_fcp_read_data
    Description: Bytes read per second via block protocol
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: fcp_read_data
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: fcp.read_data
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: node_vol_fcp_read_latency
    Description: Average time for the WAFL filesystem to process FCP protocol read operations to the volume; not including FCP protocol request processing or network communication time which will also be included in client observed FCP protocol request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: fcp_read_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: fcp.read_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_fcp_read_ops
    Description: Number of block protocol read operations per second from the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: fcp_read_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: fcp.read_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_fcp_write_data
    Description: Bytes written per second via block protocol
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: fcp_write_data
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: fcp.write_data
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: node_vol_fcp_write_latency
    Description: Average time for the WAFL filesystem to process FCP protocol write operations to the volume; not including FCP protocol request processing or network communication time which will also be included in client observed FCP protocol request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: fcp_write_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: fcp.write_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_fcp_write_ops
    Description: Number of block protocol write operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: fcp_write_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: fcp.write_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_iscsi_other_latency
    Description: Average time for the WAFL filesystem to process other iSCSI protocol operations to the volume; not including iSCSI protocol request processing or network communication time which will also be included in client observed iSCSI protocol request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: iscsi_other_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: iscsi.other_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_iscsi_other_ops
    Description: Number of other block protocol operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: iscsi_other_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: iscsi.other_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_iscsi_read_data
    Description: Bytes read per second via block protocol
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: iscsi_read_data
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: iscsi.read_data
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: node_vol_iscsi_read_latency
    Description: Average time for the WAFL filesystem to process iSCSI protocol read operations to the volume; not including iSCSI protocol request processing or network communication time which will also be included in client observed iSCSI protocol request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: iscsi_read_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: iscsi.read_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_iscsi_read_ops
    Description: Number of block protocol read operations per second from the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: iscsi_read_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: iscsi.read_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_iscsi_write_data
    Description: Bytes written per second via block protocol
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: iscsi_write_data
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: iscsi.write_data
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: node_vol_iscsi_write_latency
    Description: Average time for the WAFL filesystem to process iSCSI protocol write operations to the volume; not including iSCSI protocol request processing or network communication time which will also be included in client observed iSCSI request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: iscsi_write_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: iscsi.write_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_iscsi_write_ops
    Description: Number of block protocol write operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: iscsi_write_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: iscsi.write_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_nfs_other_latency
    Description: Average time for the WAFL filesystem to process other NFS operations to the volume; not including NFS protocol request processing or network communication time which will also be included in client observed NFS request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: nfs_other_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: nfs.other_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_nfs_other_ops
    Description: Number of other NFS operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: nfs_other_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: nfs.other_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_nfs_read_data
    Description: Bytes read per second via NFS
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: nfs_read_data
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: nfs.read_data
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: node_vol_nfs_read_latency
    Description: Average time for the WAFL filesystem to process NFS protocol read requests to the volume; not including NFS protocol request processing or network communication time which will also be included in client observed NFS request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: nfs_read_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: nfs.read_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_nfs_read_ops
    Description: Number of NFS read operations per second from the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: nfs_read_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: nfs.read_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_nfs_write_data
    Description: Bytes written per second via NFS
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: nfs_write_data
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: nfs.write_data
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: node_vol_nfs_write_latency
    Description: Average time for the WAFL filesystem to process NFS protocol write requests to the volume; not including NFS protocol request processing or network communication time, which will also be included in client observed NFS request latency
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: nfs_write_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: nfs.write_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_nfs_write_ops
    Description: Number of NFS write operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: nfs_write_ops
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: nfs.write_ops
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: node_vol_read_latency
    Description: Average latency in microseconds for the WAFL filesystem to process read request to the volume; not including request processing or network communication time
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: read_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: read_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: node_vol_write_latency
    Description: Average latency in microseconds for the WAFL filesystem to process write request to the volume; not including request processing or network communication time
    ZAPI:
      endpoint: perf-object-get-instances volume:node
      metric: write_latency
      template: conf/zapiperf/cdot/9.8.0/volume_node.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:node
      metric: write_latency
      template: conf/restperf/9.12.0/volume_node.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: nvme_lif_avg_latency
    Description: Average latency for NVMF operations
    ZAPI:
      endpoint: perf-object-get-instances nvmf_fc_lif
      metric: avg_latency
      template: conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/nvmf_lif
      metric: average_latency
      template: conf/restperf/9.12.0/nvmf_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: nvme_lif_avg_other_latency
    Description: Average latency for operations other than read, write, compare or compare-and-write.
    ZAPI:
      endpoint: perf-object-get-instances nvmf_fc_lif
      metric: avg_other_latency
      template: conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/nvmf_lif
      metric: average_other_latency
      template: conf/restperf/9.12.0/nvmf_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: nvme_lif_avg_read_latency
    Description: Average latency for read operations
    ZAPI:
      endpoint: perf-object-get-instances nvmf_fc_lif
      metric: avg_read_latency
      template: conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/nvmf_lif
      metric: average_read_latency
      template: conf/restperf/9.12.0/nvmf_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: nvme_lif_avg_write_latency
    Description: Average latency for write operations
    ZAPI:
      endpoint: perf-object-get-instances nvmf_fc_lif
      metric: avg_write_latency
      template: conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/nvmf_lif
      metric: average_write_latency
      template: conf/restperf/9.12.0/nvmf_lif.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: nvme_lif_other_ops
    Description: Number of operations that are not read, write, compare or compare-and-write.
    ZAPI:
      endpoint: perf-object-get-instances nvmf_fc_lif
      metric: other_ops
      template: conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/nvmf_lif
      metric: other_ops
      template: conf/restperf/9.12.0/nvmf_lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: nvme_lif_read_data
    Description: Amount of data read from the storage system
    ZAPI:
      endpoint: perf-object-get-instances nvmf_fc_lif
      metric: read_data
      template: conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/nvmf_lif
      metric: read_data
      template: conf/restperf/9.12.0/nvmf_lif.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: nvme_lif_read_ops
    Description: Number of read operations
    ZAPI:
      endpoint: perf-object-get-instances nvmf_fc_lif
      metric: read_ops
      template: conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/nvmf_lif
      metric: read_ops
      template: conf/restperf/9.12.0/nvmf_lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: nvme_lif_total_ops
    Description: Total number of operations.
    ZAPI:
      endpoint: perf-object-get-instances nvmf_fc_lif
      metric: total_ops
      template: conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/nvmf_lif
      metric: total_ops
      template: conf/restperf/9.12.0/nvmf_lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: nvme_lif_write_data
    Description: Amount of data written to the storage system
    ZAPI:
      endpoint: perf-object-get-instances nvmf_fc_lif
      metric: write_data
      template: conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/nvmf_lif
      metric: write_data
      template: conf/restperf/9.12.0/nvmf_lif.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: nvme_lif_write_ops
    Description: Number of write operations
    ZAPI:
      endpoint: perf-object-get-instances nvmf_fc_lif
      metric: write_ops
      template: conf/zapiperf/cdot/9.10.1/nvmf_lif.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/nvmf_lif
      metric: write_ops
      template: conf/restperf/9.12.0/nvmf_lif.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_abort_multipart_upload_failed
    Description: Number of failed Abort Multipart Upload operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: abort_multipart_upload_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_abort_multipart_upload_failed_client_close
    Description: Number of times Abort Multipart Upload operation failed because client terminated connection for operation pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: abort_multipart_upload_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_abort_multipart_upload_latency
    Description: Average latency for Abort Multipart Upload operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: abort_multipart_upload_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_abort_multipart_upload_rate
    Description: Number of Abort Multipart Upload operations per second.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: abort_multipart_upload_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_abort_multipart_upload_total
    Description: Number of Abort Multipart Upload operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: abort_multipart_upload_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_allow_access
    Description: Number of times access was allowed.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: allow_access
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_anonymous_access
    Description: Number of times anonymous access was allowed.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: anonymous_access
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_anonymous_deny_access
    Description: Number of times anonymous access was denied.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: anonymous_deny_access
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_authentication_failures
    Description: Number of authentication failures.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: authentication_failures
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_chunked_upload_reqs
    Description: Total number of object store server chunked object upload requests
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: chunked_upload_reqs
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_complete_multipart_upload_failed
    Description: Number of failed Complete Multipart Upload operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: complete_multipart_upload_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_complete_multipart_upload_failed_client_close
    Description: Number of times Complete Multipart Upload operation failed because client terminated connection for operation pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: complete_multipart_upload_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_complete_multipart_upload_latency
    Description: Average latency for Complete Multipart Upload operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: complete_multipart_upload_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_complete_multipart_upload_rate
    Description: Number of Complete Multipart Upload operations per second.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: complete_multipart_upload_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_complete_multipart_upload_total
    Description: Number of Complete Multipart Upload operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: complete_multipart_upload_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_connected_connections
    Description: Number of object store server connections currently established
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: connected_connections
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: raw

  - Harvest Metric: ontaps3_connections
    Description: Total number of object store server connections.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: connections
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_create_bucket_failed
    Description: Number of failed Create Bucket operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: create_bucket_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_create_bucket_failed_client_close
    Description: Number of times Create Bucket operation failed because client terminated connection for operation pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: create_bucket_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_create_bucket_latency
    Description: Average latency for Create Bucket operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: create_bucket_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average,no-zero-values

  - Harvest Metric: ontaps3_create_bucket_rate
    Description: Number of Create Bucket operations per second.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: create_bucket_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate,no-zero-values

  - Harvest Metric: ontaps3_create_bucket_total
    Description: Number of Create Bucket operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: create_bucket_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_default_deny_access
    Description: Number of times access was denied by default and not through any policy statement.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: default_deny_access
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_delete_bucket_failed
    Description: Number of failed Delete Bucket operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_bucket_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_delete_bucket_failed_client_close
    Description: Number of times Delete Bucket operation failed because client terminated connection for operation pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_bucket_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_delete_bucket_latency
    Description: Average latency for Delete Bucket operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_bucket_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average,no-zero-values

  - Harvest Metric: ontaps3_delete_bucket_rate
    Description: Number of Delete Bucket operations per second.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_bucket_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate,no-zero-values

  - Harvest Metric: ontaps3_delete_bucket_total
    Description: Number of Delete Bucket operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_bucket_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_delete_object_failed
    Description: Number of failed DELETE object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_object_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_delete_object_failed_client_close
    Description: Number of times DELETE object operation failed due to the case where client closed the connection while the operation was still pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_object_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_delete_object_latency
    Description: Average latency for DELETE object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_object_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_delete_object_rate
    Description: Number of DELETE object operations per sec
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_object_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_delete_object_tagging_failed
    Description: Number of failed DELETE object tagging operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_object_tagging_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_delete_object_tagging_failed_client_close
    Description: Number of times DELETE object tagging operation failed because client terminated connection for operation pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_object_tagging_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_delete_object_tagging_latency
    Description: Average latency for DELETE object tagging operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_object_tagging_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average,no-zero-values

  - Harvest Metric: ontaps3_delete_object_tagging_rate
    Description: Number of DELETE object tagging operations per sec.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_object_tagging_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate,no-zero-values

  - Harvest Metric: ontaps3_delete_object_tagging_total
    Description: Number of DELETE object tagging operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_object_tagging_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_delete_object_total
    Description: Number of DELETE object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: delete_object_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_explicit_deny_access
    Description: Number of times access was denied explicitly by a policy statement.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: explicit_deny_access
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_get_bucket_acl_failed
    Description: Number of failed GET Bucket ACL operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_bucket_acl_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_get_bucket_acl_total
    Description: Number of GET Bucket ACL operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_bucket_acl_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_get_bucket_versioning_failed
    Description: Number of failed Get Bucket Versioning operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_bucket_versioning_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_get_bucket_versioning_total
    Description: Number of Get Bucket Versioning operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_bucket_versioning_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_get_data
    Description: Rate of GET object data transfers per second
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_data
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: ontaps3_get_object_acl_failed
    Description: Number of failed GET Object ACL operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_acl_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_get_object_acl_total
    Description: Number of GET Object ACL operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_acl_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_get_object_failed
    Description: Number of failed GET object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_get_object_failed_client_close
    Description: Number of times GET object operation failed due to the case where client closed the connection while the operation was still pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_get_object_lastbyte_latency
    Description: Average last-byte latency for GET object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_lastbyte_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_get_object_latency
    Description: Average first-byte latency for GET object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_get_object_rate
    Description: Number of GET object operations per sec
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_get_object_tagging_failed
    Description: Number of failed GET object tagging operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_tagging_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_get_object_tagging_failed_client_close
    Description: Number of times GET object tagging operation failed due to the case where client closed the connection while the operation was still pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_tagging_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_get_object_tagging_latency
    Description: Average latency for GET object tagging operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_tagging_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_get_object_tagging_rate
    Description: Number of GET object tagging operations per sec
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_tagging_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_get_object_tagging_total
    Description: Number of GET object tagging operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_tagging_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_get_object_total
    Description: Number of GET object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: get_object_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_group_policy_evaluated
    Description: Number of times group policies were evaluated.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: group_policy_evaluated
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_head_bucket_failed
    Description: Number of failed HEAD bucket operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: head_bucket_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_head_bucket_failed_client_close
    Description: Number of times HEAD bucket operation failed due to the case where client closed the connection while the operation was still pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: head_bucket_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_head_bucket_latency
    Description: Average latency for HEAD bucket operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: head_bucket_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_head_bucket_rate
    Description: Number of HEAD bucket operations per sec
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: head_bucket_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_head_bucket_total
    Description: Number of HEAD bucket operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: head_bucket_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_head_object_failed
    Description: Number of failed HEAD Object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: head_object_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_head_object_failed_client_close
    Description: Number of times HEAD object operation failed due to the case where client closed the connection while the operation was still pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: head_object_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_head_object_latency
    Description: Average latency for HEAD object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: head_object_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_head_object_rate
    Description: Number of HEAD Object operations per sec
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: head_object_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_head_object_total
    Description: Number of HEAD Object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: head_object_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_initiate_multipart_upload_failed
    Description: Number of failed Initiate Multipart Upload operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: initiate_multipart_upload_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_initiate_multipart_upload_failed_client_close
    Description: Number of times Initiate Multipart Upload operation failed because client terminated connection for operation pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: initiate_multipart_upload_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_initiate_multipart_upload_latency
    Description: Average latency for Initiate Multipart Upload operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: initiate_multipart_upload_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_initiate_multipart_upload_rate
    Description: Number of Initiate Multipart Upload operations per second.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: initiate_multipart_upload_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_initiate_multipart_upload_total
    Description: Number of Initiate Multipart Upload operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: initiate_multipart_upload_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_input_flow_control_entry
    Description: Number of times input flow control was entered.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: input_flow_control_entry
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_input_flow_control_exit
    Description: Number of times input flow control was exited.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: input_flow_control_exit
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_list_buckets_failed
    Description: Number of failed LIST Buckets operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_buckets_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_list_buckets_failed_client_close
    Description: Number of times LIST Bucket operation failed due to the case where client closed the connection while the operation was still pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_buckets_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_list_buckets_latency
    Description: Average latency for LIST Buckets operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_buckets_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_list_buckets_rate
    Description: Number of LIST Buckets operations per sec
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_buckets_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_list_buckets_total
    Description: Number of LIST Buckets operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_buckets_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_list_object_versions_failed
    Description: Number of failed LIST object versions operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_object_versions_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_list_object_versions_failed_client_close
    Description: Number of times LIST object versions operation failed due to the case where client closed the connection while the operation was still pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_object_versions_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_list_object_versions_latency
    Description: Average latency for LIST Object versions operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_object_versions_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average,no-zero-values

  - Harvest Metric: ontaps3_list_object_versions_rate
    Description: Number of LIST Object Versions operations per sec
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_object_versions_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate,no-zero-values

  - Harvest Metric: ontaps3_list_object_versions_total
    Description: Number of LIST Object Versions operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_object_versions_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_list_objects_failed
    Description: Number of failed LIST objects operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_objects_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_list_objects_failed_client_close
    Description: Number of times LIST objects operation failed due to the case where client closed the connection while the operation was still pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_objects_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_list_objects_latency
    Description: Average latency for LIST Objects operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_objects_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_list_objects_rate
    Description: Number of LIST Objects operations per sec
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_objects_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_list_objects_total
    Description: Number of LIST Objects operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_objects_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_list_uploads_failed
    Description: Number of failed LIST Uploads operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_uploads_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_list_uploads_failed_client_close
    Description: Number of times LIST Uploads operation failed due to the case where client closed the connection while the operation was still pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_uploads_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_list_uploads_latency
    Description: Average latency for LIST Uploads operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_uploads_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_list_uploads_rate
    Description: Number of LIST Uploads operations per sec
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_uploads_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_list_uploads_total
    Description: Number of LIST Uploads operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: list_uploads_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_logical_used_size
    REST:
      endpoint: api/protocols/s3/buckets
      metric: logical_used_size
      template: conf/rest/9.7.0/ontap_s3.yaml

  - Harvest Metric: ontaps3_max_cmds_per_connection
    Description: Maximum commands pipelined at any instance on a connection.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: max_cmds_per_connection
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_max_connected_connections
    Description: Maximum number of object store server connections established at one time
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: max_connected_connections
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: raw

  - Harvest Metric: ontaps3_max_requests_outstanding
    Description: Maximum number of object store server requests in process at one time
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: max_requests_outstanding
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: raw

  - Harvest Metric: ontaps3_multi_delete_reqs
    Description: Total number of object store server multiple object delete requests
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: multi_delete_reqs
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_output_flow_control_entry
    Description: Number of output flow control was entered.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: output_flow_control_entry
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_output_flow_control_exit
    Description: Number of times output flow control was exited.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: output_flow_control_exit
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_presigned_url_reqs
    Description: Total number of presigned object store server URL requests.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: presigned_url_reqs
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_put_bucket_versioning_failed
    Description: Number of failed Put Bucket Versioning operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_bucket_versioning_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_put_bucket_versioning_total
    Description: Number of Put Bucket Versioning operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_bucket_versioning_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_put_data
    Description: Rate of PUT object data transfers per second
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_data
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: ontaps3_put_object_failed
    Description: Number of failed PUT object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_object_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_put_object_failed_client_close
    Description: Number of times PUT object operation failed due to the case where client closed the connection while the operation was still pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_object_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_put_object_latency
    Description: Average latency for PUT object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_object_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_put_object_rate
    Description: Number of PUT object operations per sec
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_object_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_put_object_tagging_failed
    Description: Number of failed PUT object tagging operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_object_tagging_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_put_object_tagging_failed_client_close
    Description: Number of times PUT object tagging operation failed because client terminated connection for operation pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_object_tagging_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_put_object_tagging_latency
    Description: Average latency for PUT object tagging operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_object_tagging_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average,no-zero-values

  - Harvest Metric: ontaps3_put_object_tagging_rate
    Description: Number of PUT object tagging operations per second.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_object_tagging_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate,no-zero-values

  - Harvest Metric: ontaps3_put_object_tagging_total
    Description: Number of PUT object tagging operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_object_tagging_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_put_object_total
    Description: Number of PUT object operations
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: put_object_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_request_parse_errors
    Description: Number of request parser errors due to malformed requests.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: request_parse_errors
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_requests
    Description: Total number of object store server requests
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: requests
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_requests_outstanding
    Description: Number of object store server requests in process
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: requests_outstanding
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: raw

  - Harvest Metric: ontaps3_root_user_access
    Description: Number of times access was done by root user.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: root_user_access
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_server_connection_close
    Description: Number of connection closes triggered by server due to fatal errors.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: server_connection_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_signature_v2_reqs
    Description: Total number of object store server signature V2 requests
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: signature_v2_reqs
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_signature_v4_reqs
    Description: Total number of object store server signature V4 requests
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: signature_v4_reqs
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_size
    REST:
      endpoint: api/protocols/s3/buckets
      metric: size
      template: conf/rest/9.7.0/ontap_s3.yaml

  - Harvest Metric: ontaps3_tagging
    Description: Number of requests with tagging specified.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: tagging
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta,no-zero-values

  - Harvest Metric: ontaps3_upload_part_failed
    Description: Number of failed Upload Part operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: upload_part_failed
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_upload_part_failed_client_close
    Description: Number of times Upload Part operation failed because client terminated connection for operation pending on server.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: upload_part_failed_client_close
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: ontaps3_upload_part_latency
    Description: Average latency for Upload Part operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: upload_part_latency
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: ontaps3_upload_part_rate
    Description: Number of Upload Part operations per second.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: upload_part_rate
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: ontaps3_upload_part_total
    Description: Number of Upload Part operations.
    ZAPI:
      endpoint: perf-object-get-instances object_store_server
      metric: upload_part_total
      template: conf/zapiperf/cdot/9.8.0/ontap_s3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: path_read_data
    Description: The average read throughput in kilobytes per second read from the indicated target port by the controller.
    ZAPI:
      endpoint: perf-object-get-instances path
      metric: read_data
      template: conf/zapiperf/cdot/9.8.0/path.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/path
      metric: read_data
      template: conf/restperf/9.12.0/path.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: path_read_iops
    Description: The number of I/O read operations sent from the initiator port to the indicated target port.
    ZAPI:
      endpoint: perf-object-get-instances path
      metric: read_iops
      template: conf/zapiperf/cdot/9.8.0/path.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/path
      metric: read_iops
      template: conf/restperf/9.12.0/path.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: path_read_latency
    Description: The average latency of I/O read operations sent from this controller to the indicated target port.
    ZAPI:
      endpoint: perf-object-get-instances path
      metric: read_latency
      template: conf/zapiperf/cdot/9.8.0/path.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/path
      metric: read_latency
      template: conf/restperf/9.12.0/path.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: path_total_data
    Description: The average throughput in kilobytes per second read and written from/to the indicated target port by the controller.
    ZAPI:
      endpoint: perf-object-get-instances path
      metric: total_data
      template: conf/zapiperf/cdot/9.8.0/path.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/path
      metric: total_data
      template: conf/restperf/9.12.0/path.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: path_total_iops
    Description: The number of total read/write I/O operations sent from the initiator port to the indicated target port.
    ZAPI:
      endpoint: perf-object-get-instances path
      metric: total_iops
      template: conf/zapiperf/cdot/9.8.0/path.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/path
      metric: total_iops
      template: conf/restperf/9.12.0/path.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: path_write_data
    Description: The average write throughput in kilobytes per second written to the indicated target port by the controller.
    ZAPI:
      endpoint: perf-object-get-instances path
      metric: write_data
      template: conf/zapiperf/cdot/9.8.0/path.yaml
      Unit: kb_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/path
      metric: write_data
      template: conf/restperf/9.12.0/path.yaml
      Unit: kb_per_sec
      Type: rate

  - Harvest Metric: path_write_iops
    Description: The number of I/O write operations sent from the initiator port to the indicated target port.
    ZAPI:
      endpoint: perf-object-get-instances path
      metric: write_iops
      template: conf/zapiperf/cdot/9.8.0/path.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/path
      metric: write_iops
      template: conf/restperf/9.12.0/path.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: path_write_latency
    Description: The average latency of I/O write operations sent from this controller to the indicated target port.
    ZAPI:
      endpoint: perf-object-get-instances path
      metric: write_latency
      template: conf/zapiperf/cdot/9.8.0/path.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/path
      metric: write_latency
      template: conf/restperf/9.12.0/path.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: qos_concurrency
    Description: This is the average number of concurrent requests for the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: concurrency
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: none
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: concurrency
      template: conf/restperf/9.12.0/workload.yaml
      Unit: none
      Type: rate

  - Harvest Metric: qos_detail_service_time
    Description: The workload's average service time per visit to the service center.
    ZAPI:
      endpoint: perf-object-get-instances workload_detail
      metric: service_time
      template: conf/zapiperf/cdot/9.8.0/workload_detail.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_detail
      metric: service_time
      template: conf/restperf/9.12.0/workload_detail.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: qos_detail_visits
    Description: The number of visits that the workload made to the service center; measured in visits per second.
    ZAPI:
      endpoint: perf-object-get-instances workload_detail
      metric: visits
      template: conf/zapiperf/cdot/9.8.0/workload_detail.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_detail
      metric: visits
      template: conf/restperf/9.12.0/workload_detail.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: qos_detail_volume_service_time
    Description: The workload's average service time per visit to the service center.
    ZAPI:
      endpoint: perf-object-get-instances workload_detail_volume
      metric: service_time
      template: conf/zapiperf/cdot/9.8.0/workload_detail_volume.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_detail_volume
      metric: service_time
      template: conf/restperf/9.12.0/workload_detail_volume.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: qos_detail_volume_visits
    Description: The number of visits that the workload made to the service center; measured in visits per second.
    ZAPI:
      endpoint: perf-object-get-instances workload_detail_volume
      metric: visits
      template: conf/zapiperf/cdot/9.8.0/workload_detail_volume.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_detail_volume
      metric: visits
      template: conf/restperf/9.12.0/workload_detail_volume.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: qos_detail_volume_wait_time
    Description: The workload's average wait time per visit to the service center.
    ZAPI:
      endpoint: perf-object-get-instances workload_detail_volume
      metric: wait_time
      template: conf/zapiperf/cdot/9.8.0/workload_detail_volume.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_detail_volume
      metric: wait_time
      template: conf/restperf/9.12.0/workload_detail_volume.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: qos_detail_wait_time
    Description: The workload's average wait time per visit to the service center.
    ZAPI:
      endpoint: perf-object-get-instances workload_detail
      metric: wait_time
      template: conf/zapiperf/cdot/9.8.0/workload_detail.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_detail
      metric: wait_time
      template: conf/restperf/9.12.0/workload_detail.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: qos_latency
    Description: This is the average response time for requests that were initiated by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: latency
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: latency
      template: conf/restperf/9.12.0/workload.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: qos_ops
    Description: Workload operations executed per second.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: ops
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: ops
      template: conf/restperf/9.12.0/workload.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: qos_read_data
    Description: This is the amount of data read per second from the filer by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: read_data
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: b_per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: read_data
      template: conf/restperf/9.12.0/workload.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: qos_read_io_type
    Description: This is the percentage of read requests served from various components (such as buffer cache, ext_cache, disk, etc.).
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: read_io_type
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: read_io_type_percent
      template: conf/restperf/9.12.0/workload.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: qos_read_latency
    Description: This is the average response time for read requests that were initiated by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: read_latency
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: read_latency
      template: conf/restperf/9.12.0/workload.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: qos_read_ops
    Description: This is the rate of this workload's read operations that completed during the measurement interval.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: read_ops
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: read_ops
      template: conf/restperf/9.12.0/workload.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: qos_sequential_reads
    Description: This is the percentage of reads, performed on behalf of the workload, that were sequential.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: sequential_reads
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: percent
      Type: percent,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: sequential_reads_percent
      template: conf/restperf/9.12.0/workload.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: qos_sequential_writes
    Description: This is the percentage of writes, performed on behalf of the workload, that were sequential. This counter is only available on platforms with more than 4GB of NVRAM.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: sequential_writes
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: percent
      Type: percent,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: sequential_writes_percent
      template: conf/restperf/9.12.0/workload.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: qos_total_data
    Description: This is the total amount of data read/written per second from/to the filer by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: total_data
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: b_per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: total_data
      template: conf/restperf/9.12.0/workload.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: qos_volume_latency
    Description: This is the average response time for requests that were initiated by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: latency
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: latency
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: qos_volume_ops
    Description: This field is the workload's rate of operations that completed during the measurement interval; measured per second.
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: ops
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: ops
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: qos_volume_read_data
    Description: This is the amount of data read per second from the filer by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: read_data
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: b_per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: read_data
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: qos_volume_read_io_type
    Description: This is the percentage of read requests served from various components (such as buffer cache, ext_cache, disk, etc.).
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: read_io_type
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: read_io_type_percent
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: qos_volume_read_latency
    Description: This is the average response time for read requests that were initiated by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: read_latency
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: read_latency
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: qos_volume_read_ops
    Description: This is the rate of this workload's read operations that completed during the measurement interval.
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: read_ops
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: read_ops
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: qos_volume_sequential_reads
    Description: This is the percentage of reads, performed on behalf of the workload, that were sequential.
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: sequential_reads
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: percent
      Type: percent,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: sequential_reads_percent
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: qos_volume_sequential_writes
    Description: This is the percentage of writes, performed on behalf of the workload, that were sequential. This counter is only available on platforms with more than 4GB of NVRAM.
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: sequential_writes
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: percent
      Type: percent,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: sequential_writes_percent
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: qos_volume_total_data
    Description: This is the total amount of data read/written per second from/to the filer by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: total_data
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: b_per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: total_data
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: qos_volume_write_data
    Description: This is the amount of data written per second to the filer by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: write_data
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: b_per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: write_data
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: qos_volume_write_latency
    Description: This is the average response time for write requests that were initiated by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: write_latency
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: write_latency
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: qos_volume_write_ops
    Description: This is the workload's write operations that completed during the measurement interval; measured per second.
    ZAPI:
      endpoint: perf-object-get-instances workload_volume
      metric: write_ops
      template: conf/zapiperf/cdot/9.8.0/workload_volume.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos_volume
      metric: write_ops
      template: conf/restperf/9.12.0/workload_volume.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: qos_write_data
    Description: This is the amount of data written per second to the filer by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: write_data
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: b_per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: write_data
      template: conf/restperf/9.12.0/workload.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: qos_write_latency
    Description: This is the average response time for write requests that were initiated by the workload.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: write_latency
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: write_latency
      template: conf/restperf/9.12.0/workload.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: qos_write_ops
    Description: This is the workload's write operations that completed during the measurement interval; measured per second.
    ZAPI:
      endpoint: perf-object-get-instances workload
      metric: write_ops
      template: conf/zapiperf/cdot/9.8.0/workload.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/qos
      metric: write_ops
      template: conf/restperf/9.12.0/workload.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: qtree_cifs_ops
    Description: Number of CIFS operations per second to the qtree
    ZAPI:
      endpoint: perf-object-get-instances qtree
      metric: cifs_ops
      template: conf/zapiperf/cdot/9.8.0/qtree.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/qtree
      metric: cifs_ops
      template: conf/restperf/9.12.0/qtree.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: qtree_id
    Description: The identifier for the qtree, unique within the qtree's volume.
    REST:
      endpoint: api/storage/qtrees
      metric: id
      template: conf/rest/9.12.0/qtree.yaml

  - Harvest Metric: qtree_internal_ops
    Description: Number of internal operations generated by activites such as snapmirror and backup per second to the qtree
    ZAPI:
      endpoint: perf-object-get-instances qtree
      metric: internal_ops
      template: conf/zapiperf/cdot/9.8.0/qtree.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/qtree
      metric: internal_ops
      template: conf/restperf/9.12.0/qtree.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: qtree_nfs_ops
    Description: Number of NFS operations per second to the qtree
    ZAPI:
      endpoint: perf-object-get-instances qtree
      metric: nfs_ops
      template: conf/zapiperf/cdot/9.8.0/qtree.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/qtree
      metric: nfs_ops
      template: conf/restperf/9.12.0/qtree.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: qtree_total_ops
    Description: Summation of NFS ops, CIFS ops, CSS ops and internal ops
    ZAPI:
      endpoint: perf-object-get-instances qtree
      metric: total_ops
      template: conf/zapiperf/cdot/9.8.0/qtree.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/qtree
      metric: total_ops
      template: conf/restperf/9.12.0/qtree.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: security_audit_destination_port
    ZAPI:
      endpoint: cluster-log-forward-get-iter
      metric: cluster-log-forward-info.port
      template: conf/zapi/cdot/9.8.0/security_audit_dest.yaml

  - Harvest Metric: security_certificate_expiry_time
    Description: Certificate expiration time. Can be provided on POST if creating self-signed certificate. The expiration time range is between 1 day to 10 years.
    ZAPI:
      endpoint: security-certificate-get-iter
      metric: certificate-info.expiration-date
      template: conf/zapi/cdot/9.8.0/security_certificate.yaml
    REST:
      endpoint: api/security/certificates
      metric: expiry_time
      template: conf/rest/9.12.0/security_certificate.yaml

  - Harvest Metric: security_ssh_max_instances
    REST:
      endpoint: api/security/ssh
      metric: max_instances
      template: conf/rest/9.12.0/security_ssh.yaml

  - Harvest Metric: shelf_disk_count
    ZAPI:
      endpoint: storage-shelf-info-get-iter
      metric: storage-shelf-info.disk-count
      template: conf/zapi/cdot/9.8.0/shelf.yaml
    REST:
      endpoint: api/storage/shelves
      metric: disk_count
      template: conf/rest/9.12.0/shelf.yaml

  - Harvest Metric: snapmirror_break_failed_count
    Description: The number of failed SnapMirror break operations for the relationship
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.break-failed-count
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: break_failed_count
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_break_successful_count
    Description: The number of successful SnapMirror break operations for the relationship
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.break-successful-count
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: break_successful_count
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_lag_time
    Description: Amount of time since the last snapmirror transfer in seconds
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.lag-time
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: lag_time
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_last_transfer_duration
    Description: Duration of the last SnapMirror transfer in seconds
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.last-transfer-duration
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: last_transfer_duration
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_last_transfer_end_timestamp
    Description: The Timestamp of the end of the last transfer
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.last-transfer-end-timestamp
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: last_transfer_end_timestamp
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_last_transfer_size
    Description: Size in kilobytes (1024 bytes) of the last transfer
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.last-transfer-size
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: last_transfer_size
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_newest_snapshot_timestamp
    Description: The timestamp of the newest Snapshot copy on the destination volume
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.newest-snapshot-timestamp
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: newest_snapshot_timestamp
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_resync_failed_count
    Description: The number of failed SnapMirror resync operations for the relationship
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.resync-failed-count
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: resync_failed_count
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_resync_successful_count
    Description: The number of successful SnapMirror resync operations for the relationship
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.resync-successful-count
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: resync_successful_count
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_total_transfer_bytes
    Description: Cumulative bytes transferred for the relationship
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.total-transfer-bytes
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: total_transfer_bytes
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_total_transfer_time_secs
    Description: Cumulative total transfer time in seconds for the relationship
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.total-transfer-time-secs
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: total_transfer_time_secs
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_update_failed_count
    Description: The number of successful SnapMirror update operations for the relationship
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.update-failed-count
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: update_failed_count
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapmirror_update_successful_count
    Description: Number of Successful Updates
    ZAPI:
      endpoint: snapmirror-get-iter
      metric: snapmirror-info.update-successful-count
      template: conf/zapi/cdot/9.8.0/snapmirror.yaml
    REST:
      endpoint: api/private/cli/snapmirror
      metric: update_successful_count
      template: conf/rest/9.12.0/snapmirror.yaml

  - Harvest Metric: snapshot_policy_total_schedules
    Description: Total Number of Schedules in this Policy
    ZAPI:
      endpoint: snapshot-policy-get-iter
      metric: snapshot-policy-info.total-schedules
      template: conf/zapi/cdot/9.8.0/snapshotPolicy.yaml
    REST:
      endpoint: api/private/cli/snapshot/policy
      metric: total_schedules
      template: conf/rest/9.12.0/snapshotPolicy.yaml

  - Harvest Metric: svm_cifs_connections
    Description: Number of connections
    ZAPI:
      endpoint: perf-object-get-instances cifs:vserver
      metric: connections
      template: conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs
      metric: connections
      template: conf/restperf/9.12.0/cifs_svm.yaml
      Unit: none
      Type: raw

  - Harvest Metric: svm_cifs_established_sessions
    Description: Number of established SMB and SMB2 sessions
    ZAPI:
      endpoint: perf-object-get-instances cifs:vserver
      metric: established_sessions
      template: conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs
      metric: established_sessions
      template: conf/restperf/9.12.0/cifs_svm.yaml
      Unit: none
      Type: raw

  - Harvest Metric: svm_cifs_latency
    Description: Average latency for CIFS operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:vserver
      metric: cifs_latency
      template: conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs
      metric: latency
      template: conf/restperf/9.12.0/cifs_svm.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_cifs_op_count
    Description: Array of select CIFS operation counts
    ZAPI:
      endpoint: perf-object-get-instances cifs:vserver
      metric: cifs_op_count
      template: conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs
      metric: op_count
      template: conf/restperf/9.12.0/cifs_svm.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_cifs_open_files
    Description: Number of open files over SMB and SMB2
    ZAPI:
      endpoint: perf-object-get-instances cifs:vserver
      metric: open_files
      template: conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs
      metric: open_files
      template: conf/restperf/9.12.0/cifs_svm.yaml
      Unit: none
      Type: raw

  - Harvest Metric: svm_cifs_ops
    Description: Total number of CIFS operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:vserver
      metric: cifs_ops
      template: conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs
      metric: total_ops
      template: conf/restperf/9.12.0/cifs_svm.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_cifs_read_latency
    Description: Average latency for CIFS read operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:vserver
      metric: cifs_read_latency
      template: conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs
      metric: average_read_latency
      template: conf/restperf/9.12.0/cifs_svm.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_cifs_read_ops
    Description: Total number of CIFS read operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:vserver
      metric: cifs_read_ops
      template: conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs
      metric: total_read_ops
      template: conf/restperf/9.12.0/cifs_svm.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_cifs_signed_sessions
    Description: Number of signed SMB and SMB2 sessions.
    ZAPI:
      endpoint: perf-object-get-instances cifs:vserver
      metric: signed_sessions
      template: conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml
      Unit: none
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs
      metric: signed_sessions
      template: conf/restperf/9.12.0/cifs_svm.yaml
      Unit: none
      Type: raw

  - Harvest Metric: svm_cifs_write_latency
    Description: Average latency for CIFS write operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:vserver
      metric: cifs_write_latency
      template: conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs
      metric: average_write_latency
      template: conf/restperf/9.12.0/cifs_svm.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_cifs_write_ops
    Description: Total number of CIFS write operations
    ZAPI:
      endpoint: perf-object-get-instances cifs:vserver
      metric: cifs_write_ops
      template: conf/zapiperf/cdot/9.8.0/cifs_vserver.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_cifs
      metric: total_write_ops
      template: conf/restperf/9.12.0/cifs_svm.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_nfs_access_avg_latency
    Description: Average latency of NFSv4.2 ACCESS operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: access_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: access.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_access_total
    Description: Total number of NFSv4.2 ACCESS operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: access_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: access.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_backchannel_ctl_avg_latency
    Description: Average latency of NFSv4.2 BACKCHANNEL_CTL operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: backchannel_ctl_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: backchannel_ctl.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_backchannel_ctl_total
    Description: Total number of NFSv4.2 BACKCHANNEL_CTL operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: backchannel_ctl_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: backchannel_ctl.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_bind_conn_to_session_avg_latency
    Description: Average latency of NFSv4.2 BIND_CONN_TO_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: bind_conn_to_session_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: bind_conn_to_session.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_bind_conn_to_session_total
    Description: Total number of NFSv4.2 BIND_CONN_TO_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: bind_conn_to_session_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: bind_conn_to_session.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: delta

  - Harvest Metric: svm_nfs_close_avg_latency
    Description: Average latency of NFSv4.2 CLOSE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: close_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: close.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_close_total
    Description: Total number of NFSv4.2 CLOSE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: close_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: close.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_commit_avg_latency
    Description: Average latency of NFSv4.2 COMMIT operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: commit_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: commit.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_commit_total
    Description: Total number of NFSv4.2 COMMIT operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: commit_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: commit.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_create_avg_latency
    Description: Average latency of NFSv4.2 CREATE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: create_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: create.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_create_session_avg_latency
    Description: Average latency of NFSv4.2 CREATE_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: create_session_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: create_session.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_create_session_total
    Description: Total number of NFSv4.2 CREATE_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: create_session_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: create_session.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_create_total
    Description: Total number of NFSv4.2 CREATE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: create_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: create.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_delegpurge_avg_latency
    Description: Average latency of NFSv4.2 DELEGPURGE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: delegpurge_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: delegpurge.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_delegpurge_total
    Description: Total number of NFSv4.2 DELEGPURGE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: delegpurge_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: delegpurge.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_delegreturn_avg_latency
    Description: Average latency of NFSv4.2 DELEGRETURN operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: delegreturn_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: delegreturn.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_delegreturn_total
    Description: Total number of NFSv4.2 DELEGRETURN operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: delegreturn_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: delegreturn.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_destroy_clientid_avg_latency
    Description: Average latency of NFSv4.2 DESTROY_CLIENTID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: destroy_clientid_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: destroy_clientid.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_destroy_clientid_total
    Description: Total number of NFSv4.2 DESTROY_CLIENTID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: destroy_clientid_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: destroy_clientid.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_destroy_session_avg_latency
    Description: Average latency of NFSv4.2 DESTROY_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: destroy_session_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: destroy_session.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_destroy_session_total
    Description: Total number of NFSv4.2 DESTROY_SESSION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: destroy_session_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: destroy_session.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_exchange_id_avg_latency
    Description: Average latency of NFSv4.2 EXCHANGE_ID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: exchange_id_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: exchange_id.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_exchange_id_total
    Description: Total number of NFSv4.2 EXCHANGE_ID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: exchange_id_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: exchange_id.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_free_stateid_avg_latency
    Description: Average latency of NFSv4.2 FREE_STATEID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: free_stateid_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: free_stateid.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_free_stateid_total
    Description: Total number of NFSv4.2 FREE_STATEID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: free_stateid_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: free_stateid.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_fsinfo_avg_latency
    Description: Average latency of FSInfo procedure requests. The counter keeps track of the average response time of FSInfo requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: fsinfo_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: fsinfo.average_latency
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_fsinfo_total
    Description: Total number FSInfo of procedure requests. It is the total number of FSInfo success and FSInfo error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: fsinfo_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: fsinfo.total
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_fsstat_avg_latency
    Description: Average latency of FSStat procedure requests. The counter keeps track of the average response time of FSStat requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: fsstat_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: fsstat.average_latency
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_fsstat_total
    Description: Total number FSStat of procedure requests. It is the total number of FSStat success and FSStat error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: fsstat_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: fsstat.total
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_get_dir_delegation_avg_latency
    Description: Average latency of NFSv4.2 GET_DIR_DELEGATION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: get_dir_delegation_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: get_dir_delegation.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_get_dir_delegation_total
    Description: Total number of NFSv4.2 GET_DIR_DELEGATION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: get_dir_delegation_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: get_dir_delegation.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_getattr_avg_latency
    Description: Average latency of NFSv4.2 GETATTR operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: getattr_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: getattr.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_getattr_total
    Description: Total number of NFSv4.2 GETATTR operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: getattr_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: getattr.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_getdeviceinfo_avg_latency
    Description: Average latency of NFSv4.2 GETDEVICEINFO operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: getdeviceinfo_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: getdeviceinfo.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_getdeviceinfo_total
    Description: Total number of NFSv4.2 GETDEVICEINFO operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: getdeviceinfo_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: getdeviceinfo.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_getdevicelist_avg_latency
    Description: Average latency of NFSv4.2 GETDEVICELIST operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: getdevicelist_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: getdevicelist.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_getdevicelist_total
    Description: Total number of NFSv4.2 GETDEVICELIST operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: getdevicelist_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: getdevicelist.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_getfh_avg_latency
    Description: Average latency of NFSv4.2 GETFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: getfh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: getfh.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_getfh_total
    Description: Total number of NFSv4.2 GETFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: getfh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: getfh.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_latency
    Description: Average latency of nfsv42 requests. This counter keeps track of the average response time of nfsv42 requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_layoutcommit_avg_latency
    Description: Average latency of NFSv4.2 LAYOUTCOMMIT operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: layoutcommit_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: layoutcommit.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_layoutcommit_total
    Description: Total number of NFSv4.2 LAYOUTCOMMIT operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: layoutcommit_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: layoutcommit.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_layoutget_avg_latency
    Description: Average latency of NFSv4.2 LAYOUTGET operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: layoutget_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: layoutget.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_layoutget_total
    Description: Total number of NFSv4.2 LAYOUTGET operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: layoutget_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: layoutget.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_layoutreturn_avg_latency
    Description: Average latency of NFSv4.2 LAYOUTRETURN operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: layoutreturn_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: layoutreturn.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_layoutreturn_total
    Description: Total number of NFSv4.2 LAYOUTRETURN operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: layoutreturn_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: layoutreturn.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_link_avg_latency
    Description: Average latency of NFSv4.2 LINK operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: link_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: link.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_link_total
    Description: Total number of NFSv4.2 LINK operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: link_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: link.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_lock_avg_latency
    Description: Average latency of NFSv4.2 LOCK operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: lock_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: lock.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_lock_total
    Description: Total number of NFSv4.2 LOCK operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: lock_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: lock.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_lockt_avg_latency
    Description: Average latency of NFSv4.2 LOCKT operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: lockt_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: lockt.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_lockt_total
    Description: Total number of NFSv4.2 LOCKT operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: lockt_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: lockt.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_locku_avg_latency
    Description: Average latency of NFSv4.2 LOCKU operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: locku_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: locku.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_locku_total
    Description: Total number of NFSv4.2 LOCKU operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: locku_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: locku.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_lookup_avg_latency
    Description: Average latency of NFSv4.2 LOOKUP operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: lookup_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: lookup.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_lookup_total
    Description: Total number of NFSv4.2 LOOKUP operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: lookup_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: lookup.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_lookupp_avg_latency
    Description: Average latency of NFSv4.2 LOOKUPP operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: lookupp_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: lookupp.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_lookupp_total
    Description: Total number of NFSv4.2 LOOKUPP operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: lookupp_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: lookupp.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_mkdir_avg_latency
    Description: Average latency of MkDir procedure requests. The counter keeps track of the average response time of MkDir requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: mkdir_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: mkdir.average_latency
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_mkdir_total
    Description: Total number MkDir of procedure requests. It is the total number of MkDir success and MkDir error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: mkdir_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: mkdir.total
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_mknod_avg_latency
    Description: Average latency of MkNod procedure requests. The counter keeps track of the average response time of MkNod requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: mknod_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: mknod.average_latency
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_mknod_total
    Description: Total number MkNod of procedure requests. It is the total number of MkNod success and MkNod error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: mknod_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: mknod.total
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_null_avg_latency
    Description: Average latency of NFSv4.2 NULL procedures.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: null_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: null.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_null_total
    Description: Total number of NFSv4.2 NULL procedures.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: null_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: null.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_nverify_avg_latency
    Description: Average latency of NFSv4.2 NVERIFY operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: nverify_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: nverify.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_nverify_total
    Description: Total number of NFSv4.2 NVERIFY operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: nverify_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: nverify.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_open_avg_latency
    Description: Average latency of NFSv4.2 OPEN operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: open_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: open.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_open_confirm_avg_latency
    Description: Average latency of OPEN_CONFIRM procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4
      metric: open_confirm_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4
      metric: open_confirm.average_latency
      template: conf/restperf/9.12.0/nfsv4.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_open_confirm_total
    Description: Total number of OPEN_CONFIRM procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4
      metric: open_confirm_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4
      metric: open_confirm.total
      template: conf/restperf/9.12.0/nfsv4.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_open_downgrade_avg_latency
    Description: Average latency of NFSv4.2 OPEN_DOWNGRADE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: open_downgrade_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: open_downgrade.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_open_downgrade_total
    Description: Total number of NFSv4.2 OPEN_DOWNGRADE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: open_downgrade_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: open_downgrade.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_open_total
    Description: Total number of NFSv4.2 OPEN operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: open_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: open.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_openattr_avg_latency
    Description: Average latency of NFSv4.2 OPENATTR operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: openattr_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: openattr.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_openattr_total
    Description: Total number of NFSv4.2 OPENATTR operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: openattr_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: openattr.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_ops
    Description: Total number of nfsv42 requests per sec.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: total_ops
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: total_ops
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_nfs_pathconf_avg_latency
    Description: Average latency of PathConf procedure requests. The counter keeps track of the average response time of PathConf requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: pathconf_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: pathconf.average_latency
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_pathconf_total
    Description: Total number PathConf of procedure requests. It is the total number of PathConf success and PathConf error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: pathconf_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: pathconf.total
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_putfh_avg_latency
    Description: Average latency of NFSv4.2 PUTFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: putfh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: putfh.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_putfh_total
    Description: Total number of NFSv4.2 PUTFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: putfh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: putfh.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_putpubfh_avg_latency
    Description: Average latency of NFSv4.2 PUTPUBFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: putpubfh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: putpubfh.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_putpubfh_total
    Description: Total number of NFSv4.2 PUTPUBFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: putpubfh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: putpubfh.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_putrootfh_avg_latency
    Description: Average latency of NFSv4.2 PUTROOTFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: putrootfh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: putrootfh.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_putrootfh_total
    Description: Total number of NFSv4.2 PUTROOTFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: putrootfh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: putrootfh.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_read_avg_latency
    Description: Average latency of NFSv4.2 READ operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: read_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: read.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_read_ops
    Description: Total observed NFSv3 read operations per second.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: nfsv3_read_ops
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: read_ops
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_nfs_read_symlink_avg_latency
    Description: Average latency of ReadSymLink procedure requests. The counter keeps track of the average response time of ReadSymLink requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: read_symlink_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: read_symlink.average_latency
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_read_symlink_total
    Description: Total number of ReadSymLink procedure requests. It is the total number of read symlink success and read symlink error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: read_symlink_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: read_symlink.total
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: none
      Type: delta

  - Harvest Metric: svm_nfs_read_throughput
    Description: NFSv4.2 read data transfers.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: nfs41_read_throughput
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: total.read_throughput
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_nfs_read_total
    Description: Total number of NFSv4.2 READ operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: read_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: read.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_readdir_avg_latency
    Description: Average latency of NFSv4.2 READDIR operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: readdir_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: readdir.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_readdir_total
    Description: Total number of NFSv4.2 READDIR operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: readdir_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: readdir.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_readdirplus_avg_latency
    Description: Average latency of ReadDirPlus procedure requests. The counter keeps track of the average response time of ReadDirPlus requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: readdirplus_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: readdirplus.average_latency
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_readdirplus_total
    Description: Total number ReadDirPlus of procedure requests. It is the total number of ReadDirPlus success and ReadDirPlus error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: readdirplus_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: readdirplus.total
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_readlink_avg_latency
    Description: Average latency of NFSv4.2 READLINK operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: readlink_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: readlink.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_readlink_total
    Description: Total number of NFSv4.2 READLINK operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: readlink_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: readlink.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_reclaim_complete_avg_latency
    Description: Average latency of NFSv4.2 RECLAIM_complete operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: reclaim_complete_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: reclaim_complete.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_reclaim_complete_total
    Description: Total number of NFSv4.2 RECLAIM_complete operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: reclaim_complete_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: reclaim_complete.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_release_lock_owner_avg_latency
    Description: Average Latency of RELEASE_LOCKOWNER procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4
      metric: release_lock_owner_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4
      metric: release_lock_owner.average_latency
      template: conf/restperf/9.12.0/nfsv4.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_release_lock_owner_total
    Description: Total number of RELEASE_LOCKOWNER procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4
      metric: release_lock_owner_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4
      metric: release_lock_owner.total
      template: conf/restperf/9.12.0/nfsv4.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_remove_avg_latency
    Description: Average latency of NFSv4.2 REMOVE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: remove_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: remove.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_remove_total
    Description: Total number of NFSv4.2 REMOVE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: remove_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: remove.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_rename_avg_latency
    Description: Average latency of NFSv4.2 RENAME operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: rename_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: rename.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_rename_total
    Description: Total number of NFSv4.2 RENAME operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: rename_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: rename.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_renew_avg_latency
    Description: Average latency of RENEW procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4
      metric: renew_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4
      metric: renew.average_latency
      template: conf/restperf/9.12.0/nfsv4.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_renew_total
    Description: Total number of RENEW procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4
      metric: renew_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4
      metric: renew.total
      template: conf/restperf/9.12.0/nfsv4.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_restorefh_avg_latency
    Description: Average latency of NFSv4.2 RESTOREFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: restorefh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: restorefh.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_restorefh_total
    Description: Total number of NFSv4.2 RESTOREFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: restorefh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: restorefh.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_rmdir_avg_latency
    Description: Average latency of RmDir procedure requests. The counter keeps track of the average response time of RmDir requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: rmdir_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: rmdir.average_latency
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_rmdir_total
    Description: Total number RmDir of procedure requests. It is the total number of RmDir success and RmDir error requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: rmdir_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: rmdir.total
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_savefh_avg_latency
    Description: Average latency of NFSv4.2 SAVEFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: savefh_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: savefh.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_savefh_total
    Description: Total number of NFSv4.2 SAVEFH operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: savefh_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: savefh.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_secinfo_avg_latency
    Description: Average latency of NFSv4.2 SECINFO operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: secinfo_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: secinfo.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_secinfo_no_name_avg_latency
    Description: Average latency of NFSv4.2 SECINFO_NO_NAME operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: secinfo_no_name_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: secinfo_no_name.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_secinfo_no_name_total
    Description: Total number of NFSv4.2 SECINFO_NO_NAME operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: secinfo_no_name_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: secinfo_no_name.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_secinfo_total
    Description: Total number of NFSv4.2 SECINFO operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: secinfo_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: secinfo.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_sequence_avg_latency
    Description: Average latency of NFSv4.2 SEQUENCE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: sequence_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: sequence.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_sequence_total
    Description: Total number of NFSv4.2 SEQUENCE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: sequence_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: sequence.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_set_ssv_avg_latency
    Description: Average latency of NFSv4.2 SET_SSV operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: set_ssv_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: set_ssv.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_set_ssv_total
    Description: Total number of NFSv4.2 SET_SSV operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: set_ssv_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: set_ssv.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_setattr_avg_latency
    Description: Average latency of NFSv4.2 SETATTR operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: setattr_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: setattr.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_setattr_total
    Description: Total number of NFSv4.2 SETATTR operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: setattr_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: setattr.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_setclientid_avg_latency
    Description: Average latency of SETCLIENTID procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4
      metric: setclientid_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4
      metric: setclientid.average_latency
      template: conf/restperf/9.12.0/nfsv4.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_setclientid_confirm_avg_latency
    Description: Average latency of SETCLIENTID_CONFIRM procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4
      metric: setclientid_confirm_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4
      metric: setclientid_confirm.average_latency
      template: conf/restperf/9.12.0/nfsv4.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_setclientid_confirm_total
    Description: Total number of SETCLIENTID_CONFIRM procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4
      metric: setclientid_confirm_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4
      metric: setclientid_confirm.total
      template: conf/restperf/9.12.0/nfsv4.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_setclientid_total
    Description: Total number of SETCLIENTID procedures
    ZAPI:
      endpoint: perf-object-get-instances nfsv4
      metric: setclientid_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v4
      metric: setclientid.total
      template: conf/restperf/9.12.0/nfsv4.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_symlink_avg_latency
    Description: Average latency of SymLink procedure requests. The counter keeps track of the average response time of SymLink requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: symlink_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: symlink.average_latency
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_symlink_total
    Description: Total number SymLink of procedure requests. It is the total number of SymLink success and create SymLink requests.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: symlink_total
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: symlink.total
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_test_stateid_avg_latency
    Description: Average latency of NFSv4.2 TEST_STATEID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: test_stateid_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: test_stateid.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_test_stateid_total
    Description: Total number of NFSv4.2 TEST_STATEID operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: test_stateid_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: test_stateid.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_throughput
    Description: NFSv4.2 write data transfers.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: nfs41_throughput
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: total.write_throughput
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_nfs_verify_avg_latency
    Description: Average latency of NFSv4.2 VERIFY operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: verify_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: verify.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_verify_total
    Description: Total number of NFSv4.2 VERIFY operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: verify_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: verify.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_want_delegation_avg_latency
    Description: Average latency of NFSv4.2 WANT_DELEGATION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: want_delegation_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: want_delegation.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_want_delegation_total
    Description: Total number of NFSv4.2 WANT_DELEGATION operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: want_delegation_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: want_delegation.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_nfs_write_avg_latency
    Description: Average latency of NFSv4.2 WRITE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: write_avg_latency
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: microsec
      Type: average,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: write.average_latency
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_nfs_write_ops
    Description: Total observed NFSv3 write operations per second.
    ZAPI:
      endpoint: perf-object-get-instances nfsv3
      metric: nfsv3_write_ops
      template: conf/zapiperf/cdot/9.8.0/nfsv3.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v3
      metric: write_ops
      template: conf/restperf/9.12.0/nfsv3.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_nfs_write_throughput
    Description: NFSv4.2 data transfers.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: nfs41_write_throughput
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: per_sec
      Type: rate,no-zero-values
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: total.throughput
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_nfs_write_total
    Description: Total number of NFSv4.2 WRITE operations.
    ZAPI:
      endpoint: perf-object-get-instances nfsv4_1
      metric: write_total
      template: conf/zapiperf/cdot/9.8.0/nfsv4_1.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/svm_nfs_v42
      metric: write.total
      template: conf/restperf/9.12.0/nfsv4_2.yaml
      Unit: none
      Type: rate

  - Harvest Metric: svm_vol_avg_latency
    Description: Average latency in microseconds for the WAFL filesystem to process all the operations on the volume; not including request processing or network communication time
    ZAPI:
      endpoint: perf-object-get-instances volume:vserver
      metric: avg_latency
      template: conf/zapiperf/cdot/9.8.0/volume_svm.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:svm
      metric: average_latency
      template: conf/restperf/9.12.0/volume_svm.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_vol_other_latency
    Description: Average latency in microseconds for the WAFL filesystem to process other operations to the volume; not including request processing or network communication time
    ZAPI:
      endpoint: perf-object-get-instances volume:vserver
      metric: other_latency
      template: conf/zapiperf/cdot/9.8.0/volume_svm.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:svm
      metric: other_latency
      template: conf/restperf/9.12.0/volume_svm.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_vol_other_ops
    Description: Number of other operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:vserver
      metric: other_ops
      template: conf/zapiperf/cdot/9.8.0/volume_svm.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:svm
      metric: total_other_ops
      template: conf/restperf/9.12.0/volume_svm.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_vol_read_data
    Description: Bytes read per second
    ZAPI:
      endpoint: perf-object-get-instances volume:vserver
      metric: read_data
      template: conf/zapiperf/cdot/9.8.0/volume_svm.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:svm
      metric: bytes_read
      template: conf/restperf/9.12.0/volume_svm.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: svm_vol_read_latency
    Description: Average latency in microseconds for the WAFL filesystem to process read request to the volume; not including request processing or network communication time
    ZAPI:
      endpoint: perf-object-get-instances volume:vserver
      metric: read_latency
      template: conf/zapiperf/cdot/9.8.0/volume_svm.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:svm
      metric: read_latency
      template: conf/restperf/9.12.0/volume_svm.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_vol_read_ops
    Description: Number of read operations per second from the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:vserver
      metric: read_ops
      template: conf/zapiperf/cdot/9.8.0/volume_svm.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:svm
      metric: total_read_ops
      template: conf/restperf/9.12.0/volume_svm.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_vol_total_ops
    Description: Number of operations per second serviced by the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:vserver
      metric: total_ops
      template: conf/zapiperf/cdot/9.8.0/volume_svm.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:svm
      metric: total_ops
      template: conf/restperf/9.12.0/volume_svm.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_vol_write_data
    Description: Bytes written per second
    ZAPI:
      endpoint: perf-object-get-instances volume:vserver
      metric: write_data
      template: conf/zapiperf/cdot/9.8.0/volume_svm.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:svm
      metric: bytes_written
      template: conf/restperf/9.12.0/volume_svm.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: svm_vol_write_latency
    Description: Average latency in microseconds for the WAFL filesystem to process write request to the volume; not including request processing or network communication time
    ZAPI:
      endpoint: perf-object-get-instances volume:vserver
      metric: write_latency
      template: conf/zapiperf/cdot/9.8.0/volume_svm.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume:svm
      metric: write_latency
      template: conf/restperf/9.12.0/volume_svm.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_vol_write_ops
    Description: Number of write operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume:vserver
      metric: write_ops
      template: conf/zapiperf/cdot/9.8.0/volume_svm.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume:svm
      metric: total_write_ops
      template: conf/restperf/9.12.0/volume_svm.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_vscan_connections_active
    Description: Total number of current active connections
    ZAPI:
      endpoint: perf-object-get-instances offbox_vscan
      metric: connections_active
      template: conf/zapiperf/cdot/9.8.0/vscan_svm.yaml
      Unit: none
      Type: raw

  - Harvest Metric: svm_vscan_dispatch_latency
    Description: Average dispatch latency
    ZAPI:
      endpoint: perf-object-get-instances offbox_vscan
      metric: dispatch_latency
      template: conf/zapiperf/cdot/9.8.0/vscan_svm.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_vscan_scan_latency
    Description: Average scan latency
    ZAPI:
      endpoint: perf-object-get-instances offbox_vscan
      metric: scan_latency
      template: conf/zapiperf/cdot/9.8.0/vscan_svm.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: svm_vscan_scan_noti_received_rate
    Description: Total number of scan notifications received by the dispatcher per second
    ZAPI:
      endpoint: perf-object-get-instances offbox_vscan
      metric: scan_noti_received_rate
      template: conf/zapiperf/cdot/9.8.0/vscan_svm.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: svm_vscan_scan_request_dispatched_rate
    Description: Total number of scan requests sent to the Vscanner per second
    ZAPI:
      endpoint: perf-object-get-instances offbox_vscan
      metric: scan_request_dispatched_rate
      template: conf/zapiperf/cdot/9.8.0/vscan_svm.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: token_copy_bytes
    Description: Total number of bytes copied.
    ZAPI:
      endpoint: perf-object-get-instances token_manager
      metric: token_copy_bytes
      template: conf/zapiperf/cdot/9.8.0/token_manager.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/token_manager
      metric: token_copy.bytes
      template: conf/restperf/9.12.0/token_manager.yaml
      Unit: none
      Type: rate

  - Harvest Metric: token_copy_failure
    Description: Number of failed token copy requests.
    ZAPI:
      endpoint: perf-object-get-instances token_manager
      metric: token_copy_failure
      template: conf/zapiperf/cdot/9.8.0/token_manager.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/token_manager
      metric: token_copy.failures
      template: conf/restperf/9.12.0/token_manager.yaml
      Unit: none
      Type: delta

  - Harvest Metric: token_copy_success
    Description: Number of successful token copy requests.
    ZAPI:
      endpoint: perf-object-get-instances token_manager
      metric: token_copy_success
      template: conf/zapiperf/cdot/9.8.0/token_manager.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/token_manager
      metric: token_copy.successes
      template: conf/restperf/9.12.0/token_manager.yaml
      Unit: none
      Type: delta

  - Harvest Metric: token_create_bytes
    Description: Total number of bytes for which tokens are created.
    ZAPI:
      endpoint: perf-object-get-instances token_manager
      metric: token_create_bytes
      template: conf/zapiperf/cdot/9.8.0/token_manager.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/token_manager
      metric: token_create.bytes
      template: conf/restperf/9.12.0/token_manager.yaml
      Unit: none
      Type: rate

  - Harvest Metric: token_create_failure
    Description: Number of failed token create requests.
    ZAPI:
      endpoint: perf-object-get-instances token_manager
      metric: token_create_failure
      template: conf/zapiperf/cdot/9.8.0/token_manager.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/token_manager
      metric: token_create.failures
      template: conf/restperf/9.12.0/token_manager.yaml
      Unit: none
      Type: delta

  - Harvest Metric: token_create_success
    Description: Number of successful token create requests.
    ZAPI:
      endpoint: perf-object-get-instances token_manager
      metric: token_create_success
      template: conf/zapiperf/cdot/9.8.0/token_manager.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/token_manager
      metric: token_create.successes
      template: conf/restperf/9.12.0/token_manager.yaml
      Unit: none
      Type: delta

  - Harvest Metric: token_zero_bytes
    Description: Total number of bytes zeroed.
    ZAPI:
      endpoint: perf-object-get-instances token_manager
      metric: token_zero_bytes
      template: conf/zapiperf/cdot/9.8.0/token_manager.yaml
      Unit: none
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/token_manager
      metric: token_zero.bytes
      template: conf/restperf/9.12.0/token_manager.yaml
      Unit: none
      Type: rate

  - Harvest Metric: token_zero_failure
    Description: Number of failed token zero requests.
    ZAPI:
      endpoint: perf-object-get-instances token_manager
      metric: token_zero_failure
      template: conf/zapiperf/cdot/9.8.0/token_manager.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/token_manager
      metric: token_zero.failures
      template: conf/restperf/9.12.0/token_manager.yaml
      Unit: none
      Type: delta

  - Harvest Metric: token_zero_success
    Description: Number of successful token zero requests.
    ZAPI:
      endpoint: perf-object-get-instances token_manager
      metric: token_zero_success
      template: conf/zapiperf/cdot/9.8.0/token_manager.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/token_manager
      metric: token_zero.successes
      template: conf/restperf/9.12.0/token_manager.yaml
      Unit: none
      Type: delta

  - Harvest Metric: volume_autosize_grow_threshold_percent
    Description: Used space threshold size, in percentage, for the automatic growth of the volume. When the amount of used space in the volume becomes greater than this threhold, the volume automatically grows unless it has reached the maximum size. The volume grows when 'space.used' is greater than this percent of 'space.size'. The 'grow_threshold' size cannot be less than or equal to the 'shrink_threshold' size..
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-autosize-attributes.grow-threshold-percent
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: autosize.grow_threshold
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_autosize_maximum_size
    Description: Maximum size in bytes up to which a volume grows automatically. This size cannot be less than the current volume size, or less than or equal to the minimum size of volume.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-autosize-attributes.maximum-size
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: autosize.maximum
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_avg_latency
    Description: Average latency in microseconds for the WAFL filesystem to process all the operations on the volume; not including request processing or network communication time
    ZAPI:
      endpoint: perf-object-get-instances volume
      metric: avg_latency
      template: conf/zapiperf/cdot/9.8.0/volume.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume
      metric: average_latency
      template: conf/restperf/9.12.0/volume.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: volume_filesystem_size
    Description: Total usable size of the volume, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.filesystem-size
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.filesystem_size
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_inode_files_total
    Description: Total user-visible file (inode) count, i.e., current maximum number of user-visible files (inodes) that this volume can currently hold.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-inode-attributes.files-total
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/private/cli/volume
      metric: files
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_inode_files_used
    Description: Number of user-visible files (inodes) used. This field is valid only when the volume is online.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-inode-attributes.files-used
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/private/cli/volume
      metric: files_used
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_other_latency
    Description: Average latency in microseconds for the WAFL filesystem to process other operations to the volume; not including request processing or network communication time
    ZAPI:
      endpoint: perf-object-get-instances volume
      metric: other_latency
      template: conf/zapiperf/cdot/9.8.0/volume.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume
      metric: other_latency
      template: conf/restperf/9.12.0/volume.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: volume_other_ops
    Description: Number of other operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume
      metric: other_ops
      template: conf/zapiperf/cdot/9.8.0/volume.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume
      metric: total_other_ops
      template: conf/restperf/9.12.0/volume.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: volume_read_data
    Description: Bytes read per second
    ZAPI:
      endpoint: perf-object-get-instances volume
      metric: read_data
      template: conf/zapiperf/cdot/9.8.0/volume.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume
      metric: bytes_read
      template: conf/restperf/9.12.0/volume.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: volume_read_latency
    Description: Average latency in microseconds for the WAFL filesystem to process read request to the volume; not including request processing or network communication time
    ZAPI:
      endpoint: perf-object-get-instances volume
      metric: read_latency
      template: conf/zapiperf/cdot/9.8.0/volume.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume
      metric: read_latency
      template: conf/restperf/9.12.0/volume.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: volume_read_ops
    Description: Number of read operations per second from the volume
    ZAPI:
      endpoint: perf-object-get-instances volume
      metric: read_ops
      template: conf/zapiperf/cdot/9.8.0/volume.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume
      metric: total_read_ops
      template: conf/restperf/9.12.0/volume.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: volume_sis_compress_saved
    Description: The total disk space (in bytes) that is saved by compressing blocks on the referenced file system.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-sis-attributes.compression-space-saved
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/private/cli/volume
      metric: compression_space_saved
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_sis_compress_saved_percent
    Description: Percentage of the total disk space that is saved by compressing blocks on the referenced file system
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-sis-attributes.percentage-compression-space-saved
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/private/cli/volume
      metric: compression_space_saved_percent
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_sis_dedup_saved
    Description: The total disk space (in bytes) that is saved by deduplication and file cloning.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-sis-attributes.deduplication-space-saved
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/private/cli/volume
      metric: dedupe_space_saved
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_sis_dedup_saved_percent
    Description: Percentage of the total disk space that is saved by deduplication and file cloning.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-sis-attributes.percentage-deduplication-space-saved
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/private/cli/volume
      metric: dedupe_space_saved_percent
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_sis_total_saved
    Description: Total space saved (in bytes) in the volume due to deduplication, compression, and file cloning.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-sis-attributes.total-space-saved
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/private/cli/volume
      metric: sis_space_saved
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_sis_total_saved_percent
    Description: Percentage of total disk space that is saved by compressing blocks, deduplication and file cloning.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-sis-attributes.percentage-total-space-saved
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/private/cli/volume
      metric: sis_space_saved_percent
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_size
    Description: Total provisioned size. The default size is equal to the minimum size of 20MB, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.size
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.size
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_size_available
    Description: The available space, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.size-available
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.available
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_size_total
    Description: Total size of AFS, excluding snap-reserve, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.size-total
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.afs_total
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_size_used
    Description: The virtual space used (includes volume reserves) before storage efficiency, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.size-used
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.used
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_size_used_percent
    Description: Percentage of the volume size that is used.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.percentage-size-used
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.percent_used
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_snapshot_count
    Description: Number of Snapshot copies in the volume.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-snapshot-attributes.snapshot-count
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: snapshot_count
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_snapshot_reserve_available
    Description: Size available for Snapshot copies within the Snapshot copy reserve, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.snapshot-reserve-available
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.snapshot.reserve_available
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_snapshot_reserve_percent
    Description: The space that has been set aside as a reserve for Snapshot copy usage, in percent.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.percentage-snapshot-reserve
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.snapshot.reserve_percent
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_snapshot_reserve_size
    Description: Size in the volume that has been set aside as a reserve for Snapshot copy usage, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.snapshot-reserve-size
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.snapshot.reserve_size
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_snapshot_reserve_used_percent
    Description: Percentage of snapshot reserve size that has been used.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.percentage-snapshot-reserve-used
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.snapshot.space_used_percent
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_snapshots_size_available
    Description: Available space for Snapshot copies from snap-reserve, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.size-available-for-snapshots
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.size_available_for_snapshots
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_snapshots_size_used
    Description: The total space used by Snapshot copies in the volume, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.size-used-by-snapshots
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.snapshot.used
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_space_expected_available
    Description: Size that should be available for the volume, irrespective of available size in the aggregate, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.expected-available
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.expected_available
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_space_logical_available
    Description: The amount of space available in this volume with storage efficiency space considered used, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.logical-available
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.logical_space.available
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_space_logical_used
    Description: SUM of (physical-used, shared_refs, compression_saved_in_plane0, vbn_zero, future_blk_cnt), in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.logical-used
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.logical_space.used
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_space_logical_used_by_afs
    Description: The virtual space used by AFS alone (includes volume reserves) and along with storage efficiency, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.logical-used-by-afs
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.logical_space.used_by_afs
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_space_logical_used_by_snapshots
    Description: Size that is logically used across all Snapshot copies in the volume, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.logical-used-by-snapshots
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.logical_space.used_by_snapshots
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_space_logical_used_percent
    Description: SUM of (physical-used, shared_refs, compression_saved_in_plane0, vbn_zero, future_blk_cnt), as a percentage.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.logical-used-percent
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.logical_space.used_percent
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_space_physical_used
    Description: Size that is physically used in the volume, in bytes.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.physical-used
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.physical_used
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_space_physical_used_percent
    Description: Size that is physically used in the volume, as a percentage.
    ZAPI:
      endpoint: volume-get-iter
      metric: volume-attributes.volume-space-attributes.physical-used-percent
      template: conf/zapi/cdot/9.8.0/volume.yaml
    REST:
      endpoint: api/storage/volumes
      metric: space.physical_used_percent
      template: conf/rest/9.12.0/volume.yaml

  - Harvest Metric: volume_total_ops
    Description: Number of operations per second serviced by the volume
    ZAPI:
      endpoint: perf-object-get-instances volume
      metric: total_ops
      template: conf/zapiperf/cdot/9.8.0/volume.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume
      metric: total_ops
      template: conf/restperf/9.12.0/volume.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: volume_write_data
    Description: Bytes written per second
    ZAPI:
      endpoint: perf-object-get-instances volume
      metric: write_data
      template: conf/zapiperf/cdot/9.8.0/volume.yaml
      Unit: b_per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume
      metric: bytes_written
      template: conf/restperf/9.12.0/volume.yaml
      Unit: b_per_sec
      Type: rate

  - Harvest Metric: volume_write_latency
    Description: Average latency in microseconds for the WAFL filesystem to process write request to the volume; not including request processing or network communication time
    ZAPI:
      endpoint: perf-object-get-instances volume
      metric: write_latency
      template: conf/zapiperf/cdot/9.8.0/volume.yaml
      Unit: microsec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/volume
      metric: write_latency
      template: conf/restperf/9.12.0/volume.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: volume_write_ops
    Description: Number of write operations per second to the volume
    ZAPI:
      endpoint: perf-object-get-instances volume
      metric: write_ops
      template: conf/zapiperf/cdot/9.8.0/volume.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/volume
      metric: total_write_ops
      template: conf/restperf/9.12.0/volume.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: vscan_scan_latency
    Description: Average scan latency
    ZAPI:
      endpoint: perf-object-get-instances offbox_vscan_server
      metric: scan_latency
      template: conf/zapiperf/cdot/9.8.0/vscan.yaml
      Unit: microsec
      Type: average

  - Harvest Metric: vscan_scan_request_dispatched_rate
    Description: Total number of scan requests sent to the Vscanner per second
    ZAPI:
      endpoint: perf-object-get-instances offbox_vscan_server
      metric: scan_request_dispatched_rate
      template: conf/zapiperf/cdot/9.8.0/vscan.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: vscan_scanner_stats_pct_cpu_used
    Description: Percentage CPU utilization on scanner
    ZAPI:
      endpoint: perf-object-get-instances offbox_vscan_server
      metric: scanner_stats_pct_cpu_used
      template: conf/zapiperf/cdot/9.8.0/vscan.yaml
      Unit: none
      Type: raw

  - Harvest Metric: vscan_scanner_stats_pct_mem_used
    Description: Percentage RAM utilization on scanner
    ZAPI:
      endpoint: perf-object-get-instances offbox_vscan_server
      metric: scanner_stats_pct_mem_used
      template: conf/zapiperf/cdot/9.8.0/vscan.yaml
      Unit: none
      Type: raw

  - Harvest Metric: vscan_scanner_stats_pct_network_used
    Description: Percentage network utilization on scanner
    ZAPI:
      endpoint: perf-object-get-instances offbox_vscan_server
      metric: scanner_stats_pct_network_used
      template: conf/zapiperf/cdot/9.8.0/vscan.yaml
      Unit: none
      Type: raw

  - Harvest Metric: wafl_avg_msg_latency
    Description: Average turnaround time for WAFL messages in milliseconds.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: avg_wafl_msg_latency
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: millisec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: average_msg_latency
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: millisec
      Type: average

  - Harvest Metric: wafl_avg_non_wafl_msg_latency
    Description: Average turnaround time for non-WAFL messages in milliseconds.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: avg_non_wafl_msg_latency
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: millisec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: average_non_wafl_msg_latency
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: millisec
      Type: average

  - Harvest Metric: wafl_avg_repl_msg_latency
    Description: Average turnaround time for replication WAFL messages in milliseconds.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: avg_wafl_repl_msg_latency
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: millisec
      Type: average
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: average_replication_msg_latency
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: millisec
      Type: average

  - Harvest Metric: wafl_cp_count
    Description: Array of counts of different types of Consistency Points (CP).
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: cp_count
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: cp_count
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: none
      Type: delta

  - Harvest Metric: wafl_cp_phase_times
    Description: Array of percentage time spent in different phases of Consistency Point (CP).
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: cp_phase_times
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: cp_phase_times
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: wafl_memory_free
    Description: The current WAFL memory available in the system.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_memory_free
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: mb
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: memory_free
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: mb
      Type: raw

  - Harvest Metric: wafl_memory_used
    Description: The current WAFL memory used in the system.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_memory_used
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: mb
      Type: raw
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: memory_used
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: mb
      Type: raw

  - Harvest Metric: wafl_msg_total
    Description: Total number of WAFL messages per second.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_msg_total
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: msg_total
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: wafl_non_wafl_msg_total
    Description: Total number of non-WAFL messages per second.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: non_wafl_msg_total
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: non_wafl_msg_total
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: wafl_read_io_type
    Description: Percentage of reads served from buffer cache, external cache, or disk.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: read_io_type
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: read_io_type
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: percent
      Type: percent

  - Harvest Metric: wafl_reads_from_cache
    Description: WAFL reads from cache.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_reads_from_cache
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: reads_from_cache
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: none
      Type: delta

  - Harvest Metric: wafl_reads_from_cloud
    Description: WAFL reads from cloud storage.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_reads_from_cloud
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: reads_from_cloud
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: none
      Type: delta

  - Harvest Metric: wafl_reads_from_cloud_s2c_bin
    Description: WAFL reads from cloud storage via s2c bin.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_reads_from_cloud_s2c_bin
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: reads_from_cloud_s2c_bin
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: none
      Type: delta

  - Harvest Metric: wafl_reads_from_disk
    Description: WAFL reads from disk.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_reads_from_disk
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: reads_from_disk
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: none
      Type: delta

  - Harvest Metric: wafl_reads_from_ext_cache
    Description: WAFL reads from external cache.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_reads_from_ext_cache
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: reads_from_external_cache
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: none
      Type: delta

  - Harvest Metric: wafl_reads_from_fc_miss
    Description: WAFL reads from remote volume for fc_miss.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_reads_from_fc_miss
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: reads_from_fc_miss
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: none
      Type: delta

  - Harvest Metric: wafl_reads_from_pmem
    Description: Wafl reads from persistent mmeory.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_reads_from_pmem
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: none
      Type: delta

  - Harvest Metric: wafl_reads_from_ssd
    Description: WAFL reads from SSD.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_reads_from_ssd
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: none
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: reads_from_ssd
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: none
      Type: delta

  - Harvest Metric: wafl_repl_msg_total
    Description: Total number of replication WAFL messages per second.
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: wafl_repl_msg_total
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: per_sec
      Type: rate
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: replication_msg_total
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: per_sec
      Type: rate

  - Harvest Metric: wafl_total_cp_msecs
    Description: Milliseconds spent in Consistency Point (CP).
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: total_cp_msecs
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: millisec
      Type: delta
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: total_cp_msecs
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: millisec
      Type: delta

  - Harvest Metric: wafl_total_cp_util
    Description: Percentage of time spent in a Consistency Point (CP).
    ZAPI:
      endpoint: perf-object-get-instances wafl
      metric: total_cp_util
      template: conf/zapiperf/cdot/9.8.0/wafl.yaml
      Unit: percent
      Type: percent
    REST:
      endpoint: api/cluster/counter/tables/wafl
      metric: total_cp_util
      template: conf/restperf/9.12.0/wafl.yaml
      Unit: percent
      Type: percent


```