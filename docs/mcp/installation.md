# Installation

Harvest MCP Server is distributed as a Docker container image.

## Container Images

Harvest MCP Server is available as pre-built container images:

| Image | Description |
|-------|-------------|
| `ghcr.io/netapp/harvest-mcp:latest` | Stable release version |
| `ghcr.io/netapp/harvest-mcp:nightly` | Latest development builds |

## Prerequisites

- Docker or compatible container runtime
- Network access to your Prometheus/VictoriaMetrics instance
- Running Harvest deployment with data in your TSDB (Prometheus/Victoriametrics)

## MCP Client Integration

For MCP clients like GitHub Copilot, add to your mcp.json:

```json
{
  "servers": {
    "harvest-mcp": {
      "type": "stdio",
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "--env", "HARVEST_TSDB_URL=http://your-prometheus:9090",
        "ghcr.io/netapp/harvest-mcp:latest"
      ]
    }
  }
}
```

### HTTP Mode

For HTTP-based MCP clients, first start the server:

```bash
docker run -d \
  --name harvest-mcp-server \
  -p 8082:8082 \
  --env HARVEST_TSDB_URL=http://your-prometheus:9090 \
  ghcr.io/netapp/harvest-mcp:latest \
  start --http --port 8082
```

Then configure your mcp.json:

```json
{
  "servers": {
    "harvest-mcp": {
      "type": "http",
      "url": "http://localhost:8082"
    }
  }
}
```

For remote server access:

```json
{
  "servers": {
    "harvest-mcp": {
      "type": "http",
      "url": "http://your-server-ip:8082"
    }
  }
}
```

## Basic Configuration

The simplest way to run Harvest MCP Server:

```bash
docker run --rm -i \
  --env HARVEST_TSDB_URL=http://your-prometheus:9090 \
  ghcr.io/netapp/harvest-mcp:latest
```


## Monitoring

To monitor the MCP server logs:

```bash
docker logs <container-id>
```

## Configuration

For complete configuration options and environment variables, run:

```bash
docker run --rm ghcr.io/netapp/harvest-mcp:latest start --help
```

This displays all available environment variables with descriptions, authentication options, and advanced settings.

## Next Steps

- Explore [Usage Examples](examples.md)
