package optic

import (
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"os"
	"testing"
)

func TestNewOpticModels(t *testing.T) {
	tests := []struct {
		name    string
		want    Model
		wantErr bool
	}{
		{
			name: "9336c", want: Model{Name: ""},
		},
		{
			name: "9336c", want: Model{Name: ""},
		},
		{
			name: "9336c", want: Model{Name: "Ethernet1/15", RxPower: -3.72, TxPower: -2.38},
		},
		{
			name: "9336c", want: Model{Name: ""},
		},
		{
			name: "9336c", want: Model{Name: "Ethernet1/22/4", RxPower: 0, TxPower: -2.80},
		},
	}

	// Read the file from the testdata directory
	filename := "testdata/N9K-C9336C-FX2_10.3.4-show_interface_transceiver_details.json"
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("failed to read %s file: %v", filename, err)
	}

	result := gjson.ParseBytes(data)
	jsons := result.Get("TABLE_interface.ROW_interface").Array()

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOpticModel(jsons[i], slog.Default())
			diff1 := cmp.Diff(tt.want, got)
			if diff1 != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff1)
			}
		})
	}
}
