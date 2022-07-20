package promAlerts

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"time"
)

const PrometheusAlertURL string = "http://localhost:9090/api/v1/alerts"

var volumeArwState = []string{"\"disable-in-progress\"", "\"disabled\"", "\"dry-run\"", "\"dry-run-paused\"", "\"enable-paused\"", "\"enabled\""}
var vserverArwState = []string{"\"enabled\"", "\"dry-run\""}

type EmsData struct {
	name    string
	bookend bool
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
	resolvingEms := make([]EmsData, 0)
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
			totalEms = append(totalEms, EmsData{name: emsName, bookend: true})
			resolvingEms = append(resolvingEms, EmsData{name: resolveEms.GetChildContentS("name"), bookend: true})
			bookendEmsCount++
		} else {
			totalEms = append(totalEms, EmsData{name: emsName, bookend: false})
		}
	}

	log.Debug().Msgf("Total ems configured: %d, Bookend ems configured:%d", len(totalEms), bookendEmsCount)
	return totalEms, resolvingEms
}

func GenerateEvents(emsNames []EmsData) map[bool][]string {
	supportedEmsMap := make(map[bool][]string)
	addr, user, pass := GetPollerDetail()
	url := "https://" + addr + "/api/private/cli/event/generate"
	method := "POST"

	volumeArwCount := 0
	vserverArwCount := 0
	for _, emsName := range emsNames {
		value := "1"
		if emsName.name == "arw.volume.state" {
			value = volumeArwState[volumeArwCount]
			volumeArwCount++
		}
		if emsName.name == "arw.vserver.state" {
			value = vserverArwState[vserverArwCount]
			vserverArwCount++
		}

		jsonValue := []byte(fmt.Sprintf(`{"message-name": "%s", "values": [%s,2,3,4,5,6,7,8,9]}`, emsName.name, value))
		var data map[string]interface{}
		data = SendPostReqAndGetRes(url, method, jsonValue, user, pass)
		if response := data["error"]; response != nil {
			errorDetail := response.(map[string]interface{})
			code := errorDetail["code"].(string)
			target := errorDetail["target"].(string)
			if !(code == "2" && target == "message-name") {
				supportedEmsMap[emsName.bookend] = append(supportedEmsMap[emsName.bookend], emsName.name)
			}
		} else {
			supportedEmsMap[emsName.bookend] = append(supportedEmsMap[emsName.bookend], emsName.name)
		}
	}

	return supportedEmsMap
}

func SendPostReqAndGetRes(url string, method string,
	buf []byte, user string, pass string) map[string]interface{} {
	tlsConfig := &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(buf))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(user, pass)
	res, err := client.Do(req)
	utils.PanicIfNotNil(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	utils.PanicIfNotNil(err)
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	utils.PanicIfNotNil(err)
	return data
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
