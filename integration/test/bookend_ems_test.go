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

var alertsData []string
var totalEmsNames []string
var bookendEmsNames []string

type EmsTestSuite struct {
	suite.Suite
}

func (suite *EmsTestSuite) SetupSuite() {
	emsConfigDir := utils.GetHarvestRootDir() + "/conf/ems/9.6.0"
	log.Info().Str("EmsConfigDir", emsConfigDir).Msg("Directory path")

	// Fetch ems configured in template
	totalEmsNames, bookendEmsNames = promAlerts.GetEmsAlerts(emsConfigDir, "ems.yaml")

	// Fetch prometheus alerts
	alertsData = promAlerts.GetAlerts()
	if len(alertsData) == 0 {
		log.Info().Msg("No alerts found in prometheus")
	}
	log.Info().Msgf("Total firing alerts %d", len(alertsData))
}

// Evaluate bookend active ems events
func (suite *EmsTestSuite) TestBookendEmsAlerts() {
	foundBookendEms := make([]string, 0)

	// active non-bookend ems alerts should be equal to or more than non-bookend ems configured in template
	if len(alertsData) >= (len(totalEmsNames) - len(bookendEmsNames)) {
		for _, bookendEmsName := range bookendEmsNames {
			if utils.Contains(alertsData, bookendEmsName) {
				foundBookendEms = append(foundBookendEms, bookendEmsName)
			}
		}
		if len(foundBookendEms) > 0 {
			log.Error().Msg("The following bookend ems alerts have found.")
			assert.Fail(suite.T(), fmt.Sprintf("One or more extra bookend ems alerts %s have been raised", foundBookendEms))
		}
	} else {
		assert.Fail(suite.T(), "Bookend ems test validation is failed due to having extra ems alerts. Pls check logs above")
	}
}

func TestEmsTestSuite(t *testing.T) {
	utils.SetupLogging()
	suite.Run(t, new(EmsTestSuite))
}
