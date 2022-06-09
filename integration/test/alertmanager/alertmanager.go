package promAlerts

import (
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

const PrometheusAlertURL string = "http://localhost:9090/api/v1/alerts"

func GetAlerts() ([]string, []byte) {
	alertNames := make([]string, 0)

	response, err := utils.GetResponseBody(PrometheusAlertURL)
	utils.PanicIfNotNil(err)

	results := gjson.GetManyBytes(response, "data")
	alertData := results[0].Get("alerts")

	for _, alert := range alertData.Array() {
		for key, value := range alert.Get("labels").Map() {
			if key == "alertname" {
				alertNames = append(alertNames, value.String())
				break
			}
		}
	}
	return alertNames, response
}

func GetAlertRules(alertRuleNames *[]string, alertRules *[]string, dir string, fileName string) {
	alertRulesFilePath := dir + "/" + fileName
	log.Info().Str("alertRulesFilePath", alertRulesFilePath).Msg("alert rules file path")

	data, err := tree.ImportYaml(alertRulesFilePath)
	if err != nil {
		utils.PanicIfNotNil(err)
	}

	for _, v := range data.GetChildS("groups").GetChildren() {
		if v.GetNameS() == "rules" {
			for _, a := range v.GetChildren() {
				if a.GetNameS() == "alert" {
					alertname := a.GetContentS()
					*alertRuleNames = append(*alertRuleNames, alertname)
				}
				if a.GetNameS() == "expr" {
					alertexp := a.GetContentS()
					*alertRules = append(*alertRules, alertexp)
				}
			}
		}
	}
}
