# Harvest Config Collection Guide

This guide is designed to help you validate your Harvest configuration (`harvest.yml`) on various platforms.
The commands in this guide will generate redacted output that personally identifiable information (PII) removed.
This makes it safe for you to share the output.
Follow the instructions specific to your platform.
If you wish to share it with the Harvest team,
please email them at [ng-harvest-files@netapp.com](mailto:ng-harvest-files@netapp.com).

## RPM, DEB, and Native Installations

To print a redacted version of your Harvest configuration to the console, use the following command:

```bash
cd /opt/harvest
export CONFIG_FILE_NAME=harvest.yml
bin/harvest doctor --print --config $CONFIG_FILE_NAME
```

## Docker Container

For Docker containers, use the following command to print a redacted version of your Harvest configuration to the console:

```bash
cd to/where/your/harvest.yml/is
export CONFIG_FILE_NAME=harvest.yml
docker run --rm --entrypoint "bin/harvest" --volume "$(pwd)/$CONFIG_FILE_NAME:/opt/harvest/harvest.yml" ghcr.io/netapp/harvest doctor --print
```

## NABox3

If you're using NABox3, you'll need to [ssh](https://nabox.org/documentation/configuration/) into your NABox instance.
Then, use the following command to print a redacted version of your Harvest configuration to the console:

```bash
dc exec -w /conf nabox-harvest2 /netapp-harvest/bin/harvest doctor --print
```

## NABox4

If you're using NABox4, you'll need to [ssh](https://nabox.org/documentation/configuration/) into your NABox instance.
Then, use the following command to print a redacted version of your Harvest configuration to the console:

```bash
dc exec -w /harvest -e HARVEST_CONF=/harvest-conf harvest /harvest/bin/harvest doctor --print
```

If your configuration file name is different from the default `harvest.yml`,
remember to change the `CONFIG_FILE_NAME` environment variable to match your file name.
