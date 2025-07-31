package lldp

import (
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"os"
	"testing"
)

func TestNewLLDPModel(t *testing.T) {
	tests := []struct {
		name string
		want Model
	}{
		{
			name: "1",
			want: Model{
				Capabilities:   []string{"Station"},
				ChassisID:      "dead.beef.7fef",
				RemotePlatform: "AFF-A800, NetApp Release 9.13.1P9: Fri Apr 19 13:13:02 EDT 2024",
				RemoteName:     "na2a-mc1-nbg3",
				LocalPort:      "mgmt0",
				RemotePort:     "e0M",
				TTL:            114,
			},
		},
	}

	// Read the file from the testdata directory
	filename := "testdata/lldp.json"
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("failed to read %s file: %v", filename, err)
	}

	result := gjson.ParseBytes(data)
	jsons := result.Get("TABLE_nbor_detail.ROW_nbor_detail").Array()

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLLDPModel(jsons[i])
			diff1 := cmp.Diff(tt.want, got)
			if diff1 != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff1)
			}
		})
	}
}
