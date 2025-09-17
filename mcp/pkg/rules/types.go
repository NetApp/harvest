package rules

import "time"

// AlertRule represents a Prometheus alerting rule
type AlertRule struct {
	Alert       string            `yaml:"alert"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// RuleGroup represents a group of alerting rules
type RuleGroup struct {
	Name     string      `yaml:"name"`
	Interval string      `yaml:"interval,omitempty"`
	Rules    []AlertRule `yaml:"rules"`
}

// RuleFile represents the structure of a Prometheus rules file
type RuleFile struct {
	Groups []RuleGroup `yaml:"groups"`
}

// RuleFileInfo contains metadata about a rule file
type RuleFileInfo struct {
	Path         string
	LastModified time.Time
	RuleCount    int
	GroupCount   int
}

// CreateRuleRequest represents the input for creating a new alert rule
type CreateRuleRequest struct {
	RuleName    string `json:"rule_name" jsonschema:"Name for the alert rule"`
	Expression  string `json:"expression" jsonschema:"PromQL expression. See docs://ontap-metrics docs://storagegrid-metrics docs://cisco-switch-metrics for available metrics"`
	Duration    string `json:"duration,omitempty" jsonschema:"Duration the condition must be true to fire alert (e.g. 5m, 30s)"`
	Severity    string `json:"severity,omitempty" jsonschema:"Alert severity: critical, warning, info"`
	Summary     string `json:"summary,omitempty" jsonschema:"Brief description of what the alert indicates"`
	Description string `json:"description,omitempty" jsonschema:"Detailed description of the alert and recommended actions"`
	Runbook     string `json:"runbook_url,omitempty" jsonschema:"URL to runbook or documentation for this alert"`
	GroupName   string `json:"group_name,omitempty" jsonschema:"Rule group name (defaults to harvest.rules or ems.rules)"`
}

// UpdateRuleRequest represents the input for updating an existing alert rule
type UpdateRuleRequest struct {
	RuleName       string `json:"rule_name" jsonschema:"Name of the existing alert rule to update"`
	NewExpression  string `json:"new_expression,omitempty" jsonschema:"New PromQL expression"`
	NewDuration    string `json:"new_duration,omitempty" jsonschema:"New duration for the alert condition"`
	NewSeverity    string `json:"new_severity,omitempty" jsonschema:"New alert severity"`
	NewSummary     string `json:"new_summary,omitempty" jsonschema:"New alert summary"`
	NewDescription string `json:"new_description,omitempty" jsonschema:"New alert description"`
	NewRunbook     string `json:"new_runbook_url,omitempty" jsonschema:"New runbook URL"`
}

// DeleteRuleRequest represents the input for deleting an alert rule
type DeleteRuleRequest struct {
	RuleName string `json:"rule_name" jsonschema:"Name of the alert rule to delete"`
}

// ValidateRuleRequest represents the input for validating a rule without saving
type ValidateRuleRequest struct {
	Expression string `json:"expression" jsonschema:"PromQL expression to validate"`
	RuleName   string `json:"rule_name,omitempty" jsonschema:"Optional rule name for validation"`
}

// RuleListResponse represents the response from listing rules
type RuleListResponse struct {
	AlertRules   []RuleInfo `json:"alert_rules"`
	EMSRules     []RuleInfo `json:"ems_rules"`
	TotalRules   int        `json:"total_rules"`
	LastModified string     `json:"last_modified"`
}

// RuleInfo provides information about a specific alert rule
type RuleInfo struct {
	Alert       string            `json:"alert" yaml:"alert"`
	Expr        string            `json:"expr" yaml:"expr"`
	For         string            `json:"for,omitempty" yaml:"for,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Group       string            `json:"group" yaml:"group"`
	File        string            `json:"file" yaml:"file"`
}
