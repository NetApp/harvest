package ems

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/rs/zerolog/log"
	"os"
	"testing"
)

const EmsPollerName = "testEms"

func Test_Ems(t *testing.T) {
	// Initialize the Ems collector
	e := NewEms()
	// Non-bookend test
	NonBookendEmsTest(t, e)
	// Bookend issuing-resolving ems test
	BookendEmsTest(t, e)
}

func NonBookendEmsTest(t *testing.T, e *Ems) {
	e.updateMatrix()

	// Simulated nonBookend ems "wafl.vol.autoSize.done" and ems "arw.volume.state" with op value as "disable-in-progress"
	nonBookendEmsNames := []string{"wafl.vol.autoSize.done", "arw.volume.state"}
	results := collectors.JSONToGson("testdata/ems-poll-1.json", true)
	// Polling ems collector to handle results
	emsMetric, emsCount := e.HandleResults(results, e.emsProp)
	if emsCount == 0 {
		t.Fatal("Failed to fetch data")
	}

	// Check and evaluate ems events
	var notGeneratedEmsNames []string
	for generatedEmsName, mx := range emsMetric {
		if util.Contains(nonBookendEmsNames, generatedEmsName) {
			metr, ok := mx.GetMetrics()["events"]
			// e.Matrix map would have one entry of Ems(parent) in map, skipping that as it's not required for testing.
			if !ok && generatedEmsName != "Ems" {
				t.Fatalf("Failed to get netric for Ems %s", generatedEmsName)
			}
			for _, instance := range mx.GetInstances() {
				// If value not exist for that instance or metric value 0 indicate ems hasn't been raised.
				if val, ok := metr.GetValueFloat64(instance); !ok || val == 0 {
					notGeneratedEmsNames = append(notGeneratedEmsNames, generatedEmsName)
				}
				// Test for matches - filter
				if generatedEmsName == "arw.volume.state" {
					if instance.GetLabel("op") == "disable-in-progress" {
						// OK
					} else {
						t.Errorf("Labels op= %s, expected: disable-in-progress", instance.GetLabel("op"))
					}
				}
			}
		} else if generatedEmsName != "Ems" {
			t.Errorf("Extra non-bookend ems event found= %s", generatedEmsName)
		}
	}

	count := len(notGeneratedEmsNames)
	if count > 0 {
		t.Fatalf("These Non-Bookend Ems haven't been raised: %s", notGeneratedEmsNames)
	}

	// AutoResolve testcase: In next poll, updateMatrix would remove all non-bookend ems and retained only bookend ems.
	e.updateMatrix()

	results = collectors.JSONToGson("testdata/ems-poll-2.json", true)
	// Polling ems collector to handle results
	emsMetric, emsCount = e.HandleResults(results, e.emsProp)
	if emsCount == 0 {
		t.Fatal("Failed to fetch data")
	}

	// Check and evaluate non-bookend ems events got auto resolved successfully.
	for generatedEmsName := range emsMetric {
		if util.Contains(nonBookendEmsNames, generatedEmsName) {
			t.Errorf("Non-bookend ems event hasn't been auto resolved= %s", generatedEmsName)
		}
	}
}

func BookendEmsTest(t *testing.T, e *Ems) {
	e.updateMatrix()

	// Simulated bookend issuing ems "wafl.vvol.offline" and ems "hm.alert.raised" with alert_id value as "RaidLeftBehindAggrAlert"
	issuingEmsNames := []string{"wafl.vvol.offline", "hm.alert.raised"}
	results := collectors.JSONToGson("testdata/ems-poll-2.json", true)
	// Polling ems collector to handle results
	emsMetric, emsCount := e.HandleResults(results, e.emsProp)
	if emsCount == 0 {
		t.Fatal("Failed to fetch data")
	}

	// Check and evaluate ems events
	var notGeneratedEmsNames []string
	for generatedEmsName, mx := range emsMetric {
		if util.Contains(issuingEmsNames, generatedEmsName) {
			metr, ok := mx.GetMetrics()["events"]
			// e.Matrix map would have one entry of Ems(parent) in map, skipping that as it's not required for testing.
			if !ok && generatedEmsName != "Ems" {
				t.Fatalf("Failed to get netric for Ems %s", generatedEmsName)
			}
			for _, instance := range mx.GetInstances() {
				// If value not exist for that instance or metric value 0 indicate ems hasn't been raised.
				if val, ok := metr.GetValueFloat64(instance); !ok || val == 0 {
					notGeneratedEmsNames = append(notGeneratedEmsNames, generatedEmsName)
				}
				// Test for matches - filter
				if generatedEmsName == "hm.alert.raised" {
					if instance.GetLabel("alert_id") == "RaidLeftBehindAggrAlert" {
						// OK
					} else {
						t.Errorf("Labels alert_id= %s, expected: RaidLeftBehindAggrAlert", instance.GetLabel("alert_id"))
					}
				}
			}
		} else if generatedEmsName != "Ems" {
			t.Errorf("Extra bookend issuing ems event found= %s", generatedEmsName)
		}
	}
	count := len(notGeneratedEmsNames)
	if count > 0 {
		t.Fatalf("These Bookend Ems haven't been raised: %s", notGeneratedEmsNames)
	}

	// update matrix
	e.updateMatrix()

	// Simulated bookend resolving ems "wafl.vvol.online" and ems "hm.alert.cleared" with alert_id value as "RaidLeftBehindAggrAlert"
	results = collectors.JSONToGson("testdata/ems-poll-3.json", true)
	// Polling ems collector to handle results
	emsMetric, emsCount = e.HandleResults(results, e.emsProp)
	if emsCount == 0 {
		t.Fatal("Failed to fetch data")
	}

	// Check and evaluate ems events
	var notResolvedEmsNames []string
	for generatedEmsName, mx := range emsMetric {
		if util.Contains(issuingEmsNames, generatedEmsName) {
			metr, ok := mx.GetMetrics()["events"]
			// e.Matrix map would have one entry of Ems(parent) in map, skipping that as it's not required for testing.
			if !ok && generatedEmsName != "Ems" {
				t.Fatalf("Failed to get netric for Ems %s", generatedEmsName)
			}
			for _, instance := range mx.GetInstances() {
				// If value exist for that instance and metric value 1 indicate ems hasn't been resolved.
				val, ok := metr.GetValueFloat64(instance)
				if ok && val == 1 {
					notResolvedEmsNames = append(notResolvedEmsNames, generatedEmsName)
				}
				// Test for matches - filter
				if generatedEmsName == "hm.alert.raised" {
					if instance.GetLabel("alert_id") == "RaidLeftBehindAggrAlert" && ok && val == 0.0 {
						// OK
					} else {
						t.Errorf("Labels alert_id= %s, expected: RaidLeftBehindAggrAlert, metric value = %f, expected: 0.0", instance.GetLabel("alert_id"), val)
					}
				}
			}
		} else if generatedEmsName != "Ems" {
			t.Errorf("Extra bookend issuing ems event found= %s", generatedEmsName)
		}
	}
	// After resolving ems, all bookend ems should resolved
	count = len(notResolvedEmsNames)
	if count > 0 {
		t.Errorf("These Bookend Ems haven't been resolved: %s", notResolvedEmsNames)
	}
}

func NewEms() *Ems {
	// homepath is harvest directory level
	homePath := "../../../"
	emsConfgPath := homePath + "conf/ems/default.yaml"

	conf.TestLoadHarvestConfig("testdata/config.yml")
	opts := options.Options{
		Poller:   EmsPollerName,
		HomePath: homePath,
		IsTest:   true,
	}
	ac := collector.New("Ems", "Ems", &opts, emsParams(emsConfgPath), nil)
	e := &Ems{}
	err := e.Init(ac)
	if err != nil {
		log.Fatal().Err(err)
	}
	return e
}

func emsParams(emsConfigPath string) *node.Node {
	bytes, err := os.ReadFile(emsConfigPath)
	if err != nil {
		panic(err)
	}

	root, err := tree.LoadYaml(bytes)
	if err != nil {
		panic(err)
	}
	return root
}
