package collector

import (
	"log/slog"
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func newParamCollector(params map[string]string) *AbstractCollector {
	n := node.New([]byte("root"))
	for k, v := range params {
		n.NewChildS(k, v)
	}

	return &AbstractCollector{Params: n, Logger: slog.Default()}
}

func TestLoadParam(t *testing.T) {
	c := newParamCollector(map[string]string{
		"client_timeout":  "30",
		"batch_size":      "700",
		"max_records":     "9000000000",
		"latency_io_reqd": "1.5",
		"instance_key":    "uuid",
		"is_enabled":      "true",
		"empty":           "",
		"malformed":       "not-a-number",
	})

	// T is inferred from the default -- no explicit instantiation anywhere
	assert.Equal(t, c.LoadParam("client_timeout", 5), 30)
	assert.Equal(t, c.LoadParam("batch_size", 500), 700)
	assert.Equal(t, c.LoadParam("max_records", int64(1)), int64(9000000000))
	assert.Equal(t, c.LoadParam("latency_io_reqd", 0.0), 1.5)
	assert.Equal(t, c.LoadParam("instance_key", "name"), "uuid")
	assert.Equal(t, c.LoadParam("is_enabled", false), true)
}

func TestLoadParamFallsBackToDefault(t *testing.T) {
	c := newParamCollector(map[string]string{
		"empty":     "",
		"malformed": "not-a-number",
	})

	// absent parameter
	assert.Equal(t, c.LoadParam("nope", 500), 500)
	assert.Equal(t, c.LoadParam("nope", "fallback"), "fallback")
	assert.Equal(t, c.LoadParam("nope", false), false)

	// present but empty
	assert.Equal(t, c.LoadParam("empty", 500), 500)
	assert.Equal(t, c.LoadParam("empty", int64(7)), int64(7))

	// present but unparsable -- must not panic, must use the default
	assert.Equal(t, c.LoadParam("malformed", 500), 500)
	assert.Equal(t, c.LoadParam("malformed", int64(7)), int64(7))
	assert.Equal(t, c.LoadParam("malformed", 2.5), 2.5)
	assert.Equal(t, c.LoadParam("malformed", true), true)

	// a string parameter is never malformed
	assert.Equal(t, c.LoadParam("malformed", "fallback"), "not-a-number")
}
