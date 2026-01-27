/*
 * Copyright NetApp Inc, 2025 All rights reserved
 */

package rest

import (
	"log/slog"
	"testing"

	"github.com/netapp/harvest/v2/pkg/collector"
)

func TestClient_normalizeBundleVersion(t *testing.T) {
	client := &Client{
		Logger:   slog.Default(),
		Metadata: &collector.Metadata{},
	}

	tests := []struct {
		name           string
		bundleDisplay  string
		expectedOutput string
	}{
		{
			name:           "NetApp branded with patch and revision",
			bundleDisplay:  "11.70.4R1",
			expectedOutput: "11.70.4",
		},
		{
			name:           "NetApp branded with revision only",
			bundleDisplay:  "11.90.R4",
			expectedOutput: "11.90.0",
		},
		{
			name:           "NetApp branded with GA suffix",
			bundleDisplay:  "12.00GA",
			expectedOutput: "12.00.0",
		},
		{
			name:           "NetApp branded with patch GA",
			bundleDisplay:  "11.30.5GA",
			expectedOutput: "11.30.5",
		},
		{
			name:           "Two component version",
			bundleDisplay:  "11.30",
			expectedOutput: "11.30.0",
		},
		{
			name:           "Patch with R suffix",
			bundleDisplay:  "11.50.3R1",
			expectedOutput: "11.50.3",
		},
		{
			name:           "Patch with P suffix",
			bundleDisplay:  "11.50.3R1P4",
			expectedOutput: "11.50.3",
		},
		{
			name:           "Zero patch version",
			bundleDisplay:  "11.70.0R1",
			expectedOutput: "11.70.0",
		},
		{
			name:           "Non-NetApp branded OEM version",
			bundleDisplay:  "08.74.02.00.005-LEN",
			expectedOutput: "08.74.02",
		},
		{
			name:           "Empty string",
			bundleDisplay:  "",
			expectedOutput: "",
		},
		{
			name:           "Single component",
			bundleDisplay:  "11",
			expectedOutput: "",
		},
		{
			name:           "Version with dot but no patch content",
			bundleDisplay:  "11.70.",
			expectedOutput: "11.70.0",
		},
		{
			name:           "Version with only alphabetic patch",
			bundleDisplay:  "11.70.R4",
			expectedOutput: "11.70.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.normalizeBundleVersion(tt.bundleDisplay)
			if result != tt.expectedOutput {
				t.Errorf("normalizeBundleVersion(%q) = %q, want %q",
					tt.bundleDisplay, result, tt.expectedOutput)
			}
		})
	}
}
