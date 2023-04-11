//go:build bookendemstest

package main

import (
	"fmt"
	promAlerts "github.com/Netapp/harvest-automation/test/alert"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

var issuingEmsNames []string
var resolvingEmsNames []string
var supportedIssuingEms []string
var supportedResolvingEms []string
var oldAlertsData map[string]int
var newAlertsData map[string]int

// These bookend issuing ems are node scoped and have bookendKey as node-name only.
var nodeScopedIssuingEmsList = []string{
	"callhome.battery.low",
	"sp.ipmi.lost.shutdown",
	"sp.notConfigured",
}

// These bookend resolving ems are node scoped and have bookendKey as node-name only.
var nodeScopedResolvingEmsList = []string{
	"nvram.battery.charging.normal",
	"sp.heartbeat.resumed",
	"callhome.battery.low",
	"sp.ipmi.lost.shutdown",
	"sp.notConfigured",
}

type EmsTestSuite struct {
	suite.Suite
}

func (suite *EmsTestSuite) SetupSuite() {
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

// Evaluate bookend active ems events
func (suite *EmsTestSuite) TestBookendEmsAlerts() {
	foundBookendEms := make([]string, 0)

	for _, issuingEms := range supportedIssuingEms {
		// If the issuingEms wasn't exist prior then ignore the test-case.
		if oldAlertsData[issuingEms] > 0 {
			v := oldAlertsData[issuingEms] - newAlertsData[issuingEms]
			if v < 1 {
				foundBookendEms = append(foundBookendEms, issuingEms)
			}
		} else {
			log.Info().Msg("There is no active IssuingEms exist")
		}
	}
	if len(foundBookendEms) > 0 {
		log.Error().Msg("The following bookend ems alerts have found.")
		assert.Fail(suite.T(), fmt.Sprintf("One or more extra bookend ems alerts %s have been raised", foundBookendEms))
	}
}

func TestEmsTestSuite(t *testing.T) {
	utils.SetupLogging()
	suite.Run(t, new(EmsTestSuite))
}
