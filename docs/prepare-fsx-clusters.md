
## Prepare Amazon FSx for ONTAP

To set up Harvest and FSx make sure you read through 
[Monitoring FSx for ONTAP file systems using Harvest and Grafana](https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/monitoring-harvest-grafana.html)

## Supported Harvest Dashboards

Amazon FSx for ONTAP exposes a different set of metrics than ONTAP cDOT.
That means a limited set of out-of-the-box dashboards are supported and
some panels may be missing information. 

The dashboards that work with FSx are tagged with `fsx` and listed below:

* ONTAP: Volume
* ONTAP: SVM
* ONTAP: Security
* ONTAP: Data Protection Snapshots
* ONTAP: Compliance