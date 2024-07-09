Follow this method if your goal is to establish a separate harvest container for each poller defined in `harvest.yml` file. Please note that these containers must be incorporated into your current infrastructure, which might include systems like Prometheus or Grafana.

### Setup harvest.yml

- Create a `harvest.yml` file with your cluster details, below is an example with annotated comments. Modify as needed
  for your scenario.

This config is using the Prometheus
exporter [port_range](../prometheus-exporter.md#port_range)
feature, so you don't have to manage the Prometheus exporter port mappings for each poller.

```
Exporters:
  prometheus1:
    exporter: Prometheus
    addr: 0.0.0.0
    port_range: 2000-2030  # <====== adjust to be greater than equal to the number of monitored clusters

Defaults:
  collectors:
    - Zapi
    - ZapiPerf
    - EMS
  use_insecure_tls: true   # <====== adjust as needed to enable/disable TLS checks 
  exporters:
    - prometheus1

Pollers:
  infinity:                # <====== add your cluster(s) here, they use the exporter defined three lines above
    datacenter: DC-01
    addr: 10.0.1.2
    auth_style: basic_auth
    username: user
    password: 123#abc
  # next cluster ....  
```

### Generate a Docker compose for your Pollers

- Generate a Docker compose file from your `harvest.yml`

```sh
docker run --rm \
  --env UID=$(id -u) --env GID=$(id -g) \
  --entrypoint "bin/harvest" \
  --volume "$(pwd):/opt/temp" \
  --volume "$(pwd)/harvest.yml:/opt/harvest/harvest.yml" \
  ghcr.io/netapp/harvest \
  generate docker \
  --output harvest-compose.yml
```

### Start everything

Bring everything up :rocket:

```
docker compose -f harvest-compose.yml up -d --remove-orphans
```


## Manage pollers

### How do I add a new poller?

1. Add poller to `harvest.yml`
2. Regenerate compose file by running [harvest generate](#generate-a-docker-compose-for-your-pollers)
3. Run [docker compose up](#start-everything), for example,

```bash
docker compose -f harvest-compose.yml up -d --remove-orphans
```

### Stop all containers

```
docker compose-f harvest-compose.yml down
```

### Upgrade Harvest

To upgrade Harvest:

1. Retrieve the most recent version of the Harvest Docker image by executing the following command.This is needed since the new version may contain new templates, dashboards, or other files not included in the Docker
   image.
   ```
   docker pull ghcr.io/netapp/harvest
   ```

2. [Stop all containers](#stop-all-containers)

3. Regenerate your `harvest-compose.yml` file by
   running [harvest generate](#generate-a-docker-compose-for-your-pollers)
   By default, generate will use the `latest` tag. If you want to upgrade to a `nightly` build see the twisty.

    ??? question "I want to upgrade to a nightly build"

        Tell the `generate` cmd to use a different tag like so:

        ```sh
        docker run --rm \
          --env UID=$(id -u) --env GID=$(id -g) \
          --entrypoint "bin/harvest" \
          --volume "$(pwd):/opt/temp" \
          --volume "$(pwd)/harvest.yml:/opt/harvest/harvest.yml" \
          ghcr.io/netapp/harvest:nightly \
          generate docker \
          --image ghcr.io/netapp/harvest:nightly \
          --output harvest-compose.yml
        ```

4. Restart your containers using the following:

   ```
   docker compose -f harvest-compose.yml up -d --remove-orphans
   ```
