package installer

import (
	"goharvest2/integration/test/utils"
	"log"
)

type RPM struct {
	path string
}

func (r *RPM) Init(path string) {
	r.path = path
}

func (r *RPM) Install() bool {
	//harvestFile := "harvest.yml"
	rpmFileName := "harvest.rpm"
	utils.RemoveSafely(rpmFileName)
	err := utils.DownloadFile(rpmFileName, r.path)
	if err != nil {
		panic(err)
	}
	log.Println("Downloaded: " + r.path)
	log.Println("Check and remove harvest ")
	unInstallOutput := utils.Run("yum", "remove", "-y", "harvest")
	log.Println(unInstallOutput)
	log.Println("Installing " + rpmFileName)
	installOutput := utils.Run("yum", "install", "-y", rpmFileName)
	log.Println(installOutput)
	//utils.RemoveSafely(harvestFile)
	//utils.CopyFile(HARVEST_HOME+"/harvest.yml", harvestFile) use file directly from the repo
	harvestObj := new(Harvest)
	harvestObj.Start()
	status := harvestObj.AllRunning()
	return status
}
