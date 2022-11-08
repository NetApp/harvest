## 1. Configuration file

Harvest's configuration information is defined in `harvest.yml`. There are a few ways to tell Harvest how to load this file:

* If you don't use the `--config` flag, the `harvest.yml` file located in the current working directory will be used

* If you specify the `--config` flag like so `harvest status --config /opt/harvest/harvest.yml`, Harvest will use that file

To start collecting metrics, you need to define at least one `poller` and one `exporter` in your  configuration file.
The default configuration comes with a pre-configured poller named `unix` which collects metrics from the local system. 
This is useful if you want to monitor resource usage by Harvest and serves as a good example. 
Feel free to delete it if you want.

The next step is to add pollers for your ONTAP clusters in the [Pollers](configure-harvest-basic.md#pollers) 
section of the Harvest configuration file, `harvest.yml`.

## 2. Start Harvest

Start all Harvest pollers as daemons:

```bash
bin/harvest start
```

Or start a specific poller(s). In this case, we're staring two pollers named `jamaica` and `jamaica`.

```bash
bin/harvest start jamaica jamaica
```

Replace `jamaica` and `grenada` with the poller names you defined in `harvest.yml`. 
The logs of each poller can be found in `/var/log/harvest/`.

## 3. Import Grafana dashboards

The Grafana dashboards are located in the `$HARVEST_HOME/grafana` directory. 
You can manually import the dashboards or use the `bin/harvest grafana` command
([more documentation](dashboards.md)).

Note: the current dashboards specify Prometheus as the datasource. 
If you use the InfluxDB exporter, you will need to create your own dashboards.

## 4. Verify the metrics

If you use a Prometheus Exporter, open a browser and navigate to [http://0.0.0.0:12990/](http://0.0.0.0:12990/)
(replace `12990` with the port number of your poller). 
This is the Harvest created HTTP end-point for your Prometheus exporter. 
This page provides a real-time generated list of running collectors and names of exported metrics.

The metric data that is exported for Prometheus to scrap is 
available at [http://0.0.0.0:12990/metrics/](http://0.0.0.0:12990/metrics/). 

More information on configuring the exporter can be found in the
[Prometheus exporter](prometheus-exporter.md) documentation.

If you can't access the URL, check the logs of your pollers. These are located in `/var/log/harvest/`.

## 5. (Optional) Setup Systemd service files

If you're running Harvest on a system with Systemd, you may want 
to [take advantage of systemd instantiated units](https://github.com/NetApp/harvest/tree/main/service/contrib) 
to manage your pollers.  
