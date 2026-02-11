package version

import (
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/assert"
	"testing"
)

func Test_parseRCF(t *testing.T) {
	tests := []struct {
		name   string
		banner string
		want   rcf
	}{
		{
			name: "Empty", banner: ``, want: rcf{},
		},
		{
			name: "Download", banner: ` * Filename : NX3232C-RCF-v1.10-Cluster-HA.txt
* Date     : 10-04-2023
* Version  : v1.10`,
			want: rcf{Version: "v1.10", Filename: "NX3232C-RCF-v1.10-Cluster-HA.txt"},
		},
		{
			name: "Generator", banner: `* Filename  : NX9336C-FX2_v2.10_Switch-B2.txt
* Date      : Generator: v1.6c 2023-12-05_001, file creation: 2024-07-29, 11:19:36`,
			want: rcf{Version: "v1.6c", Filename: "NX9336C-FX2_v2.10_Switch-B2.txt"},
		},
		{
			name: "Storage Switch", banner: `Filename  : Nexus-9336C-RCF-v1.8-Storage.txt
* Version   : v1.8`,
			want: rcf{Version: "v1.8", Filename: "Nexus-9336C-RCF-v1.8-Storage.txt"},
		},
		{
			name: "Generator with underscore", banner: `* Filename  : NX3132Q-V_v2.00_Switch-A1.txt
* Date      : Generator: v1.6b_2023-07-18_001, file creation: 2024-02-15, 10:28:44`,
			want: rcf{Version: "v1.6b_2023-07-18_001", Filename: "NX3132Q-V_v2.00_Switch-A1.txt"},
		},
		{
			name: "Generator with version", banner: `* Filename  : NX3232_v1.90-X1_Switch-B2.txt
* Date      : Generator version: v1.4a_2022-mm-dd_001, file creation: 2024-02-15, 10:28:44`,
			want: rcf{Version: "v1.4a_2022-mm-dd_001", Filename: "NX3232_v1.90-X1_Switch-B2.txt"},
		},
		{
			name: "Generator no filename", banner: `N3K NetApp Reference Configuration File (RCF) version 1.1-24p10g-26p40g (2016-08-23)`,
			want: rcf{Version: "1.1-24p10g-26p40g", Filename: ""},
		},
		{
			name: "With carriage returns", banner: "\n******************************************************************************\r\n* NetApp Reference Configuration File (RCF)\r\n*\r\n* Switch    : NX3232C (direct storage, L2 Networks, direct ISL)\r\n* Filename  : NX3232C_v2.20_Switch-B1.txt\r\n* Date      : Generator: v1.7a 2024-05-23_001, file creation: 2024-10-10, 07:37:31\r\n*\r\n* Platforms : 4: MetroCluster 1 : FAS8200, AFF-A300\r\n*\r\n* Port Usage:\r\n* Ports  1- 2: Ports not used\r\n* Ports  3- 4: Ports not used\r\n* Ports  5- 6: Ports not used\r\n* Ports  7- 8: Intra-Cluster ISL Ports, local cluster, VLAN 204\r\n* Ports  9-10: Ports not used\r\n* Ports 11-12: Ports not used\r\n* Ports 13-14: Ports not used\r\n* Ports 15-20: MetroCluster ISL, VLAN 10, Port Channel 10, 40G / 100G\r\n* Ports 21-24: MetroCluster ISL, VLAN 10, Port Channel 11, 4x10G breakout\r\n* Ports 25-26: MetroCluster 1, Node Ports, VLAN 10, 4x25G breakout\r\n* Ports 27-28: Ports not used\r\n* Ports 29-30: Intra-Cluster Node Ports, Cluster: MetroCluster 1, VLAN 204, 4x10G breakout\r\n* Ports 31-32: Ports not used\r\n*\r\n* 10G Port Usage:\r\n* Ports 33-34: Ports not used\r\n*\r\n******************************************************************************\r\n",
			want: rcf{Version: "v1.7a", Filename: "NX3232C_v2.20_Switch-B1.txt"},
		},
		{
			name: "Generator no filename", banner: "NX3232 NetApp Reference Configuration File (RCF) version 1.0-24p10g-26p100 (2017-09-07) *\n*",
			want: rcf{Version: "1.0-24p10g-26p100", Filename: ""},
		},
		{
			name: "", banner: "\n* NetApp Reference Configuration File (RCF)\r\n*\r\n* Switch   : NX9336C-FX2\r\n* Filename : NX9336C-FX2-RCF-1.13-1-Cluster-HA-Breakout.txt\r\n* Date     : 05-22-2025\r\n* Version  : v1.13\r\n* Port Usage:\r\n* Ports  1- 3: Breakout mode (4x10G) Intra-Cluster/HA Ports, int e1/1/1-4, e1/2/1-4, e1/3/1-4\r\n* Ports  4- 6: Breakout mode (4x25G) Intra-Cluster/HA Ports, int e1/4/1-4, e1/5/1-4, e1/6/1-4\r\n* Ports  7-34: 40/100GbE Intra-Cluster/HA Ports, int e1/7-34\r\n* Ports 35-36: Intra-Cluster ISL Ports, int e1/35-36\r\n*\r\n* IMPORTANT NOTES\r\n*\r\n* Interface port-channel999 is reserved to identify the version of this file.\r\n* This RCF supports Clustering, HA, RDMA, and DCTCP using a single port profile.",
			want: rcf{Version: "v1.13", Filename: "NX9336C-FX2-RCF-1.13-1-Cluster-HA-Breakout.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRCF(tt.banner)
			diff := cmp.Diff(got, tt.want)
			assert.Equal(t, diff, "")
		})
	}
}
