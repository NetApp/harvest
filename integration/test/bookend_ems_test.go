package main

import (
	promAlerts "github.com/Netapp/harvest-automation/test/alert"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"testing"
)

var issuingEmsNames []string
var resolvingEmsNames []string
var supportedIssuingEms []string
var supportedResolvingEms []string
var oldAlertsData map[string]int
var newAlertsData map[string]int

// These bookend issuing EMS events are node scoped and have bookendKey as node-name only.
var nodeScopedIssuingEmsList = []string{
	"callhome.battery.low",
	"sp.ipmi.lost.shutdown",
	"sp.notConfigured",
}

// These bookend resolving EMS events are node scoped and have bookendKey as node-name only.
var nodeScopedResolvingEmsList = []string{
	"nvram.battery.charging.normal",
	"sp.heartbeat.resumed",
	"callhome.battery.low",
	"sp.ipmi.lost.shutdown",
	"sp.notConfigured",
}

func setupAlerts() {
	totalAlerts := 0
	emsConfigDir := utils.GetHarvestRootDir() + "/conf/ems/9.6.0"
	log.Info().Str("EmsConfigDir", emsConfigDir).Msg("Directory path")

	// Fetch ems configured in template
	_, issuingEmsNames, resolvingEmsNames = promAlerts.GetEmsAlerts(emsConfigDir, "ems.yaml")

	// Identify supported issuing ems names for the given cluster
	supportedIssuingEms = promAlerts.GenerateEvents(issuingEmsNames, nodeScopedIssuingEmsList)
	log.Info().Msgf("Total supported issuing ems: %d", len(supportedIssuingEms))

	// Fetch previous prometheus alerts
	oldAlertsData, _ = promAlerts.GetAlerts()

	// Identify supported ems names for the given cluster
	supportedResolvingEms = promAlerts.GenerateEvents(resolvingEmsNames, nodeScopedResolvingEmsList)
	log.Info().Msgf("Total supported resolving ems:%d", len(supportedResolvingEms))

	// Fetch current prometheus alerts
	newAlertsData, totalAlerts = promAlerts.GetAlerts()
	if totalAlerts == 0 {
		log.Info().Msg("No alerts found in prometheus")
	}
	log.Info().Msgf("Total firing alerts %d", totalAlerts)
}

func TestEmsTestSuite(t *testing.T) {
	utils.SkipIfMissing(t, utils.BookendEms)
	setupAlerts()

	// Evaluate bookend active ems events
	foundBookendEms := make([]string, 0)

	for _, issuingEms := range supportedIssuingEms {
		// If the issuingEms did not exit before, then ignore the test-case.
		if oldAlertsData[issuingEms] > 0 {
			v := oldAlertsData[issuingEms] - newAlertsData[issuingEms]
			if v < 1 {
				foundBookendEms = append(foundBookendEms, issuingEms)
			}
		} else {
			log.Info().Str("issuingEms", issuingEms).Msg("There is no active issuingEms")
		}
	}
	if len(foundBookendEms) > 0 {
		log.Error().Strs("foundBookendEms", foundBookendEms).Msg("Unexpected bookendEms found")
		t.Errorf("One or more extra bookend ems alerts %s have been raised", foundBookendEms)
	}
}
