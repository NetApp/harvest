package installer

import (
	"github.com/Netapp/harvest-automation/test/cmds"
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
	_, _ = cmds.Run("yum", "remove", "-y", "harvest")
}

func UninstallNativePkg() {
	slog.Info("Uninstalling native pkg if any")
	if cmds.FileExists(HarvestHome) {
		harvestObj := new(Harvest)
		if cmds.FileExists(HarvestHome + "/bin/harvest") {
			if harvestObj.AllRunning() {
				harvestObj.Stop()
			}
		}
		_, _ = cmds.Run("rm", "-rf", HarvestHome)
	} else {
		slog.Info("Harvest doesnt exists.", slog.String("HarvestHome", HarvestHome))
	}
}

func CleanLogDir() {
	if cmds.FileExists(LogDir) {
		_, _ = cmds.Run("rm", "-rf", LogDir)
	}
}

func CreateLogDir() {
	if !cmds.FileExists(LogDir) {
		cmds.MkDir(LogDir)
	}
}
