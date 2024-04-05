## How do I migrate from Harvest 1.6 to 2.0?
There currently is not a tool to migrate data from Harvest 1.6 to 2.0. The most common workaround is to run both, 1.6 and 2.0, in parallel. Run both, until the 1.6 data expires due to normal retention policy, and then fully cut over to 2.0.

Technically, it’s possible to take a Graphite DB, extract the data, and send it to a Prometheus db, but it’s not an area we’ve invested in.
If you want to explore that option, check out the [promtool](https://github.com/prometheus/prometheus/issues/8280) which supports importing, but probably not worth the effort.

## How do I share sensitive log files with NetApp?
Email them to [ng-harvest-files@netapp.com](mailto:ng-harvest-files@netapp.com)
This mail address is accessible to NetApp Harvest employees only.

## Multi-tenancy
### Question
Is there a way to allow per SVM level user views?  I need to offer 1 tenant per SVM. Can I limit visibility to specific SVMs? Is there an SVM dashboard available?

### Answer
You can do this with Grafana. Harvest can provide the labels for SVMs. The pieces are there but need to be put together.

Grafana templates support the $__user variable to make pre-selections and decisions.
You can use that + metadata mapping the user <-> SVM. With both of those you can build SVM specific dashboards.

There is a German service provider who is doing this. They have service managers responsible for a set of customers – and only want to see the data/dashboards of their corresponding customers.

## Harvest Authentication and Permissions
### Question
What permissions does Harvest need to talk to ONTAP?
### Answer
Permissions, authentication, role based security, and creating a Harvest user are covered [here](https://netapp.github.io/harvest/prepare-cdot-clusters/).

## ONTAP counters are missing
### Question
How do I make Harvest collect additional ONTAP counters?

### Answer
Instead of modifying the out-of-the-box templates in the `conf/` directory, it is better to create your own custom templates following these [instructions](https://github.com/NetApp/harvest/blob/main/cmd/collectors/zapiperf/README.md#creatingediting-subtemplates).

## Capacity Metrics
### Question
How are capacity and other metrics calculated by Harvest?
### Answer
Each collector has its own way of collecting and post-processing metrics. Check the documentation of each individual collector (usually under section #Metrics). Capacity and hardware-related metrics are collected by the [Zapi collector](https://github.com/NetApp/harvest/tree/main/cmd/collectors/zapi/README.md#Metrics) which emits metrics as they are without any additional calculation. Performance metrics are collected by the [ZapiPerf collector](https://github.com/NetApp/harvest/tree/main/cmd/collectors/zapiperf/README.md#Metrics) and the final values are calculated from the delta of two consequent polls.

## Tagging Volumes
### Question
How do I tag ONTAP volumes with metadata and surface that data in Harvest?
### Answer
See [volume tagging issue](https://github.com/NetApp/harvest/issues/209#issuecomment-879751733) and [volume tagging via sub-templates](https://github.com/NetApp/harvest/issues/309#issuecomment-882553771)

## REST and Zapi Documentation
### Question
How do I relate ONTAP REST endpoints to ZAPI APIs and attributes?
### Answer
Please refer to the [ONTAPI to REST API mapping document](https://library.netapp.com/ecm/ecm_download_file/ECMLP2874886).

## Sizing
How much disk space is required by Prometheus?

This depends on the collectors you've added, # of nodes monitored, cardinality of labels, # instances, retention, ingest rate, etc. A good approximation is to curl your Harvest exporter and count the number of samples that it publishes and then feed that information into a Prometheus sizing formula.

Prometheus stores an average of 1-2 bytes per sample. To plan the capacity of a Prometheus server, you can use the rough formula: needed_disk_space = retention_time_seconds * ingested_samples_per_second * bytes_per_sample

A rough approximation is outlined https://devops.stackexchange.com/questions/9298/how-to-calculate-disk-space-required-by-prometheus-v2-2

## Topk usage in Grafana

### Question
In Grafana, why do I see more results from topk than I asked for?

### Answer
[Topk](https://prometheus.io/docs/prometheus/latest/querying/operators/#aggregation-operators) is one of Prometheus's out-of-the-box aggregation operators, and is used to calculate the **largest k elements by sample value**.

Depending on the time range you select, Prometheus will often return more results than you asked for. That's because Prometheus is picking the topk for each time in the graph. In other words, different time series are the topk at different times in the graph. When you use a large duration, there are often many time series.

This is a limitation of Prometheus and can be mitigated by:

- reducing the time range to a smaller duration that includes fewer topk results - something like a five to ten minute range works well for most of Harvest's charts
- the panel's table shows the current topk rows and that data can be used to supplement the additional series shown in the charts

Additional details: [here](https://stackoverflow.com/questions/38783424/prometheus-topk-returns-more-results-than-expected), [here](https://www.robustperception.io/graph-top-n-time-series-in-grafana), and [here](https://grafana.com/docs/grafana/latest/datasources/prometheus/)

## Where are Harvest container images published?

Harvest container images are published to both GitHub's image registry (ghcr.io) and Docker's image registry (hub.docker.com). By default, `ghcr.io` is used for pulling images.

Please note that `cr.netapp.io` is no longer being maintained. If you have been using `cr.netapp.io` to pull Harvest images, we encourage you to switch to `ghcr.io` or Docker Hub as your container image registry. Starting in 2024, we will cease publishing Harvest container images to `cr.netapp.io`.

### How do I switch from Github (ghcr.io) to Docker's image registry (hub.docker.com) or vice-versa?

### Answer

Replace all instances of `ghcr.io/netapp/harvest:latest` with `rahulguptajss/harvest:latest`

- Edit your docker-compose file and make those replacements or regenerate the compose file using the `--image rahulguptajss/harvest:latest` option)

- Update any shell or Ansible scripts you have that are also using those images

- After making these changes, you should stop your containers, pull new images, and restart.

You can verify that you're using the Docker Hub images like so:

**Before**

```
docker image ls -a
REPOSITORY                  TAG       IMAGE ID       CREATED        SIZE
ghcr.io/netapp/harvest      latest    80061bbe1c2c   10 days ago    56.4MB <=== GitHub Container Registry
prom/prometheus             v2.33.1   e528f02c45a6   3 weeks ago    204MB
grafana/grafana             8.3.4     4a34578e4374   5 weeks ago    274MB
```

**Pull image from Docker Hub**

```
docker pull rahulguptajss/harvest:latest
Using default tag: latest
latest: Pulling from rahulguptajss/harvest
Digest: sha256:6ff88153812ebb61e9dd176182bf8a792cde847748c5654d65f4630e61b1f3ae
Status: Image is up to date for rahulguptajss/harvest:latest
rahulguptajss/harvest:latest
```

Notice that the `IMAGE ID` for both images are identical since the images are the same.

```
docker image ls -a
REPOSITORY                  TAG       IMAGE ID       CREATED        SIZE
rahulguptajss/harvest       latest    80061bbe1c2c   10 days ago    56.4MB  <== Harvest image from Docker Hub
ghcr.io/netapp/harvest      latest    80061bbe1c2c   10 days ago    56.4MB
prom/prometheus             v2.33.1   e528f02c45a6   3 weeks ago    204MB
grafana/grafana             8.3.4     4a34578e4374   5 weeks ago    274MB
```

We can now remove the DockerHub pulled image

```
docker image rm ghcr.io/netapp/harvest:latest
Untagged: ghcr.io/netapp/harvest:latest
Untagged: ghcr.io/netapp/harvest@sha256:6ff88153812ebb61e9dd176182bf8a792cde847748c5654d65f4630e61b1f3ae

docker image ls -a
REPOSITORY              TAG       IMAGE ID       CREATED        SIZE
rahulguptajss/harvest   latest    80061bbe1c2c   10 days ago    56.4MB
prom/prometheus         v2.33.1   e528f02c45a6   3 weeks ago    204MB
grafana/grafana         8.3.4     4a34578e4374   5 weeks ago    274MB
```

## Ports

### What ports does Harvest use?

### Answer

<!-- Mermaid used to create SVG

https://mermaid-js.github.io/mermaid-live-editor/edit#pako:eNp90MGKwjAQgOFXkTkrpI0XIywou3gRLbsecxnMVAtNU6YJy2J8982WlKUXc0r4vyEwD7g6Q6DgxtjfF5_Hre4W6VSubYmL1eotqvVaxvPpsquKWSznsZxFOY9yiuws-TuFYex9elaOfRHzf69ZmVn5msnMpk8PjDV2OJqN2Ij4P5fFnt33QDwKKYSIeSRlWIIlttiYtKTHH9eQRi1pUOlqqMbQeg26eyYaeoOePkzjHYOqsR1oCRi8-_rprqA8B5rQe4Np5zar5y8B_YDH

graph LR;
    Poller1->|:443|ONTAP1;
    Poller2->|:443|ONTAP2;
    Poller3->|:443|ONTAP3;
    Prometheus->|:promPort1|Poller1;
    Prometheus->|:promPort2|Poller2;
    Prometheus->|:promPort3|Poller3;
    Grafana->|:9090|Prometheus;
    Browser->|:3000|Grafana;
-->

The default ports are shown in the following diagram.

![h](https://user-images.githubusercontent.com/242252/172162362-49909154-623e-4feb-9e13-e9f9f75d2b69.svg)

- Harvest's pollers use ZAPI or REST to communicate with ONTAP on port `443`
- Each poller exposes the Prometheus port defined in your `harvest.yml` file
- Prometheus scrapes each poller-exposed Prometheus port (`promPort1`, `promPort2`, `promPort3`)
- Prometheus's default port is `9090`
- Grafana's default port is `3000`


## Snapmirror_labels

### Why do my snapmirror_labels have an empty source_node?

### Answer

Snapmirror relationships have a source and destination node. ONTAP however does not expose the source side of that relationship, only the destination side is returned via ZAPI/REST APIs. Because of that, the Prometheus metric named, `snapmirror_labels`, will have an empty `source_node` label.

The dashboards show the correct value for `source_node` since we join multiple metrics in the Grafana panels to synthesize that information.

In short: don't rely on the `snapmirror_labels` for `source_node` labels. If you need `source_node` you will need to do a similar join as the Snapmirror dashboard does.

See https://github.com/NetApp/harvest/issues/1192 for more information and linked pull requests for REST and ZAPI.

## NFS Clients Dashboard

### Why do my NFS Clients Dashboard have no data?

### Answer

NFS Clients dashboard is only available through Rest Collector. This information is not available through Zapi. You must enable the Rest collector in your harvest.yml config and uncomment the nfs_clients.yaml section in your [default.yaml](https://github.com/NetApp/harvest/blob/main/conf/rest/default.yaml) file.

**Note:** Enabling nfs_clients.yaml may slow down data collection.

## File Analytics Dashboard

### Why do my File Analytics Dashboard have no data?

### Answer

This dashboard requires ONTAP 9.8+ and the APIs are only available via REST. Please enable the REST collector in your harvest config. To collect and display usage data such as capacity analytics, you need to enable File System Analytics on a volume. Please see https://docs.netapp.com/us-en/ontap/task_nas_file_system_analytics_enable.html for more details.

## Why do I have Volume Sis Stat panel empty in Volume dashboard?

### Answer

This panel requires ONTAP 9.12+ and the APIs are only available via REST. Enable the REST collector in your `harvest.yml` config.