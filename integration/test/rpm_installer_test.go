package main

import (
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/installer"
	"log/slog"
	"os"
	"testing"
)

func TestRHELInstall(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.InstallRPM)
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
	token := cmds.CreateGrafanaToken()
	cmds.WriteToken(token)

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
	cmds.AddPrometheusToGrafana()
}

func TestRHELStop(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.STOP)
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
