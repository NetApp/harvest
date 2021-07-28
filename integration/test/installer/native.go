package installer

import (
	"goharvest2/integration/test/utils"
	"log"
)

type Native struct {
	path string
}

func (r *Native) Init(path string) {
	r.path = path
}

func (r *Native) Install() bool {
	harvestFile := "harvest.yml"
	harvestObj := new(Harvest)
	tarFileName := "harvest.tar.gz"
	utils.RemoveSafely(tarFileName)
	err := utils.DownloadFile(tarFileName, r.path)
	if err != nil {
		panic(err)
	}
	log.Println("Downloaded: " + r.path)
	log.Println("Check and remove harvest ")
	if utils.FileExists(HarvestHome) && !harvestObj.AllStopped() {
		harvestObj.Stop()
	}
	unInstallOutput := utils.Run("rm", "-rf", HarvestHome)
	log.Println(unInstallOutput)
	log.Println("Installing " + tarFileName)
	unTarOutput := utils.Run("tar", "-xf", tarFileName, "--one-top-level=harvest", "--strip-components", "1", "-C", "/opt")
	log.Println(unTarOutput)
	utils.RemoveSafely(HarvestHome + "/" + harvestFile)
	utils.CopyFile(harvestFile, HarvestHome+"/"+harvestFile)
	harvestObj.Start()
	status := harvestObj.AllRunning()
	return status
}
