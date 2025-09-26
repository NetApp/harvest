# Harvest MCP Server

The Harvest MCP (Model Context Protocol) Server provides MCP clients (like GitHub Copilot, Claude Desktop, and other LLMs) with intelligent access to your infrastructure monitoring data collected by Harvest from ONTAP, StorageGRID, and Cisco systems.

## What is MCP?

The Model Context Protocol (MCP) is an open standard that enables interactions between MCP clients and external data sources.
Harvest MCP Server acts as a bridge between MCP clients and your infrastructure monitoring data stored in Prometheus or
VictoriaMetrics, allowing for intelligent analysis and insights.

## What You Can Ask

Transform your monitoring data into actionable insights through natural language questions:

**Simple Health Checks**

- "What's the overall health of my infrastructure?"
- "Are there any active alerts I should know about?"

**Capacity Analysis**

- "Which volumes are running out of space?"
- "Show me the top 5 volumes by utilization"

**Performance Investigation**

- "Which systems are experiencing high latency?"
- "Find volumes with performance issues"

**Advanced Analytics**

- "Analyze storage growth trends over the past month"
- "Show me performance bottlenecks across my clusters"

## Architecture

The Harvest MCP Server operates as a lightweight service that:

1. Connects to your existing Prometheus/VictoriaMetrics instance containing Harvest data
2. Provides a standardized MCP interface for MCP clients (GitHub Copilot, Claude Desktop, etc.)
3. Enables natural language queries against your infrastructure data
4. Returns structured insights suitable for decision making

```mermaid
graph LR
    A[MCP Client<br/>GitHub Copilot<br/>Claude Desktop] --> B[Harvest MCP Server]
    B --> C[Prometheus/VictoriaMetrics]
    C --> D[Harvest Pollers]
    D --> F[ONTAP Clusters]
    D --> G[StorageGRID]  
    D --> H[Cisco Switches]
```

## Prerequisites

- Running Harvest
- Prometheus or VictoriaMetrics instance with Harvest data
- Docker environment for running the MCP server
- Network connectivity from MCP server to your TSDB

For information about Harvest deployment and configuration, see:

- [Harvest Concepts](../concepts.md)
- [Installation Overview](../install/overview.md)
- [System Requirements](../system-requirements.md)

## Next Steps

- [Install the MCP Server](installation.md)
- [Configure Environment Variables](configuration.md)
- Try the [Usage Examples](examples.md)