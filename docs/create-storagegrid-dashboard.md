This guide walks you through the process of creating custom Grafana dashboards in Harvest that are similar to StorageGRID's built-in dashboards.

## Step 1: Identify StorageGrid Dashboard Metrics

First examine the metrics used in StorageGrid's built-in dashboards to understand what needs to be collected by Harvest.

### Access StorageGrid's Native Dashboards

1. Log into your StorageGrid Grid Management Interface
2. Navigate to **SUPPORT > Tools > Metrics**
3. Access the built-in Grafana instance
4. Review the available dashboards

### Inspect Dashboard Queries

For each dashboard panel you want to check the Grafana panel to identify which metrics are being used. Note the metric names for your template configuration.

## Step 2: Configure Harvest to Collect StorageGrid Metrics

Now configure Harvest to collect the metrics identified in Step 1.

### Update the Metrics Template

Edit the StorageGrid metrics template to include the required metrics:

```bash
# Edit the template file
vim conf/storagegrid/11.6.0/storagegrid_metrics.yaml
```

Add the identified metrics to the `counters` section of the template file.

### Verify Metric Availability  

Before adding metrics to your template verify they're available from your StorageGrid Prometheus endpoint (usually found at `https://sg_ip/metrics/graph`).

## Step 3: Restart Harvest and Verify Collection

After updating the template, restart your Harvest poller to collect the new metrics. Verify that the metrics are available in your time-series database (Prometheus, InfluxDB, etc.)

## Step 4: Create Similar Dashboards in Harvest Grafana

Now create dashboards in your Harvest Grafana instance that mirror the functionality of StorageGRID's built-in dashboards, using the metrics being collected by Harvest.

For detailed guidance on creating dashboards with Harvest metrics, refer to the [Creating a Custom Grafana Dashboard](dashboards.md#creating-a-custom-grafana-dashboard-with-harvest-metrics-stored-in-prometheus) guide.

You can also use the existing [StorageGrid dashboards](https://github.com/NetApp/harvest/tree/main/grafana/dashboards/storagegrid) as references for panel designs and query patterns.

Note that StorageGrid and Harvest may use different label names for the same concepts so update your dashboard queries to use Harvest's label conventions as needed.


## Troubleshooting

### Common Issues and Solutions

- **Metrics not appearing in Prometheus**: Check that metrics are available in StorageGrid's native Prometheus endpoint. Verify template syntax and restart poller. Check Harvest logs for collection errors. 

- **Dashboard panels showing "No data"**: Verify metric names match exactly between template and dashboard. Check label selectors and variable names. Ensure time range covers period when metrics were collected.

## Related Documentation

- [Configure StorageGrid Collector](configure-storagegrid.md)
- [StorageGrid Metrics Reference](storagegrid-metrics.md)  
- [Creating Custom Dashboards](dashboards.md#creating-a-custom-grafana-dashboard-with-harvest-metrics-stored-in-prometheus)