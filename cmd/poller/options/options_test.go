package options

import (
	"github.com/netapp/harvest/v2/assert"
	"testing"
)

// Make sure that if --config is passed on the cmdline that it is not overwritten
// See https://github.com/NetApp/harvest/issues/28
func TestConfigPath(t *testing.T) {
	want := "foo"
	options := New(WithConfigPath(want))

	assert.Equal(t, options.Config, want)
}
