# REST Strategy

## Status <!-- one of: In Progress, Accepted, Rejected, Superseded, Deprecated -->

In Progress

## Context

ONTAP has published a customer product communiqué [(CPC-00410)](https://mysupport.netapp.com/info/communications/ECMLP2880232.html?access=a)
announcing that ZAPIs will reach end of availability (EOA) in ONTAP `9.13.1` released Q2 2023.

This document describes how Harvest is making the ONTAP transition, from ZAPI to REST, seamless for Harvest customers.

## Harvest API Transition

By default, Harvest will use ZAPIs up until ONTAP version `9.12.1`.
Beginning with ONTAP `9.12.1` and after, Harvest will default to REST.

Harvest includes a full set of REST templates that export identical metrics as the included ZAPI templates.
No changes to dashboards or downstream metric-consumers will be required.

Read on if you want to know how you can use REST sooner, or you want to take advantage of REST-only features in ONTAP.

## Frequently Asked Questions

### How does Harvest decide whether to use REST or ZAPI APIs?

Harvest picks the appropriate API based on the following rules:

1. If you specify a particular collector in your `harvest.yml`, Harvest will use it.
2. Otherwise, Harvest will ask the cluster if it supports ZAPIs.
If the cluster says yes, Harvest will use ZAPIs, otherwise it will use REST. 

### Why would I switch to REST before `9.13.1`?

- You have advanced use cases to validate before ONTAP removes ZAPIs in `9.13.1`
- You want to take advantage of new ONTAP features that are only available via REST (e.g. cloud features, event remediation's, name services, cluster peers, etc.)
- You want to collect a metric that is not available via ZAPI
- You want to collect a metric from the ONTAP CLI. The REST API includes a private CLI pass-through to access any ONTAP CLI command

### Can I start using REST before `9.13.1`?

Yes. Several customers already are. There are a few caveats to be aware of:

1. Harvest collects config counters via REST by enabling the `Rest` collector in your `harvest.yml`,
   but ONTAP did not include performance counters via REST until [9.11.1](https://docs.netapp.com/us-en/ontap-automation/migrate/performance-counters.html#accessing-performance-counters-using-the-ontap-rest-api).
   That means Harvest's `RestPerf` collector won't work until `9.11.1`.
   ONTAP supports a subset of performance counters in `9.11.1`. The full set is available in `9.12.1`.

2. It's better to publish a set of metrics once instead of multiple times. In other words, it does not make sense to
   enable both the `Zapi` and `Rest` collector for an overlapping set of objects on the same cluster.
   It will work, but you'll put more load on the cluster and push duplicate metrics to Prometheus. 
   See [belo](#can-i-use-the-rest-and-zapi-collectors-at-the-same-time) for details on how to use both collectors at the same time. 

3. There may be performance metrics missing from versions of ONTAP earlier than `9.11.1`.

### A counter is missing from REST. What do I do?

The Harvest team has ensured that all of the out-of-the-box ZAPI templates have matching REST templates with the same metrics.
Any additional counters you have added may be missing in REST. 

Join the [Harvest discord channel](https://github.com/NetApp/harvest/blob/main/SUPPORT.md#getting-help) and ask us about the counter.
Sometimes we may know which release the missing counter is coming in, otherwise we can point you to the ONTAP
process to [request new counters](https://kb.netapp.com/Advice_and_Troubleshooting/Data_Storage_Software/ONTAP_OS/How_to_request_a_feature_for_ONTAP_REST_API).

### Can I use the REST and ZAPI collectors at the same time?

Yes. 

It's best when using both collectors to ensure that you aren't collecting the same object(s) multiple times. 
For example, there is nothing to be gained by collecting `disk` from both collectors. 
Harvest won't do anything to prevent you from doing that, but our recommendation when using both collectors, is to use a non-overlapping set of objects.

Typically, you will use ZAPI collectors with the out-of-the-box templates and add new REST templates for new objects. 
For example, if you want to [collect controller RAM status](https://github.com/NetApp/harvest/discussions/1187) you must use the REST collector,
since there is no ZAPI that returns that metric.

## Reference

Table of ONTAP versions, dates and API notes.

| ONTAP<br/>version | Release<br/>Date | ONTAP<br/>Notes                                                                                                                                                                         |
|------------------:|------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|          `9.11.1` | Q2 2022          | First version with [REST performance metrics](https://docs.netapp.com/us-en/ontap-automation/migrate/performance-counters.html#accessing-performance-counters-using-the-ontap-rest-api) |
|          `9.12.1` | Q4 2022          | ZAPIs still supported - REST performance metrics have parity with Harvest collected ZAPI performance metrics                                                                            |
|          `9.13.1` | Q2 2023          | ZAPIs removed. REST only release - REST config and performance parity with ZAPIs                                                                                                        |
