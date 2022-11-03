package main

import (
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"os"
	"reflect"
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
		t.Fatal("got nil, want labels")
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
		isTLS  bool
		listen string
		want   string
	}

	tests := []test{
		{name: "localhost", isTLS: false, listen: "localhost:8118", want: "http://localhost:8118/api/v1/sd"},
		{name: "all interfaces", isTLS: false, listen: ":8118", want: "http://127.0.0.1:8118/api/v1/sd"},
		{name: "ip", isTLS: false, listen: "10.0.1.1:8118", want: "http://10.0.1.1:8118/api/v1/sd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf.Config.Admin.Httpsd = conf.Httpsd{}
			if tt.isTLS {
				conf.Config.Admin.Httpsd.TLS = conf.TLS{
					CertFile: "a",
					KeyFile:  "a",
				}
			}
			conf.Config.Admin.Httpsd.Listen = tt.listen
			got := poller.makePublishURL()
			if got != tt.want {
				t.Errorf("makePublishURL got = [%v] want [%v]", got, tt.want)
			}
		})
	}
}

func TestCollectorUpgrade(t *testing.T) {
	poller := Poller{}

	type test struct {
		name           string
		clusterVersion string
		askFor         string
		wantCollector  string
		setEnvVar      bool
	}

	tests := []test{
		{name: "9.11 use ZAPI", clusterVersion: "9.11.1", askFor: "Zapi", wantCollector: "Zapi"},
		{name: "9.11 use REST", clusterVersion: "9.11.1", askFor: "Rest", wantCollector: "Rest"},
		{name: "9.12 upgrade", clusterVersion: "9.12.1", askFor: "Zapi", wantCollector: "Rest"},
		{name: "9.12 w/ envar", clusterVersion: "9.12.1", askFor: "Zapi", wantCollector: "Zapi", setEnvVar: true},
		{name: "9.12 REST", clusterVersion: "9.12.3", askFor: "Rest", wantCollector: "Rest", setEnvVar: true},
		{name: "9.13 RestPerf", clusterVersion: "9.13.1", askFor: "ZapiPerf", wantCollector: "RestPerf"},
		{name: "9.13 REST w/ envar", clusterVersion: "9.13.1", askFor: "Rest", wantCollector: "Rest", setEnvVar: true},
		{name: "9.13 REST", clusterVersion: "9.13.1", askFor: "Rest", wantCollector: "Rest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := conf.Collector{
				Name: tt.askFor,
			}
			if tt.setEnvVar {
				_ = os.Setenv(NoUpgrade, "1")
			} else {
				_ = os.Unsetenv(NoUpgrade)
			}
			newCollector := poller.negotiateAPI(collector, tt.clusterVersion)
			if newCollector.Name != tt.wantCollector {
				t.Errorf("got = [%s] want [%s]", newCollector.Name, tt.wantCollector)
			}
		})
	}
}

func Test_nonOverlappingCollectors(t *testing.T) {
	tests := []struct {
		name string
		args []objectCollector
		want []objectCollector
	}{
		{name: "empty", args: make([]objectCollector, 0), want: make([]objectCollector, 0)},
		{name: "one", args: ocs("Rest"), want: ocs("Rest")},
		{name: "no overlap", args: ocs("Rest", "ZapiPerf"), want: ocs("Rest", "ZapiPerf")},
		{name: "w overlap1", args: ocs("Rest", "Zapi"), want: ocs("Rest")},
		{name: "w overlap2", args: ocs("Zapi", "Rest"), want: ocs("Zapi")},
		{name: "w overlap3",
			args: ocs("Zapi", "Rest", "Rest", "Rest", "Rest", "Rest", "Zapi", "Zapi", "Zapi", "Zapi", "Zapi"),
			want: ocs("Zapi")},
		{name: "non ontap", args: ocs("Rest", "SG"), want: ocs("Rest", "SG")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nonOverlappingCollectors(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got=[%v], want=[%v]", got, tt.want)
			}
		})
	}
}

func ocs(names ...string) []objectCollector {
	collectors := make([]objectCollector, 0, len(names))
	for _, n := range names {
		collectors = append(collectors, objectCollector{class: n})
	}
	return collectors
}
