# Migrate Prometheus Docker Volume

If you want to keep your historical Prometheus data, and you generated your `harvest-compose.yml` file via `bin/harvest generate` before Harvest `22.11`, please follow the steps below to migrate your historical Prometheus data.

This is not required if you generated your `harvest-compose.yml` file via `bin/harvest generate` at Harvest release `22.11` or after.

Outline of steps:
1. [Stop Prometheus container so data acquiesces](#stop-prometheus-container)
2. [Find historical Prometheus volume and create new Prometheus data volume](#find-the-name-of-the-prometheus-volume-that-has-the-historical-data)
3. [Create a new Prometheus volume that Harvest 22.11 and after will use](#create-new-prometheus-volume)
4. [Copy the historical Prometheus data from the old volume to the new one](#copy-the-historical-prometheus-data)
5. [Optionally remove the historical Prometheus volume](#optionally-remove-historical-prometheus-data)

## Stop Prometheus container

It's safe to run the `stop` and `rm` commands below regardless if Prometheus is running or not since removing the container does not touch the historical data stored in the volume.

Stop all containers named Prometheus and remove them.

```bash
docker stop (docker ps -fname=prometheus -q) && docker rm (docker ps -a -fname=prometheus -q)
```

Docker may complain if the container is not running, like so. You can ignore this.

<details>
    <summary>Ignorable output when container is not running (click me)</summary>

```bash
"docker stop" requires at least 1 argument.
See 'docker stop --help'.

Usage:  docker stop [OPTIONS] CONTAINER [CONTAINER...]

Stop one or more running containers
```

</details>


## Find the name of the Prometheus volume that has the historical data

```bash
docker volume ls -f name=prometheus -q
```

Output should look like this:
```bash
harvest-22080-1_linux_amd64_prometheus_data  # historical Prometheus data here
harvest_prometheus_data                      # it is fine if this line is missing
```

We want to copy the historical data from `harvest-22080-1_linux_amd64_prometheus_data` to `harvest_prometheus_data`

> If `harvest_prometheus_data` already exists, you need to decide if you want to move that volume's data to a different volume or remove it. If you want to remove the volume, run `docker volume rm harvest_prometheus_data`. If you want to move the data, adjust the command below to first copy `harvest_prometheus_data` to a different volume and then remove it.

## Create new Prometheus volume

We're going to create a new mount named, `harvest_prometheus_data` by executing:

```bash
docker volume create --name harvest_prometheus_data
```

## Copy the historical Prometheus data

We will copy the historical Prometheus data from the old volume to the new one by
mounting both volumes and copying data between them.

```bash
# replace  `HISTORICAL_VOLUME` with the name of the Prometheus volume that contains you historical data found in step 2.
docker run --rm -it -v $HISTORICAL_VOLUME:/from -v harvest_prometheus_data:/to alpine ash -c "cd /from ; cp -av . /to"
```

Output will look something like this:

```bash
'./wal' -> '/to/./wal'
'./wal/00000000' -> '/to/./wal/00000000'
'./chunks_head' -> '/to/./chunks_head'
...
```

## Optionally remove historical Prometheus data

Before removing the historical data, [start your compose stack](https://github.com/NetApp/harvest/tree/main/docker#start-everything) and make sure everything works.

Once you're satisfied that you can destroy the old data, remove it like so.

```bash
# replace `HISTORICAL_VOLUME` with the name of the Prometheus volume that contains your historical data found in step 2.
docker volume rm $HISTORICAL_VOLUME
```

## Reference
- [Rename Docker Volume](https://github.com/moby/moby/issues/31154#issuecomment-360531460)