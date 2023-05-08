## Overview

Harvest is container-ready and supports several deployment options:

- [Stand-up Prometheus, Grafana, and Harvest via Docker Compose](#docker-compose). Choose this if
  you want to hit the ground running. Install, volume and network mounts automatically handled.

- [Poller-per-container model](https://github.com/NetApp/harvest/tree/main/container/onePollerPerContainer) that offers
  more flexibility in configuration. This deployment enables a broad range of orchestrators (Nomad, Mesosphere, Swarm,
  K8, etc.) since you pick-and-choose what gets built and how it's deployed, stronger familiarity with containers is
  recommended.

- If you prefer Ansible, David Blackwell created
  an [Ansible script](https://netapp.io/2021/05/21/monitor-all-of-your-ontap-clusters-with-harvest-easy-mode/) that
  stands up Harvest, Grafana, and Prometheus.

- Want to run Harvest on a Mac
  via [containerd and Racher Desktop](https://github.com/NetApp/harvest/tree/main/container/containerd)? We got you
  covered.

- [K8 Deployment](https://github.com/NetApp/harvest/blob/main/container/k8/README.md) via Kompose

## Docker Compose

This is a quick way to install and get started with Harvest. Follow the four steps below to:

- Setup Harvest, Grafana, and Prometheus via Docker Compose
- Harvest dashboards are automatically imported and setup in Grafana with a Prometheus data source
- A separate poller container is created for each monitored cluster
- All pollers are automatically added as Prometheus scrape targets

### Download and untar

- Download the latest version of [Harvest](https://netapp.github.io/harvest/latest/install/native/), untar, and
   cd into the harvest directory.

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

```
docker run --rm \
  --entrypoint "bin/harvest" \
  --volume "$(pwd):/opt/harvest" \
  ghcr.io/netapp/harvest generate docker full \
  --output harvest-compose.yml
```

`generate docker full` does two things:

1. Creates a Docker compose file with a container for each Harvest poller defined in your `harvest.yml`
2. Creates a matching Prometheus service discovery file for each Harvest poller (located
   in `container/prometheus/harvest_targets.yml`). Prometheus uses this file to scrape the Harvest pollers.

### Start everything

Bring everything up :rocket:

```
docker-compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans
```

## Prometheus and Grafana

The `prom-stack.yml` compose file creates a `frontend` and `backend` network. Prometheus and Grafana publish their admin
ports on the front-end network and are routable to the local machine. By default, the Harvest pollers are part of the
backend network and also expose their Prometheus web end-points. 
If you do not want their end-points exposed, remove the `--port` option from the `generate` sub-command in the [previous step](#generate-a-docker-compose-for-your-pollers).

### Prometheus

After bringing up the `prom-stack.yml` compose file, you can check Prometheus's list of targets
at `http://IP_OF_PROMETHEUS:9090/targets`.

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
docker-compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans
```

### Stop all containers

```
docker-compose -f prom-stack.yml -f harvest-compose.yml down
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

To upgrade Harvest:

1. Download the latest `tar.gz` or packaged version and install it.
   This is needed since the new version may contain new templates, dashboards, or other files not included in the Docker
   image.

2. [Stop all containers](#stop-all-containers)

3. Copy your existing `harvest.yml` into the new Harvest directory created in step #1.

4. Regenerate your `harvest-compose.yml` file by
   running [harvest generate](#generate-a-docker-compose-for-your-pollers)
   By default, generate will use the `latest` tag. If you want to upgrade to a `nightly` build see the twisty.

    ??? question "I want to upgrade to a nightly build"
    
        Tell the `generate` cmd to use a different tag like so:

        `docker run --rm --entrypoint "bin/harvest" --volume "$(pwd):/opt/harvest" ghcr.io/netapp/harvest:nightly generate docker full --output harvest-compose.yml`

5. Pull new images and restart your containers like so:

```
docker pull ghcr.io/netapp/harvest   # or if using Docker Hub: docker pull rahulguptajss/harvest
docker-compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans
```
