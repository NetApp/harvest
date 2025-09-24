package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func TestUnion2(t *testing.T) {
	configPath := "../../cmd/tools/doctor/testdata/testConfig.yml"
	n := node.NewS("foople")
	conf.TestLoadHarvestConfig(configPath)
	p, err := conf.PollerNamed("infinity2")
	assert.Nil(t, err)
	err = Union2(n, p)
	assert.Nil(t, err)

	labels := n.GetChildS("labels")
	assert.NotNil(t, labels)

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
		assert.Equal(t, c.GetNameS(), want.key)

		got := c.GetContentS()
		assert.Equal(t, got, want.val)
		if want.val != got {
			t.Errorf("got key=%s, want=%s", got, want.val)
		}
	}

	pp := n.GetChildContentS("prom_port")
	assert.Equal(t, pp, "2000")
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
			assert.Equal(t, got, tt.want)
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
			c := conf.Collector{
				Name: tt.askFor,
			}

			newCollector := poller.upgradeCollector(c, tt.remote)
			assert.Equal(t, newCollector.Name, tt.wantCollector)
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
		classes := strings.SplitSeq(after, ",")
		for class := range classes {
			class := strings.TrimSpace(class)
			objectsToCollectors[object] = append(objectsToCollectors[object], objectCollector{class: class, object: object})
		}
	}

	return objectsToCollectors
}

func TestUpgradeObjectCollector(t *testing.T) {
	poller := &Poller{
		options: &options.Options{
			ConfPaths: []string{"testdata", "../../conf"},
		},
		params: &conf.Poller{
			Collectors: []conf.Collector{
				{Name: "KeyPerf", Templates: &[]string{"default.yaml", "custom.yaml"}},
				{Name: "RestPerf", Templates: &[]string{"default.yaml", "custom.yaml"}},
			},
		},
		remote: conf.Remote{
			Version: "9.15.0", // Use a version that supports KeyPerf for this test
		},
	}

	tests := []struct {
		name           string
		inputClass     string
		inputObject    string
		templateYAML   string
		expectedClass  string
		expectedObject string
		expectUpgrade  bool
	}{
		{
			name:        "no_dsl_delegation",
			inputClass:  "RestPerf",
			inputObject: "Volume",
			templateYAML: `
collector: RestPerf
objects:
  Volume: volume.yaml
`,
			expectedClass:  "RestPerf",
			expectedObject: "Volume",
			expectUpgrade:  false,
		},
		{
			name:        "dsl_delegation_keyperf",
			inputClass:  "RestPerf",
			inputObject: "Volume",
			templateYAML: `
collector: RestPerf
objects:
  Volume: KeyPerf:volume.yaml
  Aggregate: aggr.yaml
`,
			expectedClass:  "KeyPerf",
			expectedObject: "Volume",
			expectUpgrade:  true,
		},
		{
			name:        "invalid_dsl_format",
			inputClass:  "RestPerf",
			inputObject: "Volume",
			templateYAML: `
collector: RestPerf
objects:
  Volume: InvalidFormat
`,
			expectedClass:  "RestPerf",
			expectedObject: "Volume",
			expectUpgrade:  false,
		},
		{
			name:        "empty_object_value",
			inputClass:  "RestPerf",
			inputObject: "Volume",
			templateYAML: `
collector: RestPerf
objects:
  Volume: ""
`,
			expectedClass:  "RestPerf",
			expectedObject: "Volume",
			expectUpgrade:  false,
		},
		{
			name:        "zapiperrf_to_keyperf_strips_extended_templates",
			inputClass:  "ZapiPerf",
			inputObject: "Volume",
			templateYAML: `
collector: ZapiPerf
objects:
  Volume: KeyPerf:volume.yaml,exclude_transient_volumes.yaml
`,
			expectedClass:  "KeyPerf",
			expectedObject: "Volume",
			expectUpgrade:  true,
		},
		{
			name:        "zapiperrf_to_keyperf_strips_multiple_extended_templates",
			inputClass:  "ZapiPerf",
			inputObject: "Volume",
			templateYAML: `
collector: ZapiPerf
objects:
  Volume: KeyPerf:volume.yaml,custom1.yaml,custom2.yaml
`,
			expectedClass:  "KeyPerf",
			expectedObject: "Volume",
			expectUpgrade:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := node.NewS("")
			template.NewChildS("collector", tt.inputClass)
			objectsNode := template.NewChildS("objects", "")

			lines := strings.Split(tt.templateYAML, "\n")
			var objectValue string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if after, ok := strings.CutPrefix(line, tt.inputObject+":"); ok {
					objectValue = strings.TrimSpace(after)
					break
				}
			}

			objectsNode.NewChildS(tt.inputObject, objectValue)

			oc := objectCollector{
				class:    tt.inputClass,
				object:   tt.inputObject,
				template: template,
			}

			result := poller.upgradeObjectCollector(oc)

			assert.Equal(t, result.class, tt.expectedClass)
			assert.Equal(t, result.object, tt.expectedObject)

			if tt.expectUpgrade {
				if templateObjects := result.template.GetChildS("objects"); templateObjects != nil {
					if objDef := templateObjects.GetChildS(result.object); objDef != nil {
						objectValue := objDef.GetContentS()
						assert.False(t, strings.Contains(objectValue, ":"))

						// Special check for ZapiPerf to KeyPerf upgrade: ensure extended templates are stripped
						if tt.name == "zapiperrf_to_keyperf_strips_extended_templates" {
							// Should only have the first template (volume.yaml), any extended templates stripped
							assert.Equal(t, objectValue, "volume.yaml")
							assert.False(t, strings.Contains(objectValue, ","))
						}
						// Special check for ZapiPerf to KeyPerf with multiple extended templates
						if tt.name == "zapiperrf_to_keyperf_strips_multiple_extended_templates" {
							// Should only have the first template (volume.yaml), all extended templates stripped
							assert.Equal(t, objectValue, "volume.yaml")
							assert.False(t, strings.Contains(objectValue, ","))
							assert.False(t, strings.Contains(objectValue, "custom1.yaml"))
							assert.False(t, strings.Contains(objectValue, "custom2.yaml"))
						}
					}
				}
			}
		})
	}
}

func TestUpgradeObjectCollectorVersionAware(t *testing.T) {
	tests := []struct {
		name           string
		ontapVersion   string
		inputClass     string
		inputObject    string
		templateYAML   string
		expectedClass  string
		expectedObject string
		expectUpgrade  bool
		description    string
	}{
		{
			name:         "volume_keyperf_upgrade_ontap_9_10",
			ontapVersion: "9.10.0",
			inputClass:   "ZapiPerf",
			inputObject:  "Volume",
			templateYAML: `
collector: ZapiPerf
objects:
  Volume: KeyPerf:volume.yaml
`,
			expectedClass:  "KeyPerf",
			expectedObject: "Volume",
			expectUpgrade:  true,
			description:    "ONTAP 9.10+ should allow volume KeyPerf upgrades",
		},
		{
			name:         "non_volume_keyperf_upgrade_ontap_9_6_allowed",
			ontapVersion: "9.6.0",
			inputClass:   "RestPerf",
			inputObject:  "Aggregate",
			templateYAML: `
collector: RestPerf
objects:
  Aggregate: KeyPerf:aggr.yaml
`,
			expectedClass:  "KeyPerf",
			expectedObject: "Aggregate",
			expectUpgrade:  true,
			description:    "Non-volume objects should upgrade to KeyPerf regardless of version",
		},
		{
			name:         "volume_keyperf_upgrade_ontap_9_6_blocked",
			ontapVersion: "9.6.0",
			inputClass:   "ZapiPerf",
			inputObject:  "Volume",
			templateYAML: `
collector: ZapiPerf
objects:
  Volume: KeyPerf:volume.yaml
`,
			expectedClass:  "ZapiPerf",
			expectedObject: "Volume",
			expectUpgrade:  false,
			description:    "Volume objects should NOT upgrade to KeyPerf on ONTAP 9.6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poller := &Poller{
				options: &options.Options{
					ConfPaths: []string{"../../conf"},
				},
				params: &conf.Poller{
					Collectors: []conf.Collector{
						{Name: "KeyPerf", Templates: &[]string{"default.yaml"}},
						{Name: "RestPerf", Templates: &[]string{"default.yaml"}},
						{Name: "ZapiPerf", Templates: &[]string{"default.yaml"}},
					},
				},
				remote: conf.Remote{
					Version: tt.ontapVersion,
				},
			}

			template := node.NewS("")
			template.NewChildS("collector", tt.inputClass)
			objectsNode := template.NewChildS("objects", "")

			lines := strings.Split(tt.templateYAML, "\n")
			var objectValue string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if after, ok := strings.CutPrefix(line, tt.inputObject+":"); ok {
					objectValue = strings.TrimSpace(after)
					break
				}
			}

			objectsNode.NewChildS(tt.inputObject, objectValue)

			oc := objectCollector{
				class:    tt.inputClass,
				object:   tt.inputObject,
				template: template,
			}

			result := poller.upgradeObjectCollector(oc)

			assert.Equal(t, result.class, tt.expectedClass)
			assert.Equal(t, result.object, tt.expectedObject)

			if tt.expectUpgrade {
				// Verify upgrade happened by checking class changed
				assert.NotEqual(t, result.class, tt.inputClass)
			} else {
				// Verify no upgrade happened
				assert.Equal(t, result.class, tt.inputClass)

				// For blocked volume upgrades, verify the KeyPerf: prefix was removed from template
				if strings.Contains(strings.ToLower(tt.inputObject), "volume") &&
					strings.Contains(tt.templateYAML, "KeyPerf:") {
					if templateObjects := result.template.GetChildS("objects"); templateObjects != nil {
						if objDef := templateObjects.GetChildS(result.object); objDef != nil {
							objectValue := objDef.GetContentS()
							// Should not contain KeyPerf: prefix anymore
							assert.False(t, strings.Contains(objectValue, "KeyPerf:"))
							// Should just be the template name
							assert.Equal(t, "volume.yaml", objectValue)
						}
					}
				}
			}
		})
	}
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
			wantErr:    false,
		},
		{
			name:       "No ZAPIs",
			remoteZapi: conf.Remote{},
			remoteRest: conf.Remote{UUID: "abc", Version: "9.17.1", HasREST: true},
			errZapi:    errors.New("no ZAPI"),
			errRest:    nil,
			want:       conf.Remote{UUID: "abc", Version: "9.17.1", HasREST: true},
			wantErr:    false,
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

func TestGetTemplatesForCollector(t *testing.T) {
	tests := []struct {
		name              string
		collectors        []conf.Collector
		collectorClass    string
		expectedTemplates []string
	}{
		{
			name: "collector_with_custom_templates",
			collectors: []conf.Collector{
				{Name: "KeyPerf", Templates: &[]string{"custom1.yaml", "custom2.yaml"}},
				{Name: "RestPerf", Templates: &[]string{"default.yaml", "custom.yaml"}},
			},
			collectorClass:    "KeyPerf",
			expectedTemplates: []string{"custom1.yaml", "custom2.yaml"},
		},
		{
			name: "collector_with_default_templates",
			collectors: []conf.Collector{
				{Name: "RestPerf", Templates: conf.DefaultTemplates},
			},
			collectorClass:    "RestPerf",
			expectedTemplates: []string{"default.yaml", "custom.yaml"},
		},
		{
			name: "collector_not_configured",
			collectors: []conf.Collector{
				{Name: "RestPerf", Templates: &[]string{"default.yaml", "custom.yaml"}},
			},
			collectorClass:    "KeyPerf",
			expectedTemplates: []string{"default.yaml", "custom.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poller := &Poller{
				params: &conf.Poller{
					Collectors: tt.collectors,
				},
			}

			result := poller.fetchCollectorTemplates(tt.collectorClass)
			assert.Equal(t, result, tt.expectedTemplates)
		})
	}
}
