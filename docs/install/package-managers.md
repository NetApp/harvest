
## Redhat

> Installation and upgrade of the Harvest package may require root or administrator privileges

Download the latest rpm of [Harvest](https://github.com/NetApp/harvest/releases/latest) from the releases
tab and install or upgrade with yum.

```
sudo yum install harvest.XXX.rpm
```

Once the installation has finished, edit the [harvest.yml configuration](../configure-harvest-basic.md) file
located in `/opt/harvest/harvest.yml`

After editing `/opt/harvest/harvest.yml`, manage Harvest with `systemctl start|stop|restart harvest`.

After upgrade, re-import all dashboards (either `bin/harvest grafana import` cli or via the Grafana UI) to
get any new enhancements in dashboards.

> To ensure that you don't run
> into [permission issues](https://github.com/NetApp/harvest/issues/122#issuecomment-856138831), make sure you manage
> Harvest using `systemctl` instead of running the harvest binary directly.

??? quote "Changes install makes"

    * Directories `/var/log/harvest/` and `/var/log/run/` are created
    * A `harvest` user and group are created and the installed files are chowned to harvest
    * Systemd `/etc/systemd/system/harvest.service` file is created and enabled

## Debian

> Installation and upgrade of the Harvest package may require root or administrator privileges

Download the latest deb of [Harvest](https://github.com/NetApp/harvest/releases/latest) from the releases
tab and install or upgrade with apt.

```
sudo apt update
sudo apt install|upgrade ./harvest-<RELEASE>.amd64.deb  
```

Once the installation has finished, edit the [harvest.yml configuration](../configure-harvest-basic.md) file
located in `/opt/harvest/harvest.yml`

After editing `/opt/harvest/harvest.yml`, manage Harvest with `systemctl start|stop|restart harvest`.

After upgrade, re-import all dashboards (either `bin/harvest grafana import` cli or via the Grafana UI) to
get any new enhancements in dashboards.

> To ensure that you don't run
> into [permission issues](https://github.com/NetApp/harvest/issues/122#issuecomment-856138831), make sure you manage
> Harvest using `systemctl` instead of running the harvest binary directly.

??? quote "Changes install makes"

    * Directories `/var/log/harvest/` and `/var/log/run/` are created
    * A `harvest` user and group are created and the installed files are chowned to harvest
    * Systemd `/etc/systemd/system/harvest.service` file is created and enabled
