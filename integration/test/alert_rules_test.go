//go:build alerttest

package main

import (
	"fmt"
	promAlerts "github.com/Netapp/harvest-automation/test/alertmanager"
	"github.com/Netapp/harvest-automation/test/data"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"net/url"
	"strings"
	"testing"
)

var alertRuleNames []string
var alertRules []string
var alerts []string
var alertResponse []byte

type AlertRulesTestSuite struct {
	suite.Suite
}

func (suite *AlertRulesTestSuite) SetupSuite() {
	// Start the webhook http server
	promAlerts.StartServer()
	dir := utils.GetHarvestRootDir() + "/integration/test"
	log.Info().Str("Dir", dir).Msg("directory path")

	// Fetch alert rules
	alertRuleNames, alertRules = promAlerts.GetAllAlertRules(dir)
	if len(alertRules) == 0 {
		assert.Fail(suite.T(), "No alert rules found @ "+dir)
	}
	log.Info().Int("rules", len(alertRules)).Msg("Alert Rules")

	// Fetch prometheus alerts
	alerts, alertResponse = promAlerts.GetAlerts()
	if len(alerts) == 0 {
		log.Info().Msg("No alerts found in prometheus")
	}
	log.Info().Msgf("Total firing alerts %d", len(alerts))
}

// Evaluate alert rule expressions
func (suite *AlertRulesTestSuite) TestExpression() {
	activeAlerts := make([]string, 0)

	for index, expr := range alertRules {
		activeAlertCount := EvaluateExpr(expr)
		log.Debug().Msgf("active alerts for %s is %d", alertRuleNames[index], activeAlertCount)
		for count := 0; count < activeAlertCount; count++ {
			activeAlerts = append(activeAlerts, alertRuleNames[index])
		}
	}
	log.Info().Msgf("active alerts name: %s", strings.Join(activeAlerts, ", "))

	// active alerts should be equal to prometheus alerts
	if len(activeAlerts) == len(alerts) {
		alertArray := strings.Join(alerts, ",")
		for _, activeAlert := range activeAlerts {
			if !strings.Contains(alertArray, activeAlert) {
				assert.Fail(suite.T(), "Test validation is failed for %s. Pls check logs above", activeAlert)
			}
		}

		// Send alert notification
		promAlerts.SendNotification(alertResponse)
	} else {
		assert.Fail(suite.T(), "Alert rules Test validation is failed due to count mismatch. Pls check logs above")
	}
}

// Evaluate expr and return number of active alert count
func EvaluateExpr(query string) int {
	query = fmt.Sprintf("%s", query)
	log.Debug().Msg("Evaluating the alert rule " + query)
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", data.PrometheusURL,
		url.QueryEscape(query))
	resp, err := utils.GetResponse(queryURL)
	if err == nil && gjson.Get(resp, "status").String() == "success" {
		value := gjson.Get(resp, "data.result")
		if value.Exists() && value.IsArray() && (len(value.Array()) > 0) {
			length := len(value.Array())
			return length
		}
		return 0
	}
	return 0
}

func TestAlertRulesTestSuite(t *testing.T) {
	utils.SetupLogging()
	suite.Run(t, new(AlertRulesTestSuite))
}
