package ems

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"os"
	"slices"
	"testing"
	"time"
)

// Bookend EMS testing: Simulated bookend issuing ems "wafl.vvol.offline" and ems "hm.alert.raised" with alert_id value as "RaidLeftBehindAggrAlert"
var issuingEmsNames = []string{"wafl.vvol.offline", "hm.alert.raised"}

// Default labels per ems is 5, "hm.alert.raised" ems has 11 labels and "wafl.vvol.offline" has 4 labels, total instance labels would be 24
const expectedInstanceLabelCount = 25

// Auto resolve EMS testing: Simulated bookend issuing ems "LUN.offline" and ems "monitor.fan.critical"
var autoresolveEmsNames = []string{"LUN.offline", "monitor.fan.critical"}

func Test_Ems(t *testing.T) {
	// Initialize the Ems collector
	e := NewEms()
	// Bookend issuing-resolving ems test
	BookendEmsTest(t, e)
}

// Bookend ems test-case would also handle the workflow of non-bookend ems test-case.
func BookendEmsTest(t *testing.T, e *Ems) {
	// Step 1: Generate Bookend issuing ems
	e.testBookendIssuingEms(t, "testdata/issuingEms.json")

	// Step 2: Generate Bookend resolving ems
	e.testBookendResolvingEms(t, "testdata/resolvingEms.json")

	// Step 3: Generate Bookend issuing ems and validate auto resolution functionality
	e.testAutoResolvingEms(t, "testdata/autoresolveEms.json")
}

func NewEms() *Ems {
	// homePath is harvest directory level
	homePath := "../../../"
	emsConfigPath := homePath + "conf/ems/default.yaml"

	conf.TestLoadHarvestConfig("testdata/config.yml")
	opts := options.New(options.WithConfPath(homePath + "conf"))
	opts.Poller = "testEms"
	opts.HomePath = homePath
	opts.IsTest = true

	ac := collector.New("Ems", "Ems", opts, emsParams(emsConfigPath), nil, conf.Remote{})
	e := &Ems{}
	if err := e.Init(ac); err != nil {
		slog.Error("", slogx.Err(err))
		os.Exit(1)
	}
	// Changed the resolve_after for 2 issuing ems for auto resolve testing
	e.resolveAfter["LUN.offline"] = 1 * time.Second
	e.resolveAfter["monitor.fan.critical"] = 2 * time.Second
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

func (e *Ems) testBookendIssuingEms(t *testing.T, path string) {
	e.updateMatrix(time.Now())

	results := collectors.JSONToGson(path, true)
	// Polling ems collector to handle results
	if _, emsCount, _ := e.HandleResults(results, e.emsProp); emsCount != expectedInstanceLabelCount {
		t.Fatalf("Instance labels count mismatch detected. Expected labels: %d actual labels: %d", expectedInstanceLabelCount, emsCount)
	}

	// Check and evaluate ems events
	var notGeneratedEmsNames []string
	for generatedEmsName, mx := range e.Matrix {
		if slices.Contains(issuingEmsNames, generatedEmsName) {
			metr, ok := mx.GetMetrics()["events"]
			if !ok {
				t.Fatalf("Failed to get netric for Ems %s", generatedEmsName)
			}
			for _, instance := range mx.GetInstances() {
				// If value not exist for that instance or metric value 0 indicate ems hasn't been raised.
				if val, ok := metr.GetValueFloat64(instance); !ok || val == 0 {
					notGeneratedEmsNames = append(notGeneratedEmsNames, generatedEmsName)
				}
				// Test for matches - filter
				if generatedEmsName == "hm.alert.raised" {
					if instance.GetLabel("alert_id") != "RaidLeftBehindAggrAlert" {
						t.Errorf("Labels alert_id= %s, expected: RaidLeftBehindAggrAlert", instance.GetLabel("alert_id"))
					}
				}
			}
		}
	}
	if len(notGeneratedEmsNames) > 0 {
		t.Fatalf("These Bookend Ems haven't been raised: %s", notGeneratedEmsNames)
	}
}

func (e *Ems) testBookendResolvingEms(t *testing.T, path string) {
	e.updateMatrix(time.Now())

	// Simulated bookend resolving ems "wafl.vvol.online" and ems "hm.alert.cleared" with alert_id value as "RaidLeftBehindAggrAlert"
	results := collectors.JSONToGson(path, true)
	// Polling ems collector to handle results
	if _, emsCount, _ := e.HandleResults(results, e.emsProp); emsCount == 0 {
		t.Fatal("Failed to fetch data")
	}

	// Check and evaluate ems events
	var notResolvedEmsNames []string
	for generatedEmsName, mx := range e.Matrix {
		if slices.Contains(issuingEmsNames, generatedEmsName) {
			metr, ok := mx.GetMetrics()["events"]
			if !ok {
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
					if instance.GetLabel("alert_id") != "RaidLeftBehindAggrAlert" || !ok || val != 0.0 {
						t.Errorf("Labels alert_id= %s, expected: RaidLeftBehindAggrAlert, metric value = %f, expected: 0.0", instance.GetLabel("alert_id"), val)
					}
				}
			}
		}
	}
	// After resolving ems, all bookend ems should resolved
	if len(notResolvedEmsNames) > 0 {
		t.Errorf("These Bookend Ems haven't been resolved: %s", notResolvedEmsNames)
	}
}

func (e *Ems) testAutoResolvingEms(t *testing.T, path string) {
	var notGeneratedEmsNames, notAutoResolvedEmsNames []string
	e.updateMatrix(time.Now())

	results := collectors.JSONToGson(path, true)
	// Polling ems collector to handle results
	if _, emsCount, _ := e.HandleResults(results, e.emsProp); emsCount == 0 {
		t.Fatal("Failed to fetch data")
	}

	// Check and evaluate ems events
	for generatedEmsName, mx := range e.Matrix {
		if slices.Contains(autoresolveEmsNames, generatedEmsName) {
			if metr, ok := mx.GetMetrics()["events"]; ok {
				for _, instance := range mx.GetInstances() {
					// If value not exist for that instance or metric value 0 indicate ems hasn't been raised.
					if val, ok := metr.GetValueFloat64(instance); !ok || val == 0 {
						notGeneratedEmsNames = append(notGeneratedEmsNames, generatedEmsName)
					}
				}
			} else {
				t.Fatalf("Failed to get netric for Ems %s", generatedEmsName)
			}
		}
	}
	if len(notGeneratedEmsNames) > 0 {
		t.Fatalf("These Bookend Ems haven't been raised: %s", notGeneratedEmsNames)
	}

	// Evaluate the cache for existence of these auto resolve ems.
	// Simulate one second in the future and check that the LUN.offline ems event is auto resolved
	e.updateMatrix(time.Now().Add(1 * time.Second))
	// Check and evaluate bookend ems events got auto resolved successfully.
	for generatedEmsName, mx := range e.Matrix {
		if slices.Contains(autoresolveEmsNames, generatedEmsName) {
			if metr, ok := mx.GetMetrics()["events"]; ok {
				for _, instance := range mx.GetInstances() {
					// If value not exist for that instance or metric value 0 indicate ems hasn't been raised.
					if val, ok := metr.GetValueFloat64(instance); !ok || val == 1 {
						notAutoResolvedEmsNames = append(notAutoResolvedEmsNames, generatedEmsName)
					}
				}
			} else {
				t.Fatalf("Failed to get netric for Ems %s", generatedEmsName)
			}
		}
	}
	if slices.Contains(notAutoResolvedEmsNames, "LUN.offline") {
		t.Fatalf("These Bookend Ems haven't been auto resolved: %s", notAutoResolvedEmsNames)
	}

	// Simulate two seconds in future and check that the "LUN.offline" ems event was removed from the cache and
	// "monitor.fan.critical" is auto resolved
	e.updateMatrix(time.Now().Add(2 * time.Second))
	notAutoResolvedEmsNames = make([]string, 0)
	// Check bookend ems event got removed from cache successfully.
	if e.Matrix["LUN.offline"] != nil {
		if len(e.Matrix["LUN.offline"].GetInstances()) != 0 {
			t.Fatalf("Instances haven't been removed from bookend cache: LUN.offline")
		}
	}
	for generatedEmsName, mx := range e.Matrix {
		if slices.Contains(autoresolveEmsNames, generatedEmsName) {
			if metr, ok := mx.GetMetrics()["events"]; ok {
				for _, instance := range mx.GetInstances() {
					// If value not exist for that instance or metric value 0 indicate ems hasn't been raised.
					if val, ok := metr.GetValueFloat64(instance); !ok || val == 1 {
						notAutoResolvedEmsNames = append(notAutoResolvedEmsNames, generatedEmsName)
					}
				}
			} else {
				t.Fatalf("Failed to get netric for Ems %s", generatedEmsName)
			}
		}
	}
	if slices.Contains(notAutoResolvedEmsNames, "monitor.fan.critical") {
		t.Fatalf("These Bookend Ems haven't been auto resolved: %s", notAutoResolvedEmsNames)
	}
}
