package promAlerts

import (
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

const PrometheusAlertURL string = "http://localhost:9090/api/v1/alerts"

func GetAlerts() []string {
	alertsData := make([]string, 0)

	response, err := utils.GetResponseBody(PrometheusAlertURL)
	utils.PanicIfNotNil(err)

	results := gjson.GetManyBytes(response, "data")
	if results[0].Exists() {
		alerts := results[0].Get("alerts")

		for _, alert := range alerts.Array() {
			labels := alert.Get("labels").Map()
			alertData := labels["message"].String()
			alertsData = append(alertsData, alertData)
		}
	}
	return alertsData
}

func GetEmsAlerts(dir string, fileName string) ([]string, []string) {
	totalEmsNames := make([]string, 0)
	bookendEmsNames := make([]string, 0)

	emsConfigFilePath := dir + "/" + fileName
	log.Debug().Str("emsConfigFilePath", emsConfigFilePath).Msg("")

	data, err := tree.ImportYaml(emsConfigFilePath)
	if err != nil {
		utils.PanicIfNotNil(err)
	}

	for _, child := range data.GetChildS("events").GetChildren() {
		emsName := child.GetChildContentS("name")
		totalEmsNames = append(totalEmsNames, emsName)

		if child.GetChildS("resolve_when_ems") != nil {
			bookendEmsNames = append(bookendEmsNames, emsName)
		}
	}

	log.Debug().Msgf("Total ems configured: %d, Bookend ems configured:%d", len(totalEmsNames), len(bookendEmsNames))
	return totalEmsNames, bookendEmsNames
}