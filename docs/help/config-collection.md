# Harvest Config Collection Guide

This guide is designed to help you validate your Harvest configuration on various platforms. The commands provided in this guide will generate output that has been redacted to remove any personally identifiable information (PII). This makes it safe for you to share the output with the Harvest team if needed. Follow the instructions specific to your platform. If you wish to share it with the Harvest team, please email them at [ng-harvest-files@netapp.com](mailto:ng-harvest-files@netapp.com).


## RPM, DEB, and Native Installations

To print a redacted version of your Harvest configuration to the console, use the following command:

```bash
export CONFIG_FILE_NAME=harvest.yml
bin/harvest doctor --print --config $CONFIG_FILE_NAME
```

## Docker Container

For Docker containers, use the following command to print a redacted version of your Harvest configuration to the console:

```bash
export CONFIG_FILE_NAME=harvest.yml
docker run --rm --entrypoint "bin/harvest" --volume "$(pwd)/$CONFIG_FILE_NAME:/opt/harvest/harvest.yml" ghcr.io/netapp/harvest doctor --print
```

## NABox

If you're using NABox, you'll need to [ssh](https://nabox.org/documentation/configuration/) as root into your NABox instance. Then, use the following command to print a redacted version of your Harvest configuration to the console:

```bash
dc exec -w /conf nabox-harvest2 /netapp-harvest/bin/harvest doctor --print
```

If your configuration file name is different from the default `harvest.yml`, remember to set the `CONFIG_FILE_NAME` environment variable to match your specific configuration file name for relevant platform.
