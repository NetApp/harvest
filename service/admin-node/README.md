# Systemd Integration for HTTP Service Discovery Admin Node

If you want to set up a service file to restart the Harvest HTTP service discovery admin node, you can use the following configuration:

Adjust paths as needed, and don't forget to enable `httpsd` in your `harvest.yml` config.

```
echo '[Unit]
Description="NetApp Harvest HTTPSD"
PartOf=harvest.target

[Service]
User=harvest
Group=harvest
Type=simple
Restart=on-failure 
WorkingDirectory=/opt/harvest
ExecStart=/opt/harvest/bin/harvest --config /opt/harvest/harvest.yml admin start

[Install]
WantedBy=harvest.target' | sudo tee /etc/systemd/system/harvest-admin.service

```
### How to use

`systemctl daemon-reload`

`systemctl start|stop|restart harvest-admin`

`systemctl status harvest-admin`

`systemctl enable harvest-admin` will enable service at machine boot time.

### systemd: Logs

```
journalctl -fu harvest-admin # follow, tail like behavior for http sd
journalctl -u harvest-admin  # show logs for http sd
```
