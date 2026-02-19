package descriptions

const InfrastructureHealthDesc = `Think of this as your ONTAP system's health checkup - like taking vital signs before diagnosing what's wrong.
Combines multiple health indicators into a unified operational status view.
Coverage: system availability, capacity utilization, performance baselines, known failure patterns.
Output: Current status with trending indicators for operational planning.
Workflow: Excellent starting point for infrastructure analysis and assessment.

REACH FOR THIS WHEN:
- General health questions, 'any problems?', starting any troubleshooting

ANALYSIS APPROACH:
1. Start here for overall infrastructure status
2. If issues found → use get_active_alerts for detailed alert information
3. For capacity problems → use metrics_query with topk() to identify top consumers
4. For performance issues → use metrics_range_query to analyze trends over time
5. Drill down to component level (FlexGroup → constituents, aggregates → volumes)

KEY METRICS CHECKED:
- Status (*_new_status): 0=offline, 1=online
- Capacity: *_size_used_percent, *_space_used_percent
- Health: health_* indicators, error counters
- Thresholds: Volumes >95% can cause app failures, Aggregates >80% need planning attention`

const MetricsQueryDesc = `
Approach: Start with simple metric queries, then add label filters to narrow scope. Use aggregation functions (sum, avg, max) for infrastructure-wide views.
Context: Always combine with range queries to understand trends and historical patterns.
State Queries: For status metrics (*_new_status), 0 = offline, 1 = online

CRITICAL - METRIC TYPE UNDERSTANDING:
- ONTAP Metrics: Performance metrics (IOPS, latency, throughput) are PRE-CALCULATED. Use directly WITHOUT rate() functions.
  GOOD: volume_ops_total, node_latency_avg
  BAD: rate(volume_ops_total[5m]) - WRONG for ONTAP!
- StorageGrid Metrics: Often raw counters. MAY NEED rate() or increase() functions.
- Cisco Metrics: Mixed types. Use get_metric_description to understand before querying.
- Discovery: Always use get_metric_description and search_metrics to understand any metric before querying.
- Example: Database team reports slowness → check node_latency_avg{cluster='prod'} to see if storage latency is the culprit

COMMON QUERY PATTERNS:
- High utilization: volume_size_used_percent > 90
- Top consumers: topk(10, volume_size_used_percent)
- Offline components: node_new_status != 1
- Cluster-wide summary: sum by (cluster) (volume_size_total)
- Filter by label: volume_ops_total{cluster="prod",node="node1"}`

const MetricsRangeQueryDesc = `Use for trend analysis, growth patterns, historical baselines, and identifying when problems started.

USE THIS TOOL FOR:
- Capacity Planning: Analyze growth trends with *_size_used_percent over 7-30 days
- Performance Baselines: Track *_ops, *_latency, *_data over time to identify anomalies
- Root Cause Analysis: Historical trends to understand when issues started
- Time-to-Full Projections: Calculate future capacity exhaustion based on growth rates
- Storage problems usually build over time - this shows you if that 85% full volume got there slowly (normal growth) or suddenly (potential issue)

PARAMETER RECOMMENDATIONS:
- Use smaller steps (5m, 15m) for detailed recent trends, larger steps (1h, 6h) for longer-term patterns

ANALYSIS PATTERNS:
- Growth trend: volume_size_total - start='now-30d', end='now', step='6h'
- Performance baseline: volume_latency_avg - start='now-7d', end='now', step='1h'
- Identify when problem started: Compare current vs historical averages`

const SearchMetricsDesc = `Search for metrics by name, description, or object type using a pattern.
Use this for discovery when you don't know which metric to query.

USE THIS TOOL FIRST WHEN:
- You don't know which metric to query (discovery phase)
- User asks about a concept: 'storage capacity', 'network performance', 'latency'
- User mentions an ONTAP object: 'volumes', 'aggregates', 'LUNs', 'qtrees', 'snapshots'
- Need to find the right metric before querying with metrics_query
- Perfect when someone says 'our snapshots aren't working' but you're not sure which snapshot-related metrics exist

SEARCH STRATEGIES:
- Concept search: pattern='latency' → finds all latency-related metrics
- Object search: pattern='volume' → finds all volume metrics
- Specific feature: pattern='snapshot' → finds snapshot-related metrics
- Wildcard regex: pattern='.*space.*used.*' → complex pattern matching

AFTER THIS TOOL:
1. Found the right metric → use metrics_query to get current values
2. Understand metric better → use get_metric_description for full details
3. Browse all available → use list_metrics to see everything`

const GetMetricDescriptionDesc = `Get description and metadata for a specific metric by name.
Provides detailed information about what the metric measures, its units, and how to use it.

USE THIS TOOL WHEN:
- You found a metric name but don't understand what it measures
- Need to verify metric type before using rate() or other functions
- Want to know the units (bytes, milliseconds, percentage, etc.)
- Understanding if it's a counter, gauge, or pre-calculated value

IMPORTANT FOR QUERY CONSTRUCTION:
- ONTAP metrics are typically pre-calculated - no rate() needed
- StorageGrid metrics may be raw counters - may need rate()
- Always check description to understand metric semantics

AFTER THIS TOOL:
- Use metrics_query with the correct syntax (with or without rate())
- Apply appropriate filters based on available labels`

const ListMetricsDesc = `List all available metrics from Prometheus or VictoriaMetrics with advanced filtering and optional descriptions.
When 'match' or 'matches' filters are applied, metric descriptions are automatically included.
Use: 1) 'match' for simple/regex patterns, 2) 'matches' for efficient server-side label matchers

USE THIS TOOL WHEN:
- Exploring what metrics are available in your system
- You need to browse all metrics (no filter) or filter by pattern
- Verifying that expected metrics are being collected
- Discovery phase before search_metrics (broader view)

FILTERING OPTIONS:
- No filter: Returns all available metric names (can be large!)
- match='volume': Simple string matching in metric names
- match='.*volume.*(latency|data|throughput).*': Regex pattern for complex filtering
- matches='{__name__=~"volume.*latency.*"},{__name__=~"volume.*data$"}': Comma-separated PromQL label matchers for efficient server-side filtering

TIP: For large systems, use 'match' with regex alternation (e.g., '.*volume.*(latency|data|throughput).*') or search_metrics for targeted results.`

const GetActiveAlertsDesc = `Get active alerts from Prometheus or VictoriaMetrics with summary by severity level.
Provides grouped view of critical, warning, and info alerts for quick operational assessment.

USE THIS TOOL WHEN:
- infrastructure_health reports issues and you need details
- User asks 'what alerts are firing?' or 'any active problems?'
- Need to prioritize which issues to address first (by severity)
- Starting incident investigation or troubleshooting session

OUTPUT FORMAT:
- Summary by severity: Critical, Warning, Info
- Full alert details with labels, annotations, and state
- Actionable information for escalation or remediation

AFTER THIS TOOL:
- Use metrics_query to investigate specific components mentioned in alerts
- Use metrics_range_query to see when the alert condition started
- For alert rule management, use list_alert_rules, create_alert_rule, etc.`

const ListLabelValuesDesc = `Get all available values for a specific label (e.g., cluster names, node names, volume names) with optional regex filtering.
Useful for discovering what infrastructure components exist and for building targeted queries.

USE THIS TOOL WHEN:
- Discovering what clusters, nodes, volumes, or other components exist
- Building filters for metrics_query (need to know valid label values)
- User asks 'what clusters do we have?' or 'show me all volumes'
- Verifying naming conventions or finding specific component names

COMMON LABELS TO QUERY:
- label='cluster' → All cluster names being monitored
- label='node' → All node names across clusters
- label='volume' → All volume names (can be large!)
- label='aggr' → All aggregate names
- label='svm' → All Storage Virtual Machine (SVM) names

TIP: Use 'match' parameter to filter results with regex patterns (e.g., match='prod.*' for production clusters)`

const ListAllLabelNamesDesc = `Get all available label names (dimensions) that can be used to filter metrics in Prometheus or VictoriaMetrics.
Shows what labels exist across all metrics for building filters and group-by queries.

USE THIS TOOL WHEN:
- Understanding what dimensions are available for filtering
- Building complex queries with multiple label filters
- Discovering what labels can be used in group-by operations
- Learning the data model structure

COMMON LABEL NAMES IN HARVEST:
- cluster, node, aggr, volume, svm - Infrastructure components
- datacenter - Logical grouping of clusters
- type, style - Object classification labels

AFTER THIS TOOL:
- Use list_label_values to see what values exist for a specific label
- Build filtered queries: metric_name{cluster='X',node='Y'}`

const GetResponseFormatTemplateDesc = `Get comprehensive multi-audience response format template for detailed infrastructure analysis.
Use when user requests comprehensive reports, detailed analysis, or management-ready output.

USE THIS TOOL WHEN:
- User asks for "detailed report" or "comprehensive analysis"
- User asks for "management summary" or "executive report"
- Analysis needs to be shared with multiple audiences (executives, engineers, operators)
- User wants structured output with business impact and technical details
- Creating documentation or formal reports

TEMPLATE INCLUDES:
- Executive Summary: Business impact, risk assessment, high-level recommendations
- Technical Analysis: Detailed metrics, root cause, configuration recommendations
- Operational Actions: Immediate tasks, preventive measures, escalation triggers
- Proactive Monitoring: Key metrics to watch, thresholds, alert recommendations

NOT NEEDED FOR:
- Quick status checks (use infrastructure_health)
- Simple metric queries
- Ad-hoc troubleshooting conversations
- Informal analysis

The template provides structure for comprehensive reporting while core response principles (impact/urgency/actions) are automatically applied to all responses.`

const CoreResponseFormat = `

## Core Response Principles

- **Severity**: Critical/High/Medium/Low with business impact
- **Timeline**: Immediate (hours) / Soon (days) / Planning (weeks)  
- **Next Steps**: What to do, how to verify, when to escalate
- **Key Metrics**: What to monitor going forward`

const Instructions = `You are a NetApp storage infrastructure expert specializing in ONTAP, StorageGRID and Cisco Switch analysis using Harvest metrics.

## Your Role
Help users understand storage health, diagnose problems, and make informed decisions about their NetApp infrastructure. You have access to real-time metrics and can provide actionable insights for storage administrators, engineers, and management.

## When Users Ask Questions, Always Consider:
- **Business Impact**: Will this affect applications or users?
- **Urgency**: Does this need immediate attention or can it wait?  
- **Next Steps**: What specific actions should they take?
- **Prevention**: How can they avoid this in the future?

## Key ONTAP/NetApp Knowledge:
- ONTAP metrics are pre-calculated (never use rate() functions)
- Volume capacity >95% = urgent, >85% = needs attention
- Aggregate capacity >80% = planning required
- Status metrics: 0=offline, 1=online
- FlexGroups contain constituents, aggregates contain volumes

## Response Guidelines:
Always classify findings by severity:
- **Critical**: Service disruption risk, immediate action needed
- **High**: Performance impact, action needed soon  
- **Medium**: Potential issues, proactive attention recommended
- **Low**: Optimization opportunities

Always specify urgency timeline:
- **Immediate** (1-2 hours): Offline components, volumes at 100%
- **Same Day**: Volumes >95%, critical alerts
- **This Week**: Aggregates >85%, performance issues  
- **Planning**: Growth trends, optimization opportunities

## Tool Selection Strategy:
1. **Start with infrastructure_health** for general "how are things?" questions
2. **Use search_metrics** when you need to find the right metric first
3. **Use metrics_query** for current values and real-time status
4. **Use metrics_range_query** for trends and historical analysis
5. **Use get_active_alerts** when infrastructure_health shows problems

Be conversational and helpful while providing technically accurate information.`
