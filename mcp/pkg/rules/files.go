package rules

import (
	"errors"
	"fmt"
	"github.com/goccy/go-yaml"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	AlertRulesFile   = "alert_rules.yml"
	EMSRulesFile     = "ems_alert_rules.yml"
	BackupExtension  = ".old"
	DefaultGroupName = "harvest.rules"
	EMSGroupName     = "ems.rules"
)

// Manager handles alert rule file operations
type Manager struct {
	rulesPath string
}

// NewManager creates a new rules manager
func NewManager(rulesPath string) *Manager {
	return &Manager{
		rulesPath: rulesPath,
	}
}

// determineTargetFile decides which file to use based on rule content
func (m *Manager) determineTargetFile(ruleName, expression string) string {
	// Check if this is an EMS-related rule
	lowerName := strings.ToLower(ruleName)
	lowerExpr := strings.ToLower(expression)

	if strings.Contains(lowerName, "ems") ||
		strings.Contains(lowerExpr, "ems_") ||
		strings.Contains(lowerExpr, "event_") {
		return EMSRulesFile
	}

	return AlertRulesFile
}

// getFilePath returns the full path for a rule file
func (m *Manager) getFilePath(filename string) string {
	return filepath.Join(m.rulesPath, filename)
}

// getBackupPath returns the backup file path
func (m *Manager) getBackupPath(filename string) string {
	return m.getFilePath(filename + BackupExtension)
}

// fileExists checks if a file exists
func (m *Manager) fileExists(filename string) bool {
	_, err := os.Stat(m.getFilePath(filename))
	return err == nil
}

// readRuleFile reads and parses a rule file
func (m *Manager) readRuleFile(filename string) (*RuleFile, error) {
	filePath := m.getFilePath(filename)

	// If file doesn't exist, return empty structure
	if !m.fileExists(filename) {
		return &RuleFile{Groups: []RuleGroup{}}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rule file %s: %w", filename, err)
	}

	var ruleFile RuleFile
	if err := yaml.Unmarshal(data, &ruleFile); err != nil {
		return nil, fmt.Errorf("failed to parse rule file %s: %w", filename, err)
	}

	return &ruleFile, nil
}

// writeRuleFile writes a rule file with backup
func (m *Manager) writeRuleFile(filename string, ruleFile *RuleFile) error {
	filePath := m.getFilePath(filename)
	backupPath := m.getBackupPath(filename)

	// Create backup if original file exists
	if m.fileExists(filename) {
		if err := m.createBackup(filename); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Marshal to YAML
	data, err := yaml.Marshal(ruleFile)
	if err != nil {
		return fmt.Errorf("failed to marshal rule file: %w", err)
	}

	// Write to temporary file first
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Rename temporary file to final location
	if err := os.Rename(tempPath, filePath); err != nil {
		// Clean up temp file
		err := os.Remove(tempPath)
		if err != nil {
			return err
		}
		// Restore backup if write failed
		if m.fileExists(filename + BackupExtension) {
			err := os.Rename(backupPath, filePath)
			if err != nil {
				return err
			}
		}
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// createBackup creates a backup of the specified file
func (m *Manager) createBackup(filename string) error {
	originalPath := m.getFilePath(filename)
	backupPath := m.getBackupPath(filename)

	// Read original file
	data, err := os.ReadFile(originalPath)
	if err != nil {
		return fmt.Errorf("failed to read original file for backup: %w", err)
	}

	// Write backup (this overwrites any existing .old file)
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// ensureRulesDirectory creates the rules directory if it doesn't exist
func (m *Manager) ensureRulesDirectory() error {
	if err := os.MkdirAll(m.rulesPath, 0750); err != nil {
		return fmt.Errorf("failed to create rules directory: %w", err)
	}
	return nil
}

// getFileInfo returns metadata about a rule file
func (m *Manager) getFileInfo(filename string) (*RuleFileInfo, error) {
	filePath := m.getFilePath(filename)

	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &RuleFileInfo{
				Path:         filePath,
				LastModified: time.Time{},
				RuleCount:    0,
				GroupCount:   0,
			}, nil
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	ruleFile, err := m.readRuleFile(filename)
	if err != nil {
		return nil, err
	}

	ruleCount := 0
	for _, group := range ruleFile.Groups {
		ruleCount += len(group.Rules)
	}

	return &RuleFileInfo{
		Path:         filePath,
		LastModified: stat.ModTime(),
		RuleCount:    ruleCount,
		GroupCount:   len(ruleFile.Groups),
	}, nil
}

// findRuleInFile searches for a rule by name in the specified file
func (m *Manager) findRuleInFile(filename, ruleName string) (*AlertRule, error) {
	ruleFile, err := m.readRuleFile(filename)
	if err != nil {
		return nil, err
	}

	for _, group := range ruleFile.Groups {
		for _, rule := range group.Rules {
			if rule.Alert == ruleName {
				return &rule, nil
			}
		}
	}

	return nil, fmt.Errorf("rule %s not found in %s", ruleName, filename)
}

// findRule searches for a rule across both rule files
func (m *Manager) findRule(ruleName string) (string, *AlertRule, error) {
	// Try alert_rules.yml first
	if rule, err := m.findRuleInFile(AlertRulesFile, ruleName); err == nil {
		return AlertRulesFile, rule, nil
	}

	// Try ems_alert_rules.yml
	if rule, err := m.findRuleInFile(EMSRulesFile, ruleName); err == nil {
		return EMSRulesFile, rule, nil
	}

	return "", nil, fmt.Errorf("rule %s not found in any rule file", ruleName)
}

// getAllRules returns all rules from both files
func (m *Manager) getAllRules() ([]RuleInfo, []RuleInfo, error) {
	var (
		alertRules, emsRules []RuleInfo
		alertError           error
		emsError             error
		alertRuleFile        *RuleFile
		emsRuleFile          *RuleFile
	)

	// Get rules from alert_rules.yml
	alertRuleFile, alertError = m.readRuleFile(AlertRulesFile)
	if alertError == nil {
		for _, group := range alertRuleFile.Groups {
			for _, rule := range group.Rules {
				alertRules = append(alertRules, RuleInfo{
					Alert:       rule.Alert,
					Expr:        rule.Expr,
					For:         rule.For,
					Labels:      rule.Labels,
					Annotations: rule.Annotations,
					Group:       group.Name,
					File:        AlertRulesFile,
				})
			}
		}
	}

	// Get rules from ems_alert_rules.yml
	emsRuleFile, emsError = m.readRuleFile(EMSRulesFile)
	if emsError == nil {
		for _, group := range emsRuleFile.Groups {
			for _, rule := range group.Rules {
				emsRules = append(emsRules, RuleInfo{
					Alert:       rule.Alert,
					Expr:        rule.Expr,
					For:         rule.For,
					Labels:      rule.Labels,
					Annotations: rule.Annotations,
					Group:       group.Name,
					File:        EMSRulesFile,
				})
			}
		}
	}

	return alertRules, emsRules, errors.Join(alertError, emsError)
}
