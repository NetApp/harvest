package main

import (
	promAlerts "github.com/Netapp/harvest-automation/test/alert"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"testing"
)

var nonBookendEmsNames []string
var supportedNonBookendEms []string
var alertsData map[string]int

// These EMS events are node scoped, and are not always raised from ONTAP, even when simulated via POST call.
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

func setup() {
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

func TestAlertRules(t *testing.T) {
	utils.SkipIfMissing(t, utils.NonBookendEms)
	setup()

	// Evaluate all active ems events
	notFoundNonBookendEms := make([]string, 0)

	for _, nonBookendEms := range supportedNonBookendEms {
		if !(alertsData[nonBookendEms] != 0 || utils.Contains(skippedEmsList, nonBookendEms)) {
			notFoundNonBookendEms = append(notFoundNonBookendEms, nonBookendEms)
		}
	}
	if len(notFoundNonBookendEms) > 0 {
		log.Error().Strs("notFoundNonBookendEms", notFoundNonBookendEms).Msg("Expected all to be found")
		t.Errorf("One or more ems alerts %s have not been raised", notFoundNonBookendEms)
	}
}
