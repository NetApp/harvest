# Containerized Harvest on Linux using Rootless Podman

RHEL 8 ships with [Podman](https://github.com/containers/podman) instead of Docker. There are two ways to run containers with Podman: rootless or with root. Both setups are outlined below. The Podman ecosystem is changing rapidly so the shelf life of these instructions may be short. Make sure you have at least the same [versions](#versions) of the tools listed below. 

If you don't want to bother with Podman, you can also install Docker on RHEL 8 and use it to run [Harvest per normal](https://github.com/NetApp/harvest/tree/main/docker).

## Setup

```bash
sudo yum remove docker-ce
sudo yum module enable -y container-tools:rhel8
sudo yum module install -y container-tools:rhel8
sudo yum install podman podman-docker podman-plugins
```

We also need to install Docker Compose since Podman uses it for compose workflows. Install `docker-compose` like this:
```bash
VERSION=1.29.2
sudo curl -L "https://github.com/docker/compose/releases/download/$VERSION/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
sudo ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose
```

## Containerized Harvest on Linux using Rootful Podman

Make sure you're able to curl the endpoint.

```bash
sudo curl -H "Content-Type: application/json" --unix-socket /var/run/docker.sock http://localhost/_ping
```

If the `sudo curl` does not print `OK⏎` troubleshoot before continuing.

Proceed to [Running Harvest](#running-harvest)

## Containerized Harvest on Linux using Rootless Podman

To run Podman rootless, we'll create a non-root user named: `harvest` to run Harvest.

```bash
# as root or sudo
usermod --append --groups wheel harvest
```

Login with the harvest user, setup the podman.socket, and make sure the curl below works. `su` or `sudo` aren't sufficient, you need to `ssh` into the machine as the harvest user or use `machinectl login`. See [sudo-rootless-podman](https://www.redhat.com/sysadmin/sudo-rootless-podman) for details.

```bash
# these must be run as the harvest user
systemctl --user enable podman.socket
systemctl --user start podman.socket
systemctl --user status podman.socket
export DOCKER_HOST=unix:///run/user/$UID/podman/podman.sock

sudo curl -H "Content-Type: application/json" --unix-socket /var/run/docker.sock http://localhost/_ping
```

If the `sudo curl` does not print `OK⏎` troubleshoot before continuing.

Run podman info and make sure `runRoot` points to `/run/user/$UID/containers` (see below). If it doesn't, you'll probably run into problems when restarting the machine. See [errors after rebooting](#errors-after-rebooting).

```bash
podman info | grep runRoot
  runRoot: /run/user/1001/containers
```

## Running Harvest

By default, Cockpit runs on port 9090, same as Prometheus. We'll change Prometheus's host port to 9091 so we can run both Cockpit and Prometheus.

Edit `prom-stack.yml` and change the prometheus ports section from `9090:9090` to `9091:9090`

```yaml
services:
    prometheus:
        ports:
            - 9091:9090   # <======== change this line
```

With these changes, the [standard Harvest compose instructions](https://github.com/NetApp/harvest/tree/main/docker) can be followed as normal now. In summary,
1. Add the clusters, exporters, etc. to your `harvest.yml` file
2. Generate a compose file from your `harvest.yml` by running `bin/harvest generate docker full --port --output harvest-compose.yml`
3. Bring everything up with `docker-compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans`

After starting the containers, you can view them with `podman ps -a` or using Cockpit `https://host-ip:9090/podman`.

```bash
podman ps -a
CONTAINER ID  IMAGE                                   COMMAND               CREATED        STATUS            PORTS                     NAMES
45fd00307d0a  docker.io/rahulguptajss/harvest:latest  --poller unix --p...  5 seconds ago  Up 5 seconds ago  0.0.0.0:12990->12990/tcp  poller_unix_v21.11.0
d40585bb903c  localhost/prom/prometheus:latest        --config.file=/et...  5 seconds ago  Up 5 seconds ago  0.0.0.0:9091->9090/tcp    prometheus
17a2784bc282  localhost/grafana/grafana:latest                              4 seconds ago  Up 5 seconds ago  0.0.0.0:3000->3000/tcp    grafana
```

## Troubleshooting

Check [Podman's troubleshooting docs](https://github.com/containers/podman/blob/main/troubleshooting.md)

### Nothing works

Make sure the `DOCKER_HOST` [env variable is set](#setup) and that this curl works.
```bash
sudo curl -H "Content-Type: application/json" --unix-socket /var/run/docker.sock http://localhost/_ping
```

Make sure your containers can talk to each other.

```bash
ping prometheus
PING prometheus (10.88.2.3): 56 data bytes
64 bytes from 10.88.2.3: seq=0 ttl=42 time=0.059 ms
64 bytes from 10.88.2.3: seq=1 ttl=42 time=0.065 ms
```
### Errors after rebooting
After restarting the machine, I see errors like these when running `podman ps`.

```
podman ps -a
ERRO[0000] error joining network namespace for container 424df6c: error retrieving network namespace at /run/user/1001/netns/cni-5fb97adc-b6ef-17e8-565b-0481b311ba09: failed to Statfs "/run/user/1001/netns/cni-5fb97adc-b6ef-17e8-565b-0481b311ba09": no such file or directory
```

Run `podman info` and make sure `runRoot` points to `/run/user/$UID/containers` (see below). If it instead points to `/tmp/podman-run-$UID` you will likely have problems when restarting the machine. Typically this happens because you used su to become the harvest user or ran podman as root. You can fix this by logging in as the `harvest` user and running `podman system reset`.

```bash
podman info | grep runRoot
  runRoot: /run/user/1001/containers
```

### Linger errors
When you logout, [systemd](https://github.com/containers/podman/blob/main/troubleshooting.md#21-a-rootless-container-running-in-detached-mode-is-closed-at-logout) may remove some temp files and tear down Podman's rootless network. **Workaround** is to run the following as the harvest user. Details [here](https://github.com/containers/podman/issues/6800)

```bash
loginctl enable-linger
```

## Versions

The following versions were used to validate this workflow.

```bash
podman version

Version:      3.2.3
API Version:  3.2.3
Go Version:   go1.15.7
Built:        Thu Jul 29 11:02:43 2021
OS/Arch:      linux/amd64

docker-compose -v
docker-compose version 1.29.2, build 5becea4c

cat /etc/redhat-release
Red Hat Enterprise Linux release 8.4 (Ootpa)
```

# References
- https://github.com/containers/podman
- https://www.redhat.com/sysadmin/sudo-rootless-podman
- https://www.redhat.com/sysadmin/podman-docker-compose
- https://fedoramagazine.org/use-docker-compose-with-podman-to-orchestrate-containers-on-fedora/
- https://podman.io/getting-started/network.html mentions the need for `podman-plugins`, otherwise rootless containers running in separate containers cannot see each other
- [Troubleshoot Podman](https://github.com/containers/podman/blob/main/troubleshooting.md)