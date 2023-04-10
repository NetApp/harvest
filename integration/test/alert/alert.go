package promAlerts

import (
	"fmt"
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
const User = "admin"

var volumeArwState = []string{
	`"disable-in-progress"`,
	`"disabled"`,
	`"dry-run"`,
	`"dry-run-paused"`,
	`"enable-paused"`,
	`"enabled"`,
}
var vserverArwState = []string{
	`"enabled"`,
	`"dry-run"`,
}

type PromAlert struct {
	message string
	count   int
}

func GetAlerts() (map[string]int, int) {
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

func GenerateEvents(emsNames []string, nodeScopedEms []string) []string {
	supportedEms := make([]string, 0)
	var jsonValue []byte
	addr, user, pass, nodeName := GetPollerDetail()
	url := "https://" + addr + "/api/private/cli/event/generate"
	method := "POST"

	volumeArwCount := 0
	vserverArwCount := 0
	for _, ems := range emsNames {
		value := "1"
		if ems == "arw.volume.state" {
			value = volumeArwState[volumeArwCount]
			volumeArwCount++
		}
		if ems == "arw.vserver.state" {
			value = vserverArwState[vserverArwCount]
			vserverArwCount++
		}

		// Handle for node-scoped ems, Passing node-name as input
		if utils.Contains(nodeScopedEms, ems) {
			jsonValue = []byte(fmt.Sprintf(`{"message-name": "%s", "values": [%s,2,3,4,5,6,7,8,9], "node": "%s"}`, ems, value, nodeName))
		} else {
			jsonValue = []byte(fmt.Sprintf(`{"message-name": "%s", "values": [%s,2,3,4,5,6,7,8,9]}`, ems, value))
		}

		var data map[string]interface{}
		data = utils.SendPostReqAndGetRes(url, method, jsonValue, user, pass)
		if response := data["error"]; response != nil {
			errorDetail := response.(map[string]interface{})
			code := errorDetail["code"].(string)
			target := errorDetail["target"].(string)
			if !(code == "2" && target == "message-name") {
				supportedEms = append(supportedEms, ems)
			}
		} else {
			supportedEms = append(supportedEms, ems)
		}
	}

	return supportedEms
}

func GetPollerDetail() (string, string, string, string) {
	var (
		err    error
		poller *conf.Poller
	)

	if err = conf.LoadHarvestConfig(utils.GetConfigDir() + "/harvest.yml"); err != nil {
		utils.PanicIfNotNil(err)
	}

	if poller, err = conf.PollerNamed(TestClusterName); err != nil {
		utils.PanicIfNotNil(err)
	}

	return poller.Addr, User, poller.Password, TestNodeName
}
