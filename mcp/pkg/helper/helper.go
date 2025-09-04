package helper

import (
	"errors"
	"strings"
)

// ValidateQueryArgs validates PromQL query arguments
func ValidateQueryArgs(query string) error {
	if strings.TrimSpace(query) == "" {
		return errors.New("query parameter cannot be empty")
	}
	return nil
}

func ValidateRangeQueryArgs(query, start, end, step string) error {
	if err := ValidateQueryArgs(query); err != nil {
		return err
	}

	if strings.TrimSpace(start) == "" {
		return errors.New("start parameter cannot be empty")
	}

	if strings.TrimSpace(end) == "" {
		return errors.New("end parameter cannot be empty")
	}

	if strings.TrimSpace(step) == "" {
		return errors.New("step parameter cannot be empty")
	}

	return nil
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}
