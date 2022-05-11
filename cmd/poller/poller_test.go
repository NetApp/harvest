package main

import (
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
)

func TestPingParsing(t *testing.T) {
	poller := Poller{}

	type test struct {
		name string
		out  string
		ping float32
		isOK bool
	}

	tests := []test{
		{
			name: "NotBusy",
			ping: 0.032,
			isOK: true,
			out: `PING 127.0.0.1 (127.0.0.1) 56(84) bytes of data.

	--- 127.0.0.1 ping statistics ---
	1 packets transmitted, 1 received, 0% packet loss, time 0ms
	rtt min/avg/max/mdev = 0.032/0.032/0.032/0.000 ms`,
		},
		{
			name: "BusyBox",
			ping: 0.088,
			isOK: true,
			out: `PING 127.0.0.1 (127.0.0.1): 56 data bytes

--- 127.0.0.1 ping statistics ---
1 packets transmitted, 1 packets received, 0% packet loss
round-trip min/avg/max = 0.088/0.088/0.088 ms`,
		},
		{
			name: "BadInput",
			ping: 0,
			isOK: false,
			out:  `foo`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ping, b := poller.parsePing(tt.out)
			if ping != tt.ping {
				t.Errorf("parsePing ping got = %v, want %v", ping, tt.ping)
			}
			if b != tt.isOK {
				t.Errorf("parsePing isOK got = %v, want %v", b, tt.isOK)
			}
		})
	}
}

func TestUnion2(t *testing.T) {
	configPath := "../../cmd/tools/doctor/testdata/testConfig.yml"
	n := node.NewS("foople")
	conf.TestLoadHarvestConfig(configPath)
	p, err := conf.PollerNamed("infinity2")
	if err != nil {
		panic(err)
	}
	Union2(n, p)
	labels := n.GetChildS("labels")
	if labels == nil {
		t.Errorf("got nil, want labels")
	}
	type label struct {
		key string
		val string
	}
	wants := []label{
		{key: "org", val: "abc"},
		{key: "site", val: "RTP"},
		{key: "floor", val: "3"},
	}
	for i, c := range labels.Children {
		want := wants[i]
		if want.key != c.GetNameS() {
			t.Errorf("got key=%s, want=%s", c.GetNameS(), want.key)
		}
		if want.val != c.GetContentS() {
			t.Errorf("got key=%s, want=%s", c.GetContentS(), want.val)
		}
	}
}

func TestPublishUrl(t *testing.T) {
	poller := Poller{}

	type test struct {
		name   string
		isTls  bool
		listen string
		want   string
	}

	tests := []test{
		{name: "localhost", isTls: false, listen: "localhost:8118", want: "http://localhost:8118/api/v1/sd"},
		{name: "all interfaces", isTls: false, listen: ":8118", want: "http://127.0.0.1:8118/api/v1/sd"},
		{name: "ip", isTls: false, listen: "10.0.1.1:8118", want: "http://10.0.1.1:8118/api/v1/sd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf.Config.Admin.Httpsd = conf.Httpsd{}
			if tt.isTls {
				conf.Config.Admin.Httpsd.TLS = conf.TLS{
					CertFile: "a",
					KeyFile:  "a",
				}
			}
			conf.Config.Admin.Httpsd.Listen = tt.listen
			got := poller.makePublishUrl()
			if got != tt.want {
				t.Errorf("makePublishUrl got = [%v] want [%v]", got, tt.want)
			}
		})
	}
}
