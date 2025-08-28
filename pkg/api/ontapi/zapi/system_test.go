package zapi

import (
	"fmt"
	"github.com/netapp/harvest/v2/assert"
	"testing"
)

func Test_7modeParsing(t *testing.T) {
	type test struct {
		name    string
		release string
		want    string
		wantErr bool
	}

	tests := []test{
		{name: "issue-376", release: "NetApp Release 8.2P4 7-Mode: Tue Oct 1 11:24:04 PDT 2013", want: "8.2.4"},
		{name: "7.0.1", release: "NetApp Release 7.0.0.1: Wed Mar 2 22:20:44 PST 2005", want: "7.0.0"},
		{name: "7.0.4", release: "NetApp Release 7.0.4: Sun Feb  5 00:52:53 PST 2006", want: "7.0.4"},
		{name: "6.4.1", release: "NetApp Release 6.4R1: Thu Mar 13 23:56:30 PST 2003", want: "6.4.1"},
		{name: "7.3.1", release: "NetApp Release 7.3.1: Thu Jan 8 01:31:42 PST 2009 ", want: "7.3.1"},
		{name: "fail", release: "cromulent house", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v0, v1, v2, err := parse7mode(tt.release)
			if err != nil {
				assert.True(t, tt.wantErr)
				if tt.wantErr {
					return
				}
			}
			got := fmt.Sprintf("%d.%d.%d", v0, v1, v2)
			assert.Equal(t, got, tt.want)
		})
	}
}
