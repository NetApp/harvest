This document describes each E-Series major event log (MEL) event that Harvest collects and remediation steps.

### Synth Drv PFA

**Impact**: Protection

**MEL Event**: `MEL_EV_SYNTH_DRIVE_PFA` (0x101E)

Impending drive failure detected by controller

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### PI Drive Locked Out

**Impact**: Protection

**MEL Event**: `MEL_EV_PI_DRIVE_LOCKED_OUT` (0x1020)

Data assurance drive has been locked out

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Excessive Reboots Detected

**Impact**: Availability

**MEL Event**: `MEL_EV_EXCESSIVE_REBOOTS_DETECTED` (0x1403)

Excessive reboots (exceptions) have occurred on the controller

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Channel Failover

**Impact**: Availability

**MEL Event**: `MEL_EV_DFC_CHANNEL_FAILOVER` (0x1513)

Individual drive - Degraded path

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Controller Wide Port Has Gone To Failed State

**Impact**: Availability

**MEL Event**: `MEL_EV_DSAS_WPORT_DEG_TO_FAIL_CTLR` (0x1710)

Controller wide port has gone to failed state

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Missing Drive Lockdown

**Impact**: Availability

**MEL Event**: `MEL_EV_MISSING_DRIVE_LOCKDOWN` (0x1907)

Controller is locked down due to too many missing drives

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Cache Battery Failure

**Impact**: Protection

**MEL Event**: `MEL_EV_CACHE_BATTERY_FAILURE` (0x210C)

Controller cache battery failed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### OCB Setting Conflict

**Impact**: Configuration

**MEL Event**: `MEL_EV_OCB_SETTING_CONFLICT` (0x211B)

Batteries present but NVSRAM file configured for no batteries

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Write-back Caching Forcibly Disabled

**Impact**: Protection

**MEL Event**: `MEL_EV_WB_CACHING_FORCIBLY_DISABLED` (0x212B)

Write-back caching forcibly disabled

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Dirty Cache Not Flushed On The Only Active Controller

**Impact**: Protection

**MEL Event**: `MEL_EV_CACHE_NOT_FLUSHED_ON_ONLY_CTLR` (0x2131)

Dirty cache not flushed on the only active controller

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Wrong Sector Size

**Impact**: Configuration

**MEL Event**: `MEL_EV_WRONG_SECTOR_SIZE` (0x224A)

Drive has wrong block size

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Drive No Response

**Impact**: Availability

**MEL Event**: `MEL_EV_DRV_NO_RESPONSE` (0x224D)

Drive failed - no response at start of day

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### LUN Down

**Impact**: Availability

**MEL Event**: `MEL_EV_LUN_DOWN` (0x2250)

Volume failure

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Uncertified Drive

**Impact**: Configuration

**MEL Event**: `MEL_EV_UNCERTIFIED_DRIVE` (0x2260)

Uncertified Drive Detected

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Configuration Wrong Drive Type

**Impact**: Configuration

**MEL Event**: `MEL_EV_CFG_WRONG_DRIVE_TYPE` (0x2262)

Failed drive replaced with wrong drive type

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Reconfiguration Failed

**Impact**: Protection

**MEL Event**: `MEL_EV_RECONFIGURATION_FAILED` (0x2266)

Volume modification operation failed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Incompatible Drive Due To Invalid Configuration

**Impact**: Configuration

**MEL Event**: `MEL_EV_INCOMPAT_DRIVE_INVALID_CONFIG` (0x2267)

Incompatible drive due to invalid configuration on drive

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Drive Failure

**Impact**: Availability

**MEL Event**: `MEL_EV_CFG_DRIVE_FAILURE` (0x226C)

Drive failure

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Drive In Volume Group Or Hot Spare In Use Removed

**Impact**: Availability

**MEL Event**: `MEL_EV_DRIVE_IN_VG_OR_HOT_SPARE_REMOVED` (0x226D)

Assigned drive or hot spare-in use drive removed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Physical Drive Has Unsupported Capacity

**Impact**: Configuration

**MEL Event**: `MEL_EV_DRIVE_UNSUPPORTED_CAPACITY` (0x2271)

Physical drive has unsupported capacity

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Volume Group Missing

**Impact**: Availability

**MEL Event**: `MEL_EV_VOLUME_GROUP_MISSING` (0x2274)

Component is missing

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Volume Group Incomplete

**Impact**: Availability

**MEL Event**: `MEL_EV_VOLUME_GROUP_INCOMPLETE` (0x2275)

Component incomplete

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Incompatible Alignment for Emulation Drive

**Impact**: Configuration

**MEL Event**: `MEL_EV_INCOMPATIBLE_ALIGNMENT_FOR_EMULATION_DRIVE` (0x2278)

Incompatible alignment for emulation drive

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Copy Then Fail No Spare

**Impact**: Protection

**MEL Event**: `MEL_EV_COPY_THEN_FAIL_NO_SPARE` (0x227C)

Waiting for eligible copy destination to start drive copy

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Drive PFA 2

**Impact**: Protection

**MEL Event**: `MEL_EV_DRIVE_PFA2` (0x2285)

Impending drive failure detected by drive

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Drive Fast Format Failure

**Impact**: Configuration

**MEL Event**: `MEL_EV_FAST_FORMAT_FAILED` (0x228C)

Drive fast format failed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Drive Flash Technology Unsupported

**Impact**: Configuration

**MEL Event**: `MEL_EV_DRIVE_FLASH_TECHNOLOGY_UNSUPPORTED` (0x2295)

The drive's flash technology is not supported on this system. For example an NVMe QLC drive in a system that doesn't support QLC drives, or a non-QLC drive in a system that requires QLC drives

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Controller Removed

**Impact**: Availability

**MEL Event**: `MEL_EV_CONTROLLER` (0x2500)

Controller removed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Line Missing

**Impact**: Availability

**MEL Event**: `MEL_EV_LINE_MISSING` (0x280A)

Controller tray component removed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Line Failed

**Impact**: Availability

**MEL Event**: `MEL_EV_LINE_FAILED` (0x280B)

Controller tray component failed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Enclosure Fail

**Impact**: Availability

**MEL Event**: `MEL_EV_ENCL_FAIL` (0x280D)

Drive tray component failed or removed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Temperature Sensor Warning

**Impact**: Availability

**MEL Event**: `MEL_EV_TEMP_SENSOR_WARNING` (0x281B)

Nominal temperature exceeded

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Temperature Sensor Failed

**Impact**: Availability

**MEL Event**: `MEL_EV_TEMP_SENSOR_FAIL` (0x281C)

Maximum temperature exceeded

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Temperature Sensor Missing

**Impact**: Availability

**MEL Event**: `MEL_EV_TEMP_SENSOR_MISSING` (0x281D)

Temperature sensor removed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Bypass Generic

**Impact**: Availability

**MEL Event**: `MEL_EV_BYPASS_GENERIC` (0x2823)

Drive by-passed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Cont Redundancy Loss

**Impact**: Availability

**MEL Event**: `MEL_EV_CONT_REDUNDANCY_LOSS` (0x2829)

Controller redundancy lost

### Tray Redundancy Loss

**Impact**: Availability

**MEL Event**: `MEL_EV_TRAY_REDUNDANCY_LOSS` (0x282B)

Drive tray path redundancy lost

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Drive Redundancy Loss

**Impact**: Availability

**MEL Event**: `MEL_EV_DRIVE_REDUNDANCY_LOSS` (0x282D)

Drive path redundancy lost

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Power Supply Failure

**Impact**: Availability

**MEL Event**: `MEL_EV_POWER_SUPPLY_FAIL` (0x283B)

Power supply failed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### ESM Drv Bypass

**Impact**: Availability

**MEL Event**: `MEL_EV_ESM_DRIVE_BYPASS` (0x2854)

Drive port bypassed - Error thresholds exceeded

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Drawer Open Or Removed

**Impact**: Availability

**MEL Event**: `MEL_EV_DRAWER_OPEN` (0x2857)

Drawer open or removed

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Power Supply No Power Input

**Impact**: Availability

**MEL Event**: `MEL_EV_POWER_SUPPLY_NO_INPUT` (0x2869)

Power supply has no power input

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### SSD Cache Failed Cache Size Mismatch

**Impact**: Performance

**MEL Event**: `MEL_EV_FLASH_CACHE_FAILED_CACHE_SIZE_MISMATCH` (0x3604)

SSD cache failed due to cache size mismatch on the two controllers

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### SSD Cache Non-Optimal Drives

**Impact**: Performance

**MEL Event**: `MEL_EV_FLASH_CACHE_NON_OPTIMAL_DRIVES` (0x3605)

SSD cache has associated non-optimal drives

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Disk Pool Reconstruction Reserved Drive Count Below Threshold

**Impact**: Capacity

**MEL Event**: `MEL_EV_DISK_POOL_REC_RDRVCNT_BEL_THRSHLD` (0x3803)

Disk pool reconstruction reserved drive count is below threshold

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Disk pool utilization warning

**Impact**: Capacity

**MEL Event**: `MEL_EV_DISK_POOL_UTILIZATION_WARNING` (0x3804)

Disk pool utilization exceeded the warning threshold

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Disk Pool Utilization Critical

**Impact**: Capacity

**MEL Event**: `MEL_EV_DISK_POOL_UTILIZATION_CRITICAL` (0x3805)

Disk pool utilization exceeded the critical threshold

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Disk Pool Capacity Depleted

**Impact**: Capacity

**MEL Event**: `MEL_EV_DISK_POOL_CAPACITY_DEPLETED` (0x3809)

All of the disk pool's free capacity has been used

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Disk Pool Insufficient Memory

**Impact**: Capacity

**MEL Event**: `MEL_EV_DISK_POOL_INSUFFICIENT_MEMORY` (0x380C)

Disk pool configuration has insufficient memory

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Invalid Pool Version

**Impact**: Configuration

**MEL Event**: `MEL_EV_DISK_POOL_INVALID_VERSION` (0x3811)

Lockdown due to invalid Disk Pool Version

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Vol Xfer Alert

**Impact**: Availability

**MEL Event**: `MEL_EV_VOL_XFER_ALERT` (0x4011)

Volume not on preferred path due to AVT/RDAC failover

### Set Controller Failed

**Impact**: Availability

**MEL Event**: `MEL_EV_SYMBOL_CONT_FAIL` (0x5005)

Place controller offline

### SYMbol Cont Service Mode

**Impact**: Availability

**MEL Event**: `MEL_EV_SYMBOL_CONT_SERVICE_MODE` (0x5040)

Place controller in service mode

### DBM Hck Altctl Not Func

**Impact**: Availability

**MEL Event**: `MEL_EV_DBM_HCK_ALTCTL_NOT_FUNC` (0x6107)

This controller's alternate is non-functional and is being held in reset

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### BBU Overheated

**Impact**: Protection

**MEL Event**: `MEL_EV_BBU_OVERHEATED` (0x7300)

Battery backup unit overheated

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Insufficient Learned Capacity

**Impact**: Protection

**MEL Event**: `MEL_EV_INSUFFICIENT_LEARNED_CAPACITY` (0x7301)

Insufficient learned battery capacity

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Battery Missing

**Impact**: Protection

**MEL Event**: `MEL_EV_BATTERY_MISSING` (0x7306)

Battery missing

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Drive Firmware Unverified

**Impact**: Configuration

**MEL Event**: `MEL_EV_DRIVE_NEW_DRIVE_FW_UNVERIFIED` (0x7E0C)

Down revision drive firmware detected

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Host Connection Redundancy Lost

**Impact**: Availability

**MEL Event**: `MEL_EV_HOST_REDUNDANCY_LOST` (0x9102)

Loss of host-side connection redundancy detected

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.

### Multipath Configuration Error

**Impact**: Configuration

**MEL Event**: `MEL_EV_MULTIPATH_CONFIG_ERROR` (0x9103)

Host multipath driver configuration error detected

**Remediation**

Open SANtricity System Manager and go to the Recovery Guru. If this problem is listed, follow the steps shown there. If it is not listed, the condition has already cleared.
