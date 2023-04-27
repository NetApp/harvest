package promAlerts

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"time"
)

const PrometheusAlertURL string = "http://localhost:9090/api/v1/alerts"
const TestClusterName = "umeng-aff300-05-06"
const TestNodeName = "umeng-aff300-06"
const Admin = "admin"

var volumeArwState = []string{
	`"disable-in-progress"`,
	`"disabled"`,
	`"dry-run"`,
	`"dry-run-paused"`,
	`"enable-paused"`,
	`"enabled"`,
}
var vserverArwState = []string{
	`"disabled"`,
	`"dry-run"`,
}

type PromAlert struct {
	message string
	count   int
}

func GetAlerts() (map[string]int, int) {
	now := time.Now()
	alertsData := make(map[string]int)
	totalAlerts := 0

	duration, _ := time.ParseDuration("3m15s")
	time.Sleep(duration)
	response, err := utils.GetResponseBody(PrometheusAlertURL)
	utils.PanicIfNotNil(err)

	results := gjson.GetManyBytes(response, "data")
	if results[0].Exists() {
		alerts := results[0].Get("alerts")

		for _, alert := range alerts.Array() {
			labels := alert.Get("labels").Map()
			alertData := labels["message"].String()
			clusterName := labels["cluster"].String()
			if clusterName == TestClusterName {
				alertsData[alertData] = alertsData[alertData] + 1
				totalAlerts++
			}
		}
	}
	log.Info().
		Int("alertsData", len(alertsData)).
		Int("totalAlerts", totalAlerts).
		Str("dur", time.Since(now).Round(time.Millisecond).String()).
		Msg("Get Prometheus alerts")
	return alertsData, totalAlerts
}

func GetEmsAlerts(dir string, fileName string) ([]string, []string, []string) {
	nonBookendEms := make([]string, 0)
	issuingEms := make([]string, 0)
	resolvingEms := make([]string, 0)
	nonBookendEmsCount := 0
	bookendEmsCount := 0

	emsConfigFilePath := dir + "/" + fileName
	log.Debug().Str("emsConfigFilePath", emsConfigFilePath).Msg("")

	data, err := tree.ImportYaml(emsConfigFilePath)
	if err != nil {
		utils.PanicIfNotNil(err)
	}

	for _, child := range data.GetChildS("events").GetChildren() {
		emsName := child.GetChildContentS("name")
		if resolveEms := child.GetChildS("resolve_when_ems"); resolveEms != nil {
			issuingEms = append(issuingEms, emsName)
			resolvingEms = append(resolvingEms, resolveEms.GetChildContentS("name"))
			bookendEmsCount++
		} else {
			nonBookendEms = append(nonBookendEms, emsName)
			nonBookendEmsCount++
		}
	}

	log.Info().Msgf("Total ems configured: %d, Non-Bookend ems configured:%d, Bookend ems configured:%d", nonBookendEmsCount+bookendEmsCount, nonBookendEmsCount, bookendEmsCount)
	return nonBookendEms, issuingEms, resolvingEms
}

func GenerateEvents(emsNames []string, nodeScopedEms []string) map[string]bool {
	supportedEms := make(map[string]bool)
	var jsonValue []byte
	err := conf.LoadHarvestConfig(installer.HarvestConfigFile)
	poller, err2 := conf.PollerNamed(TestClusterName)
	dc1, err3 := conf.PollerNamed("dc1")
	if err != nil && err2 != nil && err3 != nil {
		log.Fatal().Errs("errors", []error{err, err2, err3}).Msg("Failed to load config")
	}
	url := "https://" + poller.Addr + "/api/private/cli/event/generate"
	method := "POST"

	volumeArwCount := 0
	vserverArwCount := 0
	for _, ems := range emsNames {
		arg1 := "1"
		arg2 := "2"
		if ems == "arw.volume.state" {
			arg1 = volumeArwState[volumeArwCount]
			volumeArwCount++
		}
		if ems == "arw.vserver.state" {
			arg1 = vserverArwState[vserverArwCount]
			vserverArwCount++
		}
		// special case as arg order is different in issuing ems than resolving ems
		if ems == "sm.mediator.misconfigured" {
			arg2 = "1"
		}

		// Handle for node-scoped ems, Passing node-name as input
		if utils.Contains(nodeScopedEms, ems) {
			jsonValue = []byte(fmt.Sprintf(`{"message-name": "%s", "values": [%s,%s,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20], "node": "%s"}`, ems, arg1, arg2, TestNodeName))
		} else {
			jsonValue = []byte(fmt.Sprintf(`{"message-name": "%s", "values": [%s,%s,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20]}`, ems, arg1, arg2))
		}

		var data map[string]interface{}
		data = utils.SendPostReqAndGetRes(url, method, jsonValue, Admin, dc1.Password)
		if response := data["error"]; response != nil {
			errorDetail := response.(map[string]interface{})
			code := errorDetail["code"].(string)
			target := errorDetail["target"].(string)
			if !(code == "2" && target == "message-name") {
				supportedEms[ems] = true
			}
		} else {
			supportedEms[ems] = true
		}
	}

	return supportedEms
}
