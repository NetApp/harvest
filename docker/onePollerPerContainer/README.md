# Harvest for Docker

Harvest is container-ready and supports several deployment strategies:
 
- A poller-per-container model that offers a simple, flexible way to run Harvest. This deployment also enables a broad range of orchestrators (Nomad, Mesosphere, Swarm, K8, etc.)

- A Harvest subcommand `harvest generate docker` can be used to create a Docker compose file with a container for each poller in your `harvest.yaml` file as well as the necessary volume and network mounts.

- Docker + Ansible workflow to stand-up Harvest, Grafana, and Prometheus. A nice way to kick the tires and try everything out. See [Monitor all of your ONTAP clusters with Harvest (easy mode)](https://netapp.io/2021/05/21/monitor-all-of-your-ontap-clusters-with-harvest-easy-mode/) for details.
## Deployment

Make sure Docker is installed and `docker --version` meets the [minimum required version](https://github.com/NetApp/harvest#requirements).

Harvest releases are published on NetApp's Container Registry (https://cr.netapp.io) and [Dockerhub](https://hub.docker.com/r/rahulguptajss/harvest).

### Docker Compose Poller per Container

If you want to create a separate container for each poller in your `harvest.yaml` file, download the latest version of Harvest and run 

```
bin/harvest generate docker --output docker-compose.yml

docker-compose up -d --remove-orphans
```

Stop docker containers

```
docker-compose down
```
### Poller per Container

This option is typically used when you want to use an orchestrator or need more control over the volume, network, or configuration of your container(s).

Adjust example command below as needed:

- `PROM_PORT` is set to 12991 and is used to bind port 12991 of the container to TCP port 12991 of the host
- The `harvest.yml` file in the local directory is mapped to `/opt/harvest.yml` in the container. If you want to run with a different Harvest config file, change `$(pwd)/harvest.yml` to point to the desired local file.
   
```
PROM_PORT=12991 ; docker run --rm -it \
-p $PROM_PORT:$PROM_PORT \
--volume $(pwd)/harvest.yml:/opt/harvest/harvest.yml \
 cr.netapp.io/harvest \
 --poller infinity --promPort $PROM_PORT
```

You can also generate a [docker-compose file as described above](#docker-compose-poller-per-container) and use it as a guide without using Docker compose.

# Building Harvest Docker Image

You only need to build a Harvest Docker image if you want a custom image. If instead, you are wanting to run Harvest with Docker see [deployment](#deployment) above.

Build Docker Image

```
docker build -f docker/onePollerPerContainer/Dockerfile -t harvest:latest . --no-cache
or
nerdctl build -f docker/onePollerPerContainer/Dockerfile -t harvest:latest . --no-cache
```
