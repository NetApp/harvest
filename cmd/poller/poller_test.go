package main

import (
	"errors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strings"
	"testing"
)

func TestUnion2(t *testing.T) {
	configPath := "../../cmd/tools/doctor/testdata/testConfig.yml"
	n := node.NewS("foople")
	conf.TestLoadHarvestConfig(configPath)
	p, err := conf.PollerNamed("infinity2")
	if err != nil {
		panic(err)
	}
	err = Union2(n, p)
	if err != nil {
		panic(err)
	}
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
		got := c.GetContentS()
		if want.val != got {
			t.Errorf("got key=%s, want=%s", got, want.val)
		}
	}

	pp := n.GetChildContentS("prom_port")
	if pp != "2000" {
		t.Errorf("got prom_port=%s, want=2000", pp)
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
	poller := Poller{params: &conf.Poller{}}

	type test struct {
		name          string
		askFor        string
		wantCollector string
		remote        conf.Remote
	}

	ontap911 := conf.Remote{Version: "9.11.1", ZAPIsExist: true}
	ontap917 := conf.Remote{Version: "9.17.1", ZAPIsExist: false}
	asaR2 := conf.Remote{Version: "9.16.1", ZAPIsExist: false, IsDisaggregated: true, IsSanOptimized: true}
	keyPerf := conf.Remote{Version: "9.17.1", ZAPIsExist: false, IsDisaggregated: true}
	keyPerfWithZapi := conf.Remote{Version: "9.17.1", ZAPIsExist: true, IsDisaggregated: true}

	tests := []test{
		{name: "9.11 w/ ZAPI", remote: ontap911, askFor: "Zapi", wantCollector: "Zapi"},
		{name: "9.11 w/ ZAPI", remote: ontap911, askFor: "ZapiPerf", wantCollector: "ZapiPerf"},
		{name: "9.11 w/ ZAPI", remote: ontap911, askFor: "Rest", wantCollector: "Rest"},
		{name: "9.11 w/ ZAPI", remote: ontap911, askFor: "KeyPerf", wantCollector: "KeyPerf"},

		{name: "9.17 no ZAPI", remote: ontap917, askFor: "Zapi", wantCollector: "Rest"},
		{name: "9.17 no ZAPI", remote: ontap917, askFor: "ZapiPerf", wantCollector: "RestPerf"},
		{name: "9.17 no ZAPI", remote: ontap917, askFor: "KeyPerf", wantCollector: "KeyPerf"},

		{name: "KeyPerf", remote: keyPerf, askFor: "Zapi", wantCollector: "Rest"},
		{name: "KeyPerf", remote: keyPerf, askFor: "Rest", wantCollector: "Rest"},
		{name: "KeyPerf", remote: keyPerf, askFor: "ZapiPerf", wantCollector: "KeyPerf"},
		{name: "KeyPerf", remote: keyPerf, askFor: "RestPerf", wantCollector: "KeyPerf"},

		{name: "KeyPerf w/ ZAPI", remote: keyPerfWithZapi, askFor: "Zapi", wantCollector: "Zapi"},
		{name: "KeyPerf w/ ZAPI", remote: keyPerfWithZapi, askFor: "ZapiPerf", wantCollector: "KeyPerf"},
		{name: "KeyPerf w/ ZAPI", remote: keyPerfWithZapi, askFor: "RestPerf", wantCollector: "KeyPerf"},

		{name: "ASA R2", remote: asaR2, askFor: "Zapi", wantCollector: "Rest"},
		{name: "ASA R2", remote: asaR2, askFor: "RestPerf", wantCollector: "KeyPerf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := conf.Collector{
				Name: tt.askFor,
			}

			newCollector := poller.upgradeCollector(collector, tt.remote)
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
		{name: "no overlap", args: ocs("Rest", "KeyPerf"), want: ocs("Rest", "KeyPerf")},
		{name: "overlap", args: ocs("RestPerf", "KeyPerf"), want: ocs("RestPerf")},
		{name: "overlap", args: ocs("KeyPerf", "KeyPerf"), want: ocs("KeyPerf")},
		{name: "w overlap StatPerf1", args: ocs("StatPerf", "ZapiPerf"), want: ocs("StatPerf")},
		{name: "w overlap StatPerf2", args: ocs("RestPerf", "StatPerf"), want: ocs("RestPerf")},
		{name: "w overlap StatPerf3", args: ocs("KeyPerf", "StatPerf"), want: ocs("KeyPerf")},
		{name: "no overlap StatPerf", args: ocs("Rest", "StatPerf"), want: ocs("Rest", "StatPerf")},
		{name: "overlap statperf", args: ocs("StatPerf", "StatPerf"), want: ocs("StatPerf")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nonOverlappingCollectors(tt.args)
			diff := cmp.Diff(got, tt.want, cmp.AllowUnexported(objectCollector{}))
			if diff != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func ocs(names ...string) []objectCollector {
	objectCollectors := make([]objectCollector, 0, len(names))
	for _, n := range names {
		objectCollectors = append(objectCollectors, objectCollector{class: n})
	}
	return objectCollectors
}

func Test_uniquifyObjectCollectors(t *testing.T) {
	tests := []struct {
		name string
		args map[string][]objectCollector
		want []objectCollector
	}{
		{name: "empty", args: make(map[string][]objectCollector), want: []objectCollector{}},
		{name: "volume-rest", args: objectCollectorMap("Volume: Rest, Zapi"), want: []objectCollector{{class: "Rest", object: "Volume"}}},
		{name: "qtree-rest", args: objectCollectorMap("Qtree: Rest, Zapi"), want: []objectCollector{{class: "Rest", object: "Qtree"}}},
		{name: "qtree-zapi", args: objectCollectorMap("Qtree: Zapi, Rest"), want: []objectCollector{{class: "Zapi", object: "Qtree"}}},
		{name: "qtree-rest-quota", args: objectCollectorMap("Qtree: Rest, Zapi", "Quota: Rest"),
			want: []objectCollector{{class: "Rest", object: "Qtree"}, {class: "Rest", object: "Quota"}}},
		{name: "qtree-zapi-disable-quota", args: objectCollectorMap("Qtree: Zapi, Rest", "Quota: Rest"),
			want: []objectCollector{{class: "Zapi", object: "Qtree"}}},
		{name: "volume-restperf", args: objectCollectorMap("Volume: RestPerf, KeyPerf"),
			want: []objectCollector{{class: "RestPerf", object: "Volume"}}},
		{name: "volume-keyperf", args: objectCollectorMap("Volume: KeyPerf, RestPerf"),
			want: []objectCollector{{class: "KeyPerf", object: "Volume"}}},
		{name: "multi-keyperf", args: objectCollectorMap("Volume: RestPerf", "Aggregate: KeyPerf"),
			want: []objectCollector{{class: "RestPerf", object: "Volume"}, {class: "KeyPerf", object: "Aggregate"}}},
		{name: "volume-statperf", args: objectCollectorMap("Volume: StatPerf, ZapiPerf"),
			want: []objectCollector{{class: "StatPerf", object: "Volume"}}},
		{name: "multi-statperf", args: objectCollectorMap("Volume: StatPerf", "Aggregate: RestPerf"),
			want: []objectCollector{{class: "StatPerf", object: "Volume"}, {class: "RestPerf", object: "Aggregate"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uniquifyObjectCollectors(tt.args)

			diff := cmp.Diff(got, tt.want, cmp.AllowUnexported(objectCollector{}), cmpopts.SortSlices(func(a, b objectCollector) bool {
				return a.class+a.object < b.class+b.object
			}))

			if diff != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func objectCollectorMap(constructors ...string) map[string][]objectCollector {
	objectsToCollectors := make(map[string][]objectCollector)

	for _, template := range constructors {
		before, after, _ := strings.Cut(template, ":")
		object := before
		classes := strings.Split(after, ",")
		for _, class := range classes {
			class := strings.TrimSpace(class)
			objectsToCollectors[object] = append(objectsToCollectors[object], objectCollector{class: class, object: object})
		}
	}

	return objectsToCollectors
}

func TestMergeRemotes(t *testing.T) {

	type test struct {
		name       string
		remoteZapi conf.Remote
		remoteRest conf.Remote
		errZapi    error
		errRest    error
		want       conf.Remote
		wantErr    bool
	}

	tests := []test{
		{
			name:       "No ONTAP",
			remoteZapi: conf.Remote{},
			remoteRest: conf.Remote{},
			errZapi:    errors.New("no ZAPI"),
			errRest:    errors.New("no REST"),
			want:       conf.Remote{},
			wantErr:    true,
		},
		{
			name:       "No REST",
			remoteZapi: conf.Remote{UUID: "abc", Version: "9.11.1", ZAPIsExist: true},
			remoteRest: conf.Remote{},
			errZapi:    nil,
			errRest:    errors.New("no REST"),
			want:       conf.Remote{UUID: "abc", Version: "9.11.1", ZAPIsExist: true},
			wantErr:    true,
		},
		{
			name:       "No ZAPIs",
			remoteZapi: conf.Remote{},
			remoteRest: conf.Remote{UUID: "abc", Version: "9.17.1", HasREST: true},
			errZapi:    errors.New("no ZAPI"),
			errRest:    nil,
			want:       conf.Remote{UUID: "abc", Version: "9.17.1", HasREST: true},
			wantErr:    true,
		},
		{
			name:       "Both",
			remoteZapi: conf.Remote{UUID: "abc", Version: "9.17.1", ZAPIsExist: true, HasREST: true},
			remoteRest: conf.Remote{UUID: "abc", Version: "9.17.1", ZAPIsExist: true, HasREST: true},
			errZapi:    nil,
			errRest:    nil,
			want:       conf.Remote{UUID: "abc", Version: "9.17.1", ZAPIsExist: true, HasREST: true},
			wantErr:    false,
		},
		{
			name:       "Both Fail",
			remoteZapi: conf.Remote{ZAPIsExist: true},
			remoteRest: conf.Remote{},
			errZapi:    errors.New("auth error"),
			errRest:    errors.New("auth error"),
			want:       conf.Remote{ZAPIsExist: true},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRemote, err := collectors.MergeRemotes(tt.remoteZapi, tt.remoteRest, tt.errZapi, tt.errRest)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeRemotes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(gotRemote, tt.want); diff != "" {
				t.Errorf("MergeRemotes() mismatch (-gotRemote +want):\n%s", diff)
			}
		})
	}
}
