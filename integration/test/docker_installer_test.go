//+build install_docker

package main

import (
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
	"os"
	"testing"
)

func TestDockerInstall(t *testing.T) {

	var path = os.Getenv("BUILD_PATH")
	if len(path) == 0 {
		panic("BUILD_PATH variable is not set.")
	}
	installObject, err := installer.GetInstaller(installer.GRAFANA, "grafana/grafana")
	if err != nil {
		log.Println("Unable to initialize installer object for " + installer.GRAFANA)
		panic(err)
	}
	if !installObject.Install() {
		panic(installer.GRAFANA + " installation is failed.")
	}
	token := utils.CreateGrafanaToken()
	utils.WriteToken(token)
	installObject, err = installer.GetInstaller(installer.DOCKER, path)
	if err != nil {
		log.Println("Unable to initialize installer object")
		panic(err)
	}
	if installObject.Install() {
		log.Println("Harvest installation is successful..")
	} else {
		panic("installation is failed.")
	}

	installObject, err = installer.GetInstaller(installer.PROMETHEUS, "prom/prometheus")
	if err != nil {
		log.Println("Unable to initialize installer object for " + installer.PROMETHEUS)
		panic(err)
	}
	if !installObject.Install() {
		panic(installer.PROMETHEUS + " installation is failed.")
	}
	utils.AddPPrometheusToGrafana()
}
