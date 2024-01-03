## Installation

Visit the [Releases page](https://github.com/NetApp/harvest/releases) and copy the `tar.gz` link
for the latest release. For example, to download the `23.08.0` release:

```bash
VERSION=23.08.0
wget https://github.com/NetApp/harvest/releases/download/v${VERSION}/harvest-${VERSION}-1_linux_amd64.tar.gz
tar -xvf harvest-${VERSION}-1_linux_amd64.tar.gz
cd harvest-${VERSION}-1_linux_amd64

# Run Harvest with the default unix localhost collector
bin/harvest start
```

??? info "With curl"

    If you don't have `wget` installed, you can use `curl` like so:

    ```
    curl -L -O https://github.com/NetApp/harvest/releases/download/v22.08.0/harvest-22.08.0-1_linux_amd64.tar.gz
    ```

## Upgrade

Stop Harvest:
```
cd <existing harvest directory>
bin/harvest stop
```

Verify that all pollers have stopped:
```
bin/harvest status
or
pgrep --full '\-\-poller'  # should return nothing if all pollers are stopped
```

Download the latest release and extract it to a new directory. For example, to upgrade to the 23.11.0 release:

```bash
VERSION=23.11.0
wget https://github.com/NetApp/harvest/releases/download/v${VERSION}/harvest-${VERSION}-1_linux_amd64.tar.gz
tar -xvf harvest-${VERSION}-1_linux_amd64.tar.gz
cd harvest-${VERSION}-1_linux_amd64
```

Copy your old `harvest.yml` into the new install directory:
```
cp /path/to/old/harvest/harvest.yml /path/to/new/harvest/harvest.yml
```

After upgrade, re-import all dashboards (either `bin/harvest grafana import` cli or via the Grafana UI) to
get any new enhancements in dashboards. For more details, see the [dashboards documentation](../dashboards.md).

It's best to run Harvest as a non-root user. Make sure the user running Harvest can write to `/var/log/harvest/` or tell Harvest to write the logs somewhere else with the `HARVEST_LOGS` environment variable.

If something goes wrong, examine the logs files in `/var/log/harvest`, check out
the [troubleshooting](https://github.com/NetApp/harvest/wiki/Troubleshooting-Harvest) section on the wiki and jump
onto [Discord](https://github.com/NetApp/harvest/blob/main/SUPPORT.md#getting-help) and ask for help.