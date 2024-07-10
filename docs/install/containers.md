## Overview

Harvest is container-ready and supports several deployment options:

- [Stand-up Prometheus, Grafana, and Harvest via Docker Compose](#docker-compose). Choose this if
  you want to hit the ground running. Install, volume and network mounts automatically handled.

- [Stand-up Harvest via Docker Compose](harvest-containers.md) that offers
  more flexibility in configuration. Choose this if you only want to run Harvest containers. Since you pick-and-choose what gets built and how it's deployed, stronger familiarity with containers is
  recommended.

- If you prefer Ansible, David Blackwell created
  an [Ansible script](https://github.com/NetApp/harvest_install) that
  stands up Harvest, Grafana, and Prometheus.

- Want to run Harvest on a Mac
  via [containerd and Rancher Desktop](containerd.md)? We got you
  covered.

- [K8 Deployment](k8.md) via Kompose

## Docker Compose

This is a quick way to install and get started with Harvest. Follow the four steps below to:

- Setup Harvest, Grafana, and Prometheus via Docker Compose
- Harvest dashboards are automatically imported and setup in Grafana with a Prometheus data source
- A separate poller container is created for each monitored cluster
- All pollers are automatically added as Prometheus scrape targets

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
  generate docker full \
  --output harvest-compose.yml
```

By default, the above command uses the harvest configuration file(`harvest.yml`) located in the current directory. If you want to use a harvest config from a different location.
??? question "What if my harvest configuration file is somewhere else or not named harvest.yml"
    Use the following docker run command, updating the `HYML` variable with the absolute path to your `harvest.yml`.

    ```sh
    HYML="/opt/custom_harvest.yml"; \
    docker run --rm \
    --env UID=$(id -u) --env GID=$(id -g) \
    --entrypoint "bin/harvest" \
    --volume "$(pwd):/opt/temp" \
    --volume "${HYML}:${HYML}" \
    ghcr.io/netapp/harvest:latest \
    generate docker full \
    --output harvest-compose.yml \
    --config "${HYML}"
    ```

`generate docker full` does two things:

1. Creates a Docker compose file with a container for each Harvest poller defined in your `harvest.yml`
2. Creates a matching Prometheus service discovery file for each Harvest poller (located
   in `container/prometheus/harvest_targets.yml`). Prometheus uses this file to scrape the Harvest pollers.

### Start everything

Bring everything up :rocket:

```
docker compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans
```

### Note on Docker Logging Configuration

By default, Docker uses the `json-file` logging driver which does not limit the size of the logs. This can cause your system to run out of disk space. Docker provides several options for logging configuration, including different logging drivers and options for log rotation.

Docker recommends using the `local` driver to prevent disk-exhaustion. More details can be found in [Docker logging documentation](https://docs.docker.com/config/containers/logging/configure/)

## Prometheus and Grafana

The `prom-stack.yml` compose file creates a `frontend` and `backend` network. Prometheus and Grafana publish their admin
ports on the front-end network and are routable to the local machine. By default, the Harvest pollers are part of the
backend network and also expose their Prometheus web end-points. 
If you do not want their end-points exposed, add the `--port=false` option to the `generate` sub-command in the [previous step](#generate-a-docker-compose-for-your-pollers).

### Prometheus

After bringing up the `prom-stack.yml` compose file, you can check Prometheus's list of targets
at `http://IP_OF_PROMETHEUS:9090/targets`.

### Customize Prometheus's Retention Time

By default, `prom-stack.yml` is configured for a one year data retention period.
To increase this, for example, to two years, you can create a specific configuration file and make your changes there.
This prevents your custom settings from being overwritten if you regenerate the default `prom-stack.yml` file.
Here's the process:

- Copy the original `prom-stack.yml` to a new file named `prom-stack-prod.yml`:

```sh
cp prom-stack.yml prom-stack-prod.yml
```

- Edit `prom-stack-prod.yml` to include the extended data retention setting by updating the `--storage.tsdb.retention.time=2y` line under the **Prometheus** service's `command` section:

```yaml
command:
  - '--config.file=/etc/prometheus/prometheus.yml'
  - '--storage.tsdb.path=/prometheus'
  - '--storage.tsdb.retention.time=2y'       # Sets data retention to 2 years
  - '--web.console.libraries=/usr/share/prometheus/console_libraries'
  - '--web.console.templates=/usr/share/prometheus/consoles'
```

- Save the changes to `prom-stack-prod.yml`.

Now,
you can start your Docker containers with the updated configuration
that includes the 1-year data retention period by executing the command below:

```sh
docker compose -f prom-stack-prod.yml -f harvest-compose.yml up -d --remove-orphans
```

### Grafana

After bringing up the `prom-stack.yml` compose file, you can access Grafana at `http://IP_OF_GRAFANA:3000`.

You will be prompted to create a new password the first time you log in. Grafana's default credentials are

```
username: admin
password: admin
```

## Manage pollers

### How do I add a new poller?

1. Add poller to `harvest.yml`
2. Regenerate compose file by running [harvest generate](#generate-a-docker-compose-for-your-pollers)
3. Run [docker compose up](#start-everything), for example,

```bash
docker compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans
```

### Stop all containers

```
docker compose -f prom-stack.yml -f harvest-compose.yml down
```

If you encounter the following error message while attempting to stop your Docker containers using `docker-compose down`

```
Error response from daemon: Conflict. The container name "/poller-u2" is already in use by container
```

This error is likely due to running `docker-compose down` from a different directory than where you initially ran `docker-compose up`.

To resolve this issue, make sure to run the `docker-compose down` command from the same directory where you ran `docker-compose up`. This will ensure that Docker can correctly match the container names and IDs with the directory you are working in. 
Alternatively, you can stop the Harvest, Prometheus, and Grafana containers by using the following command:

```
docker ps -aq --filter "name=prometheus" --filter "name=grafana" --filter "name=poller-" | xargs docker stop | xargs docker rm
```

Note: Deleting or stopping Docker containers does not remove the data stored in Docker volumes.

### Upgrade Harvest

> Note: If you want to keep your historical Prometheus data, and you set up your Docker Compose workflow before
> Harvest `22.11`, please read how
> to [migrate your Prometheus volume](https://github.com/NetApp/harvest/blob/main/docs/MigratePrometheusDocker.md)
> before continuing with the upgrade steps below.
> 
> If you need to customize your Prometheus configuration, such as changing the data retention period, please refer to the instructions on [customizing the Prometheus configuration](#customize-prometheuss-retention-time).

To upgrade Harvest:

1. Retrieve the most recent version of the Harvest Docker image by executing the following command.This is needed since the new version may contain new templates, dashboards, or other files not included in the Docker
   image.
   ```
   docker pull ghcr.io/netapp/harvest
   ```

2. [Stop all containers](#stop-all-containers)

3. Regenerate your `harvest-compose.yml` file by
   running [harvest generate](#generate-a-docker-compose-for-your-pollers).
Make sure you don't skip this step. It is essential as it updates local copies of templates and dashboards, which are then mounted to the containers. If this step is skipped, Harvest will run with older templates and dashboards which will likely cause problems.
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
          generate docker full \
          --image ghcr.io/netapp/harvest:nightly \
          --output harvest-compose.yml
        ```

4. Restart your containers using the following:

```
docker compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans
```

## Building Harvest Docker Image

Building a custom Harvest Docker image is only necessary if you require a tailored solution. If your intention is to run Harvest using Docker without any customizations, please refer to the [Overview](#docker-compose) section above.

```sh
source .harvest.env
docker build -f container/onePollerPerContainer/Dockerfile --build-arg GO_VERSION=${GO_VERSION} -t harvest:latest . --no-cache
```