# harvest

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 26.02.0](https://img.shields.io/badge/AppVersion-26.02.0-informational?style=flat-square)

NetApp Harvest exporter for NetApp ONTAP storage systems.

## Source Code

* <https://github.com/NetApp/harvest>

## Overview

This chart deploys [NetApp Harvest](https://github.com/NetApp/harvest) on Kubernetes. It supports two scrape modes:

1. **Per-poller Prometheus port** (default): each poller gets its own container port and a `PodMonitor`, more info at [per-poller-prom_port](https://netapp.github.io/harvest/latest/prometheus-exporter/#per-poller-prom_port).
2. **HTTP Service Discovery** (optional): `admin` Deployment exposes Harvest's `/api/v1/sd` endpoint. A `ScrapeConfig` points Prometheus at it so new pollers are discovered automatically, more info at [enable-http-service-discovery-in-harvest](https://netapp.github.io/harvest/latest/prometheus-exporter/#enable-http-service-discovery-in-harvest).

For each entry under `harvestConfig.Pollers`, the chart declares:
- a `Deployment` named `<release>-poller-<name>` (1 replica)
- a `PodMonitor` that scrapes the pod on the poller's `prom_port`
- optional collector-extension `ConfigMap`s, if a poller sets `collectorsExtensions`

## Prerequisites

- Kubernetes 1.21+
- Helm 3.10+
- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator) for `PodMonitor` (and `ScrapeConfig`, which needs Operator v0.65+)

## Installing the Chart

```bash
helm install harvest ./deploy/helm/harvest \
  --set "harvestConfig.Pollers.ontap-prod.addr=10.0.0.1" \
  --set "harvestConfig.Pollers.ontap-prod.prom_port=12990" \
  --set "harvestConfig.Pollers.ontap-prod.username=harvest" \
  --set "harvestConfig.Pollers.ontap-prod.password=secret"
```

Or with a values file: `helm install harvest ./deploy/helm/harvest --values my-values.yaml`.

> The inline password above is only for a quick test; it ends up in a ConfigMap. For anything
> real, keep credentials out of the config and source them from a Secret (see [Credentials & Secrets](#credentials--secrets)).

## Configuration

### Adding a poller

Each key under `harvestConfig.Pollers` becomes a separate Deployment. Pick a unique `prom_port` per poller.

```yaml
harvestConfig:
  Pollers:
    ontap-prod-1:
      addr: 10.0.0.1
      username: harvest
      password: secret
      prom_port: 12990

    ontap-prod-2:
      addr: 10.0.0.2
      username: harvest
      password: secret
      prom_port: 12991
```

Each entry holds the native Harvest poller config (written to `harvest.yml`) plus a few chart-only keys (stripped before `harvest.yml` is written).

Native Harvest fields (see the [poller docs](https://netapp.github.io/harvest/latest/prometheus-exporter/#per-poller-prom_port) for the full list):

| Field | Notes |
|-------|-------|
| `addr` | ONTAP management LIF FQDN/IP (required) |
| `prom_port` | Prometheus port, unique per poller; also drives the container port and PodMonitor target (required) |
| `username` | ONTAP user |
| `password` | ONTAP password; use `${VAR}` and a Secret, see [Credentials & Secrets](#credentials--secrets) |
| `collectors` | list of collectors (e.g. `Rest`, `RestPerf`, `Zapi`, `ZapiPerf`, `Ems`) |
| `datacenter`, `auth_style`, `use_insecure_tls`, `tls_min_version`, … | other Harvest poller fields |

Chart-only keys, not sent to Harvest:

| Key | Purpose |
|-----|---------|
| `extraEnvVars` | env injected into this poller's pod only |
| `collectorsExtensions` | per-collector `{objects: {<Name>: <file>.yaml}}`, generates and mounts a `custom.yaml` |
| `podMonitor` | per-poller PodMonitor overrides: `enabled`, `path`, `portName`, `interval`, `scrapeTimeout`, `honorLabels`, `relabelings`, `metricRelabelings` |

To prepare an ONTAP cluster for Harvest see [Prepare cDOT clusters](https://netapp.github.io/harvest/latest/prepare-cdot-clusters/).

### Credentials & Secrets

The chart doesn't manage ONTAP credentials. Manage the Secret separately; the chart only references it.

Keep passwords out of `harvest.yml` and out of Helm values. Put a `${VAR}` placeholder in the
poller's `password` and supply the real value from a Secret through the poller's `extraEnvVars`.
Harvest resolves `${VAR}` from the environment at startup
([variable expansion](https://netapp.github.io/harvest/latest/configure-harvest-advanced/#variable-expansion)),
so only the placeholder lands in `harvest.yml` and the password reaches that poller's pod, and no other.

Create the Secret with `kubectl`, Vault, External Secrets, or sealed-secrets.
For example, with `kubectl`:

```bash
kubectl create secret generic ontap-prod-credentials \
  --namespace <release-namespace> \
  --from-literal=password='<ontap-password>'
```

Then reference it:

```yaml
harvestConfig:
  Pollers:
    ontap-prod:
      addr: 10.0.0.1
      username: harvest
      password: ${ONTAP_PASSWORD}
      prom_port: 12990
      extraEnvVars:
        - name: ONTAP_PASSWORD
          valueFrom:
            secretKeyRef:
              name: ontap-prod-credentials
              key: password
```

Environment variables shared by every poller (a proxy, for example) go in the top-level `poller.extraEnvVars`.
For credentials that rotate, point the poller at Harvest's
[`credentials_script`](https://netapp.github.io/harvest/latest/configure-harvest-basic/#credentials-script) instead.

To manage `harvest.yml` outside the chart (e.g. a ConfigMap
rendered by GitOps tooling), set `config.existingConfigMap` to a ConfigMap containing a
`harvest.yml` key; the chart mounts it and skips generation:

```yaml
config:
  existingConfigMap: my-harvest-config
```

### Extending collector objects

`collectorsExtensions` writes each collector's block straight to a `custom.yaml`, mounted at `conf/<collector>/custom.yaml`, which Harvest merges with the collector's `default.yaml`.

```yaml
harvestConfig:
  Pollers:
    ontap-prod:
      addr: 10.0.0.1
      prom_port: 12990
      collectors:
        - Rest
        - RestPerf
      collectorsExtensions:
        Rest:
          objects:
            NFSClients: nfs_clients.yaml
        RestPerf:
          objects:
            Qtree: qtree.yaml
```

See [Extend an existing object template](https://netapp.github.io/harvest/latest/configure-templates/#extend-an-existing-object-template).

> The chart mounts `custom.yaml` but not the object template `.yaml` files it references.
> Enabling an object the image already ships works (many built-in objects are disabled by default).
> A brand-new object whose template is not in the image needs that template baked into the image (left as a follow-up).

### Enabling HTTP Service Discovery (admin mode)

> Currently experimental

To use Harvest's admin SD endpoint instead of per-poller PodMonitors, enable `adminSD` to
bring up the admin component and turn off per-poller PodMonitors:

```yaml
monitoring:
  scrape:
    adminSD:
      enabled: true
    podMonitor:
      enabled: false
```

This creates an `admin` Deployment, a headless `Service`, and a `ScrapeConfig` pointing Prometheus at the SD endpoint. Tune the admin workload under the `admin` values.

See [Enable HTTP Service Discovery](https://netapp.github.io/harvest/latest/prometheus-exporter/#enable-http-service-discovery-in-harvest).

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| admin | object | see the admin.* keys below | Deployment settings for the admin service-discovery pod (single instance), created when monitoring.scrape.adminSD.enabled is true. |
| admin.affinity | object | `{}` | Affinity rules for pod scheduling. (corev1.Affinity) |
| admin.args | list | `[]` | Additional args appended to the admin command. (string[]) |
| admin.automountServiceAccountToken | bool | `false` | Mount the ServiceAccount token. The admin node never calls the Kubernetes API, so it is off by default (least privilege). (bool) |
| admin.command | list | `["/opt/harvest/bin/harvest","admin","start","--config","/opt/harvest/harvest.yml"]` | Admin entrypoint command. Leave as default unless customizing. (string[]) |
| admin.containerPorts | object | `{"httpsd":8887}` | Container ports exposed by the admin pod. Names must match probes and Service. |
| admin.containerSecurityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"add":["NET_RAW","NET_BIND_SERVICE"],"drop":["ALL"]},"readOnlyRootFilesystem":false,"seccompProfile":{"type":"RuntimeDefault"}}` | Container-level security posture for admin. Leave empty ({}) to omit. |
| admin.containerSecurityContext.capabilities.add | list | `["NET_RAW","NET_BIND_SERVICE"]` | Admin may need networking and port binding, keep minimal caps. |
| admin.extraEnvVars | list | `[]` | Extra environment variables injected into the admin container. (corev1.EnvVar[]) |
| admin.livenessProbe | object | `{"enabled":false}` | Liveness probe for admin API. Disable if startup is slow. |
| admin.nodeSelector | object | `{}` | Node selector for admin pods. (map[string]string) |
| admin.podSecurityContext | object | `{}` | Pod-level security context for admin. Leave empty ({}) to omit. |
| admin.priorityClassName | string | `""` | PriorityClass for admin pods. (string) |
| admin.readinessProbe | object | `{"enabled":true,"failureThreshold":6,"httpGet":{"httpHeaders":[{"name":"Accept","value":"application/json"}],"path":"/api/v1/sd","port":"httpsd"},"initialDelaySeconds":5,"periodSeconds":5,"timeoutSeconds":2}` | Readiness probe to mark admin pod ready for traffic. |
| admin.readinessProbe.failureThreshold | int | `6` | Failures required to mark NotReady. |
| admin.readinessProbe.initialDelaySeconds | int | `5` | Delay before first readiness check. |
| admin.readinessProbe.periodSeconds | int | `5` | Probe period. |
| admin.readinessProbe.timeoutSeconds | int | `2` | Single probe timeout. |
| admin.replicaCount | int | `1` | Number of admin (harvest admin SD) replicas. |
| admin.resources | object | `{}` | Container resource requests and limits. Empty sets none — set these for production. (corev1.ResourceRequirements) |
| admin.service.clusterIP | string | `"None"` | Set to 'None' for a headless service. |
| admin.service.ports | object | `{"httpsd":8887}` | Service ports exposed. Names must match containerPorts. |
| admin.service.type | string | `"ClusterIP"` | Service type for admin. ClusterIP recommended. |
| admin.startupProbe.failureThreshold | int | `24` | Failures required to mark startup as failed. |
| admin.startupProbe.httpGet.httpHeaders | list | `[{"name":"Accept","value":"application/json"}]` | Optional headers for the probe request. |
| admin.startupProbe.httpGet.path | string | `"/api/v1/sd"` | Admin API health/SD endpoint. |
| admin.startupProbe.httpGet.port | string | `"httpsd"` | Port name/number to hit. Must exist in containerPorts. |
| admin.startupProbe.initialDelaySeconds | int | `3` | Delay before probing starts. |
| admin.startupProbe.periodSeconds | int | `5` | Probe period. |
| admin.terminationGracePeriodSeconds | int | `20` | Graceful termination window to allow cleanup. |
| admin.tolerations | list | `[]` | Tolerations for admin pods. (corev1.Toleration[]) |
| admin.topologySpreadConstraints | list | `[]` | Topology spread constraints to evenly distribute admin pods. (corev1.TopologySpreadConstraint[]) |
| admin.updateStrategy | object | `{"type":"Recreate"}` | Update strategy for the admin workload. Recreate ensures single active instance. (object) |
| config.existingConfigMap | string | `""` | Mount an existing ConfigMap (with a `harvest.yml` key) instead of generating one. `harvestConfig` is then only used to enumerate pollers, not rendered. (string) |
| fullnameOverride | string | `""` | Override the full release name. Leave empty to use the chart's computed fullname. |
| harvestConfig | object | see the harvestConfig.* keys below | Harvest configuration rendered into harvest.yml: Exporters, Defaults, and per-cluster Pollers. |
| harvestConfig.Defaults | object | `{"exporters":["prometheus"]}` | Defaults applied to pollers unless overridden. |
| harvestConfig.Defaults.exporters | list | `["prometheus"]` | Exporters enabled by default for all pollers. (string[]) |
| harvestConfig.Exporters | object | `{"prometheus":{"add_meta_tags":true,"exporter":"Prometheus","sort_labels":true}}` | Exporter definitions for Harvest. Keys under `Exporters` name the exporters. |
| harvestConfig.Exporters.prometheus.add_meta_tags | bool | `true` | If true, add ONTAP/cluster metadata as Prometheus labels |
| harvestConfig.Exporters.prometheus.exporter | string | `"Prometheus"` | Exporter implementation name. Must match a Harvest exporter. |
| harvestConfig.Exporters.prometheus.sort_labels | bool | `true` | If true, sort labels for deterministic output. (bool) |
| harvestConfig.Pollers | object | `{}` | Map of poller name to config: native Harvest fields (written to harvest.yml) plus chart-only keys (extraEnvVars, collectorsExtensions, podMonitor). See the README "Adding a poller" section for the entry schema. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. One of: Always|IfNotPresent|Never. |
| image.pullSecrets | list | `[]` | List of imagePullSecrets to use for private registries. Example: ["my-regcred"] |
| image.registry | string | `"ghcr.io"` | Container registry hosting the image. |
| image.repository | string | `"netapp/harvest"` | Repository (image name) within the registry. |
| image.tag | string | `"26.02.0-1"` | Image tag to deploy. Pin to an explicit version to overwrite chart appVersion. |
| monitoring.scrape.adminSD | object | `{"enabled":false,"labels":{"release":"prometheus-operator"}}` | Admin service-discovery endpoint scrape settings. |
| monitoring.scrape.adminSD.enabled | bool | `false` | If true, enable scraping the admin SD endpoint (Prometheus target discovery). (bool) |
| monitoring.scrape.adminSD.labels | object | `{"release":"prometheus-operator"}` | Labels for the ScrapeConfig, used by the Prometheus Operator's scrapeConfigSelector to discover it. (map[string]string) |
| monitoring.scrape.podMonitor | object | `{"defaultInterval":"30s","defaultScrapeTimeout":"20s","enabled":true,"labels":{"release":"prometheus-operator"}}` | Prometheus Operator PodMonitor configuration for pollers/admin. |
| monitoring.scrape.podMonitor.defaultInterval | string | `"30s"` | Default scrape interval for PodMonitors. (duration) |
| monitoring.scrape.podMonitor.defaultScrapeTimeout | string | `"20s"` | Default scrape timeout for PodMonitors. Must be < interval. (duration) |
| monitoring.scrape.podMonitor.enabled | bool | `true` | If true, create PodMonitor resources. Requires kube-prometheus-stack or prometheus-operator. (bool) |
| monitoring.scrape.podMonitor.labels | object | `{"release":"prometheus-operator"}` | Extra labels to attach to PodMonitors for selector-based discovery. (map[string]string) |
| poller | object | see the poller.* keys below | Deployment settings shared by every poller pod. Per-cluster Harvest config (addr, credentials, collectors) lives under harvestConfig.Pollers, not here. |
| poller.affinity | object | `{}` | Affinity rules for pod scheduling. (corev1.Affinity) |
| poller.automountServiceAccountToken | bool | `false` | Mount the ServiceAccount token. Harvest never calls the Kubernetes API, so it is off by default (least privilege). Set true only for a sidecar that needs API access. (bool) |
| poller.containerSecurityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"add":["NET_RAW"],"drop":["ALL"]},"readOnlyRootFilesystem":true,"seccompProfile":{"type":"RuntimeDefault"}}` | Container-level security context. Tighten when possible. Leave empty ({}) to omit. See: https://kubesec.io/basics/securitycontext-capabilities/ See: https://www.aquasec.com/cloud-native-academy/kubernetes-in-production/kubernetes-security-context/#section-5 |
| poller.containerSecurityContext.allowPrivilegeEscalation | bool | `false` | Prevent privilege escalation via setuid binaries. (bool) |
| poller.containerSecurityContext.capabilities.add | list | `["NET_RAW"]` | We need to allow ICMP for harvest poller status metric. |
| poller.containerSecurityContext.capabilities.drop | list | `["ALL"]` | Drop all |
| poller.containerSecurityContext.readOnlyRootFilesystem | bool | `true` | If true, mount filesystem read-only; may need writable paths for Harvest cache. (bool) |
| poller.containerSecurityContext.seccompProfile | object | `{"type":"RuntimeDefault"}` | Seccomp profile to use; RuntimeDefault is recommended. (seccompProfile) |
| poller.extraEnvVars | list | `[]` | Environment variables injected into every poller container. For per-poller secrets, use extraEnvVars on the poller entry under harvestConfig.Pollers. (corev1.EnvVar[]) |
| poller.lifecycleHooks | object | `{}` | Container lifecycle hooks (postStart/preStop). (corev1.Lifecycle) |
| poller.livenessProbe | object | `{"enabled":false}` | Liveness probe toggles. Disable if startup is slow and causes restarts. (bool) |
| poller.nodeSelector | object | `{}` | Node selector for scheduling. (map[string]string) |
| poller.podAnnotations | object | `{}` | Extra annotations for poller pods. (map[string]string) |
| poller.podSecurityContext | object | `{}` | Pod-level security context. Leave empty ({}) to omit. Some exporters require root; Known bug: poller status exporter may fail if not root. |
| poller.priorityClassName | string | `""` | PriorityClass for poller pods. (string) |
| poller.readinessProbe | object | `{"enabled":true,"failureThreshold":3,"periodSeconds":10,"tcpSocket":{"port":"http"},"timeoutSeconds":2}` | Readiness probe to gate traffic until ready. |
| poller.readinessProbe.failureThreshold | int | `3` | Failures required to mark NotReady. (int) |
| poller.readinessProbe.periodSeconds | int | `10` | Frequency of readiness checks. (int seconds) |
| poller.readinessProbe.tcpSocket.port | string | `"http"` | Port (name or number) to check open. Should match the metrics or HTTP port name. (string|int) |
| poller.readinessProbe.timeoutSeconds | int | `2` | Single probe timeout. (int seconds) |
| poller.replicaCount | int | `1` | Number of poller replicas. Typically 1 per ONTAP cluster to avoid duplicate scraping. (int) |
| poller.resources | object | `{}` | Container resource requests and limits. Empty sets none — set these for production. (corev1.ResourceRequirements) |
| poller.startupProbe | object | `{"enabled":true,"failureThreshold":12,"httpGet":{"path":"/metrics","port":"http"},"initialDelaySeconds":5,"periodSeconds":5}` | Startup probe checks readiness before liveness begins. Tune thresholds for large clusters. |
| poller.startupProbe.failureThreshold | int | `12` | Consecutive failures before marking container unhealthy during startup. (int) |
| poller.startupProbe.httpGet.path | string | `"/metrics"` | Metrics endpoint path exposed by the exporter. (string) |
| poller.startupProbe.httpGet.port | string | `"http"` | Name of the container port that serves metrics (must match container port name). (string|int) |
| poller.startupProbe.initialDelaySeconds | int | `5` | Delay before probing starts. (int seconds) |
| poller.startupProbe.periodSeconds | int | `5` | Probe period. (int seconds) |
| poller.tolerations | list | `[]` | Tolerations for tainted nodes. (corev1.Toleration[]) |
| poller.topologySpreadConstraints | list | `[]` | Topology spread constraints to distribute poller pods. (corev1.TopologySpreadConstraint[]) |
| poller.updateStrategy | object | `{"type":"Recreate"}` | Update strategy for the poller Deployment. Recreate avoids dual-scrape. (object) |
