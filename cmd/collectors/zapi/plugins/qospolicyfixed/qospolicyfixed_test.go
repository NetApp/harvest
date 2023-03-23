package qospolicyfixed

import (
	"strconv"
	"testing"
)

func Test_zapiXputToRest(t *testing.T) {
	tests := []struct {
		zapi  string
		want  MaxXput
		isErr bool
	}{
		// Adaptive QOS uses this form, test it here too
		{zapi: "6144IOPS/TB", want: MaxXput{IOPS: "6144", Mbps: "0"}},

		{zapi: "100IOPS", want: MaxXput{IOPS: "100", Mbps: "0"}},
		{zapi: "100iops", want: MaxXput{IOPS: "100", Mbps: "0"}},
		{zapi: "111111IOPS", want: MaxXput{IOPS: "111111", Mbps: "0"}},
		{zapi: "0", want: MaxXput{IOPS: "0", Mbps: "0"}},
		{zapi: "INF", want: MaxXput{IOPS: "0", Mbps: "0"}},

		{zapi: "1GB/s", want: MaxXput{IOPS: "0", Mbps: "1024"}},
		{zapi: "100B/s", want: MaxXput{IOPS: "0", Mbps: "0"}},
		{zapi: "10KB/s", want: MaxXput{IOPS: "0", Mbps: "0"}},
		{zapi: "1mb/s", want: MaxXput{IOPS: "0", Mbps: "1"}},
		{zapi: "1tb/s", want: MaxXput{IOPS: "0", Mbps: "1048576"}},
		{zapi: "1000KB/s", want: MaxXput{IOPS: "0", Mbps: "0"}},
		{zapi: "15000IOPS,468.8MB/s", want: MaxXput{IOPS: "15000", Mbps: "468"}},
		{zapi: "50000IOPS,1.53GB/s", want: MaxXput{IOPS: "50000", Mbps: "1566"}},

		{zapi: "1 foople/s", want: MaxXput{IOPS: "0", Mbps: "0"}, isErr: true},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := ZapiXputToRest(tt.zapi)
			if tt.isErr && err == nil {
				t.Errorf("ZapiXputToRest(%s) got=%+v, want err but got err=%s", tt.zapi, got, err)
				return
			}
			if got.IOPS != tt.want.IOPS || got.Mbps != tt.want.Mbps {
				t.Errorf("ZapiXputToRest(%s) got=%+v, want=%+v", tt.zapi, got, tt.want)
			}
		})
	}
}
