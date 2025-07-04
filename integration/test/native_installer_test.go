package main

import (
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/installer"
	"log"
	"os"
	"testing"
)

func TestNativeInstall(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.InstallNative)
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
	token := cmds.CreateGrafanaToken()
	cmds.WriteToken(token)

	installObject, err2 := installer.GetInstaller(installer.NATIVE, path)
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
	cmds.AddPrometheusToGrafana()

}

func TestNativeStop(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.STOP)
	var path = os.Getenv("BUILD_PATH")
	installObject, err2 := installer.GetInstaller(installer.NATIVE, path)
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
