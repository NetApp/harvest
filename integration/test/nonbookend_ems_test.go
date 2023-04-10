//go:build nonbookendemstest

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

var nonBookendEmsNames []string
var supportedNonBookendEms []string
var alertsData map[string]int

// These few ems are node scoped and They won't be raised always from ONTAP even if we simulate via POST call.
var skippedEmsList = []string{
	"callhome.hainterconnect.down",
	"fabricpool.full",
	"fabricpool.nearly.full",
	"Nblade.cifsNoPrivShare",
	"Nblade.nfsV4PoolExhaust",
	"Nblade.vscanBadUserPrivAccess",
	"Nblade.vscanNoRegdScanner",
	"Nblade.vscanConnInactive",
	"cloud.aws.iamNotInitialized",
	"scsitarget.fct.port.full",
}

type AlertRulesTestSuite struct {
	suite.Suite
}

func (suite *AlertRulesTestSuite) SetupSuite() {
	totalAlerts := 0
	emsConfigDir := utils.GetHarvestRootDir() + "/conf/ems/9.6.0"
	log.Info().Str("EmsConfigDir", emsConfigDir).Msg("Directory path")

	// Fetch non-bookend ems configured in template
	nonBookendEmsNames, _, _ = promAlerts.GetEmsAlerts(emsConfigDir, "ems.yaml")

	// Identify supported non-bookend ems names for the given cluster
	supportedNonBookendEms = promAlerts.GenerateEvents(nonBookendEmsNames, []string{})
	log.Info().Msgf("Total supported non-bookend ems: %d", len(supportedNonBookendEms))

	// Fetch prometheus alerts
	alertsData, totalAlerts = promAlerts.GetAlerts()
	if totalAlerts == 0 {
		log.Info().Msg("No alerts found in prometheus")
	}
	log.Info().Msgf("Total firing alerts %d", totalAlerts)
}

// Evaluate all active ems events
func (suite *AlertRulesTestSuite) TestEmsAlerts() {
	notFoundNonBookendEms := make([]string, 0)

	for _, nonBookendEms := range supportedNonBookendEms {
		if !(alertsData[nonBookendEms] != 0 || utils.Contains(skippedEmsList, nonBookendEms)) {
			notFoundNonBookendEms = append(notFoundNonBookendEms, nonBookendEms)
		}
	}
	if len(notFoundNonBookendEms) > 0 {
		log.Error().Msg("The following ems alerts have not found.")
		assert.Fail(suite.T(), fmt.Sprintf("One or more ems alerts %s have not been raised", notFoundNonBookendEms))
	}
}

func TestAlertRulesTestSuite(t *testing.T) {
	utils.SetupLogging()
	suite.Run(t, new(AlertRulesTestSuite))
}
