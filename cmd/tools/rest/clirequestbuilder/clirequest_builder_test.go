package clirequestbuilder

import (
	"encoding/json"
	"testing"
)

func TestCLIRequestBuilder(t *testing.T) {
	t.Run("Build with all fields set", func(t *testing.T) {
		builder := New().
			BaseSet("baseSet").
			APIPath("apiPath").
			Query("query").
			Object("object").
			Filter("filter").
			Fields([]string{"field1", "field2"}).
			Counters([]string{"counter1", "counter2"}).
			Instances([]string{"instance1", "instance2"})

		result, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		var payload map[string]string
		if err := json.Unmarshal(result, &payload); err != nil {
			t.Fatalf("failed to unmarshal result: %v", err)
		}

		expected := "baseSet query -object object -filter filter -fields field1,field2 -instance instance1|instance2 -counter counter1|counter2"
		if payload["input"] != expected {
			t.Errorf("expected %v, got %v", expected, payload["input"])
		}
	})

	t.Run("Build with mandatory fields only", func(t *testing.T) {
		builder := New().
			Query("query")

		result, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		var payload map[string]string
		if err := json.Unmarshal(result, &payload); err != nil {
			t.Fatalf("failed to unmarshal result: %v", err)
		}

		expected := "query"
		if payload["input"] != expected {
			t.Errorf("expected %v, got %v", expected, payload["input"])
		}
	})

	t.Run("Build with missing mandatory fields", func(t *testing.T) {
		builder := New()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error, got none")
		}

		expectedError := "query must be provided"
		if err.Error() != expectedError {
			t.Errorf("expected error %v, got %v", expectedError, err.Error())
		}
	})
}
