package installer

import (
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
)

const LogDir = "/var/log/harvest"

func Uninstall() {
	log.Println("Check and remove harvest ")
	UninstallRPM()
	UninstallNativePkg()
	UninstallPollerDocker()
	CleanLogDir()
}

func UninstallRPM() {
	utils.Run("yum", "remove", "-y", "harvest")
}

func UninstallNativePkg() {
	log.Println("Uninstalling  native pkg if any")
	if utils.FileExists(HarvestHome) {
		harvestObj := new(Harvest)
		if utils.FileExists(HarvestHome + "bin/harvest") {
			if harvestObj.AllRunning() {
				harvestObj.Stop()
			}
		}
		utils.Run("rm", "-rf", HarvestHome)
	} else {
		log.Printf(" %s doesnt exists.\n", HarvestHome)
	}
}

func UninstallPollerDocker() {
	pollerProcessName := "bin/poller"
	containerIds := docker.GetContainerID(pollerProcessName)
	if len(containerIds) > 0 {
		docker.StopContainers(pollerProcessName)
		docker.RemoveImage("harvest")
	} else {
		log.Println("No pollers were running")
	}
}

func CleanLogDir() {
	if utils.FileExists(LogDir) {
		utils.Run("rm", "-rf", LogDir)
	}
}

func CreateLogDir() {
	if !utils.FileExists(LogDir) {
		utils.MkDir(LogDir)
	}
}
