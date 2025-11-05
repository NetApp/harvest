
## Prepare Amazon FSx for ONTAP

To set up Harvest and FSx make sure you read through 
[Monitoring FSx for ONTAP file systems using Harvest and Grafana](https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/monitoring-harvest-grafana.html)

## Supported Harvest Dashboards

Amazon FSx for ONTAP exposes a different set of metrics than ONTAP cDOT.
That means a limited set of out-of-the-box dashboards are supported and
some panels may be missing information.

You can also enable the [KeyPerf](configure-keyperf.md) collector for FSx systems to collect performance metrics that are not available via the `ZapiPerf/RestPerf` collector.

The dashboards that work with FSx are tagged with `fsx` and listed below:

* ONTAP: cDOT
* ONTAP: Cluster
* ONTAP: Data Protection
* ONTAP: Datacenter
* ONTAP: FlexCache
* ONTAP: FlexGroup
* ONTAP: FPolicy
* ONTAP: LUN
* ONTAP: NFS Troubleshooting
* ONTAP: Quota
* ONTAP: Security
* ONTAP: SVM
* ONTAP: Volume
* ONTAP: Volume by SVM
* ONTAP: Volume Deep Dive
