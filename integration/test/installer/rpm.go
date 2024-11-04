package installer

import (
	"errors"
	"github.com/Netapp/harvest-automation/test/utils"
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
	utils.RemoveSafely(rpmFileName)
	err := utils.DownloadFile(rpmFileName, r.path)
	if err != nil {
		panic(err)
	}
	slog.Info("Downloaded: " + r.path)
	Uninstall()
	harvestObj := new(Harvest)
	slog.Info("Installing " + rpmFileName)
	installOutput, err := utils.Run("yum", "install", "-y", rpmFileName)
	if err != nil {
		slog.Error("", slogx.Err(err))
		panic(err)
	}
	slog.Info(installOutput)
	slog.Info("Stopping harvest")
	harvestObj.Stop()
	_, err = utils.Run("cp", harvestFile, HarvestHome+"/"+harvestFile)
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
	utils.RemoveSafely(rpmFileName)
	harvestObj := new(Harvest)
	if !harvestObj.AllRunning("keyperf") {
		utils.PanicIfNotNil(errors.New("pollers are not in a running state before upgrade"))
	}
	versionCmd := []string{"-qa", "harvest"}
	out, err := utils.Run("rpm", versionCmd...)
	if err != nil {
		slog.Error("", slogx.Err(err))
		panic(err)
	}
	previousVersion := strings.TrimSpace(out)
	err = utils.DownloadFile(rpmFileName, r.path)
	utils.PanicIfNotNil(err)
	slog.Info("Downloaded: " + r.path)
	slog.Info("Updating " + rpmFileName)
	installOutput, _ := utils.Run("yum", "upgrade", "-y", rpmFileName)
	slog.Info(installOutput)
	out, _ = utils.Run("rpm", versionCmd...)
	installedVersion := strings.TrimSpace(out)
	if previousVersion == installedVersion {
		utils.PanicIfNotNil(errors.New("upgrade is failed"))
	}
	_, _ = utils.Run("cp", GetPerfFileWithQosCounters(ZapiPerfDefaultFile, "defaultZapi.yaml"), HarvestHome+"/"+ZapiPerfDefaultFile)
	_, _ = utils.Run("cp", GetPerfFileWithQosCounters(RestPerfDefaultFile, "defaultRest.yaml"), HarvestHome+"/"+RestPerfDefaultFile)
	harvestObj.Stop()
	harvestObj.Start()
	status := harvestObj.AllRunning()
	asupExecPath := HarvestHome + "/autosupport/asup"
	isValidAsup := harvestObj.IsValidAsup(asupExecPath)
	return status && isValidAsup
}

func (r *RPM) Stop() bool {
	if utils.FileExists(HarvestHome) {
		harvestObj := new(Harvest)
		if utils.FileExists(HarvestHome + "/bin/harvest") {
			if harvestObj.AllRunning() {
				harvestObj.Stop()
			}
		}
	}
	return true
}
