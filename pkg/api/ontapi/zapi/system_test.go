package zapi

import (
	"testing"
)

func Test_7modeParsing(t *testing.T) {
	type test struct {
		name    string
		release string
		want    system
		wantErr bool
	}

	tests := []test{
		{name: "issue-376", release: "NetApp Release 8.2P4 7-Mode: Tue Oct 1 11:24:04 PDT 2013", want: system{version: [3]int{8, 2, 4}}},
		{name: "7.0.1", release: "NetApp Release 7.0.0.1: Wed Mar 2 22:20:44 PST 2005", want: system{version: [3]int{7, 0, 0}}},
		{name: "7.0.4", release: "NetApp Release 7.0.4: Sun Feb  5 00:52:53 PST 2006", want: system{version: [3]int{7, 0, 4}}},
		{name: "6.4.1", release: "NetApp Release 6.4R1: Thu Mar 13 23:56:30 PST 2003", want: system{version: [3]int{6, 4, 1}}},
		{name: "7.3.1", release: "NetApp Release 7.3.1: Thu Jan 8 01:31:42 PST 2009 ", want: system{version: [3]int{7, 3, 1}}},
		{name: "fail", release: "cromulent house", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := system{}
			err := s.parse7mode(tt.release)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("got error=%v, want %v", err, tt.want)

			}
			if s.version[0] == 0 {
				t.Errorf("got zero version=%v, want %v", s.version, tt.want.version)
			}
			if s.version[0] != tt.want.version[0] || s.version[1] != tt.want.version[1] || s.version[2] != tt.want.version[2] {
				t.Errorf("got version=%v, want %v", s.version, tt.want.version)
			}
		})
	}
}
