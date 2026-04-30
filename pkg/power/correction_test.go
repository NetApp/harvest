package power

import "testing"

func TestNeedsPsuPowerDrawnCorrection(t *testing.T) {
	tests := []struct {
		name            string
		moduleType      string
		firmwareVersion string
		want            bool
	}{
		// IOM12E boundary: fix is in 0240
		{name: "IOM12E below threshold", moduleType: "IOM12E", firmwareVersion: "0239", want: true},
		{name: "IOM12E at threshold", moduleType: "IOM12E", firmwareVersion: "0240", want: false},
		{name: "IOM12E above threshold", moduleType: "IOM12E", firmwareVersion: "0300", want: false},

		// IOM12 boundary: fix is in 0290
		{name: "IOM12 below threshold", moduleType: "IOM12", firmwareVersion: "0289", want: true},
		{name: "IOM12 at threshold", moduleType: "IOM12", firmwareVersion: "0290", want: false},
		{name: "IOM12 above threshold", moduleType: "IOM12", firmwareVersion: "0350", want: false},

		// Case insensitivity
		{name: "lowercase iom12e", moduleType: "iom12e", firmwareVersion: "0100", want: true},
		{name: "lowercase iom12", moduleType: "iom12", firmwareVersion: "0100", want: true},
		{name: "mixed case IOM12E", moduleType: "Iom12E", firmwareVersion: "0241", want: false},

		// Unaffected module types
		{name: "NSM100 not affected", moduleType: "NSM100", firmwareVersion: "0100", want: false},
		{name: "IOM6 not affected", moduleType: "IOM6", firmwareVersion: "0001", want: false},
		{name: "unknown module", moduleType: "UNKNOWN", firmwareVersion: "0001", want: false},
		{name: "empty module type", moduleType: "", firmwareVersion: "0100", want: false},

		// Empty/invalid firmware versions
		{name: "empty firmware", moduleType: "IOM12E", firmwareVersion: "", want: false},
		{name: "non-numeric firmware", moduleType: "IOM12E", firmwareVersion: "abc", want: false},
		{name: "firmware with spaces", moduleType: "IOM12E", firmwareVersion: " 0100 ", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsPsuPowerDrawnCorrection(tt.moduleType, tt.firmwareVersion)
			if got != tt.want {
				t.Fatalf("NeedsPsuPowerDrawnCorrection(%q, %q) = %v, want %v",
					tt.moduleType, tt.firmwareVersion, got, tt.want)
			}
		})
	}
}

func TestMinFirmwareVersion(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want string
	}{
		{name: "a lower", a: "0100", b: "0200", want: "0100"},
		{name: "b lower", a: "0200", b: "0100", want: "0100"},
		{name: "equal", a: "0240", b: "0240", want: "0240"},
		{name: "a empty", a: "", b: "0240", want: "0240"},
		{name: "b empty", a: "0240", b: "", want: "0240"},
		{name: "both empty", a: "", b: "", want: ""},
		{name: "a invalid", a: "abc", b: "0240", want: "0240"},
		{name: "b invalid", a: "0240", b: "xyz", want: "0240"},
		{name: "both invalid", a: "abc", b: "xyz", want: "xyz"},
		{name: "leading spaces", a: " 0239 ", b: "0240", want: " 0239 "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MinFirmwareVersion(tt.a, tt.b)
			if got != tt.want {
				t.Fatalf("MinFirmwareVersion(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
