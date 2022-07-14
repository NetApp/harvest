//go:build alerttest

package main

import (
	"fmt"
	promAlerts "github.com/Netapp/harvest-automation/test/alert"
	"github.com/Netapp/harvest-automation/test/data"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/julienroland/usg"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"net/url"
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
	dir := utils.GetHarvestRootDir() + "/docker/prometheus"
	log.Info().Str("Dir", dir).Msg("directory path")

	// Fetch alert rules
	alertRuleNames = make([]string, 0)
	alertRules = make([]string, 0)
	promAlerts.GetAlertRules(&alertRuleNames, &alertRules, dir, "alert_rules.yml")
	promAlerts.GetAlertRules(&alertRuleNames, &alertRules, dir, "ems_alert_rules.yml")
	if len(alertRules) == 0 {
		assert.Fail(suite.T(), "No alert rules found @ "+dir)
	}
	log.Info().Int("count", len(alertRules)).Msg("Alert Rules")

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
	notMatchingAlerts := make([]string, 0)
	var isFailed = false

	for index, expr := range alertRules {
		activeAlertCount, failed := EvaluateExpr(expr)
		if failed {
			isFailed = true
		}
		log.Debug().Msgf("active alerts for %s is %d", alertRuleNames[index], activeAlertCount)
		for count := 0; count < activeAlertCount; count++ {
			activeAlerts = append(activeAlerts, alertRuleNames[index])
		}
	}

	if isFailed {
		assert.Fail(suite.T(), "Alert rules evaluation failed. Pls check logs above")
	} else {
		log.Info().Msg("Alert rules look good!")
	}

	log.Info().Msgf("active alerts name: %v", activeAlerts)

	// active alerts should be equal to prometheus alerts
	if len(activeAlerts) == len(alerts) {
		for _, activeAlert := range activeAlerts {
			if !(utils.Contains(alerts, activeAlert)) {
				notMatchingAlerts = append(notMatchingAlerts, activeAlert)
			}
		}
		if len(notMatchingAlerts) > 0 {
			log.Info().Msg("The following alerts were not matching successfully.")
			assert.Fail(suite.T(), fmt.Sprintf("One or more alerts %s were mot matching", notMatchingAlerts))
		}
	} else {
		assert.Fail(suite.T(), "Alert rules Test validation is failed due to count mismatch. Pls check logs above")
	}
}

// Evaluate expr and return number of active alert count with API error state
func EvaluateExpr(query string) (int, bool) {
	query = fmt.Sprintf("%s", query)
	log.Debug().Msg("Evaluating the alert rule " + query)
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", data.PrometheusURL,
		url.QueryEscape(query))
	resp, err := utils.GetResponse(queryURL)

	// when API has been failed
	if err != nil {
		fmt.Println(usg.Get.Cross, fmt.Sprintf(" ERROR: Failed to evaluate query [%s], error: %v", query, err))
		return 0, true
	}

	// API call succeed, but error due to bad/invalid data or any other reason
	if err == nil && gjson.Get(resp, "status").String() == "error" {
		errorType := gjson.Get(resp, "errorType")
		errorDetail := gjson.Get(resp, "error")
		if errorType.Exists() && errorDetail.Exists() {
			fmt.Println(usg.Get.Cross, fmt.Sprintf(" ERROR: Query [%s] has failed with %s, reason: [%s]", query, errorType.String(), errorDetail.String()))
			return 0, true
		}
	} else {
		// API call succeed with proper data
		value := gjson.Get(resp, "data.result")
		if value.Exists() && value.IsArray() && (len(value.Array()) > 0) {
			length := len(value.Array())
			return length, false
		}
	}
	return 0, false
}

func TestAlertRulesTestSuite(t *testing.T) {
	utils.SetupLogging()
	suite.Run(t, new(AlertRulesTestSuite))
}
