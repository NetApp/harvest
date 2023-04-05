# Local K8 Deployment

## Requirements
- Kompose: `v1.25` or higher https://github.com/kubernetes/kompose/

Execute below commands to run harvest artifacts in kubernetes

1. ```bin/harvest generate docker full --port --output harvest-compose.yml```
2. ```kompose convert --file harvest-compose.yml --file prom-stack.yml --out kub.yaml --volumes hostPath```
3. ```kubectl apply --filename kub.yaml```

### Stop all containers

```kubectl delete --filename kub.yaml```

## Helm Chart

Generate helm charts with below command

```
kompose convert --file harvest-compose.yml --file prom-stack.yml --chart --volumes hostPath
```