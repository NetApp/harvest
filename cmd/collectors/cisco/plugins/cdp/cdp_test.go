package cdp

import (
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"os"
	"testing"
)

func TestNewCDPModel(t *testing.T) {
	tests := []struct {
		name    string
		want    Model
		wantErr bool
	}{
		{
			name: "1",
			want: Model{
				Capabilities:       []string{"Supports-STP-Dispute", "router", "switch"},
				RemoteName:         "lamb-mc1-nbg3(FLM12345678)",
				LocalInterfaceMAC:  "de:ad:be:ef:d5:f0",
				RemotePlatform:     "N9K-C9336C-FX2",
				LocalPort:          "mgmt0",
				RemoteInterfaceMAC: "de:ad:be:ef:d5:f0",
				TTL:                179,
				RemoteVersion:      "Cisco Nexus Operating System (NX-OS) Software, Version 10.2(4)",
				RemotePort:         "mgmt0",
			},
		},
		{
			name: "2",
			want: Model{
				Capabilities:       []string{"host"},
				RemoteName:         "na2a-mc1-nbg3",
				LocalInterfaceMAC:  "de:ad:be:ef:d5:f8",
				RemotePlatform:     "AFF-A800",
				LocalPort:          "Ethernet1/1",
				RemoteInterfaceMAC: "10:17:43:78:7e:90",
				TTL:                157,
				RemoteVersion:      "NetApp Release 9.13.1P9: Fri Apr 19 13:13:02 EDT 2024",
				RemotePort:         "e1a",
			},
		},
	}

	// Read the file from the testdata directory
	filename := "testdata/cdp.json"
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("failed to read %s file: %v", filename, err)
	}

	result := gjson.ParseBytes(data)
	jsons := result.Get("TABLE_cdp_neighbor_detail_info.ROW_cdp_neighbor_detail_info").Array()

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCDPModel(jsons[i])
			diff1 := cmp.Diff(tt.want, got)
			if diff1 != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff1)
			}
		})
	}
}
