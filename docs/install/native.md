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
