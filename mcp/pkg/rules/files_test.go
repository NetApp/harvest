package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

// shippedAlertRules is the rule file Harvest actually ships. It is read from
// the harvest module's tree rather than copied into testdata here: a copy would
// silently stop tracking the real file when a rule is renamed or added, and
// tracking it is the point. mcp is a separate Go module, but that only
// restricts imports -- reading the file at test time is fine.
//
// Note the shipped alert names contain spaces ("Volume Used Percentage Breach"),
// so they would be rejected by Validator.ValidateRuleName. That is intentional
// and not a defect: these were authored by hand as examples, while
// ValidateRuleName guards names arriving through the create_alert_rule tool.
// The round-trip tests below deliberately do not validate names.
const shippedAlertRules = "../../../container/prometheus/alert_rules.yml"

// newTestManager returns a Manager rooted at an empty temp dir.
func newTestManager(t *testing.T) (*Manager, string) {
	t.Helper()
	dir := t.TempDir()
	return NewManager(dir, discardLogger()), dir
}

// newSeededManager returns a Manager whose directory already holds a copy of
// the shipped alert_rules.yml, so tests read and rewrite a real rule file.
func newSeededManager(t *testing.T) (*Manager, string) {
	t.Helper()
	m, dir := newTestManager(t)
	data, err := os.ReadFile(shippedAlertRules)
	if err != nil {
		t.Fatalf("read %s: %v", shippedAlertRules, err)
	}
	// #nosec G703 -- dir is t.TempDir() and the filename is a package constant
	if err := os.WriteFile(filepath.Join(dir, AlertRulesFile), data, 0600); err != nil {
		t.Fatalf("seed %s: %v", AlertRulesFile, err)
	}
	return m, dir
}

// TestWriteRuleFileEmitsNoRulesMapKey is the regression test for GitHub issue
// #4206 "create_alert_rule writes non-standard rulesmap field that crashes
// vmalert", which shipped with the status/untested label.
//
// RuleGroup carries a rulesMap lookup index alongside the rules themselves.
// While that field was exported and had no yaml tag, marshalling round-tripped
// it back into alert_rules.yml as a "rulesmap:" key, which vmalert rejects on
// load. The fix was to unexport it -- a one-character invariant that nothing
// else protects, and that any refactor suggesting "export this" would undo.
func TestWriteRuleFileEmitsNoRulesMapKey(t *testing.T) {
	m, dir := newSeededManager(t)

	// Read the shipped file, then write it straight back out.
	ruleFile, err := m.readRuleFile(AlertRulesFile)
	if err != nil {
		t.Fatalf("readRuleFile: %v", err)
	}
	if len(ruleFile.Groups) == 0 {
		t.Fatal("expected at least one group in the shipped rule file")
	}
	// readRuleFile populates the lookup index, so this round-trip is exactly the
	// situation that produced the bad key.
	if ruleFile.Groups[0].rulesMap == nil {
		t.Fatal("expected readRuleFile to populate the rules lookup index")
	}

	if err := m.writeRuleFile(AlertRulesFile, ruleFile); err != nil {
		t.Fatalf("writeRuleFile: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(dir, AlertRulesFile))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}

	lower := strings.ToLower(string(out))
	if strings.Contains(lower, "rulesmap") {
		t.Errorf("written rule file contains a rulesmap key, which vmalert rejects:\n%s", out)
	}
}

// TestWrittenGroupKeysAreOnlyTheSchemaKeys is the stronger form of the test
// above: rather than denying one known-bad key, it asserts the whole set of
// keys a group may emit. A future field added without a yaml:"-" tag fails here
// even if it is not called rulesMap.
func TestWrittenGroupKeysAreOnlyTheSchemaKeys(t *testing.T) {
	m, dir := newSeededManager(t)

	ruleFile, err := m.readRuleFile(AlertRulesFile)
	if err != nil {
		t.Fatalf("readRuleFile: %v", err)
	}
	if err := m.writeRuleFile(AlertRulesFile, ruleFile); err != nil {
		t.Fatalf("writeRuleFile: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(dir, AlertRulesFile))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}

	var generic struct {
		Groups []map[string]any `yaml:"groups"`
	}
	if err := yaml.Unmarshal(out, &generic); err != nil {
		t.Fatalf("written file is not valid YAML: %v", err)
	}
	if len(generic.Groups) == 0 {
		t.Fatal("written file has no groups")
	}

	allowed := map[string]bool{"name": true, "interval": true, "rules": true}
	for i, group := range generic.Groups {
		for key := range group {
			if !allowed[key] {
				t.Errorf("group %d emitted unexpected key %q; only name/interval/rules belong in a Prometheus rule group", i, key)
			}
		}
	}
}

// TestRoundTripPreservesRules pins that the round-trip above is lossless, so
// the #4206 test cannot be satisfied by writing nothing useful.
func TestRoundTripPreservesRules(t *testing.T) {
	m, _ := newSeededManager(t)

	before, err := m.readRuleFile(AlertRulesFile)
	if err != nil {
		t.Fatalf("readRuleFile: %v", err)
	}
	if err := m.writeRuleFile(AlertRulesFile, before); err != nil {
		t.Fatalf("writeRuleFile: %v", err)
	}
	after, err := m.readRuleFile(AlertRulesFile)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	if len(after.Groups) != len(before.Groups) {
		t.Fatalf("group count: got %d, want %d", len(after.Groups), len(before.Groups))
	}
	for i := range before.Groups {
		bg, ag := before.Groups[i], after.Groups[i]
		if ag.Name != bg.Name {
			t.Errorf("group %d name: got %q, want %q", i, ag.Name, bg.Name)
		}
		if len(ag.Rules) != len(bg.Rules) {
			t.Fatalf("group %q rule count: got %d, want %d", bg.Name, len(ag.Rules), len(bg.Rules))
		}
		for j := range bg.Rules {
			if ag.Rules[j].Alert != bg.Rules[j].Alert {
				t.Errorf("group %q rule %d alert: got %q, want %q", bg.Name, j, ag.Rules[j].Alert, bg.Rules[j].Alert)
			}
			if ag.Rules[j].Expr != bg.Rules[j].Expr {
				t.Errorf("rule %q expr: got %q, want %q", bg.Rules[j].Alert, ag.Rules[j].Expr, bg.Rules[j].Expr)
			}
			if ag.Rules[j].For != bg.Rules[j].For {
				t.Errorf("rule %q for: got %q, want %q", bg.Rules[j].Alert, ag.Rules[j].For, bg.Rules[j].For)
			}
		}
	}
}

// TestWriteRuleFileCreatesBackup pins the backup-then-rename behaviour, since
// these writes mutate a file Prometheus is actively loading.
func TestWriteRuleFileCreatesBackup(t *testing.T) {
	m, dir := newSeededManager(t)

	ruleFile, err := m.readRuleFile(AlertRulesFile)
	if err != nil {
		t.Fatalf("readRuleFile: %v", err)
	}
	if err := m.writeRuleFile(AlertRulesFile, ruleFile); err != nil {
		t.Fatalf("writeRuleFile: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, AlertRulesFile+BackupExtension)); err != nil {
		t.Errorf("expected a %s backup beside the rule file: %v", BackupExtension, err)
	}
	// The temp file must not be left behind.
	if _, err := os.Stat(filepath.Join(dir, AlertRulesFile+".tmp")); err == nil {
		t.Error("temporary file was not cleaned up")
	}
}

// TestReadRuleFileMissingReturnsEmpty pins that a not-yet-created rule file is
// an empty structure rather than an error, which is what CreateRule relies on
// when it writes the first rule.
func TestReadRuleFileMissingReturnsEmpty(t *testing.T) {
	m, _ := newTestManager(t)

	ruleFile, err := m.readRuleFile(AlertRulesFile)
	if err != nil {
		t.Fatalf("readRuleFile on a missing file: %v", err)
	}
	if ruleFile == nil {
		t.Fatal("got nil RuleFile")
	}
	if len(ruleFile.Groups) != 0 {
		t.Errorf("got %d groups, want 0", len(ruleFile.Groups))
	}
}

// TestDetermineTargetFile pins the routing between the two rule files. Getting
// this wrong writes EMS rules into the ONTAP file, where their group name does
// not match what Prometheus expects to reload.
func TestDetermineTargetFile(t *testing.T) {
	m, _ := newTestManager(t)

	tests := []struct {
		name     string
		ruleName string
		expr     string
		want     string
	}{
		{name: "ems in the rule name", ruleName: "EmsNodeDown", expr: "up == 0", want: EMSRulesFile},
		{name: "ems in the name, lowercase", ruleName: "my_ems_alert", expr: "up == 0", want: EMSRulesFile},
		{name: "ems_ prefix in the expression", ruleName: "NodeDown", expr: "ems_events{} > 0", want: EMSRulesFile},
		{name: "plain ontap rule", ruleName: "VolumeFull", expr: "volume_size_used_percent > 90", want: AlertRulesFile},
		{name: "unrelated name and expression", ruleName: "VolumeFull", expr: "aggr_space_used > 1", want: AlertRulesFile},

		// The routing is a plain substring match, so any rule name that merely
		// contains "ems" is routed to the EMS file -- "ItemsHigh" and
		// "SystemsDown" both match. That is surprising, but it is the shipped
		// behaviour and users' existing files depend on where their rules
		// landed, so it is pinned here rather than changed.
		{name: "substring match routes ItemsHigh to EMS", ruleName: "ItemsHigh", expr: "items > 1", want: EMSRulesFile},
		{name: "substring match routes SystemsDown to EMS", ruleName: "SystemsDown", expr: "up == 0", want: EMSRulesFile},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := m.determineTargetFile(tc.ruleName, tc.expr); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
