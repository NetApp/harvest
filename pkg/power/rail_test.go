package power

import "testing"

func TestRailFromLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  Rail
	}{
		{name: "empty", label: "", want: RailUnknown},
		{name: "input", label: "Input", want: RailInput},
		{name: "vin", label: "VIN", want: RailInput},
		{name: "iin", label: "iin", want: RailInput},
		{name: "output", label: "output", want: RailOutput},
		{name: "vout", label: "VOUT", want: RailOutput},
		{name: "iout", label: "iout", want: RailOutput},
		{name: "both", label: "input/output", want: RailUnknown},
		{name: "vin-vout", label: "vin vout", want: RailUnknown},
		{name: "no-match", label: "12v", want: RailUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RailFromLabel(tt.label); got != tt.want {
				t.Fatalf("RailFromLabel(%q)=%v want %v", tt.label, got, tt.want)
			}
		})
	}
}

func TestClassifyRailFromLabels(t *testing.T) {
	tests := []struct {
		name   string
		labels []string
		want   Rail
	}{
		{name: "first-input", labels: []string{"", "VIN"}, want: RailInput},
		{name: "first-output", labels: []string{"vout", "vin"}, want: RailOutput},
		{name: "skip-unknown", labels: []string{"unknown", "input"}, want: RailInput},
		{name: "none", labels: []string{"", " ", "foo"}, want: RailUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifyRailFromLabels(tt.labels...); got != tt.want {
				t.Fatalf("ClassifyRailFromLabels(%v)=%v want %v", tt.labels, got, tt.want)
			}
		})
	}
}

func TestResolveRail(t *testing.T) {
	tests := []struct {
		name         string
		voltageRail  Rail
		currentRail  Rail
		wantResolved Rail
	}{
		{name: "input-only", voltageRail: RailInput, currentRail: RailUnknown, wantResolved: RailInput},
		{name: "output-only", voltageRail: RailUnknown, currentRail: RailOutput, wantResolved: RailOutput},
		{name: "match-input", voltageRail: RailInput, currentRail: RailInput, wantResolved: RailInput},
		{name: "match-output", voltageRail: RailOutput, currentRail: RailOutput, wantResolved: RailOutput},
		{name: "conflict", voltageRail: RailInput, currentRail: RailOutput, wantResolved: RailUnknown},
		{name: "unknown", voltageRail: RailUnknown, currentRail: RailUnknown, wantResolved: RailUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveRail(tt.voltageRail, tt.currentRail); got != tt.wantResolved {
				t.Fatalf("ResolveRail(%v,%v)=%v want %v", tt.voltageRail, tt.currentRail, got, tt.wantResolved)
			}
		})
	}
}
