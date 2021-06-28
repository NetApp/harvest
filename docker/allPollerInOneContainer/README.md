## Docker

Build Docker Image from source code

```
docker build -f docker/allPollerInOneContainer/Dockerfile -t harvest:latest  .
```

Modify docker-compose (docker/allPollerInOneContainer/docker-compose.yml) file for port mappings and start docker containers

```
docker-compose -f docker/allPollerInOneContainer/docker-compose.yml up -d
```

Stop docker containers

```
docker-compose -f docker/allPollerInOneContainer/docker-compose.yml down
```