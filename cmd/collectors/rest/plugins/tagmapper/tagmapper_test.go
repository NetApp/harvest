package tagmapper

import (
	"reflect"
	"testing"
)

func TestTagMapper_parseTagsToMap(t *testing.T) {
	tagMapper := &TagMapper{}

	tests := []struct {
		name     string
		tags     string
		expected map[string]string
	}{
		{
			name: "basic key-value pairs",
			tags: "severity:page,team:foobar",
			expected: map[string]string{
				"severity": "page",
				"team":     "foobar",
			},
		},
		{
			name: "single key-value pair",
			tags: "environment:production",
			expected: map[string]string{
				"environment": "production",
			},
		},
		{
			name:     "empty string",
			tags:     "",
			expected: map[string]string{},
		},
		{
			name: "value with colon",
			tags: "url:http://example.com:8080,type:service",
			expected: map[string]string{
				"url":  "http://example.com:8080",
				"type": "service",
			},
		},
		{
			name: "tags with spaces",
			tags: " severity : critical , team : backend ",
			expected: map[string]string{
				"severity": "critical",
				"team":     "backend",
			},
		},
		{
			name: "empty values and keys",
			tags: "severity:,team:foobar,:value,key:",
			expected: map[string]string{
				"team": "foobar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tagMapper.parseTagsToMap(tt.tags)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseTagsToMap() = %v, want %v", result, tt.expected)
			}
		})
	}
}
