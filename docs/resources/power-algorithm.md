
Gathering power metrics requires a cluster with:

* ONTAP versions 9.6+
* [REST enabled](../prepare-cdot-clusters.md), even when using the ZAPI collector

REST is required because it is the only way to collect chassis field-replaceable-unit (FRU) information via the
REST API `/api/private/cli/system/chassis/fru`.

## How does Harvest calculate cluster power?

Cluster power is the sum of a cluster's [node(s) power](#node-power) + 
the sum of attached [disk shelve(s) power](#disk-shelf-power).

Redundant power supplies (PSU) load-share the total load.
With n PSUs, each PSU does roughly (1/n) the work
(the actual amount is slightly more than a single PSU due to additional fans.)

## Node power

Node power is calculated by collecting power supply unit (PSU) power, as reported by REST 
`/api/private/cli/system/environment/sensors` or by ZAPI `environment-sensors-get-iter`.

When a power supply is shared between controllers,
the PSU's power will be evenly divided across the controllers due to load-sharing.

For example:

* FAS2750 models have two power supplies that power both controllers. Each PSU is shared between the two controllers.
* A800 models have four power supplies. `PSU1` and `PSU2` power `Controller1` and `PSU3` and `PSU4` power `Controller2`. Each PSU provides power to a single controller.

Harvest determines whether a PSU is shared between controllers by consulting the `connected_nodes` of each PSU,
as reported by ONTAP via `/api/private/cli/system/chassis/fru`

## Disk shelf power

Disk shelf power is calculated by collecting `psu.power_drawn`, as reported by REST, via
`/api/storage/shelves` or `sensor-reading`, as reported by ZAPI `storage-shelf-info-get-iter`.

The power for [embedded shelves](https://kb.netapp.com/onprem/ontap/hardware/FAQ%3A_How_do_shelf_product_IDs_and_modules_in_ONTAP_map_to_a_model_of_a_shelf_or_storage_system_with_embedded_storage)
is ignored, since that power is already accounted for in the controller's power draw.

## Examples

### FAS2750

```
# Power Metrics for 10.61.183.200

## ONTAP version NetApp Release 9.8P16: Fri Dec 02 02:05:05 UTC 2022

## Nodes
system show
       Node         |  Model  | SerialNumber  
----------------------+---------+---------------
cie-na2750-g1344-01 | FAS2750 | 621841000123  
cie-na2750-g1344-02 | FAS2750 | 621841000124

## Chassis
system chassis fru show
 ChassisId   |      Name       |         Fru         |    Type    | Status | NumNodes |              ConnectedNodes               
---------------+-----------------+---------------------+------------+--------+----------+-------------------------------------------
021827030435 | 621841000123    | cie-na2750-g1344-01 | controller | ok     |        1 | cie-na2750-g1344-01                       
021827030435 | 621841000124    | cie-na2750-g1344-02 | controller | ok     |        1 | cie-na2750-g1344-02                       
021827030435 | PSQ094182201794 | PSU2 FRU            | psu        | ok     |        2 | cie-na2750-g1344-02, cie-na2750-g1344-01  
021827030435 | PSQ094182201797 | PSU1 FRU            | psu        | ok     |        2 | cie-na2750-g1344-02, cie-na2750-g1344-01

## Sensors
system environment sensors show
(filtered by power, voltage, current)
       Node         |     Name      |  Type   | State  | Value | Units  
----------------------+---------------+---------+--------+-------+--------
cie-na2750-g1344-01 | PSU1 12V Curr | current | normal |  9920 | mA     
cie-na2750-g1344-01 | PSU1 12V      | voltage | normal | 12180 | mV     
cie-na2750-g1344-01 | PSU1 5V Curr  | current | normal |  4490 | mA     
cie-na2750-g1344-01 | PSU1 5V       | voltage | normal |  5110 | mV     
cie-na2750-g1344-01 | PSU2 12V Curr | current | normal |  9140 | mA     
cie-na2750-g1344-01 | PSU2 12V      | voltage | normal | 12100 | mV     
cie-na2750-g1344-01 | PSU2 5V Curr  | current | normal |  4880 | mA     
cie-na2750-g1344-01 | PSU2 5V       | voltage | normal |  5070 | mV     
cie-na2750-g1344-02 | PSU1 12V Curr | current | normal |  9920 | mA     
cie-na2750-g1344-02 | PSU1 12V      | voltage | normal | 12180 | mV     
cie-na2750-g1344-02 | PSU1 5V Curr  | current | normal |  4330 | mA     
cie-na2750-g1344-02 | PSU1 5V       | voltage | normal |  5110 | mV     
cie-na2750-g1344-02 | PSU2 12V Curr | current | normal |  9170 | mA     
cie-na2750-g1344-02 | PSU2 12V      | voltage | normal | 12100 | mV     
cie-na2750-g1344-02 | PSU2 5V Curr  | current | normal |  4720 | mA     
cie-na2750-g1344-02 | PSU2 5V       | voltage | normal |  5070 | mV

## Shelf PSUs
storage shelf show
Shelf | ProductId | ModuleType | PSUId | PSUIsEnabled | PSUPowerDrawn | Embedded  
------+-----------+------------+-------+--------------+---------------+---------
  1.0 | DS224-12  | iom12e     | 1,2   | true,true    | 1397,1318     | true

### Controller Power From Sum(InVoltage * InCurrent)/NumNodes
Power: 256W
```

### AFF A800

```
# Power Metrics for 10.61.124.110

## ONTAP version NetApp Release 9.13.1P1: Tue Jul 25 10:19:28 UTC 2023

## Nodes
system show
  Node    |  Model   | SerialNumber  
----------+----------+-------------
a800-1-01 | AFF-A800 | 941825000071  
a800-1-02 | AFF-A800 | 941825000072

## Chassis
system chassis fru show
   ChassisId    |      Name      |    Fru    |    Type    | Status | NumNodes | ConnectedNodes  
----------------+----------------+-----------+------------+--------+----------+---------------
SHFFG1826000154 | 941825000071   | a800-1-01 | controller | ok     |        1 | a800-1-01       
SHFFG1826000154 | 941825000072   | a800-1-02 | controller | ok     |        1 | a800-1-02       
SHFFG1826000154 | EEQT1822002800 | PSU1 FRU  | psu        | ok     |        1 | a800-1-02       
SHFFG1826000154 | EEQT1822002804 | PSU2 FRU  | psu        | ok     |        1 | a800-1-02       
SHFFG1826000154 | EEQT1822002805 | PSU2 FRU  | psu        | ok     |        1 | a800-1-01       
SHFFG1826000154 | EEQT1822002806 | PSU1 FRU  | psu        | ok     |        1 | a800-1-01

## Sensors
system environment sensors show
(filtered by power, voltage, current)
  Node    |     Name      |  Type   | State  | Value | Units  
----------+---------------+---------+--------+-------+------
a800-1-01 | PSU1 Power In | unknown | normal |   376 | W      
a800-1-01 | PSU2 Power In | unknown | normal |   411 | W      
a800-1-02 | PSU1 Power In | unknown | normal |   383 | W      
a800-1-02 | PSU2 Power In | unknown | normal |   433 | W

## Shelf PSUs
storage shelf show
Shelf |  ProductId  | ModuleType | PSUId | PSUIsEnabled | PSUPowerDrawn | Embedded  
------+-------------+------------+-------+--------------+---------------+---------
  1.0 | FS4483PSM3E | psm3e      |       |              |               | true      

### Controller Power From Sum(InPower sensors)
Power: 1603W
```