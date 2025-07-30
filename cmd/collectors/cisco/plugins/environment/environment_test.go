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
				RedunMode:      "Redundant",
				OperationMode:  "Redundant",
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
				RedunMode:      "PS-Redundant",
				OperationMode:  "PS-Redundant",
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

func TestFanSpeed(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  FanModel
	}{
		{
			name:  "3000 series",
			input: "NX3132V_9.3.10-show-environment-detail.json",
			want: FanModel{
				Fans: []*FanData{
					{Name: "Fan1(sys_fan1)", Model: "NXA-FAN-30CFM-F", Speed: 83, TrayFanNum: "fan1", Status: "ok"},
					{Name: "Fan1(sys_fan1)", Model: "NXA-FAN-30CFM-F", Speed: 64, TrayFanNum: "fan2", Status: "ok"},
					{Name: "Fan2(sys_fan2)", Model: "NXA-FAN-30CFM-F", Speed: 84, TrayFanNum: "fan1", Status: "ok"},
					{Name: "Fan2(sys_fan2)", Model: "NXA-FAN-30CFM-F", Speed: 64, TrayFanNum: "fan2", Status: "ok"},
					{Name: "Fan3(sys_fan3)", Model: "NXA-FAN-30CFM-F", Speed: 88, TrayFanNum: "fan1", Status: "ok"},
					{Name: "Fan3(sys_fan3)", Model: "NXA-FAN-30CFM-F", Speed: 62, TrayFanNum: "fan2", Status: "ok"},
					{Name: "Fan4(sys_fan4)", Model: "NXA-FAN-30CFM-F", Speed: 83, TrayFanNum: "fan1", Status: "ok"},
					{Name: "Fan4(sys_fan4)", Model: "NXA-FAN-30CFM-F", Speed: 63, TrayFanNum: "fan2", Status: "ok"},
					{Name: "Fan_in_PS1", Model: "--", Speed: -1, Status: "ok"},
					{Name: "Fan_in_PS2", Model: "--", Speed: -1, Status: "shutdown"},
				},
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
			got := NewFanModel(gjson.ParseBytes(data), slog.Default())
			diff1 := cmp.Diff(tt.want, got)
			if diff1 != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff1)
			}
		})
	}
}
