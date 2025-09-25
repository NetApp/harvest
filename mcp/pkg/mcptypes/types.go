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

type QueryArgs struct {
	Query string `json:"query" jsonschema:"PromQL query string"`
}

type RangeQueryArgs struct {
	Query string `json:"query" jsonschema:"PromQL query string"`
	Start string `json:"start" jsonschema:"Start timestamp (RFC3339 or Unix timestamp)"`
	End   string `json:"end" jsonschema:"End timestamp (RFC3339 or Unix timestamp)"`
	Step  string `json:"step" jsonschema:"Query resolution step width (e.g., '15s', '1m', '1h')"`
}

type ListMetricsArgs struct {
	Match   string   `json:"match,omitempty" jsonschema:"Optional metric name pattern to filter results. Supports: 1) Simple string matching (e.g., 'volume'), 2) Regex patterns (e.g., '.*volume.*space.*'), 3) PromQL label matchers (e.g., '{__name__=~\".*volume.*\"}')"`
	Matches []string `json:"matches,omitempty" jsonschema:"Array of PromQL label matchers for server-side filtering (e.g., ['{__name__=~\".*volume.*space.*\"}']). More efficient than 'match' for complex patterns."`
}

type InfrastructureHealthArgs struct {
	IncludeDetails bool `json:"includeDetails,omitempty" jsonschema:"Include detailed metrics in the response"`
}

type ListLabelValuesArgs struct {
	Label string `json:"label" jsonschema:"Label name to get values for (e.g., 'cluster', 'node', 'volume')"`
	Match string `json:"match,omitempty" jsonschema:"Optional pattern to filter label values. Supports simple string matching or regex patterns (e.g., '.*prod.*', '^cluster_[0-9]+$')"`
}
