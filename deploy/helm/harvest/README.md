# Harvest Helm chart

A Helm chart that runs [NetApp Harvest](https://github.com/NetApp/harvest) on Kubernetes, one pod per monitored
system. It is a reference: use it as is or change it to fit your environment.

Each Harvest release publishes the chart to `oci://ghcr.io/netapp/charts/harvest`. The chart version is the Harvest
version, so chart `26.11.0` runs Harvest `26.11.0`.

- Install, credentials, adding and removing systems, upgrades and troubleshooting:
  [Kubernetes install guide](https://netapp.github.io/harvest/latest/install/k8/)
- Every chart value, with its default and description: [values.yaml](values.yaml), or
  `helm show values oci://ghcr.io/netapp/charts/harvest --version 26.11.0`

Requirements: Kubernetes 1.21+, Helm 3.10+, and the Prometheus Operator unless you scrape the pods yourself.

```bash
helm install harvest oci://ghcr.io/netapp/charts/harvest --version 26.11.0 \
  --namespace harvest --create-namespace --values values.yaml
```

To install the chart from this repository instead, for example with a nightly build, see
[Install from source](https://netapp.github.io/harvest/latest/install/k8/#install-from-source).
