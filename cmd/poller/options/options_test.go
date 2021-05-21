package options

import "testing"

// Make sure that if --config is passed on the cmdline that it is not overwritten
// See https://github.com/NetApp/harvest/issues/28
func TestConfigPath(t *testing.T) {
	want := "foo"
	options := Options{Config: want}
	SetPathsAndHostname(&options)

	if options.Config != want {
		t.Fatalf(`options.Config expected=[%q], actual was=[%q]`, want, options.Config)
	}
}
