Harvest is written in Go, which means it runs on recent Linux systems.
It also runs on Macs for development.

Hardware requirements depend on how many clusters you monitor and the number of metrics you chose to collect.
With the default configuration, when monitoring 10 clusters, we recommend:

- CPU: 2 cores
- Memory: 1 GB
- Disk: 500 MB (mostly used by log files)

**Note:** These CPU, memory, and disk requirements are just for Harvest and do not include Prometheus, InfluxDB, Grafana.

Harvest is compatible with:

- Prometheus: `2.33` or higher
- InfluxDB: `v2`
- Grafana: `8.1.X` or higher
- Docker: `20.10.0` or higher and compatible Docker Compose