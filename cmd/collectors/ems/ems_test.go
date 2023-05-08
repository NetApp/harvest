package ems

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"testing"
	"time"
)

// This poller should be present in harvest.yml file for evaluating the ems test-case.
const TestPoller = "dc1"
const Admin = "admin"

var nonBookendEmsNames []string
var issuingEmsNames []string
var resolvingEmsNames []string

func Setup() (*Ems, *conf.Poller) {
	conf.TestLoadHarvestConfig("../../../harvest.yml")
	poller, err := conf.PollerNamed(TestPoller)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err)
	}
	splits := strings.Split(pwd, "/")
	homePath := strings.Join(splits[:len(splits)-3], "/")

	// Fetch ems configured in template
	emsConfigDir := homePath + "/conf/ems/9.6.0"
	getEmsNames(emsConfigDir, "ems.yaml")

	// Initialize the Ems collector
	e := newEms(homePath, poller)
	return e, poller
}

func NonBookendEmsTest(t *testing.T, e *Ems, poller *conf.Poller) {
	// remove all ems matrix except parent object
	mat := e.Matrix[e.Object]
	e.Matrix = make(map[string]*matrix.Matrix)
	e.Matrix[e.Object] = mat
	e.PollInstance() //nolint:errcheck

	// Generate NonBookendEms: Check below non-bookend ems are supported for the given cluster
	now := time.Now()
	nonBookendEmsNames = []string{"wafl.vol.autoSize.done", "arw.volume.state"}
	supportedEms := generateEvents(nonBookendEmsNames, poller)
	log.Info().
		Strs("supportedEms", supportedEms.Slice()).
		Int("count", supportedEms.Size()).
		Str("dur", time.Since(now).Round(time.Millisecond).String()).
		Msg("Generated non-bookend ems")

	// Polling ems collector
	if _, err := e.PollData(); err != nil {
		t.Fatalf("Failed to fetch data %v", err)
	}

	// Check and evaluate ems events
	var notGeneratedEmsNames []string
	for generatedEmsName, mx := range e.Matrix {
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
		}
	}

	count := len(notGeneratedEmsNames)
	if count > 0 {
		t.Fatalf("These Non-Bookend Ems haven't been raised: %s", notGeneratedEmsNames)
	}
}

func BookendEmsTest(t *testing.T, e *Ems, poller *conf.Poller) {
	// remove all ems matrix except parent object
	mat := e.Matrix[e.Object]
	e.Matrix = make(map[string]*matrix.Matrix)
	e.Matrix[e.Object] = mat
	e.PollInstance() //nolint:errcheck

	// Generate BookendEms: Check below bookend issuing ems are supported for the given cluster
	now := time.Now()
	issuingEmsNames = []string{"wafl.vvol.offline", "hm.alert.raised"}
	supportedIssuingEms := generateEvents(issuingEmsNames, poller)
	log.Info().
		Strs("supportedIssuingEms", supportedIssuingEms.Slice()).
		Int("count", supportedIssuingEms.Size()).
		Str("dur", time.Since(now).Round(time.Millisecond).String()).
		Msg("Generated bookend issuing ems")

	// Polling ems collector
	if _, err := e.PollData(); err != nil {
		t.Fatalf("Failed to fetch data %v", err)
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
		}
	}
	count := len(notGeneratedEmsNames)
	if count > 0 {
		t.Fatalf("These Bookend Ems haven't been raised: %s", notGeneratedEmsNames)
	}

	// Resolve BookendEms: Check below bookend resolving ems are supported for the given cluster
	now = time.Now()
	resolvingEmsNames = []string{"wafl.vvol.online", "hm.alert.cleared"}
	supportedResolvingEms := generateEvents(resolvingEmsNames, poller)
	log.Info().
		Strs("supportedResolvingEms", supportedResolvingEms.Slice()).
		Int("count", supportedResolvingEms.Size()).
		Str("dur", time.Since(now).Round(time.Millisecond).String()).
		Msg("Generated resolving ems")

	// Polling ems collector
	if _, err := e.PollData(); err != nil {
		t.Fatalf("Failed to fetch data %v", err)
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
		}
	}
	// After resolving ems, all bookend ems should resolved
	count = len(notResolvedEmsNames)
	if count > 0 {
		t.Errorf("These Bookend Ems haven't been resolved: %s", notResolvedEmsNames)
	}
}

func Test_Ems(t *testing.T) {
	e, poller := Setup()
	NonBookendEmsTest(t, e, poller)
	BookendEmsTest(t, e, poller)
}

func newEms(homePath string, poller *conf.Poller) *Ems {
	opts := options.Options{
		Poller:   TestPoller,
		HomePath: homePath,
	}
	auth := auth.NewCredentials(poller, logging.Get())
	ac := collector.New("Ems", "Ems", &opts, emsParams(), auth)
	e := &Ems{}
	err := e.Init(ac)
	if err != nil {
		log.Fatal().Err(err)
	}
	return e
}

func emsParams() *node.Node {
	yml := `
collector: Ems

client_timeout: 1m
schedule:
  - instance: 24h
  - data:     3m

objects:
  Ems: ems.yaml
`
	root, err := tree.LoadYaml([]byte(yml))
	if err != nil {
		panic(err)
	}
	return root
}

func getEmsNames(dir string, fileName string) {
	nonBookendEmsNames = make([]string, 0)
	issuingEmsNames = make([]string, 0)
	resolvingEmsNames = make([]string, 0)
	nonBookendEmsCount := 0
	bookendEmsCount := 0

	emsConfigFilePath := dir + "/" + fileName
	log.Debug().Str("emsConfigFilePath", emsConfigFilePath).Msg("")

	data, err := tree.ImportYaml(emsConfigFilePath)
	if err != nil {
		log.Fatal().Err(err)
	}

	for _, child := range data.GetChildS("events").GetChildren() {
		emsName := child.GetChildContentS("name")
		if resolveEms := child.GetChildS("resolve_when_ems"); resolveEms != nil {
			issuingEmsNames = append(issuingEmsNames, emsName)
			resolvingEmsNames = append(resolvingEmsNames, resolveEms.GetChildContentS("name"))
			bookendEmsCount++
		} else {
			nonBookendEmsNames = append(nonBookendEmsNames, emsName)
			nonBookendEmsCount++
		}
	}

	log.Info().Msgf("Total ems configured: %d, Non-Bookend ems configured:%d, Bookend ems configured:%d", nonBookendEmsCount+bookendEmsCount, nonBookendEmsCount, bookendEmsCount)
}

func generateEvents(emsNames []string, poller *conf.Poller) *set.Set {
	supportedEms := set.New()
	var jsonValue []byte
	url := "https://" + poller.Addr + "/api/private/cli/event/generate"
	method := "POST"

	for _, ems := range emsNames {
		arg1 := "1"
		arg3 := "3"
		if ems == "arw.volume.state" {
			arg1 = "disable-in-progress"
		}
		if ems == "hm.alert.raised" || ems == "hm.alert.cleared" {
			arg3 = "RaidLeftBehindAggrAlert"
		}

		jsonValue = []byte(fmt.Sprintf(`{"message-name": "%s", "values": ["%s",2,"%s",4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20]}`, ems, arg1, arg3))

		data, err := util.SendPostReqAndGetRes(url, method, jsonValue, Admin, poller.Password)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load config")
		}
		if response := data["error"]; response != nil {
			errorDetail := response.(map[string]interface{})
			code := errorDetail["code"].(string)
			target := errorDetail["target"].(string)
			if !(code == "2" && target == "message-name") {
				supportedEms.Add(ems)
			}
		} else {
			supportedEms.Add(ems)
		}
	}

	return supportedEms
}
