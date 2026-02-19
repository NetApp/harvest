package mcptypes

type MetricsResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     any    `json:"result"`
	} `json:"data"`
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"errorType,omitempty"`
}

type AlertsResponse struct {
	Status string `json:"status"`
	Data   struct {
		Alerts []any `json:"alerts"`
	} `json:"data"`
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"errorType,omitempty"`
}

type LabelsResponse struct {
	Status    string   `json:"status"`
	Data      []string `json:"data"`
	Error     string   `json:"error,omitempty"`
	ErrorType string   `json:"errorType,omitempty"`
}

// TSDBOverride allows per-request override of the Prometheus/VictoriaMetrics URL and credentials
// This is useful when a single MCP server needs to query multiple TSDB instances
type TSDBOverride struct {
	URL      string `json:"tsdb_url,omitempty" jsonschema:"Optional override for Prometheus/VictoriaMetrics URL. If not provided, uses the default HARVEST_TSDB_URL from server configuration. Example: http://prometheus-prod:9090"`
	Username string `json:"tsdb_username,omitempty" jsonschema:"Optional basic auth username. Only used if tsdb_url is provided. Overrides default authentication."`
	Password string `json:"tsdb_password,omitempty" jsonschema:"Optional basic auth password. Only used if tsdb_url is provided. Overrides default authentication."` //nolint:gosec
}

type QueryRequest struct {
	Query        string       `json:"query" jsonschema:"PromQL query string"`
	TSDBOverride TSDBOverride `json:"tsdb_override,omitzero" jsonschema:"Optional override for TSDB connection"`
}

type RangeQueryRequest struct {
	Query        string       `json:"query" jsonschema:"PromQL query string"`
	Start        string       `json:"start" jsonschema:"Start timestamp (RFC3339 or Unix timestamp)"`
	End          string       `json:"end" jsonschema:"End timestamp (RFC3339 or Unix timestamp)"`
	Step         string       `json:"step" jsonschema:"Query resolution step width (e.g., '15s', '1m', '1h')"`
	TSDBOverride TSDBOverride `json:"tsdb_override,omitzero" jsonschema:"Optional override for TSDB connection"`
}

type ListMetricsRequest struct {
	Match        string       `json:"match,omitempty" jsonschema:"Optional metric name pattern to filter results. Supports: 1) Simple string matching (e.g., 'volume'), 2) Regex patterns (e.g., '.*volume.*(latency|data|throughput).*'), 3) PromQL label matchers (e.g., '{__name__=~\".*volume.*\"}')"`
	Matches      string       `json:"matches,omitempty" jsonschema:"Comma-separated PromQL label matchers for server-side filtering. Example: '{__name__=~\"volume.*latency.*\"},{__name__=~\"volume.*data$\"}'. Each matcher is a PromQL selector passed to Prometheus for efficient server-side filtering."`
	TSDBOverride TSDBOverride `json:"tsdb_override,omitzero" jsonschema:"Optional override for TSDB connection"`
}

type InfrastructureHealthRequest struct {
	IncludeDetails bool         `json:"includeDetails,omitempty" jsonschema:"Include detailed metrics in the response"`
	TSDBOverride   TSDBOverride `json:"tsdb_override,omitzero" jsonschema:"Optional override for TSDB connection"`
}

type ListLabelValuesRequest struct {
	Label        string       `json:"label" jsonschema:"Label name to get values for (e.g., 'cluster', 'node', 'volume')"`
	Match        string       `json:"match,omitempty" jsonschema:"Optional pattern to filter label values. Supports simple string matching or regex patterns (e.g., '.*prod.*', '^cluster_[0-9]+$')"`
	TSDBOverride TSDBOverride `json:"tsdb_override,omitzero" jsonschema:"Optional override for TSDB connection"`
}

type GetMetricDescriptionRequest struct {
	MetricName   string       `json:"metricName" jsonschema:"The name of the metric to get description for"`
	TSDBOverride TSDBOverride `json:"tsdb_override,omitzero" jsonschema:"Optional override for TSDB connection"`
}

type SearchMetricsRequest struct {
	Pattern      string       `json:"pattern" jsonschema:"Search pattern to match against metric names and descriptions"`
	TSDBOverride TSDBOverride `json:"tsdb_override,omitzero" jsonschema:"Optional override for TSDB connection"`
}

type GetActiveAlertsRequest struct {
	TSDBOverride TSDBOverride `json:"tsdb_override,omitzero" jsonschema:"Optional override for TSDB connection"`
}

type ListAllLabelNamesRequest struct {
	TSDBOverride TSDBOverride `json:"tsdb_override,omitzero" jsonschema:"Optional override for TSDB connection"`
}
