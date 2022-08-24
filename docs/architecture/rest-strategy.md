# REST Strategy

## Status <!-- one of: In Progress, Accepted, Rejected, Superseded, Deprecated -->

In Progress

## Context

ONTAP has published a customer product communiqué [(CPC-00410)](https://mysupport.netapp.com/info/communications/ECMLP2880232.html?access=a)
announcing that NetApp ONTAPI (ZAPI) will reach end of availability (EOA) in January 2023.

Table of ONTAP version, dates and API notes.

| ONTAP<br/>version | Release<br/>Date | ONTAP API Notes                                                                                                                                                                         |
|------------------:|------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|          `9.11.1` | Q2 2022          | First version with [REST performance metrics](https://docs.netapp.com/us-en/ontap-automation/migrate/performance-counters.html#accessing-performance-counters-using-the-ontap-rest-api) |
|          `9.12.1` | Q4 2022          | ZAPIs still supported - REST performance metrics have parity with Harvest collected ZAPI performance metrics                                                                            |
|          `9.13.1` | Q2 2023          | ZAPIs removed. REST only release - REST config and performance parity with ZAPIs                                                                                                        |

## Harvest API Transition

Harvest works with all previous versions of ONTAP from cDot to 7mode. 

ONTAP's transition from ZAPI to REST will be seamless for Harvest customers.

By the time ONTAP `9.12.1` is released:
- Harvest will pick the most appropriate API (ZAPI or REST) according to the underlying ONTAP version
- All of Harvest's included ZAPI templates will have equivalent REST templates
- Harvest REST templates will publish identical metrics as the ZAPI ones, and no changes to dashboards or downstream metrics consumers will be required

## Frequently Asked Questions

### How does Harvest decide which API to use?

Harvest release `22.11` introduces two new collectors named: `Ontap` and `OntapPerf`. 
These new collectors will interrogate the cluster and pick the most appropriate API (ZAPI or REST).

Harvest picks the appropriate API based on the following rules:

1. If you specify a particular collector in your `harvest.yml`, Harvest will use it. 
2. If the cluster version is `9.12.1` or earlier, and you specify the `Ontap` collector, the ZAPI collector is used by default
3. If the cluster is `9.13.1` or later, and you specify the `Ontap` collector, the REST collector is used by default

**Examples**

| ONTAP<br/>version | Collector<br/>in `harvest.yml` | API Picked       |
|------------------:|--------------------------------|------------------|
|           `9.8.X` | Zapi                           | Zapi             |
|           `9.8.X` | Rest                           | Rest             |
|           `9.8.X` | Ontap                          | Zapi             |
|          `9.12.X` | Ontap                          | Zapi             |
|          `9.13.X` | Zapi                           | Zapi (will fail) |
|          `9.13.X` | Ontap                          | Rest             |

### Why would I switch to REST before 9.12.1?

- Newer ONTAP features are only available via REST (e.g. cloud features, event remediation's, name services, cluster peers, etc.)
- You need to collect a configuration metric that is not available via ZAPI
- You need to collect something from the ONTAP CLI. The REST API supports a private CLI pass-through to access any ONTAP CLI command
- You want to ensure a smooth transition from ZAPI to REST and have advanced use cases to validate

### Can I start using REST before `9.12.1`?

Yes. Several customers already are. There are a few caveats to be aware of:

1. Harvest can collect config counters via REST by enabling the `Rest` collector in your `harvest.yml`, 
but ONTAP did not include performance counters via REST until [9.11.1](https://docs.netapp.com/us-en/ontap-automation/migrate/performance-counters.html#accessing-performance-counters-using-the-ontap-rest-api).
That means Harvest's `RestPerf` collector won't work until `9.11.1` or later.
Be mindful that ONTAP supports a subset of performance counters in `9.11.1`, and the full set is not included until `9.12.1`. 
 
2. It's better to not publish the same metrics from multiple collectors. In other words, don't enable  
both the `Zapi` and `Rest` collector for an overlapping set of objects on the same cluster. 
It will work, but you'll put more load on the cluster and push duplicate metrics to Prometheus.

3. Most of the REST config templates will work before `9.11`, but there may be metrics missing from earlier versions of ONTAP.

### A counter is missing from REST. What do I do?

Join the [Harvest discord channel](https://github.com/NetApp/harvest/blob/main/SUPPORT.md#getting-help) and ask us about the counter. 
Sometimes we may know which release the missing counter is coming in, otherwise we can point you to the ONTAP 
process to [request new counters](https://kb.netapp.com/Advice_and_Troubleshooting/Data_Storage_Software/ONTAP_OS/How_to_request_a_feature_for_ONTAP_REST_API).  

