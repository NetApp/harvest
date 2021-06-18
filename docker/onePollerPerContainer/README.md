## Docker

Build Docker Image from source code

```
docker build -f docker/onePollerPerContainer/Dockerfile -t harvest:latest  . --no-cache
```

Generate docker-compose file and save contents as docker-compose.yml
```
make build
bin/harvest generate docker --config ./harvest.yml
```

Start docker containers

```
docker-compose -f docker-compose.yml up -d
```
