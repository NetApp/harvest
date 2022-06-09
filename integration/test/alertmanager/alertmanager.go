package promAlerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
)

const PrometheusAlertURL string = "http://localhost:9090/api/v1/alerts"
const WebhookEndpoint string = "/webhook"
const WebhookPort string = ":8080"
const Webhook = "http://localhost" + WebhookPort + WebhookEndpoint

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

func GetAllAlertRules(dir string) ([]string, []string) {
	alertNames := make([]string, 0)
	exprList := make([]string, 0)
	alertRulesFilePath := dir + "/alert_rules.yml"
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

func SendNotification(jsonData []byte) {
	httpURL := Webhook
	response, err := http.Post(httpURL, "", bytes.NewBuffer(jsonData))
	utils.PanicIfNotNil(err)
	log.Debug().Str("URL", httpURL).Bytes("json", jsonData).Msg("notification sent")

	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		_, err := ioutil.ReadAll(response.Body)
		utils.PanicIfNotNil(err)
	} else {
		log.Error().Str("status", response.Status).Msg("Http post call failed with errors")
	}
}

func processWebhook(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != WebhookEndpoint {
		http.Error(writer, "404 page not found.", http.StatusNotFound)
		return
	}

	switch request.Method {
	case "POST":
		webhookData := make(map[string]interface{})
		err := json.NewDecoder(request.Body).Decode(&webhookData)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Info().Msg("Webhook payload data received")
		alertData := webhookData["data"].(map[string]interface{})
		alerts := alertData["alerts"].([]interface{})
		for _, alert := range alerts {
			// We only want to log the alertname, If required we can log all the labels
			for key, value := range alert.(map[string]interface{})["labels"].(map[string]interface{}) {
				if key == "alertname" {
					log.Info().Msgf("%s", value.(string))
				}
				//else {
				//	log.Info().Msgf("%s : %v\n", key, value.(string))
				//}
			}
			// If required we can log all the annotations
			//for key, value := range alert.(map[string]interface{})["annotations"].(map[string]interface{}) {
			//	log.Info().Msgf("%s : %v\n", key, value.(string))
			//}
		}
	default:
		fmt.Fprintf(writer, "Only POST call is supported on this webhook")
	}
}

func StartServer() {
	log.Info().Msgf("Server started to receive webhook at %s", Webhook)
	http.HandleFunc(WebhookEndpoint, processWebhook)
	go func() {
		log.Error().Err(http.ListenAndServe(WebhookPort, nil))
	}()
}
