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
	"time"
)

func Test_Ems(t *testing.T) {
	// Initialize the Ems collector
	e := NewEms()
	// Bookend issuing-resolving ems test
	BookendEmsTest(t, e)
}

// Bookend ems test-case would also handle the workflow of non-bookend ems test-case.
func BookendEmsTest(t *testing.T, e *Ems) {
	// Step 1: Generate Bookend issuing ems
	e.updateMatrix()

	// Simulated bookend issuing ems "wafl.vvol.offline" and ems "hm.alert.raised" with alert_id value as "RaidLeftBehindAggrAlert"
	issuingEmsNames := []string{"wafl.vvol.offline", "hm.alert.raised"}
	results := collectors.JSONToGson("testdata/ems-poll-1.json", true)
	// Polling ems collector to handle results
	if _, emsCount := e.HandleResults(results, e.emsProp); emsCount == 0 {
		t.Fatal("Failed to fetch data")
	}

	// Check and evaluate ems events
	var notGeneratedEmsNames []string
	for generatedEmsName, mx := range e.Matrix {
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
	if len(notGeneratedEmsNames) > 0 {
		t.Fatalf("These Bookend Ems haven't been raised: %s", notGeneratedEmsNames)
	}

	// Step 2: Generate Bookend resolving ems
	e.updateMatrix()

	// Simulated bookend resolving ems "wafl.vvol.online" and ems "hm.alert.cleared" with alert_id value as "RaidLeftBehindAggrAlert"
	results = collectors.JSONToGson("testdata/ems-poll-2.json", true)
	// Polling ems collector to handle results
	if _, emsCount := e.HandleResults(results, e.emsProp); emsCount == 0 {
		t.Fatal("Failed to fetch data")
	}

	// Check and evaluate ems events
	var notResolvedEmsNames []string
	for generatedEmsName, mx := range e.Matrix {
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
	if len(notResolvedEmsNames) > 0 {
		t.Errorf("These Bookend Ems haven't been resolved: %s", notResolvedEmsNames)
	}

	// Step 3: Auto-resolved bookend ems after time expired
	e.updateMatrix()

	// Simulated bookend issuing ems "LUN.offline" and ems "monitor.fan.critical"
	issuingEmsNames = []string{"LUN.offline", "monitor.fan.critical"}
	results = collectors.JSONToGson("testdata/ems-poll-3.json", true)
	// Polling ems collector to handle results
	if _, emsCount := e.HandleResults(results, e.emsProp); emsCount == 0 {
		t.Fatal("Failed to fetch data")
	}

	// Check and evaluate bookend ems events got generated successfully, e.Matrix map would have one entry of Ems(parent) in map.
	if len(e.Matrix) != len(issuingEmsNames)+1 {
		t.Error("Not all bookend ems event have been generated")
	}
	for generatedEmsName := range e.Matrix {
		if !util.Contains(issuingEmsNames, generatedEmsName) && generatedEmsName != "Ems" {
			t.Errorf("Extra ems event has been detected= %s", generatedEmsName)
		}
	}
	// Changed the resolve_after for 2 issuing ems
	e.resolveAfter["LUN.offline"] = 1 * time.Minute
	e.resolveAfter["monitor.fan.critical"] = 2 * time.Minute

	// Sleep for 1 minute and check LUN.offline ems got auto resolved
	time.Sleep(1 * time.Minute)
	e.updateMatrix()
	// Check and evaluate bookend ems events got generated successfully, e.Matrix map would have one entry of Ems(parent) in map.
	if len(e.Matrix) != 2 {
		t.Error("Bookend ems event haven't been auto resolved")
	}
	for generatedEmsName := range e.Matrix {
		if generatedEmsName != "monitor.fan.critical" && generatedEmsName != "Ems" {
			t.Errorf("This bookend ems event haven't been auto resolved= %s", generatedEmsName)
		}
	}

	// Sleep for another 1 minute and check both the ems got auto resolved
	time.Sleep(1 * time.Minute)
	e.updateMatrix()
	// Check and evaluate bookend ems events got generated successfully, e.Matrix map would have one entry of Ems(parent) in map.
	if len(e.Matrix) != 1 {
		t.Error("Bookend ems event haven't been auto resolved")
	}
	for generatedEmsName := range e.Matrix {
		if generatedEmsName != "Ems" {
			t.Errorf("This bookend ems event haven't been auto resolved= %s", generatedEmsName)
		}
	}
}

func NewEms() *Ems {
	// homepath is harvest directory level
	homePath := "../../../"
	emsConfgPath := homePath + "conf/ems/default.yaml"
	emsPoller := "testEms"

	conf.TestLoadHarvestConfig("testdata/config.yml")
	opts := options.Options{
		Poller:   emsPoller,
		HomePath: homePath,
		IsTest:   true,
	}
	ac := collector.New("Ems", "Ems", &opts, emsParams(emsConfgPath), nil)
	e := &Ems{}
	if err := e.Init(ac); err != nil {
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
