# Harvest and Containers

Harvest is container-ready and supports several deployment options:

- [Stand-up Prometheus, Grafana, and Harvest via Docker Compose](#harvest-grafana-prometheus-install). Choose this if you want to hit the ground running. Install, volume and network mounts automatically handled.
  
- [Poller-per-container model](https://github.com/NetApp/harvest/tree/main/docker/onePollerPerContainer) that offers more flexibility in configuration. This deployment enables a broad range of orchestrators (Nomad, Mesosphere, Swarm, K8, etc.) since you pick-and-choose what gets built and how it's deployed, stronger familiarity with containers is recommended.

- If you prefer Ansible, David Blackwell created an [Ansible script](https://netapp.io/2021/05/21/monitor-all-of-your-ontap-clusters-with-harvest-easy-mode/) that stands up Harvest, Grafana, and Prometheus.

- Want to run Harvest on a Mac via [containerd and Racher Desktop](https://github.com/NetApp/harvest/tree/main/docker/containerd)? We got you covered.

- [Local K8 Deployment](k8/README.md) via Kompose

# Harvest Grafana Prometheus Install

This is a quick way to install and get started with Harvest. Follow the four steps below to:

- Setup Harvest, Grafana, and Prometheus via Docker Compose
- Harvest dashboards are automatically imported and setup in Grafana with a Prometheus data source
- A separate poller container is created for each monitored cluster
- All pollers are automatically added as Prometheus scrape targets

### Requirements
- Docker: `20.10.0` or higher and compatible Docker Compose

## Download and untar

1. Download the latest version of [Harvest](https://github.com/NetApp/harvest#installation), untar, and cd into the harvest directory

## Setup your harvest.yml

2. Create a `harvest.yml` file with your cluster details, below is an example with annotated comments. Modify as needed for your scenario.

This config is using the Prometheus exporter [port_range](https://github.com/NetApp/harvest/blob/main/cmd/exporters/prometheus/README.md#parameters) feature, so you don't have to manage the Prometheus exporter port mappings for each poller.

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
   
## Generate a Docker compose for your Pollers

3. Generate a Docker compose file from your `harvest.yml`
   
```
bin/harvest generate docker full --output harvest-compose.yml
```

`generate docker full` does two things:
1. Creates a Docker compose file with a container for each Harvest poller defined in your `harvest.yml`
2. Creates a matching Prometheus service discovery file for each Harvest poller (located in `docker/prometheus/harvest_targets.yml`). Prometheus uses this file to scrape the Harvest pollers. 

## Start everything

Bring everything up :rocket:
   
```
docker-compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans
```

## Config and Using

The `prom-stack.yml` compose file creates a `frontend` and `backend` network. Prometheus and Grafana publish their admin ports on the front-end network and are routable to the local machine. By default, the Harvest pollers are part of the backend network and do not expose their Prometheus web end-points. If you want their end-points exposed, pass the `--port` flag to the `generate` sub-command in the [previous step](#generate-a-docker-compose-for-your-pollers), like so:

```
bin/harvest generate docker full --port --output harvest-compose.yml
```

### Grafana

[Grafana](http://localhost:3000/)

Default credentials - you'll be prompted to create a new password the first time you log in

```
username: admin
password: admin
```

### Prometheus

- [List of targets](http://localhost:9090/targets)

### Managing

#### Stop all containers

```
docker-compose -f prom-stack.yml -f harvest-compose.yml down
```

### Upgrade Harvest

To upgrade Harvest, use the `restart` command - a newer image of Harvest will be pulled if available.

```
docker pull rahulguptajss/harvest
docker-compose -f prom-stack.yml -f harvest-compose.yml restart
```

You typically do not need to regenerate your `harvest-compose.yml` file when upgrading Harvest. If that's required, it will be noted in the release notes.