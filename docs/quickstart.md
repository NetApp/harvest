Welcome to the NetApp Harvest Getting Started Guide. This tutorial will guide you through the steps required to deploy a basic instance of NetApp Harvest on a Linux platform to monitor an ONTAP cluster.

### 1. Set Installation Path

First, set the installation path as an environment variable. For example, we'll use `/opt/netapp/harvest`.

```bash
HARVEST_INSTALL_PATH=/opt/netapp/harvest
mkdir -p ${HARVEST_INSTALL_PATH}
```

### 2. Install Harvest

Harvest is distributed as a container, tarball, and RPM and Debs. Pick the one that works best for you.
More details can be found in the [installation](https://netapp.github.io/harvest/latest/install/overview/) documentation.

For this guide, we'll use the tarball package as an example.

Visit the [Releases page](https://github.com/NetApp/harvest/releases) and copy the `tar.gz` link for the latest release. For example, to download the `24.05.2` release:

```bash
HARVEST_VERSION=24.05.2
cd ${HARVEST_INSTALL_PATH}
wget https://github.com/NetApp/harvest/releases/download/v${HARVEST_VERSION}/harvest-${HARVEST_VERSION}-1_linux_amd64.tar.gz
tar -xvf harvest-${HARVEST_VERSION}-1_linux_amd64.tar.gz
```

### 3. Install Prometheus

To install Prometheus, follow these steps:

```bash
PROMETHEUS_VERSION=2.49.1
cd ${HARVEST_INSTALL_PATH}
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
ExecStart=<HARVEST_INSTALL_PATH>/prometheus-<PROMETHEUS_VERSION>/prometheus --config.file=<HARVEST_INSTALL_PATH>/prometheus-<PROMETHEUS_VERSION>/prometheus.yml

[Install]
WantedBy=multi-user.target
```

For example, if `HARVEST_INSTALL_PATH` is `/opt/netapp/harvest` and `PROMETHEUS_VERSION` is `2.49.1`, the `ExecStart` line would be:

```ini
ExecStart=/opt/netapp/harvest/prometheus-2.49.1/prometheus --config.file=/opt/netapp/harvest/prometheus-2.49.1/prometheus.yml
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
cd ${HARVEST_INSTALL_PATH}
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
ExecStart=<HARVEST_INSTALL_PATH>/grafana-v<GRAFANA_VERSION>/bin/grafana-server --config=<HARVEST_INSTALL_PATH>/grafana-v<GRAFANA_VERSION>/conf/defaults.ini --homepath=<HARVEST_INSTALL_PATH>/grafana-v<GRAFANA_VERSION>

[Install]
WantedBy=multi-user.target
```

For example, if `HARVEST_INSTALL_PATH` is `/opt/netapp/harvest` and `GRAFANA_VERSION` is `10.4.5`, the `ExecStart` line would be:

```ini
ExecStart=/opt/netapp/harvest/grafana-v10.4.5/bin/grafana-server --config=/opt/netapp/harvest/grafana-v10.4.5/conf/defaults.ini --homepath=/opt/netapp/harvest/grafana-v10.4.5
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
cd ${HARVEST_INSTALL_PATH}/harvest-${HARVEST_VERSION}-1_linux_amd64
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
  jamaica:
    datacenter: DC-01
    addr: ClusterManagementIP
    auth_style: basic_auth
    username: YourUsername
    password: YourPassword
    exporters:
      - prometheus1
```

**Note:** The ONTAP user specified in this configuration must have the appropriate permissions as outlined in the [Prepare cDot Clusters](/prepare-cdot-clusters/#least-privilege-approach) documentation.

### 6. Edit Prometheus Config File

Edit the Prometheus configuration file:

```sh
cd ${HARVEST_INSTALL_PATH}/prometheus-${PROMETHEUS_VERSION}
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

To start the Harvest pollers, follow these steps:

```bash
cd ${HARVEST_INSTALL_PATH}/harvest-${HARVEST_VERSION}-1_linux_amd64
bin/harvest start
```

Verify that the pollers have started successfully by checking their status:

```bash
bin/harvest status
```

The output should look similar to this:

```
Datacenter | Poller  |   PID   | PromPort | Status
-------------+---------+---------+----------+----------
DC-01      | jamaica | 1280145 |    13000 | running
```

The logs of each poller can be found in `/var/log/harvest/`.

### 8. Add Prometheus Datasource in Grafana

To add a Prometheus datasource in Grafana, follow these steps:

1. Open your web browser and navigate to Grafana, which is running on port 3000. You can access it by entering one of the following URLs in your browser's address bar:
    - If you are accessing Grafana from the same machine where it is installed, use:

        ```bash
        http://localhost:3000
        ```

    - If you are accessing Grafana from a different machine, replace `localhost` with the IP address of the machine where Grafana is running. For example:

        ```bash
        http://<machine-ip>:3000
        ```

2. From the "three lines" button (also known as the hamburger menu) at the top left, navigate to **Connections** and then **Data Sources**.
3. Click on **Add data source**.
4. Select **Prometheus** from the list of available data sources.
5. In the **Prometheus server URL** field, enter `http://localhost:9090`
6. Click on **Save and test**.
7. At the bottom of the page, you should see the message 'Successfully queried the Prometheus API.'
For detailed instructions, please refer to the [Configure Prometheus Data Source documentation](https://grafana.com/docs/grafana/latest/datasources/prometheus/configure-prometheus-data-source/).

### 9. Generate Grafana API Token

To import Grafana dashboards using the `bin/harvest grafana import` command, you need a Grafana API token. Follow these steps to generate it:

1. Log in to Grafana at `http://localhost:3000` or `http://<machine-ip>:3000`.
2. Click on the "three lines" button at the top left of the screen to open the menu.
3. Navigate to **Administration** -> **Users and access** and then select **Service Account**.
4. Click on **Add Service Account**.
5. Enter the display name **Harvest**.
6. Set the role to **Editor**.
7. Click on **Create**. The service account will appear in the dashboard.
8. Navigate back to **Service Account**.
9. Click on **Add service account token** for the Harvest service account.
10. Click on **Generate Token**.
11. Click on **Copy to clipboard and close**.

**IMPORTANT:** This is the only opportunity to save the token. Immediately paste it into a text file and save it. The token will be needed by Harvest later on.

For detailed instructions, please refer to the [Grafana API Keys documentation](https://grafana.com/docs/grafana/latest/administration/api-keys/).

### 10. Import Grafana Dashboards

To import Grafana dashboards, use the following command:

```bash
cd ${HARVEST_INSTALL_PATH}/harvest-${HARVEST_VERSION}-1_linux_amd64
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

### 9. Verify Dashboards in Grafana

After adding the Prometheus datasource, you can verify that your dashboards are correctly displaying data. Follow these steps:

1. Open your web browser and navigate to `http://localhost:3000` or `http://<machine-ip>:3000`
2. Enter your Grafana credentials to log in. The default username is `admin` and the default password is also `admin` (you may have changed this during setup).
3. Click on the "three lines" button (also known as the hamburger menu) in the top left corner of the Grafana interface. From the menu, select **Dashboards**.
4. In the Dashboards section, find and click on the **Volume Dashboard**. This dashboard should have been pre-configured or imported as part of your setup. Once the Volume Dashboard is open, you should see various panels displaying data

### Troubleshooting

If you encounter issues, check the logs in `/var/log/harvest` and refer to the [troubleshooting](https://github.com/NetApp/harvest/wiki/Troubleshooting-Harvest) section on the wiki.
You can also reach out for help on [Discord](https://github.com/NetApp/harvest/blob/main/SUPPORT.md#getting-help) or via email at ng-harvest-files@netapp.com.

### Conclusion

Congratulations! You have successfully set up NetApp Harvest along with Prometheus and Grafana. Enjoy monitoring your systems and feel free to provide feedback.