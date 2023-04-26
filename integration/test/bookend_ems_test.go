package main

import (
	promAlerts "github.com/Netapp/harvest-automation/test/alert"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"testing"
	"time"
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
	numAlerts := 0
	emsConfigDir := utils.GetHarvestRootDir() + "/conf/ems/9.6.0"
	log.Info().Str("EmsConfigDir", emsConfigDir).Msg("Directory path")

	// Fetch ems configured in template
	_, issuingEmsNames, resolvingEmsNames = promAlerts.GetEmsAlerts(emsConfigDir, "ems.yaml")

	// Identify supported issuing ems names for the given cluster
	now := time.Now()
	supportedIssuingEms = promAlerts.GenerateEvents(issuingEmsNames, nodeScopedIssuingEmsList)
	log.Info().
		Int("supportedIssuingEms", len(supportedIssuingEms)).
		Str("dur", time.Since(now).Round(time.Millisecond).String()).
		Msg("Supported issuing ems")

	// Fetch previous prometheus alerts
	oldAlertsData, _ = promAlerts.GetAlerts()

	// Identify supported ems names for the given cluster
	now = time.Now()
	supportedResolvingEms = promAlerts.GenerateEvents(resolvingEmsNames, nodeScopedResolvingEmsList)
	log.Info().
		Int("supportedResolvingEms", len(supportedResolvingEms)).
		Str("dur", time.Since(now).Round(time.Millisecond).String()).
		Msg("Supported resolving ems")

	// Fetch current prometheus alerts
	newAlertsData, numAlerts = promAlerts.GetAlerts()
	if numAlerts == 0 {
		log.Info().Msg("No alerts found in prometheus")
	}
	log.Info().Int("numAlerts", numAlerts).Msg("Firing alerts")
}

func TestEmsTestSuite(t *testing.T) {
	utils.SkipIfMissing(t, utils.BookendEms)
	setupAlerts()

	// Evaluate bookend active ems events
	for _, issuingEms := range supportedIssuingEms {
		// If the issuingEms did not exist before, then ignore the test-case.
		if oldAlertsData[issuingEms] > 0 {
			v := newAlertsData[issuingEms] - oldAlertsData[issuingEms]
			if v >= 0 {
				t.Errorf("Extra bookend ems alerts raised event=%s, count=%d", issuingEms, v)
			}
		} else {
			log.Info().Str("issuingEms", issuingEms).Msg("Ignore. Did not exist before")
		}
	}
}
