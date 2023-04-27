package main

import (
	promAlerts "github.com/Netapp/harvest-automation/test/alert"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"testing"
)

var supportedNonBookendEms map[string]bool
var alertsData map[string]int

func setup() {
	totalAlerts := 0
	// testing this non-bookend ems in CI
	var nonBookendEmsName = []string{"wafl.vol.autoSize.done"}

	// Check if non-bookend ems name is supported for the given cluster
	supportedNonBookendEms = promAlerts.GenerateEvents(nonBookendEmsName, []string{})
	log.Info().Msgf("Supported non-bookend ems: %v", supportedNonBookendEms)

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

	for nonBookendEms, _ := range supportedNonBookendEms {
		if alertsData[nonBookendEms] == 0 {
			notFoundNonBookendEms = append(notFoundNonBookendEms, nonBookendEms)
		}
	}
	if len(notFoundNonBookendEms) > 0 {
		log.Error().Strs("notFoundNonBookendEms", notFoundNonBookendEms).Msg("Expected all to be found")
		t.Errorf("Ems alerts %s have not been raised", notFoundNonBookendEms)
	} else {
		log.Info().Msg("Non bookend ems test passed")
	}
}
