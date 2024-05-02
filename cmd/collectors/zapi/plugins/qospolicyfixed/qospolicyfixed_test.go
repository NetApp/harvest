package qospolicyfixed

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"strconv"
	"testing"
)

func Test_zapiXputToRest(t *testing.T) {
	tests := []struct {
		zapi  string
		want  collectors.MaxXput
		isErr bool
	}{
		// Adaptive QOS uses this form, test it here too
		{zapi: "6144IOPS/TB", want: collectors.MaxXput{IOPS: "6144", Mbps: ""}},
		{zapi: "6144IOPS/GB", want: collectors.MaxXput{IOPS: "6144000", Mbps: ""}},

		{zapi: "100IOPS", want: collectors.MaxXput{IOPS: "100", Mbps: ""}},
		{zapi: "100iops", want: collectors.MaxXput{IOPS: "100", Mbps: ""}},
		{zapi: "111111IOPS", want: collectors.MaxXput{IOPS: "111111", Mbps: ""}},
		{zapi: "0", want: collectors.MaxXput{IOPS: "", Mbps: ""}},
		{zapi: "", want: collectors.MaxXput{IOPS: "", Mbps: ""}},
		{zapi: "INF", want: collectors.MaxXput{IOPS: "", Mbps: ""}},

		{zapi: "1GB/s", want: collectors.MaxXput{IOPS: "", Mbps: "1000"}},
		{zapi: "100B/s", want: collectors.MaxXput{IOPS: "", Mbps: "0"}},
		{zapi: "10KB/s", want: collectors.MaxXput{IOPS: "", Mbps: "0.01"}},
		{zapi: "1mb/s", want: collectors.MaxXput{IOPS: "", Mbps: "1"}},
		{zapi: "1tb/s", want: collectors.MaxXput{IOPS: "", Mbps: "1000000"}},
		{zapi: "1000KB/s", want: collectors.MaxXput{IOPS: "", Mbps: "1"}},
		{zapi: "15000IOPS,468.8MB/s", want: collectors.MaxXput{IOPS: "15000", Mbps: "468.8"}},
		{zapi: "50000IOPS,1.53GB/s", want: collectors.MaxXput{IOPS: "50000", Mbps: "1530"}},

		{zapi: "1 foople/s", want: collectors.MaxXput{IOPS: "", Mbps: ""}, isErr: true},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := collectors.ZapiXputToRest(tt.zapi)
			if err != nil && !tt.isErr {
				t.Errorf("ZapiXputToRest(%s) got=%+v, want no err but got err=%s", tt.zapi, got, err)
				return
			}
			if got.IOPS != tt.want.IOPS || got.Mbps != tt.want.Mbps {
				t.Errorf("ZapiXputToRest(%s) got=%+v, want=%+v", tt.zapi, got, tt.want)
			}
		})
	}
}
