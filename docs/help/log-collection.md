# Harvest Logs Collection Guide

This guide will help you collect Harvest logs on various platforms.
Follow the instructions specific to your platform.
If you would like to share the collected logs with the Harvest team,
please email them to [ng-harvest-files@netapp.com](mailto:ng-harvest-files@netapp.com).

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

For NABox installations,
[ssh](https://nabox.org/documentation/configuration/) into your nabox instance,
and use the following command to create a compressed tar file containing the logs:

```bash
dc logs nabox-api > nabox-api.log; dc logs nabox-harvest2 > nabox-harvest2.log;\
  tar -czf nabox-logs-`date +%Y-%m-%d_%H:%M:%S`.tgz *
```

This command will create a file named `nabox-logs-$date.tgz` containing the nabox-api and Harvest poller logs.

For more information,
see the [NABox documentation on collecting logs](https://nabox.org/documentation/troubleshooting/#collecting-logs)