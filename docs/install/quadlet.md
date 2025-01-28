# Containerized Harvest on Linux using Podman Quadlets

This documentation describes how to run Harvest in a container using Podman Quadlets. Quadlets are a way to manage Podman containers with systemd. This is useful for running Harvest as a service on Linux.

In this guide, we'll create a Podman Quadlet for each Harvest poller and a systemd service that starts the pollers when the system boots. 

## Summary of Steps

1. Install Harvest, set up your `harvest.yml`, and generate a `harvest-compose.yml` using the Harvest CLI per the [Harvest documentation](containers.md#setup-harvestyml). [details](#generate-harvest-compose-files)
2. Use [podlet](https://github.com/containers/podlet) to generate Quadlet files from the `harvest-compose.yml` file from step 1. [details](#podlet)
3. Make a few edits to the generated Quadlet files and move them to `/etc/containers/systemd/`. [details](#edit-quadlet-files)
4. Tell systemd to reload the services by running `systemctl daemon-reload`.
5. Start the systemd managed pollers (e.g., `sudo systemctl start poller-sar`).
6. Reboot the machine and verify that all pollers restart.

### Generate Harvest Compose Files
Install Harvest and [generate harvest-compose.yml](containers.md#generate-a-docker-compose-for-your-pollers).

```bash
# cat example harvest-compose.yml

services:
  u2:
    image: ghcr.io/netapp/harvest:latest
    container_name: poller-u2
    ports:
      - "12990:12990"
    command: '--poller u2 --promPort 12990 --config /opt/harvest.yml'
    volumes:
      - ./cert:/opt/harvest/cert
      - ./harvest.yml:/opt/harvest.yml
      - ./conf:/opt/harvest/conf
    networks:
      - backend

  sar:
    image: ghcr.io/netapp/harvest:latest
    container_name: poller-sar
    ports:
      - "12991:12991"
    command: '--poller sar --promPort 12991 --config /opt/harvest.yml'
    volumes:
      - ./cert:/opt/harvest/cert
      - ./harvest.yml:/opt/harvest.yml
      - ./conf:/opt/harvest/conf
    networks:
      - backend
```

### Podlet

Install [podlet](https://github.com/containers/podlet) or compare the example `harvest-compose.yml` above with the generated quadlets below and make similar edits for your `harvest-compose.yml`.

```bash
podlet compose harvest-compose.yml
```

```bash
# u2.container
[Container]
ContainerName=poller-u2
Exec=--poller u2 --promPort 12990 --config /opt/harvest.yml
Image=ghcr.io/netapp/harvest:latest
Network=backend
PublishPort=12990:12990
Volume=./cert:/opt/harvest/cert
Volume=./harvest.yml:/opt/harvest.yml
Volume=./conf:/opt/harvest/conf

---

# sar.container
[Container]
ContainerName=poller-sar
Exec=--poller sar --promPort 12991 --config /opt/harvest.yml
Image=ghcr.io/netapp/harvest:latest
Network=backend
PublishPort=12991:12991
Volume=./cert:/opt/harvest/cert
Volume=./harvest.yml:/opt/harvest.yml
Volume=./conf:/opt/harvest/conf
```

### Edit Quadlet Files

Podlet created two YAML documents as shown above, one for each poller.
We are going to make the following adjustments to the podlet output and copy/paste the final output into `/etc/containers/systemd/poller-u2.container` and `/etc/containers/systemd/poller-sar.container`.

The edits are:

1. Remove the `Network=backend` line that was included for Prometheus.
2. Change all Volume paths to fully qualified paths. (e.g., change `./cert` to `/opt/harvest/cert` if you installed Harvest at `/opt/harvest`)
3. Because we want to restart the pollers when the machine reboots add the following section:

```service
[Install]
# Start by default on boot
WantedBy=multi-user.target default.target

[Service]
Restart=always
```

Here is the final output for the sar poller service with all the edits applied: 

```service
# sar.container
[Container]
ContainerName=poller-sar
Exec=--poller sar --promPort 12991 --config /opt/harvest.yml
Image=ghcr.io/netapp/harvest:latest
PublishPort=12991:12991
Volume=/opt/harvest/cert:/opt/harvest/cert
Volume=/opt/harvest/harvest.yml:/opt/harvest.yml
Volume=/opt/harvest/conf:/opt/harvest/conf

[Install]
# Start by default on boot
WantedBy=multi-user.target default.target

[Service]
Restart=always
```

### Move Quadlet Files

Move each service file to `/etc/containers/systemd/` e.g. `mv poller-sar.container /etc/containers/systemd/poller-sar.container`

# References
- [Make systemd better for Podman with Quadlet](https://www.redhat.com/en/blog/quadlet-podman)
- https://docs.podman.io/en/latest/markdown/podman-systemd.unit.5.html
- https://github.com/containers/podlet