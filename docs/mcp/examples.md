# Usage Examples

This section provides practical example questions you can ask your MCP client (GitHub Copilot, Claude Desktop, etc.) when using the Harvest MCP Server.

## Getting the Best Results

### Use Harvest MCP Provided Prompt

For optimal analysis and insights, always start by setting the **Analysis Expert** prompt available in the Harvest MCP server. This prompt provides your MCP client with best practices for analyzing Harvest data.

**Access the prompt**: Most MCP clients support prompts - look for a `/mcp` command or prompts menu to select the "Analysis Expert" prompt.

### Use Advanced Language Models

Higher-capability language models provide significantly better analysis and insights:

- **Recommended**: GPT-4, Claude 3.5 Sonnet, or equivalent advanced models
- **Good**: GPT-3.5, Claude 3 Haiku for basic queries
- **Best Results**: Latest model versions with larger context windows

## Reference Questions

Here are example questions organized by complexity that work well with the Harvest MCP Server:

For more examples and community discussions about MCP usage, see: [Harvest MCP Discussion](https://github.com/NetApp/harvest/discussions/3902)

### Infrastructure Health

**"What's the overall health of my infrastructure?"**

Expected response: Comprehensive health summary showing cluster status, active alerts, capacity issues, and immediate action items with priority levels.

---

### Capacity Analysis

**"Which volumes are approaching capacity limits and need attention?"**

Expected response: List of volumes above 90% utilization with details about clusters, SVMs, growth trends, and recommended actions including timeline for capacity expansion.

---

### Performance Investigation

**"Show me systems experiencing performance issues with high latency or IOPS"**

Expected response: Analysis of volume and node performance metrics, identification of hotspots, correlation with capacity utilization, and specific recommendations for optimization.

---

### Trend Analysis

**"Analyze storage growth patterns across my clusters over the past 30 days and predict when I'll need to add capacity"**

Expected response: Detailed growth analysis by cluster and aggregate, mathematical projections of space consumption, identification of fastest-growing workloads, and recommended expansion timeline with sizing guidance.

## Integration Tips

1. **Always set the prompt first** before asking questions
2. **Use specific questions** rather than vague requests
3. **Ask follow-up questions** to dive deeper into specific areas
4. **Combine multiple areas** (e.g., "Show me capacity and performance issues together")
5. **Request different perspectives** (executive summary vs. technical details)

## MCP Clients

Common MCP clients that work with Harvest MCP Server:

- **GitHub Copilot**: Integrated in VS Code, supports MCP connections
- **Claude Desktop**: Anthropic's desktop application with MCP support
- **Custom MCP Clients**: Any application implementing the MCP standard

## Next Steps

- [Installation Guide](installation.md)