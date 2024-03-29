# Improved systemd Integration

systemd instantiated units for each poller and a target to tie them together. Use wildcards to start|stop|restart

## Poller Service via systemd instantiated services

Create one instantiated service for a poller. Adjust paths as needed

```
echo '[Unit]
Description="NetApp Harvest Poller instance %I"
PartOf=harvest.target
After=network-online.target
Wants=network-online.target

[Service]
User=harvest
Group=harvest
Type=simple
Restart=on-failure 
WorkingDirectory=/opt/harvest
ExecStart=/opt/harvest/bin/harvest --config /opt/harvest/harvest.yml start -f %i

[Install]
WantedBy=harvest.target' | sudo tee /etc/systemd/system/poller@.service
```

> NOTE: By default, the instantiated service will log to `stdout`, which means the 
> log messages will be included in `journald`,and on some operating systems (RHEL8),
> `/var/log/messages`. 
> 
> If you want to skip `journald` and `syslog`, you can tell
> Harvest to log to a file by adding the `--logtofile` command line argument 
> like so: `ExecStart=/opt/harvest/bin/harvest --config /opt/harvest/harvest.yml start --logtofile -f %i`   

### Harvest Target

Target files are how systemd groups a set of services together. 
We'll use it as a way to start|stop all pollers as a single unit. Nice on reboot or upgrade.

Harvest provides a `generate` subcommand to make setting up instantiated instances easier. Use like this:

```
bin/harvest generate systemd
```

If you like the output, redirect like so `bin/harvest generate systemd | sudo tee /etc/systemd/system/harvest.target`

### How to use

`systemctl daemon-reload`

Assuming your `harvest.yml` contains pollers like so:

```
cluster-01:
  collectors:
    - ZAPI
unix2:
  collectors:
    - Unix
...
```

Example commands to Manage Pollers

```
systemctl start poller@cluster-01 poller@unix2 ....

systemctl list-units --type=service "poller*"

systemctl status "poller*"

systemctl stop "poller*"

systemctl start|stop|restart harvest.target

```

### systemd: Logs

```
journalctl -fu poller@cluster-01 # follow, tail like behavior for poller named cluster-01
journalctl -u poller@unix2  # show logs for poller named unix2
```
