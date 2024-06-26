package main

import (
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
	"os"
	"testing"
)

func TestRHELInstall(t *testing.T) {
	utils.SkipIfMissing(t, utils.InstallRPM)
	var path = os.Getenv("BUILD_PATH")
	if path == "" {
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

	installObject, err2 := installer.GetInstaller(installer.RHEL, path)
	if err2 != nil {
		log.Println("Unable to initialize installer object")
		panic(err2)
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

func TestRHELStop(t *testing.T) {
	utils.SkipIfMissing(t, utils.STOP)
	var path = os.Getenv("BUILD_PATH")
	installObject, err2 := installer.GetInstaller(installer.RHEL, path)
	if err2 != nil {
		log.Println("Unable to initialize installer object")
		panic(err2)
	}
	if installObject.Stop() {
		log.Println("Stop is successful..")
	} else {
		panic("Stop is failed.")
	}
}
