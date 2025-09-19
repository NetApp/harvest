package rules

import (
	"errors"
	"fmt"
	"github.com/goccy/go-yaml"
	"regexp"
	"strings"
)

// Validator handles rule validation
type Validator struct{}

// NewValidator creates a new rule validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRule validates a complete rule structure
func (v *Validator) ValidateRule(rule *AlertRule) error {
	if err := v.ValidateRuleName(rule.Alert); err != nil {
		return err
	}

	if err := v.ValidatePromQL(rule.Expr); err != nil {
		return err
	}

	if rule.For != "" {
		if err := v.ValidateDuration(rule.For); err != nil {
			return err
		}
	}

	return nil
}

// ValidateRuleName validates the alert rule name
func (v *Validator) ValidateRuleName(name string) error {
	if name == "" {
		return errors.New("rule name cannot be empty")
	}

	// Rule names should follow Prometheus conventions
	// Must be valid metric name: [a-zA-Z_:][a-zA-Z0-9_:]*
	matched, err := regexp.MatchString(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`, name)
	if err != nil {
		return fmt.Errorf("error validating rule name: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid rule name '%s': must match pattern [a-zA-Z_:][a-zA-Z0-9_:]*", name)
	}

	return nil
}

// ValidatePromQL performs basic PromQL syntax validation
func (v *Validator) ValidatePromQL(expr string) error {
	if expr == "" {
		return errors.New("PromQL expression cannot be empty")
	}

	// Basic syntax checks
	if err := v.checkBasicPromQLSyntax(expr); err != nil {
		return err
	}

	return nil
}

// ValidateDuration validates the duration format
func (v *Validator) ValidateDuration(duration string) error {
	if duration == "" {
		return nil // Optional field
	}

	// Prometheus duration format: [0-9]+(ms|s|m|h|d|w|y)
	matched, err := regexp.MatchString(`^[0-9]+(\.[0-9]+)?(ms|s|m|h|d|w|y)$`, duration)
	if err != nil {
		return fmt.Errorf("error validating duration: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid duration format '%s': must be like '5m', '30s', '1h', etc", duration)
	}

	return nil
}

// ValidateYAMLSyntax validates YAML syntax for a rule file
func (v *Validator) ValidateYAMLSyntax(yamlContent []byte) error {
	var ruleFile RuleFile
	if err := yaml.Unmarshal(yamlContent, &ruleFile); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	// Validate structure
	if len(ruleFile.Groups) == 0 {
		return errors.New("rule file must contain at least one group")
	}

	for i, group := range ruleFile.Groups {
		if group.Name == "" {
			return fmt.Errorf("group %d must have a name", i)
		}

		if len(group.Rules) == 0 {
			return fmt.Errorf("group '%s' must contain at least one rule", group.Name)
		}

		for j, rule := range group.Rules {
			if err := v.ValidateRule(&rule); err != nil {
				return fmt.Errorf("invalid rule %d in group '%s': %w", j, group.Name, err)
			}
		}
	}

	return nil
}

// checkBasicPromQLSyntax performs basic PromQL syntax validation
func (v *Validator) checkBasicPromQLSyntax(expr string) error {
	expr = strings.TrimSpace(expr)

	// Check for balanced parentheses
	if err := v.checkBalanced(expr, '(', ')', "unmatched parentheses"); err != nil {
		return err
	}

	// Check for balanced brackets
	if err := v.checkBalanced(expr, '[', ']', "unmatched square braces"); err != nil {
		return err
	}

	// Check for balanced braces
	if err := v.checkBalanced(expr, '{', '}', "unmatched curly braces"); err != nil {
		return err
	}

	// Check for some common syntax errors
	if strings.Contains(expr, ",,") {
		return errors.New("invalid syntax: double comma in expression")
	}

	if strings.HasSuffix(expr, ",") || strings.HasSuffix(expr, "+") ||
		strings.HasSuffix(expr, "-") || strings.HasSuffix(expr, "*") ||
		strings.HasSuffix(expr, "/") {
		return errors.New("expression cannot end with an operator")
	}

	return nil
}

func (v *Validator) checkBalanced(expr string, opened int32, closed int32, errorText string) error {
	count := 0
	for _, char := range expr {
		switch char {
		case opened:
			count++
		case closed:
			count--
			if count < 0 {
				return errors.New(errorText)
			}
		}
	}

	if count != 0 {
		return errors.New(errorText)
	}

	return nil
}
