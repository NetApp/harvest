# Installation

Harvest MCP Server is distributed as a Docker container image, or you can build it from source.

## Container Images

Harvest MCP Server is available as pre-built container images:

| Image | Description |
|-------|-------------|
| `ghcr.io/netapp/harvest-mcp:latest` | Stable release version |
| `ghcr.io/netapp/harvest-mcp:nightly` | Latest development builds |

## MCP Client Integration

### Stdio Mode

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
  start --http --port 8082 --host 0.0.0.0
```

If you only want to bind to localhost, omit the `--host` option.

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

## Building from Source

### Prerequisites

- Go(check `.harvest.env` in the repository root for the exact required version)
- Git
- Make
- Docker (optional, for building Docker images)

### Clone the Repository

First, clone the Harvest repository:

```bash
git clone https://github.com/NetApp/harvest.git
cd harvest
```

### Build Docker Image

Build your own Docker image from source:

```bash
# Navigate to the mcp directory
cd mcp

# Build the Docker image using make (creates harvest-mcp:latest by default)
make docker-build

# Or specify a custom tag
make docker-build DOCKER_TAG=harvest-mcp:local
```

Alternatively, build directly with Docker:

```bash
# From the harvest repository root
docker build -f mcp/Dockerfile -t harvest-mcp:local .
```

### Running the Built Docker Image

After building, use your local image in your MCP client configuration. See [MCP Client Integration](#mcp-client-integration) above for configuration examples - just replace `ghcr.io/netapp/harvest-mcp:latest` with your local image tag (e.g., `harvest-mcp:local`).

### Build Native Binary

Create a standalone binary package:

```bash
# Navigate to the mcp directory
cd mcp

# build for specific platforms:
GOOS=linux GOARCH=amd64 make package    # Linux AMD64
GOOS=darwin GOARCH=arm64 make package   # macOS ARM64 (Apple Silicon)

# Creates: dist/harvest-mcp-<version>-<release>_<os>_<arch>.tar.gz
```

### Running the Native Binary

After extracting the package, configure it in mcp.json for MCP clients.

#### Configure in mcp.json

For MCP clients like GitHub Copilot, add to your mcp.json:

```json
{
  "servers": {
    "harvest-mcp": {
      "type": "stdio",
      "command": "/path/to/harvest-mcp/bin/harvest-mcp",
      "args": ["start"],
      "env": {
        "HARVEST_TSDB_URL": "http://your-prometheus:9090"
      }
    }
  }
}
```

For HTTP mode, start the server first:

```bash
# Start the server
HARVEST_TSDB_URL=http://your-prometheus:9090 /path/to/harvest-mcp/bin/harvest-mcp start --http --port 8082
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

## Logs

To view the MCP server logs:

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
