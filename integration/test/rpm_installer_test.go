package main

import (
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"log/slog"
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
		slog.Error("Unable to initialize installer object", slog.String("object", installer.GRAFANA))
		panic(err)
	}
	if !installObject.Install() {
		panic(installer.GRAFANA + " installation is failed.")
	}
	token := utils.CreateGrafanaToken()
	utils.WriteToken(token)

	installObject, err2 := installer.GetInstaller(installer.RHEL, path)
	if err2 != nil {
		slog.Error("Unable to initialize installer object", slog.String("object", installer.RHEL))
		panic(err2)
	}
	if installObject.Install() {
		slog.Info("Installation is successful..")
	} else {
		slog.Error("Installation failed")
		panic("installation failed.")
	}
	harvestObj := new(installer.Harvest)
	if harvestObj.AllRunning("keyperf") {
		slog.Info("All pollers but keyperf are running")
	} else {
		t.Errorf("One or more pollers are not running.")
	}
	installObject, err = installer.GetInstaller(installer.PROMETHEUS, "prom/prometheus")
	if err != nil {
		slog.Error("Unable to initialize installer object", slog.String("object", installer.PROMETHEUS))
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
		slog.Info("Unable to initialize installer object")
		panic(err2)
	}
	if installObject.Stop() {
		slog.Info("Stop is successful..")
	} else {
		panic("Stop is failed.")
	}
}
