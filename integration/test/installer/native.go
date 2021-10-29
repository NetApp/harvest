package installer

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/setup"
	"github.com/Netapp/harvest-automation/test/utils"
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
	Uninstall()
	log.Println("Installing " + tarFileName)
	unTarOutput := utils.Run("tar", "-xf", tarFileName, "--one-top-level=harvest", "--strip-components", "1", "-C", "/opt")
	log.Println(unTarOutput)
	utils.RemoveSafely(HarvestHome + "/" + harvestFile)
	log.Println("Copy certificates files into harvest directory")
	path := HarvestHome + "/certificates"
	if utils.FileExists(path) {
		err = utils.RemoveDir(path)
		utils.PanicIfNotNil(err)
	}

	utils.Run("mkdir", "-p", path)
	utils.Run("cp", "-R", utils.GetConfigDir()+"/certificates", HarvestHome)
	utils.Run("cp", setup.GetZapiPerfFileWithQosCounters(), HarvestHome+"/"+setup.ZapiPerfDefaultFile)
	err = utils.CopyFile(harvestFile, HarvestHome+"/"+harvestFile)
	if err != nil {
		panic(err)
	}
	harvestObj.Start()
	status := harvestObj.AllRunning()
	return status
}

func (n *Native) Upgrade() bool {
	utils.PanicIfNotNil(fmt.Errorf("not supported"))
	return false
}
