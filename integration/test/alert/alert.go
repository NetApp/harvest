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

var volumeArwState = []string{"\"disable-in-progress\"", "\"disabled\"", "\"dry-run\"", "\"dry-run-paused\"", "\"enable-paused\"", "\"enabled\""}
var vserverArwState = []string{"\"enabled\"", "\"dry-run\""}

type EmsData struct {
	EmsName      string
	resolvingEms string
}

func GetAlerts() []string {
	alertsData := make([]string, 0)

	time.Sleep(3 * time.Minute)
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

func GenerateEvents(emsNames []EmsData) []string {
	supportedEms := make([]string, 0)
	addr, user, pass := GetPollerDetail()
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

		jsonValue := []byte(fmt.Sprintf(`{"message-name": "%s", "values": [%s,2,3,4,5,6,7,8,9]}`, ems, value))
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

func GetPollerDetail() (string, string, string) {
	var (
		addr string
		err  error
		pass string
	)

	filename := "/harvest.yml"
	user := "admin"
	err = conf.LoadHarvestConfig(utils.GetConfigDir() + filename)
	utils.PanicIfNotNil(err)
	poller := conf.Config.Pollers["umeng_aff300"]
	if poller != nil {
		addr = poller.Addr
		pass = poller.Password
	}
	return addr, user, pass
}
