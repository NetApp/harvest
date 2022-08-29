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
var supportedEms []string
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

// These bookend issuing ems are node scoped and have bookendKey as node-name only.
var nodeScopedIssuingEmsList = []string{
	"callhome.battery.low",
	"sp.ipmi.lost.shutdown",
	"sp.notConfigured",
}

type AlertRulesTestSuite struct {
	suite.Suite
}

func (suite *AlertRulesTestSuite) SetupSuite() {
	totalAlerts := 0
	emsConfigDir := utils.GetHarvestRootDir() + "/conf/ems/9.6.0"
	log.Info().Str("EmsConfigDir", emsConfigDir).Msg("Directory path")

	// Fetch ems configured in template
	totalEmsNames, _ = promAlerts.GetEmsAlerts(emsConfigDir, "ems.yaml")

	// Identify supported ems names for the given cluster
	supportedEms = promAlerts.GenerateEvents(totalEmsNames, nodeScopedIssuingEmsList)
	log.Info().Msgf("Total supported ems: %d", len(supportedEms))

	// Fetch prometheus alerts
	alertsData, totalAlerts = promAlerts.GetAlerts()
	if totalAlerts == 0 {
		log.Info().Msg("No alerts found in prometheus")
	}
	log.Info().Msgf("Total firing alerts %d", totalAlerts)
}

// Evaluate all active ems events
func (suite *AlertRulesTestSuite) TestEmsAlerts() {
	notFoundEms := make([]string, 0)

	for _, emsName := range supportedEms {
		if !(alertsData[emsName] != 0 || utils.Contains(skippedEmsList, emsName)) {
			notFoundEms = append(notFoundEms, emsName)
		}
	}
	if len(notFoundEms) > 0 {
		log.Error().Msg("The following ems alerts have not found.")
		assert.Fail(suite.T(), fmt.Sprintf("One or more ems alerts %s have not been raised", notFoundEms))
	}
}

func TestAlertRulesTestSuite(t *testing.T) {
	utils.SetupLogging()
	suite.Run(t, new(AlertRulesTestSuite))
}
