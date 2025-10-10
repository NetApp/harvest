package qospolicyfixed

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"strconv"
	"testing"
)

func Test_zapiXputToRest(t *testing.T) {
	tests := []struct {
		zapi  string
		want  collectors.Xput
		isErr bool
	}{
		// Adaptive QOS uses this form, test it here too
		{zapi: "6144IOPS/TB", want: collectors.Xput{IOPS: "6144", Mbps: ""}},
		{zapi: "6144IOPS/GB", want: collectors.Xput{IOPS: "6144000", Mbps: ""}},

		{zapi: "100IOPS", want: collectors.Xput{IOPS: "100", Mbps: ""}},
		{zapi: "100iops", want: collectors.Xput{IOPS: "100", Mbps: ""}},
		{zapi: "111111IOPS", want: collectors.Xput{IOPS: "111111", Mbps: ""}},
		{zapi: "0", want: collectors.Xput{IOPS: "", Mbps: ""}},
		{zapi: "", want: collectors.Xput{IOPS: "", Mbps: ""}},
		{zapi: "INF", want: collectors.Xput{IOPS: "", Mbps: ""}},

		{zapi: "1GB/s", want: collectors.Xput{IOPS: "", Mbps: "1000"}},
		{zapi: "100B/s", want: collectors.Xput{IOPS: "", Mbps: "0"}},
		{zapi: "10KB/s", want: collectors.Xput{IOPS: "", Mbps: "0.01"}},
		{zapi: "1mb/s", want: collectors.Xput{IOPS: "", Mbps: "1"}},
		{zapi: "1tb/s", want: collectors.Xput{IOPS: "", Mbps: "1000000"}},
		{zapi: "1000KB/s", want: collectors.Xput{IOPS: "", Mbps: "1"}},
		{zapi: "15000IOPS,468.8MB/s", want: collectors.Xput{IOPS: "15000", Mbps: "468.8"}},
		{zapi: "50000IOPS,1.53GB/s", want: collectors.Xput{IOPS: "50000", Mbps: "1530"}},

		{zapi: "1 foople/s", want: collectors.Xput{IOPS: "", Mbps: ""}, isErr: true},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := collectors.ZapiXputToRest(tt.zapi)
			if err != nil {
				assert.True(t, tt.isErr)
				return
			}
			assert.Equal(t, got.IOPS, tt.want.IOPS)
			assert.Equal(t, got.Mbps, tt.want.Mbps)
		})
	}
}
