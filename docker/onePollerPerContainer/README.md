## Docker

Build Docker Image

```
docker build -f docker/onePollerPerContainer/Dockerfile -t harvest:latest  . --no-cache
```

Generate docker-compose file and save contents as docker-compose.yml
```
bin/harvest generate docker --config ./harvest.yml
```

Start docker containers

```
docker-compose -f docker-compose.yml up -d --remove-orphans
```

Stop docker containers

```
docker-compose -f docker-compose.yml down
```