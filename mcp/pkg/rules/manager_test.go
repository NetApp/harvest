package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
)

// newTestRuleManager builds a RuleManager rooted at a temp dir. NewRuleManager
// reads HARVEST_RULES_PATH through auth.GetTSDBConfig, and the Prometheus
// reload API defaults to disabled, so reloadPrometheus is a no-op here.
func newTestRuleManager(t *testing.T) (*RuleManager, string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HARVEST_RULES_PATH", dir)

	rm, err := NewRuleManager(discardLogger())
	if err != nil {
		t.Fatalf("NewRuleManager: %v", err)
	}
	return rm, dir
}

// seedAlertRules writes the alert rule file directly into the manager's
// directory, so a test can start from a known file rather than building one
// through the tool API.
func seedAlertRules(t *testing.T, dir string, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, AlertRulesFile), []byte(body), 0600); err != nil {
		t.Fatalf("seed %s: %v", AlertRulesFile, err)
	}
}

// readGroups re-reads a rule file as generic YAML, so assertions see exactly
// what Prometheus would load rather than the Go structs.
func readGroups(t *testing.T, dir string, filename string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	var generic struct {
		Groups []map[string]any `yaml:"groups"`
	}
	if err := yaml.Unmarshal(data, &generic); err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	return generic.Groups
}

// TestCreateRuleWritesValidYAML is the end-to-end form of the GitHub issue
// #4206 regression: the tool that customers actually call must produce a file
// vmalert can load.
func TestCreateRuleWritesValidYAML(t *testing.T) {
	rm, dir := newTestRuleManager(t)

	err := rm.CreateRule(&CreateRuleRequest{
		RuleName:    "VolumeFull",
		Expression:  "volume_size_used_percent > 90",
		Duration:    "5m",
		Severity:    "critical",
		Summary:     "Volume is nearly full",
		Description: "Free space before it fills",
	})
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, AlertRulesFile))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}

	// The file the tool just wrote must pass the project's own validator.
	if err := NewValidator().ValidateYAMLSyntax(data); err != nil {
		t.Errorf("CreateRule produced a file that fails validation: %v\n%s", err, data)
	}

	groups := readGroups(t, dir, AlertRulesFile)
	if len(groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(groups))
	}
	allowed := map[string]bool{"name": true, "interval": true, "rules": true}
	for key := range groups[0] {
		if !allowed[key] {
			t.Errorf("group emitted unexpected key %q", key)
		}
	}
}

// TestCreateRuleMapsSeverityAndAnnotations pins where each request field lands.
// Severity has to become a label -- routing and silencing key off labels, not
// annotations -- which is the surface of GitHub issue #3142 "Usage of wrong
// labels in alert rules definition".
func TestCreateRuleMapsSeverityAndAnnotations(t *testing.T) {
	rm, _ := newTestRuleManager(t)

	err := rm.CreateRule(&CreateRuleRequest{
		RuleName:    "VolumeFull",
		Expression:  "volume_size_used_percent > 90",
		Severity:    "warning",
		Summary:     "nearly full",
		Description: "add space",
		Runbook:     "https://example.com/runbook",
	})
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	_, rule, err := rm.fileManager.findRule("VolumeFull")
	if err != nil {
		t.Fatalf("findRule: %v", err)
	}

	if got := rule.Labels["severity"]; got != "warning" {
		t.Errorf("severity label: got %q, want %q", got, "warning")
	}
	if got := rule.Annotations["summary"]; got != "nearly full" {
		t.Errorf("summary annotation: got %q, want %q", got, "nearly full")
	}
	if got := rule.Annotations["description"]; got != "add space" {
		t.Errorf("description annotation: got %q, want %q", got, "add space")
	}
	if got := rule.Annotations["runbook_url"]; got != "https://example.com/runbook" {
		t.Errorf("runbook_url annotation: got %q, want %q", got, "https://example.com/runbook")
	}
	// Severity must not also be duplicated into annotations.
	if _, ok := rule.Annotations["severity"]; ok {
		t.Error("severity must be a label, not an annotation")
	}
}

// TestCreateRuleRejectsInvalidRule pins that validation happens before the file
// is touched, so one bad tool call cannot corrupt a working rule file.
func TestCreateRuleRejectsInvalidRule(t *testing.T) {
	rm, dir := newTestRuleManager(t)

	err := rm.CreateRule(&CreateRuleRequest{
		RuleName:   "Volume Full", // spaces are not a valid alert name
		Expression: "up == 0",
	})
	if err == nil {
		t.Fatal("got nil, want a validation error")
	}

	if _, statErr := os.Stat(filepath.Join(dir, AlertRulesFile)); statErr == nil {
		t.Error("an invalid rule must not create a rule file")
	}
}

// TestDeleteRuleLeavesNoEmptyGroup documents a real gap adjacent to GitHub issue
// #4206: removing a group's only rule leaves "rules: []" behind, and vmalert
// rejects a group with no rules just as it rejects an unknown key.
//
// This is the project's own validator disagreeing with its own writer, so the
// test asserts against ValidateYAMLSyntax rather than inventing a new rule.
func TestDeleteRuleLeavesNoEmptyGroup(t *testing.T) {
	rm, dir := newTestRuleManager(t)

	seedAlertRules(t, dir, `
groups:
  - name: harvest.rules
    rules:
      - alert: OnlyRule
        expr: up == 0
  - name: other.rules
    rules:
      - alert: KeepMe
        expr: up == 1
`)

	if err := rm.DeleteRule(&DeleteRuleRequest{RuleName: "OnlyRule"}); err != nil {
		t.Fatalf("DeleteRule: %v", err)
	}

	// The surviving rule must still be there.
	if _, _, err := rm.fileManager.findRule("KeepMe"); err != nil {
		t.Errorf("KeepMe should have survived the delete: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, AlertRulesFile))
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	// The file DeleteRule just wrote must still load.
	if err := NewValidator().ValidateYAMLSyntax(data); err != nil {
		t.Errorf("DeleteRule left a file that fails validation: %v\n%s", err, data)
	}
}

// TestDeleteRuleUnknownRule pins that deleting something absent is an error
// rather than a silent no-op that rewrites the file.
func TestDeleteRuleUnknownRule(t *testing.T) {
	rm, dir := newTestRuleManager(t)

	seedAlertRules(t, dir, `
groups:
  - name: harvest.rules
    rules:
      - alert: KeepMe
        expr: up == 1
`)

	if err := rm.DeleteRule(&DeleteRuleRequest{RuleName: "NoSuchRule"}); err == nil {
		t.Error("got nil, want a not-found error")
	}

	if _, _, err := rm.fileManager.findRule("KeepMe"); err != nil {
		t.Errorf("KeepMe must be untouched: %v", err)
	}
}

// TestDeleteLastRuleEmptiesTheFile documents the boundary of the empty-group
// cleanup: when the deleted rule was the only one in the file, the result is
// "groups: []". Prometheus loads that happily as "no rules", though the
// project's own ValidateYAMLSyntax rejects it, since that validator exists to
// check user-submitted rule content rather than an intentionally emptied file.
func TestDeleteLastRuleEmptiesTheFile(t *testing.T) {
	rm, dir := newTestRuleManager(t)

	seedAlertRules(t, dir, `
groups:
  - name: harvest.rules
    rules:
      - alert: OnlyRule
        expr: up == 0
`)

	if err := rm.DeleteRule(&DeleteRuleRequest{RuleName: "OnlyRule"}); err != nil {
		t.Fatalf("DeleteRule: %v", err)
	}

	groups := readGroups(t, dir, AlertRulesFile)
	if len(groups) != 0 {
		t.Errorf("got %d groups, want 0; a group with no rules must not be left behind: %v", len(groups), groups)
	}
}

// TestDeleteRuleKeepsOtherGroupsIntact pins that the empty-group cleanup only
// removes the group it emptied, and does not disturb its neighbours.
func TestDeleteRuleKeepsOtherGroupsIntact(t *testing.T) {
	rm, dir := newTestRuleManager(t)

	seedAlertRules(t, dir, `
groups:
  - name: first.rules
    rules:
      - alert: DeleteMe
        expr: up == 0
  - name: second.rules
    rules:
      - alert: KeepOne
        expr: up == 1
      - alert: KeepTwo
        expr: up == 2
`)

	if err := rm.DeleteRule(&DeleteRuleRequest{RuleName: "DeleteMe"}); err != nil {
		t.Fatalf("DeleteRule: %v", err)
	}

	groups := readGroups(t, dir, AlertRulesFile)
	if len(groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(groups))
	}
	if got := groups[0]["name"]; got != "second.rules" {
		t.Errorf("surviving group: got %v, want second.rules", got)
	}
	rules, ok := groups[0]["rules"].([]any)
	if !ok {
		t.Fatalf("rules is not a list: %T", groups[0]["rules"])
	}
	if len(rules) != 2 {
		t.Errorf("got %d rules in second.rules, want 2", len(rules))
	}
}
