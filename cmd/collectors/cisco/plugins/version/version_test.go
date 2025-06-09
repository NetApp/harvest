package version

import (
	"github.com/google/go-cmp/cmp"
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
			want: rcf{Version: "v1.6b", Filename: "NX3132Q-V_v2.00_Switch-A1.txt"},
		},
		{
			name: "Generator with version", banner: `* Filename  : NX3232_v1.90-X1_Switch-B2.txt
* Date      : Generator version: v1.4a_2022-mm-dd_001, file creation: 2024-02-15, 10:28:44`,
			want: rcf{Version: "v1.4a", Filename: "NX3232_v1.90-X1_Switch-B2.txt"},
		},
		{
			name: "Generator no filename", banner: `N3K NetApp Reference Configuration File (RCF) version 1.1-24p10g-26p40g (2016-08-23)`,
			want: rcf{Version: "1.1-24p10g-26p40g", Filename: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRCF(tt.banner)
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
