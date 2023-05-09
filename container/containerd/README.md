# Containerized Harvest on Mac using containerd 

Harvest runs natively on a Mac already. If you need that, git clone and use `GOOS=darwin make build`. 

This page describes how to run Harvest on your Mac in a containerized environment (Compose, K8, etc.)
The documentation below uses Rancher Desktop, but lima works just as well. Keep in mind, both of them
are considered alpha. They work, but are still undergoing a lot of change.

# Setup

We're going to:
- Install and Start [Rancher Desktop](https://rancherdesktop.io/)
- (Optional) Create Harvest Docker image by following Harvest's existing documentation
- Generate a Compose file following Harvest' existing documentation
- Concatenate the Prometheus/Grafana compose file with the harvest compose file since Rancher doesn't support multiple compose files yet
  - Fixup the concatenated file
- Start containers

Under the hood, Rancher is using [lima](https://github.com/lima-vm/lima). If you want to skip Rancher and use lima directly that works too.

## Install and Start Rancher Desktop

We'll use brew to install Rancher.

```sh
brew install rancher
```

After Rancher Desktop installs, start it `Cmd + Space` type: Rancher and wait for it to start a VM and download images. Once everything is started continue.

## Create Harvest Docker image

You only need to [create a new image](https://github.com/NetApp/harvest/tree/main/container/onePollerPerContainer#building-harvest-docker-image) if you've made changes to Harvest. If you just want to use the latest version of Harvest, skip this step.

These are the same steps outline on [Building Harvest Docker Image](https://github.com/NetApp/harvest/tree/main/container/onePollerPerContainer#building-harvest-docker-image) except we replace `docker build` with `nerdctl` like so:

```sh
nerdctl build -f container/onePollerPerContainer/Dockerfile -t harvest:latest . --no-cache 
```

## Generate a Harvest compose file

Follow the existing documentation to setup your `harvest.yml` file
- https://github.com/NetApp/harvest/tree/main/docker#setup-your-harvestyml

Create your `harvest-compose.yml` file like this:

```sh
docker run --rm \
  --entrypoint "bin/harvest" \
  --volume "$(pwd):/opt/harvest" \
  ghcr.io/netapp/harvest generate docker full \
  --output harvest-compose.yml # --image tag, if you built a new image above
```


## Combine Prometheus/Grafana and Harvest compose file

Currently `nerdctl compose` does not support running with multiple compose files, so we'll concat the `prom-stack.yml` and the `harvest-compose.yml` into one file and then fix it up.

```sh
cat prom-stack.yml harvest-compose.yml > both.yml

# jump to line 45 and remove redundant version and services lines (lines 45, 46, 47 should be removed)
# fix indentation of remaining lines - in vim, starting at line 46
# Shift V
# Shift G
# Shift .
# Esc
# Shift ZZ
```

## Start containers

```sh
nerdctl compose -f both.yml up -d

nerdctl ps -a

CONTAINER ID    IMAGE                               COMMAND                   CREATED               STATUS    PORTS                       NAMES
bd7131291960    docker.io/grafana/grafana:latest    "/run.sh"                 About a minute ago    Up        0.0.0.0:3000->3000/tcp      grafana
f911553a14e2    docker.io/prom/prometheus:latest    "/bin/prometheus --c…"    About a minute ago    Up        0.0.0.0:9090->9090/tcp      prometheus
037a4785bfad    docker.io/library/cbg:latest        "bin/poller --poller…"    About a minute ago    Up        0.0.0.0:15007->15007/tcp    poller_simple7_v21.11.0513
03fb951cfe26    docker.io/library/cbg:latest        "bin/poller --poller…"    59 seconds ago        Up        0.0.0.0:15025->15025/tcp    poller_simple25_v21.11.0513
049d0d65b434    docker.io/library/cbg:latest        "bin/poller --poller…"    About a minute ago    Up        0.0.0.0:16050->16050/tcp    poller_simple49_v21.11.0513
0b77dd1bc0ff    docker.io/library/cbg:latest        "bin/poller --poller…"    About a minute ago    Up        0.0.0.0:16067->16067/tcp    poller_u2_v21.11.0513
1cabd1633c6f    docker.io/library/cbg:latest        "bin/poller --poller…"    About a minute ago    Up        0.0.0.0:15015->15015/tcp    poller_simple15_v21.11.0513
1d78c1bf605f    docker.io/library/cbg:latest        "bin/poller --poller…"    About a minute ago    Up        0.0.0.0:15062->15062/tcp    poller_sandhya_v21.11.0513
286271eabc1d    docker.io/library/cbg:latest        "bin/poller --poller…"    About a minute ago    Up        0.0.0.0:15010->15010/tcp    poller_simple10_v21.11.0513
29710da013d4    docker.io/library/cbg:latest        "bin/poller --poller…"    About a minute ago    Up        0.0.0.0:12990->12990/tcp    poller_simple1_v21.11.0513
321ae28637b6    docker.io/library/cbg:latest        "bin/poller --poller…"    About a minute ago    Up        0.0.0.0:15020->15020/tcp    poller_simple20_v21.11.0513
39c91ae54d68    docker.io/library/cbg:latest        "bin/poller --poller…"    About a minute ago    Up        0.0.0.0:15053->15053/tcp    poller_simple-53_v21.11.0513

nerdctl logs poller_simple1_v21.11.0513
nerdctl compose -f both.yml down

# http://localhost:9090/targets   Prometheus
# http://localhost:3000           Grafana
# http://localhost:15062/metrics  Poller metrics
```

