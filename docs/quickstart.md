Welcome to the NetApp Harvest Getting Started Guide. This tutorial will guide you through the steps required to deploy an instance of NetApp Harvest, Prometheus, and Grafana on a Linux platform to monitor an ONTAP cluster.

This tutorial uses systemd to manage Harvest, Prometheus, and Grafana. If you would rather, run the processes directly, feel free to ignore the sections of the tutorial that setup systemd service files.

### 1. Set Installation Path

First, set the installation path as an environment variable. For example, we'll use `/opt/netapp/harvest`.

```bash
HARVEST_INSTALL_PATH=/opt/netapp/harvest
mkdir -p ${HARVEST_INSTALL_PATH}
```

### 2. Install Harvest

Harvest is distributed as a container, native tarball, and RPM and Debs. Pick the one that works best for you.
More details can be found in the [installation](https://netapp.github.io/harvest/latest/install/overview/) documentation.

For this guide, we'll use the tarball package as an example.

Visit the releases page and take note of the latest release. Update the `HARVEST_VERSION` environment variable with the latest release in the script below. For example, to download the `24.05.2` release you would use `HARVEST_VERSION=24.05.2`

After updating the `HARVEST_VERSION` environment variable run the bash script to download Harvest and untar it into your `HARVEST_INSTALL_PATH` directory.

```bash
HARVEST_VERSION=24.05.2
cd ${HARVEST_INSTALL_PATH}
wget https://github.com/NetApp/harvest/releases/download/v${HARVEST_VERSION}/harvest-${HARVEST_VERSION}-1_linux_amd64.tar.gz
tar -xvf harvest-${HARVEST_VERSION}-1_linux_amd64.tar.gz
```

### 3. Install Prometheus

To install Prometheus, follow these steps. For more details see [Prometheus installation](https://prometheus.io/docs/prometheus/latest/installation/).

```bash
PROMETHEUS_VERSION=2.49.1
cd ${HARVEST_INSTALL_PATH}
wget https://github.com/prometheus/prometheus/releases/download/v${PROMETHEUS_VERSION}/prometheus-${PROMETHEUS_VERSION}.linux-amd64.tar.gz
tar -xvf prometheus-${PROMETHEUS_VERSION}.linux-amd64.tar.gz
mv prometheus-${PROMETHEUS_VERSION}.linux-amd64 prometheus-${PROMETHEUS_VERSION}
```
If you want to manage Prometheus with `systemd`, you can create a service file for Prometheus like so. This step is optional.
A service file will attempt to restart Prometheus automatically when the machine is restarted.

Create a service file for Prometheus:

```bash
cat << EOF | sudo tee /etc/systemd/system/prometheus.service
[Unit]
Description=Prometheus Server
Documentation=https://prometheus.io/docs/introduction/overview/
After=network-online.target

[Service]
Environment="HARVEST_INSTALL_PATH=$HARVEST_INSTALL_PATH"
Environment="PROMETHEUS_VERSION=$PROMETHEUS_VERSION"
User=root
Restart=on-failure
ExecStart=/bin/bash -c '\${HARVEST_INSTALL_PATH}/prometheus-\${PROMETHEUS_VERSION}/prometheus --config.file=\${HARVEST_INSTALL_PATH}/prometheus-\${PROMETHEUS_VERSION}/prometheus.yml'

[Install]
WantedBy=multi-user.target
EOF
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

<details>
  <summary>Alternative: Start Prometheus Directly</summary>

  If you would rather start Prometheus directly and kick the tires before creating a service file, you can run the following command to start Prometheus in the background:

  ```bash
  nohup ${HARVEST_INSTALL_PATH}/prometheus-${PROMETHEUS_VERSION}/prometheus --config.file=${HARVEST_INSTALL_PATH}/prometheus-${PROMETHEUS_VERSION}/prometheus.yml > prometheus.log 2>&1 &
  ```

  This command uses <code>nohup</code> to run Prometheus in the background and redirects the output to <code>prometheus.log</code>.

</details>

### 4. Install Grafana

To install Grafana, follow these steps:

```bash
GRAFANA_VERSION=10.4.5
cd ${HARVEST_INSTALL_PATH}
wget https://dl.grafana.com/oss/release/grafana-${GRAFANA_VERSION}.linux-amd64.tar.gz
tar -xvf grafana-${GRAFANA_VERSION}.linux-amd64.tar.gz
```

If you want to manage Grafana with `systemd`, you can create a service file for Grafana like so. This step is optional.
A service file will attempt to restart Grafana automatically when the machine is restarted.

Create a service file for Grafana:

```bash
cat << EOF | sudo tee /etc/systemd/system/grafana.service
[Unit]
Description=Grafana Server
Documentation=https://grafana.com/docs/grafana/latest/setup-grafana/installation/
After=network-online.target

[Service]
Environment="HARVEST_INSTALL_PATH=$HARVEST_INSTALL_PATH"
Environment="GRAFANA_VERSION=$GRAFANA_VERSION"
User=root
Restart=on-failure
ExecStart=/bin/bash -c '\${HARVEST_INSTALL_PATH}/grafana-v\${GRAFANA_VERSION}/bin/grafana-server --config=\${HARVEST_INSTALL_PATH}/grafana-v\${GRAFANA_VERSION}/conf/defaults.ini --homepath=\${HARVEST_INSTALL_PATH}/grafana-v\${GRAFANA_VERSION}'

[Install]
WantedBy=multi-user.target
EOF
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

<details>
  <summary>Alternative: Start Grafana Directly</summary>

  If you would rather start Grafana directly and kick the tires before creating a service file, you can run the following command to start Grafana in the background:

  ```bash
  nohup ${HARVEST_INSTALL_PATH}/grafana-v${GRAFANA_VERSION}/bin/grafana-server --config=${HARVEST_INSTALL_PATH}/grafana-v${GRAFANA_VERSION}/conf/defaults.ini --homepath=${HARVEST_INSTALL_PATH}/grafana-v${GRAFANA_VERSION} > grafana.log 2>&1 &
  ```

  This command uses <code>nohup</code> to run Grafana in the background and redirects the output to <code>grafana.log</code>.

</details>


### 5. Configuration File

By default, Harvest loads its configuration information from the `./harvest.yml` file.
If you would rather use a different file, use the `--config` command line argument flag to specify the path to your config file.

To start collecting metrics, you need to define at least one `poller` and one `exporter` in your configuration file.
This is useful if you want to monitor resource usage by Harvest and serves as a good example. Feel free to delete it if you want.

The next step is to add pollers for your ONTAP clusters in the [Pollers](configure-harvest-basic.md#pollers) section of the Harvest configuration file, `harvest.yml`.

Edit the Harvest configuration file:

```sh
cd ${HARVEST_INSTALL_PATH}/harvest-${HARVEST_VERSION}-1_linux_amd64
vi harvest.yml
```

Copy and paste the following YAML configuration into your editor and update the `$cluster-management-ip`, `$username`, and `$password` sections to match your ONTAP system.

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
    addr: $cluster-management-ip
    auth_style: basic_auth
    username: $username
    password: $password
    exporters:
      - prometheus1
```

**Note:** The ONTAP user specified in this configuration must have the appropriate permissions as outlined in the [Prepare cDot Clusters](prepare-cdot-clusters.md) documentation.

### 6. Edit Prometheus Config File

Edit the Prometheus configuration file:

```sh
cd ${HARVEST_INSTALL_PATH}/prometheus-${PROMETHEUS_VERSION}
vi prometheus.yml
```

Add the following under the `scrape_configs` section. The targets you are adding should match the range of ports you specified in your `harvest.yml` file (in the example above, we use the port_range `13000-13100`).

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

To start the Harvest pollers, follow these steps. For more details see [Harvest service](https://github.com/NetApp/harvest/blob/main/service/contrib/README.md).

Create a systemd service file for Harvest pollers:

```bash
cat << EOF | sudo tee /etc/systemd/system/poller@.service
[Unit]
Description="NetApp Harvest Poller instance %I"
PartOf=harvest.target
After=network-online.target
Wants=network-online.target

[Service]
Environment="HARVEST_INSTALL_PATH=$HARVEST_INSTALL_PATH"
Environment="HARVEST_VERSION=$HARVEST_VERSION"
User=harvest
Group=harvest
Type=simple
Restart=on-failure
ExecStart=/bin/bash -c 'cd \${HARVEST_INSTALL_PATH}/harvest-\${HARVEST_VERSION}-1_linux_amd64 && \${HARVEST_INSTALL_PATH}/harvest-\${HARVEST_VERSION}-1_linux_amd64/bin/harvest --config \${HARVEST_INSTALL_PATH}/harvest-\${HARVEST_VERSION}-1_linux_amd64/harvest.yml start -f %i'

[Install]
WantedBy=harvest.target
EOF
```

Create a target file for Harvest:

```bash
cd ${HARVEST_INSTALL_PATH}/harvest-${HARVEST_VERSION}-1_linux_amd64
bin/harvest generate systemd | sudo tee /etc/systemd/system/harvest.target
```

Reload the systemd configuration and start Harvest:

```bash
sudo systemctl daemon-reload
sudo systemctl enable harvest.target
sudo systemctl start harvest.target
```

Verify that the pollers have started successfully by checking their status:

```bash
systemctl status "poller*"
```

<details>
  <summary>Alternative: Start Harvest Directly</summary>

  If you would rather start Harvest directly and kick the tires before creating a service file, you can run the following command to start Harvest:

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
  -----------+---------+---------+----------+----------
  DC-01      | jamaica | 1280145 |    13000 | running
  ```

    <p>The <a href="https://netapp.github.io/harvest/latest/help/log-collection/">logs</a> of each poller can be found in <code>/var/log/harvest/</code>.</p>
</details>

### 8. Add Prometheus Datasource in Grafana

To add a Prometheus datasource in Grafana, follow these steps:

1. Open your web browser and navigate to Grafana ([http://localhost:3000](http://localhost:3000)). When prompted for credentials, use Grafana defaults admin/admin. You should change the default credentials once you log in.
2. Navigate to the data sources section by visiting [http://localhost:3000/connections/datasources](http://localhost:3000/connections/datasources) or by clicking the hamburger menu (three horizontal lines) at the top-left of the page and navigate to **Connections** and then **Data Sources**.
3. Click on **Add data source**.
4. Select **Prometheus** from the list of available data sources.
5. In the **Prometheus server URL** field, enter ([http://localhost:9090](http://localhost:9090)).
6. Click on **Save and test**.
7. At the bottom of the page, you should see the message 'Successfully queried the Prometheus API.'
For detailed instructions, please refer to the [configure Prometheus Data Source documentation](https://grafana.com/docs/grafana/latest/datasources/prometheus/configure-prometheus-data-source/).

### 9. Generate Grafana API Token

To import Grafana dashboards using the `bin/harvest grafana import` command, you need a Grafana API token. Follow these steps to generate it:

1. Open your web browser and navigate to Grafana ([http://localhost:3000](http://localhost:3000)). Enter your Grafana credentials to log in. The default username and password are `admin`.
2. Click the hamburger menu (three horizontal lines) at the top-left of the page and Navigate to **Administration** -> **Users and access** and then select **Service Account**.
3. Click on **Add Service Account**.
4. Enter the display name **Harvest**.
5. Set the role to **Editor**.
6. Click on **Create**. The service account will appear in the dashboard.
7. Navigate back to **Service Account**.
8. Click on **Add service account token** for the Harvest service account.
9. Click on **Generate Token**.
10. Click on **Copy to clipboard and close**.

**IMPORTANT:** This is the only opportunity to save the token. Immediately paste it into a text file and save it. The token will be needed by Harvest later on.

For detailed instructions, please refer to the [Grafana API Keys documentation](https://grafana.com/docs/grafana/latest/administration/api-keys/).

### 10. Import Grafana Dashboards

To import Grafana dashboards, use the following command:

```bash
cd ${HARVEST_INSTALL_PATH}/harvest-${HARVEST_VERSION}-1_linux_amd64
bin/harvest grafana import --token YOUR_TOKEN_HERE
```

Replace `YOUR_TOKEN_HERE` with the token obtained in step 10.

You will be prompted to save your API key (token) for later use. Press `n` to not save the token in your harvest.yml file.

After a few seconds, all the dashboards will be imported into Grafana.

### 9. Verify Dashboards in Grafana

After adding the Prometheus datasource, you can verify that your dashboards are correctly displaying data. Follow these steps:

1. Open your web browser and navigate to Grafana ([http://localhost:3000](http://localhost:3000)). Enter your Grafana credentials to log in. The default username and password are `admin`.
2. Click on the "three lines" button (also known as the hamburger menu) in the top left corner of the Grafana interface. From the menu, select **Dashboards**.
3. Open the [Volume](http://localhost:3000/d/cdot-volume/ontap3a-volume?orgId=1) dashboard. Once the dashboard opens, you should see volume data displayed.

### Troubleshooting

If you encounter issues, check the logs in `/var/log/harvest` and refer to the [troubleshooting](https://github.com/NetApp/harvest/wiki/Troubleshooting-Harvest) section on the wiki.
You can also reach out for help on [Discord](https://github.com/NetApp/harvest/blob/main/SUPPORT.md#getting-help) or via email at ng-harvest-files@netapp.com.

### Conclusion

🎊 Congratulations! You have successfully set up NetApp Harvest, Prometheus, and Grafana.
Enjoy monitoring your systems and feel free to reach out on [Discord, GitHub, or email](https://github.com/NetApp/harvest/blob/main/SUPPORT.md#getting-help).