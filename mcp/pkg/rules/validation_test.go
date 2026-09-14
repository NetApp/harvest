package rules

import (
	"strings"
	"testing"
)

// TestValidateRuleName covers the Prometheus metric-name rule that alert names
// must satisfy. Rules related to GitHub issue #3142 "Usage of wrong labels in
// alert rules definition" enter through here.
func TestValidateRuleName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "simple name", in: "VolumeFull", wantErr: false},
		{name: "underscore prefix", in: "_internal", wantErr: false},
		{name: "colon prefix", in: ":recording", wantErr: false},
		{name: "digits after the first character", in: "Volume90Full", wantErr: false},
		{name: "underscores and colons", in: "harvest:volume_full", wantErr: false},
		{name: "empty", in: "", wantErr: true},
		{name: "leading digit", in: "9Lives", wantErr: true},
		{name: "contains a space", in: "Volume Full", wantErr: true},
		{name: "contains a hyphen", in: "volume-full", wantErr: true},
		{name: "contains a dot", in: "volume.full", wantErr: true},
	}

	v := NewValidator()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateRuleName(tc.in)
			if tc.wantErr && err == nil {
				t.Errorf("ValidateRuleName(%q): got nil, want an error", tc.in)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("ValidateRuleName(%q): got %v, want nil", tc.in, err)
			}
		})
	}
}

// TestValidateDuration covers the Prometheus duration format accepted in a
// rule's "for" field.
func TestValidateDuration(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "empty is allowed, the field is optional", in: "", wantErr: false},
		{name: "minutes", in: "5m", wantErr: false},
		{name: "seconds", in: "30s", wantErr: false},
		{name: "milliseconds", in: "500ms", wantErr: false},
		{name: "hours", in: "1h", wantErr: false},
		{name: "days", in: "2d", wantErr: false},
		{name: "weeks", in: "1w", wantErr: false},
		{name: "years", in: "1y", wantErr: false},
		{name: "fractional", in: "1.5h", wantErr: false},
		{name: "no unit", in: "5", wantErr: true},
		{name: "unknown unit", in: "5x", wantErr: true},
		{name: "unit only", in: "m", wantErr: true},
		{name: "negative", in: "-5m", wantErr: true},
		{name: "compound durations are not supported", in: "1h30m", wantErr: true},
		{name: "trailing space", in: "5m ", wantErr: true},
	}

	v := NewValidator()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateDuration(tc.in)
			if tc.wantErr && err == nil {
				t.Errorf("ValidateDuration(%q): got nil, want an error", tc.in)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("ValidateDuration(%q): got %v, want nil", tc.in, err)
			}
		})
	}
}

// TestValidatePromQL covers the balance and trailing-operator checks that catch
// the most common hand-written expression mistakes before they reach the rule
// file, where they would break the whole file's load rather than one rule.
func TestValidatePromQL(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "simple comparison", in: "up == 0", wantErr: false},
		{name: "selector with labels", in: `volume_size_used_percent{cluster="a"} > 90`, wantErr: false},
		{name: "range vector", in: "rate(node_total_data[5m]) > 1", wantErr: false},
		{name: "nested functions", in: "sum(rate(qos_read_data[5m])) by (cluster) > 1", wantErr: false},
		{name: "empty", in: "", wantErr: true},
		{name: "unmatched open paren", in: "sum(rate(up[5m]) > 1", wantErr: true},
		{name: "unmatched close paren", in: "sum rate(up[5m])) > 1", wantErr: true},
		{name: "unmatched square brace", in: "rate(up[5m > 1", wantErr: true},
		{name: "unmatched curly brace", in: `up{cluster="a" == 0`, wantErr: true},
		{name: "closing before opening", in: "up) == (0", wantErr: true},
		{name: "double comma", in: "sum(up) by (a,,b)", wantErr: true},
		{name: "trailing comma", in: "sum(up) by (a),", wantErr: true},
		{name: "trailing plus", in: "up +", wantErr: true},
		{name: "trailing slash", in: "up /", wantErr: true},
	}

	v := NewValidator()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidatePromQL(tc.in)
			if tc.wantErr && err == nil {
				t.Errorf("ValidatePromQL(%q): got nil, want an error", tc.in)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("ValidatePromQL(%q): got %v, want nil", tc.in, err)
			}
		})
	}
}

// TestValidateRule covers the composite check, including that a bad "for" value
// is rejected while an absent one is not.
func TestValidateRule(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		rule    AlertRule
		wantErr bool
	}{
		{
			name:    "valid rule with a duration",
			rule:    AlertRule{Alert: "VolumeFull", Expr: "volume_size_used_percent > 90", For: "5m"},
			wantErr: false,
		},
		{
			name:    "valid rule without a duration",
			rule:    AlertRule{Alert: "VolumeFull", Expr: "volume_size_used_percent > 90"},
			wantErr: false,
		},
		{
			name:    "bad name",
			rule:    AlertRule{Alert: "Volume Full", Expr: "up == 0"},
			wantErr: true,
		},
		{
			name:    "bad expression",
			rule:    AlertRule{Alert: "VolumeFull", Expr: "sum(up"},
			wantErr: true,
		},
		{
			name:    "bad duration",
			rule:    AlertRule{Alert: "VolumeFull", Expr: "up == 0", For: "soon"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateRule(&tc.rule)
			if tc.wantErr && err == nil {
				t.Error("got nil, want an error")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("got %v, want nil", err)
			}
		})
	}
}

// TestValidateYAMLSyntaxRejectsEmptyGroup pins the structural rule that matters
// most in practice: a group with no rules. vmalert refuses to load such a file,
// which is the same failure class as GitHub issue #4206 -- and DeleteRule can
// produce exactly this shape when it removes a group's last rule.
func TestValidateYAMLSyntaxRejectsEmptyGroup(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "valid file",
			yaml: `
groups:
  - name: harvest.rules
    rules:
      - alert: VolumeFull
        expr: volume_size_used_percent > 90
`,
		},
		{
			name: "group with no rules",
			yaml: `
groups:
  - name: harvest.rules
    rules: []
`,
			wantErr: "must contain at least one rule",
		},
		{
			name: "group with no name",
			yaml: `
groups:
  - name: ""
    rules:
      - alert: VolumeFull
        expr: up == 0
`,
			wantErr: "must have a name",
		},
		{
			name:    "no groups at all",
			yaml:    "groups: []\n",
			wantErr: "at least one group",
		},
		{
			name:    "not YAML",
			yaml:    "groups: [oops\n",
			wantErr: "invalid YAML syntax",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateYAMLSyntax([]byte(tc.yaml))
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("got %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("got nil, want an error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("got %q, want it to contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}
