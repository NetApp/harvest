Welcome to the NetApp Harvest getting started guide. This tutorial will walk you through the steps needed to deploy a basic instance of NetApp Harvest on a Linux platform.

### 1. Set Installation Path

First, set the installation path as an environment variable. By default, we'll use `/opt`.

```bash
INSTALL_PATH=/opt
```

### 2. Install Harvest

Harvest is distributed as a container, tarball, and RPM and Debs. Pick the one that works best for you.
More details can be found in the [installation](https://netapp.github.io/harvest/latest/install/overview/) documentation.

For this guide, we'll use the tarball package as an example.

Visit the [Releases page](https://github.com/NetApp/harvest/releases) and copy the `tar.gz` link for the latest release. For example, to download the `24.05.2` release:

```bash
HARVEST_VERSION=24.05.2
cd $INSTALL_PATH
wget https://github.com/NetApp/harvest/releases/download/v${VERSION}/harvest-${HARVEST_VERSION}-1_linux_amd64.tar.gz
tar -xvf harvest-${HARVEST_VERSION}-1_linux_amd64.tar.gz
cd harvest-${HARVEST_VERSION}-1_linux_amd64
```

### 3. Install Prometheus

To install Prometheus, follow these steps:

```bash
PROMETHEUS_VERSION=2.49.1
cd $INSTALL_PATH
wget https://github.com/prometheus/prometheus/releases/download/v${PROMETHEUS_VERSION}/prometheus-${PROMETHEUS_VERSION}.linux-amd64.tar.gz
tar -xvf prometheus-${PROMETHEUS_VERSION}.linux-amd64.tar.gz
mv prometheus-${PROMETHEUS_VERSION}.linux-amd64 prometheus-${PROMETHEUS_VERSION}
```

Create a service file for Prometheus:

```bash
sudo vi /etc/systemd/system/prometheus.service
```

Add the following content:

```ini
[Unit]
Description=Prometheus Server
Documentation=https://prometheus.io/docs/introduction/overview/
After=network-online.target

[Service]
User=root
Restart=on-failure
ExecStart=<INSTALL_PATH>/prometheus-<PROMETHEUS_VERSION>/prometheus --config.file=<INSTALL_PATH>/prometheus-<PROMETHEUS_VERSION>/prometheus.yml

[Install]
WantedBy=multi-user.target
```

For example, if `INSTALL_PATH` is `/opt` and `PROMETHEUS_VERSION` is `2.49.1`, the `ExecStart` line would be:

```ini
ExecStart=/opt/prometheus-2.49.1/prometheus --config.file=/opt/prometheus-2.49.1/prometheus.yml
```

Reload the systemd configuration and start Prometheus:

```bash
sudo systemctl daemon-reload
sudo systemctl enable prometheus
sudo systemctl start prometheus
```

Check if Prometheus is up and running:

```bash
sudo systemctl status prometheus
```

You should see output indicating that the Prometheus service is active and running.

### 4. Install Grafana

To install Grafana, follow these steps:

```bash
GRAFANA_VERSION=10.4.5
cd $INSTALL_PATH
wget https://dl.grafana.com/oss/release/grafana-${GRAFANA_VERSION}.linux-amd64.tar.gz
tar -xvf grafana-${GRAFANA_VERSION}.linux-amd64.tar.gz
```

Create a service file for Grafana:

```bash
sudo vi /etc/systemd/system/grafana.service
```

Add the following content:

```ini
[Unit]
Description=Grafana Server
Documentation=https://grafana.com/docs/grafana/latest/setup-grafana/installation/
After=network-online.target

[Service]
User=root
Restart=on-failure
ExecStart=<INSTALL_PATH>/grafana-v<GRAFANA_VERSION>/bin/grafana-server --config=<INSTALL_PATH>/grafana-v<GRAFANA_VERSION>/conf/defaults.ini --homepath=<INSTALL_PATH>/grafana-v<GRAFANA_VERSION>

[Install]
WantedBy=multi-user.target
```

For example, if `INSTALL_PATH` is `/opt` and `GRAFANA_VERSION` is `10.4.5`, the `ExecStart` line would be:

```ini
ExecStart=/opt/grafana-v10.4.5/bin/grafana-server --config=/opt/grafana-v10.4.5/conf/defaults.ini --homepath=/opt/grafana-v10.4.5
```

Reload the systemd configuration and start Grafana:

```bash
sudo systemctl daemon-reload
sudo systemctl enable grafana
sudo systemctl start grafana
```

Check if Grafana is up and running:

```bash
sudo systemctl status grafana
```

You should see output indicating that the Grafana service is active and running.

### 5. Configuration File

Harvest's configuration information is defined in `harvest.yml`. There are a few ways to tell Harvest how to load this file:

- If you don't use the `--config` flag, the `harvest.yml` file located in the current working directory will be used.
- If you specify the `--config` flag like so `harvest status --config /opt/harvest/harvest.yml`, Harvest will use that file.

To start collecting metrics, you need to define at least one `poller` and one `exporter` in your configuration file.
The default configuration comes with a pre-configured poller named `unix` which collects metrics from the local system.
This is useful if you want to monitor resource usage by Harvest and serves as a good example. Feel free to delete it if you want.

The next step is to add pollers for your ONTAP clusters in the [Pollers](configure-harvest-basic.md#pollers) section of the Harvest configuration file, `harvest.yml`.

Edit the Harvest configuration file:

```sh
cd $INSTALL_PATH/harvest-${VERSION}-1_linux_amd64
sudo vi harvest.yml
```

Make the necessary changes to monitor your ONTAP system. Example configuration:

```yaml
Exporters:
  prometheus1:
    exporter: Prometheus
    port_range: 13000-13100

Defaults:
  collectors:
    - Zapi
    - ZapiPerf
    - Ems
    - Rest
    - RestPerf
  use_insecure_tls: true

Pollers:
  unix:
    datacenter: local
    addr: localhost
    collectors:
      - Unix
    exporters:
      - prometheus1

  jamaica:
    datacenter: DC-01
    addr: ClusterManagementIP
    auth_style: basic_auth
    username: YourUsername
    password: YourPassword
    exporters:
      - prometheus1
```

### 6. Edit Prometheus Config File

Edit the Prometheus configuration file:

```sh
cd $INSTALL_PATH/prometheus-${PROMETHEUS_VERSION}
sudo vi prometheus.yml
```

Add the following lines to the existing `scrape_configs` block:

```yaml
  - job_name: 'harvest'
    static_configs:
      - targets: ['localhost:13000', 'localhost:13001', 'localhost:13002']  # Add ports as defined in the port range
```

For example, if your port range in the Harvest configuration is `13000-13100`, you should add the ports within this range that you plan to use.

Restart Prometheus to apply the changes:

```bash
sudo systemctl restart prometheus
```

Check if Prometheus is up and running:

```bash
sudo systemctl status prometheus
```

### 7. Start Harvest

Start all Harvest pollers as daemons:

```bash
cd $INSTALL_PATH/harvest-${VERSION}-1_linux_amd64
bin/harvest start
```

Or start specific poller(s). In this case, we're starting two pollers named `jamaica` and `grenada`:

```bash
bin/harvest start jamaica grenada
```

Replace `jamaica` and `grenada` with the poller names you defined in `harvest.yml`. The logs of each poller can be found in `/var/log/harvest/`.

### 8. Add Prometheus Datasource in Grafana

From the "three lines" button at the top left, navigate to **Connections** and then **Data Sources**.

1. Click on **Add data source**.
2. Select **Prometheus**.
3. Enter `http://localhost:9090` in the Prometheus server URL field.
4. Click on **Save and test**.

At the bottom, you should see the message **'Successfully queried the Prometheus API.'**

For detailed instructions, please refer to the [Configure Prometheus Data Source](https://grafana.com/docs/grafana/latest/datasources/prometheus/configure-prometheus-data-source/) documentation.

### 9. Generate Grafana API Token

To import Grafana dashboards using the `bin/harvest grafana import` command, you need a Grafana API token. Follow these steps to generate it:

1. Log in to Grafana at `http://localhost:3000`.
2. Click on the "three lines" button at the top left of the screen to open the menu.
3. Navigate to **Administration** and then select **Service Account**.
4. Click on **Add Service Account**.
5. Enter the display name **Harvest**.
6. Set the role to **Editor**.
7. Click on **Create**. The service account will appear in the dashboard.
8. Navigate back to **Service Account**.
9. Click on **Add Token** for the Harvest service account.
10. Click on **Generate Token**.
11. Click on **Copy to clipboard and close**.

**IMPORTANT:** This is the only opportunity to save the token. Immediately paste it into a text file and save it. The token will be needed by Harvest later on.

For detailed instructions, please refer to the [Grafana API Keys documentation](https://grafana.com/docs/grafana/latest/administration/api-keys/).

### 10. Import Grafana Dashboards

To import Grafana dashboards, use the following command:

```bash
cd $INSTALL_PATH/harvest-${HARVEST_VERSION}-1_linux_amd64
bin/harvest grafana import --addr localhost:3000
```

You will be prompted to enter the Grafana API token. Paste the token you generated in the previous step.

```bash
#### You will be requested to 'Enter API Token'
#### Enter the token saved before
#### You will be asked to save token for later use, enter 'Y'

### If you have removed and re-installed Grafana, you will get an error telling that current API Key is not valid any more
### You need to enter the new API Key

using API token from config
.error connect: (401 - 401 Unauthorized) Unauthorized
enter API token:
```

It will take a few seconds, but at the end, all dashboards will be imported.

### 8. Verify the Metrics

If you use a Prometheus Exporter, open a browser and navigate to [http://0.0.0.0:12990/](http://0.0.0.0:12990/)
(replace `12990` with the port number of your poller).
This is the Harvest created HTTP end-point for your Prometheus exporter.
This page provides a real-time generated list of running collectors and names of exported metrics.

The metric data that is exported for Prometheus to scrape is
available at [http://0.0.0.0:12990/metrics/](http://0.0.0.0:12990/metrics/).

More information on configuring the exporter can be found in the
[Prometheus exporter](prometheus-exporter.md) documentation.

If you can't access the URL, check the logs of your pollers. These are located in `/var/log/harvest/`.

### Troubleshooting

If you encounter issues, check the logs in `/var/log/harvest` and refer to the troubleshooting section on the wiki.
You can also reach out for help on [Discord](https://github.com/NetApp/harvest/blob/main/SUPPORT.md#getting-help) or via email at ng-harvest-files@netapp.com.

### Conclusion

Congratulations! You have successfully set up NetApp Harvest along with Prometheus and Grafana. Enjoy monitoring your systems and feel free to provide feedback.