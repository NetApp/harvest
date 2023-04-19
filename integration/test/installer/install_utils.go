package installer

import (
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
)

const LogDir = "/var/log/harvest"

func Uninstall() {
	log.Println("Check and remove harvest ")
	UninstallRPM()
	UninstallNativePkg()
	CleanLogDir()
}

func UninstallRPM() {
	_, _ = utils.Run("yum", "remove", "-y", "harvest")
}

func UninstallNativePkg() {
	log.Println("Uninstalling  native pkg if any")
	if utils.FileExists(HarvestHome) {
		harvestObj := new(Harvest)
		if utils.FileExists(HarvestHome + "/bin/harvest") {
			if harvestObj.AllRunning() {
				harvestObj.Stop()
			}
		}
		_, _ = utils.Run("rm", "-rf", HarvestHome)
	} else {
		log.Printf(" %s doesnt exists.\n", HarvestHome)
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
