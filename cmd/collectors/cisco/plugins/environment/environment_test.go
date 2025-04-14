package environment

import (
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"os"
	"testing"
)

func TestNewPowerModel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  PowerModel
	}{
		{
			name:  "3000 series",
			input: "NX3132V_9.3.10-show-environment.json",
			want: PowerModel{
				PowerSupplies: []PowerSupply{
					{Num: "1", Model: "N2200-PAC-400W", Status: "ok", ActualIn: 396},
					{Num: "2", Model: "N2200-PAC-400W", Status: "ok", ActualIn: 396},
				},
				TotalPowerDraw: 348,
			},
		}, {
			name:  "9000 series",
			input: "N9K-C9336C-FX2_10.2.3-show-environment.json",
			want: PowerModel{
				PowerSupplies: []PowerSupply{
					{Num: "1", Model: "NXA-PAC-1100W-PE2", Status: "ok", ActualIn: 128, ActualOut: 112, TotalCapacity: 1100},
					{Num: "2", Model: "NXA-PAC-1100W-PE2", Status: "ok", ActualIn: 148, ActualOut: 132, TotalCapacity: 1100},
				},
				TotalPowerDraw: 244,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the file from the testdata directory
			data, err := os.ReadFile("testdata/" + tt.input)
			if err != nil {
				t.Errorf("failed to read %s file: %v", tt.input, err)
			}
			got := NewPowerModel(gjson.ParseBytes(data), slog.Default())
			diff1 := cmp.Diff(tt.want, got)
			if diff1 != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff1)
			}
		})
	}
}
