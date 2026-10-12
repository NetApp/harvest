# Harvest Helm chart

An example Helm chart that runs [NetApp Harvest](https://github.com/NetApp/harvest) on Kubernetes, one pod per
monitored system. It is a reference: use it as is or change it to fit your environment.

- Install, credentials, adding and removing systems, upgrades and troubleshooting:
  [Kubernetes install guide](https://netapp.github.io/harvest/latest/install/k8/)
- Every chart value, with its default and description: [values.yaml](values.yaml)

Requirements: Kubernetes 1.21+, Helm 3.10+, and the Prometheus Operator unless you scrape the pods yourself.

```bash
helm install harvest ./deploy/helm/harvest --namespace harvest --create-namespace --values values.yaml
```
