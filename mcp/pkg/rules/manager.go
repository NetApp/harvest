package rules

import (
	"errors"
	"fmt"
	"log/slog"
	"mcp-server/pkg/auth"
	"slices"
	"sync"
	"time"
)

// RuleManager combines all rule management functionality
type RuleManager struct {
	fileManager      *Manager
	validator        *Validator
	prometheusClient *PrometheusClient
	logger           *slog.Logger
	mu               *sync.Mutex
}

// NewRuleManager creates a new rule manager
func NewRuleManager(logger *slog.Logger) (*RuleManager, error) {
	config := auth.GetTSDBConfig()

	if config.RulesPath == "" {
		return nil, errors.New("HARVEST_RULES_PATH environment variable not set")
	}

	fileManager := NewManager(config.RulesPath, logger)
	validator := NewValidator()
	prometheusClient := NewPrometheusClient(config, logger)

	// Ensure rules directory exists
	if err := fileManager.ensureRulesDirectory(); err != nil {
		return nil, fmt.Errorf("failed to ensure rules directory: %w", err)
	}

	return &RuleManager{
		fileManager:      fileManager,
		validator:        validator,
		prometheusClient: prometheusClient,
		logger:           logger,
		mu:               &sync.Mutex{},
	}, nil
}

// ListRules returns all rules from both files
func (rm *RuleManager) ListRules() (*RuleListResponse, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	alertRules, emsRules, err := rm.fileManager.getAllRules()
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	totalRules := len(alertRules) + len(emsRules)

	// Get the most recent modification time
	var lastModified time.Time
	if alertInfo, err := rm.fileManager.getFileInfo(AlertRulesFile); err == nil {
		if alertInfo.LastModified.After(lastModified) {
			lastModified = alertInfo.LastModified
		}
	}
	if emsInfo, err := rm.fileManager.getFileInfo(EMSRulesFile); err == nil {
		if emsInfo.LastModified.After(lastModified) {
			lastModified = emsInfo.LastModified
		}
	}

	return &RuleListResponse{
		AlertRules:   alertRules,
		EMSRules:     emsRules,
		TotalRules:   totalRules,
		LastModified: lastModified.Format(time.RFC3339),
	}, nil
}

// CreateRule creates a new alert rule
func (rm *RuleManager) CreateRule(req *CreateRuleRequest) error {
	// Validate the rule
	rule := &AlertRule{
		Alert:       req.RuleName,
		Expr:        req.Expression,
		For:         req.Duration,
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}

	// Add severity label if provided
	if req.Severity != "" {
		rule.Labels["severity"] = req.Severity
	}

	// Add annotations if provided
	if req.Summary != "" {
		rule.Annotations["summary"] = req.Summary
	}
	if req.Description != "" {
		rule.Annotations["description"] = req.Description
	}
	if req.Runbook != "" {
		rule.Annotations["runbook_url"] = req.Runbook
	}

	// Validate the rule
	if err := rm.validator.ValidateRule(rule); err != nil {
		return fmt.Errorf("rule validation failed: %w", err)
	}

	// Determine target file
	targetFile := rm.fileManager.determineTargetFile(req.RuleName, req.Expression)

	rm.mu.Lock()
	defer rm.mu.Unlock()
	// Check if rule already exists
	if _, _, err := rm.fileManager.findRule(req.RuleName); err == nil {
		return fmt.Errorf("rule '%s' already exists", req.RuleName)
	}

	// Read the target file
	ruleFile, err := rm.fileManager.readRuleFile(targetFile)
	if err != nil {
		return fmt.Errorf("failed to read rule file: %w", err)
	}

	// Determine group name
	groupName := req.GroupName
	if groupName == "" {
		if targetFile == EMSRulesFile {
			groupName = EMSGroupName
		} else {
			groupName = DefaultGroupName
		}
	}

	// Find or create the group
	var targetGroup *RuleGroup
	for i := range ruleFile.Groups {
		if ruleFile.Groups[i].Name == groupName {
			targetGroup = &ruleFile.Groups[i]
			break
		}
	}

	if targetGroup == nil {
		// Create new group
		newGroup := RuleGroup{
			Name:  groupName,
			Rules: []AlertRule{},
		}
		ruleFile.Groups = append(ruleFile.Groups, newGroup)
		targetGroup = &ruleFile.Groups[len(ruleFile.Groups)-1]
	}

	// Add the rule to the group
	targetGroup.Rules = append(targetGroup.Rules, *rule)

	// Write the file
	if err := rm.fileManager.writeRuleFile(targetFile, ruleFile); err != nil {
		return fmt.Errorf("failed to write rule file: %w", err)
	}

	rm.logger.Info("created new alert rule",
		slog.String("rule", req.RuleName),
		slog.String("file", targetFile),
		slog.String("group", groupName))

	// Reload Prometheus if possible
	if err := rm.reloadPrometheus(); err != nil {
		rm.logger.Warn("failed to reload Prometheus", slog.Any("error", err))
		return fmt.Errorf("rule created but failed to reload Prometheus: %w", err)
	}

	return nil
}

// UpdateRule updates an existing alert rule
func (rm *RuleManager) UpdateRule(req *UpdateRuleRequest) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Find the existing rule
	filename, rule, err := rm.fileManager.findRule(req.RuleName)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}

	// Update fields if provided
	if req.NewExpression != "" {
		rule.Expr = req.NewExpression
	}
	if req.NewDuration != "" {
		rule.For = req.NewDuration
	}
	if req.NewSeverity != "" {
		if rule.Labels == nil {
			rule.Labels = make(map[string]string)
		}
		rule.Labels["severity"] = req.NewSeverity
	}
	if req.NewSummary != "" {
		if rule.Annotations == nil {
			rule.Annotations = make(map[string]string)
		}
		rule.Annotations["summary"] = req.NewSummary
	}
	if req.NewDescription != "" {
		if rule.Annotations == nil {
			rule.Annotations = make(map[string]string)
		}
		rule.Annotations["description"] = req.NewDescription
	}
	if req.NewRunbook != "" {
		if rule.Annotations == nil {
			rule.Annotations = make(map[string]string)
		}
		rule.Annotations["runbook_url"] = req.NewRunbook
	}

	// Validate the updated rule
	if err := rm.validator.ValidateRule(rule); err != nil {
		return fmt.Errorf("updated rule validation failed: %w", err)
	}

	// Read the file and update the rule
	ruleFile, err := rm.fileManager.readRuleFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read rule file: %w", err)
	}

	// Find and update the rule in the file
	found := false
	for groupIdx := range ruleFile.Groups {
		for i, r := range ruleFile.Groups[groupIdx].Rules {
			if r.Alert == req.RuleName {
				ruleFile.Groups[groupIdx].Rules[i] = *rule
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return errors.New("rule not found in file structure")
	}

	// Write the file
	if err := rm.fileManager.writeRuleFile(filename, ruleFile); err != nil {
		return fmt.Errorf("failed to write rule file: %w", err)
	}

	rm.logger.Info("updated alert rule",
		slog.String("rule", req.RuleName),
		slog.String("file", filename))

	// Reload Prometheus if possible
	if err := rm.reloadPrometheus(); err != nil {
		rm.logger.Warn("failed to reload Prometheus", slog.Any("error", err))
		return fmt.Errorf("rule updated but failed to reload Prometheus: %w", err)
	}

	return nil
}

// DeleteRule deletes an alert rule
func (rm *RuleManager) DeleteRule(req *DeleteRuleRequest) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Find the rule
	filename, _, err := rm.fileManager.findRule(req.RuleName)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}

	// Read the file
	ruleFile, err := rm.fileManager.readRuleFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read rule file: %w", err)
	}

	// Find and remove the rule
	found := false
	for groupIdx := range ruleFile.Groups {
		for ruleIdx, rule := range ruleFile.Groups[groupIdx].Rules {
			if rule.Alert == req.RuleName {
				// Remove the rule from the slice
				ruleFile.Groups[groupIdx].Rules = slices.Delete(ruleFile.Groups[groupIdx].Rules, ruleIdx, ruleIdx+1)
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return errors.New("rule not found in file structure")
	}

	// Write the file
	if err := rm.fileManager.writeRuleFile(filename, ruleFile); err != nil {
		return fmt.Errorf("failed to write rule file: %w", err)
	}

	rm.logger.Info("deleted alert rule",
		slog.String("rule", req.RuleName),
		slog.String("file", filename))

	// Reload Prometheus if possible
	if err := rm.reloadPrometheus(); err != nil {
		rm.logger.Warn("failed to reload Prometheus", slog.Any("error", err))
		return fmt.Errorf("rule deleted but failed to reload Prometheus: %w", err)
	}

	return nil
}

// ValidateRule validates a rule without saving it
func (rm *RuleManager) ValidateRule(req *ValidateRuleRequest) error {
	rule := &AlertRule{
		Alert: req.RuleName,
		Expr:  req.Expression,
	}

	return rm.validator.ValidateRule(rule)
}

// ReloadPrometheus manually triggers a Prometheus reload
func (rm *RuleManager) ReloadPrometheus() error {
	return rm.reloadPrometheus()
}

// GetReloadInstructions returns reload instructions
func (rm *RuleManager) GetReloadInstructions() string {
	return rm.prometheusClient.GetReloadInstructions()
}

// reloadPrometheus internal method to reload Prometheus
func (rm *RuleManager) reloadPrometheus() error {
	if !rm.prometheusClient.CanReload() {
		rm.logger.Debug("Prometheus reload API disabled, skipping reload")
		return nil
	}

	return rm.prometheusClient.ReloadConfig()
}
