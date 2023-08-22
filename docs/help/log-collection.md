# Harvest Logs Collection Guide

This guide will help you collect logs for Harvest on various platforms. Follow the instructions specific to your platform. If you would like to share the collected logs with the Harvest team, feel free to email them to `ng-harvest-files@netapp.com`.

## RPM, DEB, and Native Installations

For RPM, DEB, and native installations, use the following command to create a compressed tar file containing the logs:

```bash
tar -czvf harvest_logs.tar.gz -C /var/log harvest
```

This command will create a file named `harvest_logs.tar.gz` with the contents of the `/var/log/harvest` directory.

## Docker Container

For Docker containers, first, identify the container ID for your Harvest instance. Then, replace `<container_id>` with the actual container ID in the following command:

```bash
docker logs <container_id> &> harvest_logs.txt && tar -czvf harvest_logs.tar.gz harvest_logs.txt
```

This command will create a file named `harvest_logs.tar.gz` containing the logs from the specified container.

## NABox

For NABox installations, refer to the NABox documentation on collecting logs:

[NABox Troubleshooting - Collecting Logs](https://nabox.org/documentation/troubleshooting/#collecting-logs)