package clirequestbuilder

import (
	"encoding/json"
	"github.com/netapp/harvest/v2/assert"
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
		assert.Nil(t, err)

		var payload map[string]string
		err = json.Unmarshal(result, &payload)
		assert.Nil(t, err)

		expected := "baseSet query -object object -filter filter -fields field1,field2 -instance instance1|instance2 -counter counter1|counter2"
		assert.Equal(t, payload["input"], expected)
	})

	t.Run("Build with mandatory fields only", func(t *testing.T) {
		builder := New().
			Query("query")

		result, err := builder.Build()
		assert.Nil(t, err)

		var payload map[string]string
		err = json.Unmarshal(result, &payload)
		assert.Nil(t, err)

		expected := "query"
		assert.Equal(t, payload["input"], expected)
	})

	t.Run("Build with missing mandatory fields", func(t *testing.T) {
		builder := New()

		_, err := builder.Build()
		assert.NotNil(t, err)

		expectedError := "query must be provided"
		assert.Equal(t, err.Error(), expectedError)
	})
}
