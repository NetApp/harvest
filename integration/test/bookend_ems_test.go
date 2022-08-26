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

var resolvingEmsNames []promAlerts.EmsData
var bookendEmsNames []string
var supportedEms []string
var oldAlertsData map[string]int
var newAlertsData map[string]int

// These bookend ems (both issuing and resolving ems) are node scoped and have bookendKey as node-name only. They won't be raised/resolved always from ONTAP even if we simulate via POST call.
var skippedBookendEmsList = []string{
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
	_, resolvingEmsNames = promAlerts.GetEmsAlerts(emsConfigDir, "ems.yaml")

	// Fetch previous prometheus alerts
	oldAlertsData, _ = promAlerts.GetAlerts()

	// Identify supported ems names for the given cluster
	supportedEms = promAlerts.GenerateEvents(resolvingEmsNames, skippedBookendEmsList)
	log.Info().Msgf("Supported Bookend ems:%d", len(supportedEms))

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

	for _, bookendEmsName := range supportedEms {
		v := oldAlertsData[bookendEmsName] - newAlertsData[bookendEmsName]
		if v < 1 {
			foundBookendEms = append(foundBookendEms, bookendEmsName)
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
