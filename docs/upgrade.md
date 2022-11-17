To upgrade Harvest

Stop harvest
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

Follow the [installation](install/overview.md) instructions to download and install Harvest and then
copy your old `harvest.yml` into the new install directory like so:

```
cp /path/to/old/harvest/harvest.yml /path/to/new/harvest.yml
```

After upgrade, re-import all dashboards (either `bin/harvest grafana import` cli or via the Grafana UI) to 
get any new enhancements in dashboards.
