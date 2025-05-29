package installer

import (
	"errors"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/errs"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"strings"
)

type RPM struct {
	path string
}

func (r *RPM) Init(path string) {
	r.path = path
}

func (r *RPM) Install() bool {
	harvestFile := "harvest.yml"
	rpmFileName := "harvest.rpm"
	cmds.RemoveSafely(rpmFileName)
	err := cmds.DownloadFile(rpmFileName, r.path)
	if err != nil {
		panic(err)
	}
	slog.Info("Downloaded: " + r.path)
	Uninstall()
	harvestObj := new(Harvest)
	slog.Info("Installing " + rpmFileName)
	installOutput, err := cmds.Run("yum", "install", "-y", rpmFileName)
	if err != nil {
		slog.Error("", slogx.Err(err))
		panic(err)
	}
	slog.Info(installOutput)
	slog.Info("Stopping harvest")
	harvestObj.Stop()
	_, err = cmds.Run("cp", harvestFile, HarvestHome+"/"+harvestFile)
	if err != nil {
		return false
	} // use file directly from the repo
	harvestObj.Start()
	status := harvestObj.AllRunning("keyperf")
	asupExecPath := HarvestHome + "/autosupport/asup"
	isValidAsup := harvestObj.IsValidAsup(asupExecPath)
	return status && isValidAsup
}

func (r *RPM) Upgrade() bool {
	rpmFileName := "harvest.rpm"
	cmds.RemoveSafely(rpmFileName)
	harvestObj := new(Harvest)
	if !harvestObj.AllRunning("keyperf") {
		errs.PanicIfNotNil(errors.New("pollers are not in a running state before upgrade"))
	}
	versionCmd := []string{"-qa", "harvest"}
	out, err := cmds.Run("rpm", versionCmd...)
	if err != nil {
		slog.Error("", slogx.Err(err))
		panic(err)
	}
	previousVersion := strings.TrimSpace(out)
	err = cmds.DownloadFile(rpmFileName, r.path)
	errs.PanicIfNotNil(err)
	slog.Info("Downloaded: " + r.path)
	slog.Info("Updating " + rpmFileName)
	installOutput, _ := cmds.Run("yum", "upgrade", "-y", rpmFileName)
	slog.Info(installOutput)
	out, _ = cmds.Run("rpm", versionCmd...)
	installedVersion := strings.TrimSpace(out)
	if previousVersion == installedVersion {
		errs.PanicIfNotNil(errors.New("upgrade is failed"))
	}
	_, _ = cmds.Run("cp", GetPerfFileWithQosCounters(ZapiPerfDefaultFile, "defaultZapi.yaml"), HarvestHome+"/"+ZapiPerfDefaultFile)
	_, _ = cmds.Run("cp", GetPerfFileWithQosCounters(RestPerfDefaultFile, "defaultRest.yaml"), HarvestHome+"/"+RestPerfDefaultFile)
	harvestObj.Stop()
	harvestObj.Start()
	status := harvestObj.AllRunning()
	asupExecPath := HarvestHome + "/autosupport/asup"
	isValidAsup := harvestObj.IsValidAsup(asupExecPath)
	return status && isValidAsup
}

func (r *RPM) Stop() bool {
	if cmds.FileExists(HarvestHome) {
		harvestObj := new(Harvest)
		if cmds.FileExists(HarvestHome + "/bin/harvest") {
			if harvestObj.AllRunning() {
				harvestObj.Stop()
			}
		}
	}
	return true
}
