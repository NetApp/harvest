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
type EmsData struct {
	EmsName      string
	resolvingEms string
}

func GetAlerts() (map[string]int, int) {
	alertsData := make(map[string]int)
	totalAlerts := 0

	time.Sleep(3 * time.Minute)
	response, err := utils.GetResponseBody(PrometheusAlertURL)
	utils.PanicIfNotNil(err)

	results := gjson.GetManyBytes(response, "data")
	if results[0].Exists() {
		alerts := results[0].Get("alerts")

		for _, alert := range alerts.Array() {
			labels := alert.Get("labels").Map()
			alertData := labels["message"].String()
			alertsData[alertData] = alertsData[alertData] + 1
			totalAlerts++
		}
	}
	return alertsData, totalAlerts
}

func GetEmsAlerts(dir string, fileName string) ([]EmsData, []EmsData) {
	totalEms := make([]EmsData, 0)
	bookendEms := make([]EmsData, 0)
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
			bookendEms = append(bookendEms, EmsData{EmsName: emsName, resolvingEms: resolveEms.GetChildContentS("name")})
			bookendEmsCount++
		}
		totalEms = append(totalEms, EmsData{EmsName: emsName})
	}

	log.Debug().Msgf("Total ems configured: %d, Bookend ems configured:%d", len(totalEms), bookendEmsCount)
	return totalEms, bookendEms
}

func GenerateEvents(emsNames []EmsData, nodeScopedEms []string) []string {
	supportedEms := make([]string, 0)
	var jsonValue []byte
	addr, user, pass, nodeName := GetPollerDetail()
	url := "https://" + addr + "/api/private/cli/event/generate"
	method := "POST"

	volumeArwCount := 0
	vserverArwCount := 0
	for _, e := range emsNames {
		value := "1"
		ems := ""
		if ems = e.resolvingEms; ems == "" {
			ems = e.EmsName
		}
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
				supportedEms = append(supportedEms, e.EmsName)
			}
		} else {
			supportedEms = append(supportedEms, e.EmsName)
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

	if poller, err = conf.PollerNamed("umeng_aff300"); err != nil {
		utils.PanicIfNotNil(err)
	}

	return poller.Addr, "admin", poller.Password, "umeng-aff300-06"
}
