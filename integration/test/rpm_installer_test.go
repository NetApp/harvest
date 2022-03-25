//go:build install_rpm

package main

import (
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
	"os"
	"testing"
)

func TestRHELInstall(t *testing.T) {
	utils.SetupLogging()
	var path = os.Getenv("BUILD_PATH")
	if len(path) == 0 {
		panic("BUILD_PATH variable is not set.")
	}
	utils.UseCertFile()
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

	installObject, error := installer.GetInstaller(installer.RHEL, path)
	if error != nil {
		log.Println("Unable to initialize installer object")
		panic(error)
	}
	if installObject.Install() {
		log.Println("Installation is successful..")
	} else {
		log.Println("Setup completed")
		panic("installation is failed.")
	}
	harvestObj := new(installer.Harvest)
	if harvestObj.AllRunning() {
		log.Println("All pollers are running")
	} else {
		t.Errorf("One or more pollers are not running.")
	}
	installObject, err = installer.GetInstaller(installer.PROMETHEUS, "prom/prometheus")
	if err != nil {
		log.Println("Unable to initialize installer object for " + installer.PROMETHEUS)
		panic(err)
	}
	if !installObject.Install() {
		panic(installer.PROMETHEUS + " installation is failed.")
	}
	utils.AddPrometheusToGrafana()

}
