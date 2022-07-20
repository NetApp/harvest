//go:build allemstest

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

var totalEmsNames []promAlerts.EmsData
var bookendEmsNames []string
var supportedEms map[bool][]string
var alertsData []string

type AlertRulesTestSuite struct {
	suite.Suite
}

func (suite *AlertRulesTestSuite) SetupSuite() {
	emsConfigDir := utils.GetHarvestRootDir() + "/conf/ems/9.6.0"
	log.Info().Str("EmsConfigDir", emsConfigDir).Msg("Directory path")

	// Fetch ems configured in template
	totalEmsNames, _ = promAlerts.GetEmsAlerts(emsConfigDir, "ems.yaml")

	// Identify supported ems names for the given cluster
	supportedEms = promAlerts.GenerateEvents(totalEmsNames)
	log.Info().Msgf("Total supported ems: %d, supported Bookend ems:%d", len(supportedEms[true])+len(supportedEms[false]), len(supportedEms[true]))

	// Fetch prometheus alerts
	alertsData = promAlerts.GetAlerts()
	if len(alertsData) == 0 {
		log.Info().Msg("No alerts found in prometheus")
	}
	log.Info().Msgf("Total firing alerts %d", len(alertsData))
}

// Evaluate all active ems events
func (suite *AlertRulesTestSuite) TestEmsAlerts() {
	notFoundEms := make([]string, 0)

	// active alerts should be equal to or more than ems configured in template
	if len(alertsData) >= (len(supportedEms[false]) + len(supportedEms[true])) {
		for _, emsName := range supportedEms[false] {
			if !(utils.Contains(alertsData, emsName)) {
				notFoundEms = append(notFoundEms, emsName)
			}
		}
		for _, emsName := range supportedEms[true] {
			if !(utils.Contains(alertsData, emsName)) {
				notFoundEms = append(notFoundEms, emsName)
			}
		}
		if len(notFoundEms) > 0 {
			log.Error().Msg("The following ems alerts have not found.")
			assert.Fail(suite.T(), fmt.Sprintf("One or more ems alerts %s have not been raised", notFoundEms))
		}
	} else {
		assert.Fail(suite.T(), "Ems alerts test validation is failed due to missing few ems alerts. Pls check logs above")
	}
}

func TestAlertRulesTestSuite(t *testing.T) {
	utils.SetupLogging()
	suite.Run(t, new(AlertRulesTestSuite))
}
