# REST Strategy

## Status <!-- one of: In Progress, Accepted, Rejected, Superseded, Deprecated -->

Accepted

## Context

ONTAP has published a customer product communiqué [(CPC-00410)](https://mysupport.netapp.com/info/communications/ECMLP2880232.html?access=a)
announcing that ZAPIs will reach end of availability (EOA) in ONTAP `9.13.1` released Q2 2023.

This document describes how Harvest handles the ONTAP transition from ZAPI to REST. 
In most cases, no action is required on your part.

## Harvest API Transition

Harvest attempts to use the collector defined in your `harvest.yml` config file.

When specifying the ZAPI collector, Harvest will use the ZAPI protocol unless the cluster no longer speaks Zapi,
in which cause, Harvest will switch to REST.

If you specify the REST collector, Harvest will use the REST protocol.

Harvest includes a full set of REST templates that export identical metrics as the included ZAPI templates.
No changes to dashboards or downstream metric-consumers should be required. 
See [below](#ive-added-counters-to-existing-zapi-templates-will-those-counters-work-in-rest) if you have 
added metrics to the Harvest out-of-the-box templates.

Read on if you want to know how you can use REST sooner, or you want to take advantage of REST-only features in ONTAP.

## Frequently Asked Questions

### How does Harvest decide whether to use REST or ZAPI APIs?

Harvest attempts to use the collector defined in your `harvest.yml` config file.
 
- If you specify the ZAPI collector, Harvest will use the ZAPI protocol as long as the cluster still speaks Zapi. 
  If the cluster no longer understands Zapi, Harvest will switch to Rest.

- If you specify the REST collector, Harvest will use REST.

Earlier versions of Harvest included a `prefer_zapi` poller option and a `HARVEST_NO_COLLECTOR_UPGRADE` environment variable.
Both of these options are ignored in Harvest versions `23.08` onwards.

### Why would I switch to REST before `9.13.1`?

- You have advanced use cases to validate before ONTAP removes ZAPIs
- You want to take advantage of new ONTAP features that are only available via REST (e.g., cloud features, event remediation, name services, cluster peers, etc.)
- You want to collect a metric that is not available via ZAPI
- You want to collect a metric from the ONTAP CLI. The REST API includes a private CLI pass-through to access any ONTAP CLI command

### Can I start using REST before `9.13.1`?

Yes. Many customers do. Be aware of the following limitations:

1. ONTAP includes a subset of performance counters via REST beginning in ONTAP [9.11.1](https://docs.netapp.com/us-en/ontap-automation/migrate/performance-counters.html#accessing-performance-counters-using-the-ontap-rest-api).
2. There may be performance metrics missing from versions of ONTAP earlier than `9.11.1`.

Where performance metrics are concerned, because of point #2,
our recommendation is to wait until at least ONTAP `9.12.1` before switching to the `RestPerf` collector.
You can continue using the `ZapiPerf` collector until you switch.

### A counter is missing from REST. What do I do?

The Harvest team has ensured
that all the out-of-the-box ZAPI templates have matching REST templates with identical metrics as of Harvest `22.11` and ONTAP `9.12.1`.
Any additional ZAPI Perf counters you have added may be missing from ONTAP REST Perf. 

Join the [Harvest discord channel](https://github.com/NetApp/harvest/blob/main/SUPPORT.md#getting-help) and ask us about the counter.
Sometimes we may know which release the missing counter is coming in, otherwise we can point you to the ONTAP
process to [request new counters](https://kb.netapp.com/Advice_and_Troubleshooting/Data_Storage_Software/ONTAP_OS/How_to_request_a_feature_for_ONTAP_REST_API).

### Can I use the REST and ZAPI collectors at the same time?

Yes. Harvest ensures that duplicate resources are not collected from both collectors.

When there is potential duplication, Harvest first resolves the conflict in the order collectors are defined in your
poller and then negotiates with the cluster on 
the most appropriate API to use [per above](#how-does-harvest-decide-whether-to-use-rest-or-zapi-apis).

Let's take a look at a few examples using the following poller definition:

```yaml
cluster-1:
    datacenter: dc-1
    addr: 10.1.1.1
    collectors:
        - Zapi
        - Rest
```

- When `cluster-1` is running ONTAP `9.9.X` (ONTAP still supports ZAPIs), the Zapi collector will be used since it is
  listed first in the list of `collectors`. When collecting a REST-only resource like, `nfs_client`, the Rest collector will be used
  since `nfs_client` objects are only available via REST.

- When `cluster-1` is running ONTAP `9.18.1` (ONTAP no longer supports ZAPIs),
  the Rest collector will be used since ONTAP can no longer speak the ZAPI protocol.

If you want the REST collector to be used in all cases, change the order in
the `collectors` section so `Rest` comes before `Zapi`.

If the resource does not exist for the first collector, the next collector will be tried.
Using the example above, when collecting `VolumeAnalytics` resources,
the Zapi collector will not run for `VolumeAnalytics` objects since that resource is only available via REST.
The Rest collector will run and collect the `VolumeAnalytics` objects.

### I've added counters to existing ZAPI templates. Will those counters work in REST?

`ZAPI` config metrics often have a REST equivalent that can be found in ONTAP's [ONTAPI to REST mapping document](https://library.netapp.com/ecm/ecm_download_file/ECMLP2882104).

ZAPI performance metrics may be missing in REST.
If you have added new metrics or templates to the `ZapiPerf` collector, those metrics likely aren't available via REST. 
You can [check if the performance counter is available](https://docs.netapp.com/us-en/ontap-automation/migrate/performance-counters.html#discover-the-available-performance-counter-tables), [ask the Harvest team on Discord](#a-counter-is-missing-from-rest-what-do-i-do),
or [ask ONTAP to add the counter](https://kb.netapp.com/Advice_and_Troubleshooting/Data_Storage_Software/ONTAP_OS/How_to_request_a_feature_for_ONTAP_REST_API) you need.

## Reference

Table of ONTAP versions, dates and API notes.

| ONTAP<br/>version | Release<br/>Date | ONTAP<br/>Notes                                                                                                                                                                                                     |
|------------------:|------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|          `9.11.1` | Q2 2022          | First version of ONTAP with [REST performance metrics](https://docs.netapp.com/us-en/ontap-automation/migrate/performance-counters.html#accessing-performance-counters-using-the-ontap-rest-api)                    |
|          `9.12.1` | Q4 2022          | ZAPIs still supported - REST performance metrics have parity with Harvest `22.11` collected ZAPI performance metrics                                                                                                |
| `9.14.1`-`9.15.1` |                  | ZAPIs enabled if ONTAP upgrade detects they were being used earlier. New ONTAP installs default to REST only. ZAPIs may be enabled via CLI                                                                          |
| `9.16.1`-`9.17.1` |                  | ZAPIs disabled. See [ONTAP communique](https://kb.netapp.com/onprem/ontap/dm/REST_API/FAQs_on_ZAPI_to_ONTAP_REST_API_transformation_for_CPC_(Customer_Product_Communiques)_notification) for details on re-enabling |
|          `9.18.1` |                  | ZAPIs removed. No way to re-enable                                                                                                                                                                                  |

