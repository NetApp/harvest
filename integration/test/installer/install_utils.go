package installer

import (
	"github.com/Netapp/harvest-automation/test/utils"
	"log/slog"
)

const LogDir = "/var/log/harvest"

func Uninstall() {
	slog.Info("Check and remove harvest ")
	UninstallRPM()
	UninstallNativePkg()
	CleanLogDir()
}

func UninstallRPM() {
	_, _ = utils.Run("yum", "remove", "-y", "harvest")
}

func UninstallNativePkg() {
	slog.Info("Uninstalling native pkg if any")
	if utils.FileExists(HarvestHome) {
		harvestObj := new(Harvest)
		if utils.FileExists(HarvestHome + "/bin/harvest") {
			if harvestObj.AllRunning() {
				harvestObj.Stop()
			}
		}
		_, _ = utils.Run("rm", "-rf", HarvestHome)
	} else {
		slog.Info("Harvest doesnt exists.", slog.String("HarvestHome", HarvestHome))
	}
}

func CleanLogDir() {
	if utils.FileExists(LogDir) {
		_, _ = utils.Run("rm", "-rf", LogDir)
	}
}

func CreateLogDir() {
	if !utils.FileExists(LogDir) {
		utils.MkDir(LogDir)
	}
}
