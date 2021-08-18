# Improved systemd Integration

systemd instantiated units for each poller and a target to tie them together. Use wildcards to start|stop|restart

## Poller Service via systemd instantiated services

Create one instantiated service for a poller. Adjust paths as needed

```
echo '[Unit]
Description="Poller instance %I"
PartOf=harvest.target

[Service]
Type=simple
Restart=on-failure 
WorkingDirectory=/path/to/harvest
ExecStart=/path/to/harvest/bin/harvest --config /path/to/harvest/harvest.yml start -f %i' | sudo tee /etc/systemd/system/poller@.service
```

### Harvest Target

Target files are how systemd groups a set of services together. We'll use it as a way to start|stop all pollers as a single unit. Nice on reboot or upgrade.

This is more tedious to setup and change. We'll add a `harvest systemd generate` command that creates this target from a template.

For now you'll need to create this yourself similar to the example below:

```
echo '[Unit]
Description="Harvest"
Wants=poller@unix1.service poller@unix2.service poller@unix3.service poller@unix4.service poller@unix5.service poller@unix6.service poller@unix7.service poller@unix8.service poller@unix9.service

[Install]
WantedBy=multi-user.target' | sudo tee /etc/systemd/system/harvest.target                               
```

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
