package statperf

import (
	"log/slog"
	"os"
	"testing"
)

type testCase struct {
	name     string
	fileName string
}

func TestParseCounters(t *testing.T) {

	testCases := []testCase{
		{
			name:     "admin_login",
			fileName: "testdata/counters.txt",
		}, {
			name:     "user_login",
			fileName: "testdata/counters_1.txt",
		},
		{
			name:     "user_login_permission_changed",
			fileName: "testdata/counters_2.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(tc.fileName)
			if err != nil {
				t.Fatalf("Failed to read from file: %v", err)
			}

			sp := &StatPerf{}
			counters, err := sp.parseCounters(string(content))
			if err != nil {
				t.Fatalf("Unexpected error during parseCounters: %v", err)
			}
			if len(counters) != 32 {
				t.Errorf("Expected 32 counter, got %d", len(counters))
			}

			cp, exists := counters["nix_skipped_reason_offline"]
			if !exists {
				t.Fatalf("Expected to find counter 'nix_skipped_reason_offline'")
			}

			if cp.BaseCounter != "-" {
				t.Errorf("Expected BaseCounter '-', got '%s'", cp.BaseCounter)
			}

			if cp.Properties != "delta,no-zero-values" {
				t.Errorf("Expected Properties 'delta,no-zero-values', got '%s'", cp.Properties)
			}

			if cp.Type != "-" {
				t.Errorf("Expected Type '-', got '%s'", cp.Type)
			}
			if cp.Deprecated != "false" {
				t.Errorf("Expected Deprecated 'false', got '%s'", cp.Deprecated)
			}
			if cp.ReplacedBy != "-" {
				t.Errorf("Expected ReplacedBy '-', got '%s'", cp.ReplacedBy)
			}
		})
	}
}

func TestParseInstances(t *testing.T) {

	testCases := []testCase{
		{
			name:     "admin_login",
			fileName: "testdata/instances.txt",
		}, {
			name:     "user_login",
			fileName: "testdata/instances_1.txt",
		},
		{
			name:     "user_login_permission_changed",
			fileName: "testdata/instances_2.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(tc.fileName)
			if err != nil {
				t.Fatalf("Failed to read from file: %v", err)
			}
			sp := &StatPerf{}
			instances := sp.parseInstances(string(content))

			if len(instances) != 6 {
				t.Errorf("Expected 6 instances, got %d", len(instances))
			}
		})
	}
}

func TestParseData(t *testing.T) {
	testCases := []testCase{
		{
			name:     "admin_login",
			fileName: "testdata/data.txt",
		}, {
			name:     "user_login",
			fileName: "testdata/data_1.txt",
		},
		{
			name:     "user_login_permission_changed",
			fileName: "testdata/data_2.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			content, err := os.ReadFile(tc.fileName)
			if err != nil {
				t.Fatalf("Failed to read from file: %v", err)
			}
			sp := &StatPerf{}

			res, err := sp.parseData(string(content), slog.Default())
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !res.IsArray() {
				t.Fatalf("Expected JSON result to be an array, got: %v", res.Type)
			}

			groupsCount := len(res.Array())
			if groupsCount < 6 {
				t.Errorf("Expected 6 groups, got %d", groupsCount)
			}

			foundInstance := false
			for _, group := range res.Array() {
				if group.Get("instance_name").String() == "hs_1" {
					foundInstance = true
					break
				}
			}
			if !foundInstance {
				t.Errorf("Expected to find a group with instance_name 'hs_1'")
			}
		})
	}
}
