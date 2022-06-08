package promAlerts

import (
	"encoding/json"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/tidwall/gjson"
)

const PrometheusAlertURL string = "http://localhost:9090/api/v1/alerts"

// AlertBody sent as notification
type AlertBody struct {
	alertName   string
	labels      map[string]string
	annotations map[string]string
}

func GetAlerts() ([]string, []AlertBody) {
	var (
		alerts      []AlertBody
		alertNames  []string
		name        string
		labels      map[string]string
		annotations map[string]string
	)

	body, err := utils.GetResponseBody(PrometheusAlertURL)
	utils.PanicIfNotNil(err)

	results := gjson.GetManyBytes(body, "data")
	alertData := results[0].Get("alerts")

	for _, alert := range alertData.Array() {
		name = ""
		labels = make(map[string]string)
		annotations = make(map[string]string)

		for i, j := range alert.Get("labels").Map() {
			if i == "alertname" {
				name = j.String()
			} else {
				labels[i] = j.String()
			}
		}
		for i, j := range alert.Get("annotations").Map() {
			annotations[i] = j.String()
		}

		alertDetail := AlertBody{
			alertName:   name,
			labels:      labels,
			annotations: annotations,
		}

		alerts = append(alerts, alertDetail)
		alertNames = append(alertNames, name)
	}
	return alertNames, alerts
}

func GenerateJson(alert AlertBody) []byte {
	alertJson := make(map[string]interface{})
	sections := make([]map[string]interface{}, 0)
	sectionsMap := make(map[string]interface{})
	facts := make([]map[string]string, 0)

	alertJson["@type"] = "MessageCard"
	alertJson["@context"] = "http://schema.org/extensions"
	alertJson["themeColor"] = "0076D7"
	alertJson["summary"] = "Harvest Alerts"

	for name, value := range alert.labels {
		factsMap := make(map[string]string)
		factsMap["name"] = name
		factsMap["value"] = value
		facts = append(facts, factsMap)
	}

	for name, value := range alert.annotations {
		factsMap := make(map[string]string)
		factsMap["name"] = name
		factsMap["value"] = value
		facts = append(facts, factsMap)
	}

	sectionsMap["activityTitle"] = alert.alertName
	sectionsMap["facts"] = facts

	sections = append(sections, sectionsMap)
	alertJson["sections"] = sections
	//log.Debug().Msgf("alert json %s", alertJson)

	//Convert AlertBody to byte using Json.Marshal, Ignoring error.
	jsonBody, err := json.Marshal(alertJson)
	utils.PanicIfNotNil(err)
	return jsonBody
}

func GetAllAlertRules(dir string) ([]string, []string) {
	alertNames := make([]string, 0)
	exprList := make([]string, 0)
	alertRulesFilePath := dir + "/alert_rules.yml"
	//log.Info().Str("alertRulesFilePath", alertRulesFilePath).Msg("alert rules file path")

	data, err := tree.ImportYaml(alertRulesFilePath)
	if err != nil {
		utils.PanicIfNotNil(err)
	}

	for _, v := range data.GetChildS("groups").GetChildren() {
		if v.GetNameS() == "rules" {
			for _, a := range v.GetChildren() {
				if a.GetNameS() == "alert" {
					alertname := a.GetContentS()
					alertNames = append(alertNames, alertname)
				}
				if a.GetNameS() == "expr" {
					alertexp := a.GetContentS()
					exprList = append(exprList, alertexp)
				}
			}
		}
	}
	return alertNames, exprList
}
