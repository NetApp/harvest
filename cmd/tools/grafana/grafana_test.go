package grafana

import (
	"testing"
)

func TestCheckVersion(t *testing.T) {

	inputVersion := []string{"7.2.3.4", "abc.1.3", "4.5.4", "7.1.0", "7.5.5"}
	expectedOutPut := []bool{true, false, false, true, true}
	// version length greater than 3

	for i, s := range inputVersion {
		c := checkVersion(s)
		if c != expectedOutPut[i] {
			t.Errorf("Expected %t but got %t for input %s", expectedOutPut[i], c, inputVersion[i])
		}
	}
}

func TestHttpsAddr(t *testing.T) {
	opts.addr = "https://1.1.1.1:3000"
	adjustOptions()
	if opts.addr != "https://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "https://1.1.1.1:3000", opts.addr)
	}

	opts.addr = "https://1.1.1.1:3000"
	opts.useHttps = false // addr takes precedence over useHttps
	adjustOptions()
	if opts.addr != "https://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "https://1.1.1.1:3000", opts.addr)
	}

	opts.addr = "http://1.1.1.1:3000"
	adjustOptions()
	if opts.addr != "http://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "http://1.1.1.1:3000", opts.addr)
	}

	// Old way of specifying https
	opts.addr = "http://1.1.1.1:3000"
	opts.useHttps = true
	adjustOptions()
	if opts.addr != "https://1.1.1.1:3000" {
		t.Errorf("Expected opts.addr to be %s but got %s", "https://1.1.1.1:3000", opts.addr)
	}
}
