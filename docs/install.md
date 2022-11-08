Get up and running with Harvest on your preferred platform.
We provide pre-compiled binaries for Linux, RPMs, Debs, as well 
as prebuilt container images for both [Nightly](https://github.com/NetApp/harvest/releases/tag/nightly) 
and stable [releases](https://github.com/NetApp/harvest/releases).

## Native

Visit the [Releases page](https://github.com/NetApp/harvest/releases) and copy the `tar.gz` link 
for the latest release. For example, to download the `v22.08.0` release:
```
wget https://github.com/NetApp/harvest/releases/download/v22.08.0/harvest-22.08.0-1_linux_amd64.tar.gz
tar -xvf harvest-22.08.0-1_linux_amd64.tar.gz
cd harvest-22.08.0-1_linux_amd64

# Run Harvest with the default unix localhost collector
bin/harvest start
```

??? info "With curl"

    If you don't have `wget` installed, you can use `curl` like so:
    ```
    curl -L -O https://github.com/NetApp/harvest/releases/download/v22.08.0/harvest-22.08.0-1_linux_amd64.tar.gz
    ```

It's best to run Harvest as a non-root user. Make sure the user running Harvest can write to `/var/log/harvest/` or tell Harvest to write the logs somewhere else with the `HARVEST_LOGS` environment variable.

If something goes wrong, examine the logs files in `/var/log/harvest`, check out
the [troubleshooting](https://github.com/NetApp/harvest/wiki/Troubleshooting-Harvest) section on the wiki and jump
onto [Discord](https://github.com/NetApp/harvest/blob/main/SUPPORT.md#getting-help) and ask for help.

## Containers

See [Harvest and containers](https://github.com/NetApp/harvest/blob/main/docker/README.md).

## Package managers

### Redhat

> Installation and upgrade of the Harvest package may require root or administrator privileges

Download the latest rpm of [Harvest](https://github.com/NetApp/harvest/releases/latest) from the releases 
tab and install or upgrade with yum.

```
sudo yum install harvest.XXX.rpm
```

Once the installation has finished, edit the [harvest.yml configuration](configure-harvest-basic.md) file 
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

### Debian

> Installation and upgrade of the Harvest package may require root or administrator privileges

Download the latest deb of [Harvest](https://github.com/NetApp/harvest/releases/latest) from the releases 
tab and install or upgrade with apt.

```
sudo apt update
sudo apt install|upgrade ./harvest-<RELEASE>.amd64.deb  
```

Once the installation has finished, edit the [harvest.yml configuration](configure-harvest-basic.md) file 
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

## Nabox

Instructions on how to install Harvest via [NAbox](https://nabox.org/documentation/installation/).

## Source

To build Harvest from source code, first make sure you have a working Go environment 
with [version 1.19 or greater installed](https://golang.org/doc/install).

Clone the repo and build everything.

```
git clone https://github.com/NetApp/harvest.git
cd harvest
make build
bin/harvest version
```

If you're building on a Mac use `GOOS=darwin make build`

Checkout the `Makefile` for other targets of interest.
