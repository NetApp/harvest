package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mcp-server/cmd/descriptions"
	"mcp-server/cmd/loader"
	"mcp-server/cmd/version"
	"mcp-server/pkg/auth"
	"mcp-server/pkg/helper"
	"mcp-server/pkg/mcptypes"
	"mcp-server/pkg/rules"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/spf13/cobra"
)

// setupLogger configures the logger based on environment variables
func setupLogger() *slog.Logger {
	level := slog.LevelInfo // default level

	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		switch strings.ToUpper(envLevel) {
		case "DEBUG":
			level = slog.LevelDebug
		case "INFO":
			level = slog.LevelInfo
		case "WARN", "WARNING":
			level = slog.LevelWarn
		case "ERROR":
			level = slog.LevelError
		}
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				source.File = filepath.Base(source.File)
			}
			return a
		},
	})

	return slog.New(handler)
}

var logger = setupLogger()

var metricDescriptions map[string]string
var ruleManager *rules.RuleManager
var tsdbConfig auth.TSDBConfig

const (
	AppName = "harvest-mcp"
)

const envVarHelpText = `Required Environment Variables:
  HARVEST_TSDB_URL              URL of your Prometheus or VictoriaMetrics server
                                Example: http://localhost:9090
                                For NABox4: Enable Victoria Metrics guest access and use https://nabox_url/vm

Optional Environment Variables (Authentication):
  HARVEST_TSDB_AUTH_TYPE        Authentication type: none, basic, cert (default: none)
  HARVEST_TSDB_USERNAME         Username for basic authentication
  HARVEST_TSDB_PASSWORD         Password for basic authentication
  HARVEST_TSDB_CERT_FILE        Path to client certificate file (for cert auth)
  HARVEST_TSDB_KEY_FILE         Path to client private key file (for cert auth)  
  HARVEST_TSDB_CA_FILE          Path to CA certificate file (optional, for cert auth)

Optional Environment Variables (Rules):
HARVEST_RULES_PATH              Path to directory containing alert_rules.yml and ems_alert_rules.yml


Optional Environment Variables (TLS):
  HARVEST_TSDB_TLS_INSECURE     Skip TLS certificate verification (true/false, default: false)
                                WARNING: Only use for development/testing

Optional Environment Variables (Timeout):
  HARVEST_TSDB_TIMEOUT          Request timeout duration (e.g., '30s', '1m', '90s', default: 30s)

Optional Environment Variables (Logging):
  LOG_LEVEL                 Log level: DEBUG, INFO, WARN, ERROR (default: INFO)

Transport Options:
  --http                        Enable HTTP transport (default: stdio)
  --port                        Port for HTTP transport (default: 8080)
  --host                        Host for HTTP transport (default: localhost)

Prometheus Reload Control:
  HARVEST_TSDB_AUTO_RELOAD      Enable automatic reload after rule changes: true/false (default: true)
  HARVEST_TSDB_AUTO_RELOAD_URL  URL called to reload rules (default: <HARVEST_TSDB_URL>/-/reload)

Examples:
  # Start with stdio transport (default)
  HARVEST_TSDB_URL=http://localhost:9090 ./bin/harvest-mcp start
  
  # Start with HTTP transport
  HARVEST_TSDB_URL=http://localhost:9090 ./bin/harvest-mcp start --http --port 8080
  
  # Start with basic authentication
  HARVEST_TSDB_URL=http://localhost:9090 \
  HARVEST_TSDB_AUTH_TYPE=basic \
  HARVEST_TSDB_USERNAME=admin \
  HARVEST_TSDB_PASSWORD=secret \
  ./bin/harvest-mcp start
  
  # Start with certificate authentication
  HARVEST_TSDB_URL=https://localhost:9090 \
  HARVEST_TSDB_AUTH_TYPE=cert \
  HARVEST_TSDB_CERT_FILE=/path/to/client.crt \
  HARVEST_TSDB_KEY_FILE=/path/to/client.key \
  HARVEST_TSDB_CA_FILE=/path/to/ca.crt \
  ./bin/harvest-mcp start
  
  # Start with basic auth and insecure TLS (for self-signed certs)
  HARVEST_TSDB_URL=https://localhost:9090 \
  HARVEST_TSDB_AUTH_TYPE=basic \
  HARVEST_TSDB_USERNAME=admin \
  HARVEST_TSDB_PASSWORD=secret \
  HARVEST_TSDB_TLS_INSECURE=true \
  ./bin/harvest-mcp start --http
  
  # Start with no auth but insecure TLS (for development)
  HARVEST_TSDB_URL=https://localhost:9090 \
  HARVEST_TSDB_TLS_INSECURE=true \
  ./bin/harvest-mcp start
  
  # Start with debug logging
  HARVEST_TSDB_URL=http://localhost:9090 \
  LOG_LEVEL=DEBUG \
  ./bin/harvest-mcp start`

func handlePrometheusError(err error, operation string) *mcp.CallToolResult {
	logger.Error(operation+" failed", slogx.Err(err))
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%s failed: %v", operation, err)},
		},
		IsError: true,
	}
}

func handleValidationError(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
		IsError: true,
	}
}

func makePrometheusAPICall(config auth.TSDBConfig, endpoint string) ([]byte, error) {
	fullURL := config.URL + endpoint

	logger.Debug("Making Prometheus API call", slog.String("url", fullURL))

	resp, err := auth.MakeRequest(config, fullURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func formatDataResponse(data any) (*mcp.CallToolResult, any, error) {
	content, err := formatJSONResponse(data)
	if err != nil {
		logger.Error("failed to format data response", slogx.Err(err), slog.String("data_type", fmt.Sprintf("%T", data)))
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error formatting response: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{Content: content}, nil, nil
}

// resolveTSDBConfig returns the appropriate TSDBConfig to use for a request
// If override parameters are provided, creates a new config otherwise uses default
func resolveTSDBConfig(override *mcptypes.TSDBOverride) auth.TSDBConfig {
	if override == nil || override.URL == "" {
		return tsdbConfig
	}

	logger.Debug("using per-request TSDB URL override",
		slog.String("url", override.URL),
		slog.Bool("custom_auth", override.Username != ""))

	config := auth.TSDBConfig{
		URL:       override.URL,
		Timeout:   tsdbConfig.Timeout,
		RulesPath: tsdbConfig.RulesPath,
		Auth: auth.Config{
			Type:            auth.None,
			InsecureSkipTLS: tsdbConfig.Auth.InsecureSkipTLS,
		},
	}

	if override.Username != "" && override.Password != "" {
		config.Auth.Type = auth.Basic
		config.Auth.Username = override.Username
		config.Auth.Password = override.Password
	}

	return config
}

func addTool[T any](server *mcp.Server, name, description string, handler func(context.Context, *mcp.CallToolRequest, T) (*mcp.CallToolResult, any, error)) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        name,
		Description: description,
	}, handler)
}

func executeTSDBQuery(config auth.TSDBConfig, queryURL string, params url.Values) (*mcptypes.MetricsResponse, error) {
	fullURL := fmt.Sprintf("%s?%s", queryURL, params.Encode())
	resp, err := auth.MakeRequest(config, fullURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var promResp mcptypes.MetricsResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if promResp.Status != "success" {
		return nil, fmt.Errorf("prometheus error: %s - %s", promResp.ErrorType, promResp.Error)
	}

	return &promResp, nil
}

func formatJSONResponse(data any) ([]mcp.Content, error) {
	resultJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logger.Error("failed to marshal JSON response", slogx.Err(err), slog.String("data_type", fmt.Sprintf("%T", data)))
		return nil, err
	}

	return []mcp.Content{
		&mcp.TextContent{Text: string(resultJSON)},
	}, nil
}

// MetricsQuery executes a time series database instant query
func MetricsQuery(_ context.Context, _ *mcp.CallToolRequest, args mcptypes.QueryRequest) (*mcp.CallToolResult, any, error) {
	if err := helper.ValidateQueryArgs(args.Query); err != nil {
		return handleValidationError(err.Error()), nil, err
	}

	config := resolveTSDBConfig(args.TSDBOverride)

	queryURL := config.URL + "/api/v1/query"
	urlValues := url.Values{}
	urlValues.Set("query", args.Query)

	logger.Debug("Executing Prometheus instant query",
		slog.String("query", args.Query),
		slog.String("url", queryURL))

	promResp, err := executeTSDBQuery(config, queryURL, urlValues)
	if err != nil {
		logger.Error("Prometheus query failed", slogx.Err(err), slog.String("query", helper.TruncateString(args.Query, 100)))
		return handlePrometheusError(err, "Prometheus query"), nil, nil
	}

	return formatDataResponse(promResp)
}

// MetricsRangeQuery executes a time series database range query
func MetricsRangeQuery(_ context.Context, _ *mcp.CallToolRequest, args mcptypes.RangeQueryRequest) (*mcp.CallToolResult, any, error) {
	if err := helper.ValidateRangeQueryArgs(args.Query, args.Start, args.End, args.Step); err != nil {
		return handleValidationError(err.Error()), nil, err
	}

	config := resolveTSDBConfig(args.TSDBOverride)

	queryURL := config.URL + "/api/v1/query_range"
	urlValues := url.Values{}
	urlValues.Set("query", args.Query)
	urlValues.Set("start", args.Start)
	urlValues.Set("end", args.End)
	urlValues.Set("step", args.Step)

	logger.Debug("Executing Prometheus range query",
		slog.String("query", args.Query),
		slog.String("start", args.Start),
		slog.String("end", args.End),
		slog.String("step", args.Step),
		slog.String("url", queryURL))

	promResp, err := executeTSDBQuery(config, queryURL, urlValues)
	if err != nil {
		logger.Error("Prometheus range query failed", slogx.Err(err), slog.String("query", helper.TruncateString(args.Query, 100)))
		return handlePrometheusError(err, "Prometheus range query"), nil, nil
	}

	return formatDataResponse(promResp)
}

func filterStrings(items []string, pattern string) []string {
	if pattern == "" {
		return items
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		logger.Debug("Pattern is not valid regex, using string matching", slog.String("pattern", pattern), slogx.Err(err))
		var filtered []string
		lowerPattern := strings.ToLower(pattern)
		for _, item := range items {
			if strings.Contains(strings.ToLower(item), lowerPattern) {
				filtered = append(filtered, item)
			}
		}
		return filtered
	}

	// Use regex matching
	var filtered []string
	for _, item := range items {
		if regex.MatchString(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// makePrometheusAPICallWithMatches performs API call with optional label matchers
func makePrometheusAPICallWithMatches(config auth.TSDBConfig, endpoint string, matches []string) ([]byte, error) {
	var fullURL string
	if len(matches) > 0 {
		// Build URL with matches parameter
		params := url.Values{}
		for _, match := range matches {
			params.Add("match[]", match)
		}
		fullURL = config.URL + endpoint + "?" + params.Encode()
	} else {
		fullURL = config.URL + endpoint
	}

	logger.Debug("Making Prometheus API call with matches",
		slog.String("url", fullURL),
		slog.Any("matches", matches))

	resp, err := auth.MakeRequest(config, fullURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// ListMetrics lists available metrics from time series database
func ListMetrics(_ context.Context, _ *mcp.CallToolRequest, args mcptypes.ListMetricsRequest) (*mcp.CallToolResult, any, error) {
	var body []byte
	var err error

	config := resolveTSDBConfig(args.TSDBOverride)

	if len(args.Matches) > 0 {
		logger.Debug("Using server-side filtering with matches", slog.Any("matches", args.Matches))
		body, err = makePrometheusAPICallWithMatches(config, "/api/v1/label/__name__/values", args.Matches)
	} else {
		body, err = makePrometheusAPICall(config, "/api/v1/label/__name__/values")
	}

	if err != nil {
		return handlePrometheusError(err, "query Prometheus metrics"), nil, nil
	}

	var promResp mcptypes.LabelsResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return handlePrometheusError(err, "parse metrics response"), nil, nil
	}

	if promResp.Status != "success" {
		return handlePrometheusError(errors.New(promResp.Error), "Prometheus metrics query"), nil, nil
	}

	metrics := promResp.Data
	if args.Match != "" && len(args.Matches) == 0 {
		metrics = filterStrings(metrics, args.Match)
	}

	// Include descriptions only when filtering is applied to limit response size
	includeDescriptions := (args.Match != "" || len(args.Matches) > 0) && len(metricDescriptions) > 0

	metricsArray := make([]map[string]any, 0, len(metrics))
	for _, metric := range metrics {
		metricInfo := map[string]any{"name": metric}
		if includeDescriptions {
			if description, found := metricDescriptions[metric]; found {
				metricInfo["description"] = description
			}
		}
		metricsArray = append(metricsArray, metricInfo)
	}

	response := map[string]any{
		"status": "success",
		"data": map[string]any{
			"total_count": len(metrics),
			"filtering": map[string]any{
				"server_side_matches": len(args.Matches) > 0,
				"client_side_pattern": args.Match != "" && len(args.Matches) == 0,
				"pattern_used":        args.Match,
				"matches_used":        args.Matches,
			},
			"descriptions_included": includeDescriptions,
			"metrics":               metricsArray,
		},
	}

	return formatDataResponse(response)
}

// ListLabelValues lists available values for a specific label from Prometheus
func ListLabelValues(_ context.Context, _ *mcp.CallToolRequest, args mcptypes.ListLabelValuesRequest) (*mcp.CallToolResult, any, error) {
	if args.Label == "" {
		return handleValidationError("label parameter is required"), nil, errors.New("label parameter is required")
	}

	config := resolveTSDBConfig(args.TSDBOverride)

	body, err := makePrometheusAPICall(config, "/api/v1/label/"+args.Label+"/values")
	if err != nil {
		logger.Error("Failed to query Prometheus label values", slogx.Err(err), slog.String("label", args.Label))
		return handlePrometheusError(err, fmt.Sprintf("query label values for '%s'", args.Label)), nil, nil
	}

	var promResp mcptypes.LabelsResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return handlePrometheusError(err, "parse label values response"), nil, nil
	}

	if promResp.Status != "success" {
		return handlePrometheusError(errors.New(promResp.Error), fmt.Sprintf("Prometheus label values query for '%s'", args.Label)), nil, nil
	}

	values := promResp.Data
	if args.Match != "" {
		values = filterStrings(values, args.Match)
	}

	response := map[string]any{
		"status": "success",
		"data": map[string]any{
			"label_name":   args.Label,
			"label_values": values,
			"total_count":  len(values),
		},
	}

	return formatDataResponse(response)
}

// ListAllLabelNames lists all available label names (dimensions) from Prometheus
func ListAllLabelNames(_ context.Context, _ *mcp.CallToolRequest, args mcptypes.ListAllLabelNamesRequest) (*mcp.CallToolResult, any, error) {
	config := resolveTSDBConfig(args.TSDBOverride)
	body, err := makePrometheusAPICall(config, "/api/v1/labels")
	if err != nil {
		return handlePrometheusError(err, "query Prometheus label names"), nil, nil
	}

	var promResp mcptypes.LabelsResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return handlePrometheusError(err, "parse label names response"), nil, nil
	}

	if promResp.Status != "success" {
		return handlePrometheusError(errors.New(promResp.Error), "Prometheus label names query"), nil, nil
	}

	response := map[string]any{
		"status": "success",
		"data": map[string]any{
			"label_names": promResp.Data,
			"total_count": len(promResp.Data),
		},
	}

	return formatDataResponse(response)
}

func GetActiveAlerts(_ context.Context, _ *mcp.CallToolRequest, args mcptypes.GetActiveAlertsRequest) (*mcp.CallToolResult, any, error) {
	config := resolveTSDBConfig(args.TSDBOverride)
	queryURL := config.URL + "/api/v1/alerts"

	resp, err := auth.MakeRequest(config, queryURL)
	if err != nil {
		logger.Error("Failed to query Prometheus alerts", slogx.Err(err))
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to query Prometheus alerts: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to read alerts response: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	var promResp mcptypes.AlertsResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to parse alerts response: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	alertReport := "## Prometheus Active Alerts\n\n"

	alerts := promResp.Data.Alerts
	if len(alerts) == 0 {
		alertReport += "✅ **No active alerts found**\n\n"
	} else {
		alertReport += fmt.Sprintf("🚨 **%d active alerts found:**\n\n", len(alerts))

		// Group alerts by severity
		criticalCount, warningCount, infoCount := countAlertsBySeverity(alerts)

		if criticalCount > 0 {
			alertReport += fmt.Sprintf("🔴 **Critical**: %d alerts\n", criticalCount)
		}
		if warningCount > 0 {
			alertReport += fmt.Sprintf("🟡 **Warning**: %d alerts\n", warningCount)
		}
		if infoCount > 0 {
			alertReport += fmt.Sprintf("🔵 **Info**: %d alerts\n", infoCount)
		}
		otherCount := len(alerts) - (criticalCount + warningCount + infoCount)
		if otherCount > 0 {
			alertReport += fmt.Sprintf("⚪ **Other**: %d alerts\n", otherCount)
		}
	}

	resultJSON, err := json.MarshalIndent(promResp, "", "  ")
	if err != nil {
		alertReport += "\n**Error formatting detailed response**"
	} else {
		alertReport += "\n**Detailed Response:**\n```json\n" + string(resultJSON) + "\n```"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: alertReport},
		},
	}, nil, nil
}

func countAlertsBySeverity(alerts []any) (int, int, int) {
	critical, warning, info := 0, 0, 0
	for _, alert := range alerts {
		if alertMap, ok := alert.(map[string]any); ok {
			if labels, ok := alertMap["labels"].(map[string]any); ok {
				if severity, ok := labels["severity"].(string); ok {
					switch strings.ToLower(severity) {
					case "critical":
						critical++
					case "warning":
						warning++
					case "info":
						info++
					}
				}
			}
		}
	}
	return critical, warning, info
}

func InfrastructureHealth(_ context.Context, _ *mcp.CallToolRequest, args mcptypes.InfrastructureHealthRequest) (*mcp.CallToolResult, any, error) {
	healthReport := strings.Builder{}
	var report string
	healthReport.WriteString("## ONTAP Infrastructure Health Report\n\n")
	issuesFound := false

	// Health checks to perform
	healthChecks := []struct {
		name        string
		query       string
		description string
		critical    bool
	}{
		{"Cluster Status", "cluster_new_status != 1", "Clusters not in healthy state", true},
		{"Node Status", "node_new_status != 1", "Nodes not online", true},
		{"Aggregate Status", "aggr_new_status != 1", "Aggregates not online", true},
		{"Volume Status", "volume_new_status != 1", "Volumes not online", false},
		{"Full Volumes", "volume_size_used_percent == 100", "Volumes at 100% capacity", true},
		{"High Volume Usage", "volume_size_used_percent > 95", "Volumes over 95% capacity", false},
		{"High Aggregate Usage", "aggr_inode_used_percent > 90", "Aggregates over 90% capacity", false},
		{"Health Alerts", "{__name__=~\"health_.*\"}", "Active health alerts", true},
	}

	config := resolveTSDBConfig(args.TSDBOverride)

	for _, check := range healthChecks {
		queryURL := config.URL + "/api/v1/query"
		urlValues := url.Values{}
		urlValues.Set("query", check.query)

		// Debug logging for infrastructure health check query
		logger.Debug("Executing infrastructure health check",
			slog.String("check_name", check.name),
			slog.String("query", check.query),
			slog.String("url", queryURL))

		promResp, err := executeTSDBQuery(config, queryURL, urlValues)
		if err != nil {
			fmt.Fprintf(&healthReport, "❌ **%s**: Error querying - %v\n", check.name, err)
			continue
		}

		// Check if there are any results
		if resultSlice, ok := promResp.Data.Result.([]any); ok && len(resultSlice) > 0 {
			issuesFound = true
			icon := "⚠️"
			if check.critical {
				icon = "🚨"
			}

			fmt.Fprintf(&healthReport, "%s **%s**: %d issues found - %s\n", icon, check.name, len(resultSlice), check.description) //nolint:gosec

			// Add details if requested
			if args.IncludeDetails {
				healthReport.WriteString("   Details:\n")
				for i, result := range resultSlice {
					if i >= 5 { // Limit to first 5 for readability
						fmt.Fprintf(&healthReport, "   ... and %d more\n", len(resultSlice)-5) //nolint:gosec
						break
					}
					if resultMap, ok := result.(map[string]any); ok {
						if metric, ok := resultMap["metric"].(map[string]any); ok {
							name := extractIdentifiers(metric)
							fmt.Fprintf(&healthReport, "   - %s\n", name) //nolint:gosec
						}
					}
				}
			}
		} else {
			fmt.Fprintf(&healthReport, "✅ **%s**: No issues found\n", check.name)
		}
	}

	// Summary
	if issuesFound {
		report = "🚨 **HEALTH ISSUES DETECTED** 🚨\n\n" + healthReport.String() +
			"\n**Recommendation**: Review and address the issues above, prioritizing critical (🚨) alerts.\n\n" +
			"**Health Monitoring Context**: \n" +
			"- Status metrics (cluster_new_status, node_new_status, aggr_new_status, volume_new_status): 1 = healthy/online, 0 = unhealthy/offline\n" +
			"- State metrics (*_state): When querying state metrics, 0 = object is offline, 1 = object is online\n" +
			"- Health metrics (health_*): Any value ≥ 1 indicates an active alert or issue\n" +
			"- Capacity metrics: Monitor volume_size_used_percent and aggr_inode_used_percent for space issues\n"
	} else {
		report = "✅ **ALL SYSTEMS HEALTHY** ✅\n\n" + healthReport.String() +
			"\n**Status**: Your ONTAP infrastructure appears to be operating normally.\n"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: report},
		},
	}, nil, nil
}

func extractIdentifiers(metric map[string]any) string {
	var identifiers []string

	if cluster, ok := metric["cluster"].(string); ok {
		identifiers = append(identifiers, "cluster:"+cluster)
	}
	if node, ok := metric["node"].(string); ok {
		identifiers = append(identifiers, "node:"+node)
	}
	if volume, ok := metric["volume"].(string); ok {
		identifiers = append(identifiers, "volume:"+volume)
	}
	if aggr, ok := metric["aggr"].(string); ok {
		identifiers = append(identifiers, "aggr:"+aggr)
	}
	if severity, ok := metric["severity"].(string); ok {
		identifiers = append(identifiers, "severity:"+severity)
	}

	if len(identifiers) == 0 {
		return "unknown"
	}
	return strings.Join(identifiers, " ")
}

// getResourcePath returns path to a resource, trying common locations
func getResourcePath(resource string) string {
	paths := []string{
		filepath.Join("..", resource),        // From mcp/bin directory
		filepath.Join("..", "mcp", resource), // From harvest/bin
		filepath.Join("..", "..", resource),  // From mcp/cmd/server
		resource,                             // From mcp root or container
	}

	executable, err := os.Executable()
	if err != nil {
		logger.Error("failed to get executable path", slogx.Err(err))
		return resource
	}

	for _, path := range paths {
		fullPath := filepath.Join(executable, path)

		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	return resource
}

// validateTSDBConnection tests the connection to the time-series database (Prometheus/VictoriaMetrics)
// Retries 5 times with 10 second delay between attempts
func validateTSDBConnection(ctx context.Context, config auth.TSDBConfig) error {
	const maxRetries = 5
	const retryDelay = 10 * time.Second

	buildInfoURL := config.URL + "/api/v1/status/buildinfo"
	logger.Info("validating time-series database connection", slog.String("url", config.URL))

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
			}
			logger.Info("retrying connection", slog.Int("attempt", attempt))
		}

		resp, err := auth.MakeRequest(config, buildInfoURL)
		if err != nil {
			lastErr = fmt.Errorf("time-series database connection attempt failed: %w", err)
			continue
		}

		statusCode := resp.StatusCode
		_ = resp.Body.Close()

		if statusCode < 200 || statusCode >= 300 {
			lastErr = fmt.Errorf("time-series database returned HTTP %d", statusCode)
			continue
		}

		logger.Info("connection validated successfully", slog.Int("attempts", attempt))
		return nil
	}

	return fmt.Errorf("time-series database connection validation failed after %d attempts: %w", maxRetries, lastErr)
}

var rootCmd = &cobra.Command{
	Use:   "harvest-mcp",
	Short: "NetApp Harvest MCP Server - Model Context Protocol server for time-series metrics",
	Long: `NetApp Harvest MCP Server

A Model Context Protocol (MCP) server that provides access to Prometheus and VictoriaMetrics 
time-series data for ONTAP infrastructure monitoring and analysis.

Use "harvest-mcp start" to start the MCP server.

` + envVarHelpText,
}

// Command-line flags
var (
	httpMode bool
	httpPort int
	httpHost string
)

func init() {
	// Add flags to the start command
	startCmd.Flags().BoolVar(&httpMode, "http", false, "Enable HTTP transport mode (default: stdio)")
	startCmd.Flags().IntVar(&httpPort, "port", 8080, "Port for HTTP transport (default: 8080)")
	startCmd.Flags().StringVar(&httpHost, "host", "localhost", "Host for HTTP transport (default: localhost)")
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the MCP server",
	Long: `Start the NetApp Harvest MCP Server.

This command starts the Model Context Protocol (MCP) server that provides access to 
Prometheus and VictoriaMetrics time-series data for ONTAP infrastructure monitoring 
and analysis.

` + envVarHelpText,
	Run: runMcpServer,
}

func runMcpServer(_ *cobra.Command, _ []string) {
	logger.Info("starting harvest mcp server", slog.String("version", version.Info()))

	tsdbURL := os.Getenv("HARVEST_TSDB_URL")
	if tsdbURL == "" {
		logger.Error("HARVEST_TSDB_URL environment variable is required but not set")
		logger.Error("Please set HARVEST_TSDB_URL to point to your Prometheus/VictoriaMetrics server")
		os.Exit(1)
	}

	tsdbConfig = auth.GetTSDBConfig()

	logger.Info("server configuration",
		slog.String("tsdb_url", tsdbConfig.URL),
		slog.String("auth_type", string(tsdbConfig.Auth.Type)),
		slog.Bool("username_set", tsdbConfig.Auth.Username != ""),
		slog.Bool("password_set", tsdbConfig.Auth.Password != ""))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := validateTSDBConnection(ctx, tsdbConfig); err != nil {
		logger.Error("failed to connect to time-series database server",
			slogx.Err(err),
			slog.String("url", tsdbConfig.URL))
		logger.Error("please verify HARVEST_TSDB_URL and authentication settings")
		os.Exit(1) //nolint:gocritic
	}

	server := createMCPServer()

	if httpMode {
		runHTTPServer(server)
	} else {
		runStdioServer(server)
	}
}

func createMCPServer() *mcp.Server {
	instructions := "IMPORTANT:" + descriptions.Instructions

	server := mcp.NewServer(&mcp.Implementation{Name: AppName, Version: version.Info()}, &mcp.ServerOptions{
		Instructions: instructions,
	})

	metricDescriptions = loader.LoadMetricDescriptions(getResourcePath("metadata"), logger)

	addTool(server, "get_metric_description", descriptions.GetMetricDescriptionDesc, GetMetricDescription)
	addTool(server, "search_metrics", descriptions.SearchMetricsDesc, SearchMetrics)
	addTool(server, "metrics_query", descriptions.MetricsQueryDesc, MetricsQuery)
	addTool(server, "metrics_range_query", descriptions.MetricsRangeQueryDesc, MetricsRangeQuery)
	addTool(server, "list_metrics", descriptions.ListMetricsDesc, ListMetrics)
	addTool(server, "get_active_alerts", descriptions.GetActiveAlertsDesc, GetActiveAlerts)
	addTool(server, "infrastructure_health", descriptions.InfrastructureHealthDesc, InfrastructureHealth)
	addTool(server, "list_label_values", descriptions.ListLabelValuesDesc, ListLabelValues)
	addTool(server, "list_all_label_names", descriptions.ListAllLabelNamesDesc, ListAllLabelNames)
	addTool(server, "get_response_format_template", descriptions.GetResponseFormatTemplateDesc, GetResponseFormatTemplate)

	// Initialize rule manager
	var err error
	ruleManager, err = rules.NewRuleManager(logger)
	if err != nil {
		logger.Warn("failed to initialize rule manager", slogx.Err(err))
		logger.Info("rule management tools will be disabled - see environment configuration")
	} else {
		// Add rule management tools
		addTool(server, "list_alert_rules", "List all Prometheus alert rules from alert_rules.yml and ems_alert_rules.yml files", ListAlertRules)
		addTool(server, "create_alert_rule", "Create a new Prometheus alert rule in the appropriate file (alert_rules.yml or ems_alert_rules.yml)", CreateAlertRule)
		addTool(server, "update_alert_rule", "Update an existing Prometheus alert rule", UpdateAlertRule)
		addTool(server, "delete_alert_rule", "Delete a Prometheus alert rule", DeleteAlertRule)
		addTool(server, "validate_alert_syntax", "Validate the syntax of a PromQL expression for an alert rule", ValidateAlertSyntax)
		addTool(server, "reload_prometheus_rules", "Manually trigger a Prometheus configuration reload to apply rule changes", ReloadPrometheusRules)
		addTool(server, "get_rules_config_help", "Get help and documentation for configuring the rules management system", GetRulesConfigHelp)
	}

	promptDefinitions, err := loader.LoadPromptDefinitions(getResourcePath("prompts"), logger)
	if err != nil {
		logger.Error("failed to load prompt definitions", slogx.Err(err))
	}

	for _, promptDef := range promptDefinitions {
		content := promptDef.Content
		server.AddPrompt(
			&mcp.Prompt{
				Name:        promptDef.Name,
				Title:       promptDef.Title,
				Description: promptDef.Description,
			},
			func(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
				return &mcp.GetPromptResult{
					Messages: []*mcp.PromptMessage{
						{
							Role: "user",
							Content: &mcp.TextContent{
								Text: content,
							},
						},
					},
				}, nil
			},
		)
	}

	if len(promptDefinitions) == 0 {
		logger.Warn("no prompts loaded - MCP server will run without prompts")
	} else {
		logger.Info("registered prompts", slog.Int("count", len(promptDefinitions)))
	}

	return server
}

func runStdioServer(server *mcp.Server) {
	logger.Info("starting MCP server over stdio transport")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Error("mcp server failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func runHTTPServer(server *mcp.Server) {
	address := httpHost + ":" + strconv.Itoa(httpPort)
	logger.Info("starting MCP server over HTTP transport",
		slog.String("address", address),
		slog.String("host", httpHost),
		slog.Int("port", httpPort))

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, nil)

	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip MCP handler for health endpoint
		if r.URL.Path == "/health" {
			http.DefaultServeMux.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Protocol-Version, Mcp-Session-Id")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})

	logger.Info("MCP server endpoint available", slog.String("url", "http://"+address))
	logger.Info("Server ready to accept connections")

	httpServer := &http.Server{
		Addr:              address,
		Handler:           wrappedHandler,
		ReadHeaderTimeout: 60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		logger.Error("http server failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("mcp server shutdown gracefully")
}

func GetMetricDescription(_ context.Context, _ *mcp.CallToolRequest, params mcptypes.GetMetricDescriptionRequest) (*mcp.CallToolResult, any, error) {
	if len(metricDescriptions) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Metadata not available. Please ensure metadata files are generated and accessible."},
			},
			IsError: true,
		}, nil, nil
	}

	if params.MetricName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "MetricName parameter is required"},
			},
			IsError: true,
		}, nil, nil
	}

	description, found := metricDescriptions[params.MetricName]
	if !found {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No metadata found for metric: " + params.MetricName},
			},
		}, nil, nil
	}

	responseText := fmt.Sprintf("**Metric:** %s\n\n**Description:** %s", params.MetricName, description)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: responseText},
		},
	}, nil, nil
}

func SearchMetrics(_ context.Context, _ *mcp.CallToolRequest, params mcptypes.SearchMetricsRequest) (*mcp.CallToolResult, any, error) {
	if len(metricDescriptions) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Metadata not available. Please ensure metadata files are generated and accessible."},
			},
			IsError: true,
		}, nil, nil
	}

	if params.Pattern == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Pattern parameter is required"},
			},
			IsError: true,
		}, nil, nil
	}

	pattern := strings.ToLower(params.Pattern)
	var matches []struct {
		name        string
		description string
	}

	for metricName, description := range metricDescriptions {
		if strings.Contains(strings.ToLower(metricName), pattern) ||
			strings.Contains(strings.ToLower(description), pattern) {
			matches = append(matches, struct {
				name        string
				description string
			}{metricName, description})
		}
	}

	if len(matches) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No metrics found matching pattern: " + params.Pattern},
			},
		}, nil, nil
	}

	var responseBuilder strings.Builder
	fmt.Fprintf(&responseBuilder, "Found %d metrics matching pattern '%s':\n\n", len(matches), params.Pattern)

	for i, match := range matches {
		if i > 0 {
			responseBuilder.WriteString("\n---\n\n")
		}
		fmt.Fprintf(&responseBuilder, "**%s**\n%s", match.name, match.description)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: responseBuilder.String()},
		},
	}, nil, nil
}

func GetResponseFormatTemplate(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: descriptions.CoreResponseFormat},
		},
	}, nil, nil
}

// Rule management handlers

func ListAlertRules(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	if ruleManager == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Rule management is not available. Please configure HARVEST_RULES_PATH environment variable."},
			},
		}, nil, nil
	}

	response, err := ruleManager.ListRules()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error listing rules: %v", err)},
			},
		}, nil, nil
	}

	var builder strings.Builder
	builder.WriteString("**Alert Rules Summary**\n")
	fmt.Fprintf(&builder, "Total Rules: %d\n", response.TotalRules)
	fmt.Fprintf(&builder, "Last Modified: %s\n\n", response.LastModified)

	if len(response.AlertRules) > 0 {
		builder.WriteString("**Standard Alert Rules (alert_rules.yml):**\n")
		for _, rule := range response.AlertRules {
			fmt.Fprintf(&builder, "- **%s**\n", rule.Alert)
			fmt.Fprintf(&builder, "  Expression: `%s`\n", rule.Expr)
			if rule.For != "" {
				fmt.Fprintf(&builder, "  Duration: %s\n", rule.For)
			}
			if severity, ok := rule.Labels["severity"]; ok {
				fmt.Fprintf(&builder, "  Severity: %s\n", severity)
			}
			if summary, ok := rule.Annotations["summary"]; ok {
				fmt.Fprintf(&builder, "  Summary: %s\n", summary)
			}
			builder.WriteString("\n")
		}
	}

	if len(response.EMSRules) > 0 {
		builder.WriteString("**EMS Alert Rules (ems_alert_rules.yml):**\n")
		for _, rule := range response.EMSRules {
			fmt.Fprintf(&builder, "- **%s**\n", rule.Alert)
			fmt.Fprintf(&builder, "  Expression: `%s`\n", rule.Expr)
			if rule.For != "" {
				fmt.Fprintf(&builder, "  Duration: %s\n", rule.For)
			}
			if severity, ok := rule.Labels["severity"]; ok {
				fmt.Fprintf(&builder, "  Severity: %s\n", severity)
			}
			if summary, ok := rule.Annotations["summary"]; ok {
				fmt.Fprintf(&builder, "  Summary: %s\n", summary)
			}
			builder.WriteString("\n")
		}
	}

	if response.TotalRules == 0 {
		builder.WriteString("No alert rules found.")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: builder.String()},
		},
	}, response, nil
}

func CreateAlertRule(_ context.Context, _ *mcp.CallToolRequest, params rules.CreateRuleRequest) (*mcp.CallToolResult, any, error) {
	if ruleManager == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Rule management is not available. Please configure HARVEST_RULES_PATH environment variable."},
			},
		}, nil, nil
	}

	err := ruleManager.CreateRule(&params)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error creating rule: %v", err)},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Successfully created alert rule '%s'", params.RuleName)},
		},
	}, nil, nil
}

func UpdateAlertRule(_ context.Context, _ *mcp.CallToolRequest, params rules.UpdateRuleRequest) (*mcp.CallToolResult, any, error) {
	if ruleManager == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Rule management is not available. Please configure HARVEST_RULES_PATH environment variable."},
			},
		}, nil, nil
	}

	err := ruleManager.UpdateRule(&params)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error updating rule: %v", err)},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Successfully updated alert rule '%s'", params.RuleName)},
		},
	}, nil, nil
}

func DeleteAlertRule(_ context.Context, _ *mcp.CallToolRequest, params rules.DeleteRuleRequest) (*mcp.CallToolResult, any, error) {
	if ruleManager == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Rule management is not available. Please configure HARVEST_RULES_PATH environment variable."},
			},
		}, nil, nil
	}

	err := ruleManager.DeleteRule(&params)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error deleting rule: %v", err)},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Successfully deleted alert rule '%s'", params.RuleName)},
		},
	}, nil, nil
}

func ValidateAlertSyntax(_ context.Context, _ *mcp.CallToolRequest, params rules.ValidateRuleRequest) (*mcp.CallToolResult, any, error) {
	if ruleManager == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Rule management is not available. Please configure HARVEST_RULES_PATH environment variable."},
			},
		}, nil, nil
	}

	err := ruleManager.ValidateRule(&params)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Validation failed: %v", err)},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Rule '%s' with expression '%s' is valid", params.RuleName, params.Expression)},
		},
	}, nil, nil
}

func ReloadPrometheusRules(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	if ruleManager == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Rule management is not available. Please configure HARVEST_RULES_PATH environment variable."},
			},
		}, nil, nil
	}

	err := ruleManager.ReloadPrometheus()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error reloading Prometheus: %v\n\n%s", err, ruleManager.GetReloadInstructions())},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Successfully triggered Prometheus configuration reload"},
		},
	}, nil, nil
}

func GetRulesConfigHelp(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: envVarHelpText},
		},
	}, nil, nil
}

func main() {
	auth.SetLogger(logger)
	// Add version command
	rootCmd.AddCommand(version.Cmd())
	// Add start command
	rootCmd.AddCommand(startCmd)

	rootCmd.Version = version.String()
	rootCmd.SetVersionTemplate(version.String())

	cobra.CheckErr(rootCmd.Execute())
}
