# K8 Deployment

The following steps are provided for reference purposes only. Depending on the specifics of your k8 configuration, you may need to make modifications to the steps or files as necessary.

### Requirements
- Kompose: `v1.25` or higher https://github.com/kubernetes/kompose/

### Download and untar

- Download the latest version of [Harvest](https://netapp.github.io/harvest/latest/install/native/), untar, and
  cd into the harvest directory.

## Deployment

* [Local k8 Deployment](#local-k8-deployment)
* [Cloud Deployment](#cloud-deployment)

## Local k8 Deployment

To run Harvest resources in Kubernetes, please execute the following commands:

1. After configuring the clusters in `harvest.yml`, generate `harvest-compose.yml` and `prom-stack.yml`.

```
bin/harvest generate docker full --port --output harvest-compose.yml
```

<details><summary>harvest.yml</summary>
<p>

```yaml
Tools:
Exporters:
    prometheus1:
        exporter: Prometheus
        port_range: 12990-14000
        add_meta_tags: false
Defaults:
    use_insecure_tls: true
    prefer_zapi: true
Pollers:
    u2:
        datacenter: u2
        addr: ADDRESS
        username: USER
        password: PASS
        collectors:
            - Rest
        exporters:
            - prometheus1
```
</p>
</details>

<details><summary>harvest-compose.yml</summary>
<p>

```yaml
version: "3.7"

services:

  u2:
    image: ghcr.io/netapp/harvest:latest
    container_name: poller-u2
    restart: unless-stopped
    ports:
      - 14999:14999
    command: '--poller u2 --promPort 14999 --config /opt/harvest.yml'
    volumes:
      - /Users/harvest/conf:/opt/harvest/conf
      - /Users/harvest/cert:/opt/harvest/cert
      - /Users/harvest/harvest.yml:/opt/harvest.yml
    networks:
      - backend
```

</p>
</details>

2. Using `kompose`, convert `harvest-compose.yml` and `prom-stack.yml` into Kubernetes resources and save them as `kub.yaml`.

```
kompose convert --file harvest-compose.yml --file prom-stack.yml --out kub.yaml --volumes hostPath
```

<details><summary>kub.yaml</summary>
<p>

```yaml
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    kompose.cmd: kompose convert --file harvest-compose.yml --file prom-stack.yml --out kub.yaml --volumes hostPath
    kompose.service.type: nodeport
    kompose.version: 1.28.0 (HEAD)
  creationTimestamp: null
  labels:
    io.kompose.service: grafana
  name: grafana
spec:
  ports:
    - name: "3000"
      port: 3000
      targetPort: 3000
  selector:
    io.kompose.service: grafana
  type: NodePort
status:
  loadBalancer: {}

---
apiVersion: v1
kind: Service
metadata:
  annotations:
    kompose.cmd: kompose convert --file harvest-compose.yml --file prom-stack.yml --out kub.yaml --volumes hostPath
    kompose.service.type: nodeport
    kompose.version: 1.28.0 (HEAD)
  creationTimestamp: null
  labels:
    io.kompose.service: prometheus
  name: prometheus
spec:
  ports:
    - name: "9090"
      port: 9090
      targetPort: 9090
  selector:
    io.kompose.service: prometheus
  type: NodePort
status:
  loadBalancer: {}

---
apiVersion: v1
kind: Service
metadata:
  annotations:
    kompose.cmd: kompose convert --file harvest-compose.yml --file prom-stack.yml --out kub.yaml --volumes hostPath
    kompose.version: 1.28.0 (HEAD)
  creationTimestamp: null
  labels:
    io.kompose.service: u2
  name: u2
spec:
  ports:
    - name: "14999"
      port: 14999
      targetPort: 14999
  selector:
    io.kompose.service: u2
status:
  loadBalancer: {}

---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    kompose.cmd: kompose convert --file harvest-compose.yml --file prom-stack.yml --out kub.yaml --volumes hostPath
    kompose.service.type: nodeport
    kompose.version: 1.28.0 (HEAD)
  creationTimestamp: null
  labels:
    io.kompose.service: grafana
  name: grafana
spec:
  replicas: 1
  selector:
    matchLabels:
      io.kompose.service: grafana
  strategy:
    type: Recreate
  template:
    metadata:
      annotations:
        kompose.cmd: kompose convert --file harvest-compose.yml --file prom-stack.yml --out kub.yaml --volumes hostPath
        kompose.service.type: nodeport
        kompose.version: 1.28.0 (HEAD)
      creationTimestamp: null
      labels:
        io.kompose.network/harvest-backend: "true"
        io.kompose.network/harvest-frontend: "true"
        io.kompose.service: grafana
    spec:
      containers:
        - image: grafana/grafana:8.3.4
          name: grafana
          ports:
            - containerPort: 3000
          resources: {}
          volumeMounts:
            - mountPath: /var/lib/grafana
              name: grafana-data
            - mountPath: /etc/grafana/provisioning
              name: grafana-hostpath1
      restartPolicy: Always
      volumes:
        - hostPath:
            path: /Users/harvest
          name: grafana-data
        - hostPath:
            path: /Users/harvest/grafana
          name: grafana-hostpath1
status: {}

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  creationTimestamp: null
  name: harvest-backend
spec:
  ingress:
    - from:
        - podSelector:
            matchLabels:
              io.kompose.network/harvest-backend: "true"
  podSelector:
    matchLabels:
      io.kompose.network/harvest-backend: "true"

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  creationTimestamp: null
  name: harvest-frontend
spec:
  ingress:
    - from:
        - podSelector:
            matchLabels:
              io.kompose.network/harvest-frontend: "true"
  podSelector:
    matchLabels:
      io.kompose.network/harvest-frontend: "true"

---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    kompose.cmd: kompose convert --file harvest-compose.yml --file prom-stack.yml --out kub.yaml --volumes hostPath
    kompose.service.type: nodeport
    kompose.version: 1.28.0 (HEAD)
  creationTimestamp: null
  labels:
    io.kompose.service: prometheus
  name: prometheus
spec:
  replicas: 1
  selector:
    matchLabels:
      io.kompose.service: prometheus
  strategy:
    type: Recreate
  template:
    metadata:
      annotations:
        kompose.cmd: kompose convert --file harvest-compose.yml --file prom-stack.yml --out kub.yaml --volumes hostPath
        kompose.service.type: nodeport
        kompose.version: 1.28.0 (HEAD)
      creationTimestamp: null
      labels:
        io.kompose.network/harvest-backend: "true"
        io.kompose.service: prometheus
    spec:
      containers:
        - args:
            - --config.file=/etc/prometheus/prometheus.yml
            - --storage.tsdb.path=/prometheus
            - --web.console.libraries=/usr/share/prometheus/console_libraries
            - --web.console.templates=/usr/share/prometheus/consoles
          image: prom/prometheus:v2.33.1
          name: prometheus
          ports:
            - containerPort: 9090
          resources: {}
          volumeMounts:
            - mountPath: /etc/prometheus
              name: prometheus-hostpath0
            - mountPath: /prometheus
              name: prometheus-data
      restartPolicy: Always
      volumes:
        - hostPath:
            path: /Users/harvest/container/prometheus
          name: prometheus-hostpath0
        - hostPath:
            path: /Users/harvest
          name: prometheus-data
status: {}

---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    kompose.cmd: kompose convert --file harvest-compose.yml --file prom-stack.yml --out kub.yaml --volumes hostPath
    kompose.version: 1.28.0 (HEAD)
  creationTimestamp: null
  labels:
    io.kompose.service: u2
  name: u2
spec:
  replicas: 1
  selector:
    matchLabels:
      io.kompose.service: u2
  strategy:
    type: Recreate
  template:
    metadata:
      annotations:
        kompose.cmd: kompose convert --file harvest-compose.yml --file prom-stack.yml --out kub.yaml --volumes hostPath
        kompose.version: 1.28.0 (HEAD)
      creationTimestamp: null
      labels:
        io.kompose.network/harvest-backend: "true"
        io.kompose.service: u2
    spec:
      containers:
        - args:
            - --poller
            - u2
            - --promPort
            - "14999"
            - --config
            - /opt/harvest.yml
          image: ghcr.io/netapp/harvest:latest
          name: poller-u2
          ports:
            - containerPort: 14999
          resources: {}
          volumeMounts:
            - mountPath: /opt/harvest/conf
              name: u2-hostpath0
            - mountPath: /opt/harvest/cert
              name: u2-hostpath1
            - mountPath: /opt/harvest.yml
              name: u2-hostpath2
      restartPolicy: Always
      volumes:
        - hostPath:
            path: /Users/harvest/conf
          name: u2-hostpath0
        - hostPath:
            path: /Users/harvest/cert
          name: u2-hostpath1
        - hostPath:
            path: /Users/harvest/harvest.yml
          name: u2-hostpath2
status: {}
```
</p>
</details>

3. Apply `kub.yaml` to k8.

```
kubectl apply --filename kub.yaml
```

4. List running pods.

```
kubectl get pods
```

<details><summary>pods</summary>
<p>

```text
NAME                          READY   STATUS    RESTARTS   AGE
prometheus-666fc7b64d-xfkvk   1/1     Running   0          43m
grafana-7cd8bdc9c9-wmsxh      1/1     Running   0          43m
u2-7dfb76b5f6-zbfm6           1/1     Running   0          43m
```
</p>
</details>

#### Remove all Harvest resources from k8

```kubectl delete --filename kub.yaml```

#### Helm Chart

Generate helm charts

```
kompose convert --file harvest-compose.yml --file prom-stack.yml --chart --volumes hostPath --out harvestchart
```

## Cloud Deployment

We will utilize `configMap` to generate Kubernetes resources for deploying Harvest pollers in a cloud environment. Please note the following assumptions for the steps below:

- The steps provided are solely for the deployment of Harvest pollers pods. Separate configurations will be necessary for setting up Prometheus and Grafana.
- Networking between Harvest and Prometheus must be configured, and this can be accomplished by adding the network configuration in `harvest-compose.yaml`.

1. After configuring the clusters in `harvest.yml`, generate `harvest-compose.yml`.

```
bin/harvest generate docker --port --output harvest-compose.yml
sed -i '/\/conf/s/^/#/g' harvest-compose.yml
```

<details><summary>harvest-compose.yml</summary>
<p>

```yaml
version: "3.7"

services:

  u2:
    image: ghcr.io/netapp/harvest:latest
    container_name: poller-u2
    restart: unless-stopped
    ports:
      - 12990:12990
    command: '--poller u2 --promPort 12990 --config /opt/harvest.yml'
    volumes:
      #      - /Users/harvest/conf:/opt/harvest/conf
      - /Users/harvest/cert:/opt/harvest/cert
      - /Users/harvest/harvest.yml:/opt/harvest.yml
```
</p>
</details>

2. Using `kompose`, convert `harvest-compose.yml` into Kubernetes resources and save them as `kub.yaml`.

```
kompose convert --file harvest-compose.yml --volumes configMap -o kub.yaml
```

<details><summary>kub.yaml</summary>
<p>

```yaml
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    kompose.cmd: kompose convert --file harvest-compose.yml --volumes configMap -o kub.yaml
    kompose.version: 1.28.0 (HEAD)
  creationTimestamp: null
  labels:
    io.kompose.service: u2
  name: u2
spec:
  ports:
    - name: "12990"
      port: 12990
      targetPort: 12990
  selector:
    io.kompose.service: u2
status:
  loadBalancer: {}

---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    kompose.cmd: kompose convert --file harvest-compose.yml --volumes configMap -o kub.yaml
    kompose.version: 1.28.0 (HEAD)
  creationTimestamp: null
  labels:
    io.kompose.service: u2
  name: u2
spec:
  replicas: 1
  selector:
    matchLabels:
      io.kompose.service: u2
  strategy:
    type: Recreate
  template:
    metadata:
      annotations:
        kompose.cmd: kompose convert --file harvest-compose.yml --volumes configMap -o kub.yaml
        kompose.version: 1.28.0 (HEAD)
      creationTimestamp: null
      labels:
        io.kompose.network/harvest-default: "true"
        io.kompose.service: u2
    spec:
      containers:
        - args:
            - --poller
            - u2
            - --promPort
            - "12990"
            - --config
            - /opt/harvest.yml
          image: ghcr.io/netapp/harvest:latest
          name: poller-u2
          ports:
            - containerPort: 12990
          resources: {}
          volumeMounts:
            - mountPath: /opt/harvest/cert
              name: u2-cm0
            - mountPath: /opt/harvest.yml
              name: u2-cm1
              subPath: harvest.yml
      restartPolicy: Always
      volumes:
        - configMap:
            name: u2-cm0
          name: u2-cm0
        - configMap:
            items:
              - key: harvest.yml
                path: harvest.yml
            name: u2-cm1
          name: u2-cm1
status: {}

---
apiVersion: v1
kind: ConfigMap
metadata:
  creationTimestamp: null
  labels:
    io.kompose.service: u2
  name: u2-cm0

---
apiVersion: v1
data:
  harvest.yml: |+
    Tools:
    Exporters:
        prometheus1:
            exporter: Prometheus
            port_range: 12990-14000
            add_meta_tags: false
    Defaults:
        use_insecure_tls: true
        prefer_zapi: true
    Pollers:

        u2:
            datacenter: u2
            addr: ADDRESS
            username: USER
            password: PASS
            collectors:
                - Rest
            exporters:
                - prometheus1

kind: ConfigMap
metadata:
  annotations:
    use-subpath: "true"
  creationTimestamp: null
  labels:
    io.kompose.service: u2
  name: u2-cm1

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  creationTimestamp: null
  name: harvest-default
spec:
  ingress:
    - from:
        - podSelector:
            matchLabels:
              io.kompose.network/harvest-default: "true"
  podSelector:
    matchLabels:
      io.kompose.network/harvest-default: "true"
```

</p>
</details>

3. Apply `kub.yaml` to k8.

```
kubectl apply --filename kub.yaml
```

4. List running pods.

```
kubectl get pods
```

<details><summary>pods</summary>
<p>

```text
NAME                  READY   STATUS    RESTARTS   AGE
u2-6864cc7dbc-v6444   1/1     Running   0          6m27s
```

</p>
</details>

#### Remove all Harvest resources from k8

```kubectl delete --filename kub.yaml```

#### Helm Chart

Generate helm charts

```
kompose convert --file harvest-compose.yml --chart --volumes configMap --out harvestchart
```